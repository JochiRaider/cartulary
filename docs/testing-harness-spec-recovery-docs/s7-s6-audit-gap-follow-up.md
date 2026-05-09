---
doc_id: THR-S7-S6-AUDIT-GAP-FOLLOW-UP-2026-05-09
title: S7 S6 Audit Gap Follow-Up Track
status: active
role: s7-gap-follow-up
---

# S7 S6 Audit Gap Follow-Up Track

## Audit-Gap Follow-Up Objective

Close or deliberately preserve the remaining S6 hazards, resources, and timing
gaps so S7 can draft the harness NLSpec without upgrading static evidence into
runtime guarantees. The default S7 posture is:

- draft from existing S0 through S6 and S8 evidence;
- keep runtime-sensitive behavior labeled `source_limit`;
- keep maintainer-owned behavior labeled `maintainer_decision_required`;
- record later runtime commands as authorized evidence collection, not as
  commands executed by this documentation pass.

This track implements the follow-up plan for the S6 audit verdict
`pass_with_source_limits_preserved`.

Companion preflight input: `s0-s6-gap-closure-plan.md` consolidates S0 through
S6 completion evidence, remaining source limits, open owner questions, and
readiness criteria. S7 must use it together with this S6-focused follow-up
track before drafting normative harness requirements.

## Scope and Non-Scope

In scope:

- S0 through S6 readiness criteria from `s0-s6-gap-closure-plan.md`.
- S7 drafting rules for carrying S6 source limits and owner questions forward.
- A grouped inventory of remaining S6 gaps.
- Later evidence-collection phases and scenarios, all marked as requiring
  separate authorization when they may create runtime artifacts or external
  state.
- Owner questions and decision owners for public or authority-sensitive harness
  contracts.
- S7 blockers versus non-blocking roadmap follow-ups.

Out of scope:

- Harness rewrites, generated output changes, lockfile or package-manager
  changes, formatter or generator runs, baseline refreshes, snapshot updates,
  cleanup execution, Docker/testcontainers, Docker Compose, browser E2E,
  service-backed execution, reset calls, package-test execution, and broad
  verification.
- Product behavior inference from harness implementation, especially reset
  route behavior, local credentials, local-dev topology, environment precedence,
  visual snapshots, or CI provider behavior.
- Any change to timing, retry, sharding, scheduler, cleanup, resource
  allocation, reset, tests, fixtures, generated outputs, lockfiles, or package
  manager state.

## Authority Boundary

Core 00 through Core 04 remain the authority for product behavior,
deployment/security boundaries, public API contracts, workbook surfaces, record
models, and conformance. This follow-up track may define only harness recovery
and validation-planning mechanics. It must not turn implementation behavior into
Core behavior.

`docs/domain.md` remains vocabulary and owner-section context only. It is not a
runtime behavior owner.

## Grouped Gap Inventory

| Gap group | Controlling rows | Remaining gap | S7 handling |
|---|---|---|---|
| S0-S6 closure preflight | `s0-s6-gap-closure-plan.md`, S0-S6 handoffs and audits | S0-S6 are complete only if source limits, named TODOs, and owner questions remain preserved. | Use the closure plan as the S7 preflight gate before drafting requirements. |
| Runtime readiness and services | `SL-0012`, `SL-FU-0001`, `RTR-0004`, `SVC-0001..SVC-0015` | Live Docker/testcontainers, Postgres, MinIO, Compose, browser stack, service-backed targets, reset route, and readiness behavior remain unobserved. | Draft as source-observed lifecycle and keep live behavior `source_limit` unless later authorized evidence exists. |
| Cleanup and timing guarantees | `SL-0014`, `SL-FU-0002`, `RTR-0007`, `RTR-0008`, `RTR-0012` | Timeout, interrupt, parent-death, detached reaper completion, active DB connection cleanup, stale janitor execution, and port release remain source-limited. | Do not claim guaranteed cleanup completion; distinguish cleanup path, cleanup scheduling, best effort, and verified completion. |
| Artifact provenance | `SL-0004`, `SL-0009`, `SL-0010`, `SL-FU-0005`, `RTR-0016`, `PRES-0017` | Retained artifact freshness, failure-only bundles, and newest-run fallback need explicit selected-run provenance. | Normative or durable claims require explicit `RESULTS_DIR`, `RUN_ID`, command, platform/tool profile, and artifact paths. |
| Environment and platform contracts | `SL-0013`, `SL-0015`, `AUTH-0006`, `AUTH-0008`, `PRES-0014`, `PRES-0015` | Public env-var set, override precedence, supported OS/tool profile, and missing-tool behavior need owner decisions. | Record observed env reads/defaults only; keep precedence and platform support owner-required. |
| Authority-only surfaces | `AUTH-0003`, `AUTH-0005`, `AUTH-0009`, `AUTH-0014`, `AUTH-0015` | Direct package scripts, local-dev Compose and `make dev`, external Go caches, visual snapshot update authority, and CI provider workflow/annotation behavior remain open owner questions. | Keep as developer/local/diagnostic or out-of-contract surfaces unless maintainers adopt them. |

