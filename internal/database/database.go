package database

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"gokb-embedder/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

// Database предоставляет методы для работы с базой данных
type Database struct {
	db *sql.DB
}

// NewDatabase создаёт новое подключение к базе данных
func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия базы данных: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	database := &Database{db: db}
	if err := database.initTables(); err != nil {
		return nil, fmt.Errorf("ошибка инициализации таблиц: %w", err)
	}

	return database, nil
}

// Close закрывает подключение к базе данных
func (d *Database) Close() error {
	return d.db.Close()
}

// initTables создаёт необходимые таблицы
func (d *Database) initTables() error {
	// Таблица для эмбедингов
	embeddingsTable := `
	CREATE TABLE IF NOT EXISTS embeddings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		embedding TEXT NOT NULL,
		file_path TEXT NOT NULL,
		relative_path TEXT NOT NULL,
		block_type TEXT NOT NULL,
		class_name TEXT,
		method_name TEXT,
		start_line INTEGER NOT NULL,
		end_line INTEGER NOT NULL,
		commit_messages TEXT,
		raw_text TEXT NOT NULL,
		embedding_text TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

	// Таблица для хешей файлов
	fileHashesTable := `
	CREATE TABLE IF NOT EXISTS file_hashes (
		file_path TEXT PRIMARY KEY,
		file_hash TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

	// Создаём таблицы
	if _, err := d.db.Exec(embeddingsTable); err != nil {
		return fmt.Errorf("ошибка создания таблицы embeddings: %w", err)
	}

	if _, err := d.db.Exec(fileHashesTable); err != nil {
		return fmt.Errorf("ошибка создания таблицы file_hashes: %w", err)
	}

	return nil
}

// SaveEmbedding сохраняет эмбединг в базу данных
func (d *Database) SaveEmbedding(block *models.CodeBlock, embedding []float64, embeddingText string) error {
	// Сериализуем эмбединг в JSON
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("ошибка сериализации эмбединга: %w", err)
	}

	// Сериализуем сообщения коммитов в JSON
	var commitMessagesJSON *string
	if len(block.CommitMessages) > 0 {
		commitJSON, err := json.Marshal(block.CommitMessages)
		if err != nil {
			return fmt.Errorf("ошибка сериализации сообщений коммитов: %w", err)
		}
		commitStr := string(commitJSON)
		commitMessagesJSON = &commitStr
	}

	// Подготавливаем значения для вставки
	className := ""
	if block.ClassName != nil {
		className = *block.ClassName
	}

	methodName := ""
	if block.MethodName != nil {
		methodName = *block.MethodName
	}

	// Вставляем запись
	query := `
	INSERT INTO embeddings 
	(embedding, file_path, relative_path, block_type, class_name, method_name, 
	 start_line, end_line, commit_messages, raw_text, embedding_text)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = d.db.Exec(query,
		string(embeddingJSON),
		block.FilePath,
		block.GetRelativePath(),
		block.BlockType,
		className,
		methodName,
		block.StartLine,
		block.EndLine,
		commitMessagesJSON,
		block.RawText,
		embeddingText,
	)

	if err != nil {
		return fmt.Errorf("ошибка вставки эмбединга: %w", err)
	}

	return nil
}

// GetFileHash возвращает хеш файла из базы данных
func (d *Database) GetFileHash(filePath string) (string, error) {
	var hash string
	err := d.db.QueryRow("SELECT file_hash FROM file_hashes WHERE file_path = ?", filePath).Scan(&hash)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("ошибка получения хеша файла: %w", err)
	}
	return hash, nil
}

// UpdateFileHash обновляет хеш файла в базе данных
func (d *Database) UpdateFileHash(filePath, hash string) error {
	query := `
	INSERT OR REPLACE INTO file_hashes (file_path, file_hash, updated_at)
	VALUES (?, ?, CURRENT_TIMESTAMP)`

	_, err := d.db.Exec(query, filePath, hash)
	if err != nil {
		return fmt.Errorf("ошибка обновления хеша файла: %w", err)
	}

	return nil
}

// DeleteFileBlocks удаляет все блоки для файла
func (d *Database) DeleteFileBlocks(filePath string) error {
	_, err := d.db.Exec("DELETE FROM embeddings WHERE file_path = ?", filePath)
	if err != nil {
		return fmt.Errorf("ошибка удаления блоков файла: %w", err)
	}
	return nil
}

// BlockExists проверяет, существует ли блок с такими параметрами
func (d *Database) BlockExists(block *models.CodeBlock) (bool, error) {
	className := ""
	if block.ClassName != nil {
		className = *block.ClassName
	}

	methodName := ""
	if block.MethodName != nil {
		methodName = *block.MethodName
	}

	var count int
	err := d.db.QueryRow(`
		SELECT COUNT(*) FROM embeddings
		WHERE file_path = ? AND class_name = ? AND method_name = ? 
		AND start_line = ? AND end_line = ? AND block_type = ?`,
		block.FilePath, className, methodName, block.StartLine, block.EndLine, block.BlockType,
	).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("ошибка проверки существования блока: %w", err)
	}

	return count > 0, nil
}

// GetAllFilePaths возвращает все пути файлов из базы данных
func (d *Database) GetAllFilePaths() ([]string, error) {
	rows, err := d.db.Query("SELECT DISTINCT file_path FROM embeddings")
	if err != nil {
		return nil, fmt.Errorf("ошибка получения путей файлов: %w", err)
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, fmt.Errorf("ошибка сканирования пути файла: %w", err)
		}
		paths = append(paths, path)
	}

	return paths, nil
}

// SaveBlockWithoutEmbedding сохраняет блок кода без эмбединга (только embedding_text)
func (d *Database) SaveBlockWithoutEmbedding(block *models.CodeBlock, embeddingText string) error {
	// Сериализуем сообщения коммитов в JSON
	var commitMessagesJSON *string
	if len(block.CommitMessages) > 0 {
		commitJSON, err := json.Marshal(block.CommitMessages)
		if err != nil {
			return fmt.Errorf("ошибка сериализации сообщений коммитов: %w", err)
		}
		commitStr := string(commitJSON)
		commitMessagesJSON = &commitStr
	}

	// Подготавливаем значения для вставки
	className := ""
	if block.ClassName != nil {
		className = *block.ClassName
	}

	methodName := ""
	if block.MethodName != nil {
		methodName = *block.MethodName
	}

	// Вставляем запись с пустым embedding
	query := `
	INSERT INTO embeddings 
	(embedding, file_path, relative_path, block_type, class_name, method_name, 
	 start_line, end_line, commit_messages, raw_text, embedding_text)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := d.db.Exec(query,
		"", // пустой embedding
		block.FilePath,
		block.GetRelativePath(),
		block.BlockType,
		className,
		methodName,
		block.StartLine,
		block.EndLine,
		commitMessagesJSON,
		block.RawText,
		embeddingText,
	)

	if err != nil {
		return fmt.Errorf("ошибка вставки блока: %w", err)
	}

	return nil
}

