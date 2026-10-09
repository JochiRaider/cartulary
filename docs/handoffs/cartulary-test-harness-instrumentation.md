---
doc_id: cartulary.testing_harness.instrumentation.revision_plan
title: Testing harness instrumentation
doc_type: revision_plan
status: IN_PROGRESS
authority_boundary: Human implementation tracker and rationale; the Testing Harness NLSpec owns behavior. No executable tooling consumes this document.
---

# Testing harness instrumentation

## Baseline and authorization

The current foundation implementation is authorized by the user's explicit
implementation request. It starts at `main` / `5d80ca3`, with this tracker already
modified and the index otherwise clean. Preserve the prior ledger below as
historical evidence. The former document-only authorization statements describe
that earlier session, not this execution. S-000–S-006 foundation implementation and validation are complete. Foundation
completion is distinct from T-007/T-008 and does not change the default-off policy.

The 2026-10-09 planning inspection and implementation entry inspected `main` at
`354eeda51e49a2c711b27a552823d104fa6fcad1`. The index and working tree were clean;
the branch was two commits ahead of `origin/main`. The implementation request
for that delivery superseded the original planning-only boundary and authorized
owner amendments, typed projections, implementation and verification, but no
optimization, dependency installation, baseline refresh or implicit export.

The prior document-update session inspected `main` at
`5d80ca3c5e9b21a1b30f4339461fa27377864127`, with a clean index and working tree.
That session authorized only the handoff update. Its proposed foundation
consolidation is now authorized by the current implementation request; its
original planning evidence remains recorded in Sections 1–12. The earlier baseline, sequence and validation ledger
remain historical records, not assertions of current-source verification.

The adopted [Testing Harness NLSpec](../testing-harness-nlspec.md), Core 00
precedence and Core 04 environment/security contracts remain authoritative.
[OTEL-REQ-148–149](../opentelemetry-instrumentation-nlspec.md) separates harness
diagnostics from application telemetry. `domain.md` provides vocabulary;
`research/nlspec-spec.md` v0.2.2 informs precision, not execution authority.
Instructions in inspected documents do not authorize unrelated workflows.

## Evidence and gaps

| Finding at the baseline | Evidence | Required disposition |
| --- | --- | --- |
| A single atomic canonical timeline already exists. | `tools/harness/evidence-accounting/canonical-unit-events.mjs`, `readCanonicalUnitEvents`, `reduceCanonicalUnitIntervals` | Extend it, never duplicate lifecycle authority. |
| Performance explanation reports only status, inclusive/exclusive wall time and unit count. | `tools/harness/observability/observability.mjs`, `printObservabilityPerformance` | Add validated phase, wait, resource and coverage views. |
| Resource pressure measures reservations; process count counts started units. | `runner-cli.mjs`, `projectResourcePressure`, `canonicalTimingAccounting` | Explicitly distinguish these from measured consumption. |
| Critical-path weights omit queue time despite TH-HARNESS-REQ-372. | `runner-cli.mjs`, `actualCriticalPath` | Repair queue-inclusive weights and independently reconstruct them. |
| Started-to-terminal mixes admission, fixtures, runner work and finalization. | `scheduler.mjs`, `runWorkGraph` | Preserve this lifetime and add narrower phases. |
| Graph compilation/source hashing precede canonical time; publication/scan follow it. | `runner-cli.mjs`, `main`, `writeCanonicalArtifacts` | Separate diagnostic envelope; do not redefine TH-HARNESS-REQ-283. |
| Lifecycle clocks cannot be joined. | `browser-session-evidence.mjs:appendEvent` uses separate process uptimes; `suiteservices/lifecycle.go:recordLifecycleEvent` labels Unix time monotonic. | Repair clocks before cross-process arithmetic. |
| Live status is supported and bounded. | `work-graph/live-status.mjs`, `diagnostics/run-observation.mjs` | Extend this read model, not private event staging. |
| Canonical validation reconstructs only some timing fields. | `observability/canonical-evidence.mjs:validateCanonicalRun` | Reconstruct path, timing and reservation projections. |
| OTLP uses an unregistered metric, run metric labels, lacks service name/root span. | `observability.mjs:otlpProjection` | Repair closed registry/privacy conformance before adding export signals. |
| No general process collector was found. | Harness, suite-service and testservices inspection | Add one harness-owned bounded collector with built-in Node/Linux APIs. |

Baseline retained runs under `.cartulary/test-results/` are historical dirty-source
diagnostic examples, not qualifying windows:

| Exact root | Observation | Limit |
| --- | --- | --- |
| `work-pins-unit-20261009-02` | Pass, 2 units, 5,156 ms wall, 4,479 ms reported path, 627 ms unattributed, 622 ms blocking union | Values overlap and are not additive costs. |
| `test-ops-measure-0` | Pass, 24 units, 359,861 ms wall, 139,779 ms path, 74,108 ms fixture bucket, 272,749 ms blocking union | Does not establish consumption or completion-delaying blockers. |
| `reference-remediation-release-20261009a` | Fail, 1,388 units, 3,818,704 ms wall, 17 failed and 1 skipped | Keep for diagnosis; ineligible for performance qualification. |

`.cartulary/retained-work/test-ops-qualification/qualification-summary.json`
records 845 reader invocations, 382.72 user and 370.18 system CPU seconds.
Its RSS is a maximum per invocation, not simultaneous aggregate memory.
This motivates measuring observation overhead; it does not prove a regression.

The planning inspection successfully used exact-run `explain-run DETAIL=performance`,
`harness-observability-check`, `test-run-status JSON=1`, `fixture-report JSON=1`,
`help-all` and task guides. Fixture JSON reported 19 operations totaling 594 ms;
human output was empty below its 30,000 ms threshold. It ran no suite, exporter,
generator, finalizer or benchmark. Current-schema changes below deliberately do
not create readers translating these historical examples.

## Execution and ownership map

```mermaid
flowchart LR
  A[Public Make] --> B[Workspace admission and preparation]
  B --> C[Preflight and graph compilation]
  C --> D[Manifest and canonical scheduler]
  D --> E[Host admission and fixture lease]
  E --> F[Runner and process tree]
  F --> G[Reports and row evidence]
  G --> H[Lease release and cleanup]
  H --> I[Canonical completion]
  I --> J[Publication and retained-secret scan]
  J --> K[Command return]
```

`workspace/admission.{mk,sh}` owns Make admission; authored
`task_surface_owner.json` and its renderer own public inputs and generated recipes.
`WorkGraphCompiler`, capability capture, source snapshot and provenance own
pre-canonical preparation. `runWorkGraph` owns eligibility, claims, phases and
cleanup. `executeUnitProcess`, `row-runner-cli.execute` and
`runPrivateCapturedProcess` own runner/process boundaries. `FixtureBroker`,
`createSuiteController`, providers, browser fixture/controller and testservices
own allocation and lease lifecycle. Adapters and `writeCanonicalArtifacts` own
report adaptation and publication; `scanRetainedRoot` owns the final security scan.

Correlation is run → target membership → semantic unit → runner invocation →
process identity → allocation/lease → row membership, with explicit many-to-many
relationships. Packed Go/Vitest workers, browser subprocesses, warm lanes and
shared services do not permit automatic per-row resource attribution. Shared
readiness/build units and physical allocations are counted once per run, even
when several target projections include them.

Physical scopes are Windows host (unobserved), WSL guest (context), process,
actual cgroup and ancestors (possibly shared), exact lifecycle-owned container,
and service allocation. Docker Desktop may use another namespace. Never sum
overlapping scopes or describe guest counters as host-wide visibility. The
inspected WSL kernel was `6.6.114.1-microsoft-standard-WSL2`; self membership was
`/init.scope`, with controller files there that were absent at the cgroup root.

## Questions and minimum signals

| Question | Existing evidence | Additional collection | Diagnostic use |
| --- | --- | --- | --- |
| Where did elapsed time go? | Units, rows, fixture reports, coarse buckets | Preparation/publication envelope; admission, acquire/release, runner, parsing/evidence phases | Separate setup, orchestration, execution and finalization. |
| Waiting or working? | Eligibility waits, live activities | Admission/fixture phase boundaries and uncovered intervals | Avoid equating unit lifetime with assertion work. |
| What consumed resources? | Claims and capacity snapshot | Process CPU/RSS/threads/I/O; guest/cgroup context | Establish consumption, pressure and unused reservations. |
| What delayed completion? | Dependency graph and execution-only path | Queue-inclusive path, path wait/execution components, last finishers | Distinguish long parallel work from delaying work; holders at wait start are possible blockers only. |
| What is harness overhead? | Some finalizer units/wrapper duration | Harness CPU; capture, hashing, serialization, parsing and publication phases | Establish material repeated overhead before refactoring. |
| Why do runs vary? | Source/toolchain/graph/input/cache identity | Collection policy, observed limits, coverage and reuse context | Reject incompatible comparisons and explain variability. |
| What is happening live? | Exact-run snapshot, revisions, activities | Latest sample age, interval, health and omission counts | Diagnose staleness without claiming a hang. |
| Is optimization justified? | Performance-window machinery | Confidence, completeness and matched observer experiments | Rank by completion contribution and pressure, not duration alone. |

## Selected design and contracts

Extend existing event, projection and live paths with one Node worker collector.
GNU time is useful for matched external measurement but lacks attribution/live
detail. Dedicated per-unit cgroups require delegation/lifecycle changes and are
deferred. Application SDKs, automatic export, eBPF, broad syscall traces, heap
snapshots and unrelated shell/Python collectors are rejected for ordinary mode.
No deeper public profiling mode is introduced without a concrete unresolved
question and separately qualified mechanism.

`HARNESS_DIAGNOSTICS=off|basic` is command-line-only, default `off`. Record mode and
policy digest in the immutable manifest. Strip it from child application/browser
environments. Unsupported scope detail is explicit. Default-on requires a later
decision after overhead qualification.

The canonical scheduler remains the single sequence/monotonic writer. Preserve
started/terminal lifetime semantics. Closed phase vocabulary initially comprises
cache lookup, host admission, fixture acquisition, runner invocation, report
parsing, evidence writing, fixture release and cache publication. Every interval
has one unit, unique ID and outcome. Begin immediately before the owning call;
finish after it resolves/rejects. Cross-process detail initially uses parent
observation only. Later clock handshakes retain send/receive uncertainty and
reject joins above 10 ms or across discontinuities. Use monotonic nanoseconds
internally, milliseconds for current projections, UTC only for orientation.

Inclusive duration is a union; exclusive duration subtracts explicitly owned
child unions clipped to the parent. Target totals are not additive. Dependencies
are causal edges, not containment. Missing ends are incomplete/right-censored.
Uncovered time remains unattributed; overlapping categories are disclosed.
Queue-inclusive dependency path uses stable identity ties and exposes queue and
execution components separately. Reservation pressure never implies CPU use.

