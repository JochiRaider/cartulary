# jobs-platform Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

This tracker is the controlling execution and handoff artifact for the
authorized Jobs Platform remediation. Normative terms such as `MUST`, `MUST
NOT`, `SHOULD`, and `MAY` describe the required projection into adopted owners,
contracts, implementation, and evidence; they do not elevate this tracker above
an adopted owner document. If this tracker and an adopted owner conflict, the
owner controls and the implementation work is `BLOCKED: owner contradiction`.

| Item | Required posture |
| --- | --- |
| Target path | `internal/platform/jobs` |
| Target label | `jobs-platform`; lowercase kebab case with no spaces, separators, shell metacharacters, or unsafe filename characters |
| Output path | `docs/handoffs/jobs-platform-module-refactor-tracker.md` |
| Original repository baseline | Branch `main`, commit `7a39756616f12d06b5d1c981295acc35e517cc1d`; the worktree was clean before the original tracker was created |
| Authorized implementation baseline | Branch `main`, commit `2e80c849048e012867dc1ff860102ff1a99ca69a`; worktree clean at the 2026-08-08 execution bootstrap |
| Current task | Implement the complete Jobs Platform remediation through JS-07 in the strict sequence in Section 7 |
| Permitted change | Adopted owner specifications, authored contracts and verification inputs, Jobs and affected owner implementations/tests/composition, authored migration and boundary policy, generator-produced outputs, and this tracker, limited to the active slice |
| Prohibited changes | Unrelated behavior, public job protocol shape changes, hand edits to generated artifacts, mixed v1/v2 runtime support, dual writers, inferred historical progress repair, and automatic database reset/reseed |
| Implementation status | JS-01 through JS-07 are `DONE` in the required sequence; repository implementation, deployment-faithful cutover rehearsal, validation, and handoff are complete |

### Controlling checkpoint protocol

The only valid execution sequence is `JS-01` → `JS-06A` → `JS-06B` →
`JS-06C` → `JS-06D` → `JS-06E` → `JS-02` → `JS-03` → `JS-04` →
`JS-05` → `JS-07`.

While execution is incomplete, exactly one slice is `IN_PROGRESS`. The active
slice is implemented and validated in isolation; changed files, commands,
result roots, failures, rollback posture, and remaining blockers are recorded
here. A slice is marked `DONE` only after all binary exit criteria pass, or
`BLOCKED` when they cannot pass. The tracker is then validated with
`make lint-markdown` before the next slice becomes `IN_PROGRESS`. After JS-07
is `DONE`, no slice remains active. Planned or expected results never count as
completion evidence. JS-07 updates JP-AC-101 through JP-AC-110 from actual
retained evidence only.

### Identity map

| Identity | Namespace | Binding meaning | Must not be confused with |
| --- | --- | --- | --- |
| `platform_jobs` | Core 01 implementation refinement and schema ownership | Shared background-job storage and lifecycle substrate | Harness owner IDs or public HTTP ownership |
| `platform.jobs` | Testing Harness semantic owner | Generic jobs lifecycle, progress, durable execution, recovery, leases, attempts, and jobs telemetry evidence | The package name or every test physically stored below `internal/platform/jobs` |
| `module.jobapi` | Module and Testing Harness owner | Public job HTTP routes, request-time authorization, not-found masking, CSRF, and response mapping | Jobs-table persistence or durable worker execution |

`JP-REQ-001`: The implementation MUST retain `internal/platform/jobs` as the
shared jobs-platform mechanism boundary unless a later adopted owner changes
that assignment. Package existence alone is not authority; Core 01's explicit
`platform_jobs` refinement supplies the owner basis.

`JP-REQ-002`: Public job routes, envelopes, status tokens, authorization
outcomes, WebSocket members, and the public `{completed,total}` progress shape
MUST remain unchanged. The internal `progress_unit_id` defined by this tracker
MUST NOT enter `jobs.Resource`, HTTP, WebSocket, OpenAPI, TypeScript, frontend,
logs, or telemetry surfaces.

`JP-REQ-003`: Core 05 remains inapplicable because this work concerns
implementation conformance rather than timed or fixture-sensitive public claim
publication.

The authority order is:

1. Adopted subsystem NLSpecs within their named scopes.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. Domain vocabulary and implementation-support guides.
5. Current repository code and tests.
6. This tracker, and the planning framework as evidence
   and handoff material only.

No owner contradiction was found. The original RB-001 through RB-003 questions
are resolved in Sections 4 and 11. They are not implementation-complete.

Owner, authoring, and support material inspected:

- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/extension-subsystem-nlspec.md`
- `docs/reporting-subsystem-nlspec.md`
- `docs/report-composition-nlspec.md`
- `docs/opentelemetry-instrumentation-nlspec.md`
- `docs/testing-harness-nlspec.md`
- `docs/domain.md`
- `docs/guides/cartulary-dev-guide.md`
- `docs/guides/cartulary_implementation_testing_guide.md`
- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `docs/research/nlspec-spec.md`

## 2. Current-State Repository Inventory

### Target files

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/platform/jobs/.gitkeep` | Removed obsolete placeholder from the populated package in JS-03 | None | None | None | None | None | `platform_jobs` housekeeping | none | Cleanup is complete. |
| `internal/platform/jobs/collaboration_producer.go` | Transaction-scoped job creation/terminal operations and canonical incident `job_progress` intent production | `ProgressIntent`, `ProgressIntentAppender`, `TransactionService`, constructor, and transaction methods | Server and test-support composition; imports, incident bundles, reporting, reference data, and extension finalizers | `pgx`, jobs lifecycle helpers, injected intent appender | Progress persistence and rollback tests in `jobs_test.go` | WebSocket job-progress contract and generated WS catalog | `platform_jobs` producer with application-owned Collaboration translation | high | Existing consumer-owned port direction is intentional. |
| `internal/platform/jobs/durable.go` | Handler payload access, recovery selection, claims, leases, attempts, error recording, and exhaustion failure | Durable Manager methods | `Runner` and tests | PostgreSQL jobs rows, lifecycle transitions, progress intents | Runner dispatch, recovery, and exhaustion tests | Jobs migrations and recovery catalog | `platform_jobs` | critical | Claim currently acquires a lease without atomically changing queued status to running. |
| `internal/platform/jobs/extensions.go` | Extension job admission, contract lookup, terminal-result validation, and canonical encoding | `ExtensionResourceRefContract`, `ExtensionJobContract`, contract configuration/lookup, canonical helpers | Extension assembly, extension store/finalizers, extension-capable producers | Generated owner facts supplied through application composition | `extensions_test.go` and extension rows | Extension job-kind fragments and generated bindings | Generic `platform_jobs` mechanism constrained by Extensions/Reporting owners | high | Future projection must carry owner-declared `progress_unit_id`. |
| `internal/platform/jobs/extensions_test.go` | Characterizes extension admission and terminal resource contracts | Two package tests | Go runner and semantic family rows | Jobs public API and fixtures | Self | Existing `module.extensions` unit row | `module.extensions` verification evidence | medium | These tests remain Extensions-owned by postcondition. |
| `internal/platform/jobs/jobs.go` | Public job DTOs/errors; Manager; creation, lookup, running/terminal transitions, cancellation, idempotency, and extension cancellation observation | `Manager`, DTOs, admission/idempotency types, lifecycle methods, and `LockTransitionTx` | Job API, server assembly, modules, extension store, runner, and test support | Jobs table, Auth-owned `route_idempotency`, Extensions-owned cancellation observation, telemetry | Manager and Job API tests | OpenAPI/TypeScript job resource, migrations, ownership manifest | Primarily `platform_jobs`; two foreign-owner persistence operations are misplaced | critical | Contains direct queued terminal writes and incomplete progress enforcement. |
| `internal/platform/jobs/jobs_test.go` | Characterizes cancellation replay, terminal resource retention, progress-intent atomicity, runner dispatch/recovery, configuration, and exhaustion | Eight package tests | Go runner and semantic family rows | PostgreSQL test support and jobs API | Self | Some rows map to `module.extensions`; six tests were previously unmapped | `platform.jobs` plus existing cross-owner evidence | high | Queued recovery currently characterizes a direct queued-to-success path and must be corrected, not preserved. |
| `internal/platform/jobs/recovery_state.go` | Declares the authoritative `jobs` recovery contribution | `RecoveryStateContribution` | Recovery assembly | Recovery contribution contract | Recovery catalog/assembly tests | Recovery state catalog | `platform_jobs` | medium | Legitimate thin owner contribution. |
| `internal/platform/jobs/runner.go` | Handler registration, dequeue admission, dispatch, recovery activation, panic/error handling, and shutdown | Runner errors, `HandlerFunc`, `DequeueGate`, `Runner`, and lifecycle methods | Server assembly and module handler registration | Manager durable operations, concurrency, UUIDs | Runner tests and worker integration rows | Named portability recovery evidence | `platform_jobs` | critical | Handler invocation must occur only after a committed running claim. |
| `internal/platform/jobs/telemetry.go` | Jobs tracing, duration and active metrics, active-job query, and safe tokens | Private jobs instrumentation helpers | Manager and Runner | OpenTelemetry API and jobs table | `telemetry_test.go` | Adopted OpenTelemetry vocabulary | `platform_jobs` | high | `progress_unit_id` must never become a telemetry attribute; `cartulary.job_kind` currently reports a scope surrogate rather than the catalog-backed job kind. |
| `internal/platform/jobs/telemetry_test.go` | Characterizes vocabulary, safe tokens, and no-SDK ownership | Three package tests | Go runner | Telemetry helpers | Self | No current semantic row | `platform.jobs` verification evidence | medium | Must be routed after `platform.jobs` is registered. |

Every target entry is inventoried; none is silently excluded.

### Execution-baseline runtime and security gaps

The authorized baseline contains all of the following defects, and JS-06B
must establish regression evidence before corrected writers are activated:

- a durable claim can invoke a handler while the public job is still queued;
- runner-wide lease-owner reuse can admit duplicate dispatch within one runner;
- a nil handler return can release a lease while leaving a mutable job
  unfinished;
- raw handler errors and recovered panic values can enter persisted or public
  diagnostics;
- lease and attempt bookkeeping can change public `updated_at` and emit a
  public progress intent despite no public resource change;
- the extension reconciliation path directly updates the Jobs table;
- Jobs directly accesses Auth route-idempotency and Extensions cancellation
  observation tables;
- Imports combines the exported `LockTransitionTx` capability with direct Jobs
  queries; and
- Jobs telemetry emits an incident/deployment scope surrogate as
  `cartulary.job_kind`.

The corrected runtime uses a unique attempt identity for every dispatch,
closed safe failure tokens and fixed operator-safe summaries, a private
transition matrix, and catalog-backed actual job kinds. Raw handler errors,
panic values, job IDs, and progress-unit IDs are forbidden from public errors,
logs, and telemetry.

### Production jobs-table writer map

`JP-REQ-004`: Every production insert or update of `jobs` MUST execute through
the jobs-platform owner boundary. Tests and migrations MAY write fixtures or
perform schema transitions, but MUST NOT define runtime behavior.

