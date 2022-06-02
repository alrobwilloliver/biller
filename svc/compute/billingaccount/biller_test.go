package billingaccount

import (
	"context"
	"database/sql"
	"errors"
	"math/big"
	"testing"
	"time"

	"biller/lib/conv"
	"biller/svc/compute/store"

	"github.com/cockroachdb/apd/v2"
	"go.uber.org/zap/zaptest"
)

func (txq FakeTxQuerier) ListOrdersByBillingAccountId(ctx context.Context, billingAccountID string) ([]store.Order, error) {
	return txq.orders, txq.listOrdersByBillingAccountIdError
}

func (txq FakeTxQuerier) ListLeasesForTimeRangeByOrderId(ctx context.Context, arg store.ListLeasesForTimeRangeByOrderIdParams) ([]store.Lease, error) {
	return txq.leasesForTimeRange, txq.leasesForTimeRangeError
}

func (txq *FakeTxQuerier) CreateProjectSpend(ctx context.Context, arg store.CreateProjectSpendParams) (store.ProjectSpend, error) {
	txq.projectSpend = arg.Spend
	return txq.createProjectSpend, txq.createProjectSpendError
}

func (txq *FakeTxQuerier) CreateOrderSpend(ctx context.Context, arg store.CreateOrderSpendParams) (store.OrderSpend, error) {
	txq.orderSpend = arg.Spend
	return txq.createOrderSpend, txq.createOrderSpendError
}

func (txq *FakeTxQuerier) CreateBillingAccountSpend(ctx context.Context, arg store.CreateBillingAccountSpendParams) (store.BillingAccountSpend, error) {
	txq.billingAccountSpend = arg.Spend
	return txq.createBillingAccountSpend, txq.createBillingAccountSpendError
}

