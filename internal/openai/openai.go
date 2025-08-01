package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client предоставляет методы для работы с OpenAI API
type Client struct {
	apiKey string
	client *http.Client
}

// NewClient создаёт новый клиент OpenAI
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

// GetEmbedding получает эмбединг для текста
func (c *Client) GetEmbedding(ctx context.Context, text string) ([]float64, error) {
	// Создаём запрос к OpenAI API
	requestBody := map[string]interface{}{
		"input": text,
		"model": "text-embedding-3-small",
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации запроса: %w", err)
	}

	// Создаём HTTP запрос
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Выполняем запрос
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ошибка API: %s - %s", resp.Status, string(body))
	}

	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Парсим ответ
	var response struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("пустой ответ от OpenAI API")
	}

	return response.Data[0].Embedding, nil
}

// GetEmbeddings получает эмбединги для нескольких текстов
func (c *Client) GetEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	// Создаём запрос к OpenAI API
	requestBody := map[string]interface{}{
		"input": texts,
		"model": "text-embedding-3-small",
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации запроса: %w", err)
	}

	// Создаём HTTP запрос
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Выполняем запрос
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ошибка API: %s - %s", resp.Status, string(body))
	}

	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Парсим ответ
	var response struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	embeddings := make([][]float64, len(response.Data))
	for i, data := range response.Data {
		embeddings[i] = data.Embedding
	}

	return embeddings, nil
}
