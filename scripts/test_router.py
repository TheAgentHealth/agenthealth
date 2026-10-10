#!/usr/bin/env python3
"""Real LiteLLM process liveness signal, with no model calls or routing claims."""
import argparse
import json
import os
from pathlib import Path
import socket
import secrets
import subprocess
import tempfile
import time
import urllib.request

parser = argparse.ArgumentParser()
parser.add_argument('agenthealth', type=Path)
parser.add_argument('litellm', type=Path)
args = parser.parse_args()
with socket.socket() as listener:
    listener.bind(('127.0.0.1', 0))
    port = listener.getsockname()[1]
with tempfile.TemporaryDirectory() as directory:
    config = Path(directory) / 'router.yaml'
    config.write_text('model_list: []\nlitellm_settings:\n  telemetry: false\n')
    with (Path(directory) / 'router.log').open('w+') as log:
        process = subprocess.Popen([str(args.litellm.resolve()), '--config', str(config),
                                    '--host', '127.0.0.1', '--port', str(port)], stdout=log, stderr=log,
                                    env={**{key: os.environ[key] for key in ('PATH', 'LANG', 'LC_ALL', 'HOME', 'SSL_CERT_FILE', 'SSL_CERT_DIR') if key in os.environ},
                                         'PATH': str(args.litellm.resolve().parent) + os.pathsep + os.environ.get('PATH', ''),
                                         'LITELLM_MASTER_KEY': 'sk-' + secrets.token_hex(32),
                                         'LITELLM_TELEMETRY': 'false'}, cwd=directory)
        try:
            endpoint = f'http://127.0.0.1:{port}/health/liveliness'
            deadline = time.monotonic() + 30
            while True:
                if process.poll() is not None:
                    raise SystemExit('LiteLLM exited before liveness')
                try:
                    urllib.request.urlopen(endpoint, timeout=1).close()
                    break
                except OSError:
                    if time.monotonic() > deadline:
                        raise SystemExit('LiteLLM liveness timed out')
                    time.sleep(0.1)
            health = Path(directory) / 'health.json'
            health.write_text(json.dumps({'version': 'v1', 'targets': [{
                'name': 'real-router', 'type': 'router', 'endpoint': endpoint,
                'checks': ['reachability', 'protocol']
            }]}))
            run = subprocess.run([str(args.agenthealth.resolve()), 'check', str(health), '--format', 'json'],
                                 capture_output=True, text=True, timeout=30)
            if run.returncode != 0 or json.loads(run.stdout)['status'] != 'HEALTHY':
                raise SystemExit('Real LiteLLM router HTTP health signal failed')
            print('LiteLLM process liveness endpoint: HEALTHY')
        finally:
            process.terminate()
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait()
