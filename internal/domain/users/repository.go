package users

import (
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	FindByID(ctx context.Context, id UserID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, page, limit int) ([]*User, error)
	Count(ctx context.Context) (int, error)
}
