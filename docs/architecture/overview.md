# StateSight Baseline Architecture Overview

## Purpose

StateSight baseline provides a minimal but extensible end-to-end path:

1. API accepts analyze/webhook requests.
2. API records job metadata in Postgres and enqueues job message in Redis.
3. Worker consumes jobs and writes snapshots, incidents or suppression audits, and evidence.
4. Web app reads API data and renders overview, application, and incident pages.

The current worker obtains desired state by cloning configured Git sources and obtains live state through `kubectl`. Failure to collect live state fails an analysis by default; synthetic live state is an explicit local-demo option and is not trustworthy forensic evidence.

## Evidence Provenance

- Every persisted incident records the Git source, source path, and analyzed revision that supplied its desired-state comparison.
- Every persisted incident records the live collection source. `kubectl` observations are trusted collection evidence; explicit synthetic fallback records are marked untrusted.
- The `kubectl` collector requests managed-field output explicitly. For drift fields currently supported by the semantic diff engine, the worker reads live-object `metadata.managedFields` and persists manager evidence only when `fieldsV1` contains the exact compared path. Container-image ownership is resolved through the live container name rather than an assumed list index.
- A reported field manager establishes ownership metadata only. It is not treated as proof of who caused the observed difference, and the worker does not invent actor identities when no such signal exists.

## Services

- `apps/api`: HTTP entrypoint, routing, request middleware, API contracts.
- `apps/worker`: asynchronous processing for `analyze_application` and `ingest_github_event`.
- `apps/web`: operator-facing UI.

## Data and Queue

- PostgreSQL: source of truth for applications, snapshots, incidents, suppressed findings, scoped ignore rules, evidence, jobs, and event metadata.
- Redis: lightweight queue transport for asynchronous work.

## Worker Runtime Configuration

- `GIT_BIN`: Git executable used to fetch desired-state repositories.
- `GIT_CACHE_DIR`: temporary checkout parent directory.
- `KUBECTL_BIN`: executable used for cluster collection.
- `ALLOW_SYNTHETIC_LIVE_STATE`: optional demo-only fallback, disabled by default.

## Ignore-Rule Evaluation

- The worker loads active rules applicable to the analyzed application only when analysis produces drift candidates.
- Each baseline `match_expression` is an exact, case-sensitive drift field path after trimming surrounding whitespace.
- Application-managed rules carry an `application_id` and may additionally specify an exact `resource_ref`; legacy rules without an application remain inherited workspace rules.
- Resource-scoped application rules are considered before application-wide rules, then inherited workspace rules. Creation time and ID provide deterministic ordering inside a scope.
- A suppression writes a `suppressed_findings` audit record linked to the analysis snapshots and a snapshot of the rule name/reason, then appears in application details.
- Application details display applicable rules and allow creation or activation changes for application-owned rules; inherited workspace rules are read-only through this surface.

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
