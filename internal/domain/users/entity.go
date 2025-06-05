package users

import (
	"net/mail"
	"strings"
	"thermondo/internal/domain/shared"
	"time"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type UserID string

// User represents a user entity
type User struct {
	ID        UserID    `json:"id" db:"id"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"password" db:"password"`
	Role      Role      `json:"role" db:"role"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// NewUser creates a new user entity
func NewUser(
	firstName string,
	lastName string,
	email string,
	password string,
	idGenerator shared.IDGenerator,
	timeProvider shared.TimeProvider,
	options ...UserOption,
) (*User, error) {

	user := User{
		ID:        UserID(idGenerator.Generate()),
		FirstName: strings.TrimSpace(firstName),
		LastName:  strings.TrimSpace(lastName),
		Email:     strings.ToLower(strings.TrimSpace(email)),
		Password:  password,
		Role:      RoleUser,
		IsActive:  true,
		CreatedAt: timeProvider.Now(),
		UpdatedAt: timeProvider.Now(),
	}

	// Validate the user before applying options
	if err := user.Validate(); err != nil {
		return nil, err
	}

	for _, option := range options {
		option(&user)
	}

	return &user, nil
}

func (u *User) Validate() error {
	if u.FirstName == "" {
		return ErrEmptyFirstName
	}

	if u.LastName == "" {
		return ErrEmptyLastName
	}

	if u.Email == "" {
		return ErrEmptyEmail
	}

	if u.Password == "" {
		return ErrEmptyPassword
	}

	if _, err := mail.ParseAddress(u.Email); err != nil {
		return ErrInvalidEmail
	}

	return nil
}
