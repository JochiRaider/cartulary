# Jobs Platform Legacy Removal and Production-Readiness Tracker

## 1. Control, Authority, and Current Posture

This is the controlling execution and handoff artifact for the second Jobs
Platform refactoring iteration. It plans removal of legacy and dead internal
surfaces, hardens durable execution, and moves the subsystem toward a
production-ready operating posture.

Normative terms in this tracker describe changes that must be adopted by the
applicable specification owner before implementation. This document does not
supersede Core or an adopted subsystem NLSpec. If a planned requirement
conflicts with an adopted owner, the active slice is `BLOCKED` until the owner
is corrected or the plan is revised.

| Item | Current value |
| --- | --- |
| Target package | `internal/platform/jobs` |
| Semantic verification owner | `platform.jobs` |
| Storage/implementation refinement | `platform_jobs` |
| Public transport owner | `module.jobapi` |
| Planning baseline | Branch `main`, commit `a43f9ddf6fe83023007ba6deed6cc6a5b5676a3d` |
| Baseline worktree | Clean at planning bootstrap on 2026-08-08 |
| Iteration sequence | `JPR-01` -> `JPR-02` -> `JPR-03` -> `JPR-04` -> `JPR-05` -> `JPR-06` -> `JPR-07` |
| Iteration status | `JPR-01 DONE`; `JPR-02 DONE`; `JPR-03 DONE`; `JPR-04 DONE`; `JPR-05 DONE`; `JPR-06 DONE`; `JPR-07 DONE`; no active slice |
| Document-update scope | Tracker, adopted owners, authored verification routing, Jobs composition, server/consumer assembly, and tests; migrations and generated roots remain unchanged through JPR-02 |
| Public compatibility | HTTP, WebSocket, authorization, error, status, and `{completed,total}` shapes remain unchanged |
| Runtime compatibility | Clean internal cutover; no mixed writers or compatibility shims |
| Data posture | Pre-release; unsupported legacy rows require an explicitly approved reset/reseed, never inferred repair |

The authority order remains:

1. Adopted subsystem NLSpecs within their named scopes.
2. Core 00 through Core 04 for product and implementation conformance.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. Typed contracts as projections of their adopted owners.
5. Repository code and tests as implementation evidence.
6. This tracker as planning, execution, and handoff evidence only.

`docs/domain.md` remains unchanged. Job leases, attempts, progress units,
supervision, and package identities remain implementation/specification
details delegated to Core and subsystem owners.

### Checkpoint protocol

The only authorized implementation order is:

`JPR-01` -> `JPR-02` -> `JPR-03` -> `JPR-04` -> `JPR-05` -> `JPR-06` ->
`JPR-07`.

For every future slice:

1. Mark only that slice `IN_PROGRESS`.
2. Implement and validate only that slice.
3. Record changed files, commands, retained result roots, failures, rollback
   posture, and remaining blockers.
4. Mark the slice `DONE` only when every binary exit criterion passes;
   otherwise mark it `BLOCKED`.
5. Run `make lint-markdown` after updating this tracker.
6. Do not begin the next slice until the checkpoint is recorded and
   lint-clean.

Planned or expected results never count as completion evidence. JPR-07 must
update every JPR acceptance criterion from actual retained evidence.

## 2. Completed Iteration Baseline

The preceding Jobs Platform specification and implementation remediation is
complete at commit `a43f9ddf6fe83023007ba6deed6cc6a5b5676a3d`. That iteration established:

- active and unique `platform.jobs` verification ownership;
- the adopted v2 job-kind contract and ten exact progress-unit bindings;
- one private lifecycle/progress transition substrate;
- catalog-backed `job_kind` and private `progress_unit_id` storage;
- running-before-handler claims and unique dispatch attempts;
- closed safe handler-failure diagnostics;
- owner transaction ports for Auth and Extensions persistence;
- Jobs-owned mutable-state writes and runnable validation;
- narrow consumer capabilities and executable boundary guards; and
- a drained v1-to-v2 cutover rehearsal with compatible public behavior.

The detailed JS-01 through JS-07 ledger, prior JP requirements, and JP-AC-101
through JP-AC-110 evidence remain available in that commit and are historical,
not an active plan. Durable final evidence includes:

| Evidence | Retained result root | Result |
| --- | --- | --- |
| Complete Jobs unit slice | `.cartulary/test-results/20260808T060830Z-p3402740` | Pass |
| Complete Jobs service-backed slice | `.cartulary/test-results/20260808T060855Z-p3405169` | Pass |
| Drained cutover rehearsal | `.cartulary/test-results/20260808T060812Z-p3401018` | Pass |
| Migration drift | `.cartulary/test-results/20260808T061905Z-p3683805` | Pass |
| Generation drift | `.cartulary/test-results/20260808T061918Z-p3686587` | Pass |
| Backend boundary check | `.cartulary/test-results/20260808T061956Z-p3690377` | Pass |
| Stateful browser evidence | `.cartulary/test-results/20260808T062214Z-p3755259` | Pass, 36/36 |
| Agent finalization | `.cartulary/test-results/20260808T062931Z-p3934214` | Pass; retained-run maintenance skipped because `RESULTS_DIR` was unset |
| Full check | `.cartulary/test-results/20260808T063551Z-p4119107` | Pass, 730/730; inspected with `make explain-run` |

This iteration starts from that corrected substrate. It must not recreate the
superseded writers, v1 runtime reader, broad dependencies, or transition
bypasses removed by the completed iteration.

## 3. Current-State Inventory and Gap Disposition

### Current Jobs package

| Concern | Current implementation homes | Current responsibility | Next-iteration disposition |
| --- | --- | --- | --- |
| Public internal values | `public_types.go`, `owner_ports.go` | Common job values, commands, sentinels, and owner transaction ports | Keep public resource values; replace broad command unions and duplicate identity inputs with typed operations. |
| Composition facade | `manager.go`, `transaction_service.go` | Configurable Manager plus transaction-scoped admission/finalization | Replace zero-value/mutable setup with immutable error-returning construction and one shared catalog. |
| Definition catalog | `definition_catalog.go`, `extension_admission.go`, `extension_terminal.go` | Extension-derived definition facts, admission evidence, and terminal proof validation | Store generic immutable definitions; keep extension policy optional and owner-specific. |
| Lifecycle persistence | `lifecycle_persistence.go`, `transition_policy.go`, `transition_persistence.go` | Creation, reads, progress/state transitions, cancellation, and publication | Retain the private matrix; require execution fencing for handler-owned changes and logical expiry for public reads. |
| Stored projection | `stored_job.go` | Private storage identity and projection to unchanged public resource | Add execution/tombstone fields without exposing them publicly. |
| Durable persistence | `durable_persistence.go` | Payload, recovery selection, claims, leases, attempts, and failure closure | Replace unrenewed leases and claim-count retry semantics with fenced renewable attempts and persisted retry eligibility. |
| Runner | `runner.go` | Handler registration, one-shot recovery, goroutine dispatch, and shutdown | Replace with a bounded continuously supervised runner. |
| Telemetry | `telemetry.go` | Enqueue/terminal spans, active gauge, duration metric, safe tokens | Correct actual-kind aggregation and instrument full attempts, renewal failure, and expiry safely. |
| Recovery declaration | `recovery_state.go` | Jobs recovery-state contribution | Retain; update only if authored schema objects change. |
| Evidence | `correction_test.go`, `cutover_test.go`, `extensions_test.go`, `jobs_test.go`, `telemetry_test.go` | Lifecycle, storage, security, cutover, extension, and telemetry tests | Replace obsolete compatibility characterization and add supervision/fencing/expiry evidence. |

### Legacy and dead surface classification

