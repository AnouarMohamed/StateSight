# Decisions Log

Use this file to record meaningful project decisions as the codebase grows.

## Template

- Date:
- Decision:
- Context:
- Options considered:
- Chosen approach:
- Consequences:

## Entries

- Date: 2026-03-22
  Decision: Product name is `StateSight`.
  Context: Initial drafts used DriftLens. Repository and baseline scaffold will use StateSight naming.
  Options considered: keep DriftLens, switch to StateSight now, switch later.
  Chosen approach: switch now before deeper implementation to reduce rename churn.
  Consequences: package, docs, and service naming use StateSight going forward.

- Date: 2026-03-22
  Decision: Baseline architecture uses one repo with API + worker + web services.
  Context: We needed a practical starting point with clear boundaries and low operational overhead.
  Options considered: monolith only, microservices split, balanced 3-service baseline.
  Chosen approach: 3-service baseline with shared internal packages.
  Consequences: easier incremental growth without over-committing to high service complexity.
