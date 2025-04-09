package dto

import "github.com/go-playground/validator/v10"

type DefineSentenceRequest struct {
	Sentence       string `json:"sentence"`
	NativeLanguage string `json:"nativeLanguage" `
}

func (dsr DefineSentenceRequest) Validate() error {
	return validator.New().Struct(dsr)
}
