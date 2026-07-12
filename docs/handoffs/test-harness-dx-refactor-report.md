# Slice-harness developer-loop investigation

Status: investigation and implementation planning only  
Measured: 2026-07-11 through 2026-07-12, America/New_York  
Evidence class: developer-experience engineering evidence, not Core 05 benchmark or
claim-publication evidence

## 1. Executive conclusion

Cartulary can make individual slice runs materially faster without weakening evidence,
fixture, or cleanup rules, but the work should start with structure rather than higher
limits. The current slice scheduler is reasonably effective once it starts. The larger
avoidable cost is before it starts: setup targets are invoked serially, services are
then started by an outer wrapper, and the slice is re-executed before any schedulable
test work begins. Successful warm service-backed runs spent roughly 10--11 seconds
outside service readiness, the phase scheduler, and the final cleanup handoff.

The highest-value public developer-experience change is a namespace-aware `ROWS`
selector that supports exact base-phase evidence rows. Base slices are currently
phase-row scoped, but they cannot be narrowed below a phase. The frontend namespace
does accept `ROWS`, while omitted `ROWS` means cumulative implemented rows from
FE-P0 through the requested phase. The public behavior change requires a Testing
Harness NLSpec amendment and `cartulary.phase_slice_plan.v2`; it should preserve the
existing Make target names and command IDs. An exact partial selection must never be
reported as complete phase execution.

Before performance work on frontend browser slices, a current correctness defect must
be fixed. All four measured cumulative FE-P8 runs and all four exact FE-P8 browser-row
runs failed in the harness because selected-row accounting reached the non-evidence
`build-web` prerequisite. `build-web` then required each selected evidence row to map
to itself and rejected it. This contradicts the adopted behavior that executable
frontend slices select and enforce their frontend evidence rows. It is a private
implementation defect, not a reason to weaken row accounting.

The successful measurements lead to these conclusions:

* Warm median external wall time was 11.08 seconds for one FE-P8 Vitest row,
  41.83 seconds for Phase 4 service-backed, 39.99 seconds for Phase 9
  service-backed, 27.55 seconds for Phase 12 service-backed, and 41.13 seconds
  for full Phase 9.
* The selected phase scheduler itself was stable: its warm median was 9.852,
  25.579, 24.884, 12.942, and 25.213 seconds for those cases, respectively.
* In Phase 9, the browser work unit occupied essentially the entire scheduler
  critical path. Its nested Playwright test body ran three shards in parallel, but
  the outer scheduler saw one `process=1` unit.
* Phase 12 emitted 76 backend-integration Go processes for its 76 scenario-tagged
  backend-integration rows. The whole service-backed plan had 84 units. This is a
  future-growth and maintainability risk, but not the current wall-clock bottleneck:
  the warm scheduler median was 12.942 seconds and its browser unit was the slowest
  ordinary work unit.
* Three retained successful full warm runs had a median `check-service-backed`
  scheduler time of 119.499 seconds, service-session setup of 8.154 seconds, and
  cleanup of 4.271 seconds. `host_io` and `postgres_clone` were the dominant
  blockers there.
* The 24-CPU, 31-GiB WSL2 host retained at least 10.15 GiB available memory in all
  successful cold and warm cases. No measured invocation grew swap. More memory or
  blanket worker increases are therefore unlikely to be the first-order fix.

The recommended order is: repair frontend prerequisite scope; add the public exact-row
contract; normalize setup and service sessions into a scheduler-visible DAG; flatten
browser groups into truthful scheduler units; then A/B-test compatible scenario-shard
consolidation and fixture-policy changes. The acceptance threshold for a future
optimization is a warm median reduction of at least 20 percent or at least 5 seconds,
with all correctness and resource gates in Section 9 satisfied.

## 2. Current slice execution architecture

### Public entry and plan construction

The live path is:

1. Generated Make bindings in `tools/task_surface.generated.mk:425-431` expose
   `phase-slice` and `service-backed-slice` and pass `PHASE`, `PHASE_NAMESPACE`,
   `ROWS`, runner limits, and runtime paths to `RUN_MAKE_NODE_TOOL`.
2. `makeNodeTools["phase-slice"]` and
   `makeNodeTools["service-backed-slice"]` in
   `tools/harness/command-surface/make-node-tools.mjs:77-105,130-159` map both
   commands to `phase-slice-cli.mjs` and translate public inputs to arguments.
3. `parseArgs` and the namespace branch in
   `tools/harness/phase-accounting/phase-slice-cli.mjs:25-101` build either a base
   or frontend plan. The base call at line 93 does not pass `options.rows`; a base
   `ROWS` value is therefore ignored today rather than rejected or selected.
4. `buildPhaseSlicePlan` in
   `tools/harness/phase-accounting/phase-slice-plan.mjs:92-199` groups executable
   base rows by target, emits Go, Vitest, and browser units, adds finalizers, resolves
   resource limits, and validates the plan.
5. `buildFrontendPhaseSlicePlan` in
   `tools/harness/phase-accounting/frontend-phase-slice-plan.mjs:254-339` emits one
   Make-target unit per selected frontend target. This target granularity means a
   row selector constrains test and accounting selection where the child supports it,
   but a target-level check such as typecheck still runs the full target.

`cartulary.phase_slice_plan.v1` is defined in
`tools/harness/phase-accounting/phase-slice-plan-contract.mjs:6,119-190` and attached
at `docs/testing-harness-nlspec.md:1123,1160`. The work-unit object deliberately has
an open private extension zone, but the schema ID and its current namespace meaning
are public.

### Row scope

Current selection is not symmetric:

| Invocation | Current row scope | Completion meaning |
| ---------- | ----------------- | ------------------ |
| Base `phase-slice` | Every executable row in the one selected active base phase. `phaseRows` returns the phase guidance rows at `row-selection.mjs:16-35`. | Full selected base-phase execution, including incomplete claim status for blocked owner rows. |
| Base `service-backed-slice` | The same phase rows filtered by `executionDependencyInfo(...).service_backed`. | Full service-backed subset, not full phase execution. |
| Frontend slice without `ROWS` | Implemented rows in every active frontend phase from FE-P0 through the requested phase. See `rowsThroughSelectedActiveFrontendPhase` at `frontend/row-scope.mjs:54-73` and the branch at `frontend-phase-slice-plan.mjs:271-281`. | Cumulative frontend selection through the phase. |
| Frontend slice with `ROWS` | Exact supplied row IDs are syntax checked, searched in active phases through the requested phase, and required to be implemented. See `frontend/row-scope.mjs:75-112`. | Partial selected-row execution; target-level work can remain broader internally, but accounting must close only the selected rows. |

The current explicit frontend selector intentionally accepts rows from earlier phases
through the requested phase. The proposed v2 contract should simplify this: an
explicit selector belongs to exactly the requested namespace and phase; cross-phase
or cross-namespace IDs fail with usage exit 2. Omitted frontend input retains the
useful cumulative behavior. There is no continuing value in preserving implicit
base-`ROWS` ignore behavior or cross-phase explicit selection.

### Setup, services, scheduling, and nested runners

`phaseSliceSetupTargets` in
`tools/harness/scheduler/phase-slice-execution.mjs:107-136` derives
`frontend-install`, `build-server-harness`, runtime-binary producers,
`build-migrate`, and `test-service-images`. `runPhaseSliceSetup` at lines 148-156
runs these Make targets in a serial loop. `runPhaseSliceExecution` at lines 511-547
runs that loop before it invokes the service wrapper; `reexecPhaseSliceInsideServiceWrapper`
at lines 180-211 starts owned services and launches the CLI again with
`--inside-service-wrapper`. Only then does `runPhaseSliceScheduler` begin.

Inside the scheduler:

* Go shards use `goShardRuntimeCommand` and the backend target runner at
  `phase-slice-execution.mjs:330-357`. Go plan packing is owned by
  `tools/harness/backend/go-shard-plan.mjs`; `packShardLane` at lines 305-349
  creates a dedicated bin for every item with `scenario_id`.
* Frontend units call `tools/harness/execution/run-frontend-unit.sh` at
  `phase-slice-execution.mjs:360-369`. The public default is
  `VITEST_MAX_WORKERS=4` at `Makefile:20`.
* Browser units call `tools/harness/browser/run-browser-e2e-target.sh` at
  `phase-slice-execution.mjs:372-381`. `addBrowserUnit` in
  `phase-slice-planning/browser-work-units.mjs:93-120` emits one outer unit with
  `postgres=1`, `object_store=1`, `process=1`, `browser_stack=1`, and one browser
  stage claim.
* The inner browser batch resolves automatic functional shard concurrency from
  `PLAYWRIGHT_WORKERS` in
  `tools/harness/browser/run-playwright-webserver-batch.sh:87-109`, then launches
  up to that many one-worker Playwright processes in the background at lines
  319-384. The public default is three workers at `Makefile:18`.
* Frontend row scope is placed on a work-unit environment by
  `frontendRowAccountingEnv` at `phase-slice-execution.mjs:246-258`. This is the
  immediate path involved in the FE-P8 defect: `frontendRowsForAccountingTarget`
  rejects selected IDs not mapped to the current target at
  `frontend/row-scope.mjs:185-201`, while `browser-e2e-support` declares
  `build-web` as a prerequisite in
  `tools/execution_topology_manifest.json:2232-2244`.

Owned service setup, readiness, template creation, and cleanup are implemented in
`tools/testservices/main.go:332-465,2165-2284`. The wrapper emits
`service-scope.json`, fixture events, timing spans, lifecycle events, a lease, and
reaper output. Scheduler reporting emits `scheduler-summary.json`,
`scheduler-events.jsonl`, `pressure-summary.json`, target and phase summaries,
timing, finalizer records, and the public tool-run summary. Cleanup remains mandatory
after child completion; this investigation does not recommend persistent sharing or
removed reset boundaries.

### Generated and owner boundaries

`tools/execution_topology_manifest.json`, phase maps, frontend phase maps,
`tools/scheduler_resource_registry.json`, and duration baselines are owner inputs.
`tools/task_surface.generated.mk`, `tools/task_surface_manifest.json`,
`tools/scheduler_manifest.json`, and `tools/browser_e2e_batch_manifest.json` are
derived outputs and must only change through their owner inputs and Make generators.
The phase-slice target names and command IDs at
`docs/testing-harness-nlspec.md:529-530` should remain stable; no legacy alias or
second command family is needed.

## 3. Measurement results

### Method and environment

Existing retained artifacts were inspected before new execution. Diagnostics included
`make explain-phase`, `make explain-target`, `make explain-run`, fixture and pressure
data, scheduler summaries and events, target timing, child logs, service scope, and
lifecycle streams.

Each new principal case used one isolated cold invocation and three warm invocations.
For each case, the cold invocation used fresh external readiness, build-artifact, and
Go build-cache directories under `/tmp/cartulary-dx-refactor-20260712`; the warm
invocations reused those directories. The shared Go module cache and existing source
tree were not deleted. A cache can still report an intra-invocation hit when an earlier
prerequisite created the record. Owned Postgres and object-store services remained cold
per invocation because the adopted lifecycle starts and cleans an owned suite each
time.

External wall time, average aggregate CPU, and largest-child maximum RSS came from
`/usr/bin/time -v`. A one-second sampler recorded host `MemAvailable`, used swap, and
the sum of visible process RSS. GNU time does not expose peak aggregate CPU and its RSS
is not whole-tree RSS; peak CPU is therefore unavailable. The host RSS sampler includes
unrelated WSL processes and is context only. Scheduler summaries remain the timing
source of truth inside the harness.

The environment was WSL2 kernel `6.6.114.1-microsoft-standard-WSL2` on a Microsoft
hypervisor, with 24 logical CPUs, one thread per core, 31 GiB RAM, and 8 GiB swap.
The repository was on `/dev/sdd`, ext4. Docker Desktop 29.5.2 exposed 24 CPUs and
33,353,142,272 bytes of memory through overlayfs and cgroup v2. No explicit WSL cgroup
CPU or memory cap was observable. Process and open-file limits were 127,193 and
1,048,576.

### New measurements

Times in `median [range]` form use the three warm invocations. `Critical` is scheduler
critical-path wall time. `Fixture` is summed template-clone event duration, not an
additive wall-time bucket because clones overlap. `Wait` is unavailable in milliseconds:
the v4 pressure summary emits blocker observation counts, not blocked duration. `Test`
names the slowest ordinary scheduler unit; it includes that unit's nested setup and
teardown. `Cleanup` is from the applied cleanup start through the last cleanup event.

