package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gokb-embedder/internal/app"
	"gokb-embedder/internal/config"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

// CLI представляет интерактивный интерфейс командной строки
type CLI struct {
	config *config.Config
}

// NewCLI создаёт новый CLI интерфейс
func NewCLI() *CLI {
	return &CLI{}
}

// Run запускает интерактивный CLI
func (c *CLI) Run() (*config.Config, error) {
	color.Cyan("🚀 GoKB Embedder - Интерактивная настройка")
	fmt.Println()

	// Показываем информацию о быстром режиме
	color.Yellow("💡 Для быстрого запуска без интерфейса используйте флаг --quick")
	color.Yellow("   ./gokb-embedder --quick")
	fmt.Println()

	// Проверяем существующий .env файл
	envExists := c.checkEnvFile()

	if envExists {
		// Загружаем существующую конфигурацию
		cfg, err := config.Load()
		if err != nil {
			color.Yellow("⚠️  Ошибка загрузки .env файла: %v", err)
			fmt.Println()
		} else {
			c.config = cfg
			color.Green("✅ Конфигурация загружена из .env файла")
			fmt.Println()
		}
	} else {
		color.Yellow("⚠️  Файл .env не найден")
		fmt.Println()
	}

	// Показываем главное меню
	return c.showMainMenu()
}

// checkEnvFile проверяет существование .env файла
func (c *CLI) checkEnvFile() bool {
	if _, err := os.Stat(".env"); err == nil {
		return true
	}
	return false
}

// quickStart выполняет быструю настройку с предустановленными значениями
func (c *CLI) quickStart() error {
	color.Cyan("🚀 Быстрый старт")
	fmt.Println()
	color.Yellow("Настраиваем GoKB Embedder с рекомендуемыми параметрами...")
	fmt.Println()

	if c.config == nil {
		c.config = &config.Config{}
	}

	// Запрашиваем только обязательные параметры
	if c.config.OpenAIAPIKey == "" {
		color.Yellow("🔑 OpenAI API Key (обязательно)")
		prompt := promptui.Prompt{
			Label: "Введите ваш OpenAI API Key",
			Mask:  '*',
		}
		apiKey, err := prompt.Run()
		if err != nil {
			return err
		}
		c.config.OpenAIAPIKey = apiKey
	}

	// Устанавливаем рекомендуемые значения
	c.config.RootDir = "."
	c.config.DBPath = "embeddings.sqlite3"
	c.config.NCommits = 3
	c.config.TokenLimit = 1600
	c.config.LogLevel = "info"
	c.config.FileExtensions = []string{".py", ".js", ".php", ".md", ".yml", ".conf"}

	color.Green("✅ Рекомендуемые настройки:")
	fmt.Printf("   📁 Root Directory: %s\n", c.config.RootDir)
	fmt.Printf("   💾 Database Path: %s\n", c.config.DBPath)
	fmt.Printf("   📚 Number of Commits: %d\n", c.config.NCommits)
	fmt.Printf("   🔢 Token Limit: %d\n", c.config.TokenLimit)
	fmt.Printf("   📊 Log Level: %s\n", c.config.LogLevel)
	fmt.Printf("   📝 File Extensions: %s\n", strings.Join(c.config.FileExtensions, ", "))
	fmt.Println()

	// Сохраняем в .env файл
	if err := c.saveToEnv(); err != nil {
		return err
	}

	color.Green("✅ Быстрая настройка завершена!")
	color.Cyan("💡 Теперь вы можете:")
	color.Cyan("   • Запустить генерацию эмбедингов")
	color.Cyan("   • Настроить дополнительные параметры")
	color.Cyan("   • Изменить парсеры")
	fmt.Println()
	return nil
}

