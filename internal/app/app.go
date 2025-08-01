package app

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"

	"gokb-embedder/internal/config"
	"gokb-embedder/internal/database"
	"gokb-embedder/internal/git"
	"gokb-embedder/internal/models"
	"gokb-embedder/internal/openai"
	"gokb-embedder/internal/parsers"
	"gokb-embedder/internal/scanner"
)

// App представляет основное приложение
type App struct {
	config     *config.Config
	logger     *logrus.Logger
	database   *database.Database
	openai     *openai.Client
	scanner    *scanner.Scanner
	parsers    *parsers.ParserRegistry
	gitService *git.GitService
}

// New создаёт новое приложение
func New(cfg *config.Config) *App {
	// Настраиваем логирование
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	return &App{
		config:  cfg,
		logger:  logger,
		parsers: parsers.NewParserRegistry(),
	}
}

// Run запускает приложение
func (r *App) Run() error {
	r.logger.Info("🚀 Запуск генератора эмбедингов")

	// Инициализируем компоненты
	if err := r.initialize(); err != nil {
		return fmt.Errorf("ошибка инициализации: %w", err)
	}
	defer r.cleanup()

	// Сканируем файлы
	files, err := r.scanFiles()
	if err != nil {
		return fmt.Errorf("ошибка сканирования файлов: %w", err)
	}

	// Проверяем изменения файлов
	filesToProcess, err := r.checkFileChanges(files)
	if err != nil {
		return fmt.Errorf("ошибка проверки изменений файлов: %w", err)
	}

	if len(filesToProcess) == 0 {
		r.logger.Info("✅ Все файлы актуальны, обновление не требуется!")
		return nil
	}

	// Обрабатываем файлы
	if err := r.processFiles(filesToProcess); err != nil {
		return fmt.Errorf("ошибка обработки файлов: %w", err)
	}

	r.logger.Info("✅ Готово! Эмбединги сохранены в " + r.config.DBPath)
	return nil
}

// initialize инициализирует все компоненты приложения
func (r *App) initialize() error {
	r.logger.Info("🔧 Инициализация компонентов...")

	// Инициализируем базу данных
	db, err := database.NewDatabase(r.config.DBPath)
	if err != nil {
		return fmt.Errorf("ошибка инициализации базы данных: %w", err)
	}
	r.database = db

	// Инициализируем OpenAI клиент
	r.openai = openai.NewClient(r.config.OpenAIAPIKey)

	// Инициализируем сканер
	r.scanner = scanner.NewScanner(r.config.RootDir, r.config.FileExtensions)
	if err := r.scanner.LoadGitignore(); err != nil {
		r.logger.Warn("⚠️ Не удалось загрузить .gitignore")
	}

	// Регистрируем парсеры
	r.registerParsers()

	// Инициализируем Git сервис (если это Git репозиторий)
	if git.IsGitRepository(r.config.RootDir) {
		gitService, err := git.NewGitService(r.config.RootDir)
		if err != nil {
			r.logger.Warn("⚠️ Не удалось инициализировать Git сервис")
		} else {
			r.gitService = gitService
		}
	}

	return nil
}

// cleanup очищает ресурсы
func (r *App) cleanup() {
	if r.database != nil {
		if err := r.database.Close(); err != nil {
			r.logger.Error("Ошибка закрытия базы данных:", err)
		}
	}
}

// registerParsers регистрирует все доступные парсеры
func (r *App) registerParsers() {
	// Регистрируем Python парсер
	pythonParser := parsers.NewPythonParser()
	r.parsers.Register(pythonParser)

	// Регистрируем текстовый парсер
	textParser := parsers.NewTextParser(r.config.TokenLimit)
	r.parsers.Register(textParser)

	r.logger.Infof("📝 Зарегистрировано парсеров: %d", len(r.parsers.GetAllParsers()))
}

// scanFiles сканирует файлы в проекте
func (r *App) scanFiles() ([]string, error) {
	r.logger.Info("🔍 Сканирование файлов...")

	files, err := r.scanner.ScanFiles()
	if err != nil {
		return nil, err
	}

	r.logger.Infof("📁 Найдено файлов: %d", len(files))
	return files, nil
}

