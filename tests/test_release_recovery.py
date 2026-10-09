"""Current workflow must not invoke Phase 13 helpers absent from an older tag."""
from pathlib import Path
import os
import subprocess
import shlex
import sys
import yaml
import pytest

ROOT = Path(__file__).parents[1]


@pytest.mark.parametrize('phase13', [False, True])
def test_recovery_helper_guards(tmp_path, phase13):
    workflow = yaml.safe_load((ROOT / '.github/workflows/release.yml').read_text())
    scripts = tmp_path / 'scripts'
    scripts.mkdir()
    def helper(name, message):
        (scripts / name).write_text(
            'from pathlib import Path\n'
            f'with Path("invocations").open("a") as output: output.write("{message}\\n")\n')
    helper('build_release.py', 'build')
    if phase13:
        helper('test_release_binary.py', 'smoke')
        helper('package_release.py', 'package')
    for job, step_name in [('release', 'Build archives and checksums')]:
        step = next(s for s in workflow['jobs'][job]['steps'] if s.get('name') == step_name)
        command = step['run'].replace('python3 ', shlex.quote(sys.executable) + ' ').replace('python ', shlex.quote(sys.executable) + ' ')
        subprocess.run(['bash', '-e', '-c', command], cwd=tmp_path, check=True,
                       env={**os.environ, 'RELEASE_TAG': 'v0.11.0' if phase13 else 'v0.10.0'})
    assert (tmp_path / 'invocations').read_text().splitlines() == (
        ['build', 'package'] if phase13 else ['build'])
