## quickstart:

go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.15.1
migrate -database "postgres://compute@localhost/compute?sslmode=disable" -path ./svc/compute/store/migrations up
go run ./svc/compute

## sqlc set up
make
