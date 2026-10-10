---
doc_id: cartulary.testing_harness.instrumentation.revision_plan
title: Testing harness instrumentation
doc_type: revision_plan
status: IN_PROGRESS
authority_boundary: Human implementation tracker and rationale; the Testing Harness NLSpec owns behavior. No executable tooling consumes this document.
---

# Testing harness instrumentation

## Baseline and authorization

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

**Selected iteration: harness execution identity and owner-local lifecycle diagnostics.**
Artifact mode: `normal`. The user selected identity/allocation correlation and
explicitly included migration, fixture-reset and report-parsing timings measured
by their owners. S-000–S-006 is complete. The current user request authorizes
S-007–S-013; all seven slices have passed their exits and final handoff is complete.
The preceding planning invocation authorized only the tracker update, not
implementation; that historical authorization does not limit the current request.

Repository root is `/home/jochi/code/cartulary`. The existing handoff establishes
the literal inventory target `tools/harness/observability`, label
`harness-instrumentation`, with the coupled launch, fixture, runtime and producer
seams in Section 2. The admitted output is
`docs/handoffs/cartulary-test-harness-instrumentation.md`. It resolves inside the
repository, is a regular file with one hard link, and has no symlinked path
components. Its declared target identity matches the requested work.

Planning follows [the repository framework](cartulary_modular_refactor_planning_framework.md)
and the refactor-tracker format. Core 00 sections 2 and 5 allocate authority.
The [Testing Harness NLSpec](../testing-harness-nlspec.md) is adopted/current,
`cartulary.testing_harness.v3`; its command-failure, fixture ownership, public-change
and REQ-830–844 contracts govern the affected boundaries. S-007 adopted C-006–C-009
before the implementing slices changed machine projections.
OTEL-REQ-148–149 keeps harness diagnostics separate from application telemetry.
`domain.md` supplies vocabulary, not new harness behavior.

[NLSpec guidance](../research/nlspec-spec.md) v0.2.2 informs conceptual fidelity,
explicit interfaces/defaults, define-once mappings and testable acceptance.
Its examples and instructions are reference material, not execution authority.
The tracker records contracts and gates; adoption belongs to the Testing Harness owner.

Container counters, full wrapper/post-scan coverage, cross-clock joins and total
observer/0/1/4-reader qualification remain deferred. There is no new public
command, daemon, plugin system, dependency, application telemetry integration,
automatic export, performance claim or default-on decision. Product behavior,
test assertions and fixture ownership are preserved.

| Baseline item | Starting evidence / identity | Final comparison / outcome | Limits |
| --- | --- | --- | --- |
| Branch / commit | `main`, `9359eb0cd359ab6bbc2eead16769950685d197b5` | E-011 / E-018 | Prior `5d80ca3` foundation evidence remains historical. |
| Index | Clean; raw index SHA-256 `008c4fc2bd5d022bdaddd05fe9085df2aac8feda7c72ea30b83c7f834bd78c0c` | E-018 compares bytes; Git optional writes disabled | Different from E-010's earlier staged-entry fingerprint method. |
| Working tree / untracked | Clean; no non-ignored untracked files; SHA-256 and lstat snapshot of 56 target/governance/coupled files | E-018 | All 27 target files are covered; not a hash of the entire repository or ignored outputs. |
| Tracker | Initial SHA-256 `47bd2142e11090a993253fe2119eef4b93cb28914906cff7ca2b681f1f632752`; device 2096, inode 3903446, links 1 | Recheck identity/content immediately before write; E-018 | Only admitted persistent repository change. |
| Ignored outputs | Existing `.cartulary/test-results` and retained-work evidence | Existing ledgers preserved; no new run created | Historical passes are not current-source verification. |

The preceding read-only planning inspection and this update use the same clean
commit. Source changes from the earlier foundation baseline were reinspected at
the selected seams; obsolete foundation diagnoses are retained as history rather
than reused as current findings. The universe is tracked plus non-ignored untracked
files under the literal target. All 27 are tracked authored regular files; no
generated, vendor, binary, symlink or submodule entry occurs there. Ignored build,
cache and dependency trees are excluded. This is a bounded subsystem plan, not
an exhaustive repository architecture audit.

## 2. Planning Baseline Repository Inventory

Paths in the first table are relative to `tools/harness/observability`.
There are 18 directly inspected entries and nine metadata-only entries.
Direct inspection is limited to the named responsibilities/assertions; it is not
a whole-file security audit. Metadata-only entries have no proposed movement;
behavioral changes there require a new direct inspection. E-011–E-016 record the
current evidence. Earlier 16-file accounting in E-002/E-010 belongs to the
foundation planning baseline.

