---
doc_id: THR-S6-AUDIT-PLAN-IMPLEMENTATION-2026-05-09
title: S6 Hazards Resources Timing Audit Plan Implementation
status: complete
role: recovery-audit
---

# S6 Hazards, Resources, and Timing Audit Plan Implementation

## Audit Objective

Verify that S6 produced a complete, accurate, evidence-grounded account of the
testing harness's concurrency hazards, shared mutable resources, race
conditions, timing assumptions, retry behavior, timeout behavior, sharding
behavior, resource conflicts, and authority-sensitive behavior needed by S7.

## Audit Verdict

`pass_with_source_limits_preserved`

S6 is ready to support S7 as documentation evidence, provided S7 preserves the
recorded `source_limit` and `maintainer_decision_required` classifications. No
documentation issue was found that should block progression to S7.

## Scope and Non-Scope

In scope:

- `race_timing_resource_register`: `race-timing-resource-register.md`.
- `concurrency_model_notes`: `concurrency-model-notes.md`.
- `timeout-retry-register.md`.
- `preservation_matrix`: `preservation-matrix.md`.
- S6 updates to `ambiguity-register.md`.
- `harness_authority_map`: `harness-authority-map.md`.
- `main_spec_conflict_list`: `main-spec-conflict-list.md`.
- S6 handoff, source-limit log, and the existing S6 audit.
- S3, S4, and S5 inputs for resources, lifecycle, cleanup, artifacts,
  failures, partial-completion states, observable interfaces, and sequencing.

Out of scope:

- No harness implementation changes.
- No changes to timing, timeout, retry, sharding, scheduler behavior, fixtures,
  cleanup behavior, resource allocation, reset behavior, tests, generated
  outputs, lockfiles, or package-manager state.
- No service-backed, Docker, Compose, browser, reset, cleanup, formatter,
  generator, baseline-refresh, snapshot-update, package-test, broad
  verification, or S7 drafting command execution.

## Source Documents and Repository Areas Inspected

Primary recovery documents:

- `03-sprint-plan.md` S6 validation and handoff sections.
- `race-timing-resource-register.md`, `concurrency-model-notes.md`,
  `timeout-retry-register.md`, `shared-state-hazard-list.md`,
  `preservation-matrix.md`, `harness-authority-map.md`,
  `main-spec-conflict-list.md`, `ambiguity-register.md`,
  `source-limit-log.md`, and the S6 handoff.
- S3/S4/S5 inputs: `artifact-ownership-matrix.md`,
  `cleanup-lifecycle-matrix.md`, `service-lifecycle-map.md`,
  `resource-allocation-register.md`, `failure-mode-register.md`,
  `partial-completion-state-list.md`, `observable-interface-map.md`, and
  `sequencing-assumption-list.md`.

Repository areas inspected by static search or cited S6 evidence:

- `Makefile`, task-surface manifests, scheduler manifests, scheduler resource
  registry, generated task surface, and generated artifact policy.
- `scripts/**`, including scheduler, sharding, Playwright/browser, process
  lifecycle, artifact discovery, drift, and runner helpers.
- `tools/testservices/**`, `internal/testutil/**`,
  `internal/app/test_runtime_reset.go`, browser E2E configs and fixtures,
  package manifests, `docker-compose.dev.yml`, and `.gitignore`.
- Core 00 through Core 04 authority posture through Core 00 and
  `docs/domain.md` status text.

## Output Presence and Evidence Posture

| Check | Result | Evidence |
|---|---|---|
| Required S6 outputs exist. | pass | Static file-presence check found all required S6 artifacts, S6 handoff, source-limit log, and existing S6 audit. |
| S6 row inventories are concrete and bounded. | pass | Current S6 docs contain 21 `RTR-*`, 33 `TMR-*`, 14 `CONC-*`, 20 `PRES-*`, 15 `AUTH-*`, 10 `MSC-*`, 10 `AMB-FU-*`, and 15 `HAZ-FU-*` unique rows. |
| Evidence statuses are row-level and preserve uncertainty. | pass | S6 rows use `observed`, `observed/source_limit`, `source_limit`, and `maintainer_decision_required`; no runtime-sensitive row was found upgraded without runtime evidence. |
| No-change posture is explicit. | pass | S6 docs and handoff state that S6 did not run mutating writer, cleanup, service, browser, reset, formatter, generator, baseline, snapshot, package-test, or broad verification commands. |
| Existing dirty state was not treated as audit evidence of implementation change. | pass | `git status --short --branch` showed pre-existing modified recovery-doc files before this audit report was written; no implementation path was touched by this audit. |

