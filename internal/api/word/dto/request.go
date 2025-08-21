package dto

import "github.com/go-playground/validator/v10"

type WordRequest struct {
	Word           string `json:"word" `
	NativeLanguage string `json:"nativeLanguage" `
}

func (wr WordRequest) Validate() error {
	return validator.New().Struct(wr)
}
