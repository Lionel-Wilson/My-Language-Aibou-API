package models

type DefineSentenceRequest struct {
	Sentence string `json:"sentence"`
	//Tier   string `json:"tier"`
	//TargetLanguage string `json:"targetLanguage"`
	NativeLanguage string `json:"nativeLanguage"`
}
type DefineWordRequest struct {
	Word string `json:"word"`
	//Tier string `json:"tier"`
	//TargetLanguage string `json:"targetLanguage"`
	NativeLanguage string `json:"nativeLanguage"`
}
