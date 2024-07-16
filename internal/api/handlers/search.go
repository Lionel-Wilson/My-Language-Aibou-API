package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"unicode"
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
	if utils.ContainsNumber(word) {
		app.ErrorLog.Printf("User provided a word(%s) that contained a number.", word)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Words should not contain numbers.", []string{})
		return
	}
	if isNotAWord(word) {
		app.ErrorLog.Printf("User provided a phrase(%s) instead of a word.", word)
		utils.NewErrorResponse(c, http.StatusBadRequest, "This looks like a phrase. Please use the 'Analyzer'.", []string{})
		return
	}
	if isNonsensical(word) {
		app.ErrorLog.Printf("User provided nonsense(%s) instead of a word.", word)
		utils.NewErrorResponse(c, http.StatusBadRequest, "This doesn't look like a word. Please provide a valid word.", []string{})
		return
	}
	if utf8.RuneCountInString(word) > 30 {
		app.ErrorLog.Printf("Word '%s' length too long. Must be less than 30 characters.", word)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Word length too long. Must be less than 30 characters.If this is a sentence, please use the analyser.", []string{})
		return
	}

	jsonBody := constructWordDefinitionBody(word, requestBody.NativeLanguage)

	OpenAIApiResponse, err := utils.MakeOpenAIApiRequest(jsonBody, c, *app.OpenApiKey)
	if err != nil {
		app.ErrorLog.Println(err.Error())
		utils.ServerErrorResponse(c, err, "Failed to process your sentence(s).Please make sure you remove any line breaks and large gaps between your sentences and try again")
		return
	}

	fmt.Printf("Word Definition: %s\n", OpenAIApiResponse.Choices[0].Message.Content)
	fmt.Printf(`Prompt Tokens: %d`, OpenAIApiResponse.Usage.PromptTokens)
	fmt.Printf(`Response Tokens: %d`, OpenAIApiResponse.Usage.CompletionTokens)
	fmt.Printf(`Total Tokens used: %d`, OpenAIApiResponse.Usage.TotalTokens)

	c.JSON(http.StatusOK, OpenAIApiResponse.Choices[0].Message.Content)
}

func (app *Application) DefineSentence(c *gin.Context) {
	var requestBody models.DefineSentenceRequest

	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		app.ErrorLog.Println(err.Error())
		utils.ServerErrorResponse(c, err, "")
		return
	}

	sentence := strings.TrimSpace(requestBody.Sentence)

	if sentence == "" {
		app.ErrorLog.Printf("User didn't provide a sentence: %s", sentence)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Please provide a sentence", []string{})
		return
	}

	if utf8.RuneCountInString(sentence) > 150 {
		app.ErrorLog.Printf("Sentence '%s' length too long. Must be less than 150 characters.", sentence)
		utils.NewErrorResponse(c, http.StatusBadRequest, "The provided sentence is too long(must be less than 150 characters). Try breaking down the sentence into smaller parts", []string{})
		return
	}

	jsonBody := constructPhraseBody(sentence, requestBody.NativeLanguage)

	OpenAIApiResponse, err := utils.MakeOpenAIApiRequest(jsonBody, c, *app.OpenApiKey)
	if err != nil {
		app.ErrorLog.Println(err.Error())
		utils.ServerErrorResponse(c, err, "Failed to process request. Please try again later")
		return
	}

	fmt.Printf("Phrase explanation: %s\n", OpenAIApiResponse.Choices[0].Message.Content)
	fmt.Printf("Prompt Tokens: %d\n", OpenAIApiResponse.Usage.PromptTokens)
	fmt.Printf("Response Tokens: %d\n", OpenAIApiResponse.Usage.CompletionTokens)
	fmt.Printf("Total Tokens used: %d\n", OpenAIApiResponse.Usage.TotalTokens)

	c.JSON(http.StatusOK, OpenAIApiResponse.Choices[0].Message.Content)
}

func constructPhraseBody(sentence, userNativeLanguage string) *strings.Reader {
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

func constructWordDefinitionBody(word, userNativeLanguage string) *strings.Reader {
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

// isNotAWord is used to check if the user is using the dictionary to define phrases as opposed to a single word
func isNotAWord(s string) bool {
	return strings.Count(s, " ") > 1
}

func isNonsensical(s string) bool {
	// Check condition 1: The string contains special characters
	hasSpecialCharacters := false
	for _, ch := range s {
		if unicode.IsPunct(ch) || unicode.IsSymbol(ch) {
			hasSpecialCharacters = true
			break
		}
	}

	// Check condition 2: The string contains the same character more than 3 times in a row
	hasRepeatingCharacters := false
	for i := 0; i < len(s)-3; i++ {
		if s[i] == s[i+1] && s[i] == s[i+2] && s[i] == s[i+3] {
			hasRepeatingCharacters = true
			break
		}
	}

	// Combine all conditions to determine if the string is nonsensical
	return hasSpecialCharacters || hasRepeatingCharacters
}
