---
doc_id: TODO-HARNESS-NLSPEC
title: TODO Testing Harness Natural Language Specification
status: draft
role: recovered-nlspec-template
---

# TODO Testing Harness Natural Language Specification

## Document status

This document is a draft recovered specification for the testing harness. It is
not binding until adopted by the project maintainer under the project's normal
specification governance.

This file is a template. It must not be used as evidence of current harness
behavior, current test coverage, runtime success, or maintainer intent.

## Document role

This document defines the observable testing-harness contract after recovery
evidence has been collected. It must specify harness behavior, interfaces,
defaults, boundaries, lifecycle rules, failure behavior, resource policy,
cleanup behavior, diagnostics, schemas, source limits, and acceptance criteria.

A completed harness NLSpec must be sufficient for a competent implementer to
recreate caller-visible harness behavior for every non-source-limited
requirement without consulting implementation code.

## Template use rules

Use this template for the current Cartulary testing-harness recovery package
only with these baseline assumptions:

- S0 through S12 recovery and package stages are complete unless a later
  recovery-process document supersedes that state.
- `make` is the canonical harness command surface unless a later maintainer
  decision supersedes `MD-S7-0001`.
- Direct package scripts are developer conveniences unless they re-enter
  Make-owned wrappers.
- Generated task, schedule, Go, and TypeScript artifacts are downstream
  execution inputs when fresh. They do not own behavior and must not be
  hand-edited.
- Source limits and owner-required decisions must remain explicit. They must
  not be closed by inference.
- Acceptance rows must use `HAC-*` for binary criteria and `HAC-GAP-*` for
  known blockers. Do not introduce another acceptance ID family.
- Every final `must` or `must not` must map to exactly one `HAC-*` row or to a
  named `HAC-GAP-*`, source-limit, contradiction, or owner-decision blocker.

## Purpose and scope

TODO: Define the testing harness purpose in one or two behavior-level
paragraphs.

The harness specification owns TODO: Make command contracts, orchestration,
fixture and golden lifecycle, generated harness inputs, runtime services,
environment lifecycle, resource allocation, timing, retries, timeout,
cancellation, cleanup, diagnostics, retained artifacts, machine schemas,
provider-neutral CI validation, and harness acceptance criteria.

The harness specification must state whether each in-scope behavior is final,
source-limited, compatibility-only, future work, or owner-required.

## Non-goals

The harness specification does not define product behavior owned by Core 00
through Core 04 or by another active project owner specification.

The harness specification may define how the harness validates product behavior,
but it must not promote tests, fixtures, reset routes, reports, or local policy
to product authority.

The harness specification does not require implementation refactors unless
those refactors are explicitly classified as future work.

## Normative language and evidence boundary

Use lowercase normative words consistently:

| Term | Required meaning |
|---|---|
| `must` | Required harness contract behavior for the stated scope. |
| `must not` | Prohibited harness contract behavior for the stated scope. |
| `may` | Permitted behavior that callers cannot require. |
| `default` | Behavior when the caller omits a configurable input. |
| `source-limited` | Evidence is insufficient for a final normative claim. |
| `owner-required` | A maintainer or governing owner must decide before the claim can become final. |

Do not use advisory language such as "should", "normally", "appropriate",
"robust", or "best" for contract behavior. Replace it with a bounded
requirement, an explicit permission, or a source limit.

### Evidence status mapping

| Evidence status | May support final requirement | Required treatment |
|---|---:|---|
| `selected_runtime_observed` | yes, for the selected command, environment, inputs, artifacts, and exit status only | Record run identity, command or target, platform/tool profile, exit status, and artifact paths. |
| `runtime_observed` | conditional | Convert to `selected_runtime_observed` for durable claims or scope the claim to the recorded run. |
| `source_observed` | yes, for source-level behavior only | Cite exact source files, declarations, manifests, docs, tests, fixtures, schemas, or committed artifacts. |
| `observed` | conditional | Prefer a more precise status. Do not use as runtime-success proof. |
| `maintainer_decision` | yes | Cite the decision ID and confirm no primary owner conflict. |
| `owner_required` | no | Route to an ambiguity, source-limit, `HAC-GAP-*`, or roadmap row. |
| `maintainer_decision_required` | no | Treat as equivalent to `owner_required` for current package routing. |
| `source_limit` | no | State what was inspected, what was not inspected, the blocked claim, and follow-up needed. |
| `inferred` | no by itself | Use only to connect observed facts or describe a draft hypothesis. |
| `assumed` | no | Replace with evidence, source limit, or owner decision before finalization. |
| `contradiction` | no | Record conflicting sources, owner, blocked claim, and current treatment. |

