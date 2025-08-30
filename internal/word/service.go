package word

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/word/domain"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/request"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/coocood/freecache"
	"go.uber.org/zap"

	openai "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/clients/open-ai"
	openaierrors "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/clients/open-ai/errors"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/utils"
)

//go:generate mockgen -source=service.go -destination=mock/service.go
type Service interface {
	GetWordDefinition(ctx context.Context, word string, nativeLanguage string) (*string, error)
	GetWordSynonyms(ctx context.Context, word string, nativeLanguage string) (*string, error)
	ValidateWord(word string) error
	GetWordHistory(ctx context.Context, word string, nativeLanguage string) (*string, error)
	Lookup(ctx context.Context, word string, nativeLanguage string) (*domain.LookupDetails, error)
}

type service struct {
	logger       *zap.Logger
	openAiClient openai.Client
	cache        *freecache.Cache
}

func NewWordService(
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

var wordCacheExpiration = int(time.Hour * 24 * 90) // 90 days

type Details struct {
	kind         string
	requestBody  io.Reader
	response     *http.Response
	responseBody []byte
	err          error
}

func (s *service) Lookup(ctx context.Context, word, nativeLanguage string) (*domain.LookupDetails, error) {
	// 1) Build payloads sequentially (no races)
	defBody, err := s.wordToOpenAiDefinitionRequestBody(word, nativeLanguage)
	if err != nil {
		return nil, fmt.Errorf("marshal word definition openai request: %w", err)
	}
	synBody, err := s.wordToOpenAiSynonymsRequestBody(word, nativeLanguage)
	if err != nil {
		return nil, fmt.Errorf("marshal word synonyms openai request: %w", err)
	}
	histBody, err := s.wordToOpenAiHistoryRequestBody(word, nativeLanguage)
	if err != nil {
		return nil, fmt.Errorf("marshal word history openai request: %w", err)
	}

	items := []*Details{
		{kind: "definition", requestBody: defBody},
		{kind: "synonyms", requestBody: synBody},
		{kind: "history", requestBody: histBody},
	}

	// 2) Fire requests in parallel with errgroup (ctx-aware)
	g, ctx := errgroup.WithContext(ctx)
	for i := range items {
		d := items[i] // capture pointer
		g.Go(func() error {
			resp, body, reqErr := s.openAiClient.MakeRequest(ctx, d.requestBody)
			d.response, d.responseBody, d.err = resp, body, reqErr
			if reqErr != nil {
				s.logger.Error("failed to make openai request", zap.String("kind", d.kind), zap.Error(reqErr))
				return reqErr
			}
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("openai non-OK: kind=%s status=%d body=%s", d.kind, resp.StatusCode, string(body))
			}
			return nil
		})
	}
	err = g.Wait()
	if err != nil {
		return nil, fmt.Errorf("one or more OpenAI requests failed: %w", err)
	}

	// 3) Unmarshal each response and assemble result
	var result domain.LookupDetails
	for _, d := range items {
		var comp openai.ChatCompletion
		err = json.Unmarshal(d.responseBody, &comp)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s response: %w", d.kind, err)
		}
		if len(comp.Choices) == 0 {
			return nil, openaierrors.ErrNoChoicesFound
		}
		content := comp.Choices[0].Message.Content
		switch d.kind {
		case "definition":
			result.Definition = content
		case "synonyms":
			result.Synonyms = content
		case "history":
			result.History = content
		}
	}

	return &result, nil
}
func (s *service) GetWordHistory(ctx context.Context, word string, nativeLanguage string) (*string, error) {
	cacheKey := []byte(fmt.Sprintf("%s word history in %s", word, nativeLanguage))

	cached, err := s.cache.Get(cacheKey)
	if err == nil {
		cachedResponse := string(cached)
		return &cachedResponse, nil
	}

	jsonEncodedBody, err := s.wordToOpenAiHistoryRequestBody(word, nativeLanguage)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal openai request: %w", err)
	}

	resp, responseBody, err := s.openAiClient.MakeRequest(ctx, jsonEncodedBody)
	if err != nil {
		return nil, fmt.Errorf("failed to make open ai request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai API returned non-OK status responseBody=%v statusCode=%v", resp, resp.StatusCode)
	}

	var OpenAIApiResponse openai.ChatCompletion

	err = json.Unmarshal(responseBody, &OpenAIApiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarhsal json body: %w", err)
	}

	if len(OpenAIApiResponse.Choices) == 0 {
		return nil, openaierrors.ErrNoChoicesFound
	}

	s.logger.Info("Successfully got word history",
		zap.String("word", word),
		zap.Any("jsonEncodedBody", jsonEncodedBody),
		zap.String("nativeLanguage", nativeLanguage),
		zap.Int("promptTokens", OpenAIApiResponse.Usage.PromptTokens),
		zap.Int("completionTokens", OpenAIApiResponse.Usage.CompletionTokens),
		zap.Int("totalTokens", OpenAIApiResponse.Usage.TotalTokens),
	)

	cacheValue := []byte(OpenAIApiResponse.Choices[0].Message.Content)

	err = s.cache.Set(cacheKey, cacheValue, wordCacheExpiration)
	if err != nil {
		s.logger.Warn("word history cache set failed", zap.Error(err))
	}

	return &OpenAIApiResponse.Choices[0].Message.Content, nil
}