## Shared Mutable Resource Completeness

Audit criterion: every shared mutable resource identified by S3/S4 must map to
an S6 hazard row and, where timing applies, a timing/concurrency row or an
explicit source-limit/exclusion.

| Resource class | S6 hazard coverage | Timing/concurrency coverage | Audit result |
|---|---|---|---|
| Scheduler host/service/process/browser lanes | `RTR-0001`, `RTR-0002`, `RTR-0015` | `CONC-0002`, `CONC-0003`, `TMR-0020`, `TMR-0021` | pass |
| Generated schedules, ledgers, generated task/code outputs | `RTR-0002` | `CONC-0013`, `TMR-0027` | pass |
| Retained result roots, run summaries, service events, fixture reports | `RTR-0016` | `CONC-0013`, `TMR-0027` | pass |
| Docker/testcontainers leases, managed containers, detached reaper | `RTR-0004`, `RTR-0007`, `RTR-0008`, `RTR-0009` | `CONC-0005`, `TMR-0001`, `TMR-0002`, `TMR-0004`, `TMR-0017`, `TMR-0028` | pass |
| Postgres templates, cloned DBs, package DBs, transactions, reset tables | `RTR-0005`, `RTR-0009`, `RTR-0010`, `RTR-0018` | `CONC-0006`, `TMR-0002`, `TMR-0003`, `TMR-0006`, `TMR-0017`, `TMR-0022` | pass |
| MinIO buckets, package buckets, prefixes, object namespaces | `RTR-0006`, `RTR-0009`, `RTR-0010` | `CONC-0007`, `TMR-0004`, `TMR-0005`, `TMR-0017`, `TMR-0023` | pass |
| Browser/dev/Compose ports, processes, process groups, runtime roots | `RTR-0011`, `RTR-0012`, `RTR-0019` | `CONC-0008`, `CONC-0009`, `CONC-0012`, `TMR-0008`, `TMR-0009`, `TMR-0010`, `TMR-0013`, `TMR-0014`, `TMR-0015`, `TMR-0029` | pass |
| Playwright state dir, lock, worker admin manifest, cleanup markers, sessions | `RTR-0013` | `CONC-0010`, `TMR-0011`, `TMR-0012`, `TMR-0024`, `TMR-0025` | pass |
| Reset boundary files and app test runtime reset state | `RTR-0014` | `TMR-0016`, `TMR-0030` | pass |
| Shell harness scratch dirs | `RTR-0020` | `CONC-0014`, `TMR-0031` | pass |
| External Go caches | `RTR-0017` | `CONC-0014`, `TMR-0026` | pass |
| Direct package-script artifacts and tool-native outputs | `RTR-0015` | `CONC-0011`, `TMR-0032` | pass |
| Visual snapshots and browser platform baselines | `RTR-0021` | `TMR-0012`, `TMR-0025` | pass |

Allocation details remain in `resource-allocation-register.md`; S6 links those
rows instead of duplicating allocation ownership.

## Timing, Retry, Wait, and Sharding Coverage

Audit criterion: every fixed sleep, timeout, retry, polling loop, debounce or
watch trigger, readiness check, cleanup wait, signal wait, lock wait, watchdog,
and sharding assumption must be documented in `TMR-*`, `CONC-*`, or an explicit
source-limit row.

| Timing class | S6 coverage | Audit result |
|---|---|---|
| Fixed sleeps and polling loops | `TMR-0003`, `TMR-0005`, `TMR-0008`, `TMR-0011`, `TMR-0013`, `TMR-0014`, `TMR-0015`, `TMR-0018`, `TMR-0019`, `TMR-0020`, `TMR-0021`, `TMR-0025`, `TMR-0029` | pass |
| Timeouts and retry windows | `TMR-0001` through `TMR-0006`, `TMR-0012`, `TMR-0017`, `TMR-0018`, `TMR-0019`, `TMR-0022`, `TMR-0023`, `TMR-0024`, `TMR-0028`, `TMR-0030` | pass |
| Readiness checks | `TMR-0003`, `TMR-0005`, `TMR-0008`, `TMR-0013`, `TMR-0014`, `TMR-0015` | pass |
| Cleanup, signal, process, and port waits | `TMR-0009`, `TMR-0010`, `TMR-0017`, `TMR-0028`, `TMR-0029`, `TMR-0031` | pass |
| Lock waits and stale-lock behavior | `TMR-0011`, `TMR-0018`, `TMR-0019`; linked to `RTR-0013` and `RTR-0018` | pass |
| Watchdog, watch, and debounce surfaces | `TMR-0020`, `TMR-0025`, `TMR-0033`; debounce absence/grouping is explicit | pass |
| Sharding and worker assumptions | `CONC-0004`, `CONC-0008`, `CONC-0009`, `CONC-0010`; `TMR-0018`, `TMR-0019`, `TMR-0021`, `TMR-0032` | pass |

