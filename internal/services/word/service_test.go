package word_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/log"
	openai "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/clients/open-ai"
	mockopenai "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/clients/open-ai/mock"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/services/word"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TO DO: UNIT TESTS
func TestValidateWord(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockOpenAiClient := mockopenai.NewMockClient(ctrl)
	logger := log.New()

	wordService := word.New(logger, mockOpenAiClient)

	testCases := []struct {
		name        string
		word        string
		expectedErr error
	}{
		{
			name:        "No word provided",
			word:        "",
			expectedErr: errors.New("please provide a word"),
		},
		{
			name:        "Numbers are included in the word",
			word:        "hello123",
			expectedErr: errors.New("words should not contain numbers"),
		},
		{
			name:        "Word is too long",
			word:        "superlongwordthatexceedscharactersss",
			expectedErr: errors.New("word length too long. Must be less than 30 characters.If this is a sentence, please use the analyser"),
		},
		{
			name:        "Is not a word",
			word:        "This is a sentence",
			expectedErr: errors.New("this looks like a phrase. Please use the 'Analyzer'"),
		},
		{
			name:        "Is Nonsensical",
			word:        "ssssssss!!!aaa@:{P}{}",
			expectedErr: errors.New("this doesn't look like a word. Please provide a valid word"),
		},
	}

	for _, tt := range testCases {
		err := wordService.ValidateWord(tt.word)
		assert.EqualError(t, err, tt.expectedErr.Error())
	}
}

func TestGetWordDefinition(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockOpenAiClient := mockopenai.NewMockClient(ctrl)
	logger := log.New()

	wordService := word.New(logger, mockOpenAiClient)

	testCases := []struct {
		name             string
		word             string
		nativeLanguage   string
		expectedResponse *openai.ChatCompletion
		expectedErr      error
		mock             func()
	}{
		{
			name:             "Failed OpenAi Request",
			word:             "어떻게",
			nativeLanguage:   "english",
			expectedResponse: &openai.ChatCompletion{},
			expectedErr:      errors.New("an error"),
			mock: func() {
				mockOpenAiClient.EXPECT().MakeRequest(gomock.Any()).Return(&http.Response{}, []byte{}, errors.New("an error"))
			},
		},
	}

	for _, tc := range testCases {
		tc.mock()

		response, err := wordService.GetWordDefinition(tc.word, tc.nativeLanguage)

		// Check if the error matches the expected outcome
		if tc.expectedErr != nil {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedResponse, response)
		}

	}
}
