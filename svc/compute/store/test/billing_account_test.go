package store_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"biller/lib/postgresql"
	"biller/svc/compute/store"

	"github.com/prometheus/client_golang/prometheus"
)

// test billing_account queries
func TestFindBillingAccountById(t *testing.T) {
	IsEnabled(t)
	dbTest := "findbillingaccountbyid"
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

	_, err = conn.Exec(newCtx, `
		INSERT INTO billing_account(id, create_time, demand_enabled, supply_enabled)
		VALUES
			('fakeid', '2022-01-10', true, true)
	`)

	if err != nil {
		t.Fatal(err)
	}

	id := "fakeid"

	billingAccount, err := postgresqlQueries.FindBillingAccountById(newCtx, id)
	if err != nil {
		t.Errorf("FindBillingAccountById() error: %v", err)
	}

	if billingAccount.ID != id {
		t.Errorf("billingAccount id want: %s, got: %s", id, billingAccount.ID)
	}

	if !strings.Contains(billingAccount.CreateTime.String(), "2022-01-10 00:00:00 +0000") {
		t.Errorf("billingAccount create_time want to contain: %s, got: %s", "2022-01-10 00:00:00 +0000", billingAccount.CreateTime.String())
	}

	if !billingAccount.DemandEnabled {
		t.Errorf("billing account should have demand_enabled true, got: %t", billingAccount.DemandEnabled)
	}

	if !billingAccount.SupplyEnabled {
		t.Errorf("billing account should have supply_enabled true, got: %t", billingAccount.SupplyEnabled)
	}
}