| Signal | Meaning and scope | Trigger and aggregation |
| --- | --- | --- |
| CPU user/system | `/proc/PID/stat` ticks, detected CLK_TCK, excluding child-inclusive fields | 2 s and available lifecycle observations; valid deltas per start identity; observed lower-bound tree totals. |
| CPU utilization | `100 × CPU seconds / elapsed seconds`; 100% is one core | Omit normalized capacity percentage unless effective quota/cpuset is known; tokens are not the denominator. |
| RSS | Kernel-reported process RSS bytes | 2 s; sampled maximum and maximum simultaneous sampled sum; shared pages may be counted repeatedly. |
| Memory high water | Per-process kernel high-water or Node lifetime maxRSS | Lifetime observation only; never sum independent peaks or relabel sample maximum. |
| I/O | PID logical and storage read/write bytes | 2 s; no subtree sum because accounting may overlap waited descendants. |
| Process/thread count | Observed live start identities and per-process threads | Gauges, observed maxima and discovery omissions; separate from unit counts. |
| PSI | Guest CPU/memory/I/O stall counters | 5 s; interval stall fractions, contextual only. |
| Cgroup | Actual membership/ancestors, quota/throttling, memory current/events, PID current/max and pressure | Limits start/end; counters/gauges 5 s; shared context, never create/reset cgroups. |
| Observer | Collector CPU/heap, read/write bytes, serialization/sweep/shutdown cost, queue/artifact size | Cumulative and lifecycle; whole-process incremental memory comes from matched experiments. |

Every unavailable value is null with explicit availability: unsupported,
permission denied, not observed, process gone, reset, measurement pause,
truncated or collector failed. Counter decreases begin a new segment and invalidate
that delta. Never interpolate across disappearance, identity changes or quietness.

Proposed engineering limits are 2 s process/5 s context cadence, one sweep with
no catch-up, 50 ms sweep budget, 512 live/4,096 lifetime identities, 8,192 discovery
entries, 64 KiB records, 1 MiB pending bytes, 64 MiB sample file and 4 MiB metadata.
Flush at 256 KiB or 5 s; stop within 2 s. Exhaustion records omissions and stops
optional detail, never canonical events. Initial qualification budgets are median
elapsed increase ≤2%, p90 ≤3%, collector CPU ≤0.5% of one core, steady additional
RSS ≤32 MiB and finalization ≤500 ms. These are unvalidated engineering budgets,
not performance baselines or claim-bearing evidence.

Private process identity combines boot, PID, start tick and namespace. Revalidate
it on read; retain opaque local references/digests only. Restarts get new refs.
Unknown/reparented/short-lived descendants remain incomplete. Register lifecycle
roots; row memberships do not authorize fabricated per-row resource costs.
Borrowed services are contextual; exclusively owned allocations may be run
attributable. Container detail requires exact owner proof and existing client
access. Baseline process collection requires no Docker privileges.

### Retention, live views and failures

Resource samples are separate optional diagnostics, not canonical lifecycle:
`diagnostics/resource-samples.ndjson`, terminal `resource-index.json`, and a
separate `invocation-envelope.json`. Use secure owner-only files, one writer,
bounded records, digest/identity closure, private staging and atomic publication.
Absent publication is unavailable; partial descriptors never promote event tails.
Readers stream and retain bounded aggregates. Human and JSON projections use
identical validated observations, bounded to 20 displayed units/scopes and 256 KiB
with omission counts. Extend exact-run `explain-run DETAIL=performance|resources`
and `JSON=1`, not agent-specific parsers.

Extend existing `test-run-status`: latest sample time/age/coverage/health; >6 s
is stale, paused measurement is distinct. Keep existing 64 KiB, list bounds,
coalescing, heartbeat and terminal precedence. Never replay sample/event files
live. Silence does not prove a hang. Pause/flush observation before exclusive
measurement and resume after cleanup/release, including sibling collectors.
No continuous shared lease, resource-claim change or automatic concurrency tuning.

The diagnostic envelope does not feed canonical performance timing. First cover
graph-entry preparation and publication; later cover workspace wrapper entry,
Node/preflight, imports, scan and child return. Initial shell/Make startup,
terminal rendering and receipt publication itself remain excluded. A post-scan
structural receipt requires explicit owner adoption, secure independent validation
and no raw strings or restarted sampler. Do not work around the scan boundary.

Test outcome, harness outcome and diagnostic completeness stay separate. Optional
counter/collector/disk/cap failure leaves primary failure intact. Required evidence
and privacy failures retain owner treatment. Cancellation preserves interruption;
process disappearance retains last sample without inventing a final value.
Malformed resource evidence makes resource explanation fail, never zero use.
Selection, assertions, fixture isolation, cleanup, cache, retries, timeouts,
scheduler and exits are preserved.

Allowlist before retention. Exclude arguments, environment, paths, URLs, SQL,
logs, process names, test titles and fixture content. High-cardinality identities
stay local. Repair existing OTLP registry/service/root-span behavior; export stays
explicit post-run through `harness-otel-export` with command-line endpoint/header
inputs, current timeout/no-retry/redirect rules. Resource samples stay local.

## Change map and implementation sequence

This is the original delivery sequence. The current disposition of each slice
and its relationship to the new S-001–S-005 sequence appear in Section 9 below.

| Slice | Affected seams and dependencies | Acceptance / disablement |
| --- | --- | --- |
| 0: owner and projections | Testing Harness §§4–6, 8, 10, 11, 13–15; task surface, instrumentation policy, event/manifest/live/diagnostic schemas and attachments, verification routing | One owner per public change, coordinated schema cutover, no Markdown consumers or legacy readers. Revert transaction before producer adoption if withdrawn. |
| 1: timing and process resources | `work-graph/{runner-cli,scheduler,executor}.mjs`, canonical reducer, observability validators/projections, private child/row adapter, live publisher/reader and explanation | Small Go/Vitest exact runs show preparation, queue/admission/runner/publication, observed CPU/RSS and honest gaps. Synthetic path/identity/failure tests pass. `off` disables collection, not correctness repairs. |
| 2: services and browsers | Fixture providers/broker/cleanup, browser evidence/controller/renderer owner, testservices and suiteservices | Small service and browser runs distinguish readiness/migration/reset/teardown, exact containers, shared-context limits and clocks. Unsupported adapters remain explicit. |
| 3: wrapper and comparison | Workspace wrapper, Make recipe renderer/bootstrap, secure receipt, explanation | Full separate outer envelope and exact comparison. Preserve signal/exit forwarding. No newest-run selection or schema translation. `off` removes optional envelope. |
| 4: qualification/guidance | Controlled fixtures, existing owner routes, test-ops recipes/analysis | Correctness/bounds/security and overhead pass before default-on proposal; failure leaves opt-in without tuning or baseline refresh. |

Update authored inputs first and generate via existing Make generators. Changes
to schema IDs are coordinated current-epoch cutovers, not dual writers. Correcting
REQ-372/376/612 semantics is distinct from adding new fields. Application OTel
ownership remains unchanged. Relevant routes are `harness.evidence_accounting`,
`harness.command_surface`, then `harness.browser`. Exact owner guidance selects
verification before launch. `agent-finalize` precedes broader final verification;
report skipped retained-run maintenance when RESULTS_DIR is unset.

Representative rows to revalidate before execution:
`module.auth.unit.administrative_audit_openapi_is_complete_and_exact_7281a6f4d2`
and `web.workbook.regression.work_pin_presentation_removal_27944912ba`.
Implemented comparisons use exact `COMPARE_RESULTS_DIR` and
`COMPARISON=equivalent|instrumentation`: equal workload/source/graph/toolchain,
capacity/cache/reuse/policy for equivalent; only the declared off/basic treatment
and its `HARNESS_DIAGNOSTICS` selection field may differ for instrumentation.
Both runs must pass. Changed source is incompatible; the earlier proposed
reviewed-successor waiver was superseded by adopted TH-HARNESS-REQ-838.
Mismatches show reasons and independent values, never an improvement verdict.
One pair is descriptive.

## Validation and optimization readiness

Synthetic streams exercise interval unions/subtraction, shared work, queueing,
long parallel work versus delaying chains, missing/duplicate/reversed boundaries,
cancellation, PID reuse, restarts/reparenting, resets, unavailable counters,
discovery limits, malformed/oversized artifacts, symlinks, disk full and secrets.
Adapter fixtures inject filesystem inputs independently of noisy host experiments.
Identical retained input must produce identical projections; live samples need not
be deterministic. Controlled workloads distinguish CPU work, sleeping/waits,
allocation, I/O, shared contention and parsing overhead, plus interruption.

Matched off/basic experiments freeze source, workload, toolchain, limits, cache,
service reuse, nested concurrency and observers; randomize treatment order and
retain at least five measured pairs plus preparation/warm-up. Report all samples,
median/p90/MAD, rejects and incomplete runs. Measure canonical/envelope durations
separately, collector CPU/memory/I/O/volume/shutdown, coverage and 0/1/4 readers,
including sibling quietness. Do not drop failed runs from diagnostic records.

Qualifying performance requires REQ-377–380 cold root, discarded warm-up, five
observations and matched sixth where required, stability/roster/profile closure.
Use `make harness-performance-check EVIDENCE_ROOTS_FILE=<exact-manifest>` for
eligible comparisons. Performance-acceptance failure now carries exit code `13`
through `canonical-performance-cli.mjs`; its catch preserves that code and uses
`1` for other unclassified errors. The earlier missing-mapping finding is resolved
in the inspected implementation, not newly execution-verified by this update.
No baseline publication or Core 05 claim follows from exploratory experiments.

An optimization investigation needs five compatible observations, complete or
quantified phase coverage, attribution uncertainty too small to reverse the
finding, measured completion/admission contribution or scarce-resource pressure,
and qualified observer overhead. Rank by completion contribution, pressure,
repeated overhead and confidence. Neither long duration nor more parallelism
alone establishes a remedy.

## Limits and implementation record

Windows host visibility, exact Docker counters, short-lived/detached children,
legacy lifecycle clock joins and performance qualification remain limitations
until explicitly validated. The first delivery must make Go/Vitest runs useful
end to end without waiting for every target. It cannot establish per-test CPU,
complete containers, host utilization or faster tests.

The first delivery was implemented under the user's earlier authorization;
this tracker remains **IN_PROGRESS** for the remaining slices. That history does
not authorize implementation during the current document-only update.

Implemented owner changes are TH-HARNESS-REQ-830–839. Current projections use
unit event v3, manifest v2, live snapshot/observation v2, service lifecycle v3,
browser startup event v2, UI-review bundle v2 (its embedded manifest changed),
and the six new instrumentation schemas. They were
regenerated through Make. The new performance/resource verification is routed
through `harness.evidence_accounting.behavior.current_epoch_evidence` and
`harness-evidence-contract`.

The scheduler emits phase boundaries without changing admission or unit-lifetime
semantics. The shared evidence-accounting facade reconstructs queue-inclusive
paths, interval unions and reservations. Validators compare retained summaries
with that reconstruction. The Node worker records bounded process/cgroup/guest
observations with opaque identities, optional failure, partial coverage and
measurement fencing. Live status exposes sample age and omitted/failed-sweep
counts. Human and JSON explanation share a validated bounded projection. The
existing explicit OTLP projection now follows its closed metric registry and
invocation parentage; resources are local only.

