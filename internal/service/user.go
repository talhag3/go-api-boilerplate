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

// basic regex for email validation.
var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// UserService is the contract for our business logic.
// The HTTP handler will depend on this, not the repository directly.
type UserService interface {
	Create(ctx context.Context, in domain.CreateUserInput) (domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, error)
	List(ctx context.Context, page, pageSize int) ([]domain.User, error)
	Update(ctx context.Context, id uuid.UUID, in domain.UpdateUserInput) (domain.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type userService struct {
	repo repository.UserRepository
	log  *slog.Logger
}

// NewUserService sets up the service with its dependencies.
func NewUserService(repo repository.UserRepository, log *slog.Logger) UserService {
	return &userService{
		repo: repo,
		log:  log,
	}
}

func (s *userService) Create(ctx context.Context, in domain.CreateUserInput) (domain.User, error) {
	// don't save empty strings
	in.FullName = strings.TrimSpace(in.FullName)
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))

	// simple validation before hitting the database
	if in.FullName == "" {
		return domain.User{}, ErrInvalidInput("full_name is required")
	}
	if !emailRegex.MatchString(in.Email) {
		return domain.User{}, ErrInvalidInput("invalid email format")
	}

	user, err := s.repo.Create(ctx, in)
	if err != nil {
		// log the actual error for debugging, but pass the error up to the handler
		s.log.Error("failed to create user in DB", "email", in.Email, "error", err)
		return domain.User{}, err
	}

	s.log.Info("user created successfully", "id", user.ID)
	return user, nil
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	// quick check to prevent hitting the DB with an empty UUID
	if id == uuid.Nil {
		return domain.User{}, ErrInvalidInput("id is required")
	}

	return s.repo.GetByID(ctx, id)
}

func (s *userService) List(ctx context.Context, page, pageSize int) ([]domain.User, error) {
	// make sure we don't pass negative or zero values to the DB
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		// default to 20 if they ask for 0 or something crazy like 5000
		pageSize = 20
	}

	// calculate offset for pagination (page 1 -> offset 0, page 2 -> offset 20)
	offset := int32((page - 1) * pageSize)

	return s.repo.List(ctx, int32(pageSize), offset)
}

func (s *userService) Update(ctx context.Context, id uuid.UUID, in domain.UpdateUserInput) (domain.User, error) {
	// because we use pointers for partial updates, we only validate if they actually sent the field
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

// custom error type so the handler knows to return a 400 Bad Request instead of a 500
type ErrInvalidInput string

func (e ErrInvalidInput) Error() string {
	return string(e)
}
