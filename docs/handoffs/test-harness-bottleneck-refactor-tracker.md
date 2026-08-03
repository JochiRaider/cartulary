# Test Harness Bottleneck Refactor Tracker and Handoff

## 1. Scope and Source Posture

| Item | Current posture |
| --- | --- |
| Primary seam | Public test target to workload selection, scheduling, resource and fixture lifecycle, execution, and evidence publication |
| Output path | `docs/handoffs/test-harness-bottleneck-refactor-tracker.md` |
| Mode | Tracker-gated implementation complete; WF-00 through WF-09 and CP-00 through CP-12 are `DONE`; clean-source multi-window performance publication was explicitly `DROPPED` by user direction and no performance claim or v3 reference was published |
| Allowed change in this phase | The active workstream's adopted harness owner, authored machine inputs, implementation, focused tests, generated projections through Make, and this tracker |
| Future implementation scope | Authored harness owner documents and machine inputs, harness tests and runtime helpers, test-only fixture support, schemas, and generated projections refreshed through Make |
| Non-goals | No product behavior, production API, assertion weakening, dependency upgrade, generated-file hand edit, or performance claim based only on a smaller workload |
| Planning baseline | `244f639d19d2e307748b1a8f17b62aa711d5f64d` on `main`, one commit ahead of `origin/main`, with a clean worktree before this tracker was added |
| Implementation start | `9af2e005e75bb636a9114b3ff90cbe5cc8061d76` on `main`; the planning baseline remains historical provenance rather than the implementation HEAD |
| WF-02 restart baseline | `f4b0176f877f3edb0fe003ee35e5bb0234c77bea` on `main`, equal to `origin/main`, with a clean worktree before the WF-02 tracker gate |
| WF-03 restart state | Same HEAD `f4b0176f877f3edb0fe003ee35e5bb0234c77bea`; dirty only with the complete WF-02 review unit; 146/98 target and 1,019-row disposition revalidated; WF-02 final lint passed at `.cartulary/test-results/20260802T174829Z-p4057900` |
| WF-04 restart state | Same HEAD; dirty only with completed WF-02/WF-03 review units; disposition check and `harness.test_catalog` task guide passed; WF-03 final lint `.cartulary/test-results/20260802T180850Z-p4093309` |
| WF-05 restart state | Same HEAD; dirty only with completed WF-02 through WF-04 review units; disposition check and `harness.browser` task guide passed; WF-04 final lint `.cartulary/test-results/20260802T182102Z-p4111949` |
| WF-06 restart state | Same HEAD; cumulative WF-02 through WF-05 review unit only; disposition check and `harness.command_surface` task guide passed; WF-05 final lint `.cartulary/test-results/20260802T183206Z-p4132437` |
| WF-07 restart state | Same HEAD; cumulative WF-02 through WF-06 review unit only; disposition check and `harness.generated_artifacts` task guide passed; WF-06 final lint `.cartulary/test-results/20260802T185222Z-p4159921` |
| WF-08 restart state | Same HEAD; cumulative WF-02 through WF-07 review unit only; 146 target records, 98 active public targets, and 1,019 active rows revalidated; command-surface task guide passed; WF-07 final lint `.cartulary/test-results/20260802T191222Z-p15121` |
| WF-09 restart state | Same HEAD; cumulative WF-02 through WF-08 review unit only; 143 task records, 98 active public targets, 47 canonical measured bindings, and 1,019 active rows revalidated; WF-08 final lint `.cartulary/test-results/20260803T005038Z-p1884498`; v2 baseline remains archival and CP-12 starts without accepted v3 performance evidence |
| WF-09 completion state | Same HEAD; cumulative WF-02 through WF-09 review unit remains uncommitted; 143 task records, 98 active public targets, 47 canonical measured bindings, and 1,019 active rows close; all functional and structural gates pass at source digest `sha256:6df8593b8b23f64a71d2521f0118ba747fbd31a8f13753308a2335b94b183068`; performance publication was explicitly skipped and the v2 reference remains archival |
| Analysis posture | `temp/analysis-notes.md` is diagnostic evidence. Its inconsistent checked-in baseline and single-sample observations are not publishable performance proof. |
| Compatibility decision | Selective hard cutover: retain useful public Make vocabulary and roles; permit coordinated command, schema, artifact, lifecycle, topology, and redundant-target changes without compatibility aliases |
| Coverage decision | Tier by purpose: fast feedback may be narrower, while full, CI, and release entry points preserve the complete correctness and release-evidence union |
| Validation decision | Accumulated affected-test ledger: each source version reruns its deterministic affected closure; unchanged compatible passes carry forward with their original provenance |

This tracker applies the structure of
`docs/handoffs/cartulary_modular_refactor_planning_framework.md` to one primary
harness seam. It began as a planning handoff and now controls implementation.
Every implementation session MUST update it as the restart, validation,
risk, and sequencing record.

The user's explicit direction changes the normal authority posture for this
refactor: the adopted Testing Harness NLSpec is the owner of current behavior,
but its scheduling, lifecycle, telemetry, artifact, and compatibility rules are
not preservation requirements by default. Relevant current clauses are
classified in Section 4 as `RETAIN`, `REVISE`, or `DELETE`. A revised owner
contract MUST be adopted before implementation projects the new behavior into
machine inputs or code. Product owners and test assertions remain authoritative
for product behavior.

The key words `MUST`, `MUST NOT`, `SHOULD`, and `MAY` are normative for future
implementation of this tracker:

| Term | Meaning |
| --- | --- |
| `MUST` / `MUST NOT` | Binary completion requirement or prohibition |
| `SHOULD` / `SHOULD NOT` | Expected design; deviation requires evidence, rationale, risk, and compensating validation in the handoff |
| `MAY` | Optional action still constrained by the scope and completion gates |

Decision states are closed:

| State | Meaning |
| --- | --- |
| `RESOLVED` | The tracker defines the target decision. Implementation remains gated by its prerequisites. |
| `IMPLEMENTATION_GATED` | The target is known, but source work waits for named evidence or an owner change. |
| `DEFERRED` | Evidence does not justify change in this refactor. The current behavior remains. |
| `BLOCKED: <reason>` | The named work stops until the stated condition is repaired. Independent work may continue. |

### Sources inspected

Planning and owner sources inspected directly:

- `AGENTS.md` instructions supplied for the repository session
- `temp/analysis-notes.md`
- `docs/testing-harness-nlspec.md`
- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `docs/handoffs/tasksdecisions-module-refactor-tracker.md`
- `docs/handoffs/ui-contracts-module-refactor-tracker.md`

Authored machine owners and generated projections inspected directly or queried:

- `tools/task_surface_owner.json`
- `tools/test_catalog_owner.json`
- `tools/execution_topology_manifest.json`
- `tools/scheduler_resource_registry.json`
- `tools/browser_e2e_batch_manifest.json`
- `tools/task_surface_manifest.json`
- `tools/scheduler_manifest.json`
- `tools/execution_topology_render_index.json`
- `tools/harness_public_target_duration_baselines.json`
- relevant schemas under `tools/schemas/**`

Representative implementation and test paths inspected directly:

- `tools/harness/scheduler/scheduler/engine.mjs`
- `tools/harness/scheduler/scheduler/state.mjs`
- `tools/harness/scheduler/scheduler-resource-policy.mjs`
- `tools/harness/scheduler/scheduler-resources.mjs`
- `tools/harness/browser/browser-stage-scheduler-cli.mjs`
- `tools/harness/browser/run-browser-e2e-batch.sh`
- `tools/harness/finalization/agent-finalize-action-plan.mjs`
- `tools/harness/finalization/agent-finalize-action-cache.mjs`
- `tools/harness/tests/test-harness-contracts.mjs`
- `internal/testutil/pgtest/pgtest.go`

Adjacent helpers under `tools/harness/**`, owner manifests under
`contracts/verification/owners/**`, and test-family manifests under
`tools/test_families/**` were inventoried by path and targeted search. They MUST
be inspected directly before their workstream edits them; a search hit is not an
edit source.

## 2. Planning-Baseline Repository Inventory

This section preserves the execution-control inventory used to authorize the
refactor. It is historical after the WF-08 cutover; the current post-cutover
inventory and source posture are recorded below and in the WF-08 handoff.

### Execution-control inventory

| Area | Current owner or implementation | Current responsibility | Principal finding | Target disposition |
| --- | --- | --- | --- | --- |
| Public task surface | `tools/task_surface_owner.json` | 146 target records, including 98 active public targets, plus command IDs, inputs, output policies, inclusion sets, and observability policy | The performance roster requires 47 targets, while the checked-in baseline does not close over it. | Retain as the authored public owner; revise target roles, inputs, IDs, and observability projection. |
| Test selection | `tools/test_catalog_owner.json`, `tools/test_families/**` | Owner routing for 1,019 active rows, runner selectors, evidence classes, runtime/resource/fixture profiles, and default-check selection | Tier intent is distributed and `default_check` cannot express the new entry-point roles. | Add one minimum execution tier per active row and generate target membership. |
| Execution topology | `tools/execution_topology_manifest.json` | Runtime, resource, fixture profiles; sequence and service-backed schedules | `test-fast` and `test` have explicit one-job local-to-service barriers; CI and release work wait on whole `check`. | Replace phase schedules with authored graph composition and capability declarations. |
| Scheduler resources | `tools/scheduler_resource_registry.json` | Logical capacity names and automatic policies | Capacity is primarily CPU-derived; the measured host resolved clone capacity to eight and browser work still serialized. | Replace family-specific capacity logic with one run capability snapshot and unit-sized claims. |
| Generated schedule | `tools/scheduler_manifest.json` | Projected check and service-backed work units | Direct and aggregate paths project different units and scheduling behavior. | Generate one graph representation used by every selector. Never hand-edit. |
| Browser manifest | `tools/browser_e2e_batch_manifest.json` | Stages, groups, sessions, reset and isolation metadata | Direct stages create one unit per session and the shell batch loops groups serially. | Preserve semantic group and reset metadata; project each group as a work unit and remove batch-loop scheduling. |
| Scheduler runtime | `tools/harness/scheduler/**` | Multiple scheduler facades, resource reservation, process execution, event and summary emission | Priority reservation can prevent otherwise fitting work; nested scheduler families fragment the critical path. | Converge on one graph compiler, scheduler engine, resource broker, and evidence projector. |
| Browser runtime | `tools/harness/browser/**` | Stack ownership, stage scheduling, resets, Playwright invocation, leaf finalization | A generated stage resource is forced to capacity one and `run-browser-e2e-batch.sh` serializes groups within a session. | Use scheduler-owned stack leases and group commands; retain only lifecycle adapters and semantic reset actions. |
| PostgreSQL fixtures | `internal/testutil/pgtest/pgtest.go` | Suite template, row/package/group databases, transactions, resets, migration scratch | The default template-clone path creates substantial clone pressure; group reuse is narrow and migration replay is repeated. | Introduce explicit lease capabilities, asynchronous pool replenishment, and owner-reviewed fixture migrations. |
| Finalization | `tools/harness/finalization/**` | Schema, duration, scheduler, generation, and drift actions with action caching | No-results finalization spends about 16.8 seconds running generation and drift as separate render passes; broad cache inputs repeatedly miss. | Render once into scratch, validate and compare once, publish atomically when selected, and delete ineffective special caching. |
| Harness contracts | `tools/harness/tests/test-harness-contracts.mjs` | Most harness contract fixtures in a 6,557-line, 104-test file | Two Node test files expose little file-level parallelism and repeatedly load or derive shared indexes. | Partition by semantic harness owner and share one immutable per-run index artifact. |
| Observability | `tools/harness/observability/**`, duration-accounting helpers and schemas | Target spans, scheduler events, baselines, hotspot and pressure summaries | Browser leaf baselines measure dispatch rather than work; wrappers are unattributed; portfolio cardinality and sums disagree. | Make unit events the timing source and derive all target, critical-path, and resource summaries from them. |

Generated files listed by `tools/generated_artifact_policy.json`, generated Make
includes, scheduler manifests, browser batch projections, and topology render
indexes MUST be changed only by updating authored owners and running the
Make-owned generators.

### Current execution shape

```text
public Make target
  -> target-specific sequence/check/service-backed/test-slice facade
    -> nested target or scheduler
      -> coarse session/shard work unit
        -> shell or runner loop over actual test groups
          -> separate finalizers and target summaries
```

This shape hides runnable work below scheduler units, creates whole-target
barriers, duplicates planning/finalization, and makes target timing depend on
which wrapper happened to emit a span.

### WF-08 post-cutover inventory

The current generated task surface contains 143 records: 98 active public
targets, ten check-internal targets, and 35 internal helpers. All 47 measured
public targets route through the canonical graph and event projections. The
active catalog remains exactly 1,019 rows. The three-record reduction from the
146-record WF-02 baseline is the intended retirement of obsolete internal
phase helpers; no public Make name was removed.

The live execution shape is now:

```text
public Make target
  -> canonical selector and work-graph compiler
    -> one resource-fitting scheduler and run-scoped fixture broker
      -> canonical work units and unit events
        -> run, target, hotspot, pressure, and performance projections
```

Static negative fixtures retain removed IDs and schema names only to prove
that they are rejected. The archival v2 public duration baseline retains its
historical `release-browser-readiness` row as bytes; no current command reads
or translates it.

### Measured bottleneck baseline

The measurements below are evidence from `temp/analysis-notes.md`. They are
prioritization inputs, not acceptance baselines.

| Bottleneck | Diagnostic wall time | Critical-path or blocking evidence | Structural cause |
| --- | ---: | --- | --- |
| `browser-e2e` | 322.77 s | 308.91 s critical path; 290.63 s wall-relevant `browser_stack` blocking; 878.50 s cumulative overlapping queue wait | One stage lane, coarse sessions, and serial group loop |
| `test` | 272.74 s | 241.81 s service-backed branch; 187.06 s wall-relevant resource blocking | Local/service phase barrier, clone pressure, excessive Go shards |
| `test-fast` | 175.60 s | 143.79 s service-backed branch; 91.65 s wall-relevant blocking | Nearly full service-backed inventory and serial aggregate composition |
| `check` | 189.26 s | Backend integration 136.27 s; 139.43 s wall-relevant blocking | Single suite service lane and resource-heavy Go work |
| `release-check` | 314.30 s | `check` 180.96 s, then release browser 96.18 s | Whole-`check` dependency before independent release work |
| `ci` | 208.45 s, failed | `check` 171.31 s, then about 37.14 s of post-check work | Whole-`check` barrier and incorrect duration drift after successful tests |
| `agent-finalize` | about 22.5 s | `generate-drift` 9.90 s and `generate` 5.78 s | Same generated structure rendered twice; action cache does not help |
| `harness-contract` | 27–28 s | Owner evidence audit about 4.40 s; target finalization about 3.02 s; repeated catalog/index work | Giant sequential contract suite and repeated setup |
| `go-vulncheck` | 18.13–19.20 s | Scanner dominates; checked-in reference 2.55 s | External whole-repository analysis, no safe same-run deduplication or freshness-keyed result reuse |

Named service-backed hotspots that MUST retain their own work items are:

- extension integration and migration work, observed at 35.21 seconds for one
  shard and as much as 46.16 seconds of aggregate fixture work;
- recovery process evidence, observed at 32.57 seconds;
- PostgreSQL support, observed at 22.31 seconds;
- the default browser session at 118.99 seconds and stateful-default session at
  108.98 seconds; and
- unattributed standup, SeaweedFS, fallow-static, and duration-coverage wrapper
  envelopes.

### Measurement defects to repair before optimization claims

The authored observability policy requires 47 targets. The checked-in public
duration baseline declares 48 portfolio targets, lists only 46 public names,
omits `openapi-compatibility-check`, and reaches 48 rows by mixing in two
internal gates. Its declared total is 2,721,126 ms while its 46 listed public
medians sum to 2,636,302.5 ms, a difference of 84,823.5 ms or 3.22%.