| Surface | Evidence at baseline | Decision | Rationale and long-term benefit | Compatibility impact | Risk if retained |
| --- | --- | --- | --- | --- | --- |
| `NewManager()` followed by `Configure`, `ConfigureTelemetry`, and `ConfigureExtensionContracts` | Production server assembly performs phased mutable setup | Remove | One constructor makes invalid partial states unrepresentable and removes catalog/telemetry ordering races. | Internal composition break only. | Runtime can be observed or used before all dependencies and the catalog are ready. |
| `NewRunner()` followed by `Configure` and `ConfigureDequeueGate` | Production assembly mutates the runner after construction | Remove | Immutable dependencies give startup one fail-fast boundary. | Internal composition and tests change. | Nil/default paths remain reachable and readiness becomes harder to reason about. |
| `Manager.Create` | No production caller; retained for HTTP test support | Remove | Tests must use the real transaction admission path rather than preserve production API for fixtures. | Test-support migration. | A second non-owner transaction path remains indefinitely. |
| `Manager.ExtensionContract` | No production caller | Remove | Catalog data remains private and consumers receive only semantic results. | Tests use catalog construction evidence. | Future callers can couple to owner metadata instead of needed behavior. |
| `Runner.Dispatch` | Anonymous work exists only in characterization tests | Remove | All production work remains named, durable, recoverable, and fenced. | Replace obsolete test. | Unmanaged anonymous work can bypass job storage and recovery. |
| `Runner.DispatchJob(string)` | Reporting and composition preview parse string IDs | Remove | UUID-typed notification eliminates compatibility parsing and caller-supplied handler routing. | Internal callers migrate atomically. | Invalid identifier and handler-name coupling persists. |
| `Runner.RecoverHandler` | Modules perform construction-time one-shot recovery | Remove | A single long-lived supervisor becomes authoritative for initial and continuing recovery. | Module recovery methods disappear. | Failed attempts remain stuck until restart. |
| `Runner.ValidateConfiguration` | Incident Portability checks mutable setup | Remove | Constructor success is the configuration proof. | Startup tests assert constructor failure. | Configuration state remains mutable and duplicated. |
| Variadic definitions in `NewTransactionService` | Used by test support | Remove | Every service shares exactly one immutable complete catalog. | Test helpers construct the catalog explicitly. | Partial catalogs and production/test drift remain possible. |
| Duplicate handler tolerance | Imports and Reference Data ignore `ErrHandlerAlreadyRegistered` | Remove | One handler owns one name and duplicate composition fails before readiness. | Startup now fails instead of silently accepting duplication. | Assembly errors remain hidden and handler ownership is ambiguous. |
| `retiredProfileHandler` denylist | Hard-coded aliases supplement exact catalog matching | Remove | Exact catalog admission is the single handler authority. | Retired aliases remain rejected as unknown definitions. | Every retired alias adds permanent compatibility code. |
| `CreateParams.JobKind` plus `ExtensionJobAdmission.JobKind` | Caller supplies the same semantic identity twice | Remove duplication | One enqueue identity prevents disagreement and simplifies future non-extension definitions. | Internal producer migration. | Callers can create inconsistent identity combinations. |
| `MarkRunning` from handlers | Durable claim already commits `running` before invocation | Replace with fenced progress update | Claim owns execution start; handlers own only progress and terminal work. | Worker interfaces and tests change. | Workers can recreate a queued-to-running bypass outside attempt ownership. |
| `TransitionParams` union | One type admits result/error combinations for every terminal operation | Replace with typed success/failure/cancellation completions | Invalid combinations become unrepresentable and call sites state intent clearly. | Internal compile-time break. | Every caller must understand validation combinations that the type permits. |
| Raw reconciliation status and terminal JSON | Extensions passes open strings and JSON after a separate validation call | Replace with an opaque grant and typed outcome | Preserves owner transaction atomicity while preventing protocol misuse. | Extensions adapter changes. | Callers can mismatch validation, state, and terminal payload. |
| Nullable terminal `job_kind` and `progress_unit_id` | Migration 58 permits immutable legacy terminal rows to remain null | Remove | Every retained row has durable semantic identity and storage constraints become simpler. | Unsafe pre-release databases require approved reset/reseed. | Every query and future migration must carry a legacy-null branch. |

### Intentionally retained versioned behavior

The following are current contracts, not legacy cleanup candidates:

- public job route and WebSocket shapes;
- the six public status tokens and existing authorization behavior;
- `cartulary.extension_job_kind_contract.v2`;
- the ten current job-kind values ending in `_v1`;
- `cartulary.route_scoped_idempotency_identity.v1`;
- all authored migration history;
- Auth and Extensions owner transaction ports;
- Collaboration intent translation at application composition; and
- the `internal/platform/jobs` package boundary.

## 4. Target Production Contract

### Fixed runtime policy

The policy is one immutable typed Jobs value. Production assembly uses these
exact values; tests may inject shorter durations through test-only
composition. This iteration does not change deployment-config schema v2.

| Policy member | Value | Meaning |
| --- | --- | --- |
| Handler lease | 30 seconds | Maximum ownership interval without successful renewal. |
| Lease renewal | 10 seconds | Renewal cadence while a handler attempt is active. |
| Recovery scan | 5 seconds | Continuous eligible-job scan cadence. |
| Recovery batch | 100 rows | Maximum candidates selected in one scan. |
| Handler concurrency | 8 | Maximum in-flight attempts across all handler kinds. |
| Maximum failures | 3 | Failure count at which a mutable job fails closed. |
| Retry delays | 5 seconds, then 30 seconds | Persisted eligibility delay after failures one and two. |
| Expiry sweep | 5 minutes | Interval between compact-tombstone sweeps. |
| Expiry batch | 1,000 rows | Maximum rows compacted per transaction. |

### Target internal interfaces

- One immutable `Catalog` is constructed from `[]Definition` and injected into
  `TransactionService`, `Manager`, and `Runner`.
- `Definition` contains `JobKind`, `ProgressUnitID`, `HandlerName`, and an
  optional extension-owner policy. Current v2 owner facts populate that
  policy; future non-extension owners do not need fake extension metadata.
- `EnqueueParams.JobKind` is the sole job-definition selector. Extension
  admission contains owner and idempotency evidence but no second job kind.
- `Execution` contains the job ID and an opaque attempt token produced only by
  a successful claim. Handlers receive `Execution`, not an independently
  supplied job ID and handler name.
- Payload reads, execution observation, progress update, and handler-owned
  completion validate the same live unexpired attempt under the Jobs
  transaction lock.
- Success, failure, and cancellation use distinct typed completion values.
- Transactional owner finalizers validate `Execution` inside the caller's
  transaction before mutating owner state.
- Inactive reconciliation uses an opaque Jobs validation grant plus a typed
  terminal outcome; it does not accept an open status string or raw terminal
  JSON.
- Runner notification is `Notify(jobID uuid.UUID)`. It is a non-blocking
  acceleration hint; periodic recovery remains authoritative if a hint is
  dropped or admission is closed.

The public `jobs.Resource` projection remains unchanged. Attempt IDs,
progress-unit IDs, handler metadata, retry state, and tombstone state never
enter HTTP, WebSocket, OpenAPI, TypeScript, frontend, log, or telemetry
payloads.

### Execution and supervision behavior

1. Publication activation starts the runner after handlers are registered and
   the dequeue gate is open.
2. The runner performs one synchronous initial recovery scan. Failure closes
   readiness and publication.
3. A claim creates a unique attempt, commits `queued -> running` when needed,
   and returns an opaque `Execution` before invocation.
4. The supervisor renews the lease every ten seconds. Ownership loss or
   renewal failure cancels the attempt context.
5. Stale attempts cannot read payloads, publish progress, validate owner work,
   or complete the job even when a handler ignores cancellation.
6. Handler error, panic, incomplete nil return, or expiry of an attempt that
   was not conditionally released by graceful shutdown increments the failure
   count, stores only a closed safe reason, clears the attempt, and schedules
   the next retry. A graceful conditional release does not consume the failure
   budget.
7. The third failure completes the job failed with the existing safe exhausted
   summary.
8. A graceful shutdown stops new claims, cancels handlers, conditionally
   releases attempts, and drains within the application deadline. Released or
   expired attempts are recovered by the next process.
9. A supervisor panic or unexpected loop exit is `job_dequeue` component loss.
   A transient scan error remains inside the live supervisor with bounded
   retry and safe telemetry.

### Expiry and tombstones