| Path | Inspection disposition | Current responsibility / consumers | Dependencies / test surface | Contract / disposition / risk |
| --- | --- | --- | --- | --- |
| `canonical-evidence-audit-cli.mjs` | Metadata-only | Existing audit CLI; no change proposed | Canonical audit routing | C-001; preserve; low |
| `canonical-evidence-drift-suite-cli.mjs` | Metadata-only | Existing drift-suite entrypoint; no change proposed | Canonical audit routing | C-005; preserve; low |
| `canonical-evidence.mjs` | Direct: retained identity/reconstruction | Canonical run validation used by check/explain/export | Secure reads, normalized accounting; tamper fixtures | C-001, C-009; preserve; high |
| `canonical-performance-cli.mjs` | Metadata-only | Performance qualification CLI; no change proposed | Qualified-window builder | C-005; preserve exit 13; low |
| `canonical-performance.mjs` | Metadata-only | Qualified windows and baseline comparison | Existing baseline schemas and performance tests | C-005; qualification deferred |
| `collection-engine.mjs` | Direct | Collection lifecycle, process and lease facts | Worker, injected gate/adapter/store; engine tests | C-002, C-007; correlation seam; high |
| `diagnostic-session.mjs` | Direct | Collector/live shutdown and pre-scan receipt | Graph composition, secure publication | C-002, C-009; preserve quiescence; high |
| `observability-check-cli.mjs` | Direct | Exact-run canonical and optional diagnostic check | Read facade, resource projection; command fixtures | C-004, C-009; preserve public exits; medium |
| `observability.mjs` | Direct | Exact selection, validated loading, explanation composition | Canonical validator, resource projection, formatter | C-004, C-009; extend optional analysis; medium |
| `otel-export-cli.mjs` | Metadata-only | Explicit export transport/security; no change proposed | Export-only projection and existing transport tests | C-004; no automatic/new export |
| `otlp-projection.mjs` | Metadata-only | Export payload construction; no change proposed | Explicit exporter only | C-004; preserve isolation |
| `performance-analysis.mjs` | Direct | Pure selected-target explanation | Validated facts and optional resource/envelope data | C-001, C-008, C-009; separate local timing; high |
| `performance-presentation.mjs` | Direct | Bounded human/JSON formatting | Explanation schema; public reader | C-004, C-009; bounded parity; medium |
| `resource-accumulator.mjs` | Direct | Counter/gauge aggregation, joint CPU windows, unit correlation | Generated descriptors; independent arithmetic fixtures | C-003, C-007; shared attribution risk; high |
| `resource-collector-worker.mjs` | Direct | Worker composition and control/correlation transport | Engine, runtime gate, Linux adapter, store | C-002, C-007; thin transport; medium |
| `resource-collector.mjs` | Direct | Controller, proof registration, bounded correlation queue | Graph caller; worker and instrumentation policy | C-002, C-006, C-007; context migration; high |
| `resource-linux.mjs` | Direct | Physical process proof and process/guest/cgroup sampling | Built-in Linux APIs; PID-reuse/counter fixtures | C-003, C-007; keep physical identity; medium |
| `resource-presentation.mjs` | Direct | Bounded resource scopes, leases and unit CPU view | Pure accumulator and resource facade | C-003, C-007, C-009; shared work disclosure; high |
| `resource-projection.mjs` | Direct | Composition of secure read, accumulation and presentation | Read facade/check | C-003, C-009; keep boundary; medium |
| `resource-reader.mjs` | Direct | Bounded schema, identity, order and digest validation | Secure-local-file APIs; malformed artifact tests | C-003, C-007, C-009; integrity; high |
| `resource-store.mjs` | Direct | One bounded sample writer and atomic index publication | Engine; writer-failure/cap tests | C-002, C-003; keep writer ownership; medium |
| `run-comparison.mjs` | Direct | Strict same-source descriptive comparison | Explanation and compatibility fixtures | C-005; no waiver or timing gate change |
| `tests/test-canonical-performance.mjs` | Metadata-only | Existing qualification tests; no change proposed | Performance builder | C-005; presence is not execution |
| `tests/test-collection-engine.mjs` | Metadata-only: scenario inventory | Existing deterministic lifecycle scenarios | Evidence-contract route; no fixture move proposed | C-002; preserve real and deterministic coverage |
| `tests/test-foundation-characterization.mjs` | Direct: relevant assertions | Independent accounting, phase limit, schema and export fixtures | Evidence-contract route; generated extension fixture | C-001, C-003, C-004; preserve baseline |
| `tests/test-instrumentation.mjs` | Direct: affected assertions | Process proof, shared leases, resource integrity and real quietness | Evidence-contract route | C-002, C-003, C-007, C-009; expand characterization |
| `tests/test-observability.mjs` | Metadata-only: scenario inventory | Existing canonical/public recipe invariants | Harness-contract route | C-001, C-004; no test pass inferred |

The directly inspected expansion outside this directory is below. Consumer
searches used `rg` across repository source and schema/routing inputs, excluding
generated bodies where irrelevant. Searches do not prove the absence of dynamic
or external private imports; no speculative external compatibility promise is made.

| Coupled boundary / directly inspected seams | Observation and intended ownership | Evidence / limit |
| --- | --- | --- |
| `runtime/command-failure.mjs`, `runtime/private-child-process.mjs` | Fresh identity currently belongs to failure-channel creation; runtime owns launch identity and secure private capture | E-012; preserve required failure and cleanup semantics |
| `scheduler/work-graph/{executor,runner-cli,scheduler}.mjs` | Unit launch callbacks register PID/unit; scheduler owns canonical time and parent result; graph composes pre-scan closure | E-012/E-015; no scheduler redesign |
| `execution/runners/row-runner-cli.mjs` | Builds packed Go/Vitest/shell invocations; selected rows can share a launch; adaptation is an explicit report boundary | E-012/E-014; framework-internal workers are not automatically harness launch attempts |
| `browser/browser-catalog-group-cli.mjs` | Playwright launch/capture and report adaptation; cleanup is required even on interruption | E-012/E-014; product browser behavior unchanged |
| `scheduler/fixture-broker/{index,providers}.mjs` | Allocation reference derives from an acquisition lease ID; successful lease publication supplies correlation; shared browser processes use allocation/null-unit ownership | E-013; create allocation identity before acquisition, preserve lifecycle authority |
| `browser/{browser-session-evidence,browser-reset-attempt}.mjs` and `browser/lifecycle/reset-route.sh` | Existing startup records and reset outcome/duration artifacts | E-014; inspect changed producer paths again before implementation |
| `internal/testutil/{pgtest/pgtest.go,suiteservices/lifecycle.go,suiteservices/diagnostics.go}` | Existing monotonic migration durations, helper lifecycle clocks and fixture summaries; observed migration recording follows successful completion | E-014; reuse measured facts, add owner-local failure coverage |
| `runtime/observation-gate.mjs` | Runtime owns fences, watcher coalescing and worker liveness | E-015; preserve foundation boundary |
| Definitions/policy, resource index schema, schema attachments and generated-artifact policy | Closed signals/phases, separate policy limits, generated validator membership | E-015; inspect exact new schema closure in S-007, generate only in later authorized work |
| `tools/task_surface_owner.json` and selected owner/test routes | Evidence/command contracts and broader harness routing; Markdown wrapper has readiness/cache/artifact effects | E-016; task guides and additional affected owner rows resolved at implementation entry |