// showMainMenu показывает главное меню
func (c *CLI) showMainMenu() (*config.Config, error) {
	for {
		prompt := promptui.Select{
			Label: "Выберите действие",
			Items: []string{
				"🚀 Быстрый старт (рекомендуется)",
				"🔧 Настроить конфигурацию",
				"📝 Настроить парсеры",
				"🔍 Проверить настройки",
				"📊 Статистика базы данных",
				"📤 Экспорт базы данных в CSV",
				"📝 Предварительная обработка файлов",
				"🧠 Генерация эмбедингов",
				"▶️  Полная обработка (файлы + эмбединги)",
				"❌ Выход",
			},
		}

		_, result, err := prompt.Run()
		if err != nil {
			return nil, err
		}

		switch result {
		case "🚀 Быстрый старт (рекомендуется)":
			if err := c.quickStart(); err != nil {
				color.Red("❌ Ошибка быстрого старта: %v", err)
			}
		case "🔧 Настроить конфигурацию":
			if err := c.configureSettings(); err != nil {
				color.Red("❌ Ошибка настройки: %v", err)
			}
		case "📝 Настроить парсеры":
			if err := c.configureParsers(); err != nil {
				color.Red("❌ Ошибка настройки парсеров: %v", err)
			}
		case "🔍 Проверить настройки":
			c.showCurrentConfig()
		case "📊 Статистика базы данных":
			if c.config == nil {
				color.Red("❌ Сначала настройте конфигурацию!")
				continue
			}
			c.config.OperationMode = "statistics"
			return c.config, nil
		case "📤 Экспорт базы данных в CSV":
			if c.config == nil {
				color.Red("❌ Сначала настройте конфигурацию!")
				continue
			}
			if err := c.exportToCSV(); err != nil {
				color.Red("❌ Ошибка экспорта: %v", err)
			}
		case "📝 Предварительная обработка файлов":
			if c.config == nil {
				color.Red("❌ Сначала настройте конфигурацию!")
				continue
			}
			c.config.OperationMode = "preprocess"
			return c.config, nil
		case "🧠 Генерация эмбедингов":
			if c.config == nil {
				color.Red("❌ Сначала настройте конфигурацию!")
				continue
			}
			c.config.OperationMode = "embeddings_only"
			return c.config, nil
		case "▶️  Полная обработка (файлы + эмбединги)":
			if c.config == nil {
				color.Red("❌ Сначала настройте конфигурацию!")
				continue
			}
			c.config.OperationMode = "full"
			return c.config, nil
		case "❌ Выход":
			color.Yellow("👋 До свидания!")
			c.config.OperationMode = "exit"
			return c.config, nil
		}
	}
}

// configureSettings настраивает основные параметры
func (c *CLI) configureSettings() error {
	color.Cyan("🔧 Настройка конфигурации")
	fmt.Println()

	if c.config == nil {
		c.config = &config.Config{}
	}

	// OpenAI API Key
	color.Yellow("🔑 OpenAI API Key (обязательно)")
	if c.config.OpenAIAPIKey == "" {
		prompt := promptui.Prompt{
			Label: "Введите ваш OpenAI API Key",
			Mask:  '*',
		}
		apiKey, err := prompt.Run()
		if err != nil {
			return err
		}
		c.config.OpenAIAPIKey = apiKey
	} else {
		color.Green("✅ OpenAI API Key уже настроен")
	}

	// Root Directory
	color.Yellow("📁 Корневая директория")
	prompt := promptui.Prompt{
		Label:   "Путь к корневой директории проекта",
		Default: c.config.RootDir,
	}
	rootDir, err := prompt.Run()
	if err != nil {
		return err
	}
	c.config.RootDir = rootDir

	// Database Path
	color.Yellow("💾 База данных")
	prompt = promptui.Prompt{
		Label:   "Путь к файлу базы данных",
		Default: c.config.DBPath,
	}
	dbPath, err := prompt.Run()
	if err != nil {
		return err
	}
	c.config.DBPath = dbPath

	// Number of Commits
	color.Yellow("📚 История коммитов")
	prompt = promptui.Prompt{
		Label:   "Количество последних коммитов для получения истории",
		Default: strconv.Itoa(c.config.NCommits),
	}
	nCommitsStr, err := prompt.Run()
	if err != nil {
		return err
	}
	if nCommits, err := strconv.Atoi(nCommitsStr); err == nil {
		c.config.NCommits = nCommits
	}

	// Token Limit
	color.Yellow("🔢 Лимит токенов")
	prompt = promptui.Prompt{
		Label:   "Лимит токенов на блок для текстовых файлов",
		Default: strconv.Itoa(c.config.TokenLimit),
	}
	tokenLimitStr, err := prompt.Run()
	if err != nil {
		return err
	}
	if tokenLimit, err := strconv.Atoi(tokenLimitStr); err == nil {
		c.config.TokenLimit = tokenLimit
	}

	// Log Level
	color.Yellow("📊 Уровень логирования")
	logPrompt := promptui.Select{
		Label: "Выберите уровень логирования",
		Items: []string{"debug", "info", "warn", "error"},
	}
	_, logLevel, err := logPrompt.Run()
	if err != nil {
		return err
	}
	c.config.LogLevel = logLevel

	// Сохраняем в .env файл
	if err := c.saveToEnv(); err != nil {
		return err
	}

	color.Green("✅ Конфигурация сохранена!")
	fmt.Println()
	return nil
}