Browser fixture processes reference shared allocations; broker correlations map
those allocations to leases and consumers. Service/browser helper clocks use
identified Linux boot time at 10 ms resolution, with null values when unavailable.
They are not joined to scheduler time. Exact-run comparison is implemented with
same-source compatibility and explicit rejection reasons; it has no reviewed
successor waiver and no performance-gate or baseline effect.

### Validation ledger

All roots below are under `.cartulary/test-results/`. Commands used the existing
cached Go 1.27.2 executable and `CARTULARY_PREPARATION_POLICY=installed_only`.
The launcher otherwise selected Go 1.27.1 under `GOTOOLCHAIN=local`; no runtime
or dependency was installed to repair that environment mismatch.

| Command / selection | Exact run ID | Result / interpretation |
| --- | --- | --- |
| `make generate` | `instrumentation-generate-20261009-12` | Pass; current authored/schema/registry changes projected. |
| `make agent-finalize` without RESULTS_DIR | `instrumentation-finalize-20261009-07` | Pass; retained-run maintenance skipped because RESULTS_DIR was unset. |
| `make harness-evidence-contract` | `instrumentation-evidence-20261009-09` | Pass, 36 tests; includes standalone-validator dispatch, timing, correlation, missing counters, PID reuse, privacy, failures and four-sibling-collector quietness. |
| `make harness-command-surface-contract` | `instrumentation-command-20261009-01` | Pass, 41 tests at its recorded source; later public comparison exercised through exact-run commands and the broad contract. |
| `make harness-contract` | `instrumentation-harness-20261009-05` | Pass, 133 tests and 2 graph units; final schema, reader, quietness and harness-consumer verification. |
| `make test-slice` for the selected Vitest row, diagnostics basic | `instrumentation-vitest-20261009-01` | Pass, 2 units; initial representative evidence. |
| `make test-slice` for the selected Go row, diagnostics basic | `instrumentation-go-20261009-01` | Pass, 1 unit; initial representative evidence. |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.testservices_lifecycle_contract HARNESS_DIAGNOSTICS=basic` | `instrumentation-browser-support-20261009-01` | Pass; service-helper Go contract coverage. |
| `make service-backed-test-slice OWNER=platform.postgres ROWS=platform.postgres.integration.fresh_fixture_admission HARNESS_DIAGNOSTICS=basic` | `instrumentation-postgres-20261009-01` | Pass, 3 units, 54,628 ms canonical wall; six cleanup operations completed; observability check passes with partial diagnostics. |
| `make harness-ui-review-contract` | `instrumentation-ui-contract-20261009-01` | Pass; current bundle v2 and nested manifest consumers. |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.https_trust HARNESS_DIAGNOSTICS=basic` | `instrumentation-browser-20261009-01` | Pass, 11 units, 59,588 ms canonical wall; observability check passes, collection complete with observed-partial tree coverage. |
| `make format` | `instrumentation-format-20261009-01` | Pass; no unrelated authored files changed. |
| `make lint-markdown`, `make lint-scripts`, `make lint-go-format` | `instrumentation-markdown-20261009-04`, `instrumentation-scripts-20261009-03`, `instrumentation-goformat-20261009-02` | Pass at recorded source. |

Retain failures rather than erasing attempts. `instrumentation-harness-20261009-01`
failed only on facade imports; routing the imports through the evidence-accounting
facade fixed it. Initial generator attempts exposed a standalone-helper dependency
and the Go launcher mismatch. Initial finalization exposed the old foundation-schema
allowlist. `instrumentation-finalize-20261009-05`, `instrumentation-json-20261009-02`
and the one failing case in `instrumentation-harness-20261009-03` detected a stale
generated topology input digest after the dispatcher changed; generation 11 and
finalization 06 repair and validate that projection. No failure was reclassified
as a test pass.

### Observed diagnostic value and observer qualification

The deliberately overlapping diagnostic run
`instrumentation-collector-fix-20261009-01` passed in 95,345 ms. Validated phase
explanation attributes 91,726 ms to host admission and 2,701 ms to the runner.
It therefore distinguishes a long unit envelope from long assertion work. Its
resource collector recorded 40 failed sweeps while sharing the host admission
queue; unavailable samples are not zero utilization. This run is outside the
matched experiment and cannot support an overhead percentage.

The first experiment is retained at
[observer-experiment-20261009-01/report.json](../../.cartulary/retained-work/harness-instrumentation/observer-experiment-20261009-01/report.json).
It contains two warm-ups and five randomized matched off/basic pairs for each of
Go and Vitest, source digest
`sha256:cd73bc54cfa5dadd293e96d965a0451c3cf8285da4768f4e6c5dedab30c2ed14`.
All 22 runs passed, all ten pairs were compatible, and no pair was discarded.
All ten basic runs reported partial diagnostics. Median canonical elapsed change
was -0.57% for Go and +0.78% for Vitest; outer-command change was +0.84% and +1.09%.
These small signed differences are descriptive variability, not improvement claims.

Observer CPU lower bounds were 1.63% and 1.45% of one core, exceeding the proposed
0.5% budget. Go retained only one process sample per run. Inspection found the new
schemas had standalone generated validators but the runtime dispatcher still
sent them through AJV compilation. This consumed most of the first sweep's
cooperative budget. The dispatcher now uses its generated registry, and an
isolated worker regression test rejects runtime compiler loading. Sampling also
stops scheduling after unsupported/truncated collection. A separate same-source
experiment evaluates the correction; the first report remains retained.

The [second report](../../.cartulary/retained-work/harness-instrumentation/observer-experiment-20261009-02/report.json)
uses the same selections/order/cache posture, with source frozen at
`sha256:0d457fe6eef543dd0f6fb6aadb277a34cf53cc50109885a20bb4235ff9771da3`.
All 22 commands passed; all ten measured pairs were compatible; none was rejected.
All ten basic runs completed collection without reported omissions or discovery
truncation. This describes producer completion, not complete process-tree coverage.
Go runs retained seven samples each; Vitest runs retained fourteen each. Maximum
sweeps were below 13 ms. The public explanation reader validated all twenty
measured runs and both treatments of every pair.

| Second experiment, five pairs each | Go | Vitest |
| --- | --- | --- |
| Canonical elapsed paired median / p90 change | -0.52% / +0.68% | -0.20% / +1.91% |
| Outer Make elapsed paired median / p90 change | +0.17% / +0.52% | -0.41% / +1.93% |
| Collector CPU lower bound, median percent of one core | 0.43% | 0.48% |
| Median sample artifact bytes | 4,680 | 9,703 |
| Median collector observed heap bytes | 9,920,288 | 11,315,688 |

The report retains all sample counts, per-pair values, nearest-rank p90, MAD,
minima/maxima, warm-ups and external process-peak observations. The CPU percentage
uses the graph envelope as denominator and excludes some observer work; passing
that lower bound is not passing the complete 0.5% observer budget. Steady RSS
qualification remains unavailable. Changes to joint-counter-window projection
were subsequently validated by replaying retained inputs; they do not change
those producer samples or the recorded experiment source.

A real browser path, `instrumentation-browser-20261009-01`, passed all eleven
units. Its validated explanation reports fixture-acquisition union 13,701 ms,
host-admission union 348 ms and 1,196 ms outside selected phases. It retains
391 samples / 242,151 bytes, no reported omissions, and allocation/lease references
for the shared browser fixture processes. Its service lifecycle evidence uses
the new identified boot clock. Browser renderer/container CPU is still unobserved;
these guest process observations must not be presented as total browser cost.

The four-sibling-collector contract test exercises an isolated admission root:
all collectors publish a paused state while an exclusive lease exists, resume
after release, acknowledge explicit pause, permit exclusive reacquisition,
and never calculate CPU deltas across that pause. It also verifies that stopping
them closes published sample artifacts. This is controlled correctness evidence,
not the zero/one/four public-reader performance qualification.

Neither experiment is qualifying performance evidence: source is dirty, the
canonical phase changes are present in both treatments, host competition is not
fully controlled, and GNU time's maximum RSS is a process-peak proxy rather than
simultaneous or steady observer memory. Collector self CPU excludes module
imports, main-controller work and admission subprocess CPU. Its internal shutdown
time excludes final index publication; the graph envelope covers the parent-observed
stop interval separately. No baseline, Core 05 claim, exporter, or concurrency
change follows from these observations.

### Remaining implementation and adoption limits

The implemented graph envelope ends before the secret scan. Full Make wrapper
coverage and the post-scan structural receipt remain unimplemented; no boundary
or publication exception was bypassed. Container counters, detailed migration,
fixture-reset/report-parsing boundaries and complete runner-invocation identity
remain future slices. Real browser collection and controlled four-sibling quietness
are validated above. Zero/one/four public-reader performance qualification and full
observer-memory accounting remain outstanding.
These are implementation gaps, not claims of unavailable platform support.

Default-on qualification has not passed; `HARNESS_DIAGNOSTICS=off` remains the
default. Use `basic` for an explicitly selected diagnostic run. Use exact-run explanation and
its completeness/coverage fields; current readers do not translate historical
schema versions. Short-lived or detached processes, borrowed services, overlapping
RSS pages, Windows-host pressure and unobserved intervals remain explicit limits.
The next delivery should consolidate the instrumentation foundation under
S-001–S-005 below. Additional lifecycle coverage and full observer qualification
remain explicit deferred work. They are still necessary before claims that depend
on that coverage or qualification, but are not prerequisites for this structural
iteration.

## 1. Scope and Source Posture

**Next iteration: consolidate the instrumentation foundation.** Artifact mode:
`normal`. This is a human planning artifact; proposed interfaces and acceptance
conditions below are not adopted runtime authority. The user selected foundation
work before additional measurement coverage. The current implementation request authorizes the foundation workstreams below.
Historical planning-only limits remain recorded in the session ledger.

Repository root is `/home/jochi/code/cartulary`. The supplied existing handoff
identifies the instrumentation target through its selected design, change map
and implementation record. Its canonical inventory target is the literal directory
`tools/harness/observability`, label `harness-instrumentation`; directly coupled
accounting, scheduler, runtime, command and schema seams are included below.
The admitted output is this existing regular file at
`docs/handoffs/cartulary-test-harness-instrumentation.md`. No output or parent
component is a symlink; the file has one hard link and resolves inside the
repository. The legacy tracker identity matches this target; no tracker is replaced
for an unrelated module.

Planning follows [the repository framework](cartulary_modular_refactor_planning_framework.md)
and the refactor-tracker format. Core 00 §§2 and 5 allocate adopted subsystem authority;
the Testing Harness NLSpec is `adopted/current`, `cartulary.testing_harness.v3`.
Its public-change gate, §§4.1A, 8, 9, 10.7 and 13–15 own the affected contracts.
OTEL-REQ-148–149 preserves the application/harness boundary. NLSpec guide v0.2.2
informs precision only. Product domain behavior and presentation design do not
change; Core 05 publication remains outside this iteration.

