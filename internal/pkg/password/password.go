package password

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidPassword = errors.New("invalid password")
)

// HashPassword takes a plain text password and returns a bcrypt hash
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	saltedPassword := append([]byte(password), salt...)

	hash, err := bcrypt.GenerateFromPassword(saltedPassword, bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	combined := append(salt, hash...)
	return base64.StdEncoding.EncodeToString(combined), nil
}

// VerifyPassword checks if a plain text password matches a stored hash
func VerifyPassword(password, storedHash string) error {
	decoded, err := base64.StdEncoding.DecodeString(storedHash)
	if err != nil {
		return fmt.Errorf("invalid stored hash: %w", err)
	}

	// Extract salt and hash
	if len(decoded) < 16 {
		return ErrInvalidPassword
	}
	salt := decoded[:16]
	hash := decoded[16:]

	saltedPassword := append([]byte(password), salt...)
	if err := bcrypt.CompareHashAndPassword(hash, saltedPassword); err != nil {
		return ErrInvalidPassword
	}

	return nil
}
