package store_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"biller/lib/postgresql"
	"biller/svc/compute/store"

	"github.com/cockroachdb/apd/v2"
	"github.com/prometheus/client_golang/prometheus"
)

func Test_CreateBillingAccountSpend(t *testing.T) {
	IsEnabled(t)
	dbTest := "createbillingaccountspend"
	err := setUpDatabase(dbTest)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		t.Fatal(err)
	}

	// create connection
	clientConfig := &postgresql.ClientConfig{
		User:     User,
		Pass:     DbPass,
		Host:     Host,
		Port:     PgPort,
		Database: Dbname + "_" + dbTest,
	}
	registry := prometheus.NewRegistry()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second))
	defer cancel()

	postgresqlDb, err := postgresql.NewClient(ctx, clientConfig, registry)
	if err != nil {
		t.Fatal(err)
	}

	newCtx := context.Background()
	_, err = postgresqlDb.Exec(newCtx, `
		INSERT INTO billing_account(id, create_time, supply_enabled, demand_enabled)
			VALUES('billing-account-id', '2022-01-20', true, false)
	`)
	if err != nil {
		t.Fatal(err)
	}

	// generate queries struct to be able to make sqlc call
	postgresqlQueries := store.NewTxQueries(postgresqlDb)
	// defer postgresqlDb.Close() - times out - should use conn.Release()
	conn, err := postgresqlDb.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	spend := *apd.New(123, 0)
	res, err := postgresqlQueries.CreateBillingAccountSpend(newCtx, store.CreateBillingAccountSpendParams{
		BillingAccountID: "billing-account-id",
		Spend:            spend,
		StartTime:        time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
		EndTime:          time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.BillingAccountID != "billing-account-id" {
		t.Errorf("expected billing account id to be %s, got %s", "billing-account-id", res.BillingAccountID)
	}
	if res.Spend.String() != "123.000000000000000000" {
		t.Errorf("expected spend to be %s, got %s", "123.000000000000000000", res.Spend.String())
	}
	if res.StartTime.String() != time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC).String() {
		t.Errorf("expected start time to be %s, got %s", time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC), res.StartTime)
	}
	if res.EndTime.String() != time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC).String() {
		t.Errorf("expected end time to be %s, got %s", time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC), res.EndTime)
	}
}

func Test_CreateOrderSpend(t *testing.T) {
	IsEnabled(t)
	dbTest := "createorderspend"
	err := setUpDatabase(dbTest)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		t.Fatal(err)
	}

	// create connection
	clientConfig := &postgresql.ClientConfig{
		User:     User,
		Pass:     DbPass,
		Host:     Host,
		Port:     PgPort,
		Database: Dbname + "_" + dbTest,
	}
	registry := prometheus.NewRegistry()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second))
	defer cancel()

	postgresqlDb, err := postgresql.NewClient(ctx, clientConfig, registry)
	if err != nil {
		t.Fatal(err)
	}

	newCtx := context.Background()
	billingAccountId := "billing-account-id"
	_, err = postgresqlDb.Exec(newCtx, fmt.Sprintf(`
		INSERT INTO billing_account(id, create_time, supply_enabled, demand_enabled)
		VALUES
			('%s', '2022-01-10', true, true)
	`, billingAccountId))
	projectId := "project-id"
	_, err = postgresqlDb.Exec(newCtx, fmt.Sprintf(`
		INSERT INTO project(id, create_time, billing_account_id)
			VALUES('%s', '2022-01-20', '%s')
	`, projectId, billingAccountId))

	orderId := "order-id"
	_, err = postgresqlDb.Exec(newCtx, fmt.Sprintf(`
		INSERT INTO "order" (id, infra_type, project_id, description, quantity, create_time, price_hr, billing_account_id)
			VALUES ('%s', 'dedicated', '%s', 'description', 1, '2022-01-20', 100, '%s')
	`, orderId, projectId, billingAccountId))
	if err != nil {
		t.Fatal(err)
	}

	// generate queries struct to be able to make sqlc call
	postgresqlQueries := store.NewTxQueries(postgresqlDb)
	// defer postgresqlDb.Close() - times out - should use conn.Release()
	conn, err := postgresqlDb.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	spend := *apd.New(1, 0)
	res, err := postgresqlQueries.CreateOrderSpend(newCtx, store.CreateOrderSpendParams{
		OrderID:   orderId,
		Spend:     spend,
		StartTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.OrderID != orderId {
		t.Errorf("expected order id to be %s, got %s", orderId, res.OrderID)
	}
	if res.Spend.String() != "1.000000000000000000" {
		t.Errorf("expected spend to be %s, got %s", "1.000000000000000000", res.Spend.String())
	}
	if res.StartTime.String() != time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC).String() {
		t.Errorf("expected start time to be %s, got %s", time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC), res.StartTime)
	}
	if res.EndTime.String() != time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC).String() {
		t.Errorf("expected end time to be %s, got %s", time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC), res.EndTime)
	}
}