| Slice and selected inventory | Namespace | State/status | External wall | Critical | Readiness | Fixture/setup | Test | Wait | Cleanup | CPU and memory | Retained run root(s) |
| ---------------------------- | --------- | ------------ | ------------- | -------- | --------- | ------------- | ---- | ---- | ------- | -------------- | -------------------- |
| `phase-slice FE-P8 ROWS=FE-U-P8-01`; one exact row | frontend | cold/pass | 13.23s | 10.390s | none | serial readiness only | frontend-unit 9.995s | no blockers | none | avg 308%; child RSS 383 MiB | `.cartulary/test-results/20260712T020824Z-p79240` |
| Same | frontend | warm/pass | 11.08s [11.06--11.11] | 9.852s [9.849--9.861] | none | about 1.23s outside scheduler | frontend-unit 9.47s; Vitest body 9.15s in median run | no blockers | none | avg 347%; child RSS 374 MiB; min available 17.22 GiB | `.../20260712T020838Z-p80964`, `.../20260712T020849Z-p82392`, `.../20260712T020900Z-p83790` |
| `service-backed-slice phase4`; full service-backed phase subset | base | cold/pass | 72.15s | 41.076s | 2.929s | 54 clones, 7.784s sum | browser 37.109s | clone 189 samples | 1.446s | avg 1000%; child RSS 597 MiB | `.cartulary/test-results/20260712T020930Z-p85289` |
| Same | base | warm/pass | 41.83s [41.53--56.93] | 25.579s [25.541--26.158] | 4.259s [4.053--19.046] | 54 clones; 7.754s [7.561--13.348] sum; about 10s serial/residual | browser about 25.8s; Playwright body 14.857s in median-wall run | clone 181--188 samples | 1.669s [1.394--1.695] | avg 446%; child RSS 574 MiB; min available 12.50 GiB | `.../20260712T021042Z-p95428`, `.../20260712T021140Z-p12674`, `.../20260712T021222Z-p29196` |
| `service-backed-slice phase9`; service-backed rows | base | cold/pass | 63.37s | 35.498s | 2.839s | 29 clones, 4.781s sum | browser 35.329s | Go CPU 91 samples | 1.060s | avg 944%; child RSS 575 MiB | `.cartulary/test-results/20260712T021324Z-p45715` |
| Same | base | warm/pass | 39.99s [39.71--109.69] | 24.884s [24.650--25.476] | 3.006s [2.918--72.947] | 29 clones; 6.231s [5.766--6.255] sum; about 11s serial/residual | browser about 25.0s; Playwright body 13.751s in median-wall run | Go CPU 91 samples | 1.036s [1.010--1.060] | avg 305%; child RSS 575 MiB; min available 13.09 GiB | `.../20260712T021428Z-p56449`, `.../20260712T021618Z-p71520`, `.../20260712T021659Z-p86090` |
| `service-backed-slice phase12`; future-growth case | base | cold/pass | 62.47s | 33.971s | 3.274s | 7 observed clones, 0.754s sum | support Go shard 15.472s | clone 2,985 samples | 0.730s | avg 792%; child RSS 566 MiB | `.cartulary/test-results/20260712T021759Z-p993` |
| Same | base | warm/pass | 27.55s [27.28--28.56] | 12.942s [12.811--12.963] | 2.951s [2.895--3.199] | 7 observed clones; 0.777s [0.631--0.780] sum; about 11s serial/residual | browser 12.507s; Playwright body 1.094s in median run | clone 2,892--2,893 samples | 0.736s [0.708--0.789] | avg 504%; child RSS 558 MiB; min available 14.49 GiB | `.../20260712T021902Z-p82544`, `.../20260712T021931Z-p1233`, `.../20260712T021958Z-p19073` |
| `phase-slice phase9`; all 34 normalized phase rows | base | cold/pass | 64.50s | 36.554s | 2.906s | 29 clones, 5.801s sum | browser 36.395s | Go CPU 119 samples | 1.192s | avg 981%; child RSS 576 MiB | `.cartulary/test-results/20260712T022045Z-p37018` |
| Same | base | warm/pass | 41.13s [41.08--41.54] | 25.213s [25.202--25.621] | 3.969s [3.695--4.214] | 29 clones; 7.756s [7.715--15.651] sum; about 11s serial/residual | browser about 25.2s; Playwright body 14.447s in median-wall run | Go CPU 105 samples | 1.221s [1.190--1.286] | avg 327%; child RSS 575 MiB; min available 12.08 GiB | `.../20260712T022150Z-p49162`, `.../20260712T022232Z-p64299`, `.../20260712T022313Z-p79223` |
| `phase-slice FE-P8`; 57 cumulative implemented rows, 11 target units | frontend | cold/fail | 4m21.57s | 236.497s | 3.221s | 4 clones, 0.226s sum | stateful browser 118.054s | summary counts only | 0.822s | avg 144%; child RSS 822 MiB | `.cartulary/test-results/20260712T022425Z-p94470` |
| Same | frontend | warm/fail | 4m09.45s [4m09.40--4m09.97] | 237.650s [237.632--238.591] | 2.925s [2.848--3.138] | 4 clones; 0.218s [0.209--0.229] sum | stateful browser about 119.1s; later support target failed | summary counts only | 0.822s [0.803--0.848] | avg 108%; child RSS 826 MiB; min available 12.54 GiB | `.../20260712T022846Z-p36308`, `.../20260712T023256Z-p67854`, `.../20260712T023705Z-p99288` |
| `service-backed-slice FE-P8 ROWS=FE-B-P8-01`; exact row maps to two browser targets | frontend | cold/fail | 36.95s | 11.828s | 2.877s | 1 clone, 0.065s sum | webserver-backed 10.733s; support failed | summary counts only | 0.765s | avg 347%; child RSS 442 MiB | `.cartulary/test-results/20260712T024128Z-p30971` |
| Same | frontend | warm/fail | 23.30s [22.96--23.74] | 11.925s [11.878--11.940] | 2.956s [2.866--3.141] | 1 clone; 0.063s [0.062--0.065] sum | support failed before product test attribution | summary counts only | 0.717s [0.696--0.780] | avg 96%; child RSS 248 MiB; min available 13.55 GiB | `.../20260712T024205Z-p51478`, `.../20260712T024229Z-p61947`, `.../20260712T024252Z-p72161` |

The two failing cases are diagnostics, not performance baselines. Every failure named
`browser-e2e-support` and contained `selected frontend row id(s) not mapped to
build-web`; the cumulative case named `FE-B-P3-01,FE-B-P8-01,FE-S-P1-01`, while the
exact case named `FE-B-P8-01`. Product assertions were not the primary failure.

Two warm service outliers are real lifecycle evidence rather than test variance.
Phase 4 warm run `p95428` retried Postgres after a 15.718-second failed readiness
attempt, producing 19.046 seconds total readiness. Phase 9 warm run `p56449` retried
the object store after a 70.298-second authenticated-readiness timeout, producing
72.947 seconds readiness. The test schedulers in those runs remained at 25.579 and
24.650 seconds. Median and range retain these outliers rather than discarding them.

### Existing retained full warm runs

The three successful retained roots below predate the isolated cases and represent
full warm `check` evidence. Their `check-service-backed` scheduler, pressure, service
session, fixture, lifecycle, and cleanup artifacts were inspected.

| Run root | Scheduler | Service setup | Cleanup | Max running | Dominant blocker observations | Fixture observations |
| -------- | --------- | ------------- | ------- | ----------- | ----------------------------- | -------------------- |
| `.cartulary/test-results/20260712T013319Z-p24822` | 119.499s | 9.303s | 6.511s | 17 | `host_io` 44,475; clone 14,223; service session 9,570; CPU 8,106 | 319 actual clones, 46.380s summed clone events |
| `.cartulary/test-results/20260712T013623Z-p35104` | 115.898s | 6.935s | 4.086s | 17 | `host_io` 38,114; CPU 13,269; clone 12,884 | 318 actual clones, 39.689s summed clone events |
| `.cartulary/test-results/20260712T014318Z-p94370` | 123.860s | 8.154s | 4.271s | 17 | `host_io` 43,483; clone 12,219; service session 9,184; CPU 5,309 | 318 actual clones, 44.115s summed clone events |
| Median [range] | 119.499s [115.898--123.860] | 8.154s [6.935--9.303] | 4.271s [4.086--6.511] | 17 | I/O first in all runs | 318 clones; 44.115s [39.689--46.380] summed events |

These are developer-loop observations only. They do not satisfy Core 05 benchmark
acquisition or publication rules.

### Resource observations

The successful cases reached high aggregate CPU only during cold builds: 792--1000
percent on a 24-CPU host. Warm successful medians ranged from 305 to 504 percent for
service-backed cases and 347 percent for the selected frontend unit row. Available
memory never fell below 10.15 GiB; warm minima were 12.08--17.22 GiB. The sampled
swap value began around 1.2 GiB and changed by 0 MiB in every invocation. Host visible
RSS peaks ranged from about 4.7 to 15.9 GiB, but include unrelated WSL processes.
Memory is not the observed cap. Storage is virtual ext4 and retained full-check
pressure consistently identifies I/O as the first scheduler blocker.

## 4. Bottleneck analysis

