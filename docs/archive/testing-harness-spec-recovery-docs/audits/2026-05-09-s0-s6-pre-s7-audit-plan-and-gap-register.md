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
packet. The existing S7 runtime evidence pass is treated as authorized evidence
only; it does not start or satisfy the S7 NLSpec, acceptance, roadmap, or review
deliverables.

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
- `docs/testing-harness-spec-recovery-docs/runtime-evidence-register.md`
- `docs/testing-harness-spec-recovery-docs/cleanup-signal-evidence-register.md`
- `docs/testing-harness-spec-recovery-docs/harness-authority-map.md`
- `docs/testing-harness-spec-recovery-docs/preservation-matrix.md`
- `docs/testing-harness-spec-recovery-docs/main-spec-conflict-list.md`
- Remaining `TODO:` markers in completed S0 through S6 outputs.

Current source-control check before this reconciliation: `git status --short
--branch` printed only `## main...origin/main`. The latest commit was
`f037aee S7 Testing-Harness Recovery`, and its changed paths were limited to
`docs/testing-harness-spec-recovery-docs/**`.

## Audit verdict

S0 through S6 are documentation-complete enough to support S7. S7 may begin
only when source limits, named TODOs, selected-evidence bounds, and
maintainer-decision-required authority questions are preserved.

The prior pre-S7 audit wording is reconciled by this document with
`SL-FU-0008` and the authorized runtime evidence registers. Passing selected
runs may support only current-host, selected-command claims with explicit run
IDs and result directories. They do not close provider CI behavior, failed
CI/release readiness, parent-death cleanup, live active-DB cleanup, stale
janitor authority, browser post-start cleanup, local-dev contract status,
platform matrix, env precedence, package-script authority, or visual snapshot
update authority.

## Sprint audit plan

| Sprint | Required audit checks | Current disposition |
|---|---|---|
| S0 | Verify charter, source-limit log, write boundary, S0 handoff/audit, `.github/**` absence routing. | Complete; preserve `SL-0001` and `AMB-0001`. |
| S1 | Verify inventory, uninvoked surfaces, embedded logic, ambiguity/source-limit rows, handoff/audit. | Complete; planned phase7/phase8, reset-route, generated-output, retained-artifact, cleanup, and local-dev questions stay routed. |
| S2 | Verify entrypoint map, sequencing assumptions, package scripts, CI-adjacent scripts, mutating command classification. | Complete; Make remains canonical by default; broad runtime and package-script behavior stay limited unless selected evidence or owner decision exists. |
| S3 | Verify artifact ownership, cleanup matrix, hazards, handoff/audit, S3 follow-ups. | Complete; fixture/golden/snapshot update authority, package-script cleanup, retained artifacts, and failure bundles remain open. |
| S4 | Verify service lifecycle, env observations, resources, hazards, handoff/audit. | Complete; live readiness, env precedence, platform/tool support, cleanup completion, and reset-route behavior stay limited. |
| S5 | Verify observable interfaces, schemas, lifecycle, transitions, partial states, failure taxonomy/register, handoff/audits. | Complete; S5 link blocker resolved; `schema_unknown`, CI annotations, Playwright bundles, controlled failures, and retained provenance stay explicit. |
| S6 | Verify timing/resource/concurrency/register outputs, preservation/authority/conflict routing, handoff/audits. | Complete; source limits and owner questions are preserved; scheduler lanes are not capacity guarantees. |

## Per-sprint verification matrix