| Current writer | Current write | Target treatment | Closure evidence |
| --- | --- | --- | --- |
| `internal/platform/jobs/jobs.go` | Creation, running, terminal, and cancellation updates | Keep creation and lifecycle persistence owner-local; route all state/progress changes through the central transition substrate | No lifecycle SQL exists outside the substrate; transition/progress evidence passes |
| `internal/platform/jobs/durable.go` | Lease/attempt claims, error/exhaustion failure | Keep lease persistence owner-local; combine queued claim with `queued → running`; delegate terminal failure to the transition substrate | Concurrent claim, recovery, exhaustion, and event-atomicity tests pass |
| `internal/platform/extensionstore/reconciliation.go` | Direct terminal update for inactive extension jobs | Remove the direct update and call a jobs-owned transaction operation while retaining Extensions-owned proof/observation validation | Source scan finds no `UPDATE jobs` outside jobs-platform; reconciliation evidence passes |
| `internal/platform/postgres/extension_job_cutover_migration_test.go` | Test fixture insert | Keep test-only; it does not authorize a production writer | Boundary scan distinguishes test fixtures from production writes |
| `internal/modules/reportcomposition/release_tuple_facade_test.go` | Test fixture insert | Keep test-only; it does not authorize a production writer | Boundary scan distinguishes test fixtures from production writes |

## 3. Module Boundary Diagnosis

The target is a legitimate persistence-adjacent platform service, durable
runner, and mutation coordinator. It is not a public transport owner, domain
catch-all, frontend controller, projection layer, saved-view owner, view-schema
adapter, or grid-vendor integration package.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Normative disposition |
| --- | --- | --- | --- | --- | --- |
| Shared job resource, lifecycle, retention, persistence, progress, and durable execution | Jobs target | `platform_jobs` | keep | Core 01, jobs migrations, recovery/schema manifests | MUST remain behind one jobs-platform facade and writer boundary. |
| Generic jobs evidence | Existing rows plus unmapped target tests | `platform.jobs` | move/reroute | Harness grammar and postcondition ownership | MUST use the new active owner without taking Job API or extension-specific rows. |
| Public job HTTP transport and authorization | `internal/modules/jobapi/routes.go` | `module.jobapi` | keep | Core 01/Core 04, OpenAPI owner, route tests | MUST remain outside jobs-platform. |
| Progress intent production and Collaboration translation | Jobs producer plus application adapters | Jobs produces; application translates; Collaboration persists/delivers | keep | Core 01, intent adapters, boundary policy | MUST retain the injected appender and transaction rollback behavior. |
| Route-scoped idempotency persistence | Jobs cancellation SQL | Auth owner behind a jobs-consumer port | move | Schema ownership manifest | MUST move after lifecycle correction without changing replay/conflict behavior. |
| Extension cancellation observation | Jobs cancellation SQL | Extensions owner behind a jobs-consumer port | move | Extensions NLSpec and migration | MUST move after lifecycle correction while remaining in the cancellation transaction. |
| Inactive-extension terminal job update | Extensions reconciliation | Jobs owner invoked through a transaction operation | move | Direct `UPDATE jobs` inspection | MUST migrate during JS-06E before the exclusive-writer requirement closes. |
| Extension job admission and canonical result validation | Jobs extension mechanism plus owner catalogs | Jobs mechanism; named job-kind owners define semantics | keep | Core 01 and adopted profile owners | MUST gain owner-declared progress-unit projection without making Jobs interpret the unit. |
| Public DTOs, validation, SQL, and foreign side effects in one large file | `jobs.go` | Private cohesive jobs implementation behind existing facade | split | Direct source inspection | MUST split only after corrections and foreign-owner port extraction. |
| Frontend polling/replay consumption | `apps/web` | Web owners | defer | Consumer inspection | MUST remain contract-consumer-only for this refactor. |
| Timeline, projections, saved views, view schema, entities, indicators, evidence, links, grid adapter, and UI selectors | Not present in target | Existing semantic owners | defer | Import, SQL, and caller scan | MUST NOT be introduced into jobs-platform. |

`JP-REQ-005`: Application composition MUST remain the only edge that assembles
owner catalogs, Jobs operations, Auth/Extensions adapters, and Collaboration
intent translation. Jobs MUST NOT discover peer owners or import peer internals.

## 4. Public Contract and Behavior Freeze Map

### Observable contract map

| Contract | Current owner | Evidence | Existing tests | Required evidence | Refactor risk | Normative disposition |
| --- | --- | --- | --- | --- | --- | --- |
| `GET /api/v1/jobs/{job_id}` and `POST /api/v1/jobs/{job_id}:cancel` | Core 01; `module.jobapi` transport | Core 01 §§3.3.9-3.3.9.1, Job API routes, OpenAPI source | Three Job API service-backed rows | Preserve paths, envelopes, CSRF, authorization, replay/conflict, and not-found masking | critical | No route or public error-code change is permitted. |
| Six job states and transitions | Core 01 | Core 01 transition matrix; Core 04 AC-257 | Manager, runner, and worker tests | Exhaustive matrix, concurrency, recovery, cancellation race, HTTP and WS observation | critical | Repository behavior must be corrected to the matrix below. |
| Progress and terminal equality | Core 01 | Core 01 progress rules; Core 04 AC-258 | Partial Manager/worker coverage | Full matrix, concurrency, success, recovery, HTTP/WS parity, and extension coverage | critical | Internal unit identity closes durable semantic ambiguity without changing wire shape. |
| Cancellation idempotency and observation atomicity | Core 01, Auth storage owner, Extensions owner | Jobs cancellation SQL, owner manifests, Extensions requirements | Manager cancellation, Job API, and reconciliation rows | First use, exact replay, conflict, absent/unauthorized job, rollback, and success-wins | critical | Port extraction must preserve one cancellation transaction. |
| `job_progress` authorization, replay, and delivery | Core 01 and Collaboration | REQ-01-268, WS source/projection, app translator | Progress persistence/rollback and Collaboration rows | Legal sequences only, rejected update emits nothing, HTTP/live/replay parity | critical | Jobs must not gain a direct Collaboration dependency. |
| Durable registration, gate, lease, attempts, recovery, panic, and close | Core 01 and jobs platform | Runner/durable sources and stage-6 gate | Dispatch, recovery, named configuration, and exhaustion tests | Gate, contention, expiry, panic, missing handler, exhaustion, idempotent close | critical | Handler execution must follow a committed claim. |
| Extension job admission, proof, cancellation, reconciliation, and canonical terminal result | Core 01 plus named profile owners | Extension fragments, Extensions/Reporting owners | Extensions unit and service-backed rows | All ten job kinds have units; proof/observation ordering and finalization remain atomic | high | Profile semantics stay owner-defined. |
| Jobs storage, recovery, and retention | Core 01 and `platform_jobs` | Migrations, schema ownership, recovery catalog | Integration and recovery checks | Migration preflight/backfill/constraints and seven-day retention | critical | Schema changes are authored migrations; generated projections are never hand-edited. |
| Jobs telemetry | OpenTelemetry instrumentation owner | Adopted jobs instrumentation section | Three target telemetry tests | `platform.jobs` routing and no new unsafe attributes | medium | Job IDs, incident content, payloads, and progress unit IDs remain forbidden. |
| OpenAPI, WebSocket, Go, and TypeScript projections | Source contract owners | OpenAPI/WS sources and generated outputs | Drift checks | Drift-clean after any authored source change | high | Public projections do not gain `progress_unit_id`. |
| Frontend polling/replay | Core 03 and Web owners | Import coordinator, import panel, collaboration session | Existing frontend rows | Stateful browser gate only if observable surface changes | medium | Clients may observe previously skipped legal states; no new state token exists. |
| Harness accounting | Testing Harness NLSpec | Owner registries, contracts, and test families | Partial Extensions/Job API routing | Active `platform.jobs` owner and complete owner slices | high | Evidence routing does not define runtime architecture. |

### State transition contract

`JP-REQ-006`: The legal transition matrix is closed and MUST be implemented
exactly as follows.

| Current state | Allowed next states | All other requests |
| --- | --- | --- |
| `queued` | `running`, `cancel_requested`, `failed` | Reject without row mutation or event |
| `running` | `cancel_requested`, `succeeded`, `failed` | Reject without row mutation or event |
| `cancel_requested` | `canceled`, `failed`, `succeeded` | Reject without row mutation or event |
| `succeeded` | none | Reject as terminal immutable |
| `failed` | none | Reject as terminal immutable |
| `canceled` | none | Reject as terminal immutable |

`queued → succeeded` and `queued → canceled` are prohibited. `queued → failed`
remains legal for permanent pre-execution failures. `cancel_requested →
succeeded` remains legal only for the defined success-wins race after the
authoritative effect commits.

`JP-REQ-007`: Initial dispatch and queued recovery MUST atomically commit
`queued → running`, `started_at`, lease ownership, lease expiry, and attempt
accounting before invoking the handler. Reclaiming an expired running lease MAY
retain `running` while changing lease/attempt metadata. A handler MUST NOT run
while its durable job remains queued.

`JP-REQ-008`: A newly accepted cancellation MUST atomically commit
`queued|running → cancel_requested`, set `cancelable=false`, persist route
idempotency, persist an extension cancellation observation when applicable, and
append the canonical progress intent. A later terminal state MUST start from
`cancel_requested`. A failed transaction MUST persist none of those effects.

### Transaction interface

`JP-REQ-009`: All production state changes MUST delegate to one jobs-owned,
transaction-scoped semantic operation equivalent to:

```text
TransitionTx(
    context,
    transaction,
    job_id,
    expected_current_states,
    requested_next_state,
    optional_final_progress,
    optional_terminal_summary
) -> updated_job | typed_transition_error
```

The operation:

- MUST use the caller's transaction and MUST NOT begin, commit, or roll it back;
- MUST compare persisted current state and requested next state atomically;
- MUST use conditional update and classify a zero-row result inside the same
  transaction;
- MUST update timestamps, retention, summaries, proof coordination, lease
  release, progress, and progress intent atomically when applicable;
- MUST return a typed internal reason without exposing SQL, table names,
  constraint names, payloads, incident content, or raw job identifiers.

Narrow operations equivalent to `ClaimRunningTx`, `RequestCancelTx`,
`CompleteSucceededTx`, `CompleteFailedTx`, and `CompleteCanceledTx` MAY remain
available, but MUST delegate to this substrate and MUST NOT carry independent
state matrices.

`JP-REQ-010`: The default concurrency mechanism is an atomic conditional
`UPDATE ... WHERE status = ANY(expected_states) RETURNING ...`. Row-local
`CHECK` constraints enforce static shape. A database transition trigger MUST
NOT be added unless an authorized later task proves that bypassing writers
cannot be removed or guarded.

### Job-definition and progress-unit contract

`JP-REQ-010A`: The active runtime contract is the closed
`cartulary.extension_job_kind_contract.v2` shape. It requires
`progress_unit_id`, participates in the existing canonical digest algorithm,
and has no v1 compatibility reader. One immutable runtime job-definition
catalog is assembled from owner facts. Callers identify a job kind; Jobs
derives the unit from the catalog and never trusts a caller-supplied unit.
Internal storage uses the generic name `job_kind`, not
`extension_job_kind`.

`JP-REQ-011`: Every new job and every mutable nonterminal job MUST have one
immutable internal `progress_unit_id`. Its grammar is:

```text
semantic_segment = [a-z][a-z0-9_]{0,62}
version_segment  = "v" [1-9][0-9]*
progress_unit_id = semantic_segment "." semantic_segment
                   *("." semantic_segment) "." version_segment
```

The complete value MUST contain no more than 191 ASCII bytes. It has no default.
Omission, an invalid value, an unknown mapping, or an ambiguous mapping MUST
fail closed for new or nonterminal jobs. Jobs stores and compares the value
exactly but MUST NOT interpret it. A semantic change requires a new versioned
identifier.

The applicable job-kind owner MUST author each exact value before JS-06D. This
tracker intentionally does not infer unit meaning from visible counters,
messages, route names, results, or mutable registries.

