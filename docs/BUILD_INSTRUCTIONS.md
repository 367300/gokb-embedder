# Инструкции по сборке и использованию

## Сборка исполняемых файлов

### Автоматическая сборка (рекомендуется)

Используйте Makefile для автоматической сборки всех платформ:

```bash
# Сборка для всех платформ (Linux, macOS, Windows)
make build-all

# Сборка только для текущей платформы
make build

# Очистка и сборка
make clean build
```

### Ручная сборка

#### Для Ubuntu/Linux:
```bash
# AMD64 (x86_64)
GOOS=linux GOARCH=amd64 go build -o gokb-embedder-linux-amd64 cmd/main.go

# ARM64
GOOS=linux GOARCH=arm64 go build -o gokb-embedder-linux-arm64 cmd/main.go
```

#### Для Windows:
```bash
# AMD64 (x86_64)
GOOS=windows GOARCH=amd64 go build -o gokb-embedder-windows-amd64.exe cmd/main.go
```

#### Для macOS:
```bash
# AMD64 (Intel)
GOOS=darwin GOARCH=amd64 go build -o gokb-embedder-darwin-amd64 cmd/main.go

# ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o gokb-embedder-darwin-arm64 cmd/main.go
```

## Размеры файлов

После сборки вы получите следующие файлы:

| Платформа | Архитектура | Размер | Файл |
|-----------|-------------|--------|------|
| Linux | AMD64 | ~12 МБ | `gokb-embedder-linux-amd64` |
| Linux | ARM64 | ~8 МБ | `gokb-embedder-linux-arm64` |
| macOS | AMD64 | ~8.5 МБ | `gokb-embedder-darwin-amd64` |
| macOS | ARM64 | ~8 МБ | `gokb-embedder-darwin-arm64` |
| Windows | AMD64 | ~8.8 МБ | `gokb-embedder-windows-amd64.exe` |

## Использование

### 1. Подготовка

1. Скопируйте нужный исполняемый файл в целевую систему
2. Создайте файл `.env` на основе `env.example`:

```bash
# Скопируйте пример конфигурации
cp env.example .env

# Отредактируйте файл
nano .env
```

3. Укажите ваш OpenAI API ключ в `.env`:

```env
OPENAI_API_KEY=your_openai_api_key_here
ROOT_DIR=.
FILE_EXTENSIONS=.py,.md,.yml,.conf
DB_PATH=embeddings.sqlite3
N_COMMITS=3
TOKEN_LIMIT=1600
LOG_LEVEL=info
```

### 2. Запуск

#### Linux/macOS:
```bash
# Сделайте файл исполняемым
chmod +x gokb-embedder-linux-amd64

# Запустите
./gokb-embedder-linux-amd64
```

#### Windows:
```cmd
# Запустите из командной строки
gokb-embedder-windows-amd64.exe
```

### 3. Альтернативные способы запуска

#### С переменными окружения:
```bash
OPENAI_API_KEY=your_key ROOT_DIR=/path/to/project ./gokb-embedder-linux-amd64
```

#### С кастомной конфигурацией:
```bash
# Создайте .env файл в нужной директории
echo "OPENAI_API_KEY=your_key" > .env
echo "ROOT_DIR=/path/to/project" >> .env

# Запустите из этой директории
./gokb-embedder-linux-amd64
```

## Требования к системе

### Минимальные требования:
- **Linux**: glibc 2.17+ (Ubuntu 18.04+, CentOS 7+)
- **Windows**: Windows 7+ (64-bit)
- **macOS**: macOS 10.12+ (Sierra)

### Дополнительные требования:
- **Git**: для получения истории коммитов (опционально)
- **SQLite3**: встроен в исполняемый файл
- **Интернет**: для доступа к OpenAI API

## Проверка работоспособности

### 1. Проверка версии:
```bash
./gokb-embedder-linux-amd64 --version
```

### 2. Тестовый запуск:
```bash
# Создайте тестовый проект
mkdir test-project
cd test-project

# Создайте .env файл
echo "OPENAI_API_KEY=your_key" > .env

# Создайте тестовый Python файл
echo 'def hello():
    return "Hello, World!"' > test.py

# Запустите
../gokb-embedder-linux-amd64
```

### 3. Проверка результатов:
```bash
# Проверьте, что база данных создалась
ls -la embeddings.sqlite3

# Проверьте содержимое базы данных
sqlite3 embeddings.sqlite3 "SELECT COUNT(*) FROM embeddings;"
```

## Распространение

### Для внутреннего использования:
1. Скопируйте нужный исполняемый файл
2. Создайте инструкцию по настройке `.env`
3. Предоставьте примеры конфигурации

### Для публичного распространения:
1. Создайте релиз с помощью `make release`
2. Загрузите файлы на GitHub Releases
3. Предоставьте документацию по установке

## Устранение неполадок

### Ошибка "permission denied":
```bash
chmod +x gokb-embedder-linux-amd64
```

### Ошибка "command not found":
```bash
# Добавьте в PATH или используйте полный путь
./gokb-embedder-linux-amd64
```

### Ошибка OpenAI API:
- Проверьте правильность API ключа
- Убедитесь в наличии средств на аккаунте
- Проверьте доступность api.openai.com

### Ошибка Git:
- Убедитесь, что Git установлен
- Проверьте, что проект является Git репозиторием
- Проверьте права доступа к `.git` директории

## Оптимизация

### Уменьшение размера файла:
```bash
# Сборка с оптимизациями
go build -ldflags="-s -w" -o gokb-embedder cmd/main.go
```

### Сборка для конкретной архитектуры:
```bash
# Только для вашей системы
go build -o gokb-embedder cmd/main.go
```

## Безопасность

- Не включайте API ключи в исполняемые файлы
- Используйте переменные окружения или `.env` файлы
- Ограничьте доступ к исполняемым файлам
- Регулярно обновляйте API ключи 