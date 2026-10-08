# Docker distribution

**Status:** Phase 12, source-only. Images are not yet published; building and
running locally is documented below. See the [Docker Distribution roadmap](../ROADMAP.md#phase-12--docker-distribution).

The [Dockerfile](../Dockerfile) builds a minimal, non-root AgentHealth image
from source. It packages the same CLI described in [docs/cli.md](cli.md); no
container-specific behavior is introduced.

## Image contents

- Multi-stage build: `golang:1.27.1-bookworm` compiles a static (`CGO_ENABLED=0`)
  binary; the runtime stage is `FROM scratch`.
- Runtime image contains only the `agenthealth` binary, CA certificates,
  and `LICENSE`. No shell, package manager, or other OS tooling is present.
- Runs as the fixed non-root user `65532:65532`.
- `WORKDIR /config`, so relative configuration paths resolve against a mounted
  configuration volume.
- `EXPOSE 8080` documents the conventional port for [AHP serving](ahp.md);
  the image does not run a server by default.
- Default `CMD ["--help"]`; override the command to run `ping`, `check`,
  `doctor`, or `serve`.
- `org.opencontainers.image.*` labels carry `VERSION`/`REVISION` build args,
  set from the release tag and commit SHA in CI.

## Build locally

Requires a Buildx-enabled Docker CLI (`docker buildx version`):

```bash
docker buildx build --load --build-arg VERSION=dev -t agenthealth:dev .
```

Cross-compilation uses `BUILDPLATFORM`/`TARGETARCH` build args set
automatically by Buildx; building `linux/amd64` and `linux/arm64` together
requires a `docker-container` builder:

```bash
docker buildx create --use
docker buildx build --platform linux/amd64,linux/arm64 -t agenthealth:dev .
```

## Run

```bash
docker run --rm agenthealth:dev version
docker run --rm agenthealth:dev ping http https://example.com
```

Mount a configuration file read-only at `/config` to use `check` or `doctor`
with a configuration file, matching the default working directory:

```bash
docker run --rm -v "$PWD/agenthealth.yaml:/config/agenthealth.yaml:ro" \
  agenthealth:dev check agenthealth.yaml --format json
```

Targets that reference `localhost`/`127.0.0.1` resolve relative to the
container's own network namespace; use `--network host` (Linux only) or a
container-reachable hostname to reach services on the host.

## Published tags

`.github/workflows/container.yml` validates both platforms on every pull
request and push, and publishes to `ghcr.io/theagenthealth/agenthealth` only
for tagged releases (`v*`) that already have a `docs/releases/<tag>.md` file:

- `ghcr.io/theagenthealth/agenthealth:vX.Y.Z` for every release tag.
- `ghcr.io/theagenthealth/agenthealth:vX` and `:latest` are also updated,
  except for pre-release tags (tags containing `-`), which only publish the
  full `vX.Y.Z` tag.

Published images carry build-provenance and SBOM
[attestations](https://docs.github.com/actions/security-guides/using-artifact-attestations-to-establish-provenance-for-builds)
and are signed with [cosign](https://github.com/sigstore/cosign) using GitHub
OIDC (keyless signing). Verify with:

```bash
gh attestation verify oci://ghcr.io/theagenthealth/agenthealth:vX.Y.Z --owner TheAgentHealth
cosign verify ghcr.io/theagenthealth/agenthealth:vX.Y.Z \
  --certificate-identity-regexp 'https://github.com/TheAgentHealth/agenthealth/.+' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```

See [installation.md](installation.md) for standalone binary archives and
their own provenance/SBOM verification, and [RELEASING.md](../RELEASING.md)
for the release process that both distribution channels share.
