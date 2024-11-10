package word

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	log "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/log"
	openai "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/clients/open-ai"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/utils"
	"github.com/gin-gonic/gin"
)

//go:generate mockgen -source=service.go -destination=mock/service.go
type Service interface {
	GetWordDefinition(c *gin.Context, word string, nativeLanguage string) (*openai.ChatCompletion, error)
	GetWordSynonyms(c *gin.Context, word string, nativeLanguage string) (*openai.ChatCompletion, error)
	ValidateWord(word string) error
}

type service struct {
	logger       log.Logger
	openAiClient openai.Client
}

func New(logger log.Logger, openAiClient openai.Client) Service {
	return &service{
		logger:       logger,
		openAiClient: openAiClient,
	}
}

func (s *service) GetWordSynonyms(c *gin.Context, word string, nativeLanguage string) (*openai.ChatCompletion, error) {
	jsonBody := wordToOpenAiSynonymsRequestBody(word, nativeLanguage)

	resp, responseBody, err := s.openAiClient.MakeRequest(jsonBody, c)
	if err != nil {
		s.logger.Error(err.Error())
		utils.ServerErrorResponse(c, err, "Failed to process your word.Please make sure you remove any extra spaces & special characters and try again")
		return &openai.ChatCompletion{}, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("OpenAI API returned non-OK status. ")
		utils.ServerErrorResponse(c, err, "Failed to process your word.Please make sure you remove any extra spaces & special characters and try again")
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
		err = fmt.Errorf("OpenAI API response contains no choices")
		utils.ServerErrorResponse(c, err, "Failed to process your word.Please make sure you remove any extra spaces & special characters and try again")
		return &openai.ChatCompletion{}, err
	}

	fmt.Printf("Word Synonyms: %s\n", OpenAIApiResponse.Choices[0].Message.Content)
	fmt.Printf(`Prompt Tokens: %d`, OpenAIApiResponse.Usage.PromptTokens)
	fmt.Printf(`Response Tokens: %d`, OpenAIApiResponse.Usage.CompletionTokens)
	fmt.Printf(`Total Tokens used: %d`, OpenAIApiResponse.Usage.TotalTokens)

	return &OpenAIApiResponse, nil
}

func (s *service) GetWordDefinition(c *gin.Context, word string, nativeLanguage string) (*openai.ChatCompletion, error) {

	jsonBody := wordToOpenAiDefinitionRequestBody(word, nativeLanguage)

	resp, responseBody, err := s.openAiClient.MakeRequest(jsonBody, c)
	if err != nil {
		s.logger.Error(err.Error())
		return &openai.ChatCompletion{}, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("OpenAI API returned non-OK status. ")
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
		err = fmt.Errorf("OpenAI API response contains no choices")
		return &openai.ChatCompletion{}, err
	}

	fmt.Printf("response: %v\n", OpenAIApiResponse)
	fmt.Printf("Word Definition: %s\n", OpenAIApiResponse.Choices[0].Message.Content)
	fmt.Printf(`Prompt Tokens: %d\n`, OpenAIApiResponse.Usage.PromptTokens)
	fmt.Printf(`Response Tokens: %d\n`, OpenAIApiResponse.Usage.CompletionTokens)
	fmt.Printf(`Total Tokens used: %d\n`, OpenAIApiResponse.Usage.TotalTokens)

	return &OpenAIApiResponse, nil
}

func (s *service) ValidateWord(word string) error {
	if word == "" {
		s.logger.Error(fmt.Printf("User didn't provide a word: %s", word))
		return errors.New("Please provide a word")
	}
	if utils.ContainsNumber(word) {
		s.logger.Error(fmt.Printf("User provided a word(%s) that contained a number.", word))
		return errors.New("Words should not contain numbers.")
	}
	if utf8.RuneCountInString(word) > 30 {
		s.logger.Error(fmt.Printf("Word '%s' length too long. Must be less than 30 characters.", word))
		return errors.New("Word length too long. Must be less than 30 characters.If this is a sentence, please use the analyser.")
	}
	if isNotAWord(word) {
		s.logger.Error(fmt.Printf("User provided a phrase(%s) instead of a word.", word))
		return errors.New("This looks like a phrase. Please use the 'Analyzer'.")
	}
	if isNonsensical(word) {
		s.logger.Error(fmt.Printf("User provided nonsense(%s) instead of a word.", word))
		return errors.New("This doesn't look like a word. Please provide a valid word.")
	}

	return nil
}

func wordToOpenAiDefinitionRequestBody(word, userNativeLanguage string) *strings.Reader {
	//var maxWordCount string
	//var MaxTokens string

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

	//fmt.Printf("Tier: %s\n", userTier)
	//fmt.Printf("Body: %s\n", body)
	fmt.Printf("Word prompt: %s\n", content)

	return strings.NewReader(body)
}

func wordToOpenAiSynonymsRequestBody(word, userNativeLanguage string) *strings.Reader {
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

	fmt.Printf("Word prompt: %s\n", content)

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