func TestCreateBillingAccount(t *testing.T) {
	IsEnabled(t)
	dbTest := "createbillingaccountbyid"
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

	billingaccount, err := postgresqlQueries.CreateBillingAccount(newCtx, "fakeid")

	if err != nil {
		t.Errorf("CreateBillingAccount() error: %v", err)
	}

	id := billingaccount.ID
	sqlQuery := `
		SELECT * FROM billing_account WHERE id = $1
	`

	res, err := conn.Query(ctx, sqlQuery, id)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Close()
	var billingAcc store.BillingAccount
	if res.Next() {
		err = res.Scan(
			&billingAcc.ID,
			&billingAcc.CreateTime,
			&billingAcc.SupplyEnabled,
			&billingAcc.DemandEnabled,
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	if billingAcc.ID != billingaccount.ID {
		t.Errorf("Expected created billing account id to equal %s, but got %s", billingaccount.ID, billingAcc.ID)
	}
}

func TestListBillingAccounts(t *testing.T) {
	IsEnabled(t)
	dbTest := "listbillingaccounts"
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
	sqlQuery := `
		INSERT INTO billing_account(id, create_time, supply_enabled, demand_enabled)
			VALUES
				('fakeid', '2022-01-19', false, false),
				('fakeid1', '2022-01-20', false, true),
				('fakeid2', '2022-01-21', true, false),
				('fakeid3', '2022-01-22', false, true),
				('fakeid4', '2022-01-23', false, false)

	`

	_, err = conn.Exec(newCtx, sqlQuery)
	if err != nil {
		t.Fatal(err)
	}

	res, err := postgresqlQueries.ListBillingAccounts(newCtx, store.ListBillingAccountsParams{
		Limit: 5,
		Ids:   []string{"fakeid", "fakeid1", "fakeid2", "fakeid3", "fakeid4"},
	})
	if err != nil {
		t.Errorf("ListBillingAccounts() error: %v", err)
	}
	if len(res) != 5 {
		t.Errorf("ListBillingAccounts() want 5, got: %d", len(res))
	}

	res, err = postgresqlQueries.ListBillingAccounts(newCtx, store.ListBillingAccountsParams{
		Limit: 2,
		Ids:   []string{"fakeid2", "fakeid3", "fakeid4"},
	})
	if err != nil {
		t.Errorf("ListBillingAccounts() error: %v", err)
	}
	if len(res) != 2 {
		t.Errorf("ListBillingAccounts() want 2, got: %d", len(res))
	}
	if res[0].ID != "fakeid2" {
		t.Errorf("ListBillingAccounts() want 'fakeid2', got: %s", res[0].ID)
	}
	if res[0].DemandEnabled != false {
		t.Errorf("ListBillingAccounts() want true, got: %t", res[0].DemandEnabled)
	}
	if res[0].SupplyEnabled != true {
		t.Errorf("ListBillingAccounts() want false, got: %t", res[0].SupplyEnabled)
	}
	if res[0].CreateTime.String() != time.Date(2022, time.January, 21, 0, 0, 0, 0, time.UTC).String() {
		t.Errorf("ListBillingAccounts() want %v, got: %s", time.Date(2022, time.January, 21, 0, 0, 0, 0, time.UTC), res[0].CreateTime)
	}
	if res[1].ID != "fakeid3" {
		t.Errorf("ListBillingAccounts() want 'fakeid3', got: %s", res[1].ID)
	}
	if res[1].DemandEnabled != true {
		t.Errorf("ListBillingAccounts() want true, got: %t", res[1].DemandEnabled)
	}
	if res[1].SupplyEnabled != false {
		t.Errorf("ListBillingAccounts() want false, got: %t", res[1].SupplyEnabled)
	}
	if res[1].CreateTime.String() != time.Date(2022, time.January, 22, 0, 0, 0, 0, time.UTC).String() {
		t.Errorf("ListBillingAccounts() want %v, got: %s", time.Date(2022, time.January, 22, 0, 0, 0, 0, time.UTC), res[0].CreateTime)
	}
}

func TestListAllBillingAccounts(t *testing.T) {
	IsEnabled(t)
	dbTest := "listallbillingaccounts"
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
	sqlQuery := `
		INSERT INTO billing_account(id, create_time, supply_enabled, demand_enabled)
			VALUES
				('fakeid', '2022-01-19', false, false),
				('fakeid1', '2022-01-20', false, true),
				('fakeid2', '2022-01-21', true, false),
				('fakeid3', '2022-01-22', false, true),
				('fakeid4', '2022-01-23', false, false)

	`

	_, err = conn.Exec(newCtx, sqlQuery)
	if err != nil {
		t.Fatal(err)
	}

	res, err := postgresqlQueries.ListAllBillingAccounts(newCtx)
	if err != nil {
		t.Errorf("ListBillingAccounts() error: %v", err)
	}
	if len(res) != 5 {
		t.Errorf("ListBillingAccounts() want 5, got: %d", len(res))
	}
}

func TestEnableBillingAccountDemand(t *testing.T) {
	IsEnabled(t)
	dbTest := "enablebillingaccountdemand"
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
	sqlQuery := `
	INSERT INTO billing_account(id, create_time, supply_enabled, demand_enabled)
	VALUES('fakeid', '2022-01-20', false, false)
	`

	_, err = conn.Exec(newCtx, sqlQuery)
	if err != nil {
		t.Fatal(err)
	}

	id := "fakeid"
	_, err = postgresqlQueries.SelectBillingAccountForUpdate(newCtx, id)
	if err != nil {
		t.Errorf("Error calling SelectBillingAccountForUpdate() = %v", err)
	}
	account, err := postgresqlQueries.EnableBillingAccountDemand(newCtx, id)
	if err != nil {
		t.Errorf("Error calling EnableBillingAccountDemand() = %v", err)
	}

	if !account.DemandEnabled {
		t.Errorf("Expected demand_enabled on billing account to equal %t, got %t", true, account.DemandEnabled)
	}
	if account.SupplyEnabled {
		t.Errorf("Expected supply_enabled on billing account to equal %t, got %t", false, account.SupplyEnabled)
	}
}

func TestEnableBillingAccountSupply(t *testing.T) {
	IsEnabled(t)
	dbTest := "enablebillingaccountsupply"
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
	sqlQuery := `
	INSERT INTO billing_account(id, create_time, demand_enabled, supply_enabled)
	VALUES('d88f5dec-1929-406e-ab90-49eed3eb79bb', '2022-01-20', false, false)
	`

	_, err = conn.Exec(newCtx, sqlQuery)
	if err != nil {
		t.Fatal(err)
	}

	id := "d88f5dec-1929-406e-ab90-49eed3eb79bb"
	_, err = postgresqlQueries.SelectBillingAccountForUpdate(newCtx, id)
	if err != nil {
		t.Errorf("Error calling SelectBillingAccountForUpdate() = %v", err)
	}
	account, err := postgresqlQueries.EnableBillingAccountDemand(newCtx, id)
	if err != nil {
		t.Errorf("Error calling EnableBillingAccountDemand() = %v", err)
	}

	if !account.DemandEnabled {
		t.Errorf("Expected demand_enabled on billing account to equal %t, got %t", true, account.DemandEnabled)
	}
}