| Sprint | Expected outputs | Validation and exit criteria | Handoff usability | Source limits and ambiguity preservation | Audit follow-ups and no-change scope |
|---|---|---|---|---|---|
| S0 | `recovery-charter.md`, `source-limit-log.md`, charter boundary list, sprint-plan update, S0 handoff, S0 audit. | Charter records permitted writes, prohibited edits, evidence labels, repository state, and provisional boundary; S1 can begin without transcript memory. | S0 handoff names inspected surfaces, source limits, pending CI question, and implementation-change audit. | `.github/**` absence preserved as `SL-0001` and `AMB-0001`. | S0 audit passes; no implementation, test, fixture, CI, cleanup, generated, or lockfile edits. |
| S1 | `harness-inventory.md`, `uninvoked-surface-list.md`, `embedded-harness-logic-list.md`, `ambiguity-register.md`, source-limit updates, S1 handoff, S1 audit. | Inventory covers discovered harness surfaces, separates uninvoked and embedded logic, and records generated/temp ownership hypotheses or explicit unknowns. | S1 handoff gives S2 entrypoint seeds and tells later sprints not to merge product assertions into harness mechanics. | `SL-0001..SL-0005` and `AMB-0001..AMB-0010` stay open or routed. | S1 audit passes with follow-ups routed to S2/S3/S4/S5/S6/S7/S8; no prohibited file classes changed. |
| S2 | `entrypoint-command-map.md`, `sequencing-assumption-list.md`, inventory/register updates, S2 handoff, S2 audit. | Current Make surface reconciles to target families; package scripts, mutating commands, cleanup commands, CI-adjacent scripts, and broad runtime gaps are classified. | S2 handoff gives stable `EP-*` and `SEQ-*` IDs for S3-S6 consumers. | Runtime behavior, provider CI, env precedence, package scripts, and uninvoked script status remain source-limited or owner-required unless selected evidence applies. | S2 audit passes with follow-ups routed; no broad gates, service targets, browser targets, cleanup, generation, formatting, baseline refresh, or failure scenarios were run during S2. |
| S3 | `artifact-ownership-matrix.md`, `cleanup-lifecycle-matrix.md`, `shared-state-hazard-list.md`, register updates, S3 handoff, S3 audit. | Artifacts, cleanup surfaces, external state, canonical fixtures, derived reports, reused/failure-only artifacts, and package-script artifacts are represented with owners or explicit unknowns. | S3 handoff gives S4 concrete external-state and cleanup seed rows. | Fixture/snapshot update authority, retained artifact freshness, timeout/interrupt cleanup, stale janitor bounds, package-script cleanup, and external cache scope stay open. | S3 audit follow-up for future S4 hazard placeholders is complete; other follow-ups are routed to S4/S5/S6/S8; no fixture, snapshot, cleanup, generated, lockfile, or service files changed. |
| S4 | `service-lifecycle-map.md`, `environment-contract-observations.md`, `resource-allocation-register.md`, `HAZ-S4-*` rows, register updates, S4 handoff, S4 audit. | Every service has readiness or explicit unknown; resources have allocation/conflict/release rows; env assumptions are entrypoint-linked; logical lanes are separated from concrete capacity. | S4 handoff gives S5 interface/failure anchors, S6 resource/timing seeds, and S8 authority questions. | `SL-0012..SL-0015`, `AMB-0023..AMB-0028`, reset route authority, platform/tool gaps, and env precedence remain preserved; selected S7 evidence narrows only explicit current-host run claims. | S4 audit passes with follow-ups routed; no service-backed/browser/Docker/Compose/reset/cleanup/runtime/generator/formatter commands were run during S4. |
| S5 | Observable interface, structured schema, consumer, lifecycle, transition, partial-state, failure register, taxonomy, S5 handoff, S5 audits. | Machine-consumed outputs have `SCHEMA-*` or `TODO: schema_unknown`; major entrypoints have lifecycle/terminal states; failure classes and modes have retryability. | S5 handoff gives S6 hazard seeds, S7 interface/schema caution, and S8 authority routing. | Controlled failure examples, Playwright bundles, CI annotations, retained provenance, cleanup strength, and runtime readiness remain source-limited unless selected S7 evidence is cited. | S5 semantic link blocker is resolved; S7 controlled-failure evidence is evidence-only and does not draft S7 deliverables. |
| S6 | `race-timing-resource-register.md`, `concurrency-model-notes.md`, `timeout-retry-register.md`, updated hazards, preservation, authority, conflict, ambiguity/source-limit docs, S6 handoff, S6 audits. | Shared resources, timing/wait/retry/lock/signal surfaces, preservation classifications, authority prompts, and main-spec-sensitive risks are covered or explicitly source-limited. | S6 handoff, `s0-s6-gap-closure-plan.md`, and `s7-s6-audit-gap-follow-up.md` provide mandatory S7 carry-forward inputs. | Runtime readiness, cleanup/signal behavior, retained provenance, env/platform, package scripts, reset route, local-dev, visual snapshots, caches, and CI remain preserved unless selected evidence or owner decisions are cited. | S6 audits pass with source limits preserved; no implementation, fixture, cleanup, CI, generated, lockfile, runtime, package-manager, or snapshot behavior changed during S6. |

## Gap register