Backend application packages, web UI, SQL migrations and application OTel are
not proposed movement targets. The next discovery boundary is the exact
producer/caller and test-row inventory in S-008, not a new broad repository scan.

## 3. Module Boundary Diagnosis

The foundation already separates collection control, runtime admission, storage,
pure accounting, retained loading, presentation and export. Keep those boundaries.
The remaining work connects identities and measurements across their actual
owners; it does not justify another collector or accounting engine.

| Responsibility | Current location | Owner / keep, move, split or defer | Evidence | Finding / contract |
| --- | --- | --- | --- | --- |
| Launch identity | Private command-failure context; partially repeated in runner adapters | Split neutral invocation context into runtime; failure transport consumes that identity | E-012 | F-010 / C-006 |
| Physical process facts | Linux proof, collector and resource index | Keep physical identity immutable; separate explicit/inferred attribution relationships | E-013/E-015 | F-011 / C-007 |
| Allocation and lease lifecycle | Fixture broker/providers | Keep lifecycle authority; create allocation identity before acquisition and observe relationships | E-013 | F-011 / C-007 |
| Migration/reset/report detail | Go fixture measurements and Node/shell producer boundaries | Keep measurement at each owner; project closed, safe timing facts | E-014 | F-012 / C-008 |
| Retained analysis | Secure readers, pure projections and bounded formatters | Keep split; normalize new facts once and validate relationship closure | E-015 | F-011/F-012 / C-009 |
| Canonical timing, admission and export | Foundation facades | Keep; no duplicate event replay, raw admission coupling or eager export | E-015 | C-001–C-005 |
| Container/full wrapper and qualification | Deferred original delivery | Defer under T-007/T-008; do not infer unavailable attribution | Foundation ledger | F-007 / C-005 |
| Current/historical documentation | This tracker | Replace obsolete current summaries; keep stable historical evidence | E-011/E-017 | F-013 |

## 4. Public Contract and Behavior Freeze Map

Adopted owners constrain implementation. Existing defects are not compatibility
requirements. S-007 adopted the extensions below before projection changes; no
schema change follows merely from a private refactor.

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

HTTP, WebSocket, entity mutations, saved views, revisions, UI selectors and product
authorization are not applicable: no product contract is changed. Required
privacy/cleanup errors retain their existing consequences; optional observation
loss cannot replace a primary result. If characterization reveals an adopted-owner
violation, record its trigger, owner clause, separate correction and regression
test. Do not silently broaden structural preservation or bless it with a golden
update. No applicable-primary-owner contradiction was identified.

## 5. Coupling and Boundary Findings

`must_fix` is scoped to the named future slice or acceptance condition, not
unrelated cleanup. F-001–F-006 and F-008/F-009 are completed foundation findings;
their original diagnoses/actions below are historical, with execution closure in
the foundation checkpoints. F-007 remains an explicit coverage/qualification
limitation. No completed foundation finding is reopened by this plan.

### Foundation findings and stable IDs — historical

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

### Current findings and remediation

| Finding / evidence / contracts | Classification / affected areas | Remediation | Rationale / long-term benefit | Compatibility / migration | Risk if unresolved | Validation / slices |
| --- | --- | --- | --- | --- | --- | --- |
| F-010 invocation identity is confined to private failure channels; E-012; C-006 | must_fix for S-009; high; specification, runtime, runner adapters, tests | Neutral launch context shared by observation and failure handling; fresh identity before each harness-owned attempt, explicit parent/unit/rows and outcome | Successful, failed, cancelled and unlaunched attempts use one identity model; failure transport no longer owns general execution meaning | Migrate private callers together; retain strict failure-envelope identity checks, cleanup and public exits; no forwarding aliases | Nested/packed runners remain ambiguous and depend on process sampling for correlation | V-007/V-008; S-007/S-008/S-009; independent Go/Vitest/shell/Playwright and unit-launch fixtures |
| F-011 allocation/process correlation lacks complete lifecycle relationships; E-013/E-015; C-007/C-009 | must_fix for S-010/S-012; high; specification, broker/providers, collection/readers, tests | Allocate opaque physical-allocation identity before acquisition; separate lease, invocation and process relationships, failed attempts and reuse | Ownership remains cohesive; shared physical work is counted once without assigning it to its first consumer | Preserve ownership/reuse/cleanup; remove redundant private correlation paths after atomic caller migration; version incompatible retained shapes | False per-unit attribution, shared-work double counting and lost failed-acquisition context | V-009/V-011; S-010/S-012; shared/borrowed/reused/failed acquisition, PID replacement/reuse, missing proof and conflict fixtures |
| F-012 detailed timings are disconnected from execution identity; E-014; C-008/C-009 | must_fix for S-011/S-012; medium; specification, catalog, producer owners, retained analysis, tests | Reuse migration/reset measurements; add missing failure outcomes and report-parse boundaries with owner clock and invocation/allocation links | Measures work at a known boundary and avoids duplicate measurements or invented scheduler time | New optional diagnostic data; no canonical duration/path changes; missing/incompatible clocks remain unavailable | Setup/parsing costs stay obscured; success-only data and cross-process subtraction can mislead | V-010/V-011; S-011/S-012; hand-calculated success/failure/cancellation/gap/nesting/overlap/shared-clock cases |
| F-013 current tracker summaries describe completed foundation work as future work; E-011/E-017 | should_fix; documentation; this update | Replace current summaries, inventory, readiness and next-entry text; preserve stable IDs and historical ledgers under their original baselines | One unambiguous starting point and no repeated foundation delivery | No runtime migration; original run identities, failed attempts and applicability limits retained | Repeated work or historical success presented as qualification | V-013 / T-011; complete finding→contract→slice→validation trace and tracker-only preservation |

