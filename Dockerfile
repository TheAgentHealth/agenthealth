# syntax=docker/dockerfile:1
#
# Official AgentHealth container image (Phase 12).
#
# The builder cross-compiles on the native build platform, so multi-platform
# builds do not need CPU emulation. Build flags match scripts/build_release.py,
# so image binaries are byte-identical to the attested release archives.
# Base images are pinned by digest; update tag and digest together.
#
#   docker build --build-arg VERSION=v0.4.0 -t agenthealth .
#   docker run --rm agenthealth ping http https://example.com

ARG GO_IMAGE=golang:1.27.2-bookworm@sha256:5cf287a799e6b94384bad13d16b14904c531f51ba65792237e122ce42b392f61
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
ARG REVISION=unknown
LABEL org.opencontainers.image.title="agenthealth" \
      org.opencontainers.image.description="Universal health and readiness checks for agentic systems" \
      org.opencontainers.image.source="https://github.com/TheAgentHealth/agenthealth" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${REVISION}"
COPY --from=build /out/agenthealth /usr/local/bin/agenthealth
COPY LICENSE /usr/share/doc/agenthealth/LICENSE
USER 65532:65532
WORKDIR /config
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/agenthealth"]
CMD ["--help"]
