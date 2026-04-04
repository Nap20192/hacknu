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

# Бинарник alerts сервиса
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bin/alerts ./cmd/alerts

# =============================================================
# Stage 2 — runtime
# =============================================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /bin/alerts ./alerts

EXPOSE 8081

CMD ["./alerts"]
