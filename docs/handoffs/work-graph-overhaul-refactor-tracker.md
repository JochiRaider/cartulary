# Work-Graph Overhaul Tracker

**Status.** Audit complete, 2026-11-23. Stage 0 tried and reverted; Stage 1 scoped, not started. The owner
asked for an audit of how the work graph is built and executed after G29
(test-efficiency tracker, Section 5) showed that a one-rule scheduler change
moved `check` wall time by +200 s through effects in another layer. This
tracker records the audit, the measured costs and the decision between
overhauling graph generation in place and replacing it with a Go
implementation.

**Method.** Two read-only code audits (graph creation; graph execution)
and analysis of retained diagnostic runs `g28a-final-check01`,
`g28a-final-check02` (first-fit backfill, 702–710 s) and `g29-check01`,
`g29-check02` (logical deficit-disjoint backfill, 899–903 s). File
references are repository-relative.

## 1. How the graph is built

- **Entry.** The Go runtime runs `node tools/harness/scheduler/work-graph/prepare-cli.mjs`
  for every invocation (5-minute deadline) and reads the plan
  `cartulary.harness_runtime_plan.v7` back over file descriptor 3; the plan
  is never written to disk (`internal/testutil/harnessruntime/prepare_linux.go:24-170`).
- **Stages.** `prepareRuntimePlan` (`work-graph/prepare-plan.mjs:17-86`)
  loads the owner model and topology, captures the host snapshot, compiles the
  selection (`compileAggregatePlan` for `check`), computes cache facts and
  emits graph, manifest, capacities, aging quantum, projections and bindings.
  `buildWorkGraph` (`work-graph/model.mjs:189-254`) adds the shared
  `host_activity` lock and each unit's resident `physical_allocations`.
- **Units in `check` (364).** About 311 Go shards, 10 Vitest cohorts, 9 shell
  rows and 34 operations. Go rows pack first-fit by compatibility digest up to
  300 s (225 s for heavy) (`backend/go-lpt-shards.mjs:28-107`); about 186
  fixture-free shards are roughly one per package, and about 99 have estimates
  of 7 s or less, 5.25 s of it fixed overhead. Vitest rows pack first-fit per
  file against the 600 s unit timeout (`work-graph/cohorts.mjs:9-20`), which
  yields one 377-row browser-unit cohort (378 s estimate, 2 CPU) and one
  277-row node cohort (278 s).
- **Claims.** Resource profile plus service minimums
  (`work-graph/resource-claims.mjs:32-46`); Go PostgreSQL units are bounded to
  one CPU (`work-graph/compiler.mjs:114-134`), so every Go test runs with
  `GOMAXPROCS=1 -p=1`; Vitest adds coordinator claims (cpu 1, 384 MB,
  3 processes) with one worker. Operations claim one CPU but set no
  `GOMAXPROCS` or thread limit (`work-graph/compiler.mjs:657-670`).
- **Capacities.** CPU tokens = 80% of host parallelism (19 here), PostgreSQL
  lanes `min(16, cpu)`, object-store lanes `min(8, cpu)`, heavy lanes 3
  (`work-graph/capability.mjs:58-145`). The mapping is defined three times
  (registry, `capability.mjs:280-296`, `plan.go:410-438`).
- **Estimates and rank.** Row estimate or evidence-class default
  (`cohorts.mjs:3-5`): 943 of 1,165 Go rows, all 709 Vitest rows and all
  shell rows use defaults; operations use authored constants, or 1000 ms when
  missing. Overheads are flat. No measured duration feeds back automatically.
  Rank is estimate plus longest downstream path plus wait age
  (`engine.go:358-381`, `engine.go:491-512`) and ignores claim width, scarce
  lanes and residency.

## 2. How the graph is executed

Two schedulers admit every unit, in sequence, against the same capacity
numbers but charging different footprints.

1. **Logical layer** (`internal/testutil/harnessruntime/engine.go`). Ranked
   first-fit with backfill and no reservation (REQ-802). At admission it
   reserves the unit's claims, emits `admitted` and `started`, and only then
   starts the worker (`engine.go:623-632`). Claims stay reserved until the
   engine processes completion (`engine.go:706-713`).
2. **Physical host admission** (`internal/testutil/harnessadmission`,
   reached from `lifecycle.go:74-82`). Charges the same active claims plus
   resident allocations (`harnessfixtures/work_linux.go:95-120`), cleanup
   headroom and other processes' leases; older waiters keep priority with
   deficit-disjoint backfill and heavy-lane precedence (REQ-803,
   `harnessadmission/admission_linux.go:719-778`). It retries on a fixed 50 ms
   timer under a host-wide file lock, with no wake on release
   (`harnessadmission/lease_linux.go:110-176`).
3. **Fixtures.** Exclusive PostgreSQL units start their own cluster
   generation in `fixture_acquire` and terminate it synchronously in
   `fixture_release`; shared clusters and the object store stay resident until
   their last consumer ends.

A resident PostgreSQL cluster claims one `postgres` lane, 8 process slots and
256 MB (`tools/execution_topology_manifest.json`,
`physical_resource_claims`). The logical layer never charges residents, so an
exclusive PostgreSQL unit costs one logical lane and two physical ones.

## 3. Measured costs

| Measure | `g28a-final-check01` (first-fit) | `g29-check02` (G29 rule) |
| --- | ---: | ---: |
| Graph wall | 710 s | 903 s |
| Ordinary PostgreSQL units: `host_admission` | 4,139 unit-s | 9,146 unit-s |
| Same units: `runner` | 1,972 unit-s | — |
| CPU tokens held while waiting in host admission | 31% of CPU capacity over the run | 10.3 of 16.8 held (average) |
| PostgreSQL lanes held while waiting in host admission | 37% of lane capacity over the run | 10.3 of 14.2 held (average) |
| Resident clusters' share of the 16 `postgres` lanes at denials | 7–8 | 10–12 |
| Physical denials (no current deficit, priority only) | 321 of 780 | 569 of 1,302 |
| Logical utilization: cpu, postgres, heavy, object store | 79%, 72%, 83%, 74% | — |

Other measurements from `g28a-final-check01`:

- Actual-to-estimated execution: fixture-free Go 0.44, PostgreSQL Go 1.81
  (host admission included), heavy Go 0.79, operations 3.0, Vitest 0.79.
- Running units fall from 19 to 7 after 500 s and to 1.5 after 600 s; the
  critical path is `frontend_install` followed by the 377-row Vitest cohort,
  which starts at 380 s because first-fit backfill never leaves two tokens
  free (G29).
- In `g29-check02` the critical path became the build chain into one heavy
  unit that waited from 100 s to 821 s for a `postgres` lane held by residents
  and by units that were themselves waiting physically.

## 4. Findings

Ranked by measured or likely effect on wall time.

1. **Two capacity authorities with different footprints.** Logical admission
   ignores residents and over-admits PostgreSQL work about 2×. Admitted units
   then wait in host admission holding CPU tokens and PostgreSQL lanes:
   4.1k unit-s in the best run, 9.1k in G29. This is the G4b head-of-line
   coupling, and it explains both G24 and the G29 regression.
2. **Opposite fairness policies.** The logical layer forbids reservation
   (REQ-802 step 4); the physical layer reserves for older waiters (REQ-803).
   Logical rank only sets arrival order at the physical queue, which then
   reorders by sequence, deficits and heavy precedence. A change in either
   layer shifts the other's queue unpredictably.
3. **Residency is unplanned.** How many shared clusters stay resident depends
   on admission order, not on a plan. G29 raised residents from 7–8 to 10–12
   lanes and cut active PostgreSQL work to about 4 lanes.
4. **Multi-token claims starve.** Aging only reorders units that already fit,
   so REQ-802's "backfill cannot starve work" fails for the 2-CPU Vitest
   cohorts (268 one-token backfills ahead of one in `g28a-final-check02`).
5. **Oversized indivisible units.** Packing minimizes the number of cohorts,
   not makespan: one 378 s Vitest cohort with one worker sets the critical
   path floor, and editing any of its 207 files reruns all of it.
6. **Estimates are mostly defaults with flat overheads**, never fed back, and
   rank uses them directly; actual/estimate ranges from 0.44 to 3.0 by kind.
7. **Claims do not match use.** Go tests are pinned to one CPU (serial
   compile), operations claim one CPU but use many, Vitest uses one worker.
8. **Logical claims held through non-compute phases**: physical wait,
   synchronous cluster start and teardown, evidence and cache publication.
9. **Physical admission mechanics**: 50 ms polling, a global lock with full
   state read and write per attempt (285 s of `admission_transaction` unit
   time in `g28a`, 518 s in G29), no wake signal, no wait timeout.
10. **Observability hides the coupling.** Queue time excludes host admission;
    physical decisions exist only under diagnostics; the logical choice
    (effective rank, skipped units, backfills) is never recorded.
11. **Duplicated logic across JavaScript and Go**: capacity mapping (3×),
    feasibility including physical allocations (2×), the Go shard estimate
    formula (4×), digests and the `host_activity` rule, and row IDs parsed back
    out of evidence paths in both languages.
12. **Spec and implementation diverge**: REQ-354 and REQ-357 still say
    capacity-derived LPT while REQ-805 and the code use bounded first-fit;
    REQ-865 forbids silent defaults that the compiler applies
    (`compiler.mjs:649`, `compiler.mjs:690`); shell-row units ignore their
    operation's policy estimate, cache policy, profile and timeout; service
    readiness is hard-coded true although REQ-802 says the snapshot records
    it; measurement Go rows get the 600 s default timeout below their
    1,800 s estimate.