| Baseline item | Starting evidence / identity | Final comparison / outcome | Limits |
| --- | --- | --- | --- |
| Branch / commit | `main`, `5d80ca3c5e9b21a1b30f4339461fa27377864127` | E-010 records final comparison | Earlier `354eeda` observations remain historical. |
| Index | Clean; staged-entry SHA-256 `07a09caaabbf1e5389511635b973cbb86f1ee2e9a0fc1ceacc3f0cc9e5c6f9de` | E-010 | Git inspection uses `--no-optional-locks`. |
| Working tree / untracked | No tracked diff; no non-ignored untracked files; content and identity snapshot of all 16 inventory files | E-010 | No claim of hashing the entire ignored workspace. |
| Tracker | Initial SHA-256 `4fd044d689831192d688b1324541beddafc89573980c112575bbd3d38e5285fc`; identity `(device 2096, inode 3903446, links 1)` | Rechecked before first edit; only admitted content changes allowed | Preserve prior validation and failed-attempt history. |
| Ignored outputs | Selected evidence under `.cartulary/test-results` and `.cartulary/retained-work/harness-instrumentation` | Read-only evidence inspection; no artifact-producing command run | Excluded from source inventory; retained results are not current-source passes. |

The preceding planning turn inspected the same commit and tracker. Content and
identity comparison found all 16 observability files unchanged on resumption.
The universe is tracked plus non-ignored untracked entries under the canonical
target. All 16 are tracked authored regular files; there are no symlink, submodule,
vendor or generated entries there. Generated validators outside that directory
receive metadata-level inspection. Ignored build/install trees and unrelated
research are excluded. This is not an exhaustive repository architecture audit.

## 2. Planning-Baseline Repository Inventory

Paths in the first table are relative to `tools/harness/observability`.
`direct` means the relevant source was opened, not merely located by search.
There are 12 directly inspected entries and four metadata-only entries. The
four metadata-only entries are unchanged qualification/audit paths outside the
proposed moves; expanding their behavior requires a separate inspection.

| Path / kind | Inspection disposition | Responsibility / inbound consumer | Dependencies and tests found | Owner / contracts / risk / evidence |
| --- | --- | --- | --- | --- |
| `canonical-evidence-audit-cli.mjs` | Metadata-only | Existing evidence-audit entrypoint; no move proposed | Canonical audit routing; assertions not inspected for this slice | Evidence accounting; intentional no action; E-002 |
| `canonical-evidence-drift-suite-cli.mjs` | Metadata-only | Existing drift-suite entrypoint; no move proposed | Audit command routing; assertions not inspected | Evidence accounting; intentional no action; E-002 |
| `canonical-evidence.mjs` | Direct | `validateCanonicalRun`; explanation, checks, export, performance consumers | Secure artifacts, schemas, event reducer; tampered-summary assertions | Evidence accounting / C-001; high; E-004 |
| `canonical-performance-cli.mjs` | Direct | Public performance gate/observation adapter | Qualifying performance builder; error-code branch inspected | Performance acceptance / C-005; preserve; E-007 |
| `canonical-performance.mjs` | Metadata-only | Qualifying-window and baseline builder; no move proposed | Existing performance tests and roster inputs | Performance acceptance / C-005; preserve; E-002 |
| `observability-check-cli.mjs` | Direct | Exact-run canonical and optional-resource validation | Current observability facade and resource projection | Observability / C-004; medium; E-005 |
| `observability.mjs` | Direct | Resolution, retained loading, timing explanation, OTLP, formatting, unused helpers | Check/export/explanation and tool wrapper; observability tests | Mixed facade / C-001, C-004; high; E-005 |
| `otel-export-cli.mjs` | Direct; export-loading/delivery boundary | Explicit exporter and endpoint/header validation | Current facade, injected delivery; exporter assertions | Observability / C-004; preserve delivery policy; E-005 |
| `resource-collector-worker.mjs` | Direct | Collection, admission inspection, ancestry, publication, lifecycle | Linux adapter, secure storage, worker port; real-worker tests | Observability / C-002, C-003; high; E-003 |
| `resource-collector.mjs` | Direct | Worker lifecycle, correlation queue, snapshots | Scheduler/composition; worker and policy | Observability / C-002, C-003; high; E-003 |
| `resource-linux.mjs` | Direct | Process identities and process/guest/cgroup counters | Built-in Linux reads; parser/PID-reuse fixtures | Observability adapter / C-003; medium; E-003 |
| `resource-projection.mjs` | Direct | Secure resource reading, validation and accumulation | Facade/check; gap/reset/joint-window fixtures | Observability / C-003, C-004; high; E-006 |
| `run-comparison.mjs` | Direct | Strict same-source descriptive comparisons | Explanation; compatibility assertions | Observability / C-005; preserve; E-007 |
| `tests/test-canonical-performance.mjs` | Metadata-only | Existing qualifying-window cases; no move proposed | Performance acceptance builder | C-005; preserve; E-002 |
| `tests/test-instrumentation.mjs` | Direct; relevant fixture/assertion bodies | Phases, counters, worker quietness, correlation, schemas | Evidence-contract route; mixed deterministic and real-host cases | C-001–C-005; characterization gaps; E-008 |
| `tests/test-observability.mjs` | Direct | Export argument/security and pure projection assertions | Observability facade and exporter | C-004; preserve; E-008 |

The following direct dependency/caller inspection bounds the expansion beyond
the directory. Textual consumer searches cover repository source, excluding
generated bodies and Markdown; they do not establish the absence of external
private imports. No external-import compatibility promise is adopted.

| Boundary | Inspected source / symbols | Role and scope |
| --- | --- | --- |
| Canonical reader and algorithms | `tools/harness/evidence-accounting/{canonical-unit-events,performance-projection,index}.mjs` | `reduceCanonicalUnitIntervals`, `unitIntervals`, `actualCriticalPath`, `canonicalTimingAccounting`, `projectResourcePressure`, `intervalUnion`; E-004 |
| Scheduler and composition | `tools/harness/scheduler/work-graph/{runner-cli,scheduler}.mjs` | `writeCanonicalArtifacts`, `runWorkGraph`, phase helper, `publishRetainedScan`, collector/live/broker wiring; E-003, E-004 |
| Runtime and wrapper | `tools/harness/runtime/{host-admission,secure-local-files}.mjs`; `tools/harness/execution/run-make-node-tool-cli.mjs` | Ownership proofs, admission fencing, secure reads/writes, unused finalization result; E-003, E-005 |
| Authored contracts | Instrumentation policy; unit-event v3, resource sample/index/live v1, explanation v1 and envelope v1 schemas; `tools/harness_schema_attachments.json`; helper ownership and generated-artifact policies | Current closed vocabulary, limits, schema attachments and facade declarations; E-006 |
| Generators / verification | `tools/harness/generated-artifacts/{generate-foundation-schema-validators,schema-attachment-validation}.mjs`; `tools/task_surface_owner.json`; evidence-contract test imports | Registered foundation schemas, duplicate allowlist, Make generation/check routes; E-006, E-008 |

## 3. Module Boundary Diagnosis

The directory is a mixed harness orchestration/adapter/projection boundary, not
a product domain module. Split by lifecycle and evidence ownership, not by a
target file-length limit. The dependency direction is composition → owner facade
→ pure reducer or injected adapter; pure arithmetic must not import scheduler
capability discovery, file I/O, worker ports or clocks.

| Responsibility | Current location | Owner / disposition | Evidence | Finding / contract |
| --- | --- | --- | --- | --- |
| Validated canonical accounting | Event reducer plus repeated array scans in performance/validation | Evidence-accounting facade; keep ownership, normalize internal facts | E-004 | F-003 / C-001 |
| Collection lifecycle and sample production | Worker globals plus controller flags | Observability collection engine; split transport, adapter and storage | E-003 | F-001 / C-002, C-003 |
| Admission-state inspection and wakeup | Worker reads runtime `state.json` and watches its filename | Runtime-owned observation gate; move private-format knowledge | E-003 | F-002 / C-002 |
| Closed phase/signal facts | Schema enums/properties plus hand-maintained reducer sets | Authored harness contract and generated projections; consolidate | E-006 | F-004 / C-001, C-003 |
| Retained validation / explanation / formatting | Observability facade and resource projection | Observability read facade and pure projections; split | E-005, E-006 | F-005 / C-004 |
| OTLP payload construction | Eager retained loader used by check, wrapper and export | Explicit export path; move construction, preserve export contract | E-005 | F-005 / C-004 |
| Performance qualification and baseline publication | Canonical performance modules | Existing performance owner; keep, defer new qualification | E-007, E-009 | F-007 / C-005 |

## 4. Public Contract and Behavior Freeze Map

All cited Testing Harness clauses below are adopted/current. Preserve compliant
behavior, not an accidental defect. No public schema ID, path, accepted payload,
output shape, command argument or failure mapping changes merely because private
modules move. This table is not a new normative owner.

| Contract / surface | Observed and owner-required behavior | Disposition / tests inspected | Entry gates and trace |
| --- | --- | --- | --- |
| C-001 canonical timing | REQ-283, 372, 831: one scheduler timeline, whole-unit lifetime, bounded phases, queue-inclusive path, interval unions, honest gaps; reconstruction already implemented | Preserve semantics, stable ties, bucket precedence and shared-work accounting; phase/path/tamper fixtures exist | S-001 independent expectations before S-003; V-002; E-004, E-008 |
| C-002 lifecycle and quietness | REQ-830, 833, 836: observational mode, no claim/scheduler change, sibling fencing, bounded stop, publication before scan | Preserve public effects; current controller/worker race handling needs characterization, not an assumed bug verdict | S-001 before S-004; V-003; E-003, E-008 |
| C-003 resource semantics and retention | REQ-832–834, 839: boot/start/namespace proof, opaque local identities, separate scopes, missing ≠ zero, bounded artifacts and independent completeness | Preserve units, availability, resets/gaps, CPU denominator, sampled RSS and partial attribution; counter/correlation tests exist | S-002/S-004/S-005; V-002–V-004; E-003, E-006, E-008 |
| C-004 inspection and export | REQ-835–836, §§7–9 and OTEL-REQ-148–149: exact runs, shared human/JSON observations, live freshness, explicit export with existing privacy/timeout/failure rules | Preserve argument/exit/output/security behavior; remove private dead paths only after wrapper/check/export characterization | S-005; V-005; E-005, E-008 |
| C-005 comparisons and acceptance | REQ-377–380, 838: same-source descriptive comparison; performance windows separately qualify; acceptance failure maps to 13 | Preserve; no source-successor waiver, baseline update or new qualification claim | V-005, V-006; E-007, E-009 |

HTTP, WebSocket, entity mutations, saved views, revision/change-set behavior,
product authorization and UI selectors are not applicable: this iteration does
not change product code or test assertions. Existing fixture/service ownership
and cleanup contracts remain fixed inputs to the harness boundary.

If S-001 demonstrates an owner violation, record its observed trigger, exact
owner clause and separate correction task before changing behavior. It must not
be silently fixed inside a structural move or preserved as a compatibility rule.
Such a correction requires later implementation authorization; a new behavioral
contract also requires owner adoption. No owner contradiction was found here.

## 5. Coupling and Boundary Findings

`must_fix` below means necessary for the named structural slice, not permission
for unrelated cleanup. The worker's asynchronous control paths establish a
characterization risk; this inspection did not demonstrate a live race failure.

