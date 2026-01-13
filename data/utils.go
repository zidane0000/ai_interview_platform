package data

import (
	"github.com/google/uuid"
)

// GenerateID generates a new UUID string
func GenerateID() string {
	return uuid.New().String()
}
