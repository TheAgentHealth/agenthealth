#!/usr/bin/env python3
"""Install an exact CLI release after checksum and GitHub provenance verification.

Requires authenticated gh with release download and attestation verify support.
No installation occurs until verification succeeds. Output must not exist.
"""
import argparse
import hashlib
import io
from pathlib import Path
import platform
import re
import subprocess
import tarfile
import tempfile
import zipfile

REPOSITORY = 'TheAgentHealth/agenthealth'


def install(version, output):
    if not re.fullmatch(r'v[0-9]+\.[0-9]+\.[0-9]+', version):
        raise ValueError('Expected exact stable vX.Y.Z')
    system = {'Linux': 'linux', 'Darwin': 'darwin', 'Windows': 'windows'}[platform.system()]
    arch = {'x86_64': 'amd64', 'AMD64': 'amd64', 'aarch64': 'arm64', 'arm64': 'arm64'}[platform.machine()]
    if system == 'windows' and arch != 'amd64':
        raise ValueError('Unsupported platform')
    extension = 'zip' if system == 'windows' else 'tar.gz'
    name = f'agenthealth_{version}_{system}_{arch}.{extension}'
    with tempfile.TemporaryDirectory(prefix='agenthealth-ci-') as temporary:
        subprocess.run(['gh', 'release', 'download', version, '--repo', REPOSITORY,
                        '--pattern', name, '--pattern', 'checksums.txt', '--dir', temporary], check=True, timeout=120)
        archive = Path(temporary) / name
        expected = next(line.split()[0] for line in (Path(temporary) / 'checksums.txt').read_text().splitlines()
                        if line.split()[-1] == name)
        if hashlib.sha256(archive.read_bytes()).hexdigest() != expected:
            raise ValueError('Checksum mismatch')
        subprocess.run(['gh', 'attestation', 'verify', str(archive), '--repo', REPOSITORY,
                        '--signer-workflow', REPOSITORY + '/.github/workflows/release.yml',
                        '--source-ref', 'refs/tags/' + version, '--deny-self-hosted-runners'], check=True, timeout=120)
        member = 'agenthealth.exe' if system == 'windows' else 'agenthealth'
        if extension == 'zip':
            with zipfile.ZipFile(archive) as bundle:
                data = bundle.read(member)
        else:
            with tarfile.open(fileobj=io.BytesIO(archive.read_bytes()), mode='r:gz') as bundle:
                entry = bundle.getmember(member)
                if not entry.isfile():
                    raise ValueError('Unexpected binary member')
                data = bundle.extractfile(entry).read()
        output.parent.mkdir(parents=True, exist_ok=True)
        with output.open('xb') as binary:
            binary.write(data)
        output.chmod(0o755)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('version')
    parser.add_argument('output', type=Path)
    args = parser.parse_args()
    install(args.version, args.output)
