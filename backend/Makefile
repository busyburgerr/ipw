.PHONY: help up down logs run build test lint tidy migrate-create

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

up: ## Start local infra (postgres, redis, minio)
	docker compose up -d

down: ## Stop local infra
	docker compose down

logs: ## Tail infra logs
	docker compose logs -f

run: ## Run the API (applies migrations on boot)
	go run ./cmd/api

build: ## Build the API binary
	go build -o bin/api ./cmd/api

test: ## Run tests
	go test ./... -race -count=1

lint: ## Run golangci-lint
	golangci-lint run

tidy: ## Tidy go.mod
	go mod tidy

migrate-create: ## Create a migration pair: make migrate-create name=add_projects
	@test -n "$(name)" || (echo "usage: make migrate-create name=<snake_case>"; exit 1)
	@dir=internal/platform/postgres/migrations; \
	next=$$(printf "%04d" $$(( $$(ls $$dir | grep -oE '^[0-9]+' | sort -n | tail -1 | sed 's/^0*//' ) + 1 ))); \
	touch $$dir/$${next}_$(name).up.sql $$dir/$${next}_$(name).down.sql; \
	echo "created $$dir/$${next}_$(name).{up,down}.sql"
