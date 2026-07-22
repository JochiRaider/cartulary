# Test Harness Observability and Wall-Clock Optimization Tracker

## 1. Tracker control

| Field | Value |
| --- | --- |
| State | ACTIVE |
| Primary seam | Public Make invocation -> harness execution graph -> retained timing graph -> derived OpenTelemetry diagnostics |
| Initial source | `00522cfed1b6e5ca0936fb703de96c4c019544f3` on `revision/grid-adapter` |
| Current source | T-011 browser-capacity remediation complete; T-012 clean candidate acceptance is next |
| Last updated | 2026-07-22 |
| Active item | T-012 |
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
- Public Make names, selected tests, owner and evidence routing, failure
  precedence, artifact roles, and cleanup behavior are preserved. The two
  performance command IDs intentionally advance to v2 with their incompatible
  evidence contracts.
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

T-001 replaced this seed with target/provider-scoped windows containing one
validated discarded warm-up and exactly two measured observations. Aggregate
runs supply only exact-once scheduler-work leaf envelopes; direct roots use the
public invocation envelope, and backend finalization uses the native
`report_collation` interval union. Failed, interrupted, stale,
capacity-mismatched, source-mismatched, or retry-contaminated runs remain
retained but do not enter the accepted sample set.

The current authored inventory classifies 48 public testing entry points as
observability-required, four post-run diagnostic/export/maintenance commands
as owner-cited exclusions, and 45 public commands as explicitly out of scope.
Performance evidence adds one internal `release-browser-readiness` profile and
one synthetic `backend-output-finalizer` metric, for 50 baseline rows total.
The provisional machine profile is frozen by these digests:

| Profile component | SHA-256 |
| --- | --- |
| Host | `c31c805927c8912d7b56ec09c32cc91b8597404122198b80c923970c546f542d` |
| Capacity | `bd042c914469a69fa6ffda55f0f96e65e968d3a34a64aecd58790dada66d2df1` |
| Workload/evidence inventory | `ff7d5a4d9cab0b3ff2fae01ca0537fd19ad884b6c4bca98fce4b533ae9a88f1d` |
| Toolchain | `877e09fe43b77b76264d81f02ac74aaaaeafe700bba05511449f890ffb18c9ac` |

The old global digests and single seed observation are superseded. The v2
baseline retains target-local provenance because independent reference targets
may come from different immutable clean snapshots. Candidate windows remain
strict-current evidence from one frozen clean commit.

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
| T-001 | Qualify retained target/provider reference windows and generate the v2 public-target duration baseline | WS-00 baseline | DONE | T-007 | harness performance | retained 12-hour audit, v2 roots manifest, generated baseline artifact, strict public-invocation identity proof, semantic execution-policy digest proof, and isolated digest-only v1 comparison proof | one discarded warm-up plus exactly two measured observations cover 48 public rows, one internal row, and one synthetic row; all rejected roots retain reasons; every strict direct provider retains its exact public invocation boundary; producer and validator share one order-independent execution-policy canonicalization contract; digest-only v1 rows can prove unchanged policy without weakening strict candidates |
| T-002 | Correct observability, artifact, security, scheduler, and performance requirements and tracker ownership | WS-01 specification | DONE | none | harness specification | `docs/testing-harness-nlspec.md`; requirements and acceptance crosswalk | every behavior has one normative owner and every measurement identity is reproducible |
| T-003 | Add the application-versus-harness OTel boundary | WS-01 specification | DONE | T-002 | telemetry specification | `docs/opentelemetry-instrumentation-nlspec.md`; OTel conformance fixtures | application scopes and runtime signals remain unchanged |
| T-004 | Add explicit target dispositions, stable measurement profiles, retained execution context, corrected schemas, attachments, and generated projections | WS-01 contracts | DONE | T-002, T-003 | harness contracts | closed public inventory, authored schemas, and reproducible generated projections | omissions, overlap, unknowns, duplicates, and unowned exclusions fail generation |
| T-005 | Implement retained-provenance deterministic reconstruction and interval-union hotspot analysis | WS-02 observability | DONE | T-004 | diagnostics | immutable context plus deterministic native, trace, metric, hotspot, and digest fixtures | retained roots reconstruct independently of the checkout and explicit graph parentage, paths, waits, gaps, and digests validate |
| T-006 | Unify sequence and scheduler lifecycle evidence, topology ownership, cancellation, and deterministic failure behavior | WS-02 observability | DONE | T-005 | execution runtime | scheduler v7, shared sequence scheduler, and lifecycle fixtures | required transitions and dependencies are attributable exactly once under success, failure, and interruption |
| T-007 | Make local validation read-only and exact-selected; correct OTLP export, privacy, and failure semantics | WS-02 observability | DONE | T-005, T-006 | diagnostics/export | tamper, exact-selection, OTLP decode, failure-class, redirect, timeout, and egress fixtures | selected source evidence is never mutated and export conforms exactly |
| T-008 | Consolidate compatible backend-unit exact symbols and run compatible groups concurrently | WS-03 optimization | DONE | T-001 | backend runner | retained 255-test parity run and 30-process plan/run proof | every symbol and row is proven exactly once across complete compatibility keys and failure paths |
| T-009 | Parse each physical Go report once and parallelize deterministic family projection emission | WS-03 optimization | DONE | T-008 | output/finalizers | worker failure fixtures, retained parity evidence, and strict warm-up diagnosis | output identity, partial-success retention, and primary-failure selection are stable; strict candidate finalizer union clears its improvement gate |
| T-010 | Execute `lint`, `ci`, and `release-check` through the topology-owned shared scheduler | WS-03 optimization | DONE | T-001, T-007 | scheduler/task surface | serial and DAG parity evidence for all three aggregates plus concurrent websocket teardown, browser-worker admission, retained-session, priority-liveness, prerequisite-admission, bounded security-analysis regressions, and exact transition-validator parity | dependency, resource, cancellation, output, cleanup, primary-failure behavior, and normalized transition validation are stable |
| T-011 | Make release browser readiness own its five-session schedule and capacity two | WS-03 optimization | DONE | T-010 | browser scheduler | static schedule proof, retained focused lifecycle evidence, exact policy-transition fixtures, and managed-service aggregate concurrency proof | direct aggregate behavior matches release behavior, leaf summaries remain distinct, no visual or fixture drift occurs, and both browser policies are ready for T-012 quantitative acceptance |
| T-012 | Generate public-target baselines and enforce baseline-derived acceptance | WS-04 acceptance | IN_PROGRESS | T-008, T-009, T-010, T-011 | harness performance | baseline and performance-check summaries | required hotspots improve and all other targets stay within budget |
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
  process count from 34 to 30. Physical reports are parsed once, and immutable
  family projections emit through the six-worker host-derived pool. Functional
  parity is closed; the quantitative comparison remains T-012-owned.
- Shared DAG execution is implemented and validated for `lint`, `ci`, and
  `release-check`. Sequence-owned internal Make prerequisites honor explicit
  scheduler suppression without weakening direct invocation; full-capacity
  security analysis and deferred OTel source validation have retained timing
  proof, and release browser readiness remains a browser-only service-backed
  aggregate.

## 6. Optimization contracts

Backend-unit exact-symbol rows share one process only when package selection,
sorted runtime-binary set, complete fixture profile/policy/budget, isolation
policy, and evidence class match. Raw selectors remain separate. Compatible
backend-unit groups run with
`min(group_count, clamp(floor(availableParallelism/4), 1, 8))` workers and
scheduler-owned child `GOMAXPROCS=max(1,floor(availableParallelism/workers))`.
On the current 24-CPU host this resolves to six workers and four Go scheduler
threads per child. Missing, duplicate, unexpected, crashed, or partial Go JSON
output fails closed. Other backend targets retain the host parallelism assigned
to their scheduler unit; the backend-unit capture partition must not escape
into those independently scheduled children.

Each physical Go report is parsed and validated exactly once. Explicit frozen
family-projection requests then merge and emit concurrently through the same
host-derived pool as capture. All workers settle before final summaries,
artifact arrays, diagnostics, and primary failures are selected in authored
order. Successful worker output remains eligible when a sibling fails.

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

For each two-sample measured set, `m` is the arithmetic midpoint and `d` is the
arithmetic midpoint of the absolute deviations. The no-regression limit is
`m + max(1000 ms, 3d, 0.05m)`. A required hotspot improves only when the
candidate median is at least `max(1000 ms, 3d, 0.10m)` below the baseline
median. Required improvement gates are backend-unit, backend finalization,
release browser readiness, and release-check. All other covered commands use
the no-regression limit. Every one of the 48 public rows must pass and their
candidate median sum must be strictly below the reference sum. The separate
one-warm-plus-five-measured `make check`
acceptance with maximum below 120 seconds remains controlling.

Verification order is schema and unit fixtures; generated policy and drift;
harness and OTel conformance; focused backend, sequence, browser, exporter, and
observability targets; `test-fast`; `check`; `ci`; `release-check`; performance
windows; then `agent-finalize` against the successful full warm root.

### Current evidence ledger

Successful retained evidence on the current worktree lineage:

| Command/evidence | Run root | Result and diagnostic value |
| --- | --- | --- |
| retained v1 reference migration audit | `/home/jochi/code/cartulary-reference-baseline/.cartulary/test-results`, cutoff `2026-07-21T10:04:31Z` through `2026-07-21T22:04:31Z` | 368 contexts inspected; 291 nominally eligible; 218 strict accepts; 73 strict rejects; 21 deduplicated windows bind 50 rows without new baseline execution |
| `make harness-public-target-duration-baselines EVIDENCE_ROOTS_FILE=.cartulary/performance-reference-roots.v2.json` | `.cartulary/test-results/20260721T224115Z-p3902667` | PASS; deterministic qualified v2 baseline with 48-target portfolio total `2721126 ms`, internal browser row, synthetic native-finalizer row, and all 73 strict rejection reasons |
| `make generate` owner-manifest refresh and ordinary generation | `.cartulary/test-results/20260721T224243Z-p3904051`; `.cartulary/test-results/20260721T224259Z-p3905640` | PASS; harness owner digest refreshed before ordinary generated projections |
| T-001 focused drift and shape gates | `.cartulary/test-results/20260721T224419Z-p3907674`; `.cartulary/test-results/20260721T224419Z-p3907695`; `.cartulary/test-results/20260721T224456Z-p3913421` | `generate-drift`, generated-artifact policy, and corrected JSON shape PASS |
| `make harness-contract` | `.cartulary/test-results/20260721T224549Z-p3915005` | PASS after intentional v2 public command-interface checksum update |
| T-008 planner and contract fixtures | `.cartulary/test-results/20260721T230009Z-p3937762` | `harness-contract` PASS; current catalog fixture proves 30 physical capture groups, 251 exact symbols, one isolated raw selector, 34 family projections, six workers, and child `GOMAXPROCS=4` |
| `make backend-unit` | `.cartulary/test-results/20260721T230040Z-p3938914` | PASS, 255 tests, failed/missing `0`, 30 complete physical report roots, 34 logical emissions (`actual=30`, `reused=4`), 32.744 s wall; quantitative acceptance remains T-012-owned |
| T-008 generated drift and shape | `.cartulary/test-results/20260721T230204Z-p3941056`; `.cartulary/test-results/20260721T230204Z-p3941058` | `generate-drift` and `json-shape-check` PASS after removing the obsolete fixed shard-job input |
| T-010 retained-browser scheduler matrix | `.cartulary/test-results/20260722T152800Z-p584361` | Extended harness smoke PASS on the final implementation, including exact retained lifecycle claims, two-token browser CPU admission, dual-session feasibility, measurement isolation, reservation-liveness fallback, failure drain, and cleanup |
| `make test` | `.cartulary/test-results/20260722T151807Z-p302020` | PASS, 651 tests and all 147 service-backed work units; observed ceiling 18 concurrent units, Go CPU 16, Go I/O 24, process 12, and both isolated measurement sessions completed without deadlock |
| T-010 owner and aggregate validation | `.cartulary/test-results/20260722T150928Z-p210325`; `.cartulary/test-results/20260722T152209Z-p374034`; `.cartulary/test-results/20260722T152426Z-p460599`; `.cartulary/test-results/20260722T152449Z-p466433` | Exact workbook browser row, `check`, `lint`, and `ci` PASS; standalone `check` was 129.057 s diagnostic-only, while CI's nested `check` was 112.937 s |
| `make release-check` | `.cartulary/test-results/20260722T153047Z-p627614` | PASS, 12/12 sequence units and 510 tests in 286.918 s; release browser readiness 139.721 s; nested contract, drift, shape, security, build, SeaweedFS, SBOM/license, and readiness branches all passed |
| T-010 final generated validation | `.cartulary/test-results/20260722T152730Z-p580020`; `.cartulary/test-results/20260722T152730Z-p580049`; `.cartulary/test-results/20260722T152730Z-p580025` | `generate-drift`, `json-shape-check`, and generated-artifact policy PASS on implementation commit `64dd1023` |
| T-011 focused browser validation | `.cartulary/test-results/20260722T225518Z-p2927420`; `.cartulary/test-results/20260722T225808Z-p2953187` | `release-browser-readiness` PASS in 132.596 s with five profile-compatible sessions, peak `browser_stack=2`, and visual/accessibility capacity one; `browser-e2e-visual` PASS in 122.500 s with no golden drift |
| T-011 managed-service concurrency proof | `.cartulary/test-results/20260722T231142Z-p3012400` | `make test` PASS, 651 tests in 230.524 s; both isolated webserver sessions started six milliseconds apart, peak `browser_stage_webserver_backed=2`, all 147 service-backed units and finalizers completed, and no fixture or cleanup failure occurred |
| T-011 contract and generated validation | `.cartulary/test-results/20260722T231109Z-p3007366`; `.cartulary/test-results/20260722T231109Z-p3007361`; `.cartulary/test-results/20260722T231109Z-p3007386` | `json-shape-check`, generated-artifact policy, and `generate-drift` PASS; `harness-contract` also passed from the same validation batch |
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
| `.cartulary/test-results/20260722T230342Z-p2980285` | Failed diagnostic experiment; exclude | A proposed direct two-session webserver scheduler raced the shared development Compose/CORS service boundary; one session failed during `compose up`. The direct-capacity experiment and v7 manifest draft were removed. Direct leaves remain isolated, while the adopted two-lane policy is limited to the concurrency-safe managed test-service aggregate. |

