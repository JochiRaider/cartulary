---
doc_id: THR-020
title: Testing Harness NLSpec Writing Guide
status: draft
role: writing-guide
---

# Testing Harness NLSpec Writing Guide

## Document role

This guide defines how agents maintain and revise the recovered testing harness
NLSpec package after the initial recovery pass. The current process state is
recorded in `01-recovery-process.md`: S0 through S12 are complete, the package
is reviewable, and later work is maintenance or targeted follow-up.

Use this guide when changing any of these package outputs:

| Package output | Writing-guide use |
|---|---|
| `harness-nlspec.md` | Keep the harness contract complete, evidence-bounded, and acceptance-mapped. |
| `harness-acceptance-matrix.md` | Keep `HAC-*` criteria binary and keep `HAC-GAP-*` blockers explicit. |
| `harness-implementation-roadmap.md` | Keep future remediation separate from recovered specification claims. |
| `harness-review-packet.md` | Keep review status, verification, source limits, and handoff material current. |
| `maintainer-decision-summary-2026-05-09.md` | Use recorded maintainer decisions as settled inputs unless a later owner decision supersedes them. |
| `source-limit-log.md` and `ambiguity-register.md` | Preserve unresolved limits and owner questions without inference. |
| Recovery registers and audits | Use as evidence sources; do not treat them as independent behavior owners. |

This guide is not evidence of harness behavior. It does not authorize
implementation, fixture, generated-artifact, CI, cleanup, lockfile,
package-manager, runtime-service, or product-behavior changes.

## Writing objective

The harness NLSpec must define the observable contract of the testing harness.
A future implementer must be able to recreate caller-visible harness behavior
from the NLSpec package without guessing about command authority, defaults,
sequencing, resources, artifacts, schemas, errors, cleanup, diagnostics,
source limits, or acceptance criteria.

The harness NLSpec must not merely describe the current implementation. It must
separate normative harness behavior, selected evidence, source-observed
behavior, compatibility-only behavior, source limits, contradictions,
maintainer-decision-required items, and future redesign candidates.

## NLSpec quality gates

Every revision must pass these gates before it is presented as reviewable
NLSpec material.

| Gate | Required result |
|---|---|
| Behavioral completeness | Every externally observable harness behavior in scope is specified, intentionally excluded, or marked with a source limit or owner decision. |
| Interface completeness | Command inputs, environment variables, config files, report schemas, artifact paths, failure classes, lifecycle states, and acceptance rows have explicit fields and meanings. |
| Boundary completeness | Defaults, omitted cases, invalid inputs, limits, timeouts, retry bounds, cleanup guarantees, and compatibility boundaries are stated where they affect callers. |
| Mapping completeness | Tables cover translations across command surfaces, artifact authority, cleanup tiers, schema status, failure classes, service states, resource classes, and CI entrypoints whenever prose would hide cases. |
| Acceptance completeness | Every final `must` or `must not` maps to one `HAC-*` criterion or to a named `HAC-GAP-*`, source-limit, or owner-decision blocker. |
| Recreatability | If implementation code were unavailable, the NLSpec package would still let a competent implementer recreate the harness contract for all non-source-limited behavior. |
| Interchangeability | Two independent implementations following the NLSpec would be caller-interchangeable for all non-source-limited behavior. |
| Spec economy | Each fact is defined once, in the owner section, and referenced elsewhere by ID or stable term. Repetition is allowed only when it prevents misreading and does not create a second source of truth. |

If a revision cannot pass a gate, it must add or update the relevant
`source_limit`, `maintainer_decision_required`, `HAC-GAP-*`, ambiguity, or
roadmap entry instead of filling the gap with guessed behavior.

## Normative language

Use these words consistently.

| Term | Meaning |
|---|---|
| `must` | Required contract behavior for the stated scope. |
| `must not` | Prohibited contract behavior for the stated scope. |
| `may` | Permitted behavior that callers cannot require. |
| `default` | Behavior when the caller omits a configurable input. |
| `source-limited` | Evidence is insufficient for a final normative claim. |
| `owner-required` | A maintainer or governing specification must decide before the claim can become final. |

Rules:

- Use lowercase `must`, `must not`, `may`, and `default` for contract-bearing
  harness behavior, matching the current S7 package style.
- Do not use advisory language for contract behavior. Rewrite it as `must`,
  `must not`, `may`, or an explicit source limit.
- Do not use undefined quality adjectives or open-ended modifiers. If the
  behavior needs a tolerance, bound, retry count, timeout, artifact field, exit
  status, or cleanup tier, state it directly.
