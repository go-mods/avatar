# syntax=docker/dockerfile:1

# Same minor as the go line of go.mod
ARG GO_VERSION=1.26.10

# Build the server: a static binary, without downloading another toolchain
FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src
ENV CGO_ENABLED=0 GOTOOLCHAIN=local

# Dependencies first: this layer is rebuilt only when go.mod or go.sum change
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/avatar ./cmd/avatar

# A distroless image has no shell to create this directory
RUN mkdir -p /out/fonts

# Run it in an image which only holds the certificates, the time zones and an unprivileged user
FROM gcr.io/distroless/static-debian13:nonroot
ARG VERSION=dev
LABEL org.opencontainers.image.version="${VERSION}" \
    org.opencontainers.image.source="https://github.com/go-mods/avatar"

COPY --from=build /out/avatar /usr/local/bin/avatar
# A new named volume takes the owner of this directory: the server can write its fonts
COPY --from=build --chown=nonroot:nonroot /out/fonts /var/cache/avatar/fonts

USER nonroot:nonroot
EXPOSE 8080

# The binary probes itself: the image has no curl nor shell
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD ["/usr/local/bin/avatar", "healthcheck"]

ENTRYPOINT ["/usr/local/bin/avatar", "-fontdir", "/var/cache/avatar/fonts"]
