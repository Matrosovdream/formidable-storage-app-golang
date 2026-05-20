package model

import "time"

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=1"`
	Remember bool   `json:"remember"`
}

type RegisterRequest struct {
	Name                 string `json:"name"                  validate:"required,max=255"`
	Email                string `json:"email"                 validate:"required,email,max=255"`
	Password             string `json:"password"              validate:"required,min=8"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required,eqfield=Password"`
}

type UserResponse struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type LoginResponse struct {
	User    UserResponse `json:"user"`
	Message string       `json:"message"`
	Token   string       `json:"token,omitempty"`
}
