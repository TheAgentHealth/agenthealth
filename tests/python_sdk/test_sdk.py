"""SDK integration against the real CLI and bounded subprocess failure contracts."""
import asyncio
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import json
import os
from pathlib import Path
import subprocess
import sys
import threading
import time

import pytest

ROOT = Path(__file__).resolve().parents[2]
# Repository tests exercise source; the dedicated CI job installs built artifacts.
if not os.environ.get('AGENTHEALTH_INSTALLED_SDK'):
    sys.path.insert(0, str(ROOT / 'sdk/python/src'))
from agenthealth import (Client, Configuration, HealthStatus, Relationship, check,
                        load_config, BinaryNotFoundError, CompatibilityError,
                        ConfigurationError, ExecutionError, OutputError, CheckTimeoutError)
from agenthealth.models import parse_report
from agenthealth.version import __version__


@pytest.fixture(scope='session')
def binary(tmp_path_factory):
    supplied = os.environ.get('AGENTHEALTH_SDK_BINARY')
    if supplied:
        return supplied
    path = tmp_path_factory.mktemp('sdk-cli') / ('agenthealth.exe' if os.name == 'nt' else 'agenthealth')
    subprocess.run(['go', 'build', '-ldflags', f'-X main.version=v{__version__}', '-o', str(path),
                    './cmd/agenthealth'], cwd=ROOT, check=True)
    return str(path)


@pytest.fixture
def endpoint():
    state = {'status': 200, 'posts': 0}
    class Handler(BaseHTTPRequestHandler):
        def do_HEAD(self):
            self.send_response(401 if self.path == '/auth' and self.headers.get('Authorization') != 'Bearer sdk-secret' else state['status'])
            self.end_headers()
        def do_GET(self):
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps({'version': 'v1', 'name': 'first', 'live': True,
                'ready': True, 'capabilities': ['answer'], 'dependencies': []}).encode())
        def do_POST(self):
            state['posts'] += 1
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(b'{"completed":true,"success":true,"downstream":"peer"}')
        def log_message(self, *_):
            pass
    server = ThreadingHTTPServer(('127.0.0.1', 0), Handler)
    thread = threading.Thread(target=server.serve_forever, daemon=True)
    thread.start()
    try:
        yield f'http://127.0.0.1:{server.server_port}', state
    finally:
        server.shutdown()
        server.server_close()
        thread.join()


def test_sync_async_and_health_failures(binary, endpoint):
    url, state = endpoint
    client = Client(binary)
    report = client.check(target_type='http', endpoint=url)
    assert report.status == HealthStatus.HEALTHY and report.exit_code == 0
    assert report.result.target.type == 'http'
    assert asyncio.run(client.async_check(target_type='http', endpoint=url)).to_dict()['status'] == 'HEALTHY'
    state['status'] = 503
    failed = client.check(target_type='http', endpoint=url, checks=['protocol'])
    assert failed.status == HealthStatus.UNHEALTHY and failed.exit_code == 2
    async def in_loop():
        with pytest.raises(RuntimeError, match='await'):
            client.check(target_type='http', endpoint=url)
    asyncio.run(in_loop())


def test_graph_roles_paths_and_batch(binary, endpoint):
    url, state = endpoint
    config = {'version': 'v1', 'targets': [
        {'id': 'first', 'name': 'first', 'type': 'agent', 'endpoint': url,
         'agent': {'required_capabilities': ['answer']}, 'dependencies': [
             {'ref': 'peer', 'relationship': 'downstream', 'critical': False},
             {'ref': 'path', 'relationship': 'path'},
             {'ref': 'gateway', 'relationship': 'gateway'},
             {'ref': 'router', 'relationship': 'router'}]},
        {'id': 'peer', 'name': 'peer', 'type': 'agent', 'endpoint': url},
        {'id': 'path', 'name': 'path', 'type': 'agent', 'endpoint': url, 'checks': ['functional'],
         'agent': {'functional': {'safe': True, 'text': 'health', 'downstream': 'peer'}}},
        {'id': 'gateway', 'name': 'gateway', 'type': 'gateway', 'endpoint': url},
        {'id': 'router', 'name': 'router', 'type': 'router', 'endpoint': url}]}
    report = Client(binary).check(config)
    assert report.exit_code == 0, report.to_dict()
    root = report.results[0]
    assert root.target.id == 'first'
    assert [child.relationship for child in root.dependencies] == [Relationship.DOWNSTREAM, Relationship.PATH, Relationship.GATEWAY, Relationship.ROUTER]
    assert root.dependencies[0].critical is False
    assert root.dependencies[1].checks['functional'].status == HealthStatus.HEALTHY
    assert state['posts'] == 1  # shared node executes once, not once per projection
    with pytest.raises(ValueError):
        report.result
    raw = report.to_dict()
    raw['results'][0]['target']['name'] = 'changed'
    assert root.target.name == 'first'


