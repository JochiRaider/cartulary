# Test Harness Bottleneck Refactor Tracker and Handoff

## 1. Scope and Source Posture

| Item | Current posture |
| --- | --- |
| Primary seam | Public test target to workload selection, scheduling, resource and fixture lifecycle, execution, and evidence publication |
| Output path | `docs/handoffs/test-harness-bottleneck-refactor-tracker.md` |
| Mode | Planning-only analysis and restartable implementation handoff |
| Allowed change in this phase | This tracker only |
| Future implementation scope | Authored harness owner documents and machine inputs, harness tests and runtime helpers, test-only fixture support, schemas, and generated projections refreshed through Make |
| Non-goals | No product behavior, production API, assertion weakening, dependency upgrade, generated-file hand edit, or performance claim based only on a smaller workload |
| Planning baseline | `244f639d19d2e307748b1a8f17b62aa711d5f64d` on `main`, one commit ahead of `origin/main`, with a clean worktree before this tracker was added |
| Analysis posture | `temp/analysis-notes.md` is diagnostic evidence. Its inconsistent checked-in baseline and single-sample observations are not publishable performance proof. |
| Compatibility decision | Selective hard cutover: retain useful public Make vocabulary and roles; permit coordinated command, schema, artifact, lifecycle, topology, and redundant-target changes without compatibility aliases |
| Coverage decision | Tier by purpose: fast feedback may be narrower, while full, CI, and release entry points preserve the complete correctness and release-evidence union |

This tracker applies the structure of
`docs/handoffs/cartulary_modular_refactor_planning_framework.md` to one primary
harness seam. It does not authorize implementation. A future implementation
session MUST update this tracker as its controlling restart record.

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

## 2. Current-State Repository Inventory

### Execution-control inventory

| Area | Current owner or implementation | Current responsibility | Principal finding | Target disposition |
| --- | --- | --- | --- | --- |
| Public task surface | `tools/task_surface_owner.json` | 145 target records, command IDs, inputs, output policies, inclusion sets, observability policy | The performance roster requires 47 targets, while the checked-in baseline does not close over it. | Retain as the authored public owner; revise target roles, inputs, IDs, and observability projection. |
| Test selection | `tools/test_catalog_owner.json`, `tools/test_families/**` | Owner routing, runner selectors, evidence classes, runtime/resource/fixture profiles, default-check selection | The planning analysis counted 1,018 active rows; tier intent is distributed and `default_check` cannot express the new entry-point roles. | Add one minimum execution tier per active row and generate target membership. |
| Execution topology | `tools/execution_topology_manifest.json` | Runtime, resource, fixture profiles; sequence and service-backed schedules | `test-fast` and `test` have explicit one-job local-to-service barriers; CI and release work wait on whole `check`. | Replace phase schedules with authored graph composition and capability declarations. |
| Scheduler resources | `tools/scheduler_resource_registry.json` | Logical capacity names and automatic policies | Capacity is primarily CPU-derived; the measured host resolved clone capacity to eight and browser work still serialized. | Replace family-specific capacity logic with one run capability snapshot and unit-sized claims. |
| Generated schedule | `tools/scheduler_manifest.json` | Projected check and service-backed work units | Direct and aggregate paths project different units and scheduling behavior. | Generate one graph representation used by every selector. Never hand-edit. |
| Browser manifest | `tools/browser_e2e_batch_manifest.json` | Stages, groups, sessions, reset and isolation metadata | Direct stages create one unit per session and the shell batch loops groups serially. | Preserve semantic group and reset metadata; project each group as a work unit and remove batch-loop scheduling. |
| Scheduler runtime | `tools/harness/scheduler/**` | Multiple scheduler facades, resource reservation, process execution, event and summary emission | Priority reservation can prevent otherwise fitting work; nested scheduler families fragment the critical path. | Converge on one graph compiler, scheduler engine, resource broker, and evidence projector. |
| Browser runtime | `tools/harness/browser/**` | Stack ownership, stage scheduling, resets, Playwright invocation, leaf finalization | A generated stage resource is forced to capacity one and `run-browser-e2e-batch.sh` serializes groups within a session. | Use scheduler-owned stack leases and group commands; retain only lifecycle adapters and semantic reset actions. |
| PostgreSQL fixtures | `internal/testutil/pgtest/pgtest.go` | Suite template, row/package/group databases, transactions, resets, migration scratch | The default template-clone path creates substantial clone pressure; group reuse is narrow and migration replay is repeated. | Introduce explicit lease capabilities, asynchronous pool replenishment, and owner-reviewed fixture migrations. |
| Finalization | `tools/harness/finalization/**` | Schema, duration, scheduler, generation, and drift actions with action caching | No-results finalization spends about 16.8 seconds running generation and drift as separate render passes; broad cache inputs repeatedly miss. | Render once into scratch, validate and compare once, publish atomically when selected, and delete ineffective special caching. |
| Harness contracts | `tools/harness/tests/test-harness-contracts.mjs` | Most harness contract fixtures in a 6,528-line, 104-test file | Two Node test files expose little file-level parallelism and repeatedly load or derive shared indexes. | Partition by semantic harness owner and share one immutable per-run index artifact. |
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
continuing value after graph cutover. `release-browser-readiness` may remain as
an internal/direct selector only if the WF-02 caller inventory finds a real
consumer; otherwise its work is selected directly by `release-check` and its
target is removed.

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

