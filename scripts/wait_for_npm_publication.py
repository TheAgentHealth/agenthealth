#!/usr/bin/env python3
"""Dispatch the npm-trusted workflow and gate final release publication on it."""
import json
import os
import re
import subprocess
import sys
import time


def publish(tag, request_id, timeout=1200):
    if not re.fullmatch(r'v\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?', tag):
        raise ValueError('expected semantic version tag')
    if not re.fullmatch(r'[0-9]+-[0-9]+', request_id):
        raise ValueError('expected release run-attempt identity')
    repository = os.environ['GITHUB_REPOSITORY']
    path = f'repos/{repository}/actions/workflows/publish.yml'
    payload = {'ref': tag, 'inputs': {'release_tag': tag, 'request_id': request_id}}
    subprocess.run(['gh', 'api', '--method', 'POST', path + '/dispatches', '--input', '-'],
                   input=json.dumps(payload), text=True, check=True)
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        response = json.loads(subprocess.check_output(
            ['gh', 'api', path + '/runs?branch=' + tag + '&event=workflow_dispatch&per_page=100'], text=True))
        matches = [r for r in response['workflow_runs']
                   if r['display_title'] == f'npm {tag} {request_id}' and r['head_branch'] == tag
                   and r['event'] == 'workflow_dispatch']
        if matches:
            run = max(matches, key=lambda r: r['id'])
            if run['status'] == 'completed':
                if run['conclusion'] != 'success':
                    raise RuntimeError('npm publication failed: ' + run['html_url'])
                print('Verified npm publication: ' + run['html_url'], flush=True)
                return
        print('Waiting for trusted npm publication...', flush=True)
        time.sleep(15)
    raise RuntimeError('npm publication timed out; release remains staged')


if __name__ == '__main__':
    publish(sys.argv[1], sys.argv[2])