func (s *service) GetWordSynonyms(ctx context.Context, word string, nativeLanguage string) (*string, error) {
	cacheKey := []byte(fmt.Sprintf("%s word synonyms in %s", word, nativeLanguage))

	cached, err := s.cache.Get(cacheKey)
	if err == nil {
		cachedResponse := string(cached)
		return &cachedResponse, nil
	}

	jsonEncodedBody, err := s.wordToOpenAiSynonymsRequestBody(word, nativeLanguage)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal openai request: %w", err)
	}

	resp, responseBody, err := s.openAiClient.MakeRequest(ctx, jsonEncodedBody)
	if err != nil {
		return nil, fmt.Errorf("failed to make open ai request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai API returned non-OK status response=%v statusCode=%v", resp, resp.StatusCode)
	}

	var OpenAIApiResponse openai.ChatCompletion

	err = json.Unmarshal(responseBody, &OpenAIApiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarhsal json body: %w", err)
	}

	if len(OpenAIApiResponse.Choices) == 0 {
		return nil, openaierrors.ErrNoChoicesFound
	}

	s.logger.Info("Successfully got word synonyms",
		zap.String("word", word),
		zap.String("nativeLanguage", nativeLanguage),
		zap.Any("jsonEncodedBody", jsonEncodedBody),
		zap.Int("promptTokens", OpenAIApiResponse.Usage.PromptTokens),
		zap.Int("completionTokens", OpenAIApiResponse.Usage.CompletionTokens),
		zap.Int("totalTokens", OpenAIApiResponse.Usage.TotalTokens),
	)

	result := &OpenAIApiResponse.Choices[0].Message.Content

	cacheValue := []byte(*result)

	err = s.cache.Set(cacheKey, cacheValue, wordCacheExpiration)
	if err != nil {
		s.logger.Warn("word synonyms cache set failed", zap.Error(err))
	}

	return result, nil
}

func (s *service) GetWordDefinition(ctx context.Context, word string, nativeLanguage string) (*string, error) {
	cacheKey := []byte(fmt.Sprintf("%s word definition in %s", word, nativeLanguage))

	cached, err := s.cache.Get(cacheKey)
	if err == nil {
		cachedResponse := string(cached)
		return &cachedResponse, nil
	}

	jsonEncodedBody, err := s.wordToOpenAiDefinitionRequestBody(word, nativeLanguage)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal openai request: %w", err)
	}

	resp, responseBody, err := s.openAiClient.MakeRequest(ctx, jsonEncodedBody)
	if err != nil {
		return nil, fmt.Errorf("failed to make open ai request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai api returned non-OK status response=%v statusCode=%v", resp, resp.StatusCode)
	}

	var OpenAIApiResponse openai.ChatCompletion

	err = json.Unmarshal(responseBody, &OpenAIApiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarhsal json body responseBody=%v : %w", responseBody, err)
	}

	if len(OpenAIApiResponse.Choices) == 0 {
		return nil, openaierrors.ErrNoChoicesFound
	}

	result := &OpenAIApiResponse.Choices[0].Message.Content

	s.logger.Info("Successfully got word definition.",
		zap.String("word", word),
		zap.Any("jsonEncodedBody", jsonEncodedBody),
		zap.String("nativeLanguage", nativeLanguage),
		zap.Int("promptTokens", OpenAIApiResponse.Usage.PromptTokens),
		zap.Int("completionTokens", OpenAIApiResponse.Usage.CompletionTokens),
		zap.Int("totalTokens", OpenAIApiResponse.Usage.TotalTokens),
	)

	cacheValue := []byte(*result)

	err = s.cache.Set(cacheKey, cacheValue, wordCacheExpiration)
	if err != nil {
		s.logger.Warn("word definition cache set failed", zap.Error(err))
	}

	return result, nil
}

