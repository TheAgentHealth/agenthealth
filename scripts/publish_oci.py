#!/usr/bin/env python3
"""Publish immutable Helm/binary OCI references from signed release assets."""
import argparse
import hashlib
import json
import os
import re
from pathlib import Path
import subprocess
import urllib.request
from urllib.parse import quote

from prepare_synchronized_release import asset_names

SOURCE = 'https://github.com/TheAgentHealth/agenthealth'
CHART = 'ghcr.io/theagenthealth/charts/agenthealth'
BINARIES = 'ghcr.io/theagenthealth/agenthealth-binaries'
BINARY_TYPE = 'application/vnd.agenthealth.release.v1'


def command(args, **kwargs):
    return subprocess.run(args, check=True, capture_output=True, text=True, **kwargs).stdout


def manifest(reference):
    result = subprocess.run(['oras', 'manifest', 'fetch', reference], capture_output=True, text=True)
    if result.returncode:
        error = result.stderr.lower()
        # GHCR can deny a pull-only token for a package that has not been created.
        # Confirm absence through the authenticated Packages API before first push.
        if any(code in error for code in ('403', 'denied', 'unauthorized')):
            repository, tag = reference.removeprefix('ghcr.io/').rsplit(':', 1)
            owner, name = repository.split('/', 1)
            endpoint = f'orgs/{owner}/packages/container/{quote(name, safe="")}/versions'
            lookup = subprocess.run(['gh', 'api', endpoint, '--paginate', '--jq',
                '.[] | .metadata.container.tags[]?'], capture_output=True, text=True)
            if lookup.returncode == 0 and tag not in lookup.stdout.splitlines():
                return None
            if lookup.returncode and '404' in lookup.stderr:
                return None
        if any(code in error for code in ('404', 'manifest_unknown', 'name_unknown', 'not found')):
            return None
        raise RuntimeError('Registry inspection failed; existing references will not be overwritten')
    return json.loads(result.stdout)


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def validate_assets(tag, output):
    if {p.name for p in output.iterdir()} != asset_names(tag) or not all(p.is_file() for p in output.iterdir()):
        raise ValueError('unexpected or missing release assets')
    lines = (output / 'checksums.txt').read_text().splitlines()
    hashes = dict((name, digest) for digest, name in (line.split('  ', 1) for line in lines))
    if set(hashes) != asset_names(tag) - {'checksums.txt'} or len(hashes) != len(lines):
        raise ValueError('checksum coverage must match the complete asset set')
    for name, digest in hashes.items():
        if sha(output / name) != digest:
            raise ValueError('asset checksum mismatch')


def validate_manifest(document, expected, version, revision, binary=False):
    annotations = document.get('annotations', {})
    if any(annotations.get(key) != value for key, value in {
            'org.opencontainers.image.source': SOURCE,
            'org.opencontainers.image.revision': revision,
            'org.opencontainers.image.version': version}.items()):
        raise ValueError('OCI source/version does not match this immutable release')
    layers = document.get('layers', [])
    if binary:
        actual = {layer.get('annotations', {}).get('org.opencontainers.image.title'): layer['digest']
                  for layer in layers}
        if document.get('artifactType') != BINARY_TYPE or len(actual) != len(layers) or actual != expected:
            raise ValueError('existing binary bundle differs; refusing to replace the version')
    elif (len(layers) != 1 or layers[0]['digest'] != next(iter(expected.values()))
          or layers[0]['mediaType'] != 'application/vnd.cncf.helm.chart.content.v1.tar+gzip'):
        raise ValueError('existing chart differs; refusing to replace the version')


