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

	if utf8.RuneCountInString(requestBody.Word) > 20 {
		app.ErrorLog.Printf(`Word ""%s length too long. Must be less than 20 characters.`, requestBody.Word)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Word length too long. Must be less than 20 characters. Could this be a phrase?", []string{})
		return
	}

	jsonBody := constructWordDefinitionBody(requestBody.Word, requestBody.Tier, requestBody.NativeLanguage)

	OpenAIApiResponse, err := utils.MakeOpenAIApiRequest(jsonBody, c, *app.OpenApiKey)
	if err != nil {
		app.ErrorLog.Println(err.Error())
		utils.ServerErrorResponse(c, err, "Failed to make request to OpenAI API")
		return
	}

	fmt.Println(OpenAIApiResponse)
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

	jsonBody := constructPhraseBody(requestBody.Phrase, requestBody.Tier, requestBody.TargetLanguage, requestBody.NativeLanguage)

	OpenAIApiResponse, err := utils.MakeOpenAIApiRequest(jsonBody, c, *app.OpenApiKey)
	if err != nil {
		app.ErrorLog.Println(err.Error())
		utils.ServerErrorResponse(c, err, "Failed to make request to OpenAI API")
		return
	}

	fmt.Println(OpenAIApiResponse)
	fmt.Printf("Prompt Tokens: %d\n", OpenAIApiResponse.Usage.PromptTokens)
	fmt.Printf("Response Tokens: %d\n", OpenAIApiResponse.Usage.CompletionTokens)
	fmt.Printf("Total Tokens used: %d\n", OpenAIApiResponse.Usage.TotalTokens)

	c.JSON(http.StatusOK, OpenAIApiResponse.Choices[0].Message.Content)
}

func constructPhraseBody(phrase, userTier, userTargetLanguage, userNativeLanguage string) *strings.Reader {
	var maxWordCount string
	var MaxTokens string

	if userTier == "Basic" {
		MaxTokens = "110"
		maxWordCount = "80"

	} else if userTier == "Premium" {
		MaxTokens = "330"
		maxWordCount = "200"
	}

	content := fmt.Sprintf("Explain the meaning & grammar used in this %s sentence in max %s words.Respond in %s-%s", userTargetLanguage, maxWordCount, userNativeLanguage, phrase)

	body := fmt.Sprintf(`{
	"model":"gpt-3.5-turbo",
	"messages": [{
		"role": "system",
		"content": "You are a helpful assistant."
	  },
	  {
		"role": "user",
		"content": "%q"
	  }],
	"temperature": 0.7,
	"max_tokens": %s
	}`, content, MaxTokens)

	return strings.NewReader(body)
}

func constructWordDefinitionBody(word, userTier, userNativeLanguage string) *strings.Reader {
	var maxWordCount string
	var MaxTokens string
	var content string

	if userTier == "Basic" {
		MaxTokens = "50"
		maxWordCount = "20"
		content = fmt.Sprintf("Define %s in %s & max %s words", word, userNativeLanguage, maxWordCount)

	} else if userTier == "Premium" {
		MaxTokens = "210"
		maxWordCount = "100"
		content = fmt.Sprintf("Define %s in %s & max %s words.Give 3 example sentences & explain the dictionary form", word, userNativeLanguage, maxWordCount)
	}

	body := fmt.Sprintf(`{
	"model":"gpt-3.5-turbo",
	"messages": [{
		"role": "system",
		"content": "You are a helpful assistant."
	  },
	  {
		"role": "user",
		"content": "%q"
	  }],
	"temperature": 0.7,
	"max_tokens": %s
	}`, content, MaxTokens)

	return strings.NewReader(body)
}
