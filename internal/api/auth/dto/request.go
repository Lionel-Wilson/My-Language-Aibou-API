package dto

import "github.com/go-playground/validator/v10"

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UpdateDetailsRequest struct {
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	Password string `json:"password,omitempty"`
}

func (rr RegisterRequest) Validate() error {
	return validator.New().Struct(rr)
}

func (lr LoginRequest) Validate() error {
	return validator.New().Struct(lr)
}

func (udr UpdateDetailsRequest) Validate() error {
	return validator.New().Struct(udr)
}
