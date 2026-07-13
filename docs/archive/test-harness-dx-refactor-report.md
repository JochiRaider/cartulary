# Slice-harness developer-loop investigation

Status: investigation and implementation planning only  
Measured: 2026-07-11 through 2026-07-12, America/New_York  
Evidence class: developer-experience engineering evidence, not Core 05 benchmark or
claim-publication evidence
Remediation status: DX-008 through DX-020 completed or explicitly dropped on evidence;
Sections 11 and 13 are the current execution and handoff record

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
| DX-003 | Repair selected frontend row scope at non-evidence prerequisites. | correctness | C2 | DONE | DX-001,DX-002 | harness maintainer | Exact-row pass roots ending `p43364`, `p59043`, `p75445`, `p96659`; cumulative pass roots ending `p71001`, `p36962`, `p69436`, `p2172`; prerequisite and support-selection tests | Both prescribed frontend browser commands pass cold and three warm runs; selected support and functional scenarios close, non-evidence `build-web` emits no row accounting, and strict target binding remains active. |
| DX-004 | Resolve or owner-explain duplicate illegal cleanup events. | lifecycle | C2 | DONE | DX-002 | service-lifecycle owner | `make service-backed-unit-check`; `make run-harness-smoke-lifecycle`; eight DX-003 lifecycle streams ending `p43364`, `p59043`, `p75445`, `p96659`, `p71001`, `p36962`, `p69436`, `p2172` | Detached reaper scheduling is a non-transition service event; the terminating path owns the single cleanup terminal event; all retained acceptance streams end in one legal applied cleanup pair with no illegal transition. |
| DX-005 | Approve namespace-aware exact-phase `ROWS` semantics and partial-completion representation. | contract | C4 | DONE | DX-001,DX-002 | Testing Harness NLSpec owner | Owner approval received 2026-07-11; approved v2 `selection` object records mode, phase span, dependency scope, completion scope, and requested/resolved row IDs | Exact namespace/phase behavior, invalid-input exit 2, omitted-input behavior, and `completion_scope=selected_subset` for partial execution are approved. |
| DX-006 | Amend the NLSpec and introduce `cartulary.phase_slice_plan.v2`. | contract | C4 | DONE | DX-005 | Testing Harness NLSpec owner | `make json-shape-check` root ending `p45476`; `make harness-contract`; `make generated-artifact-policy-check` root ending `p59564`; `make phase-schedule-drift` root ending `p59745`; `make lint-markdown` | Sections 4.2, 4.3, TH-HARNESS-REQ-114, 8, 10/10.1, and 17 define v2 selection and partial-completion semantics; v2 validates before output and retention; v1 is historical-diagnostic only. |
| DX-007 | Implement namespace-aware base/frontend exact row selection. | selection | C4 | DONE | DX-006 | harness maintainer | Selector smoke root ending `p7941`; exact base unit `p79601`; exact base service-backed `p78643`; exact base Vitest `p96427`; exact base browser `p97031`; exact frontend unit `p79868`; exact frontend browser/support `p81242`; invalid duplicate root `p7560` | Omitted selection preserves full/cumulative behavior; exact rows propagate into Go, Vitest, Playwright, scheduler, finalizer, and accounting inventories; malformed or ineligible selectors fail before setup with config/usage_error and exit 2; retained and JSON v2 plans share one serializer. |
| DX-008 | Prototype scheduler-visible setup and owned service-session DAG units. | scheduler | C2 | DONE | DX-003,DX-004 | scheduler and service-lifecycle maintainers | Cold/warm Phase 4 roots ending `p7659`, `p22732`, `p42487`, `p61914`; Phase 9 `p81373`, `p96325`, `p13866`, `p30903`; Phase 12 `p47914`, `p36580`, `p58867`, `p81111`; `make run-harness-smoke-execution`; `make run-harness-smoke-extended` | Shared scheduler-owned service sessions replace serial setup/re-execution. Warm medians are 31.388s, 31.713s, and 19.346s for Phases 4, 9, and 12, each more than 5s faster than Section 3; row inventories are stable, cleanup passes, and lifecycle streams contain no illegal transition. |
| DX-009 | Run the Vitest worker experiment at 2, 4, and 6 workers. | configuration | C1 | DONE | DX-002 | harness investigator | Worker 2 roots ending `p5557`, `p7326`, `p8716`, `p10126`; worker 4 `p11515`, `p13163`, `p14500`, `p15842`; worker 6 `p17179`, `p18793`, `p20116`, `p21435` | Warm medians were 16.084s, 10.533s, and 8.520s. Six workers improved 19.1% and 2.013s over four, below both Section 9 thresholds, so the public default remains four. All runs closed only `FE-U-P8-01` with no missing row. The slice worker inputs are now owner-declared instead of forwarded-but-rejected drift. |
| DX-010 | Run the Playwright worker experiment at 2, 3, and 4 workers. | configuration | C1 | DONE | DX-002 | harness investigator | Worker 2 roots ending `p22981`, `p39023`, `p56274`, `p73279`; worker 3 `p90764`, `p6442`, `p24030`, `p41001`; worker 4 `p58075`, `p72730`, `p89944`, `p7377` | Warm medians were 36.986s, 31.360s, and 28.410s. Four workers improved 9.4% and 2.950s over three, below both Section 9 thresholds, so the public default remains three. All runs retained the same 25 rows, passed cleanup, and emitted no illegal lifecycle transition. |
| DX-011 | Run the Postgres clone-capacity experiment at 4, 6, and 8. | configuration | C1 | DONE | DX-002 | harness investigator | Phase 4 clone 4 roots ending `p24862`, `p32709`, `p52625`, `p71938`; clone 6 `p91366`, `p4606`, `p24220`, `p43510`; clone 8 `p62936`, `p78444`, `p98298`, `p17954`; Phase 12 clone 4 `p37247`, `p7412`, `p30120`, `p52340`; clone 6 `p74594`, `p54830`, `p77093`, `p99248`; clone 8 `p21926`, `p11188`, `p33568`, `p55811` | Phase 4 warm medians were 30.462s, 32.184s, and 32.522s; four improved only 6.3% and 2.060s over eight. Phase 12 medians were 27.883s, 21.689s, and 19.964s and favored eight. No single lower value passed Section 9 across both phases, so the auto-policy remains unchanged. All runs passed with stable row inventories and cleanup. |
| DX-012 | Test Go CPU limits after clone pressure is removed as the active cap. | configuration | C1 | DONE | DX-011 | harness investigator | CPU 8 Phase 9 roots ending `p81451`, `p98633`, `p16107`, `p33146`, Phase 12 `p50370`, `p66578`, `p85034`, `p98330`; CPU 12 Phase 9 `p12131`, `p29226`, `p46279`, `p63405`, Phase 12 `p80602`, `p93855`, `p7467`, `p20676`; CPU 16 Phase 9 `p34030`, `p51051`, `p68200`, `p85366`, Phase 12 `p2800`, `p16072`, `p29503`, `p42728` | Phase 9 warm medians were 32.836s, 32.128s, and 32.462s; Phase 12 medians were 17.463s, 16.812s, and 17.463s. Twelve was numerically best but cleared neither Section 9 threshold in either phase, so the current auto resolver remains unchanged. Row/test inventories were stable with no missing evidence. |
| DX-013 | Flatten browser groups into scheduler-visible units with truthful claims and worker ranges. | topology | C3 | DROPPED | DX-003,DX-010 | browser topology owner | Prototype cold/root ending `p12641`; warm roots ending `p37353`, `p57343`, `p75328`; scheduler summary in `p12641` proves one session, four selected groups, 40/40 units, and cleanup pass | The prototype preserved 25 rows, 58 tests, deterministic three-lane ranges, one stack, and legal cleanup, but its 38.531s warm median regressed 22.9% from the accepted DX-010 31.360s median. The prototype was removed and the coarse browser target retained; no private compatibility path remains. |
| DX-014 | Reconcile Phase 12 declared clone pressure with observed fixture operations. | topology and fixtures | C3 | DONE | DX-002 | Go topology and fixture owners | Exact static row root ending `p96327`; full 112-row root ending `p96806`; service-backed 15-row root ending `p23615`; pressure summary in `p23615` | The 76 repository/in-memory Network Flow selectors are backend-unit rows with no fixture policy; their file is renamed `phase12_network_flow_contract_test.go`. Service-backed Phase 12 now excludes them and reports six clone-claiming work units, all owned harness-control/process support. The full row inventory remains 112 with no missing or unmapped evidence. |
| DX-015 | A/B-test consolidation of compatible Phase 12 scenario shards. | sharding | C3 | DROPPED | DX-014 | Go topology owner | Isolated control roots ending `p78191`, `p4109`, `p29494`, `p55032`; size 2 `p41378`, `p64355`, `p91861`, `p13638`; size 4 `p34407`, `p53437`, `p71771`, `p90116`; size 8 `p8863`, `p26335`, `p43472`, `p60765` | Warm medians were 22.056s isolated, 20.700s at 2, 19.295s at 4, and 20.238s at 8. Size 4 improved only 12.5% and 2.761s, below both Section 9 thresholds. Every run retained 112 rows and 134 tests with no missing evidence. The optional owner field, planner branch, and candidate metadata were removed rather than carrying an unproven compatibility surface. |
| DX-016 | Prove and apply any fixture-tier lowering. | fixtures | C3 | DROPPED | DX-014 | fixture-policy owner | Phase 12 root ending `p23615`, including `fixture_tier_proofs` for every remaining harness-control symbol; six clone-claiming work units and no former Network Flow contract selectors | No remaining candidate earned a lowering proof. The accepted retained proofs enumerate Postgres and process lifecycle, require clone isolation for committed runtime-fixture reset behavior, and report no unproven dirty-table reset surface. Template-clone policy is retained for those real users; no new fixture token or unsafe lowering was added. |
| DX-017 | Run implementation acceptance and generated-artifact verification. | validation | implementation gate | DONE | DX-007 and any implemented C2/C3 item | implementing maintainer | `make check` pass root ending `p47908`; execution, lifecycle, and extended smoke passes; final v2 cumulative roots ending `p58607`, `p91088`, `p23781`, `p56399`; exact roots ending `p88864`, `p1662`, `p14078`, `p26478` | Every applicable binary criterion passes; 335 check units and 1,074 tests pass; final browser inventories agree; all eight lifecycles contain one legal cleanup pair and no illegal transition; no baseline was refreshed. |
| DX-018 | Update this tracker and write the implementation handoff. | handoff | process | DONE | DX-017 | implementing maintainer | Section 12 completion record; final Markdown, schema, generated-policy, and drift checks | Current behavior, changed files, compatibility, validation roots, retries, skipped work, and the next optional item are recorded without rewriting the historical investigation evidence. |
| DX-019 | Run final remediation acceptance and generated-artifact verification. | validation | implementation gate | DONE | DX-008,DX-012,DX-013,DX-014,DX-015,DX-016 | implementing maintainer | Final Phase 4 roots ending `p3858` and `p23634`; Phase 9 `p43001` and `p60862`; Phase 12 `p78039` and `p580`; browser roots ending `p13909`, `p31004`, `p49374`, and `p62451`; full `make check` and `make release-check` root ending `p1334` | Every applicable Section 9 criterion passes; all owner-derived artifacts are current; execution, lifecycle, frontend, browser, Phase 4/9/12, check, and release gates pass. The final release root contains 14 of 14 work units, 1,131 tests, and no failed or missing evidence. |
| DX-020 | Update the tracker and write the final remediation handoff. | handoff | process | DONE | DX-019 | implementing maintainer | Section 13 final remediation completion record | Final decisions, compatibility, migrations, measurements, failures, retained roots, validation, and remaining risks are recorded; no item in this effort remains active or deferred. |

