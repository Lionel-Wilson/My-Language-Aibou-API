package word

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	"go.uber.org/zap"

	openai "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/clients/open-ai"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/utils"
)

var FailedToProcessWord = "Failed to process your word.Please make sure you remove any extra spaces & special characters and try again"

//go:generate mockgen -source=service.go -destination=mock/service.go
type Service interface {
	GetWordDefinition(word string, nativeLanguage string) (*openai.ChatCompletion, error)
	GetWordSynonyms(word string, nativeLanguage string) (*openai.ChatCompletion, error)
	ValidateWord(word string) error
}

type service struct {
	logger       *zap.Logger
	openAiClient openai.Client
}

func NewWordService(logger *zap.Logger, openAiClient openai.Client) Service {
	return &service{
		logger:       logger,
		openAiClient: openAiClient,
	}
}

func (s *service) GetWordSynonyms(word string, nativeLanguage string) (*openai.ChatCompletion, error) {
	jsonBody := s.wordToOpenAiSynonymsRequestBody(word, nativeLanguage)

	resp, responseBody, err := s.openAiClient.MakeRequest(jsonBody)
	if err != nil {
		s.logger.Error(err.Error())

		return &openai.ChatCompletion{}, err
	}

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("OpenAI API returned non-OK status. ")

		return &openai.ChatCompletion{}, err
	}

	var OpenAIApiResponse openai.ChatCompletion

	err = json.Unmarshal(responseBody, &OpenAIApiResponse)
	if err != nil {
		s.logger.Info("Failed to unmarshal json body")
		return &openai.ChatCompletion{}, err
	}

	if len(OpenAIApiResponse.Choices) == 0 {
		s.logger.Info("OpenAI API response contains no choices")

		return &openai.ChatCompletion{}, fmt.Errorf("OpenAI API response contains no choices")
	}

	s.logger.Sugar().Infof("Word Synonyms: %s\n", OpenAIApiResponse.Choices[0].Message.Content)
	s.logger.Sugar().Infof(`Prompt Tokens: %d`, OpenAIApiResponse.Usage.PromptTokens)
	s.logger.Sugar().Infof(`Response Tokens: %d`, OpenAIApiResponse.Usage.CompletionTokens)
	s.logger.Sugar().Infof(`Total Tokens used: %d`, OpenAIApiResponse.Usage.TotalTokens)

	return &OpenAIApiResponse, nil
}

func (s *service) GetWordDefinition(word string, nativeLanguage string) (*openai.ChatCompletion, error) {
	jsonBody := s.wordToOpenAiDefinitionRequestBody(word, nativeLanguage)

	resp, responseBody, err := s.openAiClient.MakeRequest(jsonBody)
	if err != nil {
		s.logger.Error(err.Error())
		return &openai.ChatCompletion{}, err
	}

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("OpenAI API returned non-OK status. ")
		return &openai.ChatCompletion{}, err
	}

	var OpenAIApiResponse openai.ChatCompletion

	err = json.Unmarshal(responseBody, &OpenAIApiResponse)
	if err != nil {
		s.logger.With(zap.Error(err)).Error("Failed to unmarshal json body")
		return &openai.ChatCompletion{}, err
	}

	if len(OpenAIApiResponse.Choices) == 0 {
		s.logger.Info("OpenAI API response contains no choices")

		return &openai.ChatCompletion{}, fmt.Errorf("OpenAI API response contains no choices")
	}

	s.logger.Sugar().Infof("response: %v\n", OpenAIApiResponse)
	s.logger.Sugar().Infof("Word Definition: %s\n", OpenAIApiResponse.Choices[0].Message.Content)
	s.logger.Sugar().Infof(`Prompt Tokens: %d\n`, OpenAIApiResponse.Usage.PromptTokens)
	s.logger.Sugar().Infof(`Response Tokens: %d\n`, OpenAIApiResponse.Usage.CompletionTokens)
	s.logger.Sugar().Infof(`Total Tokens used: %d\n`, OpenAIApiResponse.Usage.TotalTokens)

	return &OpenAIApiResponse, nil
}

func (s *service) ValidateWord(word string) error {
	if word == "" {
		s.logger.Sugar().Errorf("User didn't provide a word: %s", word)
		return errors.New("please provide a word")
	}

	if utils.ContainsNumber(word) {
		s.logger.Sugar().Errorf("User provided a word(%s) that contained a number.", word)
		return errors.New("words should not contain numbers")
	}

	if utf8.RuneCountInString(word) > 30 {
		s.logger.Sugar().Errorf("Word '%s' length too long. Must be less than 30 characters.", word)
		return errors.New("word length too long. Must be less than 30 characters.If this is a sentence, please use the analyser")
	}

	if isNotAWord(word) {
		s.logger.Sugar().Errorf("User provided a phrase(%s) instead of a word.", word)
		return errors.New("this looks like a phrase. Please use the 'Analyzer'")
	}

	if isNonsensical(word) {
		s.logger.Sugar().Errorf("User provided nonsense(%s) instead of a word.", word)
		return errors.New("this doesn't look like a word. Please provide a valid word")
	}

	return nil
}

func (s *service) wordToOpenAiDefinitionRequestBody(word, userNativeLanguage string) *strings.Reader {
	// var maxWordCount string
	// var MaxTokens string
	/*if userTier == "Basic" {
		MaxTokens = "75"
		maxWordCount = "40"
		content = fmt.Sprintf("Explain the meaning of '%s', ensuring the explanation is in %s & max %s words", word, userNativeLanguage, maxWordCount)

	} else if userTier == "Premium" {
		MaxTokens = "250"
		//maxWordCount = "180"

	}*/
	content := fmt.Sprintf("Explain the meaning of '%s'ensuring the explanation is in %s. Provide 2 example sentences using the word '%s', ensuring you translate them into %s", word, userNativeLanguage, word, userNativeLanguage)

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
	"max_tokens": 300
	}`, content)

	// s.logger.Sugar().Infof("Tier: %s\n", userTier)
	// s.logger.Sugar().Infof("Body: %s\n", body)
	s.logger.Sugar().Infof("Word prompt: %s\n", content)

	return strings.NewReader(body)
}

func (s *service) wordToOpenAiSynonymsRequestBody(word, userNativeLanguage string) *strings.Reader {
	content := fmt.Sprintf("List out for me some simple synonyms for the word '%s'", word)

	if userNativeLanguage != "English" {
		content = fmt.Sprintf("List out for me some simple synonyms for the word '%s'. Respond in %s", word, userNativeLanguage)
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
	"max_tokens": 300
	}`, content)

	s.logger.Sugar().Infof("Word prompt: %s\n", content)

	return strings.NewReader(body)
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
