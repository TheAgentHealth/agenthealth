"""Validate manifests, embedded AHS configurations, and Helm variants."""
import json
from pathlib import Path
import shutil
import subprocess

import jsonschema
import pytest
import yaml

ROOT = Path(__file__).resolve().parents[2]
CHART = ROOT / 'deploy/helm/agenthealth'
SCHEMA = json.loads((ROOT / 'spec/schemas/configuration.schema.json').read_text())
HELM = shutil.which('helm')


def pod_spec(doc):
    if doc['kind'] in ('Deployment', 'Job'):
        return doc['spec']['template']['spec']
    if doc['kind'] == 'CronJob':
        return doc['spec']['jobTemplate']['spec']['template']['spec']


def validate(doc):
    if doc['kind'] == 'ConfigMap' and 'config.yaml' in doc.get('data', {}):
        jsonschema.validate(yaml.safe_load(doc['data']['config.yaml']), SCHEMA)
    pod = pod_spec(doc)
    if not pod:
        return
    assert pod['automountServiceAccountToken'] is False
    assert pod['securityContext']['runAsNonRoot'] is True
    for c in pod['containers'] + pod.get('initContainers', []):
        assert c['securityContext']['readOnlyRootFilesystem'] is True
        assert c['securityContext']['allowPrivilegeEscalation'] is False
        assert c['securityContext']['capabilities']['drop'] == ['ALL']
        assert ':latest' not in c['image']
        for probe in ('startupProbe', 'readinessProbe'):
            if 'exec' in c.get(probe, {}):
                assert c[probe]['exec']['command'][0] == '/usr/local/bin/agenthealth'
                assert c[probe]['timeoutSeconds'] >= 10
        if c['name'] == 'agenthealth' and 'livenessProbe' in c:
            assert c['livenessProbe']['httpGet']['path'] == '/live'
    if doc['kind'] in ('Job', 'CronJob'):
        spec = doc['spec'] if doc['kind'] == 'Job' else doc['spec']['jobTemplate']['spec']
        assert spec['backoffLimit'] == 0
        assert spec['activeDeadlineSeconds'] == 30
        assert pod['restartPolicy'] == 'Never'


def test_manifests_and_embedded_configs():
    kinds = set()
    for directory in ('deploy/kubernetes', 'examples/kubernetes'):
        for path in (ROOT / directory).glob('*.yaml'):
            for doc in yaml.safe_load_all(path.read_text()):
                if isinstance(doc, dict) and 'kind' in doc:
                    kinds.add(doc['kind'])
                    validate(doc)
    assert {'Deployment', 'Job', 'CronJob', 'ConfigMap', 'Service'} <= kinds
    base_config = yaml.safe_load((ROOT / 'deploy/kubernetes/agenthealth.yaml').read_text())
    jsonschema.validate(base_config, SCHEMA)
    assert base_config == yaml.safe_load(yaml.safe_load((ROOT / 'deploy/kubernetes/configmap.yaml').read_text())['data']['config.yaml'])
    values = yaml.safe_load((CHART / 'values.yaml').read_text())
    jsonschema.validate(yaml.safe_load(values['config']), SCHEMA)
    jsonschema.validate(yaml.safe_load((ROOT / 'examples/kubernetes/agenthealth.yaml').read_text()), SCHEMA)


@pytest.mark.skipif(not HELM, reason='Helm required')
@pytest.mark.parametrize('mode', ['serve', 'job', 'cronjob'])
def test_chart_modes(mode, tmp_path):
    subprocess.run([HELM, 'lint', '--kube-version', '1.32.2', str(CHART), '--strict', '--set', f'mode={mode}'], check=True, capture_output=True)
    output = subprocess.check_output([HELM, 'template', '--kube-version', '1.32.2', 'review', str(CHART), '--set', f'mode={mode}'], text=True)
    docs = list(yaml.safe_load_all(output))
    assert len(docs) == 2
    for doc in docs:
        validate(doc)
    assert docs[-1]['kind'] == {'serve': 'Deployment', 'job': 'Job', 'cronjob': 'CronJob'}[mode]
    subprocess.run([HELM, 'package', str(CHART), '--destination', str(tmp_path)], check=True, capture_output=True)
    assert (tmp_path / 'agenthealth-0.1.0.tgz').exists()


@pytest.mark.skipif(not HELM, reason='Helm required')
def test_chart_secrets_existing_config_and_digest():
    output = subprocess.check_output([HELM, 'template', '--kube-version', '1.32.2', 'review', str(CHART), '-f', str(ROOT / 'examples/kubernetes/secret-values.yaml'), '--set', 'existingConfigMap=external', '--set', 'image.repository=docker.io/theagenthealth/agenthealth', '--set', 'image.digest=sha256:' + 'a' * 64], text=True)
    docs = list(yaml.safe_load_all(output))
    assert len(docs) == 1
    spec = pod_spec(docs[0])
    assert spec['volumes'][0]['configMap']['name'] == 'external'
    c = spec['containers'][0]
    assert c['image'] == 'docker.io/theagenthealth/agenthealth@sha256:' + 'a' * 64
    assert c['args'][-2:] == ['--token-env', 'AHP_TOKEN']
    assert all('secretKeyRef' in e['valueFrom'] and 'value' not in e for e in c['env'])


