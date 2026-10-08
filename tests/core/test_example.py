"""Verify real core output against the language-neutral result schema."""
import http.server
import json
import os
import pathlib
import shutil
import subprocess
import threading

import jsonschema
import pytest

ROOT = pathlib.Path(__file__).resolve().parents[2]

@pytest.mark.skipif(shutil.which("go") is None, reason="Go toolchain required")
def test_core_example_output_matches_schema(tmp_path):
    class Handler(http.server.BaseHTTPRequestHandler):
        def do_HEAD(self):
            self.send_response(204 if self.headers.get("Authorization") == "Bearer private-token" else 401)
            self.end_headers()

        def log_message(self, *args):
            pass

    server = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Handler)
    worker = threading.Thread(target=server.serve_forever, daemon=True)
    worker.start()
    config = tmp_path / "config.yaml"
    target = f"""  - name: example
    type: http
    endpoint: http://127.0.0.1:{server.server_port}/health
    auth: {{bearer_env: AGENTHEALTH_EXAMPLE_TOKEN}}
"""
    try:
        for count in (1, 2):
            config.write_text("version: v1\ntargets:\n" + target * count)
            completed = subprocess.run(
                ["go", "run", "./examples/core-check", str(config)],
                cwd=ROOT,
                env={**os.environ, "AGENTHEALTH_EXAMPLE_TOKEN": "private-token"},
                capture_output=True, text=True, check=True, timeout=60,
            )
            assert "private-token" not in completed.stdout + completed.stderr
            output = json.loads(completed.stdout)
            schema = json.loads((ROOT / "spec/schemas/result.schema.json").read_text())
            jsonschema.validate(output, schema)
            results = [output] if count == 1 else output["results"]
            assert len(results) == count
            assert all(r["status"] == "HEALTHY" for r in results)
        # Validate real graph output against the language-neutral schema too.
        graph_target = target.replace("  - name:", "  - id: root\n    name:", 1)
        graph_target += "    dependencies:\n      - ref: backend\n        relationship: supporting\n        critical: false\n      - ref: backend\n        relationship: downstream\n"
        backend = target.replace("  - name:", "  - id: backend\n    name:", 1)
        config.write_text("version: v1\nconcurrency: 2\ntargets:\n" + graph_target + backend)
        completed = subprocess.run(
            ["go", "run", "./examples/core-check", str(config)], cwd=ROOT,
            env={**os.environ, "AGENTHEALTH_EXAMPLE_TOKEN": "private-token"},
            capture_output=True, text=True, check=True, timeout=60,
        )
        assert "private-token" not in completed.stdout + completed.stderr
        output = json.loads(completed.stdout)
        jsonschema.validate(output, schema)
        root, shared = output["results"]
        assert root["status"] == shared["status"] == "HEALTHY"
        assert [d["target"]["id"] for d in root["dependencies"]] == ["backend", "backend"]
        assert [d["critical"] for d in root["dependencies"]] == [False, True]
        assert [d["relationship"] for d in root["dependencies"]] == ["supporting", "downstream"]
    finally:
        server.shutdown()
        server.server_close()
        worker.join(timeout=5)