The current optimization implementation has passed its focused generation,
scheduler, browser, check, lint, and CI validation. Candidate qualification
must use only roots collected after the clean T-012 activation checkpoint; do
not reuse an excluded or diagnostic root for T-001 or T-012.

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

### 2026-07-21 — T-001 Timeline mention-focus continuity correction opened

- Source: clean serial reference snapshot
  `15af60c9448ceca088875637ed64c3163219b722` before this correction.
- Qualification finding: discarded `test` warm-up root
  `.cartulary/test-results/20260721T163540Z-p517797` failed row
  `module.entities.browser.the_browser_workbook_shows_auto_resolution_only_758d41a16f`
  after a successful auto-resolution Undo request. The notice disappeared, the
  relationship became unresolved, the inspected row remained visible, and the
  grid retained its inspector-active semantics, but the Timeline synopsis cell
  lost DOM focus for the full five-second assertion window.
- Root cause: focus return is an existing behavior contract, already asserted
  by the component test that focuses the Undo button before activation. The
  live implementation restores focus only synchronously, on the next animation
  frame, and at timeout zero. A later React/grid reconciliation can replace the
  focused cell after those attempts, leaving focus on the document even though
  the action and viewport continuity succeeded.
- Required correction: make mention-action focus return bounded and resilient
  across late reconciliation. Re-resolve the semantic cell on each attempt,
  retry for a short closed window, and stop immediately if the user moves focus
  to a different connected element. Preserve mention mutation behavior,
  viewport/scroll continuity, inspector selection, accessibility intent,
  timeouts, public commands, and browser topology. Add component coverage for
  late target replacement and user-directed focus cancellation, then prove the
  exact service-backed row repeatedly and the broad `test` path.
- Excluded evidence: every warm-up and accepted observation from source
  snapshot `15af60c9448ceca088875637ed64c3163219b722` is diagnostic-only. This
  includes all completed direct-profile windows and their retained module-auth
  audit fixture. Auth slice root `20260721T153554Z-p111938` is independently
  excluded for `retry_observed`. No root from this snapshot may be migrated.
- Active: T-001 remains the only `IN_PROGRESS` task. Correct and validate focus
  continuity, commit a new clean snapshot, then restart every reference
  baseline profile from a new discarded warm-up.

### 2026-07-21 — T-001 Timeline mention-focus continuity correction complete

- Source: implementation commit
  `1455fbcaa51db24133ea1e5faf6d4f03514e3adf` on the serial reference branch;
  all validation roots below were collected from the immediately preceding
  dirty implementation state and remain diagnostic-only.
- Corrected diagnosis: the opened entry's late-reconciliation hypothesis was
  incomplete. Instrumented retained root
  `.cartulary/test-results/20260721T165540Z-p636165` proved that Undo created
  the focus-bearing `row-inspect` continuity token, then the same mutation's
  lower-strength live-update path created a `scroll-only` token in the same
  user-interaction generation and superseded it. The replacement token never
  reached an advance boundary, so no owner remained able to restore focus.
  The temporary diagnostic instrumentation was removed before validation.
- Completed correction: Timeline focus and scroll restoration now have one
  owner. Mention actions no longer run a private synchronous/frame/timeout
  focus helper. The viewport-continuity controller retains its active request
  synchronously, coalesces same-interaction `scroll-only` work into the active
  semantic request, and advances that request instead of replacing it. A new
  pointer, keyboard, or wheel interaction creates a new generation and may
  supersede or cancel older work. The controller continues its existing
  bounded retries until the semantic target is connected, focused, and fully
  visible, then clears immediately. Its redundant uncancellable inner
  animation-frame loop was removed.
- Rejected tactical variant: an intermediate one-second post-success polling
  window fixed the focus row but overrode a later intentional visual-fixture
  scroll. Failed broad root
  `.cartulary/test-results/20260721T171020Z-p806897` and focused visual root
  `.cartulary/test-results/20260721T171511Z-p875851` retain that evidence. The
  polling variant was removed; no visual golden, product mutation, inspector
  behavior, timeout, public command, or browser topology changed.
- Focused proof: the exact
  `module.entities.browser.the_browser_workbook_shows_auto_resolution_only_758d41a16f`
  row passed three consecutive final-shape service-backed owner slices at
  `.cartulary/test-results/20260721T171942Z-p899832`,
  `.cartulary/test-results/20260721T172042Z-p915612`, and
  `.cartulary/test-results/20260721T172138Z-p930665`, with one selected test
  executed exactly once per root. Earlier failed focused roots
  `20260721T164826Z-p592076`, `20260721T165117Z-p612444`, and
  `20260721T165540Z-p636165` are retained as diagnostic progression only.
- Regression proof: `make frontend-unit` passed at
  `.cartulary/test-results/20260721T171906Z-p897784`;
  `make frontend-typecheck` passed at
  `.cartulary/test-results/20260721T171933Z-p899443`; focused
  `make browser-e2e-visual` passed without golden changes at
  `.cartulary/test-results/20260721T172240Z-p945537`; and `make test` passed at
  `.cartulary/test-results/20260721T172458Z-p964463` with 651 tests, zero
  failures, and both aggregate work units complete in 233.741 s.
- Active: T-001 remains the only `IN_PROGRESS` task. Every required reference
  profile restarts from the clean tracker-completion commit containing this
  ledger entry. All roots from
  `15af60c9448ceca088875637ed64c3163219b722` and all dirty correction roots
  above remain diagnostic-only and cannot enter the qualified baseline window.

### 2026-07-21 — T-001 Go vulnerability qualification correction complete

- Source: clean pre-correction measurement snapshot
  `ada4993b5fe7f56c8864293bb1472d780982a5a6`; implementation commit
  `45d54ef7` contains the correction. The tracker-completion commit containing
  this entry is the next clean reference source.
- Qualification finding: `ci` warm-up root
  `.cartulary/test-results/20260721T195208Z-p2706015` failed the required
  `go-vulncheck` gate with four symbol-reachable instances of `GO-2026-5970`
  (`CVE-2026-56852`), an infinite-loop vulnerability in
  `golang.org/x/text`. The retained findings identify `v0.37.0` as affected,
  `v0.39.0` as fixed, and the reachable path through pgx SCRAM normalization
  into the PostgreSQL test harness. The scan database was last modified on
  2026-07-17; the earlier direct pass used stale cached vulnerability data and
  cannot qualify a later aggregate that saw the current database.
- Completed correction: the runtime module requirement is now
  `golang.org/x/text v0.39.0`; tool-owned module reconciliation also advanced
  the compatible indirect `golang.org/x/sync` requirement to `v0.21.0` and
  refreshed only those checksums. The immutable Network Flow Unicode 17
  normalization sources and provenance remain correctly attributed to their
  vendored `x/text v0.37.0` source snapshot; generated Unicode tables and
  product behavior are unchanged.
- Validation proof: `make go-vulncheck` passed at
  `.cartulary/test-results/20260721T195629Z-p2788014`; `make build-server`
  passed at `.cartulary/test-results/20260721T195647Z-p2788519`;
  `make generate-drift` passed at
  `.cartulary/test-results/20260721T195708Z-p2798028`; `make toolchain-drift`
  passed at `.cartulary/test-results/20260721T195708Z-p2798029`; and
  `make test-fast` passed at
  `.cartulary/test-results/20260721T195726Z-p2801325` with 651 tests, zero
  failures, and both aggregate work units complete. These roots were produced
  before the clean implementation/tracker commits and remain diagnostic-only.
- Excluded measurement evidence: every warm-up and accepted observation from
  source snapshot `ada4993b5fe7f56c8864293bb1472d780982a5a6` is
  diagnostic-only, including all completed direct profiles and the `lint`,
  `test-fast`, and `test` aggregate windows. Failed `test` roots
  `20260721T191059Z-p2131856` and `20260721T192730Z-p2314125` are additionally
  excluded for independent browser interaction and worker-teardown
  infrastructure failures; each exact affected row then passed three
  consecutive diagnostic owner slices. Failed `ci` root
  `20260721T195208Z-p2706015` is excluded for the security finding. No root
  from this snapshot may be migrated into the restarted baseline.
- Active: T-001 remains the only `IN_PROGRESS` task. Restart every required
  profile from the clean tracker-completion commit with a fresh discarded
  warm-up. Retain all prior roots for diagnosis and rejection-reason evidence.

### 2026-07-21 — T-001 partial clean reference collection wrapped

- Source: clean serial reference measurement commit
  `334465751e05c9d50e00416fc4a071d579bfd719`, source snapshot digest
  `6a84166396613d24eeb9d43433244ce70165797fba49ca85979ccad4f5be8a85`,
  host digest
  `60b1b04a7acd7959935651192dec03d8d910225859996bc97dea68cc58ba6f76`,
  capacity digest
  `0eb1de6a10f344907ea84b64550d893e0151b679026ced392464f491ca7b1664`,
  and toolchain digest
  `877e09fe43b77b76264d81f02ac74aaaaeafe700bba05511449f890ffb18c9ac`.
  Every accepted root below retains `source_state=clean`, `status=passed`,
  `warm_eligibility=eligible`, `retry_count=0`, `interrupted=false`, and no
  contamination reason.
