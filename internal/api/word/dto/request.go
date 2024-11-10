package dto

type DefineWordRequest struct {
	Word           string `json:"word"`
	NativeLanguage string `json:"nativeLanguage"`
}

type GetSynonymsRequest struct {
	Word           string `json:"word"`
	NativeLanguage string `json:"nativeLanguage"`
}
