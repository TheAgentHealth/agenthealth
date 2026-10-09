#!/usr/bin/env python3
"""Package the same-version Helm chart for a synchronized software release."""
import argparse
import gzip
import hashlib
import io
import os
from pathlib import Path
import subprocess
import tarfile

import yaml

ROOT = Path(__file__).resolve().parents[1]


def normalize(package, epoch, revision):
    """Keep rebuilt charts byte-identical for provenance-backed recovery."""
    with tarfile.open(package) as archive:
        members = archive.getmembers()
        if any(not m.isfile() or not m.name.startswith('agenthealth/') or
               '..' in Path(m.name).parts for m in members):
            raise ValueError('unexpected chart archive member')
        if len({m.name for m in members}) != len(members):
            raise ValueError('duplicate chart archive member')
        files = [(member.name, archive.extractfile(member).read()) for member in members]
    metadata_name = 'agenthealth/Chart.yaml'
    updated = []
    for name, data in files:
        if name == metadata_name:
            metadata = yaml.safe_load(data)
            metadata.setdefault('annotations', {}).update({
                'org.opencontainers.image.source': 'https://github.com/TheAgentHealth/agenthealth',
                'org.opencontainers.image.revision': revision})
            data = yaml.safe_dump(metadata, sort_keys=False).encode()
        updated.append((name, data))
    files = updated
    with package.open('wb') as output:
        with gzip.GzipFile(filename='', mode='wb', fileobj=output, mtime=epoch) as compressed:
            with tarfile.open(fileobj=compressed, mode='w', format=tarfile.USTAR_FORMAT) as archive:
                for name, data in sorted(files):
                    entry = tarfile.TarInfo(name)
                    entry.size = len(data)
                    entry.mode = 0o644
                    entry.mtime = epoch
                    archive.addfile(entry, io.BytesIO(data))


def package_chart(tag, output, lint=True):
    chart = ROOT / 'deploy/helm/agenthealth'
    metadata = yaml.safe_load((chart / 'Chart.yaml').read_text())
    values = yaml.safe_load((chart / 'values.yaml').read_text())
    if tag != 'v' + metadata['version'] or metadata['appVersion'] != tag:
        raise ValueError('software tag, chart version and appVersion must match')
    if values['image']['tag'] != tag or values['image'].get('digest'):
        raise ValueError('chart default must pin the matching version tag')
    if not (ROOT / 'docs/releases' / (tag + '.md')).is_file():
        raise ValueError('software release notes are required')
    if output.exists() and any(output.iterdir()):
        raise ValueError('output directory must be empty')
    output.mkdir(parents=True, exist_ok=True)
    if lint:
        subprocess.run(['helm', 'lint', str(chart), '--strict', '--kube-version', '1.32.2'], check=True)
    subprocess.run(['helm', 'package', str(chart), '--destination', str(output)], check=True)
    package = output / f"agenthealth-{metadata['version']}.tgz"
    epoch = int(os.environ.get('SOURCE_DATE_EPOCH', subprocess.check_output(
        ['git', 'log', '-1', '--format=%ct'], cwd=ROOT, text=True).strip()))
    revision = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, text=True).strip()
    normalize(package, epoch, revision)
    (output / 'checksums.txt').write_text(f'{hashlib.sha256(package.read_bytes()).hexdigest()}  {package.name}\n')
    return package


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('tag', help='shared vMAJOR.MINOR.PATCH software tag')
    parser.add_argument('--output', type=Path, required=True)
    parser.add_argument('--skip-lint', action='store_true', help='package after lint ran in source validation')
    args = parser.parse_args()
    try:
        package_chart(args.tag, args.output, lint=not args.skip_lint)
    except ValueError as error:
        parser.error(str(error))


if __name__ == '__main__':
    main()
