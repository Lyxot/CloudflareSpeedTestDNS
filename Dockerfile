# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
ARG VERSION
ARG GIT_COMMIT
RUN go build -o /cfstd -ldflags="-s -w -X main.version=${VERSION} -X main.gitCommit=${GIT_COMMIT}"

# Final stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /cfstd /app/
COPY conf/config.example.toml /app/config.toml
COPY ip.txt /app/ip.txt
COPY ipv6.txt /app/ipv6.txt
ENTRYPOINT ["/app/cfstd"]
