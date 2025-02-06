FROM golang:1.23-alpine

RUN apk add --no-cache git

WORKDIR /app
COPY . .

RUN go build -o /go-git-commit-action ./cmd/main.go

ENTRYPOINT ["/go-git-commit-action"]