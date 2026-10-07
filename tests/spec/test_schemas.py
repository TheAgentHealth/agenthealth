"""Validates spec/schemas/*.json against the fixtures in tests/spec/fixtures/.

Valid fixtures MUST pass; invalid fixtures MUST fail. This is what would have
caught every schema-related finding from the Phase 1 specification reviews
(unknown target types, negative latency, empty batches, flattened dependency
shapes) before they reached docs.
"""
import json
import pathlib

import jsonschema
import pytest
import yaml

ROOT = pathlib.Path(__file__).resolve().parents[2]
SCHEMAS_DIR = ROOT / "spec" / "schemas"
FIXTURES_DIR = pathlib.Path(__file__).resolve().parent / "fixtures"

RESULT_SCHEMA = json.loads((SCHEMAS_DIR / "result.schema.json").read_text())
CONFIGURATION_SCHEMA = json.loads((SCHEMAS_DIR / "configuration.schema.json").read_text())


def _load(path: pathlib.Path):
    if path.suffix in (".yaml", ".yml"):
        return yaml.safe_load(path.read_text())
    return json.loads(path.read_text())


def _fixtures(kind: str, prefix: str) -> list[pathlib.Path]:
    directory = FIXTURES_DIR / kind
    return sorted(directory.glob(f"{prefix}.*"))


VALID_RESULT_FIXTURES = _fixtures("valid", "result")
INVALID_RESULT_FIXTURES = _fixtures("invalid", "result")
VALID_CONFIG_FIXTURES = _fixtures("valid", "configuration")
INVALID_CONFIG_FIXTURES = _fixtures("invalid", "configuration")


def test_schemas_are_valid_json_schema():
    jsonschema.Draft202012Validator.check_schema(RESULT_SCHEMA)
    jsonschema.Draft202012Validator.check_schema(CONFIGURATION_SCHEMA)


def test_fixture_directories_are_not_empty():
    assert VALID_RESULT_FIXTURES, "expected at least one valid result fixture"
    assert INVALID_RESULT_FIXTURES, "expected at least one invalid result fixture"
    assert VALID_CONFIG_FIXTURES, "expected at least one valid configuration fixture"
    assert INVALID_CONFIG_FIXTURES, "expected at least one invalid configuration fixture"


@pytest.mark.parametrize("path", VALID_RESULT_FIXTURES, ids=lambda p: p.name)
def test_valid_result_fixtures_pass(path: pathlib.Path):
    jsonschema.validate(_load(path), RESULT_SCHEMA)


@pytest.mark.parametrize("path", INVALID_RESULT_FIXTURES, ids=lambda p: p.name)
def test_invalid_result_fixtures_fail(path: pathlib.Path):
    with pytest.raises(jsonschema.ValidationError):
        jsonschema.validate(_load(path), RESULT_SCHEMA)


@pytest.mark.parametrize("path", VALID_CONFIG_FIXTURES, ids=lambda p: p.name)
def test_valid_configuration_fixtures_pass(path: pathlib.Path):
    jsonschema.validate(_load(path), CONFIGURATION_SCHEMA)


@pytest.mark.parametrize("path", INVALID_CONFIG_FIXTURES, ids=lambda p: p.name)
def test_invalid_configuration_fixtures_fail(path: pathlib.Path):
    with pytest.raises(jsonschema.ValidationError):
        jsonschema.validate(_load(path), CONFIGURATION_SCHEMA)