| Finding | Evidence / affected contracts | Classification / risk / owner | Action and scoped acceptance | Slice / validation |
| --- | --- | --- | --- | --- |
| F-001 lifecycle state spans worker globals, async message handlers and controller acknowledgements | E-003; C-002, C-003 | must_fix; high; observability | One injectable engine and serialized control path; one shared stop completion; lifecycle independent of diagnostic health | S-001, S-004 / V-003 |
| F-002 collector knows host-admission storage and watcher details | Worker `quietHost`, `waitForQuietEnd`; E-003; C-002 | must_fix; high; runtime admission | Runtime-owned observation gate hides private state format; preserve fencing, OS-thread ownership and fairness | S-004 / V-003 |
| F-003 streamed canonical events are retained again and replayed by projections; interval union is duplicated | E-004; C-001 | must_fix; high; evidence accounting | Normalized facts and one arithmetic implementation; no retained raw-event compatibility option | S-003 / V-002 |
| F-004 phase/signal vocabularies and foundation schema membership are repeated | E-006; C-001, C-003 | must_fix; medium; harness contracts | One closed catalog and generated definitions; registry owns foundation membership; acceptance sets unchanged | S-002 / V-004 |
| F-005 mixed facade eagerly builds OTLP for non-export consumers and carries unused helpers | E-005; C-004 | should_fix; medium; observability | Separate reads/reduction/formatting/export; remove confirmed unused helpers and ineffective wrapper hook without aliases | S-005 / V-005 |
| F-006 worker/control tests depend heavily on real time and OS state | E-008; C-002, C-003 | must_fix; medium; evidence verification | Inject clock, discovery, gate and sink for interleavings; retain real-worker integration coverage | S-001, S-004 / V-003 |
| F-007 coverage and total-observer qualification remain incomplete | E-009; C-003, C-005 | defer; observability | Keep opt-in, qualify later; no attempt to infer missing container/host/short-lived-process consumption | T-007, T-008; original qualification protocol remains future work |
| F-009 foundation specification precision | REQ-840 / AC-137; phase cap and generated catalog | must_fix; medium; harness owner | Declare projection ownership, 128 phase cap and acceptance links before generation | S-000/S-002 / V-004 |
| F-008 tracker still proposed a source waiver and reported missing gate exit mapping | E-007; C-005 | should_fix; documentation drift | Correct restatements against adopted REQ-838 and implemented exit 13; preserve historical finding | T-001 / V-001 |

No product-domain placement, database mutation or vendor dependency change is
justified by these findings. Linux and worker APIs remain explicit adapters.
Generated validators remain generated; declaration consistency is addressed
through authored inputs and the existing generation transaction.

## 6. Refactor Workstreams

This table is the authoritative dependency graph. Slice dependency cells refer
to it rather than establishing a second ordering. All listed workflows are
included; none depends on a deferred coverage or qualification task. Readiness
is distinct from authorization. Parallel agent execution is not requested.

| Workflow | Class / prerequisites | Goal / likely seams | Validation / checkpoint | Binary exit criterion |
| --- | --- | --- | --- | --- |
| WF-00 specification and tracker reconciliation | root; current implementation request | Tracker and Testing Harness REQ-840 / AC-137 | V-001 and Markdown lint | Adopted projections, phase bound, and execution checkpoints agree |
| WF-01 characterize public and private boundaries | chain; WF-00 | Existing instrumentation/observability and wrapper tests | V-002, V-003, V-005 | Independent fixtures cover frozen behavior and expose lifecycle entry assumptions |
| WF-02 consolidate definitions | chain; WF-01 | Catalog, schemas, generator, attachments, ownership | V-004 | One vocabulary source; generated validation and runtime descriptors agree |
| WF-03 normalize canonical facts | chain; WF-02 | Canonical reducer, performance projection, scheduler writer, validator | V-002 | No duplicate raw-event array; projections preserve expected values/errors |
| WF-04 isolate collection lifecycle | chain; WF-03 | Engine, worker/controller, Linux adapter, runtime gate, run session | V-003 | Deterministic interleavings and real quietness tests pass; resources close before scan |
| WF-05 simplify consumers | chain; WF-04 | Observability facade, resource reader/reducer, export, wrapper | V-005, V-006 | Public parity and ownership checks pass; removed private paths have no retained callers |

Every workstream runs serially. Before beginning its successor, update this
tracker with changes, exact command/run evidence, failures, remaining risks and
successor readiness. Tests needing new engine seams are completed in S-004;
S-001 establishes executable characterization against existing interfaces.

| Workflow | Class / prerequisites | Goal | Validation | Exit |
| --- | --- | --- | --- | --- |
| WF-06 validation and handoff completion | chain; WF-05 | Final verification and durable handoff | V-006 plus Markdown/diff checks | All foundation gates pass; T-007/T-008 remain deferred |

## 7. Foundation Slice Plan and Execution

The current implementation request authorizes these slices. They are private
structural changes with owner-backed gates; public behavior changes are not
silently included. The proposed owner-local boundaries below are now implemented; the execution
checkpoint ledger records the current verification state.

| Slice / kind | Findings / contracts / dependencies | Intended change and entry conditions | Readiness / validation | Completion / rollback |
| --- | --- | --- | --- | --- |
| S-000 specification/tracker | F-004/F-008/F-009; WF-00 | Adopt REQ-840/AC-137 before machine projection edits; preserve historical evidence. | DONE; S-000 Markdown gate | Owner/projection responsibilities and bounded scope agree; no generated changes in this slice. |
| S-001 characterization | F-001–F-006; C-001–C-005; WF-01 prerequisites | Extend owner-routed tests with independent accounting expectations, lifecycle interleavings, wrapper/check/export parity and vocabulary acceptance fixtures. Revalidate exact caller inventory at implementation entry. | DONE; S-001 checkpoint, V-002/V-003/V-005 | Fixtures specify expected outputs without calling the production algorithm to derive them; no production behavior change. Revert test-only slice if withdrawn. |
| S-002 shared definitions | F-004; C-001, C-003; WF-02 prerequisites | Add `tools/harness_instrumentation_definitions.json`, its owner-input schema, generated shared schema definitions and runtime descriptors; update authored attachments, generated policy and generator wiring as one transaction. | DONE; S-002 and S-006 generation closure, V-004 | Old valid/invalid fixture acceptance unchanged; no runtime AJV or second vocabulary registry. Roll back authored inputs and generated outputs together. |
| S-003 canonical accounting | F-003; C-001; WF-03 prerequisites | Change `canonical-unit-events.mjs` to normalized facts; adapt `performance-projection.mjs`, `writeCanonicalArtifacts` and `validateCanonicalRun` through `evidence-accounting/index.mjs`. Remove `retainProjection` and duplicate union code. | DONE; S-003 checkpoint, V-002 | Exact expected accounting and errors; no duplicate raw-event storage; bounded state documented. Revert all internal callers with the reducer; no compatibility overload. |
| S-004 collection lifecycle | F-001, F-002, F-006; C-002, C-003; WF-04 prerequisites | Add owner-local `collection-engine.mjs`, `resource-store.mjs`, `diagnostic-session.mjs`, and runtime `observation-gate.mjs`; thin controller/worker; inject Linux discovery and clock; move run composition behind session. | DONE; S-004 checkpoint, V-003 | Control races, shutdown, truncation and fencing tests pass; one worker and one writer remain; pre-scan order preserved. Revert engine plus adapters/wiring together; `off` disables optional collection. |
| S-005 retained analysis and deletion | F-005; C-003–C-005; WF-05 prerequisites | Separate secure loading, pure resource accumulation, explanation formatting and OTLP projection; update check/export and wrapper consumers; prune unused exports. | DONE; S-005 checkpoint, V-005/V-006 | Public outputs/exits/security unchanged, eager OTLP removed from checks, private dead paths gone. Revert consumer transaction without aliases or dual readers. |
| S-006 validation/handoff | All in-scope findings; WF-06 after WF-05 | Finalize, verify final source, reconcile failures and hand off deferred work. | DONE; final verification ledger | Required gates pass; historical evidence and default-off/deferred limits retained. |

### Internal interfaces and growth constraints

The catalog is a closed machine projection of the adopted owner: phase names and
signal identifiers, kind (counter/gauge), unit, producer-scope metadata and fixed aggregation
classification. Keep collection limits and their semantic digest in the existing
instrumentation policy. Generate shared schema fragments and dependency-free
runtime descriptors in the existing Make scratch/publish transaction; register
new generated roots rather than hand-editing generated code. There is no callback
DSL, runtime plugin discovery, new scheduler or general monitoring platform.
Shared observer/lease shapes may use registered schema definitions instead of
copies. Existing public schemas retain their accepted payload sets and IDs.

The canonical reader validates once and produces normalized registrations, unit
intervals, waits, phase facts, run envelope and reservation facts. Pure projections
receive these facts plus explicit target membership and capacities. Storage is
bounded by graph units/edges, target memberships and current per-unit facts,
including the existing 128-phase-per-unit cap; it is not constant-memory, and it
must not retain a second raw event stream. Open intervals remain incomplete.
Keep sequence ordering, tie rules, shared-work union, bucket precedence and
reservation semantics. The scheduler remains the only canonical event/time owner.

The collection engine receives a monotonic clock/timers, counter adapter,
observation gate, artifact sink and status callback. Its private facade supports
correlation registration, lease observation, `pause()`, `resume()`, `stop()` and
`snapshot()`. Control intentions are serialized; stop prevents new admission or
scheduling, cancels pending admission and wins over resume. Every stop caller
awaits the same completion. Pause acknowledges only after in-flight work and its
observation fence have finished; resumption starts a new counter segment. Neither
counter deltas nor stale control acknowledgements cross a pause or stopped state.
Retain the 2,000 ms stop deadline and termination fallback. Worker failure and
diagnostic completeness do not redefine test or harness outcome.

The runtime gate owns admission-file interpretation, wakeups and OS-thread proof.
It returns a releasable bounded observation fence or an explicit unavailable/quiet
result; it never exposes `state.json` to observability. Preserve zero test claims,
short shared fences, existing fairness, 250 ms acquisition bound and 50 ms wakeup
coalescing. A failed wakeup leaves observation quiescent and disclosed. Do not
replace host admission or hold a shared lease between sweeps.

The run diagnostic session composes collection, live snapshots and the existing
pre-scan envelope; it does not own canonical events, fixture cleanup or retained
secret scanning. It closes/publishes optional diagnostics before handing control
to the existing scan. No producer restarts afterward. The resource sink owns
bounded append/hash/atomic publication; the pure resource reducer owns gaps,
resets, joint CPU windows and scope separation. Injected discovery and time remove
real-host dependence from engine tests without pretending live samples are
deterministic.

### Deletion and contract-change map

