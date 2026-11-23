# Work-Graph Overhaul Tracker

**Status.** Audit complete, 2026-11-23; no overhaul work started. The owner
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

## 6. Decisions requested

- Adopt Option B (staged Go replacement) or Option A.
- Start with Stage 0 (footprint-aware logical admission) before any rewrite.
