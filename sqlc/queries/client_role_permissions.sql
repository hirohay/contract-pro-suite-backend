-- name: GetClientRolePermission :one
SELECT * FROM client_role_permissions
WHERE role_id = $1
  AND feature = $2
  AND action = $3
  AND deleted_at IS NULL;

-- name: GetClientRolePermissionsByRoleID :many
SELECT * FROM client_role_permissions
WHERE role_id = $1
  AND deleted_at IS NULL
ORDER BY feature, action;

-- name: GetClientRolePermissionsByFeatureAndAction :many
SELECT * FROM client_role_permissions
WHERE role_id = $1
  AND feature = $2
  AND action = $3
  AND deleted_at IS NULL
  AND granted = true;

-- name: CreateClientRolePermission :one
INSERT INTO client_role_permissions (
    role_id,
    feature,
    action,
    granted,
    conditions
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: UpdateClientRolePermission :one
UPDATE client_role_permissions
SET
    granted = COALESCE($4, granted),
    conditions = COALESCE($5, conditions)
WHERE role_id = $1
  AND feature = $2
  AND action = $3
  AND deleted_at IS NULL
RETURNING *;

-- name: DeleteClientRolePermission :exec
UPDATE client_role_permissions
SET
    deleted_at = now(),
    deleted_by = $4
WHERE role_id = $1
  AND feature = $2
  AND action = $3
  AND deleted_at IS NULL;

-- name: DeleteClientRolePermissionsByRoleID :exec
UPDATE client_role_permissions
SET
    deleted_at = now(),
    deleted_by = $2
WHERE role_id = $1
  AND deleted_at IS NULL;

