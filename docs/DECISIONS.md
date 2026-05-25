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

- Date: 2026-05-25
  Decision: Synthetic live-state collection is disabled by default and must be explicitly enabled for demonstrations.
  Context: A forensic analysis must not emit incidents from invented live resources when `kubectl` is unavailable or misconfigured.
  Options considered: always fall back to synthetic resources, remove synthetic behavior entirely, retain it behind an explicit runtime option.
  Chosen approach: fail live-state collection by default and permit `ALLOW_SYNTHETIC_LIVE_STATE=true` only for local demonstrations.
  Consequences: real analyses fail clearly when cluster access is unavailable; demo output can still be generated but must be treated as synthetic.

- Date: 2026-05-25
  Decision: Initial ignore-rule evaluation supports exact field-path matching within a workspace.
  Context: The schema contains persisted ignore rules, but the worker previously ignored them. A broad or implicit expression language would make suppression difficult to reason about before rule-management and audit surfaces exist.
  Options considered: keep rules inactive until UI work, interpret expressions as glob or regular-expression patterns, implement exact field-path matching first.
  Chosen approach: load active rules for an application's workspace and suppress only candidates whose canonical field path exactly matches `match_expression`.
  Consequences: common known-noise fields can be suppressed deterministically across a workspace; wildcard matching remains out of scope, while later management work can narrow rule reach.

- Date: 2026-05-25
  Decision: Suppressed findings are persisted as audit records instead of existing only in logs.
  Context: Ignore-rule evaluation prevents incident creation, but operators still need an inspectable record of drift that was withheld and the rule that caused it.
  Options considered: log suppression only, model suppressed findings as ordinary incidents, store dedicated suppression audit records.
  Chosen approach: store a `suppressed_findings` record linked to desired/live snapshots and capture matching rule name/reason at suppression time.
  Consequences: application details can show hidden drift without conflating it with actionable incidents; future rule editing or deletion does not erase the recorded explanation.

- Date: 2026-05-25
  Decision: Newly managed ignore rules are application-scoped with optional exact resource restriction.
  Context: Workspace-wide field suppression is too broad for routine operator management, but existing rules must keep their behavior after an upgrade.
  Options considered: keep workspace-only rules, require every rule to specify one resource, support application scope with optional exact resource matching while retaining legacy rows.
  Chosen approach: API/UI-created rules always reference one application and may store one exact `resource_ref`; null-application rows remain inherited workspace rules and appear read-only in application details.
  Consequences: operators can suppress expected drift with narrower blast radius, resource-specific rules take precedence over broader effective rules, and editing or administering inherited workspace rules remains future work.

- Date: 2026-05-25
  Decision: Merge gating includes security, database smoke, and deployable-image validation.
  Context: A forensic platform can hide or misattribute operational drift if its dependencies, migrations, or artifacts are not continuously verified. The original workflow exercised only ordinary Go tests and `go vet`.
  Options considered: retain minimal Go CI, add quality checks without security enforcement, require reproducible cross-stack and security gates before merge.
  Chosen approach: require Go race/static/vulnerability checks, audited frontend builds, migration and API smoke verification, production-image scanning, dependency review, and CodeQL; pin workflow actions to reviewed commits.
  Consequences: pull requests receive slower but materially stronger validation; branch protection must require the documented checks, and dependency/image updates become routine maintenance rather than deferred risk.

- Date: 2026-05-25
  Decision: Incident evidence records provenance and exact field ownership without claiming unobserved causality.
  Context: The baseline generated plausible but fabricated actor evidence for every incident, which is incompatible with forensic use.
  Options considered: keep placeholder evidence until audit integration exists, remove evidence records entirely, or persist only signals available from the analyzed Git and Kubernetes inputs.
  Chosen approach: store Git revision and live-collection provenance for each incident; add `managedFields` manager evidence only when the live resource owns the exact compared field; mark synthetic fallback evidence untrusted.
  Consequences: incident evidence is truthful with current inputs and can support future correlation, but attributing who caused drift requires additional audit, deployment, or controller signals.