DX-008 through DX-012, DX-014, DX-019, and DX-020 are complete. DX-013, DX-015,
and DX-016 were measured or proved and dropped because their candidates did not clear
the acceptance gate. No workstream in this remediation pass remains active or
deferred.

## 12. Implementation completion record

### 12.1 Adopted decisions and behavior

- Current phase-slice producers emit only `cartulary.phase_slice_plan.v2`. V1
  remains readable only for historical diagnostics and has no dual-emission or
  compatibility alias.
- V2 requires `phase_namespace` and the closed `selection` object. Explicit rows
  are sorted, duplicate-checked, exact-namespace, and exact-phase. Every exact
  selection and every service-backed slice reports `selected_subset`; a complete
  row-claim rollup does not represent complete phase execution for that scope.
- Base and frontend selectors share one parser and one canonical plan serializer.
  Retained plans are validated and written before setup, no-op completion, or
  child execution. JSON and retained output are byte-equivalent after parsing.
- Exact base rows propagate into Go shard discovery, Go finalization and manifest
  verification, Vitest title/file selection, and Playwright shard selection. Full
  target emissions do not inherit exact-row filtering.
- Selected frontend accounting is privately bound to the intended evidence
  target. Non-evidence prerequisites such as `build-web` cannot consume or emit
  selected-row accounting; mismatched evidence targets fail closed.