F-010–F-012 are planned extensions, not claims that the inspected implementation
violates newly invented requirements. S-008 must distinguish an existing adopted
owner defect from missing future coverage. Each actual defect receives an explicit
correction obligation and regression fixture.

## 6. Refactor Workstreams

### Completed foundation dependency history

These workflow IDs remain stable and DONE at their recorded execution baseline.
The goals below explain the delivered structure; they are not current work.

| Workflow | Historical prerequisites | Delivered goal / seams | Historical validation | Recorded exit |
| --- | --- | --- | --- | --- |
| WF-00 specification and tracker reconciliation | root; current implementation request | Tracker and Testing Harness REQ-840 / AC-137 | V-001 and Markdown lint | Adopted projections, phase bound, and execution checkpoints agree |
| WF-01 characterize public and private boundaries | chain; WF-00 | Existing instrumentation/observability and wrapper tests | V-002, V-003, V-005 | Independent fixtures cover frozen behavior and expose lifecycle entry assumptions |
| WF-02 consolidate definitions | chain; WF-01 | Catalog, schemas, generator, attachments, ownership | V-004 | One vocabulary source; generated validation and runtime descriptors agree |
| WF-03 normalize canonical facts | chain; WF-02 | Canonical reducer, performance projection, scheduler writer, validator | V-002 | No duplicate raw-event array; projections preserve expected values/errors |
| WF-04 isolate collection lifecycle | chain; WF-03 | Engine, worker/controller, Linux adapter, runtime gate, run session | V-003 | Deterministic interleavings and real quietness tests pass; resources close before scan |
| WF-05 simplify consumers | chain; WF-04 | Observability facade, resource reader/reducer, export, wrapper | V-005, V-006 | Public parity and ownership checks pass; removed private paths have no retained callers |
| WF-06 validation and handoff completion | chain; WF-05 | Final verification and durable handoff | V-006 plus Markdown/diff checks | All foundation gates pass; T-007/T-008 remain deferred |

### Next iteration — authoritative prerequisites

This table is the sole prerequisite list for WF-07–WF-13. Slice rows reference it.
All seven workflows are included and execute serially; none depends on T-007's
deferred remainder or T-008. No parallel agent work is requested.

| Workflow / slice | Class / authoritative prerequisites | Goal / likely seams | Validation / checkpoint | Principal risk / binary exit |
| --- | --- | --- | --- | --- |
| WF-07 / S-007 owner reconciliation | root; completed WF-06 foundation baseline, E-011–E-017 revalidated at entry | Testing Harness owner: identities, relationships, local clocks, privacy, bounds, incomplete evidence and acceptance map | V-007; record adoption and projection map before machine changes | Promoting mechanisms into authority; owner inputs/outputs/defaults/rejections stated once, with no unresolved owner contradiction |
| WF-08 / S-008 characterization | chain; WF-07 | Exact launch/lifecycle consumer and route inventory; independent fixtures on existing interfaces | V-008–V-011 baseline fixtures; checkpoint executed results/correction obligations | Golden preservation of defects; owner-aligned baseline passes and proposed outcomes have explicit expected fixtures |
| WF-09 / S-009 invocation context | chain; WF-08 | Runtime context, executor, Go/Vitest/shell row runners, browser and lifecycle adapters | V-008 and V-007 generation when inputs change; checkpoint migrated callers | Stale/inherited identity; all supported harness launch attempts have explicit unique context/outcome while failure channels remain private |
| WF-10 / S-010 allocation/process relationships | chain; WF-09 | Broker/provider acquisition entry, observation and retained correlation seams | V-009 plus applicable V-007; checkpoint ownership/failure parity | Changing cleanup authority or first-consumer attribution; relationship closure preserves physical ownership and cleanup |
| WF-11 / S-011 owner-local timings | chain; WF-10 | Go migration, browser reset, row/browser report parsing and catalog | V-010 plus applicable V-007; checkpoint local-clock and canonical parity evidence | Duplicate measurements or cross-clock arithmetic; timing/outcome fixtures pass with unchanged canonical accounting |
| WF-12 / S-012 retained analysis | chain; WF-11 | Secure validated loading, pure normalized analysis, bounded human/JSON projection | V-011 plus applicable V-007; checkpoint integrity/privacy/public outcomes | Unsafe reads or double counting; declared relationships close, shared/ambiguous work stays contextual, export remains isolated |
| WF-13 / S-013 validation/handoff | chain; WF-12 | Final source, generated outputs, owner routing, findings and durable handoff | V-012; final checkpoint with exact commands/runs/failures/skips | Overstated qualification; all selected obligations have current-source evidence and deferred work remains explicit |

Derived sequence: **WF-07 → WF-08 → WF-09 → WF-10 → WF-11 → WF-12 → WF-13**.

After completing each workstream, update this tracker **before beginning its
successor**. Record status, substantive edits, exact commands and run identities,
results including failed attempts, compatibility impact, unresolved risks and
successor readiness. A failed required gate blocks dependent work. Roll back a
slice as a coherent producer/consumer/authored-input/generated-output transaction,
preserving pre-existing changes and retained evidence. Do not erase a failed
attempt after a successful repair.

