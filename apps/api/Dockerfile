# syntax=docker/dockerfile:1.7

FROM golang:1.26.3-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Сборка.
# CGO_ENABLED=0 — статическая линковка, нет зависимости от libc.
# GOOS=linux — явно указываем целевую ОС.
# -ldflags="-w -s" — убираем debug символы: -w (DWARF), -s (symbol table).
#   Уменьшает размер бинарника на ~30%.
# -trimpath — убирает пути к исходникам из бинарника (безопасность).
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -trimpath \
    -o /app/bin/api \
    ./cmd/api/

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /app/bin/api /api

COPY --from=builder /app/migrations /migrations

EXPOSE 8080

USER 65534:65534

ENTRYPOINT ["/api"]
