---
doc_id: cartulary.testing_harness.instrumentation.revision_plan
title: Testing harness instrumentation
doc_type: revision_plan
status: IN_PROGRESS
authority_boundary: Human implementation tracker and rationale; the Testing Harness NLSpec owns behavior. No executable tooling consumes this document.
---

# Testing harness instrumentation

## Baseline and authorization

The current user request authorizes implementation of **Complete Go lifecycle
ownership and retire residual harness infrastructure**, starting from clean main
`b0bc9737531c7feaf573e28715348fd81ae0584a`. This supersedes the preceding
planning-only restriction. The sole active chain is **S-036 → S-037 → S-038 →
S-039 → S-040 → S-041 → S-042**. Save a checkpoint after every workstream and
before beginning its successor; a failed exit never counts as DONE.

The prior graph/fixture cutover remains complete in its adopted scope:
S-035, S-034, S-031/A–F and S-032/V-033 remain DONE. Its final executable source is
`0725873db841e6db2a52013948d6982ffc392cf6`, digest
`sha256:1bdb3547e772d5011e1bd7ab66b85264eb5a25fe4198ad860fe06a17c9103307`.
E-040 remains historical final evidence, including 981/981 full-check units.
No new validation is inferred from that pass. Salvage source and evidence remain
untouched; this iteration requires no import, reset, archive or evidence deletion.

Go owns shared private-runtime state, admission, recovery, ordered cleanup,
retained scanning and terminal publication. Graph and review call the same native
fixture owner. Node retains compilation/selection, row runners, browser operations,
report adaptation and offline readers; physical service/browser adapters retain
their distinct operations. This is lifecycle consolidation, not a whole-harness
language rewrite. S-042 is the required completion and handoff gate.

**Functional remediation is complete.** S-036–S-042/V-037–V-043 are DONE, with
individual checkpoints E-042–E-048. Final executable source
`aed9737cbb77683451bdd4dfd47d97b42c3bb1ed` passed the full warm check, 982/982
units, plus final contracts, lint and projection gates. The final tracker checkpoint
changes documentation only. E-048 records source identity, successful evidence,
failed attempts, repairs, skips and migration limits.

Diagnostics remain default-off and local-only. Qualification is DEFERRED;
F-022/RB-006 and S-020–S-022 remain unresolved/deferred. CPU remains the signed,
equally weighted paired mean strictly below 2.5% of one core per cell. All other
budgets and eligibility/stability rules remain unchanged. The experimental ledger
remains 44.952/90 minutes. Functional checks are not qualification observations.

Sections 1–12 contain current state above explicitly historical sections. Preserve
historical IDs, failed attempts, source identities and dispositions.

## Original delivery evidence and gaps — historical

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

## Execution and ownership map — original delivery history

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

## Questions and minimum signals — original delivery history

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

## Selected design and contracts — original delivery history

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

## Change map and implementation sequence — original delivery history

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

## Validation and optimization readiness — original delivery history

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

## Limits and implementation record — original delivery history

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

### Remaining implementation and adoption limits — historical S-013 scope

The foundation implementation is complete. The graph envelope still ends before
the secret scan. Container counters and full Make wrapper/post-scan receipt
coverage remain deferred; no boundary or publication exception was bypassed.
The selected next iteration covers complete harness-owned runner identity,
allocation/process relationships and owner-local migration/reset/report-parsing
timings through S-007–S-013. Cross-clock joins remain deferred.

Real browser collection and controlled four-sibling quietness have historical
validation above. Total observer accounting and zero/one/four public-reader
performance qualification remain outstanding under T-008. These are coverage and
qualification gaps, not claims of unavailable platform support.

`HARNESS_DIAGNOSTICS=off` remains the default. Exact-run readers retain explicit
completeness/coverage limits and do not translate historical schema versions.
Short-lived or detached processes, borrowed services, overlapping RSS pages,
Windows-host pressure and unobserved intervals remain limitations. The selected
iteration neither closes deferred T-007/T-008 work nor supports a performance
improvement or default-on claim.

## 1. Scope and Source Posture

### Current lifecycle iteration scope and source

Artifact mode: normal; implementation authorized. Repository and admitted tracker
are unchanged. Testing Harness REQ-852–855 and REQ-416/502/616 own behavior;
domain remains vocabulary/navigation. Executable checks consume typed inputs,
never this tracker or other Markdown.

| Baseline | Evidence / treatment |
| --- | --- |
| Main and index | Clean `b0bc9737531c7feaf573e28715348fd81ae0584a`; no staged or untracked work at entry |
| Prior planning comparison | 6,917 tracked/nonignored files; aggregate `11c33deb2fe91d940aeaba6e54d3927b48a315bea2db82c4b98b45aceb24c53f`; tracker `a617e7fc008f29f19e709f22365f7954d85a12ddb83c4fd44df4d5b85715852b` |
| Current source | S-036 `439f4665f`, S-037 `f64bc88b7`, S-038 `2ae231627`, S-039 `f15f6c717`, S-040 `081f5b0e6`, S-041 `2b800928c`; S-042 repairs `ab2afb5e0` and final executable freeze `aed9737cbb77683451bdd4dfd47d97b42c3bb1ed`, digest `sha256:6dfb070d1f521a638b2121d285f6f5dd3e39ef20ebe700062933b5dbb010e7e8`; final checkpoint changes only this tracker |
| Evidence | Prior primary/salvage/qualification roots preserved; each writing command uses a fresh explicit run ID |
| Exclusions | Product behavior, compiler/reader rewrite, new daemons, performance qualification, golden refresh and Core 05 publication |

Inspection covers the shared runtime/scanner, graph and fixture consumers, review
controller/session/recovery, standalone callers, authored routing/schema generation
and affected tests. Other browser/product internals are not claimed fully audited.


### Previous functional iteration — current mainline scope and baseline (historical)

Canonical target: `tools/harness/observability`, expanded through graph runtime,
fixture callers, typed projections and owner tests named in Section 7. Sole
controlling tracker: this regular file. Human owner: Testing Harness NLSpec;
`docs/domain.md` remains vocabulary/navigation, with no new domain concept.

| Baseline item | Starting evidence | Preservation / applicability |
| --- | --- | --- |
| Main | `3781d9a45`, clean index/working tree | Implement sequentially; coherent source/projection checkpoints |
| Executable foundation | `229936209`; S-033/V-034/E-031 | Retained behavior, not new validation |
| Salvage candidate | `79aecc1ca` plus existing dirty S-031 | Read-only input; 6,937-entry content fingerprint; no automatic merge |
| Existing evidence | Primary, salvage and qualification retained roots | Preserve failures; no clean/distclean or relabeling |
| Qualification | F-022/RB-006, 44.952/90 minutes | DEFERRED; no new collection |

Discovery covers the named runtime/report/fixture/producer/reader/routing seams;
it is not an exhaustive product audit. Generated outputs are regenerated from
owner inputs, never hand-edited. Baseline candidates are reviewed per workstream.

### Salvage scope — historical off-main salvage

Implementation is authorized for S-023–S-032, with sequential checkpoint gates.
The established target `tools/harness/observability` expands through the graph
runtime and shared fixture callers enumerated below. Source archive admission
captured 6,903 tracked/non-ignored entries and 406 changed entries (198 authored,
208 generated); deleted entries are recorded as absent. This is complete path
accounting, not a claim of complete behavioral inspection of every file.
The durable archive is `/home/jochi/cartulary-instrumentation-archive/20261011T182827Z`.
`source-manifest.json` and `dispositions.json` record exact paths, hashes,
authorship and per-file dispositions. `evidence-manifest.json` covers all primary
retained results, retained-work, qualification-checkout results, matching RAM-disk
attempts and available temporary freeze/analysis journals. Runtime caches are
reproducible inputs, not retained experiment evidence; their contents are not
claimed archived. Existing complete source history is in `source-history.bundle`.
No original evidence or source was removed. E-029 records preservation checks.

### Prior iteration — historical scope and evidence


**Selected iteration: complete diagnostic coverage and qualify observer cost.**
Artifact mode: `normal` for the bounded recovery decision plan in Section 7.
The latest invocation is planning-only; the earlier implementation authorization
does not require continuing experiments while the user asks to step back.
Owner amendments precede typed projections and production changes. Root-cause
confirmation and any resulting code movement remain future work, not discoveries
claimed complete by this plan.

Repository root is `/home/jochi/code/cartulary`. The established literal target is
`tools/harness/observability`, a directory with canonical repository-relative
identity `tools/harness/observability` and label `harness-instrumentation`.
Coupled runtime, wrapper, suite-service, producer, schema and routing seams are
listed in Section 2. The sole output is
`docs/handoffs/cartulary-test-harness-instrumentation.md`; the file and every parent
resolve inside the repository without symlinks. The existing file is regular and
has one hard link. Its declared target matches the requested update.

Planning follows [the repository framework](cartulary_modular_refactor_planning_framework.md)
and the refactor-tracker format. Core 00 §§2/5 allocate authority. The Testing
Harness NLSpec is `adopted/current`, `cartulary.testing_harness.v3`; REQ-830–844
and REQ-377–380 govern existing diagnostics and performance mechanics. C-010–C-014
are adopted through S-014, REQ-845–849 / AC-142–146. REQ-836/840/844 now
name the bounded receipt exception; implementation and qualification remain pending.
Core 05 remains relevant only to separately authorized claim-bearing publication.
No primary-owner contradiction was found; extending an explicit current limit
requires an amendment, not treating the tracker as authority.

The selected work closes the actionable T-007/T-008 remainder: exact-container
counters, clock relationships, wrapper/scan/return coverage, complete declared
observer accounting, and 0/1/4-reader qualification. Native Windows-host counters,
Docker-daemon-internal cost, complete detached/short-lived-process discovery,
automatic export, application telemetry, scheduler/concurrency redesign, baseline
refresh and Core 05 claims are outside this local WSL2 qualification. Residual
coverage limits remain explicit even after the selected work passes. Default-off
is a decided policy, not an unresolved default-on milestone.

The following baseline table and original 30-file inventory describe the earlier
planning pass, not the current dirty implementation checkout. E-026–E-028 and the
systematic recovery checkpoint supply the current source/evidence posture.

| Baseline item | Starting evidence / identity | Final comparison / outcome | Limits |
| --- | --- | --- | --- |
| Branch / commit | Clean `main`, `1d423c6e5f9cd204f8210fa9958c780fd933d293` | E-019/E-025 | Earlier `9359eb0`, `5d80ca3` and original-delivery runs retain their own source limits |
| Index / staged changes | No staged changes; index SHA-256 `3daa7809ae7a8bfa449e9995e53376e8dcc53320ef8855764a34e151914e2584` | E-025 compares raw index bytes; Git optional writes disabled | No staging or commit in this update |
| Working tree / untracked | Clean; 6,608 tracked/non-ignored entries captured with SHA-256 and lstat identity; no non-ignored untracked files | E-025 compares all entries, excluding only the admitted tracker | Comparison material is outside the repository |
| Tracker | SHA-256 `1a883ccc54af844aa82e124a44dee4d1ce02d20d2b26753eba500cf47895928e`; regular, one hard link | Recheck identity/content before write and compare history afterward; E-025 | Only permitted repository change |
| Ignored outputs | Existing `.cartulary/test-results`, `.cartulary/retained-work`, `.cache`, `node_modules`, `.cartulary` | Root identity/timestamp comparison; no artifact-producing command run | Contents are not recursively hashed; no current runtime pass claimed |

The prior planning evidence was revalidated at the same clean commit. The target
universe is tracked plus non-ignored untracked entries: 30 tracked authored regular
files, with no symlink, submodule, binary, vendor or generated entry under this
target. Ignored caches, dependencies and retained outputs are excluded from source
inventory. This is a bounded subsystem inspection, not a whole-repository audit.

## 2. Current-State Repository Inventory

### Current lifecycle inventory and ownership

| Inspected seam | Current responsibility / callers | Disposition / finding |
| --- | --- | --- |
| `internal/testutil/harnesssession/{runtime,preservation,protocol,tree,lock,physical}_linux.go`; Node `runtime/session-client.mjs` | Shared native private roots, secrets, exact journals, janitor and bounded private transactions for producer adapters | Extraction/caller migration E-043–E-045, deletion E-046; F-037/F-039 |
| `internal/testutil/harnesssession/scan_linux.go`; graph `scan_linux.go` publication adapter | One native scanner shared by graph and review; distinct publication ordering | Policy/digest and retained schemas preserved; F-037 |
| `internal/testutil/harnessreview`; `browser/ui-review/{native-client,adapter,ownership,diagnostic-processes,input-files}.mjs`; `browser/review-recipes.mjs` | Go registry, admission, cancellation, recovery and terminal authority; Node retains browser/process operations, input compilation and preparation recipes | Direct native fixture library, no parallel Node coordination; E-044/E-046, F-038 |
| `internal/testutil/harnessfixtures` | Direct shared allocation/lease/reuse/recovery library for graph/review | Fixture-session transport, subcommand and private schema removed; F-038/F-040 |
| `execution/run-make-node-tool-cli.mjs`, `smoke/run-harness-smoke-cli.mjs`, `readiness/frontend-artifact.mjs`, `services/local-session.mjs` | Distinct Node producers/physical adapters using native runtime authority | Migration E-045, physical public checks E-047; F-039 |
| Native session/review/fixture/runtime tests; retained Node compiler/reader/producer/browser tests | Independent expected lifecycle/security facts and actual owner execution; no JavaScript scanner oracle | Assertion migration ledger closed at E-046; F-040 |
| Task surface owner, helper ownership, test families, schema attachments and native schema generator | Authored executable routing and generated projections for the remaining owners | Generated and checked in E-046; no handwritten generated roots |
| Active browser design/readiness guide and source-scoped historical handoffs | Current owner navigation and preserved historical evidence | Obsolete LIVE claims corrected and historical scope explicit; E-046, F-041 |

Removed implementations are listed in E-046 and the historical inventory, not as
current owners. Retained Node adapters have physical, compilation, producer or
reader responsibilities; none owns shared runtime or review-session policy.

### Previous functional iteration — current mainline disposition and navigation (historical)

The preserved S-033 baseline per-path ledger is
`/home/jochi/.codex/worktrees/instrumentation-main-cleanup/cartulary/.cartulary/retained-work/main-cleanup/disposition.json`.
It records initial and candidate hashes: 227 paths retained byte-for-byte, 132
updated (including generated projection changes), 30 experimental files removed,
16 existing deletions retained, and this tracker reconciled separately. To inspect
the exact committed result use `git diff --no-renames 1d423c6e5 229936209`;
`git diff --no-renames c68c8cfa4 229936209` isolates cleanup of the dirty baseline.
Do not treat generated symbol renumbering as a new behavioral feature.

| Path / component | Main disposition and responsibility | Validation / future boundary |
| --- | --- | --- |
| `tools/harness/test-catalog/`, `tests/contract-test-context.mjs`, `test-catalog-metadata-fixtures.mjs`, `execution/runners/row-runner-cli.mjs` | Retain metadata-only case fixtures, fresh selected-row validation, full public preflight and bounded hashing | Command/evidence tests; no source ASTs or broad catalog needed by metadata-only fixtures |
| `scheduler/work-graph/cache*.mjs`, `compiler.mjs`, `browser.mjs`; `readiness/cache-artifact.sh` | Retain shared catalog access, disabled-cache bypass and bounded hash batches | Cache invalidation, independent digest oracle and unsafe/current-input tests |
| `contract/`, `generated-artifacts/*schema*`, generated Node/native validators | Retain one generated acceptance source, strict JSON and secure immutable retained loading | Parity corpus, duplicate keys, hostile files, one resource traversal; no second schema registry |
| `internal/testutil/harnessruntime`, `harnessfixtures`, `harnessschema`, `harnessadmission`, `harnesscollect`, `harnessdiag`, `harnesslocal`; `tools/harnessruntime` | One installed graph binary with cohesive libraries; graph and UI-review share fixture policy | Native race/schema/physical gates; separate helper composition roots deleted |
| `observability/`, `runtime/`, workspace admission and suite-service lease proof | Retain clocks, resources, relationships, envelope/receipt, reader/security/lifecycle behavior | Current retained schemas and publication ordering preserved |
| `observability/qualification/{analysis,measurement,policy,trace,validator}.mjs` and validation fixtures | Retain independent evidence checking and authored qualification policy | No collection-only exports or launch machinery; CPU equality at 2.5% fails |
| `tools/harnessmeasure`, `internal/testutil/harnessmeasure`; qualification `runner`, `runner-cli`, `readers`, `gate-client`, `gate-helper-cli`, `diagnostic`, `diagnostic-cli`, `diagnostic-policy`, `detail`, `lifetime` | Removed from active main; recover only from existing Git if an explicitly adopted future need arises | No active import, Make target or build reference; no source archive created |
| `runtime/memory-lifetime*.mjs`, lifetime test hooks/tests; diagnostic-plan/manifest v2, measurement-detail/memory-lifetime v1 and generated validators | Removed; case-scoped fixture repair retained | Existing historical evidence keeps its matching source/readers; no translator |
| `internal/testutil/s3test` | Retain independent single-master fixture fix | Real service/browser runs plus focused fixture test; no instrumentation performance claim |
| Node graph scheduler, cache runtime, live publisher and JS fixture broker | Removed in S-031/E with migrated callers | Keep Node compiler/cache-plan facts, row runners and offline readers; no fallback |
| Go graph/fixture candidate and dirty S-031 test changes | Reviewed integration through A–E; duplicated candidate simulations rejected | Section 7 test dispositions and current-source gates supersede historical passes |
| `.markdownlint-cli2.jsonc`, deployment marker | Retain tracker lint coverage and authored deployment layout | Markdown lint and release-evidence contract respectively |


### Salvage disposition ledger — historical off-main salvage

Every changed path is assigned in the external `dispositions.json`; generated
membership is derived from `tools/generated_artifact_policy.json`. The following
transactions define the meaning of those per-file assignments. Mixed files retain
their nonexperimental behavior and lose only the named hooks/registrations.

| Transaction | Exact family / boundary | Disposition and gate |
| --- | --- | --- |
| Retained foundation | Metadata-only catalog access, selected-row validation, case-scoped fixtures; shared/precompiled validation, strict JSON, secure retained loading; bounded source hashing, cache-off bypass, readiness hashing, process/file ownership and cleanup notifications | RETAIN through S-025/V-026; preserve full public preflight and product/owner-required assertions; reorganize harness-internal tests by Section 7 |
| Service correction | `internal/testutil/s3test/s3test.go` and its focused test | RETAIN separately, without attributing unrelated observer savings |
| Retained observations | Execution record/index v2, resource sample v2/index v3, envelope v2, explanation v3, instrumentation policy v3, terminal receipt v1; clock placement, container proof, local activities and relationships | RETAIN and integrate in S-030; no acceptance-set changes or historical translators |
| Native libraries | `internal/testutil/harnessadmission`, `harnesscollect`, `harnessdiag`, `harnesslocal`, suite-service lease proof | RETAIN as cohesive libraries; consolidate composition roots |
| Node preparation and rows | Compiler/model/selector/source discovery; `tools/harness/execution/runners` and runner report adapters | RETAIN; compiler exits before execution; remove only lifetime probes |
| Graph runtime | Work-graph scheduler, runner composition, executor, cache runtime and live publication | REPLACE in Go; REMOVE superseded runtime after S-031/V-032, retaining compiler/model/offline analysis |
| Shared fixtures | `tools/harness/scheduler/fixture-broker`; graph plus `tools/harness/browser/review-preparation.mjs` and UI-review cleanup callers | REPLACE once in Go; remove JS policy only after all callers migrate in S-029/S-031 |
| Admission/bridges | Bash workspace admission; Node native-collector bridge and graph diagnostic-session orchestration | RETAIN narrow bootstrap-safe Bash wrapper; REPLACE graph bridges with Go owner, preserving receipts |
| Native entrypoints | `tools/harnessadmission`, `tools/harnesscollect`, `tools/harnessdiag` | REPLACE with `tools/harnessruntime`; remove old entrypoints/build variables only after cutover |
| External meter | `tools/harnessmeasure`, `internal/testutil/harnessmeasure` | ARCHIVE then REMOVE in S-026; archive includes matching source/tests/results |
| Experiment modules | Under `tools/harness/observability/qualification`: `runner`, `runner-cli`, `readers`, `gate-client`, `gate-helper-cli`, `diagnostic`, `diagnostic-cli`, `diagnostic-policy`, `detail`, `lifetime` | ARCHIVE then REMOVE in S-026 |
| Independent validation | Same directory: `analysis`, `measurement`, `policy`, `trace`, `validator`; qualification plan v3, manifest/result v2 and measurement trace v1 | RETAIN; remove only collection-only exports; preserve `harness-performance-check` |
| Experimental formats/hooks | Diagnostic plan/manifest v2, measurement-detail v1, memory-lifetime v1; `runtime/memory-lifetime.mjs`, `memory-lifetime-producer.mjs`, lifetime imports/calls/tests | ARCHIVE then REMOVE from active source/attachments; archived reader remains version-pinned |
| Mixed tests | `observability/tests/test-qualification.mjs` | SPLIT: retain independent arithmetic, eligibility, integrity, stability and validator cases; archive collection and meter-lifecycle cases |
| Public surface | `harness-instrumentation-qualify`, `harness-instrumentation-diagnose` | RETIRE through S-024 owner amendment and S-026 authored projection; no alias or general runner |
| Documentation/history | Tracker Markdown coverage; rejected Worker variant, failed attempts and superseded experiments | RETAIN lint coverage and historical ledgers; rejected code remains ARCHIVED only |
| Unrelated content | Staged tracker patch and `deploy/mvp/runtime-roots/tmp/.keep` | PRESERVE unchanged; no blanket reset/stash/clean or unrelated staging |

### Prior iteration — historical scope and evidence


The 30-file table below is the original planning inventory. Its removed worker,
engine and store paths are historical, not an instruction to recreate them.
Current implementation deltas are recorded in the execution ledgers. The recovery
planning pass directly inspected the following bounded measurement seams; it
does not claim a fresh exhaustive caller audit.

| Current recovery surface | Inspection / accounting | Responsibility and decision boundary |
| --- | --- | --- |
| `qualification/{analysis,measurement,runner,policy,validator}.mjs` | Five files directly read | Acquisition/summary/validation separation, complete matrix, current arithmetic, identity and CPU plan; retain owners, extend only proven evidence gaps |
| `qualification/{readers,gate-client,runner-cli,gate-helper-cli}.mjs` | Four metadata-only entries from enumeration | No movement proposed; inspect exact entry/control callers in S-020/A before extending experimental routing or changing supervision |
| `tests/test-qualification.mjs` | Directly read | Independent arithmetic, sample semantics, bound/eligibility, manifest and death fixtures; tests were inspected, not executed |
| `internal/testutil/harnessmeasure/{main_linux,proc_linux,measurement_linux_test}.go` | Three files directly read | Native child supervision, process identity, resident-page acquisition, independent CPU/adoption fixtures; S-020/B owner |
| `internal/testutil/harnessmeasure/main_unsupported.go` | Metadata-only entry | No change proposed to unsupported-platform behavior |

This bounded inventory contains 14 entries: nine directly read and five
metadata-only. Selected graph-runner compilation/cache/execution seams, authored
qualification plan, exact test-family routing, Make recipes, Markdown wrapper and
configuration were also inspected as dependencies. The proposed graph-lifetime
path remains blocked on retaining-path evidence and a complete consumer review;
no uninspected code is approved for movement. All other observability source is
outside this incremental behavioral inspection, retaining its prior ledgers and
requiring reinspection if a selected repair reaches it. Ignored candidate-16
results are explicitly included as evidence, not executable source or new tests.

Paths below are relative to `tools/harness/observability`. There are **26 directly
inspected entries and four metadata-only entries**. Direct means the named seam
or assertions were opened during the same-source planning/update passes, not a
whole-file security audit. Metadata-only entries have no proposed behavioral
movement; inspect them directly before any newly necessary change. All entries
are tracked authored regular files. Presence of a fixture is not a test pass.

| Path | Disposition / depth | Responsibility / action | Callers, dependencies, tests / contracts | Risk | Evidence |
| --- | --- | --- | --- | --- | --- |
| `canonical-evidence-audit-cli.mjs` | Metadata-only | Canonical audit entry; unchanged | Audit routes; C-001 | low | E-019 |
| `canonical-evidence-drift-suite-cli.mjs` | Metadata-only | Canonical drift entry; unchanged | Audit routes; C-001 | low | E-019 |
| `canonical-evidence.mjs` | Direct: identity/reconstruction seam | Validated canonical facts; preserve | Read facade, event reducer; C-001 | high | E-020 |
| `canonical-performance-cli.mjs` | Direct: dispatch/exits | Separate typed qualification-manifest dispatch | Public performance target; C-005/C-013 | high | E-024 |
| `canonical-performance.mjs` | Direct: eligibility/statistics seam | Keep canonical portfolio; separate instrumentation policy | CLI, canonical performance fixtures; C-005/C-013 | high | E-024 |
| `collection-engine.mjs` | Direct | Sampling cadence, fencing, pause/stop, observer facts | Worker, adapter, gate/store; C-002/C-011/C-013 | high | E-021/E-023 |
| `diagnostic-session.mjs` | Direct | Drain producers and publish pre-scan evidence | Graph runner, execution session; C-002/C-012 | high | E-022 |
| `execution-analysis.mjs` | Direct | Bounded launch/allocation/activity view | Performance analysis; C-010/C-014 | medium | E-020/E-021 |
| `execution-reader.mjs` | Direct: schema/reference closure | Secure execution-index validation | Private projection, canonical leases; C-009/C-010/C-014 | high | E-020/E-021 |
| `observability-check-cli.mjs` | Direct | Exact-run checking currently repeats resource loading | Read facade, command fixtures; C-004/C-014 | high | E-020 |
| `observability.mjs` | Direct | Selection, loading, explanation; envelope checked only here | Canonical/execution/resource readers; C-014 | high | E-020 |
| `otel-export-cli.mjs` | Direct: read-facade/import boundary | Keep explicit exporter; adapt loader consumer only | Facade, OTLP projection, transport tests; C-004 | medium | E-020 |
| `otlp-projection.mjs` | Metadata-only | Export-only projection; no signal expansion | Explicit exporter; C-004 | low | E-019 |
| `performance-analysis.mjs` | Direct | Pure phase/target analysis and optional diagnostics | Normalized facts; C-001/C-010/C-012/C-014 | high | E-020 |
| `performance-presentation.mjs` | Direct | Shared bounded human/JSON formatting | Explanation schema; C-004/C-014 | medium | E-020 |
| `resource-accumulator.mjs` | Direct | Segmented counters, RSS and attributed lower bounds | Generated signal descriptors; C-003/C-011/C-013 | high | E-020/E-023 |
| `resource-collector-worker.mjs` | Direct | Worker composition and bounded control transport | Engine, gate, adapter/store; C-011/C-013 | high | E-023 |
| `resource-collector.mjs` | Direct | Controller, registrations, acknowledgements and shutdown | Graph composition, private process proof; C-002/C-011/C-013 | high | E-023 |
| `resource-linux.mjs` | Direct | Process/guest/cgroup adapter; no container scope | Linux counters/proof; C-003/C-011 | high | E-021 |
| `resource-presentation.mjs` | Direct | Scope/observer presentation | Validated aggregates; C-011/C-013/C-014 | medium | E-020 |
| `resource-projection.mjs` | Direct | Compose one secure read and bounded aggregation | Facade/check; C-014 | medium | E-020 |
| `resource-reader.mjs` | Direct | Bounded stream, digest, identity and attribution closure | Secure local files; C-003/C-011/C-014 | high | E-020 |
| `resource-store.mjs` | Direct | Single sample writer and terminal resource index | Engine, atomic publication; C-003/C-011/C-013 | high | E-023 |
| `run-comparison.mjs` | Direct | Strict same-source descriptive comparison | Explanation; C-005/C-013 | high | E-024 |
| `tests/test-canonical-performance.mjs` | Direct: relevant assertions | Canonical roster and statistics expectations | Performance policy fixtures; V-021 | medium | E-024 |
| `tests/test-collection-engine.mjs` | Direct: control scenarios | Injected interleavings, pause/drain/failure cases | Evidence-contract route; V-015/V-018/V-020 | high | E-023/E-024 |
| `tests/test-execution-index.mjs` | Direct: security scenarios | Closure, corruption, omission and late-write cases | Evidence-contract route; V-016/V-017 | high | E-020/E-024 |
| `tests/test-foundation-characterization.mjs` | Direct: selected assertions | Independent accounting, lifecycle and local timing fixtures | Evidence-contract route; V-015/V-017 | high | E-021/E-024 |
| `tests/test-instrumentation.mjs` | Direct: selected assertions | Resource security, worker integration, reader/export parity | Evidence-contract route; V-015/V-016/V-018/V-020 | high | E-020/E-023 |
| `tests/test-observability.mjs` | Metadata-only: fixture inventory | Existing synthetic observability coverage; preserve | Evidence/command routes; C-001/C-004 | medium | E-019/E-024 |

The following coupled boundaries were directly inspected at their selected seams.
Searches used `rg` plus exact reads and authored catalogs; they do not establish
absence of dynamic consumers. S-015 completes the executable caller/route roster.

| Coupled seam | Responsibility / disposition | Evidence / future gate |
| --- | --- | --- |
| `runtime/{boot-clock,local-activity,execution-observations,secure-local-files}.mjs`; `internal/testutil/suiteservices/{activity,lifecycle_clock}.go` | Runtime clock identity/private transport; owner-local timing and bounded retained facts | E-021; S-017 anchors without redefining durations |
| `scheduler/fixture-broker/{providers,cleanup-lifecycle}.mjs`; `tools/testservices/main.go`; suite-service private ledgers | Exact ownership and existing Docker client; lifecycle stays with owners | E-021; S-018 adds read-only adapter and proves cleanup parity |
| `workspace/admission.{mk,sh}`, `workspace/child-process.mjs`; `scheduler/work-graph/{runner-cli,scheduler}.mjs` | Workspace lock inheritance, managed launch, canonical epoch and pre-scan closure | E-022; S-019 supervision and receipt boundary |
| `runtime/{observation-gate,host-admission,lease-lock}.mjs` | Observation fences and short-lived admission helpers | E-023; S-020 counts cost without changing fairness/claims |
| Definitions/policy, envelope/resource/execution schemas, schema attachments and generated-artifact policy | Authored machine inputs and generated consumer ownership | E-021/E-024; S-014 declares cutovers, later slices generate atomically |
| `tools/task_surface_owner.json`, selected `tools/test_families/*.json`, verification inputs, Markdown wrapper/config | Public command contracts, routed owner rows, documentation-check side effects | E-024; S-015 freezes routes; no Make command executed here |

Production application packages, SQL, UI components/selectors, application OTel,
HTTP/WebSocket APIs and product authorization have no planned movement. Browser
work concerns harness service/clock/cleanup support, not visual design or golden
changes. The next discovery boundary is S-015's exact dynamic wrapper/helper and
qualification workload roster, under the already specified owners and gates.

## 3. Module Boundary Diagnosis

### Current lifecycle architecture

`harnesssession` sits below graph, fixture and review coordination and depends on
secure local files, schema validation and process identity. It must not import
graph or fixture orchestration. `harnessreview` owns review lifecycle and calls
`harnessfixtures` directly. The installed harness binary remains a thin composition
root. Node adapters report facts/results and execute their specialized producer
operations; they do not independently decide session terminal transitions.

No generic command runner, plugin registry, engine selector, permanent legacy
transport or background service is introduced. Borrowed resources retain their
physical owner. Persistent service proof survives short-lived commands.


### Previous functional iteration — current diagnosis and architectural choice (historical)

F-022 did not establish a single shell RSS root cause. The fixture lifetime repair
is useful independently; the rejected worker-thread variant increased measured
row-runner RSS. Repeated matrix retries and more observer machinery did not prove
repeatability or meet the fixed gates. Preserve those conclusions without turning
failed measurement into a reason to discard working observation features.

| Option | Benefit / cost | Decision and evidence limit |
| --- | --- | --- |
| Keep the former Node runtime indefinitely | Least migration risk; retains resident source/compiler state and multiple helper composition paths | Historical baseline; replaced by the validated native boundary |
| Short-lived Node preparation plus Go graph/fixture coordination | Explicit preparation lifetime; reuse current native ownership/admission/collection libraries; remove duplicate coordination paths | Implemented through S-031/A–F using reviewed candidate inputs and fresh proof. Smaller RSS or CPU remains unproven |
| Short-lived preparation plus a compact Node coordinator | Could remove resident catalog state with less language migration | Not selected; no fallback engine or runtime selector is retained |
| Rewrite compilers, row runners, adapters or all tests in Go | Larger language uniformity, but duplicates mature source-aware and runner-specific behavior | Rejected for this delivery: high migration cost without demonstrated benefit |
| Add a daemon, profiler framework, RAM-disk dependency, forced GC or allocator tuning | Changes operating assumptions and adds another subsystem to diagnose | Rejected; storage is not a CPU-budget remedy and no new collector is needed for F-032 |

Prioritize useful answers: preparation versus execution versus finalization,
queue/admission/fixture waits, completion-delaying units, and observed consumption
with explicit scope and coverage. Consolidation must reduce active owners, resident
state and supported entrypoints, not merely move their complexity into Go.


### Salvage ownership — historical off-main salvage

Go graph runtime owns scheduling, admission, cache transactions, launch/wait/cancel,
canonical publication and finalization ordering. Node prepares validated immutable
facts and exits before graph work; Node row runners and offline analysis remain.
`internal/testutil/harnessfixtures` owns shared fixture policy for graph and
UI-review. Service/browser lifecycle adapters remain at their existing owners.
The isolated collector remains a run-owned child where required for fault
isolation and quietness, using the same installed binary. There is no daemon.

### Prior iteration — historical scope and evidence


The target is an observability subsystem with collection orchestration, adapters,
secure readers, pure analysis, formatting and explicit export. Retain those
cohesive boundaries. The needed structural correction is a shared validated read
result; remaining changes add facts through existing runtime and lifecycle owners.
No second scheduler, diagnostic daemon, generic plugin layer or application SDK
is justified.

| Responsibility | Current location | Owner / keep, move, split | Evidence | Finding / contract |
| --- | --- | --- | --- | --- |
| Retained validation | Facade, resource projection and checker | Split one validated loading result from analysis/formatting; remove duplicate reads | E-020 | F-018 / C-014 |
| Local clocks and placement | Runtime/Go producer clocks; local activities | Keep measurements at owners; add runtime anchors and pure validated joins | E-021 | F-014 / C-010 |
| Container identity/lifecycle | Suite-service private leases and helper | Keep ownership; add exact-owner read adapter, driven by existing collector cadence | E-021 | F-015 / C-011 |
| Outer completion | Workspace admission and graph pre-scan session | Keep lock/exit authority; supervise child and add narrowly adopted terminal receipt | E-022 | F-016 / C-012 |
| Observer cost | Worker self-counters and external historical experiments | Keep self-counters explicitly scoped; add complete declared external measurement and typed qualification | E-023/E-024 | F-017 / C-013 |
| Canonical accounting and export | Evidence-accounting and explicit exporter | Keep; no new canonical clock or automatic export | E-020/E-024 | C-001/C-004/C-005 |
| Current/historical plan | This tracker | Replace current summaries; preserve completed histories and evidence limitations | E-019/E-025 | F-019 / V-023 |

## 4. Public Contract and Behavior Freeze Map

### Current lifecycle contract map

| Contract | Owner and required behavior | Validation |
| --- | --- | --- |
| C-022 private runtime/scanner | REQ-852/855: one native root, secret, journal, stale cleanup and scan owner; ownership proof before deletion; unresolved proof survives | V-038/V-040 |
| C-023 review coordination | REQ-853 and REQ-416/502/616: Go admission/registry/recovery/terminal owner; exact dependency order, primary cause, immutable terminal repetition | V-039/V-042 |
| C-024 private interfaces/public freeze | REQ-854: 32 MiB graph plan, 1 MiB private frame, existing 64 KiB UI request; strict schema/reference/proof binding; no arbitrary commands or secret argv | V-037–V-042 |
| C-025 deletion and evidence | REQ-855: independent expectations, no old policy/oracle or active compatibility forwarding; retained schemas and public behavior preserved | V-041/V-043 |

Graph scanning follows producer quiescence with only its bounded outer receipt
permitted afterward. UI terminal publication follows its own REQ-502/616 protocol.
Public commands, selection, modes, failure precedence, deadlines and retained
acceptance sets remain stable. No product API or domain concept changes.


### Previous functional iteration — current adopted contracts and validation boundary (historical)

| Contract | Owner / expected behavior | Disposition and gate |
| --- | --- | --- |
| C-010–C-014 | NLSpec REQ-845–849: clock uncertainty, exact container proof, bounded receipt, unchanged qualification arithmetic and one validated retained model | Retained; fresh native/public evidence E-035–E-039 preserves acceptance and canonical timing |
| C-017 | REQ-848/850/851 distinguish functional delivery from qualification; retire experimental acquisition | Adopted on main; validation remains active, certification deferred |
| C-015/C-016/C-018 | Adopted REQ-852–855: one Go graph owner, shared fixtures and private interfaces | Adopted in S-035; S-031/A–F implementation and validation complete |
| C-019 scope-aware presentation | REQ-849 and existing resource/analysis semantics: display validated scope-specific facts without new filesystem reads or invented totals | DONE S-034/V-035/E-033; fresh native report parity E-039; retained schema unchanged |
| C-020 mainline cleanup | Completed prior cleanup and preserved evidence | S-033/V-034/E-031 historical completion |
| C-021 coherent current specification | REQ-833/835/840/844/849/852–855; schema versions, availability, reports and native ownership | DONE S-035/V-036/E-032; human-reviewed typed projections, no Markdown executable dependency |

Current formats remain execution record/index v2, resource sample v2/index v3,
envelope v2, explanation v3, instrumentation policy v3 and receipt v1. Qualification
plan v3, manifest/result v2 and measurement trace v1 remain validation inputs.
Removed versions are rejected, never translated. Credential/private proof fields
must not cross into retained diagnostic output. Domain vocabulary needs no new
concept; HTTP/product/UI behavior remains outside this harness remediation.


### Salvage contracts — historical off-main salvage

| Contract | Owner amendment / preservation | Validation |
| --- | --- | --- |
| C-015 graph runtime | Go graph execution, unchanged scheduling/failure/cache/canonical semantics; Node preparation and row execution retained | S-027/S-028/S-031; V-028/V-029/V-032 |
| C-016 shared fixture policy | One Go lifecycle policy, direct graph calls and narrow UI-review client; owned/borrowed cleanup and recovery unchanged | S-029; V-030 |
| C-017 deferred certification | Preserve REQ-848 budgets and independent validation; retire collection commands and experimental obligations without promoting failed evidence | S-024/S-026; V-025/V-027 |
| C-018 private interfaces | Runtime plan v1, 32 MiB, inherited private descriptor; fixture session v1, 1 MiB frames, closed acquire/release/recover/close with identity and EOF cleanup | S-024/S-027/S-029; V-025/V-028/V-030 |

These are adopted by the authorized S-024 specification transaction before their
machine projections. C-001–C-014 remain required except the explicit collection
deferral/retirement. Public selection/results and retained schemas are preserved.

### Prior iteration — historical scope and evidence


Existing owner-aligned contracts remain constraints. Existing defects do not
become compatibility promises. C-010–C-014 were adopted and implemented at their
recorded checkpoints; proposal/baseline observations in the original mapping
below retain that historical context. C-013 now has the explicit recovery entry
and pending user-authorized CPU revision. Required corrections stay separately
named and tested.

### Adopted foundation and identity contracts — retained IDs

| Contract / surface | Adopted behavior / current observation | Disposition and verification |
| --- | --- | --- |
| C-001 canonical timing | REQ-283, 372, 831, 840: sole scheduler timeline, whole-unit lifetimes, normalized accounting, queue-inclusive paths and bounded phases | Preserve; local activities never modify canonical durations/critical paths. V-010/V-012; foundation closure in S-006 |
| C-002 lifecycle/quietness | REQ-830, 833, 836, 840: optional diagnostics, bounded fences/stop, independent health, quiescence before scan | Preserve; new producers must join the same pre-scan closure obligation. V-009/V-011/V-012 |
| C-003 resources/retention | REQ-832–834, 839: physical proof, opaque identities, scope separation, bounded secure artifacts, incomplete data disclosed | Preserve; no extra per-row attribution from invocation membership. V-009/V-011 |
| C-004 public readers/export | REQ-835–836 and OTEL-REQ-148–149: exact runs, bounded human/JSON data, explicit export/privacy/failure policy | Preserve arguments/exits and export isolation; adopt changed explanation shapes explicitly. V-011 |
| C-005 comparisons/qualification | REQ-377–380, 838: same-source descriptive comparison, separate performance qualification, gate exit 13 | Preserve; no successor waiver, baseline refresh or claim. V-011/V-012; T-009 stays DROPPED |
| C-006 invocation context — adopted extension | REQ-841: neutral fresh launch attempts, explicit parent/unit/rows and outcomes; private command-failure v1 consumes that identity with exact validation and parent-owned cancellation | S-007 adopted neutral identity, causal parent, unit/row membership and attempt outcomes; failure envelope remains private and validated. F-010; V-007/V-008 |
| C-007 correlation — adopted extension | REQ-842: independent allocation attempts, successful leases, immutable physical proof and explicit relationship provenance/conflicts | S-007 adopted explicit relations and incomplete/conflicting evidence rules; fixture ownership/reuse/cleanup remain fixed. F-011; V-007/V-009 |
| C-008 owner-local activities — adopted extension | REQ-843: owner-local migration/reset/parsing activities with invocation/allocation linkage, nullable local clocks and unsuccessful/incomplete outcomes | S-007 adopted owner-local activity inputs/outputs, clocks, outcomes and gaps without clock joins. F-012; V-007/V-010 |
| C-009 retained execution diagnostics — adopted extension | REQ-844: versioned terminal execution index, bounded private publication, secure closure and pre-scan quiescence | S-007 adopted the new projection and schema identities, bounds, relationships, availability and public failure handling before generation. F-011/F-012; V-007/V-011 |

The V-001–V-013 references above belong to earlier execution. Current regression
obligations are V-015–V-022. C-002/C-009 currently require pre-scan producer
quiescence; C-012 proposes only a separate structural receipt exception, not a
blanket permission for post-scan writes.

### Coverage and qualification contracts

| ID / surface | Observed behavior / evidence | Owner and proposed required behavior | Disposition / prerequisites | Risk / finding / slice / validation |
| --- | --- | --- | --- | --- |
| C-010 clock relationships | Identified Go/Node local clocks and 10 ms helper boot clock, no anchors; E-021 | Testing Harness REQ-837/843: bracketed anchors, bounded offsets/uncertainty, placement unavailable on incompatible identity/discontinuity; retain local duration | Proposed extension; S-014 adoption and S-015 independent fixtures before producer changes | False precision; F-014; S-017 / V-017 |
| C-011 container observations | Private exact-container lease proof; process/guest/cgroup samples only; E-021 | REQ-832–834/839/842: lifecycle-owned proof, opaque retained identities, bounded existing-client adapter, separate physical scopes, borrowed/shared context | Proposed extension; same owners acquire/release; S-014/S-015 prerequisites | Ownership leakage/double counting; F-015; S-018 / V-018 |
| C-012 outer envelope and receipt | Graph-main-entry to pre-scan v1 envelope; wrapper execs child; E-022 | Amend REQ-836/840/844: wrapper-entry through child-exit observation, explicit exclusions, scan/envelope-bound closed receipt after producer drain | Proposed exception; adoption precedes any post-scan write; required failure/cleanup/privacy consequences preserved | Scan bypass/exit drift; F-016; S-019 / V-019 |
| C-013 observer accounting and qualification | External native measurement and public validation implemented; candidate 16 fails; E-026/E-027 | Adopted REQ-848: complete cost and bounded uncertainty. CPU mean <2.5% is adopted/projected by completed S-020/A; all other gates unchanged | Preserve canonical REQ-377–380 portfolio, baseline writer and default-off; validate the measuring instrument and controls before another cohort | False qualification or endless optimization of noise; F-017/F-022; S-020/S-021 / V-020/V-021 |
| C-014 validated retained facts | Checker reads resource evidence twice and omits envelope validation used by explanation; E-020 | Clarify REQ-834–836/844: all published selected diagnostics validate once before aggregates, with consistent check/explain acceptance | Separate validation correction plus structural consolidation; S-014 clarification/S-015 regression before S-016; no compatibility guarantee for malformed envelopes | Unsafe/inconsistent acceptance; F-018; S-016 / V-016 |

Public readers remain `explain-run DETAIL=performance|resources`,
`harness-observability-check` and `harness-performance-check`. Add the semantic
`harness-instrumentation-qualify QUALIFICATION_PLAN=<file>` collection target
and a typed qualification manifest to the validator's `EVIDENCE_ROOTS_FILE` input. Preserve canonical performance manifests, portfolios,
baseline publication, same-source descriptive comparison, exact selection and
failure taxonomy. Human/JSON views share bounded facts; ordinary readers do not
construct OTLP. HTTP, WebSocket, mutations, saved views, revisions and UI selectors
are not applicable because no product contract changes.

## 5. Coupling and Boundary Findings

### Current lifecycle findings

| ID | Remediation / areas | Benefit / unresolved risk | Migration / exit |
| --- | --- | --- | --- |
| F-037 | Shared native runtime and scanner; implementation/tests/spec | One security and lifetime policy; duplication permits divergent ownership/path decisions | CLOSED: C-022, S-037/V-038, E-043/E-046/E-048; retained formats and scan identity stable |
| F-038 | Native review lifecycle and direct fixture library; implementation/projections/tests/spec | Single terminal authority; split ownership risks late publication and premature cleanup | CLOSED: C-023/024, S-038/V-039, E-044/E-047/E-048; other environments must drain old sessions with their running version |
| F-039 | Migrate standalone runtime callers; implementation/tests | Removes permanent dependency on Node policy; stranded roots or borrowed deletion otherwise remain possible | CLOSED: C-022, S-039/V-040, E-045/E-047/E-048; specialized adapters and persistent proof preserved |
| F-040 | Relocate tests and delete old oracles/routing; tests/implementation/projections | Tests survive structural changes and detect independent faults | CLOSED: C-025, S-036/S-040/V-037/V-041, E-042/E-046/E-048; per-assertion ledger complete |
| F-041 | Correct current owner navigation; documentation | Prevents edits against retired owners | CLOSED: C-025, S-040/V-041, E-046/E-048; historical paths remain source-scoped evidence |
| F-042 | Preserve cancellation through cache freshness and ordinary graph termination; implementation/tests | Prevents a user stop from becoming configuration failure and losing canonical evidence | CLOSED: existing cancellation contract, C-024, S-042/V-043/E-048; independent regressions, installed interruption and final check passed; no schema/public input changes |


### Previous functional iteration — current findings and final dispositions (historical)

| Finding | Remediation / areas | Rationale and long-term benefit | Compatibility / migration | Risk if unresolved | Completion evidence |
| --- | --- | --- | --- | --- | --- |
| F-033 specification drift | Reconcile definition v2, container availability, retired acquisition wording, scope-aware reporting and native ownership; specification/docs | One coherent current owner, fewer obsolete compatibility assumptions | Existing retained formats unchanged; native private interfaces adopted | Contradictory versions, restored retired commands or ownerless migration | S-035/C-021/V-036: owner/acceptance/traceability review and docs lint |
| F-019 misleading readiness | Replace current summaries and identify main versus off-main completion; documentation | One trustworthy restart point and accurate evidence scope | Preserve IDs and historical ledgers; supersede earlier next actions | Another agent merges an incomplete cutover or repeats experiments | S-033/V-034: clean main, commit/evidence ledger and explicit deferred items |
| F-029 experiments coupled to delivery | Remove collection/probe/build/schema registrations; retain qualification validation; specification, implementation, tests, documentation | Ordinary instrumentation has fewer dependencies and no self-profiling experiment path | Explicit retirement of qualify/diagnose commands; historical source in Git | Ongoing complexity and false completion claims | Main removal complete at `229936209`; active-source search and contract/projection gates pass |
| F-032 collected resources are poorly surfaced | Render container memory/limit, block-I/O deltas and PID gauges from existing validated metrics; label single-sample CPU unavailable; presentation, tests, owner clarification, docs | Make existing service observations actionable without adding collection | Human output only if current JSON suffices; preserve bounded output and explicit omitted counts | Users see `sampled_rss_max_bytes=not_observed` for containers despite retained memory observations | S-034/C-019/V-035: independent fixtures and fresh exact-run human/JSON fact parity; E-031 service run demonstrates the gap |
| F-028 fragmented/resident coordination | Short-lived source preparation and one Go graph/lifecycle owner; specification, implementation, tests, generated inputs, docs | Explicit lifetimes, fewer process boundaries and clearer cancellation/publication ownership | Stable public commands/results; coordinated private plan interface; no runtime-selection flag | Resident broad state and duplicated glue continue complicating attribution and maintenance | S-031/V-028/V-029/V-032: compiler reaped before units, scheduler/cache/failure parity and deletion closure |
| F-030 shared fixture policy across callers | One Go policy for graph and UI-review; retain physical service/browser adapters; multiple areas | One lease/reuse/recovery/cleanup implementation | Both callers migrate together before JS broker deletion; borrowed services remain borrowed | Duplicate cleanup authority, ownership leaks or hidden retained processes | S-031/V-030/V-032: ownership, partial acquisition, owner loss, recovery and real service/UI-review gates |
| F-031 tests duplicate implementation structure | Retire copied scheduler/wrapper algorithms; independent owner fixtures plus public boundaries; tests, routing, implementation cleanup | Tests assert durable behavior and can survive language changes | No product assertions or lifecycle/security cases removed merely because old APIs disappear | False confidence from testing the same algorithm twice; test-only compatibility burden | S-031/V-032: each changed test group has a keep/move/remove reason and a behavioral replacement where required |
| F-034 local-session TLS projection drift | Include existing TLS certificate/key/trust bindings in the private descriptor schema and derive environment selection from that schema; implementation/projections/tests/docs | One closed acceptance owner prevents physical attachment fields and descriptor selection diverging | Optional private fields accepted; existing descriptor payloads and schema version remain valid | Real persistent service creation fails before native borrowing; mocks conceal missing physical coverage | S-031/C/V-030: actual native borrowing, live-owner stop rejection, repeated native close, borrowed services still active, then exact owner cleanup |
| F-035 conformance tests retired source layout | Validate the actual pure OTLP projection's emitted scopes instead of a removed source-level export; implementation/tests/documentation | Conformance follows behavior across refactors without restoring an obsolete API | No wire or public command change; keep local-only reconstruction and explicit export gates | Correct harness output is rejected, encouraging compatibility shims or bypassed checks | S-032/V-033: otel-conformance and full warm check pass on final source |
| F-036 child summary collides with canonical root | Graph-child smoke publishes target summaries only; standalone smoke keeps its own root summary; implementation/tests/documentation | One root publisher and preserved producer-drain/scan ownership | Child artifact ownership is corrected; public canonical and accepted legacy formats unchanged | Root publication fails, live status cannot become terminal, and required scan is never reached | S-032/V-033: public full-check smoke integration, canonical summary, terminal status and scan audit pass |
| F-022/RB-006 overhead and repeatability | Defer certification; preserve validator and all failures; any future acquisition requires a separately adopted, bounded plan | Prevents a new profiling system from becoming the objective | No numerical relaxation or relabeling; 44.952-minute history persists | Unjustified default-on use or an unsupported overhead claim | S-020–S-022 remain DEFERRED; only eligible, stable complete frozen evidence can close them |


### Salvage findings — historical off-main salvage

| Finding | Evidence and risk | Required disposition |
| --- | --- | --- |
| F-028 fragmented graph execution ownership | Node graph composition retains broad preparation state and coordinates multiple native roots; E-029 | must_fix for salvage: C-015/C-018, S-027/S-028/S-030/S-031; V-028/V-029/V-031/V-032 |
| F-029 experiments coupled to delivery | Lifetime hooks and external collection extend ordinary dependencies despite failed repeatability; historical F-022 evidence retained | must_fix for salvage: C-017, S-024/S-026; V-025/V-027; budgets and failed dispositions unchanged |
| F-030 shared fixture ownership crosses graph/UI-review | `review-preparation.mjs` and UI-review cleanup import JS broker policy | must_fix before deletion: C-016, S-029/S-031; V-030/V-032; migrate all callers to one Go owner |
| F-031 non-production tests preserve obsolete execution structure | E-030: pending `simulation.mjs` duplicates scheduler policy for test-only callers; `wrapper-fixture.mjs` copies old coordination logic; smoke/matrix entries invoke the same full native suites | must_fix for S-031/V-032: C-015/C-016/C-018; retire duplicate algorithms, test behavior at its owner, and separate bounded smoke from boundary suites. No loss of required security/lifecycle/product coverage |

F-022/RB-006 remain unresolved and qualification is DEFERRED; they do not block
functional salvage. F-019 current-state correction is included in every checkpoint.

### Prior iteration — historical scope and evidence


F-001–F-006 and F-008–F-013 are completed historical findings at their recorded
sources. F-007's actionable remainder is selected through F-014–F-017, not silently
closed. The following table records current findings; `must_fix` applies only to
the mapped slice/gate, not unrelated cleanup.

| Finding / evidence / contracts | Remediation / affected areas | Rationale / long-term benefit | Compatibility / migration | Risk if unresolved | Validation / workstream |
| --- | --- | --- | --- | --- | --- |
| F-014 reliable cross-process placement missing; E-021; C-010 | Specification, runtime/producer implementation, execution schema, analysis, tests and docs: retain local durations and add measured anchors with uncertainty | Interpretable activity placement without a second scheduler clock or generic clock-sync framework | Version changed execution records/index; no UTC fallback or invented placement | Overlap/order explanations remain ambiguous or falsely precise | V-014/V-015/V-017; S-014/S-015/S-017; quantization, namespace mismatch, missing anchors, reversal, suspend and incomplete operations |
| F-015 lifecycle-owned containers unobserved; E-021; C-011 | Specification, suite-service adapter, collection, schemas/readers, tests and docs: exact private proof plus existing Docker client, collector-driven requests | Extends coverage without relocating lifecycle authority or enumerating unrelated resources | Version resource shapes; ownership/reuse/cleanup unchanged; no legacy translator | Process-only views understate service consumption or wrongly charge shared services | V-014/V-015/V-018; S-018; rejection/reuse/replacement/restart/borrowed/shared/disappearance/permission/cleanup cases |
| F-016 surrounding envelope omits wrapper, scan and completion; E-022; C-012 | Specification, wrapper/graph composition, publication, readers, tests/docs: supervised outer envelope and narrow terminal receipt | Exposes surrounding cost while keeping canonical time and scan authority clear | Version envelope; add closed receipt; preserve signals, workspace lock and exit precedence | Expensive preparation/finalization remains invisible; ad hoc post-scan writes weaken privacy | V-014/V-015/V-019; S-019; early failure, multiple/nested runs, scan failure, cancellation and interrupted/late publication |
| F-017 worker self-counters cannot qualify total observer cost; E-023/E-024; C-013 | Specification, measurement support, typed policy, tests/docs: scoped self-counters plus matched external measurements including helpers/parent/readers | Separates useful diagnostics from complete budget evidence; makes later extensions qualify against a known scope | Replace ambiguous observer fields atomically; separate qualification-manifest dispatch from canonical portfolios | Lower-bound CPU, process-peak RSS or omitted helper work can produce a false pass | V-014/V-015/V-020/V-021; S-020/S-021; exact accounting, uncertainty, memory and eligibility/budget fixtures |
| F-018 retained validation duplicated/inconsistent; E-020; C-014 | Implementation/tests/docs and owner clarification: one immutable validated loading result shared by check/explain | Removes repeated streaming and establishes one rejection policy before new formats arrive | Preserve command names/exits; malformed published envelopes become consistently rejected | Reader cost grows with artifact size and checker/explanation disagree | V-015/V-016; S-016; single traversal, malformed-envelope regression, security, output parity and export isolation |
| F-019 current tracker retains superseded scope/inventory/readiness; E-019; existing C-001–C-009 plus proposed C-010–C-014 | Documentation only: current baseline, 30-file inventory, next slices and named gates; preserve historical ledgers | Reliable restart point and evidence provenance; no repeated completed iteration | No runtime migration; historical IDs, source identities and failed attempts preserved | Repeated work, skipped prerequisites or historical success overstated as current qualification | V-023 / T-019 now; V-022 / S-022 final handoff |
| F-022 complete-observer qualification still fails; E-026/E-027; C-013 | Specification/projections, measurement implementation, tests and documentation as selected by S-020/A–E: validate measurement, establish identical-treatment variation, attribute CPU/RSS by lifecycle, then repair the responsible boundary | Replaces repeated full-matrix trials with falsifiable experiments; distinguishes observer cost, ordinary workload variation and measurement effects | CPU plan revision requires a new version and fresh evidence; version any new raw-evidence contract coherently; preserve all old dispositions and reject obsolete formats | Further small optimizations may never resolve elapsed/RSS instability; aggregate-only evidence cannot explain process-level causes | S-020 / V-020, then S-021 / V-021 and S-022 / V-022–V-023; F-022 and RB-006 remain open |

F-014–F-017 are selected extensions, not retrospective defects under unadopted
requirements. F-018 is a source-observed validation inconsistency, not a current
executed regression result; S-015 must characterize it explicitly and S-016 must
correct it separately from behavior-preserving movement. Any additional adopted
owner violation receives a new named correction and regression test.

### Prior findings — historical diagnoses and acceptance mappings

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

| Finding / evidence / contracts | Classification / affected areas | Remediation | Rationale / long-term benefit | Compatibility / migration | Risk if unresolved | Validation / slices |
| --- | --- | --- | --- | --- | --- | --- |
| F-010 invocation identity is confined to private failure channels; E-012; C-006 | must_fix for S-009; high; specification, runtime, runner adapters, tests | Neutral launch context shared by observation and failure handling; fresh identity before each harness-owned attempt, explicit parent/unit/rows and outcome | Successful, failed, cancelled and unlaunched attempts use one identity model; failure transport no longer owns general execution meaning | Migrate private callers together; retain strict failure-envelope identity checks, cleanup and public exits; no forwarding aliases | Nested/packed runners remain ambiguous and depend on process sampling for correlation | V-007/V-008; S-007/S-008/S-009; independent Go/Vitest/shell/Playwright and unit-launch fixtures |
| F-011 allocation/process correlation lacks complete lifecycle relationships; E-013/E-015; C-007/C-009 | must_fix for S-010/S-012; high; specification, broker/providers, collection/readers, tests | Allocate opaque physical-allocation identity before acquisition; separate lease, invocation and process relationships, failed attempts and reuse | Ownership remains cohesive; shared physical work is counted once without assigning it to its first consumer | Preserve ownership/reuse/cleanup; remove redundant private correlation paths after atomic caller migration; version incompatible retained shapes | False per-unit attribution, shared-work double counting and lost failed-acquisition context | V-009/V-011; S-010/S-012; shared/borrowed/reused/failed acquisition, PID replacement/reuse, missing proof and conflict fixtures |
| F-012 detailed timings are disconnected from execution identity; E-014; C-008/C-009 | must_fix for S-011/S-012; medium; specification, catalog, producer owners, retained analysis, tests | Reuse migration/reset measurements; add missing failure outcomes and report-parse boundaries with owner clock and invocation/allocation links | Measures work at a known boundary and avoids duplicate measurements or invented scheduler time | New optional diagnostic data; no canonical duration/path changes; missing/incompatible clocks remain unavailable | Setup/parsing costs stay obscured; success-only data and cross-process subtraction can mislead | V-010/V-011; S-011/S-012; hand-calculated success/failure/cancellation/gap/nesting/overlap/shared-clock cases |
| F-013 current tracker summaries describe completed foundation work as future work; E-011/E-017 | should_fix; documentation; this update | Replace current summaries, inventory, readiness and next-entry text; preserve stable IDs and historical ledgers under their original baselines | One unambiguous starting point and no repeated foundation delivery | No runtime migration; original run identities, failed attempts and applicability limits retained | Repeated work or historical success presented as qualification | V-013 / T-011; complete finding→contract→slice→validation trace and tracker-only preservation |

## 6. Refactor Workstreams

### Current lifecycle workstreams

The sole active dependency chain is **S-036 → S-037 → S-038 → S-039 → S-040 →
S-041 → S-042**. Every predecessor must pass and its checkpoint must be saved
before successor work begins. Prior completed graph/fixture work remains DONE.

| Workflow / slice | State | Required exit |
| --- | --- | --- |
| WF-36 / S-036 contracts and characterization | DONE; E-042 | V-037: current owner, caller/test ledger, independent expectations and fresh baseline pass |
| WF-37 / S-037 shared native runtime/scanner | DONE; E-043 | V-038: native ownership/security/race checks and shared consumers |
| WF-38 / S-038 native review lifecycle | DONE; E-044 | V-039: real IPC, recovery and physical boundaries in every mode |
| WF-39 / S-039 remaining callers | DONE; E-045 | V-040: owned/borrowed/persistent callers pass failure and interruption cases |
| WF-40 / S-040 deletion/routing/docs | DONE; E-046 | V-041: import/build/generated closure and every assertion disposition closed |
| WF-41 / S-041 public boundary validation | DONE; E-047 | V-042: installed public surface, producer interoperability and physical cleanup |
| WF-42 / S-042 final validation/handoff | DONE; E-048 | V-043: finalize, frozen-source contracts/lint/projections/full warm check and saved handoff |


### Previous functional iteration — current ordered workstreams (historical)

This is the sole dependency list for the current effort. Each saved checkpoint
must precede any work on its successor. S-031 rolls up A–F; each is an independent
workstream with its own result and checkpoint.

| Workflow | Slice / prerequisite | Status | Exit / checkpoint |
| --- | --- | --- | --- |
| WF-33 | S-033; historical cleanup | DONE | V-034/E-031 |
| WF-35 Specification | S-035 after S-033 | DONE | V-036: reviewed current owner, acceptance and lint |
| WF-34 Reports | S-034 after saved S-035 | DONE | V-035: scope-correct fixtures and fresh exact-run parity |
| WF-31A Characterization | S-031/A after saved S-034 | DONE | Explicit test dispositions and executed baseline obligations |
| WF-31B Native core | S-031/B after saved S-031/A | DONE | V-028/V-029: plan, compiler lifetime, engine/cache/lifecycle race gates |
| WF-31C Fixtures | S-031/C after saved S-031/B | DONE | V-030: one fixture policy, real lifecycle/recovery and race gates |
| WF-31D Observation | S-031/D after saved S-031/C | DONE | V-031: collection/drain/scan/receipt and independent reader gates |
| WF-31E Cutover | S-031/E after saved S-031/D | DONE | One public runtime/fixture owner, caller/generated/deletion closure |
| WF-31F Public proof | S-031/F after saved S-031/E | DONE | V-032: current-source public success/failure/cancel/cleanup; S-031 closes |
| WF-32 Handoff | S-032 after saved S-031/F | DONE | V-033/E-040: finalization, frozen-source broad checks, warm check and handoff |
| WF-20–22 Qualification | S-020–S-022 | DEFERRED | Separate adoption and complete eligible stable evidence |

Each checkpoint records files/changes, exact commands, source commit/state/digest,
environment, roots, outer exits, failures/skips, compatibility and next readiness.
Rollback uses one coherent owner/projection transaction. Preserve failed roots;
no alternate-engine switch or new source archive.

### Salvage workflow dependency authority — historical off-main salvage

All workstreams are sequential chains; the dependency column is authoritative.
Save this tracker after each completion and before starting its successor.

| Workflow | Slice | Depends on | Exit gate |
| --- | --- | --- | --- |
| WF-23 Preservation and inventory | S-023 | none | V-024: Restore archive bytes; primary index/source preserved; exact dispositions |
| WF-24 Specification reconciliation | S-024 | S-023 | V-025: Adopt runtime/fixture boundaries and retire experiments; budgets unchanged |
| WF-25 Retained foundation | S-025 | S-024 | V-026: Narrow owner checks; validation/selection/security invariants preserved |
| WF-26 Experimental machinery removal | S-026 | S-025 | V-027: No experimental imports/launches; independent validator unchanged |
| WF-27 Native composition and preparation | S-027 | S-026 | V-028: Valid private plans; compiler exits before execution; status never builds |
| WF-28 Go graph engine | S-028 | S-027 | V-029: Independent scheduling/cache/failure/cancellation/timing parity |
| WF-29 Shared fixture owner | S-029 | S-028 | V-030: Graph/UI-review ownership, recovery, cleanup and real service parity |
| WF-30 Observation integration | S-030 | S-029 | V-031: Clock/container/receipt bounds, quietness and reader/export parity |
| WF-31 Public cutover and deletion | S-031 | S-030 | V-032: Single runtime/fixture owner, behavior-based test cutover, public and generated closure |
| WF-32 Functional validation and handoff | S-032 | S-031 | V-033: Finalization then broader functional checks; qualification still deferred |

### Prior iteration — historical scope and evidence


This is the **sole authoritative prerequisite list for WF-14–WF-22**. Slice rows
reference it. All workstreams are included, serial and separately checkpointed;
no parallel agent work is requested. The current implementation request supplies execution authorization.

| Workstream / slice | Class / prerequisite | Goal / likely seams | Validation / saved checkpoint | Principal risk / binary exit |
| --- | --- | --- | --- | --- |
| WF-14 / S-014 owner reconciliation | root; completed S-013 history and E-019–E-024 revalidated on entry | Adopt C-010–C-014 in Testing Harness; exact privacy/limits/formats/qualification semantics | V-014; owner review and projection map saved | Contradictory authority; no unresolved semantics or acceptance mapping; specifications/tracker only |
| WF-15 / S-015 characterization/routing | chain; WF-14 | Wrapper/helper/service/clock/reader consumers and authored test/workload routes | V-015; executed baseline and correction cases saved | Unexecuted fixtures; owner-conforming characterization passes and every changed boundary has explicit expected cases |
| WF-16 / S-016 unified retained validation | chain; WF-15 | One secure normalized loading result; check/explain/export consumer migration | V-016; exact rejection/parity and traversal evidence saved | Partial aggregates escaping closure; all security/selection/parity gates pass |
| WF-17 / S-017 clock anchoring | chain; WF-16 | Runtime/Go/helper anchors, execution projection and uncertainty-aware analysis | V-017 plus V-014 projection gates; clock/canonical parity saved | False precision; independent arithmetic/discontinuity tests pass and unsupported placement stays unavailable |
| WF-18 / S-018 container observations | chain; WF-17 | Lifecycle proof, bounded existing-client helper, collection and resource projections | V-018 plus V-014; real service and cleanup evidence saved | Ownership leakage/double counting; all owned/shared/borrowed/failure/quietness cases pass |
| WF-19 / S-019 wrapper/terminal receipt | chain; WF-18 | Workspace supervision, graph envelope and securely bound receipt | V-019 plus V-014; publication and signal/exit evidence saved | Scan bypass/exit drift; security/order/membership/interruption cases pass |
| WF-20 / S-020 observer measurement | chain; WF-19 | Component scope, controlled external measurement, typed manifests, bounded views and obsolete path removal | V-020 plus V-014; coverage/measurement fixtures saved | Lower bounds promoted to totals; exact accounting and measurement-loss/quietness gates pass |
| WF-21 / S-021 qualification | chain; WF-20; completed clean frozen source and adopted WSL2 environment prerequisites | Execute frozen workload × reader × off/basic matrix; evaluate typed budgets | V-021; every root, treatment, exclusion, statistic and gate result saved | Noise/mismatch/uncertainty; every required cell eligible, stable and within every budget |
| WF-22 / S-022 validation/handoff | chain; WF-21 | Final current-source verification, migrations/docs, generated/import review and complete reconciliation | V-022/V-023; final handoff saved | Historical evidence reused; every selected obligation closes against completed source |

Derived sequence: **S-014 → S-015 → S-016 → S-017 → S-018 → S-019 → S-020 → S-021 → S-022**.
After each workstream passes its required exit, update this tracker **before**
beginning its successor. Record substantive files/changes, exact commands,
source commit/state/digest, environment and run roots, results including failed
attempts, compatibility effects, unresolved risks and successor readiness. A
failed gate blocks its dependents. Never erase failed attempts after repair.

### Completed workflow dependency history

WF-00–WF-13 remain DONE at their recorded baselines. The following prerequisite
lists are historical and do not redefine the current sequence.

| Workflow | Historical prerequisites | Delivered goal / seams | Historical validation | Recorded exit |
| --- | --- | --- | --- | --- |
| WF-00 specification and tracker reconciliation | root; current implementation request | Tracker and Testing Harness REQ-840 / AC-137 | V-001 and Markdown lint | Adopted projections, phase bound, and execution checkpoints agree |
| WF-01 characterize public and private boundaries | chain; WF-00 | Existing instrumentation/observability and wrapper tests | V-002, V-003, V-005 | Independent fixtures cover frozen behavior and expose lifecycle entry assumptions |
| WF-02 consolidate definitions | chain; WF-01 | Catalog, schemas, generator, attachments, ownership | V-004 | One vocabulary source; generated validation and runtime descriptors agree |
| WF-03 normalize canonical facts | chain; WF-02 | Canonical reducer, performance projection, scheduler writer, validator | V-002 | No duplicate raw-event array; projections preserve expected values/errors |
| WF-04 isolate collection lifecycle | chain; WF-03 | Engine, worker/controller, Linux adapter, runtime gate, run session | V-003 | Deterministic interleavings and real quietness tests pass; resources close before scan |
| WF-05 simplify consumers | chain; WF-04 | Observability facade, resource reader/reducer, export, wrapper | V-005, V-006 | Public parity and ownership checks pass; removed private paths have no retained callers |
| WF-06 validation and handoff completion | chain; WF-05 | Final verification and durable handoff | V-006 plus Markdown/diff checks | All foundation gates pass; T-007/T-008 remain deferred |

| Workflow / slice | Class / authoritative prerequisites | Goal / likely seams | Validation / checkpoint | Principal risk / binary exit |
| --- | --- | --- | --- | --- |
| WF-07 / S-007 owner reconciliation | root; completed WF-06 foundation baseline, E-011–E-017 revalidated at entry | Testing Harness owner: identities, relationships, local clocks, privacy, bounds, incomplete evidence and acceptance map | V-007; record adoption and projection map before machine changes | Promoting mechanisms into authority; owner inputs/outputs/defaults/rejections stated once, with no unresolved owner contradiction |
| WF-08 / S-008 characterization | chain; WF-07 | Exact launch/lifecycle consumer and route inventory; independent fixtures on existing interfaces | V-008–V-011 baseline fixtures; checkpoint executed results/correction obligations | Golden preservation of defects; owner-aligned baseline passes and proposed outcomes have explicit expected fixtures |
| WF-09 / S-009 invocation context | chain; WF-08 | Runtime context, executor, Go/Vitest/shell row runners, browser and lifecycle adapters | V-008 and V-007 generation when inputs change; checkpoint migrated callers | Stale/inherited identity; all supported harness launch attempts have explicit unique context/outcome while failure channels remain private |
| WF-10 / S-010 allocation/process relationships | chain; WF-09 | Broker/provider acquisition entry, observation and retained correlation seams | V-009 plus applicable V-007; checkpoint ownership/failure parity | Changing cleanup authority or first-consumer attribution; relationship closure preserves physical ownership and cleanup |
| WF-11 / S-011 owner-local timings | chain; WF-10 | Go migration, browser reset, row/browser report parsing and catalog | V-010 plus applicable V-007; checkpoint local-clock and canonical parity evidence | Duplicate measurements or cross-clock arithmetic; timing/outcome fixtures pass with unchanged canonical accounting |
| WF-12 / S-012 retained analysis | chain; WF-11 | Secure validated loading, pure normalized analysis, bounded human/JSON projection | V-011 plus applicable V-007; checkpoint integrity/privacy/public outcomes | Unsafe reads or double counting; declared relationships close, shared/ambiguous work stays contextual, export remains isolated |
| WF-13 / S-013 validation/handoff | chain; WF-12 | Final source, generated outputs, owner routing, findings and durable handoff | V-012; final checkpoint with exact commands/runs/failures/skips | Overstated qualification; all selected obligations have current-source evidence and deferred work remains explicit |

## 7. Remediation Slice Plan

### Current lifecycle slice implementation and test dispositions

**S-036:** adopt native shared/session/review ownership, enumerate callers and
freeze independent expectations before moving production policy. Baseline native
runtime/fixture and UI contract gates must pass. Specification review distinguishes
required behavior from implementation accidents; missing characterization blocks
movement. **S-037:** extract native runtime, journal, preservation and scanner into
`harnesssession`; graph/fixtures call it directly; expose bounded private operations
for Node callers. Test security, concurrency and independent scan expectations.
**S-038:** move registry authority, admission, stopping, recovery serialization,
cleanup and terminal publication to `harnessreview`; retain browser actions and
preparation recipes as adapters. Test real IPC and service/browser boundaries.
**S-039:** migrate smoke, wrapped tools, frontend producers and local-service root
ownership; inherit existing authority or create owned native roots; no daemon.
**S-040:** delete superseded policy and fixture transport, relocate tests, update
routing/schema/helper owner inputs, regenerate, and repair active documentation.
**S-041:** exercise public boundaries and repair defects at owners. **S-042:** run
finalize before final contracts/lint/projections/full warm check, freeze source,
and save exact evidence and remaining qualification limits.

| Test group / obligation | Keep / move / remove disposition | Owner / route |
| --- | --- | --- |
| Native runtime private proof, preservation/janitor | MOVE unchanged obligations, extend concurrent/closed behavior | `harnesssession`; runtime contract |
| Native scanner Node parity | REMOVE Node oracle; retain fixed expected accept/reject/count/policy identity and unsafe filesystem cases | native scanner; runtime contract |
| Node suite-runtime contract group | MOVE ownership/scan assertions; retain Node producer umask boundary | native session plus command/evidence contracts |
| Collection-engine-named process exit test | KEEP actual exit-wakeup versus identity proof obligation; relocate to process adapter test if still needed | producer boundary; evidence contract |
| UI controller/preparation/seeded/lifecycle/recovery | MOVE policy assertions to actual Go owner; KEEP browser adapter and real IPC/physical tests | UI contract/lifecycle/seeded routes |
| Fixture broker allocation/reuse/recovery and physical owners | KEEP; remove transport-only tests only when private protocol is deleted | fixture contract and native fixture row |
| Frontend producer/private capture/command failure/local services/smoke | KEEP behavioral cases; use native runtime rather than legacy test constructor | exact owner selections and command surface |
| Compiler, catalog, Node/Go schema, offline readers, qualification validator | KEEP distinct responsibilities; never reroute through legacy lifecycle policy | existing owner routes |

Principal risks are unsafe deletion, lost independent cleanup, primary failure
replacement, hidden callers and stale evidence. Every removed assertion receives
an exact retained/relocated/obsolete disposition in the completion checkpoint.
Rollback is a coherent source/projection transaction. A private protocol change
requires old live sessions to drain with their running version; unverifiable
ownership blocks that environment's cutover. No legacy reader or receipt rewrite.


### Previous functional iteration — adopted workstream plan and completed test dispositions (historical)

S-033 is complete. The following adopted sequence is complete through S-032;
Sections 6, 9 and E-032–E-040 record its individual checkpoints and results.

**S-035 — Reconcile specification and current tracker.** Correct REQ-840 v2
identities, REQ-833 container availability and REQ-844 retired acquisition wording;
remove obsolete milestone prose. Adopt REQ-852–855/AC-147–148 with one graph/fixture
owner, bounded private interfaces and independent behavior tests. Clarify C-019
report acceptance in REQ-835/849 and AC-146. Extend acceptance traceability.
Keep implementation sequencing here and vocabulary in its existing owner.
V-036 requires human owner/projection review, `make lint-markdown`, no runtime
changes for editorial correction, and the saved checkpoint before S-034.

**S-034 — Make existing observations useful.** Start at
`tools/harness/observability/performance-presentation.mjs`,
`resource-presentation.mjs`, `resource-accumulator.mjs` and
`tests/test-instrumentation.mjs`. In E-031's service run, each of two owned
containers has a valid memory/limit/PID and cumulative CPU/I/O sample. Human
`[SCOPE]` output only prints process RSS; JSON contains the container facts.
One CPU sample cannot establish utilization or an interval delta. Clarify C-019's
owner acceptance, then display the existing scope-correct facts without another
loader, cache or sampler. Keep context, containers and processes separate.
Independently authored fixtures cover one/two samples, reset, unavailable metrics,
shared allocations and truncation. Run `make harness-evidence-contract` and fresh
exact-run human/JSON explanations. No new schema or abstraction is justified just
to format an already-validated field. Checkpoint V-035 before changing runtime
ownership. Main command spelling and result/failure classifications remain stable.

**S-031/A — Reconcile contracts and test ownership.** Compare main against committed
salvage references `3528f4379` (composition/preparation), `9fbcecdcd` (engine),
`ba543bf69` (fixtures) and `79aecc1ca` (observations). Inspect the existing dirty
S-031 difference separately; do not import it wholesale. Apply S-035's adopted
REQ-852–855/AC-147–148 boundaries when updating typed projections. Record a short keep/move/remove disposition for changed test groups,
not a new general-purpose test registry. Preserve selection, cache correctness,
admission/fairness, canonical accounting, security and cleanup guarantees.
Copied `simulation.mjs` and `wrapper-fixture.mjs` coordination are not independent
oracles; replace required assertions with explicit inputs/expected events at the
owning Go boundary. Entry: S-034 checkpoint. Exit: each moved responsibility and
removed test group has one owner and a concrete acceptance obligation. Record any implementation defect as a failed gate; the selected ownership
destination is fixed, not another architecture investigation.

#### S-031/A test ownership disposition

The following is human migration accounting, not a new executable registry.
Existing typed owner/family routing remains authoritative. Retiring an API does
not retire the observable assertion; E/F must close each row against final source.

| Changed group | Disposition / durable assertion | Destination and route |
| --- | --- | --- |
| Graph compiler/model, browser groups, selectors, tiers and capability/cache input facts | KEEP Node; deterministic inputs, exact row/output closure and source validation | Existing graph/browser contracts and harness-contract-tests |
| Scheduler simulation, graph cutover scheduling modes, contract-suite scheduling/cancel cases | MOVE expected admission/events/failure/cancel assertions; REMOVE duplicated JS algorithms | Go TestRuntimeScheduling/Lifecycle/FailurePrecedence; harness-runtime-contract and exact native_runtime row |
| Browser product/reset failure simulations | KEEP compiled failure-policy assertions in Node; MOVE execution proof, no retry and later fresh allocation | Native engine/fixture integration with explicit product/reset failure fixtures |
| Graph cache restore/write/security tests | MOVE transaction assertions; KEEP source dependency/digest/scanner/paired-output policy tests in Node | Native TestRuntimeCacheTransactions/PublicationRollback; Node compiler/evidence policy contracts |
| JS broker/pool/cleanup and physical proof fixtures | MOVE policy/concurrency/idempotence/primary failure to one Go owner; KEEP physical adapter assertions | harness-fixture-contract plus native_fixtures row and real service boundaries |
| Go/Vitest/Playwright/shell row producers and private failure channel | KEEP producer/report assertions; MOVE executor binding, wait/cancel and failure-scope ownership | Native producer interop, command-surface contracts and exact public rows |
| Canonical reducer, strict retained loading, clocks, reports, schema and export isolation | KEEP independent Node expected facts and rejection fixtures; replace only runtime-driving setup | harness-evidence-contract; native canonical/observation/schema parity |
| JS collection engine/Linux adapter/transport tests | MOVE actual collection/death/pause/stop/proof assertions to native owner; REMOVE dead JS mirrors | Native collection/admission/runtime contracts, physical boundary route |
| Wrapper registration coordinator fixture | REMOVE copied coordinator; hand-author closed protocol data for reader/publisher tests; exercise real registration natively | Native observation/receipt tests, independent Node receipt validation, real wrapper signal/nesting cases |
| UI-review acquisition/recovery/cleanup | KEEP user-facing lifecycle and exact proof failure/recovery; migrate client transport only | harness-ui-review-contract/lifecycle/seeded-default |
| Work-graph smoke/matrix wrappers | REMOVE repeated whole native-suite launches and obsolete modes; preserve bounded compiler smoke and catalog routing | Authored task surface and owner routes, command-surface/public checks |
| Generated native validators and failure policy | REGENERATE from shared authored owners; independent acceptance/arithmetic expected values | native schema parity, generation/drift/policy/shape gates |

Candidate revisions inspected independently: `3528f4379` (preparation),
`9fbcecdcd` (engine), `ba543bf69` (fixtures), `79aecc1ca` (observations).
Dirty S-031 is a separate candidate: its simulation and wrapper copy are rejected,
not automatically imported. No product assertion is removed because of language
or internal API changes. B–D retain current public dispatch until E's transaction.


E's disposition closure: native `TestRuntimeBrowserFailureContinuation` covers
product/reset failure, one attempt, fresh later allocation and typed reset failure;
`TestRuntimeProducerInterop` covers selected Node binding, private diagnostics
through Make, cancellation/reaping and private capture nonpublication/cleanup.
`test-command-failure.mjs` retains all eleven scanner causes as direct producer
fixtures plus actual canonical shell-row artifact failure and private cleanup.
Native engine tests independently cover aggregate continuation, journal cause and
exit precedence. Node compiler/group/report/schema tests remain independent.
Wrapper tests author closed protocol data and invoke the real native publisher;
there is no copied wrapper coordinator. Scheduler and broker smoke/matrix checks
select named Go tests instead of launching whole suites. Bootstrap failure tests
assert the real pre-runtime diagnostic; no synthetic canonical run is required
before the runtime can be built. JSON assertions parse semantics, not indentation.
No required lifecycle, security, accounting or product assertion is dropped.

**S-031/B — Integrate the native core in small source-coherent transactions.** Reuse
reviewed code under `internal/testutil/harnessruntime` and the thin
`tools/harnessruntime` root; retain Node `work-graph` compiler/model/selector logic.
The private compiler result carries immutable graph, projection, provenance,
capacity, launch and cache facts, not ASTs, catalog closures or credentials.
Validate the inherited-descriptor plan strictly with the existing generated schema
machinery, enforce the 32 MiB bound, reject closure/identity/version errors before
workload side effects, and reap the compiler before any unit starts. No second
handwritten schema registry or public arbitrary-command plan reader. Keep public
execution on the current main runtime until the complete cutover gate is ready;
do not add a public runtime selector. Restore the existing candidate's Make-owned
`harness-runtime-contract` and `harness-runtime-boundary-contract` routes through
authored task inputs when their implementations are integrated. Inject clocks,
executors and fixtures; expected scheduling/canonical arithmetic must be authored
independently of production algorithms. Gate V-028/V-029 before fixture integration.

**S-031/C — Migrate fixture policy for every caller.** Use
`internal/testutil/harnessfixtures` as the single coordination owner. Graph execution
calls it directly; `tools/harness/browser/review-preparation.mjs` and UI-review
cleanup use a narrow private session client. The existing service/browser adapters
retain physical provisioning and termination. Closed acquire/release/recover/close
frames carry request IDs and lease/run proof, bounded to 1 MiB; EOF/owner loss
initiates owned cleanup, never stopping borrowed services. Test partial acquisition,
replacement, reset, pooling, idempotent close, primary failure precedence and
recovery after forced child death. Integrate `harness-fixture-contract`; run real
PostgreSQL/object-store/browser rows and `harness-ui-review-contract` plus
`harness-ui-review-lifecycle`. Gate V-030 before deleting any JS policy.

**S-031/D — Connect observation at the lifecycle owner.** Call admission libraries
directly and use an isolated collector child of the consolidated binary where
fault isolation/quietness requires it. Reuse the current collection schedule,
private container proof, clock anchors, validated retained readers and schemas.
Do not create another timer or collector because the runtime language changes.
Required order stays unit completion → graph publication → ordinary producer
drain → retained scan → bounded outer receipt. Test collector death, pause/stop
cancellation and drain, unavailable clocks, namespace/discontinuity rejection,
container replacement/disappearance and multiple/nested graphs. A missing/failed
scan cannot yield complete-success receipt; no further post-scan write is allowed.
Gate V-031 against the existing independent fixtures and fresh public runs.

**S-031/E — Cut over once and delete superseded owners.** Change authored Make
routing/build inputs, regenerate, migrate remaining imports and delete the old
Node scheduling/execution coordination, graph cache runtime/live publication,
JS fixture policy, collector bridge and standalone admission/collection/diagnostic
composition roots. Keep source compilation, row runners, runner adapters and
offline analysis. Consolidate private operations under `cartulary-harness` without
mixing library responsibilities. Read-only status must use an already provisioned
binary and must never initialize writer state. Delete copied test coordinators and
obsolete helper fixtures after coverage moves; do not retain dead APIs to satisfy
old test structure. Bound smoke routes to smoke work instead of running full
native suites repeatedly. Gate: source/generated/caller closure, no dual engine,
no permanent fallback, no new experimental entrypoint and no Markdown dependency.

**S-031/F — Prove the public boundary and checkpoint.** Run native unit/fixture
routes first, then evidence and command-surface contracts, exact shell/local
migration rows, real schema-bootstrap/HTTPS-trust routes, and UI-review lifecycle
routes. After projections change, run generate, drift, generated-artifact policy
and JSON-shape checks. Require current-source success/failure/cancellation and
owned cleanup through public Make, exact-run status, observability validation and
human/JSON performance/resource reports. Characterize source preparation and live
process lifetimes using retained phases/ordinary observations; do not infer an
RSS improvement from process size alone. A defect blocks V-032; repair its owner
and rerun the affected boundary. Save S-031 before beginning S-032.

**S-032 — Functional validation and handoff.** Run `make agent-finalize`, review
any generated refresh, then freeze the executable commit/digest before broader
checks: `harness-contract`, script,
shell and Markdown lint, applicable generation/drift and a full warm `make check`.
Supply `RESULTS_DIR` to finalization only when an already eligible successful full
warm check exists; otherwise record retained-run maintenance as skipped. Recheck
fresh exact-run evidence and export isolation locally, without transmitting
telemetry. Record every failed attempt, migration/removal, test disposition and
residual physical coverage limitation. Close V-033 only on this final source.
Functional completion is separate from S-020–S-022 and does not establish an
overhead budget, publish baselines or enable diagnostics by default.

**Later qualification decision.** No acquisition implementation or new cohort is
scheduled here. A later adopted plan must first demonstrate calibrated independent
repeatability and bounded measurement error, using the unchanged workloads,
reader schedule, budgets and 44.952-minute investigation history. If those gates
cannot be met, report the limitation. Do not revive the removed experiment stack,
lengthen workloads or select only favorable attempts merely to produce a pass.


### Go consolidation salvage — historical off-main salvage

The accepted implementation plan has no permanent dual-engine switch. Rollback is
a complete authored/projection/producer/reader transaction to preserved source.
No generated file is hand-edited. The S-031 working candidate already selects Go
for public graph dispatch. Its later test/deletion edits are incomplete; only the
committed S-030 baseline has the full recorded validation. The cutover is accepted
only when the whole S-031 transaction passes V-032.

| Slice | Work | Depends on | Status / readiness | Binary completion |
| --- | --- | --- | --- | --- |
| S-023 | Preservation and inventory | none | DONE; complete | V-024: Restore archive bytes; primary index/source preserved; exact dispositions |
| S-024 | Specification reconciliation | S-023 | DONE; complete | V-025: Adopt runtime/fixture boundaries and retire experiments; budgets unchanged |
| S-025 | Retained foundation | S-024 | DONE; complete | V-026: Narrow owner checks; validation/selection/security invariants preserved |
| S-026 | Experimental machinery removal | S-025 | DONE; complete | V-027: No experimental imports/launches; independent validator unchanged |
| S-027 | Native composition and preparation | S-026 | DONE; complete | V-028: Valid private plans; compiler exits before execution; status never builds |
| S-028 | Go graph engine | S-027 | DONE; complete | V-029: Independent scheduling/cache/failure/cancellation/timing parity |
| S-029 | Shared fixture owner | S-028 | DONE; complete | V-030: Graph/UI-review ownership, recovery, cleanup and real service parity |
| S-030 | Observation integration | S-029 | DONE; complete | V-031: Clock/container/receipt bounds, quietness and reader/export parity |
| S-031 | Public cutover and deletion | S-030 | IN_PROGRESS; ready for the ordered repair below, not completion | V-032: Single runtime/fixture owner, behavior-based test cutover, public and generated closure |
| S-032 | Functional validation and handoff | S-031 | TODO; blocked on predecessor | V-033: Finalization then broader functional checks; qualification still deferred |

S-024 reconciles REQ-845–851 and adds the graph/fixture/interface owner clauses.
S-025 retains the observed metadata-fixture repair and independent foundation
fixes, without claiming overall qualification. S-026 executes the exact archive
ledger in Section 2 and splits mixed tests. S-027 introduces `cartulary-harness`
at `tools/harnessruntime` with coordination in `internal/testutil/harnessruntime`;
source-aware Node compilation emits graph/projections/provenance/capacity/launch
and cache facts, not ASTs, closures, credentials or callable behavior. S-028 ports
graph cache, scheduling, live status, process lifecycle and required publication.
S-029 shares `internal/testutil/harnessfixtures` between graph and UI-review.
S-030 preserves unit completion → required publication → producer drain → scan
→ bounded outer receipt, and reuses native observation libraries. S-031 migrates
every caller before deleting JS runtime/broker policy and standalone native roots.
S-032 supplies a functional handoff only; S-020–S-022 remain DEFERRED.

### Clean harness and non-production test cutover — historical off-main salvage

The user asks to follow the purpose of the tests, not preserve their old internal
shape. This refinement supersedes one-for-one test-porting obligations in the
unfinished S-031 notes. Preserve developer-visible guarantees and product
assertions. Test filenames, JavaScript classes, callback order, helper signatures,
internal smoke labels and historical test counts are not compatibility contracts.
Where an adopted clause actually fixes a route or behavior, amend that owner and
its typed projection together before changing it. This tracker is never an
executable acceptance registry.

Scope is the graph runtime, shared fixture coordination, harness self-tests and
their non-production adapters/helpers. Product implementation, application test
assertions, browser scenarios, golden expectations and performance acceptance
criteria are outside the simplification. Keep Go/Vitest/shell/Playwright row
execution and report adaptation in their current owners. A clean cutover means
one execution owner and one fixture policy, with small explicit language
boundaries. It does not require converting every test to Go.

The observable contract is: exact current-input selection; correct admission,
cache and failure semantics; drained owned children/services; untouched borrowed
resources; trustworthy canonical evidence; and bounded, honest diagnostics.
Keep schema versions, uncertainty and omission semantics, scan/receipt ordering,
private proof, default-off diagnostics and explicit export unchanged. Useful
output must answer where time went, what waited, what resource consumption was
actually observed, and which evidence is unavailable. No new measurement system
or performance qualification is part of this cutover.

| Current code/test family | Concrete disposition | Reason, migration impact and acceptance |
| --- | --- | --- |
| `internal/testutil/harnessruntime/{engine,lifecycle,process_linux,cache_linux,canonical,publication_linux,observation_linux}.go` and their tests | RETAIN as the execution owner. Put scheduler/cache/lifecycle expectations here using controlled clocks, executors and failures. Extend only for a demonstrated missing contract. | One tested policy. Expected admissions, durations and failures are authored independently; old JS results are supporting history, not an oracle. Go process tests must still exercise real spawn, wait and cancellation. |
| `tools/harness/scheduler/work-graph/{compiler,model,preparation,prepare-plan,cache-dependencies,cache-plan,launch-binding}.mjs` | RETAIN preparation, exact selection, current-source validation and compact plan facts. Keep Node tests of those outputs. | Compiler exits before work. Keep dependency invalidation, full preflight and selected-input freshness; do not keep execution objects merely for test imports. |
| Pending `work-graph/simulation.mjs` and its exports in `index.mjs` | REMOVE. Replace test-only uses in `contract-suite-support.mjs`, `test-work-graph-cutover.mjs` and `browser/tests/test-browser-work-graph.mjs` with compiler assertions or Go engine tests, according to the behavior under test. | Search found only test callers. `diagnostics/target-plan-cli.mjs` compiles graph facts and does not use this simulation. Do not install a second scheduler as an expected-result generator. Any deterministic simulation needed for native tests uses the actual Go policy with injected ports. |
| `internal/testutil/harnessfixtures` and `tools/harness/runtime/fixture-session.mjs` | RETAIN Go lease/reuse/recovery policy and a transport-only Node client. Test graph and UI-review use of the same owner. | Physical browser/service adapters keep acquisition/reset/stop responsibilities. Ownership, partial failure, replacement, EOF and repeated close remain explicit; no duplicated JS broker survives. |
| `contract-suite-support.mjs`, `test-work-graph-cutover.mjs`, old broker/executor/live-publisher tests | SPLIT by tested behavior. Retain compiler/catalog/source checks in Node; assign native execution rules to existing Go tests; retire cases that only exercise deleted APIs. | No obligation to preserve test names/counts or rebuild legacy facades. Each removed behavior-bearing group needs an owner test or an explicit obsolete-behavior reason, recorded here at V-032. Missing safety coverage blocks deletion acceptance. |
| `test-command-failure.mjs`, `private-child-capture-cases.mjs`, frontend producer and runner-adapter tests | RETAIN producer/parser/privacy tests near the producer. Use a compact Go-to-Node/Make integration set for failure propagation and cleanup. | Scanner report interpretation stays with the scanner adapter, not duplicated as a complete runtime matrix. Verify clean result, classified failure, contradictory/malformed publication and cancellation across the real handoff; preserve all distinct required result classes in owner tests. |
| `browser/tests/test-browser-acquisition.mjs`, `ui-review-fixture-cleanup.mjs`, native fixture tests | KEEP physical process/port/proof tests and UI-review shutdown coverage; replace only broker-specific setup and assertions. | Mock broker tests alone cannot prove no leaked listener or borrowed-service termination. Cover pending acquisition, owner death before readiness, late completion, cleanup failure and explicit recovery at the responsible owner. |
| Pending `observability/tests/wrapper-fixture.mjs` | REPLACE the copied former coordinator with small authored protocol documents and explicit tampering fixtures. | Moving a complete former implementation under a test directory preserves its maintenance burden. Retain independent native receipt rejection and multiple/nested graph tests without a second state machine. |
| `test-support/execution-channel-fixture.mjs`, retained observation/qualification tests | KEEP small data builders and independent arithmetic/security readers; prune unused helpers/imports. | Fixtures may construct bounded input records; they must not reproduce scheduling, ownership, aggregation or validation algorithms. CPU equality at 2.5% still fails; all other qualification rules stay intact. |
| `observability/tests/test-collection-engine.mjs` | MOVE its remaining process-exit notification test to the runtime process-owner tests, then remove the misleading empty collector-suite shell and registration. | The Node owned-process producer helper still has callers and deserves its regression; it does not justify retaining the removed JS collector. |
| `task_surface_owner.json`, `test_families/harness.command_surface.json`, `test_families/harness.browser.json`, `harness_helper_ownership.json`, generated Make/topology and task-surface tests | UPDATE as one transaction. Give each semantic check one route per aggregate; keep a bounded native smoke for the existing scheduler/fixture smoke roles and separate full contract/boundary runs. | Current smoke/matrix substitutions call the same full suite repeatedly. Remove that duplication and stale source-path assertions. Preserve exact product row selection; internal helper routing can change coherently without compatibility aliases. |
| Deleted JS graph runtime/broker/collector bridges and standalone `tools/harness{admission,collect,diag}` roots | COMPLETE caller migration, then accept removal with Make/build variables, registrations and dead tests. Archive at the preserved S-030 revision. | No dual-engine fallback, forwarding modules or hidden old binary dependency. Read-only status requires an installed binary and never builds it. Rollback restores a complete archived transaction. |

Tests have three roles. Native owner tests prove scheduling, cache, failure and
fixture rules cheaply with controlled inputs. Node compiler/adapter/reader tests
prove source interpretation and format semantics independently. A small real
boundary suite proves those components cooperate, including actual process
reaping, service ownership, observation quietness and retained publication. Do
not make every compiler or parser assertion launch a complete nested graph.
Conversely, do not replace process cleanup or private-channel tests with mocks.
Classify real compiler/selected-graph tests such as `TestRuntimePreparationParity`
and `TestRuntimeSelectedGraphExecution` accurately in authored routing; they are
boundary work even while their files share the runtime package. Split their
route from the cheap smoke without introducing test-only public runtime commands.

The following workstreams refine WF-31/WF-32; their order is authoritative within
those existing slices. They introduce no parallel tracker or new qualification
phase. Save a tracker checkpoint after each before its successor.

| Order / existing slice | Work and affected areas | Entry and exit gate | Risk and response |
| --- | --- | --- | --- |
| 1 / S-031: Freeze behavior and test ownership | Archive the current dirty candidate separately from S-030. Review the changed/deleted test groups against the table above; record keep/move/retire plus the behavioral reason. Reconcile only any genuinely affected owner clauses and typed test routing. | Entry: S-030 checkpoint and preserved primary index. Exit: every changed test group has a named owner and meaningful acceptance obligation; discarded implementation details are distinguished from required behavior. | A large deletion hides lost coverage. Audit behavior groups, not line counts or one-to-one legacy test names. Do not manufacture a new coverage framework. |
| 2 / S-031: Complete one candidate transaction | Finish caller/import cleanup, remove the JS simulation and copied wrapper coordinator, finish the thin compiler boundary and private producer interfaces, consolidate binary dispatch/builds, and simplify the existing test routes. Regenerate from authored inputs. | Depends on step 1. Exit: no active old engine/broker/helper root, fallback or duplicated test scheduler; compiler/row/reader boundaries remain; generated files match authored inputs and every registered test/command resolves. | Generated recipes or mixed tests still reference removed paths. Fix the responsible projection/test; do not restore dead production APIs just to satisfy old imports. |
| 3 / S-031: Prove functional boundaries and close V-032 | Run the ordered narrow and boundary gates below on this candidate. Repair only concrete failures. Record the final deletion/test-disposition map, exact commands, failures and source identity, then commit the coherent S-031 code transaction separately from the preserved primary staging. | Depends on step 2. Exit: all owner contracts and real cross-process/public Make acceptance pass on final source; no required invariant has only a removed test. Save S-031 completion before S-032. | Passing mocks or an earlier smoke conceal a broken public path. Require current-source public results and independent retained reading. Each failed attempt remains retained. |
| 4 / S-032: Finalize and hand off | Run finalization, broader functional verification, a full warm check and exact-run diagnostic review. Record source digest, rollback/archive references, migrations and physical coverage limits. | Depends on saved V-032 completion. Exit: V-033 passes and the final checkpoint clearly separates functional salvage from deferred certification. | Repairs after validation invalidate affected evidence. Rerun the failed/affected boundary first; broaden again only when its risk requires it. Never relabel old passes. |

V-032's acceptance is behavioral. Preserve independently calculated scheduling
and accounting cases; current-input/cache rejection; all required failure classes
and primary/secondary precedence; duplicate JSON/digest/path/size rejection;
owned-versus-borrowed cleanup; and unavailable-versus-zero diagnostic semantics.
Retain one real four-collector exclusive-window/drain test under the native
observation owner, not four copies of the legacy worker implementation. Cover
Go/Node/Make failure handoff, child readiness then cancellation, partial browser
acquisition/owner loss, and graph/UI-review fixture use. A removed scenario whose
behavior is already independently covered does not need another integration test.
A scenario exposing a distinct untested boundary must be retained or replaced.

Run narrow routes first: `harness-runtime-contract`, `harness-fixture-contract`,
`harness-evidence-contract`, `harness-command-surface-contract`, and the compiler/
runner owner routes discovered through `task-guide`. After projection edits run
`generate`, `generate-drift`, `generated-artifact-policy-check`, `json-shape-check`.
Then run `harness-runtime-boundary-contract`, `harness-ui-review-contract` and
`harness-ui-review-lifecycle`, and a fresh public `test-slice` selecting
`harness.evidence_accounting.behavior.local_migration_activities` with basic
diagnostics. The boundary route covers real schema bootstrap and HTTPS trust.
Use the existing Go/Vitest/shell routes to demonstrate each retained adapter;
reuse current-source coverage from the appropriate gate rather than adding a
second blanket matrix. Run these through Make in the isolated checkout, with a
unique environment-supplied `CARTULARY_TEST_RUN_ID` per invocation.

For an exact fresh result directory require `harness-observability-check`,
`test-run-status JSON=1`, and human/JSON `explain-run` performance/resources views.
Check useful phase/wait/resource facts and honest omissions, not just successful
exit or JSON shape. Exercise diagnostic failure without hiding the primary test
result, and export isolation without external telemetry. Demonstrate default-off
execution and cache current-input behavior. No timings from these functional
runs establish an overhead claim.

S-032 keeps the existing final gate: `agent-finalize` before broader
`harness-contract`, script/shell/Markdown lint, applicable drift and full warm
`check`. Supply `RESULTS_DIR` to finalization only for an already eligible full
warm successful check; otherwise record maintenance skipped. No automatic rerun
loop: classify a failure, fix its owner, rerun the narrow affected gate, and
preserve the failed root. A failed functional gate leaves S-031 or S-032 open.
F-022/RB-006 and the original qualification work remain unresolved/deferred even
if every functional gate passes.


### Prior iteration — historical scope and evidence


S-014–S-022 retain the underlying implementation authorization. This invocation
implements the recovery. The readiness column tracks the serial
prerequisite gates independently of authorization. Historical execution checkpoints
remain evidence of their own source revisions.

| Slice / kind | Findings / contracts / dependency source | Intended change / entry gate | Readiness / tests | Migration / rollback / completion |
| --- | --- | --- | --- | --- |
| S-014 adopted owner amendment | F-014–F-018; C-010–C-014; WF-14 | Revalidate baseline; declare required/nullable fields, clock/receipt/ownership rules, caps, qualification formulas and acceptance before machine changes | DONE; V-014 and RB-004 closed at recorded checkpoint | Specs/tracker only; no generated change. Exit: owner review resolves semantics and conflicting old clauses |
| S-015 characterization/routing | F-014–F-018; C-001–C-014; WF-15 | Inventory direct/dynamic wrapper, helper, reader and owner routes; independent baseline and future expected fixtures; freeze qualification workload/reader schedule in typed inputs | DONE; V-015 and RB-005 closed at recorded checkpoint | Test/routing transaction; no golden preservation of defects. Exit: executed owner-conforming baseline and complete explicit scenario map |
| S-016 separate validation correction and structural consolidation | F-018; C-014; WF-16 | Share validated facts; fix malformed-envelope acceptance; migrate checker/explanation and loader consumers | DONE; V-016 checkpoint retained | No retained format cutover solely for refactoring; revert loader/callers together, no forwarding aliases. Exit: one resource traversal and uniform rejection/parity |
| S-017 adopted clock extension | F-014; C-010; WF-17 | Owner-produced bracketed anchors, bounded uncertainty and local activity placement; preserve canonical/local durations | DONE; V-017 checkpoint retained | Version incompatible execution shapes; producers/minimum readers/generated outputs in one transaction. Exit: arithmetic and unavailable/discontinuity/canonical parity cases pass |
| S-018 adopted container extension | F-015; C-011; WF-18 | Lifecycle proof to collector-driven, run-owned Go helper using existing client; resource facts and bounded views | DONE; V-018 checkpoint retained | Version resource shapes; lifecycle stays at owner. Revert helper/producers/readers together. Exit: real owned service plus replacement/borrowed/conflict/cleanup cases pass |
| S-019 adopted wrapper/receipt extension | F-016; C-012; WF-19 | Supervise managed Make child; retain pre-scan closure; publish independently validated, scan/envelope-bound terminal receipt | DONE; V-019 checkpoint retained | Version envelope/add closed receipt; rollback wrapper/protocol/readers atomically. Exit: lock/signal/exit/privacy and incomplete-publication tests pass |
| S-020 adopted observer measurement | F-017/F-022; C-013; WF-20 | Scoped component facts plus external total incremental measurement; execute the recovery decision sequence below | IN_PROGRESS; A–B DONE; revision-27 C controls failed (Go canonical p90, shell missing RSS); D row-worker rejected/reverted, metadata-fixture repair accepted; E blocked. Historical evidence is not qualification | Version affected qualification inputs/evidence together; preserve canonical manifests. Exit: V-020 plus measurement/control/repair gates and saved completion checkpoint |
| S-021 engineering qualification | F-017/F-022; C-013; WF-21 | Execute retained frozen matrix and typed-budget evaluation, with all failures/exclusions retained | IN_PROGRESS; candidate 16 failed; new collection blocked on S-020/A–E and clean freeze; RB-006 remains open | No baseline refresh/default change. Failed budgets or inadequate uncertainty retain named findings. Exit: every required cell passes |
| S-022 final verification/handoff | F-014–F-019; C-001–C-014; WF-22 | Finalize before broader verification; reconcile final source, generated/import changes, failures/skips, migrations and evidence | blocked on WF-21; V-022/V-023 | Source repairs remain separate from validation; invalidate affected qualification cohorts. Exit: selected gates and durable handoff complete, default still off |

Rollback is a coherent slice transaction across authored inputs, generated
projections, producers and consumers. Preserve pre-existing work and retained
evidence. Disabling diagnostics contains an issue but does not complete a required
correction. Every format-changing slice ships minimum readers at the same time;
S-022 cannot be used to repair intentionally unreadable intermediate artifacts.

### Systematic F-022 recovery plan — planning reset, 2026-10-11

This replaces the next-action recommendations of the candidate-16 attribution
ledger. It preserves S-020 → S-021 → S-022 and their existing owners and validation
IDs. A–E below are decision checkpoints within S-020, not replacement workstreams.
Save the result of every checkpoint, including a rejected hypothesis, before its
successor. The latest request is to plan; no experiment below has run.

#### Diagnosis of the unsuccessful approach

We repeatedly optimized plausible components and then used an approximately
hour-long matrix as the next diagnostic. Several structural improvements are
valid, but their functional success or lower component cost did not establish
that they addressed the failures deciding qualification. The investigation has
not yet separated actual incremental observer cost, normal workload variation,
measurement disturbance, and uncertainty in the retained measurement.

The current off/zero-reader cell compares each observation to itself. Its zero
cost and passing gates do not demonstrate that two independent identical runs
are reproducible. Independent arithmetic on retained summaries checks the gate
formulas; it does not independently validate acquisition or recover discarded
memory samples. This is a missing experimental control, not proof that every
historical measurement is wrong.

E-026 independently rechecks candidate-16 result digests and derives:

| Workload | Median reported support CPU / outer duration, % of one core, off/0 | Observed off/0 canonical range, ms | Observed off/0 RSS range, MiB |
| --- | --- | --- | --- |
| Go | 13.483 | 310–322 | 122.47–127.64 |
| Vitest | 14.406 | 2344–2415 | 383.54–395.34 |
| Shell | 14.369 | 19400–19581 | 672.80–690.23 |
| Service | 12.837 | 12660–14320 | 390.33–403.91 |
| Browser | 14.992 | 23117–23621 | 648.52–663.20 |

These support figures include the reported supervision boundary and its allowance;
they are not direct steady utilization measurements, observer-budget charges, or
proof of causal interference. They make testing the measuring instrument a higher
priority than another few-millisecond schema optimization. Native measurement
currently wakes every 2 ms, walks `/proc` and selected process resident pages at
100 ms cadence, and observes gate/parent support too; the parent checks readiness
every 25 ms. Identical support code in both treatments does not establish identical
disturbance when their process populations and memory differ.

All five failing RSS upper estimates have medians below 32 MiB:

| Failed cell | Paired median MiB | Three paired MADs MiB | Declared error MiB | Upper estimate MiB |
| --- | --- | --- | --- | --- |
| Shell off/1 | 16.965 | 17.130 | 0.098 | 34.192 |
| Shell basic/0 | 25.182 | 19.350 | 0.109 | 44.642 |
| Shell basic/4 | 21.057 | 31.818 | 0.105 | 52.981 |
| Browser basic/0 | 20.587 | 18.524 | 0.207 | 39.318 |
| Browser basic/4 | 20.544 | 14.342 | 0.227 | 35.113 |

The failures remain valid under their policy; this decomposition identifies the
problem to investigate. With five or six observations, nearest-rank p90 is the
maximum. Removing a unique failing extreme can change that gate and trigger
instability even when elapsed-window MAD is below 10%. No p90 or stability rule
is changed by recognizing this property.

#### Paths already tried and limits on repeating them

| Path | Established result | Next disposition |
| --- | --- | --- |
| Native public reader and collector; shared generated root validators; selected-row discovery; bounded source hashing; parent-free syntax | Adopted and functionally verified, yet candidate 16 still fails | Keep their evidence; change them only if new attribution identifies a material remaining cost |
| Private publication/fsync and boot-clock calls | Retained attribution shows small costs; `/tmp` and qualification output already use RAM | Do not weaken publication or introduce another RAM migration; RAM addresses wall time |
| Collector initialization/process spawn | About 2.8–3.0 ms CPU including about 1.6 ms spawn in the tested workloads | Do not move lifecycle ownership on the rejected large-fork-cost hypothesis |
| Trivial schema inlining and earlier fragment sharing | No material end-to-end repair established; experiments restored and retained | Do not repeat without a newly identified hot path and predicted gate effect |
| Repeated complete matrices | Exposed real correctness issues but repeatedly failed elapsed/RSS/stability after CPU improvements | Full qualification becomes the final admission gate, not the next exploratory experiment |

#### Ordered decision checkpoints inside S-020

| Checkpoint / readiness | Dependency | Work and explicit output | Binary exit / stop rule |
| --- | --- | --- | --- |
| S-020/A — policy and experiment contract; DONE | Saved candidate-16 rejection and this plan | Adopt authorized CPU mean <2.5% in REQ-848 and a new plan version, with equality/boundary tests and coordinated generated projections. Freeze hypotheses, fixed experiment counts, source, preparation, analysis and expected signatures before measurements. | Owner and projections agree; old plan/evidence retains its original identity; all other budgets, workload selections and qualification schedules are unchanged |
| S-020/B — validate the measuring instrument; DONE after descriptor calibration | A | Audit acquisition, retain reconstructible evidence, add independent lifecycle/memory calibration, and test support disturbance using the bounded design below | Every claimed CPU/RSS component and bound can be independently checked; missing/racing observations fail closed. A discovered meter defect is repaired and calibrated before C |
| S-020/C — identical-treatment controls; FAILED, retained | B | Fixed six independent off/0-versus-off/0 pairs per workload, split into the sentinel and remaining batches below; retain order, preparation, phase and support evidence | Record every control result. Failure of unchanged elapsed/RSS/CPU/stability rules blocks a full matrix and selects a specific repeatability/measurement repair; do not repeat until lucky |
| S-020/D — causal CPU/memory repair; fixture repair accepted, residual causes unresolved | C, including its recorded pass or failure and attributed cause | Select one responsible boundary using the decision table below; implement one coherent repair; repeat only its predeclared affected diagnostic experiment and inspect all other metrics | Benefit exceeds that experiment's measured repeatability/error, explains the targeted excess, and introduces no coverage, security, timing or cleanup regression. An unattributed result is inconclusive, not success |
| S-020/E — functional closure and qualification admission; blocked pending B–D | B and C passing after any D repair, and D's retained disposition | Remove optional attribution probes; run owner routes, projection gates and exact-run readers; review the complete executable delta and source posture | V-020 and relevant V-014–V-019 gates pass; no unexplained failing control remains; save S-020 DONE before clean freeze and S-021 |

C failure is useful evidence and permits its selected D repair, not S-021. If D
changes the meter or workload lifecycle, repeat only the invalidated B/C controls
under a new experiment identity before E. This is an explicit repair loop; a
failed checkpoint never silently satisfies a dependency. There is no requirement
to invent a production optimization if a calibrated measurement repair is the
only demonstrated cause. Conversely, passing controls do not establish observer
qualification.

#### New path 1: measurement integrity and disturbance

Keep the native supervisor's ownership boundary and existing public Make routes.
Extend its closed evidence, rather than creating a general-purpose profiler or
runner. Keep raw aggregate samples with timestamps/sweep spans, sample counts,
root/reap accounting, coverage/losses and support measurements sufficient to
reconstruct each accepted summary. Bind a bounded, versioned trace to the result
with a digest; validation reconstructs the RSS integral and coverage instead of
trusting only the stored scalar. Prototype per-process role/lifecycle detail only
in attribution evidence when aggregate traces cannot answer the question.

Use opaque start identities, closed role tags and numeric counters. Retain no
command lines, environment, arbitrary paths or Docker credentials. Define byte,
record and lifetime caps, atomic publication, unsafe-path rejection and failure
outcomes before the first publishing change. These are outer qualification
artifacts; they do not broaden the workload's post-scan receipt exception. Publish
traces after the measured command, measure/disclose support cost, and use the same
bounded mechanism in both treatments. Version affected result/manifest shapes
coherently; no historical raw samples can be reconstructed from summaries.

Review two specific code seams without assuming a bug: PID/start identity can
change between population discovery and resident-page reading; page-quantization
error alone does not demonstrate that a sequential process sweep represents a
simultaneous sample within a bounded error. Preserve the sampled-RSS estimand and
100 ms cadence. Independently test overlapping/non-overlapping touched allocations,
process exit/replacement during a sweep, short-lived children, worker threads,
reaping/adoption and the central-80% integral. Audit existing waited-descendant
tests first; extend missing cases rather than redo demonstrated coverage. Use
kernel/self counters and authored numeric fixtures as independent expectations,
not the production aggregation algorithm. Unbounded error remains ineligible even
if a correct bound makes qualification harder.

Run one fixed diagnostic comparison of normal support against accounting-only
support, using identical workload commands, off mode and zero readers. Begin with
Go and shell, six counterbalanced pairs each; the latter variant has no RSS
qualification claim. It retains external monotonic timing and CPU accounting,
changing only support work for this causal experiment. If this exposes material
disturbance, prefer event-driven child completion and deadline-driven observation,
and proven-identity discovery improvements, while retaining 100 ms qualification
sampling, exact wait accounting, bounded finalization and cancellation. Do not
replace resident-page semantics with approximate counters merely to lower cost.
If the effect is within observed variation, record that limit and move on; support
cost alone is not a reason to rewrite it.

#### New path 2: identical-run controls and treatment-order effects

The current self-reference is replaced only in the diagnostic experiment by two
independent off/0 executions, labelled A and B. Use fixed paired order balanced
between A-first and B-first, predetermined before collection. Preserve ordinary
commands, cache posture, fresh service ownership and results storage. Retain one
discarded warm-up for each label, then exactly six pairs per workload; no padding,
extra warm-ups, changed cadence or selective exclusions. Begin with Go and shell
(28 commands including warm-ups); then cover Vitest, service and browser (42).
The diagnostic labels/order are separate from the unchanged seeded qualification
matrix. Route this closed control through authored measurement-owner tests; add
no arbitrary-command option to the public qualification runner.

Apply the existing numerical and leave-one-out rules descriptively to these
independent pairs, with the authorized CPU policy. Record position in the pair,
elapsed phase changes, memory support spans and service creation/readiness/cleanup
durations. Do not subtract a learned noise estimate from qualification results.
Do not borrow A/A pairs as qualifying blocks. If the identical-run controls fail,
identify whether the failure follows sample phase, process lifetime, service
readiness/cleanup, allocation/GC, or measured support. The user reports no other
intensive work; external contention is not an available default explanation.

#### New path 3: role and lifetime attribution, then structural repair

Build a reconciliation view with workload, graph coordinator, row runner,
observer helper, public readers and external support as distinct roles. Include
only proven non-overlapping CPU accounting; exit self-counters and observed
samples supplement rather than replace complete wait accounting. For memory,
plot each process's residency over the actual command timeline, mark the 10%/90%
boundaries and lifecycle phases, and reconstruct the original sampled integral.
Any unclassified contribution remains explicit. Distinguish retained objects,
temporary allocation/GC, changed process overlap, and different phase weighting
caused by runtime length. Do not phase-normalize the qualification metric.

| Observed signature | Preferred structural path | Areas / long-term benefit | Risk, migration and acceptance |
| --- | --- | --- | --- |
| Meter variation or missing/error-prone samples dominate controls | A bounded measurement trace and calibrated acquisition; reduce demonstrated support work at its native owner | Specification, typed evidence, measurement, tests and docs; reproducible independent accounting | New evidence version and new cohorts; retain security/coverage/cadence. Accept only reconstructed accounting and calibrated bounds, never a smaller unsupported error claim |
| Graph/runner allocations remain live beyond discovery and explain RSS or GC variation | Separate compilation lifetime from execution data; retain only the immutable selected execution facts and required cache inputs | Harness implementation and tests; smaller retained state and clearer ownership | The current graph runner retains `compiler` through a cache callback. Inspect all consumers and retaining paths before choosing scope release or a disposable preparation process. Count any new process fully; reject a fix whose IPC/startup cost consumes the gain. No global cache or skipped catalog validation |
| Residency is small but process overlap differs materially | Shorten the responsible process's actual ownership lifetime and release references/resources promptly | Runtime/runner implementation and lifecycle tests; reduces overlap without changing semantic work | Preserve signal propagation, admission descriptors, waits and cleanup. No forced GC, worker reuse, concurrency reduction or delayed accounting just for a pass |
| Service readiness/cleanup accounts for isolated elapsed extremes | Repair a reproduced readiness/lifecycle defect at its existing owner | Test-service implementation and tests; removes variable unnecessary work for every caller | Existing single-master and forced-drop changes are already tried. Require a new causal trace. Do not change reuse posture, remove required cleanup, retry away slow runs or loosen timeouts |
| Controls pass, but basic-only parent/publication work is repeatably material | Eliminate proven duplicate work across an already trusted immutable boundary, or use a bounded owner-local batch | Producer/reader implementation, privacy and parity tests; lower repeated work without a new authority layer | Never remove validation across a trust boundary or exceed publication/pause/stop bounds. Require a complete CPU/RSS/finalization comparison, not an isolated microbenchmark |

The graph lifetime path is a hypothesis, not a claim of a leak: process RSS may
remain resident after references disappear. Its entry evidence must identify a
retaining path and demonstrate an actual residency benefit. No scheduler redesign,
permanent diagnostic daemon or unbounded memoization is selected in advance.

#### Work limits, full qualification and handoff

Before each diagnostic experiment record one hypothesis, fixed count, source,
signature, expected budget impact and a stop decision. Permit one baseline and
one changed variant per hypothesis. At 90 minutes of experimental collection
without a reconciled cause, save a decision checkpoint and reassess the design;
do not automatically launch another full cohort. This time box ends exploration,
not required child cleanup or an already admitted full S-021 collection. A failed
or inconclusive result remains retained and cannot justify a candidate freeze.

S-021 entry requires A–E and a saved S-020 completion checkpoint. Use the existing
isolated clean checkout; preserve the primary index. Freeze the final executable
commit and new CPU plan, then run the full five-workload × six-treatment matrix
once with its cold/warm-up/five-or-six-block protocol. No reused cells, cadence
changes, affinity/capacity tuning, threshold relaxation, workload lengthening,
extra idle time, or discarded unfavorable observations. Validate the exact manifest
through `make harness-performance-check EVIDENCE_ROOTS_FILE=<exact-manifest>` and
independent arithmetic/raw-evidence reconstruction. A failure reopens its named
owner work with retained evidence; it never triggers an unchanged retry.

If calibrated identical-treatment controls still cannot meet the fixed rules and
no owner defect explains them, report that the present WSL2 setup has not
demonstrated qualification feasibility. Any different environment or statistical
contract would require a separate explicit decision; this plan changes neither.
Do not promise S-021 success from a CPU threshold increase. S-022 stays gated.

After S-021 passes, save its checkpoint and perform S-022: `make agent-finalize`
first, then required broad contracts/lints and projection checks. Without an
eligible successful full warm `check`, leave `RESULTS_DIR` unset and record
retained-run maintenance as skipped. Add this tracker to Markdown-lint coverage;
review whether that configuration affects executable identity before labeling
evidence. Keep the exact qualified commit/digest distinct from later documentation
commits. Any executable repair invalidates affected cohorts. Finish only after
V-022/V-023, the complete failed-evidence ledger, migrations, physical limits,
rollback instructions and final S-022 checkpoint are saved.

### Selected architecture and interface decisions

- Runtime owns clock anchors; producers own operation boundaries; observability
  owns validated joins. Derive bounded offset intervals from bracketed local
  monotonic/boot-clock observations and retain uncertainty in placement. Reject
  identity mismatches, reversal and discontinuity; unavailable placement does
  not erase a valid local duration. No UTC-derived elapsed time, generic clock
  synchronization, interpolation across suspend/gaps, or second scheduler clock.
- Suite-service ownership supplies exact private container proof to a bounded,
  run-owned helper using the existing Go Docker client. The helper responds to
  collector cadence with no independent polling schedule, discovery of unrelated
  containers, privilege expansion or cleanup authority. Pause/stop drain requests
  before quietness acknowledgement. Raw IDs/labels remain private; CPU, memory,
  I/O and PID counters retain their actual semantics. Processes, containers,
  cgroups and guest scopes are never added into an invented total. Shared/borrowed
  service observations remain contextual; replacements receive fresh identities.
- Supervise the managed Make child from workspace-wrapper entry through child
  exit, preserving inherited admission descriptors, nested invocation semantics,
  signal forwarding and exit precedence. Initial shell/Make startup and receipt
  publication itself remain explicit exclusions. Multiple graph runs share the
  outer invocation identity; views must not each charge its duration additively.
- Ordinary producers drain and publish before the retained-secret scan. After
  child completion, the wrapper may publish only the separately adopted bounded
  structural receipt: identities/digests, numeric timing and closed outcomes,
  bound to the scan result and pre-scan envelope. No free-form messages, environment
  values, restarted sampling, rewritten scanned artifact or arbitrary post-scan
  output is permitted. Missing/failed scan or child completion cannot imply a
  successful complete receipt. Independent secure loading validates the exception.
- Secure loading and integrity closure return immutable normalized facts once
  per command. Analysis/formatting consume them without rereading files; ordinary
  readers remain independent of OTLP construction. Malformed published evidence
  fails diagnostic validation without rewriting retained test outcomes.
- Component self-counters and external incremental cost remain distinct. Include
  module initialization, parent processing, admission helpers, container sampling,
  serialization, publication and public readers in the declared observer scope.
  Count each physical process once; a worker thread adds no second process RSS.
  Measure simultaneous sampled memory separately from lifetime maxima. External
  measurement support is identical in both treatments and reports its own cost,
  resolution and coverage. Missing cost is unavailable, not zero.
- Extend existing public commands. A separately typed instrumentation-qualification
  manifest is accepted through `harness-performance-check EVIDENCE_ROOTS_FILE=...`;
  canonical performance portfolios and baseline publication remain unchanged.
  Keep ordinary explanations descriptive. No export, optimization, concurrency
  change or baseline refresh is implied by engineering qualification.
- S-014 adopts exact closed projection versions, required/nullable fields, new
  cardinality limits and rejection rules before generation. Preserve existing
  64 KiB record, 1 MiB pending and 4 MiB metadata limits, with explicit container,
  anchor and observer record bounds rather than overloading process counts.
  Version incompatible shapes and migrate producers/minimum readers/attachments/
  generated outputs together. No legacy translators or forwarding aliases.

### Completed slices — historical intent, exits and rollback rationale

S-000–S-013 remain DONE; their earlier authorization and tests do not authorize
S-014–S-022 or qualify its completed source. These rows retain prior stable IDs.

| Slice / kind | Findings / contracts / dependencies | Intended change and entry conditions | Readiness / validation | Completion / rollback |
| --- | --- | --- | --- | --- |
| S-000 specification/tracker | F-004/F-008/F-009; WF-00 | Adopt REQ-840/AC-137 before machine projection edits; preserve historical evidence. | DONE; S-000 Markdown gate | Owner/projection responsibilities and bounded scope agree; no generated changes in this slice. |
| S-001 characterization | F-001–F-006; C-001–C-005; WF-01 prerequisites | Extend owner-routed tests with independent accounting expectations, lifecycle interleavings, wrapper/check/export parity and vocabulary acceptance fixtures. Revalidate exact caller inventory at implementation entry. | DONE; S-001 checkpoint, V-002/V-003/V-005 | Fixtures specify expected outputs without calling the production algorithm to derive them; no production behavior change. Revert test-only slice if withdrawn. |
| S-002 shared definitions | F-004; C-001, C-003; WF-02 prerequisites | Add `tools/harness_instrumentation_definitions.json`, its owner-input schema, generated shared schema definitions and runtime descriptors; update authored attachments, generated policy and generator wiring as one transaction. | DONE; S-002 and S-006 generation closure, V-004 | Old valid/invalid fixture acceptance unchanged; no runtime AJV or second vocabulary registry. Roll back authored inputs and generated outputs together. |
| S-003 canonical accounting | F-003; C-001; WF-03 prerequisites | Change `canonical-unit-events.mjs` to normalized facts; adapt `performance-projection.mjs`, `writeCanonicalArtifacts` and `validateCanonicalRun` through `evidence-accounting/index.mjs`. Remove `retainProjection` and duplicate union code. | DONE; S-003 checkpoint, V-002 | Exact expected accounting and errors; no duplicate raw-event storage; bounded state documented. Revert all internal callers with the reducer; no compatibility overload. |
| S-004 collection lifecycle | F-001, F-002, F-006; C-002, C-003; WF-04 prerequisites | Add owner-local `collection-engine.mjs`, `resource-store.mjs`, `diagnostic-session.mjs`, and runtime `observation-gate.mjs`; thin controller/worker; inject Linux discovery and clock; move run composition behind session. | DONE; S-004 checkpoint, V-003 | Control races, shutdown, truncation and fencing tests pass; one worker and one writer remain; pre-scan order preserved. Revert engine plus adapters/wiring together; `off` disables optional collection. |
| S-005 retained analysis and deletion | F-005; C-003–C-005; WF-05 prerequisites | Separate secure loading, pure resource accumulation, explanation formatting and OTLP projection; update check/export and wrapper consumers; prune unused exports. | DONE; S-005 checkpoint, V-005/V-006 | Public outputs/exits/security unchanged, eager OTLP removed from checks, private dead paths gone. Revert consumer transaction without aliases or dual readers. |
| S-006 validation/handoff | All in-scope findings; WF-06 after WF-05 | Finalize, verify final source, reconcile failures and hand off deferred work. | DONE; final verification ledger | Required gates pass; historical evidence and default-off/deferred limits retained. |

| Slice / kind | Findings / contracts / dependency source | Intended change / entry gate | Current readiness / tests | Migration / rollback / binary exit |
| --- | --- | --- | --- | --- |
| S-007 specification amendment | F-010–F-013; C-006–C-009; WF-07 | Adopt owner semantics and acceptance mappings before schemas/producers; review exact definitions, policy and schema attachment changes | DONE; executed S-007 checkpoint and assigned validation | Owner amendments only in this slice; no generated changes. Exit: agreed inputs/outputs/defaults, privacy/bounds, failure and clock rules |
| S-008 characterization | F-010–F-012; C-001–C-009; WF-08 | Inventory direct/dynamic launch consumers and owner-routed tests; baseline failure envelopes, packed runners, allocation reuse, local measurements and explanation | DONE; executed S-008 checkpoint and assigned validation | Tests/fixtures and routing together; no defect blessed as compatibility. Exit: independent expected values and passing baseline, with corrections recorded |
| S-009 structural identity extraction and adopted extension | F-010; C-006; WF-09 | Introduce neutral runtime identity; migrate all selected launch adapters and failure-channel construction together | DONE; executed S-009 checkpoint and assigned validation | Revert context and callers together; no old-interface aliases. Exit: fresh parent-linked attempts across success/spawn failure/cancellation with unchanged public outcomes |
| S-010 adopted correlation extension | F-011; C-007/C-009; WF-10 | Create allocation identity before acquisition; observe leases/reuse/failures/process replacement without transferring lifecycle authority | DONE; executed S-010 checkpoint and assigned validation | Revert broker/providers/observation/schema callers together. Exit: exact relationship closure, shared/borrowed context preserved, required cleanup unchanged |
| S-011 adopted producer timing extension | F-012; C-008; WF-11 | Reuse migration/reset clocks and facts; add failure coverage and actual report parsing boundaries; closed catalog declarations | DONE; executed S-011 checkpoint and assigned validation | Producer/catalog/generated changes form one transaction. Exit: local timing fixtures pass; incomplete work stays incomplete; canonical summaries unchanged |
| S-012 retained reader/presentation extension | F-011/F-012; C-003/C-004/C-007–C-009; WF-12 | Validate observations and references before publishing normalized analysis; extend bounded public explanation | DONE; executed S-012 checkpoint and assigned validation | Incompatible formats cut over with all consumers; no compatibility translator. Exit: security, integrity, bounds, exact selection/exits and export-isolation fixtures pass |
| S-013 validation and handoff | F-010–F-013 and preserved foundation invariants; WF-13 | Complete V-012 on final source, classify failures, reconcile generated/import/documentation changes, save checkpoint | DONE; final S-013 checkpoint, current-source V-012/V-013 gates | No source rollback hidden inside validation. Exit: every selected gap has evidence; residual coverage/qualification limitations and next entry are recorded |

## 8. Validation Plan

### Current lifecycle validation

V-037–V-043 map in order to S-036–S-042. Use Make-owned runtime, fixture,
runtime-boundary, evidence, command-surface, UI-review contract/lifecycle/seeded
routes. Preserve three-role fast smoke; broad native suites remain owner gates.
Extend authored package selections before extraction and regenerate projections.

Required cases: ownership/permissions/containment; symlink/hardlink/inode changes;
malformed proof; cross-chunk secrets; registration versus finalization; aged live
owners; partial/late acquisition; cancellation before readiness; repeated/concurrent
stop; stalled independent cleanup; parent/controller death and PID reuse; exact
resource dependencies; borrowed protection; primary/secondary publication failure;
immutable terminal repetition; no late writes; inherited/owned/persistent callers;
quiet capture and frontend staging failure. Status neither provisions nor writes.

After authored changes: `make generate`, then separately `make generate-drift`,
`make generated-artifact-policy-check`, `make json-shape-check`. S-042 runs
`make agent-finalize` before final `make harness-contract`, applicable lint,
projection checks and full warm `make check`. Supply finalization RESULTS_DIR only
for eligible successful full-warm evidence; otherwise record its maintenance skip.
Every writing invocation uses a fresh explicit run ID. Record process exit
independently of harness summary; inspect exact-run human/JSON evidence and local
export isolation. No qualification acquisition is scheduled.


### Previous functional iteration — current validation and operating contract (historical)

| Signal / operation | Stable semantics and boundary | Acceptance |
| --- | --- | --- |
| Canonical events, queue-inclusive path and phase intervals | Scheduler owns sequencing; local activity durations remain independent; nested/parallel categories are not additive | Independent arithmetic, failure/cancel closure and unchanged row/result classifications |
| Process resources | 2 s cadence, observed CPU deltas and sampled RSS; identity reuse resets attribution | State lower-bound/partial coverage; never sum independent peaks or overlapping scopes |
| Context and exact containers | 5 s cadence; one batch, 250 ms deadline, 32 live/256 lifetime containers; exact private lifecycle proof | CPU, memory, block I/O and PID semantics preserved; unavailable/reset/timeout explicit; no unrelated discovery |
| Clock placement | Two bounded anchors/activity; boot/time namespace compatible; uncertainty width at most 25 ms | Reject placement on discontinuity/mismatch; preserve valid local duration |
| Receipt | At most 64 KiB/64 graph memberships; bind envelope and scan; one shared outer duration | Signals/exits and primary failure unchanged; late writes/forged bindings fail |
| Readers | Exact `RESULTS_DIR`; strict JSON, secure files, digest/reference closure; one resource traversal | Immutable facts; analysis/presentation perform no reads; human/JSON expose consistent facts and omissions |
| Export | Existing explicit boundary only | Local conformance/fixture validation; no external transmission or newly admitted labels |
| Qualification | CPU mean <2.5%; canonical and outer median ≤2%, p90 ≤3%; RSS median + 3 MAD + error ≤32 MiB; finalization maximum ≤500 ms | Signed/equal-weight arithmetic, eligibility/stability/integrity unchanged; no pass without complete eligible evidence |

V-034 records the completed S-033 baseline cleanup. V-035 and V-036 close
presentation and specification reconciliation. Current-source V-028–032 are
complete through E-035–E-039; historical candidate passes were not substituted.
The native graph and shared fixture policy are integrated, UI-review migration
and real lifecycle suites pass. The final full warm product check passed after
successful finalization: 981/981 units, owning exit 0, exact canonical/scan audit
PASS. V-033 closes through E-040. Qualification/experimental acquisition remain
deferred; no functional result is relabeled as performance certification.

### Salvage gates — historical off-main salvage

New permanent native-runtime and fixture Make contract routes belong to existing
`harness.command_surface` and `harness.browser` owners. Independent fixtures cover
cycles/dependencies/selection, locks/capacity/aging, exclusive observation gates,
simultaneous failure, cancellation and lingering descendants; cache hit/miss/off,
corruption and current-input changes; owned/borrowed fixture recovery and UI-review
shutdown; duplicate JSON, forged identities/digests, unsafe paths, bounds, disk
failure, incomplete/late publication; clock discontinuity, container replacement
and nested/multiple-graph receipts. Qualification arithmetic tests preserve all
thresholds, missing-evidence rejection, signed/equal-weight CPU and stability.

| Validation | Stage / commands or proof | Failure consequence |
| --- | --- | --- |
| V-024 / S-023 | Restore archive bytes; primary index/source preserved; exact dispositions | Block successor; preserve failed evidence |
| V-025 / S-024 | Adopt runtime/fixture boundaries and retire experiments; budgets unchanged | Block successor; preserve failed evidence |
| V-026 / S-025 | Narrow owner checks; validation/selection/security invariants preserved | Block successor; preserve failed evidence |
| V-027 / S-026 | No experimental imports/launches; independent validator unchanged | Block successor; preserve failed evidence |
| V-028 / S-027 | Valid private plans; compiler exits before execution; status never builds | Block successor; preserve failed evidence |
| V-029 / S-028 | Independent scheduling/cache/failure/cancellation/timing parity | Block successor; preserve failed evidence |
| V-030 / S-029 | Graph/UI-review ownership, recovery, cleanup and real service parity | Block successor; preserve failed evidence |
| V-031 / S-030 | Clock/container/receipt bounds, quietness and reader/export parity | Block successor; preserve failed evidence |
| V-032 / S-031 | Section 7 behavioral test/caller cutover; one runtime/fixture owner; fresh public and generated closure | Block successor; preserve failed evidence |
| V-033 / S-032 | Finalization then broader functional checks; qualification still deferred | Block successor; preserve failed evidence |

Narrow commands: `make harness-evidence-contract`,
`make harness-command-surface-contract`; exact owner rows for local migration,
native collection/observation/admission, service-backed schema bootstrap and
browser HTTPS trust. Shared fixture movement also requires
`make harness-ui-review-contract` and `make harness-ui-review-lifecycle`.
Projection changes require `make generate`, `make generate-drift`,
`make generated-artifact-policy-check`, `make json-shape-check`.
Fresh exact-run checks require `make harness-observability-check RESULTS_DIR=...`,
`make test-run-status RESULTS_DIR=... JSON=1` and `make explain-run RESULTS_DIR=...
DETAIL=performance` / `DETAIL=resources`, human and JSON. Test export isolation
without sending external telemetry. Exact commands and result roots are recorded
at execution, not fabricated here.
S-032 runs `make agent-finalize` before broader `harness-contract`, script/shell/
Markdown lint, applicable drift and full warm `check`. `RESULTS_DIR` is supplied
to finalization only for an already eligible full warm successful check; otherwise
retained-run maintenance is recorded skipped. No qualification collection runs.

### Prior iteration — historical scope and evidence


All commands below run from `/home/jochi/code/cartulary` through public Make targets
in a later authorized execution. S-014 owner review is human review; tests,
generators, runtime metadata, conformance and release evidence consume typed
machine inputs, never this tracker, specifications or other Markdown.

At execution entry, run `make task-guide ROLE=module-author OWNER=<owner-id>` for
`harness.evidence_accounting`, `harness.command_surface`, `harness.browser`,
`platform.postgres`, and `module.database_migrations`. Select the narrow routed
rows with `make test-slice OWNER=<owner-id> ROWS=<row-id,...>` or
`make service-backed-test-slice OWNER=<owner-id> ROWS=<row-id,...>`.
E-024 identifies current rows for evidence accounting, local migration activities,
public-registry parity, browser HTTPS trust and testservices lifecycle. S-015 must
save the exact selected/new row IDs and treatment roster before dependent work.

| Validation / layer | Findings / slices | Stage / exact command family | Required result / failure consequence |
| --- | --- | --- | --- |
| V-014 owner/projection closure | F-014–F-018; S-014 and format-changing slices | Human owner review first; later `make generate`, `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check` | No contradictory old clause, unresolved format/statistic/default, hand-edited generated output or mixed producer/reader cutover; failed review/drift blocks dependents |
| V-015 independent characterization | F-014–F-018; S-015 | `make harness-evidence-contract`, `make harness-command-surface-contract` and selected owner/service rows | Routed independent expected values, current baseline failures classified, exact dynamic consumers and future scenarios named; fixture presence alone cannot close RB-005 |
| V-016 retained validation/public parity | F-018; S-016 | Evidence/command contracts; fresh exact-run `make harness-observability-check RESULTS_DIR=<run-dir>` and `make explain-run RESULTS_DIR=<run-dir> DETAIL=performance` / `DETAIL=resources JSON=1` | Single resource stream traversal; uniform malformed-envelope rejection, exact selection, no provisional aggregate leakage, human/JSON parity and export isolation |
| V-017 clocks/accounting | F-014; S-017 | Evidence contract plus Go local-migration and affected browser/producer routes | Hand-calculated brackets/uncertainty, 10 ms quantization, missing/foreign/reversed/discontinuous anchors, suspend, failure/cancellation/fatal exit, incomplete/nested/shared work; canonical path/phase/cache/reservation and local durations unchanged |
| V-018 containers/cleanup | F-015; S-018 | Evidence contract; affected `service-backed-test-slice` and browser HTTPS/lifecycle routes with `HARNESS_DIAGNOSTICS=basic` | Owned service counters, rejected foreign proof, reuse/restart/replacement, borrowed/shared scope, counter reset/disappearance, permission/timeout/cap failures, no double counting, required cleanup and sibling quietness parity |
| V-019 envelope/privacy | F-016; S-019 | Evidence/command contracts plus routed workspace-wrapper, service and browser tests | Wrapper/graph/scan/child boundaries, nesting/multiple runs, preserved admission/signals/exits, scan failure, forged receipt binding, interrupted publication, unsafe paths and writes after shutdown |
| V-020 complete observer accounting | F-017; S-020 | Evidence/command contracts and routed measurement-support fixtures | Hand-calculated process/thread/helper totals, no parent/child or RSS duplication, unavailable measurements, sampled simultaneous versus lifetime semantics, external-measurer parity, forced worker death and quietness |
| V-021 qualification | F-017; S-021 | Frozen public Make workload matrix; `make harness-performance-check EVIDENCE_ROOTS_FILE=<exact-instrumentation-manifest>` | Every treatment cell eligible and stable, complete declared scope and uncertainty, every adopted budget passes; malformed/incomplete evidence cannot qualify; budget failure preserves exit 13 |
| V-022 final current-source regression/handoff | F-014–F-019; S-022 | `make agent-finalize` before broader verification, then `make harness-contract`, `make lint-scripts`, `make lint-markdown`, applicable V-014 checks and final Git review | Current-source evidence closes every selected obligation; repaired source invalidates affected qualification and requires coherent reruns; exact commands/source/roots/failures/skips retained |
| V-023 document/preservation | F-019; T-019 now and S-022 later | Read-only Git/content/identity comparison plus manual Markdown, references, status/dependency and historical-ledger review | All twelve sections, 30-file inventory, stable IDs and one acyclic sequence; only admitted tracker differs; no runtime pass inferred |

### Qualification protocol and adoption outputs

S-014 must adopt explicit typed formulas and eligibility before executable
qualification. S-015 projects the frozen workload/reader commands, cadence,
treatment-order seed, cache/service preparation, measurement resolution and
steady-memory interval into reviewed machine inputs. These are implementation
entry outputs, not permission to improvise the protocol during S-021.

The adopted WSL2 matrix covers Go, Vitest, shell and service/browser work. For each
workload and reader count **0, 1 and 4**, compare off/basic at one clean frozen
source, workload, graph, toolchain, capacity, policy, cache and service-reuse
posture. Readers use existing public commands at the frozen cadence. Preserve
exclusive-measurement quietness; do not bypass fences to manufacture active samples.
Record active/paused coverage and the actual reader invocations.

Retain cold preparation, discarded warm-ups, five measured pairs, randomized
recorded treatment order, all failures/exclusions, per-pair observations, median,
nearest-rank p90 and MAD. Add the matched sixth pair when the adopted stability
rule requires it; unresolved instability fails qualification. Canonical elapsed,
outer elapsed, total incremental CPU, steady additional RSS and diagnostic
finalization are distinct metrics. Never substitute per-process maximum RSS for
simultaneous/steady memory or worker CPU for the complete declared observer scope.

The adopted budgets, projected by S-020/A in plan v3, are median
elapsed increase **≤2%**, p90 **≤3%**, arithmetic mean total incremental observer
CPU **<2.5% of one core per cell**, steady additional RSS upper estimate **≤32 MiB**,
and diagnostic finalization **≤500 ms**. CPU uses equal-weight signed paired
percentages; median, MAD, measurement-error bound and the former conservative
upper estimate remain diagnostics. RSS retains median plus three MADs plus error.
Both canonical and outer elapsed gates remain binding. Missing coverage or uncertainty that prevents a decision cannot pass.
No post-hoc threshold/exclusion changes are permitted to obtain completion.
Windows-host and Docker-daemon-internal costs are outside this declared local
qualification claim, not measured zero. A passing cell cannot stand in for another.

Fault injection also covers spawn/cancellation failures, conflicting process or
container proof, private-channel corruption, exhausted record/byte caps, disk/write
failure, interrupted publication and writes after shutdown. Inject clocks and
operations; do not calculate expected results with the production algorithm.
Preserve existing real-worker and four-sibling quietness tests alongside fixtures.

For S-022, pass `RESULTS_DIR` to finalization only for an eligible successful full
warm `check` run. Otherwise record retained-run maintenance skipped because
`RESULTS_DIR` was unset. No broad product/visual/golden suite or canonical duration
baseline refresh is implied; broaden only for a touched owner or uncovered risk.

| Current document-update command / effects | Disposition | Evidence / applicability |
| --- | --- | --- |
| Git with `--no-optional-locks`, `rg`, exact reads and SHA-256/lstat snapshots | Executed, read-only except the admitted tracker write | E-019–E-025; no runtime verification |
| `make lint-markdown` | Skipped | Inspected wrapper writes static-analysis cache/stamp and retained artifacts; exceeds tracker-only boundary. Current default globs also do not select this tracker; no lint pass claimed |
| Task guides, test suites, generation and `make agent-finalize` | Skipped | Future execution gates; artifact/preparation effects exceed this document update. No new RESULTS_DIR and no retained-run maintenance |
| Qualification experiments and export | Not executed | Qualification is selected future work, not current evidence; export is outside scope |

### Prior validation obligations — historical

V-001–V-013 retain their original meanings and recorded source applicability.
Their commands/results are not current-source evidence for S-014–S-022.

| Validation / layer | Findings / slices | Stage / exact commands | Expected evidence and failure gate |
| --- | --- | --- | --- |
| V-001 document and preservation | F-008; WF-00 | Planning-session: `git --no-optional-locks diff --check`; diff/status/index and content/identity comparison | Only this handoff differs; preserved historical ledger; all references and readiness links reviewed. Any unexpected change is reported, not reverted. |
| V-002 accounting and resource arithmetic | F-003; S-001/S-003 | Pre-move/post-slice: `make task-guide ROLE=module-author OWNER=harness.evidence_accounting`; `make harness-evidence-contract` | Independent synthetic expected values, tampered artifact rejection and large-stream bounds; failures block movement. |
| V-003 engine and integration | F-001/F-002/F-006; S-001/S-004 | Pre-move/post-slice: `make harness-evidence-contract` with new cases routed through existing imports | Injected interleavings plus real worker/thread-death and four-sibling quietness; artifacts in each exact retained run. No current baseline race failure is assumed. |
| V-004 definitions and generated drift | F-004; S-002 | Post-slice: `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make harness-evidence-contract` | Authored/generated agreement, dependency-free validators, equivalent acceptance and extension fixtures; stop on mixed generated state. |
| V-005 public consumer parity | F-005/F-008; S-001/S-005 | Pre-move/post-slice: `make task-guide ROLE=module-author OWNER=harness.command_surface`; `make harness-command-surface-contract`; `make harness-evidence-contract` | Wrapper/check/explanation/export arguments, outputs and exit codes match fixtures; delivery uses injected transport, no external export. |
| V-006 final regression / boundaries | S-005, C-001–C-005 | `make agent-finalize` before `make harness-contract`; `make lint-scripts`; relevant generation checks from V-004 | Current-source owner/import/schema checks pass. Without eligible `RESULTS_DIR`, report retained-run maintenance skipped. No baseline refresh. |

| Validation / layer | Findings / slices | Stage / commands | Required evidence / failure consequence |
| --- | --- | --- | --- |
| V-007 owner/projection closure | F-010–F-012; S-007 and later changed-input slices | Human owner review first; after generated-input changes: `make generate`, `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check` | Closed mappings and declared schema/policy metadata; valid/invalid corpus, synthetic extension and dependency-free validation. Owner or mixed-generation failure blocks dependents |
| V-008 launch identity | F-010; S-008/S-009 | Baseline/post-slice: `make harness-command-surface-contract`, `make harness-evidence-contract` and resolved runner owner slices | Unique nested/packed/repeated attempts; cache hits have none; spawn failure/timeout/cancellation/stale context handled; failure-channel scope/removal and public exits unchanged |
| V-009 ownership/correlation/lifecycle | F-011; S-008/S-010 | Baseline/post-slice: `make harness-evidence-contract` and resolved broker/browser/fixture slices | Shared allocation once, distinct leases, borrowed context, failed acquisition, replacement/PID reuse, missing proof, conflict and cleanup parity; preserve real thread-death/four-sibling quietness tests |
| V-010 local timing/accounting | F-012; S-008/S-011 | Baseline/post-slice: `make harness-evidence-contract` plus resolved Go fixture and row/browser adapter slices | Independent durations/outcomes for migration/reset/parsing; unavailable/incompatible clocks, incomplete ends, overlaps/nesting and shared work; canonical duration/path/cache/reservation results unchanged |
| V-011 retained/public/security | F-011/F-012; S-008/S-012 | Baseline/post-slice: `make harness-evidence-contract` and `make harness-command-surface-contract` | Dangling/forged/duplicate identities, malformed/oversized/symlinked/digest-mismatched artifacts and forbidden values rejected; bounds, human/JSON parity, exact selection, public exits and export isolation pass |
| V-012 final source/handoff | F-010–F-012; S-013 | `make agent-finalize` before broader checks; then `make harness-contract`, `make lint-scripts`, `make lint-markdown` and applicable V-007 drift checks; final `git --no-optional-locks diff --check` | Current-source routed evidence and final diff review; failed required gates block completion. Record exact source, environment, commands, run roots, failures, skips and residual limits |
| V-013 document/preservation | F-013; T-011 and S-013 | Historical E-018 was tracker-only. Execution: staged-byte identity, final Markdown/dependency/status review and staged/unstaged Git whitespace checks | Stable IDs and historical inventories/ledgers remain intact; current status and exact evidence reconcile without reusing historical acceptance |

| Current-session command / inspected side effects | Disposition | Evidence and limit |
| --- | --- | --- |
| Git with `--no-optional-locks`, `rg`, file reads and in-memory SHA-256/lstat comparison | Executed; read-only | E-011–E-018; no harness runtime verification |
| `make lint-markdown` | Skipped | Authored target has Node/frontend readiness; wrapper writes static-analysis cache/stamp and retained run artifacts. Those effects exceed the tracker-only session boundary |
| Task guides, Make suites, generation and finalization | Skipped in this document update | Public wrappers can prepare/write artifacts; future implementation routing/gates are obligations, not current passes. No retained-run maintenance or new RESULTS_DIR |
| Qualification experiments and external export | Not in scope | T-008 remains deferred; no transport invoked |

## 9. Top-Level Work Tracker

### Current lifecycle work status

| ID | Work | State / evidence |
| --- | --- | --- |
| T-042 | Adopt iteration and preserve prior history | DONE; E-041/E-042 |
| T-043 / S-036 | Contracts and characterization | DONE; V-037/E-042 |
| T-044 / S-037 | Native shared runtime/scanner | DONE; V-038/E-043 |
| T-045 / S-038 | Native review lifecycle | DONE; V-039/E-044 |
| T-046 / S-039 | Remaining callers | DONE; V-040/E-045 |
| T-047 / S-040 | Deletion/routing/documentation | DONE; V-041/E-046 |
| T-048 / S-041 | Public boundaries | DONE; V-042/E-047 |
| T-049 / S-042 | Final validation and handoff | DONE; V-043/E-048 |


### Previous functional iteration — current work status on main (historical)

| ID | Work | Status / evidence |
| --- | --- | --- |
| T-039 / S-033 | Prior cleanup | DONE; E-031 |
| T-041 / S-035 | Specification and tracker reconciliation | DONE; C-021/F-033/V-036/E-032 |
| T-040 / S-034 | Scope-aware resource report | DONE; C-019/F-032/V-035/E-033 |
| T-037 / S-031 | Native cutover, separately checkpointed A–F | DONE; V-032/E-039, all checkpoints saved |
| T-038 / S-032 | Functional validation and handoff | DONE; V-033/E-040, final-source full warm check and audit PASS |
| T-033–T-036 / S-027–S-030 | Historical native candidate stages | DONE on salvage source only; fresh validation required |
| T-007 | Original functional coverage | DONE for adopted functional scope; E-040; coverage limits remain explicit |
| T-008, T-026–T-028 | Qualification/original certification handoff | DEFERRED; F-022/RB-006 unresolved |

### Salvage work — historical off-main salvage

| ID | Slice / work | Status | Dependency / gate |
| --- | --- | --- | --- |
| T-029 | S-023: Preservation and inventory | DONE | none; V-024 |
| T-030 | S-024: Specification reconciliation | DONE | S-023; V-025 |
| T-031 | S-025: Retained foundation | DONE | S-024; V-026 |
| T-032 | S-026: Experimental machinery removal | DONE | S-025; V-027 |
| T-033 | S-027: Native composition and preparation | DONE | S-026; V-028 |
| T-034 | S-028: Go graph engine | DONE | S-027; V-029 |
| T-035 | S-029: Shared fixture owner | DONE | S-028; V-030 |
| T-036 | S-030: Observation integration | DONE | S-029; V-031 |
| T-037 | S-031: Public cutover and deletion | IN_PROGRESS | S-030; V-032 |
| T-038 | S-032: Functional validation and handoff | TODO | S-031; V-033 |

Current dispositions supersede only prior readiness: T-007 IN_PROGRESS through
salvage coverage; T-008 and further T-026/T-027/T-028 execution DEFERRED.
S-020, S-021 and S-022 are DEFERRED, not completed. Original failed attempts and
earlier workstream completions retain their recorded source and policy.

### Prior iteration — historical scope and evidence


S-000–S-013 and their original task rows remain DONE at the historical baselines.
T-007/T-008 are authorized and in progress. Their final closure requires S-022;
completed implementation slices do not establish successful qualification.

| ID | Work item / class | Workflow / dependency source | Status | Evidence / exit condition |
| --- | --- | --- | --- | --- |
| T-001 | Foundation planning reconciliation | WF-00; F-008 | DONE | E-001–E-010 apply to the prior planning session |
| T-002 | S-001 characterization | WF-01 | DONE | Foundation checkpoint and S-006 final-source confirmation |
| T-003 | S-002 definitions | WF-02 | DONE | Foundation checkpoint and S-006 final-source confirmation |
| T-004 | S-003 accounting | WF-03 | DONE | Foundation checkpoint and S-006 final-source confirmation |
| T-005 | S-004 lifecycle | WF-04 | DONE | Foundation checkpoint and S-006 final-source confirmation |
| T-006 | S-005 readers/deletion | WF-05 | DONE | Foundation checkpoint and S-006 final-source confirmation |
| T-007 | Remaining original coverage, now selected | F-007/F-014–F-016; WF-17–WF-19 | IN_PROGRESS | Container counters, clock joins and wrapper/receipt coverage selected under T-023–T-025; final closure requires S-022. Earlier DEFERRED dispositions remain historical |
| T-008 | Total observer and 0/1/4-reader qualification; default stays off | F-007/F-017; WF-20/WF-21 | IN_PROGRESS | T-026/T-027 implement and qualify the declared scope; S-022 closes evidence. Default-off decision is settled independently of qualification |
| T-009 | Source-successor comparison waiver | C-005; F-008 | DROPPED | REQ-838 rejects changed source; do not revive |
| T-010 | S-006 foundation validation/handoff | WF-06 | DONE | Final foundation verification ledger; distinct from new iteration |
| T-011 | Selected iteration plan and tracker update / documentation | F-013; V-013 | DONE | E-011–E-018; tracker-only preservation, historical ledgers and dependency review verified |
| T-012 | S-007 owner reconciliation / specification | WF-07 | DONE | S-007 executed checkpoint and final S-013 acceptance; C-006–C-009 / V-007–V-013 as mapped |
| T-013 | S-008 characterization / tests | WF-08 | DONE | S-008 executed checkpoint and final S-013 acceptance; C-006–C-009 / V-007–V-013 as mapped |
| T-014 | S-009 invocation context / implementation | WF-09 | DONE | S-009 executed checkpoint and final S-013 acceptance; C-006–C-009 / V-007–V-013 as mapped |
| T-015 | S-010 allocation/process relationships / implementation | WF-10 | DONE | S-010 executed checkpoint and final S-013 acceptance; C-006–C-009 / V-007–V-013 as mapped |
| T-016 | S-011 owner-local timings / implementation | WF-11 | DONE | S-011 executed checkpoint and final S-013 acceptance; C-006–C-009 / V-007–V-013 as mapped |
| T-017 | S-012 retained analysis / implementation | WF-12 | DONE | S-012 executed checkpoint and final S-013 acceptance; C-006–C-009 / V-007–V-013 as mapped |
| T-018 | S-013 validation/handoff / verification | WF-13 | DONE | S-013 executed checkpoint and final S-013 acceptance; C-006–C-009 / V-007–V-013 as mapped |
| T-019 | Next iteration plan and tracker update / documentation | F-019; V-023 | DONE | E-019–E-025; current scope, inventory, mappings, history and tracker-only preservation verified; no implementation performed |
| T-020 | S-014 owner reconciliation / specification | WF-14 | DONE | V-014: REQ-845–849 / AC-142–146, owner-clause review, Markdown-01 and diff checks pass; execution checkpoint below |
| T-021 | S-015 characterization and routing / tests | WF-15 | DONE | V-015 closes RB-005 with executed independent evidence |
| T-022 | S-016 unified retained validation / correction and structure | WF-16 | DONE | V-016 closes F-018 with uniform rejection and one traversal |
| T-023 | S-017 clock anchoring / extension | WF-17 | DONE | V-017 closes F-014; no fabricated placement or canonical changes |
| T-024 | S-018 container observations / extension | WF-18 | DONE | V-018 closes F-015; real coverage and unchanged cleanup authority |
| T-025 | S-019 wrapper/terminal receipt / extension | WF-19 | DONE | V-019 closes F-016; adopted privacy, ordering and exit parity |
| T-026 | S-020 complete observer measurement / extension | WF-20 | IN_PROGRESS | Candidate-16 failure saved; Section 7 recovery gates A–E precede the next completion checkpoint and cohort |
| T-027 | S-021 qualification / engineering evidence | WF-21 | IN_PROGRESS | V-021 closes RB-006; every required matrix cell passes |
| T-028 | S-022 validation and handoff / final verification | WF-22 | TODO | V-022/V-023 reconcile all selected obligations and final source; then T-007/T-008 and execution items may be DONE |

## 10. Session Handoff Log

### E-048 — S-042/V-043 final validation and handoff — DONE

Finalization first: `env -u RESULTS_DIR CARTULARY_TEST_RUN_ID=lifecycle-s042-finalize01
make agent-finalize` completed with owning exit 0, PASS, 21,794 ms. Its
`unit-artifacts/finalize-summary.json` reports zero generated changes. Canonical
retained-run and scheduler-drift maintenance were explicitly skipped because
RESULTS_DIR was unset (`results-dir-not-provided`); there was no eligible current
full-warm evidence to supply. No skipped maintenance is reported as a pass.

The first pre-repair source freeze was `2b800928c` (only the tracker differed from
implementation commit `081f5b0e6`), digest
`sha256:4550a159bda71123d1c20e7fe23fb94fc770984eceacea92ac638e14b6db42c3`.
Final pre-check gates used fresh `lifecycle-s042-<suffix>` IDs under
`.cartulary/test-results/`, with observed owning exit 0 and reported PASS:

| Suffix | Make target | Evidence |
| --- | --- | --- |
| `contract01` | `harness-contract` | 2/2 graph units; 182/182 extended assertions; 100,025 ms |
| `scripts01` | `lint-scripts` | 694 ms |
| `shell01` | `lint-shell` | 16,982 ms |
| `markdown01` | `lint-markdown` | 8,180 ms; final handoff prose will receive its own narrow check |
| `drift01` | `generate-drift` | 18,632 ms |
| `policy01` | `generated-artifact-policy-check` | 1,066 ms |
| `json01` | `json-shape-check` | 3,696 ms |

`CARTULARY_TEST_RUN_ID=lifecycle-s042-check01 make check` FAILED, owning exit 2.
Frontend typecheck reported missing declaration for the shared incident-bundle
fixture and an unused Buffer import left by S-038; Biome independently reported
the unused import (not a formatter failure). The graph continued independent
work, so the agent sent SIGINT through a pidfd after verifying the exact native
graph owner's run identity. Cancellation exposed F-042: cache freshness mapped
context cancellation to configuration failure; graph admission treated it as a
required publication failure and discarded the canonical event stream. The public
outer diagnostic was config/configuration_error. No run-summary or cleanup summary
was published; the last live observation recorded 231 passed, 4 failed, 747
cancelled of 982 units. These counts are a live observation, not completed evidence.
The retained scan passed over 1,626 files. The native private root and Docker
containers were absent afterward, but missing canonical cleanup evidence is not
reported as a passed cleanup gate. This failed run remains intact.

Repair scope: restore the typed shared fixture interface with an authored `.d.mts`
and remove the unused import, without copying its implementation back into the
browser helper. Independently reproduce cancelled freshness and interrupted cache
admission/publication through the actual native owners, preserve context errors,
and treat interrupted cache lookup as normal cancellation. No failure threshold,
cache acceptance rule, deadline or schema is relaxed. New source identity and
affected runtime/typecheck/lint/finalization/full-check gates are required.
`runtime-repro01 make harness-runtime-contract` reproduced freshness cancellation
as config/configuration_error (owning exit 2); its initial publication fixture
also failed private directory permissions before exercising the owner. The fixture
now explicitly sets 0700 as required. `runtime-repair01` then exposed the second
part: cancellation inside synchronous cache admission reached the scheduler's
no-progress check before pending units were settled. The owner loop now settles
cancellation after the admission loop, before checking progress; cancelled cache
reads/restores preserve existing entries and outputs rather than quarantining a
valid record. Regression expectations cover all three cache phases and complete
canonical cancellation publication. No gate failure is hidden.

`format01`/`format02 make format` passed. `types01 make frontend-typecheck` passed
(21,478 ms), and `biome01 make lint-biome` passed (2,044 ms); these close the typed
fixture interface and unused import findings. `runtime-repair02 make harness-runtime-contract` passed with race checks for
runtime/schema/session/review; all new cancellation regressions passed.
`cancel01 make test-slice OWNER=harness.command_surface
ROWS=harness.command_surface.unit.native_runtime CARTULARY_HARNESS_CACHE_MODE=off`
then exercised the installed public graph: after observing its one running unit,
SIGINT was sent to its exact verified owner. Native exit 130 / outer Make exit 2
and canonical `cancelled` were expected. Run/events/cleanup/scan evidence exists;
read-only observability audit and summary explanation both passed. This is an
intentional negative boundary test, distinct from check01's publication failure.
No source change was made while this invocation ran.

Repair commit `ab2afb5e0` is the new executable freeze, digest
`sha256:3e5473d6d94660c085252c27a593bbc8bf5083b34b1588d5bfec0a1d077609a2`.
`env -u RESULTS_DIR CARTULARY_TEST_RUN_ID=lifecycle-s042-finalize02 make
agent-finalize` passed (22,028 ms, owning exit 0) and refreshed zero files.
Retained-run maintenance remains explicitly skipped for unset RESULTS_DIR.
Repaired-source final gates all passed with observed owning exit 0:
`contract02 make harness-contract` (182/182 assertions, 99,660 ms),
`scripts02 make lint-scripts` (580 ms), `shell02 make lint-shell` (1,608 ms),
`markdown02 make lint-markdown` (7,370 ms), `drift02 make generate-drift`
(19,087 ms), `policy02 make generated-artifact-policy-check` (1,076 ms), and
`json02 make json-shape-check` (3,707 ms). Each uses the full
`CARTULARY_TEST_RUN_ID=lifecycle-s042-<suffix>` prefix. Earlier source-scoped
passes are not substituted for this validation.
`CARTULARY_TEST_RUN_ID=lifecycle-s042-check02 make check` completed with owning
Make exit 2, native child exit 1 and canonical FAIL: 981/982 passed, one failed,
zero skipped/cancelled; 704,462 ms. Only `target:lint-go` failed. Staticcheck found
unused `harnessfixtures.within` left by transport deletion (U1000) and redundant
field copying from `adapterOutcome` to `Operation` (S1016). All 1,241 cleanup
records completed; the retained scan passed over 6,185 files. Cache disposition
was 0 hits, 819 misses and 163 bypasses; all misses remain current-run evidence.
The completed failed root is preserved unchanged.

Remove the unused helper/imports and use the equivalent checked Go conversion
`Operation(o)`. This closes residual support-code cleanup without changing any
public or retained behavior. `lintgo01 make lint-go`, `fixture01 make
harness-fixture-contract`, and `runtime-repair03 make harness-runtime-contract`
all passed with observed owning exit 0. Runtime/schema/session/review and fixture
race tests are complete; the selected graphs all report source digest
`sha256:6dfb070d1f521a638b2121d285f6f5dd3e39ef20ebe700062933b5dbb010e7e8`.
Final executable freeze is `aed9737cbb77683451bdd4dfd47d97b42c3bb1ed`, with the
same digest above. Only this tracker changes after that freeze.
`env -u RESULTS_DIR CARTULARY_TEST_RUN_ID=lifecycle-s042-finalize03 make
agent-finalize` passed (21,534 ms, owning exit 0), changing no generated files.
The unset-RESULTS_DIR maintenance skip remains explicit. Final-source gates passed
with observed owning exit 0 after this finalization. All rows use
`CARTULARY_TEST_RUN_ID=lifecycle-s042-<suffix> make <target>` and retained roots
`.cartulary/test-results/lifecycle-s042-<suffix>`:

| Suffix | Make target | Final-source result |
| --- | --- | --- |
| `contract03` | `harness-contract` | PASS; 2/2 graph units, 182/182 assertions; 97,462 ms |
| `scripts03` | `lint-scripts` | PASS; 562 ms |
| `shell03` | `lint-shell` | PASS; 1,496 ms |
| `markdown03` | `lint-markdown` | PASS; 7,810 ms |
| `lintgo02` | `lint-go` | PASS |
| `drift03` | `generate-drift` | PASS; 18,762 ms |
| `policy03` | `generated-artifact-policy-check` | PASS; 1,050 ms |
| `json03` | `json-shape-check` | PASS; 3,662 ms |

`CARTULARY_TEST_RUN_ID=lifecycle-s042-check03 make check` PASSED on that frozen
source: observed owning Make exit 0, canonical PASS, 982/982 units passed, zero
failed/skipped/cancelled, 837,329 ms canonical wall. Retained root is
`.cartulary/test-results/lifecycle-s042-check03`. Cache accounting is explicit:
211 hits, 608 misses, 163 bypasses; cache hits are accepted reuse, not newly
executed producer evidence. All 1,030 cleanup records completed with no failure or
blocked cleanup. `retained-secret-scan.json` reports PASS over 6,063 files with
unchanged policy digest
`sha256:eaa74efae69ea73a6e0f18823a94ccc70a19b242e2cb2880e446611c3084ee11`.
No check03 private runtime root remained; Docker's running-container inventory
was empty. The earlier failed roots and exact recovery proof remain preserved.

Environment: Linux/WSL, effective Go 1.27.2 selected by launcher 1.27.1, repository
pinned Node/tooling and Docker. Final manifest records clean executable source,
toolchain digest
`sha256:4b62e2558e6dcbe1d3c632fb9c0bb47e5a89e5dda08f103a72f61347b5ef1810`
and system digest
`sha256:546793d134ea4063515a25068aa22b78078fc1613c1ee56fb5f9e41c5c550588`.
Contract03 and check03 agree on executable commit and digest. Markdown-only
checkpoint edits do not participate in those executable identities.

Exact-run read-only commands all exited 0:
`make test-run-status RESULTS_DIR=.cartulary/test-results/lifecycle-s042-check03 JSON=1`,
`make explain-run RESULTS_DIR=.cartulary/test-results/lifecycle-s042-check03 TARGET=check DETAIL=summary`,
and `make harness-observability-check RESULTS_DIR=.cartulary/test-results/lifecycle-s042-check03`.
The independent audit passed one invocation/four sources. Separate
`make explain-run RESULTS_DIR=.cartulary/test-results/lifecycle-s042-check03 DETAIL=performance`
and `DETAIL=resources`, each with and without `JSON=1`, passed. Views are retained
locally at `/tmp/cartulary-lifecycle-s042-views/`; both JSON views are the same
immutable explanation model. Human canonical timing agrees (837,329 ms);
resources/execution are explicitly `not_observed` and the terminal receipt is
unavailable because diagnostics are off. The JSON model preserves the bounded
20-unit view's 962 omitted units. No sample, consumption total or qualification
result is inferred from this default-off run. Reader limitations remain explicit:
they do not attest current source or observe the outer exit; those were separately
captured above. E-047's diagnostic service run and E-046/E-048's contract fixtures
supply the positive observation/receipt and local export-isolation coverage.
No export command or external telemetry transmission ran.

Final handoff: F-037–F-042 are CLOSED for this scope. Shared runtime, scanner,
recovery and cleanup policy live in `harnesssession`; `harnessreview` owns review
session transitions/publication and calls `harnessfixtures` directly, as graph
does. Superseded Node policy, fixture-session transport/schema/subcommand, copied
test coordinators and obsolete routing are removed. Retained Node compiler,
selection, row/browser/physical adapters, preparation recipes and offline readers
have distinct owners and obligations; no compatibility engine or daemon remains.
Public commands, required failures, retained acceptance sets, scan policy identity
and diagnostics defaults are preserved. The private session protocol change
requires other environments to drain old sessions with their running version;
unresolved ownership blocks that environment's cutover. Rollback restores the
coherent source/projection transaction, without an old-reader shim or receipt
rewrite. Graph final scan/outer receipt and UI terminal publication/scan remain
separate contracts.

Every predecessor checkpoint was saved before its successor. S-042/V-043 is DONE;
functional remediation is complete. The final checkpoint changes this tracker
only. `CARTULARY_TEST_RUN_ID=lifecycle-s042-markdown04 make lint-markdown` passed
with observed owning exit 0 (7,930 ms); its retained root is
`.cartulary/test-results/lifecycle-s042-markdown04`. Final `git diff --check`
passed, twelve numbered sections remain, and executable files stayed unchanged
through evidence inspection and handoff. All failed attempts and maintenance
skips above remain part of the handoff.
F-022/RB-006 and S-020–S-022 remain unresolved/deferred, with unchanged thresholds
and 44.952/90-minute investigation ledger. No overhead, RSS/CPU improvement,
default-on readiness, or Core 05 publication is claimed.

### E-047 — S-041/V-042 installed public boundaries — DONE

Saved before S-042. Executable source is clean
`081f5b0e6fb08ee822b4d99e6bcf030b3631cdd8`, source digest
`sha256:4550a159bda71123d1c20e7fe23fb94fc770984eceacea92ac638e14b6db42c3`.
Only this tracker changed during the slice; no executable repair was necessary.
Environment remains Linux/WSL with pinned installed tools and Docker. Discovery
through `make help`, `make help-all`, module-author task guides for
`harness.browser` and `module.database_migrations`, and summary explanations for
seeded/native-boundary routes all exited 0. Two guessed test filenames did not
exist; actual import paths were inspected instead, without affecting any gate.

Every writing command uses `CARTULARY_TEST_RUN_ID=lifecycle-s041-<suffix>`;
root is `.cartulary/test-results/lifecycle-s041-<suffix>`. All commands below
completed with observed owning exit 0 and reported pass, independently collected.

| Suffix | Exact Make target and arguments | Outcome |
| --- | --- | --- |
| `command01` | `harness-command-surface-contract` | 42/42; actual Go/Node/shell/Vitest child failures, cancellation/capture, frontend staging, status without provisioning |
| `ui01` | `harness-ui-review-contract` | 10/10; public input and preparation rejection |
| `lifecycle01` | `harness-ui-review-lifecycle` | 31/31; artifact/dev, repeated stop, controller/parent loss, preparation drain and recovery |
| `evidence01` | `harness-evidence-contract` | 90/90; hostile retained facts, scans/receipts, scope formatting, explicit export isolation and independent readers |
| `seeded01` | `harness-ui-review-seeded-default` | Real seeded editor/viewer/capture/artifact workflow, interruption and complete private cleanup |
| `service01` | `service-backed-test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.schema_bootstrap HARNESS_DIAGNOSTICS=basic CARTULARY_HARNESS_CACHE_MODE=off` | 3/3; 18,677 ms canonical wall |
| `object01` | `service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.integration.object_store_fixture_roundtrip CARTULARY_HARNESS_CACHE_MODE=off` | 3/3; 28,241 ms |
| `local01` | `test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.local_migration_activities HARNESS_DIAGNOSTICS=basic CARTULARY_HARNESS_CACHE_MODE=off` | 1/1; 2,878 ms |
| `https01` | `service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.https_trust HARNESS_DIAGNOSTICS=basic CARTULARY_HARNESS_CACHE_MODE=off` | 11/11; 26,507 ms |
| `design01` | `browser-design-review-smoke REVIEW_PROFILE=default` | Real design preparation/scenarios followed by closed review and 7/7 successful cleanup records |

Read-only `make test-run-status RESULTS_DIR=<exact-root> JSON=1` for service01,
seeded01 and https01 exited 0 and reported terminal pass. Observer limitations
(current source not audited and outer exit not observed by the reader) remain
explicit; the owning exits above were collected separately. For service01,
`make harness-observability-check RESULTS_DIR=<root>` passed: one invocation,
four sources, complete execution/diagnostic evidence. Separate
`make explain-run RESULTS_DIR=<root> DETAIL=performance` and `DETAIL=resources`,
each with and without `JSON=1`, exited 0. Both JSON views are the same immutable
model; human timing/sample/scope/unavailable/omission values agree. Temporary
views are `/tmp/cartulary-lifecycle-s041-views/{performance,resources}.{json,txt}`.
The short service run legitimately has no sampled container identity; no workload
extension or invented totals were used. Broader physical sampling remains ordinary
diagnostic coverage, never a qualification claim.

The design private root was absent after completion; `review-session.json` is
closed and `cleanup-results.json` has no failures. Docker process inventory was
empty after all calls. Prior S-040 boundary02 additionally covers actual borrowed
service preservation, owned cleanup and intentionally failed exact recovery on
the same production source. Public/retained acceptance sets are unchanged;
expected negative outcomes in contract fixtures are assertions, not failed gates.
No export command ran and local readers did not load export code (evidence suite).
S-041/V-042 is DONE; S-042 is READY. Full warm check and finalization remain
required; no current-source final handoff or qualification is claimed yet.

### E-046 — S-040/V-041 deletion, routing and documentation — DONE

Parent `f15f6c7173605120fcccd84b16c06caf280e274c`; Linux/WSL, installed
pinned Go/Node and Docker physical owners. This checkpoint supersedes current
legacy-deletion summaries, not prior completion evidence.

Deleted Node suite-runtime/scanner/recovery policy, fixture-session client and
Go transport/schema/subcommand, review controller/session/preparation/recovery/
terminal wrappers and their copied test coordinators. Specialized Node browser,
service and preparation operations remain. Native session owns physical-stop
private dispatch; review and fixture owners keep their own cleanup ordering.
Native output now validates a closed generated envelope; secondary failures have
a generated 64-item bound, preserve the primary, and reject malformed/truncated/
foreign/duplicate adapter replies. Authored task/schema/helper/test inputs were
updated, then generated normally. Active guide ownership was corrected and five
historical inventories explicitly source-scoped; historical failures stay intact.

Every command below has the form `CARTULARY_TEST_RUN_ID=<ID> make <target>`;
run root is `.cartulary/test-results/<ID>`. Successful entries mean both observed
owning exit 0 and reported pass; failures retain owning exit 2 and failed gate.

| ID suffix after `lifecycle-s040-` | Target | Observed result |
| --- | --- | --- |
| `format-01` | `format` | FAIL: unused Go import in driver; removed import |
| `generate-01`, `generate-02`, `generate-03`, `generate-04` | `generate` | PASS; intermediate source |
| `format-02`, `format-03`, `format-04` | `format` | PASS; intermediate source |
| `runtime-01`, `fixture-01`, `ui-01`, `lifecycle-01` | respective `harness-runtime-contract`, `harness-fixture-contract`, `harness-ui-review-contract`, `harness-ui-review-lifecycle` | PASS; intermediate source |
| `generate-05` | `generate` | FAIL: native schema projection rejected JavaScript Unicode regex escape; authored pattern corrected |
| `generate-06`, `generate-07`, `generate-08` | `generate` | PASS after repair |
| `runtime-02` | `harness-runtime-contract` | FAIL: selected graph wrapper receipt rejected two distinct source digests; schema/session/review packages passed |
| `boundary-01` | `harness-runtime-boundary-contract` | FAIL: borrowed physical fixture incorrectly selected owned-only `ui-review`; owned/recovery variants passed |
| `contracts-01` | `harness-contract-tests` | PASS 190/190 |
| `lifecycle-02` | `harness-ui-review-lifecycle` | PASS 31/31 |
| `ui-02`, `scripts-01`, `smoke-01` | `harness-ui-review-contract`, `lint-scripts`, `harness-smoke` | PASS |
| `drift-01`, `policy-01`, `json-01` | `generate-drift`, `generated-artifact-policy-check`, `json-shape-check` | PASS |
| `format05`, `format06` | `format` | PASS after boundary fixture/assertion repairs |
| `boundary02` | `harness-runtime-boundary-contract` | PASS actual borrowed/owned/failed-cleanup recovery, native PostgreSQL/HTTPS graphs and session-client boundary |
| `generate09` | `generate` | PASS final authored projections |

`runtime-02` recorded source digests `326a8905d5c9e75dd17bd5de95ef08b7a4fe9bf05c08a6c07874808875051ca9`
and `db252a8abbd8e2708213a4a1e0d76d33ab60941ede5b5a4f76bb5ff4c2794f7e`
under one wrapper while other verification/build work ran. The mixed evidence is
ineligible; no receipt gate was relaxed. The affected rerun is serialized after
source-changing work. The exact source-writing cause was not established.

`boundary-01` also exposed a test-fixture proof lifetime defect: Go's `t.TempDir`
removed the failed persistent owner's descriptor. The revised fixture stores it
under the preserved native private runtime and uses the supported
`service-backed-test-slice` attachment target. Two orphaned containers from that
failed test were manually removed by exact full container ID after checking
session `24b2b0998bc503d3cacd3c62`, creation at 03:44:43Z on 2026-10-12 and
expiry `2026-10-13T03:44:43.067Z`: `5f509b2d1c1830279e1683ab917020e7ee638ad2cfbf9f2774b4c07ccafd2282`
and `4d849f6578de8b3d87011b35d3146ab0c2df73f5b88469337bfc4d6611a3cd3d`.
Operational teardown used `docker inspect` and `docker rm -f` for those exact
identities only (exit 0); it is not native recovery validation. No receipt was
rewritten and no unknown ownership was deleted. The failed runtime's unresolved
journal remains private historical proof. `boundary02` validates the repaired
fixture's complete supported cleanup path.

Final closure: `runtime03` (`harness-runtime-contract`, race enabled), `drift02`
(`generate-drift`), `policy02` (`generated-artifact-policy-check`), `json02`
(`json-shape-check`) and `markdown01` (`lint-markdown`) all completed with owning
exit 0 and reported pass. `git diff --check` passed. Actual runtime/session/review
packages include failed journal registration, duplicate journal rejection and
existing-terminal preservation; the former Node assertions below are closed.
Source digest in all three selected graphs is
`sha256:4550a159bda71123d1c20e7fe23fb94fc770984eceacea92ac638e14b6db42c3`;
staged executable/projection diff SHA-256 against the parent is
`de38e603435f475130f2b18bd99d1e10448ecb2b482b40c0794af0a38b2788ac`.

Import/build/schema/backing-script searches found no deleted implementation
consumer. Only distinct compiler/producer/browser/physical/reader adapters remain;
fast smoke still has its three roles and broad suites stay in owner routes.
Current guide plus historical annotations passed Markdown lint. Public and retained
schema acceptance sets are unchanged; only superseded private transport is removed.
No fallback, new service, secret-bearing argv or replacement scanner oracle exists.
S-040/V-041 is DONE and saved before S-041; S-041 is READY. Remaining risks are
public composition and final-source validation, addressed by S-041/S-042.
F-022/RB-006 and the 44.952/90 ledger remain unchanged.

### S-040 assertion migration ledger — CLOSED at E-046

Assertion dispositions below identify actual owners, independently specified expectations, and obsolete mechanisms. V-041 passed the final affected gates recorded above.

| Former assertion | Native obligation |
| --- | --- |
| `test-ui-review-lifecycle.mjs`: hung cleanup is bounded and stale cleanup preserves unresolved proof | MOVE → `TestReviewLifecycle/undrained_preparation_blocks_cleanup_and_terminal`, `TestSessionRuntimeProof/preservation_and_janitor`; runtime contract |
| `test-ui-review-lifecycle.mjs`: stop drains a resource returned after preparation cancellation | MOVE → `TestReviewLifecycle/late_acquisition_is_closed_before_terminal`; runtime contract |
| `test-ui-review-lifecycle.mjs`: unavailable borrowed origin expires without acquiring or resetting services | MOVE → `TestReviewLifecycle/borrowed_origin_deadline_has_no_destructive_effect`; runtime contract |
| `test-ui-review-lifecycle.mjs`: partial cleanup attempts all owners, preserves the primary failure and retains failed ownership | MOVE → `TestReviewLifecycle/primary_precedes_independent_cleanup_failure`, `TestFixtureRecovery`, `TestSessionRuntimeProof/preservation_and_janitor`; runtime/fixture contracts |
| `test-ui-review-lifecycle.mjs`: every normalized primary failure survives secondary cleanup failure and terminal repetition | MOVE → `TestReviewLifecycle/every_primary_survives_cleanup_and_repetition` (independent 23-code/exit matrix); runtime contract |
| `test-ui-review-lifecycle.mjs`: session serialization bounds contention and never overlaps operations | MOVE → `TestReviewLifecycle/single_admission_and_stop_drains_before_terminal`, native contention/deadline cases; runtime contract |
| `test-ui-review-lifecycle.mjs`: eight-hour expiry closes the session at the exact lifetime boundary | MOVE → `TestReviewLifecycle/exact_eight_hour_lifetime` (virtual clock, boundary minus one ns and exact expiry); runtime contract |
| `test-ui-review-lifecycle.mjs`: unsafe terminal publication cannot report success or overwrite an existing file | MOVE → `TestReviewPublication` scan rejection, exclusive publication and immutable receipt cases; runtime contract |
| `test-ui-review-lifecycle.mjs`: a preparation failure already in flight survives an exact stop during cleanup | MOVE → `TestReviewLifecycle/late_preparation_failure_survives_stop`; runtime contract |
| `test-ui-review-preparation.mjs`: shared preparation cleanup alone has cleanup classification and preserves an earlier operational failure | MOVE → `TestReviewLifecycle/cleanup_only_failure_and_failed_publication` and primary matrix; runtime contract |
| `test-ui-review-preparation.mjs`: preparation IPC retains validated context and latches malformed outcomes | MOVE semantic rejection/context obligation → `TestReviewIPC/independent_response_rejections`, actual private adapter; REMOVE obsolete process.send protocol assumptions |
| `test-ui-review-preparation.mjs`: cancellation in each preparation phase rejects late ready publication and drains the IPC owner | MOVE → native late-acquisition, undrained-preparation and actual IPC-owner drain; runtime contract; KEEP actual public stop/death cases in UI lifecycle |
| `test-ui-review-preparation.mjs`: IPC rejects duplicate terminal outcomes and success followed by child failure | MOVE → `TestReviewIPC` duplicate/truncated/exit mismatch rejection and exact process drain; REMOVE old preparation IPC message format |
| `test-ui-review-preparation.mjs`: preparation cleanup-only IPC outcome retains cleanup context after cancellation | MOVE → native cleanup-only/context and primary matrix cases; REMOVE superseded IPC envelope assertions |
| `test-ui-review-preparation.mjs`: a failing Make producer crosses preparation IPC into the terminal receipt and safe human guidance | MOVE → native validated failure mapping/IPC and actual Node Make producer envelope tests retained in preparation route; KEEP command-surface public failures |
| `test-ui-review-preparation.mjs`: preparation bounds secondary diagnostics without replacing the primary outcome | MOVE → `TestReviewLifecycle/bounded_secondary_diagnostics_preserve_primary`, IPC 65-failure rejection; runtime contract |
| `test-ui-review-preparation.mjs`: missing preparation IPC stays primary when reaping acknowledgement and later stop fail | MOVE → native undrained-preparation and primary/cleanup precedence; REMOVE superseded missing-process.send classification path |
| `test-browser-acquisition.mjs`: borrowed runtime controller records do not hide or block preparation dependencies | REMOVE controller-record filter: retired controller/fixture transport no longer shares that policy. KEEP actual runtime run/lease binding and native dependency-order fixtures |
| `test-browser-acquisition.mjs`: failed registration acknowledgement retains dependency and exact acquisition proof | MOVE → native durable journal fault/security cases and physical acquisition proof tests; REMOVE obsolete fixture-session acknowledgement mechanism |
| `test-browser-acquisition.mjs`: missing and duplicate durable browser dependencies fail closed without releasing a suite | MOVE → native exact resource filename identity and `TestFixtureRecovery` missing/unknown dependency rejection; no duplicate executable recovery oracle |

Additional groups: removed `TestFixtureSession` authenticated socket/request tests
are obsolete with the deleted transport; native session/review IPC now cover
run/lease/owner binding and hostile input. Broker race, close-during-acquisition,
release/reuse and recovery tests remain in the fixture route. Removed
`ui-review-fixture-cleanup.mjs` cases move to
`TestNativeFixturePhysicalLifecycle/{borrowed,owned,failed_cleanup}` through the
native boundary route; these use actual service/browser owners. Removed
design-review coordinator loops move to the native lifecycle/failure tests and
actual public design boundary; profile/TOTP/archive recipes remain Node tests.
Node suite-runtime/scanner tests move to `TestSessionRuntimeProof`,
`TestSessionProtocol`, and independent `TestSessionScan` fixtures. Node producer
umask, frontend staging, quiet capture, command failures and local-service tests
remain and use the native session client. Compiler/selection, schema parity,
offline explanation, retained evidence and qualification validator tests remain.
No removed test is retained as a second lifecycle or scanner implementation.

### E-045 — S-039/V-040 remaining caller migration — DONE

Saved before S-040. Parent `2ae231627`; staged source/projection diff SHA-256
`3beff8df299f680b92e73fa227b7923cd2094777028768eb7d4263040a909d4c`
(excludes tracker). Environment: same local Linux workspace and pinned installed
Go/Node/browser tools, actual local PostgreSQL/object-store owners. Source identity
is this checkpoint's commit. No qualification acquisition occurred.

Migrated smoke, wrapped tools, Vitest diagnostics, execution observations,
command-failure capture, fixture attachments, browser fixtures, frontend producers
and local-service sessions to the native session client. Captured children register
exact native process proof, drain descendants before release, and preserve failed
proof. Frontend builds share that physical capture adapter. Wrapped tools borrow
inherited authority; independent callers create an owned native root. Persistent
service commands journal pending/acquired/released suite proof and recover the
original runtime only after exact physical cleanup. No supervisor was introduced.
Build prerequisites include the installed native binary for migrated writers;
repeated generated Make predicates were factored without raising density budgets.

All writing commands below use `CARTULARY_TEST_RUN_ID=lifecycle-s039-<suffix>`;
roots are `.cartulary/test-results/lifecycle-s039-<suffix>`. Step logs are
`<target>/<target>/{stdout,stderr}.log`; graph summaries remain at their exact roots.
Exits are the owning Make process, separately from harness results.

| Suffix | Make target / arguments | Exit / harness result |
| --- | --- | --- |
| generate-01 | generate | 2 / fail: generated Make density budget |
| generate-02–07 | generate (six distinct invocations) | 0 / pass each; final 9,840 ms |
| boundary-01 | harness-boundary-contract | 2 / no such target, no execution |
| contracts-01 | harness-contract-tests | 2 / failed borrowed-wrapper expectation and private browser import boundary |
| contracts-02 | harness-contract-tests | 2 / premature test-owned close and stale projection digest |
| contracts-03 | harness-contract-tests | 2 / frontend producer used unsupported launch label |
| contracts-04 | harness-contract-tests | 0 / 190 passed, 0 failed/skipped, 101,232 ms |
| runtime-boundary-01 | harness-runtime-boundary-contract | 0 / pass, 40.478 s |
| runtime-boundary-02 | harness-runtime-boundary-contract | 2 / failed frontend label and non-private test directory |
| runtime-boundary-03 | harness-runtime-boundary-contract | 0 / pass, 41.559 s; actual PostgreSQL/HTTPS and native session IPC |
| runtime-01 | harness-runtime-contract | 0 / all native race suites pass (runtime/schema/session/review) |
| smoke-01 | run-harness-smoke-fast | 0 / pass, three-role surface preserved |
| services-up-01 | test-services-session-up CARTULARY_TEST_SERVICES_SESSION_FILE=/tmp/cartulary-lifecycle-s039/session.json | 0 / pass, 6,670 ms |
| services-status-01 | test-services-session-status with same explicit session file | 0 / active compatible services, zero borrowers |
| services-down-01 | test-services-session-down with same explicit session file | 0 / pass, 1,120 ms; physical services and exact runtime released |
| format-01–02 | format (two distinct invocations) | 0 / pass each |
| build-web-01 | build-web | 2 / harness exit 11, unsupported producer label; repaired and covered by runtime-boundary-03 and contracts-04 |
| drift-01 | generate-drift | 0 / pass, 18,182 ms |
| policy-01 | generated-artifact-policy-check | 0 / pass, 1,012 ms |
| json-01 | json-shape-check | 0 / pass, 3,636 ms |

Read-only discovery: `make task-guide ROLE=module-author OWNER=harness.runtime`
failed (unknown owner); corrected `OWNER=harness.command_surface` and
`make explain-target TARGET=harness-runtime-boundary-contract DETAIL=summary`
passed. Default `make test-services-session-status` reported absent before the
explicit isolated persistent session. `git diff --cached --check` passed.

Repairs retain their failed evidence: existing launch label `private_capture`
replaces the invalid label, and the boundary fixture explicitly sets 0700 as the
scanner requires. Installed-binary tests prove forged lease rejection, borrowed
close protection, durable shared secret registration, successful scanning, repeated
close, owner-loss recovery rejection while proof is outstanding, and successful
recovery only after release. Existing caller tests retain failure/interruption,
staging, quiet capture and inherited-runtime assertions. Broad finalization/lint
remain S-042; this is not a final completion claim.

Compatibility: public commands and retained schemas unchanged; private runtime
ownership is native. Legacy Node policy/tests still await S-040 deletion; no fallback
is used by migrated production callers. S-039 is DONE; S-040 is READY. F-022/RB-006,
S-020–S-022, qualification thresholds and the 44.952/90-minute ledger are unchanged.

### E-044 — S-038/V-039 native review lifecycle — DONE

Saved exit checkpoint before S-039. Parent source is
`f64bc88b771fcb523725bb207d92fb3d48b2fe40`; staged executable/source projection
diff SHA-256 is `adce43945a246a30df557c5043d13392190ceca1c6a621e9d6f69fac064fc6c8`
(excludes tracker). Final source is this checkpoint's commit; final broad-source
validation remains S-042.
The pre-cutover `/tmp/cartulary-ui-review-<uid>` registry contained no JSON session
records; no active old session required draining. Environment remains the local
Linux workspace, installed pinned Node/Go/browser tools and local service owners.

Implemented: `harnessreview` owns preparation, session admission,
serialization, cancellation, private registry/IPC, recovery and public receipts;
`harnessadmission.Lease` supplies shared admission waiting. UI commands now compile
inputs through `native-client.mjs`; browser, artifact and preparation recipes stay
in bounded Node adapters. Design review uses the same native coordinator and
fixture library with helper-only presentation. Seeded UI build/fixture evidence
stays in the private runtime. Shared bounds and failure mappings project from
`harness_lifecycle_policy.json`; private request schemas and native validation
attachments use the existing generator. Independent lifecycle, private-channel,
request-boundary and publication tests live with the actual Go owner. The existing
Node lifecycle tests remain temporary removal obligations for S-040, not an oracle
for native behavior.

Fresh evidence (each root is `.cartulary/test-results/<run-id>`; commands
are `CARTULARY_TEST_RUN_ID=<run-id> make <target>` unless selection is shown):

| Run ID | Exact target / selection | Owning Make exit | Retained result |
| --- | --- | --- | --- |
| `lifecycle-s038-generate-01` | `generate` | 0 | pass; 9810 ms |
| `lifecycle-s038-generate-02` | `generate` | 2 | see exact tool summary |
| `lifecycle-s038-generate-03` | `generate` | 0 | pass; 11220 ms |
| `lifecycle-s038-format-01` | `format` | 0 | pass; 8883 ms |
| `lifecycle-s038-runtime-01` | `harness-runtime-contract` | 2 | fail; 23350 ms |
| `lifecycle-s038-generate-04` | `generate` | 0 | pass; 10380 ms |
| `lifecycle-s038-format-02` | `format` | unavailable; retained pass only | pass; 8938 ms |
| `lifecycle-s038-runtime-02` | `harness-runtime-contract` | 2 | fail; 43740 ms |
| `lifecycle-s038-generate-05` | `generate` | 0 | pass; 10120 ms |
| `lifecycle-s038-format-03` | `format` | 0 | pass; 8586 ms |
| `lifecycle-s038-runtime-03` | `harness-runtime-contract` | 0 | pass; 45390 ms |
| `lifecycle-s038-generate-06` | `generate` | 0 | pass; 10020 ms |
| `lifecycle-s038-format-04` | `format` | 0 | pass; 8915 ms |
| `lifecycle-s038-runtime-04` | `harness-runtime-contract` | 0 | pass; 45290 ms |
| `lifecycle-s038-ui-lifecycle-01` | `harness-ui-review-lifecycle` | 2 | fail; 65570 ms |
| `lifecycle-s038-ui-workflow-01` | `harness-ui-review-workflow` | 2 | fail; 6440 ms |
| `lifecycle-s038-runtime-05` | `harness-runtime-contract` | 2 | fail; 43120 ms |
| `lifecycle-s038-ui-workflow-02` | `harness-ui-review-workflow` | 2 | fail; 6230 ms |
| `lifecycle-s038-generate-07` | `generate` | 0 | pass; 10130 ms |
| `lifecycle-s038-native-review-01` | `test-slice OWNER=harness.browser ROWS=harness.browser.unit.native_review` | 0 | pass; 2191 ms |
| `lifecycle-s038-ui-workflow-03` | `harness-ui-review-workflow` | 0 | pass; 16740 ms |
| `lifecycle-s038-ui-seeded-01` | `harness-ui-review-seeded-default` | 2 | fail; 41440 ms |

Failures and repairs: generate-02 rejected newly reachable unsupported path
patterns; explicit native predicates and parity cases now cover them. Runtime-01
found a generator fixture missing the authored lifecycle-policy input; fixture
closure repaired. Runtime-02 exposed missing receipt-parent directory creation;
the native writer now creates and verifies private parents. Public lifecycle-01,
workflow-01/02 and runtime-05 exposed valid actions rejected because their schema
IDs were reachable internally but not attached as callable native validators;
authored attachments and an independent public-request regression now cover that
boundary. Seeded-01 reached real services and browser operations but failed capture
because the new recipe lacked the attachment-check callback; the callback has been
restored at the physical recipe boundary. Seeded-02 is the pending rerun at this
checkpoint. No failed gate is DONE.

Native-review-01 passed against the manifest's dirty executable digest
`sha256:28ca667d0d2e173612ae867ad20676b48acf9909cc0ba3dd7124aaee65fb6eb1`.
Runtime-04 passed the native race/contract route. Workflow-03 passed all seven
public/diagnostic/font tests, including dev and artifact modes, exact browser
operations, reports, repeated terminal reads and private-detail removal. These
passes are source-scoped partial evidence; later edits require affected reruns.

Completion evidence supplements the earlier attempts above. Each invocation uses
`CARTULARY_TEST_RUN_ID=lifecycle-s038-<suffix> make <target>` and the matching
`.cartulary/test-results/lifecycle-s038-<suffix>` root. Diagnostic `make explain-run`
reads and target/owner explanations were read-only.

| Suffix | Target / selection | Owning exit | Retained outcome |
| --- | --- | --- | --- |
| `ui-seeded-02` | `harness-ui-review-seeded-default` | 2 | pass; 840 ms |
| `generate-08` | `generate` | 0 | pass; 9880 ms |
| `generate-09` | `generate` | 0 | pass; 10020 ms |
| `format-05` | `format` | 0 | pass; — ms |
| `native-review-02` | `test-slice OWNER=harness.browser ROWS=harness.browser.unit.native_review` | 0 | pass; — ms |
| `ui-seeded-03` | `harness-ui-review-seeded-default` | 0 | pass; 120960 ms |
| `native-review-03` | `test-slice OWNER=harness.browser ROWS=harness.browser.unit.native_review` | 0 | pass; — ms |
| `design-smoke-01` | `browser-design-review-smoke REVIEW_PROFILE=default` | 2 | failed; terminal exit 3; cleanup complete |
| `design-smoke-02` | `browser-design-review-smoke REVIEW_PROFILE=default` | 2 | failed; terminal exit 3; cleanup complete |
| `design-smoke-03` | `browser-design-review-smoke REVIEW_PROFILE=default` | unavailable | failed; terminal exit 3; cleanup complete |
| `generate-10` | `generate` | 0 | pass; 11240 ms |
| `generate-11` | `generate` | 0 | pass; 10090 ms |
| `format-06` | `format` | 0 | pass; — ms |
| `design-smoke-04` | `browser-design-review-smoke REVIEW_PROFILE=default` | 0 | closed; terminal exit 0; cleanup complete |
| `runtime-06` | `harness-runtime-contract` | 0 | pass; 820 ms |
| `ui-lifecycle-02` | `harness-ui-review-lifecycle` | 2 | fail; 65880 ms |
| `projection-drift-01` | `generate-drift` | 2 | fail; — ms |
| `projection-drift-02` | `generate-drift VERBOSE=1` | 2 | usage rejection; no admitted run |
| `projection-drift-03` | `generate-drift (VERBOSE=1 environment)` | 2 | fail; — ms |
| `ui-lifecycle-03` | `harness-ui-review-lifecycle` | 0 | pass; 65190 ms |
| `fixture-01` | `harness-fixture-contract` | 0 | pass; 4090 ms |
| `projection-drift-04` | `generate-drift` | 2 | fail; — ms |
| `ui-contract-01` | `harness-ui-review-contract` | 0 | pass; 2110 ms |
| `policy-01` | `generated-artifact-policy-check` | 0 | pass; — ms |
| `generate-12` | `generate` | 0 | pass; 10110 ms |
| `json-01` | `json-shape-check` | 2 | fail; — ms |
| `projection-drift-05` | `generate-drift` | 0 | pass; — ms |
| `json-02` | `json-shape-check` | 0 | pass; — ms |

Additional repairs and dispositions: seeded-02 exposed physical recovery output
escaping the UI private preparation root; native recovery now uses the same private
root as acquisition. Native recovery fixtures cover failed publication, preserved
primary failure, exact dead-owner recovery and immutable repetition. Concurrent IPC
request IDs serialize allocation and writes. Undrained preparation withholds
terminal publication and cleanup beneath live dependents.

Design-smoke-01/02 found a retired v4 incident sample. A shared current v5 archive
fixture now serves design review and the existing browser import helper. Attempt 02
used a temporary numeric source-location diagnostic; it was removed. Attempt 03
completed the import probes but failed the original Timeline row-count invariant
after those probes added rows. That invariant now runs before mutating probes;
all import and role assertions remain. Attempt 04 passed with cleanup complete.
Attempt 03's owning exit was not retained; its failed terminal is evidence only.

UI-lifecycle-02 exposed `target_unavailable` on intentional cancellation of busy
navigation. The physical adapter now reports interruption when its abort is active;
UI-lifecycle-03 passes all 51 checks, including exact public cancellation and borrowed
origin survival. Runtime-06 passes native race contracts; fixture-01 and UI-contract-01
pass. Seeded-03 passes the real seeded/editor/viewer and controller-death workflow;
its temporary legacy cleanup cases are not claimed as native policy evidence and
remain S-040 removal obligations.

Projection-drift-01/03/04 and json-01 rejected stale projection digests during
repairs. The scratch manifest now includes lifecycle policy inputs/outputs; Make's
binary build inputs include both new native packages, and design commands require
the installed binary. Generate-12, drift-05, policy-01 and json-02 establish closure.
Drift-02 rejected a command-line `VERBOSE` input; attempt 03 used its supported
environment form. One edit attempt stopped at a wrong JSON key after moving the
seed verification; the subsequent edit used `make_recipes` and completed normally.
Staged whitespace validation found trailing empty lines in the two new archive helpers;
those were removed and `git diff --cached --check` passed. Broad lint/finalization
remain S-042 obligations.

Compatibility: public commands, request bounds, retained versions and borrowed
resources stay stable. UI and design sessions now use direct native fixture policy;
old controller/transport files remain solely until the S-040 caller/test deletion
transaction. Node retains browser work and preparation recipes. No fallback engine
or resident service was added. S-038/V-039 is DONE; S-039 is READY. S-040–S-042
remain predecessor-gated. Qualification remains deferred and unchanged.

### E-043 — S-037/V-038 shared native runtime and scanner — DONE

Saved before S-038 begins. `harnesssession` now owns the native private-root,
secret registry, resource journal, preservation, stale cleanup, pinned-tree removal
and scanner implementations. Graph and fixtures import it directly; it imports
neither graph nor fixture orchestration. Graph-specific scan publication stays at
its existing owner. Removed fixture runtime/preservation tests moved to native
session tests; scanner fixtures retain independent expected outcomes and policy
identity. New contract/catalog routing includes the session owner without adding
full suites to fast smoke.

Secret registration is durable across short-lived producers, serialized across
handles with the private-directory lock, and fenced by finalization. Invalid,
foreign or replaced proof cannot authorize removal. Strict generated resource
schemas reject open process/renderer/diagnostic proof; semantic checks bind UID
and renderer identity. The secret-name predicate now matches mixed names such as
`PASSWORD_TOKENS` while still excluding ordinary capacity tokens. A bounded,
one-shot private-pipe session interface derives caller identity from the real
parent and keeps values out of arguments and retained output. It adds no daemon.
Tests include actual child transactions, forged ownership, concurrent handles,
late registration, preserved recovery and independent scanner fixtures.

Source: parent `439f4665f` plus staged source diff SHA-256
`364561a9d1838e48d6a00d29b2e68bd81a08887ad977382c0dfa1d1e5ae904de`.
This digest includes authored/new files and generated projections, excludes this
checkpoint, and identifies the final V-038 source. Environment remains Linux/WSL,
Go launcher 1.27.1/effective 1.27.2, Make-selected Node and existing local caches.
Each run below has its own root `.cartulary/test-results/<run-id>`; no root reused.

| Exact command | Owning exit / retained result |
| --- | --- |
| `CARTULARY_TEST_RUN_ID=lifecycle-s037-generate-01 make generate` | 2 / fail; new family row was not ASCII sorted |
| `CARTULARY_TEST_RUN_ID=lifecycle-s037-format-01 make format` | 2 / native build failed: old tree-helper name in browser adapter |
| `CARTULARY_TEST_RUN_ID=lifecycle-s037-generate-02 make generate` | 2 / fail; new selector test list was not ASCII sorted |
| `CARTULARY_TEST_RUN_ID=lifecycle-s037-format-02 make format` | 2 / graph preparation rejected the unreconciled catalog |
| `CARTULARY_TEST_RUN_ID=lifecycle-s037-generate-03 make generate` | 0 / pass, 10,250 ms |
| `CARTULARY_TEST_RUN_ID=lifecycle-s037-generate-04 make generate` | 0 / pass, 9,840 ms; typed responses added |
| `CARTULARY_TEST_RUN_ID=lifecycle-s037-format-03 make format` | 0 / pass, 8,919 ms |
| `CARTULARY_TEST_RUN_ID=lifecycle-s037-runtime-01 make harness-runtime-contract` | 2 / fail; test publisher signature and private fixture mode required migration |
| `CARTULARY_TEST_RUN_ID=lifecycle-s037-runtime-02 make harness-runtime-contract` | 0 / pass, 44,340 ms |
| `CARTULARY_TEST_RUN_ID=lifecycle-s037-generate-05 make generate` | 0 / pass, 10,400 ms; strict resource projection added |
| `CARTULARY_TEST_RUN_ID=lifecycle-s037-format-04 make format` | 0 / pass, 9,695 ms |
| `CARTULARY_TEST_RUN_ID=lifecycle-s037-runtime-03 make harness-runtime-contract` | 0 / pass, 45,920 ms; runtime/schema/session races |
| `CARTULARY_TEST_RUN_ID=lifecycle-s037-fixture-01 make harness-fixture-contract` | 0 / pass, 5,490 ms; direct shared-owner integration/races |
| `CARTULARY_TEST_RUN_ID=lifecycle-s037-drift-01 make generate-drift` | 0 / pass, 18,832 ms |
| `CARTULARY_TEST_RUN_ID=lifecycle-s037-policy-01 make generated-artifact-policy-check` | 0 / pass, 921 ms |
| `CARTULARY_TEST_RUN_ID=lifecycle-s037-json-01 make json-shape-check` | 2 / fail; schema attachment registry required sorted entries |
| `CARTULARY_TEST_RUN_ID=lifecycle-s037-generate-06 make generate` | 0 / pass, 9,780 ms; sorted authored attachment registry |
| `CARTULARY_TEST_RUN_ID=lifecycle-s037-drift-02 make generate-drift` | 0 / pass, 17,655 ms |
| `CARTULARY_TEST_RUN_ID=lifecycle-s037-policy-02 make generated-artifact-policy-check` | 0 / pass, 902 ms |
| `CARTULARY_TEST_RUN_ID=lifecycle-s037-json-02 make json-shape-check` | 0 / pass, 3,653 ms |
| `CARTULARY_TEST_RUN_ID=lifecycle-s037-runtime-04 make harness-runtime-contract` | 0 / pass, 44,930 ms on final generated source |
| `git diff --check` | 0 |

Failed gates were repaired at their authored owners; none counts as completion.
Generator and native contract details are in their target tool/step summaries;
graph targets retain graph/unit evidence. Failure diagnosis used exact-run
`make explain-run RESULTS_DIR=.cartulary/test-results/lifecycle-s037-json-01 DETAIL=logs TARGET=json-shape-check`
and bounded declared step logs for generator failures. A mistaken query for a
nonexistent helper filename was corrected by inspecting actual imports.

Public commands and retained acceptance sets are unchanged. Private session and
resource projections are new internal interfaces. Existing Node runtime/scanner
and fixture transport remain active only for callers scheduled in S-038/S-039;
F-037/F-040 cannot close until deletion. Real service/UI-review boundaries remain
S-038/S-041, full final checks S-042; those gates are not claimed here. S-038 may
begin after this checkpoint is saved. Qualification deferral and ledger unchanged.

### E-042 — S-036/V-037 contracts and characterization — DONE

Saved before S-037 begins. Owner REQ-852–855/AC-147–148 and REQ-502 now name the
shared native private-runtime and Go review lifecycle boundaries. Graph and UI
publication protocols stay distinct. Section 7 records the test disposition
ledger and Section 2 the concrete caller inventory. Native scan characterization
now uses fixed independent outcomes and policy digest rather than invoking Node.
No production implementation or generated projection changed in this slice.

Source: clean entry `b0bc9737531c7feaf573e28715348fd81ae0584a` plus the adopted owner,
tracker and one test change. Non-tracker diff SHA-256:
`35b09584ed64f7042ce5e135bbecc859c1d375d5a151074a4687d3f5b7526c07`.
The scanner test file SHA-256 is
`233f1ffcadd43a614dbda449c40187bbd158d055cf4e7788c11adeae4070fce2`.
Existing Linux/WSL environment, Go launcher 1.27.1/effective pinned 1.27.2,
Make-selected pinned Node and existing local caches. Roots below are under
`.cartulary/test-results/`; each command had its own native process handle.

| Exact command | Root suffix | Owning exit / harness result |
| --- | --- | --- |
| `CARTULARY_TEST_RUN_ID=lifecycle-s036-runtime-01 make harness-runtime-contract` | `lifecycle-s036-runtime-01` | 0 / pass, 45,910 ms, race-enabled runtime/schema suites |
| `CARTULARY_TEST_RUN_ID=lifecycle-s036-fixture-01 make harness-fixture-contract` | `lifecycle-s036-fixture-01` | 0 / pass, 3,920 ms, race-enabled fixture suite |
| `CARTULARY_TEST_RUN_ID=lifecycle-s036-ui-01 make harness-ui-review-contract` | `lifecycle-s036-ui-01` | 0 / pass, 2,050 ms |
| `CARTULARY_TEST_RUN_ID=lifecycle-s036-markdown-01 make lint-markdown` | `lifecycle-s036-markdown-01` | 0 / pass, 15,590 ms |
| `git diff --check` | none | 0 |

These direct contract routes publish target tool summaries, not an outer graph
manifest; an early attempt to inspect a nonexistent outer manifest is not a test
failure. Discovery of one nonexistent helper filename was corrected from actual
imports. No gate failed. No service acquisition or qualification run was required
for this characterization gate; real physical validation remains S-038/S-041.
Current public behavior is unchanged; the normative destination is adopted, with
implementation still pending. S-037 is ready after this saved checkpoint. Prior
completion evidence and F-022/RB-006 remain unchanged.


### E-041 — Lifecycle iteration admission and S-036 entry

Implementation authorization supersedes the prior planning-only turn. Clean main
is `b0bc9737531c7feaf573e28715348fd81ae0584a`. Inspection verified only the tracker
differs from prior executable `0725873db`. The earlier planning pass compared
6,917 file identities/content and the Git index unchanged; no tests ran then.
Current discovery: `make help`, `make help-all`, module-author task guides for
`harness.command_surface` and `harness.browser`, and summary explanations for
runtime/UI contract routes all exit 0. Environment is the existing Linux/WSL
checkout with pinned Make-selected toolchains. Baseline tests and their exact
roots/exits will be recorded before S-036 is marked DONE. No historical evidence,
salvage source or qualification inputs changed. An initial owner-text patch failed
context matching without changing files; corrected using exact current clauses.


### E-040 — S-032 final validation and handoff — DONE

Initial frozen executable commit `38b8758c8` (clean), digest
`sha256:d11fef10e28232f528c48534e3085a3b7ad5c6b41f91d2c6367ad4d8af6d1f86`.
`env CARTULARY_TEST_RUN_ID=remediate-s032-finalize-01 make agent-finalize`
passed, outer exit 0, 1/1, 22,906 ms. No source/projection refresh was produced.
`RESULTS_DIR` was unset: canonical-evidence and scheduler-drift retained-run
maintenance were explicitly skipped (`results-dir-not-provided`), recorded in
`finalize-01/unit-artifacts/finalize-summary.json`. Other maintenance passed.

Final gates use fresh `remediate-s032-<suffix>` roots in `.cartulary/test-results`.
`contract-01 make harness-contract` passed (99,201 ms), as did `scripts-01 make
lint-scripts` (661 ms), `shell-01 make lint-shell` (17,587 ms), `markdown-01 make
lint-markdown` (17,360 ms), `drift-01 make generate-drift` (23,364 ms),
`policy-01 make generated-artifact-policy-check` (1,063 ms), and `shapes-01 make
json-shape-check` (3,910 ms). All owning exits 0 and summaries PASS.
`check-01 make check` FAILED: owning Make exit 2, native primary exit 10;
978/981 units passed, three failed, none skipped/cancelled, 743,207 ms inclusive target wall.
All 1,240 cleanup steps passed. Exact root and failures are retained unchanged.
The selected native graph test consumed its shared 90-second deadline waiting for
whole-host Go admission (basic probe 45,014 ms admission versus 311 ms runner).
Use an existing one-token shell row for graph/receipt proof; keep the deadline,
independent readers and separate Go interop coverage. Stop dependent receipt
assertions after a failed graph so the primary error stays clear.

Another unit failure was `target:otel-conformance`: its source-text assertion
required a retired `harnessScope` export in the old facade. F-035 checks actual
pure projection output instead; no obsolete API is restored. Final status then
returned invalid/artifact_error (reader exit 11), revealing F-036: child smoke
had written legacy `cartulary.test_run_summary.v6` to the native root. Native
exclusive publication correctly refused the collision; canonical target evidence
remains, but root publication and scanning did not complete. Do not treat this as
a successful full run. Restrict graph-child smoke to target summaries and cover
that ownership through the supported public full check.

Post-failure repair checks: `format-01/02 make format-go` and `generate-01
make generate` passed (generation 9,560 ms); `otel-01 make otel-conformance`
passed (4,101 ms). `runtime-01 make harness-runtime-contract` failed (outer 2,
42,350 ms) because a new test attempted to select internal `check-harness-smoke`
as a standalone graph entry without a command identity. Removed that unsupported
test composition; the public full check remains the integration regression gate.
No runtime interface was expanded. `runtime-02 make harness-runtime-contract` passed (43,180 ms, outer 0).
`native-row-01 make test-slice OWNER=harness.command_surface
ROWS=harness.command_surface.unit.native_runtime CARTULARY_HARNESS_CACHE_MODE=off`
also passed through the inherited-admission public graph (36,716 ms, outer 0).
The final full check covers smoke root ownership through its supported entrypoint.

Repairs changed executable source and required a new freeze, agent-finalize,
affected checks and a fresh full warm run. The initial evidence remains scoped
to its original source; it is not substituted for the repeated final gates.

Read-only preservation comparison found all 6,937 salvage entries unchanged,
no additional tracked/nonignored path, and HEAD still `79aecc1ca`. No archive,
reset, evidence deletion, experimental acquisition or external export occurred.
Only tracker prose changed during check-01; the repair then changed the three
named executable files. `finalize-02 make agent-finalize` passed, outer 0,
21,062 ms, with no generated/source refresh. Retained-run maintenance remains
skipped because RESULTS_DIR is unset; check-01 is failed and ineligible. The new
second executable freeze has digest
`sha256:4da43d53e5da1260631541433f32f87d2f08245abf3d4b662e24c4d1d2ff3a65`;
the second broad validation pass uses clean frozen commit
`e7fe065cbdf58c5cacdcc0e263db06a1e559dabb`.

Repeated `scripts-02 make lint-scripts` (733 ms), `shell-02 make lint-shell`
(1,803 ms), `markdown-02 make lint-markdown` (8,600 ms), `drift-02 make
generate-drift` (19,520 ms), `policy-02 make generated-artifact-policy-check`
(1,152 ms), and `shapes-02 make json-shape-check` (3,970 ms) passed with owning
exits 0. `contract-02 make harness-contract` published a canonical PASS summary
(93,038 ms) and retained scan; the session had closed across context recovery
without its owning exit retained in the working context. Repeat that gate with
a fresh identity to record both outcomes explicitly. `check-02 make check` ran
under native session 4424 on that executable digest.

`check-02 make check` finished with owning Make exit 2: 980/981 units passed,
one failed, none skipped/cancelled, 697,070 ms canonical wall. All 1,239 cleanup
steps passed. `explain-run TARGET=check DETAIL=summary` and independent
`harness-observability-check` passed (one invocation/four sources; default-off
diagnostics/execution not observed). Canonical root publication and scan now
complete, closing the specific F-036 collision even though the full gate failed.
The sole failing unit was `target:lint-go` / `harness/child_target_failure`:
Staticcheck found unused broker acquisition state, redundant stdout initialization,
an overwritten test parameter and a repeated-call expression in the collector's
close-idempotence test. Remove dead state/assignments and name both close results
so both calls always execute. No deadline, interface or failure policy changes.
Reviewing the retained first run confirmed the same lint failure there; its
earlier checkpoint used a pre-terminal count and omitted this third failure.
The corrected final count and all three causes are preserved here.

The test-ops result-analysis reference now separates completed functional
service/container/wrapper coverage from deferred qualification. This is a
documentation-only correction, not an executable input or additional claim.

Post-lint repair: `format-03 make format-go`, `go-lint-01 make lint-go`
(format 8,620 ms, vet 1,160 ms, Staticcheck 1,570 ms), `fixture-01 make
harness-fixture-contract` (race enabled, 3,980 ms) and `collection-01 make
test-slice OWNER=harness.evidence_accounting
ROWS=harness.evidence_accounting.behavior.native_collection
CARTULARY_HARNESS_CACHE_MODE=off` (631 ms canonical wall) all passed, owning
exits 0. `finalize-03 make agent-finalize` passed, outer 0, 21,640 ms, without
generated/source changes. Retained-run maintenance was skipped because
RESULTS_DIR was unset; both prior full checks are failed and ineligible.
The repaired executable digest is
`sha256:1bdb3547e772d5011e1bd7ab66b85264eb5a25fe4198ad860fe06a17c9103307`.
Fresh final-source broad checks, full warm acceptance and the handoff below
close V-033; earlier failures remain retained with their original source.

Current executable freeze is clean `0725873db841e6db2a52013948d6982ffc392cf6`
with the `1bdb3547...` digest above. `contract-03 make harness-contract` passed,
owning exit 0, 123,452 ms; this supplies both required outcomes absent from the
contract-02 session record. `markdown-03 make lint-markdown` passed, outer 0,
8,210 ms. `check-03 make check` completed under native session 94526. These
functional checks overlap during initial execution; none is qualification evidence.

Fresh final-source physical/report checks (all owning exits 0):
`service-01 make service-backed-test-slice OWNER=module.database_migrations
ROWS=module.database_migrations.integration.schema_bootstrap
HARNESS_DIAGNOSTICS=basic CARTULARY_HARNESS_CACHE_MODE=off` passed in 14,791 ms;
`https-01 make service-backed-test-slice OWNER=harness.browser
ROWS=harness.browser.boundary_support.https_trust HARNESS_DIAGNOSTICS=basic
CARTULARY_HARNESS_CACHE_MODE=off` passed in 29,119 ms. Both exact roots passed
`make harness-observability-check RESULTS_DIR=<root>` (one invocation/four
sources, diagnostics/execution complete), `make test-run-status
RESULTS_DIR=<root> JSON=1`, and `make explain-run RESULTS_DIR=<root>
DETAIL=performance|resources` separately in human and `JSON=1` forms.
Captures are `/tmp/cartulary-s032-{service,https}-{performance,resources}.{txt,json}`
and `/tmp/cartulary-s032-{service,https}-status.json`; underlying exact roots
remain authoritative. Performance/resource JSON views are identical.

The short service run legitimately has no observed containers. The HTTPS run has
two owned containers with three samples each: memory maxima 180,613,120 and
78,041,088 bytes, limit 33,353,146,368 bytes, PID maxima 13 and 18, observed block
writes 32,202,752 and 94,208 bytes over 12,098 ms. Both observed block-read deltas
are measured zero. Container CPU percentages derive from valid paired samples;
allocation remains unavailable where the validated roster has no binding.
Human resource output includes both containers within 7,726 bytes and discloses
35 omitted scopes and 50 omitted lines. No totals across overlapping scopes,
new samples, longer workload or qualification claim were invented for reporting.

Final full warm command `env CARTULARY_TEST_RUN_ID=remediate-s032-check-03 make
check` passed, owning Make exit 0, canonical status PASS, 981/981 work units,
zero failed/skipped/cancelled, 848,697 ms canonical wall. Cache disposition is
209 hits, 609 misses and 163 bypasses; cached units are accepted warm evidence,
not claimed freshly executed tests or timing observations. All 1,031 cleanup
steps passed. This final-source graph includes `lint-go`, `lint-scripts`,
`lint-shell`, `lint-biome`, `generate-drift`, `generated-artifact-policy-check`,
`json-shape-check`, `otel-conformance`, smoke, product checks and builds. Their
actual final target/unit summaries are retained beneath the exact root.
The environment remains Linux/WSL with selected Go 1.27.2 and pinned repository
tools; toolchain digest is
`sha256:4b62e2558e6dcbe1d3c632fb9c0bb47e5a89e5dda08f103a72f61347b5ef1810`,
system digest
`sha256:546793d134ea4063515a25068aa22b78078fc1613c1ee56fb5f9e41c5c550588`.

Exact-run `make explain-run RESULTS_DIR=.cartulary/test-results/remediate-s032-check-03
TARGET=check DETAIL=summary`, `make test-run-status RESULTS_DIR=<same-root>
JSON=1`, `make harness-observability-check RESULTS_DIR=<same-root>`, and separate
human/JSON `make explain-run RESULTS_DIR=<same-root> DETAIL=performance|resources`
all exited 0. Canonical root schema is `cartulary.harness_run_summary.v1`, live
observation resolves to terminal PASS, and the retained-secret scan passed over
6,059 files. The independent audit passed one invocation/four sources; diagnostics
and execution observations are correctly not observed in this default-off run.
Performance/resource JSON facts match. Temporary views are
`/tmp/cartulary-s032-check-{performance,resources}.{txt,json}` and
`/tmp/cartulary-s032-check-status.json`. No export transport was invoked; the
passing evidence/harness contracts retain local projection and export-isolation
coverage. Scanner, status-reader and owning-process outcomes remain distinct.

Retained-run maintenance was skipped at each finalization because RESULTS_DIR
was unset and no eligible successful current-source full warm check existed then.
This final successful root is now available for future explicitly selected
maintenance; no maintenance pass is retroactively claimed. Experimental
acquisition, qualification, baseline refresh, default-on promotion and Core 05
publication were intentionally excluded. No required functional gate remains
skipped or failed on the final executable source.

Final salvage preservation comparison passed: all 6,937 tracked/nonignored
entries, modes, hashes and absence markers match; no additional path; HEAD remains
`79aecc1cae7a9474c92f4a9b8f77092bc4363662`. A first comparison helper incorrectly
compared the snapshot's `['absent']` marker to null and reported the 19 already
absent paths; correcting that comparison yielded zero differences. No source
was restored or modified in salvage. No new archive or evidence deletion occurred.

Current finding closure is explicit:

| Finding | Final disposition / evidence |
| --- | --- |
| F-019 | DONE; E-032–E-040 replace current summaries and preserve historical source-scoped failures |
| F-028 | DONE; E-035/E-037/E-038/E-039 and final contract/check: one Go runtime, compiler reaping, scheduler/cache/lifecycle/publication ownership |
| F-029 | Removal remains DONE; E-038 generated/import/build closure and final contract/check preserve retired experimental paths and independent qualification validator |
| F-030 | DONE; E-036/E-038/E-039, fixture race proof and final physical runs: one policy for graph/UI-review, physical owners retained |
| F-031 | DONE; Section 7 disposition ledger and E-038 close all changed test groups; no copied scheduler/coordinator or repeated full-suite smoke owner |
| F-032 | DONE; E-033/E-039 and final-source HTTPS fact/byte/omission parity above |
| F-033 | DONE; S-035/C-021/V-036/E-032 owner reconciliation and final documentation review |
| F-034 | DONE; E-036 real borrowed-session/TLS proof, unchanged optional private projection and final fixture/physical checks |
| F-035 | DONE; actual projection-scope assertions, final otel-conformance and full check PASS |
| F-036 | DONE; supported full check publishes one canonical root, terminal status and completed scan/audit |
| F-022/RB-006 | Unresolved; S-020–S-022 DEFERRED, all thresholds, failed dispositions and 44.952/90-minute ledger unchanged |

Public Make names, selection, failure semantics and retained acceptance versions
remain stable. Human reports now expose scope-specific observations. Private
runtime-plan/session interfaces migrate atomically to their adopted 32 MiB/1 MiB
bounds; obsolete private entrypoints, standalone composition roots, JS scheduler,
cache transactions, live writer, broker policy and duplicate test coordinators
are removed. Compilation, row runners, physical service/browser adapters and
offline readers remain. No permanent alternate-engine flag, compatibility shim
or new domain concept is introduced. Rollback requires a coherent source plus
authored/generated-projection transaction, not mixed owners or runtime selection.

Each predecessor checkpoint was saved before its successor:

| Workstream | Saved checkpoint |
| --- | --- |
| S-035 / E-032 | `6cb56a531` |
| S-034 / E-033 | `531fe6f92` |
| S-031/A / E-034 | `ae39f79f0` |
| S-031/B / E-035 | `4e37fdd3f`, whitespace closure `4ebcdacdd` |
| S-031/C / E-036 | `60767f7b1` |
| S-031/D / E-037 | `8fe6c3d3c` |
| S-031/E / E-038 | `52713a9bf` |
| S-031/F / E-039; parent S-031 | `38b8758c8` |
| S-032 / E-040 | Repair freezes `e7fe065cb` and `0725873db`; this final handoff-only checkpoint closes V-033 |

Functional remediation is complete. Residual observations remain sampled lower
bounds with unavailable Windows-host/Docker-internal/short-lived-process detail;
no aggregate attribution or overhead certification follows. The original broader
effort stays IN_PROGRESS only for explicitly deferred qualification. A future
qualification effort requires separate adoption and complete eligible, stable
evidence under unchanged gates. No further functional successor is pending.
Final documentation gate `env CARTULARY_TEST_RUN_ID=remediate-s032-markdown-04
make lint-markdown` passed, owning exit 0, 7,050 ms. `git diff --check HEAD`
passed, all twelve numbered sections remain, and only this tracker changed
after the validated executable freeze.

### E-039 — S-031/F public boundary checkpoint; S-031 complete

Executed the installed public surface on clean `52713a9bf` with executable digest
`sha256:d11fef10e28232f528c48534e3085a3b7ad5c6b41f91d2c6367ad4d8af6d1f86`,
Linux/WSL and unchanged pinned toolchain. No executable edits were necessary.
Discovery used current owner task guides and target explanations. Each writing
command used `env CARTULARY_TEST_RUN_ID=remediate-s031f-<suffix> make <target>`;
all owning process exits 0, all retained summaries PASS. Primary roots are
`.cartulary/test-results/remediate-s031f-<suffix>`; no failed F attempt or skipped
required boundary. Concurrency is functional evidence only.

| Suffix | Exact target | Result |
| --- | --- | --- |
| runtime-01 | harness-runtime-contract | PASS, 16,590 ms |
| fixture-01 | harness-fixture-contract | PASS, 3,550 ms |
| evidence-01 | harness-evidence-contract | PASS, 8,510 ms |
| command-01 | harness-command-surface-contract | PASS, 53,380 ms |
| ui-01 | harness-ui-review-contract | PASS, 2,180 ms |
| lifecycle-01 | harness-ui-review-lifecycle | PASS, 68,210 ms |
| seeded-01 | harness-ui-review-seeded-default | PASS, 121,630 ms |
| boundary-01 | harness-runtime-boundary-contract | PASS, 37,290 ms |
| local-01 | test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.local_migration_activities HARNESS_DIAGNOSTICS=basic CARTULARY_HARNESS_CACHE_MODE=off | PASS, 1,842 ms |
| shell-01 | test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.current_epoch_evidence HARNESS_DIAGNOSTICS=basic CARTULARY_HARNESS_CACHE_MODE=off | PASS, 9,912 ms |
| vitest-01 | test-slice OWNER=package.ui ROWS=package.ui.boundary_support.design_tokens_exposes_explicit_font_role_tokens_55f6ac41b0 HARNESS_DIAGNOSTICS=basic CARTULARY_HARNESS_CACHE_MODE=off | PASS, 4,662 ms |
| service-01 | service-backed-test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.schema_bootstrap HARNESS_DIAGNOSTICS=basic CARTULARY_HARNESS_CACHE_MODE=off | PASS, 14,223 ms |
| https-01 | service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.https_trust HARNESS_DIAGNOSTICS=basic CARTULARY_HARNESS_CACHE_MODE=off | PASS, 30,352 ms |
| object-01 | service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.integration.object_store_fixture_roundtrip CARTULARY_HARNESS_CACHE_MODE=off | PASS, 12,799 ms |

Evidence and command contracts cover malformed public input, real wrapper
nesting/early failure/SIGINT/SIGTERM, native failure/cancellation/descendant drain,
private output isolation, bounded readers and no-provisioning status. UI-review
seeded lifecycle proves actual native shared borrowing, rejects stopping a live
borrower, tolerates repeated native close and preserves borrowed services until
the physical owner closes them. Runtime/fixture race contracts pass. Native
boundary retains independent-reader PostgreSQL/HTTPS roots beneath
`boundary-01/native-boundary/{postgres,https}`.

Read-only `make test-run-status RESULTS_DIR=<service-01-root> JSON=1` first
observed live revision 3, executing with one passed/one running/one pending unit,
then terminal pass. `make harness-observability-check RESULTS_DIR=<root>` passes
for service-01 and https-01 (one invocation, four sources, diagnostics and
execution complete). `make explain-run RESULTS_DIR=<root> DETAIL=resources`
and `JSON=1` pass for both; performance human/JSON pass for service-01. Saved
views are `/tmp/cartulary-s031f-{service,https}-resources.{txt,json}` and
`/tmp/cartulary-s031f-service-performance.{txt,json}`, with terminal status in
`/tmp/cartulary-s031f-service-status.json`. Retained roots are authoritative.
Human/JSON inspected values agree: service container memory 127,504,384 and
76,034,048 bytes, limits 33,353,146,368, PID gauges 9/17; HTTPS memory
177,057,792 and 77,070,336, PID 13/18. Each has one sample, so CPU and I/O deltas
are unavailable, not zero. Allocation is honestly unobserved where the validated
roster has no allocation reference. No change to cadence or accepted payloads.

All Section 7 dispositions and E deletion obligations close. One production graph
owner and one fixture policy serve migrated callers; no retired experimental
route remains. F-028/F-030/F-031 pass their functional V-032 boundary. S-031/A–F
and parent S-031 are DONE. S-032 may start after this checkpoint; it must run
agent-finalize before frozen-source broad checks and full warm check. Broader
original qualification stays unresolved/deferred with unchanged budgets/history.

### E-038 — S-031/E atomic cutover and deletion checkpoint

Starting main `8fe6c3d3c682d7e66b199c309abc3e70f4da9a6d`, Linux/WSL and the
pinned E-035 toolchain. Selectively reviewed dirty salvage contributions; never
imported its scheduler simulation or copied wrapper coordinator. Public graph and
UI-review callers now use one installed `cartulary-harness` and shared Go fixture
policy. Authored task surface, renderer, build inputs, helper ownership, family
routing and reachability inputs were changed; projections regenerated. Removed
JS scheduler/executor/cache transactions/live publication, broker policy, collector
bridge, wrapper coordination, staged native review module and three standalone
composition roots. Retained compiler/cache-plan facts, Node row runners, physical
adapters, native collector isolation and independent readers.

Additional boundary repairs: closed argv binding for Go smoke commands and Go
readiness prerequisites; bounded named native smoke tests; semantic generated
routing/bootstrap tests; no JSON-whitespace coupling. Added actual browser
failure/reset continuation and private-capture cancellation/cleanup tests. Retained
scanner and shell-row diagnostics directly at producer boundaries. Section 7
records the assertion dispositions. Source/import/build searches show no retired
owner, candidate simulation/copy or experimental acquisition route.

All commands used distinct `env CARTULARY_TEST_RUN_ID=remediate-s031e-<suffix>
make <target>` roots under `.cartulary/test-results/`. Successful owning exits are
0; failed Make attempts exit 2 and retain their failure summaries.

| Suffix | Exact target | Result / applicability |
| --- | --- | --- |
| format-01/02/03/04 | format-go | PASS; native test/source formatting |
| generate-01/02 | generate | FAIL: unsafe authored regexp smoke token; second correction targeted wrong token index |
| generate-03/04/05/06 | generate | PASS; final 9,830 ms |
| runtime-01 | harness-runtime-contract | FAIL: test expected normalized instead of raw child exit and omitted private run directory |
| runtime-02/03 | harness-runtime-contract | PASS race; final 19,040 ms including new private-capture cancellation |
| contracts-01 | harness-contract-tests | FAIL, 90,380 ms: unsorted helper facade registry |
| contracts-02/03/04 | harness-contract-tests | PASS; final 93,490 ms, scanner and shell-row cases retained |
| shapes-01 | json-shape-check | FAIL: same registry ordering defect |
| shapes-02/03, closure-01 | json-shape-check | PASS; final 3,644 ms |
| projection-01/02 | generated-artifact-policy-check | PASS; final 928 ms |
| drift-01/02/03 | generate-drift | PASS; final 19,444 ms |
| smoke-01 | harness-smoke-work-graph-scheduler-smoke | Rejected: internal check ID has no Make target; no run evidence |
| smoke-02 | run-harness-smoke-fast | FAIL: Go argv bindings were passed literally |
| smoke-03 | run-harness-smoke-fast | PASS, 6,342 ms; named owner tests and public wrapper |
| extended-01 | run-harness-smoke-extended | FAIL: stale dispatcher test name and removed compiler fixture |
| extended-02 | run-harness-smoke-extended | FAIL: stale separate-helper bootstrap expectation |
| extended-03 | run-harness-smoke-extended | FAIL: JSON indentation assertion and canonical-summary assumption before runtime bootstrap |
| extended-04 | run-harness-smoke-extended | PASS, 51,785 ms; all 43 checks including real native boundary |

Initial `git diff --check` rejected an added EOF blank line before format-01; it
was removed. Failed roots were inspected through exact `make explain-run
RESULTS_DIR=<root> TARGET=<target>` and referenced logs, never overwritten.
Extended-04 native HTTPS manifest records dirty source digest
`sha256:bb003bf031c6ec247489917f7093b8cecce59720c3b4229a7684e829fed6cfa0`.
Subsequent scanner/row test restoration is validated by contracts-04; final
shapes-03/drift-03/projection-02 record dirty source at starting commit, digest
`sha256:d11fef10e28232f528c48534e3085a3b7ad5c6b41f91d2c6367ad4d8af6d1f86`.
`git diff --check HEAD` passes. No executable change follows these final checks.

S-031/E is DONE. Compatibility: public commands and retained acceptance versions
stay stable; private JS APIs and separate installed helpers are retired without
aliases. Pre-runtime bootstrap errors remain owning-process diagnostics. F may
begin only after this saved checkpoint. F still must prove all public physical
boundaries and exact-run views; S-032 still owns final frozen-source validation.
Default-off/local-only and all qualification deferrals/budgets remain unchanged.

### E-037 — S-031/D observation and finalization checkpoint

Starting main `60767f7b1`, Linux/WSL and E-035's pinned toolchain. Reviewed and
integrated authored observation contributions from `ba543bf69..79aecc1ca`;
retained C's stronger race routes, TLS repair and staged review client. Added
native graph composition, live snapshots, bracketed clocks, execution observation,
collector client/identity proof, exact container registration, final producer
drain, retained scanning, recovery preservation and wrapper receipt registration.
One runtime binary composes these libraries; public dispatch changes remain E.
Current cadence, resource/execution/envelope/receipt versions and local-only,
default-off posture are unchanged. No second collector or late ordinary writer.

Commands used `env CARTULARY_TEST_RUN_ID=<id> make <target>`, distinct primary
`.cartulary/test-results/<id>` roots. All summaries PASS and owning process exits 0:

| Run ID suffix after `remediate-s031d-` | Target | Result |
| --- | --- | --- |
| format-01 | format-go | PASS |
| generate-01 | generate | PASS, 9,650 ms |
| runtime-01 | harness-runtime-contract | PASS, 18,220 ms; race, off/basic, two native graphs and independent receipt reader; session 98111 |
| fixture-01 | harness-fixture-contract | PASS, 3,950 ms; race and new preservation cases; session 57500 |
| evidence-01 | harness-evidence-contract | PASS, 16,620 ms; independent retained reader/hostile facts; session 36670 |
| boundary-01 | harness-runtime-boundary-contract | PASS, 42,180 ms; session 38906 |
| drift-01 | generate-drift | PASS, 4/4, 18,771 ms; session 21075 |

The boundary route retained exact nested roots `boundary-01/native-boundary/
postgres` (3/3, 15,597 ms) and `https` (11/11, 23,512 ms). Both report source
commit `60767f7b1`, dirty, digest
`sha256:7059c594b484de9b038ba2e81bba6cd114dd62b0553ee8e9eeabac6cd4c86806`.
Independent Node readers accepted native canonical/live/resource/execution/envelope
facts. Runtime integration separately exercised off/basic and multi-graph receipt
binding; scan fixtures reject unsafe paths and secrets, collector death drains,
clock closure and quiet-window assertions pass. No source changed during checks.

Exact HTTPS root read-only checks: `audit-01 make harness-observability-check
RESULTS_DIR=.cartulary/test-results/remediate-s031d-boundary-01/native-boundary/https`
PASS (complete diagnostics/execution, 4 sources); `status-01 make test-run-status`
with the same RESULTS_DIR and JSON=1 reports terminal pass; `resources-01 make
explain-run` with the same RESULTS_DIR and DETAIL=resources exposes both owned
containers' memory/limit, PID, CPU and block-I/O deltas. Human output saved at
`/tmp/cartulary-s031d-resources.txt`; retained facts remain in the exact root.
Reader success does not substitute for the separately observed native process exit.
`git diff --check` passed. No failed D attempt.

S-031/D/V-031 is DONE. E may begin after this saved checkpoint. Remaining risks
are public wiring, producer interoperability and deletion closure, explicitly
reserved for E/F. Qualification evidence and budgets remain unchanged.

### E-036 — S-031/C shared fixture checkpoint

Starting main `4ebcdacdd1ea96caaa8e3f34bc2a29ef20df6cff`, Linux/WSL and the same
pinned toolchain as E-035. Reviewed authored `9fbcecdcd..ba543bf69` contributions:
`harnessfixtures` owns lease/reuse/taint/partial-acquisition/cleanup/recovery and
bounded private sessions; graph calls its library. Physical service/browser and
local-session adapters remain their existing owners. The private UI client and
`review-preparation-native.mjs` are staged for contract/lifecycle validation;
public graph and review dispatch remain together on their old composition until E.
This temporary staging module must replace the public module and disappear in E.
Generated validators and routes were regenerated; both native routes use race
checks. Existing private/public versions remain unchanged.

Added a real persistent borrowed-suite case: actual native acquisition, owner-stop
rejection while borrowed, repeated close, zero borrowers, same live services, then
owner cleanup. It exposed F-034: private local-session schema rejected existing
TLS certificate/key/trust bindings. Added those optional fields and derived closed
environment selection from the schema instead of maintaining another allowlist.
No secrets are added to retained diagnostics; the descriptor stays private.

Every command used `env CARTULARY_TEST_RUN_ID=<id> make <target>` and fresh primary
`.cartulary/test-results/<id>` roots. PASS means summary plus owning process exit 0:

| Run ID suffix after `remediate-s031c-` | Target / inputs | Outcome |
| --- | --- | --- |
| generate-01 | generate, after rejected preflight patch; no candidate edit yet | PASS, 9,550 ms |
| generate-02 / format-01 | generate / format-go | PASS, 9,680 ms / exit 0 |
| fixture-01 | harness-fixture-contract | PASS, 4,690 ms, session 73724 |
| runtime-01 | harness-runtime-contract | PASS, 12,270 ms, session 95502 |
| ui-contract-01 | harness-ui-review-contract | PASS, 2,020 ms, session 97907 |
| ui-lifecycle-01 | harness-ui-review-lifecycle | PASS, 65,450 ms, session 56158 |
| ui-seeded-01 | harness-ui-review-seeded-default | PASS, 115,020 ms, session 61988 |
| service-01 | service-backed-test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.schema_bootstrap HARNESS_DIAGNOSTICS=basic CARTULARY_HARNESS_CACHE_MODE=off | PASS, 3/3, 14,971 ms, session 31537 |
| browser-guide-01 | task-guide ROLE=module-author OWNER=harness.browser | PASS |
| https-01 | service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.https_trust CARTULARY_HARNESS_CACHE_MODE=off | PASS, 11/11, 26,951 ms, session 74202 |
| object-01 | service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.integration.object_store_fixture_roundtrip CARTULARY_HARNESS_CACHE_MODE=off | PASS, 3/3, 9,543 ms, session 76428 |
| drift-01 | generate-drift | FAIL, 3/4, outer exit 2; topology digest stale after EOF formatting |
| drift-02 | generate-drift VERBOSE=1 | REJECTED config/configuration_error, outer exit 2; VERBOSE is not an accepted public input |
| generate-03 / drift-03 | generate / generate-drift | PASS, 9,590 ms / 4/4 |
| shapes-01 | json-shape-check | PASS, 3/3 |
| generate-04 / drift-04 | generate / generate-drift after adding borrowed case | PASS, 9,620 ms / 4/4 |
| ui-seeded-02 | harness-ui-review-seeded-default | FAIL, outer exit 2; F-034 schema rejection during real persistent creation |
| generate-05 / drift-05 | generate / generate-drift after F-034 repair | PASS, 9,670 ms / 4/4 |
| ui-seeded-03 | harness-ui-review-seeded-default, includes actual borrowed case | PASS, 118,990 ms, session 2343 |
| contracts-01 | harness-contract-tests, includes local-session cases | PASS, 108,400 ms, session 12814 |

Failed evidence remains retained. `drift-explain-01` summary and
`drift-explain-03` logs diagnosed the drift; `drift-explain-02` used unsupported
DETAIL=artifacts and was rejected with usage_error/exit 2. `seeded-explain-01`
identified the exact failed seeded run; its referenced stdout identified F-034.
Initial `git apply --check` rejected a generated render-index hunk before edits;
removed derived outputs from the authored import and generated them instead.
Removed one extra EOF blank line before execution. `git diff --check` passes.

Pre-F-034 service manifest source digest was
`sha256:2909c95ae47a841a0b02b199868af46d9f2389faa88ecfc8fde41ed35623d1ea`;
final drift manifest is source `4ebcdacdd1`, dirty, digest
`sha256:d7eaad72f7d000ad2e423327400c875e4b7abd13215029beb8dc2c559b38db3e`.
Final seeded and contract checks ran on that repaired executable source. Source
was unchanged during each final validation. Generation refreshed only owner
projections; exact summaries/logs are within the named roots.

S-031/C/V-030 and F-034 are DONE. Shared policy is validated; public activation
and JS-policy deletion remain E obligations. D may begin after this saved
checkpoint. No qualification acquisition, telemetry export or certification claim.

### E-035 — S-031/B native core checkpoint

Starting main `ae39f79f033792bf59f061d936ea936dcc4f38d1`, Linux/WSL, pinned
Go effective 1.27.2. Reviewed and applied only authored native preparation/core
changes from `c09e8ba87..9fbcecdcd`; regenerated projections from those owners.
No salvage checkout content was modified. Added cohesive `harnessruntime` and
shared `harnessschema` libraries, a thin binary composition root, bounded strict
plan/compiler handoff, scheduling/admission/process drain, cache transactions,
canonical accounting and independent contract fixtures. Public graph dispatch is
still Node for this checkpoint. The restored runtime route uses `-race -count=1`.

All commands below used `env CARTULARY_TEST_RUN_ID=<id> make <target>`, distinct
primary `.cartulary/test-results/<id>` roots and owning process exit 0:

| Run ID | Target | Result |
| --- | --- | --- |
| remediate-s031b-generate-01 | generate | PASS, 9,650 ms |
| remediate-s031b-format-01 | format-go | PASS |
| remediate-s031b-generate-02 | generate | PASS, 9,530 ms; includes race route |
| remediate-s031b-runtime-01 | harness-runtime-contract | PASS, 12,220 ms; native core/race and schema parity |
| remediate-s031b-projection-01 | generated-artifact-policy-check | PASS, 3/3 units |
| remediate-s031b-shapes-01 | json-shape-check | PASS, 3/3 units |
| remediate-s031b-command-01 | harness-command-surface-contract | PASS, 51,130 ms |
| remediate-s031b-drift-01 | generate-drift | PASS, 4/4 units |

Native execution manifest is `remediate-s031b-runtime-01/native-engine/
selected-graph-38759/run-manifest.json`, source commit `ae39f79f0`, dirty,
source digest `sha256:098c084d0472f59e2f7ec7dd79739ecc0530136c63be0878a5d0dcf585787ee2`,
toolchain `sha256:4b62e2558e6dcbe1d3c632fb9c0bb47e5a89e5dda08f103a72f61347b5ef1810`.
Native sessions 88745/21977/52588/47623/92504 all drained with exit 0; tool summaries
and independent Node canonical validation agree. The staged whitespace check then
found one extra EOF blank line in new `preparation.mjs`; removed before C.
The prior unstaged check did not inspect that untracked file.
A read-only `remediate-s031b-status-01 make test-run-status
RESULTS_DIR=.cartulary/test-results/remediate-s031b-command-01 JSON=1` returned
usage_error/exit 2 because that direct contract run has no canonical graph status;
its owning session and terminal tool summary establish the result instead.
After the EOF correction, `remediate-s031b-runtime-02 make harness-runtime-contract`
PASS/exit 0 (session 66366), source commit `4e37fdd3f`, dirty, digest
`sha256:ef849d94dd176458584433c72ee90bdae25fabbd0ff9f270ca2aa1359a80426b`. Staged whitespace check now passes.
An initial editor command selected a nonexistent JSON key and exited before any
write; corrected the owner key. No runtime/test gate failed.

V-028/V-029 and S-031/B are DONE on current mainline source. Private plan/schema
and generated build inputs migrate together; retained public formats are unchanged.
Fixture and observation composition remain for C/D, public cutover/deletion for E.
C may begin after this saved checkpoint. No qualification or broad product gate
ran; remaining certification limits are unchanged.

### E-034 — S-031/A characterization checkpoint

At `531fe6f92`, reviewed the committed native stages and dirty candidate separately.
Section 7 records twelve concrete test groups and their durable behavior/owner/
route disposition. Source/compiler and independent reader fixtures remain Node;
scheduler, executor, cache transaction and fixture policy move to actual Go owners.
Copied simulation/wrapper algorithms and repeated full-suite smoke are rejected.
The inspected native tests use explicit expected ordering, arithmetic and outcomes;
they are candidate fixtures, not current-main passes.

Current-main baseline: `env CARTULARY_TEST_RUN_ID=remediate-s031a-guide-01 make
task-guide ROLE=module-author OWNER=harness.command_surface` PASS/exit 0.
`remediate-s031a-command-01 make harness-command-surface-contract` PASS/exit 0,
52,560 ms, session 73279. `remediate-s031a-contracts-01 make harness-contract-tests`
PASS/exit 0, 104,160 ms, session 93290. Exact roots are primary
`.cartulary/test-results/<run-id>`; each has its own tool-run summary. S-034's
evidence contract covers the unchanged retained-reader/observation baseline.
No source changed while either baseline check was active. No service reset,
qualification acquisition, or broad product gate ran. No failed A attempt.

S-031/A is DONE. B may begin after this saved checkpoint. B's current-source
native tests must pass before C; no historical candidate pass discharges that gate.
The test disposition ledger remains open for E/F closure. Qualification and its
failed evidence/44.952-minute ledger are unchanged.

### E-033 — S-034 report completion

Starting commit `6cb56a531`, Linux/WSL and pinned toolchain digest
`sha256:4b62e2558e6dcbe1d3c632fb9c0bb47e5a89e5dda08f103a72f61347b5ef1810`.
Changed `performance-presentation.mjs`, `resource-presentation.mjs` and
`tests/test-instrumentation.mjs` under observability. Human rendering now consumes
catalog-owned scopes/aggregation/units and the validated physical roster. Selected
containers precede verbose process detail within the existing human bound.
Gauges, valid counter intervals, unavailable/zero, resets, borrowed/shared context
and omissions remain explicit. JSON accepted shapes and selected facts are unchanged.
No sampler, filesystem read, mutable cache or new schema was added.

All following Make invocations use `env CARTULARY_TEST_RUN_ID=<root>`; result roots
are `.cartulary/test-results/<root>` in primary. `remediate-s034-guide-01`:
`task-guide ROLE=module-author OWNER=harness.evidence_accounting`, PASS/exit 0.
`remediate-s034-evidence-01` and `-02`: `harness-evidence-contract`, FAIL/outer
exit 2; newly authored fixtures respectively used forbidden container:0 and
shadowed the metric helper. Both repaired; failures retained. `-03`: same target,
PASS/exit 0 (native session 32133). Independent report arithmetic and truncation
fixtures pass, including all four scopes and non-additive container accounting.

`remediate-s034-service-01` and `-02`: `service-backed-test-slice
OWNER=module.database_migrations ROWS=module.database_migrations.integration.schema_bootstrap
HARNESS_DIAGNOSTICS=basic CARTULARY_HARNESS_CACHE_MODE=off`, both PASS/exit 0,
3/3 units. First run exposed human byte-budget starvation; compact metrics and
container-first presentation repaired it. Final run duration 15,141 ms, native
session 49761; source digest
`sha256:cf230c92c59116c06faec49738627b9da6af9b2f65d77f5f95a7144b6528b943`.
Earlier run retains its different source identity.

Final exact service root passed `harness-observability-check` and `explain-run
DETAIL=resources` human plus `DETAIL=resources|performance JSON=1` (exit 0).
Both JSON facts are identical; human output shows both owned containers' sampled
memory/limits/PIDs and valid CPU/block-I/O intervals. Full output remains bounded
with explicit omissions. `git diff --check` passes. This is functional evidence;
no external telemetry, qualification acquisition, baseline update or broad check.
V-035/F-032 complete; S-031/A ready after this saved checkpoint. Rollback is the
three-file presentation/test transaction; all retained evidence stays in place.

### E-032 — S-035 specification completion

2026-10-11, primary main starting at `3781d9a45`, clean index/source. Only the
Testing Harness NLSpec and this tracker changed. Human review reconciled v2
instrumentation definitions, all eleven availability tokens, current container
and wrapper coverage, qualification validation/retirement, scope-aware human
reports and the final native ownership/interface requirements. AC-146–149 and
Section 17.2 traceability now cover the adopted boundaries. Runtime schema
acceptance sets are unchanged; native private projections are future B/C work.
Domain vocabulary required no amendment. No executable consumer reads Markdown.

`env CARTULARY_TEST_RUN_ID=remediate-help-01 make help` and the corresponding
`remediate-help-all-01 make help-all` both exited 0. `env
CARTULARY_TEST_RUN_ID=remediate-s035-markdown-01 make lint-markdown` PASS, outer
exit 0, 15,700 ms; exact root `.cartulary/test-results/remediate-s035-markdown-01`,
summary `adhoc/lint-markdown/tool-run-summary.json`. Native session 86631 drained.
`git diff --check` passed. These are documentation checks, not product or
qualification evidence. No generator, formatter, service or product test ran.

V-036/F-033 are complete. The salvage checkout remains a read-only candidate.
S-034 may begin only after this saved checkpoint. Rollback is the two-document
transaction; preserve the previous history and all evidence. Qualification remains
DEFERRED, F-022/RB-006 unresolved and the ledger 44.952/90 minutes.

### E-031 — Mainline cleanup and planning checkpoint

Environment: Linux `6.6.114.1-microsoft-standard-WSL2`, pinned Node 24.15.0 and
selected Go 1.27.2 (launcher 1.27.1), existing Docker Desktop integration. User
reported no other intensive workload during the earlier experiments; the current
functional runs make no timing-noise or qualification claim.

| Commit on main | Coherent change / validation |
| --- | --- |
| `9c021eedb` | Preserve authored deployment tmp marker; release-evidence contract |
| `4502046c2` | Bootstrap disposable SeaweedFS with exactly one master; focused fixture and real service/browser gates |
| `ca073dab3` | Batch readiness content hashing, independent byte/digest/failure oracle |
| `cd8fe7cf6` | Bounded preparation/hash lifetimes, selected-row checks, metadata fixtures and lazy cache bypass; command/evidence and projection tests |
| `229936209` | Coordinated native observation/strict retained validation/schema cutover; independent qualification validator retained; experiments excluded; native/evidence/service/public gates |

The code commit's executable tree equals `c09e8ba87` exactly. Only the NLSpec
changes relative to that reference: keep experiment retirement/deferred
qualification, remove not-yet-integrated native runtime requirements, and use
existing Git history instead of requiring new code archives. No generated file
was hand-edited. The unfinished salvage worktree and qualification worktree were
not modified. Hash comparison covered all starting primary tracked/non-ignored
files immediately before the exact-path transfer. A final hash/index/HEAD check
also confirmed the unfinished salvage checkout remained untouched. The original staged tracker
was preserved until this authorized documentation reconciliation.

All cleanup-checkout roots below are under
`/home/jochi/.codex/worktrees/instrumentation-main-cleanup/cartulary/.cartulary/test-results/`.
Each row used `env CARTULARY_TEST_RUN_ID=<root> make <arguments>` from that checkout.
No elapsed value below is an overhead qualification result.

| Root | Make arguments | Result |
| --- | --- | --- |
| `main-cleanup-bootstrap-01` | `frontend-install` | PASS; isolated checkout setup |
| `main-cleanup-deployment-01` | `release-evidence-contract` | PASS |
| `main-cleanup-object-store-01` | `test-slice OWNER=harness.browser ROWS=harness.browser.unit.object_store_fixture_admission` | PASS; 3/3 graph units |
| `main-cleanup-preparation-01` | `harness-contract-tests` | FAIL preserved; 191/192 passed, stale generated topology after assembling catalog inputs; new readiness/catalog tests passed |
| `main-cleanup-preparation-generate-01` | `generate` | PASS; repaired stale projection |
| `main-cleanup-preparation-shapes-01` | `json-shape-check` | PASS |
| `main-cleanup-preparation-command-01` | `harness-command-surface-contract` | PASS |
| `main-cleanup-observation-generate-01` | `generate` | PASS |
| `main-cleanup-evidence-01` | `harness-evidence-contract` | PASS, including retained validation, clock, receipt, collector-death and qualification arithmetic fixtures |
| `main-cleanup-command-01` | `harness-command-surface-contract` | PASS; negative fixture child failures are expected assertions |
| `main-cleanup-native-01` | `test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_collection,harness.evidence_accounting.behavior.native_observation,harness.evidence_accounting.behavior.observation_admission,harness.evidence_accounting.behavior.local_migration_activities` | PASS; 4/4 graph units |
| `main-cleanup-drift-01` | `generate-drift` | PASS |
| `main-cleanup-policy-01` | `generated-artifact-policy-check` | PASS |
| `main-cleanup-shapes-01` | `json-shape-check` | FAIL preserved; obsolete v1 schemas remained because assembly initially used rename-detected path enumeration; repaired with explicit no-rename deletions |
| `main-cleanup-shapes-02` | `json-shape-check` | PASS after removal |
| `main-cleanup-service-01` | `service-backed-test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.schema_bootstrap HARNESS_DIAGNOSTICS=basic` | PASS; two owned containers sampled, no container deadline overruns/omissions; retained scan/receipt complete |
| `main-cleanup-testservices-01` | `test-slice OWNER=harness.browser ROWS=harness.browser.unit.testservices_lifecycle_contract` | PASS |
| `main-cleanup-browser-01` | `service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.https_trust HARNESS_DIAGNOSTICS=basic` | PASS; 11/11 graph units |
| `main-cleanup-finalize-01` | `agent-finalize` | PASS, generated files unchanged; retained-run/performance maintenance SKIPPED because RESULTS_DIR unset |
| `main-cleanup-final-contract-01` | `harness-contract` | PASS on committed executable source; 2/2 graph units |
| `main-cleanup-final-scripts-01` | `lint-scripts` | PASS |
| `main-cleanup-final-shell-01` | `lint-shell` | PASS |
| `main-cleanup-final-drift-01` | `generate-drift` | PASS |

Fresh service reader commands passed: `make harness-observability-check
RESULTS_DIR=.cartulary/test-results/main-cleanup-service-01`, `make test-run-status
RESULTS_DIR=.cartulary/test-results/main-cleanup-service-01 JSON=1`, and
`make explain-run RESULTS_DIR=.cartulary/test-results/main-cleanup-service-01
DETAIL=performance|resources` in human and JSON forms (each DETAIL value run
separately). Outputs are retained in cleanup `.cartulary/retained-work/main-cleanup/`
as `service-status.json`, `service-performance.{json,txt}` and
`service-resources.{json,txt}`. JSON facts and retained validation pass; F-032 records
the human presentation omission rather than claiming complete field parity.

Source scope: early assembly runs record dirty predecessor commits. Service and
browser runs carry the final executable digest `sha256:c88d98ad80bb5a04d92046efa3faa746a367c9aeced20edef035e09e1e860fe6`;
finalization and broader verification use commit `229936209`. No manifest was
rewritten. The two cleanup failures above remain separate from historical
qualification failures. No baseline publication, external export or qualification
attempt ran. Full warm `check` is skipped for this cleanup; S-032 still requires it.

`disposition.json` and `main-transfer.json` in the same retained-work directory
record file hashes, exact removals and the checked fast-forward. Preservation
metadata under `/home/jochi/cartulary-instrumentation-archive/20261011-main-cleanup`
contains no retained new source tar/patch/index copy; these were removed when the
user said not to archive old code. Existing earlier archives and evidence were
left untouched. Git history and the unchanged historical checkouts retain source.
Final primary-checkout checks (from `/home/jochi/code/cartulary`):

- `env CARTULARY_TEST_RUN_ID=main-cleanup-main-shell-01 make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.current_epoch_evidence HARNESS_DIAGNOSTICS=basic`:
  PASS, one graph unit, 17,979 ms; manifest records commit `229936209` and clean
  executable state. Exact run: `.cartulary/test-results/main-cleanup-main-shell-01`.
- `make harness-observability-check RESULTS_DIR=.cartulary/test-results/main-cleanup-main-shell-01`:
  PASS; complete diagnostics/execution and read-only validation.
- `make test-run-status RESULTS_DIR=.cartulary/test-results/main-cleanup-main-shell-01 JSON=1`;
  `make explain-run RESULTS_DIR=.cartulary/test-results/main-cleanup-main-shell-01 DETAIL=performance`
  and `DETAIL=resources`, each with and without `JSON=1`: PASS. Outputs retained
  in primary `.cartulary/retained-work/main-cleanup/`. JSON models agree; human
  canonical summary values reconcile independently. F-032 remains the disclosed
  scope-specific presentation gap.
- `env CARTULARY_TEST_RUN_ID=main-cleanup-markdown-01 make lint-markdown`: PASS
  on the first complete plan draft.
  `env CARTULARY_TEST_RUN_ID=main-cleanup-markdown-02 make lint-markdown`: PASS
  on the reconciled final plan, including both source identities.
- `git diff --check`, exact tree comparison and initial primary content/index
  checks: PASS. Main transferred by a checked fast-forward after restoring only
  enumerated committed paths and removing the 30 verified experimental files.
  The deployment marker retained its original bytes. No push occurred.

**Source identity clarification.** The primary run's digest is
`sha256:4ad3142d6eb8e78cb6729390df615fc7570c832c3eda62df0a7ac332fd625d31`;
the clean integration checkout's digest is
`sha256:c88d98ad80bb5a04d92046efa3faa746a367c9aeced20edef035e09e1e860fe6`.
An independent reconstruction reconciles both manifests across 6,426 executable
entries with zero content/kind/path differences. There are 277 pre-existing
filesystem-mode differences (for example private visual fixtures and read-only
vendored inputs); source-snapshot v2 intentionally includes mode bits beyond
Git's executable flag. They were preserved rather than normalized. The detailed
comparison is primary `.cartulary/retained-work/main-cleanup/source-identity-comparison.json`.
Functional results are identified by checkout and digest; neither identity is
qualified and their manifests must not be relabeled or combined into a cohort.
A future freeze must explicitly select and verify its checkout identity.

The final checkpoint is documentation only. Broad code checks were already run
after finalization on the exact committed executable content; the fresh primary
shell/readers check validates operation with the preserved primary modes. The
next phase has not been implemented. S-033 closes cleanup only; qualification,
S-031/S-032 migration acceptance and the original overall completion remain open.


### Salvage evidence — historical off-main salvage

| Evidence | Source / identity | Result / limitations |
| --- | --- | --- |
| E-029 | Durable archive `/home/jochi/cartulary-instrumentation-archive/20261011T182827Z`; frozen source `c68c8cfa45e0a3e365bb168008e99c40e6ff5f1f`; primary baseline in `source-manifest.json` | V-024 PASS: archived content and restoration checked; 406 changed paths assigned; no original evidence removed. This is preservation evidence, not executable validation or qualification |
| E-030 | S-031 planning review at integration HEAD `79aecc1cae7a9474c92f4a9b8f77092bc4363662` plus unfinished working changes; exact reads of native engine/process/fixture tests, pending simulation/wrapper fixture, compiler/index/target-plan, test routes and owner REQ-852–855 | Direct caller search finds test-only JS simulation; duplicated full-suite smoke dispatch and copied legacy coordinator are pending design defects. Existing S-030 results are historical; no current-source runtime pass from this planning review. Temporary preservation snapshot: `/tmp/cartulary-s031-clean-cutover-planning-baseline.json`; durable implementation archive remains E-029 and S-030 checkpoint |


### Prior iteration — historical scope and evidence


### Historical recovery scope and authority

The current invocation is a tracker-only planning reset. S-014–S-019 remain
DONE. S-020 is IN_PROGRESS after the candidate-16 rejection; its functional
repairs and unsuccessful experiments remain preserved. S-021 has no qualified
revision and S-022 has not begun. Section 7 now defines measurement validation,
identical-treatment controls and causally selected repairs before another freeze.
The latest CPU request is mean <2.5% per cell; adoption and machine projection
remain pending. All other limits stay unchanged. RAM is for wall time. No new
runtime evidence, production change, test pass or implementation completion is
claimed by this planning session. E-026–E-028 record its scope and evidence.

### Planning-session scope and authority — historical

| Session / baseline | Scope / result | Next entry |
| --- | --- | --- |
| Coverage/qualification planning and tracker-only update; clean `main` at `1d423c6e5f9cd204f8210fa9958c780fd933d293` | User selected all remaining T-007/T-008 work and default-off after qualification; no production implementation authorized | Revalidate baseline/evidence, then enter S-014 only in a separately authorized task |

### Planning baseline — backend and frontend boundaries

| Boundary | Inspected / changed | Future gate |
| --- | --- | --- |
| Backend test support | Suite-service clock/activity/private lease and Docker-client seams inspected; no Go source changed | S-017/S-018 owner-routed fixtures and service coverage |
| Browser/wrapper support | Startup/clock, managed suite, wrapper/lock and scan seams inspected; no runtime source changed | S-018/S-019 lifecycle and signal/exit tests |
| Product backend/frontend | Not applicable: no product behavior, SQL, component, selector or golden move | No product or visual pass inferred |

### Planning baseline — contracts and generation

| Observation | Action / next gate |
| --- | --- |
| Adopted C-001–C-009; proposed C-010–C-014 and explicit format cutovers | Only tracker changes now; S-014 adoption precedes typed inputs and Make generation |
| Canonical and failure-envelope contracts remain stable | Incompatible optional diagnostics cut over atomically; no legacy translator or alias |

### Planning baseline — tests and harness

| Evidence | Result / limit | Next gate |
| --- | --- | --- |
| E-020–E-024 source/fixture/routing reads | No suite, generator, finalizer, benchmark or current-source runtime pass | S-015 executes characterization; S-021 requires clean frozen qualification evidence |
| V-023 / E-025 editorial and preservation checks | Passed; exact checkpoint below distinguishes the corrected editorial-check attempt from repository verification | T-019 DONE; future execution tasks remain TODO |

### Planning baseline — security and authorization

| Boundary | Required disposition |
| --- | --- |
| Session writes | Only this admitted regular tracker; no staging, source/spec/schema/test/dependency edit or service mutation |
| Future receipt | Existing pre-scan prohibition stays authoritative until S-014 adopts the narrow exception; no sampler restart or scan bypass |
| Optional observation | Loss stays independent of primary result; required privacy/cleanup failures retain their owner treatment |

### Planning baseline — open risks and next session

| Risk / blocker | Impact / required next evidence |
| --- | --- |
| RB-004 proposed owner semantics | S-014 must settle exact formats, bounds, qualification statistics and receipt exception before producers change |
| RB-005 characterization absent at completed future source | S-015 must execute independent baseline and save exact caller/row/treatment roster |
| RB-006 qualification not established | Existing historical results cannot close new total-cost/reader budgets; source/profile/uncertainty and every treatment cell must qualify |
| Latest requested mean CPU below 2.5% and RSS upper 32 MiB budgets | S-020/A must adopt/project the new CPU version; all other gates and fresh source-bound qualification remain required. Historical 2% evidence is unchanged |
| Optional Docker/helper failures and clock discontinuity | Explicit unavailable/partial observations; functional diagnostics alone never establish qualification |

### Planning baseline — evidence index

| Evidence ID | Source / inspected command or seam | Baseline / observation | Applicability / next gate |
| --- | --- | --- | --- |
| E-019 | `git --no-optional-locks rev-parse --show-toplevel`, `branch --show-current`, `rev-parse HEAD`, `status --short`, `ls-files --cached --others --exclude-standard -z`; SHA-256/lstat snapshot; root instructions/framework/owners | Clean `1d423c6e5`; 30 target entries; 6,608 repository entries; admitted tracker identity and proposed scope confirmed | Same-source planning and update, not runtime verification |
| E-020 | Observability facade/checker, resource loader/projection/accumulator, execution reader/analysis, performance analysis/presentation, export consumer boundary and selected fixtures | Duplicate resource traversal and missing checker envelope validation observed directly; existing secure closure and bounded views identified | F-018 / S-016; no failing runtime test claimed |
| E-021 | Runtime boot/local clocks and private execution observations; Go suite-service activity/lifecycle clocks; fixture providers/controller; testservices private lease/client; resource adapter/schemas | Identified clocks have no anchors; exact private container proof exists; retained sample scopes exclude containers | F-014/F-015 / S-017/S-018; lifecycle authority stays with owners |
| E-022 | Workspace admission/child wrapper; graph runner/scheduler; diagnostic session and envelope schema | Wrapper execs managed child; graph session drains/publishes before scan; current envelope excludes wrapper/imports/scan/return | F-016 / S-019; proposed exception requires adoption |
| E-023 | Worker/controller/collection engine/store; observation gate, host-admission/lease-lock; selected deterministic and real-worker tests | Worker CPU starts after imports; gate CPU explicitly unobserved; parent/index publication not fully covered by self-counters | F-017 / S-020; tests inspected, not rerun |
| E-024 | Task-surface owner, owner/test-family inputs, generated/schema ownership, canonical performance CLI/eligibility/statistics/tests, run comparison, Markdown wrapper/config | Existing public routes and clean-window rules; typed instrumentation manifest is future work; lint writes outside tracker and default globs omit it | V-014–V-022 future commands; no new Make run |
| E-025 | Pre-write identity/content guard; post-write manual Markdown/dependency/history review; Git staged/unstaged whitespace and repository snapshot comparison | Passed: 12 sections, 30 inventory entries, stable IDs, acyclic serial prerequisites, verbatim execution ledgers, raw index and 6,607 other repository entries unchanged | V-023 / T-019 DONE; no historical run promoted to current evidence |

Temporary comparison material is outside the repository at
`/tmp/cartulary-instrumentation-tracker-zkqdnktb/{baseline.json,original.md}`.
It records all 6,608 tracked/non-ignored entries and raw index identity; ignored
output roots have metadata checks only. No command capable of writing those
ignored outputs was run. Do not depend on these temporary files at future
implementation entry; revalidate the repository and affected source then.

### Historical systematic recovery evidence index

| Evidence ID | Exact source / scope | Observation | Limits / next action |
| --- | --- | --- | --- |
| E-026 | Isolated checkout `.cartulary/test-results/instrumentation-s021-qualification-16/instrumentation-s021-qualification-16/qualification/{qualification.json,results/*.json}`; commit `42831e37e9be710b02bb646152efcecc1c9dac23` | Read all 240 bound result files, checked each SHA-256 against its manifest, then independently decomposed measured RSS bounds and normalized reported support CPU using Python standard-library arithmetic. Section 7 records results | Original 2% CPU policy and FAILED disposition retained. No new run or validator execution; result-digest checking does not re-audit every archived workload file |
| E-027 | Current native meter `main_linux.go` / `proc_linux.go`, qualification acquisition/measurement/analysis/validator/policy, associated tests, graph-runner seams and owner/routing inputs | Raw sample arrays are summarized without a retained trace; support wakes at 2 ms, memory cadence is 100 ms, parent readiness check is 25 ms; off/0 reference is self-comparison; p90 is maximum for n=5/6. Compiler remains reachable through cache callback | Sampling race/bounds, support interference, GC retention and phase weighting are hypotheses to test, not newly proven implementation defects. Existing lifecycle tests constrain future movement |
| E-028 | `/tmp/cartulary-systematic-plan-ad37xeyh.json`; read-only Git/content/lstat comparison and tracker guard | Starting `main` / `1d423c6e5f9cd204f8210fa9958c780fd933d293`, 6,878 tracked/non-ignored entries; dirty implementation and staged tracker recorded. Entry tracker SHA-256 `ac56d5cd844ba8f7783ffc5b5af882fa0b09d74b6d6772e1d07c4804633c7ee1` | Final comparison recorded in the reset checkpoint. Baseline material is temporary; next entry must revalidate. Ignored caches/results were only read where named, not recursively snapshotted |

`make lint-markdown` was inspected but not run during this planning invocation:
its Make prerequisites can prepare dependencies and its wrapper writes a cache
stamp and command evidence outside the admitted tracker. Its default globs still
omit this tracker. Source edits, generators, tests, profiling, qualification and
`agent-finalize` were not run. Manual Markdown/traceability review and read-only
Git whitespace/content checks are the planning validation; they are not a lint or
runtime pass.

### Planning-session document completion checkpoint — historical

| Check / exact command or method | Result / evidence limit |
| --- | --- |
| `git --no-optional-locks diff --check` and `git --no-optional-locks diff --cached --check` | Passed; no staged changes and no whitespace errors |
| Manual current-scope, contract/finding/slice/validation, readiness and link review; temporary editorial assertions | Passed: twelve ordered sections, all stable ID ranges, 30 unique target paths, consistent table widths and nine serial workstreams |
| First temporary preservation/dependency assertion attempt | Failed after preservation checks because its regular expression omitted workflow labels following slice IDs and found zero rows; corrected the editorial helper, not repository source |
| `python3 /tmp/cartulary-instrumentation-tracker-zkqdnktb/verify_tracker.py` | Passed after correcting that assertion: all 6,607 other tracked/non-ignored entries retain content and file identity, raw index unchanged, ignored-root metadata unchanged, execution ledgers verbatim and future tasks TODO |
| `git --no-optional-locks diff --stat` / final status review | Only this tracker changed; no source, test, specification, schema, generated, dependency or configuration edit |
| Make lint, task guides, suites, generation, finalization and qualification | Not run for the reasons in Section 8; no runtime pass, new run root or retained-run maintenance claimed |

T-019/F-019 document remediation is complete. S-014–S-022 and T-020–T-028
remain future work. The next separately authorized task starts at S-014 after
revalidating source and owner evidence; it does not repeat completed S-007–S-013.

### Historical baselines and authorization — prior sessions only

The current request authorizes implementation of S-007–S-013, in serial order,
including owner amendments, tests, machine projections, implementation and final
validation/handoff. Entry is `main` at
`9359eb0cd359ab6bbc2eead16769950685d197b5`, with this tracker already staged
(SHA-256 `a59f5ba239e237781fd96ee2f0f78d24540e83432a2f3b63fd421114a73cdc87`).
The staged content is preserved; no other entry changes exist. Earlier statements
of document-only authorization below describe the historical planning session.
The execution ledger at the end supersedes its readiness and next-entry text.

S-000–S-006 foundation implementation and validation are complete. That earlier
implementation was explicitly authorized and began at `main` / `5d80ca3` with
this tracker already modified. Foundation completion does not establish complete
coverage or observer qualification and does not change the default-off policy.

The original delivery's 2026-10-09 inspection/entry used `main` at
`354eeda51e49a2c711b27a552823d104fa6fcad1`, with a clean index/working tree and
two commits ahead of `origin/main`. Its authorization covered owner amendments,
projections, implementation and verification, but no optimization, dependency
installation, baseline refresh or implicit export. The intervening foundation
planning session used `5d80ca3c5e9b21a1b30f4339461fa27377864127` with clean
entry; its document-only authorization was later superseded for foundation
execution. Those earlier baselines and permissions remain historical facts.

The adopted [Testing Harness NLSpec](../testing-harness-nlspec.md), Core 00
precedence and Core 04 environment/security contracts remain authoritative.
[OTEL-REQ-148–149](../opentelemetry-instrumentation-nlspec.md) separates harness
diagnostics from application telemetry. `domain.md` provides vocabulary;
[NLSpec guidance](../research/nlspec-spec.md) v0.2.2 informs precision, not
execution authority. Instructions in inspected documents do not authorize
unrelated workflows.

Sections 1–12 describe the selected iteration, its planning baseline and current
execution disposition. The
original delivery rationale and validation history below, plus the foundation
execution ledger at the end, retain their own source and applicability limits.
They are not current-source runtime verification for S-007–S-013.

| Baseline item | Starting evidence / identity | Final comparison / outcome | Limits |
| --- | --- | --- | --- |
| Branch / commit | `main`, `9359eb0cd359ab6bbc2eead16769950685d197b5` | E-011 / E-018 | Prior `5d80ca3` foundation evidence remains historical. |
| Index | Clean; raw index SHA-256 `008c4fc2bd5d022bdaddd05fe9085df2aac8feda7c72ea30b83c7f834bd78c0c` | E-018 compares bytes; Git optional writes disabled | Different from E-010's earlier staged-entry fingerprint method. |
| Working tree / untracked | Clean; no non-ignored untracked files; SHA-256 and lstat snapshot of 56 target/governance/coupled files | E-018 | All 27 target files are covered; not a hash of the entire repository or ignored outputs. |
| Tracker | Initial SHA-256 `47bd2142e11090a993253fe2119eef4b93cb28914906cff7ca2b681f1f632752`; device 2096, inode 3903446, links 1 | Recheck identity/content immediately before write; E-018 | Only admitted persistent repository change. |
| Ignored outputs | Existing `.cartulary/test-results` and retained-work evidence | Existing ledgers preserved; no new run created | Historical passes are not current-source verification. |

### Foundation planning handoff — historical

The following prior-session rows and E-001–E-010 are preserved under their
`5d80ca3` baseline. Their then-future next actions were superseded by S-000–S-006
execution. They neither describe current source nor authorize S-007–S-013.

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


### Historical identity-planning scope and authority

| Time / session | Baseline and state | Files / commands / evidence | Supersession / next action |
| --- | --- | --- | --- |
| 2026-10-09 read-only plan, followed by tracker update | `main` at `9359eb0cd359ab6bbc2eead16769950685d197b5`; clean entry | E-011–E-018; root instructions, framework, adopted owners and NLSpec reference | User selected identity/allocation correlation plus owner-local timing. This update changes the handoff only; S-007 is the next separately authorized task |

### Historical identity-planning backend and frontend boundaries

| Boundary | Observation / files | Current action / future gate |
| --- | --- | --- |
| Backend test support | `internal/testutil/pgtest/pgtest.go`; suite-service lifecycle/diagnostic producers; E-014 | Inspected measurement seams; no Go change. S-011 must use their owner-routed tests |
| Product backend / frontend | No product behavior, SQL migration, web component, selector or assertion move | Not applicable to implementation scope; no broad product test or visual pass inferred |
| Browser harness | Group runner/report adapter, startup and reset producers; E-012/E-014 | Selected future harness identity/timing work; preserve browser resource cleanup and public outcomes |

### Historical identity-planning contracts and generation

| Baseline | Current evidence | Files / action | Next gate |
| --- | --- | --- | --- |
| `9359eb0` | Adopted REQ-830–840 and private failure/fixture contracts; proposed C-006–C-009 | Inspected definitions/policy, attachments, resource schema and generator ownership; edited no owner or machine contract | S-007 declares projections and semantics; later generated changes use existing Make transaction and V-007 |

### Historical identity-planning tests and harness

| Baseline | Current evidence | Files / action | Next gate |
| --- | --- | --- | --- |
| `9359eb0` | E-012–E-016; assertions/scenarios inspected, no test execution | Foundation and shared-lease fixtures, command/routing metadata; historical passes retain source limits | S-008 revalidates exact consumers, resolves affected owner rows and executes independent characterization before movement |
| `9359eb0` | V-013 / E-018 | Read-only Git, content/identity snapshot, manual Markdown and link/dependency review | Complete document checkpoint; no artifact-producing Make command used |

### Historical identity-planning security and authorization

| Baseline | Current evidence | Files / action | Next gate |
| --- | --- | --- | --- |
| `9359eb0` | C-002–C-004/C-006–C-009; E-011/E-015 | Only the admitted regular tracker may change; no export, service mutation, raw counters or credentials copied | Preserve exact identity checks, secure bounded reads, required cleanup and pre-scan producer quiescence |
| `9359eb0` | Explicit document-only authorization | No production source, test, schema, generator or dependency edit | Implementation authorization not granted by this planning invocation; later task begins with S-007 |

### Planning-session open risks and next entry — historical

| Baseline | State / evidence | Completed action | Next gate / restart |
| --- | --- | --- | --- |
| `9359eb0` | F-010–F-012 planned; RB-002/RB-003 block dependent implementation | Defined contracts, slice order, validation, compatibility and residual scope in this tracker | Recheck `git --no-optional-locks status --short` and `git --no-optional-locks rev-parse HEAD`; inspect changed evidence; enter S-007 only in a separately authorized task |
| `9359eb0` | F-013 / T-011 DONE | Replaced obsolete current foundation plan and refreshed 27-file inventory; preserved historical ledger | E-018 confirms preservation and document completion; next work is separately authorized S-007 |

### Planning-session evidence index — historical

Paths are repository-relative. Source inspection below is current at `9359eb0`;
historical runtime results are retained in their original ledgers. E-018 is the
document checkpoint, not a runtime-validation result.

| Evidence | Exact source / command | Observation / result | Applicability / limit |
| --- | --- | --- | --- |
| E-011 | `git --no-optional-locks status --short`, `rev-parse HEAD`, `branch --show-current`, `ls-files`; in-memory SHA-256/lstat/index snapshot | Clean `main` at `9359eb0`; 27-file target and 56-file scoped baseline; admitted tracker identity/content recorded | This update's baseline; prior historical results are not current-source passes |
| E-012 | `runtime/command-failure.mjs:createCommandFailureContext`; `runtime/private-child-process.mjs:runPrivateCapturedProcess`; work-graph executor/runner; row-runner `execute`/`main`; browser group launch | Private failure identity is fresh but not a neutral retained launch context; parent PID/unit observation and packed row invocation seams identified | F-010/C-006; source observation, no failing runtime reproduction claimed |
| E-013 | `scheduler/fixture-broker/index.mjs:acquireOnce`; providers' browser process callback; collection engine `register`/`lease`; shared-lease test | Existing allocation reference derives from entry lease ID; correlation follows successful lease publication; shared allocation/null-unit behavior already tested | F-011/C-007; preserve existing ownership and reuse; tests inspected, not run |
| E-014 | `internal/testutil/pgtest/pgtest.go` preparation/template migration timing; suite-service lifecycle/diagnostics; browser reset route/attempt; row/browser report adaptation | Existing local migration/reset measurements should be reused; inspected migration events follow success; parsing boundaries lack the proposed execution-linked timing projection | F-012/C-008; new outcome/identity coverage is a proposed extension, not an invented owner violation |
| E-015 | Collection engine/controller/worker/session/store; resource reader/accumulator/presentation; performance analysis/formatting; comparison and check CLI; runtime gate; instrumentation definitions/policy/resource index | Foundation boundaries are present; retained resource data has process/unit/allocation relations but no complete invocation model; canonical/export/quietness protections remain | C-001–C-009; current architectural inspection, not current runtime qualification |
| E-016 | `tools/task_surface_owner.json` evidence/command/harness/Markdown recipes; `tools/harness/static-analysis/markdownlint.sh`; routed fixture names | Public contract routes exist; Markdown wrapper writes cache/stamp and retained output with readiness prerequisites | Supports future V-007–V-012 and present skip; exact additional owner rows resolved in S-008 |
| E-017 | User scope selections; NLSpec guidance v0.2.2; current tracker review | Selected identity/lifecycle plus owner-local timings; deferred cross-clock/full-wrapper/container/qualification scope; current-state drift identified | Planning rationale only; owner adoption and implementation remain future gates |
| E-018 | Pre-write identity/content recheck; final status/diff/index and 56-file comparison; historical-section comparison; `git --no-optional-locks diff --check` | PASS: only this tracker differs; all 55 other scoped file identities/contents, tracker identity, index, branch and commit unchanged; no untracked files; whitespace check passes | Original validation/foundation ledgers and prior finding/slice/validation/handoff tables preserved byte-for-byte; twelve sections, 27 inventory entries, table structure and serial dependencies reviewed. No suite, generator, finalizer or Markdown wrapper run |

## 11. Open Questions and Blockers

### Current lifecycle entry gates

S-036–S-042 are DONE at E-042–E-048. No functional blocker remains in this scope.
Other environments must check exact old-session proof before protocol cutover;
unresolved ownership blocks that environment until its running version drains or
recovers the resources. Historical failed attempts remain failed evidence, with
their repairs and subsequent passes recorded separately. F-022/RB-006 remains
explicitly unresolved; qualification is outside this iteration.


### Previous functional iteration — current blockers and stop decisions (historical)

| Scope | State | Required resolution |
| --- | --- | --- |
| F-022/RB-006 | Unresolved, qualification DEFERRED | Separate adopted acquisition and unchanged qualification gates; outside this effort |
| S-035 | DONE | V-036/E-032 saved checkpoint |
| S-034/S-031A–F/S-032 | DONE | Section 6 saved checkpoint gates and current-source V-033/E-040; no remaining functional blocker |

Implementation is authorized. No permission gate follows an ordinary failed
functional test; repair its owner and repeat affected validation. Do not restart
experimental collection, alter budgets, force GC or pad workloads.

### Salvage gates — historical off-main salvage

There is no unresolved planning choice: Go graph first, Node rows retained, all
added observations retained, shared Go fixture policy and deferred certification
were explicitly selected. Later slices are gated by their predecessors and fresh
characterization, not by permission requests. RB-006 remains an unresolved
qualification blocker and does not block functional salvage. A failed salvage
gate blocks its dependent slices until repaired; no automatic qualification retry.
The current S-031 candidate is incomplete: generated test expectations still name
the deleted JS dispatcher, native interop coverage is unfinished, and the pending
test-only simulation/copy must be removed as specified in Section 7. These are
implementation gates, not a reason to restart architectural research. S-032 is
blocked on V-032. F-031 closes only with the behavioral test/caller cutover.

### Prior iteration — historical scope and evidence


User preferences are resolved: all remaining work is selected and diagnostics
stay default-off after qualification. There is no unresolved primary-owner
contradiction. The following gates control implementation and completion. Qualification
has exposed F-022–F-027; all failed observations and repair checkpoints remain
retained below. F-022/RB-006 are open after candidate 16. Closed functional
repairs retain their own evidence and do not establish qualification.

| ID | Blocker / affected slices | Required authority or evidence | Status / resolution |
| --- | --- | --- | --- |
| RB-004 | C-010–C-014 adoption gate; S-015–S-021 | S-014 owner review, closed projection map, receipt exception, policy bounds and qualification formulas/acceptance | DONE; REQ-845–849 / AC-142–146 and S-014 checkpoint |
| RB-005 | Current-source pre-move characterization absent; S-016–S-020 | S-015 executed V-015 baseline, exact wrapper/helper/caller and owner-row inventory, independent future cases and typed workload/reader protocol | DONE; executed S-015 baseline and correction obligations below |
| RB-006 | Completed-source total observer/reader qualification absent; S-021/S-022 | S-020 complete measurement contract, clean frozen source/profile, eligible stable off/basic × 0/1/4 matrix with uncertainty and all budgets passing | IN_PROGRESS; candidate 16 failed 14 numeric cells and all five stability assessments under its original policy; S-020/A–E gate the next full matrix. All earlier failures/interruption remain retained |

S-014–S-019 and prior S-020 repairs passed their recorded functional gates.
S-020 is reopened for the systematic recovery plan, with measurement and
repeatability questions unresolved. All earlier frozen commits and failed evidence remain retained;
the isolated checkout advances only after saved repair checkpoints. Do not start
S-022 until frozen executable source passes all required S-021 cells.
Pre-existing staged tracker content is preserved.

### Closed historical blockers

| ID | Blocker / affected slices | Needed authority or evidence | Status / resolution |
| --- | --- | --- | --- |
| RB-001 | Historical missing foundation characterization, S-002–S-005 | S-001 baseline plus S-004/S-005 fixtures and final S-006 results | DONE; closed by recorded foundation execution, not reopened |
| RB-002 | New identity/activity/projection semantics not adopted; S-008–S-012 | S-007 owner amendment and reviewed C-006–C-009 inputs/outputs/defaults/bounds/acceptance mappings | DONE; S-007 adopted REQ-841–844 / AC-138–141 before projection changes |
| RB-003 | Independent pre-move characterization not executed for new launch/correlation/timing boundaries; S-009–S-012 | S-008 current-source V-008–V-011 baseline results and exact consumer/test routing inventory; explicitly assigned correction obligations | DONE; S-008 executed characterization and saved caller/routing inventory before movement |

## 12. Binary Completion Criteria

### Current lifecycle completion boundary

| Outcome | Required / current state |
| --- | --- |
| Planning | Complete for selected lifecycle scope; no whole-repository behavior audit claimed |
| Authorization | Current user explicitly authorizes S-036–S-042 implementation |
| Readiness | S-036–S-042 DONE with individual E-042–E-048 checkpoints |
| Prior functional scope | S-035/S-034/S-031A–F/S-032 remain DONE at historical sources |
| Current completion | Functional remediation complete; S-042/V-043/E-048: one native shared owner, direct fixture coordination, no obsolete policy/oracle/wiring; final full warm check 982/982 |
| Handoff | E-042–E-048 record changed files, commands, process/harness outcomes, source identity, roots, failures, skips and compatibility effects; final checkpoint is documentation only |
| Qualification | DEFERRED; F-022/RB-006 and S-020–S-022 unchanged, 44.952/90 minutes, no RSS/CPU/default-on/Core 05 claim |
| Original effort status | IN_PROGRESS solely for deferred qualification; this does not reopen the completed functional lifecycle scope |


### Previous functional iteration — current completion boundary (historical)

| Outcome | Required / current state |
| --- | --- |
| Planning | Complete; user explicitly requested execution |
| Implementation | S-035/S-034/S-031A–F/S-032 DONE; one native graph and shared fixture policy |
| Functional completion | Complete; S-035, S-034, S-031/A–F and S-032/V-033 individually checkpointed |
| Evidence | E-040 final source, commands, owning exits, exact roots and audits; failed attempts and candidate content preserved |
| Qualification | DEFERRED; F-022/RB-006 unresolved, unchanged budgets and 44.952/90 minutes |
| Original effort | IN_PROGRESS solely for deferred qualification; functional delivery does not imply overhead certification |

V-033 is complete. Finalization preceded broader verification; both executable
repairs were re-frozen and revalidated. Ordinary readers stay local and diagnostics
default-off. No overhead certification, improved RSS/CPU, default-on readiness or
Core 05 publication is claimed.

### Salvage completion — historical off-main salvage

S-023–S-030 are complete. S-031–S-032 remain uncompleted. Functional salvage
requires V-025–V-033 and the S-032 handoff; it does not qualify instrumentation.
Qualification is DEFERRED, budgets unchanged, F-022/RB-006 unresolved, S-020–S-022
DEFERRED and the overall original effort IN_PROGRESS. Implementation authorization
for the existing salvage scope came from the explicit request of 2026-10-11. The
latest planning request adds no execution in this checkpoint; no code was changed
after that steering. Primary staged work remains preserved. The cutover direction
and ordered gates are complete as a plan; per-test retirement accounting and
current-source acceptance remain implementation prerequisites, not claimed passes.

### Prior iteration — historical scope and evidence


| Outcome | Current statement / required evidence |
| --- | --- |
| Planning completeness | complete for the bounded recovery decision sequence in Section 7; E-026/E-027 establish its entry evidence. Root cause and any specific architectural repair remain hypotheses requiring B–D. This is not an exhaustive new subsystem audit |
| Per-slice readiness | S-014–S-019 DONE; S-020 IN_PROGRESS with A ready for resumption and B–E gated as specified; S-021 new collection blocked on E; S-022 blocked on passing S-021/RB-006 |
| Implementation authorization | not granted by this planning invocation. Earlier S-014–S-022 authorization is retained as context; the latest request calls for a plan, so this session performs no implementation |
| Prior delivery completion | S-000–S-013 remains DONE at its historical sources, with failed attempts and ledgers preserved |
| Session write-boundary verification | E-028 records the current tracker-only comparison; E-025 remains historical. Pre-existing implementation changes and staged tracker must remain unchanged |
| Document completion | T-019 planning checkpoint remains DONE. Current-source V-023 and final handoff completion remain pending S-022 |
| Selected future completion | S-014–S-022/T-020–T-028 and T-007/T-008 become DONE only after S-022 current-source verification and final saved checkpoint |
| Overall status/default | IN_PROGRESS now; completion requires all selected coverage/qualification gates. `HARNESS_DIAGNOSTICS=off` remains the default even on success; T-009 remains DROPPED |

The final implementation handoff must identify changed files by workstream,
adopted owner contracts, schema migration effects, exact commands/results and
run roots, failed attempts and repairs, skipped checks with reasons, final source
commit/state/digest, remaining physical coverage/uncertainty limits, and the
separate disposition of functionality and engineering qualification. No baseline
refresh or Core 05 claim follows. If a required qualification gate remains failed,
record the finding and keep overall work IN_PROGRESS; do not report completion.

The execution ledgers below are preserved verbatim. Their “current” or “next”
statements describe the original sessions and do not supersede Sections 1–12.

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

## Planning next entry — historical identity and lifecycle handoff

The document update selects S-007–S-013; it does not start their execution.
The next separately authorized task begins with S-007 owner reconciliation,
using the prerequisite table in Section 6. Revalidate the source baseline and
affected evidence before editing. Resolve routing at implementation entry,
characterize before movement, and update this tracker after every completed
workstream before starting the next. S-013 final validation/handoff is required.

Preserve exact-run selection, same-source comparison, performance-gate exit 13
and pre-scan quiescence. Container counters, full wrapper/post-scan coverage,
cross-clock joins and total observer/0/1/4-reader qualification remain deferred.
`HARNESS_DIAGNOSTICS=off` stays the default. E-018 records this document update's
verification; the historical S-006 ledger is not evidence that this iteration ran.

## Execution identity iteration — 2026-10-09

This ledger controls S-007–S-013 execution. Each completed checkpoint is saved
before its successor starts. Public Make commands use fresh `identity-*` run IDs,
the installed pinned runtime, and `CARTULARY_PREPARATION_POLICY=installed_only`.
Results remain under `.cartulary/test-results/`; no retained run is overwritten.

| Slice | Status | Changes, validation and successor gate |
| --- | --- | --- |
| S-007 | DONE | REQ-841–844 / AC-138–141 adopted; versioned cutover and owner boundaries reconciled. `make lint-markdown`, `identity-s007-markdown-01`, exit 0. V-007 owner review complete; RB-002 closed. S-008 ready. |
| S-008 | DONE | Added failure/acquisition/reuse characterization; `identity-s008-evidence-01` and `identity-s008-command-01` both exit 0. Existing routed baseline passes; RB-003 closed for movement. S-009 ready. |
| S-009 | DONE | Neutral context, private single-writer slots, execution-index reader/session, and failure/capture caller migration. `identity-s009-command-01`, evidence-03, generate-03, json-02, policy-01, drift-01 all exit 0. S-010 ready. |
| S-010 | DONE | Independent allocation attempts, successful leases, immutable process relationships; resource-index v2. Evidence-02, command-01, browser-01, generate-01, shapes-01 (JSON only), policy-01 and drift-01 pass. S-011 ready. |
| S-011 | DONE | Actual initialization/apply measurements, reset-attempt v2, parsing activities, typed activity vocabulary and Go routing. V-010, owner, service-backed and projection gates pass; S-012 ready. |
| S-012 | DONE | Secure execution/lease/process/clock closure, bounded explanation v2, corruption/cap/late-write/security and export-isolation fixtures. V-011 passes; S-013 ready. |
| S-013 | DONE | Finalize-04 precedes contract-03 and final current-source gates. Fresh Go/Vitest/shell/service/browser/reset evidence, generated/import review and documentation reconciliation pass; final handoff below. |

Entry discovery: `make task-guide ROLE=module-author OWNER=<owner>` passed for
`harness.evidence_accounting`, `harness.command_surface`, `harness.browser`,
`platform.postgres`, and `module.database_migrations`. Source and staged tracker
match the planning baseline; staged and unstaged whitespace checks passed.

### S-008 consumer and correction checkpoint

Neutral identity callers: work-graph executor; row Go/Vitest/shell adapter;
private child capture and its wrapped-tool, snapshot-builder and Playwright
callers; browser acquisition and review launchers. Command-failure tests and
Vitest diagnostic fixtures are direct construction consumers. Framework workers
remain inferred; standalone review/wrapper callers migrate identity without
claiming graph-retained coverage. Resource-only observations cannot manufacture
invocations. Go suite-service activities use explicit inherited test context.

Existing evidence-contract imports execute instrumentation, collection-engine and
foundation characterization tests; command-surface imports execute private
capture, classified failure, frontend graph and lifecycle fixtures. Added baseline
cases establish no lease on failed acquisition, shared allocation/distinct leases/
one cleanup, and independent repeated private invocation identities. Expected new
cases cover explicit causal attempts, partial terminal outcomes, conflicts,
nullable local clocks and retained reference closure in their implementing slices.

Owner routes include `platform.postgres.integration.fresh_fixture_admission`,
`platform.postgres.support_unit.certificate_fixture_lifecycle`,
`module.database_migrations.unit.test_harness_targeted_operation_validation`, and
`harness.browser.unit.testservices_lifecycle_contract`. Suite-service package unit
coverage is also part of the authored backend-unit execution topology. Additional
new tests must be included in their owner selectors before claiming coverage.

Correction obligations: collector registration currently overwrites attribution;
reset's helper falls back to UTC and clamps invalid elapsed values; migration
measurements include some preparation/finalization and record only success.
These receive explicit regression fixtures in S-010/S-011, not compatibility
assertions. Existing canonical accounting/failure/cleanup expectations remain.

### S-009 executed evidence

Commands: `make harness-command-surface-contract`, `make harness-evidence-contract`,
`make generate`, `make json-shape-check`, `make generated-artifact-policy-check`,
`make generate-drift`, using run IDs in the slice row and
`GO=/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.2.linux-amd64/bin/go`.
Generated standalone validators include the new launch/index and shared record
schemas; the policy moved to v2 with its independent 4,096-record bound.

Failed attempts retained: generate-01 selected Go 1.27.1 under installed-only and
failed OpenAPI generation; explicit installed Go 1.27.2 fixed it (generate-02/03).
Evidence-01 exposed Buffer/text handling, evidence-02 exposed missing explicit
atomic replacement for mutable private records; evidence-03 passes after repairs.
JSON-01 rejected the record schema's public classification; it is now correctly
registered as a shared component, and JSON-02 passes. Two invalid explain-run
DETAIL invocations failed usage validation; the supported summary/log selection
then diagnosed the exact JSON failure. No baseline or retained evidence was reset.

### S-010 executed evidence

V-009 passes: failed acquisition, failed canonical lease publication, shared reuse,
borrowed replacement, idempotent cleanup, conflicting explicit registrations and
late-registration CPU intervals have independent assertions. Existing PID-proof,
worker-death, sibling-quietness and browser acquisition/interruption fixtures pass.
Resource-index v2 removes mutable attribution and leases; execution-index v1 owns
those relationships. Physical consumption remains contextual for shared/borrowed
allocations. No lifecycle authority moved into diagnostics.

Commands and exact run suffixes (prefix `identity-s010-`):
`make harness-evidence-contract` (evidence-01/02),
`make harness-command-surface-contract` (command-01),
`make harness-ui-review-lifecycle` (browser-01), `make generate` (generate-01),
`make json-shape-check` (shapes-01), `make generated-artifact-policy-check`
(policy-01), and `make generate-drift` (drift-01), all target exits 0.
The first multi-target invocation exited 2 after JSON passed: the next target
correctly rejected reuse of its nonempty run root. Separate fresh identities
completed the remaining gates. No retained evidence was removed or overwritten.
Source is the entry commit plus this worktree's recorded implementation changes;
no commit was created. S-011 begins only after this saved checkpoint.

### S-011 executed evidence

Owner-local activities carry opaque clock identity, resolution, nullable duration,
explicit availability and incomplete/failed/cancelled outcomes. Initialization and
apply timers exclude source loading and template finalization; the same measured
duration feeds existing success events. The suite template producer records once.
Reset-attempt v2 and its consumers replace UTC fallback/zero clamping. Node report
reading/adaptation includes missing/malformed evidence. Canonical timing, phase,
reservation and cache accounting remain untouched and their independent fixtures
pass. Private slots now use bounded exclusive files shared by Node/Go producers.

Validation (run prefix `identity-s011-`): `make harness-evidence-contract`
(evidence-02), `make harness-command-surface-contract` (command-01), `make test-slice`
for activity routing (migration-01), PostgreSQL support (pg-01), migration harness
(apply-01), and browser suite-service lifecycle (services-01), all exit 0.
`make service-backed-test-slice OWNER=platform.postgres
ROWS=platform.postgres.integration.fresh_fixture_admission HARNESS_DIAGNOSTICS=basic`
(integration-01) passed 3/3 units, with one shared Go template migration, one parsing
activity, allocation/lease facts and complete retained execution closure.
`make harness-observability-check RESULTS_DIR=<exact root>` passed for migration-01
and integration-01. These establish function, not performance qualification.

`make generate` (generate-03), `make generate-drift` (drift-02),
`make generated-artifact-policy-check` (policy-01), `make json-shape-check`
(json-01), `make lint-scripts` (scripts-01), `make format-go` (format-03) passed.
Failures preserved: generate-01 rejected the new Go route's copied shell-only
verification IDs; corrected typed verification routing and generate-02 passed.
Evidence-01 selected an activity where its fixture intended a launch; explicit
kind selection fixes it. Drift-01 detected the stale generated topology index
after subsequent producer edits; generate-03 and drift-02 close it.

Inspection corrections: row adapters erase runner kind when constructing their
private invocation; pass it explicitly to retain Go/Vitest/shell producer identity.
The shared template's direct service-helper launch now has a neutral context,
which supplies its activity's parent identity without assigning shared work to
its first consumer. Standalone context validation uses its explicit standalone
run scope. These changes are named corrections, not compatibility preservation.

### S-012 executed evidence

The reader validates unique identities, launch disposition and parent cycles,
canonical units/rows and lease artifacts, allocation outcomes, activity clock
consistency, physical proof linkage and every declared omission before exposing
aggregates. Corrupt/capped beginnings remain explicit omissions. Resource CPU
attribution excludes conflicting registrations and pre-registration intervals.
Explanation v2 separates bounded execution facts from canonical accounting,
labels inferred/conflicted/allocation context, and preserves null local durations.
Old resource relationship/lease projections were removed; no translation aliases
or new public command were added. Existing OTLP remains isolated.

Additional inspection correction: the error path previously attempted the retained
scan before broker/runtime cleanup. It now drains producers before capture and
publication. Go activity channels now reuse suite-runtime root validation without
requiring an unrelated service-suite identity; private permissions, exact record
identity and shutdown checks precede completion writes. Service state retains its
existing suite requirement. Routed root-validation fixtures cover the extraction.

All commands below exited 0 (prefix `identity-s012-`):
`make harness-evidence-contract` (evidence-01/02/03),
`make harness-command-surface-contract` (command-01), `make generate`
(generate-01/02/03/04), `make generate-drift` (drift-01/02/03),
`make json-shape-check` (json-01/02), `make generated-artifact-policy-check`
(policy-01), `make format-go` (format-01/02), and the routed local activity/root
validation `make test-slice` (go-01/02, basic diagnostics). Incremental reruns
followed new clock-consistency and secure-channel assertions, not historical
pass reuse. `make harness-observability-check` and `make explain-run
DETAIL=performance` passed on exact integration-01 from S-011 and exercised the
new reader/presentation. Those reads validate retained producer evidence; S-013
still requires final-source execution. Staged and unstaged whitespace checks pass.


### S-013 source repairs and initial attempts

The first broad gate (`identity-s013-contract-01`, `make harness-contract`,
outer exit 2; normalized `harness/child_target_failure`) exposed stale launch
inheritance in nested fixture runs. Retained
`unit-logs/target-harness-contract/stdout.log` identifies private-capture wrapper,
frontend lifecycle, and command-failure cases with `launch context run mismatch`.
The repair builds the final child environment before identity allocation and
rejects stale/malformed parent inheritance while allocating a fresh context.
Direct context readers and failure-envelope scope validation remain strict.
The characterization now tests both rejection and fresh independent scope.

Final inspection also found nested private launches had identity without explicit
physical-process correlation. Spawn captures optional Linux boot/start/namespace
proof through the bounded private channel; capture forwards it to the collector
before shutdown. The shared runtime proof parser prevents divergent identity
algorithms. Raw process proof never enters retained execution records. Late
registration still cannot assign earlier lifetime consumption. Browser acquisition
callbacks now retain their launch identity alongside allocation identity.
A review-child callback failure after spawn is recorded as failed, not spawn_failed.
These are explicit S-013 source repairs, not validation-only edits.

Before these repairs, `identity-s013-finalize-01`, scripts-01, markdown-01,
go-01, vitest-01, integration-01 and browser-01 passed. The four representative
graph runs executed freshly with cache off and basic diagnostics (1/1, 2/2,
3/3 and 11/11 units respectively). Their evidence remains pre-repair.
`identity-s013-generate-01` failed because the new private schema used a filename
instead of the registered schema ID in its reference; corrected generation
`identity-s013-generate-02` passed. Evidence-01, command-01 and finalize-02 passed
on the process-proof repair before the stale-inheritance repair. Finalization
uses no RESULTS_DIR: no eligible successful full warm check was selected, so
retained-run maintenance was skipped.


The diagnostic reproduction `make harness-contract-tests`
(`identity-s013-contract-tests-01`, exit 2) started before the inheritance repair
and preserved the same private-wrapper mismatch in its retained stdout. It is
superseded only for acceptance by a fresh completed-source broad gate, not erased.
Generation-03 and evidence-02 passed after the inheritance repair;
finalize-03 passed before the final broader gates. The test-operations retained
analysis guidance now describes post-registration CPU intervals and contextual
shared work, matching explanation v2.


A final browser-boundary review found that syntactically valid but structurally
invalid reports could be labeled as successful parsing. The browser producer
now also consumes adapter artifact/accounting failures when settling the local
activity, matching the row runner. Existing malformed/missing/ambiguous Playwright
adapter fixtures provide the independent failure classification. This changes
only optional activity outcome; canonical row classification remains unchanged.


### S-013 completed-source verification and handoff

Final source baseline: `main` at
`9359eb0cd359ab6bbc2eead16769950685d197b5` plus the unstaged implementation;
source digest `sha256:536362e6cb6ad8106521bbb95d9626ebd426d4727518e89414302521d742bb8b`.
No commit, dependency installation, golden/baseline refresh or export was made.
The pre-existing staged tracker is preserved byte-for-byte:
`a59f5ba239e237781fd96ee2f0f78d24540e83432a2f3b63fd421114a73cdc87`.

All commands execute at the repository root. Go-requiring commands use
`GO=/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.2.linux-amd64/bin/go`
and `CARTULARY_PREPARATION_POLICY=installed_only`. Each table row sets
`CARTULARY_TEST_RUN_ID=identity-s013-<suffix>` independently; its exact retained
root is `.cartulary/test-results/identity-s013-<suffix>`. Representative runs use
`CARTULARY_HARNESS_CACHE_MODE=off` and command-line `HARNESS_DIAGNOSTICS=basic`.
The final command/result table is recorded below after terminal verification.

Compatibility cutover: resource index, performance explanation, instrumentation
policy and required browser-reset attempt advance to v2. Execution index and
neutral launch context start at v1; the private launch-proof support schema is
not a retained API. Current readers reject superseded shapes. No translator,
forwarding alias, new public command or default-on behavior was added. Canonical
unit events, failure envelope, lease lifecycle and public failure classification
retain their versions. Producers, schemas, minimum readers and generated outputs
moved together in the implementing slices; S-012 completes secure presentation.

Inspected authority/navigation includes `docs/domain.md` (unchanged), the adopted
Testing Harness owner, authored topology/catalog/routing, schema attachments,
generated-artifact policy and the owner-local migration/reset/runner consumers.
Generated review covers standalone validators, instrumentation descriptors/shared
schema definitions and the topology render index; authored owner inputs are their
sources. No Go module/lockfile, frontend dependency lockfile or SQL change occurs.
Executable validation reads typed machine inputs, never this tracker or Markdown.

Changed-file inventory for this iteration (the staged tracker remains separately
owned by the pre-existing planning session):

- `.agents/skills/cartulary-test-ops/references/result-analysis.md`
- `contracts/verification/owners/harness.evidence_accounting.json`
- `docs/guides/cartulary_implementation_testing_guide.md`
- `docs/handoffs/cartulary-test-harness-instrumentation.md`
- `docs/testing-harness-nlspec.md`
- `internal/testutil/pgtest/pgtest.go`
- `internal/testutil/suiteservices/activity.go`
- `internal/testutil/suiteservices/activity_test.go`
- `internal/testutil/suiteservices/env.go`
- `tools/execution_topology_render_index.json`
- `tools/harness/browser/browser-catalog-group-cli.mjs`
- `tools/harness/browser/browser-reset-attempt.mjs`
- `tools/harness/browser/lifecycle/reset-route.sh`
- `tools/harness/browser/review-child.mjs`
- `tools/harness/browser/tests/test-browser-acquisition.mjs`
- `tools/harness/browser/tests/test-browser-work-graph.mjs`
- `tools/harness/contract/generated/foundation-schema-validators.cjs`
- `tools/harness/contract/generated/instrumentation-definitions.mjs`
- `tools/harness/execution/run-make-node-tool-cli.mjs`
- `tools/harness/execution/runners/row-runner-cli.mjs`
- `tools/harness/execution/tests/test-vitest-diagnostics.mjs`
- `tools/harness/generated-artifacts/instrumentation-definitions.mjs`
- `tools/harness/generated-artifacts/instrumentation-shapes.mjs`
- `tools/harness/observability/collection-engine.mjs`
- `tools/harness/observability/diagnostic-session.mjs`
- `tools/harness/observability/execution-analysis.mjs`
- `tools/harness/observability/execution-reader.mjs`
- `tools/harness/observability/observability-check-cli.mjs`
- `tools/harness/observability/observability.mjs`
- `tools/harness/observability/performance-analysis.mjs`
- `tools/harness/observability/performance-presentation.mjs`
- `tools/harness/observability/resource-accumulator.mjs`
- `tools/harness/observability/resource-collector-worker.mjs`
- `tools/harness/observability/resource-collector.mjs`
- `tools/harness/observability/resource-linux.mjs`
- `tools/harness/observability/resource-presentation.mjs`
- `tools/harness/observability/resource-projection.mjs`
- `tools/harness/observability/resource-reader.mjs`
- `tools/harness/observability/tests/test-collection-engine.mjs`
- `tools/harness/observability/tests/test-execution-index.mjs`
- `tools/harness/observability/tests/test-foundation-characterization.mjs`
- `tools/harness/observability/tests/test-instrumentation.mjs`
- `tools/harness/performance-fixture/snapshot-builder-cli.mjs`
- `tools/harness/runtime/command-failure.mjs`
- `tools/harness/runtime/execution-observations.mjs`
- `tools/harness/runtime/launch-context.mjs`
- `tools/harness/runtime/local-activity.mjs`
- `tools/harness/runtime/private-child-process.mjs`
- `tools/harness/runtime/process-proof.mjs`
- `tools/harness/scheduler/fixture-broker/index.mjs`
- `tools/harness/scheduler/fixture-broker/providers.mjs`
- `tools/harness/scheduler/work-graph/executor.mjs`
- `tools/harness/scheduler/work-graph/runner-cli.mjs`
- `tools/harness/tests/test-command-failure.mjs`
- `tools/harness/tests/test-harness-evidence-contracts.mjs`
- `tools/harness_helper_ownership.json`
- `tools/harness_instrumentation_definitions.json`
- `tools/harness_instrumentation_policy.json`
- `tools/harness_schema_attachments.json`
- `tools/schemas/cartulary.browser_reset_attempt.v1.schema.json`
- `tools/schemas/cartulary.browser_reset_attempt.v2.schema.json`
- `tools/schemas/cartulary.harness_execution_index.v1.schema.json`
- `tools/schemas/cartulary.harness_execution_record.v1.schema.json`
- `tools/schemas/cartulary.harness_instrumentation_definitions.v1.schema.json`
- `tools/schemas/cartulary.harness_instrumentation_defs.v1.schema.json`
- `tools/schemas/cartulary.harness_instrumentation_policy.v1.schema.json`
- `tools/schemas/cartulary.harness_instrumentation_policy.v2.schema.json`
- `tools/schemas/cartulary.harness_launch_context.v1.schema.json`
- `tools/schemas/cartulary.harness_performance_explanation.v1.schema.json`
- `tools/schemas/cartulary.harness_performance_explanation.v2.schema.json`
- `tools/schemas/cartulary.harness_private_launch_observation.v1.schema.json`
- `tools/schemas/cartulary.harness_resource_index.v1.schema.json`
- `tools/schemas/cartulary.harness_resource_index.v2.schema.json`
- `tools/test_families/harness.evidence_accounting.json`
- `tools/testservices/main.go`


Additional preserved attempts: contract-02 passed 2/2 units after the scope fix;
go-02, vitest-02, integration-02, scripts-02, pg-01, apply-01 and services-01
also passed. Their source snapshot predates the last browser activity correction.
Browser-02 failed (outer exit 2, normalized `artifact/artifact_error`): build-web
rejected changed source inputs after snapshot, and the browser target projection
reported the consequent missing group result; 6 passed, 2 failed, 3 skipped units.
`cleanup-results.json` passed all 10 steps. This is caused by the explicit source
repair during validation, not an unexplained flaky browser test. No evidence or
build-source guard was bypassed. Fresh final runs use the frozen digest above.
Finalize-04 passed on that digest before those final broader runs.


Final coverage discovery found the small HTTPS browser row does not execute a
reset boundary. `make task-guide ROLE=module-author OWNER=module.collaboration`
and authored group/row review identify two small stateful groups sharing one
allocation. Reset-01 selects only session recovery (valid first-group behavior,
no reset); reset-02 selects both stateful rows to exercise the actual inter-group
reset. The attempted `make target-plan TARGET=test-slice OWNER=module.collaboration
ROWS=module.collaboration.browser_stateful.deterministic_route_owned_browser_revocation_sou_a134c3e26b`
was rejected before execution with usage exit 2 because target-plan does not
accept OWNER. No test was run by that rejected discovery command.


Final accepted commands (all outer exits **0**, source digest above):

| Run suffix | Exact Make target/selection | Result |
| --- | --- | --- |
| generate-04 | `make generate` | Passed authored-to-generated transaction |
| finalize-04 | `make agent-finalize` | 1/1; before broader gates; RESULTS_DIR unset, retained-run maintenance skipped |
| contract-03 | `make harness-contract` | 2/2; full extended contract route including failure/security/canonical/quietness fixtures |
| scripts-03 | `make lint-scripts` | 2/2 |
| drift-02 | `make generate-drift` | 4/4 |
| json-02 | `make json-shape-check` | 3/3 |
| policy-02 | `make generated-artifact-policy-check` | 3/3 |
| go-03 | `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.local_migration_activities HARNESS_DIAGNOSTICS=basic` | 1/1 |
| shell-01 | `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.current_epoch_evidence HARNESS_DIAGNOSTICS=basic` | 1/1; actual shell runner identity and parsing |
| vitest-03 | `make test-slice OWNER=package.ui ROWS=package.ui.boundary_support.design_tokens_exposes_explicit_font_role_tokens_55f6ac41b0 HARNESS_DIAGNOSTICS=basic` | 2/2 |
| integration-03 | `make service-backed-test-slice OWNER=platform.postgres ROWS=platform.postgres.integration.fresh_fixture_admission HARNESS_DIAGNOSTICS=basic` | 3/3; one shared migration plus report parsing |
| pg-02 | `make test-slice OWNER=platform.postgres ROWS=platform.postgres.support_unit.certificate_fixture_lifecycle` | 1/1 |
| apply-02 | `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.unit.test_harness_targeted_operation_validation` | 1/1 |
| services-02 | `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.testservices_lifecycle_contract` | 1/1 |
| browser-03 | `make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.https_trust HARNESS_DIAGNOSTICS=basic` | 11/11; 14/14 cleanup steps completed |
| reset-01 | `make test-slice OWNER=module.collaboration ROWS=module.collaboration.browser_stateful.deterministic_route_owned_browser_revocation_sou_a134c3e26b HARNESS_DIAGNOSTICS=basic` | 11/11; first-group stateful behavior |
| reset-02 | `make test-slice OWNER=module.collaboration ROWS=module.collaboration.browser_stateful.deterministic_route_owned_browser_revocation_sou_a134c3e26b,module.collaboration.browser_stateful.verify_multi_client_live_row_update_presence_anc_4c65a275c1 HARNESS_DIAGNOSTICS=basic` | 13/13; 16/16 cleanup steps completed |
| markdown-02 | `make lint-markdown` | Passed documentation before final status reconciliation; final checkpoint lint recorded below |

Read-only `make harness-observability-check RESULTS_DIR=<exact root>` passed for
final go-03, vitest-03, shell-01, integration-03, browser-03, reset-01 and reset-02.
All seven execution indexes are complete; sampled resource diagnostics remain
explicitly partial. Go/Vitest retain launch/process correlation even when no
resource sweep could sample their work; no CPU is fabricated for those intervals.
`make explain-run RESULTS_DIR=.cartulary/test-results/identity-s013-go-03
DETAIL=performance` and `make explain-run
RESULTS_DIR=.cartulary/test-results/identity-s013-vitest-03 DETAIL=resources JSON=1`
both exited 0 and expose explanation v2 through the existing public interface.
The routed independent fixtures establish human/JSON parity and export isolation.

Reset-02 retains 17 launches, one allocation, three leases, 176 relationships and
four activities. The activities are one shared Go migration, one helper-clock
reset and two Node report parses; all completed successfully. The reset duration
is 2,290 ms at the declared 10 ms resolution, linked to the reused allocation;
its required attempt artifact is v2. These observed durations establish producer
function only. No activities are added to canonical phase/path accounting.

| Obligation | Final disposition and evidence |
| --- | --- |
| F-010 / C-006 / V-008 | CLOSED for selected scope: neutral identity, caller migration, exact parent/run/row scope, failure precedence and fresh runner proofs; contract-03 and Go/Vitest/shell/browser runs |
| F-011 / C-007/C-009 / V-009/V-011 | CLOSED for selected scope: independent acquisition/lease facts, immutable proof and explicit conflict/late-attribution handling; contract-03, service/browser/reset runs and exact retained checks |
| F-012 / C-008/C-009 / V-010 | CLOSED for selected scope: actual owner boundaries, unsuccessful/incomplete/unavailable fixtures, generated activity vocabulary and canonical parity; Go/owner routes, contract-03, integration-03 and reset-02 |
| F-013 / V-013 | CLOSED: staged tracker preserved, stable IDs and historical failures retained, current summaries reconciled, final command/source/file ledger and handoff saved |
| V-007 | CLOSED: owner REQ-841–844 / AC-138–141, schema cutover, generated definitions/validators/topology, drift/shape/policy gates |
| V-012 | CLOSED: finalization before broader gates, current-source routed verification, source repairs and failed attempts separately disclosed |
| C-001–C-005 | PRESERVED: canonical timing, lifecycle/quietness, secure resources/readers, export isolation and same-source descriptive comparison remain covered by contract-03 |

S-007–S-013 and T-012–T-018 are DONE. No selected finding, prerequisite or required
repair remains open. Overall instrumentation remains **IN_PROGRESS**: T-007
(container counters, full wrapper/post-scan coverage and cross-clock joins) and
T-008 (complete observer and reader qualification) remain **DEFERRED**; T-009
(source-successor waiver) remains **DROPPED**. `HARNESS_DIAGNOSTICS` stays default-off.
No performance qualification, optimization, release/check acceptance or default-on
claim follows from these results. Full `check`/release suites and qualification
experiments were not run because they exceed this iteration's selected gates.
Next work, if authorized, begins with the deferred coverage/qualification items;
it must not repeat this completed identity/local-activity iteration.


Final checkpoint: `make lint-markdown`, run `identity-s013-markdown-03`, exited 0
(summary `adhoc/lint-markdown/tool-run-summary.json`). Staged and unstaged
`git --no-optional-locks diff --check` passed. Final manual status/evidence review
confirmed all accepted graph manifests share the recorded completed-source digest,
all selected gates passed, and the staged tracker hash is unchanged. This last
checkpoint records the results without changing executable source or projections.


## Coverage and qualification execution — 2026-10-09

### S-014 owner reconciliation — DONE

The current implementation request supersedes the previous tracker-only
authorization. Baseline and staged-tracker hashes are recorded at the top.
The adopted owner now declares REQ-845–849 and AC-142–146: bounded anchors,
exact-container observations, structural terminal receipt, complete observer
qualification and one validated retained model. Existing REQ-830/834/836/837/
840/842/843/844 wording was reconciled with the cutover and receipt exception.
Only the NLSpec and tracker have changed; no typed projection or runtime change
was made in this slice. `make task-guide ROLE=module-author
OWNER=harness.evidence_accounting` exited 0 and selects `test-slice` plus
`harness-evidence-contract`. Source remains the entry commit with document edits.

Accepted qualification selections are the five exact row IDs in the user's plan.
Off/basic × 0/1/4 readers share frozen source and preparation; off/0 is the total
reference. CPU/RSS upper estimate is median + 3 MAD + measurement error.
Elapsed median/p90 limits are 2%/3%, CPU 0.5% of one core, steady RSS 32 MiB,
and maximum diagnostic finalization 500 ms. Every cell must pass. No added idle
time, threshold relaxation, workload extension or source-successor waiver is
permitted. Ordinary self-counters cannot qualify the total observer.

S-014 exit: clause-by-clause owner review found no remaining prohibition on the
adopted bounded joins/receipt and no change to canonical, cleanup or export
authority. `CARTULARY_TEST_RUN_ID=instrumentation-s014-markdown-01 make
lint-markdown` exited 0; retained summary is
`.cartulary/test-results/instrumentation-s014-markdown-01/adhoc/lint-markdown/tool-run-summary.json`.
`git --no-optional-locks diff --check` passed. The default Markdown selector does
not cover this tracker yet; its coverage change remains S-022 and no tracker
lint pass is inferred. Staged entry patch is unchanged. RB-004 is closed.
S-015 is the next authorized workstream; S-016 onward still requires executed
characterization. This checkpoint was saved before S-015 work began.


### S-015 characterization and routing — DONE

Authored `tools/harness_instrumentation_qualification_plan.json` freezes the five
exact selected rows, six treatments, 30-second readers, 100-ms memory sampling,
seed 20261009, observation counts, preparation and hard budgets. A closed plan
schema is attached as an owner input. Build caches warm after discarded warm-up;
measured workload result caches are off. Each command owns a fresh managed
service session; reuse within that command remains with the lifecycle owner.
No workload extension or idle padding is permitted. This input is not itself
qualification evidence or an implemented runner.

The executable caller roster is:

| Boundary | Direct consumers / routing | Independent correction obligations |
| --- | --- | --- |
| Validated evidence | `observability.mjs`, check CLI, explain CLI, run comparison, explicit OTLP export | S-016: published foreign/duplicate-key envelope rejection, one resource traversal, immutable facts, human/JSON parity, export isolation |
| Local clocks | `local-activity.mjs`, row/browser report parsers, browser reset attempt, Go suiteservices migration and lifecycle, graph diagnostic session | S-017: bracket arithmetic, quantization, missing/foreign/reversed/suspended anchors, failed/cancelled/incomplete activity, unchanged local/canonical durations |
| Containers | `tools/testservices/main.go` private service lease and exact Docker client; suite providers feed execution/resource collector | S-018: foreign/replaced/restarted/borrowed/shared containers, counter resets, disappearance, deadline/caps, quiet acknowledgement and cleanup |
| Wrapper | `workspace/admission.mk`, `admission.sh`, `child-process.mjs`; graph runner closes diagnostics then retained scan | S-019: descriptor inheritance, nested/multiple graphs, early failure, signals/exit precedence, unsafe paths, forged binding, failed scan and interrupted/late publication |
| Complete observer | collection worker/engine/store, host observation gate, public exact-run status, Make owner and performance validator | S-020: independent process accounting, uncertainty, time-weighted memory, measurement loss, helper termination and exclusive quietness; S-021 executes all frozen cells |

All selected rows exist and are active in the authored owner families. The new
routed characterization checks their exact owner, runner and service dependency
posture. Existing evidence/command routes include secure filesystem, strict
resource parsing, integrity closure, cap/failure/quietness, export isolation,
admission and producer tests. Two new envelope correction tests deliberately
execute as TODO failures until S-016; they do not establish a current pass for
that behavior and do not preserve the defect as a contract.

Commands and evidence (all run in this WSL2 repository; main entry commit plus
listed dirty owner/test/input changes, no clean-source qualification claim):

| Command | Run identity / result |
| --- | --- |
| `make harness-evidence-contract` | `instrumentation-s015-evidence-01`, exit 0, 17,700 ms; two explicitly assigned TODO correction failures |
| `make harness-command-surface-contract` | `instrumentation-s015-command-01`, exit 0 |
| `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.local_migration_activities` | `instrumentation-s015-local-01`, exit 0; 1/1 graph units, 2,785 ms |
| `make task-guide ROLE=module-author OWNER=<owner>` | `harness.evidence_accounting`, `harness.command_surface`, `harness.browser`, `module.database_migrations`; all exit 0 |

Each run root is `.cartulary/test-results/<run identity>`; contract tool summaries
are `<target>/tool-run-summary.json`, and the Go route retains `run-manifest.json`
and `run-summary.json`. Projection validation for the newly attached schema is
part of the first schema publishing transaction in S-017. No service/browser
claim is made from this baseline; their changed owners must execute in S-018.
V-015 closes RB-005. The next workstream is S-016. This checkpoint was saved
before changing reader implementation; staged entry work remains untouched.


### S-016 unified retained validation — DONE

`loadRetainedObservability` now owns canonical, execution, envelope and resource
validation and returns immutable normalized facts. The checker reuses the loaded
resource view. Explanation and comparison use this same boundary; an identical
comparison selection reuses the same snapshot. Analysis/presentation perform no
filesystem reads. Explicit export remains separate and excludes diagnostics.
Canonical JSON and envelope JSON now use the strict parser. Every published
selected envelope must close identity, mode, schema and contiguous intervals;
missing optional envelopes remain unavailable.

The S-015 envelope correction tests are enabled and passing. An isolated worker
counts exactly one resource-file open per selected run across the loader,
explanation, self-comparison and checker. The same worker rejects imports of OTLP
modules, and fixtures check frozen nested facts/read-only maps, human/JSON parity,
foreign envelopes, duplicate keys and existing resource integrity faults.
No format version changes were made in this slice; malformed formerly accepted
envelopes are intentionally rejected with the existing artifact exit 11.

Validation in WSL2 at entry commit plus dirty implementation:

| Command | Evidence and outcome |
| --- | --- |
| `make harness-evidence-contract` | `instrumentation-s016-evidence-01`: outer Make exit 2, native test exit 1; strict-parser Buffer conversion regression; repaired, attempt retained |
| `make harness-evidence-contract` | `instrumentation-s016-evidence-02`: exit 0; no TODO correction failures |
| `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.local_migration_activities HARNESS_DIAGNOSTICS=basic` | `instrumentation-s016-local-01`: exit 0, 1/1 units, 2,651 ms |
| `make harness-observability-check RESULTS_DIR=.cartulary/test-results/instrumentation-s016-local-01` | exit 0; complete diagnostic/execution publication, observed-partial physical coverage |
| `make explain-run RESULTS_DIR=.cartulary/test-results/instrumentation-s016-local-01 DETAIL=performance` | exit 0, bounded human view |
| `make explain-run RESULTS_DIR=.cartulary/test-results/instrumentation-s016-local-01 DETAIL=resources JSON=1` | exit 0, explanation v2 |
| `git diff --check` | exit 0 |

Run roots and summary conventions match S-015. Fresh graph source digest:
`sha256:c64394df70312868b424af30d7252a5a1f60786de49004f99c7ccf195d680cd3`;
toolchain `sha256:4b62e2558e6dcbe1d3c632fb9c0bb47e5a89e5dda08f103a72f61347b5ef1810`;
system `sha256:546793d134ea4063515a25068aa22b78078fc1613c1ee56fb5f9e41c5c550588`.
V-016 passes; F-018 implementation correction is complete, subject to final
integrated validation. S-017 is ready. This checkpoint precedes the schema and
clock work. No qualification or overall completion is claimed.


### S-017 clock anchoring and coordinated schema cutover — DONE

The runtime clock boundary brackets boot-clock reads with producer-local clocks.
Node activities, cross-process boot-clock helpers and Go migration observations
retain beginning/end anchors without replacing independently measured duration.
The scheduler exposes its existing epoch to the diagnostic session; its canonical
clock, event boundaries and accounting algorithm are unchanged. The execution
index retains two graph anchors and the canonical epoch. Pure placement analysis
intersects consistent producer/graph offsets, composes uncertainty, enforces the
25-ms bound and reports explicit unavailability. Human and JSON views expose the
same bounded placements. Failed and cancelled local activities remain measurable;
incomplete, reversed, suspended and incompatible evidence cannot invent ordering.

Installed complete target shapes and minimum readers: execution record/index v2,
resource sample v2/index v3, envelope v2, explanation v3, policy v3, shared
definitions/catalog v2, private launch projection v2, terminal receipt v1 and
qualification plan/manifest/result v1. Container roster, receipt binding and
qualification fields have closed shapes; capabilities still awaiting S-018–S-020
remain absent/null/not-observed. The retained boundary already rejects malformed
published receipts. Collector counters now explicitly name the measured thread,
JavaScript heap and excluded startup/parent/helper/reader/publication work; they
cannot establish qualification. Obsolete diagnostic formats have no translators.
Generated definitions/validators were replaced through the Make generator; its
superseded output is retired only after installing the new closure.

Independent arithmetic uses literal expected offsets [3999,4010] and [4848,4860],
composition [838,861], placement [48,71]/[78,101], and exact 25-ms acceptance.
Tests cover quantization, read order, namespace mismatch, reversed/local/suspend
clocks, missing anchors, unsuccessful and incomplete operations, immutability,
unchanged local duration, and existing canonical accounting/producer contracts.

Executed commands and attempts:

| Command | Result / retained run identity |
| --- | --- |
| `make generate` | First two pre-bootstrap attempts failed before publication because primitive qualification arrays lacked explicit schema types and introduced an AJV runtime equality dependency; corrected without adding runtime dependencies |
| `make generate` | exits 0: `20261010T023519Z-p79113`, `20261010T023722Z-p83997`, final `instrumentation-s017-generate-03` |
| `make format` | `instrumentation-s017-format-01`, exit 0, 2/2 units |
| `make harness-evidence-contract` | `instrumentation-s017-evidence-01` failed one scope/reference consistency fixture; tightened the new sample schema; `instrumentation-s017-evidence-02` exited 0, 89 tests, no TODOs |
| `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.local_migration_activities HARNESS_DIAGNOSTICS=basic` | `instrumentation-s017-local-01` and final `instrumentation-s017-local-02` exit 0; final 1/1 units, 2,942 ms |
| `make json-shape-check` | `instrumentation-s017-shapes-01` failed three new registry validation labels; corrected to the registry's `json-shape-check` constant; `instrumentation-s017-shapes-03` exit 0, 3/3 units |
| `make json-shape-check VERBOSE=1` | `instrumentation-s017-shapes-02` failed preflight (unknown Make input VERBOSE), no graph evidence; diagnosis used retained unit logs instead |
| `make generated-artifact-policy-check` | `instrumentation-s017-policy-01`, exit 0, 3/3 units |
| `make generate-drift` | `instrumentation-s017-drift-01` and final `instrumentation-s017-drift-02`, exit 0, 4/4 units |
| `make harness-command-surface-contract` | `instrumentation-s017-command-01`, native exit 0 |
| `make harness-observability-check RESULTS_DIR=.cartulary/test-results/instrumentation-s017-local-02` | exit 0, complete diagnostic/execution publication |
| `make explain-run RESULTS_DIR=.cartulary/test-results/instrumentation-s017-local-01 DETAIL=performance` | exit 0, real placement uncertainty 15.712 ms |
| `make explain-run RESULTS_DIR=.cartulary/test-results/instrumentation-s017-local-02 DETAIL=resources JSON=1` | exit 0, explanation v3; real placement uncertainty 12.919 ms |
| `git diff --check` | exit 0 |

All retained roots are `.cartulary/test-results/<identity>`; contract/generator
summaries follow S-015 conventions and graph summaries retain source metadata.
Final executable source digest is
`sha256:604821ca2a642dec1f00526d03ae65f7d0c91773afa4fb8765f054de19387574` at
the entry commit with dirty work; toolchain/system remain the S-016 values.
These runs are functional evidence, not performance qualification. Source changes
require fresh S-021 cohorts. V-017 and projection gates pass. S-018 is ready;
this checkpoint precedes its container implementation. Preserve rollback across
all changed schemas, producers, consumers and generated outputs together.


### S-018 lifecycle-owned container observations — DONE

The lifecycle owners provide exact private service-lease proof for owned and
borrowed sessions. A run-owned helper uses the existing Go Docker client through
a port containing only inspect and non-streaming single-sample stats. It checks
lease/run/borrower binding, labels, exact container ID and unchanged start
identity before/after each observation. It never lists, acquires, resets or removes
containers. Raw IDs, names, labels and lease paths remain private; retained scopes
use opaque identities and actual Docker CPU, memory usage/limit, block-I/O and
PID counters. Missing fields remain unavailable. Shared containers remain context,
independent of process/cgroup/guest totals and canonical fixture ownership.

The existing collector drives batches at its five-second context cadence. There
is one helper per run and one in-flight batch, a 250-ms deadline, 32 live and 256
retained lifetime identities. The helper has no polling loop. A private socket
keeps requests/cancellation independent of synchronous parent setup; anchored
file-descriptor socket paths avoid Unix path-length failures. The parent owns
helper termination/reaping, Linux parent-death signaling prevents an orphan after
parent death, and worker failure drains the helper before quiet acknowledgement.
Pause/stop cancels requests and waits for drain. Counter resets and restart
identities cannot bridge incompatible intervals. Omission/deadline facts remain
separate from test outcomes and resource coverage remains observed-partial.

Independent Go/Node tests cover exact read-only access, missing versus zero,
foreign/replaced/restarted containers, duplicate/unsafe private proof, borrowed
revocation, caps, cancellation/deadline, counter resets, delayed drain,
malformed replies, forced worker death and no restart after failure.

| Command | Result / retained identity |
| --- | --- |
| `make format` | `instrumentation-s018-format-01`, `-02`, `-03`; all exit 0 |
| `make generate` | `instrumentation-s018-generate-01`, `-02`; both exit 0 |
| `make harness-evidence-contract` | `instrumentation-s018-evidence-01`, `-02`; both exit 0 |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.testservices_lifecycle_contract` | `instrumentation-s018-services-unit-01`, `-02`; both exit 0; final 1/1 unit, 8,575 ms |
| `make service-backed-test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.schema_bootstrap HARNESS_DIAGNOSTICS=basic` | `instrumentation-s018-service-01`; exit 0, 3/3 units, 60,715 ms |
| `make service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.https_trust HARNESS_DIAGNOSTICS=basic` | `instrumentation-s018-browser-01`; exit 0, 11/11 units, 73,396 ms |
| `make explain-run RESULTS_DIR=.cartulary/test-results/instrumentation-s018-service-01 DETAIL=resources JSON=1` | First attempt failed explanation schema validation because analysis correlation mutated container roster entries. Repaired by copying analysis entries; repeated exact-run explanation passed. Original retained artifacts were unchanged |
| `make explain-run RESULTS_DIR=.cartulary/test-results/instrumentation-s018-browser-01 DETAIL=resources JSON=1` | exit 0 |
| `make harness-observability-check RESULTS_DIR=<exact service/browser root above>` | both exit 0; complete diagnostic/execution publication |
| `make harness-command-surface-contract` | `instrumentation-s018-command-01`; exit 0 |
| `make generate-drift` | `instrumentation-s018-drift-01`; exit 0, 4/4 units |
| `make json-shape-check` | `instrumentation-s018-shapes-01`; exit 0, 3/3 units |
| `make generated-artifact-policy-check` | `instrumentation-s018-policy-01`; exit 0, 3/3 units |
| `git diff --check` | exit 0 |

Real service evidence retains two owned container identities with two samples
each; browser evidence retains two with three samples each. Both have zero
container omissions/deadline overruns, all six counter names, no unit attribution,
and successful required cleanup. Both source digests are
`sha256:3f3ade424f9ecf359671c8a82065ff6e7bec4297786260393f29151a94de9cdc`.
The subsequent reader-only roster correction and regression tests were verified
against those unchanged artifacts and the final contract suites. Final routed
source digest: `sha256:834dd30aae20258150fe10a953db51b3665ad96c2298305a8c77e720c7dcc422`. Entry commit remains dirty;
toolchain/system are unchanged. No diagnostic schema acceptance set changed
beyond the S-017 target shapes. V-018 and projection gates pass. S-019 is ready;
this checkpoint precedes wrapper implementation. Helper self-cost and readers
still require complete S-020/S-021 accounting; no qualification claim is made.


### S-019 supervised wrapper and terminal receipt — DONE

The outer basic-mode admission wrapper now supervises the managed Make child,
retains descriptor 9 across descendants, forwards INT/TERM to the child group,
waits for cleanup and preserves the child's exit. Nested admission reuses the
outer invocation. Unsupported early observation remains unavailable. Private
owner/graph registrations bind scan start/end/outcome and graph completion to
one live wrapper identity. Graph envelopes close before the retained scan;
ordinary producers remain stopped. The sole optional post-scan writer validates
all memberships before publishing the closed receipt, with no scanned-artifact
rewrite. Receipt loss, disk failure and partial multi-graph publication preserve
the primary result and cannot imply complete success for a missing membership.

Receipts are bounded to 64 KiB/64 memberships, bind exact envelope/scan digests,
and reject forged, duplicate, unsafe or stale bindings. Shared outer time is
reported once with an invocation identity; human/JSON views expose the same
validated facts. No schema acceptance set changed. Readiness and initial
shell/Make startup exclusions remain explicit, and external receipt-publication
measurement belongs to S-020. Independent fixtures cover nested admission,
multiple graphs, early failure, missing/failed scans, INT/TERM and cleanup,
primary exits, unsafe paths, overflow, interrupted/disk-failed publication and
changes to artifacts after publication.

| Command | Result / retained identity |
| --- | --- |
| `make harness-evidence-contract` | `instrumentation-s019-evidence-01` failed two manifest-less diagnostic-session fixtures; optional registration was corrected. `-02` and `-03` exit 0; final 101 tests, no skipped/todo |
| `make harness-command-surface-contract` | `instrumentation-s019-command-01`; exit 0 |
| `make lint-shell` | `instrumentation-s019-shell-01`, `-02`; both exit 0, 4/4 units |
| `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.local_migration_activities HARNESS_DIAGNOSTICS=basic` | `instrumentation-s019-local-01`; exit 0, 1/1 unit, 11,302 ms; concurrent verification contention is functional evidence only |
| `make harness-observability-check RESULTS_DIR=.cartulary/test-results/instrumentation-s019-local-01` | exit 0; valid bound terminal receipt, partial resources, complete execution |
| `make explain-run-performance RESULTS_DIR=.cartulary/test-results/instrumentation-s019-local-01 JSON=1` | rejected: nonexistent target, exit 2; no graph and no evidence change |
| `make explain-run RESULTS_DIR=.cartulary/test-results/instrumentation-s019-local-01 DETAIL=performance` | exit 0; human receipt facts, shared outer duration 13,890 ms |
| `make explain-run RESULTS_DIR=.cartulary/test-results/instrumentation-s019-local-01 DETAIL=resources JSON=1` | exit 0; explanation v3 |
| `make generate-drift HARNESS_DIAGNOSTICS=basic` | `instrumentation-s019-drift-01` failed stale generated topology index after runner edits; exit 2 |
| `make json-shape-check` | `instrumentation-s019-shapes-01` failed the same stale index; after generation, `-02` exit 0, 3/3 units |
| `make generate` | `instrumentation-s019-generate-01`; exit 0, refreshed topology index through its owner |
| `make generate-drift` | `instrumentation-s019-drift-02`; exit 0, 4/4 units |
| `make generated-artifact-policy-check` | `instrumentation-s019-policy-01`; exit 0, 3/3 units |
| `git diff --check` | exit 0 |

All roots are `.cartulary/test-results/<literal identity above>`. Entry commit
remains `1d423c6e5f9cd204f8210fa9958c780fd933d293`, dirty. Real local-run source:
`sha256:8db8c159a76f660e7e3808b78a8b5cf51a14a7777dc6a49b73be7eb5fde8704b`.
Final projection-check source:
`sha256:791eeb7d446df7ade8ccea1f140bd90b30936b3be99766bfab4ae6bdc7252690`.
Toolchain/system match the S-016 recorded identities; WSL2 environment is
unchanged. Failed attempts remain retained. V-019 and projection gates pass.
S-020 is ready; this saved checkpoint precedes complete observer measurement.
Rollback remains coordinated across wrapper, private protocol, envelope producer,
receipt publisher and retained readers. No engineering qualification is claimed.


### S-020 complete observer measurement — DONE

The semantic Make qualification entry owns the closed authored workload plan,
clean-source freeze, preparation, deterministic treatment order, native process
supervision, public readers and atomic evidence assembly. It is outside ordinary
checks and canonical performance portfolios. The existing performance checker
now dispatches qualification manifests through a separate strict validator;
canonical manifests and baseline publication retain their existing boundary.
Every selected retained run is loaded once through the unified model, with exact
source/commit, graph, toolchain, capacity, policy, preparation, selection, result
and receipt bindings. Partial matrices fail qualification; malformed evidence
fails inspection. No arbitrary command runner or format translator was added.

The native Linux subreaper uses one wait4 owner for complete workload and reader
CPU, including waited descendants once. The parent includes native startup,
serialization and exit in a conservative support-CPU estimate. RSS samples sum
process leaders once, using smaps_rollup, at 100 ms with recorded sweep spans;
central-80% time weighting, gaps, page quantization and sampled peaks remain
separate. This is a sampled-memory estimand, not an unsampled peak guarantee.
Support I/O uses Linux resource-accounting storage blocks (512-byte units), not
logical syscall bytes. Support CPU, sampled simultaneous support RSS and I/O
are reported separately from the observer treatment. Windows-host and daemon
internals remain outside coverage.

A persistent private adapter uses the existing observation gate to fence public
readers during exclusive work. Each configured reader has one outstanding
request, starts after initial publication and resumes at the fixed 30-second
cadence without catch-up bursts. Cancellation drains owned children. Native
meter death cannot turn lost measurement into zero cost; failed cleanup preserves
private recovery proof and prevents a successor attempt. Independent fixtures
cover waited CPU arithmetic, simultaneous memory, signed paired medians/MAD,
p90, uncertainty, finalization interval union, instability, strict matrix order,
incomplete/forged/unsafe evidence and forced helper death. Cold/warm-up attempts
remain excluded from paired statistics but included in observed finalization
maxima. Component collector counters are explicitly diagnostic-only.

| Command | Result / retained identity |
| --- | --- |
| `make format` | `instrumentation-s020-format-01` through `-11`; all exit 0 |
| `make generate` | `instrumentation-s020-generate-01` failed the disallowed tools-package selector; implementation moved to internal/testutil with a thin binary. `-02` failed unsorted authored rows; rows sorted. `-03` through `-08` exit 0 |
| `make harness-evidence-contract` | `instrumentation-s020-evidence-01` through `-05`; all exit 0, including complete synthetic 210-attempt manifest, missing coverage and forced-meter-death fixtures |
| `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.instrumentation_measurement` | `instrumentation-s020-native-01` passed initial arithmetic fixtures. `-02`/`-03` failed invalid private-to-public run-root arguments; moved run identity to the existing environment boundary. `-04` failed nested host-admission timeout; private functional tests preserve the caller lease while public qualification rejects one. `-05` failed missing receipt caused by inherited descriptor 9; only the proven descriptor is now CLOEXEC for the fresh measured invocation. `-06` passed. `-07`/`-08`/`-09` failed inaccessible zombie I/O counters; replaced that approach with waited resource accounting. `-10` exit 0, 1/1 unit, 10,655 ms; real public workload, public reader, complete parent bridge, independent waited CPU and cancellation pass |
| `make harness-instrumentation-qualify QUALIFICATION_PLAN=tools/harness_instrumentation_qualification_plan.json` | `instrumentation-s020-entry-01`; expected configuration rejection of dirty checkout, exit 2; no qualification attempts |
| `make harness-smoke-make-node-tools` | nonexistent public target rejected, exit 2; no graph. This name is a harness check, not a Make target |
| `make harness-command-surface-contract` | `instrumentation-s020-command-01`; exit 0 |
| `make lint-scripts` | `instrumentation-s020-scripts-01`, `-02`; both exit 0, final 2/2 units |
| `make generate-drift` | `instrumentation-s020-drift-01`; exit 0, 4/4 units |
| `make json-shape-check` | `instrumentation-s020-shapes-01`; exit 0, 3/3 units |
| `make generated-artifact-policy-check` | `instrumentation-s020-policy-01`; exit 0, 3/3 units |
| `git diff --check` | exit 0 |

All roots are `.cartulary/test-results/<literal identity above>`. Failed attempts
are preserved. Entry commit remains `1d423c6e5f9cd204f8210fa9958c780fd933d293`,
dirty, with the original staged tracker unchanged. Native integration source:
`sha256:c2ec29d91396fa0d2c307fe27abb8d95f117562a20ef6e655c84f3e4d78013a6`.
Final projection-check source:
`sha256:d675cfa49e072147fd6c09136fcf0e1d0328a4f91dc32bd8a2c8cfa5ff8ebc5e`.
Toolchain/system match S-016; WSL2 is unchanged. The S-017 qualification v1
schemas are now implemented without acceptance-set changes. V-020 passes.
S-021 is ready after this saved checkpoint; RB-006 remains open and no budget
qualification is claimed. Rollback includes the authored public target, native
supervisor, closed plan, reader adapter, validator, routing and generated outputs.


### S-021 engineering qualification — IN_PROGRESS

Candidate 1 is the isolated clean commit
`97c989d143110a8f30fab7cf49392aa1d2dd5c97`, named
`codex/instrumentation-qualification-01`, checked out at
`/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary`.
The commit uses an alternate index and excludes the controlling tracker, so the
original staged tracker is unchanged. This primary tracker remains the sole
execution ledger. The app created the checkout but its asynchronous status did
not advance; explicit attachment succeeded and Git verified the clean revision.

Started there:
`CARTULARY_TEST_RUN_ID=instrumentation-s021-qualification-01 make harness-instrumentation-qualify QUALIFICATION_PLAN=tools/harness_instrumentation_qualification_plan.json`.
The exact output root is that checkout's
`.cartulary/test-results/instrumentation-s021-qualification-01/qualification`.
No qualification or readiness conclusion is drawn before the complete matrix
and independent manifest validation finish.


Candidate 1 entry failed before collecting attempts: the new native-build
prerequisite used schema validation before the fresh checkout installed AJV.
`instrumentation-s021-qualification-01` exited 2 at the helper build. Named
qualification finding **F-020: measurement readiness omitted frontend schema
dependencies** is assigned back to S-020. The native build now explicitly depends
on the existing public `frontend-install` readiness boundary. This is an
executable repair; candidate 1 cannot be relabeled. A new freeze is required.
`instrumentation-s021-prepare-01 make frontend-install` prepares the isolated
checkout through its public owner; no lockfile edits or private installs.


Pre-freeze inspection found **F-021: provenance cleanliness included human
documentation**, even though source hashing already excluded Markdown and the
typed restricted roots. This would let a later tracker checkpoint affect runtime
eligibility. The graph and qualification runner now share executable provenance
with Git pathspec exclusions derived from the typed input policy, plus Markdown
and README exclusions. Source hashing uses the same README boundary. A fixture
edits restricted-root content, Markdown and README without changing provenance
or digest, then confirms that executable changes do change both. This S-020
binding repair precedes the replacement freeze; no collected cohort is reused.


F-020/F-021 repairs passed `make generate` (roots
`instrumentation-s021-repair-generate-01`, `-02`), `make harness-evidence-contract`
(`instrumentation-s021-repair-evidence-01`), `make harness-command-surface-contract`
(`instrumentation-s021-repair-command-01`), `make lint-scripts`
(`instrumentation-s021-repair-scripts-01`), `make generate-drift`
(`instrumentation-s021-repair-drift-01`, `-02`), `make json-shape-check`
(`instrumentation-s021-repair-shapes-01`), `make generated-artifact-policy-check`
(`instrumentation-s021-repair-policy-01`) and `git diff --check`; all exit 0.
These roots are in the primary checkout. F-020/F-021 are functionally resolved;
the replacement clean entry remains part of qualification verification.

Candidate 2 is commit `86dfc7b107edf14a3948100f5f58c08813a4d3a3`, branch
`codex/instrumentation-qualification-02`, executable source
`sha256:bf2748fdbe55ed7bd32efb58b76be45ab7e68f734b389110846259d79e9596d1`.
The same isolated checkout switched cleanly to this revision. Started the same
public qualification command with run ID `instrumentation-s021-qualification-02`.
Candidate 1's failure and preparation evidence remain intact. Candidate 2's
result parent is the same checkout's `.cartulary/test-results/`.


Candidate 2 entry exposed a narrower F-020 follow-up: a public target used as a
Make prerequisite inherited the qualification-only input and correctly rejected
it (`QUALIFICATION_PLAN is not declared for target frontend-install`, exit 2).
No attempts were collected. The measurement binary now depends on the existing
Make-owned frontend installation stamp, matching other native build owners,
instead of invoking a public command with incompatible input scope. F-020 is
reopened until the corrected entry executes on a new frozen candidate.


The corrected entry passed preparation and reached the expected dirty-source
rejection in the primary checkout (`instrumentation-s021-repair-entry-01`,
exit 2 with `qualification requires a clean frozen checkout`). `make generate`
(`instrumentation-s021-repair-generate-03`) and `git diff --check` passed.
Candidate 3 is clean commit `5130545f3e6f8ce0a59ffe46a13a6beeca890ea9`, branch
`codex/instrumentation-qualification-03`, in the same isolated checkout.
Started the unchanged qualification command with run ID
`instrumentation-s021-qualification-03`; no predecessor evidence is relabeled.


Candidate 3 has completed the Go cohort, including the prescribed sixth block.
All Go observations have complete coverage. **F-022: complete observer exceeds
qualification budgets** is now confirmed for this workload. Independent arithmetic
on the six retained measured blocks gives CPU median + 3 paired MADs (before the
additional measurement-error bound) of 18.894% for off/1 reader, 90.843% for off/4,
9.199% for basic/0, 27.343% for basic/1 and 93.045% for basic/4, against 0.5%.
Basic/0 outer elapsed median/p90 are 4.748%/6.128%, against 2%/3%. These are
intermediate diagnostic calculations, not the final public validator result.
The full frozen attempt continues; no unfavorable observation is removed.

F-022 belongs to the S-016 public-reader/validation and S-018/S-019 collector/
wrapper cost boundaries, with S-020 accounting retained. Runtime schema
compilation and repeated public-command startup are concrete candidate costs;
no change is made to the frozen candidate during collection. Any repair requires
new executable provenance and affected qualification cohorts. S-022 remains gated.


**F-023: qualification loses descendant coverage for fault-injection workloads.**
Candidate 3's shell commands pass, but the meter marks lifecycle coverage
incomplete. The same symptom occurred in one Vitest sixth-block treatment.
The shell row includes intentional worker-death fixtures. The first candidate
retains coverage flags but not native loss codes, limiting exact diagnosis.
An S-020 repair is being prepared in the primary checkout while the isolated
candidate continues unchanged: move the observation-gate helper outside the
native subreaper's measured cohort. Kernel adoption then proves that otherwise
unattributed descendants belong to the measured workload/reader union, while
external support remains separate. Count their CPU once in the cohort residual;
retain direct reader CPU as a component breakdown. Sample surviving adopted
children, prove the sibling gate identity, drain it before acknowledgement, and
retain closed loss/adoption diagnostics. Required validation includes an
independent adoption experiment and the real fault-injection shell workload.
No existing failed observation is relabeled eligible by this proposed repair.


The F-022 repair is prepared only in the primary checkout: hot retained/live
schemas use the existing generated standalone-validator mechanism; the authored
status recipe selects an explicitly capable read-only entrypoint that performs
shared public preflight and argument normalization in one process. It removes
the redundant preflight/launcher process chain while preserving secure selection,
rejection, signals and output contracts. Recipe validation restricts this path
to registered public entrypoints with no side effects or retained writer.
No diagnostic schema acceptance set, cadence, workload selection or budget is
changed. Existing status-command hostility/selection fixtures and wrapper,
reader, evidence, generation and drift gates must pass before the next freeze.
Generation and test execution are deliberately deferred until candidate 3 drains
to avoid concurrent verification load on its measured windows.


S-014/F-019 current-summary follow-up: removed remaining planning-session
authorization/readiness statements from current Sections 9–12 and labeled their
old checkpoints historical. T-007/T-008/T-027 now show IN_PROGRESS. This is an
editorial correction, not a newly passing runtime or qualification gate.

F-022 preparation also splits generated standalone validators into lazy modules
within one generated transaction and declared generated root. The existing
registry API and schema acceptance remain unchanged; the dependency-free
bootstrap test now verifies that one requested schema loads one module. All
new generated outputs and verification remain pending until candidate 3 drains.


Candidate 3 was deliberately stopped through native session 2631 (`SIGINT`,
outer exit 130) after confirmed failures required a new executable candidate
for every cohort. This avoids continuing a superseded failed experiment; no
favorable subset is selected. The manifest remains `incomplete` with 154
published attempts: Go 48, Vitest 48, shell 42 and service 16. Browser did not
start. The interrupted `qualification-service-measured-1-basic-1` run and command
log remain alongside the manifest but have no completed qualification result.
This interrupted publication is not silently discarded or treated as eligible.

Independent frozen-candidate validation command:
`CARTULARY_TEST_RUN_ID=instrumentation-s021-validation-03 make
harness-performance-check EVIDENCE_ROOTS_FILE=/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary/.cartulary/test-results/instrumentation-s021-qualification-03/qualification/qualification.json`.
Result: validator exit 13, public Make exit 2. Go and Vitest are unstable and
exceed CPU budgets; Vitest also has ineligible evidence; shell coverage fails;
service/browser are incomplete. The public validator reports Go CPU upper
estimates 18.902%, 90.882%, 9.2025%, 27.355%, 93.086% for off/1, off/4, basic/0,
basic/1, basic/4 respectively. Go basic/0 additional RSS upper estimate is
38.408 MiB (limit 32 MiB); basic/4 is 74.733 MiB. These results supersede no
historical evidence and establish no successful qualification.

Repair verification began only after the candidate stopped: `make format`
(`instrumentation-s021-repair-format-01`) passed, 2/2 units. `make generate`
(`instrumentation-s021-repair-generate-04` and `-05`) failed before publication
because the newly generated schemas require AJV's equality helper. The generator
now embeds the pinned helper and its license, preserving validation semantics
without requiring installed dependencies in generated validators. Generation
`-06` and remaining functional gates are in progress.


S-020 repair verification update (primary checkout, WSL2; public Make exits 0):

| Exact command (run-ID prefix `instrumentation-s021-repair-`) | Result/root suffix |
| --- | --- |
| `make generate` | `generate-06`, pass |
| `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.instrumentation_measurement` | `native-01`, 1/1, 34,518 ms |
| `make harness-evidence-contract` | `evidence-02`, pass |
| `make harness-command-surface-contract` | `command-02`, pass |
| `make lint-scripts` | `scripts-02`, 2/2 |
| `make generate-drift` | `drift-03`, 4/4 |
| `make json-shape-check` | `shapes-02`, 3/3 |
| `make generated-artifact-policy-check` | `policy-02`, 3/3 |
| `make task-guide ROLE=module-author OWNER=harness.command_surface` | `guide-01`, confirmed public routing |

Retained graph/tool roots, where the command publishes them, are under primary
`.cartulary/test-results/<prefix><suffix>`. Direct contract commands returned exit
0 without manufacturing a retained graph. The native gate's source digest was
`sha256:dba983290608d8e393cfbb4360ecbb7b178aa2e60f9eb39d2d2d05c6c027ba16`;
the only subsequent change added command-surface rejection cases. Projection
gates bind source `sha256:b3de6d78474c512123951b50403e70ea4a629bac3b0a7d96653555d7a78769aa`.
The native gate covers ordinary completion, the actual fault-injection shell
workload, cancellation, and an independent adopted-descendant accounting
experiment. F-023 is functionally repaired; all failed observations remain
unchanged. Public readers preserve input rejection and the sole-goal restriction.
`make harness-contract` (`contract-01`) is running to cover the changed generated
validator's dependency-free bootstrap boundary before the next freeze; this is a
repair boundary gate, not the final S-022 handoff invocation.


`harness-contract` repair gate `contract-01` failed (204/205 assertions, public
exit 2; run summary and `unit-logs/target-harness-contract/stdout.log` under its
exact root). The related failure is an owner-boundary violation: three callers
of the F-021 provenance helper imported its private module. The helper is now
exported through the existing test-catalog facade and all three imports use it.
The contract gate now explicitly runs the dependency-free foundation fixture.
Qualification cell presentation also requires at least five matched blocks before
reporting eligibility; interrupted partial cohorts cannot display a passing cell
even though the overall validator already rejected them. No threshold or eligible
completed-window acceptance changes. New repair gates follow before freezing.


Repair generation `instrumentation-s021-repair-generate-07` passed. The full
`harness-contract` gate `instrumentation-s021-repair-contract-02` reports pass
after facade repair and explicitly covers the dependency-free validator fixture.
Its source digest is `sha256:338e8160625e11acc2460d8e73da6da7f98e0bb20a71b3e48a34212b75acfc76`.
`lint-scripts` (`scripts-03`) also passed. Native and drift repeat gates were
queued by host admission behind the contract gate; quiet output was not treated
as a timeout or reason to rerun.


Native repair repeat `native-02` failed (public exit 2; inner test assertion
exit 10). `run-summary.json` and
`unit-logs/go-1fd2784951e2f699-38d0548a57a6/stderr.log` retain the failure:
waited-descendant and memory coverage pass, but reader lifecycle fails in the
complete and fault-workload cases; short-run finalization also lacks coverage.
These checks ran alongside the full contract and drift gates; an observation-gate
pause is a hypothesis, not an established cause. Drift `drift-04` passed (4/4).
The meter now retains reader start/skip/failure counters for diagnosis. External
envelope observation is separated from the 100 ms memory tick and uses the
already-owned command-completion poll; a short finalization window can otherwise
fall entirely between memory samples. Required cadence, accounting and budgets
are unchanged. The reader Make recipe also binds canonical task metadata for
argument normalization; its regression fixture supplies hostile inherited
manifest selectors. Format `format-02` passed; generation `generate-08` and
serial native verification follow. S-020 remains reopened until these gates pass.


### S-020 follow-up repair checkpoint — DONE; S-021 remains IN_PROGRESS

Final repair source: `sha256:c53efc145e8ee088eddbfaabada5dac12216cc3b8c3da98ffbafc8308ddfe951`
at the primary dirty checkout; the clean executable commit is created next.
`make generate` (`generate-08`), `make harness-command-surface-contract`
(`command-03`), `make generate-drift` (`drift-05`), `make format` (`format-03`),
the native owner row (`native-04`, 1/1, 33,066 ms), and `make lint-scripts`
(`scripts-05`, 2/2) pass. These use prefix `instrumentation-s021-repair-`;
previous repair gates and exact roots are listed above. Full `harness-contract`
`contract-02` passed, 2/2 units, 175,502 ms; the subsequent changes are the
canonical reader-manifest binding, separated finalization polling, closed reader
counters and their tests. Generation, schema and generated-policy gates for the
validator split remain applicable; no schema acceptance set changed.

The native integration test now distinguishes an admitted reader from the
explicit outcome where the observation gate skips every launch. It verifies
unavailable reader/quietness coverage in that case; the qualification validator
still rejects it. It never substitutes zero reader cost for missing coverage.
Closed reader counters are included in qualification diagnostic output.

F-020/F-021 readiness/provenance fixes and F-023 descendant ownership fixes are
functionally complete. F-022 remains open until the entire fresh matrix passes;
CPU, RSS, elapsed, finalization and stability budgets are unchanged. The next
candidate includes lazy generated validators, one-process read-only status
preflight, bounded sibling-gate support, independent finalization observation and
strict partial-cell rejection. The pre-existing staged tracker remains intact.

Before reusing the isolated checkout, the exact public service route
`CARTULARY_TEST_RUN_ID=instrumentation-s021-cancel-cleanup-03 make
service-backed-test-slice OWNER=module.database_migrations
ROWS=module.database_migrations.integration.schema_bootstrap
HARNESS_DIAGNOSTICS=basic` is running at candidate 3 to exercise its lifecycle
owner's stale-resource cleanup after the interrupted attempt. This is functional
cleanup evidence, not another qualifying observation. No unproven process or
container is selected manually. This saved repair checkpoint precedes candidate
4 collection; S-022 has not begun.


Candidate 4 was frozen with an alternate Git index as commit
`1894851cbff3fbd748e09e4d140ab8becc462e47`, branch
`codex/instrumentation-qualification-04`, parent candidate 3. Primary `main`,
the normal index and the staged tracker blob are unchanged. The tracker is
excluded from the snapshot; its working copy remains the sole execution ledger.
Executable source digest is the final repair digest above, to be confirmed by
the new qualification manifest before accepting observations.


Candidate 3 cleanup verification passed: `instrumentation-s021-cancel-cleanup-03`,
3/3 units, 47,021 ms, public exit 0. `cleanup-results.json` records all six steps
completed, including services/runtime close. Exact-run `make fixture-report
RESULTS_DIR=.cartulary/test-results/instrumentation-s021-cancel-cleanup-03 JSON=1`
also succeeded. This new report does not retrospectively supply a missing
terminal report for the interrupted attempt. The clean isolated checkout was
then switched to `codex/instrumentation-qualification-04`; ignored historical
results remain retained. Candidate 4 collection uses
`CARTULARY_TEST_RUN_ID=instrumentation-s021-qualification-04 make
harness-instrumentation-qualify
QUALIFICATION_PLAN=tools/harness_instrumentation_qualification_plan.json`.
No primary-checkout verification will run concurrently with its measured windows.


### S-021 candidate 4 rejection and next repair entry — IN_PROGRESS

Frozen candidate 4 manifest binds commit `1894851cbff3fbd748e09e4d140ab8becc462e47`
and source `sha256:9342cf2bf359f9a78abbb922ae90356c59efccb5d19ccb179cea151237e1dfcb`.
This differs from the primary dirty-checkout digest: the clean Git checkout
normalizes existing non-executable file modes (including pre-existing 0600
goldens and 0444 vendored source) to 0644. Those primary modes were not changed.
The manifest's actual frozen identity is authoritative for these observations.

Native session 43242 was stopped through `SIGINT` (outer exit 130) after all six
Go blocks failed CPU budgets. It retains 52 published attempts (48 Go, 4 Vitest
cold), all with complete coverage. `qualification-vitest-cold-0-basic-0` retains
its interrupted run without a completed result; later workloads did not start.
The manifest remains incomplete. Exact independent validation was
`CARTULARY_TEST_RUN_ID=instrumentation-s021-validation-04 make
harness-performance-check EVIDENCE_ROOTS_FILE=/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary/.cartulary/test-results/instrumentation-s021-qualification-04/qualification/qualification.json`:
validator exit 13, public Make exit 2.

Go CPU upper estimates are 12.171%, 62.109%, 4.3979%, 16.738%, and 65.652%
for off/1, off/4, basic/0, basic/1, basic/4. Basic/0 RSS upper estimate is
33.687 MiB and basic/4 is 35.639 MiB. Diagnostic finalization improves to
124.5–127.31 ms, but CPU/RSS and stability gates still fail. No thresholds,
workloads or cadence will be changed to accept this candidate.

Post-rejection profiling used the public `make test-run-status` path, with a
temporary installed-node wrapper enabling the native Node CPU profiler; it is
diagnostic evidence, never qualification. The first attempt passed `NODE_BIN`
on the Make command line and was correctly rejected as an undeclared input.
The corrected inherited `NODE_BIN` invocation succeeded. Profiles are retained
at `/tmp/cartulary-reader-profile-wgtc1yzw/` (including the rejected attempt).
The Node profile spans 42.919 ms, while the complete command uses approximately
0.50 s CPU/wall. `time -p make --just-print test-run-status` with the same exact
run selection uses 0.41 s (0.18 user/0.23 system), before reader execution. A
read-only process-family observation of that dry run observed 92 Bash and 87
date processes: the recursively expanded, exported default run-ID expression
repeatedly forks date. F-022 therefore needs a Make identity-resolution repair
in addition to reader/runtime work.

**F-024: read-only preflight can change result-root permissions.** Static review
found `validateResultRootSecurity` calling `secureMkdir` even without a creation
request. Repair validation to be non-mutating and leave permission changes at
explicit artifact creation. Preserve symlink/ownership rejection and add an
independent permission-preservation fixture. This belongs to the S-016/S-020
reader boundary and must pass before another freeze. S-020 is reopened; S-022
remains gated. Candidate 4 and its evidence are unchanged.


### S-020 Make identity and read-only preflight repair checkpoint — DONE

The default Make invocation identity is now evaluated exactly once, preserving
explicit/inherited identities while removing repeated date/shell forks. A
Make-owned fixture checks identity equality across multiple child environments
and explicit/inherited precedence. Read-only result-root validation no longer
creates directories or changes modes; explicit creation still secures its leaf.
It rejects symlink ancestors, including dangling links, and retains ownership
checks. Independent fixtures verify mode/inode/ownership/mtime/ctime preservation
and explicit creation. The common contract facade exports the validation boundary.

Primary executable source is `sha256:a7b3c6d8cd68bfc5f7587c314f7b95a3db3ddb4b649ce064ed55d8127bebb773`; primary `main` and
the staged tracker fingerprint remain unchanged. No published schema acceptance,
qualification threshold, workload or reader schedule changed. Exact commands
(all from the primary root) and outcomes:

| Invocation ID | Public Make command | Result |
| --- | --- | --- |
| `instrumentation-s021-repair-command-04` | `make harness-command-surface-contract` | FAILED: new fixture imported a not-yet-exported facade member. Preserved stdout identifies the module-link failure; related, corrected by exporting the existing boundary. |
| `instrumentation-s021-repair-command-05` | `make harness-command-surface-contract` | PASS, public exit 0, including identity and non-mutating validation fixtures. |
| `instrumentation-s021-repair-evidence-03` | `make harness-evidence-contract` | PASS, public exit 0, retained-reader and accounting coverage. |
| `instrumentation-s021-repair-scripts-06` | `make lint-scripts` | PASS, 2/2 units, 768 ms, public exit 0. |

Each command uses `CARTULARY_TEST_RUN_ID=<Invocation ID>`; result roots are
`.cartulary/test-results/<Invocation ID>`. The failed command's log is
`harness-command-surface-contract/harness-command-surface-contract/stdout.log`.

Diagnostic `time -p make --just-print test-run-status
RESULTS_DIR=.cartulary/test-results/instrumentation-s021-repair-native-04 JSON=1`
reports 0.01 s versus the pre-repair 0.41 s. The actual same-selection public
status command reports 0.07 s wall, 0.05 user and 0.02 system. These single-run
diagnostics are not qualification and imply no budget pass. Node reader and
collector/admission/receipt startup remain possible costs. F-024 is closed;
F-022 remains open. Candidate 5 can now freeze this repair and remeasure.
No S-022 finalization, broad final acceptance or completion is claimed.


Candidate 5 is frozen at `0b158001bf4ea4e2ffa455edddba0cf3a32bc335` on
`codex/instrumentation-qualification-05` in the existing clean isolated checkout.
An alternate Git index preserved primary HEAD/index and excluded the tracker.
Collection uses `CARTULARY_TEST_RUN_ID=instrumentation-s021-qualification-05
make harness-instrumentation-qualify
QUALIFICATION_PLAN=tools/harness_instrumentation_qualification_plan.json`.
The manifest will provide the actual clean-checkout source digest.


### S-021 candidate 5 rejection and required repair handoff — IN_PROGRESS

This is a failed phase checkpoint, **not** the S-022 completion checkpoint.
The clean frozen executable is `0b158001bf4ea4e2ffa455edddba0cf3a32bc335`,
source `sha256:cea8faf2dc7f230a47b3ee813a1a11854c070540a50dac6a6c20dfcce948cc51`.
It is retained on `codex/instrumentation-qualification-05` in
`/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary`.
**No revision has qualified.** Primary executable content remains at the prior
repair checkpoint; its different file-mode-sensitive source digest is recorded
there. Later tracker edits do not relabel either identity.

The public qualification command recorded above ran in native session 32587.
Its retained root is the isolated checkout's
`.cartulary/test-results/instrumentation-s021-qualification-05/qualification`.
It published 45 attempts: 42 Go (cold, warm-up, five measured blocks for every
treatment) and three Vitest cold attempts. Every published attempt has complete
measurement coverage and no measurement failure. Go's five-block assessment
is stable, so the unchanged policy did not request a sixth block.

After the complete Go cohort failed, the superseded candidate was stopped via
SIGINT to its verified native session (outer exit 130). The unlisted interrupted
`runs/qualification-vitest-cold-0-off-4` directory is retained without a completed
result. Exact public `make test-run-status RESULTS_DIR=<that exact run> JSON=1`
returned `not_published`, exit 0; that observation is not cleanup or acceptance
proof. A subsequent read-only process listing showed no qualification runner,
native meter or gate helper remaining. No service/browser attempt started, no
service resources were manually removed, and no valid unfavorable observation
was excluded. The qualification manifest remains `incomplete`.

Independent public validation was:

```sh
CARTULARY_TEST_RUN_ID=instrumentation-s021-validation-05 make harness-performance-check EVIDENCE_ROOTS_FILE=/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary/.cartulary/test-results/instrumentation-s021-qualification-05/qualification/qualification.json
```

The validator rejected the manifest (validator exit 13; outer Make exit 2).
All non-Go cohorts are incomplete/ineligible. The complete Go cells are eligible
but fail as follows; off/0 is the zero-cost reference, not an observer pass:

| Treatment | CPU conservative upper, % of one core | Additional RSS upper, MiB | Finalization upper, ms | Outer median / p90 change, % |
| --- | --- | --- | --- | --- |
| off / 0 | 0 | 0 | 0 | 0 / 0 |
| off / 1 | 3.5184 | 6.1945 | 0 | 0.5687 / 1.4534 |
| off / 4 | 8.4415 | 12.166 | 0 | 0.5372 / 2.1401 |
| basic / 0 | 4.7817 | 22.513 | 118.48 | 2.4572 / 3.5841 |
| basic / 1 | 8.4849 | 29.238 | 125.23 | 2.1820 / 3.7535 |
| basic / 4 | 11.535 | 27.399 | 125.51 | 2.3168 / 3.3824 |

CPU/RSS/finalization values are the public validator's printed results. Outer
percentages were independently reconstructed from the five retained paired
measurements using the median and nearest-rank p90. Canonical elapsed gates,
RSS and finalization pass for this cohort. CPU exceeds 0.5% for every observer
treatment; basic outer elapsed exceeds the 2% median / 3% p90 gates. Reader CPU
alone is approximately 67 ms for one reader and 277 ms for four in the collected
attempts. The Make identity repair materially reduced cost, but did not qualify
the complete observer. No unsupported attribution of workload variance to a
specific helper is made.

**F-022 remains unresolved; S-020 performance repair is reopened.** The next
implementation must reduce the complete public reader and basic diagnostics
execution cost, including startup, admission transactions and receipt
publication. Preserve one validation owner, strict parsing/security and the
observation-gate boundary; do not add unchecked cached outputs, weaken schema
validation, classify observer work as measurement support, or alter workloads,
reader cadence or budgets. A native implementation is a possible structural
repair to evaluate against these ownership and maintenance constraints; no such
port has been implemented or validated at this checkpoint. Use profiling and
independent accounting to establish the next repair, validate it in its owning
workstream, save the tracker, then freeze a new clean candidate and rerun every
affected cohort. All previous candidates and failed attempts remain retained.

The latest repair's `harness-command-surface-contract`,
`harness-evidence-contract` and `lint-scripts` passed as recorded above.
`git diff --check` and `git diff --cached --check` passed after the repair.
Primary HEAD remains `1d423c6e5f9cd204f8210fa9958c780fd933d293`; the pre-existing
staged tracker SHA-256 is still
`9ab829e44fa8df7ba1ee86e794518acd97129cea6af416966fbdda4b33f83426`.

S-022 is deliberately not entered: the user's dependency rule blocks it while
S-021 fails. Therefore its `agent-finalize`, final broad verification,
tracker-to-Markdown-lint registration and final handoff completion remain
outstanding. No eligible successful full warm `check` root was selected and no
retained-run maintenance is claimed. T-007/T-008, F-017 and RB-006 remain open;
F-014–F-016/F-018 have functional evidence but overall completion still requires
the final qualified validation/handoff. F-019's final readiness reconciliation
also remains contingent on that checkpoint.

Operational containment remains `HARNESS_DIAGNOSTICS=off` by default. Optional
resource evidence stays local-only, WSL2 coverage excludes Windows-host and
Docker-daemon-internal cost, sampled memory is not an unsampled peak guarantee,
and no Core 05 publication claim is made. Rollback, if selected, must restore
one coherent authored/generated/producer/reader transaction while preserving
historical evidence and the user's staged work; default-off alone does not
close F-022 or qualification.


### S-020 F-022 continuation entry — IN_PROGRESS

The user explicitly requested continued structural repair and completion of
S-020–S-022 with unchanged budgets. Primary HEAD/index and candidate 5 evidence
were rechecked and remain intact. The next repair begins with a native public
reader and shared native schema projection compiled from the existing typed
foundation schemas. This avoids a second authored acceptance registry. Unsupported
schema vocabulary must stop generation, and differential/fault fixtures must
confirm parity before cutting over the public target. No new candidate is frozen.

The user also asked to consider a RAM disk. Read-only filesystem checks show
`/tmp` already uses tmpfs (16 GiB capacity); `/dev/shm` is an unused, executable
16 GiB tmpfs, while the current checkout/results use ext4. Available guest
memory was about 27 GiB. A matched RAM-backed preparation/storage experiment
can distinguish filesystem effects from startup CPU; it cannot replace the
native-reader cost repair or alter thresholds. Any selected storage posture
must be identical across treatments, recorded with its evidence, and durable
copies of all results must be preserved before releasing RAM-backed storage.

### S-020 F-022 native-reader repair checkpoint — IN_PROGRESS

The replacement public reader uses `internal/testutil/harnessdiag` and a thin
`tools/harnessdiag` composition root. Foundation JSON Schemas generate native
predicates; the authored task input contract generates its native input policy.
Make owns installation through bootstrap/verification preparation. The observer
requires that installed tool and initializes no writer identity, compiler or
cache. It preserves exact selection, bounded waiting, secure descriptor-relative
reads, strict JSON, schema closure, live membership validation, terminal precedence,
closed output, and read-only interruption. There is no qualifying performance
claim yet. Basic collection, admission and terminal-publication cost still require
repair. The former JavaScript reader remains temporarily as a differential test
reference pending removal of its obsolete public path.

Primary HEAD remains `1d423c6e5f9cd204f8210fa9958c780fd933d293`. These are dirty-tree
functional checks, not a frozen executable candidate. Generated native files are
registered with generated-artifact policy and drift inputs; no schema acceptance
set or observer budget changed. The pre-existing staged tracker remains separate.
All result roots below are under the primary `.cartulary/test-results/` directory.

| Run ID / exact public command | Outcome and limitation |
| --- | --- |
| `instrumentation-f022-native-generate-01`: `make generate` | Failed before publication: native projection lacked `maxProperties`. No retained run summary was produced at the prerequisite boundary. |
| `instrumentation-f022-native-generate-02`: `make generate` | PASS after adding property bounds and Boolean-schema handling. |
| `instrumentation-f022-native-generate-03`: `make generate` | PASS after native command-policy and public recipe projection. |
| `instrumentation-f022-native-command-01`: `make harness-command-surface-contract` | FAILED native build: generated empty-object branch had an unused variable. Retained under `harness-diag-build/`; corrected in the generator. |
| `instrumentation-f022-native-generate-04`: `make generate` | FAILED authored routing: live-observation verification did not yet admit Go evidence. Added the native owner's Go evidence kind; no row was reassigned to an unrelated obligation. |
| `instrumentation-f022-native-generate-05`: `make generate` | PASS after routing repair and bootstrap readiness wiring. |
| `instrumentation-f022-native-command-02`: `make harness-command-surface-contract` | PASS, including unknown inputs, inherited-selection isolation, missing installed tool, and sole-goal enforcement. |
| `instrumentation-f022-native-format-01`: `make format-go` | PASS; only authored Go sources are formatting inputs. |
| `instrumentation-f022-native-tests-01`: `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_observation` | PASS, 1/1 unit, 6440 ms; strict-number/duplicate/surrogate rejection, schema closure, missing exact selection, cancellation, symlink/ancestor/hardlink and size-bound cases. |
| `instrumentation-f022-native-evidence-01`: `make harness-evidence-contract` | FAILED differential validation: native failure output exposed partial sample identity. Fixed by admitting a complete sample only after validation; no malformed sample facts escape. |
| `instrumentation-f022-native-evidence-02`: `make harness-evidence-contract` | FAILED only generator-transaction fixture setup: its isolated root omitted the new typed task-owner input. Native/JavaScript reader comparisons passed. |
| `instrumentation-f022-native-generate-06`: `make generate` | PASS after separating read-only Make initialization from writer configuration. |
| `instrumentation-f022-native-evidence-03`: `make harness-evidence-contract` | PASS after repairing the fixture input roster; differential and hostile-evidence checks pass. |

The user-requested RAM-disk assessment is retained at
`.cartulary/test-results/instrumentation-f022-ram-probe-01/`. It contains the
storage/selection declaration, every alternating timing observation, copied
artifact digests, and a durable copy of the RAM-backed retained fixture. This is
an external component probe of the same native public `make test-run-status
RESULTS_DIR=<exact> JSON=1`, not engineering qualification. Eight alternating
blocks include one declared warm-up; seven measured blocks gave ext4 median
wall/CPU 11.546/8.856 ms and tmpfs 10.911/9.058 ms. No CPU improvement was observed.
The RAM location is
`/dev/shm/cartulary-instrumentation-f022-ram-probe-01/`; no failed qualification
evidence was moved or removed. Keep ext4 qualification posture unless later
matched evidence establishes a reason to change it. CPU cost must be reduced
structurally; a storage change cannot be used to relabel the failed candidates.

S-020 remains open for complete-observer repair; S-021 remains failed/incomplete
and S-022 has not begun. All candidate-1–5 evidence and exclusions remain intact.


### S-020 F-022 native observer continuation — IN_PROGRESS

Native public reading, terminal receipt publication, and host admission now use
one generated schema/input-policy boundary and one weighted admission owner.
The native collector replaces the JavaScript worker without changing cadence,
limits, resource scopes, quietness, lifetime ownership or qualification budgets.
It reuses the admission owner in process; kernel notifications wake it after an
exclusive measurement. A pinned collection thread owns bounded streaming and
publication. JavaScript heap is truly absent (zero), while complete native heap,
other threads and helper cost remain covered by external qualification. REQ-833
now names the already-published v3 component fields and their actual semantics;
no schema acceptance set changed. The obsolete JavaScript reader, collector,
receipt entrypoint and admission implementation have been removed. Native fault
coverage is being completed before refreezing; this is not qualification success.

All roots below are under primary `.cartulary/test-results/`; commands use the
explicit `CARTULARY_TEST_RUN_ID` shown. Historical failures remain retained.

| Run ID suffix after `instrumentation-f022-` | Exact command and outcome |
| --- | --- |
| `native-evidence-04` | `make harness-evidence-contract`: PASS, 18050 ms; native receipt parity. |
| `native-generate-07` | `make generate`: FAILED unsorted authored test selection. |
| `native-tests-02` | `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_observation`: FAILED graph admission before publication following the routing failure; no retained summary. |
| `native-generate-08` | `make generate`: PASS, 11190 ms after sorting routing. |
| `native-tests-03` | Same exact native-observation slice: PASS 1/1, 2796 ms, including AJV-authored native schema parity probes. |
| `native-generate-09` | `make generate`: PASS, 11200 ms; minimal native Make route. |
| `native-generate-10` | `make generate`: PASS, 11060 ms after removing the obsolete JavaScript reader. |
| `native-command-03` | `make harness-command-surface-contract`: PASS, 62400 ms. |
| `native-generate-11` | `make generate`: FAILED stale admission backing-file roster and impermissible implicit tool installation for UI review; removed installation from those read-only flows and updated the roster. |
| `native-format-02` | `make format-go`: PASS. |
| `native-generate-12` | `make generate`: FAILED generated Make bindings exceeded the existing 184320-byte bound. Factored the shared recursive prerequisite recipe; did not change the bound. |
| `native-generate-13` | `make generate`: PASS, 11530 ms. |
| `native-evidence-05` | `make harness-evidence-contract`: PASS 110/110, 19050 ms; native admission and receipt. |
| `native-generate-14` | `make generate`: PASS, 11490 ms; admission row routing. |
| `native-tests-04` | `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_observation,harness.evidence_accounting.behavior.observation_admission`: FAILED 1/2 units, 2919 ms. New admission fixtures incorrectly supplied a non-private Go temporary root; fixed fixtures to request their owned private subdirectory. Strict production permissions remain unchanged. |
| `native-format-03` | `make format-go`: PASS. |
| `native-evidence-06` | `make harness-evidence-contract`: PASS 110/110, 19090 ms, including four actual native collectors across quietness/pause/resume. |
| `native-format-04` | `make format-go`: PASS. |
| `native-generate-15` | `make generate`: PASS, 11900 ms; native collection and receipt fault row routing. |

Two diagnostic attempts to select the tool-only `native-evidence-05` root without
`TARGET=harness-evidence-contract` returned configuration errors; the exact
published target summary establishes its pass. A graph invocation was then
selected successfully by the public native reader. No diagnostic failure was
relabeled as workload failure. Current executable work is dirty-tree functional
validation. Candidate 5 remains the latest frozen failed qualification; S-020
and F-022 stay open, and S-021/S-022 cannot be marked complete from these checks.


### S-020 F-022 native observer validation ledger — IN_PROGRESS

Further native validation retained the following attempts. All roots have prefix
`.cartulary/test-results/instrumentation-f022-` in the primary checkout. The table
uses the suffix as the explicit `CARTULARY_TEST_RUN_ID` suffix; source identities
remain in each exact run manifest. Functional checks use dirty source and cannot
establish S-021 qualification.

| Suffix | Exact public command and result |
| --- | --- |
| `native-tests-05` | `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_collection,harness.evidence_accounting.behavior.native_observation,harness.evidence_accounting.behavior.observation_admission`: PASS 3/3, 3257 ms. |
| `native-evidence-07` | `make harness-evidence-contract`: PASS, 19040 ms. |
| `native-format-05` through `native-format-11` | Each `make format-go`: PASS; authored native code only. |
| `native-generate-16` through `native-generate-20` | Each `make generate`: PASS; respective durations 10620, 10800, 10850, 11080, 11270 ms. |
| `native-tests-06` | Same three-row native slice: PASS 3/3, 3092 ms. |
| `native-command-04` | `make harness-command-surface-contract`: FAILED 44/45 cases. Native Make `unexport` changed unset cache input origins into empty declared values. Fixed the generator to capture origins before unexport, preserving explicit invalid-cache rejection and inherited unrelated-input isolation. |
| `native-tests-07` | Same three-row native slice with `HARNESS_DIAGNOSTICS=basic`: PASS 3/3, 3192 ms. |
| `native-command-05` | `make harness-command-surface-contract`: PASS, 65850 ms. |
| `native-evidence-08` | `make harness-evidence-contract`: PASS, 19390 ms. |
| `native-lint-01`, `native-lint-02`, `native-lint-03` | Each `make lint-scripts`: PASS 2/2; 792, 815, 777 ms. |
| `native-service-01` | `make service-backed-test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.schema_bootstrap HARNESS_DIAGNOSTICS=basic`: workload PASS 3/3, 59446 ms; **not valid evidence of resource collection**, as explained below. |
| `native-browser-01` | `make service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.https_trust HARNESS_DIAGNOSTICS=basic`: workload PASS 11/11, 68747 ms; **not valid evidence of resource collection**. |
| `native-observability-01` through `native-observability-03` | `make harness-observability-check RESULTS_DIR=<exact absolute root>` on `native-tests-07`, `native-service-01`, `native-browser-01`: PASS integrity/closure with partial diagnostics. These audits do not prove successful sampling. |
| `native-tests-08` | Same three-row native slice with `HARNESS_DIAGNOSTICS=basic`: PASS 3/3, 3061 ms; includes actual native malformed-container transport/timeout termination and secure storage failure cases. |
| `native-lifecycle-01` | `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.testservices_lifecycle_contract`: PASS 1/1, 5187 ms. |
| `native-drift-01` | `make generate-drift`: PASS 4/4, 24360 ms. |
| `native-policy-01` | `make generated-artifact-policy-check`: PASS 3/3, 2837 ms. |
| `native-shapes-01` | `make json-shape-check`: PASS 3/3, 5090 ms. |
| `native-evidence-09` | `make harness-evidence-contract`: FAILED 96/97 cases after strengthening treatment eligibility; the formerly successful synthetic basic qualification fixture omitted resource publication. Added an actual closed empty resource index/stream to valid fixtures and a missing-index rejection case. |
| `native-evidence-10` | `make harness-evidence-contract`: PASS, 19280 ms. |
| `native-ui-lifecycle-01` | `make harness-ui-review-lifecycle`: PASS, 68520 ms; shared native host admission retains UI lifecycle behavior. |
| `native-evidence-11` | `make harness-evidence-contract`: FAILED shared immutable-object traversal after exposing validated resource-index facts. Replaced repeated freezing with one memoized object/map traversal. |
| `native-service-02` | Same service-backed database slice with basic: workload PASS 3/3, 58131 ms; successful actual collection after the capacity projection repair. |
| `native-browser-02` | Same browser HTTPS trust slice with basic: FAILED artifact admission, 6 passed/2 failed/3 skipped, 40628 ms. `target-build-web` reports frontend build inputs changed after the run source snapshot while repairs continued. Cleanup reports 10 steps, 0 failures/blocked. This attempt is retained and is not qualifying evidence. |
| `native-observability-04` | `make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/instrumentation-f022-native-service-02`: PASS, complete execution and diagnostic closure. |

Exact human/JSON performance and resource views of `native-service-01` were
written through `make explain-run RESULTS_DIR=<exact absolute root>
DETAIL=performance|resources [JSON=1]` under `native-views-01/`. They exposed the
native capacity-domain defect: all graph resource dimensions had been passed to
host admission instead of its five host dimensions. Every sweep failed before
sampling; retained integrity alone was therefore insufficient. The native
boundary now projects those dimensions as the existing host owner does, and the
real four-collector quietness fixture includes graph-only dimensions so this
cannot recur unnoticed. `native-service-02` now retains 268 samples, two exact
owned container identities, no container omissions/overruns, and complete index
status; its additional public JSON view is
`native-views-01/resources-service-02.json`.

The S-016 validated model now also returns the already-validated resource index
without a second stream traversal. Qualification's F-017 treatment-completeness
check rejects missing/failed collectors and failed sweeps/container overruns,
even if the external meter and receipt otherwise close. This tightens eligibility
without changing any budget or published schema acceptance set. An empty,
complete stream remains valid for an invocation shorter than its sampling cadence.

Other structural corrections include ordered execution-v2 relationship hashing
(with an independently calculated literal test digest), failure-aware staging
cleanup, a private fixed-size native control protocol, periodic flushing through
the same admission fence, and removal of the obsolete JavaScript facade roster.
The worker fault tests now terminate the actual native collector and verify that
the parent drains its proven container helper.

`make task-guide ROLE=module-author OWNER=harness.testservices` was rejected:
that owner does not exist. Discovery identified the lifecycle row under
`harness.browser`; its task-guide route and exact selected row then passed.
`native-tests-08` source is
`sha256:e2574b9b5ce0300ff556286b9c6a0be57c60bfaa166b4e2e3f2107cbd155c5a8`;
`native-service-01`/`native-browser-01` source is
`sha256:691354fd7c01f6c058610718332cd36b32329331bdf65243f7a836a42bba2cae`;
`native-service-02`/`native-browser-02` source is
`sha256:1ffb6f0f4521d45bfda32f58ca16c97a886e4e5bac573b71ef518dd3b2dddcae`.
These identities are not interchangeable with the next clean qualification source.

The pre-existing staged tracker SHA-256 remains
`9ab829e44fa8df7ba1ee86e794518acd97129cea6af416966fbdda4b33f83426`;
primary HEAD remains unchanged. No historical failed evidence was removed.
S-020 awaits its final functional checkpoint; F-022/S-021 still require the full
unchanged matrix, and S-022 remains dependent on that result.


### S-020 native observer repair completion checkpoint — DONE

The complete observer's functional repair is ready for S-021. This checkpoint
closes the reopened S-020 workstream before beginning qualification. F-022 itself
remains open until every required cell passes the unchanged hard gates.

`instrumentation-f022-native-evidence-12` (`make harness-evidence-contract`)
passes, 19180 ms, including the unified immutable model, treatment-completeness
rejections, native worker death, host quietness, native reader and constrained
receipt tests. `instrumentation-f022-native-browser-03` repeats the exact HTTPS
trust service-backed selection with basic diagnostics and passes 11/11 units,
58008 ms. Its retained resource index is complete: 419 samples, two exact owned
container identities, zero container omissions and deadline overruns. The public
`harness-observability-check` (`native-observability-05`) passes complete resource
and execution closure. The browser JSON resource explanation is retained under
`native-views-01/resources-browser-03.json`. Fresh human/JSON performance and
resource explanations of `native-service-02` also pass and are retained there.
The failed browser-02 run and failed evidence-09/11 checks remain preserved.

Earlier functional gates in the preceding ledger also pass: all native owner
rows, testservices lifecycle, public command surface, shared UI lifecycle,
script lint, generation, drift, generated-artifact policy and JSON shapes.
The last repairs corrected graph-to-host capacity projection, shared immutable
facts, and eligibility of the actual observer treatment. They changed no public
schema acceptance set, cadence, workload selection, preparation policy or budget.
They do not alter canonical timing, primary-failure precedence, fixture ownership,
cleanup or explicit export boundaries.

Qualification must now freeze all executable changes together in the existing
isolated checkout and rerun all five workloads × off/basic × 0/1/4 readers.
Keep seed 20261009, 30-second reader cadence, cold and discarded warm-up attempts,
five measured blocks and the existing sixth-block rule. Ext4 storage remains
selected; the recorded RAM probe did not establish a CPU benefit. Previous
candidate-1–5 evidence, source identities, exclusions and failures are retained.
No result from a dirty-tree functional run is a qualification pass. Primary HEAD
and the user's staged tracker remain untouched. S-022 is still gated by S-021.


### S-021 candidate 6 freeze and dispatch — IN_PROGRESS

After the saved S-020 completion checkpoint, executable work was frozen using an
alternate Git index, preserving primary HEAD and the pre-existing staged tracker.
The clean isolated checkout is
`/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary`, branch
`codex/instrumentation-qualification-06`, commit
`6665c50215814c2e2181dc3e2cac4b9e52141975`, executable source digest
`sha256:f3c34b856237ce154499b299ca87044eec6d2238801ceb21e81ce01ca1a93f54`.
This is a new candidate, not a relabeling of prior evidence.

Dispatched there:
`CARTULARY_TEST_RUN_ID=instrumentation-s021-qualification-06 make harness-instrumentation-qualify QUALIFICATION_PLAN=tools/harness_instrumentation_qualification_plan.json`.
Native execution session: `56107`. The sole candidate manifest is
`.cartulary/test-results/instrumentation-s021-qualification-06/qualification/qualification.json`.
The tool owns preparation, the randomized matrix, supervision, cleanup and result
assembly. No concurrent repository verification is running during measurement.
Initial cold attempts have complete external coverage, but cold observations do
not establish budget success. All gates and treatment/preparation inputs remain
unchanged. S-021 and F-022 remain open; S-022 has not begun.


### S-021 candidate 6 rejection; S-020 reopened — IN_PROGRESS

Candidate 6 remains frozen at `6665c50215814c2e2181dc3e2cac4b9e52141975`,
source `sha256:f3c34b856237ce154499b299ca87044eec6d2238801ceb21e81ce01ca1a93f54`.
The unchanged policy required a sixth Go block. The resulting Go cohort is
eligible but unstable and exceeds CPU/elapsed gates. After this completed
cohort failed, the verified qualification process group received SIGINT; native
session 56107 exited 130. All 53 published attempts remain: 48 Go and five Vitest
cold attempts, including cancelled `vitest-cold-0-basic-0` (137). The manifest is
`incomplete`; no service/browser cohort started. A subsequent process listing
found no qualification runner, gate helper or measurement helper. No unfavorable
valid observation was excluded and no retained evidence was removed.

Independent validation in the isolated checkout:
`CARTULARY_TEST_RUN_ID=instrumentation-s021-validation-06 make harness-performance-check EVIDENCE_ROOTS_FILE=/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary/.cartulary/test-results/instrumentation-s021-qualification-06/qualification/qualification.json`.
Validator exit 13; outer Make exit 2. Result roots remain under that checkout's
`.cartulary/test-results/instrumentation-s021-qualification-06` and
`.cartulary/test-results/instrumentation-s021-validation-06`.

| Go treatment | CPU upper, % one core | RSS upper, MiB | Finalization upper, ms | Outer median / p90, % |
| --- | --- | --- | --- | --- |
| off / 0 | 0 | 0 | 0 | 0 / 0 |
| off / 1 | 1.7428 | 7.3162 | 0 | 0.7771 / 1.5134 |
| off / 4 | 1.6427 | 4.2265 | 0 | -0.0892 / 2.9254 |
| basic / 0 | 2.2021 | 14.938 | 86.804 | 1.4359 / 3.9250 |
| basic / 1 | 3.8116 | 12.16 | 76.040 | 2.0923 / 8.6355 |
| basic / 4 | 3.5534 | 23.22 | 95.419 | 1.7801 / 3.4134 |

CPU/RSS/finalization are the public validator's results; outer statistics were
independently reconstructed from the retained six paired blocks. Native readers
reduced their measured CPU to approximately 12 ms per request, but this is still
material against the complete observer budget. Sampling and wrapper costs also
remain. These data do not isolate workload variance from observer CPU.

S-020 is reopened for a structural cost repair before another clean freeze.
F-022 remains open, S-021 incomplete and S-022 not started. Investigate native
process startup, shared validator dependencies and Make dispatch overhead;
retain strict validation and full accounting. RAM storage remains a rejected
performance assumption: the preserved component probe showed no CPU benefit.
No budget, workload, cadence, preparation or schema acceptance change is adopted.


### S-020 candidate 7 preparation ledger — IN_PROGRESS

The next repair projects simple anchored ASCII token patterns into generated
byte predicates; other patterns retain the existing closed RE2 path. Independent
AJV parity now covers all authored patterns with ASCII, Unicode and length-edge
probes. Native command token checks have an independent regexp oracle. The
collector receives its own composition binary, separating its network/runtime
imports from the public reader. Make's generated native route executes a simple
command directly while retaining missing-tool failure and input validation.
Secure local directory traversal opens existing ancestors directly; it creates
only missing components and retains descriptor anchoring, no-follow opens and
private-leaf checks. These changes preserve schema acceptance and publication
security; they do not adopt a cache of unvalidated evidence.

Primary WSL2 public validation so far (all roots under `.cartulary/test-results`):

| Exact command after `CARTULARY_TEST_RUN_ID=` | Result root suffix | Outcome |
| --- | --- | --- |
| `make generate` | `instrumentation-f022-token-generate-01` | PASS, 11240 ms |
| `make format` | `instrumentation-f022-token-format-01` | PASS, 8431 ms |
| `make generate` | `instrumentation-f022-token-generate-02` | PASS, 11050 ms |
| `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_observation,harness.evidence_accounting.behavior.native_collection` | `instrumentation-f022-token-tests-01` | PASS, 2/2, 3016 ms |
| `make harness-evidence-contract` | `instrumentation-f022-token-evidence-01` | PASS |
| `make generate` | `instrumentation-f022-token-generate-03` | PASS, 11080 ms; subsequent status dry-run exposed recipe ordering bug |
| `make format` | `instrumentation-f022-token-format-02` | PASS, 8671 ms |
| `make -n test-run-status RESULTS_DIR=/tmp/absent JSON=1` | no retained run | FAIL, Make 2: recipe preceded its target after a target-specific assignment |
| `make generate` | `instrumentation-f022-token-generate-04` | PASS, 11030 ms; status dry-run subsequently PASS after moving assignment before target header |

All component probes invoke public `make --no-print-directory test-run-status
RESULTS_DIR=<exact root> JSON=1`; none qualifies the complete observer. Roots
`instrumentation-f022-reader-probe-02` through `-05` retain every observation.
Probe 02 included default/serial-runtime/dry-run dispatch comparisons and Go
initialization tracing. Probes 03/04 accidentally reused the previously installed
binary: owner Go tests compile their test executable, not the installed public
reader. Their apparent timing difference is **not** evidence for the repair.
After the public evidence target rebuilt the binary, probe 05 recorded its binary
digest and approximately 8.4 ms median terminal-reader CPU (one declared warm-up,
11 measured requests). This remains a component diagnostic, not qualification.
Candidate 6 and all earlier failed attempts remain untouched. S-020 stays open
pending binary-split, filesystem, public command and projection validation.


Candidate 7 preparation update: public `harness-evidence-contract` with
`CARTULARY_TEST_RUN_ID=instrumentation-f022-token-evidence-02` passed after the
binary split and directory traversal repair. Probe 06 retains its binary digest
and all 24 public requests; median terminal CPU was approximately 6.9 ms.
This was a component probe during command-contract preparation, so it is not
paired workload evidence. Budget compliance remains unestablished.


Temporary profiling was built only in the primary checkout through
`CARTULARY_TEST_RUN_ID=instrumentation-f022-profile-build-01 make harness-evidence-contract`
(PASS). It measured native CPU before entry and within the request, activated by
a temporary private probe environment variable. Public requests are retained in
`instrumentation-f022-reader-profile-01`; request processing itself was roughly
0.4–0.7 ms CPU, with most process CPU already accumulated before native entry.
The profile does not distinguish kernel execution/startup from earlier exec
stages, and changing native runtime parallelism did not remove that cost. The
temporary probe code was removed before further validation/freeze; it introduces
no supported command or input. This limits attribution rather than establishing
an overhead pass. The binary split and schema/IO repairs now proceed to complete
functional validation and a fresh qualification attempt.


Candidate 7 validation continuation (primary WSL2; exact commands use the listed
`CARTULARY_TEST_RUN_ID`, and roots are `.cartulary/test-results/<id>`):

| ID | Command | Outcome |
| --- | --- | --- |
| `instrumentation-f022-token-command-01` | `make harness-command-surface-contract` | PASS, 62520 ms |
| `instrumentation-f022-token-evidence-03` | `make harness-evidence-contract` | PASS after removing temporary profiling |
| `instrumentation-f022-token-basic-01` | `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_observation,harness.evidence_accounting.behavior.native_collection,harness.evidence_accounting.behavior.local_migration_activities HARNESS_DIAGNOSTICS=basic` | PASS, 3/3, 2816 ms; complete resources, one sweep, ten samples, zero omissions/failures |
| `instrumentation-f022-token-audit-01` | `make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/instrumentation-f022-token-basic-01` | PASS, complete diagnostics/execution |
| `instrumentation-f022-token-browser-01` | `make service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.https_trust HARNESS_DIAGNOSTICS=basic` | PASS, 11/11, 65017 ms; 32 sweeps/463 samples, two owned containers, zero omissions/overruns/failed sweeps |
| `instrumentation-f022-token-audit-02` | `make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/instrumentation-f022-token-browser-01` | PASS, complete diagnostics/execution |
| `instrumentation-f022-token-drift-01` | `make generate-drift` | PASS, 4/4, 20089 ms |
| `instrumentation-f022-token-policy-01` | `make generated-artifact-policy-check` | PASS, 3/3, 2584 ms |
| `instrumentation-f022-token-generate-05` | `make generate` | PASS, 11510 ms; projects direct target-specific exports |

The browser run's executable source is
`sha256:47184c2b451795d4b39ad7e1ed902b15e3fd7b23bff0029af6b404c531d2e86e`.
Its human/JSON performance and resource explanations all passed via public
`make explain-run RESULTS_DIR=<that exact root> DETAIL=performance|resources
[JSON=1]`; exact argv and outputs are in `instrumentation-f022-token-views-01`.
This predates the final native Make dispatch change and is not relabeled as
qualification evidence.

The final dispatch repair uses target-specific exports of literal input values,
with source classification captured first, then invokes the installed binary
without an intermediate shell or `env` process. It retains the same generated
input policy. Probe 07 (saved binary digest, all 24 observations, one discarded
warm-up per selection) measured approximately 4.06 ms terminal-reader CPU. These
are component diagnostics only. The primary staged tracker hash still matches
`9ab829e44fa8df7ba1ee86e794518acd97129cea6af416966fbdda4b33f83426` and primary HEAD
is unchanged. Further command/projection gates must pass before freezing.


### S-020 candidate 7 structural repair completion checkpoint — DONE

This checkpoint closes the reopened implementation workstream before its
successor begins. It does **not** close F-022 or establish qualification. Final
primary executable source is
`sha256:aa969a3672f401202517ec5a45d642f513c8978eaa7f9904c311b208b4bf0a02`
(dirty checkout at unchanged `main` HEAD). A clean freeze is required next.

The completed change projects token validation without per-request automaton
construction, separates collection from the public reader executable, uses
literal target-specific Make exports/direct dispatch, and removes repeated
creation attempts/stat calls from secure directory traversal. Strict parsing,
closed schemas, no-follow/ownership/link checks, immutable publication, signal
handling, observation quietness and total external accounting are preserved.
A new installed `cartulary-harness-collect` binary is prepared by the existing
public bootstrap/contract/qualification prerequisites. Existing installations
must rebuild through those Make routes; the read-only status target never
provisions tools. No schema acceptance, workload, preparation, cadence, capacity
or budget changes were made. Temporary profiling code is absent from source.

Final public gates (primary WSL2, exact `CARTULARY_TEST_RUN_ID` and root suffix):

| ID | Exact command | Result |
| --- | --- | --- |
| `instrumentation-f022-token-command-02` | `make harness-command-surface-contract` | PASS, 65080 ms |
| `instrumentation-f022-token-command-03` | `make harness-command-surface-contract` | PASS, 62380 ms; includes literal Make/shell input non-execution regression |
| `instrumentation-f022-token-evidence-04` | `make harness-evidence-contract` | PASS, 18820 ms |
| `instrumentation-f022-token-drift-02` | `make generate-drift` | PASS, 4/4, 19638 ms |
| `instrumentation-f022-token-policy-02` | `make generated-artifact-policy-check` | PASS, 3/3, 2521 ms |
| `instrumentation-f022-token-shapes-01` | `make json-shape-check` | PASS, 3/3, 4783 ms |
| `instrumentation-f022-token-lint-01` | `make lint-scripts` | PASS, 2/2, 661 ms |

Native parity/security/collector tests, diagnostics-enabled migration checks,
real browser/service observations and exact-run audits/explanations are recorded
above. Earlier source identities and failures remain distinct. Both Git diff
checks passed. The staged tracker hash remains the entry hash and no user-staged
work was changed. Rollback must restore authored inputs, generated predicates/
recipes, both native composition roots and their callers together; never delete
failed retained evidence or translate obsolete schemas. Default-off remains an
operational containment option, not closure of F-022.

S-021 is ready to freeze and collect a new candidate; S-022 remains blocked on
successful engineering qualification. Local-only diagnostics, unobserved Windows
host/Docker-daemon internals and no Core 05 publication claim remain explicit.


### S-021 candidate 7 freeze and dispatch — IN_PROGRESS

After saving the S-020 completion checkpoint, an alternate Git index froze the
repair as commit `257e3c2787441d7bf4059041b9f07d6fdbe66481`, branch
`codex/instrumentation-qualification-07`, in the clean isolated checkout
`/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary`.
Primary HEAD and the staged tracker hash remain unchanged. All prior branches,
qualification roots, unsuccessful probes and attempts are retained.

Dispatched in that checkout:
`CARTULARY_TEST_RUN_ID=instrumentation-s021-qualification-07 make harness-instrumentation-qualify QUALIFICATION_PLAN=tools/harness_instrumentation_qualification_plan.json`.
Its sole manifest is
`.cartulary/test-results/instrumentation-s021-qualification-07/qualification/qualification.json`.
The manifest will bind the frozen source digest; the primary dirty digest must
not be substituted. The runner owns preparation, all randomized treatments,
measurement and cleanup. No repository tests or generators run concurrently with
qualification. All budgets and the plan remain unchanged. S-021/F-022 are open,
and S-022 has not begun.


Candidate 7 runtime binding: native session `50291`; manifest source digest
`sha256:99a2eab71c0b23770a7648c00b6f1aefd05810abeb9ce6f0da87af3340fac592`.
The first nine Go cold/warm-up attempts are published. This is dispatch evidence,
not a qualification pass. The frozen executable commit remains
`257e3c2787441d7bf4059041b9f07d6fdbe66481`.


Candidate 7 progress checkpoint (collection still active): Go and Vitest each
completed six blocks (48 attempts each, including cold/warm-up). The policy
requested the sixth blocks. Shell cold collection follows; at this checkpoint
101 total results are retained and no measurement loss has been reported.
Independent diagnostic arithmetic over Go's six paired blocks gives CPU upper
bounds of approximately 3.8293% (off/1), 5.6439% (off/4), 3.8596% (basic/0),
3.5685% (basic/1), and 3.6589% (basic/4). These are provisional arithmetic from
published results, not a substitute for the final public validator. Reader CPU
medians are about 4.6 ms for one request and 16.4–17.6 ms for four. The off/1
paired median is near zero but MAD is about 1.28 percentage points; therefore
component improvement does not establish the required conservative total-cost
bound. The candidate remains running to complete the remaining workload matrix.
No executable source or plan changes are being made during collection.


Candidate 7 progress checkpoint: Go, Vitest, shell and service each completed
48 attempts (cold, warm-up and six measured blocks), for 192 total. Every
published result has a non-null measurement and null failure at this checkpoint.
All four cohorts needed the policy-prescribed sixth block. Browser collection
follows. This is collection coverage, not a budget or stability pass. The final
public validator has not yet run; executable source, storage posture, plan,
cadence, preparation, treatment order and budgets remain unchanged. No S-022
work has begun.


### S-021 candidate 7 complete-matrix rejection checkpoint — IN_PROGRESS

The entire unchanged matrix completed naturally; no treatment was interrupted
or selectively discarded. The retained manifest status is `failed`. All five
workloads required the sixth matched block and each has 48 attempts: six cold,
six discarded warm-up and 36 measured. All 240 commands succeeded, every result
has a non-null measurement and null failure, and every declared coverage flag
is true. This establishes eligible collection, not successful qualification.

Frozen executable: `257e3c2787441d7bf4059041b9f07d6fdbe66481`, source
`sha256:99a2eab71c0b23770a7648c00b6f1aefd05810abeb9ce6f0da87af3340fac592`,
branch `codex/instrumentation-qualification-07`. Environment remains the adopted
local WSL2 guest and the frozen ext4 checkout/storage posture. No source, plan,
capacity, preparation, service-reuse, reader-cadence or budget changes were made
during collection. The isolated checkout remained clean afterward. No executable
revision has qualified.

Exact commands in
`/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary`:

| Command | Result / retained evidence |
| --- | --- |
| `CARTULARY_TEST_RUN_ID=instrumentation-s021-qualification-07 make harness-instrumentation-qualify QUALIFICATION_PLAN=tools/harness_instrumentation_qualification_plan.json` | Runner exit 1 / Make 2, 7630654 ms; complete failed matrix under `.cartulary/test-results/instrumentation-s021-qualification-07/qualification/qualification.json`; tool summary/stdout/stderr under the sibling `harness-instrumentation-qualify/` directory |
| `CARTULARY_TEST_RUN_ID=instrumentation-s021-validation-07 make harness-performance-check EVIDENCE_ROOTS_FILE=/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary/.cartulary/test-results/instrumentation-s021-qualification-07/qualification/qualification.json` | Validator exit 13 / Make 2; all 30 cells eligible, all five workloads unstable after six blocks; qualification rejected |

The read-only validator does not create a retained root from
`CARTULARY_TEST_RUN_ID`. Its first invocation's failure is recorded in the tool
transcript. It was repeated solely to capture durable stdout/stderr and exact
argv/environment/exit metadata under
`.cartulary/test-results/instrumentation-s021-validation-07/`. That directory is
an external validation capture, not a canonical harness run. Both invocations
rejected the same immutable evidence. `independent-arithmetic.json` there retains
signed paired arithmetic computed independently from the result JSON; it is
diagnostic support, not a replacement for the public validator.

The validator rejected 23 of 25 non-reference cells on total incremental CPU.
Four cells exceeded the 32 MiB RSS bound. Elapsed gates also failed (11 canonical
p90, 12 outer p90, one canonical median and three outer median cell failures).
All finalization bounds passed; the largest was 129.600384 ms against 500 ms.
The largest additional-RSS upper estimate was 43.032510 MiB against 32 MiB.
All five reference cells and only the Vitest off/one-reader non-reference cell
passed their numeric gates; every workload still failed the stability rule.

Values below are conservative CPU/RSS upper estimates. Canonical and outer
columns are paired median / nearest-rank p90 percentage changes. Signed values
remain intact and negative values do not establish an improvement. Full
precision, declared errors, individual failures and signed pairs are in the
retained JSON; rounded values here do not decide gates.

| Workload | Treatment | CPU % one core | RSS MiB | Canonical median / p90 % | Outer median / p90 % |
| --- | --- | --- | --- | --- | --- |
| go | off/1 | 3.8293 | 16.915 | 0.155 / 1.049 | -0.162 / 1.147 |
| go | off/4 | 5.6439 | 5.845 | 0.998 / 1.477 | 0.794 / 2.199 |
| go | basic/0 | 3.8596 | 21.030 | 1.054 / 1.345 | 2.096 / 3.200 |
| go | basic/1 | 3.5685 | 23.796 | 0.974 / 3.109 | 2.029 / 2.831 |
| go | basic/4 | 3.6589 | 20.136 | 0.816 / 4.002 | 1.510 / 3.088 |
| vitest | off/1 | 0.1945 | 6.141 | -0.263 / 1.001 | -0.374 / 0.996 |
| vitest | off/4 | 1.9290 | 11.359 | -0.451 / 0.927 | 0.046 / 0.726 |
| vitest | basic/0 | 3.0923 | 31.505 | -0.177 / 2.022 | 0.552 / 1.014 |
| vitest | basic/1 | 1.9266 | 17.702 | 0.095 / 1.460 | 1.022 / 1.700 |
| vitest | basic/4 | 2.7175 | 33.073 | 0.453 / 1.523 | 0.753 / 3.497 |
| shell | off/1 | 1.9388 | 14.506 | 0.159 / 0.850 | 0.411 / 1.346 |
| shell | off/4 | 1.5187 | 5.957 | 0.039 / 1.554 | 0.045 / 1.742 |
| shell | basic/0 | 2.4197 | 23.249 | 0.402 / 1.040 | 0.710 / 1.114 |
| shell | basic/1 | 3.2625 | 30.515 | 0.094 / 1.297 | 0.463 / 1.794 |
| shell | basic/4 | 2.7795 | 30.734 | 0.356 / 1.060 | 0.282 / 1.817 |
| service | off/1 | 4.3036 | 6.022 | -1.731 / 4.279 | -1.691 / 4.192 |
| service | off/4 | 5.3032 | 14.631 | 1.520 / 9.538 | 1.454 / 8.907 |
| service | basic/0 | 3.2545 | 23.708 | -0.390 / 2.872 | -0.235 / 2.947 |
| service | basic/1 | 0.9501 | 31.368 | -0.453 / 11.355 | -0.154 / 10.873 |
| service | basic/4 | 0.3601 | 23.147 | -0.433 / 8.998 | -0.299 / 8.670 |
| browser | off/1 | 1.9485 | 0.285 | 0.401 / 9.529 | 0.360 / 9.067 |
| browser | off/4 | 6.5064 | 2.361 | 0.681 / 5.146 | 0.681 / 5.077 |
| browser | basic/0 | 3.3743 | 35.171 | 0.817 / 3.839 | 0.860 / 3.878 |
| browser | basic/1 | 3.0150 | 43.033 | 2.945 / 9.105 | 3.035 / 8.724 |
| browser | basic/4 | 2.4985 | 40.922 | 1.034 / 8.374 | 1.114 / 7.952 |

F-022 remains unresolved. Candidate 7 materially reduced public-reader component
cost (the Go cohort reports approximately 4.6 ms for one request and 16.4–17.6 ms
for four), but this is insufficient to establish the complete observer budget.
For example, Go off/one-reader has a near-zero paired CPU median and about
1.28 percentage points of paired MAD; service/browser reader-only treatments
also fail. Baseline variation and observer cost must be investigated separately;
these results do not identify Windows activity, storage, or any particular
process as the cause. A repeat with no mechanism repair must not be used to
select a lucky passing cohort.

The user-requested RAM-disk consideration is resolved as a bounded diagnostic
experiment, not a qualification change. The preserved retained-data tmpfs probe
recorded approximately 9.058 ms CPU versus 8.856 ms on ext4, with no CPU benefit.
It did not test moving the entire executable checkout to RAM and does not prove
that all storage effects are absent. The current candidate kept its frozen
storage configuration. The container helper was also inspected: startup already
occurs only upon lifecycle-proof registration, so unconditional startup on
service-free workloads is not an established cause.

S-020's candidate 7 structural/functional checkpoint remains DONE. S-021, F-022,
RB-006 and the overall effort remain IN_PROGRESS/open. S-022 has not begun: its
`agent-finalize`, broad end-of-effort checks, tracker Markdown-lint coverage and
V-022/V-023 completion checkpoint are explicitly outstanding because S-021
failed. No eligible full warm `check` run was supplied for retained-run
maintenance; no finalization or maintenance pass is claimed. T-007/T-008 and
final F-014–F-019 dispositions remain dependent on S-022.

No qualification or validation helpers remained after completion. Primary HEAD
remains `1d423c6e5f9cd204f8210fa9958c780fd933d293`; the staged tracker SHA-256 is
still `9ab829e44fa8df7ba1ee86e794518acd97129cea6af416966fbdda4b33f83426`.
Both Git diff checks passed before saving this checkpoint. All failed candidate
branches, result roots, exclusions and unsuccessful probes remain intact. Do not
archive the qualification worktree: its ignored outputs contain the evidence.
Rollback remains a coherent producer/reader/generator/schema transaction; keep
retained evidence. Default-off, local-only diagnostics, unobserved Windows-host
and Docker-daemon-internal cost, sampled-memory limits and no Core 05 publication
claim remain explicit.


### S-020 post-candidate-7 attribution entry — IN_PROGRESS

The S-021 rejection checkpoint above was saved before reopening implementation.
The next investigation distinguishes conditional producer/validator startup,
parent processing/publication, collector/helper cost and measurement-support
interference. Generated JavaScript validators are a candidate seam because
basic-only schemas are lazily loaded by several producers; this is a hypothesis,
not an established explanation or an authorization to change acceptance.
Require a measured mechanism and functional/security parity before another
freeze. Preserve all candidate 7 evidence and budgets. S-022 remains gated.


### S-020 runtime schema compilation attribution and repair — IN_PROGRESS

A temporary `process.cpuUsage` wrapper around `validateSchemaSync` measured
per-process schema counts, first-call CPU and total validation CPU while the
existing public local-migration row ran with diagnostics off/basic and cache
mode off. This was component attribution, not external total-cost qualification.
No private event staging was inspected and no benchmark workload was lengthened.
Both temporary probes were removed by restoring the exact pre-probe source after
their commands exited. JSON profiles and source-backup digests are retained under
`.cartulary/test-results/instrumentation-f022-schema-profile-01/` and `-02/`.

Exact commands used `CARTULARY_TEST_RUN_ID=instrumentation-f022-schema-profile-<mode>-<01|02>`
with `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.local_migration_activities HARNESS_DIAGNOSTICS=<mode> CARTULARY_HARNESS_CACHE_MODE=off`.
Each command passed 1/1. Their canonical durations/source identities are:

| Mode / probe | Canonical ms | Dirty source digest |
| --- | --- | --- |
| off/01 | 2466 | `sha256:8440922e6b7f02f8f6ff56da2df7783452b814f043558801e84026cb14d1f604` |
| off/02 | 2332 | `sha256:3421216c9c1a922d5565407d716b6fb3f821bbe98bae7fda6a0f68fa3dac5cc8` |
| basic/01 | 2458 | `sha256:8440922e6b7f02f8f6ff56da2df7783452b814f043558801e84026cb14d1f604` |
| basic/02 | 2091 | `sha256:3421216c9c1a922d5565407d716b6fb3f821bbe98bae7fda6a0f68fa3dac5cc8` |

Before repair, parent plus row-runner schema validation consumed 1372777 µs
(off) and 1370505 µs (basic). These totals include the small attribution wrapper.
The expensive calls constructed AJV compilers and scanned/read support schemas
at runtime. The diagnostic-only schema calls accounted for approximately 17 ms;
they did not explain the entire failed qualification.

The repair compiles every registered JavaScript schema during generation using
the existing pinned AJV compiler/options. Generated modules remain lazy and
dependency-free. The native renderer keeps the original bounded foundation
closure, so JavaScript-only owner schemas introduce no native schema coupling.
Both synchronous and asynchronous runtime validation now use that generated
registry; the runtime compiler, per-process compiler cache, schema-directory
discovery and ad-hoc unregistered-schema fallback are removed. All 195 existing
schema identities remain registered with unchanged acceptance sets. An unknown
identity is deliberately rejected even if a matching file exists on disk; new
schemas must join the authored registry and regeneration transaction.

The bootstrap test now checks full registry membership, lazy loading of a
non-foundation owner schema without authored schemas or installed packages,
valid/invalid sync/async parity, and rejection of an unregistered on-disk schema.
After repair, parent plus row validation measured 101460 µs (off) and 116838 µs
(basic). This confirms removal of a runtime compilation cost; it does not prove
the remaining CPU/RSS/elapsed/stability gates or attribute all prior variance.

Primary WSL2 functional/projection ledger so far:

| Run ID | Exact command | Result |
| --- | --- | --- |
| `instrumentation-f022-compiled-generate-01` | `make generate` | PASS, 11220 ms |
| `instrumentation-f022-compiled-evidence-01` | `make harness-evidence-contract` | PASS, 19210 ms |
| `instrumentation-f022-compiled-command-01` | `make harness-command-surface-contract` | PASS, 56490 ms |
| `instrumentation-f022-compiled-drift-01` | `make generate-drift` | FAIL, 3/4, 19532 ms; `tools/execution_topology_render_index.json` still reflected source before the subsequent runtime/test edits; exact failed root retained |
| `instrumentation-f022-compiled-policy-01` | `make generated-artifact-policy-check` | PASS, 3/3, 2105 ms |

The projection mismatch is repaired through `make generate`, never by editing
the generated index. S-020 remains open until the remaining functional gates and
saved completion checkpoint. Candidate 7 and its rejection are unchanged; a new
clean executable freeze and qualification are required. S-022 remains gated.


### S-020 candidate 8 structural repair completion checkpoint — DONE

Saved before the successor qualification workstream. The measured repair moves
registered JavaScript schema compilation into the generation transaction and
uses one lazy runtime validation boundary for sync/async consumers. Native
foundation predicates keep their original bounded closure. No schema acceptance,
qualification policy, workload, cadence, preparation, capacity, storage posture
or budget changed. Installed generated validators now cover all 195 authored
registry entries (approximately 7.23 MB of generated source); no runtime compiler
or schema-directory enumeration remains in this boundary. This is not a second
schema authority or a compatibility translator.

The temporary attribution probes are absent from executable source. Their
component results and unfrozen source identities are retained above. Primary
final executable source is
`sha256:684a7dec30073af0f7265a818cd2136ae5492a7d0054ffe2133b2c31590a7d7c`
at unchanged dirty `main`; the forthcoming clean freeze must bind its own
identity rather than borrowing this digest.

Additional exact public gates (each ID is `CARTULARY_TEST_RUN_ID` and retained
root suffix under primary `.cartulary/test-results/`):

| ID | Command | Result |
| --- | --- | --- |
| `instrumentation-f022-compiled-generate-02` | `make generate` | PASS, 10140 ms; refreshes the stale topology projection |
| `instrumentation-f022-compiled-drift-02` | `make generate-drift` | PASS, 4/4, 18989 ms |
| `instrumentation-f022-compiled-shapes-01` | `make json-shape-check` | PASS, 3/3, 3993 ms |
| `instrumentation-f022-compiled-lint-01` | `make lint-scripts` | PASS, 2/2, 514 ms |
| `instrumentation-f022-compiled-basic-01` | `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_observation,harness.evidence_accounting.behavior.native_collection,harness.evidence_accounting.behavior.local_migration_activities HARNESS_DIAGNOSTICS=basic` | PASS, 3/3, 2442 ms |
| `instrumentation-f022-compiled-browser-01` | `make service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.https_trust HARNESS_DIAGNOSTICS=basic` | PASS, 11/11, 55589 ms; resource index complete, 27 sweeps/446 samples, two owned containers, zero failed sweeps/omissions/deadline overruns |

`make harness-observability-check RESULTS_DIR=.cartulary/test-results/instrumentation-f022-compiled-browser-01`
passed with complete diagnostics/execution. Both human and JSON performance and
resource views passed via `make explain-run` for that exact root, `DETAIL=performance|resources`
and optional `JSON=1`; argv/exit/output captures are in
`.cartulary/test-results/instrumentation-f022-compiled-views-01/`. The evidence
contract, command-surface contract and artifact-policy pass are recorded above,
as is the superseded failed drift attempt.

S-020 is functionally complete and ready to freeze. F-022/RB-006 remain open;
S-021 requires a new complete-observer matrix against the frozen executable.
Rollback restores the generator, complete generated registry/modules and both
runtime consumers together; do not restore a runtime fallback in isolation.
Existing authored schema additions must be registered and regenerated. Default
remains off, observations remain local, no Core 05 claim is made, and S-022 has
not begun. Preserve all earlier failed evidence and the staged user tracker.


### S-021 candidate 8 freeze and dispatch — IN_PROGRESS

The S-020 completion checkpoint was saved first. An alternate Git index then
froze commit `2ee483899657dc83c3c42568aac6346b3a06185e`, branch
`codex/instrumentation-qualification-08`, in the existing clean isolated checkout
`/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary`. The delta
from candidate 7 contains the runtime loader, its bootstrap test, generator,
generated validator registry/modules and topology render index only. Primary
HEAD/index and the staged user tracker remain unchanged. All prior branches and
ignored qualification evidence remain present.

Next exact dispatch in that checkout:
`CARTULARY_TEST_RUN_ID=instrumentation-s021-qualification-08 make harness-instrumentation-qualify QUALIFICATION_PLAN=tools/harness_instrumentation_qualification_plan.json`.
Its manifest is
`.cartulary/test-results/instrumentation-s021-qualification-08/qualification/qualification.json`.
The manifest binds the clean source digest; no primary dirty digest is reused.
Use the complete unchanged plan and budgets. Do not run tests, generators or
profiling concurrently with the matrix or change executable source during it.
F-022/S-021 remain open; S-022 has not begun.


Candidate 8 runtime binding: native session `1237`; manifest source digest
`sha256:4982b86dc1c7dd4c0a546096ad53ef5915b09f9f49769c11c9ca5940c4bf2097`.
The first 12 Go cold/warm-up attempts are retained. This is dispatch/coverage
evidence, not a qualification pass. The frozen commit remains
`2ee483899657dc83c3c42568aac6346b3a06185e`.


Candidate 8 progress checkpoint: Go completed six measured blocks (48 total
attempts), Vitest completed five without requiring a sixth (42), and shell
completed six (48). All 138 published results have null failure at this point.
Service collection follows. Provisional Go CPU upper estimates before the small
error term are approximately 2.7245% (off/1), 1.6490% (off/4), 3.4996% (basic/0),
3.0891% (basic/1), and 2.8714% (basic/4); they still exceed 0.5%. These are
independent arithmetic over published results, not the final public validator.
Reader medians remain approximately 4.7 ms for one request and 16.2–16.4 ms for
four. The complete matrix continues unchanged to assess the other workloads and
RSS. No executable changes, profiling or tests run concurrently. S-022 remains
gated and no passing qualification is claimed.


Candidate 8 progress checkpoint: all non-browser cohorts are complete: Go 48,
Vitest 42, shell 48 and service 48 attempts. The first 186 results all have
non-null measurements, null failures and complete declared coverage. Service
required the sixth block. Its provisional CPU upper estimates before error are
1.0768%, 2.7498%, 1.6995%, 1.4283%, and 1.1950% for off/1, off/4, basic/0,
basic/1, and basic/4 respectively; these still exceed 0.5%. Browser cold
collection has begun. Source/settings remain frozen and final public validation
has not run. No completed qualification or S-022 handoff is claimed.


Candidate 8 progress checkpoint: 216 attempts are retained; browser cold, warm-up
and three measured blocks are complete. Published attempts continue to have
valid measurements and successful workload exits. Independent arithmetic over
the completed non-browser cohorts finds CPU upper estimates above 0.5% in
19/20 non-reference cells; shell/basic/1 and shell/basic/4 also exceed 32 MiB
(36.1571 and 33.9459 MiB respectively). Service p90 elapsed differences exceed
3% in every non-reference treatment. The complete matrix and public validator
remain required; no unfavorable valid observation is excluded.

Read-only inspection found that catalog selector resolution rescans Go package
files for each catalog row. This is a candidate for later component attribution,
not an established explanation of F-022 or permission to modify the frozen run.
No source changes or profiling are occurring alongside qualification. Preserve
all candidate evidence and save the S-021 disposition before reopening S-020.


### S-021 candidate 8 complete-matrix rejection checkpoint — IN_PROGRESS

Candidate 8 completed naturally against frozen executable commit
`2ee483899657dc83c3c42568aac6346b3a06185e`, source digest
`sha256:4982b86dc1c7dd4c0a546096ad53ef5915b09f9f49769c11c9ca5940c4bf2097`,
in the isolated qualification checkout. Native session `1237` terminated with
runner exit 1 / Make exit 2 after 7115980 ms. All 234 observations have non-null
measurements, null failures and complete declared coverage. Go, shell, service
and browser each retained 48 attempts (including six measured blocks); Vitest
retained 42 (five measured blocks, with no stability extension required). No
unfavorable valid observation was discarded. The manifest status is `failed`.

Exact collection command and result root are recorded in the dispatch checkpoint
above. Its terminal summary is
`.cartulary/test-results/instrumentation-s021-qualification-08/harness-instrumentation-qualify/tool-run-summary.json`;
the manifest and all per-attempt evidence remain under that root's `qualification/`.

Independent public validation ran once, after collection, in the same checkout:
`CARTULARY_TEST_RUN_ID=instrumentation-s021-validation-08 make harness-performance-check EVIDENCE_ROOTS_FILE=/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary/.cartulary/test-results/instrumentation-s021-qualification-08/qualification/qualification.json`.
It rejected with validator exit 13 / Make exit 2. Exact argv, environment override,
UTC timestamps, exit and stdout/stderr are retained in
`.cartulary/test-results/instrumentation-s021-validation-08/{command.json,stdout.log,stderr.log}`.
This is an external capture directory, not a canonical harness run. The reader
target does not manufacture a run root from its run-ID environment variable.

All 30 treatment cells are eligible. Only the five reference cells and
shell/off/1 pass every numeric budget. CPU fails in 24/25 non-reference cells;
RSS fails in six cells. Canonical p90 fails in eight, outer p90 in eleven, and
outer median in two; no canonical-median or finalization gate fails. Maximum
finalization is 148.602 ms, below 500 ms. Maximum RSS upper estimate is
48.1895 MiB, above 32 MiB. The public disposition is
`go:unstable,vitest:budget_or_coverage,shell:unstable,service:unstable,browser:unstable,manifest_not_complete`.

The following rounded arithmetic is independently calculated from retained
results; `independent-arithmetic.json` beside the validation capture preserves
full precision, signed pairs, CPU medians/MADs and failed gate names. The public
validator remains the disposition owner. Elapsed thresholds remain median 2%
and nearest-rank p90 3%; CPU remains 0.5% of one core, RSS 32 MiB and finalization
500 ms. CPU/RSS upper estimates include three paired MADs and declared error.

| Workload / treatment | Canonical median / p90 % | Outer median / p90 % | CPU upper % | RSS upper MiB | Finalization ms |
| --- | --- | --- | --- | --- | --- |
| go/off/1 | 0.908 / 1.367 | 0.832 / 1.777 | 2.7257 | 13.9733 | 0.000 |
| go/off/4 | 1.340 / 2.407 | 0.863 / 1.724 | 1.6500 | 4.7150 | 0.000 |
| go/basic/0 | 0.477 / 1.781 | 2.260 / 3.202 | 3.5010 | 12.2292 | 80.978 |
| go/basic/1 | 1.214 / 2.311 | 2.209 / 3.170 | 3.0902 | 14.4139 | 93.264 |
| go/basic/4 | 0.904 / 2.744 | 1.672 / 3.243 | 2.8729 | 14.4997 | 83.340 |
| vitest/off/1 | -0.882 / 0.458 | -0.518 / 0.231 | 2.6192 | 3.0733 | 0.000 |
| vitest/off/4 | -0.911 / 0.603 | -0.724 / 0.050 | 1.0521 | 3.5973 | 0.000 |
| vitest/basic/0 | -0.763 / 0.935 | -0.097 / 2.545 | 3.2982 | 31.0761 | 89.862 |
| vitest/basic/1 | 0.671 / 2.066 | 0.842 / 1.759 | 1.6446 | 18.8274 | 88.892 |
| vitest/basic/4 | 0.241 / 0.596 | 0.670 / 1.461 | 1.0995 | 26.1553 | 84.745 |
| shell/off/1 | -0.244 / 0.385 | -0.235 / 0.504 | 0.2288 | 3.8855 | 0.000 |
| shell/off/4 | -0.550 / -0.182 | -0.370 / -0.151 | 0.7956 | 26.8162 | 0.000 |
| shell/basic/0 | -0.026 / 0.341 | 0.254 / 0.692 | 1.3787 | 24.6573 | 81.081 |
| shell/basic/1 | -0.296 / 0.713 | -0.003 / 1.160 | 2.0033 | 36.1571 | 144.643 |
| shell/basic/4 | -0.345 / 0.536 | 0.232 / 0.850 | 0.6936 | 33.9459 | 80.408 |
| service/off/1 | -0.570 / 11.360 | -0.513 / 10.639 | 1.0768 | 7.9793 | 0.000 |
| service/off/4 | -0.125 / 11.590 | -0.174 / 10.813 | 2.7499 | 8.0906 | 0.000 |
| service/basic/0 | 0.263 / 10.104 | 0.356 / 9.611 | 1.6996 | 30.5210 | 114.632 |
| service/basic/1 | -0.009 / 8.640 | 0.159 / 8.515 | 1.4283 | 27.9642 | 107.564 |
| service/basic/4 | -1.094 / 9.426 | -0.916 / 9.103 | 1.1951 | 25.1797 | 99.019 |
| browser/off/1 | 0.658 / 9.385 | 0.677 / 8.911 | 3.1610 | 35.6602 | 0.000 |
| browser/off/4 | -0.042 / 10.021 | 0.007 / 9.558 | 2.4229 | 24.1729 | 0.000 |
| browser/basic/0 | 0.083 / 0.900 | 0.161 / 0.931 | 3.5601 | 40.1960 | 148.602 |
| browser/basic/1 | 0.679 / 8.682 | 0.803 / 8.039 | 5.5961 | 48.1895 | 110.709 |
| browser/basic/4 | -0.006 / 0.584 | 0.102 / 0.801 | 0.8062 | 34.8831 | 130.190 |

F-022 remains unresolved. Moving registered schema compilation out of runtime
removed measured component work, but did not establish complete-observer
qualification. Cross-candidate differences are not matched treatment evidence
and are not claimed as causal improvements. The previous RAM probe concerned
retained evidence only and found no CPU benefit; the executable checkout has not
been moved to RAM. No capacity, preparation, service posture, storage, cadence,
workload, statistics or budget was changed for this candidate.

Post-run checks confirm the isolated checkout is clean at its frozen commit and
no qualification runner, meter, gate helper or collector remains. Primary HEAD
remains `1d423c6e5f9cd204f8210fa9958c780fd933d293`; staged tracker SHA-256 remains
`9ab829e44fa8df7ba1ee86e794518acd97129cea6af416966fbdda4b33f83426`.
Both primary Git diff whitespace checks pass. All older candidate branches and
failed evidence remain preserved. Do not archive the qualification worktree; its
ignored retained outputs are needed.

S-020's candidate 8 functional checkpoint remains DONE. S-021, F-022, RB-006 and
the overall effort remain IN_PROGRESS/open. No executable commit is qualified.
S-022 has not begun: finalization, broad final checks, tracker Markdown-lint
coverage, final source/evidence reconciliation and the V-022/V-023 completion
checkpoint remain outstanding under the sequential gate. Historical finalizers
and functional passes do not discharge those obligations. Default-off, local-only
resource diagnostics, Windows-host/Docker-daemon-internal exclusions, sampled
memory limitations and absence of a Core 05 publication claim remain explicit.


### S-020 post-candidate-8 catalog attribution entry — IN_PROGRESS

The complete S-021 candidate 8 rejection checkpoint above is saved before this
workstream resumes. Investigate catalog selector resolution as a bounded
structural cost seam: current Go selector resolution scans all test sources in
a package again for each row, while module-global selector caches outlive a
single catalog load. Measure component CPU/counts through the existing public
local-migration row before selecting a repair. Keep cache lifetime tied to one
validated catalog load, preserve exact selectors and path/build-context checks,
and require fresh-load mutation cases if implementation changes. This is an
attribution hypothesis, not a qualification claim or a reason to tune budgets.
All temporary profiling code must be removed before functional gates or freezing.
S-020/F-022 remain open; S-021 requires a new complete frozen cohort after any
executable repair, and S-022 remains gated. Candidate 8 evidence is immutable.


S-020 catalog attribution checkpoint: the exact public command
`make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.local_migration_activities HARNESS_DIAGNOSTICS=<off|basic> CARTULARY_HARNESS_CACHE_MODE=off`
passed 1/1 with run IDs `instrumentation-f022-catalog-profile-off-01` (2057 ms)
and `instrumentation-f022-catalog-profile-basic-01` (2043 ms). Temporary per-call
`process.cpuUsage` counters measured each graph process loading the catalog twice
and resolving 1656 Go rows across 161 packages; each row process loaded it once
and resolved 828 Go rows across the same 161 packages. Before repair, Go symbol
scans consumed approximately 559–574 ms CPU per graph process and 282–300 ms per
row process. Whole catalog loads consumed approximately 2.49–2.50 s per graph
process and 1.80–1.84 s per row process. Nested component totals overlap and must
not be summed; this is attribution, not qualification.

A separate stack-only probe used run ID `instrumentation-f022-catalog-stack-01`,
the same row/target with diagnostics off and cache mode off, and passed 1/1 in
2002 ms. It identifies the second graph-process load as
`WorkGraphCache.lookup → context → dependencyClosure → resolveCacheDependencyClosure`;
cache mode was checked only after constructing the fingerprint. Profiles, stacks
and exact pre-probe source digests are preserved in primary
`.cartulary/test-results/instrumentation-f022-catalog-attribution-01/`. Both sets
of temporary source edits were restored byte-for-byte before permanent repair.

The repair now scopes package symbols, source text, TypeScript titles and Go build
context to one catalog load. Per-selector path checks remain active. Fresh loads
observe source, imported-suite and toolchain-pin edits; module-global selector
caches are removed. Cache eligibility is separate from fingerprint construction;
cache-off retains eligibility outcomes and bypasses dependency/input hashing.
The graph passes its validated catalog lazily to cache closure and directly to
browser planning, avoiding repeated full loads within the same graph. No public
selection, schema, budget or cache-reuse contract changes. New regressions cover
fresh-load edits, duplicate symbols, pin changes, imported-suite changes, symlink
replacement despite cached facts, and disabled-cache input isolation.

Public discovery selected `harness.test_catalog` via `make task-guide` and
`make explain-test-owner`. Two attempted discovery names were invalid:
`make task-guide ROLE=module-author OWNER=harness.scheduler` and
`make explain-target TARGET=harness-test-catalog-tests DETAIL=rows` each returned
configuration/usage exit 2; no test ran and no retained test root was created.
The authored catalog maps its row to `harness-contract-tests`; use the discovered
owner route. Current verification is
`CARTULARY_TEST_RUN_ID=instrumentation-f022-catalog-contract-01 make test-slice OWNER=harness.test_catalog`,
native session `22659`. Its outcome is not yet established. S-020 stays open.


S-020 follow-up evidence: the first catalog owner route failed 0/1 after 101453 ms
with `product/test_assertion_failure` (harness exit 10 / Make 2); cleanup passed
3/3. The 205 inner assertions had 203 passes, including every new selector/cache
case. Two failures were a stale topology render index after authored work-graph
edits and the old `missing schema attachment` expectation, which no longer
matched candidate 8's deliberate registered-schema boundary. The fixture now
expects rejection of that exact retired identity as unregistered. No acceptance
set or compatibility translator was restored. The full failed root is preserved.
`make explain-run RESULTS_DIR=.cartulary/test-results/instrumentation-f022-catalog-contract-01 TARGET=test-slice DETAIL=logs`
returned the referenced diagnostics. An earlier `TARGET=harness-contract-tests`
explanation failed because that nested command has no canonical target projection
in this owner-slice root; it was corrected to the actual selected target.

Post-repair attribution used the same public row/selection, off/basic and cache
mode off, with run IDs `instrumentation-f022-catalog-profile-off-02` (PASS 1/1,
1412 ms) and `instrumentation-f022-catalog-profile-basic-02` (PASS 1/1, 1142 ms).
Each graph process now loaded the catalog once. Go symbol-resolution CPU was
30141/26746 µs in the two graph processes and 37870/37350 µs in their row
processes. Whole catalog CPU was 1501773/1435879 µs in graph processes and
1430817/1432372 µs in row processes. Calls still resolve every row, while package
scans and source reads share only load-local facts. Full profiles, run bindings
and source-backup hashes are retained in
`.cartulary/test-results/instrumentation-f022-catalog-attribution-02/`. This is
component attribution with overlapping counters; timing differences between the
two modes are not evidence of an improvement. Both temporary profiler files were
restored exactly, and marker searches confirm no temporary profiling code remains.

`CARTULARY_TEST_RUN_ID=instrumentation-f022-catalog-generate-01 make generate`
passed in 10150 ms, refreshing the generated topology index through its owner.
The repaired owner route is now running with ID
`instrumentation-f022-catalog-contract-02`, native session `37125`.
S-020 remains open until all required functional/projection checks pass and its
completion checkpoint is saved; no qualification has begun for this source.


Catalog attribution source bindings (unfrozen, temporary probes only):

| Run ID | Source digest |
| --- | --- |
| `instrumentation-f022-catalog-profile-off-01` | `sha256:7970341a26e9e17e13933cbacf2b6096900d8361fad6db3524be3d46cfb095a5` |
| `instrumentation-f022-catalog-profile-basic-01` | `sha256:7970341a26e9e17e13933cbacf2b6096900d8361fad6db3524be3d46cfb095a5` |
| `instrumentation-f022-catalog-stack-01` | `sha256:7be3197dfd4a15251c7287f66d603a490d456219360d36153c70fd36a0ed23db` |
| `instrumentation-f022-catalog-profile-off-02` | `sha256:9d7bb7f77b4b679e95057ea11827687994829d7edc466061d43e97621989b009` |
| `instrumentation-f022-catalog-profile-basic-02` | `sha256:9d7bb7f77b4b679e95057ea11827687994829d7edc466061d43e97621989b009` |


### S-020 candidate 9 structural repair completion checkpoint — DONE

Saved before the successor qualification workstream. The confirmed cost repair
uses one selector-resolution scope per catalog load and shares the graph's
validated catalog with cache and browser consumers. Package scans and source
reads are reused only within that scope; path checks remain per selector and
fresh loads observe edits to sources, imported registration suites and Go pins.
Cache eligibility is separate from fingerprint construction. `off` no longer
reads unused cache inputs or constructs a fingerprint; normal/cold reuse and
security/freshness checks remain unchanged. The prior process-global selector
caches are removed. No schema, public selector, workload, cadence, preparation,
capacity, service-reuse posture, storage or budget changed. A byte comparison
against candidate 8 confirms the plan, policy and all 195 schemas are unchanged.

Compatibility impact: disabled caching no longer fails on an unsafe/inaccessible
input used solely for cache fingerprinting; the same input remains rejected when
caching needs its fingerprint. This does not relax workload input validation.
Fresh catalog loads now reject stale selector assumptions after pin or imported
suite edits. Graphs, cache identities for enabled caching and published evidence
formats keep their existing contracts. Roll back selector loading, graph/catalog
injection, cache eligibility/key construction, regression fixtures and the
generated topology index together. Preserve all qualified-candidate history;
no prior candidate is qualified or relabeled by this checkpoint.

Temporary profilers are removed. The independent component attribution above
confirms avoided work but does not prove total-observer budgets. Final primary
executable digest for these functional gates is
`sha256:299d33fec73ad94555e0c1c64eeefe9b31f48a9de73ff4bb5780cab2190b1092`
at dirty `main`; a clean freeze must record its own identity.

WSL2 validation ledger (every ID is the explicit `CARTULARY_TEST_RUN_ID` and
retained-root suffix under primary `.cartulary/test-results/`):

| ID | Exact command | Result |
| --- | --- | --- |
| `instrumentation-f022-catalog-generate-01` | `make generate` | PASS, 10150 ms |
| `instrumentation-f022-catalog-contract-02` | `make test-slice OWNER=harness.test_catalog` | PASS, 1/1, 102994 ms; all 205 inner cases pass |
| `instrumentation-f022-catalog-evidence-01` | `make harness-evidence-contract` | PASS, 18140 ms |
| `instrumentation-f022-catalog-command-01` | `make harness-command-surface-contract` | PASS, 51330 ms |
| `instrumentation-f022-catalog-drift-01` | `make generate-drift` | PASS, 4/4, 18465 ms |
| `instrumentation-f022-catalog-policy-01` | `make generated-artifact-policy-check` | PASS, 3/3, 1789 ms |
| `instrumentation-f022-catalog-shapes-01` | `make json-shape-check` | PASS, 3/3, 3884 ms |
| `instrumentation-f022-catalog-lint-01` | `make lint-scripts` | PASS, 2/2, 579 ms |
| `instrumentation-f022-catalog-basic-01` | `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_observation,harness.evidence_accounting.behavior.native_collection,harness.evidence_accounting.behavior.local_migration_activities HARNESS_DIAGNOSTICS=basic` | PASS, 3/3, 1467 ms |
| `instrumentation-f022-catalog-browser-01` | `make service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.https_trust HARNESS_DIAGNOSTICS=basic` | PASS, 11/11, 55995 ms |

The browser resource index is complete: 27 sweeps, 396 samples, two owned
containers, zero failed sweeps, omissions or container deadline overruns.
`make harness-observability-check RESULTS_DIR=.cartulary/test-results/instrumentation-f022-catalog-browser-01`
passes complete diagnostics/execution. Human and JSON performance/resources
views also pass via `make explain-run` for that exact root, `DETAIL=performance|resources`
and optional `JSON=1`; exact argv/exit/output captures are retained in
`.cartulary/test-results/instrumentation-f022-catalog-views-01/`.

The failed first catalog route, its stale-index/error-expectation diagnosis,
invalid discovery/explanation attempts, before/after attribution probes, and all
candidate 1–8 failures remain preserved above and at their exact roots. S-020 is
functionally complete and ready for a clean freeze. F-022/RB-006 remain open and
S-021 requires the entire unchanged matrix. S-022 has not begun; default-off,
local-only observations and the absence of a Core 05 claim remain explicit.


### S-021 candidate 9 freeze and dispatch — IN_PROGRESS

The saved S-020 completion checkpoint precedes this workstream. An alternate
Git index froze executable commit `bf79a14a0a7440e25ad3d85299005652977634bc`,
branch `codex/instrumentation-qualification-09`, with candidate 8 as parent.
The delta is exactly eleven files: seven catalog/graph/cache implementation
modules, the test-support facade and two regression modules, plus the generated
topology render index. Primary HEAD, the full primary index bytes and staged
tracker content are unchanged. The existing isolated checkout
`/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary` is clean at
that commit. All older branches and ignored failed evidence remain present.
Freeze metadata is `/tmp/cartulary-qualification-freeze-09.json`.

Next exact command in that checkout:
`CARTULARY_TEST_RUN_ID=instrumentation-s021-qualification-09 make harness-instrumentation-qualify QUALIFICATION_PLAN=tools/harness_instrumentation_qualification_plan.json`.
The manifest is
`.cartulary/test-results/instrumentation-s021-qualification-09/qualification/qualification.json`;
it must bind its own clean source digest. Execute the entire unchanged matrix
and retain every attempt and exclusion. No executable edits, test suites,
generators or component profiles run concurrently. Candidate 8 remains failed;
S-021/F-022 are open and S-022 has not begun.


Candidate 9 runtime binding: native session `5946`; manifest source digest
`sha256:924eb6dc7fd2a7beb5f0d18097fb38bbf72f6024b0c27ecbb168f5c6fbaccb30`.
The first 4 cold Go observations are published with null failure. The clean
commit remains `bf79a14a0a7440e25ad3d85299005652977634bc`. This confirms
dispatch and provenance only; no qualification pass is claimed.


Candidate 9 collection milestone: all 48 observations each for Go, Vitest and
shell are retained (144 total). Each cohort required six measured blocks; every
published observation has a non-null measurement and null failure. Service and
browser cohorts remain outstanding. No executable source, workload, preparation,
reader cadence or budget has changed during collection.

Independent provisional arithmetic on the completed Go cohort gives CPU median
plus three paired MADs, before the measurement-error term, of 1.4398%, 4.1040%,
4.8831%, 2.2502% and 3.9693% for off/1, off/4, basic/0, basic/1 and basic/4.
The corresponding Vitest estimates are 1.6096%, 3.0806%, 1.8992%, 2.9976% and
4.6598%. These already exceed 0.5%; they are diagnostic arithmetic, not a public
validator disposition. The entire matrix continues unchanged, preserving every
observation. S-021/F-022/RB-006 remain open and S-022 has not begun.


### S-021 candidate 9 policy-superseded interruption checkpoint — NOT QUALIFIED

The user explicitly amended the CPU acceptance criterion: “Allow an average
below 2%.” This supersedes the previous instruction to retain the 0.5% CPU gate;
it does not amend any other budget, cadence, workload, preparation or coverage
requirement. Candidate 9 remains bound to its original policy and source.

The exact proven runner PID 54517 received SIGINT through a Linux pidfd after
checking its command arguments, result-root binding and checkout. Collection
ended after 3107960 ms with runner exit 1 / Make exit 2, preserving 173 attempts:
48 Go, 48 Vitest, 48 shell and 29 service. The first 172 have null failure;
`service-measured-3-basic-1` is retained as `cancelled`, child exit 143. The
manifest remains `incomplete`. No browser observations were collected. Process
inspection found no remaining qualification runner, native meter or gate helper.

Before changing source or policy, the frozen candidate independently ran
`CARTULARY_TEST_RUN_ID=instrumentation-s021-validation-09 make harness-performance-check EVIDENCE_ROOTS_FILE=/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary/.cartulary/test-results/instrumentation-s021-qualification-09/qualification/qualification.json`.
The validator returned timing exit 13 / Make 2: Go, Vitest and shell unstable;
service incomplete, ineligible and budget/coverage failed; browser incomplete;
manifest not complete. Fourteen of fifteen completed non-reference cells exceed
the old CPU limit. Vitest basic/0 and shell basic/1 and basic/4 exceed 32 MiB.
The interrupted run is not qualified under either policy.

Original plan, manifest, attempts and retained run roots remain under isolated
`.cartulary/test-results/instrumentation-s021-qualification-09/`. Exact public
validator argv, UTC times, exit status and stdout/stderr are in
`.cartulary/test-results/instrumentation-s021-validation-09/`; this capture is not
a canonical test run. Commit `bf79a14a0a7440e25ad3d85299005652977634bc` and
source digest `sha256:924eb6dc7fd2a7beb5f0d18097fb38bbf72f6024b0c27ecbb168f5c6fbaccb30`
remain the exact candidate identity. All earlier failed candidates remain intact.

### S-020 user-authorized CPU policy revision — IN_PROGRESS

Begin only after saving the preceding S-021 interruption checkpoint. Adopt the
arithmetic mean of signed per-block total incremental observer CPU percentages,
strictly below 2% of one core in every workload/treatment cell. Continue reporting
CPU median, paired MAD, bounded measurement error and the former conservative
upper estimate as diagnostics, not as an additional CPU budget. Negative pairs
remain signed and do not establish an improvement. The leave-one-block-out rule
continues to apply to every actual gate, now including the CPU mean gate.

Update REQ-848, the versioned typed qualification plan, arithmetic and public
reporting, independent boundary/mean/uncertainty fixtures and characterization.
Reject obsolete plan versions without translation. Preserve measurement shapes,
complete observer accounting, elapsed 2%/3%, RSS upper 32 MiB, finalization 500 ms,
all eligibility requirements, default-off and local-only/no-Core-05 limits.
A new clean freeze and qualification matrix are required; historical evidence
will not be relabeled. S-022 remains gated.


### S-020 CPU-policy revision completion checkpoint — DONE

Saved before the successor freeze. REQ-848 now adopts a strictly-below-2% CPU
arithmetic mean per workload/treatment cell, equal weighting of signed paired
percentages, with CPU median/MAD/bounded error/conservative estimate reported as
diagnostics. No pooling across cells or elapsed denominators is permitted.
Equality fails. CPU's leave-one-block-out stability gate uses the actual mean
disposition. RSS, both elapsed gates, finalization, complete accounting, eligibility,
reader cadence, randomized treatments, workload and preparation remain unchanged.
This is the user's explicit policy amendment, not an executable optimization.

`cartulary.harness_instrumentation_qualification_plan.v2` replaces plan v1 and
renames the budget to `cpu_mean_percent_one_core`. The authored plan sets 2.
Old plans and altered authored budgets are rejected without translation. The
manifest/result schemas remain v1: their measurement shapes are unchanged and
their exact plan digest binds the revised semantics. Instrumentation policy v3,
resource/execution/receipt shapes and canonical performance portfolios are
unchanged. Generated JavaScript/native schema consumers were refreshed through
Make, retiring the old generated plan validator. Public cell output labels
`cpu_mean`, `cpu_mad`, `cpu_error_bound` and `cpu_upper_diagnostic` distinctly.

Independent fixtures cover mean versus median, signed values, equal-weight
percentages versus pooled denominators, exact threshold rejection, reader CPU,
per-cell failure without pooling, leave-one-block-out mean instability, bounded
error reporting, missing coverage, and obsolete/altered plan rejection.

All checks below passed in the primary WSL2 checkout. Each ID is the explicit
`CARTULARY_TEST_RUN_ID` and retained root suffix under `.cartulary/test-results/`.
Functional source digest is
`sha256:f0d76d88e322b6a3bef7f758992fd35bf74f4aea5da1b113e2ab20153e3fc096`
at dirty main; the clean successor must record its own digest.

| ID | Exact command | Result |
| --- | --- | --- |
| `instrumentation-cpu-mean-generate-01` | `make generate` | PASS, 10160 ms |
| `instrumentation-cpu-mean-evidence-01` | `make harness-evidence-contract` | PASS, 18240 ms |
| `instrumentation-cpu-mean-command-01` | `make harness-command-surface-contract` | PASS, 54110 ms |
| `instrumentation-cpu-mean-native-01` | `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_observation,harness.evidence_accounting.behavior.instrumentation_measurement` | PASS, 2/2, 35424 ms |
| `instrumentation-cpu-mean-drift-01` | `make generate-drift` | PASS, 4/4, 20853 ms |
| `instrumentation-cpu-mean-policy-01` | `make generated-artifact-policy-check` | PASS, 3/3, 1897 ms |
| `instrumentation-cpu-mean-shapes-01` | `make json-shape-check` | PASS, 3/3, 4202 ms |
| `instrumentation-cpu-mean-lint-01` | `make lint-scripts` | PASS, 2/2, 599 ms |

Discovery used `make task-guide ROLE=module-author OWNER=harness.evidence_accounting`.
Git diff checks pass; primary HEAD and the original staged tracker hash remain
unchanged. No new qualification result exists yet. No additional service/browser
functional run was needed for this arithmetic/plan-only change; their unchanged
observation paths retain candidate 9 functional evidence and will be exercised
by the new full matrix. S-022 finalization remains gated by S-021.

Rollback this policy revision as one transaction across REQ-848, plan/schema
registration, arithmetic, public reporting, fixtures and generated consumers;
do not change the meaning of existing evidence. All candidates 1–9, including
failed and cancelled observations, remain preserved at their original identities.
Default-off, local-only diagnostics and no Core 05 claim remain explicit.


### S-021 candidate 10 freeze and dispatch — IN_PROGRESS

The preceding S-020 completion checkpoint was saved before this workstream.
Candidate 10 is clean commit `a9f86713367d85ed19110a6b5c00cfab6c05fe4f`,
branch `codex/instrumentation-qualification-10`, parent candidate 9. The existing
isolated checkout is `/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary`.
The alternate-index freeze preserved primary HEAD, full primary index bytes and
the original staged tracker. Metadata is
`/tmp/cartulary-qualification-freeze-10.json`. All earlier candidate branches and
ignored retained evidence remain present.

Exact dispatch command:
`CARTULARY_TEST_RUN_ID=instrumentation-s021-qualification-10 make harness-instrumentation-qualify QUALIFICATION_PLAN=tools/harness_instrumentation_qualification_plan.json`.
Manifest: `.cartulary/test-results/instrumentation-s021-qualification-10/qualification/qualification.json`.
The runner must bind the new clean source and v2 plan. CPU mean is strictly below
2% per cell; other gates remain unchanged. No prior result is reused or relabeled.
No executable edits, generators, test suites or attribution experiments run
concurrently. Source/digest, full matrix, independent validation and final
disposition remain outstanding; S-022 has not begun.


Candidate 10 runtime binding: native session `45235`, commit
`a9f86713367d85ed19110a6b5c00cfab6c05fe4f`, source digest
`sha256:4c51feae880ee4a4adf837e819fc790551bc18eb2c13d78de0ea7f93a32320b4`,
plan digest `sha256:1152b853dc9f2c8ceeb20343245e305806b729ac413f8cf0a686118627ba38af`.
The first 31 Go observations have null failure. This records provenance and
progress only; no qualification pass is claimed.


Candidate 10 collection milestone: Go, Vitest and shell each completed 48
observations, including a required sixth measured block (144 total). A read of
the published results confirms no failed or missing measurements. Service and
browser cohorts remain outstanding. No executable source or protocol changes
have occurred during collection.

Independent provisional arithmetic finds Go CPU means of 0.1118%, 1.7821%,
3.1227%, 1.6107% and 2.8887% for off/1, off/4, basic/0, basic/1 and basic/4.
Thus basic/0 and basic/4 still exceed the new strict 2% gate. Go also has
elapsed median/p90 failures; its RSS estimates are below 32 MiB. Vitest CPU
means for the same treatment order are -1.8437%, -1.2236%, 0.0104%, 0.1018% and
1.4184%; all are below 2%, while basic/4 RSS is 32.9278 MiB, exceeding 32 MiB.
Signed negative values do not establish an improvement. These are provisional
independent calculations, not a public-validator verdict. Full collection and
independent public validation continue; S-021/F-022/RB-006 remain open and
S-022 has not begun.


Candidate 10 service milestone: the first four workloads each completed 48
observations, including six measured blocks (192 total). Every published result
has a non-null measurement and null failure. Browser qualification remains.
The frozen source, v2 plan, preparation and all gates remain unchanged; final
public validation and disposition are pending.


### S-021 candidate 10 complete-matrix rejection checkpoint — NOT QUALIFIED

Candidate 10 completed all 240 observations (48 per workload, including six
measured blocks each) in 6933382 ms, runner exit 1 / Make exit 2. All attempts
have non-null measurements, null failure, and complete declared coverage. The
manifest status is `failed`. No observation was replaced or selectively excluded.
The isolated checkout remains clean at `a9f86713367d85ed19110a6b5c00cfab6c05fe4f`,
source `sha256:4c51feae880ee4a4adf837e819fc790551bc18eb2c13d78de0ea7f93a32320b4`,
plan `sha256:1152b853dc9f2c8ceeb20343245e305806b729ac413f8cf0a686118627ba38af`.
Process inspection found no remaining qualification runner, meter or gate helper.

Independent public validation used exactly
`CARTULARY_TEST_RUN_ID=instrumentation-s021-validation-10 make harness-performance-check EVIDENCE_ROOTS_FILE=/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary/.cartulary/test-results/instrumentation-s021-qualification-10/qualification/qualification.json`.
It returned timing exit 13 / Make 2: all five workloads are unstable after six
blocks, and the manifest is not complete. All 30 cells are eligible; 15 of the
25 non-reference cells fail one or more numeric gates. Two fail CPU mean <2%,
six fail RSS upper ≤32 MiB, nine fail canonical p90, eleven fail outer p90,
and three fail outer median. Canonical median and finalization gates pass
throughout; maximum finalization upper bound is 128.773 ms. The new CPU rule is
working as adopted, but it does not establish overall qualification.

The full run remains under isolated
`.cartulary/test-results/instrumentation-s021-qualification-10/`. Public validator
stdout/stderr, exact argv/environment/cwd/UTC times/exit status and independently
calculated signed arithmetic are retained under
`.cartulary/test-results/instrumentation-s021-validation-10/`. These captures are
external analysis, not a canonical test run. Negative pairs remain signed and
do not imply an improvement. The public validator owns the final disposition.

| Workload/mode/readers | CPU mean, % one core | RSS upper, MiB | Canonical median / p90, % | Outer median / p90, % | Finalization upper, ms |
| --- | --- | --- | --- | --- | --- |
| go/off/1 | 0.1118 | 2.2780 | 0.784 / 1.330 | 0.114 / 1.001 | 0.000 |
| go/off/4 | 1.7821 | 9.0084 | 1.609 / 3.490 | 1.005 / 4.237 | 0.000 |
| go/basic/0 | 3.1227 | 22.1060 | 0.307 / 6.114 | 3.206 / 6.651 | 83.518 |
| go/basic/1 | 1.6107 | 14.5204 | 1.176 / 2.748 | 2.332 / 4.917 | 85.920 |
| go/basic/4 | 2.8887 | 18.2518 | 1.085 / 2.660 | 2.762 / 4.218 | 69.853 |
| vitest/off/1 | -1.8437 | 7.0688 | -0.796 / 2.299 | -1.458 / 1.344 | 0.000 |
| vitest/off/4 | -1.2236 | 12.0116 | -1.801 / -0.249 | -1.518 / 0.428 | 0.000 |
| vitest/basic/0 | 0.0104 | 19.1570 | -1.370 / 2.423 | -0.132 / 2.130 | 92.508 |
| vitest/basic/1 | 0.1018 | 20.9673 | 0.368 / 1.674 | 1.157 / 2.123 | 92.458 |
| vitest/basic/4 | 1.4184 | 32.9278 | -0.136 / 1.933 | 0.927 / 2.411 | 91.978 |
| shell/off/1 | -0.1666 | 4.7551 | 0.258 / 1.352 | 0.443 / 1.190 | 0.000 |
| shell/off/4 | 0.6623 | 34.6456 | 0.586 / 1.735 | 0.492 / 1.855 | 0.000 |
| shell/basic/0 | 0.7409 | 29.8289 | 0.070 / 1.872 | 0.358 / 2.140 | 100.973 |
| shell/basic/1 | 1.1459 | 25.8835 | 0.528 / 1.319 | 0.710 / 1.541 | 101.597 |
| shell/basic/4 | 0.8819 | 61.3323 | 0.805 / 1.809 | 1.071 / 2.025 | 82.035 |
| service/off/1 | -1.2336 | 2.5269 | 0.150 / 7.595 | 0.224 / 7.279 | 0.000 |
| service/off/4 | -1.9069 | 4.6800 | -0.803 / 0.255 | -0.659 / 0.066 | 0.000 |
| service/basic/0 | -1.0779 | 24.3704 | 1.204 / 10.346 | 1.229 / 10.053 | 100.148 |
| service/basic/1 | -0.8338 | 22.9124 | 0.383 / 8.326 | 0.579 / 8.291 | 106.326 |
| service/basic/4 | -0.8290 | 22.8993 | -1.857 / -0.052 | -1.568 / 0.264 | 99.782 |
| browser/off/1 | -2.1517 | 16.9865 | -0.331 / 8.052 | -0.356 / 7.836 | 0.000 |
| browser/off/4 | -2.3892 | 13.7364 | -0.908 / 4.511 | -0.904 / 4.278 | 0.000 |
| browser/basic/0 | -2.3119 | 42.6615 | -0.469 / 3.456 | -0.365 / 3.489 | 122.078 |
| browser/basic/1 | -1.6994 | 40.1413 | -1.931 / 0.080 | -1.720 / 0.238 | 118.578 |
| browser/basic/4 | -1.5289 | 46.0891 | -0.507 / 8.370 | -0.340 / 8.173 | 128.773 |

S-020's user-authorized CPU-policy implementation remains functionally complete.
F-022/RB-006/S-021 remain open; S-022 is not started and cannot be marked complete.
All budgets retain their adopted values, including CPU mean strictly below 2%.
Default-off, local-only diagnostics and no Core 05 publication claim remain
explicit. Further attribution/repair must save its own S-020 checkpoint before
refreezing; the rejected candidate and all prior attempts remain unchanged.


### S-020 F-022 publication-cost attribution — IN_PROGRESS

Begin after the saved candidate 10 rejection. The user's RAM-disk suggestion
warrants measuring actual publication, which the earlier retained-reader-only
probe did not cover. Use the unchanged clean candidate 10 executable and exact
Go local-migration owner row via public Make, cache mode off. Compare off/basic
with retained output on ext4 versus a newly owned tmpfs directory, without public
readers. Retain one preparation observation per mode/storage cell and eight
balanced measured blocks. Record every command, source, result, signed CPU/wall
observation and failure, then copy RAM evidence durably. This is component-path
attribution with waited-child CPU and external wall timing, not the complete
qualification protocol: no independent subreaper/RSS/reader accounting or
qualification pass is claimed. It changes no executable source, adopted budget
or frozen qualification evidence. Any benefit must inform a structural repair
or an explicitly bounded environment decision; do not relabel candidate 10.


F-022 publication-path attribution completed on unchanged clean candidate 10:
all 36 public Go workload commands passed (one preparation observation per
mode/storage cell, eight balanced four-cell blocks). Exact argv, environment,
source, mode/order, timestamps, result roots and signed waited-child CPU/external
wall measurements are in isolated
`.cartulary/test-results/instrumentation-f022-publication-probe-01/probe.json`.
Each invokes `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.local_migration_activities HARNESS_DIAGNOSTICS=<off|basic> CARTULARY_HARNESS_CACHE_MODE=off`,
with explicit unique `CARTULARY_TEST_RUN_ID` and disk/tmpfs
`CARTULARY_TEST_RESULTS_DIR`. Original RAM root is
`/dev/shm/cartulary-instrumentation-f022-publication-probe-01/`; all 243 files
were durably copied to `retained-ram/` beneath the probe root, with matching
SHA-256 digests in `ram-copy-digests.json`. No failed observations occurred.

Independent `arithmetic.json` preserves every signed pair. Disk basic-minus-off
CPU mean/median/MAD are 58.9725/64.8645/47.5090 ms; RAM values are
5.74275/19.7595/38.3165 ms. The paired RAM-minus-disk effect has mean -53.22975 ms
but median +0.7275 ms and MAD 44.8615 ms: the mean improvement is not a consistent
CPU benefit. The analogous wall effect is mean -97.434676 ms, median -62.870343 ms
and MAD 69.578596 ms. RAM can reduce publication wall time, but this probe does
not establish total observer qualification or justify changing the CPU claim.
Storage remains unchanged. These are waited-child attribution observations,
without qualification subreaper/RSS/reader accounting, not passing qualification.

Next attribution candidate: generated JavaScript validators repeatedly embed
shared schema definitions. The execution record/index/private-launch validators
alone occupy 93,505/122,779/109,975 bytes. Investigate compiler-owned reference
factoring with the pinned AJV generator and existing acceptance parity; do not
introduce a runtime compiler, cache or translator, or weaken validation. S-020
remains open pending measured benefit and functional/projection gates.

The durable RAM-copy audit also passed: `make harness-observability-check RESULTS_DIR=/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary/.cartulary/test-results/instrumentation-f022-publication-probe-01/retained-ram/runs/instrumentation-f022-publication-ram-basic-measured-8` reported one invocation, four sources, complete diagnostics/execution and read-only inspection.

The user clarified that RAM is intended to reduce wall time, not to establish a
CPU improvement. Keep that environment option separate from the shared-validator
repair. Any RAM-backed qualification must freeze the same storage posture for
all treatments and preserve all evidence durably; CPU mean remains strictly below
2% and all other budgets remain unchanged.

### S-020 F-022 shared validator factoring — IN_PROGRESS

The user explicitly authorized continuing this refactor. Extract generation-time
standalone compilation into one narrowly owned compiler and use pinned AJV
reference factoring. Keep each lazy schema module dependency-free with a complete
closure; repeated references share functions within that closure. No schema
acceptance change, runtime compiler or historical translator is authorized.
Validate all registered schemas against the original inline-reference compiler,
plus independent boundary cases for repeated/external/recursive references,
combinators, Unicode, conditionals and evaluated-property closure. Preserve the
pre-refactor generated closure and byte inventory in primary
`.cartulary/test-results/instrumentation-f022-validator-factoring-01/`.

Initial `make generate` passed at primary run
`instrumentation-f022-factoring-generate-01` (9,820 ms). Reference factoring alone
reduced the total generated JS closure from 7,233,968 to 6,763,069 bytes, largely
in performance explanations; execution record/index/private-launch sizes barely
changed. This is a code-size observation, not qualification.

`CARTULARY_TEST_RUN_ID=instrumentation-f022-factoring-contract-01 make harness-contract CARTULARY_HARNESS_CACHE_MODE=off`
failed (197/198 tests passed, graph 104,944 ms, Make exit 2). Retained failure is
`unit-logs/target-harness-contract/stdout.log` under that run: the new parity
fixture passed host-realm objects into a VM-realm validator, and pinned
fast-deep-equal correctly distinguishes those constructors. Fix the fixture to
parse serialized inputs in the same realm as its validator. No production
acceptance mismatch was established. Preserve the failed run.

The compiler also uses first-rejection validation: a valid artifact must still
satisfy the complete schema, while invalid artifacts need an actionable rejection
witness, not exhaustive error allocation in every unsuccessful union arm. Schema
IDs and failure categories stay unchanged; private diagnostic enumeration may
be shorter. Parity compares this policy against the original all-errors/inlined
compiler, with independent authored boundary expectations.

Component attribution completed under
`CARTULARY_TEST_RUN_ID=instrumentation-f022-factoring-profile-01 make harness-contract CARTULARY_HARNESS_CACHE_MODE=off`:
all 198 tests and both graph units passed (105,424 ms). Temporary test-only
profiling launched fresh pinned-Node children for 34 retained public artifacts
from candidate 10, spanning 13 registered schemas, with one preparation pair and
15 alternating measured pairs. The saved original generated closure supplies
the baseline; the current generated closure supplies the candidate. Fixture
bytes, exact child argv, probe source, all observations and independent paired
arithmetic are in `instrumentation-f022-validator-factoring-01/`. This isolates
module loading/validation, not whole-observer CPU/RSS or qualification.

First-load/validation median CPU fell 17,081 to 15,686 microseconds (paired median
-1,433; paired MAD 393). After 100 additional validation passes, median total CPU
fell 40,366 to 35,224 microseconds (paired median -5,445; MAD 1,200). RSS growth
medians were 21,929,984 and 18,788,352 bytes (paired median -3,448,832; MAD 757,760).
Heap growth did not improve consistently. Total generated JS is 6,141,672 bytes,
down from 7,233,968; execution record/index/private-launch are now
90,404/116,851/104,331 bytes. Temporary profiling is removed before final functional
validation and freeze; it is not part of the executable qualification candidate.

For the next qualification, use an explicitly owned `/dev/shm` result parent for
every treatment to evaluate the separate wall-time benefit. Source/tool caches
and service posture remain unchanged. The entire cohort must be durably archived
and digest-verified after the command exits, including failed/interrupted
evidence. Qualification applies only to that recorded RAM-backed output posture;
no disk-output qualification or CPU improvement from RAM is claimed.

### S-020 shared-validator functional checkpoint — DONE

Saved before beginning the successor freeze/qualification. The final authored
change is `schema-validator-compiler.mjs`, extracted from foundation generation,
using pinned AJV reference factoring and first-rejection validation. Generation
still publishes complete lazy dependency-free modules atomically. Make dependency
and authored backing-script routing include the compiler. Expanded parity tests
compare all 195 registered schemas to the original compiler and exercise repeated,
external and recursive references, combinations, nullability, Unicode, conditionals
and evaluated-property closure. Bootstrap isolation and unknown-schema rejection
remain intact. No schema acceptance set, schema ID, budget, workload or cadence
changed; shorter private diagnostic enumeration is the only diagnostic impact.
Native generated schemas/command policy and the qualification plan are byte-equal
to candidate 10. No production files read, stat or hash Markdown.

Final dirty functional source is
`sha256:0fd8868e23c0339ba9f5df3720f2a5615a347b69a0eb33f70d737cdc8cedc33c`
at primary `main` / `1d423c6e5f9cd204f8210fa9958c780fd933d293`, on WSL2 with
existing pinned toolchain/capacity. The source is not a qualification revision.
The pre-existing staged tracker SHA-256 remains
`9ab829e44fa8df7ba1ee86e794518acd97129cea6af416966fbdda4b33f83426`.

All commands below ran from the primary root with the listed explicit run ID.
Result roots are `.cartulary/test-results/<run-id>`. Exact argv, environment,
outer exits and captured output for the sequential validation are retained in
`instrumentation-f022-factoring-validation-01/commands.json` and adjacent logs.

| Run ID | Command | Result |
| --- | --- | --- |
| instrumentation-f022-factoring-generate-02 | `make generate` | PASS, 9,640 ms; final compiler policy |
| instrumentation-f022-factoring-contract-02 | `make harness-contract CARTULARY_HARNESS_CACHE_MODE=off` | PASS, 198 tests / 2 graph units, 105,264 ms; temporary probes removed |
| instrumentation-f022-factoring-evidence-01 | `make harness-evidence-contract` | PASS, outer 23,356 ms |
| instrumentation-f022-factoring-command-01 | `make harness-command-surface-contract` | PASS, outer 55,291 ms |
| instrumentation-f022-factoring-native-01 | `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_observation,harness.evidence_accounting.behavior.instrumentation_measurement` | PASS, 2 rows, outer 35,638 ms |
| instrumentation-f022-factoring-drift-01 | `make generate-drift` | PASS, outer 20,102 ms |
| instrumentation-f022-factoring-policy-01 | `make generated-artifact-policy-check` | PASS, outer 3,229 ms |
| instrumentation-f022-factoring-shapes-01 | `make json-shape-check` | PASS, outer 5,322 ms |
| instrumentation-f022-factoring-lint-01 | `make lint-scripts` | PASS, outer 1,973 ms |
| instrumentation-f022-factoring-basic-01 | `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.local_migration_activities HARNESS_DIAGNOSTICS=basic CARTULARY_HARNESS_CACHE_MODE=off` | PASS, 1 row, canonical 1,167 ms |

For that exact basic run, `make harness-observability-check RESULTS_DIR=.cartulary/test-results/instrumentation-f022-factoring-basic-01`
passed with one invocation/four sources and complete diagnostics/execution.
`make explain-run RESULTS_DIR=.cartulary/test-results/instrumentation-f022-factoring-basic-01 DETAIL=performance`
and `DETAIL=resources`, each human and `JSON=1`, passed; exact commands and output
are retained under the validation root. The initial cross-realm parity-fixture
failure and all earlier qualification failures remain retained.

The measured component benefit justifies carrying this structural simplification
forward but does not establish total observer qualification. F-022/RB-006 remain
open. S-021 must freeze a clean executable revision and run every required cell
with CPU mean strictly below 2% and all other gates unchanged. Retained output
will use an owned RAM directory consistently, then be archived durably; the
physical coverage claim is limited to that storage posture. S-022 remains gated.
Rollback is the compiler/generator/test/Make/backing-input/generated-JS transaction;
retain failed evidence and never restore runtime compilation or historical schema
translation. No finalizer or S-022 completion is claimed at this checkpoint.

### S-021 candidate 11 freeze — IN_PROGRESS

Begins after the saved S-020 shared-validator functional checkpoint. Clean isolated
branch `codex/instrumentation-qualification-11` is frozen at
`6713042e9625e41756954575a2aa3c3bbfe2d346` (parent candidate 10).
The 202-file delta is 195 regenerated validator modules plus the compiler,
generator, parity test, Make dependency, authored task surface and its two
generated projections. No acceptance schema, budget, workload or cadence changed.
Freeze metadata is `/tmp/cartulary-qualification-freeze-11.json`, also retained in
primary `instrumentation-f022-validator-factoring-01/freeze-11.json`. Primary
index bytes and the pre-existing staged tracker were preserved exactly during
freezing; the isolated checkout is clean.

Run `CARTULARY_TEST_RUN_ID=instrumentation-s021-qualification-11 CARTULARY_TEST_RESULTS_DIR=/dev/shm/cartulary-instrumentation-qualification-11 make harness-instrumentation-qualify QUALIFICATION_PLAN=tools/harness_instrumentation_qualification_plan.json`
from `/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary`.
All treatments use this same owned RAM result parent; source, installed tools,
build caches, private runtime storage and service policy remain unchanged.
The recorded mount is existing tmpfs `/dev/shm`, approximately 16 GiB capacity;
candidate 10 retained about 69 MiB. No mount or global cache changes are made.

The native execution wrapper records exact argv/environment, elapsed time, outer
exit and stdout/stderr durably at isolated
`.cartulary/test-results/instrumentation-s021-qualification-11/`. Once Make and its
owned children have exited, archive the complete RAM run into that durable root
and verify every file digest, including any failure/interruption. Do not copy
during measurement or modify source, run other tests/generators, or profile the
active cohort. The qualification manifest supplies its executable source digest;
record it without substituting dirty functional-source evidence. CPU mean must
remain strictly below 2%; every other budget, coverage and six-block stability
rule remains unchanged. S-022 remains unstarted and gated.

Candidate 11 is running in native session `78565`; the externally recorded Make
PID is `10830` (identity context is in the command journal, not inferred ownership).
The published manifest confirms clean source commit `6713042e9625e41756954575a2aa3c3bbfe2d346`,
source digest `sha256:d92dd9b42621add7f851ac53c39258062428180be6673e52bbd16cc5c81df08e` and
unchanged v2 plan digest `sha256:1152b853dc9f2c8ceeb20343245e305806b729ac413f8cf0a686118627ba38af`.
RAM manifest is
`/dev/shm/cartulary-instrumentation-qualification-11/instrumentation-s021-qualification-11/qualification/qualification.json`.
Durable command journal/logs are in isolated
`.cartulary/test-results/instrumentation-s021-qualification-11/`; the native
wrapper archives and digest-verifies the RAM artifacts after Make exits. Initial
17 Go attempts had no command/measurement failures. No qualification pass is yet
claimed, and all previous failed evidence remains unchanged.

Candidate 11 progress checkpoint: Go, Vitest and shell each completed 48 attempts
(six measured blocks plus cold/warm-up). All 150 completed attempts, including
the first six service observations, had successful exits, non-null measurements,
null failure and every coverage flag true. Independent preliminary per-cell
arithmetic shows every CPU mean in these three completed workloads below 2%:
maxima Go 1.2107%, Vitest 1.5528%, shell 0.7627%. These are this combined
source/storage cohort's observations, not an isolated causal CPU claim for RAM.

Known numeric failures are Go canonical elapsed p90 off/1 = 3.0729%, basic/0 =
3.4241%, basic/4 = 3.5119%; Vitest basic/0 outer elapsed p90 = 3.1771%; shell
basic/4 RSS upper estimate = 32.7069 MiB. Every other preliminary CPU/RSS/elapsed
cell in these three workloads passes its numeric threshold. This is not a
qualification disposition: finalization maxima and six-block stability remain
for the independent final validator, and service/browser collection continues.
No observation was discarded, no threshold changed and no executable work was
performed during collection. The complete cohort will be retained, including
these unfavorable observations; S-022 remains gated.


### S-021 candidate 11 complete-matrix rejection checkpoint — NOT QUALIFIED

The RAM-output cohort completed all 240 attempts (48 per workload; six measured
blocks each) in 6,961,828 ms, Make exit 2. All attempts have successful workload
exits, null failure, non-null measurements and complete declared coverage. The
manifest status is `failed`. No valid unfavorable observation was excluded or
replaced. The isolated checkout remains clean at
`6713042e9625e41756954575a2aa3c3bbfe2d346`, source
`sha256:d92dd9b42621add7f851ac53c39258062428180be6673e52bbd16cc5c81df08e`,
plan `sha256:1152b853dc9f2c8ceeb20343245e305806b729ac413f8cf0a686118627ba38af`.
Inspection after exit found no remaining qualification runner, meter or gate helper.

Make wrote retained output to `/dev/shm/cartulary-instrumentation-qualification-11/`;
all other frozen source/tool/cache/runtime/service preparation remained unchanged.
The supervising wrapper copied and digest-verified all 8,285 files into isolated
`.cartulary/test-results/instrumentation-s021-qualification-11/` only after Make
exited. `archive-command.json` records the mount, source, argv, environment,
start/end, elapsed time, exit and `archive_verified: true`; `archive-digests.json`
records each file. Original RAM evidence remains present. RAM addresses wall time;
these combined source/storage observations do not establish a causal CPU benefit.

Independent validation used exactly
`CARTULARY_TEST_RUN_ID=instrumentation-s021-validation-11 make harness-performance-check EVIDENCE_ROOTS_FILE=/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary/.cartulary/test-results/instrumentation-s021-qualification-11/qualification/qualification.json`.
It returned timing exit 13 / Make 2 in 1,132 ms: all five workloads remain unstable
after six blocks and the manifest is not complete. All 30 cells are eligible;
14 of 25 non-reference cells fail numeric gates. CPU arithmetic means pass
throughout (maximum 1.5528%); three RSS upper estimates, eleven canonical p90,
nine outer p90, three canonical median and three outer median gates fail.
Finalization passes throughout; the maximum is 106.581 ms against 500 ms.

The validator command journal/stdout/stderr and independently computed signed
pairs, error diagnostics, finalization unions and per-gate leave-one-out changes
are retained at isolated `.cartulary/test-results/instrumentation-s021-validation-11/`.
`independent-arithmetic.json` and `independent-table.md` agree with the public
validator. Negative differences remain signed and do not establish improvement.

| Workload/mode/readers | CPU mean, % one core | RSS upper, MiB | Canonical median / p90, % | Outer median / p90, % | Finalization upper, ms |
| --- | --- | --- | --- | --- | --- |
| go/off/1 | -0.3152 | 6.7627 | 0.263 / 3.073 | 0.021 / 2.046 | 0.000 |
| go/off/4 | 0.5132 | 7.6614 | 0.827 / 1.836 | 0.330 / 1.659 | 0.000 |
| go/basic/0 | 0.8510 | 25.4443 | 1.483 / 3.424 | 0.355 / 1.928 | 62.754 |
| go/basic/1 | 1.1056 | 16.8449 | 1.962 / 2.342 | 1.208 / 2.815 | 66.833 |
| go/basic/4 | 1.2107 | 22.9133 | 0.607 / 3.512 | 0.993 / 2.400 | 62.787 |
| vitest/off/1 | 0.5614 | -0.9056 | -0.203 / 2.144 | 0.796 / 2.282 | 0.000 |
| vitest/off/4 | 0.7250 | 1.6492 | 1.025 / 1.698 | 0.860 / 2.146 | 0.000 |
| vitest/basic/0 | 1.5528 | 7.4733 | 0.472 / 2.546 | 1.814 / 3.177 | 63.189 |
| vitest/basic/1 | 1.2873 | 18.4269 | 0.534 / 1.040 | 1.184 / 2.534 | 62.503 |
| vitest/basic/4 | 1.4516 | 9.0795 | -0.327 / 2.554 | 0.974 / 2.690 | 61.468 |
| shell/off/1 | 0.4089 | 18.6308 | 0.024 / 1.048 | 0.110 / 1.392 | 0.000 |
| shell/off/4 | 0.2692 | 6.2610 | 0.191 / 0.745 | 0.186 / 0.628 | 0.000 |
| shell/basic/0 | 0.6185 | 21.0773 | -0.039 / 0.503 | 0.057 / 0.913 | 62.928 |
| shell/basic/1 | 0.5301 | 16.9521 | 0.293 / 0.773 | 0.201 / 0.959 | 58.939 |
| shell/basic/4 | 0.7627 | 32.7069 | 0.233 / 1.038 | 0.235 / 1.091 | 57.950 |
| service/off/1 | 0.1368 | 3.3823 | 2.792 / 10.636 | 2.989 / 10.142 | 0.000 |
| service/off/4 | -1.2162 | 6.0529 | -0.513 / 11.108 | -0.528 / 10.753 | 0.000 |
| service/basic/0 | 0.6102 | 20.4798 | 2.202 / 11.983 | 2.232 / 11.738 | 90.356 |
| service/basic/1 | -0.1714 | 27.7945 | 0.156 / 1.790 | 0.336 / 2.054 | 88.114 |
| service/basic/4 | -0.2052 | 21.2154 | -0.346 / 14.840 | -0.275 / 14.208 | 89.841 |
| browser/off/1 | -0.4505 | 22.2638 | -0.328 / 3.938 | -0.343 / 3.665 | 0.000 |
| browser/off/4 | -0.9406 | 13.1399 | 1.428 / 8.383 | 1.437 / 8.069 | 0.000 |
| browser/basic/0 | 0.2962 | 40.4375 | -0.328 / 1.405 | -0.394 / 1.323 | 106.581 |
| browser/basic/1 | -1.8962 | 30.7620 | 5.739 / 9.277 | 5.404 / 9.031 | 102.909 |
| browser/basic/4 | -0.8716 | 39.9801 | 0.653 / 6.978 | 0.645 / 6.710 | 102.845 |

This saves the failed S-021 disposition before reopening implementation. F-022
and RB-006 remain open; S-022 has not begun. All budgets remain unchanged,
including strict CPU mean <2%. A further candidate requires an evidence-backed
repair, complete S-020 functional checkpoint and clean freeze. Default-off,
local-only diagnostics and absence of a Core 05 publication claim remain explicit.


### S-020 post-candidate-11 structural repair entry — IN_PROGRESS

The complete failed S-021 checkpoint is saved above. Investigate sharing generated
compiled references across modules, including shared definition fragments, while
preserving pinned AJV acceptance, lazy loading, complete generated closure and
package-free bootstrap. Current per-module factoring still duplicates referenced
validators across consumers. Confine compiler linkage to generation and reject
unresolved runtime dependencies; independent recursion/composition/parity tests
must establish acceptance before adoption. Also characterize the duplicated
initial private execution publication and service setup cost before choosing any
further repair. Do not change schemas, workload, reader cadence, measurement or
budgets. Candidate 11 and all prior evidence remain untouched. S-022 is gated.

First linked-validator attempt: `CARTULARY_TEST_RUN_ID=instrumentation-f022-linking-generate-01 make generate` passed (10,090 ms). `CARTULARY_TEST_RUN_ID=instrumentation-f022-linking-contract-01 make harness-contract CARTULARY_HARNESS_CACHE_MODE=off` failed (1/2 units, 101,723 ms, Make 2). The independent shared-fragment assertion found that compiling roots before fragment entry points left separate compiled function identities, so the intended fragment module was never loaded. Pinned acceptance comparisons passed up to that assertion; no acceptance change is inferred. Preserve the root and `unit-logs/target-harness-contract/stdout.log`. Correct generation ordering by resolving fragment entry points first through AJV's registry, then rerun. No runtime policy, schema or budget change.

Second linked-validator attempt: `instrumentation-f022-linking-generate-02` passed `make generate` (9,840 ms). `instrumentation-f022-linking-contract-02` failed `make harness-contract CARTULARY_HARNESS_CACHE_MODE=off` (1/2 units, 102,498 ms, Make 2). Shared-fragment, recursion, dynamic-anchor, composition and acceptance checks passed; the bootstrap fixture still required exactly one loaded module and rejected the now-intended shared dependencies. Replace that obsolete count with the independently authored expected summary reference closure (summary, free-object, extensions), plus the owner registry after its separate load; unrelated roots remain forbidden. All failed roots/logs are retained. Private execution publication now validates and bounds its initial record once, removes its duplicate replacement, and validates proof-bearing updates through the complete envelope once; boundary coverage includes oversized beginnings and rejected replacements preserving the incomplete fact.

`instrumentation-f022-linking-profile-01` passed `make harness-contract CARTULARY_HARNESS_CACHE_MODE=off` (2/2 units, 104,792 ms), including the temporary Make-owned component probe. The 15 measured alternating pairs show that sharing small fragments worsened retained-document first-validation CPU (median 16,277 → 20,199 µs; paired median +3,973/MAD 686 µs), total CPU (+3,859/MAD 1,215 µs), and RSS (+1,015,808/MAD 446,464 bytes). Generated byte savings are not a performance claim. Preserve this failed optimization's full generated closure, compiler, commands, observations and independent arithmetic at primary `.cartulary/test-results/instrumentation-f022-linked-validators-01/`. Narrow sharing to registered roots, keep small fragments local, and compare both retained-reader and producer use (record plus private-envelope loading). These are component experiments, not qualification; no threshold changes.

Root-only linkage: `instrumentation-f022-linking-generate-03` passed `make generate` (9,850 ms); `instrumentation-f022-linking-profile-02` passed `make harness-contract CARTULARY_HARNESS_CACHE_MODE=off` (2/2 units, 107,592 ms). The component test retains 96 observations: one warm-up and 15 measured blocks for retained and producer datasets across baseline, fragment-sharing and root-sharing variants. Producer data adds independently authored private proof and execution records to the retained fixture set. Root sharing yields first-validation CPU paired median −3,905/MAD 328 µs and RSS −5,980,160/MAD 1,830,912 bytes; retained-document first CPU rises +597/MAD 435 µs and RSS +581,632/MAD 270,336 bytes. Adopt root sharing for the producer benefit and explicit shared contract implementation; reject separate fragment modules, whose loading cost erased the benefit. All raw observations and signed arithmetic remain at `root-linking-measurements.json`/`root-linking-arithmetic.json`. Temporary profiling code was removed; this is not a functional completion or qualification checkpoint.

Public exact-run performance explanations for candidate 11 service outliers are retained with commands under the same component root (`retained-explanation-commands.json`, `retained-outlier-commands.json`). Fixture acquisition is 9,420–9,490 ms in four selected outliers versus 4,649–4,703 ms in matched references. Public service-scope facts show successful single container-start attempts; the extra interval falls after object-store container startup. Cause remains under investigation; do not label it reader or collector overhead. The S3 transport already disables SDK retries (`RetryMaxAttempts: 1`), so nested SDK retry is not an established defect.

S-020 readiness attribution starts after the root-linking component run drained. A temporary, locally retained numeric-only S3 operation trace records stage durations for four exact public service-backed migration selections. It changes no retry, timeout, readiness acceptance, selected test or service reuse behavior, and must be removed before functional verification/freeze. Original/instrumented source and command evidence are preserved at primary `.cartulary/test-results/instrumentation-f022-readiness-attribution-01/`. This is attribution, not qualification; do not treat a faster individual run as acceptance.

All four original readiness attribution commands passed, outer elapsed 49,280 / 46,811 / 51,057 / 46,666 ms. The detailed trace narrows the slow third run to an initial PUT of 4,960 ms, versus about 200 ms in the others, in addition to initial LIST polling; initial readiness was 8,494 ms versus about 3,740 ms. No readiness deadline or retry changed. This corrects any earlier suggestion that all delay was between fast calls.

The pinned [SeaweedFS 4.48 server command](https://raw.githubusercontent.com/seaweedfs/seaweedfs/4.48/weed/command/server.go) forwards `master.peers` to its master and actual component addresses. Its [master startup](https://raw.githubusercontent.com/seaweedfs/seaweedfs/4.48/weed/command/master.go) explicitly bootstraps a fresh single master immediately when peers are `none`; the current disposable fixture leaves the peer configuration implicit. Test this explicit single-master topology with four more exact selections, preserving all original observations and the same readiness mutation/cleanup/deadline contract. This is a causal hypothesis and component experiment, not a qualification pass or a claim that RAM fixes CPU cost.

All four explicit single-master attribution commands passed, outer elapsed 47,348 / 45,443 / 45,193 / 45,852 ms. `all-arithmetic.json` retains every stage duration and both variants. The new configuration's initial readiness stayed near 2.74 seconds in these four observations; this supports the bounded startup repair, but does not prove qualification stability or eliminate every possible provider delay. Restore the original source, retaining only `-master.peers=none` and its explanation; all temporary trace code is removed. The adopted owner now makes the existing disposable one-master topology explicit without changing readiness, deadlines, retry, ownership or cleanup. The existing container-configuration fixture is named accordingly and assigned to `harness.browser.unit.object_store_fixture_admission`; generate routing from the authored family. Shared root validators, single initial publication, and explicit standalone startup are now ready for functional/projection validation. S-020 remains open until those gates pass; S-021 requires a fresh full freeze and S-022 is unstarted.


### S-020 registered-root sharing and bounded startup checkpoint — DONE

Functional completion is saved before the next S-021 freeze. The adopted change
shares exact compiled registered-root validator identities through generated
sibling modules. Small fragments remain local after their measured regression;
no runtime compiler, new validation cache, schema acceptance change or historical
translator is introduced. The generated closure is 195 modules / 5,604,419 bytes.
Pinned acceptance, recursive/dynamic references, both mutual-cycle load orders,
lazy closure and package-free bootstrap checks pass. Producer component benefit
and the small retained-reader regression remain recorded above; neither is a
qualification claim.

Private initial execution publication now validates/serializes once, enforces the
byte bound before reserving a slot, and avoids the immediate duplicate write.
One encoder governs initial, ordinary replacement and resumed replacement, with
one complete envelope validation for proof-bearing records. Independent oversized
and malformed replacement cases preserve the incomplete prior fact. The resumed
byte-bound gap was found in final review and repaired before freeze.

Disposable object-store startup explicitly uses `-master.peers=none`, matching
its existing single-master ownership topology. The adopted owner and routed
container-configuration fixture reflect that requirement. Readiness mutations,
timeouts, retry, cleanup, workload selections and reuse policy are unchanged.
All temporary component/readiness instrumentation is removed from executable
source; original sources, both experiments, rejected fragment closure and failed
contract runs remain in the recorded result roots.

Primary validation journals are
`.cartulary/test-results/instrumentation-f022-linking-validation-01/commands.json`
and `instrumentation-f022-linking-validation-02/commands.json`; each records exact
argv, run ID, cwd, start, outer elapsed and exit, with full stdout/stderr. All
commands listed below passed. Source before the final resumed-bound repair was
`sha256:9d8fb11b5bdc641e262985684507ecce90901e0be06ac9c642a0e9d2c5ed4077`;
the fresh basic run after that repair records `sha256:edb3eb4ddace6c0aa4cb1d79941c0b22c8c520459fbe0181be7cd3a056b7011d`
at primary HEAD `1d423c6e5f9cd204f8210fa9958c780fd933d293`, dirty executable state.
This functional identity is not the forthcoming clean qualification identity.

| Run ID suffix after `instrumentation-f022-linking-` | Exact public operation | Result |
| --- | --- | --- |
| generate-04 | `make generate` | PASS, 9,840 ms |
| final-guide-evidence-01 / final-guide-browser-01 | `make task-guide ROLE=module-author OWNER=harness.evidence_accounting` / `OWNER=harness.browser` | PASS |
| final-evidence-01 | `make harness-evidence-contract` | PASS, 22,084 ms |
| final-command-01 | `make harness-command-surface-contract` | PASS, 53,442 ms |
| final-contract-01 | `make harness-contract CARTULARY_HARNESS_CACHE_MODE=off` | PASS, 105,682 ms |
| final-native-01 | `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_observation,harness.evidence_accounting.behavior.instrumentation_measurement CARTULARY_HARNESS_CACHE_MODE=off` | PASS, 34,617 ms |
| final-services-01 | `make service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.unit.object_store_fixture_admission,harness.browser.integration.object_store_fixture_roundtrip,harness.browser.unit.testservices_lifecycle_contract,harness.browser.boundary_support.https_trust HARNESS_DIAGNOSTICS=basic CARTULARY_HARNESS_CACHE_MODE=off` | PASS, 58,612 ms |
| final-drift-01 / final-policy-01 / final-shapes-01 | `make generate-drift` / `make generated-artifact-policy-check` / `make json-shape-check` | PASS, 19,505 / 3,097 / 5,240 ms |
| final-lint-01 | `make lint-scripts` | PASS, 1,977 ms |
| evidence-02 | `make harness-evidence-contract` after resumed-bound repair | PASS, 22,256 ms |
| basic-02 | `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.local_migration_activities HARNESS_DIAGNOSTICS=basic CARTULARY_HARNESS_CACHE_MODE=off` | PASS, 3,460 ms |
| audit-02 / services-audit-02 | `make harness-observability-check RESULTS_DIR=.cartulary/test-results/instrumentation-f022-linking-basic-02` / exact `instrumentation-f022-linking-final-services-01` root | PASS |
| performance-human-02 / performance-json-02 / resources-human-02 / resources-json-02 | `make explain-run RESULTS_DIR=.cartulary/test-results/instrumentation-f022-linking-basic-02 DETAIL=performance` or `DETAIL=resources`, human and `JSON=1` | PASS |
| lint-02 | `make lint-scripts` after resumed-bound repair | PASS, 1,951 ms |

Migration impact: generated consumers are regenerated together; published schema
versions/acceptance remain unchanged. Oversized resumed private observations now
retain the incomplete beginning instead of publishing unreadable evidence. There
is no retained-format translation or runtime default change. Roll back the
compiler/generator/tests/generated closure, execution writer/tests and fixture
configuration/routing/owner clarification as a coherent source transaction while
preserving every attempt. Source and generated projections remain downstream of
the adopted owners; tooling does not consume this tracker.

F-022/RB-006 remain open. S-021 must freeze the changed executable source and
collect the complete unchanged matrix; RAM retained output addresses wall time,
not a CPU guarantee. CPU mean remains strictly below 2%, with all other budgets
and stability/coverage rules unchanged. Retained-reader overhead, provider noise
and total RSS remain qualification risks. S-022 has not begun. Default-off,
local-only diagnostics and no Core 05 publication claim remain explicit.


### S-021 candidate 12 freeze — IN_PROGRESS

Begins after the saved S-020 completion checkpoint. Clean isolated branch
`codex/instrumentation-qualification-12` is frozen at
`d36682b60ba9088f7f7c4e7edd06452bf16f0093`, parent candidate 11. The 25-file
delta contains registered-root validator linkage and its affected generated
modules, bounded private publication/tests, and single-master fixture startup
with owner/test-routing projections. Schema acceptance, measurement plan,
budgets, cadence and workload selection are unchanged. Freeze metadata is
`/tmp/cartulary-qualification-freeze-12.json`, also retained at primary
`instrumentation-f022-linked-validators-01/freeze-12.json`. The primary index and
pre-existing staged tracker were preserved byte-for-byte; isolated status is clean.

Run `CARTULARY_TEST_RUN_ID=instrumentation-s021-qualification-12 CARTULARY_TEST_RESULTS_DIR=/dev/shm/cartulary-instrumentation-qualification-12 make harness-instrumentation-qualify QUALIFICATION_PLAN=tools/harness_instrumentation_qualification_plan.json`
from `/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary`.
Use the same existing tmpfs output posture for every treatment, with source,
build cache, private runtime and service preparation governed by the unchanged
plan. RAM addresses wall time; no causal CPU benefit is claimed. Record source
and plan digests from the manifest. The native supervisor retains command identity,
mount/storage, stdout/stderr, outer elapsed/exit, then archives and digest-verifies
the complete run only after Make exits. Preserve every failure and original RAM
artifact. No executable edits, competing tests, generation or profiling during
collection. S-022 remains unstarted and gated by every required S-021 cell.


Candidate 12 active native session is `55282`, supervising Make PID `97210` as
recorded in the command journal. The manifest confirms clean commit
`d36682b60ba9088f7f7c4e7edd06452bf16f0093`, source
`sha256:769d0309925149d37c3a8921902e53bf9e93670b2ad326b090cd5a6a1af0ca3e`,
unchanged plan
`sha256:1152b853dc9f2c8ceeb20343245e305806b729ac413f8cf0a686118627ba38af`.
RAM manifest:
`/dev/shm/cartulary-instrumentation-qualification-12/instrumentation-s021-qualification-12/qualification/qualification.json`.
Durable command journal/stdout/stderr are at isolated
`.cartulary/test-results/instrumentation-s021-qualification-12/`; the supervisor
will archive and verify all retained files after Make exits.

Progress at 117 published attempts: Go and Vitest each completed 48 (six measured
blocks); shell has 21. All Go/Vitest command exits, measurements and coverage
flags are valid. Independent preliminary arithmetic finds Go off/4 CPU mean
2.1843%, so this candidate already cannot qualify. Go's other CPU means and all
RSS estimates pass; several outer elapsed gates and basic/1 canonical p90 fail.
Vitest CPU means and RSS estimates pass; off/4 and basic/4 canonical p90 and
basic/0, basic/1, basic/4 outer p90 fail. These are preliminary cell calculations,
not a final eligibility/stability disposition. Collection continues through the
full service/browser matrix to evaluate the explicit single-master fixture
change; no valid unfavorable observation is discarded or replaced. Final
independent public validation and complete archive remain required before any
new S-020 repair. S-022 is still unstarted and gated.


Candidate 12 shell checkpoint: all 48 attempts are retained. Measured CPU means,
RSS upper estimates and elapsed numeric gates are within bounds; shell basic/0
RSS is 31.4566 MiB and basic/4 is 31.5401 MiB. This does not establish eligibility:
`shell-cold-0-basic-1` returned Make 2 / `command_failed`, with complete measurement
coverage. Its run summary is product `test_assertion_failure`; terminal log
`qualification/runs/qualification-shell-cold-0-basic-1/unit-logs/row-harness.evidence_accounting.behavior.current_epoch_evidence/stdout.log`
reports 100 passing/1 failing tests. The failing sibling-collector quietness test
expected aggregate process CPU null but observed 0.4982561036372696. Correction
to initial observation wording: coverage was complete, and this is an aggregate
assertion, not direct evidence of a cross-pause delta in the first resumed sample.

Read-only inspection confirms the accumulator requires equal segment and
identity before accumulating CPU. The real-worker fixture waits for all four
collectors before pausing/stopping and then assumes every segment has only one
sample. Investigate that asynchronous assumption with an independent per-segment
oracle after this cohort drains; do not weaken the pause boundary or reclassify
the failed cold attempt as successful. No source/test changes during collection.
At this checkpoint service has ten cold/warm-up observations; the full cohort,
archive and independent public validator are still pending.


### S-021 candidate 12 complete-matrix rejection checkpoint — NOT QUALIFIED

Candidate 12 completed all 240 attempts (48/workload, six measured blocks each),
outer Make exit 2, elapsed 6,784,933 ms. Frozen commit is
`d36682b60ba9088f7f7c4e7edd06452bf16f0093`, source
`sha256:769d0309925149d37c3a8921902e53bf9e93670b2ad326b090cd5a6a1af0ca3e`,
plan `sha256:1152b853dc9f2c8ceeb20343245e305806b729ac413f8cf0a686118627ba38af`.
The isolated checkout remains clean. Native session `55282` completed; its
supervisor copied and digest-verified all 8,285 files after Make exited. Durable
root is `/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary/.cartulary/test-results/instrumentation-s021-qualification-12/`.
`archive-command.json` and `archive-digests.json` retain the exact invocation,
mount/storage, source, PID, elapsed/exit and every file hash. Original RAM evidence
remains at `/dev/shm/cartulary-instrumentation-qualification-12/`.

Independent public validation:
`CARTULARY_TEST_RUN_ID=instrumentation-s021-validation-12 make harness-performance-check EVIDENCE_ROOTS_FILE=/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary/.cartulary/test-results/instrumentation-s021-qualification-12/qualification/qualification.json`
returned timing exit 13 / Make 2 in 1,107 ms. Command journal/stdout/stderr and
independent signed pairs, CPU error diagnostics, interval unions, same-reader
comparisons and per-gate leave-one-out changes are at isolated
`.cartulary/test-results/instrumentation-s021-validation-12/`.

All 30 measured cells are eligible; 239/240 complete attempts are eligible.
The shell cold basic/1 command failure remains independently disqualifying.
Seventeen of 25 non-reference cells fail numeric gates: two CPU means, four RSS
upper estimates, eight canonical p90, thirteen outer p90 and three outer median.
No canonical median or finalization gate fails; maximum finalization is
444.507 ms. Go, Vitest, shell and service remain unstable after six blocks.
Browser is stable but its three basic-mode RSS cells fail. The independent
arithmetic agrees with the public validator. No valid unfavorable observation was
excluded or replaced; negative differences remain signed without claiming benefit.

| Workload/mode/readers | CPU mean, % one core | RSS upper, MiB | Canonical median / p90, % | Outer median / p90, % | Finalization upper, ms |
| --- | --- | --- | --- | --- | --- |
| go/off/1 | -0.1997 | 11.8352 | -0.307 / 2.827 | 0.996 / 3.496 | 0.000 |
| go/off/4 | 2.1843 | 5.9831 | -0.264 / 2.473 | 2.305 / 3.384 | 0.000 |
| go/basic/0 | 1.5797 | 9.2753 | 0.703 / 2.366 | 2.280 / 4.407 | 60.708 |
| go/basic/1 | 1.3616 | 16.1240 | 1.760 / 3.687 | 1.585 / 5.864 | 57.544 |
| go/basic/4 | 1.8165 | 20.0668 | 0.440 / 1.753 | 2.162 / 4.602 | 60.712 |
| vitest/off/1 | 0.6966 | 10.6678 | 0.001 / 1.332 | -0.167 / 2.480 | 0.000 |
| vitest/off/4 | 0.6826 | 8.1921 | 0.964 / 4.003 | 0.275 / 2.989 | 0.000 |
| vitest/basic/0 | 1.4070 | 22.7057 | 0.776 / 2.594 | 1.137 / 3.134 | 59.444 |
| vitest/basic/1 | 1.9186 | 21.3095 | 1.695 / 2.498 | 1.562 / 3.991 | 59.729 |
| vitest/basic/4 | 1.8693 | 18.3090 | 0.855 / 3.746 | 1.346 / 3.796 | 61.723 |
| shell/off/1 | 0.2080 | 9.4433 | 0.297 / 0.751 | 0.318 / 0.597 | 0.000 |
| shell/off/4 | 0.3203 | 13.2115 | -0.051 / 1.346 | 0.093 / 1.463 | 0.000 |
| shell/basic/0 | 0.6649 | 31.4566 | 0.053 / 0.951 | 0.453 / 2.168 | 444.507 |
| shell/basic/1 | 0.9611 | 28.9180 | 0.527 / 0.730 | 0.750 / 0.918 | 58.073 |
| shell/basic/4 | 0.7125 | 31.5401 | 0.068 / 0.497 | 0.273 / 0.776 | 56.903 |
| service/off/1 | 1.3156 | 9.9386 | 0.155 / 6.584 | 0.119 / 6.241 | 0.000 |
| service/off/4 | 2.2393 | 11.3598 | 0.657 / 15.981 | 0.787 / 15.209 | 0.000 |
| service/basic/0 | 1.8928 | 36.0148 | 1.485 / 6.745 | 1.557 / 6.488 | 90.879 |
| service/basic/1 | 1.3377 | 28.9238 | -0.207 / 5.954 | -0.141 / 6.040 | 90.737 |
| service/basic/4 | 0.3595 | 28.3408 | -0.473 / 9.509 | -0.377 / 9.271 | 96.435 |
| browser/off/1 | -0.0002 | 17.0071 | 0.022 / 1.167 | 0.032 / 1.135 | 0.000 |
| browser/off/4 | 0.4327 | 11.9907 | 0.484 / 2.983 | 0.371 / 2.824 | 0.000 |
| browser/basic/0 | 0.2717 | 33.9778 | 0.018 / 0.622 | 0.091 / 0.997 | 107.065 |
| browser/basic/1 | 0.6641 | 42.4988 | 0.181 / 1.580 | 0.197 / 1.669 | 104.951 |
| browser/basic/4 | 0.3696 | 41.0330 | 0.099 / 0.995 | 0.098 / 1.188 | 107.623 |

The single-master change did not eliminate service timing outliers in this full
cohort. Shared-root component gains do not establish a total-observer pass.
RAM was used consistently for retained-output wall time; this is not an isolated
causal CPU claim. Concurrent Windows/other-WSL activity was queried but remains
unknown; do not attribute the failures to it without evidence. Physical coverage
still excludes Windows-host and Docker-daemon internal cost, and sampled RSS is
not an unsampled peak guarantee. All budgets and the typed v2 plan are unchanged.

This saves the failed S-021 disposition before implementation resumes. S-022 has
not begun; no finalizer or final handoff completion is claimed. F-022/RB-006 stay
open and the overall effort remains IN_PROGRESS. Earlier commits, failed
optimizations, command failures, attempts and archives remain untouched.

### S-020 post-candidate-12 repair entry — IN_PROGRESS

**F-025: the real sibling-collector quietness test assumes one sample per segment.**
Owner C-013 / REQ-833/838 observation quietness; S-020 / V-020 and the existing
`harness.evidence_accounting.behavior.current_epoch_evidence` route. The cold
basic/1 assertion failure is preserved above. The test waits for all four
independently scheduled collectors, then expects aggregate CPU null. Legitimate
within-segment pairs may already exist while waiting for the last collector.
The accumulator explicitly refuses cross-segment deltas; direct retained segment
boundaries and independent multi-sample arithmetic must distinguish this from a
producer defect. Repair the assertion without adding waits, lengthening the
qualification workload, weakening quietness, changing cadence or budgets.

Remaining F-022 needs evidence-backed attribution, especially service/browser
RSS and CPU/elapsed variability present even in off/4 treatments. A source change
must be followed by narrow functional verification, a saved S-020 checkpoint and
fresh affected qualification evidence. Do not rerun unchanged failed cells for
luck or silently reinterpret the qualified source. Default-off, local-only
resource diagnostics and no Core 05 claim remain explicit.


F-025 functional repair checkpoint: the real sibling fixture now consumes the
validated retained sample stream and checks that both sides of the acknowledged
pause are present, every post-pause process sample has a newer segment, and a
continuing process identity spans that boundary. It no longer assumes one sample
per segment while waiting for independent siblings. A separate authored fixture
has 300 ms CPU in the first observed second and 400 ms in the second, separated
by a three-second pause containing 1,000 ms CPU; expected utilization is exactly
35%, with only 500,000 user / 200,000 system µs accumulated over 2,000 ms. This
independent oracle retains valid within-segment deltas and rejects bridging the
pause. No waits, deadlines, cadence, production code, schemas or budgets changed.

All passed: `CARTULARY_TEST_RUN_ID=instrumentation-f025-segments-evidence-01 make harness-evidence-contract`
(20,922 ms), `CARTULARY_TEST_RUN_ID=instrumentation-f025-segments-cold-01 make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.current_epoch_evidence HARNESS_DIAGNOSTICS=basic CARTULARY_HARNESS_CACHE_MODE=cold`
(23,447 ms), and `CARTULARY_TEST_RUN_ID=instrumentation-f025-segments-lint-01 make lint-scripts`
(1,949 ms). Exact commands/outer exits/stdout/stderr are in primary
`.cartulary/test-results/instrumentation-f025-segments-validation-01/`.
F-025 is functionally repaired; candidate 12's cold failure remains a failure.
S-020 remains open for F-022 attribution, with S-021/S-022 uncompleted.


F-022 attribution follow-up: the user reports **no other intensive work** on
Windows or other WSL sessions during qualification. This supersedes the earlier
unknown host-activity note; no host-contention cause is established.

Public exact-run explanations for service measured-2 off/0 and basic/0, service
measured-5 off/0, off/4 and basic/4, and browser measured-5 basic/1 all passed.
Commands and human-independent JSON output are in primary
`.cartulary/test-results/instrumentation-f022-candidate12-attribution-01/`.
The service measured-2 basic/0 canonical excess is approximately three seconds,
mostly in the two build/readiness checks; measured-5 off/4 also includes about
two seconds of extra fixture acquisition. No sole-cause claim is warranted.

A temporary numeric phase probe of the existing cache helper was invoked only
through `make testservices-build`, with run IDs
`instrumentation-f022-cache-attribution-01` and `-02` (both passed, 20,474 and
17,356 ms). The second warm call spent 16,428 ms hashing 2,529 declared inputs,
then 120 ms hashing outputs. Original and instrumented source plus commands and
logs are preserved at primary
`.cartulary/test-results/instrumentation-f022-cache-attribution-01/`.
The probe was removed before the batching repair. This is component attribution,
not qualification evidence. The attempted discovery
`CARTULARY_TEST_RUN_ID=instrumentation-f022-cache-guide-01 make explain-target TARGET=harness-smoke-cache-artifact DETAIL=summary`
failed with usage exit 2 because that name is an internal check, not a declared
public target. Its existing public extended smoke route is used for validation.

The in-progress repair batches content hashing without skipping declared inputs
or outputs, changing preparation policy, or introducing another runtime dependency.
Every batch is bounded and its complete exit/output is checked before use.
Independent digest fixtures cover multiple batches, missing inputs, unusual
filenames, output changes and hashing failure. S-020 remains IN_PROGRESS;
subsequent qualification remains blocked on validation and remaining attribution.


S-020 cache/component checkpoint (still IN_PROGRESS):
`make harness-contract` with run ID `instrumentation-f022-cache-contract-01`
passed 201 tests, canonical duration 106,796 ms, including the independently
computed multi-batch cache-key oracle and fail-closed hashing fixture. The same
cache fixture is now included in the ordinary harness boundary contract route;
it remains in extended smoke as well. The earlier extended smoke attempts
`instrumentation-f022-cache-extended-01` and `-02` failed before reaching it.
The first exposed native helper installation outside the staged bootstrap step;
those helpers now belong to `bootstrap-tool-installations`. The second exposed a
superseded literal recipe assertion after the admitted Make-prerequisite macro
cutover, plus a toolchain-pin fixture diagnostic failure. The bootstrap assertion
now includes existing Go-analysis readiness and the recipe assertion recognizes
the admitted macro. The remaining extended-smoke failure is not claimed fixed.
`instrumentation-f022-cache-generate-01 make generate` passed; all logs are in
the cache-attribution root above. Preserve both failed extended attempts.

Component probe evidence is at primary
`.cartulary/test-results/instrumentation-f022-component-proof-01/`.
Attempt 01 was rejected before execution because its proposed run root already
contained the probe journal. Attempt 02's workload passed but the temporary probe
shadowed `processProof` in the container bridge; it is invalid for container-cost
attribution. The saved original/probe sources and command/output remain retained.
After fixing only that probe, exact run
`instrumentation-f022-component-probe-run-03` passed the browser HTTPS-trust row
in 25,384 ms. Its audit and performance/resource JSON explanations all passed.
Three producer-side opaque process identities bind the parent, native collector
and container helper to retained samples. Eleven collector samples show 7.5–10.96
MiB RSS; seven container-helper samples show 49.69–51 MiB RSS. Both owned service
containers are available with no omissions or deadline overruns. These sampled
component observations are not a total-observer qualification or peak guarantee.
All probe edits were removed. No executable source reads this tracker.

The next F-022 structural repair consolidates the read-only Docker adapter into
the existing native collector and places shared lease types/labels with the
suite-service lifecycle owner. It removes the extra observer process/socket and
its duplicated protocol, keeps the same exact private proof and client port,
and drains the same bounded requests before pause/stop acknowledgement. Public
formats, all budgets, cadence and preparation policy remain unchanged. The
existing collector remains one run-owned helper. Validate forced death, revocation,
strict proof, cancellation, deadline, caps and real PostgreSQL/object-store scope
before freezing another cohort. Default-off and S-022 gating remain unchanged.


S-020 consolidated collector functional checkpoint — IN_PROGRESS:

- `internal/testutil/suiteservices/leaseproof` now owns the unchanged private lease
  shape and required labels. Acquisition/cleanup and observation share this
  narrow package, without coupling the collector to suite-event or application
  initialization. No retained schema changed.
- The collector invokes the read-only Docker client directly under its existing
  cadence/fence. There is no separate `testservices observe-containers` mode,
  observer process, Unix socket server, or socket wire protocol. Strict private
  proof, exact before/after container identity, single-sample statistics, response
  closure, batch deadline/cap, lifetime limit, counter semantics and omissions
  remain. A null private proof explicitly revokes observation.
- Cancellation joins all in-flight requests before control acknowledgement.
  Collector EOF cancels active work; parent death terminates the collector.
  Direct-child exit notification wakes owned-process cleanup but never replaces
  process identity/group-liveness proof. Tests retain an early false-notification
  case and forced collector death with no late proof publication or restart.
- Observer Go cases move with their implementation into the existing stable
  `harness.evidence_accounting.behavior.native_collection` row. Lifecycle cases
  remain with `harness.browser`; both are exercised.

The first native run `instrumentation-f022-collector-native-01` failed compilation
because the extracted adapter retained an unused `time` import. Its structured
explanation and compiler diagnostics remain under that exact root. Removing that
import is the named repair; it is not an exclusion from qualification.

All passed with exact command journals at primary
`.cartulary/test-results/instrumentation-f022-integrated-collector-validation-01/`:
`make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_collection CARTULARY_HARNESS_CACHE_MODE=off`
(3,849 ms), `make harness-evidence-contract` (21,979 ms),
`make service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.testservices_lifecycle_contract,harness.browser.boundary_support.https_trust HARNESS_DIAGNOSTICS=basic CARTULARY_HARNESS_CACHE_MODE=off`
(25,860 ms), `make lint-scripts` (1,902 ms), and `make lint-shell` (16,295 ms).
Their explicit run IDs use `instrumentation-f022-integrated-{native,evidence,service,scripts,shell}-01`.
Generation and formatting calls/logs are in the preceding component-probe root.
The primary staged tracker hash remains
`9ab829e44fa8df7ba1ee86e794518acd97129cea6af416966fbdda4b33f83426`.

Fresh component probe `instrumentation-f022-consolidated-probe-run-01` passed the
same browser HTTPS-trust selection in 25,658 ms. Producer-side opaque identities
bind eleven collector samples at 11.44–15.36 MiB RSS and eleven parent samples at
341.96 MiB. Both exact owned containers are available, with no omissions or
batch overruns. Its retained audit and resource/performance JSON explanations
passed; the shutdown envelope interval is 25 ms. The temporary identity probe
was removed, with original/instrumented sources preserved at primary
`.cartulary/test-results/instrumentation-f022-consolidated-probe-01/`.
Two subsequent warm `make testservices-build` calls passed in 1,366 and 1,385 ms.
These are component observations on dirty source, not qualifying total CPU/RSS
estimates, elapsed gates, or peak guarantees. RAM remains a wall-time decision.

**F-026: fixture/command-surface projections missed native-helper dependencies.**
This S-020 validation finding belongs to the bootstrap/command-surface owner and
its existing contracts. The extended smoke negative toolchain fixture skips
preparation while copying only source; it therefore failed host admission before
reaching the deliberately invalid pin. Supply its admitted native executable from
the real prepared checkout, with an explicit prerequisite on the owning smoke
tiers. Preserve the asserted pin mismatch and proof that it reaches its checker.
The staged bootstrap and macro assertions above are part of the same reconciliation.
The failed extended roots remain retained. Broader validation is pending and no
S-020 completion, S-021 pass or S-022 start is claimed by this checkpoint.


### S-020 candidate 13 structural repair checkpoint — DONE

Saved before freezing or beginning the next S-021 cohort. F-025 is functionally
repaired and F-026 is closed. The earlier failed extended-smoke runs remain in
the ledger, including `instrumentation-f022-integrated-extended-02`, which still
expected the old bootstrap rendering and exited 2 after 21,418 ms. Updating that
fixture to the existing staged Go-analysis and native-helper prerequisites is the
explicit repair; the following successful run does not erase that failure.

The final review also restored the adopted REQ-846 cooperative boundary: a due
Docker batch cannot start after the 50 ms sweep allowance is exhausted. It records
an omission and advances that cadence attempt without catch-up. A deterministic
zero-budget fixture in the existing native-collection bounds case proves no
Docker inspection or statistics call is made. The separate 250 ms request limit
is unchanged. This closes a concrete F-022 collection-boundary defect rather than
changing any qualification budget.

All commands below passed. Commands, environment/run identity, outer exit and
elapsed time are journaled under primary `.cartulary/test-results/` in the named
validation directory; stdout/stderr remain alongside `commands.json`.

`instrumentation-f022-integrated-collector-validation-03/`:

| Run ID | Exact Make command | Outer elapsed ms |
| --- | --- | --- |
| `instrumentation-f022-integrated-extended-03` | `make run-harness-smoke-extended` | 47607.820 |
| `instrumentation-f022-integrated-command-03` | `make harness-command-surface-contract` | 55039.741 |
| `instrumentation-f022-integrated-contract-03` | `make harness-contract` | 109011.441 |
| `instrumentation-f022-integrated-drift-03` | `make generate-drift` | 19397.151 |
| `instrumentation-f022-integrated-policy-03` | `make generated-artifact-policy-check` | 3093.705 |
| `instrumentation-f022-integrated-shape-03` | `make json-shape-check` | 5132.276 |
| `instrumentation-f022-integrated-scripts-03` | `make lint-scripts` | 1900.677 |
| `instrumentation-f022-integrated-shell-03` | `make lint-shell` | 16103.544 |

`instrumentation-f022-sweep-boundary-validation-01/`:

| Run ID | Exact Make command | Outer elapsed ms |
| --- | --- | --- |
| `instrumentation-f022-sweep-format-01` | `make format` | 10981.863 |
| `instrumentation-f022-sweep-native-01` | `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_collection CARTULARY_HARNESS_CACHE_MODE=off` | 3540.254 |
| `instrumentation-f022-sweep-evidence-01` | `make harness-evidence-contract` | 21324.807 |
| `instrumentation-f022-sweep-audit-01` | `make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/instrumentation-f022-integrated-service-01` | 335.270 |
| `instrumentation-f022-sweep-resources-01` | `make explain-run RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/instrumentation-f022-integrated-service-01 DETAIL=resources` | 353.604 |
| `instrumentation-f022-sweep-resources-json-01` | `make explain-run RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/instrumentation-f022-integrated-service-01 DETAIL=resources JSON=1` | 380.670 |
| `instrumentation-f022-sweep-performance-01` | `make explain-run RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/instrumentation-f022-integrated-service-01 DETAIL=performance` | 368.568 |
| `instrumentation-f022-sweep-performance-json-01` | `make explain-run RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/instrumentation-f022-integrated-service-01 DETAIL=performance JSON=1` | 374.916 |
| `instrumentation-f022-sweep-drift-01` | `make generate-drift` | 19661.081 |

The exact audited service root is
`.cartulary/test-results/instrumentation-f022-integrated-service-01`.
Its resource/performance JSON reports are identical normalized models, and the
human views agree on canonical 23,633 ms, outer receipt 25,820 ms, 162 retained
samples and two separate owned-container scopes. These are functional reports,
not qualifying measurements. Native and evidence tests passed again after the
sweep guard; generation drift passed. Existing integrated full harness, command,
schema/policy and lint gates remain applicable to the otherwise unchanged source.

Compatibility: public retained schemas, command selections, policies, cadence,
lease/cleanup contracts and failure precedence remain unchanged. Removed private
socket/observer mode has no translator; its callers moved together. The cache
helper's changed source invalidates old helper-keyed entries normally. Bootstrap
projections now explicitly prepare their required native executables. No runtime
consumer reads this tracker or NLSpec Markdown.

The shared registered-root validators and bounded publication repairs remain
adopted. The complete observer still needs the unchanged external qualification;
component RSS/build timings do not close F-022 or RB-006. Candidate 13 must use a
clean executable commit, unchanged typed plan, frozen randomized matrix and
RAM-backed retained output for wall time. Every earlier attempt, probe and
failed gate remains preserved. The user reports no other intensive host work.
Primary staged tracker content is preserved at its recorded entry hash. S-021
is ready; S-022 is NOT STARTED. Default-off, local-only, physical coverage limits
and no Core 05 publication claim remain explicit. If qualification fails, save
its failure checkpoint before reopening the responsible workstream.


### S-021 candidate 13 execution entry — IN_PROGRESS

Begins after the saved S-020 completion checkpoint. The clean isolated checkout
`/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary` now uses
branch `codex/instrumentation-qualification-13`, commit
`755e1d529159aecc2babeda7433c8dd289ca6f70`, parent
`d36682b60ba9088f7f7c4e7edd06452bf16f0093`, tree
`20e40ddb7dfc59761037a35230daa3826578000c`. The alternate-index freeze journal is
`/tmp/cartulary-qualification-freeze-13.json`, copied to the primary sweep-boundary
validation root. The primary index bytes were unchanged by the freeze and the
pre-existing staged tracker retains its recorded hash. No stashing or absorption
of unrelated staged work occurred.

Native session `6362` owns this exact command from the isolated root:

```sh
CARTULARY_TEST_RUN_ID=instrumentation-s021-qualification-13 \
CARTULARY_TEST_RESULTS_DIR=/dev/shm/cartulary-instrumentation-qualification-13 \
make harness-instrumentation-qualify QUALIFICATION_PLAN=tools/harness_instrumentation_qualification_plan.json
```

The archive supervisor records argv, environment, commit, mount, capacity, PID,
start/end, outer exit and elapsed time at isolated
`.cartulary/test-results/instrumentation-s021-qualification-13/archive-command.json`.
It retains stdout/stderr there and copies all RAM evidence after Make exits,
checking every file digest. Both original and archive remain. Source digest and
plan identity will be taken from the emitted typed manifest; no executable
source is changed during the cohort. Preparation, all six treatments, cold and
warm-up observations, five measured blocks and any prescribed sixth block remain
unchanged. No source edits, profiling or unrelated verification runs in parallel.
S-021 cannot pass from partial cells; S-022 remains gated. RAM is for wall time.

Candidate 13 emitted source digest `sha256:cf58fd60a73a9df79b8c107194edf1ab873cb49f243c53496be96e58a5606cb8` and unchanged plan digest `sha256:1152b853dc9f2c8ceeb20343245e305806b729ac413f8cf0a686118627ba38af`. The typed manifest identifies WSL2/Linux subreaper wait4/proc accounting, seed 20261009, policy `sha256:c3f10c497b2c3eed5e0b647a02316d375b139fc5621d8b152ef93541bb6f4930`, toolchain `sha256:4b62e2558e6dcbe1d3c632fb9c0bb47e5a89e5dda08f103a72f61347b5ef1810` and system `sha256:546793d134ea4063515a25068aa22b78078fc1613c1ee56fb5f9e41c5c550588`. No acceptance is inferred from initial publication.


### S-021 candidate 13 rejection checkpoint — IN_PROGRESS

The exact frozen command above exited Make 2 after 3,448,379.421 ms. It completed
234 attempts: 48 each for Go, Vitest, shell and service, and 42 for browser.
Every attempt is eligible; there are no command or measurement-loss exclusions.
Every cold treatment and discarded warm-up passed. Browser required only the
five standard measured blocks and is stable. The other four workloads took the
prescribed sixth block and remain unstable because removing a block changes an
individual gate. None fails the actual elapsed-window MAD rule.

The complete archive contains 7,742 digest-verified files at isolated
`.cartulary/test-results/instrumentation-s021-qualification-13/`, with original
RAM evidence still at
`/dev/shm/cartulary-instrumentation-qualification-13/instrumentation-s021-qualification-13`.
`archive-command.json` and `archive-digests.json` bind the exact command, source,
environment, storage, start/end, exit and every archived file. The isolated
checkout remains clean at `755e1d529159aecc2babeda7433c8dd289ca6f70` with source
`sha256:cf58fd60a73a9df79b8c107194edf1ab873cb49f243c53496be96e58a5606cb8`.

Independent public validation:

```sh
CARTULARY_TEST_RUN_ID=instrumentation-s021-validation-13 \
make harness-performance-check EVIDENCE_ROOTS_FILE=/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary/.cartulary/test-results/instrumentation-s021-qualification-13/qualification/qualification.json
```

It exited Make 2 / inner gate 13 after 908.702 ms. Command, stdout and stderr are
retained at isolated `.cartulary/test-results/instrumentation-s021-validation-13/`.
Independent Python arithmetic uses standard median/MAD and nearest rank, direct
paired equations and endpoint integration for interval unions; it does not import
production analysis. `independent-arithmetic.json` preserves every signed pair,
all error estimates, actual-window tests and leave-one-block-out gate changes.
All 30 eligible/numeric dispositions and reported statistics agree with the
public validator. Its script and table remain with the evidence.

| Workload | Treatment | Blocks | CPU mean % | RSS upper MiB | Canonical median / p90 % | Outer median / p90 % | Final ms | Failed numeric gates |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| go | off/0 | 6 | 0.0000 | 0.0000 | 0.0000 / 0.0000 | 0.0000 / 0.0000 | 0.000 | none |
| go | off/1 | 6 | -1.2593 | 4.2239 | -0.8802 / 1.0601 | 0.7617 / 2.2344 | 0.000 | none |
| go | off/4 | 6 | -0.8876 | 13.5450 | -0.8368 / 0.3512 | -1.1199 / 2.3511 | 0.000 | none |
| go | basic/0 | 6 | -0.3271 | 14.5805 | -0.8331 / 2.2085 | 0.5405 / 1.6954 | 55.006 | none |
| go | basic/1 | 6 | -0.4271 | 24.7966 | -0.3069 / 1.5018 | -0.3237 / 3.0087 | 60.299 | outer_p90 |
| go | basic/4 | 6 | 0.1641 | 12.6781 | -0.5279 / 1.5901 | 0.1327 / 2.6213 | 58.615 | none |
| vitest | off/0 | 6 | 0.0000 | 0.0000 | 0.0000 / 0.0000 | 0.0000 / 0.0000 | 0.000 | none |
| vitest | off/1 | 6 | -0.1038 | 1.4006 | -0.4974 / 1.4502 | -0.2828 / 1.2742 | 0.000 | none |
| vitest | off/4 | 6 | -0.0188 | 5.6254 | 0.4222 / 3.5597 | 0.3628 / 2.6495 | 0.000 | canonical_p90 |
| vitest | basic/0 | 6 | 0.2848 | 12.8397 | 0.4800 / 3.2630 | 1.0512 / 2.6788 | 61.029 | canonical_p90 |
| vitest | basic/1 | 6 | 0.2577 | 12.9088 | -0.0642 / 2.7357 | 0.8519 / 2.2505 | 58.734 | none |
| vitest | basic/4 | 6 | 1.7188 | 23.2391 | 0.6299 / 2.0105 | 1.1061 / 3.1252 | 58.911 | outer_p90 |
| shell | off/0 | 6 | 0.0000 | 0.0000 | 0.0000 / 0.0000 | 0.0000 / 0.0000 | 0.000 | none |
| shell | off/1 | 6 | 0.2792 | 10.0922 | 0.0831 / 0.2799 | 0.0371 / 0.2353 | 0.000 | none |
| shell | off/4 | 6 | 0.2036 | 20.3924 | -0.0896 / 0.7585 | -0.1210 / 0.9038 | 0.000 | none |
| shell | basic/0 | 6 | 0.2160 | 28.7533 | -0.1476 / 0.2274 | -0.1510 / 0.4815 | 57.102 | none |
| shell | basic/1 | 6 | 0.2999 | 43.1422 | 0.1615 / 0.6679 | 0.0778 / 1.0238 | 56.155 | rss_upper |
| shell | basic/4 | 6 | 0.4583 | 31.3757 | -0.0664 / 0.7490 | -0.0055 / 0.7315 | 61.215 | none |
| service | off/0 | 6 | 0.0000 | 0.0000 | 0.0000 / 0.0000 | 0.0000 / 0.0000 | 0.000 | none |
| service | off/1 | 6 | 0.0863 | 1.1553 | 0.2119 / 1.0779 | 0.2305 / 1.3946 | 0.000 | none |
| service | off/4 | 6 | 0.1586 | 5.5939 | 0.1263 / 4.4885 | 0.1271 / 3.4884 | 0.000 | canonical_p90, outer_p90 |
| service | basic/0 | 6 | 0.7085 | 25.5963 | 0.0737 / 1.1185 | -0.0783 / 1.2672 | 55.907 | none |
| service | basic/1 | 6 | 0.5492 | 19.8731 | -0.4346 / 0.7530 | -0.1892 / 0.6322 | 60.970 | none |
| service | basic/4 | 6 | 0.7589 | 17.5678 | 0.4641 / 0.5980 | 0.2615 / 0.8616 | 59.534 | none |
| browser | off/0 | 5 | 0.0000 | 0.0000 | 0.0000 / 0.0000 | 0.0000 / 0.0000 | 0.000 | none |
| browser | off/1 | 5 | 0.2354 | 4.1419 | 0.2354 / 1.0401 | 0.2242 / 0.9407 | 0.000 | none |
| browser | off/4 | 5 | -0.0088 | 6.0643 | 0.3653 / 0.8762 | 0.3836 / 0.6056 | 0.000 | none |
| browser | basic/0 | 5 | 0.5553 | 19.3953 | -0.0817 / 0.0838 | -0.1754 / 0.4395 | 63.541 | none |
| browser | basic/1 | 5 | 0.7003 | 22.9662 | 0.1281 / 0.7349 | 0.4484 / 1.0404 | 63.185 | none |
| browser | basic/4 | 5 | 0.5587 | 22.0562 | -0.1667 / 0.9672 | -0.0618 / 1.1539 | 70.824 | none |

All CPU means are strictly below 2%; the maximum is Vitest basic/4 at 1.7187943%.
Signed negative observations are retained and do not establish an improvement.
All elapsed medians pass. Six of 25 non-reference cells fail numeric gates:
Go basic/1 outer p90; Vitest off/4 and basic/0 canonical p90; Vitest basic/4 outer
p90; service off/4 both p90s; shell basic/1 RSS. Shell basic/1's conservative RSS
upper is 43.1422 MiB, above 32 MiB. All finalization unions pass; the maximum is
70.824245 ms. Browser's five-block matrix is stable and passes every gate.

This is a rejection, not qualification completion. It reduces the observed
failure set relative to candidate 12 but does not permit selecting favorable
cells across revisions or relabeling earlier evidence. The full cohort, all
prior failures, and all attribution attempts remain retained. CPU, elapsed,
RSS, finalization, sample, reader, preparation and stability rules are unchanged.
F-022 / C-013 remains assigned to S-020 for evidence-backed attribution of the
remaining cost/variation; RB-006 and S-021 stay open. The user reports no other
intensive host work; no host-contention explanation is asserted. Reopening
S-020 requires this saved rejection checkpoint first. S-022 remains NOT STARTED,
gated on a complete passing current-source matrix. Default-off, local-only and
no Core 05 publication claim remain explicit.


### S-020 post-candidate-13 attribution entry — IN_PROGRESS

The saved S-021 rejection precedes this entry. Focus F-022/C-013 on the six
remaining failures, using exact retained paired runs before changing source.
Selection arithmetic identifies Go basic/1 block 2 (outer-only increase), Vitest
off/4 and basic/0 block 6 (canonical p90), Vitest basic/4 block 5 (outer p90),
and service off/4 block 3 (both elapsed p90s). Nine distinct exact retained roots
are recorded in candidate 13's validation directory as
`attribution-selections.json`. Public performance explanations and canonical
unit evidence will distinguish preparation, command execution and surrounding
cost. These are terminal inspections, not another qualification attempt.

Shell basic/1 RSS paired differences are 9.3502, 12.4782, 27.7679, 25.8481,
5.8016 and 24.1017 MiB. Each individual difference is below 32 MiB, but the
required median + three MADs + error estimate is 43.1422 MiB and fails. This
explains the arithmetic, not its physical cause; do not relabel it as a pass or
attribute it to host contention. Investigate retained component scope and the
measurement/producer allocation boundaries before selecting a structural repair.
No budgets, workloads, observation counts, cadence or exclusions change. No
unchanged cohort is rerun for a favorable outcome. S-022 remains gated.


S-020 retained attribution checkpoint: nine exact public performance explanations
passed; their command journal and JSON are under candidate 13 validation
`attribution/`. Go's selected pair has canonical 1,137 versus 1,139 ms, so its
outer miss is outside canonical execution. Vitest block 6 has runner-unit
2,768 ms (off/4) and 2,764 ms (basic/0), versus 2,687 ms reference; frontend
preparation also differs. Service off/4 block 3 has runner-unit 10,051 versus
9,551 ms reference, while fixture-acquisition union differs by 25 ms. These
identify time regions, not a proven physical cause or an exclusion rationale.

Validated retained basic shell samples show the graph parent near 340–342.5 MiB,
another long-lived process near 282–286 MiB, and a descendant whose sparse samples
range roughly 124–554 MiB during the unit. The collector-sized process is near
11–15 MiB. `component-samples.json` preserves the exact opaque scopes and samples;
only the graph parent has a code-established first-registration role here. Do
not infer command names for other scopes or promote sparse component samples
to external total accounting. This points to allocation/retention boundaries for
further investigation but does not prove the RSS dispersion's cause.

**F-027: strict JSON construction loses prototype-named object members.**
The shared strict parser assigns parsed keys to a normal object. JavaScript's
`__proto__` setter can therefore consume a JSON member or alter the result's
prototype before schema validation. This conflicts with C-014's strict retained
validation boundary and closed-shape rejection. The clean repair keeps lexical
duplicate/I-JSON validation while giving native `JSON.parse` sole ownership of
value/object construction. It must preserve decoded duplicate detection, Unicode,
finite/safe numeric rules, ordinary object prototypes and every own JSON member.
No schema or compatibility translator is needed; malformed published shapes
that slipped through are intentionally rejected. This is an implementation,
security-regression and documentation correction, not a budget change.

`instrumentation-f027-characterization-01` currently runs the public evidence
contract with independently authored prototype/escaped-duplicate fixtures and
a real retained resource-index closed-shape regression. Preserve its failure
before fixing the parser. Measure representative parser cost before choosing
whether further allocation changes belong in the same structural repair. No
qualification is active; candidate 13 and its archive remain unchanged.


F-027 characterization and component checkpoint — IN_PROGRESS:

- First `make harness-evidence-contract`, run
  `instrumentation-f027-characterization-01`, failed Make 2 in 20,247.734 ms.
  It proves the lost/prototype-mutating JSON member directly. Its initial
  resource-index fixture had an independently invalid empty stream, so that
  failure is not evidence of retained acceptance. This limitation is preserved.
- Correct the resource fixture to one valid sample and first assert its
  unmodified projection succeeds. Add independently authored ordinary, escaped,
  Unicode, duplicate, malformed and non-interoperable numeric cases. Repeating
  the characterization with the original parser as
  `instrumentation-f027-characterization-02` failed exactly the two intended
  regressions (104/106 pass) in 20,908.998 ms. The retained-index regression now
  reports `Missing expected exception`, proving the original closed-schema bypass.
  Restore the repaired parser immediately afterward. Both failure roots remain.
- Native `JSON.parse` now solely constructs objects/arrays after the unchanged
  duplicate-member and I-JSON lexical checks. This removes manual property writes
  and the scanner's intermediate object graph. REQ-849/AC-146 clarify preservation
  of own JSON members and rejection of undeclared prototype-like names. No
  generated schema acceptance set or retained format changes.
- `instrumentation-f027-evidence-01` passed in 20,251.186 ms. The corrected fixtures
  and temporary attribution test passed in `instrumentation-f027-profile-01`
  (22,198.998 ms). Under that Make-owned test, 139 typed inputs were parsed in
  fresh child processes, with alternating variant order, one discarded block and
  eight measured pairs. Equal input counts and serialized output byte counts are
  asserted; full output digests were not retained, so this is not a semantic
  parity proof. The independent regression corpus supplies that validation. Native construction
  changes median CPU by −5,344.5 µs (MAD 1,172), RSS by −6,103,040 bytes (MAD
  1,189,888), process high-water RSS by −5,660,672 bytes (MAD
  1,589,248), and wall by +0.243 ms (MAD 0.792). Heap-used rises by 2,068,632 bytes
  (MAD 31,804); do not confuse this with RSS. These component observations do not
  prove a total observer budget or peak guarantee.
- Raw observations, arithmetic, both parser sources, input paths, exact command
  journals and original/instrumented fixture source are at primary
  `.cartulary/test-results/instrumentation-f027-strict-json-01/`. Temporary profiler
  code was removed. The complete candidate-13 source/evidence remains unchanged.

The repaired source now needs fresh functional and projection gates before a
new S-020 completion checkpoint. F-022/RB-006 remain open. The additional
regressions validate the security boundary; they do not extend the workload to
obtain a budget pass. No cadence, threshold, treatment, cache or exclusion changes.


F-027 functional validation checkpoint: `instrumentation-f027-evidence-02`
passed in 21,016.037 ms; the basic/cold exact current-epoch row
`instrumentation-f027-cold-02` passed in 23,619.846 ms; command-surface contract
`instrumentation-f027-command-02` passed in 53,748.220 ms. Full
`instrumentation-f027-contract-02` failed Make 2 after 107,265.031 ms (1/2 graph
units passed). Public `make explain-run RESULTS_DIR=.cartulary/test-results/instrumentation-f027-contract-02`
reported `child_target_failure` with all four cleanup steps complete.
The referenced target stdout identifies the sole failing boundary case: the JSON
shape checker rejects stale generated topology input provenance because
`tools/harness/contract/strict-json.mjs` changed. This is related generation drift,
not a parser functional rejection. The pipeline stopped there; its later drift,
policy, shape and script checks have not run. Regenerate through `make generate`,
then rerun the failed full contract and remaining gates. Preserve this failed root
and the exact command journal at primary
`.cartulary/test-results/instrumentation-f027-validation-01/`.


### S-020 candidate 14 strict-JSON repair checkpoint — DONE

Saved before freezing or beginning another S-021 cohort. F-027 is closed:
standard JSON construction preserves own data members and ordinary prototypes;
lexical duplicate and I-JSON checks remain. The corrected original-parser
characterization fails both intended security cases, while repaired evidence,
cold-row and full contract checks pass. REQ-849/AC-146 state the own-member
boundary explicitly. No schema version, public selection, retained format,
canonical timing, observer cadence, qualification plan or budget changed.

The temporary profiler and baseline source substitution are removed. The native
parser and permanent regression file match their saved repaired snapshots. The
staged tracker retains its entry hash. Candidate 13 is still clean, immutable
and rejected; component cost reductions do not qualify this successor.

The earlier validation-01 narrow results and related full-contract drift failure
remain above. Following `make generate`, all commands in primary
`.cartulary/test-results/instrumentation-f027-validation-02/commands.json` passed:

| Run ID | Exact Make command | Outer elapsed ms |
| --- | --- | --- |
| `instrumentation-f027-generate-03` | `make generate` | 11888.708 |
| `instrumentation-f027-contract-03` | `make harness-contract` | 106549.726 |
| `instrumentation-f027-drift-03` | `make generate-drift` | 19899.606 |
| `instrumentation-f027-policy-03` | `make generated-artifact-policy-check` | 3056.953 |
| `instrumentation-f027-shape-03` | `make json-shape-check` | 5107.546 |
| `instrumentation-f027-scripts-03` | `make lint-scripts` | 1869.493 |
| `instrumentation-f027-audit-03` | `make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/instrumentation-f027-cold-02` | 333.290 |
| `instrumentation-f027-resources-03` | `make explain-run RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/instrumentation-f027-cold-02 DETAIL=resources` | 342.570 |
| `instrumentation-f027-resources-json-03` | `make explain-run RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/instrumentation-f027-cold-02 DETAIL=resources JSON=1` | 340.701 |
| `instrumentation-f027-performance-03` | `make explain-run RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/instrumentation-f027-cold-02 DETAIL=performance` | 347.872 |
| `instrumentation-f027-performance-json-03` | `make explain-run RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/instrumentation-f027-cold-02 DETAIL=performance JSON=1` | 335.467 |

Fresh resource/performance JSON reports for the basic/cold current-epoch run are
identical normalized models. Their human views and the exact-run audit pass.
Generation updates the topology input provenance for the shared parser through
its owner; generated policy, drift and shape gates pass. Shared-parser fixtures
cover ordinary values, nested objects/arrays, escaped duplicate keys, Unicode
pairs and lone surrogates, malformed syntax, negative zero, non-finite and unsafe
integer values. The full contract also covers generated-validator parity and
bootstrap operation.

Migration: previously hidden undeclared prototype-named fields now cause the
intended malformed-evidence rejection, without rewriting retained test outcomes
or translating historical formats. Authored valid JSON retains native semantics.
The source change affects the shared input/retained parsing boundary, so rerun the
full frozen matrix, including the previously passing browser cells. Preserve
all earlier failures and do not substitute historical cells. Rollback, if needed,
must restore parser, regressions, owner clarification and generated provenance
coherently; it would reopen the security finding and is not a qualification pass.
F-022, RB-006 and S-021 remain open. S-022 is NOT STARTED. Default-off, local-only,
physical coverage limitations and no Core 05 publication claim remain explicit.


### S-021 candidate 14 execution entry — IN_PROGRESS

Begins after the saved S-020 strict-JSON completion checkpoint. The clean isolated
checkout `/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary`
uses branch `codex/instrumentation-qualification-14`, commit
`59ea4783a8934de6f15e6b4d77a76f4bd8f2f5cf`, parent
`755e1d529159aecc2babeda7433c8dd289ca6f70`, tree
`40dd1d7d1aaee9579bf3035f46e65fd4f83cf959`. Exactly four paths differ from candidate
13: the parser, its regression file, NLSpec clarification and generated topology
input provenance. Freeze journal `/tmp/cartulary-qualification-freeze-14.json`
is copied to the primary F-027 validation-02 root. The alternate index preserves
the primary index bytes and pre-existing staged tracker hash.

Native session `2547` owns this command from the clean isolated root:

```sh
CARTULARY_TEST_RUN_ID=instrumentation-s021-qualification-14 \
CARTULARY_TEST_RESULTS_DIR=/dev/shm/cartulary-instrumentation-qualification-14 \
make harness-instrumentation-qualify QUALIFICATION_PLAN=tools/harness_instrumentation_qualification_plan.json
```

Durable command/output/archive evidence is assigned to isolated
`.cartulary/test-results/instrumentation-s021-qualification-14/`; the supervisor
copies and digest-verifies all RAM evidence after Make exits, retaining both.
No executable changes, profiling or other verification runs during the matrix.
All five workloads and six treatments rerun with the unchanged plan, reader
cadence, CPU mean <2%, other budgets and stability rules. Candidate 13's failures
remain independent retained evidence. S-022 remains gated; no qualification
claim is inferred from launch. RAM continues to serve wall time only.

Candidate 14 emitted source digest `sha256:4c5221d4c2377beb77804126537924c3385cfffbfeffd9aa85e699d097844516`, unchanged plan digest `sha256:1152b853dc9f2c8ceeb20343245e305806b729ac413f8cf0a686118627ba38af`, and unchanged policy digest `sha256:c3f10c497b2c3eed5e0b647a02316d375b139fc5621d8b152ef93541bb6f4930`. Its typed manifest supplies the frozen toolchain/system identity; initial publication is not acceptance.


### S-021 candidate 14 rejection checkpoint — NOT QUALIFIED

The frozen command exited Make 2 after 3,344,794.524 ms. All 228 attempts are
eligible: 48 each for Go, Vitest and shell, and 42 each for service and browser.
All cold treatments and discarded warm-ups passed. There were no command,
measurement-loss or coverage failures. Go, Vitest and shell required the
prescribed sixth block and remain unstable because removing a block changes an
individual gate. No workload fails the actual elapsed-window MAD rule. Service
is stable and passes every gate. Browser is stable but its basic/1 RSS gate fails;
stability does not imply a budget pass.

The isolated checkout remains clean at
`59ea4783a8934de6f15e6b4d77a76f4bd8f2f5cf`, with source
`sha256:4c5221d4c2377beb77804126537924c3385cfffbfeffd9aa85e699d097844516`.
Plan and policy digests remain those frozen above. Complete evidence is retained
at isolated `.cartulary/test-results/instrumentation-s021-qualification-14/`:
7,559 files were copied and SHA-256 verified. `archive-command.json` and
`archive-digests.json` preserve command, source, environment, tmpfs placement,
start/end, exit and file identities. Original evidence remains at
`/dev/shm/cartulary-instrumentation-qualification-14/instrumentation-s021-qualification-14`.
RAM storage concerns wall time only; it does not change CPU accounting or budgets.
The user reports no other intensive host work; no contention cause is asserted.

Independent public validation:

```sh
CARTULARY_TEST_RUN_ID=instrumentation-s021-validation-14 \
make harness-performance-check EVIDENCE_ROOTS_FILE=/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary/.cartulary/test-results/instrumentation-s021-qualification-14/qualification/qualification.json
```

The validator exited Make 2 / inner gate 13 after 923.674 ms. Its command, stdout,
stderr, independent arithmetic script, signed pairs and complete table are retained
at isolated `.cartulary/test-results/instrumentation-s021-validation-14/`.
Independent standard-library arithmetic uses the frozen equations, nearest-rank
p90, actual-window MAD, leave-one-block-out gates and endpoint integration of
finalization intervals. It imports no production analysis. All 30 public cell
statistics and eligible/numeric dispositions match.

| Workload | Treatment | Blocks | CPU mean % | RSS upper MiB | Canonical median / p90 % | Outer median / p90 % | Final ms | Failed numeric gates |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| go | off/0 | 6 | 0.0000 | 0.0000 | 0.0000 / 0.0000 | 0.0000 / 0.0000 | 0.000 | none |
| go | off/1 | 6 | -0.0986 | 7.4551 | 0.8348 / 3.5682 | -0.0053 / 2.4091 | 0.000 | canonical_p90 |
| go | off/4 | 6 | -0.0754 | 24.7397 | 0.1331 / 1.1691 | -0.4043 / 2.9775 | 0.000 | none |
| go | basic/0 | 6 | -0.3656 | 16.2589 | 0.6679 / 2.2482 | 0.0656 / 1.3513 | 59.019 | none |
| go | basic/1 | 6 | 0.9506 | 27.3061 | 1.4290 / 2.0517 | 2.0821 / 3.1190 | 62.052 | outer_median, outer_p90 |
| go | basic/4 | 6 | 0.4699 | 17.7253 | 1.2018 / 2.4086 | 1.7484 / 2.7500 | 55.438 | none |
| vitest | off/0 | 6 | 0.0000 | 0.0000 | 0.0000 / 0.0000 | 0.0000 / 0.0000 | 0.000 | none |
| vitest | off/1 | 6 | -0.1608 | 7.8383 | -1.0096 / -0.3540 | -0.4043 / 0.6807 | 0.000 | none |
| vitest | off/4 | 6 | 1.2842 | 16.1329 | -0.2730 / 2.4390 | 0.4832 / 1.3706 | 0.000 | none |
| vitest | basic/0 | 6 | 0.8964 | 34.8585 | -1.2019 / 3.5795 | 0.3593 / 2.6970 | 53.141 | canonical_p90, rss_upper |
| vitest | basic/1 | 6 | 0.6291 | 16.6843 | -0.8840 / 0.1925 | 0.6813 / 1.1097 | 59.032 | none |
| vitest | basic/4 | 6 | 1.2691 | 19.9728 | -0.1284 / 1.9249 | 0.5118 / 1.3499 | 59.077 | none |
| shell | off/0 | 6 | 0.0000 | 0.0000 | 0.0000 / 0.0000 | 0.0000 / 0.0000 | 0.000 | none |
| shell | off/1 | 6 | 0.1880 | 10.3272 | 0.1603 / 0.8857 | 0.1638 / 0.8671 | 0.000 | none |
| shell | off/4 | 6 | 0.0154 | 11.9933 | -0.1832 / 1.6141 | -0.2155 / 1.3442 | 0.000 | none |
| shell | basic/0 | 6 | 0.4828 | 33.2776 | 0.3675 / 0.4788 | 0.3627 / 0.5970 | 62.870 | rss_upper |
| shell | basic/1 | 6 | 0.5227 | 24.8691 | 0.1696 / 0.4823 | 0.2260 / 0.5503 | 54.252 | none |
| shell | basic/4 | 6 | 0.8290 | 49.7457 | 0.0814 / 1.4503 | 0.1562 / 1.3667 | 61.283 | rss_upper |
| service | off/0 | 5 | 0.0000 | 0.0000 | 0.0000 / 0.0000 | 0.0000 / 0.0000 | 0.000 | none |
| service | off/1 | 5 | -0.0859 | 3.3242 | -0.4302 / 0.5842 | -0.4370 / 0.5654 | 0.000 | none |
| service | off/4 | 5 | 0.0836 | 1.5177 | 0.7224 / 1.0501 | 0.6240 / 1.2490 | 0.000 | none |
| service | basic/0 | 5 | 0.1852 | 21.1324 | 0.0369 / 0.7419 | 0.0553 / 1.1242 | 54.703 | none |
| service | basic/1 | 5 | 0.2484 | 24.2612 | -1.0186 / 0.2514 | -0.6903 / 0.3814 | 60.763 | none |
| service | basic/4 | 5 | 0.6018 | 22.9948 | -0.5104 / 1.6935 | 0.1969 / 1.3985 | 58.931 | none |
| browser | off/0 | 5 | 0.0000 | 0.0000 | 0.0000 / 0.0000 | 0.0000 / 0.0000 | 0.000 | none |
| browser | off/1 | 5 | -0.2222 | -1.1744 | -0.4149 / 0.1137 | -0.7104 / 0.3781 | 0.000 | none |
| browser | off/4 | 5 | 0.6510 | 6.6984 | 0.1616 / 1.6678 | 0.1984 / 1.3561 | 0.000 | none |
| browser | basic/0 | 5 | 0.6717 | 31.7719 | 0.9303 / 0.9722 | 0.6563 / 0.9509 | 65.816 | none |
| browser | basic/1 | 5 | 0.4315 | 39.3681 | -0.5448 / 0.0854 | -0.7335 / 0.1954 | 65.640 | rss_upper |
| browser | basic/4 | 5 | 0.6243 | 26.5935 | 0.1152 / 0.2892 | 0.0845 / 0.4257 | 71.794 | none |

All CPU arithmetic means are strictly below 2%; the maximum is 1.2841773% of one
core (Vitest off/4). Negative signed observations remain retained and do not
establish an improvement. All finalization gates pass; the maximum union upper
bound is 71.794181 ms. Six of 25 non-reference cells fail numeric gates, comprising
two canonical p90, one outer median, one outer p90 and four RSS failures. Their
limits remain 2% elapsed median, 3% elapsed p90 and 32 MiB conservative RSS upper.
No cell, block or earlier revision is selected to create a favorable combined claim.

This saves S-021's failed disposition before reopening S-020. F-022/C-013 and
RB-006 remain open. The strict-JSON security repair remains functionally complete,
but does not establish observer qualification. S-022 remains NOT STARTED; no
finalizer, final handoff completion or Core 05 publication is claimed. Default-off,
local-only and the physical measurement exclusions remain explicit. The primary
staged tracker still has SHA-256
`9ab829e44fa8df7ba1ee86e794518acd97129cea6af416966fbdda4b33f83426`.


### S-020 post-candidate-14 catalog boundary entry — IN_PROGRESS

Begins after the saved complete S-021 rejection above. Code review finds that
`row-runner-cli.mjs` calls the full `loadTestCatalog` source-discovery path even
though public graph preflight has already validated the full catalog. This loads
unrelated Go and TypeScript selector sources into every private Go, Vitest and
shell row-runner process. Candidate-13 component observations bound substantial
Node RSS, and candidate 14 still has four RSS gate failures; this motivates
attribution, but does not prove that catalog discovery causes those failures.

Separate complete typed catalog metadata validation from source-selector
resolution. Preserve full catalog source validation, overlap checks and summary
publication in public preflight. Give private row runners a narrow operation that
validates current metadata and all requested row identities, resolves the selected
source selectors and returns only those rows. It must not emit a complete-catalog
validation report, add a general validation bypass, weaken parent preflight or
accept a partially valid selection. Preserve source/path checks, exact selection,
runner adapters, results, failure categories and canonical timing boundaries.

Owner C-013 and TH-HARNESS-REQ-016/074 and AC-063 selector requirements remain controlling;
reconcile the exact owner clause before editing implementation. Validate full-model
and selected-row parity, malformed global metadata, unknown/duplicate selection,
selected source changes and unsafe paths, and continued rejection of invalid
unselected selectors by full preflight. Route new regression work through its
command/preflight owner, without padding the qualification workload. Retain all
component attempts and output equivalence evidence. Adopt a change only with
functional and component evidence; remove temporary attribution code before
validation and freezing. No unchanged qualification rerun is authorized by this
checkpoint. CPU mean stays strictly below 2% per cell; all other gates, workload
selections and schedules remain unchanged. S-022 remains gated and unstarted.


S-020 catalog validation checkpoint — IN_PROGRESS: `make task-guide
ROLE=module-author OWNER=harness.test_catalog` passed and selected `make test-slice
OWNER=harness.test_catalog` with broader `make harness-contract-tests`. The first
`make generate` attempt reused the already occupied run ID
`instrumentation-f022-catalog-generate-01` and was rejected before execution;
no existing evidence was overwritten. Fresh
`instrumentation-f022-row-selection-generate-01` passed in 11,415.982 ms.
`make harness-command-surface-contract`, run
`instrumentation-f022-row-selection-command-01`, failed Make 2 after 54,032.567 ms.
The new private-row fixture omitted the approved `cmd` directory, so path
validation correctly failed with ENOENT before its intended assertion. This is
related fixture construction failure, not evidence that the new boundary passed.
The fixture now creates every declared approved root without unrelated sources.
All commands and output remain at primary
`.cartulary/test-results/instrumentation-f022-catalog-selection-01/`; the failing
nested target summary/logs remain under the exact command run root. The scheduled
script lint did not run after failure. No S-020 completion or S-021 rerun is claimed.


S-020 selected-row boundary component checkpoint — IN_PROGRESS:

- TH-HARNESS-REQ-016 now explicitly assigns full source closure to public
  preflight and limits private row children to current metadata plus selected
  source resolution, without a complete-catalog success report or a caller bypass.
- `readCatalogDefinition` owns shared metadata validation; `loadTestCatalog`
  still resolves every source and publishes the unchanged complete summary.
  `loadSelectedTestRows` checks all requested identities before source work,
  resolves only selected sources, and returns only those rows. The private
  Go/Vitest/shell row CLI uses it; public graph preflight remains unchanged.
  No public format, schema, selection, adapter or failure category changes.
- `private_row_selection_boundary` runs on the command/preflight contract route,
  outside the selected shell qualification workload. It verifies real-row parity,
  full-preflight rejection of missing unselected sources, invalid/unknown/duplicate
  requests, changed/missing/ambiguous selected symbols, symlinked packages,
  overlapping selected rows and malformed unselected metadata.
- Corrected `make harness-command-surface-contract` passed as
  `instrumentation-f022-row-selection-command-02` in 53,491.591 ms;
  `make lint-scripts` passed as `instrumentation-f022-row-selection-scripts-01`
  in 1,976.929 ms. The earlier fixture and occupied-run-ID failures remain retained.
- A temporary Make-owned component test ran through
  `make harness-command-surface-contract`,
  `instrumentation-f022-row-selection-profile-01`, PASS in 93,571.540 ms.
  It retains 69 fresh-process observations: 15 parity observations for all five
  frozen workload selections, six discarded warm-up observations and 48 measured
  observations (eight alternating before/after pairs for Go, Vitest and shell).
  Every child exited zero. Before/after snapshots differ only in this refactor;
  the baseline probe rewrites relative import locations to absolute file URLs,
  retaining the original algorithm. Raw source snapshots and linked probe copies
  remain alongside command, output and numeric evidence.
- All selected row SHA-256 values match. Full catalog semantic and summary
  digests match on all five selections. No production arithmetic is imported by
  `component-arithmetic.json`; it retains every signed pair and paired MAD.
  Temporary test instrumentation was removed and byte equality with its original
  test entrypoint was verified.

| Component | Paired median CPU delta µs / MAD | Paired median wall delta ms / MAD | Paired median RSS delta bytes / MAD |
| --- | --- | --- | --- |
| Go row discovery | -1,241,496.5 / 946.5 | -781.521 / 1.198 | -204,627,968 / 1,560,576 |
| Vitest row discovery | -1,118,544 / 8,342 | -679.657 / 7.640 | -174,630,912 / 1,409,024 |
| Shell row discovery | -1,241,949 / 3,898 | -779.936 / 5.578 | -204,634,112 / 1,384,448 |

These are fresh Node import/catalog-load component observations, not whole-command
qualification, time-weighted steady RSS or an unsampled-peak guarantee. Process
high-water and heap-used values are retained separately. This removes redundant
unrelated source discovery from each row child, reducing coupling as the catalog
grows; it does not change observer policy or count a faster workload as a budget
pass. Remaining functional, real-row and projection gates must pass before S-020
completion, a clean freeze and another full S-021 matrix. S-022 stays unstarted.
Evidence: primary `.cartulary/test-results/instrumentation-f022-catalog-selection-01/`.


### S-020 selected-row catalog boundary completion checkpoint — DONE

Saved before freezing or starting another qualification cohort. The owner
clarification, metadata/source separation, private child migration and command
boundary regression are functionally complete. Full public preflight continues
to validate every selector and overlap before setup. No public caller receives
the selected-row operation, no complete-catalog success is inferred from it, and
all requested row identities are resolved before any selected source work.
Current metadata closure, exact adapter selection, fixture ownership, canonical
timing boundaries and failure classification remain intact. There is no public
schema cutover or compatibility translation for this consolidation.

Permanent edits: the Testing Harness NLSpec, test-catalog facade/loader,
`row-runner-cli.mjs`, command-contract catalog fixtures/registration, and the
Make-generated execution topology render index. The profiling entrypoint was
restored byte-for-byte; no temporary probe is in executable source. Full catalog
summary/digest and all five selected row digests match the prior implementation.
The saved 69 component observations support removing repeated global discovery;
they are not qualification evidence. The previous occupied-ID rejection and
fixture-construction failure are retained with their corrections.

Validation ran sequentially in the primary WSL2 checkout through public Make
routes; diagnostics are basic on the real selected rows, with forced-cold cache
mode on the shell row. Exact commands, IDs, stdout, stderr and results are at
primary `.cartulary/test-results/instrumentation-f022-row-selection-validation-01/`:

| Command | Exact run ID | Result / elapsed ms |
| --- | --- | --- |
| `make generate` | `instrumentation-f022-row-selection-generate-02` | PASS / 11,395.829 |
| `make test-slice OWNER=harness.test_catalog ROWS=harness.test_catalog.behavior.routing_contract HARNESS_DIAGNOSTICS=basic` | `instrumentation-f022-row-selection-catalog-01` | PASS / 107,770.142 |
| `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.local_migration_activities HARNESS_DIAGNOSTICS=basic` | `instrumentation-f022-row-selection-go-01` | PASS / 2,491.098 |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.work_pin_presentation_removal_27944912ba HARNESS_DIAGNOSTICS=basic` | `instrumentation-f022-row-selection-vitest-01` | PASS / 4,752.040 |
| `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.current_epoch_evidence HARNESS_DIAGNOSTICS=basic CARTULARY_HARNESS_CACHE_MODE=cold` | `instrumentation-f022-row-selection-shell-01` | PASS / 21,967.400 |
| `make service-backed-test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.schema_bootstrap HARNESS_DIAGNOSTICS=basic` | `instrumentation-f022-row-selection-service-01` | PASS / 15,660.776 |
| `make service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.https_trust HARNESS_DIAGNOSTICS=basic` | `instrumentation-f022-row-selection-browser-01` | PASS / 25,386.953 |
| `make harness-evidence-contract` | `instrumentation-f022-row-selection-evidence-01` | PASS / 19,929.071 |
| `make generate-drift` | `instrumentation-f022-row-selection-drift-01` | PASS / 18,608.368 |
| `make generated-artifact-policy-check` | `instrumentation-f022-row-selection-policy-01` | PASS / 2,288.873 |
| `make json-shape-check` | `instrumentation-f022-row-selection-shape-01` | PASS / 5,021.265 |
| `make lint-scripts` | `instrumentation-f022-row-selection-scripts-02` | PASS / 1,863.071 |

The preceding corrected `make harness-command-surface-contract` and Make-owned
component contract pass remain at
`.cartulary/test-results/instrumentation-f022-catalog-selection-01/`. The catalog
owner slice above executes its routed complete contract suite on the final source.
Fresh `harness-observability-check` and human/JSON `explain-run DETAIL=resources`
and `DETAIL=performance` all pass against exact run
`instrumentation-f022-row-selection-browser-01`; the two JSON normalized models
are identical. Reader commands, output and parity are retained in the validation
root. `git diff --check` passes. Primary staged tracker SHA-256 remains
`9ab829e44fa8df7ba1ee86e794518acd97129cea6af416966fbdda4b33f83426`.

No broad finalizer or Markdown-lint completion is claimed here; S-022 remains gated.
Shell implementation is unchanged since candidate 13's passing shell lint and
will receive the required final S-022 check. Default-off, local-only and no Core 05
publication claim remain explicit. F-022/RB-006 require a clean new executable
freeze and the complete unchanged qualification matrix. Shorter startup does not
relax relative elapsed or CPU budgets, and memory variation still needs external
qualification. Roll back this repair coherently across owner prose, catalog API,
private consumer, tests and generated input identities if necessary; preserve all
evidence. Successor S-021 is ready; S-022 is NOT STARTED.


### S-021 candidate 15 execution entry — IN_PROGRESS

Begins after the saved S-020 selected-row completion checkpoint. The clean isolated
checkout `/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary`
is on branch `codex/instrumentation-qualification-15`, commit
`7b2371bb63fb1230e0694f3ae409894d27a476cb`, tree
`38d488d15f20fcc07cbcb06d62698445b45c480c`, parent candidate 14
`59ea4783a8934de6f15e6b4d77a76f4bd8f2f5cf`. Seven paths differ: the NLSpec,
Make-generated topology render index, private row CLI, catalog facade/loader,
and catalog fixture/command-case registration. No temporary profiling code is
included. `/tmp/cartulary-qualification-freeze-15.json` is copied to the primary
row-selection validation root. The alternate-index freeze preserves primary staged
content and excludes the controlling tracker. The primary index is byte-identical
before/after this freeze; its current hash is recorded in that journal.

```sh
CARTULARY_TEST_RUN_ID=instrumentation-s021-qualification-15 \
CARTULARY_TEST_RESULTS_DIR=/dev/shm/cartulary-instrumentation-qualification-15 \
make harness-instrumentation-qualify QUALIFICATION_PLAN=tools/harness_instrumentation_qualification_plan.json
```

The complete five-workload, six-treatment matrix retains the same seed, cold and
warm-up requirements, five/six measured-block rule, preparation posture and reader
schedule. CPU arithmetic mean must remain strictly below 2% per cell; all other
budgets and eligibility/stability rules remain unchanged. Result storage uses
WSL2 tmpfs for wall time only. The external supervisor preserves command/environment,
source, mount, timing and exit, then copies and digest-verifies all evidence into
isolated `.cartulary/test-results/instrumentation-s021-qualification-15/`.
Original RAM evidence and all earlier failures remain retained. No source edits,
profiling or verification run concurrently with this cohort. S-022 is gated and
unstarted; default-off, local-only and no Core 05 publication claim remain explicit.

Candidate-15 published identities (from its typed manifest):

- `source_commit`: `7b2371bb63fb1230e0694f3ae409894d27a476cb`.
- `source_digest`: `sha256:245a3a514d17b41a98547b7f94ff5d675357ab864fb877bb92ba66b8b7b2efa0`.
- `plan_digest`: `sha256:1152b853dc9f2c8ceeb20343245e305806b729ac413f8cf0a686118627ba38af`.
- `policy_digest`: `sha256:c3f10c497b2c3eed5e0b647a02316d375b139fc5621d8b152ef93541bb6f4930`.
- `toolchain_digest`: `sha256:4b62e2558e6dcbe1d3c632fb9c0bb47e5a89e5dda08f103a72f61347b5ef1810`.
- `system_digest`: `sha256:546793d134ea4063515a25068aa22b78078fc1613c1ee56fb5f9e41c5c550588`.


### S-021 candidate 15 rejection checkpoint — FAILED; saved before S-020 repair

Candidate 15 remains bound to clean commit
`7b2371bb63fb1230e0694f3ae409894d27a476cb` and source digest
`sha256:245a3a514d17b41a98547b7f94ff5d675357ab864fb877bb92ba66b8b7b2efa0`.
Its exact command, typed identities and environment are recorded in the preceding
execution entry and `archive-command.json`. The owning Make command exited 2
in 3,450,890.184282 ms. All five workloads completed 48 attempts each: one cold,
one discarded warm-up and six measured blocks for every treatment. All 240
attempts and all 30 cells are eligible. No command, measurement-loss or cleanup
failure occurred. Every attempt and exclusion remains in the manifest; none was
removed for an unfavorable valid observation. The user confirms no other
intensive work was running. That is recorded context, not proof of a timing cause.

The supervisor copied and SHA-256 verified 8,285 files into isolated
`.cartulary/test-results/instrumentation-s021-qualification-15/`;
`archive-digests.json` and `archive-command.json` record successful verification.
Original `/dev/shm/cartulary-instrumentation-qualification-15/` evidence remains.
RAM storage is a wall-time choice, not a CPU concession. Candidate 14 and every
older failure, interruption, probe and source branch remain preserved.

Independent public validation:

```sh
CARTULARY_TEST_RUN_ID=instrumentation-s021-validation-15 \
make harness-performance-check EVIDENCE_ROOTS_FILE=/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary/.cartulary/test-results/instrumentation-s021-qualification-15/qualification/qualification.json
```

This exits Make 2 / owning validator 13 in 940.922734 ms. Exact argv, output and
exit are saved under isolated
`.cartulary/test-results/instrumentation-s021-validation-15/`. The separately
retained `independent-arithmetic.py` uses standard-library equations and endpoint
integration, imports no production algorithms, and matches all 30 public cell
statistics and numeric dispositions. Signed observations, paired errors/MADs,
all leave-one-block outcomes and the complete table are retained in
`independent-arithmetic.json` and `independent-table.md`.

Thirteen cells fail numeric gates: nine outer-p90, seven canonical-p90, two
outer-median, two canonical-median, five RSS and one CPU failure (some cells
fail several gates). CPU arithmetic mean remains strictly below 2% in every
cell except Go/basic/4 at 2.028136969796112%; that failure is not rounded down.
Maximum finalization union upper bound is 87.052347 ms, passing 500 ms. All five
workloads are unstable after six blocks because removing a block changes an
individual gate. No actual elapsed window exceeds the 10% MAD threshold.
Numeric cell passes do not override workload instability.

| Workload | Treatment | Blocks | CPU mean % | RSS upper MiB | Canonical median / p90 % | Outer median / p90 % | Final ms | Failed numeric gates |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| go | off/0 | 6 | 0.0000 | 0.0000 | 0.0000 / 0.0000 | 0.0000 / 0.0000 | 0.000 | none |
| go | off/1 | 6 | -1.1872 | 4.1627 | -0.1633 / 1.9231 | -0.3581 / 1.0971 | 0.000 | none |
| go | off/4 | 6 | -0.0852 | 5.3365 | 0.0062 / 0.9677 | 0.1951 / 3.9661 | 0.000 | outer_p90 |
| go | basic/0 | 6 | 1.7727 | 17.1604 | 1.5975 / 13.6364 | 2.0879 / 5.0250 | 57.295 | canonical_p90, outer_median, outer_p90 |
| go | basic/1 | 6 | 1.2977 | 9.6968 | 3.5195 / 6.4935 | 2.5269 / 4.0911 | 87.052 | canonical_median, canonical_p90, outer_median, outer_p90 |
| go | basic/4 | 6 | 2.0281 | 15.4378 | 3.0347 / 8.4416 | 1.4122 / 5.1673 | 59.475 | canonical_median, canonical_p90, outer_p90, cpu_mean |
| vitest | off/0 | 6 | 0.0000 | 0.0000 | 0.0000 / 0.0000 | 0.0000 / 0.0000 | 0.000 | none |
| vitest | off/1 | 6 | -0.9495 | 3.1165 | -1.3613 / 1.7140 | -0.1737 / 0.9924 | 0.000 | none |
| vitest | off/4 | 6 | -0.4858 | 3.4556 | -0.3090 / 1.1264 | -0.5150 / 1.0158 | 0.000 | none |
| vitest | basic/0 | 6 | 0.3175 | 11.4253 | 0.2316 / 2.8644 | 1.3574 / 3.1118 | 67.842 | outer_p90 |
| vitest | basic/1 | 6 | 0.5929 | 21.4960 | 0.5966 / 2.3411 | 0.4827 / 2.3912 | 54.482 | none |
| vitest | basic/4 | 6 | 0.9298 | 17.0384 | -0.2031 / 2.7952 | 1.1317 / 2.3089 | 59.034 | none |
| shell | off/0 | 6 | 0.0000 | 0.0000 | 0.0000 / 0.0000 | 0.0000 / 0.0000 | 0.000 | none |
| shell | off/1 | 6 | 1.1975 | 38.1004 | 0.7541 / 1.4249 | 0.7178 / 1.6193 | 0.000 | rss_upper |
| shell | off/4 | 6 | 0.3952 | 23.4470 | 0.5374 / 1.5951 | 0.4232 / 1.6293 | 0.000 | none |
| shell | basic/0 | 6 | 0.5070 | 32.4125 | 0.3321 / 1.2049 | 0.4784 / 1.1369 | 61.482 | rss_upper |
| shell | basic/1 | 6 | 1.2900 | 60.1405 | 1.0668 / 1.6703 | 1.2088 / 1.8505 | 50.823 | rss_upper |
| shell | basic/4 | 6 | 0.7949 | 36.1211 | 0.1625 / 1.2199 | 0.3085 / 1.4792 | 56.766 | rss_upper |
| service | off/0 | 6 | 0.0000 | 0.0000 | 0.0000 / 0.0000 | 0.0000 / 0.0000 | 0.000 | none |
| service | off/1 | 6 | 0.0091 | 18.9025 | 0.1489 / 1.9227 | 0.1855 / 1.6898 | 0.000 | none |
| service | off/4 | 6 | 0.2099 | 9.1439 | 0.2433 / 1.9504 | 0.3438 / 1.6331 | 0.000 | none |
| service | basic/0 | 6 | 0.4200 | 17.7402 | 0.9018 / 13.5596 | 1.0574 / 11.5603 | 57.174 | canonical_p90, outer_p90 |
| service | basic/1 | 6 | 0.6475 | 23.2607 | 0.8149 / 13.9037 | 1.0794 / 12.3041 | 59.108 | canonical_p90, outer_p90 |
| service | basic/4 | 6 | 0.6816 | 28.9547 | 0.0095 / 12.4179 | 0.4086 / 10.5178 | 60.874 | canonical_p90, outer_p90 |
| browser | off/0 | 6 | 0.0000 | 0.0000 | 0.0000 / 0.0000 | 0.0000 / 0.0000 | 0.000 | none |
| browser | off/1 | 6 | 0.1142 | 9.9056 | 0.4587 / 0.7462 | 0.3235 / 0.6531 | 0.000 | none |
| browser | off/4 | 6 | 0.5465 | -1.5971 | -0.2226 / 1.2429 | -0.0585 / 1.1400 | 0.000 | none |
| browser | basic/0 | 6 | 0.4069 | 28.9074 | -0.1679 / 1.7916 | -0.0598 / 1.5931 | 65.770 | none |
| browser | basic/1 | 6 | 0.6507 | 37.0177 | 0.2769 / 5.5945 | 0.3370 / 4.9415 | 63.646 | canonical_p90, outer_p90, rss_upper |
| browser | basic/4 | 6 | 0.3199 | 21.9005 | -0.0342 / 0.4013 | -0.0165 / 0.7523 | 63.409 | none |


`component-attribution.json` records bounded, read-only inspection of already
published artifacts. Go's 18 basic measured observations have zero periodic
sweeps; collector self CPU is only a component diagnostic. Rewriting process
discovery is therefore not supported as a repair for the short-Go failure.
Signed Go phase differences show about 2–3 ms setup and 1–3 ms wrapper additions,
with more variable execution differences. Private publication and startup need
measured attribution before selecting a repair.

Service database-drop totals are about 1.6–1.7 seconds through block 3 and the
first three off treatments of block 4. At order 177 they rise to about 2.9–3.0
seconds, persisting in later off and basic observations. The first migration
scratch database drop dominates; migration activity timings rise much less than
total row duration. All cleanup statuses succeed. This is a wall-time observation,
not evidence of extra observer CPU or permission to discard the block. Shell
RSS also fails in off/1, so collector memory alone cannot explain its dispersion.

The isolated executable checkout is still clean. Primary staged tracker SHA-256
is unchanged at `9ab829e44fa8df7ba1ee86e794518acd97129cea6af416966fbdda4b33f83426`.
F-022/RB-006 remain open; no executable source is qualified. This saves the failed
S-021 workstream checkpoint before reopening S-020. S-022 is NOT STARTED;
finalization, broad end-of-run checks, Markdown-lint coverage and final handoff
are not claimed. Default-off, local-only and no Core 05 publication claim remain.
The next repair must use preserved component evidence, then functionally validate,
save its S-020 completion, freeze clean executable source and run a new complete
matrix. Budgets, workload selections, cadence and stability rules remain unchanged.

### S-020 publication/startup attribution entry — IN_PROGRESS

Begins only after the saved candidate-15 S-021 rejection. Measure synchronous
private publication, clock reads and schema validation on exact public Make
workloads before selecting an implementation change. Temporary probes must be
saved with before/instrumented source and removed byte-for-byte afterward;
component observations cannot qualify or replace the 240-attempt failure.
Evaluate scratch placement for wall time only. Creation and stale cleanup must
resolve a scratch override consistently; host-wide admission must retain one
namespace. Retained evidence, ownership and recovery publication retain their
security and durability requirements. Preserve private-message atomicity,
incomplete beginnings, shutdown rejection, canonical timing and primary failures.
No PostgreSQL cleanup or service-storage change is authorized merely by the
observed timing shift: first establish the relevant contract and causal evidence.
CPU mean stays below 2% per cell; all other budgets and schedules remain unchanged.
S-021 is blocked on a completed validated repair and a fresh full cohort; S-022
remains unstarted. The primary index and all prior evidence stay untouched.

S-020 publication attribution checkpoint — IN_PROGRESS: the public owner guide
passed (`make task-guide ROLE=module-author OWNER=harness.evidence_accounting`).
Sixteen exact `make test-slice` component commands passed: Go off/basic with one
discarded warm-up and four alternating measured pairs, and shell off/basic with
one discarded warm-up and two measured pairs, all with cache mode off. Full argv,
IDs, stdout/stderr, before/instrumented sources, 132 per-process numeric profiles,
independent aggregation and restoration hashes remain at primary
`.cartulary/test-results/instrumentation-f022-publication-attribution-01/`.
The five instrumented source files were restored byte-for-byte automatically.
Nested component counters overlap and cannot be summed into total observer cost.

The result rejects the private-fsync hypothesis. Go graph private execution
publication takes median 0.790 ms / 802.5 microseconds CPU; row publication takes
0.862 ms / 1,152.5 microseconds CPU. Their fsync calls total only 0.012 / 0.018 ms.
The current `/tmp` mount is already tmpfs; qualification results also used tmpfs.
Clock reads total about 0.344 ms in the graph and 0.225 ms in the row process.
Normal-disk retained graph publication takes 23.271 ms basic versus 15.911 ms off,
which is wall-time evidence and is not the qualification RAM-storage environment.
No durability or scratch-placement change is justified by these measurements.
The scratch-override creation/cleanup discrepancy remains a source-review note,
not an attributed F-022 cause or an adopted storage migration.

Graph schema validation is about 67,375 microseconds CPU basic versus 54,336 off;
row validation is about 42,445 versus 42,189. Those are component medians, not a
paired qualification statistic. Full catalog syntax discovery grows graph RSS
from about 67 MiB to 277–278 MiB and leaves about 186 MiB of V8 heap capacity;
source hashing then observes about 331–334 MiB RSS. The next bounded experiment
will remove unused TypeScript parent links from syntax-only selector discovery,
compare complete catalog acceptance/digests, CPU and memory, and retain a losing
experiment rather than adopting it. Existing shared-root schema linkage remains;
previous failed fragment-sharing experiments are not repeated. S-020 is not
complete and no new S-021 cohort is started.

S-020 syntax/source-memory repair checkpoint — IN_PROGRESS:

- `instrumentation-f022-selector-links-profile-01` passed
  `make harness-command-surface-contract` in 72,336.305011 ms. Twenty-two fresh
  component processes (one discarded pair plus ten alternating measured pairs)
  retain equal selected-row, full catalog and summary digests. Removing unused
  TypeScript parent links yields paired median CPU delta -127,564 microseconds
  (MAD 23,300) and wall delta -96.249281 ms (MAD 9.145714). RSS delta
  -14,198,784 bytes has MAD 17,121,280, so it does not establish stable memory
  improvement. Adopt the syntax-only representation because discovery never uses
  parent navigation and the CPU cost reduction is measured. Both syntax entry
  points omit parent links; complete source closure remains unchanged.
- Source hashing now uses one reusable 64 KiB buffer across regular files,
  retaining v2 byte/mode/path identity. Descriptor and path metadata checks reject
  truncation, growth, replacement and permission changes during hashing; read
  failure closes the descriptor. No file-sized allocation or new public option
  is introduced. Symbolic links retain their own link-text digest and are not
  followed. Documentation filtering still precedes filesystem access.
- A command-surface regression independently authors the aggregate serialization
  and hashes binary, empty, multi-chunk and symlink inputs. Synchronous fault
  injection proves mutation rejection and descriptor closure, with core filesystem
  bindings restored in `finally`. It is routed under AC-066 and is outside the
  selected shell qualification workload; that workload is not padded.
- `instrumentation-f022-streamed-source-generate-01` passed `make generate`
  in 12,690.592275 ms. `instrumentation-f022-streamed-source-profile-01` passed
  `make harness-command-surface-contract` in 73,103.908508 ms, including the
  new identity/fault case and 22 fresh-process component comparisons after full
  catalog loading. All 6,430 source entries and aggregate digests match.
  Streaming's paired median CPU delta is -30,239 microseconds (MAD 14,851.5),
  wall -14.665110 ms (MAD 5.142081), post-hash RSS -42,010,624 bytes
  (MAD 8,646,656), and external/array-buffer bytes -28,258,872 (MAD zero).
  Heap-used observations rise and are retained separately; neither RSS endpoint
  nor lifetime maximum is a steady-memory qualification claim.
- All component commands passed. Full commands, source snapshots, observations,
  signed arithmetic and restoration proofs remain under primary
  `.cartulary/test-results/instrumentation-f022-selector-links-01/` and
  `.cartulary/test-results/instrumentation-f022-streamed-source-01/`.
  Temporary test entrypoints were restored byte-for-byte. `git diff --check`
  passes. No schema acceptance change or migration translator is introduced;
  a concurrently changing source now fails identity construction deliberately.

Functional owner/real-row and projection gates remain required. These repairs
reduce unnecessary catalog work and unbounded source-buffer churn; they do not
establish that any failed qualification cell now passes or explain PostgreSQL
cleanup variation. S-020 stays IN_PROGRESS, S-021 requires a later clean freeze
and complete new matrix, and S-022 remains NOT STARTED. All budgets are unchanged.

### S-020 syntax/source-memory completion — DONE

Saved before beginning the next S-021 workstream. The two measured structural
changes are adopted: selector syntax trees omit unused parent links and source
snapshot hashing reuses a 64 KiB buffer with descriptor/path mutation checks.
Source identity remains v2 and unchanged for stable bytes, modes and paths;
concurrent modification deliberately fails instead of publishing mixed identity.
The independent identity/fault fixture is outside the frozen workload selection.
No diagnostic schema, workload, cadence, budget, service posture, publication
durability or scratch placement changed. The shared schema validator remains
in place. No qualification claim follows from the component improvements.

All twelve public functional/projection commands passed on primary source.
Their full argv, fresh IDs, durations, stdout/stderr and exit codes are recorded
at `.cartulary/test-results/instrumentation-f022-source-memory-validation-01/commands.json`.
Run IDs use `instrumentation-f022-source-memory-<name>-01`:

| Name | Public command / selected scope | Exit | Elapsed ms |
| --- | --- | --- | --- |
| catalog | `make test-slice OWNER=harness.test_catalog ROWS=harness.test_catalog.behavior.routing_contract HARNESS_DIAGNOSTICS=basic` | 0 | 102668.490897 |
| go | `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.local_migration_activities HARNESS_DIAGNOSTICS=basic` | 0 | 2621.078851 |
| vitest | `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.work_pin_presentation_removal_27944912ba HARNESS_DIAGNOSTICS=basic` | 0 | 4675.734436 |
| shell | `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.current_epoch_evidence HARNESS_DIAGNOSTICS=basic CARTULARY_HARNESS_CACHE_MODE=cold` | 0 | 21145.505622 |
| service | `make service-backed-test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.schema_bootstrap HARNESS_DIAGNOSTICS=basic` | 0 | 15964.445948 |
| browser | `make service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.https_trust HARNESS_DIAGNOSTICS=basic` | 0 | 26864.695744 |
| evidence | `make harness-evidence-contract` | 0 | 19326.773216 |
| command | `make harness-command-surface-contract` | 0 | 51646.508778 |
| drift | `make generate-drift` | 0 | 18600.264255 |
| policy | `make generated-artifact-policy-check` | 0 | 2103.792394 |
| shape | `make json-shape-check` | 0 | 4487.577019 |
| scripts | `make lint-scripts` | 0 | 1741.546575 |

Fresh exact-run validation of the browser root passed
`make harness-observability-check RESULTS_DIR=<exact-browser-root>` and all four
human/JSON `make explain-run RESULTS_DIR=<exact-browser-root>` combinations of
`DETAIL=resources|performance`. Both JSON views return the identical normalized
model; human views show the same timing, receipt and resource observations.
The browser command and its cleanup passed; resource coverage is observed_partial,
not a claim of unsampled physical coverage. Exact reader commands and parity
are retained in `reader-commands.json` and `reader-parity.json` at the validation
root. `git diff --check` passes. The staged tracker SHA-256 remains
`9ab829e44fa8df7ba1ee86e794518acd97129cea6af416966fbdda4b33f83426`.

Component experiments, rejected hypotheses and every prior failed cohort remain
retained. Their signed measurements do not establish steady RSS or an elapsed
qualification pass. PostgreSQL cleanup variation remains unattributed. The user
confirmed no other intensive work; there is no host-contention exclusion.
S-021 may now freeze the five-path executable delta from candidate 15 and run
the complete unchanged matrix. F-022/RB-006 stay open; S-022 remains NOT STARTED.
Default-off, local-only diagnostics and no Core 05 publication claim remain.

### S-021 candidate 16 entry — IN_PROGRESS

Begins after the saved S-020 syntax/source-memory completion. Isolated branch
`codex/instrumentation-qualification-16` is clean at commit
`42831e37e9be710b02bb646152efcecc1c9dac23`, tree
`226e151e293685a96e042a80893e008f95b9e766`, parent candidate 15
`7b2371bb63fb1230e0694f3ae409894d27a476cb`. The five-path delta is the
selector, source snapshot, authored contract case/routing and generated render
index. The alternate-index transaction preserved the primary index byte-for-byte
and preserved the original staged tracker. Freeze evidence is saved at
`.cartulary/test-results/instrumentation-f022-source-memory-validation-01/freeze-16.json`.

The public command in `/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary` is
`CARTULARY_TEST_RUN_ID=instrumentation-s021-qualification-16 CARTULARY_TEST_RESULTS_DIR=/dev/shm/cartulary-instrumentation-qualification-16 make harness-instrumentation-qualify QUALIFICATION_PLAN=tools/harness_instrumentation_qualification_plan.json`.
The RAM result parent is retained for wall time only; all CPU accounting remains
complete and unchanged. Durable command/evidence archive is
`.cartulary/test-results/instrumentation-s021-qualification-16/` in that checkout.
No source edits, tests, generators or profiling run concurrently with the cohort.
Every attempt, extension, failure and exclusion will be preserved and independently
validated. No prior favorable cell is reused. CPU mean must stay below 2% per cell;
all other budgets, stability rules, workload selections, preparation and reader
schedules remain unchanged. S-022 is blocked pending successful qualification.

### S-021 candidate 16 rejection checkpoint — FAILED

This completes the candidate-16 workstream before any further implementation.
The full frozen cohort finished with Make exit 2 in 3,369,158.534228 ms.
All 240 workload commands passed and all attempts/cells are eligible: no lifecycle,
measurement, quietness, cleanup or finalization-coverage failure explains rejection.
Fourteen of 30 cells fail numeric gates. All five workloads remain unstable after
six blocks because removing a block changes at least one individual gate;
no elapsed window exceeds the 10% MAD condition. Cold/warm-up exclusions are the
specified phases only. Negative observations remain signed, without an improvement
claim. CPU mean remains strictly below 2% as required; thresholds are unchanged.

Source commit is `42831e37e9be710b02bb646152efcecc1c9dac23`; source digest is
`sha256:5ebe5b34912cc3807283d75f5a68b602ea3c3ef608e05ab00c9758b6802a8b49`.
Plan digest is `sha256:1152b853dc9f2c8ceeb20343245e305806b729ac413f8cf0a686118627ba38af`;
policy digest is `sha256:c3f10c497b2c3eed5e0b647a02316d375b139fc5621d8b152ef93541bb6f4930`.
Toolchain and system digests match candidate 15. The user confirmed no other
intensive work; no contention-based exclusion is asserted. RAM storage is for
wall time and does not reduce the CPU coverage or commitment.

Durable archive in the isolated checkout is
`.cartulary/test-results/instrumentation-s021-qualification-16/`.
Its 8,285 files are digest-verified against the original RAM root, which remains
retained. `archive-command.json` preserves command, environment, mounts, source,
Make PID, start/end, exit and elapsed time; `archive-digests.json` closes the copy.
The exact manifest is nested at
`instrumentation-s021-qualification-16/qualification/qualification.json` within
that archive. No previous archive or qualification branch was overwritten.

The public command was
`CARTULARY_TEST_RUN_ID=instrumentation-s021-validation-16 make harness-performance-check EVIDENCE_ROOTS_FILE=/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary/.cartulary/test-results/instrumentation-s021-qualification-16/instrumentation-s021-qualification-16/qualification/qualification.json`.
It rejected the cohort with Make exit 2 / inner exit 13 in 989.148926 ms.
Full command/stdout/stderr and independently authored standard-library arithmetic
are retained at isolated `.cartulary/test-results/instrumentation-s021-validation-16/`.
All 30 public cell dispositions and displayed statistics match independent
calculation. The table below comes from that arithmetic, not production analysis.

| Workload | Treatment | Blocks | CPU mean % | RSS upper MiB | Canonical median / p90 % | Outer median / p90 % | Final ms | Failed numeric gates |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| go | off/0 | 6 | 0.0000 | 0.0000 | 0.0000 / 0.0000 | 0.0000 / 0.0000 | 0.000 | none |
| go | off/1 | 6 | -0.1475 | 10.1677 | 0.6240 / 2.2581 | -1.1423 / 0.4344 | 0.000 | none |
| go | off/4 | 6 | 0.9019 | 10.4603 | 1.8809 / 5.5901 | 0.1627 / 2.9519 | 0.000 | canonical_p90 |
| go | basic/0 | 6 | 1.2063 | 21.3097 | 4.0616 / 5.9561 | 0.8034 / 2.8499 | 53.054 | canonical_median, canonical_p90 |
| go | basic/1 | 6 | 1.7932 | 10.3228 | 0.9497 / 5.0157 | 1.1079 / 3.0224 | 57.212 | canonical_p90, outer_p90 |
| go | basic/4 | 6 | 2.3285 | 17.4728 | 1.7168 / 6.8966 | 2.0529 / 3.9034 | 58.379 | canonical_p90, outer_median, outer_p90, cpu_mean |
| vitest | off/0 | 6 | 0.0000 | 0.0000 | 0.0000 / 0.0000 | 0.0000 / 0.0000 | 0.000 | none |
| vitest | off/1 | 6 | 1.4271 | 12.9792 | 1.0497 / 5.3259 | 1.0970 / 3.6815 | 0.000 | canonical_p90, outer_p90 |
| vitest | off/4 | 6 | 0.4301 | 10.6024 | 1.4146 / 1.8747 | 0.5250 / 1.7286 | 0.000 | none |
| vitest | basic/0 | 6 | 1.5812 | 24.3166 | 1.5271 / 3.0717 | 1.5143 / 2.1634 | 57.011 | canonical_p90 |
| vitest | basic/1 | 6 | 2.5038 | 17.1524 | 1.7628 / 3.0290 | 1.9793 / 4.3310 | 58.701 | canonical_p90, outer_p90, cpu_mean |
| vitest | basic/4 | 6 | 2.5109 | 15.9826 | 1.4898 / 3.5410 | 1.5302 / 2.4835 | 59.236 | canonical_p90, cpu_mean |
| shell | off/0 | 6 | 0.0000 | 0.0000 | 0.0000 / 0.0000 | 0.0000 / 0.0000 | 0.000 | none |
| shell | off/1 | 6 | 0.5992 | 34.1922 | 0.1979 / 0.8940 | 0.2470 / 0.8509 | 0.000 | rss_upper |
| shell | off/4 | 6 | 0.3606 | 11.4436 | 0.2080 / 1.2177 | 0.3696 / 1.3361 | 0.000 | none |
| shell | basic/0 | 6 | 0.8788 | 44.6421 | 0.3846 / 1.1406 | 0.4684 / 1.2637 | 58.348 | rss_upper |
| shell | basic/1 | 6 | 0.3417 | 26.4155 | 0.0568 / 1.0361 | 0.3488 / 1.0470 | 57.227 | none |
| shell | basic/4 | 6 | 0.3378 | 52.9807 | -0.0719 / 0.5567 | 0.0633 / 0.8718 | 60.549 | rss_upper |
| service | off/0 | 6 | 0.0000 | 0.0000 | 0.0000 / 0.0000 | 0.0000 / 0.0000 | 0.000 | none |
| service | off/1 | 6 | 0.4209 | 13.1289 | -0.1759 / 1.0085 | 0.0345 / 1.2632 | 0.000 | none |
| service | off/4 | 6 | 0.2552 | 9.7884 | 0.1887 / 0.8031 | 0.2386 / 0.9648 | 0.000 | none |
| service | basic/0 | 6 | 0.7044 | 24.7672 | -0.0406 / 1.3966 | 0.2491 / 1.5917 | 59.231 | none |
| service | basic/1 | 6 | 0.7607 | 20.7503 | 0.1615 / 0.6736 | 0.4990 / 0.9331 | 58.309 | none |
| service | basic/4 | 6 | 0.9115 | 27.0970 | 0.3709 / 4.9302 | 0.3301 / 4.6182 | 57.124 | canonical_p90, outer_p90 |
| browser | off/0 | 6 | 0.0000 | 0.0000 | 0.0000 / 0.0000 | 0.0000 / 0.0000 | 0.000 | none |
| browser | off/1 | 6 | 0.0672 | 12.3689 | 0.2823 / 1.0728 | 0.2482 / 1.0594 | 0.000 | none |
| browser | off/4 | 6 | 0.0796 | 16.4405 | -0.0462 / 2.6925 | -0.1353 / 2.5685 | 0.000 | none |
| browser | basic/0 | 6 | 0.4995 | 39.3185 | 0.0262 / 0.6012 | 0.3427 / 0.7718 | 108.958 | rss_upper |
| browser | basic/1 | 6 | 0.0058 | 27.9459 | 0.0358 / 0.6056 | -0.0282 / 1.1205 | 65.289 | none |
| browser | basic/4 | 6 | 0.5309 | 35.1132 | 0.0794 / 0.7830 | 0.1511 / 0.8975 | 60.069 | rss_upper |

Three CPU cells fail: Go/basic/4 at 2.3285169014963745%, Vitest/basic/1 at
2.503840057699092%, and Vitest/basic/4 at 2.510884825639547%. Five RSS cells
fail: shell/off/1, shell/basic/0, shell/basic/4, browser/basic/0 and browser/basic/4.
Elapsed failures include nine canonical p90, one canonical median, five outer
p90 and one outer median gates. Every finalization gate passes; the maximum is
108.957705 ms against 500 ms. The bounded-source and syntax changes are valid
functional repairs but do not establish a successful engineering qualification.

The isolated checkout remains clean. Primary staged tracker SHA-256 remains
`9ab829e44fa8df7ba1ee86e794518acd97129cea6af416966fbdda4b33f83426`.
F-022/RB-006 remain open. S-022 is NOT STARTED; finalization, broad final checks,
tracker Markdown-lint coverage and handoff completion are not claimed. No revision
is qualified. Default-off, local-only and no Core 05 publication claim remain.
A successor must identify and measure a structural repair, save S-020 completion,
freeze changed executable source and rerun the full affected matrix. An unchanged
rerun or favorable-cell reuse is not an acceptable successor.

### S-020 short-workload process CPU attribution entry — IN_PROGRESS

Begins after the saved candidate-16 rejection. Existing publication profiles show
that private fsync and clock reads are small. Per-schema reanalysis further shows
only about 2–3.5 ms of execution-record validation in the Go row/graph processes;
the earlier aggregate schema difference includes variable common family validation.
That evidence does not justify another compiler refactor by itself. Do not repeat
the failed small-fragment sharing experiment or optimize a guessed bottleneck.

Measure absolute Node process CPU alongside per-schema component counters on
unchanged exact Go and Vitest public selections, with one discarded warm-up and
five alternating off/basic pairs each. Keep complete source before/instrumented
copies, commands and all signed observations, then restore the sole temporary
source edit byte-for-byte. This is attribution only; component medians and process
peaks cannot qualify total CPU or steady RSS. No source edits will reach the clean
candidate-16 checkout or alter its archived failure. No schema, budget, workload,
reader cadence, preparation or service-posture change is authorized by this probe.
S-021 requires an adopted, validated repair and another clean freeze; S-022 stays
NOT STARTED. The original staged tracker and all prior evidence remain preserved.

S-020 process/schema attribution checkpoint — IN_PROGRESS: all 24 exact public
Go/Vitest component commands passed (one discarded warm-up pair and five
alternating measured pairs per workload), with 108 numeric process profiles.
The sole temporary `harness-contract.mjs` edit was restored byte-for-byte.
Evidence, commands, before/instrumented source, restoration digest and independent
paired arithmetic remain at primary
`.cartulary/test-results/instrumentation-f022-process-schema-attribution-01/`.
Node process CPU excludes child processes; schema counters are contained within
it and cannot be added. Median paired graph CPU is +11,601 microseconds for Go
(MAD 10,885) and +35,532 for Vitest (MAD 2,403). Row-process deltas are +3,307
(MAD 3,586) and +6,310 (MAD 2,108). Graph schema deltas are +21,004 and +8,527;
the overlapping, noisy components do not establish total cost or a schema cause.
No compiler change is justified yet. Measure synchronous collector startup and
native child creation directly, retaining the same source-restoration guarantees,
before choosing a repair. S-020 remains open; no new S-021 freeze is authorized
by component success alone. Budgets and matrix remain unchanged.

S-020 startup attribution checkpoint — IN_PROGRESS: all 24 Go/Vitest component
commands passed and both temporarily instrumented collector files were restored
byte-for-byte. Evidence remains under primary
`.cartulary/test-results/instrumentation-f022-collector-startup-attribution-01/`.
Basic collector initialization costs median 2,831 / 2,985 microseconds CPU in
Go/Vitest (MAD 34 / 92), including spawn at 1,555 / 1,606 microseconds (MAD 30 / 16).
This rejects large-process-fork startup as the dominant cause; do not move helper
ownership or weaken shutdown merely to reduce a small component. Retained native
helper examples report about 3.7 ms CPU with zero Go sweeps and 7.0 ms with one
Vitest sweep. They remain lower-bound component diagnostics.

Candidate-16 independent CPU decomposition is retained in validation-16
`cpu-components.json`: readers average roughly 4 ms each, while Vitest/off/1's
workload CPU delta is about +57.9 ms versus +2.8 ms for off/4. Thus workload CPU
variation is material and no single reader or helper counter proves total cost.
The next bounded experiment uses the pinned compiler's supported `inlineRefs: 2`
option to inline trivial predicates while retaining shared complex roots. Compare
retained-document, producer and family-metadata workloads against the current
compiler, with independent acceptance and recursion/closure tests. Restore source
if the component result loses. No new acceptance rule, schema version, runtime
compiler, qualification exclusion or relaxed threshold is introduced.

### S-020 post-candidate-16 attribution checkpoint — IN_PROGRESS

The attribution experiments are complete and retained; no new executable repair
is adopted after candidate 16. The trivial-predicate compiler experiment passed
`make generate` (`instrumentation-f022-trivial-inline-generate-01`,
11,371.613375 ms) and `make harness-contract CARTULARY_HARNESS_CACHE_MODE=off`
(`instrumentation-f022-trivial-inline-profile-01`, 117,201.476902 ms).
The latter includes independent acceptance, recursion, dynamic-reference,
shared-root, opposite-cycle-order, lazy closure and package-free bootstrap checks,
plus 96 fresh-process observations across retained, producer and family datasets.
One warm-up and 15 alternating measured pairs per dataset are preserved.

First-validation CPU paired medians are -2,273 microseconds / MAD 243 for retained
inputs, -1,739 / MAD 813 for producers, and -99 / MAD 398 for family metadata.
Producer first-validation wall time rises +1.047788 ms / MAD 1.542938; producer
RSS delta is -8,192 bytes / MAD 4,231,168. Family total CPU delta is -307
microseconds / MAD 5,473. These results do not establish a material repair of the
failed qualification cells or a steady-memory improvement. Reject the experiment
rather than starting another full cohort on this weak attribution. The supported
compiler option was acceptance-equivalent but is not adopted as a new setting.

Primary `.cartulary/test-results/instrumentation-f022-trivial-inline-01/` retains
both complete 195-module generated closures, before/candidate compiler, fixtures,
probe, temporary test source, exact commands, all observations and independent
signed arithmetic. Restoring authored compiler and running `make generate`
(`instrumentation-f022-trivial-inline-restore-01`) passed in 11,465.601457 ms.
Both generated closure and temporary test source are restored byte-for-byte.
An alternate-index audit confirms the entire executable source tree equals
candidate 16 tree `226e151e293685a96e042a80893e008f95b9e766`, excluding only
this controlling tracker. `restoration-audit.json` records this check, unchanged
primary index and staged tracker hash. `git diff --check` passes.

Current restart point: candidate 16 is a complete, independently validated failed
cohort, not a qualified revision. S-020 remains IN_PROGRESS for F-022 remediation;
S-021 cannot close and S-022 remains NOT STARTED. There are no active component
or qualification commands, no temporary source probes, and no candidate-17 freeze.
The shared-root schema validator, selected-row boundary, bounded source hashing
and parent-free syntax discovery remain functionally validated. The rejected
compiler experiment and all earlier failures remain preserved.

The unresolved work is to attribute workload CPU variation and sampled aggregate
RSS dispersion to specific process lifecycles, then adopt a structural repair
with measured benefit. In particular, an off/one-reader shell RSS failure and a
large Vitest/off/one-reader workload CPU difference rule out blaming collection
alone. A next investigation may collect component process/thread CPU and memory
lifecycle evidence through temporary Make-owned probes; it must preserve every
observation and remove instrumentation before validation/freeze. Per-schema and
collector-startup evidence already obtained must not be repeated as though unknown.

### S-020 systematic recovery planning checkpoint — IN_PROGRESS

The user requested a step back after more than twenty hours without a solution.
This tracker-only session replaces the next investigative action with the
Section 7 decision sequence. No component experiment, source edit, test, generator,
qualification, finalization or service action ran. Candidate 16 remains FAILED
under its original policy; no candidate 17 exists or is dispatched by this plan.
The latest CPU request is mean strictly below 2.5%; the unchanged executable plan
and NLSpec still say 2%. S-020/A must perform the coordinated versioned amendment
on implementation resumption, without relabeling historical evidence.

Newly inspected evidence distinguishes three unresolved questions: correctness
and disturbance of measurement, independent identical-run variation, and actual
incremental observer/workload cost. E-026 records support CPU normalized by outer
duration and the median/dispersion/error components of all five RSS failures.
E-027 records the absence of retained raw samples and the specific native and
graph-lifetime seams behind the proposed experiments. These are decision inputs,
not proof of a new root cause. Prior schema, publication and startup experiments
remain retained and are explicitly protected against unmotivated repetition.

Planning completeness is complete for this bounded recovery sequence; root-cause
proof and specific code movement remain pending experiments and consumer review.
S-020/A is ready for implementation resumption; B–E have explicit gates. S-021
new collection is blocked on S-020 completion and S-022 on successful S-021.
F-022/RB-006, T-007/T-008 and overall work remain IN_PROGRESS. Default-off,
local-only WSL2 scope and absence of a Core 05 publication claim remain unchanged.

The planning pass read the owner/framework/skill instructions, the exact native
measurement and qualification seams, their tests/routing, selected graph-runner
seams, and candidate-16 bound result files. It updated current readiness/policy
summaries and added the systematic recovery plan and E-026–E-028 only to this
controlling document. It did not rewrite any historical execution checkpoint.
`make lint-markdown` was skipped for the inspected write effects and omitted
tracker coverage documented in Section 10; no runtime pass is claimed.

E-028 preservation comparison passed: all 6,877 other tracked/non-ignored entries
retain their content and recorded file identities; raw index bytes and staged
tracker content are unchanged. The staged tracker SHA-256 remains
`9ab829e44fa8df7ba1ee86e794518acd97129cea6af416966fbdda4b33f83426`.
The admitted tracker remains a regular, single-link file. All twelve required
sections and serial workstream dependencies remain present. Read-only
`git --no-optional-locks diff --check` passed. Comparison details are outside the
repository at `/tmp/cartulary-systematic-plan-comparison.json`; the baseline is
`/tmp/cartulary-systematic-plan-ad37xeyh.json`. Ignored archives/caches were not
recursively snapshotted; no operation writing them was executed. Recheck source
and affected evidence at the next implementation entry.
No new ownership layer, weakened validation, delayed cleanup, altered workload,
extra idle time, discarded valid observation or threshold relaxation is justified
by these findings. Successful completion still requires a new validated repair,
a clean freeze, the full passing qualification matrix and the saved S-022 checkpoint.

`make agent-finalize`, final broad lint/contract checks, adding this tracker to
Markdown-lint coverage, and handoff closure are intentionally not performed as
S-022 work while its prerequisite fails. No eligible successful full warm `check`
run is claimed. Default-off, local-only diagnostics, explicit physical coverage
limits and no Core 05 publication claim remain unchanged.

### S-020/A accepted recovery execution entry — IN_PROGRESS

The user explicitly authorized implementation of the complete recovery plan.
The accepted retention decision is aggregate raw measurements for every
qualification attempt; detailed process attribution is confined to diagnostic
experiments. The implementation adds a closed explicit
`harness-instrumentation-diagnose DIAGNOSTIC_PLAN=<file>` operation, outside ordinary
checks, alongside the unchanged qualification entry point. Aggregate trace v1
and qualification result/manifest v2 publish together during B; CPU plan v3 is
adopted in A. All other budgets and the workload matrix remain unchanged.

The native supervisor remains the acquisition owner, secure loading closes raw
evidence, and pure analysis reconstructs summaries. Limits remain 64 MiB trace,
360,000 memory samples and 64 KiB records. Unknown measurements stay unavailable;
trace loss cannot qualify. Diagnostic plans admit only support-interference,
repeatability and attribution with closed workload selections, fixed counts,
balanced order and retained outcomes. Ordinary qualification retains no process
identities, commands or environment values. No new post-scan exception is created.

A records policy and protocols; B validates measurement and support; C executes
independent controls; D repairs only an attributed cause; E verifies the final
source and saves S-020 completion. Each checkpoint precedes its successor.
Qualification and S-022 remain blocked on those prerequisites. A 90-minute
experimental-collection limit without a reconciled cause ends exploration after
cleanup and a saved decision checkpoint, never by loosening a gate.

Entry preserves dirty primary `main` at `1d423c6e5f9cd204f8210fa9958c780fd933d293`.
`.cartulary/test-results/instrumentation-recovery-entry-01/` retains the 6,878-entry
content baseline and pre-edit tracker. Staged tracker SHA-256 remains
`9ab829e44fa8df7ba1ee86e794518acd97129cea6af416966fbdda4b33f83426`.
`CARTULARY_TEST_RUN_ID=instrumentation-recovery-guide-01 make task-guide ROLE=module-author OWNER=harness.evidence_accounting`
passed, confirming the focused `test-slice` and broader evidence-contract routes.
No candidate-17 freeze or qualification is started.

S-020/A first validation retained: `make generate` passed (11,846.379760 ms);
`make harness-evidence-contract` failed 105/106 (Make 2, 22,002.556644 ms),
root `instrumentation-recovery-policy-evidence-01`. The aggregate arithmetic
fixture's injected reader cost produced 2.1%, now correctly below the authorized
2.5% limit. Its expected rejection was stale; the independent injected value is
now 1.5 seconds, producing 3.1%, without changing production arithmetic.
Remaining commands were not run after this failure. Evidence/logs remain at
`.cartulary/test-results/instrumentation-recovery-policy-validation-01/`;
this is a policy-test correction, not a qualification exclusion.

### S-020/A policy and experiment contract checkpoint — DONE

Saved before B begins. REQ-848 now requires equal-weight signed CPU mean <2.5%,
plan v3 is the only current plan, and independent tests cover equality, below-limit
values, cell isolation, signed weighting and rejection of v1/v2 plans. No other
budget, workload, preparation rule or qualification cadence changed. REQ-850/851
adopt bounded aggregate retention/reconstruction and closed diagnostic protocols
before their implementation. Qualification v2/trace v1 publish together in B.

Validation commands from the primary checkout (exact roots use these run IDs):

| Command | Run ID | Result | Outer ms |
| --- | --- | --- | --- |
| `make generate` | `instrumentation-recovery-policy-generate-01` | PASS | 11846.38 |
| `make harness-evidence-contract` | `instrumentation-recovery-policy-evidence-01` | FAIL, retained | 22002.557 |
| `make harness-evidence-contract` | `instrumentation-recovery-policy-evidence-02` | PASS | 20195.444 |
| `make harness-command-surface-contract` | `instrumentation-recovery-policy-command-01` | PASS | 53413.135 |
| `make generate-drift` | `instrumentation-recovery-policy-drift-01` | PASS | 19415.862 |
| `make generated-artifact-policy-check` | `instrumentation-recovery-policy-policy-01` | PASS | 2135.943 |
| `make json-shape-check` | `instrumentation-recovery-policy-shape-01` | PASS | 4620.023 |

Full logs and command records remain in
`.cartulary/test-results/instrumentation-recovery-policy-validation-01/`.
The original failed fixture attempt is preserved. Generated validators and
attachments cut over through `make generate`, without a runtime translator.
The staged tracker digest is unchanged. S-020 overall remains IN_PROGRESS; B is
ready, C–E and qualification are gated, S-022 is unstarted.

### S-020/B reconstructible measurement entry — IN_PROGRESS

Implement bounded aggregate trace publication and single-pass secure loading,
then independently reconstruct retained summaries. Reject disappearance/replacement
as unavailable instead of zero. Establish sampling uncertainty before any control
or qualification can be eligible; ordinary trace evidence never retains process
identities. Detailed process attribution remains diagnostic-only. Prior candidate
evidence keeps its original formats and outcomes.

S-020/B first generation failed at authored routing validation: the newly added
`TestMeasurementResidentIdentity` selector was not ASCII-sorted. Result root
`instrumentation-recovery-measurement-generate-01` and full command logs in
`instrumentation-recovery-measurement-validation-01` are retained. The authored
selector ordering is corrected; formatting and tests did not run after this gate.
No experiment or qualification attempt has started.

B evidence contract attempt `instrumentation-recovery-measurement-evidence-01`
failed 107/108: the new synthetic envelope began its sole interval at 9900 ms
instead of the required zero origin. The fixture now uses local [0,10] ms and
keeps external observation timestamps separate. Production envelope validation
correctly rejected it and is unchanged. Generation attempt 02 and `format-go`
passed; the failed evidence and generation artifacts remain retained.

B evidence contract attempt 02 passed 108/108. Native owner attempt
`instrumentation-recovery-measurement-native-01` failed: the prior real-shell
fixture asserted all RSS sweeps were available. New identity checks exposed an
unavailable sample, correctly setting memory coverage/error to unavailable.
Accounting, reader lifecycle and finalization were complete; cancellation and
independent waited/adopted CPU checks passed. The fixture now checks truthful
unavailability without relaxing qualification eligibility. The retained failing
row and logs remain at that exact run root. This does not establish a performance
cause or permit qualification.

B native owner attempt 02, evidence attempts 03/04, command-surface contract,
generation drift and generated-artifact policy passed. JSON-shape attempt 01
failed because the new NDJSON trace schema declared identity only inside its
header branch. The public schema policy requires a closed top-level schema ID.
Every record now carries the same versioned schema ID with closed variant fields;
this remains the initial unpublished cutover. Failed root
`instrumentation-recovery-measurement-shape-01` is preserved.

### S-020/B functional checkpoint — diagnostic collection admitted

New aggregate trace v1, qualification manifest/result v2 and bounded diagnostic
plan/manifest/detail v1 now pass generation, evidence contract (109 tests), native
measurement row, command-surface contract, generation drift, artifact policy and
JSON shapes. Full command/exit records and every failed gate remain in
`instrumentation-recovery-measurement-validation-01`. All acquisition errors stay
unavailable; no RSS or CPU gate was loosened. CPU support points and allowances
are distinct; trace publication cost is separately disclosed. Qualification and
diagnostic readers share one retained-attempt validation boundary.

Diagnostic freeze `9450628e371468ea933e33cdf9688a648c1a3f08` (tree `8918f5d33506eb74257ea0e4f1b8ba3b00a605c8`), branch
`codex/instrumentation-diagnostic-17`, uses the existing isolated checkout. The alternate index leaves the
primary index and staged tracker unchanged. The complete freeze record is
`instrumentation-recovery-measurement-validation-01/freeze-diagnostic-17.json`.
This is not an S-021 qualification candidate or a completion claim.

The next fixed experiment is support-interference: Go and shell, off/zero-reader,
full versus accounting-only support, one warm-up each and six balanced measured
pairs per workload. Expected signature is a consistent support CPU difference
and/or elapsed effect larger than paired variation; absent that evidence, do not
attribute failed observer gates to measurement interference. Every command and
aggregate trace is retained. Accounting-only has no RSS eligibility claim.
B remains IN_PROGRESS until calibration results and limitations are checkpointed.

Isolated freeze-17 JSON-shape validation failed (Make 2, normalized harness
`child_target_failure`): three task declarations require
`deploy/mvp/runtime-roots/tmp/.keep`, but the broad `tmp/` ignore rule excludes
that existing authored marker from every clean checkout. Primary checks had
passed using the ignored local marker. A narrow ignore exception now retains
only `.keep`, preserving exclusion of runtime contents. This is required clean
checkout reproducibility; no performance experiment ran on freeze 17. Its exact
failed root is the isolated `instrumentation-recovery-freeze-17-shape-01`; primary
logs/exit are retained in `instrumentation-recovery-measurement-validation-01`.
Before the first real trace publication, accounting-only support RSS is explicitly
nullable instead of displaying a zero where the protocol makes no RSS observation.
A replacement diagnostic freeze will include these corrections.

B replacement diagnostic freeze: `21aaf19a79d9e63b6d07357254669ca3a0a61d13`, source digest
`sha256:54fea13d21a7dc74a2e5b56a6a340b8a1b1e9f315505183df745f5b240e27420`. Primary generation/evidence/shape/drift and isolated
`instrumentation-recovery-freeze-18-shape-01` pass. Freeze-17 failure is retained.
The exact support-interference input is
`.cartulary/test-results/instrumentation-recovery-support-18-01-command/plan.json`.
Command: `CARTULARY_TEST_RUN_ID=instrumentation-recovery-support-18-01 make harness-instrumentation-diagnose DIAGNOSTIC_PLAN=/home/jochi/code/cartulary/.cartulary/test-results/instrumentation-recovery-support-18-01-command/plan.json`
in the isolated checkout. Result parent is its ordinary `.cartulary/test-results`;
no RAM-disk/environment switch is applied within or between treatments. All
preparation is the fixed off-cache public workload plus protocol warm-ups.
No qualification matrix is admitted.

### S-020/B support-interference checkpoint — calibration gap remains

Frozen diagnostic 18 completed all 28 commands, all workload exits 0, outer Make
exit 0, elapsed 355130.031 ms. The shared secure loader closed every raw trace and
reconstructed every reported aggregate. Isolated root:
`.cartulary/test-results/instrumentation-recovery-support-18-01/diagnostic/`.
Primary command, plan, exact source identity and exit are retained in
`instrumentation-recovery-support-18-01-command`. This is diagnostic completion,
not qualification or S-020 completion.

Full support adds 6.707% (Go) / 13.156% (shell) of one core relative to
accounting-only support, with paired MAD 0.227% / 0.507% and maximum CPU error
allowance 1.740% / 0.183%. This establishes substantial sampling-support cost.
It does **not** establish substantial wall-time interference: paired outer medians
are 0.229% / 0.091%, smaller than their paired MADs 0.673% / 0.315%. Workload CPU
mean changes are 0.935% / 0.461%, with mixed signs. No support estimate is subtracted
from an observer budget, and the original failed qualifications remain failed.

The new aggregate evidence also exposes unavailable RSS observations within the
central 80% window, not only at startup/exit. Almost every full-support attempt is
RSS-ineligible. B calibration therefore remains open and blocks C, E and S-021.
The next bounded diagnostic will distinguish resident acquisition failure stages
before any repair; do not hide these samples, replace them with zero, or relax the
coverage/budget gates. Approximately 5.92 minutes of experimental collection have
been spent against the 90-minute unexplained-cause stop rule.

B next experiment is bounded attribution on Go/off/zero readers (two independent
full-support commands, warm-up each plus six balanced pairs), with the already
adopted diagnostic-only process journal. Before its first publication, the closed
journal vocabulary records only acquisition stage (`process_open`, identity
before/after, page read/format, unknown), never error strings or paths. Hypothesis:
unavailable samples are caused by lifecycle transitions during acquisition;
pre-read versus post-read stages select the responsible ownership boundary.
If failures instead indicate malformed pages/unsupported evidence, repair or
report that limitation before controls. This experiment cannot qualify or replace
any qualification cell. Aggregate trace v1 and qualification v2 shapes are unchanged.

Resident-stage attribution is frozen at `b7b1d8ad8921dfba44b394360f4648c921101436`, source
`sha256:63fec3d6c072f81f9a384d669c20c3af02c2f2bdbe6ff9a7f433a0cfcc57c1c5`; generation, formatting, native owner row, evidence
contract and primary/isolated shape checks pass. Exact validation commands and
freeze record are in `instrumentation-recovery-resident-validation-01`. The typed
plan is `instrumentation-recovery-resident-19-01-command/plan.json`.
Command: `CARTULARY_TEST_RUN_ID=instrumentation-recovery-resident-19-01 make harness-instrumentation-diagnose DIAGNOSTIC_PLAN=/home/jochi/code/cartulary/.cartulary/test-results/instrumentation-recovery-resident-19-01-command/plan.json`
in the isolated checkout, same ordinary result storage. The process journal is
bounded diagnostic evidence; sampled self-CPU is never added to wait4 totals.
All prior evidence remains in place. B and qualification gates stay open.

### S-020/B resident attribution checkpoint — acquisition repair selected

Diagnostic 19 completed 14/14 public commands (all exits 0); diagnostic acceptance
failed because 12/14 attempts had incomplete RSS, so outer Make exited 2 (the
diagnostic CLI's ineligible result is not a product failure). Exact isolated root
`instrumentation-recovery-resident-19-01/diagnostic` and primary command logs are
preserved. Native aggregate/detail digests and every retained attempt validated.
There were 24 process-open losses, seven page-read losses and one pre-read identity
loss; no loss occurred after a successful page read. Thus the proposed explanation
that post-read identity checks reject otherwise valid pages is not supported.

Selected B repair: the meter currently enumerates every host process, then resolves
ownership and later opens resident pages. Replace this sampling path with bounded
traversal of exact owned process trees, reading each resident observation next to
its discovery. Enumerate children of every thread; verify parent/start ownership,
keep the subreaper and gate scopes disjoint, and retain wait4 as the only total CPU
owner. A transient proved identity lost during actual acquisition stays unavailable.
Do not substitute a zero, discard a sample, change the RSS estimand/cadence or
change budgets. This also targets demonstrated sampling-support cost, but its
wall-time impact remains unproven. Cleanup retains its independent ownership proof.

Prediction for the fixed follow-up attribution protocol: eliminate the stale
process-open pattern and reduce sweep spans/support CPU; page-read loss is still
an explicit rejection if it persists. Require independent threaded-child,
reparenting, replacement, scope separation and bounds fixtures before re-freezing.
Experimental collection spent so far is approximately 6.59 minutes; C and all
qualification/handoff completion gates remain blocked on B calibration.

Owned-tree sampling functional checks passed in
`instrumentation-recovery-owned-census-validation-01`: generation, Go formatting,
native measurement row, evidence contract, drift and JSON shapes. Discovery reads
all thread child lists, validates parent/start ownership, closes over adoption,
rejects foreign relationships and exhausts a bounded roster. The actual resident
acquisition and strict availability rules are unchanged. A final scope correction
uses the already-nullable support RSS field when support pages are unavailable;
it does not invalidate an independently complete workload RSS sample. Per-thread
child-list reads are capped at 64 KiB. CPU remains solely wait4 accounting.

Owned-census follow-up uses freeze `49a3b266b8fcab46a631af5b91e56c0c061f8565`, source
`sha256:1eb542828e0adeec326128d825e6101eb3fd745d60e5bcc35001706616bba569`. All three native validation attempts passed, including
replacement of a parent during child discovery; no foreign relationship grants
sampling/cleanup ownership. The same 14-command Go attribution protocol and
fixed preparation will test the declared repair signature. Exact plan/command
record: `instrumentation-recovery-owned-census-20-01-command`.
Command: `CARTULARY_TEST_RUN_ID=instrumentation-recovery-owned-census-20-01 make harness-instrumentation-diagnose DIAGNOSTIC_PLAN=/home/jochi/code/cartulary/.cartulary/test-results/instrumentation-recovery-owned-census-20-01-command/plan.json`
in the isolated checkout. Freeze 19 and its failed acceptance remain unchanged.

### S-020/B owned-census checkpoint — partial repair confirmed

Freeze 20 completed all 14 workload commands successfully but failed diagnostic
acceptance (Make 2): five attempts remained RSS-ineligible. All 24 process-open
losses from freeze 19 disappeared; remaining losses were six page reads and one
post-read identity check, all final observations of those opaque identities.
Median sweep span changed from 6.113 to 5.326 ms; median reported support CPU from
12.077% to 11.208% of one core. These cross-freeze medians are descriptive, not a
randomized performance claim. The failure/root remains
`instrumentation-recovery-owned-census-20-01/diagnostic`; no C or S-021 admission.

The next correction follows independent kernel and arithmetic evidence. The
running kernel is `6.6.114.1-microsoft-standard-WSL2`. Its
[proc stat implementation](https://raw.githubusercontent.com/microsoft/WSL2-Linux-Kernel/linux-msft-wsl-6.6.114.1/fs/proc/array.c)
reports virtual size and resident pages as zero when no address space exists,
independently of the earlier task-state snapshot. A verified empty address space
is different from an inaccessible/disappeared process. Preserve a valid resident
read across a later same-identity zombie state; accept zero only with explicit
empty-address-space proof; missing/replaced identities still fail observation.

Independent review also found that the new implementation rejected central-80%
RSS estimates for missing observations entirely outside that adopted window.
REQ-850 clarifies coverage by the actual estimand, retaining every outside sample
and rejecting any unavailable held interval/uncertainty intersecting the window.
Peak remains the maximum valid sample, never a physical unsampled guarantee.
No numerical gate, cadence, workload or preparation rule changes. Independent
empty-state and window-boundary fixtures must precede another fixed diagnostic.
Approximately 7.26 experimental minutes have been collected.

Empty-address-space and fixed-window fixtures pass with evidence/native contracts
(`instrumentation-recovery-window-validation-01/commands.json`). Diagnostic freeze
`5da4918c7a65676e32050febe6c5715124efb88a`, source `sha256:59ce7407087562b8ca7b8aa4704dd2e4d79880c52442e22be74c5f544a90ad6d`, passes isolated shape checks.
The unchanged 14-command Go/off/zero-reader attribution protocol is fixed in
`instrumentation-recovery-window-21-01-command/plan.json`. Hypothesis: explicit
empty-address-space proof removes exit page-read losses; window-correct coverage
retains valid steady estimates while all remaining central losses fail closed.
Any remaining acquisition failure or unexplained control failure blocks C.
Command in the isolated checkout: `CARTULARY_TEST_RUN_ID=instrumentation-recovery-window-21-01 make harness-instrumentation-diagnose DIAGNOSTIC_PLAN=/home/jochi/code/cartulary/.cartulary/test-results/instrumentation-recovery-window-21-01-command/plan.json`.
No earlier failure is relabeled, and all numeric budgets remain unchanged.

### S-020/B window correction checkpoint — complete Go measurement

Freeze 21 completed all 14 commands with exit 0 and complete reconstructed RSS;
the diagnostic passed with six stable independent pairs. Canonical p90 2.5974%,
outer p90 1.5031%, CPU mean -0.4897% and RSS upper 9,865,909 bytes. Signed CPU is
retained as an observation, not an improvement claim. All outside samples remain
in the traces; acquisition-loss counts are {}. Exact retained root is
`instrumentation-recovery-window-21-01/diagnostic`, Make exit 0, 39,012.912 ms.
This supports the measured repair signature, not qualification. Experimental
collection totals approximately 7.91 minutes. Before C, add independent live
touched-mapping calibration and enforce the planned 28-command sentinel decision
boundary in REQ-851 and the diagnostic runner. The controlled support experiment
found substantial support CPU but no wall effect beyond observed variation; it
does not justify another support rewrite or subtracting its cost from budgets.

### S-020/B completion checkpoint — C admitted

Independent live resident calibration passed: two overlapping 32 MiB touched
anonymous mappings increase residence separately; releasing each decreases
residence separately. Existing wait/adoption, threaded-child, replacement,
missing-sample, cancellation, bounds and arithmetic fixtures pass. Commands in
`instrumentation-recovery-calibration-validation-01/commands.json`: `make format-go`,
`make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.instrumentation_measurement`,
`make harness-evidence-contract`, all exit 0. `git diff --check` also passes.

B adopts exact owned-tree sampling, anchored resident acquisition, explicit empty
address-space proof, separate support availability, reconstructed fixed-window
RSS and per-identity transition uncertainty. CPU remains non-overlapping wait4
accounting. Go attribution 21 has zero acquisition losses and every pair passes.
Support experiment 18 established CPU cost but not significant wall interference;
no noise subtraction or further support rewrite is selected. Old failures retain
their old identities and dispositions. Plan v3, result/manifest v2 and raw/detail
v1 migration boundaries are unchanged by these last corrections. Physical limits
remain sampled WSL2 Linux residence/CPU; no host/daemon/unsampled-peak claim.

C hypothesis: calibrated independent off/zero-reader pairs satisfy the fixed
rules. Its typed protocol is 28 Go/shell commands followed, only on both passing,
by 42 Vitest/service/browser commands. One warm-up per label and six balanced
pairs per workload; cache off, ordinary storage, unchanged commands/cadence.
Failure selects attribution or a feasibility decision, never a full matrix.
The runner now enforces the sentinel decision boundary. A new clean freeze binds
that dispatch change and live-calibration test before collection. S-020 overall
remains IN_PROGRESS, D/E and S-021/S-022 remain unadmitted.

C frozen source: `1517ac72da4ebca0d56cfe91d52c10c686ef6333`, digest
`sha256:0bf3f006de8cb426881d223a70c750cf3577d2f812eb4215c6234bc26439dcea`; isolated shape passes.
Exact typed input and command logs: `instrumentation-recovery-controls-22-01-command`.
Command in the isolated checkout: `CARTULARY_TEST_RUN_ID=instrumentation-recovery-controls-22-01 make harness-instrumentation-diagnose DIAGNOSTIC_PLAN=/home/jochi/code/cartulary/.cartulary/test-results/instrumentation-recovery-controls-22-01-command/plan.json`.
The sentinel stop rule is active. No prior control or failed cohort substitutes
for these independent observations.

### S-020/C sentinel checkpoint — RSS stability blocks progression

Freeze 22 completed 28/28 Go/shell commands successfully with complete validated
measurements, then stopped before Vitest/service/browser because shell stability
failed. Make exit 2, elapsed 350,099.149 ms. Exact root:
`instrumentation-recovery-controls-22-01/diagnostic`; command/plan and independent
leave-one-out arithmetic are in `instrumentation-recovery-controls-22-01-command`.
Go passes every numeric/stability rule: CPU mean 0.8125%, RSS upper 9,821,618 bytes.
Shell full-window gates pass (CPU mean -0.0767%, RSS upper 21,989,547 bytes), but
omitting blocks 1, 2 or 3 yields RSS uppers 37,992,576 / 35,835,194 / 35,739,092
bytes, above 33,554,432. The reader and observer are both off; this is demonstrated
ordinary-workload memory variation, not established observer cost. No valid
observation is excluded. The three remaining controls are uncollected, not passed.
C remains failed and E/S-021/S-022 remain blocked. Experimental collection totals
approximately 13.75 minutes.

D attribution hypothesis: retained Node compilation/runner residency or process
overlap explains the shell RSS dispersion. On the same frozen source, collect
exactly one warm-up per label plus six independent shell/off/zero-reader pairs
with the bounded opaque process journal. Reconcile each process contribution and
lifetime to the raw aggregate central-window integral; sampled CPU remains a
lower bound, never added to wait4. Select a structural lifetime repair only if
that evidence identifies the responsible boundary; otherwise record uncertainty
and reassess rather than launch qualification. Fixed input/logs:
`instrumentation-recovery-shell-attribution-22-01-command`.
Command in the isolated checkout: `CARTULARY_TEST_RUN_ID=instrumentation-recovery-shell-attribution-22-01 make harness-instrumentation-diagnose DIAGNOSTIC_PLAN=/home/jochi/code/cartulary/.cartulary/test-results/instrumentation-recovery-shell-attribution-22-01-command/plan.json`.

### S-020/D attribution checkpoint — B reopened for snapshot ownership

Shell attribution 22 completed all 14 commands successfully, but two attempts
were RSS-ineligible. Make exit 2, elapsed 315,031.108 ms; exact root
`instrumentation-recovery-shell-attribution-22-01/diagnostic`. Three lost observations
(all `after_identity`) occurred after a resident/empty snapshot was obtained; two
intersected steady windows. The first independent reconstruction script stopped
on a null integral, correctly refusing to treat the missing value as zero; its
partial exploratory output is not acceptance evidence. Experimental collection is
approximately 19.00 minutes. The B completion above is superseded, not erased.

Kernel-source inspection of WSL 6.6.114.1
[proc inode ownership](https://raw.githubusercontent.com/microsoft/WSL2-Linux-Kernel/linux-msft-wsl-6.6.114.1/fs/proc/base.c) and
[resident snapshot acquisition](https://raw.githubusercontent.com/microsoft/WSL2-Linux-Kernel/linux-msft-wsl-6.6.114.1/fs/proc/task_mmu.c)
establishes that the held proc directory binds the kernel PID instance. Relative
`openat` reads cannot retarget a recycled numeric PID. A completed snapshot needs
no subsequent liveness check. REQ-850 now states that descriptor ownership rule;
identity still validates before acquisition, and missing/failed acquisition stays
unavailable. Implement bounded, closed-name relative reads and independently
exercise a real child that exits after its snapshot bytes are acquired. This
repair neither fills missing data nor changes the RSS estimand/budgets. Retain
the stage enum for already-published detail v1; no schema acceptance change.

Exploratory per-process sums exactly reconciled available samples. The outer
graph retained roughly 276–281 MiB, but variation also appears in row/test
processes. No graph-lifetime rewrite is justified from size alone. Pinned Node
uses `MainThread` in proc comm, so current closed `other` tags remain explicitly
unclassified; parent topology provides context, not a hidden relabeling.

Descriptor-bound acquisition tests, live exit-after-snapshot calibration, evidence
contract and isolated JSON shapes pass. Validation commands/results are retained
in `instrumentation-recovery-descriptor-validation-01`. New freeze
`f77abf262a5803860a68eb206b5af4176b174ab9`, source
`sha256:06831172e645e1898560ed3e48fe4aa53f75a5e8ffa5ee8ae949c5d805298da8`.
Repeat the exact 14-command shell attribution protocol once as the changed
variant. Prediction: no post-acquisition losses; complete RSS reconstruction.
This corrects measurement loss, and does not assert that C RSS dispersion is fixed.
Expected signature/count/preparation remain in the typed plan; acceptance failures
stay retained and block qualification. Command in the isolated checkout:
`CARTULARY_TEST_RUN_ID=instrumentation-recovery-shell-descriptor-23-01 make harness-instrumentation-diagnose DIAGNOSTIC_PLAN=/home/jochi/code/cartulary/.cartulary/test-results/instrumentation-recovery-shell-descriptor-23-01-command/plan.json`.

### S-020/B descriptor completion and D lifetime repair entry

Freeze 23 shell attribution completed 14/14 commands with no lost observations;
Make exit 0, elapsed 316,354.135 ms. Every detail sample summed exactly to its
aggregate, and independent integration reconstructed every steady RSS value
(`instrumentation-recovery-shell-descriptor-23-01-command/independent-reconciliation.json`).
Shell numerical/stability assessment passes: CPU mean -0.0997%, canonical p90
0.9910%, outer p90 0.9043%, RSS upper 3,171,782 bytes. Signed RSS pairs still span
-13.14 to +14.05 MiB; the negative center is not evidence of an improvement.
B closes again on descriptor ownership, independent fixtures and complete real
Go/shell acquisition. C historical RSS stability failure remains open; this
attribution run is not a substitute control or qualification block. Approximately
24.27 experimental minutes have been collected.

The measured waiting row-runner depth contributes -10.53 MiB of block 5
(-13.14 MiB total), -7.15 MiB of block 6 (-8.39 MiB total), and +6.54 MiB in
block 1 (offset by nested-test variation). The outer graph contribution is much
smaller; no graph/compiler rewrite is selected. Source inspection confirms that
private row preparation allocates complete typed metadata/validators, returns
selected rows, then keeps the preparing Node isolate alive while its adapter runs.
Unreachable heap need not immediately leave RSS; this is a lifetime boundary,
not proof of a reachable-object leak.

D hypothesis: one-shot worker isolation for private row preparation releases that
preparation heap before adapter execution and reduces both its persistent RSS and
its contribution to paired variation. Keep exactly the same metadata/source/path
checks, private selection semantics, workload commands and canonical boundaries.
Return immutable selected facts only; await successful worker exit before launch;
all worker CPU/memory remains part of the containing process, counted once.
REQ-016 adopts the boundary. No forced GC, cache reuse, changed workload, validation
bypass or new daemon. Test result parity and complete rejection before a new
frozen shell attribution variant. Require a residency reduction larger than the
observed row-process variation, then rerun independent C controls; otherwise reject
the repair hypothesis and save a decision checkpoint.

D row-preparation command/evidence contracts and local-migration owner row pass.
The first JSON-shape check fails because generated topology input closure is stale
for the two new private worker modules and changed row runner. Exact failed root:
`instrumentation-recovery-row-preparation-shape-01` (Make 2). This is related
projection drift; regenerate from authored inputs before freezing. The failed
attempt and unit log remain retained. No diagnostic variant has started.

D row-worker functional/projection gates pass after the retained shape failure:
command-surface and evidence contracts, local-migration and native-measurement
owner rows, generation, drift, generated policy, primary/isolated JSON shapes.
Exact commands/results: `instrumentation-recovery-row-preparation-validation-01`.
Frozen variant `7a5b7153c5c19f5f2e2caec9f7a42f2d8731cd98`, digest
`sha256:47b8c069a48a1b8b47715931181819cec71492b57e855c55f59d31a95f6b990a`.
Compare one fixed 14-command shell attribution variant against frozen baseline 23.
The predeclared repair signature is at least 12 MiB lower median time-weighted
row-runner contribution (larger than its observed 10.53 MiB paired variation),
with lower row contribution dispersion and no acquisition/CPU/elapsed regression.
Cross-source reduction is diagnostic attribution, never qualification. Preserve
all 14 attempts; failure rejects this variant rather than admitting another
unchanged retry. Command in the isolated checkout:
`CARTULARY_TEST_RUN_ID=instrumentation-recovery-shell-worker-24-01 make harness-instrumentation-diagnose DIAGNOSTIC_PLAN=/home/jochi/code/cartulary/.cartulary/test-results/instrumentation-recovery-shell-worker-24-01-command/plan.json`.

### S-020/D row-worker checkpoint — hypothesis rejected and rolled back

Frozen variant 24 completed 14/14 commands with complete, independently reconciled
measurements, but failed shell stability (Make 2, 314,161.802 ms). Every full-window
numeric gate passes; RSS upper is 21,079,093 bytes and CPU mean 0.0721%. The
predeclared residency signature fails: row-depth median increased from 95.090 to
96.503 MiB, rather than decreasing by at least 12 MiB. Row MAD changed from 2.035
to 1.662 MiB, insufficient to establish the required structural benefit. Outer
graph median stayed about 280 MiB; nested-test memory variation remains material.
The earlier warm-up drop was not representative of steady contribution.

Exact retained root: `instrumentation-recovery-shell-worker-24-01/diagnostic`.
Independent sample/integral closure and lifetime comparison are retained in
`instrumentation-recovery-shell-worker-24-01-command`. All results and source
commit `7a5b7153c5c19f5f2e2caec9f7a42f2d8731cd98` remain preserved. The two new
worker modules, runner/test changes, REQ-016 clause and topology render-index
projection are restored together to diagnostic-23 content, without touching the
primary index or reverting validated measurement changes. No optional worker
mechanism remains. Experimental collection totals approximately 29.51 minutes.

D has not resolved C RSS dispersion; E, S-021 and S-022 remain blocked. A compiler
or parser rewrite cannot be justified by aggregate process size, and unchanged
control/qualification retries are not authorized by this failure.

### S-020 decision checkpoint — retained repairs and unresolved feasibility

The worker rollback is content-exact: an alternate-index tree comparison matches
`f77abf262a5803860a68eb206b5af4176b174ab9` / tree
`2259f23fb68e75d47704a2b89580171e7b26751e` across authored/generated files, excluding
only this controlling tracker. The isolated checkout is clean on
`codex/instrumentation-diagnostic-23`. Rejected commit
`7a5b7153c5c19f5f2e2caec9f7a42f2d8731cd98` remains on its diagnostic branch.
The primary index was not used for either freeze or rollback; staged tracker
SHA-256 remains `9ab829e44fa8df7ba1ee86e794518acd97129cea6af416966fbdda4b33f83426`.
Exact proof: `instrumentation-recovery-shell-worker-24-01-command/rollback-verification.json`.

Retained changes include plan v3's strict CPU mean <2.5%, result/manifest v2 and
bounded aggregate trace/detail contracts, shared immutable evidence validation,
closed diagnostic dispatch, explicit support accounting, owned-tree discovery,
descriptor-bound resident snapshots, fixed-window coverage/transition uncertainty,
independent live allocation/exit calibration and the sentinel stop gate. No
row-preparation worker, additional profiler daemon, forced GC, cadence/workload
change, RAM-disk switch or relaxed budget remains. Raw qualification evidence is
required on future attempts; process detail remains diagnostic-only.

The following new evidence is not interchangeable:

- Go attribution 21: complete acquisition and stable six-pair numerical pass.
- Controls 22: Go passes; shell fails RSS stability with complete acquisition;
  Vitest/service/browser were not collected because the sentinel gate stopped.
- Shell attribution 22: two ineligible attempts exposed post-acquisition liveness
  rejection, which reopened B and selected the descriptor repair.
- Shell descriptor 23: complete acquisition and stable numerical pass; signed
  memory differences remain substantial, so no improvement is inferred.
- Shell worker 24: complete acquisition but failed RSS stability and failed the
  predeclared >=12 MiB residency reduction; the entire variant was reverted.

All experiment result roots are under
`/home/jochi/.codex/worktrees/instrumentation-qualification/cartulary/.cartulary/test-results/`;
primary `*-command` directories hold typed inputs, exact invocations, exits,
timings and independent reconstructions. Prior result roots, including candidate
16 and earlier failed cohorts, were neither rewritten nor deleted.

Verification of the retained descriptor revision includes `make format-go`,
`make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.instrumentation_measurement`,
`make harness-evidence-contract`, isolated `make json-shape-check`, and the fixed
shell diagnostic, all passing at their recorded source. Earlier measurement
projection gates and command-surface checks retain their exact source identities.
The reverted worker had passing functional/projection checks but failed the
experiment; those passes do not qualify the retained source. The failed initial
worker shape check remains recorded as related generated-input drift.

No remaining attributed owner defect presently explains shell RSS dispersion.
The row-isolate experiment specifically falsified the predicted resident release;
process size alone does not justify rewriting graph compilation, test execution
or schema validation. Before another implementation experiment, establish whether
the waiting process's native allocation retention or later execution setup owns
that residence, and whether it explains the *paired variation*, not merely total
size. This is a new hypothesis requiring a fixed diagnostic contract and evidence,
not permission to substitute process isolation, retry controls until lucky, or
borrow attribution pairs as qualification. The current WSL2 setup has not yet
demonstrated feasibility under all fixed rules.

S-020/E, the next S-021 matrix, `make harness-performance-check` on a new qualifying
manifest, and S-022 finalization/broad verification remain skipped because their
prerequisite gates are unmet. `make agent-finalize` was not run; `RESULTS_DIR`
was not supplied and no retained-run maintenance was performed. S-022's tracker
Markdown coverage, final lint/import review and final handoff remain uncompleted.
Before a future qualification freeze, account for the executable identity effect
of the Markdown-lint configuration change so a later documentation checkpoint
cannot silently relabel evidence. F-022/RB-006, T-007/T-008 and S-020–S-022 remain
open. This is a saved decision checkpoint, not the S-022 completion checkpoint.


### S-020 memory-lifetime resumption — IN_PROGRESS

The user authorized the four-workstream recovery plan. Baseline is retained
revision `f77abf262a5803860a68eb206b5af4176b174ab9`; the isolated qualification
checkout is clean. Primary staged tracker SHA-256 remains
`9ab829e44fa8df7ba1ee86e794518acd97129cea6af416966fbdda4b33f83426`.
Content baseline is retained at
`.cartulary/test-results/instrumentation-recovery-lifetime-entry-01/baseline.json`.

Workstream 1 implements diagnostic plan/manifest v2 and a bounded lifetime v1
sidecar: shell off/zero readers, one warm-up per label and six balanced measured
pairs; A retains external process detail, B adds numeric boundary/GC probes.
The hypothesis is that a named allocation lifetime or changed phase overlap
explains paired central-window RSS variation. Raw samples and phase contributions
must reconcile; a materially perturbing probe cannot establish the cause.
No repair is selected yet. A source-owned cause and its independent reproducer
are prerequisites for workstream 2; controls, qualification and finalization
remain gated. CPU mean remains strictly below 2.5%; every other budget remains
unchanged. Experimental collection starts from approximately 29.51 minutes
against the existing cumulative 90-minute stop ceiling. Failed evidence and
the rejected worker revision remain immutable.

Workstream 1 implementation verification: `instrumentation-recovery-lifetime-generate-01`
passed `make generate` (10,340 ms); `instrumentation-recovery-lifetime-format-01`
passed `make format-go`. The native measurement row failed in
`instrumentation-recovery-lifetime-native-01` (Make 2, product assertion, 28,482 ms).
Its new live-transport test found a missing default case ordinal in the direct
probe API; all retained measurement tests passed. Cleanup passed 3/3. The failed
root and `make explain-run ... TARGET=test-slice DETAIL=logs` diagnosis remain
retained. The probe now supplies ordinal zero, and the peer handshake additionally
binds a fresh challenge to the still-live proven process instance. No experiment
or qualification was collected; the cumulative collection clock is unchanged.

The corrected native row passed as `instrumentation-recovery-lifetime-native-02`
(28,167 ms). Generation passed again as `instrumentation-recovery-lifetime-generate-02`
(9,950 ms). The sequential public checks in
`instrumentation-recovery-lifetime-validation-01/commands.json` passed:
command-surface contract 56,125.789 ms, evidence contract 19,956.449 ms and JSON
shape 4,676.079 ms. New lifetime regressions are routed outside the frozen shell
workload, under command-surface contracts and the measurement Go row. Before the
first lifetime publication, schema bindings are tightened to the exact aggregate
trace and external process-detail digests; the targeted reader cases must pass
again after regeneration. No qualification evidence is claimed.

### S-020 memory-lifetime acquisition freeze — IN_PROGRESS

Lifetime infrastructure and its independent functional gates pass; causal
attribution remains unproven. `instrumentation-recovery-lifetime-validation-02`
records command-surface 52,950.869 ms, shape 4,660.318 ms, generate-drift
19,036.210 ms and generated-artifact policy 2,124.504 ms, all passing. Generation
03 passed in 9,700 ms. The isolated shape check passed 3/3 in 3,458 ms.

Frozen diagnostic revision: `3fad783d93551fc34240af43d66004d7f862e731`, source
`sha256:0a3d4e29e58cbccfd980d464e016cda4de839b5c1c3d90dfccc2150b112f40cb`, tree
`8bebe449a3a63680ca72ffb2d4d81ce36a9b199e`. Its branch is
`codex/instrumentation-diagnostic-25`. The primary index and staged tracker digest
are preserved; freeze proof is in `instrumentation-recovery-lifetime-freeze-25`.
Generated validator changes include deterministic function renumbering after
registering the new schema, not acceptance changes to existing schema versions.

The predeclared next experiment is exactly 14 shell commands: one warm-up and six
measured pairs, balanced A/B order, A external detail and B external detail plus
phase/GC probes, both off/zero readers. Inputs and exact native command are in
`instrumentation-recovery-memory-lifetime-25-01-command`. Result root is isolated
`.cartulary/test-results/instrumentation-recovery-memory-lifetime-25-01/diagnostic`.
Preparation/storage match the retained recovery experiments; no RAM-disk or
environment change. Probe perturbation, uncertainty and incomplete acquisition
are assessed before causal conclusions. No functional repair or qualification is
claimed; workstreams 2–4 remain gated. Cumulative prior collection is 29.51 minutes.

### S-020 memory-lifetime baseline 25 — acquisition complete, attribution IN_PROGRESS

The fixed 14-command cohort completed successfully at revision 25. Exact command
and independent native exit are retained in
`instrumentation-recovery-memory-lifetime-25-01-command/command.json`; the isolated
diagnostic root is unchanged. All measurements and lifetime sidecars are eligible,
and the probe comparison passes the fixed numerical and leave-one-out rules.
Elapsed collection was 318,211.629 ms; cumulative experimental collection is
approximately **34.814 minutes**, including the prior 29.51 minutes.

Probe CPU paired mean was +0.527391% of one core. RSS conservative upper estimate
was 19,014,377 bytes. All six paired RSS differences were positive (approximately
2.4–18.6 MiB), so passing gates do not establish absence of probe perturbation.
Independent sample/integral and phase reconstruction is still required. Preliminary
observations show varying row-runner heap use during its child wait and repeated
large catalog allocations in the nested contract tests. These are hypotheses, not
a causal disposition. Intervals after a case-end marker can contain other
uninstrumented test work and must not be labeled as idle or attributed to that
completed case. No repair, control rerun, qualification or handoff is claimed.

### S-020 workstream 1 checkpoint — bounded attribution complete

Independent reconstruction is saved with its Python source in
`instrumentation-recovery-memory-lifetime-25-01-command/independent-reconciliation.json`.
All 14 trace/detail sample sums, reference digests, clock placements and central-80%
integrals reconcile (sub-byte arithmetic residual). Authenticated B producer roles
confirm the three largest processes' ancestry mapping used for A. Clock placement
uncertainty remains bounded by the adopted 25 ms limit. Heap observations are not
added to RSS. All paired process contributions sum to their aggregate differences.

The row runner occupies the entire steady window in every measured attempt. Its
A RSS spans 92.37–103.34 MiB and B spans 94.74–103.24 MiB. Natural major GC occurs
at different positions within preparation; retained heap/capacity differs during
the long wait. This rules out changed row-process overlap as the explanation, but
does not identify a specific unnecessary row allocation. The rejected Worker
variant remains rejected; no row-runtime rewrite is authorized by this evidence.

A distinct source-owned mechanism is demonstrated in the nested fixture owner:
PostgreSQL catalog closure and four row-metadata partition/coverage cases request
`context.catalog`, which resolves all Go and TypeScript source selectors although
those five cases assert only validated metadata. Each case owns a fresh resolver;
source parsing repeats and discarded syntax-tree allocation expands the process
heap across cases. The five named intervals total 2.84–2.88 s and contribute
69.53–77.11 MiB to the central integral in B. Their changing levels, rather than
material duration change, explain that 7.58 MiB contribution range. Browser graph
compilation and selector-specific cases still require complete source closure.

This identifies a repairable allocation mechanism, not the sole cause of F-022
or proof about old cohorts lacking phase samples. Probe B adds a median 6.64 MiB
aggregate RSS and 0.527391% CPU; its six evidence-process differences are
+0.18–6.22 MiB. Thus uninstrumented controls remain mandatory and absolute probe
readings cannot be substituted for ordinary qualification. The present cohort
has no failing leave-one-out assessment; older failures retain their limitations.

### S-020 workstream 2 admission — one fixture variant

Predeclared repair: give the five metadata-only cases a case-owned validated
metadata dependency. Preserve every existing assertion, current-input validation,
all selector-specific tests, browser compilation, exact shell row and public full
preflight. Do not share contexts across cases or retain a global catalog cache.

Independent reproducer: observe source-file acquisition for full and metadata-only
loads of the same repository, assert identical row/owner/PostgreSQL facts, prove
metadata loading opens no source selectors, and exercise changed/malformed metadata
on a fresh case. This confirms removal of the demonstrated acquisition mechanism
without a memory-size assertion or forced GC. The new regression is routed outside
the frozen shell workload.

Predicted signature: the nested evidence process's median steady contribution
falls by **at least 20 MiB** against revision 25, for both A and B, and the five
metadata-case phases no longer show repeated source-parsing growth. This exceeds
baseline process dispersion (approximately 6 MiB) plus observed paired error
(approximately 3.3 MiB). Reconstruct the same fixed 14-command diagnostic schedule
independently. Reject the variant if the acquisition reproducer fails, the stated
signature is absent, or selection/lifecycle assertions change. A passing aggregate
alone is insufficient. Row-runtime variation is not predicted to improve. Only
a demonstrated repair permits uninstrumented controls; a remaining control failure
requires a new decision checkpoint, not another automatic qualification attempt.

Variant 26 functional checks: command-surface contract passed in 55,624.528 ms
(including the independent source-acquisition regression); evidence contract passed
in 17,928.910 ms. `generate-drift` failed in 19,425.066 ms because the changed catalog-loader
renderer-source digest changes `tools/execution_topology_render_index.json`. The failed root
`instrumentation-recovery-fixture-26-drift-01` and public `explain-run` diagnosis
are preserved; cleanup passed 6/6. This is related generated-input drift, not a
workload assertion failure. Freeze is not committed. Run the owner generator, then
repeat drift/shape/policy gates before acquisition.

Variant 26 repaired projection gates passed: generation 13,021.020 ms, drift
20,059.496 ms, generated-artifact policy 2,160.293 ms and shape 4,659.566 ms.
Frozen commit `10d43252f3169242b3705d481bbbdb1510bff62d`, tree
`07ae78e1824dedb576322b685cc9973bf00fc31a`, source
`sha256:3d1c182f60e6df1e6fb429ad474541eeaec554d70a8a101b408eecab8c931395`.
Branch `codex/instrumentation-diagnostic-26` is clean; isolated shape passes 3/3
(3,527 ms graph duration). Primary staged tracker and index are preserved across
the freeze. Proof is in `instrumentation-recovery-fixture-freeze-26/freeze.json`.

The single predeclared variant now runs the same 14-command lifetime protocol.
Exact command/plan and external exit are retained in
`instrumentation-recovery-memory-lifetime-26-01-command`; isolated result parent is
`instrumentation-recovery-memory-lifetime-26-01`. There are no policy, cadence,
preparation, storage or workload-selection changes. Cumulative collection entering
this experiment is 34.814 minutes. No control or qualification admission yet.

### S-020 workstream 2 checkpoint — fixture repair accepted

Revision 26 completed the one prescribed variant cohort: 14/14 commands eligible,
complete lifetime publications, all fixed numerical and stability checks passing.
Make exited 0; external collection elapsed 286,612.483 ms. Cumulative experimental
collection is approximately **39.590 minutes**. Original revision 25, rejected
worker 24, failed projection checks and all older evidence remain immutable.

Independent reconstruction and the cross-revision comparison are in
`instrumentation-recovery-memory-lifetime-26-01-command/{independent-reconciliation,variant-comparison}.json`,
with the independent Python arithmetic source alongside. All samples/integrals
reconcile. The predicted nested-process median reduction exceeds 20 MiB in both
configurations: **35.015 MiB A, 39.124 MiB B**. The five metadata-only case intervals
shrink from 2.838–2.880 s to 0.393–0.410 s; their central RSS contributions shrink
from 69.53–77.11 MiB to 9.05–9.45 MiB. Source acquisition independently proves that
these cases no longer parse source selectors, while metadata facts, fresh-input
validation, all existing assertions and complete source preflight are preserved.
No forced GC, tuning, process isolation, padding or coverage reduction was used.

The effect exceeds baseline variation and measurement error, so this specific
fixture repair is retained. It is not a claim that all RSS instability is solved.
Row-runner ranges remain 12.53 MiB A and 10.79 MiB B, and graph A has a 10.15 MiB
range. Nested-process A dispersion increased despite the lower median; the B range
is essentially unchanged. The probe comparison CPU mean is -0.040032% (no improvement
claim); RSS upper estimate is 13,507,612 bytes. Only the upcoming independent,
uninstrumented controls can establish readiness for qualification.

### S-020 workstream 3 admission — ordinary controls

Disable lifetime probes through the existing repeatability protocol; use ordinary
100 ms measurement and all unchanged rules. Validate measurement owner, evidence
and command-surface contracts, exact local-migration and shell rows, and projections.
Add the tracker to Markdown-lint coverage before freezing control/qualification
source. Collect six independent pairs plus the two warm-ups per workload, Go then
shell first. The runner must stop after those 28 commands if either sentinel fails;
only both passing admit Vitest, service and browser controls. Every observation,
including a failing or unstable control, remains retained. S-020 is not complete
and S-021/S-022 remain unstarted.

Workstream 3 functional gates pass: measurement row 29,648.375 ms, exact local
migration row 3,272.227 ms, exact shell row 21,217.738 ms and Markdown lint
11,115.195 ms. Commands and outputs are retained in
`instrumentation-recovery-control-validation-27-01`. Earlier revision 26 owner
contracts/projection gates cover the unchanged executable fixture repair.

The tracker is now explicitly included in `.markdownlint-cli2.jsonc`; lint passes
with that coverage. The clean control freeze is commit
`c68c8cfa45e0a3e365bb168008e99c40e6ff5f1f`, source
`sha256:e8a8e516e2070e5fc18683713d4423766c618a09054533cfa7ece701bfe287fd`,
tree `9dcf5acd0ed6caddc638e7f7179b36e76cce4142`. Revision 27 differs from 26 only
in the Markdown-lint configuration; source identities are not relabeled.
Isolated shape passes 3/3 in 3,429 ms. Freeze proof is
`instrumentation-recovery-control-freeze-27/freeze.json`; primary staging remains
preserved.

The ordinary repeatability experiment starts with 39.590 cumulative experimental
minutes. Its typed plan and command are in `instrumentation-recovery-controls-27-01-command`,
with the isolated result parent `instrumentation-recovery-controls-27-01`. Lifetime
probes and external per-process detail are disabled. No full qualification starts
unless all five controls are eligible and stable under the fixed rules.

### S-020 workstream 3 decision checkpoint — controls FAILED, effort IN_PROGRESS

Revision 27's repeatability runner correctly stopped at the 28-command sentinel
boundary. Native Make exit was 2; manifest status is `stopped`, not a complete
five-workload result. External elapsed was 321,715.465 ms, bringing cumulative
experimental collection to **44.952 minutes**. No observations were excluded or
replaced. Vitest, service and browser controls were not admitted. No qualification
cohort was started or retried.

Exact evidence is the isolated
`.cartulary/test-results/instrumentation-recovery-controls-27-01/diagnostic/diagnostic.json`
and `assessment.json`. The primary command directory retains the typed plan,
external command/exit, stdout/stderr and independent Python reconstruction in
`instrumentation-recovery-controls-27-01-command/independent-assessment.json`.
All 28 workload commands exited 0; all retained test summaries and secret scans
passed; every recorded cleanup operation completed. Lifetime/detail references
are null in all 28 attempts, confirming ordinary measurement configuration.

Go is eligible but fails canonical p90: measured block 5 is 319 ms A versus
335 ms B, **+5.015674%** against the unchanged 3% gate. Canonical accounting locates
the 16 ms difference in execution (+12 ms), setup (+3 ms) and unattributed (+1 ms);
this does not establish their physical cause. Canonical median is -0.466325%; outer
median/p90 are -1.174070%/+1.983024%. CPU mean is -1.585374% and RSS upper estimate
is 8,081,817 bytes. Negative signed observations establish no improvement. Omitting
block 5 changes p90 to 2.140673% and removes the failure, so the required leave-one-out
stability gate also fails. No canonical accounting discrepancy was found.

Shell's measured block 4 B has one unavailable RSS sample at 2,612.483666 ms,
covering 99.625061 ms of the central window until the next sample at 2,712.108727 ms.
The aggregate has count 7, sweep span 9.387650 ms, `rss: null`, and availability
`unavailable`. Ordinary mode does not retain individual failure reasons or process
identities, so attributing that event to a particular disappearing process would
be speculation. The validator correctly reports `insufficient_samples`; bounded
RSS error and steady RSS remain unavailable. No zero-fill, interpolation or pair
discard is permitted. All other 27 central integrals independently reconcile.
The shell control is **ineligible**, with no RSS or stability pass claim. Its full
six-pair CPU mean (-0.006982%), canonical median/p90 (+0.358653%/+0.827144%) and
outer median/p90 (+0.357888%/+0.573212%) are descriptive only.

Disposition: retain the demonstrated metadata-fixture repair and bounded lifetime
diagnostic infrastructure. Do not reintroduce the failed Worker variant. F-022 and
RB-006 remain OPEN; S-020 and T-007/T-008 remain IN_PROGRESS; new S-021 collection and S-022 remain
unstarted. S-020's end-to-end exit criterion is unmet despite functional and
projection checks passing. This checkpoint completes the prescribed stop decision,
not S-020 or S-022. Default diagnostics remain off; all evidence is local engineering
evidence with WSL2/daemon/unsampled-memory limitations unchanged.

Before any new collection, choose and predeclare a diagnostic capable of separating
Go preparation/child timing and the cause of unavailable resident observations.
The latter requires bounded failure provenance, not a weaker coverage rule. No
additional variant or automatic control/full-matrix retry is selected by this
checkpoint. Retain the original signed observations and policies. The cumulative
90-minute ceiling has approximately 45.048 experimental minutes remaining.

Skipped due to the failed prerequisite: remaining three controls, full S-021
qualification, qualification-manifest `harness-performance-check`, S-022
`agent-finalize`, broader harness/script/shell verification, final exact-run
explanation/export audit and S-022 completion. `RESULTS_DIR` was not supplied; no
retained-run maintenance was performed. Markdown coverage and its earlier lint
pass are complete, but do not constitute S-022 finalization.

Rollback guidance: the accepted fixture transaction is isolated as revision 26
(on top of diagnostic 25), with revision 27 adding only lint coverage. Preserve
those source identities and all associated evidence. If that transaction later
requires rollback, revert its authored fixture/catalog changes and regenerate
the renderer-input projection together; do not restore the rejected Worker or
translate old evidence. The primary staged tracker digest remains
`9ab829e44fa8df7ba1ee86e794518acd97129cea6af416966fbdda4b33f83426`.

Checkpoint documentation validation: `make lint-markdown` passed with the tracker
in scope as `instrumentation-recovery-control-27-checkpoint-markdown-02` (7,680 ms
tool duration). `git diff --check` passed; the isolated revision-27 checkout is
clean and the staged primary tracker digest is unchanged. Current status tables
now distinguish the accepted fixture repair from the failed ordinary control gate.


### Superseded pre-salvage current summary — historical, 2026-10-11

#### Earlier baseline and authorization

The underlying user request authorizes implementation of S-014–S-022, including
specification amendments, typed projections, implementation, tests, generation,
engineering qualification and finalization. Entry is `main` at
`1d423c6e5f9cd204f8210fa9958c780fd933d293` with this tracker already staged;
the staged patch is preserved at `/tmp/cartulary-instrumentation-staged-entry.patch`
(SHA-256 `dfc9cd94a8b356b2afd78654f5209896ae4423c968b9b6b73e8bc0fc66b4caae`).
The entry tracker hash is `9ab829e44fa8df7ba1ee86e794518acd97129cea6af416966fbdda4b33f83426`.
No staging or commit of pre-existing work is implied. The latest user request authorizes implementation of the proposed recovery plan.
S-020/A–B retain their completed policy/calibration dispositions. C controls remain
failed: revision 27 has a Go canonical-p90 failure and one ineligible shell RSS
observation. D row-worker hypothesis failed and was reverted; the later metadata
fixture repair met its independent acceptance criterion and is retained. Clean
freezes, failed controls and evidence limitations are saved below. The Section 7 sequence governs this execution and blocks qualification
until measurement, controls and functional closure pass.

The accepted implementation plan fixes 25 ms maximum placement uncertainty,
250 ms container batches with 32 live/256 lifetime identities, a 64 KiB/64-graph
terminal receipt, 30-second readers, 100 ms external memory sampling and seed
20261009. The latest user-authorized CPU limit is arithmetic mean total incremental
observer CPU strictly below **2.5% of one core per workload/treatment cell**,
including readers. S-020/A adopted and projected that revision as plan v3. Candidate 16 and earlier
evidence retain their original policies and failures. CPU dispersion/error remain reported; all other
budgets and eligibility/stability requirements remain unchanged. Default remains off. `harness-instrumentation-qualify
QUALIFICATION_PLAN=<file>` owns collection; `harness-performance-check` remains
an independent validator. The execution checkpoints at the end of this file
record the current slice; each is saved before its successor begins.

Current execution: S-014–S-019 remain DONE. S-020/A adopted plan v3 and
S-020/B now has independently calibrated, reconstructible measurement evidence.
C historical controls passed Go but failed shell RSS leave-one-out stability.
Fresh revision 27 controls stopped after the two sentinels: Go failed canonical
p90/stability, and shell was ineligible because one central-window sample was
unavailable. This supersedes the readiness inference, not historical evidence.
D repaired resident acquisition and tested a row-preparation worker hypothesis;
that worker failed its predeclared residency signature and was completely reverted.
The bounded memory-lifetime protocol and its baseline cohort are complete.
Independent reconstruction identifies repeated full source-catalog acquisition in
metadata-only contract fixtures as a source-owned allocation mechanism. The single
fixture variant met its predeclared 20 MiB reduction criterion in
both configurations. The fixture repair is retained; row-runner variation remains
a separate contributor, and F-022 is not closed. Uninstrumented controls failed;
dependent work is stopped at the saved decision
checkpoint below.

The current clean diagnostic commit is
`c68c8cfa45e0a3e365bb168008e99c40e6ff5f1f`, source
`sha256:e8a8e516e2070e5fc18683713d4423766c618a09054533cfa7ece701bfe287fd`.
It is not qualified. Revision 26 demonstrated the fixture repair; revision 27 adds
tracker Markdown coverage and fails the ordinary control gate. Cumulative
experimental collection is approximately 44.952 minutes of the unchanged 90-minute
ceiling. The failed gate, not expiration of that ceiling, stops dependent work.
S-020/E, S-021 and S-022 remain gated. All historical
revisions and failed evidence are preserved; no Core 05 claim is made.

Candidate 16 remains historical rejection evidence: its 240 eligible attempts
had 14 numerically failing cells and five unstable workloads under its original
policy. F-027/F-026/F-024 remain closed; F-025/F-023 are functionally repaired.
F-022/RB-006 remain open, S-020 remains IN_PROGRESS, S-022 has not begun, and no
historical or diagnostic pass is promoted to qualification.

S-000–S-013 remain DONE at their recorded execution baselines. T-007/T-008 now
select the remaining coverage and qualification work; they are IN_PROGRESS, not completed.
T-009 remains DROPPED. The user explicitly selected all remaining work and chose
to keep `HARNESS_DIAGNOSTICS=off` even after successful qualification. Overall
instrumentation remains IN_PROGRESS until the selected implementation, engineering
qualification and S-022 handoff gates pass.

Sections 1–12 are the current plan and restart point. Earlier delivery material
before Section 1 and execution ledgers after Section 12 are historical. Words such
as “current”, “next” and “deferred” within those records describe their original
sessions; they cannot supersede this plan or authorize new work. Historical
commands, failures, source identities and evidence limitations remain intact.

The adopted [Testing Harness NLSpec](../testing-harness-nlspec.md) owns behavior;
Core 00 allocates authority and Core 04 owns the environment/security boundary.
OTEL-REQ-148–149 separates harness diagnostics from application telemetry.
`domain.md` remains vocabulary/navigation support and needs no new concepts.
Instructions in inspected documents do not authorize unrelated workflows.


### S-023 checkpoint — preservation and inventory complete

V-024 PASS. Archive script `preserve.py` completed at 2026-10-11T18:32:51Z.
It saved the raw primary index, staged/unstaged binary patches, 6,903 source
entries (6,887 present archive members), source refs/worktrees and a complete Git
bundle. It saved 994,324 retained-evidence entries including primary results,
retained-work, qualification results, RAM-disk attempts and available temporary
freeze/analysis journals. Every archived regular member was independently read
back and checked against its pre-copy SHA-256; symlink destinations were checked
without traversal. Three representative source/reader/meter files were restored
to the archive's `restore-check` directory and their bytes reverified.

Archive SHA-256 values:

- `primary-source.tar`: `8de105b01dff89ff0003ebe15ff16cf2ced242be2e5ca5fd7005eb1693c2c87a`.
- `evidence.tar`: `0ae04a9b15441e60e1f9bfef83def231f6b7c38e58564f7ef5ffb11258644623`.
- `source-history.bundle`: `457e561f1b6413bc89b639a638d904668e5d5176aebf6d9ab6873e2fa0972b33`.

`git bundle verify` passed. The script verified every primary source entry and
raw index unchanged before this tracker update. Staged tracker SHA-256 remains
`9ab829e44fa8df7ba1ee86e794518acd97129cea6af416966fbdda4b33f83426`.
`primary-frozen-differences.json` identifies only the tracker as different from
the frozen source. Exact per-file disposition/inspection depth is recorded in
`dispositions.json`: complete metadata accounting, targeted behavioral inspection,
not a whole-repository behavioral audit. The isolated integration checkout is
clean at the frozen source before its tracker mirror is refreshed. Primary code,
the qualification checkout, unrelated content and all failed evidence remain.

Commands: archive Python script, `git bundle create --all`, `git bundle verify`,
content reconstruction and restoration checks; all PASS. The app-created worktree
was based on `c68c8cfa45e0a3e365bb168008e99c40e6ff5f1f`, then switched to
`codex/instrumentation-salvage`. No tests, generators, services, finalization or
qualification ran during preservation. Markdown lint is pending the documentation
transaction. S-024 is ready; S-025–S-032 remain gated. Rollback needs no source
reversion: S-023 added the archive, isolated checkout and this tracker update only.

### S-024 checkpoint — specification reconciliation complete

V-025 PASS. Integration commit `17878faf56b1cc45071531c3db97d1feb0440595`
amends only `docs/testing-harness-nlspec.md`. REQ-852–855 / AC-147–149 adopt
Go graph ownership, shared Go fixture policy, bounded private plan/session
interfaces and functional salvage completion. REQ-848 retains the entire
qualification budget/matrix/eligibility/stability policy and independent validator
while retiring the collection command. REQ-850 remains an evidence/validation
contract without requiring an active meter; REQ-851 retires active diagnostic
collection, formats and lifetime probes. REQ-845–847/849 remain unchanged.
The section introduction now names actual qualification plan v3 and manifest/result
v2 rather than its obsolete all-v1 statement. No domain vocabulary changed.

Owner-edit review compared preserved clauses and checked every numerical budget;
this was human specification-edit support, not executable conformance evidence.
`CARTULARY_TEST_RUN_ID=salvage-s024-markdown-01 make lint-markdown` PASS, exit 0,
5,440 ms tool duration. Root: integration checkout
`.cartulary/test-results/salvage-s024-markdown-01`, summary
`adhoc/lint-markdown/tool-run-summary.json`. `git diff --check` PASS.
Make provisioned the isolated pinned frontend runtime/dependencies as required
for Markdown lint. No product tests, generators, services, qualification or
finalization ran. No failed command in this workstream. The primary code and
staged tracker remain unchanged; this checkpoint is mirrored into the integration
checkout. Native runtime/projection changes remain future work.

S-025 is ready. Later workstreams remain gated. Rollback is the single owner
amendment commit; it does not delete the external archive or rewrite evidence.

### S-025 checkpoint — retained foundation complete

V-026 PASS. Integration commit `ff48689f0290818040b9023ab4633944ea3b49e0`
retains the frozen foundation and adds the previously unrouted metadata-lifetime
regression to the evidence contract, with authored routing and regenerated
projections. Frozen source already includes metadata-only catalog access with
fresh selection validation, case-owned fixtures, shared generated validators,
unified immutable retained reading, secure process/file ownership, bounded source
hashing and the separate single-master SeaweedFS repair. No experimental removal
or runtime port is included. The retained-reader fixture checks one resource
traversal per independent read, strict envelope rejection and export isolation;
the foundation boundary route checks validator acceptance parity and bootstrap
cache identity. The metadata fixture checks both full source preflight and fresh
case rejection of malformed metadata.

Environment: isolated Linux/WSL checkout, pinned Node and Go through Make; source
identity `sha256:b11396de2a7b8884bfd3266193c1bb9598770c54a8b905815f5ba1653b5cd3fe`
for the final functional/projection checks. Result roots below are relative to
integration `.cartulary/test-results/`. Each command used the root's name as
`CARTULARY_TEST_RUN_ID`:

- `make task-guide ROLE=module-author OWNER=harness.evidence_accounting`:
  PASS, `salvage-s025-guide-01`.
- `make harness-evidence-contract`: PASS, `salvage-s025-evidence-01`
  (15,750 ms), then after routing correction `salvage-s025-evidence-02`
  (16,340 ms), outer exit 0 for both.
- `make harness-command-surface-contract`: PASS, `salvage-s025-command-01`
  (50,960 ms), exit 0; final harness-contract includes this route again.
- `make generate`: PASS, `salvage-s025-generate-01`, exit 0, 10,500 ms.
- `make harness-contract generate-drift generated-artifact-policy-check json-shape-check`:
  `salvage-s025-foundation-01`; harness-contract passed (107,195 ms), then
  generate-drift preflight rejected the nonempty run-ID reuse; overall exit 2.
  The remaining checks did not run in this attempt. This is a command-shaping
  error, not a projection or product failure; the successful graph summary is
  not represented as a successful outer multi-goal command.
- `make task-guide ROLE=module-author OWNER=harness.browser`: PASS,
  `salvage-s025-browser-guide-01`.
- `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.object_store_fixture_admission,harness.browser.unit.testservices_lifecycle_contract`:
  PASS, `salvage-s025-services-01`, exit 0, 4/4 graph units, 17,978 ms.
- `make harness-smoke-bootstrap-node-runtime harness-smoke-cache-artifact`:
  FAILED before execution, `salvage-s025-bootstrap-01`, exit 2; internal smoke
  labels were mistaken for Make targets. No result root was published. The
  already selected harness-contract boundary route executes these fixtures.
- `make generate-drift`: PASS, `salvage-s025-drift-01`, exit 0, 17,448 ms.
- `make generated-artifact-policy-check`: PASS,
  `salvage-s025-artifact-policy-01`, exit 0, 924 ms.
- `make json-shape-check`: PASS, `salvage-s025-json-01`, exit 0, 3,573 ms.
- `make test-run-status RESULTS_DIR=.cartulary/test-results/salvage-s025-foundation-01 JSON=1`:
  PASS, read-only live observation; no workload acceptance claim.
- `git diff --check`: PASS.

All results and failed attempts remain. No qualification or diagnostic cohort
ran; 44.952 minutes of experimental collection remains unchanged. Primary code
and the index were not modified. No schema acceptance set changed. S-026 is ready
only after this checkpoint. Rollback is the focused routing commit; historical
foundation and archive remain intact. Broader real service/browser migration
checks remain assigned to S-029/S-030, not claimed by these mocked owner rows.

### S-026 checkpoint — experimental machinery removal complete

V-027 PASS. Integration commit `c09e8ba87e33110de85da7b131e94801bc52af5b`
removes the native measurement supervisor, qualification/diagnosis collection
entrypoints, reader-launch/gate machinery, memory-detail/lifetime collection,
producer hooks, collection-specific tests and four retired schema attachments.
The two retired Make commands and measurement row/build registration are removed
with regenerated projections. Graph and row runners retain their selection,
execution, local activities, cleanup and failure behavior. The demonstrated
case-scoped fixture repair remains. Trace publication is now a test-only fixture
that constructs independent bytes; production trace loading and validation remain.
Qualification arithmetic, measurement analysis, policy, validator and the authored
qualification plan are byte-identical to the predecessor. Generated validator
symbol renumbering does not change retained schema inputs or acceptance sets.

`restore-s026/` in the external archive contains all 26 deleted authored files,
restored from `primary-source.tar` and independently checked against deletion-time
hashes. `s026-removed-files.json` and `s026-restoration.json` record paths and hashes.
The matching frozen source/readers and all original evidence remain; no historical
qualification was rerun or relabeled. A repository-wide search across Makefile,
`tools`, `internal` and `contracts` found no remaining experimental target, meter,
lifetime hook or retired diagnostic-schema reference. Ordinary execution has no
experimental collection launch path.

Environment: same isolated Linux/WSL checkout and pinned Make tooling as S-025.
Functional source digest:
`sha256:c88d98ad80bb5a04d92046efa3faa746a367c9aeced20edef035e09e1e860fe6`.
Each root below is under integration `.cartulary/test-results/`; its name was the
explicit `CARTULARY_TEST_RUN_ID`:

- `make generate`: PASS, `salvage-s026-generate-01`, exit 0, 9,450 ms.
- `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.current_epoch_evidence,harness.evidence_accounting.behavior.local_migration_activities`:
  PASS, `salvage-s026-rows-01`, exit 0, 18,815 ms. This executes the evidence
  contract including independent qualification arithmetic, strict CPU equality
  rejection, eligibility, stability and integrity fixtures through the public
  shell row, plus the local-migration Go row.
- `make harness-command-surface-contract`: PASS, `salvage-s026-command-01`,
  exit 0, 51,120 ms.
- `make generate-drift`: PASS, `salvage-s026-drift-01`, exit 0, 17,614 ms.
- `make generated-artifact-policy-check`: PASS, `salvage-s026-policy-01`,
  exit 0, 874 ms.
- `make json-shape-check`: PASS, `salvage-s026-shapes-01`, exit 0, 3,549 ms.
- `git diff --check` and archive restoration/hash verification: PASS.

No failed validation command in S-026. No qualification, diagnostic collection or
broader service/browser run was launched. The external archive and historical
checkout were not changed; only restored copies were added to the archive.
Primary code/index remain preserved. S-027 is ready after this saved checkpoint.
Rollback is the whole removal commit together with its generated projections;
restoration must use matching historical readers, never a translator or a changed
historical schema. Certification remains deferred; F-022/RB-006 remain unresolved.

### S-027 checkpoint — native composition and preparation complete

V-028 PASS. Integration commit `3528f4379a3f74e91fb946f46e30818a312cabd4`
adds the thin `tools/harnessruntime` composition root and Make-built
`cartulary-harness`. Closed admission, collector and diagnostic subcommands reuse
the existing libraries. The private preparation operation starts only the
repository Node compiler with an inherited pipe, receives at most 32 MiB, waits
for successful compiler exit and validates strict JSON, generated schema shape,
unit/graph/snapshot digests, projection/dependency closure, cycles, resource claims,
locks, cache bindings and safe relative outputs before returning a plan.
Cancellation and terminal-state observation drain the compiler group before
reaping, keeping its identity reserved during signaling. Workspace admission is
inherited via a duplicated descriptor; the original descriptor is not closed.

`cartulary.harness_runtime_plan.v1` carries compact graph/projection/capacity facts,
launch bindings, preparation clock anchors and phase intervals, and cache identity
facts with shared source-entry indices. Node source content, catalog objects and
ASTs never cross the boundary. Off mode omits cache-input metadata. Normal/cold
source identities are checked against the full executable snapshot digest. Go
cache use/current-input revalidation belongs to S-028; no cache transaction starts
in this workstream. Preparation creates no retained run root or graph units.
The old public scheduler remains active. Shared selection/environment preparation
helpers were extracted without changing its execution dispatch.

`internal/testutil/harnessschema` now owns strict JSON and generated native
validation; diagnostic presentation delegates to it. The existing schema registry
and parity corpus remain the only projection inputs. Native regex parity probes
now follow compiled references rather than unused definitions in imported schema
files. No retained schema acceptance set changed. The permanent
`harness-runtime-contract` Make route and native-runtime/schema catalog rows were
added under the existing owners. Read-only status dispatch still has no binary
provisioning prerequisite. S-031 will migrate that dispatch to the new binary.

Environment: isolated Linux/WSL checkout and pinned Make tooling. Final native and
projection tests used source digest
`sha256:e77aaa922f2902ab4e26b1adb3c2e35233a3f8b7b656d70429373ee26a1dc26b`.
Earlier broader contract evidence used
`sha256:239e78ddb106b55c5982da56f168c531c0dc0772573a41d719d7f1aad57d1ea5`
before the final private cache-input fields; it is not relabeled as final-source
qualification. Final cache-field changes were covered by the native and projection
routes. Each command below used the named root as explicit `CARTULARY_TEST_RUN_ID`;
roots are under integration `.cartulary/test-results/`:

- `make generate`: attempts `salvage-s027-generate-01` and `-02` FAILED,
  exit 2, before the normal retained wrapper: unused catalog path-pattern parity
  projection was unsupported (`-02` added precise generator diagnostics).
  `-03` FAILED, exit 2, on another unused catalog-definition pattern. The repair
  selects actual compiled reference patterns; no authored schema was weakened.
- `make generate`: `salvage-s027-generate-04` FAILED, exit 2, tool exit 1,
  9,610 ms; the initial native Go row incorrectly selected the shell-only
  public-registry verification. Its step summary records the exact unsupported
  runner error. A dedicated native-runtime verification registration fixes it.
- `make generate`: PASS, `salvage-s027-generate-05` (10,510 ms), then final
  private cache-input projection `salvage-s027-generate-06` (9,870 ms), exit 0.
- `make format-go`: PASS, `salvage-s027-format-01`, `-02`, `-03`, exit 0.
- `make harness-runtime-contract`: PASS, `salvage-s027-native-01` (2,190 ms),
  `-02` (2,450 ms; terminal group drainage), `-03` (4,030 ms; normal/cold/off
  compiler parity), exit 0. Includes bounded/duplicate/incomplete/forged plan,
  graph closure, compiler failure, cancellation and lingering-descriptor cases,
  real selected-row preparation and generated Node/Go schema parity.
- `make harness-contract`: PASS, `salvage-s027-contract-01`, exit 0,
  104,655 ms; covers the extraction, evidence, command and foundation boundaries.
- `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_schema,harness.evidence_accounting.behavior.native_observation,harness.evidence_accounting.behavior.native_collection`:
  PASS, `salvage-s027-observation-01`, exit 0, 1,857 ms.
- `make generate-drift`: PASS, `salvage-s027-drift-01` (19,734 ms), final
  `salvage-s027-drift-02` (18,781 ms), exit 0.
- `make generated-artifact-policy-check`: PASS, `salvage-s027-policy-01`
  (1,029 ms), final `salvage-s027-policy-02` (1,040 ms), exit 0.
- `make json-shape-check`: PASS, `salvage-s027-shapes-01` (3,975 ms), final
  `salvage-s027-shapes-02` (3,809 ms), exit 0.
- `git diff --check`: PASS.

All failed attempts remain. No service/browser migration, qualification,
experimental collection or finalization was performed. Primary code/index remain
preserved. S-028 is ready after this checkpoint; S-029–S-032 remain gated. Rollback
is this complete preparation/projection/route commit, keeping S-026 retirement
and all archived evidence. This functional boundary does not resolve F-022/RB-006
or establish an overhead benefit. The experimental counter remains 44.952 minutes.

### S-028 checkpoint — Go graph engine complete

V-029 PASS. Integration commit `9fbcecdcdf9802f970ee20d12b39a5085171d6d1` adds one native coordinator
for DAG admission, critical-path rank, aging credits, capacity/lock reservations,
exclusive windows, dependency propagation and cancellation. Lifecycle workers
own acquired resources and acknowledge transitions through a bounded mailbox;
canonical event order and scheduling state have one writer. Independent cleanup
contexts drain host and fixture leases after cancellation or required writer
failure. Primary-failure selection uses the existing projected taxonomy and
keeps cleanup secondary to a non-cleanup failure. The private coordinator honors
the authored `continue_independent` flag; all currently compiled public units set
it true, preserving their dispatch behavior.

Native process execution preserves pinned Node launch, workspace descriptor 9,
fresh launch identities, private typed failure channels, bounded output tails,
timeouts and signals. WNOWAIT pins the leader until group signaling and pipe
drainage finish. Host admission calls the existing library directly. Command IDs,
failure ordering and phase limits are generated from their existing authored
owners; there is no new registry. Optional graph observation is isolated from
required outcomes and a failed sink is reported in the internal engine result.

The native cache preserves off/cold/normal and same-run behavior, fresh known
input validation, digest/manifest closure, corrupt-entry quarantine and immutable
concurrent entries. Artifact payloads stream through bounded buffers. Restoration
stages all outputs, validates descriptor/inode ownership and rolls back a failed
transaction; an incomplete rollback is an artifact failure. Same-run storage is
owned by one cache instance and removed only with matching inode proof. Full
source discovery remains in Node preparation, including new-file detection on
each invocation. Warm-fixture cache publication waits for physical cleanup.

Required publication emits the existing manifest, unit events/results, cleanup,
target and run summary formats. Canonical accounting retains per-unit/phase facts
rather than an accumulated event stream. Independently authored arithmetic agrees
with the existing Node reader; malformed closure fails publication. Unit logs are
bounded/redacted, including empty logs for unexecuted units. The generated native
schema closure adds existing required-result/cache/launch shapes and prefix-items
and path-pattern parity; no retained schema acceptance set changed.

Public Make execution still uses the existing scheduler. The real native-engine
integration fixture executes the exact local-migration row after Node preparation
has exited. Its test-only composition uses the existing private runtime owner;
shared fixtures, native private-runtime finalization and all observation/scan
wiring remain S-029/S-030 obligations. This is a native engine gate, not public
cutover, service/browser migration or overhead evidence.

Environment: isolated Linux/WSL checkout, pinned Make Go/Node toolchains.
The final catalog-selected row used source digest
`sha256:a80d1eb1383f7a02f37451e05cac4f958b163629fbfad1863f0afa90b35c1730`.
Race and broader contracts used
`sha256:688785dfd6d353a43594c844ccd81a63b12abea1f5a66ee1e6d9ed0bdf07d1d7`
before the final additional concurrent-failure test; executable implementation
was unchanged by that test addition. Earlier successes remain tied to their own
recorded sources. No source is qualified.

Every root below is under integration `.cartulary/test-results/`. Unless noted,
commands used `env CARTULARY_TEST_RUN_ID=<root> make <target>`:

- `make generate CARTULARY_TEST_RUN_ID=salvage-s028-generate-01`: FAILED,
  exit 2 before a retained wrapper; the run identity is not an admitted Make
  command-line input. Subsequent commands pass it in the environment.
- `make generate`: `salvage-s028-generate-02`, `-03`, `-04` FAILED, exit 2
  before normal wrapper publication, respectively on unsupported native
  `prefixItems`, the existing legacy-reference negative lookahead, and catalog
  reserved-segment lookaheads. The generator and independent parity corpus were
  extended; authored schema restrictions were not weakened.
- `make generate`: PASS, `salvage-s028-generate-05` through `-12`, exit 0;
  durations respectively 10,710, 10,410, 10,640, 11,400, 9,980, 10,010, 10,490
  and 9,570 ms as implementation/projection steps were added.
- `make format-go`: PASS, `salvage-s028-format-01` through `-09`, exit 0.
- `make harness-runtime-contract`: PASS, `salvage-s028-native-01` through
  `-05`, exit 0, respectively 4,300, 4,390, 4,790, 5,110 and 4,720 ms.
  These cover scheduling, cancellation, process drainage, exact private failure
  binding, admission, cache integrity, rollback and independent canonical parity.
- The same target `salvage-s028-native-06` FAILED, tool exit 1/Make exit 2,
  5,830 ms: the new real-row test composition omitted the required private suite
  runtime. Its nested failed graph remains at
  `salvage-s028-native-06/native-engine/selected-graph-34168`.
  `salvage-s028-native-07` PASS, exit 0, 6,590 ms after supplying the real
  existing owner rather than relaxing the row runner's binding check.
- `env GOFLAGS=-race CARTULARY_TEST_RUN_ID=<root> make harness-runtime-contract`:
  `salvage-s028-race-01` PASS (10,880 ms); `salvage-s028-race-02` FAILED
  at `harness-runtime-build/build-harness-runtime/stderr.log`, tool exit 1/Make
  exit 2 (190 ms), because command projection emitted an undefined ID for
  private helpers. Projection now emits only authored command IDs.
  `salvage-s028-race-03` PASS, exit 0, 12,630 ms, including the real row,
  same-run ownership, quietness and failed optional observation.
- `make harness-evidence-contract harness-command-surface-contract`: PASS,
  `salvage-s028-contract-01`, exit 0; targets 16,290 and 50,510 ms.
- `make generate-drift generated-artifact-policy-check json-shape-check`:
  `salvage-s028-projections-01` PARTIAL/FAILED, exit 2. Drift PASS, 18,573 ms;
  the next target correctly rejected reuse of that nonempty graph root. No
  failed root was cleared or reused.
- `make generated-artifact-policy-check`: PASS, `salvage-s028-policy-01`,
  exit 0, 999 ms. `make json-shape-check`: PASS, `salvage-s028-shapes-01`,
  exit 0, 3,596 ms, using separate roots.
- `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.unit.native_runtime`:
  PASS, `salvage-s028-row-01`, exit 0, 6,470 ms. Confirms authored exact
  selection of all native-runtime tests, including concurrent primary-failure
  precedence, and retains its nested real local-migration execution.
- `git diff --check`: PASS. Primary index and staged tracker SHA-256 are
  unchanged: `12215bad3a73c93e30b0f0afb11b54a561ffafb19b5c3ade72ee96e6fab875b6`
  and `9ab829e44fa8df7ba1ee86e794518acd97129cea6af416966fbdda4b33f83426`.

V-029 is complete; S-029 is ready only after this saved checkpoint. No service
or browser lifecycle run, finalization, qualification, or experimental collection
was performed. All failures and nested evidence remain. Rollback is this whole
engine/projection/test commit; public dispatch is still unchanged. F-022/RB-006
remain unresolved, certification remains deferred, and the original effort remains
IN_PROGRESS. The experimental counter remains 44.952 minutes.

### S-029 checkpoint — Shared fixture owner complete

V-030 PASS at integration commit `ba543bf694c862e15eb5c7e5f976e4883e5a1d90`. `harnessfixtures` now owns
fixture acquisition, lease identity, reference counts, warm reuse, taint,
partial acquisition, dependency ordering and recovery. Graph and UI-review
callers use the same owner through the private consolidated-binary session;
the Go graph engine has a direct library adapter. Public graph scheduling still
uses Node until S-031. No production caller invokes the JS fixture broker or
provider policy. Those definitions and their remaining characterization tests
are retained until the S-031 deletion transaction; their presence is not a
second active fixture engine. F-030 still requires V-032 deletion closure.

The session uses strict generated schema validation, inherited Unix-stream
socket peer/process proof, run/runtime-lease binding, monotonic request IDs,
1 MiB frames, 64 in-flight requests and a 65,536-request limit. It accepts only
acquire/release/recover/close operations. EOF/cancellation drains owned work;
repeated close joins its original outcome. Credentials and container proofs stay
private. The client transports facts and presents cleanup results; it does not
choose reuse, service ownership or recovery order.

Existing physical service, browser, renderer and proxy lifecycle owners remain.
The native owner retains exact proof through stop acknowledgement, keeps waited
helper identities pinned until group cleanup, blocks service closure beneath
unresolved dependents, and detaches borrowed suites without terminating their
containers. Browser attachments bind exact run/suite/session retained paths.
Borrowed browser allocations receive a run-scoped attachment proof rather than
ownership of their persistent service. Directory removal verifies and moves the
owned inode before recursion. Recovery rejects unknown/forged dependencies.

Make declares the new package in the consolidated binary build inputs and passes
its selected Go executable to children. This repairs stale cached fixture code
and the proxy-builder configuration failure encountered during validation. Shared
native schema generation now supports `patternProperties`, with an independent
parity probe, and projects the existing fixture/service/runtime shapes plus the
new private session schema. Retained public acceptance sets are unchanged.

Cleanup presentation now supplies positive sequence ordinals to the existing
failure selector. An independent first-artifact/later-cleanup fixture verifies
that the earliest cleanup failure remains primary. Session I/O rejection is a
typed artifact failure; acknowledgement failures stay secondary to the original
rejection. The UI live-cleanup test withholds safe access to real ownership proof,
checks retained failure facts and a still-live dependent, restores that exact
proof, rejects foreign-run recovery, and then confirms exact cleanup. It no
longer mutates a JS broker implementation that is absent from production.

Environment: isolated Linux/WSL checkout and pinned Make Go/Node toolchains.
Final committed-source manifest digest: `sha256:ffaf09dc9b05c223a67884753adb9dcb2217fb43b21a754cd50eaf156f44088c`; toolchain digest: `sha256:4b62e2558e6dcbe1d3c632fb9c0bb47e5a89e5dda08f103a72f61347b5ef1810`.. Evidence is functional validation, not overhead qualification.
Diagnostics remain default-off. No qualification/diagnostic collection ran;
44.952/90 experimental minutes, all budgets, F-022/RB-006, and deferred
S-020–S-022 dispositions are unchanged.

Commands used `env CARTULARY_TEST_RUN_ID=salvage-s029-<suffix> make <target>`
from the integration root. Exact target/outcome/root ledger is
`/home/jochi/cartulary-instrumentation-archive/20261011T182827Z/s029-validation-ledger.json`.
Final passes: `harness-fixture-contract` (fixtures-16, race enabled),
`harness-runtime-contract` (runtime-01), `harness-evidence-contract` (evidence-01),
`harness-command-surface-contract` (commands-02), `harness-ui-review-contract`
(ui-contract-02), `harness-ui-review-lifecycle` (ui-lifecycle-05),
`harness-ui-review-seeded-default` (ui-seeded-06), `harness-contract-tests`
(contract-tests-02), `generate` (generate-13), `format-go` (format-20),
`generate-drift` (drift-02), `json-shape-check` (shapes-02),
`generated-artifact-policy-check` (policy-02). The final four projection/row roots
were collected after the committed freeze; earlier passes retain their own source identities.

Exact selections passed: `test-slice OWNER=harness.browser
ROWS=harness.browser.unit.native_fixtures` (native-row-02);
`service-backed-test-slice OWNER=module.database_migrations
ROWS=module.database_migrations.integration.schema_bootstrap` (schema-bootstrap-02);
`service-backed-test-slice OWNER=harness.browser
ROWS=harness.browser.boundary_support.https_trust` (https-03); and the same owner
with `ROWS=harness.browser.integration.object_store_fixture_roundtrip`
(object-store-02). `task-guide ROLE=module-author OWNER=harness.browser`
(guides-01) established routing. `git diff --check` passed.

Failure dispositions: generate-03 attached an unnecessary unsupported browser
schema; generate-09 had unsorted catalog assertions. fixtures-03 had an unused
import; fixtures-04 omitted synthetic parent directories; fixtures-06 serialized
empty secrets as null; fixtures-10/11 used an incorrect synthetic run root.
https-01/02/debug-01 exposed missing binary build inputs and browser acquisition
binding/path defects. ui-seeded-01 was invalidated by a concurrent source edit;
ui-seeded-02 through -05 exposed cleanup expectation, rejection typing and
primary-failure/idempotence defects, all repaired before ui-seeded-06.
ui-lifecycle-02 synthetic callers lacked required run/private-root identity;
contract-tests-01 exposed two forbidden imports; object-store-01 exposed the
missing selected-Go environment binding. Nested deliberate fault-fixture
failures are not relabeled as failed parent targets. A temporary debug hook was
removed before the final freeze; debug-02 is supporting evidence only.
No end-of-effort finalization, broad warm check or certification ran here; those
remain S-032 or explicitly deferred, respectively.

Failed attempts remain in their original roots, with original dispositions.
Intermediate dirty-source attempts are not relabeled as final-source passes.
External archive manifest `s029-archive-manifest.json` records the verified full
Git bundle `salvage-through-s029.bundle` (SHA-256
`0e7227662c81f8483b39cb51bee5d5fc395d364207b9ff578f9037dc4f6e4aba`)
and `salvage-evidence-through-s029.tar` (SHA-256
`c981deb1369e96da300613681ec4ecf3bc6785d5e42340500342d86970470856`).
All integration result roots through S-029, including failed attempts, were
included; archived summary bytes were reconstructed and compared with source.
Earlier S-023/S-026 archives and original evidence remain unchanged.

Primary staged tracker SHA-256 remains
`9ab829e44fa8df7ba1ee86e794518acd97129cea6af416966fbdda4b33f83426`;
primary index SHA-256 remains
`12215bad3a73c93e30b0f0afb11b54a561ffafb19b5c3ade72ee96e6fab875b6`.
Primary code and unrelated `.keep` content are untouched. Rollback is the whole
S-029 integration transaction, including authored routes and generated outputs;
keep evidence and use the archive/preceding revision, never reset the primary
checkout. S-030 is ready only after this checkpoint is saved. Native graph
observation, private-runtime finalization/scan integration, and public cutover
remain S-030/S-031. Final functional handoff remains S-032; salvage is not complete.

### S-030 checkpoint — Native graph observation integration complete

S-030 / WF-30 / V-031 is DONE at integration commit
`79aecc1cae7a9474c92f4a9b8f77092bc4363662`, source digest
`sha256:9928699c97998338bc39879b97c5183151fe876b87e5ae4f70d820cdbbe42852`.
The integration checkout is `/home/jochi/.codex/worktrees/instrumentation-salvage/cartulary`.
Pinned Linux/WSL toolchain digest remains
`sha256:4b62e2558e6dcbe1d3c632fb9c0bb47e5a89e5dda08f103a72f61347b5ef1810`.
This is functional evidence, not qualification.

The native graph composition now owns canonical publication, clock anchors,
execution records, exact-container registration, isolated same-binary collection,
coherent live status, envelope publication, private-runtime shutdown and retained
scanning. Collector termination drains readers/writers and process groups before
acknowledgement; inability to drain is a required cleanup failure. Optional
observation failure cannot overwrite a primary workload failure. Live publication
pauses through the existing observation gate. Execution publication closes its
reference set and rejects writes after shutdown. The compiler exits before graph
units launch; its elapsed preparation is observed in the graph clock as a coarse
runtime_prepare interval, without pretending its private subphase clock is the
graph clock or adding preparation to canonical execution.

Wrapper graph registration uses the existing private owner proof. All ordinary
producers close before the scan; only the existing schema-closed outer receipt
can follow it. Multi-graph tests validate one shared outer duration and two exact
memberships through the independent Node reader. Retained schema acceptance sets
are unchanged. Go scans preserve the existing policy digest and rejection rules;
independent Node parity covers cross-chunk secrets and unsafe retained entries.
Private roots now have explicit preservation/recovery and bounded stale-root
cleanup. A unit's host-admission deadline uses its declared timeout, matching the
existing scheduling policy; a regression distinguishes cancellation from an
admission resource conflict and verifies released queue state.

Permanent routes include race-enabled `harness-runtime-contract` and the separate
`harness-runtime-boundary-contract` for real service/browser observations.
The native boundary route runs exact schema_bootstrap and HTTPS-trust selections,
validates retained/live evidence independently, and requires real owned-container
observations in the HTTPS run. Public graph dispatch still uses Node until S-031.

Commands used `env CARTULARY_TEST_RUN_ID=salvage-s030-<suffix> make <target>`.
Exact command/target/outcome/root records are in external
`s030-validation-ledger.json`; reader commands and complete output are in
`s030-reader-ledger.json` and `s030-*.log`. Final relevant passes:
`harness-runtime-contract` (runtime-07, race enabled),
`harness-runtime-boundary-contract` (boundary-02, committed source),
`harness-fixture-contract` (fixtures-01), `harness-evidence-contract` (evidence-01),
`harness-command-surface-contract` (commands-01), `harness-contract-tests`
(contract-tests-01), `generate` (generate-04), `format` (format-08),
`generate-drift` (drift-01), `generated-artifact-policy-check` (policy-01),
`json-shape-check` (shapes-01). Exact `test-slice OWNER=harness.evidence_accounting`
with native_collection, native_observation and observation_admission rows passed
(observation-rows-01, committed source). The native runtime route also exercises
exact local-migration rows with diagnostics off/basic. `task-guide
ROLE=module-author OWNER=harness.evidence_accounting` confirmed routing.
`git diff --check` passed.

Fresh exact native root
`.cartulary/test-results/salvage-s030-boundary-02/native-boundary/https`
passed `harness-observability-check`, `test-run-status JSON=1`, and
`explain-run DETAIL=performance|resources` in human/JSON forms. Both JSON views
share identical validated facts. Ordinary reader/export isolation is covered by
the passing evidence contract; no external telemetry was sent. The direct native
service fixture deliberately has no outer wrapper receipt; the separately
wrapped multiple-graph fixture supplies that acceptance evidence. Earlier roots
keep their own source identities and are not relabeled as committed-source passes.

Failed attempts are preserved: runtime-02 exposed a stale descriptor pathname
after staging-root rename; the descriptor is now relabeled through a close-on-exec
duplicate. runtime-03 exposed fractional timestamp precision outside the existing
live schema; publication now uses exact millisecond UTC timestamps. All subsequent
narrow runtime checks pass. Deliberate nested fault-fixture failures remain test
evidence, not failed parent outcomes. No source-sensitive test was edited while
its final run was active.

External archive `/home/jochi/cartulary-instrumentation-archive/20261011T182827Z`
contains `salvage-through-s030.bundle` (SHA-256
`20fcda97fa578fe813e227447164d0f50899104712f69439221c23db7058eca1`),
`salvage-evidence-s030.tar` (SHA-256
`4f82982b9c635cd6584a4e191099cc6f9200475eb84d573a8da1936c49567434`),
and `s030-archive-manifest.json`. Bundle verification and byte comparison of
78 restored summary artifacts passed. Every S-030 root, including failures,
remains in place and in the archive. Previous archives are unchanged.

Primary index SHA-256 remains
`12215bad3a73c93e30b0f0afb11b54a561ffafb19b5c3ade72ee96e6fab875b6`;
staged tracker SHA-256 remains
`9ab829e44fa8df7ba1ee86e794518acd97129cea6af416966fbdda4b33f83426`.
Primary code and unrelated content are untouched. Rollback is the full S-030
transaction including routing and generated outputs, using the preserved prior
revision; never reset the primary checkout. S-031 becomes ready after this saved
checkpoint. End-of-effort finalization and full warm verification remain S-032.
No certification or diagnostic collection ran: 44.952/90 experimental minutes,
CPU mean <2.5%, all other budgets, default-off diagnostics, deferred S-020–S-022,
and unresolved F-022/RB-006 remain unchanged. Overall original effort IN_PROGRESS.

### S-031 clean-cutover planning checkpoint — IN_PROGRESS

The latest user steering requests a clean harness and non-production test cutover,
following the purpose of existing tests. The previous S-031 deletion/porting work
is preserved, unfinished, in the isolated integration checkout. No additional
code, test, projection, configuration or generated edit was made after that
request; this checkpoint updates the sole tracker and its established mirror.
It does not mark V-032 or V-033 passed.

Integration HEAD remains `79aecc1cae7a9474c92f4a9b8f77092bc4363662`. Before this
planning boundary, `env CARTULARY_TEST_RUN_ID=salvage-s031-generate-01 make generate`
and `env CARTULARY_TEST_RUN_ID=salvage-s031-cutover-smoke-01 make test-slice
OWNER=harness.evidence_accounting
ROWS=harness.evidence_accounting.behavior.local_migration_activities
HARNESS_DIAGNOSTICS=basic CARTULARY_HARNESS_CACHE_MODE=off` passed. Their retained
roots are under the integration checkout's `.cartulary/test-results/` with those
exact IDs. Subsequent edits/deletions invalidate them as passes for the current
candidate. There is no new frozen executable digest or qualification result.

Section 7 now defines the concrete remaining sequence: agree behavior/test
ownership; complete one source/test/routing transaction; pass meaningful native,
producer, public and physical-boundary checks; then finalize S-032. F-031 records
newly identified duplication in the test-only JS scheduler, copied wrapper
coordinator and repeated full-suite smoke routes. Test-internal structure can be
simplified; required security, lifecycle, accounting and product coverage cannot
be discarded. Existing native tests are the starting point, not a mandate to
port every old assertion or build a second testing framework.

Read-only inspection used `git --no-optional-locks`, `rg`, exact file reads and
SHA-256/lstat snapshots. The review covers the changed scheduler/compiler,
fixture, producer/reader and authored routing seams named in Section 7; it is not
an exhaustive application-test audit. All 6,903 primary and 6,937 integration
tracked/non-ignored entries were fingerprinted for preservation, not claimed
behaviorally inspected. Ignored caches/evidence were not rehashed in this
planning checkpoint; no build, test, cleanup, service or collection command was
run after the planning boundary. Existing archive hashes/restoration results
retain their original applicability. Referenced files not read in full retain
their prior inspection limits; final caller/coverage closure is step 1/2 work.

`make lint-markdown` was skipped: its inspected wrapper prepares Node/frontend
inputs and writes static-analysis cache/stamp plus retained output. Generation,
contract tests, service routes and `agent-finalize` were also skipped during this
planning-only update. Finalization remains the first broader completion action
in S-032. No skipped command is reported as a pass.

The planning snapshot records primary index SHA-256
`12215bad3a73c93e30b0f0afb11b54a561ffafb19b5c3ade72ee96e6fab875b6` and staged tracker
SHA-256 `9ab829e44fa8df7ba1ee86e794518acd97129cea6af416966fbdda4b33f83426`;
integration index SHA-256
`b09c72118e89ebd243c28a59b1ff714cbb92df4f22239e96dcfbd0dde7d627e8`.
Both tracker copies began with SHA-256
`e494cae568968aca6e1a0c2043df2b2b51bdafb99c5e305e0ad68f090e4a7e04` and were
regular singly linked files with no symlinked path components. Their content was
rechecked immediately before writing. Final preservation verification is recorded
below after comparing all snapshot entries and index bytes.

S-031 is ready for the first ordered implementation step; completion is blocked
on V-032. S-032 remains blocked on the saved S-031 completion checkpoint.
Original effort IN_PROGRESS; qualification/S-020–S-022 DEFERRED; F-022/RB-006
unresolved; CPU mean <2.5%, all other budgets and the 44.952/90 experimental-minute
ledger unchanged. No further collection is scheduled.

Final planning verification: both tracker-scoped `git --no-optional-locks diff
--check` commands passed. All 6,903 primary and 6,937 integration snapshot entries,
including missing/deleted and untracked entries, matched outside the tracker;
file modes/identities and both indexes/staged tracker bytes matched. All twelve
numbered sections remain in order, F-031/E-030 are unique entries, and the mirror
matches the controlling tracker. No unexpected non-tracker changes were observed
after the planning boundary. This is preservation/document validation only.
