"""Release coordination must preserve complete, immutable and publicly verifiable artifacts."""
import copy
import hashlib
import importlib
import io
import json
from pathlib import Path
import tarfile

import pytest
import yaml

ROOT = Path(__file__).parents[1]
TAG = 'v0.12.0'
REVISION = 'a' * 40


@pytest.fixture
def helpers(monkeypatch):
    monkeypatch.syspath_prepend(str(ROOT / 'scripts'))
    return tuple(importlib.import_module(name) for name in
                 ('package_helm', 'prepare_synchronized_release', 'publish_oci'))


def cli_assets(directory, prepare):
    directory.mkdir()
    for name in prepare.asset_names(TAG) - {'checksums.txt', 'agenthealth-0.12.0.tgz'}:
        (directory / name).write_bytes(name.encode())
    (directory / 'checksums.txt').write_text(''.join(
        f'{hashlib.sha256(p.read_bytes()).hexdigest()}  {p.name}\n' for p in sorted(directory.iterdir())))


def test_complete_asset_checksum_set_and_no_stale_files(tmp_path, helpers, monkeypatch):
    _, prepare, publisher = helpers
    output = tmp_path / 'assets'
    cli_assets(output, prepare)
    def chart(tag, directory, lint):
        assert tag == TAG and not lint
        path = directory / 'agenthealth-0.12.0.tgz'
        path.write_bytes(b'packaged-chart')
        return path
    monkeypatch.setattr(prepare, 'package_chart', chart)
    prepare.prepare(TAG, output)
    publisher.validate_assets(TAG, output)
    assert len(list(output.iterdir())) == 18
    (output / 'unrelated.txt').write_text('stale')
    with pytest.raises(ValueError, match='unexpected'):
        publisher.validate_assets(TAG, output)


@pytest.mark.parametrize('damage', ['missing', 'duplicate', 'tampered'])
def test_preparation_rejects_incomplete_or_tampered_cli_assets(tmp_path, helpers, damage):
    _, prepare, _ = helpers
    output = tmp_path / 'assets'
    cli_assets(output, prepare)
    manifest = output / 'checksums.txt'
    lines = manifest.read_text().splitlines(True)
    if damage == 'missing':
        manifest.write_text(''.join(lines[1:]))
    elif damage == 'duplicate':
        manifest.write_text(''.join(lines + lines[:1]))
    else:
        (output / 'agenthealth.rb').write_text('changed-after-signing')
    with pytest.raises(ValueError):
        prepare.prepare(TAG, output)
    assert not (output / 'agenthealth-0.12.0.tgz').exists()


def test_chart_rebuild_is_deterministic_and_contains_source_binding(tmp_path, helpers):
    helm, _, _ = helpers
    results = []
    for index in range(2):
        package = tmp_path / f'chart-{index}.tgz'
        with tarfile.open(package, 'w:gz') as archive:
            for name, data in [('agenthealth/Chart.yaml', b'name: agenthealth\nversion: 0.12.0\n'),
                               ('agenthealth/templates/workload.yaml', b'kind: Job\n')]:
                member = tarfile.TarInfo(name)
                member.size = len(data)
                member.mtime = 1000 + index
                member.uid = index + 100
                archive.addfile(member, io.BytesIO(data))
        helm.normalize(package, 12345, REVISION)
        results.append(package.read_bytes())
        with tarfile.open(package) as archive:
            metadata = yaml.safe_load(archive.extractfile('agenthealth/Chart.yaml'))
            assert metadata['annotations']['org.opencontainers.image.revision'] == REVISION
            assert all(m.uid == 0 and m.gid == 0 and m.mtime == 12345 for m in archive.getmembers())
    assert results[0] == results[1]


@pytest.mark.parametrize('tag,app,image', [('v0.12.1', TAG, TAG),
                                         (TAG, 'v0.11.1', TAG),
                                         (TAG, TAG, 'v0.11.1')])
def test_mixed_component_versions_fail_before_packaging(tmp_path, helpers, monkeypatch, tag, app, image):
    helm, _, _ = helpers
    root = tmp_path / 'source'
    chart = root / 'deploy/helm/agenthealth'
    chart.mkdir(parents=True)
    (chart / 'Chart.yaml').write_text(yaml.safe_dump({'version': '0.12.0', 'appVersion': app}))
    (chart / 'values.yaml').write_text(yaml.safe_dump({'image': {'tag': image}}))
    monkeypatch.setattr(helm, 'ROOT', root)
    with pytest.raises(ValueError):
        helm.package_chart(tag, tmp_path / 'output')
    assert not (tmp_path / 'output').exists()


