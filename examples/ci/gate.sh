#!/usr/bin/env sh
# Use an exact-version, provenance-verified CLI already on PATH.
# Preserve every health failure and code 6; do not pipe through tee.
set -eu
configuration=${1:?Usage: gate.sh configuration.yaml [report.xml]}
report=${2:-agenthealth-junit.xml}
if [ "$configuration" = "$report" ] || [ "$configuration" -ef "$report" ]; then
  echo 'AgentHealth report must not overwrite configuration' >&2
  exit 6
fi
if [ -L "$report" ]; then
  echo 'AgentHealth report must not be a symlink' >&2
  exit 6
fi
exec agenthealth check "$configuration" --format junit > "$report"
