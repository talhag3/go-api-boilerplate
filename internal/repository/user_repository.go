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

// Custom errors that we define so we don't return database-specific errors to the handler.
var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailTaken   = errors.New("email already taken")
)

// UserRepository is an interface. My senior told me to make this an interface
// so we can mock it when writing unit tests. I still need to learn how to write Go tests!
type UserRepository interface {
	Create(ctx context.Context, in domain.CreateUserInput) (domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, error)
	List(ctx context.Context, limit, offset int32) ([]domain.User, error)
	Update(ctx context.Context, id uuid.UUID, in domain.UpdateUserInput) (domain.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// sqlcUserRepository is the concrete struct that actually talks to the database.
// It implements the UserRepository interface.
type sqlcUserRepository struct {
	q *sqlc.Queries // sqlc generates this for us!
}

// NewUserRepository is a constructor function. It returns the interface, not the concrete struct.
func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &sqlcUserRepository{q: sqlc.New(pool)}
}

// Create inserts a new user into the database.
func (r *sqlcUserRepository) Create(ctx context.Context, in domain.CreateUserInput) (domain.User, error) {
	// Call the sqlc-generated CreateUser function
	row, err := r.q.CreateUser(ctx, sqlc.CreateUserParams{
		FullName: in.FullName,
		Email:    in.Email,
	})
	if err != nil {
		// "23505" is the Postgres error code for unique constraint violation (like email already exists!).
		// We use errors.As to cast the error to a pgconn.PgError.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.User{}, fmt.Errorf("%w: email=%s", ErrEmailTaken, in.Email)
		}
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}
	// Convert the sqlc struct to our domain struct
	return toDomain(row), nil
}

// GetByID fetches a single user by their UUID.
func (r *sqlcUserRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	// sqlc expects a pgtype.UUID instead of google/uuid.UUID.
	// So we have to copy the bytes and set Valid to true!
	pgID := pgtype.UUID{Bytes: id, Valid: true}

	row, err := r.q.GetUser(ctx, pgID)
	if err != nil {
		// If pgx says no rows were returned, we return our custom ErrUserNotFound error
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("get user: %w", err)
	}
	return toDomain(row), nil
}

// List gets a list of users with pagination (limit and offset).
func (r *sqlcUserRepository) List(ctx context.Context, limit, offset int32) ([]domain.User, error) {
	rows, err := r.q.ListUsers(ctx, sqlc.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	// Map each sqlc database row to our domain User struct
	users := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, toDomain(row))
	}
	return users, nil
}

// Update updates a user's details.
func (r *sqlcUserRepository) Update(ctx context.Context, id uuid.UUID, in domain.UpdateUserInput) (domain.User, error) {
	// First, fetch the current state of the user.
	// This is important because if the user only wanted to update their email,
	// the Name field in `in` would be nil, and we don't want to overwrite their Name with an empty string!
	current, err := r.GetByID(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	fullName := current.FullName
	if in.FullName != nil {
		fullName = *in.FullName // Dereference the pointer to get the actual string value
	}

	email := current.Email
	if in.Email != nil {
		email = *in.Email // Dereference the pointer
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

// Delete deletes a user from the database.
func (r *sqlcUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	pgID := pgtype.UUID{Bytes: id, Valid: true}

	if err := r.q.DeleteUser(ctx, pgID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

// toDomain converts the sqlc-generated struct to our clean domain struct.
// This is nice because pgtype.UUID and pgtype.Timestamptz are database-specific types,
// and we don't want those leaking into our services or handlers!
func toDomain(u sqlc.User) domain.User {
	var createdAt, updatedAt time.Time

	// pgtype fields have a Valid flag. We must check it before using the value!
	if u.CreatedAt.Valid {
		createdAt = u.CreatedAt.Time
	}
	if u.UpdatedAt.Valid {
		updatedAt = u.UpdatedAt.Time
	}

	// Cast pgtype.UUID's Bytes back to google/uuid.UUID
	domainID := uuid.UUID(u.ID.Bytes)

	return domain.User{
		ID:        domainID,
		FullName:  u.FullName,
		Email:     u.Email,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}