| Job kind | Profile/semantic owner | Current authored source | Required declaration state |
| --- | --- | --- | --- |
| `import.discovery_v1` | Import contract owner (`core01` fragment) | `contracts/extensions/fragments/core01.profile-jobs.json` | `import.discovery.session.v1` |
| `import.apply_v1` | Import contract owner (`core01` fragment) | `contracts/extensions/fragments/core01.profile-jobs.json` | `import.apply.import_unit.v1` |
| `incident_portability.export_v1` | Incident portability contract owner (`core01` fragment) | `contracts/extensions/fragments/core01.profile-jobs.json` | `incident_portability.export.request.v1` |
| `incident_portability.import_v1` | Incident portability contract owner (`core01` fragment) | `contracts/extensions/fragments/core01.profile-jobs.json` | `incident_portability.import.request.v1` |
| `reference_pack.import_v1` | Reference-pack contract owner (`core01` fragment) | `contracts/extensions/fragments/core01.profile-jobs.json` | `reference_pack.import.request.v1` |
| `reference_pack.refresh_v1` | Reference-pack contract owner (`core01` fragment) | `contracts/extensions/fragments/core01.profile-jobs.json` | `reference_pack.refresh.pack_key.v1` |
| `reference_pack.reverify_v1` | Reference-pack contract owner (`core01` fragment) | `contracts/extensions/fragments/core01.profile-jobs.json` | `reference_pack.reverify.pack_version.v1` |
| `snapshot_reporting.composition_preview_v1` | Reporting | `contracts/extensions/fragments/snapshot_reporting.participation.json` | `snapshot_reporting.composition_preview.render_attempt.v1` |
| `snapshot_reporting.release_create_v1` | Reporting | `contracts/extensions/fragments/snapshot_reporting.participation.json` | `snapshot_reporting.release_create.render_attempt.v1` |
| `snapshot_reporting.snapshot_create_v1` | Reporting | `contracts/extensions/fragments/snapshot_reporting.participation.json` | `snapshot_reporting.snapshot_create.materialization.v1` |

The unit MUST traverse the owner-to-storage boundary as follows:

| Source or boundary | Required representation | Validation responsibility |
| --- | --- | --- |
| Adopted job-kind owner input | Required `progress_unit_id` member in each job-kind contract | Named job-kind owner defines meaning and version |
| Extensions runtime projection | Required `ProgressUnitID` on the projected `extensions.JobKindContract` | Extensions projection validates presence, grammar, and canonical digest participation |
| Application-to-Jobs projection | Required `ProgressUnitID` on `jobs.ExtensionJobContract` | Application composition copies the exact owner value without reinterpretation |
| Jobs creation input | Required internal `ProgressUnitID` in the semantic queued-job creation input | Jobs validates grammar and, for an admitted extension job, exact equality with the configured job-kind contract |
| Jobs row | Required `progress_unit_id` for new/mutable rows | Jobs persists the exact value at creation and never changes it |
| Public resource/event | No member | HTTP, WS, TypeScript protocol, and telemetry surfaces reject accidental projection |

An extension job with a missing, unknown, or mismatched catalog definition MUST
fail admission before any jobs row is inserted. A future non-extension producer
or internal caller MUST add an explicitly owner-declared catalog fact; an
admission-time fallback or caller-chosen unit is forbidden.

### Progress update contract

`JP-REQ-012`: Progress updates MUST satisfy every applicable row in this table.

| Prior value or state | Proposed value | Required result | Event behavior |
| --- | --- | --- | --- |
| `completed=N` | `< N` | Reject `completed_regressed` | Emit nothing |
| `completed=N` | `N` with otherwise identical progress and no state change | Return current resource as an idempotent no-op | Emit nothing |
| `completed=N` | `> N` | Accept when every other invariant holds | Emit one intent on commit |
| `total=null` | `null` | Accept when every other invariant holds | Follow actual row/state change |
| `total=null` | positive integer | Accept only when `completed <= total` | Emit one intent on commit |
| `total=T` | `null` | Reject `total_cleared` | Emit nothing |
| `total=T` | `< T` | Reject `total_regressed` | Emit nothing |
| `total=T` | `T` or `> T` | Accept only when unit is unchanged and `completed <= total` | Emit one intent only for an actual row/state change |
| Any unit ID | Different unit ID | Reject `progress_unit_changed` | Emit nothing |
| Known total | `completed > total` | Reject `completed_exceeds_total` | Emit nothing |
| `succeeded` with known total | `completed != total` | Reject `incomplete_success` | Emit nothing |
| `succeeded` with `total=null` | Any nonnegative completed value | Accept when transition is otherwise legal | Emit one terminal intent on commit |
| `failed` or `canceled` with known total | `completed < total` | Accept when transition is otherwise legal | Emit one terminal intent on commit |
| Any terminal job | Progress-only update | Reject `terminal_job_immutable` | Emit nothing |

JS-06D MUST add or retain named row-local constraints equivalent to:

```sql
CONSTRAINT jobs_progress_unit_required_ck CHECK (
  progress_unit_id IS NOT NULL
  OR status IN ('succeeded', 'failed', 'canceled')
)

CONSTRAINT jobs_progress_unit_shape_ck CHECK (
  progress_unit_id IS NULL
  OR (
    octet_length(progress_unit_id) <= 191
    AND progress_unit_id ~
      '^[a-z][a-z0-9_]{0,62}(\.[a-z][a-z0-9_]{0,62})+\.v[1-9][0-9]*$'
  )
)

CONSTRAINT jobs_succeeded_progress_complete_ck CHECK (
  status <> 'succeeded'
  OR progress_total IS NULL
  OR progress_completed = progress_total
)
```

The existing nonnegative-completed, positive-known-total, and
completed-not-over-total checks MUST remain. Prior-row monotonicity and legal
state transitions MUST remain in the atomic Jobs write path rather than a
row-local check.

`JP-REQ-013`: Successful completion MAY submit explicit final progress in the
terminal transaction. When total is known, the submitted or existing completed
value MUST equal total. The platform MUST NOT silently set completed to total.

`JP-REQ-014`: Job-row progress, state transition, result/error summary,
applicable extension proof, and canonical progress intent MUST commit in one
transaction. A stale or rejected update MUST overwrite nothing and emit no
intent. HTTP polling and live/replayed WebSocket resources MUST expose the same
committed progress.

`JP-REQ-015`: Internal progress/transition contract reasons include exactly the
following minimum closed set for this refactor:

| Reason | Retry posture | Public exposure |
| --- | --- | --- |
| `completed_regressed` | non-retryable contract violation | Map through an existing safe public surface; do not expose the token automatically |
| `total_cleared` | non-retryable contract violation | Same |
| `total_regressed` | non-retryable contract violation | Same |
| `progress_unit_changed` | non-retryable contract violation | Same |
| `completed_exceeds_total` | non-retryable contract violation | Same |
| `incomplete_success` | non-retryable contract violation | Same |
| `terminal_job_immutable` | non-retryable contract violation | Same |

The runner MUST record only a secret-safe diagnostic and MAY attempt a legal
failure transition only when doing so cannot conceal an already committed
authoritative effect.

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Jobs reads/writes Auth-owned `route_idempotency`. | Jobs cancellation SQL and schema ownership manifest | Cross-owner persistence and idempotency drift | `must_fix` | Auth persistence behind a Jobs consumer port | Execute JS-02 after conformance correction; preserve one cancellation transaction and exact replay/conflict behavior. |
| Jobs inserts Extensions-owned cancellation observations. | Jobs SQL, Extensions NLSpec, migration 00034 | Reconciliation drift and authority leakage | `must_fix` | Extensions persistence behind a Jobs consumer port | Execute JS-02 after conformance correction; preserve precommit observation and success-wins behavior. |
| Extensions reconciliation updates `jobs` directly. | `internal/platform/extensionstore/reconciliation.go` | Bypasses Jobs transition/progress enforcement | `must_fix` | Jobs transaction operation | Migrate in JS-06E before claiming exclusive writer closure. |
| Durable claim does not perform queued-to-running. | `durable.go` conditional lease update | Handler can run while durable state remains queued | `must_fix` | Jobs transition substrate | Correct in JS-06C/JS-06E and prove concurrent claim uniqueness. |
| `jobs.go` combines DTOs, SQL, lifecycle, Auth, and Extensions effects. | Direct source inspection | Large regression surface | `should_fix` | Private cohesive Jobs implementation | Split in JS-03 after corrections and port extraction. |
| Broad concrete Jobs values flow through application/HTTP dependency sets. | Server assembly and consumer inspection | Capability leakage | `should_fix` | Consumer-owned narrow interfaces | Migrate one consumer per JS-04 checkpoint. |
| Boundary policy does not reject all foreign storage or jobs-table bypasses. | Backend boundary policy inspection | Regression can reintroduce leaks | `should_fix` | Boundary policy owner | Add guards in JS-05 after current violations are removed. |
| Progress intent uses an injected appender and app translation. | Jobs producer and app/test adapters | Low when preserved | `intentional/no_action` | Jobs producer plus Collaboration consumer | Retain and strengthen atomicity evidence. |
| Jobs-owned SQL and recovery contribution remain in Jobs. | Target sources and recovery catalog | Expected persistence adjacency | `intentional/no_action` | `platform_jobs` | Keep private; do not introduce a generic repository abstraction without a behavior need. |
| Jobs telemetry remains API-only and safe-token bounded. | OTel owner and target tests | Low when unit ID remains absent | `intentional/no_action` | `platform_jobs` | Route tests to `platform.jobs`; add no new attribute. |
| No target dependency on timeline, projections, saved views, view schema, grid vendor, or frontend shell exists. | Target/caller scan | Scope expansion | `intentional/no_action` | Existing owners | Keep these concerns out of the refactor. |
| `.gitkeep` remained in a populated directory. | Directory inspection | No runtime risk | `should_fix` | Housekeeping | Resolved in JS-03 by deleting the obsolete placeholder. |