| Priority | Sprint | Gap / issue | Evidence source | Required follow-up | Blocks S7? | Validation / exit criteria |
|---|---|---|---|---|---|---|
| P0 | S0-S6 | S7 may proceed only if source limits, `TODO:*_unknown`, and owner questions are preserved. | `03-sprint-plan.md`; `s0-s6-gap-closure-plan.md`; `s7-s6-audit-gap-follow-up.md`; `source-limit-log.md` | Build S7 carry-forward checklist before drafting. | Yes, if not preserved. | Every S7 claim cites row IDs and keeps evidence labels. |
| P0 | Pre-S7 audit | Existing pre-S7 audit wording was stale relative to `SL-FU-0008` and authorized S7 evidence. | This audit document; `SL-FU-0008`; `runtime-evidence-register.md`; `cleanup-signal-evidence-register.md` | This revision supersedes the stale wording and records selected-evidence semantics. | Yes, if stale wording is treated as canonical. | Audit lists `SL-FU-0008` and states selected evidence closes only current-host, selected-command claims. |
| P1 | S0/S2/S5/S6 | Provider CI workflows/annotations are absent; selected `make ci` failed. | `SL-0001`; `AUTH-0015`; `PRES-0019`; run `05-ci`; `FAIL-0028` | Draft repo-local CI only; route provider behavior to owner; record CI failure. | No; blocks CI readiness/provider claims. | No `.github/**` or annotation behavior invented; no CI readiness claim. |
| P1 | S5/S7 evidence | `release-check` selected run failed. | `runtime-evidence-register.md` run `06-release-check`; `FAIL-0026` | Preserve as release-readiness gap. | No; blocks release readiness claims. | S7 does not claim release gate is passing. |
| P1 | S2/S4/S6 | Selected runtime evidence narrows but does not erase broad runtime limits. | `SL-FU-0008`; selected runs `02`-`19`, `24`, and `35` | Use selected run IDs only for current-host selected-command claims. | No. | Static evidence is not upgraded beyond selected runs. |
| P1 | S3/S4/S6 | Cleanup guarantees remain incomplete for parent death, live active DB cleanup, stale janitor authority, browser post-start failure cleanup, and detached reaper guarantee. | `SL-0014`; `SL-FU-0008`; `cleanup-signal-evidence-register.md` | Preserve best-effort/source-limited wording; route destructive cleanup to owner. | No; blocks guaranteed cleanup language. | Distinguish observed after-state, scheduling, best effort, and guarantees. |
| P1 | S1/S4/S6 | Reset route authority remains owner-required despite resettable browser evidence. | `AMB-0006`; `AUTH-0004`; `PRES-0010`; run `17-browser-resettable` | Route to maintainer plus product/security owner. | No; blocks public/reset-route contract. | Reset route is not presented as product API. |
| P1 | S2/S4/S6 | Env var public API and precedence unresolved. | `SL-0015`; `AMB-0012`; `AMB-0026`; `AUTH-0006`; `PRES-0014` | Record observed reads/defaults only; owner decides precedence. | No; blocks env normativity. | Every env rule is source-observed or owner-required. |
| P1 | S4/S6 | Platform/tool matrix unresolved; current host observed only. | `SL-0013`; `AUTH-0008`; `PRES-0015`; platform profile run `01` | Keep portability claims owner-required. | No; blocks portability guarantees. | Current WSL/Linux evidence is not generalized. |
| P1 | S2/S3/S4/S6 | Direct package scripts bypass Make result-root, scheduler, cleanup, env, and output guarantees. | `AMB-0011`; `AMB-0020`; `AUTH-0003`; `PRES-0011` | Default to Make canonical; route package-script status to owner. | No; blocks package-script contract. | Package scripts stay separate unless adopted. |
| P1 | S3/S5/S6 | Retained artifact provenance requires explicit selection. | `SL-0004`; `SL-0009`; `SL-0010`; `S7-DEC-0002`; `AUTH-0013` | Require `RESULTS_DIR`, `RUN_ID`, command, platform, exit status, and artifact paths. | No; blocks ambient artifact claims. | No newest-run evidence supports normative claims. |
| P2 | S3/S5/S6 | Machine schemas remain unknown for Playwright tool artifacts, shell log contents, CI annotations, and some failure bundles. | `SCHEMA-0002`; `SCHEMA-0013`; `SCHEMA-0019`; `TODO: schema_unknown` | Keep `schema_unknown` unless selected artifacts are authorized and inspected. | No; blocks stable schema requirements. | Every unknown machine surface stays explicit. |
| P2 | S3/S6 | Fixture, golden, and visual snapshot update authority unresolved. | `AMB-0015`; `AMB-0022`; `AUTH-0014`; `PRES-0018` | Default to owner-reviewed source edits; no refresh workflow. | No; blocks mutation/update rules. | No snapshot update authority is drafted. |
| P2 | S4/S6 | Local-dev Compose, `make dev`, fixed ports, persistent volumes, and MinIO reset gap remain authority-required. | `AUTH-0005`; `PRES-0012`; selected runs `20`-`23b` | Keep local-dev guidance separate from verification unless adopted. | No. | Local-dev evidence is not conformance contract. |
| P2 | S3/S6 | External Go cache cleanup scope unresolved. | `AMB-0021`; `AUTH-0009`; `PRES-0016` | Keep tool-managed/outside cleanup by default. | No. | S7 does not add `/tmp/cartulary-go-*` cleanup. |
| P2 | S4/S6 | Scheduler lanes are logical constraints, not host/service capacity guarantees. | `AUTH-0011`; `PRES-0003`; S6 audit | Draft lanes separately from concrete capacity. | No; blocks capacity guarantees. | No lane equals Docker/Postgres/MinIO/browser capacity. |
| P3 | S7 | NLSpec, acceptance matrix, roadmap, and review packet are intentionally absent. | `03-sprint-plan.md` S7 row; `TODO: harness_nlspec_draft_path`; `TODO: acceptance_matrix` | Do not draft them in this step. | No. | S7 deliverables remain not started. |

