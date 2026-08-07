# Build stage
# Pinned to the BUILD platform, not the target: the image is published
# multi-arch, and letting the toolchain stage run under QEMU emulation to
# produce an arm64 binary is minutes of emulated compilation for no reason. Go
# cross-compiles natively instead, driven by the TARGET* args buildx injects.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

WORKDIR /app

# Supplied per target platform. The defaults keep a plain `docker build` (a
# local one-off, no --platform) working: TARGETARCH resolves empty there, and
# an empty GOARCH means "host default", which is what that build wants anyway.
#
# BuildKit is required either way — `$BUILDPLATFORM` on the FROM above is a
# BuildKit-only variable, and the legacy builder fails to parse the line rather
# than ignoring it. That is the default builder since Docker 23, so this only
# bites an explicit DOCKER_BUILDKIT=0.
ARG TARGETOS=linux
ARG TARGETARCH

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /go-git-commit-action ./cmd/main.go

# Final stage
FROM alpine:3.24

# Install required git packages
RUN apk add --no-cache \
    git \
    github-cli \
    curl

# Set working directory
WORKDIR /app

COPY --from=builder /go-git-commit-action /go-git-commit-action

ENTRYPOINT ["/go-git-commit-action"]