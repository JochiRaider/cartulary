# Test Harness Observability and Wall-Clock Optimization Tracker

## 1. Tracker control

| Field | Value |
| --- | --- |
| State | ACTIVE |
| Primary seam | Public Make invocation -> harness execution graph -> retained timing graph -> derived OpenTelemetry diagnostics |
| Initial source | `00522cfed1b6e5ca0936fb703de96c4c019544f3` on `revision/grid-adapter` |
| Current source | Serial reference substrate plus qualification-fixture corrections `7d7a87e19a8052981b1a309a496fd1a46da3f1a3`; safety candidate lineage begins at `528ef57d03c12dd47c0991a03380b0167ae54459` |
| Last updated | 2026-07-21 |
| Active item | T-001 |
| Successor to | `docs/handoffs/test-harness-subsystem-migration-refactor-tracker.md` |
| Product behavior | Preserved |
| Harness behavior | Additive diagnostics plus explicitly adopted scheduling and duration changes |

Status values are `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`, and
`DROPPED`. Exactly one task is `IN_PROGRESS` until all work is terminal. A task
becomes `DONE` only after its evidence and exit condition are both satisfied.
The 2026-07-20 remediation audit reopened work whose implementation presence was
mistaken for demonstrated conformance. Historical status transitions and roots
remain in this tracker for provenance, but every root collected before the
remediation audit is diagnostic-only and cannot qualify T-001 or T-012.

## 2. Authority and locked decisions

Authority order is Core 00 through Core 04 for product behavior, Core 05 for
claim-bearing benchmark publication, `docs/testing-harness-nlspec.md` for
harness mechanics, `docs/opentelemetry-instrumentation-nlspec.md` for the
application telemetry boundary, and `docs/domain.md` for product vocabulary.

The domain vocabulary review is complete and requires no edit. Targets, spans,
runners, work units, resource claims, and duration baselines are
implementation-support terms, not product-domain concepts.

Locked decisions:

- Native harness JSON remains authoritative; OTLP payloads are derived
  diagnostics.
- Local diagnostic generation is the default. Network export is a separate,
  explicit post-run command and inherited `OTEL_*` variables have no effect.
- The application telemetry scopes, `cartulary.module` registry, runtime
  bootstrap, and browser prohibition are unchanged.
- Public command IDs, selected tests, owner and evidence routing, failure
  precedence, artifact roles, and cleanup behavior are preserved.
- Generated roots and lockfiles are never hand-edited. This effort adds no OTel
  SDK dependency.
- Code-level profiles, cold tool provisioning, hosted-CI dashboards, and Core
  05 publication are outside this effort.

## 3. Current-state baseline

The retained run
`.cartulary/test-results/20260720T141740Z-p2992081` is diagnostic seed evidence,
not a qualifying baseline. It recorded:

| Observation | Duration |
| --- | ---: |
| `release-check` wall | 733.069 s |
| Standalone release browser targets | 297.490 s |
| `check` wall | 124.565 s |
| `backend-unit` scheduler occupancy | 87.244 s |
| Recorded backend Go capture/execution spans | 12.595 s |
| Post-scheduler/finalizer tail | 16.655 s |
| Cold license-tool provisioning | 43.280 s |

T-001 must replace this seed with three consecutive warm observations for every
in-scope public testing command. Aggregate runs may supply leaf observations;
commands absent from aggregates must be invoked directly. Failed, interrupted,
stale, capacity-mismatched, source-mismatched, or retry-contaminated runs remain
retained but cannot enter the accepted sample set.

The current authored inventory classifies 50 public testing entry points as
observability-required, four post-run diagnostic/export/maintenance commands
as owner-cited exclusions, and 43 public commands as explicitly out of scope.
The provisional machine profile is frozen by these digests:

| Profile component | SHA-256 |
| --- | --- |
| Host | `c31c805927c8912d7b56ec09c32cc91b8597404122198b80c923970c546f542d` |
| Capacity | `bd042c914469a69fa6ffda55f0f96e65e968d3a34a64aecd58790dada66d2df1` |
| Workload/evidence inventory | `ff7d5a4d9cab0b3ff2fae01ca0537fd19ad884b6c4bca98fce4b533ae9a88f1d` |
| Toolchain | `877e09fe43b77b76264d81f02ac74aaaaeafe700bba05511449f890ffb18c9ac` |

These digests and the single seed observation are not a completed baseline.
The evidence-roots manifest still requires three exact warm baseline roots and
three exact candidate roots per covered target.

## 4. Target architecture and public contracts

Every public target receives exactly one authored observability disposition of
`required`, `excluded`, or `out_of_scope`. An exclusion or out-of-scope entry
must cite an owner reason. Every required target also owns one stable
measurement profile covering canonical inputs, direct or aggregate eligibility,
warm-up policy, and performance gate. Non-test development, cleanup, and
interactive commands are normally out of scope but may still appear as child
spans inside a traced aggregate.

One top-level invocation maps to one trace. Stable span classes are invocation,
sequence step, scheduler wait, scheduler work unit, service lifecycle, runner,
artifact/finalization, and report collation. Span names are low-cardinality;
target, command, family, runner, work-unit, status, timing-bucket, wait-reason,
and logical-resource identities use closed attributes. Error messages, raw
commands, output, paths, environment values, headers, credentials, hostnames,
SQL, and product-authored data are forbidden.

The local artifact tree is:

```text
<run-root>/_shared/harness-observability/
  execution-context.json
  observability-index.json
  <invocation-id>/
    trace-bundle.json
    trace.otlp.json
    metrics.otlp.json
    hotspot-summary.json
```

Trace and span IDs are deterministic nonzero lowercase hexadecimal identifiers
derived from run identity plus stable source occurrence. Source artifact
digests make reconstruction auditable. Existing scheduler
`critical_path_wall_duration_ms` remains the full scheduler envelope. The
hotspot summary separately records the actual dependency critical path, queue
wait, resource-blocked time, union-based direct-child coverage, and unattributed
parent-envelope time.

Public additions:

- `make harness-observability-check RESULTS_DIR=<root|run-dir> [RUN_ID=<id>]`
- `make harness-otel-export RESULTS_DIR=<root|run-dir> [RUN_ID=<id>] HARNESS_OTLP_ENDPOINT=<url> [HARNESS_OTLP_HEADERS_FILE=<0600-json-file>]`
- `make explain-run ... DETAIL=performance`
- `make harness-performance-check EVIDENCE_ROOTS_FILE=<manifest>`
- `make harness-public-target-duration-baselines EVIDENCE_ROOTS_FILE=<baseline-window>`

Collector export accepts HTTPS or loopback HTTP, rejects credentials, query,
and fragment components, disables redirects, appends `/v1/traces` and
`/v1/metrics`, never retains header values, and never mutates the selected run.
Ordinary test commands perform no network export. Local diagnostic failure
retains a partial marker and warning without changing the underlying test
result; the explicit observability check fails closed.

## 5. Implementation tracker

