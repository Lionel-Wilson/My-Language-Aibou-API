package word

import (
	"encoding/json"
	"errors"
	"fmt"
	openaierrors "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/clients/open-ai/errors"
	"net/http"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/coocood/freecache"
	"go.uber.org/zap"

	openai "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/clients/open-ai"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/utils"
)

//go:generate mockgen -source=service.go -destination=mock/service.go
type Service interface {
	GetWordDefinition(word string, nativeLanguage string) (*string, error)
	GetWordSynonyms(word string, nativeLanguage string) (*string, error)
	ValidateWord(word string) error
	GetWordHistory(word string, nativeLanguage string) (*string, error)
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

var (
	wordCacheExpiration = int(time.Hour * 24 * 90) //90 days
)

func (s *service) GetWordHistory(word string, nativeLanguage string) (*string, error) {
	cacheKey := []byte(fmt.Sprintf("%s word history in %s", word, nativeLanguage))

	cached, err := s.cache.Get(cacheKey)
	if err == nil {
		cachedResponse := string(cached)
		return &cachedResponse, nil
	}

	jsonBody := s.wordToOpenAiHistoryRequestBody(word, nativeLanguage)

	resp, responseBody, err := s.openAiClient.MakeRequest(jsonBody)
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

	cacheValue := []byte(OpenAIApiResponse.Choices[0].Message.Content)
	err = s.cache.Set(cacheKey, cacheValue, wordCacheExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to cache word definition: %w", err)
	}

	return &OpenAIApiResponse.Choices[0].Message.Content, nil
}

func (s *service) GetWordSynonyms(word string, nativeLanguage string) (*string, error) {
	cacheKey := []byte(fmt.Sprintf("%s word synonyms in %s", word, nativeLanguage))

	cached, err := s.cache.Get(cacheKey)
	if err == nil {
		cachedResponse := string(cached)
		return &cachedResponse, nil
	}

	jsonEncodedBody := s.wordToOpenAiSynonymsRequestBody(word, nativeLanguage)

	resp, responseBody, err := s.openAiClient.MakeRequest(jsonEncodedBody)
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

	result := &OpenAIApiResponse.Choices[0].Message.Content

	cacheValue := []byte(*result)
	err = s.cache.Set(cacheKey, cacheValue, wordCacheExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to cache word definition: %w", err)
	}

	return result, nil
}

func (s *service) GetWordDefinition(word string, nativeLanguage string) (*string, error) {
	cacheKey := []byte(fmt.Sprintf("%s word definition in %s", word, nativeLanguage))

	cached, err := s.cache.Get(cacheKey)
	if err == nil {
		cachedResponse := string(cached)
		return &cachedResponse, nil
	}

	jsonEncodedBody := s.wordToOpenAiDefinitionRequestBody(word, nativeLanguage)

	resp, responseBody, err := s.openAiClient.MakeRequest(jsonEncodedBody)
	if err != nil {
		return nil, fmt.Errorf("failed to make open ai request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai api returned non-OK status responseBody=%v statusCode=%v", resp, resp.StatusCode)
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

	cacheValue := []byte(*result)
	err = s.cache.Set(cacheKey, cacheValue, wordCacheExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to cache word definition: %w", err)
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

func (s *service) wordToOpenAiHistoryRequestBody(word, userNativeLanguage string) *strings.Reader {
	content := fmt.Sprintf("Give me the history and origin of the word '%s', ensuring the explanation is in %s. (If the word is Japanese, include furigana for any kanji used, but do not mention whether it is or isn’t Japanese.)", word, userNativeLanguage)

	return strings.NewReader(fmt.Sprintf(`{
	"model":"gpt-4o",
	"messages": [{
		"role": "system",
		"content": "You are a helpful multilingual assistant that supports users learning foreign languages."
	  },
	  {
		"role": "user",
		"content": "%s"
	  }],
	"temperature": 0.4,
	"max_tokens": 300
	}`, content))
}

func (s *service) wordToOpenAiDefinitionRequestBody(word, lang string) *strings.Reader {
	content := fmt.Sprintf(
		"Explain the meaning of '%s' in %s. Provide 2 example sentences using the word '%s', with translations into %s. "+
			"If the word is Japanese, include furigana for any kanji used, but do not mention whether it is or isn’t Japanese.",
		word, lang, word, lang,
	)

	req := openai.OpenAIRequest{
		Model:       "gpt-4o",
		Temperature: 0.4,
		MaxTokens:   300,
	}
	req.Messages = append(req.Messages,
		openai.Message{Role: "system", Content: "You are a helpful multilingual assistant that supports users learning foreign languages."},
		openai.Message{Role: "user", Content: content},
	)
	b, _ := json.Marshal(&req)
	return strings.NewReader(string(b))
}

func (s *service) wordToOpenAiSynonymsRequestBody(word, userNativeLanguage string) *strings.Reader {
	// Construct the dynamic content prompt
	content := fmt.Sprintf(
		"The user has provided the word '%s'. First, detect what language this word is in. "+
			"Then, list some simple synonyms for it in that same language. "+
			"Respond in %s, but make sure the synonyms themselves are written in the original language of the word.",
		word, userNativeLanguage,
	)

	// Optionally add a more specific system instruction
	systemPrompt := "You are a helpful multilingual assistant that supports users learning foreign languages."

	// Construct the full request body
	return strings.NewReader(fmt.Sprintf(`{
		"model": "gpt-4o",
		"messages": [
			{
				"role": "system",
				"content": "%s"
			},
			{
				"role": "user",
				"content": "%s"
			}
		],
		"temperature": 0.4,
		"max_tokens": 400
	}`, systemPrompt, content))
}

// isNotAWord is used to check if the user is using the dictionary to define phrases as opposed to a single word
func isNotAWord(word string) bool {
	return strings.Count(word, " ") > 1
}

func isNonsensical(word string) bool {
	// Check condition 1: The string contains special characters
	hasSpecialCharacters := false

	for _, ch := range word {
		if unicode.IsPunct(ch) || unicode.IsSymbol(ch) {
			hasSpecialCharacters = true
			break
		}
	}

	// Check condition 2: The string contains the same character more than 3 times in a row
	hasRepeatingCharacters := false

	for i := 0; i < len(word)-3; i++ {
		if word[i] == word[i+1] && word[i] == word[i+2] && word[i] == word[i+3] {
			hasRepeatingCharacters = true
			break
		}
	}

	// Combine all conditions to determine if the string is nonsensical
	return hasSpecialCharacters || hasRepeatingCharacters
}