- At `now >= retained_until`, Jobs read and cancellation operations return the
  existing masked not-found result, regardless of sweep progress.
- Only terminal rows may gain `expired_at`.
- Compaction is ordered by `retained_until`, then `job_id`, and uses a fixed
  cycle cutoff and exactly one transaction of at most 1,000 rows per sweep.
- Compaction clears handler payload, attempt/lease/retry diagnostics, message,
  result/error summaries, and extension idempotency evidence.
- It retains job identity, scope and ownership/provenance, catalog identity,
  terminal status, progress, and lifecycle timestamps.
- Expiry emits no `job_progress` event and performs no hard delete.
- Durable outputs and cross-owner provenance remain untouched. A future
  physical-deletion feature requires a separately adopted cross-owner
  retention contract.

## 5. Workstream Plan

### JPR-01 - Adopt production execution policy and evidence

Depends on: this planning revision.

Areas: Core 01, Core 04, OpenTelemetry NLSpec, verification contracts, tests,
generated harness topology, tracker.

Remediation:

- Adopt continuous recovery, bounded concurrency, renewable fenced attempts,
  persisted retry delays, logical expiry, compact tombstones, and safe
  supervisor diagnostics.
- Adopt the exact policy values in Section 4.
- Define the required `platform.jobs` verification families for immutable
  composition, execution fencing, supervision, retry, and expiry. Activate
  each verification contract and test-family row only in the implementation
  slice that adds its selected test.
- Record the tests that must prove one-shot recovery, unrenewed leases,
  unbounded dispatch, stale-attempt finalization, and post-expiry visibility
  wrong. Each later implementation slice observes its new test fail before the
  fix and ends with all routed rows green; JPR-01 does not retain red tests.

Rationale and long-term benefit: Behavior remains specification-owned and
gains executable evidence before internal APIs or storage change. Future job
families inherit one bounded execution model rather than copying worker loops.

Compatibility and migration: No public behavior changes in this slice. New
tests intentionally fail only their old-behavior assertions.

Risk if unresolved: Cleanup could remove APIs while retaining unsafe runtime
semantics, and later tests could bless a non-production-ready runner.

Binary exit:

- owner review is complete;
- selectors are unique and cross-owner rows retain their semantic owners;
- generation, generated-policy, and drift checks pass; and
- the active verification catalog has no orphan verification ID or selector;
  and existing routed evidence remains green.

### JPR-02 - Immutable composition and dead API removal

Depends on: JPR-01.

Areas: Jobs internals, server and test composition, tests, boundary policy,
tracker.

Remediation:

- Construct one immutable catalog and inject it into all Jobs services.
- Replace extension-only stored catalog values with generic definitions plus
  optional extension policy.
- Replace mutable setup with error-returning constructors requiring every
  dependency.
- Make job kind the sole enqueue identity.
- Remove the dead construction, create, lookup, dispatch, validation, alias,
  variadic-catalog, and duplicate-handler compatibility surfaces in Section 3.
- Make test support enqueue through the production transaction admission path.
- Add boundary rules rejecting reintroduced mutable setup or removed APIs.

Rationale and long-term benefit: Invalid partial states and redundant identity
sources become unrepresentable. The catalog can accept future non-extension
definitions without inventing extension ownership.

Compatibility and migration: Internal Go call sites and test harness
constructors break and migrate atomically. Public protocols and the active v2
owner contract do not change.

Risk if unresolved: Production and test construction remain different,
catalog setup remains order-sensitive, and dead exports attract new callers.

Binary exit:

- missing dependencies fail construction before readiness or goroutine start;
- caller mutation cannot alter the constructed catalog;
- no production or test-support source uses a removed symbol;
- target Jobs, app-server, and boundary evidence passes; and
- no compatibility facade remains.

### JPR-03 - Fenced execution substrate and schema cleanup

Depends on: JPR-02.

Areas: migration `00059_platform_jobs_execution_fencing.sql`, stored model,
durable/lifecycle persistence, typed APIs, owner finalizers, tests, tracker.

Remediation:

- Add opaque execution identity and conditional claim, renewal, observation,
  payload, progress, release, and completion operations.
- Replace terminal unions with typed success, failure, and cancellation
  commands.
- Replace inactive-reconciliation strings/JSON with an opaque grant and typed
  outcome.
- Atomically migrate every handler signature and Imports, Incident Portability,
  Reference Data, Reporting, Report Composition, and Extensions owner
  finalizer to consume the Jobs-owned execution fence.
- Rename lease ownership to attempt identity; replace attempt-count retry
  semantics with failure count, maximum failures, and `handler_next_attempt_at`.
- Require non-null `job_kind`, `progress_unit_id`, and `handler_name`.
- Preflight rejects active leases, `running` or `cancel_requested` rows, queued
  rows with prior execution metadata, incomplete definitions, and missing
  handler identity before schema mutation.
- Backfill or infer no semantic value. Unsupported rows require the explicit
  pre-release reset/reseed path.

Rationale and long-term benefit: Lease ownership becomes an enforceable fence,
not advisory bookkeeping. Typed completion and reconciliation APIs eliminate
illegal combinations and sequencing ambiguity.

Compatibility and migration: Drained internal schema break. Queued jobs survive
only when they have complete v2 identity and no prior execution metadata.
Downgrade is supported only before the corrected writer uses the new fields.

Risk if unresolved: An expired attempt can commit domain work or terminal state
after another worker acquires the job, and every query must retain legacy-null
branches.

Binary exit:

- fresh install and valid drained upgrade pass;
- every invalid corpus fails before partial mutation;
- stale, expired, and wrong attempts are mutation- and event-free;
- renewal/finalization races preserve one winner;
- owner transactions reject stale execution before owner mutation; and
- migration drift plus complete Jobs storage/execution evidence passes.

### JPR-04 - Supervised runner and consumer migration

Depends on: JPR-03.

Areas: runner, Imports, Incident Portability, Reference Data, Reporting,
Report Composition preview, server lifecycle, tests, tracker.

Remediation:

- Replace one-shot recovery with synchronous initial recovery followed by a
  continuous supervisor.
- Bound concurrency, deduplicate in-flight jobs, renew leases, and honor
  persisted retry eligibility.
- Replace handler-name/string dispatch with typed non-blocking job
  notification.
- Permit handler registration only before runner start; duplicate or late
  registration fails startup.
- Remove caller-selected notification routing and every module-local recovery
  entry point after the supervisor becomes authoritative.
- Remove module recovery functions, `MarkRunning`/resume helpers, ignored
  duplicate registration, and string parsing.
- Make shutdown stop claims, cancel attempts, conditionally release leases,
  and drain under the existing application deadline.

Rationale and long-term benefit: Durable jobs retry and recover without a
process restart, long work remains exclusively owned, and load is bounded for
future worker families.

Compatibility and migration: Internal worker interfaces break. Running becomes
entirely claim-owned; public state ordering remains unchanged.

Risk if unresolved: Failed jobs remain stuck until restart, long handlers can
be duplicated after 30 seconds, and unbounded goroutines can exhaust the
process.

Binary exit:

- a handler running beyond two lease periods is not reclaimed;
- a forced ownership loss cancels the attempt and fences late commits;
- failed work retries without restart at the exact delays and exhausts once;
- in-flight execution never exceeds eight and one job runs once at a time;
- dropped notifications are recovered by scanning;
- graceful restart recovers released or expired work; and
- all affected owner slices pass.

### JPR-05 - Logical expiry and compact tombstones

Depends on: JPR-04.

Areas: migration `00060_platform_jobs_expiry_tombstones.sql`, reads,
cancellation, Jobs maintenance, indexes, tests, tracker.

Remediation:

- Add `expired_at` and terminal/tombstone constraints.
- Apply exact-cutoff not-found behavior to reads and cancellation.
- Compact expired rows in deterministic bounded batches after the first full
  sweep interval.
- Clear only Jobs-owned operational/public-summary material listed in Section
  4 and retain required identity/provenance.
- Replace the current retention lookup index with an unexpired-candidate
  index.
- Emit no public transition and perform no row or durable-output deletion.