| Bottleneck or stage | Type | Evidence and critical-path position | Can resources reduce it? |
| ------------------- | ---- | ----------------------------------- | ------------------------ |
| Serial pre-scheduler setup and CLI re-execution | Dependency/fixed overhead | `runPhaseSliceSetup` serializes every setup target before service startup. Successful warm cases retained about 10--11 seconds outside readiness, scheduler, and cleanup. It is before the reported scheduler critical path and therefore invisible to scheduler prioritization. | Not primarily. Parallelizing independent cached build validation and service startup is the structural fix. |
| Owned service readiness | Lifecycle/fixed with retry tail | Normal warm readiness was about 2.9--4.3 seconds, but two retries added 15 and 70 seconds. Postgres and object-store starts already overlap where allowed in `tools/testservices/main.go`. | More host memory does not address the observed refusal or authenticated-readiness timeout. Better attribution and retry diagnosis are needed; persistent sharing is not justified. |
| Browser target as one coarse phase unit | Runner-bound and fixed browser-stack overhead | Phase 4 and Phase 9 browser targets occupied essentially the complete warm scheduler critical path. Phase 9's median-run target spent 1.583s setup, 0.805s migration, 1.434s frontend startup, 13.966s Playwright body, and 0.256s teardown, yet the outer unit lasted 25.461s. | Bounded Playwright worker A/B tests may help the test body. Flattening groups and startup dependencies is needed to improve scheduler visibility. |
| Nested Playwright demand not represented by `process=1` | Resource-model mismatch | The phase unit claims one process and one browser stack, while `run-playwright-webserver-batch.sh` launches up to three concurrent one-worker Playwright processes. The outer scheduler cannot assign truthful worker ranges or account for nested demand. | A blanket increase risks oversubscription. Make each group or shard scheduler-visible with an explicit worker range and claims. |
| Clone and I/O pressure in broad/full schedules | Fixture/resource-bound | Full retained warm checks saturated `host_io=17` and `postgres_clone=8`; Phase 4 warm slices had 54 clones and clone was their top blocker. | Clone limit 4/6/8 is a valid controlled experiment. The hard max is 8; raising beyond it is neither supported nor a substitute for fixture proof. |
| Phase 12 scenario process fan-out | Granularity/future growth | `packShardLane` isolates every `scenario_id`. The measured plan had 76 backend-integration units for 76 rows, plus five integration-support units, one store unit, one process unit, and one browser unit: 84 total. Backend-integration finalization alone took about 4.7s. | Compatible-item consolidation can reduce command, clone-claim, and finalizer overhead, but current wall time is already short and browser-dominated. Require an A/B result before adoption. |
| Phase 12 claim/actual mismatch | Resource-model/fixture | Pressure classified 82 units as `template_clone` and recorded about 2,893 clone-blocker samples, while service scope observed only seven clone operations totaling under 0.8s in warm runs. | Do not merely raise clone capacity. Reconcile declared fixture policy, unit claims, and actual fixture boundaries first. |
| Frontend target-level expansion | Selection/runner-bound | One `FE-U-P8-01` row still invokes the frontend-unit target; it took about 9.85s with four Vitest workers. Cumulative FE-P8 selected 57 rows and 11 target units, including stateful, support, visual, and accessibility stages. Typecheck and similar semantic targets remain full-target checks. | Vitest 2/4/6 is a supported experiment. Exact row selection cannot make inherently target-wide checks row-local without changing their semantics. |
| Frontend prerequisite accounting leak | Correctness defect | Both browser-bearing frontend cases failed when `build-web` inherited selected-row scope. The failure happened after other work, taking 23 seconds for the exact row and more than four minutes for cumulative FE-P8. | Resources cannot fix it. Validate the plan/prerequisite scope earlier and keep non-evidence prerequisites outside selected-row closure. |
| Finalizers and aggregation | Fixed/runner-bound | Phase 12 finalizer durations summed about 7.4s, with backend integration about 4.7s. Phase 4 summed about 3.1--3.6s; Phase 9 about 1.9s. Finalizers correctly wait for their selected shards. | Consolidating compatible tiny shards can lower fan-in. Starting evidence collation earlier is safe only for immutable completed shards and must preserve final summaries. |
| Cleanup artifact anomaly | Lifecycle/accounting | Successful slice lifecycle streams first recorded an applied cleanup transition, then another `cleanup_started`/`cleanup_succeeded` pair marked `illegal` with `scheduler_accounting_error`. Public summaries still passed. | Not a speed lever. It should be fixed or explained before lifecycle artifacts are used as an implementation acceptance gate. |

Logical capacities resolved to `go_cpu=16`, `go_io=24`, `process=6`,
`postgres=32`, `object_store=32`, `browser_stack=1`, and `postgres_clone=8` in
the measured service-backed plans. These are scheduler tokens, not physical resources.
The resource owner is `tools/scheduler_resource_registry.json`, and the auto algorithms
are in `scheduler-resource-policy.mjs:180-226`. Current evidence does not justify
higher global capacities. Browser stack serialization is correctness-sensitive; clone
eight is already the specification-owned maximum; and physical CPU was not saturated
through most warm wall time.

Work that can overlap safely includes independent readiness-cache validation, cached
Go/frontend builds, and service startup when neither consumes the other's output.
Template creation must follow Postgres readiness and migrations. A test unit must wait
for its declared runtime binary, service session, and fixture. Browser groups that
share or taint a browser session, destructive reset boundary, database identity, or
observer state must remain serialized.

## 5. Optimization option matrix

Every option is assigned exactly one required classification: C1 supported
configuration, C2 private implementation, C3 owner-input/generated topology, C4 public
NLSpec behavior, or C5 rejected.

| Priority | Option | Bottleneck addressed | Expected wall-clock effect | Required changes | Specification impact | Correctness or isolation risk | Measurement needed |
| -------- | ------ | -------------------- | -------------------------- | ---------------- | -------------------- | ----------------------------- | ------------------ |
| 0 (C2) | Stop selected frontend row scope at non-evidence prerequisites and validate this boundary before expensive browser work. | Current FE-P8 harness failure and late error. | Makes the prescribed frontend browser paths executable; may avoid minutes of doomed work. No speed percentage is claimed. | Private execution/prerequisite scope and contract tests. | None; implements existing Sections 4, 5, 8, and 10 behavior. | Medium if scope is disabled on an evidence-bearing target; require positive and negative row closure tests. | Repeat both failing cases cold plus three warm; require pass and exact row accounting. |
| 1 (C4) | Generalize `ROWS` to a namespace-aware exact phase-row selector, including base phases. | Base slices are too broad for feature work. | Potentially large for single-row development, but depends on row and target. Do not invent a percentage. | Public input semantics, base planner, plan schema v2, accounting, diagnostics, tests, NLSpec. | Amend Sections 4.2, 4.3, TH-HARNESS-REQ-114, Section 8 schema attachment/semantics, Section 10/10.1 selection, and Section 17 acceptance. | Medium: partial selection could be mislabeled complete or omit shared prerequisites. | Compare exact row with full phase; prove identical assertions for the row and explicit partial status. |
| 2 (C2) | Replace serial setup/re-execution with normalized readiness and service-session DAG units. | About 10--11s warm fixed overhead and hidden dependencies. | Bounded by overlap among cached builds and 3--4s service readiness; numeric saving requires prototype data. | `phase-slice-execution` and scheduler service-session integration; retain current summaries and cleanup. | None if public inputs, artifacts, lifecycle, and failure mapping remain identical. | Medium: early work must not run without its binary/service or escape cleanup. | Timeline A/B on Phases 4, 9, 12; validate every dependency and cleanup artifact. |
| 3 (C3) | Flatten browser groups/shards into phase-scheduler-visible units with explicit worker ranges and truthful process/browser claims. | Coarse critical browser unit and nested demand mismatch. | Better concurrency and diagnostics; test-body savings depend on group balance. | Browser topology owner, browser unit builder, worker allocation, generated batch/scheduler outputs. | No public command change; stay within current scheduler algorithm and artifact contract. | High around shared session/reset/taint rules. | Playwright 2/3/4 A/B, non-overlapping ranges, reset/taint/leak tests. |
| 4 (C3) | Consolidate compatible tiny scenario shards, especially Phase 12, while retaining row identity and summaries. | 76 processes, 82 clone-class claims, and large finalizer fan-in. | Likely lowers process/finalizer overhead; current Phase 12 is not slow enough to assume a material win. | Phase-map/topology owner metadata, Go shard planner inputs or packing rules, duration baselines, generated plans. | None if selection and public artifacts remain stable. | High if rows share process globals, schema/database identity, destructive residue, controls, or observers. | One-factor A/B with identical row inventory and fixture proof; reject if median does not pass Section 9. |
| 5 (C3) | Lower eligible clone fixtures to transaction or group reuse only with accepted `cartulary.fixture_tier_proof.v1`. | Clone and I/O pressure. | Could reduce broad-run I/O; no numeric estimate before proof. | Fixture-policy owner inputs, proof artifacts, affected tests, generated pressure expectations. | Existing Section 11 proof model is sufficient unless a new policy token is introduced. | High contamination risk. | State-surface proof, repeated/random-order runs, leak/reset checks, clone event reduction. |
| Experiment (C1) | Compare `PLAYWRIGHT_WORKERS=2,3,4` on Phase 9. | Browser test body and nested worker saturation. | Unknown; bounded experiment only. | Caller configuration only. | None; current default/override is adopted. | Medium service/browser pressure at four. | Cold plus three warm per value, fixed caches and selected rows. |
| Experiment (C1) | Compare `VITEST_MAX_WORKERS=2,4,6` on `FE-U-P8-01`. | Selected frontend-unit duration. | Unknown; current warm median is 11.08s external. | Caller configuration only. | None. | Low; watch memory and flakes. | Three warm per value after one cold preparation. |
| Experiment (C1) | Compare clone limits 4, 6, and 8 on Phases 4 and 12. | Clone queueing and I/O saturation. | Unknown; may expose I/O as the next cap. | `CARTULARY_SERVICE_BACKED_POSTGRES_CLONE_LIMIT` only. | None within the current 1--8 bound at NLSpec line 1649. | Medium database/I/O pressure. | Same row inventory, fixture events, host memory, swap, queue samples, wall median. |
| Experiment (C1) | Test lower/equal Go CPU limits only after clone pressure is not the active cap. | Possible CPU scheduling balance. | Unknown. | `CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT` only. | None within registry bounds. | Low to medium; can lengthen queues. | One-factor sweep around resolved 16, after clone experiment. |
| Deferred (C1) | Increase WSL/Docker physical resources only after a run demonstrates a physical cap. | Potential future CPU or memory cap. | None expected for current warm slices. | Host configuration only. | None. | Higher I/O and thermal contention; reduced reproducibility. | Repeat baseline with declared host change and identical logical limits. |
| Rejected (C5) | Cache product-test success or reuse old retained runs as current execution. | Test body. | Would appear fast by skipping evidence. | Prohibited. | Violates Sections 4.1B, 6, 8, and 10. | Unacceptable evidence and correctness loss. | None. |
| Rejected (C5) | Share a persistent unproven service/browser stack across invocations. | Service startup. | Could save seconds but changes isolation and cleanup ownership. | Prohibited until a separately adopted lifecycle and contamination proof exists. | Would require public lifecycle work, not an implicit optimization. | Unacceptable today. | None until authority adopts a design. |
| Rejected (C5) | Remove database resets, browser resets, leak checks, or cleanup. | Fixture/teardown. | Unsafe apparent saving. | Prohibited. | Violates Sections 11--13. | Unacceptable contamination and leak risk. | None. |
| Rejected (C5) | Blanket worker, scheduler, clone, or process increases. | Assumed underutilization. | Unproven and may worsen I/O. | Prohibited as a recommendation. | May exceed owned bounds. | Oversubscription, flake, and service overload. | Use the bounded C1 experiments instead. |
| Rejected (C5) | Report a partial `ROWS` selection as full phase completion. | Cosmetic accounting. | No legitimate saving. | Prohibited. | Violates Sections 5, 8, and 10. | Unacceptable false conformance. | Negative acceptance fixture required. |

