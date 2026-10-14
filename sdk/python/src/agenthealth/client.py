"""Bounded CLI execution with native async cancellation and no shell invocation."""
import asyncio
from contextlib import contextmanager
import json
import math
import os
import re
from pathlib import Path
import shutil
import signal
import tempfile
import time
from typing import Any, Mapping, Sequence

from .config import Configuration, load_config
from .errors import (BinaryNotFoundError, CheckTimeoutError, CompatibilityError,
                     ConfigurationError, ExecutionError, OutputError)
from .models import CheckReport, parse_report
from .version import __version__

MAX_OUTPUT_BYTES = 16 * 1024 * 1024
MAX_DIAGNOSTIC_BYTES = 64 * 1024


async def _signal_group(process: asyncio.subprocess.Process, sig: int) -> None:
    try:
        os.killpg(process.pid, sig)
    except ProcessLookupError:
        pass
    except PermissionError:
        # Darwin can report EPERM for an already-exited process group. Confirm
        # root exit before accepting it; a live permission failure remains fatal.
        if process.returncode is None:
            try:
                await asyncio.wait_for(process.wait(), 1)
            except asyncio.TimeoutError:
                raise ExecutionError("Unable to stop AgentHealth CLI process") from None


async def _stop(process: asyncio.subprocess.Process) -> None:
    # Kill the dedicated process group even if the root exited, so an MCP stdio
    # descendant cannot retain credentials or the temporary output descriptors.
    try:
        if os.name == "posix":
            await _signal_group(process, signal.SIGTERM)
        elif process.returncode is None:
            process.terminate()
    except ProcessLookupError:
        pass
    try:
        await asyncio.wait_for(process.wait(), 1)
    except asyncio.TimeoutError:
        pass
    finally:
        try:
            if os.name == "posix":
                await _signal_group(process, signal.SIGKILL)
            elif process.returncode is None:
                process.kill()
        except ProcessLookupError:
            pass
        await process.wait()


async def _run(command: list[str], *, timeout: float, cwd: Path,
               environment: Mapping[str, str] | None) -> tuple[int, bytes]:
    env = None if environment is None else {**os.environ, **environment}
    with tempfile.TemporaryFile() as output, tempfile.TemporaryFile() as diagnostic:
        try:
            process = await asyncio.create_subprocess_exec(
                *command, stdin=asyncio.subprocess.DEVNULL, stdout=output, stderr=diagnostic,
                cwd=cwd, env=env, start_new_session=os.name == "posix")
        except OSError:
            raise BinaryNotFoundError("Unable to start AgentHealth CLI") from None
        deadline = time.monotonic() + timeout
        try:
            while True:
                if (os.fstat(output.fileno()).st_size > MAX_OUTPUT_BYTES or
                        os.fstat(diagnostic.fileno()).st_size > MAX_DIAGNOSTIC_BYTES):
                    raise OutputError("AgentHealth output exceeded the SDK limit")
                if process.returncode is not None:
                    break
                remaining = deadline - time.monotonic()
                if remaining <= 0:
                    raise CheckTimeoutError("AgentHealth check deadline expired")
                await asyncio.sleep(min(0.02, remaining))
            output.seek(0)
            return process.returncode, output.read(MAX_OUTPUT_BYTES + 1)
        finally:
            # A separate task keeps cleanup alive if another cancellation arrives.
            cleanup = asyncio.create_task(_stop(process))
            try:
                await asyncio.shield(cleanup)
            except asyncio.CancelledError:
                await cleanup
                raise


@contextmanager
def _config_file(configuration: Configuration):
    with tempfile.TemporaryDirectory(prefix="agenthealth-sdk-") as directory:
        path = Path(directory) / "configuration.json"
        # Directory mode is 0700; file is never passed in command text or stderr.
        descriptor = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
        with os.fdopen(descriptor, "w") as stream:
            json.dump(configuration.to_dict(), stream, allow_nan=False)
        yield path