## Authority relationship

| Surface | Owns | Does not own | May drive execution | Conflict behavior | Evidence status |
|---|---|---|---|---|---|
| Core 00 through Core 04 | Product behavior and product conformance meaning. | Harness mechanics unless an owner section says otherwise. | no | Product owner text governs product behavior. Harness text must route conflicts as authority gaps. | `source_observed` |
| `docs/domain.md` | Project vocabulary and domain-boundary interpretation. | Runtime behavior not owned by active specs. | no | Use for terminology only; do not use it to override owner specs. | `source_observed` |
| Harness NLSpec package | Adopted harness mechanics, validation orchestration, artifacts, cleanup bounds, diagnostics, and acceptance mapping. | Product semantics. | yes, only after adoption. | Later adopted harness spec text governs harness behavior unless a primary product owner conflicts. | TODO |
| `make` public surface | Canonical harness command surface. | Product behavior. | yes | Direct package scripts remain subordinate unless routed through Make. | `maintainer_decision` |
| Direct package scripts | Developer convenience entrypoints. | First-class harness authority unless they re-enter Make wrappers. | conditional | Do not promote them to canonical contracts without owner decision. | TODO |
| Generated task, schedule, Go, and TypeScript artifacts | Fresh downstream execution inputs. | Behavioral ownership. | conditional | Regenerate from owners; do not hand-edit. | TODO |
| Harness implementation, tests, and fixtures | Evidence of current behavior or validation expectations. | Normative truth when conflicting with adopted specs or owner decisions. | conditional | Treat conflicts as authority gaps. | TODO |
| CI configuration | Provider-neutral operational gate behavior when present. | Product or harness semantics not specified elsewhere. | conditional | Do not invent provider behavior for absent workflows. | TODO |
| Logs, reports, traces, screenshots, dashboards | Diagnostic or derived views by default. | Canonical pass/fail or state authority unless explicitly adopted. | no by default | Promote fields only through schema adoption or owner decision. | TODO |
| Local policy and retained local artifacts | Local execution evidence for the recorded context. | Portable harness contract by default. | no by default | Durable claims require explicit run identity. | TODO |

## Terminology

| Term | Meaning | Boundary |
|---|---|---|
| Harness | TODO: Define the orchestration and validation subsystem covered by this spec. | The definition must not include product behavior merely because tests exercise it. |
| Harness run | TODO: Define one invocation boundary, including command, environment, run identity, lifecycle, artifacts, and terminal result. | A run must not be inferred from newest retained artifacts without run identity. |
| Entrypoint | TODO: Define a supported invocation surface; for the current package this is a Make target unless adopted otherwise. | Direct scripts are not canonical unless this spec explicitly adopts them. |
| Fixture | TODO: Define committed or controlled test input state. | Fixtures must not be treated as product authority by themselves. |
| Golden or snapshot | TODO: Define expected-output artifacts and refresh authority. | Visual snapshot refresh remains source-limited unless owner bounds are supplied. |
| Generated artifact | TODO: Define derived files produced from upstream owners. | Must not be hand-edited or treated as behavior owner. |
| Retained artifact | TODO: Define run-local output kept for investigation or evidence. | Durable claims require explicit run identity. |
| Diagnostic surface | TODO: Define logs, reports, traces, screenshots, dashboards, and summaries. | Diagnostic by default unless a stable schema is adopted. |
| Service dependency | TODO: Define external runtime dependency such as database, object store, browser, server, or emulator. | Deployment conformance is out of scope unless product specs say otherwise. |
| Resource lease | TODO: Define ownership claim over ports, databases, buckets, temp dirs, browser profiles, locks, processes, or containers. | Physical capacity is not guaranteed by logical scheduler claims. |
| Cleanup tier | TODO: Define the evidence-backed cleanup claim strength. | Do not state stronger cleanup guarantees than evidence supports. |
| Failure class | TODO: Define normalized caller-visible failure categories. | `unknown_failure` must not remain in final accepted claims. |

