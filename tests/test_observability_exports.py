"""Validate CLI payloads with independent Prometheus and OTLP parsers."""
import json
from pathlib import Path
import subprocess

import pytest
from google.protobuf.json_format import Parse
from opentelemetry.proto.collector.logs.v1.logs_service_pb2 import ExportLogsServiceRequest
from prometheus_client.parser import text_string_to_metric_families

ROOT = Path(__file__).resolve().parents[1]


@pytest.fixture(scope="module")
def export_cli(tmp_path_factory):
    binary = tmp_path_factory.mktemp("observability") / "agenthealth"
    subprocess.run(["go", "build", "-o", str(binary), "./cmd/agenthealth"], cwd=ROOT, check=True)
    return binary


def test_independent_export_parsers(export_cli, tmp_path):
    # No services or external credentials: refusing localhost yields real
    # nonhealthy engine evidence and still must produce consumable exports.
    config = tmp_path / "health.yaml"
    config.write_text("""version: v1
targets:
  - id: private-peer
    name: private-peer-name
    type: http
    endpoint: http://127.0.0.1:1
  - id: private-root
    name: private-root-name
    type: http
    endpoint: http://127.0.0.1:1
    dependencies:
      - ref: private-peer
        critical: true
        relationship: supporting
      - ref: private-peer
        critical: true
        relationship: supporting
""")
    payloads = {}
    for fmt in ("prometheus", "logs", "otlp", "webhook"):
        result = subprocess.run([str(export_cli), "check", str(config), "--format", fmt],
                                capture_output=True, text=True, timeout=15)
        assert 0 < result.returncode < 6, result.stderr
        assert not result.stderr
        assert "private" not in result.stdout
        assert "127.0.0.1" not in result.stdout
        payloads[fmt] = result.stdout
    families = list(text_string_to_metric_families(payloads["prometheus"]))
    samples = [sample for family in families for sample in family.samples]
    identities = [(s.name, tuple(sorted(s.labels.items()))) for s in samples]
    assert len(identities) == len(set(identities)), "duplicate metric series"
    states = {}
    for sample in samples:
        if sample.name == "agenthealth_status":
            key = tuple(sorted((k, v) for k, v in sample.labels.items() if k != "status"))
            states.setdefault(key, []).append(sample.value)
    assert all(len(values) == 6 and sum(values) == 1 for values in states.values())
    otlp = Parse(payloads["otlp"], ExportLogsServiceRequest())
    logs = otlp.resource_logs[0].scope_logs[0].log_records
    records = [json.loads(line) for line in payloads["logs"].splitlines()]
    assert [json.loads(log.body.string_value) for log in logs] == records
    assert all(log.observed_time_unix_nano > 0 for log in logs)
    assert json.loads(payloads["webhook"]) == {"export_version": "v1", "records": records}
    assert sum(record["kind"] == "node" for record in records) == 2
    edges = [record for record in records if record["kind"] == "edge"]
    assert len(edges) == 2 and len({edge["edge"] for edge in edges}) == 2
