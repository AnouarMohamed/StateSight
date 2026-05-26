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
- The `kubectl` collector requests managed-field output explicitly. For drift fields currently supported by the semantic diff engine, the worker reads live-object `metadata.managedFields` and persists manager evidence only when `fieldsV1` contains the exact compared leaf. Qualified annotation, label, Service selector, and resource quantity keys are treated as single Kubernetes map keys; container-image ownership is resolved through the live container name rather than an assumed list index.
- A reported field manager establishes ownership metadata only. It is not treated as proof of who caused the observed difference, and the worker does not invent actor identities when no such signal exists.

## Semantic Diff Boundary

- The worker emits semantic findings for resource presence, replica count, first-container image, annotation, metadata label, Service selector, named pod-template container presence, named environment entry, and named container resource request/limit differences.
- Annotation, label, and Service selector values are compared by exact key; additions and removals are represented explicitly, and findings are ordered deterministically.
- Pod-template containers and environment entries use name-qualified canonical paths. Environment values include literal and `valueFrom` configuration; resource quantities use Kubernetes semantic equality so equivalent representations do not drift.
- Named-container presence and environment-entry findings are aggregate values and therefore do not emit `managedFields` ownership evidence. Request/limit quantity fields are exact leaves and can emit that evidence when present in the live object.
- Volumes and probes are not yet semantically compared, and first-container image paths remain positional for compatibility with existing exact ignore rules.

## Services

- `apps/api`: HTTP entrypoint, routing, request middleware, API contracts.
- `apps/worker`: asynchronous processing for `analyze_application` and `ingest_github_event`.
- `apps/web`: operator-facing UI.

## Data and Queue

- PostgreSQL: source of truth for applications, snapshots, incidents, suppressed findings, scoped ignore rules, evidence, jobs, and event metadata.
- Redis: lightweight queue transport for asynchronous work.
- Workspace-qualified foreign keys prevent applications from binding cross-workspace clusters or sources and prevent application-scoped ignore rules from crossing the tenant boundary.

## Worker Runtime Configuration

- `GIT_BIN`: Git executable used to fetch desired-state repositories.
- `GIT_CACHE_DIR`: temporary checkout parent directory.
- `KUBECTL_BIN`: executable used for cluster collection.
- `ALLOW_SYNTHETIC_LIVE_STATE`: optional demo-only fallback, disabled by default.

## Authentication Boundary

- With `AUTH_REQUIRED=true`, the API starts only after discovering the configured OIDC provider and constructing a verifier for the configured audience.
- Protected API requests require a JWT bearer token validated against the discovered JWKS, issuer, audience, and token lifetime. HTTPS is required for issuer and JWKS retrieval unless a local-development insecure override is explicitly configured.
- Verified `(issuer, subject)` identities resolve through `user_identities` to local users. Workspace access and editor actions remain governed by `workspace_memberships`.
- `X-Workspace-ID` is only a selected scope for workspace collection endpoints; it never establishes caller identity. Resource-addressed endpoints authorize against the resource's persisted workspace.
- The GitHub webhook remains outside the OIDC middleware and is authenticated through its HMAC webhook signature.
- The static web client no longer supplies user identity headers. Interactive browser OIDC login is a separate integration still to be implemented.

## Ignore-Rule Evaluation

- The worker loads active rules applicable to the analyzed application only when analysis produces drift candidates.
- Each baseline `match_expression` is an exact, case-sensitive drift field path after trimming surrounding whitespace.
- Application-managed rules carry an `application_id` and may additionally specify an exact `resource_ref`; legacy rules without an application remain inherited workspace rules.
- Resource-scoped application rules are considered before application-wide rules, then inherited workspace rules. Creation time and ID provide deterministic ordering inside a scope.
- A suppression writes a `suppressed_findings` audit record linked to the analysis snapshots and a snapshot of the rule name/reason, then appears in application details.
- Application details display applicable rules and allow creation, edits, activation changes, and confirmed deletion for application-owned rules; inherited workspace rules are read-only through this surface.
- Edits and deletion govern future analysis only. Existing suppression audit rows store the matching rule name and reason captured when suppression occurred, and rule deletion nulls only their optional rule reference.

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