## Harness-owned surfaces

| Surface class | Required contract content | Authority class | Evidence needed | Acceptance or blocker |
|---|---|---|---|---|
| Make entrypoints | Command, caller, mode, inputs, defaults, outputs, side effects, ordering, parallel safety, failure behavior. | canonical harness surface | TODO | TODO: `HAC-0001` |
| Direct scripts | Whether they are convenience-only, wrapper-backed, deprecated, or excluded. | subordinate unless adopted | TODO | TODO |
| Orchestration scripts | Scheduling, fanout/fanin, readiness, reporting, and cleanup behavior observable to callers. | harness mechanics | TODO | TODO |
| Fixtures and goldens | Ownership, update workflow, absence/staleness behavior, and review expectations. | validation input | TODO | TODO |
| Generated artifacts | Upstream owner, generation command, freshness rule, execution effect, and drift behavior. | downstream execution input | TODO | TODO |
| Retained artifacts | Identity fields, authority status, retention, freshness, and investigation limits. | evidence or diagnostic | TODO | TODO |
| Reports and schemas | Stable fields, schema status, null/omission semantics, ordering, and tool-owned exclusions. | stable only when adopted | TODO | TODO |
| Services | Provision, ready, reset, stop, cleanup, failure, and shared-state behavior. | harness runtime dependency | TODO | TODO |
| Resource policy | Allocation, collision, release, reuse, and parallel-safety behavior. | harness scheduler/resource rule | TODO | TODO |
| Cleanup | Trigger, order, idempotence, tier, destructive proof gates, retained diagnostics, and non-guarantees. | harness safety rule | TODO | TODO |
| CI | Provider-neutral entrypoints, absent-provider limits, summaries, and release gate relationship. | operational gate | TODO | TODO |

## Command contracts

### Entrypoint contract table

| Entrypoint ID | Make target or command | Caller | Mode | Inputs | Defaults and omitted cases | Outputs | Side effects | Ordering dependencies | Parallel safety | Failure class | Evidence | Acceptance |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| EP-0001 | `TODO` | `developer/ci/release/agent/other` | `test/check/service_start/service_stop/cleanup/fixture_update/debug/other` | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO | `HAC-0001` |

### Command mode mapping

| Mode | Allowed side effects | Disallowed side effects | Default cleanup | Failure behavior | Acceptance or blocker |
|---|---|---|---|---|---|
| `test` | TODO | TODO | TODO | TODO | TODO |
| `check` | TODO | TODO | TODO | TODO | TODO |
| `service_start` | TODO | TODO | TODO | TODO | TODO |
| `service_stop` | TODO | TODO | TODO | TODO | TODO |
| `cleanup` | TODO | TODO | TODO | TODO | TODO |
| `fixture_update` | TODO | TODO | TODO | TODO | TODO |
| `debug` | TODO | TODO | TODO | TODO | TODO |

### Command contract rules

TODO: Define command-level invariants.

Every command contract must state:

- valid callers and invocation modes;
- positional arguments, flags, environment variables, config files, and default
  behavior when each input is omitted;
- invalid-input behavior and the resulting failure class;
- stdout, stderr, exit code, report, and artifact behavior;
- side effects on fixtures, generated artifacts, services, temp paths, retained
  artifacts, and external state;
- ordering dependencies and legal concurrent invocation behavior;
- cleanup behavior on success, failure, timeout, and cancellation;
- evidence status and `HAC-*` or blocker mapping.

## Run lifecycle

### Lifecycle states

| State | Meaning | Allowed next states | Required artifacts | Caller-visible status | Evidence | Acceptance |
|---|---|---|---|---|---|---|
| `TODO` | TODO | TODO | TODO | TODO | TODO | TODO |

