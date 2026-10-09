"""Release gates reject unrelated, skipped, failed, and incomplete CI evidence."""
import importlib.util
from pathlib import Path
import pytest
import yaml

ROOT = Path(__file__).parents[1]
spec = importlib.util.spec_from_file_location('release_ci', ROOT / 'scripts/wait_for_release_ci.py')
gate = importlib.util.module_from_spec(spec)
spec.loader.exec_module(gate)


def run(**updates):
    result = dict(id=10, head_sha='abc', head_branch='main', event='workflow_dispatch',
                  status='completed', conclusion='success', html_url='https://example.com/run',
                  run_attempt=2)
    result.update(updates)
    return result


def mock_api(monkeypatch, runs, jobs=None):
    jobs = jobs if jobs is not None else [
        dict(name=f'Binary distribution ({platform})', conclusion='success')
        for platform in gate.PLATFORMS]
    def api(path):
        if '/attempts/' in path:
            assert '/attempts/2/jobs?' in path
            return {'jobs': jobs}
        assert 'branch=main&event=workflow_dispatch' in path
        return {'workflow_runs': runs}
    monkeypatch.setattr(gate, 'api', api)


def test_success(monkeypatch):
    mock_api(monkeypatch, [run()])
    assert gate.check_ci('owner/repo', 'abc')


@pytest.mark.parametrize('updates', [dict(head_sha='other'), dict(head_branch='feature'),
                                    dict(event='pull_request'), dict(status='in_progress')])
def test_waits_for_exact_release_dispatch(monkeypatch, updates):
    mock_api(monkeypatch, [run(**updates)])
    assert not gate.check_ci('owner/repo', 'abc')


def test_latest_failed_run_blocks_old_success(monkeypatch):
    mock_api(monkeypatch, [run(), run(id=11, conclusion='failure')])
    with pytest.raises(RuntimeError, match='did not succeed'):
        gate.check_ci('owner/repo', 'abc')


@pytest.mark.parametrize('conclusion', ['skipped', 'failure', None])
def test_all_five_jobs_required(monkeypatch, conclusion):
    jobs = [dict(name=f'Binary distribution ({p})', conclusion='success') for p in gate.PLATFORMS]
    jobs[-1]['conclusion'] = conclusion
    mock_api(monkeypatch, [run()], jobs)
    with pytest.raises(RuntimeError, match='Missing or unsuccessful'):
        gate.check_ci('owner/repo', 'abc')


def test_missing_matrix_blocks_release(monkeypatch):
    mock_api(monkeypatch, [run()], [])
    with pytest.raises(RuntimeError, match='Missing or unsuccessful'):
        gate.check_ci('owner/repo', 'abc')


def test_missing_ci_times_out(monkeypatch):
    mock_api(monkeypatch, [])
    with pytest.raises(RuntimeError, match='No completed pre-release CI'):
        gate.wait_for_ci('owner/repo', 'abc', timeout=0)


def test_publication_requires_gate():
    release = yaml.safe_load((ROOT / '.github/workflows/release.yml').read_text())['jobs']
    assert 'binary-smoke' not in release
    for job in ['release', 'container']:
        assert 'ci-gate' in release[job]['needs']
    ci = yaml.safe_load((ROOT / '.github/workflows/ci.yml').read_text())['jobs']
    assert ci['binary-distribution']['if'] == "github.event_name == 'pull_request' || (github.event_name == 'workflow_dispatch' && github.ref == 'refs/heads/main' && inputs.release_sha != '')"
    assert ci['binary-distribution']['strategy']['matrix']['os'] == list(gate.PLATFORMS)


@pytest.mark.parametrize('conclusion', ['success', 'skipped', 'failure', None, 'missing'])
def test_synchronized_release_requires_kubernetes(monkeypatch, conclusion):
    jobs = [dict(name=f'Binary distribution ({p})', conclusion='success') for p in gate.PLATFORMS]
    if conclusion != 'missing':
        jobs.append(dict(name='Kubernetes manifests and Helm', conclusion=conclusion))
    mock_api(monkeypatch, [run()], jobs)
    if conclusion == 'success':
        assert gate.check_ci('owner/repo', 'abc', require_kubernetes=True)
    else:
        with pytest.raises(RuntimeError, match='Kubernetes manifests and Helm'):
            gate.check_ci('owner/repo', 'abc', require_kubernetes=True)
    # Older standalone layouts continue to need only binary evidence.
    assert gate.check_ci('owner/repo', 'abc')