// configureParsers настраивает парсеры
func (c *CLI) configureParsers() error {
	color.Cyan("📝 Настройка парсеров")
	fmt.Println()

	// Доступные парсеры с подробным описанием
	availableParsers := map[string]map[string]string{
		".py": {
			"name":        "Python Parser",
			"description": "Извлекает методы, функции и классы из Python файлов",
			"features":    "• Автоматическое определение классов и методов\n• Извлечение документации (docstrings)\n• Поддержка вложенных функций",
			"parser":      "python",
		},
		".js": {
			"name":        "JavaScript Parser",
			"description": "Извлекает функции, методы и классы из JavaScript файлов",
			"features":    "• Поддержка ES6+ синтаксиса\n• Извлечение стрелочных функций\n• Обработка React компонентов\n• Поддержка TypeScript",
			"parser":      "javascript",
		},
		".php": {
			"name":        "PHP Parser",
			"description": "Извлекает функции, методы и классы из PHP файлов",
			"features":    "• Поддержка ООП (классы, интерфейсы, трейты)\n• Извлечение namespace\n• Обработка абстрактных классов\n• Поддержка анонимных функций",
			"parser":      "php",
		},
		".md": {
			"name":        "Markdown Parser",
			"description": "Обрабатывает документацию и README файлы",
			"features":    "• Разбивка на логические блоки\n• Сохранение структуры заголовков\n• Обработка кодовых блоков",
			"parser":      "text",
		},
		".yml": {
			"name":        "YAML Parser",
			"description": "Обрабатывает конфигурационные файлы YAML",
			"features":    "• Разбивка по секциям конфигурации\n• Сохранение иерархии ключей\n• Обработка комментариев",
			"parser":      "text",
		},
		".yaml": {
			"name":        "YAML Parser",
			"description": "Обрабатывает конфигурационные файлы YAML",
			"features":    "• Разбивка по секциям конфигурации\n• Сохранение иерархии ключей\n• Обработка комментариев",
			"parser":      "text",
		},
		".conf": {
			"name":        "Config Parser",
			"description": "Обрабатывает конфигурационные файлы",
			"features":    "• Разбивка по секциям\n• Обработка комментариев\n• Сохранение структуры",
			"parser":      "text",
		},
		".txt": {
			"name":        "Text Parser",
			"description": "Обрабатывает простые текстовые файлы",
			"features":    "• Разбивка по токенам\n• Настраиваемый лимит токенов\n• Сохранение контекста",
			"parser":      "text",
		},
	}

	// Текущие расширения
	currentExtensions := make(map[string]bool)
	for _, ext := range c.config.FileExtensions {
		currentExtensions[ext] = true
	}

	color.Yellow("Выберите расширения файлов для обработки:")
	fmt.Println()

	var selectedExtensions []string
	for ext, info := range availableParsers {
		isSelected := currentExtensions[ext]
		status := "❌"
		if isSelected {
			status = "✅"
		}

		// Показываем подробную информацию о парсере
		color.Cyan("📋 %s %s", status, ext)
		fmt.Printf("   Название: %s\n", info["name"])
		fmt.Printf("   Описание: %s\n", info["description"])
		fmt.Printf("   Возможности:\n%s\n", info["features"])
		fmt.Println()

		prompt := promptui.Select{
			Label: fmt.Sprintf("Статус для %s (%s):", ext, info["name"]),
			Items: []string{"✅ Включить", "❌ Отключить"},
		}

		_, result, err := prompt.Run()
		if err != nil {
			return err
		}

		if strings.Contains(result, "Включить") {
			selectedExtensions = append(selectedExtensions, ext)
		}
	}

	c.config.FileExtensions = selectedExtensions

	// Показываем статистику
	color.Green("📊 Статистика выбранных парсеров:")
	pythonCount := 0
	javascriptCount := 0
	phpCount := 0
	textCount := 0
	for _, ext := range selectedExtensions {
		if availableParsers[ext]["parser"] == "python" {
			pythonCount++
		} else if availableParsers[ext]["parser"] == "javascript" {
			javascriptCount++
		} else if availableParsers[ext]["parser"] == "php" {
			phpCount++
		} else if availableParsers[ext]["parser"] == "text" {
			textCount++
		}
	}
	fmt.Printf("   🐍 Python парсер: %d расширений\n", pythonCount)
	fmt.Printf("   🟨 JavaScript парсер: %d расширений\n", javascriptCount)
	fmt.Printf("   🟦 PHP парсер: %d расширений\n", phpCount)
	fmt.Printf("   📝 Text парсер: %d расширений\n", textCount)
	fmt.Printf("   📁 Всего расширений: %d\n", len(selectedExtensions))
	fmt.Println()

	// Сохраняем в .env файл
	if err := c.saveToEnv(); err != nil {
		return err
	}

	color.Green("✅ Настройки парсеров сохранены!")
	fmt.Println()
	return nil
}

