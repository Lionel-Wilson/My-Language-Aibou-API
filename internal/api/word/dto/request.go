package dto

import "github.com/go-playground/validator/v10"

type DefineWordRequest struct {
	Word           string `json:"word" `
	NativeLanguage string `json:"nativeLanguage" `
}

type GetSynonymsRequest struct {
	Word           string `json:"word"`
	NativeLanguage string `json:"nativeLanguage"`
}

type GetHistoryRequest struct {
	Word           string `json:"word"`
	NativeLanguage string `json:"nativeLanguage"`
}

func (ghr GetHistoryRequest) Validate() error {
	return validator.New().Struct(ghr)
}

func (dwr DefineWordRequest) Validate() error {
	return validator.New().Struct(dwr)
}

func (gsr GetSynonymsRequest) Validate() error {
	return validator.New().Struct(gsr)
}
