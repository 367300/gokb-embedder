package config

import "errors"

var (
	// ErrMissingOpenAIKey ошибка отсутствия ключа OpenAI API
	ErrMissingOpenAIKey = errors.New("отсутствует ключ OpenAI API (OPENAI_API_KEY)")
) 