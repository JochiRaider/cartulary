# Test Execution Efficiency Remediation Plan

**Status.** Active remediation tracker, last updated 2026-11-22. Committed through `500e3bdda`: G3 fixes 1–3 and 2b, G4a, G5a, G9 stage 1, G19 with hardening, G20. Uncommitted in the working tree: G8, G10, G21, G22, G23, the database-retirement event, and fixes for regressions the regression suite exposed in G19 and G20. `make distclean` on 2026-11-22 deleted every retained run, so the run IDs in Section 11 before that date are a record, not retrievable evidence.

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
| `check` graph wall | 1,022 s warm (`g24-check01`); best earlier 970 s (`g19b-check01`, before the unexplained disk slowdown) | ≤ 750 s |
| `check` outer wall | 1,097 s (`g19b-check01`) | ≤ 800 s |
| Post-graph overhead per invocation | ~60 s (`post-distclean-check02`; was ~121 s before G8) | ≤ 15 s |
| CPU-only units waiting more than 300 s | 26 (`g5-check01`; was 188) | 0 |
| Successful `check` run root | 222 MB (`post-distclean-check02`; was 1.1 GB) | ≤ 150 MB |
| `unit-events.ndjson` per `check` | ~163 MB (98% wait-segment pairs) | ≤ 10 MB |
| Temporary directories leaked per clean run | 0 (G10; was ~20) | 0 |
| Retained results | Reset by `make distclean` 2026-11-22; still unbounded | Bounded by an owner-set budget |
| `ci` graph wall | 3,279 s (`ws1-ci01`; not re-measured) | ≤ 1,500 s |
| `release-check` wall | Unmeasured | Measured; correctness portion ≤ 1.5 × `check` |

## 4. Gap Register

| ID | Gap | Areas | Workstream | Priority |
| --- | --- | --- | --- | --- |
| G1 | `gofmt` drift keeps `make check` red at HEAD (resolved in `3bd5875af`) | Implementation | WS0 | Closed |
| G2 | Heavy-lane default too low (resolved in `6712f0408`) | Specification, implementation, tests | WS0 follow-up | Closed; a 4-lane re-measure is now worthwhile (heavy lanes alone set the wall after G5) |
| G3 | Heavy tests repeat expensive setup; a few oversized rows dominate | Tests, test support | WS3 | High; fixes 1–3 and 2b committed (`500e3bdda`); fixes 4–10 open; remaining plan in Section 12 (G3.4–G3.9) |
| G4 | Graph admission ignores physical footprint (head-of-line blocking) | Specification, implementation, tests | WS2 | G4a (host backfill) committed (`500e3bdda`); G4b (graph footprint) deferred: after G5a, host admission no longer blocks `check` |
| G5 | PostgreSQL concurrency emerges from process-slot arithmetic and double-charged lanes | Specification, machine inputs, implementation | WS2 | G5a (process slots from the OS limit) committed (`500e3bdda`); G5b (single lane) reverted |
| G6 | Exclusive PostgreSQL generation startup and residency cost | Specification, implementation | W2-4 | High; plan rewritten in Section 12 (harness-assigned identity, then a seed image) |
| G7 | Retained-secret scan costs ~120 s per `check` | Implementation, tests | WS4 | High |
| G8 | Run roots retain bulk fixture bytes on success | Specification, test support | WS4 | High (remediated, uncommitted) |
| G9 | Wait-segment and admission-record churn | Specification, implementation | WS4 | Stage 1 committed (`500e3bdda`); stage 2 (wait segments) open: G4/G5 did not reduce it, and it is now most of each run root |
| G10 | Temporary-directory leaks, including private keys | Implementation, tests, specification | WS5 | High (security hygiene; remediated, uncommitted) |
| G11 | Unbounded retained-result growth | Specification, implementation | WS5 | Medium |
| G12 | Go build-cache growth without a budget | Implementation, specification | WS5 | Low |
| G13 | `release-check` unmeasured; quiet measurement runs inline in aggregates | Measurement, specification, process | WS1, WS6 | Medium (`ci` measured: measurement is 96 s) |
| G14 | Estimates drift and diagnostics truncate | Machine inputs, implementation | WS1, WS4 | High; heavy-row estimates rebalanced (uncommitted), little wall effect |
| G15 | Fallow web dead-code policy fails `ci` on 5 unused exports | Implementation (frontend) | WS0 follow-up | Immediate for a green `ci` |
| G16 | Admission waiters re-run a full cross-process transaction every 50 ms | Implementation, specification | WS2 | High |
| G17 | Go subtest timings are computed by the runner and discarded; only top-level row durations survive | Implementation, specification | WS4 | Medium |
| G18 | Retained ATT&CK release metadata expires 2026-12-08T20:40:00Z; tests import it against the real clock | Fixtures, tests | Separate task | Urgent (outside efficiency scope) |
| G19 | `DROP DATABASE` checkpoints can exceed the fixed 35 s cleanup budget under I/O load, poisoning the unit's generation | Specification, implementation | WS2/WS3 | High; generation retirement and hardening committed (`500e3bdda`); retirement event uncommitted |
| G20 | A Git repository nested in the checkout (such as an agent worktree under `.claude/worktrees/`) breaks source enumeration and is scanned by a whole-checkout filesystem walk | Specification, implementation | WS4 | Medium (remediated, committed `500e3bdda`; `spawnSync` regression fix uncommitted) |
| G21 | `clean`/`distclean` miss declared build outputs and harness scratch (root `dist`, reference-pack builders and module staging, `apps/web/.cartulary`) | Specification, implementation | WS5 | Medium (remediated, uncommitted) |
| G22 | Readiness cache stamps vouch for external Docker images that may have been removed | Specification, implementation | WS5 | Low (remediated, uncommitted) |
| G23 | The pinned browser-trust package was removed upstream; only the local readiness cache kept bootstrap working | Implementation | WS0 | High while broken (pin updated, uncommitted) |
| G24 | Heavy units hold a heavy lane while queued in host admission behind ordinary PostgreSQL units | Specification, implementation | WS2 | High (remediated, uncommitted: about 150 s of `check` wall) |
| G25 | Unattributed I/O regression: checkpoint sync and graph wall up since the morning's passing runs | Measurement, possibly implementation | W2-1 | High (open, Section 12) |
| G26 | Packing omits PostgreSQL generation acquisition | Specification, machine inputs, implementation | W2-5 | Medium (open, Section 12) |

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