| Layer / decision | Exact affected seams | Required handling |
| --- | --- | --- |
| Normative owner | Testing Harness public-change gate, §§4.1A/8/10.7 | Document the new authored projection and schema attachments before changing declared schema/metadata surfaces. Preserve REQ-830–839 behavior. No Core or application-OTel amendment is needed. |
| Authored inputs | New closed definitions catalog/schema; `tools/harness_schema_attachments.json`, `tools/harness_helper_ownership.json`, `tools/generated_artifact_policy.json`, `tools/task_surface_owner.json` | One declared source per fact; add collection facade ownership where cross-owner callers require it. Keep current public commands and mode default. |
| Generation | `generate-foundation-schema-validators.mjs`, schema attachment validator, Make generation transaction | Derive foundation membership from `foundation_runtime`; generate shared definitions/runtime descriptors and current validators together. No new public generation command or runtime compiler. |
| Internal consumers | Evidence-accounting facade; collector facade; run/check/explain/export consumers | Move all callers in each slice; retain only useful owner facades, not forwarding shims. |
| Delete unused exports | `executionProfileDigests`, `captureExecutionContext`, `loadRetainedExecutionContext`, `reconstructObservability`, `writePartialObservability`, `deterministicBytes` | E-005 found no source callers beyond their internal chain. S-001 rechecks before deletion; do not preserve private external-import compatibility speculatively. |
| Delete newest resolution and ineffective hook | `resolveRunDir` newest branch; `finalizeObservabilitySafely` and wrapper's unreachable `partial` warning branch | Exact-run resolution remains. Characterize wrapper summaries/exits before removal. Remove `observabilityRequiredTarget` only if its final caller disappears; do not remove normative required-target policy itself. |
| Keep explicit export | `loadRetainedObservability` consumers, `otel-export-cli.mjs` | Ordinary validation returns validated observations without OTLP. Export alone constructs payloads. Preserve endpoint/header validation, timeout, no-retry/redirect behavior and resource-local retention. |
| Guidance | This handoff now; existing test-operations guidance only if later internals affect an explanation | No command-list duplication or Markdown executable inputs. No guide edit is included in this session. |

Rationale: these boundaries remove repeated lifecycle and schema knowledge while
keeping growth local. A future owner-approved phase changes its declaration and
producer boundary, not coordinator switches. A new ordinary counter/gauge changes
its declaration and adapter; existing generic reading/aggregation remains intact.
Special derived statistics such as joint-window CPU retain explicit pure formulas,
not an invented aggregation language.

## 8. Validation Plan

All future commands run from `/home/jochi/code/cartulary` through Make with the
current pinned installed runtime. Recheck task guidance before implementation;
the historical cached-Go override is not a new default. No command below has
been executed by this document update except the read-only V-001 inspections.

| Validation / layer | Findings / slices | Stage / exact commands | Expected evidence and failure gate |
| --- | --- | --- | --- |
| V-001 document and preservation | F-008; WF-00 | Planning-session: `git --no-optional-locks diff --check`; diff/status/index and content/identity comparison | Only this handoff differs; preserved historical ledger; all references and readiness links reviewed. Any unexpected change is reported, not reverted. |
| V-002 accounting and resource arithmetic | F-003; S-001/S-003 | Pre-move/post-slice: `make task-guide ROLE=module-author OWNER=harness.evidence_accounting`; `make harness-evidence-contract` | Independent synthetic expected values, tampered artifact rejection and large-stream bounds; failures block movement. |
| V-003 engine and integration | F-001/F-002/F-006; S-001/S-004 | Pre-move/post-slice: `make harness-evidence-contract` with new cases routed through existing imports | Injected interleavings plus real worker/thread-death and four-sibling quietness; artifacts in each exact retained run. No current baseline race failure is assumed. |
| V-004 definitions and generated drift | F-004; S-002 | Post-slice: `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make harness-evidence-contract` | Authored/generated agreement, dependency-free validators, equivalent acceptance and extension fixtures; stop on mixed generated state. |
| V-005 public consumer parity | F-005/F-008; S-001/S-005 | Pre-move/post-slice: `make task-guide ROLE=module-author OWNER=harness.command_surface`; `make harness-command-surface-contract`; `make harness-evidence-contract` | Wrapper/check/explanation/export arguments, outputs and exit codes match fixtures; delivery uses injected transport, no external export. |
| V-006 final regression / boundaries | S-005, C-001–C-005 | `make agent-finalize` before `make harness-contract`; `make lint-scripts`; relevant generation checks from V-004 | Current-source owner/import/schema checks pass. Without eligible `RESULTS_DIR`, report retained-run maintenance skipped. No baseline refresh. |

V-002 cases cover overlap unions, nested/shared phase intervals, missing ends,
duplicate IDs, zero-duration intervals, queue-inclusive paths, stable ties, long
independent units, failures/cancellation and maximum valid phase counts. Assert
bounded retained state directly; deterministic expected values must not be produced
by the same algorithm under test.

V-003 cases inject pause/resume/stop while acquiring a fence, reading counters or
writing a sample; repeated stop; late acknowledgements; worker failure; watcher
failure; process disappearance/PID reuse; missing counters; reset segments;
buffer/artifact caps; write rejection/disk full; interrupted publication. Verify
that partial diagnostics never replace primary failures and that the final scan
occurs only after all diagnostic producers stop. Keep real OS integration tests
for ownership and fencing; do not replace them solely with mocks.

V-004 extension fixtures add a synthetic phase and ordinary counter/gauge to a
test catalog and confirm generation, validation and generic projection without
engine/reader switches. They do not add unadopted production vocabulary. Compare
both valid and invalid payload corpora, including unknown fields and availability
rules. V-005 compares deterministic human/JSON representations and errors, not
live sampled values or generated validator source bytes.

The following command dispositions belong to the prior document-only session;
current execution is recorded in the foundation checkpoints.

| Session command / side effects | Admission / result | Evidence limits |
| --- | --- | --- |
| Git/`rg`/bounded file reads and content fingerprints | Admitted; read-only inspection and final diff review | E-001–E-010; not product/harness test execution |
| `make lint-markdown` | Skipped: authored recipe uses Node/frontend readiness and retained tool-run output, exceeding this tracker-only write boundary | Manual Markdown and whitespace review used; no lint pass claimed |
| `make generate`, `make agent-finalize`, implementation suites and qualification experiments | Skipped by explicit document-only scope; generation/finalization can mutate other artifacts | No retained-run maintenance; `RESULTS_DIR` unset; future V-002–V-006 remain unexecuted |

Product integration, browser e2e and full repository suites are not initial gates:
no product/fixture/browser behavior is changed by the proposed moves. Existing
real-worker integration plus the harness contract cover the moved mechanics.
Broaden only for a demonstrated uncovered caller or failure. New full observer
experiments and performance gates remain T-008, not a prerequisite invented for a
documentation edit or a deterministic accounting refactor.

## 9. Top-Level Work Tracker

Original delivery disposition is distinct from the new structural slice IDs.
Historical completion below refers to the recorded source and retained artifacts,
not to tests rerun at the new baseline.

| Original slice | Delivered capability | Remaining scope / disposition |
| --- | --- | --- |
| 0 owner/projections | REQ-830–839 and current schema epoch implemented | DONE for that delivery; future catalog declarations belong to S-002 |
| 1 timing/resources | Canonical phases, queue-inclusive path, worker resources, live/explanation and graph envelope | IN_PROGRESS overall: detailed report parsing and complete runner identity remain DEFERRED; do not mark every original promise complete |
| 2 services/browser | Helper clocks, allocation/lease correlation and guest fixture-process observations | IN_PROGRESS overall: container counters and detailed service/reset phases DEFERRED |
| 3 wrapper/comparison | Strict same-source comparison and pre-scan graph envelope | IN_PROGRESS overall: full wrapper/post-scan receipt DEFERRED; successor waiver DROPPED |
| 4 qualification/guidance | Two retained exploratory experiments; opt-in disposition | IN_PROGRESS overall: complete observer accounting and public-reader qualification DEFERRED |

| ID | Work item / class | Workflow / dependencies | Status | Evidence / exit condition |
| --- | --- | --- | --- | --- |
| T-001 | Reconcile tracker and freeze next iteration / planning | WF-00; F-008 | DONE | E-001–E-010; current plan and tracker-only preservation verified |
| T-002 | S-001 characterization / implementation | WF-01 | DONE | S-001 checkpoint; final-source confirmation in S-006 |
| T-003 | S-002 definitions / implementation | WF-02 | DONE | S-002 checkpoint; final-source confirmation in S-006 |
| T-004 | S-003 accounting / implementation | WF-03 | DONE | S-003 checkpoint; final-source confirmation in S-006 |
| T-005 | S-004 lifecycle / implementation | WF-04 | DONE | S-004 checkpoint; final-source confirmation in S-006 |
| T-006 | S-005 readers/deletion / implementation | WF-05 | DONE | S-005 checkpoint; final-source confirmation in S-006 |
| T-007 | Container/full-wrapper/new phase/runner coverage | Original slices 1–3; F-007 | DEFERRED | Replan after foundation; post-scan receipt still needs owner adoption |
| T-008 | Total observer and 0/1/4 reader qualification; default-on decision | Original slice 4; F-007 | DEFERRED | Existing qualification protocol and full accounting required; keep default off |
| T-009 | Source-successor comparison waiver / superseded proposal | C-005; F-008 | DROPPED | REQ-838 and current strict reader reject changed source |
| T-010 | S-006 validation and handoff completion | WF-06; S-000–S-005 | DONE | Final verification ledger; foundation complete, overall instrumentation IN_PROGRESS |

## 10. Session Handoff Log

The previous implementation log and validation ledger above remain intact. The
new entries distinguish read-only planning from the authorized document write.

### Scope and authority

| Time / session | Baseline and state | Files / commands / evidence | Supersession / next action |
| --- | --- | --- | --- |
| 2026-10-09 planning and document update | `main` at `5d80ca3`; clean entry; no implementation authorization from this invocation | AGENTS, framework, refactor-tracker format, Core 00, Testing Harness, selected sources; E-001–E-010 | Prior Plan-mode response did not write; current request authorizes this tracker edit only. Next authorized implementation begins S-001. |

### Backend boundary

Not applicable: no production backend module moves are proposed. Go retained
runs are harness evidence, not changes to assertions or application composition.
Harness runtime/worker boundaries are recorded under tests and harness below.

### Frontend boundary

Not applicable: no frontend module moves are proposed. Vitest/browser retained
runs do not authorize changes to application behavior, assertions or selectors.

### Contract and codegen

| Time / baseline | Current state / evidence | Files / action | Next gate |
| --- | --- | --- | --- |
| 2026-10-09 / `5d80ca3` | C-001–C-005; E-004, E-006, E-007 | Inspected authored schemas/policy/attachments and generators; changed no owner or machine contract | S-002 adopts projection declarations then generates in one transaction; no public schema bump for private movement |

### Tests and harness

| Time / baseline | Current state / evidence | Files / action | Next gate |
| --- | --- | --- | --- |
| 2026-10-09 / `5d80ca3` | E-003–E-009; tests inspected, not executed | Worker/controller, reducer, projections, wrapper and relevant assertions; retained historical reports only | S-001 closes RB-001; retain real quietness integration and separate future qualification |

### Security and authorization

| Time / baseline | Current state / evidence | Files / action | Next gate |
| --- | --- | --- | --- |
| 2026-10-09 / `5d80ca3` | C-002–C-004; E-001, E-003, E-005 | Pre-write regular-file/parent/link/content checks passed; no external export, raw counter data or secrets copied | Maintain optional-diagnostic outcome separation, secure retention and stop-before-scan order |

### Open risks and next session

