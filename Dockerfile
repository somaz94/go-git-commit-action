# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY . .

RUN go build -o /go-git-commit-action ./cmd/main.go

# Final stage
FROM alpine:latest

# Install required git packages
RUN apk add --no-cache \
    git \
    github-cli \
    curl

# Set working directory
WORKDIR /app

COPY --from=builder /go-git-commit-action /go-git-commit-action

ENTRYPOINT ["/go-git-commit-action"]