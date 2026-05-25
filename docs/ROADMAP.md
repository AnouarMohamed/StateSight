# StateSight Roadmap

## Completed Baseline

- repository structure, collaboration documents, local Docker Compose workflow
- Go API and asynchronous Redis-backed worker with PostgreSQL persistence
- React operator UI for overview, applications, incidents, and timelines
- Git manifest fetcher and `kubectl` live-state collector
- initial resource normalization and semantic diff engine
- baseline workspace RBAC boundary and GitHub webhook ingestion

## Current: Baseline Stabilization

- keep documentation aligned with implemented capabilities and limitations
- enforce real-live-state collection by default; synthetic results are explicit demo behavior only
- pass worker runtime configuration to source and cluster adapters
- execute active workspace ignore rules for exact drift field-path matches
- persist and surface audit records for suppressed findings
- manage application-scoped ignore rules with optional exact resource matching
- establish CI and focused tests for critical analysis boundaries

## Next: Trustworthy Analysis

- replace placeholder evidence attribution with real provenance signals
- extend ignore-rule administration with edit/delete flows and deliberate workspace-wide controls
- improve incident grouping and semantic normalization/diff coverage
- replace trusted request headers with a real authentication integration

## Later: Integrations and Operations

- support GitOps rendering and controller integrations
- deepen Kubernetes collection and multi-cluster operational controls
- harden reliability, observability, deployment, and scale behavior

Roadmap is expected to evolve as implementation learns from real usage.