`JP-REQ-016`: After JS-05, boundary checks MUST reject production jobs-table
writes outside the Jobs owner, Auth/Extensions table access from Jobs, direct
Collaboration persistence from Jobs, and peer-internal imports. Owner adapters
at application composition remain permitted.

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap | root | none | WF-01 | Preserve authority, baseline, constraints, and history. | Tracker and source materials | Read-only status/source inspection | Scope and sole-write rule recorded. |
| WF-01 | Current-state inventory | chain | WF-00 | WF-02, WF-03 | Inventory target, callers, contracts, writers, tests, and generated surfaces. | Jobs target and adjacent evidence | Direct reads and source scans | Every target file and production writer accounted for. |
| WF-02 | Owner and public-contract map | chain | WF-01 | WF-03, WF-04 | Freeze public behavior and assign semantic owners. | Core/subsystem owners and contract sources | Owner/repository cross-check | No contradiction and no unowned contract. |
| WF-03 | Verification-owner closure | chain | WF-02 | WF-04 | Register `platform.jobs` and route generic mechanism evidence exactly once. | Verification registry/contract and family inputs | Owner explain/guide and complete slices | RB-001 implementation evidence complete. |
| WF-04 | Normative correction contract | chain | WF-02, WF-03 | WF-05 | Adopt minimal Core clarifications, unit declarations, and owner-aligned acceptance. | Core 01/Core 04 and job-kind owner inputs | Human owner review plus contract generation | RB-002/RB-003 requirements are authoritative before code. |
| WF-05 | Transition/progress implementation chain | chain | WF-04 | WF-06 | Add tests, central substrate, migration, and migrate every writer/worker. | Jobs, migration, writers, workers, tests | Platform/Job API/Extensions slices and migration checks | Single corrected writer version serves work. |
| WF-06 | Structural boundary refactor | chain | WF-05 | WF-07 | Extract foreign storage ports, split internals, and narrow consumers. | Jobs, owner adapters, composition, consumers | Narrow owner rows and boundary checks | Behavior remains stable after each checkpoint. |
| WF-07 | Harness and anti-drift accounting | parallel | WF-03, WF-05, WF-06 | WF-08 | Close semantic routing, generated drift, and boundary regression policy. | Harness/boundary owner inputs | Generation, drift, owner, and boundary targets | All selected rows resolve once and guards reject bypasses. |
| WF-08 | Cutover, validation, and handoff | chain | WF-05, WF-06, WF-07 | none | Drain old writers, migrate, reopen, recover, validate, and publish evidence. | Operations, validation artifacts, tracker | Final command matrix | Actual results and residual deferrals recorded. |

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| JS-01 | tracker bootstrap | Register active `platform.jobs`; add lifecycle, progress, durable-runner, recovery, and telemetry verification families; route generic Jobs tests exactly once while preserving Job API, Extensions, Collaboration, and profile ownership by postcondition. | Authored verification/catalog inputs; generated topology via generator | Misassignment by file location or recycled identity | Selector uniqueness, owner closure, existing cross-owner ownership | `make generate`; generation/policy drift; owner explain/guide; complete `platform.jobs` slices | Revert authored owner inputs and regenerate; no product code changes. | Owner is active/unique; every target test resolves exactly once; generated outputs are drift-clean. |
| JS-06A | JS-01 | Adopt minimal Core 01 transition/recovery/progress clarification, Core 04 acceptance updates, and exact unit declarations for all ten job kinds. | Core 01/Core 04 and owner fragments/contracts | Tracker must not become behavior owner; unit meaning must not be guessed | Owner review and contract validation | `make generate`; `make generate-drift`; `make generated-artifact-policy-check` | Revert owner edits as one specification checkpoint before code. | Every future behavior has one adopted owner statement and every job kind has one exact unit. |
| JS-06B | JS-06A | Add owner-aligned transition, progress, concurrency, recovery, atomicity, public-parity, and extension-unit evidence. | Jobs/Job API/Extensions/Collaboration tests and semantic rows | Tests could bless current queued terminal drift | Full matrices in Section 8 | Exact owner slices | Revert tests/rows without touching production behavior. | Tests fail against current defects for the intended reasons and contain no duplicated selectors. |
| JS-06C | JS-06B | Introduce the central atomic transition/progress substrate behind existing Jobs operations; do not activate unsupported writers yet. | Jobs lifecycle/transaction internals | Public errors, atomicity, or status behavior can drift | Unit and service-backed substrate evidence | `make test-slice OWNER=platform.jobs`; `make service-backed-test-slice OWNER=platform.jobs` | Keep old call routing until substrate evidence passes; never dual-write. | One internal state/progress matrix exists; current facade delegates where activated. |
| JS-06D | JS-06C | Add next available authored migration, currently `00058`, with `progress_unit_id`, preflight/backfill, constraints, and compatibility gates. | `db/migrations/00058_platform_jobs_state_progress_contract.sql` or next free number; job-kind projections | Unsafe normalization, unresolved units, mixed writers, invalid retained rows | Migration upgrade/drift, invalid corpus, backfill, rollback/readiness tests | `make migration-drift`; relevant platform owner slice | Close dequeue and retain rollback point before migration; no inferred repair. | Preflight is safe, all mutable rows have units, constraints validate, and unresolved nonterminal rows block startup. |
| JS-06E | JS-06D | Route durable claim/recovery, cancellation, terminal finalizers, Extensions reconciliation, reporting, imports, incident bundles, and reference data through the substrate; remove bypass writes. | Jobs, `extensionstore`, module workers/finalizers, app composition | Handler ordering, success-wins, proof/observation, and events | All correction and cross-owner rows | Platform, Job API, and Extensions service-backed slices; source/boundary scan | One writer/caller group per checkpoint; rollback before moving the next. | No production Jobs write bypass remains; queued work runs only after running claim; all evidence passes. |
| JS-02 | JS-06E | Replace direct Auth idempotency and Extensions observation SQL with transaction-scoped owner adapters. | Jobs, Auth adapter, Extensions store, server/test composition | Cancellation replay/conflict and observation atomicity | Manager cancellation, Job API authorization, reconciliation, rollback | Job API and Extensions service-backed slices; boundary check | Remove old SQL only when adapter path passes; no dual write. | Jobs contains no Auth/Extensions table name and public cancellation behavior is unchanged. |
| JS-03 | JS-02 | Split public facade/DTO, private Jobs persistence, transition validation, durable persistence, and telemetry without changing package path. | `internal/platform/jobs` | Exported surface and SQL scan behavior | All `platform.jobs` rows | Platform owner slices; `make test-fast` | Move one concern per checkpoint; no compatibility shim. | One semantic home per private concern and unchanged exported behavior. |
| JS-04 | JS-03 | Narrow concrete consumers one at a time while application composition supplies adapters. | Job API, imports, incident bundles, reporting, reference data, extension store, HTTP/server/test composition | Capability loss or broad facade recreation | Each consumer's current owner rows | Consumer `make task-guide` slice plus boundary check | One consumer per checkpoint. | Consumers receive only required Jobs capabilities. |
| JS-05 | JS-02, JS-04 | Add authored boundary guards for foreign SQL, jobs-table bypasses, and peer imports. | Backend boundary owner input; generated outputs through generator | Overbroad rules can reject owner adapters/tests | Positive owner-adapter fixtures and negative bypass fixtures | `make backend-module-boundary-check`; drift checks | Revert a rule that cannot distinguish production bypass from allowed owner/test paths. | Reintroduced bypasses fail canonically and valid composition passes. |
| JS-07 | JS-05 | Perform drained cutover, replay-horizon check, complete validation, optional `.gitkeep` cleanup, and final handoff. | Runtime operations, validation artifacts, tracker | Mixed writers, stale replay, premature cleanup, unsupported claims | Complete owner and broad evidence | `make agent-finalize`; `make check`; all Section 8 gates | Keep pre-cutover rollback point; reopen dequeue only after corrected writer readiness. | Single corrected writer serves work; all binary criteria pass; results and deferrals are recorded. |

### Slice checkpoint ledger

