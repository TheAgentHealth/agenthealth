"""The container image must stay pinned, minimal, and build like the release archives."""
from pathlib import Path
import re

ROOT = Path(__file__).parents[1]
DOCKERFILE = (ROOT / "Dockerfile").read_text()
BUILD_RELEASE = (ROOT / "scripts/build_release.py").read_text()


def test_base_images_are_pinned_by_digest():
    images = re.findall(r"^ARG \w+_IMAGE=(\S+)$", DOCKERFILE, re.MULTILINE)
    assert len(images) == 2
    for image in images:
        assert re.fullmatch(r"[^@\s]+:[^@\s]+@sha256:[0-9a-f]{64}", image), image


def test_runtime_is_non_root_distroless():
    runtime = re.search(r"^ARG RUNTIME_IMAGE=(\S+)$", DOCKERFILE, re.MULTILINE).group(1)
    assert runtime.startswith("gcr.io/distroless/static")
    assert re.search(r"^USER 65532:65532$", DOCKERFILE, re.MULTILINE)
    assert 'ENTRYPOINT ["/usr/local/bin/agenthealth"]' in DOCKERFILE


def test_build_flags_match_release_archives():
    # Identical flags keep image binaries byte-identical to the attested archives.
    for flag in ("-trimpath", "-buildvcs=false", "-s -w -X main.version=", "CGO_ENABLED"):
        assert flag in DOCKERFILE and flag in BUILD_RELEASE, flag
    go_version = re.search(r'go-version: "([\d.]+)"', (ROOT / ".github/workflows/release.yml").read_text()).group(1)
    assert f"golang:{go_version}-" in DOCKERFILE


def test_build_context_excludes_local_state():
    ignored = [line for line in (ROOT / ".dockerignore").read_text().splitlines() if line and not line.startswith("#")]
    assert ignored[0] == "*"
    assert "!.credentials/" not in ignored and "!dist/" not in ignored
