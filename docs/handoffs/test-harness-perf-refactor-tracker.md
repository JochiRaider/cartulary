# Test Harness Performance Refactor Tracker

## 1. Status, scope, and authority

Status: implementation complete; PERF-010 through PERF-080 are `DONE`.

This document is a non-normative execution handoff. It records retained evidence,
implementation candidates, adoption gates, and the minimum validation needed to
improve test-harness wall-clock time. It does not establish test-harness behavior.

The adopted [Testing Harness NLSpec](../testing-harness-nlspec.md) owns harness
behavior. Machine-readable contracts under `tools/` project that owner. If this
tracker conflicts with an adopted owner or its corrected projection, the owner
wins and this tracker must be updated.

The plan follows the
[Cartulary Modular Refactor Planning Framework](cartulary_modular_refactor_planning_framework.md)
and uses the precision, boundary, mapping, and acceptance-criteria practices in
the [NLSpec writing standard](../research/nlspec-spec.md). References to those
documents are for human planning only. Harness execution, conformance, and
release evidence must not read, hash, stat, or otherwise depend on Markdown.

The approved implementation request authorizes the serial workstreams recorded
here. This tracker itself remains non-normative and does not authorize:

- removal, skipping, weakening, or rerouting of any catalog row, tier, evidence
  route, retry rule, measurement rule, isolation rule, or release gate;
- automatic runtime tuning or publication of a normative performance baseline;
- repeated full-suite benchmarks or a P95 study.

The retained logs are sufficient diagnostic evidence for prioritization. Several
strong runs used dirty source states, and no retained run is asserted to match
current `HEAD`. The results in this document are therefore engineering baselines,
not formal TH-377--TH-380 performance publications.

## 2. Top-level work tracker

Allowed status values are `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, and
`DROPPED`. A workstream moves to `DONE` only when its implementation,
acceptance evidence, cleanup, and documentation obligations are complete.

| ID | Workstream | Status | Depends on | Real-target budget | Completion gate |
| --- | --- | --- | --- | --- | --- |
| PERF-000 | Evidence capture and implementation plan | DONE | None | No new product run | Baseline, bottlenecks, constraints, risks, and ranked work are recorded |
| PERF-010 | Owner and evidence-contract correction | DONE | None | Contract fixtures only | Adopted NLSpec has one coherent scheduler, browser, service-session, and event contract |
| PERF-020 | Cache and scheduler fast path | DONE | PERF-010 | One warm `make test-fast` | Contract simulation passes and warm target improves without semantic drift |
| PERF-030 | Explicit local service-session reuse | DONE | PERF-020 | One small attached service-backed slice | Attach/ownership/cleanup proofs pass and slice reaches the adoption range |
| PERF-040 | PostgreSQL and fixture throughput | DONE | PERF-030 | One matched capacity-8/capacity-12 `module.networkflow` pair | Typed policy and row audit pass; capacity 12 is retained only if it improves at least 15% without contention or isolation failure |
| PERF-050 | Performance-fixture critical path | DONE | PERF-040 | Narrow lint and architecture repair plus the exact owner slice | Semantic artifact, owner boundary, and prior performance decision remain valid |
| PERF-060 | Browser topology and utilization | DONE | PERF-050 | One `make browser-e2e-webserver-backed` | All current owner-derived rows pass, isolation holds, and wall time improves at least 20% |
| PERF-070 | Granular and artifact-bearing caches | DONE | PERF-060 | Contract fixtures plus one exact real-row miss/hit pair | Positive, negative, corruption, containment, and fresh-evidence matrices pass |
| PERF-080 | Convergence and release validation | DONE | All accepted workstreams | One `make check` and one warm `make release-check` | Exact closure and semantics hold; accepted improvements meet their gates |

`BLOCKED` here describes a planned dependency, not an active implementation
impasse. Only the one `IN_PROGRESS` row may receive structural implementation;
later rows remain dependency-blocked until their predecessor reaches a terminal
status and this tracker is updated and validated.

## 3. Workflow dependency map

```text
PERF-000 retained evidence and plan
    -> PERF-010 owner and evidence contracts
    -> PERF-020 cache and scheduler fast path
    -> PERF-030 explicit local service session
    -> PERF-040 PostgreSQL and fixture throughput
    -> PERF-050 performance-fixture critical path
    -> PERF-060 browser topology
    -> PERF-070 granular and artifact caches
    -> PERF-080 convergence and handoff
