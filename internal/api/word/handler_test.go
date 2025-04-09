package word_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/word"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/word/dto"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/clients/open-ai"
	wordmock "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/word/mock"
)

func TestDefineWordHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := wordmock.NewMockService(ctrl)
	mockLogger := zaptest.NewLogger(t)
	handler := word.NewWordHandler(mockLogger, mockService)

	r := chi.NewRouter()
	r.Post("/api/v1/word/definition", handler.DefineWord())

	testCases := []struct {
		name           string
		requestBody    dto.DefineWordRequest
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "empty word",
			requestBody: dto.DefineWordRequest{
				Word:           "",
				NativeLanguage: "english",
			},
			mockSetup: func() {
				mockService.EXPECT().ValidateWord("").Return(errors.New("please provide a word"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "please provide a word",
		},
		{
			name: "valid word but API fails",
			requestBody: dto.DefineWordRequest{
				Word:           "hello",
				NativeLanguage: "english",
			},
			mockSetup: func() {
				mockService.EXPECT().ValidateWord("hello").Return(nil)
				mockService.EXPECT().GetWordDefinition("hello", "english").Return(nil, errors.New("api error"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Failed to process your word",
		},
		{
			name: "successful definition",
			requestBody: dto.DefineWordRequest{
				Word:           "hello",
				NativeLanguage: "english",
			},
			mockSetup: func() {
				mockService.EXPECT().ValidateWord("hello").Return(nil)
				mockService.EXPECT().GetWordDefinition("hello", "english").Return(&openai.ChatCompletion{
					Choices: []openai.Choice{
						{
							Message: openai.Message{
								Content: "Definition of hello",
							},
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Definition of hello",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/word/definition", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)
		})
	}
}
