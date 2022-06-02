package project

import (
	"context"

	"biller/lib/resource"
	"biller/svc/compute/billingaccount"
	"biller/svc/compute/store"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	log     *zap.Logger
	querier store.TxQuerier
	UnimplementedProjectServiceServer
}

func NewServer(querier store.TxQuerier, log *zap.Logger) *server {
	return &server{
		log:     log,
		querier: querier,
	}
}

func (s *server) CreateProject(ctx context.Context, req *CreateProjectRequest) (*Project, error) {
	var res Project

	if req.Project.Id == "" {
		return &res, status.Error(codes.InvalidArgument, "project id is required")
	}

	if !resource.ValidResourceID(req.Project.Id) {
		return &res, status.Error(codes.InvalidArgument, "invalid project id")
	}

	if !resource.ValidResourceID(req.Project.BillingAccountId) {
		return &res, status.Error(codes.InvalidArgument, "invalid account id")
	}

	tx, txq, err := s.querier.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return &res, err
	}
	defer tx.Rollback(ctx)

	err = billingaccount.EnsureDemandEnabled(ctx, txq, req.Project.BillingAccountId)
	if err != nil {
		return &res, err
	}

	exists, err := txq.FindProjectExistsById(ctx, req.Project.Id)
	if err != nil {
		return &res, status.Error(codes.Internal, "failed to check if project exists")
	}
	if exists {
		return &res, status.Error(codes.AlreadyExists, "a project with the requested id already exists")
	}

	newProject, err := txq.CreateProject(ctx, store.CreateProjectParams{
		BillingAccountID: req.Project.BillingAccountId,
		ID:               req.Project.Id,
	})
	if err != nil {
		return &res, status.Error(codes.Internal, "creation failed")
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Error("transaction failed when creating project", zap.Error(err))
		return &res, status.Error(codes.Internal, "creation failed")
	}
	return toProjectPb(newProject), nil
}

// TODO this would need to delete all resources in project???
// would still need to archive it for billing, auditing
func (s *server) DeleteProject(ctx context.Context, req *DeleteProjectRequest) (*emptypb.Empty, error) {
	var res emptypb.Empty

	if req.Id == "" {
		return &res, status.Error(codes.InvalidArgument, "project name is required")
	}

	tx, txq, err := s.querier.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return &res, err
	}
	defer tx.Rollback(ctx)

	project, err := txq.FindProjectById(ctx, req.Id)
	if err != nil {
		return &res, err
	}

	affected, err := txq.DeleteProject(ctx, project.ID)
	if err != nil {
		return &res, status.Errorf(codes.Internal, codes.Internal.String())
	}
	if affected == 0 {
		return &res, status.Error(codes.Internal, "project could not be deleted")
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Error("transaction failed when updating project", zap.Error(err))
		return &res, status.Error(codes.Internal, "update failed")
	}
	return &res, nil
}

func (s *server) GetProject(ctx context.Context, req *GetProjectRequest) (*Project, error) {
	var res Project

	if req.Id == "" {
		return &res, status.Error(codes.InvalidArgument, "project name is required")
	}

	tx, txq, err := s.querier.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return &res, err
	}
	defer tx.Rollback(ctx)

	item, err := txq.FindProjectById(ctx, req.Id)
	if err == pgx.ErrNoRows {
		return &res, status.Errorf(codes.NotFound, "project not found")
	}
	if err != nil {
		return &res, status.Errorf(codes.Internal, codes.Internal.String())
	}

	return toProjectPb(item), nil
}

func (s *server) ListProjects(ctx context.Context, req *ListProjectsRequest) (*ListProjectsResponse, error) {
	var res ListProjectsResponse

	if req.PageSize > 100 {
		req.PageSize = 100
	}
	if req.PageSize < 10 {
		req.PageSize = 10
	}

	tx, txq, err := s.querier.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return &res, err
	}
	defer tx.Rollback(ctx)

	ids := make([]string, 100)

	project, err := txq.ListProjects(ctx, store.ListProjectsParams{
		Ids:   ids,
		Limit: req.PageSize,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, codes.Internal.String())
	}

	res.PageSize = req.PageSize

	res.Projects = make([]*Project, len(project))
	for i, row := range project {
		res.Projects[i] = &Project{
			BillingAccountId: row.BillingAccountID,
			Id:               row.ID,
		}
	}
	return &res, nil
}

func (s *server) UpdateProject(ctx context.Context, req *UpdateProjectRequest) (*Project, error) {
	var res Project

	if !resource.ValidResourceID(req.Project.Id) {
		return &res, status.Error(codes.InvalidArgument, "invalid id")
	}

	var updated store.Project

	tx, txq, err := s.querier.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return &res, err
	}
	defer tx.Rollback(ctx)

	existing, err := txq.SelectProjectForUpdate(ctx, req.Project.Id)
	if err != nil {
		return &res, err
	}

	updates := store.UpdateProjectParams{
		BillingAccountID: existing.BillingAccountID,
	}

	if req.Project.BillingAccountId != existing.ID {
		err := billingaccount.EnsureDemandEnabled(ctx, txq, req.Project.BillingAccountId)
		if err != nil {
			return &res, err
		}
		updates.BillingAccountID = req.Project.BillingAccountId
	}

	updated, err = txq.UpdateProject(ctx, updates)
	if err != nil {
		s.log.Error("query failed when updating project", zap.Error(err))
		return &res, status.Error(codes.Internal, "update failed")
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Error("transaction failed when updating project", zap.Error(err))
		return &res, status.Error(codes.Internal, "update failed")
	}

	return toProjectPb(updated), nil
}

func (s *server) GetProjectCurrentSpend(ctx context.Context, req *GetProjectCurrentSpendRequest) (*ProjectSpend, error) {
	var res ProjectSpend

	if req.Id == "" {
		return &res, status.Error(codes.InvalidArgument, "project name is required")
	}

	tx, txq, err := s.querier.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return &res, err
	}
	defer tx.Rollback(ctx)

	project, err := txq.FindProjectById(ctx, req.Id)
	if err != nil {
		return &res, err
	}

	spend, err := txq.GetProjectCurrentSpend(ctx, project.ID)

	return &ProjectSpend{
		Uid:       spend.Uid.String(),
		ProjectId: spend.ProjectID,
		// Spend:     spend.Spend,
		StartTime: timestamppb.New(spend.StartTime),
		EndTime:   timestamppb.New(spend.EndTime),
	}, nil
}

func (s *server) GetProjectSpendHistory(ctx context.Context, req *GetProjectSpendHistoryRequest) (*GetProjectSpendHistoryResponse, error) {
	var res GetProjectSpendHistoryResponse

	if req.Id == "" {
		return &res, status.Error(codes.InvalidArgument, "project name is required")
	}

	tx, txq, err := s.querier.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return &res, err
	}
	defer tx.Rollback(ctx)

	project, err := txq.FindProjectById(ctx, req.Id)
	if err != nil {
		return &res, err
	}

	spend, err := txq.GetProjectSpendHistory(ctx, project.ID)

	res.ProjectSpendHistory = make([]*ProjectSpend, len(spend))
	for i, row := range spend {
		res.ProjectSpendHistory[i] = &ProjectSpend{
			Uid:       row.Uid.String(),
			ProjectId: row.ProjectID,
			// Spend:     float32(row.Spend),
			StartTime: timestamppb.New(row.StartTime),
			EndTime:   timestamppb.New(row.EndTime),
		}
	}
	return &res, nil
}

func toProjectPb(in store.Project) *Project {
	out := Project{
		BillingAccountId: in.BillingAccountID,
		Id:               in.ID,
	}
	return &out
}
