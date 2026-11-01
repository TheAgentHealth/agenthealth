"""Phase 21 runner boundaries, report safety and verified installation."""
import importlib.util
import os
from pathlib import Path
import subprocess
import sys

import pytest

ROOT = Path(__file__).resolve().parents[1]


def module(path):
    spec = importlib.util.spec_from_file_location('ci_helper', ROOT / path)
    result = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(result)
    return result


@pytest.mark.parametrize('code', range(7))
def test_action_preserves_health_exit_and_removes_tool_report(tmp_path, monkeypatch, code):
    gate = module('integrations/github-action/gate.py')
    for key, value in {'GITHUB_WORKSPACE': str(tmp_path), 'GITHUB_OUTPUT': str(tmp_path / 'outputs'),
                       'AH_REPORT': 'report.xml', 'AH_CONFIGURATION': 'configuration.yaml',
                       'AH_BINARY': 'trusted-binary', 'AH_VERSION': 'v0.19.0', 'AH_TIMEOUT': '120'}.items():
        monkeypatch.setenv(key, value)
    calls = []

    def run(arguments, **options):
        calls.append(arguments)
        if arguments[1] == 'version':
            return subprocess.CompletedProcess(arguments, 0, b'agenthealth v0.19.0\n')
        options['stdout'].write(b'<testsuite/>')
        return subprocess.CompletedProcess(arguments, code)
    monkeypatch.setattr(gate.subprocess, 'run', run)
    assert gate.main() == code
    assert (tmp_path / 'outputs').read_text() == f'exit-code={code}\n'
    assert (tmp_path / 'report.xml').exists() == (code != 6)
    assert calls[-1] == ['trusted-binary', 'check', str(tmp_path / 'configuration.yaml'), '--format', 'junit']


@pytest.mark.parametrize('failure', ['timeout', 'version', 'outside', 'overlap', 'bad-timeout'])
def test_action_fails_closed(tmp_path, monkeypatch, failure):
    gate = module('integrations/github-action/gate.py')
    config = tmp_path / 'configuration.yaml'
    config.write_text('keep configuration')
    report = tmp_path / 'report.xml'
    report.write_text('stale pass')
    values = {'GITHUB_WORKSPACE': str(tmp_path), 'GITHUB_OUTPUT': str(tmp_path / 'outputs'),
              'AH_REPORT': 'report.xml', 'AH_CONFIGURATION': 'configuration.yaml',
              'AH_BINARY': 'trusted', 'AH_VERSION': 'v0.19.0', 'AH_TIMEOUT': '120'}
    if failure == 'outside':
        values['AH_REPORT'] = '../outside.xml'
    if failure == 'overlap':
        values['AH_REPORT'] = 'configuration.yaml'
    if failure == 'bad-timeout':
        values['AH_TIMEOUT'] = '0'
    for key, value in values.items():
        monkeypatch.setenv(key, value)

    def run(arguments, **options):
        if failure == 'version':
            return subprocess.CompletedProcess(arguments, 0, b'agenthealth v0.18.0\n')
        if arguments[1] == 'version':
            return subprocess.CompletedProcess(arguments, 0, b'agenthealth v0.19.0\n')
        raise subprocess.TimeoutExpired(arguments, 120)
    monkeypatch.setattr(gate.subprocess, 'run', run)
    assert gate.main() == 6
    assert config.read_text() == 'keep configuration'
    if failure not in ('overlap', 'outside'):
        assert not report.exists()


def test_installer_rejects_mutable_version(tmp_path):
    installer = module('scripts/install_ci_binary.py')
    with pytest.raises(ValueError):
        installer.install('latest', tmp_path / 'binary')
    assert not (tmp_path / 'binary').exists()

@pytest.mark.parametrize('damage', [None, 'checksum', 'provenance'])
def test_installer_verifies_before_writing(tmp_path, monkeypatch, damage):
    import hashlib
    import io
    import tarfile
    installer = module('scripts/install_ci_binary.py')
    monkeypatch.setattr(installer.platform, 'system', lambda: 'Linux')
    monkeypatch.setattr(installer.platform, 'machine', lambda: 'x86_64')
    calls = []

    def run(arguments, **options):
        calls.append(arguments)
        if arguments[1:3] == ['release', 'download']:
            directory = Path(arguments[-1])
            archive = directory / 'agenthealth_v0.19.0_linux_amd64.tar.gz'
            with tarfile.open(archive, 'w:gz') as bundle:
                entry = tarfile.TarInfo('agenthealth')
                entry.size = len(b'verified-cli')
                bundle.addfile(entry, io.BytesIO(b'verified-cli'))
            digest = hashlib.sha256(archive.read_bytes()).hexdigest()
            if damage == 'checksum':
                digest = '0' * 64
            (directory / 'checksums.txt').write_text(f'{digest}  {archive.name}\n')
        elif damage == 'provenance':
            raise subprocess.CalledProcessError(1, arguments)
        return subprocess.CompletedProcess(arguments, 0)
    monkeypatch.setattr(installer.subprocess, 'run', run)
    output = tmp_path / 'agenthealth'
    if damage:
        with pytest.raises((ValueError, subprocess.CalledProcessError)):
            installer.install('v0.19.0', output)
        assert not output.exists()
    else:
        installer.install('v0.19.0', output)
        assert output.read_bytes() == b'verified-cli'
        assert '--source-ref' in calls[1]
        assert 'refs/tags/v0.19.0' in calls[1]
        assert '--signer-workflow' in calls[1]
        assert '--deny-self-hosted-runners' in calls[1]
        with pytest.raises(FileExistsError):
            installer.install('v0.19.0', output)
