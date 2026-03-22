.PHONY: help setup up down migrate-up seed api worker web fmt lint test docs-check

help:
	@echo "StateSight baseline commands"
	@echo ""
	@echo "make setup      # install web dependencies"
	@echo "make up         # start docker compose stack"
	@echo "make down       # stop docker compose stack"
	@echo "make migrate-up # run SQL migrations"
	@echo "make seed       # seed baseline demo data"
	@echo "make api        # run API locally"
	@echo "make worker     # run worker locally"
	@echo "make web        # run web locally"
	@echo "make fmt        # gofmt all go files"
	@echo "make lint       # go vet"
	@echo "make test       # go tests"
	@echo "make docs-check # verify key docs exist"

setup:
	cd apps/web && npm install

up:
	docker compose up --build -d

down:
	docker compose down

migrate-up:
	go run ./scripts/migrate

seed:
	go run ./scripts/seed

api:
	go run ./apps/api

worker:
	go run ./apps/worker

web:
	cd apps/web && npm run dev

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

lint:
	go vet ./...

test:
	go test ./...

docs-check:
	test -f README.md
	test -f CONTRIBUTING.md
	test -f docs/PROJECT-OVERVIEW.md
	test -f docs/ROADMAP.md
	test -f docs/ARCHITECTURE-NOTES.md
	test -f docs/WORKFLOW.md
	test -f docs/architecture/overview.md
	@echo "Docs check passed."