- Use stable tokens for bounded concepts. For cleanup, use
  `best_effort_cleanup` only as the cleanup tier defined in this guide and in
  the harness NLSpec.
- Keep examples non-normative unless their text uses the same evidence,
  default, boundary, and acceptance structure as the contract section.

## Evidence-to-spec rules

### Evidence status mapping

| Evidence status | Allowed specification treatment |
|---|---|
| `selected_runtime_observed` | may support a final harness requirement only for the selected command, environment, inputs, artifacts, and exit status that were actually observed. |
| `runtime_observed` | may support a scoped claim when the run identity, command, platform/tool profile, exit status, and artifact paths are recorded. Otherwise convert to `source_limit`. |
| `source_observed` or `observed` | may support source-level behavior, static command contracts, declared defaults, schema declarations, or routing rules. It does not prove runtime success. |
| `maintainer_decision` | may close an authority or intent gap when the decision ID is cited and no primary owner conflicts. |
| `maintainer_decision_required` or `owner_required` | must remain non-final. Route to an ambiguity row, `HAC-GAP-*`, or roadmap item. |
| `source_limit` | must not become final normative behavior. State what was inspected, what was not inspected, the blocked claim, and the follow-up needed. |
| `inferred` | may explain a draft hypothesis or connect evidence rows. It cannot be the sole support for final normative behavior. |
| `assumed` | must not support normative behavior. Replace with evidence, source limit, or owner decision. |
| `contradiction` | must not be reconciled silently. Record the conflict, owner, blocked claim, and current non-normative treatment. |

### Authority surface mapping

| Surface | Specification treatment |
|---|---|
| Core 00 through Core 04 | Own product behavior. Harness text may define validation mechanics only. |
| `docs/domain.md` | Provides vocabulary and boundary context for domain-facing text; it does not replace owner specs. |
| Harness NLSpec package | Owns adopted harness mechanics, command authority, validation orchestration, artifacts, cleanup bounds, diagnostics, and acceptance mapping. |
| `make` public surface | Sole canonical harness command surface per `MD-S7-0001`. |
| Direct package scripts | Developer conveniences unless they re-enter Make-owned wrappers. Do not promote them to first-class harness contracts. |
| Generated task, schedule, Go, and TypeScript artifacts | Downstream execution inputs when fresh. They do not own behavior and must not be hand-edited. |
| Harness implementation, tests, and fixtures | Evidence of current behavior or validation expectations. They are not final authority when they conflict with owner specs or maintainer decisions. |
| CI configuration | Provider-neutral operational gate evidence unless provider workflow sources are present and adopted. |
| Logs, reports, traces, screenshots, dashboards | Diagnostic or derived surfaces unless a schema or owner decision explicitly makes a field contract-bearing. |
| Local policy and retained local artifacts | Evidence only for the recorded local context. Durable claims require explicit run identity. |

### Preserved source-limit mapping

The current package must preserve these boundaries until later selected evidence
or maintainer decision closes them.

| Boundary | Required wording |
|---|---|
| Environment-variable precedence | State only source-observed variables and defaults. Keep precedence as `TODO: precedence_unknown` or source-limited. |
| Visual snapshot refresh OS, browser, version, and command | State that visual snapshots are validation-only. Do not adopt a snapshot update command or platform bounds. |
| Parent-death cleanup | Do not claim guaranteed abrupt-exit cleanup. |
| Active DB cleanup | Do not claim guaranteed live-connection cleanup. |
| Detached reaper hard completion | Treat cleanup scheduling evidence as scheduling evidence, not completion proof. |
| Provider-specific CI while `.github/**` is absent | Keep CI provider-neutral. Do not invent workflow, annotation, upload, or dashboard behavior. |
| Playwright report, trace, video, and screenshot internals | Treat as tool-owned or `schema_unknown` unless a later owner adopts stable harness schemas. |
| Release readiness beyond recorded evidence | Keep readiness claims separate from stale-smoke demotion unless the named release gate passes or each failure is classified. |

### Cleanup tier mapping

| Cleanup tier | Allowed claim strength |
|---|---|
| `observed_successful_cleanup` | Selected runtime evidence shows a specific cleanup path completed successfully. |
| `observed_cleanup_scheduling` | Source or selected evidence shows cleanup was scheduled; completion is not proven. |
| `delayed_after_state_evidence` | Later observation saw the expected after-state; synchronous cleanup completion is not proven. |
| `best_effort_cleanup` | Source attempts cleanup; failures, interrupts, or detached completion are not guaranteed. |
| `source_limited_cleanup` | Source exists or routing is known; selected runtime evidence or owner decision is missing. |

