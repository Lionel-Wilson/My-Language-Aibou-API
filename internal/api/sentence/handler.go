package sentence

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/sentence/dto"
	sentence "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/services/sentence"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/utils"
)

type Handler interface {
	ExplainSentence(c *gin.Context)
	CorrectSentence(c *gin.Context)
}

type handler struct {
	logger  zap.Logger
	service sentence.Service
}

func NewSentenceHandler(
	logger zap.Logger,
	service sentence.Service,
) Handler {
	return &handler{
		logger:  logger,
		service: service,
	}
}

func (h *handler) ExplainSentence(c *gin.Context) {
	var requestBody dto.DefineSentenceRequest

	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		h.logger.Error(err.Error())
		utils.ServerErrorResponse(c, err, "Failed to process your trimmedSentence(s).Please make sure you remove any line breaks and large gaps between your sentences and try again")

		return
	}

	trimmedSentence := strings.TrimSpace(requestBody.Sentence)

	err = h.service.ValidateSentence(trimmedSentence)
	if err != nil {
		utils.NewErrorResponse(c, http.StatusBadRequest, err.Error(), []string{})
		return
	}

	response, err := h.service.GetSentenceExplanation(c, trimmedSentence, requestBody.NativeLanguage)
	if err != nil {
		utils.ServerErrorResponse(c, err, "Failed to process your trimmedSentence(s).Please make sure you remove any line breaks and large gaps between your sentences and try again")
		return
	}

	c.JSON(http.StatusOK, response.Choices[0].Message.Content)
}

func (h *handler) CorrectSentence(c *gin.Context) {
	var requestBody dto.DefineSentenceRequest

	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		h.logger.Error(err.Error())
		utils.ServerErrorResponse(c, err, "Failed to process your trimmedSentence(s).Please make sure you remove any line breaks and large gaps between your sentences and try again")

		return
	}

	trimmedSentence := strings.TrimSpace(requestBody.Sentence)

	err = h.service.ValidateSentence(trimmedSentence)
	if err != nil {
		utils.NewErrorResponse(c, http.StatusBadRequest, err.Error(), []string{})
		return
	}

	response, err := h.service.GetSentenceCorrection(c, trimmedSentence, requestBody.NativeLanguage)
	if err != nil {
		utils.ServerErrorResponse(c, err, "Failed to process your trimmedSentence(s).Please make sure you remove any line breaks and large gaps between your sentences and try again")
		return
	}

	c.JSON(http.StatusOK, response.Choices[0].Message.Content)
}
