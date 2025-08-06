package parsers

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gokb-embedder/internal/models"
)

// JavaScriptParser парсер для JavaScript файлов
type JavaScriptParser struct{}

// NewJavaScriptParser создаёт новый парсер JavaScript файлов
func NewJavaScriptParser() *JavaScriptParser {
	return &JavaScriptParser{}
}

// GetName возвращает имя парсера
func (jp *JavaScriptParser) GetName() string {
	return "javascript"
}

// CanParse проверяет, может ли парсер обработать файл с данным расширением
func (jp *JavaScriptParser) CanParse(fileExtension string) bool {
	return fileExtension == ".js" || fileExtension == ".jsx" || fileExtension == ".ts" || fileExtension == ".tsx"
}

// ParseFile парсит JavaScript файл и возвращает блоки кода
func (jp *JavaScriptParser) ParseFile(filePath string) ([]*models.CodeBlock, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла %s: %w", filePath, err)
	}

	// Парсим JavaScript код
	return jp.parseJavaScriptContent(string(content), filePath)
}

// parseJavaScriptContent парсит содержимое JavaScript файла
func (jp *JavaScriptParser) parseJavaScriptContent(content, filePath string) ([]*models.CodeBlock, error) {
	var blocks []*models.CodeBlock
	lines := strings.Split(content, "\n")

	// Регулярные выражения для поиска различных конструкций
	classRegex := regexp.MustCompile(`^(export\s+)?(class|interface)\s+(\w+)`)
	functionRegex := regexp.MustCompile(`^(export\s+)?(function\s+(\w+)|const\s+(\w+)\s*=\s*(?:async\s+)?\(|let\s+(\w+)\s*=\s*(?:async\s+)?\(|var\s+(\w+)\s*=\s*(?:async\s+)?\(|(\w+)\s*:\s*(?:async\s+)?\()`)
	arrowFunctionRegex := regexp.MustCompile(`^(export\s+)?(const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?\([^)]*\)\s*=>`)
	methodRegex := regexp.MustCompile(`^(\w+)\s*\([^)]*\)\s*\{`)

	var currentClass *string
	var currentIndent int

	for i, line := range lines {
		lineNum := i + 1
		trimmedLine := strings.TrimSpace(line)

		// Пропускаем пустые строки и комментарии
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "//") || strings.HasPrefix(trimmedLine, "/*") {
			continue
		}

		// Определяем уровень отступа
		indent := getIndentLevel(line)

		// Проверяем определение класса или интерфейса
		if classMatch := classRegex.FindStringSubmatch(trimmedLine); classMatch != nil {
			className := classMatch[3] // Имя класса/интерфейса
			currentClass = &className
			currentIndent = indent
			continue
		}

		// Проверяем определение функции
		if functionMatch := functionRegex.FindStringSubmatch(trimmedLine); functionMatch != nil {
			funcName := jp.extractFunctionName(functionMatch)

			// Определяем тип блока
			blockType := "function"
			if currentClass != nil && indent > currentIndent {
				blockType = "method"
			}

			// Извлекаем тело функции
			startLine := lineNum
			endLine, funcBody := jp.extractFunctionBody(lines, i, indent)

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

		// Проверяем стрелочные функции
		if arrowMatch := arrowFunctionRegex.FindStringSubmatch(trimmedLine); arrowMatch != nil {
			funcName := arrowMatch[3] // Имя функции

			// Определяем тип блока
			blockType := "function"
			if currentClass != nil && indent > currentIndent {
				blockType = "method"
			}

			// Извлекаем тело стрелочной функции
			startLine := lineNum
			endLine, funcBody := jp.extractArrowFunctionBody(lines, i, trimmedLine)

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

		// Проверяем методы классов (без export/function ключевых слов)
		if methodMatch := methodRegex.FindStringSubmatch(trimmedLine); methodMatch != nil {
			if currentClass != nil && indent > currentIndent {
				funcName := methodMatch[1] // Имя метода

				// Извлекаем тело метода
				startLine := lineNum
				endLine, funcBody := jp.extractFunctionBody(lines, i, indent)

				// Создаём блок кода
				block := models.NewCodeBlock(
					filePath,
					"method",
					currentClass,
					&funcName,
					startLine,
					endLine,
					funcBody,
				)

				blocks = append(blocks, block)
			}
		}
	}

	return blocks, nil
}

// extractFunctionName извлекает имя функции из совпадения регулярного выражения
func (jp *JavaScriptParser) extractFunctionName(matches []string) string {
	// Проверяем различные группы захвата
	for i := 3; i < len(matches); i++ {
		if matches[i] != "" {
			return matches[i]
		}
	}
	return "anonymous"
}

// extractFunctionBody извлекает тело функции/метода
func (jp *JavaScriptParser) extractFunctionBody(lines []string, startIndex, baseIndent int) (int, string) {
	var bodyLines []string
	var braceCount int
	var inFunction bool

	for i := startIndex; i < len(lines); i++ {
		line := lines[i]

		// Подсчитываем открывающие и закрывающие скобки
		braceCount += strings.Count(line, "{")
		braceCount -= strings.Count(line, "}")

		if !inFunction && strings.Contains(line, "{") {
			inFunction = true
		}

		if inFunction {
			bodyLines = append(bodyLines, line)
		}

		// Если все скобки закрыты и мы были в функции
		if inFunction && braceCount <= 0 {
			break
		}
	}

	body := strings.Join(bodyLines, "\n")
	return startIndex + len(bodyLines), body
}

// extractArrowFunctionBody извлекает тело стрелочной функции
func (jp *JavaScriptParser) extractArrowFunctionBody(lines []string, startIndex int, firstLine string) (int, string) {
	var bodyLines []string
	bodyLines = append(bodyLines, firstLine)

	// Если стрелочная функция в одну строку
	if strings.Contains(firstLine, "=>") && strings.Contains(firstLine, "{") {
		// Ищем закрывающую скобку в той же строке или в следующих
		var braceCount int
		braceCount += strings.Count(firstLine, "{")
		braceCount -= strings.Count(firstLine, "}")

		if braceCount > 0 {
			// Многострочная стрелочная функция
			for i := startIndex + 1; i < len(lines); i++ {
				line := lines[i]
				bodyLines = append(bodyLines, line)

				braceCount += strings.Count(line, "{")
				braceCount -= strings.Count(line, "}")

				if braceCount <= 0 {
					break
				}
			}
		}
	} else if strings.Contains(firstLine, "=>") && !strings.Contains(firstLine, "{") {
		// Стрелочная функция с неявным возвратом
		// Ищем конец выражения (обычно точка с запятой или новая строка)
		for i := startIndex + 1; i < len(lines); i++ {
			line := lines[i]
			trimmedLine := strings.TrimSpace(line)

			if trimmedLine == "" || strings.HasSuffix(trimmedLine, ";") {
				break
			}

			bodyLines = append(bodyLines, line)
		}
	}

	body := strings.Join(bodyLines, "\n")
	return startIndex + len(bodyLines), body
}
