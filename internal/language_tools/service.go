package languagetools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"unicode"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/config"
	log "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/log"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/models"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/utils"
	"github.com/gin-gonic/gin"
)

type Service interface {
	IsNonsensical(s string) bool
	IsNotAWord(s string) bool
	GetWordDefinition(c *gin.Context, word string, nativeLanguage string) (*models.ChatCompletion, error)
	GetWordSynonyms(c *gin.Context, word string, nativeLanguage string) (*models.ChatCompletion, error)
	GetSentenceExplanation(c *gin.Context, sentence string, nativeLanguage string) (*models.ChatCompletion, error)
}

type service struct {
	config *config.Config
	logger *log.Logger
}

func New(config *config.Config, logger *log.Logger) Service {
	return &service{
		config: config,
		logger: logger,
	}
}

func (s *service) GetWordSynonyms(c *gin.Context, word string, nativeLanguage string) (*models.ChatCompletion, error) {
	jsonBody := wordToOpenAiSynonymsRequestBody(word, nativeLanguage)

	resp, responseBody, err := utils.MakeOpenAIApiRequest(jsonBody, c, s.config.OpenAi.Key)
	if err != nil {
		s.logger.ErrorLog.Println(err.Error())
		utils.ServerErrorResponse(c, err, "Failed to process your word.Please make sure you remove any extra spaces & special characters and try again")
		return &models.ChatCompletion{}, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("OpenAI API returned non-OK status. ")
		utils.ServerErrorResponse(c, err, "Failed to process your word.Please make sure you remove any extra spaces & special characters and try again")
		return &models.ChatCompletion{}, err
	}

	var OpenAIApiResponse models.ChatCompletion

	err = json.Unmarshal(responseBody, &OpenAIApiResponse)
	if err != nil {
		fmt.Println("Failed to unmarshal json body")
		return &models.ChatCompletion{}, err
	}

	if len(OpenAIApiResponse.Choices) == 0 {
		fmt.Println("OpenAI API response contains no choices")
		err = fmt.Errorf("OpenAI API response contains no choices")
		utils.ServerErrorResponse(c, err, "Failed to process your word.Please make sure you remove any extra spaces & special characters and try again")
		return &models.ChatCompletion{}, err
	}

	fmt.Printf("Word Synonyms: %s\n", OpenAIApiResponse.Choices[0].Message.Content)
	fmt.Printf(`Prompt Tokens: %d`, OpenAIApiResponse.Usage.PromptTokens)
	fmt.Printf(`Response Tokens: %d`, OpenAIApiResponse.Usage.CompletionTokens)
	fmt.Printf(`Total Tokens used: %d`, OpenAIApiResponse.Usage.TotalTokens)

	return &OpenAIApiResponse, nil
}

func (s *service) GetWordDefinition(c *gin.Context, word string, nativeLanguage string) (*models.ChatCompletion, error) {

	jsonBody := wordToOpenAiDefinitionRequestBody(word, nativeLanguage)

	resp, responseBody, err := utils.MakeOpenAIApiRequest(jsonBody, c, s.config.OpenAi.Key)
	if err != nil {
		s.logger.ErrorLog.Println(err.Error())
		return &models.ChatCompletion{}, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("OpenAI API returned non-OK status. ")
		return &models.ChatCompletion{}, err
	}

	var OpenAIApiResponse models.ChatCompletion

	err = json.Unmarshal(responseBody, &OpenAIApiResponse)
	if err != nil {
		fmt.Println("Failed to unmarshal json body")
		return &models.ChatCompletion{}, err
	}

	if len(OpenAIApiResponse.Choices) == 0 {
		fmt.Println("OpenAI API response contains no choices")
		err = fmt.Errorf("OpenAI API response contains no choices")
		return &models.ChatCompletion{}, err
	}

	fmt.Printf("Word Definition: %s\n", OpenAIApiResponse.Choices[0].Message.Content)
	fmt.Printf(`Prompt Tokens: %d`, OpenAIApiResponse.Usage.PromptTokens)
	fmt.Printf(`Response Tokens: %d`, OpenAIApiResponse.Usage.CompletionTokens)
	fmt.Printf(`Total Tokens used: %d`, OpenAIApiResponse.Usage.TotalTokens)

	return &OpenAIApiResponse, nil

}

func (s *service) GetSentenceExplanation(c *gin.Context, sentence string, nativeLanguage string) (*models.ChatCompletion, error) {
	jsonBody := sentenceToOpenAiExplanationRequestBody(sentence, nativeLanguage)

	resp, responseBody, err := utils.MakeOpenAIApiRequest(jsonBody, c, s.config.OpenAi.Key)
	if err != nil {
		s.logger.ErrorLog.Println(err.Error())
		utils.ServerErrorResponse(c, err, "Failed to process your sentence(s).Please make sure you remove any line breaks and large gaps between your sentences and try again")
		return &models.ChatCompletion{}, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("OpenAI API returned non-OK status. ")
		utils.ServerErrorResponse(c, err, "Failed to process your sentence(s).Please make sure you remove any line breaks and large gaps between your sentences and try again")
		return &models.ChatCompletion{}, err
	}

	var OpenAIApiResponse models.ChatCompletion

	err = json.Unmarshal(responseBody, &OpenAIApiResponse)
	if err != nil {
		fmt.Println("Failed to unmarshal json body")
		return &models.ChatCompletion{}, err
	}

	if len(OpenAIApiResponse.Choices) == 0 {
		fmt.Println("OpenAI API response contains no choices")
		err = fmt.Errorf("OpenAI API response contains no choices")
		utils.ServerErrorResponse(c, err, "Failed to process your sentence(s).Please make sure you remove any line breaks and large gaps between your sentences and try again")
		return &models.ChatCompletion{}, err
	}

	fmt.Printf("Phrase explanation: %s\n", OpenAIApiResponse.Choices[0].Message.Content)
	fmt.Printf("Prompt Tokens: %d\n", OpenAIApiResponse.Usage.PromptTokens)
	fmt.Printf("Response Tokens: %d\n", OpenAIApiResponse.Usage.CompletionTokens)
	fmt.Printf("Total Tokens used: %d\n", OpenAIApiResponse.Usage.TotalTokens)

	return &OpenAIApiResponse, nil
}

// isNotAWord is used to check if the user is using the dictionary to define phrases as opposed to a single word
func (s *service) IsNotAWord(word string) bool {
	return strings.Count(word, " ") > 1
}

func (s *service) IsNonsensical(word string) bool {
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

func sentenceToOpenAiExplanationRequestBody(sentence, userNativeLanguage string) *strings.Reader {
	//var maxWordCount string
	//var MaxTokens string

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

	//fmt.Printf("Tier: %s\n", userTier)
	//fmt.Printf("Body: %s\n", body)
	fmt.Printf("Phrase prompt: %s\n", content)

	return strings.NewReader(body)
}
