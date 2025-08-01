package main

import (
	"log"

	"gokb-embedder/internal/app"
	"gokb-embedder/internal/config"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Создаём и запускаем приложение
	application := app.New(cfg)
	if err := application.Run(); err != nil {
		log.Fatalf("Ошибка выполнения приложения: %v", err)
	}
}
