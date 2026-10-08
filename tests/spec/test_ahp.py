"""AHP envelope fixtures validate independently of the Go server."""
import json
from pathlib import Path

import jsonschema
import pytest
from referencing import Registry, Resource

ROOT = Path(__file__).resolve().parents[2]
SCHEMAS = ROOT / 'spec' / 'schemas'
AHP = json.loads((SCHEMAS / 'ahp.schema.json').read_text())
RESULT = json.loads((SCHEMAS / 'result.schema.json').read_text())
REGISTRY = Registry().with_resource(RESULT['$id'], Resource.from_contents(RESULT))
VALIDATOR = jsonschema.Draft202012Validator(AHP, registry=REGISTRY, format_checker=jsonschema.FormatChecker())


def test_ahp_schema():
    jsonschema.Draft202012Validator.check_schema(AHP)


@pytest.mark.parametrize('kind', ['valid', 'invalid'])
def test_ahp_fixtures(kind):
    fixtures = sorted((ROOT / 'tests' / 'spec' / 'fixtures' / kind).glob('ahp.*.json'))
    assert fixtures
    for path in fixtures:
        errors = list(VALIDATOR.iter_errors(json.loads(path.read_text())))
        assert bool(errors) == (kind == 'invalid'), (path, errors)


def test_ahp_has_canonical_id():
    assert AHP["$id"] == "https://github.com/TheAgentHealth/agenthealth/spec/schemas/ahp.schema.json"
    assert not list(VALIDATOR.iter_errors(json.loads((ROOT / "tests/spec/fixtures/valid/ahp.graph.json").read_text())))
