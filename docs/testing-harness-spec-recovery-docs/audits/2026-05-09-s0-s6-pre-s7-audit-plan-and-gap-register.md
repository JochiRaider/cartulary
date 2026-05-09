---
doc_id: THR-S0-S6-PRE-S7-AUDIT-PLAN-AND-GAP-REGISTER-2026-05-09
title: S0-S6 Pre-S7 Audit Plan and Gap Register
status: complete
role: s7-readiness-audit
---

# S0-S6 Pre-S7 Audit Plan and Gap Register

## Document role

This document implements the S0 through S6 pre-S7 audit plan and gap register.
It verifies that the completed recovery sprints are safe S7 inputs only if S7
preserves every `source_limit`, `TODO:*_unknown`, and
`maintainer_decision_required` item.

This document does not draft the NLSpec, acceptance matrix, roadmap, or review
packet.

## Evidence base

The audit uses these source-of-truth inputs:

- `docs/testing-harness-spec-recovery-docs/03-sprint-plan.md`
- S0 through S6 outputs under `docs/testing-harness-spec-recovery-docs/`
- S0 through S6 handoffs under `docs/testing-harness-spec-recovery-docs/handoffs/`
- S0 through S6 audit documents under `docs/testing-harness-spec-recovery-docs/audits/`
- `docs/testing-harness-spec-recovery-docs/source-limit-log.md`
- `docs/testing-harness-spec-recovery-docs/ambiguity-register.md`
- `docs/testing-harness-spec-recovery-docs/s0-s6-gap-closure-plan.md`
- `docs/testing-harness-spec-recovery-docs/s7-s6-audit-gap-follow-up.md`
- `docs/testing-harness-spec-recovery-docs/harness-authority-map.md`
- `docs/testing-harness-spec-recovery-docs/preservation-matrix.md`
- `docs/testing-harness-spec-recovery-docs/main-spec-conflict-list.md`
- Remaining `TODO:` markers in completed S0 through S6 outputs.

Current scope check before this document was added: `git status --short --branch`
printed only `## main...origin/main`, and both unstaged and staged diff name
checks were empty.

## Audit verdict

S0 through S6 are documentation-complete enough to support S7. No standalone
documentation blocker exists for starting S7 if S7 preserves open source
limits, explicit unknowns, and maintainer-decision-required authority
questions.

The only S7-start blocker is failure to preserve those labels and questions.
The gaps in this register block final normative S7 `MUST` language where noted,
but they do not block beginning S7 as a source-limited drafting pass.

## Sprint audit plan

| Sprint | Audit result | Remaining pre-S7 handling |
|---|---|---|
| S0 | Charter, write boundary, source-limit log, handoff, and audit pass. | Preserve absent `.github/**` as `SL-0001`; do not invent CI provider behavior. |
| S1 | Inventory, uninvoked surfaces, embedded logic, ambiguity/source-limit rows, handoff, and audit pass with follow-ups. | Preserve planned phase7/phase8 absence, reset-route authority, retained-artifact provenance, generated-output authority, and local-dev boundary questions. |
| S2 | Entrypoint map and sequencing list reconcile Make targets, package scripts, CI-adjacent scripts, and audits. | Treat Make as canonical; keep package scripts, broad runtime behavior, env precedence, and uninvoked script behavior unresolved unless owner decides. |
| S3 | Artifact matrix, cleanup matrix, shared-state hazards, handoff, and audit pass. S3 future-reference follow-up is complete. | Preserve fixture/snapshot update authority, package-script cleanup gaps, retained-artifact freshness, failure bundles, and live cleanup limits. |
| S4 | Service lifecycle, env observations, resources, hazards, handoff, and audit pass. | Preserve live readiness, platform/tool, timeout/interrupt, detached reaper, stale janitor, and env precedence limits. |
| S5 | Observable interfaces, schemas, lifecycle, transitions, partial states, failure taxonomy/register, handoff, and audit pass. S5 link blocker is resolved. | Preserve `schema_unknown`, controlled-failure absence, Playwright bundle gaps, CI annotation gaps, and retained-run provenance. |
| S6 | Race/timing/resource, concurrency, timeout/retry, preservation, authority, conflict routing, handoff, and audit pass. | Use `s7-s6-audit-gap-follow-up.md`; scheduler lanes remain logical limits, not concrete capacity guarantees. |

