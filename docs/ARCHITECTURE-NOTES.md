# Architecture Notes (Early Draft)

This is a high-level direction, not a fixed final architecture.

## Likely Stack

- Frontend: React + TypeScript
- Backend services: Go
- Data: PostgreSQL
- Queue/cache: Redis

## Likely Services

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

## Important Note

Architecture will evolve during the baseline scaffold and early implementation phases. Changes should be recorded in `docs/DECISIONS.md`.
