package dto

type DefineWordRequest struct {
	Word string `json:"word"`
	//Tier string `json:"tier"`
	//TargetLanguage string `json:"targetLanguage"`
	NativeLanguage string `json:"nativeLanguage"`
}

type GetSynonymsRequest struct {
	Word           string `json:"word"`
	NativeLanguage string `json:"nativeLanguage"`
}
