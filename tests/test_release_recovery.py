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


@pytest.mark.parametrize('new_tests', [False, True])
def test_ci_pytest_selects_existing_tag_tests(tmp_path, new_tests):
    workflow = yaml.safe_load((ROOT / '.github/workflows/ci.yml').read_text())
    spec_tests = tmp_path / 'tests/spec'
    spec_tests.mkdir(parents=True)
    (spec_tests / 'test_present.py').write_text('def test_present(): assert True\n')
    if new_tests:
        (tmp_path / 'tests/test_release_ci.py').write_text('def test_release(): assert True\n')
    step = next(s for s in workflow['jobs']['spec-schema-tests']['steps']
                if s.get('name') == 'Run spec schema tests')
    command = step['run'].replace('python3 ', shlex.quote(sys.executable) + ' ')
    result = subprocess.run(['bash', '-e', '-c', command], cwd=tmp_path,
                            check=True, text=True, capture_output=True)
    assert ('2 passed' if new_tests else '1 passed') in result.stdout


@pytest.mark.parametrize('phase13', [False, True])
def test_ci_native_smoke_uses_workflow_helper(tmp_path, phase13):
    workflow = yaml.safe_load((ROOT / '.github/workflows/ci.yml').read_text())
    scripts = tmp_path / 'scripts'
    scripts.mkdir()
    tooling = tmp_path / '.ci-tools/scripts'
    tooling.mkdir(parents=True)
    def helper(path, message):
        path.write_text('from pathlib import Path\n'
                        f'with Path("invocations").open("a") as out: out.write("{message}\\n")\n')
    helper(scripts / 'build_release.py', 'build-tagged-source')
    helper(tooling / 'test_release_binary.py', 'native-smoke')
    if phase13:
        helper(scripts / 'test_release_binary.py', 'old-helper')
    step = next(s for s in workflow['jobs']['binary-distribution']['steps']
                if s.get('name') == 'Build and smoke-test native archive')
    command = step['run'].replace('python ', shlex.quote(sys.executable) + ' ')
    subprocess.run(['bash', '-e', '-c', command], cwd=tmp_path, check=True)
    assert (tmp_path / 'invocations').read_text().splitlines() == ['build-tagged-source', 'native-smoke']


def test_ci_package_guard_for_pre_phase13_tag(tmp_path):
    workflow = yaml.safe_load((ROOT / '.github/workflows/ci.yml').read_text())
    step = next(s for s in workflow['jobs']['binary-distribution']['steps']
                if s.get('name') == 'Build package channels and verify DEB payload')
    subprocess.run(['bash', '-e', '-c', step['run']], cwd=tmp_path, check=True)