## 6. Recommended sequence

### Immediate configuration experiments

After the FE-P8 correctness defect is repaired, run the Playwright 2/3/4 and Vitest
2/4/6 sweeps. Run clone 4/6/8 on Phases 4 and 12 with all other controls fixed. Do
not change Go CPU until the clone sweep demonstrates that clone pressure is no longer
the active constraint. These are experiments, not default changes.

### Low-risk implementation changes

First repair frontend row-accounting propagation so build/readiness prerequisites are
not treated as row-closing targets. Fail plan validation before work if a selected
target or evidence prerequisite cannot consume the selector correctly.

Next model setup and the owned service session as scheduler-visible DAG units. Reuse
the existing check-scheduler service-session design rather than creating another
lifecycle. Allow independent cached build verification and service startup to overlap;
keep template migration behind Postgres readiness and keep every test behind all of
its declared prerequisites. Preserve output budgets, failure classification, timing,
service scope, finalizers, and cleanup.

### Fixture or sharding changes requiring validation

Flatten browser work at the same semantic reset/session boundaries already owned by
the browser topology. Allocate deterministic, non-overlapping Playwright worker ranges
and claim actual process/browser demand.

Treat Phase 12 consolidation as an experiment, not an assumed win. Compatible tiny
scenario items may share one Go command only when row summaries, scenario identity,
and fixture policy remain separately attributable. Never combine rows requiring
process-global isolation, schema or database-identity isolation, destructive residue,
independent runtime controls, or observer isolation. Reconcile why 82 units claim the
clone class while only seven clone operations are observed before altering capacity.

Fixture lowering follows only accepted `cartulary.fixture_tier_proof.v1` evidence.
Transaction rollback or group-clone reuse must enumerate the state surfaces restored;
clone capacity must not be raised to hide an invalid fixture tier.

### Public-contract and NLSpec changes

Adopt namespace-aware exact row selection in a single public contract revision:

* Omitted `ROWS` retains the existing full executable base phase and cumulative
  frontend-through-phase behavior.
* Supplied `ROWS` contains IDs owned by exactly the selected namespace and exact
  phase. Unknown, duplicate after normalization, cross-phase, cross-namespace,
  blocked, or non-executable IDs fail as usage exit 2 before setup.
* Base `service-backed-slice` additionally rejects an exact row that is not
  service-backed rather than silently dropping it.
* Selected-row artifacts enumerate the requested and resolved inventory. Missing,
  duplicate, or unmapped evidence fails closed.
* A partial selection emits an explicit partial selection status and cannot claim
  complete phase execution. Target-level checks may run a wider semantic target, but
  only selected evidence rows count toward closure.
* Existing `phase-slice` and `service-backed-slice` names and
  `cartulary.harness.command.*.v1` command IDs remain. No compatibility aliases are
  added.
* Introduce `cartulary.phase_slice_plan.v2`. Retain v1 only for old-run diagnostics;
  do not reinterpret v1's frontend-only namespace field under the same ID.

### Rejected approaches

Do not cache successful product tests, accept old run evidence as current, suppress
summaries, skip security or drift checks selected by a row, remove service readiness,
remove reset or cleanup, create a persistent default service pool, or raise all limits.
Do not preserve implicit base `ROWS` ignore behavior merely for compatibility.

## 7. Experiment plan

Use a fresh external cache root per experiment family, reuse it for three warm runs,
and keep source revision, Docker state, selected rows, and all non-principal factors
fixed. Each command remains Make-owned. Record the same artifacts and host samples as
Section 3.

### Frontend correctness gate

Hypothesis: separating selected evidence scope from non-evidence build prerequisites
allows both frontend commands to pass without broadening closure.

```sh
make phase-slice PHASE_NAMESPACE=frontend PHASE=FE-P8
make service-backed-slice PHASE_NAMESPACE=frontend PHASE=FE-P8 ROWS=FE-B-P8-01
```

Run one cold and three warm invocations each. Success requires all selected row IDs,
no unrelated row closure, valid browser/target/scheduler/service/cleanup artifacts,
and no `build-web` row-accounting failure. Roll back if the fix disables accounting
on any evidence-bearing target or lets an unmapped selected row pass.

### Playwright worker range

Hypothesis: Phase 9's 13--15 second Playwright body may improve at four workers, or
may prove that three is already the saturation point.

```sh
make service-backed-slice PHASE=phase9 PLAYWRIGHT_WORKERS=2
make service-backed-slice PHASE=phase9 PLAYWRIGHT_WORKERS=3
make service-backed-slice PHASE=phase9 PLAYWRIGHT_WORKERS=4
```

Compare external wall, browser work-unit wall, Playwright executed and wall duration,
stack startup, CPU, memory, service errors, worker ranges, and flaky outcomes. A value
is eligible only if it meets the global warm-median threshold, keeps at least 4 GiB
available, and produces no reset, leak, retry, or flake regression. Revert to three if
four merely moves the cap to service or I/O pressure.

