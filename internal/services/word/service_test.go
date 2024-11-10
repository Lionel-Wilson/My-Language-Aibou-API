package word_test

import (
	"errors"
	"testing"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/config"
	log "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/log"
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
			expectedErr: errors.New("Please provide a word"),
		},
		{
			name:        "Numbers are included in the word",
			word:        "hello123",
			expectedErr: errors.New("Words should not contain numbers."),
		},
		{
			name:        "Word is too long",
			word:        "superlongwordthatexceedscharactersss",
			expectedErr: errors.New("Word length too long. Must be less than 30 characters.If this is a sentence, please use the analyser."),
		},
		{
			name:        "Is not a word",
			word:        "This is a sentence",
			expectedErr: errors.New("This looks like a phrase. Please use the 'Analyzer'."),
		},
		{
			name:        "Is Nonsensical",
			word:        "ssssssss!!!aaa@:{P}{}",
			expectedErr: errors.New("This doesn't look like a word. Please provide a valid word."),
		},
	}

	for _, tt := range testCases {
		err := wordService.ValidateWord(tt.word)
		assert.EqualError(t, err, tt.expectedErr.Error())
	}
}
