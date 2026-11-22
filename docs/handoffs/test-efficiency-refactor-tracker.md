# Test Execution Efficiency Remediation Plan

**Status.** Planning handoff. This document authorizes no implementation by itself.
Each specification amendment in [Section 8](#8-specification-amendment-ledger) needs
owner approval before its workstream starts. Workstream 0 is the only work that
needs no new decision.

## 1. Scope, Authority and Evidence Posture

| Item | Baseline |
| --- | --- |
| Goal | Reduce wall time and host resource use of `make check`, `ci` and `release-check`, plus the shared harness costs every graph target pays. |
| Checkout | `/home/jochi/code/cartulary`, branch `main`, HEAD `6712f0408` ("Increase default PostgreSQL heavy lane capacity"). |
| Host | WSL2, 24 logical CPUs, 31 GiB memory. All timings below come from this host. |
| Authority | The [Testing Harness NLSpec](../testing-harness-nlspec.md) owns harness mechanics. Machine inputs (`tools/scheduler_resource_registry.json`, `tools/execution_topology_manifest.json`, `tools/harness_work_graph_owner.json`, schemas) are projections of it. Product test changes stay inside their owning module's NLSpec and Core sections. |
| Fixed constraint | TH-HARNESS-REQ-804 preserves `fsync`, full-page writes, synchronous commit, checksums and normal checkpoint behavior for harness PostgreSQL. No remediation below weakens durability. |
| Evidence | Retained runs `eff-base-check01` (2 heavy lanes) and `eff-heavy3-check01` (3 heavy lanes), both `make check HARNESS_DIAGNOSTICS=basic CARTULARY_HARNESS_CACHE_MODE=off` on commit `fa24d95f6`; ten earlier `check` runs (`slowrows-*`, `post1210-*`, `tcr-*`); `make target-plan` output for `test-fast`, `check`, `ci` and `release-check`; source inspection cited per gap. |
| Evidence limits | One run per treatment. The ten historical `check` runs varied between 1,077 s and 1,352 s. `ci` and `release-check` have no retained runs; their figures are authored planner estimates. Host-admission diagnostics retained 3,153 of 173,449 records in the baseline, so admission counts are samples. |

## 2. Measured Baseline

| Measure | 2 heavy lanes | 3 heavy lanes (current default) |
| --- | ---: | ---: |
| `check` graph wall | 1,341 s | 1,164 s |
| `check` outer command wall | 1,471 s | 1,290 s |
| Time before graph start | 6.7 s | 4.2 s |
| Time after graph end (retained-secret scan) | ~120 s | ~121 s |
| Heavy-lane work (17 units) | 2,589 s | 3,294 s |
| CPU-only units waiting more than 300 s to start | 197 of 239 | 199 of 239 |
| CPU-only ready-to-start wait, summed | 108,391 s | 143,766 s |
| Fixture acquisition, summed | 1,243 s | 1,391 s |
| Peak sampled RSS sum | 12.9 GB | 15.0 GB |
| Retained run root | 1.1 GB | 1.1 GB |

Additional facts:

- **Head-of-line blocking.** For the first ~600 s, 9–12 of 19 logical CPU tokens are held by PostgreSQL units waiting in physical host admission. The graph engine reserves logical claims and emits `started` before the unit enters the `host_admission` phase ([engine.go](../../internal/testutil/harnessruntime/engine.go) `admit` → `reserve`; [lifecycle.go](../../internal/testutil/harnessruntime/lifecycle.go) `host_admission`).
- **Admission deficits.** Sampled denials name `process` (1,386) and `postgres` (1,377). Each exclusive unit is charged once for its own claim and again for its PostgreSQL container, whose resident claim is 8 process slots and 1 PostgreSQL lane ([execution_topology_manifest.json](../../tools/execution_topology_manifest.json) `physical_resource_claims`). The pool is 76 process slots (4 × 19 CPU tokens) and 16 PostgreSQL lanes. Both run out at about 8 concurrent exclusive units.
- **Heavy-lane contention.** Moving from 2 to 3 lanes made heavy units run 21–27% slower, and restore or checkpoint rows up to 3× slower (`recovery.integration.reset_preserves_crypto_format_identity`: 42 s → 135 s).
- **Heavy work concentration.** Seven rows account for 45% of heavy-row time: `imports.integration.current_workflows` (226 s), `incidentbundles.integration.network_flow_portability` (194 s), `recovery.integration.attack_production_recovery` (138 s), `networkflow.integration.saved_graph_lifecycle_v2` (114 s), `incidentbundles.integration.attack_pack_portability_…` (114 s), `incidentbundles.integration.incident_bundle_export_and_import_…` (102 s) and `reportcomposition.integration.preview_reporting_delegation_…` (98 s).
- **Run-root composition.** 599 MB of each 1.1 GB root is `_shared/test-services`, mostly six copies of a 95,482,880-byte signed ATT&CK `container.pack` that reference-data test support deliberately writes ([dataset_fixture.go](../../internal/modules/reference_data/testsupport/dataset_fixture.go)). `unit-events.ndjson` is 159 MB and holds 110,873 `wait_started`/`wait_ended` pairs; one unit records 680 segments. Network Flow fixture materializations add up to 113 MB per fixture (38 directories per run).
- **Secret scan.** The post-graph scan walks the root on one thread and runs `bytes.Contains` once per secret per 64 KiB chunk ([harnesssession/scan_linux.go](../../internal/testutil/harnesssession/scan_linux.go)).
- **Temporary-directory leaks.** `~/.cache/cartulary/go/tmp` holds 52,843 `go-build*` directories and about 2,600 test temporary directories (3.5 GB), back to 2026-08-06. A clean `check` leaks about 20 directories from five `_Process` tests. The children re-execute the test binary and leave via `os.Exit`, which skips `t.Cleanup` ([publication_deadline_test.go](../../internal/app/server/publication_deadline_test.go)). The leaked directories contain test TLS private keys. Clean runs leak no `go-build*` directories; those come from interrupted runs.
- **Retention.** `.cartulary/test-results` holds 15,994 runs (122 GB). `make clean` removes all of them or none. The Go build cache is 182 GB, and nearly all of it was accessed within the last five days, so Go's own 5-day trim is working and the volume is real churn.
- **Aggregates.** `ci` plans 506 units and adds 33 browser units (about 15,800 s estimated), including 4 quiet measurement groups (7,324 s). `release-check` plans 585 units, including 33 quiet measurement groups of about 1,831 s each (60,423 s). Quiet groups take the exclusive `host_activity` lock ([browser.mjs](../../tools/harness/scheduler/work-graph/browser.mjs) lines 395–396), so they run alone.

## 3. Principles and Targets

Decision rules applied to every gap:

1. Fix the model that produces a symptom, not the symptom. Capacity should be stated explicitly, not emerge from arithmetic side effects.
2. Retained data must earn its bytes. Keep evidence needed to judge a result; keep bulk bytes only for failures.
3. Schema changes cut over. Older retained runs stay historical evidence and are not translated, matching the NLSpec's current-epoch reader rule.
4. Measure before and after every change with the protocol in [Section 9](#9-validation-protocol).
5. Durability, isolation scope and secret-scan coverage are not traded for speed.

Targets are hypotheses to confirm against measurements, not commitments.

| Metric | Current | Target |
| --- | ---: | ---: |
| `check` graph wall | 1,164 s | ≤ 750 s |
| `check` outer wall | 1,290 s | ≤ 800 s |
| Post-graph overhead per invocation | ~121 s | ≤ 15 s |
| CPU-only units waiting more than 300 s | 199 | 0 |
| Successful `check` run root | 1.1 GB | ≤ 150 MB |
| `unit-events.ndjson` per `check` | 159 MB | ≤ 10 MB |
| Temporary directories leaked per clean run | ~20 | 0 |
| Retained results | 122 GB, unbounded | Bounded by an owner-set budget |
| `ci`, `release-check` wall | Unmeasured | Measured; correctness portion of `release-check` ≤ 1.5 × `check` |

## 4. Gap Register

| ID | Gap | Areas | Workstream | Priority |
| --- | --- | --- | --- | --- |
| G1 | `gofmt` drift keeps `make check` red at HEAD | Implementation | WS0 | Immediate |
| G2 | Heavy-lane default too low (resolved in `6712f0408`) | Specification, implementation, tests | WS0 follow-up | Closed; revisit after WS3 |
| G3 | Heavy tests repeat expensive setup; a few oversized rows dominate | Tests, test support | WS3 | High |
| G4 | Graph admission ignores physical footprint (head-of-line blocking) | Specification, implementation, tests | WS2 | High |
| G5 | PostgreSQL concurrency emerges from process-slot arithmetic and double-charged lanes | Specification, machine inputs, implementation | WS2 | High |
| G6 | Exclusive PostgreSQL generation startup and residency cost | Specification, implementation | WS2, WS3 | Medium |
| G7 | Retained-secret scan costs ~120 s per `check` | Implementation, tests | WS4 | High |
| G8 | Run roots retain bulk fixture bytes on success | Specification, test support | WS4 | High |
| G9 | Wait-segment and admission-record churn | Specification, schemas, implementation | WS4 | Medium |
| G10 | Temporary-directory leaks, including private keys | Implementation, tests, specification | WS5 | High (security hygiene) |
| G11 | Unbounded retained-result growth | Specification, implementation | WS5 | Medium |
| G12 | Go build-cache growth without a budget | Implementation, specification | WS5 | Low |
| G13 | `ci`/`release-check` unmeasured; quiet measurement runs inline in aggregates | Measurement, specification, process | WS1, WS6 | High |
| G14 | Estimates drift and diagnostics truncate | Machine inputs, implementation | WS1, WS4 | Medium |

## 5. Gap Remediation Detail

### G1 — `gofmt` drift at HEAD

- **Evidence.** Both measured runs failed only `operation:harness.lint_go_format` on `internal/modules/links/internal/sourcestate/prepare.go`.
- **Remediation.** Run `make format`, review the diff, commit.
- **Area.** Implementation.
- **Rationale.** Every later measurement needs a passing `check`; otherwise each comparison carries a known failure.
- **Long-term benefit.** A green baseline makes regressions visible.
- **Compatibility.** None.
- **Risk if unresolved.** Every aggregate stays red, and a real failure can hide behind the known one.
- **Validation.** `make lint` passes; the next `make check` reports `status=pass`.

### G2 — Heavy-lane default (closed)

- **Evidence.** 2 → 3 lanes cut graph wall 1,341 s → 1,164 s. Committed as `6712f0408` in the NLSpec Section 10.2 capacity table, the registry default and `test-compilation.mjs`.
- **Remediation.** None now. After WS3 shrinks heavy work, re-measure 3 against 4 lanes once.
- **Area.** Specification and machine input, already changed.
- **Rationale.** Contention is I/O-bound, so the lane count stays an authored policy, not a CPU-derived one.
- **Long-term benefit.** The default reflects measured contention instead of a guess.
- **Compatibility.** Override files that set `postgres_heavy_lanes` still apply.
- **Risk if unresolved.** Not applicable.
- **Validation.** The WS7 re-baseline records 3- and 4-lane walls and per-row slowdown.

### G3 — Heavy test cost

- **Evidence.** Seven rows are 45% of heavy-row time. Reference-data support signs a fresh 95 MB ATT&CK container per test. The envelope embeds expiry times derived from the caller's `now` (`signDataset`), so each test produces a distinct digest; six per run. `imports.integration.current_workflows` is one 226 s row, the longest unbreakable item in the heavy tail.
- **Remediation.**
  1. Profile the seven rows: Go CPU profile, per-subtest timing and PostgreSQL statement statistics. Classify each cost as setup that recurs, setup unique to the test, or the behavior under test.
  2. Build immutable fixtures once per test process from deterministic inputs: fixed signing time and key, content-addressed in memory. Identical inputs yield one signature and one digest.
  3. Where several tests need the same seeded database state and do not test the seeding itself, seed once into a sealed snapshot and clone it, reusing the existing performance-fixture snapshot mechanism (builder unit, fixture profile, snapshot key) instead of adding a new one.
  4. Split rows over the 225 s heavy packing target into independently selectable rows along real behavioral seams, so LPT packing can balance the tail.
- **Area.** Tests and test support; catalog rows in `tools/test_families/**`; no harness specification change.
- **Rationale.** Heavy work is the critical path: wall ≈ heavy work ÷ lanes. Removing repeated work is the only lever that shortens the path without adding contention.
- **Long-term benefit.** Faster heavy tests also speed `test-slice` runs and leave room in the lane budget as the suite grows.
- **Compatibility.** Row splits change row IDs; follow the existing row-migration process (`tools/test_catalog_row_migrations.json`). Deterministic signing must not weaken tests that assert time-bound signature behavior; those keep their explicit clock.
- **Risk if unresolved.** Heavy work grows with every new import or recovery feature and `check` grows with it.
- **Validation.** Heavy-row time in `check` drops by at least 30% at unchanged row coverage. Each changed row keeps its assertions, verified by the owner's targeted rows. ATT&CK container digests are identical across tests in one run.

### G4 — Graph admission ignores physical footprint

- **Evidence.** Up to 12 of 19 CPU tokens held by units blocked in host admission; 197 CPU-only units waited more than 300 s while CPUs were idle. The compiler already computes combined physical allocations per unit ([model.mjs](../../tools/harness/scheduler/work-graph/model.mjs) around line 155), but the engine does not use them for admission.
- **Remediation.** Make graph-local admission footprint-aware. A unit's admission footprint is its active claims plus the prospective resident claims it would create (an exclusive cluster, a first shared service). The engine starts a unit only when the footprint fits the graph's view of capacity, and otherwise keeps ranking smaller fitting work. Host admission stays the cross-process authority (TH-HARNESS-REQ-381 and TH-HARNESS-REQ-803 oldest-waiter priority), so the two queues stay separate as the observation rules require.
- **Area.** Specification (Section 2.1, REQ-803; Section 10.2), implementation (`engine.go`, `model.mjs`), tests (scheduler self-tests and engine tests).
- **Rationale.** One claim model, read from the topology, drives both queues. The graph stops making promises the host will refuse, which removes the blocking at its source instead of compensating for it.
- **Long-term benefit.** Adding a service type or resident profile changes one topology entry, and the scheduler follows automatically. Starvation becomes structurally impossible within one invocation.
- **Compatibility.** Unit events keep their vocabulary. Infeasible units stay queued instead of emitting `started` early. The admission-record schema gains footprint fields (version bump shared with WS4).
- **Risk if unresolved.** Every aggregate that mixes PostgreSQL and CPU-only work loses 10+ minutes of parallelism, and adding more service-backed rows makes it worse.
- **Validation.** No CPU-only unit waits more than 30 s while CPU tokens are free. The `check` wall falls with heavy work held constant. Scheduler fairness tests (oldest-waiter, exclusive barrier, rank) pass. New engine tests cover a footprint-infeasible unit not blocking a fitting one.

### G5 — PostgreSQL concurrency as an emergent property

- **Evidence.** Exclusive units are charged a PostgreSQL lane twice and 9 process slots against a pool derived from CPU count. The effective cap of about 8 concurrent exclusive units appears nowhere as a stated policy.
- **Remediation.** Introduce an explicit `postgres_clusters` capacity for concurrent exclusive generations, with an authored default and a memory-derived bound, in the Section 10.2 capacity table. Charge a resident cluster against it instead of `postgres` lanes. Keep `postgres` lanes for active connections only. Restate `process_slots` as an OS-process ceiling derived from the effective process limit, so resident claims stop competing with active work on an arbitrary 4 × CPU pool.
- **Area.** Specification (Section 10.2, REQ-804 resident charging), machine inputs (registry, topology claims, schemas: registry v8 → v9, capability snapshot v1 → v2, override v1 → v2), implementation (`capability.mjs`, `scheduler-resources.mjs`, native admission), tests.
- **Rationale.** Capacity that matters should be named and owned, so it can be tuned and reasoned about.
- **Long-term benefit.** Each knob means one thing. Hosts with more memory can run more clusters without touching unrelated pools.
- **Compatibility.** Override files using v1 must be rewritten; no dual schema support. Retained snapshots stay historical.
- **Risk if unresolved.** Tuning any pool has unpredictable effects on PostgreSQL concurrency, and the hidden cap persists after G4.
- **Validation.** Capability snapshot reports `postgres_clusters` with its source. Admission diagnostics show no `process` or `postgres` deficits caused by resident double charging. With G4, PostgreSQL-backed concurrency reaches the authored cluster cap.

### G6 — Exclusive generation startup and residency

- **Evidence.** Fixture acquisition averages about 10 s per service-backed unit (4–28 s on heavy units, which sit on the critical path). Single-use generations are retired asynchronously after terminal release, and admission still logged 68 sampled pressure-retirement phases. The execution index is capped at 4,096 records, so no phase breakdown of startup exists yet.
- **Remediation.**
  1. Add allocation-lifecycle phase records for container start, readiness, role initialization, template migration, sealing checkpoint and retirement (WS1 profiling, made affordable by WS4 volume cuts).
  2. If template migration dominates, seed new exclusive generations from a sealed, migrated data directory keyed by provisioning identity (image digest, migration digest, role policy). Sealing and its checkpoint already exist (REQ-804); the snapshot persists that sealed state across generations.
  3. If retirement lag holds capacity, release a single-use generation's claims at settlement of its terminal release and make retirement concurrency an authored bound.
- **Area.** Specification (REQ-804 provisioning identity and retirement), implementation (`harnessfixtures`), tests.
- **Rationale.** Startup cost is paid 125 times per `check`. Caching the deterministic part removes it without sharing mutable state between units.
- **Long-term benefit.** Adding migrations no longer increases every exclusive unit's startup.
- **Compatibility.** Provisioning identity gains a snapshot digest; attached sessions with older identity are rejected and rebuilt.
- **Risk if unresolved.** Startup grows with schema size and multiplies with exclusive row count.
- **Validation.** Median exclusive acquisition ≤ 3 s. Template isolation and durable-WAL acceptance criteria in REQ-804 still pass.

### G7 — Retained-secret scan cost

- **Evidence.** ~120 s after every `check`; single-threaded; per-secret substring search; a new slice allocated per 64 KiB chunk.
- **Remediation.** Keep the walk and its per-entry safety checks (ownership, symlink refusal, mode, link count, stability). Scan file contents in a bounded worker pool. Replace per-secret `bytes.Contains` with one multi-pattern automaton built once per scan. Reuse buffers.
- **Area.** Implementation (`harnesssession/scan_linux.go`), tests. TH-HARNESS-REQ-600 already permits this; no specification change.
- **Rationale.** The scan is mandatory and must stay complete; making it efficient is the only acceptable lever besides retaining fewer bytes (G8).
- **Long-term benefit.** Scan cost scales with bytes ÷ cores instead of bytes × secrets.
- **Compatibility.** Verdicts and the `cartulary.harness_retained_secret_scan.v1` record are unchanged.
- **Risk if unresolved.** Every graph target pays a cost proportional to retained bytes, and it grows as fixtures grow.
- **Validation.** Same verdicts on the redaction and scan fixtures behind TH-HARNESS-AC-011 and TH-HARNESS-AC-160, plus new cases for chunk-boundary matches and many secrets. Post-graph time ≤ 15 s on a successful `check` root after G8.

### G8 — Bulk fixture bytes retained on success

- **Evidence.** Six 95 MB ATT&CK packs per run; Network Flow materializations up to 113 MB, which TH-HARNESS-REQ-407 requires under the run root until result-root cleanup.
- **Remediation.**
  1. Reference-data support records each envelope's manifest, digests and signature metadata in the run root and writes pack bytes only when the owning test fails. Deterministic signing (G3) collapses duplicates.
  2. Amend REQ-407 so a passing Network Flow fixture comparison keeps `actual.json`, transcript comparisons and digests, and releases the remaining materialization at unit finalization. A failing comparison keeps the full bundle.
- **Area.** Specification (REQ-407), test support, tests.
- **Rationale.** Evidence must let a reviewer judge the result. A digest plus the compared outputs does that for a pass; the full bundle matters only for diagnosis.
- **Long-term benefit.** Run roots shrink tenfold, which also shrinks scan time, retention pressure and copy cost.
- **Compatibility.** Tools that read retained packs from passing runs, if any, must read the digest record. Search confirms before cut-over.
- **Risk if unresolved.** Disk growth of about 1 GB per `check` and scan cost both scale with fixture size.
- **Validation.** Successful `check` root ≤ 150 MB. A deliberately failing fixture comparison retains its full bundle. A deliberately failing reference-data test retains its pack.

### G9 — Wait-segment and admission-record churn

- **Evidence.** 110,873 wait-segment pairs per `check` (63,988 `resources`, 46,885 `rank`). TH-HARNESS-REQ-371 closes a segment whenever blocking unit IDs change, so every admission re-segments every waiting unit. The spec already tells reducers to compact holders. Host-admission records: 173,449 in the baseline and 311,463 with 3 lanes, of which about 3,200 fit the cap.
- **Remediation.** Amend REQ-371 so segment identity is reason plus blocking resource set, with blocking unit IDs captured when the segment opens. Emit host-admission decision records on state change (first denial, deficit change, grant), not on every re-evaluation. Bump `cartulary.harness_unit_event` to v5 and `cartulary.harness_host_admission_record` to v4, and cut over readers together.
- **Area.** Specification (REQ-371, observation records), schemas, implementation (engine, admission observation, readers), tests.
- **Rationale.** The record should describe why a unit waited, not every change in who else ran.
- **Long-term benefit.** Event volume scales with units instead of units × transitions. Diagnostics stop truncating, which unblocks G6 profiling.
- **Compatibility.** Readers drop v4/v3 support; older runs stay historical evidence. `explain-run` and `test-run-status` change in step.
- **Risk if unresolved.** Event volume grows quadratically with graph size; diagnostics stay sampled.
- **Validation.** `unit-events.ndjson` ≤ 10 MB for `check`. Queue totals and partitions equal the v4 values on a replayed graph. Admission diagnostics report zero omitted records on `check`.

### G10 — Temporary-directory leaks

- **Evidence.** About 20 leaked directories per clean `check` from five `_Process` tests, containing TLS private keys. 52,843 `go-build*` directories from interrupted runs. All test processes share one machine-wide `TMPDIR`/`GOTMPDIR`.
- **Remediation.**
  1. Give every graph invocation a private scratch root for `TMPDIR` and `GOTMPDIR`, removed by invocation cleanup and reclaimed by the stale janitor after dead-owner proof, like other harness runtime roots.
  2. Fix the re-exec pattern: the child helper runs its work and returns an exit code without calling `os.Exit` inside a test function. Any directory it needs comes from the parent's `t.TempDir()` by explicit argument. Add a shared test helper so new process tests follow it.
  3. After both land and with owner approval, remove the existing leaked directories once.
- **Area.** Specification (Section 13 runtime-root ownership), implementation (runner environment), tests (`internal/app/server`, `internal/app/operator`).
- **Rationale.** Scoping temporary space to an owner makes leaks impossible to outlive the invocation, whatever a test does.
- **Long-term benefit.** No private key material lingers on disk; no manual cleanup.
- **Compatibility.** Tests that assume a stable machine-wide temporary directory break visibly; none is known.
- **Risk if unresolved.** Unbounded disk growth and lingering secrets outside the retained-secret scan's reach.
- **Validation.** After a clean `check`, a cancelled `check` and a `test-slice`, no new entries exist under the machine temporary roots. The janitor reclaims a scratch root left by a killed invocation.

### G11 — Unbounded retained-result growth

- **Evidence.** 15,994 runs, 122 GB. `make clean` removes all default results or none, and the NLSpec forbids caller-supplied deletion lists.
- **Remediation.** Add a public retention command with an authored policy: maximum age, maximum total size and a minimum number of recent runs per target. Protect pinned roots: any root named by an evidence-roots file, release evidence or an explicit pin. Follow the Section 13.4 dry-run contract and report what would be removed. Optionally warn at run start when the budget is exceeded; never delete automatically.
- **Area.** Specification (Section 13), task surface, implementation, tests.
- **Rationale.** Evidence retention needs a lifecycle; all-or-nothing deletion pushes users to keep everything.
- **Long-term benefit.** Disk use stays bounded without risking roots that release evidence depends on.
- **Compatibility.** Additive command. `make clean` is unchanged.
- **Risk if unresolved.** Disk exhaustion. Users fall back to `make clean` and lose evidence they meant to keep.
- **Validation.** Dry-run lists exactly the expected roots on a synthetic tree. Pinned and referenced roots survive. The command refuses roots it cannot prove it owns.

### G12 — Go build-cache growth

- **Evidence.** 182 GB, about 40 GB per day of genuine churn; Go's trim works.
- **Remediation.** First measure cache bytes written per invocation and per build variant (race, tags, CGO setting). Remove variants that carry no semantic weight. Then add a bounded cache-maintenance command under the readiness-cache owner with a size budget, separate from `make clean`, which must not remove external caches.
- **Area.** Implementation, specification (Section 13 cache ownership).
- **Rationale.** Fewer variants reduce both cache growth and recompilation time; a budget handles the rest.
- **Long-term benefit.** Predictable disk use; faster cold builds.
- **Compatibility.** None for results.
- **Risk if unresolved.** Disk pressure on smaller hosts.
- **Validation.** Recorded bytes per variant before and after; cache size stays under budget across a week of normal use.

### G13 — Aggregate composition and missing measurements

- **Evidence.** No retained `ci` or `release-check` run. Planner estimates put quiet measurement at 7,324 s of `ci` and 60,423 s of `release-check`, run serially under the exclusive `host_activity` lock. Browser estimates are unreliable (the six a11y groups: 1,419 s estimated, 658 s observed in total).
- **Remediation.**
  1. Measure one `ci` and one `release-check` with diagnostics (WS1), on an otherwise idle host.
  2. Recalibrate authored estimates from qualified runs, as TH-HARNESS-REQ-865 permits.
  3. Decide with the Core 05 owner whether performance measurement stays inline. The recommended structure: `release-check` verifies measurement evidence from explicitly supplied retained roots through `release-evidence-verify` (2026-11-06 targeted release amendment), and measurement runs as its own explicit qualification command. `ci` runs no quiet measurement groups.
- **Area.** Measurement; specification (aggregate membership clauses); Core 05 alignment for benchmark evidence; machine inputs.
- **Rationale.** Correctness gates and performance qualification have different cadence and host requirements. Coupling them makes every correctness run pay for quiet, serialized measurement.
- **Long-term benefit.** `ci` and `release-check` become fast enough to run routinely; measurement keeps its quiet-host fidelity on its own schedule.
- **Compatibility.** Release procedures must name measurement roots explicitly. Existing aggregates lose members, recorded in the spec.
- **Risk if unresolved.** If the estimates hold, `release-check` takes many hours and is skipped in practice.
- **Validation.** Measured walls recorded. Release evidence verification fails without measurement roots and passes with qualifying ones.

### G14 — Estimate drift and diagnostic truncation

- **Evidence.** Execution index capped at 4,096 records; host-admission records about 2% retained; browser estimates off by about 2×.
- **Remediation.** After G9, confirm diagnostics fit their caps; raise caps only if a measured `check` still truncates. Run the existing estimate audit (`estimate_audit`, `heavy_lane_candidate`) on each qualified baseline and update authored estimates in the same change as the behavior that moved them.
- **Area.** Machine inputs, implementation.
- **Rationale.** Scheduling quality depends on estimates; diagnosis depends on untruncated records.
- **Long-term benefit.** LPT packing stays accurate as tests change.
- **Compatibility.** None.
- **Risk if unresolved.** Packing and ranking degrade silently.
- **Validation.** Zero omitted records on `check`; estimate audit reports no findings above its threshold.

## 6. Rejected Alternatives

| Alternative | Reason rejected |
| --- | --- |
| Turn off `fsync`, full-page writes or synchronous commit, or put PGDATA on tmpfs | Forbidden by TH-HARNESS-REQ-804, and recovery and restore tests would stop exercising real durability. |
| Four or more heavy lanes now | Three lanes already slowed heavy units 21–27% and restore rows up to 3×. Revisit only after G3. |
| Release logical CPU tokens while a unit waits in host admission | Causes thrash, starves large units and breaks oldest-waiter priority (REQ-803). G4 removes the cause instead. |
| Raise `process_slots` through the override file | Bounded at 4 × CPU by policy, treats a symptom and keeps the hidden cluster cap. |
| Skip the secret scan for large or binary files | A security regression under REQ-600. |
| Prune old results automatically inside every run | A surprising destructive side effect, contrary to Section 13's no-implicit-deletion stance. |

## 7. Workstreams, Sequencing and Exit Criteria

```text
WS0 ──┬─> WS1 ──┬─> WS2 ──┐
      │         ├─> WS3 ──┼─> WS7
      │         └─> WS6 ──┤
      ├─> WS4 ─────────────┤
      └─> WS5 ─────────────┘
```

WS2 and WS3 can run in parallel but are measured together, because both move the heavy tail. WS4 and WS5 need only WS0. WS4's G9 change should land before WS2's admission-record changes so the schema bumps happen once.

### WS0 — Baseline hygiene

- **Scope.** G1; confirm G2 in a passing run.
- **Dependencies.** None.
- **Steps.** `make format`; commit; one `make check HARNESS_DIAGNOSTICS=basic CARTULARY_HARNESS_CACHE_MODE=off` as the new reference.
- **Risks.** None material.
- **Exit criteria.** `make check` passes at HEAD; the reference run ID is recorded in this tracker.

### WS1 — Measurement and decision inputs

- **Scope.** G13 measurement, G14 audit, G3 and G6 profiling.
- **Dependencies.** WS0.
- **Steps.** One `ci` and one `release-check` with diagnostics on an idle host (schedule overnight). Profile the seven heavy rows. Add temporary local instrumentation for generation startup phases, or wait for WS4's G9 to lift truncation. Run the estimate audit.
- **Risks.** `release-check` may run for many hours. Bound it by measuring its measurement projection separately if the owner prefers.
- **Exit criteria.** Measured walls and critical paths for `ci` and `release-check`; per-row cost breakdowns for the seven heavy rows; generation startup breakdown; owner decisions D1–D5 recorded in [Section 10](#10-open-decisions-and-handoff).

### WS2 — Scheduler and admission model

- **Scope.** G4, G5, G6 residency.
- **Dependencies.** WS1 baseline; approved amendments A1–A3.
- **Steps.** Amend the NLSpec (Sections 2.1, 10.2), then registry, topology and schemas, then regenerate through `make generate`. Implement footprint-aware graph admission, then the explicit cluster capacity. Extend scheduler and engine self-tests before switching defaults.
- **Risks.** Admission deadlock or fairness regressions; capacity mis-tuning on smaller hosts. Mitigate with the existing fairness tests, new starvation tests and a two-host check if a second host is available.
- **Exit criteria.** No CPU-only unit waits more than 30 s while tokens are free; no resident double-charge deficits; scheduler, engine and admission self-test rows pass; `make source-check CHECK=generated` passes; `check` graph wall improves measurably at constant heavy work.

### WS3 — Test and fixture cost

- **Scope.** G3, G6 startup, deterministic fixtures for G8.
- **Dependencies.** WS1 profiles. G6 snapshot seeding needs amendment A3.
- **Steps.** Per row: remove repeated setup, introduce deterministic fixture construction, adopt snapshot clones where the test does not exercise seeding, split rows over the packing target with catalog migrations. Then implement snapshot-seeded generations if WS1 shows migration dominates startup.
- **Risks.** Weakening a test's behavioral coverage. Each change keeps its assertions and runs its owner's targeted rows plus a mutation check where practical.
- **Exit criteria.** Heavy-row time down at least 30%; median exclusive acquisition ≤ 3 s if seeding lands; affected owners' targeted rows pass.

### WS4 — Evidence pipeline efficiency

- **Scope.** G7, G8, G9, G14 diagnostics.
- **Dependencies.** WS0. G8 needs amendment A5; G9 needs A4.
- **Steps.** G7 first (implementation only). Then G9 schema cut-over with all readers. Then G8 retention changes in test support.
- **Risks.** Scan regressions are security regressions. Gate G7 on the redaction acceptance fixtures and new boundary cases. Reader cut-over can break `explain-run`; update readers in the same change.
- **Exit criteria.** Post-graph overhead ≤ 15 s; successful `check` root ≤ 150 MB; events ≤ 10 MB; zero omitted diagnostic records on `check`.

### WS5 — Host hygiene and retention

- **Scope.** G10, G11, G12.
- **Dependencies.** WS0. G10 and G11 need amendment A6.
- **Steps.** Fix the re-exec tests (no amendment needed) first. Then invocation-scoped scratch roots, then the retention command, then build-variant measurement and the cache budget. Finally, with owner approval, a one-time removal of existing leaks and old results through the new commands.
- **Risks.** Destructive commands. Use the dry-run contract, ownership proofs and pin protection; never delete without dry-run output reviewed.
- **Exit criteria.** Zero leaked temporary entries after clean and cancelled runs; retention dry-run and apply pass their tests; result and cache totals under the owner's budgets.

### WS6 — Aggregate composition

- **Scope.** G13 structural change.
- **Dependencies.** WS1 measurements; decision D3; amendment A7.
- **Steps.** Amend aggregate membership; move quiet measurement to its explicit qualification command; extend `release-evidence-verify` inputs to require measurement roots; recalibrate estimates.
- **Risks.** Release gating gaps if measurement evidence becomes optional by accident. Verification must fail closed without qualifying roots.
- **Exit criteria.** `ci` contains no quiet measurement groups; `release-check` correctness portion ≤ 1.5 × `check`; release verification fails closed without measurement roots.

### WS7 — Validation and handoff

- **Scope.** Whole programme.
- **Dependencies.** All other workstreams that were approved.
- **Steps.** Re-baseline `check`, `ci` and `release-check` with the Section 9 protocol; re-measure 3 against 4 heavy lanes (G2); update `.agents/skills/cartulary-test-ops` recipes if public commands changed; record final numbers here.
- **Risks.** Host noise masking small gains; use two repetitions per treatment.
- **Exit criteria.** Section 3 targets met or each miss explained with data; every amendment adopted with its projections regenerated; tracker updated with run IDs.

## 8. Specification Amendment Ledger

| ID | Section or requirement | Change | Gap | Schema impact |
| --- | --- | --- | --- | --- |
| A0 | §10.2 capacity table | Heavy lanes default 3 (adopted, `6712f0408`) | G2 | None |
| A1 | §10.2; REQ-804 resident charging | Add `postgres_clusters`; charge resident clusters there; `postgres` lanes for connections only; restate `process_slots` | G5 | Registry v9, capability snapshot v2, override v2 |
| A2 | §2.1 REQ-803; §10.2 | Graph-local admission uses the unit's footprint (active plus prospective resident claims); host admission stays authoritative | G4 | Host-admission record v4 (with A4) |
| A3 | REQ-804 | Snapshot-seeded provisioning identity; single-use generation claims released at terminal settlement | G6 | Provisioning identity fields |
| A4 | REQ-371; observation records | Segment identity = reason + blocking resources; admission records on state change | G9 | Unit event v5, host-admission record v4 |
| A5 | REQ-407 | Passing Network Flow comparisons keep outputs and digests; bulk materialization released at unit finalization | G8 | None expected |
| A6 | §13 | Invocation-scoped `TMPDIR`/`GOTMPDIR`; retention command with pins and dry-run; cache-maintenance command | G10–G12 | New command identities in the task surface |
| A7 | Aggregate membership; Core 05 alignment | Quiet measurement leaves `ci`/`release-check`; release verification requires measurement roots | G13 | Aggregate projections |

Each amendment lands spec-first in one change with its machine projections, regenerated outputs (`make generate`) and the tests that exercise it, following [tools/AGENTS.md](../../tools/AGENTS.md).

## 9. Validation Protocol

1. **Host state.** No other harness invocation, no running harness containers, normal background load.
2. **Command.** `CARTULARY_TEST_RUN_ID=<fresh-id> CARTULARY_OUTPUT_MODE=ci make check HARNESS_DIAGNOSTICS=basic CARTULARY_HARNESS_CACHE_MODE=off`, with an override file only when the treatment is a capacity change.
3. **Repetitions.** Two per treatment, sequential, same commit apart from the change under test.
4. **Comparison.** `make explain-run RESULTS_DIR=<run> DETAIL=performance JSON=1 COMPARE_RESULTS_DIR=<reference> COMPARISON=equivalent`, plus the metrics in Section 3. The comparison is descriptive, never a gate (REQ-865).
5. **Correctness.** Each change runs its targeted rows per [AGENTS.md](../../AGENTS.md). Spec and machine-input changes also run `make source-check CHECK=generated` and, for documents, `make lint-markdown`. A full `check` runs once per workstream exit, `ci` at WS6 and WS7 exit.
6. **Recording.** Append run IDs, walls and deltas to Section 11 of this tracker.

## 10. Open Decisions and Handoff

| ID | Decision | Needed by | Recommendation |
| --- | --- | --- | --- |
| D1 | Approve amendments A1–A7 individually | Each workstream's start | Approve A4 and A2 first; they unblock diagnostics and the largest structural gain. |
| D2 | Retention budget defaults and what counts as pinned | WS5 | Age 14 days, 30 GB, keep the 5 newest per target; pin anything named in evidence-roots files. |
| D3 | Whether quiet measurement leaves `ci`/`release-check` | WS6 | Yes, with fail-closed release verification. Needs Core 05 owner alignment. |
| D4 | One-time removal of existing leaked directories and old results | WS5 exit | Yes, through the new commands' dry-run then apply. |
| D5 | Confirm or adjust the Section 3 targets | WS1 exit | Keep, then revise with `ci`/`release-check` data. |

**Handoff checklist for the next session.**

- Read this tracker, the [test-operations skill](../../.agents/skills/cartulary-test-ops/SKILL.md) and the NLSpec sections in the amendment ledger.
- Start with WS0; do not start an amendment-dependent workstream without its decision recorded above.
- Keep measurement runs sequential and never attach to the newest run implicitly.
- Record every run ID used as evidence in Section 11.

## 11. Execution Log

| Date | Workstream | Change | Evidence run IDs | Result |
| --- | --- | --- | --- | --- |
| 2026-11-22 | Baseline | Measurement, 2 heavy lanes | `eff-base-check01` | Graph 1,341 s; `lint_go_format` failed (G1) |
| 2026-11-22 | Baseline | Measurement, 3 heavy lanes via override | `eff-heavy3-check01` | Graph 1,164 s; `lint_go_format` failed (G1) |
| 2026-11-22 | G2 | Default heavy lanes 2 → 3 | `eff-heavy3-slice01`, `eff-heavy3-generated01`, `eff-heavy3-mdlint01` | Committed `6712f0408` |
