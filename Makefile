# Makefile для GoKB Embedder

# Переменные
BINARY_NAME=gokb-embedder
BUILD_DIR=build
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

# Цвета для вывода
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

.PHONY: all build clean test lint run help install deps

# Цель по умолчанию
all: clean build

# Сборка приложения
build:
	@echo "$(GREEN)🔨 Сборка приложения...$(NC)"
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/main.go
	@echo "$(GREEN)✅ Приложение собрано: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

# Сборка для Windows
build-windows:
	@echo "$(GREEN)🔨 Сборка для Windows...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe cmd/main.go
	@echo "$(GREEN)✅ Приложение собрано для Windows: $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe$(NC)"

# Сборка для разных платформ
build-all: clean
	@echo "$(GREEN)🔨 Сборка для всех платформ...$(NC)"
	@mkdir -p $(BUILD_DIR)
	
	# Linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 cmd/main.go
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 cmd/main.go
	
	# macOS
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 cmd/main.go
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 cmd/main.go
	
	# Windows
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe cmd/main.go
	
	@echo "$(GREEN)✅ Сборка завершена для всех платформ$(NC)"

# Установка зависимостей
deps:
	@echo "$(GREEN)📦 Установка зависимостей...$(NC)"
	go mod download
	go mod tidy
	@echo "$(GREEN)✅ Зависимости установлены$(NC)"

# Установка инструментов разработки
install-tools:
	@echo "$(GREEN)🔧 Установка инструментов разработки...$(NC)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@echo "$(GREEN)✅ Инструменты установлены$(NC)"

# Установка системных зависимостей
install-deps:
	@echo "$(GREEN)🔧 Установка системных зависимостей...$(NC)"
	@if command -v apt-get >/dev/null 2>&1; then \
		sudo apt-get update && sudo apt-get install -y zip unzip; \
	elif command -v yum >/dev/null 2>&1; then \
		sudo yum install -y zip unzip; \
	elif command -v dnf >/dev/null 2>&1; then \
		sudo dnf install -y zip unzip; \
	elif command -v pacman >/dev/null 2>&1; then \
		sudo pacman -S --noconfirm zip unzip; \
	else \
		echo "$(YELLOW)⚠️  Не удалось определить пакетный менеджер. Установите zip вручную.$(NC)"; \
	fi
	@echo "$(GREEN)✅ Системные зависимости установлены$(NC)"

# Запуск приложения
run: build
	@echo "$(GREEN)🚀 Запуск приложения...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME)

# Запуск в режиме разработки
dev:
	@echo "$(GREEN)🔧 Запуск в режиме разработки...$(NC)"
	go run cmd/main.go

# Тестирование
test:
	@echo "$(GREEN)🧪 Запуск тестов...$(NC)"
	go test -v ./...
	@echo "$(GREEN)✅ Тесты завершены$(NC)"

# Покрытие тестами
test-coverage:
	@echo "$(GREEN)🧪 Запуск тестов с покрытием...$(NC)"
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ Отчёт о покрытии создан: coverage.html$(NC)"

# Линтинг
lint:
	@echo "$(GREEN)🔍 Проверка кода...$(NC)"
	golangci-lint run
	@echo "$(GREEN)✅ Линтинг завершён$(NC)"

# Форматирование кода
fmt:
	@echo "$(GREEN)🎨 Форматирование кода...$(NC)"
	go fmt ./...
	@echo "$(GREEN)✅ Код отформатирован$(NC)"

# Проверка безопасности
security:
	@echo "$(GREEN)🔒 Проверка безопасности...$(NC)"
	gosec ./...
	@echo "$(GREEN)✅ Проверка безопасности завершена$(NC)"

# Очистка
clean:
	@echo "$(YELLOW)🧹 Очистка...$(NC)"
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "$(GREEN)✅ Очистка завершена$(NC)"

# Установка в систему
install: build
	@echo "$(GREEN)📦 Установка в систему...$(NC)"
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "$(GREEN)✅ Приложение установлено$(NC)"

# Создание релиза
release: clean build-all
	@echo "$(GREEN)📦 Создание релиза...$(NC)"
	@mkdir -p release
	cd $(BUILD_DIR) && tar -czf ../release/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	cd $(BUILD_DIR) && tar -czf ../release/$(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64
	cd $(BUILD_DIR) && tar -czf ../release/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64
	cd $(BUILD_DIR) && tar -czf ../release/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64
	@if command -v zip >/dev/null 2>&1; then \
		cd $(BUILD_DIR) && zip ../release/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe; \
	else \
		echo "$(YELLOW)⚠️  zip не найден, создаём tar.gz для Windows...$(NC)"; \
		cd $(BUILD_DIR) && tar -czf ../release/$(BINARY_NAME)-$(VERSION)-windows-amd64.tar.gz $(BINARY_NAME)-windows-amd64.exe; \
	fi
	@echo "$(GREEN)✅ Релиз создан в папке release/$(NC)"

# Проверка версии
version:
	@echo "$(GREEN)📋 Версия: $(VERSION)$(NC)"

# Помощь
help:
	@echo "$(GREEN)📖 Доступные команды:$(NC)"
	@echo "  $(YELLOW)build$(NC)        - Сборка приложения"
	@echo "  $(YELLOW)build-windows$(NC) - Сборка для Windows"
	@echo "  $(YELLOW)build-all$(NC)    - Сборка для всех платформ"
	@echo "  $(YELLOW)deps$(NC)         - Установка зависимостей"
	@echo "  $(YELLOW)install-tools$(NC) - Установка инструментов разработки"
	@echo "  $(YELLOW)install-deps$(NC) - Установка системных зависимостей"
	@echo "  $(YELLOW)run$(NC)          - Запуск приложения"
	@echo "  $(YELLOW)dev$(NC)          - Запуск в режиме разработки"
	@echo "  $(YELLOW)test$(NC)         - Запуск тестов"
	@echo "  $(YELLOW)test-coverage$(NC) - Тесты с покрытием"
	@echo "  $(YELLOW)lint$(NC)         - Проверка кода"
	@echo "  $(YELLOW)fmt$(NC)          - Форматирование кода"
	@echo "  $(YELLOW)security$(NC)     - Проверка безопасности"
	@echo "  $(YELLOW)clean$(NC)        - Очистка"
	@echo "  $(YELLOW)install$(NC)      - Установка в систему"
	@echo "  $(YELLOW)release$(NC)      - Создание релиза"
	@echo "  $(YELLOW)version$(NC)      - Показать версию"
	@echo "  $(YELLOW)help$(NC)         - Показать эту справку" 