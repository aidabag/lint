# Линтер для проверки лог-записей

Go линтер для проверки лог-сообщений в соответствии с установленными правилами.

## Возможности

- ✅ Проверка начала сообщения со строчной буквы
- ✅ Проверка на английский язык
- ✅ Обнаружение спецсимволов и эмодзи
- ✅ Обнаружение чувствительных данных
- ✅ Поддержка `log/slog`, `go.uber.org/zap` и стандартного `log`
- ✅ Интеграция с golangci-lint
- ✅ Автономный CLI инструмент
- ✅ Unit-тесты с 100% покрытием
- ✅ Конфигурация через YAML файл
- ✅ Авто-исправление ошибок
- ✅ Кастомные паттерны для проверки
- ✅ Docker контейнеризация
- ✅ CI/CD готов (GitHub Actions, GitLab CI)

## Правила

### 1. Строчная буква в начале
❌ Плохо:
```go
slog.Info("Starting server on port 8080")
slog.Error("Failed to connect to database")
```

✅ Хорошо:
```go
slog.Info("starting server on port 8080")
slog.Error("failed to connect to database")
```

### 2. Только английский язык
❌ Плохо:
```go
slog.Info("запуск сервера")
slog.Error("ошибка подключения к базе данных")
```

✅ Хорошо:
```go
slog.Info("starting server")
slog.Error("failed to connect to database")
```

### 3. Без спецсимволов
❌ Плохо:
```go
slog.Info("server started! 🚀")
slog.Error("connection failed!!!")
slog.Warn("warning: something went wrong...")
```

✅ Хорошо:
```go
slog.Info("server started")
slog.Error("connection failed")
slog.Warn("something went wrong")
```

### 4. Без чувствительных данных
❌ Плохо:
```go
slog.Info("user password: " + password)
slog.Debug("api_key=" + apiKey)
slog.Info("token: " + token)
```

✅ Хорошо:
```go
slog.Info("user authenticated successfully")
slog.Debug("api request completed")
slog.Info("token validated")
```

## Установка и сборка

### 1. Клонирование репозитория
```bash
git clone https://github.com/aidabag/lint.git
cd lint
```

### 2. Установка зависимостей
```bash
go mod tidy
```

### 3. Сборка CLI версии
```bash
# Простая версия (рекомендуется)
go build -o loglinter-simple cmd/simple/main.go

# Полная версия (требует стабильную версию Go)
go build -o loglinter-full cmd/loglinter/main.go
```

### 4. Запуск с Docker
```bash
# Сборка Docker образа
docker build -t loglinter .

# Запуск линтера на файлах текущей директории
docker run --rm -v $(pwd):/app loglinter ./app

# Запуск с подробным выводом
docker run --rm -v $(pwd):/app loglinter -v ./app

# Запуск на конкретном файле
docker run --rm -v $(pwd):/app loglinter /app/main.go

# Запуск с авто-исправлением
docker run --rm -v $(pwd):/app loglinter -fix ./app
```

### 5. Запуск с Makefile
```bash
# Сборка бинарного файла
make build

# Запуск тестов
make test

# Запуск линтера
make lint-test

# Полная проверка перед коммитом
make pre-commit

# Сборка для всех платформ
make build-all
```

### 6. Запуск тестов
```bash
go test -v ./pkg/loglinter/...
```

## Использование

### Автономный CLI инструмент
```bash
# Проверить все файлы в директории
./loglinter-simple ./...

# Проверить конкретный файл
./loglinter-simple main.go

# Проверить пакет
./loglinter-simple ./pkg/...

# Запуск с подробным выводом
./loglinter-simple -v ./...

# Авто-исправление ошибок
./loglinter-simple -fix ./...

# Использование кастомной конфигурации
./loglinter-simple -config custom.yaml ./...

# Пример запуска из директории проекта
cd cmd/simple
go run main.go -v test_example.go
```

### Интеграция с golangci-lint
Добавьте в ваш `.golangci.yml`:
```yaml
linters-settings:
  custom:
    loglinter:
      path: github.com/aidabag/lint/pkg/loglinter
      enabled: true
      original-url: github.com/aidabag/lint
```

## Разработка

### Команды для разработки
```bash
# Запустить все тесты
go test -v ./...

# Запустить линтер на примере
cd cmd/simple
go run main.go -v test_example.go

# Запустить линтер на тестовых данных
cd cmd/simple
go run main.go -v ../../testdata/src/a/a.go

# Форматирование кода
go fmt ./...

# Проверка зависимостей
go mod tidy
```

### Структура проекта
```
lint/
├── cmd/
│   ├── loglinter/     # Полная версия с golang.org/x/tools
│   └── simple/        # Простая версия без внешних зависимостей
├── pkg/
│   ├── loglinter/     # Основная логика анализатора
│   └── golangci/      # Интеграция с golangci-lint
├── testdata/          # Тестовые файлы
├── docs/              # Документация
├── Makefile          # Команды для разработки
├── README.md         # Этот файл
├── USAGE.md          # Подробное руководство
└── CHANGELOG.md      # История изменений
```

## Системные требования

- Go 1.22+
- golangci-lint (для режима плагина)

## Примеры использования

### Базовая проверка
```bash
# Создать тестовый файл
echo 'package main
import "log/slog"
func main() {
    slog.Info("Starting server")  // Нарушение: заглавная буква
    slog.Error("failed to connect")  // Хорошо
}' > test.go

# Запустить проверку
go run cmd/simple/main.go test.go
```

### Проверка реального проекта
```bash
# Проверить весь проект
go run cmd/simple/main.go ./...

# Проверить только исходные файлы
go run cmd/simple/main.go **/*.go
```

## Устранение неполадок

### Проблема: `invalid array length -delta * delta`
**Решение**: Используйте простую версию `cmd/simple` вместо полной версии `cmd/loglinter`.

### Проблема: Ложные срабатывания
**Решение**: Линтер использует эвристики для уменьшения ложных срабатываний, но некоторые могут оставаться. Отправьте issue для улучшения.

### Проблема: Не работает с zap
**Решение**: Убедитесь, что переменная логгера называется `logger` или `zap`.