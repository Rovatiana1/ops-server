# ── Stage 1: Builder ─────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /bin/ops-server ./cmd/api/main.go

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM scratch

# Copy CA certs for HTTPS calls
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
# Copy binary
COPY --from=builder /bin/ops-server /ops-server
# Copy config (can be overridden by volume mount)
COPY --from=builder /app/configs/config.yaml /configs/config.yaml

EXPOSE 8080

ENTRYPOINT ["/ops-server"]
