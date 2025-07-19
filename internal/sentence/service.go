package sentence

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/coocood/freecache"
	"go.uber.org/zap"

	openai "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/clients/open-ai"
)

var ErrOpenAiNoChoices = errors.New("OpenAI API response contains no choices")

//go:generate mockgen -source=service.go -destination=mock/service.go
type Service interface {
	GetSentenceExplanation(sentence string, nativeLanguage string, isDetailed bool) (*string, error)
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

var (
	sentenceCacheExpiration = int(time.Hour * 24 * 30) //30 days
)

func (s *service) GetSentenceCorrection(sentence string, nativeLanguage string) (*string, error) {
	s.logger.Debug("Getting sentence correction", zap.String("sentence", sentence), zap.String("nativeLanguage", nativeLanguage))
	cacheKey := []byte(fmt.Sprintf("%s sentence correction in %s", sentence, nativeLanguage))

	cached, err := s.cache.Get(cacheKey)
	if err == nil {
		cachedResponse := string(cached)
		return &cachedResponse, err
	}
	jsonBody := s.sentenceToOpenAiSentenceCorrectionRequestBody(sentence, nativeLanguage)

	resp, responseBody, err := s.openAiClient.MakeRequest(jsonBody)
	if err != nil {
		s.logger.Error(err.Error())

		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openAI API returned non-OK status(%s): %s", resp.StatusCode, responseBody)
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
	err = s.cache.Set(cacheKey, cacheValue, sentenceCacheExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to cache sentence correction: %w", err)
	}

	return result, nil
}

func (s *service) GetSentenceExplanation(sentence string, nativeLanguage string, isDetailed bool) (*string, error) {
	s.logger.Info("Getting sentence explanation", zap.String("sentence", sentence), zap.String("nativeLanguage", nativeLanguage))
	cacheKey := []byte(fmt.Sprintf("%s sentence explanation in %s(%v)", sentence, nativeLanguage, isDetailed))

	cached, err := s.cache.Get(cacheKey)
	if err == nil {
		cachedResponse := string(cached)
		return &cachedResponse, err
	}

	var jsonBody io.Reader
	if isDetailed {
		jsonBody = s.sentenceToOpenAiExplanationRequestBody(sentence, nativeLanguage)
	} else {
		jsonBody = s.sentenceToOpenAiSimpleTranslationRequestBody(sentence, nativeLanguage)
	}

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
	err = s.cache.Set(cacheKey, cacheValue, sentenceCacheExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to cache sentence explanation: %w", err)
	}

	return result, nil
}

func (s *service) ValidateSentence(sentence string) error {
	if sentence == "" {
		return errors.New("Please provide a sentence")
	}

	if utf8.RuneCountInString(sentence) > 120 {
		return errors.New("The sentence must be less than 120 characters.")
	}

	return nil
}

func (s *service) sentenceToOpenAiSimpleTranslationRequestBody(sentence, userNativeLanguage string) *bytes.Reader {
	content := fmt.Sprintf(
		"Translate the following sentence into %s. Do not provide any explanation or context—only the translated sentence.\n\nSentence: %s",
		userNativeLanguage,
		sentence,
	)

	payload := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a precise translator. Only return the translation without explanations.",
			},
			{
				"role":    "user",
				"content": content,
			},
		},
		"temperature": 0.2,
		"max_tokens":  300,
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error("Failed to marshal simple translation payload", zap.Error(err))
		return bytes.NewReader([]byte{})
	}

	return bytes.NewReader(jsonBody)
}

func (s *service) sentenceToOpenAiExplanationRequestBody(sentence, userNativeLanguage string) *strings.Reader {
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

	return strings.NewReader(body)
}

func (s *service) sentenceToOpenAiSentenceCorrectionRequestBody(sentence, userNativeLanguage string) *bytes.Reader {
	var content string

	if userNativeLanguage == "English" {
		content = fmt.Sprintf(
			"Is this sentence correct? If not, correct it and briefly explain why. Do not ask follow-up questions or encourage further conversation. Just provide the correction and explanation in a single, complete answer.\n\nSentence: %s",
			sentence,
		)
	} else {
		content = fmt.Sprintf(
			"Is this sentence correct? If not, correct it and briefly explain why. Do not ask follow-up questions or encourage further conversation. Just provide the correction and explanation in a single, complete answer. Respond in %s as if you're a language teacher teaching a native %s speaker.\n\nSentence: %s",
			userNativeLanguage,
			userNativeLanguage,
			sentence,
		)
	}

	payload := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a concise language assistant that explains sentence corrections in a clear and brief manner without engaging in back-and-forth conversation.",
			},
			{
				"role":    "user",
				"content": content,
			},
		},
		"temperature": 0.4,
		"max_tokens":  800,
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error("Failed to marshal sentence correction payload", zap.Error(err))
		return bytes.NewReader([]byte{}) // fallback to empty body
	}

	return bytes.NewReader(jsonBody)
}