| ID | Work item | Workstream | Status | Depends on | Owner | Evidence/artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- | --- |
| T-001 | Collect the clean serial reference window and generate the qualified public-target duration baseline | WS-00 baseline | IN_PROGRESS | T-007 | harness performance | exact retained execution contexts, baseline roots manifest, and generated baseline artifact | one discarded warm-up and three accepted observations exist for every required measurement profile |
| T-002 | Correct observability, artifact, security, scheduler, and performance requirements and tracker ownership | WS-01 specification | DONE | none | harness specification | `docs/testing-harness-nlspec.md`; requirements and acceptance crosswalk | every behavior has one normative owner and every measurement identity is reproducible |
| T-003 | Add the application-versus-harness OTel boundary | WS-01 specification | DONE | T-002 | telemetry specification | `docs/opentelemetry-instrumentation-nlspec.md`; OTel conformance fixtures | application scopes and runtime signals remain unchanged |
| T-004 | Add explicit target dispositions, stable measurement profiles, retained execution context, corrected schemas, attachments, and generated projections | WS-01 contracts | DONE | T-002, T-003 | harness contracts | closed public inventory, authored schemas, and reproducible generated projections | omissions, overlap, unknowns, duplicates, and unowned exclusions fail generation |
| T-005 | Implement retained-provenance deterministic reconstruction and interval-union hotspot analysis | WS-02 observability | DONE | T-004 | diagnostics | immutable context plus deterministic native, trace, metric, hotspot, and digest fixtures | retained roots reconstruct independently of the checkout and explicit graph parentage, paths, waits, gaps, and digests validate |
| T-006 | Unify sequence and scheduler lifecycle evidence, topology ownership, cancellation, and deterministic failure behavior | WS-02 observability | DONE | T-005 | execution runtime | scheduler v7, shared sequence scheduler, and lifecycle fixtures | required transitions and dependencies are attributable exactly once under success, failure, and interruption |
| T-007 | Make local validation read-only and exact-selected; correct OTLP export, privacy, and failure semantics | WS-02 observability | DONE | T-005, T-006 | diagnostics/export | tamper, exact-selection, OTLP decode, failure-class, redirect, timeout, and egress fixtures | selected source evidence is never mutated and export conforms exactly |
| T-008 | Consolidate compatible backend-unit exact symbols and run compatible groups concurrently | WS-03 optimization | TODO | T-001 | backend runner | retained 255-test parity run and expected process reduction | every symbol and row is proven exactly once across complete compatibility keys and failure paths |
| T-009 | Batch per-family output ingestion and parallelize deterministic report emission | WS-03 optimization | TODO | T-008 | output/finalizers | batch worker implemented; retained parity is green; qualified improvement gate remains open | output identity is stable and the hotspot clears its gate |
| T-010 | Execute `lint`, `ci`, and `release-check` through the topology-owned shared scheduler | WS-03 optimization | TODO | T-001, T-007 | scheduler/task surface | serial and DAG parity evidence for all three aggregates | dependency, resource, cancellation, output, cleanup, and primary-failure behavior are stable |
| T-011 | Make release browser readiness own its five-session schedule and capacity two | WS-03 optimization | TODO | T-010 | browser scheduler | static schedule proof and retained focused lifecycle evidence | direct aggregate behavior matches release behavior, leaf summaries remain distinct, and no visual or fixture drift occurs |
| T-012 | Generate public-target baselines and enforce baseline-derived acceptance | WS-04 acceptance | TODO | T-008, T-009, T-010, T-011 | harness performance | baseline and performance-check summaries | required hotspots improve and all other targets stay within budget |
| T-013 | Run broad verification and close the handoff | WS-04 handoff | TODO | T-012 | integrator | final verification matrix and handoff log | clean tree, terminal tasks, no unresolved blocker |

Provisional implementation currently present in the worktree (none of these
statements closes a reopened task until its corrected exit condition passes):

- The harness NLSpec owns observability coverage, reconstruction, privacy,
  failure isolation, export, and performance-window behavior. The OTel NLSpec
  delegates the harness profile and preserves application telemetry behavior.
- Seven authored observability/performance schemas are attached to the harness
  registry. Generated contracts, task-surface outputs, scheduler inputs, and
  browser projections were produced through `make generate`.
- `tools/harness/observability/` reconstructs traces from authoritative native
  evidence with built-in Node facilities, validates closed/redacted bundles,
  renders performance diagnostics, checks evidence windows, and performs only
  explicit post-run OTLP export.
- Normal finalization records an observability warning in retained structured
  output when generation is partial; it does not change the underlying test
  result or consume the command's stderr budget.
- Backend-unit exact-symbol compatibility grouping reduced the observed Go
  process count from 34 to 30. Deterministic batch ingestion and bounded report
  emission are implemented, but T-009 remains open until a qualifying baseline
  establishes and clears its improvement gate.
- Shared DAG execution is implemented for `lint`, `ci`, and `release-check`.
  The first release layout was rejected after it exposed shared Compose service
  interference; the corrected browser-only service-backed aggregate is
  generated but still needs focused and full release execution evidence.

## 6. Optimization contracts

Backend exact-symbol rows share one process only when package selection,
runtime-binary set, fixture and isolation policy, and evidence class match. Raw
selectors remain separate. Compatible groups run with
`clamp(floor(availableParallelism/4), 1, 4)` workers. Missing, duplicate,
unexpected, crashed, or partial Go JSON output fails closed.

Batch emission may perform independent parsing and file writes concurrently,
but final summaries, artifact arrays, and failure selection use the existing
deterministic orders.

`check` remains the first dependency of `ci` and `release-check`. After it
passes, harness-contract, security, licensing/SBOM, builds, SeaweedFS leaves,
and browser readiness may overlap under declared resource claims. The SeaweedFS
release gate waits for compatibility, migration, license, and SBOM. Release
readiness waits for every required release summary. Direct browser leaf targets
remain isolated. The release sequence invokes one internal
`release-browser-readiness` service-backed schedule rather than concurrently
invoking direct leaf targets against shared development services. That schedule
owns one isolated test-services stack, admits at most two browser sessions, and
retains distinct support, visual, accessibility, and aggregate summaries. Its
five sessions preserve default versus claimed Network Flow runtime profiles and
carry explicit isolation reasons. Every additional isolation boundary requires
an owner reason.

## 7. Verification and acceptance

Trace fixtures cover direct targets, nested sequences, parallel siblings,
scheduler waits, resource blockers, shared services, repeated targets,
finalizers, failure, interruption, malformed clocks, missing artifacts, and
byte-identical reconstruction. Privacy fixtures inject credentials, URLs,
paths, commands, environment values, SQL-like text, and product-like output.
Export fixtures use a local receiver and prove endpoint construction, payload
decoding, no redirects, timeout behavior, redaction, and zero ordinary-run
egress.

For a three-sample warm set, let `m` be the median duration and `d` the median
absolute deviation. The no-regression limit is
`m + max(1000 ms, 3d, 0.05m)`. A required hotspot improves only when the
candidate median is at least `max(1000 ms, 3d, 0.10m)` below the baseline
median. Required improvement gates are backend-unit, backend finalization,
release browser readiness, and release-check. All other covered commands use
the no-regression limit. The separate one-warm-plus-five-measured `make check`
acceptance with maximum below 120 seconds remains controlling.

Verification order is schema and unit fixtures; generated policy and drift;
harness and OTel conformance; focused backend, sequence, browser, exporter, and
observability targets; `test-fast`; `check`; `ci`; `release-check`; performance
windows; then `agent-finalize` against the successful full warm root.

### Current evidence ledger

Successful retained evidence on the current worktree lineage:

| Command/evidence | Run root | Result and diagnostic value |
| --- | --- | --- |
| `make run-harness-smoke-extended` | `.cartulary/test-results/20260720T171451Z-p3532051` | PASS, 180.115 s; lifecycle and harness fixture coverage before the release-browser correction |
| `make json-shape-check` | `.cartulary/test-results/20260720T171827Z-p3575733` | PASS |
| `make generated-artifact-policy-check` | `.cartulary/test-results/20260720T171829Z-p3576053` | PASS |
| `make generate-drift` | `.cartulary/test-results/20260720T171830Z-p3576225` | PASS |
| `make harness-contract` | `.cartulary/test-results/20260720T171840Z-p3579100` | PASS, 21.840 s |
| `make otel-conformance` | `.cartulary/test-results/20260720T171903Z-p3579927` | PASS |
| `make lint-shell` | `.cartulary/test-results/20260720T171910Z-p3582860` | PASS |
| `make lint-scripts` | `.cartulary/test-results/20260720T171923Z-p3583763` | PASS |
| `make backend-unit` | `.cartulary/test-results/20260720T171939Z-p3584225` | PASS, 255 tests in 13.110 s; actual dependency critical path 4.790 s, batch finalizer 2.990 s, unattributed envelope 0.610 s |
| `make test-fast` | `.cartulary/test-results/20260720T172016Z-p3587832` | PASS, 651 tests in 124.185 s |
| `make check` | `.cartulary/test-results/20260720T172228Z-p3629581` | PASS, 122.369 s; observability check PASS, but above the strict 120 s acceptance and not a qualifying five-run window |
| `make ci` | `.cartulary/test-results/20260720T172600Z-p3716898` | PASS, 146.607 s; nested `check` 117.158 s and four post-check steps overlapped |
| `make generate` | `.cartulary/test-results/20260720T174945Z-p3997826` | PASS after the browser-only release schedule correction |

