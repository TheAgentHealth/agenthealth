#!/usr/bin/env python3
"""Verify and run the native executable from its release archive."""
import argparse
import hashlib
import platform
import subprocess
import tarfile
import tempfile
import zipfile
from pathlib import Path


def smoke(version, output):
    system = {'Linux': 'linux', 'Darwin': 'darwin', 'Windows': 'windows'}[platform.system()]
    arch = {'x86_64': 'amd64', 'AMD64': 'amd64', 'aarch64': 'arm64', 'arm64': 'arm64'}[platform.machine()]
    suffix = '.zip' if system == 'windows' else '.tar.gz'
    archive = output / f'agenthealth_{version}_{system}_{arch}{suffix}'
    hashes = dict(line.split('  ', 1)[::-1] for line in (output / 'checksums.txt').read_text().splitlines())
    assert hashlib.sha256(archive.read_bytes()).hexdigest() == hashes[archive.name]
    with tempfile.TemporaryDirectory() as temporary:
        binary = Path(temporary) / ('agenthealth.exe' if system == 'windows' else 'agenthealth')
        if system == 'windows':
            with zipfile.ZipFile(archive) as bundle:
                binary.write_bytes(bundle.read(binary.name))
        else:
            with tarfile.open(archive) as bundle:
                binary.write_bytes(bundle.extractfile('agenthealth').read())
            binary.chmod(0o755)
        assert subprocess.check_output([str(binary), 'version'], text=True).strip() == f'agenthealth {version}'
        subprocess.run([str(binary), '--help'], check=True, capture_output=True)
        result = subprocess.run([str(binary), 'check', str(Path(temporary) / 'missing.yaml')], capture_output=True)
        assert result.returncode == 6, result.returncode


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('version')
    parser.add_argument('--output', type=Path, default=Path('dist'))
    args = parser.parse_args()
    smoke(args.version, args.output)
