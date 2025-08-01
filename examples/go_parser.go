package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"gokb-embedder/internal/models"
)

// GoParser парсер для Go файлов
type GoParser struct{}

// NewGoParser создаёт новый парсер Go файлов
func NewGoParser() *GoParser {
	return &GoParser{}
}

// GetName возвращает имя парсера
func (gp *GoParser) GetName() string {
	return "go"
}

// CanParse проверяет, может ли парсер обработать файл с данным расширением
func (gp *GoParser) CanParse(fileExtension string) bool {
	return fileExtension == ".go"
}

// ParseFile парсит Go файл и возвращает блоки кода
func (gp *GoParser) ParseFile(filePath string) ([]*models.CodeBlock, error) {
	// Создаём набор токенов
	fset := token.NewFileSet()

	// Парсим файл
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга Go файла %s: %w", filePath, err)
	}

	var blocks []*models.CodeBlock

	// Обрабатываем функции и методы
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			// Это функция или метод
			block := gp.extractFunctionBlock(fset, x, filePath)
			if block != nil {
				blocks = append(blocks, block)
			}
		case *ast.TypeSpec:
			// Это определение типа (структура, интерфейс и т.д.)
			if structType, ok := x.Type.(*ast.StructType); ok {
				block := gp.extractStructBlock(fset, x, structType, filePath)
				if block != nil {
					blocks = append(blocks, block)
				}
			}
		}
		return true
	})

	return blocks, nil
}

// extractFunctionBlock извлекает блок функции или метода
func (gp *GoParser) extractFunctionBlock(fset *token.FileSet, funcDecl *ast.FuncDecl, filePath string) *models.CodeBlock {
	// Определяем тип блока
	blockType := "function"
	var className *string
	var methodName *string

	if funcDecl.Recv != nil {
		// Это метод
		blockType = "method"
		// Получаем имя типа получателя
		if len(funcDecl.Recv.List) > 0 {
			if starExpr, ok := funcDecl.Recv.List[0].Type.(*ast.StarExpr); ok {
				if ident, ok := starExpr.X.(*ast.Ident); ok {
					className = &ident.Name
				}
			} else if ident, ok := funcDecl.Recv.List[0].Type.(*ast.Ident); ok {
				className = &ident.Name
			}
		}
	}

	// Имя функции/метода
	funcName := funcDecl.Name.Name
	methodName = &funcName

	// Получаем диапазон строк
	startPos := fset.Position(funcDecl.Pos())
	endPos := fset.Position(funcDecl.End())

	startLine := startPos.Line
	endLine := endPos.Line

	// Читаем исходный код
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	// Извлекаем текст функции
	lines := strings.Split(string(content), "\n")
	if startLine > len(lines) || endLine > len(lines) {
		return nil
	}

	funcLines := lines[startLine-1 : endLine]
	funcText := strings.Join(funcLines, "\n")

	// Добавляем комментарий о классе для методов
	if className != nil {
		funcText = fmt.Sprintf("// Method of %s\n%s", *className, funcText)
	}

	return models.NewCodeBlock(
		filePath,
		blockType,
		className,
		methodName,
		startLine,
		endLine,
		funcText,
	)
}

// extractStructBlock извлекает блок структуры
func (gp *GoParser) extractStructBlock(fset *token.FileSet, typeSpec *ast.TypeSpec, structType *ast.StructType, filePath string) *models.CodeBlock {
	// Получаем диапазон строк
	startPos := fset.Position(typeSpec.Pos())
	endPos := fset.Position(typeSpec.End())

	startLine := startPos.Line
	endLine := endPos.Line

	// Читаем исходный код
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	// Извлекаем текст структуры
	lines := strings.Split(string(content), "\n")
	if startLine > len(lines) || endLine > len(lines) {
		return nil
	}

	structLines := lines[startLine-1 : endLine]
	structText := strings.Join(structLines, "\n")

	// Добавляем комментарий о типе
	structText = fmt.Sprintf("// Type definition\n%s", structText)

	className := typeSpec.Name.Name

	return models.NewCodeBlock(
		filePath,
		"struct",
		&className,
		nil, // нет метода для структуры
		startLine,
		endLine,
		structText,
	)
}

// Пример использования:
func main() {
	parser := NewGoParser()

	// Проверяем, может ли парсер обработать Go файл
	if parser.CanParse(".go") {
		fmt.Println("Парсер может обработать Go файлы")
	}

	// Парсим файл (пример)
	// blocks, err := parser.ParseFile("example.go")
	// if err != nil {
	//     fmt.Printf("Ошибка: %v\n", err)
	//     return
	// }

	// for _, block := range blocks {
	//     fmt.Printf("Блок: %s\n", block)
	// }
}