// GetBlocksWithoutEmbeddings возвращает все блоки без эмбедингов
func (d *Database) GetBlocksWithoutEmbeddings() ([]*models.CodeBlock, error) {
	rows, err := d.db.Query(`
		SELECT id, file_path, relative_path, block_type, class_name, method_name, 
		       start_line, end_line, commit_messages, raw_text, embedding_text
		FROM embeddings 
		WHERE embedding = '' OR embedding IS NULL
		ORDER BY file_path, start_line`)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения блоков без эмбедингов: %w", err)
	}
	defer rows.Close()

	var blocks []*models.CodeBlock
	for rows.Next() {
		var id, startLine, endLine int
		var filePath, relativePath, blockType, className, methodName, commitMessages, rawText, embeddingText string

		err := rows.Scan(&id, &filePath, &relativePath, &blockType, &className, &methodName,
			&startLine, &endLine, &commitMessages, &rawText, &embeddingText)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования блока: %w", err)
		}

		// Десериализуем сообщения коммитов
		var commitMsgs []string
		if commitMessages != "" {
			if err := json.Unmarshal([]byte(commitMessages), &commitMsgs); err != nil {
				return nil, fmt.Errorf("ошибка десериализации сообщений коммитов: %w", err)
			}
		}

		// Создаём блок
		block := &models.CodeBlock{
			FilePath:       filePath,
			RelativePath:   relativePath,
			BlockType:      blockType,
			StartLine:      startLine,
			EndLine:        endLine,
			RawText:        rawText,
			CommitMessages: commitMsgs,
		}

		// Устанавливаем опциональные поля
		if className != "" {
			block.ClassName = &className
		}
		if methodName != "" {
			block.MethodName = &methodName
		}

		blocks = append(blocks, block)
	}

	return blocks, nil
}

