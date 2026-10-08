#!/usr/bin/env python3
"""Smoke-test a built AgentHealth container image's runtime behavior."""
import argparse
import http.server
import json
import subprocess
import tempfile
import threading
from pathlib import Path


def docker(*args):
    return subprocess.run(["docker", *args], capture_output=True, text=True)


def check(condition, message):
    if not condition:
        raise SystemExit(f"FAIL: {message}")


def test_version(image, version):
    result = docker("run", "--rm", image, "version")
    check(result.returncode == 0, f"version command failed: {result.stderr}")
    check(result.stdout.strip() == f"agenthealth {version}", f"unexpected version output: {result.stdout!r}")


def test_default_help(image):
    result = docker("run", "--rm", image)
    check(result.returncode == 0, f"default command failed: {result.stderr}")
    check("Usage:" in result.stdout, f"missing usage text: {result.stdout!r}")


def test_image_config(image):
    result = docker("inspect", image, "--format", "{{json .Config}}")
    check(result.returncode == 0, f"docker inspect failed: {result.stderr}")
    config = json.loads(result.stdout)
    check(config["User"] == "65532:65532", f"expected a non-root user, got {config['User']!r}")
    check(config["Entrypoint"] == ["/usr/local/bin/agenthealth"], f"unexpected entrypoint: {config['Entrypoint']!r}")
    check("8080/tcp" in config.get("ExposedPorts", {}), "missing EXPOSE 8080")


def test_functional_check(image):
    """Run an actual health check from inside the container against a host HTTP server."""
    class Handler(http.server.BaseHTTPRequestHandler):
        def do_GET(self):
            self.send_response(200)
            self.end_headers()

        def log_message(self, *_args):
            pass

    server = http.server.HTTPServer(("127.0.0.1", 0), Handler)
    thread = threading.Thread(target=server.serve_forever, daemon=True)
    thread.start()
    try:
        with tempfile.TemporaryDirectory() as tmp:
            config = Path(tmp) / "agenthealth.yaml"
            config.write_text(
                "version: v1\n"
                "targets:\n"
                "  - name: container-check\n"
                "    type: http\n"
                f"    endpoint: http://127.0.0.1:{server.server_port}/\n"
                "    checks:\n"
                "      - reachability\n"
            )
            # The container runs as a fixed non-root UID; both must be world-readable.
            Path(tmp).chmod(0o755)
            config.chmod(0o644)
            result = docker(
                "run", "--rm", "--network", "host",
                "-v", f"{tmp}:/config:ro",
                image, "check", "/config/agenthealth.yaml", "--format", "json",
            )
            check(result.returncode == 0, f"check command failed (exit {result.returncode}): {result.stderr}\n{result.stdout}")
            payload = json.loads(result.stdout)
            check(payload["status"] == "HEALTHY", f"expected HEALTHY status, got {payload!r}")
    finally:
        server.shutdown()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("image", help="built image tag, such as agenthealth:test")
    parser.add_argument("--version", required=True, help="expected `agenthealth version` output suffix")
    args = parser.parse_args()
    test_version(args.image, args.version)
    test_default_help(args.image)
    test_image_config(args.image)
    test_functional_check(args.image)
    print("All container checks passed.", flush=True)


if __name__ == "__main__":
    main()