The audit found no S6 statement that treats scheduler lanes as concrete host,
Docker, Postgres, MinIO, browser, process, or port capacity guarantees.
`RTR-0001`, `CONC-0002`, `CONC-0003`, `AUTH-0011`, and `PRES-0003` consistently
state the scheduler-accounting boundary.

## Failure and Partial-Completion Linkage

Audit criterion: recurring failures and partial-completion states must be
linked to hazards only where evidence supports the connection, or retained as
source-limited/out-of-scope.

| Area | Result | Evidence |
|---|---|---|
| Failure-mode linkage | pass | `race-timing-resource-register.md` links `FAIL-0005` through `FAIL-0028` groups to `RTR-*` hazards without claiming new runtime examples. |
| Partial-completion linkage | pass | `PCS-0001` through `PCS-0019` are linked to S6 hazards for startup, cleanup, reset, retained artifact, package-script, scheduler, and browser partial states. |
| Recurring failure interpretation | pass | S6 distinguishes product assertion failures, harness operational failures, authority-required failures, and source-limited runtime failures. |
| Unsupported runtime proof | pass | Cleanup completion, interrupt behavior, detached reaper completion, active DB connection cleanup, stale janitor execution, and port release remain source-limited. |

## Preservation, Classification, and Authority Checks

Audit criterion: every major subsystem must have preservation classification;
behavior classification must be consistent; authority-required behavior must
have a maintainer or owner question.

| Check | Result | Evidence |
|---|---|---|
| Major subsystem classification | pass | `PRES-0001` through `PRES-0020` cover command surface, generated artifacts, schedulers, Go sharding/report locks, services, Postgres, MinIO, browser stack, Playwright state, reset route, package scripts, local-dev services, stale janitors, env contracts, platform/tools, caches, retained artifacts, visual snapshots, CI gaps, and scratch dirs. |
| Classification vocabulary consistency | pass | `preservation-matrix.md` defines `preserve`, `preserve_with_clarification`, `compatibility_only`, `accidental`, `deprecated`, `redesign_required`, `authority_required`, and `exclude_from_contract`; current rows use allowed values and do not overclaim accidental or deprecated behavior. |
| Authority-required owner prompts | pass | `AUTH-0003` through `AUTH-0015` and `PRES-0010` through `PRES-0018` contain concrete owner questions for direct package scripts, reset route, local-dev services, env precedence, stale janitors, platform profile, external caches, retained artifact selection, visual snapshots, and CI provider gaps. |
| Main-spec-sensitive routing | pass | `MSC-0001` through `MSC-0010` route reset route, local credentials, local-dev topology, env contracts, generated artifacts, domain vocabulary, visual platform behavior, stale janitors, retained artifacts, and CI provider gaps without redefining Core 00 through Core 04. |

## Unsupported Assumptions, Ambiguity, and Source Limits

The audit found the following gaps correctly recorded as source limits or owner
questions, not blockers:

| Follow-up area | Current routing | Why non-blocking for S7 |
|---|---|---|
| Live Docker/testcontainers, Postgres, MinIO, Compose, browser, reset, and service-backed readiness | `SL-0012`, `SL-FU-0001`, `RTR-0004`, `TMR-*` readiness rows | S6 keeps these as static/source-limited and does not claim runtime observation. |
| Timeout, interrupt, parent-death, detached reaper, active DB cleanup, stale janitor, and port-release behavior | `SL-0014`, `SL-FU-0002`, `RTR-0007`, `RTR-0008`, `RTR-0012`, `AUTH-0012` | S6 distinguishes cleanup paths from cleanup guarantees. |
| Platform/tool profile and missing-tool behavior | `SL-0013`, `SL-FU-0003`, `AUTH-0008`, `PRES-0015` | Supported platform remains an owner decision. |
| Environment override precedence | `SL-0015`, `SL-FU-0004`, `AUTH-0006`, `PRES-0014` | S6 documents exact observed reads/defaults but does not infer cross-layer precedence. |
| Retained artifact freshness and failure-only bundles | `SL-0004`, `SL-0009`, `SL-0010`, `SL-FU-0005`, `RTR-0016`, `AUTH-0013` | S7 can preserve explicit run-selection and provenance limits. |
| Package scripts, local-dev services, external Go caches, visual snapshot refresh, and CI provider workflow gaps | `AUTH-0003`, `AUTH-0005`, `AUTH-0009`, `AUTH-0014`, `AUTH-0015`; `PRES-*`; `MSC-*` | Each has an owner question and is not silently promoted to a first-class contract. |

