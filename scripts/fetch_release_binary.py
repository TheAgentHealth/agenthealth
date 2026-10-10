#!/usr/bin/env python3
"""Fetch a checksum-verified Linux AMD64 release binary for compatibility tests."""
import argparse
import hashlib
import io
from pathlib import Path
import re
import tarfile
import urllib.request

parser = argparse.ArgumentParser()
parser.add_argument('version')
parser.add_argument('output', type=Path)
args = parser.parse_args()
if not re.fullmatch(r'v\d+\.\d+\.\d+', args.version):
    raise SystemExit('Expected a stable vX.Y.Z version')
name = f'agenthealth_{args.version}_linux_amd64.tar.gz'
base = f'https://github.com/TheAgentHealth/agenthealth/releases/download/{args.version}/'
with urllib.request.urlopen(base + 'checksums.txt', timeout=30) as response:
    checksums = response.read(1048576).decode()
expected = next(line.split()[0] for line in checksums.splitlines() if line.split()[-1] == name)
with urllib.request.urlopen(base + name, timeout=60) as response:
    archive = response.read(128 * 1024 * 1024 + 1)
if len(archive) > 128 * 1024 * 1024 or hashlib.sha256(archive).hexdigest() != expected:
    raise SystemExit('Release archive checksum/size mismatch')
with tarfile.open(fileobj=io.BytesIO(archive), mode='r:gz') as bundle:
    member = bundle.getmember('agenthealth')
    if not member.isfile() or member.size > 128 * 1024 * 1024:
        raise SystemExit('Unexpected binary archive member')
    with bundle.extractfile(member) as binary:
        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_bytes(binary.read())
args.output.chmod(0o755)
print(f'Fetched and verified {args.version}')