Rationale and long-term benefit: Expiry becomes deterministic and sensitive or
large worker material is not retained indefinitely, while cross-owner
provenance remains safe.

Compatibility and migration: Public callers receive the already defined
not-found result after expiry. Before the first compaction, rollback is
ordinary; after compaction, cleared summaries require backup restoration or a
forward fix.

Risk if unresolved: Retention deadlines have no operational effect, payloads
and diagnostics grow indefinitely, and a tactical hard delete could cascade
into durable outputs.

Binary exit:

- reads succeed immediately before and fail at the exact cutoff;
- cancellation uses the same masked result;
- batches are ordered, bounded, restart-safe, and concurrency-safe;
- sensitive fields are cleared and retained fields are exact;
- no Collaboration event or durable-output mutation occurs; and
- migration, recovery, Job API, and affected owner evidence passes.

### JPR-06 - Operational observability and anti-regression guards

Depends on: JPR-05.

Areas: Jobs telemetry, application fatal lifecycle, boundary policy, generated
harness artifacts, tests, tracker.

Remediation:

- Correct active-job aggregation to select and group by catalog-backed
  `job_kind`.
- Cover the complete handler attempt with `cartulary.jobs.run`.
- Add `cartulary.jobs.attempts`,
  `cartulary.jobs.lease_renewal.failures`, and `cartulary.jobs.expired` counters
  using only catalog job kind and closed result/error tokens.
- Forbid job ID, attempt token, progress-unit ID, payload, panic value, and raw
  error text from telemetry and diagnostics.
- Treat initial recovery failure as pre-readiness failure and unexpected
  supervisor termination as `job_dequeue` component loss.
- Keep transient scan errors inside the live supervisor with bounded retry.
- Add negative fixtures for removed construction, dispatch, transition,
  recovery, and unfenced-finalization paths.

Rationale and long-term benefit: Operators can distinguish backlog, retries,
lease loss, expiry, and component failure without high-cardinality or secret
data, and structural regressions fail automatically.

Compatibility and migration: Telemetry gains adopted bounded signals; existing
public protocols remain unchanged. Dashboards may add the new counters but do
not require scope-to-kind migration again.

Risk if unresolved: The active gauge can misaggregate or fail at runtime,
attempt health is invisible, and removed compatibility paths can return.

Binary exit:

- telemetry conformance and safe-sentinel tests pass;
- multi-kind active counts are exact;
- attempt spans cover handler execution and terminal classification;
- lifecycle failure tests prove pre-readiness and component-loss behavior; and
- boundary, generation, generated-policy, and drift checks pass.

### JPR-07 - Validation and handoff

Depends on: every prior JPR slice.

Areas: all affected verification layers, cutover, documentation, tracker.

Remediation:

- Perform a drained migration and corrected-writer startup rehearsal.
- Run all commands and owner slices in Section 7.
- Inspect retained full-check evidence and classify every failure.
- Update the ledger, acceptance criteria, rollback posture, residual risks,
  and operational handoff from actual evidence.

Rationale and long-term benefit: Repository success, migration safety, runtime
supervision, browser-observable ordering, and handoff are proven through the
same path a future deployment uses.

Compatibility and migration: Planned downtime remains required. Mixed-version
or zero-downtime rollout requires a separately adopted protocol and is out of
scope.

Risk if unresolved: A locally correct refactor could still fail during schema
cutover, startup recovery, shutdown, or externally observed state handling.

Binary exit:

- every JPR acceptance criterion is `PASS` from actual evidence;
- one fenced supervised writer is active;
- expired resources are inaccessible and compact safely;
- no removed compatibility surface remains; and
- the final tracker is complete and lint-clean.

## 6. Migration, Cutover, and Rollback

### Migration 00059

`00059_platform_jobs_execution_fencing.sql` is applied only after closing job
admission, stopping the old process under the deployment-global process lease,
and proving there is no active lease or running/cancel-requested row.

The migration:

- rejects unsafe rows before any schema mutation;
- preserves eligible queued jobs;
- removes nullable definition compatibility;
- installs attempt/failure/retry fields and constraints; and
- supports downgrade only before the new writer records execution state.

### Migration 00060

`00060_platform_jobs_expiry_tombstones.sql` adds tombstone state and its
constraints/indexes. The corrected writer starts with the compactor delayed for
one full five-minute interval so startup validation can complete before an
irreversible compaction effect.

Rollback posture:

- before the first compaction, stop the corrected writer and downgrade through
  the normal migration path;
- after compaction, do not attempt to infer cleared summaries or payloads;
  restore the pre-cutover backup or roll forward; and
- never run old and new writers concurrently.

### Unsupported retained data

Unknown definitions, missing handlers, nullable job identity, active old
leases, or replay-incompatible state block the cutover. The implementation
must report only bounded counts and safe job-kind/status tokens. It must not
report job IDs, attempt IDs, payloads, incident content, paths, progress-unit
IDs, or secrets. Reset/reseed is never automatic.

## 7. Validation Matrix

### Required commands

| Gate | Command | Required stage |
| --- | --- | --- |
| Tracker checkpoint | `make lint-markdown` | After every tracker update and before the next slice |
| Owner routing | `make explain-test-owner OWNER=platform.jobs` | JPR-01 and JPR-07 |
| Owner guide | `make task-guide ROLE=module-author OWNER=platform.jobs` | Before selecting Jobs rows |
| Jobs unit | `make test-slice OWNER=platform.jobs` | Every implementation slice affecting Jobs |
| Jobs service-backed | `make service-backed-test-slice OWNER=platform.jobs` | Every storage/runtime slice |
| Generation | `make generate` | Authored verification or boundary inputs change |
| Generation drift | `make generate-drift` | JPR-01, JPR-06, JPR-07 |
| Generated policy | `make generated-artifact-policy-check` | JPR-01, JPR-06, JPR-07 |
| JSON shapes | `make json-shape-check` | Contract/migration and final gates |
| OpenAPI compatibility | `make openapi-compatibility-check` | JPR-07 |
| Migration drift | `make migration-drift` | JPR-03, JPR-05, JPR-07 |
| Backend boundaries | `make backend-module-boundary-check` | JPR-02, JPR-06, JPR-07 |
| Security | `make go-gosec-targeted` | JPR-03, JPR-04, JPR-06 |
| Fast broad | `make test-fast` | After structural consumer migration and JPR-07 |
| Browser state | `make browser-e2e-stateful` | JPR-07, unconditionally |
| Finalization | `make agent-finalize` | Before final broad verification |
| Full check | `make check` | Final JPR-07 gate |
| Retained-root inspection | `make explain-run RESULTS_DIR=<run-root>` | Final successful full-check root |

`RESULTS_DIR` is supplied to `make agent-finalize` only for an eligible
successful full warm-check root. Otherwise the tracker records that retained
run maintenance was skipped because `RESULTS_DIR` was unset.

### Required owner evidence

JPR-07 runs the complete `platform.jobs` unit and service-backed inventories
plus affected slices for:

- `app.server`;
- `module.jobapi`;
- `module.collaboration`;
- `module.extensions`;
- Imports;
- Incident Portability;
- Reference Data;
- Reporting;
- Report Composition;
- Postgres/migration ownership; and
- Telemetry.

Each consumer is validated after its migration. A later consumer does not
begin while the prior consumer's targeted evidence is red.

### Required behavior scenarios

- constructor rejection and catalog immutability;
- exact enqueue identity and unknown-definition rejection;
- every legal and illegal lifecycle/progress branch;
- initial recovery and continuous recovery without restart;
- bounded concurrency and in-flight deduplication;
- handler execution longer than two lease periods;
- renewal success, storage failure, ownership loss, and stale finalization;
- persisted 5-second and 30-second retries plus third-failure exhaustion;
- panic, raw error, and incomplete nil-return secrecy;
- graceful drain, forced timeout, restart, and component-loss classification;
- exact retention cutoff and not-found masking;
- deterministic tombstone batches and restart;
- preservation of every cross-owner durable output and foreign-key row;
- no expiry `job_progress` event;
- HTTP polling, live WebSocket, replay, authorization, and frontend parity;
- actual-kind active metrics and safe attempt/renewal/expiry telemetry; and
- absence of every removed source surface through canonical boundary fixtures.

