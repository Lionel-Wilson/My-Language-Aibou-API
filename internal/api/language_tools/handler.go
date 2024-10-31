package handler

import (
	"net/http"
	"strings"

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

	err = h.service.ValidateWord(word)
	if err != nil {
		utils.NewErrorResponse(c, http.StatusBadRequest, err.Error(), []string{})
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

	err = h.service.ValidateSentence(sentence)
	if err != nil {
		utils.NewErrorResponse(c, http.StatusBadRequest, err.Error(), []string{})
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

	err = h.service.ValidateWord(word)
	if err != nil {
		utils.NewErrorResponse(c, http.StatusBadRequest, err.Error(), []string{})
		return
	}

	response, err := h.service.GetWordSynonyms(c, word, requestBody.NativeLanguage)
	if err != nil {
		utils.ServerErrorResponse(c, err, "Failed to process your word.Please make sure you remove any extra spaces & special characters and try again")
		return
	}

	c.JSON(http.StatusOK, response.Choices[0].Message.Content)
}
