package users

import "errors"

var (
	ErrInvalidEmail      = errors.New("invalid email address")
	ErrEmptyEmail        = errors.New("email cannot be empty")
	ErrEmptyFirstName    = errors.New("first name cannot be empty")
	ErrEmptyLastName     = errors.New("last name cannot be empty")
	ErrEmptyPassword     = errors.New("password cannot be empty")
	ErrUserAlreadyExists = errors.New("user already exists")
)
