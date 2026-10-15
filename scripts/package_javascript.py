#!/usr/bin/env python3
"""Build the exact-version npm tarball included in signed synchronized releases."""
import argparse
import gzip
import io
import json
import os
from pathlib import Path
import re
import subprocess
import tarfile
import tempfile

ROOT = Path(__file__).resolve().parents[1]


def has_javascript(tag):
    match = re.fullmatch(r'v(\d+)\.(\d+)\.(\d+)(?:-[0-9A-Za-z.-]+)?', tag)
    return bool(match and tuple(map(int, match.groups())) >= (0, 14, 0))


def asset_names(tag):
    if not re.fullmatch(r'v\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?', tag):
        raise ValueError('npm distribution requires a semantic version tag')
    return {f'theagenthealth-agenthealth-{tag[1:]}.tgz'}


def build_javascript(tag, output):
    expected = asset_names(tag)
    package = ROOT / 'sdk/javascript'
    metadata = json.loads((package / 'package.json').read_text())
    if metadata['version'] != tag[1:] or f"export const version = '{tag[1:]}';" not in (package / 'index.js').read_text():
        raise ValueError('JavaScript SDK source version must match software tag')
    for name in ('configuration', 'result'):
        if (package / f'schemas/{name}.schema.json').read_bytes() != (ROOT / f'spec/schemas/{name}.schema.json').read_bytes():
            raise ValueError('JavaScript SDK schemas must match AHS schemas')
    if any((output / name).exists() for name in expected):
        raise ValueError('npm distribution output already exists')
    with tempfile.TemporaryDirectory(prefix='agenthealth-npm-') as temporary:
        subprocess.run(['npm', 'pack', '--ignore-scripts', '--pack-destination', temporary], cwd=package, check=True)
        archives = list(Path(temporary).iterdir())
        if {p.name for p in archives} != expected:
            raise ValueError('Unexpected npm distribution set')
        epoch = int(os.environ.get('SOURCE_DATE_EPOCH') or subprocess.check_output(
            ['git', 'log', '-1', '--format=%ct'], cwd=ROOT, text=True).strip())
        with tarfile.open(archives[0]) as archive:
            entries = [(m.name, archive.extractfile(m).read()) for m in archive.getmembers() if m.isfile()]
        required = {'package/package.json', 'package/index.js', 'package/index.d.ts', 'package/README.md',
                    'package/LICENSE', 'package/schemas/configuration.schema.json', 'package/schemas/result.schema.json'}
        if {name for name, _ in entries} != required:
            raise ValueError('npm package contains missing or unexpected files')
        output.mkdir(parents=True, exist_ok=True)
        with (output / archives[0].name).open('wb') as stream:
            with gzip.GzipFile(filename='', mode='wb', fileobj=stream, mtime=epoch) as compressed:
                with tarfile.open(fileobj=compressed, mode='w', format=tarfile.PAX_FORMAT) as archive:
                    for name, data in sorted(entries):
                        entry = tarfile.TarInfo(name)
                        entry.size = len(data)
                        entry.mode = 0o644
                        entry.mtime = epoch
                        archive.addfile(entry, io.BytesIO(data))


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('tag')
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    build_javascript(args.tag, args.output)