- Browser-backed `FE-B-P8-01` closes both its functional and support evidence
  without broad or duplicate row claims. Empty support selections remain empty
  rather than becoming a blank synthetic phase.
- The service cleanup owner emits one legal cleanup start and one terminal cleanup
  event. Detached scheduling delegates terminal ownership to the reaper; fallback
  uses the same coordinator. Startup failure remains terminal and resource proof
  does not mutate it through cleanup lifecycle events.
- Unknown, blank, normalized-duplicate, cross-phase, cross-namespace, blocked,
  unmapped, inactive, otherwise non-executable, and invalid service-backed row
  selections fail before setup as `config`/`usage_error`, public exit `2`.

Public Make targets and `cartulary.harness.command.*.v1` IDs are unchanged. The
intentional compatibility breaks are that base `ROWS` is no longer ignored and an
explicit frontend selector no longer accepts an earlier-phase row. Omitted base
and frontend selection retains exact-base and cumulative-frontend behavior,
respectively.

### 12.2 Changed files

Specification and handoff:

- `docs/testing-harness-nlspec.md`
- `docs/handoffs/test-harness-dx-refactor-report.md`

Contract, topology owner inputs, and generated projections:

- `tools/schemas/cartulary.phase_slice_plan.v2.schema.json`
- `tools/execution_topology_manifest.json`
- `tools/execution_topology_render_index.json`
- `tools/task_surface_manifest.json`
- `tools/harness/generated-artifacts/task-surface.mjs`
- `tools/harness/command-surface/make-node-tools.mjs`

