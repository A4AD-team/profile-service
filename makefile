.PHONY: run build migrate-up migrate-down migrate-down-1 test lint deps docker-up docker-down

include .env
export

run:
	go run ./cmd/server

build:
	go build -o bin/profile-service ./cmd/server

migrate-up:
	migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DATABASE_URL)" down

migrate-down-1:
	migrate -path ./migrations -database "$(DATABASE_URL)" down 1

test:
	go test -race -cover ./...

lint:
	golangci-lint run ./...

deps:
	go mod download && go mod tidy

docker-up:
	docker compose up -d

docker-down:
	docker compose down
