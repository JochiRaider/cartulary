# Release Qualification Redesign Tracker

Status: adopted 2026-11-24 (owner adopted R1–R5); implementation in
progress on branch `test-efficiency-phase2-plan`. The evidence-heavy fixture
question (Section 9) is open.

## 1. Purpose

Replace the monolithic `release-check` aggregate with a design built from the
release requirements, not from the current graph. The owner asked for a fresh
approach that carries no bad design forward.

## 2. What a release must prove

- **Implementation conformance (Core 00–04).** Product behaviour through the
  catalog rows; Core 04 §9 makes the timed and fixture-sensitive criteria
  pass/fail conformance, measured under the closed fixtures (REQ-04-157) on a
  quiet host. Core 05 claim publication is separate and stays out of scope.
- **Release artifacts.** One package producer; archive inspection; package
  startup, Reference Pack and operational-recovery consumers; credential
  capacity; all bound to one manifest and archive digest (targeted release
  amendment, 2026-11-06).
- **Supply chain and policy.** SBOM, licence report, release inventory,
  cryptographic-policy assessment, static and security analysis, generated and
  boundary checks.
- **Compatibility.** The SeaweedFS S3 profile and its strict release gate.
- **Evidence integrity.** Every record bound to one source digest and
  toolchain, cleanup closed, retained-secret scan passed, nothing inferred
  from absence. Visual and UI-Review stay explicit and outside release.

## 3. What the current design does

The 4e run (`20261124T112817Z-p93015`, Go planner, 4,717 s) shows the shape:

| Work | Units | Work time | Notes |
| --- | ---: | ---: | --- |
| Product catalog rows (Go, Vitest) | 338 | 10,231 s | Same work as `ci` |
| Webserver-backed browser cohorts | 23 | 4,112 s | Same as `ci`; the critical path |
| Quiet measurement sessions | 33 | 2,024 s | Serial, host-exclusive |
| Release and package operations | 14 | 1,635 s | Package chain, smokes, supply chain |
| Snapshot builds | 4 | 1,012 s | Evidence-heavy alone 566 s (21 GB) |
| Harness self-tests (contracts, rows) | 85 | 1,497 s | Same as `ci` |
| Other (a11y, finalizers, static, builds) | 89 | 1,778 s | Mostly `ci` |

Total work is 22,288 s; the host has 19 CPU tokens, yet average concurrency
was 4.7. `release-check` is the 507-unit `ci` graph plus 80 units: the
measurement campaign, the package chain and the release operations.

## 4. Design problems not to carry forward

1. **Two definitions of a release.** The `release-check` aggregate runs every
   tier (2,749 catalog rows) as one graph. The targeted path
   (`release-evidence-verify` with `contracts/verification/release-qualification.v1.json`)
   qualifies from retained roots against 167 obligation rows and 33
   operations, and calls the aggregates "optional conveniences".
2. **Two scheduling regimes in one graph.** Correctness work wants a saturated
   host; quiet measurement wants an idle one. Reconciling them inside one
   scheduler needs a `host_activity` lock on all 586 units for the sake of 35,
   an exclusive barrier, renderer retirement before measurement, a lifecycle
   pause port, and now the A21 quiet-readiness rule. The 4e run still lost
   about 450 s draining before measurement and 789 s of critical-path
   starvation after it, because aging accumulated behind the barrier.
   Measurement groups also sit inside `ci`.
3. **No evidence reuse.** A release candidate reruns the whole `ci` graph for
   a commit that already passed `ci`. One unrelated failure (Fallow) costs a
   full 80-minute rerun.
4. **Coupled retries.** A failed measurement session, package smoke or
   compatibility probe forces everything to run again.
5. **Shared mutable outputs.** Nested commands inside the release graph
   rebuilt tool and runtime binaries in place while other units executed them
   (fixed by atomic publication in `3150155c4`; the coupling remains).

## 5. Proposed design

Separate the work by regime, run each part once per source digest, and make
one evidence decision.

