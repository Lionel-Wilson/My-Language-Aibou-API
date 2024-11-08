package word

import (
	"net/http"
	"strings"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/log"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/word/dto"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/utils"
	word "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/word"
	"github.com/gin-gonic/gin"
)

type Handler interface {
	DefineWord(c *gin.Context)
	GetSynonyms(c *gin.Context)
}

type wordHandler struct {
	logger  log.Logger
	service word.Service
}

func NewHandler(
	logger log.Logger,
	service word.Service,
) Handler {
	return &wordHandler{
		logger:  logger,
		service: service,
	}
}

func (h *wordHandler) DefineWord(c *gin.Context) {
	var requestBody dto.DefineWordRequest

	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		h.logger.Error(err.Error())
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

func (h *wordHandler) GetSynonyms(c *gin.Context) {
	var requestBody dto.GetSynonymsRequest

	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		h.logger.Error(err.Error())
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