Planner, selector, runtime, and accounting implementation:

- `tools/harness/phase-accounting/phase-row-selector.mjs`
- `tools/harness/phase-accounting/phase-slice-plan-contract.mjs`
- `tools/harness/phase-accounting/phase-slice-plan.mjs`
- `tools/harness/phase-accounting/frontend-phase-slice-plan.mjs`
- `tools/harness/phase-accounting/phase-slice-cli.mjs`
- `tools/harness/phase-accounting/index.mjs`
- `tools/harness/phase-accounting/phase-selection.mjs`
- `tools/harness/phase-accounting/phase-manifest-cli.mjs`
- `tools/harness/phase-accounting/frontend-row-accounting.mjs`
- `tools/harness/phase-accounting/frontend/row-scope.mjs`
- `tools/harness/phase-accounting/phase-slice-planning/row-selection.mjs`
- `tools/harness/phase-accounting/phase-slice-planning/backend-work-units.mjs`
- `tools/harness/scheduler/phase-slice-execution.mjs`
- `tools/harness/backend/target-execution/context.mjs`
- `tools/harness/backend/target-execution/planning.mjs`
- `tools/harness/backend/target-execution/summary-emission.mjs`
- `tools/harness/execution/run-make-node-tool-cli.mjs`
- `tools/harness/execution/run-vitest-manifest-phase.sh`
- `tools/harness/browser/run-browser-e2e-batch.sh`
- `tools/harness/browser/run-playwright-webserver-batch.sh`
- `tools/harness/output/test-output/runners/go.mjs`

Lifecycle implementation:

- `internal/testutil/suiteservices/lifecycle.go`
- `tools/testservices/main.go`

Tests and acceptance fixtures:

- `tools/testservices/main_test.go`
- `tools/harness/phase-accounting/tests/test-run-phase-slice.sh`
- `tools/harness/generated-artifacts/tests/test-json-shapes.sh`
- `tools/harness/generated-artifacts/tests/test-generate-drift.sh`
- `tools/harness/scheduler/tests/test-check-scheduler.sh`
- `tools/harness/tests/test-run-make-sequence-fast.sh`

The topology projections were regenerated from owner inputs through
`make phase-schedules`; no generated root was hand-edited.

### 12.3 Validation and retained evidence

The final narrow-to-broad acceptance passed:

- `make lint-markdown`
- `make generated-artifact-policy-check`, root ending `p63305`
- `make json-shape-check`, root ending `p64074`
- `make phase-ledger-drift`, root ending `p64448`
- `make phase-schedule-drift`, root ending `p64831`
- `make harness-contract`
- `make run-harness-smoke-execution`, root ending `p11231`
- `make run-harness-smoke-lifecycle`, root ending `p83515`
- `make run-harness-smoke-extended`, root ending `p90978`
- `env -u RESULTS_DIR make agent-finalize`, root ending `p56451`; retained-run
  maintenance was skipped because `RESULTS_DIR` was intentionally unset
- `make check`, root ending `p47908`; 335 of 335 work units and 1,074 tests passed
  with zero failed or missing tests

After this handoff was written, final `make lint-markdown` passed, as did
`make generated-artifact-policy-check` at root ending `p42531`,
`make phase-schedule-drift` at root ending `p42718`, and
`make json-shape-check` at root ending `p42905`.

Representative exact base and frontend commands passed:

- Base unit `U-4-08`, root ending `p79601`
- Base service-backed `U-4-01`, final confirmation root ending `p46944`
- Base Vitest `U-4-WB-01`, root ending `p96427`
- Base browser `E-4-01`, root ending `p97031`
- Frontend unit `FE-U-P8-01`, root ending `p79868`

Final current-v2 cumulative FE-P8 roots end in `p58607`, `p91088`, `p23781`, and
`p56399`. Each passed 11 of 11 units and 85 tests with zero failed, missing, or
unmapped rows. Their plans use namespace `frontend`, default `through_phase`, full
completion scope, and the same 57-row resolved inventory.

Final current-v2 exact `FE-B-P8-01` roots end in `p88864`, `p1662`, `p14078`, and
`p26478`. Each passed 2 of 2 units and three tests with zero failed, missing, or
unmapped rows. Their plans contain one requested and resolved row, exact phase,
service-backed dependency scope, and selected-subset completion. Each retained
both closure-bearing accounting artifacts with only `FE-B-P8-01` closed.

All eight final service lifecycles have six unique sequential events, exactly one
applied `cleanup_started`, exactly one applied `cleanup_succeeded`, zero illegal
transitions, and terminal state `cleaned`.

### 12.4 Failures, retries, and skipped work

- JSON-shape root ending `p63486` was retry-contaminated because planner inputs
  changed after generation. `make phase-schedules` regenerated owner-derived
  outputs, after which root `p64074` passed.
- Exact base service-backed roots ending `p74902` and `p76991` exposed runtime
  reconstruction and Go manifest-filtering gaps. Exact row identity now reaches
  shard execution and finalization; final roots `p78643`, `p46944`, and the broad
  check pass confirm the repair.
- Earlier extended roots ending `p85878`, `p36615`, `p4662`, and `p46492` exposed
  stale fake-summary producers, a fake SQL generator incompatible with
  clean-before-write, a blank support-phase sentinel, and a host-load-sensitive
  completion-order assertion. Their fixtures now model current owner contracts;
  root `p90978` passed without an override.
- Initial check root ending `p57892` showed exact selected IDs leaking into full
  aggregate Go verification. The filter is now exact-context-only; root `p47908`
  passed all work.
- No duration, worker, clone, fixture, sharding, or resource baseline was changed.
  The Section 9 optimization threshold was not applied because this was
  correctness and contract remediation. DX-008 through DX-016 remain unchanged
  follow-up work and no acceptance sample was rescued by a baseline refresh.