| Slice | Status | Changed files | Commands and result roots | Failures or blockers | Rollback posture |
| --- | --- | --- | --- | --- | --- |
| JS-01 | DONE | `contracts/verification/owners/platform.jobs.json`; `contracts/verification/registry.json`; `tools/test_families/platform.jobs.json`; `tools/test_catalog_owner.json`; generated `tools/execution_topology_render_index.json`; this tracker | `make generate` pass: `.cartulary/test-results/20260808T040638Z-p2208396`; `make generate-drift` pass: `.cartulary/test-results/20260808T040657Z-p2210760`; generated policy pass: `.cartulary/test-results/20260808T040709Z-p2213484`; owner explain/guide pass; owner slice pass: `.cartulary/test-results/20260808T040721Z-p2214373`; service-backed slice pass: `.cartulary/test-results/20260808T040734Z-p2216155` | Initial generation attempts failed on ASCII ordering at `.cartulary/test-results/20260808T040533Z-p2202673` and `.cartulary/test-results/20260808T040609Z-p2205561`; authored order corrected; no remaining blocker | Revert the four authored owner/catalog changes and regenerate; no product behavior changed |
| JS-06A | DONE | Core 01/Core 04, Extensions/Reporting/OTel NLSpecs; extension contract definition and both job owner fragments; generator validation; Extensions parser/canonical model; application/Jobs projections and affected fixtures; generated `internal/gen/contractextensions/artifacts_gen.go`; this tracker | `make format` pass: `.cartulary/test-results/20260808T041353Z-p2220810`; `make generate` pass: `.cartulary/test-results/20260808T041402Z-p2223875`; JSON shape pass: `.cartulary/test-results/20260808T041429Z-p2226373`; generation drift pass: `.cartulary/test-results/20260808T041434Z-p2226775`; generated policy pass: `.cartulary/test-results/20260808T041447Z-p2229519`; Extensions unit rows pass: `.cartulary/test-results/20260808T041452Z-p2229951`; affected service-backed rows pass: `.cartulary/test-results/20260808T041514Z-p2231247`; Markdown lint pass: `.cartulary/test-results/20260808T041603Z-p2234634` | First Markdown lint run failed at `.cartulary/test-results/20260808T041543Z-p2233683` because a new Core 01 table lacked a trailing blank line; corrected; no remaining blocker | Revert the adopted owner/contract/projection checkpoint as one unit and regenerate; no v1 runtime reader exists to preserve |
| JS-06B | DONE | New `internal/platform/jobs/correction_test.go`; strengthened Jobs telemetry/public-projection and Extensions ten-unit tests; `platform.jobs` failure-security verification and three semantic rows; regenerated topology render index; this tracker | `make format` passes: `.cartulary/test-results/20260808T042343Z-p2238282` and `.cartulary/test-results/20260808T042713Z-p2261502`; `make generate` pass: `.cartulary/test-results/20260808T042355Z-p2241305`; intended transition/progress failure: `.cartulary/test-results/20260808T042722Z-p2264555`; intended claim/recovery failure: `.cartulary/test-results/20260808T042440Z-p2245721`; intended failure-security failure: `.cartulary/test-results/20260808T042459Z-p2247332`; intended telemetry failure: `.cartulary/test-results/20260808T042515Z-p2248968`; retained Jobs rows pass: `.cartulary/test-results/20260808T042524Z-p2249343`; Extensions exact-unit rows pass: `.cartulary/test-results/20260808T042542Z-p2251032`; Job API pass: `.cartulary/test-results/20260808T042558Z-p2251856`; Collaboration protocol pass: `.cartulary/test-results/20260808T042558Z-p2251852`; generation drift/policy pass: `.cartulary/test-results/20260808T042622Z-p2258121` and `.cartulary/test-results/20260808T042622Z-p2258130`; checkpoint lint pass: `.cartulary/test-results/20260808T042936Z-p2266725` | Four rows fail only their intended old-behavior assertions: illegal terminal paths/progress semantics; claim/recovery/duplicate dispatch; nil/raw/panic/exhaustion secrecy; scope-surrogate telemetry. These are activation gates for JS-06C/JS-06E, not unrelated regressions | Revert the new tests/semantic rows and regenerate; no production behavior changed in this slice |
| JS-06C | DONE | New private `internal/platform/jobs/transition.go`; lifecycle facade delegation in `jobs.go`; corrected direct-terminal fixture in `jobs_test.go`; this tracker | `make format` pass: `.cartulary/test-results/20260808T043515Z-p2309837`; central transition/progress and retained lifecycle rows pass: `.cartulary/test-results/20260808T043518Z-p2312848`; earlier isolated passes: `.cartulary/test-results/20260808T043331Z-p2272978` and `.cartulary/test-results/20260808T043342Z-p2274788`; Job API pass: `.cartulary/test-results/20260808T043421Z-p2276471`; Extensions service-backed slice passes after correcting owner routing | Durable claim, recovery, handler-failure secrecy, and telemetry correction rows intentionally remain activation gates for JS-06E. An attempted `OWNER=platform.extensionstore` service slice failed because that owner has no service-backed rows; the valid `module.extensions` service slice passed | Revert `transition.go` and the facade/test edits before JS-06D; no schema mutation or dual writer exists yet |
| JS-06D | DONE | Migration 58; migration history; generated SQL model; Jobs definition admission/catalog/startup validation; server composition gate; migration/storage tests and owner rows; renamed storage queries and affected fixtures; this tracker | `make format` pass: `.cartulary/test-results/20260808T045404Z-p2398997`; migration/storage invalid-corpus pass: `.cartulary/test-results/20260808T045251Z-p2393808`; central/storage rows pass: `.cartulary/test-results/20260808T045407Z-p2402020`; migration drift pass: `.cartulary/test-results/20260808T045311Z-p2395836`; generation drift, generated policy, and JSON shape pass: `.cartulary/test-results/20260808T045430Z-p2404673`, `.cartulary/test-results/20260808T045437Z-p2407394`, `.cartulary/test-results/20260808T045439Z-p2407794`; app.server slice passes | Initial migration fixture failed on its FK setup at `.cartulary/test-results/20260808T044544Z-p2324287`; initial storage validation used an incomplete test definition at `.cartulary/test-results/20260808T044841Z-p2336233`; initial drift identified the required history entry at `.cartulary/test-results/20260808T044951Z-p2342014`. All were corrected and rerun. No reset ran | Restore the pre-58 snapshot and old writer before JS-06E writes new rows; after corrected writes, downgrade is unsupported |
| JS-06E | DONE | Durable claim/error/recovery and runner implementation; catalog-backed telemetry; inactive-extension Jobs transaction operation and composition ordering; extensionstore writer removal; affected tests; this tracker | `make format` pass: `.cartulary/test-results/20260808T050632Z-p2643373`; complete Jobs unit slice pass: `.cartulary/test-results/20260808T050638Z-p2651867`; complete Jobs service-backed slice passes; first correction/security group pass: `.cartulary/test-results/20260808T045951Z-p2413545`; telemetry row pass: `.cartulary/test-results/20260808T050042Z-p2419280`; complete affected Job API, Extensions, Imports, Incident Portability, Reference Data, Reporting, Collaboration, and app.server unit/service-backed invocations pass; production source scan finds no Jobs-table write outside Jobs | No remaining JS-06E blocker. Raw handler/panic values are absent by construction; nil mutable returns persist only `job_handler_incomplete`; lease-only recovery is public-event-free | Stop the corrected process; schema downgrade is unsupported after v2 writes. Revert the runner/durable/telemetry and inactive-reconciliation routing together; never restore an old writer onto v2 rows |
| JS-02 | DONE | New Jobs owner-port contracts; Auth transaction lookup/update operations; Extensions transaction append; application and test adapters; finalizer, cancellation, admission, and composition migrations; affected tests; this tracker | `make format` pass: `.cartulary/test-results/20260808T051344Z-p2669895`; complete Jobs unit slice pass: `.cartulary/test-results/20260808T051404Z-p2673429`; complete Jobs service-backed slice pass: `.cartulary/test-results/20260808T051435Z-p2677730`; complete Job API service-backed slice pass: `.cartulary/test-results/20260808T051455Z-p2679318`; complete Extensions service-backed slice pass: `.cartulary/test-results/20260808T051509Z-p2685393`; boundary check pass: `.cartulary/test-results/20260808T051601Z-p2710387`; production source scans pass | An initial `platform.jobs.unit` row alias was rejected because `ROWS` requires canonical unique row IDs; the complete owner slice was used and passed. No remaining blocker; Jobs contains neither foreign table name nor Auth import/type | Revert ports, owner persistence operations, and composition adapters as one checkpoint. Cancellation/finalization still use one transaction and no dual write exists |
| JS-03 | DONE | Jobs public types, manager facade, extension admission/terminal, immutable definition catalog, stored projection, transition policy/persistence, lifecycle persistence, durable persistence, transaction service, and telemetry concern files; removed unused canonical JSON helper and `.gitkeep`; this tracker | `make format` passes: `.cartulary/test-results/20260808T052222Z-p2713847`, `.cartulary/test-results/20260808T052426Z-p2721515`, and `.cartulary/test-results/20260808T052449Z-p2724730`; complete Jobs unit slice pass: `.cartulary/test-results/20260808T052452Z-p2727760`; complete Jobs service-backed slice pass: `.cartulary/test-results/20260808T052526Z-p2732068`; `make test-fast` pass: `.cartulary/test-results/20260808T052546Z-p2733748` | No failed validation or remaining blocker. The internal catalog cleanup strengthened equality from progress-unit-only matching to the complete immutable definition | Revert the cohesive file split and shared-catalog change together; no package path, protocol, or compatibility shim was introduced |
| JS-04 | DONE | Job API read/cancel port; Imports, Incident Bundles, Reference Data, Reporting, and Report Composition admission/runtime/runner ports; Jobs runnable/finalization transaction operations; narrowed Extensions finalizer and application adapters; HTTP dependency bag cleanup; runner configuration contract; affected tests and boundary allowlist; this tracker | Job API pass: `.cartulary/test-results/20260808T053003Z-p2819050` and `.cartulary/test-results/20260808T053013Z-p2821618`; Imports pass: `.cartulary/test-results/20260808T054929Z-p3241453` and `.cartulary/test-results/20260808T055004Z-p3268921`; Incident Bundles pass: `.cartulary/test-results/20260808T054844Z-p3228740` and `.cartulary/test-results/20260808T054916Z-p3239791`; Reference Data pass: `.cartulary/test-results/20260808T053833Z-p3000782` and `.cartulary/test-results/20260808T053904Z-p3025002`; Reporting/Composition pass: `.cartulary/test-results/20260808T054014Z-p3050629`, `.cartulary/test-results/20260808T054027Z-p3055453`, `.cartulary/test-results/20260808T054034Z-p3056903`, and `.cartulary/test-results/20260808T054043Z-p3058478`; Extensions pass: `.cartulary/test-results/20260808T054730Z-p3175101` and `.cartulary/test-results/20260808T054806Z-p3205888`; app.server pass: `.cartulary/test-results/20260808T055050Z-p3292348`; complete Jobs slices pass: `.cartulary/test-results/20260808T055125Z-p3319608` and `.cartulary/test-results/20260808T055145Z-p3321859`; boundary pass: `.cartulary/test-results/20260808T055204Z-p3323496` | Initial Imports root `.cartulary/test-results/20260808T053233Z-p2827134` failed because the next consumer still called the removed runner signature; corrected before rerun. Reference roots `.cartulary/test-results/20260808T053659Z-p2948623` and `.cartulary/test-results/20260808T053732Z-p2975118` exposed two obsolete queued-during-handler assertions; corrected to required running-before-handler. Extensions roots `.cartulary/test-results/20260808T054114Z-p3063123` and `.cartulary/test-results/20260808T054149Z-p3089718` exposed a partial test catalog; the fixture now supplies one exact shared catalog. Boundary root `.cartulary/test-results/20260808T054447Z-p3167326` exposed the JS-03 file-move allowlist and was corrected. No blocker remains | Revert consumer ports and composition as one checkpoint. No compatibility facade remains; restore neither the broad HTTP dependency fields nor external transition locks |
| JS-05 | DONE | Authored backend boundary policy; static-analysis lock detection and Jobs-specific fixtures; Jobs inactive-reconciliation validation operation; Extensions reconciliation caller; this tracker | `make format` pass: `.cartulary/test-results/20260808T060036Z-p3327961`; Jobs slices pass: `.cartulary/test-results/20260808T060051Z-p3331158` and `.cartulary/test-results/20260808T060117Z-p3334286`; Extensions slices pass: `.cartulary/test-results/20260808T060143Z-p3335986` and `.cartulary/test-results/20260808T060221Z-p3362394`; `make generate` pass: `.cartulary/test-results/20260808T060253Z-p3385359`; generation drift pass: `.cartulary/test-results/20260808T060303Z-p3387669`; generated policy pass: `.cartulary/test-results/20260808T060313Z-p3390427`; final boundary pass: `.cartulary/test-results/20260808T060317Z-p3390933`; JSON shape pass: `.cartulary/test-results/20260808T060322Z-p3391280`; checkpoint Markdown lint pass: `.cartulary/test-results/20260808T060427Z-p3392233` | Policy design exposed the remaining inactive-reconciliation `FOR UPDATE` on Jobs. It was replaced with Jobs-owned `ValidateInactiveJobTx`, preserving the caller transaction and candidate identity checks. Fixtures prove owner/app adapter, migration, test, and ordinary read positives plus non-owner write/lock, foreign storage, broad capability, raw lock, and peer-persistence negatives. No blocker remains | Revert policy/checker and semantic reconciliation validation together if rollback occurs; do not restore the external row lock without also removing its guard |
| JS-07 | DONE | New deployment-faithful cutover/recovery integration test and `platform.jobs` row; generated topology index; lowercase Job API constructor diagnostic; final tracker/handoff | Owner explain/task-guide pass; `make format`: `.cartulary/test-results/20260808T060751Z-p3395579`; `make generate`: `.cartulary/test-results/20260808T060801Z-p3398673`; isolated drained cutover: `.cartulary/test-results/20260808T060812Z-p3401018`; complete Jobs: `.cartulary/test-results/20260808T060830Z-p3402740` and `.cartulary/test-results/20260808T060855Z-p3405169`; Extensions: `.cartulary/test-results/20260808T060935Z-p3407048` and `.cartulary/test-results/20260808T061012Z-p3431822`; final Job API: `.cartulary/test-results/20260808T062908Z-p3930691` and `.cartulary/test-results/20260808T062923Z-p3932737`; Imports: `.cartulary/test-results/20260808T061104Z-p3459224` and `.cartulary/test-results/20260808T061142Z-p3486531`; Incident Bundles: `.cartulary/test-results/20260808T061216Z-p3509890` and `.cartulary/test-results/20260808T061247Z-p3516767`; Reference Data: `.cartulary/test-results/20260808T061305Z-p3518515` and `.cartulary/test-results/20260808T061338Z-p3544232`; Reporting/Composition: `.cartulary/test-results/20260808T061411Z-p3566728`, `.cartulary/test-results/20260808T061425Z-p3571444`, `.cartulary/test-results/20260808T061433Z-p3572927`, and `.cartulary/test-results/20260808T061446Z-p3574548`; Collaboration: `.cartulary/test-results/20260808T061451Z-p3576022` and `.cartulary/test-results/20260808T061623Z-p3608393`; app.server: `.cartulary/test-results/20260808T061750Z-p3635117` and `.cartulary/test-results/20260808T061830Z-p3661598`; drift/policy/boundary roots: `.cartulary/test-results/20260808T061905Z-p3683805`, `.cartulary/test-results/20260808T061918Z-p3686587`, `.cartulary/test-results/20260808T061931Z-p3689350`, `.cartulary/test-results/20260808T061935Z-p3689790`, and `.cartulary/test-results/20260808T061956Z-p3690377`; `make test-fast`: `.cartulary/test-results/20260808T062005Z-p3690872`; stateful browser: `.cartulary/test-results/20260808T062214Z-p3755259`; final `make agent-finalize`: `.cartulary/test-results/20260808T062931Z-p3934214`; final `make check` and explain: `.cartulary/test-results/20260808T063551Z-p4119107`; checkpoint Markdown lint: `.cartulary/test-results/20260808T064001Z-p2067` | First check `.cartulary/test-results/20260808T062431Z-p3784022` failed only related `lint-go` ST1005 on the new capitalized Job API error; corrected and focused lint/Job API evidence passed. Second check `.cartulary/test-results/20260808T062947Z-p3936932` failed an unrelated two-second standalone-server reset timeout; exact canonical row rerun passed at `.cartulary/test-results/20260808T063521Z-p4099687`. Final check passed 730/730. `RESULTS_DIR` was unset for finalization, so retained-run maintenance was skipped. No reset/reseed or external deployment mutation ran | The isolated rehearsal retains its pre-58 database boundary until migration, proves no active lease or 24-hour replay event, and opens dequeue only after migration/catalog validation. For an external environment, retain the pre-cutover rollback point and execute the same drained sequence; never start a v1 writer after v2 writes |

### Migration and rollout defaults

`JP-REQ-017`: The compatibility preflight MUST report counts and bounded safe
job-kind/status tokens for negative completed, nonpositive totals,
completed-over-total, incomplete succeeded rows, unresolved unit mappings,
active leases/writer versions, and still-replayable illegal sequences. It MUST
NOT report payloads, incident content, raw job IDs, or storage secrets.

