sqlc:
	cd svc/compute/store; sqlc generate; cd ../../..


migrate := migrate -database "postgres://compute@localhost/compute?sslmode=disable" -path ./svc/compute/store/migrations
init-db:
	$(migrate) force 1
	$(migrate) down --all
	$(migrate) up

migrate:
	$(migrate) up

test:
	go test ./...

tools:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.15.1
	go install github.com/bufbuild/buf/cmd/buf@v1.3.1
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.10.0
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.10.0
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.0
	go install github.com/kyleconroy/sqlc/cmd/sqlc@v1.12.0
