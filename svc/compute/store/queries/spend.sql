-- name: CreateBillingAccountSpend :one
INSERT INTO "billing_account_spend" (uid, billing_account_id, spend, start_time, end_time)
VALUES (
    @uid,
    @billing_account_id,
    @spend,
    @start_time,
    @end_time
)
ON CONFLICT (billing_account_id, start_time, end_time)
  DO UPDATE SET spend = @spend
RETURNING *;

-- name: CreateOrderSpend :one
INSERT INTO "order_spend" (uid, order_id, spend, start_time, end_time)
VALUES (
    @uid,
    @order_id,
    @spend,
    @start_time,
    @end_time
)
ON CONFLICT (order_id, start_time, end_time)
  DO UPDATE SET spend = @spend
RETURNING *;

-- name: CreateProjectSpend :one
INSERT INTO "project_spend" (uid, project_id, spend, start_time, end_time)
VALUES (
    @uid,
    @project_id,
    @spend,
    @start_time,
    @end_time
)
ON CONFLICT (project_id, start_time, end_time)
  DO UPDATE SET spend = @spend
RETURNING *;

-- name: FindBillingAccountSpendForTimeRange :one
SELECT *
FROM "billing_account_spend"
WHERE billing_account_id = @billing_account_id
  AND start_time < @end_time
  AND end_time >= @start_time;
  
  -- name: FindOrderSpendForTimeRange :one
SELECT *
FROM "order_spend"
WHERE order_id = @order_id
  AND start_time < @end_time
  AND end_time >= @start_time;
  
  -- name: FindProjectSpendForTimeRange :one
SELECT *
FROM "project_spend"
WHERE project_id = @project_id
  AND start_time < @end_time
  AND end_time >= @start_time;
