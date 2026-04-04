# =============================================================
# Stage 1 — builder
# =============================================================
FROM golang:1.26.1-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Зависимости (кешируются отдельно от кода)
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Бинарник simulation сервиса
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bin/simulation ./cmd/simulation

# =============================================================
# Stage 2 — runtime
# =============================================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /bin/simulation ./simulation
COPY --from=builder /app/metrics.yaml ./metrics.yaml
COPY --from=builder /app/db/migrations ./db/migrations

CMD ["./simulation"]
