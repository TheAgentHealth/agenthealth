"""Opt-in integration tests against official SDK protocol implementations."""
import importlib.metadata
import json
import os
from pathlib import Path
import socket
import subprocess
import sys
import time
import urllib.error
import urllib.request

import pytest
import yaml

pytestmark = pytest.mark.skipif(os.environ.get("AGENTHEALTH_INTEROP") != "1", reason="official SDK suite runs separately")
SERVER = Path(__file__).with_name("server.py")


@pytest.fixture
def binary():
    path = os.environ.get("AGENTHEALTH_BINARY")
    assert path and Path(path).is_file(), "set AGENTHEALTH_BINARY to the built CLI"
    return path


def check(binary, tmp_path, target, expected=0):
    config = tmp_path / "check.yaml"
    config.write_text(yaml.safe_dump({"version": "v1", "targets": [target]}))
    run = subprocess.run([binary, "check", str(config), "--format", "json"], capture_output=True, text=True, timeout=30)
    assert run.returncode == expected, run.stdout + run.stderr
    result = json.loads(run.stdout)
    assert result["status"] == ("HEALTHY" if expected == 0 else "UNHEALTHY"), result
    return result


@pytest.mark.parametrize("protocol", ["mcp", "a2a"])
def test_official_http_server(binary, tmp_path, protocol):
    with socket.socket() as listener:
        listener.bind(("127.0.0.1", 0))
        port = listener.getsockname()[1]
    endpoint = f"http://127.0.0.1:{port}"
    with (tmp_path / "server.log").open("w+") as log:
        process = subprocess.Popen([sys.executable, str(SERVER), protocol, "--port", str(port)], stdout=log, stderr=log)
        try:
            ready = False
            for _ in range(100):
                if process.poll() is not None:
                    log.seek(0)
                    pytest.fail(log.read())
                try:
                    urllib.request.urlopen(endpoint + ("/mcp" if protocol == "mcp" else "/.well-known/agent-card.json"), timeout=0.2).close()
                    ready = True
                    break
                except urllib.error.HTTPError:
                    ready = True
                    break
                except (OSError, urllib.error.URLError):
                    time.sleep(0.05)
            assert ready, "SDK server did not start"
            target = {"name": "official", "type": protocol, "endpoint": endpoint,
                      "checks": ["authentication", "protocol", "capability", "functional"]}
            if protocol == "mcp":
                target["endpoint"] += "/mcp"
                target["mcp"] = {"required_tools": ["health"], "functional": {"tool": "health", "safe": True}}
            else:
                target["a2a"] = {"required_skills": ["health"], "functional": {"safe": True, "text": "health"}}
                if importlib.metadata.version("a2a-sdk").startswith("0."):
                    target["a2a"]["protocol_version"] = "0.3.0"
            check(binary, tmp_path, target)
            target["checks"] = ["authentication", "protocol", "capability"]
            options = target[protocol]
            options.pop("functional")
            options["required_tools" if protocol == "mcp" else "required_skills"] = ["missing"]
            result = check(binary, tmp_path, target, expected=2)
            assert result["checks"]["capability"]["code"] == protocol + "_required"
        finally:
            process.terminate()
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait(timeout=5)


def test_official_mcp_stdio(binary, tmp_path):
    target = {"name": "official-stdio", "type": "mcp", "endpoint": "stdio://official-sdk",
              "checks": ["protocol", "capability", "functional"],
              "mcp": {"transport": "stdio", "stdio": {"command": sys.executable, "args": [str(SERVER), "mcp", "--transport", "stdio"]},
                      "required_tools": ["health"], "functional": {"safe": True, "tool": "health"}}}
    check(binary, tmp_path, target)