## Blocking Issues for S7

No S7 blocker was found.

The audit specifically checked for the requested blocker classes and found none:

- No S3/S4 shared mutable resource lacks S6 hazard coverage or explicit
  source-limit/exclusion routing.
- No fixed sleep, timeout, retry, poll, wait, readiness check, debounce/watch
  trigger, cleanup wait, signal wait, lock wait, or sharding assumption was
  found undocumented.
- Scheduler lanes are not described as concrete capacity guarantees.
- Runtime-sensitive behavior is not promoted without runtime evidence.
- No major subsystem lacks preservation classification.
- Authority-required behavior has concrete owner questions.
- Product/Core-spec-adjacent behavior is routed through `main-spec-conflict-list.md`
  or `harness-authority-map.md`.

## Follow-Up Work

These are valid follow-ups, but they should not block S7 if S7 preserves the
source-limit and authority-required classifications:

- Collect controlled runtime evidence for live service readiness, browser
  readiness, Docker/Compose behavior, reset-route behavior, timeout/interrupt
  cleanup, detached reaper completion, active DB cleanup, stale janitors, and
  port release.
- Decide first-class contract status for direct package scripts and local-dev
  Compose behavior.
- Decide public environment-variable contracts and cross-layer precedence.
- Decide supported platform/tool profile for Docker, Compose, `ss`, `curl`,
  `setsid`, `realpath`, shell, localhost networking, Node/pnpm, and browsers.
- Decide cleanup-scope treatment for external Go caches.
- Decide visual snapshot platform/browser/update authority.
- Supply CI provider workflow source or keep provider workflow behavior out of
  the harness contract.

## Commands Run

Only static inspection commands were run for this audit implementation:

- `git status --short --branch`
- `git rev-parse HEAD`
- `date --iso-8601=seconds`
- `uname -a`
- `test -f` file-presence checks
- `rg` searches and row-count checks
- `sed` reads of recovery docs, Core 00, and `docs/domain.md`

No service-backed, browser, Docker, Compose, reset, cleanup, formatter,
generator, baseline-refresh, snapshot-update, package-test, or broad
verification command was run.

## No-Change Audit

| Check | Result |
|---|---|
| Harness implementation files modified by this audit | no |
| Test logic modified by this audit | no |
| CI behavior modified by this audit | no |
| Fixture contents modified by this audit | no |
| Cleanup scripts modified by this audit | no |
| Timing, retry, sharding, scheduler, resource allocation, or reset behavior modified by this audit | no |
| Generated outputs or lockfiles modified by this audit | no |
| Audit output limited to recovery documentation | yes |

The worktree was already dirty with S6 recovery-document modifications before
this audit implementation began. This audit adds a documentation-only audit
report and does not alter those pre-existing recovery-doc edits.

## Validation Criteria

| Criterion | Result |
|---|---|
| Every required S6 output is present and internally consistent. | pass |
| Every shared mutable resource and timing surface is covered or explicitly source-limited. | pass |
| Failure and partial-completion links are evidence-backed. | pass |
| Preservation and authority classifications are complete. | pass |
| Unsupported assumptions are not promoted into S7-ready facts. | pass |
| Final verdict clearly states whether S7 may proceed and what limits S7 must preserve. | pass |

## Final Audit Checklist

- [x] Required S6 artifacts reviewed.
- [x] S3/S4/S5 inputs cross-checked.
- [x] Shared mutable resource coverage complete.
- [x] Sleeps, waits, retries, timeouts, polls, readiness checks, debounce/watch
      triggers, cleanup waits, signal waits, locks, and sharding assumptions
      covered.
- [x] Recurring failures and partial-completion states linked only where
      evidence supports them.
- [x] Every major subsystem has a preservation classification.
- [x] Classifications are consistent and not overclaimed.
- [x] Authority-required behavior has owner questions.
- [x] Main-spec conflicts are routed, not resolved by inference.
- [x] Unsupported assumptions, missing evidence, and source limits are recorded.
- [x] S7 blockers and follow-ups are separated.
- [x] No implementation or harness behavior changed.
- [x] Final verdict and S7 readiness statement recorded.