def _configuration(config: Configuration | Mapping[str, Any] | str | Path | None,
                   target_type: str | None, endpoint: str | None, name: str,
                   checks: Sequence[str] | None, options: Mapping[str, Any] | None) -> Configuration:
    if config is not None:
        if target_type is not None or endpoint is not None or checks is not None or options is not None:
            raise ConfigurationError("Choose a configuration or individual target arguments")
        if isinstance(config, Configuration):
            return config
        if isinstance(config, Mapping):
            return Configuration(config)
        return load_config(config)
    if target_type is None or endpoint is None:
        raise ConfigurationError("A configuration or target_type and endpoint are required")
    target: dict[str, Any] = dict(options or {})
    if any(key in target for key in ("name", "type", "endpoint", "checks")):
        raise ConfigurationError("Target options must not override named arguments")
    target.update(name=name, type=target_type, endpoint=endpoint)
    if checks is not None:
        target["checks"] = list(checks)
    return Configuration({"version": "v1", "targets": [target]})


class Client:
    """Use an installed CLI of exactly the SDK version; never download or execute
    a binary at import/install time. Pass a verified binary path when pinning it.
    """

    def __init__(self, binary: str | Path = "agenthealth", *, timeout: float = 30,
                 environment: Mapping[str, str] | None = None):
        if isinstance(timeout, bool) or not isinstance(timeout, (int, float)) or not math.isfinite(timeout) or timeout <= 0:
            raise ValueError("timeout must be a positive finite number")
        self.binary = str(binary)
        self.timeout = timeout
        self.environment = None if environment is None else dict(environment)

    async def async_check(self, config: Configuration | Mapping[str, Any] | str | Path | None = None,
                          *, target_type: str | None = None, endpoint: str | None = None,
                          name: str = "target", checks: Sequence[str] | None = None,
                          options: Mapping[str, Any] | None = None) -> CheckReport:
        configuration = _configuration(config, target_type, endpoint, name, checks, options)
        # Resolve before changing cwd; relative explicit binary paths retain their meaning.
        binary = shutil.which(self.binary)
        if binary is None:
            raise BinaryNotFoundError("AgentHealth CLI executable was not found")
        binary = str(Path(binary).resolve())
        deadline = time.monotonic() + self.timeout
        code, version = await _run([binary, "version"], timeout=min(self.timeout, 5),
                                   cwd=configuration.base_dir, environment=self.environment)
        cli_version = re.sub(r"(a|b|rc)([0-9]+)$", r"-\1.\2", __version__)
        if code != 0 or version.strip() != f"agenthealth v{cli_version}".encode():
            raise CompatibilityError("AgentHealth CLI must match the SDK software version")
        remaining = deadline - time.monotonic()
        if remaining <= 0:
            raise CheckTimeoutError("AgentHealth check deadline expired")
        with _config_file(configuration) as path:
            code, output = await _run([binary, "check", "--format", "json", "--", str(path)],
                                      timeout=remaining, cwd=configuration.base_dir,
                                      environment=self.environment)
        if code not in range(6):
            raise ExecutionError("AgentHealth CLI execution failed")
        return parse_report(output, code)

    def check(self, config: Configuration | Mapping[str, Any] | str | Path | None = None,
              *, target_type: str | None = None, endpoint: str | None = None,
              name: str = "target", checks: Sequence[str] | None = None,
              options: Mapping[str, Any] | None = None) -> CheckReport:
        try:
            asyncio.get_running_loop()
        except RuntimeError:
            return asyncio.run(self.async_check(config, target_type=target_type, endpoint=endpoint,
                                                name=name, checks=checks, options=options))
        raise RuntimeError("Use await client.async_check() inside an event loop")


def check(config: Configuration | Mapping[str, Any] | str | Path | None = None, *,
          target_type: str | None = None, endpoint: str | None = None, name: str = "target",
          checks: Sequence[str] | None = None, options: Mapping[str, Any] | None = None,
          binary: str | Path = "agenthealth", timeout: float = 30,
          environment: Mapping[str, str] | None = None) -> CheckReport:
    return Client(binary, timeout=timeout, environment=environment).check(
        config, target_type=target_type, endpoint=endpoint, name=name, checks=checks, options=options)


async def async_check(config: Configuration | Mapping[str, Any] | str | Path | None = None, *,
                      target_type: str | None = None, endpoint: str | None = None, name: str = "target",
                      checks: Sequence[str] | None = None, options: Mapping[str, Any] | None = None,
                      binary: str | Path = "agenthealth", timeout: float = 30,
                      environment: Mapping[str, str] | None = None) -> CheckReport:
    return await Client(binary, timeout=timeout, environment=environment).async_check(
        config, target_type=target_type, endpoint=endpoint, name=name, checks=checks, options=options)
