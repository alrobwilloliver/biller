// Code generated by sqlc. DO NOT EDIT.
// source: spend.sql

package store

import (
	"context"
	"time"

	apd "github.com/cockroachdb/apd/v2"
	"github.com/google/uuid"
)

const createBillingAccountSpend = `-- name: CreateBillingAccountSpend :one
INSERT INTO "billing_account_spend" (uid, billing_account_id, spend, start_time, end_time)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
ON CONFLICT (billing_account_id, start_time, end_time)
  DO UPDATE SET spend = $3
RETURNING uid, billing_account_id, spend, start_time, end_time
`

type CreateBillingAccountSpendParams struct {
	Uid              uuid.UUID
	BillingAccountID string
	Spend            apd.Decimal
	StartTime        time.Time
	EndTime          time.Time
}

func (q *Queries) CreateBillingAccountSpend(ctx context.Context, arg CreateBillingAccountSpendParams) (BillingAccountSpend, error) {
	row := q.db.QueryRow(ctx, createBillingAccountSpend,
		arg.Uid,
		arg.BillingAccountID,
		arg.Spend,
		arg.StartTime,
		arg.EndTime,
	)
	var i BillingAccountSpend
	err := row.Scan(
		&i.Uid,
		&i.BillingAccountID,
		&i.Spend,
		&i.StartTime,
		&i.EndTime,
	)
	return i, err
}

const createOrderSpend = `-- name: CreateOrderSpend :one
INSERT INTO "order_spend" (uid, order_id, spend, start_time, end_time)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
ON CONFLICT (order_id, start_time, end_time)
  DO UPDATE SET spend = $3
RETURNING uid, order_id, spend, start_time, end_time
`

type CreateOrderSpendParams struct {
	Uid       uuid.UUID
	OrderID   string
	Spend     apd.Decimal
	StartTime time.Time
	EndTime   time.Time
}

func (q *Queries) CreateOrderSpend(ctx context.Context, arg CreateOrderSpendParams) (OrderSpend, error) {
	row := q.db.QueryRow(ctx, createOrderSpend,
		arg.Uid,
		arg.OrderID,
		arg.Spend,
		arg.StartTime,
		arg.EndTime,
	)
	var i OrderSpend
	err := row.Scan(
		&i.Uid,
		&i.OrderID,
		&i.Spend,
		&i.StartTime,
		&i.EndTime,
	)
	return i, err
}

const createProjectSpend = `-- name: CreateProjectSpend :one
INSERT INTO "project_spend" (uid, project_id, spend, start_time, end_time)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
ON CONFLICT (project_id, start_time, end_time)
  DO UPDATE SET spend = $3
RETURNING uid, project_id, spend, start_time, end_time
`

type CreateProjectSpendParams struct {
	Uid       uuid.UUID
	ProjectID string
	Spend     apd.Decimal
	StartTime time.Time
	EndTime   time.Time
}

func (q *Queries) CreateProjectSpend(ctx context.Context, arg CreateProjectSpendParams) (ProjectSpend, error) {
	row := q.db.QueryRow(ctx, createProjectSpend,
		arg.Uid,
		arg.ProjectID,
		arg.Spend,
		arg.StartTime,
		arg.EndTime,
	)
	var i ProjectSpend
	err := row.Scan(
		&i.Uid,
		&i.ProjectID,
		&i.Spend,
		&i.StartTime,
		&i.EndTime,
	)
	return i, err
}

const findBillingAccountSpendForTimeRange = `-- name: FindBillingAccountSpendForTimeRange :one
SELECT uid, billing_account_id, spend, start_time, end_time
FROM "billing_account_spend"
WHERE billing_account_id = $1
  AND start_time < $2
  AND end_time >= $3
`

type FindBillingAccountSpendForTimeRangeParams struct {
	BillingAccountID string
	EndTime          time.Time
	StartTime        time.Time
}

func (q *Queries) FindBillingAccountSpendForTimeRange(ctx context.Context, arg FindBillingAccountSpendForTimeRangeParams) (BillingAccountSpend, error) {
	row := q.db.QueryRow(ctx, findBillingAccountSpendForTimeRange, arg.BillingAccountID, arg.EndTime, arg.StartTime)
	var i BillingAccountSpend
	err := row.Scan(
		&i.Uid,
		&i.BillingAccountID,
		&i.Spend,
		&i.StartTime,
		&i.EndTime,
	)
	return i, err
}

const findOrderSpendForTimeRange = `-- name: FindOrderSpendForTimeRange :one
SELECT uid, order_id, spend, start_time, end_time
FROM "order_spend"
WHERE order_id = $1
  AND start_time < $2
  AND end_time >= $3
`

type FindOrderSpendForTimeRangeParams struct {
	OrderID   string
	EndTime   time.Time
	StartTime time.Time
}

func (q *Queries) FindOrderSpendForTimeRange(ctx context.Context, arg FindOrderSpendForTimeRangeParams) (OrderSpend, error) {
	row := q.db.QueryRow(ctx, findOrderSpendForTimeRange, arg.OrderID, arg.EndTime, arg.StartTime)
	var i OrderSpend
	err := row.Scan(
		&i.Uid,
		&i.OrderID,
		&i.Spend,
		&i.StartTime,
		&i.EndTime,
	)
	return i, err
}

const findProjectSpendForTimeRange = `-- name: FindProjectSpendForTimeRange :one
SELECT uid, project_id, spend, start_time, end_time
FROM "project_spend"
WHERE project_id = $1
  AND start_time < $2
  AND end_time >= $3
`

type FindProjectSpendForTimeRangeParams struct {
	ProjectID string
	EndTime   time.Time
	StartTime time.Time
}

func (q *Queries) FindProjectSpendForTimeRange(ctx context.Context, arg FindProjectSpendForTimeRangeParams) (ProjectSpend, error) {
	row := q.db.QueryRow(ctx, findProjectSpendForTimeRange, arg.ProjectID, arg.EndTime, arg.StartTime)
	var i ProjectSpend
	err := row.Scan(
		&i.Uid,
		&i.ProjectID,
		&i.Spend,
		&i.StartTime,
		&i.EndTime,
	)
	return i, err
}