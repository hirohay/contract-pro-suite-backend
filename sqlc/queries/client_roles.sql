-- name: GetClientRole :one
SELECT * FROM client_roles
WHERE role_id = $1
  AND deleted_at IS NULL;

-- name: GetClientRoleByCode :one
SELECT * FROM client_roles
WHERE client_id = $1
  AND code = $2
  AND deleted_at IS NULL;

-- name: ListClientRoles :many
SELECT * FROM client_roles
WHERE client_id = $1
  AND deleted_at IS NULL
ORDER BY is_system DESC, created_at ASC;

-- name: CreateClientRole :one
INSERT INTO client_roles (
    role_id,
    client_id,
    code,
    name,
    description,
    is_system
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateClientRole :one
UPDATE client_roles
SET
    name = COALESCE($2, name),
    description = COALESCE($3, description)
WHERE role_id = $1
  AND deleted_at IS NULL
  AND is_system = false  -- システムロールは更新不可
RETURNING *;

-- name: DeleteClientRole :exec
UPDATE client_roles
SET
    deleted_at = now(),
    deleted_by = $2
WHERE role_id = $1
  AND deleted_at IS NULL
  AND is_system = false;  -- システムロールは削除不可

