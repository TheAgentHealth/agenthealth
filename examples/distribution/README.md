# Standalone distribution preparation

Phase 13 tooling is included starting in v0.11.0. This
example prepares local assets; it does not publish a release or package channel.
Run from the repository root with Go 1.26.4 or newer for nFPM (the CLI itself
supports Go 1.23+) and Python 3. Use a fresh, empty output directory.

```bash
python3 scripts/build_release.py v0.0.0-local --output /tmp/agenthealth-local-assets
python3 scripts/package_release.py v0.0.0-local --output /tmp/agenthealth-local-assets
python3 scripts/test_release_binary.py v0.0.0-local --output /tmp/agenthealth-local-assets
```

The smoke test selects the host OS/architecture and verifies the archive hash,
version, help and missing-configuration tool-failure exit code. It supports
Linux AMD64/ARM64, macOS AMD64/ARM64 and Windows AMD64; it cannot execute other
platforms' binaries. On Windows choose a writable Windows output path instead.

On Linux, verify every generated asset and inspect a DEB without installing it:

```bash
cd /tmp/agenthealth-local-assets
sha256sum --check checksums.txt
mkdir deb-inspect
dpkg-deb --extract agenthealth_v0.0.0-local_linux_amd64.deb deb-inspect
./deb-inspect/usr/bin/agenthealth version
```

Use the ARM64 DEB on an ARM64 host. RPM and DEB packages reuse archive binaries
and include a CA certificate dependency. The local-version Homebrew/Scoop
manifests contain URLs for a tag that does not exist and must not be installed;
they demonstrate generated metadata only. For published versions, follow
[package installation](../../docs/installation.md#package-channels-phase-13-v0110)
and [release policy](../../RELEASING.md#phase-13-package-assets).

Existing [health configuration examples](../README.md) work with the same CLI
regardless of installation method. Services, credentials, trusted MCP stdio
executables and safe runtime probe handlers must still be supplied by users.
