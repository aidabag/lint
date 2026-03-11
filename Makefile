# Makefile для Log Linter

.PHONY: build test clean install lint docker ci-local security benchmark

# Переменные
BINARY_NAME=loglinter
MAIN_PATH=cmd/simple/main.go
PKG_PATH=./pkg/loglinter
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags="-X main.version=$(VERSION) -s -w"

# Сборка бинарного файла
build:
	@echo "Сборка бинарного файла..."
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) $(MAIN_PATH)

# Сборка для всех платформ
build-all:
	@echo "Сборка для всех платформ..."
	mkdir -p dist
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

# Запуск тестов
test:
	@echo "Запуск тестов..."
	go test -v -race -coverprofile=coverage.out ./...

# Запуск тестов с покрытием
test-coverage:
	@echo "Запуск тестов с покрытием..."
	go test -v -covermode=atomic -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Отчет о покрытии сохранен в coverage.html"

# Запуск Race тестов
test-race:
	@echo "Запуск race тестов..."
	go test -race -short ./...

# Запуск бенчмарков
benchmark:
	@echo "Запуск бенчмарков..."
	go test -bench=. -benchmem -count=3 ./... | tee benchmark.txt

# Установка линтера
install:
	@echo "Установка линтера..."
	go install $(LDFLAGS) $(MAIN_PATH)

# Очистка
clean:
	@echo "Очистка..."
	go clean
	rm -rf bin/
	rm -rf dist/
	rm -f coverage.out coverage.html benchmark.txt
	rm -f *.backup

# Запуск линтера на тестовых данных
lint-test:
	@echo "Запуск линтера на тестовых данных..."
	go run $(MAIN_PATH) -v ./testdata/...

# Форматирование кода
fmt:
	@echo "Форматирование кода..."
	go fmt ./...

# Запуск golangci-lint
lint:
	@echo "Запуск golangci-lint..."
	golangci-lint run --timeout=5m

# Проверка зависимостей
mod-tidy:
	@echo "Проверка зависимостей..."
	go mod tidy
	go mod verify

# Проверка уязвимостей
security:
	@echo "Проверка уязвимостей..."
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

# Запуск gosec
security-scan:
	@echo "Запуск gosec..."
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	gosec ./...

# Docker команды
docker-build:
	@echo "Сборка Docker образа..."
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest

docker-run:
	@echo "Запуск Docker контейнера..."
	docker run --rm -v $(pwd):/app $(BINARY_NAME):latest ./app

docker-push:
	@echo "Публикация Docker образа..."
	docker push $(BINARY_NAME):$(VERSION)
	docker push $(BINARY_NAME):latest

# Полная проверка перед коммитом
pre-commit: fmt mod-tidy test lint-test lint
	@echo "Все проверки пройдены!"

# Полная проверка CI
ci: fmt mod-tidy test-race test-coverage lint-test lint security security-scan docker-build
	@echo "CI проверки завершены!"

# Локальный CI (использует act для GitHub Actions)
ci-local:
	@echo "Запуск локального CI..."
	act -j test
	act -j lint
	act -j build

# Генерация документации
docs:
	@echo "Генерация документации..."
	godoc -http=:6060 &
	@echo "Документация доступна на http://localhost:6060"

# Профилирование
profile-cpu:
	@echo "Профилирование CPU..."
	go test -cpuprofile=cpu.prof -bench=. ./...
	go tool pprof cpu.prof

profile-mem:
	@echo "Профилирование памяти..."
	go test -memprofile=mem.prof -bench=. ./...
	go tool pprof mem.prof

# Версия
version:
	@echo "Версия: $(VERSION)"

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  build          - Сборка бинарного файла"
	@echo "  build-all      - Сборка для всех платформ"
	@echo "  test           - Запуск тестов"
	@echo "  test-coverage  - Запуск тестов с покрытием"
	@echo "  test-race      - Запуск race тестов"
	@echo "  benchmark      - Запуск бенчмарков"
	@echo "  install        - Установка линтера"
	@echo "  clean          - Очистка"
	@echo "  lint-test      - Запуск линтера на тестовых данных"
	@echo "  fmt            - Форматирование кода"
	@echo "  lint           - Запуск golangci-lint"
	@echo "  mod-tidy       - Проверка зависимостей"
	@echo "  security       - Проверка уязвимостей"
	@echo "  security-scan  - Запуск gosec"
	@echo "  docker-build   - Сборка Docker образа"
	@echo "  docker-run     - Запуск Docker контейнера"
	@echo "  docker-push    - Публикация Docker образа"
	@echo "  pre-commit     - Полная проверка перед коммитом"
	@echo "  ci             - Полная проверка CI"
	@echo "  ci-local       - Локальный CI"
	@echo "  docs           - Генерация документации"
	@echo "  profile-cpu    - Профилирование CPU"
	@echo "  profile-mem    - Профилирование памяти"
	@echo "  version        - Показать версию"
	@echo "  help           - Показать эту справку"

# Release задачи
release: clean test build-all
	@echo "Подготовка релиза $(VERSION)..."
	cd dist && sha256sum * > checksums.txt
	@echo "Артефакты готовы в директории dist/"
	@echo "Не забудьте создать тег: git tag v$(VERSION)"
	@echo "И запушить его: git push origin v$(VERSION)"

# Сборка релиза
release: clean test build
	@echo "Сборка релиза завершена"

# Запуск на примерах
example:
	go run $(MAIN_PATH) ./testdata/src/a/...