Additional defects are:

- browser leaf references of two to three milliseconds measure dispatch or
  summary work rather than the 20–131 second browser workload;
- Go command overhead is planned at about 67.3 seconds per shard while observed
  overhead is about 0.4–3.2 seconds, causing 127 false overplanned-shard reports
  in the diagnostic CI run;
- finalizer/wrapper spans are sometimes missing or counted inconsistently;
- parameterized slice observations do not always match their canonical owner
  workload; and
- several observations are single samples rather than qualified windows.

The unrelated Go-vet failure reported by the analysis in
`internal/modules/revisions/incident_bundle_portability.go` MUST be revalidated
at implementation start and recorded as related or unrelated evidence. It MUST
not be silently attributed to this refactor.

## 3. Target Boundary and Architecture

### Primary seam

The target module is the harness execution-control plane. Its public facade is a
Make-owned target invocation. It hides workload selection, graph expansion,
resource and fixture leasing, process execution, caching, failure propagation,
cleanup, event capture, artifact publication, and target-summary projection.

Runner adapters own only translation to and from Go, Vitest, Playwright, or
shell protocols. Fixture providers own lifecycle mechanics. Catalog owners own
test identity and selection. Public targets MUST NOT encode scheduler policy in
shell loops or nested Make dependencies.

The target flow is:

```text
public Make target + declared inputs
  -> target selector and tier policy
    -> canonical workload graph compiler
      -> one resource/fixture-aware scheduler
        -> runner and lifecycle adapters
          -> canonical unit event stream
            -> target, hotspot, pressure, and run-summary projections
```

### Execution tiers

Add the closed test-row enum `minimum_tier = fast | standard | full | release`.
For test rows, the order is monotonic: `fast < standard < full < release`.
Policy, packaging, and release units are selected by target policy rather than
masquerading as product test rows.

| Public entry point | Test-row selection | Additional policy work | Explicit exclusions |
| --- | --- | --- | --- |
| `test-fast` | `minimum_tier=fast` | Only readiness needed by selected fast rows | Browser measurement/visual/a11y, full migration replay, recovery/process isolation, release/security policy unless a row is explicitly proven fast and required |
| `check` | Fast and standard rows | Static analysis, type/boundary checks, generated drift, standard security and evidence checks | Full-only functional inventory and release-only claims |
| `test` | Fast, standard, and full rows | Test evidence collation only | Release-only measurement, packaging, and publication policy |
| `ci` | Full functional inventory | CI security, policy, deployable-shape, and duration-health units | Release-only claims and packaging unless independently required by CI policy |
| `release-check` | Complete release tier | CI policy plus license, SBOM, compatibility, build, release browser, visual, accessibility, measurement, and release-evidence units | None from the adopted release inventory |

Leaf browser/backend/frontend targets and owner slices remain explicit graph
selectors. An omitted owner slice remains the complete owner inventory, not the
entry point's tier-limited subset. Every active row MUST have one minimum tier,
and every active row MUST be reachable from `test`, `ci`, or `release-check` as
appropriate. A removed row requires an owner-backed redundancy record naming
the surviving row and proving equivalent behavior, runner, fixture, and
evidence semantics.

### Canonical work-unit contract

The authored graph model MUST project every schedulable action to this logical
contract. Exact schema names and command IDs are assigned in WF-02 and receive
new versions in one hard cutover.

| Field | Required semantics |
| --- | --- |
| `unit_id` | Stable semantic identity independent of aggregate target and queue position |
| `owner_id` | Verification or harness owner responsible for the work |
| `kind` | Closed runner, readiness, lifecycle, finalizer, policy, or artifact kind |
| `command` | Fully resolved executable, arguments, and declared environment projection |
| `needs` | Stable predecessor unit IDs; no implicit phase order |
| `resource_claims` | CPU, memory, IO, process, database, object-store, browser, and exclusive-write tokens required while running |
| `fixture_lease` | Explicit fixture capability and affinity; `none` is explicit and there is no fallback |
| `cache_policy` | `none`, `same_run`, or `content_addressed`, with a registered input-closure profile |
| `timeout_ms` | Owner-declared bounded timeout |
| `evidence_outputs` | Required schemas and artifact roles produced by the unit |
| `failure_policy` | Descendant blocking, independent-work continuation, and aggregate effect |
| `estimated_work_ms` | Qualified work duration only; wrapper/setup cost is modeled by its own unit |
| `semantic_digest` | Digest of selection, command, dependencies, profiles, and evidence contract |

One `unit_id` and semantic digest MAY contribute to multiple target summaries in
one graph, but it MUST execute no more than once. A direct target and its
aggregate projection MUST resolve the same unit identities for the same
semantic selection.

### Scheduler algorithm

The unified scheduler MUST use the following deterministic policy:

1. Validate all IDs, dependency closure, cycles, resource names, maximum claims,
   fixture capabilities, cache profiles, and output contracts before starting
   child work.
2. Capture one immutable host-capability snapshot for the run.
3. Compute each ready unit's rank as its qualified estimated work plus the
   longest estimated downstream path.
4. Among resource-fitting units, start the highest-rank unit. Stable `unit_id`
   ordering breaks ties.
5. Backfill with any lower-ranked fitting unit; a blocked high-ranked unit MUST
   NOT reserve resources it does not hold.
6. Increase a ready unit's effective rank by monotonic waiting age so continual
   backfill cannot starve it.
7. Stop scheduling descendants after a required predecessor fails. Continue
   independent diagnostic work only when target failure policy permits it.
8. On cancellation, stop admission, terminate owned process groups, release or
   destroy leases, and publish a partial non-success summary.

The initial implementation MUST remain deterministic from the authored graph,
qualified cost model, and captured capacity. Live self-tuning is out of scope;
capacity changes between runs are permitted and recorded.

### Capacity and resource model

Capacity discovery MUST inspect the effective CPU set, cgroup or host memory,
process limit, writable-volume availability, and declared service readiness.
Every resource profile declares CPU tokens, memory bytes, IO weight, process
slots, and scarce service capabilities. Parallel capacity for a unit class is
the minimum capacity implied by each declared dimension. Missing or invalid
capacity data fails safe to one lane for the affected scarce resource rather
than inventing host capacity.

Browser-stack capacity is the minimum of CPU, memory, process, and port-lane
capacity. PostgreSQL clone, reset, and migration capacity is independently
declared and calibrated; a general `postgres=32` token is not proof that 32
clone operations are useful. Environment or Make overrides remain explicit
public inputs, are range checked, and appear with their source in the run
manifest.

Resource telemetry MUST report requested capacity, resolved capacity, peak
use, saturation intervals, per-unit queue duration, blocking resources, and the
units that actually held those resources.

### Service and fixture plane

One run-scoped broker owns infrastructure readiness and cleanup. It MAY attach
to already managed Postgres and object-store services, but every application or
browser stack, database, namespace, and process it creates is owned by a lease.
Borrowed resources are never closed by the broker.

The fixture capability set is:

| Capability | Isolation and reuse rule |
| --- | --- |
| `none` | No service state; omission is invalid. |
| `postgres_transaction` | One compatible database lease, per-test transaction, mandatory rollback; unavailable to tests requiring committed cross-connection state. |
| `postgres_group` | One database for a declared execution group with owner-defined reset between groups and affinity within the group. |
| `postgres_dedicated` | Exclusive database for committed-state, process, recovery, or isolation-sensitive work. |
| `postgres_migration` | Empty or milestone database for tests that inspect migration behavior; full-chain replay remains mandatory where it is the assertion. |
| `object_store_namespace` | Unique bucket or prefix with bounded cleanup and no cross-unit visibility. |
| `managed_process` | Dedicated process group and runtime-binary identity with bounded termination. |
| `browser_stack` | Application server, database/object namespace, port set, and browser runtime identity leased as one lane. |

The broker SHOULD create clean dedicated databases ahead of demand and
replenish the pool asynchronously while tests run. A lease that fails reset,
health, or contamination checks is destroyed and replaced, never returned to
the pool. Infrastructure setup MAY retry once on a newly created lane; product
test failures MUST NOT be retried by scheduler policy.

Migration work retains at least one full migration-chain proof for every owned
chain. Tests that require only a known migrated state MAY use a digest-keyed
milestone template after an owner test proves it was produced by the same
migration chain. Extension, recovery, and PostgreSQL support hotspots each need
an owner-specific fixture audit before profile conversion.

Compatible Go rows share a process by exact module, package, runtime, fixture,
isolation, selector, environment, and evidence key. Remove the fixed eight-test
and twelve-second packing rules. For each compatibility group, request twice as
many predicted shards as available Go CPU lanes, distribute rows by
longest-processing-time order, and split a package group only when its predicted
weight exceeds twice that derived target shard weight and its selectors are
independently executable. Go JSON remains the source for row-level outcomes.

### Browser execution

Each browser manifest group becomes an ordinary work unit. The scheduler leases
stack lanes; shell code no longer owns group ordering or capacity.

- Isolated functional, accessibility, and read-only visual groups MAY run on
  separate lanes subject to host tokens.
- Stateful groups declare an affinity key and explicit dependency chain. Only
  those groups share a lane, and resets are graph units at the exact required
  boundary.
- Visual validation MAY run concurrently when outputs are read-only. Snapshot
  update units claim an exclusive write key per snapshot output and cannot
  collide.
- Measurement units enter only when no ordinary unit holds the CPU, IO, browser,
  or process tokens covered by the measurement profile. Their quiet-lane rule
  does not serialize unrelated work before that readiness boundary.
- A failed or unhealthy stack lane is quarantined and cleaned. Its test is not
  retried unless the failure is classified as infrastructure before product
  work began.

The current generated stage-capacity-one resource and
`run-browser-e2e-batch.sh` group loop are deleted after parity. Direct and
aggregate browser selectors use the same graph and lifecycle adapter.

### Aggregates, deduplication, and caching

Aggregate targets compile a union graph rather than invoke child targets.
`test-local`, `test-fast-service-backed`, `test-service-backed`, and
`check-service-backed` are transitional internal phase wrappers and have no
continuing value after graph cutover. The WF-02 restart inventory found no
direct consumer or distinct public diagnostic role for
`release-browser-readiness`; its work is selected directly by `release-check`
and the internal target is removed at cutover. `browser-e2e-support` remains as
distinct release-support work.

Identical readiness, build, scanner, test, and finalizer units are deduplicated
within one graph. License, SBOM, service readiness, static checks, builds,
browser stacks, and scanners start when their actual inputs are ready rather
than after all of `check` succeeds. Resource ranks prevent noncritical release
work from starving the functional critical path.

Cross-run content caching is allowed only when a registered profile closes over
tracked and relevant untracked source, configuration, declared environment,
tool and runtime digests, dependency/lock inputs, selection, expected output
schemas, and helper implementations. A hit validates output digests and still
emits a current-run unit event and target projection. Missing, corrupt, stale,
or mismatched entries are misses; they never produce success.

Stateful tests, service lifecycle, migration assertions, browser work,
measurement, cleanup, destructive safeguards, generated drift verdicts, and
release publication remain uncached. Vulnerability results are cacheable only
when a lightweight freshness check resolves a stable vulnerability-database
revision and that revision is part of the key; if no revision can be proven,
the scan runs. Security scan findings and exit behavior must match an uncached
scan exactly.

### Evidence and artifact model

The hard-cutover run format contains:

| Artifact | Required contents |
| --- | --- |
| `run-manifest.json` | Run/command ID, selected target and declared inputs, source/toolchain/system digests, capability snapshot, graph digest, start time, and cache mode |
| `unit-events.ndjson` | Ordered queued, admitted, started, cache, fixture, resource, completed, failed, skipped, cancellation, and cleanup events with monotonic offsets |
| `run-summary.json` | Final status, failure class, unit counts, wall interval, critical path, resource pressure, cache accounting, and artifact refs |
| `target-summaries/<target>.json` | Projection of relevant unit events, inclusive/exclusive wall, children, status, and evidence refs without rerunning work |

Target summaries, pressure summaries, hotspots, and duration baselines are
derived from the canonical event stream. Browser leaves use their group and
lane intervals, not dispatch spans. Wrapper/setup time is either a first-class
unit or explicitly reported as unattributed error; it is never silently added
to every leaf weight.

The baseline writer MUST prove equality among the authored 47-target roster,
eligible command/profile bindings, observed target names, baseline rows, target
count, and recomputed portfolio sum. Internal gates have a separate roster and
total. Old artifacts remain archival bytes but no new command reads them as
current input after cutover; there are no compatibility readers, dual schema
writers, or aliases.

## 4. Current Contract Disposition Register

This register covers the Testing Harness NLSpec clauses affected by the target
seam. Unlisted product-selector, security, and test-support clauses remain
unchanged unless a workstream adds an explicit row before editing them.

| ID | Current clause | Disposition | Successor contract or rationale | Owning workstream |
| --- | --- | --- | --- | --- |
| CD-001 | Section 1, TH-HARNESS-REQ-003: Make binding, stable command ID, and public behavior gate | `REVISE` | Retain Make names where valuable, bump command IDs for changed semantics, and permit one owner-first hard cutover of schemas, paths, inputs, and failure projection. | WF-02, WF-08 |
| CD-002 | Section 1, TH-HARNESS-REQ-004: authored owners precede generated projections and generated files are not hand-edited | `RETAIN` | This prevents generated output from becoming an accidental behavior owner. | WF-08 |
| CD-003 | Section 3.4, TH-HARNESS-REQ-017: closed runtime/resource/fixture profile registries with no implicit fallback | `REVISE` | Retain explicit closed capabilities; replace current profiles and budgets with the resource and lease model in Section 3. | WF-02, WF-04 |
| CD-004 | Section 4 default-check row placement through `default_check` | `REVISE` | Replace one Boolean with `minimum_tier`; generate default target selections from tier and policy owners. | WF-02 |
| CD-005 | Section 4, TH-HARNESS-REQ-055: every selected local check unit executes and reusable cache is absent | `DELETE` | Replace with validated per-unit `cache_policy`; cold evidence remains mandatory for performance acceptance. | WF-02, WF-06 |
| CD-006 | Section 4 cache restrictions that forbid security, drift, aggregate, service, and browser reuse as one blanket | `REVISE` | Cache only closed hermetic units; keep stateful/lifecycle/measurement exclusions and add freshness-keyed vulnerability reuse. | WF-06, WF-07 |
| CD-007 | Section 4 public target inventory and exact current membership | `REVISE` | Keep useful names, tier inventories by purpose, and prove complete full/release evidence closure. | WF-02 |
| CD-008 | Section 4 browser direct-target session ownership and stage behavior | `REVISE` | Preserve owned cleanup and semantic isolation; replace one coarse session unit with group units and stack leases. | WF-05 |
| CD-009 | Section 4.1A, TH-HARNESS-REQ-061: private helpers may move only while every public command/schema/artifact contract remains unchanged | `REVISE` | Private movement remains free; this refactor also authorizes versioned public hard cutovers after owner adoption. | WF-02, WF-08 |
| CD-010 | Section 4 warm `check-service-backed` 155,000 ms cap and fixed peer-balance policy | `DELETE` | Use same-source statistical before/after gates and resource/critical-path evidence; no fixed global seconds target. | WF-01, WF-09 |
| CD-011 | Section 8 current target/tool/scheduler summary schema families and retained paths | `REVISE` | Replace parallel timing authorities with the canonical run/event/projection model and new schema versions. | WF-01, WF-08 |
| CD-012 | Section 8 same-run helper provenance and fail-closed artifact refs | `RETAIN` | Same-run provenance remains valuable but projects through canonical unit IDs and current schema refs. | WF-03, WF-06 |
| CD-013 | Section 8 finalizer `generated_structure_refresh` runs `generate` and `generate-drift` once each | `DELETE` | Render once, validate once, compare once, and publish atomically when refresh is selected. | WF-07 |
| CD-014 | Section 10 separate sequence, check, service-backed, and test-slice scheduler families | `DELETE` | One normalized graph and scheduler serve all selectors. | WF-03 |
| CD-015 | Section 10 FIFO/priority resource reservation for earlier blocked work | `DELETE` | Resource-fit backfill with critical-path rank and aging provides throughput and fairness without unheld reservations. | WF-03 |
| CD-016 | Section 10 aggregate rules that place all CI/release work after `check` | `DELETE` | Express actual unit dependencies and allow independent work to overlap. | WF-06 |
| CD-017 | Section 10 Go packing defaults of eight symbols and 12 seconds, with backend-process limits of 16/24 | `DELETE` | Use compatibility grouping and capacity-derived LPT splitting with corrected measured work costs. | WF-04 |
| CD-018 | Section 10 browser-stack capacity and stage-lane serialization rules | `DELETE` | Use multi-dimensional stack capacity, explicit affinity, and measurement quiet-lane claims. | WF-03, WF-05 |
| CD-019 | Section 10.5 browser leaf and aggregate timing-source rules and current public-target baseline closure | `REVISE` | Unit events are authoritative; exact roster/row/sum equality is required and internal targets are separate. | WF-01 |
| CD-020 | Section 10.5 / TH-HARNESS-AC-079: one discarded warm-up, exactly two measured roots, fixed gates, and current portfolio formula | `REVISE` | Use one cold run, one discarded warm-up, five warm observations, variability-derived materiality, and semantic-digest equality. | WF-01, WF-09 |
| CD-021 | Section 11 transaction, package reset, group clone, template clone, migration scratch, and service-stack lifecycle profiles | `REVISE` | Replace policy names with explicit lease capabilities and broker ownership; retain isolation semantics only where tests require them. | WF-04 |
| CD-022 | Section 11 clone/reset caps as calibration-only adjustments | `REVISE` | Capacity still cannot replace structural repair, but pool replenishment, fixture migration, batching, and calibrated lanes are all required. | WF-04 |
| CD-023 | Section 8/10 current scheduler and helper artifact backward compatibility | `DELETE` | Historical artifacts are archival only; no live old-schema reader or dual writer survives cutover. | WF-08 |
| CD-024 | Public concise output, meaningful failure class/exit status, owned cleanup, and generated-artifact policy | `RETAIN` | These provide continuing developer and safety value independent of the old topology. | All |

