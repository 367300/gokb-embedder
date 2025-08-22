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

	level := logrus.InfoLevel
	if cfg != nil {
		if parsedLevel, err := logrus.ParseLevel(cfg.LogLevel); err == nil {
			level = parsedLevel
		}
	}
	logger.SetLevel(level)

	return &App{
		config:  cfg,
		logger:  logger,
		parsers: parsers.NewParserRegistry(),
	}
}

// UpdateConfig обновляет конфигурацию приложения
func (r *App) UpdateConfig(cfg *config.Config) {
	r.config = cfg
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

// RunPreprocess запускает предварительную обработку файлов
func (r *App) RunPreprocess() error {
	r.logger.Info("📝 Запуск предварительной обработки файлов")

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

	// Обрабатываем файлы без эмбедингов
	if err := r.processFilesWithoutEmbeddings(filesToProcess); err != nil {
		return fmt.Errorf("ошибка предварительной обработки файлов: %w", err)
	}

	r.logger.Info("✅ Готово! Блоки сохранены в " + r.config.DBPath)
	return nil
}

// initialize инициализирует все компоненты приложения
func (r *App) initialize() error {
	r.logger.Info("🔧 Инициализация компонентов...")
	r.logger.Infof("📁 ROOT_DIR: %s", r.config.RootDir)
	r.logger.Infof("📝 FILE_EXTENSIONS: %v", r.config.FileExtensions)

	// Инициализируем базу данных
	r.logger.Debug("Инициализация базы данных...")
	db, err := database.NewDatabase(r.config.DBPath)
	if err != nil {
		return fmt.Errorf("ошибка инициализации базы данных: %w", err)
	}
	r.database = db
	r.logger.Debug("✅ База данных инициализирована")

	// Инициализируем OpenAI клиент
	r.logger.Debug("Инициализация OpenAI клиента...")
	r.openai = openai.NewClient(r.config.OpenAIAPIKey)
	r.logger.Debug("✅ OpenAI клиент инициализирован")

	// Инициализируем сканер
	r.logger.Debug("Инициализация сканера...")
	r.scanner = scanner.NewScanner(r.config.RootDir, r.config.FileExtensions)
	if r.scanner == nil {
		return fmt.Errorf("не удалось создать сканер")
	}
	r.logger.Debug("✅ Сканер создан")

	if err := r.scanner.LoadGitignore(); err != nil {
		r.logger.Warnf("⚠️ Не удалось загрузить .gitignore: %v", err)
	} else {
		r.logger.Debug("✅ .gitignore загружен")
	}

	// Регистрируем парсеры
	r.logger.Debug("Регистрация парсеров...")
	r.registerParsers()

	// Инициализируем Git сервис (если это Git репозиторий)
	r.logger.Debug("Проверка Git репозитория...")
	if git.IsGitRepository(r.config.RootDir) {
		r.logger.Debug("Git репозиторий найден, инициализация Git сервиса...")
		gitService, err := git.NewGitService(r.config.RootDir)
		if err != nil {
			r.logger.Warnf("⚠️ Не удалось инициализировать Git сервис: %v", err)
		} else {
			r.gitService = gitService
			r.logger.Debug("✅ Git сервис инициализирован")
		}
	} else {
		r.logger.Debug("Git репозиторий не найден, Git сервис не инициализирован")
	}

	r.logger.Info("✅ Все компоненты инициализированы")
	return nil
}

// InitializeDatabase инициализирует только базу данных
func (r *App) InitializeDatabase() error {
	r.logger.Info("🔧 Инициализация базы данных...")

	// Инициализируем базу данных
	r.logger.Debug("Инициализация базы данных...")
	db, err := database.NewDatabase(r.config.DBPath)
	if err != nil {
		return fmt.Errorf("ошибка инициализации базы данных: %w", err)
	}
	r.database = db
	r.logger.Debug("✅ База данных инициализирована")

	return nil
}

// InitializeForEmbeddings инициализирует базу данных и OpenAI клиент
func (r *App) InitializeForEmbeddings() error {
	r.logger.Info("🔧 Инициализация для генерации эмбедингов...")

	// Инициализируем базу данных
	r.logger.Debug("Инициализация базы данных...")
	db, err := database.NewDatabase(r.config.DBPath)
	if err != nil {
		return fmt.Errorf("ошибка инициализации базы данных: %w", err)
	}
	r.database = db
	r.logger.Debug("✅ База данных инициализирована")

	// Инициализируем OpenAI клиент
	r.logger.Debug("Инициализация OpenAI клиента...")
	r.openai = openai.NewClient(r.config.OpenAIAPIKey)
	r.logger.Debug("✅ OpenAI клиент инициализирован")

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
	r.logger.Debug("Начинаем регистрацию парсеров...")

	// Создаем карту выбранных расширений для быстрого поиска
	selectedExtensions := make(map[string]bool)
	for _, ext := range r.config.FileExtensions {
		selectedExtensions[ext] = true
	}

	// Регистрируем Python парсер (если выбраны .py файлы)
	if selectedExtensions[".py"] {
		r.logger.Debug("Регистрация Python парсера...")
		pythonParser := parsers.NewPythonParser()
		if pythonParser == nil {
			r.logger.Error("❌ Не удалось создать Python парсер")
		} else {
			r.parsers.Register(pythonParser)
			r.logger.Debugf("✅ Python парсер зарегистрирован: %s", pythonParser.GetName())
		}
	} else {
		r.logger.Debug("⏭️ Python парсер пропущен (файлы .py не выбраны)")
	}

	// Регистрируем JavaScript парсер (если выбраны JS/TS файлы)
	if selectedExtensions[".js"] {
		r.logger.Debug("Регистрация JavaScript парсера...")
		javascriptParser := parsers.NewJavaScriptParser()
		if javascriptParser == nil {
			r.logger.Error("❌ Не удалось создать JavaScript парсер")
		} else {
			r.parsers.Register(javascriptParser)
			r.logger.Debugf("✅ JavaScript парсер зарегистрирован: %s", javascriptParser.GetName())
		}
	} else {
		r.logger.Debug("⏭️ JavaScript парсер пропущен (JS файлы не выбраны)")
	}

	// Регистрируем PHP парсер (если выбраны .php файлы)
	if selectedExtensions[".php"] {
		r.logger.Debug("Регистрация PHP парсера...")
		phpParser := parsers.NewPHPParser()
		if phpParser == nil {
			r.logger.Error("❌ Не удалось создать PHP парсер")
		} else {
			r.parsers.Register(phpParser)
			r.logger.Debugf("✅ PHP парсер зарегистрирован: %s", phpParser.GetName())
		}
	} else {
		r.logger.Debug("⏭️ PHP парсер пропущен (файлы .php не выбраны)")
	}

	// Регистрируем текстовый парсер (если выбраны текстовые файлы)
	textExtensions := []string{".md", ".yml", ".yaml", ".conf", ".txt"}
	hasTextFiles := false
	for _, ext := range textExtensions {
		if selectedExtensions[ext] {
			hasTextFiles = true
			break
		}
	}

	if hasTextFiles {
		r.logger.Debug("Регистрация текстового парсера...")
		textParser := parsers.NewTextParser(r.config.TokenLimit)
		if textParser == nil {
			r.logger.Error("❌ Не удалось создать текстовый парсер")
		} else {
			r.parsers.Register(textParser)
			r.logger.Debugf("✅ Текстовый парсер зарегистрирован: %s", textParser.GetName())
		}
	} else {
		r.logger.Debug("⏭️ Текстовый парсер пропущен (текстовые файлы не выбраны)")
	}

	allParsers := r.parsers.GetAllParsers()
	r.logger.Infof("📝 Зарегистрировано парсеров: %d", len(allParsers))

	for _, parser := range allParsers {
		r.logger.Debugf("  - %s", parser.GetName())
	}
}

// scanFiles сканирует файлы в проекте
func (r *App) scanFiles() ([]string, error) {
	r.logger.Info("🔍 Сканирование файлов...")

	if r.scanner == nil {
		return nil, fmt.Errorf("сканер не инициализирован")
	}

	r.logger.Debugf("Сканирование в директории: %s", r.config.RootDir)
	r.logger.Debugf("Ищем файлы с расширениями: %v", r.config.FileExtensions)

	files, err := r.scanner.ScanFiles()
	if err != nil {
		r.logger.Errorf("Ошибка сканирования файлов: %v", err)
		return nil, err
	}

	r.logger.Infof("📁 Найдено файлов: %d", len(files))
	if len(files) == 0 {
		r.logger.Warn("⚠️ Файлы не найдены! Проверьте:")
		r.logger.Warn("  - Правильность пути ROOT_DIR")
		r.logger.Warn("  - Наличие файлов с указанными расширениями")
		r.logger.Warn("  - Правила .gitignore")
	} else {
		r.logger.Debug("Найденные файлы:")
		for i, file := range files {
			if i < 10 { // Показываем только первые 10 файлов
				r.logger.Debugf("  - %s", file)
			} else if i == 10 {
				r.logger.Debugf("  ... и ещё %d файлов", len(files)-10)
				break
			}
		}
	}

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

		// Устанавливаем относительные пути и получаем сообщения коммитов
		for _, block := range blocks {
			// Устанавливаем относительный путь от корня проекта
			block.SetRelativePath(file)

			// Получаем сообщения коммитов
			if r.gitService != nil {
				commitMessages, err := r.gitService.GetLastCommitMessages(fullPath, r.config.NCommits)
				if err != nil {
					r.logger.Debugf("Не удалось получить коммиты для %s: %v", file, err)
				} else {
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

// processFilesWithoutEmbeddings обрабатывает файлы без создания эмбедингов
func (r *App) processFilesWithoutEmbeddings(files []string) error {
	r.logger.Info("📝 Предварительная обработка файлов (без эмбедингов)...")

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

		// Устанавливаем относительные пути и получаем сообщения коммитов
		for _, block := range blocks {
			// Устанавливаем относительный путь от корня проекта
			block.SetRelativePath(file)

			// Получаем сообщения коммитов
			if r.gitService != nil {
				commitMessages, err := r.gitService.GetLastCommitMessages(fullPath, r.config.NCommits)
				if err != nil {
					r.logger.Debugf("Не удалось получить коммиты для %s: %v", file, err)
				} else {
					block.SetCommitMessages(commitMessages)
				}
			}
		}

		allBlocks = append(allBlocks, blocks...)
	}

	bar.Finish()
	r.logger.Infof("📦 Всего блоков обработано: %d", len(allBlocks))

	// Сохраняем блоки без эмбедингов
	return r.saveBlocksWithoutEmbeddings(allBlocks)
}

// saveBlocksWithoutEmbeddings сохраняет блоки в базу данных без эмбедингов
func (r *App) saveBlocksWithoutEmbeddings(blocks []*models.CodeBlock) error {
	r.logger.Info("💾 Сохранение блоков в базу данных...")

	// Создаём прогресс-бар
	bar := progressbar.Default(int64(len(blocks)), "Сохранение блоков")

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

		// Сохраняем блок без эмбединга
		if err := r.database.SaveBlockWithoutEmbedding(block, embeddingText); err != nil {
			r.logger.Warnf("⚠️ Ошибка сохранения блока %s: %v", block, err)
			continue
		}
	}

	bar.Finish()
	return nil
}

// GenerateEmbeddingsOnly генерирует эмбединги только для блоков без эмбедингов
func (r *App) GenerateEmbeddingsOnly() error {
	r.logger.Info("🧠 Генерация эмбедингов для существующих блоков...")

	// Получаем блоки без эмбедингов
	blocks, err := r.database.GetBlocksWithoutEmbeddings()
	if err != nil {
		return fmt.Errorf("ошибка получения блоков без эмбедингов: %w", err)
	}

	if len(blocks) == 0 {
		r.logger.Info("✅ Все блоки уже имеют эмбединги!")
		return nil
	}

	r.logger.Infof("📦 Найдено блоков без эмбедингов: %d", len(blocks))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Создаём прогресс-бар
	bar := progressbar.Default(int64(len(blocks)), "Генерация эмбедингов")

	for _, block := range blocks {
		bar.Add(1)

		// Формируем текст для эмбединга
		embeddingText := block.GetEmbeddingText()

		// Получаем эмбединг
		embedding, err := r.openai.GetEmbedding(ctx, embeddingText)
		if err != nil {
			r.logger.Warnf("⚠️ Ошибка получения эмбединга для блока %s: %v", block, err)
			continue
		}

		// Обновляем эмбединг
		if err := r.database.UpdateEmbedding(block, embedding); err != nil {
			r.logger.Warnf("⚠️ Ошибка обновления эмбединга для блока %s: %v", block, err)
			continue
		}
	}

	bar.Finish()
	return nil
}

// ShowDatabaseStatistics показывает статистику базы данных
func (r *App) ShowDatabaseStatistics() error {
	r.logger.Info("📊 Статистика базы данных...")

	stats, err := r.database.GetStatistics()
	if err != nil {
		return fmt.Errorf("ошибка получения статистики: %w", err)
	}

	r.logger.Infof("📁 Всего файлов: %d", stats["file_count"])
	r.logger.Infof("📦 Всего блоков: %d", stats["total_blocks"])
	r.logger.Infof("✅ Блоков с эмбедингами: %d", stats["blocks_with_embeddings"])
	r.logger.Infof("⏳ Блоков без эмбедингов: %d", stats["blocks_without_embeddings"])

	// Показываем статистику по типам блоков
	if blockTypes, ok := stats["block_types"].(map[string]int); ok {
		r.logger.Info("📝 Статистика по типам блоков:")
		for blockType, count := range blockTypes {
			r.logger.Infof("   • %s: %d", blockType, count)
		}
	}

	return nil
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
