# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.27.1-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY core ./core
COPY adapters ./adapters
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -buildvcs=false -ldflags "-s -w -X main.version=$VERSION" -o /out/agenthealth ./cmd/agenthealth

FROM scratch
ARG VERSION=dev
ARG REVISION=unknown
LABEL org.opencontainers.image.title="AgentHealth" \
      org.opencontainers.image.source="https://github.com/TheAgentHealth/agenthealth" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.version=$VERSION \
      org.opencontainers.image.revision=$REVISION
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/agenthealth /usr/local/bin/agenthealth
COPY LICENSE /LICENSE
USER 65532:65532
WORKDIR /config
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/agenthealth"]
CMD ["--help"]
