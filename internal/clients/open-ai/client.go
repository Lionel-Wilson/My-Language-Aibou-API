package openai

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

//go:generate mockgen -source=client.go -destination=mock/client.go
type Client interface {
	MakeRequest(ctx context.Context, body io.Reader) (*http.Response, []byte, error)
}

type (
	OpenAIRequest struct {
		Model       string    `json:"model"`
		Messages    []Message `json:"messages"`
		Temperature float32   `json:"temperature"`
		MaxTokens   int       `json:"max_tokens"`
	}

	ChatCompletion struct {
		ID                string   `json:"id"`
		Object            string   `json:"object"`
		Created           int64    `json:"created"`
		Model             string   `json:"model"`
		SystemFingerprint string   `json:"system_fingerprint"`
		Choices           []Choice `json:"choices"`
		Usage             Usage    `json:"usage"`
	}

	Choice struct {
		Index        int     `json:"index"`
		Message      Message `json:"message"`
		Logprobs     *string `json:"logprobs"`
		FinishReason string  `json:"finish_reason"`
	}

	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	}
)

type openAiClient struct {
	Key    string
	logger *zap.Logger
}

func NewClient(apiKey string, logger *zap.Logger) Client {
	return &openAiClient{
		Key:    apiKey,
		logger: logger,
	}
}

func (c openAiClient) MakeRequest(ctx context.Context, body io.Reader) (*http.Response, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create post request : %w", err)
	}

	req.Header.Add("Content-Type", `application/json`)
	req.Header.Add("Authorization", `Bearer `+c.Key)

	client := &http.Client{
		Timeout: time.Second * 45,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to make request to OpenAI API : %w", err)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read AI response body %v: %w", resp.Body, err)
	}

	return resp, responseBody, nil
}
