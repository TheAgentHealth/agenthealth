#!/usr/bin/env python3
"""Verify release-bound npm tarballs and registry integrity without npm tokens."""
import argparse
import base64
import hashlib
import io
import json
import os
from pathlib import Path
import re
import subprocess
import tarfile
import time
import urllib.error
import urllib.request

REPOSITORY = 'TheAgentHealth/agenthealth'


def fetch(url):
    with urllib.request.urlopen(url, timeout=60) as response:
        data = response.read(32 * 1024 * 1024 + 1)
    if len(data) > 32 * 1024 * 1024:
        raise ValueError('npm response exceeds limit')
    return data


def registry_document(version, wait=False, timeout=600):
    """Allow npm's asynchronous publication to reach the public registry."""
    deadline = time.monotonic() + timeout
    while True:
        try:
            return json.loads(fetch('https://registry.npmjs.org/@theagenthealth%2fagenthealth/' + version))
        except urllib.error.HTTPError as error:
            if not wait or error.code not in (404, 429, 500, 502, 503, 504):
                raise
            if time.monotonic() >= deadline:
                raise TimeoutError('npm registry publication did not propagate within the deadline') from error
            print('Waiting for npm registry publication to propagate...', flush=True)
            time.sleep(min(10, max(0, deadline - time.monotonic())))


def verify(tag, output, registry=False):
    if not re.fullmatch(r'v\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?', tag):
        raise ValueError('expected semantic version tag')
    version = tag[1:]
    name = f'theagenthealth-agenthealth-{version}.tgz'
    if not registry:
        output.mkdir(parents=True, exist_ok=False)
        subprocess.run(['git', 'fetch', 'origin', 'refs/tags/' + tag], check=True)
        sha = subprocess.check_output(['git', 'rev-parse', 'FETCH_HEAD^{commit}'], text=True).strip()
        subprocess.run(['git', 'merge-base', '--is-ancestor', sha, 'origin/main'], check=True)
        if os.environ.get('GITHUB_ACTIONS') == 'true':
            tagged = os.environ.get('GITHUB_REF') == 'refs/tags/' + tag and os.environ.get('GITHUB_SHA') == sha
            recovery = os.environ.get('NPM_PUBLICATION_RECOVERY') == 'true' and os.environ.get('GITHUB_REF') == 'refs/heads/main'
            if not (tagged or recovery):
                raise ValueError('npm publication requires the source tag or explicit main-only infrastructure recovery')
        subprocess.run(['gh', 'release', 'download', tag, '--repo', REPOSITORY, '--dir', str(output),
                        '--pattern', name, '--pattern', 'checksums.txt'], check=True)
        for filename in (name, 'checksums.txt'):
            subprocess.run(['gh', 'attestation', 'verify', str(output / filename), '--repo', REPOSITORY,
                            '--signer-workflow', REPOSITORY + '/.github/workflows/release.yml',
                            '--source-digest', sha, '--source-ref', 'refs/tags/' + tag,
                            '--deny-self-hosted-runners'], check=True)
    data = (output / name).read_bytes()
    lines = (output / 'checksums.txt').read_text().splitlines()
    if lines.count(hashlib.sha256(data).hexdigest() + '  ' + name) != 1:
        raise ValueError('signed npm checksum mismatch')
    with tarfile.open(fileobj=io.BytesIO(data)) as archive:
        metadata = json.load(archive.extractfile('package/package.json'))
        if metadata['name'] != '@theagenthealth/agenthealth' or metadata['version'] != version:
            raise ValueError('npm package identity mismatch')
        if metadata.get('scripts', {}).get('postinstall') or metadata.get('scripts', {}).get('preinstall'):
            raise ValueError('npm package must not have installation scripts')
    try:
        document = registry_document(version, wait=registry)
    except urllib.error.HTTPError as error:
        if error.code == 404 and not registry:
            return
        raise
    dist = document['dist']
    if fetch(dist['tarball']) != data:
        raise ValueError('registry npm bytes differ from signed release tarball')
    if dist['integrity'] != 'sha512-' + base64.b64encode(hashlib.sha512(data).digest()).decode():
        raise ValueError('npm registry integrity mismatch')
    if not dist.get('attestations'):
        raise ValueError('npm provenance is required')
    if not registry:
        (output / 'already-published').write_text('Verified existing identical publication\n')
    print('Verified public npm tarball integrity and provenance reference')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('tag')
    parser.add_argument('--output', type=Path, required=True)
    parser.add_argument('--registry', action='store_true')
    args = parser.parse_args()
    verify(args.tag, args.output, args.registry)
