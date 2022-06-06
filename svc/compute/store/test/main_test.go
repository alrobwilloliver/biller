package store_test

import (
	"database/sql"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	migrate_postgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
)

var (
	Host           = "localhost"
	PgPort         = 5432
	User           = "postgres"
	Dbname         = "compute_queries_test"
	DbPass         = "test"
	MigrationsPath = ""
	enabled        = true
)

func IsEnabled(t *testing.T) {
	t.Helper()

	if !enabled {
		t.SkipNow()
	}
}

func getRootPath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			path, err := filepath.Abs(dir)
			if err != nil {
				return "", err
			}
			return path, nil
		}

		dir = filepath.Join(dir, "..")
	}
}

// migration tool postgres implementation
func setUpDatabase(testName string) error {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		Host, PgPort, User, DbPass, Dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}

	row, err := db.Query(fmt.Sprintf("DROP DATABASE IF EXISTS %s_%s;", Dbname, testName))
	if err != nil {
		// fmt.Print("err", err)
		return err
	}
	row.Close()

	row, err = db.Query(fmt.Sprintf("CREATE DATABASE %s_%s;", Dbname, testName))
	if err != nil {
		// fmt.Print("err", err)
		return err
	}
	row.Close()
	err = db.Close()
	if err != nil {
		return err
	}
	psqlNewInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		Host, PgPort, User, DbPass, Dbname+"_"+testName)
	db, err = sql.Open("postgres", psqlNewInfo)
	if err != nil {
		fmt.Printf("err- problem %v", err)

		return err
	}

	driver, err := migrate_postgres.WithInstance(db, &migrate_postgres.Config{
		DatabaseName:    Dbname + "_" + testName,
		MigrationsTable: "migrations-table",
	})
	if err != nil {
		fmt.Print("migration error \n")
		return err
	}
	defer driver.Close()

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%v", MigrationsPath),
		"postgresDbInstance",
		driver,
	)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil {
		return err
	}
	return nil
}

func TestMain(m *testing.M) {
	// we must call `flag.Parse(), otherwise `testing.Short()` will always be false`
	flag.Parse()

	// this allows us to skip tests if we are in short mode, or don't have docker
	if testing.Short() {
		fmt.Println("skipping as we are in short mode")
		enabled = false
		os.Exit(m.Run())
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		fmt.Printf("error creating docker pool: %s\n", err)
	}
	rootPath, err := getRootPath()
	if err != nil {
		fmt.Printf("could not get root path: %s\n", err)
	}

	MigrationsPath = fmt.Sprintf("%s/svc/%s/store/migrations/", rootPath, "compute")
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14",
		Env: []string{
			"POSTGRES_DB=compute_queries_test",
			"POSTGRES_PASSWORD=test",
			"POSTGRES_USER=postgres",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}

		// as the database is only temporary, we can store it in ram, which dramatically speeds up
		// run time
		config.Tmpfs = map[string]string{"/var/lib/postgresql/data": "rw"}
	})

	if err != nil {
		fmt.Printf("could not create resource (docker probably either not installed/running): %s\n", err)
		fmt.Println("skipping as docker could not be connected to")
		enabled = false
		os.Exit(m.Run())
	}

	err = resource.Expire(120)
	if err != nil {
		fmt.Printf("could not set expiry on resource: %s\n", err)
		os.Exit(m.Run())
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	dbhost, portString, _ := net.SplitHostPort(hostAndPort)
	port, err := strconv.Atoi(portString)
	if err != nil {
		fmt.Printf("could not convert port into int: %s\n", err)
	}
	Host = dbhost
	PgPort = port

	if err := pool.Retry(func() error {
		db, err := sql.Open("postgres", fmt.Sprintf("postgres://postgres:test@%s/compute_queries_test?sslmode=disable", hostAndPort))
		if err != nil {
			return err
		}

		if err := db.Ping(); err != nil {
			return err
		}
		return nil
	}); err != nil {
		fmt.Printf("could not open postgres connection: %s\n", err)
	}

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		fmt.Printf("could not purge resource: %s\n", err)
	}

	os.Exit(code)
}
