#!/usr/bin/env python3
"""Exercise unchanged authenticated/graph configurations on released/new/released CLI."""
import argparse
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import json
import os
from pathlib import Path
import secrets
import subprocess
import tempfile
import threading


def run_suite(old_binary, new_binary):
    token = secrets.token_hex(24)
    rejected_token = secrets.token_hex(24)
    state = {'outage': False}

    class Endpoint(BaseHTTPRequestHandler):
        def respond(self, body=False):
            status = 204
            if self.path == '/auth' and self.headers.get('Authorization') != 'Bearer ' + token:
                status = 401
            elif self.path == '/dependency' and state['outage']:
                status = 503
            document = {'version': 'v1', 'name': 'agent', 'live': True, 'ready': True,
                        'capabilities': ['answer'], 'dependencies': []}
            if self.path == '/agent':
                status = 200
            self.send_response(status)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            if body and self.path == '/agent':
                self.wfile.write(json.dumps(document).encode())
        def do_HEAD(self):
            self.respond()
        def do_GET(self):
            self.respond(True)
        def log_message(self, *_args):
            pass

    server = ThreadingHTTPServer(('127.0.0.1', 0), Endpoint)
    thread = threading.Thread(target=server.serve_forever, daemon=True)
    thread.start()
    base = f'http://127.0.0.1:{server.server_port}'
    def target(name, path='/health', **options):
        return {'name': name, 'type': 'http', 'endpoint': base + path,
                'checks': ['protocol'], **options}
    basic = {'version': 'v1', 'targets': [target('http')]}
    auth = {'version': 'v1', 'targets': [target('authenticated', '/auth',
            checks=['authentication', 'protocol'], auth={'bearer_env': 'AGENTHEALTH_UPGRADE_TOKEN'})]}
    nested = {'version': 'v1', 'targets': [target('root', dependencies=[
        target('support', '/dependency', critical=False)])]}
    graph = {'version': 'v1', 'targets': [
        target('direct', id='direct', dependencies=[{'ref': 'peer', 'relationship': 'downstream', 'critical': False},
                                                   {'ref': 'gateway', 'relationship': 'gateway'},
                                                   {'ref': 'router', 'relationship': 'router'}]),
        target('peer', '/dependency', id='peer'),
        target('gateway', id='gateway', type='gateway'),
        target('router', id='router', type='router')]}
    agent = {'version': 'v1', 'targets': [target('direct-agent', '/agent', type='agent',
             checks=['protocol', 'capability'], agent={'required_capabilities': ['answer']})]}
    # Expected roots preserve health status and edge metadata through both directions.
    cases = [('http', basic, False, token, 0),
             ('bearer-accepted', auth, False, token, 0),
             ('bearer-rejected', auth, False, rejected_token, 4),
             ('bearer-missing', auth, False, '', 4),
             ('nested-healthy', nested, False, token, 0),
             ('nested-optional-outage', nested, True, token, 1),
             ('graph-healthy', graph, False, token, 0),
             ('graph-outage', graph, True, token, 2),
             ('agent-contract', agent, False, token, 0)]

    def stable(value):
        if isinstance(value, dict):
            return {key: stable(child) for key, child in value.items()
                    if key not in ('latency_ms', 'message', 'steps')}
        if isinstance(value, list):
            return [stable(child) for child in value]
        return value

    try:
        with tempfile.TemporaryDirectory() as directory:
            config = Path(directory) / 'released.json'
            for name, document, outage, credential, exit_code in cases:
                state['outage'] = outage
                config.write_text(json.dumps(document))
                previous = None
                for binary in [old_binary, new_binary, old_binary]:
                    env = {**os.environ, 'AGENTHEALTH_UPGRADE_TOKEN': credential}
                    version = subprocess.check_output([str(binary.resolve()), 'version'], text=True).strip()
                    run = subprocess.run([str(binary.resolve()), 'check', str(config), '--format', 'json'],
                                         capture_output=True, text=True, env=env, timeout=30)
                    if any(value and (value in run.stdout or value in run.stderr)
                           for value in (token, credential)):
                        raise RuntimeError(f'{name}: credential disclosure')
                    if run.returncode != exit_code:
                        raise RuntimeError(f'{name}: expected exit {exit_code}, got {run.returncode}')
                    evidence = stable(json.loads(run.stdout))
                    if previous is not None and previous != evidence:
                        raise RuntimeError(f'{name}: health/graph evidence changed during upgrade or rollback')
                    previous = evidence
                    print(f'{version}: {name} exit={exit_code} evidence preserved', flush=True)
    finally:
        server.shutdown()
        server.server_close()
        thread.join()


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('old_binary', type=Path)
    parser.add_argument('new_binary', type=Path)
    args = parser.parse_args()
    run_suite(args.old_binary, args.new_binary)
