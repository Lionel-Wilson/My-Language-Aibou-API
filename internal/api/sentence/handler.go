package handler

import (
	"net/http"
	"strings"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/log"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/sentence/dto"
	sentence "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/sentence"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/utils"
	"github.com/gin-gonic/gin"
)

type Handler interface {
	ExplainSentence(c *gin.Context)
	CorrectSentence(c *gin.Context)
}

type sentenceHandler struct {
	logger  *log.Logger
	service sentence.Service
}

func NewHandler(
	logger *log.Logger,
	service sentence.Service,
) Handler {
	return &sentenceHandler{
		logger:  logger,
		service: service,
	}
}

func (h *sentenceHandler) ExplainSentence(c *gin.Context) {
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

func (h *sentenceHandler) CorrectSentence(c *gin.Context) {
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

	response, err := h.service.GetSentenceCorrection(c, sentence, requestBody.NativeLanguage)
	if err != nil {
		utils.ServerErrorResponse(c, err, "Failed to process your sentence(s).Please make sure you remove any line breaks and large gaps between your sentences and try again")
		return
	}

	c.JSON(http.StatusOK, response.Choices[0].Message.Content)
}
