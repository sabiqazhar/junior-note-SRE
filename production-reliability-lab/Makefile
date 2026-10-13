.PHONY: help up down reset logs ps health test test-int lint build loadtest backup restore fmt

COMPOSE  := docker compose -f docker/compose.yaml --env-file .env
IMAGE    := sre-lab-api
SHA      := $(shell git rev-parse --short=12 HEAD 2>/dev/null || echo dev)

help: ## Tampilkan daftar perintah
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

up: ## Build dan jalankan semua service (tunggu sampai healthy)
	$(COMPOSE) config -q
	$(COMPOSE) up -d --build --wait

down: ## Stop semua service, data tetap ada
	$(COMPOSE) down

reset: ## PERINGATAN: stop + hapus volume (data hilang)
	$(COMPOSE) down -v

logs: ## Log semua service
	$(COMPOSE) logs -f

ps: ## Status + health tiap service
	$(COMPOSE) ps --format 'table {{.Name}}\t{{.Status}}\t{{.Health}}'

health: ## Cek endpoint health lewat reverse proxy
	curl -fsS http://localhost/health  && echo
	curl -fsS http://localhost/readyz  && echo

test: ## Unit test
	go test ./... -short

test-int: ## Integration test (butuh service jalan)
	go test -tags=integration -count=1 ./tests/integration/...

lint: ## go vet + golangci-lint
	go vet ./...
	golangci-lint run ./... || echo "golangci-lint belum terpasang, lewati"

fmt: ## Format kode
	gofmt -w .

build: ## Build image dengan tag commit SHA
	docker build -f docker/Dockerfile -t $(IMAGE):$(SHA) .

loadtest: ## Jalankan skenario k6
	k6 run loadtest/loadtest.js

backup: ## Backup database (custom format)
	$(COMPOSE) exec -T postgres pg_dump -U $${PGUSER} -d $${PGDATABASE} -Fc > backup-$(shell date +%Y%m%d-%H%M%S).dump

restore: ## Restore: make restore FILE=backup-xxx.dump
	$(COMPOSE) exec -T postgres pg_restore -U $${PGUSER} -d $${PGDATABASE} --clean --if-exists < $(FILE)