## 5. Coupling and Bottleneck Findings

| ID | Finding | Consequence | Required structural response |
| --- | --- | --- | --- |
| BF-001 | Runnable browser groups are hidden inside session commands and a shell loop. | Scheduler sees too few units and cannot fill available lanes. | Project group/reset/finalizer units directly and lease stack lanes. |
| BF-002 | Local and service-backed work are separate child targets under one-job sequence schedules. | Ready work waits behind a phase boundary unrelated to dependencies. | Compile their unit union into one graph. |
| BF-003 | CI and release depend on completion of whole `check`. | License, scanning, build, service readiness, and browser setup start late. | Replace target dependencies with exact producer-unit edges. |
| BF-004 | Clone-heavy is the default profile for about 200 analyzed rows, and clone capacity saturated at eight. | Database creation dominates queue pressure and slow packages. | Require explicit fixture capabilities, migrate compatible tests, and replenish clean databases asynchronously. |
| BF-005 | Generated Go overhead is much larger than observed overhead. | Shard weights, critical-path selection, and CI drift verdicts are wrong. | Separate wrapper/setup units and rebuild weights from qualified unit events. |
| BF-006 | Scheduler reservation protects blocked earlier units by resource name. | A blocked unit can prevent resource-fitting backfill and reduce utilization. | Use non-reserving fit selection plus aging. |
| BF-007 | Browser `browser_exclusive` and stage resources conflate isolation with capacity one. | Semantically independent sessions serialize. | Model affinity/exclusive writes/measurement quietness separately from stack count. |
| BF-008 | Finalizers and child summaries execute as many small, separately planned commands. | Setup/index/render work is repeated and extends the tail. | Derive projections in-process from one event/index model. |
| BF-009 | `agent-finalize` renders generated state twice and hashes broad changing inputs for caches. | The finalizer is far above its reference and cache records rarely hit. | One transactional render plan; use the general cache only for closed actions. |
| BF-010 | The harness contract is concentrated in a giant test file. | Node file concurrency is ineffective and immutable inputs are repeatedly rebuilt. | Partition by owner behavior and create one shared validated index per invocation. |
| BF-011 | Security scanning is repeated by nested or separate aggregates and has no trustworthy result key. | Aggregate tails and the direct target remain tool-bound. | One graph unit plus vulnerability-database-revision cache with exact finding parity. |
| BF-012 | Baseline artifacts allow roster/count/sum drift and dispatch-only timing. | Performance gates can pass or fail for accounting defects instead of execution changes. | Close all derived fields against the owner roster and canonical events. |
| SG-001 | The adopted owner still names superseded owner-slice/accounting schema versions, and `harness.command_surface` has no active test-family row. | The specification and verification routing disagree with the live reviewed projections before the v3 refactor begins. | Repair the current owner references, register an owner-routable command-surface row, and validate the corrected routing before runtime changes. |

## 6. Refactor Workstreams

| ID | Workstream | Class | Depends on | Status | Required outcome |
| --- | --- | --- | --- | --- | --- |
| WF-00 | Source, scope, and authority bootstrap | root | none | `DONE (planning)` | Inspected inventory, clean baseline, non-goals, decisions, and clause disposition are recorded. Revalidate at implementation start. |
| WF-01 | Measurement truth repair | chain | WF-00 | `DONE` | The 47-target roster closes exactly in the v3 writer; internal targets are separate; browser and wrapper work is attributed; Go overhead is corrected; validation closes through the versioned affected-test ledger below. |
| WF-02 | Public contract and tier freeze | chain | WF-01 | `DONE` | Every target/input/ID/schema/artifact/row has a disposition; four monotonic row tiers and five aggregate entry-point policies are machine-projectable. |
| WF-03 | Unified graph and scheduler | chain | WF-02 | `DONE` | Direct and aggregate selectors share graph IDs/digests; scheduling is deterministic, resource-fit, fair, cancellable, and deadlock checked. |
| WF-04 | Service and backend fixture plane | parallel | WF-03 | `DONE` | One broker, explicit leases, pooled replenishment, compatible Go grouping, and owner-specific hotspot fixture repairs are complete. |
| WF-05 | Browser decomposition and stack pooling | chain | WF-03, WF-04 | `DONE` | Browser groups schedule independently; stateful affinity, resets, snapshot writes, measurement quietness, and lane cleanup are explicit. |
| WF-06 | Aggregate composition, deduplication, and cache | chain | WF-03, WF-04, WF-05 | `DONE` | Public roots union/deduplicate units, use actual dependencies, and safely reuse only validated hermetic work. |
| WF-07 | Finalization and tool overhead | chain | WF-06 | `DONE` | Single-pass finalization, partitioned harness contracts, and once-per-graph/freshness-keyed vulnerability scans are complete. |
| WF-08 | Owner-first hard cutover and cleanup | chain | WF-03, WF-04, WF-05, WF-06, WF-07 | `DONE` | Owners and authored inputs changed first; projections regenerated; public targets switched atomically; obsolete paths and schemas were removed. |
| WF-09 | Accumulated validation and handoff | chain | WF-08 | `DONE` | Functional, lifecycle, artifact, generation, and public-entry-point gates pass at one source state; performance publication is explicitly dropped without a performance claim or baseline mutation. |

### WF-01 — Measurement truth repair

Edits started in the observability owner and schema inputs, not in the
checked-in baseline. Completed tasks:

1. Derive the required public measurement roster directly from
   `tools/task_surface_owner.json`; reject missing, extra, duplicate, or internal
   target rows.
2. Add `openapi-compatibility-check` and separate internal-gate accounting.
3. Make the writer recompute and verify target list, command/profile binding,
   row count, medians, and portfolio sum before writing.
4. Emit real work intervals for browser groups/sessions and explicit units for
   all wrapper/setup work.
5. Recalculate Go command and fixture overhead from eligible events; never add
   one aggregate command envelope to every shard.
6. Create a semantic-compatible legacy workload selector for before/after
   executor comparisons. It is private measurement support, not a public alias.
7. Retain the v7 timing windows for unchanged targets, use v8 only for changed
   or previously missing target windows, and preserve every rejected attempt.
8. Adopt TH-HARNESS-REQ-675 so later source versions rerun only their directly
   affected validation closure while compatible passing results carry forward.

Exit: the v3 writer and schema close exactly over the 47-target owner roster,
internal diagnostics are separate, no accepted timing observation is
dispatch-only or relabeled, and the accumulated ledger below resolves every
WF-01 functional obligation. The checked v2 baseline remains archival input;
the v3 reference is published only after the final accepted performance
comparison, not synthesized from incomplete windows.

#### WF-01 completed implementation

- Repaired stale owner-slice, accounting, diagnostic, and schema references in
  the Testing Harness NLSpec and registered the active
  `harness.command_surface` family and verification route.
- Corrected the tracker to four monotonic row tiers across five aggregate entry
  points and retained the original planning baseline separately from the
  implementation start.
- Added v3 performance-evidence and public-baseline schemas and a sole writer
  that derives the exact 47-target roster, bindings, row count, statistics, and
  portfolio sum from the task-surface owner. Internal diagnostics no longer
  enter the public count or total.
- Added exact per-target `source_windows` closure. Qualified target windows may
  come from different frozen source versions, but each row retains its real
  commit and snapshot; copying or relabeling an observation is forbidden.
- Replaced aggregate-envelope timing with canonical target timing, exact
  scheduler-unit start/finish intervals, and non-overlapping setup, fixture,
  execution, collation, and wrapper buckets.
- Replaced Go duration baseline v5 planning with v6 exclusive command-overhead
  attribution, removed aggregate-envelope multiplication across shards, and
  corrected deterministic batching estimates and generated duration data.
- Corrected the CI sequence self-reference and incremented the changed CI
  command ID once. Regenerated task-surface, topology, schedule, and duration
  projections through their Make-owned paths.
- Made generated-source publication preserve repository-owned modes, service
  scope snapshots atomic, owned process-group completion exact, browser port
  leases live for the owned group, and browser port ranges non-overlapping with
  Linux ephemeral allocation.
- Stabilized only directly affected test setup and synchronization boundaries:
  pagination classification, rollback matrix bounds, browser-unit timeout,
  Timeline post-edit response, workbook filter response, keyboard virtualization,
  collaboration queue visibility, recovery and serving-lease success setup,
  browser support request matching, and idempotent absent-bucket cleanup.
  Product assertions, blocked-lease conflict deadlines, and evidence selectors
  remain exact.
- Removed the retained-v1 performance migration reader. Historical artifacts
  remain archival bytes and are not current writer inputs.

#### WF-01 qualified timing ledger

Unchanged timings use v7 as directed. V8 contributes only the changed browser
aggregate and the previously missing `test-fast` window. All rows use one
forced-cold observation, one discarded warm-up, and five measured observations
unless explicitly identified below as functional characterization only.

| Provider or target set | Source | Measured samples (ms) | p50 / p90 / MAD (ms) | Disposition |
| --- | --- | --- | --- | --- |
| `release-check` and 29 exact child targets | v7, `9fd25002`, snapshot `870abee1...` | 313654, 303912, 299881, 315603, 315122 | 313654 / 315603 / 1949 | Qualified and retained; v8 is update-only validation. |
| `agent-finalize` | v7 | 1150, 1080, 1100, 1090, 1090 | 1090 / 1150 / 10 | Qualified. |
| `benchmark-claim-check` | v7 | 110, 120, 110, 110, 110 | 110 / 120 / 0 | Qualified. |
| `frontend-fallow-static` | v7 | 7779, 7720, 7710, 7736, 7824 | 7736 / 7824 / 26 | Qualified. |
| `lint` | v7 | 11334, 11659, 11340, 11275, 11358 | 11340 / 11659 / 18 | Qualified. |
| `migration-drift` | v7 | 3730, 3720, 3770, 3720, 4520 | 3730 / 4520 / 10 | Qualified. |
| `openapi-compatibility-check` | v7 | 170, 170, 170, 180, 170 | 170 / 180 / 0 | Qualified; restores the missing public row. |
| `seaweedfs-release-evidence` | v7 | 460, 470, 470, 470, 530 | 470 / 530 / 0 | Qualified. |
| `standup-operational-recovery-smoke` | v7 | 27480, 27570, 24780, 22590, 22770 | 24780 / 27570 / 2190 | Qualified; MAD is below ten percent. |
| `standup-package-smoke` | v7 | 15470, 15100, 16670, 17160, 17390 | 16670 / 17390 / 720 | Qualified. |
| `test-slice OWNER=module.auth` | v7 | 139814, 143659, 140054, 139462, 145498 | 140054 / 145498 / 592 | Qualified; exact 53-row closure. |
| `service-backed-test-slice OWNER=module.auth` | v7 | 146139, 134619, 141902, 145572, 136282 | 141902 / 146139 / 4237 | Qualified; exact 30-row closure. |
| `test-evidence-audit OWNER=module.auth` | v7 | 2452, 2437, 2412, 2432, 2413 | 2432 / 2452 / 19 | Qualified; all 53 rows and ten target partitions. |
| `browser-e2e` and its exact measurement child | v8, `6fdd0605`, snapshot `a97b7d5a...` | 282184, 282045, 285118, 321151, 297928 | 285118 / 321151 / 12934 | Qualified after cleanup fix; every lane cleaned. |
| `test-fast` | v8 | 195484, 201633, 202229, 202480, 195804 | 201633 / 202480 / 847 | Qualified and previously missing. |
| `test` | v10 cold characterization | 320299 cold | not a warm window | Functional broad pass only; no value is published as a v3 reference sample. |
| `ci` | `wf01-ci-proof-v2` | 207184 characterization | not a warm window | Functional broad pass carried forward; no value is published as a v3 reference sample. |

The v7 release provider supplies 30 exact public observations, twelve v7 direct
windows supply their exact public targets, and the v8 browser and `test-fast`
windows supply three changed or missing observations. The remaining `test` and
`ci` entries have retained functional characterization but no synthetic v3
performance row. WF-09 must measure their final changed implementations before
publishing the complete v3 reference/candidate comparison.

#### WF-01 accumulated validation ledger

This ledger implements TH-HARNESS-REQ-675. Each later version reran only the
rows or targets directly affected by its changes; earlier compatible passes
remain valid for unaffected scopes. Dirty focused proof roots retain their exact
source snapshot and are functional evidence only, never performance samples.

