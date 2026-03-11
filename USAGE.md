# Log Linter Usage Guide

## Установка

### 1. Как standalone линтер

```bash
go install github.com/aidabag/lint/cmd/loglinter@latest
```

### 2. Как плагин для golangci-lint

Добавьте в ваш `.golangci.yml`:

```yaml
linters-settings:
  custom:
    loglinter:
      path: github.com/aidabag/lint/pkg/loglinter
      enabled: true
      original-url: github.com/aidabag/lint
```

## Правила проверки

### 1. Строчная буква в начале

❌ Плохо:
```go
slog.Info("Starting server")
log.Error("Failed to connect")
```

✅ Хорошо:
```go
slog.Info("starting server")
log.Error("failed to connect")
```

### 2. Только английский язык

❌ Плохо:
```go
slog.Info("запуск сервера")
log.Error("ошибка подключения")
```

✅ Хорошо:
```go
slog.Info("starting server")
log.Error("connection failed")
```

### 3. Без спецсимволов и эмодзи

❌ Плохо:
```go
slog.Info("server started! 🚀")
log.Warn("warning: check config!!!")
```

✅ Хорошо:
```go
slog.Info("server started")
log.Warn("check configuration")
```

### 4. Без чувствительных данных

❌ Плохо:
```go
slog.Info("user password: " + password)
log.Debug("api_key=" + apiKey)
```

✅ Хорошо:
```go
slog.Info("user authenticated successfully")
log.Debug("api request completed")
```

## Поддерживаемые логгеры

- `log/slog`
- `go.uber.org/zap`
- Стандартный `log`

## Примеры использования

### Запуск из командной строки

```bash
# Проверить все файлы в проекте
loglinter ./...

# Проверить конкретный пакет
loglinter ./pkg/...

# Проверить конкретный файл
loglinter main.go
```

### Интеграция с golangci-lint

```bash
# Установить линтер
go install github.com/aidabag/lint@latest

# Запустить golangci-lint с новым правилом
golangci-lint run
```

## Тестирование

```bash
# Запустить все тесты
go test ./...

# Запустить тесты с покрытием
go test -cover ./...

# Запустить конкретный тест
go test ./pkg/loglinter/...
```

## Разработка

```bash
# Сборка проекта
go build ./cmd/loglinter

# Запуск в режиме отладки
go run cmd/loglinter/main.go ./testdata/...
```