## 5. In-place overhaul or a Go implementation

**What has to change.** Findings 1–4 and 8–10 are execution problems in the
Go runtime and host admission. They need one resource model that both
planning and admission share: a unit's full footprint including residents,
one capacity authority and one fairness policy. Findings 5–7 and 11–12 are
graph-generation problems: unit formation, estimates, claims and duplicated
definitions. The largest measured cost, finding 1, does not depend on where
graph generation lives.

**Option A: overhaul the JavaScript compiler in place.** Keeps the 20 or so
JavaScript consumers (diagnostics, generated Make and topology artifacts,
release evidence, browser review) working unchanged and avoids a
parity migration. It leaves the language boundary in the middle of the
model: claims, footprints, capacities and estimates would still be defined in
JavaScript and re-derived in Go, so every model change is made twice and the
two can drift, as findings 11–12 show they already have.

**Option B: a Go implementation, run alongside, then the JavaScript
compiler removed.** The planner and the scheduler share one package, one
resource model and one capacity definition, and the runtime stops spawning
Node for every invocation. Planning features the audit calls for
(residency-aware ordering, capacity-aware cohort sizing, scarce-lane rank,
reviewed measured estimates) are written once, next to admission. Costs: about
6,600 lines of JavaScript (work-graph, scheduler, test-catalog) and 30 test
files to replace; the catalog and selector resolution must move or be
exposed; unit IDs and semantic digests change unless reproduced exactly, so
the plan version and cache identity break once; the JavaScript consumers need
a Go CLI that emits the graph as JSON until they are ported.

**Recommendation: Option B, staged, with the execution fix first.**

1. *Stage 0, language-independent.* Make logical admission charge the full
   physical footprint (active claims plus prospective residents, already in
   `unit.Physical`), so a unit is admitted only when it can run; record the
   logical choice and keep host-admission time out of execution in the
   canonical accounting. Validate with one `check`; this targets the
   4.1k–9.1k unit-s of held capacity directly and tells the new model what it
   must express.
2. *Stage 1.* A Go planner package that reproduces today's plan from the
   catalog, run beside the JavaScript compiler with a test that diffs the two
   plans for `check`, `ci` and owner slices. No behaviour change.
3. *Stage 2.* Switch the runtime to the Go planner (plan v8), add a Go CLI
   that emits the graph for the remaining JavaScript consumers, and delete the
   JavaScript compiler, cohort packing, capacity derivation and their tests.
4. *Stage 3.* Planning improvements in Go: one fairness policy across both
   layers, residency-aware ordering by cluster family, cohort sizing against
   capacity rather than the timeout, rank that weighs scarce lanes, measured
   estimates through a reviewed generated file, and CPU claims that match use.

Each stage changes the NLSpec first (REQ-371, REQ-800–817, REQ-802, REQ-803,
REQ-805, REQ-865 and the affected acceptance criteria) and is validated with
one diagnostic `check` against `g28a-final-check01`/`-02`.

**Risks.** Stage 0 changes admission order again, so it can shift residency
the way G29 did; it must be judged on idle-held capacity and resident counts,
not wall time alone. Stage 2 breaks cache identity once. The spec's
`harness_runtime_plan` and unit-event schemas are the stable contract, and
module paths are not public bindings (NLSpec implementation-package rule), so
the replacement needs no change for path reasons.

## 6. Stage 0 result: reverted

Stage 0 made logical admission charge residents as host admission keys them:
an exclusive unit's own generation for its admission, the shared generation
and shared object store from first consumer to last (`147651a9b`; scheduler
tests passed). One diagnostic `check`, `s0-check01`, passed all 1,890 rows but
ran 971 s against 702–710 s:

| Measure | `g28a-final-check01` | `s0-check01` |
| --- | ---: | ---: |
| Graph wall | 710 s | 971 s |
| Logical utilization: cpu, postgres | 79%, 72% | 56%, 46% |
| PostgreSQL units: queue (logical) | 35,559 unit-s | 75,107 unit-s |
| PostgreSQL units: `host_admission` | 4,139 unit-s | 3,325 unit-s |
| At physical `postgres` denials: resident lanes, active lanes | 7–8, 7–9 | 10–12, 4–5 |

The logical layer now admitted less, yet physical admission still denied:
exclusive generations stay resident after their unit completes until the
cleanup dispatcher drains them, and those drains need admission themselves. A
logical copy of the resident rules releases at unit completion and cannot see
that asynchronous retirement, so the two ledgers still disagree, now in the
other direction. Reverted in `0e193c15b`. The lesson for the design: resident
lifetime, retirement and admission must live in one ledger with one authority
(Stage 3), with retirement completing inside that ledger rather than
asynchronously beside it.

## 7. Stage 1 scope: a clean Go planner

**Decision (owner, 2026-11-23): Stage 0, then a Go planner.** The planner
is designed from the NLSpec's requirements, not ported: a current behaviour
is carried forward only when it materially improves the subsystem, and
compatibility burden is removed rather than reproduced. Byte-identical
digests with today's plan are not a goal; the plan and graph schema versions
change once.

**Inventory.** The JavaScript planner reaches about 7,100 lines in about 35
modules: graph logic (about 3,200), catalog and selector resolution (about
2,150, including TypeScript AST parsing of every Vitest and Playwright file on
each prepare), topology, browser, fixture and shard packing (about 1,250),
and input resolution (about 500). It validates about 20 schemas at plan time
that have no Go validator yet. About 15 other JavaScript tools call into it.

**Node.** Node is not kept because the NLSpec names it (owner, 2026-11-23).
It remains only where the tool under test is itself Node: running Vitest and
Playwright. Harness logic that runs in Node today (the planner, the row runner
that wraps `go test`, evidence and diagnostics tools, generated-artifact
renderers) moves to Go as each milestone reaches it, and REQ-852's Node
ownership clause is amended with milestone 1.

**What the planner is.** A pure Go package that turns authored inputs (owner
model, topology, catalog rows, task surface, capacity snapshot, measured
estimates) and a selection into a plan, plus a CLI that prints the plan or
graph as JSON. It does not parse sources for validation, spawn Node, or read
the clock for identity.

| Current behaviour | Decision | Reason |
| --- | --- | --- |
| Selection kinds: aggregate tiers, commands, targets, owner and row slices | Keep, one resolver | Public contract (`make check`, `test-slice OWNER=… ROWS=…`) |
| Policy operations and dependency closure (`needs` to exact producers) | Keep | Required ordering; already exact |
| Per-row resource profiles and service minimums | Keep, as one footprint | Becomes a unit's full footprint: active claims plus named residents (below) |
| Resident allocations re-derived in two places (Stage 0 mirror, host admission) | Replace | The plan names each unit's residents with key and scope (`unit`, `graph`, `persistent`); engine and host admission both read that one record |
| Go shard packing by compatibility digest to a fixed 300 s bound ("LPT" in name only) | Simplify | Unit = package and fixture class; size from measured estimates; drop LPT naming and the separate shard plan schema |
| Vitest cohorts packed first-fit to the 600 s timeout | Replace | Size cohorts against a target duration and capacity so no cohort sets the critical-path floor; allow more than one worker per cohort |
| Unit IDs from packing results and resolved titles | Replace | ID = digest of the unit's row set and execution contract; editing an estimate never renames a unit |
| TypeScript AST parsing for Vitest and Playwright titles and registration-level coverage on every prepare | Drop (removed from the NLSpec, 2026-11-23) | Rows already carry exact full titles; the pinned runner resolves them at execution and a title it does not collect exactly once fails; frontend completeness is file-level (every matching test file has a catalog row), which needs no parsing |
| TypeScript import closure with silent broad fallback for cache keys | Simplify | Workspace-level digests for frontend units; `go list -deps` closures for Go units; no silent fallback |
| Estimates: row values or class defaults, flat overheads, `?? 1000` defaults | Replace | A reviewed, generated estimates file from retained runs (per unit kind, including fixed overheads), with explicit failure when an estimate is missing |
| Capacity mapping defined in registry, JavaScript and Go | Simplify | One Go definition; the registry stays the authored input |
| Host profile from Node (`x64`, `os.release`) | Replace | Go host facts; identity strings change once |
| Shell rows silently overriding their operation's policy | Drop | Each unit takes its own declared policy; a conflict is an authoring error |
| Absolute binary and checkout paths inside unit environments (graph digest depends on checkout path) | Replace | Paths are bound at launch, outside unit identity |
| Plan never written to disk; retained runs replayed by recompiling (`compileRetainedRun`) | Replace | The plan is written to the run root; readers use it instead of recompiling |
| `validateWorkGraph` in every row-runner child | Drop | The runtime hands each child its validated unit; no per-child revalidation |
| Non-canonical browser execution key, locale-sensitive lane ordering, insertion-order lane chains | Replace | Canonical digests and explicit orderings |
| Browser lanes chained at compile time through `needs` | Revisit in browser milestone | A third ordering layer; candidate for scheduler-owned lane assignment |

