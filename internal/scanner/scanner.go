package scanner

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Scanner предоставляет методы для сканирования файлов
type Scanner struct {
	rootDir           string
	fileExtensions    []string
	gitignorePatterns []string
}

// NewScanner создаёт новый сканер файлов
func NewScanner(rootDir string, fileExtensions []string) *Scanner {
	return &Scanner{
		rootDir:        rootDir,
		fileExtensions: fileExtensions,
	}
}

// LoadGitignore загружает правила из .gitignore файла
func (s *Scanner) LoadGitignore() error {
	gitignorePath := filepath.Join(s.rootDir, ".gitignore")

	file, err := os.Open(gitignorePath)
	if err != nil {
		// .gitignore может не существовать, это нормально
		return nil
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}

	s.gitignorePatterns = patterns
	return scanner.Err()
}

// ScanFiles сканирует файлы в директории
func (s *Scanner) ScanFiles() ([]string, error) {
	var files []string
	var totalFiles, matchedFiles, ignoredFiles int

	fmt.Printf("🔍 Сканирование в: %s\n", s.rootDir)
	fmt.Printf("📝 Ищем расширения: %v\n", s.fileExtensions)
	fmt.Printf("🚫 Gitignore паттернов: %d\n", len(s.gitignorePatterns))

	err := filepath.Walk(s.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Пропускаем директории
		if info.IsDir() {
			return nil
		}

		totalFiles++
		ext := filepath.Ext(path)

		// Проверяем расширение файла
		if !contains(s.fileExtensions, ext) {
			return nil
		}

		matchedFiles++

		// Проверяем .gitignore
		if s.shouldIgnoreFile(path) {
			ignoredFiles++
			return nil
		}

		// Получаем относительный путь
		relPath, err := filepath.Rel(s.rootDir, path)
		if err != nil {
			return fmt.Errorf("ошибка получения относительного пути: %w", err)
		}

		files = append(files, relPath)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("ошибка сканирования файлов: %w", err)
	}

	fmt.Printf("📊 Статистика сканирования:\n")
	fmt.Printf("  - Всего файлов: %d\n", totalFiles)
	fmt.Printf("  - Подходящих расширений: %d\n", matchedFiles)
	fmt.Printf("  - Игнорировано .gitignore: %d\n", ignoredFiles)
	fmt.Printf("  - Обработано: %d\n", len(files))

	return files, nil
}

// shouldIgnoreFile проверяет, должен ли файл быть проигнорирован
func (s *Scanner) shouldIgnoreFile(filePath string) bool {
	relPath, err := filepath.Rel(s.rootDir, filePath)
	if err != nil {
		return false
	}

	// Нормализуем разделители путей
	relPath = strings.ReplaceAll(relPath, "\\", "/")

	for _, pattern := range s.gitignorePatterns {
		if s.matchesPattern(relPath, pattern) {
			return true
		}
	}

	return false
}

// matchesPattern проверяет, соответствует ли путь паттерну
func (s *Scanner) matchesPattern(path, pattern string) bool {
	// Убираем слеш в начале, если есть
	if strings.HasPrefix(pattern, "/") {
		pattern = pattern[1:]
	}

	// Проверяем точное совпадение
	if s.globMatch(path, pattern) {
		return true
	}

	// Проверяем, находится ли файл внутри директории
	if strings.HasSuffix(pattern, "/") {
		dirPattern := pattern[:len(pattern)-1] // убираем trailing slash
		pathParts := strings.Split(path, "/")
		for _, part := range pathParts {
			if part == dirPattern {
				return true
			}
		}
	} else {
		// Проверяем, начинается ли путь с этой директории
		if strings.HasPrefix(path, pattern+"/") {
			return true
		}

		// Проверяем точное совпадение с файлом
		if path == pattern {
			return true
		}

		// Проверяем, является ли паттерн директорией в пути
		pathParts := strings.Split(path, "/")
		for _, part := range pathParts {
			if part == pattern {
				return true
			}
		}
	}

	return false
}

// globMatch проверяет соответствие пути glob-паттерну
func (s *Scanner) globMatch(path, pattern string) bool {
	// Простая реализация glob-паттернов
	// В реальном проекте можно использовать более сложную логику

	// Заменяем * на .*
	pattern = strings.ReplaceAll(pattern, "*", ".*")

	// Добавляем начало и конец строки
	pattern = "^" + pattern + "$"

	// Простая проверка (в реальном проекте лучше использовать regexp)
	return strings.Contains(path, strings.ReplaceAll(pattern, ".*", ""))
}

// contains проверяет, содержится ли элемент в слайсе
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
