# jobs-platform Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

This tracker is a planning and handoff contract. Normative terms such as
`MUST`, `MUST NOT`, `SHOULD`, and `MAY` prescribe the work that a later
authorized implementation must perform; they do not elevate this tracker above
an adopted owner document. If this tracker and an adopted owner conflict, the
owner controls and the implementation work is `BLOCKED: owner contradiction`.

| Item | Required posture |
| --- | --- |
| Target path | `internal/platform/jobs` |
| Target label | `jobs-platform`; lowercase kebab case with no spaces, separators, shell metacharacters, or unsafe filename characters |
| Output path | `docs/handoffs/jobs-platform-module-refactor-tracker.md` |
| Original repository baseline | Branch `main`, commit `7a39756616f12d06b5d1c981295acc35e517cc1d`; the worktree was clean before the original tracker was created |
| Current task | Revise this tracker into an NLSpec-grade, decision-complete handoff |
| Permitted change | This tracker file only |
| Prohibited changes | Production code, tests, adopted owners, contracts, migrations, generated artifacts, package configuration, and harness inputs or outputs |
| Implementation status | Not started; every production, owner, contract, migration, test, generation, and harness change requires a later authorized task |

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
MUST NOT enter HTTP, WebSocket, frontend, or telemetry surfaces.

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
| `internal/platform/jobs/.gitkeep` | Empty placeholder in a populated package | None | None | None | None | None | `platform_jobs` housekeeping | low | Non-runtime; optional deletion remains deferred. |
| `internal/platform/jobs/collaboration_producer.go` | Transaction-scoped job creation/terminal operations and canonical incident `job_progress` intent production | `ProgressIntent`, `ProgressIntentAppender`, `TransactionService`, constructor, and transaction methods | Server and test-support composition; imports, incident bundles, reporting, reference data, and extension finalizers | `pgx`, jobs lifecycle helpers, injected intent appender | Progress persistence and rollback tests in `jobs_test.go` | WebSocket job-progress contract and generated WS catalog | `platform_jobs` producer with application-owned Collaboration translation | high | Existing consumer-owned port direction is intentional. |
| `internal/platform/jobs/durable.go` | Handler payload access, recovery selection, claims, leases, attempts, error recording, and exhaustion failure | Durable Manager methods | `Runner` and tests | PostgreSQL jobs rows, lifecycle transitions, progress intents | Runner dispatch, recovery, and exhaustion tests | Jobs migrations and recovery catalog | `platform_jobs` | critical | Claim currently acquires a lease without atomically changing queued status to running. |
| `internal/platform/jobs/extensions.go` | Extension job admission, contract lookup, terminal-result validation, and canonical encoding | `ExtensionResourceRefContract`, `ExtensionJobContract`, contract configuration/lookup, canonical helpers | Extension assembly, extension store/finalizers, extension-capable producers | Generated owner facts supplied through application composition | `extensions_test.go` and extension rows | Extension job-kind fragments and generated bindings | Generic `platform_jobs` mechanism constrained by Extensions/Reporting owners | high | Future projection must carry owner-declared `progress_unit_id`. |
| `internal/platform/jobs/extensions_test.go` | Characterizes extension admission and terminal resource contracts | Two package tests | Go runner and semantic family rows | Jobs public API and fixtures | Self | Existing `module.extensions` unit row | `module.extensions` verification evidence | medium | These tests remain Extensions-owned by postcondition. |
| `internal/platform/jobs/jobs.go` | Public job DTOs/errors; Manager; creation, lookup, running/terminal transitions, cancellation, idempotency, and extension cancellation observation | `Manager`, DTOs, admission/idempotency types, lifecycle methods, and `LockTransitionTx` | Job API, server assembly, modules, extension store, runner, and test support | Jobs table, Auth-owned `route_idempotency`, Extensions-owned cancellation observation, telemetry | Manager and Job API tests | OpenAPI/TypeScript job resource, migrations, ownership manifest | Primarily `platform_jobs`; two foreign-owner persistence operations are misplaced | critical | Contains direct queued terminal writes and incomplete progress enforcement. |
| `internal/platform/jobs/jobs_test.go` | Characterizes cancellation replay, terminal resource retention, progress-intent atomicity, runner dispatch/recovery, configuration, and exhaustion | Eight package tests | Go runner and semantic family rows | PostgreSQL test support and jobs API | Self | Some rows map to `module.extensions`; six tests were previously unmapped | `platform.jobs` plus existing cross-owner evidence | high | Queued recovery currently characterizes a direct queued-to-success path and must be corrected, not preserved. |
| `internal/platform/jobs/recovery_state.go` | Declares the authoritative `jobs` recovery contribution | `RecoveryStateContribution` | Recovery assembly | Recovery contribution contract | Recovery catalog/assembly tests | Recovery state catalog | `platform_jobs` | medium | Legitimate thin owner contribution. |
| `internal/platform/jobs/runner.go` | Handler registration, dequeue admission, dispatch, recovery activation, panic/error handling, and shutdown | Runner errors, `HandlerFunc`, `DequeueGate`, `Runner`, and lifecycle methods | Server assembly and module handler registration | Manager durable operations, concurrency, UUIDs | Runner tests and worker integration rows | Named portability recovery evidence | `platform_jobs` | critical | Handler invocation must occur only after a committed running claim. |
| `internal/platform/jobs/telemetry.go` | Jobs tracing, duration and active metrics, active-job query, and safe tokens | Private jobs instrumentation helpers | Manager and Runner | OpenTelemetry API and jobs table | `telemetry_test.go` | Adopted OpenTelemetry vocabulary | `platform_jobs` | medium | `progress_unit_id` must never become a telemetry attribute. |
| `internal/platform/jobs/telemetry_test.go` | Characterizes vocabulary, safe tokens, and no-SDK ownership | Three package tests | Go runner | Telemetry helpers | Self | No current semantic row | `platform.jobs` verification evidence | medium | Must be routed after `platform.jobs` is registered. |

