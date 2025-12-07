-- name: GetOperatorAssignment :one
SELECT * FROM operator_assignments
WHERE client_id = $1
  AND operator_id = $2
  AND deleted_at IS NULL;

-- name: GetOperatorAssignmentsByOperatorID :many
SELECT * FROM operator_assignments
WHERE operator_id = $1
  AND deleted_at IS NULL
  AND status = 'ACTIVE'
ORDER BY assigned_at DESC;

-- name: GetOperatorAssignmentsByClientID :many
SELECT * FROM operator_assignments
WHERE client_id = $1
  AND deleted_at IS NULL
ORDER BY assigned_at DESC;

-- name: CreateOperatorAssignment :one
INSERT INTO operator_assignments (
    client_id,
    operator_id,
    role,
    status,
    assigned_at
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: UpdateOperatorAssignment :one
UPDATE operator_assignments
SET
    role = COALESCE($3, role),
    status = COALESCE($4, status),
    unassigned_at = COALESCE($5, unassigned_at)
WHERE client_id = $1
  AND operator_id = $2
  AND deleted_at IS NULL
RETURNING *;

-- name: DeleteOperatorAssignment :exec
UPDATE operator_assignments
SET
    deleted_at = now(),
    deleted_by = $3
WHERE client_id = $1
  AND operator_id = $2
  AND deleted_at IS NULL;

