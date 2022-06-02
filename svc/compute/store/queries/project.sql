-- name: CreateProject :one
INSERT INTO "project" (id, billing_account_id)
VALUES (
    @id,
    @billing_account_id
)
ON CONFLICT DO NOTHING
RETURNING *;

-- name: DeleteProject :execrows
DELETE FROM "project"
WHERE
    id = @id;

-- name: FindProjectById :one
SELECT *
FROM "project"
WHERE
    id = @id;

-- name: FindProjectExistsById :one
SELECT EXISTS (
    SELECT *
    FROM "project"
    WHERE
        id = @id
);

-- name: ListProjects :many
SELECT *
FROM "project"
WHERE id = ANY(@ids::varchar[])
ORDER BY id
LIMIT $1;

-- name: SelectProjectForUpdate :one
SELECT *
FROM "project"
WHERE id = @id
FOR UPDATE;

-- name: UpdateProject :one
UPDATE "project" 
SET billing_account_id = @billing_account_id
WHERE id = @id
RETURNING *;

-- name: GetProjectCurrentSpend :one
SELECT *
FROM "project_spend"
WHERE project_id = @project_id
ORDER BY start_time DESC
LIMIT 1;

-- name: GetProjectSpendHistory :many
SELECT *
FROM "project_spend"
WHERE project_id = @project_id
ORDER BY start_time DESC;
