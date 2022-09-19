FROM golang:1.19-bullseye

WORKDIR /go/src/github.com/zalgonoise/zlog
COPY go.mod ./
COPY go.sum ./
RUN go mod download \
    && go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.49.0

