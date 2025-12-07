-- name: GetClientUserRole :one
SELECT * FROM client_user_roles
WHERE client_id = $1
  AND client_user_id = $2
  AND role_id = $3
  AND deleted_at IS NULL;

-- name: GetClientUserRolesByUserID :many
SELECT * FROM client_user_roles
WHERE client_id = $1
  AND client_user_id = $2
  AND deleted_at IS NULL
  AND revoked_at IS NULL  -- 有効なロールのみ
ORDER BY assigned_at DESC;

-- name: GetClientUserRolesByRoleID :many
SELECT * FROM client_user_roles
WHERE client_id = $1
  AND role_id = $2
  AND deleted_at IS NULL
  AND revoked_at IS NULL  -- 有効なロールのみ
ORDER BY assigned_at DESC;

-- name: CreateClientUserRole :one
INSERT INTO client_user_roles (
    client_id,
    client_user_id,
    role_id,
    assigned_at
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: RevokeClientUserRole :exec
UPDATE client_user_roles
SET
    revoked_at = now()
WHERE client_id = $1
  AND client_user_id = $2
  AND role_id = $3
  AND deleted_at IS NULL
  AND revoked_at IS NULL;

-- name: DeleteClientUserRole :exec
UPDATE client_user_roles
SET
    deleted_at = now(),
    deleted_by = $4
WHERE client_id = $1
  AND client_user_id = $2
  AND role_id = $3
  AND deleted_at IS NULL;

