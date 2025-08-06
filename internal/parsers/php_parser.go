package parsers

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gokb-embedder/internal/models"
)

// PHPParser парсер для PHP файлов
type PHPParser struct{}

// NewPHPParser создаёт новый парсер PHP файлов
func NewPHPParser() *PHPParser {
	return &PHPParser{}
}

// GetName возвращает имя парсера
func (pp *PHPParser) GetName() string {
	return "php"
}

// CanParse проверяет, может ли парсер обработать файл с данным расширением
func (pp *PHPParser) CanParse(fileExtension string) bool {
	return fileExtension == ".php"
}

// ParseFile парсит PHP файл и возвращает блоки кода
func (pp *PHPParser) ParseFile(filePath string) ([]*models.CodeBlock, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла %s: %w", filePath, err)
	}

	// Парсим PHP код
	return pp.parsePHPContent(string(content), filePath)
}

// parsePHPContent парсит содержимое PHP файла
func (pp *PHPParser) parsePHPContent(content, filePath string) ([]*models.CodeBlock, error) {
	var blocks []*models.CodeBlock
	lines := strings.Split(content, "\n")

	// Регулярные выражения для поиска различных конструкций
	classRegex := regexp.MustCompile(`^(abstract\s+)?(class|interface|trait)\s+(\w+)`)
	functionRegex := regexp.MustCompile(`^(public|private|protected|static\s+)?(function\s+(\w+)|(\w+)\s*\([^)]*\)\s*\{)`)
	namespaceRegex := regexp.MustCompile(`^namespace\s+([^;]+)`)

	var currentClass *string
	var currentIndent int

	for i, line := range lines {
		lineNum := i + 1
		trimmedLine := strings.TrimSpace(line)

		// Пропускаем пустые строки и комментарии
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "//") || strings.HasPrefix(trimmedLine, "#") || strings.HasPrefix(trimmedLine, "/*") {
			continue
		}

		// Определяем уровень отступа
		indent := getIndentLevel(line)

		// Проверяем namespace (пока не используется, но может быть полезно в будущем)
		if namespaceMatch := namespaceRegex.FindStringSubmatch(trimmedLine); namespaceMatch != nil {
			// namespace := strings.TrimSpace(namespaceMatch[1])
			// currentNamespace = &namespace
			continue
		}

		// Проверяем определение класса, интерфейса или трейта
		if classMatch := classRegex.FindStringSubmatch(trimmedLine); classMatch != nil {
			className := classMatch[3] // Имя класса/интерфейса/трейта
			currentClass = &className
			currentIndent = indent
			continue
		}

		// Проверяем определение функции
		if functionMatch := functionRegex.FindStringSubmatch(trimmedLine); functionMatch != nil {
			funcName := pp.extractFunctionName(functionMatch)

			// Определяем тип блока
			blockType := "function"
			if currentClass != nil && indent > currentIndent {
				blockType = "method"
			}

			// Извлекаем тело функции
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

// extractFunctionName извлекает имя функции из совпадения регулярного выражения
func (pp *PHPParser) extractFunctionName(matches []string) string {
	// Проверяем различные группы захвата
	for i := 3; i < len(matches); i++ {
		if matches[i] != "" {
			return matches[i]
		}
	}
	return "anonymous"
}

// extractFunctionBody извлекает тело функции/метода
func (pp *PHPParser) extractFunctionBody(lines []string, startIndex, baseIndent int) (int, string) {
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