- Direct-profile result: all 16 profiles that are not supplied by the locked
  aggregate set completed one discarded warm-up followed by three accepted
  observations. Root IDs below are relative to `.cartulary/test-results/`.

  | Profile | Discarded warm-up | Three accepted observations |
  | --- | --- | --- |
  | `agent-finalize` | `20260721T200234Z-p2860669` | `20260721T200237Z-p2860847`, `20260721T200239Z-p2861030`, `20260721T200241Z-p2861213` |
  | `benchmark-claim-check` | `20260721T200243Z-p2861396` | `20260721T200244Z-p2861556`, `20260721T200246Z-p2861722`, `20260721T200247Z-p2861882` |
  | `browser-e2e` | `20260721T205819Z-p3421944` | `20260721T210244Z-p3454354`, `20260721T210718Z-p3486881`, `20260721T211147Z-p3519261` |
  | `frontend-fallow-static` | `20260721T200248Z-p2862044` | `20260721T200255Z-p2862565`, `20260721T200303Z-p2863104`, `20260721T200310Z-p2863619` |
  | `go-gosec-targeted` | `20260721T200434Z-p2875843` | `20260721T200443Z-p2895954`, `20260721T200450Z-p2914450`, `20260721T200457Z-p2932899` |
  | `go-test-duration-baseline-coverage` | `20260721T200317Z-p2864140` | `20260721T200319Z-p2864183`, `20260721T200321Z-p2864231`, `20260721T200322Z-p2864286` |
  | `go-vulncheck` | `20260721T200420Z-p2874129` | `20260721T200424Z-p2874565`, `20260721T200427Z-p2874995`, `20260721T200430Z-p2875420` |
  | `json-shape-check` | `20260721T200324Z-p2864328` | `20260721T200327Z-p2864670`, `20260721T200330Z-p2865011`, `20260721T200332Z-p2865352` |
  | `migration-drift` | `20260721T200335Z-p2865697` | `20260721T200346Z-p2867674`, `20260721T200355Z-p2869380`, `20260721T200405Z-p2871082` |
  | `seaweedfs-release-evidence` | `20260721T201019Z-p3093872` | `20260721T201322Z-p3114931`, `20260721T201544Z-p3134547`, `20260721T201758Z-p3154150` |
  | `service-backed-test-slice OWNER=module.auth` | `20260721T203938Z-p3300326` | `20260721T204406Z-p3330786`, `20260721T204845Z-p3361233`, `20260721T205329Z-p3391580` |
  | `standup-operational-recovery-smoke` | `20260721T200742Z-p3023656` | `20260721T200821Z-p3041202`, `20260721T200900Z-p3058744`, `20260721T200940Z-p3076324` |
  | `standup-package-smoke` | `20260721T200528Z-p2952151` | `20260721T200604Z-p2970265`, `20260721T200636Z-p2987893`, `20260721T200709Z-p3005774` |
  | `test-evidence-audit OWNER=module.auth` | `20260721T203918Z-p3300057` | `20260721T203920Z-p3300108`, `20260721T203922Z-p3300163`, `20260721T203924Z-p3300218` |
  | `test-slice OWNER=module.auth` | `20260721T202019Z-p3173789` | `20260721T202453Z-p3207267`, `20260721T202933Z-p3238095`, `20260721T203413Z-p3269142` |
  | `toolchain-drift` | `20260721T200414Z-p2872778` | `20260721T200416Z-p2873118`, `20260721T200417Z-p2873452`, `20260721T200418Z-p2873792` |

- Parameterized-profile proof: both slice commands used the locked
  `OWNER=module.auth` input. Each `test-slice` root executed 48 tests across
  nine rows, each service-backed slice root executed 29 tests across nine
  rows, and the audit fixture was built from retained slice root
  `20260721T203413Z-p3269142`; each audit observation accounted for all 48
  expected tests.
- Completed aggregate result: `lint` passed all eight sequence steps. Its
  discarded warm-up is `20260721T211614Z-p3551561`; accepted roots are
  `20260721T211648Z-p3565016`, `20260721T211716Z-p3571305`, and
  `20260721T211744Z-p3577572`.
- Completed aggregate result: `test-fast` passed 651 tests with both aggregate
  work units complete and no missing evidence. Its discarded warm-up is
  `20260721T211819Z-p3583897`; accepted roots are
  `20260721T212103Z-p3625349`, `20260721T212351Z-p3666804`, and
  `20260721T212640Z-p3708442`.
- Incomplete aggregate result: `test` warm-up root
  `20260721T212932Z-p3749876` passed 651 tests with both aggregate work units
  complete in 235.246 s. Collection was then stopped at the user's wrap-up
  request while the first measured observation was active. Retained root
  `20260721T213330Z-p3820013` records `status=failed`,
  `warm_eligibility=ineligible`, `contamination_reasons=[failed_execution]`,
  and `interrupted=false`; it is excluded as an operator-cancelled observation.
  The retained interruption boolean is reported as written rather than
  reinterpreted. The affected `test` window must restart with a new discarded
  warm-up.
- Not collected: no observation from this snapshot exists for `ci` or
  `release-check`. The public-target roots manifest was therefore not created,
  `tools/harness_public_target_duration_baselines.json` was not regenerated,
  and no qualification or performance claim is made.
- Commands: the direct commands above were invoked four times each through
  their public Make targets; `test-slice` and `service-backed-test-slice` used
  `OWNER=module.auth`; `test-evidence-audit` used that owner plus the retained
  auth-slice evidence manifest. `make lint` and `make test-fast` were invoked
  four times each. `make test` completed once and its next invocation was
  cancelled during collection. `ci` and `release-check` were not invoked from
  this source snapshot.
- Worktree and sequencing: T-001 remains the sole `IN_PROGRESS` task; T-008 is
  not unlocked. Committing this ledger advances the tracker branch but does not
  mutate retained contexts. To reuse the accepted roots, resume the missing
  `test`, `ci`, and `release-check` windows from a separate clean worktree at
  exact commit `334465751e05c9d50e00416fc4a071d579bfd719`. Mixing roots from a
  later source snapshot is forbidden. After all three missing windows pass,
  verify every context, build the exact evidence manifest, generate the
  baseline only through
  `make harness-public-target-duration-baselines EVIDENCE_ROOTS_FILE=...`, and
  update this tracker before beginning T-008.

### 2026-07-21 — T-001 retained v2 reference qualification complete

- Source: implementation based on clean `main` commit
  `fd8630ac`; no new baseline target observations were executed.
- Retained audit: the frozen UTC interval `2026-07-21T10:04:31Z` through
  `2026-07-21T22:04:31Z` contains 368 execution contexts. Of 291 nominally
  eligible contexts, 218 satisfy the strict complete-root contract and 73 are
  retained with strict rejection reasons. The isolated
  `retained_v1_reference_migration` reader recovered eligible historical roots
  lacking only the later invocation marker by validating terminal native
  summaries and exact timing boundaries; candidate evidence cannot use this
  path.
- Completed: T-001. The performance contract is v2 and target/provider scoped.
  Twenty-one deduplicated windows bind all 48 required public targets, the
  internal release-browser profile, and the synthetic backend finalizer. Each
  binding retains one discarded warm-up and exactly two measured roots. The
  backend finalizer is derived from the native `report_collation` interval
  union. The checked-in v2 baseline has 50 rows and a 48-public-target
  reference portfolio total of `2721126 ms`.
- Contract and compatibility: performance command IDs advanced to v2; v1
  manifests are rejected in normal operation. Original retained roots remain
  immutable and the migration baseline remains implementation-support, not
  Core 05 publication evidence. Public Make names, application OTel behavior,
  product behavior, domain vocabulary, and ordinary-run zero-egress behavior
  are unchanged.
- Verification: the sole writer passed at
  `.cartulary/test-results/20260721T224115Z-p3902667`; owner-manifest refresh and
  ordinary generation passed at `20260721T224243Z-p3904051` and
  `20260721T224259Z-p3905640`; `generate-drift` passed at
  `20260721T224419Z-p3907674`; generated-artifact policy passed at
  `20260721T224419Z-p3907695`; JSON shape passed at
  `20260721T224456Z-p3913421`; and `harness-contract` passed at
  `20260721T224549Z-p3915005`. The first JSON-shape attempt correctly exposed
  the stale warm-policy enum and the first harness-contract attempt correctly
  exposed the intentional command-interface digest change; both owner fixtures
  were corrected before the passing reruns.
- Generated status: owner documents, extension-owner dependencies, generated
  contracts, task-surface projection, and topology render index agree with
  their owners. The v2 baseline was written only by its public Make target.
- Active: T-008 is the only `IN_PROGRESS` task. Its checkpoint must record the
  exact 30-process/255-test proof before T-009 begins.

### 2026-07-21 — T-008 backend capture consolidation complete

- Source: T-001 checkpoint `ecd2ef40`; worktree implementation was validated
  before this mandatory T-008 checkpoint.
- Completed: T-008. Exact-symbol rows are partitioned by normalized package
  selection, sorted runtime binaries, runtime/resource/fixture profiles,
  complete fixture policy and budget, isolation policy, evidence class, and
  coverage identity. Raw selectors always form one isolated physical group.
  The current catalog deterministically resolves 34 logical families to 29
  exact groups plus one raw group. The host-derived pool uses six workers on
  this 24-CPU host and owns child `GOMAXPROCS=4`; the obsolete fixed
  `BACKEND_INTEGRATION_SHARD_JOBS` surface was removed rather than retained as
  an ignored compatibility burden.
- Projection integrity: physical capture identity is separate from logical
  evidence identity. Families spanning multiple compatible package groups are
  merged once into their stable family report before emission; shared physical
  execution is accounted once as actual and subsequent family ownership as
  reuse. Row order, family step identity, target summary schema, cleanup, and
  authored failure precedence remain stable.
- Verification: `harness-contract` passed at
  `.cartulary/test-results/20260721T230009Z-p3937762`; `backend-unit` passed at
  `.cartulary/test-results/20260721T230040Z-p3938914` with all 255 tests,
  failed/missing `0`, 30 completed physical captures, and accounting
  `actual=30`, `reused=4`; `generate-drift` and `json-shape-check` passed at
  `20260721T230204Z-p3941056` and `20260721T230204Z-p3941058`.
- Retained failure: `.cartulary/test-results/20260721T225357Z-p3919588` failed
  family closure because the first implementation emitted package fragments
  independently. It is excluded and retained as the regression case that led
  to family-owned merged projection. The corrected roots close every row and
  exact symbol once.
- Performance disposition: the functional slice is complete; its 32.744 s
  dirty-worktree wall is diagnostic only. No threshold was evaluated or
  relaxed. The qualified backend and portfolio comparisons remain exclusively
  T-012-owned.
- Active: T-009 is the only `IN_PROGRESS` task. It will eliminate repeated
  physical report parsing and parallelize immutable family requests before any
  T-010 topology work begins.

### 2026-07-21 — T-009 backend report finalization complete

- Source: T-008 checkpoint `1c97ec89`; worktree implementation was validated
  before this mandatory T-009 checkpoint.
- Completed: T-009. Every completed physical Go report is read and JSON-
  validated exactly once into a frozen record. Frozen family requests then
  merge and emit independently through the same host-derived pool as capture;
  on this host that is six workers. All workers settle before target summary
  emission, diagnostics, and authored-order primary-failure selection.
  Successful family evidence is preserved when a sibling parser or emitter
  fails. The undeclared `CARTULARY_GO_TARGET_FINALIZER_EMIT_JOBS` environment
  control was removed with no alias or ignored compatibility path.
- Artifact-completion correction: failed retained root
  `.cartulary/test-results/20260721T231625Z-p4003786` exposed a real race once
  parsing moved immediately behind capture: the child process had closed while
  its artifact write streams were still draining, producing an incomplete
  family projection. Capture completion now waits for both stdout and stderr
  write-stream `finish` events before publishing the physical report's
  `complete` marker. Consecutive corrected `backend-unit` roots
  `20260721T231719Z-p4006027` and `20260721T231740Z-p4007834` each passed all
  255 tests across 30 physical packages with zero failed or missing tests.
