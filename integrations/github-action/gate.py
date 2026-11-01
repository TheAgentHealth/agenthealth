#!/usr/bin/env python3
"""GitHub Action runner; never interpret workflow inputs as shell code."""
import os
from pathlib import Path
import re
import subprocess
import sys


def main():
    code = 6
    report = None
    try:
        workspace = Path(os.environ['GITHUB_WORKSPACE']).resolve()
        report = (workspace / os.environ['AH_REPORT']).resolve()
        if not report.is_relative_to(workspace) or report == workspace:
            raise ValueError('report path')
        configuration = (workspace / os.environ['AH_CONFIGURATION']).resolve()
        if report == configuration:
            raise ValueError('report overlaps configuration')
        # Remove previous evidence before running; a failed tool cannot leave a stale pass.
        report.unlink(missing_ok=True)
        expected = os.environ['AH_VERSION']
        if not re.fullmatch(r'v[0-9]+\.[0-9]+\.[0-9]+', expected):
            raise ValueError('version')
        timeout = int(os.environ['AH_TIMEOUT'])
        if not 1 <= timeout <= 3600:
            raise ValueError('timeout')
        binary = os.environ['AH_BINARY']
        version = subprocess.run([binary, 'version'], capture_output=True, timeout=10, check=True)
        if version.stdout.decode().strip() != 'agenthealth ' + expected:
            raise ValueError('CLI version mismatch')
        report.parent.mkdir(parents=True, exist_ok=True)
        with report.open('wb') as output:
            result = subprocess.run([binary, 'check', str(configuration), '--format', 'junit'],
                                    stdout=output, stderr=subprocess.DEVNULL, timeout=timeout)
        code = result.returncode if result.returncode in range(6) else 6
        if code == 6 or not report.stat().st_size:
            code = 6
            report.unlink(missing_ok=True)
    except (OSError, ValueError, KeyError, subprocess.SubprocessError):
        # Keep exception text, command arguments and credentials out of workflow commands.
        if report is not None and report.is_relative_to(Path(os.environ['GITHUB_WORKSPACE']).resolve()) and report.is_file() and report != locals().get('configuration'):
            report.unlink(missing_ok=True)
    with open(os.environ['GITHUB_OUTPUT'], 'a', encoding='utf-8') as output:
        output.write(f'exit-code={code}\n')
    if code:
        print(f'::error title=AgentHealth::Deployment gate failed (exit {code}); inspect the report if present.')
    else:
        print('::notice title=AgentHealth::Deployment gate passed.')
    return code


if __name__ == '__main__':
    sys.exit(main())
