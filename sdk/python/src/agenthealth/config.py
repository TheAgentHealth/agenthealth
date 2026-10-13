"""Load trusted configuration without executing checks or resolving credentials."""
from copy import deepcopy
from importlib.resources import files
import json
from pathlib import Path
from typing import Any, Mapping

from jsonschema import Draft202012Validator
import yaml

from .errors import ConfigurationError

MAX_CONFIG_BYTES = 4 * 1024 * 1024


def validator(name: str) -> Draft202012Validator:
    schema = json.loads(files("agenthealth").joinpath("schemas", name + ".schema.json").read_text())
    return Draft202012Validator(schema)


class Configuration:
    """A validated copy of an AHS v1 document; semantic validation stays in Go."""

    def __init__(self, data: Mapping[str, Any], *, base_dir: str | Path | None = None):
        try:
            document = json.loads(json.dumps(dict(data), allow_nan=False))
            validator("configuration").validate(document)
            if len(json.dumps(document).encode()) > MAX_CONFIG_BYTES:
                raise ValueError()
        except Exception:
            # Schema/parser exceptions may contain credentials and configured values.
            raise ConfigurationError("Invalid AgentHealth configuration") from None
        self._data = document
        self.base_dir = Path(base_dir).resolve() if base_dir is not None else Path.cwd()

    def to_dict(self) -> dict[str, Any]:
        return deepcopy(self._data)


class _UniqueLoader(yaml.SafeLoader):
    pass


def _mapping(loader: _UniqueLoader, node: yaml.MappingNode, deep: bool = False):
    result = {}
    for key_node, value_node in node.value:
        key = loader.construct_object(key_node, deep=deep)
        if key in result:
            raise ValueError("Duplicate YAML key")
        result[key] = loader.construct_object(value_node, deep=deep)
    return result


_UniqueLoader.add_constructor(yaml.resolver.BaseResolver.DEFAULT_MAPPING_TAG, _mapping)


def load_config(path: str | Path) -> Configuration:
    """Read bounded JSON/YAML. Relative credential/token paths use the caller's cwd,
    matching CLI behavior. Pass base_dir explicitly to Configuration to change it.
    """
    try:
        with Path(path).open("rb") as stream:
            data = stream.read(MAX_CONFIG_BYTES + 1)
        if len(data) > MAX_CONFIG_BYTES:
            raise ValueError()
        document = yaml.load(data, Loader=_UniqueLoader)
        if not isinstance(document, dict):
            raise ValueError()
        return Configuration(document)
    except Exception:
        raise ConfigurationError("Unable to load AgentHealth configuration") from None
