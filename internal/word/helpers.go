package word

import (
	"bytes"
	"fmt"
	openai "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/clients/open-ai"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/request"
	"strings"
	"unicode"
)

func (s *service) wordToOpenAiHistoryRequestBody(word, userNativeLanguage, wordLanguage string) (*bytes.Reader, error) {
	content := fmt.Sprintf(
		"Give me the history and origin of the %s word '%s', ensuring the explanation is in %s. "+
			"(If the word is Japanese, include furigana for any kanji used, but do not mention whether it is or isn’t Japanese.)",
		wordLanguage, word, userNativeLanguage,
	)

	return request.JsonReader(mapToOpenAiRequest(content))
}

func (s *service) wordToOpenAiContextualUsageRequestBody(word, userNativeLanguage, wordLanguage string) (*bytes.Reader, error) {
	content := fmt.Sprintf(
		"Explain the contextual usage of the %s word '%s'. e.g. whether it's formal/casual, who would say this and to whom, when would you say this etc. Make sure to respond in %s.",
		wordLanguage, word, userNativeLanguage,
	)

	return request.JsonReader(mapToOpenAiRequest(content))
}

func (s *service) wordToOpenAiWordLanguageRequestBody(word string) (*bytes.Reader, error) {
	content := fmt.Sprintf(
		"what language is this word '%s'. reply with just the language name.",
		word,
	)

	return request.JsonReader(mapToOpenAiRequest(content))
}

func (s *service) wordToOpenAiDefinitionRequestBody(word, userNativeLanguage, wordLanguage string) (*bytes.Reader, error) {
	content := fmt.Sprintf(
		"Explain the meaning of the %s word '%s'. Provide 2 example sentences using the word '%s', with translations into %s.Make sure to respond in %s.",
		wordLanguage, word, word, userNativeLanguage, userNativeLanguage,
	)

	return request.JsonReader(mapToOpenAiRequest(content))
}

func (s *service) wordToOpenAiSynonymsRequestBody(word, userNativeLanguage, wordLanguage string) (*bytes.Reader, error) {
	content := fmt.Sprintf(
		"The user has provided the %s word '%s'."+"List some simple synonyms for it in %s. "+
			"Respond in %s, but make sure the synonyms themselves are written in %s",
		wordLanguage, word, wordLanguage, userNativeLanguage, wordLanguage,
	)

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