```

This order is intentionally serial. A terminal tracker update and its Markdown
and diff validation are required before the next workstream starts.

## 4. Source and implementation inventory

### 4.1 Owner and machine-contract inputs

| Path | Relevance |
| --- | --- |
| `docs/testing-harness-nlspec.md` | Adopted owner for execution, evidence, lifecycle, security, and performance semantics |
| `tools/execution_topology_manifest.json` | Authored target topology and resource profiles |
| `tools/scheduler_resource_registry.json` | Logical resource definitions and capacity resolution |
| `tools/harness_work_graph_owner.json` | Work-graph ownership, policies, and projections |
| `tools/harness_cache_registry.json` | Cache profiles, eligibility, inputs, and storage policy |
| `tools/browser_e2e_batch_manifest.json` | Browser rows, groups, stages, profiles, and estimates |
| `tools/performance_fixture_snapshot_owner.json` | Performance snapshot shape, contributions, and receipts |
| `tools/test_catalog_owner.json` | Active rows and verification metadata |
| `tools/test_families/**` | Verification-family routing and row projections |

Generated topology and schedule outputs remain projections. Implementers must
change owner inputs and use the owning generation or drift target; they must not
hand-edit generated files.

### 4.2 Scheduler, cache, browser, and fixture implementation

| Path | Relevance |
| --- | --- |
| `tools/harness/scheduler/work-graph/runner-cli.mjs` | Run setup, source snapshot, cache wiring, event retention, and summaries |
| `tools/harness/scheduler/work-graph/scheduler.mjs` | Readiness, resource fitting, admission, wait accounting, and completion |
| `tools/harness/scheduler/work-graph/cache.mjs` | Cache input digests, lookup, store, and record validation |
| `tools/harness/scheduler/work-graph/compiler.mjs` | Canonical units, grouping, compatibility, estimates, and resource claims |
| `tools/harness/scheduler/work-graph/browser.mjs` | Browser grouping and invocation boundary |
| `tools/harness/scheduler/fixture-broker/providers.mjs` | Fixture leases, owned service suite, and cleanup |
| `tools/testservices/main.go` | Service-suite start, readiness, descriptors, and teardown |
| `internal/testutil/pgtest/**` | PostgreSQL templates, databases, transaction fixtures, and cleanup |
| `internal/testutil/performancefixture*/**` | Cross-owner performance-fixture assembly and validation |
| `internal/modules/*/testsupport/performancefixture/**` | Source-owner fixture contributions and persistence adapters |

### 4.3 Retained evidence roots

All retained roots are beneath `.cartulary/test-results/`. They are observations,
not runtime dependencies. Implementers should use `make explain-run
RESULTS_DIR=<run-dir>` and retained summaries before scheduling another broad run.

| Run ID | Target/result | Role in this plan |
| --- | --- | --- |
| `20260816T213414Z-p492324` | `release-check`, pass | Matched warm release baseline |
| `20260816T205838Z-p187726` | `release-check`, fail | Same source/graph mixed-cache worst case |
| `20260816T204956Z-p4190652` | `check`, pass | Cold/edit-cycle baseline |
| `20260815T051959Z-p2970211` | `check`, pass | Directional warm check; source/graph differs |
| `20260815T063829Z-p3617456` | `test-fast`, pass | Warm cache/accounting baseline |
| `20260816T201227Z-p3682320` | `browser-e2e`, pass | Direct aggregate browser evidence |
| `20260816T202002Z-p3739653` | `browser-e2e-webserver-backed`, pass | Functional browser adoption baseline |
| `20260816T202455Z-p3821599` | `browser-e2e-stateful`, pass | Stateful isolation baseline |
| `20260816T202718Z-p3850837` | `browser-e2e-a11y`, pass | Accessibility baseline |
| `20260816T203531Z-p3917617` | `browser-e2e-visual`, pass | Visual baseline |
| `20260816T212622Z-p449485` | `browser-e2e-measurement`, pass | Measurement and snapshot baseline |
| `20260816T163014Z-p2241301` | Small service-backed slice, pass | Fixture-startup baseline |
| `20260816T175432Z-p2576514` | Small service-backed slice, pass | Fixture-startup baseline |
| `20260816T195230Z-p3265609` | Network Flow owner slice, pass | PostgreSQL workload evidence |
| `20260816T195511Z-p3311962` | Network Flow service-backed slice, pass | PostgreSQL capacity comparison baseline |
| `20260817T001900Z-p1027378` | Network Flow attached service-backed slice, pass | Matched current-source capacity-8 PERF-040 baseline |
| `20260817T002111Z-p1069227` | Network Flow attached service-backed slice, pass | Matched current-source capacity-12 candidate; rejected below threshold |
| `20260817T023037Z-p1648051` | `check`, fail | Related PERF-050 ownership/lint failure; repaired, retained, and excluded from completion evidence |
| `20260817T024813Z-p1881416` | `check`, pass | Final exact 776-unit/968-row closure |
| `20260817T025434Z-p1994125` | `release-check`, pass | Final warm exact 974-unit/1,202-row release closure |

## 5. Current execution architecture

The current execution path is:

```text
authored target, catalog, family, and topology inputs
    -> WorkGraphCompiler
    -> canonical graph and capacity snapshot
    -> resource-fit scheduler
    -> cache, fixture broker, and unit runners
    -> unit events and row results
    -> target summary and run summary
```

The v3 graph, resource-fit scheduler, fixture broker, deterministic ordering, and
browser batching already exist. This plan improves their hot paths and boundaries;
it does not propose another wholesale harness rewrite.

For each scheduler iteration, the current scheduler scans the graph for ready
units. For each ready unit that does not fit, it derives a holder snapshot from
the running set. A holder-set change can publish another `resource_wait` event.
Events are also retained in memory and appended synchronously to NDJSON. The cost
therefore grows with ready units, running holders, completions, and serialized
event volume rather than only with actual wait intervals.

The cache calculates broad input-root digests during lookup and calculates them
again during store. The runner has already created a source snapshot over about
3,581 executable inputs, but the useful per-file index is discarded after its
aggregate digest and file count are produced. Cache traversal is broader than the
executable-source boundary and can reach Markdown, which is not an acceptable
runtime dependency. Cache lookup also occurs after resource admission, so a
confirmed hit can consume logical CPU, PostgreSQL, browser, or fixture capacity
and participate in wait accounting despite doing no product work.

Go graph units are admitted against the captured CPU capacity. Compiler grouping,
however, does not include every exact compatibility partition named by the owner,
and package groups are invoked sequentially within a shard. Concurrent `go`
processes can each inherit host-level parallelism. The intended worker/process
budget therefore needs to be projected exactly before increasing Go concurrency.

Ordinary browser profiles currently claim four PostgreSQL lanes. With PostgreSQL
capacity eight, no more than two stacks fit even though browser, port, and service
stack capacities are four. Non-stateful groups commonly pay for separate stack
lifecycle work, and group estimates use a coarse evidence-class value instead of
the selected rows' summed estimates. Long groups can consequently start too late.

The managed service provider always starts and owns a suite. The service command
rejects nested active suites. The adopted owner describes attach semantics, but
the work-graph path does not yet offer an ergonomic, explicit local session that
developers can safely reuse across runs.

The performance snapshot constructs 20,000 Timeline rows through 40 batches of
500 via the application clipboard mutation path. It is both the largest known
uncacheable setup unit and a gate on measurement execution. Its current resource
profile reserves browser, port, service-stack, object-store, and four PostgreSQL
lanes even though its primary work is database fixture construction.

## 6. Log-derived baseline

### 6.1 Release and developer-loop results

| Evidence | Units | Wall time | Cache hit/miss/bypass | Selected detail | Conclusion |
| --- | ---: | ---: | ---: | --- | --- |
| Matched warm `release-check` | 974 | 907.217s | 790/6/178 | PostgreSQL saturated 535.262s; 161,357 waits; 219 MB events | Warm release is browser/PostgreSQL constrained, not generally CPU- or memory-constrained |
| Same-source/graph mixed-cache failed release | 974 | 1,603.003s | 689/107/178 | 519,072 waits; 731 MB events | Cache state plus harness accounting explains much of the approximately 30-minute case |
| Cold `check` | 784 | 507.405s | 0/739/45 | 276,470 waits; 526 MB events | Edit-cycle cost is dominated by broad invalidation, hashing, runnable work, and accounting |
| Comparable warm `check` | 766 | 92.706s | 715/6/45 | Different source and graph | Cache effectiveness is the largest developer-loop lever; comparison is directional only |
| Warm `test-fast` | 365 | 19.175s | 362/0/3 | 43,156 waits; 102 MB events | Cached units still reserve resources and create disproportionate accounting work |

The matched warm release reported 57.672s setup, 255.094s fixture work,
550.451s execution, 40.489s collation, and 3.508s wrapper work. Resource blocking
totaled 865.362s across units. The failed mixed-cache release reported 82.733s
setup, 248.045s fixture work, 1,223.395s execution, and 43.852s collation. These
phases overlap at the target level and must not be added to estimate a new wall
time.

The warm release event stream contained about 2.98 million holder references,
with up to 24 holders in one wait event. Its largest reason counts were CPU-only,
PostgreSQL-only, CPU plus object store, and object-store-only waits. The failed
release amplified this pattern to more than half a million wait events. Event
volume is therefore a bottleneck and an observability-model defect, not merely a
storage concern.

### 6.2 Catalog and fixture shape

| Dimension | Retained inventory |
| --- | --- |
| Active rows | 1,202 |
| Runners | 650 Go; 389 Vitest; 154 Playwright; 9 shell |
| Tiers | 577 fast; 385 standard; 228 full; 12 release |
| Evidence classes | 625 unit; 339 integration; 110 browser; 73 static; 24 visual; 13 accessibility; 9 security; 8 measurement; 1 release |
| Fixtures | PERF-040 audited current PostgreSQL closure: 252 `postgres_dedicated`; 81 transaction; 0 group; 8 migration; other retained counts in this table remain historical context |
| Service dependencies | 698 none; 351 PostgreSQL plus object store; 151 PostgreSQL; 2 object store |

PERF-040 audited all 341 current PostgreSQL rows. The three former group rows
required committed or cross-connection behavior but had no deterministic reset
and failure-contamination contract, so they moved conservatively to dedicated
scope. The 81 transaction rows are closed by an exact typed approval roster;
future shared-scope additions fail catalog validation until explicitly audited.

### 6.3 Browser and performance-fixture results

| Target or unit | Wall time | Relevant observation |
| --- | ---: | --- |
| `browser-e2e` | 448.978s | Effective peak was two browser stacks and eight PostgreSQL lanes |
| `browser-e2e-webserver-backed` | 290.536s | 62 graph units, 83 rows; 167.951s fixture and 107.269s execution |
| `browser-e2e-stateful` | 136.288s | Stateful isolation remains a hard boundary |
| `browser-e2e-a11y` | 82.928s | Separate evidence and artifact obligations remain |
| `browser-e2e-visual` | 99.064s | Visual isolation and artifacts remain |
| `browser-e2e-measurement` | 415.266s | Quiet measurement and snapshot construction dominate |
| Direct performance snapshot builder | 222.604s | Largest single known uncacheable setup unit |
| Snapshot builder during warm release | 263.070s | Contention extends the release critical path |

The browser manifest contains 26 webserver-backed groups covering 83 rows, five
support groups covering 12 rows, 13 stateful groups covering 15 rows, seven
measurement groups covering seven rows, two accessibility groups covering 13
rows, and two visual groups covering 24 rows. Startup, coarse estimates, quiet
locks, and effective two-stack concurrency dominate the warm release browser
path.

### 6.4 Service-backed slices and host resources

Small service-backed slices commonly completed in 21--26s, with about 18s spent
in fixture or service startup. The retained examples completed in 21.806s with
17.923s fixture time and 25.516s with 18.444s fixture time. An explicit local
session can therefore remove most latency after the session has been started,
without changing per-run data isolation.

The retained Network Flow owner slice completed 89 units in 151.577s. The
service-backed slice completed 47 units in 139.699s, including 13.282s setup,
51.697s fixture work, 60.693s execution, and 48.486s PostgreSQL saturation. This
is the selected single-run workload for a PostgreSQL capacity experiment.

The captured host exposed 24 CPUs and about 31 GiB memory. The warm release
briefly used all 24 CPU tokens but saturated them for only 54.499s, peaked at
6,912 MiB memory, and saturated PostgreSQL for 535.262s. Initial CPU, memory,
process, and I/O increases are not justified by this evidence.

## 7. Bottlenecks and constraint classification

### 7.1 Ranked bottlenecks

| Rank | Bottleneck | Evidence and likely cause | Primary workstream |
| ---: | --- | --- | --- |
| 1 | Cache misses and cache-path overhead | 92.706s directional warm check versus 507.405s cold check; repeated broad hashing and resource admission on hits | PERF-020, PERF-070 |
| 2 | Browser/PostgreSQL critical path | 535.262s PostgreSQL saturation; functional stacks claim half the capacity each | PERF-040, PERF-060 |
| 3 | Performance snapshot construction | 222.604s direct and 263.070s under release contention | PERF-050 |
| 4 | Scheduler/event amplification | 43,156 waits and 102 MB events for an almost entirely cached 19.175s run | PERF-020 |
| 5 | Repeated local service startup | About 18s setup in 21--26s small slices | PERF-030 |
| 6 | Dedicated-fixture distribution | 249 dedicated rows, three group fixtures, and sustained PostgreSQL pressure | PERF-040 |
| 7 | Coarse grouping and estimates | Separate browser lifecycles and evidence-class estimates extend the tail | PERF-060 |
| 8 | Broad invalidation and missing artifact restore | Local changes invalidate unrelated rows; release artifacts remain fresh work | PERF-070 |

### 7.2 Constraints that must be preserved

- Exact selected row, tier, owner, family, and evidence closure.
- Deterministic graph identity, capacity snapshot, ordering, and result collation.
- Existing failure visibility, failure precedence, diagnostics, and per-row
  outcomes, including when compatible rows share a process or affinity lane.
- Measurement exclusivity, quiet-state proof, one worker, 100 operations, no
  retries, and current measurement artifact semantics.
- Dedicated lifecycle for migration, process, cross-connection, committed-state,
  recovery, identity-sensitive, and other tests whose behavior requires it.
- Unique per-run databases, object-store namespaces, ports, runtime roots, leases,
  cleanup evidence, and zero cross-run state leakage.
- Security, owner-only permissions, secret redaction, path containment, and
  symlink rejection at artifact and session boundaries.
- Owned service mode for CI and `release-check`; ambient attached sessions cannot
  influence release evidence.
- Authored estimates and captured resource capacity. Runtime observations can
  inform a reviewed owner change but cannot silently retune a run.

### 7.3 Constraints that can change after owner adoption

- Event schema and internal ready, blocked, holder, and interval data structures.
- Cache indexing, digest memoization, lookup admission point, invalidation
  granularity, and safe artifact restoration.
- Compatible grouping, deterministic ranking, and duration estimates when exact
  result and isolation semantics remain unchanged.
- Local service-container lifetime through an explicit attach session.
- Logical PostgreSQL capacity and ordinary functional-browser resource claims
  after the specified safety gates pass.
- Fixture capability for rows proven safe under rollback or explicit group reset.
- Snapshot batching or source-owner persistence implementation when semantic
  receipts and artifacts remain identical.
- Ordinary functional-browser process affinity and reuse with a proven reset
  between groups.

### 7.4 Initial resource decisions

| Resource | Current retained capacity | Initial decision | Candidate experiment |
| --- | ---: | --- | --- |
| CPU | 24 | Keep | None until CPU becomes the measured critical constraint |
| Memory | 31,808 MiB | Keep | None; warm release peak was 6,912 MiB |
| Process | 96 | Keep | Enforce the corrected Go process budget within it |
| I/O | 48 | Keep | Reassess only after event-volume correction |
| PostgreSQL | 8 logical lanes | Keep | Capacity 12 was rejected in PERF-040 after a matched 2.584% improvement missed the 15% gate |
| Browser stack | 4 | Keep | Make four usable for compatible functional groups |
| Port | 4 | Keep | Preserve unique preallocated ports |
| Service stack | 4 | Keep | Preserve lifecycle and affinity boundaries |
| Object store | 4 | Keep | Remove false claims; do not raise capacity initially |

## 8. Owner contradictions and adoption blocker

PERF-010 is required before structural implementation because the adopted owner
currently contains conflicting instructions:

- Section 2.1 depends on the deleted historical
  `docs/handoffs/test-harness-bottleneck-refactor-tracker.md`. That tracker was
  removed in repository history and cannot remain a source of current
  requirements.
- TH-367 retains legacy FIFO reservation, browser-stage serialization,
  `test_slice`, and finalizer drain/order rules. Those rules conflict with the
  newer resource-fit backfill and ordinary browser-unit rules in TH-802 and
  TH-806 and with the newer timing semantics.
- `cartulary.harness_unit_event.v1` and its producer expose holder snapshots but
  do not encode the explicit eligibility, wait-start, and wait-end boundaries
  needed by TH-370 and TH-371.

The owner correction must be reviewed and adopted before projections or runtime
behavior change. The correction should:

1. Remove the dependency on the deleted tracker and incorporate any still-owned
   rule directly in the NLSpec.
2. Retire or replace the conflicting TH-367 execution rules so one scheduler and
   browser contract remains.
3. Define the local attached-service command and lifecycle boundary before the
   public targets are added in PERF-030.
4. Adopt `cartulary.harness_unit_event.v2` with explicit `eligible`,
   `wait_started`, and `wait_ended` boundaries and the closed wait reasons
   `resources`, `earlier_overlapping_ready`, `capacity`, and `scheduler_stop`.
5. Treat v1 only as archival retained evidence for manual inspection. Current
   commands must not read v1. All current producers, validators, summaries, and
   schemas must cut over atomically to v2; a run must not mix versions.

Owner-review acceptance is binary: the NLSpec has one answer for readiness,
reservation/backfill, browser serialization, finalizer order, event intervals,
and attached-service ownership, and every changed machine projection has an
identified generator or authored owner.

## 9. Detailed future workstreams

### PERF-010: Owner and evidence correction

Objective: make the adopted requirements internally coherent before changing
runtime structure.

Implementation outline:

- Prepare one owner change covering the stale dependency, superseded scheduler
  text, browser-unit rules, attached-session surface, and event v2 semantics.
- Map each corrected owner clause to its schema, registry, generator, contract
  fixture, producer, and consumer.
- Remove historical v1 readers from current commands. Retained v1 roots remain
  archival inputs for manual inspection only; do not retain dual producers or
  compatibility readers.
- Regenerate projections through their owning Make targets and prove drift clean.

Minimum validation:

- NLSpec and contract review resolves every contradiction listed in Section 8.
- Schema fixtures cover every event type, wait reason, unmatched interval,
  duplicate boundary, stop path, and deterministic ordering case.
- Harness contract checks prove that v2 closure derives the same or more precise
  timing and failure evidence; no product suite run is needed.

Risk control: do not encode implementation convenience as owner law. Eligibility,
dependency, failure, cleanup, and quiet-state boundaries must remain testable and
independent of a particular data structure.

### PERF-020: Cache and scheduler fast path

Objective: make cache hits cheap and make scheduler overhead proportional to
state transitions instead of graph size and holder churn.

Implementation outline:

- Build one immutable executable-source index per run. Derive every cache-profile
  digest from that index. The index must exclude Markdown and all other inputs
  outside the executable evidence boundary.
- Memoize each profile digest and reuse it for every lookup and store in the run.
- Probe cache eligibility after dependencies become eligible but before reserving
  product resources or fixtures. A confirmed hit restores/publishes its declared
  evidence and completes without consuming CPU, PostgreSQL, object-store,
  browser, port, or service-stack lanes.
- Represent waits with v2 interval boundaries. Maintain indexed ready and blocked
  sets plus incremental resource occupancy rather than repeatedly scanning the
  full graph and full holder set.
- Derive browser-group estimates by summing the selected row estimates.
- Restore exact Go compatibility partitions and the TH-379 budget: on the
  retained 24-CPU host, `min(group_count, clamp(floor(24/4), 1, 8))` gives at
  most six workers, each with `GOMAXPROCS=4`, subject to captured capacity and
  exact owner rules.

Expected impact: 30--70% off warm `test-fast`, tens of seconds or more off
`check`, at least 80% smaller unit-event artifacts, and 15--60s off warm release.
These ranges are estimates and not additive.

Adoption gate:

- Deterministic simulations cover ready/backfill ordering, dependency completion,
  a cache hit behind a dependency, misses, corruption, failures, cancellation,
  finalization, and every resource overlap.
- Cache-off and cache-hit simulations have identical selected-row closure,
  result identity, failure precedence, cleanup, redaction, and current-run
  evidence semantics.
- One warm `make test-fast` completes all 365 expected baseline units or the
  current owner-derived equivalent, produces valid v2 intervals, reduces event
  bytes by at least 80% against 102 MB, and falls within the estimated time range
  or records why another bottleneck now dominates.

The time range guides engineering judgment; semantic equivalence and event-volume
gates are mandatory. Do not repeat the target solely to establish a percentile.

### PERF-030: Explicit local service-session reuse

Objective: amortize container and migrated-template startup for local developer
slices while retaining per-run data isolation.

Adopted public surface:

- `make test-services-session-up`
- `make test-services-session-status`
- `make test-services-session-down`

The default descriptor should be
`${CARTULARY_MACHINE_CACHE_DIR}/test-services/session.json`. It must be outside
retained result roots, owned by the invoking user, mode `0600`, safely created,
and treated as secret-capable. `CARTULARY_TEST_SERVICES_MODE=attach` selects
attach mode. An optional absolute `CARTULARY_TEST_SERVICES_SESSION_FILE`
overrides the default after validation and containment checks. Externally
supplied `CARTULARY_TEST_SERVICES_ACTIVE` is rejected; it remains private
child-process state only.

Attach mode borrows containers and migrated templates. Each harness run still
creates unique databases, object-store namespaces, ports, runtime roots, leases,
cleanup receipts, and schema-digest validation. The run closes only resources it
owns. Session-down closes only the exact session described by a validated lease.
CI and `release-check` remain owned-mode only and must reject an ambient attach
request or active descriptor.

Expected impact: reduce a typical 21--26s small service-backed slice to about
3--8s after the explicit session is ready.

Minimum validation:

- Contract fixtures cover missing, stale, wrong-owner, wrong-mode, malformed,
  mismatched-schema, dead-container, and concurrently used descriptors.
- One small attached service-backed slice proves readiness, unique resource
  identities, current-run evidence, cleanup of run-owned state, and survival of
  borrowed session resources.
- Owned-mode fixtures prove teardown and parent-death cleanup. Release-mode
  fixtures prove attach rejection.

### PERF-040: PostgreSQL and fixture throughput

Status: `DONE`, including the final-validation repair.

The v6 scheduler resource registry now owns typed automatic capacity policies
and bounds. Runtime hard-coded lane defaults were removed. PostgreSQL remains at
eight lanes by default; the candidate declaration is a schema-validated test
fixture and cannot exceed the detected CPU policy bound.

The current manifest contains 341 PostgreSQL rows: 252 dedicated, 81
transaction, zero group, and eight migration. The typed PostgreSQL fixture
policy registry maps each capability to its isolation rationale, closes the
exact transaction-row approval roster, and rejects stale or surplus shared
scope. The three previously grouped rows moved to dedicated scope because they
exercise committed or cross-connection behavior without a complete reset and
failure-contamination contract. Their obsolete group helper path was removed;
future group rows fail closed until the complete adopted contract exists.

Performance-fixture construction now has an independent builder policy and
`performance_fixture_builder` topology profile. Its graph unit claims bounded
CPU, I/O, memory, and process capacity plus one PostgreSQL lane. It does not
claim browser, port, or object-store resources inherited from the measurement
consumer.

The matched cache-off attached Network Flow comparison used source digest
`sha256:4d89047825492c2ea3c08bdd598c52fbe073f5391bd9d0ff05e9972041c70627`,
workload digest
`sha256:c86686a3c1a9e55bf99b40eca6a41c6556d9813215a90d6f48ee089e42a50d02`,
24 CPU tokens, the same explicit service session, 45 units, and 33 rows:

- capacity 8 passed at `20260817T001900Z-p1027378` in 125.094s, with
  PostgreSQL peak 8 and saturation 44.334s;
- capacity 12 passed at `20260817T002111Z-p1069227` in 121.861s, with
  PostgreSQL peak 12 and saturation 26.916s.

The 3.233s or 2.584% wall-time improvement missed the required 15% gate. The
candidate also increased peak browser stacks from two to three, peak modeled
memory from 1,792 to 2,048 MiB, and the dependency critical path from 54.647s to
55.477s. Capacity 12 was therefore rejected and the clean typed default of eight
was retained. Both valid runs passed all 45 units, emitted exact 426-event
streams of 183,463 and 183,447 bytes, bypassed cache for all units as required,
passed retained-secret scans over 530 files, returned the borrower count to
zero, and left cleanup to the exact session-down command.

### PERF-050: Performance-fixture critical path

Objective: reduce the 20,000-row snapshot build while preserving its exact
semantic artifact and quiet measurement boundary.

Implemented outcome:

- Added the redacted
  `cartulary.performance_fixture_build_diagnostics.v1` current-run artifact.
  Per-contribution and batch timings remain outside the snapshot receipt,
  snapshot key, and semantic digest.
- Rejected the 2,000-row batch candidate because the existing tabular mutation
  contract correctly rejects batches above 500. No compatibility exception was
  added.
- Added a Timeline-owned set-oriented construction boundary spanning owner
  application validation, record persistence, workbook projection persistence,
  and owner read-back validation. It preserves the production store and
  transaction boundaries without exposing cross-owner raw SQL.
- Preserved exactly 20,000 Timeline rows, stable contribution identities, the
  complete relationship and condition set, the semantic digest, measurement
  quietness, browser predicate, cleanup, and redaction behavior.
- Repaired the atomic event writer's same-run finalizer boundary: finalizers may
  read only the validated run-contained staging stream, while the canonical v2
  stream is still published atomically after terminal scheduler completion.

Adoption evidence:

- The exact cache-off Timeline measurement row passed at
  `.cartulary/test-results/20260817T005901Z-p1240805`, with all 14 units, one
  selected row, 122 v2 events, and 44,691 event bytes.
- The builder completed in 100.515s versus the directional 222.604s reference,
  a 122.089s or 54.846% improvement. Timeline's set-oriented contribution took
  8.610s; Entities is now the dominant contribution at 90.939s.
- Semantic digest
  `sha256:b9549c30dab549504fd945fc211da7649223573e8ea67d57b7bb17a2ff06b5a7`
  matched the earlier successful build and the independent two-build owner
  integration proof. The measurement aggregate qualified with 100 samples,
  zero retries, and complete owned-resource cleanup.

PERF-050 is complete. The additive diagnostics contract does not alter product
APIs or fixture meaning. Transitive testservices build invalidation remains a
PERF-070 dependency-closure concern; the retained broad fallback is safe but can
temporarily reuse a stale helper binary until that workstream closes the input
model.

### PERF-060: Browser topology and resource utilization

Status: `DONE`.

The adopted owner now defines ordinary functional browser lanes separately from
stateful, measurement, accessibility, and visual isolation. The typed
`browser_functional` profile claims two PostgreSQL tokens, so the retained
capacity of eight admits four compatible stacks. The compiler partitions by
runtime profile, calculates every group cost from the sum of selected row
estimates, and applies stable longest-processing-time-first placement with
group-ID and lane-ID tie breaks. The current webserver-backed closure is 26
groups and 83 rows across four lanes: three default-profile lanes and one
`network_flow_claimed` lane.

Each lane retains only its process stack. Every group is a separate Playwright
invocation with fresh context, output root, row closure, and test-route
generation. Each of the 22 inter-group reset units restores the lane's exact
database to the proved semantic baseline without changing its immutable DSN,
recreates its exact object-store namespace, resets application state through the
authenticated test route, and clears Playwright state. Product failures are not
retried; reset remains finalizable. Reset failure quarantines the allocation,
and a later group receives a new concrete session identity rather than reusing
conflicting evidence.

The first full run at `20260817T012414Z-p1302576` is invalid adoption evidence.
All 26 product groups passed, but physical database replacement terminated the
warm backend connection; the 22 resets failed and the run ended after 39/62
units in 267.660s. The physical-drop candidate was removed. The second full run
at `20260817T013614Z-p1401118` is also invalid: all 22 in-place resets passed,
but every group failed before execution because the new concrete session ID was
not exported by the fixture provider. It ended after 35/62 units in 197.629s.
Both failures proved reset finalization, quarantine, fresh replacement
allocation identities, and continued independent work; neither is averaged
into the performance result.

After correcting the provider boundary, the focused cache-off Auth browser row
passed 12/12 units at `20260817T014148Z-p1488937`. The accepted full
`make browser-e2e-webserver-backed` passed at
`.cartulary/test-results/20260817T014259Z-p1516987`: 62/62 units, 26/26 groups,
and the exact 83-row current closure in 165.251s. This is 125.285s or 43.121%
faster than the directional 290.536s reference and exceeds the 20% adoption
gate. Cache was 0/0/62, the v2 event stream was 604 events and 328,954 bytes,
capacity was four browser/object-store/port lanes and eight PostgreSQL tokens,
and peak use was exactly four lanes and eight PostgreSQL tokens. All 22 reset
statuses were HTTP 200; all 52 broker leases were released; four unique
database, bucket, backend-port, frontend-port, runtime, and concrete session
identities were retired; all eight retained process identities were stopped;
there were no terminal resource holders; and the retained-secret scan passed
over 752 files. No measurement, stateful, accessibility, visual, retry, or row
semantics changed.

### PERF-070: Granular and artifact-bearing caches

Objective: prevent localized edits from invalidating unrelated rows and safely
restore expensive declared artifacts.

Implementation outline:

- Derive conservative Go-package dependency closures and TypeScript-workspace
  dependency closures from executable owner inputs. Fall back to the current
  broad profile whenever a complete closure cannot be proven.
- Version cache records that contain artifacts. Record declared artifact type,
  relative path, mode, digest, producer identity, and semantic input digest.
- Restore only beneath a validated current-run destination. Reject absolute
  paths, traversal, undeclared files, type changes, special files, links,
  symlinks, mode escalation, digest mismatch, and partial records.
- Regenerate current-run evidence, summaries, timestamps, and run identities even
  when immutable product artifacts are restored.
- Consider build, license, and SBOM artifacts only after their complete inputs and
  freshness rules are modeled. Browser, migration, lifecycle, stateful,
  measurement, and other freshness-sensitive work remains uncached.

Minimum validation uses small contract fixtures, not duplicate full-suite runs:

- Positive matrix: unrelated edit stays a hit; declared dependency edit misses;
  toolchain, environment, owner, schema, selector, and generator changes miss.
- Negative matrix: missing dependency knowledge falls back broad; deleted,
  renamed, or newly discovered executable inputs invalidate safely.
- Corruption matrix: content, digest, metadata, mode, path, symlink, partial
  record, and producer mismatch all miss closed without publishing artifacts.
- Artifact matrix: cold output and restored output have identical declared bytes,
  type, mode, semantic digest, and release consumption, while evidence belongs to
  the current run.

Expected impact: localized edits stop invalidating hundreds of unrelated rows;
artifact restore can remove 60--120s of warm release work. The actual gain depends
on proven closure coverage and which artifact producers remain eligible.

Outcome: adopted cache-record v2 under a distinct
`.cache/cartulary/graph-v2` root, conservative unit-aware dependency closures,
immutable content-addressed entries, and transactional typed artifact restore.
Go closures contain exact selected packages and repository-local transitive
imports; TypeScript closures contain selected workspaces and declared,
referenced, or relative transitive repository workspaces. Both include current
owner/selector contracts, tracked and relevant untracked inputs, manifests,
lockfiles, toolchain, schemas, helpers, configuration, and generator identities.
Any incomplete selection or dependency proof falls back to the registered broad
closure. Resolution results are memoized only within one immutable source index
and are shared by package/workspace plus owner rather than rescanned per row.

The v2 entry manifest closes artifact type, normalized relative path,
destination class, mode, digest, producer, semantic input, and exact directory
members. Entries are written through exclusive staging and a valid concurrent
entry wins without deletion or replacement. Lookup rejects absolute and
traversing paths, undeclared or surplus payloads, wrong destinations, producers,
types, modes, digests, symlinks, hard links, special files, unsafe ancestry, and
partial records. Invalid entries are quarantined; restore validates and stages
the complete set, publishes all-or-none with rollback, and regenerates row and
unit evidence with current timestamps. Service-backed, stateful, migration,
browser, measurement, security, lifecycle, drift, cleanup, and publication work
is non-cacheable. Build, license, and SBOM producers remain disabled because no
complete work-graph reusable-output/freshness declaration was proven; this is an
intentional safe outcome rather than an incomplete candidate.

The contract matrix covers add, delete, rename, transitive Go and workspace
changes, manifests, lockfiles, toolchain, schemas, untracked executable inputs,
broad fallback, immutable concurrent writers, corrupt entries, absolute and
traversing paths, symlink/hard-link/FIFO attacks, modes, missing and surplus
payloads, producer and destination mismatch, directory members, partial restore
rollback, and fresh evidence replay. Final `make harness-contract` passed at
`.cartulary/test-results/20260817T022430Z-p1633453`. The exact current row
`platform.config.unit.catalog_snapshot_lifecycle` then passed on a v2 miss at
`.cartulary/test-results/20260817T022445Z-p1634018` in 1.419s and on a v2 hit at
`.cartulary/test-results/20260817T022453Z-p1634416` in 0.339s. Both used source
digest `sha256:2fec3d94faaca204d82dba79ec0c96fe868fb223d41ad7461347328210d010dd`,
capacity 24 CPU/48 I/O/96 processes, and the exact one-unit/one-row closure. The
miss emitted 12 events/3,088 bytes and one admission; the hit emitted 10
events/2,741 bytes, a fresh `2026-08-17T02:24:55.477Z` row timestamp, and no
admission or held resource. No service, fixture, secret, cleanup, or redaction
surface participated in this exact pure-unit cache proof.

### PERF-080: Convergence and release validation

Status: `DONE`.

Objective: prove the accepted workstreams compose without weakening coverage,
determinism, isolation, diagnostics, or cleanup.

Before broad validation:

- Complete every workstream's contract fixtures and single candidate run.
- Regenerate owned projections and pass all relevant drift checks.
- Compare selected catalog rows, tiers, evidence routes, graph digests, cache
  dispositions, retries, failures, redaction, and cleanup receipts with the owner.
- Record accepted and rejected candidates in the execution log in Section 12.

Then run, in order:

1. `make agent-finalize`, passing the successful retained result root when the
   target supports it.
2. One `make check`.
3. One warm `make release-check`.

Do not run a second release solely to estimate variance. Use the formal
TH-377--TH-380 window once only if the team later chooses to publish a normative
performance baseline.

The directional outcome, if the first six workstreams pass their gates, is a
9--12 minute warm release versus the retained 15m07.217s baseline. Workstream
ranges overlap and are not additive. No cold-release target is claimed until the
targeted evidence identifies its remaining critical path.

The completed current-source validation used source digest
`sha256:0bc146b0c0d6b4a4c5af339703d87f18a41c0b998c83b77a90ac303a6b1fa01d`.
The final `make check` passed at
`.cartulary/test-results/20260817T024813Z-p1881416`: 776/776 units, exact
968-row closure, 358.508s wall time, cache 398/66/312 hit/miss/bypass, and a
5,954-event/2,327,944-byte v2 stream. Its retained-secret scan passed over
6,251 files, its resource peaks remained within the captured capacities, and
its cleanup and run terminal boundaries both passed.

The immediately following warm `make release-check` passed at
`.cartulary/test-results/20260817T025434Z-p1994125`: 974/974 units, exact
1,202-row closure (`go` 650, Playwright 154, shell 9, Vitest 389), 1,094.157s
wall time, cache 456/34/484 hit/miss/bypass, and a
7,724-event/3,260,645-byte v2 stream. All 974 units have one queued, eligible,
wait-started, wait-ended, and completed boundary; 518 executed units have one
admission/start pair, and 456 hits have neither. Fixture acquire/release counts
both equal 420. The retained-secret scan passed over 8,766 files; release,
readiness, inventory, failure-finalizer, cleanup, and run-completion evidence
all passed. The owned service teardown's last span passed, no scope-labeled
container or event staging file remained, and exact browser/database/bucket
identity and collision checks passed through their release rows and finalizers.
Deterministic contract fixtures supplied fail-fast/finalizer evidence; no
second deliberately failing release was manufactured.

The release result is 186.940s (20.6%) slower than the historical 907.217s
directional reference, so the projected 9--12 minute outcome is rejected and no
performance improvement is claimed for the aggregate release. That historical
root differs in source and cache conditions and is not a formal matched
baseline. The current release's PostgreSQL and object-store saturation
(524.709s and 389.897s) remains a follow-up performance risk, but it does not
invalidate the completed correctness, isolation, security, cleanup, or evidence
gates and does not authorize another broad run. No duration baseline was
published or changed.

## 10. Local and release optimization boundaries

| Optimization | Local developer workflows | Release/CI validation |
| --- | --- | --- |
| Cache-hit admission and scheduler/event fast path | Enabled after contract gate | Enabled after same gate |
| Explicit attached service session | Opt-in and visible | Forbidden; owned mode only |
| PostgreSQL capacity 12 | Adopted only after slice gate; same authored default where applicable | Adopted only if owner and release resource model approve it |
| Dedicated-to-transaction fixture conversion | Available to proven rows | Same exact row contract; no local-only weakening |
| Browser affinity lanes | Useful for direct functional target | Same deterministic groups; quiet/stateful exceptions preserved |
| Performance snapshot optimization | Speeds exact measurement rows | Same real snapshot and semantic receipts |
| Granular row caches | Faster localized edit loop | Same conservative closure and miss-closed behavior |
| Build/license/SBOM artifact cache | Optional only after proof | Enabled only after freshness and release-consumption matrices pass |

Local speed must not create a second coverage definition. The only local-only
feature is the explicit lifetime of borrowed service containers and templates;
selected rows and their isolation remain identical.

## 11. Risk register

| Risk | Trigger | Mitigation and proof | Owner workstream |
| --- | --- | --- | --- |
| Lost or inflated wait visibility | Event compression or admission changes | v2 eligibility/start/end closure, deterministic replay, unmatched-boundary rejection | PERF-010, PERF-020 |
| Stale cache result | Incomplete source closure or memoization error | Cache-off comparison; positive/negative invalidation; broad fallback | PERF-020, PERF-070 |
| Cache poisoning or unsafe restore | Corrupt record, traversal, link, mode, or secret | Digest/type/mode validation, containment, symlink rejection, redaction scan | PERF-070 |
| Database races or exhaustion | Capacity 12 or narrower claims admit too much | Single capacity gate, connection metrics, DDL checks, cleanup and contamination proof | PERF-040 |
| Port or filesystem collision | Four browser lanes or reused process | Preallocated unique ports, per-run/lane roots, collision fixtures | PERF-060 |
| Shared browser state | Affinity reuse crosses group boundary | Complete reset after success/failure; stateful and measurement exclusions | PERF-060 |
| Failure masking | New groups report only process failure | Exact per-row results, stable failure precedence, retained diagnostics | PERF-020, PERF-060 |
| Snapshot semantic drift | Larger batches alter order, receipts, or mutation behavior | Exact semantic digest, counts, identities, receipts, and validation queries | PERF-050 |
| Nondeterministic auto-tuning | Runtime measurements change current schedule | Authored estimates, captured capacity, stable LPT and tie-breaks | PERF-010, PERF-060 |
| Local session leaks or ownership error | Stale descriptor or indiscriminate teardown | Owner/mode/lease validation, exact session-down, borrowed-resource rules | PERF-030 |
| Cleanup gap after cancellation | Fast path bypasses broker/finalizer | Cancellation and partial-start fixtures; idempotent reverse-order cleanup | All structural work |
| Resource regression hidden by overlap | Phase totals misread as wall time | Compare target wall time and critical resource intervals, not summed phase durations | PERF-040--PERF-080 |
| Excessive validation cost | Broad targets repeated per small change | Contract simulation first; one specified real target at each checkpoint | All workstreams |

## 12. Execution protocol and handoff records

### 12.1 Required protocol for each implementation session

1. Confirm the active owner text and this tracker's dependency state.
2. Run `make task-guide ROLE=module-author OWNER=<owner-id>` before any owner
   slice.
3. Use `make explain-test-owner`, `make explain-target`, `make explain-run`, and
   retained artifacts to narrow the candidate before running it.
4. Implement one coherent checkpoint and run deterministic simulations or
   contract fixtures first.
5. Spend the workstream's one real-target budget only after the cheap checks pass.
6. Record command, result root, exact selected closure, before/after wall time,
   resource peaks/saturation, cache counts, event bytes, cleanup outcome, and
   adoption decision.
7. Revert or correct a candidate that misses a mandatory safety gate. Do not hide
   a miss by changing coverage, retry, measurement, or isolation semantics.

### 12.2 Real-target run budget

| Order | Checkpoint | Permitted real target | Baseline/adoption comparison |
| ---: | --- | --- | --- |
| 1 | Cache, scheduler, event, and Go corrections | One warm `make test-fast` | 19.175s, 102 MB events, 365 baseline units |
| 2 | Attached local service session | One small attached service-backed slice | 21--26s total, about 18s startup; target 3--8s |
| 3 | PostgreSQL capacity and fixture profile | One matched capacity-8/capacity-12 `module.networkflow` pair | Same current source and service state; require at least 15% improvement |
| 4 | Snapshot batching/persistence | One exact Timeline measurement row | 222.604s builder; require at least 20% improvement |
| 5 | Functional browser topology | One `make browser-e2e-webserver-backed` | 290.536s and 83 rows; require at least 20% improvement |
| 6 | Accepted-work convergence | One `make check`, then one warm `make release-check` | 507.405s cold check context; 907.217s warm release |

Repeated executions are allowed only when the first execution is invalid because
the harness did not run the intended candidate, the target failed before reaching
the measurement boundary for an unrelated environmental reason, or a correctness
defect must be verified after repair. Record the reason; do not average the runs.

### 12.3 Implementation-session record template

```text
Date and implementer:
Workstream and tracker status:
Owner revision/digest:
Source revision and dirty-state summary:
Paths changed:
Contract/simulation commands and results:
Real target command, if budgeted:
Retained result root:
Selected row/tier/evidence closure:
Baseline wall time and candidate wall time:
Cache hit/miss/bypass:
Resource capacity, peak, saturation, and waits:
Event artifact bytes and interval validation:
Isolation, cleanup, redaction, and failure evidence:
Adoption decision and gate rationale:
Remaining blocker or next workstream:
```

## 13. Acceptance criteria

The performance refactor is complete only when all of the following are true:

- [x] PERF-010 is reviewed and adopted; no active requirement depends on a
  deleted handoff or conflicts with another scheduler/browser clause.
- [x] Current event producers and consumers use one adopted v2 schema, and every
  eligible wait has exact, closed, deterministic interval evidence.
- [x] Cache hits do not reserve product resources or fixtures and do not weaken
  dependency, artifact, or failure semantics.
- [x] Source and cache indexing has no Markdown/documentation dependency and does
  not reread unchanged executable inputs per lookup/store.
- [x] Go grouping and worker/process budgets exactly project the adopted owner.
- [x] Local service attachment is explicit, owner-only, safe against stale or
  malicious descriptors, and impossible in release/CI mode.
- [x] Every PostgreSQL capacity or fixture-class change has connection, cleanup,
  and contamination evidence.
- [x] The performance snapshot preserves exact semantic identity and measurement
  behavior.
- [x] Browser lanes preserve every row outcome, artifact, isolation boundary,
  reset, retry rule, and failure path.
- [x] Granular and artifact caches miss closed for incomplete knowledge,
  invalidation, corruption, traversal, link, and mode hazards.
- [x] Each completed checkpoint meets its stated wall-time gate or records an
  explicit evidence-backed decision to retain the old behavior.
- [x] Final `make check` and warm `make release-check` pass once, with exact
  owner-derived closure, valid evidence, zero leaked resources, and no loss of
  failure visibility.
- [x] The implementation log names every retained result root and distinguishes
  diagnostic estimates from any formally published baseline.

## 14. Current execution-step record

| Field | Value |
| --- | --- |
| Scope | PERF-010 through PERF-080 completed serially; no workstream is active or blocked |
| Baseline source | Existing retained logs listed in Section 4.3 |
| Executable changes | PERF-020 scheduler/cache/runtime corrections, PERF-030 explicit local service sessions, PERF-040 typed PostgreSQL capacity/fixture policy, PERF-050 owner set-oriented fixture construction, PERF-060 warm browser lanes, and PERF-070 typed dependency/artifact caches are complete; PERF-080 added no product behavior |
| Generated-artifact changes | Regenerated through `make generate`; no generated file was hand-edited |
| Test, coverage, retry, or release-gate changes | Contract fixtures expanded; selected row closure, retry policy, and release gates are unchanged |
| Required verification | Complete: every workstream's named gates, serial tracker checkpoints, final generation/contract gates, `agent-finalize`, `check`, warm `release-check`, Markdown lint, and diff validation |
| Product-suite verification | PERF-020 warm `make test-fast` passed at `.cartulary/test-results/20260816T231242Z-p823738`; PERF-030 attached row passed with exact per-run cleanup at `.cartulary/test-results/20260816T234959Z-p899374`; PERF-040 matched Network Flow runs passed at `.cartulary/test-results/20260817T001900Z-p1027378` and `.cartulary/test-results/20260817T002111Z-p1069227`; PERF-050 exact Timeline measurement passed at `.cartulary/test-results/20260817T005901Z-p1240805` and its final ownership repair revalidation passed at `.cartulary/test-results/20260817T024127Z-p1837874`; PERF-060 exact webserver-backed browser closure passed at `.cartulary/test-results/20260817T014259Z-p1516987`; PERF-070 exact pure-unit cache miss/hit pair passed at `.cartulary/test-results/20260817T022445Z-p1634018` and `.cartulary/test-results/20260817T022453Z-p1634416`; final `check` passed at `.cartulary/test-results/20260817T024813Z-p1881416`; final warm `release-check` passed at `.cartulary/test-results/20260817T025434Z-p1994125` |

PERF-010 adopted the owner correction before structural scheduler work. PERF-020
then completed the cache-before-admission, incremental scheduler, bounded event
writer, current-run replay, executable-source index, and Go execution-budget
changes. PERF-030 added explicit persistent local service sessions with secure
descriptors, live borrower leases, redacted status, owned-only target enforcement,
and per-borrower cleanup. PERF-040 moved capacity policy into typed ownership,
closed all current PostgreSQL row classifications, removed the unproven group
path, and separated builder resources from browser consumption. PERF-050 added
separate redacted build diagnostics and an owner set-oriented Timeline fixture
boundary while preserving semantic identity and measurement behavior. PERF-060
added deterministic four-lane ordinary browser scheduling, explicit resets, and
warm process reuse with fresh group state. PERF-070 added conservative typed
dependency closures and immutable all-or-none artifact caching while leaving
unproven artifact producers disabled. PERF-080 reconciled the specification,
projections, implementation, tests, and developer guide; repaired the related
Timeline ownership defect discovered by the first broad run; and closed the
current exact check and release closures. The handoff is complete with no active
or blocked mandatory workstream.

## 15. Implementation log

| Timestamp | Workstream | Source identity | Change and evidence | Decision | Next action |
| --- | --- | --- | --- | --- | --- |
| 2026-08-16T18:40:20-04:00 | PERF-010 activation | `HEAD 03a3505a6a86e3ff6286e572acbdbc69ad5dd4d8`; staged tracker-only diff SHA-256 `d03d302c98d4eae30d68b50dfa42b7fb8958b6ecee418ce2d396b4004a492c6e` | Reconciled the approved serial execution order and hard v2 cutover with this controlling tracker. No runtime, contract, generated, or product file changed at activation. | PERF-010 is the only `IN_PROGRESS` slice; PERF-020 through PERF-080 remain dependency-blocked. | Adopt the coherent Testing Harness owner text before changing projections or runtime behavior. |
| 2026-08-16T18:54:45-04:00 | PERF-010 close and PERF-020 activation | `HEAD 03a3505a6a86e3ff6286e572acbdbc69ad5dd4d8`; complete dirty-state SHA-256 `06fa8e15c4dfd480402368c90509ce3c5abdabc9496218371b54d08d3f6582f0` | Adopted one ranked resource-fit scheduler rule, explicit local-session ownership, event v2 interval boundaries, work-graph v4 evidence/artifact separation, and cache-record v2. Updated all current producers, readers, schemas, attachments, contract fixtures, and generated topology index; no compatibility reader remains. `make generate` passed at `.cartulary/test-results/20260816T225309Z-p747318`; `make json-shape-check` passed at `20260816T225329Z-p750827`; `make generate-drift` passed at `20260816T225340Z-p751371`; `make generated-artifact-policy-check` passed at `20260816T225340Z-p751384`; `make harness-contract` passed at `20260816T225340Z-p751612`; `make lint-markdown` passed at `20260816T225407Z-p755200`; retired-identity scan and `git diff --check` passed. Contract-only slice used no product-run budget. | PERF-010 `DONE`; current commands accept v2/v4/v2 only, wait closure is matched and deterministic, and the staged tracker remains non-executable. PERF-020 is `IN_PROGRESS`. | Implement the scheduler/cache fast path, bounded event writer, and corrected Go worker budget before spending the one warm `test-fast` run. |
| 2026-08-16T19:13:20-04:00 | PERF-020 close and PERF-030 activation | `HEAD 03a3505a6a86e3ff6286e572acbdbc69ad5dd4d8`; pre-tracker complete dirty-state SHA-256 `0c78075e29e560a6486e5671edcd5f912ae08d777fd92cb76c676450154af8c5`; adoption-run source digest `sha256:766fd2613ed469b257b6194a36f8285ed9aad003afd921c0e78f425d82ea8966` | Replaced graph rescans with incremental pending/ready/successor indexes, moved cache lookup before resource or fixture admission, added fresh semantic-result replay, memoized executable-source profile digests, streamed v2 events through a bounded atomic writer, and applied exact Go compatibility plus the TH-379 child CPU budget. Changed the NLSpec, cache/event/work-graph/Go-plan schemas and attachments, scheduler/compiler/cache/runner/source-index implementations, evidence consumers, diagnostics, readiness cache producer, and contract/backend/browser/observability tests; regenerated `tools/execution_topology_render_index.json`. Initial `make harness-contract` failed at `20260816T230720Z-p765082` because the immutable cache-fixture snapshot test expected an in-run source mutation to be rescanned; the fixture was corrected and `make generate-drift harness-contract` passed at roots `20260816T230808Z-p769393` and `20260816T230808Z-p769467`. The first post-cutover `make test-fast` passed cold at `20260816T230834Z-p772787` in 115.299s with 405/405 units, 577 rows, cache 0/402/3, and 1,132,080 event bytes, so it was retained as an invalid warm comparison. The next warm attempt reached 402 hits but failed before summary publication at `20260816T231117Z-p813110` because legacy accounting subtracted hits that no longer emit `started`; its otherwise complete v2 stream ended at 0.879s. The accounting fix initially produced expected generated drift at `20260816T231147Z-p813855`; `make generate` passed at `20260816T231202Z-p816981`, then `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check`, and `make harness-contract` passed at `20260816T231215Z-p819912`, `20260816T231215Z-p819914`, `20260816T231215Z-p819916`, and `20260816T231215Z-p819986`. The valid warm `make test-fast` passed at `.cartulary/test-results/20260816T231242Z-p823738`: 405/405 units, exact 577-row fast closure, 0.883s versus the directional 19.175s reference, cache 402/0/3, capacity 24 CPU/8 PostgreSQL, peak three CPU/process/I/O claims, no saturation, 405 single matched wait intervals, 2,440 events, 939,464 bytes (more than 99% below 102 MB), fresh current-run row timestamps, owner-only retained files, complete cleanup, and no resource claim for 402 hits. Bounded-writer abort, cancellation/finalizer, cache corruption, and no-Markdown source-boundary fixtures passed in `harness-contract`; no service or secret-bearing resource participated. | PERF-020 `DONE`. The cold hard-cutover and failed warm attempts are retained, not averaged. The accepted warm result exceeds the directional duration target and mandatory event-volume gate without semantic, isolation, cleanup, or evidence loss. Detailed dependency and artifact security hardening remains assigned to PERF-070 rather than extending this slice. PERF-030 is `IN_PROGRESS`. | Implement the explicit owner-only local service descriptor, borrower leases, redacted status, public commands, and owned-only target rejection before running one attached service-backed slice. |
| 2026-08-16T19:50:56-04:00 | PERF-030 close and PERF-040 activation | `HEAD 03a3505a6a86e3ff6286e572acbdbc69ad5dd4d8`; pre-tracker complete dirty-state SHA-256 `6b90a4f2d29cc282e22d2bd000676b31aff2ef394810a766001a80d7014d7e47`; valid attached-run source digest `sha256:4543f662de2a9b45b9cf4eab9c379db339dd2c03886cd0a695d9f7606c379d9a` | Added `cartulary.test_services.local_session.v1` and redacted status schemas; public up/status/down commands; explicit `owned|attach` selection on exactly the three adopted leaf targets; owner-only, non-symlink descriptor, state, lock, borrower, and runtime boundaries; current schema/tool/config/image/container/UID/expiry proof; live borrower refusal; exact idempotent teardown; persistent-container janitor exclusion; and per-borrower database cleanup. Changed the Testing Harness owner, task-surface owner and generated projections, schema attachments, session manager/CLI/tests, graph runner, suite runtime, testservices lifecycle/tests, `suiteservices` and `pgtest`, the harness browser family row, implementation testing guide, and generated topology index; `docs/domain.md` remains unchanged and no runtime reads Markdown. Contract fixtures cover missing, malformed, permissive, wrong-owner, wrong-mode, stale digest, expired, dead-container, symlink, concurrent borrower, live-down, repeated-down, redaction, every owned-only target, and internal-selector injection. `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.testservices_lifecycle_contract CARTULARY_HARNESS_CACHE_MODE=off` passed at `20260816T234833Z-p886400`. Final `make generate` passed at `20260816T234902Z-p890849`; `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check`, and `make harness-contract` passed at `20260816T234917Z-p893889`, `20260816T234917Z-p893905`, `20260816T234917Z-p893927`, and `20260816T234917Z-p894229`. The first up attempt failed at `20260816T234045Z-p854445` because the descriptor projection used `object-store` instead of the existing typed `object_store` identity. After correction, up passed at `20260816T234223Z-p866391`; the first attached slice passed in 3.218s at `20260816T234253Z-p867976` but was rejected as invalid adoption evidence because its per-test database was retained. The first down attempt then failed at `20260816T234305Z-p870023` because the terminator lacked suite-runtime proof; the repaired down and idempotent repeat passed at `20260816T234411Z-p871282` and `20260816T234422Z-p872036`. After separating persistent borrowers from intra-run attachment, final up passed at `20260816T234933Z-p898016` and the valid cache-off attached row `app.operator.integration.migration_evidence_transport_and_resource_lifecy_6f6f81a872` passed at `.cartulary/test-results/20260816T234959Z-p899374`: exact 1-row closure, 3/3 units, 3.650s versus the directional 21--26s total and 3--8s adoption range, cache 0/0/3, capacity 24 CPU/8 PostgreSQL, peak 24 CPU/one PostgreSQL/two I/O/one process, 2.501s CPU saturation, 30 events/8,789 bytes, fresh evidence, redaction pass over 19 files, unique borrower suite identity, and an exact `postgres-db-dropped` terminal event. Status then proved both borrowed containers active with zero borrowers. Final down and repeated down passed at `20260816T235022Z-p901551` and `20260816T235030Z-p902298`; redacted status reported absent, the descriptor/state root was removed, and an exact session-label Docker query returned no containers. | PERF-030 `DONE`. Adopted explicit attachment and a hard selector cutover; rejected aliases, automatic attachment, caller-controlled internal state, and retaining borrower-owned databases. The failed/semantically invalid attempts are retained but not averaged. Persistent containers intentionally survive borrower cleanup and are removed only by exact session down; expired descriptors still require proof-driven down rather than ordinary janitor deletion. PERF-040 is `IN_PROGRESS`. | Audit current PostgreSQL fixture rows, move capacity/profile policy into typed ownership, and perform the single matched Network Flow capacity comparison only after structural and failure-injection gates pass. |
| 2026-08-16T20:25:08-04:00 | PERF-040 close and PERF-050 activation | `HEAD 03a3505a6a86e3ff6286e572acbdbc69ad5dd4d8`; pre-tracker complete dirty-state SHA-256 `4e93631cb90476828449f671235734f60de228aeca878aebc10b11ee6d9cb3cf`; matched-run and final-check source digest `sha256:4d89047825492c2ea3c08bdd598c52fbe073f5391bd9d0ff05e9972041c70627` | Updated the Testing Harness owner and replaced scheduler resource registry v5 with v6 typed capacity policies and host-bound override validation. Added `cartulary.postgres_fixture_policy_registry.v1`, its exact 81-row transaction approval roster, current-catalog closure validation, and conservative rationales for all 341 active PostgreSQL rows. Reclassified the three unproven group rows in Evidence, Incidents, and Records to dedicated scope, changed their tests to isolated databases, and removed the unused group helper path. Added `cartulary.performance_fixture_builder_policy.v1` and the `performance_fixture_builder` topology profile; the snapshot builder now claims CPU 1, I/O 1, memory 512 MiB, process 1, and PostgreSQL 1 only. Added fixture quarantine failure evidence and repaired a PERF-020 concurrency defect discovered by the real workload by serializing event publication and emitting lifecycle boundaries as atomic batches. Changed the owner/specification, scheduler registry/schema/loader/capability resolver, topology owner and generated projections, fixture policy/schema/catalog validator, performance builder policy/schema/loader/compiler/tests, three owner manifests and integration tests, `pgtest`/`appsupport` group helpers, scheduler event runtime/tests, schema attachments, generated-input registry, and generated render index; `docs/domain.md` remains unchanged. `make task-guide ROLE=module-author OWNER=module.networkflow` ran before the candidate. The three changed owner rows passed cache-off attached slices at `20260817T001035Z-p922980`, `20260817T001048Z-p924851`, and `20260817T001057Z-p926509`. The first capacity-8 attempt at `20260817T001128Z-p928830` was invalid because final accounting found a monotonic event regression; after timestamp ordering, the second at `20260817T001507Z-p980604` was invalid because a fixture event interleaved inside a wait/admission boundary. Both correctness defects were repaired and neither run entered the comparison. The valid matched cache-off pair used the same 24-CPU host, attached session, source digest, 45-unit/33-row closure, and workload digest: capacity 8 passed in 125.094s at `20260817T001900Z-p1027378`; capacity 12 passed in 121.861s at `20260817T002111Z-p1069227`. Cache was 0/0/45 in both. Event streams were exact 426-event v2 files of 183,463 and 183,447 bytes. PostgreSQL peak/saturation changed from 8/44.334s to 12/26.916s; browser peak changed from 2 to 3, modeled memory peak from 1,792 to 2,048 MiB, and dependency critical path from 54.647s to 55.477s. Both retained-secret scans passed over 530 files, all units and rows passed, the borrower count returned to zero, and exact session down passed at `20260817T002402Z-p1115629` with no labeled containers left. `make format` passed at `20260817T002352Z-p1112053`; final `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check`, and `make harness-contract` passed at `20260817T002411Z-p1116555`, `20260817T002411Z-p1116575`, `20260817T002411Z-p1116592`, and `20260817T002411Z-p1116887`. | PERF-040 `DONE`. Adopted typed capacity ownership, closed row classification, conservative dedicated isolation, independent builder topology, and atomic event boundaries. Rejected capacity 12 because its 3.233s/2.584% improvement missed the 15% gate; retained default 8. No transaction conversion was inferred, no group compatibility bridge remains, and the checked-in capacity-12 declaration remains only a negative/adoption test fixture. PERF-050 is `IN_PROGRESS`. | Add diagnostics outside the semantic receipt, evaluate the 2,000-row Timeline batch at the exact current measurement row, and retain it only if semantics and the 20% gate pass; otherwise implement the owner set-oriented fallback. |
| 2026-08-16T21:05:43-04:00 | PERF-050 close and PERF-060 activation | `HEAD 03a3505a6a86e3ff6286e572acbdbc69ad5dd4d8`; pre-tracker complete dirty-state SHA-256 `4162310514a5c75ad6ad6d3c5b1f8874d84c7d2140ebc1b19ee78fa823650ebc`; accepted-run source digest `sha256:381f45cea7233a0f20df9e79bf28cafcd3b65d0d1e79c3eba104b024fd5445a7` | Updated REQ-04-157 and TH-813, added the redacted `cartulary.performance_fixture_build_diagnostics.v1` schema/projection, instrumented contribution and semantic-validation timing outside all semantic identities, and added success/failure publication. The 2,000-row candidate was rejected at `20260817T003639Z-p1141522` because the production mutation contract limits tabular batches to 500. Added a Timeline-owned set-oriented fixture command and store boundary, typed record and projection batch ports, one transaction, strict input validation, exact relationship hydration, and owner read-back validation; no cross-owner SQL or product API changed. Two independent 20,000-row builds with identical digests passed at `20260817T004442Z-p1173482`. A stale testservices helper invalidated the first fallback attempt at `20260817T004906Z-p1183268`; the broad helper build cache omitted transitive internal-module inputs, so `make testservices-build CARTULARY_FORCE_REBUILD=1` rebuilt it and the dependency gap remains assigned to PERF-070. The next exact run at `20260817T005317Z-p1209622` built and exercised the browser successfully but exposed a finalizer/event-publication race. TH-281 and the event runtime now permit validated run-contained staging reads by same-run finalizers while retaining atomic canonical publication; contract tests cover publication, consumption, containment, and cleanup. The accepted cache-off command `make test-slice OWNER=module.timeline ROWS=module.timeline.measurement.timeline_blank_row_creation_satisfies_the_paint_afddd2ce13 CARTULARY_HARNESS_CACHE_MODE=off` passed at `20260817T005901Z-p1240805`: 14/14 units, exact one-row closure, cache 0/0/14, 169.975s target wall, capacity 24 CPU/8 PostgreSQL, peak CPU/process/I/O 4, memory 1,024 MiB, PostgreSQL 4, object store/port 1, and no saturation. Its v2 stream had 122 events and 44,691 bytes with no retained staging file. The builder took 100.515s versus 222.604s, improving 122.089s/54.846%; Timeline took 8.610s in one 20,000-item owner set-oriented batch, while Entities took 90.939s. Exact count, ID, relationship, condition, quietness, 100-sample/zero-retry browser predicate, cleanup, secret scan, and qualified aggregate checks passed. Semantic digest `sha256:b9549c30dab549504fd945fc211da7649223573e8ea67d57b7bb17a2ff06b5a7` matched prior successful builds. `make generate` passed at `20260817T010429Z-p1272085`; the first concurrently launched final checks at `20260817T010241Z-p1266574` and `20260817T010241Z-p1266611` correctly reported stale generated inputs after the event fix and are invalid closure evidence. Serial `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check`, and `make harness-contract` then passed at `20260817T010442Z-p1275040`, `20260817T010453Z-p1277999`, `20260817T010457Z-p1278451`, and `20260817T010504Z-p1278994`. `docs/domain.md` remains unchanged and no executable path reads Markdown. | PERF-050 `DONE`. Adopted the separate diagnostics artifact and the owner set-oriented boundary because the clean batch-size candidate could not preserve the existing input contract. The accepted builder exceeds the 20% gate without semantic, measurement, redaction, isolation, cleanup, or evidence regression. The invalid attempts are retained and not averaged. PERF-060 is `IN_PROGRESS`. | Add the typed `browser_functional` profile, stable selected-row LPT lane plan, warm lane lifecycle, fresh per-group state, reset/failure quarantine, and simulations before spending the single webserver-backed target run. |
| 2026-08-16T21:46:52-04:00 | PERF-060 close and PERF-070 activation | `HEAD 03a3505a6a86e3ff6286e572acbdbc69ad5dd4d8`; pre-tracker complete dirty-state SHA-256 `30dfa1496e484186a8487d28895c46d7c8aa4dd22feb3d3ce24b60dc2ca36af4`; accepted-run source digest `sha256:79acf486fc493e430d88277bfdf3ebaf987e30d4c4be247d161e8f63fdad0912` | Updated TH-806 first, added the typed `browser_functional` profile, migrated all 83 current ordinary webserver-backed rows, and compiled 26 groups into four runtime-compatible lanes using selected-row cost and stable LPT placement. Added warm affinity retention, explicit inter-group resource renewal and application reset, reset finalization, failure quarantine, allocation-unique session evidence, reset-aware target finalization, and deterministic success/failure simulations. Stateful, measurement, accessibility, and visual profiles are unchanged. `make format` passed at `20260817T013444Z-p1389131` and `20260817T014106Z-p1481506`; `make generate` passed at `20260817T013451Z-p1392678` and `20260817T014112Z-p1485050`; `make harness-contract` passed at `20260817T013505Z-p1395716` and `20260817T014132Z-p1488353`; the cache-off testservices contract row passed at `20260817T013533Z-p1396617`; and the focused cache-off Auth browser row passed 12/12 at `20260817T014148Z-p1488937`. Final `make generate-drift`, `make generated-artifact-policy-check`, and `make json-shape-check` passed at `20260817T014235Z-p1513037`, `20260817T014246Z-p1515978`, and `20260817T014251Z-p1516433`. The first full attempt at `20260817T012414Z-p1302576` was invalid: all 26 groups passed, physical database replacement terminated the four warm backends, and 22 resets failed; it ended after 39/62 units in 267.660s. The second at `20260817T013614Z-p1401118` was invalid: all 22 in-place resets passed, but all groups rejected a missing exported concrete session ID; it ended after 35/62 units in 197.629s. The accepted `make browser-e2e-webserver-backed` passed at `.cartulary/test-results/20260817T014259Z-p1516987`: 62/62 units, 26 groups, exact 83-row closure, cache 0/0/62, and 165.251s versus 290.536s (43.121% faster). Its event stream contained 604 events/328,954 bytes. Four browser/object-store/port lanes and eight PostgreSQL tokens were used at exact peak; 22 resets returned HTTP 200; 52 broker leases were released; four unique session/database/bucket/runtime/port identities were retired; all eight retained browser-stack process identities were stopped; the target finalizer and retained-secret scan over 752 files passed; and no resource holder, collision, staging file, or secret remained. `docs/domain.md` remains unchanged and no runtime reads Markdown. | PERF-060 `DONE`. Adopted deterministic four-lane warm process reuse with fresh semantic resources and exact evidence. Rejected physical database replacement because it violates the immutable warm-connection boundary. Both invalid full attempts are retained and not averaged. The accepted result exceeds the 20% gate without row, retry, evidence, failure-finalizer, isolation, cleanup, redaction, or special-profile regression. PERF-070 is `IN_PROGRESS`. | Implement conservative typed Go and TypeScript dependency closures plus all-or-none artifact manifests/restoration and the complete corruption/containment matrix before any additional producer becomes cacheable. |
| 2026-08-16T22:25:47-04:00 | PERF-070 close and PERF-080 activation | `HEAD 03a3505a6a86e3ff6286e572acbdbc69ad5dd4d8`; pre-tracker complete dirty-state SHA-256 `41e328010207692c416f3fb8a136efa68a2b3e20a2b161352980010156abfc73`; accepted-run source digest `sha256:2fec3d94faaca204d82dba79ec0c96fe868fb223d41ad7461347328210d010dd` | Updated TH-807 first, extended the typed cache registry with unit-aware versus broad strategies, moved v2 entries to `.cache/cartulary/graph-v2`, and implemented immutable source-index-derived Go package and TypeScript workspace closures with safe broad fallback and per-snapshot memoization. Replaced byte replay with semantic results plus exact typed file/directory manifests, exclusive staging, immutable concurrent publication, quarantine, complete entry validation, transactional restore/rollback, and fresh row/unit evidence. Service-backed and security rows now join all other freshness-sensitive work as non-cacheable; build/license/SBOM remain disabled pending complete producer output and freshness declarations. Changed the Testing Harness owner; cache registry/schema and cache-record v2 schema; dependency resolver, cache, compiler, runner, model, and exports; contract/cutover tests; fixture quarantine terminal state; and generated topology index. Contract matrices cover source add/delete/rename, local/transitive Go and workspace changes, manifest/lock/toolchain/schema/helper/generator changes, untracked inputs, incomplete resolution, concurrent writers, content/metadata/path/destination/producer/type/mode corruption, missing/surplus/partial payloads, directory closure, symlink/hard-link/FIFO attacks, rollback, and fresh semantic evidence. Initial `make harness-contract` attempts failed at `20260817T020315Z-p1565361` (fixture lacked the new dependency strategy) and `20260817T020353Z-p1566405` (the corruption fixture still addressed the v1 payload shape); both tests were corrected. `make harness-smoke-cache-artifact` was an invalid command because no such Make target exists. `make run-harness-smoke-fast` failed at `20260817T021049Z-p1578122` because explicit quarantine was collapsed into destroyed state; the broker now retains the required terminal distinction, and the smoke passed at `20260817T021133Z-p1584711` and finally at `20260817T022048Z-p1617896` in 7.273s. An `explain-run` attempt with unsupported `DETAIL=rows` failed usage validation; `DETAIL=summary` then passed. A first real miss/hit pair at `20260817T022141Z-p1621212` and `20260817T022148Z-p1621604` passed but was superseded after it exposed that v2 should use a distinct root rather than merely a schema-key-separated entry. Final `make format` and `make generate` passed at `20260817T022347Z-p1622998` and `20260817T022354Z-p1626545`; `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check`, and `make harness-contract` passed at `20260817T022406Z-p1629474`, `20260817T022417Z-p1632452`, `20260817T022420Z-p1632901`, and `20260817T022430Z-p1633453`. `make task-guide ROLE=module-author OWNER=platform.config` selected the exact pure-unit proof. Its v2 miss passed at `20260817T022445Z-p1634018` in 1.419s with one admission, 12 events, and 3,088 bytes; its v2 hit passed at `20260817T022453Z-p1634416` in 0.339s with no admission, 10 events, 2,741 bytes, and fresh evidence. Both used one exact unit/row, normal cache mode, capacity 24 CPU/48 I/O/96 processes, and no service or fixture. `docs/domain.md` remains unchanged and executable source indexing excludes Markdown. | PERF-070 `DONE`. Adopted conservative unit-aware row closures, a hard v2 root cutover, immutable entry publication, complete typed artifact validation, transactional restore, corruption quarantine, and current-run evidence regeneration. Rejected enabling build, license, or SBOM caching because their work-graph artifact/freshness declarations are incomplete; safe bypass is the intended completed state. No full-suite performance budget was consumed. PERF-080 is `IN_PROGRESS`. | Reconcile documentation and every owner projection, run final narrow gates, execute `make agent-finalize`, then obtain one successful `make check` followed by one warm successful `make release-check` before closing the tracker. |
| 2026-08-16T22:37:47-04:00 | PERF-080 validation reopens PERF-050 | `HEAD 03a3505a6a86e3ff6286e572acbdbc69ad5dd4d8`; pre-tracker complete dirty-state SHA-256 `6816ae6d3fc953a6ad6552bea26f6b564dcb4daad73bd18f9a1e8a2d1fc88817`; failed-run source digest `sha256:2fec3d94faaca204d82dba79ec0c96fe868fb223d41ad7461347328210d010dd` | Pre-broad `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check`, `make harness-contract`, and `make lint-markdown` passed at `20260817T022918Z-p1639526`, `20260817T022929Z-p1642477`, `20260817T022933Z-p1642926`, `20260817T022941Z-p1643472`, and `20260817T022956Z-p1644084`; `git diff --check` passed. `env -u RESULTS_DIR make agent-finalize` passed at `20260817T023014Z-p1645086` and correctly recorded `results-dir-not-provided`. The first `make check` failed at `.cartulary/test-results/20260817T023037Z-p1648051` after 395.027s with 774/776 units and cache 0/464/312. `target:lint-go` reported 13 capitalized Timeline fixture error strings, and `module.records.architecture.boundary` reported two production reads of `timeline_grid_projection` from `internal/modules/timeline/performance_fixture_store.go`. An attempted `explain-run DETAIL=logs` without `TARGET` also failed usage validation; direct canonical unit logs identified both defects. | PERF-050 is reopened and is the only `IN_PROGRESS` slice; PERF-080 is `BLOCKED`. The failure is related implementation evidence, not an environmental exception, and the failed full run is retained but cannot satisfy final validation. | Repair the Timeline error and projection-storage ownership boundaries without suppressing static or architecture enforcement, run narrow owner/architecture/lint gates, then checkpoint PERF-050 before reactivating PERF-080. |
| 2026-08-16T22:45:48-04:00 | Reopened PERF-050 close and PERF-080 reactivation | `HEAD 03a3505a6a86e3ff6286e572acbdbc69ad5dd4d8`; pre-tracker complete dirty-state SHA-256 `ac91b4087d9f9c4a7abd2370906bc10b4a0f60fa20baf07a7b81b0763716677b`; accepted-run source digest `sha256:0bc146b0c0d6b4a4c5af339703d87f18a41c0b998c83b77a90ac303a6b1fa01d` | Lowercased the new Timeline error strings and replaced both cross-owner projection-table reads with a construction-only `workbookprojection.FixtureBatchPort` count capability. Timeline retains application validation and the Projections owner retains all `timeline_grid_projection` SQL in its declared physical storage. `make format` passed at `20260817T024026Z-p1819744`; all three `make lint-go` children passed at `20260817T024033Z-p1823469`, `20260817T024035Z-p1826781`, and `20260817T024043Z-p1835480`; the exact cache-off architecture row passed at `20260817T024109Z-p1837238`; and `make task-guide ROLE=module-author OWNER=module.timeline` confirmed the owner route. The exact cache-off Timeline measurement passed at `.cartulary/test-results/20260817T024127Z-p1837874`: 14/14 units in 171.066s, cache 0/0/14, 122 events/44,691 bytes, builder 100.388s, semantic validation 0.016s, Timeline 8.642s in one 20,000-item set-oriented batch, exact 20,000 Timeline rows, all counts and conditions passed, semantic digest `sha256:b9549c30dab549504fd945fc211da7649223573e8ea67d57b7bb17a2ff06b5a7`, cleanup complete, and retained-secret scan pass over 118 files. `make generate` passed at `20260817T024450Z-p1863209`; final `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check`, and `make harness-contract` passed at `20260817T024504Z-p1866175`, `20260817T024515Z-p1869122`, `20260817T024519Z-p1869579`, and `20260817T024525Z-p1870120`. | PERF-050 is `DONE` again. The clean owner-port repair preserves the prior 54.846% builder adoption decision, exact semantic identity, set-oriented topology, ownership, redaction, isolation, and cleanup. PERF-080 is `IN_PROGRESS`. | Checkpoint this tracker edit, rerun the mandated pre-broad gates and `agent-finalize`, then obtain a successful `make check` followed by one warm successful `make release-check`. |
| 2026-08-16T23:16:04-04:00 | PERF-080 close and handoff | `HEAD 03a3505a6a86e3ff6286e572acbdbc69ad5dd4d8`; tracker-excluded dirty-state SHA-256 `b134a5d3dfeeb08a1674c78ac1f5dabe750c6c70842e53111de44290bc9388cc`; accepted-run source digest `sha256:0bc146b0c0d6b4a4c5af339703d87f18a41c0b998c83b77a90ac303a6b1fa01d` | Reconciled the Testing Harness owner, typed projections, implementation, contract fixtures, generated outputs, and user documentation. Added explicit session usage and hard-cutover migration notes to `docs/guides/cartulary_implementation_testing_guide.md`; left `docs/domain.md` and `tools/harness_public_target_duration_baselines.v3.json` unchanged, retained only the adopted redacted performance-fixture diagnostic, and confirmed executable source indexing excludes Markdown. The final pre-broad `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check`, `make harness-contract`, and `make lint-markdown` passed at `20260817T024710Z-p1873103`, `20260817T024721Z-p1876031`, `20260817T024725Z-p1876486`, `20260817T024735Z-p1877054`, and `20260817T024752Z-p1877694`; `git diff --check` passed. `env -u RESULTS_DIR make agent-finalize` passed at `20260817T024757Z-p1878507` and recorded the required `results-dir-not-provided` retained-run-maintenance skip because no successful full warm current-source root existed yet. Final `make check` passed at `.cartulary/test-results/20260817T024813Z-p1881416`: graph digest `sha256:668de02c8aeceb82fa18467f672a4928ffe22b25b6e80ce01ec80cb4741c1c40`, 776/776 units, exact 968 rows, 358.508s, cache 398/66/312, 5,954 events/2,327,944 bytes, capacity 24 CPU/8 PostgreSQL, retained-secret scan pass over 6,251 files, and exact cleanup/run completion. The immediately following warm `make release-check` passed at `.cartulary/test-results/20260817T025434Z-p1994125`: graph digest `sha256:5b0cc62d2d154be4d34f824aca32b4a25fa4ba5a9b7e3f9c3552db83fe8055d2`, 974/974 units, exact 1,202 rows (`go` 650/Playwright 154/shell 9/Vitest 389), 1,094.157s, cache 456/34/484, 7,724 events/3,260,645 bytes, capacity 24 CPU/8 PostgreSQL, exact 974 wait lifecycles, 518 admission/start pairs, 456 resource-free hits, matched 420 fixture acquire/releases, 61 passing target summaries, retained-secret scan pass over 8,766 files, and passing release/readiness/inventory/cleanup/run terminal evidence. Exact final scope/container inspection found no scope-labeled container or event staging file after the passing owned teardown. `harness-contract` supplied deterministic cancellation, bounded-writer, cache corruption, finalizer, browser reset failure/quarantine, redaction, and fail-fast proof; no deliberately failing full release was run. | PERF-080 and the overall effort are `DONE`. The warm release is 186.940s/20.6% slower than the non-matched historical 907.217s reference, so no aggregate performance improvement or formal baseline is claimed. PostgreSQL/object-store saturation remains a documented follow-up risk. All mandatory correctness, exact closure, evidence, isolation, security, redaction, cleanup, generated-state, and documentation gates passed; no legacy reader, selector alias, automatic attachment, capacity-12 adoption, 2,000-row batch, unsafe fixture grouping, physical browser database replacement, or unproven artifact-cache producer was retained. | Handoff complete. Preserve the retained roots, use `make explain-run` for future analysis, and treat any formal TH-377--TH-380 baseline publication or further service-path optimization as a separately authorized effort. |

The PERF-030 tracker checkpoint passed `make lint-markdown` at
`.cartulary/test-results/20260816T235315Z-p904431`; `git diff --check` also
passed. No PERF-040 implementation began before that checkpoint validation.

The PERF-040 tracker checkpoint passed `make lint-markdown` at
`.cartulary/test-results/20260817T002704Z-p1122215`; `git diff --check` also
passed. No PERF-050 implementation began before that checkpoint validation.

The PERF-050 tracker checkpoint passed `make lint-markdown` at
`.cartulary/test-results/20260817T010705Z-p1280306`; `git diff --check` also
passed. No PERF-060 implementation began before that checkpoint validation.

The PERF-060 tracker checkpoint passed `make lint-markdown` at
`.cartulary/test-results/20260817T014813Z-p1555445`; `git diff --check` also
passed. No PERF-070 implementation began before that checkpoint validation.

The PERF-070 tracker checkpoint passed `make lint-markdown` at
`.cartulary/test-results/20260817T022708Z-p1635624`; `git diff --check` also
passed. No PERF-080 implementation began before that checkpoint validation.

The reopened PERF-050 tracker checkpoint passed `make lint-markdown` at
`.cartulary/test-results/20260817T024626Z-p1871004`; `git diff --check` also
passed. PERF-080 did not resume before that repair checkpoint validation.

The PERF-080 completion checkpoint passed `make lint-markdown` at
`.cartulary/test-results/20260817T031743Z-p2208301`; its post-record lint passed
at `.cartulary/test-results/20260817T031807Z-p2209364`; `git diff --check` also
passed. All serial workstreams and the overall handoff are closed.
