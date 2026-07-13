-- name: GetUser :one
SELECT id, full_name, email, created_at, updated_at
FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT id, full_name, email, created_at, updated_at
FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateUser :one
INSERT INTO users (full_name, email)
VALUES ($1, $2)
RETURNING id, full_name, email, created_at, updated_at;

-- name: UpdateUser :one
UPDATE users
SET full_name = $2,
    email     = $3,
    updated_at = now()
WHERE id = $1
RETURNING id, full_name, email, created_at, updated_at;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;