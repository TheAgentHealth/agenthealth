"""Package manifests must bind to verified archive bytes."""
import hashlib
import importlib.util
import json
from pathlib import Path
import tarfile
import io
import pytest

spec = importlib.util.spec_from_file_location('packages', Path(__file__).parents[1] / 'scripts/package_release.py')
packages = importlib.util.module_from_spec(spec)
spec.loader.exec_module(packages)


def assets(path):
    names = []
    for system, arch in [('linux', 'amd64'), ('linux', 'arm64'), ('darwin', 'amd64'), ('darwin', 'arm64'), ('windows', 'amd64')]:
        name = f'agenthealth_v1.2.3_{system}_{arch}' + ('.zip' if system == 'windows' else '.tar.gz')
        source = path / name
        if system == 'linux':
            with tarfile.open(source, 'w:gz') as bundle:
                entry = tarfile.TarInfo('agenthealth')
                entry.size = 6
                bundle.addfile(entry, io.BytesIO(b'binary'))
        else:
            source.write_bytes(b'archive')
        names.append(source)
    (path / 'checksums.txt').write_text(''.join(f'{hashlib.sha256(p.read_bytes()).hexdigest()}  {p.name}\n' for p in names))


def test_manifests_and_packages(tmp_path, monkeypatch):
    assets(tmp_path)
    configs = []
    def package(command, **kwargs):
        configs.append(json.loads(Path(command[command.index('--config') + 1]).read_text()))
        Path(command[-1]).write_bytes(b'package')
    monkeypatch.setattr(packages.subprocess, 'run', package)
    packages.generate('v1.2.3', tmp_path)
    manifest = json.loads((tmp_path / 'agenthealth.json').read_text())
    assert manifest['bin'] == 'agenthealth.exe'
    assert manifest['architecture']['64bit']['hash'] == hashlib.sha256(b'archive').hexdigest()
    assert manifest['architecture']['64bit']['url'].endswith('agenthealth_v1.2.3_windows_amd64.zip')
    formula = (tmp_path / 'agenthealth.rb').read_text()
    assert 'on_macos do' in formula and 'on_linux do' in formula
    assert formula.count('sha256 ') == 4
    assert [c['arch'] for c in configs] == ['amd64', 'amd64', 'arm64', 'arm64']
    assert all(c['depends'] == ['ca-certificates'] for c in configs)
    for line in (tmp_path / 'checksums.txt').read_text().splitlines():
        digest, name = line.split('  ')
        assert hashlib.sha256((tmp_path / name).read_bytes()).hexdigest() == digest
    with pytest.raises(ValueError, match='already exist'):
        packages.generate('v1.2.3', tmp_path)


def test_tampered_archive_rejected(tmp_path):
    assets(tmp_path)
    (tmp_path / 'agenthealth_v1.2.3_windows_amd64.zip').write_bytes(b'tampered')
    with pytest.raises(ValueError, match='checksum mismatch'):
        packages.generate('v1.2.3', tmp_path)
    assert not (tmp_path / 'agenthealth.rb').exists()


@pytest.mark.parametrize('version', ['../bad', 'v1.2', 'v1.2.3;bad'])
def test_invalid_version(tmp_path, version):
    with pytest.raises(ValueError, match='invalid release version'):
        packages.generate(version, tmp_path)
