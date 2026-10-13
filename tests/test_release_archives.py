"""Archive metadata must not depend on the builder's filesystem or identity."""
import importlib.util
from pathlib import Path
import tarfile
import zipfile

spec = importlib.util.spec_from_file_location("build_release", Path(__file__).parents[1] / "scripts/build_release.py")
release = importlib.util.module_from_spec(spec)
spec.loader.exec_module(release)


def test_archives_are_deterministic_and_normalized(tmp_path):
    binary = tmp_path / "binary"
    binary.write_bytes(b"executable")
    license = tmp_path / "license"
    license.write_text("license")
    files = [(binary, "agenthealth"), (license, "LICENSE")]
    for windows in (False, True):
        first, second = tmp_path / "first", tmp_path / "second"
        release.write_archive(first, files, 1700000000, windows)
        binary.chmod(0o600)
        release.write_archive(second, files, 1700000000, windows)
        assert first.read_bytes() == second.read_bytes()
        if windows:
            with zipfile.ZipFile(first) as archive:
                assert archive.getinfo("agenthealth").external_attr >> 16 == 0o100755
                assert archive.getinfo("LICENSE").external_attr >> 16 == 0o100644
        else:
            with tarfile.open(first) as archive:
                for entry in archive.getmembers():
                    assert entry.uid == entry.gid == 0
                    assert entry.uname == entry.gname == ""
                    assert entry.mtime == 1700000000
                assert archive.getmember("agenthealth").mode == 0o755
                assert archive.getmember("LICENSE").mode == 0o644


def test_sbom_describes_linked_binary(tmp_path, monkeypatch):
    import hashlib
    import json
    binary = tmp_path / "agenthealth"
    binary.write_bytes(b"executable")
    monkeypatch.setattr(release.subprocess, "check_output", lambda *args, **kwargs:
                        "binary: go1.27.1\n\tpath\tgithub.com/TheAgentHealth/agenthealth/cmd/agenthealth\n\tmod\tgithub.com/TheAgentHealth/agenthealth\t(devel)\n\tdep\tgopkg.in/yaml.v3\tv3.0.1\th1:example\n")
    output = tmp_path / "sbom.json"
    release.write_sbom(output, binary, "v0.4.0", "linux", "amd64")
    bom = json.loads(output.read_text())
    assert bom["bomFormat"] == "CycloneDX" and bom["specVersion"] == "1.6"
    components = {c["bom-ref"]: c for c in bom["components"]}
    assert components["gopkg.in/yaml.v3"]["version"] == "v3.0.1"
    assert components["go-stdlib"]["version"] == "go1.27.1"
    assert components["executable"]["hashes"][0]["content"] == hashlib.sha256(binary.read_bytes()).hexdigest()
    assert set(bom["dependencies"][0]["dependsOn"]) == set(components)


def test_embedded_installation_guide_tracks_archive_version(tmp_path, monkeypatch):
    docs = tmp_path / 'docs'
    docs.mkdir()
    (docs / 'installation.md').write_text('<!-- release-version: v0.12.0 -->\n'
        'Download /v0.12.0/agenthealth_v0.12.0_linux_amd64.tar.gz\n'
        'Chart 0.12.0 and CLI v0.12.0; historical baseline v0.11.0.\n')
    monkeypatch.setattr(release, 'ROOT', tmp_path)
    for version in ('v0.13.1', 'v0.0.0-ci'):
        guide = release.installation_guide(version)
        assert version in guide and version[1:] in guide
        assert '0.12.0' not in guide
        assert 'historical baseline v0.11.0' in guide
        path = tmp_path / 'INSTALL.md'
        path.write_text(guide)
        for windows in (False, True):
            archive_path = tmp_path / ('guide.zip' if windows else 'guide.tar.gz')
            release.write_archive(archive_path, [(path, 'INSTALL.md')], 1700000000, windows)
            if windows:
                with zipfile.ZipFile(archive_path) as archive:
                    archived = archive.read('INSTALL.md')
            else:
                with tarfile.open(archive_path) as archive:
                    archived = archive.extractfile('INSTALL.md').read()
            assert archived == guide.encode()


def test_installation_guide_requires_explicit_template_version(tmp_path, monkeypatch):
    import pytest
    docs = tmp_path / 'docs'
    docs.mkdir()
    monkeypatch.setattr(release, 'ROOT', tmp_path)
    for text in ('unversioned', '<!-- release-version: v0.12.0 -->\n<!-- release-version: v0.13.0 -->'):
        (docs / 'installation.md').write_text(text)
        with pytest.raises(ValueError, match='exactly one'):
            release.installation_guide('v0.13.1')
    with pytest.raises(ValueError, match='invalid installation version'):
        release.installation_guide('../../secret')
