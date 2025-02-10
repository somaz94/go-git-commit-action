FROM golang:1.23-alpine

RUN apk add --no-cache git github-cli curl

WORKDIR /app

COPY . .

RUN go build -o /go-git-commit-action ./cmd/main.go

ENTRYPOINT ["/go-git-commit-action"]