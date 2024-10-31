package dto

type DefineSentenceRequest struct {
	Sentence string `json:"sentence"`
	//Tier   string `json:"tier"`
	//TargetLanguage string `json:"targetLanguage"`
	NativeLanguage string `json:"nativeLanguage"`
}
