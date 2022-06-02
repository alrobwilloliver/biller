-- name: CreateBillingAccount :one
INSERT INTO "billing_account" (id)
VALUES (
    @id
)
ON CONFLICT DO NOTHING
RETURNING *;

-- name: FindBillingAccountById :one
SELECT *
FROM "billing_account"
WHERE
    id = @id;

-- name: ListBillingAccounts :many
SELECT *
FROM "billing_account"
WHERE id = ANY(@ids::varchar[])
ORDER BY create_time
LIMIT $1;

-- name: ListAllBillingAccounts :many
SELECT *
FROM "billing_account"
ORDER BY create_time;

-- name: SelectBillingAccountForUpdate :one
SELECT *
FROM "billing_account"
WHERE id = @id
FOR UPDATE;

-- name: EnableBillingAccountDemand :one
UPDATE "billing_account"
SET demand_enabled = true
WHERE id = @id
RETURNING *;

-- name: EnableBillingAccountSupply :one
UPDATE "billing_account"
SET supply_enabled = true
WHERE id = @id
RETURNING *;
