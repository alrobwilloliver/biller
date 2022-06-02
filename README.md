## Description

This is a simple billing system for billing accounts, projects and leases based on time. It is an gRPC API with a REST gateway using Postgres and SQLC. I stripped out much of this code from a Compute Market beta in a simplified form to demonstrate how I built and maintained our testing suite, as well as the apd.Decimal library that I added to the biller, replacing floats for more precise calculations when multiplying lease price per hour by the length of the lease and incrementally adding to the total spend.

## quickstart:

```shell
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.15.1
migrate -database "postgres://compute@localhost/compute?sslmode=disable" -path ./svc/compute/store/migrations up
go run ./svc/compute
```

## sqlc set up
make
