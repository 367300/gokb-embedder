package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config содержит все настройки приложения
type Config struct {
	// OpenAI настройки
	OpenAIAPIKey string

	// Настройки проекта
	RootDir        string
	FileExtensions []string
	DBPath         string
	NCommits       int
	TokenLimit     int

	// Настройки логирования
	LogLevel string

	// Режим работы (для CLI)
	OperationMode string
}

// Load загружает конфигурацию из .env файла и переменных окружения
func Load() (*Config, error) {
	// Загружаем .env файл если он существует
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return nil, err
		}
	}

	// Получаем значения из переменных окружения
	openAIKey := getEnv("OPENAI_API_KEY", "")
	if openAIKey == "" {
		return nil, ErrMissingOpenAIKey
	}

	rootDir := getEnv("ROOT_DIR", ".")
	fileExtensionsStr := getEnv("FILE_EXTENSIONS", ".py,.md,.yml,.conf")
	fileExtensions := parseFileExtensions(fileExtensionsStr)
	dbPath := getEnv("DB_PATH", "embeddings.sqlite3")
	nCommits := getEnvAsInt("N_COMMITS", 3)
	tokenLimit := getEnvAsInt("TOKEN_LIMIT", 1600)
	logLevel := getEnv("LOG_LEVEL", "info")

	return &Config{
		OpenAIAPIKey:   openAIKey,
		RootDir:        rootDir,
		FileExtensions: fileExtensions,
		DBPath:         dbPath,
		NCommits:       nCommits,
		TokenLimit:     tokenLimit,
		LogLevel:       logLevel,
	}, nil
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt получает целочисленное значение переменной окружения
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// parseFileExtensions парсит строку расширений файлов в слайс
func parseFileExtensions(extensions string) []string {
	if extensions == "" {
		return []string{".py", ".md", ".yml", ".conf"}
	}

	var result []string
	// Разделяем по запятой, а не по filepath.SplitList
	parts := strings.Split(extensions, ",")
	for _, ext := range parts {
		ext = strings.TrimSpace(ext)
		if ext != "" {
			result = append(result, ext)
		}
	}
	return result
}