// showCurrentConfig показывает текущую конфигурацию
func (c *CLI) showCurrentConfig() {
	color.Cyan("🔍 Текущая конфигурация")
	fmt.Println()

	if c.config == nil {
		color.Red("❌ Конфигурация не загружена")
		fmt.Println()
		return
	}

	// Проверяем валидность конфигурации
	issues := c.validateConfig()
	if len(issues) > 0 {
		color.Yellow("⚠️  Обнаружены проблемы в конфигурации:")
		for _, issue := range issues {
			color.Yellow("   • %s", issue)
		}
		fmt.Println()
	} else {
		color.Green("✅ Конфигурация валидна")
		fmt.Println()
	}

	fmt.Printf("🔑 OpenAI API Key: %s\n", maskAPIKey(c.config.OpenAIAPIKey))
	fmt.Printf("📁 Root Directory: %s\n", c.config.RootDir)
	fmt.Printf("💾 Database Path: %s\n", c.config.DBPath)
	fmt.Printf("📚 Number of Commits: %d\n", c.config.NCommits)
	fmt.Printf("🔢 Token Limit: %d\n", c.config.TokenLimit)
	fmt.Printf("📊 Log Level: %s\n", c.config.LogLevel)
	fmt.Printf("📝 File Extensions: %s\n", strings.Join(c.config.FileExtensions, ", "))

	// Показываем статистику парсеров
	if len(c.config.FileExtensions) > 0 {
		fmt.Println()
		color.Cyan("📊 Статистика парсеров:")
		pythonCount := 0
		javascriptCount := 0
		phpCount := 0
		textCount := 0
		for _, ext := range c.config.FileExtensions {
			if ext == ".py" {
				pythonCount++
			} else if ext == ".js" {
				javascriptCount++
			} else if ext == ".php" {
				phpCount++
			} else {
				textCount++
			}
		}
		fmt.Printf("   🐍 Python парсер: %d расширений\n", pythonCount)
		fmt.Printf("   🟨 JavaScript парсер: %d расширений\n", javascriptCount)
		fmt.Printf("   🟦 PHP парсер: %d расширений\n", phpCount)
		fmt.Printf("   📝 Text парсер: %d расширений\n", textCount)
	}
	fmt.Println()
}

// validateConfig проверяет валидность конфигурации
func (c *CLI) validateConfig() []string {
	var issues []string

	// Проверяем OpenAI API Key
	if c.config.OpenAIAPIKey == "" {
		issues = append(issues, "Отсутствует OpenAI API Key")
	} else if len(c.config.OpenAIAPIKey) < 20 {
		issues = append(issues, "OpenAI API Key слишком короткий")
	}

	// Проверяем корневую директорию
	if c.config.RootDir == "" {
		issues = append(issues, "Не указана корневая директория")
	}

	// Проверяем расширения файлов
	if len(c.config.FileExtensions) == 0 {
		issues = append(issues, "Не выбраны расширения файлов для обработки")
	}

	// Проверяем количество коммитов
	if c.config.NCommits < 0 {
		issues = append(issues, "Количество коммитов не может быть отрицательным")
	}

	// Проверяем лимит токенов
	if c.config.TokenLimit < 100 {
		issues = append(issues, "Лимит токенов слишком мал (рекомендуется >= 100)")
	} else if c.config.TokenLimit > 8000 {
		issues = append(issues, "Лимит токенов слишком велик (рекомендуется <= 8000)")
	}

	return issues
}

