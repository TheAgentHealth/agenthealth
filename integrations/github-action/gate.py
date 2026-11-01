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
    created_report = False
    try:
        workspace = Path(os.environ['GITHUB_WORKSPACE']).resolve()
        # Normalize '..' without following report symlinks. Check every component
        # before unlinking stale evidence or creating output directories.
        report = Path(os.path.abspath(workspace / os.environ['AH_REPORT']))
        if not report.is_relative_to(workspace) or report == workspace:
            raise ValueError('report path')
        for component in (report, *report.parents):
            if component == workspace:
                break
            if component.is_symlink():
                raise ValueError('symlinked report path')
        configuration = (workspace / os.environ['AH_CONFIGURATION']).resolve()
        if report == configuration or (report.exists() and configuration.exists() and report.samefile(configuration)):
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
        with report.open('xb') as output:
            created_report = True
            result = subprocess.run([binary, 'check', str(configuration), '--format', 'junit'],
                                    stdout=output, stderr=subprocess.DEVNULL, timeout=timeout)
        code = result.returncode if result.returncode in range(6) else 6
        if code == 6 or not report.stat().st_size:
            code = 6
            report.unlink(missing_ok=True)
    except (OSError, ValueError, KeyError, subprocess.SubprocessError):
        code = 6
        # Only remove output created by this invocation; rejected paths must never
        # trigger cleanup of preexisting files. Cleanup failures remain tool errors.
        if created_report:
            try:
                report.unlink(missing_ok=True)
            except OSError:
                pass
    with open(os.environ['GITHUB_OUTPUT'], 'a', encoding='utf-8') as output:
        output.write(f'exit-code={code}\n')
    if code:
        print(f'::error title=AgentHealth::Deployment gate failed (exit {code}); inspect the report if present.')
    else:
        print('::notice title=AgentHealth::Deployment gate passed.')
    return code


if __name__ == '__main__':
    sys.exit(main())