**Spec changes before code.** The TypeScript source-parsing contract is
already removed (2026-11-23): rows carry exact full titles, the runner resolves
them, and frontend completeness is file-level. The JavaScript planner still
parses until milestone 3 removes it. REQ-802 ("one graph compiler") and
REQ-852 ("Node owns source-aware compilation") name the Go planner; REQ-363's byte-identical graph
stays as determinism of the new graph; the plan becomes
`harness_runtime_plan.v8` and the graph `work_graph.v9`; residents become an
explicit plan field (REQ-802, REQ-803); estimates come from a reviewed
generated file (REQ-865); the line forbidding a retired JavaScript
implementation as a native test oracle means the JavaScript planner is a
comparison aid only during Stage 1 and is deleted in Stage 2.

**Milestones.**

1. *Inputs and model.* Go loaders and generated validators for the owner,
   topology, task-surface and catalog schemas; one capacity definition; the
   plan v8 and graph v9 schemas with explicit residents.
2. *Go rows and operations.* Owner and row slices and operation targets
   (`make test-slice` for Go owners), plan v8 accepted by the runtime behind a
   planner selector. The JavaScript planner is not a reference (owner,
   2026-11-23: it never worked well, so nothing is copied for parity). The Go
   planner is judged on its own correctness: every selected row bound exactly
   once, every need resolved and acyclic, each unit waiting for what its rows
   require, footprints and residents that follow the declared services, and
   the outcome of running Go-planned slices.
3. *Aggregates and Vitest.* `test-fast` and `check`, with Vitest cohorts
   sized to a target duration and authored titles passed through to the
   runner; file-level frontend completeness; TypeScript parsing removed from
   the JavaScript catalog path.
4. *Commands and browser.* Command selections, `ci` and `release-check`
   browser stages; decide lane assignment.
5. *Switch and delete (Stage 2).* Go planner by default, plan written to the
   run root, a Go CLI for the remaining JavaScript readers, then deletion of
   the JavaScript planner, selector resolution's planning half, retained-run
   recompilation and their tests.

Each milestone is validated by targeted checks and one diagnostic `check`
once `check` is covered. The scheduler redesign (one admission decision
reading the plan's footprints, one fairness policy) follows as Stage 3 and
replaces Stage 0's mirrored resident keys.

**Risks.** Without source parsing a misauthored title fails when its file
runs rather than at prepare, and a test registered in a file that already has
a row but is named by no row is no longer detected. Unit identities and cache keys change once.
About 15 JavaScript readers must move to the CLI or the retained plan in
Stage 2; `operations.mjs` run-time dispatch is not planner code and stays.

## 8. Progress

**Milestone 1, first slice (2026-11-23).** New package
`internal/testutil/harnessplan`:

- `Load` reads the owner model, execution topology, scheduler resource
  registry and owner catalog with strict JSON decoding and typed structs for
  the fields planning uses, then checks the invariants planning depends on:
  unique rows, registered owners, runner-specific selectors, resolvable
  resource and runtime profiles, services consistent with PostgreSQL
  isolation, policy needs and aggregate roots that resolve to operations or
  target unions, known runtime binaries and preparation operations, required
  rows, and the capacity policies capture needs. Authored-input schema
  validation stays with the source checks; the planner does not re-run it on
  every prepare as the JavaScript planner does.
- `Capture` and `Capacities` are the single Go definition of the capability
  snapshot and the scheduler resource mapping, with host detection separate
  (`DetectHost`, Linux) so the rules are tested over synthetic hosts. The
  snapshot satisfies `harness_capability_snapshot.v1`. Compatibility branches
  for historical snapshots are not carried over.
- Operations whose policy unit is missing get no silent default; target
  members and operation execution definitions are typed when milestone 2
  needs them.

Tests: `TestPlannerInputs` (the real inputs load; eight mutations each break
one invariant) and `TestCapabilityCapture` (policies, CPU bounds, cgroup
memory, override bounds, unavailable services, schema and digest
determinism), registered as
`harness.command_surface.unit.work_graph_planner`. Checks: `plan-lintgo02`,
`plan-slice05`, `plan-srcgen01`, `plan-srccat01`; all pass.

Not adopted: generating Go validators for the planner's input schemas. They
use regex lookaheads the RE2-based native validators cannot express; the
native renderer already whitelists some lookahead patterns and silently drops
the phase-label exclusions, so native validation is weaker than the
JavaScript one for 19 patterns in 30 schemas. That is a separate cleanup:
rewrite each lookahead as a positive `pattern` plus a `not` pattern and
delete the whitelist.

**Milestone 1, second slice.** The runtime validates plan capacities through
`harnessplan.Capacities` (`7887d4d45`); its own mapping and the branch for
snapshots recorded before heavy lanes existed are gone. Checks: `cap-lintgo01`,
`cap-slice02` (native runtime, scheduler and planner rows, including the
preparation parity test that decodes a real prepared plan); all pass.

**Plan v8 and graph v9 design.** The schemas land with the Go producer in
milestone 2, so no schema exists without a writer. The contract:

| Field | v9/v8 | Reason |
| --- | --- | --- |
| Unit `physical_allocations` (service to claims) | Replaced by `residents`: key, scope and claims per resident | The planner names each resident once (`postgres:unit:<unit>`, `postgres:shared`, `object_store:shared`, browser stacks, renderer slots, measurement generations) with its scope: `unit` (retired with the unit), `graph` (retired after its last consumer) or `attached` (held by a persistent owner, not charged). The runtime and host admission read these instead of re-deriving keys (`clusterKey`), the root of the Stage 0 failure. A key has the same claims wherever it appears |
| Unit ID | Digest of the unit's row set and execution contract | Independent of packing results and estimates |
| `current_run_evidence_outputs` | Kept; row identity comes only from `row_bindings` | Readers stop parsing row IDs back out of `rows/<id>.json` paths |
| Command environment | Binding names, not absolute paths | Checkout and binary paths are bound at launch, so identity does not depend on the checkout location |
| Claims, locks, needs, timeout, estimate, fixture lease, services, PostgreSQL isolation, cache and failure policy, renderer, browser keys, semantic digest | Kept | Required by scheduling and execution |
| Plan `host_profile` | Go host facts (`amd64`) | Node strings dropped |
| Plan graph, manifest, projections, capacities, aging quantum, bindings, input environment, cache facts and source inputs, helper digest, preparation intervals, browser lane limit | Kept | Required by the runtime |
| Plan location | Written to the run root as `plan.json` | Retained-run readers use it instead of recompiling |

Milestone 1 is complete apart from this contract's schemas, which move to
milestone 2 with their producer.

**Milestone 2, slice 2a: unit compilation (2026-11-23).** `harnessplan.Compile`
turns a selection (exact rows, an owner slice, or an operation or private
target union) into a validated graph:

- *Operations* compile to one unit each under their own policy unit; there is
  no default policy. Dependencies attach through each dependency's end units,
  and a union's own policy needs attach to its member units.
- *Shell rows* bind to their operation's unit, which runs the row runner and
  publishes their evidence under the operation's policy. This removes the
  silent override. It exposed authoring conflicts, now fixed in authored data:
  six full-tier shell rows declared `standard` where their operation's policy
  declares `harness_contract`, so the rows take `harness_contract`; the
  UI-review contract policy declared `standard` where its row declares
  `ui_review_raster`, so the policy takes `ui_review_raster`; and eleven
  operations named by shell rows had no policy unit, so each now has one
  carrying the values that already applied (the row's profile, no needs, no
  cache, its evidence-class estimate).
- *Go rows* group by execution contract (package, profiles, fixture, services,
  binaries, preparation, cache, CPU, race) and split into balanced units:
  `ceil(work / target)` units filled longest-first into the least-loaded unit
  (225 s target for heavy-lane units, 300 s otherwise). Process-isolated rows,
  managed processes and rows with an explicit context run alone. Unit IDs are
  digests of the contract and the row set.
- *Residents* are explicit per unit: `postgres:unit:<unit>` (scope `unit`) for
  exclusive isolation, `postgres:shared` and `object_store:shared` (scope
  `graph`).
- Runtime binaries are listed per unit; their paths are bound at launch.
- Vitest, Playwright and measurement rows return `ErrUnsupported` until
  milestones 3 and 4.
- The Go suite of a row is still inferred from its family name's last
  segment, as before; an explicit suite field in the catalog is a later
  cleanup.

Tests: `TestPlannerCompilation` (every plannable row bound once and
deterministically; an operation's dependency closure; residents and scopes;
balanced units; shell rows under the operation's policy; unsupported and
invalid selections). Checks: `m2-lintgo01`, `m2-slice02`, `m2-srccat01`,
`m2-srcgen01`, `m2-srcjson01`; all pass.

**Milestone 2, slice 2b: the plan document (2026-11-23).** The NLSpec gains
the "Go work-graph planner amendment — 2026-11-23" (staged replacement,
REQ-852's Node ownership moving to Go per selection kind, graph v9, plan v8,
owner v11). Schemas `cartulary.harness_work_graph.v9` and
`cartulary.harness_runtime_plan.v8` are attached with Go validators; both
admit additive fields until the switch.

- *Graph v9.* `residents` replaces `physical_allocations`; `failure_policy`
  keeps only `block_descendants` (`continue_independent` was always true and
  `aggregate_effect` had no reader); `runtime_binary_ids` is added; the
  graph-level `selector` is gone. Renderer, affinity, fixture-profile and
  snapshot fields return with the milestones that compile that work.
- *Finalization.* `Compile` canonicalizes every unit (sorted sets, empty
  lists instead of null, the unit result `unit-results/<unit>.json` in the
  evidence outputs) and computes unit and graph digests with one helper, which
  the capability snapshot digest now shares.
- *Plan v8.* `Assemble` binds a graph to an invocation's facts (run identity,
  declared inputs and input environment, snapshot, Go worker budget, source
  inputs and provenance, cache mode and facts, instrumentation, host profile,
  launch bindings) and computes the manifest digests. `Plan.validate` holds
  the invariants the runtime's v7 decoder checks today (digests, modes, worker
  budget, lane limit, input environment, capacities, footprints including
  residents, evidence and artifact outputs, projections, cache facts) so the
  runtime and retained-run readers share one check. `helper_digest` and
  `preparation` are dropped.
