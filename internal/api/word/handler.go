package word

import (
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/word/dto"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/word"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/render"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/request"
)

var FailedToProcessWord = "Failed to process your word. Please make sure you remove any extra spaces and special characters and try again"

type Handler interface {
	DefineWord() http.HandlerFunc
	GetSynonyms() http.HandlerFunc
}

type handler struct {
	logger  *zap.Logger
	service word.Service
}

func NewWordHandler(
	logger *zap.Logger,
	service word.Service,
) Handler {
	return &handler{
		logger:  logger,
		service: service,
	}
}

func (h *handler) DefineWord() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody dto.DefineWordRequest

		// Validates and decodes request
		if err := request.DecodeAndValidate(r.Body, &requestBody); err != nil {
			h.logger.Sugar().Errorw("failed to decode and validate define word request body",
				"error", err)

			render.Json(w, http.StatusBadRequest, FailedToProcessWord)

			return
		}

		spaceTrimmedWord := strings.TrimSpace(requestBody.Word)

		err := h.service.ValidateWord(spaceTrimmedWord)
		if err != nil {
			h.logger.Sugar().Errorw("failed to validate word", "word", spaceTrimmedWord, "native language", requestBody.NativeLanguage, "error", err)
			render.Json(w, http.StatusBadRequest, err.Error())

			return
		}

		response, err := h.service.GetWordDefinition(spaceTrimmedWord, requestBody.NativeLanguage)
		if err != nil {
			h.logger.Sugar().Errorw("failed to define word", "error", err, "word", spaceTrimmedWord, "native language", requestBody.NativeLanguage)
			render.Json(w, http.StatusBadRequest, FailedToProcessWord)

			return
		}

		render.Json(w, http.StatusOK, response.Choices[0].Message.Content)
	}
}

func (h *handler) GetSynonyms() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody dto.GetSynonymsRequest

		// Validates and decodes request
		if err := request.DecodeAndValidate(r.Body, &requestBody); err != nil {
			h.logger.Sugar().Errorw("failed to decode and validate define word request body",
				"error", err)

			render.Json(w, http.StatusBadRequest, FailedToProcessWord)

			return
		}

		spaceTrimmedWord := strings.TrimSpace(requestBody.Word)

		err := h.service.ValidateWord(spaceTrimmedWord)
		if err != nil {
			h.logger.Sugar().Errorw("failed to validate word", "word", spaceTrimmedWord, "error", err)
			render.Json(w, http.StatusBadRequest, err.Error())

			return
		}

		response, err := h.service.GetWordSynonyms(spaceTrimmedWord, requestBody.NativeLanguage)
		if err != nil {
			h.logger.Sugar().Errorw("failed to get word synonyms",
				"word", spaceTrimmedWord, "native language", requestBody.NativeLanguage, "error", err)

			render.Json(w, http.StatusInternalServerError, FailedToProcessWord)

			return
		}

		render.Json(w, http.StatusOK, response.Choices[0].Message.Content)
	}
}