func Test_CreateProjectSpend(t *testing.T) {
	IsEnabled(t)
	dbTest := "createprojectspend"
	err := setUpDatabase(dbTest)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		t.Fatal(err)
	}

	// create connection
	clientConfig := &postgresql.ClientConfig{
		User:     User,
		Pass:     DbPass,
		Host:     Host,
		Port:     PgPort,
		Database: Dbname + "_" + dbTest,
	}
	registry := prometheus.NewRegistry()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second))
	defer cancel()

	postgresqlDb, err := postgresql.NewClient(ctx, clientConfig, registry)
	if err != nil {
		t.Fatal(err)
	}

	newCtx := context.Background()
	_, err = postgresqlDb.Exec(newCtx, `
		INSERT INTO billing_account(id, create_time, supply_enabled, demand_enabled)
		VALUES
			('billing-account-id', '2022-01-10', true, true)
	`)

	_, err = postgresqlDb.Exec(newCtx, `
		INSERT INTO project(id, create_time, billing_account_id)
			VALUES('project-id', '2022-01-20', 'billing-account-id')
	`)

	// generate queries struct to be able to make sqlc call
	postgresqlQueries := store.NewTxQueries(postgresqlDb)
	// defer postgresqlDb.Close() - times out - should use conn.Release()
	conn, err := postgresqlDb.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer conn.Release()

	spend := *apd.New(123445365764743, -15)
	res, err := postgresqlQueries.CreateProjectSpend(newCtx, store.CreateProjectSpendParams{
		ProjectID: "project-id",
		Spend:     spend,
		StartTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ProjectID != "project-id" {
		t.Errorf("expected project id to be %s, got %s", "project-id", res.ProjectID)
	}
	if res.Spend.String() != "0.123445365764743000" {
		t.Errorf("expected spend to be %s, got %s", "0.123445365764743000", res.Spend.String())
	}
	if res.StartTime.String() != time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC).String() {
		t.Errorf("expected start time to be %s, got %s", time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC), res.StartTime)
	}
	if res.EndTime.String() != time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC).String() {
		t.Errorf("expected end time to be %s, got %s", time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC), res.EndTime)
	}
}

func Test_FindBillingAccountSpendForTimeRange(t *testing.T) {
	IsEnabled(t)
	dbTest := "findbillingaccountspendfortimerange"
	err := setUpDatabase(dbTest)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		t.Fatal(err)
	}

	// create connection
	clientConfig := &postgresql.ClientConfig{
		User:     User,
		Pass:     DbPass,
		Host:     Host,
		Port:     PgPort,
		Database: Dbname + "_" + dbTest,
	}
	registry := prometheus.NewRegistry()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second))
	defer cancel()

	postgresqlDb, err := postgresql.NewClient(ctx, clientConfig, registry)
	if err != nil {
		t.Fatal(err)
	}

	// generate queries struct to be able to make sqlc call
	postgresqlQueries := store.NewTxQueries(postgresqlDb)
	// defer postgresqlDb.Close() - times out - should use conn.Release()
	conn, err := postgresqlDb.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	newCtx := context.Background()
	_, err = postgresqlDb.Exec(newCtx, `
		INSERT INTO billing_account(id, create_time, supply_enabled, demand_enabled)
			VALUES('billing-account-id', '2022-01-20', true, false)
	`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = postgresqlDb.Exec(newCtx, `
		INSERT INTO billing_account_spend(billing_account_id, spend, start_time, end_time)
			VALUES('billing-account-id', 1.0, '2020-01-01', '2020-01-02')
	`)
	if err != nil {
		t.Fatal(err)
	}

	res, err := postgresqlQueries.FindBillingAccountSpendForTimeRange(newCtx, store.FindBillingAccountSpendForTimeRangeParams{
		BillingAccountID: "billing-account-id",
		StartTime:        time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
		EndTime:          time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.BillingAccountID != "billing-account-id" {
		t.Errorf("expected billing account id to be %s, got %s", "billing-account-id", res.BillingAccountID)
	}
	if res.Spend.String() != "1.000000000000000000" {
		t.Errorf("expected spend to be %s, got %s", "1.000000000000000000", res.Spend.String())
	}
	if res.StartTime.String() != time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC).String() {
		t.Errorf("expected start time to be %s, got %s", time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC), res.StartTime)
	}
	if res.EndTime.String() != time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC).String() {
		t.Errorf("expected end time to be %s, got %s", time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC), res.EndTime)
	}
}

