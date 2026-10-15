#!/usr/bin/env python3
"""Require any existing PyPI version to match original signed release files."""
import argparse
import hashlib
import json
from pathlib import Path
import urllib.error
import urllib.request

from package_python import asset_names, python_version


def verify(tag, output):
    try:
        with urllib.request.urlopen(f'https://pypi.org/pypi/agenthealth/{python_version(tag)}/json', timeout=30) as response:
            metadata = json.load(response)
    except urllib.error.HTTPError as error:
        if error.code == 404:
            return False
        raise
    files = metadata['urls']
    if {item['filename'] for item in files} != asset_names(tag):
        raise ValueError('PyPI distribution set differs from the signed release')
    for item in files:
        original = (output / item['filename']).read_bytes()
        with urllib.request.urlopen(item['url'], timeout=60) as response:
            data = response.read(len(original) + 1)
        if data != original or hashlib.sha256(data).hexdigest() != item['digests']['sha256'] or item['yanked']:
            raise ValueError('Existing PyPI file differs from signed release bytes')
    print('Existing PyPI files match original signed assets; safe to skip their upload')
    return True


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('tag')
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    verify(args.tag, args.output)
