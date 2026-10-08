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

## Container image

Starting with the first release after Phase 12, each release tag publishes a
multi-platform image (`linux/amd64`, `linux/arm64`) to
`ghcr.io/theagenthealth/agenthealth`. Plain version tags publish `X.Y.Z`, `X.Y`
and `latest`; suffixed preview tags publish only their exact version. Image
tags omit the leading `v`. Until that release exists, build the image locally:

```bash
docker build --build-arg VERSION=dev -t agenthealth .
docker run --rm agenthealth version
```

The image uses a distroless static base with CA certificates, runs as the
non-root user `65532`, and contains no shell. Its entrypoint is `agenthealth`,
so arguments are CLI arguments. The binaries inside are byte-identical to the
binaries in the matching Linux release archives.

```bash
docker run --rm ghcr.io/theagenthealth/agenthealth ping http https://example.com

# Mount configuration read-only and pass referenced credentials by name.
docker run --rm \
  -v "$PWD/agenthealth.yaml:/config/agenthealth.yaml:ro" \
  -e AGENT_API_TOKEN \
  ghcr.io/theagenthealth/agenthealth check /config/agenthealth.yaml --format json
```

Exit codes are the [CLI exit codes](../spec/exit-codes.md), so the container
can gate CI jobs and Kubernetes probes directly. Use `host.docker.internal`
(Docker Desktop) or a container network to reach services on the host.

Limitations: MCP stdio targets run their configured executable inside the
container, so build a derived image that adds that server. `agenthealth login`
needs a browser callback and is intended for a workstation binary. Token files
require owner-only permissions and may be rewritten on refresh, so a container
reusing one needs a writable volume owned by UID 65532.

Verify a published image's signed attestation with the GitHub CLI:

```bash
gh attestation verify oci://ghcr.io/theagenthealth/agenthealth:<version> --repo TheAgentHealth/agenthealth
```