`docs/domain.md` was consulted. This remediation changes harness mechanics and no
product-domain vocabulary, record model, workbook surface, or concept boundary.
The original investigation measurements in Sections 2 through 10 remain
historical developer-experience evidence and were not rewritten as post-change
benchmarks.

## 13. Final remediation completion record

This section records the DX-008 through DX-016 implementation and experiment pass,
DX-019 acceptance, and DX-020 handoff. Sections 2 through 10 remain the historical
investigation and baseline record. Section 11 is the canonical per-workstream root
index and disposition ledger.

### 13.1 Sequence and final dispositions

Work ran in tracker order. Each row was marked `IN_PROGRESS` before its implementation
or experiment and was resolved before the next row started:

1. DX-008 adopted the scheduler-visible setup and owned service-session DAG.
2. DX-009, DX-010, and DX-011 calibrated Vitest, Playwright, and Postgres clone
   capacity. Their experiments completed and retained the existing defaults.
3. DX-013 implemented and measured browser-group flattening, then removed it after a
   material Phase 9 regression.
4. DX-014 corrected Phase 12 execution and fixture ownership before any packing or
   lowering experiment.
5. DX-015 implemented and measured explicit scenario packing at sizes 2, 4, and 8,
   then removed the field, metadata, and planner branch because no size passed the
   global improvement threshold.
6. DX-016 completed fixture-tier proofs for the remaining clone users. None supported
   safe lowering, so the current fixture tier remains.
7. DX-012 calibrated Go CPU only after the topology was final. The current resolver
   remains.
8. DX-019 completed narrow, representative, full-check, and release acceptance.
9. DX-020 recorded the final state without modifying historical measurements.

Final status is `DONE` for DX-008 through DX-012, DX-014, DX-019, and DX-020. DX-013,
DX-015, and DX-016 are `DROPPED` with completed evidence. No item in this pass is
`TODO`, `IN_PROGRESS`, `BLOCKED`, or `DEFERRED`.

### 13.2 Adopted specification, owner, and implementation changes

DX-008 replaces two phase-slice lifecycle paths with one normalized schedule:

- `service-session-runtime.mjs` is the shared private service coordinator used by
  `check` and phase slices. It owns service environment files, leases, child lifecycle
  events, final cleanup, and scheduler-summary projection.
- `scheduler-dag.mjs` adds setup producers, service startup, child test units,
  aggregation, finalizers, and service completion to one dependency graph. Runtime
  binaries and browser prerequisites are represented by producer/consumer edges.
- The serial `runPhaseSliceSetup` loop, private `--inside-service-wrapper` flag, and
  phase-slice re-execution path are removed. There is no private legacy alias.
- Setup and readiness failure remain attributed to their real units. Cleanup remains
  scheduler-owned and executes once after success, failure, or interruption.

The NLSpec and topology owner input now declare the already-supported
`VITEST_MAX_WORKERS` and `PLAYWRIGHT_WORKERS` phase-slice inputs. This closes the
forwarded-but-rejected input drift found during DX-009 and DX-010. It does not change
the defaults: direct Vitest remains 4, scheduler-owned `frontend-unit` remains 2, and
Playwright remains 3.

DX-014 changes the Phase 12 owner classification, not product behavior. The 76
repository/in-memory Network Flow contract selectors now use `backend_unit`, have no
fixture policy, and live in `phase12_network_flow_contract_test.go`. The full Phase 12
inventory remains 112 rows; the service-backed subset is 15 rows and now contains six
real clone-claiming harness-control/process work units. The coverage ledger and
generated schedules were regenerated from the corrected owner map.

Final acceptance also exposed a release-sequence composition defect. `release-check`
now runs `seaweedfs-compatibility` and `seaweedfs-migration-preservation` as explicit
sequence steps, then invokes `seaweedfs-release-gate` with prerequisites suppressed
and a fail-closed marker proving that the sequence completed them. A direct skipped
gate without that marker still fails. The release evidence test covers the legal
sequence-owned case, and direct `seaweedfs-release-gate` retains its full prerequisite
list.