- Determinism and failure coverage: focused fixtures prove bounded concurrency,
  stable result indexing under simultaneous worker failures, partial-success
  preservation, frozen parsed records, rejection of malformed physical JSON,
  and byte-equivalent repeated projections after the source report is made
  unreadable as valid JSON. Existing runner fixtures continue to cover missing,
  duplicate, unexpected, partial, crashed, and missing-metadata evidence and
  stable failure classification.
- Verification: `harness-contract` passed at
  `.cartulary/test-results/20260721T231806Z-p4009638`; final `backend-unit`
  parity passed at `20260721T231740Z-p4007834` in `9.493 s`; `test-fast`
  passed at `20260721T231831Z-p4010588` with 651 tests and both aggregate work
  units complete; and `lint-biome` passed at
  `.cartulary/test-results/20260721T232100Z-p4051497`. Native backend
  finalization is now fully attributed through the `report_collation` interval
  union. These dirty-worktree roots are functional evidence only; no
  performance threshold was evaluated or relaxed.
- Performance disposition: the backend total and finalizer improvement gates
  remain exclusively T-012-owned and require strict evidence from the later
  frozen candidate commit.
- Active: T-010 is the only `IN_PROGRESS` task. Topology-owned adaptive
  sequence resources must be complete and checkpointed before browser capacity
  changes begin.

### 2026-07-21 — T-010 adaptive aggregate DAGs complete

- Source: T-009 checkpoint `dfebc233`; worktree implementation was validated
  before this mandatory T-010 checkpoint.
- Completed: T-010. The shared resource registry now admits the `sequence`
  scheduler through the topology-owned `sequence_adaptive` profile. On this
  24-CPU host it resolves `host_cpu=20`, `host_io=24`, and `process=8`.
  Closed work profiles own CPU, I/O, process, and nested Make concurrency;
  every sequence step claims one process slot. Outer sequence capacity ignores
  inherited check overrides, while registry-validated forwarding passes exact
  claimed budgets to nested check and service-backed schedulers.
- Topology: `lint` establishes Node, frontend-install, and shell readiness once,
  then admits all eight independent steps. `ci` starts only `check`, then admits
  harness contract, security audit, deployable shape, and duration drift
  independently. `release-check` starts only `check`; its harness, security,
  license/SBOM, build/deployable, SeaweedFS, and browser branches use the exact
  dependency graph in TH-HARNESS-REQ-374. Authored summary and failure order
  remain stable regardless of completion order.
- Retained functional evidence: fast sequence smoke passed at
  `.cartulary/test-results/20260721T234510Z-p4067014`; extended
  `harness-contract` passed at `20260721T234522Z-p4068163`; `lint` passed at
  `20260721T234606Z-p4069227` in 21.633 s with all eight units running at peak
  and peak claims CPU/I/O/process `20/11/8`; `ci` passed at
  `20260721T235101Z-p22359` in 154.163 s. Its nested `check` received exactly
  CPU/I/O `20/24`, and all four post-check branches ran concurrently.
  `generate`, `json-shape-check`, and `generate-drift` passed at
  `20260721T234257Z-p4064437`, `20260721T234308Z-p4066013`, and
  `20260721T235352Z-p137567`.
- Retained host-state failures: `20260721T234652Z-p4085995` was rejected because
  the prior baseline workspace still owned host port 5432. That exact stale
  Compose project was stopped without deleting volumes. The initially failed
  current-repo container then retained invalid unpublished-port metadata, so
  `20260721T234901Z-p4145697` was rejected after a connection refusal. The
  current Compose containers/network were recreated without deleting volumes;
  the unchanged next run passed. Neither root is performance evidence.
- Performance disposition: these dirty-worktree roots prove behavior and
  resource ownership only. The exact sequence transition is now validated from
  its normalized policy projection; quantitative acceptance remains T-012.
- Active: T-011 is the only `IN_PROGRESS` task. Release browser schedule
  capacity and five-session ownership must be checkpointed before candidate
  collection begins.

### 2026-07-21 — T-011 release browser capacity complete

- Source: T-010 checkpoint `d31f1f0d`; worktree implementation was validated
  before this mandatory T-011 checkpoint.
- Completed: T-011. `release-browser-readiness` now owns `browser_stack=2`
  directly in its authored service-backed schedule. Its exact five sessions are
  support default, visual default, visual Network Flow claimed, accessibility
  default, and accessibility Network Flow claimed. Every session has an
  immutable isolation reason and runtime profile. Visual and accessibility
  stage capacities remain one, so same-stage sessions never overlap. The outer
  `release-check` sequence forwards only its 2/4 CPU/I/O child budget and does
  not own or override browser-stack capacity.
- Contract enforcement: service schedule generation rejects any release
  schedule that changes stack or stage capacity, introduces non-browser work,
  omits an isolation reason, gives one session inconsistent identity, or
  differs from the exact five-session inventory. Strict performance comparison
  now validates the normalized release-browser session projection rather than
  accepting only a changed digest.
- Retained verification: `harness-contract` passed at
  `.cartulary/test-results/20260722T000055Z-p148719`; the full extended Harness
  matrix passed at `20260722T000635Z-p213252`; the real release browser schedule
  passed at `20260722T000949Z-p267160` in 134.034 s with five session starts,
  peak browser stacks `2`, visual/accessibility lane peaks `1`, peak running
  work `2`, no skipped work or finalizer failure, and five fixture-reclaimed
  events. The isolated public `browser-e2e-visual` leaf passed at
  `20260722T001838Z-p317683` in 121.017 s with browser-stack capacity one and no
  visual drift. `json-shape-check`, `generate-drift`, and generated-artifact
  policy passed at `20260722T000428Z-p179420`, `20260722T002110Z-p336702`, and
  `20260722T002110Z-p336694`.
- Retained corrections: extended roots `20260722T000154Z-p150129`,
  `20260722T000259Z-p155261`, and `20260722T000441Z-p180032` exposed stale
  synthetic-resource, scheduler-registry, and process-source fixture
  assumptions from T-010; each was corrected before the passing matrix. Direct
  visual root `20260722T001311Z-p280154` was rejected because an abandoned
  reference-baseline `s3corsproxy` process still owned port 8333. The exact
  stale process was terminated; the unchanged next visual run passed.
- Performance disposition: all roots above are dirty-worktree functional
  evidence. No T-012 threshold was evaluated or relaxed.
- Active: T-012 is the only `IN_PROGRESS` task. Candidate evidence must now be
  collected from one clean frozen commit and may not use retained-v1 migration.

### 2026-07-22 — T-012 warm-up rejects backend finalizer and reopens T-009

- Source: clean frozen T-011 checkpoint `5602039c19f235b54b19ad08d1d1bdebb8fca936`.
  The first strict candidate `release-check` warm-up passed at
  `.cartulary/test-results/20260722T002542Z-p344761` in `240.606 s`; its
  execution context retained the expected commit, clean source snapshot,
  24-CPU capacity, and no contamination reason.
- Positive diagnostic results: the warm root recorded `release-check` at
  `240.914 s`, nested `check` at `116.517 s`, backend-unit at `20.697 s`, and
  release-browser readiness at `98.488 s`. It retained exact-once target spans
  for the expected release, browser, backend, frontend, security, and
  SeaweedFS branches.
- Rejection: the authoritative backend-unit native `report_collation` interval
  union was `28.526 s`, above the required T-001 baseline improvement limit of
  `5.991 s` (`6.991 s` reference median less the required `1.000 s`
  improvement). No measured observation was started and the warm root cannot
  enter a later candidate window.
- Diagnosis: all 30 physical reports parse once in milliseconds, but 34
  logical family projections still pay cold output-helper startup in six
  host-pool waves. Their interval envelope is approximately `15.34 s`.
  Initial target-summary collation then spends `13.177 s` generating owner
  evidence because the catalog/accounting selection is reloaded twice for
  each owner. This is repeated initialization and lookup work, not retained
  product or evidence behavior.
- Disposition: T-009 is reopened as the sole `IN_PROGRESS` item and T-012 is
  restored to `TODO`. Replace per-family cold helpers with bounded reusable
  output workers, reuse one immutable catalog/accounting context during owner
  finalization, preserve stable authored failure order and partial-success
  evidence, and re-run focused finalizer plus backend-unit validation. Only
  after the strict finalizer warm diagnostic clears the existing threshold may
  T-009 return to `DONE` and T-012 restart with a new clean commit.

### 2026-07-22 — T-009 reusable backend finalization complete

- Source: remediation implementation commit `995bfe48`, following the
  mandatory reopen checkpoint `f809d392`.
- Completed: immutable family-projection requests are partitioned across the
  same host-derived six-worker pool as backend capture, but each output worker
  now initializes once and processes several requests sequentially. The
  34-family backend-unit inventory resolves to balanced worker batches of
  `6/6/6/6/5/5`. Report collation remains separate from emission, all worker
  results settle into authored family order, and custom test helpers retain the
  existing bounded fallback path.
- Owner finalization: one validated catalog/accounting context is loaded per
  target and reused across every owner partition. The context retains exact
  command routing, catalog rows, profile digests, and source identity; it does
  not cache across target invocations or weaken owner-shard validation.
- Failure and policy proof: focused fixtures retain a failed worker request and
  still complete its successful sibling, reject invalid worker counts, prove
  stable 34-request partitioning, and reject forged accounting contexts. The
  normalized backend transition now records
  `emission=bounded_reusable_host_derived_pool` and
  `owner_accounting_context=once_per_target` in addition to exact physical
  parse-once identity.
- Functional verification: `harness-contract` passed at
  `.cartulary/test-results/20260722T004511Z-p519347`; consecutive dirty-worktree
  `backend-unit` runs `20260722T003843Z-p511250` and
  `20260722T004540Z-p520390` each retained 255 tests, 26 owner shards, and
  finalizer unions of `3.177 s` and `3.161 s`; `test-fast` passed at
  `20260722T004610Z-p522630` with 651 tests and backend-unit/store/integration
  finalizer unions of `3.868 s`, `3.304 s`, and `2.879 s`. `generate`,
  `generate-drift`, JSON shape, generated-artifact policy, Biome, and Markdown
  validation passed at `20260722T004453Z-p517660`,
  `20260722T004822Z-p563573`, `20260722T004838Z-p568692`,
  `20260722T004838Z-p568698`, `20260722T004847Z-p569300`, and
  `20260722T004847Z-p569306`.
- Strict quantitative proof: clean-source root
  `.cartulary/test-results/20260722T004956Z-p571520` retained execution-context
  v2 for commit `995bfe48d90a2e7dbabedb361df8afb50e2860aa`, the public invocation
  boundary, no contamination, all 255 tests, and all 26 owner shards. Backend
  wall was `4.645 s` and the authoritative native `report_collation` union was
  `3.225 s`, below the unchanged `5.991 s` improvement limit.
- Rejected generation attempts `20260722T004240Z-p515040`,
  `20260722T004345Z-p516091`, and `20260722T004425Z-p516901` exposed the
  expected owner-document then canonical dependency-digest refresh sequence;
  none is performance evidence.
- Active: T-012 is again the sole `IN_PROGRESS` task. Its prior release warm-up
  remains rejected. Candidate evidence must restart from the new clean tracker
  checkpoint and cannot reuse any root from commit `995bfe48`.

### 2026-07-22 — T-012 strict qualification reopens T-001 invocation identity

- Source: clean frozen candidate checkpoint
  `d829ccee3a2b7f337e679fba29a3f36b1c83deb8`. Candidate collection produced
  one successful warm-up plus exactly two measured roots for each of the 21
  reference-mirrored providers. All 63 provider contexts are v2, clean,
  terminal-successful, warm-eligible, uncontaminated, and share the frozen
  commit and source snapshot.
- Rejection: the first strict comparison stopped before duration evaluation at
  `frontend-fallow-static` because its direct roots do not retain the public
  invocation boundary. The same structural defect affects
  `go-test-duration-baseline-coverage`, `test-slice`,
  `service-backed-test-slice`, and `test-evidence-audit`. The roots remain
  diagnostic-only and will not be rerun unchanged.
