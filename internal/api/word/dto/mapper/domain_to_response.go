package mapper

import (
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/word/dto"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/word/domain"
)

func MapToLookUpResponse(details *domain.LookupDetails) dto.LookupResponse {
	if details == nil {
		return dto.LookupResponse{}
	}

	return dto.LookupResponse{
		Definition: details.Definition,
		Synonyms:   details.Synonyms,
		History:    details.History,
	}
}
