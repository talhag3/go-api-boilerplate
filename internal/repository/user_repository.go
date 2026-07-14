/*
Concepts
Interfaces: UserRepository is a contract. The service depends on it, not on sqlcUserRepository.
*/

package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/talhag3/go-api-boilerplate/internal/db/sqlc"
	"github.com/talhag3/go-api-boilerplate/internal/domain"
)

// Custom errors
var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailTaken   = errors.New("email already taken")
)

// UserRepository is the contract the service layer depends on.
type UserRepository interface {
	Create(ctx context.Context, in domain.CreateUserInput) (domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, error)
	List(ctx context.Context, limit, offset int32) ([]domain.User, error)
	Update(ctx context.Context, id uuid.UUID, in domain.UpdateUserInput) (domain.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type sqlcUserRepository struct {
	q *sqlc.Queries
}

// NewUserRepository returns the interface, not the concrete struct.
// This makes it easy to swap out for a mock during tests.
func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &sqlcUserRepository{q: sqlc.New(pool)}
}

func (r *sqlcUserRepository) Create(ctx context.Context, in domain.CreateUserInput) (domain.User, error) {
	row, err := r.q.CreateUser(ctx, sqlc.CreateUserParams{
		FullName: in.FullName,
		Email:    in.Email,
	})
	if err != nil {
		// "23505" is the Postgres error code for unique constraint violation
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.User{}, fmt.Errorf("%w: email=%s", ErrEmailTaken, in.Email)
		}
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}
	return toDomain(row), nil
}

func (r *sqlcUserRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	// Translate google/uuid to pgtype.UUID for sqlc
	pgID := pgtype.UUID{Bytes: id, Valid: true}

	row, err := r.q.GetUser(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("get user: %w", err)
	}
	return toDomain(row), nil
}

func (r *sqlcUserRepository) List(ctx context.Context, limit, offset int32) ([]domain.User, error) {
	rows, err := r.q.ListUsers(ctx, sqlc.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	users := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, toDomain(row))
	}
	return users, nil
}

func (r *sqlcUserRepository) Update(ctx context.Context, id uuid.UUID, in domain.UpdateUserInput) (domain.User, error) {
	// Fetch current state to handle partial updates safely.
	// Otherwise, passing nil pointers would overwrite existing data with empty strings.
	current, err := r.GetByID(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	fullName := current.FullName
	if in.FullName != nil {
		fullName = *in.FullName
	}

	email := current.Email
	if in.Email != nil {
		email = *in.Email
	}

	pgID := pgtype.UUID{Bytes: id, Valid: true}
	row, err := r.q.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:       pgID,
		FullName: fullName,
		Email:    email,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.User{}, fmt.Errorf("%w: email=%s", ErrEmailTaken, email)
		}
		return domain.User{}, fmt.Errorf("update user: %w", err)
	}
	return toDomain(row), nil
}

func (r *sqlcUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	pgID := pgtype.UUID{Bytes: id, Valid: true}

	if err := r.q.DeleteUser(ctx, pgID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

// toDomain converts the sqlc generated struct to our pure domain struct.
// This keeps database types (like pgtype) from leaking into the rest of the app.
func toDomain(u sqlc.User) domain.User {
	var createdAt, updatedAt time.Time

	// pgtype.Timestamptz requires a validity check before accessing the time
	if u.CreatedAt.Valid {
		createdAt = u.CreatedAt.Time
	}
	if u.UpdatedAt.Valid {
		updatedAt = u.UpdatedAt.Time
	}

	// Convert pgtype.UUID back to google/uuid
	domainID := uuid.UUID(u.ID.Bytes)

	return domain.User{
		ID:        domainID,
		FullName:  u.FullName,
		Email:     u.Email,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}
