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
- manage application-scoped ignore rules with optional exact resource matching and lifecycle controls
- enforce merge-gating CI, security analysis, migration smoke verification, and deployable-image scanning
- derive incident evidence from Git revisions, live collection sources, and exact Kubernetes field ownership
- present drift and provenance in a compact evidence-first operator console
- verify OIDC bearer identities at the API boundary before applying workspace RBAC
- enforce workspace-qualified application and scoped-rule relationships in persistence
- compare metadata labels and Service selectors with deterministic key-level findings and exact field-ownership lookup
- compare named pod-template container presence, environment entries, and Kubernetes-normalized resource quantities

## Next: Trustworthy Analysis

- correlate provenance with audit, deployment, and controller signals without inferring causality from ownership alone
- design deliberate workspace-wide ignore-rule administration with authorization and blast-radius controls
- improve incident grouping and semantic normalization/diff coverage
- complete protected operator access with browser OIDC login and managed identity provisioning workflows
- expand semantic diff coverage for volumes, probes, and a compatible named-container image path
- add alerting integrations (webhooks, Slack) for new incidents and severity changes

## Later: Integrations and Operations

- support GitOps rendering and controller integrations
- deepen Kubernetes collection and multi-cluster operational controls
- harden reliability, observability, deployment, and scale behavior
- add multi-cluster views with clear scoping and per-cluster drift summaries

Roadmap is expected to evolve as implementation learns from real usage.
