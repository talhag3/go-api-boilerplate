package service

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/google/uuid"

	"github.com/talhag3/go-api-boilerplate/internal/domain"
	"github.com/talhag3/go-api-boilerplate/internal/repository"
)

// basic regex for email validation. I copied this regex pattern from Google!
var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// UserService is an interface that describes all the actions we can do with users.
// The HTTP handler talks to this service, so we keep handler and repo decoupled.
type UserService interface {
	Create(ctx context.Context, in domain.CreateUserInput) (domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, error)
	List(ctx context.Context, page, pageSize int) ([]domain.User, error)
	Update(ctx context.Context, id uuid.UUID, in domain.UpdateUserInput) (domain.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// userService is the actual implementation of the interface above.
type userService struct {
	repo repository.UserRepository // Needs the repository to store/retrieve data
	log  *slog.Logger              // Needs a logger to log events
}

// NewUserService is the constructor function for our user service.
func NewUserService(repo repository.UserRepository, log *slog.Logger) UserService {
	return &userService{
		repo: repo,
		log:  log,
	}
}

// Create handles validation and creation of a new user.
func (s *userService) Create(ctx context.Context, in domain.CreateUserInput) (domain.User, error) {
	// Clean up whitespace and convert email to lowercase so searching is case-insensitive!
	in.FullName = strings.TrimSpace(in.FullName)
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))

	// Simple validation rules before we do a database call
	if in.FullName == "" {
		return domain.User{}, ErrInvalidInput("full_name is required")
	}
	if !emailRegex.MatchString(in.Email) {
		return domain.User{}, ErrInvalidInput("invalid email format")
	}

	// Call repository to save the user
	user, err := s.repo.Create(ctx, in)
	if err != nil {
		// Log the error here so we can find it in our logs, then return it
		s.log.Error("failed to create user in DB", "email", in.Email, "error", err)
		return domain.User{}, err
	}

	s.log.Info("user created successfully", "id", user.ID)
	return user, nil
}

// GetByID retrieves a user by their UUID.
func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	// Don't waste time querying the DB if they sent an empty UUID
	if id == uuid.Nil {
		return domain.User{}, ErrInvalidInput("id is required")
	}

	return s.repo.GetByID(ctx, id)
}

// List gets a paginated list of users.
func (s *userService) List(ctx context.Context, page, pageSize int) ([]domain.User, error) {
	// Make sure the page is at least 1, we can't have page 0 or negative pages!
	if page < 1 {
		page = 1
	}
	// Limit page size to maximum 100 so users can't crash the server by asking for 1,000,000 users!
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20 // Default to 20
	}

	// Calculate the SQL OFFSET.
	// Example: page 1, pageSize 20 -> offset = (1-1)*20 = 0.
	// Example: page 2, pageSize 20 -> offset = (2-1)*20 = 20.
	offset := int32((page - 1) * pageSize)

	return s.repo.List(ctx, int32(pageSize), offset)
}

// Update validates updates and saves them.
func (s *userService) Update(ctx context.Context, id uuid.UUID, in domain.UpdateUserInput) (domain.User, error) {
	// Since fields are pointers, we check if they are not nil before validating them.
	if in.FullName != nil {
		*in.FullName = strings.TrimSpace(*in.FullName)
		if *in.FullName == "" {
			return domain.User{}, ErrInvalidInput("full_name cannot be empty")
		}
	}
	if in.Email != nil {
		*in.Email = strings.TrimSpace(strings.ToLower(*in.Email))
		if !emailRegex.MatchString(*in.Email) {
			return domain.User{}, ErrInvalidInput("invalid email format")
		}
	}

	return s.repo.Update(ctx, id, in)
}

// Delete removes a user.
func (s *userService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrInvalidInput("id is required")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("service delete failed: %w", err)
	}

	s.log.Info("user deleted", "id", id)
	return nil
}

// ErrInvalidInput is a custom error type we use for validation errors.
// This helps the HTTP handler know it was a 400 Bad Request instead of a 500 Server Error!
type ErrInvalidInput string

func (e ErrInvalidInput) Error() string {
	return string(e)
}
