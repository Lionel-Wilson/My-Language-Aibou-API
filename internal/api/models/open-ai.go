package models

// Define the structure for the choices array
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	Logprobs     *string `json:"logprobs"`
	FinishReason string  `json:"finish_reason"`
}

// Define the structure for the message field within choices
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Define the structure for the usage field
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Define the main structure
type ChatCompletion struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	SystemFingerprint string   `json:"system_fingerprint"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
}