## 6. Refactor Workstreams

| ID | Workstream | Class | Depends on | Status | Required outcome |
| --- | --- | --- | --- | --- | --- |
| WF-00 | Source, scope, and authority bootstrap | root | none | `DONE (planning)` | Inspected inventory, clean baseline, non-goals, decisions, and clause disposition are recorded. Revalidate at implementation start. |
| WF-01 | Measurement truth repair | chain | WF-00 | `TODO` | The 47-target roster closes exactly; internal targets are separate; browser and wrapper work is attributed; Go overhead is corrected. |
| WF-02 | Public contract and tier freeze | chain | WF-01 | `TODO` | Every target/input/ID/schema/artifact/row has a disposition and the five tiers are machine-projectable. |
| WF-03 | Unified graph and scheduler | chain | WF-02 | `TODO` | Direct and aggregate selectors share graph IDs/digests; scheduling is deterministic, resource-fit, fair, cancellable, and deadlock checked. |
| WF-04 | Service and backend fixture plane | parallel | WF-03 | `TODO` | One broker, explicit leases, pooled replenishment, compatible Go grouping, and owner-specific hotspot fixture repairs are complete. |
| WF-05 | Browser decomposition and stack pooling | chain | WF-03, WF-04 | `TODO` | Browser groups schedule independently; stateful affinity, resets, snapshot writes, measurement quietness, and lane cleanup are explicit. |
| WF-06 | Aggregate composition, deduplication, and cache | chain | WF-03, WF-04, WF-05 | `TODO` | Public roots union/deduplicate units, use actual dependencies, and safely reuse only validated hermetic work. |
| WF-07 | Finalization and tool overhead | parallel | WF-01, WF-03 | `TODO` | Single-pass finalization, partitioned harness contracts, and once-per-graph/freshness-keyed vulnerability scans are complete. |
| WF-08 | Owner-first hard cutover and cleanup | chain | WF-03, WF-04, WF-05, WF-06, WF-07 | `TODO` | Owners and authored inputs change first; projections regenerate; public targets switch atomically; obsolete paths and schemas are removed. |
| WF-09 | Accumulated validation and handoff | chain | WF-08 | `TODO` | Functional, lifecycle, artifact, performance, generation, and public-entry-point gates pass at one source state. |

### WF-01 — Measurement truth repair

Future edits start in the observability owner and schema inputs, not in the
checked-in baseline. Required tasks:

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
7. Capture the qualified baseline described in Section 8 before tier or
   scheduling changes.

Exit: baseline closure tests pass, no unattributed material interval remains,
and every retained timing row names one canonical event-derived source.

### WF-02 — Public contract and tier freeze

1. Amend the Testing Harness NLSpec with the retained contracts, tier table,
   work-unit shape, resource/fixture capabilities, caching rules, artifact
   model, and hard-cutover policy in this tracker.
2. Produce a complete machine-generated disposition report for all 145 current
   target records and a complete minimum-tier report for every active test row.
3. Inventory every declared Make variable, inherited environment input,
   command ID, schema ID, artifact path, and public caller. Keep only inputs
   that provide current value; removed inputs fail as undeclared after cutover.
4. Decide `release-browser-readiness` mechanically: retain it only if a direct
   caller outside `release-check` or a distinct human diagnostic role exists;
   otherwise remove it and project its units directly into `release-check`.
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
pass. Complete the trackers, risks, cleanup ledger, and restart record.

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

### Performance protocol

There is no fixed global wall-clock target. Performance closure uses comparable
observations and variability-derived materiality.

For each bottlenecked direct or aggregate target:

1. Require identical source, toolchain, platform, declared inputs, semantic
   selection digest, and capacity profile for executor before/after comparison.
2. Record one cold invocation with reusable work forced off.
3. Discard one warm-up invocation, then record five successful warm invocations.
4. Report p50, p90, median absolute deviation, critical path, resource blocking,
   utilization, setup/fixture work, process count, cache posture, and graph digest.
5. Define material warm improvement as a p50 reduction greater than three times
   the pooled before/after median absolute deviation. The after p90 MUST NOT
   regress beyond that same variability band.
6. Require the addressed structural metric—critical-path contribution,
   resource-blocking wall, setup interval, process count, or duplicate work—to
   decrease as predicted. A smaller tier inventory alone is not executor proof.
7. Treat a cold regression as unresolved even if content-cache hits improve warm
   use. Cache hits are a separate benefit class.

