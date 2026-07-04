---
doc_id: THR-S5-HANDOFF-2026-05-08
title: S5 Lifecycle Interfaces Failures Handoff
status: active
role: sprint-handoff
---

# S5 Lifecycle, Interfaces, and Failures Handoff

## Session Metadata

| Field | Value |
|---|---|
| Sprint | S5: Lifecycle, interfaces, and failures |
| Status | `complete` |
| Repository root | `/home/askahn/code/cartulary` |
| Branch | `main...origin/main` |
| HEAD revision | `3cfde09450b3217f0cb9613a78a708e866ad37f5` |
| Timestamp recorded | `2026-05-08T22:48:54-04:00` |
| Runtime platform | `Linux DeskRip 6.6.114.1-microsoft-standard-WSL2 x86_64 GNU/Linux` |
| Initial dirty state | clean before S5 recovery-doc edits |
| Recovery write boundary | `docs/testing-harness-spec-recovery-docs/**` |

## Inputs Used

- S2: `entrypoint-command-map.md`, `sequencing-assumption-list.md`, S2
  handoff, and S2 audit follow-ups.
- S3: `artifact-ownership-matrix.md`, `cleanup-lifecycle-matrix.md`,
  `shared-state-hazard-list.md`, and S3 ambiguity/source-limit rows.
- S4: `service-lifecycle-map.md`, `environment-contract-observations.md`,
  `resource-allocation-register.md`, S4 hazards, S4 handoff, and S4 audit
  follow-ups.
- Process controls: `recovery-charter.md`,
  `04-registers-and-checklists.md`, `source-limit-log.md`, and
  `ambiguity-register.md`.
- Static source evidence from Make targets, scheduler scripts, service/browser
  wrappers, reset scripts, output helpers, package manifests, CI helpers, and
  retained-artifact readers.

## Outputs Produced

| Output | Status | Blockers | Findings | Source limits | Ambiguities | Handoff notes | Evidence status |
|---|---|---|---|---|---|---|---|
| `observable-interface-map.md` | `complete` | CI provider annotations and controlled failure bundles unavailable. | `OI-0001` through `OI-0017` map stdout, stderr, exit status, reports, logs, failure artifacts, durable artifacts, local/CI stability, authority class, and evidence for all S2 major entrypoint families. | broad gates, Docker/Compose/browser/reset/cleanup/runtime failures, stale retained artifacts, `.github/**`. | package scripts, reset authority, CI annotations, local-dev persistence. | S6/S7 can use OI rows as caller-visible output inventory. | `observed`, `observed/source_limit`, `source_limit` |
| `structured-output-schema-notes.md` | `complete` | Playwright failure bundle schemas, CI annotations, and service event exact schema are incomplete or unavailable. | `SCHEMA-0001` through `SCHEMA-0020` classify every machine-consumed output as `complete`, `partial`, `source_observed`, `schema_unknown`, or `authority_unknown`; unknown machine outputs have explicit `TODO: schema_unknown`. | selected failure-only artifacts unavailable; provider annotations absent; retained artifacts unselected. | external provider annotations and direct package reports. | S7/NLSpec should not promote unknown schemas without owner decision or selected evidence. | `observed`, `source_limit`, `authority_unknown` |
| `output-consumer-map.md` | `complete` | none beyond linked source limits. | `CONS-0001` through `CONS-0014` identify Make, schedulers, testservices, browser wrappers, summarizers, explain tools, drift/baseline tools, CI/release gates, package users, and maintainers as consumers. | direct package and CI provider consumers remain authority/source-limited. | package scripts and provider annotations. | Use as consumer matrix for NLSpec output contracts. | `observed`, `observed/source_limit` |
| `harness-lifecycle-map.md` | `complete` | Runtime cleanup/readiness/timeout/interrupt examples unavailable. | `LIFE-0001` through `LIFE-0015` cover Make phases, aggregates, check scheduler, service-backed scheduler, Go, frontend/Vitest, browser stack, reset boundary, testservices, local dev, package scripts, CI, investigation, maintenance, and uninvoked/source-limited entrypoints. | runtime cleanup, reaper completion, broad gates, service/browser execution. | local-dev and package-script contract status, reset route, supported platform. | S6 should use lifecycle phases as probe boundaries. | `observed`, `observed/source_limit`, `source_limit` |
| `phase-transition-table.md` | `complete` | Runtime transition evidence unavailable for broad/service/browser/reset/cleanup cases. | `PT-0001` through `PT-0035` record gates, failure transitions, cleanup transitions, and transition outputs for all `LIFE-*` rows. | event ordering, real host capacity, timeout cleanup, parent death, port release, reaper completion. | reset route, platform profile, package scripts, local-dev persistence. | S6 should validate transition timing and cleanup claims before normativity. | `observed`, `observed/source_limit`, `source_limit` |
| `partial-completion-state-list.md` | `complete` | Runtime partial examples unavailable. | `PCS-0001` through `PCS-0019` list partial service, scheduler, browser, reset, package, retained-artifact, maintenance, Go, and Vitest states with cleanup expectations and retryability. | cleanup/reaper completion, active connection cleanup, port release, stale provenance, failure-only schemas. | reset side effects, package scripts, local-dev persistence, external caches. | S6 should treat these as hazard candidates; S8 should resolve authority-bound rows. | `observed`, `observed/source_limit` |
| `failure-class-taxonomy.md` | `complete` | Controlled failure examples unavailable. | Required classes are defined with detection, reporting location, exit/report/artifact surface, retryability, ownership, and cleanup/artifact/resource follow-up. | runtime-only timeout, interrupt, cleanup, unsupported platform, unknown failure. | authority-required class captures owner decisions. | Use as classification vocabulary for S6 and NLSpec. | `observed`, `observed/source_limit`, `source_limit`, `maintainer_decision_required` |
| `failure-mode-register.md` | `complete` | Controlled failures and provider CI evidence unavailable. | `FAIL-0001` through `FAIL-0029` register recurring/plausible failure modes from S2/S3/S4/S5 evidence with explicit retryability. | runtime failures, cleanup, interrupts, stale artifacts, failure bundles, provider annotations. | package scripts, env precedence, platform profile, reset/janitor/local-dev authority. | S6 should probe runtime rows; S8 should settle authority rows before NLSpec normativity. | `observed`, `observed/source_limit`, `source_limit` |