- Diagnosis: each public preflight wrote a valid invocation marker to a sibling
  run ID while its generated node-tool recipe explicitly re-expanded the
  recursive Make default `CARTULARY_TEST_RUN_ID` for the actual summary root.
  For example, audit summary root `20260722T022824Z-p2225741` has its marker in
  sibling root `20260722T022824Z-p2225743`. This splits one public invocation
  across two retained roots and violates strict-current evidence identity.
- Disposition: T-001 is reopened as the sole `IN_PROGRESS` item and T-012
  returns to `TODO`. Freeze `CARTULARY_TEST_RUN_ID` once at target scope for
  every public artifact-emitting generated recipe, add generator and contract
  coverage that proves preflight and summary identity agree, regenerate owned
  outputs, and validate focused direct roots. Only the five affected provider
  windows require recollection after the corrected clean checkpoint; every
  other candidate window remains rejected for the next source snapshot because
  strict candidate rows must share one frozen commit.

### 2026-07-22 — T-001 public invocation identity complete

- Source: implementation commit `4d8b01ff`. The task-surface generator now
  emits one immediate target-scoped `CARTULARY_TEST_RUN_ID` export for every
  public artifact-emitting recipe unless the owner already declares the same
  freeze. All 78 applicable public recipes therefore reuse one run identity
  across preflight, prerequisites, child work, summaries, cleanup, and
  observability finalization.
- Contract: TH-HARNESS-REQ-283 now requires the invocation marker and terminal
  summary to occupy the same frozen run root and classifies sibling generated
  IDs as artifact-incomplete. TH-HARNESS-AC-015 covers generated default
  identities for public node-tool and owner-slice targets. The renderer fixture
  checks every public artifact target, so a new recipe cannot omit the freeze.
- Clean retained proof at `4d8b01ff`: `frontend-fallow-static`
  (`20260722T024252Z-p2244874`),
  `go-test-duration-baseline-coverage`
  (`20260722T024304Z-p2245596`), one-row `test-slice`
  (`20260722T024314Z-p2245824`), and one-row
  `service-backed-test-slice` (`20260722T024322Z-p2246184`) all passed with
  clean source state, no contamination, and
  `invocation_boundary_retained=true` in the exact terminal root. The audit
  wrapper also retained the exact boundary during its expected rejection of
  stale pre-change full-owner evidence; successful canonical audit proof is
  coupled to the new full-owner T-012 window.
- Validation: `make generate` passed at
  `.cartulary/test-results/20260722T023834Z-p2229772`; `harness-contract`
  passed at `20260722T023857Z-p2231558`; `generate-drift`, JSON shape, and
  generated-artifact policy passed at `20260722T024120Z-p2235179`,
  `20260722T024120Z-p2235215`, and `20260722T024120Z-p2235186`; Biome,
  ShellCheck, and Markdown lint passed at `20260722T024141Z-p2241436`,
  `20260722T024141Z-p2241451`, and `20260722T024141Z-p2241461`.
- Active: T-012 is again the sole `IN_PROGRESS` item. Because strict candidate
  rows must share one clean frozen commit and snapshot, all provider windows
  must be recollected after this mandatory tracker checkpoint; none of the
  prior `d829ccee` candidate roots can enter the accepted comparison.

### 2026-07-22 — T-012 browser lifecycle failure reopens T-011

- Source: clean frozen candidate checkpoint
  `8720868b7663704420462009ace192125e790348`. Candidate collection retained
  successful strict-current warm-up and two-sample windows for release-check,
  test-fast, ci, lint, agent-finalize, benchmark-claim-check, duration-baseline
  coverage, JSON shape, migration drift, toolchain drift, frontend fallow,
  gosec, govulncheck, both standup smoke targets, SeaweedFS release evidence,
  and browser-e2e. These roots remain diagnostic-only after the next source
  change and cannot be mixed into the replacement candidate snapshot.
- Rejection: full `test` warm-up root
  `.cartulary/test-results/20260722T034226Z-p3585244` passed 651 tests, but the
  first measured attempt
  `.cartulary/test-results/20260722T034608Z-p3655308` failed the
  `stateful-network-flow-claimed-network-flow` group while starting the
  recovery import after a table lifecycle deletion. The failed root is
  retained with reason `failed_execution`; no unchanged second measurement
  was started.
- Diagnosis: the trace proves upload, job, unit, and preview requests completed
  in approximately 23--150 ms, excluding backend starvation and a timeout
  remedy. The test used the WebSocket `extension_resource_changed` event as the
  deletion-completion boundary. That event can arrive just before the delete
  request's UI completion callback. The next import then reserves a generation
  which the late callback resets, leaving the intentionally stale result
  ignored while the view still reports `Discovering source columns.`
- Disposition: T-011 is reopened as the sole `IN_PROGRESS` item and T-012
  returns to `TODO`. Synchronize the browser lifecycle scenario on the UI's
  committed deletion acknowledgement after retaining the socket assertion,
  add focused regression coverage for the boundary, and rerun the owned
  stateful browser evidence. Product behavior, request timing, and test timeout
  budgets remain unchanged. Only after that proof may T-011 return to `DONE`
  and T-012 restart from a new clean checkpoint.

### 2026-07-22 — T-011 browser lifecycle synchronization complete

- Source: implementation commit `96c08d65`, following the mandatory reopen
  checkpoint `e52b501b`. The network-flow lifecycle scenario now retains its
  WebSocket removal assertion and also waits for the visible
  `lifecycle-source was deleted.` acknowledgement before starting recovery.
  This is the UI commit boundary after `clearResources`; it prevents a stale
  deletion callback from invalidating the next import generation without
  changing application behavior or increasing a timeout.
- Focused proof: `make browser-e2e-stateful` passed at
  `.cartulary/test-results/20260722T035444Z-p3722698` with both direct sessions
  complete and the claimed network-flow lifecycle scenario passing. Its
  scheduler completed in `175.30 s` with the direct target's authored
  single-stack isolation.
- Release proof: `make release-browser-readiness` passed at
  `.cartulary/test-results/20260722T035814Z-p3757406`; all 14 scheduler work
  units completed, `browser_stack` capacity was two, observed peak usage was
  exactly two, the visual/accessibility stage capacities remained one, no work
  unit or finalizer failed, and scheduler duration was `134.161 s`.
- Visual and source validation: `make browser-e2e-visual` passed at
  `.cartulary/test-results/20260722T040037Z-p3770742` with both sessions and no
  golden drift. `make lint-biome` passed at
  `.cartulary/test-results/20260722T040305Z-p3790056`.
- Active: T-012 is again the sole `IN_PROGRESS` item. Strict candidate roots
  must now share the next clean tracker checkpoint; all roots from
  `8720868b`, including successful windows, are retained rejects after this
  source change and cannot enter the accepted comparison.

### 2026-07-22 — T-012 strict comparison reopens T-001 policy digest contract

- Source: clean frozen candidate checkpoint
  `706aa7ed62832ce77871577921c9f841bd884a87`. Collection produced 21
  provider windows and 63 selected strict-current roots covering all 50
  target bindings. Every selected root passed, retained a clean source state,
  reported zero retries and no contamination, preserved its public invocation
  boundary, and completed observability indexing.
- Retained rejects: test-fast root
  `.cartulary/test-results/20260722T042926Z-p302152` and ci root
  `.cartulary/test-results/20260722T043420Z-p382608` are rejected as
  `retry_observed`; test-fast root
  `.cartulary/test-results/20260722T054231Z-p1282645` is rejected as
  `artifact_incomplete`; the prior failed full-test root remains rejected as
  `failed_execution`; three overlapping standup roots remain rejected as
  `external_activity`. The retry-free replacement windows passed strict
  retained qualification.
- Rejection: the read-only comparison reached the first target and failed with
  `agent-finalize retained execution-policy projection digest mismatch` before
  evaluating any duration threshold. The retained producer hashes formatted,
  insertion-ordered JSON while the v2 performance validator hashes recursively
  sorted semantic JSON. The same policy value therefore has two deterministic
  digests. This is a contract implementation defect, not a candidate timing
  regression, and no threshold was relaxed.
- Disposition: T-001 is reopened as the sole `IN_PROGRESS` item and T-012
  returns to `TODO`. Define one I-JSON-safe, recursively key-sorted semantic
  encoding for execution-policy projections, use the shared canonicalizer in
  both producer and validator, and prove key-order independence plus
  producer/validator agreement. Retained reference roots remain immutable and
  readable; all `706aa7ed` candidate roots become diagnostic-only after the
  correction. T-012 must recollect every provider window from the next clean
  tracker checkpoint.

### 2026-07-22 — T-001 semantic execution-policy digest correction complete

- Source: implementation commit `45639626`, following mandatory reopen
  checkpoint `b27b43a6`. Section 10.5 now binds normalized policy projections
  to the Section 3.6 recursively key-sorted, I-JSON-safe semantic encoding and
  its bare SHA-256 value. The retained context producer and v2 validator call
  the same shared digest implementation. The performance implementation also
  removed its private JSON canonicalizer and uses the shared semantic encoding
  for policy equality and window signatures. Retained artifact file formatting
  and all unrelated legacy digests remain unchanged.
- Regression proof: harness fixtures assert that top-level and target-local
  measurement-policy digests equal the validator digest and remain equal after
  object-key reordering. The semantic JSON contract separately proves that its
  prefixed and bare forms share one byte encoding.
- Generated owners: `make generate` passed at
  `.cartulary/test-results/20260722T060818Z-p1771304` and refreshed the testing
  harness owner-document digest, extension dependency projection, generated Go
  contracts, and execution-topology render index through their owner inputs.
- Validation: `make harness-contract` passed at
  `.cartulary/test-results/20260722T060831Z-p1773514`; JSON shape,
  generated-artifact policy, and generation drift passed at
  `20260722T060831Z-p1773133`, `20260722T060831Z-p1773138`, and
  `20260722T060831Z-p1773154`. Script and Markdown lint passed at
  `20260722T060831Z-p1773612` and `20260722T060831Z-p1773653`.
- Active: T-012 is again the sole `IN_PROGRESS` item. Every candidate provider
  window must be collected from the next clean tracker checkpoint; none of the
  otherwise-qualified `706aa7ed` roots can enter the accepted candidate
  snapshot after this source change.

### 2026-07-22 — T-012 strict comparison reopens T-001 digest-only v1 bridge

- Source: clean frozen candidate checkpoint
  `206c343ce849e7b2a69cdc5ea9d408e15760136d`. Collection produced 21 primary
  provider windows and 63 selected strict-current roots, all passed, clean,
  retry-free, uncontaminated, invocation-boundary complete, observability-index
  complete, and bound to one source snapshot. Two additional direct windows
  for `generate-drift` and `generated-artifact-policy-check` were retained
  after their `ci` trace lacked an exact target envelope; this follows the
  direct-provider fallback instead of weakening exact-once aggregate proof.
- First comparison rejection: `generate-drift aggregate timing is not
  exact-once`. Diagnosis proved one unique scheduler-work interval but no
  target span for each wrapper-style target. Direct warm-up plus two measured
  windows passed for both commands and preserve the reference timing identity.
- Second comparison rejection: `agent-finalize has an undeclared
  execution-policy change`. Every retained baseline row is a deliberately
  migrated v1 digest-only policy record. Its legacy formatted digest cannot
  equal the corrected strict candidate's semantic digest even when the policy
  projection is unchanged. Duration thresholds were not reached or relaxed.
- Disposition: T-001 is reopened as the sole `IN_PROGRESS` item and T-012
  returns to `TODO`. The migration-only comparator must recompute the historical
  formatted digest over the strict candidate projection to prove an unchanged
  legacy policy. That bridge is forbidden for candidate qualification, v2
  candidate digests, and publication evidence; declared policy transitions
  continue to require their exact normalized projections. All `206c343c`
  candidate roots become diagnostic-only after the correction, and T-012 must
  recollect from the next clean checkpoint.

### 2026-07-22 — T-001 digest-only v1 policy bridge complete