Material improvement is required for `browser-e2e`, its affected browser leaf
targets, `test-fast`, `test`, `check`, `ci`, `release-check`, `agent-finalize`,
`harness-contract`, and `go-vulncheck`. Every other retained public verification
target MUST be non-regressing within its measured variability band.

If three observations do not produce a stable comparison, extend both windows in
matched pairs up to six measured runs. If the conclusion remains unstable, mark
the gate `BLOCKED: unstable measurement environment`; do not refresh the
reference.

### Rebaseline rule

Baseline refresh is the last action after functional, evidence, and performance
acceptance. The writer MUST reject a mixed semantic window or incomplete public
roster. Legacy-equivalent executor results and new tier user-experience results
are reported separately so reduced scope is never presented as scheduler
acceleration.

### Tracker-document verification

This planning-only phase changes Markdown only. Its required command is:

```sh
make lint-markdown
```

Product, generated, conformance, and release checks are intentionally not run
for creation of this tracker.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Owner | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- | --- |
| T-001 | Define target module and scope | scope | `DONE` | none | planning actor | Section 1 | One execution-control seam and exclusions are explicit. |
| T-002 | Inspect current repo state | discovery | `DONE` | T-001 | planning actor | Sections 1–2 | Owners, projections, representative runtime paths, tests, and diagnostic evidence are inventoried. |
| T-003 | Map owner contracts | contracts | `DONE (planning)` | T-002 | harness owners | Section 4 | Every impacted current contract family has retain/revise/delete posture and successor. Owner adoption remains WF-02. |
| T-004 | Freeze characterization evidence | tests | `TODO` | T-003 | harness evidence accounting | CP-01, CP-02 | Corrected legacy-equivalent baseline and missing characterization are retained. |
| T-005 | Plan boundary guardrails | architecture | `DONE` | T-003 | harness command surface | Sections 3, 8 | Graph, ownership, generated, cache, artifact, and compatibility guardrails are binary. |
| T-006 | Plan behavior-preserving moves | implementation | `DONE` | T-004, T-005 | workstream owners | Sections 6–7 | Ordered checkpoints, dependencies, rollback points, and hotspot tasks are defined. |
| T-007 | Plan validation loop | validation | `DONE` | T-006 | harness evidence accounting | Section 8 | Focused, aggregate, lifecycle, artifact, and performance gates are named. |
| T-008 | Update docs/contracts if required | docs | `TODO` | T-003 | Testing Harness owner | WF-02, WF-08 | Owner changes precede authored projections and generated refresh. |
| T-009 | Execute or hand off | handoff | `DONE (planning handoff)` | T-006, T-007 | planning actor | Section 10 | Another actor can begin WF-01 without rediscovery; implementation remains unauthorized by this planning session. |

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

### Future handoff record template

Append one record before ending every implementation session.

| Field | Value |
| --- | --- |
| Date/time | TODO |
| Branch/commit | TODO |
| Worktree state and intentional pre-existing changes | TODO |
| Current workflow/checkpoint | TODO |
| Completed workflows/checkpoints | TODO |
| Changed authored files | TODO |
| Regenerated files and owning generator | TODO |
| Commands run | TODO |
| Passing validation and run roots | TODO |
| Failing validation and classification | TODO |
| Performance source/system/semantic digests | TODO |
| Decisions and deviations | TODO |
| Open risks or blockers | TODO |
| Rollback state | TODO |
| Next workflow/checkpoint | TODO |
| Safe restart command | TODO |

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
| THBR-AC-014 | Functional, artifact, generated, focused owner, direct target, aggregate, finalizer, CI, and release validation passes at one source state, with exact run roots recorded. |
| THBR-AC-015 | Every named bottleneck target materially improves under the variability-derived protocol; every other retained public verification target is non-regressing; baselines refresh only afterward. |
| THBR-AC-016 | The final tracker records changed/authored/generated files, commands, run roots, semantic/system digests, failures, cleanup, decisions, residual risks, and safe restart state. |

### Final checklist

- [ ] WF-01 corrected measurement truth is retained before optimization.
- [ ] Every active row and public input has a target-state disposition.
- [ ] Owner changes precede schemas, authored projections, generation, and code.
- [ ] Direct and aggregate work use one graph and scheduler.
- [ ] Fixture and browser concurrency preserves explicit isolation.
- [ ] No phase wrapper, serial browser loop, duplicate finalizer, or nested runner remains.
- [ ] Cache closure and security freshness fail safe.
- [ ] Artifact roster, timing, critical path, and totals close exactly.
- [ ] Executor improvement is separated from tier inventory reduction.
- [ ] Focused and accumulated validation pass at one source state.
- [ ] Generated output was refreshed only through Make.
- [ ] Obsolete compatibility paths and schema readers are removed.
- [ ] Performance references were refreshed only after acceptance.
- [ ] The final implementation handoff is complete and restartable.
