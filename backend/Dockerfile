# =============================================================
# Stage 1 — builder
# =============================================================
FROM golang:1.26.1-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata curl

WORKDIR /app

# Зависимости (кешируются отдельно от кода)
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Бинарники приложения
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bin/server    ./cmd/server     && \
    go build -ldflags="-s -w" -o /bin/simulator ./cmd/simulation

# golang-migrate CLI для запуска миграций
RUN MIGRATE_VER=v4.18.1 && \
    curl -sSL "https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VER}/migrate.linux-amd64.tar.gz" \
    | tar -xz -C /bin && \
    chmod +x /bin/migrate

# =============================================================
# Stage 2 — runtime
# =============================================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /bin/server    ./server
COPY --from=builder /bin/simulator ./simulator
COPY --from=builder /bin/migrate   ./migrate
COPY --from=builder /app/metrics.yaml ./metrics.yaml
COPY --from=builder /app/db/migrations ./db/migrations

# entrypoint.sh: сначала миграции, потом сервер
COPY entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["./entrypoint.sh"]
