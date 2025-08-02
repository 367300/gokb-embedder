package models

import (
	"encoding/json"
	"fmt"
	"path/filepath"
)

// CodeBlock представляет блок кода с метаинформацией
type CodeBlock struct {
	FilePath       string   `json:"file_path"`     // Абсолютный путь к файлу
	RelativePath   string   `json:"relative_path"` // Относительный путь от корня проекта
	BlockType      string   `json:"block_type"`
	ClassName      *string  `json:"class_name,omitempty"`
	MethodName     *string  `json:"method_name,omitempty"`
	StartLine      int      `json:"start_line"`
	EndLine        int      `json:"end_line"`
	RawText        string   `json:"raw_text"`
	CommitMessages []string `json:"commit_messages,omitempty"`
}

// NewCodeBlock создаёт новый блок кода
func NewCodeBlock(filePath, blockType string, className, methodName *string, startLine, endLine int, rawText string) *CodeBlock {
	return &CodeBlock{
		FilePath:     filePath,
		RelativePath: filePath, // По умолчанию относительный путь равен абсолютному
		BlockType:    blockType,
		ClassName:    className,
		MethodName:   methodName,
		StartLine:    startLine,
		EndLine:      endLine,
		RawText:      rawText,
	}
}

// SetCommitMessages устанавливает сообщения коммитов
func (cb *CodeBlock) SetCommitMessages(messages []string) {
	cb.CommitMessages = messages
}

// GetEmbeddingText формирует текст для создания эмбединга
func (cb *CodeBlock) GetEmbeddingText() string {
	// Используем относительный путь для отображения в тексте эмбединга
	displayPath := cb.RelativePath
	if displayPath == "" {
		displayPath = cb.FilePath // Fallback на абсолютный путь
	}

	text := fmt.Sprintf("File: %s\n", displayPath)

	if cb.ClassName != nil {
		text += fmt.Sprintf("Class: %s\n", *cb.ClassName)
	}

	if cb.MethodName != nil {
		text += fmt.Sprintf("Method/Function: %s\n", *cb.MethodName)
	}

	text += fmt.Sprintf("Lines: %d-%d\n", cb.StartLine, cb.EndLine)

	if len(cb.CommitMessages) > 0 {
		text += fmt.Sprintf("Recent commits: %s\n", joinStrings(cb.CommitMessages, "; "))
	}

	text += fmt.Sprintf("\nCode:\n%s", cb.RawText)
	return text
}

// GetFileName возвращает имя файла без пути
func (cb *CodeBlock) GetFileName() string {
	return filepath.Base(cb.FilePath)
}

// GetRelativePath возвращает относительный путь файла
func (cb *CodeBlock) GetRelativePath() string {
	if cb.RelativePath != "" {
		return cb.RelativePath
	}
	return cb.FilePath
}

// SetRelativePath устанавливает относительный путь файла
func (cb *CodeBlock) SetRelativePath(relativePath string) {
	cb.RelativePath = relativePath
}

// String возвращает строковое представление блока
func (cb *CodeBlock) String() string {
	className := ""
	if cb.ClassName != nil {
		className = *cb.ClassName
	}

	methodName := ""
	if cb.MethodName != nil {
		methodName = *cb.MethodName
	}

	return fmt.Sprintf("<Block %s %s %s %s:%d-%d>",
		cb.BlockType, className, methodName, cb.FilePath, cb.StartLine, cb.EndLine)
}

// MarshalJSON кастомная сериализация для JSON
func (cb *CodeBlock) MarshalJSON() ([]byte, error) {
	type Alias CodeBlock

	aux := &struct {
		*Alias
		ClassName  interface{} `json:"class_name,omitempty"`
		MethodName interface{} `json:"method_name,omitempty"`
	}{
		Alias: (*Alias)(cb),
	}

	if cb.ClassName != nil {
		aux.ClassName = *cb.ClassName
	}

	if cb.MethodName != nil {
		aux.MethodName = *cb.MethodName
	}

	return json.Marshal(aux)
}

// joinStrings объединяет строки с разделителем
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