### Vitest worker range

Hypothesis: six workers may reduce the approximately 9.15-second selected Vitest test
body without meaningful memory or process pressure.

```sh
make phase-slice PHASE_NAMESPACE=frontend PHASE=FE-P8 ROWS=FE-U-P8-01 VITEST_MAX_WORKERS=2
make phase-slice PHASE_NAMESPACE=frontend PHASE=FE-P8 ROWS=FE-U-P8-01 VITEST_MAX_WORKERS=4
make phase-slice PHASE_NAMESPACE=frontend PHASE=FE-P8 ROWS=FE-U-P8-01 VITEST_MAX_WORKERS=6
```

Compare external wall, frontend-unit timing, CPU, largest-child RSS, selected titles,
and retries/flakes. Reject any value that does not meet the global threshold, changes
the selected title inventory, or breaches resource gates.

### Postgres clone capacity

Hypothesis: Phase 4 may benefit from six or eight clone slots, while Phase 12's claim
pattern may show no benefit because actual clone work is already small.

```sh
CARTULARY_SERVICE_BACKED_POSTGRES_CLONE_LIMIT=4 make service-backed-slice PHASE=phase4
CARTULARY_SERVICE_BACKED_POSTGRES_CLONE_LIMIT=6 make service-backed-slice PHASE=phase4
CARTULARY_SERVICE_BACKED_POSTGRES_CLONE_LIMIT=8 make service-backed-slice PHASE=phase4
CARTULARY_SERVICE_BACKED_POSTGRES_CLONE_LIMIT=4 make service-backed-slice PHASE=phase12
CARTULARY_SERVICE_BACKED_POSTGRES_CLONE_LIMIT=6 make service-backed-slice PHASE=phase12
CARTULARY_SERVICE_BACKED_POSTGRES_CLONE_LIMIT=8 make service-backed-slice PHASE=phase12
```

Compare wall, clone event sum, blocker observations, Go I/O queueing, database errors,
available memory, swap, and fixture inventory. Roll back a candidate if I/O becomes
worse, fixture failures appear, or wall does not meet the global threshold. Never test
above the current specification-owned maximum of eight.

### Setup/service DAG prototype

Hypothesis: cached build verification and owned service startup can overlap enough to
remove at least 5 seconds from Phase 9 or 20 percent from a shorter service-backed
slice.

Use the unchanged public commands for Phases 4, 9, and 12 on the prototype branch.
Compare a scheduler-event timeline that includes setup/readiness units with the current
external residual. Require identical setup targets, cache decisions, selected rows,
service scope, failure classification, and cleanup. Roll back if any dependency is
weakened, timing disappears from summaries, or cleanup ownership changes.

### Browser flattening prototype

Hypothesis: scheduler-visible browser groups improve balance and make nested demand
truthful without adding stacks or reset risk.

Use `make service-backed-slice PHASE=phase9` with workers two, three, and four on the
prototype. Validate explicit contiguous non-overlapping worker ranges, one owned stack
per declared session group, reset/taint boundaries, and unchanged selected tests.
Reject the design if it creates another stack, shares state across an existing
isolation boundary, or fails the global threshold.

### Scenario consolidation prototype

Hypothesis: packing compatible Phase 12 scenario rows lowers process and finalizer
overhead while retaining row-level evidence.

Use `make service-backed-slice PHASE=phase12` with one owner-controlled packing factor
at a time. Compare process count, scheduler wall, backend-integration finalizer time,
fixture events, selected row IDs, per-row summaries, and test assertions. Run repeated
and randomized group-order contamination checks. Reject if any row loses independent
attribution or if the warm median fails the global threshold.

## 8. Minimal change map

No file in this map was changed during this investigation.

### Frontend prerequisite correctness, C2

Likely implementation and tests:

* `tools/harness/scheduler/phase-slice-execution.mjs` --
  `frontendRowAccountingEnv`, browser command construction, and future setup units.
* `tools/harness/phase-accounting/frontend/row-scope.mjs` -- keep strict evidence
  target mapping; do not weaken the `unmapped` guard globally.
* `tools/harness/browser/run-browser-e2e-target.sh` and browser lifecycle wrappers --
  establish the evidence-target boundary around prerequisites.
* `tools/harness/phase-accounting/tests/test-run-phase-slice.sh` and
  `tools/harness/browser/tests/test-run-playwright-webserver-batch.sh` -- positive
  FE-P8 browser selection and negative unmapped-row fixtures.
* If owner topology must mark evidence-consuming prerequisites,
  `tools/execution_topology_manifest.json` is the owner input; regenerate, do not edit,
  `tools/task_surface_manifest.json`, `tools/task_surface.generated.mk`,
  `tools/scheduler_manifest.json`, `tools/browser_e2e_batch_manifest.json`, or
  `tools/execution_topology_render_index.json`.

### Namespace-aware exact `ROWS`, C4

Specification and schema:

* `docs/testing-harness-nlspec.md` -- Section 4.2 phase-slice family; Section 4.3
  `phase-slice` and `service-backed-slice` rows; TH-HARNESS-REQ-114 input matrix at
  lines 844--848; Section 8 schema table and plan semantics at lines 1123 and 1160;
  Section 10 and 10.1 namespace/selection semantics beginning at line 1497; Section
  17 positive and negative acceptance coverage.
* New `tools/schemas/cartulary.phase_slice_plan.v2.schema.json`; retain v1 for old
  diagnostics only.

Implementation and tests:

* `tools/harness/command-surface/make-node-tools.mjs` -- usage and public input help.
* `tools/harness/phase-accounting/phase-slice-cli.mjs` -- pass normalized base rows
  to the planner and map selector errors to usage exit 2.
* `tools/harness/phase-accounting/phase-slice-plan.mjs` and
  `phase-slice-planning/row-selection.mjs` -- exact namespace/phase filtering and
  partial selection status.
* `tools/harness/phase-accounting/frontend-phase-slice-plan.mjs` and
  `frontend/row-scope.mjs` -- exact-phase explicit selection while retaining omitted
  cumulative selection.
* `tools/harness/phase-accounting/phase-slice-plan-contract.mjs` -- v2 semantics.
* `tools/harness/generated-artifacts/tests/test-json-shapes.sh`,
  `tools/harness/phase-accounting/tests/test-run-phase-slice.sh`, and
  `tools/harness/tests/test-harness-contracts.mjs` -- schema, selector, accounting,
  invalid-input, and no-false-completion tests.
* The generated Make/task surface changes only if its owner registry/input changes;
  retain both public target names and command IDs.

### Setup and service-session DAG, C2

* `tools/harness/scheduler/phase-slice-execution.mjs` -- replace
  `runPhaseSliceSetup` and re-execution with normalized units and dependencies.
* `tools/harness/scheduler/check-schedule-manifest.mjs`,
  `scheduler/engine.mjs`, and service-session helpers -- reuse current owned-session
  semantics where a private generalization is needed.
* `tools/testservices/main.go` only if the existing start/lease/terminate interface
  cannot be reused unchanged.
* `tools/harness/phase-accounting/tests/test-run-phase-slice.sh`,
  `tools/harness/scheduler/tests/test-check-scheduler.sh`, lifecycle tests, and timing
  drift fixtures -- dependency, failure, retry, finalizer, and cleanup coverage.
* Existing summary schemas should not change unless the prototype proves a required
  public field. Private work-unit fields remain in the v1/v2 extension zone.

### Browser topology flattening, C3

* `tools/execution_topology_manifest.json` and
  `tools/browser_e2e_duration_baselines.json` -- owner grouping and weights.
* `tools/harness/phase-accounting/phase-slice-planning/browser-work-units.mjs` --
  emit visible group units and truthful claims.
