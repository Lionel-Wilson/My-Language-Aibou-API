package handler

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/language_tools/dto"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/log"
	languagetools "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/language_tools"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/utils"
	"github.com/gin-gonic/gin"
)

type Handler interface {
	DefineWord(c *gin.Context)
	ExplainSentence(c *gin.Context)
	GetSynonyms(c *gin.Context)
}

type languageToolsHandler struct {
	logger  *log.Logger
	service languagetools.Service
}

func NewHandler(
	logger *log.Logger,
	service languagetools.Service,
) Handler {
	return &languageToolsHandler{
		logger:  logger,
		service: service,
	}
}

func (h *languageToolsHandler) DefineWord(c *gin.Context) {
	var requestBody dto.DefineWordRequest

	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		h.logger.ErrorLog.Println(err.Error())
		utils.ServerErrorResponse(c, err, "Failed to process your word.Please make sure you remove any extra spaces & special characters and try again")
		return
	}

	word := strings.TrimSpace(requestBody.Word)

	if word == "" {
		h.logger.ErrorLog.Printf("User didn't provide a word: %s", word)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Please provide a word", []string{})
		return
	}
	if utils.ContainsNumber(word) {
		h.logger.ErrorLog.Printf("User provided a word(%s) that contained a number.", word)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Words should not contain numbers.", []string{})
		return
	}
	if utf8.RuneCountInString(word) > 30 {
		h.logger.ErrorLog.Printf("Word '%s' length too long. Must be less than 30 characters.", word)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Word length too long. Must be less than 30 characters.If this is a sentence, please use the analyser.", []string{})
		return
	}
	if h.service.IsNotAWord(word) {
		h.logger.ErrorLog.Printf("User provided a phrase(%s) instead of a word.", word)
		utils.NewErrorResponse(c, http.StatusBadRequest, "This looks like a phrase. Please use the 'Analyzer'.", []string{})
		return
	}
	if h.service.IsNonsensical(word) {
		h.logger.ErrorLog.Printf("User provided nonsense(%s) instead of a word.", word)
		utils.NewErrorResponse(c, http.StatusBadRequest, "This doesn't look like a word. Please provide a valid word.", []string{})
		return
	}

	response, err := h.service.GetWordDefinition(c, word, requestBody.NativeLanguage)
	if err != nil {
		utils.ServerErrorResponse(c, err, "Failed to process your word. Please make sure you remove any extra spaces & special characters and try again")
		return
	}

	c.JSON(http.StatusOK, response.Choices[0].Message.Content)
}

func (h *languageToolsHandler) ExplainSentence(c *gin.Context) {
	var requestBody dto.DefineSentenceRequest

	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		h.logger.ErrorLog.Println(err.Error())
		utils.ServerErrorResponse(c, err, "Failed to process your sentence(s).Please make sure you remove any line breaks and large gaps between your sentences and try again")
		return
	}

	sentence := strings.TrimSpace(requestBody.Sentence)

	if sentence == "" {
		h.logger.ErrorLog.Printf("User didn't provide a sentence: %s", sentence)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Please provide a sentence", []string{})
		return
	}

	if utf8.RuneCountInString(sentence) > 100 {
		h.logger.ErrorLog.Printf("Sentence '%s' length too long. Must be less than 100 characters.", sentence)
		utils.NewErrorResponse(c, http.StatusBadRequest, "The sentence must be less than 100 characters.", []string{})
		return
	}

	response, err := h.service.GetSentenceExplanation(c, sentence, requestBody.NativeLanguage)
	if err != nil {
		utils.ServerErrorResponse(c, err, "Failed to process your sentence(s).Please make sure you remove any line breaks and large gaps between your sentences and try again")
		return
	}

	c.JSON(http.StatusOK, response.Choices[0].Message.Content)
}

func (h *languageToolsHandler) GetSynonyms(c *gin.Context) {
	var requestBody dto.GetSynonymsRequest

	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		h.logger.ErrorLog.Println(err.Error())
		utils.ServerErrorResponse(c, err, "Failed to process your word. Please make sure you remove any extra spaces & special characters and try again")
		return
	}

	word := strings.TrimSpace(requestBody.Word)

	if word == "" {
		h.logger.ErrorLog.Printf("User didn't provide a word: %s", word)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Please provide a word", []string{})
		return
	}
	if utils.ContainsNumber(word) {
		h.logger.ErrorLog.Printf("User provided a word(%s) that contained a number.", word)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Words should not contain numbers.", []string{})
		return
	}
	if utf8.RuneCountInString(word) > 30 {
		h.logger.ErrorLog.Printf("Word '%s' length too long. Must be less than 30 characters.", word)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Word length too long. Must be less than 30 characters.If this is a sentence, please use the analyser.", []string{})
		return
	}
	if h.service.IsNotAWord(word) {
		h.logger.ErrorLog.Printf("User provided a phrase(%s) instead of a word.", word)
		utils.NewErrorResponse(c, http.StatusBadRequest, "This looks like a phrase. Please use the 'Analyzer'.", []string{})
		return
	}
	if h.service.IsNonsensical(word) {
		h.logger.ErrorLog.Printf("User provided nonsense(%s) instead of a word.", word)
		utils.NewErrorResponse(c, http.StatusBadRequest, "This doesn't look like a word. Please provide a valid word.", []string{})
		return
	}

	response, err := h.service.GetWordSynonyms(c, word, requestBody.NativeLanguage)
	if err != nil {
		utils.ServerErrorResponse(c, err, "Failed to process your word.Please make sure you remove any extra spaces & special characters and try again")
		return
	}

	c.JSON(http.StatusOK, response.Choices[0].Message.Content)
}
