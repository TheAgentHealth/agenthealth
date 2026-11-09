"""Conformance evidence must not silently pass missing or skipped tests."""
import importlib.util
import json
import jsonschema
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
spec = importlib.util.spec_from_file_location('conformance', ROOT / 'scripts/test_conformance.py')
suite = importlib.util.module_from_spec(spec)
spec.loader.exec_module(suite)


def events():
    return [dict(Action='pass', Package=suite.PREFIX + package, Test=name)
            for package, names in suite.CASES.items() for name in names]


def test_suite_rejects_missing_failed_skipped_and_malformed_evidence():
    complete = events()
    encode = lambda rows: '\n'.join(json.dumps(row) for row in rows)
    assert suite.evaluate(encode(complete), 0)['passed']
    assert not suite.evaluate(encode(complete[:-1]), 0)['passed']
    assert not suite.evaluate(encode(complete), 1)['passed']
    assert not suite.evaluate(encode(complete) + '\ninvalid', 0)['passed']
    for action in ('skip', 'fail'):
        changed = complete + [dict(complete[0], Action=action)]
        assert not suite.evaluate(encode(changed), 0)['passed']
    skipped_child = dict(complete[0], Action='skip', Test=complete[0]['Test'] + '/child')
    assert not suite.evaluate(encode(complete + [skipped_child]), 0)['passed']


def test_mandatory_cases_exist_in_source():
    for package, names in suite.CASES.items():
        source = '\n'.join(path.read_text(encoding='utf-8') for path in (ROOT / package).glob('*_test.go'))
        for name in names:
            assert f'func {name}(' in source, (package, name)


def test_examples_have_valid_schema_but_distinct_semantics():
    schema = json.loads((ROOT / 'spec/schemas/result.schema.json').read_text())
    for path in (ROOT / 'examples/conformance').glob('*.json'):
        document = json.loads(path.read_text())
        jsonschema.validate(document, schema)
    inconsistent = json.loads((ROOT / 'examples/conformance/inconsistent.json').read_text())
    assert inconsistent['status'] != inconsistent['checks']['reachability']['status']
