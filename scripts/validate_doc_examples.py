#!/usr/bin/env python3
"""Validates fenced code examples in documentation that are explicitly marked
for validation against spec/schemas/*.json.

This closes the gap the pytest fixture suite (tests/spec/) cannot: fixtures
only prove the schemas behave correctly in isolation, they do not prove that
the prose examples in README.md/ROADMAP.md/spec/*.md actually stay in sync
with those schemas. Doc authors mark each fenced example explicitly:

    <!-- spec-example: result -->
    ```json
    { ... }
    ```

    <!-- spec-example: configuration -->
    ```yaml
    ...
    ```

    <!-- spec-example: skip reason="why this one is intentionally exempt" -->
    ```json
    { ... }
    ```

Any ```json/```yaml fenced block NOT preceded by one of these markers is
simply ignored by this script (it is not assumed to be a spec example at
all) - only explicitly marked blocks are checked, and skips must carry a
reason so exemptions are intentional and documented, not silent.
"""
import json
import pathlib
import re
import sys

import jsonschema
import yaml

ROOT = pathlib.Path(__file__).resolve().parents[1]
SCHEMAS_DIR = ROOT / "spec" / "schemas"

RESULT_SCHEMA = json.loads((SCHEMAS_DIR / "result.schema.json").read_text())
CONFIGURATION_SCHEMA = json.loads((SCHEMAS_DIR / "configuration.schema.json").read_text())

SCHEMAS_BY_KIND = {
    "result": RESULT_SCHEMA,
    "configuration": CONFIGURATION_SCHEMA,
}

MARKER_RE = re.compile(
    r"<!--\s*spec-example:\s*(result|configuration|skip)"
    r"(?:\s+reason=\"([^\"]*)\")?\s*-->\s*\n"
    r"```(\w*)\n(.*?)\n```",
    re.DOTALL,
)

FILES_TO_CHECK = [
    ROOT / "README.md",
    ROOT / "ROADMAP.md",
    ROOT / "spec" / "result-schema.md",
    ROOT / "spec" / "configuration.md",
    ROOT / "spec" / "protocol.md",
    ROOT / "docs" / "http-adapter.md",
    ROOT / "docs" / "mcp-adapter.md",
    ROOT / "docs" / "a2a-adapter.md",
    ROOT / "examples" / "README.md",
]


def _load_block(lang: str, body: str):
    if lang in ("yaml", "yml"):
        return yaml.safe_load(body)
    return json.loads(body)


def main() -> int:
    errors: list[str] = []
    checked = 0
    skipped = 0

    for path in FILES_TO_CHECK:
        text = path.read_text(encoding="utf-8")
        for match in MARKER_RE.finditer(text):
            kind, reason, lang, body = match.groups()

            if kind == "skip":
                if not reason:
                    errors.append(
                        f"{path}: 'spec-example: skip' with no reason=\"...\" - "
                        "exemptions must be documented, not silent"
                    )
                else:
                    skipped += 1
                continue

            checked += 1
            try:
                data = _load_block(lang, body)
            except Exception as exc:  # noqa: BLE001 - report any parse failure
                errors.append(f"{path}: failed to parse '{kind}' example as {lang or 'JSON'}: {exc}")
                continue

            schema = SCHEMAS_BY_KIND[kind]
            try:
                jsonschema.validate(data, schema)
            except jsonschema.ValidationError as exc:
                errors.append(f"{path}: '{kind}' example failed schema validation: {exc.message}")

    # Standalone configuration templates are also user-facing examples.
    configs = sorted((ROOT / "examples").rglob("*.yaml"))
    configs += sorted((ROOT / "integrations").rglob("agenthealth.yaml"))
    for path in configs:
        try:
            data = yaml.safe_load(path.read_text(encoding="utf-8"))
            jsonschema.validate(data, CONFIGURATION_SCHEMA)
        except (yaml.YAMLError, jsonschema.ValidationError, OSError) as exc:
            errors.append(f"{path}: configuration example failed validation: {exc}")

    if errors:
        print(f"Found {len(errors)} documentation example issue(s):")
        for err in errors:
            print(f"  - {err}")
        return 1

    print(f"All {checked} marked documentation example(s) validated OK ({skipped} explicitly exempt).")
    print(f"All {len(configs)} standalone configuration example(s) validated OK.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
