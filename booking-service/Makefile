COMPOSE_FILE := dev/deployments/docker-compose.yml
MIGRATE_DSN := postgres://booking:booking@localhost:5433/booking?sslmode=disable

.PHONY: up down migrate tests

up:
	docker compose -f $(COMPOSE_FILE) up -d --build

tests:
	go test ./... -cover

migrate:
	POSTGRES_DSN=$(MIGRATE_DSN) go run ./cmd/migrator

down:
	docker compose -f $(COMPOSE_FILE) down
