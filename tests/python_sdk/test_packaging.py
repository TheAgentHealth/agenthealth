"""Python artifacts must join the complete immutable release and gate."""
import importlib
from pathlib import Path

import pytest
import yaml

ROOT = Path(__file__).resolve().parents[2]


@pytest.fixture
def packaging(monkeypatch):
    monkeypatch.syspath_prepend(str(ROOT / 'scripts'))
    return importlib.import_module('package_python')


def test_version_mapping_and_historical_layout(packaging):
    assert packaging.python_version('v0.13.0-rc.1') == '0.13.0rc1'
    assert packaging.python_version('v0.13.0') == '0.13.0'
    assert not packaging.has_python('v0.12.0')
    assert packaging.has_python('v0.13.0-rc.1')
    for tag in ('latest', 'v1.0.0-preview', '../../v1.0.0'):
        with pytest.raises(ValueError):
            packaging.python_version(tag)
    prepare = importlib.import_module('prepare_synchronized_release')
    assert len(prepare.asset_names('v0.12.0')) == 18
    assert len(prepare.asset_names('v0.13.0')) == 20


def test_source_version_mismatch_fails_before_build(packaging, tmp_path):
    with pytest.raises(ValueError, match='source version'):
        packaging.build_python('v0.14.0', tmp_path)
    assert not list(tmp_path.iterdir())


def test_pypi_reuses_staged_assets_and_gates_final_release():
    jobs = yaml.safe_load((ROOT / '.github/workflows/release.yml').read_text())['jobs']
    pypi = jobs['pypi']
    assert pypi['environment']['name'] == 'pypi'
    assert pypi['permissions'] == {'contents': 'read', 'id-token': 'write'}
    assert pypi['needs'] == ['release', 'container', 'packages']
    assert any(step.get('with', {}).get('attestations') is True for step in pypi['steps'])
    assert 'pypi' in jobs['publish']['needs']
    assert "needs.pypi.result == 'success'" in jobs['publish']['if']
    assert any('--require-python' in step.get('run', '') for step in jobs['ci-gate']['steps'])


def test_release_gate_rejects_skipped_sdk_matrix(packaging, monkeypatch):
    gate = importlib.import_module('wait_for_release_ci')
    run = {'id': 1, 'head_sha': 'abc', 'head_branch': 'main', 'event': 'workflow_dispatch',
           'status': 'completed', 'conclusion': 'success', 'html_url': 'example', 'run_attempt': 1}
    jobs = [{'name': f'Binary distribution ({os})', 'conclusion': 'success'} for os in gate.PLATFORMS]
    jobs += [{'name': f'Python SDK ({os}, {version})', 'conclusion': 'success'}
             for os in ('ubuntu-latest', 'macos-latest', 'windows-latest') for version in ('3.10', '3.13')]
    jobs += [{'name': f'Kubernetes interoperability ({version})', 'conclusion': 'success'}
             for version in ('1.31.9', '1.33.1')]
    monkeypatch.setattr(gate, 'api', lambda path: {'jobs': jobs} if '/jobs?' in path else {'workflow_runs': [run]})
    assert gate.check_ci('org/repo', 'abc', require_python=True)
    jobs[6]['conclusion'] = 'skipped'
    with pytest.raises(RuntimeError, match='Python SDK'):
        gate.check_ci('org/repo', 'abc', require_python=True)