* `tools/harness/browser/run-playwright-webserver-batch.sh`,
  `browser-shard-plan.mjs`, and browser scheduler-dependency helpers -- accept
  scheduler-owned group and worker range.
* `tools/scheduler_resource_registry.json` only if an existing resource cannot model
  the proven contention; do not add tokens preemptively.
* Browser batch, scheduler, phase-slice, reset, taint, and worker-range tests.
* Regenerate the task surface, scheduler manifest, browser batch manifest, topology
  render index, phase schedules, and ledgers only through their owner generators.

### Go scenario consolidation and fixture proof, C3

* Affected `tools/phase*_test_map.json` owner rows, especially
  `tools/phase12_test_map.json`, if compatibility/isolation metadata is row-owned.
* `tools/harness/backend/go-shard-plan.mjs` -- packing behavior only after owner
  metadata can express compatibility; preserve `shard_isolation` fail-closed behavior.
* `tools/go_test_duration_baselines.json` and affected duration owner inputs after a
  successful measured topology, not before.
* Fixture-policy owner rows and affected Go tests only with accepted
  `cartulary.fixture_tier_proof.v1` artifacts; schema already exists at
  `tools/schemas/cartulary.fixture_tier_proof.v1.schema.json`.
* Backend shard-plan, aggregation, fixture-budget, contamination, and phase-slice
  tests; generated schedules/ledgers only via generators.

## 9. Acceptance criteria for the next implementation pass

An optimization passes only if every applicable criterion is true:

1. On the same declared host and cache protocol, three successful warm candidate runs
   reduce median external wall time by at least 20 percent or at least 5.0 seconds
   against three successful warm baseline runs. Failed, retry-contaminated, cold, or
   old-run samples cannot establish the performance win, though all remain reported.
2. The requested and resolved selected-row inventory is byte-for-byte identical to
   baseline for a private/topology change. For the new public exact-row feature, it is
   identical to the approved selector fixture. Partial execution is explicitly marked
   and never claims complete phase execution.
3. Product assertion titles/symbols and counts are identical. There are no new
   unmapped, missing, duplicated, unexpectedly broad, or silently dropped evidence
   rows.
4. Scheduler summary/events, pressure summary, target and phase summaries,
   tool-run summary, fixture artifacts, service scope/lease/lifecycle, child logs,
   timing, finalizers, and cleanup artifacts all validate and agree on status.
5. Frontend selected browser cases pass, and a negative unmapped frontend row still
   fails before expensive child work. Non-evidence prerequisites do not claim rows;
   evidence targets cannot evade selected-row accounting.
6. Product versus harness failure classification and public exit codes are unchanged.
   Retry and cleanup failures remain visible and cannot be overwritten by a speed
   path.
7. Fixture policy, reset, taint, database identity, browser session, object-store
   cleanup, and observer boundaries are unchanged unless an accepted fixture-tier or
   topology proof explicitly authorizes the change. Repeated and randomized-order
   contamination tests pass.
8. No managed container, process, port, database, bucket, lease, runtime, or result
   artifact leaks. Lifecycle streams contain no unexplained illegal transition; the
   duplicate illegal cleanup events observed in this report are resolved or an adopted
   owner explanation and test establishes why they are valid.
9. There is no increase in flaky outcomes over repeated baseline/candidate runs and no
   additional service readiness retry. A single unexplained new flake rejects the
   candidate until diagnosed.
10. Largest-child peak RSS and sampled host RSS are no more than 25 percent above
    baseline, at least 4 GiB host memory remains available, and swap grows by no more
    than 512 MiB during an invocation. CPU and I/O saturation must be reported rather
    than hidden by a median.
11. Generated outputs change only after their owner inputs and only through Make-owned
    generators. Generated-artifact policy, JSON shape, phase ledger/schedule drift,
    and affected duration-baseline drift checks pass.
12. The narrow affected harness tests pass first, followed by the required harness
    smoke and drift checks selected by `make task-guide`. No baseline is refreshed to
    make a regression disappear.

## 10. Assumptions, limits, and open decisions

### Assumptions and limits

* Measurements are local developer-experience evidence from one WSL2/Docker Desktop
  host. They are not portable performance claims and are not Core 05 evidence.
* Cold means fresh external readiness/build/Go build-cache directories, not a deleted
  module cache, Docker image store, pnpm install, browser installation, source output,
  or operating-system page cache. This avoids mutating existing caches and mirrors a
  safe isolated first invocation.
* Owned services are cold on every invocation under the current lifecycle. Warm refers
  to allowed tool/readiness/build caches, not service persistence.
* Peak whole-tree RSS and peak CPU were unavailable. GNU time reports largest-child
  RSS and average aggregate CPU. The host sampler's RSS includes unrelated WSL
  processes. Scheduler queue wait is available only as blocker observation counts,
  not milliseconds.
* The Phase 4 and Phase 9 warm ranges include one successful readiness retry each.
  Medians were not selectively cleaned. Future performance acceptance excludes
  retry-contaminated samples from the performance trio but must separately prove no
  retry regression.
* Failed frontend cases do not establish successful performance baselines. Their
  timings show failure cost and scope expansion only.
* Plan construction/validation, Make process startup, and individual serial setup
  target wall time do not have dedicated slice-level timing buckets. They are included
  in the external residual and can be separated only after the proposed DAG makes them
  scheduler-visible.
* Phase 12's 76 backend-integration processes were observed in retained scheduler
  units and follow directly from the scenario isolation branch. Consolidation still
  requires proof because process count alone did not control current wall time.

### Documentation and implementation drift

1. The adopted input table at `docs/testing-harness-nlspec.md:845` defines `ROWS` as
   frontend row IDs. The live Make surface accepts `ROWS` for either namespace, but
   the base CLI branch ignores it. The next contract should fail invalid base input
   until namespace-aware base selection is adopted; silent ignore is unsafe.
2. Sections 4 and 5 require executable frontend slices to enforce selected implemented
   rows. The current FE-P8 browser paths instead pass selected scope into `build-web`,
   a non-evidence prerequisite, and fail after substantial work. This is live drift
   and should be corrected before optimization measurement.
3. Successful slice lifecycle streams contained a second cleanup pair marked
   `transition_status="illegal"` and `failure_reason="scheduler_accounting_error"`
   after an applied cleanup pair. Public summaries still passed. Section 11.2 and
   TH-HARNESS-AC-017 require legal terminal transitions; maintainers must decide
   whether the reaper emission is a bug or the lifecycle owner needs a different
   non-transition event.
4. `make explain-target TARGET=phase-slice` describes the target without a concrete
   phase and therefore cannot expose conditional service requirements. This is a
   diagnostic limitation, not evidence that slice plans have no services; concrete
   plans and service scopes are authoritative for the invocation.

### Open maintainer decisions

* Approve or reject the C4 exact-phase selector semantics, especially the deliberate
  removal of earlier-phase IDs from explicit frontend `ROWS` while retaining omitted
  cumulative behavior.
* Decide the artifact field that distinguishes partial selected-row success from full
  phase completion before approving `cartulary.phase_slice_plan.v2`.
* Approve the minimal compatibility metadata needed to group scenario rows. Prefer a
  fail-closed isolation declaration over filename, package, or scenario-name inference.
* Determine whether setup/service DAG normalization can reuse the check scheduler's
  private service-session representation unchanged. Do not create a second lifecycle
  merely to avoid a small generalization.
* Resolve the illegal duplicate cleanup events before using lifecycle validity as a
  performance acceptance signal.

### Investigation verification and repository state

The report is the only intended tracked change. Validation for this documentation-only
pass completed as follows:

* `make lint-markdown` -- pass.
* `make generated-artifact-policy-check` -- pass; retained root
  `.cartulary/test-results/20260712T025820Z-p90175`.
