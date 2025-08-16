# Dockerfile
FROM golang:1.24

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o /go-lcov-summary ./cmd/go-lcov-summary

ENTRYPOINT ["/go-lcov-summary"]
