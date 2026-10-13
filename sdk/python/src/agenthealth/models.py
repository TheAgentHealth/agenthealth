"""Typed views of AHS v1 output; preserve additional evidence in to_dict()."""
from copy import deepcopy
from dataclasses import dataclass, field
from enum import Enum
import json
from types import MappingProxyType
from typing import Any, Mapping

from .config import validator
from .errors import OutputError


class HealthStatus(str, Enum):
    HEALTHY = "HEALTHY"
    DEGRADED = "DEGRADED"
    UNHEALTHY = "UNHEALTHY"
    UNREACHABLE = "UNREACHABLE"
    MISCONFIGURED = "MISCONFIGURED"
    UNKNOWN = "UNKNOWN"


class Relationship(str, Enum):
    SUPPORTING = "supporting"
    DOWNSTREAM = "downstream"
    PATH = "path"
    GATEWAY = "gateway"
    ROUTER = "router"


@dataclass(frozen=True)
class Target:
    name: str
    type: str
    id: str | None = None


@dataclass(frozen=True)
class CheckEntry:
    status: HealthStatus
    message: str | None
    code: str | None
    steps: Mapping[str, HealthStatus]


@dataclass(frozen=True)
class HealthResult:
    target: Target
    status: HealthStatus
    checks: Mapping[str, CheckEntry]
    dependencies: tuple["HealthResult", ...]
    latency_ms: float | None
    relationship: Relationship | None
    critical: bool | None
    spec_version: str
    _document: dict[str, Any] = field(repr=False, compare=False)

    def to_dict(self) -> dict[str, Any]:
        return deepcopy(self._document)


@dataclass(frozen=True)
class CheckReport:
    """Ordered root results and the CLI's aggregate exit code (0–5)."""
    results: tuple[HealthResult, ...]
    exit_code: int
    spec_version: str
    _document: dict[str, Any] = field(repr=False, compare=False)

    @property
    def status(self) -> HealthStatus:
        return tuple(HealthStatus)[self.exit_code]

    @property
    def result(self) -> HealthResult:
        """Single root convenience; batches require explicit results iteration."""
        if len(self.results) != 1:
            raise ValueError("Expected a single root result")
        return self.results[0]

    def to_dict(self) -> dict[str, Any]:
        return deepcopy(self._document)


def _result(data: dict[str, Any], version: str) -> HealthResult:
    checks = {name: CheckEntry(HealthStatus(check["status"]), check.get("message"),
                              check.get("code"), MappingProxyType({key: HealthStatus(value)
                              for key, value in check.get("steps", {}).items()}))
              for name, check in data["checks"].items()}
    target = data["target"]
    return HealthResult(Target(target["name"], target["type"], target.get("id")),
                        HealthStatus(data["status"]), MappingProxyType(checks),
                        tuple(_result(child, version) for child in data["dependencies"]),
                        data.get("latency_ms"),
                        Relationship(data["relationship"]) if "relationship" in data else None,
                        data.get("critical"), version, deepcopy(data))


def parse_report(output: bytes, exit_code: int) -> CheckReport:
    try:
        document = json.loads(output, parse_constant=lambda _: (_ for _ in ()).throw(ValueError()))
        validator("result").validate(document)
        if document["spec_version"] != "v1" or exit_code not in range(6):
            raise ValueError()
        roots = document["results"] if "results" in document else [document]
        results = tuple(_result(root, "v1") for root in roots)
        expected = max(tuple(HealthStatus).index(root.status) for root in results)
        if exit_code != expected:
            raise ValueError()
        return CheckReport(results, exit_code, "v1", deepcopy(document))
    except Exception:
        raise OutputError("Invalid AgentHealth result or exit-code contract") from None