def binary_manifest(publisher):
    return {'artifactType': publisher.BINARY_TYPE, 'annotations': {
        'org.opencontainers.image.source': publisher.SOURCE,
        'org.opencontainers.image.revision': REVISION,
        'org.opencontainers.image.version': TAG}, 'layers': [{
            'digest': 'sha256:' + 'b' * 64,
            'annotations': {'org.opencontainers.image.title': 'archive.tar.gz'}}]}


@pytest.mark.parametrize('damage', ['version', 'revision', 'source', 'blob', 'duplicate', 'type'])
def test_existing_oci_versions_cannot_be_replaced_with_different_content(helpers, damage):
    _, _, publisher = helpers
    document = binary_manifest(publisher)
    expected = {'archive.tar.gz': 'sha256:' + 'b' * 64}
    publisher.validate_manifest(document, expected, TAG, REVISION, binary=True)
    changed = copy.deepcopy(document)
    if damage in ('version', 'revision', 'source'):
        changed['annotations']['org.opencontainers.image.' + damage] = 'different'
    elif damage == 'blob':
        changed['layers'][0]['digest'] = 'sha256:' + 'c' * 64
    elif damage == 'duplicate':
        changed['layers'] += changed['layers'][:1]
    else:
        changed['artifactType'] = 'application/other'
    with pytest.raises(ValueError):
        publisher.validate_manifest(changed, expected, TAG, REVISION, binary=True)


@pytest.mark.parametrize('tampered', [False, True])
def test_public_blob_checks_use_anonymous_credentials(tmp_path, helpers, monkeypatch, tampered):
    _, prepare, publisher = helpers
    output = tmp_path / 'assets'
    cli_assets(output, prepare)
    chart = output / 'agenthealth-0.12.0.tgz'
    chart.write_bytes(b'chart')
    (output / 'checksums.txt').write_text(''.join(
        f'{hashlib.sha256(p.read_bytes()).hexdigest()}  {p.name}\n'
        for p in sorted(output.iterdir()) if p.name != 'checksums.txt'))
    blob_digest = 'sha256:' + hashlib.sha256(chart.read_bytes()).hexdigest()
    document = {'annotations': {'org.opencontainers.image.source': publisher.SOURCE,
                               'org.opencontainers.image.revision': REVISION,
                               'org.opencontainers.image.version': '0.12.0'},
                'layers': [{'mediaType': 'application/vnd.cncf.helm.chart.content.v1.tar+gzip',
                            'digest': blob_digest}]}
    config = json.dumps({'name': 'agenthealth', 'version': '0.12.0', 'appVersion': TAG}).encode()
    config_digest = 'sha256:' + hashlib.sha256(config).hexdigest()
    document['config'] = {'digest': config_digest}
    raw = json.dumps(document).encode()
    digest = 'sha256:' + hashlib.sha256(raw).hexdigest()
    report = tmp_path / 'report.json'
    report.write_text(json.dumps({'tag': TAG, 'source_commit': REVISION,
        'packages': [{'reference': publisher.CHART + ':0.12.0', 'digest': digest, 'binary': False}]}))
    monkeypatch.setenv('GH_TOKEN', 'secret-fixture')
    def fetch(request, timeout):
        if isinstance(request, str):
            assert 'ghcr.io/token?' in request
            return io.BytesIO(b'{"token":"anonymous-fixture"}')
        assert request.get_header('Authorization') == 'Bearer anonymous-fixture'
        assert 'secret-fixture' not in str(request.headers)
        if '/manifests/' in request.full_url:
            return io.BytesIO(raw)
        if config_digest in request.full_url:
            return io.BytesIO(config)
        return io.BytesIO(b'corrupted-blob' if tampered else chart.read_bytes())
    monkeypatch.setattr(publisher.urllib.request, 'urlopen', fetch)
    if tampered:
        with pytest.raises(ValueError, match="blob checksum"):
            publisher.public_verification(report, output)
    else:
        publisher.public_verification(report, output)


def test_public_release_waits_for_all_required_publication_jobs():
    workflow = yaml.safe_load((ROOT / '.github/workflows/release.yml').read_text())
    publish = workflow['jobs']['publish']
    assert set(publish['needs']) == {'release', 'container', 'packages'}
    assert "needs.packages.result == 'success'" in publish['if']
    assert "needs.release.outputs.synchronized != 'true'" in publish['if']
    stage = next(s for s in workflow['jobs']['release']['steps'] if s.get('name') == 'Stage draft release')
    assert '--draft' in stage['run'] and '--draft=false' not in stage['run']