### G17 — Subtest timings are discarded

- **Evidence.** The Go runner records every test and subtest's terminal action and elapsed time from `go test -json` (the `terminal` map in [go.mjs](../../tools/harness/execution/runners/go.mjs)) but keeps only top-level row durations. Passing runs retain no Go output, including in verbose mode (`g3-verbose-probe01`). TH-HARNESS-REQ-844 forbids test titles in the secure execution index.
- **Remediation.** Under `HARNESS_DIAGNOSTICS=basic`, retain a bounded per-unit artifact of subtest durations keyed by the row's catalog identity plus an ordinal or digest of the subtest name, with names kept only where the spec permits. Decide the naming rule in the spec first.
- **Area.** Specification (REQ-844 diagnostics), implementation (Go runner), tests.
- **Rationale.** Splits within the slowest tests are currently inferred from code; measuring them would make G3 work targeted and verifiable.
- **Risk if unresolved.** Test-cost work keeps relying on estimates.
- **Validation.** A diagnostic `check` retains subtest timings for every Go unit within the bound, and the secret scan and redaction tests pass.

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
| A14 | Lifecycle estimates beside REQ-865 | `postgres_generation_acquire` joins the closed set (proposed, Section 12) | G26 | Work-graph owner input |
| A15 | §10.2 capacity table | Heavy lanes default 4, conditional on the G2 criteria (proposed, Section 12) | G2 | None |
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

## 12. Phase 2 Plan — `check` Wall Clock

**Status.** Proposed 2026-11-22, after G24. Nothing in this section is
implemented. It supersedes Section 7 for the remaining `check` wall-clock work;
Section 7 keeps the phase-1 history.

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

#### G2 (revisited) — Heavy lane count

- **Remediation.** After G25 is settled and W2-2 to W2-4 land, run three lanes
  against four (capacity override), twice each. Adopt four as the default only
  if graph wall improves beyond noise, checkpoint sync stays within 20% of the
  three-lane runs, and no cleanup or deadline failure appears (the G19 pattern).
- **Area.** Measurement; specification §10.2 default if adopted (A15).
- **Risk.** Disk saturation turns speed into flakiness, as G5b showed.
- **Validation.** The acceptance criteria above, recorded with run IDs.

#### G7 (scheduled) — Secret scan

Unchanged from Section 5. Post-graph time is now about 28 s, so the expected
saving is about 13 s of outer wall. It runs last because it is independent and
small.

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

#### W2-3 — Shared fixture support

- **Scope.** G3.7 helper, then G3.4.
- **Dependencies.** W2-0; G3.4 step 2 needs the G3.7 helper.
- **Risks.** Shared state leaking between tests; the isolation probe is
  mandatory.
- **Exit criteria.** Identical ATT&CK digests within a run; isolation probes
  pass; heavy work down by a further 80 s or more.

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

### 12.6 Amendments and Decisions

| ID | Section or requirement | Change | Gap |
| --- | --- | --- | --- |
| A3 (rewritten) | REQ-804, REQ-811; line 760 clarification | Harness-assigned generation identity; exclusive generations start from a content-addressed seed proven by a full-chain build; the seed is an immutable build artifact, not a retained allocation | G6 |
| A14 | Lifecycle estimates beside REQ-865 | `postgres_generation_acquire` joins the closed set and is reserved once per exclusive cohort | G26 |
| A15 (conditional) | §10.2 capacity table | Heavy lanes default 4, only if the G2 criteria hold | G2 |

| ID | Decision | Needed by | Recommendation |
| --- | --- | --- | --- |
| D6 | Approve rewritten A3 and the seed-image approach after the spike | W2-4 | Approve G6a now; decide G6b on the spike's numbers |
| D7 | Add `server.Options.JobPolicy` (product composition) | W2-2 | Approve; Core 01 already permits it |
| D8 | Lifecycle "replay": receipt replay or second import | W2-3 | Reference-data owner decides; keep whichever the NLSpec claim names |
| D9 | Commit plan for the uncommitted work | W2-0 | Logical commits, so W2-1 can bisect |
| D10 | Four-lane adoption criteria | W2-5 | As stated under G2 |