Do not write a stronger cleanup claim than the tier supports.

### Schema status mapping

| Schema status | Specification treatment |
|---|---|
| Stable schema ID adopted by source or acceptance matrix | may define field names, types, required/default/conditional fields, ordering, null/omission semantics, and compatibility rules. |
| `partial` | may define only inspected fields and must mark unknown fields or cases as source-limited. |
| `schema_unknown` | must remain diagnostic or tool-owned. Do not specify field contracts. |
| `authority_unknown` | must route to owner decision before field contracts become final. |
| Tool-owned report format | may be referenced as diagnostic output. Do not re-specify internals unless the harness adopts them as stable schemas. |

## Harness NLSpec topology

Maintain the current harness NLSpec around these behavior areas. A section may
be compact, but it must close the listed decisions or preserve the gap
explicitly.

| Section area | Required content |
|---|---|
| Document role and scope | Current package status, harness-only scope, product-behavior exclusions, and evidence split. |
| Authority | Core ownership, Make authority, generated-artifact authority, direct-script status, and settled maintainer decisions. |
| Make command surface | Canonical entrypoints, caller modes, inputs, defaults, outputs, side effects, and failure behavior. |
| Local-dev verification | `make db-up`, `make services-up`, `make db-reset`, `make dev`, config-file use, and local-only boundaries. |
| Scheduler model | Logical lanes, claims, scheduling constraints, duration-baseline role, and non-guarantees for physical capacity. |
| Environment and platform | Source-observed variables and defaults, unresolved precedence, WSL/Linux observation, and non-Linux/source-limit boundaries. |
| Retained artifacts | Required run identity fields, newest-run investigation limits, artifact authority, freshness, and diagnostic status. |
| Reset route | Harness ownership, disabled default, explicit test-route wiring, schema ID, product-API exclusion, and partial-failure limits. |
| Cleanup and destructive safety | Cleanup tiers, deletion proof gates, cache exclusions, interrupt/crash limits, and cleanup-error behavior. |
| Fixtures, goldens, and snapshots | Controlled update workflow, missing-item review, canonical-vs-generated distinction, and visual-snapshot update limits. |
| Provider-neutral CI and stale smoke | `make ci`, `make release-check`, absent provider workflow limits, stale-smoke demotion, and diagnostic target status. |
| Failure taxonomy | Failure class meanings, trigger phase, exit/report behavior, cleanup behavior, retryability, and owner. |
| Acceptance hooks | `HAC-*` mapping, `HAC-GAP-*` blockers, selected evidence status, and current coverage. |
| Open limits and roadmap routing | Source limits, owner decisions, future remediation, and classifications that prevent accidental normativity. |

## Contract writing pattern

Use this structure for every major behavior that carries contract weight.

```markdown
### <Behavior name>

The harness must <observable requirement> when <scope condition>.

Authority and evidence:
- <owner spec, maintainer decision, source row, or selected runtime row>

Inputs:
- <input name, type, source, and valid values>

Defaults and omitted cases:
- <default behavior when input is absent>

Outputs:
- <stdout/stderr/report/artifact/exit behavior>

Side effects:
- <created, mutated, deleted, or retained state>

Failure behavior:
- `<failure_class>`: <caller-visible result, cleanup behavior, retryability>

Source limits:
- <none, or exact blocked claim and follow-up needed>

Acceptance:
- `<HAC-####>`: <binary pass condition>
```

If declarative prose cannot determine ordering, selection, retry, cleanup,
resource allocation, or report fan-in behavior, use deterministic pseudocode.

```text
Algorithm: <algorithm name>

Input:
  <input name and type>

Preconditions:
  <required state>

Steps:
  1. <deterministic step>
  2. <deterministic step>
  3. <deterministic step>

Output:
  <output value or artifact>

Errors:
  - <failure_class>: <observable behavior>

Acceptance:
  - <HAC-####>: <binary check>
