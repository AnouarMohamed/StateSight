.PHONY: help setup up down migrate-up seed api worker web fmt lint test test-race security-go verify-web workflow-lint script-lint docs-check

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
	@echo "make test-race  # race-enabled Go tests with coverage"
	@echo "make security-go # static and vulnerability analysis for Go"
	@echo "make verify-web # install, audit, lint, type-check, and build web"
	@echo "make workflow-lint # validate GitHub Actions workflow syntax"
	@echo "make script-lint # parse CI shell helper scripts"
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
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*' -not -path './apps/web/node_modules/*' -not -path './apps/web/dist/*')

lint:
	@packages="$$(./scripts/ci/go-packages.sh)" && go vet $$packages

test:
	@packages="$$(./scripts/ci/go-packages.sh)" && go test $$packages

test-race:
	@packages="$$(./scripts/ci/go-packages.sh)" && go test -race -shuffle=on -covermode=atomic -coverprofile=coverage.out $$packages

security-go:
	@packages="$$(./scripts/ci/go-packages.sh)" && go run honnef.co/go/tools/cmd/staticcheck@v0.7.0 $$packages
	@packages="$$(./scripts/ci/go-packages.sh)" && go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 $$packages

verify-web:
	cd apps/web && npm ci --ignore-scripts
	cd apps/web && npm audit --audit-level=moderate
	cd apps/web && npm run lint
	cd apps/web && npm run typecheck
	cd apps/web && npm run build

workflow-lint:
	go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12 -color

script-lint:
	bash -n scripts/ci/*.sh

docs-check:
	test -f README.md
	test -f CONTRIBUTING.md
	test -f docs/PROJECT-OVERVIEW.md
	test -f docs/ROADMAP.md
	test -f docs/ARCHITECTURE-NOTES.md
	test -f docs/WORKFLOW.md
	test -f docs/CI.md
	test -f docs/architecture/overview.md
	@echo "Docs check passed."
