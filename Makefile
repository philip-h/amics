build:
	go build -o tmp/main ./cmd/main.go

run: build
	./tmp/main

.PHONY: migrateup
migrateup:
	migrate -path internal/db/migrations/ -database "sqlite3://amics.db" up

.PHONY: migratedown
migratedown:
	migrate -path internal/db/migrations/ -database "sqlite3://amics.db" down