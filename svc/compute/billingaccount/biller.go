package billingaccount

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"biller/lib/conv"
	"biller/svc/compute/store"

	"github.com/cockroachdb/apd/v2"
	"go.uber.org/zap"
)

// Simple, non-killbill billing helper to show demand spend & supplier earnings
// Calculates earnings per lease and sums them up at host, host group inventory and data center level

type Biller struct {
	querier store.TxQuerier
	log     *zap.Logger
}

func NewBiller(querier store.TxQuerier, log *zap.Logger) *Biller {
	return &Biller{
		querier: querier,
		log:     log,
	}
}

func (b *Biller) Run(ctx context.Context) error {
	billingAccounts, err := b.querier.ListAllBillingAccounts(ctx)

	if err != nil {
		b.log.Error("list billing accounts failed", zap.Error(err))
		return err
	}

	// TODO: Set start/end times based on ctx or run parameters
	// This is definitely something that could benefit from being run in temporal
	// endTime and startTime are currently the first & last nanoseconds of the current month
	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endTime := startTime.AddDate(0, 1, 0)

	err = b.calculateDemandSpend(ctx, billingAccounts, startTime, endTime)
	b.log.Error("error calculating demand spend: %v", zap.Error(err))
	return nil
}

type DemandSpend struct {
	BillingAccountID string
	Spend            *apd.Decimal
	Projects         map[string]*ProjectSpend
}

type ProjectSpend struct {
	ProjectID string
	Spend     *apd.Decimal
	Orders    map[string]*OrderSpend
}

type OrderSpend struct {
	OrderID string
	Spend   *apd.Decimal
}

// Calculate the spend of demand customers
// Get billing account -> orders -> leases
func (b *Biller) calculateDemandSpend(ctx context.Context, billingAccounts []store.BillingAccount, startTime time.Time, endTime time.Time) error {
	for _, billingAccount := range billingAccounts {
		spend := &DemandSpend{
			BillingAccountID: billingAccount.ID,
			Projects:         make(map[string]*ProjectSpend),
			Spend:            apd.New(0, 0),
		}

		orders, err := b.querier.ListOrdersByBillingAccountId(ctx, billingAccount.ID) // TODO paginate this at some point

		if err != nil {
			return fmt.Errorf("error when listing orders for billing account: %w", err)
		}

		for _, order := range orders {
			if _, ok := spend.Projects[order.ProjectID]; !ok {
				spend.Projects[order.ProjectID] = &ProjectSpend{
					ProjectID: order.ProjectID,
					Orders:    make(map[string]*OrderSpend),
					Spend:     apd.New(0, 0),
				}
			}
			spend.Projects[order.ProjectID].Orders[order.ID] = &OrderSpend{
				OrderID: order.ID,
				Spend:   apd.New(0, 0),
			}

			leases, err := b.querier.ListLeasesForTimeRangeByOrderId(ctx, store.ListLeasesForTimeRangeByOrderIdParams{
				OrderID: order.ID,
				StartTime: sql.NullTime{
					Time:  startTime,
					Valid: true,
				},
				EndTime: endTime,
			})

			if err != nil {
				return fmt.Errorf("list leases for time range by order id failed: %w", err)
			}

			for _, lease := range leases {
				// calculate number of hours the lease was active in the specified time range
				var (
					leaseStartTime time.Time
					leaseEndTime   time.Time
				)

				if lease.CreateTime.After(startTime) {
					leaseStartTime = lease.CreateTime
				} else {
					leaseStartTime = startTime
				}

				if lease.EndTime.Time.Before(endTime) && lease.EndTime.Valid {
					leaseEndTime = lease.EndTime.Time
				} else {
					leaseEndTime = endTime.Add(time.Nanosecond * -1)
				}

				leaseDuration := leaseEndTime.Sub(leaseStartTime)
				leaseHours := leaseDuration.Hours()
				// convert priceHr to decimal from float64
				priceHrDecimal, err := conv.FromFloat(lease.PriceHr)
				if err != nil {
					return fmt.Errorf("error converting price hour to decimal: %w", err)
				}
				// convert leaseHours to decimal from float64
				leaseHoursDecimal, err := conv.FromFloat(leaseHours)
				if err != nil {
					return fmt.Errorf("error converting lease hours to decimal: %w", err)
				}

				// define apd context
				apdContext := apd.Context{
					MaxExponent: 65,
					MinExponent: -18,
					Precision:   65,
				}
				leaseSpendDecimal := apd.NullDecimal{
					Decimal: *apd.New(0, 0),
					Valid:   false,
				}
				cond, err := apdContext.Mul(&leaseSpendDecimal.Decimal, &leaseHoursDecimal, &priceHrDecimal)
				if err != nil {
					return fmt.Errorf("error calculating lease spend: %w", err)
				}
				if cond.Any() {
					return fmt.Errorf("error calculating lease spend: %w", err)
				}

				cond, err = apdContext.Add(
					spend.Projects[order.ProjectID].Orders[order.ID].Spend,
					spend.Projects[order.ProjectID].Orders[order.ID].Spend,
					&leaseSpendDecimal.Decimal,
				)
				apdContext.Add(
					spend.Projects[order.ProjectID].Spend,
					spend.Projects[order.ProjectID].Spend,
					&leaseSpendDecimal.Decimal,
				)
				if err != nil {
					return fmt.Errorf("error adding lease spend to order spend: %w", err)
				}
				if cond.Any() {
					return fmt.Errorf("error adding lease spend to project spend: %w", err)
				}
				apdContext.Add(
					spend.Spend,
					spend.Spend,
					&leaseSpendDecimal.Decimal,
				)
			}
			// write order spend
			_, err = b.querier.CreateOrderSpend(ctx, store.CreateOrderSpendParams{
				OrderID:   order.ID,
				Spend:     *spend.Projects[order.ProjectID].Orders[order.ID].Spend,
				StartTime: startTime,
				EndTime:   endTime,
			})
			if err != nil {
				return fmt.Errorf("create order spend failed: %w", err)
			}
			// write project spend
			_, err = b.querier.CreateProjectSpend(ctx, store.CreateProjectSpendParams{
				ProjectID: order.ProjectID,
				Spend:     *spend.Projects[order.ProjectID].Spend,
				StartTime: startTime,
				EndTime:   endTime,
			})
			if err != nil {
				return fmt.Errorf("create project spend failed: %w", err)
			}
		}
		// write billing account spend
		_, err = b.querier.CreateBillingAccountSpend(ctx, store.CreateBillingAccountSpendParams{
			BillingAccountID: billingAccount.ID,
			Spend:            *spend.Spend,
			StartTime:        startTime,
			EndTime:          endTime,
		})
		if err != nil {
			return fmt.Errorf("create billing account spend failed: %w", err)
		}
	}
	return nil
}
