# Install AgentHealth

Download the archive for your operating system and CPU from [GitHub Releases](https://github.com/TheAgentHealth/agenthealth/releases). Go is not required for downloaded binaries.

| Platform | Archive suffix |
|---|---|
| Linux x86-64 | `linux_amd64.tar.gz` |
| Linux ARM64 | `linux_arm64.tar.gz` |
| macOS Intel | `darwin_amd64.tar.gz` |
| macOS Apple Silicon | `darwin_arm64.tar.gz` |
| Windows x86-64 | `windows_amd64.zip` |

## Linux x86-64 example

```bash
curl -fLO https://github.com/TheAgentHealth/agenthealth/releases/download/v0.4.0/agenthealth_v0.4.0_linux_amd64.tar.gz
curl -fLO https://github.com/TheAgentHealth/agenthealth/releases/download/v0.4.0/checksums.txt
sha256sum --check --ignore-missing checksums.txt
tar -xzf agenthealth_v0.4.0_linux_amd64.tar.gz
./agenthealth version
./agenthealth ping http https://example.com
```

To run from any directory, place the binary in a directory on your PATH. For example:

```bash
mkdir -p "$HOME/.local/bin"
install -m 755 agenthealth "$HOME/.local/bin/agenthealth"
export PATH="$HOME/.local/bin:$PATH"
agenthealth version
```

Persist the PATH addition in your shell configuration if needed.

On macOS, use the appropriate tar archive and `shasum -a 256` to compare its hash with `checksums.txt`. On Windows, extract the ZIP and run `./agenthealth.exe version` in PowerShell; use `Get-FileHash -Algorithm SHA256` to compare the archive hash. Release artifacts carry GitHub keyless signed provenance and per-platform SBOM attestations. OS-native executable signing and macOS notarization remain future distribution work.

## Build from source

With Go 1.23 or newer:

```bash
git clone https://github.com/TheAgentHealth/agenthealth.git
cd agenthealth
git checkout v0.4.0
go build -ldflags '-X main.version=v0.4.0' -o agenthealth ./cmd/agenthealth
./agenthealth version
```

See [CLI usage](https://github.com/TheAgentHealth/agenthealth/blob/main/docs/cli.md) and [HTTP configuration](https://github.com/TheAgentHealth/agenthealth/blob/main/docs/http-adapter.md). The v0.4.0 release supports HTTP/API, MCP HTTP/stdio with OAuth, and A2A 1.0 JSON-RPC targets with explicit 0.3.0 compatibility. See [MCP configuration](mcp-adapter.md) and [A2A configuration](a2a-adapter.md).

## Verify provenance and SBOMs

Each archive has a matching `agenthealth_v0.4.0_<os>_<arch>.cdx.json` CycloneDX
1.6 SBOM describing linked Go modules, the standard-library version and the
executable SHA-256. `checksums.txt` covers both archives and SBOMs.

With a current GitHub CLI supporting `gh attestation`:

```bash
gh attestation verify agenthealth_v0.4.0_linux_amd64.tar.gz --repo TheAgentHealth/agenthealth
gh attestation verify agenthealth_v0.4.0_linux_amd64.tar.gz --repo TheAgentHealth/agenthealth --predicate-type https://cyclonedx.org/bom
```

Attestations bind artifacts to this repository’s release workflow through a
GitHub OIDC identity. Checksums alone do not authenticate the publisher.
Archives normalize owner IDs, names, modes, timestamps and gzip/ZIP metadata.
Reproduction requires the same source commit, Go toolchain and build environment;
archive timestamp defaults to the commit time and can use `SOURCE_DATE_EPOCH`.
