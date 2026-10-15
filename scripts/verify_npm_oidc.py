#!/usr/bin/env python3
"""Diagnose npm trusted-publisher identity without printing credentials or JWTs."""
import base64
import json
import os
import urllib.error
import urllib.parse
import urllib.request


def verify():
    request_url = os.environ['ACTIONS_ID_TOKEN_REQUEST_URL']
    request_token = os.environ['ACTIONS_ID_TOKEN_REQUEST_TOKEN']
    parts = urllib.parse.urlsplit(request_url)
    query = urllib.parse.parse_qsl(parts.query)
    query.append(('audience', 'npm:registry.npmjs.org'))
    url = urllib.parse.urlunsplit(parts._replace(query=urllib.parse.urlencode(query)))
    request = urllib.request.Request(url, headers={'Authorization': 'Bearer ' + request_token})
    with urllib.request.urlopen(request, timeout=30) as response:
        token = json.load(response)['value']
    payload = token.split('.')[1]
    claims = json.loads(base64.urlsafe_b64decode(payload + '=' * (-len(payload) % 4)))
    allowed = ('repository', 'repository_owner', 'repository_visibility', 'workflow_ref',
               'job_workflow_ref', 'ref', 'aud', 'runner_environment', 'environment')
    print('OIDC public identity: ' + json.dumps({key: claims[key] for key in allowed if key in claims}), flush=True)
    request = urllib.request.Request(
        'https://registry.npmjs.org/-/npm/v1/oidc/token/exchange/package/@theagenthealth%2fagenthealth',
        data=b'', method='POST', headers={'Authorization': 'Bearer ' + token})
    try:
        with urllib.request.urlopen(request, timeout=30) as response:
            document = json.load(response)
        if not document.get('token'):
            raise RuntimeError('npm token exchange returned no publishing token')
    except urllib.error.HTTPError as error:
        try:
            document = json.loads(error.read(16384))
        except (ValueError, UnicodeDecodeError):
            document = {}
        message = str(document.get('message') or document.get('error') or 'No registry diagnostic')[:1000]
        message = message.replace(token, '[redacted]').replace(request_token, '[redacted]')
        raise RuntimeError(f'npm OIDC exchange rejected (HTTP {error.code}): {message}') from None
    print('npm OIDC token exchange succeeded; no token was logged', flush=True)


if __name__ == '__main__':
    verify()
