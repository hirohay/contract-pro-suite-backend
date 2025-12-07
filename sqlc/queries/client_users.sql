-- name: GetClientUser :one
SELECT * FROM client_users
WHERE client_user_id = $1
  AND client_id = $2
  AND deleted_at IS NULL;

-- name: GetClientUserByEmail :one
SELECT * FROM client_users
WHERE client_id = $1
  AND email = $2
  AND deleted_at IS NULL;

-- name: GetClientUserByUserIDOnly :one
SELECT * FROM client_users
WHERE client_user_id = $1
  AND deleted_at IS NULL;

-- name: ListClientUsers :many
SELECT * FROM client_users
WHERE client_id = $1
  AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateClientUser :one
INSERT INTO client_users (
    client_user_id,
    client_id,
    email,
    first_name,
    last_name,
    department,
    position,
    settings,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: UpdateClientUser :one
UPDATE client_users
SET
    email = COALESCE($3, email),
    first_name = COALESCE($4, first_name),
    last_name = COALESCE($5, last_name),
    department = COALESCE($6, department),
    position = COALESCE($7, position),
    settings = COALESCE($8, settings),
    status = COALESCE($9, status),
    updated_at = now()
WHERE client_user_id = $1
  AND client_id = $2
  AND deleted_at IS NULL
RETURNING *;

-- name: DeleteClientUser :exec
UPDATE client_users
SET
    deleted_at = now(),
    deleted_by = $3,
    updated_at = now()
WHERE client_user_id = $1
  AND client_id = $2
  AND deleted_at IS NULL;

