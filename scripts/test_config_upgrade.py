#!/usr/bin/env python3
"""Run one unchanged released configuration against old/new/old binaries."""
import argparse
import json
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
import subprocess
import tempfile
import threading

parser = argparse.ArgumentParser()
parser.add_argument('old_binary', type=Path)
parser.add_argument('new_binary', type=Path)
args = parser.parse_args()

class Endpoint(BaseHTTPRequestHandler):
    def do_HEAD(self):
        self.send_response(204)
        self.end_headers()
    def log_message(self, *_args):
        pass

server = ThreadingHTTPServer(('127.0.0.1', 0), Endpoint)
thread = threading.Thread(target=server.serve_forever, daemon=True)
thread.start()
try:
    with tempfile.TemporaryDirectory() as directory:
        config = Path(directory) / 'released.json'
        config.write_text(json.dumps({'version': 'v1', 'targets': [{
            'name': 'released-http', 'type': 'http',
            'endpoint': f'http://127.0.0.1:{server.server_port}/health',
            'checks': ['protocol'], 'http': {'expected_status': [204]}}]}))
        for binary in [args.old_binary, args.new_binary, args.old_binary]:
            version = subprocess.check_output([str(binary.resolve()), 'version'], text=True).strip()
            run = subprocess.run([str(binary.resolve()), 'check', str(config), '--format', 'json'],
                                 capture_output=True, text=True, timeout=30)
            if run.returncode != 0 or json.loads(run.stdout)['status'] != 'HEALTHY':
                raise SystemExit(f'{version}: configuration compatibility failed')
            print(f'{version}: unchanged configuration HEALTHY')
finally:
    server.shutdown()
    server.server_close()
    thread.join()
