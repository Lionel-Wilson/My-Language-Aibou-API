package dto

type LookupResponse struct {
	Definition string `json:"definition"`
	Synonyms   string `json:"synonyms"`
	History    string `json:"history"`
}
