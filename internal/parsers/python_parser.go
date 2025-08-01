package parsers

import (
	"fmt"
	"os"
	"strings"

	"gokb-embedder/internal/models"
)

// PythonParser парсер для Python файлов
type PythonParser struct{}

// NewPythonParser создаёт новый парсер Python файлов
func NewPythonParser() *PythonParser {
	return &PythonParser{}
}

// GetName возвращает имя парсера
func (pp *PythonParser) GetName() string {
	return "python"
}

// CanParse проверяет, может ли парсер обработать файл с данным расширением
func (pp *PythonParser) CanParse(fileExtension string) bool {
	return fileExtension == ".py"
}

// ParseFile парсит Python файл и возвращает блоки кода
func (pp *PythonParser) ParseFile(filePath string) ([]*models.CodeBlock, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла %s: %w", filePath, err)
	}

	// Парсим Python код (упрощённая версия)
	return pp.parsePythonContent(string(content), filePath)
}

// parsePythonContent парсит содержимое Python файла
func (pp *PythonParser) parsePythonContent(content, filePath string) ([]*models.CodeBlock, error) {
	var blocks []*models.CodeBlock
	lines := strings.Split(content, "\n")

	// Простой парсер для извлечения классов и функций
	var currentClass *string
	var currentIndent int

	for i, line := range lines {
		lineNum := i + 1
		trimmedLine := strings.TrimSpace(line)

		// Пропускаем пустые строки и комментарии
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		// Определяем уровень отступа
		indent := getIndentLevel(line)

		// Проверяем определение класса
		if strings.HasPrefix(trimmedLine, "class ") && strings.Contains(trimmedLine, ":") {
			className := extractClassName(trimmedLine)
			currentClass = &className
			currentIndent = indent
			continue
		}

		// Проверяем определение функции/метода
		if strings.HasPrefix(trimmedLine, "def ") && strings.Contains(trimmedLine, ":") {
			funcName := extractFunctionName(trimmedLine)

			// Определяем тип блока
			blockType := "function"
			if currentClass != nil && indent > currentIndent {
				blockType = "method"
			}

			// Извлекаем тело функции/метода
			startLine := lineNum
			endLine, funcBody := pp.extractFunctionBody(lines, i, indent)

			// Создаём блок кода
			block := models.NewCodeBlock(
				filePath,
				blockType,
				currentClass,
				&funcName,
				startLine,
				endLine,
				funcBody,
			)

			blocks = append(blocks, block)
		}
	}

	return blocks, nil
}

// getIndentLevel возвращает уровень отступа строки
func getIndentLevel(line string) int {
	indent := 0
	for _, char := range line {
		if char == ' ' || char == '\t' {
			indent++
		} else {
			break
		}
	}
	return indent
}

// extractClassName извлекает имя класса из строки определения
func extractClassName(line string) string {
	// Убираем "class " и всё после ":"
	parts := strings.Split(line, "class ")
	if len(parts) < 2 {
		return ""
	}

	className := strings.Split(parts[1], ":")[0]
	className = strings.TrimSpace(className)

	// Убираем наследование, если есть
	if strings.Contains(className, "(") {
		className = strings.Split(className, "(")[0]
	}

	return strings.TrimSpace(className)
}

// extractFunctionName извлекает имя функции из строки определения
func extractFunctionName(line string) string {
	// Убираем "def " и всё после ":"
	parts := strings.Split(line, "def ")
	if len(parts) < 2 {
		return ""
	}

	funcName := strings.Split(parts[1], ":")[0]
	funcName = strings.TrimSpace(funcName)

	// Убираем параметры, если есть
	if strings.Contains(funcName, "(") {
		funcName = strings.Split(funcName, "(")[0]
	}

	return strings.TrimSpace(funcName)
}

// extractFunctionBody извлекает тело функции/метода
func (pp *PythonParser) extractFunctionBody(lines []string, startIndex, baseIndent int) (int, string) {
	var bodyLines []string
	endLine := startIndex + 1

	// Добавляем строку определения функции
	bodyLines = append(bodyLines, lines[startIndex])

	// Ищем конец функции по отступам
	for i := startIndex + 1; i < len(lines); i++ {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		// Пропускаем пустые строки
		if trimmedLine == "" {
			bodyLines = append(bodyLines, line)
			continue
		}

		indent := getIndentLevel(line)

		// Если отступ меньше или равен базовому, функция закончилась
		if indent <= baseIndent && trimmedLine != "" {
			break
		}

		bodyLines = append(bodyLines, line)
		endLine = i + 1
	}

	return endLine, strings.Join(bodyLines, "\n")
}