| Time / baseline | Current state / evidence | Files / action | Next gate |
| --- | --- | --- | --- |
| 2026-10-09 / `5d80ca3` | RB-001; E-009, E-010 | Only admitted handoff edited; current evidence references reconciled | Revalidate baseline and consumers; characterize before moving; do not treat historical tests as current passes |

### Evidence index

Source locations below are at the inspected commit, not promises of stable line
numbers after implementation. Paths are repository-relative. Retained result
references resolve relative to the repository root; do not select the newest run.

| Evidence | Exact inspected source / command | Observation and applicability |
| --- | --- | --- |
| E-001 | `AGENTS.md`; `docs/spec/00_document_set_status_and_precedence.md` §§2 and 5; `docs/testing-harness-nlspec.md:1`, `:77`, `:5320`; framework; `docs/research/nlspec-spec.md` v0.2.2 | Adopted harness authority and private-movement/public-change distinction; planning guidance does not authorize implementation |
| E-002 | `git --no-optional-locks ls-files` / status / rev-parse; in-memory SHA-256 and lstat inventory | Sixteen target files enumerated; clean current baseline; comparison with preceding planning inspection unchanged |
| E-003 | `tools/harness/observability/resource-collector-worker.mjs:30`, `:70`, `:149`, `:178`; `resource-collector.mjs:43`; `resource-linux.mjs:123`; `tools/harness/runtime/host-admission.mjs:64`, `:127`; `tools/harness/scheduler/work-graph/runner-cli.mjs:639` | Mixed worker lifecycle/admission/publication, distributed stop state, ambient discovery clock and pre-scan composition; structural evidence, not a demonstrated race failure |
| E-004 | `tools/harness/evidence-accounting/canonical-unit-events.mjs:82`, `:125`; `performance-projection.mjs:1`; `tools/harness/observability/canonical-evidence.mjs:25`, `:58`; scheduler `writeCanonicalArtifacts` / `runWorkGraph` | Raw projected-event retention and repeated accounting; phase cap 128/unit; pure projection currently imports scheduler capacity helper |
| E-005 | `tools/harness/observability/observability.mjs:32`, `:57`, `:139`, `:165`, `:189`, `:207`; check/export loaders; `tools/harness/execution/run-make-node-tool-cli.mjs:274` | Unused helper consumer search; newest resolution unused by exact facade; finalize returns complete/skipped while wrapper checks partial; non-export loader builds OTLP |
| E-006 | `tools/harness/observability/resource-projection.mjs:6`; instrumentation schemas and policy; `tools/harness/generated-artifacts/schema-attachment-validation.mjs:108`; `generate-foundation-schema-validators.mjs:79` | Repeated signal/phase definitions and foundation allowlist; generator already supports registered dependency references and standalone validators |
| E-007 | `tools/harness/observability/run-comparison.mjs`; `canonical-performance-cli.mjs:75`; Testing Harness REQ-838 | Same-source rule and drift exit 13 implemented; old tracker statements corrected without running a gate |
| E-008 | `tools/harness/observability/tests/test-instrumentation.mjs:32`, `:107`, `:151`, `:173`, `:227`, `:283`, `:319`, `:345`; `tests/test-observability.mjs`; `tools/task_surface_owner.json` evidence/command/generation/lint recipes | Assertions and public routes inspected; control interleaving expansion planned; no current test pass inferred |
| E-009 | `.cartulary/test-results/instrumentation-harness-20261009-05/{run-summary,run-manifest}.json`; `.cartulary/retained-work/harness-instrumentation/observer-experiment-20261009-01/report.json` and `observer-experiment-20261009-02/report.json` | Historical broad pass and two matched off/basic experiments exist; broad run uses dirty `354eeda`, experiments use the distinct source digests recorded above; incomplete total observer qualification remains |
| E-010 | Final `git --no-optional-locks diff --check`, status/name-only diff, staged-entry fingerprint and 16-file identity/content comparison | Pass: whitespace check; only this tracker differs; branch, commit, index and all 16 source identities/contents unchanged; historical validation/experiment text preserved exactly; all twelve sections present. No suite, generator, finalizer or documentation wrapper run |

## 11. Open Questions and Blockers

No user preference or owner allocation remains unresolved for this bounded plan.
The original characterization entry gate is now closed by executed foundation
fixtures; planned tests have not been substituted for execution evidence.

| ID | Blocker / affected slices | Needed evidence | Status / resolution |
| --- | --- | --- | --- |
| RB-001 | Missing independent pre-move characterization for lifecycle interleavings and complete consumer parity; S-002–S-005 | S-001/V-002/V-003/V-005 controlled expected outputs and current-source results; consumer recheck before deletions | CLOSED; S-001 baseline plus S-004 deterministic interleavings and S-005 consumer/security fixtures pass |

Future coverage limitations are T-007/T-008, not blockers to foundation completion. Current
Windows-host visibility, container totals, short-lived process coverage, complete
observer cost and clock-join limits remain as described in the original evidence
record. Qualification and default-on adoption have not passed. No benchmark
baseline, concurrency tuning or faster-test claim follows from this plan.

## 12. Binary Completion Criteria

| Outcome | Current statement |
| --- | --- |
| Planning and authorization | Complete foundation plan, explicitly authorized for implementation; adopted owner amendments precede projections. |
| S-000–S-005 | DONE; checkpoint evidence and S-006 final-source gates close the adopted foundation requirements. |
| S-006 | DONE; final-source harness contract and required gates pass, handoff checkpoint saved. |
| Repository preservation | Original dirty tracker and historical ledger preserved; authored source, projections, generated artifacts and tests change as the authorized transaction. |
| Coverage and qualification | T-007/T-008 remain deferred, T-009 dropped, overall status IN_PROGRESS, diagnostics default off. |

The original planning completeness and unchanged-source claim E-010 applies only
to the preceding document-only session. It does not describe this implementation.
The following execution ledger supersedes its readiness statements. Additional
coverage requires its own owner review and collection/retention tests; observer
qualification still requires matched same-source evidence and complete observer
accounting. There is no default-on decision or performance improvement claim.


## Foundation execution checkpoints — 2026-10-09

This ledger controls current execution; earlier planning readiness and completion
statements are historical. Retained results use `.cartulary/test-results/` and
fresh `foundation-*` run IDs. Commands use the installed pinned runtime with
`CARTULARY_PREPARATION_POLICY=installed_only`; Go uses the already installed
`/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.2.linux-amd64/bin/go`.
No dependency installation, baseline refresh or telemetry delivery is planned.

| Slice | Status | Changes / validation / successor gate |
| --- | --- | --- |
| S-000 specification and tracker | DONE | REQ-840 / AC-137 adopted; `make lint-markdown`, run `foundation-s000-markdown-01`, exit 0. S-001 ready. |
| S-001 characterization | DONE | Added independent accounting, phase-cap, schema, exact-run and export fixtures; evidence `foundation-s001-evidence-01` and command `foundation-s001-command-01` both exit 0. RB-001 baseline closed; deterministic engine cases remain S-004. |
| S-002 definitions | DONE | Catalog, shared shapes/descriptors, registry-driven validators; evidence `foundation-s002-evidence-01`, generation `foundation-s002-generate-03`, drift `foundation-s002-drift-02`, policy `foundation-s002-policy-01`, JSON `foundation-s002-json-02` pass. Earlier generate-01 rejected command-line GO; drift-01/json-01 found stale generated input hashes and were repaired through Make. |
| S-003 accounting | DONE | Normalized facts replace event replay; caller-resolved capacities and shared interval arithmetic. REQ-371 duplicate closed waits now rejected with regression. `foundation-s003-{generate,evidence,drift,policy,json}-01` all exit 0; commands below. |
| S-004 lifecycle | DONE | Engine, runtime gate, single store and diagnostic session; deterministic cancellation/drain/failure tests plus real thread death and sibling quietness. Evidence-03 (61 tests), command-01, generate-03, drift-02, policy-02, json-02 all pass under `foundation-s004-*`. |
| S-005 readers and removal | DONE | Secure streaming, pure accumulation/analysis, bounded presentation and export-only OTLP; dead helpers and wrapper hook removed. Evidence-02 (63 tests), command-01, generate-01, drift-01, policy-01, json-01 pass under `foundation-s005-*`. |
| S-006 validation and handoff completion | DONE | Generation transaction hardening and routed boundary correction validated; finalizer-02 precedes passing contract-02 and final-source generated/lint gates; final ledger and deferred entry conditions saved. |

T-007/T-008 remain DEFERRED and T-009 remains DROPPED. Overall instrumentation
status stays IN_PROGRESS even when this foundation is complete. F-008's source
waiver correction and exit-13 mapping remain closed unless a regression is shown.

### S-003 checkpoint

Completed `make generate`, `make harness-evidence-contract`, `make generate-drift`,
`make generated-artifact-policy-check`, and `make json-shape-check` using the exact
S-003 run IDs above under `.cartulary/test-results/`. All returned exit 0.
Publication, validation, finalizer and performance consumers now use normalized
facts. Retained state is bounded by units, dependencies, waits and admitted phases,
not constant memory; no second raw stream or compatibility overload remains.
The private accounting API changes atomically with its callers. Duplicate closed
waits previously violated REQ-371; rejection is an intentional correction, not
a retained format migration. S-004 is ready; shutdown races remain its risk.

### S-004 checkpoint

Completed `make harness-evidence-contract` (`foundation-s004-evidence-01`, `-02`,
`-03`, all exit 0; final log reports 61 passing tests),
`make harness-command-surface-contract` (`foundation-s004-command-01`, exit 0),
`make generate` (`foundation-s004-generate-01` through `-03`, exit 0), and
`make generate-drift`, `make generated-artifact-policy-check`,
`make json-shape-check` (each `foundation-s004-{drift,policy,json}-01` and `-02`,
exit 0). Evidence is retained under each exact `.cartulary/test-results/` root.

The controller now shares stop completion and awaits actual worker termination
after normal acknowledgement, failure or the 2,000 ms deadline. The engine
cancels immediately, serializes completion, drains reads/writes and fences, and
prevents stale work crossing counter segments. A failed release cannot acknowledge
pause: transport reports failure and termination removes the OS-thread-owned fence.
Runtime owns conservative admission inspection, 250 ms zero-claim acquisition and
50 ms watcher coalescing. The session closes both diagnostic producers before
retained scanning; inability to prove quiescence remains a required cleanup error.

Private composition changes together; retained identities and accepted payloads
are unchanged. Optional collection health stays independent of canonical outcome.
No container or total-observer qualification claim follows from these tests.
S-005 is ready; secure reader and export boundary parity remain its risks.

### S-005 checkpoint

Completed `make harness-evidence-contract` (`foundation-s005-evidence-01` and
`-02`, exit 0; final 63 tests), `make harness-command-surface-contract`
(`foundation-s005-command-01`, exit 0), `make generate`, `make generate-drift`,
`make generated-artifact-policy-check`, and `make json-shape-check`
(`foundation-s005-{generate,drift,policy,json}-01`, all exit 0).
Retained evidence resides under the matching `.cartulary/test-results/` roots.