The backend, `check`, and `ci` roots pass strict observability reconstruction.
Their measurements are useful diagnostics only; they are not interchangeable
with the exact three-root baseline/candidate windows.

Failed, contaminated, or incomplete evidence retained for diagnosis:

| Run root | Classification | Finding and disposition |
| --- | --- | --- |
| `.cartulary/test-results/20260720T172842Z-p3834220` | Contaminated; exclude | Initial `release-check` ran direct browser leaves and SeaweedFS work concurrently against shared development Compose services. Browser observations were lost and SeaweedFS was recreated. Replaced by the isolated `release-browser-readiness` schedule. |
| `.cartulary/test-results/20260720T174659Z-p3995630` | Failed generation; exclude | A stage-wide release session group mixed default and claimed Network Flow profiles. The schedule now preserves five profile-compatible sessions with group-level isolation reasons. |
| `.cartulary/test-results/20260720T175017Z-p3999438` | Failed fixture; exclude | The fast sequence smoke fake did not emit the three browser leaf summaries produced by the new aggregate. The fixture was updated. |
| `.cartulary/test-results/20260720T175230Z-p4005166` | Interrupted; exclude | Follow-up execution smoke rerun was interrupted before aggregate finalization; retained child artifacts are incomplete. |

The current post-correction generated state still requires a fresh
`generate-drift`, harness-contract/OTel pass, focused
`release-browser-readiness`, complete execution smoke, and full
`release-check`. Do not reuse any excluded root for T-001 or T-012.

## 8. Handoff protocol and log

Every status transition appends a dated record with source commit and worktree
state, completed work, substantive files or contracts changed, commands and run
roots, failed or contaminated evidence, generated status, and the one next
active task. A resumer first validates the current active task and its
dependencies, then continues from retained evidence rather than repeating
completed work.

### 2026-07-20 — Tracker creation

- Source: clean `revision/grid-adapter` at `00522cfed1b6e5ca0936fb703de96c4c019544f3`.
- Completed: authority review, domain classification, initial hotspot scan, and
  successor tracker creation.
- Active: T-001 baseline and command-inventory freeze.
- Retained seed: `.cartulary/test-results/20260720T141740Z-p2992081`.
- Qualification warning: the seed is a single observation and does not satisfy
  the three-observation performance contract.

### 2026-07-20 — Specification, observability, and backend implementation

- Source: uncommitted implementation worktree based on
  `00522cfed1b6e5ca0936fb703de96c4c019544f3`; authored and generated changes are
  present.
- Completed: T-002 through T-008. The harness and telemetry owner documents,
  policy metadata, seven schemas, deterministic reconstruction, lifecycle
  capture, local validation/finalization, explicit OTLP export, backend exact
  symbol grouping, and batch report machinery are implemented.
- Contract state: native JSON remains authoritative, no OTel SDK dependency was
  added, ordinary test execution has no network export, and application OTel
  scopes/bootstrap remain unchanged. `docs/domain.md` was reviewed without edit.
- Verification: use the successful roots in the current evidence ledger. The
  retained backend run proves 255 selected tests and the 34-to-30 process
  reduction; `check` and `ci` demonstrate the shared scheduler and local trace
  reconstruction.
- Acceptance warning: T-009 remains open because its baseline-derived finalizer
  improvement threshold cannot be evaluated from the seed observation.
- Active: T-001 remains the only `IN_PROGRESS` task.

### 2026-07-20 — Release scheduling correction and current handoff

- Source: same uncommitted worktree; generated outputs were refreshed by
  `make generate` at
  `.cartulary/test-results/20260720T174945Z-p3997826`.
- Finding: the first release DAG incorrectly overlapped direct browser leaves
  and SeaweedFS commands that reconfigured the same development Compose stack.
  The failed run is retained at
  `.cartulary/test-results/20260720T172842Z-p3834220` and is excluded from all
  performance windows.
- Correction: `release-check` now invokes one browser-only
  `release-browser-readiness` service-backed schedule after `check`; it uses the
  isolated test-services wrapper, receives browser-stack capacity two at the
  release Make boundary, retains the three leaf summaries plus its aggregate,
  and preserves default/claimed runtime-profile session boundaries.
- Fixture follow-up: the sequence smoke fake now emits the aggregate's browser
  child summaries. The next public smoke rerun was interrupted, so it does not
  qualify as verification.
- Generated status: current authored topology and generated files agree as of
  the successful `make generate`; post-correction drift and broad checks remain
  to be rerun.
- Next active task: T-001. First finish post-correction verification in the
  prescribed order, then collect exact three-observation baseline/candidate
  windows. Do not advance T-009, T-010, T-011, or T-012 from diagnostic timings
  alone.

### 2026-07-20 — Remediation audit and status correction

- Source: the same staged implementation worktree based on
  `00522cfed1b6e5ca0936fb703de96c4c019544f3`; no prior root or log entry was
  deleted or reinterpreted.
- Audit finding: implementation presence did not satisfy the adopted contracts.
  In particular, public-target coverage was incomplete, parameterized workload
  identities were not stable, retained roots inherited profiles from the
  current checkout, validation rewrote selected evidence, scheduler causality
  was incomplete, sequence execution duplicated scheduling policy, OTLP names
  diverged from the specification, and baseline regeneration had no owner.
- Status correction: T-002 and T-004 through T-008 are reopened. T-003 remains
  complete because the application-versus-harness boundary itself is unchanged,
  but it is revalidated as a dependency of T-004. T-001 now follows the
  corrected measurement and validation substrate; candidate optimizations
  follow T-001.
- Diagnostic-only probes after the audit: `make generate-drift` passed at
  `.cartulary/test-results/20260720T181200Z-p4022613`;
  `make json-shape-check` passed at
  `.cartulary/test-results/20260720T181214Z-p4025502`; `make harness-contract`
  exited successfully; and `make otel-conformance` passed at
  `.cartulary/test-results/20260720T181240Z-p4026615`. These roots predate the
  remediation and cannot enter reference or candidate windows.
- Active: T-002 is the only `IN_PROGRESS` task. The next transition is T-002 to
  `DONE` and T-004 to `IN_PROGRESS` after the corrected normative ownership and
  measurement definitions are complete.

### 2026-07-20 — T-002 specification correction complete

- Source: staged implementation worktree based on
  `00522cfed1b6e5ca0936fb703de96c4c019544f3`; worktree remains intentionally
  dirty for the remediation.
- Completed: T-002. The harness specification now owns explicit three-way public
  target dispositions, canonical measurement profiles, immutable retained
  execution context, read-only exact-run checks, scheduler v7 and sequence
  topology, union timing, schedule-owned browser capacity, exporter failure
  classes, complete backend compatibility, and the sole baseline writer.
- Boundary revalidation: T-003 remains complete. The application instrumentation
  scope, bootstrap, configuration, browser prohibition, and dependencies did not
  change; only the harness-owned diagnostic profile was refined.
