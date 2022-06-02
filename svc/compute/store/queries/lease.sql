-- name: FindLeaseInfoByLeaseId :one
SELECT *
FROM "lease" lease
         INNER JOIN "order" o ON lease.order_id = o.id
WHERE lease.id = @id
LIMIT 1;

-- name: ListLeasesForTimeRangeByOrderId :many
SELECT *
FROM "lease" l
WHERE
  l.order_id = @order_id AND
  l.create_time < @end_time AND
  (l.end_time IS NULL OR l.end_time >= @start_time);

-- name: ListActiveLeasesByOrderId :many
SELECT *
FROM "lease"
WHERE lease.order_id = @order_id
  AND status = 'active';

-- name: ListOrdersByProjectId :many
SELECT *
FROM "order" o
WHERE o.project_id = @project_id;

-- name: ListOrdersByBillingAccountId :many
SELECT *
FROM "order" o
WHERE o.billing_account_id = @billing_account_id;

-- name: CreateOrder :one
INSERT INTO "order" (
  id,
  infra_type,
  billing_account_id,
  project_id,
  quantity,
  description,
  price_hr
)
VALUES (
  @id,
  @infra_type,
  @billing_account_id,
  @project_id,
  @quantity,
  @description,
  @price_hr
)
ON CONFLICT DO NOTHING
RETURNING *;

-- name: CreateLease :one
INSERT INTO "lease" (
  id,
  infra_type,
  order_id,
  price_hr
)
VALUES (
  @id,
  @infra_type,
  @order_id,
  @price_hr
)
ON CONFLICT DO NOTHING
RETURNING *;

-- name: EndLease :one
UPDATE "lease"
SET end_time = NOW(),
    status = @status
WHERE id = @id
RETURNING *;

-- name: EndOrder :one
UPDATE "order"
SET status = @status
WHERE id = @id
RETURNING *;
