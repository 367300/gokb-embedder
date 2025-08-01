package utils

import (
	"testing"
)

func TestCountTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "пустая строка",
			input:    "",
			expected: 0,
		},
		{
			name:     "простое слово",
			input:    "hello",
			expected: 1,
		},
		{
			name:     "несколько слов",
			input:    "hello world",
			expected: 2,
		},
		{
			name:     "с знаками препинания",
			input:    "hello, world!",
			expected: 4, // hello, world, ,, !
		},
		{
			name:     "с новой строкой",
			input:    "hello\nworld",
			expected: 2,
		},
		{
			name:     "с табуляцией",
			input:    "hello\tworld",
			expected: 2,
		},
		{
			name:     "смешанный текст",
			input:    "def hello_world(): return 'Hello, World!'",
			expected: 14, // def, hello_world, (, ), :, return, ', Hello, ,, World, !, '
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountTokens(tt.input)
			if result != tt.expected {
				t.Errorf("CountTokens(%q) = %d, ожидалось %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "пустая строка",
			input:    "",
			expected: []string{},
		},
		{
			name:     "одно слово",
			input:    "hello",
			expected: []string{"hello"},
		},
		{
			name:     "два слова",
			input:    "hello world",
			expected: []string{"hello", "world"},
		},
		{
			name:     "с запятой",
			input:    "hello, world",
			expected: []string{"hello", ",", "world"},
		},
		{
			name:     "с восклицательным знаком",
			input:    "hello!",
			expected: []string{"hello", "!"},
		},
		{
			name:     "с несколькими знаками препинания",
			input:    "hello, world!",
			expected: []string{"hello", ",", "world", "!"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tokenize(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("tokenize(%q) вернул %d токенов, ожидалось %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i, token := range result {
				if i >= len(tt.expected) || token != tt.expected[i] {
					t.Errorf("tokenize(%q)[%d] = %q, ожидалось %q", tt.input, i, token, tt.expected[i])
				}
			}
		})
	}
}