## 8. Slice Ledger

Exactly one row may be `IN_PROGRESS` during implementation.

| Slice | Status | Depends on | Planned checkpoint | Actual files/evidence | Failures/blockers | Rollback posture |
| --- | --- | --- | --- | --- | --- | --- |
| JPR-01 | DONE | Tracker revision | Adopt execution/expiry policy without leaving red or orphan evidence | Tracker; Core 01; Core 04; Extensions NLSpec; OpenTelemetry NLSpec. Lint `.cartulary/test-results/20260808T135022Z-p101306`; generation drift `.cartulary/test-results/20260808T135031Z-p102078`; generated policy `.cartulary/test-results/20260808T135040Z-p104817`; Jobs unit `.cartulary/test-results/20260808T135041Z-p105235` | None | Revert the five owner/tracker inputs together; no implementation or schema change exists at this checkpoint. |
| JPR-02 | DONE | JPR-01 | Immutable construction and dead API removal | `Catalog`, `Definition`, optional `ExtensionPolicy`, `RuntimePolicy`, `ManagerOptions`, and `RunnerOptions`; transaction admission uses `EnqueueParams` and one shared catalog; server and test composition migrated; removed mutable configuration, `Manager.Create`, `Manager.ExtensionContract`, variadic catalogs, anonymous dispatch, caller handler identity, and retired aliases. Red composition root `.cartulary/test-results/20260808T135235Z-p109274`; green composition root `.cartulary/test-results/20260808T140325Z-p114497`; Jobs root `.cartulary/test-results/20260808T141314Z-p308105`; Jobs service-backed root `.cartulary/test-results/20260808T141340Z-p310601`; boundary root `.cartulary/test-results/20260808T141306Z-p307705`; fast root `.cartulary/test-results/20260808T141046Z-p245388` | First `app.server` run had one unrelated two-second server-process timeout at `.cartulary/test-results/20260808T140747Z-p190230`; isolated rerun passed at `.cartulary/test-results/20260808T140839Z-p217311`. No remaining red row or blocker. | Revert the complete composition/API checkpoint; no compatibility facade or mixed catalog writer exists. |
| JPR-03 | DONE | JPR-02 | Migration 00059 and fenced typed execution substrate | `db/migrations/00059_platform_jobs_execution_fencing.sql`; migration manifest and generated SQL model; Jobs execution, lifecycle, catalog, runner-signature, and transaction-service files; Extensions finalization/reconciliation; server/extension assembly; Imports, Incident Portability, Reference Data, Reporting, Report Composition, Collaboration, Job API, and test-support migrations; Jobs/Postgres verification rows and tests. Red typed-contract root `.cartulary/test-results/20260808T141754Z-p316139`; red migration root `.cartulary/test-results/20260808T141801Z-p316506`; green Jobs `.cartulary/test-results/20260808T151900Z-p892812`; green Jobs service-backed `.cartulary/test-results/20260808T152121Z-p1011112`; strengthened stale/superseded row `.cartulary/test-results/20260808T152450Z-p1024562`; migration-59 row `.cartulary/test-results/20260808T151900Z-p892820`; Imports `.cartulary/test-results/20260808T151131Z-p728414`; Incident Portability `.cartulary/test-results/20260808T152008Z-p955314`; Reference Data `.cartulary/test-results/20260808T152008Z-p955307`; Reporting `.cartulary/test-results/20260808T152008Z-p955300`; Report Composition `.cartulary/test-results/20260808T151608Z-p820246`; Extensions `.cartulary/test-results/20260808T152008Z-p955297`; Collaboration `.cartulary/test-results/20260808T151608Z-p820271`; Job API `.cartulary/test-results/20260808T152255Z-p1018649`; app server `.cartulary/test-results/20260808T151900Z-p892822`; fast `.cartulary/test-results/20260808T151319Z-p755267`; generation drift `.cartulary/test-results/20260808T151948Z-p924225`; migration drift `.cartulary/test-results/20260808T151948Z-p924254`; boundary `.cartulary/test-results/20260808T151948Z-p924486`; targeted security `.cartulary/test-results/20260808T151948Z-p924622` | Intended red contract/migration tests failed before implementation. Imports initially exposed asynchronous fixture assumptions and a cancellation-finalizer fence distinction at `.cartulary/test-results/20260808T150704Z-p688301`; all corrected rows pass. Extensions initially preserved success-over-cancel behavior at `.cartulary/test-results/20260808T151608Z-p820252`; cancellation now wins without owner/proof mutation. Job API fixtures initially omitted required extension admission at `.cartulary/test-results/20260808T152121Z-p1011109`; fixtures now use the production admission path. No remaining red row or blocker. | Migration preflight rejects unsupported identity or active execution before mutation. Stop the corrected writer before downgrade; downgrade is supported only before corrected execution fields are used. Otherwise restore the drained backup or roll forward. No mixed writer or semantic backfill is supported. |
| JPR-04 | DONE | JPR-03 | Supervised runner and consumer migration | Jobs runner, global recovery storage, exact handler catalog validation, server pre-publication activation/component-loss wiring, Imports/Incident Portability/Reference Data/Reporting recovery removal, and test-support migration. Red supervision root `.cartulary/test-results/20260808T153113Z-p1036084`; green focused supervision `.cartulary/test-results/20260808T154741Z-p1297324`; Jobs `.cartulary/test-results/20260808T154835Z-p1326883`; Jobs service-backed `.cartulary/test-results/20260808T154835Z-p1326897`; app server `.cartulary/test-results/20260808T154835Z-p1326876`; Imports `.cartulary/test-results/20260808T154431Z-p1206687`; Incident Portability `.cartulary/test-results/20260808T154431Z-p1206702`; Reference Data `.cartulary/test-results/20260808T154146Z-p1114434`; Reporting `.cartulary/test-results/20260808T154146Z-p1114436`; Report Composition `.cartulary/test-results/20260808T154431Z-p1206717`; Extensions `.cartulary/test-results/20260808T154523Z-p1237433`; Collaboration `.cartulary/test-results/20260808T154523Z-p1237428`; Job API `.cartulary/test-results/20260808T154523Z-p1237439`; fast `.cartulary/test-results/20260808T154146Z-p1114646`; generation drift `.cartulary/test-results/20260808T154811Z-p1299137`; boundary `.cartulary/test-results/20260808T154812Z-p1299319`; targeted security `.cartulary/test-results/20260808T154812Z-p1299450` | Intended red supervision row failed before implementation. Generation initially rejected a non-canonical verification-ID order at `.cartulary/test-results/20260808T153028Z-p1030944`; the authored owner list was sorted and regeneration passed. No completed row remains red and no blocker remains. | Revert the runner, storage query, server lifecycle, and all migrated recovery callers together. There is no old dispatcher/recovery facade or mixed execution path to preserve. |
| JPR-05 | DONE | JPR-04 | Migration 00060 and compact tombstones | `db/migrations/00060_platform_jobs_expiry_tombstones.sql`; migration manifest and generated SQL model; exact-cutoff Jobs reads/cancel replay; one-batch private compactor; runner expiry supervision; Job API masking; Jobs/Postgres verification rows and tests. Red behavior root `.cartulary/test-results/20260808T155710Z-p1368987`; red migration root `.cartulary/test-results/20260808T155710Z-p1368988`; green expiry/compaction `.cartulary/test-results/20260808T160414Z-p1402520`; green migration `.cartulary/test-results/20260808T155942Z-p1379047`; Jobs `.cartulary/test-results/20260808T160456Z-p1407245`; Jobs service-backed `.cartulary/test-results/20260808T160522Z-p1410100`; Job API `.cartulary/test-results/20260808T160549Z-p1412025`; Collaboration `.cartulary/test-results/20260808T160604Z-p1413922`; Extensions `.cartulary/test-results/20260808T160738Z-p1447481`; Postgres `.cartulary/test-results/20260808T160817Z-p1472876`; app server `.cartulary/test-results/20260808T160902Z-p1481029`; fast `.cartulary/test-results/20260808T160943Z-p1507769`; generation drift `.cartulary/test-results/20260808T160839Z-p1475410`; migration drift `.cartulary/test-results/20260808T160852Z-p1478190`; generated policy `.cartulary/test-results/20260808T161217Z-p1574103`; JSON shape `.cartulary/test-results/20260808T161223Z-p1574578` | Intended red behavior and migration rows failed before implementation. One concurrent Make invocation raced on its shared service-image stamp at `.cartulary/test-results/20260808T155942Z-p1379046`; sequential rerun passed. Test fixture SQL exposed two setup-only failures at `.cartulary/test-results/20260808T160003Z-p1384239` and `.cartulary/test-results/20260808T160101Z-p1389742`; corrected tests pass. No completed row remains red and no blocker remains. | Downgrade to 59 is guarded and supported before the first tombstone. Once `expired_at` is written, restore the pre-compaction backup or roll forward; never infer cleared summaries/evidence or start an old writer. Physical job/output deletion remains prohibited. |
| JPR-06 | DONE | JPR-05 | Telemetry, fatal lifecycle, and boundary guards | Jobs attempt spans and actual-kind active gauge; attempt, renewal-failure, and expiry counters; telemetry registry/privacy and OTel corpus; telemetry-owned SDK capture support; removed-surface architecture guard; Reporting typed notification cleanup; explicit pre-serving/published `job_dequeue` loss tests. Intended telemetry reds `.cartulary/test-results/20260808T161903Z-p1585880` and `.cartulary/test-results/20260808T161922Z-p1587722`; focused Jobs/guard `.cartulary/test-results/20260808T163301Z-p1630239`; platform telemetry `.cartulary/test-results/20260808T163047Z-p1616735`; OTel conformance `.cartulary/test-results/20260808T163314Z-p1632204`; Reporting `.cartulary/test-results/20260808T163329Z-p1637211` and `.cartulary/test-results/20260808T163342Z-p1643301`; app server `.cartulary/test-results/20260808T163425Z-p1648168`; Jobs `.cartulary/test-results/20260808T163513Z-p1676663`; Jobs service-backed `.cartulary/test-results/20260808T163538Z-p1681158`; generation drift `.cartulary/test-results/20260808T163607Z-p1683062`; generated policy `.cartulary/test-results/20260808T163614Z-p1685763`; JSON shape `.cartulary/test-results/20260808T163638Z-p1686807`; boundary `.cartulary/test-results/20260808T163641Z-p1687267`; targeted security `.cartulary/test-results/20260808T163642Z-p1687687`; fast `.cartulary/test-results/20260808T163653Z-p1713054` | Intended telemetry tests failed before implementation. A strengthened short-attempt fixture initially applied the long-attempt duration assertion to every span at `.cartulary/test-results/20260808T162952Z-p1608900`; the assertion now distinguishes blocked and immediate handlers. OTel conformance correctly rejected SDK imports outside its owner at `.cartulary/test-results/20260808T163107Z-p1622334`; capture mechanics moved under `internal/platform/telemetry/testsupport`. JSON shape then required that support root in the authored inventory at `.cartulary/test-results/20260808T163615Z-p1686164`; the inventory was updated. All resolved checks pass and no red row or blocker remains. | Signals, registry, privacy policy, corpus, and tests must be reverted together. The runner/storage behavior and schema need no rollback; no public protocol or deployment configuration changed. |
| JPR-07 | DONE | JPR-06 | Drained cutover, broad validation, and handoff | Drained migration/startup and pre-compaction rollback rehearsals; complete owner matrix; final static, browser, fast, and full-check gates; retained evidence record below | Cutover fixtures, two broad-load test budgets, and two obsolete helpers were corrected; all final routed and broad evidence is green | Planned downtime and pre-cutover backup remain mandatory. Downgrade is supported only before the first tombstone; afterward restore backup or roll forward. Never reopen an old writer on new state. |