Every target entry is inventoried; none is silently excluded.

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

### Progress-unit contract

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
| `import.apply_v1` | Import contract owner (`core01` fragment) | `contracts/extensions/fragments/core01.profile-jobs.json` | Exact `progress_unit_id` MUST be adopted before migration; no fallback |
| `import.discovery_v1` | Import contract owner (`core01` fragment) | `contracts/extensions/fragments/core01.profile-jobs.json` | Exact `progress_unit_id` MUST be adopted before migration; no fallback |
| `incident_portability.export_v1` | Incident portability contract owner (`core01` fragment) | `contracts/extensions/fragments/core01.profile-jobs.json` | Exact `progress_unit_id` MUST be adopted before migration; no fallback |
| `incident_portability.import_v1` | Incident portability contract owner (`core01` fragment) | `contracts/extensions/fragments/core01.profile-jobs.json` | Exact `progress_unit_id` MUST be adopted before migration; no fallback |
| `reference_pack.import_v1` | Reference-pack contract owner (`core01` fragment) | `contracts/extensions/fragments/core01.profile-jobs.json` | Exact `progress_unit_id` MUST be adopted before migration; no fallback |
| `reference_pack.refresh_v1` | Reference-pack contract owner (`core01` fragment) | `contracts/extensions/fragments/core01.profile-jobs.json` | Exact `progress_unit_id` MUST be adopted before migration; no fallback |
| `reference_pack.reverify_v1` | Reference-pack contract owner (`core01` fragment) | `contracts/extensions/fragments/core01.profile-jobs.json` | Exact `progress_unit_id` MUST be adopted before migration; no fallback |
| `snapshot_reporting.composition_preview_v1` | Reporting | `contracts/extensions/fragments/snapshot_reporting.participation.json` | Exact `progress_unit_id` MUST be adopted before migration; no fallback |
| `snapshot_reporting.release_create_v1` | Reporting | `contracts/extensions/fragments/snapshot_reporting.participation.json` | Exact `progress_unit_id` MUST be adopted before migration; no fallback |
| `snapshot_reporting.snapshot_create_v1` | Reporting | `contracts/extensions/fragments/snapshot_reporting.participation.json` | Exact `progress_unit_id` MUST be adopted before migration; no fallback |

The unit MUST traverse the owner-to-storage boundary as follows:

| Source or boundary | Required representation | Validation responsibility |
| --- | --- | --- |
| Adopted job-kind owner input | Required `progress_unit_id` member in each job-kind contract | Named job-kind owner defines meaning and version |
| Extensions runtime projection | Required `ProgressUnitID` on the projected `extensions.JobKindContract` | Extensions projection validates presence, grammar, and canonical digest participation |
| Application-to-Jobs projection | Required `ProgressUnitID` on `jobs.ExtensionJobContract` | Application composition copies the exact owner value without reinterpretation |
| Jobs creation input | Required internal `ProgressUnitID` in the semantic queued-job creation input | Jobs validates grammar and, for an admitted extension job, exact equality with the configured job-kind contract |
| Jobs row | Required `progress_unit_id` for new/mutable rows | Jobs persists the exact value at creation and never changes it |
| Public resource/event | No member | HTTP, WS, TypeScript protocol, and telemetry surfaces reject accidental projection |

