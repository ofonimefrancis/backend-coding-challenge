package repository

import (
	"context"
	"thermondo/internal/domain/users"

	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) users.Repository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id users.UserID) (*users.User, error) {
	return nil, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*users.User, error) {
	return nil, nil
}

func (r *userRepository) Create(ctx context.Context, user *users.User) error {
	return nil
}