`JP-REQ-018`: Backfill MUST resolve each mutable job from an immutable adopted
job-kind binding. Unknown or ambiguous nonterminal mappings block startup.
Terminal rows that cannot be safely mapped MAY retain `NULL` only when they are
immutable and cannot re-enter processing.

`JP-REQ-019`: Invalid retained data uses these defaults:

- Any retained job that violates the new state or progress invariants blocks
  cutover. It is never normalized or repaired by inference.
- Unsupported retained job data requires the explicitly approved pre-release,
  targeted `make db-reset` and reseed path. The implementation and migration
  MUST NOT run that destructive path automatically.
- Illegal Collaboration history must be absent from the current 24-hour replay
  horizon. Wait for owner-managed pruning or use the same approved reset/reseed
  path; never rewrite Collaboration events in place.

`JP-REQ-020`: The default deployment is a drained single-writer cutover:

1. Close dequeue admission.
2. Drain old workers and leases.
3. Run compatibility preflight.
4. Apply migration/backfill and validate constraints.
5. Deploy only the corrected writer version.
6. Reopen dequeue admission.
7. Recover nonterminal jobs through the corrected claim path.
8. Run complete owner evidence.

Mixed old/new writers, permanent dual representations, and implicit rolling
compatibility are prohibited. Zero-downtime deployment requires a separately
specified two-phase protocol.

## 8. Validation Plan

### Required command matrix

| Validation layer | Command | Scope and precondition | Required checkpoint | Notes |
| --- | --- | --- | --- | --- |
| Tracker documentation | `make lint-markdown` | This documentation-only revision | Current task | Only validation authorized for this tracker edit. |
| Owner discovery | `make explain-test-owner OWNER=platform.jobs` | Valid after JS-01 | JS-01 and final | Must resolve the complete active owner inventory. |
| Owner guidance | `make task-guide ROLE=module-author OWNER=platform.jobs` | Valid after JS-01 | JS-01 and every implementation session | No jobs-specific Make target is added. |
| Platform owner slice | `make test-slice OWNER=platform.jobs` | Active owner exists | JS-01, JS-06B onward | Omitted rows select the complete all-profile inventory. |
| Platform service slice | `make service-backed-test-slice OWNER=platform.jobs` | Active service-backed rows exist | JS-06B onward | Omitted rows select the complete service-backed inventory. |
| Existing Extensions unit | `make test-slice OWNER=module.extensions ROWS=module.extensions.unit.job_admission_and_terminal_contracts_64d299b374` | Existing row | Before/after extension contract projection | Remains Extensions-owned. |
| Existing Extensions integration | `make service-backed-test-slice OWNER=module.extensions ROWS=module.extensions.integration.job_progress_intent_persistence_characterization_a2b3c4d5e6,module.extensions.integration.named_portability_job_recovery_f7c496c10b,module.extensions.integration.job_finalization_cancellation_and_reconciliation_2c712826c2` | Existing rows | JS-06E, JS-02, final | Covers progress, recovery, finalization, cancellation, and reconciliation. |
| Job API integration | `make service-backed-test-slice OWNER=module.jobapi` | Existing active owner | JS-06E, JS-02, JS-04, final | Preserves route and authorization behavior. |
| Migration | `make migration-drift` | Authored migration changed | JS-06D and final | No direct migration tool command. |
| Generation | `make generate` | Authored owner/contract/harness inputs changed | JS-01 and JS-06A | Refresh outputs only through this target. |
| Generated drift | `make generate-drift` | After generation and final | JS-01, JS-06A, JS-05, final | Must be drift-clean. |
| Generated policy | `make generated-artifact-policy-check` | After any generated output | JS-01, JS-06A, JS-05, final | No hand-edited generated file. |
| Boundary/static | `make backend-module-boundary-check` | Before and after writer/port/guard changes | JS-06E, JS-02, JS-05, final | Must reject all prohibited production bypasses. |
| Fast broad | `make test-fast` | Narrow owner evidence passes | JS-03 onward | Broadens only after targeted evidence. |
| Browser stateful | `make browser-e2e-stateful` | Only if HTTP/WS/frontend-observable behavior changes or corrected sequences need browser proof | Conditional before final | Not required for an entirely backend-internal checkpoint. |
| Finalization | `make agent-finalize` | Before broad end-of-run verification | JS-07 | Supply retained `RESULTS_DIR` only for a successful full warm run; otherwise record the skip. |
| Full check | `make check` | All narrow and drift gates pass | JS-07 | Report the actual run root and any failure classification. |

No product, migration, generation, owner-slice, boundary, browser, or full-check
validation is claimed by this documentation-only revision.

### Required behavior evidence

| Test group | Required assertion | Requirement |
| --- | --- | --- |
| Exhaustive transitions | Every allowed pair succeeds; every other pair rejects without mutation/event | JP-REQ-006 |
| Direct queued terminal | Queued-to-succeeded and queued-to-canceled reject | JP-REQ-006 |
| Recovery ordering | Queued recovery commits running before handler invocation | JP-REQ-007 |
| Concurrent claim | Exactly one claimant changes queued to running | JP-REQ-007, JP-REQ-010 |
| Cancellation | Queued/running cancellation passes through cancel-requested and commits all side effects atomically | JP-REQ-008 |
| Success-wins | Cancel-requested reaches succeeded only after authoritative effect commit | JP-REQ-006, JP-REQ-008 |
| Terminal immutability | Terminal state/progress accepts no later mutation | JP-REQ-006, JP-REQ-012 |
| Completed monotonicity | Increase succeeds; equality is no-op; decrease rejects | JP-REQ-012 |
| Total discovery/persistence | Null-to-positive succeeds; known total never clears or decreases | JP-REQ-012 |
| Unit stability | Unit mapping is present and immutable for every mutable job | JP-REQ-011, JP-REQ-012 |
| Bounds/success | Completed-over-total rejects; known-total success requires explicit equality; unknown-total success is legal | JP-REQ-012, JP-REQ-013 |
| Concurrent progress | Stale lower progress loses without overwriting | JP-REQ-010, JP-REQ-012 |
| Transaction publication | Rejected update emits no intent; committed row/event/summary/proof agree | JP-REQ-014 |
| Recovery progress | Recovery preserves prior completed, total, and unit | JP-REQ-007, JP-REQ-011 |
| Public parity | HTTP and live/replayed WS expose identical legal committed sequences | JP-REQ-002, JP-REQ-014 |
| Extension coverage | All ten job kinds declare and persist one stable unit | JP-REQ-011 |
| Authorization regression | Read/cancel rederive current route-family authorization and preserve not-found masking | JP-REQ-002 |
| Writer boundary | No production jobs write bypass remains | JP-REQ-004, JP-REQ-016 |

## 9. Top-Level Work Tracker

Only `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`, and `DROPPED` are
valid statuses in this table.

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| JP-001 | Define target, authority, identities, and exclusions | WF-00 | DONE | none | Sections 1 and 3 | Scope and authority boundaries are explicit. |
| JP-002 | Inventory every target file and production jobs writer | WF-01 | DONE | JP-001 | Section 2 | Eleven target entries and all discovered production writers are recorded. |
| JP-003 | Freeze public contracts and owner mapping | WF-02 | DONE | JP-002 | Section 4 | Every discovered public contract has one owner and evidence posture. |
| JP-004 | Resolve RB-001 through RB-003 at planning level | WF-02 | DONE | JP-003 | Sections 4 and 11 | Owner ID, state correction, and progress closure are decision-complete. |
| JP-005 | Register `platform.jobs` and route evidence | WF-03 | DONE | JP-004 | JS-01 owner contract, family manifest, generated topology, and retained run roots | Owner commands accept the ID; three rows and all selected tests resolve exactly once; complete slices pass. |
| JP-006 | Adopt Core clarifications and exact job-kind unit declarations | WF-04 | DONE | JP-005 | JS-06A owners, typed facts, parser/canonical projection, generated bindings, and retained run roots | Owners contain every required behavior; all ten units project through v2; active runtime inputs contain no v1 reader. |
| JP-007 | Add owner-aligned correction evidence | WF-05 | DONE | JP-006 | JS-06B semantic tests, owner rows, intended-failure roots, and retained green owner roots | Required tests fail on the old drift for their intended assertions; selectors are unique; unrelated retained evidence passes. |
| JP-008 | Implement central transition/progress substrate | WF-05 | DONE | JP-007 | `internal/platform/jobs/transition.go`; central and retained lifecycle result roots | One private closed matrix and typed safe error boundary governs running, progress, cancellation, and terminal publication; exact repeats and rejected updates are mutation/event-free. |
| JP-009 | Apply progress-unit migration and compatibility gates | WF-05 | DONE | JP-008 | Migration 58, startup validation, migration/storage rows, and drift roots | Exact backfill, legacy-terminal null preservation, constraints, safe preflight, lease/replay rejection, and catalog-matching startup validation pass without inferred repair. |
| JP-010 | Migrate every worker, finalizer, recovery path, and external writer | WF-05 | DONE | JP-009 | Complete Jobs/affected owner slices, correction/security rows, and production writer scan | Claims publish running before invocation, recover without regression, use unique attempts, fail with closed safe tokens, clear terminal leases, and route inactive reconciliation through Jobs. |
| JP-011 | Extract Auth/Extensions persistence ports | WF-06 | DONE | JP-010 | Jobs owner-port contracts, application/test adapters, owner persistence functions, owner slice roots, and source scans | Jobs contains no foreign-owner table access; cancellation replay/conflict, observation, finalization, and rollback remain atomic. |
| JP-012 | Split Jobs private internals | WF-06 | DONE | JP-011 | Cohesive Jobs concern files, one shared private catalog, removed unused export/placeholder, complete platform roots, and `test-fast` root | Public values/facade are isolated from stored projection, policy, lifecycle/durable persistence, catalog, transaction publication, and telemetry concerns. |
| JP-013 | Narrow concrete consumers | WF-06 | DONE | JP-012 | Consumer-owned ports, composition injection, Jobs runnable/finalization operations, source scans, affected owner roots, and boundary root | Job API has read/cancel only; producers, workers, and finalizers hold narrow semantic ports; the HTTP dependency bag has no Jobs capability; external transition locks/direct mutable-status reads are gone. |
| JP-014 | Add boundary anti-regression rules | WF-07 | DONE | JP-011, JP-013 | JS-05 policy, static-analysis fixtures, semantic lock correction, and retained result roots | Canonical boundary checks reject all defined bypasses while owner adapters, migrations, tests, and permitted reads remain accepted. |
| JP-015 | Execute drained cutover and final evidence | WF-08 | DONE | JP-010, JP-014 | JS-07 cutover rehearsal, complete owner/browser/broad roots, and final acceptance table | The corrected writer recovered and served the retained job only after readiness and dequeue reopening; every implementation criterion passes. |
| JP-016 | Remove `.gitkeep` | WF-08 | DONE | JP-012 | JS-03 cleanup | The obsolete placeholder is removed. |

## 10. Session Handoff Log

