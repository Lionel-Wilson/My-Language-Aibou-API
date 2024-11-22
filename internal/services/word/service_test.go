package word_test

import (
	"errors"
	"testing"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/config"
	log "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/log"
	openai "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/clients/open-ai"
	mockopenai "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/clients/open-ai/mock"
	word "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/services/word"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func newMockConfig() *config.Config {
	var cfg config.Config

	cfg.OpenAi.Key = "test-key"

	return &cfg
}

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

func GetWordDefinition(t *testing.T) {
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
				mockOpenAiClient.EXPECT().MakeRequest(gomock.Any()).Return(gomock.Any(), gomock.Any(), errors.New("An error"))
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

/*
{chatcmpl-ARVLNP5LZ9tfXee0w4osqEQrzrEJG chat.completion 1731118705 gpt-4o-2024-08-06 fp_9e15ccd6a4 [{0 {assistant The Korean word "어떻게" means "how" in English.
It is used to inquire about the manner or method in which something is done or occurs.
Here are two example sentences using "어떻게":

1. **Korean:** 이 문제를 어떻게 해결할 수 있을까요?
   **English:** How can we solve this problem?

2. **Korean:** 당신은 이 일을 어떻게 시작했나요?
   **English:** How did you start this work?

In both examples, "어떻게" is used to ask about the method or process involved.} <nil> stop}] {53 120 173}}


/*
func GetWordSynonyms(t *testing.T) {
	cfg := config.New()
	logger := log.New()
	wordService := word.New(cfg, logger)

	testCases := []struct {
		name        string
		word        string
		expectedErr error
	}{
		{
			name:        "No word provided",
			word:        "",
			expectedErr: errors.New("Please provide a word"),
		},
	}

	for _, tt := range testCases {
		err := wordService.ValidateWord(tt.word)
		assert.EqualError(t, err, tt.expectedErr.Error())
	}
}*/