### JPR-07 final evidence record

The cutover rehearsal passed at
`.cartulary/test-results/20260808T164714Z-p1795006`. It closed admission,
proved process-lease writer exclusion, took a non-empty custom-format Postgres
backup, migrated through 00059 and 00060, started one corrected writer,
completed synchronous recovery, and proved the first compaction did not run
early. The same row proved guarded downgrade from 00060 through 00058 before
any corrected write or tombstone.

Complete owner evidence passed:

- Jobs unit `.cartulary/test-results/20260808T164744Z-p1797217` and
  service-backed `.cartulary/test-results/20260808T164808Z-p1800276`;
- app server `.cartulary/test-results/20260808T164834Z-p1802188` and
  service-backed `.cartulary/test-results/20260808T164911Z-p1828146`;
- Job API `.cartulary/test-results/20260808T164939Z-p1850579` and
  service-backed `.cartulary/test-results/20260808T164949Z-p1852763`;
- Collaboration `.cartulary/test-results/20260808T164952Z-p1854221` and
  service-backed `.cartulary/test-results/20260808T165119Z-p1884964`;
- Extensions `.cartulary/test-results/20260808T165247Z-p1911784` and
  service-backed `.cartulary/test-results/20260808T165319Z-p1936449`;
- Imports `.cartulary/test-results/20260808T165815Z-p2007246` and
  service-backed `.cartulary/test-results/20260808T165855Z-p2033245`;
- Incident Portability `.cartulary/test-results/20260808T165932Z-p2056729`
  and service-backed `.cartulary/test-results/20260808T170000Z-p2060594`;
- Reference Data `.cartulary/test-results/20260808T170021Z-p2062434` and
  service-backed `.cartulary/test-results/20260808T170053Z-p2086509`;
- Reporting `.cartulary/test-results/20260808T170121Z-p2108818` and
  service-backed `.cartulary/test-results/20260808T170132Z-p2111373`;
- Report Composition `.cartulary/test-results/20260808T170135Z-p2112817` and
  service-backed `.cartulary/test-results/20260808T170145Z-p2114435`;
- Postgres `.cartulary/test-results/20260808T170152Z-p2115936` and
  service-backed `.cartulary/test-results/20260808T170207Z-p2118246`; and
- Telemetry `.cartulary/test-results/20260808T170221Z-p2120148`. Telemetry
  has no service-backed row; the harness rejected that inapplicable command
  without running product evidence.

Final gates passed: generation drift
`.cartulary/test-results/20260808T170238Z-p2121588`, generated-artifact policy
`.cartulary/test-results/20260808T170246Z-p2124282`, JSON shape
`.cartulary/test-results/20260808T170247Z-p2124683`, migration drift
`.cartulary/test-results/20260808T170250Z-p2125107`, backend boundaries
`.cartulary/test-results/20260808T170256Z-p2127943`, targeted security
`.cartulary/test-results/20260808T170258Z-p2128363`, OpenAPI compatibility
`.cartulary/test-results/20260808T170307Z-p2152076`, OTel conformance
`.cartulary/test-results/20260808T170308Z-p2152590`, agent finalization
`.cartulary/test-results/20260808T170320Z-p2156460`, fast tests
`.cartulary/test-results/20260808T170338Z-p2159147`, and stateful browser
`.cartulary/test-results/20260808T170539Z-p2208508` (36/36). Agent
finalization ran before broad verification with `RESULTS_DIR` unset because no
eligible prior successful warm-check root existed, so retained-run maintenance
was skipped as required.

The final full check passed 736/736 at
`.cartulary/test-results/20260808T172758Z-p2775481`. `make explain-run`
inspected that root successfully; its run summary reports `status=pass`, zero
failed units, and no failure classification.

Resolved evidence is retained rather than hidden:

- cutover fixture/schema assumptions failed at
  `.cartulary/test-results/20260808T164301Z-p1782636` and a deliberately
  interrupted hung diagnostic at
  `.cartulary/test-results/20260808T164343Z-p1787708`; the backup environment,
  lease cleanup, and authored fixture were corrected before the green row;
