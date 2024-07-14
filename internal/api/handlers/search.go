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

	jsonBody := constructPhraseBody(requestBody.Phrase, requestBody.Tier, requestBody.NativeLanguage)

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
	var maxWordCount string
	var MaxTokens string

	if userTier == "Basic" {
		MaxTokens = "110"
		maxWordCount = "80"

	} else if userTier == "Premium" {
		MaxTokens = "330"
		maxWordCount = "230"
	}

	//content := fmt.Sprintf("Explain the meaning & grammar used in the following sentence in max %s words - '%s'", maxWordCount, phrase)

	body := fmt.Sprintf(`{
	"model":"gpt-3.5-turbo",
	"messages": [{
		"role": "system",
		"content": "You will be provided with a sentence in a certain language, and your task is to explain the meaning & grammar used in the sentence in max %s words. Responding in %s"
	  },
	  {
		"role": "user",
		"content": "%s"
	  }],
	"temperature": 0.7,
	"max_tokens": %s
	}`, maxWordCount, userNativeLanguage, phrase, MaxTokens)

	fmt.Printf("Tier: %s\n", userTier)
	fmt.Printf("Body: %s\n", body)
	//fmt.Printf("Phrase prompt: %s\n", content)

	return strings.NewReader(body)
}

func constructWordDefinitionBody(word, userTier, userNativeLanguage string) *strings.Reader {
	var maxWordCount string
	var MaxTokens string
	//var content string

	if userTier == "Basic" {
		MaxTokens = "60"
		maxWordCount = "40"
		//content = fmt.Sprintf("Define '%s' in max %s words", word, maxWordCount)

	} else if userTier == "Premium" {
		MaxTokens = "210"
		maxWordCount = "110"
		//content = fmt.Sprintf("Define '%s' in max %s words.Give 3 example sentences in the word's language with their translations in %s", userNativeLanguage, word, maxWordCount)
	}

	body := fmt.Sprintf(`{
	"model":"gpt-3.5-turbo",
	"messages": [{
		"role": "system",
		"content": "You will be provided a word and your tasks are to do the following in max %s words. First, define the word in %s. Second, give 3 example sentences that use that word. Third, translate the example sentences into %s for the user."
	  },
	  {
		"role": "user",
		"content": "%s"
	  }],
	"temperature": 0.7,
	"max_tokens": %s
	}`, maxWordCount, userNativeLanguage, userNativeLanguage, word, MaxTokens)

	fmt.Printf("Tier: %s\n", userTier)
	fmt.Printf("Body: %s\n", body)
	//fmt.Printf("Word prompt: %s\n", content)

	return strings.NewReader(body)
}
