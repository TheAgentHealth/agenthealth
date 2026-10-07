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