- *Owner v11.* Each policy unit declares `current_run_evidence_outputs`. The
  JavaScript compiler's hard-coded table of operation outputs is deleted and
  it reads the owner instead. `harness.test_evidence_audit` and
  `harness.release_evidence_verify` get policy units with the values the
  JavaScript default supplied (standard profile, no needs or services, no
  cache, 1000 ms), so no operation that declares evidence relies on a default.
- *Fix to 2a.* Residents follow the fixture lease: a unit without one (the
  standup and credential-capacity operations and three browser PostgreSQL
  rows) manages or attaches to its services itself and has no residents, as
  in the v8 model.

Tests: `TestPlannerCompilation` adds canonical-unit, identity-alone and
policy-evidence cases and checks lease-bound residents; `TestPlanAssembly`
validates a plan against the generated v8 validator and rejects thirteen
invalid invocations, each on its intended invariant. Checks: `make lint
CHECK=go`; the planner, JSON-shape and native-runtime rows; the graph-contract
row (`harness.test_catalog.behavior.routing_contract`); `source-check`
`generated`, `schema` and `catalog`; `lint-markdown`; all pass. The
owner-boundaries row passes its owner-schema test but fails on private browser
imports in `installed-readiness.mjs` from `a5df06952` (on main, unrelated;
tracked as its own task).

**Milestone 2, slice 2c: preparation in Go (2026-11-23).**
`harnessplan.Prepare(root, request, now, pid)` turns a graph invocation into a
validated plan without Node. The request carries the runtime's environment, so
the planner reads nothing from the process environment.

- *Selection.* `Selection` is now rows plus targets (operations or private
  unions) plus inputs forwarded to the selected operations. A preparation
  layer resolves requests into it: `command` through the command's selection
  binding (owner slices with OWNER and ROWS, fixed and input-keyed private
  selections flattened to rows and operations), `target` (operation, union,
  or a public operation command with its argv inputs forwarded), `owner` and
  `rows`. It also applies `explicit_only` placement, SCOPE=services and the
  GO_SHUFFLE_SEED Go-only rule. `make test-slice` reaches the runtime as a
  `command` selection, so this is what Go owner slices need.
- *Inputs.* The task-surface owner is loaded with the planning inputs, and the
  topology's browser lane policy is projected onto every BROWSER_LANES input.
  The resolver applies Make origins, allowed sources, defaults, empty-string
  rules, normalization, the seven input types graph commands declare (any
  other type fails closed) and input constraints. Public preflight keeps
  filesystem admission of path inputs. Declared inputs, redacted where
  declared, go to the manifest. The unit input environment clears every public
  input and forwards this command's runtime inputs. Cache, instrumentation and
  service mode come from the resolved inputs, not the raw environment.
- *Facts.* Source snapshot and provenance (`git ls-files` under the
  executable-input pathspec, documentation pruned by Git before enumeration,
  parallel hashing with change detection, nested repositories rejected); run
  identity; Go worker budget under a nested admission lease; owned or attached
  test services (attachment only for the selections the NLSpec names, with a
  session file that reaches no symbolic link); launch bindings; Go host
  profile; toolchain digest; instrumentation policy.
- *Not yet.* Cache facts are slice 2d: until then `Prepare` accepts only cache
  mode off and reports `ErrUnsupported` otherwise. Aggregates are milestone 3.
  Selections that admit `explicit_only` rows by name (the visual and UI-review
  suites) are a milestone 4 decision, since their rows are Playwright rows.
  `harnessdiag` keeps its own narrow input parser. Folding it into the
  planner's resolver is a later consolidation.

Tests: `TestInputResolution` (origins, sources, defaults, normalization,
types, constraints over a synthetic contract), `TestSourceSnapshot` (a
temporary Git repository: restricted roots, documentation, ignored, deleted,
untracked and symbolic-link entries, modes, clean and dirty state, nested
repository), `TestPlanPreparation` (a test-slice owner slice prepares a
schema-valid plan whose owner-request form compiles the same graph; a public
operation command; thirteen rejections, each on its intended rule), and a
forwarded-input case in `TestPlannerCompilation`. Checks: `make lint
CHECK=go`, the planner row (executed, cache miss), `source-check`
`generated` and `catalog`; all pass.

**Milestone 2, slice 2d: cache facts (2026-11-23).** `harnessplan.CacheFacts`
decides each unit's cache eligibility and, for an eligible unit, its input
digest and dependency closure; `Prepare` now plans in every cache mode.

- *Registry v3.* A unit-aware profile declares `unit_inputs` (`paths`, each
  with everything beneath it, and `prefixes`). That replaces the JavaScript
  closure's hard-coded file list, which now reads the profile too. The
  registry adds `contracts` wholesale. The JavaScript derivation of a contract
  root from the owner name is gone because it matched 21 of 66 owners and
  missed real contract directories. The NLSpec amendment records both changes.
- *Closures.* Broad profiles close over their registered roots. Go units close
  over their packages and every repository-local import found by parsing all
  Go files of each package (all build constraints, tests included) with
  `go/parser`, plus everything beneath those package directories, the
  enclosing module's `go.mod` and `go.sum` (a general rule replacing the
  referencepacks special case), the profile's unit inputs, the selected
  owners' manifests and the rows' fixture roots. Incomplete knowledge (an
  unresolved local import, a changed or missing file) selects the profile's
  broad closure as `broad_fallback`, as the NLSpec requires.
- *Keys.* The input digest binds the profile, strategy, unit digest, closure
  digest, platform, toolchain digest and inherited cryptographic environment.
  The helper-ownership digest is dropped: unit-aware closures already contain
  that file, and broad profiles declare their roots.
- *Not ported.* Vulnerability-database revision proof. The only profile that
  requires it covers `go_vulncheck`, which declares no reusable outputs and so
  is never reusable. The Go planner reports `freshness_unknown` for such
  profiles, so the scan always runs, until a scan gains reusable outputs and
  needs the proof.
- *Authoring fix.* `harness.projection_sql_check` declared `content_addressed`
  with no cache profile. Its only shell row ran uncached, and compiling the
  operation directly would have failed in the JavaScript planner. The policy
  now declares `none`, and `Load` rejects a cacheable policy unit without a
  matching profile, an operation bound twice, and unit-aware profiles without
  unit inputs.
- *Cost.* Closure digests are assembled from each snapshot entry's canonical
  JSON, encoded once, instead of re-serializing every closure. Facts for all
  452 plannable units take 0.45 s (2.8 s before); the snapshot of 9,581 files
  takes 0.12 s.

Tests: `TestCacheFacts` (a synthetic module: exact closure membership,
keys that change only with closure changes, broad fallback, nested-module
manifests, closure digest equal to the generic semantic digest, every
ineligibility reason, unbound cacheable units), three new `TestPlannerInputs`
mutations, and normal-mode cases in `TestPlanPreparation` (the planner row's
closure holds its packages, imports, unit inputs and owner manifest and no
unrelated product packages; deterministic keys; the scan is never reused).
Checks: `make lint CHECK=go`; the planner, scheduler-contract, JSON-shape and
native-runtime rows; the graph-contract row; `source-check` `generated`,
`schema` and `catalog`; `lint-markdown`; all pass.

Open items found here:

- *Per-unit freshness re-hashing.* The runtime re-hashes every file of an
  eligible unit's closure at lookup. Across the 227 eligible units above that
  is about 783,000 file checks, mostly the same shared inputs, although the
  source snapshot was hashed moments earlier. 2e should verify each snapshot
  file at most once per run.
- *Tests that read authored data outside their closure.* Some Go tests read
  repository files they do not import. The planner's own tests read every
  family manifest, but the closure holds only the selected owner's, so a cached
  result can be reused after another owner's manifest changes. The NLSpec lets
  rows declare extra roots only under `fixtures/`. The structural fix is a row
  declaration of data roots anywhere in the repository, which needs a spec
  change.

**Milestone 2, slice 2e: the runtime runs Go plans (2026-11-23).** With
`CARTULARY_HARNESS_PLANNER=go`, the runtime plans the invocation with
`harnessplan.Prepare` in-process. There is no compiler process, descriptor or
plan decode on that path. The JavaScript compiler stays the default. A
selection the Go planner does not support fails as a usage error; it never
falls back. Units do not inherit the selector.

- *Adoption.* `adoptPlan` maps the Go plan onto the runtime's execution model
  once. The plan itself is the published run-root document (`plan.json`, in
  place of `execution-context.json`). Runtime binaries are bound at adoption
  (`Plan.LaunchEnvironment`): a unit's listed binaries get their paths and
  every other binary is cleared. Planner rejections become the same closed
  preparation diagnostic the JavaScript path renders. Accounted runs impose
  their cache and instrumentation modes through the request, not the
  environment.
