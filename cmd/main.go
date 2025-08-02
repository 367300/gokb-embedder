package main

import (
	"fmt"
	"log"
	"os"

	"gokb-embedder/internal/app"
	"gokb-embedder/internal/cli"
	"gokb-embedder/internal/config"
)

func main() {
	// Проверяем аргументы командной строки
	if len(os.Args) > 1 && os.Args[1] == "--quick" {
		// Быстрый режим без интерфейса - загружаем конфигурацию из .env
		cfg, err := config.Load()
		if err != nil {
			log.Printf("⚠️  Ошибка загрузки конфигурации: %v", err)
			log.Println("💡 Попробуйте запустить без флага --quick для интерактивной настройки:")
			log.Println("   ./gokb-embedder")
			log.Fatalf("Ошибка: %v", err)
		}

		// Создаём и запускаем приложение
		application := app.New(cfg)
		if err := application.Run(); err != nil {
			log.Fatalf("Ошибка выполнения приложения: %v", err)
		}
		return
	}

	// Интерактивный режим по умолчанию
	cliInterface := cli.NewCLI()

	// Создаём приложение один раз (пока без конфигурации)
	application := app.New(nil)

	// Запускаем интерактивный цикл
	for {
		cfg, err := cliInterface.Run()
		if err != nil {
			log.Fatalf("Ошибка CLI: %v", err)
		}

		// Обновляем конфигурацию приложения
		application.UpdateConfig(cfg)

		// Выполняем операцию в зависимости от режима
		switch cfg.OperationMode {
		case "statistics":
			// Для статистики нужно только инициализировать базу данных
			if err := application.InitializeDatabase(); err != nil {
				log.Printf("Ошибка инициализации базы данных: %v", err)
				continue
			}
			if err := application.ShowDatabaseStatistics(); err != nil {
				log.Printf("Ошибка показа статистики: %v", err)
			}
		case "preprocess":
			if err := application.RunPreprocess(); err != nil {
				log.Printf("Ошибка предварительной обработки: %v", err)
			}
		case "embeddings_only":
			// Для генерации эмбедингов нужно инициализировать базу данных и OpenAI
			if err := application.InitializeForEmbeddings(); err != nil {
				log.Printf("Ошибка инициализации: %v", err)
				continue
			}
			if err := application.GenerateEmbeddingsOnly(); err != nil {
				log.Printf("Ошибка генерации эмбедингов: %v", err)
			}
		case "full":
			if err := application.Run(); err != nil {
				log.Printf("Ошибка выполнения приложения: %v", err)
			}
		case "exit":
			// Выход из программы
			return
		default:
			if err := application.Run(); err != nil {
				log.Printf("Ошибка выполнения приложения: %v", err)
			}
		}

		// После выполнения операции возвращаемся в главное меню
		// (кроме случая выхода)
		if cfg.OperationMode != "exit" {
			fmt.Println()
			fmt.Println("Нажмите Enter для возврата в главное меню...")
			fmt.Scanln() // Ждём нажатия Enter
		}
	}
}
