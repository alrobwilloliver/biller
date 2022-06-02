package billingaccount

import (
	"context"

	"biller/lib/resource"
	"biller/svc/compute/store"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	log     *zap.Logger
	querier store.TxQuerier
	UnimplementedBillingAccountServiceServer
}

func NewServer(querier store.TxQuerier, log *zap.Logger) *server {
	return &server{
		querier: querier,
		log:     log,
	}
}

func (s *server) CreateBillingAccount(ctx context.Context, req *CreateBillingAccountRequest) (*BillingAccount, error) {
	var res BillingAccount

	tx, txq, err := s.querier.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return &res, err
	}
	defer tx.Rollback(ctx)

	nanoID, err := resource.NewNanoID(12)
	if err != nil {
		s.log.Error("could not generate nano id", zap.Error(err))
		return &res, err
	}
	newBillingAccount, err := txq.CreateBillingAccount(ctx, nanoID)
	if err != nil {
		s.log.Error("could not create account", zap.Error(err))
		return &res, status.Error(codes.Internal, codes.Internal.String())
	}

	if err != nil {
		s.log.Error("could not add billing account owner role", zap.Error(err))
		return &res, status.Error(codes.Internal, codes.Internal.String())
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Error("transaction failed when creating billing account", zap.Error(err))
		return &res, status.Error(codes.Internal, "creation failed")
	}

	return toBillingAccountPb(newBillingAccount), nil
}

func (s *server) GetBillingAccount(ctx context.Context, req *GetBillingAccountRequest) (*BillingAccount, error) {
	var res BillingAccount

	if !resource.ValidResourceID(req.Id) {
		return &res, status.Error(codes.InvalidArgument, "invalid id")
	}

	account, err := s.querier.FindBillingAccountById(ctx, req.Id)
	if err == pgx.ErrNoRows {
		return &res, status.Error(codes.NotFound, "billing account not found")
	}
	if err != nil {
		return &res, status.Error(codes.Internal, "internal service error")
	}

	return toBillingAccountPb(account), nil
}

func (s *server) ListBillingAccounts(ctx context.Context, req *ListBillingAccountsRequest) (*ListBillingAccountsResponse, error) {
	var res ListBillingAccountsResponse

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

	datacenter, err := txq.ListBillingAccounts(ctx, store.ListBillingAccountsParams{
		Ids:   ids,
		Limit: req.PageSize,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, codes.Internal.String())
	}

	res.PageSize = req.PageSize

	res.BillingAccounts = make([]*BillingAccount, len(datacenter))
	for i, row := range datacenter {
		res.BillingAccounts[i] = toBillingAccountPb(row)
	}
	return &res, nil
}

func toBillingAccountPb(in store.BillingAccount) *BillingAccount {
	out := BillingAccount{
		Id:            in.ID,
		CreateTime:    timestamppb.New(in.CreateTime),
		DemandEnabled: in.DemandEnabled,
		SupplyEnabled: in.SupplyEnabled,
	}
	return &out
}

func EnsureDemandEnabled(ctx context.Context, querier store.Querier, accountID string) error {
	account, err := querier.FindBillingAccountById(ctx, accountID)
	if err == pgx.ErrNoRows {
		return status.Error(codes.NotFound, "billing account not found")
	}
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	if !account.DemandEnabled {
		return status.Error(codes.PermissionDenied, "billing account not enabled for demand")
	}
	return nil
}

func EnsureSupplyEnabled(ctx context.Context, querier store.Querier, accountID string) error {
	account, err := querier.FindBillingAccountById(ctx, accountID)
	if err == pgx.ErrNoRows {
		return status.Error(codes.NotFound, "billing account not found")
	}
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	if !account.SupplyEnabled {
		return status.Error(codes.PermissionDenied, "billing account not enabled for supply")
	}
	return nil
}