- Source: implementation commit `67afce9c`, following mandatory reopen
  checkpoint `2e892441`. An unchanged migrated digest-only row is now compared
  by recomputing its historical formatted-JSON SHA-256 over the strict
  candidate's normalized projection. The comparator first requires the
  migrated wrapper digest and row digest to agree. The helper is private to
  performance comparison and is selected only for
  `retained_v1_reference_migration`; v2 contexts and strict candidates continue
  to require the shared I-JSON-safe semantic digest.
- Transition isolation: migrated rows with declared backend, sequence, or
  browser transitions continue through the exact normalized transition
  validators and never through the digest-only unchanged-policy bridge.
  Regression fixtures cover an unchanged strict projection, an undeclared
  projection change, a tampered migrated wrapper, and the serial-to-DAG
  transition alongside ordinary performance gates.
- Generated owners: final `make generate` passed at
  `.cartulary/test-results/20260722T075424Z-p3501177`, refreshing the testing
  harness owner digest and generated Go contract projection through their owner
  inputs.
- Validation: `make harness-contract` passed at
  `.cartulary/test-results/20260722T075440Z-p3502914`; JSON shape,
  generated-artifact policy, and generation drift passed at
  `20260722T075522Z-p3504207`, `20260722T075522Z-p3504213`, and
  `20260722T075522Z-p3504194`. Script and Markdown lint also passed in that
  final validation group.
- Active: T-012 is again the sole `IN_PROGRESS` item. Because the candidate
  snapshot must be the final clean comparison implementation, all 69 selected
  roots from `206c343c`, including the two direct fallback windows, remain
  diagnostic-only. Recollect every provider window after the next clean
  tracker checkpoint.

### 2026-07-22 — T-012 release-check window reopens T-010 teardown safety

- Source: clean frozen candidate checkpoint
  `e1ad71ff236ecdf1b460783ce12a76924da8cc9a`. The first release-check
  warm-up passed at `.cartulary/test-results/20260722T084305Z-p316168` in
  `220.700 s`. The first measured observation failed at
  `.cartulary/test-results/20260722T084648Z-p467068` after `113.122 s`; it is
  retained as `failed_execution` and cannot enter a candidate window.
- Failure: the nested `check` backend-integration row panicked with
  `send on closed channel` while `Hub.PublishJobProgress` delivered the
  collaboration presence/replay integration scenario. A publisher currently
  snapshots subscriber channels under the hub mutex and sends after unlocking,
  while unsubscribe removes and closes a channel after unlocking. Concurrent
  teardown can therefore close a snapshotted channel before its send.
- Ownership finding: production websocket consumers terminate through request
  context and call unsubscribe only as deferred registry cleanup. Tests likewise
  use unsubscribe as cleanup and do not require it to signal completion by
  closing the data channel. Closing a sender-owned data channel from receiver
  cleanup is both unnecessary and unsafe; repeated unsubscribe also currently
  panics.
- Disposition: T-010 is reopened as the sole `IN_PROGRESS` item and T-012
  returns to `TODO`. Make subscription teardown a non-closing, idempotent
  registry removal, add concurrent publish/unsubscribe regression coverage,
  and revalidate the exact failed row plus aggregate concurrency. All otherwise
  successful roots from `e1ad71ff` become diagnostic-only after the correction;
  T-012 must recollect every provider window from the next clean checkpoint.

### 2026-07-22 — T-010 websocket teardown correction complete

- Source: implementation commit `5a2865d2`, following mandatory reopen
  checkpoint `a07f5bb6`. `SubscribeIncident` and `SubscribeRecordChanges` now
  treat unsubscribe as idempotent registry removal and intentionally leave the
  sender-owned data channel open. This matches the hub's existing revocation
  subscription lifecycle and removes the close-versus-snapshotted-send race
  without serializing publication behind consumer teardown.
- Regression proof: the selected `platform.ws` transport test races repeated
  incident and record-change publication against repeated unsubscribe for 256
  fresh hubs, then proves detached channels receive no later events. Its exact
  catalog slice passed at
  `.cartulary/test-results/20260722T085316Z-p557866`. The exact collaboration
  integration row that previously panicked passed at
  `.cartulary/test-results/20260722T085329Z-p558329`.
- Backend and local-loop proof: `make backend-unit` passed 255 tests at
  `.cartulary/test-results/20260722T085351Z-p560563`; `make test-fast` passed
  651 tests at `.cartulary/test-results/20260722T085414Z-p566522`.
- Aggregate proof: `make release-check` passed all 12 work units and 510 tests
  at `.cartulary/test-results/20260722T085626Z-p611037` with zero failed or
  skipped work units and zero finalizer failures. Its scheduler completed in
  `229.970 s`, enforced capacities `host_cpu=20`, `host_io=24`, `process=8`,
  observed the full CPU and I/O capacities, and admitted at most six concurrent
  work units. `make lint` passed all eight steps at
  `.cartulary/test-results/20260722T090103Z-p770392`; `make ci` passed all five
  work units and 510 tests at
  `.cartulary/test-results/20260722T090124Z-p776174`.
- Active: T-010 is closed and T-012 is again the sole `IN_PROGRESS` item. The
  dirty-tree validation roots above are functional evidence only. Candidate
  performance collection must use the next clean tracker checkpoint, and no
  root from the superseded `e1ad71ff` snapshot may qualify.

### 2026-07-22 — T-012 release window reopens T-010 browser admission

- Source: clean frozen candidate checkpoint
  `786e0c9e4db4888935cb981fef7ee66cb6234ea1`. Collection produced 69
  candidate contexts for all 23 provider windows: 68 passed and one failed.
  The first 22 provider windows are complete, and the release-check warm-up
  plus first measured observation passed at
  `.cartulary/test-results/20260722T103039Z-p2134645` and
  `.cartulary/test-results/20260722T103443Z-p2285233` in `224.938 s` and
  `221.862 s`. The second measured release observation failed at
  `.cartulary/test-results/20260722T103833Z-p2435294` and is retained as
  `failed_execution`; it cannot enter a candidate window and was not rerun.
- Failure evidence: four independent `browser-e2e-webserver-backed` groups
  timed out on workbook visibility, query observation, or save completion.
  Their 19 sibling browser groups began together inside the shared default
  browser session. The check scheduler recorded `max_running_work_units=19`
  and full `host_cpu=20`/`host_io=24` use, while process use peaked at only
  three of six slots because generated `browser_group` units claimed one CPU
  and one I/O token but no process token.
- Ownership finding: catalog browser rows use the closed `browser_exclusive`
  profile with `browser_stack=1` and `process=1`. The stage session correctly
  retains its stack/process ownership, but service-backed schedule generation
  drops the real Playwright-group process claim. The fixed non-sequence
  `host_process_slots=6` policy also prevents using a safe bounded fraction of
  the current 24-CPU host once group processes are accounted.
- Disposition: T-010 is reopened as the sole `IN_PROGRESS` item and T-012
  returns to `TODO`. Every scheduled Playwright browser group must claim one
  process slot in addition to its CPU/I/O claims. For `check`,
  `service_backed`, and `test_slice`, `host_process_slots` must resolve as
  `clamp(floor(available_parallelism/2),2,12)`, yielding 12 here; the sequence
  scheduler retains its separate `/3`, cap-eight policy. Add exact generation,
  capacity, admission, and sibling-drain fixtures, then revalidate browser,
  check, and release aggregates. All 69 `786e0c9e` roots become
  diagnostic-only after the correction, and T-012 must recollect every
  provider window from the next clean checkpoint.

### 2026-07-22 — T-010 browser-process admission correction complete

- Source: implementation commit `b6ad2de7`, following mandatory reopen
  checkpoint `3754f997`. Schedule generation now gives every normalized
  `browser_group` exactly one process claim in addition to its scheduler-family
  CPU and I/O claims. The retained browser session continues to own a distinct
  process slot for its backend/frontend stack. Manifest normalization rejects
  any browser group that lacks the exact process claim.
- Capacity: `check`, `service_backed`, and `test_slice` now resolve process
  capacity as `clamp(floor(available_parallelism/2),2,12)`, yielding 12 on this
  24-CPU host. Sequence scheduling retains its independent `/3`, cap-eight
  policy. Extended fixtures prove seven normalized browser groups each own one
  slot, the observed synthetic ceiling equals its six-slot limit, failure skips
  dependent same-session work while draining an isolated sibling, and all
  retained browser sessions finalize.
- Generated and contract validation: the first `make generate` at
  `.cartulary/test-results/20260722T104900Z-p2523777` failed closed on the
  intentionally stale testing-harness owner digest. After refreshing the
  authored owner/dependency chain, `make generate` passed at
  `.cartulary/test-results/20260722T105038Z-p2525875`. JSON shape,
  generated-artifact policy, generation drift, and harness contract passed at
  `20260722T105056Z-p2527818`, `20260722T105056Z-p2527756`,
  `20260722T105056Z-p2527781`, and `20260722T105056Z-p2528045`.
- Live proof: `make browser-e2e-webserver-backed` passed at
  `.cartulary/test-results/20260722T105129Z-p2534651`. `make check` passed all
  130 work units and 510 tests at
  `.cartulary/test-results/20260722T105658Z-p2561502`; it reached 16 concurrent
  work units and exactly 12 active process claims without exceeding the
  12-slot limit. Its 121.612-second wall is functional evidence, not the later
  warm acceptance window. `make lint` passed all eight steps at
  `.cartulary/test-results/20260722T105917Z-p2650473`, and `make ci` passed all
  five work units and 510 tests at
  `.cartulary/test-results/20260722T105935Z-p2656266`, with its nested check at
  117.131 seconds.
- Active: T-010 is closed and T-012 is the sole `IN_PROGRESS` item. All 69
  roots from `786e0c9e`, including 68 successful contexts and the retained
  release-check failure, are diagnostic-only. Recollect every provider window
  from the next clean tracker checkpoint; candidate evidence may not use the
  retained-v1 migration path.

### 2026-07-22 — T-012 `test` warm-up reopens T-010 retained-session admission

- Source: clean frozen candidate checkpoint
  `6a32e41e325a2287700184d78b7648081ea1ae3b`. Host `/tmp` maintenance removed
  an initial 42-context partial collection, so those roots were not retained or
  selected. Collection restarted from zero after capacity was restored.
- Successful retained collection: 48 strict-current contexts completed across
  16 provider windows. The `test-fast` measured pair passed 651 tests in
  `123.718 s` and `131.013 s`. The `ci` measured pair passed 510 tests in
  `139.893 s` and `141.628 s`, with nested `check` walls of `108.246 s` and
  `109.336 s`. Twelve short provider windows and both standup windows also
  completed successfully.
- Failure: the next provider's `test` warm-up at
  `.cartulary/test-results/20260722T141711Z-p3891735` passed its product tests
  but failed with a service-backed scheduler deadlock. A measurement-stage
  session retained `process=limit`, resolving to all 12 slots, while its
  normalized Playwright child correctly claimed the additional one process
  slot required by the preceding browser-admission correction. The pending
  child could therefore never become feasible.
- Ownership finding: a retained browser session owns one live application-stack
  process, never the worker capacity of its future children. Measurement
  exclusivity belongs on the measurement browser-group claim, which already
  owns full Go CPU, Go I/O, Postgres, and object-store capacity. Allowing a
  retained session to expand `process` or `browser_stack` to the scheduler limit
  conflates lifecycle retention with child admission and can self-deadlock any
  later explicitly accounted child.
- Disposition: T-010 is reopened as the sole `IN_PROGRESS` item and T-012
  returns to `TODO`. Normalize every retained browser stage session to exactly
  one `process` and one `browser_stack` claim while preserving its authored
  stage lane; keep measurement exclusivity on the child group. Add exact
  generation, normalization, feasibility, and failure-drain coverage, then
  revalidate `test`, `check`, and release aggregates. All 48 otherwise
  successful `6a32e41e` contexts and the failed `test` root become
  diagnostic-only after the correction; T-012 must recollect every provider
  window from the next clean checkpoint.