| Version or source snapshot | Directly affected closure | Passing retained evidence | Carried-forward closure |
| --- | --- | --- | --- |
| `b4a5f4e7`, `ef3fba0a...` | Process-group ownership plus broad aggregate behavior | `wf01-process-group-service-slice-proof`, `wf01-browser-e2e-proof`, `wf01-test-fast-proof`, `wf01-test-proof` | Earlier schema, catalog, static, security, and generation passes. |
| `575dc8df`, `9d5996d1...` | CI sequence identity, aggregate timing, command v2 | `wf01-ci-proof-v2` | Broad test/browser closure from `b4a5f4e7`. |
| `c1903e83` plus focused snapshots | Port allocation and keyboard/browser synchronization | `wf01-release-check-proof-v4`, `wf01-keyboard-fill-proof`, `wf01-collaboration-timeout-proof` | Unchanged backend, static, schema, security, and aggregate rows. |
| `60c155ca` through `9fd25002` | Recovery lease setup and browser support request identity | `wf01-target-admission-deadline-proof`, `wf01-support-selector-proof-v2`, v7 five-run release/direct windows | All unaffected v7 target windows. |
| `6fdd0605`, `a97b7d5a...` | Owned object-store cleanup and affected browser/aggregate paths | `wf01-idempotent-bucket-cleanup-proof`, v8 five-run `browser-e2e` and `test-fast` windows, four passing v8 release observations | Qualified v7 timings for every unchanged target. |
| `85b836c2`, `1f89bde4...` | Shared/exclusive serving-lease success setup | `wf01-processlease-success-acquisition-proof`, successful v10 cold `test` root | All unaffected v7/v8 functional and timing entries. |
| Main working snapshot after source-window closure | Performance source-window roster and accumulated-validation owner text | `wf01-windowed-source-roster-proof`; final Markdown/schema/contract gates in the WF-01 handoff | Every unaffected implementation and test result above. |

Rejected evidence remains immutable: v1-v6 measurement cohorts were
superseded by source changes or failed attempts; v7 browser m5 exposed absent
bucket cleanup; v8 release m5 had an unrelated frontend-unit assertion; v8
`test` warm-up exposed the serving-lease setup deadline; v9 cold was
contaminated by concurrent pnpm link mutation; v9 warm-up and v10 warm-up were
interrupted and are not passes. None contributes a timing sample or erases its
recorded failure.

### WF-02 — Public contract and tier freeze

1. Amend the Testing Harness NLSpec with the retained contracts, tier table,
   work-unit shape, resource/fixture capabilities, caching rules, artifact
   model, and hard-cutover policy in this tracker.
2. Produce a complete machine-generated disposition report for all 146 current
   target records and a complete minimum-tier report for every active test row.
3. Inventory every declared Make variable, inherited environment input,
   command ID, schema ID, artifact path, and public caller. Keep only inputs
   that provide current value; removed inputs fail as undeclared after cutover.
4. Remove `release-browser-readiness`: the restart inventory found no direct
   caller outside `release-check` and no distinct human diagnostic role. Project
   its release-support units directly into `release-check`; retain
   `browser-e2e-support` as the distinct release-support work selector.
5. Version all materially changed command and schema IDs once. Do not introduce
   aliases, old-shape readers, or dual output.

Exit: owner review and machine validation prove tier reachability and complete
evidence closure before runtime edits begin.

### WF-03 — Unified graph and scheduler

1. Add an authored graph-composition model and compiler that resolves public,
   leaf, and slice selectors to the canonical work-unit contract.
2. Introduce a private comparison runner and graph-diff command. It MAY coexist
   temporarily with the old runner for characterization, but no public target
   points to it until its selector and artifact parity gates pass.
3. Consolidate scheduler normalization, process execution, failure propagation,
   cancellation, resource accounting, and event emission behind one facade.
4. Replace reservation with critical-path-ranked fit/backfill/aging and add
   simulation fixtures for contention, starvation, failures, and cycles.
5. Capture host capabilities once and validate every unit claim before child
   work.

Exit: legacy-equivalent selectors have identical row/evidence closure; direct
and aggregate selection yields identical unit IDs and semantic digests; no
nested scheduler is required by the comparison runner.

### WF-04 — Service and backend fixture plane

1. Add the run-scoped broker and lease lifecycle with idempotent cleanup,
   borrowed-resource ownership, health/quarantine behavior, and bounded setup
   retry.
2. Make every service-backed catalog row declare a supported lease capability;
   remove implicit template-clone fallback.
3. Characterize and migrate safe clone users to transaction or group leases.
   Keep dedicated/process/migration isolation where committed state,
   cross-connection behavior, recovery, or migration replay requires it.
4. Add asynchronous dedicated-database replenishment and digest-keyed migrated
   templates with contamination tests.
5. Replace fixed symbol/time packing with the capacity-derived compatibility
   algorithm in Section 3.
6. Give extension migrations, recovery process tests, PostgreSQL support, and
   other greater-than-20-second units individual before/after ledgers.

Exit: fixture purity and cleanup pass under contention; clone/reset queue time
falls for the legacy-equivalent workload; row outcomes and failure attribution
remain exact.

### WF-05 — Browser decomposition and stack pooling

1. Project each group, reset, lane readiness, evidence finalizer, and summary
   projection as a graph unit.
2. Replace session-batch execution with single-group runner commands.
3. Implement browser-stack leasing, unique ports and database/object namespaces,
   affinity for stateful chains, quarantine, and owned cleanup.
4. Allow independent default, claimed, accessibility, and read-only visual
   groups to fill available lanes.
5. Enforce quiet measurement admission and exclusive snapshot-output writes.
6. Validate direct and aggregate browser selectors against identical group IDs,
   dependencies, runtime profiles, and evidence outputs.

Exit: the capacity-one generated stage lock and serial batch loop have no
callers; every started lane is cleaned; browser queue/blocking and critical-path
contribution materially improve under Section 8.

### WF-06 — Aggregate composition, deduplication, and cache

1. Project `test-fast`, `check`, `test`, `ci`, and `release-check` from tier and
   policy selectors into one union graph each.
2. Replace whole-child-target dependencies with exact readiness, build, scan,
   service, and evidence edges.
3. Deduplicate same-run units by ID and semantic digest and derive each child
   target summary from the shared event stream.
4. Add the registered cache-closure model, validation, miss/corruption behavior,
   bypass mode, and current-run hit evidence.
5. Ensure noncritical license/SBOM/scanner work cannot starve product critical
   paths through resource rank and claims.

Exit: no phase wrapper or nested Make invocation appears in a public aggregate
plan; identical units execute once; failure and child-summary semantics pass;
all five aggregate roots meet their performance gates.

### WF-07 — Finalization and tool overhead

1. Replace `generate` followed by `generate-drift` in finalization with one
   deterministic scratch render plan. Validate and compare that render; publish
   changed generated outputs atomically only for a refresh action.
2. Remove finalizer action-cache profiles whose input closure is too broad to
   hit; use the general cache model only where a closed action qualifies.
3. Split harness contract tests by semantic owner and failure family. Build one
   immutable validated catalog/task/topology index per invocation and pass it
   to parallel suites.
4. Represent vulnerability scanning as one graph unit. Add the stable database
   revision to its content key, verify cached artifacts and finding equivalence,
   and run normally whenever freshness cannot be proven.
5. Instrument any remaining tool-wrapper envelope as first-class work.

Exit: generation renders once per finalizer invocation; harness contract suites
retain all 104 baseline cases or named replacements; cached and uncached
vulnerability findings match; each named direct target materially improves.

### WF-08 — Owner-first hard cutover and cleanup

1. Land adopted NLSpec changes, then authored schemas/catalog/task/topology/
   resource inputs, then implementation, then Make-generated projections.
2. Run parity using the private comparison selector at the same source state.
3. Atomically point public Make recipes at the canonical runner and new schemas.
4. Delete old sequence/service-backed/test-slice orchestration, browser batch
   scheduling, duplicate finalizers, superseded cache helpers, obsolete inputs,
   and old live schema readers in the same work order.
5. Search authored and generated sources for removed IDs, paths, environment
   variables, and schema versions; regenerate rather than patch projections.

Exit: no public alias, dual write, forwarding wrapper, contextless runner, or
old artifact reader remains; generated drift and boundary checks pass.

### WF-09 — Accumulated validation and handoff

Run the validation ladder in Section 8 at one source state. Record every command,
run root, target semantic digest, system profile, cold/warm/cache posture, and
failure classification. Rebaseline only after functional and performance gates
pass. Complete the trackers, risks, cleanup ledger, and restart record. For this
execution, the user explicitly skipped clean-source multi-window performance
publication. The functional and structural ladder still closes; no v3 baseline
is written and no executor-improvement or non-regression claim is made.

## 7. Implementation Checkpoint Plan

| Checkpoint | Edit scope | Required validation | Expected diff | Rollback point |
| --- | --- | --- | --- | --- |
| CP-00 | Revalidate baseline commit, target/catalog counts, known failures, retained evidence | Read-only owner reports and target explanations | Tracker/evidence update only | Planning baseline |
| CP-01 | Observability owner, schemas, unit timing sources, closure validators | Harness evidence-accounting and command-surface slices; schema checks | Correct roster, browser attribution, wrapper units, Go overhead | CP-00 evidence |
| CP-02 | Capture legacy-equivalent cold/warm window | Performance checker in record-only mode | Retained baseline artifacts only; no semantic changes | CP-01 implementation |
| CP-03 | Adopt owner contract and authored tier/work-unit/resource/fixture schemas | Owner review, JSON shape, catalog and task reports | NLSpec and authored inputs; no public runtime switch | CP-02 baseline |
| CP-04 | Canonical graph compiler and private graph diff | Scheduler/selection fixtures and graph parity | New private modules and tests only | CP-03 owners |
| CP-05 | Unified scheduler and capacity snapshot | Simulation matrix, process failure/cancellation, artifact schema tests | New engine path behind private runner | CP-04 compiler |
| CP-06 | Service broker, fixture leases, Go grouping, hotspot conversions | PostgreSQL/service helper tests and affected owner slices | Test-only lifecycle and catalog changes | CP-05 scheduler |
| CP-07 | Browser group units and stack pool | Browser harness slice, direct affected browser targets, cleanup evidence | New group runner/lifecycle; old batch path still comparison-only | CP-06 broker |
| CP-08 | Aggregate graph, same-run deduplication, content cache | Aggregate graph fixtures, cache invalidation, failure projection | New private aggregate plans | CP-07 browser |
| CP-09 | Finalizer, harness contract, vulnerability work | Direct targets plus schema/artifact comparison | Consolidated tool paths and partitioned tests | CP-08 aggregates |
| CP-10 | Public Make/command/schema hard cutover and regeneration | Task-surface report, generate, drift, policy, JSON shape | Public binding changes and regenerated projections | CP-09 private parity |
| CP-11 | Remove obsolete paths and inputs | Source/ID/path audits, harness contracts, generated drift | Net deletion of phase wrappers, loops, readers, caches | CP-10 cutover commit |
| CP-12 | Focused, aggregate, finalization, and performance validation | Section 8 full ladder | Evidence and tracker updates; baseline refresh only after pass | CP-11 clean implementation |

Each checkpoint MUST be independently reviewable. Product failures, harness
failures, infrastructure failures, and unrelated repository failures are
recorded separately. A failed checkpoint returns to its named rollback state;
it does not update duration references or proceed to public cutover.

Checkpoint status at final handoff: `CP-00` through `CP-12` are `DONE`. CP-12's
clean-source multi-window performance-publication subgate is `DROPPED` by
explicit user direction; this is an accepted scope removal, not passing
performance evidence. Public Make bindings use the canonical graph, scheduler,
broker, cache, and event projections. The comparison runner, phase
orchestrators, serial browser batch loop, duplicate finalizer, superseded
schemas, and live legacy artifact readers are removed. Static rejection
fixtures and the archival v2 duration baseline are the only retained references
to retired IDs. The current 143-record task surface preserves all 98 public
names, and all 1,019 active rows retain complete tier and evidence reachability.

## 8. Validation and Performance Acceptance

### Functional and structural matrix

| Area | Required cases | Acceptance |
| --- | --- | --- |
| Tier selection | Every active row, monotonic tier reachability, omitted owner slice, inactive/unknown/duplicate row, redundancy removal | Every active row has one tier and reaches the complete evidence union exactly as declared. Invalid selection fails before child work. |
| Graph identity | Direct versus aggregate selector, union/deduplication, stable digest, changed command/profile/input | Equivalent selections yield identical units; changed semantics change the digest; one unit executes once. |
| Dependencies | Success, predecessor failure, independent failure, cycle, unknown ID, missing artifact | Descendants skip with reason, permitted independent work follows policy, and invalid graphs fail before execution. |
| Scheduling | Critical-path priority, resource fit, blocked large unit, backfill, aging, simultaneous completion, capacity override | No unheld reservation, deadlock, or starvation; event order and tie breaks are deterministic. |
| Cancellation | Interrupt during setup, test, finalization, and cleanup | Admission stops, owned process groups terminate, leases clean, and partial evidence reports non-success. |
| Fixtures | Transaction rollback, committed-state rejection, group reset, dedicated isolation, pool replenishment, contamination, migration milestone/full replay, borrowed services | No cross-unit state leak; unhealthy resources are destroyed; full-chain migration proofs remain. |
| Browser | Parallel isolated lanes, stateful affinity/order, exact resets, runtime profiles, port/DB/object isolation, measurement quietness, visual read/update, lane failure | Groups share only declared state, conflicting writes never overlap, and every started lane cleans. |
| Caching | Hit, miss, bypass, forced cold, changed source/config/env/tool/dependency/helper/output, corruption, missing artifact, stale/unknown vulnerability DB | Only closed hermetic work reuses; invalid entries run or fail safely; hit/miss evidence is current-run and exact. |
| Failure summaries | Product, configuration, artifact, infrastructure, timeout, cancellation, concurrent failures | Primary failure and all unit outcomes remain deterministic; concise public output and exit status remain meaningful. |
| Artifacts | Event ordering, target projections, critical path, resource holders, cache accounting, roster/count/sum closure, old schema input | All derived math closes against events; old/malformed inputs fail; direct and aggregate summaries agree. |
| Cleanup | Successful, failed, cancelled, quarantined, and partially initialized runs | Owned processes, databases, namespaces, ports, and scratch paths are cleaned exactly once; borrowed resources remain. |

### Characterization and command ladder

The future implementer MUST use `make task-guide ROLE=module-author
OWNER=<owner-id>` before selecting focused rows. Relevant owners include
`harness.command_surface`, `harness.evidence_accounting`,
`harness.generated_artifacts`, `harness.test_catalog`, `harness.browser`, and
`harness.release`.

Run the narrowest applicable ladder at each checkpoint:

1. Affected harness owner slice with `make test-slice OWNER=<owner-id>`.
2. `make harness-contract` for contract, scheduler, artifact, cache, or public
   surface changes.
3. `make json-shape-check`, `make generate-drift`, and
   `make generated-artifact-policy-check` after authored schema or generation
   changes; use `make generate` to refresh outputs.
4. Affected direct backend, browser, finalizer, scanner, or leaf target.
5. `make test-fast`, `make check`, and `make test` after tier/aggregate cutover.
6. `make ci` and `make release-check` at accumulated validation.
7. `make agent-finalize RESULTS_DIR=<successful-full-warm-check-run-root>` after
   a qualifying retained run. If no such run exists, record that retained-run
   maintenance was skipped rather than substituting unrelated evidence.

### Incremental validation protocol

Validation closure follows TH-HARNESS-REQ-675. Every source version records its
changed inputs and deterministic affected row/target closure. Only that closure
is rerun. A passing entry from an earlier version carries forward when its
command, selector, profiles, evidence contract, canonical inputs, toolchain
contract, and producer dependencies are unchanged. A later failure invalidates
only affected entries and remains visible; it can be superseded only by a
relevant source change followed by a complete affected rerun. Same-version
retry cannot erase a failure.

The workstream gate is the union of newest applicable passes, not the status of
the latest broad attempt. The union must close every current required row,
public target, schema/generation gate, security freshness boundary, lifecycle
check, and cleanup obligation exactly once. This does not authorize test-result
caching or skipping selected work within a Make invocation. Performance samples
remain same-source within each target window, while different targets may use
independently qualified windows with exact provenance.

### Performance protocol

There is no fixed global wall-clock target. Performance closure uses comparable
observations and variability-derived materiality.

For each bottlenecked direct or aggregate target:

1. Require one clean frozen source snapshot within each reference or candidate
   window. Reference and candidate snapshots MAY differ only by the reviewed
   refactor. Toolchain, platform, declared inputs, semantic selection digest,
   and capacity profile MUST match across the comparison.