def publish(tag, output, report):
    validate_assets(tag, output)
    revision = command(['git', 'rev-parse', 'HEAD']).strip()
    if command(['git', 'rev-parse', tag + '^{commit}']).strip() != revision:
        raise ValueError('publish only the exact immutable tagged source')
    chart = output / f'agenthealth-{tag[1:]}.tgz'
    records = []
    for reference, files, binary in [(CHART + ':' + tag[1:], [chart], False),
                                     (BINARIES + ':' + tag, sorted(output.iterdir()), True)]:
        expected = {p.name: 'sha256:' + sha(p) for p in files}
        document = manifest(reference)
        if document is None:
            if binary:
                stamp = command(['git', 'log', '-1', '--format=%cI']).strip()
                args = ['oras', 'push', reference, '--artifact-type', BINARY_TYPE,
                        '--annotation', 'org.opencontainers.image.source=' + SOURCE,
                        '--annotation', 'org.opencontainers.image.revision=' + revision,
                        '--annotation', 'org.opencontainers.image.version=' + tag,
                        '--annotation', 'org.opencontainers.image.created=' + stamp]
                # The complete release set includes the chart so checksums have no absent entries.
                args += [p.name + ':application/octet-stream' for p in files]
                command(args, cwd=output)
            else:
                command(['helm', 'push', str(chart), 'oci://ghcr.io/theagenthealth/charts'])
            document = manifest(reference)
        if document is None:
            raise RuntimeError('published OCI manifest is unavailable')
        validate_manifest(document, expected, tag if binary else tag[1:], revision, binary)
        descriptor = json.loads(command(['oras', 'manifest', 'fetch', '--descriptor', reference]))
        digest = descriptor['digest']
        if not re.fullmatch(r'sha256:[0-9a-f]{64}', digest):
            raise ValueError('unexpected OCI digest')
        records.append({'reference': reference, 'digest': digest, 'binary': binary})
        if os.environ.get('GITHUB_OUTPUT'):
            with open(os.environ['GITHUB_OUTPUT'], 'a') as stream:
                stream.write(('binary_digest' if binary else 'chart_digest') + '=' + digest + '\n')
    report.write_text(json.dumps({'source_commit': revision, 'tag': tag, 'packages': records}, indent=2) + '\n')


def public_verification(report, output):
    records = json.loads(report.read_text())
    validate_assets(records['tag'], output)
    for package in records['packages']:
        reference = package['reference']
        repository = reference.removeprefix('ghcr.io/').rsplit(':', 1)[0]
        url = 'https://ghcr.io/token?service=ghcr.io&scope=repository:' + repository + ':pull'
        # No GitHub token or credential store is used: this establishes public access.
        with urllib.request.urlopen(url, timeout=30) as response:
            token = json.load(response)['token']
        def fetch(path, accept='application/octet-stream'):
            request = urllib.request.Request('https://ghcr.io/v2/' + repository + '/' + path,
                headers={'Authorization': 'Bearer ' + token, 'Accept': accept})
            with urllib.request.urlopen(request, timeout=60) as response:
                return response.read()
        raw = fetch('manifests/' + package['digest'], 'application/vnd.oci.image.manifest.v1+json')
        if 'sha256:' + hashlib.sha256(raw).hexdigest() != package['digest']:
            raise ValueError('public manifest digest mismatch')
        document = json.loads(raw)
        files = sorted(output.iterdir()) if package['binary'] else [output / f"agenthealth-{records['tag'][1:]}.tgz"]
        expected = {p.name: 'sha256:' + sha(p) for p in files}
        validate_manifest(document, expected, records['tag'] if package['binary'] else records['tag'][1:],
                          records['source_commit'], package['binary'])
        config = document['config']
        config_data = fetch('blobs/' + config['digest'])
        if 'sha256:' + hashlib.sha256(config_data).hexdigest() != config['digest']:
            raise ValueError('public config checksum mismatch')
        if not package['binary']:
            metadata = json.loads(config_data)
            if metadata.get('name') != 'agenthealth' or metadata.get('version') != records['tag'][1:] or metadata.get('appVersion') != records['tag']:
                raise ValueError('published Helm config version mismatch')
        for layer in document['layers']:
            data = fetch('blobs/' + layer['digest'])
            if 'sha256:' + hashlib.sha256(data).hexdigest() != layer['digest']:
                raise ValueError('public blob checksum mismatch')
        print('Anonymous manifest and all asset blobs verified:', reference)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('tag')
    parser.add_argument('--output', type=Path, required=True)
    parser.add_argument('--report', type=Path, required=True)
    parser.add_argument('--verify-public', action='store_true')
    args = parser.parse_args()
    if args.verify_public:
        if json.loads(args.report.read_text())['tag'] != args.tag:
            parser.error('report tag mismatch')
        public_verification(args.report, args.output.resolve())
    else:
        publish(args.tag, args.output.resolve(), args.report)
