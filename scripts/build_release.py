#!/usr/bin/env python3
"""Cross-compile CLI archives and SHA-256 checksums for a version tag."""
import argparse
import hashlib
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
    with tempfile.TemporaryDirectory(prefix="agenthealth-release-") as temporary:
        for system, arch in PLATFORMS:
            binary_name = "agenthealth.exe" if system == "windows" else "agenthealth"
            binary = Path(temporary) / binary_name
            subprocess.run(
                ["go", "build", "-trimpath", "-ldflags",
                 f"-s -w -X main.version={args.version}", "-o", str(binary), "./cmd/agenthealth"],
                cwd=ROOT, check=True,
                env={**os.environ, "CGO_ENABLED": "0", "GOOS": system, "GOARCH": arch},
            )
            stem = f"agenthealth_{args.version}_{system}_{arch}"
            files = [(binary, binary_name), (ROOT / "LICENSE", "LICENSE"),
                     (ROOT / "docs/installation.md", "INSTALL.md")]
            if system == "windows":
                archive = args.output / f"{stem}.zip"
                with zipfile.ZipFile(archive, "w", zipfile.ZIP_DEFLATED) as bundle:
                    for source, name in files:
                        bundle.write(source, name)
            else:
                archive = args.output / f"{stem}.tar.gz"
                with tarfile.open(archive, "w:gz") as bundle:
                    for source, name in files:
                        bundle.add(source, arcname=name)
            archives.append(archive)
            print(f"Built {archive.name}", flush=True)
    checksums = "".join(f"{hashlib.sha256(path.read_bytes()).hexdigest()}  {path.name}\n"
                        for path in sorted(archives))
    (args.output / "checksums.txt").write_text(checksums)


if __name__ == "__main__":
    main()