2. Record one cold invocation with reusable work forced off.
3. Discard one warm-up invocation, then record five successful warm invocations.
4. Report p50, p90, median absolute deviation, critical path, resource blocking,
   utilization, setup/fixture work, process count, cache posture, and graph digest.
5. Define `variability_band = 3 * max(reference MAD, candidate MAD, 1ms)`.
   Material warm improvement requires a p50 reduction greater than that band.
   Candidate p90 MUST NOT exceed reference p90 plus the same band.
6. Require the addressed structural metric—critical-path contribution,
   resource-blocking wall, setup interval, process count, or duplicate work—to
   decrease as predicted. A smaller tier inventory alone is not executor proof.
7. Treat a cold regression as unresolved even if content-cache hits improve warm
   use. Cache hits are a separate benefit class.

Material improvement is required for `browser-e2e`, its affected browser leaf
targets, `test-fast`, `test`, `check`, `ci`, `release-check`, `agent-finalize`,
`harness-contract`, and `go-vulncheck`. Every other retained public verification
target MUST be non-regressing within its measured variability band.

If either five-run window has MAD greater than ten percent of its median, or a
leave-one-out evaluation changes the gate result, add one matched observation to
each side. Stop at six measured observations per side. If the conclusion remains
unstable, mark the gate `BLOCKED: unstable measurement environment`; do not
refresh the reference.

### Rebaseline rule

Baseline refresh is the last action after functional, evidence, and performance
acceptance. The writer MUST reject a mixed semantic window or incomplete public
roster. Legacy-equivalent executor results and new tier user-experience results
are reported separately so reduced scope is never presented as scheduler
acceleration.

### Tracker-document verification

After every workstream, update this tracker and append the workstream handoff,
then run:

```sh
make lint-markdown
```

Only after Markdown lint passes may the next workstream begin. Additional
affected gates are selected through the incremental validation protocol above.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Owner | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- | --- |
| T-001 | Define target module and scope | scope | `DONE` | none | planning actor | Section 1 | One execution-control seam and exclusions are explicit. |
| T-002 | Inspect current repo state | discovery | `DONE` | T-001 | planning actor | Sections 1–2 | Owners, projections, representative runtime paths, tests, and diagnostic evidence are inventoried. |
| T-003 | Map owner contracts | contracts | `DONE (planning)` | T-002 | harness owners | Section 4 | Every impacted current contract family has retain/revise/delete posture and successor. Owner adoption remains WF-02. |
| T-004 | Freeze characterization evidence | tests | `DONE` | T-003 | harness evidence accounting | CP-01, CP-02 and the WF-01 ledger | Corrected legacy-equivalent timing, source provenance, and rejected characterization are retained. |
| T-005 | Plan boundary guardrails | architecture | `DONE` | T-003 | harness command surface | Sections 3, 8 | Graph, ownership, generated, cache, artifact, and compatibility guardrails are binary. |
| T-006 | Plan behavior-preserving moves | implementation | `DONE` | T-004, T-005 | workstream owners | Sections 6–7 | Ordered checkpoints, dependencies, rollback points, and hotspot tasks are defined. |
| T-007 | Plan validation loop | validation | `DONE` | T-006 | harness evidence accounting | Section 8 | Focused, aggregate, lifecycle, artifact, and performance gates are named. |
| T-008 | Update docs/contracts if required | docs | `DONE` | T-003 | Testing Harness owner | Testing Harness v3, active guides, generated projections, and WF-02 through WF-09 handoffs | Owner changes preceded implementation; active projections and guides agree after the hard cutover. |
| T-009 | Execute or hand off | handoff | `DONE` | T-006, T-007, T-008 | implementation actor | Section 10 WF-09 handoff | WF-00 through WF-09 are closed, the performance-publication omission is explicit, and review can begin without rediscovery. |

Implementation status values are `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`,
`DEFERRED`, and `DROPPED`. A future actor MUST update both the workstream table
and this top-level tracker at every checkpoint.

## 10. Session Handoff Log

### Planning handoff

| Field | Value |
| --- | --- |
| Date/time | 2026-08-01, America/New_York |
| Branch/commit | `main` at `244f639d19d2e307748b1a8f17b62aa711d5f64d`, ahead of `origin/main` by one commit |
| Dirty state at start | Clean |
| Target module or seam | Harness execution-control plane from public target to evidence publication |
| Current workflow | WF-01 is next; implementation has not started |
| Completed workflows | WF-00 planning bootstrap; T-001, T-002, T-003 planning map, T-005, T-006, T-007, and T-009 planning handoff |
| Changed files | `docs/handoffs/test-harness-bottleneck-refactor-tracker.md` only |
| Commands run | Read-only `git status`, `git rev-parse`, `git log`, `rg`, `sed`, `find`, `wc`, and `jq`; final Markdown verification recorded below |
| Passing validation | `make lint-markdown` passed; run root `.cartulary/test-results/20260802T011556Z-p3059385` |
| Failing validation | Analysis reported an unrelated Go-vet failure and inconsistent performance baseline; revalidate at implementation start |
| Decisions made | Selective hard cutover; tier-by-purpose coverage; unified graph/scheduler; explicit fixture leases; event-derived evidence; no fixed seconds target |
| Open questions | None requiring user choice. Conditional target/cache decisions have mechanical evidence gates in WF-02 and WF-07. |
| Blockers | No planning blocker. WF-01 blocks all performance claims until accounting is corrected. |
| Next recommended workflow | WF-01 measurement truth repair |
| Safe restart command | `make task-guide ROLE=module-author OWNER=harness.evidence_accounting` |

### WF-01 implementation handoff

| Field | Value |
| --- | --- |
| Date/time | 2026-08-02 13:08 EDT |
| Branch/commit | Main worktree based at `9af2e005e75bb636a9114b3ff90cbe5cc8061d76`; clean detached measurement lineage ends at `85b836c2a4c87583839831c49856502643e4da6e` |
| Worktree state and intentional pre-existing changes | Dirty with the intentional WF-01 implementation, owner, test, generated, maintenance-prerequisite, and tracker changes; 69 tracked paths plus four new authored schema/family paths; `docs/domain.md` unchanged |
| Current workflow/checkpoint | WF-01 `DONE`; CP-00 through CP-02 `DONE`; WF-02 / CP-03 is the next workstream |
| Completed workflows/checkpoints | WF-00 planning; WF-01 measurement truth, SG-001 hygiene, v3 roster/timing writer, Go attribution, lifecycle stabilization, qualified timing collection, and accumulated validation adoption |
| Changed authored files | `docs/testing-harness-nlspec.md`; task-surface, topology, scheduler-resource, catalog, test-family, verification-route, schema-attachment, observability, backend-duration, browser-lifecycle, readiness, output, scheduler-test, and harness-contract sources under `tools/**`; focused browser and Go tests; test-only S3/service diagnostics; the CP-00 Go-vet/staticcheck prerequisite used keyed literals and removed dead helpers without changing product contracts |
| Regenerated files and owning generator | `tools/task_surface.generated.mk` and `tools/task_surface_manifest.json` from task-surface owner generation; `tools/scheduler_manifest.json` and `tools/execution_topology_render_index.json` from topology generation; `tools/go_test_duration_baselines.json` from the Go-duration writer; `internal/gen/contractextensions/artifacts_gen.go` from contract generation |
| Commands run | Make-owned `format`, `generate`, `generate-drift`, `lint-markdown`, `json-shape-check`, `harness-contract`, `agent-finalize`, focused owner/slice targets, direct browser/test/test-fast/check/CI/release targets, and the recorded cold/warm performance windows; read-only `explain-run`, owner/task reports, Git, `rg`, `jq`, process, and retained-artifact inspection |
| Passing validation and run roots | Broad carry-forward: `wf01-browser-e2e-proof`, `wf01-test-fast-proof`, `wf01-test-proof`, `wf01-ci-proof-v2`, `wf01-release-check-proof-v4`; focused latest-change proofs: `wf01-keyboard-fill-proof`, `wf01-collaboration-timeout-proof`, `wf01-target-admission-deadline-proof`, `wf01-support-selector-proof-v2`, `wf01-idempotent-bucket-cleanup-proof`, `wf01-processlease-success-acquisition-proof`, and main `wf01-windowed-source-roster-proof`; final affected gates: `wf01-final-json-shape-check`, `wf01-final-harness-contract`, successful command-surface task guide, and `wf01-final-tracker-lint`; qualified v7/v8 windows are listed above |
| Failing validation and classification | Retained and rejected: source-evolution v1-v6 cohorts; v7 browser cleanup failure; unrelated v8 release frontend assertion; v8 serving-lease setup failure fixed in the next version; v9 pnpm-link contamination; interrupted v9/v10 warm-ups. No rejected result is counted as a pass. |
| Performance source/system/semantic digests | v7 source `9fd25002` / snapshot `870abee1...`; v8 source `6fdd0605` / snapshot `a97b7d5a...`; v10 cold source `85b836c2` / snapshot `1f89bde4...`; common host `60b1b04a...`, capacity `0eb1de6a...`, toolchain `877e09fe...`; non-tracker implementation patch digest before final lint `bf96d2ae...` |
| Decisions and deviations | User-directed v7 timing retention; v8 updates only affected/missing surfaces. Per-target frozen windows replace one portfolio-wide source restriction. TH-HARNESS-REQ-675 permits affected-test reruns and compatible pass carry-forward. `test` and `ci` have functional characterization but no synthetic v3 warm timing row; their final changed implementations must be measured before WF-09 publication. |
| Open risks or blockers | No WF-01 blocker. WF-02 must freeze every public target/row disposition before runtime graph work. The checked v2 public baseline remains archival until accepted v3 publication; no compatibility reader was restored. |
| Rollback state | Clean measurement commits preserve each WF-01 source version. Main changes can be reviewed by authored/generated group; historical run roots are immutable. No destructive reset or deletion was used; displaced caches remain recoverable under explicit `/tmp/cartulary-wf01-cache-before-*` paths. |
| Next workflow/checkpoint | WF-02 — public contract and tier freeze / CP-03 |
| Safe restart command | `make task-guide ROLE=module-author OWNER=harness.command_surface` |

The final affected-gate commands were `make json-shape-check`,
`make harness-contract`, and
`make task-guide ROLE=module-author OWNER=harness.command_surface`. Broad
`test`, `check`, `ci`, and `release-check` were not rerun after documentation
and source-window-ledger changes because TH-HARNESS-REQ-675 carries their
unchanged passing roots above; the exact directly affected observability and
schema/contract gates were rerun instead. Retained-run finalizer maintenance was
not rerun because this closure did not select a new successful full warm
`check` root.

### WF-02 implementation handoff

| Field | Value |
| --- | --- |
| Date/time | 2026-08-02 17:47 EDT |
| Branch/commit | `main` at the unchanged WF-02 baseline `f4b0176f877f3edb0fe003ee35e5bb0234c77bea`; implementation remains an uncommitted review unit |
| Worktree state and intentional pre-existing changes | Dirty only with the cumulative WF-02 owner, policy, schema, disposition, contract-test, generated-index, and tracker changes; the worktree was clean at the recorded WF-02 start gate; `docs/domain.md` is unchanged |
| Current workflow/checkpoint | WF-02 and CP-03 `DONE`; WF-03 / CP-04 remains `TODO` until this handoff lints |
| Completed workflows/checkpoints | Adopted Testing Harness v3 execution-control contract; froze all 146 target records, 98 retained public names, 82 declared input uses, 1,019 active rows, four tiers, five entry-point policies, five internal removals, schema successors, and canonical v3 artifact paths |
| Changed authored files | `docs/testing-harness-nlspec.md`; `tools/harness_contract_cutover_policy.json`; 11 successor schemas under `tools/schemas/**`; `tools/harness_schema_attachments.json`; `tools/harness/contract/cutover-disposition.mjs`; `tools/harness/tests/test-harness-contracts.mjs`; this tracker |
| Regenerated files and owning generator | `tools/execution_topology_render_index.json` refreshed only by `make generate`; `tools/harness_contract_cutover_disposition.json` rendered by its checked contract compiler from the task surface, catalog, and cutover policy |
| Commands run | `make task-guide ROLE=module-author OWNER=harness.command_surface`; `node tools/harness/contract/cutover-disposition.mjs --write`; `node tools/harness/contract/cutover-disposition.mjs --check`; `make generate`; `make json-shape-check`; `make harness-command-surface-contract`; `make harness-contract`; `make agent-finalize`; `make lint-markdown`; read-only Git, `rg`, `jq`, `sed`, `wc`, and retained-run inspection |
| Passing validation and run roots | Final generated refresh `.cartulary/test-results/20260802T174725Z-p4046160`; final JSON shape `.cartulary/test-results/20260802T174733Z-p4048373`; command-surface slice `.cartulary/test-results/20260802T174130Z-p4031323`; full harness contract `.cartulary/test-results/20260802T174153Z-p4031816`; finalizer `.cartulary/test-results/20260802T174740Z-p4049466`; tracker lint `.cartulary/test-results/20260802T174805Z-p4056174` |
| Failing validation and classification | Expected fail-closed JSON-shape attempt `.cartulary/test-results/20260802T174052Z-p4027336` reported the newly tracked compiler absent from generated topology provenance; `make generate` refreshed the owning projection and the later JSON-shape gate passed. The failed artifact remains recorded and is not relabeled. |
| Performance source/system/semantic digests | No performance window was selected in WF-02. Task-surface semantic digest `sha256:4924a76e19c9ec8fa9b03e2e6cf9028a31313f7b2ee23ce7f917c6b62c42b825`; catalog semantic digest `sha256:137f27154b7b24fb26e8ff977007e4c4c9d0e1921c0c55ac81af6808a4a21f30`; cache posture was ordinary except finalizer actions, which reported zero hits and no `RESULTS_DIR`. |
| Decisions and deviations | `release-browser-readiness` and the other four phase-shaped internal targets are removal candidates at hard cutover; `browser-e2e-support` remains. Discovery, cleanup, and local-development command IDs retain their versions; 80 materially affected declared IDs are marked for one cutover bump, one removed ID is deleted, and 40 private records have no ID. The fixed warm-check budget and balance inputs are removed; all other declared inputs remain. |
| Open risks or blockers | No WF-02 blocker. Runtime manifests remain on their v2 families by design until the owner-first atomic cutover; the v3 disposition is the migration ledger, not a dual live reader. The checked v2 timing baseline remains archival. |
| Rollback state | CP-02 baseline `f4b0176f877f3edb0fe003ee35e5bb0234c77bea`; deleting the new owner projections and restoring the generated index returns to it without data cleanup. No external or product state changed. |
| Next workflow/checkpoint | WF-03 — unified graph and scheduler / CP-04 and CP-05 |
| Safe restart command | `make task-guide ROLE=module-author OWNER=harness.command_surface` |

### WF-03 implementation handoff

