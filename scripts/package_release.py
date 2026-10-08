#!/usr/bin/env python3
"""Generate package-channel assets from verified release archives (no rebuild)."""
import argparse
import hashlib
import json
import os
import re
import subprocess
import tarfile
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
NFPM = 'github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.47.0'
BASE = 'https://github.com/TheAgentHealth/agenthealth/releases/download'


def generate(version, output):
    if not re.fullmatch(r'v\d+\.\d+\.\d+(?:-[A-Za-z0-9.-]+)?', version):
        raise ValueError('invalid release version')
    hashes = {}
    for line in (output / 'checksums.txt').read_text().splitlines():
        digest, name = line.split('  ', 1)
        if Path(name).name != name or not re.fullmatch(r'[a-f0-9]{64}', digest):
            raise ValueError('invalid checksum entry')
        if hashlib.sha256((output / name).read_bytes()).hexdigest() != digest:
            raise ValueError(f'checksum mismatch: {name}')
        hashes[name] = digest
    # Never overwrite generated assets or silently mix release versions.
    destinations = ['agenthealth.rb', 'agenthealth.json'] + [
        f'agenthealth_{version}_linux_{arch}.{kind}'
        for arch in ('amd64', 'arm64') for kind in ('deb', 'rpm')]
    if any((output / name).exists() for name in destinations):
        raise ValueError('package assets already exist')
    def archive(system, arch):
        name = f'agenthealth_{version}_{system}_{arch}' + ('.zip' if system == 'windows' else '.tar.gz')
        return f'{BASE}/{version}/{name}', hashes[name]
    formula = ['class Agenthealth < Formula', '  desc "Vendor-neutral agent health CLI"',
               '  homepage "https://github.com/TheAgentHealth/agenthealth"',
               f'  version "{version[1:]}"', '  license "Apache-2.0"']
    for system, block in [('darwin', 'macos'), ('linux', 'linux')]:
        formula.append(f'  on_{block} do')
        for arch, cpu in [('arm64', 'arm'), ('amd64', 'intel')]:
            url, digest = archive(system, arch)
            formula.extend([f'    on_{cpu} do', f'      url "{url}"', f'      sha256 "{digest}"', '    end'])
        formula.append('  end')
    formula.extend(['  def install', '    bin.install "agenthealth"', '  end', '  test do',
                    f'    assert_equal "agenthealth {version}", shell_output("#{{bin}}/agenthealth version").strip',
                    '  end', 'end'])
    url, digest = archive('windows', 'amd64')
    scoop = {'version': version[1:], 'description': 'Vendor-neutral agent health CLI',
             'homepage': 'https://github.com/TheAgentHealth/agenthealth', 'license': 'Apache-2.0',
             'architecture': {'64bit': {'url': url, 'hash': digest}}, 'bin': 'agenthealth.exe'}
    with tempfile.TemporaryDirectory(prefix='agenthealth-packages-') as temporary:
        staging = Path(temporary)
        for arch in ('amd64', 'arm64'):
            source = output / f'agenthealth_{version}_linux_{arch}.tar.gz'
            with tarfile.open(source) as bundle:
                # Read only the allowlisted executable; never extract arbitrary paths.
                executable = bundle.getmember('agenthealth')
                if not executable.isfile():
                    raise ValueError('archive executable must be a regular file')
                binary = staging / 'agenthealth'
                binary.write_bytes(bundle.extractfile(executable).read())
                binary.chmod(0o755)
            config = {'name': 'agenthealth', 'arch': arch, 'platform': 'linux',
                      'version': version[1:], 'maintainer': 'TheAgentHealth',
                      'description': 'Vendor-neutral agent health CLI', 'license': 'Apache-2.0',
                      'homepage': 'https://github.com/TheAgentHealth/agenthealth',
                      'depends': ['ca-certificates'], 'contents': [
                          {'src': str(binary), 'dst': '/usr/bin/agenthealth'},
                          {'src': str(ROOT / 'LICENSE'), 'dst': '/usr/share/doc/agenthealth/LICENSE'},
                          {'src': str(ROOT / 'docs/installation.md'), 'dst': '/usr/share/doc/agenthealth/INSTALL.md'}]}
            config_path = staging / 'nfpm.json'
            config_path.write_text(json.dumps(config))
            for kind in ('deb', 'rpm'):
                target = output / f'agenthealth_{version}_linux_{arch}.{kind}'
                subprocess.run(['go', 'run', NFPM, 'package', '--config', str(config_path),
                                '--packager', kind, '--target', str(target.resolve())], check=True, cwd=ROOT,
                               env={**os.environ, 'SOURCE_DATE_EPOCH': str(executable.mtime)})
    (output / 'agenthealth.rb').write_text('\n'.join(formula) + '\n')
    (output / 'agenthealth.json').write_text(json.dumps(scoop, indent=2) + '\n')
    assets = sorted(path for path in output.iterdir() if path.name != 'checksums.txt')
    (output / 'checksums.txt').write_text(''.join(
        f'{hashlib.sha256(path.read_bytes()).hexdigest()}  {path.name}\n' for path in assets))


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('version')
    parser.add_argument('--output', type=Path, default=ROOT / 'dist')
    args = parser.parse_args()
    generate(args.version, args.output.resolve())
