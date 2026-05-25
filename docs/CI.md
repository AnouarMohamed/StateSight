# Continuous Integration and Security Gates

StateSight treats analysis correctness and the integrity of its build path as merge requirements. The workflows in `.github/workflows/` are designed to fail closed on test, dependency, migration, static-analysis, and deployable-image issues.

## Required Pull Request Checks

Configure branch protection on `main` to require these checks before merge:

- `Workflow Policy`
- `Go Quality and Security`
- `Web Quality and Dependency Audit`
- `Database Migration and API Smoke`
- `Container Build and Vulnerability Scan`
- `Dependency Review`
- `CodeQL (go)`
- `CodeQL (javascript-typescript)`

Require pull requests, at least one approving review, and review from code owners. Restrict direct pushes to `main`; workflows, CI scripts, migrations, deployment configuration, and dependency manifests are covered by `.github/CODEOWNERS` because they alter the trust boundary.

## Continuous Integration Workflow

`.github/workflows/go.yml` runs for pull requests, pushes to `main`, and manual invocations.

- Workflow policy: validates workflow files with fixed-version `actionlint` and parses CI shell helpers.
- Go quality and security: verifies modules, formatting, `go vet`, Staticcheck, reachable-vulnerability analysis with `govulncheck`, and race-enabled tests with a retained coverage artifact.
- Web quality and dependency audit: installs strictly from the lockfile, fails on npm advisories at moderate severity or above, runs ESLint, type-checks, and builds production assets.
- Database migration and API smoke: boots PostgreSQL and Redis, applies migrations twice to verify replay safety, seeds data, and exercises scoped ignore-rule creation/list/update through the API.
- Container build and vulnerability scan: builds the deployable API, worker, and web images from digest-pinned bases with reusable Go build caches, verifies web runtime security headers, then fails on fixable high or critical vulnerabilities found by checksum-verified Grype.

Only fixable high or critical container findings block merges. Non-fixable base-image findings remain visible in scanner output and must be evaluated during dependency/image updates rather than waived silently.

## Security Workflow

`.github/workflows/security.yml` adds:

- dependency review on pull requests, blocking new moderate-or-greater vulnerable dependencies in runtime or development scopes;
- CodeQL analysis for Go and JavaScript/TypeScript on pull requests, pushes to `main`, a weekly schedule, and manual invocation.

Actions are referenced by full commit SHA with release comments; Dockerfile bases and Compose/CI service containers are pinned by digest. `.github/dependabot.yml` proposes weekly updates for Actions, Go modules, web packages, Dockerfile base images, and Compose service images so pinning does not become stagnation.

## Local Verification

Before requesting review, run:

```bash
make test
make lint
make test-race
make security-go
make verify-web
make workflow-lint
make script-lint
make docs-check
docker compose config --quiet
```

`./scripts/ci/api-smoke.sh` requires available PostgreSQL and Redis instances through `DATABASE_URL` and `REDIS_URL`. `./scripts/ci/scan-containers.sh` expects already built image names as arguments and downloads a checksum-verified Grype release.

Go verification enumerates the declared Go source roots through `./scripts/ci/go-packages.sh`; it does not descend into the frontend tree, so installing or replacing JavaScript dependencies cannot inject unrelated Go packages into local checks.
