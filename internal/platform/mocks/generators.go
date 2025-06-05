package mocks

import "time"

// Mock ID Generator implementation for testing
type MockIDGenerator struct {
	ID string
}

func (m *MockIDGenerator) Generate() string {
	return m.ID
}

// Mock Time Provider implementation for testing
type MockTimeProvider struct {
	Time time.Time
}

func (m *MockTimeProvider) Now() time.Time {
	return m.Time
}