## Per-sprint verification matrix

| Sprint | Expected outputs | Validation and exit criteria | Handoff usability | Source limits and ambiguity preservation | Audit follow-ups and no-change scope |
|---|---|---|---|---|---|
| S0 | `recovery-charter.md`, `source-limit-log.md`, charter boundary list, sprint-plan update, S0 handoff, S0 audit. | Charter records permitted writes, prohibited edits, evidence labels, repository state, and provisional boundary; S1 can begin without transcript memory. | S0 handoff names inspected surfaces, source limits, pending CI question, and implementation-change audit. | `.github/**` absence preserved as `SL-0001` and `AMB-0001`. | S0 audit passes; no implementation, test, fixture, CI, cleanup, generated, or lockfile edits. |
| S1 | `harness-inventory.md`, `uninvoked-surface-list.md`, `embedded-harness-logic-list.md`, `ambiguity-register.md`, source-limit updates, S1 handoff, S1 audit. | Inventory covers discovered harness surfaces, separates uninvoked and embedded logic, and records generated/temp ownership hypotheses or explicit unknowns. | S1 handoff gives S2 entrypoint seeds and tells later sprints not to merge product assertions into harness mechanics. | `SL-0001..SL-0005` and `AMB-0001..AMB-0010` stay open or routed. | S1 audit passes with follow-ups routed to S2/S3/S4/S5/S6/S7/S8; no prohibited file classes changed. |
| S2 | `entrypoint-command-map.md`, `sequencing-assumption-list.md`, inventory/register updates, S2 handoff, S2 audit. | Current Make surface reconciles to S2 target families; package scripts, mutating commands, cleanup commands, CI-adjacent scripts, and broad runtime gaps are classified. | S2 handoff gives stable `EP-*` and `SEQ-*` IDs for S3-S6 consumers. | Runtime behavior, provider CI, env precedence, package scripts, and uninvoked script status remain source-limited or owner-required. | S2 audit passes with follow-ups routed; no broad gates, service targets, browser targets, cleanup, generation, formatting, baseline refresh, or failure scenarios executed. |
| S3 | `artifact-ownership-matrix.md`, `cleanup-lifecycle-matrix.md`, `shared-state-hazard-list.md`, register updates, S3 handoff, S3 audit. | Artifacts, cleanup surfaces, external state, canonical fixtures, derived reports, reused/failure-only artifacts, and package-script artifacts are represented with owners or explicit unknowns. | S3 handoff gives S4 concrete external-state and cleanup seed rows. | Fixture/snapshot update authority, retained artifact freshness, timeout/interrupt cleanup, stale janitor bounds, package-script cleanup, and external cache scope stay open. | S3 audit follow-up for future S4 hazard placeholders is complete; other follow-ups are routed to S4/S5/S6/S8; no fixture, snapshot, cleanup, generated, lockfile, or service files changed. |
| S4 | `service-lifecycle-map.md`, `environment-contract-observations.md`, `resource-allocation-register.md`, `HAZ-S4-*` rows, register updates, S4 handoff, S4 audit. | Every service has readiness or explicit unknown; resources have allocation/conflict/release rows; env assumptions are entrypoint-linked; logical lanes are separated from concrete capacity. | S4 handoff gives S5 interface/failure anchors, S6 resource/timing seeds, and S8 authority questions. | `SL-0012..SL-0015`, `AMB-0023..AMB-0028`, reset route authority, platform/tool gaps, and env precedence remain preserved. | S4 audit passes with follow-ups routed; no service-backed/browser/Docker/Compose/reset/cleanup/runtime/generator/formatter commands run. |
| S5 | Observable interface, structured schema, consumer, lifecycle, transition, partial-state, failure register, taxonomy, S5 handoff, S5 audits. | Machine-consumed outputs have `SCHEMA-*` or `TODO: schema_unknown`; major entrypoints have lifecycle/terminal states; failure classes and modes have retryability. | S5 handoff gives S6 hazard seeds, S7 interface/schema caution, and S8 authority routing. | Controlled failure examples, Playwright bundles, CI annotations, retained provenance, cleanup strength, and runtime readiness remain source-limited. | S5 semantic link blocker is resolved; no S7 drafting or runtime/failure/generator/cleanup commands run. |
| S6 | `race-timing-resource-register.md`, `concurrency-model-notes.md`, `timeout-retry-register.md`, updated hazards, preservation, authority, conflict, ambiguity/source-limit docs, S6 handoff, S6 audits. | Shared resources, timing/wait/retry/lock/signal surfaces, preservation classifications, authority prompts, and main-spec-sensitive risks are covered or explicitly source-limited. | S6 handoff and `s7-s6-audit-gap-follow-up.md` provide mandatory S7 carry-forward inputs. | Runtime readiness, cleanup/signal behavior, retained provenance, env/platform, package scripts, reset route, local-dev, visual snapshots, caches, and CI remain preserved. | S6 audits pass with source limits preserved; no implementation, fixture, cleanup, CI, generated, lockfile, runtime, package-manager, or snapshot behavior changed. |

