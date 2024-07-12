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

	c.JSON(http.StatusOK, "Here's your definition")
}

func (app *Application) DefinePhrase(c *gin.Context) {
	var requestBody models.DefinePhraseRequest

	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		utils.ServerErrorResponse(c, err, "")
		return
	}

	jsonBody := constructPhraseBody(requestBody.Phrase, requestBody.Tier, requestBody.TargetLanguage, requestBody.NativeLanguage)

	PhraseBreakdown := utils.MakeOpenAIApiRequest(jsonBody, c, *app.OpenApiKey)

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

	content := fmt.Sprintf("Break down the meaning and grammar used in the below %s sentence. Write your breakdown in %s and a maximum of %s words - %s", userTargetLanguage, userNativeLanguage, maxWordCount, phrase)

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