- *Residents.* Host admission holds a unit's residents by the keys the plan
  names (`WorkRequirement.Residents`). It derives them from `Physical` in one
  place only for JavaScript-planned units; that path is deleted at the switch.
  The planner marks the shared generation and object store `attached` in
  attach mode; admission does not charge attached residents.
- *Children.* Go units carry their rows in the command (`--row-ids`), so the
  row runner never reads or revalidates the published graph. Every argument
  is bounded at 32 KiB (the longest today is 5 KB).
- *Inputs.* Inputs that preparation turns into launch bindings (service mode,
  session file) are not forwarded. The JavaScript planner rejects an explicit
  session file as a credential input, which makes that documented input
  unusable there.
- *Freshness.* Cache lookups read a closure file again only when its identity
  (device, inode, mode, size, modification and change times) differs from its
  last proof in the run. An identity is remembered only when the file's change
  time precedes the proof by more than one second (Git's racy-timestamp rule),
  so a same-size rewrite within one timestamp tick is still detected. A first
  attempt that remembered every proof broke `input_change_blocks_reuse` and
  was replaced. The rule applies to both planners.

Runs: Go-planned `test-slice` for the planner row (cache miss, then a hit in
104 ms), `module.reportcomposition` (exclusive and shared PostgreSQL
generations and the shared object store held by planner keys; all five rows
pass) and `platform.objectstore` (all rows pass; the second run reuses all
three cacheable units). The default JavaScript-planned `module.reportcomposition`
still passes through derived residents. Tests: `TestGoPlannerAdoption`
(adoption, residents, row arguments, binary binding, attach-mode scopes,
request parsing, diagnostics), `TestRuntimeCacheFreshness`,
`TestPhysicalResidents`, and planner cases for attached scope, row arguments,
launch binding, the argument bound and imposed modes. Checks: `make lint
CHECK=go`; the planner, native-runtime, scheduler-smoke, native-fixtures and
fixture-broker rows; `source-check` `generated`, `schema` and `catalog`;
`lint-markdown`; all pass.

Milestone 2 is complete. Open items carried forward:

- Post-run JavaScript readers (`explain-run`, retained checks, agent-finalize
  retained runs) read `execution-context.json` and do not yet accept a
  Go-planned run's `plan.json`. They move to the plan in Stage 2.
- Attach mode is covered by tests, not by a run against a live local service
  session.
- The JavaScript row runner keeps `--unit-rows` for JavaScript plans until
  the switch.

**Milestone 3 plan.** 3a: Vitest rows in the Go planner. 3b: aggregate
selections (`test-fast`, `check`) with sub-target projections. 3c: remove
TypeScript title parsing from the JavaScript catalog and runner paths, and
make frontend completeness file-level. `test-fast` and `check` hold Go, shell
and Vitest rows plus policy roots; no Playwright or measurement rows sit in
those tiers.

**Milestone 3, slice 3a: Vitest in the Go planner (2026-11-23).**

- *Titles.* 89 of 758 Vitest rows authored a leaf title that the TypeScript
  parser expanded to the full title. They now carry the exact full titles,
  produced once by the existing resolver and ASCII-sorted, so resolution is
  the identity for every row. This makes the NLSpec's "rows carry exact full
  titles" true.
- *Authored needs.* The Vitest invocation policy declares
  `preparation_operations` (`harness.frontend_install`) and `unit_target_ms`
  (120 s). The four `renderer-node` integration rows declare
  `harness.playwright_install`. The JavaScript planner's two rules that
  derived these from runner and project names are replaced by reads of the
  authored data. Owner v11 was amended in place; it exists only on this
  branch.
- *Units.* Go and Vitest share one compilation path. A contract (runner, Go
  package or Vitest project, profiles, fixture, services, binaries,
  preparations, cache, CPU, isolation) groups rows. Indivisible items (a Go
  row, or a Vitest file's rows) are packed longest-first into as many
  balanced units as the work needs at the runner's target. Vitest units run
  one worker, claim the row profile plus the coordinator, and set
  `CARTULARY_TEST_TARGET=frontend-unit`. A unit's owner is its rows' common
  owner, else `harness.command_surface`; the former fixed `harness.backend`
  for Go units is gone.
- *Cache.* The unit-aware profile is split per runner: `test_rows` (Go) and
  `vitest_rows`, whose unit inputs add the Vitest runner's configuration and
  setup (`apps/web/vite.config.ts`, `apps/web/src/testing`, the app's
  manifests). A Vitest closure is the workspaces holding its files plus their
  declared workspace dependencies, read from `pnpm-workspace.yaml` and each
  `package.json`; it needs no source parsing. The JavaScript closure for a
  `packages/*` test omits the runner configuration it loads, so a
  configuration change can reuse a stale result there.
- *Plan contract.* Graph v9 allowed 4096-byte arguments (inherited from v8)
  while the planner allowed 32 KiB. The first `web.application` run failed at
  publication, reported as a `fixture_error`. The schema now allows 32 KiB,
  and `Assemble` validates the plan against its schema, so a contract
  violation is a planner error.

Runs (Go planner): `package.view_contracts` (one Vitest unit, all five rows
pass, four with migrated titles), `module.indicators` (Go and Vitest),
`web.application` (two Vitest units of 35 and 60 rows; all 95 pass). The
JavaScript-planned `package.view_contracts` also passes with the migrated
titles. Tests: Vitest compilation (one unit per file within a contract,
authored preparations, coordinator claims, project purity), runner-generic
balance and row arguments, workspace closures over a synthetic repository
(transitive declared dependencies, unrelated workspaces excluded, broad
fallback outside every workspace), and preparing `web.application` within the
schema. Checks: `make lint CHECK=go`; the planner, scheduler-contract and
JSON-shape rows; every `harness.test_catalog` row; `source-check`
`generated`, `schema` and `catalog`; all pass. The owner-boundaries row fails
only on the known `installed-readiness.mjs` import (task_77d37cee); its
owner-schema test passes.

**Milestone 3, slice 3b: aggregates (2026-11-23).**

- *Selection.* `--selection aggregate` (and a `target` request for an
  aggregate) selects every row placed at or below the tier plus the policy
  roots of the aggregate and every aggregate it inherits.
- *One packing pass.* Compilation now first walks the selection for every row
  it reaches: selected rows, union members, policies' required rows, and rows
  reached through the operations rows and policies wait for. It packs all
  Go and Vitest rows once, then builds units on demand. Before, a row reached
  twice (for example in the tier and in a policy's required rows) was compiled
  in separate calls and could land in two units.
- *Projections.* Command and aggregate plans project row suites
  (`backend-<suite>`, `frontend-unit`), every operation, and every other
  public command and gate smoke check whose own plan the graph contains:
  every candidate row is bound here, and every candidate operation unit is
  here with an identical digest. Each projection is closed under needs.
  Containment by row binding replaces the JavaScript planner's comparison of
  separately packed units, which could never match a packed row; a
  gate-smoke check that names a row is now projected whenever its row runs.
  Go integration-support rows report under `backend-integration`, the suite
  `SUITE=backend-integration` selects.

Runs (Go planner, `CARTULARY_HARNESS_PLANNER=go`):

- `make test-fast` passed: 180 units (168 Go, 8 Vitest, 4 operations), all
  1213 rows passed in 113.5 s graph time. Every cacheable unit missed because
  the Go planner's cache keys are new.
- `make check` passed: 367 units (311 Go, 13 Vitest, 43 operations) and 58
  projections, including the contained public commands (`lint`,
  `go-vulncheck`, `go-gosec-targeted`, `go-error-declaration-check`,
  `migration-drift`, `otel-conformance`) and the three gate smoke checks. All
  1891 rows passed in 856.6 s graph time (14 min 24 s for the whole command),
  with 144 cache hits partly warmed by the `test-fast` run.

Preparing the `check` plan in Go takes 0.89 s (6.5 MB, 197 cache-eligible
units). The Go planner is judged on correctness and, later, efficiency, never
by comparison with the JavaScript planner, which is being removed (owner,
2026-11-23).

Tests: aggregate preparation for `test-fast` and `check` (every tier row
bound, no explicit-only row, a projection for every root operation, suite
projections, the aggregate's projection covering the graph); a row reached
both directly and through a union is bound once; `ci` reports unsupported.

**Milestone 3, slice 3c: no TypeScript parsing for Vitest (2026-11-23).**

- *Titles pass through.* The catalog admits a Vitest row by its file (contained
  in the runner's approved roots, no symbolic links), its registered project,
  and the project that runs that file. It does not read the source.
  `resolveVitestTitles` is deleted; the JavaScript planner and the Vitest
  runner use the authored titles. The runner selects them by an anchored
  exact-name pattern and fails a row whose title it does not collect exactly
  once (`missing or ambiguous observations`), so a misauthored title fails
  when its file runs. Overlap detection still compares exact full titles.
- *File-level completeness.* Every discovered Vitest test file has a catalog
  row. A test added to a file that already has a row is no longer detected,
  the risk recorded with the Stage 1 scope.
- *Tests.* The catalog assertions replace the Vitest parser cases with
  pass-through, project-mismatch and file-level coverage cases. The Vitest
  adapter fixture rows author exact full titles (`scope one`).
- *Playwright stays parsed until milestone 4.* Its reporter needs title paths
  split into suite components for Playwright's exact test-list interface, and
  only the parser supplies them today. Milestone 4 moves title paths into the
  catalog and deletes the parser with the browser work.

Checks: the registry-parity row (Vitest adapter and catalog assertions),
every `harness.test_catalog` row, JavaScript- and Go-planned `web.application`
(95 Vitest rows), `make lint CHECK=scripts`, `source-check` `catalog` and
`generated` (after regenerating the render index); all pass. The
owner-boundaries row fails only on the known `installed-readiness.mjs` import.
Loading the complete catalog takes 0.95 s.

Milestone 3 is complete: the Go planner plans `test-fast` and `check`.

**Milestone 4 plan.** 4a: remaining command selections. 4b: Playwright title
paths in the catalog and deletion of the TypeScript parser. 4c: browser
stages in Go (groups, lane assignment, finalizers, renderer residents, group
units that carry their rows instead of reading the published graph). 4d:
measurement rows and performance fixtures. 4e: `ci` and `release-check` end
to end.

**Milestone 4, slice 4a: commands (2026-11-23).** A probe of every graph
command and selection variant found three kinds of failure, now resolved or
assigned:

- *Silently empty visual plans.* The Go command selection dropped
  explicit-only rows unless OWNER and ROWS were given. `browser-e2e
  SUITE=visual` and `browser-e2e-visual-update` therefore planned without
  their (explicit-only) visual rows. The JavaScript planner admitted them
  through a hard-coded list of three selection names. The owner model now
  declares `explicit_selections` (the two visual suites and
  `harness.contract_ui_review`, which each reach explicit-only rows). Both
  planners read it; any other bound selection still drops explicit-only rows
  (`SUITE=backend-integration` keeps excluding its five explicit-only Go
  rows).
- *Operations without policy units.* `format_frontend`, `format_go`,
  `harness_runtime_contract`, `harness_fixture_contract`,
  `protocol_ts_dead_code_check`, `source_surface`, `agent_finalize` and
  `benchmark_claim_check` now have policy units. Each operation that lists the
  frontend-install prerequisite needs `harness.frontend_install`. The two Go
  race-test contract suites take the harness-contract profile and estimate of
  their sibling suites; the rest take the standard profile and the 1000 ms
  that applied.
- *Browser and measurement work* (`ci`, `release-check`, `browser-e2e`
  suites, `harness-contract SUITE=artifact-consumers`, `test-slice
  SUITE=backend-integration`) is reported as unsupported until 4b–4d.
  `release-evidence-verify` and `test-evidence-audit` correctly refuse without
  their required inputs.

Tests: `TestCommandSelection` (explicit selections run explicit-only rows,
other selections drop them, an unknown explicit selection is rejected), and a
case that resolves and compiles every graph command and variant, accepting
only success or `ErrUnsupported` (1.1 s). Runs (Go planner): `lint
CHECK=protocol` and `source-check CHECK=surface`, which reach two of the new
policy units, pass. Checks: `make lint CHECK=go` and `CHECK=scripts`; the
planner, scheduler-contract, JSON-shape and graph-contract rows;
`source-check` `generated`, `schema` and `catalog`; all pass.

**Milestone 4, slice 4b: Playwright title paths (2026-11-23).** The harness no
longer parses TypeScript or JavaScript test sources.

- *Representation.* A Playwright title is its title path in Playwright's
  display form, components joined by ` › `. 709 of 710 cataloged Playwright
  titles were top-level tests, whose path is the title itself. The one nested
  test (`web.design.visual.account_menu_root_and_nested_viewport_states`)
  now authors `Account navigation presentation › Capture …`. The catalog keeps
  one field; no schema change.
- *Readers.* `playwrightSelection(rows)` splits titles into exact test-list
  paths for the cohort reporter, the browser group CLI and the benchmark claim
  check. Places that compare with Playwright's own test title (visual capture
  selection, visual reconciliation, UI-review import, the visual registry
  check) use the path's last part (`playwrightLeafTitle`), so they keep
  matching nested tests.
- *Catalog.* Playwright rows are admitted by file, registered project and
  stage, scenario pairing and valid title paths. A cataloged path that is a
  suite prefix of another in the same file and project is rejected. Coverage
  is per file. The runner already fails a path it does not report exactly once
  and any identity it runs outside its selection. Detecting imported
  registration suites needed source parsing, so that NLSpec clause is
  replaced by the runner's exact-identity accounting.
- *Deleted.* The TypeScript registration parser (about 400 lines of
  `selector-resolution.mjs`, its `typescript` loader and
  `resolvePlaywrightSelection`), and the generate-drift scratch copy's walk of
  registration-suite files. The metadata fixture test now asserts that loading
  the complete catalog reads no TypeScript source.

Runs: the migrated visual row in a real browser (`make test-slice
OWNER=web.design ROWS=…account_menu_root…`) passes through the split
test-list path and the leaf-title capture match. Checks: `make lint
CHECK=scripts`; every `harness.test_catalog` row; the registry-parity,
generate-drift and Playwright-adapter rows; `source-check` `generated` and
`catalog`; all pass.

**Milestone 4, slice 4c: browser stages in Go (2026-11-23).**

- *Inputs.* The planner reads the generated browser batch manifest
  (partitions and reporting stages), the topology's frontend artifacts and
  browser reset policy, and the owner's lifecycle estimates. The Playwright
  invocation policy now declares the preparations every browser group waits
  for (`build_migrate`, `build_server_harness`, `playwright_install`,
  `test_service_images`); the JavaScript planner reads them instead of its
  hard-coded list.
- *Cohorts and lanes.* Rows of one partition with the same needs form
  cohorts. Rows outside test-owned state run alone; otherwise balanced
  cohorts fit the resource profile's timeout less allocation headroom, and
  ordinary cohorts spread over `BROWSER_LANES`. Ordinary cohorts are placed
  longest first on the least-loaded lane and wait for the cohort before them.
  The lane chain stays a compile-time ordering for now. A capacity-based lane
  bound needs a new logical resource and host-admission work, so it belongs
  with the Stage 3 scheduler. Cohort names are the planner's own
  (`cohort-<digest>`). A stage that reports selected rows gets one target
  finalizer; snapshot updates report under `browser-e2e-visual-update`.
- *Units.* Group units carry their rows (`--rows`), renderer requirement,
  affinity key and stack key. The group CLI validates the cohort against its
  partition from those rows and hands the cohort to the Playwright reporter,
  so no browser child reads the published graph. The reporter now takes its
  cohort only from the CLI, for both planners.
- *Residents.* The planner names each group's stack (`stack:<digest>`) and
  the PostgreSQL cluster it owns (`postgres:<stack>`). The fixture broker uses
  the named key (`Request.StackKey`) instead of deriving one, so plan
  residents and broker allocations agree. Admission reads renderer-slot
  claims from the topology (cached once per run) when the plan names none.
- *Projections.* Playwright rows report under their stage's target, and each
  finalizer under its own.
- *Still unsupported.* Fixture-profile (measurement) browser groups and Go
  measurement rows (4d). `ci` now plans; `release-check` waits for 4d.

Runs (Go planner, real browsers): `browser-e2e SUITE=support` (4 groups on 4
lanes, 16 rows), `SUITE=stateful` (11 groups, 59 rows), `SUITE=accessibility`
(6 groups, 57 rows) and `SUITE=visual` (6 groups, 48 rows, explicit-only
visual rows included) all pass, with planner-named stacks and clusters. The
JavaScript-planned `SUITE=support` also passes with the changed group CLI and
reporter. Tests: browser plans for the functional, stateful and visual
suites and the snapshot update (group units with rows, renderer, affinity and
stack resident; lanes within `BROWSER_LANES`; finalizers with projections);
measurement work and `release-check` remain unsupported. Checks: `make lint
CHECK=go` and `CHECK=scripts`; the planner, native-runtime,
scheduler-contract, native-fixtures, fixture-broker, Playwright-adapter and
graph-contract rows; `source-check` `generated` and `schema`; all pass.

**Milestone 4, slice 4d: measurement rows and performance fixtures
(2026-11-23).** The planner now compiles every selection; `ErrUnsupported`
is gone, and the runtime's usage-error mapping with it.

- *Fixture owner.* Routing a row to its populated fixture is the fixture
  owner's job (TH-HARNESS-REQ-812), so `performancefixture.ResolveRow` now
  does it in Go from the generated profiles' verification bindings. The
  planner consumes it and `performancefixture.SnapshotKey`; it implements no
  keying of its own. The migration digest is `pgschema.Hash()`: the
  migrated-template identity both harness binaries embed and already verify.
- *Load-time routing.* Every row is routed when the inputs load. A populated
  profile binds one Playwright row to exactly one predicate. The row's
  partition declares the same profile. The builder policy
  (`tools/performance_fixture_builder_policy.json`, now a planner input)
  covers exactly the active profiles. A Go row with a populated profile is
  rejected, since nothing runs a Go clone.
- *Units.* Each populated row is its own cohort; its cohort name includes the
  snapshot key. One builder per snapshot,
  `fixture_snapshot:<runtime>:<profile>:<key>`, waits only for the service
  images. The old planner made it wait for every browser preparation too. It
  holds shared host activity. Each measurement group needs its builder and
  carries the fixture, row and predicate. Each group gets a summary
  finalizer. The measurement target finalizer waits for the summaries and
  adds the aggregate. Builder, group and summary carry `fixture_profile_id`
  and `snapshot_key` (graph v9 amended in place).
- *Cluster authority.* The planner names every PostgreSQL cluster a fixture
  lease uses (`postgres:unit:<unit>`, `postgres:shared`, `postgres:<stack>`,
  `postgres:measurement:<digest>`) and passes it as
  `CARTULARY_POSTGRES_CLUSTER_KEY`. The broker allocates the named cluster
  (`Request.ClusterKey`) instead of deriving one. The builder and every clone
  hold the snapshot's cluster as one graph resident.
- *Go measurement rows* run alone (TH-HARNESS-REQ-865). A runner unit's
  timeout is now its resource profile's `runner_timeout_ms`: the backend
  capacity profiles author 3,600 s, but runner units took the 600 s default.
  That was shorter than the rows' 1,800 s measurement estimate, and the old
  planner had the same gap.
- *Closures.* Importing the migration owner exposed an over-approximation:
  the Go closure followed the test-file imports of every imported package.
  That pulled whole modules (for example `internal/modules/artifacts`, via the
  migration owner's tests) into unrelated cache keys. Go never compiles a
  dependency's tests into a dependent's test binary, so imported packages'
  `_test.go` files and their imports are now left out. Selected packages keep
  theirs.

Runs (Go planner): `browser-e2e SUITE=measurement` passes in 2,458 s. Four
builders ran, one per profile (128 s each; 437 s for the evidence-heavy
profile). Then 33 quiet groups and 29 summaries ran, and the aggregate is
`qualified` across four profile groups. The builders' shared host activity
holds off every quiet session until the slowest builder finishes, so about
430 s of the run is builder-only. That ordering is spec-mandated; a later
scheduler may still overlap one profile's measurements with another
profile's build, if quietness allows it. `test-slice
SUITE=backend-integration` ran its PostgreSQL units under planner-named
clusters, and its Go measurement row ran alone. Two `full`-tier rows failed;
neither failure comes from this slice:
`app.server.integration.application_fixtures_own_independent_databases…`
fails identically under the JavaScript planner. Since `500e3bdda` a
unit-owned exclusive generation no longer drops per-test databases eagerly,
and the test still asserts an eager drop. Tracked separately.
`harness.browser.integration.native_fixture_lifecycle` compared a
same-run frontend receipt with the source digest after this session edited
files mid-run; on a stable tree it passes. Tests: measurement planning (one
builder per profile, keyed by the owner, resident on the snapshot cluster;
one group and summary per row; the aggregate), Go measurement rows alone
within their timeout, load-time routing rejections, the fixture owner's
`TestRowFixtureRouting`, and `release-check` among the prepared aggregates.
Checks: `make lint CHECK=go`; the planner, native-runtime, scheduler-smoke,
native-fixtures, fixture-broker, nested-cancellation and fixture-owner rows;
`source-check` `catalog`, `schema` and `generated`; `make lint-markdown`;
all pass.

### Slice 4e: aggregate runs (blocked, then fixed in isolation)

The Go-planned `release-check` graph (586 units; every `ci` unit with an
identical digest, except the measurement target finalizer) was run once to
cover both aggregates. It deadlocked 11 minutes in and sat idle for 5 hours:
56 units had passed, none was running, and nothing was printed.

- *Cause.* The host admission ledger still charged seven PostgreSQL residents
  and one object-store resident from a run stopped the day before
  (`otel-check01`). A stop by the agent's task stop kills the harness
  mid-cancel; its runtime, containers and ledger residents stay until
  `make test-run-recover`. The first measurement group's exclusive request
  needed 17 of 16 PostgreSQL lanes and was re-filed for 5 hours, about 3.7
  times a second. Retirement can reclaim only stacks and renderers. The
  scheduler's exclusive barrier held the other 418 ready units, and the
  measurement pause froze live status.
- *Startup recovery.* A top-level graph now recovers every dead, unsealed
  runtime of its repository before admitting work, through the same admission
  and cleanup evaluator as `test-run-recover`. It reports `[RECOVERED]` or a
  `[RECOVERY-FAIL]` line with the exact recovery command. Sealed roots have
  already drained their resources and stay for explicit recovery (eight such
  roots from October test runs exist; they hold no capacity).
- *Unsatisfiable admission.* The ledger (generation 4) records which process
  first charged each resident. A waiter fails with `infra/resource_conflict`
  when its persistent deficit remains after removing every resident a live
  foreign owner could release, its own owner holds no other active
  admission, and its pressure retirement can neither release an allocation
  nor finds one deferred by another request's pin. The scheduler prints
  `[ADMISSION-UNSATISFIABLE]`.
- *Known failures fixed first.* Installed readiness imported two private
  browser modules. Three browser modules imported the test catalog's private
  selector module, a slice 4b regression. The application fixture test still
  asserted an eager drop where its database retires with its generation.

Runs and checks: the owner-boundaries row, `harness-contract SUITE=ui-review`
and the application fixture row pass (Go planner). `make lint CHECK=go`;
`source-check` `catalog` and `generated`; the admission, native-session,
native-fixtures, fixture-broker, nested-cancellation, native-runtime and
scheduler-smoke rows; `make lint-markdown`: all pass.

**Rerun (2026-11-24, `73ce7cb85`, run `20261124T112817Z-p93015`).** The
graph ran to completion in 78.6 minutes, with no startup recovery and no
unsatisfiable admission. 582 of 586 units passed; three failed, and
`release_readiness_evidence` was skipped behind them.

- *`frontend_fallow_static`: pre-existing.* Five unused exports in
  `apps/web` (`reconnectBaseDelayMs`, `reconnectMaxDelayMs`,
  `publicErrorCode`, `publicErrorStatusText`, `HistoryActionOperation`)
  arrived on `main` in `353e4a99f`. This branch changes no application code.
- *`harness_scheduler_contract`: shared tool rebuilt mid-run.*
  `content_cache_security_and_closure` got `broad_fallback` where it expects
  `go_packages`; the row passes alone. The failing resolution ran at
  12:30:15Z, the second at which a nested Make invocation inside the run
  rebuilt the shared `.cache/cartulary/toolbin/cartulary-harness` with
  different bytes. The resolver executes that binary (`inspect-go`) and turns
  any error into the broad fallback. Nested Make rebuilt the binary four
  times during the run (12:29:27Z, 12:30:15Z, 12:45:59Z and 12:46:00Z), under
  four cache keys. Twenty probe resolutions beside a cold server build and
  the routing contract did not reproduce the failure. The fault class is a
  run's nested commands replacing shared tool binaries that the run's units
  are executing.
- *`seaweedfs_release_gate`: two faults.* The occurrence inventory reads every
  repository source through a 64 MiB bound, and three tracked reference-pack
  containers are 95 MB (`main`, `8501d4449`), so generation throws before
  writing a summary (pre-existing). Its recovery admission reads
  `execution-context.json`, which a Go-planned run does not publish (caused
  by the planner switch; a Stage 2b reader).

**Fixes, tested in isolation (2026-11-24).**

- *Fallow.* The five declarations are module-private (`export` removed).
- *Atomic Go artifacts.* Units receive a private `GOTMPDIR`, memory-backed
  under `/dev/shm` when eligible. A nested `go build -o` into the repository
  then crosses filesystems, so Go removes and rewrites the published binary,
  and an exec during the copy fails. A scratch reproduction counted 130
  failed execs out of 3,685 across six in-place rebuilds, and none when
  building beside the output and renaming. `build-go-artifact.sh` now
  publishes that way, and the cryptographic receipt is renamed into place.
  The Go launcher on the host was 1.27.1 and switched to the pinned 1.27.4
  from a module cache, which builds embed as GOROOT; the launcher is now the
  pinned toolchain itself, and builds with disk and memory-backed temporary
  roots are byte-identical. Nested rebuilds stay allowed: real-command tests
  deliberately run the top-level path, and a rebuild can no longer break an
  exec or change the bytes.
- *SeaweedFS occurrence scan.* `readRetainedTextBytes` streams a source and
  returns null at the first NUL byte whatever the size; only text the scan
  keeps must fit 64 MiB.
- *SeaweedFS recovery admission.* It reads the run's one plan document:
  `plan.json` (schema, recomputed graph digest and manifest binding) or, until
  Stage 2b, `execution-context.json`.

Checks: `frontend-fallow-static`; `lint` `frontend`, `types`, `scripts` and
`shell`; the artifact-assurance and SeaweedFS evidence rows;
`harness-contract SUITE=regression`; and a Go-planned
`seaweedfs-release-gate`, whose ten checks all pass. The scheduler-contract
race is not reproducible on demand; the scratch reproduction is its evidence.

### Stage 2 scope: Go planner by default, JavaScript planner deleted

**Entry condition.** The Go-planned `release-check` (slice 4e, which also
covers `ci`) passes, or fails only on causes outside planning. Every selection
kind already plans in Go; `ErrUnsupported` is gone.

**Current state (2026-11-24 inventory).**

- *Selector.* `CARTULARY_HARNESS_PLANNER=go` chooses `prepareInProcess` in
  `harnessruntime.Run`. Otherwise `Prepare` spawns Node
  (`tools/harness/scheduler/work-graph/prepare-cli.mjs`) and `plan.go`
  adopts its v7 plan and publishes `execution-context.json`. Nested graphs
  never inherit the selector, so every nested command still plans in
  JavaScript.
- *Planner package.* `tools/harness/scheduler/work-graph` (20 modules, 3,317
  lines; 9 test files, 2,187 lines), plus planner-only helpers outside it:
  `backend/go-lpt-shards.mjs`, `browser/browser-cohorts.mjs` and
  `scheduler/adapters/browser.mjs`.
- *Not planner code that lives in the package.* `operations.mjs` (operation
  definitions, environments and digests) is run-time dispatch for
  `operation-runner-cli`, the shell runner, agent-finalize and three
  generators. `policy-rows.mjs` serves the OTel conformance check.
  `capability.mjs` gives JavaScript admission clients (the local service
  session and the UI-review adapter) their host capacities. `cohorts.mjs`
  `rowEstimateMs` serves the estimate audit.
- *Recompiling readers.* `release-evidence-verify`, `agent-finalize`
  retained runs and `test-evidence-audit` rebuild a run's plan with
  `compileRetainedRun` and compare it to the retained graph.
- *JS-plan readers.* `explain-run`, retained checks, canonical context and the
  row runner's `--unit-rows` path read `execution-context.json` or revalidate
  the v8 graph (`validateWorkGraph`). The SeaweedFS release gate's recovery
  admission already reads `plan.json` and keeps an `execution-context.json`
  branch for JavaScript-planned runs.
- *Planning-only consumers.* The `target-plan` diagnostic and the
  task-surface report's scheduler-growth budget call `WorkGraphCompiler`.
- *Go code that exists only for JavaScript plans.* `harnessruntime`
  `Prepare`/`receiveCompiler` and the v7 adoption in `plan.go` (about 620
  lines with `prepare_linux.go`), the `prepare` subcommand, resident
  derivation from `Physical` (`physicalResidents`), the broker's cluster-key
  derivation when no planner key is given, and the JS row runner's
  `--unit-rows` mode.
- *Tests.* 36 JavaScript test files import the planner. Eleven operations run
  them (graph, scheduler, command-surface and boundary contracts, browser lane
  policy, UI-review lifecycle, artifacts and presentation, execution topology,
  release readiness and object-store evidence). Retained-run fixtures
  (`retained-fixture.mjs`, `retained-finalizer-cases.mjs`,
  `canonical_test.go`, the selected-services boundary test) build
  `execution-context.json` runs.

**Slices.**

1. *2a: Go entry points for the remaining consumers.*
   `cartulary-harness plan <graph arguments>` prints the plan v8 document the
   runtime would execute, for the `target-plan` diagnostic.
   `cartulary-harness capacities` prints the capability snapshot and scheduler
   capacities through `harnessplan.Capture`, for the local service session
   and the UI-review adapter. The scheduler-growth budget becomes a Go test
   over a `release-check` plan with a fixed capability snapshot, and leaves the
   task-surface report. `policy-rows.mjs`, and `rowEstimateMs` with the
   estimate audit, move to their consuming owners; `operations.mjs` moves to
   the execution owner unchanged. No behaviour change.
2. *2b: Cutover.* `Run` always plans in Go; the selector, `Prepare`, v7
   adoption, the `prepare` subcommand, `execution-context.json` publication,
   resident derivation and cluster-key derivation are deleted. Every reader
   takes `plan.json`, validated against the manifest's graph digest:
   `explain-run`, retained checks, canonical context, agent-finalize,
   `test-evidence-audit` and `release-evidence-verify`; the SeaweedFS gate drops
   its `execution-context.json` branch.
   Recompilation is removed, as the NLSpec amendment already requires. The row
   runner drops `--unit-rows` and per-child graph revalidation. Retained-run
   fixtures are re-authored as Go-planned runs. The NLSpec's
   execution-context section and the planner amendment are rewritten for the
   switch, and REQ-852's Node ownership clause names Go. Plan v8 and graph v9
   are closed at the switch.
3. *2c: Deletion.* `tools/harness/scheduler/work-graph`, the planner-only
   helpers, compile-time browser lane chaining, the plan v7, graph v8,
   execution-context v3 and shard-plan schemas, their generated validators
   and schema attachments, and the topology backing-script entries. The
   eleven contract operations keep only assertions about behaviour the Go
   planner or runtime owns; planner-internal JavaScript cases go with the
   planner. Helper-ownership and import-boundary manifests drop the removed
   owners.

**2b result (2026-11-24).** Done ahead of 2a, which only the diagnostic
and maintenance tools need. `Run` plans every invocation in Go; the selector,
`Prepare`/`receiveCompiler`, the v7 decoder, the `prepare` subcommand,
`prepare-cli.mjs`, `prepare-plan.mjs`, `compileRetainedRun`, the pipe
diagnostic and `unit-selection.mjs` are deleted. One Node reader
(`observability/retained-plan.mjs`) binds `plan.json` to the run manifest
for retained checks, agent-finalize, the owner audit, `release-evidence-verify`
and the SeaweedFS gate. `explain-run` needed no port: it reads no graph
fields. Admission takes residents only from the plan; `WorkRequirement.Physical`
became `RendererClaims`, and only the review owner's dedicated borrower names
its residents from the topology. The broker's cluster-key derivation stays:
the fixture owner still uses it for that borrower and for tests. While
porting, `plan.json` turned out to retain launch bindings (absolute binary
paths) and the input-forwarding policy, which REQ-278 forbids; the published
plan became `cartulary.harness_runtime_plan.v9` without them. Verified by the
native-runtime, planner, scheduler, fixture, review, boundary-lifecycle,
routing, evidence, schema, owner-boundary, registry, scheduler-contract,
package-readiness and SeaweedFS-evidence rows.

**Verification.** 2a: the planner, native-runtime and new-CLI rows, the
local-session and UI-review suites, `source-check` `generated`, `schema` and
`catalog`. 2b: each moved reader's rows, the retained-run and boundary rows,
then one diagnostic `make check`, because the default planner changes for
every command, nested ones included. 2c: the affected harness-contract suites,
`make generate`, the source checks and `lint`. `ci` and `release-check` are
not rerun for Stage 2; the 4e run already covers the Go plans they execute.

**Risks.**

- Nested commands inside contract tests switch planners at 2b. Assertions on
  JavaScript unit IDs, cohort names or shard files fail and must be rewritten
  against the Go plan or deleted, not preserved.
- `explain-run` (1,136 lines) reads graph v8 fields; it must be moved to v9
  (residents, no graph selector, three-field failure policy), not shimmed.
- Retained runs planned before the switch can no longer be verified or
  finalized. Release evidence already requires the current source digest, so
  no current acceptance path depends on them.
- Cache identity breaks once for default (non-selector) runs.

**Not in Stage 2.** Scheduler redesign (one admission decision, capacity-based
browser lanes, overlapping measurement with builds): Stage 3.

### Wall-clock: quiet work last

The owner set wall-clock reduction ahead of further expensive runs (2026-11-24):
targeted tests verify each change, and `release-check` reruns only after real
progress.

**Profile of the 4e run (4,717 s).** The 35 units holding `host_activity`
exclusively (33 measurement sessions, the credential-capacity operation and the
measurement target summary) became exclusive barriers as soon as their
dependencies finished. Four measurement sessions need no snapshot and were
ready at 120 s, which barred all shared work; their lock then waited for the
running shared units to drain, the last being the 569 s evidence-heavy snapshot
build. The host ran 2–17 units on about 2 CPU tokens from 120 s to 576 s. The
sessions then ran one at a time until 2,601 s (742 s of it fixture
acquisition, 69 s release, 1,205 s Playwright). Behind the barrier, 551 shared
units aged about 2,600 s each, so after measurement aging, not critical-path
rank, ordered them: the webserver-backed browser chain, the critical path,
waited 789 s for CPU behind the Go units. Measurement estimates were 30 times
the actual durations (`evidence_estimates_ms.measurement` 1,800,000 ms against
a 36.5 s mean runner time); Go, Vitest and browser estimates were within
1.0–1.7 times.

**Simulation.** A replay of the recorded unit durations under the engine's
logical rules (ranked choice with backfill, claims, shared and exclusive locks,
aging) gives 4,961 s for the current rules and 4,295–4,304 s with quiet work
last (−13%). Correcting the estimate alone changes nothing, because the
barrier ignores rank; a rank-aware barrier ran longer (5,268–5,803 s) because it
interleaves drains. The simulation ignores host admission and contention.

**Decision and change (A21).** A unit holding `host_activity` exclusively
becomes ready at the first scheduling decision at which its dependencies are
satisfied and no shared `host_activity` unit is ready, and stays ready; the
no-backfill barrier is unchanged. Release is sticky so the measurement summaries
that each session unblocks wait behind the released block instead of forcing a
drain per session. The engine's `ready()` holds unreleased quiet units back and
records the release; eligibility and aging start there. The measurement class
estimate is now 40,000 ms.

Checks: `TestRuntimeScheduling` (new cases: quiet work waits for ordinary work
then bars later shared work; quiet units released together run as one block),
the native-runtime and Go-planner rows, the graph-contract row,
`make lint CHECK=go`, `source-check CHECK=generated` and `lint-markdown`: all
pass.

**Next wall-clock levers.** The quiet block itself (about 2,000 s, serial):
per-session snapshot clone and bucket restore (about 22 s each) and session
start with traffic stabilization (about 20 s for a 153 ms single observation).
The webserver-backed chain (pure dependency path 1,447 s; lane loads up to
1,005 s on four lanes).

## 9. Decisions requested

- Decided 2026-11-23: Option B (staged Go replacement); Stage 0 tried and
  reverted; TypeScript source parsing removed from the NLSpec; Node kept only
  for running Vitest and Playwright.
- Open: whether the single-authority scheduler (Stage 3) moves ahead of the
  planner milestones, since it holds the measured wall-time cost and the
  planner rewrite alone does not remove it.