### Lifecycle phases

| Phase order | Phase | Owner | Inputs | Defaults and skipped cases | Outputs | Side effects | Failure class | Cleanup behavior | Evidence | Acceptance |
|---:|---|---|---|---|---|---|---|---|---|---|
| 1 | `TODO` | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO |

### Lifecycle algorithm

Use deterministic pseudocode when ordering, fanout, cleanup, report aggregation,
or partial-completion behavior cannot be expressed unambiguously in prose.

```text
Algorithm: TODO_harness_run_lifecycle

Input:
  command: TODO
  environment: TODO
  config: TODO

Preconditions:
  TODO

Steps:
  1. TODO

Output:
  terminal_state: TODO
  exit_behavior: TODO
  retained_artifacts: TODO

Errors:
  - TODO_failure_class: TODO caller-visible behavior, cleanup behavior, retryability.

Acceptance:
  - HAC-0001: TODO binary pass condition.
```

## Environment and platform

### Environment variable contract

| Variable | Owner | Type or valid values | Default when unset | Invalid value behavior | Precedence status | Evidence | Acceptance or blocker |
|---|---|---|---|---|---|---|---|
| `TODO_ENV_VAR` | TODO | TODO | TODO | TODO | `TODO: precedence_unknown/source_limit/defined` | TODO | TODO |

Only source-observed variables and defaults may become final requirements.
Cross-layer precedence must remain `TODO: precedence_unknown` or source-limited
until an owner supplies a precedence matrix or decision.

### Platform boundary

| Platform or tool surface | Supported or observed status | Required tool/version | Default behavior when unavailable | Failure class | Source limit | Acceptance or blocker |
|---|---|---|---|---|---|---|
| TODO | TODO | TODO | TODO | TODO | TODO | TODO |

Do not claim a full non-Linux, browser, provider-CI, or missing-tool support
matrix unless the claim is backed by owner decision or selected evidence.

## Service lifecycle

| Service ID | Service or environment | Provision owner | Start command or path | Ready condition | Poll interval | Ready timeout | Scope | Shared resources | Reset rule | Stop rule | Cleanup tier | Failure behavior | Evidence | Acceptance |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| SVC-0001 | TODO | TODO | TODO | TODO | TODO | TODO | `per_test/per_file/per_worker/per_suite/per_run/global/unknown` | TODO | TODO | TODO | TODO | TODO | TODO | TODO |

Every service must have provisioning, readiness, reset, stop, cleanup, and
failure behavior or an explicit source-limit row.

## Scheduler, resources, and concurrency

### Scheduler model

TODO: Define logical scheduler lanes, claims, fanout/fanin rules, duration
baseline use, and non-guarantees.

Scheduler lanes and resource claims are logical scheduling constraints unless
this spec explicitly states otherwise. They must not be described as physical
capacity guarantees for host CPUs, Docker, Postgres, object-store, browser processes,
network ports, or external services without owner evidence.

### Resource allocation table

| Resource ID | Resource type | Allocation rule | Scope | Collision detection | Collision behavior | Release rule | Reuse allowed | Parallel-safe | Evidence | Acceptance |
|---|---|---|---|---|---|---|---|---|---|---|
| RES-0001 | `port/temp_dir/database/browser_profile/worker/process/container/cache/file_lock/socket/other` | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO |

### Concurrency model

TODO: Define legal concurrent invocations, nested invocation behavior, shared
mutable resources, worker ownership, lock behavior, report fan-in ordering, and
partial-failure behavior.

## Timing, retry, timeout, and cancellation

### Timing and timeout table

| Timing surface | Unit | Default | Minimum | Maximum | Omitted-case behavior | Timeout result | Partial artifact behavior | Evidence | Acceptance or blocker |
|---|---|---|---|---|---|---|---|---|---|
| TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO |

Every timeout must state unit, default, minimum or omitted lower bound, maximum,
failure class, caller-visible result, cleanup behavior, and partial-artifact
behavior.

### Retry algorithm

