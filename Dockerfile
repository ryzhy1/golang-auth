FROM golang:1.22-alpine AS builder

WORKDIR /app

# Копируем go.mod, go.sum и .env файл
COPY go.mod go.sum ./
COPY .env /app/.env

# Устанавливаем зависимости
RUN go mod download
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Копируем весь исходный код
COPY . .

RUN go build -o /app/bin/app ./cmd/sso/main.go

# Финальный образ на базе alpine
FROM alpine:3.18 AS runner

WORKDIR /app

# Копируем собранное приложение и необходимые файлы из builder
COPY --from=builder /app/bin/app /app/app
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /app/.env /app/.env

# Копируем конфигурационный файл и миграции
COPY --from=builder /app/config/dev.yaml /app/config/dev.yaml
COPY --from=builder /app/migrations /app/migrations

# Устанавливаем переменные окружения
ENV CONFIG_PATH /app/config/dev.yaml

# Запускаем миграции и основное приложение
CMD ["sh", "-c", "goose -dir /app/migrations postgres 'user=postgres password=postgres dbname=postgres sslmode=disable' up && ./app"]