def test_credentials_and_unsupported_adapter_are_health_results(binary, endpoint):
    url, _ = endpoint
    report = check(target_type='http', endpoint=url + '/auth', binary=binary,
                   checks=['authentication'], options={'auth': {'bearer_env': 'SDK_TOKEN'}},
                   environment={'SDK_TOKEN': 'sdk-secret'})
    assert report.status == HealthStatus.HEALTHY
    missing = check(target_type='http', endpoint=url + '/auth', binary=binary,
                    checks=['authentication'], options={'auth': {'bearer_env': 'SDK_TOKEN'}},
                    environment={'SDK_TOKEN': ''})
    assert missing.exit_code == 4
    assert 'sdk-secret' not in json.dumps(report.to_dict())
    assert check(target_type='model', endpoint=url, binary=binary).exit_code == 4


def test_loader_validation_and_go_graph_semantics(binary, tmp_path):
    path = tmp_path / 'config.yaml'
    path.write_text('version: v1\ntargets:\n  - name: service\n    type: http\n    endpoint: http://127.0.0.1:1\n')
    config = load_config(path)
    assert config.to_dict()['targets'][0]['type'] == 'http'
    data = config.to_dict()
    data['targets'][0]['name'] = 'mutated'
    assert config.to_dict()['targets'][0]['name'] == 'service'
    path.write_text('version: v1\nversion: bad\ntargets: []')
    with pytest.raises(ConfigurationError):
        load_config(path)
    with pytest.raises(ConfigurationError):
        Configuration({'password': 'secret-value'})
    with pytest.raises(BinaryNotFoundError):
        Client(str(tmp_path / 'absent')).check(config)
    document = config.to_dict()
    document['targets'][0].update(id='loop', dependencies=[{'ref': 'loop'}])
    with pytest.raises(ExecutionError):
        Client(binary).check(document)


def test_result_validation_unknown_fields_and_exit_contract():
    document = {'spec_version': 'v1', 'target': {'name': 'target', 'type': 'http'},
                'status': 'HEALTHY', 'checks': {'protocol': {'status': 'HEALTHY'}},
                'dependencies': [], 'future_evidence': {'vendor': 'retained'}}
    assert parse_report(json.dumps(document).encode(), 0).to_dict() == document
    for output, code in [(b'not json', 0), (json.dumps(document).encode(), 2),
                         (b'{"status":"HEALTHY"}', 0)]:
        with pytest.raises(OutputError):
            parse_report(output, code)
    document['spec_version'] = 'future'
    with pytest.raises(OutputError):
        parse_report(json.dumps(document).encode(), 0)


@pytest.mark.skipif(os.name != 'posix', reason='POSIX executable/process-group contract')
@pytest.mark.parametrize('mode,error', [('version', CompatibilityError), ('tool', ExecutionError),
    ('malformed', OutputError), ('large', OutputError), ('slow', CheckTimeoutError)])
def test_subprocess_failure_contract(tmp_path, mode, error):
    path = tmp_path / 'fake-cli'
    path.write_text(f'''#!{sys.executable}
import sys,time
if sys.argv[1] == 'version':
    print('agenthealth v{'0.0.1' if mode == 'version' else __version__}')
else:
    mode = {mode!r}
    if mode == 'tool':
        print('secret-value',file=sys.stderr); sys.exit(6)
    if mode == 'large': print('x' * (17 * 1024 * 1024))
    elif mode == 'slow': time.sleep(10)
    else: print('secret-value malformed')
''')
    path.chmod(0o700)
    with pytest.raises(error) as failure:
        Client(path, timeout=0.3 if mode == 'slow' else 5).check(target_type='http', endpoint='http://localhost')
    assert 'secret-value' not in str(failure.value)


