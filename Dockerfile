# Multi-stage Dockerfile for Log Linter

# Build stage
FROM golang:1.22-alpine AS builder

# Установка зависимостей для сборки
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Копирование go mod файлов
COPY go.mod go.sum ./

# Загрузка зависимостей
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка бинарного файла
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o loglinter cmd/simple/main.go

# Final stage
FROM alpine:latest

# Установка ca-certificates для HTTPS запросов
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Копирование бинарного файла из builder stage
COPY --from=builder /app/loglinter .

# Копирование конфигурационного файла по умолчанию
COPY .loglinter.yaml .

# Создание директории для логов
RUN mkdir -p /root/logs

# Установка переменных окружения
ENV TZ=UTC

# Открываем порт для health check (если понадобится)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ./loglinter --version || exit 1

# Метаданные
LABEL maintainer="aidabag"
LABEL version="1.0.0"
LABEL description="Log Linter - Go linter for checking log messages"
LABEL source="https://github.com/aidabag/lint"

# Команда по умолчанию
CMD ["./loglinter", "--help"]