| Field | Value |
| --- | --- |
| Date/time | 2026-08-02 18:08 EDT |
| Branch/commit | `main` at `f4b0176f877f3edb0fe003ee35e5bb0234c77bea`; WF-02 and WF-03 remain one uncommitted review unit |
| Worktree state and intentional pre-existing changes | Dirty with the cumulative WF-02 review unit plus the WF-03 graph owner, schemas, scheduler-owned work-graph facade, tests, generated topology index, and this tracker; no unrelated user changes were present at either start gate |
| Current workflow/checkpoint | WF-03, CP-04, and CP-05 `DONE`; WF-04 / CP-06 remains `TODO` until this handoff lints |
| Completed workflows/checkpoints | Catalog-driven compiler for target, aggregate, owner, and exact-row selectors; canonical deduplication, unit and graph digests; private graph CLI; immutable multidimensional capability snapshot and validated override; critical-path rank, non-reserving fit/backfill, monotonic aging, stable ties, failure propagation, cancellation, event emission, owned process-group termination, and unconditional cleanup |
| Changed authored files | `tools/harness_work_graph_owner.json`; `tools/schemas/cartulary.harness_work_graph_owner.v1.schema.json`; `tools/schemas/cartulary.harness_capacity_override.v1.schema.json`; schema/helper ownership registries; `tools/harness/test-catalog/index.mjs`; scheduler facade under `tools/harness/scheduler/work-graph/**`; graph/scheduler contract fixtures in `tools/harness/tests/test-harness-contracts.mjs`; this tracker |
| Regenerated files and owning generator | `tools/execution_topology_render_index.json` refreshed only through `make generate` after each tracked helper change; no generated schedule or public binding changed |
| Commands run | Private comparison CLI for `backend-unit`; `make generate`; `make json-shape-check`; `make harness-contract`; `make agent-finalize`; read-only artifact, source, task-surface, scheduler-manifest, Git, `rg`, `jq`, and `sed` inspection |
| Passing validation and run roots | Final generation `.cartulary/test-results/20260802T180632Z-p4082125`; final JSON shape `.cartulary/test-results/20260802T180639Z-p4084316`; final harness contract `.cartulary/test-results/20260802T180642Z-p4084900`; finalizer `.cartulary/test-results/20260802T180730Z-p4086334`; the private `backend-unit` graph contained 219 canonical units and digest `sha256:d8c1eb81e915a89b2c96776580f49026565b45bebff002f09bf1d8153771750f` |
| Failing validation and classification | `.cartulary/test-results/20260802T175745Z-p4066173` failed because the initial new top-level `graph` directory violated semantic owner boundaries and one assertion matched the wrong schema-error wording; the code moved under the scheduler owner. `.cartulary/test-results/20260802T180538Z-p4080652` then failed only because the exact helper-facade inventory still expected 34 entries; registration of the new facade raised the reviewed count to 35. Both failed artifacts remain retained. |
| Performance source/system/semantic digests | No performance sample or baseline mutation was selected. Graph semantic digests are source-derived and selector-independent; capability override fixtures record override sources. Finalizer used ordinary cache posture with zero hits and no `RESULTS_DIR`. |
| Decisions and deviations | The graph implementation is a scheduler-owner facade rather than a new top-level harness owner, preserving the repository's cohesion boundary. Current public runners remain bound to v2 while the private compiler proves identity; there is no alias or live legacy reader in the new facade. |
| Open risks or blockers | No WF-03 blocker. The compiler currently translates v2 fixture profiles through the authored graph owner; WF-04 must replace that migration translation with explicit per-row capabilities before public cutover. Aggregate policy-only units and cache closure remain WF-06. |
| Rollback state | CP-03 WF-02 state; removing the scheduler work-graph facade, two WF-03 schemas, graph owner, facade registry row, tests, and generated index delta returns to it. No service, database, browser, or product state changed. |
| Next workflow/checkpoint | WF-04 — service and backend fixture plane / CP-06 |
| Safe restart command | `make task-guide ROLE=module-author OWNER=harness.test_catalog` |

### WF-04 implementation handoff

| Field | Value |
| --- | --- |
| Date/time | 2026-08-02 18:20 EDT |
| Branch/commit | `main` at `f4b0176f877f3edb0fe003ee35e5bb0234c77bea`; cumulative WF-02 through WF-04 changes remain uncommitted |
| Worktree state and intentional pre-existing changes | Dirty only with the cumulative refactor review unit and Make-generated topology index; no product source, dependencies, `docs/domain.md`, or external state changed |
| Current workflow/checkpoint | WF-04 and CP-06 `DONE`; WF-05 / CP-07 remains `TODO` until this handoff lints |
| Completed workflows/checkpoints | Explicit capability ledger for all 1,019 rows; 437 service/runtime-backed capabilities split into 147 browser stacks, eight managed processes, 200 dedicated databases, two groups, 11 migration databases, and 69 transactions; run-scoped lease broker; shared affinity leases; borrowed-resource protection; reverse/idempotent cleanup; quarantine; asynchronous dedicated replenishment; digest-keyed migrated pools; capacity-derived Go LPT planning with twice-lane target and exact compatibility closure |
| Changed authored files | Cutover disposition schema/compiler/report; v3 test-family schema; fixture broker facade under `tools/harness/scheduler/fixture-broker/**`; Go LPT planner and schema; graph compiler/runtime broker integration; helper/schema ownership registries; harness contract fixtures; this tracker |
| Regenerated files and owning generator | `tools/execution_topology_render_index.json` refreshed only through `make generate`; cutover disposition regenerated by its checked compiler after adding per-row capabilities |
| Commands run | Cutover report write/check; `make task-guide ROLE=module-author OWNER=harness.test_catalog`; `make generate`; `make json-shape-check`; `make harness-contract`; `make test-slice OWNER=harness.test_catalog`; `make agent-finalize`; read-only catalog, profile, scheduler, source, and artifact inspection |
| Passing validation and run roots | Generation `.cartulary/test-results/20260802T181804Z-p4099380`; JSON shape `.cartulary/test-results/20260802T181812Z-p4101586`; harness contract `.cartulary/test-results/20260802T181816Z-p4102189`; catalog owner slice `.cartulary/test-results/20260802T181853Z-p4103534`; finalizer `.cartulary/test-results/20260802T181936Z-p4104930` |
| Failing validation and classification | No WF-04 validation failure. Earlier WF-03 failures remain immutable and are not counted in this slice. |
| Performance source/system/semantic digests | No accepted performance window or baseline mutation was selected. The Go LPT plan digest closes lane count, exact compatibility, item IDs, weights, and isolation. Finalizer reported zero cache hits and no `RESULTS_DIR`. |
| Decisions and deviations | Runtime-only rows with no old fixture profile receive `managed_process`, avoiding an implicit `none` fallback. Existing clone-intent rows remain `postgres_dedicated` until owner tests prove transaction/group safety; pressure reduction comes from reusable replenished leases rather than unsafe automated isolation weakening. The old profile-to-capability map was removed from the graph owner after the row ledger became explicit. |
| Open risks or blockers | No WF-04 blocker. The current public v2 runner still owns legacy fixture acquisition until atomic cutover; v3 manifests will carry `fixture_capability` and omit `fixture_profile_id`. Live service pressure and hotspot timing are final-source WF-09 measurements, not inferred from pure lifecycle tests. |
| Rollback state | CP-05 state; remove the fixture broker, Go LPT planner/schema, explicit capability fields, facade rows, tests, and generated index delta. No database or service cleanup is required because all broker lifecycle tests used in-memory providers. |
| Next workflow/checkpoint | WF-05 — browser decomposition and pooling / CP-07 |
| Safe restart command | `make task-guide ROLE=module-author OWNER=harness.browser` |

### WF-05 implementation handoff

| Field | Value |
| --- | --- |
| Date/time | 2026-08-02 18:32 EDT |
| Branch/commit | `main` at `f4b0176f877f3edb0fe003ee35e5bb0234c77bea`; cumulative WF-02 through WF-05 changes remain uncommitted |
| Worktree state and intentional pre-existing changes | Dirty only with the cumulative harness refactor review unit and the Make-generated topology index; no product source, dependencies, `docs/domain.md`, or external browser/service state changed |
| Current workflow/checkpoint | WF-05 and CP-07 `DONE`; WF-06 / CP-08 remains `TODO` until this handoff lints |
| Completed workflows/checkpoints | Browser-stage graph projection for 98 authored semantic group records; independent stateless lifecycle identities; exact stateful affinities and 28 authored reset actions; single-group commands; evidence and target-summary finalizers; scheduler shared/exclusive locks for quiet measurement and snapshot publication; fixture environment injection; run-scoped healthy stack retention, replacement after quarantine, and idempotent final cleanup |
| Changed authored files | Browser graph compiler under `tools/harness/scheduler/work-graph/**`; work-graph schema/model/scheduler/executor; fixture broker; browser manifest adapter facade and single-group attachment validation; harness contract fixtures; this tracker |
| Regenerated files and owning generator | `tools/execution_topology_render_index.json` refreshed only through `make generate`; no public browser projection or generated Make binding changed before the hard cutover |
| Commands run | `make task-guide ROLE=module-author OWNER=harness.browser`; `make harness-contract` twice after one fail-closed repair; `make generate`; `make json-shape-check`; `make test-slice OWNER=harness.browser`; `make agent-finalize`; read-only manifest, lifecycle, import-boundary, artifact, and scheduler inspection |
| Passing validation and run roots | Harness contract `.cartulary/test-results/20260802T182910Z-p4119141` (117/117); generation `.cartulary/test-results/20260802T182958Z-p4120563`; JSON shape `.cartulary/test-results/20260802T183006Z-p4122761`; browser owner slice `.cartulary/test-results/20260802T183012Z-p4123342`; finalizer `.cartulary/test-results/20260802T183050Z-p4125401`; final tracker lint `.cartulary/test-results/20260802T183206Z-p4132437` |
| Failing validation and classification | Initial harness run `.cartulary/test-results/20260802T182819Z-p4117560` failed because the new graph imported the browser owner directly instead of its scheduler adapter and one assertion compared a full runner path as a bare argument. Both were harness-contract defects in the new private path; the retained failure is not relabeled. |
| Performance source/system/semantic digests | No accepted browser performance window or baseline mutation was selected. Graph digests cover exact groups, commands, lifecycle dependencies, affinities, locks, and evidence paths. Finalizer ran without `RESULTS_DIR`; retained-run maintenance was skipped and its ordinary special-action cache reported no accepted performance evidence. |
| Decisions and deviations | The old manifest `browser_session_group` is retained as a compatibility contract, not a physical lease identity. Stateless groups receive unique lease identities while stateful partitions share one affinity key. The public v2 batch path remains comparison-only until the atomic cutover; running its direct target would not execute this private graph, so the focused browser owner slice plus graph/lifecycle contracts are the meaningful CP-07 gate. |
| Open risks or blockers | No WF-05 blocker. A production browser-stack provider and public artifact projection are intentionally bound in WF-08 after WF-06 aggregate composition and WF-07 tool work. Live port/database/object isolation and critical-path improvement require the final-source public browser ladder in WF-09. |
| Rollback state | CP-06 state; remove the browser graph module, lock/affinity fields, browser lease-retention behavior, attachment-contract fallback, tests, and generated index delta. No live stack was started, so no external cleanup is required. |
| Next workflow/checkpoint | WF-06 — aggregate composition, deduplication, and cache / CP-08 |
| Safe restart command | `make task-guide ROLE=module-author OWNER=harness.command_surface` |

### WF-06 implementation handoff

| Field | Value |
| --- | --- |
| Date/time | 2026-08-02 18:52 EDT |
| Branch/commit | `main` at `f4b0176f877f3edb0fe003ee35e5bb0234c77bea`; cumulative WF-02 through WF-06 changes remain uncommitted |
| Worktree state and intentional pre-existing changes | Dirty only with the cumulative harness refactor review unit and Make-generated topology index; no product source, dependency, domain-owner, service, or external state changed |
| Current workflow/checkpoint | WF-06 and CP-08 `DONE`; WF-07 / CP-09 remains `TODO` until this handoff lints |
| Completed workflows/checkpoints | Five aggregate union plans; inherited policy roots with exact producer edges; canonical browser target selection with 50 group units covering all 147 browser rows once; same-run unit deduplication; dependency-closed target projections; one vulnerability scanner unit in `check`, `ci`, and `release-check`; cache registry and runtime for normal/cold/off modes, same-run/content-addressed policies, artifact publication, corruption/missing-output rejection, source/tool/helper mutation, and vulnerability freshness |
| Changed authored files | Work-graph owner/schema and compiler; cache registry, cache record/registry/event schemas, work-graph cache/runtime integration; static shell tier disposition compiler/report/schema; schema attachments; harness aggregate/cache fixtures; this tracker |
| Regenerated files and owning generator | `tools/execution_topology_render_index.json` refreshed only through `make generate`; `tools/harness_contract_cutover_disposition.json` refreshed through its checked compiler after the tier correction; public scheduler and Make projections remain unchanged until hard cutover |
| Commands run | Disposition write/check; private graph compilation for `check` and `release-check`; `make harness-contract` three times across retained repair attempts; `make generate`; `make json-shape-check`; `make test-slice OWNER=harness.command_surface`; `make harness-command-surface-contract`; `make agent-finalize`; read-only task, topology, schema, artifact, and graph inspection |
| Passing validation and run roots | Final harness contract `.cartulary/test-results/20260802T184850Z-p4146417` (120/120); command-surface owner slice `.cartulary/test-results/20260802T184935Z-p4147848`; command-surface contract `.cartulary/test-results/20260802T184941Z-p4148780`; final generation `.cartulary/test-results/20260802T185024Z-p4149971`; JSON shape `.cartulary/test-results/20260802T185031Z-p4152168`; finalizer `.cartulary/test-results/20260802T185035Z-p4152696`; final tracker lint `.cartulary/test-results/20260802T185222Z-p4159921`; final `check` graph has 821 units and digest `sha256:31ed6643b602351659feb0bd6e5410f97e7b7f8c19c0788ce726eb9c422219b2`; final release graph has 1,022 units and digest `sha256:9f34899ec29bad4bd8cda4ff9e7214d16561a138d84ef9a371d9341c2cfd32fd` |
| Failing validation and classification | Harness run `.cartulary/test-results/20260802T184612Z-p4140344` failed because the WF-04 capability assertion treated new browser-group runner IDs as row IDs; the assertion was repaired to close the 147 browser evidence paths. Finalizer `.cartulary/test-results/20260802T184955Z-p4149115` then failed closed because scheduler source changed after the prior generated index; `make generate` refreshed the owner projection and the retained later run passed. Neither failure is relabeled. |
| Performance source/system/semantic digests | No accepted performance window or baseline mutation was selected. Aggregate graph digests close exact rows, policy units, dependencies, cache posture, browser groups, and evidence outputs. Cache fixtures use content-derived input/output digests and current-run events. Finalizer had no `RESULTS_DIR`, so retained-run maintenance was skipped. |
| Decisions and deviations | Seven static shell-contract rows moved from `full_explicit` to `standard_static_contract`; this removes accidental full-tier sibling execution while preserving the standard check policy, yielding 477/315/220/7 tiers and 477/792/1,012/1,012/1,019 entry-point row closures. Leaf Make commands remain private comparison executors in CP-08; phase-target and nested-scheduler commands are absent, and leaf bindings are replaced atomically at CP-10 rather than through wrappers. |
| Open risks or blockers | No WF-06 blocker. Cache entries are not yet enabled on public commands. Vulnerability reuse remains fail-safe without a database revision; WF-07 supplies the direct scanner parity/freshness adapter. Failure-derived target summary rendering and canonical event intervals close in WF-09. |
| Rollback state | CP-07 state; remove aggregate policy declarations/compiler projection, cache registry/runtime/schemas/tests, revert the seven tier dispositions, and refresh the generated index. Cache tests used repo-local temporary directories and removed them. |
| Next workflow/checkpoint | WF-07 — finalization and tool overhead / CP-09 |
| Safe restart command | `make task-guide ROLE=module-author OWNER=harness.generated_artifacts` |

### WF-07 implementation handoff

