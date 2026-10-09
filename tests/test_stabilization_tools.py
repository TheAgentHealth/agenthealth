"""Exercise weighted coverage gates and reviewable skip reports."""
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]


def test_coverage_gate_weights_statements_and_excludes_examples(tmp_path):
    profile = tmp_path / 'coverage.out'
    profile.write_text('mode: atomic\n'
                       'example/core/a.go:1.1,2.1 9 1\n'
                       'example/core/a.go:3.1,4.1 1 0\n'
                       'example/examples/demo/main.go:1.1,2.1 100 0\n')
    command = [sys.executable, str(ROOT / 'scripts/check_coverage.py'), str(profile)]
    assert subprocess.run(command + ['--minimum', '90'], capture_output=True).returncode == 0
    assert subprocess.run(command + ['--minimum', '91'], capture_output=True).returncode == 1
    profile.write_text('mode: atomic\n')
    assert subprocess.run(command, capture_output=True).returncode != 0


def test_skip_report_groups_reasons_and_keeps_test_identity(tmp_path):
    report = tmp_path / 'pytest.xml'
    report.write_text('<testsuites><testsuite>'
                     '<testcase classname="tests.kube" name="helm"><skipped message="Helm required"/></testcase>'
                     '<testcase classname="tests.kube" name="chart"><skipped message="Helm required"/></testcase>'
                     '<testcase classname="tests.core" name="passed"/>'
                     '</testsuite></testsuites>')
    output = tmp_path / 'skips.md'
    subprocess.run([sys.executable, str(ROOT / 'scripts/summarize_test_skips.py'), str(report), str(output)], check=True)
    text = output.read_text()
    assert '2 skipped tests.' in text
    assert '| Helm required | 2 |' in text
    assert '`tests.kube.helm`' in text
    assert '`tests.kube.chart`' in text
    assert 'tests.core.passed' not in text
