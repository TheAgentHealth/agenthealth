"""Independently maintained filesystem server, passive inventories only."""
import json
import os
from pathlib import Path
import subprocess
import pytest

pytestmark = pytest.mark.skipif(os.environ.get('AGENTHEALTH_REAL_MCP') != '1',
                               reason='real MCP product suite runs separately')


def test_filesystem_server(tmp_path):
    server = Path(os.environ['AGENTHEALTH_FILESYSTEM_SERVER']).resolve()
    config = tmp_path / 'health.json'
    config.write_text(json.dumps({'version': 'v1', 'targets': [{
        'name': 'filesystem', 'type': 'mcp', 'endpoint': 'stdio://filesystem',
        'checks': ['protocol', 'capability'],
        'mcp': {'protocol_version': '2025-11-25', 'transport': 'stdio',
                'required_tools': ['read_file', 'list_directory'],
                'stdio': {'command': 'node', 'args': [str(server), str(tmp_path)]}}
    }]}))
    run = subprocess.run([os.environ['AGENTHEALTH_BINARY'], 'check', str(config), '--format', 'json'],
                         capture_output=True, text=True, timeout=30)
    assert run.returncode == 0, run.stdout + run.stderr
    assert json.loads(run.stdout)['status'] == 'HEALTHY'
