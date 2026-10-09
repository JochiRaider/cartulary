---
doc_id: cartulary.testing_harness.instrumentation.revision_plan
title: Testing harness instrumentation
doc_type: revision_plan
status: IN_PROGRESS
authority_boundary: Human implementation tracker and rationale; the Testing Harness NLSpec owns behavior. No executable tooling consumes this document.
---

# Testing harness instrumentation

## Baseline and authorization

The 2026-10-09 planning inspection and implementation entry inspected `main` at
`354eeda51e49a2c711b27a552823d104fa6fcad1`. The index and working tree were clean;
the branch was two commits ahead of `origin/main`. The implementation request
supersedes the original planning-only boundary. It authorizes owner amendments,
typed projections, implementation and verification, but no optimization,
dependency installation, baseline refresh or implicit telemetry export.

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
Proposed comparisons use exact `COMPARE_RESULTS_DIR` and
`COMPARISON=equivalent|instrumentation`: equal workload/source/graph/toolchain,
capacity/cache/reuse/policy for equivalent; only declared collection treatment and
reviewed source/policy successor differ for instrumentation. Mismatches show
reasons and independent values, never an improvement verdict. One pair is descriptive.

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
Use `harness-performance-check EVIDENCE_ROOTS_FILE=<exact-manifest>` only after
repairing its inspected drift exit mapping (currently 1 instead of 13). No
baseline publication or Core 05 claim follows from exploratory experiments.

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

Implementation is authorized by the user's explicit follow-up. The first delivery
is implemented; this tracker remains **IN_PROGRESS** for the remaining slices.

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
The next delivery should finish observer qualification and lifecycle coverage
before using these observations to justify a separate test-performance change.