- exact-cutoff cancellation changed the lock observed by the Imports fixture;
  failed diagnostic roots
  `.cartulary/test-results/20260808T165346Z-p1959378`,
  `.cartulary/test-results/20260808T165505Z-p1989243`,
  `.cartulary/test-results/20260808T165629Z-p1994933`, and
  `.cartulary/test-results/20260808T165718Z-p2000483` led to a deterministic
  transaction-lock observer and green owner evidence;
- the first broad check at
  `.cartulary/test-results/20260808T170802Z-p2234762` found two related dead
  helpers through static analysis and one unrelated two-second process-reset
  timeout; the helpers were removed and the process row passed at
  `.cartulary/test-results/20260808T171335Z-p2391151`;
- the next broad check at
  `.cartulary/test-results/20260808T171412Z-p2411520` exposed a ten-second
  Imports eventual-state budget under full resource contention; the same row
  passed alone, its bounded observation budget was raised to 30 seconds, and
  it passed at `.cartulary/test-results/20260808T172133Z-p2591468`; and
- the following broad check at
  `.cartulary/test-results/20260808T172149Z-p2593264` reproduced the unrelated
  server-process reset timeout. The test-only request budget now matches the
  existing ten-second process-cleanup budget and the routed row passed at
  `.cartulary/test-results/20260808T172726Z-p2755838` before the clean full
  check. No production timeout or behavior changed.

## 9. Binary Acceptance Criteria

| Acceptance ID | Binary completion condition | Current status | Actual evidence |
| --- | --- | --- | --- |
| JPR-AC-101 | Manager, transaction service, catalog, and runner construction is immutable and fails before readiness for missing dependencies. | PASS | Jobs composition and full owner evidence `.cartulary/test-results/20260808T164744Z-p1797217`; final check `.cartulary/test-results/20260808T172758Z-p2775481`. |
| JPR-AC-102 | Dead exports, mutable configuration, duplicate job identity, retired aliases, anonymous/string dispatch, one-shot recovery, and duplicate-handler tolerance are absent. | PASS | Jobs architecture guard `.cartulary/test-results/20260808T163301Z-p1630239`; backend boundary `.cartulary/test-results/20260808T170256Z-p2127943`; final check. |
| JPR-AC-103 | Every retained job has non-null catalog identity and handler identity; migration 00059 rejects unsafe data before mutation. | PASS | Cutover/rollback `.cartulary/test-results/20260808T164714Z-p1795006`; Postgres `.cartulary/test-results/20260808T170152Z-p2115936`; migration drift `.cartulary/test-results/20260808T170250Z-p2125107`. |
| JPR-AC-104 | Every handler-owned payload read, progress update, owner transaction, and terminal completion requires one live opaque execution attempt. | PASS | Jobs service-backed `.cartulary/test-results/20260808T164808Z-p1800276`; all six consumer owner/service-backed pairs in the final evidence record; final check. |
| JPR-AC-105 | Long handlers renew safely; ownership loss fences stale work; only one attempt can publish owner or job state. | PASS | Jobs unit and service-backed roots `.cartulary/test-results/20260808T164744Z-p1797217` and `.cartulary/test-results/20260808T164808Z-p1800276`; final check. |
| JPR-AC-106 | Recovery is continuous, retries use the exact persisted delays, concurrency is bounded at eight, and failed work does not require restart. | PASS | Jobs unit and service-backed roots `.cartulary/test-results/20260808T164744Z-p1797217` and `.cartulary/test-results/20260808T164808Z-p1800276`; cutover recovery `.cartulary/test-results/20260808T164714Z-p1795006`. |
| JPR-AC-107 | Startup, readiness, notification, shutdown, restart, transient scan failure, and component-loss behavior is deterministic and evidence-backed. | PASS | app server `.cartulary/test-results/20260808T164834Z-p1802188`; Jobs `.cartulary/test-results/20260808T164744Z-p1797217`; browser `.cartulary/test-results/20260808T170539Z-p2208508`; final check. |
| JPR-AC-108 | Expired resources are hidden at the exact cutoff and compact to private tombstones without deleting or mutating durable outputs or emitting progress. | PASS | Jobs service-backed `.cartulary/test-results/20260808T164808Z-p1800276`; Job API `.cartulary/test-results/20260808T164939Z-p1850579`; Collaboration `.cartulary/test-results/20260808T164952Z-p1854221`; Postgres `.cartulary/test-results/20260808T170152Z-p2115936`. |
| JPR-AC-109 | Actual-kind telemetry is correct and bounded; removed architecture paths fail canonical boundary checks; public HTTP/WS/auth/generated shapes remain compatible. | PASS | Telemetry `.cartulary/test-results/20260808T170221Z-p2120148`; OTel `.cartulary/test-results/20260808T170308Z-p2152590`; boundary `.cartulary/test-results/20260808T170256Z-p2127943`; OpenAPI `.cartulary/test-results/20260808T170307Z-p2152076`; browser and final check. |
| JPR-AC-110 | All required owner, migration, security, drift, fast, browser, finalization, and full-check gates pass with an inspected retained root and complete handoff. | PASS | All retained roots in the JPR-07 final evidence record; inspected full check `.cartulary/test-results/20260808T172758Z-p2775481`, 736/736 with zero failures. |

No criterion may be marked `PASS` from planned behavior, a local unretained
command, or another owner's unrelated evidence.

## 10. Risks, Boundaries, and Handoff

### Primary risks

| Risk | Mitigation | Residual posture |
| --- | --- | --- |
| Internal API churn breaks many workers simultaneously | Use compile-time typed migration and validate one owner at a time. | All affected owners pass; no compatibility facade is retained. |
| Attempt fencing is omitted from one owner finalizer | Require opaque execution at the transaction boundary and add canonical source guards. | Consumer owner evidence and source guards pass; a future unfenced finalizer will fail the boundary suite. |
| Recovery polling causes database load | Fixed five-second cadence, indexed eligibility, 100-row batches, and eight-worker concurrency. | The fixed policy is validated; reconfiguration requires a later adopted policy change. |
| Retry loops amplify persistent failures | Persist exact delays and cap failures at three. | Exhaustion fails closed with the existing safe summary; changing the budget requires an adopted policy revision. |
| Expiry deletes durable owner output | Prohibit hard deletion and retain compact provenance tombstones. | Field-level and cross-owner preservation evidence passes; physical deletion remains separately gated. |
| Compaction removes rollback data | Delay the first sweep and require backup restore or roll-forward after compaction. | Irreversibility after the first tombstone remains an operator responsibility; no inferred reconstruction is permitted. |
| Telemetry leaks identifiers or errors | Closed attributes and secrecy sentinel tests across spans, metrics, logs, and public summaries. | Registry, hostile corpus, and conformance pass; any future leak remains release-blocking. |
| Mixed writers corrupt attempt state | Drained migration under the process lease; no dual schema or writer. | Writer exclusion is rehearsed; zero-downtime rollout remains unsupported and out of scope. |

### Explicit boundaries

- No public route, envelope, status token, error code, authorization result, or
  progress member changes.
- No v1 reader, handler alias, dual writer, inferred semantic repair, or
  automatic reset/reseed.
- No deployment-config schema change.
- No physical deletion of job rows or owner outputs.
- No frontend implementation change unless required browser evidence exposes a
  real consumer defect.
- No generated file is hand-edited.
- Core 05 remains inapplicable.

### Completed operator handoff

Deployment remains a planned-downtime operation:

1. Close job admission and stop the old writer while holding the
   application-process lease.
2. Prove there is no active old execution state, then take and verify the
   pre-cutover backup.
3. Apply migrations 00059 and 00060 without semantic backfill.
4. Start exactly one corrected writer, reacquire the process lease, and wait
   for synchronous initial recovery before readiness.
5. Keep the compactor idle until its first full five-minute interval.

Rollback before the first tombstone requires stopping the corrected writer
and running the guarded down migrations. After any tombstone is written,
restore the verified backup or roll forward; never infer cleared data, run an
old reader/writer against the new state, or introduce mixed writers. Unknown
or incomplete retained identity remains a cutover blocker and reset/reseed
still requires explicit approval.

