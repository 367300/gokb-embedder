# Быстрый старт GoKB Embedder

## 🚀 Быстрая установка

### 1. Скачайте исполняемый файл

Выберите файл для вашей системы:

- **Ubuntu/Linux (AMD64)**: `gokb-embedder-linux-amd64`
- **Ubuntu/Linux (ARM64)**: `gokb-embedder-linux-arm64` 
- **Windows**: `gokb-embedder-windows-amd64.exe`
- **macOS (Intel)**: `gokb-embedder-darwin-amd64`
- **macOS (Apple Silicon)**: `gokb-embedder-darwin-arm64`

### 2. Настройте конфигурацию

```bash
# Создайте файл конфигурации
cat > .env << EOF
OPENAI_API_KEY=your_openai_api_key_here
ROOT_DIR=.
FILE_EXTENSIONS=.py,.md,.yml,.conf
DB_PATH=embeddings.sqlite3
N_COMMITS=3
TOKEN_LIMIT=1600
LOG_LEVEL=info
EOF
```

### 3. Запустите

```bash
# Linux/macOS
chmod +x gokb-embedder-linux-amd64
./gokb-embedder-linux-amd64

# Windows
gokb-embedder-windows-amd64.exe
```

## 📋 Что происходит

1. **Сканирование** — поиск файлов с указанными расширениями
2. **Парсинг** — извлечение блоков кода и документации
3. **Git история** — получение последних коммитов (если Git репозиторий)
4. **Эмбединги** — создание векторных представлений через OpenAI API
5. **Сохранение** — запись в SQLite базу данных

## 📊 Результат

После выполнения вы получите:
- `embeddings.sqlite3` — база данных с эмбедингами
- Логи процесса в консоли
- Прогресс-бары для отслеживания

## 🔍 Проверка результатов

```bash
# Проверьте размер базы данных
ls -lh embeddings.sqlite3

# Посмотрите количество записей
sqlite3 embeddings.sqlite3 "SELECT COUNT(*) FROM embeddings;"

# Просмотрите примеры записей
sqlite3 embeddings.sqlite3 "SELECT file_path, block_type, start_line, end_line FROM embeddings LIMIT 5;"
```

## ⚙️ Настройки

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `OPENAI_API_KEY` | Ключ OpenAI API | **обязательно** |
| `ROOT_DIR` | Директория для сканирования | `.` |
| `FILE_EXTENSIONS` | Расширения файлов | `.py,.md,.yml,.conf` |
| `DB_PATH` | Путь к базе данных | `embeddings.sqlite3` |
| `N_COMMITS` | Количество коммитов | `3` |
| `TOKEN_LIMIT` | Лимит токенов на блок | `1600` |
| `LOG_LEVEL` | Уровень логирования | `info` |

## 🚫 Игнорирование файлов

Скрипт автоматически читает `.gitignore` и игнорирует:
- Файлы и папки по паттернам `.gitignore`
- Поддиректории игнорируемых папок
- Файлы с указанными расширениями

## 🔄 Повторный запуск

При повторном запуске скрипт:
- Проверяет изменения файлов по MD5-хешам
- Обновляет только изменённые файлы
- Сохраняет время и деньги на API

## 🐛 Устранение проблем

### Ошибка "permission denied":
```bash
chmod +x gokb-embedder-linux-amd64
```

### Ошибка OpenAI API:
- Проверьте правильность API ключа
- Убедитесь в наличии средств на аккаунте

### Файлы не обрабатываются:
- Проверьте расширения в `FILE_EXTENSIONS`
- Убедитесь, что файлы не в `.gitignore`

## 📚 Дополнительная информация

- [Полная документация](README.md)
- [Архитектура проекта](docs/ARCHITECTURE.md)
- [Инструкции по сборке](docs/BUILD_INSTRUCTIONS.md)
- [Сравнение с Python версией](docs/COMPARISON.md) 