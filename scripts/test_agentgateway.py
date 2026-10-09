#!/usr/bin/env python3
"""Real Agentgateway readiness signal; does not certify routing interoperability."""
import argparse
import json
from pathlib import Path
import socket
import subprocess
import tempfile
import time
import urllib.request

parser = argparse.ArgumentParser()
parser.add_argument('agenthealth', type=Path)
parser.add_argument('agentgateway', type=Path)
args = parser.parse_args()
with socket.socket() as listener:
    listener.bind(('127.0.0.1', 0))
    port = listener.getsockname()[1]
with tempfile.TemporaryDirectory() as directory:
    config = Path(directory) / 'gateway.yaml'
    config.write_text(f'config:\n  adminAddr: 127.0.0.1:0\n  readinessAddr: 127.0.0.1:{port}\n')
    with (Path(directory) / 'gateway.log').open('w+') as log:
        process = subprocess.Popen([str(args.agentgateway.resolve()), '-f', str(config)], stdout=log, stderr=log)
        try:
            endpoint = f'http://127.0.0.1:{port}/healthz/ready'
            deadline = time.monotonic() + 20
            while True:
                if process.poll() is not None:
                    raise SystemExit('Agentgateway exited before readiness')
                try:
                    urllib.request.urlopen(endpoint, timeout=1).close()
                    break
                except OSError:
                    if time.monotonic() > deadline:
                        raise SystemExit('Agentgateway readiness timed out')
                    time.sleep(0.1)
            health = Path(directory) / 'health.json'
            health.write_text(json.dumps({'version': 'v1', 'targets': [{
                'name': 'real-agentgateway', 'type': 'gateway', 'endpoint': endpoint,
                'checks': ['reachability', 'protocol']
            }]}))
            run = subprocess.run([str(args.agenthealth.resolve()), 'check', str(health), '--format', 'json'],
                                 capture_output=True, text=True, timeout=30)
            if run.returncode != 0 or json.loads(run.stdout)['status'] != 'HEALTHY':
                raise SystemExit('Real Agentgateway health signal failed')
            print('Agentgateway readiness endpoint: HEALTHY')
        finally:
            process.terminate()
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait()