def test_wait_propagates_kubernetes_requirement(monkeypatch):
    mock_api(monkeypatch, [run()])
    with pytest.raises(RuntimeError, match='Kubernetes manifests and Helm'):
        gate.wait_for_ci('owner/repo', 'abc', require_kubernetes=True)


@pytest.mark.parametrize('synchronized', [False, True])
def test_gate_detects_layout_from_tagged_source(tmp_path, synchronized):
    import os
    import subprocess
    import sys
    import shlex
    def git(*args):
        return subprocess.check_output(['git', '-c', 'user.name=Test', '-c',
                                       'user.email=test@example.com', *args],
                                      cwd=tmp_path, text=True).strip()
    git('init', '-q')
    scripts = tmp_path / 'scripts'
    scripts.mkdir()
    helper = scripts / 'wait_for_release_ci.py'
    helper.write_text('import sys\nfrom pathlib import Path\n'
                      'Path("arguments").write_text("\\n".join(sys.argv[1:]))\n')
    if synchronized:
        (scripts / 'prepare_synchronized_release.py').write_text('# synchronized\n')
    git('add', '.')
    git('commit', '-qm', 'tagged source')
    sha = git('rev-parse', 'HEAD')
    # HEAD deliberately has the opposite layout to the target tag.
    marker = scripts / 'prepare_synchronized_release.py'
    if synchronized:
        marker.unlink()
    else:
        marker.write_text('# current source\n')
    git('add', '.')
    git('commit', '-qm', 'different current layout')
    jobs = yaml.safe_load((ROOT / '.github/workflows/release.yml').read_text())['jobs']
    step = next(s for s in jobs['ci-gate']['steps'] if s.get('name') == 'Trigger and require pre-release CI for the tagged commit')
    # Exercise the actual layout selection after tag fetch and SHA resolution.
    command = step['run'][step['run'].index('required_checks='):]
    command = command.replace('python3 ', shlex.quote(sys.executable) + ' ')
    subprocess.run(['bash', '-e', '-c', command], cwd=tmp_path, check=True,
                   env={**os.environ, 'source_sha': sha})
    assert (tmp_path / 'arguments').read_text().splitlines() == (
        [sha, '--dispatch', '--require-kubernetes'] if synchronized else [sha, '--dispatch'])


@pytest.mark.parametrize('title,accepted', [
    ('Pre-release CI abc unique', True),
    ('Pre-release CI abc previous', False),
    ('Pre-release CI different unique', False),
])
def test_correlates_fresh_dispatch_after_main_advances(monkeypatch, title, accepted):
    mock_api(monkeypatch, [run(head_sha='new-main', display_title=title)])
    assert gate.check_ci('owner/repo', 'abc', request_id='unique') == accepted


def test_dispatch_payload(monkeypatch):
    import json
    calls = []
    monkeypatch.setattr(gate.subprocess, 'run', lambda command, **kwargs: calls.append((command, kwargs)))
    request = gate.dispatch_ci('owner/repo', 'abc', require_kubernetes=True)
    command, kwargs = calls[0]
    assert 'repos/owner/repo/actions/workflows/ci.yml/dispatches' in command
    assert kwargs['check'] is True
    assert json.loads(kwargs['input']) == {
        'ref': 'main', 'inputs': {'release_sha': 'abc', 'request_id': request,
                                 'require_kubernetes': True}}


def test_ci_uses_immutable_release_commit():
    workflow = yaml.safe_load((ROOT / '.github/workflows/ci.yml').read_text())
    triggers = workflow.get('on', workflow.get(True))
    assert 'push' not in triggers
    assert 'pull_request' in triggers
    for job in workflow['jobs'].values():
        for step in job.get('steps', []):
            if step.get('uses', '').startswith('actions/checkout@'):
                if step['with'].get('path') == '.ci-tools':
                    assert step['with']['ref'] == '${{ github.workflow_sha }}'
                    assert step['with']['persist-credentials'] is False
                else:
                    assert step['with']['ref'] == '${{ inputs.release_sha || github.sha }}'
