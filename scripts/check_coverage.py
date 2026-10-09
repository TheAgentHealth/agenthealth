#!/usr/bin/env python3
"""Gate statement coverage for production packages, excluding executable examples."""
import argparse
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument('profile', type=Path)
parser.add_argument('--minimum', type=float, default=80)
args = parser.parse_args()
covered = total = 0
for line in args.profile.read_text().splitlines()[1:]:
    location, statements, count = line.rsplit(' ', 2)
    if '/examples/' in location:
        continue
    n = int(statements)
    total += n
    if int(count) > 0:
        covered += n
if not total:
    raise SystemExit('No production coverage data')
percent = 100 * covered / total
print(f'Production statement coverage: {percent:.2f}% (minimum {args.minimum:.2f}%)')
if percent < args.minimum:
    raise SystemExit(1)
