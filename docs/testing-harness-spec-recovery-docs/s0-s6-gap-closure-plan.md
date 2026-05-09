---
doc_id: THR-S0-S6-GAP-CLOSURE-PLAN-2026-05-09
title: S0-S6 Gap Closure Plan Before S7
status: complete
role: s7-readiness-closure-plan
---

# S0-S6 Gap Closure Plan Before S7

## Document role

This document is the S7 preflight closure plan for S0 through S6. It confirms
that the completed recovery sprints are usable as S7 inputs only if S7 preserves
their source limits, open owner decisions, and evidence labels.

This plan does not rewrite the harness, modify implementation behavior, close
runtime evidence gaps, settle maintainer decisions, or begin the S7 NLSpec
draft. It consolidates the remaining S0 through S6 follow-up work so the handoff
to S7 is explicit and reviewable.

Primary controlling inputs:

- `03-sprint-plan.md`
- `source-limit-log.md`
- `ambiguity-register.md`
- `harness-authority-map.md`
- `preservation-matrix.md`
- `s7-s6-audit-gap-follow-up.md`
- S0 through S6 handoffs and audits under `handoffs/` and `audits/`

## Closure verdict

S0 through S6 are documentation-complete enough to support S7. No hard blocker
exists for starting S7 if S7 keeps every runtime-sensitive item labeled
`source_limit` and every owner-sensitive item labeled
`maintainer_decision_required`.

The remaining work is evidence and authority closure, not harness
implementation closure. Runtime, cleanup, generated-output, fixture,
package-manager, lockfile, snapshot, Docker, Compose, service-backed, browser,
reset, formatter, generator, baseline-refresh, and broad verification commands
remain out of scope for this closure step.

## Sprint-by-sprint closure

| Sprint | Confirmed complete | Remaining gaps before safe S7 | Required pre-S7 closure | Deferrable beyond S7 |
|---|---|---|---|---|
| S0 Charter/setup | Charter, write boundary, evidence labels, source-limit log, handoff, and audit pass. | `.github/**` absent as `SL-0001`; CI provider behavior unknown. | Carry CI provider behavior as source-limited; draft only repo-local `make ci` and `scripts/ci/**`. | Owner decision on external, absent, or provider CI. |
| S1 Inventory/boundary | Inventory, uninvoked surfaces, embedded harness logic, ambiguity register, and audit pass with follow-ups. | Planned phase7/phase8 files absent; reset-route owner boundary, generated-output authority, retained artifacts, cleanup/service lifecycle, and local-dev boundary remain open. | Treat phase7/phase8 as planned, not active; preserve open `AMB-*` rows instead of resolving by inference. | Maintainer decisions on planned phases, embedded harness scope, and reset-route inclusion. |
| S2 Entrypoints/commands | Command map reconciles 122 Make targets, package scripts, CI-adjacent scripts, sequencing assumptions, and audit pass. | Broad gates, runtime failures, cleanup, generators/formatters, package scripts, env precedence, and uninvoked script behavior remain source-limited. | Draft Make/task-surface as canonical; keep package scripts separate unless adopted; make no broad runtime success/failure guarantee. | Runtime verification pass; owner decision on package scripts and uninvoked `scripts/test-run-go-target-fast.sh`. |
| S3 Fixtures/artifacts/cleanup | Artifact matrix, cleanup matrix, shared-state hazards, and S3 audit follow-ups recorded; future hazard placeholder fixed. | `TODO: update_rule_unknown` for fixture/golden/snapshot authority; `TODO: owner_unknown` and `TODO: cleanup_rule_unknown` for package-script artifacts; retained artifact freshness and failure bundles source-limited. | S7 must state fixture/golden/snapshot mutation is owner-reviewed or unknown; durable retained-artifact claims require explicit `RESULTS_DIR` and run ID. | Snapshot refresh policy, package-script cleanup policy, failure-bundle schema evidence, stale artifact selection rules. |
| S4 Services/env/resources | Service, env, and resource rows complete; all services have readiness or explicit unknown; audit pass with follow-ups. | Live Docker/testcontainers/browser/Compose/reset behavior unobserved; `TODO: readiness_unknown`, `TODO: platform_unknown`, `TODO: precedence_unknown`, and timeout unknowns remain. | S7 may cite S4 rows as source-observed only; do not claim live readiness, cleanup completion, env precedence, or platform support. | Authorized runtime readiness pass; env override matrix; supported platform/tool matrix. |
| S5 Lifecycle/interfaces/failures | Observable interfaces, schema notes, consumers, lifecycle, transitions, partial states, failure taxonomy/register complete; audit follow-up resolved. | Controlled failures, Playwright failure bundles, CI annotations, retained artifact provenance, and runtime cleanup/readiness examples remain unavailable; `TODO: schema_unknown` remains intentional. | Do not promote `schema_unknown`, `authority_unknown`, or `source_limit` rows into stable machine schemas or guarantees. | Controlled failure runs; selected failure artifacts; provider CI source. |
| S6 Hazards/resources/timing | `RTR-*`, `TMR-*`, `CONC-*`, `PRES-*`, `AUTH-*`, and `MSC-*` rows complete; audit verdict is pass with source limits preserved; S7 follow-up track exists. | Runtime readiness, cleanup/signal behavior, stale janitor proof, env/platform, retained artifact freshness, reset route, package scripts, local-dev, visual snapshots, and caches remain source-limited or owner-required. | Use `s7-s6-audit-gap-follow-up.md` as mandatory carry-forward; scheduler lanes must stay logical, not concrete capacity guarantees. | Runtime readiness, cleanup/signal, artifact provenance, and owner-decision passes. |

## Cross-sprint gaps

