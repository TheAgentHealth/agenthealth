#!/usr/bin/env python3
"""Build the wheel/sdist from tagged source and verify synchronized metadata."""
import argparse
import ast
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile

ROOT = Path(__file__).resolve().parents[1]


def python_version(tag):
    match = re.fullmatch(r'v(\d+\.\d+\.\d+)(?:-(a|b|rc)\.(\d+))?', tag)
    if not match:
        raise ValueError('Python distribution requires vX.Y.Z or vX.Y.Z-(a|b|rc).N')
    return match[1] + ((match[2] + match[3]) if match[2] else '')


def has_python(tag):
    # Historical releases must retain their original complete asset set.
    match = re.match(r'^v(\d+)\.(\d+)\.(\d+)(?:-|$)', tag)
    return bool(match and tuple(map(int, match.groups())) >= (0, 13, 0))


def asset_names(tag):
    version = python_version(tag)
    return {f'agenthealth-{version}-py3-none-any.whl', f'agenthealth-{version}.tar.gz'}


def build_python(tag, output):
    version = python_version(tag)
    package = ROOT / 'sdk/python'
    assignment = ast.parse((package / 'src/agenthealth/version.py').read_text()).body[-1]
    if not isinstance(assignment, ast.Assign) or ast.literal_eval(assignment.value) != version:
        raise ValueError('SDK source version must match the software tag')
    if not re.search(r'^version = "' + re.escape(version) + r'"$', (package / 'pyproject.toml').read_text(), re.M):
        raise ValueError('SDK package metadata must match the software tag')
    for name in ('configuration', 'result'):
        if (package / f'src/agenthealth/schemas/{name}.schema.json').read_bytes() != (ROOT / f'spec/schemas/{name}.schema.json').read_bytes():
            raise ValueError('SDK schemas must match the AHS schemas')
    with tempfile.TemporaryDirectory(prefix='agenthealth-python-build-') as directory:
        subprocess.run([sys.executable, '-m', 'build', str(package), '--outdir', directory], check=True)
        distributions = list(Path(directory).iterdir())
        if {path.name for path in distributions} != asset_names(tag):
            raise ValueError('Unexpected Python distribution set')
        subprocess.run([sys.executable, '-m', 'twine', 'check', '--strict',
                        *map(str, distributions)], check=True)
        output.mkdir(parents=True, exist_ok=True)
        if any((output / path.name).exists() for path in distributions):
            raise ValueError('Python distribution output already exists')
        for path in distributions:
            shutil.copyfile(path, output / path.name)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('tag')
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    build_python(args.tag, args.output)
