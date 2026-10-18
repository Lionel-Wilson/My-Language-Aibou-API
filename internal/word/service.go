package word

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/word/domain"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"unicode/utf8"

	"go.uber.org/zap"

	airesponses "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/airesponses/storage"
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
	responses    airesponses.Repository
}

func NewWordService(
	logger *zap.Logger,
	openAiClient openai.Client,
	responses airesponses.Repository,
) Service {
	return &service{
		logger:       logger,
		openAiClient: openAiClient,
		responses:    responses,
	}
}

// Bump the matching constant when the prompt for that kind changes.
const (
	promptVersionDefinition = 1
	promptVersionSynonyms   = 1
	promptVersionHistory    = 1
	promptVersionContext    = 1
)

func (s *service) getCachedResponse(ctx context.Context, kind airesponses.Kind, input, nativeLanguage string, promptVersion int) (string, bool) {
	response, found, err := s.responses.Get(ctx, kind, input, nativeLanguage, promptVersion)
	if err != nil {
		s.logger.Warn("ai response get failed", zap.String("kind", string(kind)), zap.Error(err))
		return "", false
	}
	return response, found
}

func (s *service) storeCachedResponse(ctx context.Context, kind airesponses.Kind, input, nativeLanguage string, promptVersion int, response string) {
	if err := s.responses.Upsert(ctx, kind, input, nativeLanguage, promptVersion, response); err != nil {
		s.logger.Warn("ai response upsert failed", zap.String("kind", string(kind)), zap.Error(err))
	}
}

type Details struct {
	kind         string
	requestBody  io.Reader
	response     *http.Response
	responseBody []byte
	err          error
}

func (s *service) Lookup(ctx context.Context, word, nativeLanguage string) (*domain.LookupDetails, error) {
	var result domain.LookupDetails

	lookupParts := []struct {
		detailKind     string
		storageKind    airesponses.Kind
		promptVersion  int
		dest           *string
	}{
		{"definition", airesponses.KindDefinition, promptVersionDefinition, &result.Definition},
		{"synonyms", airesponses.KindSynonyms, promptVersionSynonyms, &result.Synonyms},
		{"history", airesponses.KindHistory, promptVersionHistory, &result.History},
		{"context", airesponses.KindContext, promptVersionContext, &result.Context},
	}

	var items []*Details
	for _, part := range lookupParts {
		if cached, ok := s.getCachedResponse(ctx, part.storageKind, word, nativeLanguage, part.promptVersion); ok {
			*part.dest = cached
			continue
		}
		items = append(items, &Details{kind: part.detailKind})
	}

	if len(items) == 0 {
		return &result, nil
	}

	wordLanguage, err := s.detectWordLanguage(ctx, word)
	if err != nil {
		return nil, fmt.Errorf("failed to detect word language: %w", err)
	}

	for _, d := range items {
		var bodyErr error
		switch d.kind {
		case "definition":
			d.requestBody, bodyErr = s.wordToOpenAiDefinitionRequestBody(word, nativeLanguage, wordLanguage)
		case "synonyms":
			d.requestBody, bodyErr = s.wordToOpenAiSynonymsRequestBody(word, nativeLanguage, wordLanguage)
		case "history":
			d.requestBody, bodyErr = s.wordToOpenAiHistoryRequestBody(word, nativeLanguage, wordLanguage)
		case "context":
			d.requestBody, bodyErr = s.wordToOpenAiContextualUsageRequestBody(word, nativeLanguage, wordLanguage)
		}
		if bodyErr != nil {
			return nil, fmt.Errorf("marshal word %s openai request: %w", d.kind, bodyErr)
		}
	}

	g, ctx := errgroup.WithContext(ctx)
	for i := range items {
		d := items[i]
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
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("one or more OpenAI requests failed: %w", err)
	}

	for _, d := range items {
		var comp openai.ChatCompletion
		if err := json.Unmarshal(d.responseBody, &comp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s response: %w", d.kind, err)
		}
		if len(comp.Choices) == 0 {
			return nil, openaierrors.ErrNoChoicesFound
		}
		content := comp.Choices[0].Message.Content
		switch d.kind {
		case "definition":
			result.Definition = content
			s.storeCachedResponse(ctx, airesponses.KindDefinition, word, nativeLanguage, promptVersionDefinition, content)
		case "synonyms":
			result.Synonyms = content
			s.storeCachedResponse(ctx, airesponses.KindSynonyms, word, nativeLanguage, promptVersionSynonyms, content)
		case "history":
			result.History = content
			s.storeCachedResponse(ctx, airesponses.KindHistory, word, nativeLanguage, promptVersionHistory, content)
		case "context":
			result.Context = content
			s.storeCachedResponse(ctx, airesponses.KindContext, word, nativeLanguage, promptVersionContext, content)
		}
	}

	return &result, nil
}
func (s *service) GetWordHistory(ctx context.Context, word string, nativeLanguage string) (*string, error) {
	if cached, ok := s.getCachedResponse(ctx, airesponses.KindHistory, word, nativeLanguage, promptVersionHistory); ok {
		return &cached, nil
	}

	wordLanguage, err := s.detectWordLanguage(ctx, word)
	if err != nil {
		return nil, fmt.Errorf("failed to detect word language: %w", err)
	}

	jsonEncodedBody, err := s.wordToOpenAiHistoryRequestBody(word, nativeLanguage, wordLanguage)
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
		zap.String("nativeLanguage", nativeLanguage),
		zap.Int("promptTokens", OpenAIApiResponse.Usage.PromptTokens),
		zap.Int("completionTokens", OpenAIApiResponse.Usage.CompletionTokens),
		zap.Int("totalTokens", OpenAIApiResponse.Usage.TotalTokens),
	)

	content := OpenAIApiResponse.Choices[0].Message.Content
	s.storeCachedResponse(ctx, airesponses.KindHistory, word, nativeLanguage, promptVersionHistory, content)

	return &content, nil
}

