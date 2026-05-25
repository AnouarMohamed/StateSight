# StateSight Baseline Architecture Overview

## Purpose

StateSight baseline provides a minimal but extensible end-to-end path:

1. API accepts analyze/webhook requests.
2. API records job metadata in Postgres and enqueues job message in Redis.
3. Worker consumes jobs and writes snapshots/incidents/evidence.
4. Web app reads API data and renders overview, application, and incident pages.

The current worker obtains desired state by cloning configured Git sources and obtains live state through `kubectl`. Failure to collect live state fails an analysis by default; synthetic live state is an explicit local-demo option and is not trustworthy forensic evidence.

## Services

- `apps/api`: HTTP entrypoint, routing, request middleware, API contracts.
- `apps/worker`: asynchronous processing for `analyze_application` and `ingest_github_event`.
- `apps/web`: operator-facing UI.

## Data and Queue

- PostgreSQL: source of truth for applications, snapshots, incidents, evidence, jobs, and event metadata.
- Redis: lightweight queue transport for asynchronous work.

## Worker Runtime Configuration

- `GIT_BIN`: Git executable used to fetch desired-state repositories.
- `GIT_CACHE_DIR`: temporary checkout parent directory.
- `KUBECTL_BIN`: executable used for cluster collection.
- `ALLOW_SYNTHETIC_LIVE_STATE`: optional demo-only fallback, disabled by default.

## Ignore-Rule Evaluation

- The worker loads active rules for the application's workspace only when an analysis produces drift candidates.
- Each baseline `match_expression` is an exact, case-sensitive drift field path after trimming surrounding whitespace.
- A rule applies to that field path across its workspace; resource- or application-specific selectors are not supported yet.
- Rules are considered in creation order, with ID as a deterministic tie-breaker; the first match suppresses incident creation for that candidate.
- Suppressed candidates are currently reported in worker logs. Persisted suppression auditing and rule-management APIs are future work.

## Key Internal Boundaries

- `internal/sourceingest`: desired-state ingestion boundary.
- `internal/k8scollect`: live-state collection boundary.
- `internal/normalize`: canonicalization boundary.
- `internal/diff`: diffing boundary.
- `internal/incidents`: finding-to-incident grouping boundary.
- `internal/evidence`: attribution/evidence boundary.
- `internal/scoring`: recommendation boundary.
- `internal/ignorerules`: suppression boundary.
- `internal/timelines`: timeline construction boundary.
- `internal/storage`: Postgres repository layer.
- `internal/apihttp`: transport layer and API handlers.
- `internal/jobs`: queue and worker processor.

## Baseline Design Principles

- keep package boundaries explicit
- keep core flow observable and traceable
- keep stubs realistic and replaceable
- avoid over-abstraction at this stage