- Verification: `make lint-markdown` passed at
  `.cartulary/test-results/20260720T183906Z-p4036805`; `git diff --check` passed.
  This root is specification-workstream evidence, not a qualifying performance
  observation.
- Generated status: not yet refreshed; schema and owner-input changes belong to
  active T-004 and must be followed by `make generate`.
- Active: T-004 is the only `IN_PROGRESS` task.

### 2026-07-20 — T-004 contracts and projections complete

- Source: staged remediation worktree at
  `00522cfed1b6e5ca0936fb703de96c4c019544f3`; the worktree remains
  intentionally dirty and no qualifying performance root was collected.
- Completed: T-004. The task-surface owner now accounts for every public target
  exactly once across required, excluded, and out-of-scope dispositions; all
  required measurements bind an authored profile. Execution topology owns
  sequence dependencies, priorities, generic resource claims, and capacities.
  The attachment registry now adopts execution context and scheduler v7 while
  retiring the staged scheduler v6 and mutating observability-check summary.
- Contract fixtures: generation rejects disposition omissions, overlap,
  unknown targets, unowned exclusions, duplicate sequence targets, missing
  profiles, dependency cycles, and topology/task-surface divergence. Two
  duplicate `target` members found in the prior owner JSON were removed rather
  than allowed to remain parser-dependent.
- Verification: `make generate` passed at
  `.cartulary/test-results/20260720T191615Z-p4068960`;
  `make json-shape-check` passed at
  `.cartulary/test-results/20260720T191620Z-p4070424`;
  `make generated-artifact-policy-check` passed at
  `.cartulary/test-results/20260720T191443Z-p4060533`;
  `make generate-drift` passed at
  `.cartulary/test-results/20260720T191443Z-p4060550`; and
  `make harness-contract` passed at
  `.cartulary/test-results/20260720T191814Z-p4073065`.
- Excluded diagnostics: failed generation and contract roots from this slice are
  retained for investigation only. They exposed stale owner-document digests,
  missing help placement, attachment ordering, duplicate JSON members, and the
  expected public-interface digest update; none is eligible for performance
  evidence.
- Active: T-005 is the only `IN_PROGRESS` task. Retained execution context and
  deterministic reconstruction must complete before scheduler/sequence
  lifecycle work begins.

### 2026-07-20 — T-005 retained provenance and reconstruction complete

- Source: staged remediation worktree at
  `00522cfed1b6e5ca0936fb703de96c4c019544f3`; no reference or candidate
  performance window was started.
- Completed: T-005. Every required top-level wrapper now captures an immutable
  execution context and derives its trace, OTLP payloads, hotspot summary, and
  digest-bearing index from run-relative native evidence. Read-only loading
  verifies context, source, and derived-artifact digests before independently
  reconstructing and byte-comparing the bundle. Current-checkout policy and
  profile values are not substituted into retained roots.
- Reconstruction semantics: target parentage comes only from explicit retained
  summary relationships; scheduler dependency paths come from retained v7
  edges; direct-child attribution, per-resource blocking, and backend/report
  finalization use interval unions. Exact `cartulary.harness` service and metric
  identifiers are emitted, and external owner-only result roots do not leak
  absolute paths.
- Fixtures: the observability golden covers explicit nested parentage,
  monotonic scheduler boundaries, dependency paths, resource and capacity
  waits, overlapping finalizers, hostile source values, malformed clocks,
  multiple parents, source and derived tampering, deterministic bytes,
  read-only verification, and an external result root.
- Verification: `make generate` passed at
  `.cartulary/test-results/20260720T193341Z-p4089526`; the Make-owned
  `make harness-contract` run passed at
  `.cartulary/test-results/20260720T193347Z-p4090998`, including the
  observability golden. Its own retained observability bundle also passes
  independent loading; it remains performance-ineligible because the source is
  intentionally dirty.
- Excluded diagnostic: `make run-harness-smoke-extended` failed at
  `.cartulary/test-results/20260720T193013Z-p4082069` because the legacy
  sequence lifecycle producer still emitted retired event tokens. That is the
  controlling T-006 defect, not qualifying T-005 evidence.
- Active: T-006 is the only `IN_PROGRESS` task. The sequence runtime must now
  compile topology-owned schedules through the shared scheduler and retire the
  duplicate Bash DAG lifecycle.

### 2026-07-20 — T-006 shared sequence scheduler complete

- Source: staged remediation worktree at
  `00522cfed1b6e5ca0936fb703de96c4c019544f3`; the worktree remains
  intentionally dirty and all roots in this slice are diagnostic-only.
- Completed: T-006. Task-surface serial and DAG sequences now compile the
  execution-topology schedule into the existing scheduler engine under
  `scheduler_kind=sequence`. The former Bash scheduler is a thin Node launcher;
  it no longer parses colon-delimited topology, accounts resources, selects
  failures, polls children, or owns cancellation.
- Lifecycle and scheduling semantics: generic logical resources, topology
  priorities and dependencies, dependency skips, running-sibling drain,
  process-group interruption, shared finalizer rules, deterministic same-turn
  completion draining, and public summary order use the common engine. Retained
  scheduler v7 state and the sequence-event adapter use monotonic boundaries,
  exact dependency edges, authored step identity, resource claims, and one
  terminal transition per started or skipped step. The release browser schedule
  itself owns `browser_stack=2`; the obsolete parent Make override and its stale
  fixture expectation are removed.
- Contract changes: `cartulary.sequence_scheduler_summary.v1` gives shared
  sequence summaries an exact schema identity, and scheduler pressure evidence
  accepts the `sequence` kind. The scheduler event writer is closed before
  public target-summary finalization, preventing observability from reading a
  partially flushed event stream.
- Verification: `make generate` passed at
  `.cartulary/test-results/20260720T195017Z-p4122461`;
  `make json-shape-check` passed at
  `.cartulary/test-results/20260720T195029Z-p4123993`;
  `make harness-contract` passed at
  `.cartulary/test-results/20260720T195036Z-p4124388` in 22.660 s;
  `make run-harness-smoke-extended` passed at
  `.cartulary/test-results/20260720T195429Z-p4167515` in 159.755 s;
  `make lint-shell` passed at
  `.cartulary/test-results/20260720T195721Z-p15094`;
  `make lint-scripts` passed at
  `.cartulary/test-results/20260720T195739Z-p16041`;
  `make generate-drift` passed at
  `.cartulary/test-results/20260720T195758Z-p16493`; and
  `make generated-artifact-policy-check` passed at
  `.cartulary/test-results/20260720T195757Z-p16480`. The focused sequence suites
  also pass directly and now assert typed scheduler/sequence evidence, causal
  dependency skips, dry-run non-execution, and arbitrary logical-resource
  contention.
- Excluded diagnostic: the first extended smoke rerun at
  `.cartulary/test-results/20260720T195118Z-p4125762` retained a stale fixture
  that expected release-browser capacity from a parent environment override.
  The schedule was already correct; the fixture was updated to require
  schedule-owned capacity two. This failed root is excluded from all
  qualification windows.
- Active: T-007 is the only `IN_PROGRESS` task. Exact selected read-only
  validation, OTLP export behavior, privacy, timeout, redirect, and zero-egress
  fixtures must close before reference baseline collection.

### 2026-07-20 — T-007 read-only validation and explicit export complete

- Source: staged remediation worktree at
  `00522cfed1b6e5ca0936fb703de96c4c019544f3`; the worktree remains
  intentionally dirty and no root from this slice is performance-qualified.
- Completed: T-007. `harness-observability-check` now resolves only an explicit
  run directory or result root plus `RUN_ID`, loads and independently
  reconstructs the retained bundle entirely in memory, and emits no summary or
  other mutation into the selected run. Missing, partial, malformed, tampered,
  or nondeterministic evidence exits `11`; invalid selection exits `2`.