func Test_FindOrderSpendForTimeRange(t *testing.T) {
	IsEnabled(t)
	dbTest := "findorderspendfortimerange"
	err := setUpDatabase(dbTest)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		t.Fatal(err)
	}

	// create connection
	clientConfig := &postgresql.ClientConfig{
		User:     User,
		Pass:     DbPass,
		Host:     Host,
		Port:     PgPort,
		Database: Dbname + "_" + dbTest,
	}
	registry := prometheus.NewRegistry()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second))
	defer cancel()

	postgresqlDb, err := postgresql.NewClient(ctx, clientConfig, registry)
	if err != nil {
		t.Fatal(err)
	}

	// generate queries struct to be able to make sqlc call
	postgresqlQueries := store.NewTxQueries(postgresqlDb)
	// defer postgresqlDb.Close() - times out - should use conn.Release()
	conn, err := postgresqlDb.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	newCtx := context.Background()

	billingAccountId := "billing-account-id"
	_, err = postgresqlDb.Exec(newCtx, fmt.Sprintf(`
		INSERT INTO billing_account(id, create_time, supply_enabled, demand_enabled)
		VALUES
			('%s', '2022-01-10', true, true)
	`, billingAccountId))
	projectId := "project-id"
	_, err = postgresqlDb.Exec(newCtx, fmt.Sprintf(`
		INSERT INTO project(id, create_time, billing_account_id)
			VALUES('%s', '2022-01-20', '%s')
	`, projectId, billingAccountId))

	orderId := "order-id"
	_, err = postgresqlDb.Exec(newCtx, fmt.Sprintf(`
		INSERT INTO "order" (id, infra_type, project_id, description, quantity, create_time, price_hr, billing_account_id)
			VALUES ('%s', 'dedicated', '%s', 'description', 1, '2022-01-20', 100, '%s')
	`, orderId, projectId, billingAccountId))
	if err != nil {
		t.Fatal(err)
	}

	_, err = postgresqlDb.Exec(newCtx, `
		INSERT INTO order_spend(order_id, spend, start_time, end_time)
			VALUES('order-id', 1.0, '2020-01-01', '2020-01-02')
	`)
	if err != nil {
		t.Fatal(err)
	}

	res, err := postgresqlQueries.FindOrderSpendForTimeRange(newCtx, store.FindOrderSpendForTimeRangeParams{
		OrderID:   orderId,
		StartTime: time.Date(2019, time.December, 30, 0, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2020, time.January, 3, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.OrderID != orderId {
		t.Errorf("expected order id to be %s, got %s", orderId, res.OrderID)
	}
	if res.Spend.String() != "1.000000000000000000" {
		t.Errorf("expected spend to be %s, got %s", "1.000000000000000000", res.Spend.String())
	}
	if res.StartTime.String() != time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC).String() {
		t.Errorf("expected start time to be %s, got %s", time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC), res.StartTime)
	}
	if res.EndTime.String() != time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC).String() {
		t.Errorf("expected end time to be %s, got %s", time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC), res.EndTime)
	}
}

func Test_FindProjectSpendForTimeRange(t *testing.T) {
	IsEnabled(t)
	dbTest := "findprojectspendfortimerange"
	err := setUpDatabase(dbTest)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		t.Fatal(err)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		t.Fatal(err)
	}

	// create connection
	clientConfig := &postgresql.ClientConfig{
		User:     User,
		Pass:     DbPass,
		Host:     Host,
		Port:     PgPort,
		Database: Dbname + "_" + dbTest,
	}
	registry := prometheus.NewRegistry()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second))
	defer cancel()

	postgresqlDb, err := postgresql.NewClient(ctx, clientConfig, registry)
	if err != nil {
		t.Fatal(err)
	}

	// generate queries struct to be able to make sqlc call
	postgresqlQueries := store.NewTxQueries(postgresqlDb)
	// defer postgresqlDb.Close() - times out - should use conn.Release()
	conn, err := postgresqlDb.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	newCtx := context.Background()

	billingAccountId := "billing-account-id"
	_, err = postgresqlDb.Exec(newCtx, fmt.Sprintf(`
		INSERT INTO billing_account(id, create_time, supply_enabled, demand_enabled)
		VALUES
			('%s', '2022-01-10', true, true)
	`, billingAccountId))
	projectId := "project-id"
	_, err = postgresqlDb.Exec(newCtx, fmt.Sprintf(`
		INSERT INTO project(id, create_time, billing_account_id)
			VALUES('%s', '2022-01-20', '%s')
	`, projectId, billingAccountId))

	if err != nil {
		t.Fatal(err)
	}

	_, err = postgresqlDb.Exec(newCtx, fmt.Sprintf(`
		INSERT INTO project_spend(project_id, spend, start_time, end_time)
			VALUES('%s', 1.0, '2020-01-01', '2020-01-02')
	`, projectId))

	if err != nil {
		t.Fatal(err)
	}

	res, err := postgresqlQueries.FindProjectSpendForTimeRange(newCtx, store.FindProjectSpendForTimeRangeParams{
		ProjectID: projectId,
		StartTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ProjectID != projectId {
		t.Errorf("expected project id to be %s, got %s", projectId, res.ProjectID)
	}
	if res.Spend.String() != "1.000000000000000000" {
		t.Errorf("expected spend to be %s, got %s", "1.000000000000000000", res.Spend.String())
	}
	if res.StartTime.String() != time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC).String() {
		t.Errorf("expected start time to be %s, got %s", time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC), res.StartTime)
	}
	if res.EndTime.String() != time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC).String() {
		t.Errorf("expected end time to be %s, got %s", time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC), res.EndTime)
	}
}