// saveToEnv сохраняет конфигурацию в .env файл
func (c *CLI) saveToEnv() error {
	file, err := os.Create(".env")
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Записываем конфигурацию
	fmt.Fprintf(writer, "# OpenAI API ключ (обязательно)\n")
	fmt.Fprintf(writer, "OPENAI_API_KEY=%s\n\n", c.config.OpenAIAPIKey)

	fmt.Fprintf(writer, "# Корневая директория для поиска файлов\n")
	fmt.Fprintf(writer, "ROOT_DIR=%s\n\n", c.config.RootDir)

	fmt.Fprintf(writer, "# Расширения файлов для обработки (через запятую)\n")
	fmt.Fprintf(writer, "FILE_EXTENSIONS=%s\n\n", strings.Join(c.config.FileExtensions, ","))

	fmt.Fprintf(writer, "# Путь к файлу базы данных\n")
	fmt.Fprintf(writer, "DB_PATH=%s\n\n", c.config.DBPath)

	fmt.Fprintf(writer, "# Количество последних коммитов для получения истории\n")
	fmt.Fprintf(writer, "N_COMMITS=%d\n\n", c.config.NCommits)

	fmt.Fprintf(writer, "# Лимит токенов на блок для текстовых файлов\n")
	fmt.Fprintf(writer, "TOKEN_LIMIT=%d\n\n", c.config.TokenLimit)

	fmt.Fprintf(writer, "# Уровень логирования (debug, info, warn, error)\n")
	fmt.Fprintf(writer, "LOG_LEVEL=%s\n", c.config.LogLevel)

	return writer.Flush()
}

// maskAPIKey маскирует API ключ для безопасного отображения
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "***"
	}
	return apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
}

// exportToCSV экспортирует базу данных в CSV файл
func (c *CLI) exportToCSV() error {
	color.Cyan("📤 Экспорт базы данных в CSV")
	fmt.Println()

	// Проверяем существование базы данных
	if _, err := os.Stat(c.config.DBPath); os.IsNotExist(err) {
		color.Red("❌ База данных не найдена: %s", c.config.DBPath)
		color.Yellow("💡 Сначала создайте базу данных, запустив обработку файлов")
		fmt.Println()
		return nil
	}

	// Запрашиваем путь для сохранения CSV файла
	color.Yellow("📁 Путь для сохранения CSV файла")
	prompt := promptui.Prompt{
		Label:   "Введите путь к файлу (например: export.csv)",
		Default: "embeddings_export.csv",
	}
	outputPath, err := prompt.Run()
	if err != nil {
		return err
	}

	// Проверяем, существует ли файл
	if _, err := os.Stat(outputPath); err == nil {
		color.Yellow("⚠️  Файл уже существует: %s", outputPath)
		confirmPrompt := promptui.Select{
			Label: "Перезаписать файл?",
			Items: []string{"✅ Да, перезаписать", "❌ Нет, отменить"},
		}
		_, result, err := confirmPrompt.Run()
		if err != nil {
			return err
		}
		if strings.Contains(result, "Нет") {
			color.Yellow("📤 Экспорт отменён")
			fmt.Println()
			return nil
		}
	}

	// Создаём приложение для экспорта
	app := app.New(c.config)
	if err := app.InitializeDatabase(); err != nil {
		return fmt.Errorf("ошибка инициализации базы данных: %w", err)
	}
	// Примечание: cleanup() вызывается автоматически при завершении работы приложения

	// Выполняем экспорт
	color.Yellow("📤 Выполняется экспорт...")
	if err := app.ExportDatabaseToCSV(outputPath); err != nil {
		return err
	}

	// Показываем статистику экспортированного файла
	if fileInfo, err := os.Stat(outputPath); err == nil {
		color.Green("✅ Экспорт завершён успешно!")
		fmt.Printf("📁 Файл: %s\n", outputPath)
		fmt.Printf("📊 Размер: %.2f МБ\n", float64(fileInfo.Size())/1024/1024)
		fmt.Println()
		color.Cyan("💡 Теперь вы можете:")
		color.Cyan("   • Открыть файл в Excel или Google Sheets")
		color.Cyan("   • Импортировать данные в другие системы")
		color.Cyan("   • Анализировать структуру кодовой базы")
		fmt.Println()
	}

	return nil
}
