package word

import (
	"github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/messages"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/word/dto"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/word"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/render"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/request"
)

type Handler interface {
	DefineWord() http.HandlerFunc
	GetSynonyms() http.HandlerFunc
	GetHistory() http.HandlerFunc
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

var FailedToProcessWord = "Failed to process your word. Please make sure you remove any extra spaces and special characters and try again"

func (h *handler) GetHistory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var requestBody dto.DefineWordRequest

		// Validates and decodes request
		if err := request.DecodeAndValidate(r.Body, &requestBody); err != nil {
			h.logger.Sugar().Infow("failed to decode and validate word history request body",
				"error", err)

			render.Json(w, http.StatusBadRequest, FailedToProcessWord)

			return
		}

		spaceTrimmedWord := strings.TrimSpace(requestBody.Word)

		err := h.service.ValidateWord(spaceTrimmedWord)
		if err != nil {
			h.logger.Sugar().Infow(
				"failed to validate word",
				zap.Error(err),
				"word", spaceTrimmedWord,
				"nativeLanguage", requestBody.NativeLanguage)

			render.Json(w, http.StatusBadRequest, err.Error())

			return
		}

		response, err := h.service.GetWordHistory(ctx, spaceTrimmedWord, requestBody.NativeLanguage)
		if err != nil {
			h.logger.Sugar().Errorw(
				"failed to get word history",
				zap.Any("context", ctx),
				zap.Error(err),
				"word", spaceTrimmedWord,
				"nativeLanguage", requestBody.NativeLanguage)
			render.Json(w, http.StatusInternalServerError, messages.InternalServerErrorMsg)

			return
		}

		render.Json(w, http.StatusOK, response)
	}
}

func (h *handler) DefineWord() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var requestBody dto.DefineWordRequest

		// Validates and decodes request
		if err := request.DecodeAndValidate(r.Body, &requestBody); err != nil {
			h.logger.Sugar().Infow("failed to decode and validate define word request body",
				zap.Error(err))

			render.Json(w, http.StatusBadRequest, FailedToProcessWord)

			return
		}

		spaceTrimmedWord := strings.TrimSpace(requestBody.Word)

		err := h.service.ValidateWord(spaceTrimmedWord)
		if err != nil {
			h.logger.Sugar().Infow(
				"failed to validate word",
				zap.Any("context", ctx),
				zap.Error(err),
				"word", spaceTrimmedWord,
				"nativeLanguage", requestBody.NativeLanguage)
			render.Json(w, http.StatusBadRequest, err.Error())

			return
		}

		response, err := h.service.GetWordDefinition(ctx, spaceTrimmedWord, requestBody.NativeLanguage)
		if err != nil {
			h.logger.Sugar().Errorw(
				"failed to define word",
				zap.Any("context", ctx),
				zap.Error(err), "word",
				spaceTrimmedWord, "nativeLanguage",
				requestBody.NativeLanguage)
			render.Json(w, http.StatusInternalServerError, messages.InternalServerErrorMsg)

			return
		}

		render.Json(w, http.StatusOK, response)
	}
}

// todo: create a mapper that converts validation errorrs to user friendly responses
func (h *handler) GetSynonyms() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var requestBody dto.GetSynonymsRequest

		// Validates and decodes request
		if err := request.DecodeAndValidate(r.Body, &requestBody); err != nil {
			h.logger.Sugar().Infow("failed to decode and validate define word request body", zap.Error(err))

			render.Json(w, http.StatusBadRequest, FailedToProcessWord)

			return
		}

		spaceTrimmedWord := strings.TrimSpace(requestBody.Word)

		err := h.service.ValidateWord(spaceTrimmedWord)
		if err != nil {
			h.logger.Sugar().Infow(
				"failed to validate word", zap.Error(err), "word", spaceTrimmedWord)
			render.Json(w, http.StatusBadRequest, err.Error())

			return
		}

		response, err := h.service.GetWordSynonyms(ctx, spaceTrimmedWord, requestBody.NativeLanguage)
		if err != nil {
			h.logger.Sugar().Errorw("failed to get word synonyms",
				zap.Any("context", ctx),
				zap.Error(err),
				"word", spaceTrimmedWord,
				"nativeLanguage", requestBody.NativeLanguage)

			render.Json(w, http.StatusInternalServerError, messages.InternalServerErrorMsg)

			return
		}

		render.Json(w, http.StatusOK, response)
	}
}
