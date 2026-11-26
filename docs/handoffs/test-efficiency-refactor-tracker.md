# Test Execution Efficiency Remediation Plan

**Status.** Active remediation tracker, last updated 2026-11-25. Phases 1–3 are closed. The Go-planned `check` has run at 593–626 s graph wall since 2026-11-24 (1,341 s at the 2026-11-22 baseline). The release redesign took quiet measurement out of `ci`, which now runs in 1,618–1,735 s (3,279 s in `ws1-ci01`). Phase 4 ([Section 13](#13-phase-4-plan--daily-check-and-ci)) is the active plan. It targets the two commands used to validate everyday changes, `check` and `ci`, and the stability of `ci`, which passed 3 of its last 11 runs.

**Phase 1–3 history (2026-11-23).** All phase-1 work is committed. Phase 2 ([Section 12](#12-phase-2-plan--check-wall-clock)) landed G7 and G6; G6 brought the `check` graph wall from 1,022–1,157 s to under 900 s. In G17 (phase 3), G17b and G17c landed, G17a was not adopted, and disk stall was identified as the contention the failing tests hit. G17f withdrew A17 and added A18 (WAL compression, −22% container writes), and G17g bounded test-side waits by the test deadline. G17 is closed. G27 (heavy-row test work) is done: graph wall 800–840 s, the best recorded. G28 measured the disk flushes and, with owner-approved options A (session slides persist at most once a minute on every route) and C (Go test temp work in RAM), cut heavy work 14–27% and brought the graph wall to 702–710 s, the best recorded. The run now ends with a late-starting Vitest cohort (G29). `make distclean` on 2026-11-22 deleted every retained run, so the run IDs in Section 11 before that date are a record, not retrievable evidence.

The active plan is [Section 13](#13-phase-4-plan--daily-check-and-ci); [Section 12](#12-phase-2-plan--check-wall-clock) keeps the phase-2 record. Original framing: this document authorizes no implementation by itself.
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
was refreshed on 2026-11-25 from the Go-planned `check` runs of 2026-11-24 and the
eleven `ci` runs of the release attempts (Section 13.1). Single runs vary by about
±30 s for `check` and ±60 s for `ci`; compare treatments with two repetitions each.

| Metric | Current (2026-11-25) | Target |
| --- | ---: | ---: |
| `check` graph wall, all units run | 593–626 s (seven runs, 2026-11-24); 702–710 s after G28, 887–889 s after G6 | ≤ 500 s |
| `check` graph wall after a localized product change | Same as a full run: the 124 PostgreSQL-backed units (7,095 s of 8,814 s unit time) are never cached | ≤ 300 s for the median product change (G35) |
| `ci` graph wall | 1,618–1,735 s (eleven runs, 2026-11-24 to 25); 3,279 s in `ws1-ci01` | ≤ 1,450 s after the component work; ≤ 1,200 s after G31 |
| `ci` pass rate | 3 of 11; 7 of the 8 failing runs were latent defects that only load exposed | No load-dependent failure in the phase's validation runs; every failure diagnosable from its first occurrence |
| `check` pass rate | 6 of 7; the seventh was a real privilege defect in a new migration | No load-dependent failure |
| Post-graph overhead per invocation | 11–12 s (outer minus graph, `g6-check01`, `g6-check02`; was ~60 s, ~121 s before G8) | ≤ 15 s (met) |
| Successful run root | 212 MB `check`, 567 MB `ci`; the event log is 190 MB and 345 MB of it (112,286 wait-segment pairs per `check`, G9 stage 2) | ≤ 150 MB `check` |
| Temporary directories leaked per clean run | 0 (G10; was ~20) | 0 (met) |
| Retained results | Still unbounded | Bounded by an owner-set budget |
| `release-check` wall | 477 s and 497 s in its two passing runs; release work only (release-qualification tracker, S7) | Met |

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
| G13 | `release-check` unmeasured; quiet measurement runs inline in aggregates | Measurement, specification, process | WS1, WS6 | Closed by the release redesign (2026-11-24): `ci` has no quiet units, and `release-check` (477–497 s) and `release-measure` are measured |
| G14 | Estimates drift and diagnostics truncate | Machine inputs, implementation | WS1, WS4 | Superseded by G33 (estimate drift) and G9 (event volume) |
| G15 | Fallow web dead-code policy fails `ci` on 5 unused exports | Implementation (frontend) | WS0 follow-up | Closed: `frontend_fallow_static` passes in the passing `ci` runs |
| G16 | Admission waiters re-run a full cross-process transaction every 50 ms | Implementation, specification | WS2 | Folded into G30 (one admission authority) |
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
| G27 | Heavy-lane wall sits in a few long single tests (import workflows, ATT&CK production restore, Network Flow incident transfer, the two bundle round trips) | Tests (module owners) | Phase 3 follow-up | Done: `7a62a435a`, `9f01fdc27`; changed tests −48%, graph wall −50 s (800/840 s); remaining saving lost to disk contention (G28) |
| G28 | Heavy units spent about 15% of their time waiting on their own synchronous flushes (WAL commits and product file fsyncs); stall follows host flush latency, not guest flush or byte rates | Measurement, product, specification | Phase 3 follow-up | Done: options A and C (`d8654d75a`, `42849a1e7`); graph wall 702–710 s, heavy work −14–27%, host flushes −56% |
| G29 | The 377-row Vitest cohort waits for two CPU tokens until 372–442 s and then ends every run: first-fit backfill starves multi-token claims | Scheduling (REQ-802), coupled to host admission (G4b) | Phase 3 follow-up | Folded into G30 and G31. The Vitest cohort no longer ends `check` (it starts at 333 s and ends at 516 s); the same starvation now delays the `ci` browser chain |
| G30 | Two admission layers again: PostgreSQL units hold CPU tokens and logical lanes while they wait in host admission, because each generation is charged twice | Specification, machine inputs, implementation | Phase 4, P4-5 | Done (`1862c9c20` to `2237b4519`): paired `check` 577 s and 567 s (from about 705 s); no unit waits in host admission while holding work (13.14) |
| G31 | The `ci` browser chain, its critical path, is eligible at 134 s but starts at 504–540 s | Scheduling (REQ-802, REQ-803) | Phase 4, P4-6 | Done (`36674c969`): adopted (D17); `ci` 1,445 s (from 1,688 s) with browser groups starting at 211 s instead of 551 s; paired `check` 597 s and 591 s (13.15) |
| G32 | Browser lanes are fixed chains packed from estimates that are about twice the actual time: lane loads 904–1,155 s against an even 1,031 s | Planner, scheduling | Phase 4, P4-2 | Done (`7ebf2d0af`): browser phase −102 s in one diagnostic `ci` (13.12) |
| G33 | Heavy units start in estimate order, and the estimates have drifted: the longest heavy unit starts at 260 s | Machine inputs, planner | Phase 4, P4-1 | Done (`540af9d6e`): plans use observed estimates; builds start when eligible. `check` wall unchanged, now bound by the heavy lanes and by G30/G31 (13.11) |
| G34 | Heavy placement and capacities were tuned before G28 cut host flushes by 56% | Measurement, machine inputs | Phase 4, P4-1 | Done: four heavy lanes adopted through the gate (`d77ba4a25`); placement audit inconclusive, placement unchanged; ten generations gave no gain over eight after G30 and are not adopted (13.11, 13.14) |
| G35 | PostgreSQL-backed rows are never cached, so `check` costs the same for a one-line change as for a full run | Specification, implementation | Phase 4, P4-3 | Done (`f9bbdc749`, D12): 313 of 366 `check` units cacheable; an unchanged `check` takes 90 s (13.13) |
| G36 | Every test row's cache key includes the whole harness tree, schemas, contracts, scripts and the Makefile, so tooling commits invalidate every cached row | Machine inputs, specification | Phase 4, P4-3 | Done (`f9bbdc749`): Go rows run from the plan; keys hold the runner's closure. All-invalidating commits 101 → 60 of the last 150; the rest is G40 (13.13) |
| G37 | Builds on the `ci` critical path (renderer installation 76 s, server harness 56 s) run every time | Machine inputs, implementation | Phase 4, P4-3 | Done (`f9bbdc749`): a reused renderer installation costs 31 s instead of 94–137 s, and repeated builds no longer flip the receipt (13.13) |
| G38 | A browser cohort stops at its first failure, so the rest of its tests report `infrastructure_failed` and their outcomes are lost | Implementation, possibly specification | Phase 4, P4-2 | Done (`7ebf2d0af`); not yet exercised by a browser failure (13.12) |
| G39 | Load-dependent defects surface one per `ci` run and are reproduced by hand | Tests, harness, process | Phase 4, P4-4 | Done: host pressure and stress command (`465a30afe`), wait audit (`bdd8f4e1d`); the 13.2 rows pass `REPEAT=10` under mixed load (13.10) |
| G40 | Test-row cache keys still hold broad harness inputs (harness owner models, harness test families, schemas, generated validators, task surface), because tests read repository files through paths they compute, which no closure rule can see; 60 of the last 150 commits invalidate every cached unit | Machine inputs, specification, tests | Follow-up after Phase 4 | Open: needs per-package declared run-time inputs and an enforcement mechanism (13.13) |
| G41 | The installed-renderer test read the receipt twice, so a receipt moved by a concurrent build broke it; 13.12 credited G37 with fixing this, which was wrong | Tests | Phase 4, P4-6 | Done: every renderer in the test binds the installation it opened; proven on a segment (13.15) |
| G42 | Build-artifact cache keys held per-run locations (each graph run's Go temporary directory, the per-run web source path), so every graph run missed and re-linked shared binaries | Machine inputs, specification | Phase 4, P4-6 | Done: locations leave the keyed command and input directories are keyed by content; proven on a segment (13.15) |
| G43 | Test rows launch nested graphs and real `make` runs that rebuild shared repository outputs under the Go test's toolchain environment while the outer run uses them; their failures are cut to the first 4 KiB and reclassified as product assertions | Tests, harness, specification | Phase 5 | Closed: P5-0 fenced the Go-hosted nested graphs (14.5); P5-2 moved the six nesting rows to explicit presets and refuses a graph beneath an aggregate run of the same repository (14.22) |
| G44 | Make and the planner both own ordering, freshness, environment and prerequisite skipping: 26 operations re-enter Make, 17 targets carry two recipes chosen by `GRAPH_CHILD`, five freshness mechanisms coexist, and graph-only Make edges run work twice | Machine inputs, implementation, specification | Phase 5 | Open: P5-0 removed the renderer double run (14.5); P5-1 made every test script a row and removed five duplicate operations (14.6); P5-3 remains |
| G45 | Units inherit the graph process's whole environment (the shell's variables and up to 127 Make exports), minus hand-kept strip lists; harness code reads 478 `CARTULARY_*` names with no declaration, so leaks, stale names and typos fail silently | Machine inputs, implementation, specification | Phase 5, P5-4 | Open: P5-4a registered every harness name and enforces it (14.8); P5-4b made unit environments default-deny (14.9); P5-4c slices 1 to 3 made every outside launcher explicit (14.11 to 14.13); P5-4d slices 1 to 4 gave 57 names a single owner and enforce it in Node, Go and shell, covering every multi-writer invocation-context, fixture-binding and step-runtime name (14.14 to 14.17); the P5-4d exit is met. Left for P5-3: mode switches, strip lists and the overlay; the environment tidy classified product-derived names, folded fixture identity into the browser binding, replaced the cleanup marker with an argument, deleted the constant Playwright worker names and closed two check blind spots (14.23) |
| G46 | A runtime whose cleanup fails is preserved for recovery behind a seal while it still holds resources, but startup recovery skips every sealed root on the rule that sealed roots have drained. A stranded `p58-browser` resident held one of the eight PostgreSQL generation slots in every later run | Specification, implementation | Phase 5 | Closed: `p58-browser` recovered, startup recovery decides by resource records, and the eight legacy roots were removed with owner approval (14.19) |
| G47 | Build-artifact keys come from hand-listed source roots. The server, operator and migrate binaries import 22, 22 and 19 `internal/gen` packages that none of their keys covers, and the server key also omits `db/migrations`, which the server embeds: a migration-only or generator-only change reuses a stale binary | Machine inputs, implementation | Phase 5, P5-3a | Closed: every Go build Make runs is keyed by its import closure from `go list -deps` (14.25) |

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

**Owner-approved options (committed `9f01fdc27`, 2026-11-23).** The owner
approved every remaining option. Isolated row times are from the
`g27b-slice-*` runs:

1. *ATT&CK restore probes run concurrently.* `RestoreServer` prepares every
   probe target on the test goroutine, then runs each probe on its own
   goroutine beside the final restore; each already used its own fresh
   target. `probeTB` turns `Fatal`, `FailNow` and `Skip` into goroutine-safe
   errors that end only the probe, and the test fails once all have finished.
   `attack_production_recovery`: 109 s → 70.6 s.
2. *The Network Flow incident transfer verifies one destination.* The
   operator backup/restore leg (a third server and a second full
   verification) is removed. Restored Network Flow claims and Graph readiness
   stay qualified by `TestRestoredNetworkFlowClaimBoundary_Process`. The row no
   longer needs the operator binary, and the unused
   `BackupRestoreServerWithOperator` and `pollImportJSON` are deleted.
   `network_flow_portability`: 173 s → 119.9 s.
3. *The report composition preview imports a 1 MB Entra pack* instead of the
   95 MB ATT&CK corpus; it only checks that the current pack set binds into
   the snapshot. The row drops the ATT&CK fixture root and moves from
   `postgres_heavy` to `standard`, like the Entra binding row, because its work
   is no longer dominated by PostgreSQL bulk I/O. 91 s → 22.0 s.
4. *`appsupport.ShortJobRecoveryScan`*, a `ConfigureRuntime` hook shortening
   only the recovery scan, is used by the current-workflow tests, the import
   crash-recovery restart and the incident-bundle queued-export restart.
   `unit_commit_crash_recovery`: 26 s → 14.1 s; export job row: 9–10 s →
   2.4 s.

Estimates for the seven changed rows were scaled to the new durations. Checks:
`g27b-lintgo03` (after `make format`), `g27b-srccat02`, `g27b-srcgen01`,
`g27b-generate02`, and `g27b-slice-rec01`, `-ib01`, `-rc01`, `-imp01` (all
seven rows pass).

**Final validation (paired diagnostic `check` on `9f01fdc27`).**

| Run | Result | Graph | Outer | Heavy unit work | Changed tests (heavy) | Other heavy tests | Mean `io_full` |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| `g17-final-check01` (before G27) | Pass | 857 s | 870 s | 2,383 s | 711 s | 1,496 s | 7.3% |
| `g17-final-check02` (before G27) | Pass | 882 s | 893 s | 2,496 s | 747 s | 1,564 s | 8.7% |
| `g27-final-check01` | Pass | 800 s | 811 s | 2,243 s | 378 s | 1,714 s | 9.5% |
| `g27-final-check02` | Pass | 840 s | 853 s | 2,355 s | 376 s | 1,772 s | 11.2% |

Rows under `check` load (before pair → after pair): current workflows
208/221 → 80/81 s, current cancellation 16/15 → 6/6 s, crash recovery 26/27 →
16/15 s, Network Flow transfer 241/254 → 188/179 s, export job 10/9 → 5/2 s,
ATT&CK restore 159/174 → 106/114 s, composition preview 93/90 → 37/65 s (now
off the heavy lanes). All 1,890 rows still run; the preview row now packs into
a standard cohort, so units are 364 rather than 365.

The total change is a graph wall 50 s lower on average (857/882 → 800/840 s,
the best recorded) and 140 s less heavy unit work. The changed tests saved
352 s (−48%), but every other heavy test ran 14% slower in both runs (+213 s),
and mean `io_full` rose from 7–9% to 9–11%. Container writes did not change
(46–48 GB). The same disk work in a shorter run contends harder, which is the
G28 flush-bound limit showing up: further heavy-row cuts will keep losing part
of their saving until disk stall falls.

**Remaining.** G27's identified options are done. The next wall-clock lever is
G28 (disk flushes), not more heavy-row work.

### G28 — Synchronous flushes on the heavy-lane path

Opened 2026-11-23 as a G17 follow-up, measurement first. The premise was that
disk stall follows flush count, so cutting flushes would cut stall for every
test. The measurement does not support that premise.

**Method.** Three scratch instruments, not committed. A sampler read
per-second flush and write counters for the two virtual disks: `sdd`, the Ubuntu
root, holds the harness, Go test temp directories and caches, and `sde`, Docker
Desktop's data disk, holds PostgreSQL and SeaweedFS. Every 5 s it also read
`pg_stat_io`, `pg_stat_wal` and `pg_stat_database` from each test cluster, less
the counters a seeded cluster restores from the seed build. A tracer counted
`fsync`-family syscalls of the whole `make` tree with seccomp user notification,
recording caller, unit and file. Runs: `g28-measure-check01` (sampler; pass,
graph 742 s), `g28-measure-check02` (sampler and tracer; graph 805 s; the two
report-renderer sandbox rows failed because a process tree can hold only one
seccomp listener, an instrument artifact; every other row passed), and
`g28-prof-ib01` (the eight heaviest incident-bundle rows, per-database
counters; pass).

| Run | `sdd` flushes | `sde` flushes | `sdd` written | `sde` written | Mean `io_full` |
| --- | ---: | ---: | ---: | ---: | ---: |
| `g28-measure-check01` | 160,517 | 118,625 | 23.6 GB | 70.6 GB | 8.0% |
| `g28-measure-check02` | 153,884 | 118,920 | 23.7 GB | 70.8 GB | 9.3% |

**Findings.**

1. *PostgreSQL flushes are commits.* Clusters issued 81,445 WAL fsyncs, 74,461
   of them by client backends at commit, and the checkpointer only 337 (one
   seal checkpoint per generation). PostgreSQL wrote 40.2 GB: clone copies
   12.8 GB, background writer 9.4 GB, WAL 8.9 GB, and 7.0 GB zero-filling 417
   new 16 MB WAL segments, since short-lived clusters never recycle WAL.
2. *Host flushes are mostly product storage in tests.* Of 99.6k fsyncs to
   disk-backed paths, 61.8k came from product storage under the Go test temp
   directories (reference-pack storage, backup sets, and transfer staging at
   423 fsyncs per stage) and about 36k from harness and test-service state,
   10.4k of which are execution observations written only under diagnostics.
   The harness's 175.6k admission-state fsyncs go to `/tmp`, which is tmpfs on
   this host, so they reach no disk.
3. *Stall follows flush latency, not flush or byte rates.* Over 156 ten-second
   windows, flush and write rates on both disks explain little of `io_full`
   (joint regression R² 0.15). Stall tracks mean flush latency (r 0.64 and
   0.88). Latency rises from about 2 ms at idle to 3–6 ms under any load and
   does not scale with guest writes (r ≤ 0.32), so the host sets it. This
   corrects the G27 reading: run-to-run `io_full` differences are host
   latency, not contention from compressing the same work.
4. *Each test pays for its own flushes.* Heavy units issued 48.6k WAL and
   43.8k host fsyncs. At the run's mean latencies that is about 330 s of their
   2,245 s (15%) spent waiting on their own flushes. The largest: two
   incident-bundle round-trip units (11.5k and 10.6k WAL fsyncs), the
   reference-data fresh-runtime unit (6.1k WAL, 3.6k host), and two recovery
   units (11.0k and 9.8k host fsyncs from backup storage).
5. *Server settings cannot remove commit flushes.* Isolated `pgbench`
   micro-benchmarks on `sde`: `commit_delay` of 1 ms halves flushes per commit
   at four clients but cuts throughput 36% (23% at two clients), so it is
   rejected. `wal_init_zero=off` keeps throughput and costs 1.05 instead of
   1.00 flushes per commit, removing only the zero-fill. A larger
   `shared_buffers` would cut background-writer bytes, which do not set
   latency. Synchronous commit, `fsync` and checkpoint behavior are fixed by
   the NLSpec.
6. *Commit sources in the incident-bundle rows.* `g28-prof-ib01` saw 62k
   commits. The ATT&CK import commits about 20 rows per transaction (60k rows
   in 2.7–3.5k commits). Several databases commit 2–7k transactions with few
   inserts, mostly updates. These are not session slides, which Core 04 lets
   routes throttle to one per minute and the product does, and not the
   import. Their source (job leases and progress, or 25 ms test polling) is
   not yet attributed.

**Options (owner decisions).**

- *A. Cut per-request and job commits in the heaviest rows.* Attribute them
  with per-table counters, then change tests or product. Halving the top
  units' commits would save roughly 50–80 s of heavy work under `check`.
- *B. Batch fsyncs in product backup and transfer staging.* A product
  durability design change (recovery owner).
- *C. Put Go test temp directories on tmpfs when capacity allows.* This
  removes about 62k disk flushes, but conflicts with TH-HARNESS-AC-092 (cache
  roots resolve outside `/tmp`) and moves test data into memory.
- *D. `wal_init_zero=off`.* Needs a spec sentence like A18; small and
  uncertain gain.

**Owner decision (2026-11-23): A and C.** B and D are not pursued.

**A: profile and change.** A temporary `pg_stat_statements` preload (removed
after profiling) ranked the statements that write WAL in the heaviest
incident-bundle and reference-data rows (`g28-pgss-ib01`, `g28-pgss-rd01`).
The import itself commits in batches. The avoidable commits were session
slides: 3,457 `UPDATE user_sessions` in the incident-bundle rows and 1,634 in
the reference-data rows, about one per job poll. Core 04 REQ-04-011 lets a
route skip persisting a slide while the stored activity is under a minute
old. Only the workbook routes did; 17 other routes slid on every request (the
2026-11-19 cadence review recorded the difference as conformant). Now
`httpauth.SlideSessionIfNeeded` persists at most once per
`SessionSlideWriteInterval` on every route, and the separate workbook-only
helper is gone. Four tests that expected a write right after login now
start from activity one interval old: the httpauth cadence test, extension
discovery's service-backed slide test, and the auth credential-state and
Network Flow session-maintenance unit tests. Staged-object state transitions
(about 12k single-row writes in the incident-bundle rows) are a durable
product state machine and are unchanged. Committed `d8654d75a`, `42849a1e7`.

**C: Go test temp work in RAM (amendment A20).** A graph invocation creates
its private Go temporary-work directory beneath
`/dev/shm/cartulary-go-tmp-<uid>` when `/dev/shm` is tmpfs with at least
8 GiB available and the root is a real directory owned by the user with mode
`0700`; otherwise it uses `GO_TMP_DIR` as before. Test `TempDir` trees and
build work then live in memory, and product `fsync` calls in tests wait for
no disk flush. `GO_TMP_DIR` resolution, `doctor` and the machine-cache
defaults are unchanged. The NLSpec's invocation-scoped paragraph and
TH-HARNESS-AC-092 record the rule; `TestRuntimeGoTemporaryWork` covers
eligible, disk-backed, under-capacity, unavailable, permissive,
symbolic-link, non-directory and foreign-owned roots. Committed `d8654d75a`.

**Checks.** `g28a-format01`, `g28a-lintgo01`, `g28a-slice-cs01`
(`harness.command_surface.unit.native_runtime`), `g28a-slice-http01`,
`g28a-slice-srv01`, `g28a-slice-auth01`, `g28a-mdlint01`, `g28a-srcgen01`:
all pass. The first confirming `check`, `g28a-check01`, failed three rows: the
two unit tests above (fixed, `g28a-slice-auth02` and `g28a-slice-nf01` pass)
and `platform.jobs.integration.supervised_runner`, whose cleanup closes the
runner within a fixed 1 s; it passed on rerun (`g28a-slice-jobs01`) and
touches neither sessions nor temp directories.

**Final validation (paired diagnostic `check` on `42849a1e7`, flush
sampler attached).**

| Run | Rows | Graph | Heavy unit work | Host-disk flushes | Host-disk writes | Docker-disk flushes | WAL fsyncs | Mean `io_full` | `/dev/shm` peak |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| `g27-final-check01`, `-02` (before G28) | 1,890, 1,890 | 800 s, 840 s | 2,243 s, 2,355 s | — | — | — | — | 9.5%, 11.2% | — |
| `g28-measure-check01` (before) | 1,890 | 742 s | 2,064 s | 160,517 | 23.6 GB | 118,625 | 81,445 (`-02`) | 8.0% | — |
| `g28a-final-check01` | 1,890 | 710 s | 1,768 s | 69,432 | 5.3 GB | 96,617 | 59,465 | 5.0% | 2.2 GB |
| `g28a-final-check02` | 1,889 | 702 s | 1,725 s | 70,091 | 5.5 GB | 96,520 | 58,837 | 5.3% | 2.2 GB |

Graph wall 702–710 s is the best recorded (800/840 s before G27's pair,
742 s for the unchanged measurement run). Heavy unit work fell 14–27%, host
flushes 56%, host writes 77% and WAL fsyncs 28%; `io_full` dropped to 5%.
The largest per-row cuts were file-heavy recovery and reference-data rows:
capture-target admission 240 s → 110 s, absent optional packs 143 s → 97 s,
canonical operator process 94 s → 55 s (`g28a-check01` against
`g28-measure-check01`). Memory stayed ample: `/dev/shm` peaked at 2.2 GB and
available memory never fell below 12.1 GB (9.4 GB in `g28a-check01`).

Both runs exit non-zero, neither from this change:

- *`harness.go_vulncheck` (both runs, and `g28a-check01`).* The Go
  vulnerability database added GO-2026-6974
  (`go.opentelemetry.io/otel` AttributeValueLengthLimit bypass, fixed in
  v1.47.0) at 2026-11-23 20:05 UTC; the symbol is reachable, so the scan
  blocks every `check` until the dependency is upgraded. Tracked as a
  separate task.
- *`module.imports.integration.network_flow_public_default_limits_commit_the_ac_fc9396304e`
  (`-02` only).* The known row-limit test (G17): two source passes share the
  30 s participant budget the test asserts. The second pass took 22.0 s
  against 6.2 s for the first while Docker-disk flush latency rose from 2 ms
  to 6–9 ms; it passed in the other two runs.

**What bounds the wall now.** Heavy lanes finish at 637–646 s. The run ends
when the 377-row Vitest cohort `vitest_cohort:16a12aa1…` does; it is eligible
at 1 s but waits for two CPU tokens until 372–442 s (617 waits in
`g28a-check01`) while one-token units keep taking them, then runs 300–330 s.
Starting it early would put the wall near 650 s. That is G29.

### G29 — Long Vitest cohort starts late

Opened 2026-11-23 from the G28 validation. The 377-row frontend Vitest cohort
`vitest_cohort:16a12aa1…` (claims 2 CPU, 4 process slots) takes 300–450 s.
Across runs it started at 6, 161, 213, 372, 380 and 442 s although it is
eligible at 1 s; its rows carry no duration estimates, and its wait reason is
`cpu` (two-token deficit) behind one-token units. Since G28 it ends every run.
Candidates: estimates for its rows so ranking places it early, or splitting
the cohort.

**Cause.** Estimates are not the problem: at 378 s the cohort is already the
highest-ranked unit in the graph. TH-HARNESS-REQ-802 admits the highest-ranked
fitting unit and backfills with any lower-ranked fitting unit, without
reservation. A two-token unit is skipped whenever one token is free, and in
`g28a-final-check02` 268 lower-ranked one-token units each took a token at
exactly that moment. Aging raises rank but cannot admit a unit that never
fits, so REQ-802's "backfill cannot starve work" does not hold for multi-token
claims. In runs where it started at 6 s, only 8 such backfills occurred.

**Attempt (reverted).** The logical scheduler adopted REQ-803's
deficit-disjoint backfill: lower-ranked fitting work may not draw down what the
highest-ranked capacity-lacking unit lacks (`70565c792`), then a narrower form
that protects only resources it currently lacks (`6a006b67a`). Scheduler tests
passed, and the cohorts started at 2–4 s. Both validation runs were about
200 s slower:

| Run | Rule | Graph | Heavy-lane idle | Go `host_admission` | Rows |
| --- | --- | ---: | ---: | ---: | --- |
| `g28a-final-check02` | First-fit (before) | 702 s | 186 lane-s | 4,186 unit-s | 1,889 |
| `g29-check01` | Deficit-disjoint | 899 s | 1,113 lane-s | — | 1,889 |
| `g29-check02` | Current deficits only | 903 s | 999 lane-s | 9,298 unit-s | 1,889 |

Held-back work (`rank`) accounts for almost none of the idle heavy-lane time.
Heavy units instead waited on `postgres` and `object_store` lanes held by
ordinary units, and Go units spent twice as long in physical host admission
while holding their logical tokens and lanes. Changing the logical order shifts
which units queue at host admission, and the head-of-line coupling between the
logical and physical schedulers (G4b) dominates. Reverted in `b5d0935a8`; the
spec and engine are back to `578259802`. G29 stays open: a fix must address
the G4b coupling first, or start the cohort by about 300 s without
reordering the rest. The two failing rows were unrelated:
`module.networkflow.integration.saved_graph_lifecycle_v2` (`g29-check01`) is a
live-runner race in the test, fixed in `408e11948`, and
`module.extensions.integration.managed_service_assertions_for_state_admission_m_55b90ea265`
(`g29-check02`) is not yet investigated.

**Saved-graph test race (fixed `408e11948`).**
`TestNetworkFlowSavedGraphLifecycleRoutes_Integration` restores against a live
server whose Jobs runner keeps its 5 s recovery scan. Restore reconciliation
leaves the Network Flow job `running` with no attempt, a recovery-scan
candidate, so a scan landing between the restore commit and the test's read
re-claimed it under a fresh attempt (`8b2a7936…`, not the seeded `…9105`); one
failure in 33 retained runs. The test now asserts that the restored attempt was
released, and seeds the job in one transaction through the new
`JobsCapability.CreateTx`, so the runner never sees it queued. Checks:
`race-lintgo01`, `race-slice-nf01`, `race-slice-cs01`; all pass.

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
| A20 | §13 invocation-scoped temporary work; TH-HARNESS-AC-092 | A graph invocation places its private Go temporary-work directory beneath `/dev/shm/cartulary-go-tmp-<uid>` when `/dev/shm` is tmpfs with at least 8 GiB available and the root is a private directory owned by the user, otherwise beneath `GO_TMP_DIR` (owner decision 2026-11-23, committed `d8654d75a`) | G28 | None |
| A21 | REQ-806 quiet admission; Section 10 scheduler contract | A unit holding `host_activity` exclusively becomes ready at the first scheduling decision at which its dependencies are satisfied and no shared `host_activity` unit is ready, and stays ready; ordinary work runs before quiet work and released quiet units run as one block. The no-backfill barrier is unchanged (owner decision 2026-11-24; work-graph overhaul tracker, "Wall-clock: quiet work last") | Work-graph overhaul | None |
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

Phase 4 decisions (D11–D17) are listed in [Section 13.8](#138-decisions-and-amendments).

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
| 2026-11-23 | G27 | Four owner-approved options: concurrent restore probes, one transfer destination, Entra setup pack, shared recovery-scan hook; targeted checks and changed rows | `g27b-lintgo03`, `g27b-srccat02`, `g27b-srcgen01`, `g27b-generate02`, `g27b-slice-rec01`, `g27b-slice-ib01`, `g27b-slice-rc01`, `g27b-slice-imp01` | All pass; seven rows 31–76% faster in isolation. Committed `9f01fdc27` |
| 2026-11-23 | G27 | Final validation: paired diagnostic `check` on `9f01fdc27` | `g27-final-check01`, `g27-final-check02` | **Pass twice.** Graph 800 s and 840 s (from 857 s and 882 s); heavy unit work 2,243 s and 2,355 s; changed tests −352 s, other heavy tests +213 s on higher disk stall (`io_full` 9.5% and 11.2%) |
| 2026-11-23 | G28 | Diagnostic `check` with a per-second disk-flush and PostgreSQL I/O sampler | `g28-measure-check01` | Pass; graph 742 s. Host disk 160,517 flushes, Docker disk 118,625; PostgreSQL sampling missed seeded clusters (image-ID filter), fixed for the next run |
| 2026-11-23 | G28 | Diagnostic `check` with the sampler and a seccomp `fsync` tracer | `g28-measure-check02` | Graph 805 s; two report-renderer sandbox rows failed on the tracer's seccomp listener (instrument artifact), all other rows passed. 81,445 WAL fsyncs, 337 checkpointer fsyncs; 99.6k host fsyncs to disk; heavy units about 15% in their own flush waits |
| 2026-11-23 | G28 | `pgbench` micro-benchmarks for `commit_delay` and `wal_init_zero`; per-database commit profile of the eight heaviest incident-bundle rows | `g28bench` (scratch), `g28-prof-ib01` | `commit_delay` rejected (−23–36% throughput); `wal_init_zero=off` marginal; 62k commits, import batched about 20 rows per commit, remaining commit source unattributed |
| 2026-11-23 | G28 | Statement profile with a temporary `pg_stat_statements` preload (removed after) | `g28-pgss-ib01`, `g28-pgss-rd01` | Pass. Avoidable commits are per-request session slides (3,457 and 1,634); the import is batched |
| 2026-11-23 | G28 | Options A and C: throttled slides on every route; Go test temp work in `/dev/shm` (A20); targeted checks | `g28a-format01`, `g28a-lintgo01`, `g28a-slice-cs01`, `g28a-slice-http01`, `g28a-slice-srv01`, `g28a-slice-auth01`, `g28a-mdlint01`, `g28a-srcgen01` | All pass. Committed `d8654d75a` |
| 2026-11-23 | G28 | Confirming diagnostic `check` | `g28a-check01` | Graph 750 s, heavy work 1,732 s. Failed: two unit tests expecting an immediate slide (fixed, `42849a1e7`; `g28a-slice-auth02`, `g28a-slice-nf01` pass), the jobs runner 1 s close budget (passed on rerun, `g28a-slice-jobs01`), and `harness.go_vulncheck` on a newly published otel advisory |
| 2026-11-23 | G28 | Final validation: paired diagnostic `check` on `42849a1e7` | `g28a-final-check01`, `g28a-final-check02` | Graph 710 s and 702 s (best recorded); heavy work 1,768 s and 1,725 s; host flushes −56%, WAL fsyncs −28%. Rows 1,890 and 1,889 passed; both runs fail `harness.go_vulncheck` (GO-2026-6974, external), and `-02` failed the load-sensitive Network Flow row-limit test |
| 2026-11-23 | OTel | GO-2026-6974 advisory: OpenTelemetry Go v1.47.0 (log v1, OTLP log exporters v0.23.0), deprecated log/global calls replaced, snapshot and identity tests updated | `otel-bootstrap01`, `otel-vuln02`, `otel-lintgo02`, `otel-telemetry02`, `otel-conformance02`, `otel-slice-srv01`, `otel-slice-nf01`, `otel-srcgen01` | All pass; `govulncheck` has no blocking finding. Committed `578259802` |
| 2026-11-23 | G29 | Deficit-disjoint logical backfill, then current-deficit form; scheduler rows and one diagnostic `check` each | `g29-slice-sched01`, `g29-check01`, `g29-slice-sched02`, `g29-check02` | Cohort starts at 2–4 s, but graph 899 s and 903 s (from 702–710 s): heavy lanes idle behind host-admission waits. Reverted `b5d0935a8` |
| 2026-11-23 | G29 | Saved-graph lifecycle test race: assert released restored attempt; transactional job seeding | `race-lintgo01`, `race-slice-nf01`, `race-slice-cs01` | All pass. Committed `408e11948` |
| 2026-11-24 | Go planner (work-graph tracker) | `check` after the Go planner became the default (slices 3b–4d, Stage 2) | `20261124T022650Z-p16111`, `20261124T161533Z-p41508`, `20261124T171701Z-p83767`, `20261124T180212Z-p90231`, `20261124T185203Z-p57224`, `20261124T193704Z-p20922`, `20261124T212414Z-p1022` | Graph 856 s, then 593–626 s. Six passed; `-p1022` failed one row on a real privilege defect in a new migration. The drop from about 700 s is not attributed to a single change |
| 2026-11-24 to 25 | Release redesign (release-qualification tracker) | Eleven `ci` runs during the release attempts | `20261124T200215Z-p34171.ci` through `20261125T160715Z-p42612.ci` | Graph 1,618–1,735 s; 3 passed. Failures in Section 13.2 |
| 2026-11-25 | Phase 4 analysis | Unit-event profiles of `check` `-p20922` and `ci` `-p42612`: phase occupancy, lane loads, estimate ratios, cache facts; commit-history simulation of cache invalidation | `20261124T193704Z-p20922`, `20261125T160715Z-p42612.ci`, the plans of the five 2026-11-24 `check` runs | Findings G30–G39 and Section 13. No code change |
| 2026-11-25 | P4-0 | Reference runs at `c9502cb39`, diagnostics on, cache off | `p40-ref-check01`, `p40-ref-ci01` | **Pass both.** `check` 368/368, graph 691 s; `ci` 502/502, graph 1,791 s. Detail in 13.9 |
| 2026-11-25 | P4-4 / G39 | Failure-scoped host pressure and `make test-stress` (REQ-870, REQ-871) | Owner rows `harness.command_surface` (native runtime, stress command) | Pass. Committed `465a30afe` |
| 2026-11-25 | P4-4 / G39 | Stability gate: each 13.2 row `REPEAT=10` with `LOAD=mixed` at `465a30afe` | `p44-gate-{ext,rel,imp,inc,wb,rep}`, repetitions `-r01` to `-r10` | **60 of 60 passed.** Detail in 13.10 |
| 2026-11-25 | P4-1 / G33 | Observed estimates and `make estimate-refresh`; folded `p40-ref-check01` and `p40-ref-ci01` (2,604 rows, 80 operations) | `p41-cs`, `p41-ea`, `p41-src`, `p41-lintgo`, `p41-cs4`; diagnostic `check` `p41-g33-check01` | Rows and checks pass. `check` 719 s (reference 691 s). Committed `540af9d6e`. Detail in 13.11 |
| 2026-11-25 | P4-1 / G34 | Four heavy lanes by override, paired diagnostic `check`; G28 sampler on the first | `p41-heavy4-check01`, `p41-heavy4-check02` | **Pass both:** 706 s and 704 s. Adopted as the default (`d77ba4a25`; `p41-heavy4-cs`, `p41-heavy4-src` pass) |
| 2026-11-25 | P4-2 / G32, G38 | Browser lanes as a runtime capacity; test-owned cohorts run to completion | `p42-cs2`, `p42-hb`, `p42-src`, `p42-lintgo`, `p42-md`; diagnostic `ci` `p42-ci01` | Targeted rows pass. `ci` 2,647 of 2,648 rows, graph 1,688 s (reference 1,791 s); the failure is the renderer receipt race (13.12). Committed `7ebf2d0af` |
| 2026-11-25 | P4-3 / G35, G36, G37 | Go rows run from the plan; runner-closure and named-input keys; cacheable PostgreSQL rows; renderer reuse | `p43-src`, `p43-lintgo`, `p43-lintscripts`, `p43-md`, `p43-cs`, `p43-ea`; `check` with the cache on, `p43-check01` (all misses) and `p43-check02` (repeat) | All pass. `check` 663 s with every row executed, then 90 s with 313 of 313 eligible units reused. Committed `f9bbdc749` |
| 2026-11-25 to 26 | P4-5 / G30 | Host ledger as the admission authority; PostgreSQL work and generations stated once; four defects found and fixed (13.14) | Targeted: `p45-cs`, `p45-ea`, `p45-hb`, `p45-tc`, `p45-rel`, `p45-src2`, `p45-fix*`, `p45-ci-*`, `p45-f-*`, `p45-g-*`. Diagnostic `check`: `p45-check01` to `-05` (graph aborted, defects), `p45-check06` and `-07` (pass); `p45-gen10-check01`, `-02`; `ci` `p45-ci01`, `-02` (failed on the defects) | Paired `check` pass: 577 s and 567 s. Ten generations: 573 s and 558 s, not adopted. Committed `1862c9c20` to `2237b4519` |
| 2026-11-26 | P4-6 / G31, G41, G42 | Critical-path reservation; renderer test binds one installation; location-free build keys | Gate: `p46-check01`, `p46-check02`, `p46-ci01`, `p46-ci02` (stopped); segments `seg-ontest-before`, `seg-ontest-after`, `seg-cache01`, `seg-cache02`; `fx-lintgo`, `fx-lintscripts`, `fx-md`, `fx-src` | `check` pair pass (597 s, 591 s). `ci01` 2,648 of 2,649 rows, graph 1,445 s; the failure is G41. The owner stopped `ci02` after the same failure. Segments prove both fixes; targeted checks pass (13.15) |
| 2026-11-26 | P5-0 / G43, G44 | Nested launches undo the unit overlay; graph-only renderer prerequisite removed; nested failures keep their tail and class; Go build failures keep compiler output | Segments `p50-nested-before`, `p50-nested-after2`, `p50-nested-warm`, `p50-nested-steady`; rows `p50-units`, `p50-nest2`; `p50-lintgo`, `p50-md`, `p50-src` | HEAD's nested graph moved the receipt, replaced the shared server-harness binary and missed three build caches; fenced, it changes nothing in steady state. All rows and checks pass (14.5) |
| 2026-11-26 | P5-1 / G44 | Every test script is a shell row; catalog enforces it; five duplicate operations removed | Catalog probes; plans of `check`, `ci` and `harness-contract SUITE=regression`; rows `p51-cs`, `p51-ea`, `p51-ea2`, `p51-cs2`, `p51-ea3`; `p51-lintscripts`, `p51-md`, `p51-src2` | Catalog 2,774 rows (from 2,754). `check` and `ci` plan the same units (366, 501); each gains only `public_wrapper_smoke` as a row. New rows run alone and pass (14.6) |
| 2026-11-26 | P5-4a / G45 | Environment registry of every harness `CARTULARY_*` name and its check in `check`, `ci` and `source-check`; one dead mode switch deleted | Registry probes; plans of `check` and `ci`; rows `p54-cs`, `p54-cs2`, `p54-ea`, `p54-hb`; `p54-src2`, `p54-surface`, `p54-lintscripts` | 480 names registered and 8 declared inputs cover all 488 harness names. `check` and `ci` gain only the new check (and its test row in `ci`). All rows and checks pass (14.8) |
| 2026-11-26 | P5-4b / G45 | Default-deny unit environments derived from the registry; dead strip entries removed | Segments `p57-units`, `p57-planner`, `p57-report`, `p57-pg`, `p57-vitest`, `p57-browser`, `p57-nested`, `p57-strip`; `p57-src2`, `p56-lintgo3`, `p57-lintscripts`; diagnostic `check` `p57-check01` | Every segment passes. A captured unit environment holds 100 names, not 164; 64 shell, agent and desktop variables no longer reach units. `check` passes, 1,894 rows, graph 576 s (14.9) |
| 2026-11-26 | P5-4c slice 1 / G45 | Explicit child environments for the services session, frontend publication and browser acquisition; launcher scope enforced by the registry check | Segments `p58-*` (two failures found and fixed), `p59-hb`, `p59-browser`, `p59-ea`, `p59-cs`; `p59-src`, `p59-lintscripts` | Every segment passes after the fixes. The registry check holds the three launchers to the rule with no findings (14.11) |
| 2026-11-26 | P5-4c slice 2 / G45 | Explicit child environments for UI-review preparation and the native review client | Direct preparation tests (52); segment `p60-review` (all eight UI-review rows, including both seeded rows); `p60-cs`, `p60-src`, `p60-lintscripts` | All pass. Five launcher files are now under the registry's `explicit_launchers` scope with no findings (14.12) |
| 2026-11-26 | P5-4c slice 3 / G45 | Go test support starts product processes with `childenv.Build`; the check covers Go launchers | Segments `p61-cs`, `p61-app-server` (18 rows), `p61-module-recovery` (5 rows, including the restore browser rows), `p61-module-extensions`, `p61-module-networkflow`; `p61-lintgo2`, `p61-lintscripts`, `p61-src2` | All pass after removing a helper the change left unused (14.13) |
| 2026-11-26 | P5-4d slice 1 / G45 | Registry `owners` and a write check; owners for the suite-runtime handle, suite attachment and runtime binary names; the review server's copied binary defaults removed | `p62b-generate`, `p62b-lintgo`; segments `p62b-cs`, `p62b-ea`, `p62b-hb-units`, `p62b-hb` (including the seeded default review), `p62b-browser`; `p62b-lintscripts`, `p62b-src` | All pass. The first generate found the new Go test uncataloged; it joined the `suiteservices` helper row (14.14) |
| 2026-11-26 | P5-4d slice 2 / G45 | `run_context` owner (Node `run-context.mjs`, Go `runcontext`) for the invocation context and results root; owners for the runner script and failure context; 70 write sites converted; a stale guide assertion from P5-1 fixed | `p63-generate`, `p63-lintgo`; segments `p63-cs` (14 rows), `p63-ea` (7), `p63-rel` (3), `p63-hb-native` (15), `p63-hb`, `p63-rep`, `p63-browser`; reruns `p63r-tc`, `p63r-hb`; `p63-lintscripts`, `p63r-src` | All pass. `routing_contract` had failed since P5-1 on a stale assertion; `native_fixture_lifecycle` failed once on a tracked-file edit made during the run (14.15) |
| 2026-11-26 | P5-4d slice 3 / G45 | `browser_binding` owner (Node `browser-binding.mjs`, Go `browserbinding`); resource dependencies join `run_context`, the PostgreSQL allocation joins `suite_services` | Planner equivalence: `ci` and `release-measure` plans with and without the change (`p64-plan-*`); `p64-generate`, `p64-lintgo`; segments `p64-cs`, `p64-ea`, `p64-hb-native` (16 rows), `p64-hb`, `p64-obj`, `p64-pg`, `p64-browser`; `p64-lintscripts`, `p64-src` | All pass; every unit digest is unchanged. Seven of the 20 names are product secrets that any launcher may set, and five had only one Node or Go writer (14.16) |
| 2026-11-26 | P5-4d slice 4 / G45 | The registry check reads shell and Make recipes; `step_runtime` owner with one summary writer replacing four copies; `--target` and `--timing-bucket` step-runner options; the step artifact directory joins `run_context` | `p65-generate`, `p65-lintgo`; segments `p65-cs` (8 rows), `p65-ea` (3 step accounting rows), `p65-hb`, `p65-browser`; `p65-lintscripts`, `p65-src` | All pass. Shell checking surfaced nine writes of names already owned; entry points and the browser stack joined owner paths, and three launchers stopped writing the target (14.17) |
| 2026-11-26 | P5-4 close-out | One `check` after P5-4c and P5-4d | `p66-check01`; after the fix, `p66-lintshell` and `p66-cache` | All 1,897 rows pass. One operation failed: ShellCheck SC2034 on a variable that slice 4 left unused in `build-go-artifact.sh`. `lint CHECK=scripts` does not run ShellCheck; `lint CHECK=shell` does. The script now requires the target without assigning it, and the shell lint and build-artifact rows pass. The graph took 617 s (577 s at P5-4b), while unit work fell from 7,559 to 7,438 s. The same units finish last; the renderer ran 22 s longer. It is 20 s past the earlier 576 to 597 s spread, so one sample cannot separate noise from regression (14.17; settled in 14.18) |
| 2026-11-26 | P5-4 second check / G46 | A second `check` to settle the timing | `p67-check02`; `p67-ac604-1` to `-3` | 1,896 of 1,897 rows pass. `app.server.process.ordinary_shutdown_ac604` failed once (WebSocket close status -1, not 1001) and passed three isolated reruns; it has passed 75 of 76 runs. The graph took 674 s. Not a P5-4 regression: since `p58-browser`, whose cleanup failed, a stranded resident has held one of the eight PostgreSQL generation slots, so PostgreSQL units ran 7 at a time instead of 8 in `p66` and `p67`. This run also rebuilt the report renderer (153 s) because its cache key includes `build-go-artifact.sh`. That delayed the server-harness build to 167 s, so the server-process unit started last and ran to the end of the run (14.18) |
| 2026-11-26 | G46 repair | Recover `p58-browser`; startup recovery decides by resource records, not the seal (owner-approved spec amendment) | `p68-recover`; `p68-lintgo`, `p68-session`, `p68-fixtures`, `p68-md` | Pass. The ledger holds no residents, so all eight PostgreSQL generation slots are free. Every top-level graph now reports eight legacy October roots as `[RECOVERY-FAIL]`: their records predate the current ledger and boot, and they hold nothing physically. Removing them needs owner approval (14.19) |
| 2026-11-26 | Never-run rows | Promote the four browser accessibility self-tests from `explicit_only` to `full`; keep the eight on-demand rows (owner decision) | `p69-plan-*` (every graph target and suite); `p69-a11y`; `p70-generate`, `p70-catalog`, `p70-generated`, `p70-routing`, `p70-lintscripts`, `p70-plan-ci`, `p70-plan-check` | Pass. 12 of 2,778 rows were never planned by any target or suite. `ci` now plans the four promoted rows (2,658 rows) in its existing 506 units; `check` is unchanged (14.20) |
| 2026-11-26 | Child CPU budget | The registry's `budget` names (`GOMAXPROCS`) pass to every child that a launcher builds, and never come from the graph's environment; `processtest` helper launches pass their mode explicitly | `p71-generate`, `p71-lintgo`, `p71-lintscripts`, `p71-cs`, `p71-server`, `p71-surface`, `p71-generated`; `p72-lintgo`, `p72-cs`, `p72-server`, `p72-nf`, `p72-browser`, `p72-md` | Pass. `p71-cs` failed the `ci`-only `processtest` row, a P5-4c regression: its helper mode reached the child only by inheritance. It passes after the repair, as do the other seven `ci`-only rows that start processes through `childenv` (14.21) |
| 2026-11-26 | P5-2 / G43 | The six rows that start nested graphs are `explicit_only` in five explicit `harness-contract` presets; a graph refuses to start beneath an aggregate run of the same repository (D20) | `p73-generate`, `p73-plan-*`, `p73b-plan-*`; `p74-generate`, `p74-lintgo`, `p74-catalog`, `p74-generated`, `p74-runtime`, `p74-refused`, `p74-allowed`, `p74-nested`, `p74-realout`, `p74-lifetime`, `p74-md` | Pass. `ci` plans 2,652 rows in 500 units (2,658 in 506); `check` is unchanged; the presets keep their rows. A stand-in `ci` process refuses a real nested `test-slice`, and a stand-in `harness-contract` allows it (14.22) |
| 2026-11-26 | Environment tidy / G45 | `product_derived` role for 14 names; fixture identity in `browser_binding`; `cleanup-web-e2e --process-cleanup` replaces its marker; Playwright worker names deleted; the check counts quoted shell assignments and loop-variable writes | `p75-generate`, `p75-format`, `p75-lint*`, `p75-surface`, `p75-catalog`, `p75-identity`, `p75-generated`, `p75-cs`, `p75-hb`, `p75-incidents`, `p75-md`; `p75-plan-measure-old` and `-new` | Pass. The registry holds 477 names (480) and its check has no findings; the new loop rule's one finding is repaired. `release-measure` plans identically before and after (14.23) |
| 2026-11-26 | P5-3a / G47 | Go build keys are each binary's import closure; receipts are built for the staged binary; the 17 graph-child recipes, 11 unused Makefile variables, one runtime macro and four dead switches removed | `p77-*` (segments, real builds, pre-graph timing), `p78-lintshell`, `p78-lintscripts`, `p78-cache`, `p79-check01` | Pass. Segments found ShellCheck SC2016 on two intended strings and a test cleanup defeated by read-only module files; both repaired. `p79-check01` passed all 1,897 rows in a 543 s graph (14.25) |
| 2026-11-25 | P4-4 / G39 | Wait audit: 31 whole-test hang guards derive from the test deadline (`testdeadline`) | `p44-port-*` (16 owner slices, every row holding a converted guard) | Pass. Committed `bdd8f4e1d` |

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

## 13. Phase 4 Plan — Daily `check` and `ci`

**Status.** Approved and in execution since 2026-11-25. The owner approved
D11–D16 as recommended on 2026-11-25; D17 is decided by the 13.7 gate after
P4-5. Progress is recorded per workstream in 13.5 and Section 11. The owner set the priority. Developers use
`check` and `ci` to validate everyday changes, so their wall time and
stability come before `release-measure`, which runs rarely. Each component is
improved and measured first. Changes that make phases run concurrently (G31)
come last, because overlap can introduce instability that targeted tests do
not show.

### 13.1 Current State

**Evidence.** The profiles come from `check` `20261124T193704Z-p20922`
(605 s, pass) and `ci` `20261125T160715Z-p42612.ci` (1,660 s, pass). Neither
ran with diagnostics, so they are read from the phase events in
`unit-events.ndjson`; Section 11 lists the runs behind the ranges. The host is
as in Section 1: 24 logical CPUs and 19 CPU tokens.

`check` units by class. Unit time runs from `started` to `completed`, so it
includes host admission.

| Class | Units | Unit time | Runner | Host admission | Cache |
| --- | ---: | ---: | ---: | ---: | --- |
| PostgreSQL-backed Go | 109 | 5,495 s | 1,819 s | 3,380 s | Never (by spec) |
| Heavy-lane Go | 15 | 1,599 s | 1,509 s | 30 s | Never (by spec) |
| Plain Go | 186 | 837 s | 801 s | 1 s | 182 eligible |
| Vitest | 13 | 463 s | 462 s | 0 s | Eligible |
| Operations | 43 | 420 s | 415 s | 0 s | 2 eligible; builds are not |

`ci` runs 502 units: the same classes, a few more Go units and operations,
and 29 browser groups with 4,125 s of unit time.

**What bounds `check` (605 s).**

- *0–430 s: PostgreSQL generations.* In every 30 s sample from 30 s to 390 s,
  7–9 PostgreSQL units waited in host admission while 6–8 ran tests. The
  logical layer admits 16 PostgreSQL units: the capacity is 16 and each unit
  claims one lane. Host admission charges each unit's own lane and again its
  resident generation's lane (`physical_resource_claims.postgres` in the
  topology manifest), so only 8 run. Units that use the object store are
  charged the same way: 8 lanes, 2 per unit, so 4 at a time. Units waiting in
  host admission keep their CPU tokens. All 19 tokens were claimed for 435 s,
  and plain Go units waited 30,091 unit-seconds for one.
- *0–604 s: the three heavy lanes.* There are 15 heavy units with 1,599 s of
  work. An even split is 533 s, longest first on the actual durations is
  544 s, and the heavy lanes finished at 604 s. Heavy units start in estimate
  order, but the estimates range from 9% under to 2.5 times the actual time.
  A unit estimated at 221 s that ran 89 s started first. The longest unit (`network_flow_portability`, 193 s,
  estimated 176 s) started at 260 s.
- *After 450 s,* only heavy units run, three and then one.

**What bounds `ci` (1,660 s).**

- *Builds.* Renderer installation (76 s) and then the server harness (56 s)
  run every time (cache policy `none`). The browser chain becomes eligible at
  134 s.
- *Starved critical path.* The four browser lane heads started at
  504–540 s. Each browser group claims 2 CPU tokens, 2 PostgreSQL lanes and an
  object-store lane at once. First-fit backfill gives every freed single token
  to a ready Go unit, the G29 starvation. The spec already forbids this:
  "Noncritical policy work MUST NOT starve the functional critical path."
- *Fixed browser lanes.* From 504 s to 1,659 s the browser groups run on four
  lanes fixed at plan time. The lanes were busy 969, 1,096, 1,155 and 904 s,
  against an even 1,031 s. The planner packs longest first by estimate, and
  the estimates are about twice the actual time (actual/estimate 0.26–0.82,
  median 0.5). Lane 4 sat idle for the last 244 s.
- *Idle host.* From 690 s, six or fewer units run; from 1,110 s, four or
  fewer. From 690 s, at most five PostgreSQL-backed units run, browser
  groups included.

**Cache in daily development.** `check` and `ci` default to cache mode
`normal`.

- In `check`, 197 of 367 units are cache-eligible, with 1,300 s of the
  8,814 s unit time. The 124 PostgreSQL-backed units (7,095 s) are
  ineligible: Section 8.3 of the NLSpec keeps "stateful tests, lifecycle,
  migration assertions, browser work, measurement, cleanup, destructive
  safeguards ..." uncached. The builds are ineligible too (policy `none`).
- The five `check` runs of 2026-11-24 were on successive commits, and each
  hit 0 or 1 units. No eligible unit kept its input digest from one commit to
  the next. The `test_rows` profile puts these paths in every closure, 1,092
  files in all: `tools/harness/**`, `tools/schemas/**`, `tools/generate*`,
  `contracts`, `configs`, `scripts`, the Makefile and the task-surface files.
- Over the last 150 commits:
  - 101 touched those paths and invalidated a median of all 197 eligible
    units;
  - 9 product-only commits invalidated a median of 3;
  - 40 documentation-only commits invalidated none.
- *Simulation for PostgreSQL-backed rows.* This uses the planner's closure
  rule (selected packages and their repository-local imports). The closures
  reach a median of 344 local packages, because integration tests compose the
  whole server. In the last 800 commits there were 37 product-only Go
  commits. Under a content-addressed key, those would rerun a median of 10 of
  the 124 units (about 700 s of the 7,095 s) and a mean of 51. Two in five
  would still rerun 92–124.

**Run roots.** A `check` root is 212 MB, 190 MB of it the event log (112,286
wait-segment pairs, G9 stage 2). A `ci` root is 567 MB, with a 345 MB event
log.

### 13.2 Stability Record

`ci`, eleven runs from 2026-11-24 to 25, three passed. The release-qualification
tracker has the detail on each failure.

| Run | Failure | Cause | Resolution |
| --- | --- | --- | --- |
| `-p34171` | NF-REQ-171a: the 250,000-row commit exceeded its 30 s budget | Capacity margin under load | Row-key migration and single-pass apply; the row now takes 16 s alone |
| `-p60919` | Release readiness contract | Test runner: a module-level scenario starved by a synchronous test | Registered async test |
| `-p89370` | Membership management live pages | Product race in paging after a confirmed write | Controller waits for the follow-up access check |
| `-p92912` | State-admission result precedence | Test fixture built inside a 1 s validation deadline | Fixture built before admission |
| `-p12690` | UI review stop; renderer sandbox Mermaid bundle | Lifecycle window; for the renderer, a concurrent forced rebuild republished the receipt under the running test (found in P4-2, 13.12) | Stop waits for the owner; UI review left `ci`; renderer reuse makes repeated builds republish the same installation (G37) |
| `-p61824` | Entity Find editing | Test pressed Enter while a rescan disabled navigation | Waits for the status first |
| `-p44999` | IR corpus import: the timeline apply job stalled for 30 s | Unknown | Stall diagnostics |
| `-p97887` | Backend module boundary check | Static rule broken by the stall diagnostics change | Package restructured |

- **Seven of the eight failing runs were load-dependent.** Each row passed
  alone, and each failure was a real defect in the product, a test or test
  support, not noise. Load is what exposed them, so the load stays.
- **Each one cost a `ci` run of about 28 minutes to find.** In three runs the
  failing browser test stopped its cohort, so 26, 37 and 6 other tests
  reported `infrastructure_failed` without running.
- **Every reproduction was manual:** the row alone, then beside 48 CPU-bound
  processes.
- **`check`.** Six of the seven Go-planned runs passed; the seventh failed on a
  real privilege defect in a new migration. Phase 3's load-sensitive
  Network Flow row limit has not failed since the single-pass apply.

### 13.3 Constraints

- **Coverage.** No row is removed, made `explicit_only` or narrowed to save
  time.
- **Isolation.** Every PostgreSQL unit keeps its dedicated generation, and
  every browser group its stack. No change introduces shared mutable state.
- **Reliability.** No retries: Playwright retries are already refused, and no
  Go rerun is added. No durability change (REQ-804). A timeout changes only by
  deriving it from the test deadline or an observable state. A cache hit must
  prove its whole key; any doubt is a miss.
- **Capacity is a physical stability limit.** Raising it needs the paired-run
  gate in 13.7.
- **Order.** Components first (P4-1 to P4-4), then the scheduler's structural
  fix (P4-5), then overlap (P4-6).
- **Verification** follows the owner's rules: targeted rows for each change;
  one diagnostic `check` when scheduling or capacity changes; paired runs only
  to adopt a gap; `ci` only when a change affects browser or `ci`-only work,
  with the reason stated; no release reruns in this phase.

### 13.4 Gap Detail

#### G30 — One admission authority, capacities stated once

- **Evidence.** 13.1: about 8 PostgreSQL units hold CPU tokens and logical
  lanes while they wait in host admission for 430 s of every `check`.
- **Fix.** Stage 3 of the work-graph tracker. One ledger decides admission
  over a unit's claims and its residents together, and a unit takes no token
  until every claim, its generation included, is available. Each capacity
  names the physical thing it bounds and is charged once per resident:
  `postgres_generations` 8 (today's effective limit of 16 lanes charged
  twice) and object stores 4. Physical concurrency is unchanged, so the disk
  load is unchanged. G16's 50 ms polling is redesigned with it: waiters wake
  on release.
- **Expected effect.** Little on `check`, which the heavy lanes and the eight
  generations bound. Plain Go, Vitest and operations finish earlier. In
  `ci`, browser groups may start as soon as generations free, which is
  partial overlap; validation must compare browser start times and failures.
- **Checks.** The scheduling, admission and native-runtime rows, the planner
  rows, one diagnostic `check`; to adopt, a paired `check` and one `ci`
  (browser start times change).
- **Risk.** G29 and Stage 0 each changed one layer and lost 200–270 s. That is
  why both layers become one, delivered in separately verified slices.

#### G31 — Critical-path reservation

- **Evidence.** The browser chain waits 370–406 s in every `ci`.
- **Fix, after G30.** The highest-ranked waiting unit reserves resources as
  they free, and lower-ranked units backfill only beside the reservation.
- **Expected effect.** The browser chain starts when its builds finish. The
  Go phase stretches, because four browser stacks hold four of the eight
  generations; it still ends inside the browser phase. `ci` ≈ builds plus
  about 1,050 s of browser work, about 1,150–1,250 s.
- **Risk.** Browser tests, the most load-sensitive, would run beside the
  PostgreSQL-heavy Go phase. This is the overlap the owner deferred. Adopt
  only through the 13.7 stability gate.

#### G32 — Browser lanes as a runtime capacity

- **Fix.** Browser lanes become a runtime capacity of stacks (default 4,
  maximum 6, as today). The planner stops chaining groups; when a stack
  frees, the scheduler starts the next ready group, longest first. A stack
  still serves consecutive groups of one execution, so stack reuse is kept.
  Estimates decide only the order, no longer the assignment.
- **Expected effect.** The browser phase approaches an even split: about
  1,155 s → 1,040–1,060 s.
- **Checks.** The planner browser rows, the scheduler rows, the lane-policy
  contract, one `ci` (browser scheduling changes).

#### G33 — Estimates from observations

- **Evidence.** Heavy units ran 604 s against 544 s longest-first; browser
  estimates are about twice the actual time. The authored estimates were last
  rebalanced on 2026-11-22 (G14), before G27 and G28 changed the heavy rows.
- **Fix.** A maintenance command refreshes the authored row estimates from
  named passing runs (median of the observations) and writes them as an
  ordinary reviewable change. The source tree stays the plan's only input, so
  plans stay reproducible, and the existing estimate audit becomes the drift
  check.
- **Expected effect.** `check` −40 to −60 s; the browser order in G32 also
  depends on it.
- **Checks.** The estimate audit, the planner rows, one diagnostic `check`.

#### G34 — Heavy placement and capacity after G28

- **Evidence.** Four heavy lanes were rejected three times on 2026-11-23 for
  contention failures (the Network Flow row limit, operator restore and an
  ATT&CK round trip). That was before G28 cut host flushes by 56%, and before
  the single-pass apply cut the Network Flow row from 22 s to 16 s alone. The
  limit of eight generations dates from G19 and G5b (2026-11-22).
- **Fix.**
  - *Placement audit.* Measure each heavy row's flushes and WAL with the G28
    sampler, and move rows whose I/O is now ordinary to ordinary generations.
  - *Re-test capacity.* Try four heavy lanes, and separately ten generations,
    by capacity override, with paired diagnostic `check` runs for each.
- **Expected effect.** `check` toward 450–500 s if either holds, nothing if
  both are rejected.
- **Risk.** These are the conditions that failed before, so the 13.7 gate
  applies in full: a single contention failure rejects.

#### G35 — Cache PostgreSQL-backed rows

- **Fix (amendment).** A Go row with a dedicated generation, and optionally an
  object store, becomes eligible for content-addressed caching. Its key binds
  today's closure plus the seed template digest (which binds the
  migrations), the service image digests, the fixture roots and the
  test-services helper. Everything else in the Section 8.3 list stays
  uncached: lifecycle, migration assertions, process tests, browser,
  measurement and destructive safeguards. A row whose outcome depends on
  wall-clock or host state opts out in the catalog.
- **Reliability.** Release qualification runs its `ci` with the cache off, so
  every release still executes every row.
- **Expected effect.** For the median product change, about 700 s of service
  work instead of 7,095 s. `check` after a localized change is then bounded
  by the longest invalidated unit plus the builds: ≤ 300 s. Full runs are
  unchanged.
- **Decision.** D12.

#### G36 — Narrow the test-row cache inputs

- **Fix.** The registered broad unit inputs give way to what a row's result
  depends on: the row runner's own implementation closure (its imports; once
  the row runner is in Go, the harness binary's package closure), the schemas
  it writes, the topology and scheduler inputs it reads, the toolchain pins
  and the selected owner manifests (already included). A check fails when the
  runner reads a file outside its declared closure, so the key stays
  closed-world.
- **Expected effect.** Harness and tooling commits stop invalidating product
  rows; 101 of the last 150 commits invalidated every eligible unit.
- **Risk.** An undeclared input gives a stale hit, which is why the closure
  is derived mechanically and checked.

#### G37 — Reusable build outputs

- **Fix.** `build_report_renderer`, `build_server_harness` and the other builds
  on the browser path get content-addressed reusable-output contracts. The
  NLSpec already allows build caches under a cache profile.
- **Expected effect.** `ci` −76 s when the renderer inputs are unchanged
  (most commits), and −56 s more when no server input changed. `check` is
  unaffected, because the builds are not its critical path.

#### G38 — Browser cohorts run to completion

- **Evidence.** `tools/harness/browser/browser-catalog-group-cli.mjs` passes
  `--max-failures=1`. Three failing `ci` runs left 26, 37 and 6 tests
  unexecuted.
- **Fix.** Cohorts whose state is `test_owned` drop `--max-failures=1`: each
  test owns its state, so one failure cannot poison the next. Cohorts whose
  tests share state, and harness or stack failures, keep stopping at the
  first failure.
- **Expected effect.** No change on passing runs. A failing run reports every
  failing test, so one `ci` run finds all of a cohort's defects.
- **Decision.** D14, if the NLSpec requires stop-on-first-failure.

#### G39 — Find load-dependent defects without `ci`

- **Stress command.** `make test-stress ROWS=… REPEAT=N LOAD=<profile>` runs
  rows repeatedly beside a declared synthetic load (CPU, disk flushes, or
  both) and records each repetition. This replaces the manual runs beside
  48 CPU-bound processes. It is a new public target (D15).
- **Failure-scoped host evidence.** A 1 Hz in-memory ring of host pressure
  (CPU, I/O and memory PSI). The window around each failed unit is written
  into its unit logs, so every load-dependent failure carries its load
  context without `HARNESS_DIAGNOSTICS=basic`.
- **Wait audit.** Go tests have 84 `time.Sleep` calls and 199
  fixed-duration `context.WithTimeout` calls; the browser specs set 47
  explicit timeouts. Classify each one as deadline-derived (the G17g pattern),
  an observable-state wait, or justified. Fix the rest owner by owner.
- **Open defects.** The renderer sandbox flake (1 in 51) and the IR corpus
  apply stall now retain their cause class. Record each recurrence here.
- **Expected effect.** No wall change; fewer `ci` runs spent finding defects.
- **Outcome (P4-4).** Delivered as planned; the gate result and the audit
  inventory are in 13.10. Neither open defect recurred in 20 repetitions of
  its row under load.

### 13.5 Workstreams

| Step | Work | Gaps | Exit |
| --- | --- | --- | --- |
| P4-0 | Reference runs: one diagnostic `check` and one diagnostic `ci` at HEAD with the cache off, because the 13.1 runs have no pressure samples | — | Both pass; recorded in Section 11. **Done 2026-11-25:** `p40-ref-check01` and `p40-ref-ci01` (13.9) |
| P4-1 | Go phase components | G33, G34 | Estimates refreshed; placement audited; capacity re-tests adopted or rejected through the gate. **Done 2026-11-25** (13.11): estimates refreshed, four heavy lanes adopted, placement unchanged; the ten-generation re-test moves after P4-5 |
| P4-2 | Browser phase components | G32, G38 | Dynamic lanes adopted after one `ci`; failure policy decided. **Done 2026-11-25** (13.12): adopted; the `ci` failed one row on the renderer receipt race, fixed in P4-6 (G41; 13.12's cause was wrong) |
| P4-3 | Daily-development cache | G36, G37, then G35 | Hit rates measured on representative commits; G35 only after D12. **Done 2026-11-25** (13.13) |
| P4-4 | Stability tooling | G39 (host evidence first, then the stress command, wait audit and open defects) | The week's failing rows pass `REPEAT=10` under load. **Done 2026-11-25:** 60 of 60 repetitions passed (13.10) |
| P4-5 | One admission authority | G30 | Paired `check` and one `ci`. **Done 2026-11-26** (13.14): paired `check` pass; by owner decision its `ci` folds into P4-6's paired `ci`, and correctness was validated with targeted rows, `ci`-only rows included |
| P4-6 | Overlap | G31 | Through the 13.7 gate, or recorded as rejected. **Done 2026-11-26** (13.15): adopted; the paired `ci` exposed G41 and G42, both fixed and proven on segments; the passing paired `ci` folds into P4-7 |
| P4-7 | Validation and handoff | — | Paired `check`, paired `ci`, Section 3 refreshed |

P4-1 to P4-4 are independent and can proceed in any order. P4-5 starts after
P4-1 and P4-2, so the scheduler change is measured on stable components.
P4-6 comes last.

**Execution order (2026-11-25).** P4-0, P4-4, P4-1, P4-2, P4-3, P4-5, P4-6,
P4-7. P4-4 moved first because G34's capacity re-tests can be adopted only
through the 13.7 gate, which needs the stress command.

### 13.6 Expected Effect

| Lever | `check`, full | `check`, localized change | `ci` |
| --- | ---: | ---: | ---: |
| G33 | −40 to −60 s | same | small |
| G34 | 0 to −100 s | same | small until G31 |
| G32 | — | — | −100 to −120 s |
| G37 | — | — | −76 to −132 s when unchanged |
| G36 | — | tooling commits stop invalidating rows | same for `ci`'s cached rows |
| G35 | — | about 600 s → ≤ 300 s | its Go rows likewise |
| G30 | 0 to −30 s | — | prerequisite for G31 |
| G31 | — | — | −300 to −400 s |
| G38, G39 | — | — | fewer runs to find defects |

**Projection.**

- *`check`, full run:* about 545–565 s after P4-1, and ≤ 500 s only if G34
  holds.
- *`check`, localized product change:* ≤ 300 s after G35.
- *`ci`:* about 1,450–1,500 s after P4-2 and P4-3, and about 1,150–1,250 s
  after G31.

### 13.7 Validation Protocol and Stability Gate

- **Per change.** Targeted rows, found through `make task-guide`.
- **Scheduler, capacity or planner change.** One diagnostic `check`
  (`HARNESS_DIAGNOSTICS=basic`, cache off), compared with the P4-0 reference.
- **Browser scheduling change.** One diagnostic `ci`, with the reason stated.
- **Adopting a gap.** Paired runs of the affected command.
- **Stability gate.** Any change that raises concurrency or overlap (G31,
  G34) is adopted only if all of these hold:
  - its paired runs pass;
  - heavy-unit work and browser runner time grow by no more than 15%;
  - the rows that failed in 13.2 pass `REPEAT=10` under the stress command's
    declared load.

### 13.8 Decisions and Amendments

| ID | Decision | Needed by | Recommendation |
| --- | --- | --- | --- |
| D11 | Make Stage 3 (one admission authority) phase-4 work, after P4-1 and P4-2 | P4-5 | Approve; it answers the work-graph tracker's open question. **Approved 2026-11-25** |
| D12 | Cache PostgreSQL-backed rows under a key that binds the seed template and service images | P4-3 | Approve, with release `ci` run cache-off. **Approved 2026-11-25** |
| D13 | Narrow test-row cache inputs to the runner's own closure | P4-3 | Approve. **Approved 2026-11-25** |
| D14 | Run test-owned browser cohorts to completion | P4-2 | Approve unless the NLSpec requires stop-on-first-failure. **Approved 2026-11-25** |
| D15 | Add a public stress command | P4-4 | Approve. **Approved 2026-11-25** |
| D16 | Re-test four heavy lanes and ten generations | P4-1 | Approve as measurement; adopt only through the 13.7 gate. **Approved 2026-11-25** |
| D17 | Critical-path reservation, overlapping the browser and Go phases | P4-6 | **Adopted 2026-11-26** on the `check` pair and `p46-ci01`; the passing paired `ci` moves to P4-7 (13.15) |
| D18 | Fold P4-5's `ci` into P4-6's paired `ci`; validate P4-5 with targeted rows | P4-5 | **Owner decision 2026-11-26** |

| ID | Section or requirement | Change | Gap |
| --- | --- | --- | --- |
| A22 | Scheduling and host admission (REQ-802, REQ-803) | One admission ledger; capacities per physical resource, charged once | G30 |
| A23 | REQ-802, REQ-803 | Reservation for the highest-ranked waiter; backfill only beside it | G31 |
| A24 | Browser lane policy | Lanes are a runtime stack capacity, not a plan-time chain | G32 |
| A25 | Section 8.3 cache policy | Go rows with dedicated services become cacheable under the extended key | G35 |
| A26 | Unit-aware closure rules | The runner's implementation closure replaces the registered broad unit inputs | G36 |
| A27 | Browser failure policy | Test-owned cohorts run to completion | G38, if needed |
| A28 | Command tables | Stress command | G39 |

### 13.9 Phase 4 Reference (P4-0)

Both runs used `HARNESS_DIAGNOSTICS=basic` and `CARTULARY_HARNESS_CACHE_MODE=off`
on `c9502cb39`. Diagnostics add time: the 13.1 `check` without them took 605 s.

| Measure | `check` `p40-ref-check01` | `ci` `p40-ref-ci01` |
| --- | ---: | ---: |
| Result | Pass, 368/368 | Pass, 502/502 |
| Graph wall | 691 s | 1,791 s |
| Guest pressure: CPU some, I/O full | 5.1%, 7.5% | 4.6%, 5.3% |
| PostgreSQL units (not heavy): host admission, runner | 3,799 s, 2,132 s | 4,642 s, 2,485 s |
| Heavy units: work; lanes end | 1,836 s; 690 s (longest first on actual durations: 623 s) | 1,880 s; 745 s |
| Run root; event log | 242 MB; 174 MB | 621 MB |

- **`check` critical path.** `build_report_renderer` (137 s), then
  `build_server_harness`, eligible at 140 s but started at 372 s, then a
  heavy unit (`recovery.process.inactive_current_restore…`), eligible at 422 s,
  started at 609 s when a heavy lane freed, ending the run at 690 s. Multi-token
  and scarce-lane starvation (G31) therefore bounds `check` too, not only `ci`.
- **Heavy order.** Heavy estimates ran 0.49–1.29 of actual; the longest unit
  (`network_flow_portability`, 206 s) started at 307 s.
- **`ci` browser phase.** The four lane heads became eligible at 129 s and
  started at 600–606 s. Lane loads were 984, 1,116, 1,189 and 935 s against an
  even 1,056 s; browser runner time was a median 0.47 of its estimates
  (0.16–0.82). Fresh stack acquisition took 8–16 s and a reused stack's reset
  4 s, so the 11 reuses in a `ci` save about 66 s of lane time, about 16 s of
  wall (input to G32).
- **Build estimates.** `build_report_renderer` ran 137 s against a 15 s
  estimate, `build_server_harness` 50–68 s against 5 s (input to G33).

### 13.10 Stability Gate and Wait Audit (P4-4)

Each row that failed `ci` in 13.2 and is still in `ci` ran
`make test-stress OWNER=<owner> ROWS=<rows> REPEAT=10 LOAD=mixed` at
`465a30afe`: two busy threads per CPU and four `fdatasync` writers beside every
repetition, the selected rows executed fresh each time.

| Owner | 13.2 rows | Repetitions passed | Wall per repetition (median, range) |
| --- | --- | ---: | --- |
| `module.extensions` | State-admission result precedence | 10/10 | 61 s (57–75) |
| `harness.release` | Release readiness contract | 10/10 | 277 s (266–283) |
| `module.imports` | Network Flow row limit; IR corpus import | 10/10 | 181 s (176–222) |
| `module.incidents` | Membership management live pages | 10/10 | 161 s (154–168) |
| `module.workbook` | Entity Find editing | 10/10 | 153 s (148–159) |
| `module.reporting` | Renderer installed sandbox | 10/10 | 260 s (253–272) |

- **All 60 repetitions passed**, so the 13.7 gate's third condition holds at
  this commit, and no failure window was written. It is a baseline, not an
  adoption: a change that raises concurrency still needs its own paired runs.
- **Open defects.** The renderer sandbox and IR corpus rows passed 10 of 10
  each; neither the sandbox flake nor the apply stall recurred.
- **Wait audit, Go.** The scan found 297 `context.WithTimeout` calls with a
  fixed or computed duration (13.4's 199 counted fewer patterns) and, after
  the conversion, 87 `time.Sleep` calls.
  - *Converted (31 in 27 files, `bdd8f4e1d`).* Every whole-test hang guard,
    the first statement of a test or of a function a test runs as a subtest,
    now derives from the test deadline less a 30 s settlement allowance
    (`internal/testutil/testdeadline`), as G17g did for process tests; the
    `httptestx` poll helper uses the same rule.
  - *Justified as they stand.* 35 sleeps are poll intervals of an
    observable-state wait, 7 park a helper process, and 45 hold a negative
    window or pace a fixture; 53 timeouts bound cleanup, 11 already derive
    from a deadline, 38 are deliberately short probes of timeout behaviour
    and 7 take a computed duration; two whole-test bounds time a
    characterization (180 s) and a process harness (30 min).
  - *Next candidates.* 155 fixed bounds scope a helper, subtest or single
    call (58 under the test context, 97 under `context.Background` or
    another parent). A scoped bound fails only when its operation is slow,
    which `make test-stress` now exposes; convert them owner by owner as
    stress runs implicate them.
- **Wait audit, browser.** The 70 explicit timeouts bound observable-state
  assertions (`toBeVisible`, `toPass`, `poll`); the 64 `test.setTimeout` calls
  set test budgets; the one fixed wait is a 500 ms negative window proving no
  further queries arrive. None needs conversion.

### 13.11 Go Phase Components (P4-1)

All runs are diagnostic `check` runs with the cache off; each passed 1,893 rows.

| Measure | `p40-ref-check01` | `p41-g33-check01` (G33) | `p41-heavy4-check01`, `-02` (G33, four heavy lanes) |
| --- | ---: | ---: | ---: |
| Graph wall | 691 s | 719 s | 706 s, 704 s |
| Heavy units; their work | 15; 1,836 s | 13; 1,989 s | 13; 1,979 s, 1,966 s |
| Heavy lanes end (second-last, last unit) | 616 s, 690 s | 653 s, 718 s | 506 s, 668 s (both) |
| `build_server_harness` eligible, started | 140 s, 372 s | 127 s, 127 s | 70 s, 70 s; 65 s, 65 s |
| Ordinary PostgreSQL host-admission wait | 3,799 s | 3,173 s | 3,917 s, 3,944 s |

- **G33 removed the build delay but not the wall.** With observed estimates
  the server-harness build starts the moment it is eligible, and
  `build_report_renderer` is planned at its observed 98 s median instead of
  15 s. Accurate estimates also let the heavy packing target hold more rows,
  so the heavy rows formed 13 units instead of 15; the three largest started
  together and heavy work rose 8%, within the gate's 15%. With three heavy
  lanes the heavy phase (about 663 s of work per lane) ends the run.
- **Four heavy lanes end the heavy phase by 506 s**, without growing heavy
  work, but `check` gains only 13–15 s: the run now ends on an ordinary
  PostgreSQL unit (`app.server.process.a_fresh_process…`, 11 rows, 133 s)
  that was eligible at 144–147 s and waited 424 s for object-store and
  PostgreSQL lanes and CPU while shorter units were admitted ahead of it.
  The last heavy unit (`recovery.process.inactive_current_restore…`) waited
  the same way. That starvation is G30/G31, so the gain of four lanes
  appears only after P4-5 and P4-6.
- **Adoption.** The 13.7 gate holds: paired runs pass, heavy work does not
  grow against three lanes, and the 13.2 rows pass under load (13.10). Four
  lanes are the default from `d77ba4a25`; P4-5 and P4-6 measure with them.
- **Placement audit: inconclusive, placement unchanged.** The G28 sampler's
  per-database statistics were matched to units by test-process database
  names and time; only 23 of 122 PostgreSQL units matched, and database
  tuple writes miss the file and object-store I/O that keeps several rows
  heavy (bundle round trips, imports). Rows stay where G27 and G28 left
  them; a placement change needs per-unit I/O evidence the harness does not
  yet record.
- **Ten generations moves after P4-5.** Today a unit and its generation
  each charge a PostgreSQL lane, so ten generations need 20 lanes, and the
  policy bounds lanes by CPU tokens (19 on this host). P4-5 states
  generations as their own capacity; the re-test runs there.

### 13.12 Browser Phase Components (P4-2)

One diagnostic `ci` with the cache off, `p42-ci01` on `7ebf2d0af`, against the
P4-0 reference. It ran because G32 changes how browser groups are scheduled,
which only `ci` exercises.

| Measure | `p40-ref-ci01` | `p42-ci01` |
| --- | ---: | ---: |
| Result | Pass, 2,646 rows | Fail, 2,647 of 2,648 rows (renderer receipt race) |
| Graph wall; browser phase end | 1,791 s; 1,789 s | 1,688 s; 1,687 s |
| Browser groups; runner time | 29; 3,906 s | 27; 3,848 s |
| First four ordinary groups started | 600–606 s | 551, 569, 588, 713 s |
| Stack acquisition, median and total | 8 s; 264 s | 9 s; 291 s |

- **G32 adopted.** Groups now start whenever a lane frees, ranked by their
  observed estimates; the browser phase ends 102 s earlier with the same
  runner time, the low end of the 100–120 s expected. Every group acquires a
  fresh stack: 27 s more acquisition than the reference's reuse, as
  expected. The first groups still wait from 108 s (eligible) to 551 s for
  resources the Go phase holds; that is G31.
- **G38 adopted, not yet exercised.** No browser test failed, so the
  run-to-completion policy had nothing to show; the next browser failure
  will report its whole cohort.
- **The failure is the renderer sandbox defect of 13.2, now explained.**
  `module.reporting.integration.renderer_installed_sandbox` failed with
  "candidate toolchain is not bound to its installation". The test binds
  the installation the receipt names when it starts; at 156 s
  `harness.assurance_release_readiness_evidence`, whose recipe re-enables
  Make prerequisites and forces `$(REPORT_RENDERER_RECEIPT)`, rebuilt the
  renderer and published a receipt naming a different installation, so the
  test's later renderer loaded another one. Renderer builds are not
  reproducible (four distinct installation identities from unchanged
  sources), so every forced rebuild can flip the receipt under a running
  renderer test. G37's input-keyed reuse makes a repeated build republish
  the same installation, which removes the flip; P4-3 verifies it.

### 13.13 Daily-Development Cache (P4-3)

- **Design change from 13.4 (owner-approved).** G36's runner closure was first
  built as a scan of the JavaScript row runner's imports. That runner loaded
  the whole catalog for every Go unit, so Go rows moved to a Go row runner
  inside the harness runtime (TH-HARNESS-REQ-872): the planner emits each Go
  unit's exact `go test` command and per-row tests, and the runtime settles
  rows from the event stream. Graph and plan move to v10 and the runner
  registry to v4. The JavaScript runner keeps Vitest and shell rows.
- **Correctness addition.** Harness test support launches Node modules by
  literal path (for example the PostgreSQL resource stop module, reached by
  182 units), so a Go key also holds every repository file its Go code names
  by a literal path, with a named module's closure. Without it, narrowing
  would have reused stale results after such a module changed.
- **Measured.** `p43-check01`, the first `check` after the change, missed
  every eligible unit, executed all 1,893 rows through the new runner and
  passed in 663 s. `p43-check02`, run straight after, reused 313 of 313
  eligible units; the 53 uncached units (lifecycle, migration, measurement,
  process and similar work) bound it at 90 s.
- **Hit rates on representative commits.** Replaying the last 150 commits
  against these keys: the 44 documentation-only commits invalidate nothing;
  the 9 product-only commits invalidate a median of 6 of 313 units (mean
  33); of the 97 commits touching tools, contracts or the Makefile, 60 still
  invalidate every unit (101 of 150 did before G36). A change to a module's
  root package reaches 120–177 units, because test support assembles the
  whole application.
- **G40.** Those 60 commits change inputs every key still holds: harness
  owner models and test families, schemas, generated validators and the task
  surface. Tests read such files through paths they compute, which no
  closure rule can see, so removing them needs each package to declare what
  it reads at run time and a way to enforce it. That is a separate change.
- **G37.** The first build recorded its input key in 94 s; every later build
  in both runs, including the forced rebuild inside `build_server_harness`,
  reported the same installation "(reused)". Reuse costs 31 s, mostly hashing
  the Chromium directory and verifying the 1.4 GB root against its manifest.
  Because repeated builds now republish the same installation, the receipt
  race behind the renderer sandbox failure (13.12) cannot flip the
  installation under a running test.

### 13.14 One Admission Authority (P4-5)

G30 as planned: the scheduler evaluates its ranked ready units against one
ledger view and transacts only for units the view admits, so a waiting unit
holds no token; waiters wake on ledger changes; and PostgreSQL work (8 lanes)
and generations (8, charged by resident clusters) are stated once instead of
16 lanes charged twice. Registry v10, capability snapshot v2, override v2,
resident allocation v3, local session v5, host admission record v4,
execution index v7 and critical path v5 carry the new `postgres_generation`
resource.

| Measure | P4-1, four heavy lanes | `p45-check06`, `-07` | `p45-gen10-check01`, `-02` (ten generations) |
| --- | ---: | ---: | ---: |
| Graph wall | 706 s, 704 s | 577 s, 567 s | 573 s, 558 s |
| Heavy units' work; heavy lanes end | 1,979 s, 1,966 s; 668 s | 1,961 s, 1,939 s; 552 s, 556 s | 1,958 s, 1,915 s; 510 s, 545 s |
| Time units spent inside host admission | 3,917 s, 3,944 s | 0 s, 0 s | 0 s, 0 s |

- **`check` −130 s.** Units no longer hold CPU and lanes while they wait,
  so long ordinary units start when their resources free. The run now ends
  when PostgreSQL generation throughput runs out: its last unit
  (`schema_bootstrap`, 17 s) waited for generations behind higher-ranked
  work. Heavy work did not grow.
- **Ten generations not adopted.** They changed nothing measurable after
  G30 (573 s and 558 s against 577 s and 567 s); eight stay, and G34 is
  closed.
- **Defects found, all fixed with targeted tests.** The first diagnostic
  runs aborted (`harness/fixture_error`) whenever generations ran short:
  a unit denied for capacity held only through another unit's resident, or
  outside the graph, published a wait no event vocabulary allows. Holders now
  include residents; otherwise one unit waits in host admission, where
  pressure retirement frees idle generations, and the rest wait by rank. A
  ledger evaluation error now fails only its unit. Each single admission
  attempt is now its unit's `host_admission` phase, so diagnostic evidence
  validates. The JavaScript registry validator and the resident and
  observation schemas gained the new resource. A selected-service boundary
  test now asks for fresh execution (`--fresh`), because G35 made its probe
  rows cacheable.
- **Correctness validation (D18).** Targeted rows: the runtime, scheduler,
  admission, fixture, session, observation and collection rows, the nested
  graph and service boundary rows, the release and Playwright adapter rows,
  and one real browser group under diagnostics
  (`module.incidents.browser.membership_management_live_pages`), plus the
  source checks and lints. The `ci` measurement of G30's browser effect
  moves to P4-6's paired `ci`.

### 13.15 Overlap (P4-6)

G31 as planned: when graph units hold the deficits of the highest-ranked
waiting unit, the scheduler reserves them, and other units start only when
they claim none of the reserved resources.

| Measure | Reference | P4-6 | Gate |
| --- | ---: | ---: | --- |
| `check` graph wall | 577 s, 567 s (`p45-check06`, `-07`) | 597 s, 591 s | Both pass |
| `check` heavy-unit work | 1,961 s, 1,939 s | 1,964 s, 1,892 s | Within +15% |
| `ci` graph wall | 1,688 s (`p42-ci01`) | 1,445 s (`p46-ci01`) | — |
| `ci` first browser group start | 551 s | 211 s | — |
| `ci` browser runner time | 3,848 s | 4,279 s (+11%) | Within +15% |
| `ci` heavy-unit work | 2,147 s | 1,848 s | Within +15% |
| `ci` rows | 2,647 of 2,648 | 2,648 of 2,649 | One failure, G41 |

- **Effect.** `ci` is 243 s faster than P4-2 and 346 s faster than the P4-0
  reference (1,791 s). This includes G30, whose `ci` measurement folded here
  (D18). Browser groups start 340 s earlier. `check` is 20 s slower, inside
  the ±30 s spread. The projection was 1,150–1,250 s; `ci` now ends on the
  browser phase, whose 4,279 s of runner time over four lanes is about
  1,070 s, so browser lane capacity bounds it.
- **G41, the failure.** Both `ci` runs failed
  `module.reporting.integration.renderer_installed_sandbox`, as P4-2's `ci`
  had. The test opened the installation the receipt named when it started,
  then built each renderer by reading the receipt again. In `ci`, the nested
  graph of `harness.browser.integration.native_graph_observations` runs its
  own renderer build in the Go test's environment, reuses a different
  installation (`c9d0…` against `d438…`) and moves the receipt about 8 s after
  the test starts. 13.12's explanation (a forced rebuild from release
  readiness) and its claim that G37 removed the race were wrong: every build
  in these runs reused an installation. Every renderer in the test now binds
  the installation it opened; production still opens the receipt once at
  startup. Proof: with the receipt moved 2 s after the test binary starts,
  HEAD fails with the `ci` message (`seg-ontest-before`) and the fix passes
  (`seg-ontest-after`). Two earlier replays that moved the receipt at fixed
  times passed HEAD by luck, so the trigger follows the test process.
- **G42, found while tracing G41.** Fifteen build profiles put the Go
  temporary directory on the keyed command, and every graph run creates a new
  one, so every run missed. `build-server-harness` gained one record per run
  with an unchanged output digest. `embedded-web-assets` also keyed its
  per-run web source path. Locations now reach producers through the
  environment, input directories are keyed by content (specification
  amended), and the cache test gains a relocation case. Proof: two
  consecutive runs of `harness.browser.boundary_support.https_trust`. The
  first wrote one record per Go build profile (new key format), the second
  none, and its build operations took 10.9 s against 18.6 s.
- **Left for Phase 5.** During `ci` the nested graph wrote a different server
  binary (`7117…` against `74e9…`) into the shared path (G43).
  `build_server_harness` still repeats the renderer reuse check through a Make
  prerequisite the graph does not model: 8 s on a hit, about 20 s under
  `check` load (G44).
- **Gate.** The `check` pair passes and every work limit holds. The `ci` pair
  did not pass: `ci01` failed on G41 alone, and the owner stopped `ci02` after
  it failed the same row. G31 is adopted (D17). The passing paired `ci` folds
  into P4-7's final validation, and the stress condition keeps the 13.10
  baseline.

## 14. Phase 5 Proposal — Flat Catalog and Thin Make

Status: proposed 2026-11-26 at the owner's request. P5-0, P5-1 and P5-2 done (14.5, 14.6, 14.22); P5-4 done (14.7 to 14.21, 14.23); P5-3 planned (14.24).

### 14.1 Findings

- **Nested executions (G43).** Five `ci` rows drive real repository builds
  from inside a running graph:
  - `native_graph_observations` (two nested graphs)
  - `native_selected_service_lifecycle` (four)
  - `native_nested_cancellation` (`make test-slice`, three modes)
  - `frontend_artifact_lifetime` (two concurrent `make build`)
  - the `real_output_*` shell rows

  They inherit the Go test's toolchain environment (`CGO_ENABLED=0`,
  `GOENV=off`, `GOWORK=off`), so their build keys differ and they rename
  different binaries over shared outputs that outer units are using. Their
  failures are reported as the first 4 KiB of output (`%.4096s`) and always
  classed as product assertions, and their rows are never counted.
- **Make and planner duplication (G44).**
  - A `make check` parses the Makefile at least twice (admission re-entry).
  - The 26 artifact and installation operations run `make -- <files>` again.
    Each such call re-parses everything and hits FORCE rules that
    `cache-artifact.sh` keys separately from their Make prerequisites.
  - Six switches decide whether prerequisites run (`SKIP_PREREQUISITES`,
    `GRAPH_CHILD`, an unread `GRAPH_ARTIFACT_CHILD`, and others), and 17
    targets have two recipes chosen by `GRAPH_CHILD`.
  - Freshness has five owners: planner cache policy, `cache-artifact.sh`,
    the renderer store, web staging and Make mtimes.
  - Build profiles and test units use different toolchain environments.
- **Catalog layering.** There are 2,754 rows, 51 private selections (six
  contain other selections) and 183 operations. Some checks are policy
  operations with no row (the build cache test is reachable only through
  `harness-contract SUITE=regression`), so they cannot be run alone as a
  segment.

### 14.2 Proposed workstreams

| Step | Work | Gaps |
| --- | --- | --- |
| P5-0 | Fences: nested graphs get an allowlisted environment and a private build root instead of the shared repository outputs; `build-go-artifact.sh` publishes the binary and its crypto receipt together; nested failure output keeps the tail and the child's failure class; the graph-only renderer prerequisite of the server harness build goes. **Done 2026-11-26** (14.5): the environment fence made the private build root unnecessary, and the binary and receipt pairing moved to P5-3 | G43, G44 |
| P5-1 | Flat catalog: every check is a row with an owner, so any of them runs alone with `ROWS=`; selections list rows and operations, not other selections. **Done 2026-11-26** (14.6): every test script is a row and duplicates are gone; selection nesting stays (14.6) | G44 |
| P5-2 | Unnest: harness-mechanics probes run as first-class graph rows against private outputs, or move to an explicit harness suite outside `ci`; a source check forbids rows that launch `make` or `graph` against the shared repository | G43 |
| P5-4 | Declared environment: a registry of every harness variable, default-deny unit environments, one typed unit context document, fixture bindings in that document (14.7) | G43, G45 |
| P5-3 | Thin Make: Make only admits the workspace and forwards to the harness; build operations become data with declared inputs, environment keys and outputs; one Go content-addressed store replaces `cache-artifact.sh`, the renderer store and web staging; no unit re-enters Make; one toolchain environment from the pins for builds and tests. **Planned 2026-11-26** (14.24): five steps, P5-3a to P5-3e; P5-3a done (14.25) | G44 |

P5-0 is small and removes the hazards seen in `ci`. P5-1 makes segments
possible for every check. P5-3 is the largest and needs specification
amendments.

### 14.3 Constraints

- The NLSpec's public-behavior gate requires specification changes before
  target identity, `command_id` or task-surface metadata change. Its
  graph-child clauses encode Make prerequisite semantics.
- The generated-artifact policy requires the three generated `.mk` files.
- The execution topology names Make variables (`output_make_variable`).
- `CARTULARY_MAKE_INPUT_SOURCES` relies on Make's `$(origin)`.
- Retired Make names are part of the command contract.

### 14.4 Decisions

| ID | Decision | Recommendation |
| --- | --- | --- |
| D19 | Approve Phase 5 scope and order (P5-0, P5-1, P5-2, P5-3) | Approve P5-0 now; decide P5-1 to P5-3 after it. **P5-0 and P5-1 approved by the owner 2026-11-26**, with authority to deprecate tests covered elsewhere |
| D20 | Where harness-mechanics probes run | An explicit harness suite outside `ci`, run when harness code changes. **Decided by the owner 2026-11-26: explicit `harness-contract` presets** (14.22) |
| D21 | Approve P5-4 scope and order (P5-4a to P5-4d); P5-4c redefined by the owner on 2026-11-26 as single-writer ownership (14.10), and P5-4d as single owners (14.14) | Approve P5-4a and P5-4b now; decide P5-4c and P5-4d on P5-4b's measurements. **Approved 2026-11-26** |
| D22 | How a unit finds its context document | A path in one variable: it survives Make re-entry and child tools; the admission lease keeps its file descriptor. **Approved 2026-11-26** |
| D23 | P5-4 against P5-2 and P5-3 | P5-4a and P5-4b before P5-2, since unnesting is simpler once children start from a declared environment; mode switches leave with P5-3. **Approved 2026-11-26** |
| D24 | Where the P5-3 store lives | Repository-local `.cache/cartulary/store`, beside today's caches, so locks and collection stay within one checkout; sharing across worktrees is a later change. **Approved 2026-11-26** |
| D25 | How conventional paths stay valid | Published views swapped by rename: hard links for files and an atomically renamed tree for directories. Never a symbolic link, which the container build would copy as a link. **Approved 2026-11-26** |
| D26 | cgo for shipped binaries | `CGO_ENABLED=0` everywhere, as tests and the package probes already use; adopted only if the release package rows pass. **Approved 2026-11-26** |
| D27 | Checkout-independent binaries | `-trimpath` for product binaries, so a binary's digest does not depend on the checkout path; the one-time digest change is recorded. **Approved 2026-11-26** |
| D28 | Cache controls | Keep the public `CARTULARY_BUILD_CACHE_DISABLE` and `CARTULARY_FORCE_REBUILD` as the store's two controls; fold the readiness, reinstall and static-analysis variants into them. **Approved 2026-11-26** |

### 14.5 Fences (P5-0)

Approved by the owner on 2026-11-26 (D19, P5-0 only).

- **Nested environment.** The runtime records in
  `CARTULARY_HARNESS_UNIT_OVERLAY` what each unit's own overlay replaced: its
  planned command environment and, for Go units, the runtime's Go binding.
  `harnessruntime.NestedEnvironment` undoes it. The three Go-hosted nested
  launches (`native_graph_observations`, `native_selected_service_lifecycle`,
  `native_nested_cancellation`) start from it instead of their test's
  `CGO_ENABLED=0`, `GOENV=off` Go build context. Their builds therefore key
  like the enclosing run's and are cache hits. This replaces the planned
  private build root: a nested build with matching keys writes nothing, so a
  second output tree is unnecessary. Specification amended.
- **Renderer prerequisite.** In graph-child mode the server-harness build no
  longer re-runs the renderer build through Make; the graph already orders it.
- **Nested failures.** The nested launch sites report the last 4 KiB of
  output (they printed the first 4 KiB) and the nested run's own failure class
  and reason. While running these segments, a compile error in a boundary test
  surfaced only as `infra/missing_selector_result`, with the compiler text
  left in the unit's standard output. The Go row runner now puts the last
  4 KiB of compiler output at the head of the failure detail
  (`[GO BUILD FAILURE]`). Specification amended.
- **Moved to P5-3.** Publishing a binary and its cryptographic receipt as one
  unit needs a single pointer over a content-addressed directory, which is
  P5-3's store. With the environment fence, no nested build republishes them
  during a run.

`native_graph_observations`, one row with two nested graphs:

| Segment | Renderer receipt | Shared server-harness binary | New build cache records | Nested server-harness build |
| --- | --- | --- | --- | --- |
| `p50-nested-before` (HEAD) | Moved to `c9d0…` | Replaced (`101da0…` → `dd659a…`) | `build-migrate`, `build-server-harness`, `testservices-build`: one each | 9.6 s; re-ran the renderer |
| `p50-nested-steady` (fenced, second run) | Stayed `d438…` | Unchanged | None | 1.1 s |

The nested https graph took 28.8 s against 46.3 s. The rows whose launch code
changed pass (`p50-nest2`, `p50-units`). Not exercised: the nested launch
sites' failure reporting, which needs a failing nested graph; compilation
covers it.

### 14.6 Flat Catalog (P5-1)

Approved by the owner on 2026-11-26, with authority to deprecate tests covered
elsewhere.

- **Every test script is a row.** 99 of 150 policy operations had no row. Of
  those, 23 ran repository test scripts: `make test-slice … ROWS=` could not
  reach them, and their failures were operation failures, not row results. 20
  now have shell rows with the existing operation binding. 19 are
  `explicit_only`, because they ran only in the explicit `harness-contract`
  suites, and their suites keep running them. `public_wrapper_smoke` is
  `standard`, because it already ran in `check` and `ci` through
  `contract_smoke`. Owners follow the existing bound rows: 16 in
  `harness.command_surface` (the release task-surface test moved there in
  P5-4b, because explicit-only `harness.release` rows are release package
  rows), and two each in `harness.evidence_accounting` and
  `harness.generated_artifacts`. One shell
  verification, `harness.command_surface.verification.helper_scripts`, covers
  the harness helper scripts. Of the other 76 unbound operations, four run
  `go test` directly (below); the rest are commands (lints, static analysis,
  builds, dev and db tools) that keep their own public targets.
- **Enforced.** The catalog's source coverage now includes the shell runner,
  which it skipped before. Every operation that runs a `tests/test-*.sh` or
  `tests/test-*.mjs` script must be selected by exactly one shell row. Two
  operations may run the same script only for different cases: environment
  assignments other than output presentation, or arguments after the script.
  Violations are `catalog_coverage` failures (specification amended).
- **Duplicates removed.** Five operations re-ran tests that cataloged rows or
  bound operations already run:
  - `assurance_observability`, `assurance_agent_finalize` and
    `assurance_canonical_evidence` ran the same files as the `ci` operations
    `harness_observation_contract` and `harness_evidence_contract`;
  - `assurance_fixture_broker_matrix` and
    `assurance_work_graph_scheduler_matrix` ran Go tests that the rows
    `harness.browser.unit.native_fixtures` and
    `harness.command_surface.unit.native_runtime` select.

  `contract_regression` now lists the covering operations and rows, so the
  regression suite keeps its coverage (49 planned units).
- **Unchanged composition.** `check` and `ci` plan the same 366 and 501 units
  as before. Their only new row is `public_wrapper_smoke`, bound to a unit they
  already ran.
- **Kept.** `harness_fixture_contract` and `harness_runtime_contract` run whole
  harness packages under `-race`. That coverage exists nowhere else, so these
  are not duplicates, but they are tests outside the catalog. The catalog
  already has a per-row race context (`go_race`); a selection-level race
  context would let these suites select rows instead.
- **Selection nesting stays.** Six selections contain other selections. They
  compose lists at plan time and create no nested execution, so expanding them
  would duplicate member lists without removing a hazard.

Segments: the cache test and `public_wrapper_smoke` each ran alone through
their new rows and passed (`p51-cs`). The catalog's own test rows, which hold
the new shell-coverage test, pass (`p51-ea`, `p51-ea2`).

### 14.7 Declared Environment (P5-4, planned)

Status: planned 2026-11-26 at the owner's request. D21 to D23 approved as recommended on 2026-11-26; P5-4a and P5-4b done (14.8, 14.9).

**Problem (G45).** A unit's environment starts as a copy of the graph
process's environment. That copy holds the user's shell (72 variables on the
reference host) and up to 127 names Make exports. The planner then clears the
public inputs, and the runtime adds its own variables. Harness code reads
these names:

| Role | Names | References |
| --- | ---: | ---: |
| Harness context (run identity, results root, lease, suite runtime, fixture identity, failure context) | 52 | 2,063 |
| Fixture and service bindings (PostgreSQL, object store, browser stacks, test services) | 105 | 801 |
| Product configuration (`CARTULARY__*`) | 97 | 486 |
| Mode switches (graph child, skip prerequisites, suppress child success, preparation policy, build-cache switches) | 9 | 418 |
| Not yet classified (step runtime, readiness, layout, visual renderer, dev stack, release packaging) | 337 | 1,718 |

Consequences already seen in this phase:

- The Go test environment leaked into nested builds (G43).
- The owner keeps 342 `env -u` entries by hand.
- Each nested launch site kept its own strip list; P5-0 replaced them with
  an overlay record, which is one more variable.
- Mode switches make one Make target two programs.
- An unread or misspelled name does nothing and reports nothing.

The specification already closes public inputs: undeclared public inputs must
not reach children (`child_forwarding`). P5-4 extends that rule to every
harness variable.

**Target.** Environment variables remain at three boundaries only:

1. Public inputs of `make <target>`, already declared.
2. Variables that external tools read themselves (`PATH`, `HOME`, `LANG`,
   `GOCACHE`, `GOFLAGS`, `PLAYWRIGHT_BROWSERS_PATH`, `NODE_OPTIONS`).
3. Product configuration (`CARTULARY__*`) for product processes that tests
   start.

Everything the harness passes to its own code travels in one typed document
per unit.

**Workstreams.**

| Step | Work | Exit |
| --- | --- | --- |
| P5-4a | Environment registry. An authored `tools/harness_environment_registry.json` lists every name harness Go, Node and shell code reads or writes, with its role, writer and scope (top level, unit, nested). A source check rejects a read of an unregistered name and a registered name that nothing reads. | Every name registered or deleted, including the 337 not yet classified; the source check runs in `check` |
| P5-4b | Default-deny unit environment. The runtime builds each unit's environment from the registry's unit-scope names, the runner's declared tool variables and its own bindings, never by copying its parent. Nested launches build from the registry's top-level scope. | The 342 `env -u` entries, the remaining strip lists and `CARTULARY_HARNESS_UNIT_OVERLAY` removed; each runner's rows pass on segments, then one `check` |
| P5-4c | Single-writer ownership (redefined 2026-11-26, 14.10). Every internal harness variable has one owning component that writes it; only the planner and runtime write what the test population sees; components that start their own processes use their own names and build their children's environment explicitly. The registry declares components and writers, and its check rejects write sites outside the owner | Every invocation-context, step-runtime and fixture-binding name has one writer; the check runs in `check` |
| P5-4d | Single owners (redefined 2026-11-26, 14.14). Each remaining name has one owning component and an owner function; the registry's `owners` section declares them and its check rejects any other write. Product processes keep receiving `CARTULARY__*` configuration. | Every multi-writer invocation-context, fixture-binding and step-runtime name has an owner |

The nine mode switches leave with P5-3, when no unit re-enters Make.

**Order and validation.** P5-4a comes first because it measures and freezes
the surface. P5-4b is small and removes the leak class. P5-4c and P5-4d are
wide but mechanical. Every slice is proven on segments first: the rows of each
runner it touches, then one `check` at the end of P5-4b and of P5-4c. `ci`
folds into P4-7's final validation.

**Risks.**

- A unit that silently relies on an inherited variable fails under
  default-deny. The registry's source check finds static reads. A diagnostic
  mode that records each name the runtime withheld finds dynamic ones, such as
  names built at run time or read by child tools.
- Tools that read their own variables (`go`, `node`, `pnpm`, Playwright,
  Docker) need those names declared per runner.
- Until P5-3, operation units re-enter Make, so Make's exported variables
  remain a declared scope.

**Specification.** Amend the public-input forwarding rule (Section 5.5) to
cover every harness name, the Go row runner's binding (REQ-872) and the launch
context. Add the `cartulary.harness_unit_context.v1` schema and the runner
registry's tool-variable declarations.

### 14.8 Environment Registry (P5-4a)

- **Registry.** `tools/harness_environment_registry.json` registers 480
  `CARTULARY_*` names in eight roles; the task surface's 8 declared public
  inputs cover the rest of the 488 names harness code mentions.

  | Role | Names |
  | --- | ---: |
  | `fixture_binding` | 186 |
  | `test_internal` | 85 |
  | `invocation_context` | 49 |
  | `tool_option` | 43 |
  | `step_runtime` | 35 |
  | `location` | 34 |
  | `product_interface` | 28 |
  | `mode_switch` | 20 |

  The registry names its harness and excluded path scopes; every other
  admitted path is product code. Roles are a first classification from name
  families and where each name is mentioned; P5-4b refines them when the
  runtime first uses them.
- **Check.** `harness.environment_registry_check` fails on an unregistered
  name, a stale registration, a name in two roles, a declared input registered
  again, or a product boundary mismatch: a name both sides mention must be a
  `product_interface`, and only such a name. It runs in `check`, `ci` and
  `make source-check` (`CHECK=surface`, `CHECK=all`); its tests have their own
  row (`harness.command_surface.behavior.environment_registry`). Specification
  amended (Section 5.5).
- **Deleted.** `CARTULARY_HARNESS_GRAPH_ARTIFACT_CHILD` was set by the review
  recipes and stripped by 38 operations, but its only reader was a Make macro
  the generated surface never emits. The switch, the macro and the 38 strip
  entries are gone.
- **Boundary findings.** The run-evidence package
  `internal/platform/harnessruntime` is harness code compiled only into the
  harness server profile; it now sits in harness scope. The product reads
  three harness test hooks (`ENABLE_TEST_ROUTES`, `TEST_RUNTIME_MARKER`,
  `TEST_ROUTE_TOKEN`); the registry marks them `product_interface`, so the
  dependency is visible.
- **Left for P5-4b.** 101 names are mentioned exactly once. Most are knobs read
  with a default or child markers defined as constants, but each is a
  candidate for deletion when P5-4b decides which names reach units.

### 14.9 Default-Deny Unit Environment (P5-4b)

- **Change.** A unit inherits only declared names from the environment its graph
  starts in. The planner derives the set and passes it with the launch state;
  the runtime filters before it clears public inputs and adds forwarded inputs,
  bindings and its own variables. The set is:
  - registered names outside `product_interface` and `test_internal`, so an
    operator's shell cannot inject product configuration or test markers;
  - the registry's new `inherited` process and tool names (44) and prefixes
    (`LC_`, `TESTCONTAINERS_`);
  - the task surface's global inputs;
  - the Make variables operation executions name.

  Specification amended (Section 5.5).
- **Measured.** A unit captured from a live run had 164 variables before and
  100 after. Of the 72 shell variables, 8 still reach it (`HOME`, `LANG`,
  `LOGNAME`, `PATH`, `SHELL`, `TERM`, `USER`, `XDG_RUNTIME_DIR`). The 64
  withheld include the coding agent's variables, among them a messaging token
  and the user's email address, which every test process had received.
- **Defects found on segments.**
  - The first attempt failed every run within 2 s. The operation runner
    resolves the Make variables an operation's execution names from its
    environment, and a recipe's bindings do not cover all of them. They are
    now inherited, derived from the same owner data the Make renderer exports.
  - The planner row caught a P5-1 regression: explicit-only `harness.release`
    rows are release package rows, so the release task-surface row moved to
    `harness.command_surface`.
- **Strip entries.** The 38 `-u CARTULARY_MAKE_INPUT_SOURCES` entries were dead:
  the plan clears that name before any unit sees it. They are removed.
- **Plan correction.** 14.7 expected P5-4b to remove all strip entries and the
  P5-0 overlay record. Both stay:
  - The 226 remaining entries strip the runtime's own mode switches
    (`GRAPH_CHILD`, `SKIP_PREREQUISITES`, `SUPPRESS_CHILD_SUCCESS`) and
    inherited declared names (`CARTULARY_TEST_TARGET`, `VERBOSE`,
    `CI_VERBOSE`) before a unit runs `make` again.
  - A nested `make` launched from a Go test still starts from that test's
    process, so the overlay record keeps its `CGO_ENABLED=0` context out of
    nested builds.

  Both leave with P5-3, when no unit re-enters Make.
- **Validation.** Segments across every runner (Go, PostgreSQL-backed Go, the
  installed renderer, Vitest, a browser group, nested graphs, shell operations)
  pass. Diagnostic `check` `p57-check01` passes, 1,894 rows, graph 576 s.

### 14.10 P5-4c Redefined: Single-Writer Ownership

Status: redefined by the owner on 2026-11-26; replaces the typed context document.

- **Why.** The context document assumed each context variable had one writer,
  the runtime's process launcher. A write-site scan found 89 harness variables
  written by more than one component. The admission lease alone is written by
  the runtime, the fixture suite, the UI-review driver and the services
  session, and every child reads it as "the lease I'm under". A launcher that
  forgets to set its own value silently passes on a grandparent's. The
  suite-runtime handles belong to the services session but are also written by
  browser, readiness and review code. Changing how values travel would not fix
  that; giving each name one writer does.
- **Rule.** Every internal harness variable has one owning component, and only
  that component writes it. Only the planner and runtime write what the test
  population sees. A component that starts its own processes uses its own
  names, which units never inherit, and builds its children's environment
  explicitly instead of copying its own (default-deny at every launcher). Public
  inputs and product configuration stay writable by any launcher, because
  setting a command's inputs or a product's configuration is what launching it
  means.
- **Mechanism.** The registry declares components by path and a writer for each
  owned name; its check rejects a write site outside the owner. Tests are
  exempt, since they fake contexts on purpose.
- **Order.** Invocation context (13 multi-writer names), then step runtime (14),
  then fixture bindings (28, with P5-4d). Mode switches leave with P5-3.

| Multi-writer names by role | Names | Plan |
| --- | ---: | --- |
| Fixture bindings | 28 | One owner each |
| Product interface | 20 | Any launcher; a single product-launch helper later |
| Step runtime | 14 | One owner each |
| Invocation context | 13 | One owner each, first |
| Mode switches | 8 | Removed by P5-3 |
| Locations and public inputs | 6 | Any launcher |

### 14.11 Explicit Child Environments (P5-4c, slice 1)

- **What the write sites showed.** Most of the 89 multi-writer names are not
  rival owners. They are launchers setting a child command's inputs. The
  services session gives its start and terminate commands its own lease, and
  browser acquisition gives its cleanup the suite-runtime handle. Each sets
  the value explicitly, with one meaning. What pollutes is pass-through:
  `{ ...process.env, ... }` hands a child everything the parent had, so a
  value another component set reaches processes it was never meant for. The
  enforceable form of the single-writer rule is therefore: a component that
  starts processes for another component passes on only process and tool
  basics, and forwards or sets every harness variable explicitly.
- **Change.** `childEnvironment` (`tools/harness/runtime/child-environment.mjs`)
  builds those environments. It passes the registry's inherited basics and the
  global inputs, the names the launcher forwards (`pass`) and the values it
  sets (`set`). Libraries that read harness variables from a given environment
  declare them: `runtimeEnvironment` (the session client and native session
  command) and `launchEnvironment` (launch contexts). The services session,
  frontend publication and browser acquisition use it, and the registry's new
  `explicit_launchers` scope makes the check reject any pass-through in their
  files. The browser stack's own two spawns run inside the unit and carry an
  explicit marker. Specification amended.
- **A collision this removes.** A services session is a standalone suite
  command and should request the full service scope. Started from inside a
  unit, it inherited the unit's narrower service selection
  (`CARTULARY_HARNESS_SERVICE_DEPENDENCIES`) by pass-through.
- **Hidden dependencies the segments exposed.**
  - Browser stack retirement needs the group's object-store binding (the
    `S3TEST` names and the service selection its suite hash covers). It now
    forwards that binding by name, and the stack no longer fails to clean up.
  - Recovery records need the session client's run placement and recovery
    attempt. That interface is now declared as `runtimeEnvironment`.
- **Remaining.** Slice 2: the UI-review recipes and native client. Then the Go
  launchers outside the runtime (`harnessaccount`, `harnessreview`,
  `processtest`, `recoverybrowsertest`, `webstacklisten`). Then the step-runtime
  and fixture-binding owners, with P5-4d. Top-level tools that hand the
  operator's environment to `docker`, `go` or `package.sh` are left as they
  are, because they do not feed the test population.

### 14.12 Explicit Child Environments (P5-4c, slice 2)

- **Change.** The UI-review chain runs `make ui-review`, then the native
  client, the Go review driver, the Node adapter, the recipes and finally the
  operation runner. Its two Node launchers copied their parent's environment
  and deleted a regular expression of name families. That is default-allow:
  anything the expression missed, such as shell, agent and editor variables,
  crossed, and a Go test's `GOROOT` needed its own deletion. Both now use
  `childEnvironment` and forward one declared set, `reviewForwarded`: the Make
  variables operation executions name, minus public target inputs, plus
  `PLAYWRIGHT_BROWSERS_PATH`. The deletion expressions and the `GOROOT` special
  case are gone. The recipe environment is passed as `recipe`, because it is
  an explicit object, not the process environment.
- **Found while converting.** Operation executions name a public review input
  (`REVIEW_PROFILE`), so the operation-variable set alone would have let it
  cross. Public inputs reach the native review owner only through its request,
  so the forwarded set excludes them.
- **Validation.** All eight UI-review rows pass, including both seeded rows,
  which drive the whole chain against a live browser stack.
- **Remaining.** The Go launchers outside the runtime (`harnessaccount`,
  `harnessreview`, `processtest`, `recoverybrowsertest`, `webstacklisten`).
  Then single owners for the step-runtime and fixture-binding names, with
  P5-4d.

### 14.13 Explicit Child Environments (P5-4c, slice 3)

- **Classification.** Five Go sites copied `os.Environ()` into a child.
  - Three are correct as they are. `harnessaccount` wraps the user's public
    command for accounting, so the command must see its full environment.
    `webstacklisten` wraps the browser stack's backend inside the unit's own
    fixture. `harnessreview` copies an environment the native client already
    made explicit in slice 2.
  - Two are outside riders that start product processes from Go tests:
    `processtest`, which starts the server, and the recovery browser test's
    operator launcher. The server received the whole test-unit environment:
    harness context, fixture bindings, the Go build context and the test master
    key the runtime binds into Go units. The operator received every
    non-`CARTULARY_` variable, including the operator's shell and agent
    settings.
- **Change.** `internal/testutil/childenv` is the Go counterpart of
  `childEnvironment` and reads the same registry and task-surface
  declarations. Both launchers use it. Their configuration was already
  explicit (`configtest` sets the master key and secrets), so nothing they
  needed came by inheritance. The registry check now also flags
  `append(os.Environ(), ...)`, `range os.Environ()` and `Env = os.Environ()` in
  Go launcher files; seven files are in the launcher scope. The new package's
  test has its own row (`harness.command_surface.unit.child_environment`).
  Specification amended.
- **Validation.** Every row that starts a server or operator through these
  launchers passes: 18 `app.server` rows, the recovery process rows and the
  three restore browser rows, and the extensions and network-flow
  packaged-server rows. Staticcheck caught a helper the change left unused
  (`envPairs`); it is removed.
- **Remaining in P5-4c.** Single owners for the step-runtime and
  fixture-binding names, taken with P5-4d.


### 14.14 Single Writers (P5-4d, slice 1)

Status: P5-4d redefined by the owner on 2026-11-26. Instead of moving fixture
bindings into a typed document (14.7), it gives each name one writer, the rule
of 14.10.

- **Mechanism.** The registry gains `owners`. Each component names its paths
  and the names that only it writes. The registry check rejects any write of
  an owned name outside the owner's paths in Node or Go code that is not a
  test. A write is:
  - a bracket, property or map-key assignment;
  - a `Setenv`;
  - a `NAME=value` literal;
  - in Go, any of these through a constant declared for the name.

  A read, a switch case or forwarding by name is not a write. Shell is not
  checked, because its writes are prefixes or exports scoped to one command.
  Specification amended.
- **Owners.**

  | Component | Owned names | Writer functions | Paths |
  | --- | --- | --- | --- |
  | `runtime_session` | Suite-runtime root, lease and run ID (3) | `runtimeHandleEnvironment` | Session client, `harnesssession` |
  | `suite_services` | Suite ID, call mode, lifecycle mode, persistent borrower (4) | `suiteAttachmentEnvironment`, `suiteservices.Attachment` | Services session, `suiteservices`, `testservices` |
  | `runtime_binaries` | The seven runtime binary paths (7) | `runtimeBinaryEnvironment`, the planner's launch bindings | Runtime binary registry, `harnessplan`, `Makefile` |

- **What moved.**
  - These callers now call the owner functions instead of spelling the names:
    the services session's start, attach and terminate; browser acquisition's
    two borrows and its cleanup; frontend publication; and the UI-review
    recipes.
  - In Go, the managed-suite fixture and its recovery set the suite ID through
    `suiteservices.Attachment`. `testservices` uses it for its three owned
    starts.
- **A second writer removed.** The Go review server filled in its own defaults
  for four of the seven binary paths. That was a copy of the planner's
  bindings, and it had no operator, server or sourcecheck path. The native
  review client launches the review server. It now sets all seven paths
  through `runtimeBinaryEnvironment`, and the review server only reads them.
- **Validation.** With the owners declared, the check first found two
  writers: the managed-suite fixture's suite ID and the review server's
  defaults. Both are converted, and the check now passes with no findings.
  The owner scan adds about 0.1 to 0.4 s to the check's 3 to 4 s. Every
  segment passes:
  - the registry and child-environment rows, with two new owner tests;
  - the `suiteservices` helper row, with a new `Attachment` test, and current
    epoch evidence;
  - the native fixture, native review, test-services lifecycle and broker
    smoke rows;
  - local session, UI-review lifecycle and execution, artifact assurance and
    the seeded default review;
  - the incidents browser row.
- **Remaining.** By the 2026-11-26 write-site scan, these names still have
  more than one writer and no owner:
  - 11 invocation-context names, among them run ID, target, work unit,
    admission lease and service selection;
  - 20 fixture bindings: the browser stack, PostgreSQL and object-store
    allocation, and service-binding leases;
  - the 14 step-runtime names.

### 14.15 Single Writers: Invocation Context (P5-4d, slice 2)

- **What the write sites showed.** The 11 invocation-context names had 70
  write sites in Node and Go. Almost all were launchers giving a child command
  its context, setting each name by hand:
  - which run it is and where results go;
  - which target and work unit it reports for;
  - which admission lease it runs under;
  - where its inputs came from.

  Two names already had one writer: the runner script and the runtime's
  failure context.
- **Owner.** The new `run_context` owner holds the 10 context names plus the
  results root, which every one of those sites set alongside the run ID.
  - The Node writer is `tools/harness/contract/run-context.mjs`. The Go writer
    is `internal/testutil/runcontext`, a package with no dependencies, so the
    planner can use it without importing test support.
  - `Context` writes only the fields that are set, because an empty identity
    or lease is never meaningful.
  - The service selection and input sources always write, because an empty
    list means "none" rather than "inherit". The planner and managed suite
    rely on that difference.
  - The generated Make entry point keeps writing the names for top-level
    commands. Its renderer is in the owner's paths.

  The runner script and the failure context each get their existing writer as
  owner (`runner_context`, `runtime_failure_scope`).
- **What moved.** Every other writer now calls the owner:
  - Go: the planner, runtime, fixtures, review driver, session, stress,
    accounting and `testservices`;
  - Node: the services session, browser acquisition, frontend publication,
    review recipes and adapter, operation runner, runner CLI, finalizer,
    evidence and release CLIs, and the renderer build.

  Two false positives became simpler code:
  - the contract's two-entry defaults object is now one constant;
  - the tool-run summary records its settings from a list of names.
- **Found by the segments.**
  - The `routing_contract` row had been failing since P5-1. It runs only in
    `ci`, so the P5-1 segments did not include it. P5-1 gave the command
    surface 15 explicit-only rows, but a guide test still assumed the owner
    had none and therefore planned nothing. The test now uses an
    explicit-free owner from the catalog for that property.
  - No explicit command binds those 15 rows, so they run only through
    `test-slice`. Several of them are nested real-target runs, which are P5-2
    material.
  - `native_fixture_lifecycle` failed once because I edited a tracked file
    while the segment ran. Its fixture compares the prepared frontend's source
    digest with the tree's current digest. It passes on rerun.
- **Validation.** The registry check has no findings. Every segment passes:
  - 14 command-surface rows: the planner, native runtime, stress, scheduler,
    registry parity, the release task surface, Node bootstrap, pinned tools,
    three real-output cases and the registry and context rows;
  - 7 evidence-accounting rows, including both invocation-accounting rows;
  - 3 release rows;
  - the routing contract;
  - 15 native fixture, review and test-services rows;
  - the browser review chain with the seeded review;
  - the installed renderer row;
  - the incidents browser row.

  The renderer installation was reused, so the renderer build's changed line
  did not run.
- **Remaining.** By the same scan, 20 fixture bindings (the browser stack,
  PostgreSQL and object-store allocation, and service-binding leases) and the
  14 step-runtime names still have more than one writer.

### 14.16 Single Writers: Fixture Bindings (P5-4d, slice 3)

- **What the 20 names turned out to be.** The scan's count of fixture bindings
  with more than one writer overcounted, in three ways:
  - **Product secrets.** Seven names (the primary object-store endpoint,
    credentials, bucket and TLS switch, the primary runtime DSN and a conflict
    token) are product secrets. The server reads them by a computed name, and
    the browser stack and dev stack each set them for the server they start.
    Like product configuration, any launcher may set them. Their
    `fixture_binding` role is a misfiling, because product code computes
    their names instead of naming them.
  - **Shell reads.** Five names had only one Node or Go writer; the scan had
    counted shell reads with defaults (`${NAME:-}`) as writes. These are the
    cohort label, acquisition file, session lease file, services binding
    lease and Playwright worker offset.
  - **Real second writers.** Eight names did have several: the browser stage,
    runtime profile, service requirement, session group and reset drain
    timeout; the resource dependencies; and the PostgreSQL isolation and
    allocation reference.
- **Owners.**

  | Component | Owned names | Writer functions | Paths |
  | --- | --- | --- | --- |
  | `browser_binding` (new) | The browser stage, runtime and resource profiles, service requirement, session contract and group, stack key, preparation label and timeout, reset drain timeout, cohort label, acquisition file, session lease file and private session root (14) | `browserBindingEnvironment`, `browserbinding.Binding` | `browser-binding.mjs`, `internal/testutil/browserbinding` |
  | `run_context` | Adds the resource dependencies, an always-written JSON list | `resourceDependenciesEnvironment`, `runcontext.ResourceDependencies` | As before |
  | `suite_services` | Adds the PostgreSQL isolation and allocation reference, always written together | `postgresAllocationEnvironment`, `suiteservices.PostgresAllocation` | As before |

  The browser binding owns the single-writer browser names too. The planner
  and fixture write them in the same statements as the contested ones, so
  each site now builds one binding. `browserbinding` has no dependencies, so
  the planner can use it.
- **Write semantics.**
  - **Browser binding: only set fields.** Every reader treats an empty
    binding value as unset. All 47 batch partitions set every planner field.
  - **Resource dependencies and allocation: always written.** The old code
    wrote empty values there to clear inherited ones.
- **What moved.**
  - **Planner.** Its browser units and fixture-snapshot builders.
  - **Browser fixture.** Its stack environment, cohort label, acquisition
    file, stack attachment and reset.
  - **Six resource-dependency sites.** Attachment, proxy, renderer, managed
    suite (two) and work; the fixtures' private JSON helper is gone.
  - **Elsewhere.** The runtime's empty list, the review driver's profile,
    the UI-review recipes, the browser cohort CLI and the services session's
    allocation.
- **Planner equivalence.** I planned `ci` and `release-measure` with and
  without the planner change. Every unit's semantic digest and both graph
  digests match. The semantic digest covers the unit's command environment.
  The plans cover 27 `ci` browser groups and 33 measurement groups with their
  4 fixture-snapshot builders. Fixture-profile cohorts run only in
  measurement, which is too heavy for a segment.
- **Validation.** The registry check has no findings. Every segment passes:
  - the browser-binding, context, registry and planner rows;
  - the `suiteservices` helper row, with a new allocation test;
  - 16 native fixture, review and test-services rows, including the
    object-store namespace lifecycle;
  - the browser review chain with both seeded reviews;
  - the object-store proxy's streaming row;
  - a transaction and a dedicated PostgreSQL row;
  - the incidents browser row, which exercises the planner's browser
    binding, the stack fixture, and the renderer and work leases.
- **Remaining.**
  - The 14 step-runtime names.
  - Shell writers are still not checked. Inside the browser stack, shell
    scripts still set names for their own children; for example
    `start-web-e2e.sh` passes the session lease file to its Node helper.

### 14.17 Single Writers: Step Runtime and Shell (P5-4d, slice 4)

- **What the write sites showed.** Shell writes all 14 step-runtime names,
  so the check could not see any of them.
  - **The step protocol was written four times.** `step-runtime.sh` (twice),
    `run-vitest-step.sh` and `run-playwright-step.sh` each spelled out the
    same twelve summary names for the summary helper: label, directory,
    command, start and end, three durations, exit status, runner log and the
    two output logs.
  - **The timing bucket.** The browser stack set it as a prefix on its seven
    timed steps.
  - **The step artifact directory.** The Go runtime writes it for units and
    the step runtime for steps.
- **The check now reads shell and Make recipes.** In those files a write is:
  - an assignment or a command prefix;
  - an `export`, `local`, `readonly` or `declare`;
  - a `${NAME=...}` expansion.

  Comment lines, quoted text and test scripts are not checked. Once shell
  was checked, nine shell writes of names owned in slices 1 to 3 appeared:
  - **Entry points.** `admission.sh` creates the accounted invocation root,
    and `step-runtime.sh` resolves run placement for top-level steps. Both
    join the `run_context` owner's paths, like the Makefile.
  - **The browser stack.** `start-web-e2e.sh` and `playwright-owned-stack.sh`
    canonicalize the runtime profile they receive and pass it, with the
    session lease file, to their own processes. They join the
    `browser_binding` owner's paths, because the stack is where the binding
    takes effect.
  - **Launchers setting a child's target.**
    - `build-web-artifact.sh` now passes `--target` to the step runner.
    - `build-go-artifact.sh` re-set the target it had just read from its own
      environment, so it no longer sets it.
    - The generate drift check unsets the inherited target, so the nested
      `generate` recipe uses its own target, as it declares.
- **Owners.**
  - **`step_runtime` (new).** It owns the 12 summary names and the timing
    bucket, and its only path is `step-runtime.sh`.
  - **One summary writer.** `step_publish_summary` writes those names from
    named fields: a field that is not given is not set, and one given empty
    is set empty. All four writers now call it. Names specific to one step
    kind, such as the Vitest request or the Playwright output directory, are
    single-writer and still reach the helper from the caller.
  - **New step-runner options.** `run_step_command` and `run-step.sh` accept
    `--target` and `--timing-bucket`. They set them as function-local exports,
    so nested steps still inherit the enclosing step's bucket and the values
    are restored on return.
  - **Artifact directory.** It joins `run_context`: `runcontext.Context`
    gains `ArtifactDir`, and the Node form gains `artifactDir`.
- **Validation.** The registry check has no findings, and its tests cover
  shell writes. Every segment passes:
  - `make generate`, which runs the changed step runtime at top level;
  - the build-artifact, generate-drift, capacity-failure, release-task-surface,
    registry-parity and native-runtime rows;
  - the Vitest, Playwright and shell step accounting rows;
  - artifact assurance, the seeded review and native fixture lifecycle;
  - the incidents browser row;
  - script lint and the full source-check, whose drift check runs the nested
    `generate` with the target unset.
- **Remaining.** Every multi-writer invocation-context, fixture-binding and
  step-runtime name now has one owner, which meets the P5-4d exit criterion.
  The check does not see:
  - a quoted shell assignment such as `env "NAME=value"`;
  - a write through a variable key.

  The seven product secrets keep a misfiled `fixture_binding` role.
- **Close-out check.** `p66-check01` passed all 1,897 rows. It failed one
  operation: ShellCheck found the `target` variable this slice left unused in
  `build-go-artifact.sh`. The script now only requires the target. The shell
  lint and build-artifact rows pass, and the full `check` was not repeated.
  The graph took 617 s against 577 s at P5-4b, but unit work fell by 121 s
  and the same units finish last. Section 14.18 traces the 40 s to a
  stranded PostgreSQL generation, not to P5-4.

### 14.18 P5-4 Timing: A Stranded Generation

- **Question.** `p66-check01` took 617 s against 577 s at P5-4b
  (`p57-check01`). A second check, `p67-check02`, took 674 s.
- **Not P5-4.** Unit work fell in each run (7,559, 7,438 and 7,145 s), and so
  did PostgreSQL work (4,312, 4,048 and 3,920 s). The PostgreSQL lanes were
  busy less of the time:

  | Run | Running PostgreSQL units, steady state | Waiting units are blocked on |
  | --- | ---: | --- |
  | `p57-check01` | 8.0 | `postgres` and `postgres_generation`, 8 holders |
  | `p66-check01` | 7.0 | `postgres_generation` only, 7 holders |
  | `p67-check02` | 7.0 | `postgres_generation` only, 7 holders |

- **The holder.** Host-admission decisions in `p66` and `p67` show the same
  generation resident from t = 0 to the end of each run. The host-admission
  ledger holds three residents owned by the dead process of `p58-browser`, a
  P5-4c segment run:
  - an object store;
  - a PostgreSQL generation;
  - a browser stack.

  That run's cleanup failed at the `browser_namespace` stage, so its runtime
  was preserved for recovery. No container or process remains; only the
  ledger charge does. `make test-run-recovery-status` reports the run as
  `runtime_bound`.
- **Why startup recovery missed it.** Every close path fences detail producers
  first, and fencing writes the `finalizing` seal. That includes the failure
  path, `PreserveRecovery`, so a preserved runtime is sealed while it still
  holds resources. Startup recovery skips sealed roots because the spec says
  "a sealed root drained its resources before sealing". That is false for
  preserved runtimes. It also conflicts with the same paragraph's rule that a
  stopped or killed run never leaves admission capacity charged for later
  runs (G46).
- **Proposed repair (needs owner approval).** Amend the rule so startup
  recovery decides by resource records rather than by the seal:
  - a dead root with unresolved resource records is recovered, even when
    sealed;
  - a sealed root with none stays for exact recovery of its publication
    authority.

  Then run `make test-run-recover RESULTS_DIR=.cartulary/test-results/p58-browser`
  to release the stranded residents.
- **The `p67` tail.** `p67` rebuilt the report renderer (153 s) instead of
  reusing it (20 to 42 s in the runs before and after). Its installation key
  includes `build-go-artifact.sh`, which the ShellCheck fix changed. The chain
  that followed:
  - The server-harness build waits for the renderer, so it finished at 167 s
    rather than 27 to 49 s.
  - The server-process unit (`internal/app/serverprocess`, 138 s estimated)
    depends on that build.
  - Meanwhile the other PostgreSQL units aged in the queue, and 110
    lower-ranked units started first.
  - The server-process unit ran from 566 s to the end of the run.

  Rank aging adds a unit's wait to its rank, so a long unit that becomes
  eligible late falls behind short units that were eligible earlier.
- **Flake.** In `p67-check02`,
  `TestOrdinaryShutdownAC604_Process/job_ignores_cancellation` received WebSocket
  close status -1 instead of 1001. It then passed three isolated reruns.
  `wstest.RequireClose` does not print the underlying error, so the run cannot
  show whether the socket first saw a message, a reset or EOF.

### 14.19 Startup Recovery by Resource Records (G46)

- **Recovery.** `make test-run-recover RESULTS_DIR=.cartulary/test-results/p58-browser`
  (`p68-recover`) passed. The host-admission ledger now holds no residents,
  and the `p58-browser` runtime root is gone, so all eight PostgreSQL
  generation slots are free.
- **Spec amendment (owner-approved).** In REQ-855, resource records, not the
  finalization seal, now decide whether a dead root needs startup recovery:
  - A run whose cleanup fails seals its root before preserving the resources
    it could not release, so a seal does not prove those resources settled.
  - Recovering a sealed root releases its resources and keeps the sealed,
    resource-free root for exact recovery. A UI-review root can still hold
    terminal publication authority after its resources drain, and its own
    recovery path needs that root.
  - Sealed roots without resource records stay untouched, as before.
- **Implementation.**
  - `harnesssession.DeadRuntimes` offers a dead sealed root only when its
    validated resource journal is non-empty, and reports the seal. A seal
    with unexpected content fails closed.
  - Startup recovery settles a sealed root with `PreserveRecovery` instead of
    `Close`, so the root stays. Explicit `test-run-recover` still closes and
    removes it.
- **Validation.** All passed:
  - `make lint CHECK=go` (`p68-lintgo`);
  - `harness.evidence_accounting.behavior.native_session` (`p68-session`):
    the inspector offers a dead sealed root with a resource record and
    reports its seal. It omits drained and corruptly sealed roots, as well as
    live, busy and foreign ones;
  - `harness.browser.unit.native_fixtures` (`p68-fixtures`): startup recovery
    releases a sealed root's resource, reports it and keeps the seal. It
    leaves a drained sealed root alone;
  - `make lint-markdown` (`p68-md`).
- **Eight legacy roots now fail on every run.** Since the change, each
  top-level graph prints eight `[RECOVERY-FAIL]` lines. They come from the
  October runtimes in `~/.cache/cartulary/harness-scratch/suite-runtime`,
  which the seal rule used to hide. Current recovery cannot admit them:
  - **Six fail as unsafe.** Their resident records name the pre-generation
    ledger `/tmp/cartulary-host-admission-1000` and lack current claims such
    as `postgres_generation`.
  - **Two fail at `process_identity`.** Their helper processes were recorded
    under an earlier boot, which the harness treats as unresolved.

  They hold nothing physically. No ledger, legacy or current, holds their
  residents. No harness container remains, and every anonymous volume dates
  from 23 November or later. Their run directories were deleted, so explicit
  `test-run-recover` cannot settle them either. The owner approved removing
  them, and they were deleted on 2026-11-26. `p69-a11y`, the next graph run,
  reported no recovery failures.

### 14.20 Never-Run Rows

- **Question.** The owner asked to deprecate unused tests.
- **Method.** Every target and suite that runs a work graph was planned, and
  their row sets were compared with the catalog (`p69-plan-*`). That covers
  `check`, `ci`, `test-fast`, `release-check`, `release-measure`, the 11
  `harness-contract` suites, the five `browser-e2e` suites,
  `browser-e2e-visual-update` and `package-inspect CHECK=all`.
  `otel-conformance`, the SeaweedFS gates and `ui-review` run no work graph.
- **Result.** 12 of 2,778 rows were never planned, all `explicit_only`. None
  is dead: each is the only row for its test, and no compatibility code or
  test for retired formats remains.
  - **Four browser accessibility self-tests:** contrast, accessible names and
    presentation restoration (Vitest, `harness-node`), and the accessibility
    observation scenario (Playwright). The helpers they test are used by
    about 30 accessibility and visual specs each. These tests had not run
    since they were added on 2026-10-26, and they pass (`p69-a11y`).
  - **Three spec-required tools:** PostgreSQL transport qualification, reset
    characterization and delegated-cgroup invocation accounting. The harness
    spec defines each as an explicit capability.
  - **Five on-demand evidence rows:** the network-flow import scaling sweep,
    graph capacity certification, and the Entra, VERIS and Windows
    real-release imports. They have not been run in this review.
- **Decision (owner, 2026-11-26).** Promote the four self-tests to `full`,
  where their 23 siblings run, and keep the other eight as on-demand rows.
- **Change.** The four rows' placement is `full`. They join units `ci` already
  runs: the browser batch manifest already partitioned the observation
  scenario, so only the render index digest changed. A task-guide test used
  the contrast row as its explicit-only example; it now uses
  `visual_anchor_geometry`, which `browser-e2e SUITE=visual` plans.
- **Validation.** All passed: `make generate`, `source-check CHECK=catalog`
  and `CHECK=generated`, the `harness.test_catalog.behavior.routing_contract`
  row and `lint CHECK=scripts`. `ci` plans 2,658 rows (2,654 before) in the
  same 506 units, and `check` is unchanged at 1,897 rows. The rows were not
  run inside a full `ci`.

### 14.21 Child CPU Budget (P5-4 follow-up)

- **Defect.** Since P5-4c, a product process that a Go unit starts through
  `childenv.Build` received only the registry's `inherited` names. The plan
  binds `GOMAXPROCS` to a Go unit's CPU tokens, and that binding was dropped.
  The server or operator child therefore ran on every host CPU while its unit
  claimed only its tokens, against REQ-379: "Each child claims exactly its
  scheduler-owned `GOMAXPROCS` CPU budget."
- **Why not `inherited`.** The planner reads the same list to decide what a
  unit takes from its graph's environment. Adding `GOMAXPROCS` there would let
  an operator's shell value reach every non-Go unit.
- **Repair.** The registry gains `budget`, which lists `GOMAXPROCS`:
  - A unit never takes a budget name from its graph's environment.
  - Both child builders, Go `childenv.Build` and Node `childEnvironment`,
    keep the parent's value, so a process a unit starts runs within the
    unit's CPU claim.
  - The registry check validates the list and rejects a budget name that is
    also inherited. Specification amended (Section 5.5).
- **Tests.** The child-environment tests in Go and Node expect the budget to
  pass and the Go build context not to. The planner's preparation test asserts
  that a real plan never inherits `GOMAXPROCS`. The registry tests cover the
  new shape rules.
- **A second P5-4c regression.** The segment failed
  `harness.command_surface.process.processtest_helper_contract`, a `ci`-only
  row:
  - Its tests set the helper mode with `t.Setenv` and relied on the started
    process inheriting it.
  - Under P5-4c's explicit child environments the marker is withheld, so the
    child test binary ran the whole suite instead of serving, and two tests
    timed out.
  - Every launch now passes its helper mode explicitly.

  No segment after P5-4c ran this row, because `check` does not select it.
- **`ci`-only launcher rows.** Eight `full` rows start processes through
  `childenv` and had not run since P5-4c: this row, four `app.server` process
  smokes, the packaged network-flow server and two recovery-browser rows.
  All eight pass with the repair (`p72-cs`, `p72-server`, `p72-nf`,
  `p72-browser`). One of the recovery-browser rows runs in-process lifecycle
  tests and starts no child.
- **Validation.** All passed:
  - `make generate`, which changed nothing (`p71-generate`);
  - `lint CHECK=go` and `lint CHECK=scripts` (`p71-lintgo`,
    `p71-lintscripts`, `p72-lintgo`);
  - the child-environment, registry and planner rows (`p71-cs`);
  - the AC-604 shutdown row (`p71-server`);
  - `source-check CHECK=surface` and `CHECK=generated` (`p71-surface`,
    `p71-generated`);
  - `lint-markdown` (`p72-md`).

  The repaired `processtest` row passes (`p72-cs`). The rest of `ci` has not
  run since P5-4b, so other `ci`-only rows could hide the same inherited-marker
  pattern. The eight rows above are the ones that start processes through
  `childenv`.

### 14.22 Unnesting (P5-2)

Decision D20 (owner, 2026-11-26): nested harness probes run in explicit
`harness-contract` presets, outside `ci`.

- **What nests.** The last retained `ci` with diagnostics (`p46-ci01`) holds
  nested run manifests from exactly six rows, and `p67-check02` holds none:

  | Row | Nested graphs it starts |
  | --- | --- |
  | `native_graph_observations` | two `test-slice` |
  | `native_nested_cancellation` | three `test-slice` |
  | `native_selected_service_lifecycle` | four `harness-contract` |
  | `frontend_artifact_lifetime` | two concurrent `make build` |
  | `real_output_owner_json` | one `test-slice` |
  | `real_output_source_schema` | one `source-check` |

  Other tests call `make` read-only (status, `target-plan`, `-q`, help, retired
  targets) or in scratch fixture repositories, and the physical-fixture
  lifecycle tests start fixtures without a graph. They stay where they are.
  `real_output_invalid_usage` starts no graph and stays in `ci`.
- **Moved.** The six rows are `explicit_only`. Each was already a member of an
  explicit preset, but a selection drops explicit-only rows unless the owner
  lists it in `explicit_selections`. The owner now lists `runtime-boundary`,
  `artifact-consumers`, `execution`, `regression` and `full`. Those five
  presets select no other explicit-only rows, and the presets plan the rows
  they planned before (5, 2, 10, 40 and 57). `ci` plans 2,652 rows in 500
  units (2,658 in 506), and `check` is unchanged.
- **Guard.** The planned source check would have relied on text patterns: a
  scan of every row's sources found ten hits that are messages, scratch
  Makefiles or read-only calls, and a Playwright fixture that no file-level
  rule can see. The runtime decides exactly instead. A graph whose process
  ancestry holds an aggregate graph (`cartulary-harness graph --selection
  aggregate`) working in the same repository stops before planning with
  `config/configuration_error` at the new stage `nested_aggregate` and prints
  `[NESTED-REFUSED] enclosing_target=<target>`. It reads the process table,
  not the environment, because nested launchers clear their children's
  environments. A graph under `test-slice`, an explicit preset or
  `agent-finalize`, or in a scratch repository, is unaffected. Only the
  runtime binary calls `Run`, so no test graph runs in-process. Specification
  amended (the `harness-contract` presets and the startup stage table).
- **Trade-off.** A row that nests but is placed in a daily lane passes its own
  segment and fails in `check` or `ci`, with the refusal line naming the cause.
- **Validation.** All passed (`p74-*`):
  - `make generate`, `lint CHECK=go`, and `source-check CHECK=catalog` and
    `CHECK=generated`;
  - the runtime row, with the new ancestry tests: a fixed process table covers
    `ci`, `check`, an explicit suite, another repository, another program, a
    top-level graph, an exited ancestor and a cyclic ancestry, and a fake
    `/proc` covers the parser;
  - a stand-in process with the argument vector of a `ci` aggregate graph,
    which runs a real `make test-slice`: the nested graph stops with exit 2,
    `[NESTED-REFUSED] enclosing_target=ci` and stage `nested_aggregate`
    (`p74-refused`). The same stand-in naming `harness-contract` lets it run
    (`p74-allowed`);
  - the six moved rows under `test-slice`, whose nested graphs start
    (`p74-nested`, `p74-realout`, `p74-lifetime`);
  - `lint-markdown`.

  No aggregate run was repeated. Retained evidence shows no nested graph in
  `check`, and in `ci` only from the six moved rows.

### 14.23 Environment Tidy (P5-4 follow-up)

- **What remained.** A write-site scan after P5-4d and the budget repair found
  no owner for names that more than one component writes. Apart from the five
  mode switches that P5-3 removes, they fell into four groups.
- **Product configuration under derived names.** Fourteen registered names are
  configuration that the product reads under a name it derives from a
  reference, such as `CARTULARY_S3_<ref>_ENDPOINT` from an object-store
  service reference or `CARTULARY_SECRET_<name>` from a key-ring secret:
  - the primary migration and runtime DSNs;
  - six primary object-store names;
  - six secrets.

  Product code never spells them, so the check could not file them as
  `product_interface`, and they sat in `fixture_binding`. They now have the
  role `product_derived`. Like a product interface, any launcher may set one
  for the product process it starts, and no unit inherits one from its
  graph's environment. The check rejects a `product_derived` name that product
  code spells. Specification amended (Section 5.5).
- **Performance-fixture identity.** The planner, the Go browser fixture and the
  session-evidence module each wrote a cohort's snapshot names:
  - the profile, snapshot key, builder unit, row and predicate;
  - the clone lease and ordinal.

  They join the `browser_binding` owner as seven fields of its Go `Binding`
  and Node `browserBindingEnvironment`, which write only set values. Every
  reader trims and treats an empty value as unset, and a populated fixture
  always has a profile, key, builder and exactly one predicate. So the
  environments the planner writes are unchanged.
- **The cleanup marker becomes an argument.** Two launchers told
  `cartulary-test-services cleanup-web-e2e` through
  `CARTULARY_FIXTURE_PROCESS_CLEANUP_COMPLETE` whether the browser processes
  using a fixture had stopped, and the service set the same marker in its own
  copy of the environment. The command now takes a required
  `--process-cleanup complete|incomplete` and passes it as a parameter to the
  lease cleanup, and the name is gone.
- **Constant Playwright worker names are deleted.**
  `CARTULARY_PLAYWRIGHT_WORKER_COUNT` and `_INDEX_OFFSET` were written by two
  launchers. Every reader only checked that they were `1` and `0`, and the
  configuration already fixes one worker. The names, their readers and five
  Vitest cases that tested those readers are removed, and the configuration
  and parallel-index checks remain. The registry holds 477 names (480 before).
- **The check's blind spots.** The registry check now also counts as a write:
  - a quoted shell assignment word, such as `env "NAME=value"` or an array
    element, unless the line only prints it;
  - a Go or Node loop over a literal list of names that assigns through its
    loop variable.

  The loop rule found one real violation: the browser fixture cleared the
  input-sources name, which `run_context` owns, through a loop. It now calls
  `runcontext.InputSources`. Specification amended.
- **Validation.** All passed (`p75-*`):
  - `make generate`, `make format`;
  - `lint` for Go, scripts, shell, types and frontend;
  - `source-check` for surface (the registry check, with no findings),
    catalog, identity and generated;
  - the registry, browser-binding, planner and owned-stack lifecycle rows;
  - the test-services lifecycle contract, with the new argument rules;
  - the performance-fixture lifecycle unit and integration rows;
  - the trimmed Vitest fixture suite and the UI-review lifecycle, which runs
    the browser acquisition's retirement call;
  - the incidents browser row, which stops a real stack through the new
    cleanup argument;
  - `lint-markdown`.

  Planner equivalence: `release-measure` planned on the committed code and on
  this change has the same graph digest and the same semantic digest for all
  79 units, 62 of them fixture cohorts or builders.
- **Remaining.**
  - The five multi-writer mode switches, which leave with P5-3.
  - 183 names appear in only one file and 85 only in tests. Deleting them is
    a separate decision.

### 14.24 Thin Make (P5-3, planned)

Status: planned 2026-11-26 at the owner's request. Owner direction: retire as
much of the Makefile as needed; a thin wrapper is better than a complicated
one. Approved with D24 to D28 as recommended on 2026-11-26; P5-3a done
(14.25).

**Problem (G44, G47).** Measured at `db2304ea9`:

- **Make runs before the graph four times.** One `make check` parses the
  Makefile once to take the workspace lock and re-execute itself. Under the
  lock it parses again for the target, then again for a nested `make` of the
  Node runtime and again for the harness binary. Each artifact unit in the
  graph parses it once more.
- **Most public targets do their work inside Make.** Of the 74 public targets,
  32 launch the graph and 42 run their leaf work in a Make recipe.
- **File rules always run.** 23 of the 25 file rules depend on `FORCE`, so
  freshness is decided inside `cache-artifact.sh`, the bootstrap scripts and
  the renderer store, not by Make.
- **26 operations re-enter Make.** The 20 artifact bindings run
  `make -- <files>`, and the 6 installations do the same for `make bootstrap`.
- **17 targets carry two recipes chosen by `GRAPH_CHILD`, and the graph uses
  neither child branch.** Graph units run the operation runner directly. Only
  the wrapper smoke reaches a child branch, and it expects a preflight failure.
- **Two owners for ordering, environment and freshness.**
  - Make file prerequisites repeat the policy units' `needs`.
  - Make strips 107 names and exports 61 bindings, then the planner builds the
    unit environment again.
  - Five toolchain units are both `same_run` in the planner and cached by
    `cache-artifact.sh`.
- **Freshness has five mechanisms and none is complete:**

  | Mechanism | Stores | Defect |
  | --- | --- | --- |
  | Planner unit cache | Row results and declared outputs | Its build profile is inert: every build unit has cache policy `none`. Unit digests include estimates and resource claims, so an estimate refresh or another host changes keys |
  | `cache-artifact.sh` | Digest records only; outputs stay at fixed paths | Keys hold absolute command lines and the whole Makefile. No Go binary key covers the `internal/gen` packages it imports, and the server key omits the migrations it embeds, so a migration-only change reuses a stale server (G47). Over 400 stale records per profile, never collected |
  | Renderer store | Immutable content-addressed roots and a receipt pointer | The right model, but its key includes `build-go-artifact.sh`, the whole pins file and package tests, and misses Go settings. Reuse re-hashes the installation (20 to 42 s); no lock |
  | Web build | A per-run private build, then a copy to `apps/web/dist` | No reuse: every run rebuilds. Publication leaves a window with no `dist`. Output depends on the unkeyed output mode, which adds sourcemaps |
  | Go binaries | Binary and `.crypto.json` receipt | Published by two separate renames; every hit rewrites the receipt outside the lock |

- **Builds and tests use different Go environments.**
  - Tests pin `GOENV`, `GOWORK`, `GOFLAGS`, `CGO_ENABLED=0` and `GOROOT`.
  - Builds inherit the user's values. Shipped binaries use Go's default cgo
    setting, on when a C compiler is present.
  - `GOTOOLCHAIN` becomes `local` under the installed-only policy.
  - The renderer and receipt subprocesses run without the configured caches.
- **Cost.** Artifact units took 78 s of unit work in a `check` with nothing to
  rebuild (`p66-check01`) and 201 s with a renderer rebuild (`p67-check02`).
  The pre-graph Make passes are not yet measured.

**Target.**

1. **Make is a thin generated wrapper.** One rule per public target records
   each input's origin and executes the harness. There are no file rules,
   prerequisites, mode switches or leaf recipes. The harness takes the
   workspace lock itself, so Make parses once. Public target names, command
   IDs, input contracts, outputs and exit statuses are unchanged.
2. **Builds are data.** A producer declares its kind (Go binary, web bundle,
   embedded assets, renderer installation, Node runtime, frontend install,
   Playwright, Go tool, ShellCheck, release inventory), its inputs, the
   environment that enters its key, and its outputs. Inputs are derived where a
   tool can derive them, such as `go list -deps` with embed files. The runtime
   runs producers itself, and no unit re-enters Make.
3. **One content-addressed store.**
   - It generalizes the renderer store: immutable roots keyed by declared inputs
     and toolchain identity, with a receipt as the only mutable pointer.
   - It publishes a multi-file output, such as a binary with its receipt, as
     one unit.
   - It takes a lock per key and collects unreferenced roots.
   - Conventional paths (`build/bin/*`, `apps/web/dist`, the toolbin) are
     published views of store roots.

   It replaces `cache-artifact.sh`, the renderer's private store, web
   publication and the readiness stamps.
4. **One toolchain environment.** The planner derives Go and Node settings from
   the pins for every build and test. The P5-0 overlay record,
   `NestedEnvironment`, the 20 mode switches and the 226 strip entries leave.

**Workstreams.**

| Step | Work | Exit |
| --- | --- | --- |
| P5-3a | Correctness and dead code, in the current structure. The server, operator and migrate keys gain `internal/gen`, and the server's gains `db/migrations` (G47). A binary and its receipt are published together, inside the lock. Remove what nothing reaches: the 17 child recipes, once the wrapper smoke stops depending on them; unused Makefile variables; switches with no reader or no writer (`TARGET_EVIDENCE_FINALIZE`, `DEFER_OBSERVABILITY_FINALIZE`, `DEFER_TOOL_SUMMARY`, `READINESS_CHECK_ONLY`). Measure the pre-graph Make passes | A test shows a migration-only change rebuilds the server; generated Make shrinks; segments, then one `check` |
| P5-3b | The store, in Go, with Go binaries as its first producer. Keys come from `go list -deps -json` (Go, embed and cgo files), the module files, the pinned Go settings and the producer's own declaration, not from launcher scripts. Unit keys drop scheduling-only fields (estimates, claims) | Go binaries reuse across runs with exact keys; binary and receipt are atomic; `cache-artifact.sh` serves no Go binary |
| P5-3c | Every other producer moves to the store: the renderer (its reuse check becomes a sealed-root identity check), embedded assets, the web bundles (cacheable; flags independent of output mode), the Node runtime, frontend install, Playwright, Go tools and ShellCheck. Artifact-binding and installation operations become producer operations, and `make bootstrap` runs them through the harness | `cache-artifact.sh` and the 25 file rules deleted; 26 Make re-entries become 0; one `check` |
| P5-3d | One toolchain environment from the pins for Go and Node builds and tests (D26, D27); `NestedEnvironment` and the overlay record removed | One environment table for builds and tests; release package rows pass |
| P5-3e | The thin wrapper. The 42 targets that run leaf work in Make run through the harness's command entry. The harness takes the workspace lock. The generated Make keeps only input-origin capture and one `exec`. The mode switches and strip entries are deleted | Make parses once per command; the generated include and Makefile are small; one `check`, then `ci` with P4-7 |

**Order.** P5-3a first: it fixes a correctness gap and shrinks what later
steps move. P5-3b and P5-3c depend on each other's store. P5-3d needs P5-3b's
producer environment. P5-3e needs P5-3c, because Make keeps its file rules
until no unit asks for them. Each step is proven on segments of the producers
and rows it touches; a build producer is also proven by planning `check`,
`ci` and `release-measure` before and after and by the digests of what it
publishes. `ci` folds into P4-7's final validation.

**Specification.** Amend, step by step:

- the producer clauses that distinguish direct and graph-child invocation
  (REQ-817, frontend producers): every invocation enters the graph;
- the renderer store clause (REQ-060) and AC-028: one store and its controls;
- the cache inputs in Section 5;
- the `builds` row, whose outputs come from Make variables, and the
  execution topology's `output_make_variable`;
- the task-surface Make binding profiles (REQ-069).

`CARTULARY_MAKE_INPUT_SOURCES` stays: only Make knows an input's origin.

**Risks.**

- **Breadth.** Every build path changes. Producers move one at a time, each
  with plan and output-digest equivalence.
- **Paths other tools read.** The container build, release packaging, the
  `CARTULARY_*_BIN` bindings and developers use `build/bin/*` and
  `apps/web/dist`. Those stay as published views (D25).
- **Developer commands.** `make dev`, `make services-up` and `make build` keep
  their names and behavior.
- **Bootstrap edits module files.** `install-go-dependencies.sh` can rewrite
  `go.mod` and `go.sum`, which every Go key reads. P5-3c makes bootstrap
  read-only, or makes the edit explicit.

### 14.25 Correctness and Dead Make (P5-3a)

Approved by the owner on 2026-11-26 with P5-3 and D24 to D28 as recommended.

- **Go build keys are the import closure (G47).**
  - **New option.** `cache-artifact.sh` takes `--go-package` (with `--go-tags`).
  - **What the key covers.** It keys every repository file the build reads:
    the non-test Go, cgo, assembly and C sources and the embedded files of each
    package that `go list -deps` reports. It runs under the build's own module
    and build caches, read-only.
  - **Which builds use it.** Every Go build Make runs: the server, harness
    server, migrate, operator, harness runtime, test services, source check and
    the embedded-asset generator. The hand-listed root sets and their discovery
    macro are gone.
  - **Effect.** The server's key now holds its 31 embedded migration files and
    40 generated files, and no test file. A test-only edit no longer rebuilds a
    binary. Listing a closure takes about 0.1 s.
- **Binary and receipt.**
  - The build helper writes the receipt for the staged binary, so the receipt
    check rejects a bad binary before anything is published.
  - It then renames the binary and the receipt into place back to back, inside
    the cache lock.
  - The seven Make rules no longer rewrite the receipt after the cache step,
    outside the lock. A cache hit already proves the receipt matches its
    record. One atomic pointer for both files waits for P5-3b's store.
- **Dead Make removed.**
  - **The 17 graph-child recipes.** The renderer emits only the graph launch for
    a graph-entry target, because the graph runs the operation through the
    operation runner. The generated include lost 195 lines and has no
    `GRAPH_CHILD` test. The wrapper smoke's two calls now reach the graph
    launch's preflight, which rejects their overrides with the same exit status.
  - **Eleven unused Makefile variables and one unused runtime macro.** The
    variables are the Go and pnpm environments, the owned-stack environment, the
    step macro, three tool flag sets, `comma` and a duplicate `Q`; the macro is
    `TASK_SURFACE_GO_ENV`.
  - **Four mode switches with no reader or no writer.**
    `TARGET_EVIDENCE_FINALIZE`, `DEFER_OBSERVABILITY_FINALIZE`,
    `DEFER_TOOL_SUMMARY` and `READINESS_CHECK_ONLY`, with the cache helper's
    inspect-only branch and its test. The registry holds 16 mode switches and
    473 names.
- **Measured pre-graph cost.** A small warm `test-slice` takes 1.3 s of wall time
  around a 0.2 s graph, so about 1.1 s goes to Make, admission and preflight. A
  `lint CHECK=go` that rebuilt the harness binary spent 5.5 s outside its graph.
- **Kept for P5-3c.** `list-build-inputs.sh` now has no production caller; its
  test, operation and row go with `cache-artifact.sh`.
- **Specification.** Amended: a Go binary's key covers its import closure, and
  a builder receipts the staged binary before publishing either file.
- **Validation.**
  - **Segments (`p77-*`).**
    - `make generate`, Go and script lint, and `source-check` for surface,
      catalog and generated all passed.
    - Nine command-surface rows passed, among them the wrapper smoke, generate
      drift, task-surface report and planner. So did two generated-artifact
      rows and three evidence-accounting rows.
    - `make build` rebuilt `migrate`, `operator` and `server` on their new
      keys, then reused all three. `build COMPONENT=server-harness` rebuilt it.
  - **Two failures, both repaired (`p78-*`).**
    - ShellCheck rejected the two intended single-quoted `$` strings (SC2016).
      The Go template now carries a disable directive with its reason, and the
      test's stub producer is a heredoc.
    - The cache-artifact row failed after every assertion had passed. Under the
      unit's `GOFIPS140` setting, `go list` downloaded the module snapshot into
      the test's empty module cache, read-only, and the cleanup could not delete
      it. The test now uses the build's caches and makes its scratch writable
      before removing it.
  - **`p79-check01`.** All 1,897 rows passed. The graph took 543 s against 577 s
    at P5-4b, and 617 and 674 s while the leaked PostgreSQL slot held a lane
    (14.18). Wall time was 555 s.
    - Artifact units took 87 s (78 s in `p66`). Every Go tool's key includes the
      whole Makefile, so this change rebuilt the seven tools once.
    - The renderer's reuse check took 25 s (41 s in `p66`).