- Export boundary: `harness-otel-export` uses the same exact read-only loader,
  accepts only HTTPS or loopback HTTP without encoded authority ambiguity,
  appends the two OTLP signal paths exactly once, disables redirects, performs
  no retries, and applies the fixed 10,000 ms request timeout. Endpoint, header,
  permission, selection, or retained-input configuration fails with exit `2`;
  resolution, connection, redirect, timeout, HTTP, or other delivery failure
  exits `1`. Diagnostics contain no endpoint, path, or header value.
- Fixtures: the observability golden verifies exact run-directory and
  root-plus-ID selection, result-root ambiguity rejection, read-only success
  and tamper failure by whole-tree digest, hostile endpoint and header shapes,
  byte-counted header values, symlink and mode rejection, exact OTLP decoding
  through a loopback receiver, redirect and non-success rejection, timeout,
  bounded failure text, source-tree immutability, and zero ordinary-run fetches.
- Verification: exact public checks passed against
  `.cartulary/test-results/20260720T195036Z-p4124388` both as a run directory and
  as `.cartulary/test-results` plus `RUN_ID`; `make otel-conformance` passed at
  `.cartulary/test-results/20260720T200409Z-p24040`;
  `make harness-contract` passed at
  `.cartulary/test-results/20260720T200613Z-p28610` in 22.860 s;
  `make lint-scripts` passed at
  `.cartulary/test-results/20260720T200646Z-p29651`;
  `make json-shape-check` passed at
  `.cartulary/test-results/20260720T200646Z-p29659`;
  `make generated-artifact-policy-check` passed at
  `.cartulary/test-results/20260720T200646Z-p29666`; and
  `make generate-drift` passed at
  `.cartulary/test-results/20260720T200708Z-p30670`.
- Active: T-001 is the only `IN_PROGRESS` task. The next step is to preserve
  this candidate worktree in a local safety commit, derive a clean reference
  commit with the corrected measurement substrate and serial policies, and
  collect the exact warm reference window before changing candidate code.

### 2026-07-21 — T-005 reopened by reference-window qualification

- Source: clean safety snapshot
  `528ef57d03c12dd47c0991a03380b0167ae54459`; clean serial reference lineage
  `87d418f8d7e921b792292aa752e6df1024576c5f` before this correction.
- Qualification finding: the sole baseline writer rejected the explicit
  63-root reference manifest with `retained observability index is incomplete`
  at `.cartulary/test-results/20260721T040348Z-p2475899`. All three direct
  `browser-e2e` roots finalized observability before the enclosing scheduler
  summary settled. All three direct `seaweedfs-release-evidence` roots attempted
  to parent retained prerequisite summaries beneath the later, narrow release
  evidence step, producing intervals outside that root. The local diagnostic
  failure did not change the six commands' passing native results.
- Status correction: T-005 is reopened and is the only `IN_PROGRESS` task;
  T-001 returns to `TODO`. The correction must retain an explicit top-level
  invocation boundary and prerequisite relationships, then finalize scheduler
  diagnostics only after terminal scheduler evidence is written. T-006 remains
  complete subject to this boundary revalidation because scheduler v7,
  sequence behavior, cancellation, and failure precedence did not regress.
- Excluded evidence: every root on the pre-correction reference commit remains
  available for diagnosis but is ineligible for T-001. In particular, direct
  SeaweedFS roots `20260721T030019Z-p2041667`,
  `20260721T030241Z-p2061268`, and `20260721T030504Z-p2080884`, and direct
  browser roots `20260721T031143Z-p2132670`,
  `20260721T031601Z-p2164901`, and `20260721T032031Z-p2197290`, are partial and
  must not be rewritten or migrated. The other 57 manifest roots are also
  diagnostic-only because the source correction changes the reference
  snapshot.
- Next active task: T-005. Add fixture-backed invocation-boundary and terminal
  scheduler-finalization semantics, regenerate adopted contracts through
  `make generate`, and verify exact read-only loading before restarting T-001
  from a new clean reference commit.

### 2026-07-21 — T-005 invocation-boundary correction complete

- Source: clean serial reference substrate commit
  `91b7a277851876c4b705445c2770e7a14f3598f8`; the validation roots below were
  collected immediately before that commit and remain diagnostic-only because
  their source state was dirty.
- Completed: T-005. Public preflight now retains
  `cartulary.harness_invocation_start.v1` before prerequisite work, including a
  sorted recursive snapshot of generated Make target-prerequisite edges. The
  terminal execution context binds the marker and edge snapshot; its root span
  covers the public invocation envelope. Reconstruction filters and parents
  target summaries through retained invocation, summary, sequence, or scheduler
  relationships. Performance qualification rejects a missing boundary as
  `artifact_incomplete`.
- Scheduler boundary: direct browser targets defer observability while the
  in-scheduler target summary is produced, then finalize only after terminal
  scheduler summary, pressure, progress, and event evidence is closed. Child
  scheduler invocations remain suppressed and cannot replace the aggregate's
  top-level boundary.
- Focused defect proof: direct `seaweedfs-release-evidence` passed at
  `.cartulary/test-results/20260721T041917Z-p2486191`; its complete context spans
  161.975 s, retains nine prerequisite edges, and passes exact read-only
  reconstruction across 36 sources. Direct `browser-e2e` passed at
  `.cartulary/test-results/20260721T042215Z-p2507131`; its complete context spans
  263.846 s, retains five prerequisite edges, and passes exact read-only
  reconstruction across 66 sources. Neither root is eligible for T-001.
- Verification: `make generate` passed at
  `.cartulary/test-results/20260721T041832Z-p2483554`; `make harness-contract`
  passed at `.cartulary/test-results/20260721T041844Z-p2485089`;
  `make json-shape-check` passed at
  `.cartulary/test-results/20260721T042733Z-p2540037`;
  `make generated-artifact-policy-check` passed at
  `.cartulary/test-results/20260721T042736Z-p2540394`; `make generate-drift`
  passed at `.cartulary/test-results/20260721T042738Z-p2540575`;
  `make lint-scripts` passed at
  `.cartulary/test-results/20260721T042747Z-p2543482`; `make otel-conformance`
  passed at `.cartulary/test-results/20260721T042749Z-p2543827`; and
  `make lint-markdown` passed at
  `.cartulary/test-results/20260721T042812Z-p2546905`.
- Active: T-001 is the only `IN_PROGRESS` task. All reference warm-ups and
  accepted observations must be recollected from the next clean commit; no
  pre-correction root may be migrated or rewritten.

### 2026-07-21 — T-001 lease-fixture qualification correction opened

- Source: clean serial reference substrate
  `3242235b03837825d248975060cb963e9f70ae95` before this correction.
- Qualification finding: consecutive broad reference windows exposed a flaky
  timing assumption in
  `TestPostgresApplicationProcessLease_Integration`. The fixture constructs a
  lazy `pgxpool` and gives the first `processlease.Acquire` call only 100 ms, so
  under broad-suite load that interval can expire while opening the first
  network session before any advisory-lock attempt occurs. The production
  contract and default remain unchanged: Core 04 requires acquisition within
  the configured timeout and the adopted configuration range is 1 through 300
  seconds with a 30-second default.
- Required correction: pre-establish the real PostgreSQL sessions used by the
  fixture before starting its deliberately short acquisition and contention
  checks. Keep `processlease.Acquire`, the PostgreSQL backend, production
  defaults, error mapping, and product semantics unchanged. Add focused proof
  that session warm-up failure is reported before the lease assertions.
- Excluded evidence: `ci` root `20260721T055637Z-p3613990` and `test` root
  `20260721T063045Z-p101648` failed at the invalid fixture boundary. The prior
  failed `ci` root `20260721T054534Z-p3200942` remains excluded for its listener
  conflict. Every warm-up and accepted observation collected from source
  `3242235b03837825d248975060cb963e9f70ae95` remains diagnostic-only because the
  fixture source correction changes the reference snapshot; no root will be
  migrated, rewritten, or mixed into the replacement window.
