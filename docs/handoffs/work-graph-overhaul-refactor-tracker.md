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

**Remaining milestone 2 slices.** 2c: preparation facts in Go (task-surface
command inputs and input environment, run identity, source snapshot and
provenance, launch bindings, host profile, Go worker budget) and a
`harnessplan.Prepare` entry point. 2d: cache facts (cache profiles, input
digests, `go list -deps` closures for Go units). 2e: the runtime calls the
planner in-process for Go-planned slices behind a planner selector, writes
`plan.json` to the run root, and reads residents by key; that removes the
compiler subprocess and the plan decode round trip on that path.

## 9. Decisions requested

- Decided 2026-11-23: Option B (staged Go replacement); Stage 0 tried and
  reverted; TypeScript source parsing removed from the NLSpec; Node kept only
  for running Vitest and Playwright.
- Open: whether the single-authority scheduler (Stage 3) moves ahead of the
  planner milestones, since it holds the measured wall-time cost and the
  planner rewrite alone does not remove it.
