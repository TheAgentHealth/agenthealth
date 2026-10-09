"""Exercise PR path detection without starting a disposable cluster."""
from pathlib import Path
import os
import subprocess
import yaml
import pytest

ROOT = Path(__file__).parents[1]
WORKFLOW = yaml.safe_load((ROOT / '.github/workflows/ci.yml').read_text())['jobs']


@pytest.mark.parametrize('path,relevant', [
    ('deploy/kubernetes/configmap.yaml', True),
    ('deploy/helm/agenthealth/values.yaml', True),
    ('examples/kubernetes/Dockerfile.exec-probe', True),
    ('tests/kubernetes/test_kubernetes.py', True),
    ('scripts/package_helm.py', True),
    ('scripts/test_kubernetes.py', True),
    ('Dockerfile', True), ('.dockerignore', True), ('go.mod', True),
    ('core/serve.go', True), ('cmd/agenthealth/main.go', True),
    ('adapters/http/http.go', True),
    ('spec/schemas/configuration.schema.json', True),
    ('.github/workflows/ci.yml', True), ('tests/requirements.txt', True),
    ('README.md', False), ('docs/kubernetes.md', False),
    ('core/README.md', False), ('core/serve_test.go', False),
    ('scripts/build_release.py', False),
])
def test_pr_paths(tmp_path, path, relevant):
    subprocess.run(['git', 'init', '-q', str(tmp_path)], check=True)
    def git(*args):
        return subprocess.check_output(['git', '-c', 'user.name=Test', '-c',
                                       'user.email=test@example.com', *args],
                                      cwd=tmp_path, text=True).strip()
    git('commit', '--allow-empty', '-qm', 'base')
    base = git('rev-parse', 'HEAD')
    target = tmp_path / path
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_text('changed\n')
    git('add', '.')
    git('commit', '-qm', 'PR change')
    head = git('rev-parse', 'HEAD')
    # Advance the base branch: its unrelated changes must not trigger the PR.
    git('checkout', '--detach', base)
    (tmp_path / 'Dockerfile').write_text('unrelated base change\n')
    git('add', '.')
    git('commit', '-qm', 'base advanced')
    base = git('rev-parse', 'HEAD')
    output = tmp_path / 'output'
    step = next(step for step in WORKFLOW['kubernetes-changes']['steps'] if step.get('id') == 'paths')
    subprocess.run(['bash', '-e', '-c', step['run']], cwd=tmp_path, check=True,
                   env={**os.environ, 'BASE_SHA': base, 'HEAD_SHA': head,
                        'GITHUB_OUTPUT': str(output), 'EVENT_NAME': 'pull_request'})
    assert output.read_text() == f'relevant={str(relevant).lower()}\n'


def test_job_conditions():
    assert "github.event_name == 'pull_request'" in WORKFLOW['kubernetes-changes']['if']
    assert WORKFLOW['kubernetes-changes']['if'] == "github.event_name == 'pull_request'"
    job = WORKFLOW['kubernetes']
    assert job['needs'] == 'kubernetes-changes'
    assert "github.event_name == 'pull_request'" in job['if']
    assert "needs.kubernetes-changes.outputs.relevant == 'true'" in job['if']
    assert "needs.kubernetes-changes.result != 'success'" in job['if']
    assert "github.event_name == 'workflow_dispatch' && github.ref == 'refs/heads/main'" in job['if']
    assert 'inputs.require_kubernetes' in job['if']
    assert 'always()' in job['if']