```text
Algorithm: TODO_retry_policy

Input:
  failure_class: TODO
  attempt_number: TODO
  entrypoint_id: TODO

Preconditions:
  TODO

Steps:
  1. TODO

Output:
  retry_decision: retry | do_not_retry
  next_delay: TODO

Exhaustion result:
  TODO failure class, report behavior, cleanup behavior, and exit behavior.

Acceptance:
  - HAC-0001: TODO binary pass condition.
```

### Cancellation behavior

| Cancellation trigger | Detection rule | Shutdown sequence | Signal behavior | Partial artifact behavior | Cleanup tier | Failure class | Evidence | Acceptance or blocker |
|---|---|---|---|---|---|---|---|---|
| Caller interrupt | TODO | TODO | TODO | TODO | TODO | `cancelled` | TODO | TODO |
| CI cancellation | TODO | TODO | TODO | TODO | TODO | `cancelled` | TODO | TODO |
| Timeout | TODO | TODO | TODO | TODO | TODO | `timeout` | TODO | TODO |
| Process crash | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO |
| Service teardown failure | TODO | TODO | TODO | TODO | TODO | `cleanup_error` | TODO | TODO |

Parent-death cleanup, active DB cleanup, and detached reaper hard completion
must remain source-limited unless selected evidence or owner decision supports a
stronger claim.

## Fixture, golden, snapshot, and artifact lifecycle

### Artifact authority classes

| Authority class | Required meaning | May drive execution | Update authority | Absence or staleness behavior | Evidence needed |
|---|---|---|---|---|---|
| `canonical_fixture` | TODO | TODO | TODO | TODO | TODO |
| `canonical_runtime_state` | TODO | TODO | TODO | TODO | TODO |
| `derived_report` | TODO | TODO | TODO | TODO | TODO |
| `ephemeral_scratch` | TODO | TODO | TODO | TODO | TODO |
| `diagnostic_log` | TODO | TODO | TODO | TODO | TODO |
| `external_state` | TODO | TODO | TODO | TODO | TODO |
| `unknown_authority` | TODO | no | owner-required | TODO | TODO |

### Artifact lifecycle table

| Artifact ID | Artifact class or path | Authority class | Created by | Read by | Mutated by | Persisted where | Committed/ignored/uploaded | Cleanup owner | Cleanup trigger | Absence/staleness behavior | Evidence | Acceptance or blocker |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| ART-0001 | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO |

### Fixture and golden update workflow

Fixture, golden, and snapshot updates must use an owner-reviewable workflow:

1. Record explicit owner intent.
2. Record the targeted file list.
3. Cite source evidence, selected runtime evidence, or the test reason.
4. Run and record the targeted verification command.
5. Add a review note explaining why the change is expected.

Visual snapshots are validation-only until an owner supplies exact OS, browser,
version, and update-command bounds.

### Missing item review

| Artifact family | Missing items found | Evidence | Required follow-up | Acceptance or blocker |
|---|---|---|---|---|
| Fixtures | `yes/no/TODO` | TODO | TODO | TODO |
| Goldens | `yes/no/TODO` | TODO | TODO | TODO |
| Visual snapshots | `yes/no/TODO` | TODO | TODO | `HAC-GAP-0001` |

If no missing fixture, golden, or snapshot item is identified, this section must
include an explicit reviewed absence statement with evidence scope.

## Retained artifacts and diagnostics

### Retained artifact identity

Durable claims based on retained artifacts require explicit run identity:
`RESULTS_DIR`, `RUN_ID`, command or target, platform/tool profile, exit status,
and artifact paths. Newest-artifact fallback may be used only for human
investigation unless adopted otherwise.

| Claim type | Required identity fields | May support final requirement | Missing identity behavior | Evidence | Acceptance or blocker |
|---|---|---:|---|---|---|
| TODO | TODO | TODO | TODO | TODO | TODO |

### Diagnostic surfaces

| Diagnostic surface | Produced by | Consumer | Machine-consumed | Authority class | Schema or format | Ordering guarantee | Retention | Evidence | Acceptance or blocker |
|---|---|---|---|---|---|---|---|---|---|
| TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO |

Derived diagnostic surfaces must not drive execution unless this section
explicitly designates them as canonical for a named purpose.

