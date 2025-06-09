package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	domainUser "thermondo/internal/domain/users"
	"thermondo/internal/pkg/cache"

	"github.com/jmoiron/sqlx"
)

var (
	ErrUserNotCreated = errors.New("user was not created")
)

type userRepository struct {
	db    *sqlx.DB
	cache cache.Cache
}

func NewUserRepository(db *sqlx.DB, cache cache.Cache) domainUser.UserRepository {
	return &userRepository{
		db:    db,
		cache: cache,
	}
}

func (r *userRepository) FindByID(ctx context.Context, id domainUser.UserID) (*domainUser.User, error) {
	query := `SELECT id, first_name, last_name, email, role, is_active, created_at FROM users WHERE id = $1`
	user := &domainUser.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Role, &user.IsActive, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, err
	}
	return user, err
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domainUser.User, error) {
	query := `SELECT id, first_name, last_name, email, role, is_active, created_at FROM users WHERE email = $1`
	user := &domainUser.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Role, &user.IsActive, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *userRepository) Create(ctx context.Context, user *domainUser.User) (*domainUser.User, error) {
	query := `INSERT INTO users (id, first_name, last_name, email, password, role, is_active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	result, err := r.db.ExecContext(ctx, query, user.ID, user.FirstName, user.LastName, user.Email, user.Password, user.Role, user.IsActive, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrUserNotCreated
	}

	// Invalidate user cache
	if err := r.invalidateUserCache(ctx, user.ID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to invalidate user cache: %v\n", err)
	}

	return user, err
}

func (r *userRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM users`
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

func (r *userRepository) List(ctx context.Context, page, limit int) ([]*domainUser.User, error) {
	offset := 0
	if page > 0 {
		offset = (page - 1) * limit
	}
	query := `SELECT id, first_name, last_name, email, role, is_active, created_at FROM users ORDER BY id LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domainUser.User
	for rows.Next() {
		user := &domainUser.User{}
		if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Role, &user.IsActive, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// invalidateUserCache deletes all cached data for a user
func (r *userRepository) invalidateUserCache(ctx context.Context, userID domainUser.UserID) error {
	profilePattern := fmt.Sprintf("user:%s:*", userID)
	if err := r.cache.DeletePattern(ctx, profilePattern); err != nil {
		return fmt.Errorf("failed to delete profile cache: %w", err)
	}

	statsKey := fmt.Sprintf("user:stats:%s", userID)
	if err := r.cache.Delete(ctx, statsKey); err != nil {
		return fmt.Errorf("failed to delete stats cache: %w", err)
	}

	return nil
}