| Command | Regime | Contents |
| --- | --- | --- |
| `ci` | Throughput | Unchanged correctness scope, minus its quiet measurement groups. No quiet units. |
| `release-check` | Throughput | Release-only work: package producer, inspection and consumers (startup, Reference Pack, operational recovery), SBOM, licence, release inventory, cryptographic policy, SeaweedFS compatibility and gate. No `ci` rows. |
| `release-measure` (new) | Quiet campaign | Snapshot builds first (ordinary, parallel); then one whole-host exclusive lease for every active measurement predicate and the credential-capacity assessment, run in sequence, each with its own clone, bucket and session as REQ-814 and REQ-820 require. Credential capacity binds the package digest from the `release-check` root. |
| `release-evidence-verify` | Decision | The only release decision. Consumes the three roots and the authored qualification policy; requires one source digest and toolchain across roots, one package identity across `release-check` and `release-measure`, a complete passing `ci` root, every required operation, and every active measurement predicate. Launches nothing. |

A thin `make release` convenience MAY run the missing stages with explicit run
IDs, write the roots file and call the verifier; a failed stage reruns alone.

**Consequences for the harness.**

- No throughput graph contains quiet work, so the scheduler drops the
  exclusive barrier, quiet readiness (A21, reverted), the lifecycle pause,
  renderer retirement before measurement and the `host_activity` lock modes on
  ordinary units. Quiet measurement becomes one exclusive request in the host
  admission ledger plus a sequential runner.
- Measurement estimates leave the throughput ranker entirely.
- `release-evidence-verify` reads `plan.json` (Stage 2b) instead of
  recompiling plans, so this design follows or joins Stage 2b.

**Ordering constraints.** Credential capacity consumes the package archive,
and the SeaweedFS gate requires a passing operational-recovery row, so
`release-measure` reads the `release-check` root and the gate reads the `ci`
root's recovery row. `release-measure` may build its snapshots while
`release-check` runs; it requests the quiet lease only after the
`release-check` root is terminal. Core 04 requires an otherwise idle reference
host for performance measurement, so the quiet phase overlaps nothing on its
host.

## 6. Expected wall clock

Figures come from the 4e run's unit durations (today's code, before the
parallel snapshot restore); the `ci` figure is simulated and ignores host
admission and contention.

| Scenario | Today | Proposed |
| --- | ---: | ---: |
| Release candidate whose commit already passed `ci` | 4,717 s (reruns `ci`) | about 2,870 s: `release-check` about 700 s (builds, package 142 s, longest smoke 441 s) with snapshot builds (566 s) beside it, then the quiet phase about 2,170 s (sessions 2,024 s, credential capacity 142 s) |
| Cold release from nothing | 4,717 s | about 4,670 s (`ci` about 1,800 s first); no faster, but each stage retries alone |
| `ci` | carries 4 quiet groups, their drain and barrier | no quiet units |
| Retry after one failure | full rerun | the failed stage only |

The quiet phase is the floor on one host. The parallel snapshot restore
(commit after `3ebc20e3c`, eight workers) cut an evidence-heavy session's
stack acquisition from 239 s to 83 s (`20261124T135824Z-p18155`), about 310 s
off each release's quiet window; the snapshot build is 512 s. Per-session
start-up work is the next reduction, and the quiet phase can move to a second
reference host that runs beside `ci`.

## 7. Decisions for the owner

| ID | Decision | Recommendation |
| --- | --- | --- |
| R1 | Release qualification is evidence-composed: `release-evidence-verify` over `ci`, `release-check` and `release-measure` roots is the only release decision | Adopt |
| R2 | Quiet measurement leaves `ci` and `release-check` for `release-measure` (earlier A7/D3, now measured) | Adopt |
| R3 | `release-check` contains release-only work and no `ci` rows | Adopt |
| R4 | The qualification policy requires a complete passing `ci` root, not only the 167 obligation rows | Adopt; obligations stay as the capability map |
| R5 | Keep a thin `make release` orchestrator or document the three commands | Keep it thin; no graph of its own |

