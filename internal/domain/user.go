package domain

import (
	"time"

	"github.com/google/uuid"
)

// User is the domain-level representation, independent of DB or HTTP.
// Separate from sqlc-generated User so the business layer

type User struct {
	ID        uuid.UUID `json:"id"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserInput is the data needed to create a user (validated at edges).
type CreateUserInput struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

// UpdateUserInput allows partial updates (pointers = optional).
type UpdateUserInput struct {
	FullName *string `json:"full_name,omitempty"`
	Email    *string `json:"email,omitempty"`
}