## Proposed Evidence Collection Phases

1. S7 source-limit carry-forward, no runtime.
   Draft NLSpec sections from existing row IDs only. Every runtime-sensitive
   claim must retain `source_limit` or `maintainer_decision_required`.

2. Authorized runtime readiness pass.
   After approval, run a small, uniquely named result-root session to observe
   service-backed and browser readiness without changing code. Record selected
   run identity, platform/tool profile, command, timestamps, exit status, and
   artifact paths.

3. Authorized cleanup/signal pass.
   After approval, run controlled timeout, interrupt, and parent-death
   scenarios in disposable Docker/testcontainers/browser stacks. Record cleanup
   summaries, leases, reaper logs, service events, and port checks.

4. Authorized artifact provenance pass.
   After approval, use one selected successful run and one selected controlled
   failure run. Inspect only those result roots with explicit `RESULTS_DIR` and
   `RUN_ID`.

5. Owner-decision pass.
   Maintainers decide contract status for env vars, supported platform/tool
   profile, package scripts, local-dev behavior, external Go caches, visual
   snapshot authority, reset-route boundary, stale janitor bounds, and CI
   provider source.

## Later Commands and Scenarios

Every command in this section requires separate later authorization because it
may create runtime artifacts or external state. These commands were not run
while creating this follow-up track.

| Evidence area | Later authorized command or scenario | Required evidence record |
|---|---|---|
| Service-backed Docker/Postgres/MinIO readiness | `make backend-store`, `make backend-integration`, and `make backend-process` with unique `CARTULARY_TEST_RESULTS_DIR` and `CARTULARY_TEST_RUN_ID`. | Suite events, service leases, fixture summaries, cleanup summaries, command, run ID, result dir, platform/tool profile, timestamps, and exit status. |
| Phase/service-backed coverage | Select the phase with static `make explain-phase PHASE=<phase>`, then run `make service-backed-slice PHASE=<selected phase>`. | Phase selection rationale, scheduler summaries/events, result dir, run ID, and service-backed child artifacts. |
| Browser/reset/readiness | `make browser-e2e-webserver-backed` for shared-stack readiness and `make browser-e2e-resettable` for reset-route evidence. | Browser stack env/json, backend/frontend logs, reset response/status files, Playwright reports, result dir, run ID, and explicit reset-route authority status. |
| Compose/local-dev | `make services-up` or `make db-up`, then a controlled `make dev` session. | Readiness behavior, manual interrupt behavior, fixed-port release, persistent Compose state notes, and `AUTH-0005` decision status. |
| Cleanup/signal behavior | Controlled wrapper runs around `tools/testservices run -- <long-running child>` and browser stack sessions, then TERM/INT or parent timeout/death. | Signal sent, parent/child status, cleanup summaries, detached reaper logs, container labels, DB/bucket cleanup summaries, and `ss` port-release output where available. |
| Artifact provenance | `make explain-run RESULTS_DIR=<selected run> RUN_ID=<run id>`, `make fixture-report RESULTS_DIR=<selected run>`, and selected drift readers such as `make scheduler-event-order-drift RESULTS_DIR=<selected run>`. | Explicit selected run identity, no ambient newest-run claim, schema validity, and artifact paths. |
| Failure-only bundles | One authorized controlled failure for Go, Vitest, and/or browser if S7 needs stable machine schemas. | Failure command, selected result dir/run ID, runner logs, watchdog JSON, Playwright traces/screenshots/videos/reports where produced. |
| Visual snapshots | Validation-only target such as `make browser-e2e-visual`, after authorization. | Validation result only. Do not run snapshot update tooling until `AUTH-0014` names the OS, browser version, and update command. |
| Cleanup targets | `make clean` or `make distclean`, only with separate cleanup authorization. | Scope decision and before/after path inventory. External `/tmp/cartulary-go-*` cache removal remains out of contract by default. |

## Owner Questions and Decision Owners