The 2026-08-07 rows preserve the planning handoff. Subsequent rows are the
slice-by-slice execution record required by the controlling checkpoint
protocol.

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 America/New_York | Codex original planning session | WF-00 through WF-02 planning complete | Inspected framework, Core 00-04, subsystem NLSpecs, domain and support guides; touched only this tracker | `git status`, `git branch --show-current`, `git rev-parse`, `sed`, `rg` | Target exists; label is safe; baseline was clean; no owner contradiction found | None for tracker completion | Seek later authorization for JS-01 before implementation. |
| 2026-08-07 America/New_York | Codex NLSpec revision session | Requirements and owner boundaries are decision-complete | Inspected `analysis-notes.md`, `nlspec-spec.md`, current tracker, owner inputs, and adopted sources; touched only this tracker | `git status`, `sha256sum`, `sed`, `rg`, `jq`, `git log` | RB-001 through RB-003 resolved at planning level; no owner contradiction; implementation remains pending | None at planning level | Execute JS-01 in a separately authorized task. |
| 2026-08-08 America/New_York | Codex remediation execution | JS-01 activated after tracker bootstrap | Updated this tracker; confirmed clean baseline `2e80c849048e012867dc1ff860102ff1a99ca69a` | `git status --short`; `git rev-parse HEAD`; `rg`; `sed`; `make lint-markdown` | Authorized sequence, v2 catalog, exact units, reset/reseed posture, runtime-security gaps, and checkpoint protocol recorded | JS-01 evidence pending | Register and validate `platform.jobs`; do not begin JS-06A until checkpoint is `DONE` and lint-clean. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 America/New_York | Codex original planning session | Legitimate platform boundary with two foreign-storage leaks and mixed internals | Inspected every target file, server/runtime dependencies, Job API, extension store/assembly, imports, incident bundles, reporting, and reference data; touched only this tracker | `rg`, `sed`, `find`, `git ls-files` | Keep Jobs persistence/runner; move Auth/Extensions SQL behind owner-backed ports; split later | Evidence routing and implementation authority | Characterize before structural extraction. |
| 2026-08-07 America/New_York | Codex NLSpec revision session | Central writer/transition requirement added | Re-inspected target writes, module workers, and `internal/platform/extensionstore/reconciliation.go`; touched only this tracker | `rg`, `sed` | Extensions reconciliation is a third production write boundary and must migrate in JS-06E | No planning blocker | Register owner, adopt requirements, test, then centralize all writers. |
| 2026-08-08 America/New_York | Codex remediation execution | JS-02 complete; JS-03 activated | Added required Jobs consumer ports, Auth- and Extensions-owned transaction operations, application/test adapters, and finalizer adapter injection; removed foreign SQL/types from Jobs and Auth SQL from the Extensions finalizer | Complete Jobs unit/service-backed, Job API service-backed, and Extensions service-backed slices; boundary check; production foreign-schema/import scans; `make lint-markdown` | Cancellation replay/conflict, observation append, progress intent, final idempotency replacement, terminal state, and proof publication retain one caller transaction with no dual write | None | Split Jobs internals by concern, remove obsolete exports and `.gitkeep`, then run complete platform slices and `make test-fast`. |
| 2026-08-08 America/New_York | Codex remediation execution | JS-03 complete; JS-04 activated | Reorganized Jobs into cohesive concern files, replaced two definition maps with one shared immutable catalog, removed the unused canonical JSON export and stale placeholder, and retained the package boundary | `make format`; complete platform owner slices; `make test-fast`; file/export scans; `make lint-markdown` | Public serialization and package identity are unchanged; all 349 fast units pass; complete catalog equality now prevents partial-definition aliasing | None | Inventory each concrete consumer capability and migrate one consumer at a time. |
| 2026-08-08 America/New_York | Codex remediation execution | JS-04 complete; JS-05 activated | Added consumer-owned Jobs interfaces for every route/producer/worker/finalizer; removed all Jobs capabilities from `httpapi.DependencySet`; added Jobs-owned runnable and extension-finalization transaction operations; removed the external lock and direct mutable-state queries | Per-consumer unit/service owner slices, app.server and complete Jobs slices; source scans; backend boundary check; `make lint-markdown` | No production module or extension finalizer stores a concrete Jobs service; public behavior remains compatible; running-before-handler evidence is corrected | None | Encode these boundaries in the authored canonical policy with positive and negative fixtures. |
| 2026-08-08 America/New_York | Codex remediation execution | JS-05 complete; JS-07 activated | Added owner-only Jobs write/lock enforcement, Jobs foreign-storage and peer-import guards, narrow-capability guards, and fail-closed positive/negative fixtures; moved inactive-reconciliation locking and candidate validation into Jobs | `make format`; complete Jobs and Extensions unit/service-backed slices; `make generate`; generation/policy/JSON drift; final boundary root `.cartulary/test-results/20260808T060317Z-p3390933` | Every prohibited bypass fixture fails its intended canonical rule; owner/app adapters, migrations, tests, and ordinary reads pass; production policy is green | None | Run the drained compatibility/replay readiness path, final owner/module/browser/broad validation, and evidence-backed acceptance update. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 America/New_York | Codex original planning session | Frontend is contract-consumer-only and out of implementation scope | Inspected collaboration session, import coordinator, import panel, and related tests; touched only this tracker | `rg`, `sed` | No frontend shell, view-schema, projection, saved-view, grid-vendor, or selector logic belongs in Jobs | None | Run stateful browser evidence only for observable changes. |
| 2026-08-07 America/New_York | Codex NLSpec revision session | Public shape remains frozen; corrected legal intermediate states may become observable | Reused prior frontend evidence and Core 03/Core 04 acceptance; touched only this tracker | `rg`, `sed` | No frontend code change is planned; public progress remains `{completed,total}` | None | Add browser evidence only if corrected sequences expose a client defect. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 America/New_York | Codex original planning session | HTTP, WS, storage, recovery, extension-job, and telemetry contracts frozen | Inspected OpenAPI/WS owners, extension fragments/bindings, migrations 00005/00027/00034, ownership/recovery manifests, and generated Go/TS; touched only this tracker | `rg`, `sed`, `jq`, Make explain targets | Generated artifacts are projections; no authored contract/schema change was then planned | Original RB-002/RB-003 | Use canonical generation only after authored input changes. |
| 2026-08-07 America/New_York | Codex NLSpec revision session | Future owner and schema edits are precisely routed | Inspected ten job-kind fragments, extension contract parsing/projection, current migration tail through 00057, and jobs schema; touched only this tracker | `find`, `rg`, `sed`, `jq` | JS-06A must add owner-declared units; JS-06D uses next free migration, currently 00058; public projections remain unchanged | Owner adoption is an implementation precondition | Adopt owners first, then generate; never hand-edit projections. |
| 2026-08-08 America/New_York | Codex remediation execution | JS-06A complete; JS-06B activated | Adopted lifecycle/progress/security corrections; replaced active v1 contract with v2; added all ten owner units; updated validation, parser, digest projection, application binding, and generated contracts | `make format`; `make generate`; `make json-shape-check`; generation/policy drift; targeted Extensions unit/service-backed slices; `rg`; `jq`; `make lint-markdown` | Ten exact units project through v2 and affect canonical digests; no active v1 reader remains; all binary gates pass | None | Add correction/security evidence and capture intended old-behavior failures in JS-06B. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 America/New_York | Codex original planning session | Target tests partly route through Extensions; Job API owner active; platform owner absent | Inspected tests, owners, Job API/Extensions families, and harness owner; touched only this tracker | `make help`, `make help-all`, task/explain targets, `make lint-markdown`, `rg`, `sed`, `jq` | Existing owners discovered; invalid attempted IDs rejected; original Markdown lint passed | Original RB-001 | Harness owner decides/creates semantic owner. |
| 2026-08-07 America/New_York | Codex NLSpec revision session | `platform.jobs` selected; implementation not authored | Inspected owner grammar, current/historical registries, platform owner examples, and exact target tests; touched only this tracker | `rg`, `sed`, `jq`, `git log`, `make lint-markdown` | ID is grammar-valid, currently unused, and not found in owner history; Markdown lint passed; no product or owner slice ran | None at planning level | JS-01 authors owner inputs and runs complete owner evidence. |
| 2026-08-08 America/New_York | Codex remediation execution | JS-01 complete; JS-06A activated | Added `platform.jobs` owner/contract and three generic Jobs rows; updated both registries and generated render index; retained Extensions/Job API/profile selectors | `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; owner explain/guide; complete owner unit/service-backed slices; `make lint-markdown` | All JS-01 binary gates passed; run roots are recorded in the slice ledger | None | Adopt the v2 normative correction contract and exact unit projections in JS-06A. |
| 2026-08-08 America/New_York | Codex remediation execution | JS-06B complete; JS-06C activated | Added transition/progress, claim/recovery/publication, duplicate-dispatch, runner-failure/secrecy, telemetry, public-projection, and exact-unit evidence; updated semantic owner rows | Targeted owner slices and retained-root inspection listed in the slice ledger; generation/policy drift; `make lint-markdown` | All new defect rows fail the old implementation for the intended assertions; retained Jobs, Extensions, Job API, and Collaboration evidence passes | Expected red rows require JS-06C and JS-06E corrections | Implement the private transition/progress substrate without activating unsupported external writers. |
| 2026-08-08 America/New_York | Codex remediation execution | JS-06C complete; JS-06D activated | Added the private stored/public projection boundary, one closed transaction-scoped state/progress policy, semantic lifecycle delegation, conditional state writes, monotonic progress, exact no-op behavior, and committed-change-only intent publication | `make format`; targeted `platform.jobs` service-backed rows; complete Job API slice; Extensions service-backed slice; `make lint-markdown` | Central and retained lifecycle evidence passes; public Job API projection remains compatible | Durable writer/security/telemetry rows remain expected red until JS-06E; invalid service owner selection was corrected to `module.extensions` | Author migration `00058`, immutable catalog, safe preflight/startup gate, and invalid-corpus evidence. |
| 2026-08-08 America/New_York | Codex remediation execution | JS-06D complete; JS-06E activated | Added migration 58, generic job identity and private unit persistence, exact catalog derivation, safe preflight, no-partial-mutation rejection, startup validation before runner/listener activation, and migration/storage owner evidence | `make format`; targeted platform migration/storage/lifecycle rows; `make migration-drift`; generation/policy/JSON drift; complete app.server slice | Fresh install, exact upgrade, all ten mappings, legacy-terminal nulls, invalid progress, unknown mappings, active leases, illegal replay, startup mismatch secrecy, and rollback shape pass | No reset was run; downgrade remains valid only before a corrected writer creates v2 rows | Correct durable claim/recovery/error/finalizer and every remaining Jobs-table writer. |
| 2026-08-08 America/New_York | Codex remediation execution | JS-06E complete; JS-02 activated | Corrected durable claim/recovery/error/exhaustion, unique dispatch attempts, handler secrecy, incomplete-handler closure, actual-kind telemetry, terminal lease clearing, and inactive-extension atomic terminal routing | Complete Jobs owner slices; correction/security/telemetry rows; all affected owner slices; app.server slice; production write scan; `make lint-markdown` | One corrected Jobs writer governs every mutable row; all JS-06B red evidence is green; public state/event sequencing and secrecy sentinels pass | Jobs still directly accesses Auth idempotency and Extensions cancellation storage | Extract caller-transaction ports and composition adapters without dual writes. |
| 2026-08-08 America/New_York | Codex remediation execution | JS-02 complete; JS-03 activated | Preserved cancellation/finalizer assertions while replacing direct cross-owner persistence with composition-supplied transaction ports | Jobs unit root `.cartulary/test-results/20260808T051404Z-p2673429`; Jobs service root `.cartulary/test-results/20260808T051435Z-p2677730`; Job API service root `.cartulary/test-results/20260808T051455Z-p2679318`; Extensions service root `.cartulary/test-results/20260808T051509Z-p2685393` | Complete affected semantic owner inventories pass; source scans are clean | None | Begin only the cohesive private-internal split. |
| 2026-08-08 America/New_York | Codex remediation execution | JS-03 complete; JS-04 activated | Retained all platform evidence while moving implementation concerns and tightening the catalog identity boundary | Jobs unit root `.cartulary/test-results/20260808T052452Z-p2727760`; Jobs service root `.cartulary/test-results/20260808T052526Z-p2732068`; fast root `.cartulary/test-results/20260808T052546Z-p2733748` | Complete Jobs and broad fast evidence pass after the structural break; no compatibility shim exists | None | Narrow Job API first, then transactional producers, workers, and owner finalizers. |
| 2026-08-08 America/New_York | Codex remediation execution | JS-04 complete; JS-05 activated | Validated each consumer after capability narrowing and reran corrected evidence after every discovered failure | Pass roots are recorded in the JS-04 ledger; final app.server, Jobs, and boundary roots are `.cartulary/test-results/20260808T055050Z-p3292348`, `.cartulary/test-results/20260808T055125Z-p3319608`, `.cartulary/test-results/20260808T055145Z-p3321859`, and `.cartulary/test-results/20260808T055204Z-p3323496` | All consumer inventories pass; no broad HTTP capability, concrete module field, external transition lock, or direct module mutable-status query remains | None | Add canonical positive/negative boundary fixtures, regenerate if required, and validate drift. |
| 2026-08-08 America/New_York | Codex remediation execution | JS-07 complete; no active slice | Added and passed the isolated drained v1-to-v2 cutover/recovery row, then ran every required owner, drift, boundary, fast, browser, finalization, and broad check gate | Complete roots are recorded in the JS-07 ledger; final `make check` root `.cartulary/test-results/20260808T063551Z-p4119107` passes 730/730 and was inspected with `make explain-run` | Repository implementation and deployment rehearsal are complete; public polling/live/replay state handling remains compatible | No repository blocker. External rollout remains an operator action using the recorded fail-closed sequence | Handoff to deployment operators; do not mix writer versions or bypass retained-state preflight. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 America/New_York | Codex original planning session | Job API rederives authorization; Jobs directly accesses Auth idempotency storage | Inspected Core 04, Job API routes/tests, Jobs, and ownership manifest; touched only this tracker | `rg`, `sed`, `jq` | Preserve request-time authorization/not-found masking; move persistence behind Auth adapter | Characterization and implementation authority | Execute Auth port after characterization. |
| 2026-08-07 America/New_York | Codex NLSpec revision session | Authorization remains outside lifecycle correction; safe diagnostics are explicit | Reused Core 04/Job API evidence and inspected transition/progress error implications; touched only this tracker | `rg`, `sed` | Internal reasons must not leak SQL, IDs, payloads, incident content, or unit IDs | None at planning level | Preserve Job API evidence through JS-06E and JS-02. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 America/New_York | Codex original planning session | Planning inventory and sequencing complete; implementation not started | Inspected owner/code/test/harness evidence; touched only this tracker | Read-only discovery commands | Three stable blockers remained; no product validation or refactor ran | RB-001, RB-002, RB-003 | Start with owner decision and characterization. |
| 2026-08-07 America/New_York | Codex NLSpec revision session | No unanswered planning blocker; implementation gates remain | Inspected retained-data, replay, migration, writer, and rollout evidence; touched only this tracker | Read-only commands listed above | Defaults are fail-closed; mixed writers prohibited; RB decisions resolved but not implemented | Implementation evidence gates IG-001 through IG-004 | Start JS-01; do not combine correction with JS-02/JS-03. |
| 2026-08-08 America/New_York | Codex remediation execution | All slices and implementation evidence gates closed | Inspected final retained roots, failure artifacts, source status, acceptance criteria, and rollback posture; updated the final handoff | `make explain-run` on both failing and final passing roots; exact flaky-row rerun; final tracker lint | Related lint defect was corrected; unrelated serverprocess timeout passed exact rerun; final broad graph is green | No code/specification blocker; external environment state was not mutated or inferred | Execute the operator-owned drained cutover in each deployment and preserve the pre-v2 rollback point until migration commits. |

## 11. Open Questions and Blockers

No unanswered planning question remains. RB-001 through RB-003 are resolved
decisions, not completed implementation. They MUST NOT be marked `DONE` until
their closure evidence exists.

| ID | Resolution | Required closure evidence | Current status |
| --- | --- | --- | --- |
| RB-001 | The semantic harness owner is `platform.jobs`. Generic lifecycle, progress, durable runner, recovery, and Jobs telemetry evidence belongs to it; cross-owner postconditions retain their existing owners. | Active registries/contracts/families; exact selectors; owner explain/guide; complete all/service-backed slices; drift-clean generated outputs | RESOLVED — CLOSED by JS-01 and JS-07 evidence |
| RB-002 | Core 01 controls. Direct queued-to-succeeded/canceled transitions are prohibited; dispatch/recovery claims running before work; cancellation passes through cancel-requested; one atomic Jobs substrate governs all writers. | Exhaustive matrix, concurrent claim, recovery sequencing, cancellation race, atomicity, public polling/WS sequence, and drained-writer evidence | RESOLVED — CLOSED by JS-06C, JS-06E, and JS-07 evidence |
| RB-003 | Progress rejects completed regression, total clearing/regression, unit change, bound violation, and incomplete known-total success. Every mutable job persists an immutable owner-declared internal unit. | Adopted owner clarification/unit map; migration/backfill; constraints; concurrent progress, success, recovery, HTTP/WS, and extension parity evidence | RESOLVED — CLOSED by JS-06A through JS-06E and JS-07 evidence |

### Implementation evidence gates

| Gate | Condition that prevents completion | Fail-closed result | Current status |
| --- | --- | --- | --- |
| IG-001 | One or more of the ten job kinds lacks an adopted exact unit | JS-06D does not start; no unit is inferred | CLOSED — all ten exact v2 bindings pass owner and migration evidence |
| IG-002 | Preflight finds unresolved nonterminal mappings or unsafe invalid progress | Startup/cutover remains closed until owner-authorized resolution | CLOSED — invalid corpora reject before mutation; valid retained data pass migration and startup validation |
| IG-003 | Old illegal progress sequences remain replayable | Full conformance claim waits for replay expiry or Collaboration-owner remediation | CLOSED — illegal 24-hour replay rejects; the cutover rehearsal proves an empty accepted horizon |
| IG-004 | Old writers/runners or active leases cannot be drained | Migration and dequeue reopening do not proceed; two-phase rollout requires a separate specification | CLOSED — active-lease preflight rejects and the drained rehearsal keeps dequeue closed until the corrected writer is ready |

## 12. Binary Completion Criteria

### Tracker completeness

| Acceptance ID | Binary criterion | Requirement coverage | Result |
| --- | --- | --- | --- |
| JP-AC-001 | Target, authority, identity namespaces, and prohibited scope are explicit. | JP-REQ-001 through JP-REQ-003 | PASS |
| JP-AC-002 | Every target file and every discovered production jobs writer is inventoried. | JP-REQ-004 | PASS |
| JP-AC-003 | Public HTTP, WS, authorization, progress, telemetry, storage, generated, frontend, and harness surfaces have owners and evidence posture. | JP-REQ-002 | PASS |
| JP-AC-004 | `platform.jobs` ownership and cross-owner evidence allocation are unambiguous. | RB-001 | PASS |
| JP-AC-005 | The complete state matrix, worker/recovery sequencing, cancellation behavior, and transition interface are specified. | JP-REQ-006 through JP-REQ-010 | PASS |
| JP-AC-006 | Progress unit grammar, ownership boundary, ten-kind mapping obligation, update matrix, success rule, atomicity, and errors are specified. | JP-REQ-011 through JP-REQ-015 | PASS |
| JP-AC-007 | Foreign persistence, external Jobs writer, composition, and boundary guardrails have explicit dispositions. | JP-REQ-004, JP-REQ-005, JP-REQ-016 | PASS |
| JP-AC-008 | Preflight, backfill, invalid-data, replay, and deployment defaults are deterministic and fail closed. | JP-REQ-017 through JP-REQ-020 | PASS |
| JP-AC-009 | Every slice has dependencies, validation, rollback, and a binary completion condition. | Sections 6-8 | PASS |
| JP-AC-010 | Prior handoff history is preserved and this revision supplies a safe restart point. | Sections 9-11 | PASS |
| JP-AC-011 | No owner contradiction or unresolved planning question remains; resolved decisions are not mislabeled implementation-complete. | Section 11 | PASS |
| JP-AC-012 | Only this tracker is changed by the revision task, and no generated file hand edit is planned. | Section 1 | PASS |

### Refactor completion

| Acceptance ID | Binary completion condition | Current result | Actual evidence |
| --- | --- | --- | --- |
| JP-AC-101 | `platform.jobs` is active, unique, selects complete owner inventories, and has no overlapping selectors. | PASS | Owner explain/task-guide resolve nine active rows; complete unit/service-backed roots are `.cartulary/test-results/20260808T060830Z-p3402740` and `.cartulary/test-results/20260808T060855Z-p3405169`. |
| JP-AC-102 | Core clarifications and all ten exact owner-declared progress units are adopted and projected. | PASS | JS-06A owner/projection evidence is retained; final generation and drift roots are `.cartulary/test-results/20260808T060801Z-p3398673` and `.cartulary/test-results/20260808T061918Z-p3686587`. |
| JP-AC-103 | All required transition/progress/concurrency/recovery/public-parity tests pass. | PASS | Complete Jobs and all affected owner inventories pass; stateful browser root `.cartulary/test-results/20260808T062214Z-p3755259` passes 36/36. |
| JP-AC-104 | One Jobs-owned atomic substrate governs every production state/progress write. | PASS | Complete Jobs evidence passes; final canonical boundary root `.cartulary/test-results/20260808T061956Z-p3690377` rejects external writes and locks. |
| JP-AC-105 | Migration preflight/backfill/constraints pass and every mutable job has an immutable unit. | PASS | Drained cutover root `.cartulary/test-results/20260808T060812Z-p3401018`, complete Jobs storage rows, and migration drift root `.cartulary/test-results/20260808T061905Z-p3683805` pass. |
| JP-AC-106 | Queued work reaches running before execution; cancellation passes through cancel-requested; terminal/progress invariants hold. | PASS | Complete Jobs lifecycle/runner evidence and the cutover recovery handler assert running-before-invocation; Job API and browser evidence pass. |
| JP-AC-107 | Jobs contains no Auth/Extensions-owned storage access and no production jobs-table writer bypass exists. | PASS | JS-02 source scans and owner rows pass; JS-05 positive/negative fixtures and final boundary root `.cartulary/test-results/20260808T061956Z-p3690377` are green. |
| JP-AC-108 | Public HTTP/WS/auth behavior and generated public shapes remain compatible and drift-clean. | PASS | Final Job API roots `.cartulary/test-results/20260808T062908Z-p3930691` and `.cartulary/test-results/20260808T062923Z-p3932737`, Collaboration roots, JSON shape, generation drift, and stateful browser evidence pass. |
| JP-AC-109 | Old writers are drained, replay compatibility is resolved, corrected recovery succeeds, and dequeue is safely reopened. | PASS | The isolated deployment-faithful root `.cartulary/test-results/20260808T060812Z-p3401018` proves zero active leases, an accepted empty 24-hour replay horizon, migration/startup validation before admission, and corrected recovery after reopening. No external deployment state is inferred. |
| JP-AC-110 | Required owner slices, migration/generation/boundary gates, `make agent-finalize`, and `make check` pass with recorded artifacts. | PASS | Finalization root `.cartulary/test-results/20260808T062931Z-p3934214` and inspected final check root `.cartulary/test-results/20260808T063551Z-p4119107` pass; retained-run maintenance was skipped because `RESULTS_DIR` was unset. |

The jobs-platform remediation is complete at the repository and
deployment-rehearsal boundary. An external deployment must still execute the
recorded drained sequence against its own retained state; this handoff does not
claim an unobserved production rollout.
