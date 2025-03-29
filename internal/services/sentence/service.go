package sentence

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	openai "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/clients/open-ai"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/utils"
)

var (
	FailedToProcessSentence = "Failed to process your sentence(s).Please make sure you remove any line breaks and large gaps between your sentences and try again"
	ErrOpenAiNoChoices      = errors.New("OpenAI API response contains no choices")
)

//go:generate mockgen -source=service.go -destination=mock/service.go
type Service interface {
	GetSentenceExplanation(c *gin.Context, sentence string, nativeLanguage string) (*openai.ChatCompletion, error)
	GetSentenceCorrection(c *gin.Context, sentence string, nativeLanguage string) (*openai.ChatCompletion, error)
	ValidateSentence(sentence string) error
}

type service struct {
	logger       zap.Logger
	openAiClient openai.Client
}

func New(logger zap.Logger, openAiClient openai.Client) Service {
	return &service{
		logger:       logger,
		openAiClient: openAiClient,
	}
}

func (s *service) GetSentenceCorrection(c *gin.Context, sentence string, nativeLanguage string) (*openai.ChatCompletion, error) {
	jsonBody := sentenceToOpenAiSentenceCorrectionRequestBody(sentence, nativeLanguage)

	resp, responseBody, err := s.openAiClient.MakeRequest(jsonBody)
	if err != nil {
		s.logger.Error(err.Error())
		utils.ServerErrorResponse(c, err, FailedToProcessSentence)

		return &openai.ChatCompletion{}, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("OpenAI API returned non-OK status. ")
		utils.ServerErrorResponse(c, err, FailedToProcessSentence)

		return &openai.ChatCompletion{}, err
	}

	var OpenAIApiResponse openai.ChatCompletion

	err = json.Unmarshal(responseBody, &OpenAIApiResponse)
	if err != nil {
		fmt.Println("Failed to unmarshal json body")
		return &openai.ChatCompletion{}, err
	}

	if len(OpenAIApiResponse.Choices) == 0 {
		fmt.Println("OpenAI API response contains no choices")

		err = ErrOpenAiNoChoices
		utils.ServerErrorResponse(c, err, FailedToProcessSentence)

		return &openai.ChatCompletion{}, err
	}

	fmt.Printf("Phrase explanation: %s\n", OpenAIApiResponse.Choices[0].Message.Content)
	fmt.Printf("Prompt Tokens: %d\n", OpenAIApiResponse.Usage.PromptTokens)
	fmt.Printf("Response Tokens: %d\n", OpenAIApiResponse.Usage.CompletionTokens)
	fmt.Printf("Total Tokens used: %d\n", OpenAIApiResponse.Usage.TotalTokens)

	return &OpenAIApiResponse, nil
}

func (s *service) GetSentenceExplanation(c *gin.Context, sentence string, nativeLanguage string) (*openai.ChatCompletion, error) {
	jsonBody := sentenceToOpenAiExplanationRequestBody(sentence, nativeLanguage)

	resp, responseBody, err := s.openAiClient.MakeRequest(jsonBody)
	if err != nil {
		s.logger.Error(err.Error())
		utils.ServerErrorResponse(c, err, FailedToProcessSentence)

		return &openai.ChatCompletion{}, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("OpenAI API returned non-OK status. ")
		utils.ServerErrorResponse(c, err, FailedToProcessSentence)

		return &openai.ChatCompletion{}, err
	}

	var OpenAIApiResponse openai.ChatCompletion

	err = json.Unmarshal(responseBody, &OpenAIApiResponse)
	if err != nil {
		fmt.Println("Failed to unmarshal json body")
		return &openai.ChatCompletion{}, err
	}

	if len(OpenAIApiResponse.Choices) == 0 {
		fmt.Println("OpenAI API response contains no choices")

		err = ErrOpenAiNoChoices
		utils.ServerErrorResponse(c, err, FailedToProcessSentence)

		return &openai.ChatCompletion{}, err
	}

	fmt.Printf("Phrase explanation: %s\n", OpenAIApiResponse.Choices[0].Message.Content)
	fmt.Printf("Prompt Tokens: %d\n", OpenAIApiResponse.Usage.PromptTokens)
	fmt.Printf("Response Tokens: %d\n", OpenAIApiResponse.Usage.CompletionTokens)
	fmt.Printf("Total Tokens used: %d\n", OpenAIApiResponse.Usage.TotalTokens)

	return &OpenAIApiResponse, nil
}

func (s *service) ValidateSentence(sentence string) error {
	if sentence == "" {
		s.logger.Sugar().Errorf("User didn't provide a sentence: %s", sentence)
		return errors.New("Please provide a sentence")
	}

	if utf8.RuneCountInString(sentence) > 100 {
		s.logger.Sugar().Errorf("Sentence '%s' length too long. Must be less than 100 characters.", sentence)
		return errors.New("The sentence must be less than 100 characters.")
	}

	return nil
}

func sentenceToOpenAiExplanationRequestBody(sentence, userNativeLanguage string) *strings.Reader {
	// var maxWordCount string
	// var MaxTokens string
	/*Remove tier system.
	if userTier == "Basic" {
		MaxTokens = "110"
		maxWordCount = "80"

	} else if userTier == "Premium" {
		MaxTokens = "330"
		maxWordCount = "230"
	}*/
	content := fmt.Sprintf("Explain the meaning & grammar used in this sentence - '%s'.Respond in %s", sentence, userNativeLanguage)

	body := fmt.Sprintf(`{
	"model":"gpt-4o",
	"messages": [{
		"role": "system",
		"content": "You are a helpful assistant."
	  },
	  {
		"role": "user",
		"content": "%s"
	  }],
	"temperature": 0.4,
	"max_tokens": 800
	}`, content)

	// fmt.Printf("Tier: %s\n", userTier)
	// fmt.Printf("Body: %s\n", body)
	fmt.Printf("Phrase prompt: %s\n", content)

	return strings.NewReader(body)
}

func sentenceToOpenAiSentenceCorrectionRequestBody(sentence, userNativeLanguage string) *strings.Reader {
	content := fmt.Sprintf("Is this sentence correct? if not then correct it for me - '%s'", sentence)

	if userNativeLanguage != "English" {
		content = fmt.Sprintf("Is this sentence correct? if not then correct it for me - '%s'. Respond in %s as if you're a language teacher teaching a native %s speaker who's learning this sentence's language.", sentence, userNativeLanguage, userNativeLanguage)
	}

	body := fmt.Sprintf(`{
	"model":"gpt-4o",
	"messages": [{
		"role": "system",
		"content": "You are a helpful assistant."
	  },
	  {
		"role": "user",
		"content": "%s"
	  }],
	"temperature": 0.4,
	"max_tokens": 800
	}`, content)

	fmt.Printf("Phrase prompt: %s\n", content)

	return strings.NewReader(body)
}
