package utils

import (
	"strings"
	"unicode"
)

// CountTokens подсчитывает количество токенов в тексте
// Это упрощённая реализация, которая разбивает текст на слова
// В реальном проекте можно использовать более точные токенизаторы
func CountTokens(text string) int {
	if text == "" {
		return 0
	}

	// Разбиваем текст на токены по пробелам и знакам препинания
	tokens := tokenize(text)
	return len(tokens)
}

// tokenize разбивает текст на токены
func tokenize(text string) []string {
	// Убираем лишние пробелы
	text = strings.TrimSpace(text)

	var tokens []string
	var currentToken strings.Builder

	for _, char := range text {
		if unicode.IsSpace(char) {
			// Если накопили токен, сохраняем его
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
		} else if unicode.IsPunct(char) {
			// Если накопили токен, сохраняем его
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			// Знаки препинания тоже считаются токенами
			tokens = append(tokens, string(char))
		} else {
			// Добавляем символ к текущему токену
			currentToken.WriteRune(char)
		}
	}

	// Добавляем последний токен, если он есть
	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}

	return tokens
}