func Test_calculateDemandSpend(t *testing.T) {
	t.Run("should fail when ListOrdersByBillingAccountId query returns an error", func(t *testing.T) {
		var querier FakeTxQuerier
		querier.listOrdersByBillingAccountIdError = errors.New("list orders by billing account id error")
		biller := NewBiller(&querier, zaptest.NewLogger(t))

		ctx := context.Background()
		billingAccounts := []store.BillingAccount{
			{
				ID:            "1",
				CreateTime:    time.Date(2020, time.January, 3, 0, 0, 0, 0, time.UTC),
				SupplyEnabled: true,
				DemandEnabled: true,
			},
			{
				ID:            "2",
				CreateTime:    time.Date(2020, time.January, 5, 0, 0, 0, 0, time.UTC),
				SupplyEnabled: true,
				DemandEnabled: true,
			},
		}
		err := biller.calculateDemandSpend(ctx, billingAccounts, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, time.January, 30, 0, 0, 0, 0, time.UTC))
		if err == nil {
			t.Error("expected error, got nil")
		}
		if err.Error() != "error when listing orders for billing account: list orders by billing account id error" {
			t.Errorf("expected error message to be '%s', got %v", "error when listing orders for billing account: list orders by billing account id error", err.Error())
		}
	})
	t.Run("should fail when ListLeasesForTimeRangeByOrderId", func(t *testing.T) {
		var querier FakeTxQuerier
		querier.leasesForTimeRangeError = errors.New("list orders by order id error")
		querier.orders = []store.Order{
			{
				ID:               "1",
				BillingAccountID: "1",
				ProjectID:        "1",
				PriceHr:          100,
			},
			{
				ID:               "2",
				BillingAccountID: "1",
				ProjectID:        "2",
				PriceHr:          150,
			},
		}
		biller := NewBiller(&querier, zaptest.NewLogger(t))

		ctx := context.Background()
		billingAccounts := []store.BillingAccount{
			{
				ID:            "1",
				CreateTime:    time.Date(2020, time.January, 3, 0, 0, 0, 0, time.UTC),
				SupplyEnabled: true,
				DemandEnabled: true,
			},
			{
				ID:            "2",
				CreateTime:    time.Date(2020, time.January, 5, 0, 0, 0, 0, time.UTC),
				SupplyEnabled: true,
				DemandEnabled: true,
			},
		}
		err := biller.calculateDemandSpend(ctx, billingAccounts, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, time.January, 30, 0, 0, 0, 0, time.UTC))
		if err == nil {
			t.Error("expected error, got nil")
		}
		if err.Error() != "list leases for time range by order id failed: list orders by order id error" {
			t.Errorf("expected error message to be '%s', got %v", "list leases for time range by order id failed: list orders by order id error", err.Error())
		}
	})
	t.Run("should fail when CreateOrderSpend returns an error", func(t *testing.T) {
		var querier FakeTxQuerier
		querier.orders = []store.Order{
			{
				ID:               "1",
				BillingAccountID: "1",
				ProjectID:        "1",
				PriceHr:          100,
			},
			{
				ID:               "2",
				BillingAccountID: "1",
				ProjectID:        "2",
				PriceHr:          150,
			},
		}
		querier.leasesForTimeRange = []store.Lease{
			{
				ID:         "1",
				OrderID:    "1",
				CreateTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				EndTime: sql.NullTime{
					Time:  time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
					Valid: true,
				},
				PriceHr: 100,
			},
			{
				ID:         "2",
				OrderID:    "1",
				CreateTime: time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
				PriceHr:    100,
			},
			{
				ID:         "3",
				OrderID:    "2",
				CreateTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				EndTime: sql.NullTime{
					Time:  time.Date(2020, time.January, 3, 0, 0, 0, 0, time.UTC),
					Valid: true,
				},
				PriceHr: 150,
			},
			{
				ID:         "4",
				OrderID:    "2",
				CreateTime: time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
				PriceHr:    10,
			},
		}
		querier.createOrderSpendError = errors.New("create order spend error")
		biller := NewBiller(&querier, zaptest.NewLogger(t))

		ctx := context.Background()
		billingAccounts := []store.BillingAccount{
			{
				ID:            "1",
				CreateTime:    time.Date(2020, time.January, 3, 0, 0, 0, 0, time.UTC),
				SupplyEnabled: true,
				DemandEnabled: true,
			},
			{
				ID:            "2",
				CreateTime:    time.Date(2020, time.January, 5, 0, 0, 0, 0, time.UTC),
				SupplyEnabled: true,
				DemandEnabled: true,
			},
		}
		err := biller.calculateDemandSpend(ctx, billingAccounts, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, time.January, 30, 0, 0, 0, 0, time.UTC))
		if err == nil {
			t.Error("expected error, got nil")
		}
		if err.Error() != "create order spend failed: create order spend error" {
			t.Errorf("expected error message to be '%s', got %v", "create order spend failed: create order spend error", err.Error())
		}
	})
	t.Run("should fail when CreateProjectSpend returns an error", func(t *testing.T) {
		var querier FakeTxQuerier
		querier.orders = []store.Order{
			{
				ID:               "1",
				BillingAccountID: "1",
				ProjectID:        "1",
				PriceHr:          100,
			},
			{
				ID:               "2",
				BillingAccountID: "1",
				ProjectID:        "2",
				PriceHr:          150,
			},
			{
				ID:               "3",
				BillingAccountID: "2",
				ProjectID:        "1",
				PriceHr:          100,
			},
		}
		querier.leasesForTimeRange = []store.Lease{
			{
				ID:         "1",
				OrderID:    "1",
				CreateTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				EndTime: sql.NullTime{
					Time:  time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
					Valid: true,
				},
				PriceHr: 100,
			},
			{
				ID:         "2",
				OrderID:    "1",
				CreateTime: time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
				PriceHr:    100,
			},
			{
				ID:         "3",
				OrderID:    "2",
				CreateTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				EndTime: sql.NullTime{
					Time:  time.Date(2020, time.January, 3, 0, 0, 0, 0, time.UTC),
					Valid: true,
				},
				PriceHr: 150,
			},
			{
				ID:         "4",
				OrderID:    "2",
				CreateTime: time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
				PriceHr:    10,
			},
		}
		querier.createProjectSpendError = errors.New("create project spend error")
		biller := NewBiller(&querier, zaptest.NewLogger(t))

		ctx := context.Background()
		billingAccounts := []store.BillingAccount{
			{
				ID:            "1",
				CreateTime:    time.Date(2020, time.January, 3, 0, 0, 0, 0, time.UTC),
				SupplyEnabled: true,
				DemandEnabled: true,
			},
			{
				ID:            "2",
				CreateTime:    time.Date(2020, time.January, 5, 0, 0, 0, 0, time.UTC),
				SupplyEnabled: true,
				DemandEnabled: true,
			},
		}
		err := biller.calculateDemandSpend(ctx, billingAccounts, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, time.January, 30, 0, 0, 0, 0, time.UTC))
		if err == nil {
			t.Error("expected error, got nil")
		}
		if err.Error() != "create project spend failed: create project spend error" {
			t.Errorf("expected error message to be '%s', got %v", "create project spend failed: create project spend error", err.Error())
		}
	})
	t.Run("should fail when the CreateBillingAccountSpend returns an error", func(t *testing.T) {
		var querier FakeTxQuerier
		querier.orders = []store.Order{
			{
				ID:               "1",
				BillingAccountID: "1",
				ProjectID:        "1",
				PriceHr:          100,
			},
			{
				ID:               "2",
				BillingAccountID: "1",
				ProjectID:        "2",
				PriceHr:          150,
			},
			{
				ID:               "3",
				BillingAccountID: "2",
				ProjectID:        "1",
				PriceHr:          100,
			},
		}
		querier.leasesForTimeRange = []store.Lease{
			{
				ID:         "1",
				OrderID:    "1",
				CreateTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				EndTime: sql.NullTime{
					Time:  time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
					Valid: true,
				},
				PriceHr: 100,
			},
			{
				ID:         "2",
				OrderID:    "1",
				CreateTime: time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
				PriceHr:    100,
			},
			{
				ID:         "3",
				OrderID:    "2",
				CreateTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				EndTime: sql.NullTime{
					Time:  time.Date(2020, time.January, 3, 0, 0, 0, 0, time.UTC),
					Valid: true,
				},
				PriceHr: 150,
			},
			{
				ID:         "4",
				OrderID:    "2",
				CreateTime: time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
				PriceHr:    10,
			},
		}
		querier.createBillingAccountSpendError = errors.New("create billing account spend error")
		biller := NewBiller(&querier, zaptest.NewLogger(t))

		ctx := context.Background()
		billingAccounts := []store.BillingAccount{
			{
				ID:            "1",
				CreateTime:    time.Date(2020, time.January, 3, 0, 0, 0, 0, time.UTC),
				SupplyEnabled: true,
				DemandEnabled: true,
			},
			{
				ID:            "2",
				CreateTime:    time.Date(2020, time.January, 5, 0, 0, 0, 0, time.UTC),
				SupplyEnabled: true,
				DemandEnabled: true,
			},
		}
		err := biller.calculateDemandSpend(ctx, billingAccounts, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, time.January, 30, 0, 0, 0, 0, time.UTC))
		if err == nil {
			t.Error("expected error, got nil")
		}
		if err.Error() != "create billing account spend failed: create billing account spend error" {
			t.Errorf("expected error message to be '%s', got %v", "create billing account spend failed: create billing account spend error", err.Error())
		}
	})
	t.Run("should calculate spend for all models when end time on lease is defined", func(t *testing.T) {
		var querier FakeTxQuerier
		querier.orders = []store.Order{
			{
				ID:               "1",
				BillingAccountID: "1",
				ProjectID:        "1",
				PriceHr:          100,
			},
		}
		querier.leasesForTimeRange = []store.Lease{
			{
				ID:         "1",
				OrderID:    "1",
				CreateTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				EndTime: sql.NullTime{
					Time:  time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
					Valid: true,
				},
				PriceHr: 100,
			},
			{
				ID:         "2",
				OrderID:    "1",
				CreateTime: time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
				EndTime: sql.NullTime{
					Time:  time.Date(2020, time.January, 4, 0, 0, 0, 0, time.UTC),
					Valid: true,
				},
				PriceHr: 100,
			},
		}
		biller := NewBiller(&querier, zaptest.NewLogger(t))

		ctx := context.Background()
		billingAccounts := []store.BillingAccount{
			{
				ID:            "1",
				CreateTime:    time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				SupplyEnabled: true,
				DemandEnabled: true,
			},
		}
		err := biller.calculateDemandSpend(ctx, billingAccounts, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, time.January, 30, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if querier.orderSpend.String() != "7200" {
			t.Errorf("expected order spend to be %s, got %s", "7200", querier.orderSpend.String())
		}
		if querier.projectSpend.String() != "7200" {
			t.Errorf("expected project spend to be %s, got %s", "7200", querier.projectSpend.String())
		}
		if querier.billingAccountSpend.String() != "7200" {
			t.Errorf("expected billing account spend to be %s, got %s", "7200", querier.billingAccountSpend.String())
		}
	})
	t.Run("should calculate spend when the end time is not defined", func(t *testing.T) {
		var querier FakeTxQuerier
		querier.orders = []store.Order{
			{
				ID:               "1",
				BillingAccountID: "1",
				ProjectID:        "1",
				PriceHr:          100,
			},
		}
		querier.leasesForTimeRange = []store.Lease{
			{
				ID:         "1",
				OrderID:    "1",
				CreateTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				EndTime: sql.NullTime{
					Valid: false,
				},
				PriceHr: 100,
			},
			{
				ID:         "2",
				OrderID:    "1",
				CreateTime: time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
				EndTime: sql.NullTime{
					Valid: true,
					Time:  time.Date(2020, time.January, 4, 0, 0, 0, 0, time.UTC),
				},
				PriceHr: 100,
			},
			{
				ID:         "3",
				OrderID:    "2",
				CreateTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				EndTime: sql.NullTime{
					Valid: false,
				},
				PriceHr: 150,
			},
			{
				ID:         "4",
				OrderID:    "2",
				CreateTime: time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
				EndTime: sql.NullTime{
					Valid: false,
				},
				PriceHr: 10,
			},
		}
		biller := NewBiller(&querier, zaptest.NewLogger(t))
		billingAccounts := []store.BillingAccount{
			{
				ID:            "1",
				CreateTime:    time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				SupplyEnabled: true,
				DemandEnabled: true,
			},
		}
		err := biller.calculateDemandSpend(context.Background(), billingAccounts, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, time.January, 30, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if querier.orderSpend.String() != "185519.9999999999480" {
			t.Errorf("expected order spend to be %s, got %s", "185519.9999999999480", querier.orderSpend.String())
		}
		if querier.projectSpend.String() != "185519.9999999999480" {
			t.Errorf("expected project spend to be %s, got %s", "185519.9999999999480", querier.projectSpend.String())
		}
		if querier.billingAccountSpend.String() != "185519.9999999999480" {
			t.Errorf("expected billing account spend to be %s, got %s", "185519.9999999999480", querier.billingAccountSpend.String())
		}
	})
	t.Run("should calculate spend when the end time is after the end date", func(t *testing.T) {
		var querier FakeTxQuerier
		querier.orders = []store.Order{
			{
				ID:               "1",
				BillingAccountID: "1",
				ProjectID:        "1",
				PriceHr:          100,
			},
		}
		querier.leasesForTimeRange = []store.Lease{
			{
				ID:         "1",
				OrderID:    "1",
				CreateTime: time.Date(2020, time.January, 29, 0, 0, 0, 0, time.UTC),
				EndTime: sql.NullTime{
					Time:  time.Date(2020, time.February, 2, 0, 0, 0, 0, time.UTC),
					Valid: true,
				},
				PriceHr: 100,
			},
			{
				ID:         "2",
				OrderID:    "1",
				CreateTime: time.Date(2020, time.January, 28, 0, 0, 0, 0, time.UTC),
				EndTime: sql.NullTime{
					Time:  time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC),
					Valid: true,
				},
				PriceHr: 150,
			},
		}
		biller := NewBiller(&querier, zaptest.NewLogger(t))
		billingAccounts := []store.BillingAccount{
			{
				ID:            "1",
				CreateTime:    time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				SupplyEnabled: true,
				DemandEnabled: true,
			},
		}
		err := biller.calculateDemandSpend(context.Background(), billingAccounts, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, time.January, 30, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if querier.orderSpend.String() != "9599.999999999930300" {
			t.Errorf("expected order spend to be %s, got %s", "9599.999999999930300", querier.orderSpend.String())
		}
		if querier.projectSpend.String() != "9599.999999999930300" {
			t.Errorf("expected project spend to be %s, got %s", "9599.999999999930300", querier.projectSpend.String())
		}
		if querier.billingAccountSpend.String() != "9599.999999999930300" {
			t.Errorf("expected billing account spend to be %s, got %s", "9599.999999999930300", querier.billingAccountSpend.String())
		}
	})
}

func Test_LoopAddMultipleNullDecimals(t *testing.T) {
	// should not use as calulation should include exponent value
	t.Run("should add multiple null decimals using a big int", func(t *testing.T) {
		oneNullDecimal := apd.NullDecimal{
			Decimal: *apd.New(1123, -4),
			Valid:   false,
		}
		twoNullDecimal := apd.NullDecimal{
			Decimal: *apd.New(2345, -3),
			Valid:   false,
		}
		// add null decimals together
		bigInt := big.Int{}
		resInt := bigInt.Add(&twoNullDecimal.Decimal.Coeff, &oneNullDecimal.Decimal.Coeff)
		if resInt.String() != "3468" {
			t.Errorf("expected 3468, got %v", resInt)
		}
	})
	t.Run("should add multiple decimals when calculating lease spend", func(t *testing.T) {
		priceHrDecimal, err := conv.FromFloat(1.123)
		if err != nil {
			t.Errorf("error converting price hour to decimal: %v", err)
		}
		// convert leaseHours to decimal from float64
		leaseHoursDecimal, err := conv.FromFloat(2.123)
		if err != nil {
			t.Errorf("error converting lease hours to decimal: %v", err)
		}

		// define apd context
		apdContext := apd.Context{
			MaxExponent: 65,
			MinExponent: -18,
			Precision:   65,
		}
		leaseSpendDecimal := apd.NullDecimal{}
		cond, err := apdContext.Mul(&leaseSpendDecimal.Decimal, &leaseHoursDecimal, &priceHrDecimal)
		if err != nil {
			t.Errorf("error calculating lease spend: %v", err)
		}
		if cond.Any() {
			t.Errorf("error calculating lease spend: %v", err)
		} else {
			leaseSpendDecimal.Valid = true
		}
		firstSpend, cond, err := apd.NewFromString("0")
		if err != nil {
			t.Errorf("error creating first spend: %v", err)
		}
		if cond.Any() {
			t.Errorf("error creating first spend: %v", err)
		}
		secondSpend, cond, err := apd.NewFromString("0")
		if err != nil {
			t.Errorf("error creating second spend: %v", err)
		}
		if cond.Any() {
			t.Errorf("error creating second spend: %v", err)
		}
		thirdSpend, cond, err := apd.NewFromString("0")
		if err != nil {
			t.Errorf("error creating third spend: %v", err)
		}
		if cond.Any() {
			t.Errorf("error creating third spend: %v", err)
		}

		for i := 0; i < 3; i++ {
			if leaseSpendDecimal.Valid {
				cond, err := apdContext.Add(
					firstSpend,
					firstSpend,
					&leaseSpendDecimal.Decimal,
				)
				if err != nil {
					t.Errorf("error adding first spend: %v", err)
				}
				if cond.Any() {
					t.Errorf("error adding first spend: %v", err)
				}
				apdContext.Add(
					secondSpend,
					secondSpend,
					&leaseSpendDecimal.Decimal,
				)
				apdContext.Add(
					thirdSpend,
					thirdSpend,
					&leaseSpendDecimal.Decimal,
				)
			}
		}

		if firstSpend.String() != "7.152387" {
			t.Errorf("expected %f, got %v", 7.152387, firstSpend.String())
		}
		if secondSpend.String() != "7.152387" {
			t.Errorf("expected %f, got %v", 7.152387, secondSpend.String())
		}
		if thirdSpend.String() != "7.152387" {
			t.Errorf("expected %f, got %v", 7.152387, thirdSpend.String())
		}
	})
}
