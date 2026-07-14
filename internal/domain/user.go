package domain

import (
	"time"

	"github.com/google/uuid"
)

// User is the domain-level representation, independent of DB or HTTP.
// We keep this separate from sqlc-generated User so the business layer doesn't depend on the database representation directly.
type User struct {
	ID        uuid.UUID `json:"id"`         // Unique ID for the user. We use UUIDs instead of auto-incrementing integers for security!
	FullName  string    `json:"full_name"`  // The user's full name
	Email     string    `json:"email"`      // The user's email address
	CreatedAt time.Time `json:"created_at"` // When the user was created
	UpdatedAt time.Time `json:"updated_at"` // When the user was last updated
}

// CreateUserInput is the data needed to create a user (validated at edges).
// We don't need ID, CreatedAt, or UpdatedAt here because the database generates them.
type CreateUserInput struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

// UpdateUserInput allows partial updates.
// We use pointers here (*string) because if a field is not sent in the JSON request,
// the pointer will be nil. If we used normal strings, they would default to "",
// and we wouldn't know if the user wanted to set their name to empty or if they just didn't want to update it!
type UpdateUserInput struct {
	FullName *string `json:"full_name,omitempty"`
	Email    *string `json:"email,omitempty"`
}
