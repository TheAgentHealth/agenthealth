#!/usr/bin/env python3
"""Exercise Phase 14 on the explicitly selected disposable Kubernetes context.

Requires kubectl and helm; never selects or creates a cluster implicitly.
"""
import argparse
import json
import pathlib
import subprocess
import tempfile
import time

import yaml

ROOT = pathlib.Path(__file__).resolve().parents[1]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--context', required=True)
    parser.add_argument('--image', help='Optional locally built AgentHealth repository:tag loaded into the cluster')
    args = parser.parse_args()
    if args.image and ('@' in args.image or ':' not in args.image or not all(args.image.rsplit(':', 1))):
        parser.error('--image requires an explicit repository:tag')
    namespace = 'agenthealth-smoke-' + str(int(time.time()))

    def kube(*arguments, check=True):
        result = subprocess.run(['kubectl', '--context', args.context, '-n', namespace, *arguments], capture_output=True, text=True, timeout=150)
        if check and result.returncode:
            raise RuntimeError(f'kubectl {arguments[0]} failed: {result.stderr}')
        return result

    def apply(path):
        documents = list(yaml.safe_load_all((ROOT / path).read_text()))
        if args.image:
            def replace(value):
                if isinstance(value, dict):
                    if str(value.get('image', '')).startswith('ghcr.io/theagenthealth/agenthealth:'):
                        value['image'] = args.image
                    for child in value.values():
                        replace(child)
                elif isinstance(value, list):
                    for child in value:
                        replace(child)
            replace(documents)
        with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml') as manifest:
            yaml.safe_dump_all(documents, manifest, sort_keys=False)
            manifest.flush()
            kube('apply', '-f', manifest.name)

    def wait_readiness(ready, selector='app=agenthealth'):
        deadline = time.monotonic() + 100
        while time.monotonic() < deadline:
            pods = json.loads(kube('get', 'pods', '-l', selector, '-o', 'json').stdout)['items']
            if pods and all(any(c['status'] == 'True' for c in pod['status'].get('conditions', []) if c['type'] == 'Ready') == ready for pod in pods):
                assert all(c['restartCount'] == 0 for pod in pods for c in pod['status']['containerStatuses'])
                return
            time.sleep(2)
        raise AssertionError(f'Pod readiness did not become {ready}: {selector}')

    def wait_job(name, success=True):
        kube('wait', '--for=condition=' + ('complete' if success else 'failed'), 'job/' + name, '--timeout=90s')

    kube('create', 'namespace', namespace)
    try:
        apply('examples/kubernetes/demo.yaml')
        kube('rollout', 'status', 'deployment/demo', '--timeout=90s')
        apply('examples/kubernetes/configmap.yaml')
        # Contributor scenario: init must block while its external dependency is absent.
        kube('scale', 'deployment/demo', '--replicas=0')
        kube('wait', '--for=delete', 'pod', '-l', 'app=demo', '--timeout=60s')
        apply('examples/kubernetes/init.yaml')
        deadline = time.monotonic() + 60
        while time.monotonic() < deadline:
            pods = json.loads(kube('get', 'pods', '-l', 'app=demo-init', '-o', 'json').stdout)['items']
            if pods:
                status = pods[0]['status']
                init = status.get('initContainerStatuses', [])
                if init and (init[0]['state'].get('terminated', {}).get('exitCode', 0) != 0 or init[0]['lastState'].get('terminated', {}).get('exitCode', 0) != 0):
                    assert not any('running' in c['state'] for c in status.get('containerStatuses', []))
                    break
            time.sleep(1)
        else:
            raise AssertionError('Init container did not block startup during dependency outage')
        # Process startup must succeed even while dependency readiness is failing.
        apply('deploy/kubernetes/deployment.yaml')
        apply('examples/kubernetes/exec-probes.yaml')
        deadline = time.monotonic() + 75
        while time.monotonic() < deadline:
            pod_groups = [json.loads(kube('get', 'pods', '-l', selector, '-o', 'json').stdout)['items']
                          for selector in ('app=agenthealth', 'app=agenthealth-exec')]
            if all(pods and all(c.get('started', False) for pod in pods for c in pod['status'].get('containerStatuses', []))
                   and all(pod['status'].get('containerStatuses') for pod in pods) for pods in pod_groups):
                break
            time.sleep(1)
        else:
            raise AssertionError('Monitor process startup was blocked by a dependency outage')
        wait_readiness(False)
        wait_readiness(False, 'app=agenthealth-exec')
        kube('scale', 'deployment/demo', '--replicas=1')
        kube('rollout', 'status', 'deployment/demo', '--timeout=90s')
        kube('rollout', 'status', 'deployment/demo-init', '--timeout=90s')
        apply('deploy/kubernetes/job.yaml')
        wait_job('agenthealth-check')
        result = json.loads(kube('logs', 'job/agenthealth-check').stdout)
        assert result['status'] == 'HEALTHY', result
        assert len(result['dependencies']) == 4
        # Synthetic fixture credential: verify injection and rejection without logging it.
        kube('create', 'secret', 'generic', 'agenthealth-credentials', '--from-literal=agent-token=fixture-token')
        apply('examples/kubernetes/secret-job.yaml')
        wait_job('agenthealth-auth')
        assert 'fixture-token' not in kube('logs', 'job/agenthealth-auth').stdout
        kube('delete', 'job', 'agenthealth-auth')
        kube('delete', 'secret', 'agenthealth-credentials')
        kube('create', 'secret', 'generic', 'agenthealth-credentials', '--from-literal=agent-token=rejected-fixture-token')
        apply('examples/kubernetes/secret-job.yaml')
        wait_job('agenthealth-auth', success=False)
        auth_result = kube('logs', 'job/agenthealth-auth').stdout
        assert json.loads(auth_result)['status'] == 'MISCONFIGURED'
        assert 'rejected-fixture-token' not in auth_result
        for filename, name in [('deploy/kubernetes/deployment.yaml', 'agenthealth'), ('examples/kubernetes/sidecar.yaml', 'demo-sidecar'), ('examples/kubernetes/init.yaml', 'demo-init'), ('examples/kubernetes/exec-probes.yaml', 'agenthealth-exec')]:
            apply(filename)
            kube('rollout', 'status', 'deployment/' + name, '--timeout=90s')
        apply('deploy/kubernetes/cronjob.yaml')
        kube('create', 'job', 'scheduled-check', '--from=cronjob/agenthealth-check')
        wait_job('scheduled-check')
        for mode in ('serve', 'job', 'cronjob'):
            command = ['helm', 'template', '--kube-version', '1.32.2', 'chart-' + mode, str(ROOT / 'deploy/helm/agenthealth'), '--set', 'mode=' + mode]
            if args.image:
                repository, tag = args.image.rsplit(':', 1)
                command.extend(['--set', 'image.repository=' + repository, '--set', 'image.tag=' + tag])
            rendered = subprocess.check_output(command, text=True)
            with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml') as manifest:
                manifest.write(rendered)
                manifest.flush()
                kube('apply', '-f', manifest.name)
            name = 'chart-' + mode + '-agenthealth'
            if mode == 'serve':
                kube('rollout', 'status', 'deployment/' + name, '--timeout=90s')
            elif mode == 'job':
                wait_job(name)
            else:
                kube('create', 'job', 'chart-scheduled', '--from=cronjob/' + name)
                wait_job('chart-scheduled')
        # A nonzero CLI health exit fails a Job without retry and removes readiness.
        kube('scale', 'deployment/demo', '--replicas=0')
        kube('wait', '--for=delete', 'pod', '-l', 'app=demo', '--timeout=60s')
        kube('delete', 'job', 'agenthealth-check')
        apply('deploy/kubernetes/job.yaml')
        wait_job('agenthealth-check', success=False)
        wait_readiness(False)
        wait_readiness(False, 'app=agenthealth-exec')
        # Verify recovery, including exec readiness, without process restarts.
        kube('scale', 'deployment/demo', '--replicas=1')
        kube('rollout', 'status', 'deployment/demo', '--timeout=90s')
        wait_readiness(True)
        wait_readiness(True, 'app=agenthealth-exec')
        print('Kubernetes smoke passed: init outage/recovery, probes, sidecar, Secrets, Jobs, CronJobs, Helm and readiness recovery')
    finally:
        kube('delete', 'namespace', namespace, '--wait=false', check=False)


if __name__ == '__main__':
    main()
