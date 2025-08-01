package openai

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// Client предоставляет методы для работы с OpenAI API
type Client struct {
	client *openai.Client
}

// NewClient создаёт новый клиент OpenAI
func NewClient(apiKey string) *Client {
	return &Client{
		client: openai.NewClient(apiKey),
	}
}

// GetEmbedding получает эмбединг для текста
func (c *Client) GetEmbedding(ctx context.Context, text string) ([]float64, error) {
	resp, err := c.client.CreateEmbeddings(
		ctx,
		openai.EmbeddingRequest{
			Input: []string{text},
			Model: openai.TextEmbeddingAda002,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения эмбединга: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("пустой ответ от OpenAI API")
	}

	return resp.Data[0].Embedding, nil
}

// GetEmbeddings получает эмбединги для нескольких текстов
func (c *Client) GetEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	resp, err := c.client.CreateEmbeddings(
		ctx,
		openai.EmbeddingRequest{
			Input: texts,
			Model: openai.TextEmbeddingAda002,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения эмбедингов: %w", err)
	}

	embeddings := make([][]float64, len(resp.Data))
	for i, data := range resp.Data {
		embeddings[i] = data.Embedding
	}

	return embeddings, nil
}