An extension job with a missing, unknown, or mismatched unit MUST fail admission
as an invalid internal job definition before row insertion. A non-extension test
or internal caller MUST supply an explicitly registered Jobs-owned unit; an
arbitrary caller string is not an owner declaration.

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
| `.gitkeep` remains in a populated directory. | Directory inspection | No runtime risk | `defer` | Housekeeping | Optional JS-07 cleanup only. |

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
| JS-01 | none | Register active `platform.jobs`; add verification contract and semantic lifecycle/durable/telemetry families; route exact tests without overlaps. | Authored verification/catalog inputs; generated topology via generator | Misassignment by file location or recycled identity | Selector uniqueness, owner closure, existing Extensions/Job API ownership | `make generate`; owner explain/guide; complete `platform.jobs` slices | Revert authored owner inputs and regenerate; no product code changes. | Owner is active/unique; every target test resolves exactly once; generated outputs are drift-clean. |
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

- Incomplete succeeded rows with known totals age out under the existing
  retention policy by default. One-time normalization requires separate owner
  authorization and an affected-count record.
- Negative completed values, nonpositive totals, and completed-over-total rows
  MUST NOT be repaired by inference.
- Still-replayable illegal Collaboration history ages out under the bounded
  replay horizon by default. Only a Collaboration-owner remediation may alter
  history.

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
| JP-005 | Register `platform.jobs` and route evidence | WF-03 | TODO | JP-004 | JS-01 | Owner commands accept the ID and all tests resolve exactly once. |
| JP-006 | Adopt Core clarifications and exact job-kind unit declarations | WF-04 | TODO | JP-005 | JS-06A | Owners contain every required behavior and all ten unit mappings. |
| JP-007 | Add owner-aligned correction evidence | WF-05 | TODO | JP-006 | JS-06B | Required tests fail on old drift and pass on corrected behavior. |
| JP-008 | Implement central transition/progress substrate | WF-05 | TODO | JP-007 | JS-06C | One internal matrix and typed error boundary governs state/progress. |
| JP-009 | Apply progress-unit migration and compatibility gates | WF-05 | TODO | JP-008 | JS-06D | Mutable rows have units, constraints validate, and unsafe data fails closed. |
| JP-010 | Migrate every worker, finalizer, recovery path, and external writer | WF-05 | TODO | JP-009 | JS-06E | No production jobs write bypass remains and complete correction evidence passes. |
| JP-011 | Extract Auth/Extensions persistence ports | WF-06 | TODO | JP-010 | JS-02 | Jobs contains no foreign-owner table access. |
| JP-012 | Split Jobs private internals | WF-06 | TODO | JP-011 | JS-03 | Stable facade hides cohesive private concerns. |
| JP-013 | Narrow concrete consumers | WF-06 | TODO | JP-012 | JS-04 | Each consumer receives only required capabilities. |
| JP-014 | Add boundary anti-regression rules | WF-07 | TODO | JP-011, JP-013 | JS-05 | Canonical boundary checks reject all defined bypasses. |
| JP-015 | Execute drained cutover and final evidence | WF-08 | TODO | JP-010, JP-014 | JS-07 | Corrected single writer is serving and every implementation criterion passes. |
| JP-016 | Remove `.gitkeep` | WF-08 | DEFERRED | JP-012 | Optional JS-07 cleanup | File is removed in a scoped cleanup or deferral remains recorded. |

## 10. Session Handoff Log

The first row in each table preserves the original 2026-08-07 planning handoff.
The second row records this NLSpec-grade revision. Files touched remain limited
to this tracker; source, owner, code, contract, migration, generated, and harness
files were inspected only.

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 America/New_York | Codex original planning session | WF-00 through WF-02 planning complete | Inspected framework, Core 00-04, subsystem NLSpecs, domain and support guides; touched only this tracker | `git status`, `git branch --show-current`, `git rev-parse`, `sed`, `rg` | Target exists; label is safe; baseline was clean; no owner contradiction found | None for tracker completion | Seek later authorization for JS-01 before implementation. |
| 2026-08-07 America/New_York | Codex NLSpec revision session | Requirements and owner boundaries are decision-complete | Inspected `analysis-notes.md`, `nlspec-spec.md`, current tracker, owner inputs, and adopted sources; touched only this tracker | `git status`, `sha256sum`, `sed`, `rg`, `jq`, `git log` | RB-001 through RB-003 resolved at planning level; no owner contradiction; implementation remains pending | None at planning level | Execute JS-01 in a separately authorized task. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 America/New_York | Codex original planning session | Legitimate platform boundary with two foreign-storage leaks and mixed internals | Inspected every target file, server/runtime dependencies, Job API, extension store/assembly, imports, incident bundles, reporting, and reference data; touched only this tracker | `rg`, `sed`, `find`, `git ls-files` | Keep Jobs persistence/runner; move Auth/Extensions SQL behind owner-backed ports; split later | Evidence routing and implementation authority | Characterize before structural extraction. |
| 2026-08-07 America/New_York | Codex NLSpec revision session | Central writer/transition requirement added | Re-inspected target writes, module workers, and `internal/platform/extensionstore/reconciliation.go`; touched only this tracker | `rg`, `sed` | Extensions reconciliation is a third production write boundary and must migrate in JS-06E | No planning blocker | Register owner, adopt requirements, test, then centralize all writers. |

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

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 America/New_York | Codex original planning session | Target tests partly route through Extensions; Job API owner active; platform owner absent | Inspected tests, owners, Job API/Extensions families, and harness owner; touched only this tracker | `make help`, `make help-all`, task/explain targets, `make lint-markdown`, `rg`, `sed`, `jq` | Existing owners discovered; invalid attempted IDs rejected; original Markdown lint passed | Original RB-001 | Harness owner decides/creates semantic owner. |
| 2026-08-07 America/New_York | Codex NLSpec revision session | `platform.jobs` selected; implementation not authored | Inspected owner grammar, current/historical registries, platform owner examples, and exact target tests; touched only this tracker | `rg`, `sed`, `jq`, `git log`, `make lint-markdown` | ID is grammar-valid, currently unused, and not found in owner history; Markdown lint passed; no product or owner slice ran | None at planning level | JS-01 authors owner inputs and runs complete owner evidence. |

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