## Gap register

| Priority | Sprint | Gap / issue | Evidence source | Required follow-up | Blocks S7? | Validation / exit criteria |
|---|---|---|---|---|---|---|
| P0 | S0-S6 | S7 may proceed only if source limits and owner questions are preserved. | `03-sprint-plan.md`; `s0-s6-gap-closure-plan.md`; `s7-s6-audit-gap-follow-up.md` | Build S7 carry-forward checklist before drafting. | Yes, if not preserved. | Every S7 claim cites row IDs and keeps evidence labels. |
| P1 | S0/S1/S2/S5/S6 | Provider CI workflows and annotations absent. | `SL-0001`, `AMB-0001`, `SCHEMA-0019`, `AUTH-0015`, `PRES-0019` | Draft only repo-local `make ci`/`scripts/ci/**`; route provider CI to owner. | No; blocks CI-provider normativity. | No `.github/**` behavior or annotation schema is invented. |
| P1 | S1/S2 | Planned phase7/phase8 files absent. | `SL-0005`, `AMB-0004` | Treat as planned, not active coverage. | No. | S7 does not count phase7/phase8 as missing active harness evidence. |
| P1 | S2/S4/S5/S6 | Broad gates, service-backed, browser, Docker/Compose, reset, and controlled failures were not run. | `SL-0002`, `SL-0003`, `SL-0006`, `SL-0012`, `AMB-0014`, `AMB-0023` | Keep live behavior `source_limit`; require later authorized run for runtime guarantees. | No; blocks live-readiness `MUST` claims. | Static evidence is not upgraded to `runtime_observed`. |
| P1 | S3/S4/S6 | Cleanup on timeout/interrupt/parent death, detached reaper completion, stale janitors, active DB cleanup, and port release unproven. | `SL-0011`, `SL-0014`, `AMB-0018`, `AMB-0019`, `AMB-0024`, `AMB-0027`, `AUTH-0007`, `AUTH-0012` | Preserve as best-effort/source-limited unless owner or controlled evidence closes it. | No; blocks guaranteed cleanup language. | S7 distinguishes cleanup path, scheduling, best effort, and verified completion. |
| P1 | S1/S4/S6 | Reset route authority, readiness, timeout, and destructive side effects remain owner-required. | `AMB-0006`, `SVC-0014`, `TMR-0016`, `AUTH-0004`, `PRES-0010`, `MSC-0001` | Route to maintainer plus product/security owner. | No; blocks public/reset-route contract. | S7 does not present reset route as product API. |
| P1 | S2/S4/S6 | Public env vars and cross-layer precedence unresolved. | `SL-0015`, `AMB-0012`, `AMB-0026`, `AUTH-0006`, `PRES-0014` | Record observed reads/defaults only; route precedence to owner. | No; blocks env precedence normativity. | Every env rule is either source-observed or owner-required. |
| P1 | S4/S6 | Supported platform/tool profile unresolved. | `SL-0013`, `AMB-0025`, `AMB-0028`, `AUTH-0008`, `PRES-0015` | Keep `platform_unknown`; owner must define Linux/WSL, Docker, Compose, `ss`, `curl`, `setsid`, Node/pnpm, browser support. | No; blocks portability guarantees. | S7 names only source-proved checks/failures. |
| P1 | S2/S3/S4/S6 | Direct package scripts bypass Make result-root, scheduler, cleanup, env, and output guarantees. | `AMB-0011`, `AMB-0020`, `ART-0030`, `CLN-0020`, `AUTH-0003`, `PRES-0011` | Default Make canonical; route package-script status to owner. | No; blocks package-script contract. | Package scripts are separate unless explicitly adopted. |
| P1 | S3/S5/S6 | Retained artifact freshness, newest-run fallback, and failure-only bundles lack selected-run provenance. | `SL-0004`, `SL-0009`, `SL-0010`, `AMB-0017`, `AUTH-0013`, `PRES-0017` | Require explicit `RESULTS_DIR`, `RUN_ID`, command, platform, exit status, and artifact paths for durable claims. | No; blocks artifact-based guarantees. | No ambient newest-run evidence supports normative claims. |
| P2 | S3/S5/S6 | Machine schemas remain unknown for shell log contents, Playwright reports/traces/screenshots/videos, CI annotations, and failure-only artifacts. | `SCHEMA-0002`, `SCHEMA-0013`, `SCHEMA-0019`, `OI-0008`, `OI-0014`, `TODO: schema_unknown` | Keep `schema_unknown`; inspect selected artifacts only if authorized. | No; blocks stable schema requirements. | Every machine-consumed unknown stays explicit. |
| P2 | S3/S6 | Fixture, golden, and visual snapshot update authority unresolved. | `AMB-0015`, `AMB-0022`, `ART-0001..ART-0003`, `AUTH-0014`, `PRES-0018` | Default owner-reviewed source edits; no snapshot refresh authority. | No; blocks mutation/update rules. | S7 does not define refresh workflow without owner decision. |
| P2 | S4/S6/S8 | Local-dev Compose, `make dev`, fixed ports, persistent volumes, and MinIO reset gap are authority-required. | `AUTH-0005`, `PRES-0012`, `MSC-0003` | Keep local-dev guidance separate from verification contract unless adopted. | No. | Local-dev behavior is not used as verification conformance. |
| P2 | S3/S6 | External Go cache cleanup scope unresolved. | `AMB-0021`, `AUTH-0009`, `PRES-0016` | Default tool-managed outside cleanup unless owner expands scope. | No. | S7 does not add `/tmp/cartulary-go-*` cleanup requirement. |
| P2 | S3/S6 | Generated artifacts can drive execution but are not behavior owners. | `AMB-0003`, `AMB-0016`, `AUTH-0010`, `PRES-0002`, `MSC-0005` | Preserve upstream owner/downstream generated distinction. | No. | Generated drift checks are freshness checks, not normative ownership. |
| P2 | S4/S6 | Scheduler lanes are logical scheduling constraints, not concrete host/Docker/Postgres/MinIO/browser capacity guarantees. | `HAZ-S4-0001`, `AUTH-0011`, `PRES-0003`, S6 audit | Draft lanes separately from concrete capacity. | No; blocks capacity guarantees. | No S7 language equates lanes with service capacity. |
| P3 | S7 | S7 outputs are intentionally absent. | `03-sprint-plan.md` S7 row; `TODO: harness_nlspec_draft_path`, `TODO: acceptance_matrix`, roadmap/review placeholders | Do not draft them in this audit step. | No; this is S7 scope. | S7 remains `not_started` after this audit plan. |

