// Code generated by sqlc. DO NOT EDIT.
// source: project.sql

package store

import (
	"context"
)

const createProject = `-- name: CreateProject :one
INSERT INTO "project" (id, billing_account_id)
VALUES (
    $1,
    $2
)
ON CONFLICT DO NOTHING
RETURNING id, create_time, billing_account_id
`

type CreateProjectParams struct {
	ID               string
	BillingAccountID string
}

func (q *Queries) CreateProject(ctx context.Context, arg CreateProjectParams) (Project, error) {
	row := q.db.QueryRow(ctx, createProject, arg.ID, arg.BillingAccountID)
	var i Project
	err := row.Scan(&i.ID, &i.CreateTime, &i.BillingAccountID)
	return i, err
}

const deleteProject = `-- name: DeleteProject :execrows
DELETE FROM "project"
WHERE
    id = $1
`

func (q *Queries) DeleteProject(ctx context.Context, id string) (int64, error) {
	result, err := q.db.Exec(ctx, deleteProject, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const findProjectById = `-- name: FindProjectById :one
SELECT id, create_time, billing_account_id
FROM "project"
WHERE
    id = $1
`

func (q *Queries) FindProjectById(ctx context.Context, id string) (Project, error) {
	row := q.db.QueryRow(ctx, findProjectById, id)
	var i Project
	err := row.Scan(&i.ID, &i.CreateTime, &i.BillingAccountID)
	return i, err
}

const getProjectCurrentSpend = `-- name: GetProjectCurrentSpend :one
SELECT uid, project_id, spend, start_time, end_time
FROM "project_spend"
WHERE project_id = $1
ORDER BY start_time DESC
LIMIT 1
`

func (q *Queries) GetProjectCurrentSpend(ctx context.Context, projectID string) (ProjectSpend, error) {
	row := q.db.QueryRow(ctx, getProjectCurrentSpend, projectID)
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

const getProjectSpendHistory = `-- name: GetProjectSpendHistory :many
SELECT uid, project_id, spend, start_time, end_time
FROM "project_spend"
WHERE project_id = $1
ORDER BY start_time DESC
`

func (q *Queries) GetProjectSpendHistory(ctx context.Context, projectID string) ([]ProjectSpend, error) {
	rows, err := q.db.Query(ctx, getProjectSpendHistory, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ProjectSpend
	for rows.Next() {
		var i ProjectSpend
		if err := rows.Scan(
			&i.Uid,
			&i.ProjectID,
			&i.Spend,
			&i.StartTime,
			&i.EndTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listProjects = `-- name: ListProjects :many
SELECT id, create_time, billing_account_id
FROM "project"
WHERE id = ANY($2::varchar[])
ORDER BY id
LIMIT $1
`

type ListProjectsParams struct {
	Limit int32
	Ids   []string
}

func (q *Queries) ListProjects(ctx context.Context, arg ListProjectsParams) ([]Project, error) {
	rows, err := q.db.Query(ctx, listProjects, arg.Limit, arg.Ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Project
	for rows.Next() {
		var i Project
		if err := rows.Scan(&i.ID, &i.CreateTime, &i.BillingAccountID); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const selectProjectForUpdate = `-- name: SelectProjectForUpdate :one
SELECT id, create_time, billing_account_id
FROM "project"
WHERE id = $1
FOR UPDATE
`

func (q *Queries) SelectProjectForUpdate(ctx context.Context, id string) (Project, error) {
	row := q.db.QueryRow(ctx, selectProjectForUpdate, id)
	var i Project
	err := row.Scan(&i.ID, &i.CreateTime, &i.BillingAccountID)
	return i, err
}

const updateProject = `-- name: UpdateProject :one
UPDATE "project" 
SET billing_account_id = $1
WHERE id = $2
RETURNING id, create_time, billing_account_id
`

type UpdateProjectParams struct {
	BillingAccountID string
	ID               string
}

func (q *Queries) UpdateProject(ctx context.Context, arg UpdateProjectParams) (Project, error) {
	row := q.db.QueryRow(ctx, updateProject, arg.BillingAccountID, arg.ID)
	var i Project
	err := row.Scan(&i.ID, &i.CreateTime, &i.BillingAccountID)
	return i, err
}
