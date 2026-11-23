# Test Execution Efficiency Remediation Plan

**Status.** Active remediation tracker, last updated 2026-11-23. All phase-1 work is committed. Phase 2 ([Section 12](#12-phase-2-plan--check-wall-clock)) landed G7 and G6; G6 brought the `check` graph wall from 1,022–1,157 s to under 900 s. In G17 (phase 3), G17b and G17c landed, G17a was not adopted, and disk stall was identified as the contention the failing tests hit. G17f withdrew A17 and added A18 (WAL compression, −22% container writes), and G17g bounded test-side waits by the test deadline. G17 is closed; follow-ups are G27 (heavy-row test work; its workflow-test fix landed) and G28 (flush-bound disk stall). `make distclean` on 2026-11-22 deleted every retained run, so the run IDs in Section 11 before that date are a record, not retrievable evidence.

The active plan for the remaining `check` wall-clock work is [Section 12](#12-phase-2-plan--check-wall-clock). Original framing: this document authorizes no implementation by itself.
Each specification amendment in [Section 8](#8-specification-amendment-ledger) needs
owner approval before its workstream starts. Workstream 0 is the only work that
needs no new decision.

## 1. Scope, Authority and Evidence Posture

| Item | Baseline |
| --- | --- |
| Goal | Reduce wall time and host resource use of `make check`, `ci` and `release-check`, plus the shared harness costs every graph target pays. |
| Checkout | `/home/jochi/code/cartulary`, branch `main`, HEAD `6712f0408` ("Increase default PostgreSQL heavy lane capacity") at baseline; HEAD `500e3bdda` on 2026-11-22. |
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
- **Aggregates (planned).** `ci` plans 506 units and adds 33 browser units (about 15,800 s estimated), including 4 quiet measurement groups (7,324 s). `release-check` plans 585 units, including 33 quiet measurement groups of about 1,831 s each (60,423 s). Quiet groups take the exclusive `host_activity` lock ([browser.mjs](../../tools/harness/scheduler/work-graph/browser.mjs) lines 395–396), so they run alone.
- **`ci` (measured, `ws1-ci01`).** Graph wall 3,279 s; outer 3,464 s, of which about 181 s follows the graph (scan and publication of a 1.4 GB root). 505 of 506 units passed; the failure is unrelated (G15). The four quiet measurement groups took 96 s in total, not the 7,324 s estimated. The host sat nearly idle instead: average admitted concurrency 3.9 units, 32% of CPU tokens in use, 3 or fewer units running for 2,000 of 3,279 s, and 27,061 unit-seconds spent waiting in host admission. At 2,000 s, 11 CPU-only units (no PostgreSQL) waited in host admission while 2 units ran. The critical path ends in `row:harness.release.behavior.package_readiness_contract` (4 CPU, 20 process slots), eligible at 0.8 s but started at 3,165 s after 2,976 s of graph-level waiting for CPU (2,822 s), process (783 s) and the quiet-measurement barrier (285 s). Summed fixture time was 2,817 s. `unit-events.ndjson` reached 305 MB, and the admission log kept 3,105 records and dropped 2,269,576, so it covers only the first 669 s.

## 3. Principles and Targets

Decision rules applied to every gap:

1. Fix the model that produces a symptom, not the symptom. Capacity should be stated explicitly, not emerge from arithmetic side effects.
2. Retained data must earn its bytes. Keep evidence needed to judge a result; keep bulk bytes only for failures.
3. Schema changes cut over. Older retained runs stay historical evidence and are not translated, matching the NLSpec's current-epoch reader rule.
4. Measure before and after every change with the protocol in [Section 9](#9-validation-protocol).
5. Durability, isolation scope and secret-scan coverage are not traded for speed.

Targets are hypotheses to confirm against measurements, not commitments. "Current"
is the WS0 reference run `ws0-baseline-check01` (commit `1cd7d19e7`, default 3 heavy
lanes, passing). The earlier 3-lane override run measured 1,164 s graph wall on the same
graph, so expect about ±150 s between single runs; compare treatments with two
repetitions each.

| Metric | Current | Target |
| --- | ---: | ---: |
| `check` graph wall | 887 s and 889 s (`g6-check01`, `g6-check02`); was 1,022–1,157 s before G6 | ≤ 750 s |
| `check` outer wall | 899 s and 900 s (`g6-check01`, `g6-check02`) | ≤ 800 s |
| Post-graph overhead per invocation | 11–12 s (outer minus graph, `g6-check01`, `g6-check02`; was ~60 s, ~121 s before G8) | ≤ 15 s (met) |
| CPU-only units waiting more than 300 s | 55 and 96 with more than 300 s of `resources` wait (`g6-check01`, `g6-check02`; 119 in `w27-check01`). They queue for CPU tokens while the heavy lanes set the wall. Earlier figures used another definition | 0 |
| Successful `check` run root | 249 MB and 242 MB (`g6-check01`, `g6-check02`; was 1.1 GB) | ≤ 150 MB |
| `unit-events.ndjson` per `check` | 185 MB and 178 MB (`g6-check01`, `g6-check02`; 113,002 wait-segment pairs, G9 stage 2) | ≤ 10 MB |
| Temporary directories leaked per clean run | 0 (G10; was ~20) | 0 (met) |
| Retained results | Reset by `make distclean` 2026-11-22; still unbounded | Bounded by an owner-set budget |
| `ci` graph wall | 3,279 s (`ws1-ci01`; not re-measured) | ≤ 1,500 s |
| `release-check` wall | Unmeasured | Measured; correctness portion ≤ 1.5 × `check` |

## 4. Gap Register

| ID | Gap | Areas | Workstream | Priority |
| --- | --- | --- | --- | --- |
| G1 | `gofmt` drift keeps `make check` red at HEAD (resolved in `3bd5875af`) | Implementation | WS0 | Closed |
| G2 | Heavy-lane default too low (resolved in `6712f0408`) | Specification, implementation, tests | WS0 follow-up | Closed at three lanes: four lanes rejected twice, before and after G6 (Section 12, G2) |
| G3 | Heavy tests repeat expensive setup; a few oversized rows dominate | Tests, test support | WS3 / W2-2, W2-3 | Closed for phase 2: fixes 1–3 and 2b committed (`500e3bdda`); G3.5, G3.6 and G3.9 measured no gain and were dropped; G3.7 rejected; G3.4 deferred until G17 gives per-subtest timings |
| G4 | Graph admission ignores physical footprint (head-of-line blocking) | Specification, implementation, tests | WS2 | G4a (host backfill) committed (`500e3bdda`); G4b (graph footprint) deferred: after G5a, host admission no longer blocks `check` |
| G5 | PostgreSQL concurrency emerges from process-slot arithmetic and double-charged lanes | Specification, machine inputs, implementation | WS2 | G5a (process slots from the OS limit) committed (`500e3bdda`); G5b (single lane) reverted |
| G6 | Exclusive PostgreSQL generation startup and residency cost | Specification, implementation | W2-4 | Done (`b42e9bf20`): seeded generations and harness-assigned identity; `check` graph 887–889 s |
| G7 | Retained-secret scan costs ~120 s per `check` | Implementation, tests | WS4 / W2-6 | Done (`75b5ffea5`): memoized matcher, 21.7 s → 0.72 s on a full root |
| G8 | Run roots retain bulk fixture bytes on success | Specification, test support | WS4 | Done (`a5df06952`) |
| G9 | Wait-segment and admission-record churn | Specification, implementation | WS4 | Stage 1 committed (`500e3bdda`); stage 2 (wait segments) open: G4/G5 did not reduce it, and it is now most of each run root |
| G10 | Temporary-directory leaks, including private keys | Implementation, tests, specification | WS5 | Done (`a5df06952`); zero leaks after the G6 checks |
| G11 | Unbounded retained-result growth | Specification, implementation | WS5 | Medium |
| G12 | Go build-cache growth without a budget | Implementation, specification | WS5 | Low |
| G13 | `release-check` unmeasured; quiet measurement runs inline in aggregates | Measurement, specification, process | WS1, WS6 | Medium (`ci` measured: measurement is 96 s) |
| G14 | Estimates drift and diagnostics truncate | Machine inputs, implementation | WS1, WS4 | Medium; heavy-row estimates rebalanced (`fb5c7b63d`), little wall effect; estimates predate G6 |
| G15 | Fallow web dead-code policy fails `ci` on 5 unused exports | Implementation (frontend) | WS0 follow-up | Immediate for a green `ci` |
| G16 | Admission waiters re-run a full cross-process transaction every 50 ms | Implementation, specification | WS2 | High |
| G17 | Heavy-lane critical path: subtest timings are discarded; process tests used fixed hang guards; disk stalls (`guest:io` `full` about 10% of the run) are the binding contention, and template-clone I/O (about 38 of 62 GB per `check`) is unattributed | Specification, machine inputs, implementation, tests | Phase 3 | Closed: G17b, G17c, G17g and A18 committed; G17a not adopted; A17 withdrawn; follow-ups G27 and G28 (Section 5) |
| G18 | Retained ATT&CK release metadata expires 2026-12-08T20:40:00Z; tests import it against the real clock | Fixtures, tests | Separate task | Urgent (outside efficiency scope) |
| G19 | `DROP DATABASE` checkpoints can exceed the fixed 35 s cleanup budget under I/O load, poisoning the unit's generation | Specification, implementation | WS2/WS3 | Done: retirement and hardening (`500e3bdda`), retirement event (`a5df06952`); G6 removed most checkpoint backlog |
| G20 | A Git repository nested in the checkout (such as an agent worktree under `.claude/worktrees/`) breaks source enumeration and is scanned by a whole-checkout filesystem walk | Specification, implementation | WS4 | Done (`500e3bdda`, regression fix `a5df06952`) |
| G21 | `clean`/`distclean` miss declared build outputs and harness scratch (root `dist`, reference-pack builders and module staging, `apps/web/.cartulary`) | Specification, implementation | WS5 | Done (`a5df06952`) |
| G22 | Readiness cache stamps vouch for external Docker images that may have been removed | Specification, implementation | WS5 | Done (`a5df06952`); regression: `installed-readiness.mjs` imports private browser modules, failing `harness.command_surface.behavior.owner_boundaries` (not in `check`; separate task) |
| G23 | The pinned browser-trust package was removed upstream; only the local readiness cache kept bootstrap working | Implementation | WS0 | Done (`a5df06952`) |
| G24 | Heavy units hold a heavy lane while queued in host admission behind ordinary PostgreSQL units | Specification, implementation | WS2 | Done (`fb5c7b63d`): about 150 s of `check` wall |
| G25 | Unattributed I/O regression: checkpoint sync and graph wall up since the morning's passing runs | Measurement, possibly implementation | W2-1 | Closed: the cost was per-generation provisioning checkpoints, amplified by host disk state; G6 removed it (checkpoint sync 1,156 s → 4 s) |
| G26 | Packing omits PostgreSQL generation acquisition | Specification, machine inputs, implementation | W2-5 | Medium; not implemented, see the G26 decision in Section 12 |
| G27 | Heavy-lane wall sits in a few long single tests (import workflows, ATT&CK production restore, Network Flow incident transfer, the two bundle round trips) | Tests (module owners) | Phase 3 follow-up | High for wall. Workflow test fixed (`7a62a435a`, −99 s heavy work); the rest is product work needing owner decisions (Section 5, G27) |
| G28 | Disk stall follows flush count, not bytes (A18 cut writes 22% without moving `io_full`); 21–35 s of CPU-stall bursts per `check` remain unattributed | Measurement, then specification | Phase 3 follow-up | Medium; measure WAL flushes per unit and short-lived child CPU first |

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

#### G3 profile (2026-11-22)

Sources: row durations from five diagnostic `check` runs (`eff-base-check01` through `g5-check01`); per-database and per-generation timings from the test-services journals in `g5-check01/_shared/test-services/*/journals/*.ndjson` (`timing-span`, `postgres-db-created`, `postgres-db-dropped`, each with `duration_ms` and `test_name`); four read-only code investigations, with key claims re-checked against the code. Subtest-level timing is not retained on success (G17), so splits inside a single test function come from code structure, older failing-run logs and sibling-test estimates, and are marked as estimates.

**Where heavy time goes.** Heavy units ran 2,766 s in `g5-check01`. About 19% is infrastructure: container start 91 s and template migration 118 s across 17 exclusive generations, 266 s of `DROP DATABASE` across 479 databases, and 61 s of clones. The remaining ~81% is the tests' own work. Across all units in the run, container start totals 749 s (median 5.4 s) and template migration 1,070 s (median 9.8 s, up to 28.7 s under load); drops total 787 s (median 0.26 s, p99 8.0 s, max 24.3 s; the 43 drops over 2 s account for 332 s).

**Two kinds of slow row.** Stable rows (spread at most 1.3× across the five runs) are bound by their own work: `imports current_workflows` 217–251 s, `incidentbundles network_flow_portability` 186–226 s, `recovery attack_production_recovery` 138–177 s, `reference_data workspace_cleanup_lifecycle` 65–77 s. Contention-sensitive rows (1.5–2.7×) inflated when G5 let 15–18 units run beside the heavy lanes: incident-bundle export/import 56–152 s, ATT&CK pack portability 81–151 s, reference-pack activation 51–103 s, report-composition preview 87–149 s.

**Findings per test, ranked by expected saving on the heavy lanes.**

| # | Test (row) | Cause, with evidence | Fix | Expected saving | Needs |
| --- | --- | --- | --- | --- | --- |
| 1 | `TestReferenceWorkspaceOperationCleanup_Integration` (workspace_cleanup_lifecycle) | The `timeout` case blocks indexing and waits for the real verification deadline, configured to 60 s, the validator minimum (`workspace_lifetime_integration_test.go:36-70`, `config/validation.go:415`). Measured 60.1 s in an older log | Inject the coordinator's deadline/clock through the existing composition seam, or prove `verification_timeout` classification in a unit test; cancel and index-failure cases already prove scratch cleanup | ~59 s (measured) | Test support |
| 2 | `TestCurrentImportWorkflows_Integration` (current_workflows, 217–251 s) | 22 serial subtests, each a full deployment; 86 server starts and ~126 PBKDF2 (600,000-iteration) hashes; the replay loop restarts twice (`current_workflow_integration_test.go:238-254`); six fault subtests wait on the production job policy's 5 s retry plus 5 s recovery scan (`platform/jobs/manager.go:29-33`) | Replay once and reuse the session; give tests a short job policy (see #9) | ~25–35% of the row from replay (estimate); ~30 s from the job policy | Test change; #9 for timers |
| 3 | `TestApprovedMappingApplyAdmission` (approved_mapping, 59–91 s) | 2 profiles × 21 cases, each starting a full deployment (`approved_mapping_integration_test.go:102-124`): clone, bucket, runtime, bootstrap hash, drop | One deployment per profile, adding each case through the existing `addSource` seam with distinct IDs | ~85–95% of the row (estimate: ~2 s per case is setup) | Test change |
| 4 | ATT&CK tests (`attack_production_recovery`, `attack_pack_portability`, reference-data lifecycle, report-composition preview) | Every import loads 40,853 index rows and 14,490 lookup keys as single-row INSERTs in batches of at most 128 rows/keys, each batch its own transaction (`reference_data/canonical_index.go:80`, `writeIndexBatch`); tests import the same full pack 2–3 times; recovery restores it 4 times; signed containers differ only by random keys and `now`-based expiry, so 6 are built per run where 3 would be byte-identical | (a) Deterministic keys and fixed expiry, built once per process; (b) a sealed post-import snapshot for tests that do not test import; (c) fewer redundant full passes (lifecycle "authentic-replay" re-import, recovery fault probes) | (a) 3 builds and ~285 MB of retained bytes; (b)/(c) tens of seconds per test (estimate) | Test support; (c) needs owner judgment; batching itself is production behavior under the reference-pack NLSpec |
| 5 | `TestIncidentReferenceCatalogSurvivesUnavailablePackRoundTrip_Integration` | 12 servers; drops total 38.5 s; three identical plain sources; fault cohorts take ~5 s longer, consistent with one job retry | Share one plain source across reporting, embedded and main flows; job policy (#9) | ~10–20 s plus timers (estimate) | Test change; #9 |
| 6 | `TestNetworkFlowIncidentTransferThroughPublicOperations_Integration` (186–226 s) | One dependent chain: 4 large exports, 2 bundle imports, operator backup/restore; A-side reads repeated after A stops changing; page size 200 where the API allows 500 | Read A once; use limit 500; run the independent unclaimed-destination leg concurrently | ~10–20% (estimate) | Test change |
| 7 | Network Flow saved-graph restore targets | The same capture restored 14 times into fresh databases | Restore once, seal as a template (`pgtest` already has seal/clone helpers), clone per fault target | ~60–80% of the restore phase (estimate) | Test support |
| 8 | Every exclusive generation | Container start plus template migration ~15 s median per generation; 209 s on heavy units, 1,819 s per `check` overall | G6: seed generations from a sealed, migrated data directory keyed by provisioning identity | ~10 s per generation | Spec A3 |
| 9 | Fault tests across modules | Job retries wait the production `RecoveryScan` 5 s and `RetryDelays` 5 s/30 s | A test-scoped job policy through server options (production defaults unchanged) | ~30 s in imports, ~12–15 s in incident bundles (estimate) | Product test seam; owner decision |
| 10 | Per-test database drops in exclusive single-unit generations | 266 s on heavy tests (787 s per `check`), with 4–24 s stalls consistent with forced checkpoints; REQ-357 requires eager per-test retirement and REQ-804 preserves normal checkpoint behavior | Let databases inside a single-unit exclusive generation retire with the generation | Up to 266 s heavy work | Spec change (REQ-357); weigh against eager-cleanup intent |

Tier 1 (rows 1, 2-replay, 3, 5-sharing, 6) are test-local and need no spec change: about 200–280 s of heavy work, roughly 70–90 s of `check` wall at 3 lanes. Rows 4a, 7 are shared test-support changes. Rows 8–10 need owner or spec decisions.

#### G3 fixes applied (2026-11-22, committed `500e3bdda`)

| Fix | Change | Row before | Row after | Notes |
| --- | --- | ---: | ---: | --- |
| 1 | `TestReferenceWorkspaceOperationCleanup_Integration` keeps `content_rejection`, `cancel`, `index_failure` and the historical cases; the `timeout` case moves to `TestReferenceWorkspaceVerificationTimeout_Integration` in new row `module.reference_data.integration.verification_timeout_classifies_the_operation_an_4ef0203684` (profile `standard`, still exclusive and dedicated) | 65–77 s heavy | 4.0 s heavy + 61.5 s on an ordinary lane | Same assertions; row ID from `make author-test-row-id`; estimates 17 s and 70 s; `source-check CHECK=catalog` passes |
| 2 | `TestCurrentImportWorkflows_Integration` restarts once and replays twice on that server, instead of restarting before each replay | 194.9 s (isolated) | 185.8 s | Same assertions (replay after restart, repeated replay stable). Saves 20 server starts and logins, about 9 s; the code-read estimate of 25–35% was wrong |
| 2b | Contention robustness for the same test: seeded actor and session (`SeedActor` + `SeedSession`) instead of `SeedLocalUserFlags` and four password logins per subtest; `pg_locks` waits scoped to the test database; this test's 10–20 s waits raised to a 2-minute bound and the shared `importtest.WaitJobTerminal` bound raised from 5 s to 60 s | 185.8 s | 175.0 s | About 88 fewer 600,000-iteration PBKDF2 hashes per run. Waits still return as soon as state settles, so only a stuck run is affected. The 5 s job retry and recovery-scan waits (~45 s per run) remain; they need the job-policy seam (#9) |
| 3 | `TestApprovedMappingApplyAdmission` builds all 42 cases first and shares one deployment; each case gets its own session, unit and table IDs and idempotency keys, the injected change is armed only around its apply with counters reset, and the outcome check is scoped to its session. `publicationFixtureRuntime.bindTest` points failures at the running subtest | 59–91 s | 4.6 s | A temporary probe failing only case 42 proved all 42 cases run |

First combined measurement (`g3-check01`, fixes 1–3, before 2b): graph 959 s (984 s before); heavy work 2,386 s (2,766 s before, −14%). The run failed in an unchanged unit (G19).

### G4 — Graph admission ignores physical footprint

- **Evidence.** In `check`, up to 12 of 19 CPU tokens held by units blocked in host admission; 197 CPU-only units waited more than 300 s while CPUs were idle. In `ci`, the same blocking sets most of the wall: average concurrency 3.9, and the critical-path unit waited 2,822 s for logical CPU held largely by units stuck in host admission. CPU-only units also queue inside host admission behind larger waiters, which needs the untruncated admission log (G9) to attribute between resident occupancy and the oldest-waiter rule (REQ-803). The compiler already computes combined physical allocations per unit ([model.mjs](../../tools/harness/scheduler/work-graph/model.mjs) around line 155), but the engine does not use them for admission.
- **Attribution from the complete admission log (`g9-check01`).** The largest denial class has no resource deficit at all: 361 decision records, spanning 2,723 s of observed waiting, were denied only by queue priority. For CPU-only units it is 123 records and 1,482 s. Their persistent deficits are empty, so REQ-803 forbids backfill: fitting work may not bypass an older waiter that would fit once active work drains. Where a deficit exists it is mostly `process`; at those denials resident services hold a median of 66 of the 76 slots and active work holds 10. Blocking therefore has two layers: logical CPU tokens held by units waiting in host admission (graph level), and the oldest-waiter rule amplified by resident occupancy (host level).
- **Remediation.** Two parts. (a) Host level: amend REQ-803 to permit deficit-disjoint backfill. A fitting request may bypass an older waiter when its claims draw on none of the resources in that waiter's current deficit, so the waiter's progress is never delayed by the bypass. With G5 removing resident process-slot charges, CPU-only work becomes disjoint from PostgreSQL waiters. (b) Graph level: make graph-local admission footprint-aware.
- **Status.** (a) is committed (`500e3bdda`) as a deficit-non-increase rule rather than a disjointness test: a later request may go ahead of an older shared waiter only if admitting it increases none of that waiter's current deficits (computed with the same weighted evaluator, including the later request's prospective residents). Descendants of a branch admitted ahead of a waiter stay bound by that waiter, because weighted descendants can exceed their ancestor's claim; branches admitted before the waiter keep their drain entitlement, so no deadlock arises. Code in `mutate` in [admission_linux.go](../../internal/testutil/harnessadmission/admission_linux.go); new tests `deficit_disjoint_backfill_never_delays_an_older_waiter` and `backfilled_branch_descendants_respect_the_bypassed_waiter`; three existing tests updated where both requests fit after a release. Alone, (a) changed little (`g4a-check01`, 990 s) because nearly every waiter lacked process slots, which every unit claims. With G5 (`g5-check01`), host-admission waiting fell from 5,485 to 705 unit-seconds and only one or two units wait at a time, so (b) is deferred: re-evaluate it with `ci`, whose browser lanes add residents. A unit's admission footprint is its active claims plus the prospective resident claims it would create (an exclusive cluster, a first shared service). The engine starts a unit only when the footprint fits the graph's view of capacity, and otherwise keeps ranking smaller fitting work. Host admission stays the cross-process authority (TH-HARNESS-REQ-381 and TH-HARNESS-REQ-803 oldest-waiter priority), so the two queues stay separate as the observation rules require.
- **Area.** Specification (Section 2.1, REQ-803; Section 10.2), implementation (`engine.go`, `model.mjs`), tests (scheduler self-tests and engine tests).
- **Rationale.** One claim model, read from the topology, drives both queues. The graph stops making promises the host will refuse, which removes the blocking at its source instead of compensating for it.
- **Long-term benefit.** Adding a service type or resident profile changes one topology entry, and the scheduler follows automatically. Starvation becomes structurally impossible within one invocation.
- **Compatibility.** Unit events keep their vocabulary. Infeasible units stay queued instead of emitting `started` early. The admission-record schema gains footprint fields (version bump shared with WS4).
- **Risk if unresolved.** Every aggregate that mixes PostgreSQL and CPU-only work loses 10+ minutes of parallelism, and adding more service-backed rows makes it worse.
- **Validation.** No CPU-only unit waits more than 30 s while CPU tokens are free. CPU-only units are never denied for priority alone when they draw on none of the oldest waiter's deficit resources. The `check` wall falls with heavy work held constant. Scheduler fairness tests (oldest-waiter, exclusive barrier, rank) pass. New engine tests cover a footprint-infeasible unit not blocking a fitting one.

### G5 — PostgreSQL concurrency as an emergent property

- **Evidence.** Exclusive units are charged a PostgreSQL lane twice and 9 process slots against a pool derived from CPU count. The effective cap of about 8 concurrent exclusive units appears nowhere as a stated policy.
- **Remediation (G5a committed in `500e3bdda`; G5b reverted).** Two changes, chosen over the originally proposed `postgres_clusters` resource because each fixes the meaning of an existing resource instead of adding one:
  - **G5a.** `process_slots` now derive from the effective operating-system process limit (the smaller of the soft per-user limit and a positive cgroup `pids.max`) with a 25% margin, falling back to 4 × CPU tokens only when no limit is available. Section 10.2 capacity table amended; registry and schema cut over to `cartulary.scheduler_resource_registry.v9` (policy `process_limit_with_margin`), v8 retired. On this host: 95,394 slots instead of 76. Code: `processSlotsFor` and `processLimit` in [capability.mjs](../../tools/harness/scheduler/work-graph/capability.mjs); validation in [scheduler-resources.mjs](../../tools/harness/scheduler/scheduler-resources.mjs).
  - **G5b (reverted).** It stopped charging a Go unit a second `postgres` lane for its own exclusive generation. That doubled concurrent PostgreSQL work to about 16 units, saturated this host's disk (checkpoint sync per `check` 611–720 s before, 1,270–1,480 s after) and made I/O-heavy tests and cleanup budgets fail, without shortening the heavy-lane-bound wall. Lowering the lane default to 8 instead would throttle browser lanes (four lanes need 12). Reverting restores the calibrated accounting: about 8 concurrent Go units on 16 lanes, browser math unchanged. REQ-803 now states that the active and resident `postgres` claims are distinct and that the default lane count is calibrated for both.
- **Area.** Specification (Section 10.2, REQ-804 resident charging), machine inputs (registry, topology claims, schemas: registry v8 → v9, capability snapshot v1 → v2, override v1 → v2), implementation (`capability.mjs`, `scheduler-resources.mjs`, native admission), tests.
- **Rationale.** Capacity that matters should be named and owned, so it can be tuned and reasoned about.
- **Long-term benefit.** Each knob means one thing. Hosts with more memory can run more clusters without touching unrelated pools.
- **Compatibility.** Override files using v1 must be rewritten; no dual schema support. Retained snapshots stay historical.
- **Risk if unresolved.** Tuning any pool has unpredictable effects on PostgreSQL concurrency, and the hidden cap persists after G4.
- **Validation.** Capability snapshot reports `postgres_clusters` with its source. Admission diagnostics show no `process` or `postgres` deficits caused by resident double charging. With G4, PostgreSQL-backed concurrency reaches the authored cluster cap.

### G6 — Exclusive generation startup and residency

- **Evidence.** Measured from test-services journals in `g5-check01`: container start median 5.4 s (749 s over 89 generations) and template migration median 9.8 s, up to 28.7 s under load (1,070 s). Heavy units pay 209 s of this on the critical path. Single-use generations are retired asynchronously after terminal release.
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

- **Status (2026-11-22, uncommitted).** Done as amended REQ-407. `networkflowfixtures.Run` releases `inputs/` and `produced/` after a pass and records `materialized_input_root: null`; the summary keeps manifest, bundle and produced-artifact digests. Keeping `actual.json` was dropped: a pass means it matched the committed expectation, and it reaches 17 MB. `irfixtures` writes `comparison.json` (now with `inputs_retained`) and releases the admitted copy unless the test failed. Reference-pack fixtures always publish `public/bootstrap.json` under the container-digest directory and publish `failure/container.pack` only when the owning test fails, in a separate tree so passing and failing tests sharing a digest never conflict. Shared helper `fixtureartifacts.ReleaseTree` restores owner permissions on the sealed read-only copies before removal.

### G9 — Wait-segment and admission-record churn

- **Evidence.** 110,873 wait-segment pairs per `check` (63,988 `resources`, 46,885 `rank`). TH-HARNESS-REQ-371 closes a segment whenever blocking unit IDs change, so every admission re-segments every waiting unit. The spec already tells reducers to compact holders. Host-admission records: 173,449 in the baseline and 311,463 with 3 lanes, of which about 3,200 fit the cap.
- **Remediation.** Two stages.
  1. **Admission records (verified, uncommitted).** On `g9-check01` the host-admission section is `complete` with 0 omissions: 2,366 records (1,262 decisions) for 547,957 observations in 5.8 MB, against 186,482 omitted records in the reference run. `make harness-observability-check` passes with `full_validity=1`, and the critical-path `host-admission` view reports complete coverage. TH-HARNESS-REQ-857 now defines decision identity: request, mode, grant and transaction outcome, requested claims, effective capacities, required ancestors, the sets of resources in current and persistent deficit, and whether priority or exclusive-fence blockers exist. Consecutive decisions coalesce while identity holds; queue position, deficit amounts, residents, holders and blocker identities are context carried from the last observation. Zero claim and deficit quantities are omitted; capacities stay complete. The record shape is unchanged, so `cartulary.harness_host_admission_record.v3` and its readers stay; existing records are valid instances of the relaxed rule. Code: `decisionIdentity` and `compactDecision` in [admission_observation_linux.go](../../internal/testutil/harnessruntime/admission_observation_linux.go); test `TestRuntimeObservations/admission_decisions_coalesce_by_identity_not_retry_frequency`.
  2. **Wait segments (reopened).** Replaying `ws0-baseline-check01` showed that dropping blocking units from segment identity removes only 5% of segments (109,012 → 103,124), and keying on reason alone still leaves 93,276. The churn is waiting units flipping between `resources` and `rank` at each scheduling instant (93,000 of the transitions; up to 385 boundaries at one instant). I expected G4/G5 to shrink it by shortening waits; `g5-check01` disproved that (109,042 pairs, 175 MB) even though host-admission waiting fell 87%. The queue is inherent: 364 units contend for 19 CPU tokens for the first ~600 s, and every token release flips every ready unit. A remedy must change REQ-371's segment model itself, for example recording a ready unit's rank/resource alternation as one segment with per-reason totals; it needs its own design before any schema change.
- **Area.** Specification (REQ-371, observation records), schemas, implementation (engine, admission observation, readers), tests.
- **Rationale.** The record should describe why a unit waited, not every change in who else ran.
- **Long-term benefit.** Event volume scales with units instead of units × transitions. Diagnostics stop truncating, which unblocks G6 profiling.
- **Compatibility.** Readers drop v4/v3 support; older runs stay historical evidence. `explain-run` and `test-run-status` change in step.
- **Risk if unresolved.** Event volume grows quadratically with graph size; diagnostics stay sampled.
- **Validation.** Stage 1: host-admission section `complete` with zero omitted records on a diagnostic `ci`, and the retained-evidence readers accept the run. Event volume: `unit-events.ndjson` ≤ 10 MB for `check` after G4, with no change to segment semantics.

- **Footprint note (2026-11-22).** With G8 in place, the event log dominates a `check` root: in `post-distclean-check01` (diagnostics off), `unit-events.ndjson` is 163 of 208 MB, and 98% of it is 102,968 `wait_started`/`wait_ended` pairs. The REQ-371 wait-segment redesign is therefore also the main remaining disk lever per run.

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

- **Status (2026-11-22, uncommitted).** Cause found: Go's `t.TempDir()` creates roots under `GOTMPDIR` when set, not `TMPDIR`, so the five `_Process` tests' `TMPDIR=` override never applied and each child that exits through `os.Exit` leaked a TLS-key directory into the shared `GO_TMP_DIR` (7 per run of one test, measured in `g10-probe02`). Fixes: (1) the graph runtime creates `invocation-*` beneath `GO_TMP_DIR` with an owner proof, projects it as `GO_TMP_DIR` and `GOTMPDIR` to units and fixtures, removes it at settlement and reclaims dead-owner siblings (`harnessruntime/gotemp_linux.go`); (2) the `_Process` tests pass `GOTMPDIR` with `TMPDIR`. `TMPDIR` is deliberately not scoped: host admission and the session socket registries live there.

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

- **Evidence.** In `ws1-ci01` the four quiet measurement groups ran 96 s against 7,324 s estimated (about 76× too high), plus a 285 s drain before them. Planner estimates put `release-check` measurement at 60,423 s; at the `ci` ratio the real figure may be under 15 minutes, but that is unmeasured. Other browser estimates also miss (the six a11y groups: 1,419 s estimated, 658 s observed in total in an earlier run).
- **Remediation.**
  1. Measure one `ci` and one `release-check` with diagnostics (WS1), on an otherwise idle host.
  2. Recalibrate authored estimates from qualified runs, as TH-HARNESS-REQ-865 permits.
  3. Decide with the Core 05 owner whether performance measurement stays inline. The recommended structure: `release-check` verifies measurement evidence from explicitly supplied retained roots through `release-evidence-verify` (2026-11-06 targeted release amendment), and measurement runs as its own explicit qualification command. `ci` runs no quiet measurement groups.
- **Area.** Measurement; specification (aggregate membership clauses); Core 05 alignment for benchmark evidence; machine inputs.
- **Rationale.** Correctness gates and performance qualification have different cadence and host requirements. Coupling them makes every correctness run pay for quiet, serialized measurement.
- **Long-term benefit.** `ci` and `release-check` become fast enough to run routinely; measurement keeps its quiet-host fidelity on its own schedule.
- **Compatibility.** Release procedures must name measurement roots explicitly. Existing aggregates lose members, recorded in the spec.
- **Risk if unresolved.** Wrong estimates distort LPT ranking for every aggregate. The structural decoupling matters less for `ci` than first estimated; decide D3 after measuring `release-check`.
- **Validation.** Measured walls recorded. Release evidence verification fails without measurement roots and passes with qualifying ones.

### G14 — Estimate drift and diagnostic truncation

- **Evidence.** Execution index capped at 4,096 records. Host-admission records: about 2% retained in `check`, and 0.14% in `ci` (2,269,576 dropped; coverage ends at 669 s of 3,279 s), which hides the main `ci` bottleneck. Measurement estimates are about 76× too high; other browser estimates miss by about 2×.
- **Remediation.** After G9, confirm diagnostics fit their caps; raise caps only if a measured `check` still truncates. Run the existing estimate audit (`estimate_audit`, `heavy_lane_candidate`) on each qualified baseline and update authored estimates in the same change as the behavior that moved them.
- **Area.** Machine inputs, implementation.
- **Rationale.** Scheduling quality depends on estimates; diagnosis depends on untruncated records.
- **Long-term benefit.** LPT packing stays accurate as tests change.
- **Compatibility.** None.
- **Risk if unresolved.** Packing and ranking degrade silently.
- **Validation.** Zero omitted records on `check`; estimate audit reports no findings above its threshold.

- **Packing rebalance (2026-11-22, uncommitted).** From the qualified audit of `pack-base-check01`, 29 row estimates moved to the mean observed test time of two passing runs (`pack-base-check01`, `post-distclean-check02`), rounded up to whole seconds, wherever the change exceeded both 5 s and 25%. These were 22 heavy rows and the two drifted non-heavy cohorts' rows; 7 rows that used the 5 s evidence-class default gained their own estimates. The planner now splits the reference-data heavy rows into cohorts of 224 s and 146 s instead of one that overran at 302 s. Measured in `pack-rebal-check01`: makespan excess over a perfect three-lane split fell from 58 s to 42 s and partly idle lane time from 93 s to 66 s, but heavy work grew 3,263 → 3,375 s, so the graph wall moved 1,152 → 1,171 s, within noise. Packing is not a significant lever while G24 and fixture acquisition (G6) hold the heavy lanes; the estimates stay because they are more accurate.

### G15 — Fallow web dead-code policy failure

- **Evidence.** `operation:harness.frontend_fallow_static` failed in `ws1-ci01`. The bounded web policy found five symbols: `reconnectBaseDelayMs` and `reconnectMaxDelayMs` in `apps/web/src/collaboration/incidentCollaborationSessionPlan.ts`, `publicErrorStatusText` and `publicErrorCode` in `apps/web/src/shared/publicError.ts` (unused exports), and `HistoryActionOperation` in `apps/web/src/workbook/history/workbookHistoryOperation.ts` (unused type).
- **Remediation.** For each symbol, confirm it has no consumer, then remove the export, or the symbol if nothing uses it internally. Do not add suppressions; the policy forbids per-file suppression growth.
- **Area.** Implementation (frontend).
- **Rationale.** Same as G1: aggregates must be green before they can serve as measurement baselines.
- **Long-term benefit.** `ci` failures stay meaningful.
- **Compatibility.** None; the symbols are unused.
- **Risk if unresolved.** Every `ci` stays red.
- **Validation.** `make frontend-fallow-static` passes; `make lint CHECK=frontend` and the affected Vitest rows pass.

### G16 — Admission polling

- **Evidence.** A request that misses admission sleeps 50 ms and retries ([lease_linux.go](../../internal/testutil/harnessadmission/lease_linux.go) line 168). Each retry rebuilds logical residents, runs a cross-process transaction under `flock` (spinning in 1 ms steps on contention) and runs a pressure-retirement evaluation. `ws1-ci01` made about 2.34 million decision observations in 55 minutes, about 690 a second, with 10–17 waiters for long stretches.
- **Remediation.** Give the admission root a state generation that every committed grant, release or reclamation advances. A waiter re-runs its transaction only when the generation changes, when its own logical preparation changes, or after a bounded fallback interval that preserves dead-owner reclamation. Keep the oldest-waiter priority unchanged.
- **Area.** Implementation (`harnessadmission`), specification (REQ-803/REQ-381 waiting semantics), tests.
- **Rationale.** Admission can only change when shared state changes; retrying on a timer spends CPU and lock time to rediscover the same answer.
- **Long-term benefit.** Waiting cost becomes proportional to state changes, not waiters × time, so larger graphs and parallel invocations scale.
- **Compatibility.** Admission state root gains a generation file; the root is already versioned by generation, so a new root generation cuts over cleanly.
- **Risk if unresolved.** Admission overhead grows with the number of waiters and competes with test work for CPU and the admission lock.
- **Validation.** Same grant order on the scheduler fairness tests; admission transactions per `ci` drop by at least 10×; dead-owner reclamation tests pass.

### G17 — Heavy-lane critical path: attribution and contention

Rescoped 2026-11-23 from "Subtest timings are discarded" after phase 2. `check`
is now bound by heavy-lane test work (about 2,500 s ÷ 3 lanes). This entry holds
everything known about what that work waits on, and the fixes in order.

**Evidence.**

- *Attribution gap (the original G17).* The Go runner records every test and
  subtest's terminal action and elapsed time from `go test -json` (the
  `terminal` map in [go.mjs](../../tools/harness/execution/runners/go.mjs)) but
  keeps only top-level row durations. Passing runs retain no Go output,
  including in verbose mode (`g3-verbose-probe01`). TH-HARNESS-REQ-844 forbids
  test titles in the secure execution index.
- *Correction to the phase-2 close-out.* Heavy units do not slow each other on
  CPU. In `g6-check01` and `g2b-lane4-check02` the host uses 6–9 of 24 cores
  (`cgroup_context` samples in `diagnostics/resource-samples.ndjson`), with CPU
  stall (`guest:cpu` pressure, `some`) at 1–3% outside short bursts.
  Long-lived test processes average 0.3–0.6 cores. Heavy-lane time is
  latency-bound, and the disk still stalls work: `guest:io` `full` pressure is
  9–17% of the run. G6 removed the provisioning I/O, not the tests' own I/O.
- *CPU bursts from under-declared operations.* Static-analysis and build
  operations (`go_gosec_targeted`, `go_vulncheck`, `lint_go_staticcheck`,
  `lint_go_vet`, `lint_go_format`, `sourcecheck_build`, `codegen_toolchain`,
  `json_shape_check`, `reference_pack_runtime`) declare no resource claims in
  [harness_work_graph_owner.json](../../tools/harness_work_graph_owner.json),
  so each is admitted as one CPU token. The tools nonetheless run with default
  `GOMAXPROCS` across all 24 cores. When the block runs, the host jumps to
  14–21 cores with 21–65% CPU stall: 380–410 s in `g6-check01`, then 690 s and
  740–750 s in `g2b-lane4-check02`. The spec's rule that CPU contention is
  bounded by `cpu_tokens` (§10.2 `process_slots` row) does not hold for them;
  TH-HARNESS-REQ-379 bounds `GOMAXPROCS` only for Go test children.
- *Why the same two tests fail.* Both failures in `g2b-lane4-check02` (and in
  the pre-G6 four-lane pair) coincide with a burst. Four lanes scheduled those
  rows into the burst; on three lanes they ran in the quiet tail, overlapping
  5–6 units instead of 50–115. Both rows run single-processor: a PostgreSQL Go
  unit's `GOMAXPROCS` equals its one-lane PostgreSQL claim
  (TH-HARNESS-REQ-354), so the in-process app server shares that processor, and
  spawned operator processes inherit it. Each also has little slack:
  - `TestNetworkFlowPublicDefaultLimitsCommitAcceptedRowLimit` asserts a
    product claim: the default 250,000 accepted rows commit within the default
    30 s extension transaction-participant budget. The whole row takes 28–31 s
    on three lanes (`w27-check01`, `g6-check01`, `g6-check02`) and 34–36 s on
    four. In the failing run the two 25 MB source passes took 5.3 s and 24.1 s
    before `extension_transaction_timeout`. On three lanes the second pass
    takes 16–20 s (corrected after G17d; this said "instead of 5.3 s", which is
    the first pass).
  - `TestCanonicalOperatorRestoreBundle_Process/expired_source_retention` was
    killed by the test's own hang guard. `runOperatorBinary` kills the operator
    after a fixed 15 s; the call was a quick `stale_backup_unacknowledged`
    rejection that asserts no timing. The row takes about 31 s on three lanes
    and 78–90 s on four. Twenty-two calls share that 15 s guard
    (`operator_process_test.go` 15, `portable_restore_process_test.go` 5,
    `target_control_process_test.go` 2).
- *Where test I/O goes.* PostgreSQL containers write 62.4 GB per `check`
  (`g6-check01`; 70.9 GB in `w27-check01`) while using about 1,000 s of CPU in
  total. Template clones dominate: 1,302 per `check`, each `CREATE DATABASE …
  STRATEGY WAL_LOG` of a 19 MB template writing 9.8 MB of WAL plus 19 MB of
  dirty pages that a later checkpoint or the cluster's shutdown flushes. That is
  about 38 GB. Clones are spread at about one per test across roughly a
  thousand tests (the largest test makes 6), so there is no hotspot. Retired
  generations shut down gracefully, flushing dirty clone pages just before the
  volume is deleted. That cost is not yet measured.
- *Four lanes.* Heavy work rose 2,490–2,537 s → 3,019–3,026 s (runner time
  2,379–2,407 → 2,804–2,806 s) with checkpoint sync unchanged at 4 s, so the
  fourth lane's capacity goes to I/O contention and the CPU bursts.
- *Correction after the G17a validation.* The CPU bursts are not what the
  failing tests wait on, and operations are not their only source:
  - Total CPU stall is small: 21–35 stall-seconds per `check` in every run
    before and after G17a, concentrated in 3–5 windows of 10 s.
  - The harness samples attribute only 3–7 of the 13–19 busy cores in a burst.
    The rest belongs to short-lived children (compilers, workers) that start
    and exit between samples. Bounding operation `GOMAXPROCS` left the bursts
    in place, so their source is not yet identified.
  - The disk is the binding contention. `guest:io` `full` pressure averages
    8–12% of the run, with 9–33 windows of 10 s above 15%, often in the
    window after a write surge. During the failing step of each of the four
    post-G17a failures (two of them first seen in 20 retained runs), CPU stall
    was 0–10% and the host used 3–12 of 24 cores, while `io_full` was 13–27%:
    - `module.reference_data`, `TestRefreshAdmissionEligibilityMatrix_Integration`,
      at 188 s: the server went fatal (`extension_integrity_failure`) while
      cancelling a job blocked at a test barrier.
    - `module.incidentbundles`, the ATT&CK round-trip, at 79–109 s: an import
      job did not finish within the test's 30 s poll.
    - The network-flow row-limit test in both four-lane runs.
  - In `g2b-lane4-check02` the network-flow failure did end in a CPU burst
    (740 s) after 700–720 s at 12–15% `io_full`, so both pressures can push
    that row over its budget.

**Remediation**, in order:

1. **G17a — Bound operations to their CPU claim** (specification, machine
   inputs, implementation). Every operation unit declares a resource profile
   whose CPU claim reflects its real parallelism. The compiler gives it
   `GOMAXPROCS` (and, for `go` commands, `-p`) equal to that claim, as REQ-379
   already does for Go test children. Operations are off the heavy critical
   path, so bounding them costs no wall time. Amendment A16. **Not adopted:**
   validation failed (see Results); the attempt is kept on local branch
   `g17a-attempt` (`2f23c6e44`).
2. **G17b — Derive process-test hang guards from the test deadline** (tests).
   `runOperatorBinary` and similar helpers bound a child by the remaining
   `t.Deadline()` minus a settlement margin, not a fixed 15 s. Explicit product
   deadlines a test asserts stay as they are. No assertion changes.
3. **G17c — Per-subtest timings** (specification, implementation, tests; the
   original G17). Under `HARNESS_DIAGNOSTICS=basic`, retain a bounded per-unit
   artifact of subtest durations keyed by the row's catalog identity plus an
   ordinal or digest of the subtest name, with names kept only where the spec
   permits. Decide the naming rule in the spec first (REQ-844). **Done**
   (`b98e95e17`, TH-HARNESS-REQ-869): redacted names beside the unit logs,
   outside the execution index (see G17c and G17f results).
4. **G17d — Capacity-row margin** (measurement, then machine inputs if needed).
   After G17a, measure the row-limit test's apply duration against its 30 s
   budget. Keep the product default; it is what the test proves. If the margin
   stays under about a third, give that row a capacity profile with two CPU
   tokens (and the matching PostgreSQL claim the rule requires), so its server
   is not single-processor. **Measured; no change:** a second CPU could
   shorten only the 5 s CPU-bound first pass (see Results).
5. **G17e — Four lanes, third attempt** (measurement). Rerun the G2 criteria
   after G17a and G17b. **Rejected a third time** (see Results).
6. **G17f — Test I/O** (measurement first, then specification where needed).
   - Measure the shutdown checkpoint of retired single-unit generations from
     the container log before removal. If material, propose an immediate stop
     for generations whose databases all retire with them. The harness keeps
     graceful shutdown deliberately (`published_harness_retains_ordinary_graceful_shutdown`),
     so this needs an owner decision (A17).
   - Critical-path priority: heavy-unit containers and processes get a higher
     CPU and block-I/O weight than ordinary units. Measure first whether the
     WSL virtual disk honours I/O weights.
   - With G17c data, move tests that clone only to read or to roll back onto
     the transaction capability, row by row (successors of G3.4).
   - **Measured** (see G17c and G17f results): retired generations have no
     shutdown checkpoint (A17 withdrawn), I/O weights are unavailable here, and
     clones are spread too thinly for a few moves to matter. WAL compression
     (A18) was added instead.
7. **G17g — Test-side waits from the test deadline** (test support, tests,
   REQ-864 clarification). A zero-timeout `Poll`/`PollJob` takes the test's
   deadline less a 30 s settlement margin. The sub-minute job-poll hang guards
   drop their fixed timeouts. **Done** (`4a30fd00c`).

- **Implementation (2026-11-23).**
  - *G17a (not adopted; branch `g17a-attempt`, `2f23c6e44`).* Drafted A16 in
    TH-HARNESS-REQ-379 and the resource-profile list. The topology gained
    `parallel_toolchain` (4 CPU, 2 I/O, 1,536 MB, 16 processes), assigned to the
    Go analysis operations (`staticcheck`, `vet`, `govulncheck`, both `gosec`
    and both error-declaration operations), Biome lint, the Go and web builds,
    the Go toolchain installs, the two Go `go run` compatibility checks, and new
    policy entries for the two `go test -race` contract operations. The
    work-graph compiler set `GOMAXPROCS` and `RAYON_NUM_THREADS` to every
    operation's CPU claim. A compiler test asserted both variables equal the
    claim for every operation in `check`.
  - *G17b (committed `3373cd3b5`).* `runOperatorBinary` bounds each operator
    child by the test's remaining deadline minus 30 s (five minutes without a
    deadline). All 13 explicit 30 s and 60 s guards now use it. The
    interruption test's 30 s command context and 15 s progress waits use the
    same bound.
  - *Targeted checks.* `g17a-generate01`, `g17-lintgo01`, `g17-recovery01`
    (the four operator process rows), `g17-srcgen01`, `g17-srcschema01`,
    `g17-mdlint01b`, `g17-contract-exec01` (`harness-contract SUITE=execution`,
    which runs the compiler tests): all pass.
  - *Baseline for validation.* Every earlier run has 48–66% CPU-stall windows
    during the analysis block (`g6-check01`, `g6-check02`, `g17-probe-check01`,
    `g2b-lane4-check01`, `g2b-lane4-check02`).
- **Results (2026-11-23).** Four diagnostic `check` runs (cache off) with G17a
  and G17b applied, against `g6-check01` and `g6-check02` (pass, graph 887 s and
  889 s, heavy work 2,490 s and 2,537 s, peak CPU stall 65% and 66%):

  | Run | Lanes | Result | Graph | Heavy work | Peak CPU stall | Failed rows |
  | --- | ---: | --- | ---: | ---: | ---: | --- |
  | `g17a-check01` | 3 | Pass 365/365 | 918 s | 2,552 s | 47% | — |
  | `g17a-check02` | 3 | Fail 364/365 | 948 s | 2,631 s | 50% | `module.reference_data` refresh-admission matrix |
  | `g17a-lane4-check01` | 4 | Fail 364/365 | 966 s | 3,067 s | 49% | Network-flow row limit |
  | `g17a-lane4-check02` | 4 | Fail 363/365 | 921 s | 2,881 s | 51% | Network-flow row limit; `module.incidentbundles` ATT&CK round trip |

  - *G17a: both criteria failed; not adopted.* Each run still has 3–5 windows
    above 10% CPU stall, peaking at 47–51%. Three-lane graph wall rose to 918 s
    and 948 s. At four CPUs the analysis operations' own durations stayed
    within their earlier ranges (`gosec` 18–19 s against 18–26 s,
    `govulncheck` 12–15 s against 4–18 s, `staticcheck` 2–5 s against
    2–13 s). The 4-CPU build operations (`build_server_harness`,
    `build_report_renderer`) used to start in the run's first 110 s. They now
    started at 584–654 s on three lanes, because four free tokens appear only
    once the heavy phase thins out. Heavy work also rose, and windows above
    15% `io_full` went from 9–16 to 24–27 per three-lane run. PostgreSQL
    concurrency in the first 600 s was unchanged (11.9–15.9 units in each
    150 s window, both pairs), so the extra disk stall points to host disk
    state, as in G25. The attempt was taken out of the working branch.
  - *G17b: met.* The operator process rows passed in both four-lane runs; they
    failed in `g2-lane4-check02` and `g2b-lane4-check02`.
  - *G17d: recorded, no change.* Both 25 MB source passes share the 30 s
    participant budget: in the failing four-lane runs they stopped at 29.9 s
    and 30.0 s (5.4 + 24.5 s and 7.4 + 22.6 s, after 15.7 MB and 18.9 MB of the
    second pass). The two passes took 21.0–24.1 s on three lanes (four runs),
    a margin of 20–30%, and 28.3 s in the passing four-lane run. Even in a
    quiet window (`g17a-check01`, 1–3% `io_full`, CPU idle) they took 21.3 s.
    The test process runs at about 1.0 core for the 5 s first pass and about
    0.5 core for the second, which waits on PostgreSQL. In the failing runs that
    wait coincided with 13–27% `io_full`. A two-CPU profile could shorten only
    the first pass, by at most about 2 s, so the row is left unchanged. It runs
    the product default close to its budget on this host. That headroom is a
    product-owner question, not a harness one.
  - *G17e: four lanes rejected a third time.* Both runs failed, and graph wall
    (966 s and 921 s) showed no gain over three lanes. The row-limit test has
    now failed in 4 of 6 four-lane runs and in none of 14 three-lane runs.
  - *New failures.* The `reference_data` and `incidentbundles` failures are the
    first for those rows in 20 retained runs. Both came early in the run (188 s;
    79–109 s) during `io_full` windows of 15–25%. The `reference_data` server
    selected a fatal integrity outcome while a cancelled job unwound from a test
    barrier; the response does not carry the fatal reason code, so the cause
    (likely a 2 s default cancellation grace overrun under disk stall) is
    inferred. Neither test changed.
- **G17c and G17f results (2026-11-23).**
  - *G17c (committed `b98e95e17`).* TH-HARNESS-REQ-869 and
    `cartulary.harness_go_test_durations.v1`. Under basic diagnostics each Go
    unit writes `unit-logs/row-<row-id>/go-test-durations.json`, with every test
    and subtest's outcome and Go-reported elapsed time from the parse that
    settles row outcomes. Names are redacted like unit logs (URL credentials and
    `cursor_token` values were masked in 13 real names) and kept out of the
    secure execution index. The file holds at most 4,096 observations; overflow
    is counted (901 in the Network Flow alias-fixture unit, 14 in one graph
    projection unit). Targeted checks: `g17c-generate01`, `g17c-contract01`
    (Go row-runner assertions in `public_registry_parity`), `g17c-shapes01`,
    `g17c-srcgen01`, `g17c-srcschema01`, `g17c-lintscripts02`; all pass.
    `g17c-check01` wrote files for 312 of 312 Go units and the retained-secret
    scan passed (4,731 files). Its one failure (`incidentbundles`
    `network_flow_portability`, first in 24 runs) ran while a WAL probe of mine
    added disk load (run mean `io_full` 13.3%, the highest recorded). The
    participant-deadline expiry and the contract-compatibility rejection share
    one error mapping, so that reason code cannot tell them apart.
  - *Heavy-lane attribution* (`g17c-check01`, inflated by the probe; shares,
    not absolutes). Heavy-lane time sits in a few long single tests, not in
    setup or clones: `TestCurrentImportWorkflows_Integration` 210 s (one test,
    no subtests), `TestATTACKProductionPackRestoresPinnedHistory_Integration`
    181 s, the Network Flow incident transfer 153 s, the ATT&CK snapshot
    bundle round trip 100 s (68 s in the production-license subtest), the
    reference-catalog round trip 93 s (61 s in embedded-destination
    verification), and the recovery referential-integrity test 94 s (29 s and
    24 s in two subtests). Build and binary start add 5–25 s per heavy unit.
  - *Shutdown checkpoint: none; A17 withdrawn.* Suite termination
    (`terminateLeasedService` in `tools/testservices/main.go`) force-removes
    each generation container with its volumes. Docker events show SIGKILL
    (signal 9, exit 137), so PostgreSQL writes no shutdown checkpoint. Teardown
    took 111 s over 90 generations (median 0.93 s, max 4.8 s) in `g17a-check01`,
    and heavy-unit release 33–41 s per run: at most about 11 s of wall.
  - *I/O priority: unavailable on this host.* Every WSL virtual disk uses the
    `none` scheduler (BFQ not offered), and the cgroups have no `io.weight` or
    `io.cost`. `io.latency` exists, but Docker's container cgroups are
    root-owned. Not pursued.
  - *Clones.* 1,300 template clones per `check` across 579 tests, at most 39 in
    any one test (`TestOperatorReferencePackRetainedAdmission_Integration` 39,
    `TestServerReferencePackRetainedAdmission_Integration` 35,
    `TestCredentialStateAndBootstrapFlows_Integration` 35,
    `TestStateRuntimeAdmit_ServiceBacked_InitializationValidationRollback` 34,
    `TestRestoreRejectsBypassedReferentialIntegrity_Integration` 30). Creation
    took 384 s in total, spread across lanes. Moving the top tests to the
    transaction capability would remove about a fifth of the clone I/O through a
    module-by-module test refactor, so it is deferred to row owners.
  - *WAL compression (A18).* A seed-container probe measured the WAL of one
    `WAL_LOG` clone of the migrated template under each `wal_compression` mode
    (bytes by LSN difference, five clones each):

    | Mode | WAL per clone | Five clones | Update after checkpoint |
    | --- | ---: | ---: | ---: |
    | `off` (current) | 9,451 KiB | 1,235 ms | 77,972 KiB |
    | `pglz` | 3,458 KiB | 2,602 ms | 66,672 KiB |
    | `lz4` | 3,797 KiB | 923 ms | 67,079 KiB |
    | `zstd` | 2,680 KiB | 3,151 ms | 61,496 KiB |

    `lz4` cuts clone WAL by 60% at no time cost (about 7 GB of the 62–63 GB
    containers write per `check`); `zstd` saves more but is CPU-bound. A18 adds
    `wal_compression=lz4` to the harness server command and the REQ-804 clone
    paragraph. Full-page writes, `fsync` and synchronous commit are unchanged,
    and the seeded-generation test now asserts all four settings. The setting
    enters the provisioning identity, so a new seed image was built. Targeted
    checks: `a18-lintgo01`, `a18-slice-pg01`, `a18-slice-hb01`
    (`harness.browser.integration.postgres_template_preparation`); all pass.
    Validation: one diagnostic `check`, `a18-check01`. Per the owner's
    direction, one `check` confirms correctness and paired runs are kept for
    final validations; `a18-check02` was stopped before it started.

    | Run | Container writes | Mean `io_full` | Windows > 15% | Graph | Heavy work |
    | --- | ---: | ---: | ---: | ---: | ---: |
    | `g6-check01`, `g6-check02` | 61.8 GB, 63.0 GB | 8.1%, 9.6% | 9, 16 | 887 s, 889 s | 2,490 s, 2,537 s |
    | `g17a-check01`, `g17a-check02` | 62.3 GB, 62.8 GB | 9.7%, 10.4% | 24, 27 | 918 s, 948 s | 2,552 s, 2,631 s |
    | `a18-check01` | 48.6 GB | 10.6% | 26 | 961 s | 2,737 s |

    A18 cut container writes by 22%, about twice the WAL estimate. Disk stall
    and wall did not improve, which fits a virtual disk whose stall follows
    flush count rather than bytes (every commit still flushes WAL). Since G6,
    graph wall has drifted from 887–889 s to 918–961 s and heavy work by up to
    10% with no code on the path that explains it, as with G25's host-state
    drift. One run cannot separate A18 from that drift. The run failed one
    row: the `incidentbundles` ATT&CK snapshot round trip, whose import job did
    not finish within the test's fixed 30 s poll (`waitJob` in
    `routes_helpers_integration_test.go`). This was the second such failure
    after `g17a-lane4-check02`, again early in the run (82–112 s) at 13–21%
    `io_full` with the CPU idle. That wait is a test hang guard, not a product
    budget, like the G17b operator guards. A18 is kept for the write reduction;
    adoption awaits owner review.
- **G17g (committed `4a30fd00c`).** A zero-timeout `Poll` or `PollJob` takes
  the test's deadline less a 30 s settlement margin, never below one second,
  as REQ-864 allowed; the spec now says a hang guard uses the test's deadline
  and a fixed deadline is for a wait the test asserts. Six sub-minute
  hang guards drop their fixed timeouts: `incidentbundles` `waitJob` (30 s),
  `waitJobWithStatus` (10 s) and `waitFailedJob` (40 s), the Network Flow job
  wait (10 s), fixture publication (15 s) and the harness fault test (12 s).
  Polls with one minute or more, and product budgets the tests assert (the
  30 s participant budget, the 2 s cancellation grace), are unchanged.
  Targeted checks: `g17g-generate01`, `g17g-lintgo01`, `g17g-slice01`
  (`harness.browser.unit.httptestx_helper_contract` with the new
  `TestPollDefaultsToTheTestDeadline`), `g17g-srcgen01`, `g17g-mdlint01`; all
  pass.
- **Closed (2026-11-23).** G17a was not adopted. G17b, G17c, G17g and A18 are
  committed. A17 was withdrawn, and G17d and G17e are recorded with no change.
  The rest moves to two follow-up gaps: heavy-row test work (G27, owner-level
  test changes guided by G17c timings), and flush-bound disk stall plus the
  unattributed CPU bursts (G28, measurement first). Final validation: a
  paired diagnostic `check` on `4a30fd00c` (cache off), against the G6 pair:

  | Run | Result | Graph | Outer | Heavy work | Container writes | Mean `io_full` | Windows > 15% |
  | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
  | `g6-check01`, `g6-check02` | Pass, pass | 887 s, 889 s | 899 s, 900 s | 2,490 s, 2,537 s | 61.8 GB, 63.0 GB | 8.1%, 9.6% | 9, 16 |
  | `g17-final-check01` | Pass 365/365 | 857 s | 870 s | 2,383 s | 46.2 GB | 7.3% | 5 |
  | `g17-final-check02` | Pass 365/365 | 882 s | 893 s | 2,496 s | 48.4 GB | 8.7% | 21 |

  Both runs pass with no deadline failures, and graph wall is at or below the
  G6 pair (best yet: 857 s). Container writes are down 23–26% (A18). The
  mid-run analysis block still peaks at 67–70% CPU stall (G28). Duration files
  were written for 312 of 312 Go units in both runs; no leaked volumes.
- **Area.** Specification (A16; REQ-844 naming for G17c; A17 if adopted),
  machine inputs (operation resource profiles), implementation (work-graph
  compiler, Go runner), tests (process-test helpers).
- **Rationale.** The critical path is latency-bound heavy work that collides
  with unaccounted CPU bursts and shared disk I/O. Fixing the accounting
  model removes the bursts at their source. Per-subtest timings then show where
  the remaining heavy-lane time goes.
- **Long-term benefit.** Capacity means what the spec says it means; new
  operations cannot silently oversubscribe the host. Process tests stop
  failing on load alone. Test-cost work becomes measurable.
- **Compatibility.** Operation claims change packing and admission order, not
  results. Run-root schemas gain a bounded subtest artifact (G17c). A17 would
  change retirement behavior, so it waits for a decision.
- **Risk if unresolved.** Any change that shifts heavy rows into the analysis
  block reintroduces the two failures, even on three lanes. Four lanes stay
  out of reach, and test-cost work keeps relying on estimates.
- **Validation.**
  - G17a: in two diagnostic checks, no 10 s window exceeds 10% CPU stall; the
    analysis operations' own durations are recorded; wall is unchanged or
    better.
  - G17b: the operator rows pass under a four-lane run.
  - G17c: a diagnostic `check` retains subtest timings for every Go unit
    within the bound, and the secret scan and redaction tests pass.
  - G17d: the apply duration and budget are recorded.
  - G17e: the G2 criteria.
  - G17f: measured shutdown-checkpoint bytes and any priority effect, recorded
    here with run IDs.

### G27 — Heavy-row test work

Opened 2026-11-23 as a G17 follow-up. Heavy lanes are saturated for most of a
`check`, so cutting heavy work saves about a third of it in wall time.

**Evidence.** G17c timings from `g17-final-check01` and `g17-final-check02`:
145 heavy top-level tests, 2,259 s on average. Three tests take 28%:
`TestNetworkFlowIncidentTransferThroughPublicOperations_Integration` 247 s,
`TestCurrentImportWorkflows_Integration` 215 s and
`TestATTACKProductionPackRestoresPinnedHistory_Integration` 166 s. Because
code-reading estimates failed in W2-2, each was profiled before any change.
Temporary phase marks plus a forced failure in cleanup kept the Go JSON
stream, then both were removed (`g27-prof-cw01`, `g27-prof-nf01`,
`g27-prof-attack01`; isolated runs, so absolute times are lower than under
`check`). The harness keeps only the last 1 MiB of a failing unit's output, so
the workflow profile covers 12 of 22 subtests.

| Test | Isolated | Largest measured phases |
| --- | ---: | --- |
| Current import workflows | 191 s | Waiting for the recovered job after restart, 6.0 s per subtest (about 130 s in all); commit-fault recovery 5.5 s per fault case; server starts 0.45–0.65 s each |
| Network Flow incident transfer | 173 s | Investigation, indicator, report release and edits 44 s; operator backup and restore 37 s; transfer import 27 s; core corpus import 14 s; two re-exports 5–6 s each; two paged table comparisons 6–8 s each |
| ATT&CK production restore | 109 s | Three fault probes at 17–18 s each, each a full restore of the captured backup up to its reference-pack fault; final restore about 17 s; pack import 12 s; backup capture 7 s |

**Cause in the workflow test.** A restarted server activates its job runner
during startup publication, before admission opens. The activation scan skips
every candidate (`candidateReady` requires open admission), so a recovered job
waits for the next scan at the production 5 s cadence. That is product ordering
under Core 01 and is not changed here.

**Change (committed `7a62a435a`).** Core 01 makes the production Jobs policy
immutable but lets "harness-owned composition" inject shorter values.
`server.Options.JobPolicy` (nil selects `jobs.ProductionRuntimePolicy()`,
through `jobRuntimePolicy`, with `TestJobRuntimePolicySelection_Unit`) is that
seam; deployment configuration has no path to it. The current-workflow tests
inject a 100 ms recovery scan and keep production retry delays and leases. The
commit-fault scenarios rely on those, stopping after one failure before the
server closes. In isolation `current_workflows` went from 169–208 s to 71 s
and `current_cancellation` from 14–17 s to 4.6 s (`g27-slice-imports01`).
Targeted checks: `g27-generate01`, `g27-lintgo02`, `g27-srcgen01`,
`g27-slice-server01`, `g27-slice-imports01`; all pass. Confirming `check`
`g27-check01` on `7a62a435a`: **pass 365/365**. In it the workflow test took
116 s, against 208–221 s in the final G17 pair (−99 s of heavy work). Graph
wall was 945 s and heavy work 2,671 s, because nearly every other heavy test
ran slower on host disk state: mean `io_full` 11.5% against 7.3–8.7%, and up
to +45 s per test, as in G25. One run cannot show the wall gain under that
noise; the per-test saving is direct.

**Remaining, owner-level.** The rest of the measured time is product work
that the tests' assertions name:
- *ATT&CK restore fault probes.* Each of `missing_member`, `damaged_member` and
  `wrong_authority` runs a full restore into its own fresh target (about 53 s
  together). Options: run the independent probes concurrently, or assert two
  member faults in one restore. Both are recovery-owner decisions; clone-based
  targets were rejected in G3.7.
- *Network Flow incident transfer.* Backup/restore, the transfer import and the
  report release are the subject of the test. The remaining options are owner
  judgments about scope, such as a smaller corpus or one destination instead
  of two.
- *Setup-only ATT&CK imports.* The report composition preview imports the full
  95 MB pack (91 s test) only as setup. A smaller authentic-shaped dataset or a
  snapshot would need the reporting owner (G3.4).
- *Other restarts with recovered jobs* (incident-bundle queued export recovery,
  a few imports and reporting tests) can adopt the same `JobPolicy` seam; each
  saves about 5 s.

### G18 — ATT&CK release fixture expiry

- **Evidence.** `container.pack` under `fixtures/reference-packs/dataset-v1/attack/release/envelopes/83ec1d97…/` carries `metadata/snapshot.json`, `targets.json` and `timestamp.json` with `"expires":"2026-12-08T20:40:00Z"`. Envelope validation requires `verified_at` before `trust_valid_until` ([envelope.go:52](../../internal/modules/reference_data/internal/packformat/envelope.go)), and the recovery test imports with `CLOCK_TRUSTED=true` against the real clock. The failure after that date is inferred, not yet reproduced.
- **Remediation.** Tracked as a separate task: confirm with a pinned clock, then refresh the fixture through the reference-pack Make targets or pin the tests' verification clock, preferring the fix that does not recur.
- **Risk if unresolved.** `module.recovery` and `module.incidentbundles` ATT&CK rows fail from 2026-12-08.

### G19 — Drop checkpoints exceed the cleanup budget under load

- **Evidence.** `g3-check01` failed `module.incidentbundles.integration.incident_bundle_export_and_import_…`: a `forced_drop` in `TestIncidentReferenceCatalogSurvivesUnavailablePackRoundTrip_Integration` hit its 35 s operation budget while PostgreSQL waited on `CheckpointDone`, with the checkpointer in `DataFileSync` (`sync_ms` 70,465 for 19,425 buffers). The cleanup failure quarantined the unit's generation, so the remaining subtests failed setup with `fixture_error`. The same failure appeared on 2026-11-20 in `tcr-s1130-check02` (`attack_production_recovery`) and `tcr-s1130-check03` (`reset_preserves_crypto_format_identity`), before any change in this programme, so it is pre-existing. G5 roughly doubled concurrent PostgreSQL-backed work, which plausibly raises its frequency; one run cannot establish a rate. Drop latency in `g5-check01` already reached 24.3 s.
- **Reproduction (2026-11-22).** After the G3 fixes the failure reproduced in 2 of 2 diagnostic `check` runs (`g3-check01`, `g3-check02`), each time in `TestIncidentReferenceCatalogSurvivesUnavailablePackRoundTrip_Integration` (`sync_ms` 70,465 and 52,294). Fix 1's estimate change (workspace row 64 s → 17 s) re-packed the reference-data heavy cohorts, so the incident-bundles unit now overlaps the Network Flow heavy cohort for 260–290 s instead of 60 s (`g5-check01`, which passed). Restoring the old estimate would only move the collision; the work should not be committed until G19 has a structural fix.
- **Remediation options.** (a) Retire dedicated databases inside a single-unit exclusive generation with the generation instead of eagerly (spec change to REQ-357's eager retirement), which removes the forced checkpoints entirely; (b) bound concurrent PostgreSQL I/O explicitly, for example a lower `postgres_lanes` default or a dedicated cluster capacity, restoring the headroom the old process-slot arithmetic provided by accident. The NLSpec forbids longer cleanup budgets and weaker durability, so neither of those is an option.
- **Area.** Specification, harness capacity policy.
- **Implemented (option a, committed `500e3bdda`).** REQ-357 now retires dedicated and migration databases with a single-unit exclusive generation (allocation `postgres:unit:<unit>`) instead of eagerly; REQ-804's matching sentence is updated. Shared, browser-lane, measurement, owned and persistent clusters keep eager drops, and explicit drops remain available. Code: `retiresWithGeneration` guards `registerDatabaseCleanupT` in [pgtest.go](../../internal/testutil/pgtest/pgtest.go), keyed on the harness's admitted binding; `RetireClonedDatabaseT` applies the same rule to databases a test clones itself (used by `TestTimelineTagLabelRepair_Integration`). `TestDatabaseFixtureCleanup` gains `unit_generation` (no drop) and `lane_generation` (eager drop) modes; disabling the rule fails exactly the three `unit_generation` cases. A real attachment run (`g19-platform01`) created two databases, issued no drop attempts and terminated the cluster once.
- **Effect.** `g19-check01` issued 2 drops instead of about 1,325. Its one failure was the timeline-repair test's own backup cleanup drop (1,693 buffers, still stalled in `DataFileSync`), which is now routed through `RetireClonedDatabaseT`. Drop latency shows why G5 mattered: median 0.11 s and p99 about 4 s before G5, median 0.25 s and p99 6–8 s after, with drops over 10 s rising from 1–5 to 10–12 per run.
- **Disk saturation is the underlying limit.** `g19-check02` failed differently: an ATT&CK bundle job in `TestATTACKSnapshotReferencesSurviveBundleRoundTrip_Integration` was still running after the test's 30 s wait. PostgreSQL checkpoint diagnostics show total checkpoint sync time per `check` of 611–720 s before G5 and 1,270–1,480 s after. Generation retirement cut checkpoints from 1,464 to 92 but not the total sync (1,212–1,450 s), because the data still reaches the disk. G5 doubled concurrent PostgreSQL work without shortening the wall (heavy lanes bound it), so the extra concurrency only saturated the disk. Option (b) is therefore needed as well: an explicit bound on concurrent PostgreSQL work, replacing the accidental process-slot cap, while CPU-only units keep the G4a/G5 gains. Experiment: `postgres_lanes` 8 through the override file (`g19-pg8-check01`).
- **Postgres cap experiment (`g19-pg8-check01`, `postgres_lanes` 8 by override).** Total checkpoint sync fell to 569 s (below the pre-G5 611–720 s; max single checkpoint 18.6 s) and the graph wall was 942 s, the best so far; CPU-only units still ran 15–18 wide early. Two units still failed: (1) the deliberate source drop in `TestCanonicalOperatorRestoreBundle_Process` waited past 35 s on a checkpoint already in progress (`requested=2 done=1`); (2) `TestHostRegistryProfileConsumer_Integration` exceeded a short job wait. Finding (1) is a side effect of generation retirement: without frequent drop checkpoints, each remaining checkpoint carries a larger backlog (median 4.6 s capped, 12–17 s uncapped, against well under 0.1 s before), so the few remaining deliberate drops are more exposed than before.
- **Hardening (committed `500e3bdda`).** Scenario removals now rename instead of drop inside a single-unit generation: `pgtest.Harness.RemoveDatabase` ends the database's sessions and renames it to a fixed-length `rm_<uuid>` name, forcing no checkpoint, and drops elsewhere; REQ-357 documents this. `TestCanonicalOperatorRestoreBundle_Process` uses it for its source-loss step (`g19b-restore01`: 5 creates, 0 drop attempts). The reference-data `requireJob` bound rises from 5 s to 1 minute, like the imports waits. G5b is reverted (see G5).
- **Remaining exposure.** Deliberate drops that are part of a scenario or of drop diagnostics stay eager: the recovery operator's portable-restore source drop, the drop-diagnostic tests, the explicit-only reset characterization, and browser-lane and measurement fixtures. If they surface, option (b) caps concurrent PostgreSQL I/O.
- **Risk if unresolved.** Intermittent `check` failures under load, with no product defect behind them.
- **Validation.** Several consecutive diagnostic `check` runs without a cleanup `deadline_exceeded`, and drop latency p99 well under the budget.

### G20 — Static checks scan nested worktrees

- **Evidence.** In `g19b-check02`, `lint_biome` (nested root `biome.json`), `lint_shell` and `json_shape_check` (`executable input path is not normalized`) and `otel_conformance` (telemetry import boundary in the nested copy) all failed on files under `.claude/worktrees/zealous-vaughan-849873/`, a git worktree another session created inside the checkout at 17:50:40, seven minutes into the run. Every service-backed unit passed. While that worktree exists, `make lint-markdown` fails the same way (nested root `biome.json`), and the planning source snapshot fails too, so every graph target would fail before starting.
- **Cause.** `lint_shell`, `json_shape_check` and the source snapshot already enumerate inputs through Git (tracked plus non-ignored untracked files, as REQ-153 requires). Git lists an untracked nested repository as one `dir/` entry, which path admission rejected with an unhelpful message. `otel_conformance` walked the filesystem itself, and Biome discovers configuration files itself.
- **Remediation (done, committed `500e3bdda`).** REQ-153 now states that enumeration fails naming a non-ignored nested repository, that whole-checkout scans use this enumeration, and that third-party tools honour Git ignore rules (A10). One Git enumeration in `restricted-input-boundary.mjs` serves the source snapshot, the executable-input policy and the static-analysis file lists. `check-otel-conformance.mjs` takes its repository index from it, keeping its build-output walk. `.gitignore` ignores `/.claude/worktrees/`. Restricting checks to tracked files was rejected: it contradicts REQ-153 and would leave new files unlinted and outside cache keys.
- **Area.** Specification (REQ-153); implementation (source enumeration, OTel conformance); repository ignore rules.
- **Risk if unresolved.** Any agent worktree created under the checkout turns `make check` red without a defect.
- **Validation.** With the worktree still present: `make lint-markdown`, `make lint CHECK=shell`, `make lint CHECK=frontend`, `make source-check CHECK=schema` and `make otel-conformance` pass. The source-snapshot contract test covers the named failure and the ignored case. The Git index matches the old walk's 3,862 Go files exactly, without the nested copy.

### G21 — Cleanup layout misses declared outputs

- **Evidence.** Audit on 2026-11-22. The task surface declares root `dist` (from `build`, `package-release`, `package-inspect` and `credential-capacity-assessment`) and `.cartulary/reference-packs/builders` and `modules` (from the reference-pack producer and module-check commands) as build outputs. None is registered in `tools/workspace_layout.json`, so both cleanup commands preserved them; builders alone held 1.4 GB. `apps/web/.cartulary` holds Vitest group scratch from `test-vitest-diagnostics.mjs`, but only its legacy `test-results` child was registered. A stale Vite cache from an older fixture root sat in `apps/web/src/testing/fixtures/node_modules`. The `.cache/cartulary` children `graph-v2` (native graph cache, 1 GB) and `static-analysis` were covered by their parent but unregistered.
- **Remediation (done, uncommitted).** Registered `dist`, `.cartulary/reference-packs/builders` and `apps/web/.cartulary` (with the legacy results nested beneath it) at the `clean` tier; `.cartulary/reference-packs/modules` at `distclean` with the other dependency installs; the fixture Vite cache as legacy `clean`; and both cache children under `cache`. The REQ-505 disposition table records the new classes, and the cleanup contract test covers each path plus a preserved sibling (`.cartulary/reference-packs/isolation`, which holds qualification evidence).
- **Left preserved by design.** External caches (`~/.cache/cartulary`: Go build, module and temporary work, harness scratch), `temp`, `.cartulary/retained-work`, runtime roots, browser leases, and ad-hoc custom roots under `.cartulary` (about 4 GB of manual qualification and remediation output, including the other `.cartulary/reference-packs/*` folders).
- **Validation.** `make harness-contract SUITE=regression` (cleanup contract) and a `make distclean` preview list exactly the registered paths.

### G22 — Readiness stamps vouch for removed images

- **Evidence.** `harness-contract SUITE=regression` (`clean-regression01/02`) failed `native_graph_observations` and `native_fixture_lifecycle` at the renderer `image` stage: the pinned Playwright renderer image was no longer in the local Docker store (only the SeaweedFS and PostgreSQL images remained), while the `playwright-install` readiness stamp still recorded a successful install, so the pull never re-ran. A run with the G10 scoping disabled (`g10-ab-off01`) failed identically.
- **Remediation (done, uncommitted).** Same rule the spec already applied to test-service images, now stated for the browser installation: a hit must prove the pinned renderer image is present with its repository digest and platform (spec text beside the test-service image paragraph). `installed-readiness.mjs` gains `rendererImageReadiness`, built on the lease's own `verifyRendererImage`, used by the installed-only browser probe and a `renderer-image` CLI mode; it reports under the existing `chromium` subject with `playwright_install` recovery, so the closed failure vocabulary is unchanged. `playwright-install.sh --invalidate-stale-stamp`, run by the stamp recipe before the readiness cache, drops the stamp when that proof fails.
- **Validation.** New tests cover present, missing and replaced images for both the probe and the stamp invalidation (`ui_review_contract`, `ui_review_lifecycle` rows pass). End to end: with the image removed and the stamp still on disk, `make bootstrap PART=browser` reported the missing image, reinstalled and restored it; `native_graph_observations` and `native_fixture_lifecycle` then passed (`g22-native01`).

### G23 — Pinned security package removed upstream

- **Evidence.** After `make distclean`, `make bootstrap` failed at `playwright-install`: `libnss3-tools_3.98-1ubuntu0.2_amd64.deb` returns 404 from `security.ubuntu.com`, which drops superseded security builds. The cached copy under `.cache/cartulary/playwright/trust-tools` had hidden this. `post-distclean-check01` then failed `playwright_install` and skipped 10 dependent units.
- **Fix (uncommitted).** `browser_trust_tool` in `tools/toolchain_pins.json` moves to `2:3.98-1ubuntu0.3` (615,126 bytes, SHA-256 `0bb79191…88c0`). Provenance: the `noble-security` `InRelease` file verifies with the Ubuntu archive key, lists the `main/binary-amd64/Packages.gz` digest, and that index lists this package's size and digest; the download matches.
- **Remaining risk.** The next security update will break the pin again. A durable fix would pin to an archive that keeps superseded builds, or retain the verified package in a release-evidence cache that cleanup preserves.

### G24 — Heavy lanes idle while waiting for PostgreSQL capacity

- **Evidence (`pack-base-check01`, complete admission log).** Heavy units spent 443 s in host admission while holding a heavy lane, plus 262 s acquiring fixtures. Of the admission wait, 432 s was a `postgres` deficit: ordinary PostgreSQL Go units (each one active and one resident lane) held all 16 lanes. Six heavy units waited 40–132 s each, mostly in the first 700 s; one sat seventh in the queue. About 410 s of heavy-lane time, roughly 135 s of wall at three lanes, went unused. The graph grants the logical heavy lane first, and the heavy unit then joins the cross-process queue behind ordinary PostgreSQL waiters that arrived earlier (oldest-waiter order, REQ-803).
- **Options.** (a) Reserve PostgreSQL capacity for the heavy lanes: ordinary PostgreSQL work may use at most `postgres_lanes − 2 × postgres_heavy_lanes` lanes, so a heavy unit always finds its active and resident lanes. (b) Order host admission by graph rank, so critical-path requests precede ordinary ones. (a) is a capacity statement in the spirit of G5; (b) changes REQ-803's ordering. The `postgres_lanes` 8 experiment (`g19-pg8-check01`, best graph wall 942 s) suggests ordinary PostgreSQL work tolerates fewer lanes.
- **Validation.** Heavy-unit host-admission wait near zero; heavy-lane idle time before the tail near zero; ordinary PostgreSQL units still finish before the heavy tail.
- **Implemented (option b, uncommitted).** REQ-803 amended: waiter priority stays arrival order, except that between shared requests one that claims a `postgres_heavy` lane with no deficit on that lane precedes requests claiming no heavy lane. The existing deficit-non-increase rule then keeps ordinary work from drawing down the PostgreSQL lanes the heavy request still lacks, which reserves capacity only while a heavy unit needs it; this avoids option (a)'s static reservation, which would starve `ci` browser lanes. A heavy request still waiting for its lane gains no precedence, and exclusive, drain and observation requests keep arrival order, so the exclusive barrier is unchanged. Code: `precedes`/`heavyReady` in `mutate` ([admission_linux.go](../../internal/testutil/harnessadmission/admission_linux.go)); subtests `available_heavy_lane_precedes_older_ordinary_postgres_waiters` and `heavy_request_without_its_lane_keeps_arrival_order`.
- **Result (`g24-check01`).** Pass 365/365, graph 1,022 s against 1,152 s and 1,171 s in the two preceding runs. Heavy-unit host admission 443–454 s → 35 s; heavy work 3,263–3,375 s → 2,921 s; lanes all busy 946 s, makespan 42 s above a perfect split. Ordinary PostgreSQL units: admission wait unchanged (6,184–6,837 → 6,761 unit-seconds) and their last unit still finishes before the heavy tail (960 s against 1,015 s). CPU-only admission wait stays about 1 s. Checkpoint sync 871 s (was 1,000–1,066 s).

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

Phase-1 workstreams. For remaining `check` wall-clock work, [Section 12](#12-phase-2-plan--check-wall-clock) supersedes WS2, WS3, WS4 and WS7.

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
- **Status.** Complete on 2026-11-22. `3bd5875af` fixed the format drift; `ws0-baseline-check01` passed all 364 units on commit `1cd7d19e7` and is the reference run for later workstreams.

### WS1 — Measurement and decision inputs

- **Scope.** G13 measurement, G14 audit, G3 and G6 profiling.
- **Dependencies.** WS0.
- **Steps.** One `ci` and one `release-check` with diagnostics on an idle host (schedule overnight). Profile the seven heavy rows. Add temporary local instrumentation for generation startup phases, or wait for WS4's G9 to lift truncation. Run the estimate audit.
- **Risks.** `release-check` may run for many hours. Bound it by measuring its measurement projection separately if the owner prefers.
- **Exit criteria.** Measured walls and critical paths for `ci` and `release-check`; per-row cost breakdowns for the seven heavy rows; generation startup breakdown; owner decisions D1–D5 recorded in [Section 10](#10-open-decisions-and-handoff).
- **Status.** In progress. `ci` measured on 2026-11-22 (`ws1-ci01`); see Section 2. Its admission log truncated at 669 s, so attributing the mid-run admission stall needs G9 (or a temporarily raised record cap) before WS2 design is final. Remaining: `release-check`, heavy-row profiles, generation startup breakdown, estimate audit.

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
| A1 | §10.2 `process_slots` row; REQ-803 lane accounting | Process slots from the OS process limit with margin; active and resident `postgres` claims stated as distinct (the single-charge change was reverted) (adopted, `500e3bdda`) | G5 | Registry v9 (v8 retired) |
| A2 | §2.1 REQ-803 | Deficit-non-increase backfill at host admission, with descendants of a backfilled branch bound by the bypassed waiter (adopted, `500e3bdda`). Graph-local footprint admission (G4b) deferred | G4 | None |
| A3 | REQ-804, REQ-811 | Rewritten in Section 12.6: harness-assigned generation identity and seed-image provisioning (previous text: snapshot-seeded provisioning identity; single-use generation claims released at terminal settlement) | G6 | Attachment identity |
| A4a | REQ-857 | Admission decisions coalesce on a defined decision identity; zero quantities omitted (adopted with G9 stage 1, `500e3bdda`) | G9 | None |
| A4b | REQ-371 | Withdrawn: replay shows segment churn comes from waiting-unit volume (G4), not holder changes | G9 | None |
| A9 | REQ-357; REQ-804 | Databases in a single-unit exclusive generation retire with the generation instead of eager `DROP DATABASE` (adopted, `500e3bdda`); the `postgres-db-retired` evidence event is drafted, uncommitted | G19 | None |
| A10 | REQ-153 | Non-ignored nested repositories fail enumeration by name; whole-checkout scans use the Git enumeration; third-party tools honour ignore rules (adopted, `500e3bdda`) | G20 | None |
| A11 | REQ-505 | Disposition table adds root `dist`, reference-pack builders and module staging, and the whole `apps/web/.cartulary` (drafted, uncommitted) | G21 | None |
| A12 | Readiness cache profiles (beside the test-service image rule) | A browser-installation hit must prove the pinned renderer image is present by digest and platform; otherwise the stamp is invalidated (drafted, uncommitted) | G22 | None |
| A13 | REQ-803 | Between shared requests, one claiming an available `postgres_heavy` lane precedes requests claiming none; otherwise arrival order (drafted, uncommitted) | G24 | None |
| A14 | Lifecycle estimates beside REQ-865 | Withdrawn (Section 12, G26 decision) | G26 | None |
| A15 | §10.2 capacity table | Not adopted: four lanes gained no wall time and added contention failures (Section 12, G2 result) | G2 | None |
| A16 | REQ-379; §10.2 `cpu_tokens` and `process_slots` rows | Not adopted: bounding operation children to their CPU claim left the CPU-stall bursts in place and moved 4-CPU builds to the run's tail (G17 Results); the draft is on local branch `g17a-attempt` | G17 | None |
| A17 (conditional) | REQ-357; REQ-804 retirement | Withdrawn: suite termination already force-removes generation containers, so there is no shutdown checkpoint to skip (G17f) | G17 | None |
| A18 | REQ-804 clone paragraph | Harness clusters run with `wal_compression=lz4`, cutting `WAL_LOG` clone WAL by about 60% and container writes per `check` by 22% (`a18-check01`); durability settings unchanged; part of the provisioning identity. Disk stall and wall unchanged in one run (committed `81f709c2d`; adoption awaits owner review) | G17 | None (new seed image) |
| A19 | REQ-869 (new) | Go test and subtest durations retained under basic diagnostics beside the unit logs, redacted and bounded, outside the execution index (committed `b98e95e17`; keeping redacted names rather than digests awaits owner review) | G17 | Schema `cartulary.harness_go_test_durations.v1` |
| A8 | REQ-803; REQ-381 | Admission waiters re-evaluate on admission-state change, with a bounded fallback for dead-owner reclamation | G16 | None expected |
| A5 | REQ-407 | A passing fixture comparison keeps its digest record only (Network Flow summary with `materialized_input_root` null, IR `comparison.json`, reference-pack `bootstrap.json`); a failing one keeps its full material (drafted, uncommitted) | G8 | None |
| A6 | §5 Go path table; §14 readiness | Invocation-scoped `GOTMPDIR`/`GO_TMP_DIR` beneath the machine path, removed at settlement, dead-owner reclamation; `TMPDIR` stays host-wide (drafted, uncommitted). Retention and cache-maintenance commands (G11, G12) not started | G10 | None |
| A7 | Aggregate membership; Core 05 alignment | Quiet measurement leaves `ci`/`release-check`; release verification requires measurement roots | G13 | Aggregate projections |

Each amendment lands spec-first in one change with its machine projections, regenerated outputs (`make generate`) and the tests that exercise it, following [tools/AGENTS.md](../../tools/AGENTS.md).

## 9. Validation Protocol

1. **Host state.** No other harness invocation, no running harness containers, normal background load.
2. **Command.** `CARTULARY_TEST_RUN_ID=<fresh-id> CARTULARY_OUTPUT_MODE=ci make check HARNESS_DIAGNOSTICS=basic CARTULARY_HARNESS_CACHE_MODE=off`, with an override file only when the treatment is a capacity change.
3. **Repetitions.** Two per treatment, sequential, same commit apart from the change under test.
4. **Comparison.** `make explain-run RESULTS_DIR=<run> DETAIL=performance JSON=1 COMPARE_RESULTS_DIR=<reference> COMPARISON=equivalent`, plus the metrics in Section 3. The comparison is descriptive, never a gate (REQ-865).
5. **Correctness.** Each change runs its targeted rows per [AGENTS.md](../../AGENTS.md). Spec and machine-input changes also run `make source-check CHECK=generated` and, for documents, `make lint-markdown`. A diagnostic `make check` is the default broad verification and measurement vehicle. Owner direction (2026-11-22): keep `ci` and `release-check` runs to a minimum because of their run time, and use them only when a change affects behavior only they exercise (browser lanes, quiet measurement, release qualification), stating that reason first. If one must be stopped, settle it with `make test-run-recovery-status` and `make test-run-recover RESULTS_DIR=<exact run>`; stopping the task alone leaves harness containers running.
6. **Recording.** Append run IDs, walls and deltas to Section 11 of this tracker.

## 10. Open Decisions and Handoff

| ID | Decision | Needed by | Recommendation |
| --- | --- | --- | --- |
| D1 | Approve amendments A1–A7 individually | Each workstream's start | Approve A4 and A2 first; they unblock diagnostics and the largest structural gain. |
| D2 | Retention budget defaults and what counts as pinned | WS5 | Age 14 days, 30 GB, keep the 5 newest per target; pin anything named in evidence-roots files. |
| D3 | Whether quiet measurement leaves `ci`/`release-check` | WS6 | Defer until `release-check` is measured; in `ci` measurement costs only 96 s plus a 285 s drain. Needs Core 05 owner alignment either way. |
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
| 2026-11-22 | WS0 / G1 | `make format` removed one blank line in `prepare.go` | `ws0-format01`, `ws0-lintgo01` (`lint CHECK=go`) | Pass; committed `3bd5875af` |
| 2026-11-22 | WS0 | Reference `check`, diagnostics on, cache off, default 3 lanes | `ws0-baseline-check01` | Pass, 364/364. Graph 1,007 s; outer 1,134 s; post-graph scan 120.5 s; heavy work 2,839 s; 188 CPU-only units waited more than 300 s; 109,012 wait pairs; run root 1.1 GB; admission denials `postgres` 1,243 and `process` 1,193 (sampled, 186,482 records omitted) |
| 2026-11-22 | WS1 / G13 | `ci`, diagnostics on, cache off, default 3 lanes, 4 browser lanes | `ws1-ci01` | 505/506; `frontend_fallow_static` failed (G15). Graph 3,279 s; outer 3,464 s; post-graph about 181 s; average concurrency 3.9; CPU tokens 32% used; quiet measurement 96 s; heavy work 4,975 s; fixture time 2,817 s; run root 1.4 GB; events 305 MB; admission records 2,269,576 dropped |
| 2026-11-22 | G9 stage 1 | Decision-identity coalescing and zero-quantity omission (REQ-857 amended) | `g9-native-runtime01` (row `harness.command_surface.unit.native_runtime`), `g9-mutation01`, `g9-lintgo01`, `g9-mdlint01` | Row passed; with the old exact-match rule restored the new subtest fails ("identity-preserving retries produced 3 records"); uncommitted |
| 2026-11-22 | G9 stage 1 | Diagnostic `ci` stopped at owner direction after 16 min; replaced by a diagnostic `check` | `g9-ci01` (stopped), `20261122T173425Z-p56152` (recovery status), `20261122T173437Z-p56894` (recovery) | Stop left 9 harness containers; exact-run recovery removed all 9 |
| 2026-11-22 | G9 stage 1 | Diagnostic `check` with coalescing | `g9-check01`; `harness-observability-check CHECK=all`; `explain-run DETAIL=critical-path VIEW=host-admission` | Pass 364/364, graph 976 s. Host admission complete, 0 omitted, 2,366 records, 5.8 MB; readers `full_validity=1`. Denials: priority-only 361 records / 2,723 s (CPU-only 123 / 1,482 s); process-deficit denials show residents holding a median 66 of 76 slots |
| 2026-11-22 | G4a | Deficit-non-increase backfill (REQ-803 amended) | `g4-admission03` (row `harness.evidence_accounting.behavior.observation_admission`), `g4-mutation01`, `g4-mutation02`, `g4-runtime01`, `g4-browserfixtures01` | All pass; mutations caught (strict priority restored fails 4 subtests; descendant rule disabled fails its test) |
| 2026-11-22 | G4a | Diagnostic `check` | `g4a-check01` | Pass 364/364, graph 990 s. Priority-only denials 2,723 → 1,364 s observed (CPU-only 1,482 → 804 s); `process` now the dominant deficit, residents holding a median 58 of 76 slots; `postgres` denials show residents 7 + holders 8 of 16 lanes |
| 2026-11-22 | G5 | Process slots from OS limit (registry v9); single lane for exclusive borrowers | `g5-cmdsurface01`, `g5-browser02`, `g5-catalog01`, `g5-boundary01`, `g5-source-generated`, `g5-source-schema`, `g5-mutation01`, `g5-generate01` | All pass; one browser-lane assertion that encoded the old 76-slot cap now checks the PostgreSQL bound; mutation caught |
| 2026-11-22 | G4a + G5 | Diagnostic `check` | `g5-check01` | Pass 364/364, graph 984 s, outer about 1,109 s. Host-admission waiting 5,485 → 705 unit-seconds; admission transaction time 236 → 35 s; 15–18 units run for the first ~450 s (CPU tokens 17–18 of 19); CPU-only units waiting more than 300 s 188 → 26, last done at 616 s. Wall unchanged: after ~600 s only the 3 heavy lanes run (2,766 s ÷ 3). Wait pairs unchanged at 109,042 (175 MB), disproving the G9 stage 2 assumption. Fixture acquisition 998 → 1,921 s summed with more concurrent container starts |
| 2026-11-22 | G3 | Heavy-test profile: cross-run row durations, test-services journals, four read-only code investigations | `eff-base-check01`, `ws0-baseline-check01`, `g9-check01`, `g4a-check01`, `g5-check01`; `g3-verbose-probe01` | Heavy work 2,766 s: ~19% infrastructure (start, migration, clones, drops), ~81% test work. Ten ranked fixes recorded in the G3 profile; G6 measured; G17 and G18 opened |
| 2026-11-22 | G3 fixes 1–3 | Timeout split, single restart for replays, shared deployment for approved mapping | `g3-approved01`, `g3-approved-probe01`, `g3-currentworkflows01`, `g3-currentworkflows-base01`, `g3-workspace01`, `g3-source-catalog`, `g3-source-generated`, `g3-lintgo01` | All targeted rows pass; durations in the G3 fixes table |
| 2026-11-22 | G3 | Diagnostic `check` with fixes 1–3 | `g3-check01` | Fail 364/365: harness `fixture_error` from a 35 s forced-drop budget (G19, pre-existing mode). Graph 959 s; heavy work 2,386 s |
| 2026-11-22 | G3 fix 2b | Contention robustness for current workflows | `g3-currentworkflows02`, `g3-lintgo02` | Pass; `current_workflows` 175.0 s, `current_cancellation` 14.4 s |
| 2026-11-22 | G3 fixes 1–3, 2b | Diagnostic `check` | `g3-check02` | Fail 364/365, G19 again in the same unit (`sync_ms` 52,294). All five changed rows pass under contention: `current_workflows` 229.6 s, `current_cancellation` 16.9 s, `approved_mapping` 8.4 s, workspace 5.3 s, timeout 64.0 s. Graph 931 s; heavy work 2,326 s |
| 2026-11-22 | G19 | Generation retirement for unit-generation databases (REQ-357/804 amended) | `g19-contract01`, `g19-pgtest01`, `g19-platform01`, `g19-mutation01`, `g19-lintgo01`, `g19-mdlint01` | Pass; mutation caught; real attachment: 2 creates, 0 drop attempts, 1 cluster termination |
| 2026-11-22 | G19 | Diagnostic `check` | `g19-check01` | Fail 364/365: `TestTimelineTagLabelRepair_Integration`'s own backup drop timed out (test-owned cleanup). Drops 1,325 → 2. Graph 990 s |
| 2026-11-22 | G19 | Route test-cloned databases through the same rule (`RetireClonedDatabaseT`) | `g19-migrations01`, `g19-lintgo03` | Pass |
| 2026-11-22 | G19 | Diagnostic `check` | `g19-check02` | Fail 364/365: ATT&CK bundle job exceeded the test's 30 s wait under I/O load. Drops 2 (max 17.2 s). Checkpoint sync 1,450 s total. Graph 1,019 s |
| 2026-11-22 | G19 + cap | Diagnostic `check`, `postgres_lanes` 8 by override | `g19-pg8-check01` | Fail 363/365: deliberate restore-scenario drop past 35 s; short job wait in `TestHostRegistryProfileConsumer_Integration`. Checkpoint sync 569 s total; graph 942 s |
| 2026-11-22 | G19 hardening + cap | Revert G5b; `RemoveDatabase` rename for scenario removal; `requireJob` 1-minute bound | `g19b-restore01`, `g19b-profile01`, `g19b-fixtures01`, `g19b-pgtest01`, `g19b-lintgo01` | All pass; restore scenario issued no drops |
| 2026-11-22 | Validation 1 of 2 | Diagnostic `check`, default capacities | `g19b-check01` | **Pass 365/365.** Graph 970 s, outer 1,097 s; checkpoint sync 594 s; CPU-only host-admission wait 7 s |
| 2026-11-22 | Validation 2 of 2 | Diagnostic `check`, default capacities | `g19b-check02` | Fail 361/365: only the four static checks failed, all on a git worktree another session created inside the checkout mid-run (G20); every service-backed unit passed. Graph 996 s; checkpoint sync 692 s; admission log complete |
| 2026-11-22 | G20 remediation | Targeted: `test-slice` (command-surface parity, owner boundaries), `lint CHECK=shell`, `lint CHECK=frontend`, `source-check CHECK=schema`, `otel-conformance`, `lint-markdown`; nested worktree present throughout | `g20-slice01`, `g20-lintshell01`, `g20-lintfe01`, `g20-schema02`, `g20-otel01` | Pass. `g20-schema01` failed only on stale generated topology digests; fixed by `make generate` (`g20-generate01`) |
| 2026-11-22 | One-time cleanup | Removed fixture copies from retained runs for passed comparisons and packs not named in a failed test's log (60 GiB); leaked `GO_TMP_DIR` entries older than 1 day (55,357, 2.5 GiB); Go build-cache entries unused for 2 days (146 GiB). Disk use 462 → 253 GB | — | Kept: every summary, log, diagnostic, event log, Playwright output, 1,452 Network Flow and 50 IR folders tied to failures, 115 failure-referenced packs, 33 possibly live runs |
| 2026-11-22 | G8 + G10 validation | Targeted slices across six owners; diagnostic `check` | `g810-*01`, `g810-check01` | Slices pass; no new shared Go temp entries. `check` 363/365: a `staticcheck` ST1005 finding in `gotemp_linux.go` (fixed) and `network_flow_portability` (import participant error under load; passes alone in `g810-nfport01`). Check root 249 MB, 174 MB of it the diagnostics event log |
| 2026-11-22 | Regression suite | `harness-contract SUITE=regression` | `clean-regression01`, `clean-regression02` | First run: 5 failures. Fixed 3 caused by committed work: two tests intercept `spawnSync` for the shared Git listing (G20 had switched it to `execFileSync`), and the selected-services test required a drop for every owned database (G19 retirement; now a `postgres-db-retired` event, REQ-357). Remaining 2 are G22 |
| 2026-11-22 | Cleanup audit and `distclean` | Registered missing build outputs and scratch (G21); `make distclean` preview then run | — | 39 roots removed in 53 s; disk use 251 → 127 GB. External Go caches and custom `.cartulary` roots preserved |
| 2026-11-22 | Post-`distclean` check 1 | `make bootstrap`, `make check` | `post-distclean-check01` | Bootstrap failed (G23). Check 352/365 + 10 skipped: `playwright_install` (G23) and two stale-generation failures (generation not re-run after a late edit; fixed by `make generate`). Run root 208 MB without diagnostics |
| 2026-11-22 | Post-`distclean` check 2 | NSS pin updated (G23), `make bootstrap`, `make check` | `post-distclean-check02` | **Pass 365/365.** Graph 1,181 s with cold caches; no new entries in the shared `GO_TMP_DIR`; run root 222 MB, mostly the event log (G9) |
| 2026-11-22 | Packing baseline | Warm diagnostic `check`, cache off | `pack-base-check01` | **Pass 365/365.** Graph 1,152 s. Heavy work 3,263 s (runner 2,528 s, host admission 443 s, fixture acquire 262 s); lanes all busy 1,052 s, one or two busy 93 s; makespan 57 s above perfect balance. Estimate audit: 3 findings (reference-data heavy cohort 302 s over the 300 s bound; two drifted non-heavy cohorts). G24 opened |
| 2026-11-22 | Packing rebalance | 29 row estimates updated from observations (G14) | `pack-generate01`, `pack-catalog01`, `pack-rebal-check01` | **Pass 365/365.** Graph 1,171 s. Makespan excess 58 → 42 s; heavy work 3,375 s (host admission 454 s, fixture acquire 341 s). Both packing runs are about 180 s slower than `g19b-check01` (970 s): total checkpoint sync is now 1,000–1,066 s against 594–692 s then, for the same 91 checkpoints, so the disk is slower or busier. Cause not attributed: the earlier runs were deleted by `distclean`. Candidates: the uncommitted G8/G10/G22/G23 changes and the retirement event, the ATT&CK clock commit `181413312`, or host disk state after about 300 GB of deletions |
| 2026-11-22 | G24 | Heavy-lane precedence in host admission (REQ-803 amended) | `g24-generate01`, `g24-admission01` (row `harness.evidence_accounting.behavior.observation_admission`), `g24-runtime01` (row `harness.command_surface.unit.native_runtime`), `g24-lintgo01`; mutation check (precedence disabled fails the new subtest) | All pass |
| 2026-11-22 | G24 | Diagnostic `check` | `g24-check01` | **Pass 365/365.** Graph 1,022 s (from 1,152–1,171 s); heavy host admission 35 s (from 443–454 s); heavy work 2,921 s; checkpoint sync 871 s |
| 2026-11-22 | W2-0 | Phase-2 reference, second run on the landed code (`fb5c7b63d`; `g24-check01` is the first) | `w20-ref-check02` | **Pass 365/365.** Graph 1,054 s; heavy work 3,027 s (fixture acquire 331 s, host admission 54 s); lanes all busy 985 s; makespan 40 s above a perfect split; checkpoint sync 975 s. Reference pair: 1,022 s and 1,054 s |
| 2026-11-23 | W2-1 / G25 | Side A: `500e3bdda` plus the G23 pin, diagnostic `check` twice | `g25-a-generate01`, `g25-a-check01`, `g25-a-check02` | Pass 365/365 both. Graph 1,112 s and 1,144 s; checkpoint sync 845 s and 1,012 s; heavy host admission 413 s and 386 s (no G24). Same as side B, so the regression is not in the code |
| 2026-11-23 | W2-1 / G25 | Host disk probe: `pg_test_fsync -s 3` in the pinned PostgreSQL image, idle host | — | `fdatasync` 2,102 µs/op, `fsync` 3,785 µs/op, `open_datasync` 1,990 µs/op. Root mounted `ext4 … discard`. Recorded as host baseline |
| 2026-11-23 | W2-2 | G3.6 and G3.9 implemented, measured isolated against clean `HEAD`, then reverted | `w22-imports01`, `w22-nfport01`, `w22-before-imports02`, `w22-before-nfport02` | `network_flow_portability` 177.9 → 174.8 s; `current_workflows` 169.3 → 175.3 s; `ir_corpus_core` 15.7 s with the higher page size. No meaningful gain, so no change kept. A first before-run (`w22-before-*01`) was invalid: a path-limited `git stash` left the working tree unchanged |
| 2026-11-23 | W2-3 | G3.7 rejected (recovery binding digests forbid clone-per-target); G3.4 deferred pending G17 | — | No code change |
| 2026-11-23 | W2-4 / G6 spike | Fresh start vs seed image (copy-on-write, then volume-initialized); template clone and write cost | — | Copy-on-write rejected (clone 6.7 s vs 0.2 s; writes 8× slower). Volume-initialized seed ready in 0.7–1.0 s with fresh-volume write speed. Projected saving about 10 s per generation |
| 2026-11-23 | W2-5 / G2 | Four heavy lanes by capacity override, diagnostic `check` twice | `g2-lane4-check01`, `g2-lane4-check02` | Fail 364/365 each (contention timeouts). Graph 1,029 s and 1,052 s (three lanes: 1,022 s and 1,054 s); heavy work 3,659 s and 3,509 s; checkpoint sync 881 s and 755 s. Rejected |
| 2026-11-23 | W2-6 / G7 | Secret-scan profile, then a memoized matcher with a differential test | `g7-generate01`, `g7-session01` (row `harness.evidence_accounting.behavior.native_session`), `g7-lintgo01`; mutation check | Scan of `g24-check01` root 22.3 s → 0.72 s; all pass; mutation caught |
| 2026-11-23 | W2-7 (interim) | Diagnostic `check` with the memoized scan | `w27-check01` | **Pass 365/365.** Outer 1,168 s, graph 1,157 s; post-graph to scan completion 8 s (was about 28 s); heavy work 3,307 s; checkpoint sync 1,156 s, the highest recorded, so the graph wall reflects host disk state |
| 2026-11-23 | W2-4 / G6 | Seeded generations: targeted rows, Go lint, generated-projection check, Markdown lint | `g6-generate01`, `g6-slice-pg01`, `g6-slice-mig01`, `g6-slice-hb01`, `g6-lintgo01`, `g6-srcgen01`, `g6-mdlint01` | All pass (4, 1 and 9 units). Broker-provided exclusive generation acquired in 1.0 s |
| 2026-11-23 | W2-7 / G6 | Diagnostic `check`, first run | `g6-check01` | **Pass 365/365.** Outer 899 s, graph 887 s; heavy work 2,490 s (from 3,307 s); heavy fixture acquisition 39 s in total, median 2.1 s (from 438 s, median 25.1 s); acquisition across all 126 PostgreSQL units 251 s (from 2,126 s); checkpoint sync 4 s (from 1,156 s); no leaked volumes or builder containers |
| 2026-11-23 | G17 | Contention analysis of the three-lane and four-lane runs: `cgroup_context`, `guest` pressure, process and container samples; unit overlap; template-clone count; one seed-container probe of clone cost | `g6-check01`, `g2b-lane4-check02`, `w27-check01` | CPU 6–9 of 24 cores, stall 1–3% except bursts of 14–21 cores and 21–65% stall from static-analysis operations; both failures sit in a burst. Containers write 62.4 GB; 1,302 clones at 9.8 MB WAL plus 19 MB pages each. G17 rescoped |
| 2026-11-23 | W2-5 / G2 | Four heavy lanes by capacity override after G6, diagnostic `check` twice | `g2b-lane4-check01`, `g2b-lane4-check02` | Pass 365/365, graph 874 s; then fail 363/365, graph 825 s (contention timeouts in the network-flow row-limit and operator-restore tests). Heavy work 3,026 s and 3,019 s; checkpoint sync 4 s. Rejected again |
| 2026-11-23 | W2-7 / G6 | Diagnostic `check`, second run | `g6-check02` | **Pass 365/365.** Outer 900 s, graph 889 s; heavy work 2,537 s; heavy fixture acquisition 56 s, median 2.6 s; all PostgreSQL units 283 s; checkpoint sync 4 s; no leaks. G6 validation met in both runs |
| 2026-11-23 | G17a, G17b | Targeted checks: generation, Go lint, operator process rows, generated and schema projections, Markdown lint, `harness-contract SUITE=execution` | `g17a-generate01`, `g17-lintgo01`, `g17-recovery01`, `g17-srcgen01`, `g17-srcschema01`, `g17-mdlint01b`, `g17-contract-exec01` | All pass |
| 2026-11-23 | G17a | Diagnostic `check` twice, three lanes | `g17a-check01`, `g17a-check02` | Pass 365/365, graph 918 s; then fail 364/365, graph 948 s (`reference_data` refresh-admission matrix, fatal during cancellation in an `io_full` window). Peak CPU stall 47% and 50%. Not adopted |
| 2026-11-23 | G17b, G17e | Diagnostic `check` twice, four lanes by capacity override | `g17a-lane4-check01`, `g17a-lane4-check02` | Fail 364/365, graph 966 s; fail 363/365, graph 921 s (network-flow row limit both times; `incidentbundles` ATT&CK round trip once). Operator process rows passed: G17b met. Four lanes rejected a third time |
| 2026-11-23 | G17 | Contention re-analysis: per-window `guest` CPU and I/O pressure, unit overlap, process CPU of the network-flow unit, source-pass timings | The eight runs above and `g6-check01/02`, `g2b-lane4-check01/02` | All four failures fall in `io_full` windows of 13–27% with CPU stall 0–10%; CPU stall totals 21–35 s per `check`. G17d: passes share the 30 s budget, three-lane margin 20–30%; no change |
| 2026-11-23 | G17a, G17b | Commits; projection check after removing G17a | `g17-srcgen02` | G17b `3373cd3b5`; G17a kept on local branch `g17a-attempt` (`2f23c6e44`); generated projections match their inputs |
| 2026-11-23 | G17c | Targeted checks: generation, `harness-contract SUITE=contract`, JSON-shape and Go row-runner rows, generated and schema projections, script lint | `g17c-generate01`, `g17c-contract01`, `g17c-shapes01`, `g17c-srcgen01`, `g17c-srcschema01`, `g17c-lintscripts02` | All G17c checks pass. `g17c-contract01` also failed `harness.command_surface.behavior.owner_boundaries`: `installed-readiness.mjs` imports private browser modules, from G22's `a5df06952`; not in `check`; split into a separate task |
| 2026-11-23 | G17c | Diagnostic `check` | `g17c-check01` | Fail 364/365 (`incidentbundles` `network_flow_portability`, during a concurrent WAL probe; run mean `io_full` 13.3%). Durations retained for 312 of 312 Go units; retained-secret scan pass. Committed `b98e95e17` |
| 2026-11-23 | G17f | Shutdown-checkpoint trace (Docker events), I/O scheduler and cgroup controllers, clone census from journals, `wal_compression` probe on the seed image | `g17a-check01`, `g17c-check01` | No shutdown checkpoint (force removal, SIGKILL); no I/O weights; 1,300 clones over 579 tests; `lz4` −60% WAL per clone. A17 withdrawn, A18 drafted |
| 2026-11-23 | A18 | Go lint, PostgreSQL rows, seeded-generation template row, then one diagnostic `check` | `a18-lintgo01`, `a18-slice-pg01`, `a18-slice-hb01`, `a18-check01` | Targeted checks pass. `a18-check01` fail 364/365 (`incidentbundles` ATT&CK round trip, fixed 30 s poll under `io_full`); container writes 48.6 GB (from 61.8–63.0 GB); `io_full` 10.6%; graph 961 s |
| 2026-11-23 | G17g | Generation, Go lint, httptestx helper row with the new deadline test, generated projections, Markdown lint | `g17g-generate01`, `g17g-lintgo01`, `g17g-slice01`, `g17g-srcgen01`, `g17g-mdlint01` | All pass. Committed `4a30fd00c` |
| 2026-11-23 | G17 | Final validation: paired diagnostic `check` on `4a30fd00c` | `g17-final-check01`, `g17-final-check02` | **Pass 365/365 twice.** Graph 857 s and 882 s (G6 pair 887 s and 889 s); heavy work 2,383 s and 2,496 s; container writes 46.2 GB and 48.4 GB (from 61.8–63.0 GB); mean `io_full` 7.3% and 8.7%. G17 closed |
| 2026-11-23 | G27 | Profiling of the three heaviest tests with temporary phase marks (removed after) | `g27-prof-cw01`, `g27-prof-nf01`, `g27-prof-attack01`, `g27-prof-cw02` | Workflow test idle about 6 s per subtest waiting on the 5 s recovery scan after restart; transfer and ATT&CK time is product work. With a 100 ms injected scan the workflow row fell from 191 s to 86 s (instrumented) |
| 2026-11-23 | G27 | Targeted checks, then one confirming diagnostic `check` | `g27-generate01`, `g27-lintgo02`, `g27-srcgen01`, `g27-slice-server01`, `g27-slice-imports01`, `g27-check01` | All pass; `check` 365/365. Workflow test 116 s (from 208–221 s). Graph 945 s on noisy host disk (`io_full` 11.5%). Committed `7a62a435a` |

## 12. Phase 2 Plan — `check` Wall Clock

**Status.** Proposed 2026-11-22, after G24; worked through 2026-11-23. It
supersedes Section 7 for the remaining `check` wall-clock work; Section 7 keeps
the phase-1 history.

| Workstream | Outcome |
| --- | --- |
| W2-0 | Done: plan committed; reference pair 1,022 s and 1,054 s |
| W2-1 (G25) | Done: host disk state, not code; `pg_test_fsync` baseline recorded |
| W2-2 (G3.5, G3.6, G3.9) | Measured, no material gain; reverted or dropped |
| W2-3 (G3.7, G3.4) | G3.7 rejected (recovery binding fence); G3.4 deferred pending G17 |
| W2-4 (G6) | Done: harness-assigned identity and volume-initialized seed images; graph 887 s in `g6-check01` |
| W2-5 (G26, G2) | G26 not worth implementing; four lanes rejected, and rejected again after G6 |
| W2-6 (G7) | Done: memoized scan matcher; post-graph time about 28 s → 8 s |
| W2-7 | Done: `g6-check01` and `g6-check02` passed, graph 887 s and 889 s; Section 3 targets and the gap register refreshed |

Before G6, graph wall on unchanged code ranged 1,022–1,157 s with checkpoint
sync 755–1,156 s. G6 removed almost all of that sync (see the G25
follow-up): graph wall 887 s, checkpoint sync 4 s. `check` is again bound by
heavy-lane test work (2,490 s ÷ 3 ≈ 830 s).

### 12.1 Starting Point

`g24-check01` (warm, diagnostics on, cache off) passed in 1,022 s. The three
heavy lanes were all busy for 946 s, so the wall is heavy-lane work divided by
three plus packing imbalance:

| Component | Value |
| --- | ---: |
| Heavy-lane work | 2,921 s |
| Of which test time | 2,442 s (top 8 rows 1,097 s) |
| Of which PostgreSQL generation acquisition | 354 s (median 19 s per unit) |
| Of which host admission and fixture release | 69 s |
| Makespan above a perfect three-lane split | 42 s |
| Post-graph time (scan and publication) | about 28 s |
| Checkpoint sync, whole run | 871 s (594–692 s in the morning's passing runs) |

Every second removed from heavy-lane work saves about a third of a second of
wall at three lanes. The ≤ 750 s target needs about 800 s less heavy work at
three lanes, or a fourth lane that the disk can sustain.

### 12.2 Principles for Phase 2

The Section 3 rules still apply. In addition:

1. Remove repeated work before adding parallelism. On this host the disk, not
   the CPU, saturates first.
2. A test keeps every assertion. Shared setup is built once from the same
   public operations the test would run, then cloned; anything a test mutates
   is cloned, never shared.
3. One mechanism per concept. Snapshot, clone and seed reuse the harness's
   existing primitives (`SealTemplateDatabase`, `CloneDatabase`, the
   performance-fixture object snapshot) instead of adding parallel ones.
4. Production policy is never weakened to speed tests. Shorter timers enter only
   through harness-owned composition, which the Core specification already
   permits for job timers.
5. Measure every treatment twice. Single-run wall varies by about ±150 s, so
   judge by heavy-lane work and lane occupancy first and wall second.

### 12.3 Gap Remediation

#### G25 — Unattributed I/O regression

- **Evidence.** This morning's passing runs had graph wall 970–996 s and
  checkpoint sync 594–692 s. Every run since `distclean` shows 1,022–1,181 s
  and 871–1,066 s for the same 91 checkpoints. The earlier roots were deleted,
  so nothing can be compared directly.
- **Remediation.** Compare `500e3bdda` with the G23 pin cherry-picked against
  `a5df06952`, twice each. Pre-G23 commits cannot bootstrap: the old pinned
  package is gone upstream. Each run uses a worktree outside the checkout (G20).
  If the bundle is at fault, revert its G8 and G10 parts separately in a scratch
  worktree to isolate them. If both are slow, test the host: record raw `fsync`
  latency with `pg_test_fsync` in the pinned PostgreSQL image, then repeat after
  host-side trim and compaction of the WSL disk image. Fix the culprit, or
  record the new host baseline and move every target with it.
- **Area.** Measurement; implementation if a commit is at fault.
- **Rationale.** Every later saving is 10–70 s; an unexplained 50–150 s drift
  makes them unmeasurable.
- **Long-term benefit.** A recorded disk probe makes later I/O regressions
  attributable in minutes instead of hours.
- **Compatibility.** None.
- **Risk if unresolved.** Savings are misattributed, or a real regression ships.
- **Validation.** Two runs on each side of the bisection reproduce or rule out
  the difference beyond noise. A fix, if any, restores checkpoint sync to the
  morning range. The `fsync` probe result is recorded here.
- **Result (W2-1, 2026-11-23).** Not the code. Side A (`500e3bdda` plus the G23
  pin, the code of the morning's 970 s run) took 1,112 s and 1,144 s with
  checkpoint sync 845 s and 1,012 s (`g25-a-check01`, `g25-a-check02`). Side B
  (`a5df06952` content) took 1,152 s and 1,171 s with 1,000 s and 1,066 s
  (`pack-base-check01`, `pack-rebal-check01`). The slowdown is host disk state.
  Likely mechanism: the root filesystem is mounted with `discard`, so the
  day's roughly 300 GB of deletions were passed to the WSL virtual disk as TRIM;
  later writes land in deallocated space and pay allocation on the Windows host.
  Host baseline with `pg_test_fsync` in the pinned PostgreSQL image, idle host:
  `fdatasync` 2.1 ms per 8 kB write (476 ops/s), `fsync` 3.8 ms, `open_datasync`
  2.0 ms. Remedies are host-side and need the owner: compact the virtual disk
  from Windows after `wsl --shutdown`, or mount without `discard` and trim on a
  schedule. The phase-2 targets are measured against the reference pair
  (1,022 s and 1,054 s); rerun the probe before comparing any later treatment.
- **Follow-up (G6, 2026-11-23).** Most of that checkpoint sync was
  provisioning I/O, not test I/O. Each generation's template-seal checkpoint
  fsynced about 1,433 files that `initdb` and the migration had written: 13–15
  s per generation, 2,614 s of checkpoint time in `w27-check01`. Seeded
  generations receive those files by Docker volume copy, so each seal
  checkpoint now syncs 5 files in about 9 ms. Total checkpoint sync per `check`
  fell from 1,156 s to 4 s (`g6-check01`). The disk state explains why that
  cost varied from run to run; the host remedies above now matter much less.

#### G3.6 — Network Flow transfer re-reads frozen state

- **Evidence.** `TestNetworkFlowIncidentTransferThroughPublicOperations_Integration`
  (254 s) verifies both destinations against side A. Side A stops changing
  after line 158, yet each verification re-reads it: investigation views, three
  tables and about 23 table pages at limit 200. The API allows 500.
  `irCommittedSources` pages at limit 1.
- **Remediation.** Read side A once, after it stops changing, into an immutable
  expectation value and compare both destinations against it. Page table and
  view reads at the API maximum. Keep one deliberately multi-page traversal so
  pagination stays covered. Do not split the unclaimed-destination leg: it
  needs the full bundle, so a split repeats the whole side-A setup and adds
  work.
- **Area.** Tests.
- **Rationale.** Re-reading frozen state is repeated work, not coverage.
- **Long-term benefit.** The row falls under the 225 s packing target, so the
  packer can pair it instead of running it alone.
- **Compatibility.** None. The row ID is unchanged; its estimate is refreshed.
- **Risk if unresolved.** The longest single row keeps setting the heavy tail.
- **Validation.** Identical assertions; a probe that changes one expected value
  fails; row time drops in two checks (estimate 15–30%).

#### G3.7 — Saved-graph restore fan-out

- **Evidence.** `assertFrozenReportingRestores` captures once, then restores the
  same capture 15 times into fresh databases and object-store namespaces:
  resumption, 8 authority faults (one nested) and 5 retained-input corruptions.
  appsupport has no "restore once, clone per target" helper.
- **Remediation.** Add one appsupport helper. It restores a capture once,
  seals the database (`SealTemplateDatabase`) and the object namespace,
  then gives each target a database clone (`CloneDatabase`,
  `RetireClonedDatabaseT`) and its own copy of the objects. Reuse the
  performance-fixture object snapshot (`sealObjectSnapshot`,
  `restoreObjectSnapshot`) by moving it to shared test support, with one owner,
  rather than writing a second copier.
- **Area.** Test support (appsupport, pgtest, performance-fixture support);
  tests.
- **Rationale.** Cloning is the harness's existing isolation primitive; fifteen
  identical restores add nothing.
- **Long-term benefit.** Recovery fault probes, including the ATT&CK recovery
  test's three, can use the same helper.
- **Compatibility.** The object-snapshot code moves package; its single caller
  moves with it.
- **Risk if unresolved.** Every new fault target adds a full restore.
- **Validation.** Every fault still produces its expected outcome. An isolation
  probe shows one target's mutation is invisible to the next, for both database
  and objects. Row time drops (estimate 30–50% of 84 s).

#### G3.4 — ATT&CK fixture cost

- **Evidence.** Each test builds its own signing context with random keys, so
  identical packs get different digests. Tests import the 95 MB pack two or
  three times (the lifecycle "replay" uses a new `client_txn_id`, so it is a
  full second import); recovery restores it four times; reporting imports it
  once for setup only. Each import writes about 40,853 index rows in several
  hundred transactions.
- **Remediation.**
  1. One deterministic signing context per test process: keys derived from a
     fixed seed and `now` pinned to one second through the existing
     `ReferencePackClock`. Pack bytes are memoized per dataset and version.
     Tests that assert key or expiry behavior keep explicit contexts.
  2. For tests whose subject is not import (reporting composition preview,
     incident-bundle source setup, recovery setup), import once per package
     process into a sealed template database plus a read-only pack-storage
     snapshot, and clone per test with the G3.7 helper.
  3. Owner review of redundant passes: decide whether the lifecycle "replay"
     should assert receipt replay (same `client_txn_id`) or a second import; move
     recovery's three fault probes onto the G3.7 helper.
- **Out of scope.** COPY-based index writes would speed production imports
  too, but batching is product behavior under the reference-pack NLSpec. Raise it
  with that owner separately.
- **Area.** Test support (`reference_data/testsupport`,
  `referencepackfixture`); tests.
- **Rationale.** Identical inputs should yield one artifact; setup that is not
  under test should be built once.
- **Long-term benefit.** New reference-pack tests start from a snapshot instead
  of adding a 20–40 s import each.
- **Compatibility.** None for product code.
- **Risk if unresolved.** ATT&CK rows stay contention-sensitive (81–151 s
  spreads) and keep growing with pack size.
- **Validation.** Container digests are identical across tests in one run.
  Each changed test keeps its assertions. Owners' targeted rows pass. Heavy
  time for the four ATT&CK rows drops (estimate 60–120 s in total).

#### G3.9 — Production job timers in fault tests

- **Evidence.** `server.Options` has no job-policy field;
  `runtime_assembly.go` hard-codes `ProductionRuntimePolicy()` (5 s recovery
  scan, 5 s and 30 s retries) for manager and runner. Fault tests wait on these
  real timers (about 45 s in `current_workflows`, 12–15 s in incident bundles).
  Core 01 makes the values normative and says "Tests MAY inject shorter values
  through harness-owned composition; deployment configuration MUST NOT redefine
  these values."
- **Remediation.** Add one composition field, `server.Options.JobPolicy`, where
  nil means the production policy; it drives both manager and runner.
  appsupport sets a short policy only for fault tests that exercise retry or
  recovery and do not assert production cadence. Replace the tests that reach
  the unexported dependency hook with this field.
- **Area.** Implementation (`internal/app/server` composition, appsupport);
  tests. No specification change.
- **Rationale.** The seam the specification already permits, in one place,
  instead of per-test workarounds.
- **Long-term benefit.** Fault tests run at test speed without touching
  production defaults.
- **Compatibility.** None. `cmd/server` never sets the field.
- **Risk if unresolved.** Each new fault test adds 5–35 s of idle waiting.
- **Validation.** A composition test proves deployment configuration cannot set
  the policy and that nil yields production values. The existing manager tests
  still assert the production values. Fault tests keep their classification
  assertions and get faster.

#### G3.5 — Duplicated plain sources

- **Remediation.** Share one plain source across the reporting, embedded and
  main flows of
  `TestIncidentReferenceCatalogSurvivesUnavailablePackRoundTrip_Integration`.
- **Area.** Tests. **Benefit.** Estimate 10–20 s. **Compatibility.** None.
- **Validation.** Same assertions; row time.

#### G6 — Seeded PostgreSQL generations (replaces the earlier G6 plan)

- **Evidence.** Every exclusive generation runs `initdb`, roles, template
  creation, the full migration chain, seal and checkpoint. That is a median of
  19 s per heavy unit, 354 s of heavy work and about 1,800 s per `check` across
  89 generations. Everything is discarded at retirement. Attachments are keyed
  by PostgreSQL's `system_identifier`, which a copied or seeded data directory
  would share. `ProvisioningDigest` and `TemplateDigest` already identify
  everything a seed depends on.
- **Remediation.**
  1. **Generation identity (G6a).** Each generation receives a harness-minted
     identity at initialization, recorded in the cluster and verified on
     attach. Attachments bind it instead of `system_identifier`. This is a
     prerequisite for seeding and also removes a hidden dependency on `initdb`.
  2. **Seed image (G6b).** Build one immutable seed image per `TemplateDigest`.
     It holds an initialized, role-provisioned, migrated, sealed and
     checkpointed data directory at a PGDATA path outside the base image's
     declared volume, so containers start copy-on-write. The full migration
     chain runs in the seed build, which is the owner proof that line 840 of the
     NLSpec requires; checksums and every durability setting are unchanged.
     Exclusive generations start from the seed, set their identity, and publish
     the already-sealed template. There is no per-generation `initdb` or
     migration path left to maintain.
  3. **Ownership.** The seed is a content-addressed build artifact owned by the
     readiness cache. It is verified on each use (the G22 rule), labelled for
     cleanup, and never holds live data. It is not an allocation retained
     across runs (NLSpec line 760); the specification must say so explicitly.
- **Alternatives rejected.** Copying a data directory per generation adds I/O
  on a disk-bound host and still needs G6a. Prefetching generations ahead of the
  lane moves work instead of removing it, and resembles the asynchronous
  replenishment the NLSpec forbids.
- **Area.** Specification (A3, rewritten: REQ-804 and REQ-811 identity, seed
  provenance, artifact ownership); implementation (`pgtest` owned cluster,
  `testservices`, attachment registry, readiness cache); tests.
- **Rationale.** The migration chain is deterministic for a digest; running it
  89 times per `check` proves nothing the first run did not.
- **Long-term benefit.** New migrations stop multiplying every generation's
  start, and PostgreSQL lanes free sooner, which also helps ordinary units.
- **Compatibility.** The attachment identity cuts over; attached sessions with
  the old identity are rejected and rebuilt. The image gains a cleanup
  registration.
- **Risk if unresolved.** Startup grows with every migration, multiplied by
  every exclusive unit.
- **Main risk of the fix.** Write performance on the overlay filesystem
  (copy-up on first write of each relation file). A spike measures it before
  adoption; if it hurts, populate a named volume from the seed instead and
  measure again.
- **Validation.** Median exclusive acquisition ≤ 6 s. The REQ-804 acceptance
  checks for template and clone isolation and durable WAL progress pass. The
  migration owner keeps its full-chain proof. Two checks show heavy-unit
  acquisition below 100 s in total.
- **Implementation (2026-11-23).** Amendment A3 landed in REQ-804, REQ-811
  and the two clarifications. In `internal/testutil/pgtest`:
  - *Identity (G6a).* Each startup attempt mints a positive 63-bit decimal,
    passes it as `-c cluster_name=…` outside the provisioning digest, and
    rejects a server reporting another value. Attachment and drop observation
    read `current_setting('cluster_name')` instead of `system_identifier`.
  - *Seed (G6b).* `seed.go` resolves `cartulary-postgres-seed:<TemplateDigest>`
    under a host-wide lock: shared to use, exclusive to build, prune or reclaim
    builders. The builder is an ordinary fresh `startOwned` cluster with PGDATA
    outside the declared volume. It runs cluster prerequisites, creates and
    migrates the template without sealing it, deletes its TLS files and init
    script, and shuts down cleanly. Then it is committed with a digest label,
    and inherited Testcontainers labels are blanked so no reaper can claim the
    image. Generations start from the image ID with an anonymous volume over
    `/pgseed`. A generation renames the seed template to its own template name
    under catalog admission (same ownership ledger and creation event), then
    seals it. Builds keep three seeds and never fail startup on pruning.
  - *Corrections found while testing.* A sealed seed template is marked
    `IS_TEMPLATE`, and PostgreSQL refuses to drop it, so disposing a failed
    publication left it behind. The seed now carries an unsealed migrated
    template, and sealing stays per generation.
  - *Measured.* Seed layer 78 MB; first build about 10 s on an idle host.
    Seeded generation through the broker acquired in 1.0 s
    (`g6-slice-pg01`). Two owned seeded generations started, adopted,
    cloned and closed in 3.5 s.

#### G26 — Packing omits generation acquisition

- **Evidence.** Heavy cohorts are packed against 225 s using row estimates plus
  a 5 s Go startup allowance. The closed lifecycle-estimate set has no
  PostgreSQL generation acquisition, so about 19 s per heavy unit is invisible
  to packing and ranking.
- **Remediation.** Add `postgres_generation_acquire` to the closed
  lifecycle-estimate set. The packer reserves it once per exclusive cohort, as
  the specification already does for browser acquisition. Re-derive the value
  after G6, and refresh row estimates from two qualified runs after W2-2 and
  W2-3.
- **Area.** Specification (lifecycle estimates beside REQ-865); machine input
  (`lifecycle_estimates_ms`); implementation (compiler packing); tests.
- **Rationale.** Packing should charge every cost the cohort pays exactly once.
- **Long-term benefit.** Balance stays correct as acquisition cost changes.
- **Compatibility.** Cohort composition and unit IDs change; this is planning
  data only.
- **Risk if unresolved.** Cohorts overrun the target, and the tail stays uneven.
- **Validation.** The estimate audit reports no `over_packing_bound`; makespan
  above a perfect split ≤ 20 s in two checks.
- **Decision (2026-11-23).** Not implemented. Acquisition costs about the same
  for every heavy cohort, so charging it changes no longest-first order among
  heavy units, and charging it in packing only makes cohorts smaller, which adds
  generations that each pay about 19 s. The remaining 33–49 s imbalance is
  granularity, and G6 shrinks the per-cohort overhead directly. A14 is
  withdrawn.

#### G2 (revisited) — Heavy lane count

- **Remediation.** After G25 is settled and W2-2 to W2-4 land, run three lanes
  against four (capacity override), twice each. Adopt four as the default only
  if graph wall improves beyond noise, checkpoint sync stays within 20% of the
  three-lane runs, and no cleanup or deadline failure appears (the G19 pattern).
- **Area.** Measurement; specification §10.2 default if adopted (A15).
- **Risk.** Disk saturation turns speed into flakiness, as G5b showed.
- **Validation.** The acceptance criteria above, recorded with run IDs.
- **Result (2026-11-23).** Rejected; A15 is not adopted. Four lanes by override
  (`g2-lane4-check01`, `g2-lane4-check02`): graph 1,029 s and 1,052 s against
  1,022 s and 1,054 s on three lanes. Heavy work rose from 2,921–3,027 s to
  3,509–3,659 s (runner time alone 2,494–2,607 → 2,940–2,955 s), so the fourth
  lane's capacity went to contention. Each run also failed one unit on a
  contention timeout (`extension_transaction_timeout` in
  `TestNetworkFlowPublicDefaultLimitsCommitAcceptedRowLimit`; an operator
  process killed on its deadline in `TestCanonicalOperatorRestoreBundle_Process`).
- **Re-measured after G6 (2026-11-23).** Still rejected. The first run predated
  W2-4, contrary to the plan's sequencing, so it was repeated once G6 removed
  the checkpoint I/O (`g2b-lane4-check01`, `g2b-lane4-check02`). Graph 874 s
  (pass) and 825 s (2 units failed, so the wall is not comparable) against
  887 s and 889 s on three lanes. Heavy work rose from 2,490–2,537 s to
  3,019–3,026 s (runner 2,379–2,407 → 2,804–2,806 s); checkpoint sync stayed
  at 4 s. The same two tests failed together in the second run: the import's
  second source pass took 24.1 s instead of 5.3 s before
  `extension_transaction_timeout`, and the operator restore process was killed
  on its deadline (exit −1). The disk is no longer the constraint: heavy units
  slow each other on CPU and time-bounded work, so the fourth lane gains
  nothing beyond noise and makes two fixed-deadline tests flaky.

#### G7 (done, uncommitted at writing) — Secret scan

- **Profile (`g24-check01` root, 229 MB, 4,404 files).** The scan took 22.3 s;
  the 177 MB event log took 16.6 s, and 16.5 s of that was the syntactic
  matcher alone. Per-secret `bytes.Contains` was negligible, so the planned
  worker pool and multi-pattern automaton would have missed the cost: one large
  file dominates, and the matcher ran at about 10.7 MB/s.
- **Remediation.** The matcher was a Thompson NFA simulation walking the
  program for every rune. It now memoizes each step as a lazily built DFA
  whose states are exactly the Thompson counter sets, with a per-matcher bound
  of 2,048 states that clears on overflow. Memoization is exact because the
  policy's only assertions are end-of-text (`$`), which depends on the current
  rune alone. A program with any assertion that reads the previous rune keeps
  the Thompson path. Code: `scan_syntax_linux.go`; no specification change
  (REQ-600 already permits it).
- **Validation.** A differential subtest feeds 6,000 seeded inputs built from
  the policy's tokens, marker fragments, multi-byte and invalid UTF-8 through
  both paths at random chunk boundaries and requires identical verdicts. A
  mutation that drops half the pending counters fails it. The existing scan
  subtests pass (row `harness.evidence_accounting.behavior.native_session`),
  and `lint CHECK=go` passes. The same root now scans in 0.72 s instead of
  21.7 s, so every graph target saves about 21 s of outer wall.

### 12.4 Workstreams

```text
W2-0 ─> W2-1 ─┬─> W2-4 ──┐
              ├─> W2-2 ──┼─> W2-5 ─> W2-7
              └─> W2-3 ──┘
W2-6 (independent) ───────────────> W2-7
```

#### W2-0 — Land and baseline

- **Scope.** The phase-1 work and this plan.
- **Status.** The owner landed the work before this step:
  - `a5df06952` (pushed) bundles G8, G10, the G19 retirement event, the G20
    `spawnSync` fix, G21, G22 and G23;
  - `fb5c7b63d` holds G24 and the G14 estimates.

  The published bundle is not rewritten; W2-1 isolates its parts instead. This
  plan is the remaining commit.
- **Steps.** Commit this plan; take two warm diagnostic `check` runs on the
  landed code as the phase-2 reference (`g24-check01` ran on the same executable
  source and counts as the first).
- **Dependencies.** None.
- **Exit criteria.** Clean tree; two reference runs recorded in Section 11.
- **Result.** Complete. The plan is committed as `dd8dfdd62`, and the reference
  pair is `g24-check01` (1,022 s) and `w20-ref-check02` (1,054 s), with
  checkpoint sync of 871 s and 975 s.

#### W2-1 — Attribution

- **Scope.** G25. Optional enabler: G17 (subtest timings) if the owner settles
  its naming rule, because it makes W2-2 and W2-3 measurable inside tests.
- **Dependencies.** W2-0.
- **Risks.** Host noise; mitigate with two runs per side and the `fsync` probe.
- **Exit criteria.** Cause identified and fixed, or a host baseline recorded
  and Section 3 targets restated against it.

#### W2-2 — Test-local heavy fixes

- **Scope.** G3.6, G3.5, G3.9.
- **Dependencies.** W2-0. Can run alongside W2-1; measure after it.
- **Risks.** Lost coverage; every change keeps its assertions and gets a
  mutation probe.
- **Exit criteria.** Owners' targeted rows pass; heavy work down by at least
  100 s across two checks.
- **Result (2026-11-23).** Not met; both changes reverted after measurement.
  Isolated rows against clean `HEAD` (`w22-before-imports02`,
  `w22-before-nfport02` against `w22-imports01`, `w22-nfport01`):
  - G3.6, capturing side A once at limit 500: `network_flow_portability`
    177.9 s → 174.8 s.
  - G3.9, advancing the recovering server's clock to the persisted retry
    instant: `current_workflows` 169.3 s → 175.3 s. Closing and restarting the
    server already outlasts the 5 s retry delay, so the job was due anyway.

  The code-reading estimates (15–30% and about 45 s) were wrong: exports,
  imports, restores and server starts dominate these rows, not re-reads or
  timers. G3.5 was dropped before implementation: the reporting source needs
  template packs, so sharing would save one server start (about 3 s). D7 is not
  needed.

#### W2-3 — Shared fixture support

- **Scope.** G3.7 helper, then G3.4.
- **Dependencies.** W2-0; G3.4 step 2 needs the G3.7 helper.
- **Risks.** Shared state leaking between tests; the isolation probe is
  mandatory.
- **Exit criteria.** Identical ATT&CK digests within a run; isolation probes
  pass; heavy work down by a further 80 s or more.
- **Result (2026-11-23).** Not started; both items stopped on evidence.
  - G3.7 is rejected as designed: a restored target records binding digests
    for its database, object store, pack storage and export outputs
    (`recovery_target_state`), which the server checks at startup. A database
    clone has a different binding, so clone-per-target would need a bypass of a
    recovery safety fence.
  - G3.4 is deferred. Signing cost is mostly hashing and archiving 95 MB
    (about a second per build), and the import-snapshot saving is an unmeasured
    code-reading estimate of the kind W2-2 just disproved. Resume it only with
    per-subtest timings (G17) showing import time inside the ATT&CK rows.

#### W2-4 — Seeded generations

- **Scope.** G6a, then a G6b spike, then G6b.
- **Steps.** Spike: build one seed by hand and measure container start and an
  I/O-heavy test on overlay against a fresh generation. Then rewrite A3, land
  G6a with its attachment cut-over, then G6b.
- **Dependencies.** W2-1 (stable baseline); decision D6.
- **Risks.** Overlay write cost; attachment cut-over breaking attached
  sessions. Mitigate with the spike and with rejection, not translation, of old
  attachments.
- **Exit criteria.** G6 validation met; REQ-804 acceptance checks pass.
- **Spike result (2026-11-23, idle host, pinned PostgreSQL image).**
  - Fresh container (entrypoint `initdb`) is ready in 1.5 s warm (4.1 s cold).
    A seed with PGDATA outside the declared volume starts in 0.56 s.
  - Copy-on-write seed: rejected. On the overlay filesystem a template clone
    took 6,669 ms against 205 ms on a volume, and a 200k-row insert with
    checkpoint 6,227 ms against 740 ms. Copy-up of whole relation and WAL
    files on first write makes every test clone slow.
  - Volume-initialized seed: viable. An anonymous volume mounted over the
    seed's data path is filled from the image at create (84 MB synthetic data
    directory). Ready in 0.70–0.96 s, clone 244 ms, and writes match a fresh
    volume in alternating trials (417–467 ms against 440–484 ms).
  - Projection: each generation skips `initdb`, roles, template migration
    (median 9.8 s under load), seal and checkpoint, about 10 s per generation,
    so roughly 170–200 s of heavy work and about 60 s of wall.
  - G6b therefore uses volume initialization; Docker's copy replaces
    `initdb` and the migration's writes, so per-generation I/O stays about
    the same.

#### W2-5 — Scheduling

- **Scope.** G26, estimate refresh, then the G2 lane experiment.
- **Dependencies.** W2-2, W2-3 and W2-4, so estimates reflect the new costs.
- **Exit criteria.** G26 validation met; lane-count decision recorded with
  evidence.

#### W2-6 — Post-graph

- **Scope.** G7.
- **Dependencies.** None.
- **Exit criteria.** Same verdicts on the redaction and scan fixtures;
  post-graph time ≤ 15 s.

#### W2-7 — Validation and handoff

- **Steps.** Two final diagnostic `check` runs; Section 3 targets updated with
  the outcome; amendments adopted with projections regenerated; test-ops skill
  recipes updated only if public commands changed; tracker and memory updated.
- **Exit criteria.** Every gap above closed or explicitly deferred with data.

### 12.5 Expected Effect

Estimates, before measurement; heavy-work savings convert to wall at about
one third.

| Lever | Heavy work saved | Wall saved |
| --- | ---: | ---: |
| G25 | — | 0–150 s, unknown |
| G3.6, G3.5, G3.9 | 100–160 s | 35–55 s |
| G3.7, G3.4 | 80–160 s | 25–55 s |
| G6 | about 220 s | about 70 s |
| G26 | — | 10–20 s |
| G7 | — | about 13 s, outer only |
| G2, four lanes | — | unknown; up to about 240 s if the disk holds |

Without G2 and G25 the expected graph wall is about 820–880 s. Reaching
≤ 750 s needs either a fourth lane the disk sustains, or G25 recovering the
morning's I/O headroom.

Measured: G6 alone saved about 800 s of heavy work and brought the graph wall
from 1,022–1,157 s to 887–889 s, far above its estimate, because it also
removed the provisioning checkpoint I/O behind G25. G7 saved about 20 s outside
the graph. The other levers gave nothing measurable or were rejected, including
four lanes after G6.

**Phase 2 closed (2026-11-23).** Every Section 12 gap is closed or deferred
with data. `check` (887–889 s graph, about 900 s outer) misses the 750 s target
by about 140 s. It is bound by heavy-row test work (about 2,500 s ÷ 3 lanes),
not by provisioning, admission, the disk or the post-graph scan. The remaining
lever is that test work itself. [G17](#g17--heavy-lane-critical-path-attribution-and-contention)
now holds the plan: bound operations to their CPU claim, derive process-test
hang guards from the test deadline, retain per-subtest timings, then test
I/O. Outside wall clock, open items are G9 stage 2 (run root 242–249 MB, event
log 178–185 MB), G15, G16, G11, G12 and G13.

Update after G17a and G17b (2026-11-23): the process-test guards landed; the
operation CPU bound did not remove the CPU bursts and was not adopted; four
lanes failed a third time. Disk stall, not CPU, is the contention that the
failing tests hit, so G17c and then G17f come next. G17 has since closed (Section 5): G17c, G17g and A18 landed, A17 was withdrawn, and G27 and G28 hold what remains.

### 12.6 Amendments and Decisions

| ID | Section or requirement | Change | Gap |
| --- | --- | --- | --- |
| A3 (rewritten) | REQ-804, REQ-811; line 760 clarification | Harness-assigned generation identity; exclusive generations start from a content-addressed seed proven by a full-chain build; the seed is an immutable build artifact, not a retained allocation | G6 |
| A14 | Lifecycle estimates beside REQ-865 | `postgres_generation_acquire` joins the closed set and is reserved once per exclusive cohort | G26 |
| A15 (conditional) | §10.2 capacity table | Heavy lanes default 4, only if the G2 criteria hold | G2 |

| ID | Decision | Needed by | Recommendation |
| --- | --- | --- | --- |
| D6 | Approve rewritten A3 and the seed-image approach after the spike | W2-4 | Approved 2026-11-23 and implemented |
| D7 | Add `server.Options.JobPolicy` (product composition) | W2-2 | Approve; Core 01 already permits it |
| D8 | Lifecycle "replay": receipt replay or second import | W2-3 | Reference-data owner decides; keep whichever the NLSpec claim names |
| D9 | Commit plan for the uncommitted work | W2-0 | Logical commits, so W2-1 can bisect |
| D10 | Four-lane adoption criteria | W2-5 | As stated under G2 |