## 7. Remediation Slice Plan

### Completed foundation slices

The earlier implementation authorization and S-000–S-006 evidence remain valid
for that delivery. Their DONE state does not grant implementation authorization
for the new iteration. The table preserves their stable IDs and rollback rationale.

| Slice / kind | Findings / contracts / dependencies | Intended change and entry conditions | Readiness / validation | Completion / rollback |
| --- | --- | --- | --- | --- |
| S-000 specification/tracker | F-004/F-008/F-009; WF-00 | Adopt REQ-840/AC-137 before machine projection edits; preserve historical evidence. | DONE; S-000 Markdown gate | Owner/projection responsibilities and bounded scope agree; no generated changes in this slice. |
| S-001 characterization | F-001–F-006; C-001–C-005; WF-01 prerequisites | Extend owner-routed tests with independent accounting expectations, lifecycle interleavings, wrapper/check/export parity and vocabulary acceptance fixtures. Revalidate exact caller inventory at implementation entry. | DONE; S-001 checkpoint, V-002/V-003/V-005 | Fixtures specify expected outputs without calling the production algorithm to derive them; no production behavior change. Revert test-only slice if withdrawn. |
| S-002 shared definitions | F-004; C-001, C-003; WF-02 prerequisites | Add `tools/harness_instrumentation_definitions.json`, its owner-input schema, generated shared schema definitions and runtime descriptors; update authored attachments, generated policy and generator wiring as one transaction. | DONE; S-002 and S-006 generation closure, V-004 | Old valid/invalid fixture acceptance unchanged; no runtime AJV or second vocabulary registry. Roll back authored inputs and generated outputs together. |
| S-003 canonical accounting | F-003; C-001; WF-03 prerequisites | Change `canonical-unit-events.mjs` to normalized facts; adapt `performance-projection.mjs`, `writeCanonicalArtifacts` and `validateCanonicalRun` through `evidence-accounting/index.mjs`. Remove `retainProjection` and duplicate union code. | DONE; S-003 checkpoint, V-002 | Exact expected accounting and errors; no duplicate raw-event storage; bounded state documented. Revert all internal callers with the reducer; no compatibility overload. |
| S-004 collection lifecycle | F-001, F-002, F-006; C-002, C-003; WF-04 prerequisites | Add owner-local `collection-engine.mjs`, `resource-store.mjs`, `diagnostic-session.mjs`, and runtime `observation-gate.mjs`; thin controller/worker; inject Linux discovery and clock; move run composition behind session. | DONE; S-004 checkpoint, V-003 | Control races, shutdown, truncation and fencing tests pass; one worker and one writer remain; pre-scan order preserved. Revert engine plus adapters/wiring together; `off` disables optional collection. |
| S-005 retained analysis and deletion | F-005; C-003–C-005; WF-05 prerequisites | Separate secure loading, pure resource accumulation, explanation formatting and OTLP projection; update check/export and wrapper consumers; prune unused exports. | DONE; S-005 checkpoint, V-005/V-006 | Public outputs/exits/security unchanged, eager OTLP removed from checks, private dead paths gone. Revert consumer transaction without aliases or dual readers. |
| S-006 validation/handoff | All in-scope findings; WF-06 after WF-05 | Finalize, verify final source, reconcile failures and hand off deferred work. | DONE; final verification ledger | Required gates pass; historical evidence and default-off/deferred limits retained. |

### Selected execution slices

Implementation is authorized. The serial execution ledger records each completed
exit and the saved checkpoint before its successor. The table records current
readiness; historical planning authorization remains in Section 10.

| Slice / kind | Findings / contracts / dependency source | Intended change / entry gate | Current readiness / tests | Migration / rollback / binary exit |
| --- | --- | --- | --- | --- |
| S-007 specification amendment | F-010–F-013; C-006–C-009; WF-07 | Adopt owner semantics and acceptance mappings before schemas/producers; review exact definitions, policy and schema attachment changes | DONE; executed S-007 checkpoint and assigned validation | Owner amendments only in this slice; no generated changes. Exit: agreed inputs/outputs/defaults, privacy/bounds, failure and clock rules |
| S-008 characterization | F-010–F-012; C-001–C-009; WF-08 | Inventory direct/dynamic launch consumers and owner-routed tests; baseline failure envelopes, packed runners, allocation reuse, local measurements and explanation | DONE; executed S-008 checkpoint and assigned validation | Tests/fixtures and routing together; no defect blessed as compatibility. Exit: independent expected values and passing baseline, with corrections recorded |
| S-009 structural identity extraction and adopted extension | F-010; C-006; WF-09 | Introduce neutral runtime identity; migrate all selected launch adapters and failure-channel construction together | DONE; executed S-009 checkpoint and assigned validation | Revert context and callers together; no old-interface aliases. Exit: fresh parent-linked attempts across success/spawn failure/cancellation with unchanged public outcomes |
| S-010 adopted correlation extension | F-011; C-007/C-009; WF-10 | Create allocation identity before acquisition; observe leases/reuse/failures/process replacement without transferring lifecycle authority | DONE; executed S-010 checkpoint and assigned validation | Revert broker/providers/observation/schema callers together. Exit: exact relationship closure, shared/borrowed context preserved, required cleanup unchanged |
| S-011 adopted producer timing extension | F-012; C-008; WF-11 | Reuse migration/reset clocks and facts; add failure coverage and actual report parsing boundaries; closed catalog declarations | DONE; executed S-011 checkpoint and assigned validation | Producer/catalog/generated changes form one transaction. Exit: local timing fixtures pass; incomplete work stays incomplete; canonical summaries unchanged |
| S-012 retained reader/presentation extension | F-011/F-012; C-003/C-004/C-007–C-009; WF-12 | Validate observations and references before publishing normalized analysis; extend bounded public explanation | DONE; executed S-012 checkpoint and assigned validation | Incompatible formats cut over with all consumers; no compatibility translator. Exit: security, integrity, bounds, exact selection/exits and export-isolation fixtures pass |
| S-013 validation and handoff | F-010–F-013 and preserved foundation invariants; WF-13 | Complete V-012 on final source, classify failures, reconcile generated/import/documentation changes, save checkpoint | DONE; final S-013 checkpoint, current-source V-012/V-013 gates | No source rollback hidden inside validation. Exit: every selected gap has evidence; residual coverage/qualification limitations and next entry are recorded |