## Validation Results

- Every machine-consumed output identified in `observable-interface-map.md` has
  a corresponding `SCHEMA-*` row or an exact `TODO: schema_unknown`.
- Every major entrypoint family from S2 `EP-0002` through `EP-0019`, cleanup
  and maintenance `EP-0020`, and uninvoked/source-limited `EP-0021` has
  lifecycle and terminal-state coverage.
- Product assertion failures are explicitly separated from harness operational
  failures in `failure-class-taxonomy.md` and `failure-mode-register.md`.
- Retryability is explicit for every failure class and failure mode.
- Durable retained artifacts are marked as requiring explicit run identity or
  as investigation-only/source-limited.
- Source-limited runtime claims remain source-limited.
- No new source-limit or ambiguity rows were required beyond the existing
  register coverage; S5 rows link unresolved items to those registers and to
  S6/S8 handoff.

## Audit Follow-Up

`AUD-S5-BLOCK-0001` is resolved by correcting the affected
`failure-mode-register.md` linked output/schema references. S6 and S7 may
consume the corrected failure-mode register while preserving all existing
source-limit and authority-routing notes.

## Handoff To S6

S6 should consume `harness-lifecycle-map.md`,
`phase-transition-table.md`, `partial-completion-state-list.md`,
`failure-class-taxonomy.md`, `failure-mode-register.md`, S4 `HAZ-S4-*`
rows, and the linked `SL-*` rows.

Priority S6 probes:

- cleanup on timeout, interrupt, cancellation, parent death, and child failure;
- detached reaper scheduling versus verified completion;
- service readiness timeouts, retry exhaustion, and cleanup after partial
  service startup;
- browser process-group cleanup, port release, backend/frontend readiness, and
  Playwright failure artifact retention;
- scheduler event ordering, sibling drain behavior, and logical lanes versus
  real host/service capacity;
- stale fixture janitor behavior, active DB connection cleanup, reset partials,
  and fixture/resource leak handling.

## Handoff To S7

S7 and the later NLSpec draft can safely consume the caller-visible output
inventory, lifecycle phase vocabulary, terminal-state taxonomy, and consumer
map. S7 should not treat `schema_unknown`, `authority_unknown`, or
`source_limit` rows as normative schemas or guarantees.

## Handoff To S8

S8 should resolve maintainer-authority rows before NLSpec turns them into
contracts:

- reset route authority and reset side effects;
- package scripts as first-class caller-visible contracts or developer
  convenience surfaces;
- local-dev Compose persistence and cleanup expectations;
- public environment variable precedence;
- stale janitor destructive bounds;
- supported platform/tool profile;
- external Go caches and other state outside repo-local cleanup;
- CI provider annotations and workflow ownership if `.github/**` appears.

## Blockers

| Blocker | Status | Owner |
|---|---|---|
| Broad gates were not executed. | `source_limit` | S6/runtime validation |
| Service-backed, browser, Docker, Compose, reset, cleanup, timeout, and interrupt behavior were not exercised. | `source_limit` | S6/runtime validation |
| Controlled failure examples and failure-only bundle schemas were unavailable. | `source_limit` | S6/S7 with selected artifacts |
| `.github/**` was absent, so CI provider annotations were not inspectable. | `source_limit` | S8/CI owner |
| Direct package-script contract status is undecided. | `maintainer_decision_required` | S8 |

## Implementation-Change Audit

S5 changed recovery documentation only under
`docs/testing-harness-spec-recovery-docs/**`. It did not modify harness
implementation, command behavior, exit codes, report schemas, CI workflows,
generated artifacts, fixtures, lockfiles, cleanup scripts, or runtime state,
and it did not run service-backed, browser, Docker, Compose, reset, cleanup,
formatter, generator, baseline refresh, or broad gate commands.
