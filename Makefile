build:
	go build -o tmp/main ./cmd/main.go

run: build
	./tmp/main

.PHONY: test
test:
	go test -v ./internal/*

.PHONY: migrateup
migrateup:
	migrate -path internal/db/migrations/ -database "postgres://postgres:@localhost:5432/amics?sslmode=disable" up

.PHONY: migratedown
migratedown:
	migrate -path internal/db/migrations/ -database "postgres://postgres:@localhost:5432/amics?sslmode=disable" down