## Machine-readable schemas

### Schema status mapping

| Schema status | Required treatment |
|---|---|
| `stable` | Field names, types, required/default/conditional fields, ordering, null/omission semantics, and compatibility rules may be final requirements. |
| `partial` | Only inspected fields may be specified; unknown fields or cases must remain source-limited. |
| `schema_unknown` | Treat as diagnostic or tool-owned. Do not specify field contracts. |
| `authority_unknown` | Route to owner decision before field contracts become final. |
| `tool_owned` | Reference as diagnostic output only unless the harness adopts the schema. |

### Schema contract table

| Schema ID or surface | Status | Producer | Consumer | Required fields | Optional fields | Ordering | Null or omission semantics | Compatibility rule | Evidence | Acceptance or blocker |
|---|---|---|---|---|---|---|---|---|---|---|
| TODO | `stable/partial/schema_unknown/authority_unknown/tool_owned` | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO |

Playwright report, trace, video, and screenshot internals must remain
tool-owned or `schema_unknown` unless a later owner adopts stable harness
schemas.

## Failure taxonomy and error behavior

| Failure class | Required meaning | Trigger phase | Caller-visible result | Exit/report behavior | Cleanup behavior | Retryable | Owner | Evidence | Acceptance or blocker |
|---|---|---|---|---|---|---|---|---|---|
| `usage_error` | Caller passed invalid command, flag, path, environment, mode, or unsupported combination. | TODO | TODO | TODO | TODO | no | TODO | TODO | TODO |
| `configuration_error` | Required harness configuration is missing, invalid, or contradictory. | TODO | TODO | TODO | TODO | no | TODO | TODO | TODO |
| `preflight_error` | Harness cannot start because dependencies, secrets, platform, or prerequisites are missing. | TODO | TODO | TODO | TODO | conditional | TODO | TODO | TODO |
| `service_start_error` | Required service cannot be provisioned or started. | TODO | TODO | TODO | TODO | conditional | TODO | TODO | TODO |
| `service_readiness_timeout` | Service started but never satisfied the ready condition. | TODO | TODO | TODO | TODO | conditional | TODO | TODO | TODO |
| `fixture_error` | Fixture is missing, malformed, stale, or unauthorized to mutate. | TODO | TODO | TODO | TODO | no | TODO | TODO | TODO |
| `resource_conflict` | Port, temp path, lock, database, browser profile, worker slot, or cache conflicts. | TODO | TODO | TODO | TODO | conditional | TODO | TODO | TODO |
| `test_assertion_failure` | Product-under-test assertion failed after harness setup reached the test execution phase. | TODO | TODO | TODO | TODO | no | TODO | TODO | TODO |
| `harness_internal_error` | Harness code failed independently of a product assertion. | TODO | TODO | TODO | TODO | no | TODO | TODO | TODO |
| `timeout` | Configured timeout elapsed. | TODO | TODO | TODO | TODO | conditional | TODO | TODO | TODO |
| `cancelled` | Caller, CI, or controlling process cancelled execution. | TODO | TODO | TODO | TODO | no | TODO | TODO | TODO |
| `cleanup_error` | Cleanup failed or left ambiguous state. | TODO | TODO | TODO | TODO | conditional | TODO | TODO | TODO |
| `unsupported_platform` | Current platform cannot satisfy the declared harness contract. | TODO | TODO | TODO | TODO | no | TODO | TODO | TODO |
| `missing_secret` | Required secret or credential is unavailable. | TODO | TODO | TODO | TODO | no | TODO | TODO | TODO |
| `manifest_or_accounting_mismatch` | Harness manifests, generated task surfaces, schemas, or accounting reports disagree. | TODO | TODO | TODO | TODO | no | TODO | TODO | TODO |
| `authority_required` | A failure cannot be classified normatively until an owner decides the contract. | TODO | TODO | TODO | TODO | no | TODO | TODO | `HAC-GAP-0001` |
| `unknown_failure` | Failure was observed but not classified. This class must not remain in a final accepted claim. | TODO | TODO | TODO | TODO | no | TODO | TODO | `HAC-GAP-0001` |

