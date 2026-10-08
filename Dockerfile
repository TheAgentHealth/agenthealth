# syntax=docker/dockerfile:1
#
# Official AgentHealth container image (Phase 12).
#
# The builder cross-compiles on the native build platform, so multi-platform
# builds do not need CPU emulation. Build flags match scripts/build_release.py.
# Base images are pinned by digest; update tag and digest together.
#
#   docker build --build-arg VERSION=v0.4.0 -t agenthealth .
#   docker run --rm agenthealth ping http https://example.com

ARG GO_IMAGE=golang:1.27.1-bookworm@sha256:8d48e12ec56735e9358640898b9d9b9fcca110612ed8a5567438c0a1baa24e66
ARG RUNTIME_IMAGE=gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab

FROM --platform=$BUILDPLATFORM ${GO_IMAGE} AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY cmd ./cmd
COPY core ./core
COPY adapters ./adapters
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -buildvcs=false -ldflags "-s -w -X main.version=${VERSION}" \
    -o /out/agenthealth ./cmd/agenthealth

# Distroless static: CA certificates and tzdata, no shell or package manager,
# and a fixed non-root user (65532). MCP stdio targets need their server
# executable in the image; build a derived image for that use case.
FROM ${RUNTIME_IMAGE}
ARG VERSION=dev
LABEL org.opencontainers.image.title="agenthealth" \
      org.opencontainers.image.description="Universal health and readiness checks for agentic systems" \
      org.opencontainers.image.source="https://github.com/TheAgentHealth/agenthealth" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.version="${VERSION}"
COPY --from=build /out/agenthealth /usr/local/bin/agenthealth
COPY LICENSE /usr/share/doc/agenthealth/LICENSE
USER 65532:65532
WORKDIR /home/nonroot
ENTRYPOINT ["/usr/local/bin/agenthealth"]
CMD ["--help"]