```

## Required mapping tables

Add or maintain a table whenever there are more than two cases or when a table
reduces ambiguity. The harness package must keep these mappings explicit where
the behavior is in scope:

- Authority surface to owner and conflict behavior.
- Make entrypoint to lifecycle, artifacts, and allowed side effects.
- Command mode to allowed side effects.
- Environment variable to default, owner, and precedence status.
- Scheduler lane or resource claim to logical meaning and non-guarantee.
- Artifact authority class to execution effect and update authority.
- Retained artifact claim to required run identity fields.
- Cleanup tier and cleanup trigger to allowed claim strength.
- Stale janitor deletion proof gate to deletion permission.
- Service state to recovery action.
- Failure class to exit, report, cleanup, retryability, and owner.
- Schema status to stable-contract or source-limit treatment.
- CI surface to Make entrypoint and provider-specific limits.

## Mechanics that must be explicit

| Mechanics surface | Required specification |
|---|---|
| Command inputs | Flags, environment variables, positional args, config files, omitted-case behavior, invalid input behavior. |
| Exit behavior | Exit code meanings, stdout/stderr semantics, partial-output behavior, and CI summary behavior. |
| Report schemas | Field names, types, required/default/conditional fields, ordering, null/omission semantics, and schema status. |
| Artifact paths | Canonical path rules, run-local paths, collision behavior, cleanup owner, persistence, and absence/staleness behavior. |
| Resource allocation | Worker IDs, ports, temp dirs, database names, browser profiles, locks, leases, collision behavior, and release behavior. |
| Service readiness | Ready condition, poll interval, timeout, failure result, stale service handling, reset rule, and stop rule. |
| Retry and backoff | Eligible failures, maximum attempts, delay algorithm, reset behavior, and exhaustion result. |
| Timeout and cancellation | Units, defaults, minimums, maximums, signal or shutdown sequence, partial artifacts, and cleanup behavior. |
| Fixture and golden updates | Owner intent, targeted file list, evidence or test reason, verification command, and review note. |
| Cleanup | Trigger, order, idempotence, cleanup tier, failure behavior, retained diagnostics, and non-guarantees. |
| Ordering | Test discovery, report ordering, log ordering, artifact enumeration, parallel fan-in, and retained-run selection. |
| Compatibility | Existing caller dependence, replacement path, warning behavior, deprecation gate, and removal gate. |

## Failure taxonomy rules

The harness spec must classify failures by phase and caller-visible behavior.
Each class must map to exit behavior, report behavior, cleanup behavior,
retryability, and owner.

| Failure class | Required meaning |
|---|---|
| `usage_error` | Caller passed invalid command, flag, path, environment, mode, or unsupported combination. |
| `configuration_error` | Required harness configuration is missing, invalid, or contradictory. |
| `preflight_error` | Harness cannot start because dependencies, secrets, platform, or prerequisites are missing. |
| `service_start_error` | Required service cannot be provisioned or started. |
| `service_readiness_timeout` | Service started but never satisfied the ready condition. |
| `fixture_error` | Fixture is missing, malformed, stale, or unauthorized to mutate. |
| `resource_conflict` | Port, temp path, lock, database, browser profile, worker slot, or cache conflicts. |
| `test_assertion_failure` | Product-under-test assertion failed after harness setup reached the test execution phase. |
| `harness_internal_error` | Harness code failed independently of a product assertion. |
| `timeout` | Configured timeout elapsed. |
| `cancelled` | Caller, CI, or controlling process cancelled execution. |
| `cleanup_error` | Cleanup failed or left ambiguous state. |
| `unsupported_platform` | Current platform cannot satisfy the declared harness contract. |
| `missing_secret` | Required secret or credential is unavailable. |
| `manifest_or_accounting_mismatch` | Harness manifests, generated task surfaces, schemas, or accounting reports disagree. |
| `authority_required` | A failure cannot be classified normatively until an owner decides the contract. |
| `unknown_failure` | Failure was observed but not classified. This class must not remain in a final accepted claim. |

## Acceptance matrix rules

Acceptance criteria live in `harness-acceptance-matrix.md`.

Use `HAC-*` IDs for binary criteria and `HAC-GAP-*` IDs for known blockers.
Do not introduce a second acceptance ID family in the harness package.

| Field | Required content |
|---|---|
| `Criterion ID` | Stable ID such as `HAC-0001` or blocker ID such as `HAC-GAP-0001`. |
| `Requirement` | One binary requirement or one named blocked claim. |
| `Validation type` | Static inspection, unit test, integration test, CI run, dry run, golden fixture, manual review, owner decision, or source limit. |
| `Validation command or evidence` | Exact command, file, section, register row, selected run, or owner decision ID. |
| `Expected result` | Exact pass condition or exact blocker condition. |
| `Evidence status` | `selected_runtime_observed`, `source_observed`, `source_limit`, `maintainer_decision`, `maintainer_decision_required`, or another defined status. |

Rules:

- One criterion tests one requirement.
- A criterion that depends on maintainer judgment must use an owner-decision
  evidence status and name the decision needed.
- Existing tests may be cited only after inspection, and runtime success may be
  cited only with selected run evidence.
- Passing an existing test does not prove that the requirement is sufficient;
  the criterion must still state the contract it verifies.
- A final `must` or `must not` without an acceptance row is a package defect.

## Product authority rules

The harness specification must not redefine product behavior owned by Core 00
through Core 04. It may define how the harness validates product behavior.

Use this pattern when a harness behavior touches product semantics:

```markdown
Core <owner> owns <product behavior>. This harness specification owns
<validation behavior>. If a harness fixture, test, reset route, or report
encodes product behavior not present in the owner specification, the package
must record an authority gap and must not promote that behavior to harness
normativity by itself.
```

High-risk boundaries:

- Reset routes are test hooks, not product APIs.
- Local-dev Compose and `make dev` are local verification behavior, not
  deployment conformance.
- Browser stress margins validate harness coverage and do not change product
  session semantics.
- Generated artifacts may drive execution when fresh, but they do not own
  product or harness behavior.

## Preservation and roadmap rules

Future implementation work must be classified separately from recovered
specification work.

| Classification | Meaning |
|---|---|
| `preserve` | Make current behavior part of the harness contract. |
| `preserve_with_clarification` | Keep behavior but specify missing defaults, errors, or boundaries. |
| `refactor_preserving_behavior` | Mechanism may change; observable behavior must not. |
| `deprecate` | Keep temporarily with warning, replacement path, and removal gate. |
| `redesign_required` | Current behavior is unsafe, inconsistent, or not specifiable as-is. |
| `remove_if_unused` | Appears unused and non-contractual. Prove no dependency before removal. |
| `authority_decision_required` | Human decision required before normative closure. |
| `exclude_from_contract` | Behavior is incidental and not relied on. State the non-guarantee when reliance risk exists. |

Do not write roadmap items as completed implementation. Each roadmap item must
state the current treatment, future action, owner or evidence needed, and the
acceptance or blocker row it affects.

## Closure checklist

Before presenting a revised harness NLSpec package for review, verify:

- [ ] The revision uses the current recovery state from `01-recovery-process.md`
  and does not restart completed stages without new evidence or request.
- [ ] Settled S7 decisions are cited by `MD-S7-*` ID and are not re-decided.
- [ ] Every final `must` or `must not` maps to exactly one `HAC-*` criterion or
  to an explicit `HAC-GAP-*`, source-limit, or owner-decision blocker.
- [ ] Every command has inputs, defaults or omitted-case behavior, outputs, side
  effects, and failure behavior.
- [ ] Make remains the canonical harness command surface; direct package
  scripts remain developer conveniences unless they re-enter Make wrappers.
- [ ] Generated artifacts are described as downstream execution inputs, not
  behavior owners, and no generated path is hand-edited.
- [ ] Every generated artifact, retained artifact, log, report, fixture, golden,
  and snapshot has authority, lifecycle, absence/staleness, and cleanup
  treatment.
- [ ] Missing fixture, golden, and snapshot review is explicit, including a
  reviewed absence statement when no missing items are identified.
- [ ] Every service has provisioning, readiness, reset, stop, cleanup, and
  failure behavior or a source-limit row.
- [ ] Every timeout has unit, default, minimum or omitted lower bound, maximum,
  and timeout result.
- [ ] Every retry has eligible failures, attempt bound, delay algorithm, reset
  behavior, and exhaustion behavior.
- [ ] Every shared resource has allocation, collision, release, reuse, and
  parallel-safety behavior.
- [ ] Environment precedence remains `TODO: precedence_unknown` or
  source-limited until an owner supplies a precedence matrix or decision.
- [ ] Visual snapshots remain validation-only until an owner supplies exact OS,
  browser, version, and update-command bounds.
- [ ] Parent-death cleanup, active DB cleanup, and detached reaper completion do
  not receive stronger guarantees than the evidence supports.
- [ ] Provider-specific CI behavior is not invented while provider workflows are
  absent.
- [ ] Playwright report, trace, video, and screenshot internals remain
  tool-owned or `schema_unknown` unless adopted as stable harness schemas.
- [ ] No derived diagnostic surface is silently promoted to authority.
- [ ] No product behavior is redefined under the harness spec.
- [ ] Open contradictions have a named owner, decision prompt, blocked claim,
  and current treatment.
- [ ] The review packet has current verification status, preserved limits,
  owner decisions, implementation-change audit, and final handoff material.
- [ ] The diff scope is limited to recovery documentation unless a separate
  implementation-change instruction exists.
