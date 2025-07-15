package dto

import (
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/entity"
)

type UserDetailsResponse struct {
	Email string `json:"email,omitempty" validate:"omitempty,email"`
}

func ToResponse(user *entity.User) UserDetailsResponse {
	return UserDetailsResponse{
		Email: user.Email,
	}
}
