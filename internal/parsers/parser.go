package parsers

import (
	"gokb-embedder/internal/models"
)

// Parser интерфейс для парсеров различных типов файлов
type Parser interface {
	// ParseFile парсит файл и возвращает блоки кода
	ParseFile(filePath string) ([]*models.CodeBlock, error)

	// CanParse проверяет, может ли парсер обработать файл с данным расширением
	CanParse(fileExtension string) bool

	// GetName возвращает имя парсера
	GetName() string
}

// ParserRegistry реестр всех доступных парсеров
type ParserRegistry struct {
	parsers map[string]Parser
}

// NewParserRegistry создаёт новый реестр парсеров
func NewParserRegistry() *ParserRegistry {
	return &ParserRegistry{
		parsers: make(map[string]Parser),
	}
}

// Register регистрирует парсер в реестре
func (pr *ParserRegistry) Register(parser Parser) {
	pr.parsers[parser.GetName()] = parser
}

// GetParser возвращает парсер для указанного расширения файла
func (pr *ParserRegistry) GetParser(fileExtension string) (Parser, bool) {
	for _, parser := range pr.parsers {
		if parser.CanParse(fileExtension) {
			return parser, true
		}
	}
	return nil, false
}

// GetAllParsers возвращает все зарегистрированные парсеры
func (pr *ParserRegistry) GetAllParsers() []Parser {
	var result []Parser
	for _, parser := range pr.parsers {
		result = append(result, parser)
	}
	return result
}