- Runtime evidence is intentionally missing for broad gates, service-backed
  targets, browser E2E, Docker/testcontainers, Compose, reset routes, cleanup
  paths, timeout/interrupt/parent-death scenarios, baseline refreshes, snapshot
  updates, and controlled failures.
- Retained artifact evidence is weak unless a specific `RESULTS_DIR`, `RUN_ID`,
  command, platform profile, exit status, and artifact path set is recorded.
- Machine schemas remain incomplete for shell log contents, Playwright reports,
  traces, screenshots, videos, CI provider annotations, and uninvoked or
  source-limited surfaces.
- Authority remains open for reset route, direct package scripts, local-dev
  services, stale janitors, env contracts, supported platform, external Go
  caches, visual snapshot updates, retained artifact selection, and CI provider
  behavior.
- Template and process `TODO` placeholders for S7 outputs are expected
  scaffolding. S0 through S6 closure focuses on live `TODO` markers inside
  completed outputs.

## Source limits affecting S7 quality

Carry these source limits into S7 unchanged unless separately authorized
evidence exists:

| Source limits | S7 effect |
|---|---|
| `SL-0001` | Provider CI source and annotation behavior are absent while `.github/**` is absent. |
| `SL-0004`, `SL-0009`, `SL-0010` | Retained artifacts, selected-run provenance, and failure-only bundles cannot support durable claims without explicit run identity. |
| `SL-0008` | Mutating writer, cleanup, generation, format, baseline, and snapshot commands were not executed. |
| `SL-0011` through `SL-0014` | Live external-state cleanup, service readiness, platform behavior, timeout, interrupt, and reaper cleanup remain unobserved. |
| `SL-0015` | Environment override precedence remains unresolved. |
| `SL-FU-0001` through `SL-FU-0007` | S5/S6/S7 routing notes preserve the source limits and do not close them. |

## Authority questions requiring maintainer input

Before final normative S7 language, maintainers must still answer:

- What environment-variable precedence applies across Make, scripts,
  schedulers, package scripts, Go helpers, config files, and Playwright?
- Which exact OS, browser, version, and command may refresh visual snapshots?
- What evidence or owner decision proves parent-death cleanup, active DB
  cleanup, or detached reaper completion as stronger than best-effort?
- What provider-specific CI workflow, annotations, uploads, and dashboard
  behavior, if any, should be part of the repo contract?
- Which Playwright report, trace, video, and screenshot internals, if any,
  should be adopted as stable harness schemas?

The 2026-05-09 maintainer decisions resolved these previously open authority
questions for S7: Make is canonical, direct package scripts are developer
conveniences, the former `internal/app/test_runtime_reset.go` behavior is
harness-owned and now lives under `internal/testutil/testruntime`, local-dev
Compose and `make dev` are local verification behavior, stale janitors require
generated names plus metadata or lease evidence and conservative checks, WSL/Linux
is the primary observed environment without a full platform matrix, and
`/tmp/cartulary-go-*` caches are outside default cleanup.
- Which OS/browser/update command owns visual snapshot refresh?
- Which retained-artifact tools may use newest-run fallback?
- Is CI provider behavior external, absent, or represented only by
  provider-neutral scripts?

## Minimal execution order

1. Freeze S0 through S6 as documentation-complete and do not rerun
   implementation, cleanup, generation, formatting, runtime, browser, Docker,
   or snapshot commands.
2. Build the S7 carry-forward checklist from `SL-0001`, `SL-0004`,
   `SL-0008..SL-0015`, `SL-FU-*`, open `AMB-*`, `AUTH-*`, and `PRES-*` rows.
3. Classify every S7 candidate requirement as exactly one of:
   `source_observed`, `source_limited`, `maintainer_decision_required`, or
   `deferred_roadmap`.
4. Resolve only wording-level closure before S7. Static evidence must not
   become `runtime_observed`, and owner-required behavior must not become
   normative.
5. Defer runtime readiness, cleanup/signal, failure-bundle,
   artifact-provenance, env/platform, and owner-decision passes unless S7 needs
   final normative guarantees.

## S7 readiness criteria

S0 through S6 are complete enough to support S7 when all of the following are
true:

- All S0 through S6 expected outputs exist and their sprint status remains
  `complete`.
- Every remaining `TODO` in S0 through S6 outputs is either a named unknown
  such as `schema_unknown`, `readiness_unknown`, `precedence_unknown`,
  `platform_unknown`, or `update_rule_unknown`, or is explicitly routed to an
  owner/source-limit row.
- S7 cites existing row IDs for every requirement and preserves evidence labels.
- No final normative language is drafted for live readiness, cleanup
  completion, signal/timeout behavior, reset-route public status, env
  precedence, direct package scripts, local-dev Compose, visual snapshot
  updates, external cache cleanup, CI provider annotations, or concrete
  scheduler capacity without new evidence or owner decision.
- No implementation, generated, lockfile, fixture, cleanup, service, runtime,
  package-manager, or snapshot behavior is changed during this closure step.

## Documentation-only implementation checklist

- [x] S0 through S6 completion evidence summarized.
- [x] Incomplete or partial work separated from blockers.
- [x] Remaining named `TODO` markers consolidated.
- [x] Open ambiguity and source-limit items preserved.
- [x] Missing or weak evidence grouped by runtime, artifact, schema, and
      authority area.
- [x] Audit findings requiring action routed to S7 or later evidence/owner
      passes.
- [x] Cross-sprint gaps consolidated.
- [x] Source limits affecting specification quality listed.
- [x] Maintainer authority questions listed.
- [x] Minimal execution order recorded.
- [x] S7 readiness criteria recorded.
- [x] No S7 NLSpec draft started.
- [x] No implementation behavior changed.
