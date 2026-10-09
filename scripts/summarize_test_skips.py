#!/usr/bin/env python3
"""Turn pytest JUnit skip reasons into a reviewable CI artifact and job summary."""
import argparse
from collections import Counter
from pathlib import Path
import xml.etree.ElementTree as ET

parser = argparse.ArgumentParser()
parser.add_argument('report', type=Path)
parser.add_argument('output', type=Path)
args = parser.parse_args()
root = ET.parse(args.report).getroot()
counts = Counter()
rows = []
for test in root.iter('testcase'):
    skipped = test.find('skipped')
    if skipped is not None:
        reason = skipped.get('message', 'unspecified').replace('\n', ' ').replace('|', '\\|')
        counts[reason] += 1
        rows.append(f'- `{test.get("classname")}.{test.get("name")}`: {reason}')
lines = ['# Test skips', '', f'{len(rows)} skipped tests.', '', '| Reason | Count |', '|---|---|']
lines += [f'| {reason} | {count} |' for reason, count in sorted(counts.items())]
lines += ['', '## Tests', ''] + rows
args.output.parent.mkdir(parents=True, exist_ok=True)
args.output.write_text('\n'.join(lines) + '\n')
