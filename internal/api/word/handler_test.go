package word_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/word"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/word/dto"
	openai "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/clients/open-ai"
	wordmock "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/services/word/mock"
)

// TO-DO: Handler tests
func TestDefineWordHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := wordmock.NewMockService(ctrl)
	mockLogger := zaptest.NewLogger(t)
	handler := word.NewHandler(*mockLogger, mockService)

	router := gin.Default()
	router.POST("search/word", handler.DefineWord)

	testCases := []struct {
		name           string
		requestBody    dto.DefineWordRequest
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "no word provided",
			requestBody: dto.DefineWordRequest{
				Word:           "",
				NativeLanguage: "english",
			},
			mockSetup: func() {
				mockService.EXPECT().ValidateWord("").Return(errors.New("please provide a word")).Times(1)
			},
			expectedBody:   `please provide a word`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Failed to get word definition",
			requestBody: dto.DefineWordRequest{
				Word:           "恋愛",
				NativeLanguage: "english",
			},
			mockSetup: func() {
				mockService.EXPECT().ValidateWord("恋愛").Return(nil).Times(1)
				mockService.EXPECT().GetWordDefinition(gomock.Any(), gomock.Any()).Return(&openai.ChatCompletion{}, errors.New("an error"))
			},
			expectedBody:   "an error",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	// Run test cases
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			body, _ := json.Marshal(tt.requestBody)
			req, err := http.NewRequest(http.MethodPost, "/search/word", bytes.NewBuffer(body))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			require.Contains(t, recorder.Body.String(), tt.expectedBody)
		})
	}
}
