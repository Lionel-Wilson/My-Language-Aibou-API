package dto

type DefineSentenceRequest struct {
	Sentence       string `json:"sentence"`
	NativeLanguage string `json:"nativeLanguage"`
}
