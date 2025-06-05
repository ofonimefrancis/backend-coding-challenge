package users

import (
	"testing"
	"thermondo/internal/platform/mocks"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	fixedTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	idGen := &mocks.MockIDGenerator{ID: "test-id-123"}
	timeProvider := &mocks.MockTimeProvider{Time: fixedTime}

	tests := []struct {
		name        string
		firstName   string
		lastName    string
		email       string
		password    string
		options     []UserOption
		wantErr     error
		checkResult func(t *testing.T, user *User)
	}{
		{
			name:      "valid user creation",
			firstName: "John",
			lastName:  "Doe",
			email:     "john.doe@example.com",
			password:  "password123",
			wantErr:   nil,
			checkResult: func(t *testing.T, user *User) {
				assert.Equal(t, UserID("test-id-123"), user.ID)
				assert.Equal(t, "John", user.FirstName)
				assert.Equal(t, "Doe", user.LastName)
				assert.Equal(t, "john.doe@example.com", user.Email)
				assert.Equal(t, RoleUser, user.Role)
				assert.True(t, user.IsActive)
				assert.Equal(t, fixedTime, user.CreatedAt)
				assert.Equal(t, fixedTime, user.UpdatedAt)
			},
		},
		{
			name:      "user with admin role option",
			firstName: "Jane",
			lastName:  "Admin",
			email:     "jane@example.com",
			password:  "admin123",
			options:   []UserOption{WithRole(RoleAdmin)},
			wantErr:   nil,
			checkResult: func(t *testing.T, user *User) {
				assert.Equal(t, RoleAdmin, user.Role)
				assert.Equal(t, "Jane", user.FirstName)
				assert.Equal(t, "Admin", user.LastName)
			},
		},
		{
			name:      "user with inactive status option",
			firstName: "Bob",
			lastName:  "Inactive",
			email:     "bob@example.com",
			password:  "password123",
			options:   []UserOption{WithActiveStatus(false)},
			wantErr:   nil,
			checkResult: func(t *testing.T, user *User) {
				assert.False(t, user.IsActive)
				assert.Equal(t, RoleUser, user.Role)
			},
		},
		{
			name:      "user with multiple options",
			firstName: "Alice",
			lastName:  "Multi",
			email:     "alice@example.com",
			password:  "password123",
			options:   []UserOption{WithRole(RoleAdmin), WithActiveStatus(false)},
			wantErr:   nil,
			checkResult: func(t *testing.T, user *User) {
				assert.Equal(t, RoleAdmin, user.Role)
				assert.False(t, user.IsActive)
			},
		},
		{
			name:      "trims whitespace and normalizes email",
			firstName: "  John  ",
			lastName:  "  Doe  ",
			email:     "  JOHN.DOE@EXAMPLE.COM  ",
			password:  "password123",
			wantErr:   nil,
			checkResult: func(t *testing.T, user *User) {
				assert.Equal(t, "John", user.FirstName)
				assert.Equal(t, "Doe", user.LastName)
				assert.Equal(t, "john.doe@example.com", user.Email)
			},
		},
		{
			name:      "empty first name",
			firstName: "",
			lastName:  "Doe",
			email:     "john@example.com",
			password:  "password",
			wantErr:   ErrEmptyFirstName,
		},
		{
			name:      "empty last name",
			firstName: "John",
			lastName:  "",
			email:     "john@example.com",
			password:  "password",
			wantErr:   ErrEmptyLastName,
		},
		{
			name:      "empty email",
			firstName: "John",
			lastName:  "Doe",
			email:     "",
			password:  "password",
			wantErr:   ErrEmptyEmail,
		},
		{
			name:      "empty password",
			firstName: "John",
			lastName:  "Doe",
			email:     "john@example.com",
			password:  "",
			wantErr:   ErrEmptyPassword,
		},
		{
			name:      "invalid email",
			firstName: "John",
			lastName:  "Doe",
			email:     "invalid-email",
			password:  "password",
			wantErr:   ErrInvalidEmail,
		},
		{
			name:      "whitespace only first name",
			firstName: "   ",
			lastName:  "Doe",
			email:     "john@example.com",
			password:  "password",
			wantErr:   ErrEmptyFirstName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(
				tt.firstName,
				tt.lastName,
				tt.email,
				tt.password,
				idGen,
				timeProvider,
				tt.options...,
			)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, user)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, user)
			if tt.checkResult != nil {
				tt.checkResult(t, user)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		user    *User
		wantErr error
	}{
		{
			name: "valid user",
			user: &User{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Password:  "password123",
			},
			wantErr: nil,
		},
		{
			name: "empty first name",
			user: &User{
				FirstName: "",
				LastName:  "Doe",
				Email:     "john@example.com",
				Password:  "password123",
			},
			wantErr: ErrEmptyFirstName,
		},
		{
			name: "empty last name",
			user: &User{
				FirstName: "John",
				LastName:  "",
				Email:     "john@example.com",
				Password:  "password123",
			},
			wantErr: ErrEmptyLastName,
		},
		{
			name: "empty email",
			user: &User{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "",
				Password:  "password123",
			},
			wantErr: ErrEmptyEmail,
		},
		{
			name: "empty password",
			user: &User{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Password:  "",
			},
			wantErr: ErrEmptyPassword,
		},
		{
			name: "invalid email format",
			user: &User{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "invalid-email",
				Password:  "password123",
			},
			wantErr: ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
