#!/usr/bin/env python3
"""Verify existing signed release assets/packages for publication-only recovery.

Never build, upload, replace or delete versioned artifacts. Only emit verified
inputs for the later PyPI and final-publication jobs.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import urllib.error
import urllib.request

from package_python import asset_names as python_assets
from publish_oci import public_verification, validate_assets

REPOSITORY = 'TheAgentHealth/agenthealth'
WORKFLOW = REPOSITORY + '/.github/workflows/release.yml'
ACCEPT = ', '.join(('application/vnd.oci.image.index.v1+json',
                    'application/vnd.oci.image.manifest.v1+json',
                    'application/vnd.docker.distribution.manifest.list.v2+json',
                    'application/vnd.docker.distribution.manifest.v2+json'))


def command(arguments):
    return subprocess.check_output(arguments, text=True).strip()


def fetch(url, headers=None):
    request = urllib.request.Request(url, headers=headers or {})
    with urllib.request.urlopen(request, timeout=60) as response:
        data = response.read(4 * 1024 * 1024 + 1)
    if len(data) > 4 * 1024 * 1024:
        raise ValueError('registry metadata exceeds limit')
    return data


def registry(host, repository, reference, token_url):
    token = json.loads(fetch(token_url))['token']
    headers = {'Authorization': 'Bearer ' + token, 'Accept': ACCEPT}
    def get(kind, value):
        data = fetch(f'https://{host}/v2/{repository}/{kind}/{value}', headers)
        digest = 'sha256:' + hashlib.sha256(data).hexdigest()
        if value.startswith('sha256:') and digest != value:
            raise ValueError('registry metadata digest mismatch')
        return json.loads(data), digest
    document, digest = get('manifests', reference)
    return document, digest, get


def image(host, reference, token_url, tag, revision):
    document, digest, get = registry(host, reference, tag, token_url)
    identities = {}
    for item in document['manifests']:
        platform = item['platform']
        if platform['os'] != 'linux' or platform['architecture'] not in ('amd64', 'arm64'):
            continue
        arch = platform['architecture']
        if arch in identities:
            raise ValueError('duplicate image platform')
        manifest, _ = get('manifests', item['digest'])
        config, _ = get('blobs', manifest['config']['digest'])
        labels = config['config']['Labels']
        if (labels.get('org.opencontainers.image.revision') != revision or
                labels.get('org.opencontainers.image.version') != tag or
                labels.get('org.opencontainers.image.source') != 'https://github.com/' + REPOSITORY or
                config.get('architecture') != arch or config.get('os') != 'linux'):
            raise ValueError('image source/version/platform mismatch')
        # Mirror labels/build history can differ; runtime configuration and all
        # filesystem layers must match the signed GHCR image before promotion.
        identities[arch] = (tuple(layer['digest'] for layer in manifest['layers']),
                            config['rootfs'],
                            {key: value for key, value in config['config'].items() if key != 'Labels'})
    if set(identities) != {'amd64', 'arm64'}:
        raise ValueError('both Linux image platforms are required')
    return digest, identities


def verify_signature(subject, revision, tag):
    command(['gh', 'attestation', 'verify', subject, '--repo', REPOSITORY,
             '--signer-workflow', WORKFLOW, '--source-digest', revision,
             '--source-ref', 'refs/tags/' + tag, '--deny-self-hosted-runners'])


def prepare(tag, output):
    if not re.fullmatch(r'v\d+\.\d+\.\d+(?:-(?:a|b|rc)\.\d+)?', tag):
        raise ValueError('expected an immutable software version tag')
    if output.exists() and any(output.iterdir()):
        raise ValueError('recovery output must be empty')
    output.mkdir(parents=True, exist_ok=True)
    subprocess.run(['git', 'fetch', 'origin', 'refs/tags/' + tag], check=True)
    revision = command(['git', 'rev-parse', 'FETCH_HEAD^{commit}'])
    release = json.loads(command(['gh', 'release', 'view', tag, '--repo', REPOSITORY,
                                  '--json', 'isDraft,tagName,assets']))
    if not release['isDraft'] or release['tagName'] != tag:
        raise ValueError('publication-only recovery requires an existing draft for this tag')
    assets = output / 'assets'
    subprocess.run(['gh', 'release', 'download', tag, '--repo', REPOSITORY,
                    '--dir', str(assets)], check=True)
    validate_assets(tag, assets)
    if {item['name'] for item in release['assets']} != {path.name for path in assets.iterdir()}:
        raise ValueError('draft asset list does not match downloaded assets')
    for path in sorted(assets.iterdir()):
        verify_signature(str(path), revision, tag)
    packages = []
    values = {'synchronized': 'true', 'source_sha': revision}
    for repository, version, binary, key in (
            ('theagenthealth/charts/agenthealth', tag[1:], False, 'chart_digest'),
            ('theagenthealth/agenthealth-binaries', tag, True, 'binary_digest')):
        _, digest, _ = registry('ghcr.io', repository, version,
            'https://ghcr.io/token?service=ghcr.io&scope=repository:' + repository + ':pull')
        packages.append({'reference': 'ghcr.io/' + repository + ':' + version,
                         'digest': digest, 'binary': binary})
        values[key] = digest
        verify_signature('oci://ghcr.io/' + repository + '@' + digest, revision, tag)
    report = output / 'oci-report.json'
    report.write_text(json.dumps({'tag': tag, 'source_commit': revision, 'packages': packages}))
    public_verification(report, assets)
    values['image_digest'], ghcr_images = image('ghcr.io', 'theagenthealth/agenthealth',
        'https://ghcr.io/token?service=ghcr.io&scope=repository:theagenthealth/agenthealth:pull', tag, revision)
    verify_signature('oci://ghcr.io/theagenthealth/agenthealth@' + values['image_digest'], revision, tag)
    try:
        mirror_digest, mirror_images = image('registry-1.docker.io', 'theagenthealth/agenthealth',
            'https://auth.docker.io/token?service=registry.docker.io&scope=repository:theagenthealth/agenthealth:pull', tag, revision)
    except urllib.error.HTTPError as error:
        if error.code != 404:
            raise
        values['docker_digest'] = ''
    else:
        if mirror_images != ghcr_images:
            raise ValueError('mirror filesystem/runtime differs from signed GHCR image')
        values['docker_digest'] = mirror_digest
    python_output = output / 'python'
    python_output.mkdir()
    for name in python_assets(tag):
        shutil.copyfile(assets / name, python_output / name)
    (output / 'verified-inputs.json').write_text(json.dumps(values, indent=2) + '\n')
    return values


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('tag')
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    values = prepare(args.tag, args.output)
    if os.environ.get('GITHUB_OUTPUT'):
        with open(os.environ['GITHUB_OUTPUT'], 'a') as stream:
            for key, value in values.items():
                stream.write(key + '=' + value + '\n')
    print('Existing signed artifacts and package inputs verified; no versioned artifact was changed')
