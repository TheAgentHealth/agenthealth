#!/usr/bin/env python3
"""Validate and package an independently versioned Helm component release."""
import argparse
import hashlib
from pathlib import Path
import subprocess

import yaml

ROOT = Path(__file__).resolve().parents[1]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('tag', help='helm-v<chart-version>')
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    chart = ROOT / 'deploy/helm/agenthealth'
    metadata = yaml.safe_load((chart / 'Chart.yaml').read_text())
    if args.tag != 'helm-v' + metadata['version']:
        parser.error('tag must match the chart version')
    notes = ROOT / 'docs/releases' / (args.tag + '.md')
    if not notes.is_file():
        parser.error('component release notes are required')
    args.output.mkdir(parents=True, exist_ok=True)
    if any(args.output.iterdir()):
        parser.error('output directory must be empty')
    subprocess.run(['helm', 'lint', str(chart), '--strict', '--kube-version', '1.32.2'], check=True)
    subprocess.run(['helm', 'package', str(chart), '--destination', str(args.output)], check=True)
    package = args.output / f"agenthealth-{metadata['version']}.tgz"
    checksum = hashlib.sha256(package.read_bytes()).hexdigest()
    (args.output / 'checksums.txt').write_text(f'{checksum}  {package.name}\n')


if __name__ == '__main__':
    main()