## Cleanup and destructive safety

### Cleanup tier mapping

| Cleanup tier | Allowed claim strength |
|---|---|
| `observed_successful_cleanup` | Selected runtime evidence shows a specific cleanup path completed successfully. |
| `observed_cleanup_scheduling` | Source or selected evidence shows cleanup was scheduled, but completion is not proven. |
| `delayed_after_state_evidence` | A later observation saw the expected after-state, without proving synchronous cleanup completion. |
| `best_effort_cleanup` | Source attempts cleanup, but failures, interrupts, or detached completion are not guaranteed. |
| `source_limited_cleanup` | Source exists or routing is known, but selected runtime evidence or owner decision is missing. |

### Cleanup behavior table

| Cleanup trigger | Cleanup action | Scope | Required order | Idempotent | Cleanup tier | Runs on success | Runs on failure | Runs on timeout | Runs on interrupt | Failure behavior | Evidence | Acceptance or blocker |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO |

### Destructive proof gates

| Destructive action | Required proof gates | Prohibited when | Failure behavior | Evidence | Acceptance or blocker |
|---|---|---|---|---|---|
| Delete stale database | Generated resource name; harness metadata or lease evidence; conservative age or completed-summary check; scope-limited resource type. | TODO | TODO | TODO | TODO |
| Delete stale bucket | Generated resource name; harness metadata or lease evidence; conservative age or completed-summary check; scope-limited resource type. | TODO | TODO | TODO | TODO |
| Delete stale container | Generated resource name; harness metadata or lease evidence; conservative age or completed-summary check; scope-limited resource type. | TODO | TODO | TODO | TODO |

## CI, release, and provider boundaries

| CI or release surface | Make entrypoint | Provider-specific assumptions allowed | Outputs | Failure behavior | Source limit | Acceptance or blocker |
|---|---|---|---|---|---|---|
| Provider-neutral CI | `make ci` | no, unless provider workflow source is present and adopted | TODO | TODO | TODO | TODO |
| Release verification | `make release-check` | no, unless provider workflow source is present and adopted | TODO | TODO | TODO | TODO |
| Provider annotations/uploads | TODO | owner-required unless workflow source is present | TODO | TODO | TODO | `HAC-GAP-0001` |

Repository documentation must not invent `.github/**` workflow behavior,
provider annotations, upload paths, dashboard semantics, or release readiness
claims while provider workflow sources or selected release evidence are absent.

## Product-spec validation relationship

Core 00 through Core 04 own product behavior. This harness specification owns
only the validation mechanics named in this section.

| Product requirement or surface | Product owner | Harness validation surface | Harness-owned behavior | Authority gap | Evidence | Acceptance or blocker |
|---|---|---|---|---|---|---|
| TODO | TODO | TODO | TODO | TODO | TODO | TODO |

If a harness fixture, test, reset route, or report encodes product behavior not
present in the owner specification, this spec must record an authority gap and
must not promote that behavior to harness normativity by itself.

## Preservation classification and roadmap routing

| Behavior ID | Behavior | Classification | Current treatment | Future action | Owner or evidence needed | Acceptance or blocker |
|---|---|---|---|---|---|---|
| PRES-0001 | TODO | `preserve/preserve_with_clarification/refactor_preserving_behavior/deprecate/redesign_required/remove_if_unused/authority_decision_required/exclude_from_contract` | TODO | TODO | TODO | TODO |

Future work must stay separate from recovered specification claims. A roadmap
item must not be written as completed implementation unless selected evidence
or owner decision supports that status.

## Acceptance criteria

Acceptance criteria live in the harness acceptance matrix. Inline references in
this document must use the same IDs.

| Criterion ID | Requirement | Source section | Validation type | Validation command or evidence | Fixture required | Expected result | Fail condition | Current coverage | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|---|---|
| HAC-0001 | TODO binary requirement. | TODO | `static_inspection/unit_test/integration_test/ci_run/dry_run/golden_fixture/manual_review/owner_decision/source_limit` | TODO | `yes/no/TODO` | TODO exact pass condition. | TODO exact fail condition. | `existing/partial/missing/unknown` | TODO | TODO |
| HAC-GAP-0001 | TODO named blocked claim. | TODO | `owner_decision/source_limit` | TODO | `yes/no/TODO` | TODO exact unblock condition. | TODO exact blocker condition. | `blocked` | `source_limit/owner_required` | TODO |

