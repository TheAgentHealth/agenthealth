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
    workflow = yaml.safe_load((ROOT / '.github/workflows/publish.yml').read_text())
    job = workflow['jobs']['publish']
    assert job['permissions']['id-token'] == 'write'
    assert job['permissions']['contents'] == 'write'  # Draft asset access requires this.
    assert "github.ref == format('refs/tags/{0}', inputs.release_tag)" in job['if']
    assert "inputs.recovery && github.ref == 'refs/heads/main'" in job['if']
    assert not any('workflow_call' in str(k) for k in workflow)
    steps = '\n'.join(s.get('run', '') for s in job['steps'])
    assert '--provenance' in steps and 'npm audit signatures' in steps
    assert 'npm publish ./npm-dist/' in steps
    assert 'npm publish npm-dist/' not in steps
    assert 'NODE_AUTH_TOKEN' not in steps and 'NPM_TOKEN' not in steps
    release = yaml.safe_load((ROOT / '.github/workflows/release.yml').read_text())
    publish = release['jobs']['publish']
    npm = release['jobs']['npm']
    assert npm['permissions']['actions'] == 'write'
    assert set(npm['needs']) == {'release', 'container', 'packages', 'recover-assets'}
    assert 'npm' in publish['needs']
    assert "needs.npm.result == 'success'" in publish['if']
    assert any('wait_for_npm_publication.py' in s.get('run', '') for s in npm['steps'])
    assert not any('wait_for_npm_publication.py' in s.get('run', '') for s in publish['steps'])


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
