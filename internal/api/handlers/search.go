package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/models"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/utils"
	"github.com/gin-gonic/gin"
)

func (app *Application) DefineWord(c *gin.Context) {
	var requestBody models.DefineWordRequest

	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		app.ErrorLog.Println(err.Error())
		utils.ServerErrorResponse(c, err, "")
		return
	}

	word := strings.TrimSpace(requestBody.Word)

	if word == "" {
		app.ErrorLog.Printf("User didn't provide a word: %s", word)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Please provide a word", []string{})
		return
	}

	if utf8.RuneCountInString(word) > 30 {
		app.ErrorLog.Printf("Word '%s' length too long. Must be less than 30 characters.", word)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Word length too long. Must be less than 30 characters.If this is a sentence, please use the analyser.", []string{})
		return
	}

	jsonBody := constructWordDefinitionBody(word, requestBody.Tier, requestBody.NativeLanguage)

	OpenAIApiResponse, err := utils.MakeOpenAIApiRequest(jsonBody, c, *app.OpenApiKey)
	if err != nil {
		app.ErrorLog.Println(err.Error())
		utils.ServerErrorResponse(c, err, "Failed to make request to OpenAI API")
		return
	}

	fmt.Printf("Word Definition: %s\n", OpenAIApiResponse.Choices[0].Message.Content)
	fmt.Printf(`Prompt Tokens: %d`, OpenAIApiResponse.Usage.PromptTokens)
	fmt.Printf(`Response Tokens: %d`, OpenAIApiResponse.Usage.CompletionTokens)
	fmt.Printf(`Total Tokens used: %d`, OpenAIApiResponse.Usage.TotalTokens)

	c.JSON(http.StatusOK, OpenAIApiResponse.Choices[0].Message.Content)
}

func (app *Application) DefinePhrase(c *gin.Context) {
	var requestBody models.DefinePhraseRequest

	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		app.ErrorLog.Println(err.Error())
		utils.ServerErrorResponse(c, err, "")
		return
	}

	phrase := strings.TrimSpace(requestBody.Phrase)

	if phrase == "" {
		app.ErrorLog.Printf("User didn't provide a sentence: %s", phrase)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Please provide a sentence", []string{})
		return
	}

	jsonBody := constructPhraseBody(phrase, requestBody.Tier, requestBody.NativeLanguage)

	OpenAIApiResponse, err := utils.MakeOpenAIApiRequest(jsonBody, c, *app.OpenApiKey)
	if err != nil {
		app.ErrorLog.Println(err.Error())
		utils.ServerErrorResponse(c, err, "Failed to make request to OpenAI API")
		return
	}

	fmt.Printf("Phrase explanation: %s\n", OpenAIApiResponse.Choices[0].Message.Content)
	fmt.Printf("Prompt Tokens: %d\n", OpenAIApiResponse.Usage.PromptTokens)
	fmt.Printf("Response Tokens: %d\n", OpenAIApiResponse.Usage.CompletionTokens)
	fmt.Printf("Total Tokens used: %d\n", OpenAIApiResponse.Usage.TotalTokens)

	c.JSON(http.StatusOK, OpenAIApiResponse.Choices[0].Message.Content)
}

func constructPhraseBody(phrase, userTier, userNativeLanguage string) *strings.Reader {
	//var maxWordCount string
	var MaxTokens string

	if userTier == "Basic" {
		MaxTokens = "110"
		maxWordCount = "80"

	} else if userTier == "Premium" {
		MaxTokens = "330"
		maxWordCount = "230"
	}

	content := fmt.Sprintf("Explain the meaning & grammar used in the following sentence: '%s'.Provide a detailed explanation to help understand the structure,syntax,& semantics of the sentence.Respond in %s & in max %s", phrase, userNativeLanguage,maxWordCount )

	body := fmt.Sprintf(`{
	"model":"gpt-3.5-turbo",
	"messages": [{
		"role": "system",
		"content": "You are a helpful assistant."
	  },
	  {
		"role": "user",
		"content": "%s"
	  }],
	"temperature": 0.4,
	"max_tokens": %s
	}`, content, MaxTokens)

	fmt.Printf("Tier: %s\n", userTier)
	//fmt.Printf("Body: %s\n", body)
	fmt.Printf("Phrase prompt: %s\n", content)

	return strings.NewReader(body)
}

func constructWordDefinitionBody(word, userTier, userNativeLanguage string) *strings.Reader {
	//var maxWordCount string
	var MaxTokens string
	var content string

	if userTier == "Basic" {
		MaxTokens = "75"
		maxWordCount = "40"
		content = fmt.Sprintf("Explain the meaning of '%s', ensuring the explanation is in %s & max %s words", word, userNativeLanguage, maxWordCount)

	} else if userTier == "Premium" {
		MaxTokens = "250"
		//maxWordCount = "180"
		content = fmt.Sprintf("Explain the meaning of '%s', ensuring the explanation is in %s. Provide 3 example sentences using the word '%s', ensuring you translate them into %s", word, userNativeLanguage, word, userNativeLanguage)
	}

	body := fmt.Sprintf(`{
	"model":"gpt-3.5-turbo",
	"messages": [{
		"role": "system",
		"content": "You are a helpful assistant."
	  },
	  {
		"role": "user",
		"content": "%s"
	  }],
	"temperature": 0.4,
	"max_tokens": %s
	}`, content, MaxTokens)

	fmt.Printf("Tier: %s\n", userTier)
	//fmt.Printf("Body: %s\n", body)
	fmt.Printf("Word prompt: %s\n", content)

	return strings.NewReader(body)
}
