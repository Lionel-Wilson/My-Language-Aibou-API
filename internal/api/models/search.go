package models

type DefinePhraseRequest struct {
	Phrase         string `json:"phrase"`
	Tier           string `json:"tier"`
	TargetLanguage string `json:"targetLanguage"`
	NativeLanguage string `json:"nativeLanguage"`
}
type DefineWordRequest struct {
	Word           string `json:"word"`
	Tier           string `json:"tier"`
	TargetLanguage string `json:"targetLanguage"`
	NativeLanguage string `json:"nativeLanguage"`
}