// UpdateEmbedding обновляет эмбединг для существующего блока
func (d *Database) UpdateEmbedding(block *models.CodeBlock, embedding []float64) error {
	// Сериализуем эмбединг в JSON
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("ошибка сериализации эмбединга: %w", err)
	}

	// Подготавливаем значения для обновления
	className := ""
	if block.ClassName != nil {
		className = *block.ClassName
	}

	methodName := ""
	if block.MethodName != nil {
		methodName = *block.MethodName
	}

	// Обновляем запись
	query := `
	UPDATE embeddings 
	SET embedding = ?
	WHERE file_path = ? AND class_name = ? AND method_name = ? 
	AND start_line = ? AND end_line = ? AND block_type = ?`

	result, err := d.db.Exec(query,
		string(embeddingJSON),
		block.FilePath,
		className,
		methodName,
		block.StartLine,
		block.EndLine,
		block.BlockType,
	)

	if err != nil {
		return fmt.Errorf("ошибка обновления эмбединга: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества обновлённых строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("блок не найден для обновления")
	}

	return nil
}

// GetStatistics возвращает статистику базы данных
func (d *Database) GetStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Общее количество блоков
	var totalBlocks int
	err := d.db.QueryRow("SELECT COUNT(*) FROM embeddings").Scan(&totalBlocks)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения общего количества блоков: %w", err)
	}
	stats["total_blocks"] = totalBlocks

	// Блоки с эмбедингами
	var blocksWithEmbeddings int
	err = d.db.QueryRow("SELECT COUNT(*) FROM embeddings WHERE embedding != '' AND embedding IS NOT NULL").Scan(&blocksWithEmbeddings)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения количества блоков с эмбедингами: %w", err)
	}
	stats["blocks_with_embeddings"] = blocksWithEmbeddings

	// Блоки без эмбедингов
	var blocksWithoutEmbeddings int
	err = d.db.QueryRow("SELECT COUNT(*) FROM embeddings WHERE embedding = '' OR embedding IS NULL").Scan(&blocksWithoutEmbeddings)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения количества блоков без эмбедингов: %w", err)
	}
	stats["blocks_without_embeddings"] = blocksWithoutEmbeddings

	// Количество файлов
	var fileCount int
	err = d.db.QueryRow("SELECT COUNT(DISTINCT file_path) FROM embeddings").Scan(&fileCount)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения количества файлов: %w", err)
	}
	stats["file_count"] = fileCount

	// Статистика по типам блоков
	rows, err := d.db.Query("SELECT block_type, COUNT(*) FROM embeddings GROUP BY block_type")
	if err != nil {
		return nil, fmt.Errorf("ошибка получения статистики по типам блоков: %w", err)
	}
	defer rows.Close()

	blockTypeStats := make(map[string]int)
	for rows.Next() {
		var blockType string
		var count int
		if err := rows.Scan(&blockType, &count); err != nil {
			return nil, fmt.Errorf("ошибка сканирования статистики типов блоков: %w", err)
		}
		blockTypeStats[blockType] = count
	}
	stats["block_types"] = blockTypeStats

	return stats, nil
}
