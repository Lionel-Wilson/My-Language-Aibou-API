package models

type DefinePhraseRequest struct {
	Phrase         string `json:"phrase"`
	Tier           string `json:"tier"`
	TargetLanguage string `json:"targetLanguage"`
	NativeLanguage string `json:"nativeLanguage"`
}
