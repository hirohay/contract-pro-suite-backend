-- name: GetOperator :one
SELECT * FROM operators
WHERE operator_id = $1
  AND deleted_at IS NULL;

-- name: GetOperatorByEmail :one
SELECT * FROM operators
WHERE email = $1
  AND deleted_at IS NULL;

-- name: ListOperators :many
SELECT * FROM operators
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateOperator :one
INSERT INTO operators (
    operator_id,
    email,
    first_name,
    last_name,
    status,
    mfa_enabled
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateOperator :one
UPDATE operators
SET
    email = COALESCE($2, email),
    first_name = COALESCE($3, first_name),
    last_name = COALESCE($4, last_name),
    status = COALESCE($5, status),
    mfa_enabled = COALESCE($6, mfa_enabled),
    last_login_at = COALESCE($7, last_login_at),
    password_changed_at = COALESCE($8, password_changed_at),
    updated_at = now()
WHERE operator_id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: DeleteOperator :exec
UPDATE operators
SET
    deleted_at = now(),
    deleted_by = $2,
    updated_at = now()
WHERE operator_id = $1
  AND deleted_at IS NULL;