// todo: return actual errors and then convert to userf friendly on handler layer
func (s *service) ValidateWord(word string) error {
	if word == "" {
		return errors.New("please provide a word")
	}

	if utils.ContainsNumber(word) {
		return errors.New("words should not contain numbers")
	}

	if utf8.RuneCountInString(word) > 30 {
		return errors.New("word length too long. Must be less than 30 characters.If this is a sentence, please use the analyser")
	}

	if isNotAWord(word) {
		return errors.New("this looks like a phrase. Please use the 'Analyzer'")
	}

	if isNonsensical(word) {
		return errors.New("this doesn't look like a word. Please provide a valid word")
	}

	return nil
}

func (s *service) wordToOpenAiHistoryRequestBody(word, userNativeLanguage string) (*bytes.Reader, error) {
	content := fmt.Sprintf(
		"Give me the history and origin of the word '%s', ensuring the explanation is in %s. "+
			"(If the word is Japanese, include furigana for any kanji used, but do not mention whether it is or isn’t Japanese.)",
		word, userNativeLanguage,
	)

	s.logger.Info("wordToOpenAiHistoryRequestBody", zap.String("content", content))

	return request.JsonReader(mapToOpenAiRequest(content))
}

func (s *service) wordToOpenAiDefinitionRequestBody(word, lang string) (*bytes.Reader, error) {
	content := fmt.Sprintf(
		"Explain the meaning of '%s' in %s. Provide 2 example sentences using the word '%s', with translations into %s. "+
			"If the word is Japanese, include furigana for any kanji used, but do not mention whether it is or isn’t Japanese.",
		word, lang, word, lang,
	)

	s.logger.Info("wordToOpenAiDefinitionRequestBody", zap.String("content", content))

	return request.JsonReader(mapToOpenAiRequest(content))
}

func (s *service) wordToOpenAiSynonymsRequestBody(word, userNativeLanguage string) (*bytes.Reader, error) {
	content := fmt.Sprintf(
		"The user has provided the word '%s'. First, detect what language this word is in. "+
			"Then, list some simple synonyms for it in that same language. "+
			"Respond in %s, but make sure the synonyms themselves are written in the original language of the word.",
		word, userNativeLanguage,
	)

	s.logger.Info("wordToOpenAiSynonymsRequestBody", zap.String("content", content))

	return request.JsonReader(mapToOpenAiRequest(content))
}

// isNotAWord is used to check if the user is using the dictionary to define phrases as opposed to a single word
func isNotAWord(word string) bool {
	return strings.Count(word, " ") > 1
}

func isNonsensical(s string) bool {
	// specials: punctuation or symbol (keep hyphen/apostrophe if you want)
	for _, r := range s {
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			if r != '-' && r != '\'' {
				return true
			}
		}
	}
	// repeating characters (runes)
	var prev rune

	count := 0

	for i, r := range s {
		if i == 0 || r != prev {
			prev, count = r, 1
			continue
		}

		count++
		if count > 3 {
			return true
		}
	}

	return false
}

func mapToOpenAiRequest(content string) *openai.OpenAIRequest {
	response := openai.OpenAIRequest{
		Model:       "gpt-4o",
		Temperature: 0.4,
		MaxTokens:   400,
	}
	response.Messages = append(response.Messages,
		openai.Message{Role: "system", Content: "You are a helpful multilingual assistant that supports users learning foreign languages."},
		openai.Message{Role: "user", Content: content},
	)

	return &response
}