### 13.3 Measurements and decisions

Every experiment used one cold run plus three uncontaminated warm runs for each
candidate. Section 11 records every retained root suffix. Failed, stale, cold, and
retry-contaminated runs were not used to establish a performance win.

| Workstream | Candidate warm medians | Decision |
| ---------- | ---------------------- | -------- |
| DX-008 setup/service DAG | Phase 4 31.388s; Phase 9 31.713s; Phase 12 19.346s | Adopted. Each representative phase improved by more than 5 seconds from the Section 3 baseline with unchanged row, assertion, artifact, and lifecycle inventories. |
| DX-009 Vitest workers | 2: 16.084s; 4: 10.533s; 6: 8.520s | Retain 4. Six improved 19.1% and 2.013s, below both gates. |
| DX-010 Playwright workers | 2: 36.986s; 3: 31.360s; 4: 28.410s | Retain 3. Four improved 9.4% and 2.950s, below both gates. |
| DX-011 clone capacity, Phase 4 | 4: 30.462s; 6: 32.184s; 8: 32.522s | Retain the auto formula. Four improved only 6.3% and 2.060s over eight. |
| DX-011 clone capacity, Phase 12 | 4: 27.883s; 6: 21.689s; 8: 19.964s | Eight was fastest; no single lower value passed both representative phases and therefore no full-check confirmation was warranted. |
| DX-013 browser flattening | control 31.360s; flattened 38.531s | Dropped and removed. The candidate regressed 22.9%. |
| DX-015 scenario packing | isolated 22.056s; size 2 20.700s; size 4 19.295s; size 8 20.238s | Dropped and removed. Size 4 improved 12.5% and 2.761s, below both gates. |
| DX-012 Go CPU, Phase 9 | 8: 32.836s; 12: 32.128s; 16: 32.462s | Retain the auto resolver. Twelve cleared neither gate. |
| DX-012 Go CPU, Phase 12 | 8: 17.463s; 12: 16.812s; 16: 17.463s | Retain the auto resolver. Twelve cleared neither gate. |

DX-016 was a proof workstream rather than a timing contest. Every remaining clone
user requires committed cross-process runtime-fixture reset behavior, so transaction
rollback and group-clone reuse were rejected. No fixture token was added and no
unproven dirty-table reset surface was accepted.

Across accepted samples, row and product-assertion inventories were unchanged, no
new readiness retry or flake was introduced, lifecycle cleanup remained legal, at
least 10.15 GiB host memory remained available, and measured swap growth was zero.
CPU and I/O blocker observations remain visible in scheduler summaries; no physical
host limit or duration baseline was changed.

### 13.4 Compatibility and migration

The public commands `phase-slice` and `service-backed-slice`, their command IDs,
`ROWS` semantics, public exit codes, failure classes, result-root layout, and
`cartulary.phase_slice_plan.v2` remain unchanged. Scheduler summaries expose existing
setup and service work more truthfully but do not reinterpret child status. The new
plan work units stay within v2's private extension zone; there is no new schema ID.

The only removed interface is the private phase-slice service-wrapper re-execution
flag. It has no compatibility alias because scheduler-owned service sessions fully
replace it. The rejected `scenario_process_group` candidate is absent from the
NLSpec, owner maps, generated plans, and implementation. No downstream migration is
required for it.

Consumers of generated Phase 12 schedules will observe the truthful movement of the
76 Network Flow rows from service-backed integration work to fixture-free backend
unit work. Row IDs, evidence ownership, selected symbols, assertion titles, and total
phase coverage do not change. Maintainers should update the phase map first and use
the Make-owned ledger/schedule generators for any future classification change.

No legacy worker default, clone formula, CPU formula, fixture token, persistent
service pool, reset behavior, cleanup behavior, or public diagnostic surface was
carried forward or added without passing evidence.

### 13.5 Final validation and retained roots

Narrow contract, planner, scheduler, lifecycle, and release-evidence tests passed
through `make harness-contract`, `make run-harness-smoke-execution`,
`make run-harness-smoke-lifecycle`, and `make run-harness-smoke-extended`. The latest
final execution and extended smoke roots ended `p88778` and `p14436`; lifecycle root
ending `p84632` also passed. The release root reran applicable harness contracts.

