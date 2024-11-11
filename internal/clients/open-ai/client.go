package openai

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/log"
)

//go:generate mockgen -source=client.go -destination=mock/client.go
type Client interface {
	MakeRequest(body *strings.Reader) (*http.Response, []byte, error)
}

type (
	// Define the structure for the choices array
	Choice struct {
		Index        int     `json:"index"`
		Message      Message `json:"message"`
		Logprobs     *string `json:"logprobs"`
		FinishReason string  `json:"finish_reason"`
	}

	// Define the structure for the message field within choices
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	// Define the structure for the usage field
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	}

	// Define the main structure
	ChatCompletion struct {
		ID                string   `json:"id"`
		Object            string   `json:"object"`
		Created           int64    `json:"created"`
		Model             string   `json:"model"`
		SystemFingerprint string   `json:"system_fingerprint"`
		Choices           []Choice `json:"choices"`
		Usage             Usage    `json:"usage"`
	}
)

type openAiClient struct {
	Key    string
	logger log.Logger
}

func NewClient(apiKey string, logger log.Logger) Client {
	return &openAiClient{
		Key:    apiKey,
		logger: logger,
	}
}

func (c openAiClient) MakeRequest(body *strings.Reader) (*http.Response, []byte, error) {
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", body)
	if err != nil {
		fmt.Println("Failed to create request")
		return &http.Response{}, []byte{}, err
	}

	req.Header.Add("Content-Type", `application/json`)
	req.Header.Add("Authorization", `Bearer `+c.Key)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to make request to OpenAI API")
		return &http.Response{}, []byte{}, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Failed to read AI response body:")
		fmt.Println(string(responseBody))
		return &http.Response{}, []byte{}, err
	}

	return resp, responseBody, nil
}
