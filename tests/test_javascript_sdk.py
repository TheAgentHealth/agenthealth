"""npm packaging and publication must preserve shared-version signed artifact identity."""
import hashlib
import importlib
import io
import json
from pathlib import Path
import shutil
import subprocess
import tarfile

import pytest
import yaml

ROOT = Path(__file__).parents[1]


@pytest.fixture
def package(monkeypatch):
    monkeypatch.syspath_prepend(str(ROOT / 'scripts'))
    return importlib.import_module('package_javascript')


def test_historical_assets_and_new_sdk_assets(package):
    assert not package.has_javascript('v0.13.2')
    assert package.has_javascript('v0.14.0')
    assert package.asset_names('v0.14.0') == {'theagenthealth-agenthealth-0.14.0.tgz'}
    prepare = importlib.import_module('prepare_synchronized_release')
    assert len(prepare.asset_names('v0.14.0')) == 21
    assert len(prepare.asset_names('v0.13.2')) == 20


def test_version_mismatch_rejected_before_pack(package, tmp_path, monkeypatch):
    monkeypatch.setattr(package.subprocess, 'run', lambda *a, **kw: pytest.fail('must not pack'))
    with pytest.raises(ValueError, match='source version'):
        package.build_javascript('v9.99.99', tmp_path)


def test_npm_archive_is_reproducible_and_has_only_public_source(package, tmp_path):
    if not shutil.which('npm'):
        pytest.skip('npm required for package build')
    version = json.loads((ROOT / 'sdk/javascript/package.json').read_text())['version']
    outputs = [tmp_path / str(n) for n in range(2)]
    for directory in outputs:
        package.build_javascript('v' + version, directory)
    name, = package.asset_names('v' + version)
    assert (outputs[0] / name).read_bytes() == (outputs[1] / name).read_bytes()
    with tarfile.open(outputs[0] / name) as archive:
        assert len(archive.getnames()) == 7
        assert all(m.uid == 0 and m.mode == 0o644 for m in archive.getmembers())
        assert not any('node_modules' in m.name or '/test/' in m.name for m in archive.getmembers())


def test_trusted_publisher_and_final_release_gating():
    release = yaml.safe_load((ROOT / '.github/workflows/release.yml').read_text())
    npm = release['jobs']['npm']
    assert npm['permissions'] == {'contents': 'write', 'id-token': 'write'}
    assert npm['environment']['name'] == 'npm'
    steps = '\n'.join(s.get('run', '') for s in npm['steps'])
    assert '--provenance' in steps and 'npm audit signatures' in steps
    assert 'npm publish ./npm-dist/' in steps
    assert 'npm publish npm-dist/' not in steps
    assert 'unset NODE_AUTH_TOKEN NPM_TOKEN' in steps
    assert 'verify_npm_oidc.py' in steps
    assert not (ROOT / '.github/workflows/publish.yml').exists()
    publish = release['jobs']['publish']
    assert set(npm['needs']) == {'release', 'container', 'packages', 'recover-assets'}
    assert 'npm' in publish['needs']
    assert "needs.npm.result == 'success'" in publish['if']


def test_registry_rejects_modified_tarball_before_accepting_integrity(tmp_path, monkeypatch):
    monkeypatch.syspath_prepend(str(ROOT / 'scripts'))
    helper = importlib.import_module('verify_npm_publication')
    data = io.BytesIO()
    with tarfile.open(fileobj=data, mode='w:gz') as archive:
        document = json.dumps({'name': '@theagenthealth/agenthealth', 'version': '0.14.0'}).encode()
        info = tarfile.TarInfo('package/package.json'); info.size = len(document)
        archive.addfile(info, io.BytesIO(document))
    content = data.getvalue()
    name = 'theagenthealth-agenthealth-0.14.0.tgz'
    (tmp_path / name).write_bytes(content)
    (tmp_path / 'checksums.txt').write_text(hashlib.sha256(content).hexdigest() + '  ' + name + '\n')
    def fetch(url):
        if url.startswith('https://registry.npmjs.org/'):
            return json.dumps({'dist': {'tarball': 'https://example.com/tarball'}}).encode()
        return b'tampered'
    monkeypatch.setattr(helper, 'fetch', fetch)
    with pytest.raises(ValueError, match='bytes differ'):
        helper.verify('v0.14.0', tmp_path, registry=True)


@pytest.mark.parametrize('conclusion', ['skipped', 'failure', None, 'missing'])
def test_release_gate_requires_every_node_matrix_job(monkeypatch, conclusion):
    monkeypatch.syspath_prepend(str(ROOT / 'scripts'))
    gate = importlib.import_module('wait_for_release_ci')
    run = dict(id=10, head_sha='abc', head_branch='main', event='workflow_dispatch',
               status='completed', conclusion='success', html_url='https://example.com/run', run_attempt=1)
    jobs = [dict(name=f'Binary distribution ({p})', conclusion='success') for p in gate.PLATFORMS]
    jobs += [dict(name=f'JavaScript SDK ({p}, Node {v})', conclusion='success')
             for p in ('ubuntu-latest', 'macos-latest', 'windows-latest') for v in ('22', '24')]
    if conclusion == 'missing': jobs.pop()
    else: jobs[-1]['conclusion'] = conclusion
    monkeypatch.setattr(gate, 'api', lambda path: {'jobs': jobs} if '/jobs?' in path else {'workflow_runs': [run]})
    with pytest.raises(RuntimeError, match='JavaScript SDK'):
        gate.check_ci('owner/repo', 'abc', require_javascript=True)


