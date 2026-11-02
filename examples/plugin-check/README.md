# Local executable adapter fixture

From the repository root:

```sh
go build -o /tmp/agenthealth-host ./cmd/agenthealth
go build -o /tmp/agenthealth-probe ./examples/plugin-check
python3 - <<'PY'
import hashlib, json
from pathlib import Path
binary = Path('/tmp/agenthealth-probe').resolve()
manifest = {'protocol': 'v1', 'command': str(binary),
            'sha256': hashlib.sha256(binary.read_bytes()).hexdigest(),
            'metadata': {'name': 'protocol-fixture', 'version': '1.0.0',
                         'compatibility_version': 'v1', 'target_types': ['custom'],
                         'dimensions': ['configuration', 'reachability'], 'active_checks': []}}
Path('/tmp/agenthealth-adapter.json').write_text(json.dumps(manifest))
PY
/tmp/agenthealth-host check examples/plugin-check/agenthealth.yaml --adapter /tmp/agenthealth-adapter.json --format json
```

The fixture returns HEALTHY for configuration and UNKNOWN for reachability:
it performs no network operation and makes no remote health claim. Expected health
exit code is 5. Adapt the output path for your OS; Windows uses an `.exe` suffix.

The direct/composite names identify agent roles. Dependencies record independently
observed health, not verified communication. Add an explicit peer/path target to
measure that path. Keep built-in gateway/router, A2A and backend checks separate;
see the [extension guide](../../docs/plugins.md). Do not forward arbitrary input,
log credentials or return response bodies. Real probes must implement the safety
and response-marker rules in the [contract](../../docs/rfcs/phase-23-adapters.md).