* `make json-shape-check` -- pass; retained root
  `.cartulary/test-results/20260712T025820Z-p90194`.

Broad `make check`, `make test`, release suites, optimization experiments, baseline
refreshes, specification edits, owner-input generation, and commits are intentionally
out of scope. `make agent-finalize` retained-run maintenance is not performed because
there is no implementation or generated owner-input change and no baseline refresh is
authorized. Authority and terminology inputs inspected included
`docs/testing-harness-nlspec.md`, `docs/domain.md`, `temp/planner-prompt.md`, the
implementation/testing guides, live phase registries/maps, topology/resource owner
inputs, generated manifests, runners, service code, and retained artifacts cited above.

## 11. Work tracker

This tracker uses the modular refactor planning framework as a formatting and handoff
style guide only. The authority, classifications, sequencing, and acceptance criteria
remain those in this report and the Testing Harness NLSpec. Update the tracker at every
implementation checkpoint.

Status values are `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`, and `DROPPED`.
`BLOCKED` means an owner decision or prerequisite is required; `DEFERRED` means the
work is intentionally held until its evidence gate is satisfied.

| ID | Work item | Workstream | Class | Status | Depends on | Owner | Evidence or artifact | Exit condition |
| -- | --------- | ---------- | ----- | ------ | ---------- | ----- | -------------------- | -------------- |
| DX-001 | Trace base and frontend slice execution end to end. | discovery | investigation | DONE | none | report author | Sections 2 and 4; cited paths and symbols | Public entry, selection, setup, services, runners, summaries, and cleanup are mapped. |
| DX-002 | Establish cold and warm developer-loop baselines. | measurement | investigation | DONE | DX-001 | report author | Section 3; retained roots ending `p24822`, `p35104`, `p94370`, and the seven new case families | Required samples, medians, ranges, host limits, failures, and measurement limitations are recorded. |
| DX-003 | Repair selected frontend row scope at non-evidence prerequisites. | correctness | C2 | TODO | DX-001,DX-002 | harness maintainer | FE-P8 failures ending `p94470`, `p36308`, `p67854`, `p99288`, `p30971`, `p51478`, `p61947`, `p72161` | Both prescribed frontend browser commands pass cold and three warm runs; exact row accounting remains strict and negative unmapped-row tests still fail before expensive work. |
| DX-004 | Resolve or owner-explain duplicate illegal cleanup events. | lifecycle | C2 | BLOCKED | DX-002 | service-lifecycle owner | Successful slice `lifecycle-events.jsonl`; Section 10 drift item 3 | Successful lifecycle streams contain only adopted legal transitions, or an owner-approved non-transition event and acceptance test replaces the illegal pair. |
| DX-005 | Approve namespace-aware exact-phase `ROWS` semantics and partial-completion representation. | contract | C4 | BLOCKED | DX-001,DX-002 | Testing Harness NLSpec owner | Sections 5, 6, and 10 open decisions | Owner approves exact namespace/phase behavior, invalid-input exit 2, omitted-input behavior, and the artifact field that prevents partial selection from claiming complete phase execution. |
| DX-006 | Amend the NLSpec and introduce `cartulary.phase_slice_plan.v2`. | contract | C4 | TODO | DX-005 | Testing Harness NLSpec owner | Section 8 change map | Sections 4.2, 4.3, TH-HARNESS-REQ-114, 8, 10/10.1, and 17 are revised; v2 validates; v1 remains diagnostic-only for old runs. |
| DX-007 | Implement namespace-aware base/frontend exact row selection. | selection | C4 | TODO | DX-006 | harness maintainer | Selector tests and retained v2 plans | Omitted selection preserves current full/cumulative behavior; valid exact rows execute; unknown, cross-phase, cross-namespace, blocked, duplicate, unmapped, and non-service-backed IDs fail closed as specified. |
| DX-008 | Prototype scheduler-visible setup and owned service-session DAG units. | scheduler | C2 | TODO | DX-003,DX-004 | scheduler and service-lifecycle maintainers | Phase 4, 9, and 12 baseline timelines | Independent cached builds and service startup overlap; every dependency, summary, failure class, fixture boundary, and cleanup guarantee remains intact. |
| DX-009 | Run the Vitest worker experiment at 2, 4, and 6 workers. | configuration | C1 | TODO | DX-002 | harness investigator | `FE-U-P8-01` cold plus three warm runs per value | A candidate meets Section 9's wall-time and resource gates with identical selected titles and no flake increase; otherwise retain four. |
| DX-010 | Run the Playwright worker experiment at 2, 3, and 4 workers. | configuration | C1 | TODO | DX-002 | harness investigator | Phase 9 cold plus three warm runs per value | A candidate meets Section 9, uses valid worker ranges, and adds no service, reset, leak, memory, or flake regression; otherwise retain three. |
| DX-011 | Run the Postgres clone-capacity experiment at 4, 6, and 8. | configuration | C1 | TODO | DX-002 | harness investigator | Phase 4 and Phase 12 fixture, pressure, host, and timing artifacts | A candidate meets Section 9 without increased I/O, database errors, contamination, or resource pressure; no value above eight is tested. |
| DX-012 | Test Go CPU limits after clone pressure is removed as the active cap. | configuration | C1 | DEFERRED | DX-011 | harness investigator | Clone experiment conclusion and follow-up pressure summaries | Clone is no longer the active blocker; one-factor Go CPU results identify a faster safe value or retain resolved 16. |
| DX-013 | Flatten browser groups into scheduler-visible units with truthful claims and worker ranges. | topology | C3 | TODO | DX-003,DX-010 | browser topology owner | Phase 9 browser plans, worker-range artifacts, reset/taint tests | Selected browser groups have deterministic non-overlapping ranges and truthful claims; current session/reset/taint boundaries and all artifacts remain valid; Section 9 passes. |
| DX-014 | Reconcile Phase 12 declared clone pressure with observed fixture operations. | topology and fixtures | C3 | TODO | DX-002 | Go topology and fixture owners | Phase 12 pressure says 82 template-clone units; service scope observes seven clone operations | Each unit's fixture class and resource claim matches its real execution boundary, or an owner-backed reason and test explains the retained conservative claim. |
| DX-015 | A/B-test consolidation of compatible Phase 12 scenario shards. | sharding | C3 | DEFERRED | DX-014 | Go topology owner | Row compatibility metadata, shard plans, finalizer timing, contamination tests | Process/finalizer overhead falls enough to pass Section 9 while all 76 backend-integration rows retain assertions, scenario identity, summaries, and required isolation. |
| DX-016 | Prove and apply any fixture-tier lowering. | fixtures | C3 | DEFERRED | DX-014 | fixture-policy owner | Accepted `cartulary.fixture_tier_proof.v1` per affected row or group | Proof enumerates restored state surfaces; randomized/repeated contamination checks pass; clone work falls; Section 9 passes. |
| DX-017 | Run implementation acceptance and generated-artifact verification. | validation | implementation gate | TODO | DX-007 and any implemented C2/C3 item | implementing maintainer | Section 9 artifacts and Make-owned validation results | Every applicable binary criterion in Section 9 passes; generated outputs come only from owner inputs and generators; failures and skipped checks are recorded. |
| DX-018 | Update this tracker and write the implementation handoff. | handoff | process | TODO | DX-017 | implementing maintainer | Changed-file list, commands, retained roots, decisions, blockers, and next action | Another maintainer can resume without rediscovery; statuses and dependencies reflect the repository state. |

The next executable item is DX-003. DX-004 and DX-005 require named owner decisions
before their dependent implementation work can complete. DX-009, DX-010, and DX-011
may run in parallel with those decisions because their selected baseline commands
already pass. Only mark DX-015 or DX-016 active after their proof prerequisites are
complete.
