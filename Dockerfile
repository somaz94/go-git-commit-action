# Build natively and cross-compile via TARGETARCH instead of running the Go
# toolchain under QEMU for the arm64 image.
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS builder

WORKDIR /app

# BuildKit injects TARGETARCH; the legacy builder cannot parse the FROM above.
ARG TARGETOS=linux
ARG TARGETARCH

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /go-git-commit-action ./cmd/main.go

FROM alpine:3.24

RUN apk add --no-cache \
    git \
    github-cli \
    curl

WORKDIR /app

COPY --from=builder /go-git-commit-action /go-git-commit-action

ENTRYPOINT ["/go-git-commit-action"]