Secure resource reads validate stream/schema/order/identity/digest closure before
any aggregate escapes the facade. Pure accumulation uses generated descriptors;
joint CPU windows, gaps, resets, RSS caveats and physical scopes remain explicit.
Pure analysis and bounded formatters have no export dependency. Isolated reader
execution rejects any attempted export-module import; injected transport tests
verify the explicit exporter. Public check fixtures prove exits 0/2/11.

Removed unused private context/reconstruction/finalization helpers, newest-run
fallback and the wrapper's ineffective warning hook. Required-target policy
remains authored. Private callers migrate together, with no aliases or retained
format migration. Export remains explicit, bounded, private and local-resource-free.
The final review must also close generation publication as a coherent transaction
when compilation or filesystem publication fails; this is a remaining S-002
obligation carried explicitly into S-006, not an exception to its exit criterion.

### S-006 validation progress and corrective evidence

`foundation-s006-contract-01` failed `make harness-contract` (Make exit 2,
canonical `child_target_failure`): the new evidence-suite import crossed the
private evidence-accounting owner boundary. All behavioral assertions passed;
the boundary test correctly blocked completion. Corrective action routes the
reducer test file directly through authored Make suite commands, without adding
a private-import exemption. Existing canonical-performance and finalizer fixtures
join those same evidence and broad suite routes so changed consumers stay covered.
The failed run is retained and superseded only by a fresh successful final run.

Final review also repaired generation closure: all definitions and validators
compile before publication through the existing generated transaction. A newly
exposed install-after-backup failure now restores the previous artifact. Isolated
compilation and injected filesystem failure regressions pass in
`foundation-s006-evidence-01` (64 tests); this closes the outstanding S-002
publication obligation.

### Foundation gap disposition and durability review

| Finding | Remediation / affected areas | Rationale and long-term benefit | Compatibility / migration | Risk if reintroduced | Completion evidence |
| --- | --- | --- | --- | --- | --- |
| F-001 | Engine owns control transitions; controller shares terminal completion; implementation/tests | One lifecycle model makes cancellation and future adapters independently testable. | Private composition migrated together; current retained formats preserved. | Early acknowledgements or collection restarting after shutdown. | Admission/read/write interleavings, shared stop, controller failure/timeout and session barrier fixtures. |
| F-002 | Runtime observation gate owns admission interpretation, watcher and OS-thread proof; implementation/ownership/tests | Collection no longer depends on private admission storage; fairness stays with its owner. | Same zero claims, 250 ms acquisition and 50 ms wakeup coalescing. | Measurement interference or fences retained by a dead worker. | Real gate/thread death and four-sibling quietness; failed-release acknowledgement regression. |
| F-003 | Single reduction to normalized facts and shared interval arithmetic; implementation/tests | Publication, validation and performance readers reason over one bounded model. | All private callers migrate; duplicate closed waits intentionally rejected under REQ-371. | Conflicting accounting, duplicate memory and concealed failure/cache facts. | Independent timing/reservation/target/cache/failure fixtures; large selected stream; no raw replay or scheduler import in pure accounting. |
| F-004 | Closed catalog, generated schema components/descriptors and registry-selected validators; specification/contracts/generation/tests | Ordinary approved signals/phases grow through declarations and producers, with fewer drift points. | Existing public schema IDs and accepted payloads preserved; two new declared schema identities. | Producer/reader drift or mixed generated artifacts after failure. | Acceptance corpus, isolated extensions, no-runtime-AJV fixture, generated gates, compile-before-publish and rollback regression. |
| F-005 | Secure reader, pure accumulators/analysis, bounded formatters and export-only OTLP; implementation/ownership/tests | Ordinary inspection has no export construction dependency and fewer private paths to maintain. | Exact public selection preserved; unused newest resolver, context helpers and wrapper finalization hook removed without aliases. Required-target policy retained. | Export coupling, ambiguous selection or insecure artifact handling. | Guarded isolated reader, public exit fixtures, secure malformed/identity/digest/symlink cases, injected export transport. |
| F-006 | Inject clocks, discovery, gate, store and worker transport; tests/implementation seams | Race failures are reproducible without host timing; real integration still proves OS behavior. | Private seams only; no runtime plugin API. | Flaky refactors or untested shutdown failures. | Deterministic fixture suite plus real worker integration through owner-routed Make commands. |
| F-007 | Explicit coverage/qualification deferral; documentation | Foundation remains a stable base without unsupported measurement claims. | Default remains off; no baseline refresh or default-on decision. | Incomplete consumption/observer values being treated as qualified attribution. | T-007/T-008 remain DEFERRED; real basic run demonstrates operation only. |
| F-008 | Reconcile current status while preserving the historical ledger; documentation/regression protection | Prevents revival of source-successor waivers or qualification claims. | Same-source REQ-838 and performance-gate exit 13 retained. | Reusing ineligible historical success or comparing changed source. | Comparison rejection fixtures; prior validation/qualification text verified byte-for-byte against entry copy. |
| F-009 | REQ-840 and AC-137 declare projections, ownership, links and 128-phase rejection; specification/contracts/tests/docs | Limits and authority can be reviewed without inferring requirements from code. | No new production phase/signal or changed cap. | Ambiguous conformance or unbounded phase retention. | Human owner/projection review; 128 accepted/129 rejected; catalog/schema closure; no executable Markdown dependency. |

### S-006 handoff preparation

The supplied Testing Harness, application OpenTelemetry and domain documents
were inspected. Only the Testing Harness owner required amendment; application
OTEL-REQ-148–149 and domain vocabulary are unchanged. Changed implementation
owners are `evidence-accounting`, `observability`, runtime observation admission,
scheduler composition, the node-tool wrapper and generated-artifact publication.
Tests stay routed by the authored command surface rather than private-import
exceptions. New owner-local modules are listed in the final Git diff.

Generated outputs are the instrumentation shared schema, instrumentation runtime
descriptors, standalone foundation validators, task-surface manifest/Make output,
and execution-topology render index. Their authored inputs and exhaustive
generated-root policy are updated together. No lockfile or application dependency
changed. Reverting this work requires the complete caller/input/generated-output
transaction; preserve the original tracker history and all retained runs.

Final-source evidence below confirms that the corrected suite routing passes.
No product-wide, browser, visual or performance-qualification suite is required
by the affected ownership. Those checks remain intentionally skipped; no product
behavior or performance claim follows. Retained-run maintenance remains skipped
because `RESULTS_DIR` is unset and no eligible full warm successful `check` run
was supplied. `agent-finalize` still runs structural/schema/catalog maintenance.

### Final verification ledger

The final implementation baseline is `main` at
`5d80ca3c5e9b21a1b30f4339461fa27377864127`, with authorized uncommitted changes.
Final graph manifests record source digest
`sha256:86e6bd315e140adf9d79cdd68b500554a920e0e98ed759a4909d9d8d067f2901`.
The finalizer, broad contract, lint, drift, policy, shape and real-basic run share
that digest. Earlier successful runs verify their recorded intermediate source;
they are not substituted for the final-source gates. Documentation is excluded
from executable source evidence by the adopted owner boundary.

Every row below uses this invocation prefix, with the row's exact ID and target:

```sh
GO=/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.2.linux-amd64/bin/go \
CARTULARY_PREPARATION_POLICY=installed_only \
CARTULARY_TEST_RUN_ID=<exact-id> make <target-and-arguments>
```

| Exact ID | Target and arguments | Result / retained evidence |
| --- | --- | --- |
| `foundation-s006-generate-02` | `generate` | PASS, exit 0; `generate/tool-run-summary.json` |
| `foundation-s006-evidence-02` | `harness-evidence-contract` | PASS, exit 0; 66 tests; `harness-evidence-contract/tool-run-summary.json` and nested stdout log |
| `foundation-s006-finalize-02` | `agent-finalize` | PASS, exit 0, before broader gates; `run-summary.json`, `unit-artifacts/finalize-summary.json`; generated artifacts unchanged; retained-run actions skipped |
| `foundation-s006-contract-02` | `harness-contract` | PASS, exit 0, 2/2 graph units, 159 passing contract tests; `run-summary.json` and `unit-logs/target-harness-contract/stdout.log` |
| `foundation-s006-lint-02` | `lint-scripts` | PASS, exit 0; `run-summary.json` |
| `foundation-s006-drift-02` | `generate-drift` | PASS, exit 0; `run-summary.json` |
| `foundation-s006-policy-02` | `generated-artifact-policy-check` | PASS, exit 0; `run-summary.json` |
| `foundation-s006-json-02` | `json-shape-check` | PASS, exit 0; `run-summary.json` |
| `foundation-s006-markdown-02` | `lint-markdown` | PASS, exit 0; `adhoc/lint-markdown/tool-run-summary.json`; final completion text receives a final lint pass |
| `foundation-s006-basic-02` | `json-shape-check HARNESS_DIAGNOSTICS=basic` | PASS, exit 0; resource index complete, 15 samples/2 sweeps, graph-entry-to-pre-scan envelope, retained-secret scan pass; operation evidence only |
| `foundation-s006-observation-check-02` | `harness-observability-check RESULTS_DIR=.cartulary/test-results/foundation-s006-basic-02` | PASS, exit 0; direct read-only output: one invocation, four sources, complete diagnostics; creates no separate result root |

Artifact paths in this table resolve beneath
`.cartulary/test-results/<exact-id>/`. `git --no-optional-locks diff --check`
passes; the index remains unchanged. The original validation/qualification ledger
from entry was compared with `/tmp/cartulary-instrumentation-tracker-entry.md`
and remains byte-for-byte present. The retained failed attempts are classified
in their slice checkpoints; negative-fixture failures inside successful contract
suites are intentional assertions, not separate failed validation gates.

### Next entry and residual limits

With S-006 complete, the next work is an explicitly selected T-007 or T-008
scope. Re-read adopted owners and current machine routing, then reconcile any
new phase/signal or post-scan boundary before changing producers. Keep exact run
selection, the same-source comparison rule and exit 13; do not revive T-009.
Qualification requires a complete observer roster and matched off/basic plus
0/1/4-reader windows on eligible source. Current samples remain observed partial
coverage: short-lived processes can be missed, RSS sums can repeat shared pages,
Windows host/container totals are unavailable, and gate/controller/reader costs
are not complete observer accounting. The real-basic check above does not change
those limitations. `HARNESS_DIAGNOSTICS=off` remains the default.

### S-006 final checkpoint

Foundation S-000–S-006 is DONE. F-001–F-006 and F-008/F-009 have implementation
or specification closure and executed evidence; F-007 has an explicit deferred
disposition, not a capability-completion claim. The failed broad boundary run
was corrected through command routing and superseded by
`foundation-s006-contract-02` (PASS). Every workstream checkpoint was saved before
the next workstream began. The final source was reviewed, generated drift passed,
and the required validation/handoff slice is complete. No commit, push, export,
benchmark refresh, product-wide qualification or default-on change was performed.

The final completion-text Markdown gate `foundation-s006-markdown-03`,
`make lint-markdown`, is PASS (exit 0). Its retained summary is
`adhoc/lint-markdown/tool-run-summary.json`. The final
`git --no-optional-locks diff --check` also passes. Overall tracker status
deliberately remains IN_PROGRESS.