| Field | Value |
| --- | --- |
| Date/time | 2026-08-02 19:11 EDT |
| Branch/commit | `main` at `f4b0176f877f3edb0fe003ee35e5bb0234c77bea`; cumulative WF-02 through WF-07 changes remain uncommitted |
| Worktree state and intentional pre-existing changes | Dirty only with the cumulative harness refactor review unit and Make-generated projections; no product source, dependency, domain vocabulary, service, or external state changed |
| Current workflow/checkpoint | WF-07 and CP-09 `DONE`; WF-08 / CP-10 and CP-11 remain `TODO` until this handoff lints |
| Completed workflows/checkpoints | One-render generated transaction with compare-once and atomic refresh publication; isolated parallel-safe contract fixtures; five disjoint semantic harness contract suites retaining the 104-case baseline and 17 named additions; freshness-keyed vulnerability resolution and exact artifact/finding/exit parity; once-per-aggregate scanner identity; generated cutover disposition integrated into the Make-owned generation transaction |
| Changed authored files | Generated transaction helper and drift script; generation scratch-input owner and generator; vulnerability adapter; harness contract suite registry/schema and five suite entry files; task-surface owner bindings for contract suites; network-flow shape-check fixture-root handling; harness contracts, schema/helper attachments, and this tracker |
| Regenerated files and owning generator | `tools/harness_contract_cutover_disposition.json` now generated by `make generate`; `tools/task_surface.generated.mk`, `tools/task_surface_manifest.json`, and `tools/execution_topology_render_index.json` refreshed only through `make generate` |
| Commands run | `make task-guide ROLE=module-author OWNER=harness.generated_artifacts`; `make generate`; `make harness-contract`; the three public harness contract entry points; `make json-shape-check`; `make generated-artifact-policy-check`; `make go-vulncheck`; `make lint-scripts`; `make generate-drift`; `make test-slice OWNER=harness.generated_artifacts`; `make agent-finalize`; read-only artifact, task, source, Git, `rg`, `jq`, and retained-log inspection |
| Passing validation and run roots | Generation `.cartulary/test-results/20260802T190938Z-p4190757`; partitioned harness contract `.cartulary/test-results/20260802T190625Z-p4174918` (122/122 including observability); independent contract entries rooted at `.cartulary/test-results/20260802T190714Z-p4176919`, `.cartulary/test-results/20260802T190714Z-p4176917`, and `.cartulary/test-results/20260802T190714Z-p4176906`; JSON shape `.cartulary/test-results/20260802T190805Z-p4179249`; generated policy `.cartulary/test-results/20260802T190805Z-p4179220`; direct scanner `.cartulary/test-results/20260802T190805Z-p4179595`; generate drift `.cartulary/test-results/20260802T190950Z-p4193127`; generated-artifacts owner slice `.cartulary/test-results/20260802T191006Z-p3293`; finalizer `.cartulary/test-results/20260802T191032Z-p7684` |
| Failing validation and classification | Partition run `.cartulary/test-results/20260802T190149Z-p4169647` retained two harness failures: stale cutover disposition and a live-repository network-flow fixture race; both sources changed before the 122/122 pass. Concurrent focused runs `.cartulary/test-results/20260802T190840Z-p4181301` and `.cartulary/test-results/20260802T190840Z-p4181373` retained the same fail-closed scratch-input omission: the cutover policy was absent from the generated-drift scratch declaration. The declaration changed before the later drift and owner-slice passes; no failure was relabeled or overwritten. |
| Performance source/system/semantic digests | No accepted performance window or baseline mutation was selected. Partitioned harness wall time was 15.69 seconds versus the prior retained roughly 30-second private suite run. Direct `go-vulncheck` took 18.87 seconds and executed because no proven reusable database revision existed. Finalizer remained on the v2 comparison binding for CP-10, took 25.08 seconds, used no `RESULTS_DIR`, and accepted no performance evidence. |
| Decisions and deviations | Contract cases remain authored in one case source but are registered through disjoint semantic suite files, giving file-level parallelism without duplicating setup or case bodies. The public finalizer and its special action-cache records intentionally remain comparison-only until the atomic WF-08 binding; the one-render path is private and fully characterized, so no transitional dual writer exists. The cutover disposition is derived state and is now owned by the normal Make generation transaction rather than an out-of-band writer. |
| Open risks or blockers | No WF-07 blocker. CP-10 must switch the public finalizer, scanner, aggregates, browser commands, service-backed commands, owner slices, artifacts, and versioned command/schema families together, then CP-11 must delete the comparison paths and special finalizer caches. Final performance and canonical event interval closure remain WF-09 gates. |
| Rollback state | CP-08 state; remove the transaction/scanner adapters, suite registry and entries, generator integration, isolated fixture-root change, tests, and generated projection deltas. All scratch, cache, and fixture tests used repo-local recoverable paths and completed cleanup. |
| Next workflow/checkpoint | WF-08 — owner-first hard cutover and cleanup / CP-10 and CP-11 |
| Safe restart command | `make task-guide ROLE=module-author OWNER=harness.command_surface` |

### WF-08 implementation handoff

| Field | Value |
| --- | --- |
| Date/time | 2026-08-02 20:48 EDT |
| Branch/commit | `main` at `f4b0176f877f3edb0fe003ee35e5bb0234c77bea`, equal to `origin/main`; cumulative WF-02 through WF-08 changes remain uncommitted |
| Worktree state and intentional pre-existing changes | Dirty only with the cumulative harness refactor review unit, owner-first specification and guide changes, deleted obsolete harness paths, focused test-support repairs, and Make-generated projections; no product behavior, production API, dependency, domain vocabulary, or external service state changed |
| Current workflow/checkpoint | WF-08, CP-10, and CP-11 `DONE`; WF-09 / CP-12 is `IN_PROGRESS` after the handoff lint gate |
| Completed workflows/checkpoints | Atomic public graph cutover for all 98 public Make names; 47 canonical measured-target bindings; v3 test families, v2 task owner, v6 topology, v5 resources, v8 browser model, and v3 scheduler projection; canonical capacity, cache, fixture lease, manifest, unit-event, run-summary, and target-summary contracts; explicit fixture capability environment; group-level browser scheduling; aggregate union/deduplication; canonical artifact assertions; single-render finalizer; nested OTel owner-slice execution replaced by exact graph dependencies; old orchestration, batch, duration-accounting, schema, finalizer-cache, and compatibility-reader paths removed |
| Current inventory | 143 task-surface records: 98 public, ten check-internal, and 35 internal-helper records; all 47 measured public targets use scheduler orchestration and canonical evidence; exactly 1,019 active catalog rows |
| Changed authored files | Testing Harness v3 NLSpec; task-surface, catalog/test-family, topology, browser, scheduler-resource, work-graph, cache, helper-ownership, schema-attachment, finalization, observability, and extension-contract owners; canonical compiler/scheduler/executor/broker/cache/evidence modules; PostgreSQL and process-lifecycle test support; active developer/testing guides; contract, generated-artifact, readiness, release-surface, browser, fixture, cache, observability, and real-target smoke tests; this tracker |
| Regenerated files and owning generator | `tools/task_surface.generated.mk`, `tools/task_surface_manifest.json`, `tools/scheduler_manifest.json`, `tools/execution_topology_render_index.json`, generated topology schedules, and `internal/gen/contractextensions/artifacts_gen.go` refreshed only through `make generate` and their registered generators; generated roots were not hand-edited |
| Commands run | `make task-guide ROLE=module-author OWNER=harness.command_surface`; repeated focused `make harness-contract`, task-surface, command-surface, JSON-shape, generation, generated-policy, drift, import-boundary, shell/script lint, backend unit/integration, exact owner-row, OTel conformance, fixture lifecycle, and smoke gates; direct cache-off graph probes; `make check`; `make agent-finalize`; read-only Git, source/ID/path/environment, task/catalog count, run-manifest, and retained-artifact inspections |
| Passing validation and run roots | Accumulated harness smoke `.cartulary/test-results/wf08-smoke-full-final-pass`; backend unit `.cartulary/test-results/wf08-backend-unit-off` (52/52); backend integration `.cartulary/test-results/wf08-backend-integration-off2` (181/181); transaction fixture `.cartulary/test-results/wf08-fixture-transaction` (3/3 with one released borrowed transaction lease); migration audit `.cartulary/test-results/wf08-platform-audit-migration-row`; direct OTel `.cartulary/test-results/wf08-otel-no-nested` (10/10); aggregate `check` `.cartulary/test-results/wf08-check-no-nested` (688/688); final source shell lint `.cartulary/test-results/20260803T004557Z-p1873861`; generation `.cartulary/test-results/20260803T004612Z-p1874721`; generated policy `.cartulary/test-results/20260803T004628Z-p1877032`; drift `.cartulary/test-results/20260803T004628Z-p1877034`; JSON shape `.cartulary/test-results/20260803T004628Z-p1877066`; finalizer `.cartulary/test-results/20260803T004654Z-p1880854`; final WF-08 tracker lint `.cartulary/test-results/20260803T005038Z-p1884498` |
| Failing validation and classification | Generator defects remained visible at `.cartulary/test-results/20260802T225028Z-p572358`, `.cartulary/test-results/20260802T225114Z-p577923`, `.cartulary/test-results/20260802T225200Z-p586941`, `.cartulary/test-results/20260802T231025Z-p778421`, `.cartulary/test-results/20260802T232447Z-p783468`, `wf08-generate-drift`, and `wf08-generate-drift-2`; their generator-version, cache-path, lifecycle/evidence, pending-baseline, JavaScript-syntax, and stale-projection sources changed before the final generation/drift passes. `wf08-check-repro` was intentionally cancelled. `wf08-smoke-full-final`, `wf08-smoke-full-fix`, `wf08-smoke-extended-fix` through `wf08-smoke-extended-fix4`, `wf08-smoke-full-pass`, and `wf08-smoke-full-pass2` retained successive stale-fixture, canonical-artifact, lifecycle, PostgreSQL-capability, release-surface, and hidden nested-OTel failures; affected sources changed before `wf08-smoke-extended-fix5` and the final accumulated smoke passed. `.cartulary/test-results/20260803T004628Z-p1877120` failed closed because a nonexistent `RESULTS_DIR` was supplied to `agent-finalize`; the valid no-`RESULTS_DIR` invocation is separately recorded and does not erase that configuration failure. |
| Performance source/system/semantic digests | No accepted performance window or baseline mutation was selected. Final direct `check` evidence at `wf08-check-no-nested` used source `sha256:b1ed9a70f7670d4f7fb8f24e6bc4769de10d5e19e89af6771cc91e4ad77d9a9e`, system `sha256:24000a47102254c54328907385158903bc14cd0446e69c3a277895d52b6a3be2`, graph `sha256:03e78caefff3408127ecd8ed23d1f7f8de4705379e271bdd9b9b3ab351c5a4a7`, normal cache mode, and 62,990 ms. Finalizer source `sha256:7c417f1767e35a4005991be28260d6a89094926b325a38eb913b871170f4f9cf`, same system digest, graph `sha256:58e60baf7949d988196a6dce93337b27b8a9830d469b47c46d2e61f898adf457`, and normal cache mode. The checked v2 duration baseline remains archival and unchanged. |
| Decisions and deviations | The public vocabulary remains 98 names, but obsolete internal phase records are removed, reducing the full task-surface inventory from 146 to 143. Policy targets may declare exact owner-slice graph dependencies; this replaced the hidden OTel nested scheduler without reintroducing target-shaped phase barriers. Migration-capable Go rows explicitly request migration leases while current-head isolated database helpers map that capability to a dedicated migrated template; no implicit clone fallback was restored. Static negative fixtures retain retired identifiers solely to prove rejection. |
| Source/ID/path audit | No live consumer of the five removed internal phase targets, retired batch/scheduler commands, or superseded schema IDs remains. Search hits are limited to rejection tests, human prose, one unrelated frontend fixture label containing `test-local`, and the archival v2 duration baseline. `git diff --check` passes. |
| Open risks or blockers | No WF-08 blocker. WF-09 must close the final functional ladder, canonical interval/projection ledger, cold/warm performance protocol, non-regression roster, accepted v3 baseline publication, final retained-run maintenance, and restartable handoff. Performance improvement is not inferred from the passing functional roots. |
| Rollback state | CP-09 private-parity state. Rollback requires restoring the complete old binding/schema/orchestrator set as one unit; partial aliases, dual writers, and mixed readers are prohibited. No rollback or external cleanup is currently required. |
| Next workflow/checkpoint | WF-09 — accumulated validation and handoff / CP-12 is now open |
| Safe restart command | `make task-guide ROLE=module-author OWNER=harness.command_surface` |

### WF-09 implementation handoff

