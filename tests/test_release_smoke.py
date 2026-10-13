"""Native smoke checks select and authenticate the correct platform archive."""
import hashlib
import importlib.util
import io
from pathlib import Path
import subprocess
import tarfile
import zipfile
import pytest

spec = importlib.util.spec_from_file_location('smoke', Path(__file__).parents[1] / 'scripts/test_release_binary.py')
smoke = importlib.util.module_from_spec(spec)
spec.loader.exec_module(smoke)


@pytest.mark.parametrize('system,machine,os_name,arch', [
    ('Linux', 'x86_64', 'linux', 'amd64'), ('Linux', 'aarch64', 'linux', 'arm64'),
    ('Darwin', 'x86_64', 'darwin', 'amd64'), ('Darwin', 'arm64', 'darwin', 'arm64'),
    ('Windows', 'AMD64', 'windows', 'amd64')])
def test_native_archive_selection_and_tampering(tmp_path, monkeypatch, system, machine, os_name, arch):
    monkeypatch.setattr(smoke.platform, 'system', lambda: system)
    monkeypatch.setattr(smoke.platform, 'machine', lambda: machine)
    name = 'agenthealth.exe' if system == 'Windows' else 'agenthealth'
    archive = tmp_path / f'agenthealth_v1.2.3_{os_name}_{arch}'
    archive = Path(str(archive) + ('.zip' if system == 'Windows' else '.tar.gz'))
    if system == 'Windows':
        with zipfile.ZipFile(archive, 'w') as bundle:
            bundle.writestr(name, b'binary')
    else:
        with tarfile.open(archive, 'w:gz') as bundle:
            entry = tarfile.TarInfo(name)
            entry.size = 6
            bundle.addfile(entry, io.BytesIO(b'binary'))
    (tmp_path / 'checksums.txt').write_text(f'{hashlib.sha256(archive.read_bytes()).hexdigest()}  {archive.name}\n')
    calls = []
    def version(command, **kwargs):
        assert Path(command[0]).read_bytes() == b'binary'
        calls.append(command)
        return 'agenthealth v1.2.3\n'
    def run(command, **kwargs):
        calls.append(command)
        return subprocess.CompletedProcess(command, 6 if command[1] == 'check' else 0)
    monkeypatch.setattr(smoke.subprocess, 'check_output', version)
    monkeypatch.setattr(smoke.subprocess, 'run', run)
    smoke.smoke('v1.2.3', tmp_path)
    assert [c[1] for c in calls] == ['version', '--help', 'check']
    calls.clear()
    archive.write_bytes(b'tampered')
    with pytest.raises(AssertionError):
        smoke.smoke('v1.2.3', tmp_path)
    assert not calls


def test_current_guide_cannot_point_at_a_different_release():
    guide = b'<!-- release-version: v0.13.1 -->\n/download/v0.13.1/agenthealth_v0.13.1_linux_amd64.tar.gz'
    smoke.verify_installation_guide(guide, 'v0.13.1')
    with pytest.raises(AssertionError):
        smoke.verify_installation_guide(guide, 'v0.13.2')
    with pytest.raises(AssertionError):
        smoke.verify_installation_guide(guide.replace(b'/download/v0.13.1/', b'/download/v0.12.0/'), 'v0.13.1')
    smoke.verify_installation_guide(b'historical unmarked guide', 'v0.4.0')
