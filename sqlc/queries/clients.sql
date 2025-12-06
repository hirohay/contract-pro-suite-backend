-- name: GetClient :one
SELECT * FROM clients
WHERE client_id = $1
  AND deleted_at IS NULL;

-- name: GetClientBySlug :one
SELECT * FROM clients
WHERE slug = $1
  AND deleted_at IS NULL;

-- name: GetClientByCompanyCode :one
SELECT * FROM clients
WHERE company_code = $1
  AND deleted_at IS NULL;

-- name: ListClients :many
SELECT * FROM clients
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateClient :one
INSERT INTO clients (
    slug,
    company_code,
    name,
    e_sign_mode,
    retention_default_months,
    status,
    settings
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateClient :one
UPDATE clients
SET
    slug = COALESCE($2, slug),
    company_code = COALESCE($3, company_code),
    name = COALESCE($4, name),
    e_sign_mode = COALESCE($5, e_sign_mode),
    retention_default_months = COALESCE($6, retention_default_months),
    status = COALESCE($7, status),
    settings = COALESCE($8, settings),
    updated_at = now()
WHERE client_id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: DeleteClient :exec
UPDATE clients
SET
    deleted_at = now(),
    deleted_by = $2,
    updated_at = now()
WHERE client_id = $1
  AND deleted_at IS NULL;

