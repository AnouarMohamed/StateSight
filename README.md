# StateSight

StateSight is a GitOps forensic platform for Kubernetes.

Its purpose is to compare desired state from Git with live cluster state, explain drift, group it into incidents, and recommend actions (`ignore`, `monitor`, `investigate`, `reconcile`).

StateSight is **not** a deployment controller and does not replace Argo CD or Flux.

## What This Baseline Includes

- Go API service with versioned routes, request IDs, structured JSON responses, health/readiness, and basic metrics.
- Go worker service that consumes Redis queue jobs and writes deterministic analysis outputs to Postgres.
- React + TypeScript + Vite + Tailwind web app with routed pages and API-backed data loading.
- PostgreSQL migrations for core domain entities.
- Seed workflow with realistic sample data.
- Docker Compose local stack for Postgres, Redis, API, worker, and web.
- Makefile commands for setup, migrate, seed, run, format, test, and docs checks.

## Current Limitations (Intentional)

- Semantic diffing currently covers resource presence, replica counts, first-container images, and annotations; it is not a complete Kubernetes diff engine.
- Live-state collection uses `kubectl` for a limited resource set rather than a Kubernetes client integration.
- Evidence attribution is placeholder data, and stored ignore rules are not evaluated yet.
- GitHub webhook endpoint is baseline-only (not full GitHub App install/auth flow).
- Git desired-state ingestion reads plain YAML/JSON manifests; Helm, Kustomize, Argo CD, and Flux integrations are not implemented.
- No auto-remediation.

## Auth and RBAC Baseline

API supports workspace-aware RBAC boundaries when `AUTH_REQUIRED=true`.

Expected request headers in auth-enabled mode:

- `X-User-ID`
- `X-Workspace-ID`
- `X-User-Email` (optional)

Roles come from `workspace_memberships` (`viewer`, `editor`, `admin`).

## Analysis Safety and Configuration

The worker honors:

- `GIT_BIN` and `GIT_CACHE_DIR` for desired-state checkouts.
- `KUBECTL_BIN` for live-state collection.
- `ALLOW_SYNTHETIC_LIVE_STATE`, which defaults to `false`.

When `kubectl` cannot collect live resources, analysis fails by default. Set `ALLOW_SYNTHETIC_LIVE_STATE=true` only for local pipeline demonstrations; resulting incidents do not represent observations from a cluster.

## Architecture Overview

High-level structure:

- `apps/api`: HTTP API service
- `apps/worker`: async job processor
- `apps/web`: frontend app
- `internal/*`: service internals and pipeline boundaries
- `pkg/*`: reusable domain/model utilities
- `migrations/`: SQL schema migrations
- `scripts/migrate`, `scripts/seed`: operational bootstrap commands

Detailed notes:

- [docs/architecture/overview.md](docs/architecture/overview.md)
- [docs/ARCHITECTURE-NOTES.md](docs/ARCHITECTURE-NOTES.md)

## Local Setup

### 1) Prerequisites

- Go 1.23+
- Node 20+
- Docker + Docker Compose

### 2) Environment

```bash
cp .env.example .env
cp apps/web/.env.example apps/web/.env
```

### 3) Start Infrastructure + Services

```bash
docker compose up --build -d
```

### 4) Run Migrations

```bash
make migrate-up
```

### 5) Seed Sample Data

```bash
make seed
```

The seed data provides a prebuilt incident for UI inspection. Its source repository is illustrative, not a runnable analysis input. Running a real analysis requires an accessible manifest repository and Kubernetes credentials available to the worker. A containerized worker also needs any kubeconfig path referenced by a cluster record mounted inside its container.

### 6) Run Services Locally (optional alternative to containerized app services)

```bash
make api
make worker
make web
```

## Make Commands

- `make help`
- `make setup`
- `make up`
- `make down`
- `make migrate-up`
- `make seed`
- `make api`
- `make worker`
- `make web`
- `make fmt`
- `make test`
- `make lint`
- `make docs-check`

## Required API Endpoints in This Baseline

- `GET /healthz`
- `GET /readyz`
- `GET /api/v1/overview`
- `GET /api/v1/applications`
- `POST /api/v1/applications`
- `GET /api/v1/applications/:id`
- `POST /api/v1/applications/:id/analyze`
- `GET /api/v1/incidents/:id`
- `GET /api/v1/incidents/:id/timeline`
- `POST /api/v1/github/webhook`

## Next Suggested Implementation Steps

1. Replace placeholder evidence attribution with evidence derived from real source and cluster signals.
2. Implement ignore-rule evaluation and API/UI management around persisted rules.
3. Expand normalization, diff coverage, and incident grouping with focused tests.
4. Replace header-trusted identity with an authenticated workspace access flow.
5. Add GitOps rendering/integration support and hardened Kubernetes collection.