All five adopted by the owner on 2026-11-24.

Each needs NLSpec amendments: the aggregate tier table, the "MUST NOT
recursively invoke lower gates" clause, REQ-806 quiet admission, the targeted
release amendment and the qualification-policy schema.

## 8. Migration slices

Aggregates come from `tools/harness_work_graph_owner.json`
(`tier_order`, `aggregate_tiers`, `aggregate_policy_inherits`,
`aggregate_policy_roots`) and each row's authored `aggregate_placement`.
Today 30 measurement rows sit in the `release` tier and 4 in `full`; the
`release` tier also holds the 4 package rows and 7 `platform.telemetry` unit
rows (no placement justification).

1. **S1 Selection.** Measurement rows become `explicit_only`; the telemetry
   rows move to `full`; the `release` tier leaves `tier_order`.
   `release-check` stops being a tier aggregate and selects release-only work
   through a fixed selection (package producer, inspection and the three
   standup consumers with their rows, SBOM, licence, inventory, syft,
   cryptographic policy, SeaweedFS compatibility and gate).
   `harness.suite_browser_support` and `harness.build_production` stay with
   `ci`. New public command `release-measure` selects the measurement suite
   explicitly with its snapshot builders. NLSpec: aggregate table, recursion
   clause, REQ-865, command tables. Targeted: planner, task-surface, catalog
   and graph-contract rows.
2. **S2 Credential capacity.** Moves to `release-measure` and consumes the
   package of a named `release-check` root, keeping one package identity.
3. **S3 SeaweedFS gate.** The recovery qualification moves out of the gate
   into the verifier's policy, which reads the `ci` root's recovery row; the
   gate keeps its source, supply-chain and compatibility checks.
4. **Stage 2b** (work-graph tracker): the Go planner by default and readers on
   `plan.json`, so the verifier stops recompiling plans. Done 2026-11-24.
5. **S4 Verifier and policy v2 (R1, R4).** Required roots by command (`ci`,
   `release-check`, `release-measure`), each complete and passing, one source
   digest and toolchain, every active measurement predicate, one package
   identity. `release_readiness_evidence` folds into the verifier's output.
6. **S5 `make release` (R5).** Runs the missing stages with explicit run IDs,
   writes the roots file, calls the verifier.
7. **S6 Cleanup.** Remove aggregate inheritance and quiet-only scheduling
   paths that no throughput graph uses any more.
8. **S7 One real release**, with the owner's go-ahead.

### S1 result (2026-11-24)

- **Rows.** 34 measurement rows and the 4 release package rows are
  `explicit_only`; the 7 telemetry rows are `full`. The `release` placement is
  retired: work-graph owner schema v12 drops it from `tier_order` and drops
  `release-check` from the aggregate maps (v11 retired). The family-manifest
  schema v14 still lists `release` in its enum; validation rejects it, and the
  enum goes in S6.
- **Selections.** `harness.release_qualification` (package archive, startup,
  Reference Pack, recovery and credential capacity; SBOM; licence;
  cryptographic policy; SeaweedFS compatibility and gate; release readiness;
  the four package rows) and `harness.release_measurement` (the measurement
  suite) are explicit selections, as are the `package_*` selections, because
  the package operations declare the explicit-only package rows as policy
  dependencies (REQ-800). `harness.suite_browser_support` moved to the `ci`
  roots.
- **Commands.** `release-check` v5 uses the fixed selection, drops
  `BROWSER_LANES` and no longer inherits `ci`; `release-measure` v1 is new.
  `browser-e2e SUITE=measurement` is removed. Five `ci` commands drop
  `release-check` from their inclusion sets. `release_readiness_evidence` now
  needs only release work (package consumers, credential capacity,
  cryptographic policy, SBOM, licence, SeaweedFS gate), not `contract_default`,
  `go_gosec_audit` or the browser-support suite.
