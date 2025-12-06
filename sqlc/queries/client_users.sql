-- name: GetClientUser :one
SELECT * FROM client_users
WHERE client_user_id = $1
  AND deleted_at IS NULL;

-- name: GetClientUserByEmail :one
SELECT * FROM client_users
WHERE client_id = $1
  AND email = $2
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
    email = COALESCE($2, email),
    first_name = COALESCE($3, first_name),
    last_name = COALESCE($4, last_name),
    department = COALESCE($5, department),
    position = COALESCE($6, position),
    settings = COALESCE($7, settings),
    status = COALESCE($8, status),
    updated_at = now()
WHERE client_user_id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: DeleteClientUser :exec
UPDATE client_users
SET
    deleted_at = now(),
    deleted_by = $2,
    updated_at = now()
WHERE client_user_id = $1
  AND deleted_at IS NULL;