### Intended model and interface boundaries

- **Invocation:** one harness-owned launch attempt, distinct from a semantic unit,
  OS process or row. Allocate a fresh identity before attempting launch; a cache
  hit creates no runner invocation. Parentage is explicit. Packed invocations
  retain selected row membership without claiming row-level CPU attribution.
  Unit wrappers, Go/Vitest/shell/Playwright runners and selected lifecycle launch
  adapters are the covered producers; framework-internal subprocesses remain
  inferred unless explicitly registered by an owner.
- **Process:** physical identity uses existing boot/start/namespace proof. An
  invocation may own several processes. Physical facts and explicit or inferred
  attribution are separate; missing proof or a short-lived unsampled process is
  disclosed rather than invented. Conflicting evidence cannot silently overwrite
  an earlier owner or retroactively assign shared lifetime consumption.
- **Allocation and lease:** allocate opaque allocation identity before acquisition
  so failed attempts are representable. Reuse retains the allocation identity;
  leases identify separate consumer use. Shared processes belong to the physical
  allocation, not its first unit. Borrowed services remain contextual. Observation
  does not acquire, release, retry or otherwise control the resource.
- **Owner-local activity:** migration, fixture reset and report parsing are
  measured around actual owner calls, with explicit outcome and clock provenance.
  Reuse existing measurements; fill missing failure/cancellation coverage.
  Missing ends stay incomplete. Cross-clock subtraction and joining to scheduler
  time are deferred; incompatible/unavailable clocks yield unavailable durations.
  Overlap/shared work is disclosed and not summed into canonical cost.

Keep contexts and observations inside existing private runtime and artifact
ownership boundaries. The private failure envelope remains a failure channel,
with exact identity validation, parent-owned timeout/cancellation and removal;
do not retain it as another result summary. Optional diagnostic failure cannot
replace primary failure or required security/cleanup obligations.

Publish a bounded validated retained projection before the secret scan and prove
all its producers quiescent first. There is no diagnostics daemon, generic plugin
system, application SDK, automatic export or duplicate canonical event stream.
Normalize retained facts once; secure loading, pure analysis and formatting stay
separate. Ordinary inspection must not construct OTLP.

Extend the existing authored definitions catalog for approved activity vocabulary
and generated validators. Activity names do not implicitly extend canonical phase
semantics. Collection limits remain in their policy owner. S-007 declares exact
projection identities, required/nullable fields, caps and rejection behavior before
generation; this tracker does not invent an unadopted second schema registry.
Preserve unaffected formats. Version incompatible shapes and migrate their
producers/readers atomically, without legacy translation readers or aliases.

## 8. Validation Plan

### Foundation validation IDs — historical

V-001–V-006 retain their original meanings and execution references. Their commands
were exercised during foundation implementation as recorded in its checkpoints,
not rerun by this document update. Earlier planning-only skips remain in the
historical handoff. No foundation pass qualifies the new observations.

| Validation / layer | Findings / slices | Stage / exact commands | Expected evidence and failure gate |
| --- | --- | --- | --- |
| V-001 document and preservation | F-008; WF-00 | Planning-session: `git --no-optional-locks diff --check`; diff/status/index and content/identity comparison | Only this handoff differs; preserved historical ledger; all references and readiness links reviewed. Any unexpected change is reported, not reverted. |
| V-002 accounting and resource arithmetic | F-003; S-001/S-003 | Pre-move/post-slice: `make task-guide ROLE=module-author OWNER=harness.evidence_accounting`; `make harness-evidence-contract` | Independent synthetic expected values, tampered artifact rejection and large-stream bounds; failures block movement. |
| V-003 engine and integration | F-001/F-002/F-006; S-001/S-004 | Pre-move/post-slice: `make harness-evidence-contract` with new cases routed through existing imports | Injected interleavings plus real worker/thread-death and four-sibling quietness; artifacts in each exact retained run. No current baseline race failure is assumed. |
| V-004 definitions and generated drift | F-004; S-002 | Post-slice: `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make harness-evidence-contract` | Authored/generated agreement, dependency-free validators, equivalent acceptance and extension fixtures; stop on mixed generated state. |
| V-005 public consumer parity | F-005/F-008; S-001/S-005 | Pre-move/post-slice: `make task-guide ROLE=module-author OWNER=harness.command_surface`; `make harness-command-surface-contract`; `make harness-evidence-contract` | Wrapper/check/explanation/export arguments, outputs and exit codes match fixtures; delivery uses injected transport, no external export. |
| V-006 final regression / boundaries | S-005, C-001–C-005 | `make agent-finalize` before `make harness-contract`; `make lint-scripts`; relevant generation checks from V-004 | Current-source owner/import/schema checks pass. Without eligible `RESULTS_DIR`, report retained-run maintenance skipped. No baseline refresh. |

### New validation obligations