func (s *service) GetWordSynonyms(ctx context.Context, word string, nativeLanguage string) (*string, error) {
	if cached, ok := s.getCachedResponse(ctx, airesponses.KindSynonyms, word, nativeLanguage, promptVersionSynonyms); ok {
		return &cached, nil
	}

	wordLanguage, err := s.detectWordLanguage(ctx, word)
	if err != nil {
		return nil, fmt.Errorf("failed to detect word language: %w", err)
	}

	jsonEncodedBody, err := s.wordToOpenAiSynonymsRequestBody(word, nativeLanguage, wordLanguage)
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
		zap.Int("promptTokens", OpenAIApiResponse.Usage.PromptTokens),
		zap.Int("completionTokens", OpenAIApiResponse.Usage.CompletionTokens),
		zap.Int("totalTokens", OpenAIApiResponse.Usage.TotalTokens),
	)

	result := &OpenAIApiResponse.Choices[0].Message.Content
	s.storeCachedResponse(ctx, airesponses.KindSynonyms, word, nativeLanguage, promptVersionSynonyms, *result)

	return result, nil
}

func (s *service) GetWordDefinition(ctx context.Context, word string, nativeLanguage string) (*string, error) {
	if cached, ok := s.getCachedResponse(ctx, airesponses.KindDefinition, word, nativeLanguage, promptVersionDefinition); ok {
		return &cached, nil
	}

	wordLanguage, err := s.detectWordLanguage(ctx, word)
	if err != nil {
		return nil, fmt.Errorf("failed to detect word language: %w", err)
	}

	jsonEncodedBody, err := s.wordToOpenAiDefinitionRequestBody(word, nativeLanguage, wordLanguage)
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
		zap.String("nativeLanguage", nativeLanguage),
		zap.Int("promptTokens", OpenAIApiResponse.Usage.PromptTokens),
		zap.Int("completionTokens", OpenAIApiResponse.Usage.CompletionTokens),
		zap.Int("totalTokens", OpenAIApiResponse.Usage.TotalTokens),
	)

	s.storeCachedResponse(ctx, airesponses.KindDefinition, word, nativeLanguage, promptVersionDefinition, *result)

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

func (s *service) detectWordLanguage(ctx context.Context, word string) (string, error) {
	wordLanguageRequestBody, err := s.wordToOpenAiWordLanguageRequestBody(word)
	if err != nil {
		return "", fmt.Errorf("marshal word language openai request: %w", err)
	}
	resp, body, err := s.openAiClient.MakeRequest(ctx, wordLanguageRequestBody)
	if err != nil {
		return "", fmt.Errorf("failed to make open ai request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai API returned non-OK status response=%v statusCode=%v", resp, resp.StatusCode)
	}
	var comp openai.ChatCompletion
	err = json.Unmarshal(body, &comp)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal %s response: %w", body, err)
	}
	return comp.Choices[0].Message.Content, nil
}