// checkFileChanges проверяет изменения файлов
func (r *App) checkFileChanges(files []string) ([]string, error) {
	r.logger.Info("🔍 Проверка изменений файлов...")

	var filesToProcess []string

	for _, file := range files {
		fullPath := filepath.Join(r.config.RootDir, file)

		// Получаем хеш файла
		fileHash, err := r.getFileHash(fullPath)
		if err != nil {
			r.logger.Warnf("⚠️ Не удалось получить хеш файла %s: %v", file, err)
			continue
		}

		// Проверяем хеш в базе данных
		storedHash, err := r.database.GetFileHash(file)
		if err != nil {
			r.logger.Warnf("⚠️ Не удалось получить сохранённый хеш для %s: %v", file, err)
		}

		// Если хеш изменился или файл новый
		if storedHash == "" || storedHash != fileHash {
			filesToProcess = append(filesToProcess, file)

			// Удаляем старые блоки для изменённого файла
			if storedHash != "" {
				if err := r.database.DeleteFileBlocks(file); err != nil {
					r.logger.Warnf("⚠️ Не удалось удалить старые блоки для %s: %v", file, err)
				}
			}

			// Обновляем хеш
			if err := r.database.UpdateFileHash(file, fileHash); err != nil {
				r.logger.Warnf("⚠️ Не удалось обновить хеш для %s: %v", file, err)
			}
		}
	}

	r.logger.Infof("📝 Файлов для обработки: %d", len(filesToProcess))
	for _, file := range filesToProcess {
		r.logger.Debugf("  - %s", file)
	}

	return filesToProcess, nil
}

// processFiles обрабатывает файлы и создаёт эмбединги
func (r *App) processFiles(files []string) error {
	r.logger.Info("🔄 Обработка файлов...")

	var allBlocks []*models.CodeBlock

	// Создаём прогресс-бар
	bar := progressbar.Default(int64(len(files)), "Обработка файлов")

	for _, file := range files {
		bar.Add(1)

		fullPath := filepath.Join(r.config.RootDir, file)
		ext := filepath.Ext(file)

		// Получаем парсер для файла
		parser, found := r.parsers.GetParser(ext)
		if !found {
			r.logger.Warnf("⚠️ Не найден парсер для файла %s", file)
			continue
		}

		// Парсим файл
		blocks, err := parser.ParseFile(fullPath)
		if err != nil {
			r.logger.Warnf("⚠️ Ошибка парсинга файла %s: %v", file, err)
			continue
		}

		// Получаем сообщения коммитов
		if r.gitService != nil {
			commitMessages, err := r.gitService.GetLastCommitMessages(fullPath, r.config.NCommits)
			if err != nil {
				r.logger.Debugf("Не удалось получить коммиты для %s: %v", file, err)
			} else {
				for _, block := range blocks {
					block.SetCommitMessages(commitMessages)
				}
			}
		}

		allBlocks = append(allBlocks, blocks...)
	}

	bar.Finish()
	r.logger.Infof("📦 Всего блоков для эмбединга: %d", len(allBlocks))

	// Создаём эмбединги
	return r.createEmbeddings(allBlocks)
}

// createEmbeddings создаёт эмбединги для блоков
func (r *App) createEmbeddings(blocks []*models.CodeBlock) error {
	r.logger.Info("🧠 Генерация эмбедингов...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Создаём прогресс-бар
	bar := progressbar.Default(int64(len(blocks)), "Генерация эмбедингов")

	for _, block := range blocks {
		bar.Add(1)

		// Проверяем, существует ли уже такой блок
		exists, err := r.database.BlockExists(block)
		if err != nil {
			r.logger.Warnf("⚠️ Ошибка проверки существования блока: %v", err)
			continue
		}
		if exists {
			continue // Пропускаем существующий блок
		}

		// Формируем текст для эмбединга
		embeddingText := block.GetEmbeddingText()

		// Получаем эмбединг
		embedding, err := r.openai.GetEmbedding(ctx, embeddingText)
		if err != nil {
			r.logger.Warnf("⚠️ Ошибка получения эмбединга для блока %s: %v", block, err)
			continue
		}

		// Сохраняем эмбединг
		if err := r.database.SaveEmbedding(block, embedding, embeddingText); err != nil {
			r.logger.Warnf("⚠️ Ошибка сохранения эмбединга для блока %s: %v", block, err)
			continue
		}
	}

	bar.Finish()
	return nil
}

// getFileHash получает MD5 хеш файла
func (r *App) getFileHash(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	hash := md5.Sum(data)
	return fmt.Sprintf("%x", hash), nil
}