- **Plans** (JavaScript target plans): `ci` 501 units with no measurement;
  `release-check` 28 units (release operations and builds, the four package
  rows, and the recovery row the SeaweedFS gate still needs until S3);
  `release-measure` 78 units (33 sessions, 29 summaries, 4 snapshot builders and
  their builds).
- **Task guides.** Schema v7 adds `make release-measure` and
  `make release-check` to the explicit commands. Scheduler growth now measures
  `ci`, the largest graph.
- **Checks.** `lint` `go` and `scripts`; `source-check` `catalog`, `schema` and
  `generated`; the Go planner, native-runtime and scheduler-smoke rows; the
  browser lane policy row; `harness-contract SUITE=contract` and
  `SUITE=regression`, whose three failing rows (routing contract, package
  readiness contract, planner preparation) were fixed and rerun alone: all
  pass.

### S2 result (2026-11-24)

Credential capacity is quiet measurement and now runs in `release-measure`.
`release-measure` and the standalone `credential-capacity-assessment` take a
required `RELEASE_RESULTS_DIR` (forwarded to units); `packageReceiptRoot`
accepts it only as a passing `release-check` run of the same source digest and
otherwise reads the run's own package. The package archive already lives in
the shared, receipt-verified store and installation loads the images from it,
so consumers needed no other change. `package-inspect CHECK=all` keeps the
four same-run consumers; `release_readiness_evidence` checks those three
standup consumers, and the cross-root binding of credential capacity is S4's
verifier work. `release-check` now carries no quiet unit. Checks: `lint`
`go` and `scripts`; `source-check` `generated` and `schema`; the Go planner,
package readiness contract and routing contract rows: all pass.

### S3 decision (2026-11-24): the gate keeps its recovery prerequisite

The NLSpec requires the SeaweedFS release gate to declare current-invocation
prerequisites for `seaweedfs-compatibility` and the canonical operator
process row, and to read that row from the same invocation, never from
retained or historical evidence. Moving it to the verifier would weaken that
guarantee from "same invocation" to "same source digest", which is an owner
decision, not a refactor. `release-check` therefore keeps exactly one `ci`
row: the gate's declared, owner-justified policy dependency
(`module.recovery.process.canonical_operator_process_evidence_maps_the_imp_9808fdd4e9`),
which REQ-800 permits. It runs beside the 441 s package smoke, so it adds no
wall time. S3 is closed without a change.

## 9. Evidence-heavy fixture v3 (decided 2026-11-24)

The owner chose the count-faithful v3 fixture. Core 04 REQ-04-157 now defines
`cartulary.perf.evidence_heavy.v3`: 5,000 Timeline rows and 10,000 Evidence
records with 10,000 distinct complete 2x2 PNG objects; the first 100 in
construction order (the linked records) keep 2,097,152 bytes and the other
9,900 hold 4,096 bytes (250,265,600 bytes in total, from about 21 GB). v2 is
retired because its measured steps read only the 100 linked objects and one
new upload. The predicate IDs stay; the fixture ID, the contracts (`ac045.v3`
with schema v2, `evidence-verification.v2` with schema v2), the snapshot profile
(`evidence_heavy_snapshot_v3`) and the evidence contribution
(`evidence.objects.v2`) change. The evidence owner's fixture command takes
ordered segments, so the linked records come first, and the validator checks
each record's size and digest by class.

Measured (`20261124T150150Z-p82659`, `20261124T150949Z-p91922`, both rows
pass): a session's stack acquisition fell from 239 s (v2 sequential) and 83 s
(v2 with parallel restore) to 16–17 s, about 445 s off the quiet window per
release against the original. The snapshot build fell from 512–566 s to
287–373 s; the remaining time is the 10,000 sequential evidence-record
mutations (265–349 s), not object bytes.

Checks: `lint` `go`, `types`, `frontend` and `scripts`; `source-check`
`generated`, `schema` and `catalog`; the JSON-shape assertions; nine
`harness.browser` fixture rows, including the evidence object and inspector
integration tests; and the two evidence-heavy measurement rows.
