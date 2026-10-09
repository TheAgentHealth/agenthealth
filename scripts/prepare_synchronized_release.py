#!/usr/bin/env python3
"""Add the Helm chart and regenerate the complete signed release checksum set."""
import argparse
import hashlib
from pathlib import Path
import shutil
import tempfile

from package_helm import package_chart


def asset_names(tag):
    names = {'agenthealth.rb', 'agenthealth.json', 'checksums.txt',
             f'agenthealth-{tag.removeprefix("v")}.tgz'}
    for system, arch in [('linux', 'amd64'), ('linux', 'arm64'),
                         ('darwin', 'amd64'), ('darwin', 'arm64'), ('windows', 'amd64')]:
        stem = f'agenthealth_{tag}_{system}_{arch}'
        names.update({stem + ('.zip' if system == 'windows' else '.tar.gz'), stem + '.cdx.json'})
        if system == 'linux':
            names.update({stem + '.deb', stem + '.rpm'})
    return names


def prepare(tag, output):
    expected = asset_names(tag)
    chart_name = f'agenthealth-{tag.removeprefix("v")}.tgz'
    if not output.is_dir() or {p.name for p in output.iterdir()} != expected - {chart_name}:
        raise ValueError('complete CLI asset set is required; stale or missing files are not allowed')
    lines = (output / 'checksums.txt').read_text().splitlines()
    if {line.split('  ', 1)[1] for line in lines} != expected - {chart_name, 'checksums.txt'}:
        raise ValueError('CLI checksum manifest must cover every asset exactly once')
    if len(lines) != len(expected) - 2:
        raise ValueError('duplicate checksum entries are not allowed')
    for line in lines:
        digest, name = line.split('  ', 1)
        if name not in expected or Path(name).name != name:
            raise ValueError('invalid checksum filename')
        if hashlib.sha256((output / name).read_bytes()).hexdigest() != digest:
            raise ValueError('CLI checksum verification failed')
    with tempfile.TemporaryDirectory(prefix='agenthealth-chart-') as temporary:
        chart = package_chart(tag, Path(temporary), lint=False)
        shutil.copyfile(chart, output / chart.name)
    (output / 'checksums.txt').write_text(''.join(
        f'{hashlib.sha256(path.read_bytes()).hexdigest()}  {path.name}\n'
        for path in sorted(output.iterdir()) if path.name != 'checksums.txt'))


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('tag')
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    try:
        prepare(args.tag, args.output)
    except ValueError as error:
        parser.error(str(error))
