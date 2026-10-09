#!/usr/bin/env python3
"""Add verified package digests to the shared release description."""
import argparse
import os
from pathlib import Path
import re

ROOT = Path(__file__).resolve().parents[1]


def notes(tag, values):
    references = [('Docker GHCR', 'ghcr.io/theagenthealth/agenthealth', values['IMAGE_DIGEST']),
                  ('Helm OCI', 'ghcr.io/theagenthealth/charts/agenthealth', values['CHART_DIGEST']),
                  ('Binary OCI bundle', 'ghcr.io/theagenthealth/agenthealth-binaries', values['BINARY_DIGEST'])]
    if values.get('DOCKER_DIGEST'):
        references.append(('Docker Hub mirror', 'docker.io/theagenthealth/agenthealth', values['DOCKER_DIGEST']))
    if not all(re.fullmatch(r'sha256:[0-9a-f]{64}', digest) for _, _, digest in references):
        raise ValueError('every required publication must supply its verified digest')
    text = (ROOT / 'docs/releases' / (tag + '.md')).read_text().rstrip()
    text += '\n\n## Published package digests\n\n'
    text += 'All entries identify this release source revision. '
    text += '[Repository Packages](https://github.com/TheAgentHealth/agenthealth/packages).\n\n'
    text += '| Package | Immutable reference |\n|---|---|\n'
    for name, repository, digest in references:
        text += f'| {name} | `{repository}@{digest}` |\n'
    if not values.get('DOCKER_DIGEST'):
        text += '\nDocker Hub mirroring was not configured for this release.\n'
    return text


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('tag')
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    args.output.write_text(notes(args.tag, os.environ))