Owner and generated-artifact checks passed:

- `make generated-artifact-policy-check`
- `make json-shape-check`
- `make generate-drift`
- `make phase-ledger-drift`
- `make phase-schedule-drift`
- `make go-test-duration-baseline-drift RESULTS_DIR=<successful Phase 12 root>`
- `make service-backed-make-target-duration-baseline-drift`

The standalone duration-drift roots ended `p58725` and `p58765`. The final release
root `20260712T160719Z-p1334` reran the generated policy, JSON shape, generation,
ledger, schedule, harness-contract, and other release-owned drift gates successfully.
No baseline was refreshed.

Exact and cumulative frontend acceptance passed:

- `FE-U-P8-01`, root ending `p51122`
- exact `FE-B-P8-01`, root ending `p52541`
- cumulative FE-P8, root ending `p68459`, with 11 of 11 units and 85 tests

Representative final phase acceptance passed:

- Phase 4 base root ending `p3858`; service-backed root ending `p23634`
- Phase 9 base root ending `p43001`; service-backed root ending `p60862`
- Phase 12 base root ending `p78039`; service-backed root ending `p580`

Browser acceptance passed with functional root ending `p13909` (97 tests), stateful
root ending `p31004` (22 tests), accessibility root ending `p49374` (20 tests), and
visual root ending `p62451` (28 tests). Reset, taint, leak, fixture, and cleanup
checks remained green.

`env -u RESULTS_DIR make agent-finalize` passed at final root ending `p78448`.
Retained-run
maintenance was intentionally skipped because `RESULTS_DIR` was unset. Final
`make check` and `make release-check` passed together at root
`.cartulary/test-results/20260712T160719Z-p1334`: check completed 259 of 259 work units
and 1,074 tests; release completed 14 of 14 work units and 1,131 tests, with zero
failed or missing evidence.

### 13.6 Failures, retries, and rejected samples

- A first Go duration-drift invocation at root ending `p58571` omitted the required
  `RESULTS_DIR`. It was an invalid invocation, not a regression; the correctly scoped
  root ending `p58725` passed.
- Release root ending `p85225` encountered a load-sensitive frontend-unit failure.
  Isolated root ending `p68520` passed, and the final full release rerun passed. The
  failed sample was not retained as acceptance evidence.
- Release roots ending `p70293` and `p18765` exposed duplicate execution of
  `backend-process` through the SeaweedFS gate's Make prerequisites. The sequence now
  owns and reuses those prerequisites instead of rerunning them into one result root.
- Root ending `p79568` exposed the caller/validator mismatch in the sequence-satisfied
  marker. The shared fail-closed predicate and its regression test corrected it; a
  focused gate pass preceded the final `p1334` release pass.
- DX-013 and DX-015 candidate roots remain in Section 11 for auditability, but their
  prototypes were completely removed. DX-016's retained proofs record why lowering
  was rejected. None of these results was converted into an accepted optimization by
  changing a baseline or weakening isolation.

### 13.7 Remaining risks and future guidance

The coarse browser target remains because flattening was slower on this host. Its
nested process demand is therefore still less visible than ideal; future browser
growth may justify a different topology, but any retry should start from a new owner
design and the same isolation and performance gates rather than reviving the removed
prototype.

The six remaining Phase 12 clone users are legitimate I/O consumers. Future lowering
requires new fixture-tier proofs showing restored Postgres, auth, idempotency, job,
object-store, observer, process, and migration surfaces. Current evidence does not
authorize optimistic grouping or transaction rollback.

Worker and CPU results are local developer-experience evidence for the declared WSL2
host, not portable or Core 05 benchmark claims. Overrides remain available for local
investigation, while public defaults remain conservative and owner-tested. Physical
WSL/Docker resource changes remain out of scope.

`docs/domain.md` was consulted throughout. Phase rows, harness fixtures, schedulers,
and service sessions remain implementation-support concepts; this pass changes no
product-domain vocabulary or boundary.
