# ==========================================
# Stage 1: Build
# ==========================================
FROM golang:1.22-alpine AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /application

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/

RUN --mount=type=cache,target=/root/.cache/go-build \
    go build \
    -trimpath \
    -ldflags="-w -s" \
    -o /weather-gateway \
    cmd/gateway/main.go

# ==========================================
# Stage 2: Runtime
# ==========================================
FROM alpine:3.19

LABEL maintainer="Ooko Tonny" \
      application="weather-intelligence-gateway" \
      version="1.0.0"

RUN apk --no-cache add ca-certificates wget

RUN addgroup -S gateway && \
    adduser -S gateway -G gateway

WORKDIR /app

COPY --from=builder \
     --chown=gateway:gateway \
     /weather-gateway .

USER gateway

EXPOSE 8080

ENV SERVER_PORT=8080

HEALTHCHECK --interval=30s \
            --timeout=5s \
            --start-period=10s \
            --retries=3 \
            CMD wget --spider -q \
            http://localhost:8080/health || exit 1

CMD ["./weather-gateway"]

