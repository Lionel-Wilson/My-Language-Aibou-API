package domain

import (
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/auth/dto"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	HashedPassword string
	Email          string
}

func RegisterRequestToUserDomain(req dto.RegisterRequest) (user *User, err error) {

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &User{
		HashedPassword: string(hashedPassword),
		Email:          req.Email,
	}, nil
}
