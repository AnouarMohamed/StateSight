# StateSight Architecture Notes (Early Draft)

This is a high-level direction, not a fixed final architecture.

## Baseline Stack

- Frontend: React + TypeScript
- Backend services: Go
- Data: PostgreSQL
- Queue/cache: Redis

## Baseline Services

- web app
- API service
- worker service

## Likely Domain Objects

- Workspace
- Cluster
- SourceDefinition
- Application
- DesiredSnapshot
- LiveSnapshot
- DriftIncident
- DriftField
- EvidenceRecord
- IgnoreRule

## Design Direction

- modular code boundaries
- small, testable packages
- explicit contracts between components
- clear observability hooks
- fail analysis when real live-state collection is unavailable unless demo behavior is explicitly enabled

## Important Note

Architecture will evolve during implementation. Behavior-affecting changes should be recorded in `docs/DECISIONS.md`.
