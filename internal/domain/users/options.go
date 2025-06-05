package users

import "time"

type UserOption func(*User)

func WithRole(role Role) UserOption {
	return func(u *User) {
		u.Role = role
	}
}

func WithActiveStatus(isActive bool) UserOption {
	return func(u *User) {
		u.IsActive = isActive
	}
}

func WithPassword(password string) UserOption {
	return func(u *User) {
		u.Password = password
	}
}

func WithTimestamps(createdAt, updatedAt time.Time) UserOption {
	return func(u *User) {
		u.CreatedAt = createdAt
		u.UpdatedAt = updatedAt
	}
}
