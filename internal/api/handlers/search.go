package handlers

import (
	"fmt"
	"net/http"
	"strings"

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

	jsonBody := constructWordDefinitionBody(requestBody.Word, requestBody.Tier, requestBody.TargetLanguage, requestBody.NativeLanguage)

	wordDefinition, err := utils.MakeOpenAIApiRequest(jsonBody, c, *app.OpenApiKey)
	if err != nil {
		app.ErrorLog.Println(err.Error())
		utils.ServerErrorResponse(c, err, "Failed to make request to OpenAI API")
		return
	}

	c.JSON(http.StatusOK, wordDefinition)
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

	PhraseBreakdown, err := utils.MakeOpenAIApiRequest(jsonBody, c, *app.OpenApiKey)
	if err != nil {
		app.ErrorLog.Println(err.Error())
		utils.ServerErrorResponse(c, err, "Failed to make request to OpenAI API")
		return
	}

	c.JSON(http.StatusOK, PhraseBreakdown)
}

func constructPhraseBody(phrase, userTier, userTargetLanguage, userNativeLanguage string) *strings.Reader {
	var maxWordCount string
	var MaxTokens string

	if userTier == "Basic" {
		MaxTokens = "110"
		maxWordCount = "80"

	} else if userTier == "Premium" {
		MaxTokens = "330"
		maxWordCount = "170"
	}

	content := fmt.Sprintf("Break down the meaning and grammar used in the following %s sentence. Write your breakdown in %s and a maximum of %s words - %s", userTargetLanguage, userNativeLanguage, maxWordCount, phrase)

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
	"temperature": 0.7,
	"max_tokens": %s
	}`, content, MaxTokens)

	return strings.NewReader(body)
}

func constructWordDefinitionBody(word, userTier, userTargetLanguage, userNativeLanguage string) *strings.Reader {
	var maxWordCount string
	var MaxTokens string
	var content string

	if userTier == "Basic" {
		MaxTokens = "50"
		maxWordCount = "20"
		content = fmt.Sprintf("Define the following %s word. Define it in %s and a maximum of %s words - %s", userTargetLanguage, userNativeLanguage, maxWordCount, word)

	} else if userTier == "Premium" {
		MaxTokens = "210"
		maxWordCount = "100"
		content = fmt.Sprintf("Define the following %s word. Define it in %s and a maximum of %s words. Make sure to include 3 example sentences and explain the dictionary form - %s", userTargetLanguage, userNativeLanguage, maxWordCount, word)
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
	"temperature": 0.7,
	"max_tokens": %s
	}`, content, MaxTokens)

	return strings.NewReader(body)
}