def test_oidc_diagnostics_validate_exchange_without_printing_tokens(monkeypatch, capsys):
    import base64
    monkeypatch.syspath_prepend(str(ROOT / 'scripts'))
    helper = importlib.import_module('verify_npm_oidc')
    claims = {'repository': 'TheAgentHealth/agenthealth', 'workflow_ref': 'TheAgentHealth/agenthealth/.github/workflows/release.yml@refs/heads/main', 'private_claim': 'do-not-print'}
    payload = base64.urlsafe_b64encode(json.dumps(claims).encode()).decode().rstrip('=')
    token = 'header.' + payload + '.signature'
    monkeypatch.setenv('ACTIONS_ID_TOKEN_REQUEST_URL', 'https://example.com/token?api-version=2')
    monkeypatch.setenv('ACTIONS_ID_TOKEN_REQUEST_TOKEN', 'request-secret')
    calls = []
    def fetch(request, timeout):
        calls.append(request)
        if len(calls) == 1:
            assert 'audience=npm%3Aregistry.npmjs.org' in request.full_url
            assert request.get_header('Authorization') == 'Bearer request-secret'
            return io.BytesIO(json.dumps({'value': token}).encode())
        assert request.get_method() == 'POST'
        assert request.full_url.endswith('/package/@theagenthealth%2fagenthealth')
        assert request.get_header('Authorization') == 'Bearer ' + token
        return io.BytesIO(b'{"token":"publishing-secret"}')
    monkeypatch.setattr(helper.urllib.request, 'urlopen', fetch)
    helper.verify()
    output = capsys.readouterr().out
    assert 'TheAgentHealth/agenthealth' in output
    assert all(secret not in output for secret in (token, 'request-secret', 'publishing-secret', 'do-not-print'))


def test_oidc_rejection_is_reported_without_falling_back_to_token_auth(monkeypatch):
    import base64
    import urllib.error
    monkeypatch.syspath_prepend(str(ROOT / 'scripts'))
    helper = importlib.import_module('verify_npm_oidc')
    token = 'header.' + base64.urlsafe_b64encode(b'{}').decode().rstrip('=') + '.signature'
    monkeypatch.setenv('ACTIONS_ID_TOKEN_REQUEST_URL', 'https://example.com/token')
    monkeypatch.setenv('ACTIONS_ID_TOKEN_REQUEST_TOKEN', 'request-secret')
    def fetch(request, timeout):
        if request.full_url.startswith('https://example.com/'):
            return io.BytesIO(json.dumps({'value': token}).encode())
        raise urllib.error.HTTPError(request.full_url, 404, 'Not found', {},
                                     io.BytesIO(json.dumps({'message': 'rejected ' + token + ' request-secret'}).encode()))
    monkeypatch.setattr(helper.urllib.request, 'urlopen', fetch)
    with pytest.raises(RuntimeError, match='HTTP 404') as error:
        helper.verify()
    assert token not in str(error.value) and 'request-secret' not in str(error.value)


@pytest.mark.parametrize('tampered', [False, True])
def test_pypi_recovery_reuses_only_identical_signed_distributions(monkeypatch, tmp_path, tampered):
    monkeypatch.syspath_prepend(str(ROOT / 'scripts'))
    helper = importlib.import_module('verify_pypi_publication')
    names = {'agenthealth-0.14.0-py3-none-any.whl', 'agenthealth-0.14.0.tar.gz'}
    for name in names:
        (tmp_path / name).write_bytes(name.encode())
    metadata = {'urls': [{'filename': name, 'url': 'https://files.pythonhosted.org/' + name,
                          'digests': {'sha256': hashlib.sha256(name.encode()).hexdigest()}, 'yanked': False}
                         for name in sorted(names)]}
    def fetch(url, timeout):
        if url.startswith('https://pypi.org/'):
            return io.BytesIO(json.dumps(metadata).encode())
        return io.BytesIO(b'tampered' if tampered else url.rsplit('/', 1)[1].encode())
    monkeypatch.setattr(helper.urllib.request, 'urlopen', fetch)
    if tampered:
        with pytest.raises(ValueError, match='differs from signed'):
            helper.verify('v0.14.0', tmp_path)
    else:
        assert helper.verify('v0.14.0', tmp_path)


@pytest.mark.parametrize('status', [404, 429, 503])
def test_registry_waits_for_asynchronous_publication(monkeypatch, status):
    import urllib.error
    monkeypatch.syspath_prepend(str(ROOT / 'scripts'))
    helper = importlib.import_module('verify_npm_publication')
    calls = []
    def fetch(url):
        calls.append(url)
        if len(calls) == 1:
            raise urllib.error.HTTPError(url, status, 'processing', {}, None)
        return b'{"version":"0.14.0"}'
    monkeypatch.setattr(helper, 'fetch', fetch)
    monkeypatch.setattr(helper.time, 'sleep', lambda _: None)
    assert helper.registry_document('0.14.0', wait=True)['version'] == '0.14.0'
    assert len(calls) == 2


def test_registry_wait_is_bounded_and_auth_errors_fail_immediately(monkeypatch):
    import urllib.error
    monkeypatch.syspath_prepend(str(ROOT / 'scripts'))
    helper = importlib.import_module('verify_npm_publication')
    def unavailable(url):
        raise urllib.error.HTTPError(url, 404, 'processing', {}, None)
    monkeypatch.setattr(helper, 'fetch', unavailable)
    with pytest.raises(TimeoutError, match='deadline'):
        helper.registry_document('0.14.0', wait=True, timeout=0)
    with pytest.raises(urllib.error.HTTPError):
        helper.registry_document('0.14.0')
    def forbidden(url):
        raise urllib.error.HTTPError(url, 403, 'forbidden', {}, None)
    monkeypatch.setattr(helper, 'fetch', forbidden)
    with pytest.raises(urllib.error.HTTPError):
        helper.registry_document('0.14.0', wait=True)