## Consolidated cross-sprint gaps

- Runtime evidence remains absent for broad gates, service-backed targets,
  browser targets, Docker/testcontainers, Compose, reset routes, cleanup paths,
  timeout/interrupt/parent-death scenarios, baseline refreshes, snapshot
  updates, and controlled failures.
- Cleanup strength remains unresolved for detached reapers, stale janitors,
  active DB cleanup, port release, process groups, shell traps, and abrupt
  exits.
- Retained artifact provenance remains weak unless a selected `RESULTS_DIR`,
  `RUN_ID`, command, platform/tool profile, exit status, and artifact path set
  are recorded.
- Machine schemas remain incomplete for shell log contents, Playwright
  reports/traces/screenshots/videos, CI provider annotations, and selected
  failure-only artifacts.
- Authority remains open for reset route, direct package scripts, local-dev
  services, stale janitors, environment contracts, platform/tool profile,
  external Go caches, visual snapshot updates, retained artifact selection, and
  CI provider behavior.
- Generated artifacts can drive execution as downstream inputs, but they are
  not behavior owners.

## Unresolved source limits

Carry forward all current source limits:

- `SL-0001` through `SL-0015`
- `SL-FU-0001` through `SL-FU-0007`

Highest-impact groups for S7 are:

- CI absence and provider annotation behavior.
- Unexecuted broad/runtime/mutating commands.
- Retained artifact provenance and failure-only bundles.
- Live service/browser readiness.
- Cleanup, timeout, interrupt, parent-death, and signal behavior.
- Platform/tool support.
- Environment precedence.

