package parsers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gokb-embedder/internal/models"
	"gokb-embedder/internal/utils"
)

// TextParser парсер для текстовых файлов (md, yml, conf)
type TextParser struct {
	tokenLimit int
}

// NewTextParser создаёт новый парсер текстовых файлов
func NewTextParser(tokenLimit int) *TextParser {
	return &TextParser{
		tokenLimit: tokenLimit,
	}
}

// GetName возвращает имя парсера
func (tp *TextParser) GetName() string {
	return "text"
}

// CanParse проверяет, может ли парсер обработать файл с данным расширением
func (tp *TextParser) CanParse(fileExtension string) bool {
	extensions := []string{".md", ".yml", ".yaml", ".conf", ".config", ".txt"}
	for _, ext := range extensions {
		if fileExtension == ext {
			return true
		}
	}
	return false
}

// ParseFile парсит текстовый файл и возвращает блоки кода
func (tp *TextParser) ParseFile(filePath string) ([]*models.CodeBlock, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла %s: %w", filePath, err)
	}

	return tp.splitTextByTokens(string(content), filePath)
}

// splitTextByTokens разбивает текст на блоки по токенам
func (tp *TextParser) splitTextByTokens(content, filePath string) ([]*models.CodeBlock, error) {
	var blocks []*models.CodeBlock
	lines := strings.Split(content, "\n")

	var currentBlock []string
	currentTokens := 0
	startLine := 1

	for i, line := range lines {
		lineNum := i + 1
		lineTokens := utils.CountTokens(line)

		// Если добавление строки превысит лимит токенов и у нас уже есть блок
		if currentTokens+lineTokens > tp.tokenLimit && len(currentBlock) > 0 {
			// Сохраняем текущий блок
			blockText := strings.Join(currentBlock, "\n")
			block := models.NewCodeBlock(
				filePath,
				tp.getBlockType(filePath),
				nil, // class_name
				nil, // method_name
				startLine,
				lineNum-1,
				blockText,
			)
			blocks = append(blocks, block)

			// Начинаем новый блок
			currentBlock = []string{line}
			currentTokens = lineTokens
			startLine = lineNum
		} else {
			// Добавляем строку к текущему блоку
			currentBlock = append(currentBlock, line)
			currentTokens += lineTokens
		}
	}

	// Добавляем последний блок, если он есть
	if len(currentBlock) > 0 {
		blockText := strings.Join(currentBlock, "\n")
		block := models.NewCodeBlock(
			filePath,
			tp.getBlockType(filePath),
			nil, // class_name
			nil, // method_name
			startLine,
			len(lines),
			blockText,
		)
		blocks = append(blocks, block)
	}

	return blocks, nil
}

// getBlockType определяет тип блока на основе расширения файла
func (tp *TextParser) getBlockType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".md":
		return "markdown"
	case ".yml", ".yaml":
		return "yaml"
	case ".conf", ".config":
		return "config"
	case ".txt":
		return "text"
	default:
		return "text"
	}
}