- Active: T-001 remains the only `IN_PROGRESS` task. Correct and repeatedly
  verify the focused lease row, run the harness contract and fast regression
  gates, commit a new clean reference source, and restart every baseline
  measurement profile from that exact source.

### 2026-07-21 — T-001 lease-fixture qualification correction complete

- Source: clean correction commit
  `218aa49906a2894293a8d9a4ed723273daecf881`; the validation roots below were
  collected immediately before that commit and remain diagnostic-only because
  their source state was dirty.
- Completed correction: the real PostgreSQL integration fixture now opens and
  releases two pool sessions under a separate five-second setup deadline before
  exercising the unchanged 100 ms and 20 ms lease-acquisition cases. Holding
  both setup sessions until warm-up is complete forces the pool to establish
  both physical connections. The subsequent short intervals therefore measure
  advisory-lock acquisition and contention, including crash-release and
  loss-detection behavior, rather than lazy connection establishment.
- Focused proof: the exact
  `app.server.integration.extension_application_process_lease_43130392c4` row
  passed against three fresh harness-owned PostgreSQL stacks at roots
  `.cartulary/test-results/20260721T063845Z-p174574`,
  `.cartulary/test-results/20260721T063852Z-p174825`, and
  `.cartulary/test-results/20260721T063859Z-p175020`, with one selected test
  executed exactly once in 4.781 s, 5.029 s, and 4.785 s respectively.
- Regression proof: `make harness-contract` passed independently at
  `.cartulary/test-results/20260721T063915Z-p177564` and
  `.cartulary/test-results/20260721T063945Z-p178538`; `make test-fast` passed at
  `.cartulary/test-results/20260721T064047Z-p179753` with 651 tests and two of
  two work units complete in 163.555 s. Product lease code, configuration,
  error mapping, and the Core 04 contract are unchanged.
- Active: T-001 remains the only `IN_PROGRESS` task. Its discarded warm-ups
  and three-observation windows now restart from the clean tracker-completion
  commit that contains this ledger entry; every earlier source root remains
  diagnostic-only.

### 2026-07-21 — T-001 readiness-fixture qualification correction opened

- Source: clean serial reference snapshot
  `e2fd5ba2bc38f9f202bde1911ff7849c108bc90a` before this correction.
- Qualification finding: accepted `release-check` observation
  `.cartulary/test-results/20260721T083833Z-p1760362` failed in the
  `dev service lifecycle guards are mutation-safe` harness-contract fixture.
  The fixture configured mocked PostgreSQL and object-store readiness deadlines
  to one second. `wait_object_store` starts that deadline before building and
  launching the CORS proxy, so process scheduling under release load can cross
  the one-second boundary before the first mocked probe is attempted. The
  resulting `state=unknown health=unknown` is a fixture timing failure, not a
  service, product, or production-readiness failure.
- Required correction: give the mocked lifecycle fixture a bounded scheduling
  margin that is independent of production readiness configuration. Preserve
  the production wait loops, deadlines, diagnostics, proxy lifecycle, and
  service commands unchanged. Repeated focused harness-contract runs must prove
  that the stubbed first probe is reached under ordinary host load.
- Excluded evidence: the release warm-up
  `20260721T083022Z-p1611432`, failed accepted attempt
  `20260721T083833Z-p1760362`, and every other warm-up or accepted observation
  collected from `e2fd5ba2bc38f9f202bde1911ff7849c108bc90a` remain retained for
  diagnosis but are ineligible after the source correction. No measurement
  window or individual target root from that snapshot may be migrated.
- Active: T-001 remains the only `IN_PROGRESS` task. Correct the test-only
  deadline, run repeated harness-contract and fast regression proof, commit a
  new clean reference snapshot, and restart the complete baseline manifest.

### 2026-07-21 — T-001 readiness-fixture qualification correction complete

- Source: clean correction commit
  `b1067ca059fb1d3863535cc401de5cbdf69b6f48`; the validation roots below were
  collected immediately before that commit and remain diagnostic-only because
  their source state was dirty.
- Completed correction: the mutation-safety fixture now gives its fake
  PostgreSQL and object-store readiness commands a five-second bounded deadline
  instead of one second. This margin exists only in the test process and lets
  the mocked proxy build, launch, and first probe run even when the host is
  scheduling a full release workload. Production readiness defaults, loops,
  diagnostics, service commands, and proxy lifecycle are byte-unchanged.
- Focused proof: five consecutive `make harness-contract` runs passed at
  `.cartulary/test-results/20260721T084239Z-p1846527`,
  `.cartulary/test-results/20260721T084304Z-p1847457`,
  `.cartulary/test-results/20260721T084329Z-p1848398`,
  `.cartulary/test-results/20260721T084354Z-p1849362`, and
  `.cartulary/test-results/20260721T084419Z-p1850327`. `make lint-shell` passed
  at `.cartulary/test-results/20260721T084452Z-p1851302`.
- Regression proof: `make test-fast` passed at
  `.cartulary/test-results/20260721T084510Z-p1852230` with 651 tests and both
  aggregate work units complete in 162.670 s.
- Active: T-001 remains the only `IN_PROGRESS` task. The complete baseline
  manifest now restarts from the clean tracker-completion commit containing
  this ledger entry; all earlier source roots remain diagnostic-only.

### 2026-07-21 — T-001 browser-discovery qualification correction opened

- Source: clean serial reference snapshot
  `7d7a87e19a8052981b1a309a496fd1a46da3f1a3` before this correction.
- Qualification finding: two consecutive broad `test` windows failed in
  different browser groups while direct browser and release windows remained
  green. Root `20260721T104005Z-p3266085` retained a 350-pixel Network Flow
  timestamp-text golden delta with identical content, layout, and font-manifest
  digest; it is excluded as a visual-render diagnostic. In the restarted
  window, root `20260721T104851Z-p3402375` failed while Network Flow displayed
  the healthy intermediate status `Discovering source columns.` The shared
  import helper exhausted Playwright's generic five-second expectation before
  the real discovery request completed under the broad scheduler's observed
  five concurrent browser stacks.
- Required correction: make the Network Flow import helper wait up to 30
  seconds for its semantic mapping-dialog readiness boundary, consistent with
  other real upload/import helpers. Do not add sleeps, retries, newest-result
  fallback, or product changes. Preserve the import request, UI states,
  application deadlines, browser topology, golden bytes, and failure behavior.
  Repeated claimed-profile stateful and broad service-backed evidence must pass
  before a new clean reference snapshot is admitted.
- Excluded evidence: both failed `test` windows, including their earlier
  passing warm-ups and accepted attempts, are ineligible. Every release,
  owner-slice, audit, browser, lint, or test-fast observation collected from
  `7d7a87e19a8052981b1a309a496fd1a46da3f1a3` also remains diagnostic-only after
  this source correction; no root may be migrated or rewritten.
- Active: T-001 remains the only `IN_PROGRESS` task. Correct the helper,
  complete focused and broad browser regression proof, commit a new clean
  source snapshot, and restart the complete baseline manifest.

### 2026-07-21 — T-001 browser-discovery qualification correction complete

- Source: clean correction commit
  `fa05592de09df6206c2847f744d9bf6141bab21f`; the validation roots below were
  collected immediately before that commit and remain diagnostic-only because
  their source state was dirty.
- Completed correction: the Network Flow import helper now gives the semantic
  mapping-dialog readiness boundary an explicit 30-second deadline. The
  browser still waits on the real dialog state without a sleep or retry, and
  the import request, intermediate status, application deadlines, execution
  topology, golden bytes, and product behavior are unchanged.