### 2026-07-22 — T-010 retained-browser admission and scheduler liveness complete

- Source: implementation commit `64dd1023`, following mandatory reopen
  checkpoint `2c8a5a28`. Every retained browser stage session now releases its
  startup envelope to exactly one browser-stack slot, one stage-lane slot, and
  one process slot. Measurement children preserve full CPU, I/O, Postgres, and
  object-store isolation while claiming their own single Playwright process.
- Adaptive admission: every ordinary scheduled browser group now claims two
  scheduler-family CPU tokens, one I/O token, and one process slot. Measurement
  groups retain limit-sized CPU/I/O claims plus one process slot. This preserves
  the 12-slot host process ceiling while bounding Chromium/Playwright pressure
  through the existing CPU registry instead of adding a private control.
- Liveness: a later limit-sized measurement session could reserve `process`
  while blocked by the first retained measurement lane, preventing the first
  session's feasible child from running and releasing that lane. FIFO resource
  reservation now has an idle-only liveness fallback: when no child is running,
  the earliest otherwise feasible ready unit may proceed. Reservation behavior
  is unchanged while any child is active. Contract fixtures prove both paths.
- Browser reliability: the account-settings page helper now distinguishes an
  already-open dialog from a failed tab/control action, so it cannot fall
  through and click navigation behind a modal. The exact workbook row that
  failed under the former full-load schedule passed at
  `.cartulary/test-results/20260722T150928Z-p210325`.
- Diagnostic failures retained: full `test` at
  `.cartulary/test-results/20260722T142558Z-p3971425` exposed the broad modal
  fallback; `.cartulary/test-results/20260722T143305Z-p4073334` exposed a
  different workbook assertion under the former one-CPU browser profile; and
  `.cartulary/test-results/20260722T151040Z-p226164` exposed FIFO reservation
  deadlock between the two measurement profiles. Product rows passed in the
  deadlock run; none of these roots qualifies for T-012.
- Focused validation: final extended harness smoke passed at
  `.cartulary/test-results/20260722T152800Z-p584361`. Harness contract passed on
  the final state and again inside release-check. Generation drift, JSON shape,
  and generated-artifact policy passed at
  `.cartulary/test-results/20260722T152730Z-p580020`,
  `.cartulary/test-results/20260722T152730Z-p580049`, and
  `.cartulary/test-results/20260722T152730Z-p580025`.
- Broad validation: `make test` passed all 651 tests at
  `.cartulary/test-results/20260722T151807Z-p302020`; its service scheduler
  completed 147/147 work units while reaching 18 concurrent units and the
  declared 16 CPU, 24 I/O, and 12 process ceilings. `make check`, `make lint`,
  and `make ci` passed at `20260722T152209Z-p374034`,
  `20260722T152426Z-p460599`, and `20260722T152449Z-p466433`; CI's nested check
  was 112.937 seconds. `make release-check` passed all 12 sequence units at
  `.cartulary/test-results/20260722T153047Z-p627614` in 286.918 seconds, with
  release browser readiness at 139.721 seconds.
- Generated state: owner specification digest, dependency declaration,
  generated contracts, execution-topology render index, and scheduler manifest
  were refreshed only through their authored owners and `make generate`. No
  lockfile, application telemetry, domain vocabulary, or visual golden changed.
- Active: T-010 is closed and T-012 is the sole `IN_PROGRESS` item. Every root
  collected before this tracker activation checkpoint, including the 48
  successful `6a32e41e` contexts, is diagnostic-only. Recollect one warm-up and
  exactly two measured roots for every provider window from this clean source;
  candidate evidence may not use retained-v1 migration exceptions.

### 2026-07-22 — T-012 portfolio comparison reopens T-008 worker-policy scope

- Source: clean frozen candidate checkpoint
  `9e2fc1aafaf09eb5c0e8c8ba4df3eba48f3da091`. All 23 provider windows were
  collected as one successful warm-up plus exactly two measured observations.
  Three retry-contaminated roots were rejected and retained with
  `retry_observed`: test-fast
  `.cartulary/test-results/20260722T161707Z-p1236031`, service-backed owner
  slice `.cartulary/test-results/20260722T160507Z-p1132214`, and release-check
  `.cartulary/test-results/20260722T170244Z-p2026907`. Complete replacement
  windows were collected rather than substituting individual samples.
- Strict comparison: every selected root passed schema, clean-state, canonical
  input, workload, host/capacity/toolchain, timing-source, and exact policy
  transition qualification. All four required-improvement gates passed:
  backend-unit `40324 -> 5497 ms`, backend finalization `6991 -> 3995 ms`,
  release browser readiness `154543 -> 136375.5 ms`, and release-check
  `456198.5 -> 261940.5 ms`. The 48-target portfolio improved
  `2721126 -> 2405789 ms` (`-315337 ms`).
- No-regression failures: backend-process `24182 > 23885 ms`,
  browser-e2e-webserver-backed `115932.5 > 104108.55 ms`, harness-contract
  `27047.5 > 25678.275 ms`, otel-conformance `4274 > 1626.5 ms`,
  SeaweedFS migration preservation `92735 > 88058.775 ms`, SeaweedFS release
  evidence `142819 > 138433.575 ms`, and operational recovery smoke
  `46804.5 > 39948.3 ms`. Thresholds remain unchanged and the failed window
  will not be rerun unchanged.
- Ownership finding: the backend pool's scheduler-owned `GOMAXPROCS=4` escaped
  the T-008 backend-unit consolidation boundary and was applied to process
  families. Retained command evidence shows the same app-server process group
  slowing from `14.746 s` without that override to `22.745 s` with it. T-008
  must scope the six-worker/four-thread capture policy to backend-unit while
  keeping the shared T-009 finalizer pool. The remaining failed rows are
  assigned to a subsequent T-010 resource/topology correction after T-008 is
  closed.
- Disposition: T-008 is reopened as the sole `IN_PROGRESS` item and T-012
  returns to `TODO`. After focused backend parity and process-latency proof,
  close T-008, reopen T-010, and correct retained browser resource
  double-accounting, latency-sensitive check admission, isolated SeaweedFS
  scenario scheduling, and operational recovery startup duplication. Any code
  change invalidates all `9e2fc1aa` candidate windows; T-012 must recollect from
  the next clean checkpoint.

### 2026-07-22 — T-008 backend worker-policy scope correction complete

- Source: reopen checkpoint `525856cb4a1b750af7dc64d27c0e2ce354bb6bba`.
  The implementation, owner specification, derived contract digest, and focused
  tests were updated before this mandatory checkpoint.
- Completed: the bounded capture pool now resolves through an explicit
  target-aware policy. `backend-unit` retains six workers with child
  `GOMAXPROCS=4` on this 24-CPU host; other backend capture targets retain
  `GOMAXPROCS=24`. Backend finalization continues to use the bounded T-009 pool
  without mutating capture-child parallelism.
- Evidence refinement: the failed T-012 aggregate samples themselves used
  scheduled process shards with `GOMAXPROCS=24`; the escaped four-thread policy
  was confined to the standalone sharded-runner path. The correction is still
  required because that path is part of the public backend target contract,
  but aggregate contention remains T-010-owned rather than being attributed to
  T-008.
- Verification: `make generate` passed at
  `.cartulary/test-results/20260722T180725Z-p3141781`; `make harness-contract`
  passed at `.cartulary/test-results/20260722T180737Z-p3143659`; `make
  backend-unit` passed all 255 tests at
  `.cartulary/test-results/20260722T180804Z-p3144729`; and `make
  backend-process` passed all 36 tests in `18.440 s` at
  `.cartulary/test-results/20260722T180821Z-p3147025`, with retained commands
  proving `GOMAXPROCS=24`.
- Drift and shape closure: `make generate-drift` and `make json-shape-check`
  passed at `.cartulary/test-results/20260722T180946Z-p3165420` and
  `.cartulary/test-results/20260722T181000Z-p3168510`.
- Active: T-008 is closed and T-010 is the sole `IN_PROGRESS` item. T-010 owns
  retained browser process-token double-accounting, latency-sensitive check
  admission, isolated SeaweedFS scenario concurrency, and operational recovery
  startup duplication. T-012 remains `TODO` until that checkpoint is complete.

### 2026-07-22 — T-010 resource and critical-path correction complete

- Source: T-008 checkpoint `fd0b5bcf`. The implementation, owner
  specification, policy projection, generated artifacts, and focused tests were
  completed before this mandatory T-010 checkpoint.
- Retained browser ownership: browser stage startup still claims one process
  slot while creating its backend/frontend stack, then releases that transient
  claim. The live session retains only its browser-stack and authored stage-lane
  claims. Browser children retain their own process claims, so session lifetime
  no longer double-counts child admission capacity. The normalized release
  browser policy projection records the retained and released resource classes
  and strict comparison validates them.
- Source-boundary work: OTel conformance now builds one filtered repository file
  index and reuses immutable file text across its checks. Authored-source scans
  explicitly exclude `.cache` and `.pnpm-store`; those directories contained
  219,486 and 18,174 non-source files on this host. The direct conformance body
  fell from `2.84 s` to `0.14 s` without changing the application/harness
  telemetry boundary, privacy rules, or ordinary-run zero-egress behavior.
- Release-support critical path: the independent SeaweedFS migration pass and
  mismatch scenarios now execute as parallel top-level tests with separate
  fixtures and cleanup identities. `make seaweedfs-migration-preservation`
  passed in `46.310 s` at
  `.cartulary/test-results/20260722T182101Z-p3179041`, down from the rejected
  `92.735 s` candidate row. Operational recovery now builds its uniquely tagged
  application image once and starts Compose from that completed image; its
  focused target passed in `25.750 s` at
  `.cartulary/test-results/20260722T182155Z-p3180277`, below the retained
  `39.948 s` no-regression limit.
- Aggregate proof: `make lint` passed all eight work units in `9.806 s` at
  `.cartulary/test-results/20260722T191455Z-p4116266`. `make ci` passed all five
  outer work units and 510 tests in `148.634 s` at
  `.cartulary/test-results/20260722T191121Z-p4000045`; its nested `check` target
  completed in `117.325 s`. The sequence scheduler resolved the required
  24-CPU host profile to 20 CPU tokens, 24 I/O tokens, and eight process slots.
  The final topology uses the plan's closed shared profiles; attempted private
  backend priority and full-CPU OTel profiles were removed after retained runs
  showed worse aggregate critical paths.
- Browser proof: `make release-browser-readiness` passed at
  `.cartulary/test-results/20260722T191523Z-p4134256` in `137.788 s`. Its
  scheduler admitted at most two concurrent stacks, retained one visual and one
  accessibility lane, and completed all five authored sessions with cleanup.
  Direct `make browser-e2e-visual` also passed at
  `.cartulary/test-results/20260722T191828Z-p4147780` in `129.21 s` with no
  visual drift.
- Contract and generated closure: final `make harness-contract` passed in
  `22.170 s` at `.cartulary/test-results/20260722T192139Z-p4171135`.
  `make generate` passed at
  `.cartulary/test-results/20260722T192241Z-p4172385`; final generation drift,
  JSON shape, and generated-artifact policy checks passed at
  `.cartulary/test-results/20260722T192320Z-p4174206`,
  `.cartulary/test-results/20260722T192335Z-p4177293`, and
  `.cartulary/test-results/20260722T192347Z-p4177822`. OTel conformance passed
  again at `.cartulary/test-results/20260722T192356Z-p4178159`.
- Diagnostic rejects: CI root
  `.cartulary/test-results/20260722T185527Z-p3644302` retained the rejected
  full-CPU OTel admission policy (`175.297 s`, nested check `143.390 s`). Root
  `.cartulary/test-results/20260722T190533Z-p3881013` retained a transient
  74.276-second service startup (`199.710 s`, nested check `168.205 s`). Both
  passed functionally but do not qualify performance acceptance. The clean
  shared-profile rerun restored service startup to `9.087 s` and is the
  controlling T-010 proof.