@pytest.mark.skipif(os.name != 'posix', reason='POSIX descendant cleanup')
def test_async_cancellation_reaps_child_and_config(tmp_path):
    path = tmp_path / 'fake-cli'
    marker = tmp_path / 'marker.json'
    path.write_text(f'''#!{sys.executable}
import json,subprocess,sys,time
if sys.argv[1]=='version': print('agenthealth v{__version__}')
else:
    child=subprocess.Popen([sys.executable,'-c','import time;time.sleep(60)'])
    open({str(marker)!r},'w').write(json.dumps([child.pid,sys.argv[-1]]))
    time.sleep(60)
''')
    path.chmod(0o700)
    async def cancel():
        task = asyncio.create_task(Client(path).async_check(target_type='http', endpoint='http://localhost'))
        for _ in range(100):
            if marker.exists():
                break
            await asyncio.sleep(0.02)
        assert marker.exists()
        task.cancel()
        with pytest.raises(asyncio.CancelledError):
            await task
    asyncio.run(cancel())
    pid, config_path = json.loads(marker.read_text())
    assert not Path(config_path).exists()
    # Linux may retain an adopted zombie until its init reaps it; it is stopped.
    status = Path(f'/proc/{pid}/stat')
    if status.exists():
        assert status.read_text().split()[2] == 'Z'
    else:
        with pytest.raises(ProcessLookupError):
            os.kill(pid, 0)


def test_vendored_schemas_match_source():
    from importlib.resources import files
    for name in ('configuration', 'result'):
        assert files('agenthealth').joinpath('schemas', name + '.schema.json').read_bytes() == (ROOT / 'spec/schemas' / (name + '.schema.json')).read_bytes()


def test_upgrade_suite_rejects_stderr_credential_disclosure(binary, monkeypatch):
    import importlib
    monkeypatch.syspath_prepend(str(ROOT / 'scripts'))
    upgrade = importlib.import_module('test_config_upgrade')
    tokens = iter(('accepted-test-credential', 'rejected-test-credential'))
    monkeypatch.setattr(upgrade.secrets, 'token_hex', lambda _: next(tokens))
    original = subprocess.run
    def disclose(*args, **kwargs):
        result = original(*args, **kwargs)
        if kwargs.get('env', {}).get('AGENTHEALTH_UPGRADE_TOKEN') == 'rejected-test-credential':
            result.stderr += 'rejected-test-credential'
        return result
    monkeypatch.setattr(upgrade.subprocess, 'run', disclose)
    with pytest.raises(RuntimeError, match='credential disclosure'):
        upgrade.run_suite(Path(binary), Path(binary))


@pytest.mark.skipif(os.name != 'posix', reason='POSIX process groups')
def test_exited_group_permission_error_preserves_report(binary, endpoint, monkeypatch):
    def denied(*_):
        raise PermissionError('sensitive platform diagnostic')
    monkeypatch.setattr(os, 'killpg', denied)
    url, _ = endpoint
    assert Client(binary).check(target_type='http', endpoint=url).status == HealthStatus.HEALTHY


@pytest.mark.skipif(os.name != 'posix', reason='POSIX process groups')
def test_live_group_permission_error_is_sanitized(monkeypatch):
    from agenthealth.client import _signal_group
    class LiveProcess:
        pid = 123
        returncode = None
        async def wait(self):
            await asyncio.sleep(60)
    def denied(*_):
        raise PermissionError('sensitive platform diagnostic')
    monkeypatch.setattr(os, 'killpg', denied)
    with pytest.raises(ExecutionError, match='Unable to stop AgentHealth CLI process') as failure:
        asyncio.run(_signal_group(LiveProcess(), 15))
    assert 'sensitive' not in str(failure.value)