| Decision owner | Required decision | Current default until decided |
|---|---|---|
| Harness command owner | Confirm Make remains canonical and direct package scripts are developer conveniences unless adopted. | Make is canonical; package scripts are separate. |
| Harness/service owner plus product/security input | Decide reset-route visibility, stale janitor deletion proof, cleanup guarantee strength, and detached reaper contract. | Reset route and destructive cleanup remain authority-required; cleanup completion remains source-limited. |
| Harness config owner plus Core 04 owner for overlapping product config | Decide public env-var set and precedence across Make, generated recipes, schedulers, shell wrappers, package scripts, Go helpers, config files, and Playwright. | Observed reads/defaults only; precedence remains unresolved. |
| Harness/platform owner | Define supported platform/tool profile for Linux/WSL, Docker, Compose, `ss`, `curl`, `setsid`, `realpath`, shell, localhost networking, Node/pnpm, and browsers. | Platform/tool support remains source-limited except where source explicitly validates or fails. |
| Local-dev owner | Decide whether Compose fixed ports, persistent volumes, `db-reset` MinIO exclusion, and `make dev` are harness contract or operator guidance. | Local-dev behavior is operator/local-dev guidance, not verification contract. |
| Browser/harness owner | Decide visual snapshot platform/browser/update authority. | Validation-only; no snapshot update authority. |
| CI owner | Decide whether provider workflows are external, absent, or intentionally represented only by `scripts/ci/**`. | Provider workflow and annotations remain absent/source-limited while `.github/**` is absent. |

## S7 Blockers

No blocker exists for starting S7 if S7 preserves source limits and owner
questions.

The following block final normative S7 `MUST` language until evidence or owner
decisions exist:

- any claim that live readiness, cleanup completion, signal or timeout behavior,
  reaper completion, active DB cleanup, stale janitor deletion, or port release
  is guaranteed;
- any public contract for reset route, env precedence, direct package scripts,
  local-dev Compose, visual snapshot updates, external Go cache cleanup, or CI
  provider annotations without owner decision;
- any claim that scheduler lanes are concrete host, Docker, Postgres, MinIO, or
  browser capacity guarantees.

## Non-Blocking Follow-Ups

- Runtime evidence collection can remain roadmap work while S7 drafts
  source-limited mechanics.
- Failure-only bundle schemas can remain `schema_unknown` unless S7 makes them
  normative machine interfaces.
- Local-dev Compose and direct package scripts can remain documented
  compatibility/developer surfaces outside Make-owned verification.
- External Go caches should default to tool-managed and outside `clean` and
  `distclean`.
- Provider CI annotations should remain absent/source-limited while `.github/**`
  is absent.

## Validation Criteria

- Every S7 requirement cites existing recovery rows or a later authorized
  runtime evidence record.
- Every source-limited gap remains labeled `source_limit` until an approved run
  records selected evidence.
- Every maintainer-owned decision remains `maintainer_decision_required` until
  explicitly decided.
- Every selected runtime run records `RESULTS_DIR`, `RUN_ID`, command,
  platform/tool profile, start time, exit status, and artifact paths.
- Documentation planning does not change harness implementation,
  timing/retry/sharding/scheduler behavior, cleanup behavior, generated output,
  fixtures, lockfiles, package-manager state, or snapshot baselines.

## S7 Carry-Forward Checklist

- [ ] Use `s0-s6-gap-closure-plan.md` as the S7 preflight gate for all S0
      through S6 remaining gaps.
- [ ] Carry `SL-0012..SL-0015`, `SL-FU-0001..SL-FU-0007`, and relevant
      `AMB-FU-*` rows into S7.
- [ ] Mark runtime readiness and cleanup guarantees as source-limited unless
      authorized evidence exists.
- [ ] Route reset route, stale janitor, env, platform, package-script,
      local-dev, cache, visual, and CI questions to owners.
- [ ] Use explicit selected-run provenance for any retained artifact claim.
- [ ] Keep Core 00 through Core 04 product/deployment authority untouched.
- [ ] Record all authorized evidence commands separately from this
      documentation-only plan.

## Documentation-Only Implementation Checklist

- [x] Follow-up objective recorded.
- [x] Scope and non-scope recorded.
- [x] Grouped gap inventory recorded.
- [x] Evidence collection phases recorded.
- [x] Later commands and scenarios marked as requiring authorization.
- [x] Owner questions and decision owners recorded.
- [x] S7 blockers and non-blocking follow-ups separated.
- [x] Validation criteria recorded.
- [x] Final S7 carry-forward checklist recorded.
- [x] No runtime, service, browser, Docker, Compose, reset, cleanup, formatter,
      generator, baseline-refresh, snapshot-update, package-test, or broad
      verification command was run for this documentation pass.

## Authorized Runtime Evidence Implementation Note

On 2026-05-09, a later maintainer-authorized S7 evidence pass implemented the
runtime-evidence plan without rewriting harness code or promoting source-limited
behavior into final `MUST` language.

Evidence outputs:

- `runtime-evidence-register.md`
- `cleanup-signal-evidence-register.md`
- `source-limit-log.md` follow-up `SL-FU-0008`

The selected runs close only the rows validated by explicit
`CARTULARY_TEST_RESULTS_DIR` and `CARTULARY_TEST_RUN_ID`. CI and release runs
failed with current harness-smoke/phase-map accounting evidence; parent-death
cleanup, live active-DB cleanup, stale janitor authority, local-dev contract
status, platform matrix, CI provider behavior, and visual snapshot update
authority remain source-limited or maintainer-decision-required.
