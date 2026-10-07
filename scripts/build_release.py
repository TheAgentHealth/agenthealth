#!/usr/bin/env python3
"""Cross-compile CLI archives and SHA-256 checksums for a version tag."""
import argparse
import hashlib
import gzip
import io
import json
from datetime import datetime, timezone
from urllib.parse import quote
import os
from pathlib import Path
import re
import subprocess
import tarfile
import tempfile
import zipfile

ROOT = Path(__file__).resolve().parents[1]
PLATFORMS = [("linux", "amd64"), ("linux", "arm64"),
             ("darwin", "amd64"), ("darwin", "arm64"), ("windows", "amd64")]


def write_archive(archive, files, epoch, windows=False):
    """Normalize ownership, modes, timestamps, gzip headers and entry order."""
    if windows:
        stamp = datetime.fromtimestamp(max(epoch, 315532800), timezone.utc)
        with zipfile.ZipFile(archive, "w", zipfile.ZIP_DEFLATED) as bundle:
            for source, name in files:
                entry = zipfile.ZipInfo(name, stamp.timetuple()[:6])
                entry.create_system = 3
                entry.external_attr = (0o100755 if name.startswith("agenthealth") else 0o100644) << 16
                entry.compress_type = zipfile.ZIP_DEFLATED
                bundle.writestr(entry, source.read_bytes())
    else:
        with archive.open("wb") as output:
            with gzip.GzipFile(filename="", mode="wb", fileobj=output, mtime=epoch) as compressed:
                with tarfile.open(fileobj=compressed, mode="w", format=tarfile.USTAR_FORMAT) as bundle:
                    for source, name in files:
                        data = source.read_bytes()
                        entry = tarfile.TarInfo(name)
                        entry.size = len(data)
                        entry.uid = entry.gid = 0
                        entry.uname = entry.gname = ""
                        entry.mtime = epoch
                        entry.mode = 0o755 if name == "agenthealth" else 0o644
                        bundle.addfile(entry, io.BytesIO(data))


def write_sbom(output, binary, version, system, arch):
    """CycloneDX module inventory from the actual linked binary's build info.

    This describes shipped Go modules and the standard library, not test-only
    dependencies or the CI runner. It contains the executable hash for binding.
    """
    info = subprocess.check_output(["go", "version", "-m", str(binary)], text=True)
    compiler = info.splitlines()[0].rsplit(": ", 1)[-1].strip()
    root = {"type": "application", "bom-ref": "agenthealth", "name": "agenthealth", "version": version}
    components = [{"type": "library", "bom-ref": "go-stdlib", "name": "Go standard library", "version": compiler},
                  {"type": "file", "bom-ref": "executable", "name": binary.name,
                   "hashes": [{"alg": "SHA-256", "content": hashlib.sha256(binary.read_bytes()).hexdigest()}]}]
    for line in info.splitlines()[1:]:
        fields = line.split()
        if fields and fields[0] == "dep":
            name, revision = fields[1:3]
            component = {"type": "library", "bom-ref": name, "name": name, "version": revision,
                         "purl": f"pkg:golang/{quote(name, safe='/')}@{quote(revision, safe='')}"}
            if len(fields) > 3:
                component["properties"] = [{"name": "go:module:sum", "value": fields[3]}]
            components.append(component)
        elif fields and fields[0] == "=>":
            raise ValueError("release SBOM does not permit replaced modules")
    document = {"bomFormat": "CycloneDX", "specVersion": "1.6", "version": 1,
                "metadata": {"component": root, "properties": [{"name": "go:os", "value": system}, {"name": "go:arch", "value": arch}]},
                "components": components,
                "dependencies": [{"ref": "agenthealth", "dependsOn": [c["bom-ref"] for c in components]}]}
    output.write_text(json.dumps(document, indent=2, sort_keys=True) + "\n")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("version", help="release tag, such as v0.1.0")
    parser.add_argument("--output", type=Path, default=ROOT / "dist")
    args = parser.parse_args()
    if not re.fullmatch(r"v\d+\.\d+\.\d+(?:-[A-Za-z0-9.-]+)?", args.version):
        parser.error("version must be a vMAJOR.MINOR.PATCH tag")
    args.output.mkdir(parents=True, exist_ok=True)
    if any(args.output.iterdir()):
        parser.error("output directory must be empty to prevent stale release assets")
    archives = []
    epoch = int(os.environ.get("SOURCE_DATE_EPOCH", subprocess.check_output(
        ["git", "log", "-1", "--format=%ct"], cwd=ROOT, text=True).strip()))
    with tempfile.TemporaryDirectory(prefix="agenthealth-release-") as temporary:
        for system, arch in PLATFORMS:
            binary_name = "agenthealth.exe" if system == "windows" else "agenthealth"
            binary = Path(temporary) / binary_name
            subprocess.run(
                ["go", "build", "-trimpath", "-ldflags",
                 f"-s -w -X main.version={args.version}", "-buildvcs=false", "-o", str(binary), "./cmd/agenthealth"],
                cwd=ROOT, check=True,
                env={**os.environ, "CGO_ENABLED": "0", "GOOS": system, "GOARCH": arch},
            )
            stem = f"agenthealth_{args.version}_{system}_{arch}"
            files = [(binary, binary_name), (ROOT / "LICENSE", "LICENSE"),
                     (ROOT / "docs/installation.md", "INSTALL.md")]
            archive = args.output / (f"{stem}.zip" if system == "windows" else f"{stem}.tar.gz")
            write_archive(archive, files, epoch, windows=system == "windows")
            sbom = args.output / f"{stem}.cdx.json"
            write_sbom(sbom, binary, args.version, system, arch)
            archives.extend([archive, sbom])
            print(f"Built {archive.name}", flush=True)
    checksums = "".join(f"{hashlib.sha256(path.read_bytes()).hexdigest()}  {path.name}\n"
                        for path in sorted(archives))
    (args.output / "checksums.txt").write_text(checksums)


if __name__ == "__main__":
    main()
