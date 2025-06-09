package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"thermondo/internal/domain/users"
	"thermondo/internal/pkg/cache"

	"github.com/jmoiron/sqlx"
)

func setupUserTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("postgres", "postgres://postgres:postgres@localhost:5432/test_db?sslmode=disable")
	require.NoError(t, err)

	// Clean up the database before each test
	_, err = db.Exec(`
		TRUNCATE TABLE ratings CASCADE;
		TRUNCATE TABLE movies CASCADE;
		TRUNCATE TABLE users CASCADE;
	`)
	require.NoError(t, err)

	return db
}

func TestUserRepository_Create(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	mockCache := new(cache.MockCache)
	mockCache.On("DeletePattern", mock.Anything, "user:test-id-create:*").Return(nil)
	mockCache.On("Delete", mock.Anything, []string{"user:stats:test-id-create"}).Return(nil)

	repo := NewUserRepository(db, mockCache)

	user := &users.User{
		ID:        "test-id-create",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "test-create@example.com",
		Password:  "hashed_password",
		Role:      users.RoleUser,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	createdUser, err := repo.Create(context.Background(), user)
	require.NoError(t, err)
	assert.NotNil(t, createdUser)
	assert.Equal(t, user.ID, createdUser.ID)
	assert.Equal(t, user.Email, createdUser.Email)
	mockCache.AssertExpectations(t)
}

func TestUserRepository_FindByID(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	mockCache := new(cache.MockCache)

	repo := NewUserRepository(db, mockCache)

	expectedUser := &users.User{
		ID:        "test-id-find",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "test-find@example.com",
		Password:  "hashed_password",
		Role:      users.RoleUser,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Insert the user into the database
	_, err := db.Exec(`
		INSERT INTO users (id, first_name, last_name, email, password, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, expectedUser.ID, expectedUser.FirstName, expectedUser.LastName, expectedUser.Email, expectedUser.Password, expectedUser.Role, expectedUser.IsActive, expectedUser.CreatedAt, expectedUser.UpdatedAt)
	require.NoError(t, err)

	result, err := repo.FindByID(context.Background(), "test-id-find")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUser.ID, result.ID)
	assert.Equal(t, expectedUser.Email, result.Email)
	mockCache.AssertExpectations(t)
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	mockCache := new(cache.MockCache)

	repo := NewUserRepository(db, mockCache)

	expectedUser := &users.User{
		ID:        "test-id-find-email",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "test-find-email@example.com",
		Password:  "hashed_password",
		Role:      users.RoleUser,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Insert the user into the database
	_, err := db.Exec(`
		INSERT INTO users (id, first_name, last_name, email, password, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, expectedUser.ID, expectedUser.FirstName, expectedUser.LastName, expectedUser.Email, expectedUser.Password, expectedUser.Role, expectedUser.IsActive, expectedUser.CreatedAt, expectedUser.UpdatedAt)
	require.NoError(t, err)

	result, err := repo.FindByEmail(context.Background(), "test-find-email@example.com")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUser.ID, result.ID)
	assert.Equal(t, expectedUser.Email, result.Email)
	mockCache.AssertExpectations(t)
}

func TestUserRepository_List(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	mockCache := new(cache.MockCache)

	repo := NewUserRepository(db, mockCache)

	expectedUsers := []users.User{
		{
			ID:        "test-id-list-1",
			FirstName: "Alice",
			LastName:  "Smith",
			Email:     "test-list-1@example.com",
			Password:  "hashed_password1",
			Role:      users.RoleUser,
			IsActive:  true,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		{
			ID:        "test-id-list-2",
			FirstName: "Bob",
			LastName:  "Johnson",
			Email:     "test-list-2@example.com",
			Password:  "hashed_password2",
			Role:      users.RoleAdmin,
			IsActive:  true,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
	}

	// Insert the users into the database
	for _, u := range expectedUsers {
		_, err := db.Exec(`
			INSERT INTO users (id, first_name, last_name, email, password, role, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, u.ID, u.FirstName, u.LastName, u.Email, u.Password, u.Role, u.IsActive, u.CreatedAt, u.UpdatedAt)
		require.NoError(t, err)
	}

	result, err := repo.List(context.Background(), 0, 10)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, expectedUsers[0].ID, result[0].ID)
	assert.Equal(t, expectedUsers[1].ID, result[1].ID)
	mockCache.AssertExpectations(t)
}

func TestUserRepository_Count(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	mockCache := new(cache.MockCache)

	repo := NewUserRepository(db, mockCache)

	// Insert a user into the database
	_, err := db.Exec(`
		INSERT INTO users (id, first_name, last_name, email, password, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, "test-id-count", "John", "Doe", "test-count@example.com", "hashed_password", users.RoleUser, true, time.Now().UTC(), time.Now().UTC())
	require.NoError(t, err)

	count, err := repo.Count(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	mockCache.AssertExpectations(t)
}

func TestUserRepository_ErrorHandling(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	mockCache := new(cache.MockCache)

	repo := NewUserRepository(db, mockCache)

	// Test FindByID with non-existent ID
	user, err := repo.FindByID(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "sql: no rows in result set", err.Error())

	// Test FindByEmail with non-existent email
	user, err = repo.FindByEmail(context.Background(), "nonexistent@example.com")
	assert.NoError(t, err)
	assert.Nil(t, user)

	mockCache.AssertExpectations(t)
}