- Focused proof: three consecutive `make browser-e2e-stateful` runs passed at
  `.cartulary/test-results/20260721T105458Z-p3471070`,
  `.cartulary/test-results/20260721T105801Z-p3492880`, and
  `.cartulary/test-results/20260721T110100Z-p3514228`. Each run completed both
  stateful sessions and their finalizers with scheduler capacity
  `browser_stack=1`; the previously failing claimed Network Flow profile
  reached its mapping dialog successfully in every observation.
- Regression proof: `make test-fast` passed at
  `.cartulary/test-results/20260721T110405Z-p3535521` with 651 tests, zero
  failures, and both aggregate work units complete in 160.452 s. `make format`
  passed at `.cartulary/test-results/20260721T105448Z-p3468716` before the
  repeated browser checks.
- Active: T-001 remains the only `IN_PROGRESS` task. The complete baseline
  manifest restarts from the clean tracker-completion commit containing this
  ledger entry. Every root from this correction and every earlier reference
  source remains diagnostic-only; none may be migrated into the new window.

### 2026-07-21 — T-001 contributor-readiness qualification correction opened

- Source: clean serial reference snapshot
  `c997c00cbf31dbd1a128d6692685112f0a8fc620` before this correction.
- Qualification finding: the discarded `test` warm-up at
  `.cartulary/test-results/20260721T115205Z-p432270` failed the claimed Network
  Flow contributor golden by 350 pixels. The retained expected image contains
  the normalized `2025-01-01` contributor timestamps, while the actual image
  contains the original fixture `2026-07-10` timestamps. The contributor drawer
  becomes visible before its asynchronous query has populated the semantic
  grid, so under broad load the visual-state masker can run before those text
  nodes exist. The two failed broad roots produced byte-identical actual
  images, ruling out font or raster variance.
- Required correction: make the claimed visual fixture await retained
  contributor content inside the visible drawer before preparing the visual
  regression state. Preserve the product query, grid rendering, timestamp
  fixture, dynamic-text normalization, golden bytes, pixel threshold, and
  browser topology. Do not hide the cells, weaken comparison, add a sleep, or
  retry the failed run.
- Excluded evidence: warm-up root `20260721T115205Z-p432270` is retained only
  for diagnosis. The completed release and CI windows from
  `c997c00cbf31dbd1a128d6692685112f0a8fc620` are also ineligible after this
  source correction; no accepted root from that snapshot may be migrated.
- Active: T-001 remains the only `IN_PROGRESS` task. Add the semantic readiness
  assertion, prove the focused claimed visual repeatedly and the broad test
  path, commit a new clean source snapshot, and restart every baseline profile.

### 2026-07-21 — T-001 contributor-readiness qualification correction complete

- Source: clean correction commit
  `16d6d48fc12b2ecbe037702339a658339bc22b80`; the validation roots below were
  collected immediately before that commit and remain diagnostic-only because
  their source state was dirty.
- Completed correction: the claimed Network Flow visual fixture now waits for
  the retained `visual-flow` contributor content inside the visible drawer
  before preparing deterministic visual state. This closes the asynchronous
  query boundary without sleeping or retrying. Product data, query execution,
  grid rendering, timestamp normalization, golden bytes, zero-tolerance visual
  comparison, and execution topology are unchanged.
- Focused proof: the exact
  `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6`
  row passed three consecutive service-backed owner slices at
  `.cartulary/test-results/20260721T115931Z-p501817`,
  `.cartulary/test-results/20260721T120035Z-p517120`, and
  `.cartulary/test-results/20260721T120129Z-p531908`, with one selected test
  executed exactly once per root and the existing contributor golden accepted.
- Regression proof: `make test` passed at
  `.cartulary/test-results/20260721T120228Z-p546721` with 651 tests, zero
  failures, and both aggregate work units complete in 222.787 s. `make format`
  passed at `.cartulary/test-results/20260721T115851Z-p499231` before the
  focused checks.
- Active: T-001 remains the only `IN_PROGRESS` task. Every discarded warm-up
  and accepted observation now restarts from the clean tracker-completion
  commit containing this ledger entry. All roots from
  `c997c00cbf31dbd1a128d6692685112f0a8fc620` and earlier remain diagnostic-only.

### 2026-07-21 — T-001 grouped-conflict qualification correction opened

- Source: clean serial reference snapshot
  `e5dc295349606c5a2dff067212cd4f9596fa2a90` before this correction.
- Qualification finding: accepted `test` observation
  `.cartulary/test-results/20260721T125804Z-p1804662` failed because the
  collaboration fixture expected the inline grid editor to remain mounted
  after a same-field conflict. This scenario deliberately groups rows by
  capture state and sorts them by the edited synopsis. Applying the remote
  synopsis moves the row within the grouped/sorted projection and legitimately
  unmounts its editor. Retained page state proves the conflict resolver held the
  local and remote values and the remounted row exposed the conflict marker.
- Required correction: use the existing grouped-projection branch of
  `driveRealTimelineSummaryConflict` for this call by declaring that the edited
  cell is expected to unmount. Continue to require the remounted conflict
  marker, resolver, exact local and remote values, save-state transition, and
  server result. Preserve product grouping, sorting, conflict behavior, helper
  defaults for stable cells, timeouts, and browser topology.
- Focused refinement: dirty validation root
  `.cartulary/test-results/20260721T130308Z-p1871819` produced the other valid
  projection state: the inline editor remained mounted with the local value,
  so no conflict-marker button existed. Together with the qualification root,
  this proves that the grouped/sorted grid representation is event-order
  dependent. The correction must therefore use both existing projection
  opt-outs for this call and make the resolver values, conflict save state,
  selected resolution, and final server value authoritative. It must not poll
  for or prefer either transient cell representation.
- Excluded evidence: failed root `20260721T125804Z-p1804662` and the preceding
  accepted `test` root `20260721T125418Z-p1734532` are ineligible. The completed
  release and CI windows from
  `e5dc295349606c5a2dff067212cd4f9596fa2a90` also become diagnostic-only after
  this source correction; no root from that snapshot may be migrated.
- Active: T-001 remains the only `IN_PROGRESS` task. Correct the scenario
  declaration, prove the exact collaboration row repeatedly and the broad test
  path, commit a new clean source snapshot, and restart every baseline profile.

### 2026-07-21 — T-001 grouped-conflict qualification correction complete

- Source: clean correction commit
  `be31f9a06a3531484b9aa8aeab98d0bba39073a9`; the validation roots below were
  collected immediately before that commit and remain diagnostic-only because
  their source state was dirty.
- Completed correction: the grouped/sorted collaboration scenario now uses the
  shared helper's two projection opt-outs, so it does not require either the
  transient inline editor or the transient remounted marker. The helper still
  requires `Conflict`, a visible resolver, exact saved and unsaved values, and
  the scenario still selects `Use my unsaved value`, waits for `Saved`, and
  verifies that exact value through the server API. Stable-cell callers retain
  the stricter mounted-editor and marker defaults.
- Focused proof: the exact
  `module.collaboration.browser_stateful.verify_multi_client_live_row_update_presence_anc_4c65a275c1`
  row passed three consecutive service-backed owner slices at
  `.cartulary/test-results/20260721T130634Z-p1890248`,
  `.cartulary/test-results/20260721T130730Z-p1905550`, and
  `.cartulary/test-results/20260721T130826Z-p1920400`, with one selected test
  executed exactly once per root.
- Regression proof: `make test` passed at
  `.cartulary/test-results/20260721T130935Z-p1935289` with 651 tests, zero
  failures, and both aggregate work units complete in 215.685 s. `make format`
  passed at `.cartulary/test-results/20260721T130609Z-p1887772` before the
  repeated focused checks.
