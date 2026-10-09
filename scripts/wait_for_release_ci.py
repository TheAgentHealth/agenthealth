"""Require successful pre-release CI and native binary jobs for a release commit."""
import argparse
import json
import os
import subprocess
import time
import uuid

PLATFORMS = ('ubuntu-latest', 'ubuntu-24.04-arm', 'macos-15-intel',
             'macos-latest', 'windows-latest')


def api(path):
    return json.loads(subprocess.check_output(['gh', 'api', path], text=True))


def check_ci(repository, sha, require_kubernetes=False, request_id=None):
    query = '?branch=main&event=workflow_dispatch&per_page=100'
    if request_id is None:
        query += f'&head_sha={sha}'
    runs = api(f'repos/{repository}/actions/workflows/ci.yml/runs' + query)['workflow_runs']
    runs = [run for run in runs if (run.get('display_title') == f'Pre-release CI {sha} {request_id}'
                              if request_id else run['head_sha'] == sha)
            and run['head_branch'] == 'main' and run['event'] == 'workflow_dispatch']
    if not runs:
        return False
    run = max(runs, key=lambda item: item['id'])
    if run['status'] != 'completed':
        return False
    if run['conclusion'] != 'success':
        raise RuntimeError(f"Main CI did not succeed: {run['html_url']}")
    jobs = []
    page = 1
    while True:
        batch = api(f"repos/{repository}/actions/runs/{run['id']}/attempts/"
                    f"{run['run_attempt']}/jobs?per_page=100&page={page}")['jobs']
        jobs.extend(batch)
        if len(batch) < 100:
            break
        page += 1
    required_jobs = [f'Binary distribution ({platform})' for platform in PLATFORMS]
    if require_kubernetes:
        required_jobs.append('Kubernetes manifests and Helm')
    for name in required_jobs:
        matches = [job for job in jobs if job['name'] == name]
        if len(matches) != 1 or matches[0]['conclusion'] != 'success':
            raise RuntimeError(f"Missing or unsuccessful {name}: {run['html_url']}")
    print(f"Required pre-release CI succeeded: {run['html_url']}", flush=True)
    return True


def wait_for_ci(repository, sha, timeout=4500, require_kubernetes=False, request_id=None):
    deadline = time.monotonic() + timeout
    while True:
        if check_ci(repository, sha, require_kubernetes=require_kubernetes, request_id=request_id):
            return
        if time.monotonic() >= deadline:
            raise RuntimeError(f'No completed pre-release CI for {sha} within {timeout}s')
        print(f'Waiting for pre-release CI on {sha}...', flush=True)
        time.sleep(min(30, max(0, deadline - time.monotonic())))


def dispatch_ci(repository, sha, require_kubernetes=False):
    request_id = uuid.uuid4().hex
    payload = {'ref': 'main', 'inputs': {'release_sha': sha, 'request_id': request_id,
                                        'require_kubernetes': require_kubernetes}}
    subprocess.run(['gh', 'api', '--method', 'POST',
                    f'repos/{repository}/actions/workflows/ci.yml/dispatches',
                    '--input', '-'], input=json.dumps(payload), text=True, check=True)
    return request_id


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('sha')
    parser.add_argument('--require-kubernetes', action='store_true',
                        help='Require successful Kubernetes runtime CI for synchronized releases')
    parser.add_argument('--dispatch', action='store_true', help='Start a fresh pre-release CI run on main')
    args = parser.parse_args()
    repository = os.environ['GITHUB_REPOSITORY']
    request_id = dispatch_ci(repository, args.sha, args.require_kubernetes) if args.dispatch else None
    wait_for_ci(repository, args.sha, require_kubernetes=args.require_kubernetes,
                request_id=request_id)