- Active: T-010 is closed and T-012 is the sole `IN_PROGRESS` item. All roots
  collected before this tracker activation checkpoint are diagnostic-only for
  candidate qualification. Freeze this checkpoint, then collect one successful
  warm-up and exactly two consecutive measured roots for each minimal provider
  window; candidate evidence may not use retained-v1 migration exceptions.

### 2026-07-22 — T-012 early CI window reopens T-010 prerequisite admission

- Source: clean frozen candidate checkpoint
  `867028dee64a451e2790844142776d0b5bf1b37f`. The CI warm-up and consecutive
  measured roots passed at
  `.cartulary/test-results/20260722T192640Z-p4184011`,
  `.cartulary/test-results/20260722T192905Z-p104778`, and
  `.cartulary/test-results/20260722T193132Z-p218409`. Their public walls were
  `142.066 s`, `144.711 s`, and `143.937 s`; measured nested-check work was
  `116.318 s` and `115.599 s`.
- Early rejection: the measured target-local medians for harness contract
  (`28.296 s` against `25.678 s`), OTel conformance (`3.115 s` against
  `1.627 s`), and gosec audit (`6.024 s` against `4.705 s`) exceed their
  no-regression limits. CI and check themselves pass. Collection stopped before
  unrelated provider windows so a known-invalid candidate would not consume a
  full qualification cycle.
- Ownership finding: `deployable-shape` is a sequence-owned internal helper,
  but its authored Make prerequisites were not converted to the scheduler-aware
  prerequisite prelude used by public targets. Each CI child therefore reran
  `build-server`, `build-migrate`, and `build-operator` for roughly 20 seconds
  after `check` had already produced them, concurrently slowing harness and
  gosec work. OTel's now-subsecond source scan still needs explicit low-priority
  admission so its short process and summary envelope is not stretched by the
  initial CPU-heavy fan-out.
- Disposition: T-010 is reopened as the sole `IN_PROGRESS` item and T-012
  returns to `TODO`. Correct prerequisite suppression for sequence-owned
  internal helpers without weakening direct Make prerequisites, and give the
  OTel work unit a topology-owned non-blocking isolation policy. Revalidate
  direct helper behavior, scheduler contracts, and a fresh CI window from a new
  clean checkpoint. These three `867028de` roots remain retained diagnostic
  rejects and cannot enter the next candidate manifest.

### 2026-07-22 — T-010 prerequisite admission and bounded security analysis complete

- Source: mandatory reopen checkpoint `96c4bd14`, following the rejected early
  T-012 window from `867028de`. The implementation, owner specification,
  semantic task surface, generated topology, contract tests, and this tracker
  were completed before the new T-012 candidate freeze.
- Prerequisite ownership: every sequence-owned Make target now emits authored
  prerequisites through the scheduler-aware conditional prelude. Direct
  invocation still runs the complete prerequisite set. CI and release
  `deployable-shape` explicitly reuse the binary readiness established by their
  `check` or `build` dependency; release browser readiness retains its existing
  scheduler-owned reuse. `lint-go` deliberately does not suppress its format,
  vet, and staticcheck children because they are owned work rather than
  readiness. Contract coverage proves all three distinctions.
- Admission correction: default-check OTel conformance retains its ordinary
  CPU/I/O `1/1` profile and static-validation priority but waits for
  `check-service-backed` completion. Its exact scheduler envelope fell from the
  rejected `3.115 s` median to `1.414 s` and `1.335 s` in the two successful
  diagnostic proofs, both below the `1.6265 s` reference limit.
- Security analysis: the advisory audit now owns one physical repository scan
  over the stable union of runtime rules and minimal runtime/support package
  roots. A harness contract proves that every support-inventory root is
  subsumed. The `security_analysis` sequence profile claims the full resolved
  CPU capacity; the sequence scheduler strips inherited state and forwards its
  exact numeric CPU claim to the scan as `GOMAXPROCS`. This prevents the former
  undeclared 24-CPU scan from contending with the harness while eliminating
  duplicate Go package loading.
- Timing proof: final `make ci` passed all five outer work units and 510 tests
  in `149.643 s` at
  `.cartulary/test-results/20260722T202244Z-p963081`; nested `check` was
  `113.481 s`. Exact aggregate envelopes were harness contract `24.988 s`,
  gosec audit `4.625 s`, and OTel conformance `1.335 s`, below their retained
  limits of `25.678275 s`, `4.7045 s`, and `1.6265 s`. Deployable validation
  was `1.203 s` instead of the rejected duplicate-build interval near `22 s`.
  An earlier successful proof at
  `.cartulary/test-results/20260722T201700Z-p847073` recorded CI `146.143 s`,
  nested check `109.810 s`, harness `25.280 s`, gosec `4.531 s`, and OTel
  `1.414 s`.
- Functional and generated closure: final `make lint` passed all eight work
  units in `12.292 s` at
  `.cartulary/test-results/20260722T202131Z-p950934`. Final harness contract,
  generation drift, JSON shape, and generated-artifact policy checks passed at
  `.cartulary/test-results/20260722T202157Z-p958235`,
  `.cartulary/test-results/20260722T202157Z-p957953`,
  `.cartulary/test-results/20260722T202157Z-p957978`, and
  `.cartulary/test-results/20260722T202157Z-p957982`. `make generate` passed at
  `.cartulary/test-results/20260722T202119Z-p949144`, and the real shell gate
  passed after the bounded-audit edit.
- Diagnostic rejects retained: CI root
  `.cartulary/test-results/20260722T194531Z-p369691` proved that centralizing a
  prerequisite is insufficient without an explicit sequence skip policy;
  deployable validation still consumed `21.946 s`. Root
  `.cartulary/test-results/20260722T195117Z-p487931` proved prerequisite reuse
  but retained CPU-scan contention. Root
  `.cartulary/test-results/20260722T195657Z-p588478` isolated the two scans and
  narrowed gosec to a `4.920 s` near miss. Root
  `.cartulary/test-results/20260722T201513Z-p775395` was rejected for the
  focused ShellCheck variable-shadow diagnostic. Lint root
  `.cartulary/test-results/20260722T201959Z-p942235` exposed and rejected the
  legacy `lint-go` child-work skip policy. None qualifies for T-012.
- Active: T-010 is closed and T-012 is the sole `IN_PROGRESS` item. Freeze the
  clean checkpoint containing this tracker, then collect one successful
  warm-up and exactly two consecutive measured roots for each minimal provider
  window. Candidate evidence may not use retained-v1 migration exceptions, and
  all roots collected before the checkpoint remain diagnostic-only.

### 2026-07-22 — T-012 strict comparison reopens T-010 transition validation

- Source: clean frozen candidate checkpoint `4dc80ea57222d7b8890b58ea6316c2e71e80063c`.
  All 23 provider windows were collected from this source after rejecting and
  replacing three retry-contaminated observations. The complete selected set
  contains 69 strict-current roots with zero remaining retries, qualification
  reasons, or missing execution contexts.
- Rejection: the strict comparison reached normalized policy validation and
  rejected `ci` because the transition validator still expected the original
  `cpu_analysis` profile for `go-gosec-audit`. The owner specification,
  resource registry, execution topology, generated task surface, and retained
  candidate policy all correctly use the later adopted `security_analysis`
  profile: full resolved CPU, I/O one, process one, Make jobs equal to the CPU
  claim. The validator is stale, not the runtime policy.
- Disposition: T-010 is reopened as the sole `IN_PROGRESS` item and T-012
  returns to `TODO`. Align the exact normalized transition validator and its
  negative fixtures with the normative `security_analysis` policy, validate
  both `ci` and `release-check`, and checkpoint the correction before
  collecting a new candidate. Every `4dc80ea5` candidate root is retained as
  diagnostic evidence but cannot qualify after the source change.

### 2026-07-22 — T-010 transition validation complete; browser gates reopen T-011

- Source: mandatory T-010 reopen checkpoint `ead0ee9e`. The exact transition
  validator now recognizes `security_analysis` as full resolved CPU, I/O one,
  process one, with Make jobs equal to the CPU claim. Both `ci` and
  `release-check` require that profile for `go-gosec-audit`; the former stale
  `cpu_analysis` projection is rejected by a dedicated negative fixture.
- Validation: `make harness-contract`, `make lint-biome`,
  `make json-shape-check`, `make generated-artifact-policy-check`, and
  `make generate-drift` passed. Retained roots for the artifact checks are
  `.cartulary/test-results/20260722T223709Z-p2907059`,
  `.cartulary/test-results/20260722T223712Z-p2907552`, and
  `.cartulary/test-results/20260722T223714Z-p2907884`.
- Diagnostic comparison: the complete `4dc80ea5` manifest now passes exact
  policy validation and 46 of 48 public duration rows. The public portfolio
  falls from `2,721,126 ms` to `2,344,564.5 ms` (`-376,561.5 ms`). Backend
  unit (`40,324 -> 5,316.5 ms`), backend finalization
  (`6,991 -> 3,862.5 ms`), and release-check
  (`456,198.5 -> 290,293 ms`) pass required improvement.
- Remaining failures: `browser-e2e-webserver-backed` measured `109,475 ms`
  against a `104,108.55 ms` no-regression limit, and internal
  `release-browser-readiness` measured `139,917.5 ms` against its
  `139,088.7 ms` required-improvement limit. T-010 is closed. T-011 is
  reopened as the sole `IN_PROGRESS` item to diagnose retained browser
  critical paths and implement a durable reduction; T-012 remains `TODO`.

### 2026-07-22 — T-011 browser critical-path remediation complete

- Source: implementation worktree based on mandatory T-010 checkpoint
  `69b1a2f22d18aea3a12ff2d1aacd1568272f2ebf`; this record and its generated
  outputs form the mandatory T-011 checkpoint before new candidate collection.
- Retained diagnosis: release-check forwarded only CPU/I/O `2/4` into
  `release-browser-readiness`, so its CPU-two browser groups serialized even
  though the child schedule owned `browser_stack=2`. The webserver-backed
  managed-service stage likewise owned one stage lane for its default and
  claimed Network Flow sessions. These were topology bottlenecks, not product
  or Playwright behavior changes.
- Remediation: added the closed `nested_browser_validation` sequence profile
  with CPU/I/O `4/4`, process one, and registry-owned service-backed budget
  forwarding; release-check now uses that profile. The managed test-service
  webserver stage owns exactly two lanes. A target-local closed policy
  transition proves the exact two-session identities, runtime profiles,
  isolation reason, CPU-two group claims, and capacity change without changing
  sibling `standard_no_regression` targets.
- Safety boundary: a diagnostic direct-capacity prototype was rejected at
  `.cartulary/test-results/20260722T230342Z-p2980285` after two direct sessions
  raced the shared development Compose/CORS proxy and one failed startup. The
  prototype and draft schema bump were fully removed. Direct public browser
  leaves remain one-stack isolated; concurrency two is admitted only inside the
  managed test-service scheduler that owns fixture and cleanup isolation.
- Validation: `release-browser-readiness` passed at
  `.cartulary/test-results/20260722T225518Z-p2927420` with five sessions, peak
  stack two, and visual/accessibility capacities one. `browser-e2e-visual`
  passed without drift at
  `.cartulary/test-results/20260722T225808Z-p2953187`. `make test` passed 651
  tests at `.cartulary/test-results/20260722T231142Z-p3012400`; the two
  webserver session-start units began at monotonic 164 ms and 170 ms, peak
  stage capacity was two, and every service-backed unit/finalizer completed.
  `harness-contract`, `json-shape-check`, generated-artifact policy,
  `generate-drift`, and `lint-biome` passed on the final implementation.
- Active: T-012 is the sole `IN_PROGRESS` item. Freeze this clean checkpoint,
  discard every prior candidate window as diagnostic-only, and collect one
  successful warm-up plus exactly two consecutive measured observations for
  every minimal provider. Candidate roots may not use retained-v1 migration.