- Active: T-001 remains the only `IN_PROGRESS` task. Every baseline warm-up and
  accepted observation restarts from the clean tracker-completion commit
  containing this ledger entry. All roots from
  `e5dc295349606c5a2dff067212cd4f9596fa2a90` and earlier remain diagnostic-only.

### 2026-07-21 — T-001 Network Flow lifecycle-settlement correction opened

- Source: clean serial reference snapshot
  `c33f7097be36094c8794636d427bc326fb847298` before this correction.
- Qualification finding: accepted `test` observation
  `.cartulary/test-results/20260721T131801Z-p2075751` failed during the recovery
  import in the Network Flow lifecycle scenario. The retained network trace
  proves that the second upload was accepted three milliseconds after the
  delete response and its background job succeeded with a valid import-session
  resource in about 27 ms. The fixture had waited only for the locally updated
  empty workspace, not for the client's replayable
  `extension_resource_changed/remove` event. That later event legitimately
  clears protected state and increments the import operation generation, so it
  can invalidate a recovery import started before lifecycle settlement.
- Required correction: install the existing incident websocket monitor before
  workbook navigation, record its message boundary before deletion, and wait
  for the exact Network Flow table-removal event for the deleted resource
  before submitting the recovery import. Preserve product event handling,
  import cancellation, upload/job behavior, the 30-second semantic dialog
  deadline, and browser topology. Do not add sleeps, retries, or product-side
  self-event suppression as a test workaround.
- Excluded evidence: failed root `20260721T131801Z-p2075751` and discarded
  warm-up root `20260721T131414Z-p2005889` are diagnostic-only. No root from
  `c33f7097be36094c8794636d427bc326fb847298` may enter a qualified window.
- Active: T-001 remains the only `IN_PROGRESS` task. Add the websocket semantic
  boundary, prove the exact lifecycle row repeatedly and the broad `test` path,
  commit a new clean source snapshot, and restart every baseline profile.

### 2026-07-21 — T-001 Network Flow lifecycle-settlement correction complete

- Source: clean correction commit
  `d68790dcdfdb22bc9530b833c2c892e47817c0ae`; the validation roots below were
  collected immediately before that commit and remain diagnostic-only because
  their source state was dirty.
- Completed correction: the Network Flow opening helper now exposes a bounded
  pre-navigation incident callback. The lifecycle fixture uses it to attach the
  existing incident socket monitor before the workbook websocket is created,
  records a message boundary before deletion, and waits for the exact
  `extension_resource_changed` event with profile `network_flow_activity`,
  resource kind `network_flow_table`, change kind `remove`, and a non-empty
  resource ID before submitting recovery import. Product event processing,
  import-generation cancellation, upload/jobs, and topology are unchanged.
- Focused proof: the exact
  `module.networkflow.browser_stateful.verify_protected_network_analysis_state_is_disca_21a5de1ebf`
  row passed three consecutive final-shape service-backed owner slices at
  `.cartulary/test-results/20260721T133224Z-p2196030`,
  `.cartulary/test-results/20260721T133313Z-p2211348`, and
  `.cartulary/test-results/20260721T133406Z-p2226135`, with one selected test
  executed exactly once per root. Three earlier dirty focused roots also passed
  before the monitor reference received its type-safe shape and remain
  diagnostic-only.
- Regression proof: `make frontend-typecheck` passed at
  `.cartulary/test-results/20260721T133206Z-p2195569`; `make test` passed at
  `.cartulary/test-results/20260721T133508Z-p2240942` with 651 tests, zero
  failures, and both aggregate work units complete in 220.914 s. `make format`
  passed at `.cartulary/test-results/20260721T132704Z-p2143082`.
- Excluded deterministic failure: broad root `20260721T133029Z-p2190479`
  exposed and retained the initial TypeScript closure-narrowing error; no
  runtime test was admitted from it.
- Active: T-001 remains the only `IN_PROGRESS` task. Every baseline warm-up and
  accepted observation restarts from the clean tracker-completion commit
  containing this ledger entry. All prior roots remain diagnostic-only.

### 2026-07-21 — T-001 direct duration-coverage identity correction opened

- Source: clean serial reference snapshot
  `0ea8d0cdb01684089cd58001f8524cbcdc13bd49` before this correction.
- Qualification finding: direct required-profile invocation
  `make go-test-duration-baseline-coverage` failed at retained root
  `.cartulary/test-results/20260721T150156Z-p3921353` because the Make-node
  contract removed `CARTULARY_TEST_RESULTS_DIR` and `CARTULARY_TEST_RUN_ID`
  before the coverage CLI wrote its retained public summary. The top-level
  preflight had already created `_shared/harness-invocation-start.json`, so the
  child attempted to allocate the same caller run identity as an unprepared
  non-empty root and rejected it. This makes the required direct measurement
  profile non-invocable even though aggregate coverage execution can pass.
- Required correction: declare the retained run identity variables as runtime
  environment for the duration-baseline coverage Make-node tool, add a focused
  contract fixture that proves direct prepared-identity reuse and retained
  summary/observability closure, regenerate the task surface only through
  `make generate`, and prove drift plus direct repeated execution. Preserve the
  public command, command ID, optional baseline-file input, coverage behavior,
  summary schema, and strict rejection of arbitrary non-empty caller roots.
- Excluded evidence: all warm-up and accepted roots from source snapshot
  `0ea8d0cdb01684089cd58001f8524cbcdc13bd49` remain diagnostic-only. This
  includes the completed `test`, `test-fast`, `lint`, `ci`, `release-check`,
  `agent-finalize`, `benchmark-claim-check`, and `frontend-fallow-static`
  windows. The retry-contaminated `test-fast` root
  `20260721T135951Z-p2633099` is independently excluded for
  `retry_observed`. No observation from this snapshot may be migrated into the
  restarted reference window.
- Active: T-001 remains the only `IN_PROGRESS` task. Correct and validate the
  direct identity contract, commit a new clean source snapshot, then restart
  every reference baseline profile from a fresh discarded warm-up.

### 2026-07-21 — T-001 direct duration-coverage identity correction complete

- Source: implementation commit
  `cd8e066b` on the serial reference branch; all validation roots below were
  collected from the immediately preceding dirty implementation state and are
  diagnostic-only.
- Completed correction: the generated `RUN_MAKE_NODE_TOOL` boundary now marks
  the exact public target as using the identity already validated by its Make
  preflight. The duration-baseline coverage tool declares the retained result
  root and run ID as child runtime environment. The resolver's strict rule is
  unchanged: an unprepared caller cannot reuse a non-empty root, including a
  root containing an imitated invocation-start marker.
- Focused proof: the exact public command passed at
  `.cartulary/test-results/20260721T150657Z-p3925814` and then three consecutive
  times at `.cartulary/test-results/20260721T150759Z-p3930152`,
  `.cartulary/test-results/20260721T150800Z-p3930200`, and
  `.cartulary/test-results/20260721T150802Z-p3930242`. Each retained the
  `go-test-duration-baseline-coverage/tool-run-summary.json` artifact under its
  own preflighted root.
- Contract and drift proof: `make harness-contract` passed at
  `.cartulary/test-results/20260721T150707Z-p3925897`, including the negative
  unprepared/non-empty collision fixtures. `make generate-drift` passed at
  `.cartulary/test-results/20260721T150740Z-p3926876`, `make json-shape-check`
  passed at `.cartulary/test-results/20260721T150750Z-p3929775`, and
  `make generated-artifact-policy-check` passed at
  `.cartulary/test-results/20260721T150808Z-p3930324`. Generated task-surface
  projections were produced only by `make generate`.
- Active: T-001 remains the only `IN_PROGRESS` task. Every required profile now
  restarts from the clean tracker-completion commit containing this ledger
  entry; all roots from `0ea8d0cdb01684089cd58001f8524cbcdc13bd49` and all
  dirty validation roots above remain diagnostic-only.