## Consolidated cross-sprint gaps

- Runtime evidence is selected and current-host only; CI/release failed, and
  portability remains unresolved.
- Cleanup strength remains limited for parent death, active DB cleanup, stale
  janitors, browser process-start failures, and reaper guarantees.
- Authority remains open for reset route, package scripts, local-dev services,
  stale janitors, env precedence, platform/tool support, external caches,
  visual snapshots, retained-run selection, and CI provider behavior.
- Schemas remain unknown for Playwright tool artifacts, CI annotations, shell
  log contents, and selected failure-only surfaces.
- Generated artifacts may drive execution when fresh, but they are downstream
  inputs, not behavior owners.

## Source limits to carry forward

Carry forward all current source limits:

- `SL-0001` through `SL-0015`
- `SL-FU-0001` through `SL-FU-0008`

`SL-FU-0008` narrows only selected, current-host claims with explicit run IDs
and result directories. It does not close CI provider behavior, failed
CI/release readiness, parent-death cleanup, live active DB cleanup, stale
janitor authority, browser post-start cleanup, local-dev contract status,
platform matrix, env precedence, package-script authority, or visual snapshot
update authority.

## Maintainer decisions required

S7 must route these questions instead of resolving them by inference:

- Canonical command surface: Make only, or direct package scripts too.
- Reset route ownership and product/security visibility.
- Local-dev Compose and `make dev` contract status.
- Stale janitor deletion proof for DBs, buckets, and containers.
- Public env var set and precedence.
- Supported platform/tool profile.
- External Go cache cleanup scope.
- Visual snapshot OS/browser/update command.
- Retained-artifact newest-run fallback policy.
- CI provider workflow/annotation authority.

## S7 blockers

Beginning S7 is blocked only if the carry-forward labels are not preserved or if
stale pre-S7 audit wording is used without reconciling `SL-FU-0008`.

Final normative S7 `MUST` language is blocked for the following until evidence
or owner decisions exist:

- live readiness;
- cleanup guarantees;
- reset-route public status;
- env precedence;
- package scripts;
- local-dev Compose;
- visual snapshot updates;
- external cache cleanup;
- CI provider annotations;
- release readiness;
- concrete capacity guarantees.

## Gaps preservable as source limits during S7

S7 may preserve these gaps as source-limited or owner-required instead of
closing them:

- CI provider source and annotation behavior.
- Failed selected CI and release readiness.
- Planned phase7/phase8 files.
- Non-selected broad gate, service-backed, browser, Docker/Compose, reset, and
  controlled-failure runtime behavior.
- Cleanup timing and completion guarantees not covered by selected evidence.
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

1. Freeze S0 through S6 as complete and record the authorized S7 evidence pass
   as evidence-only.
2. Supersede or reconcile stale pre-S7 audit wording with `SL-FU-0008`.
3. Build the S7 carry-forward checklist from `SL-*`, `AMB-*`, `AUTH-*`,
   `PRES-*`, `MSC-*`, `FAIL-*`, and remaining `TODO:*_unknown` markers.
4. Classify each S7 candidate claim as exactly one of
   `selected_runtime_observed`, `source_observed`, `source_limit`,
   `maintainer_decision_required`, or `deferred_roadmap`.
5. Route owner-required claims before final normative wording.
6. Draft no NLSpec, acceptance matrix, roadmap, or review packet until the
   readiness criteria below pass.

## Final readiness criteria for beginning S7

S7 may begin when all of the following are true:

- Every unresolved source limit and owner question is preserved.
- Selected runtime evidence is cited only by explicit run ID and result
  directory.
- Stale audit wording is reconciled.
- CI and release failures are recorded as non-readiness evidence.
- Every remaining `TODO:*_unknown` is explicit or routed.
- No implementation, fixture, cleanup, CI workflow, generated artifact,
  lockfile, package-manager, or snapshot change is introduced during the
  transition.

## No-change audit

This audit plan and gap register is documentation-only. It does not change
harness implementation behavior, product implementation behavior, test logic,
fixture or snapshot contents, CI behavior, cleanup scripts, generated outputs,
lockfiles, package-manager state, runtime services, or S7 deliverables.