Acceptance rules:

- One criterion tests one requirement.
- A final `must` or `must not` without an acceptance row or blocker row is a
  specification defect.
- Runtime success may be cited only with selected runtime evidence.
- Existing tests may be cited only after inspection.
- Passing tests do not replace the requirement text; the criterion must state
  the exact contract being verified.

## Open ambiguities and owner decisions

| ID | Surface | Ambiguity type | Conflicting or missing facts | Blocked claim | Required owner | Current treatment | Decision prompt | Evidence | Evidence status | Resolution status |
|---|---|---|---|---|---|---|---|---|---|---|
| AMB-0001 | TODO | `missing_default/contradiction/implicit_sequence/resource_owner_unknown/authority_gap/schema_unknown/other` | TODO | TODO | TODO | TODO | TODO | TODO | `owner_required/source_limit/contradiction` | `open/resolved/deferred` |

Contradictions must not be reconciled silently. Record the conflict, the owner,
the blocked claim, and the current non-normative treatment.

## Source limits

| Source-limit ID | Surface | Limit type | What was inspected | What was not inspected | Blocked claim | Impact | Follow-up needed | Evidence status | Acceptance or blocker |
|---|---|---|---|---|---|---|---|---|---|
| SL-0001 | TODO | `inaccessible_file/unavailable_runtime/missing_secret/ci_only/platform_specific/partial_search/unexecuted_command/timeout/other` | TODO | TODO | TODO | TODO | TODO | `source_limit` | `HAC-GAP-0001` |

Preserve these boundaries unless later selected evidence or owner decision
closes them:

| Boundary | Required treatment |
|---|---|
| Environment-variable precedence | State only source-observed variables and defaults. Keep precedence as `TODO: precedence_unknown` or source-limited. |
| Visual snapshot refresh OS, browser, version, and command | State that visual snapshots are validation-only. Do not adopt a snapshot update command or platform bounds. |
| Parent-death cleanup | Do not claim guaranteed abrupt-exit cleanup. |
| Active DB cleanup | Do not claim guaranteed live-connection cleanup. |
| Detached reaper hard completion | Treat cleanup scheduling evidence as scheduling evidence, not completion proof. |
| Provider-specific CI while provider workflows are absent | Keep CI provider-neutral. Do not invent workflow, annotation, upload, or dashboard behavior. |
| Playwright report, trace, video, and screenshot internals | Treat as tool-owned or `schema_unknown` unless a later owner adopts stable harness schemas. |
| Release readiness beyond recorded evidence | Keep readiness claims separate from stale-smoke demotion unless the named release gate passes or each failure is classified. |

## Definition of done

The completed harness NLSpec is reviewable only when all conditions are true:

1. Every externally observable in-scope harness behavior is specified,
   intentionally excluded, or marked with a source limit or owner decision.
2. Every command has inputs, defaults or omitted-case behavior, outputs, side
   effects, ordering, parallel-safety, failure behavior, evidence, and
   acceptance mapping.
3. Every generated artifact, retained artifact, log, report, fixture, golden,
   and snapshot has authority, lifecycle, absence/staleness, and cleanup
   treatment.
4. Every service has provisioning, readiness, reset, stop, cleanup, and failure
   behavior or a source-limit row.
5. Every timeout, retry, resource allocation, cancellation path, and cleanup
   guarantee has explicit bounds or a source-limit row.
6. Environment precedence, visual snapshot refresh bounds, provider-specific CI,
   Playwright internals, and cleanup non-guarantees remain source-limited until
   later evidence or owner decision closes them.
7. No product behavior is redefined under the harness spec.
8. No derived diagnostic surface is silently promoted to authority.
9. Every final `must` or `must not` maps to one `HAC-*` criterion or to a named
   blocker.
10. Open contradictions have a named owner, decision prompt, blocked claim, and
    current treatment.
