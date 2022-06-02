package billingaccount

import (
	"context"
	"testing"
	"time"

	"biller/svc/compute/store"

	"github.com/cockroachdb/apd/v2"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap/zaptest"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FakeTxQuerier struct {
	store.TxQuerier
	billingAccountSpend               apd.Decimal
	createBillingAccountSpend         store.BillingAccountSpend
	createBillingAccountSpendError    error
	createBillingAccount              store.BillingAccount
	createOrderSpend                  store.OrderSpend
	createOrderSpendError             error
	createProjectSpend                store.ProjectSpend
	createProjectSpendError           error
	getBillingAccount                 store.BillingAccount
	listBillingAccounts               []store.BillingAccount
	listOrdersByBillingAccountIdError error
	leasesForTimeRange                []store.Lease
	leasesForTimeRangeError           error
	orders                            []store.Order
	projectSpend                      apd.Decimal
	orderSpend                        apd.Decimal
	err                               error
}

func (txq FakeTxQuerier) CreateBillingAccount(ctx context.Context, id string) (store.BillingAccount, error) {
	return txq.createBillingAccount, nil
}

func (txq FakeTxQuerier) FindBillingAccountById(ctx context.Context, id string) (store.BillingAccount, error) {
	return txq.getBillingAccount, nil
}

func (txq FakeTxQuerier) ListBillingAccounts(ctx context.Context, arg store.ListBillingAccountsParams) ([]store.BillingAccount, error) {
	return txq.listBillingAccounts, nil
}

type FakeTx struct {
	pgx.Tx
}

func (tx FakeTx) Rollback(context.Context) error {
	return nil
}

func (tx FakeTx) Commit(ctx context.Context) error {
	return nil
}

func (q FakeTxQuerier) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, store.Querier, error) {
	return FakeTx{}, &q, nil
}

func (q FakeTxQuerier) ExecWithTx(context.Context, pgx.TxOptions, func(store.Querier) error) error {
	return nil
}

func Test_CreateBillingAccount(t *testing.T) {
	t.Run("should create billing account", func(t *testing.T) {

		logger := zaptest.NewLogger(t)
		var querier FakeTxQuerier
		id := "d543d430-0b1a-4255-868d-af2f08ee4c41"
		createTime := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
		querier.createBillingAccount = store.BillingAccount{
			ID:            id,
			CreateTime:    createTime,
			SupplyEnabled: false,
			DemandEnabled: false,
		}
		server := NewServer(&querier, logger)

		res, err := server.CreateBillingAccount(context.Background(), &CreateBillingAccountRequest{})
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if res.Id != "d543d430-0b1a-4255-868d-af2f08ee4c41" {
			t.Errorf("expected empty id, got: %s", res.Id)
		}
		if res.CreateTime.AsTime().String() != timestamppb.New(createTime).AsTime().UTC().String() {
			t.Errorf("expected create time %v, got: %v", timestamppb.New(createTime).AsTime().UTC().String(), res.CreateTime.AsTime().String())
		}
		if res.SupplyEnabled {
			t.Errorf("expected supply enabled false, got: %v", res.SupplyEnabled)
		}
		if res.DemandEnabled {
			t.Errorf("expected demand enabled false, got: %v", res.DemandEnabled)
		}
	})
}

func Test_GetBillingAccount(t *testing.T) {
	t.Run("should get billing account", func(t *testing.T) {

		logger := zaptest.NewLogger(t)
		var querier FakeTxQuerier
		id := "d543d430-0b1a-4255-868d-af2f08ee4c41"
		createTime := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
		querier.getBillingAccount = store.BillingAccount{
			ID:            id,
			CreateTime:    createTime,
			SupplyEnabled: false,
			DemandEnabled: false,
		}
		server := NewServer(&querier, logger)

		res, err := server.GetBillingAccount(context.Background(), &GetBillingAccountRequest{
			Id: "d543d430-0b1a-4255-868d-af2f08ee4c41",
		})
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if res.Id != "d543d430-0b1a-4255-868d-af2f08ee4c41" {
			t.Errorf("expected empty id, got: %s", res.Id)
		}
		if res.CreateTime.AsTime().String() != timestamppb.New(createTime).AsTime().UTC().String() {
			t.Errorf("expected create time %v, got: %v", timestamppb.New(createTime).AsTime().UTC().String(), res.CreateTime.AsTime().String())
		}
		if res.SupplyEnabled {
			t.Errorf("expected supply enabled false, got: %v", res.SupplyEnabled)
		}
		if res.DemandEnabled {
			t.Errorf("expected demand enabled false, got: %v", res.DemandEnabled)
		}
	})
}

func Test_ListBillingAccounts(t *testing.T) {
	t.Run("should list billing accounts", func(t *testing.T) {

		logger := zaptest.NewLogger(t)
		var querier FakeTxQuerier
		id := "d543d430-0b1a-4255-868d-af2f08ee4c41"
		createTime := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
		querier.listBillingAccounts = []store.BillingAccount{
			{
				ID:            id,
				CreateTime:    time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				SupplyEnabled: false,
				DemandEnabled: false,
			},
		}
		server := NewServer(&querier, logger)

		res, err := server.ListBillingAccounts(context.Background(), &ListBillingAccountsRequest{})
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(res.BillingAccounts) != 1 {
			t.Errorf("expected 1 billing account, got: %d", len(res.BillingAccounts))
		}
		if res.BillingAccounts[0].Id != "d543d430-0b1a-4255-868d-af2f08ee4c41" {
			t.Errorf("expected empty id, got: %s", res.BillingAccounts[0].Id)
		}
		if res.BillingAccounts[0].CreateTime.AsTime().String() != timestamppb.New(createTime).AsTime().UTC().String() {
			t.Errorf("expected create time %v, got: %v", timestamppb.New(createTime).AsTime().UTC().String(), res.BillingAccounts[0].CreateTime.AsTime().String())
		}
		if res.BillingAccounts[0].SupplyEnabled {
			t.Errorf("expected supply enabled false, got: %v", res.BillingAccounts[0].SupplyEnabled)
		}
		if res.BillingAccounts[0].DemandEnabled {
			t.Errorf("expected payment enabled false, got: %v", res.BillingAccounts[0].DemandEnabled)
		}
	})
}