Future commands run from `/home/jochi/code/cartulary` through public Make with
the current pinned installed runtime. At implementation entry use
`make task-guide ROLE=module-author OWNER=harness.evidence_accounting` and
`make task-guide ROLE=module-author OWNER=harness.command_surface`, then resolve
affected runner/browser/fixture owner rows from current routing. Source filenames
are not test-owner IDs. Use `make test-slice OWNER=<resolved-owner> ROWS=<resolved-rows>`
or its service-backed equivalent after that discovery; do not invent row IDs.
New tests must be in executed owner suites, not merely present in the tree.

| Validation / layer | Findings / slices | Stage / commands | Required evidence / failure consequence |
| --- | --- | --- | --- |
| V-007 owner/projection closure | F-010–F-012; S-007 and later changed-input slices | Human owner review first; after generated-input changes: `make generate`, `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check` | Closed mappings and declared schema/policy metadata; valid/invalid corpus, synthetic extension and dependency-free validation. Owner or mixed-generation failure blocks dependents |
| V-008 launch identity | F-010; S-008/S-009 | Baseline/post-slice: `make harness-command-surface-contract`, `make harness-evidence-contract` and resolved runner owner slices | Unique nested/packed/repeated attempts; cache hits have none; spawn failure/timeout/cancellation/stale context handled; failure-channel scope/removal and public exits unchanged |
| V-009 ownership/correlation/lifecycle | F-011; S-008/S-010 | Baseline/post-slice: `make harness-evidence-contract` and resolved broker/browser/fixture slices | Shared allocation once, distinct leases, borrowed context, failed acquisition, replacement/PID reuse, missing proof, conflict and cleanup parity; preserve real thread-death/four-sibling quietness tests |
| V-010 local timing/accounting | F-012; S-008/S-011 | Baseline/post-slice: `make harness-evidence-contract` plus resolved Go fixture and row/browser adapter slices | Independent durations/outcomes for migration/reset/parsing; unavailable/incompatible clocks, incomplete ends, overlaps/nesting and shared work; canonical duration/path/cache/reservation results unchanged |
| V-011 retained/public/security | F-011/F-012; S-008/S-012 | Baseline/post-slice: `make harness-evidence-contract` and `make harness-command-surface-contract` | Dangling/forged/duplicate identities, malformed/oversized/symlinked/digest-mismatched artifacts and forbidden values rejected; bounds, human/JSON parity, exact selection, public exits and export isolation pass |
| V-012 final source/handoff | F-010–F-012; S-013 | `make agent-finalize` before broader checks; then `make harness-contract`, `make lint-scripts`, `make lint-markdown` and applicable V-007 drift checks; final `git --no-optional-locks diff --check` | Current-source routed evidence and final diff review; failed required gates block completion. Record exact source, environment, commands, run roots, failures, skips and residual limits |
| V-013 document/preservation | F-013; T-011 and S-013 | Historical E-018 was tracker-only. Execution: staged-byte identity, final Markdown/dependency/status review and staged/unstaged Git whitespace checks | Stable IDs and historical inventories/ledgers remain intact; current status and exact evidence reconcile without reusing historical acceptance |

Use hand-calculated expectations, injected clocks/operations and explicit
membership fixtures. Cover interrupted publication, cap exhaustion, rejected
writes and producers attempting to write after shutdown. Do not infer test
execution from names or derive expected output using the production algorithm.

Broaden service/browser execution only for an affected producer boundary or an
uncovered failure. Product-wide, visual and performance-qualification suites are
not initial gates. The changed Go fixture and browser adapter boundaries do
require their routed owner coverage; this is not a claim that pure Node tests
alone cover them. Preserve existing real-worker integration.

For V-012, pass `RESULTS_DIR` only for an eligible successful full warm `check`
run. Otherwise explicitly record retained-run maintenance skipped because
`RESULTS_DIR` is unset. No performance baseline refresh or optimization claim is
part of this iteration.

| Current-session command / inspected side effects | Disposition | Evidence and limit |
| --- | --- | --- |
| Git with `--no-optional-locks`, `rg`, file reads and in-memory SHA-256/lstat comparison | Executed; read-only | E-011–E-018; no harness runtime verification |
| `make lint-markdown` | Skipped | Authored target has Node/frontend readiness; wrapper writes static-analysis cache/stamp and retained run artifacts. Those effects exceed the tracker-only session boundary |
| Task guides, Make suites, generation and finalization | Skipped in this document update | Public wrappers can prepare/write artifacts; future implementation routing/gates are obligations, not current passes. No retained-run maintenance or new RESULTS_DIR |
| Qualification experiments and external export | Not in scope | T-008 remains deferred; no transport invoked |

## 9. Top-Level Work Tracker

Foundation completion is historical evidence at its recorded source. Original
delivery scopes remain distinguishable from the new slice IDs.

| Original delivery scope | Current disposition |
| --- | --- |
| 0 owner/projections | Foundation catalog/REQ-840 and selected REQ-841–844 contracts delivered |
| 1 timing/resources | Foundation DONE; selected runner identity and report detail delivered in S-009/S-011/S-012 |
| 2 services/browser | Selected allocation/migration/reset detail delivered in S-010/S-011; container totals deferred |
| 3 wrapper/comparison | Same-source comparison and graph pre-scan envelope delivered; full wrapper/post-scan and cross-clock joins deferred; successor waiver dropped |
| 4 qualification/guidance | Historical experiments retained; total observer and public-reader qualification deferred |