| Field | Value |
| --- | --- |
| Date/time | 2026-08-02 23:50 EDT |
| Branch/commit | `main` at `f4b0176f877f3edb0fe003ee35e5bb0234c77bea`, equal to `origin/main`; cumulative WF-02 through WF-09 changes remain uncommitted for review |
| Worktree state and intentional pre-existing changes | Dirty only with the cumulative harness-refactor review unit: adopted owner and active-guide changes, authored machine contracts and schemas, canonical graph/scheduler/broker/cache/evidence implementation, obsolete-path deletions, focused test-support repairs, and Make-generated projections. No product API, production behavior, dependency version, domain vocabulary, or external service state was changed. The unrelated retained WF-01 SeaweedFS listener was inspected but not stopped or modified. |
| Current workflow/checkpoint | WF-09 and CP-12 `DONE`; the clean-source multi-window performance-publication subgate is `DROPPED` by explicit user direction, so no performance claim or v3 reference publication is part of completion |
| Completed workflows/checkpoints | WF-00 through WF-09 and CP-00 through CP-12. Functional acceptance closes tier selection, graph identity, deterministic scheduling, explicit fixture leases, browser affinity and cleanup, aggregate deduplication, cache/security fail-safe behavior, canonical evidence, single-pass finalization, once-per-graph scanning, hard-cutover cleanup, and the restartable handoff. |
| Current inventory | 143 task-surface records: 98 public, ten check-internal, and 35 internal-helper records; 47 canonical measured-target bindings; 61 catalog owners and exactly 1,019 active rows |
| Changed authored files | Testing Harness v3 NLSpec; task-surface, test-family/catalog, topology, browser, scheduler-resource, work-graph, cache, helper-ownership, schema-attachment, finalization, observability, release-evidence, and extension-contract owners and implementation; PostgreSQL, browser, process, row-runner, generated-transaction, canonical-performance, and contract-suite tests; the two active harness guides; this tracker. `docs/domain.md` remains unchanged. |
| Regenerated files and owning generator | `tools/task_surface.generated.mk`, `tools/task_surface.runtime.generated.mk`, `tools/task_surface_manifest.json`, `tools/scheduler_manifest.json`, `tools/execution_topology_render_index.json`, generated topology/browser projections, and `internal/gen/contractextensions/artifacts_gen.go` were refreshed only through `make generate` and registered Make-owned generators; generated roots were not hand-edited. |
| Commands run | Focused owner guidance/slices; repeated `make harness-contract`; direct browser, backend, scanner, finalizer, object-store compatibility, release-evidence, generated-artifact, JSON-shape, import-boundary, task-surface, and baseline-coverage gates; `make generate`; `make test-fast`; `make check`; `make test`; `make ci`; `make release-check`; `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260803T032316Z-p90370`; read-only inventory, artifact, digest, source/ID/path/environment, and Git audits. `make harness-performance-check EVIDENCE_ROOTS_FILE=.cartulary/t012-performance-comparison-roots.v2.json` was also run as an expected negative old-schema probe. |
| Passing validation and run roots | Focused catalog slice `.cartulary/test-results/20260803T020115Z-p2468924`; harness contract `.cartulary/test-results/20260803T032241Z-p89490`; object-store compatibility `.cartulary/test-results/20260803T022708Z-p3011918`; same-run release gate `.cartulary/test-results/20260803T023756Z-p3203785`; browser support `.cartulary/test-results/20260803T025332Z-p3557055`; direct browser `.cartulary/test-results/20260803T031816Z-p51014`; direct scanner `.cartulary/test-results/20260803T031750Z-p50225`; JSON shape `.cartulary/test-results/20260803T031625Z-p4172271`; drift `.cartulary/test-results/20260803T031627Z-p4172659`; generated policy `.cartulary/test-results/20260803T031634Z-p4175191`; generation `.cartulary/test-results/20260803T032234Z-p87265`; frontend import boundary `.cartulary/test-results/20260803T032812Z-p167048`; baseline closure `.cartulary/test-results/20260803T032736Z-p165793`; final-source `test-fast` `.cartulary/test-results/20260803T032822Z-p167493` (339/339), `check` `.cartulary/test-results/20260803T032316Z-p90370` (688/688), `test` `.cartulary/test-results/20260803T032842Z-p168009` (822/822), `ci` `.cartulary/test-results/20260803T033328Z-p252871` (850/850), `release-check` `.cartulary/test-results/20260803T033833Z-p401790` (857/857), and retained-run finalizer `.cartulary/test-results/20260803T032431Z-p162280` (1/1). Tracker Markdown gates `.cartulary/test-results/wf09-tracker-lint-draft`, `.cartulary/test-results/wf09-final-markdown-lint`, and `.cartulary/test-results/wf09-final-markdown-lint-recorded` pass in sequence. All canonical timing summaries report zero or rounding-only one/two-millisecond unattributed wall time. |
| Failing validation and classification | `.cartulary/test-results/20260803T013420Z-p2143055` and `.cartulary/test-results/20260803T014225Z-p2303007` retained fail-closed profile/fixture contract defects; their authored inputs changed before later contract passes. `.cartulary/test-results/20260803T014322Z-p2306357` is the intentional failure-taxonomy probe. `.cartulary/test-results/20260803T014431Z-p2307505` exposed zero-reference browser stacks retaining PostgreSQL clients after claims released; broker ownership, terminal release edges, and resource claims changed before the final 822/822 tests. `.cartulary/test-results/20260803T021506Z-p2825424` exposed release coupling to a fixed local-development object-store port; the release graph moved to an owned run-scoped lease without modifying the unrelated listener. `.cartulary/test-results/20260803T022455Z-p3007307` exposed missing CORS behavior at the raw object-store endpoint; the owned ephemeral proxy changed before compatibility passed. `.cartulary/test-results/20260803T022749Z-p3013570`, `.cartulary/test-results/20260803T023457Z-p3192981`, and `.cartulary/test-results/20260803T023648Z-p3198495` exposed stale/pre-final-projection release-evidence reads; live atomic unit results changed before the gate and release aggregate passed. `.cartulary/test-results/20260803T024557Z-p3451532` exposed a browser request/response-generation race; exact request identity and completion synchronization changed before direct and aggregate passes. `.cartulary/test-results/20260803T032051Z-p84178` exposed missing `RESULTS_DIR` child forwarding; the declared input contract changed before finalizer and all final aggregates passed. The archival v2 performance manifest was deliberately rejected by the v3 schema; it was not translated, relabeled, or used as current evidence. |
| Performance source/system/semantic digests | Final functional roots share dirty reviewed source digest `sha256:6df8593b8b23f64a71d2521f0118ba747fbd31a8f13753308a2335b94b183068`, system digest `sha256:a0ed26bbd5d2dbd5c14ad15cb957bff29b110da73fbca87edbe02750cc6a1a1a`, and normal cache mode. Graph digests are `test-fast` `sha256:f02a279e84df40d151279d2bcb309f1953a4f79c0d6ac49fb767ed7d544f7607`, `check` `sha256:03e78caefff3408127ecd8ed23d1f7f8de4705379e271bdd9b9b3ab351c5a4a7`, `test` `sha256:7524412c2351e7019cc59bfa1e975e8076c3e480310602f42f1adad9d20a230c`, `ci` `sha256:796c711e99c100cfe03c17688421986024019a5acaf30df07e3566c99bd149f2`, `release-check` `sha256:e9ea3b5a1f99fcc49b43163b9ff4632f57787ca24d95f507071d0936a8ef0307`, and `agent-finalize` `sha256:58e60baf7949d988196a6dce93337b27b8a9830d469b47c46d2e61f898adf457`. These are functional observations, not accepted performance windows. The user explicitly skipped clean-source multi-window publication; the v2 baseline remains archival and no v3 baseline exists. |
| Decisions and deviations | Clean-source performance sampling/publication was removed from completion by explicit user direction on 2026-08-02. The implementation retains the v3 fail-closed schemas, exact 47-target writer, cold/warm protocol, and rejection of old inputs for future use, but this handoff makes no material-improvement or non-regression claim. The public 98-name vocabulary remains; retired internal phases, live compatibility readers, dual writers, and forwarding aliases remain deleted. |
| Source/ID/path audit | Live authored/generated sources contain no consumer of the five removed internal phase targets, retired batch/scheduler commands, superseded live schema IDs, or old artifact readers. Remaining hits are static rejection fixtures, explicit negative assertions, one unrelated frontend label containing `test-local`, and the archival v2 duration baseline. `make task-surface-report` closes 98/10/35, and `git diff --check` passes. |
| Open risks or blockers | No implementation blocker. Accepted residual risk: executor improvement and public-target non-regression were not statistically established because publication was explicitly skipped. No future baseline may treat these functional observations as accepted performance evidence; a later performance effort must start with clean committed reference/candidate windows. |
| Rollback state | Final hard-cutover review unit at CP-12. Rollback must restore the pre-cutover owner/schema/orchestrator set atomically from CP-09; partial aliases, dual schemas, dual writers, or mixed readers remain prohibited. No owned service/process/port/database cleanup remains, and borrowed/external resources were preserved. |
| Next workflow/checkpoint | None in this plan. Review and commit/publish are intentionally left to the user; a future optional performance-only effort starts from a reviewed clean commit. |
| Safe restart command | `git diff --check && make task-surface-report && make harness-contract` |

Do not claim a command passed unless it ran in that session or an exact retained
artifact is named. Do not claim parity, preservation, cleanup, or improvement
without the corresponding artifact. Missing evidence is `TODO`, not inference.

## 11. Risks and Blockers

| ID | Risk or blocker | Severity | Preventive control | Resolution condition |
| --- | --- | --- | --- | --- |
| RB-001 | Corrected telemetry changes the apparent bottleneck ordering. | High | Complete WF-01 before architecture performance claims; retain raw prior evidence. | Qualified unit-derived baseline is complete. |
| RB-002 | Tier reclassification silently removes product evidence. | High | Minimum-tier closure and full/release union audit for every active row. | All rows are reachable or have owner-backed redundancy records. |
| RB-003 | Increased fixture concurrency causes state contamination or Postgres saturation. | High | Explicit leases, multi-dimensional capacity, contamination tests, quarantine, cold/warm pressure evidence. | Purity tests and resource-pressure gates pass at selected capacity. |
| RB-004 | Browser stack pooling creates port, database, object, or process leaks. | High | Broker-owned lane identity, unique namespaces, process groups, cancellation and leak tests. | All success/failure/cancellation lifecycle fixtures clean. |
| RB-005 | Critical-path scheduling starves small or scarce-resource work. | Medium | Resource-fit backfill plus monotonic aging and deterministic simulations. | Starvation/contention matrix passes. |
| RB-006 | Cache closure misses an input or accepts stale security evidence. | High | Registered closures, output digest validation, vulnerability DB revision, forced-cold tests, fail-safe miss. | Mutation fixtures invalidate every affected profile; cached/uncached findings match. |
| RB-007 | Hard-cutover artifacts break unknown consumers. | Medium | Complete caller and declared-input inventory in WF-02; one atomic public switch. | All registered consumers migrate; no old reader or alias remains. |
| RB-008 | Reduced workload is misreported as execution acceleration. | High | Preserve private legacy-equivalent measurement selector and separate tier/user timing. | Executor and inventory effects are reported separately. |
| RB-009 | External scanner or host variance prevents stable improvement proof. | Medium | Matched system digests, extended paired windows, freshness-keyed reuse, explicit unstable-environment blocker. | Stable comparison passes or work remains blocked without rebaseline. |
| RB-010 | Unrelated repository failure obscures final validation. | Medium | Reproduce and classify at CP-00; record exact failing target and source state. | Failure is fixed separately or retained as explicit unrelated evidence. |
| RB-011 | Temporary comparison runner becomes permanent duplicate architecture. | Medium | Private-only binding and mandatory deletion in CP-11. | No comparison runner is reachable after cutover except retained static fixtures. |
| RB-012 | Incremental carry-forward hides a changed test or erases a historical failure. | High | TH-HARNESS-REQ-675 deterministic impact closure, exact source provenance, same-version retry prohibition, immutable failures, and exact combined roster closure. | AC-081 fixtures and the final ledger prove every current requirement resolves to the newest applicable pass exactly once. |

Final disposition:

| Risk | Final status | Evidence or accepted posture |
| --- | --- | --- |
| RB-001 | `ACCEPTED RESIDUAL` | Canonical unit-derived accounting and the 47-target writer pass, but the user explicitly skipped the clean-source publication window. No performance conclusion or reference mutation was made. |
| RB-002 | `RESOLVED` | All 1,019 active rows have one tier and the focused/aggregate catalog, tier, and evidence-union gates pass. |
| RB-003 | `RESOLVED` | Explicit fixture claims, broker lifecycle tests, final full aggregates, and released lease evidence pass at the selected capacity. |
| RB-004 | `RESOLVED` | Browser affinity, reset, isolation, failure, cancellation, quarantine, and final direct/aggregate cleanup gates pass. |
| RB-005 | `RESOLVED` | Rank, fit, backfill, aging, contention, simultaneous-completion, and starvation simulations pass. |
| RB-006 | `RESOLVED` | Registered closure mutation tests, corrupt/missing output tests, cold/off modes, and vulnerability-freshness fail-safe behavior pass. |
| RB-007 | `RESOLVED` | All registered consumers migrated atomically; source/ID/path audits find no live compatibility reader or forwarding alias. |
| RB-008 | `ACCEPTED RESIDUAL` | Tier and executor effects remain structurally distinct, and this handoff makes no acceleration claim because performance publication was skipped. |
| RB-009 | `DROPPED` | The user removed multi-window performance publication from this effort. The v2 baseline remains archival, and any future performance claim requires new clean matched windows. |
| RB-010 | `RESOLVED` | Every final-source focused and aggregate gate passes; all earlier failures retain their classifications and superseding source changes in the WF-09 handoff. |
| RB-011 | `RESOLVED` | The comparison runner and obsolete scheduler families are absent from the live surface; only static rejection fixtures remain. |
| RB-012 | `RESOLVED` | The accumulated ledger retains failed roots, relevant source changes, newest applicable passes, and final aggregate closure without relabeling retries. |

## 12. Binary Completion Criteria

### Tracker acceptance

| ID | Criterion |
| --- | --- |
| RF-AC-001 | Exactly one primary seam—the harness execution-control plane—is named. |
| RF-AC-002 | Inspected sources, representative implementations, generated boundaries, evidence limits, and unseen-adjacent-file rule are recorded. |
| RF-AC-003 | Every affected scheduling, lifecycle, telemetry, artifact, and compatibility contract family has a retain/revise/delete disposition. |
| RF-AC-004 | Product behavior preservation is separated from authorized harness behavior changes and tier redistribution. |
| RF-AC-005 | Characterization, graph, lifecycle, cache, artifact, and performance evidence requirements are explicit. |
| RF-AC-006 | The checkpoint sequence includes validation, expected diff, dependency, and rollback state for every risky move. |
| RF-AC-007 | Authored owner, generated-artifact, runner-adapter, fixture-provider, and public-facade boundaries are explicit. |
| RF-AC-008 | No generated file hand edit is planned. |
| RF-AC-009 | No phase-shaped scheduler or runtime dependency is introduced. |
| RF-AC-010 | The handoff permits WF-01 to begin without rediscovery or an unresolved user decision. |

### Future implementation definition of done

| ID | Criterion |
| --- | --- |
| THBR-AC-001 | The adopted Testing Harness owner describes the new tiers, graph, scheduler, resource/fixture, cache, artifact, and hard-cutover contracts before their implementation projection. |
| THBR-AC-002 | The authored 47-target roster, observed targets, baseline rows, counts, command/profile bindings, and recomputed totals are exactly equal; internal gates are separate. |
| THBR-AC-003 | Every active catalog row has one minimum tier and complete owner/full/release reachability, or an owner-backed redundancy removal record. |
| THBR-AC-004 | Direct, leaf, slice, and aggregate selectors compile through one graph compiler and one scheduler; equivalent selections have identical unit IDs and digests. |
| THBR-AC-005 | Scheduling is deterministic, critical-path ranked, resource fitting, backfilling, aging, deadlock checked, and cancellation safe. |
| THBR-AC-006 | Host capacity is captured once, multi-dimensional, recorded, overrideable only through declared inputs, and never inferred solely from CPU for scarce services. |
| THBR-AC-007 | Every service-backed row has an explicit fixture lease; transaction, group, dedicated, migration, object-store, process, and browser isolation pass contamination and cleanup tests. |
| THBR-AC-008 | Browser groups are independently schedulable; stateful affinity/resets, quiet measurement, snapshot writes, lane quarantine, and cleanup pass; the serial stage lock and batch loop are gone. |
| THBR-AC-009 | Public aggregates contain no whole-target phase barrier or nested Make/scheduler work and execute each identical unit once. |
| THBR-AC-010 | Cache reuse is limited to validated closures, reports current-run hit evidence, fails safe on any mismatch, and cannot satisfy cold or non-hermetic gates. |
| THBR-AC-011 | Finalization renders generated structure once; harness contracts use parallel semantic suites; vulnerability scans execute once per graph and cached findings equal uncached findings. |
| THBR-AC-012 | Canonical unit events account for all material wall intervals and derive run, target, hotspot, pressure, critical-path, and baseline artifacts without dispatch-only or duplicated timing. |
| THBR-AC-013 | Public command/schema changes use one versioned hard cutover; no alias, forwarding wrapper, dual writer, old live reader, obsolete input, or ineffective special cache remains. |
| THBR-AC-014 | Functional, artifact, generated, focused owner, direct target, aggregate, finalizer, CI, and release validation closes through the exact TH-HARNESS-REQ-675 affected-test ledger, with every source version, impact set, pass, failure, supersession, and run root recorded. |
| THBR-AC-015 | Every named bottleneck target materially improves under the variability-derived protocol; every other retained public verification target is non-regressing; baselines refresh only afterward. |
| THBR-AC-016 | The final tracker records changed/authored/generated files, commands, run roots, semantic/system digests, failures, cleanup, decisions, residual risks, and safe restart state. |

Final criterion disposition: THBR-AC-001 through THBR-AC-014 and THBR-AC-016
are satisfied. THBR-AC-015 is `DROPPED` by explicit user direction: no
clean-source multi-window performance publication was run, no v3 reference was
written, and no material-improvement or non-regression assertion is made.

### Final checklist

- [x] WF-01 corrected measurement truth and its accumulated validation ledger are retained before optimization.
- [x] Every active row and public input has a target-state disposition.
- [x] Owner changes precede schemas, authored projections, generation, and code.
- [x] Direct and aggregate work use one graph and scheduler.
- [x] Fixture and browser concurrency preserves explicit isolation.
- [x] No phase wrapper, serial browser loop, duplicate finalizer, or nested runner remains.
- [x] Cache closure and security freshness fail safe.
- [x] Artifact roster, timing, critical path, and totals close exactly.
- [x] Executor improvement is separated from tier inventory reduction; no improvement claim is made because performance publication was skipped.
- [x] Focused and accumulated validation close through exact affected-test provenance.
- [x] Generated output was refreshed only through Make.
- [x] Obsolete compatibility paths and schema readers are removed.
- [x] Performance references were not refreshed; publication was explicitly dropped before any acceptance claim.
- [x] The final implementation handoff is complete and restartable.
