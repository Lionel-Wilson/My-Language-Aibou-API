package sentence

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/config"
	log "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/log"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/models"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/utils"
	"github.com/gin-gonic/gin"
)

type Service interface {
	GetSentenceExplanation(c *gin.Context, sentence string, nativeLanguage string) (*models.ChatCompletion, error)
	ValidateSentence(sentence string) error
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

func (s *service) ValidateSentence(sentence string) error {
	if sentence == "" {
		s.logger.ErrorLog.Printf("User didn't provide a sentence: %s", sentence)
		return errors.New("Please provide a sentence")
	}

	if utf8.RuneCountInString(sentence) > 100 {
		s.logger.ErrorLog.Printf("Sentence '%s' length too long. Must be less than 100 characters.", sentence)
		return errors.New("The sentence must be less than 100 characters.")
	}

	return nil
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