| ID | Work item / class | Workflow / dependency source | Status | Evidence / exit condition |
| --- | --- | --- | --- | --- |
| T-001 | Foundation planning reconciliation | WF-00; F-008 | DONE | E-001–E-010 apply to the prior planning session |
| T-002 | S-001 characterization | WF-01 | DONE | Foundation checkpoint and S-006 final-source confirmation |
| T-003 | S-002 definitions | WF-02 | DONE | Foundation checkpoint and S-006 final-source confirmation |
| T-004 | S-003 accounting | WF-03 | DONE | Foundation checkpoint and S-006 final-source confirmation |
| T-005 | S-004 lifecycle | WF-04 | DONE | Foundation checkpoint and S-006 final-source confirmation |
| T-006 | S-005 readers/deletion | WF-05 | DONE | Foundation checkpoint and S-006 final-source confirmation |
| T-007 | Deferred remainder of original coverage | F-007; original slices 1–3 | DEFERRED | Container counters, full wrapper/post-scan coverage and cross-clock joins. Selected identity/local-timing scope transferred to T-012–T-018; remainder is not completed |
| T-008 | Total observer and 0/1/4-reader qualification; default-on decision | F-007; original slice 4 | DEFERRED | Existing qualification protocol and complete observer accounting required; default stays off |
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

## 10. Session Handoff Log

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


### Current scope and authority

| Time / session | Baseline and state | Files / commands / evidence | Supersession / next action |
| --- | --- | --- | --- |
| 2026-10-09 read-only plan, followed by tracker update | `main` at `9359eb0cd359ab6bbc2eead16769950685d197b5`; clean entry | E-011–E-018; root instructions, framework, adopted owners and NLSpec reference | User selected identity/allocation correlation plus owner-local timing. This update changes the handoff only; S-007 is the next separately authorized task |

### Current backend and frontend boundaries

| Boundary | Observation / files | Current action / future gate |
| --- | --- | --- |
| Backend test support | `internal/testutil/pgtest/pgtest.go`; suite-service lifecycle/diagnostic producers; E-014 | Inspected measurement seams; no Go change. S-011 must use their owner-routed tests |
| Product backend / frontend | No product behavior, SQL migration, web component, selector or assertion move | Not applicable to implementation scope; no broad product test or visual pass inferred |
| Browser harness | Group runner/report adapter, startup and reset producers; E-012/E-014 | Selected future harness identity/timing work; preserve browser resource cleanup and public outcomes |

### Current contracts and generation

| Baseline | Current evidence | Files / action | Next gate |
| --- | --- | --- | --- |
| `9359eb0` | Adopted REQ-830–840 and private failure/fixture contracts; proposed C-006–C-009 | Inspected definitions/policy, attachments, resource schema and generator ownership; edited no owner or machine contract | S-007 declares projections and semantics; later generated changes use existing Make transaction and V-007 |

### Current tests and harness

| Baseline | Current evidence | Files / action | Next gate |
| --- | --- | --- | --- |
| `9359eb0` | E-012–E-016; assertions/scenarios inspected, no test execution | Foundation and shared-lease fixtures, command/routing metadata; historical passes retain source limits | S-008 revalidates exact consumers, resolves affected owner rows and executes independent characterization before movement |
| `9359eb0` | V-013 / E-018 | Read-only Git, content/identity snapshot, manual Markdown and link/dependency review | Complete document checkpoint; no artifact-producing Make command used |

### Current security and authorization

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

No user scope preference remains unresolved: identity/allocation correlation and
owner-local migration/reset/report timings are selected. No primary-owner
contradiction remains unresolved. Owner adoption and characterization passed
before dependent implementation. S-013 completed final-source validation and handoff.

| ID | Blocker / affected slices | Needed authority or evidence | Status / resolution |
| --- | --- | --- | --- |
| RB-001 | Historical missing foundation characterization, S-002–S-005 | S-001 baseline plus S-004/S-005 fixtures and final S-006 results | DONE; closed by recorded foundation execution, not reopened |
| RB-002 | New identity/activity/projection semantics not adopted; S-008–S-012 | S-007 owner amendment and reviewed C-006–C-009 inputs/outputs/defaults/bounds/acceptance mappings | DONE; S-007 adopted REQ-841–844 / AC-138–141 before projection changes |
| RB-003 | Independent pre-move characterization not executed for new launch/correlation/timing boundaries; S-009–S-012 | S-008 current-source V-008–V-011 baseline results and exact consumer/test routing inventory; explicitly assigned correction obligations | DONE; S-008 executed characterization and saved caller/routing inventory before movement |

T-007's deferred remainder and T-008 are not blockers to this selected iteration.
Missing host/container/short-lived-process coverage, sampled RSS limitations and
incomplete total observer accounting remain disclosed. No baseline refresh,
default-on adoption, concurrency tuning or performance improvement follows from
either this plan or its eventual implementation.

## 12. Binary Completion Criteria

| Outcome | Required/current statement |
| --- | --- |
| Planning completeness | complete for selected identity/allocation and owner-local timing scope; F-010–F-013 map to C-006–C-009, S-007–S-013 and V-007–V-013; adopted owner/detail/characterization work has executed exits |
| Per-slice readiness | S-007–S-013 DONE with saved serial checkpoints and completed-source validation; no selected successor remains |
| Implementation authorization | Current user request authorizes implementation through S-013. Prior document-only authorization belongs to the planning session |
| Foundation completion | S-000–S-006 remains DONE at its recorded source; its source changes, failures and results are preserved as history |
| Document completion / preservation | T-011/E-018 retain their historical tracker-only boundary and skipped lint. S-013 records current Markdown lint, staged-byte preservation, final changed-file inventory, evidence and status reconciliation |
| Selected iteration completion | S-007–S-013 and T-012–T-018 DONE after the final execution checkpoint below; current-source evidence closes the selected iteration |
| Coverage / qualification | T-007 remainder and T-008 DEFERRED, T-009 DROPPED, overall IN_PROGRESS and `HARNESS_DIAGNOSTICS=off` unchanged |

The ledger below is a verbatim historical foundation record. Its references to
“current execution” and “next entry” describe that completed session. The current
execution checkpoint is the identity iteration ledger below; E-018 remains the
historical document-only checkpoint.

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