## Maintainer decision questions

S7 must route these questions instead of resolving them by inference:

- Should Make remain the sole canonical harness command surface, with direct
  package scripts treated as developer conveniences unless adopted?
- Is `internal/app/test_runtime_reset.go` harness-owned, and what
  visibility/security boundary applies?
- Which local-dev Compose and `make dev` behaviors belong in the harness
  contract?
- What proof is sufficient before stale janitors delete DBs, buckets, or
  containers?
- Which environment variables are public harness API, and what precedence
  applies across Make, schedulers, shell wrappers, package scripts, Go helpers,
  config files, and Playwright?
- What platform/tool profile is supported for Linux/WSL, Docker, Compose, `ss`,
  `curl`, `setsid`, `realpath`, shell, localhost networking, Node/pnpm, and
  browsers?
- Are `/tmp/cartulary-go-*` caches inside cleanup scope?
- Which OS, browser version, and command may refresh visual snapshots?
- Which retained-artifact tools may use newest-run fallback, and when must
  `RESULTS_DIR` or active run identity be explicit?
- Is CI provider behavior external, absent, or represented only by
  provider-neutral scripts?

## S7 blockers

No standalone documentation gap blocks starting S7.

S7 is blocked only if it fails to preserve source limits, explicit unknowns,
and maintainer-decision-required questions.

The following block final normative S7 `MUST` language until evidence or owner
decisions exist:

- Live readiness, cleanup completion, signal or timeout behavior, reaper
  completion, active DB cleanup, stale janitor deletion, or port release
  guarantees.
- Public contracts for reset route, env precedence, direct package scripts,
  local-dev Compose, visual snapshot updates, external Go cache cleanup, or CI
  provider annotations.
- Any claim that scheduler lanes are concrete host, Docker, Postgres, MinIO, or
  browser capacity guarantees.

## Gaps preservable as source limits during S7

S7 may preserve these gaps as source-limited or owner-required instead of
closing them:

- CI provider source and annotation behavior.
- Planned phase7/phase8 files.
- Broad gate, service-backed, browser, Docker/Compose, reset, and controlled
  failure runtime behavior.
- Cleanup timing and completion guarantees.
- Retained artifact freshness and newest-run fallback.
- Failure-only bundle schemas.
- Playwright report/trace/screenshot/video schemas.
- Environment precedence.
- Platform/tool support profile.
- Package-script authority and cleanup behavior.
- Fixture, golden, and visual snapshot update authority.
- Local-dev Compose and `make dev` authority.
- External Go cache cleanup scope.

## Minimal execution order

1. Freeze S0 through S6 as documentation-complete and do not run runtime,
   cleanup, generator, formatter, snapshot, package-test, Docker, Compose,
   reset, or broad verification commands.
2. Build the S7 carry-forward checklist from `SL-*`, `AMB-*`, `AUTH-*`,
   `PRES-*`, `MSC-*`, and remaining `TODO:*_unknown` markers.
3. Classify each S7 candidate claim as `observed`, `source_limit`,
   `maintainer_decision_required`, or `deferred_roadmap`.
4. Route owner-required items before any final normative language.
5. If a strong guarantee is needed, stop that section until authorized runtime
   evidence or maintainer decision exists.

## Final readiness criteria for beginning S7

S7 may begin when all of the following are true:

- Every S7 draft input preserves evidence labels from S0 through S6.
- Every unresolved `TODO:*_unknown` remains explicit or is routed to a source
  limit or owner decision.
- Every S7 requirement cites existing row IDs or a later authorized evidence
  record.
- Runtime-sensitive claims remain source-limited unless actual runtime evidence
  exists.
- Owner-sensitive claims remain `maintainer_decision_required` until explicitly
  decided.
- No implementation, fixture, cleanup, CI, generated-artifact, lockfile,
  runtime, package-manager, or snapshot change is made during the transition.

## No-change audit

This audit plan and gap register is documentation-only. It does not change
harness implementation behavior, product implementation behavior, test logic,
fixture or snapshot contents, CI behavior, cleanup scripts, generated outputs,
lockfiles, package-manager state, runtime services, or S7 deliverables.
