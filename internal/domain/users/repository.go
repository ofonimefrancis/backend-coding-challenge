package users

import "context"

// Finder interface defines methods for finding users by ID or email
type Finder interface {
	FindByID(ctx context.Context, id UserID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
}

// Creator interface defines methods for creating users
type Creator interface {
	Create(ctx context.Context, user *User) error
}

// Repository interface combines Finder and Creator interfaces
type Repository interface {
	Finder
	Creator
}
