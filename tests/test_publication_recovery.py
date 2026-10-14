"""Recovery must reuse existing signed artifacts and cannot rebuild version tags."""
import importlib
import json
from pathlib import Path

import pytest
import yaml

ROOT = Path(__file__).parents[1]
COMMIT = '9' * 40


@pytest.fixture
def helper(monkeypatch):
    monkeypatch.syspath_prepend(str(ROOT / 'scripts'))
    return importlib.import_module('recover_publication')


def test_recovery_is_opt_in_and_skips_artifact_production():
    workflow = yaml.safe_load((ROOT / '.github/workflows/release.yml').read_text())
    jobs = workflow['jobs']
    assert '!inputs.publication_only' in jobs['validate']['if']
    for name in ('release', 'container'):
        assert 'validate' in jobs[name]['needs']
    assert jobs['recover-assets']['needs'] == 'ci-gate'
    assert 'workflow_dispatch' in jobs['recover-assets']['if']
    assert 'inputs.publication_only' in jobs['recover-assets']['if']
    assert jobs['recover-assets']['permissions'] == {'contents': 'read'}
    assert jobs['recover-pypi']['permissions'] == {'contents': 'read', 'id-token': 'write'}
    assert jobs['recover-pypi']['environment']['name'] == 'pypi'
    assert jobs['recover-pypi']['needs'] == 'recover-assets'
    assert 'recover-pypi' in jobs['publish']['needs']
    assert "needs.recover-assets.result == 'success'" in jobs['publish']['if']
    assert "needs.recover-pypi.result == 'success'" in jobs['publish']['if']
    for job in ('pypi', 'recover-pypi'):
        action = next(step for step in jobs[job]['steps'] if step.get('uses', '').startswith('pypa/'))
        assert action['uses'].endswith('@ed0c53931b1dc9bd32cbe73a98c7f6766f8a527e')
        assert action['with']['attestations'] is True


@pytest.mark.parametrize('tag', ['latest', '../v0.13.1', 'v0.13.1;command'])
def test_recovery_rejects_untrusted_tag_before_git(helper, tmp_path, monkeypatch, tag):
    monkeypatch.setattr(helper.subprocess, 'run', lambda *_a, **_kw: pytest.fail('must not execute'))
    with pytest.raises(ValueError, match='immutable software version'):
        helper.prepare(tag, tmp_path / 'out')


def test_recovery_refuses_stale_files_and_public_release(helper, tmp_path, monkeypatch):
    output = tmp_path / 'out'
    output.mkdir()
    (output / 'stale').write_text('old')
    with pytest.raises(ValueError, match='must be empty'):
        helper.prepare('v0.13.1', output)
    calls = []
    monkeypatch.setattr(helper.subprocess, 'run', lambda arguments, **_kw: calls.append(arguments))
    def command(arguments):
        if arguments[0] == 'git':
            return COMMIT
        return json.dumps({'isDraft': False, 'tagName': 'v0.13.1', 'assets': []})
    monkeypatch.setattr(helper, 'command', command)
    with pytest.raises(ValueError, match='existing draft'):
        helper.prepare('v0.13.1', tmp_path / 'fresh')
    assert calls == [['git', 'fetch', 'origin', 'refs/tags/v0.13.1']]


@pytest.mark.parametrize('damage', [None, 'revision', 'platform', 'duplicate'])
def test_existing_images_require_both_platforms_and_source(helper, monkeypatch, damage):
    manifest = {'manifests': [{'platform': {'os': 'linux', 'architecture': arch}, 'digest': arch}
                             for arch in ('amd64', 'arm64')]}
    if damage == 'duplicate':
        manifest['manifests'].append(manifest['manifests'][0])
    def get(kind, value):
        if kind == 'manifests':
            return {'config': {'digest': value}, 'layers': [{'digest': 'layer-' + value}]}, 'digest'
        labels = {'org.opencontainers.image.source': 'https://github.com/TheAgentHealth/agenthealth',
                  'org.opencontainers.image.version': 'v0.13.1',
                  'org.opencontainers.image.revision': 'wrong' if damage == 'revision' else COMMIT}
        return {'architecture': 'wrong' if damage == 'platform' else value, 'os': 'linux',
                'rootfs': {'diff_ids': ['layer-' + value]},
                'config': {'User': '65532', 'Entrypoint': ['agenthealth'], 'Labels': labels}}, 'digest'
    monkeypatch.setattr(helper, 'registry', lambda *_a: (manifest, 'sha256:' + 'a' * 64, get))
    if damage:
        with pytest.raises(ValueError):
            helper.image('host', 'repo', 'token-url', 'v0.13.1', COMMIT)
    else:
        digest, identities = helper.image('host', 'repo', 'token-url', 'v0.13.1', COMMIT)
        assert set(identities) == {'amd64', 'arm64'}
        assert digest.startswith('sha256:')


def test_registry_digest_tampering_fails_before_publication(helper, monkeypatch):
    def fetch(url, headers=None):
        return b'{"token":"anonymous-token"}' if url == 'token-url' else b'{}'
    monkeypatch.setattr(helper, 'fetch', fetch)
    _, _, get = helper.registry('host', 'repo', 'v0.13.1', 'token-url')
    with pytest.raises(ValueError, match='digest mismatch'):
        get('blobs', 'sha256:' + '0' * 64)
