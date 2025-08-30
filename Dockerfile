FROM alpine:latest AS base
FROM base AS builder
ARG TARGETPLATFORM
COPY release-assests .
RUN ARTIFACT_ARCH=""; \
    if [ "$TARGETPLATFORM" = "linux/amd64" ]; then \
        ARTIFACT_ARCH="amd64"; \
    elif [ "$TARGETPLATFORM" = "linux/386" ]; then \
        ARTIFACT_ARCH="386"; \
    elif [ "$TARGETPLATFORM" = "linux/arm/v7" ]; then \
        ARTIFACT_ARCH="armv7"; \
    elif [ "$TARGETPLATFORM" = "linux/arm64" ]; then \
        ARTIFACT_ARCH="armv8"; \
    elif [ "$TARGETPLATFORM" = "linux/riscv64" ]; then \
        ARTIFACT_ARCH="riscv64"; \
    else \
        echo "Unsupported architecture: $TARGETPLATFORM"; \
        exit 1; \
    fi; \
    mv release-assests/$ARTIFACT_ARCH /app; \
    cd /app && mv config.example.toml config.toml


FROM base
COPY --from=builder /app .
RUN apk add --no-cache tzdata
WORKDIR /app
ENV TZ=Asia/Shanghai
ENTRYPOINT ["/app/cfstd"]