The remaining risks are deliberate operational boundaries, not incomplete
implementation: planned downtime, backup custody, one-writer enforcement, and
post-compaction irreversibility. Public HTTP, WebSocket, authorization,
OpenAPI, status, error, and progress contracts remain compatible. There is no
active implementation slice and no deferred compatibility facade.

### Planning revision record

| Time | State | Files changed | Validation | Result | Next action |
| --- | --- | --- | --- | --- | --- |
| 2026-08-08 America/New_York | Second iteration planned; no active slice | Only this tracker | `git diff --check`; `make lint-markdown` | Pass; lint root `.cartulary/test-results/20260808T132243Z-p87099` | Begin JPR-01 only under separate implementation authorization. |
| 2026-08-08 America/New_York | JPR-01 started under implementation authorization | This tracker; owner edits pending | Baseline `main` at `a43f9ddf6fe83023007ba6deed6cc6a5b5676a3d`; tracker checkpoint lint pending | In progress | Adopt owner requirements, validate existing evidence, and close JPR-01 before JPR-02. |
| 2026-08-08 America/New_York | JPR-01 complete; no active slice | Tracker and four adopted owner specifications | `git diff --check`; `make lint-markdown`; `make generate-drift`; `make generated-artifact-policy-check`; `make test-slice OWNER=platform.jobs` | Pass; retained roots recorded in the JPR-01 ledger | Mark JPR-02 `IN_PROGRESS`, lint the tracker, then begin immutable composition. |
| 2026-08-08 America/New_York | JPR-02 started | Tracker only at slice transition | `make lint-markdown` pending | In progress | Add red composition evidence, implement immutable construction, and finish all routed rows green. |
| 2026-08-08 America/New_York | JPR-02 complete; no active slice | Jobs catalog/options composition, server and consumer assembly, test support, verification routing, and this tracker | `make format`; `make test-slice OWNER=platform.jobs`; `make service-backed-test-slice OWNER=platform.jobs`; `make backend-module-boundary-check`; `make test-fast`; source search for removed surfaces; `git diff --check` | Pass; retained roots recorded in the JPR-02 ledger. One unrelated process timeout passed on isolated rerun. | Lint this checkpoint, then mark JPR-03 `IN_PROGRESS` before migration 00059 work. |
| 2026-08-08 America/New_York | JPR-03 started | Tracker only at slice transition | JPR-02 checkpoint lint `.cartulary/test-results/20260808T141423Z-p312523`; pre-slice lint `.cartulary/test-results/20260808T141452Z-p313526` | Pass | Add red fencing/migration evidence, then implement migration 00059 and atomically migrate consumers. |
| 2026-08-08 America/New_York | JPR-03 complete; no active slice | Migration 00059, generated SQL model, Jobs fenced execution and typed terminal APIs, owner finalizers, affected consumers, test support, verification routing, and this tracker | `make format`; `make generate`; `make generate-drift`; `make migration-drift`; `make test-slice OWNER=platform.jobs`; `make service-backed-test-slice OWNER=platform.jobs`; affected owner slices; `make backend-module-boundary-check`; `make go-gosec-targeted`; `make test-fast`; source guards; `git diff --check`; checkpoint `make lint-markdown` | Pass; retained roots and resolved red evidence are recorded in the JPR-03 ledger; checkpoint lint `.cartulary/test-results/20260808T152549Z-p1026949`. | Mark JPR-04 `IN_PROGRESS`, lint the tracker, then begin runner supervision and recovery removal. |
| 2026-08-08 America/New_York | JPR-04 started | Tracker only at slice transition | Final JPR-03 lint `.cartulary/test-results/20260808T152613Z-p1028053`; pre-slice lint pending | In progress | Add red supervision evidence, then implement the bounded continuous runner and remove every old dispatch/recovery path. |
| 2026-08-08 America/New_York | JPR-04 complete; no active slice | Supervised Jobs runner and recovery storage, server lifecycle, affected module recovery callers, verification routing, test support, and this tracker | `make format`; `make generate`; `make generate-drift`; `make test-slice OWNER=platform.jobs`; `make service-backed-test-slice OWNER=platform.jobs`; affected owner slices; `make backend-module-boundary-check`; `make go-gosec-targeted`; `make test-fast`; removed-surface source guard; `git diff --check`; checkpoint `make lint-markdown` | Pass; retained roots and resolved red generation evidence are recorded in the JPR-04 ledger; checkpoint lint `.cartulary/test-results/20260808T155115Z-p1358270`. | Mark JPR-05 `IN_PROGRESS`, lint the tracker, then begin migration 00060 and compaction work. |
| 2026-08-08 America/New_York | JPR-05 started | Tracker only at slice transition | Final JPR-04 lint `.cartulary/test-results/20260808T155134Z-p1359245`; pre-slice lint pending | In progress | Add red migration/expiry evidence, then implement exact logical expiry and bounded compact tombstones. |
| 2026-08-08 America/New_York | JPR-05 complete; no active slice | Migration 00060, generated SQL model, Jobs logical visibility and compaction, runner maintenance, Job API masking, verification routing, and this tracker | `make format`; `make generate`; `make generate-drift`; `make migration-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make test-slice OWNER=platform.jobs`; `make service-backed-test-slice OWNER=platform.jobs`; Job API, Collaboration, Extensions, Postgres, and app-server slices; `make test-fast`; `git diff --check`; checkpoint `make lint-markdown` | Pass; retained roots and resolved red/environmental evidence are recorded in the JPR-05 ledger; checkpoint lint `.cartulary/test-results/20260808T161257Z-p1575244`. | Mark JPR-06 `IN_PROGRESS`, lint the tracker, then begin telemetry and architecture-guard work. |
| 2026-08-08 America/New_York | JPR-06 started | Tracker only at slice transition | Final JPR-05 lint `.cartulary/test-results/20260808T161316Z-p1576213`; pre-slice lint pending | In progress | Add red telemetry/guard evidence, then implement exact attempt signals and removed-surface enforcement. |
| 2026-08-08 America/New_York | JPR-06 complete; no active slice | Jobs runner/telemetry/expiry; telemetry registry, privacy, corpus, and test support; Reporting notification path; server component-loss evidence; verification routing, test-support inventory, generated topology, and this tracker | `make format`; `make generate`; `make test-slice OWNER=platform.jobs`; `make service-backed-test-slice OWNER=platform.jobs`; `make test-slice OWNER=platform.telemetry`; `make otel-conformance`; Reporting and app-server owner slices; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make backend-module-boundary-check`; `make go-gosec-targeted`; `make test-fast`; `git diff --check`; checkpoint `make lint-markdown` | Pass; retained roots and all resolved red/boundary evidence are recorded in the JPR-06 ledger; checkpoint lint `.cartulary/test-results/20260808T163951Z-p1773652`. | Mark JPR-07 `IN_PROGRESS`, lint the tracker, then begin the cutover rehearsal and final gates. |
| 2026-08-08 America/New_York | JPR-07 started | Tracker only at slice transition | Final JPR-06 lint `.cartulary/test-results/20260808T164014Z-p1774607`; pre-slice lint `.cartulary/test-results/20260808T164039Z-p1775568` | Pass | Rehearse the drained two-migration cutover and guarded rollback posture, then run and inspect every final gate. |
| 2026-08-08 America/New_York | JPR-07 complete; no active slice | Cutover and rollback evidence in `internal/platform/jobs/cutover_test.go`; Jobs family routing; Imports cancellation fixture; server-process test budget; final dead-helper cleanup; this tracker | Complete owner matrix; generation, generated-policy, JSON, migration, boundary, security, OpenAPI, OTel, fast, browser, finalization, full check, retained-root inspection, `git diff --check`, and final `make lint-markdown`; roots recorded in the JPR-07 evidence record | Pass; full check `.cartulary/test-results/20260808T172758Z-p2775481` is 736/736 with zero failed units; tracker lint `.cartulary/test-results/20260808T173547Z-p2939074` | Handoff complete under the recorded planned-downtime, backup, one-writer, and post-compaction rollback posture. |

All JPR slices are complete. Every binary acceptance criterion is `PASS` from
retained evidence and there is no active slice.