@pytest.mark.skipif(not HELM, reason='Helm required')
@pytest.mark.parametrize('setting', ['mode=invalid', 'image.tag=latest', 'image.digest=sha256:bad', 'ahpTokenEnv=AHP_TOKEN', 'secretEnv[0].name=INVALID-NAME'])
def test_chart_rejects_invalid_values(setting):
    result = subprocess.run([HELM, 'template', '--kube-version', '1.32.2', 'review', str(CHART), '--set', setting], capture_output=True)
    assert result.returncode != 0


@pytest.mark.skipif(not HELM, reason='Helm required')
def test_component_package_integrity_and_preconditions(tmp_path):
    import hashlib
    import sys
    import tarfile

    out = tmp_path / 'release'
    script = ROOT / 'scripts/package_helm.py'
    result = subprocess.run([sys.executable, str(script), 'helm-v9.9.9', '--output', str(out)], capture_output=True)
    assert result.returncode != 0
    assert not out.exists()
    subprocess.run([sys.executable, str(script), 'helm-v0.1.0', '--output', str(out)], check=True, capture_output=True)
    package = out / 'agenthealth-0.1.0.tgz'
    assert (out / 'checksums.txt').read_text() == hashlib.sha256(package.read_bytes()).hexdigest() + '  agenthealth-0.1.0.tgz\n'
    with tarfile.open(package) as archive:
        metadata = yaml.safe_load(archive.extractfile('agenthealth/Chart.yaml').read())
        assert metadata['version'] == '0.1.0'
        assert metadata['appVersion'] == 'v0.11.1'
        assert 'agenthealth/templates/workload.yaml' in archive.getnames()
    result = subprocess.run([sys.executable, str(script), 'helm-v0.1.0', '--output', str(out)], capture_output=True)
    assert result.returncode != 0


@pytest.mark.skipif(not HELM, reason='Helm required')
def test_token_must_reference_named_secret_entry():
    result = subprocess.run([HELM, 'template', '--kube-version', '1.32.2', 'review', str(CHART), '-f', str(ROOT / 'examples/kubernetes/secret-values.yaml'), '--set', 'ahpTokenEnv=UNMAPPED_TOKEN'], capture_output=True, text=True)
    assert result.returncode != 0
    assert 'ahpTokenEnv must reference a secretEnv entry' in result.stderr


@pytest.mark.skipif(not HELM, reason='Helm required')
def test_long_release_name_keeps_config_reference_valid():
    output = subprocess.check_output([HELM, 'template', '--kube-version', '1.32.2', 'a' * 53, str(CHART)], text=True)
    configmap, deployment = list(yaml.safe_load_all(output))
    assert len(configmap['metadata']['name']) <= 63
    assert len(deployment['metadata']['name']) <= 63
    assert pod_spec(deployment)['volumes'][0]['configMap']['name'] == configmap['metadata']['name']


@pytest.mark.skipif(not shutil.which('kubectl'), reason='kubectl required for Kustomize')
def test_kustomize_config_hash_and_workload_references():
    output = subprocess.check_output(['kubectl', 'kustomize', str(ROOT / 'examples/kubernetes')], text=True)
    docs = list(yaml.safe_load_all(output))
    configmap = next(d for d in docs if d['kind'] == 'ConfigMap' and d['metadata']['name'].startswith('agenthealth-config-'))
    jsonschema.validate(yaml.safe_load(configmap['data']['config.yaml']), SCHEMA)
    for doc in docs:
        validate(doc)
        spec = pod_spec(doc)
        if spec:
            for volume in spec.get('volumes', []):
                name = volume.get('configMap', {}).get('name', '')
                if name.startswith('agenthealth-config'):
                    assert name == configmap['metadata']['name']
    assert all(c['image'] == 'ghcr.io/theagenthealth/agenthealth:v0.11.1'
               for doc in docs if pod_spec(doc)
               for c in pod_spec(doc)['containers'] + pod_spec(doc).get('initContainers', [])
               if c['name'] == 'agenthealth')


@pytest.mark.skipif(not HELM, reason='Helm required')
@pytest.mark.parametrize('name', ['true', 'null', '123'])
def test_yaml_scalar_names_remain_strings(name):
    output = subprocess.check_output([HELM, 'template', '--kube-version', '1.32.2', name, str(CHART), '--set-string', 'existingConfigMap=' + name], text=True)
    doc = next(yaml.safe_load_all(output))
    assert doc['spec']['selector']['matchLabels']['app.kubernetes.io/instance'] == name
    assert doc['spec']['template']['metadata']['labels']['app.kubernetes.io/instance'] == name
    assert pod_spec(doc)['volumes'][0]['configMap']['name'] == name


def test_exec_readiness_has_bounded_dependencies_and_process_startup():
    config = yaml.safe_load((ROOT / 'examples/kubernetes/agenthealth.yaml').read_text())
    embedded = yaml.safe_load(yaml.safe_load((ROOT / 'examples/kubernetes/configmap.yaml').read_text())['data']['config.yaml'])
    assert config == embedded
    pod = pod_spec(yaml.safe_load((ROOT / 'examples/kubernetes/exec-probes.yaml').read_text()))
    container = pod['containers'][0]
    assert container['startupProbe']['httpGet']['path'] == '/live'
    assert container['readinessProbe']['exec']['command'][1] == 'check'
    for target in config['targets']:
        for node in [target] + target['dependencies']:
            assert node['budget_ms'] <= 5000
            assert node['timeout_ms'] <= 2000
    assert container['readinessProbe']['timeoutSeconds'] >= 10
