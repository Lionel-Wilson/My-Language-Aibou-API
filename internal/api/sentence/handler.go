package sentence

import (
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/sentence/dto"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/sentence"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/messages"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/render"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/request"
)

type Handler interface {
	ExplainSentence() http.HandlerFunc
	CorrectSentence() http.HandlerFunc
	Simplify() http.HandlerFunc
}

type handler struct {
	logger  *zap.Logger
	service sentence.Service
}

func NewSentenceHandler(
	logger *zap.Logger,
	service sentence.Service,
) Handler {
	return &handler{
		logger:  logger,
		service: service,
	}
}

var FailedToProcessSentence = "Failed to process your sentence(s).Please make sure you remove any line breaks and large gaps between your sentences and try again"

func (h *handler) ExplainSentence() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var requestBody dto.DefineSentenceRequest

		// Validates and decodes request
		if err := request.DecodeAndValidate(r.Body, &requestBody); err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate explain sentence request body",
				"error", err)

			render.Json(w, http.StatusBadRequest, FailedToProcessSentence)

			return
		}

		trimmedSentence := strings.TrimSpace(requestBody.Sentence)

		err := h.service.ValidateSentence(trimmedSentence)
		if err != nil {
			h.logger.Sugar().Infow("sentence validation failed",
				"sentence", trimmedSentence, "error", err)
			render.Json(w, http.StatusBadRequest, err.Error())

			return
		}

		response, err := h.service.GetSentenceExplanation(ctx, trimmedSentence, requestBody.NativeLanguage, requestBody.IsDetailed)
		if err != nil {
			h.logger.Sugar().Errorw("sentence explanation failed",
				"sentence", trimmedSentence,
				"nativeLanguage", requestBody.NativeLanguage,
				"error", err)
			render.Json(w, http.StatusInternalServerError, messages.InternalServerErrorMsg)

			return
		}

		render.Json(w, http.StatusOK, response)
	}
}

func (h *handler) CorrectSentence() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var requestBody dto.DefineSentenceRequest

		// Validates and decodes request
		if err := request.DecodeAndValidate(r.Body, &requestBody); err != nil {
			h.logger.Sugar().Warnw(
				"failed to decode and validate correct sentence request body",
				"error", err)

			render.Json(w, http.StatusBadRequest, FailedToProcessSentence)

			return
		}

		trimmedSentence := strings.TrimSpace(requestBody.Sentence)

		err := h.service.ValidateSentence(trimmedSentence)
		if err != nil {
			h.logger.Sugar().Infow(
				"sentence validation failed",
				"sentence", trimmedSentence, "error", err)
			render.Json(w, http.StatusBadRequest, err.Error())

			return
		}

		response, err := h.service.GetSentenceCorrection(ctx, trimmedSentence, requestBody.NativeLanguage)
		if err != nil {
			h.logger.Sugar().Errorw("sentence correction failed",
				"sentence", trimmedSentence,
				"nativeLanguage", requestBody.NativeLanguage,
				"error", err)
			render.Json(w, http.StatusInternalServerError, messages.InternalServerErrorMsg)

			return
		}

		render.Json(w, http.StatusOK, response)
	}
}

func (h *handler) Simplify() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
