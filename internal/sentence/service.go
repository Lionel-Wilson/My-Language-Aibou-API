package sentence

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/coocood/freecache"
	"go.uber.org/zap"

	openai "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/clients/open-ai"
)

var ErrOpenAiNoChoices = errors.New("OpenAI API response contains no choices")

//go:generate mockgen -source=service.go -destination=mock/service.go
type Service interface {
	GetSentenceExplanation(sentence string, nativeLanguage string) (*string, error)
	GetSentenceCorrection(sentence string, nativeLanguage string) (*string, error)
	ValidateSentence(sentence string) error
}

type service struct {
	logger       *zap.Logger
	openAiClient openai.Client
	cache        *freecache.Cache
}

func NewSentenceService(
	logger *zap.Logger,
	openAiClient openai.Client,
	cache *freecache.Cache,
) Service {
	return &service{
		logger:       logger,
		openAiClient: openAiClient,
		cache:        cache,
	}
}

func (s *service) GetSentenceCorrection(sentence string, nativeLanguage string) (*string, error) {
	s.logger.Debug("Getting sentence correction", zap.String("sentence", sentence), zap.String("nativeLanguage", nativeLanguage))
	jsonBody := s.sentenceToOpenAiSentenceCorrectionRequestBody(sentence, nativeLanguage)

	resp, responseBody, err := s.openAiClient.MakeRequest(jsonBody)
	if err != nil {
		s.logger.Error(err.Error())

		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		s.logger.Sugar().Infof("OpenAI API returned non-OK status: %d", resp.StatusCode)

		return nil, err
	}

	var OpenAIApiResponse openai.ChatCompletion

	err = json.Unmarshal(responseBody, &OpenAIApiResponse)
	if err != nil {
		return nil, err
	}

	if len(OpenAIApiResponse.Choices) == 0 {
		return nil, ErrOpenAiNoChoices
	}

	s.logger.Sugar().Infof("Prompt Tokens: %d", OpenAIApiResponse.Usage.PromptTokens)
	s.logger.Sugar().Infof("Response Tokens: %d", OpenAIApiResponse.Usage.CompletionTokens)
	s.logger.Sugar().Infof("Total Tokens used: %d", OpenAIApiResponse.Usage.TotalTokens)

	return &OpenAIApiResponse.Choices[0].Message.Content, nil
}

func (s *service) GetSentenceExplanation(sentence string, nativeLanguage string) (*string, error) {
	s.logger.Info("Getting sentence explanation", zap.String("sentence", sentence), zap.String("nativeLanguage", nativeLanguage))
	cacheKey := []byte(fmt.Sprintf("%s sentence explanation in %s", sentence, nativeLanguage))

	cached, err := s.cache.Get(cacheKey)
	if err == nil {
		cachedResponse := string(cached)
		return &cachedResponse, err
	}
	jsonBody := s.sentenceToOpenAiExplanationRequestBody(sentence, nativeLanguage)

	resp, responseBody, err := s.openAiClient.MakeRequest(jsonBody)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		s.logger.Sugar().Infof("OpenAI API returned non-OK status: %d", resp.StatusCode)

		return nil, err
	}

	var OpenAIApiResponse openai.ChatCompletion

	err = json.Unmarshal(responseBody, &OpenAIApiResponse)
	if err != nil {
		return nil, err
	}

	if len(OpenAIApiResponse.Choices) == 0 {
		return nil, ErrOpenAiNoChoices
	}

	s.logger.Sugar().Infof("Prompt Tokens: %d", OpenAIApiResponse.Usage.PromptTokens)
	s.logger.Sugar().Infof("Response Tokens: %d", OpenAIApiResponse.Usage.CompletionTokens)
	s.logger.Sugar().Infof("Total Tokens used: %d", OpenAIApiResponse.Usage.TotalTokens)

	result := &OpenAIApiResponse.Choices[0].Message.Content

	cacheValue := []byte(*result)
	expiration := 2592000 //30 days
	err = s.cache.Set(cacheKey, cacheValue, expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to cache sentence explanation: %w", err)
	}

	return result, nil
}

func (s *service) ValidateSentence(sentence string) error {
	if sentence == "" {
		return errors.New("please provide a sentence")
	}

	if utf8.RuneCountInString(sentence) > 100 {
		return errors.New("the sentence must be less than 100 characters")
	}

	return nil
}

func (s *service) sentenceToOpenAiExplanationRequestBody(sentence, userNativeLanguage string) *strings.Reader {
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

	// s.logger.Sugar().Infof("Tier: %s\n", userTier)
	// s.logger.Sugar().Infof("Body: %s\n", body)
	s.logger.Sugar().Infof("Phrase prompt: %s\n", content)

	return strings.NewReader(body)
}

func (s *service) sentenceToOpenAiSentenceCorrectionRequestBody(sentence, userNativeLanguage string) *strings.Reader {
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

	s.logger.Sugar().Infof("Phrase prompt: %s\n", content)

	return strings.NewReader(body)
}