## 11. Open Questions and Blockers

No unanswered planning question remains. RB-001 through RB-003 are resolved
decisions, not completed implementation. They MUST NOT be marked `DONE` until
their closure evidence exists.

| ID | Resolution | Required closure evidence | Current status |
| --- | --- | --- | --- |
| RB-001 | The semantic harness owner is `platform.jobs`. Generic lifecycle, progress, durable runner, recovery, and Jobs telemetry evidence belongs to it; cross-owner postconditions retain their existing owners. | Active registries/contracts/families; exact selectors; owner explain/guide; complete all/service-backed slices; drift-clean generated outputs | RESOLVED — implementation evidence required |
| RB-002 | Core 01 controls. Direct queued-to-succeeded/canceled transitions are prohibited; dispatch/recovery claims running before work; cancellation passes through cancel-requested; one atomic Jobs substrate governs all writers. | Exhaustive matrix, concurrent claim, recovery sequencing, cancellation race, atomicity, public polling/WS sequence, and drained-writer evidence | RESOLVED — implementation evidence required |
| RB-003 | Progress rejects completed regression, total clearing/regression, unit change, bound violation, and incomplete known-total success. Every mutable job persists an immutable owner-declared internal unit. | Adopted owner clarification/unit map; migration/backfill; constraints; concurrent progress, success, recovery, HTTP/WS, and extension parity evidence | RESOLVED — implementation evidence required |

### Implementation evidence gates

| Gate | Condition that prevents completion | Fail-closed result | Current status |
| --- | --- | --- | --- |
| IG-001 | One or more of the ten job kinds lacks an adopted exact unit | JS-06D does not start; no unit is inferred | TODO |
| IG-002 | Preflight finds unresolved nonterminal mappings or unsafe invalid progress | Startup/cutover remains closed until owner-authorized resolution | TODO |
| IG-003 | Old illegal progress sequences remain replayable | Full conformance claim waits for replay expiry or Collaboration-owner remediation | TODO |
| IG-004 | Old writers/runners or active leases cannot be drained | Migration and dequeue reopening do not proceed; two-phase rollout requires a separate specification | TODO |

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

| Acceptance ID | Binary completion condition | Current result |
| --- | --- | --- |
| JP-AC-101 | `platform.jobs` is active, unique, selects complete owner inventories, and has no overlapping selectors. | NOT MET |
| JP-AC-102 | Core clarifications and all ten exact owner-declared progress units are adopted and projected. | NOT MET |
| JP-AC-103 | All required transition/progress/concurrency/recovery/public-parity tests pass. | NOT MET |
| JP-AC-104 | One Jobs-owned atomic substrate governs every production state/progress write. | NOT MET |
| JP-AC-105 | Migration preflight/backfill/constraints pass and every mutable job has an immutable unit. | NOT MET |
| JP-AC-106 | Queued work reaches running before execution; cancellation passes through cancel-requested; terminal/progress invariants hold. | NOT MET |
| JP-AC-107 | Jobs contains no Auth/Extensions-owned storage access and no production jobs-table writer bypass exists. | NOT MET |
| JP-AC-108 | Public HTTP/WS/auth behavior and generated public shapes remain compatible and drift-clean. | NOT MET |
| JP-AC-109 | Old writers are drained, replay compatibility is resolved, corrected recovery succeeds, and dequeue is safely reopened. | NOT MET |
| JP-AC-110 | Required owner slices, migration/generation/boundary gates, `make agent-finalize`, and `make check` pass with recorded artifacts. | NOT MET |

The tracker is complete only as a planning artifact. The jobs-platform refactor
is incomplete until JP-AC-101 through JP-AC-110 all read `PASS` with actual
evidence.
