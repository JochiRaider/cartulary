---
doc_id: cartulary.testing_harness.v3
title: Testing Harness NLSpec
conformance_profile_id: cartulary.testing_harness.current.v3
doc_type: nlspec
status: adopted/current
authority_boundary: Harness mechanics only; command invocation, target selection, scheduling, fixture lifecycle, service ownership, artifact emission, summary emission, cleanup, and harness verification gates.
---

## 1. Status, Scope, and Authority

This NLSpec defines the Cartulary testing harness subsystem. It is the sole
adopted current authority for the harness mechanics identified in
`authority_boundary`, including owner and row selection, catalog validation,
runner adaptation, retained-evidence auditing, and the existing scheduler,
fixture, service, artifact, cleanup, and verification-gate contracts. Adoption
does not make harness readiness evidence product conformance or Core 05
claim-publication evidence.

No harness test, generator, conformance command, or release command reads,
stats, hashes, or otherwise consumes this file or any other documentation file
as executable evidence. Versioned machine inputs are reviewed projections of
this NLSpec and the applicable product/support specifications; executable tools
validate those projections without parsing their source documents.
OpenAPI release change sets are operational review records: each entry contains
only its semantic fingerprint, compatibility classification, owning component,
and rationale. They MUST NOT copy specification requirement identifiers or use
a requirement catalog as approval authority.

**TH-HARNESS-REQ-001**
This NLSpec owns only harness mechanics: command invocation, owner and row
selection, catalog and runner validation, scheduling, fixture lifecycle,
service ownership, artifact emission, summary emission, retained-evidence
auditing, cleanup, and harness verification gates. Core 00 through Core 04
remain the sole owners of product behavior and product-profile authority. Core
05 remains the sole owner of claim-publication and benchmark-publication
activation.

Owner catalogs, verification contracts, schemas, task-surface inputs, and
execution-topology inputs are reviewed derived contracts. They MUST implement
this NLSpec and their cited product or support owners, but they MUST NOT
supersede either source. Migration trackers and handoffs are implementation
authorities only; they MUST NOT create lasting harness behavior. Guides,
generated outputs, repository source, tests, and retained artifacts MUST NOT
become alternate behavior owners.

Frontend readiness mechanics introduced by `browser-e2e-visual`,
`browser-e2e-a11y`, owner test-family manifests,
`tools/frontend_visual_fixture_registry.json`, canonical row and target results,
and frontend visual/accessibility projections are harness and implementation-
readiness mechanics only. They MUST NOT define Core product behavior, promote
visual or accessibility evidence into product conformance, or activate Core 05
claim publication unless a claim-bearing publication predicate is active.

Superseded command IDs, delivery-phase registries, schemas, artifact
identities, and retained runs are historical investigation evidence after the
v3 cutover. A v3 implementation MUST NOT provide aliases, fallback readers,
dual catalogs, dual writers, or newest-artifact fallback for those predecessor
identities. Historical artifacts MUST NOT close a v3 verification or release
gate. A command ID whose version remains current in Section 4.3 is not
superseded merely because its version suffix is lower than three.

Where repository planning, handoffs, or command descriptions use the phrase "production evidence", the current harness interpretation is release-readiness evidence. Release-readiness evidence is a release gate input, not product conformance by itself and not Core 05 claim-publication evidence unless an owner document later promotes a narrower claim-bearing publication boundary.

Harness graph-cache records under `cartulary.harness_cache_record.v2` are local
acceleration mechanics only. They MUST NOT define product behavior, weaken
target projection or event emission, replace drift/security/readiness/cleanup/
publication verdicts, or be cited as product, release, benchmark, or Core 05
evidence. Same-run sharing is represented by one canonical unit referenced by
several target projections; it has no secondary helper-artifact identity.

Fallow static-analysis mechanics introduced by `frontend-fallow-static`, `.fallowrc.json`, `tools/fallow/*`, `cartulary.fallow_reachability_owner.v1`, and `cartulary.fallow_static_summary.v2` are harness and implementation-support mechanics only. They MUST NOT define product behavior, MUST NOT replace TypeScript, Biome, frontend import-boundary checks, tests, security scans, generated-artifact drift checks, or harness gates, MUST NOT activate Fallow Runtime behavior, and MUST NOT be cited as Core 05 publication evidence.
Verified by: TH-HARNESS-AC-013, TH-HARNESS-AC-016, TH-HARNESS-AC-022, TH-HARNESS-AC-026, TH-HARNESS-AC-028

**TH-HARNESS-REQ-002**
A harness conformance claim MUST identify this NLSpec version, the exact public Make target or target set under evaluation, the conformance environment from Section 14, and the retained result root/run ID/run root when retained harness artifacts are used as evidence.
Verified by: TH-HARNESS-AC-015, TH-HARNESS-AC-016

**TH-HARNESS-REQ-003**
The current canonical public invocation surface is Make. In the current profile, a command invocation is canonical only when invoked as `make <target>` from the repository root or through a Make-owned wrapper that preserves the target identity.

Each public target MUST also declare one stable `command_id` using the form `cartulary.harness.command.<name>.v<positive_integer>`. The version is a decimal integer without a leading zero. The `command_id` identifies the command's semantic contract. The Make target name is the current invocation binding for that semantic command. A later adopted NLSpec MAY add additional invocation bindings only when they preserve the same `command_id`, configuration contract, output contract, artifact contract, failure mapping, and cleanup behavior.

Public behavior change gate. Any change to public Make target identity, stable `command_id`, declared schema IDs, retained artifact paths, output shape, failure mapping, or task-surface/topology metadata MUST be specified here before implementation and before generated task-surface or topology outputs are refreshed. Private module movement that preserves those public contracts MAY proceed without changing this gate.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-004, TH-HARNESS-AC-005

**TH-HARNESS-REQ-004**
Generated files under `internal/gen/**`, `packages/protocol-ts/src/generated/**`, `packages/ui-contracts/src/generated/**`, generated task/schedule artifacts, and generated Make includes are downstream generated artifacts. They MUST NOT be hand-edited and MUST NOT become behavior owners unless a later adopted NLSpec explicitly promotes one of them. `tools/task_surface_owner.json` is the authored machine owner for task metadata and Make binding profiles; `tools/task_surface_manifest.json` and generated Make includes are projections of that owner. `tools/execution_topology_manifest.json` owns execution topology and the closed runtime and resource profile registries only; fixture capabilities are owned by the v3 family schema and broker contract. It MUST NOT embed a second task-surface or test-catalog owner.

Make-owned generation MUST render generated harness state into one private
scratch transaction, validate it there, compare it once, and publish it
atomically only for a refresh operation. A generated renderer MUST NOT create
an implicit repo-local intermediate or publish a partial set of related
artifacts.

When a harness helper or catalog path change affects task-surface, topology, or schedule outputs, the owner input MUST be updated first and downstream artifacts MUST be refreshed through `make generate`. The current verification ladder is `make generate-drift`, `make generated-artifact-policy-check`, and `make json-shape-check`. Generated files MUST NOT be edited as the source of truth for helper-path or catalog migration.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-016

**TH-HARNESS-REQ-005**
Direct package scripts, raw scripts, raw Go/Vitest/Playwright/Biome/Vite/pnpm commands, and tool-specific reports are developer conveniences or child commands unless a public Make target invokes them. Direct invocation of those surfaces MUST NOT be treated as equivalent to a canonical harness run.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-005

## 2. Purpose, Non-Goals, and Conformance Boundary

The testing harness exists to provide a reproducible repository command surface for local developers, CI entrypoints, coding agents, and release verification. It provides deterministic target selection, bounded output, structured artifacts, explicit service ownership, controlled fixture lifecycle, stable failure classification, and destructive cleanup gates.

**TH-HARNESS-REQ-006**
The harness MUST provide all of the following for public Make targets:

- deterministic target-class and target-selection metadata;
- declared configuration resolution;
- output-mode behavior;
- exit-code mapping;
- retained artifact identity when a target declares artifacts;
- failure classification that separates product assertion failures from harness operational failures;
- cleanup predicates for every destructive operation.
Verified by: TH-HARNESS-AC-001..TH-HARNESS-AC-017

**TH-HARNESS-REQ-007**
The harness MUST NOT claim provider-specific hosted CI behavior, benchmark publication, release publication readiness, macOS support, Windows-native support, Podman support, or Playwright artifact schema stability unless those areas are explicitly included in this NLSpec's current conformance profile or in a later adopted NLSpec. The current visual-snapshot refresh authority is limited to the helper-only maintenance contract stated in Sections 6, 8, 11, 15, and 17.
Verified by: TH-HARNESS-AC-012, TH-HARNESS-AC-016

**TH-HARNESS-REQ-008**
Logical scheduler resources are execution constraints inside the harness. They MUST NOT be represented as guarantees about physical CPU, I/O, Docker, database, object-store, browser, or network capacity.
Verified by: TH-HARNESS-AC-006

**TH-HARNESS-REQ-009**
Adopted Sections 1 through 17 MUST close every current-conformance harness behavior they name. Unbounded delegation to a target, producer, tool, implementation, or applicability judgment is invalid unless the same requirement cites a closed table, schema attachment, algorithm, or explicitly non-normative diagnostic boundary. Generated manifests and generated Make includes MAY mirror a closed contract, but they MUST NOT be the only current-conformance owner for a public harness behavior.

In adopted Sections 1 through 17, `MAY` means true implementation freedom whose divergent realizations remain interchangeable to callers. Acceptance-bearing behavior MUST use `MUST` or `MUST NOT`; this document does not use an advisory normative keyword.
Verified by: TH-HARNESS-AC-016, TH-HARNESS-AC-029

**TH-HARNESS-REQ-010**
The normative main body in Sections 1 through 17 MUST contain every load-bearing default, bound, omission rule, mapping, algorithm, failure consequence, and acceptance obligation. Appendices and Section 18 MAY contain rationale, examples, sample objects, and research traceability only. A supporting section MUST NOT be the sole owner of current-conformance behavior.
Verified by: TH-HARNESS-AC-016, TH-HARNESS-AC-071

**TH-HARNESS-REQ-011**
The current normative imported format dependencies are JSON Schema Draft 2020-12, RFC 8785 JSON Canonicalization Scheme with its verified errata, SHA-256 as specified by FIPS 180-4, and RFC 3339 timestamps constrained by Section 8. A later revision MUST replace an imported dependency explicitly; implementation-library defaults MUST NOT silently change the format contract.
Verified by: TH-HARNESS-AC-062, TH-HARNESS-AC-066, TH-HARNESS-AC-071

**TH-HARNESS-REQ-021**
Repository test-harness observability is a harness diagnostic subsystem, not
application runtime telemetry. This NLSpec owns harness trace reconstruction,
harness performance metrics, retained diagnostic artifacts, explicit post-run
OTLP export, and the public commands that inspect or export them. The pinned
OpenTelemetry source and protocol baseline is consumed from machine snapshot
and dependency contracts under `contracts/otel/`; this reference MUST NOT
duplicate or silently rebaseline it. Application telemetry scopes, application
`cartulary.module` attributes, deployment telemetry configuration, and the
server telemetry bootstrap MUST NOT become harness configuration or harness
evidence owners.
Verified by: TH-HARNESS-AC-072, TH-HARNESS-AC-076

### 2.1 V3 execution-control contract

This section owns the complete current execution-control seam. No handoff or
historical tracker supplies additional current requirements. The former
fixed-cap, fixed-shard, phase-scheduler, and fixed-browser-capacity contracts
have been replaced in their original sections; the requirements below state the
shared architecture those sections implement.

**TH-HARNESS-REQ-800**
Every active test-family row MUST declare exactly one `minimum_tier` from the
closed ordered set `fast`, `standard`, `full`, or `release`. The order is
monotonic: a selector for tier N admits rows whose minimum tier is N or lower.
The authored catalog MUST NOT contain `default_check`, an inferred tier, or an
implicit tier fallback.

The five aggregate policies are closed:

| Entry point | Row selection | Additional policy work |
| --- | --- | --- |
| `test-fast` | `fast` | Readiness required by selected fast rows only. |
| `check` | `fast` and `standard` | Standard static, boundary, generated-drift, security, and evidence checks. |
| `test` | `fast`, `standard`, and `full` | Functional evidence collation only. |
| `ci` | `fast`, `standard`, and `full` | CI security, policy, and deployable-shape work. |
| `release-check` | all four tiers | CI policy plus release, build, compatibility, browser, accessibility, visual, measurement, SBOM, license, and release-evidence work. |

Policy, packaging, and release units are selected by aggregate policy and MUST
NOT masquerade as product test rows. An omitted owner slice selects the complete
active owner inventory. An explicit row slice selects exactly its valid active
rows and is not tier-filtered. Removing a row requires an owner-backed
redundancy record naming the surviving row and proving equivalent behavior,
runner, fixture, and evidence semantics.
Verified by: TH-HARNESS-AC-082

**TH-HARNESS-REQ-801**
Every schedulable action MUST compile to one canonical work unit with these
closed logical fields: `unit_id`, `owner_id`, `kind`, `command`, `needs`,
`resource_claims`, `fixture_lease`, `service_dependencies`, `cache_policy`,
`timeout_ms`, `current_run_evidence_outputs`, `reusable_artifact_outputs`,
`failure_policy`, `estimated_work_ms`, and `semantic_digest`. Unit identity is
independent of aggregate target, queue
position, run ID, and invocation timestamp. The semantic digest closes over the
selection, command and declared environment, graph dependencies, resource
profile, fixture lease, service dependencies, cache inputs, timeout, failure
policy, and evidence contract.

A unit resolved from a populated fixture profile also contains
`fixture_profile_id` and `snapshot_key`. Those values participate in its
unit identity, command environment where consumed, dependencies, and semantic
digest. Omission is valid only for a unit with no populated fixture profile.

Equivalent direct, aggregate, leaf, and owner-slice selections MUST produce the
same unit IDs and semantic digests. One semantic unit MAY contribute to several
target projections in one graph but MUST execute no more than once. Unknown IDs,
missing dependencies, cycles, unsupported resource names, infeasible claims,
unknown fixture capabilities, unknown cache profiles, or incomplete output
contracts MUST fail before child work.
Verified by: TH-HARNESS-AC-083

**TH-HARNESS-REQ-802**
One graph compiler and one scheduler facade MUST serve public, leaf, direct,
aggregate, and owner-slice selectors. The scheduler MUST:

1. capture one immutable host-capability snapshot;
2. rank each ready unit by qualified work plus longest estimated downstream path;
3. admit the highest-ranked resource-fitting unit with stable `unit_id` ties;
4. backfill with any lower-ranked fitting unit without reserving unheld resources;
5. add monotonic waiting age to effective rank so backfill cannot starve work;
6. block descendants of failed required predecessors while continuing only
   independent work permitted by target failure policy; and
7. on cancellation, stop admission, terminate owned process groups, release or
   destroy leases, and publish a partial non-success summary.

The authored graph, qualified cost model, capability snapshot, and capacity
override declaration MUST determine admission reproducibly. Live self-tuning is
outside v3. A scheduler family, shell loop, or nested Make invocation MUST NOT
encode a second scheduling policy.
Verified by: TH-HARNESS-AC-084

**TH-HARNESS-REQ-803**
The capability snapshot MUST record effective CPU set, memory bound, process
limit, writable-volume availability, port-lane availability, and declared
service readiness. Resource profiles declare CPU tokens, memory bytes, IO
weight, process slots, and scarce service capabilities. Capacity for a unit
class is the minimum capacity implied by every claimed dimension. Missing or
invalid data fails safe to one lane for the affected scarce resource.

Family-specific capacity-forwarding variables are unsupported after cutover.
Graph-backed public commands MAY accept one schema-validated capacity override
declaration file through `CARTULARY_HARNESS_CAPACITY_OVERRIDE`; inline JSON is
unsupported. Every override is range checked and recorded with its source in
the run manifest. Runner-specific `VITEST_MAX_WORKERS` and `PLAYWRIGHT_WORKERS`
remain valid only for matching runner units and their effect MUST be represented
by the unit's resource claims.

The resource profile selected by `resource_profile_id` is the sole owner of a
unit's base numeric claims. Runner-local default claim tables are forbidden.
The execution topology MUST also own one minimum resource-claim map for every
admitted service dependency. Compilers merge each declared dependency into the
selected profile with maximum-per-resource semantics, so a profile MAY reserve
more than the service minimum but no path may reserve less. Browser PostgreSQL
intensity and object-store admission claims follow this same rule; they are not
compiler constants.

Telemetry MUST expose requested and resolved capacity, peak use, saturation
intervals, queue duration, blocking resources, and actual resource holders.
Verified by: TH-HARNESS-AC-085

**TH-HARNESS-REQ-804**
One run-scoped broker owns infrastructure readiness and cleanup. It MAY attach
to already managed Postgres and object-store services, but it MUST NOT close a
borrowed resource. Every application stack, browser stack, database, namespace,
port set, and process created by the broker is owned by an explicit lease.

The fixture capability set is closed to `none`, `postgres_transaction`,
`postgres_dedicated`, `postgres_migration`,
`object_store_namespace`, `managed_process`, and `browser_stack`. Omission is
invalid and there is no implicit template-clone fallback. Failed reset, health,
or contamination proof destroys and replaces a lease. The broker MUST replenish
clean dedicated databases asynchronously. A digest-keyed migrated template is valid
only when an owner test proves the same migration chain produced it; every owned
migration chain retains at least one full-chain proof.

Fixture capability describes broker-owned isolation and lifecycle only. It is
not a declaration of every managed service a helper consumes. Each catalog row
and work unit therefore carries a separate, required, ASCII-sorted,
duplicate-free `service_dependencies` array. The current service-dependency set
is closed to `object_store` and `postgres`; adding a service requires an adopted
topology and schema revision, not a composite fixture capability.

Infrastructure setup MAY retry once on a newly created lane only when no product
work began and the failure occurred before service readiness polling began.
Once readiness polling begins, its expiry is terminal for that lane and MUST NOT
start a replacement lane. Product test failures MUST NOT be retried by scheduler
policy.
Verified by: TH-HARNESS-AC-086

**TH-HARNESS-REQ-805**
Compatible Go rows share a process only when module, exact package selection,
runtime-binary set, resource profile, fixture capability and profile, service
dependencies, owner-required partitions, process-isolation policy, selector,
environment, and evidence-class keys are exact. Resource isolation supplied by
`postgres_dedicated`, `postgres_migration`, or `object_store_namespace` is
per-test ownership and does not by itself require a separate Go process. A
`managed_process` item remains process-isolated because process lifecycle is the
subject under test. An exact-symbol Go row MAY additionally declare the typed
`process_isolation="exclusive"` catalog input only when the assertion installs
or observes process-global state that cannot coexist with another selected row;
omission means `shared`, and the compiler MUST reject the input for any other
runner. The fixed eight-symbol and twelve-second limits and
speculative extra shards are unsupported. Compatibility groups and child CPU
budgets are formed by TH-HARNESS-REQ-379; each child claims the same number of
scheduler CPU tokens as its scheduler-owned `GOMAXPROCS`. Rows within one
compatible group are ordered by descending predicted work with stable row-ID
tie-breaking. Go JSON remains authoritative for row outcomes.
Verified by: TH-HARNESS-AC-087

**TH-HARNESS-REQ-806**
Every semantic browser group is an ordinary work unit. Stack lanes are broker
leases; shell code MUST NOT own group order or capacity. Stateful groups declare
an affinity key and explicit dependency chain, and resets are graph units at the
exact required boundary. The compiler derives one browser affinity chain from
the exact tuple `(stage, browser_session_group, runtime_profile_id,
resource_profile_id, fixture_profile_id-or-mutable)`. After row/owner/target
selection, it inserts a reset only between a selected predecessor and selected
successor in that exact chain. A first selected group MUST NOT receive a leading
reset. The reset label is derived from the predecessor and successor IDs;
authored topology and browser-manifest inputs MUST NOT declare `reset_before` or
another reset-placement override. Direct, owner-sliced, and aggregate selection
therefore compile the same selected chain. Incompatible affinity tuples never
share a reset, and immutable measurement-fixture clones remain reset-free.
Read-only visual work MAY overlap; snapshot update work
claims an exclusive write key for its output. Measurement work is admitted only
when no ordinary unit holds the resources covered by its quiet profile. Every
executable graph unit that consumes CPU, memory, process, I/O, database,
object-store, service-stack, or browser resources MUST hold `host_activity` in
shared mode. A `browser_measurement_quiet` session MUST hold `host_activity` in
exclusive mode from reset through fixture preparation, readiness and traffic
stabilization, warm-up, measured operations, artifact finalization, cleanup,
and lease release. Once an exclusive waiter is ready, later shared work MUST
NOT backfill ahead of it. A central graph validator MUST reject host-active work
without exactly one valid `host_activity` mode; constructors MUST resolve the
mode and numeric claims from the catalog topology profile instead of restating
them locally.

Every p95 browser measurement consumes the typed harness policy: exactly one
discarded warm-up operation, exactly 100 finite non-negative measured
operations, nearest-rank p95 at `ceil(0.95 * N) - 1`, and zero retries. Before a
threshold assertion, and on safe partial setup or sampling failure, each
predicate MUST emit one immutable
`cartulary.frontend_measurement_observation.v1` with predicate, criterion,
policy, fixture, digest, threshold, counts, p50 and p95 when available, numeric
stage samples, and traffic qualification. The observation MUST exclude entered
text, record and transaction IDs, credentials, tokens, and payload data.

An observation is not qualification evidence. After the browser-stack lease
has released and clone cleanup has completed, the row finalizer MUST combine
the observation with the immutable snapshot-build and snapshot-lease artifacts
and scheduler overlap proof into
`cartulary.frontend_measurement_summary.v2`. The target finalizer MUST accept
only current v2 summaries and emit
`cartulary.frontend_measurement_aggregate.v2`. Missing, malformed, stale,
unredacted, v1-only, digest-inconsistent, not-cleaned, or overlapping evidence
fails closed as `environment_not_qualified`.

A failed or unhealthy lane is quarantined and cleaned, and its product work is
not retried. Direct and aggregate selectors use identical group IDs,
dependencies, profiles, and evidence outputs.

Ordinary `webserver_backed` groups MUST use the `browser_functional` resource
profile, whose authored PostgreSQL claim is two tokens. Their compiler owns at
most four run-scoped warm lanes across the selected closure. A lane is immutable
for one runtime profile; groups with incompatible runtime profiles MUST NOT share
it. The compiler MUST assign at least one lane to each selected runtime-profile
partition, then allocate each remaining lane to the partition with the greatest
current total-estimated-work-per-lane value, with stable runtime-profile ID ties.
More than four selected runtime profiles is invalid for this topology.

Within each compatible partition, group cost is the sum of the selected rows'
authored evidence-class estimates. Groups are considered in descending cost
order with stable group-ID ties and assigned to the least-loaded compatible lane
with stable lane-ID ties. The emitted dependency chain, not scheduler timing,
owns that stable longest-processing-time-first plan. Each lane has one readiness
unit, one affinity identity, and one warm process stack. Its first group follows
readiness directly; every later group follows an explicit reset unit, and every
reset follows the preceding group even when that product group failed.

Each functional group MUST receive a new browser context, group output root, and
reset generation. Before a retained stack serves the next group, reset MUST
restore the exact validated lane database to its proved migrated-template
semantic state, re-create the exact validated lane bucket, clear browser session
storage, and reset route-idempotency, fault, and clock state. The database name,
process ports, runtime profile, and secret-bearing child environment remain the
lane's immutable startup identity; physically dropping that live database is
forbidden because it invalidates the warm process lease. A product failure is
never retried and quarantines its allocation; its successor reset may obtain a
fresh allocation with a new concrete session evidence identity.
A reset failure is required target failure, quarantines the allocation, does not
retry prior product work, and does not block later independent groups from
obtaining a fresh clean stack. Stateful, measurement, accessibility, visual, and
support stages retain their existing resource profiles and isolation topology.
Verified by: TH-HARNESS-AC-088

**TH-HARNESS-REQ-807**
Aggregate targets compile a union graph and MUST NOT invoke child aggregates or
nested schedulers. Identical readiness, build, scanner, test, and finalizer units
deduplicate by unit ID and semantic digest. Exact producer dependencies replace
whole-target phase barriers. Noncritical policy work MUST NOT starve the
functional critical path.

Cache policy is closed to `none`, `same_run`, or `content_addressed`.
Current-run evidence outputs contain run identity, timestamps, row or unit
results, and summaries; they MUST be regenerated for every invocation and MUST
NOT be stored or restored as reusable artifacts. A reusable artifact output is
cacheable only when its complete type, normalized relative path, destination
class, mode, producer identity, semantic input digest, and content digest are
declared. A content-addressed profile closes tracked and relevant untracked
source, configuration, declared environment, tool/runtime and dependency
digests, selection, output schemas, and helper implementations. A hit validates
the complete reusable artifact manifest, publishes it all-or-none beneath
validated destinations, and emits fresh current-run evidence. A unit with an
incomplete reusable-output contract is cache-ineligible. Missing, corrupt,
stale, or mismatched records are misses and never success. Stateful tests,
lifecycle, migration
assertions, browser work, measurement, cleanup, destructive safeguards,
generated-drift verdicts, and release publication remain uncached.
Suite-scoped populated fixture construction and cloning under
TH-HARNESS-REQ-812 through TH-HARNESS-REQ-815 are lifecycle work, not cached
test results. They MUST execute once per resolved key within one suite and MUST
NOT be discovered, reused, translated, or retained for reuse across runs.
Vulnerability findings are reusable only when a stable vulnerability-database
revision is proven and keyed; otherwise the scan runs.

Unit-aware cache profiles MUST derive conservative dependency closures from the
run's immutable executable-source index. A selected Go package closure includes
the exact selected packages, their repository-local transitive imports, tracked
and relevant untracked files beneath those package boundaries, module and
workspace manifests and lockfiles, toolchain pins, selected owner and selector
contracts, helper and output-schema implementations, declared environment
inputs, and generator identities. A selected TypeScript closure includes the
workspace containing every selected file, all transitive repository-workspace
dependencies, the same manifest, lockfile, toolchain, selector, helper, schema,
environment, and generator classes, and all tracked and relevant untracked files
inside those workspaces. Add, delete, rename, manifest, dependency, schema,
toolchain, selector, or helper changes therefore change the semantic input
digest. Incomplete parsing, an unknown selected row or package/workspace, an
unresolved repository-local dependency, or any other incomplete dependency
knowledge MUST select the registered safe broad closure; it MUST NOT produce a
narrow key.

Content-addressed entries are immutable directories published from exclusive
staging. A writer MUST NOT delete, replace, or mutate a valid concurrent entry.
Each stored artifact manifest is matched exactly against the producer's current
declaration and contains its type, normalized relative path, destination class,
mode, content digest, producer identity, and semantic input digest. Directory
artifacts additionally close the exact normalized member set, member types,
modes, and digests. Lookup MUST reject absolute or traversing paths, destination
or producer mismatches, missing or surplus payloads, partial entries, wrong
modes or digests, symlinks, hard links, and special files in the entry,
destination ancestry, staging tree, or restored tree. It validates and stages
the complete output set before an all-or-none publish. An invalid entry is
quarantined by exact proved identity and the unit executes as a miss without
publishing any cached byte. A concurrent valid entry wins; no reader or writer
may overwrite it.
Verified by: TH-HARNESS-AC-089

**TH-HARNESS-REQ-808**
The v3 retained run contract contains `run-manifest.json`,
`unit-events.ndjson`, `run-summary.json`, and
`target-summaries/<target>.json`. The manifest records command and declared
inputs, source/toolchain/system digests, capability snapshot, graph digest,
cache mode, and start time. Unit events are the sole timing authority for target,
hotspot, pressure, resource, critical-path, cache, and baseline projections.
Material wrapper/setup work is a first-class unit or explicit unattributed
error; it is never silently copied to several leaves.

Historical artifacts remain archival bytes. Current commands MUST reject old
schema input and MUST NOT dual-write, translate, discover, or select the newest
legacy artifact.
Verified by: TH-HARNESS-AC-090

**TH-HARNESS-REQ-809**
The v3 hard cutover retains the current public Make names whose task-surface rows
own distinct semantic operations. It removes the internal phase targets
`test-local`, `test-fast-service-backed`, `test-service-backed`,
`check-service-backed`, and `release-browser-readiness`.
`browser-e2e-support` remains distinct release-support work. Materially changed
command IDs and schema IDs version exactly once. There are no aliases,
forwarding wrappers, dual readers, or dual writers.

Performance acceptance uses TH-HARNESS-REQ-377 through TH-HARNESS-REQ-380. No
fixed global seconds target or fixed peer-balance ratio is a v3 conformance
requirement. A reduced tier inventory is reported separately from unchanged-
workload executor improvement.
Verified by: TH-HARNESS-AC-091

**TH-HARNESS-REQ-810**
The PostgreSQL fixture broker MUST expose version-targeted migration execution
only through an opaque migration-database capability issued for a disposable
`postgres_migration` scratch lease. The capability MUST bind an unexported
harness identity and its harness-owned database handle, MUST always use the one
canonical embedded repository migration source, and MUST NOT accept a
caller-selected source or permit construction around an arbitrary database.

The capability MAY expose the underlying `*sql.DB` for assertions against that
same scratch database. Its apply-through operation MUST reject target versions
less than or equal to zero before source or database access. Its
rollback-through operation MUST reject targets below zero before source or
database access and MAY accept zero. These operations are fixture mechanics;
they MUST NOT be exported from production migration or PostgreSQL packages and
MUST NOT establish production rollback, recovery, or downgrade behavior.

Database preparation MUST return the prepared database and error only.
Preparation lifecycle events are the sole authoritative evidence of template
cloning, full-chain migration, failure, and cleanup; a duplicate preparation
status object is unsupported. The broker MUST preserve ordinary owned-resource
cleanup and MUST NOT close or delete a borrowed database capability.
Verified by: TH-HARNESS-AC-094

**TH-HARNESS-REQ-811**
The disposable PostgreSQL migration capability MUST use the Production DDL
Rebaseline v2 catalog and PostgreSQL major 16. It MUST provision the exact
administrator-owned `public` schema, `pgcrypto` 1.3 and `citext` 1.6
prerequisites, fixed roles, deployment logins, memberships, database grants,
schema ownership, extension privilege cleanup, and default privileges required
by Core 04 REQ-04-153 before applying authored SQL.

The full-chain fixture MUST apply exactly versions `1..29` from a pristine
database. Contamination fixtures MUST cover pre-existing Cartulary objects,
wrong extension version or schema, missing prerequisites, v1 lineage, foreign
lineage, and unmarked nonzero Goose history. The historical-line fixtures MUST
construct only the minimum synthetic ledger and lineage rows needed to prove
pre-DDL rejection and MUST NOT retain or execute v1 migration SQL.

Targeted rollback remains a harness-only operation. Rollback through version
zero MUST leave no Cartulary-authored table, view, sequence, type, routine,
trigger, constraint, index, lineage object, or application-created schema. It
MUST retain administrator-provisioned roles/logins, `pgcrypto`, `citext`, their
extension-managed objects, and Goose's exact version-zero metadata residue. No
Down section may use `CASCADE`, drop an extension, or establish a production
rollback, downgrade, restore, or recovery capability.

Harness role and ACL fixtures MUST prove newly created and recycled physical
connections establish the exact `session_user` and `current_user`; runtime and
Recovery own no object and cannot assume another fixed role; runtime lacks
`TRUNCATE`, sequence update, `session_replication_role`, DDL, and ledger
mutation; Recovery can complete backup, restore, restore verification, journal,
audit, sequence restoration, and projection rebuild without schema-owner
membership; `PUBLIC` and future-object defaults are closed; and every positive
and negative operation matches the authored object access classes.

Credential fixtures MUST cover all three purposes, resolver precedence,
retired and cross-purpose presence without value reads, bounded no-follow file
decoding, safe failure mapping, and complete diagnostic redaction. Profile
claim-state fixtures MUST prove that physical extension-profile tables and
metadata do not claim a profile or create authoritative state presence.
Verified by: TH-HARNESS-AC-095

**TH-HARNESS-REQ-812**
`tools/performance_fixture_snapshot_owner.json` is the sole authored harness
registry for populated performance-fixture profiles and MUST validate as
`cartulary.performance_fixture_snapshot_owner.v2`. Each active profile closes
its stable profile ID, fixture version and seed, compatible runner, evidence
class, selector stage, runtime profile, resource profile, fixture capability
and services, verification-to-predicate bindings, source-contract identities,
ordered source-owner contributions and receipt-count expectations, semantic
count and condition expectations, runtime credential sets, compatibility,
cleanup, artifact, and redaction policies. These collections are structural:
generic schemas and lifecycle code MUST NOT encode a particular profile's
identities, counts, conditions, or credentials.

The registry MUST generate the immutable Go descriptor catalog and the
cross-language snapshot-key vectors through the canonical generation target.
Generated lookup returns defensive copies. JavaScript and Go consumers MUST
resolve the same descriptor and MUST NOT maintain handwritten profile-fact
mirrors. Generated descriptors and key vectors MUST contain no credential,
runtime identity, or runtime path values. A synthetic second active profile
MUST render and route without a profile-specific generic-code branch.

`tools/harness/performance-fixture` is the JavaScript owner for strict registry
loading, active-profile lookup, verification-to-predicate mapping, canonical
snapshot keying, and independent profile grouping. Test-catalog, graph,
broker, browser, finalizer, diagnostic, and release code are consumers of that
owner and MUST NOT implement, re-export, or conditionally reinterpret those
mechanics. RFC 8785 canonical JSON and semantic digest primitives belong to the
neutral harness contract owner; fixture and catalog modules MUST consume them
without a compatibility re-export.

An active catalog row MAY declare `fixture_profile_id`. Every verification
bound by the registry MUST declare its exact active profile, and an unbound row
MUST omit the field. Missing, unknown, inactive, duplicate, incompatible, or
divergent profile routing MUST fail before child work. File names, titles,
scenario IDs, predicate IDs, constructor names, targets, and resource profiles
MUST NOT infer a fixture profile.

The snapshot-key input schema is selected only by the active profile's artifact
policy. Its structural envelope contains the selected schema identity,
lowercase raw 64-character migration and source-contract digests, fixture
version, integer seed, and any profile identity required by that schema
generation. `snapshot_key` is the raw lowercase SHA-256 of RFC 8785 canonical
JSON for the exact envelope. Artifact-reference digests retain the separate
`sha256:<hex>` convention. Go and JavaScript consumers MUST pass the generated
key vectors for every active profile. An unselected key-schema version is not
valid current evidence.
Verified by: TH-HARNESS-AC-096

**TH-HARNESS-REQ-813**
Populated fixture construction MUST use one closed owner-neutral contribution
contract. Auth, Incidents, Entities, Timeline, Links, and Projections each
construct their registered contribution through production application,
persistence, query, and validation boundaries owned by that source owner. The
harness validates complete membership, unique identities, versions, source
contracts, and an acyclic dependency order before the first mutation.

Contributors return only versioned bounded counts and safe semantic digests.
Credentials, account or record identities, physical database identities,
paths, timestamps, and generated entropy MUST NOT enter receipts or the
semantic validation digest. Equal profile, source contracts, and seed MUST
produce equal semantic digests. Missing, duplicate, reordered, cyclic,
incompatible, unauthorized, partially valid, or post-build mutated input MUST
fail before sealing and MUST be cleaned. Raw browser fixture SQL and a
monolithic cross-owner loader are unsupported.

Every build attempt MUST emit one separate
`cartulary.performance_fixture_build_diagnostics.v2` current-run artifact. The
artifact records only the fixture-profile and builder identities, terminal build
state, a bounded failure stage, a construction count of exactly one, total duration,
semantic-validation duration, and ordered per-contribution terminal state and duration plus
optional bounded batch count, configured batch size, and item count. It MUST use
the selected profile's redaction-policy identity and MUST NOT
contain credentials, account or record identities, database identities, paths,
payload values, error text, or generated entropy. A failed contribution MUST retain
the safe identities, terminal states, and durations of completed contributions and
the failing contribution so release-scale contention remains attributable. The diagnostic artifact is not
a contribution receipt, does not participate in the snapshot key or semantic
validation digest, and is not qualification evidence. A timing-only difference
MUST NOT alter the immutable semantic build artifact.

Before sealing, the builder MUST close its PostgreSQL pool and cross-check every
remaining template connection against the server-assigned backend process IDs
captured by that pool. Proven builder connections MAY receive one bounded drain
window because server-side disconnect observation can lag client pool closure.
An unrecognized connection fails immediately; the builder MUST NOT terminate it
or seal around it. Exhausting the drain window is a finalization failure and the
unsealed template is cleaned.

The graph-owned snapshot builder is the sole path permitted to construct the
exact active production fixture. Direct contribution and assembler contract
tests MAY validate its generated descriptors, counts, ordering, failure
propagation, semantic digest, and redaction without materializing production
volume. Snapshot lifecycle integration MUST use a bounded deterministic test
profile while retaining four-clone isolation, concurrency, partial-build,
corruption, cancellation, credential, and cleanup coverage. Removed production
assembly rows require an owner-backed catalog migration naming those direct
contracts and canonical snapshot lifecycle evidence; an independent repeated
production build is not current conformance evidence.

Generic contribution and validation mechanics MUST consume the selected
generated profile descriptor. Exact expectation interpretation belongs to the
named profile adapter and its source-owner tests. Adding a profile MUST require
only a registry entry, its explicit adapter, and owner-specific validation; it
MUST NOT require modifying a generic schema or adding a profile conditional to
generic construction, keying, lifecycle, or evidence code.
Verified by: TH-HARNESS-AC-097

**TH-HARNESS-REQ-814**
The compiler MUST resolve one canonical shared builder unit for each distinct
runtime-profile, fixture-profile, and snapshot-key tuple. Its unit identity is
`fixture_snapshot:<runtime_profile_id>:<fixture_profile_id>:<snapshot_key>`,
and the resolved profile and key participate in unit identity, semantic
digests, direct plans, row and owner slices, aggregate plans, and release
plans. Construction holds shared `host_activity` and finishes before any
dependent predicate requests its exclusive quiet session.

The builder clones the migrated template, runs the closed contribution graph,
performs exact semantic validation, closes its own connections, rejects unknown
open connections, seals the populated database, and emits one immutable
build artifact whose schema is selected by the resolved profile's active
artifact policy. Concurrent
same-key requests join one suite-local build; different keys remain separate.
Selecting any nonempty subset of the active Timeline measurement rows MUST
therefore produce exactly one production construction and one build diagnostic
with `construction_count=1`. A harness closure with no populated measurement
row MUST produce no production snapshot builder and no construction diagnostic.
Each dependent predicate receives one isolated database clone, one empty
object-store bucket, explicit lease identity and clone ordinal, and a private
copy of its typed runtime bundle. A populated-profile row has no empty-template
or live-assembly fallback.

Clone preparation, readiness, traffic, warm-up, measurement, finalization,
session and process shutdown, credential-copy deletion, database and bucket
deletion, and lease release remain inside that predicate's exclusive quiet
session. Suite cleanup unseals and drops the populated template only after all
dependent leases finish. Success, failure, cancellation, partial construction,
corruption, and finalizer failure MUST perform idempotent active cleanup;
lease-scoped bounded janitor recovery is the only fallback and MUST NOT delete
an unproven resource.
Verified by: TH-HARNESS-AC-098

**TH-HARNESS-REQ-815**
Background credentials MUST be suite-random and stored only in the suite's
external private `0700` runtime root in a `0600` bundle of the active
profile-selected runtime schema. The populated template
contains accounts and required memberships but no sessions, tokens, cookies,
browser state, traffic, object payloads, or predicate-local mutations. Each
predicate receives its own private bundle copy, authenticates through the
ordinary login path, and deletes that copy before its snapshot lease finalizes.
The suite bundle is deleted during suite cleanup.

The private root is the one suite-runtime boundary defined by
TH-HARNESS-REQ-603. Template bundles, clone copies, browser state, stack
environments, recovery DSNs, key material, service leases, and raw child
captures MUST NOT be created below a result root. Retained lease evidence is a
separate immutable projection containing opaque identity, resource classes,
and cleanup outcomes; it MUST NOT contain runtime paths, process handles,
ports, credentials, or resource administration identities.

Snapshot evidence is immutable and two-stage. The shared build artifact proves
construction and sealing. After cleanup, one
artifact of the active profile-selected lease schema proves a
predicate's creation, isolation, credential-copy deletion, session and process
shutdown, and database and bucket cleanup. A summary of the active
profile-selected schema MUST
reference both immutable artifacts with run-relative paths and
`sha256:<hex>` digests; no producer may mutate an artifact after another
artifact references its digest. The active profile-selected aggregate MUST
prove one builder, distinct
clone ordinals, one key, zero cross-clone visibility, complete cleanup, zero
scheduler overlap, exact traffic and sampling, and redaction. Historical v1
summaries and any superseded evidence generations remain inspectable bytes but
cannot qualify current source and MUST NOT be dual-written or translated.

The sole active populated-fixture evidence generation is selected by the v2
profile registry and consists of snapshot-key v2, snapshot v2, snapshot-lease
v2, runtime v2, frontend-measurement-observation v2,
frontend-measurement-summary v3, and frontend-measurement-aggregate v3. A
summary v3 contains bounded qualification rollups and immutable run-relative
references with digests to its observation, build, and retained lease
artifacts; it MUST NOT embed any of those artifacts. An aggregate v3 contains
independent profile groups, immutable summary references, and bounded
cross-row rollups; it MUST NOT embed observations or complete summaries. The
browser target result wrapper is v3 and the retained browser-stack lease is v2.

Superseded performance evidence schemas live outside the active schema root in
one digest-pinned, read-only historical registry. Only diagnostic and schema
integrity tools may load that registry. Graph compilation, producers,
finalizers, qualification, release evidence, and active schema attachment
resolution MUST reject historical identities. Historical validation proves
only that retained bytes match their former shape; it MUST NOT translate those
bytes, synthesize current evidence, or qualify current source.
Verified by: TH-HARNESS-AC-099

**TH-HARNESS-REQ-816**
The explicit local managed-service session commands are
`test-services-session-up`, `test-services-session-status`, and
`test-services-session-down`, with command IDs
`cartulary.harness.command.test_services_session_up.v1`,
`cartulary.harness.command.test_services_session_status.v1`, and
`cartulary.harness.command.test_services_session_down.v1`. The commands own one
`cartulary.test_services.local_session.v1` descriptor beneath the external
machine-cache root. The descriptor parent is mode `0700`; the descriptor is a
mode-`0600`, owner-UID regular file reached without symlink traversal. It closes
the suite/session identity, owner UID, schema, configuration, tool, image and
service digests, exact container proof, expiry, and secret-capable attach
environment. Status emits only the redacted
`cartulary.test_services.local_session_status.v1` projection.

`CARTULARY_TEST_SERVICES_MODE` is closed to `owned` or `attach` and defaults to
`owned`. `attach` requires an absolute
`CARTULARY_TEST_SERVICES_SESSION_FILE`, defaulting to
`${CARTULARY_MACHINE_CACHE_DIR}/test-services/session.json`; descriptor presence
alone MUST NOT attach. Child attachment authority consists only of the exact
suite ID, a live runtime or borrower lease, schema-valid service metadata, the
current readiness generation, and matching container proof. An ambient active
boolean is neither accepted nor forwarded. Caller-supplied reserved
`CARTULARY_TEST_SERVICES_*` state outside the documented public selector and
session-file inputs is `usage_error`. Attachment is permitted only for
`service-backed-test-slice`, `browser-e2e-webserver-backed`, and
`browser-e2e-stateful`; all aggregate, CI, release, measurement, accessibility,
visual, security, drift, and publication commands are owned-only.

Every attached invocation MUST hold a live borrower lease containing owner UID,
run ID, PID/start proof, and session identity while it uses the session. It still
owns a unique suite scope, databases, buckets, ports, runtime roots, browser
contexts, leases, and results, and it cleans only those run-owned resources.
Session down obtains exclusive lifecycle control, refuses any live borrower,
removes only the exact proven session, and succeeds idempotently when that
session is already absent. Session containers use distinct identity and expiry
labels; ordinary run janitors MUST NOT remove a live session, and session cleanup
MUST NOT act on an expired resource without complete identity proof.
Verified by: TH-HARNESS-AC-100

**TH-HARNESS-REQ-817**
`release-inventory-artifacts` is the sole producer of the canonical
`.cartulary/release-artifacts/sbom.cyclonedx.json` and
`.cartulary/release-artifacts/license-report.json` pair. The producer MUST
publish exactly those two mode-`0644` files as one complete reusable artifact
set. `sbom` and `license-report` are non-cacheable validation consumers: direct
invocation first requires the producer, while graph-child invocation relies on
its explicit producer edge and MUST NOT execute target-local production
prerequisites. Each consumer still emits fresh current-run evidence.

Canonical SBOM and license-report bytes MUST be deterministic for their
semantic inputs. The SBOM serial identity is derived from a `sha256:<hex>`
semantic input digest; canonical output contains no generation timestamp,
absolute repository path, commit identity, run-root path, command transcript,
or current-run evidence reference. The current license wrapper identity is
`cartulary.license_report.v2`; superseded wrappers are rejected without an
alias, translator, dual reader, or dual writer. CycloneDX validation remains a
required consumer check. Raw scanner captures, copied license-text trees, and
Markdown diagnostics are temporary producer work, not canonical or retained
release inventory.

Only the paired producer MAY use content-addressed cache-v2 restoration. Its
registered key closes repository source that can change inventory, Go and
workspace manifests and lockfiles, every package manifest, toolchain and
scanner identities, container inputs, output schemas, validators, and the
generator implementation. A hit validates and transactionally restores the
complete typed pair before fresh consumers execute. Missing, partial, surplus,
corrupt, traversing, linked, wrong-mode, wrong-producer, wrong-destination, or
wrong-input records are misses. Vulnerability, publication, security, drift,
service, cleanup, browser, measurement, and destructive work remain
non-cacheable. Go test-dependency discovery MUST use the explicit stable
repository package roots `cmd`, `db`, `internal`, and `tools`; it MUST NOT use a
recursive repository-root pattern that can enter runtime, cache, result, or
test-owned transient directories. Adding a new first-party Go package root
therefore requires an explicit producer and cache-input update rather than
implicit filesystem discovery.
Verified by: TH-HARNESS-AC-101

## 3. Terminology

| Term                     | Meaning                                                                                                                                                    |
| ------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| harness run              | One invocation of a canonical harness command or one child invocation explicitly tied to a result root and run ID.                                         |
| target                   | A named Make target or scheduler target selected by the harness.                                                                                           |
| public target            | A target classified as public in the public command registry and canonical only through Make.                                                              |
| child target             | A target invoked by an aggregate, sequence, scheduler, or wrapper target.                                                                                  |
| work unit                | One scheduler-visible executable unit with an identity, dependencies, resource claims, logs, status, and optional completion keys.                         |
| scheduler                | A harness runner that executes a manifest-defined DAG using logical resource claims and emits scheduler events and summaries.                              |
| lifecycle machine        | A normative finite state contract for one harness lifecycle, including its closed states, closed events, allowed transitions, failure mapping, and evidence. |
| representational lifecycle diagram | A non-normative diagram or list that explains an existing lifecycle without adding requirements.                                                   |
| state                    | One named lifecycle condition inside a lifecycle machine.                                                                                                  |
| event                    | One named input signal that can be presented to a lifecycle machine.                                                                                       |
| transition               | One allowed movement from a source state to a destination state for a specific event and guard.                                                            |
| terminal state           | A state that ends a lifecycle machine instance. No later transition is allowed from a terminal state.                                                      |
| result root              | The root directory that contains run artifacts. The default is `.cartulary/test-results`.                                                                  |
| run ID                   | The run directory name under the result root. The default format is defined in Section 6.                                                                  |
| run root                 | The directory `normalize_result_root(CARTULARY_TEST_RESULTS_DIR) / normalize_run_id(CARTULARY_TEST_RUN_ID)`.                                               |
| harness artifact         | A file or directory produced by a harness run, child command, service, scheduler, test runner, or diagnostic tool.                                         |
| retained artifact        | An artifact preserved after command exit for a specific result root, run ID, and target.                                                                   |
| release-readiness evidence | A retained harness artifact that aggregates product, frontend, visual, accessibility, harness, release, and support evidence for release gating while preserving each record's owner-derived semantic effect. |
| generated artifact       | A file produced from owner inputs by a generator and checked for drift.                                                                                    |
| cache record             | A repo-local JSON record that proves a cache key, input digest, output digest, and profile-specific output contract for a local acceleration profile.      |
| cacheable output         | A deterministic file or directory output whose digest is declared in a cache record and may be reused only when all profile validation succeeds.           |
| non-cacheable side effect | Target behavior that must still execute or be emitted for the current run, such as summaries, failure classification, cleanup, service readiness, or drift/security verdicts. |
| same-run helper artifact ref | A schema-owned retained reference that links an aggregate consumer to helper/setup artifacts produced earlier under the same run root, with declared inputs, producer artifact digests, consumer refs, and fail-closed scope. |
| fixture                  | Test setup state created for a test, package, target, scheduler group, browser stack, or service suite.                                                    |
| service-backed fixture   | A fixture that uses Postgres, object-store services, Docker/testcontainers, browser processes, or Compose-backed services.                                  |
| backing services         | Postgres, object-store services, Docker/testcontainers, Compose services, backend processes, frontend processes, and browser runtime dependencies used by harness targets. |
| output mode              | The resolved mode from Section 7 that controls stdout, stderr, and artifact summary behavior.                                                              |
| machine output           | The `machine` output mode defined in Section 7. For public Make targets that accept it, stdout is exactly one UTF-8 JSON object followed by LF.            |
| failure class            | A coarse normalized grouping for failed harness commands: `product`, `config`, `infra`, `harness`, `artifact`, `timing`, `interrupted`, or `unknown`.      |
| failure reason           | A detailed snake-case reason code used for diagnostics, exit-code mapping, automation, and handoff.                                                        |
| cleanup tier             | A named cleanup scope such as repo-local clean, repo-local distclean, service-suite cleanup, browser-stack cleanup, or stale janitor cleanup.              |
| stale janitor            | A cleanup routine that removes previously generated DBs, buckets, containers, or browser fixtures only when proof predicates match.                        |
| diagnostic-only artifact | An artifact retained for human investigation whose internal shape is not a machine-readable harness conformance contract.                                  |
| harness observability bundle | A deterministic diagnostic projection of native retained harness evidence into one invocation trace, OTLP request payloads, and one hotspot summary. |
| invocation trace         | One trace rooted at a top-level public harness invocation. Child targets, sequence steps, scheduler work, services, runners, and finalizers are spans or links inside that trace. |
| actual dependency critical path | The longest observed dependency-respecting chain of executable and wait intervals. It is distinct from the scheduler envelope stored in `critical_path_wall_duration_ms`. |
| unattributed envelope    | Parent wall time not covered by the union of its directly attributable child intervals; overlapping child time is counted once. |
| test owner               | The one module, platform, application, package, or harness boundary accountable for a catalog row's verification postcondition.                         |
| test family              | An owner-qualified semantic grouping of related catalog rows; it is not a runner, file, target, evidence class, or delivery milestone.                 |
| catalog row              | One active cross-runner executable evidence contract with one owner, exact selectors, verification references, and execution profiles.                 |
| collaborator             | A participating owner that is not accountable for the row's verification postcondition.                                                                 |
| verification contract    | A reviewed machine-readable derivation of one adopted product or support requirement used for evidence routing.                                         |
| runner adapter           | An allowlisted harness implementation that translates one closed selector kind into executable child work and normalized results.                      |
| runtime profile          | A closed startup and managed-service identity referenced by a catalog row.                                                                               |
| resource profile         | A closed set of logical scheduler-resource claims referenced by a catalog row; it does not define physical capacity.                                    |
| fixture capability       | One explicit broker capability referenced by a catalog row and materialized as a run-scoped lease.                                                       |
| fixture profile          | One explicit registry-resolved populated-fixture lifecycle bound to selected verification rows; it is independent of fixture capability.             |
| populated fixture snapshot | One suite-scoped sealed PostgreSQL template reused only as the parent of isolated predicate clones; it is not a cached test result.                    |
| snapshot build artifact  | Immutable redacted proof that one populated template was constructed, validated, connection-closed, and sealed for one snapshot key.                  |
| snapshot lease artifact  | Immutable redacted proof finalized after one predicate clone and its private runtime copy, sessions, processes, database, and bucket are cleaned.       |
| retired delivery identity | Any ordinal execution identity from the predecessor harness, unsupported by the current owner-based runtime.                                           |

Domain and product terms keep their meanings from the product specs and `docs/domain.md`.

### 3.1 Identifier grammar and lifecycle

**TH-HARNESS-REQ-012**
The current identifier grammar is:

```text
segment         = [a-z][a-z0-9_]{0,62}
owner_id        = (module|platform|app|web|package|harness) "." segment
family_id       = owner_id "." segment
row_id          = family_id "." segment
verification_id = owner_id ".verification." segment
```

Each complete identifier MUST contain at most 191 ASCII bytes. Unicode, whitespace, empty segments, `/`, `\\`, percent escaping, and shell metacharacters are invalid. No segment may case-insensitively match `phase[0-9]+` or `fe_p[0-9]+`. Serialized identifier arrays MUST be unique and ordered by ascending ASCII bytes.

Identifiers are immutable, globally unique within their category, and never recycled. An owner migration MUST allocate a new ID and record the migration crosswalk; a runtime alias is forbidden. When display metadata is omitted, diagnostics MUST render the machine ID. Display metadata MUST NOT participate in semantic identity or semantic digests.
Verified by: TH-HARNESS-AC-062, TH-HARNESS-AC-068

### 3.2 Owner, verification, and family contracts

**TH-HARNESS-REQ-013**
The current owner inputs are `contracts/verification/registry.json`,
`contracts/verification/owners/*.json`, `tools/test_catalog_owner.json`, and
`tools/test_families/*.json`. They validate respectively as
`cartulary.verification_registry.v3`, `cartulary.verification_contract.v3`,
`cartulary.test_owner_registry.v1`, and `cartulary.test_family_manifest.v6`.
Every schema uses JSON Schema Draft 2020-12, requires its exact `schema_id`,
rejects unknown properties, and closes every current enum.

Every verification resolves to an active catalog row or registered public
target. Verification v3 contains no requirement, acceptance,
specification-trace, or specification-status field. Specification completeness
is assessed against the adopted owner by human review, not inferred from
routing counts.

Catalog owner, family, row, selector, and runner totals are derived diagnostics.
They MUST reconcile with the active registries and row partitions, but an exact
repository-wide total MUST NOT be separately authored as an acceptance value.
Catalog closure is established by resolved references, unique identities and
selectors, owner and runner partition reconciliation, and semantic digests.

An active test owner registry row MUST contain `owner_id`, `manifest_path`, and
`status="active"`; it MAY contain display metadata. A verification registry
owner row contains only `owner_id` and `contract_path`. Each path MUST be a
normalized repository-relative path under the matching owner root, MUST
resolve exactly once, and MUST remain inside the repository after realpath
resolution. Every active test owner MUST own at least one active executable
row.
Verified by: TH-HARNESS-AC-062, TH-HARNESS-AC-067

**TH-HARNESS-REQ-014**
Every active catalog row MUST contain `row_id`, `owner_id`, `family_id`,
`collaborator_ids`, `verification_ids`, `runner`, `selector`, `evidence_class`,
`runtime_profile_id`, `resource_profile_id`, `fixture_capability`,
`service_dependencies`, `minimum_tier`, `claim_posture`, and `status="active"`.
Rows governed by TH-HARNESS-REQ-812 also contain the required
`fixture_profile_id`; all other rows omit it.

`collaborator_ids` is required and MAY be empty. `verification_ids` is required
and MUST be nonempty. Verification entries contain only routing semantics:
`verification_id`, `behavior_class`, `profile`, `evidence_kinds`, optional
`public_target`, and optional `skip_policy`. Reference arrays MUST be sorted and
duplicate-free, and every reference MUST resolve exactly once. `owner_id` MUST
equal the containing manifest owner. A row MUST NOT embed commands, ports,
capacities, service topology, child environment variables, fixture paths,
documentation paths, or document-derived behavior.
Verified by: TH-HARNESS-AC-062, TH-HARNESS-AC-067

**TH-HARNESS-REQ-015**
Verification contracts use the closed `behavior_class` set `product`, `security`, `architecture`, `build`, `harness`, and `claim_publication`. Catalog rows use the closed `evidence_class` set `unit`, `integration`, `browser`, `accessibility`, `visual`, `measurement`, `static`, `security`, and `release`.

Verification profiles are `base`, `support`, `extension.<profile_id>`, or `claim.<profile_id>`. Claim posture is `implementation`, `informative`, or `claim.<claim_id>`. Product and security behavior MUST use `base` or an adopted extension; build and harness behavior MUST use `support`; claim-publication behavior MUST use an active Core 05 `claim.*` profile. A `claim.*` posture and profile MUST resolve to the same active claim. Informative evidence MUST NOT satisfy conformance or release closure, and informative measurement rows MUST use `minimum_tier=full` or `minimum_tier=release`.

The default skip policy is `forbid`. An authorized skip MUST identify one closed reason, the verification owner, approval evidence, and an RFC 3339 expiry. Missing or expired authorization is equivalent to `forbid`.
Verified by: TH-HARNESS-AC-062, TH-HARNESS-AC-065, TH-HARNESS-AC-069

### 3.3 Runner registry and selectors

**TH-HARNESS-REQ-016**
`tools/test_runner_registry.json` MUST validate as `cartulary.test_runner_registry.v1`. The current runner set and selectors are closed by this table:

| Runner | Exact selector contract |
| --- | --- |
| `go` | One repository package and one nonempty ASCII-sorted array of exact top-level `Test...` symbols. |
| `vitest` | One repository-relative test file and one nonempty ASCII-sorted array of exact full test titles. |
| `playwright` | One repository-relative file, one project ID, one stage, nonempty stable scenario IDs, and matching diagnostic titles. |
| `shell` | One stable registered command ID; raw shell, argv, executable paths, and row-defined environment are forbidden. |

Go selector resolution uses the current ordinary Go-runner build context:
Linux, `amd64`, the repository-pinned Go release tags, the `gc` compiler tag,
and no private or caller-supplied build tags. Filename constraints and
`//go:build` expressions are part of selector resolution. A `Test...` symbol
that exists only in a file excluded by that context has zero resolution.
`runtime_profile_id` selects harness runtime configuration; it does not add Go
build tags. In particular, `cartulary_harness` code is verified through the
declared `server-harness` runtime binary, not through an ordinary Go catalog
row. A future tagged Go-test context requires an adopted runner-registry and
NLSpec revision before a catalog or runner implementation may admit it.

Playwright stage is exactly one of `webserver_backed`, `stateful`, `support`, `visual`, `accessibility`, or `measurement`. Before setup, selector validation MUST reject zero resolution, multiple resolution, overlap between active rows, globs, regular expressions, missing paths, symlink escape, paths outside approved roots, and shell command IDs absent from the task-surface registry.

A later runner requires an adopted runner-registry and NLSpec revision, selector and result schemas, an allowlisted checked-in adapter, and positive and negative contract fixtures. Dynamic package, plugin, or executable loading is outside the current profile.
Verified by: TH-HARNESS-AC-063, TH-HARNESS-AC-067

### 3.4 Runtime, resource, and fixture contracts

**TH-HARNESS-REQ-017**
`tools/execution_topology_manifest.json` MUST own closed top-level
`runtime_profiles`, `resource_profiles`, and `service_resource_minimums`
collections. Catalog rows reference those profile IDs, declare one fixture
capability directly, and declare every consumed managed service. Resource profiles
MAY reference logical resources from `tools/scheduler_resource_registry.json`
but MUST NOT redefine capacity. Unknown profiles or fixture capabilities,
duplicate profiles, cross-kind references, and inline capacity overrides MUST
fail before child work.

These are four orthogonal concepts: runtime profile selects which managed
services are available and their immutable configuration; resource profile
selects base numeric scheduler cost; fixture capability selects the broker lease
and isolation lifecycle; service dependencies enumerate shared managed services
actually consumed. A dependency unavailable from the runtime profile, a
fixture whose required service is absent, or an unknown, omitted, duplicate, or
unsorted dependency MUST fail before execution. No reader may infer a missing
dependency from runtime, runner, target, helper name, or fixture capability.

The current runtime profiles are:

| ID | Managed services required | Contract |
| --- | ---: | --- |
| `none` | no | No managed service or browser startup. |
| `default` | yes | Ordinary unclaimed isolated test-service/browser configuration. |
| `network_flow_claimed` | yes | Network Flow claimed startup configuration with its separately owned key-ring and secret-handling rules. |

The current resource profiles are `standard`, `io_heavy`, `managed_process`,
`backend_capacity_isolated`, `performance_fixture_builder`,
`browser_functional`, `browser_isolated`, `browser_measurement_quiet`, and
`postgres_catalog_isolated`. Every resource profile MUST
bound executable work with positive `cpu`, `io`, `memory_mb`, and `process`
claims; a zero-claim executable profile is forbidden. Their exact claims MUST
be present in the authored topology; omission has no implicit fallback. The fixture capability set is closed by
TH-HARNESS-REQ-804. Work that holds or tests a fixed cluster-scoped PostgreSQL
advisory lock MUST use a dedicated database lease and an exclusive compatibility
key because database identity does not isolate cluster-scoped lock keys.

Direct target execution, exact-row owner slices, service-backed owner slices,
and broad scheduler execution MUST derive the same exact-test and package
fixture capability, service dependencies, resource claims, and semantic digest.
A focused execution path MUST NOT omit, infer, weaken, or replace the selected
row's declarations. Catalog/call-site
mismatch remains a fail-closed harness error.
Verified by: TH-HARNESS-AC-063, TH-HARNESS-AC-065

### 3.5 Browser runtime profiles

**TH-HARNESS-REQ-018**
Browser evidence MUST obtain `runtime_profile_id` through the catalog row and the authored execution-topology registry. Unknown profiles MUST fail before Playwright starts. Arbitrary per-test environment injection and runtime routes that toggle extension claims are forbidden.

A browser runtime profile is immutable startup identity. Generated groups, shards, and sessions MUST carry the profile ID; incompatible profiles MUST use distinct browser session groups. A mixed-profile session, an attach request whose expected profile differs from the retained stack, or a profile/configuration fingerprint mismatch MUST fail before product assertions. Runtime reset is data-only and MUST NOT change the profile, extension claims, key-ring identity, or child-process environment.

Claimed-profile secrets MUST be generated in memory for each owned stack, passed only in the child server environment, and redacted from commands, logs, diagnostics, summaries, retained metadata, and failure messages. Retained browser-stack metadata MUST contain the non-secret runtime profile ID and deterministic configuration fingerprint and MUST NOT contain secret values or secret digests.

Every Make-owned browser invocation MUST resolve its selected rows and their
runtime profiles before service setup. A resolved browser session with one or
more managed services has `service_requirement=test-services` and MUST either
own one isolated suite or attach to an exact compatible active suite. A resolved
session with no managed services has `service_requirement=none` and MUST start
no service suite. Incompatible runtime profiles MUST remain separate sessions,
including when one public target selects more than one profile. Target or stage
names MUST NOT determine service need, and no browser path may fall back to the
shared development Postgres, object store, Compose project, bucket, port, or
proxy.
Verified by: TH-HARNESS-AC-063, TH-HARNESS-AC-066

### 3.6 Semantic digests and extension boundary

**TH-HARNESS-REQ-019**
Semantic JSON digests MUST reject duplicate object members, non-I-JSON numbers, non-finite values, and negative zero before RFC 8785 canonicalization. The digest is lowercase `sha256:` followed by 64 hexadecimal characters. Semantic projections MUST omit display metadata, inert documentation references, diagnostic timestamps, and other fields explicitly classified as non-semantic by the owning schema. A producer and consumer MUST use the same schema ID and semantic projection version.
Verified by: TH-HARNESS-AC-062, TH-HARNESS-AC-066

**TH-HARNESS-REQ-020**
Internal module layout, in-memory data structures, adapter function signatures, human diagnostic prose, and scheduling order among simultaneously ready independent work units remain intentionally unspecified when Sections 4 through 17 are preserved. Implementations MUST remain interchangeable at the public command, schema, artifact, failure, security, and cleanup boundaries.
Verified by: TH-HARNESS-AC-016, TH-HARNESS-AC-071

## 4. Public Command Surface

**TH-HARNESS-REQ-050**
The public command registry MUST be owned by this NLSpec, represented in the authored `tools/task_surface_owner.json`, and mirrored by generated `tools/task_surface_manifest.json` entries with `target_class="public"`. The implementation MUST provide exactly the public targets listed in the target registry below unless the owner input and this NLSpec are revised together.

`tools/execution_topology_manifest.json` MAY provide scheduler topology, child-work topology, generated schedule inputs, or resource-profile inputs. It MUST NOT independently add, remove, rename, reclassify, or change the output class, artifact policy, schema policy, side-effect declaration, command identity, or public lifecycle state of a public target.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-027

**TH-HARNESS-REQ-051**
Every public target MUST declare exactly one output class, exactly one stable-summary schema policy, and exactly one artifact policy. The output-class behavior is owned by Section 7. The schema policy is owned by Section 8.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-004, TH-HARNESS-AC-005, TH-HARNESS-AC-023

**TH-HARNESS-REQ-052**
A Make-owned wrapper MAY invoke package scripts, raw scripts, or external tools as implementation mechanisms. The wrapper remains responsible for the public target's configuration, output, artifact, failure, exit-code, and cleanup contract.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-004

**TH-HARNESS-REQ-058**
A public target MUST provide a semantic harness operation, not merely an alias for one or more child commands. A target qualifies as semantic only when it owns at least one observable behavior from the table below in addition to invoking child work.

| Semantic behavior | Observable requirement |
| --- | --- |
| `configuration_resolution` | Resolves and validates declared harness configuration before child work. |
| `evidence_normalization` | Emits or validates retained artifacts under a stable schema. |
| `failure_normalization` | Maps child or harness failures to Section 9 `failure_class`, `failure_reason`, and public exit code. |
| `service_lifecycle` | Owns service startup, readiness, fixture lifecycle, lease, or cleanup proof. |
| `scheduler_orchestration` | Selects, orders, and executes work units using the scheduler contract. |
| `destructive_safety` | Applies cleanup or reset proof predicates before mutation. |
| `security_boundary` | Applies redaction, token gating, artifact-safety, or secret-handling behavior. |
| `diagnostic_synthesis` | Converts retained evidence into a bounded human or machine diagnostic that cannot be obtained from raw child output alone. |

A target that provides none of these behaviors MUST be private child work or a developer convenience outside the public command registry.
Verified by: TH-HARNESS-AC-020

**TH-HARNESS-REQ-053**
The default `check-harness-smoke` gate MUST remain a small semantic smoke surface rather than a broad harness regression suite. Its fast tier MUST contain exactly one check for each gate role: public Make/wrapper projection, work-graph scheduler semantics, and fixture-broker semantics. Broader field-shape, topology-rendering, and graph-detail checks MUST live in owner-aligned validation such as `json-shape-check`, generated drift checks, the explicit `harness-contract` extended target, or non-default diagnostic smoke tiers. The `harness-contract` target MUST be selected by CI and release gates, not by default local `check`.

Fast smoke fixtures that create disposable repo-shaped files, directories, fake Make surfaces, manifests, or child-run workspaces MUST create them through `tools/harness/test-support/harness-scratch.sh` outside the repository checkout. `CARTULARY_HARNESS_SCRATCH_ROOT` MAY redirect that scratch root only when it still resolves outside the repository. Repo-local `tmp/` remains reserved for durable tool caches, retained run artifacts, and operator-inspectable local outputs; fast smoke fixtures MUST NOT place transient package-shaped or source-shaped scratch trees there, so concurrent source traversal such as `go list ./...` cannot observe disappearing non-package directories.

Non-default harness smoke tiers MAY carry owner-specific regression checks for harness maintenance surfaces such as finalization, evidence audit, topology generation, scheduler behavior, and wrapper compatibility. A retained harness self-test MUST be reachable from an owner-controlled tier, merged into another active owner test, or deleted with named replacement coverage; manual-only harness self-test files outside owner-controlled tiers MUST NOT be treated as active coverage.

The current owner-controlled harness smoke tiers are `fast`, `execution`, `extended`, `lifecycle`, and `full`. Make helper targets MUST expose the active tiers as `run-harness-smoke-fast`, `run-harness-smoke-execution`, `run-harness-smoke-extended`, `run-harness-smoke-lifecycle`, and `run-harness-smoke-full`. The execution helper target is the narrow validation surface for shell command execution, runner wrappers, Make-node dispatch, service-backed runner delegation, and fast Make-sequence wrapper behavior. The lifecycle helper target is the narrow validation surface for browser/dev stack lifecycle, reset, readiness, and teardown harness changes. The execution and lifecycle helper targets MUST remain helper-only and MUST NOT become default local `check` work by themselves.

A negative harness fixture whose assertion depends on injected failing work or on a caller-supplied malformed, empty, or otherwise invalid artifact MUST prevent cache reuse or producer regeneration from bypassing or replacing that exact input. The fixture MUST apply the bypass at the narrowest invocation boundary: a nested graph-backed public Make probe MUST supply its accepted cache-mode selection to that nested public invocation, and an artifact-validation probe MUST make the injected artifact itself non-remakeable rather than enumerate the producer's current prerequisites. A parent-only environment override is not sufficient when an owning wrapper strips public-input variables before launching the fixture.

Every such fixture MUST prove that the injected input reached the intended boundary. A fake executable fixture MUST retain invocation evidence identifying the selected fixture input, and an injected artifact fixture MUST prove the artifact remains byte-identical after validation. Reordering the fixture after a warm successful invocation, adding an ordinary producer prerequisite, or adding a phony readiness prerequisite MUST NOT change the negative fixture verdict.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-006, TH-HARNESS-AC-093

**TH-HARNESS-REQ-054**
The default local `check` gate MUST prioritize correctness evidence and MUST NOT
enforce performance-window acceptance. Performance validation remains available
through the explicit v3 performance commands and the final handoff ladder.
`agent-finalize RESULTS_DIR=<dir>` MAY validate a qualifying retained run but
MUST NOT publish a performance baseline before the complete acceptance window
passes.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-004, TH-HARNESS-AC-006, TH-HARNESS-AC-027

Aggregate placement MUST be selected from `minimum_tier` under
TH-HARNESS-REQ-800. Every selected row retains its verification, evidence
class, runtime profile, resource profile, and fixture capability. Task-surface
inclusion, graph reachability, and target projection membership MUST agree. An
unknown row, inactive row, missing verification, duplicate selection, or tier
selection inconsistent with the closed order MUST fail before child work.

Scheduled Go unit identity MUST close over the exact selected row universe.
Every shard receives a sorted, duplicate-free, non-empty semantic row-ID set;
capture and row-result reconciliation use that same set. A shard from one
selection MUST NOT be reinterpreted against a broader inventory. Partial,
unknown, duplicated, broadened, or inconsistent selection fails before Go
product work or artifact collation.

**TH-HARNESS-REQ-060**
The current graph cache registry is `cartulary.harness_cache_registry.v1`.
Every cacheable unit declares `same_run` or `content_addressed`; every other
unit declares `none`. A target absent from the registry is ineligible. The
retired scheduler-input-stamp and finalizer-action-cache families are not valid
v3 inputs, and current commands MUST NOT read or translate them.

Embedded web asset preparation is a build-artifact producer in the current profile. When the embedded output is consumed by Go `//go:embed`, the publisher MUST update the embedded source-tree artifact atomically from Go embed's point of view. It MUST NOT delete or rewrite a directory of hashed frontend assets in place while concurrently scheduled Go compilation can traverse that directory. Harness-only readiness stamps, cache records, and other operational metadata MUST remain outside the embedded content root unless the owning product spec explicitly makes that file served application content.

Every cached Go binary whose transitive package closure consumes the embedded web asset root MUST depend on the complete embedded asset producer tuple before compilation starts and MUST include that root in its build-artifact cache key. The authored execution topology MUST represent the same dependency whenever the producer and consumer are scheduled together. A source-file-only key, ambient Make ordering, or concurrent publisher/consumer execution is not valid evidence for such a binary.

Every content-addressed key MUST include the cache schema ID, profile ID,
platform identity where relevant, declared tool or runtime versions, declared
command/profile inputs, helper implementation digests, selected source and
configuration digests, and every declared output contract needed by the
profile. Broad timestamp-only caching is invalid. Same-run records exist only
in the current scheduler invocation and cannot be promoted to retained reuse.

Cached Go build-artifact profiles MUST pass `-buildvcs=false`. Git revision and
dirty-worktree stamping are undeclared repository-state inputs and MUST NOT alter a
binary whose cache key is otherwise unchanged. Source-snapshot, release, and audit
provenance remain owned by their retained harness evidence; a cached binary MUST NOT
silently acquire a second provenance identity from ambient Git metadata.

Cache hits MAY skip only the exact registered semantic unit. They MUST still
validate and publish every declared current-run output, emit cache events, and
contribute to the same target projections and aggregate verdict. A cache hit
MUST NOT bypass service readiness, fixture cleanup, runtime reset, destructive
guards, failure projection, or publication of canonical current-run evidence.

Test-service image readiness is a readiness cache profile only. Its cache key MUST include the testservices binary digest, image-owner source digests, helper implementation digests, and toolchain pins. A cache hit MUST still prove that every pinned service image named by the testservices helper is locally present before accepting the readiness stamp; missing images invalidate the stamp and force the ordinary image warmup command. This cache profile MUST NOT replace service startup readiness, test service lifecycle evidence, browser reset, cleanup, fixture preparation, or product-conformance evidence.

Cache records that are missing, bypassed, invalid, corrupt, or whose declared
outputs are missing or digest-mismatched MUST NOT produce success by reuse.
They execute the unit and emit a miss or bypass event, or fail as
`configuration_error` or `artifact_error` when execution cannot safely
continue. The cache mode is `normal`, `cold`, or `off` and is recorded in the
run manifest. Security findings require the database-revision rule in
TH-HARNESS-REQ-807. Stateful, lifecycle, migration, browser, measurement,
cleanup, destructive, drift-verdict, and publication units remain uncached.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-006, TH-HARNESS-AC-018, TH-HARNESS-AC-019, TH-HARNESS-AC-024, TH-HARNESS-AC-028

**TH-HARNESS-REQ-070**
The current `frontend-fallow-static` profile MUST build an effective Fallow configuration from `.fallowrc.json` plus the schema-validated owner input `tools/fallow/reachability_owner.json` using schema ID `cartulary.fallow_reachability_owner.v1`. The effective configuration MUST be retained under the target's run root as `frontend-fallow-static/fallow/resolved-fallowrc.json`, and every Fallow child invocation in the target MUST use that retained configuration. The four authored read-only report invocations (`dead-code` JSON/SARIF, `dead-code` Markdown, `dupes`, and `health`) MUST start as independent concurrent children after the effective configuration is retained. The parent MUST drain every child before normalization, retain every successfully produced artifact when siblings fail, and normalize reports, logs, warnings, and failures in that authored order so concurrent completion order cannot change output identity or primary-failure selection.

Fallow reachability MUST represent durable owner patterns, not per-file suppression growth. Current owner patterns are Vitest setup files selected by `apps/web/vite.config.ts`, task-surface and harness-check backing scripts declared by `tools/task_surface_owner.json`, Vite public asset URLs under the declared public root, and executable tooling dependencies invoked by owner scripts such as `pnpm exec cdxgen`. A missing owner-declared file, missing public asset reference, invalid owner input, or missing executable dependency declaration MUST fail `frontend-fallow-static` as `configuration_error`. Non-blocking static findings in valid Fallow reports MAY still be retained as warnings.

The effective Fallow configuration MUST NOT enable Fallow Runtime, Fallow security scans, baseline enforcement, automatic source mutation, generated-file mutation, inline suppression insertion, or lockfile mutation. A future expansion of changed-code scoping, runtime coverage, blocking enforcement, or additional dynamic reachability owner families MUST revise this NLSpec and the owner schema before implementation.
Verified by: TH-HARNESS-AC-022, TH-HARNESS-AC-028, TH-HARNESS-AC-029

**TH-HARNESS-REQ-056**
The default local `check` graph MUST exclude ordinary browser measurement
evidence. `browser-e2e-measurement` remains an explicit public target and is
selected only by a tier or aggregate policy that requires it.

Default local `make check` MUST NOT schedule current full visual or
accessibility browser work. Direct browser commands and aggregates whose tier
policy selects that work retain full-fidelity target behavior. A bounded
readiness scenario MAY enter `check` only as its own semantic catalog row with
an independently adopted verification and `minimum_tier`.

`test-slice` and `service-backed-test-slice` select visual work only from the resolved owner rows whose evidence class is `visual` and whose Playwright stage is `visual`. The ordinary `browser-e2e-visual` adapter MUST receive the exact resolved row IDs. It MUST NOT broaden the selection to other owners or infer selection from a filename, title prefix, historical registry, or visual fixture identifier.

Browser rows selected by `check` MUST declare `minimum_tier=fast` or
`minimum_tier=standard`, and their verification contract MUST identify the
cross-stack postcondition unavailable from a cheaper required gate.
Informative measurement, full visual, and full accessibility rows MUST use
`minimum_tier=full` or `minimum_tier=release`. Direct public browser targets
remain full-fidelity for their catalog-selected inventories. A bounded row MUST
NOT be represented as proof that a different full target inventory ran.

Browser graph compilation MUST apply the selected tier after evidence-class and
profile resolution and before work-unit emission. An empty browser group is
omitted. A retained group preserves reset, taint, browser-session, teardown,
and target-projection behavior. Direct public browser targets, aggregates, and
owner slices use their own closed selection and MUST NOT inherit another
entrypoint's tier.

Browser target summaries and row results MUST expose the exact selected owner
rows without broadening. A target-level browser run accounts for its catalog-
selected inventory through canonical row results, browser group results, and the
target projection. Absence of one runner family MUST NOT be represented as
success for a selected row that did not execute.

A scheduler- or owner-slice-selected browser group MUST apply its exact sorted
row-ID subset after resolving the generated batch group and before constructing
Playwright title filters. The subset MUST be non-empty, unique, and contained by
the generated group. Reopening a full batch group by stage and group name MUST NOT
discard, broaden, or replace the scheduler-owned row selection.

`cartulary.browser_e2e_batch_manifest.v11` MUST be generated from the active
catalog plus authored stage, runtime, fixture-capability, affinity, isolation,
and evidence policy. Every generated group contains exactly one selector file,
one runtime-derived service requirement, and a sorted non-empty semantic row-ID
set. Delivery phases, title-prefix inference, retired-ID translation, hidden
group loops, and renderer-owned dependency lists are forbidden. Every
applicable Playwright row occurs in exactly one group for its stage and runtime
profile. Task-surface, topology, work-graph, runtime-binary, and readiness
owners are joined without duplicating authority.

For a managed browser session, the suite-scoped browser lifecycle adapter is the
only owner of backend and frontend startup, readiness, startup events,
terminal startup diagnostics, v6 stack publication, and teardown. The
Playwright-facing adapter is attach-only. Before workers start it MUST validate
an exact `cartulary.web_e2e_stack.v6`, the suite/session/profile identities, all
referenced byte digests, the active schema/template/bucket/endpoint identities,
the frontend build digest, and live backend/frontend process proofs.
Missing, stale, v3-only, profile-mismatched, digest-mismatched, development-stack,
or incomplete attachment evidence MUST fail before Playwright assertions.
Canonical Playwright configuration MUST NOT start a web server, reuse an
existing listener, or derive origins from defaults. `--no-deps` MUST NOT bypass
the outer attachment guard.

Each session MUST retain one append-only validated event stream and one
immutable terminal diagnostic under
`_shared/test-services/<suite-id>/browser-sessions/<browser-session-id>/`.
Only the session lifecycle adapter may write them. A ready terminal requires
the complete ordered state graph `initializing`, `service_attached`,
`fixture_ready`, `backend_ready`, `frontend_ready`, `ready`; `failed` may close
any nonterminal state and can never regress. The terminal ready diagnostic MUST
be published before one immutable v4 stack binds its exact digest together with
the compact service-admission proof, lease, database, object-store namespace, fixture,
process, and frontend-build identities. Group and target results MUST carry
ordered run-relative session artifact references and SHA-256 digests. Shared or
multi-profile target projections MUST consume those artifacts without gaining
write authority. V3 stack and v1 diagnostic schemas are historical-validation
inputs only and MUST NOT be active-run admission fallbacks.

The `browser-e2e-stateful` public target MAY use generated `stateful_partition` groups when each partition declares explicit semantic row IDs and an explicit browser session group. Partitioning MUST preserve the same row inventory as the unpartitioned target. An empty adapter invocation MUST be omitted rather than represented as product success. Direct execution MUST reset between selector-file partitions. Scheduler execution MUST serialize stateful partitions that share a browser session group in authored order, and each partition's reset MUST complete before the next partition starts. Distinct browser session groups MAY overlap only when each group owns an isolated retained lifecycle. The `network_flow_claimed` profile MUST always have a distinct startup session for stateful, accessibility, visual, measurement, and webserver-backed evidence. Partitioning MUST NOT remove reset, taint, teardown, route-token, runtime identity, evidence-accounting, or target-summary evidence.

A browser session cleanup unit is a failure-tolerant scheduler finalizer. It MUST run after successful, failed, or dependency-skipped group outcomes, MUST release every retained resource claim owned by that session, and MUST NOT be dependency-skipped. A finalizer whose producer never started and therefore acquired no resource MUST complete as a successful no-op; it MUST NOT invoke lease cleanup, emit a target pass summary, or add a secondary missing-artifact failure. Valid aggregate Go evidence requires the complete declared shard set. After early scheduler stop, an aggregate finalizer for which none or only a proper subset of declared shards started MUST complete as a successful no-op and MUST NOT emit partial target evidence; once every declared shard started, missing shard metadata remains fail-closed under the ordinary artifact classification. Missing browser lease metadata for work that did start likewise remains fail-closed under the ordinary artifact or cleanup classification. A failed browser group MUST still produce the normalized scheduler and target summaries needed to retain its row/group failure classification; cleanup after failure MUST NOT degrade a product assertion into scheduler deadlock, missing-summary inference, or `unknown_failure`.

Cleanup of a proven harness-owned fixture namespace MUST be idempotent. After
the ordinary ownership and destructive-safety predicates succeed, a canonical
database-, bucket-, prefix-, or object-absent result is a successful terminal
no-op, including when another finalizer already removed the same owned
namespace. Authentication, authorization, endpoint, ownership, malformed
response, and other cleanup failures remain fail-closed. Repeated finalization
MUST NOT convert successful cleanup into `cleanup_error` merely because the
owned resource is already absent.

The Playwright result adapter MUST join observations by exact normalized selector file and exact catalog title. Zero observations, multiple observations, aggregate process success without selector observations, and an unauthorized Playwright skip are accounting failures; a product assertion failure MUST remain a product failure. Reports, stdout, and stderr MUST be retained through the redaction boundary, and only exact selector observations may close catalog rows.

Ordinary Playwright configuration MUST set `updateSnapshots=none`. Missing
goldens therefore fail ordinary visual validation. Snapshot mutation is
authorized only by `browser-e2e-visual-update`, which MUST retain the same row
selection, profile-derived service resolution, session grouping, attachment
evidence, and exact row accounting as validation and MUST NOT publish ordinary
passing visual evidence.
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-022, TH-HARNESS-AC-063, TH-HARNESS-AC-069

**TH-HARNESS-REQ-057**
Harness performance is implementation-health evidence, not product-performance
or claim-publication evidence. Acceptance uses the matched observation windows
in TH-HARNESS-REQ-377 through TH-HARNESS-REQ-380. No fixed global duration cap,
fixed lane-balance ratio, or phase-target duration is a v3 requirement.

Accepted observations MUST distinguish forced-cold provisioning, a discarded
warm-up, and warm steady-state execution. Tool installation, Go build-cache
population, frontend or Playwright install, service image preparation, and
helper builds remain valid correctness work but are first-class graph units and
are not hidden inside a measured leaf. Tier-selection reduction is reported
separately from unchanged-workload executor improvement.

Finalizer duration and mutations are outside the evaluated source window. A
performance claim after a generated or planning-input refresh MUST use a later
complete window from the resulting source state. Failed, incomplete,
contaminated, mixed-semantic, or unstable windows block publication and MUST
NOT be relabeled or overwritten by a retry at the same source state.
Verified by: TH-HARNESS-AC-018

### 4.1 Mechanism Boundary

Harness implementation packages are owner-local realization details unless a row in this NLSpec explicitly promotes a specific path. Module names and directory names under `tools/harness/**` MUST NOT be treated as public harness invocation bindings, product conformance surfaces, or durable compatibility promises. Implementation code MUST depend on owner facades for contract, catalog, output, execution, finalization, diagnostics, smoke, frontend, browser, backend, scheduler, evidence-accounting, and generated-artifact behavior instead of importing private mixed-responsibility helper paths. The harness import-boundary check MUST reject private catch-all package imports once an owner-specific facade exists for the behavior.

Catalog shape, registry semantics, evidence naming, fixture policy, row selection, selector verification, and Go/Vitest/Playwright/shell result adaptation are catalog and evidence-accounting behavior. Task guidance, target explanation, owner explanation, retained-run explanation, target-plan display, and fixture diagnostics are diagnostics behavior. Runner-specific discovery remains behind runner adapters; test-output indexing and target-start statistics belong to test-output helpers; summary topology belongs to execution helpers; task-surface validation and generated Make rendering belong to generated-artifact helpers. Public Make target names, stable `command_id` values, schema IDs, output/failure contracts, and retained artifact paths remain the compatibility surface. A diagnostic surface that combines catalog, runner, task-surface, and scheduler reachability data MUST depend on the owner facades for each source family rather than importing private owner-local helpers.

### 4.1A Private Helper Ownership, Facades, and Compatibility

**TH-HARNESS-REQ-061**
Private helper paths under `tools/harness/**` MUST be classified by semantic ownership, not by historical directory name. A helper path MAY move, be renamed, or be deleted when public Make target names, stable `command_id` values, schema IDs, output shape, retained artifact paths, failure mapping, cleanup behavior, and declared public input contracts remain unchanged.

The following compatibility statuses are closed:

| Status | Definition |
| --- | --- |
| `public_contract` | A Make target, `command_id`, schema ID, output/failure contract, retained artifact path, or declared public input contract owned by this NLSpec. |
| `owner_facade` | A declared semantic import boundary for an implementation helper family. Non-owner callers MUST import this facade when the behavior is needed from another harness subsystem. |
| `unsupported_private` | A helper path that is not a public compatibility path and may move, be renamed, or be deleted without public contract revision when public harness behavior is preserved. |

The following helper ownership registry is closed for the current profile:

Repository attachment mechanics for these rows are owned by the
schema-validated `tools/harness_helper_ownership.json` input. Its `key` set
MUST exactly match the semantic `owner_facade` rows below. Each entry declares
only current facade paths and explicit cross-owner consumers; historical paths
and moved-module aliases are not owner input.

The scheduler helper rows below own harness orchestration only. They MUST NOT define product HTTP routes, WebSocket behavior, workbook mutation/query behavior, saved-view behavior, product transport adapters, frontend shell state, or grid-vendor integration; those questions remain owned by Core 00 through Core 04 and adopted subsystem owner documents.

| Helper family | Facade key | Owner boundary | Compatibility status | Observable behavior to preserve |
| --- | --- | --- | --- | --- |
| Backend Go target row discovery and backend target-plan DTOs. | `backend_target_plan` | Backend harness target planning. | `owner_facade` | Backend row identity, ordering, target selection, backend field values consumed by `target-plan-json`, and Go fixture-policy details. |
| Backend Go shard planning and shard-plan DTOs. | `backend_shard_plan` | Backend harness shard planning. | `owner_facade` | Shard names, ordering, fixture policy, target mapping, and shard-plan JSON. |
| Backend Go selection, compatible shard planning, execution, and row-result adaptation. | `backend_target_execution` | Work-graph compiler plus row runner. | `owner_facade` | Exact row selection, capacity-derived LPT planning, runtime-binary validation, Go JSON outcome adaptation, canonical row results, and failure classification. |
| Go estimated-work planning and accepted performance observations. | `backend_duration_accounting` | Work-graph cost model plus canonical performance projection. | `owner_facade` | Estimated-work identity, exact selected-row closure, event-derived observations, and accepted-window rules. |
| Migration-history and schema-object ownership validation. | `database_contract_drift` | Generated-drift `database_contract_drift` sub-boundary. | `owner_facade` | Manifest schema IDs, current-line lineage validation, deterministic diagnostics when present, fresh and penultimate scratch migration behavior for `migration-drift`, and `json-shape-check` compatibility. |
| Govulncheck findings normalization and redaction. | `static_analysis_security_findings` | Static-analysis/security sub-boundary. | `owner_facade` | Redaction behavior, blocking/non-blocking classification, exit-code mapping, `cartulary.govulncheck_findings.v1`, and `GOVULNCHECK_DB` as the current public vulnerability database override. |
| Task-surface owner loading, topology, generated drift, schema attachment validation, and generated Make rendering. | `generated_artifact_surface` | Generated-artifact helpers. | `owner_facade` | Authored task-surface ownership, generated artifact source-of-truth parity, schema validation, drift failure mapping, and no hand-editing of generated outputs. |
| Public Make Node-tool command registry, input filtering, child environment construction, and invocation argument synthesis. | `command_surface_node_tool_dispatch` | Command-surface helpers; current facade `tools/harness/command-surface/make-node-tools.mjs`. | `owner_facade` | Public Make node-tool input matrix parity, inherited-environment stripping, runtime-env forwarding, usage diagnostics, and generated Make dispatch behavior. |
| Shared shell command runtime, timing, redaction, artifact directory policy, output mode resolution, step summaries, target summaries, and Vitest watchdog substrate. | `command_execution_runtime` | Execution runtime helper boundary. | `owner_facade` | Step and target timing spans, redacted stdout/stderr handling, retained artifact directories, output mode behavior, step-summary emission, public exit-code propagation, and watchdog sidecar behavior. |
| Cross-runner owner catalogs, row IDs, selectors, profiles, and slice planning. | `test_catalog` | Catalog and selection boundary. | `owner_facade` | Owner/verification/runner registry validation, exact selector resolution, profile resolution, catalog semantic digest, and immutable slice-plan semantics. |
| Owner row-result generation and validation. | `test_evidence_accounting` | Canonical row-result and target-projection boundary. | `owner_facade` | `cartulary.harness_row_result.v2`, terminal row closure, exact selected-row scope, and target-summary evidence references. |
| Owner retained-evidence audit. | `test_evidence_audit` | Canonical evidence-audit boundary. | `owner_facade` | `test-evidence-audit` retained-root inputs, semantic-digest validation, roster/count/sum closure, and rejection of historical evidence schemas. |
| Frontend unit target execution. | `frontend_target_execution` | Execution helper boundary. | `owner_facade` | `frontend-unit` command behavior, Vitest invocation shape, owner summaries, target summaries, exact selected-row filtering, and public failure mapping. |
| Vitest execution diagnostics and sidecars. | `vitest_execution_diagnostics` | Execution, test-output, and diagnostics helper boundaries. | `owner_facade` | Vitest row wrappers, exact-title filtering, watchdog handling, `cartulary.vitest_failure_details.v1`, and retained `vitest-failure-details.json` sidecar paths. |
| Frontend toolchain and dependency-install readiness. | `frontend_toolchain_readiness` | Readiness helper boundary. | `owner_facade` | Pinned repo-local Node/pnpm preparation, frozen-lockfile install behavior, readiness cache keys, stamp content, and configuration-error exit mapping. |
| Go effective-toolchain selection and Go-tool installation readiness. | `go_toolchain_readiness` | Readiness helper boundary. | `owner_facade` | Exact effective Go selection, read-only diagnosis, corruption classification, configuration-error exit mapping, private failure-atomic installation staging, and version-scoped external-cache diagnostics. |
| Web build and embedded web asset artifacts. | `web_build_artifact` | Readiness/build-artifact helper boundary. | `owner_facade` | `build-web` and embedded asset cache behavior, Vite build invocation, embed archive/stamp atomicity, and public build target behavior. |
| Design-token generation. | `design_token_generation` | Generated-artifact design-token sub-boundary. | `owner_facade` | `contracts/design/tokens.v1.json` machine-registry loading and validation, generated token TypeScript content, machine-input provenance identity, atomic replacement, and generated-artifact drift behavior; executable helpers MUST NOT access `docs/`. |
| Font asset validation. | `font_asset_validation` | Static-analysis helper boundary. | `owner_facade` | Font manifest validation, vendored font checksum/license checks, local CSS activation checks, and remote-font ban diagnostics. |
| Browser batch manifest loading, normalization, and target/stage metadata. | `browser_batch_manifest` | Browser manifest helper boundary; current canonical path `tools/harness/browser/browser-batch-manifest.mjs`. | `owner_facade` | `cartulary.browser_e2e_batch_manifest.v11`, catalog-derived semantic groups, exact selector files and row IDs, fixture capabilities, affinity, and target/stage metadata. |
| Browser command/dependency projection. | `browser_scheduler_adapter` | Work-graph browser compiler boundary; current facade `tools/harness/scheduler/work-graph/browser.mjs`. | `owner_facade` | Browser group, reset, finalizer, affinity, lease, evidence, and producer dependency units. |
| Scheduler DAG execution, events, finalizers, summaries, and failure mapping. | `scheduler_execution_core` | Work-graph scheduler boundary; current facade `tools/harness/scheduler/work-graph/index.mjs`. | `owner_facade` | Deterministic admission, resource fitting, aging, event emission, cleanup, cache integration, failure projection, and canonical run artifacts. |
| Scheduler manifest, graph, resource, and capability helpers. | `scheduler_contract_helpers` | Current facades `tools/harness/scheduler/scheduler-manifest.mjs`, `tools/harness/scheduler/scheduler-resources.mjs`, and `tools/harness/scheduler/work-graph/model.mjs`. | `owner_facade` | `cartulary.scheduler_manifest.v3`, work-graph identity, resource-registry validation, capability snapshots, and canonical artifact references. |
| Scheduler child process execution and log handling. | `scheduler_process_execution` | Current facade `tools/harness/scheduler/work-graph/executor.mjs`. | `owner_facade` | Child environment construction, process-group cancellation, log capture, timeout, and child failure propagation. |
| Owner-slice plan construction and selected work-unit accounting. | `test_slice_planning` | Work-graph compiler boundary. | `owner_facade` | Owner and exact-row selection, canonical target plan, selected unit ordering, fixture capability, runtime-binary readiness, and target projections. |
| Service-backed fixture and lifecycle planning. | `service_backed_schedule_planning` | Run-scoped fixture-broker boundary; current facade `tools/harness/scheduler/fixture-broker/index.mjs`. | `owner_facade` | Explicit capability leases, pool replacement, contamination handling, borrowed-resource ownership, cleanup, and graph dependency integration. |
| Event-derived performance accounting. | `scheduler_duration_accounting` | Canonical performance boundary; current facade `tools/harness/observability/canonical-performance-cli.mjs`. | `owner_facade` | Complete observation windows, variability rules, contamination rejection, matched comparisons, and baseline publication eligibility. |
| Scheduler retained event/timing diagnostics. | `scheduler_evidence_drift` | Diagnostics helper boundary; current entrypoints `tools/harness/diagnostics/scheduler-event-order-drift-cli.mjs` and `tools/harness/diagnostics/scheduler-summary-timing-drift-cli.mjs`. | `owner_facade` | Event-order drift, summary timing drift, warm-run eligibility checks, lane/session accounting diagnostics, bounded output, and retained-run evidence requirements. |
| Browser target execution and group dispatch. | `browser_target_execution` | Browser graph and single-group runner boundary. | `owner_facade` | Target selection, group execution, broker-owned stack attachment, resets, canonical group/row results, cleanup, and public failure mapping. |
| Browser Playwright selection, webserver-batch execution, report parsing, and selection artifacts. | `browser_playwright_execution` | Browser Playwright execution plus test-output adapter boundary; current stable adapter `tools/harness/output/test-output/playwright-artifacts.mjs`. | `owner_facade` | Exact catalog row and scenario selection, Playwright runner report interpretation, selected-test title/file indexing, merged report behavior, owner summaries, stdout/stderr/output artifact paths, and failure normalization. |
| Browser estimated-work and performance accounting. | `browser_duration_accounting` | Work-graph cost model and canonical performance boundary. | `owner_facade` | Exact group and row identities, event-derived intervals, readiness attribution, accepted performance windows, and no hidden shell-batch timing. |
| Browser owned-stack lifecycle, runtime identity proof, and reset controller. | `browser_lifecycle_adapter` | Browser lifecycle/test-route adapter boundary; current entrypoints `tools/harness/browser/start-web-e2e.sh` and `tools/harness/browser/reset-web-e2e-stack.sh`. | `owner_facade` | `cartulary.web_e2e_stack.v6`, `cartulary.web_e2e_backend_generation.v1`, per-session startup event and terminal-diagnostic ownership, immutable attachment evidence, preview-mode startup, port ownership, runtime root/session files, process-group cleanup, runtime identity proof, backend replacement, reset diagnostics/taint, and Playwright state cleanup. |
| Browser accessibility evidence summaries. | `browser_accessibility_evidence` | Browser helper boundary; current canonical path `tools/harness/browser/browser-catalog-group-cli.mjs`. | `owner_facade` | Accessibility summary schema, contrast record handling, retained Playwright runner references, and browser a11y target artifact paths. |
| Browser visual snapshot update helper. | `browser_visual_update_helper` | Work-graph browser compiler plus current single-group entrypoint `tools/harness/browser/browser-catalog-group-cli.mjs`. | `owner_facade` | Helper-only visual update target posture, snapshot-update mode propagation, authorized authored snapshot write path, retained browser evidence, and exclusion from default `check`, `test`, `ci`, and release gates unless separately declared. |

Migration-history and schema-object ownership validators are implementation-support evidence for database-contract drift. They MUST NOT become Core product behavior owners and MUST NOT be imported by backend target execution code except through a declared `database_contract_drift` facade.

Govulncheck findings normalization belongs to `static_analysis_security_findings`. Non-current backend helper paths for that behavior are private implementation history and are invalid as current compatibility paths.

Verified by: TH-HARNESS-AC-038, TH-HARNESS-AC-040, TH-HARNESS-AC-041

**TH-HARNESS-REQ-062**
Private helper compatibility MUST NOT be preserved by default. Archive references, historical handoffs, raw script paths, and old implementation imports do not establish compatibility support.

The current profile rejects private cross-owner imports through semantic
boundary rules derived from current ownership. The validator MUST retain the
general bans on `tools/harness/core/**`, unknown top-level harness owner roots,
private backend, browser, frontend catch-all, catalog, and evidence-accounting imports,
and scheduler-to-browser imports that bypass the browser scheduler adapter.
It MUST NOT maintain exact or prefix tombstones for historical paths.

Deletion or contraction of an `unsupported_private` family is allowed only
after current live callers are moved to declared owner facades,
generated/task-surface metadata does not reference the old path, semantic
import-boundary tests reject cross-owner access, and characterization tests
for the relevant public targets pass. A moved private module receives no
forwarding shim unless a continuing external consumer is demonstrated.

Verified by: TH-HARNESS-AC-039

**TH-HARNESS-REQ-063**
Harness implementation code MUST import declared owner facades rather than arbitrary private backend or frontend helper paths once the facade for that behavior exists. The harness import-boundary check MUST reject new non-owner imports from private implementation helpers and private catch-all imports where an owner facade is declared.

Tests MAY import private implementation fixtures only from declared test-support paths. Generated files MAY mirror declared owner paths but MUST NOT independently widen import allowances. Non-owner harness code MUST NOT import private browser implementation helpers directly once a browser owner facade exists; scheduler code MUST use the scheduler browser adapter instead of direct browser helper imports. Import-boundary failures are harness failures, not product failures.

`tools/backend_module_boundaries.json` MUST validate as
`cartulary.backend_module_boundaries.v2`. Every forbidden-import rule MUST
declare `match_kind` as exactly `exact` or `subtree`. `exact` matches only the
named import path. `subtree` matches the named import path and paths below it
on a package-segment boundary. The checker MUST NOT infer prefix semantics
from the spelling of a path. Diagnostics MUST identify the rule, match kind,
importer, and candidate import. Non-current boundary schemas and rules without
`match_kind` are unsupported. Source-table access policy remains file-exact;
directory expansion is not an accepted migration.

Verified by: TH-HARNESS-AC-038, TH-HARNESS-AC-039

**TH-HARNESS-REQ-064**
Go duration-baseline maintenance MUST use the retained-run eligibility declared for each command family in the following table. `duration_retained_run` means retained run evidence that contains the target summaries, scheduler summaries, scheduler events, and Go shard/duration artifacts required to bind observed durations to current target and shard identities.

| Command family | Retained-run rule |
| --- | --- |
| `go-test-duration-baseline-coverage` | Does not require retained-run evidence. It verifies planned baseline coverage only, is read-only, and may remain ordinary `check` evidence. |
| `go-test-duration-baseline-drift` | MAY use `RESULTS_DIR` explicitly supplied by the caller or the current retained run when invoked inside a retained-run context. It MUST be read-only and MUST reject missing, failed, incomplete, or artifact-insufficient evidence before producing a drift verdict. |
| `go-test-duration-baselines` | MUST require explicit `RESULTS_DIR`. It MUST reject ambiguous result roots unless exactly one retained run is resolved by Sections 5 and 6. It MAY mutate baseline files only after retained-run validation succeeds. |
| `agent-finalize duration_baseline_refresh` | MUST use an existing retained full warm `make check` run root. Service-backed-only, owner-slice, browser-only, and other partial roots are invalid. |
| `agent-finalize duration_baseline_drift_validation` | MUST use the same retained-run requirement as `duration_baseline_refresh`, but remains read-only. |

Duration retained-run evidence MUST be rejected if it is failed, incomplete, contaminated, non-warm where warm evidence is required, missing full-check markers where full-check evidence is required, artifact-insufficient, or older than the latest sibling retained check run without the Section 8.2 older-run override where that override applies. A mutating baseline update MUST fail before the first mutation when retained-run validation fails.

Verified by: TH-HARNESS-AC-042

**TH-HARNESS-REQ-065**
Migration-history evidence capture is database-contract or migration-evidence evidence unless a later adopted owner explicitly promotes another boundary. Harness rows for migration manifest audit, embedded SQL source audit, goose ledger inspection, schema-object ownership drift, and migration-history diagnostics MUST route through database-contract or migration-evidence ownership. They MUST NOT be classified as operator-recovery conformance evidence merely because an implementation exposes a deployment-local operator wrapper.

A deployment-local wrapper for migration evidence MAY exist as an implementation mechanism. Omission behavior: if the wrapper is absent, harness conformance MAY still be satisfied through owner-backed database-contract drift or migration-evidence targets, provided the catalog rows cite the correct owner and retained artifacts satisfy the declared schema.

Verified by: TH-HARNESS-AC-045

**TH-HARNESS-REQ-066**
When production code or tests move across packages, harness accounting source-of-truth inputs MUST be updated before generated artifacts. The required order is:

1. Update the owner catalog row, target-map row, runtime-binary declaration, helper ownership row, task-surface input, topology input, or schedule input that owns the changed path or selection.
2. Regenerate downstream task-surface, schedule, topology, and evidence-accounting artifacts only through Make-owned generation.
3. Run drift and schema checks before treating the move as complete.
4. Treat any generated-ledger or generated-manifest hand edit as non-conformant.

A path-only test move that preserves row ID, title, target, owner references, evidence class, runtime-binary use, and retained artifact shape MAY remain behavior-preserving. A move that changes command grammar, result schema, authorization outcome, runtime-binary use, public target membership, minimum tier, graph topology, fixture lifecycle, or retained artifact shape is a public harness behavior change and MUST be specified in this NLSpec before implementation.

Required validation after moved-test accounting changes is:

| Change class | Required validation |
| --- | --- |
| Catalog or row ownership only | `make json-shape-check`, private `test-catalog-check`, and the affected owner slice. |
| Task-surface or public-target metadata | `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all` and public-target parity checks. |
| Scheduler topology or generated schedule | `make generate`, `make generate-drift`, and the affected scheduler target. |
| Generated artifacts | `make generate-drift` and `make generated-artifact-policy-check`. |
| Operator runtime-binary rows | `make build-operator` and the affected scheduler-selected operator work. |

The current raw support-package execution bindings are:

| Support package | Raw aggregate | Public target | Fixture capability | Required isolation |
| --- | --- | --- | --- | --- |
| `internal/testutil/processtest` | `backend-process-processtest` | `backend-process` | `managed_process` | Self-contained inherited-FD-3 helper; no product services, product configuration fixture, or external product server binary. |
| `internal/testutil/suiteservices` | `backend-unit-suiteservices` | `backend-unit` | `none` | Pure unit execution with no managed service or managed-process lease. |

These raw aggregates are support evidence. They MUST NOT close product
conformance or replace exact catalog-row evidence.

Verified by: TH-HARNESS-AC-046

**TH-HARNESS-REQ-067**
Private backend test-support packages MUST be named for their semantic owner
and responsibility. Historical delivery labels are forbidden as evidence-accounting
identifiers, test IDs, catalog selectors, helper packages, exported helper types,
and exported helper symbols. A legitimate product-phase or execution-step term
requires the Section 15 semantic allowlist classification.

The repository MUST maintain one schema-validated test-support inventory that
classifies every shared and owner-local Go support root and every committed
shared fixture or golden root. The inventory MUST record semantic owner,
shared or owner-local posture, runtime-scan treatment, support-scan treatment,
and whether a Go root exposes service-starting `Start*` entrypoints. A support
root that is absent, duplicated, missing on disk, excluded from both runtime
and support security profiles, or inconsistent with its service-starting
classification MUST fail harness validation. Private compatibility packages
MUST NOT be retained solely to preserve an old helper import.

Verified by: TH-HARNESS-AC-056

### 4.1B Command-runtime profiles and legacy-surface disposition

**TH-HARNESS-REQ-068**
Cartulary has exactly three deployable executable identities: `server`, `migrate`,
and `operator`. `build-server-harness` is a harness-only build profile of the
existing `server` identity, not a fourth deployable, release artifact, or public
product command. `build-server` produces the production profile; it MUST NOT link
harness-route contributors or inherited-listener support.

The harness profile is selected only by the Make-owned build target and the private
`cartulary_harness` build tag. It is the only profile permitted to consume
`CARTULARY_ENABLE_TEST_ROUTES` or `CARTULARY_HTTP_LISTEN_FD`. The production profile
MUST reject either key before application runtime construction or listener
acquisition. Product HTTP, WebSocket, authorization, diagnostics, and packaged-asset
behavior shared by the two profiles remain owned by Core 00 through Core 04.
Cataloged Go tests MUST exercise this profile only as black-box consumers of the
declared `server-harness` runtime binary. The private build tag is not an
ordinary Go-test selector context.

All black-box consumers MUST receive a declared runtime binary from the topology
runtime-binary registry. The injected file MUST be a scheduler-produced regular,
executable, digest-matched artifact for its declared profile. Go tests MUST NOT build
or run a command binary through nested `make`, `go build`, or `go run` fallbacks;
missing injection is a configuration failure that directs the caller to the relevant
public Make target.

| Surface | Owner | Supported caller and final execution surface | Retirement trigger |
| --- | --- | --- | --- |
| `server` production profile | `internal/app/server` plus `internal/platform/httpruntime` | Development, packaging, release, stand-up, and deployable-shape targets through `build-server`. | None; this is the production server identity. |
| `server` harness profile | Testing Harness and the `internal/app/server` harness contribution | Process and browser evidence through `build-server-harness` and declared `server-harness` runtime-binary rows only. | Remove when no harness route or inherited-listener consumer remains. |
| `migrate up` | `internal/app/migrate` and Postgres migration owner | Deployment/bootstrap, `db-migrate`, `db-reset`, and Make-owned migration targets only. | None while forward migration remains a deployment requirement. |
| Penultimate migration application | Postgres database-contract test support | `migration-scratch-apply` and migration-drift support only; never the `migrate` executable. | Remove when migration-line verification is retired by its owner. |
| Five recovery commands | Recovery module | Declared operator runtime-binary rows and deployment-local operator execution. | None while Core 01 requires them. |
| `migration-evidence capture` | Postgres migration-evidence owner | Declared database-contract or migration-evidence harness rows. | Remove when owner-defined evidence no longer consumes it. |
| `object-store init` | Object-store platform owner | Stand-up packaging and declared operator rows. | Remove when configured-bucket initialization is no longer needed. |

Retired recovery aliases, implicit migration commands, arbitrary Goose arguments,
contextless runner wrappers, and raw-Go or nested-build test fallbacks are unsupported
private compatibility. Historical archives may describe them but MUST NOT restore them
as current workflows.

Verified by: TH-HARNESS-AC-057

Private child-runner paths are implementation details and MUST be registered through the checked-in runner registry and invoked through Make-owned execution adapters. Legacy root runners, frontend catch-all runners, and core diagnostic shims MUST NOT be recreated as compatibility paths; callers MUST use the owning execution, browser, or diagnostics boundary.

`tools/harness/execution/cartulary-runner-cli.mjs` private direct use MUST select an explicit runner subcommand. Backend Go target execution is available only through `go-target <target-or-command> [...]`; direct aliases such as `backend-unit` or `backend-store` are unsupported private compatibility and MUST fail with usage status `2`. Quiet successful child logs MUST remain suppressed for public summaries regardless of unrelated environment variables.

Harness import-boundary validation MUST cover statically resolvable shell `source` and `.` references to repo-local `tools/harness/...` paths in addition to JavaScript imports. Unsupported private helper rules apply equally to those shell-source edges. Dynamic shell source expressions that cannot be resolved to a repo-local static path remain outside this static rule and are covered by shell lint plus target-level tests.

| Surface                                                  |                                  Normative? | Required contract                                                                      |
| -------------------------------------------------------- | ------------------------------------------: | -------------------------------------------------------------------------------------- |
| Public Make target name                                  |                                         yes | Stable command surface invoked as `make <target>` from the repository root.            |
| `tools/task_surface_owner.json` public target metadata | yes | Required authored machine owner downstream of this NLSpec. |
| `tools/task_surface_manifest.json` public target_class | yes | Required generated machine-readable mirror of the public target registry. |
| Root/package `pnpm` scripts                              |                                          no | Developer convenience unless invoked by a Make-owned public target. Successful raw package-script output MUST NOT be reported as completion evidence for public harness targets. |
| Raw owner helper scripts and child CLIs                   | no | May change when public Make behavior remains unchanged.                                |
| Make-owned harness contract helper implementation path    | no as an additional invocation binding; yes as the owner implementation path used by Make wrappers | Make-owned public wrappers invoke this owner CLI for preflight, cleanup, schema validation, and related harness-contract mechanics. Direct CLI invocation remains implementation support unless a public Make target adopts it. |
| Make-owned test-output helper implementation path         | yes only for Make-owned command behavior; no for JavaScript module boundaries | Make-owned wrappers may invoke these helpers to emit lifecycle lines, machine output, summaries, and retained artifacts. Command names, arguments, output modes, schema IDs, failure taxonomy, exit codes, and retained artifact paths are the contract; helper filenames, exports, imports, and private module locations are implementation details unless explicitly promoted by this NLSpec. |
| `tools/testservices` binary path                         |                                          no | Service lifecycle behavior is normative; binary path is an implementation realization. |
| Public output classes and schema IDs listed in Section 8 |                                         yes | Required machine-output and artifact validation contracts.                              |
| Docker image tag for Postgres or object-store services   | no unless declared in a service fixture row | Exact tag is not normative unless it defines fixture semantics in Section 11.          |
| Generated Make include names, helper binaries, helper target classes, priority-band names, and generator constants | no | Implementation detail unless promoted by an explicit requirement.                      |

**TH-HARNESS-REQ-069**
Task-surface Make binding profiles are private generated-artifact implementation
details. The current closed profile types are `artifact_binding`, `aggregate`,
`readiness_projection`, `cleanup`, `print_help`, `sequence`, `check_schedule`,
`go_target`, `service_backed_target`, `service_backed_schedule`, `browser_batch`,
`owner_command`, `summary_target`, and `node_tool`. The former catch-all `alias`
profile is unsupported. Generated bindings MUST factor invariant preflight,
prerequisite, input sanitization, and summary behavior through a shared generated
runtime rather than repeating the global public-input inventory per target.

`tools/task_surface.generated.mk` MUST be no larger than 180 KiB; the shared generated
runtime include MUST be no larger than 64 KiB; their combined size MUST be no larger
than 220 KiB; and neither file may contain a physical line longer than 512 bytes. A
synthetic addition of 25 ordinary owner-style targets MUST grow generated output by no
more than 512 bytes per target on average. These are implementation-maintainability
gates and do not change public target behavior.
Verified by: TH-HARNESS-AC-058

**TH-HARNESS-REQ-071**
The current owner-first public commands are closed by this table:

| Target | Command ID | Required target-local inputs | Optional target-local inputs |
| --- | --- | --- | --- |
| `test-slice` | `cartulary.harness.command.test_slice.v2` | `OWNER` | `ROWS`, `VITEST_MAX_WORKERS`, `PLAYWRIGHT_WORKERS`, `JSON` |
| `service-backed-test-slice` | `cartulary.harness.command.service_backed_test_slice.v2` | `OWNER` | `ROWS`, `VITEST_MAX_WORKERS`, `PLAYWRIGHT_WORKERS`, `JSON`, `CARTULARY_TEST_SERVICES_MODE`, `CARTULARY_TEST_SERVICES_SESSION_FILE` |
| `explain-test-owner` | `cartulary.harness.command.explain_test_owner.v2` | `OWNER` | `JSON` |
| `task-guide` | `cartulary.harness.command.task_guide.v2` | `ROLE=module-author`, `OWNER` | `JSON` |
| `test-evidence-audit` | `cartulary.harness.command.test_evidence_audit.v3` | `OWNER`, `EVIDENCE_ROOTS_FILE` | none |

`test-catalog-check` is private check-level work, accepts no target-local public input, has no stable public command ID, and MUST be selected by `make check`. Any predecessor delivery-phase command, alias, reader, or dual writer is unsupported.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-064, TH-HARNESS-AC-070

**TH-HARNESS-REQ-072**
`test-slice` and `service-backed-test-slice` MUST use `scheduler_summary_with_artifacts`; `test-evidence-audit` MUST use `summary_with_artifacts`; `explain-test-owner` and `task-guide` MUST use `human_summary` and remain read-only. Section 7 owns the exact human, machine, and target-local JSON behavior. Section 8 owns the exact artifacts.
Verified by: TH-HARNESS-AC-004, TH-HARNESS-AC-064, TH-HARNESS-AC-071

**TH-HARNESS-REQ-073**
The public commands, machine contracts, selection inputs, artifact paths, and schema IDs declared by this specification are closed. Any undeclared predecessor identity is unsupported and MUST NOT be accepted, translated, discovered, read, or emitted. Historical migration mappings belong only in the immutable reconciliation handoff and Git history; they are not normative runtime inputs.
Verified by: TH-HARNESS-AC-068, TH-HARNESS-AC-070

**TH-HARNESS-REQ-074**
Public command preflight MUST validate every target-local input, owner, row, selector, and profile before setup or child work. Removed inputs and unknown Make command-line inputs MUST fail as `usage_error`; undeclared inherited environment variables retain the Section 5 ignore rule. A caller MUST NOT obtain partial selection by supplying an invalid row.
Verified by: TH-HARNESS-AC-064, TH-HARNESS-AC-065

**TH-HARNESS-REQ-075**
`explain-test-owner` MUST report the selected manifest, semantic families, row counts, runner and evidence distribution, execution profiles, minimum-tier distribution, and exact narrow commands. `task-guide ROLE=module-author OWNER=<owner_id>` MUST derive its focused, generation, and broader commands from the same catalog and topology snapshot. Neither command may infer ownership from paths or documentation.
Verified by: TH-HARNESS-AC-064, TH-HARNESS-AC-071

**TH-HARNESS-REQ-076**
The hard cutover MUST update this NLSpec, authored task surface, catalog, verification contracts, topology, schemas, generated outputs, public help, active guides, and deletion obligations in one merged change set. A state in which current and predecessor public selection are both supported is nonconforming.
Verified by: TH-HARNESS-AC-070, TH-HARNESS-AC-071

### 4.2 Command Family Defaults

Target membership for each family is defined only by `### 4.3 Public Target Registry`. The family-default table owns shared behavior for targets whose registry row declares that family. If a target appears in a prose command list and not in the registry, the registry governs and the prose is editorial drift.

| Family | Family ID | Required inputs | Optional inputs and defaults | Output class family | Scheduler use | Backing services | Artifact behavior | Failure contract |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| help and discovery | `help_discovery` | Target-local inputs declared by the Section 5.3 per-target input registry. | Omitted optional inputs select documented summary views according to the target's `input_contract`. | The target row's Section 4.3 output class, constrained by Section 7.2. | None | None | Does not create central run evidence unless the target row declares a summary schema. | Usage/config errors use Section 9. |
| bootstrap and toolchain | `bootstrap_toolchain` | Required local tools according to target | Tool paths default from Section 5. | `summary_with_artifacts` | None | May download/install repo-local tools. | Tool-run summary required; readiness cache artifacts MAY be retained when a cache profile is active. | Tool/config failures are `configuration_error` or `preflight_error`. |
| local services and dev | `local_services_dev` | Docker Compose and local config required by the target row. | `CONFIG_FILE=configs/dev/config.toml`, `OBJECT_STORE_BUCKET=cartulary` for rows that read those variables. | `service_summary` or `interactive_raw` | None | Compose Postgres/SeaweedFS S3 and local processes. | `service_summary` rows emit service summaries; `dev` has no verification artifact contract. | Startup/readiness/config failures are harness operational failures. |
| generated and drift | `generated_drift` | Owner inputs and manifests | `RESULTS_DIR` rows in the Section 5.3 input matrix select retained evidence. | `summary_with_artifacts` | Child work is scheduled only through Section 10 normalized scheduler work units. | Migration drift may use scratch Postgres when scheduled. | Tool summary and command-specific drift/finalizer files. | Drift mismatch is `artifact_error` or `scheduler_accounting_error`; unsafe retained-run finalization evidence is `artifact_error` or `configuration_error`. |
| owner and service slices | `test_owner_slices` | `OWNER` and any exact row selector declared by Section 5.3. | Omitted `ROWS` uses TH-HARNESS-REQ-360; worker defaults and `JSON` use Section 5.3. | `scheduler_summary_with_artifacts` | Compiles owner selection into the generic work graph. | Service-backed owner work acquires its declared broker capabilities. | Canonical run/target/unit/row artifacts and owner evidence. | Missing/invalid owner or row is `usage_error`; child failures retain child class. |
| backend and frontend leaf tests | `backend_frontend_leaf_tests` | Toolchain, catalog, and package inputs | Parallelism and worker variables from Section 5. | `summary_with_artifacts` | Service-backed targets may use a scheduler or testservices. | Store/integration/process targets require Postgres/object-store services when service-backed. | Owner, target, tool, logs, and reports. | Product assertion failures are `test_assertion_failure`; setup failures are operational. |
| browser E2E | `browser_e2e` | Node/pnpm, Playwright browser, backend/migrate/server support, services | `PLAYWRIGHT_WORKERS=3` unless overridden by Section 5. | `summary_with_artifacts` | Compiles semantic groups, resets, lifecycle work, and finalizers into the Section 10 work graph. | Broker-leased Postgres, object-store namespace, backend, frontend, and browser runtime capabilities. | Canonical run artifacts plus browser group, stack, Playwright, reset, and target evidence. | Product assertions are product failures; stack/readiness/reset failures are operational. |
| aggregates and gates | `aggregates_gates` | Toolchain and child inputs | `summary` output mode by default; `ci` defaults to `ci` mode through Make-owned CI target environment. | aggregate or scheduler output classes | Every aggregate compiles one union graph with exact producer edges and same-run unit deduplication. | Service-backed and browser units acquire broker capabilities. | Canonical run and target summaries plus target-owned evidence. | Exit nonzero if any required unit or artifact validation fails. |
| static analysis and security | `static_analysis_security` | Toolchain and source roots | Rule, flag, package, and Fallow static profiles named by public target rows and Section 5.3 inputs. Shell lint is blocking for public Make targets. `GOVULNCHECK_DB` is the only public security-profile override in the current profile. Fallow Runtime and Fallow security-scan commands are not selected by this profile. | `summary_with_artifacts` | Scheduled through Section 10 work units. | None | Tool summary and logs; cache reuse is allowed only by the registered graph policy, and vulnerability reuse additionally requires a proven database revision. | Findings are gate failures for scheduled local correctness targets. Advisory targets MUST be explicitly selected outside local `check`. |
| builds | `builds` | Build inputs and toolchain | Output paths from Make variables. | `summary_with_artifacts` | Scheduled as readiness work only through Section 10 normalized scheduler work units. | None | Tool summary and build logs; build cache artifacts MAY be retained when a cache profile is active. | Build failures are gate failures. |
| cleanup | `cleanup` | None | Uses Make path registries. | `destructive_human` | None | Does not stop Docker Compose services. | No central summary contract. | Unsafe path guard failure exits nonzero; missing paths are not failures. |
| formatting | `formatting` | Toolchain | None | `summary_with_artifacts` | None | None | Tool summary and formatter logs. | Formatter failure is operational; formatter rewrites are mutating. |

For `test_owner_slices`, omitted `ROWS` selects the inventory closed by TH-HARNESS-REQ-360. An explicit `ROWS` value selects only active executable row IDs owned by the exact requested owner. Every `service-backed-test-slice` has `dependency_scope="service_backed"`; omission still produces `completion_scope="full_owner"` over the owner's service-backed inventory, while an explicit selection produces `completion_scope="selected_subset"`.

`VITEST_MAX_WORKERS` and `PLAYWRIGHT_WORKERS` on this family apply only to
selected child work in the matching runner family. A slice that selects no
matching child MAY accept and report the bounded input but MUST NOT use it to
change another runner's concurrency or scheduler resource limits.

### 4.3 Public Target Registry

Every command below inherits the matching family defaults. `Default inclusion sets` lists direct target-policy membership only; row selection is derived from minimum tiers. `helper_only` means the target is public and directly invocable, but is not selected by default by `test`, `check`, `ci`, or `release-check` unless an aggregate policy explicitly selects it. `helper_only` MUST NOT mean private, uncontracted, or exempt from public-target output, configuration, failure, and cleanup contracts.

`Command ID` is the stable semantic command contract; the Make target is the current invocation binding. `Family ID` binds the target to Section 4.2 family defaults. `Semantic behaviors` declares the observable harness operation required by TH-HARNESS-REQ-058. `Side effects` declares the target's intentional mutation and resource contract from TH-HARNESS-REQ-059. The visible `Side effects` column MUST match the `side_effects[].class` list in `tools/task_surface_manifest.json`. `Lifecycle state` is defined by Section 4.6.

Public aggregate targets are target projections over a canonical union graph.
Distinct resource, policy, or lifecycle boundaries compile to separate semantic
units with exact producer edges. `migration-drift`, for example, retains one
public command projection while static migration-input validation and migration
database evidence retain separate unit, resource, lease, and artifact identities.

| Target | Command ID | Family ID | Default inclusion sets | Output class | Stable summary schema | Semantic behaviors | Side effects | Lifecycle state | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `help` | `cartulary.harness.command.help.v1` | `help_discovery` | `check` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `help-all` | `cartulary.harness.command.help_all.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `doctor` | `cartulary.harness.command.doctor.v2` | `bootstrap_toolchain` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `bootstrap` | `cartulary.harness.command.bootstrap.v2` | `bootstrap_toolchain` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `tool_install` | `public_active` |  |
| `bootstrap-node-runtime` | `cartulary.harness.command.bootstrap_node_runtime.v2` | `bootstrap_toolchain` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `tool_install` | `public_active` |  |
| `frontend-toolchain` | `cartulary.harness.command.frontend_toolchain.v2` | `bootstrap_toolchain` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `frontend-install` | `cartulary.harness.command.frontend_install.v2` | `bootstrap_toolchain` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `tool_install` | `public_active` |  |
| `playwright-install` | `cartulary.harness.command.playwright_install.v2` | `bootstrap_toolchain` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `tool_install` | `public_active` |  |
| `db-up` | `cartulary.harness.command.db_up.v1` | `local_services_dev` | `helper_only` | `service_summary` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start` | `public_active` |  |
| `db-migrate` | `cartulary.harness.command.db_migrate.v1` | `local_services_dev` | `helper_only` | `service_summary` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation` | `public_active` | Proves local Postgres readiness under Section 11, starting an owned instance when absent, and applies current-line migrations without resetting the database or object storage. |
| `db-reset` | `cartulary.harness.command.db_reset.v1` | `local_services_dev` | `helper_only` | `service_summary` | `cartulary.tool_run_summary.v5` | `destructive_safety` (Section 13), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `destructive_cleanup` | `public_active` | Requires `CARTULARY_DESTRUCTIVE_CONFIRM=db-reset` unless `CARTULARY_CLEANUP_DRY_RUN=1`; resets only the local database and does not reset object storage. |
| `services-up` | `cartulary.harness.command.services_up.v1` | `local_services_dev` | `helper_only` | `service_summary` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start` | `public_active` |  |
| `services-down` | `cartulary.harness.command.services_down.v1` | `local_services_dev` | `helper_only` | `service_summary` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `destructive_safety` (Section 13), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `destructive_cleanup` | `public_active` | Stops local Compose services with named volumes preserved. |
| `test-services-session-up` | `cartulary.harness.command.test_services_session_up.v1` | `local_services_dev` | `helper_only` | `service_summary` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `security_boundary` (Section 15), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start` | `public_active` | Creates one explicit, expiring local managed-service session; accepts only an optional absolute `CARTULARY_TEST_SERVICES_SESSION_FILE`. |
| `test-services-session-status` | `cartulary.harness.command.test_services_session_status.v1` | `local_services_dev` | `helper_only` | `machine_stdout_json` | `cartulary.test_services.local_session_status.v1` | `service_lifecycle` (Section 11), `security_boundary` (Section 15), `failure_normalization` (Section 9) | `none` | `public_active` | Emits one redacted session-status object and performs no lifecycle mutation. |
| `test-services-session-down` | `cartulary.harness.command.test_services_session_down.v1` | `local_services_dev` | `helper_only` | `service_summary` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `destructive_safety` (Section 13), `security_boundary` (Section 15), `failure_normalization` (Section 9) | `retained_artifacts`, `destructive_cleanup` | `public_active` | Refuses live borrowers and removes only an exactly proven session; repeated removal succeeds. |
| `object-store-init` | `cartulary.harness.command.object_store_init.v1` | `local_services_dev` | `helper_only` | `service_summary` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation` | `public_active` | Proves local object-store readiness under Section 11, starting an owned instance when absent, and initializes the configured bucket without requiring Postgres. |
| `object-store-reset` | `cartulary.harness.command.object_store_reset.v1` | `local_services_dev` | `helper_only` | `service_summary` | `cartulary.tool_run_summary.v5` | `destructive_safety` (Section 13), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `destructive_cleanup` | `public_active` | Requires `CARTULARY_DESTRUCTIVE_CONFIRM=object-store-reset` unless `CARTULARY_CLEANUP_DRY_RUN=1`; clears objects only from the configured local object-store bucket. |
| `dev` | `cartulary.harness.command.dev.v1` | `local_services_dev` | `helper_only` | `interactive_raw` | none | `service_lifecycle` (Section 11) | `service_start` | `public_active` |  |
| `generate` | `cartulary.harness.command.generate.v2` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `generated_artifacts` | `public_active` |  |
| `generate-drift` | `cartulary.harness.command.generate_drift.v2` | `generated_drift` | `check` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` |  |
| `generated-artifact-policy-check` | `cartulary.harness.command.generated_artifact_policy_check.v2` | `generated_drift` | `check` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` |  |
| `json-shape-check` | `cartulary.harness.command.json_shape_check.v2` | `generated_drift` | `check` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` |  |
| `openapi-compatibility-check` | `cartulary.harness.command.openapi_compatibility_check.v2` | `generated_drift` | `check` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` |  |
| `toolchain-drift` | `cartulary.harness.command.toolchain_drift.v2` | `generated_drift` | `check` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` |  |
| `migration-drift` | `cartulary.harness.command.migration_drift.v2` | `generated_drift` | `check` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts`, `service_resource_mutation` | `public_active` |  |
| `standup-package-smoke` | `cartulary.harness.command.standup_package_smoke.v2` | `local_services_dev` | `helper_only` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts`, `build_outputs`, `service_start`, `service_resource_mutation` | `public_active` | Builds and smokes the MVP on-prem stand-up package with one app image plus local Postgres and SeaweedFS S3, verifies embedded browser assets, package-local object-store init, readiness, persistent Docker-volume roots, and WebSocket Origin behavior. It is package smoke evidence only and MUST NOT be represented as disconnected-profile or backup/restore conformance evidence. |
| `standup-operational-recovery-smoke` | `cartulary.harness.command.standup_operational_recovery_smoke.v2` | `local_services_dev` | `helper_only` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts`, `build_outputs`, `service_start`, `service_resource_mutation` | `public_active` | Builds and smokes the MVP on-prem operational recovery workflow, creates a backup through the canonical operator recovery result schema, inspects the latest retained backup through the canonical inspect command, runs due restore verification against an isolated target, proves public recovery route-family absence, and retains command-specific recovery artifacts. It is not disconnected-profile evidence and does not reclassify `standup-package-smoke` as backup/restore conformance evidence. |
| `agent-finalize` | `cartulary.harness.command.agent_finalize.v2` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts`, `generated_artifacts` | `public_active` |  |
| `test-evidence-audit` | `cartulary.harness.command.test_evidence_audit.v3` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` | Audits one owner's catalog rows across compatible retained broad-check, support, visual, accessibility, and measurement roots. |
| `benchmark-claim-check` | `cartulary.harness.command.benchmark_claim_check.v2` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` | Validates retained Core 05 benchmark claim artifacts when the default benchmark manifest exists; absence of the default manifest is a no-claim pass, while an explicitly configured non-default missing manifest remains a harness failure. |
| `task-surface-report` | `cartulary.harness.command.task_surface_report.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `task-guide` | `cartulary.harness.command.task_guide.v2` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` | Requires `ROLE=module-author` and `OWNER`. |
| `author-test-row-id` | `cartulary.harness.command.author_test_row_id.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `test-slice` | `cartulary.harness.command.test_slice.v2` | `test_owner_slices` | `helper_only` | `scheduler_summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `scheduler_orchestration` (Section 10), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Selects every active owner row when `ROWS` is omitted. |
| `service-backed-test-slice` | `cartulary.harness.command.service_backed_test_slice.v2` | `test_owner_slices` | `helper_only` | `scheduler_summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `scheduler_orchestration` (Section 10), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Selects only rows whose runtime profile requires managed services. |
| `backend-unit` | `cartulary.harness.command.backend_unit.v2` | `backend_frontend_leaf_tests` | `test`, `check`, `ci` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` |  |
| `backend-store` | `cartulary.harness.command.backend_store.v2` | `backend_frontend_leaf_tests` | `test`, `check`, `ci` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts`, `service_start`, `service_resource_mutation` | `public_active` |  |
| `backend-integration` | `cartulary.harness.command.backend_integration.v2` | `backend_frontend_leaf_tests` | `test`, `check`, `ci` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts`, `service_start`, `service_resource_mutation` | `public_active` |  |
| `backend-process` | `cartulary.harness.command.backend_process.v2` | `backend_frontend_leaf_tests` | `test`, `check`, `ci` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts`, `service_start`, `service_resource_mutation` | `public_active` |  |
| `otel-conformance` | `cartulary.harness.command.otel_conformance.v2` | `backend_frontend_leaf_tests` | `check`, `ci` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `security_boundary` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Validates source snapshot, generated constants evidence, emitted telemetry goldens, browser non-export, retained raw capture policy, and telemetry security boundaries. |
| `target-plan` | `cartulary.harness.command.target_plan.v2` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `target-plan-json` | `cartulary.harness.command.target_plan_json.v2` | `help_discovery` | `helper_only` | `machine_stdout_json` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `fixture-report` | `cartulary.harness.command.fixture_report.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `explain-run` | `cartulary.harness.command.explain_run.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `harness-observability-check` | `cartulary.harness.command.harness_observability_check.v2` | `generated_drift` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4), `security_boundary` (Section 15) | `none` | `public_active` | Read-only validation of deterministic harness diagnostic projection for one exact retained run. |
| `harness-otel-export` | `cartulary.harness.command.harness_otel_export.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4), `security_boundary` (Section 15) | `external_network` | `public_active` | Explicit post-run OTLP export; ordinary harness commands never invoke it. |
| `harness-performance-check` | `cartulary.harness.command.harness_performance_check.v5` | `generated_drift` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4), `failure_normalization` (Section 9) | `none` | `public_active` | Validates exact target/provider-scoped baseline and candidate evidence windows under Section 10.5. |
| `harness-public-target-duration-baselines` | `cartulary.harness.command.harness_public_target_duration_baselines.v5` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `diagnostic_synthesis` (Section 4), `failure_normalization` (Section 9) | `retained_artifacts`, `generated_artifacts` | `public_active` | Sole writer of the public-target duration baseline artifact from exact qualified target/provider windows. |
| `explain-test-owner` | `cartulary.harness.command.explain_test_owner.v2` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` | Read-only catalog and topology explanation for one owner. |
| `explain-target` | `cartulary.harness.command.explain_target.v2` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `go-test-duration-baselines` | `cartulary.harness.command.go_test_duration_baselines.v4` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `go-test-duration-baseline-coverage` | `cartulary.harness.command.go_test_duration_baseline_coverage.v4` | `generated_drift` | `check` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` |  |
| `go-test-duration-baseline-drift` | `cartulary.harness.command.go_test_duration_baseline_drift.v4` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `browser-e2e-duration-baselines` | `cartulary.harness.command.browser_e2e_duration_baselines.v3` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `browser-e2e-duration-baseline-drift` | `cartulary.harness.command.browser_e2e_duration_baseline_drift.v3` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `service-backed-make-target-duration-baselines` | `cartulary.harness.command.service_backed_make_target_duration_baselines.v3` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `service-backed-make-target-duration-baseline-drift` | `cartulary.harness.command.service_backed_make_target_duration_baseline_drift.v3` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `harness-smoke-duration-baselines` | `cartulary.harness.command.harness_smoke_duration_baselines.v3` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `harness-smoke-duration-baseline-drift` | `cartulary.harness.command.harness_smoke_duration_baseline_drift.v3` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `scheduler-event-order-drift` | `cartulary.harness.command.scheduler_event_order_drift.v2` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `scheduler-summary-timing-drift` | `cartulary.harness.command.scheduler_summary_timing_drift.v2` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` | Canonical event/run/target timing closure only. |
| `frontend-typecheck` | `cartulary.harness.command.frontend_typecheck.v2` | `backend_frontend_leaf_tests` | `test`, `check`, `ci` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` |  |
| `frontend-unit` | `cartulary.harness.command.frontend_unit.v2` | `backend_frontend_leaf_tests` | `test`, `check`, `ci` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` |  |
| `frontend-import-boundary-check` | `cartulary.harness.command.frontend_import_boundary_check.v2` | `backend_frontend_leaf_tests` | `check` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` |  |
| `backend-module-boundary-check` | `cartulary.harness.command.backend_module_boundary_check.v2` | `static_analysis_security` | `check` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `security_boundary` (Section 8), `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` | Enforces backend module ownership boundaries from `tools/backend_module_boundaries.json` and emits `cartulary.backend_module_boundary_summary.v1`. |
| `frontend-fallow-static` | `cartulary.harness.command.frontend_fallow_static.v2` | `static_analysis_security` | `helper_only` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` | Current helper-only Fallow static profile. Emits `cartulary.fallow_static_summary.v2`; Fallow Runtime and Fallow security scans are not selected. |
| `lint-biome` | `cartulary.harness.command.lint_biome.v2` | `static_analysis_security` | `check` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` |  |
| `lint-scripts` | `cartulary.harness.command.lint_scripts.v2` | `static_analysis_security` | `check` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` |  |
| `lint-markdown` | `cartulary.harness.command.lint_markdown.v2` | `static_analysis_security` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` | Structural Markdown lint over authored active docs; generated ledgers remain generator-owned. |
| `harness-contract` | `cartulary.harness.command.harness_contract.v2` | `static_analysis_security` | `ci`, `release-check` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` | Runs extended harness topology, schema, and field-shape contract checks outside default local `check`. |
| `lint-shell` | `cartulary.harness.command.lint_shell.v2` | `static_analysis_security` | `check` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` |  |
| `format` | `cartulary.harness.command.format.v2` | `formatting` | `helper_only` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `authored_source_write` | `public_active` |  |
| `browser-e2e` | `cartulary.harness.command.browser_e2e.v2` | `browser_e2e` | `test`, `ci`, `release-check` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Direct aggregate remains full browser evidence and is not a default local `check` member. |
| `browser-e2e-webserver-backed` | `cartulary.harness.command.browser_e2e_webserver_backed.v2` | `browser_e2e` | `test`, `ci` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Direct and aggregate selections compile the same webserver-backed group units and full-fidelity evidence. |
| `browser-e2e-stateful` | `cartulary.harness.command.browser_e2e_stateful.v2` | `browser_e2e` | `test`, `check`, `ci` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Stateful affinity, reset, and group order are explicit graph dependencies. |
| `browser-e2e-measurement` | `cartulary.harness.command.browser_e2e_measurement.v2` | `browser_e2e` | `test`, `ci` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` |  |
| `browser-e2e-a11y` | `cartulary.harness.command.browser_e2e_a11y.v2` | `browser_e2e` | `test`, `ci`, `release-check` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Emits `cartulary.frontend_accessibility_summary.v4` for active selected accessibility rows only. Direct target evidence remains full-fidelity and default local `check` does not schedule current accessibility work. Release-check uses it as release-readiness evidence only. |
| `browser-e2e-visual` | `cartulary.harness.command.browser_e2e_visual.v2` | `browser_e2e` | `test`, `ci`, `release-check` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Direct target evidence runs the full active visual catalog inventory. Owner slices pass exact selected visual row IDs. Default local `check` does not schedule full visual work. Release-check uses it as release-readiness evidence only. |
| `browser-e2e-visual-update` | `cartulary.harness.command.browser_e2e_visual_update.v2` | `browser_e2e` | `helper_only` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `authored_source_write`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Helper-only refresh command for committed Playwright visual goldens. It runs the same visual stack and row selection as `browser-e2e-visual` with snapshot-update mode, mutates only authored snapshot PNGs under the visual spec snapshot directory, and MUST NOT be selected by `check`, `test`, `ci`, or release gates. |
| `test-fast` | `cartulary.harness.command.test_fast.v2` | `aggregates_gates` | `test`, `check`, `ci` | `aggregate_summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` |  |
| `test` | `cartulary.harness.command.test.v2` | `aggregates_gates` | `test`, `check`, `ci` | `aggregate_summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` |  |
| `lint` | `cartulary.harness.command.lint.v2` | `aggregates_gates` | `helper_only` | `aggregate_summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` | Sequence aggregate that emits run and target summaries for blocking lint/typecheck children. |
| `go-vulncheck` | `cartulary.harness.command.go_vulncheck.v2` | `static_analysis_security` | `check` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `security_boundary` (Section 15), `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` | Runs Govulncheck with structured JSON output and retains `govulncheck-findings.json` with schema ID `cartulary.govulncheck_findings.v1`. Symbol-reachable vulnerability findings are blocking `security_finding` failures; package-only and module-only findings are retained as diagnostic security evidence unless a later profile promotes them. |
| `go-gosec-targeted` | `cartulary.harness.command.go_gosec_targeted.v2` | `static_analysis_security` | `check` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `security_boundary` (Section 15), `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` |  |
| `go-gosec-audit` | `cartulary.harness.command.go_gosec_audit.v2` | `static_analysis_security` | `ci`, `release-check` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `security_boundary` (Section 15), `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` | Advisory no-fail audit evidence. It MUST NOT be selected by default local `check`. |
| `check` | `cartulary.harness.command.check.v2` | `aggregates_gates` | `check` | `scheduler_summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `scheduler_orchestration` (Section 10), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` |  |
| `ci` | `cartulary.harness.command.ci.v3` | `aggregates_gates` | `ci` | `aggregate_summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Duration-baseline drift remains an explicit post-run maintenance command and is not an in-flight CI child. |
| `release-check` | `cartulary.harness.command.release_check.v2` | `aggregates_gates` | `release-check` | `aggregate_summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Runs release child helpers for extended harness contract checks, advisory security audit evidence, SBOM/license evidence, SeaweedFS S3 release-gate evidence, builds, deployable-shape evidence, frontend support/visual/accessibility readiness children, and final release-readiness aggregation. |
| `release-readiness-evidence` | `cartulary.harness.command.release_readiness_evidence.v2` | `aggregates_gates` | `release-check` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` | Validates the exact canonical target-projection closure required by `release-check`; it writes no parallel release-evidence format and does not promote design/support evidence to product conformance or Core 05 publication evidence. |
| `seaweedfs-compatibility` | `cartulary.harness.command.seaweedfs_compatibility.v2` | `local_services_dev` | `release-check` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `scheduler_orchestration` (Section 10) | `retained_artifacts`, `service_start`, `service_resource_mutation` | `public_active` | Runs the dedicated SeaweedFS S3 compatibility profile and emits the full `SWFS-COMP-*` report outside `services-up` as a command-specific retained artifact. |
| `seaweedfs-release-evidence` | `cartulary.harness.command.seaweedfs_release_evidence.v2` | `static_analysis_security` |  | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `security_boundary` (Section 15), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` | Runs current SeaweedFS compatibility as a prerequisite, emits SeaweedFS release evidence, and emits a non-enforcing release-gate summary as command-specific retained artifacts; missing strict child evidence is reported as blocked evidence rather than hidden. |
| `seaweedfs-release-gate` | `cartulary.harness.command.seaweedfs_release_gate.v2` | `static_analysis_security` | `release-check` | `summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `security_boundary` (Section 15), `scheduler_orchestration` (Section 10) | `retained_artifacts` | `public_active` | Runs current SeaweedFS compatibility and enforces the strict SeaweedFS release gate from current-run compatibility, backup-integrity, redaction, storage-ref owner, security, license, and occurrence evidence. The release-gate summary is a command-specific retained artifact. |
| `build` | `cartulary.harness.command.build.v2` | `builds` | `helper_only` | `aggregate_summary_with_artifacts` | `cartulary.harness_run_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `build_outputs` | `public_active` |  |
| `build-server` | `cartulary.harness.command.build_server.v2` | `builds` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `build_outputs` | `public_active` | In default local `check`, build evidence is readiness for downstream service-backed work, not release deployable-shape evidence. |
| `build-server-harness` | `cartulary.harness.command.build_server_harness.v2` | `builds` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `build_outputs` | `public_active` | Builds the harness-only profile of the existing `server` identity. Default local `check` selects it as process/browser runtime readiness; it is never release deployable-shape evidence. |
| `build-migrate` | `cartulary.harness.command.build_migrate.v2` | `builds` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `build_outputs` | `public_active` | In default local `check`, build evidence is readiness for migration and service-backed work, not release deployable-shape evidence. |
| `build-operator` | `cartulary.harness.command.build_operator.v2` | `builds` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `build_outputs` | `public_active` | Selected through `build`, CI, release-shaped gates, and scheduler-visible operator runtime-binary readiness. Default local `check` builds the operator only when selected runtime-binary work declares it. |
| `build-web` | `cartulary.harness.command.build_web.v2` | `builds` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `build_outputs` | `public_active` | In default local `check`, build evidence is readiness for browser preview work, not release deployable-shape evidence. |
| `clean` | `cartulary.harness.command.clean.v1` | `cleanup` | `helper_only` | `destructive_human` | none | `destructive_safety` (Section 13), `failure_normalization` (Section 9) | `destructive_cleanup` | `public_active` |  |
| `distclean` | `cartulary.harness.command.distclean.v1` | `cleanup` | `helper_only` | `destructive_human` | none | `destructive_safety` (Section 13), `failure_normalization` (Section 9) | `destructive_cleanup` | `public_active` |  |

**TH-HARNESS-REQ-059**
Every public target MUST declare one or more side-effect classes in the public target registry source. The declaration MUST be represented as `side_effects[]`, where each entry is an object with `class`, `owner_section`, and the class-specific details required by the table below. A target that performs an undeclared side effect is non-conformant. `none` is mutually exclusive with every other side-effect class.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-004, TH-HARNESS-AC-020, TH-HARNESS-AC-023

| Side-effect class | Meaning | Required declaration |
| --- | --- | --- |
| `none` | No intentional file, service, or resource mutation outside ordinary terminal output. | Target row declares only `side_effects[].class=none`. |
| `retained_artifacts` | Writes retained run-root artifacts. | Artifact policy declares retained artifact families or paths. |
| `generated_artifacts` | Mutates checked-in generated or maintenance artifacts. | Target row declares exact generated file families. |
| `authored_source_write` | Mutates authored source files. | Target row declares source families or paths. |
| `build_outputs` | Writes reproducible build outputs. | Target row declares output roots or artifact families. |
| `tool_install` | Installs or updates repo-local tools or dependencies. | Target row declares install root and cleanup behavior. |
| `service_start` | Starts local or harness-owned services or runtime processes. | Target row declares ownership mode and lifecycle machine. |
| `service_resource_mutation` | Creates, modifies, or deletes service resources such as scratch databases, buckets, fixture resources, or local service bootstrap resources. | Target row declares ownership mode, resource families, and lifecycle machine. |
| `destructive_cleanup` | Deletes files, directories, services, databases, buckets, or other resources. | Target row cites Section 13 predicates. |
| `runtime_reset` | Mutates test runtime state through the test-only reset boundary. | Target row cites Section 12 predicates. |
| `external_network` | Sends a caller-selected diagnostic payload to an external endpoint. | Target row cites the exact input, protocol, privacy, timeout, redirect, and failure contract. |

**TH-HARNESS-REQ-077**
The authored task-surface owner MUST declare a closed `observability_policy`
that assigns every public target exactly one explicit disposition of `required`,
`excluded`, or `out_of_scope`. An `excluded` or `out_of_scope` entry MUST have a
nonempty owner section and reason. Every required target MUST bind exactly one
stable measurement profile; its target-row command ID and resolved profile's
canonical inputs, direct or aggregate eligibility, warm-up policy, and
performance gate form the measurement identity. Parameterized
slice and audit profiles MUST use `OWNER=module.auth`; the audit profile MUST
consume that owner's retained slice evidence. A target binding MAY declare one
target-specific normalized policy transition when a shared measurement
profile's gates and canonical inputs remain unchanged; that override MUST be a
closed validator-owned transition and MUST NOT affect any sibling binding.
Duplicate target steps are invalid
until an occurrence-aware artifact contract is adopted. Generated task-surface
projections MUST preserve the policy and validation MUST enumerate the complete
public surface against it, rejecting omissions, overlap, unknown targets, and
unowned exclusions. A check-internal target MAY have a measurement-profile
binding outside the public disposition sets only when a named acceptance gate
requires that exact diagnostic. Execution-context capture and performance
qualification MUST retain and evaluate every declared binding. Defaults,
runtime-family inference, and target-name inference are forbidden.
`scheduler-event-order-drift` and
`scheduler-summary-timing-drift` are explicitly `out_of_scope`: they validate
caller-selected retained evidence, so their duration describes that external
evidence selection rather than a stable command workload.
Verified by: TH-HARNESS-AC-072, TH-HARNESS-AC-073

**TH-HARNESS-REQ-078**
Every successful, failed, or interrupted top-level invocation with
`observability.disposition=required` MUST attempt local observability
finalization after native summaries and cleanup evidence are stable. One
top-level invocation produces one invocation trace. Nested public targets are
child spans selected from sequence, scheduler, summary-group, and timing
relationships; they MUST NOT become disconnected root traces merely because
their Make targets are independently public.
Verified by: TH-HARNESS-AC-073, TH-HARNESS-AC-074

**TH-HARNESS-REQ-079**
The current public observability and performance-maintenance commands are
`harness-observability-check`, `harness-otel-export`, and
`harness-performance-check`, plus the sole baseline writer
`harness-public-target-duration-baselines`; `explain-run` additionally accepts
`DETAIL=performance`. `harness-observability-check` reads one retained run,
reconstructs the selected invocation bundle in memory, validates retained
output and deterministic equivalence, and fails with `artifact_error`, exit
`11`, on missing, partial, malformed, unsafe, or nondeterministic evidence.
It MUST be strictly read-only and MUST NOT create a check-summary artifact in
the selected run. Check and export selection MUST name an exact retained run
directory or provide a result root together with `RUN_ID`; newest-run selection
is forbidden for these commands.
`harness-performance-check` reads the exact evidence-roots manifest supplied by
the caller and fails with `duration_baseline_drift`, exit `13`, when Section
10.5 acceptance fails.
Verified by: TH-HARNESS-AC-073, TH-HARNESS-AC-079

**TH-HARNESS-REQ-080**
`harness-otel-export` is the only current public harness network-export
surface. It MUST read a complete validated retained bundle without modifying
the selected run. `HARNESS_OTLP_ENDPOINT` is required and accepted from the
Make command line only. `HARNESS_OTLP_HEADERS_FILE` is optional and accepted
from the Make command line only. No ordinary test, aggregate, scheduler,
summary, finalizer, or explanation command may perform telemetry network
export.
Verified by: TH-HARNESS-AC-076

### 4.4 Direct Script and Package Boundary

| Surface                                                  | Classification                           | Contract                                                                                                  |
| -------------------------------------------------------- | ---------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| Root `package.json` scripts `build`, `test`, `typecheck` | Developer convenience                    | They do not promise Make result roots, run IDs, scheduler summaries, cleanup, or machine output.          |
| `apps/web/package.json` scripts                          | Developer convenience or child command   | Browser and unit scripts become harness child work only when invoked through Make wrappers.               |
| Raw owner-local helper scripts and CLIs                  | Tool-owned diagnostics or child commands | Their direct usage and exit codes are not public harness contracts unless a Make target adopts them.      |
| Raw Go, Vitest, Playwright, Biome, Vite, pnpm commands   | Tool-owned                               | Tool output schemas remain external or diagnostic unless consumed and normalized by a Make-owned wrapper. |

### 4.5 Public Wrapper Lifecycle

Every Make-owned public wrapper that is not `interactive_raw` MUST execute this observable lifecycle:

1. establish wrapper identity and target identity;
2. resolve output mode;
3. resolve and validate harness configuration;
4. compute result-root and run-id identity if the output class declares retained artifacts;
5. initialize redaction before capturing child output;
6. run the target's semantic behavior;
7. validate required schema-owned artifacts before success;
8. select the primary failure using Section 9.1;
9. run required cleanup or finalizers;
10. emit the target's public output according to Section 7;
11. expose the normalized public exit code through retained summaries and failure output, and exit nonzero on failure.

A target MAY skip a step only when its output class or target row explicitly declares that the step does not apply. A skipped step MUST NOT be implemented as an implicit child-command side effect.

Every public Make wrapper whose lifecycle invokes repo-owned Node tooling,
including the shared public preflight, MUST make pinned repo-local Node readiness
an explicit generated precondition before semantic behavior begins. The sole
exception is `bootstrap-node-runtime`, which MUST install or validate that
runtime before invoking any repo-owned Node module. If the pinned Node runtime
cannot be resolved, installed, downloaded, verified, or executed, the wrapper
MUST fail before semantic work with `failure_class=config`,
`failure_reason=configuration_error`, and public exit code `2`; it MUST NOT
surface a raw shell, raw `curl`, or `env` executable-not-found failure as the
public target result. Node runtime bootstrap MUST serialize mutation of the
repo-local runtime and archive paths, download into temporary files, publish an
archive only after checksum verification, remove corrupt or partial archive
candidates, and use bounded retry for transient download failures.

**TH-HARNESS-REQ-055**
Foundational public-wrapper schemas needed before frontend dependency readiness
MUST have generated, dependency-free validators committed as governed generated
artifacts. `cartulary.tool_run_summary.v5` is foundational in the current
profile. Its validator MUST execute with the pinned Node runtime when
`node_modules` is absent, MUST retain the exact schema's pass/fail behavior, and
MUST be regenerated from the registered schema rather than hand-maintained.
General schema validation MAY continue to use the pinned frontend AJV
dependency after frontend dependency readiness.
Verified by: TH-HARNESS-AC-092

Frontend dependency installation is itself foundational bootstrap work and
MUST NOT invoke a harness runner or schema validator supplied by the dependency
tree it is creating. The enclosing public wrapper MAY use only the pinned Node
runtime and dependency-free foundational validator until that install has
completed. `bootstrap` MUST model frontend dependency readiness as a distinct
first stage and MUST place all AJV-dependent tool-install and browser-install
work behind an explicit generated dependency edge from that stage. Textual
prerequisite order alone is insufficient because bootstrap MUST retain the same
ordering under parallel Make execution and as later phases are added.
Verified by: TH-HARNESS-AC-092

### 4.6 Public Target Lifecycle

A target has one of these public-lifecycle states:

| State | Meaning | Invocation behavior |
| --- | --- | --- |
| `candidate_child` | Internal or generated child work, not a public command. | MUST NOT be required for public conformance by direct invocation. |
| `public_active` | Current public command. | MUST satisfy all public target contracts. |
| `removed` | No longer public. | MUST NOT appear in the public registry. |

A target may move to `public_active` only when it passes the semantic-value test from TH-HARNESS-REQ-058. A pre-release target may move to `removed` by revising the registry and generated mirrors in one change set. `removed` is represented by absence from the public registry, not by a retained registry row.

## 5. Configuration Resolution Contract

**TH-HARNESS-REQ-100**
Every public Make target MUST resolve harness configuration through `resolve_harness_config()` before child work begins. A target that cannot resolve or validate configuration MUST fail with `failure_class=config`, `failure_reason=configuration_error`, and public exit code `2`.
Verified by: TH-HARNESS-AC-002, TH-HARNESS-AC-003, TH-HARNESS-AC-014

`resolve_harness_config()` is the normative configuration-resolution contract. Repository implementation entrypoints such as preflight helpers MAY wrap this resolver, but MUST NOT define a narrower public-target configuration contract.

**TH-HARNESS-REQ-101**
Generated manifests are execution inputs, not caller configuration. A caller-supplied variable that attempts to override a non-overridable manifest field MUST fail with `configuration_error` before child work.
Verified by: TH-HARNESS-AC-002

**TH-HARNESS-REQ-102**
When a scheduler work unit invokes a child runner that starts its own worker pool, the scheduler input MUST either declare logical resource claims equal to the child worker budget or constrain that child worker budget through scheduler-owned environment. The scheduler-owned value wins for that scheduled work unit even when the same variable has a different direct public-target default. In the current check profile, scheduled `frontend-unit` MUST run Vitest with `VITEST_MAX_WORKERS=2` and MUST claim `host_cpu=2`; direct `make frontend-unit` MAY keep a faster developer default outside the check scheduler.

Auto-derived scheduler capacity MUST NOT resolve below the largest declared claim for that resource in the normalized work-unit set. Caller overrides MAY still choose lower limits, but such overrides are configuration errors when they cannot satisfy a declared work unit.
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-021

**TH-HARNESS-REQ-103**
Frontend unit harness tests that depend on asynchronous jsdom rendering, workbook row hydration, inspector-subject hydration, controlled input replacement, row-history rendering, or virtualized grid mounting MUST use shared bounded wait helpers and stable selector builders with actionable diagnostics. The default wait budget MUST be finite and configuration-backed. When identity matters, the wait predicate MUST use stable workbook-row, inspector-subject, or row-history-item identity rather than visible count or text alone. An inspector-subject readiness helper MUST be observation-only after the invoking action and MUST match the expected `view_schema_id`, `record_id`, and `row_version`; it MUST NOT retry the action. Exact human diagnostic prose is non-normative; the diagnostic record MUST identify the expected row IDs, mounted row IDs, expected and mounted inspector-subject identity, received row-history item references, surface, inspector state, and failing selector class without including record payload values or reclassifying ordinary assertions away from `failure_class=product`.
Verified by: TH-HARNESS-AC-021

**TH-HARNESS-REQ-665**
Browser E2E helpers that perform a mutating UI action and then drive another action that depends on the committed result MUST wait for the server success response and for the rendered workbook projection to converge on the response's stable source-record identity before continuing. When the response supplies `source_record.row_version`, convergence MUST require the rendered source row to reach at least that version; a concurrent accepted version above the response version satisfies the floor, while a stale lower version does not. When the dependent action relies on optimistic concurrency, convergence MUST include the returned `row_version` rendered under the stable row identifier. A visible global save-state label such as `Saved` MAY be asserted after convergence, but it MUST NOT be the only completion predicate for a dependent mutation sequence.

A post-mutation value assertion in a refreshable or virtualized grid MUST root
each polling attempt at exactly one visible row selected by stable `record_id`,
then locate the editor or display field within that row. It MUST reacquire the
row and descendant after render-sensitive refreshes and MUST NOT choose an
editor by active-element state, DOM order, or a cross-row locator union.
Diagnostics MUST identify the expected and mounted record IDs without including
record payload values.

A helper that validates post-mutation focus or viewport continuity MUST be observation-only after the mutating action: it MUST NOT focus, scroll, click, press a key, or dispatch an input event to manufacture the postcondition it is measuring. Setup-time navigation and scrolling before the action remain allowed. A later passing invocation is distinct evidence and MUST NOT retry, replace, or reclassify an earlier product assertion failure.

This requirement owns browser synchronization and evidence only. Core 03 `REQ-03-283` remains the unchanged product authority for deterministic row-local focus restoration after same-surface follow-up rendering.
Verified by: TH-HARNESS-AC-016, TH-HARNESS-AC-021, TH-HARNESS-AC-080

**TH-HARNESS-REQ-676**
`apps/web/e2e` is a browser-verification discovery and evidence root, not a
product module or general support package. Its root TypeScript files MUST be
Playwright `*.spec.ts` entrypoints, `fixtures.ts`, `fixtures.test.ts`,
`global-setup.ts`, or `global-teardown.ts`. Support and page modules MUST live
under their semantic subdirectories, MUST be statically reachable from those
entrypoints or an E2E `*.test.ts` entrypoint, and MUST NOT export a declaration
that has no importer outside its declaring module. Type-only imports count as
reachability and use. Namespace imports and dynamic imports MUST NOT be used to
evade exact support-module liveness accounting.
Verified by: TH-HARNESS-AC-039, TH-HARNESS-AC-056

**TH-HARNESS-REQ-677**
Reusable E2E support clients that treat a selected public JSON operation as a
success MUST derive its method, path, request type, accepted success status,
and response validation from `@cartulary/protocol-ts/http`. The low-level
`requestPublicJson` client MUST remain private to `support/transport/**`.
Support-level raw requests are limited to readiness, the validated opaque
upload PUT, browser-context bulk mutation or interception, repeated-header
observation after generated success validation, and explicit owner-backed
non-success probes. Scenario-local product tests MAY retain direct raw requests
when the request or error bytes are the assertion target; that permission MUST
NOT be generalized into a reusable support client.
Verified by: TH-HARNESS-AC-039, TH-HARNESS-AC-056

**TH-HARNESS-REQ-678**
Two active browser catalog rows that resolve to the same owner, stage, project,
verification set, helper behavior, and asserted postconditions MUST NOT be
retained solely for title capitalization, fixture-label, historical scenario,
or compatibility reasons. The canonical existing row and scenario identity
MUST be retained, the duplicate MUST be removed from authored accounting, and
generated topology MUST be refreshed through `make generate`. A distinct
owner, postcondition, runtime profile, or behavior-sensitive fixture prevents
deduplication.
Verified by: TH-HARNESS-AC-082

Browser E2E tests that need the authenticated incident directory MUST navigate to that surface explicitly before asserting directory UI. Tests MUST NOT assume raw authenticated `/` always renders the incident directory, because Core 01 `REQ-01-580` makes raw root cardinality-sensitive and requires the sole visible incident to auto-open.
Verified by: TH-HARNESS-AC-016, TH-HARNESS-AC-021

**TH-HARNESS-REQ-115**
Browser E2E helpers that select a workbook surface through a menu, popover, selector, or tab strip MUST use stable selector builders rooted in canonical `view_schema_id` or `sheet_ref` identity. Such a helper MUST use bounded retries, reacquire locators after every render-sensitive action, require a single target option before selecting, and after selection converge on the active workbook shell surface signal, the canonical direct `view_schema_id` URL representation when selecting a base surface, and the target grid shell before returning. Final helper diagnostics MUST include the requested target identity, current URL, active shell surface, menu-open state, visible candidate target identities, and final retry error. This requirement governs browser-test synchronization only and MUST NOT define product behavior beyond the owning product specifications.
Verified by: TH-HARNESS-AC-016, TH-HARNESS-AC-021

**TH-HARNESS-REQ-111**
Make-owned frontend dependency installation MUST use the pinned repo-local Node and pnpm toolchain, MUST bind pnpm's content-addressable store to the repo-local `.pnpm-store` path through project configuration, MUST run without requiring a TTY or interactive confirmation, and MUST use a frozen lockfile. `frontend-install` is an install/readiness target, not a dependency-update target; if `pnpm-lock.yaml` is out of sync with workspace manifests, the target MUST fail with `failure_class=config`, `failure_reason=configuration_error`, and public exit code `2` rather than mutating the lockfile. A package-manager repair that purges and recreates repo-local `node_modules` is allowed only as part of this non-interactive install contract and only for repo-local workspace dependency roots.
Verified by: TH-HARNESS-AC-002, TH-HARNESS-AC-014, TH-HARNESS-AC-023

**TH-HARNESS-REQ-116**
`OWNER` is required for every owner-first public command. It has no default, uses `trim` normalization, remains case-sensitive after trimming, and MUST resolve to exactly one active owner. Missing, empty, malformed, inactive, and unknown values are `usage_error` with public exit `2` before setup.
Verified by: TH-HARNESS-AC-064

**TH-HARNESS-REQ-117**
`ROWS` is optional only for the two slice commands. The parser MUST retain raw comma-separated tokens until it rejects empty tokens and duplicates after trimming. Accepted values MUST be active row IDs owned by the exact `OWNER`, then normalized to unique ASCII-bytewise order. `ROWS=` and unknown, malformed, duplicate, inactive, cross-owner, or non-executable rows are `usage_error` with exit `2`. No valid subset may run when any requested token is invalid.
Verified by: TH-HARNESS-AC-064, TH-HARNESS-AC-065

**TH-HARNESS-REQ-118**
Omitted `ROWS` has the exact selection behavior in TH-HARNESS-REQ-360 and
selects the complete active owner inventory without tier filtering. A selection
that resolves to zero rows is `usage_error`; no support-only or migration
exception exists in the adopted profile.
Verified by: TH-HARNESS-AC-064, TH-HARNESS-AC-068

**TH-HARNESS-REQ-119**
`VITEST_MAX_WORKERS` defaults to decimal `4`; `PLAYWRIGHT_WORKERS` defaults to decimal `3`. Each accepts only a base-10 integer from `1` through `16`, inclusive. Validation occurs even when the selection contains no matching runner. A valid unused value MUST be recorded in `unused_inputs` and MUST NOT change another runner or scheduler resource limit.
Verified by: TH-HARNESS-AC-064

**TH-HARNESS-REQ-120**
Target-local `JSON` accepts only exact `1`; omitted or empty means human output. `JSON=1` and `CARTULARY_OUTPUT_MODE=machine` together are `usage_error` with exit `2`. For `test-slice` and `service-backed-test-slice`, `JSON=1` changes stdout only and execution still occurs; successful stdout is exactly one `cartulary.harness_run_summary.v1` object followed by LF. For `explain-test-owner` and `task-guide`, successful stdout is exactly one command-specific schema object followed by LF.
Verified by: TH-HARNESS-AC-004, TH-HARNESS-AC-064

**TH-HARNESS-REQ-121**
`test-evidence-audit` requires `EVIDENCE_ROOTS_FILE`, a caller-owned
`cartulary.harness_evidence_root_manifest.v1` file containing the exact owner ID and
ASCII-sorted unique `{target, run_root}` entries to audit. The auditor MUST derive
the required target partitions from the current catalog and verification contracts;
it MUST NOT widen the manifest, search sibling roots, or select a newest run. One
physical run root MAY be named by multiple explicit target entries. A known supplied
target that is not applicable to the selected owner MUST be reported in
`unused_inputs`; an unknown or duplicate target is `usage_error`. Missing required
target entries are `usage_error`; unsafe roots or incompatible contents are
`artifact_error`. As an atomic alternative to leaf target partitions, a manifest
MAY contain one `target=test-slice` entry whose retained canonical projection
selects the owner's complete active row set. In that mode the auditor MUST verify
the full-owner selection and all compatibility fields directly from that artifact;
it MUST treat any additional supplied target entries as unused and MUST NOT infer,
split, or search for leaf evidence. A partial or service-backed-only slice MUST NOT
satisfy the full-owner alternative.
Verified by: TH-HARNESS-AC-064, TH-HARNESS-AC-066

**TH-HARNESS-REQ-122**
`task-guide` requires exact `ROLE=module-author` and a valid `OWNER`; neither has a default. Every other role token and every delivery-phase input is unsupported in the current task-guide contract.
Verified by: TH-HARNESS-AC-064, TH-HARNESS-AC-071

**TH-HARNESS-REQ-123**
Any Make command-line input not declared by the selected target MUST fail before child work as `usage_error`; an undeclared inherited environment value MUST be ignored and MUST NOT reach child environments, as required by TH-HARNESS-REQ-112.

**TH-HARNESS-REQ-124**
Inputs that apply uniformly to every public target MUST be declared once in the
authored task-surface global-input registry and projected into the generated
manifest and Make preflight input set. A global input MUST NOT be copied into
every target-local input contract or admitted through an implementation-only
allowlist. Target-local declarations remain authoritative for inputs that do
not apply to every public target.
Verified by: TH-HARNESS-AC-002, TH-HARNESS-AC-092
Verified by: TH-HARNESS-AC-064, TH-HARNESS-AC-070

### 5.1 Precedence

| Precedence | Source                                 | Rule                                                                                                   |
| ---------: | -------------------------------------- | ------------------------------------------------------------------------------------------------------ |
|          1 | Make-owned wrapper CLI flags           | Highest priority only for flags explicitly declared by the target.                                     |
|          2 | Make command-line variables            | `VAR=value make target` overrides inherited environment for the same canonical variable.               |
|          3 | Exported environment inherited by Make | Accepted only for variables declared in the configuration table.                                       |
|          4 | Target manifest values                 | Source inputs for scheduler and target behavior, not caller overrides.                                 |
|          5 | Config files                           | Apply only to the application-under-test runtime unless the variable table declares a harness binding. |
|          6 | Hardcoded harness defaults             | Used only when all higher layers omit the value.                                                       |

### 5.2 Configuration Algorithm

```text
resolve_harness_config(target, raw_make_vars, raw_env, wrapper_cli_args):
  assert target is a public Make target or a Make-owned wrapper target
  declared = global_configuration_table + per_target_input_registry[target]
  resolved = empty map

  reject undeclared wrapper CLI flags
  reject caller overrides of manifest/internal fields
  reject undeclared public harness Make variables supplied on the Make command line

  for each declared variable in stable table order:
    candidates = [
      wrapper_cli_args value if variable has declared CLI binding,
      raw_make_vars value if supplied on the Make command line,
      raw_env value if exported before Make invocation,
      manifest value if variable has declared manifest binding,
      config_file value if variable has declared config-file binding,
      hardcoded default
    ]

    select the first candidate whose layer is allowed for the variable
    apply the variable's empty-string rule
    normalize the selected value
    validate the normalized value
    if validation fails:
      emit configuration_error summary when the target has a summary layer
      fail before child work with exit code 2
    record selected value, source layer, and normalized value

  ignore undeclared inherited environment variables
  strip undeclared public harness variables from child process environments
  emit resolved values required by Section 8 summaries
```

### 5.3 Per-Target Input Registry

**TH-HARNESS-REQ-112**
Every public target MUST declare a closed per-target input contract. This NLSpec owns the current public target-local input contract. Authored `tools/task_surface_owner.json` MUST use `cartulary.task_surface_owner.v2`, and generated `tools/task_surface_manifest.json` MUST use `cartulary.task_surface_manifest.v15` or a later adopted projection schema. Both MUST contain `input_contract` for every row with `target_class="public"`; neither may independently widen, narrow, or reinterpret the closed contract below. Execution topology may reference a target name but MUST NOT own or override its input contract.

Make-to-wrapper source transport is private. The current transport is one
`CARTULARY_MAKE_INPUT_SOURCES` value containing whitespace-separated
`NAME=cli|env|file|unset` entries for the closed public input inventory. Names MUST be
unique safe Make variable names, and unknown names or source tokens MUST fail before
child work. Per-variable `CARTULARY_MAKE_ORIGIN_<NAME>` variables are unsupported and
MUST NOT affect source resolution.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-002, TH-HARNESS-AC-027, TH-HARNESS-AC-058

Each `input_contract` MUST contain:

| Field | Required value |
| --- | --- |
| `undeclared_make_command_line` | `usage_error` |
| `undeclared_inherited_env` | `ignore` |
| `inputs[]` | Stable ordered array of accepted target-local inputs. Empty array means the public target accepts no target-local Make variables. |

Each `inputs[]` row MUST contain `name`, `binding`, `allowed_sources`, `required`, `type`, `default`, `empty_string`, `normalization`, `invalid_reason`, `summary_emission`, and `child_forwarding`. Rows MAY additionally contain bounded type metadata such as `values`, `min`, or `max`.

| Row field | Meaning |
| --- | --- |
| `name` | Uppercase Make variable name accepted by this target. |
| `binding` | Public invocation binding. The current profile accepts `make_variable`; a later profile MAY add wrapper CLI bindings only when Section 5.2 precedence remains preserved. |
| `allowed_sources` | Subset of `make_command_line`, `environment`, `makefile_default`, `internal_default`, and `manifest`. A source not listed for the row MUST NOT supply the value. |
| `required` | Whether omission after all allowed sources is a usage/configuration failure. |
| `type` | One of the Section 5.3 type tokens. |
| `default` | Default value or `null`; defaults are valid only when their source is declared. |
| `empty_string` | One of `invalid`, `omitted`, or `false`. |
| `normalization` | One of `none`, `trim`, `trim_lowercase`, or `path_token`. |
| `invalid_reason` | `usage_error` for caller selection mistakes or `configuration_error` for invalid paths, retained evidence, internal state, or manifest-derived configuration. |
| `summary_emission` | One of `none`, `value`, `redacted_value`, or `source_and_value`. |
| `child_forwarding` | One of `none`, `argv`, `runtime_env`, or `argv_and_runtime_env`; undeclared public harness inputs MUST NOT reach child environments. |

The closed target-local public input set in the current profile consists only of documented uses of `ROLE`, `OWNER`, `ROWS`, `TARGET`, `RESULTS_DIR`, explicit retained-evidence root selectors, `ALLOW_OLDER_RESULTS_DIR`, `RUN_ID`, `DETAIL`, `JSON`, worker controls, fixture report limits, duration-maintenance knobs, scheduler timing knobs, destructive-safety controls, the explicit Govulncheck database override, and the explicit `HARNESS_OTLP_ENDPOINT` and `HARNESS_OTLP_HEADERS_FILE` post-run export inputs. A public target accepts one of these names only when it appears in the normative input matrix below.

`frontend-fallow-static` accepts no target-local Make variables in the current Fallow static profile. A future changed-code audit base such as `FALLOW_CHANGED_SINCE` MUST be added to this registry before it becomes public input.

**TH-HARNESS-REQ-114**
Every public target's accepted target-local input set is closed by the normative input matrix. A public target that is not listed in the matrix accepts no target-local Make variables beyond the global variables in Section 5.5. A grouped row is valid only when every listed target has identical `type`, `default`, `allowed_sources`, `required`, `empty_string`, `normalization`, `values` or `min`/`max` bounds, `invalid_reason`, `summary_emission`, and `child_forwarding`.
Verified by: TH-HARNESS-AC-002, TH-HARNESS-AC-029

| Target(s) | Input | Type | Required | Allowed sources | Default | Omission behavior | Empty-string behavior | Normalization | Values/bounds | Invalid behavior | Summary emission | Child forwarding |
| --- | --- | --- | ---: | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `db-reset`, `services-down`, `object-store-reset`, `clean`, `distclean` | `CARTULARY_CLEANUP_DRY_RUN` | `exact_1_bool` | no | make command line, environment, makefile default | `false` | use `false` | `false` | `trim` | `exact_1_bool` | `usage_error`, exit `2` | `value` | `runtime_env` |
| `db-reset` | `CARTULARY_DESTRUCTIVE_CONFIRM` | `enum` | no | make command line | none | omitted | `invalid` | `trim` | `db-reset` | `usage_error`, exit `2` | `value` | `none` |
| `object-store-reset` | `CARTULARY_DESTRUCTIVE_CONFIRM` | `enum` | no | make command line | none | omitted | `invalid` | `trim` | `object-store-reset` | `usage_error`, exit `2` | `value` | `none` |
| `agent-finalize` | `ALLOW_OLDER_RESULTS_DIR` | `exact_1_bool` | no | make command line, environment, makefile default | none | omitted | `false` | `trim` | `exact_1_bool` | `usage_error`, exit `2` | `value` | `runtime_env` |
| `agent-finalize` | `RESULTS_DIR` | `result_selector` | no | make command line, environment, makefile default | none | omitted | `omitted` | `path_token` | `result_selector` | `usage_error`, exit `2` | `value` | `runtime_env` |
| `test-evidence-audit`, `task-guide`, `test-slice`, `service-backed-test-slice`, `explain-test-owner` | `OWNER` | `owner_id` | yes | make command line, environment, makefile default | none | missing required input | `invalid` | `trim` | `owner_id` | `usage_error`, exit `2` | `value` | `argv` |
| `test-evidence-audit`, `harness-performance-check` | `EVIDENCE_ROOTS_FILE` | `path` | yes | make command line, environment, makefile default | none | missing required input | `invalid` | `path_token` | `path` | `usage_error`, exit `2` | `value` | `argv` |
| `task-surface-report` | `TASK_SURFACE_REPORT_ARGS` | `task_surface_report_args` | no | make command line, environment, makefile default | none | omitted | `omitted` | `trim` | `task_surface_report_args` | `usage_error`, exit `2` | `value` | `argv` |
| `task-guide` | `ROLE` | `enum` | yes | make command line, environment, makefile default | none | missing required input | `invalid` | `trim` | `module-author` | `usage_error`, exit `2` | `value` | `argv` |
| `task-guide`, `test-slice`, `service-backed-test-slice`, `fixture-report`, `explain-test-owner`, `explain-target` | `JSON` | `exact_1_bool` | no | make command line, environment, makefile default | `false` | use `false` | `false` | `trim` | `exact_1_bool` | `usage_error`, exit `2` | `value` | `argv` |
| `author-test-row-id` | `FAMILY_ID` | `family_id` | yes | make command line, environment, makefile default | none | missing required input | `invalid` | `trim` | `family_id` | `usage_error`, exit `2` | `value` | `argv` |
| `author-test-row-id` | `CLAIM` | `semantic_text` | yes | make command line, environment, makefile default | none | missing required input | `invalid` | `trim` | `semantic_text` | `usage_error`, exit `2` | `value` | `argv` |
| `author-test-row-id` | `SELECTOR_KEY` | `semantic_text` | yes | make command line, environment, makefile default | none | missing required input | `invalid` | `trim` | `semantic_text` | `usage_error`, exit `2` | `value` | `argv` |
| `generate-drift`, `generated-artifact-policy-check`, `json-shape-check`, `openapi-compatibility-check`, `toolchain-drift`, `migration-drift`, `standup-package-smoke`, `standup-operational-recovery-smoke`, `agent-finalize`, `test-evidence-audit`, `benchmark-claim-check`, `test-slice`, `service-backed-test-slice`, `backend-unit`, `backend-store`, `backend-integration`, `backend-process`, `otel-conformance`, `go-test-duration-baseline-coverage`, `frontend-typecheck`, `frontend-unit`, `frontend-import-boundary-check`, `backend-module-boundary-check`, `frontend-fallow-static`, `lint-biome`, `lint-scripts`, `harness-contract`, `lint-shell`, `format`, `browser-e2e`, `browser-e2e-webserver-backed`, `browser-e2e-stateful`, `browser-e2e-measurement`, `browser-e2e-a11y`, `browser-e2e-visual`, `browser-e2e-visual-update`, `test-fast`, `test`, `lint`, `go-vulncheck`, `go-gosec-targeted`, `go-gosec-audit`, `check`, `ci`, `release-check`, `release-readiness-evidence`, `seaweedfs-compatibility`, `seaweedfs-release-evidence`, `seaweedfs-release-gate`, `build` | `CARTULARY_HARNESS_CACHE_MODE` | `enum` | no | make command line, environment, makefile default | `normal` | use `normal` | `omitted` | `trim` | `normal`, `cold`, `off` | `usage_error`, exit `2` | `value` | `runtime_env` |
| `generate-drift`, `generated-artifact-policy-check`, `json-shape-check`, `openapi-compatibility-check`, `toolchain-drift`, `migration-drift`, `standup-package-smoke`, `standup-operational-recovery-smoke`, `agent-finalize`, `test-evidence-audit`, `benchmark-claim-check`, `test-slice`, `service-backed-test-slice`, `backend-unit`, `backend-store`, `backend-integration`, `backend-process`, `otel-conformance`, `go-test-duration-baseline-coverage`, `frontend-typecheck`, `frontend-unit`, `frontend-import-boundary-check`, `backend-module-boundary-check`, `frontend-fallow-static`, `lint-biome`, `lint-scripts`, `harness-contract`, `lint-shell`, `format`, `browser-e2e`, `browser-e2e-webserver-backed`, `browser-e2e-stateful`, `browser-e2e-measurement`, `browser-e2e-a11y`, `browser-e2e-visual`, `browser-e2e-visual-update`, `test-fast`, `test`, `lint`, `go-vulncheck`, `go-gosec-targeted`, `go-gosec-audit`, `check`, `ci`, `release-check`, `release-readiness-evidence`, `seaweedfs-compatibility`, `seaweedfs-release-evidence`, `seaweedfs-release-gate`, `build` | `CARTULARY_HARNESS_CAPACITY_OVERRIDE` | `path` | no | make command line, environment, makefile default | none | omitted | `omitted` | `path_token` | `path` | `usage_error`, exit `2` | `value` | `runtime_env` |
| `test-slice`, `service-backed-test-slice` | `PLAYWRIGHT_WORKERS` | `positive_integer` | no | make command line, environment, makefile default | `3` | use `3` | `invalid` | `trim` | `1..16` | `usage_error`, exit `2` | `value` | `runtime_env` |
| `test-slice`, `service-backed-test-slice` | `ROWS` | `row_ids` | no | make command line, environment, makefile default | none | omitted | `invalid` | `trim` | `row_ids` | `usage_error`, exit `2` | `value` | `argv` |
| `test-slice`, `service-backed-test-slice`, `frontend-unit` | `VITEST_MAX_WORKERS` | `positive_integer` | no | make command line, environment, makefile default | `4` | use `4` | `invalid` | `trim` | `1..16` | `usage_error`, exit `2` | `value` | `runtime_env` |
| `target-plan`, `target-plan-json`, `fixture-report`, `explain-run`, `scheduler-event-order-drift`, `scheduler-summary-timing-drift` | `TARGET` | `target_name` | no | make command line, environment, makefile default | none | omitted | `omitted` | `trim` | `target_name` | `usage_error`, exit `2` | `value` | `argv` |
| `fixture-report` | `RESULTS_DIR` | `result_selector` | no | make command line, environment, makefile default | `.cartulary/test-results` | use `.cartulary/test-results` | `omitted` | `path_token` | `result_selector` | `usage_error`, exit `2` | `value` | `argv` |
| `fixture-report`, `explain-run`, `harness-observability-check`, `harness-otel-export` | `RUN_ID` | `run_id` | no | make command line, environment, makefile default | none | omitted | `omitted` | `trim` | `run_id` | `usage_error`, exit `2` | `value` | `argv` |
| `fixture-report` | `FIXTURE_THRESHOLD_MS` | `positive_integer` | no | make command line, environment, makefile default | `30000` | use `30000` | `omitted` | `trim` | `1..999999999` | `usage_error`, exit `2` | `value` | `argv` |
| `fixture-report` | `FIXTURE_TOP` | `positive_integer` | no | make command line, environment, makefile default | `5` | use `5` | `omitted` | `trim` | `1..999999999` | `usage_error`, exit `2` | `value` | `argv` |
| `explain-run`, `harness-observability-check`, `harness-otel-export`, `go-test-duration-baselines`, `browser-e2e-duration-baselines`, `service-backed-make-target-duration-baselines`, `harness-smoke-duration-baselines` | `RESULTS_DIR` | `result_selector` | yes | make command line, environment, makefile default | none | missing required input | `invalid` | `path_token` | `result_selector` | `usage_error`, exit `2` | `value` | `argv` |
| `explain-run` | `DETAIL` | `enum` | no | make command line, environment, makefile default | `summary` | use `summary` | `omitted` | `trim` | `summary`, `children`, `logs`, `progress`, `accounting`, `performance` | `usage_error`, exit `2` | `value` | `argv` |
| `harness-otel-export` | `HARNESS_OTLP_ENDPOINT` | `url` | yes | make command line | none | missing required input | `invalid` | `trim` | `url` | `usage_error`, exit `2` | `redacted_value` | `argv` |
| `harness-otel-export` | `HARNESS_OTLP_HEADERS_FILE` | `path` | no | make command line | none | omitted | `omitted` | `path_token` | `path` | `usage_error`, exit `2` | `redacted_value` | `argv` |
| `harness-public-target-duration-baselines` | `EVIDENCE_ROOTS_FILE` | `path` | yes | make command line | none | missing required input | `invalid` | `path_token` | `path` | `usage_error`, exit `2` | `value` | `argv` |
| `explain-target` | `TARGET` | `target_name` | yes | make command line, environment, makefile default | none | missing required input | `invalid` | `trim` | `target_name` | `usage_error`, exit `2` | `value` | `argv` |
| `explain-target` | `DETAIL` | `enum` | no | make command line, environment, makefile default | `summary` | use `summary` | `omitted` | `trim` | `summary`, `rows`, `artifacts` | `usage_error`, exit `2` | `value` | `argv` |
| `go-test-duration-baseline-drift`, `browser-e2e-duration-baseline-drift`, `service-backed-make-target-duration-baseline-drift`, `harness-smoke-duration-baseline-drift`, `scheduler-event-order-drift`, `scheduler-summary-timing-drift` | `RESULTS_DIR` | `result_selector` | no | make command line, environment, makefile default | `current-run` | use `current-run` | `omitted` | `path_token` | `result_selector` | `usage_error`, exit `2` | `value` | `argv` |
| `go-vulncheck` | `GOVULNCHECK_DB` | `path` | no | make command line, environment, makefile default | none | omitted | `omitted` | `path_token` | `path` | `usage_error`, exit `2` | `value` | `runtime_env` |
| `build-server-harness` | `SERVER_HARNESS_BIN` | `path` | no | make command line, environment, makefile default | `$(CURDIR)/server-harness` | use `$(CURDIR)/server-harness` | `invalid` | `path_token` | `path` | `configuration_error`, exit `2` | `source_and_value` | `none` |
| `build-operator` | `OPERATOR_BIN` | `path` | no | make command line, environment, makefile default | `$(CURDIR)/operator` | use `$(CURDIR)/operator` | `invalid` | `path_token` | `path` | `configuration_error`, exit `2` | `source_and_value` | `none` |

`fixture-report` remains a `human_summary` target by default. `JSON=1` selects the target-local diagnostic JSON path and is not equivalent to `CARTULARY_OUTPUT_MODE=machine`. When `JSON=1`, stdout MUST be exactly one `cartulary.fixture_report.v1` JSON object followed by one LF, and stderr follows the Section 7 failure budget for `human_summary` targets. `CARTULARY_OUTPUT_MODE=machine make fixture-report` MUST continue to fail before child work under Section 7.2 unless a later adopted registry row changes the target's output class.

**TH-HARNESS-REQ-113**
Undeclared public harness inputs MUST have one shared result:

| Caller input class | Required behavior |
| --- | --- |
| Undeclared wrapper CLI flag | Reject before child work with `failure_reason=usage_error`, exit `2`. |
| Undeclared public harness Make variable supplied on the Make command line | Reject before child work with `failure_reason=usage_error`, exit `2`. |
| Undeclared inherited environment variable | Ignore for resolution and strip from child process environments. |
| Caller override of manifest/internal fields | Reject before child work with `failure_reason=configuration_error`, exit `2`. |

Manifest and internal fields include at least `TASK_SURFACE_MANIFEST`, `CARTULARY_TASK_SURFACE_MANIFEST`, `EXECUTION_TOPOLOGY_MANIFEST`, `CARTULARY_EXECUTION_TOPOLOGY_MANIFEST`, `SCHEDULER_MANIFEST`, and `CARTULARY_OPERATOR_BIN` when supplied through public Make command-line variables. Script-level environment fallbacks such as broad manifest-path overrides, broad passthrough argument strings such as `VITEST_FLAGS`, or unbounded threshold variables are non-canonical implementation inputs unless a public target row declares a bounded `input_contract` entry.

For public `frontend-unit` evidence, `VITEST_MAX_WORKERS` is the only current public Vitest input. `VITEST_FLAGS` MUST NOT be accepted from the Make command line for `frontend-unit`; inherited `VITEST_FLAGS` from the caller environment MUST be stripped before child Vitest execution and MUST NOT narrow the canonical runner report. Filtered Vitest diagnostics, when needed, are private developer commands outside public harness evidence.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-002, TH-HARNESS-AC-003

| Type token | Valid values |
| --- | --- |
| `enum` | One of the row's `values[]` tokens after normalization. |
| `exact_1_bool` | Exact `1` when true; empty string is false only when the row says `empty_string=false`. |
| `owner_id` | One active owner ID satisfying TH-HARNESS-REQ-012. |
| `test_row_ids` | Comma-separated active row IDs parsed by TH-HARNESS-REQ-117. |
| `target_name` | A target name present in the task-surface manifest. |
| `run_id` | `run_id_v1` from Section 6.2. |
| `result_selector` | Existing result root or retained run-root path accepted by the target. |
| `path` | Filesystem path token; path existence and safety follow the Section 5.3 matrix row's `Values/bounds` and `Invalid behavior` cells unless a later adopted row narrows them. |
| `positive_integer` | Decimal integer greater than zero and inside row bounds when declared. |
| `positive_decimal` | Decimal number greater than zero and inside row bounds when declared. |
| `task_surface_report_args` | Empty string, `--all`, `--check`, `--check --all`, or `--all --check`. |

### 5.4 Empty-String Rules

| Variable family                              | Empty string behavior                                                                |
| -------------------------------------------- | ------------------------------------------------------------------------------------ |
| Output mode                                  | Treated as omitted; default resolution applies.                                      |
| Result root                                  | Invalid.                                                                             |
| Run ID                                       | Invalid.                                                                             |
| Boolean exact-`1` flags                      | Empty string is false.                                                               |
| Integer limits                               | Treated as omitted; default applies.                                                 |
| Required DSN, endpoint, credential, or token | Invalid.                                                                             |
| Optional config path                         | Treated as omitted.                                                                  |
| Comma-separated lists                        | Empty string is an empty list only when the variable row says so; otherwise invalid. |

### 5.5 Configuration Variable Table

| Name or family                                                                                  | Scope                   | Type and valid values                                                                                                 | Default                                                                                   | Allowed sources                                 | Empty-string behavior                                   | Normalization                                                                                                 | Invalid behavior                                                                   | Summary emission                                   |
| ----------------------------------------------------------------------------------------------- | ----------------------- | --------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- | ----------------------------------------------- | ------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- | -------------------------------------------------- |
| `CARTULARY_TEST_RESULTS_DIR`                                                                    | global                  | `result_root_path_v1` from Section 6                                                                                  | `.cartulary/test-results`                                                                 | Make variable, env, default                     | invalid                                                 | Path normalized by Section 6                                                                                  | `configuration_error`, exit `2`                                                    | normalized path and cleanup scope                  |
| `CARTULARY_TEST_RUN_ID`                                                                         | global                  | `run_id_v1` from Section 6                                                                                            | generated by Section 6                                                                    | Make variable, env, default                     | invalid                                                 | Grammar validation only                                                                                       | `configuration_error`, exit `2`                                                    | run ID and whether generated                       |
| `CARTULARY_OUTPUT_MODE`                                                                         | global                  | `quiet`, `summary`, `ci`, `verbose`, `debug`, `machine`                                                               | resolved by Section 7                                                                     | Make variable, env, default                     | omitted                                                 | lower-case exact token                                                                                        | `configuration_error`, exit `2`                                                    | resolved output mode and source                    |
| `VERBOSE`                                                                                       | global                  | exact `1` means verbose request; any other value false                                                                | false                                                                                     | Make variable, env, default                     | false                                                   | exact string compare                                                                                          | non-`1` ignored as false                                                           | boolean source when true                           |
| `CI_VERBOSE`                                                                                    | global                  | exact `1` means CI-output request; any other value false                                                              | false                                                                                     | Make variable, env, default                     | false                                                   | exact string compare                                                                                          | non-`1` ignored as false                                                           | boolean source when true                           |
| `CI`                                                                                            | global                  | exact `1` marks CI environment                                                                                        | false                                                                                     | env, default                                    | false                                                   | exact string compare                                                                                          | non-`1` ignored as false                                                           | boolean source when true                           |
| Scheduler resource limits                                                                       | scheduler               | positive integer `1..256` unless resource row declares narrower bound                                                 | resource registry default                                                                 | CLI flag, Make variable, env, manifest, default | omitted                                                 | decimal parse with no separators                                                                              | `configuration_error`, exit `2`                                                    | normalized limit and source                        |
| `--resource-limit name=value`                                                                   | scheduler               | declared resource name and positive integer value                                                                     | none                                                                                      | CLI flag only                                   | invalid                                                 | name exact; value decimal                                                                                     | `usage_error` for malformed flag, `configuration_error` for invalid declared value | normalized override                                |
| `CARTULARY_MACHINE_CACHE_DIR`                                                                  | global machine state    | absolute filesystem path outside the repository                                                                       | `$XDG_CACHE_HOME/cartulary` when `XDG_CACHE_HOME` is absolute and nonempty; otherwise `$HOME/.cache/cartulary`; no `/tmp` or repository fallback | Make variable, env, default | invalid | realpath-aware normalization against the repository boundary | `configuration_error`, exit `2` | normalized root and source |
| `GO`, `GO_CACHE_DIR`, `GO_MOD_CACHE_DIR`, `GO_TMP_DIR`                                         | toolchain               | executable path for `GO`; absolute external filesystem paths for the three state directories                          | Go launcher auto-discovered; `<machine-cache>/go/build`, `<machine-cache>/go/mod`, and `<machine-cache>/go/tmp` | Make variable, env, default | invalid for paths, omitted for executable | realpath-aware path normalization; state directories must be outside the repository and pairwise non-overlapping | `configuration_error`, exit `2` | launcher and normalized state paths unless redacted |
| `GOCACHE`, `GOMODCACHE`, `GOTMPDIR`                                                            | toolchain child env     | exact projections of the resolved Cartulary path variables                                                            | resolved `GO_CACHE_DIR`, `GO_MOD_CACHE_DIR`, and `GO_TMP_DIR` | internal child environment only | invalid | no caller alias translation or fallback | caller Make override is `usage_error`, exit `2`; inherited values are stripped or overwritten | omitted |
| `GO_TOOLCHAIN`, `GOTOOLCHAIN`                                                                  | toolchain               | exact reviewed Go toolchain token                                                                                     | `go1.26.5`, projected from `tools/toolchain_pins.json` and `go.mod`                       | internal Make projection only                  | invalid                                                 | exact token; `GOTOOLCHAIN` is forced to the repository pin for Make-owned Go work                            | `configuration_error`, exit `2`                                                    | launcher version, exact effective version, and source |
| `NODE_VERSION`, `PNPM_VERSION`, `NODE_RUNTIME_DIR`, `NODE_BIN`, `PNPM`, `COREPACK_HOME`, `PATH` | toolchain               | version token or filesystem path                                                                                      | Node `24.15.0`, pnpm `10.33.0`, repo-local `tmp/node-runtime`                             | Make variable, env, default                     | invalid for paths/versions unless row-specific optional | exact version token; path normalization                                                                       | `configuration_error`, exit `2`                                                    | version and runtime path                           |
| `CARTULARY_READINESS_CACHE_DIR`, `CARTULARY_READINESS_DISABLE_CACHE`, `CARTULARY_FORCE_REINSTALL` | harness cache           | repo-local path for cache dir; exact `1` for disable or force reinstall                                                | `.cache/cartulary/readiness`; false; false                                                 | Make variable, env, default                     | invalid for path; false for boolean flags                | path normalization; exact string compare for flags                                                           | invalid path is `configuration_error`; non-`1` flags are false                     | cache state and record path only                   |
| `CARTULARY_BUILD_CACHE_DIR`, `CARTULARY_BUILD_CACHE_DISABLE`, `CARTULARY_FORCE_REBUILD`          | harness cache           | repo-local path for cache dir; exact `1` for disable or force rebuild                                                  | `.cache/cartulary/build-artifacts`; false; false                                           | Make variable, env, default                     | invalid for path; false for boolean flags                | path normalization; exact string compare for flags                                                           | invalid path is `configuration_error`; non-`1` flags are false                     | cache state and record path only                   |
| `CONFIG_FILE`, `CARTULARY_CONFIG_FILE`                                                          | app runtime             | config file path                                                                                                      | `configs/dev/config.toml` for local/dev/browser targets                                   | Make variable, env, config binding, default     | omitted                                                 | path normalization; `CARTULARY_CONFIG_FILE` wins only inside application runtime when both are passed through | harness invalid path: `configuration_error`; app invalid config: target failure    | path, not file contents                            |
| `TEST_SERVICES_BIN`, `CARTULARY_TEST_SERVICES_BIN`                                              | service suite           | executable path                                                                                                       | `tmp/toolbin/cartulary-test-services`                                                     | Make variable, env, default                     | invalid                                                 | path normalization                                                                                            | `configuration_error`, exit `2`                                                    | normalized path                                    |
| `CARTULARY_OPERATOR_BIN`                                                                         | runtime binary          | scheduler-owned executable path for operator scenario Go tests                                                        | produced by `build-operator` from `OPERATOR_BIN`; current default `operator`              | scheduler/runtime wiring only; not public Make command line | invalid for canonical scheduler-selected operator scenario work | path normalization; existing regular executable file; symlinks rejected | missing, empty, non-regular, non-executable, or caller command-line override is `configuration_error`, exit `2`; build-artifact digest/provenance mismatch is `artifact_error`, exit `11` | source, normalized path, producer target, file digest, build-artifact reference |
| `CARTULARY_TEST_SERVICES_MODE`                                                                  | service suite           | exact `owned` or `attach`                                                                                            | `owned`                                                                                   | env, Make variable, default                     | invalid                                                 | exact token                                                                                                   | `usage_error`, exit `2`                                                             | mode only                                          |
| `CARTULARY_TEST_SERVICES_SESSION_FILE`                                                          | service suite           | absolute external regular-file path reached without symlink traversal; accepted only with mode `attach`             | `${CARTULARY_MACHINE_CACHE_DIR}/test-services/session.json`                              | env, Make variable, default                     | invalid outside attach mode                              | owner, mode, containment, and no-follow validation                                                               | `configuration_error`, exit `2`                                                    | normalized path; contents redacted                |
| `CARTULARY_TEST_SUITE_ID`, `CARTULARY_TEST_TARGET`                                              | service suite           | non-empty ASCII token; suite ID is 24 lowercase hex in owned mode                                                     | generated in owned mode                                                                   | service manifest, env in attach mode            | invalid in attach mode                                  | exact grammar validation                                                                                      | `configuration_error`, exit `2`                                                    | suite ID, target                                   |
| Postgres attach set                                                                             | service suite           | `CARTULARY_PGTEST_ADMIN_DSN`, `CARTULARY_PGTEST_DSN_TEMPLATE` containing `{database}`, `CARTULARY_PGTEST_TEMPLATE_DB`, optional `CARTULARY_PGTEST_SCHEMA_HASH` | none                                                                                      | env, Make variable                              | invalid                                                 | DSN redacted; template exact placeholder validation; schema hash exact match when supplied                     | partial or malformed set or schema-hash mismatch: `configuration_error`, exit `2`  | redacted DSN, attach mode, schema hash             |
| Fixture capability                                                                              | scheduler/broker        | `none`, `postgres_transaction`, `postgres_dedicated`, `postgres_migration`, `object_store_namespace`, `managed_process`, or `browser_stack` | exact catalog or authored work-unit declaration; no implicit service-backed fallback | catalog, work-graph owner                       | invalid                                                 | exact lower-case token; broker derives private adapter policy only after validation                           | missing or unknown capability is `configuration_error`, exit `2`, before child work | capability, lease ID, isolation scope, and cleanup state |
| Object-store S3 attach set                                                                      | service suite           | endpoint, access key, secret key, secure bool through `CARTULARY_S3TEST_*`                                            | none                                                                                      | env, Make variable                              | invalid for required members                            | endpoint normalized; credentials redacted; secure bool exact `true`/`false` or `1`/`0`                        | partial set or invalid bool: `configuration_error`, exit `2`                       | endpoint, secure flag, credential redaction marker |
| `CARTULARY_TEST_SERVICES_WEB_E2E_CLEANUP_WORKERS`                                               | browser/service cleanup | integer `1..16`                                                                                                       | `4`                                                                                       | Make variable, env, default                     | omitted                                                 | decimal parse                                                                                                 | invalid value falls back to `4` and records warning                                | resolved value and warning when fallback used      |
| Compose env                                                                                     | local services          | `CARTULARY_COMPOSE_FILE`, ready timeouts, `OBJECT_STORE_BUCKET`                                                       | `docker-compose.dev.yml`, Postgres `180s`, object-store `120s`, bucket `cartulary`        | Make variable, env, default                     | omitted for optional values                             | path and duration normalization                                                                               | missing Docker/Compose: Section 9 class                                            | non-secret values                                  |
| Browser owned-stack env                                                                         | browser                 | runtime roots, origins, backend/frontend port overrides, built frontend preview artifact                               | dynamic ports; frontend served from `apps/web/dist` by non-watching preview; `build-web` is a first-class prerequisite | Make variable, env, manifest, default           | invalid for required values or missing built frontend artifact | origin values lower-case scheme and host; ports decimal; backend readiness must prove owned process identity through the token-protected test runtime identity route; frontend readiness must report `frontend_mode="preview"` and `frontend_command_kind="vite-preview"`; service-backed frontend auto-allocation uses stage-owned CORS windows outside the default host ephemeral-client range, with stateful browser work in `19100-19199` and current non-stateful browser work in `19000-19099`; frontend port selection MUST use `CARTULARY_BROWSER_STAGE` or scheduler session metadata rather than target-name substring matching | config or port collision: `resource_conflict`; missing preview artifact or invalid config: `configuration_error` | origins, ports, runtime root, ownership proof, frontend mode |
| `PLAYWRIGHT_WORKERS`, worker count/index/offset envs                                            | browser                 | positive integers; worker offset `0..1024`                                                                            | Make `3`; shared config fallback `2`; direct isolated offset `0`; scheduled browser groups require scheduler-owned count and offset | Make variable, env, default, scheduler manifest | omitted only for direct isolated browser invocation      | decimal parse                                                                                                 | `configuration_error`, exit `2`                                                    | worker counts and scheduled worker slot range      |
| `VITEST_MAX_WORKERS`                                                                            | frontend unit           | positive integer `1..16`                                                                                              | matching Vitest unit default from its runner contract | Make variable, env, default, graph unit | invalid                                                 | decimal parse                                                                                                 | `usage_error`, exit `2` for public target input                                     | worker count and resource-claim effect            |
| Webserver-backed shard env                                                                      | browser                 | required grep/file values declared by target, plus selected-test artifact path for manifest-aware shard verification   | none                                                                                      | Make variable, manifest                         | invalid                                                 | exact string after JSON/shell decoding; selected tests validated as `cartulary.playwright_manifest_selection.v1` | missing required value: `configuration_error`, exit `2`                            | declared shard IDs as compatibility fallback; selected `(row_id,file,title)` entries when artifact-backed |
| `CARTULARY_ENABLE_TEST_ROUTES`                                                                  | reset/browser           | exact `1` enables test routes                                                                                         | disabled                                                                                  | Make variable, env, default                     | false                                                   | exact string compare                                                                                          | non-`1` means disabled                                                             | enabled boolean                                    |
| `CARTULARY_TEST_ROUTE_TOKEN`                                                                    | test-controls/browser   | non-empty opaque string with at least 128 bits entropy                                                                | generated by harness stack when test routes are enabled                                   | harness generated, env for attach mode          | invalid                                                 | not normalized                                                                                                | missing/low-entropy token: `configuration_error`, exit `2` before stack use        | redaction token only                               |
| Object-store runtime env                                                                        | app runtime             | `CARTULARY_S3_OBJECT_PRIMARY_*` endpoint, credentials, secure bool, bucket                                            | browser/dev SeaweedFS S3 local values                                                     | Make variable, env, config binding, default     | invalid for required members                            | endpoint normalized; credentials redacted                                                                     | app startup/reset failure according to Section 12                                  | redacted credential fields                         |
| Runtime root envs                                                                               | app runtime             | `CARTULARY__ROOTS__*__PATH` filesystem paths                                                                          | browser stack creates under runtime root                                                  | Make variable, env, config binding, default     | invalid                                                 | path normalization                                                                                            | invalid/unwritable path: `configuration_error` or app startup failure              | normalized path                                    |
| `CARTULARY_HARNESS_REPO_ROOT`, `CARTULARY_HARNESS_SCRATCH_ROOT`, `TMPDIR`                       | harness scratch         | filesystem path                                                                                                       | `${TMPDIR:-/tmp}/cartulary-harness-scratch`                                               | env, default                                    | invalid for explicit scratch                            | path normalization; scratch root must be outside repo                                                         | in-repo scratch root: `configuration_error`, exit `2`                              | normalized scratch root                            |
| `CARTULARY_HARNESS_SUITE_RUNTIME_ROOT`, `CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID`, `CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID` | private suite runtime | absolute owner-only external directory plus opaque lease and owning-run identities | scheduler-created below the validated harness scratch namespace | internal scheduler projection only | invalid or omitted for managed service/browser work | realpath, owner, mode, containment, symlink-component, and exact ownership-marker validation | `configuration_error` before child use; cleanup proof mismatch is `cleanup_error` | opaque lease identity and cleanup result only |
| `CARTULARY_CLEANUP_DRY_RUN`                                                                     | cleanup                 | exact `1`                                                                                                             | false                                                                                     | Make variable, env, default                     | false                                                   | exact string compare                                                                                          | non-`1` false                                                                      | dry-run boolean                                    |
| `CARTULARY_DESTRUCTIVE_CONFIRM`                                                                 | destructive local reset | enum equal to the target name, currently `db-reset` or `object-store-reset`                                            | none                                                                                      | Make command line only                          | invalid when supplied empty; omitted allowed only for dry-run | trim exact token                                                                                              | wrong Make command-line token is `usage_error`; inherited-env-only confirmation is ignored and cannot satisfy reset confirmation; missing token on real reset fails before mutation | selected target token                            |
| `LINT_SHELL_STRICT`                                                                             | lint                    | exact `1`; public Make lint targets force strict blocking behavior                                                    | `1` for public Make targets; raw script fallback may default false                       | Make recipe                                      | false outside public Make                            | exact string compare                                                                                          | public Make target overrides ignored by recipe-owned strict value                   | boolean when true                                  |
| `STATICCHECK_CHECKS`                                                                             | static analysis          | closed; not a public Make target input                                                                                | Staticcheck default fixed by the public target row and wrapper                           | raw script only                                  | not applicable                                           | none                                                                                                          | Make command-line use is `usage_error`, exit `2`                                  | none                                               |
| Gosec rule, flag, and pattern variables (`GOSEC_*`, `GOSEC_TARGETED_*`, `GOSEC_AUDIT_*`)         | security                | closed; not public Make target inputs                                                                                 | curated profiles from this NLSpec and task-surface owner inputs                          | raw script only                                  | not applicable                                           | none                                                                                                          | Make command-line use is `usage_error`, exit `2`                                  | retained security profile metadata                 |
| `GOVULNCHECK_DB`                                                                                 | security                | optional Govulncheck vulnerability DB path or endpoint token                                                          | omitted; Govulncheck default DB                                                           | Make variable, env, default                     | omitted                                                 | path-token validation                                                                                         | invalid value is `usage_error`, exit `2`                                           | value                                              |
| `GOVULNCHECK_FLAGS`, `GOVULNCHECK_PATTERNS`                                                      | security                | closed; not public Make target inputs                                                                                 | `-test -json` flags and authored package roots fixed by the public target row and wrapper | raw script only                                  | not applicable                                           | none                                                                                                          | Make command-line use is `usage_error`, exit `2`                                  | none                                               |

## 6. Result Roots, Run IDs, and Artifact Identity

**TH-HARNESS-REQ-150**
A public Make target that emits retained artifacts MUST compute artifact identity as:

```text
run_root = normalize_result_root(CARTULARY_TEST_RESULTS_DIR) / normalize_run_id(CARTULARY_TEST_RUN_ID)
```
Verified by: TH-HARNESS-AC-003, TH-HARNESS-AC-015

Retained raw telemetry captures from `otel-conformance` MUST be retained artifacts owned by the `otel-conformance` target below the normalized run root. They MUST NOT be written below committed golden directories such as `internal/testutil/golden/otel/`.

### 6.1 Result Root Normalization

```text
normalize_result_root(input):
  if input is omitted:
    input = ".cartulary/test-results"
  reject empty string
  reject NUL
  reject path equal to "/" after lexical normalization
  reject path equal to "." after lexical normalization
  reject any caller-supplied segment equal to ".."
  reject backslash on POSIX conformance hosts
  if relative:
    resolve against repository root
  if absolute:
    allow for artifact writing and set cleanup_scope = "external_or_custom"
  create the directory if missing
  create retained run roots and target artifact directories with owner-only permissions on POSIX conformance hosts
  fail with configuration_error if parent is not writable
  fail with configuration_error if a caller-supplied custom result root is world-writable without sticky bit
```

### 6.2 Run-ID Grammar

```text
run_id = 1*96(ALPHA / DIGIT / "-" / "_" / ".")
run_id MUST NOT equal "." or ".."
run_id MUST NOT contain "/"
run_id MUST NOT contain "\\"
run_id MUST NOT contain whitespace
```

When `CARTULARY_TEST_RUN_ID` is omitted, the wrapper MUST generate:

```text
YYYYMMDDTHHMMSSZ-p<PID>
```

`YYYYMMDDTHHMMSSZ` is the UTC wrapper start time. `<PID>` is the decimal process ID of the Make-owned top-level wrapper.
Every public target that retains artifacts MUST freeze the resolved run ID once
at target scope before its first recipe line. Public preflight, prerequisite
coordination, child invocation, summary emission, cleanup, and observability
finalization MUST reuse that exact value; no recipe line may re-evaluate the
default run-ID expression or create a sibling root for the same invocation.

### 6.3 Collision Rules

| Case                                                | Required behavior                                                                        |
| --------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| Omitted run ID                                      | Generate a default run ID.                                                               |
| Caller-supplied run ID path does not exist          | Create it.                                                                               |
| Caller-supplied run ID path exists and is empty     | Reuse it.                                                                                |
| Caller-supplied run ID path exists and is non-empty | Fail before child work with `configuration_error`, exit `2`.                             |
| Generated default run ID collides                   | Append `-n<N>` with the smallest positive decimal `N` that produces a non-existing path. |

### 6.4 Artifact-Proof Rule

**TH-HARNESS-REQ-151**
Retained artifacts prove only the explicit `{result_root, run_id, target}` triple. A newest-run fallback MAY be used for human investigation, but MUST NOT satisfy harness conformance evidence.
Verified by: TH-HARNESS-AC-015

**TH-HARNESS-REQ-152**
Every post-cutover row-bearing evidence artifact MUST identify its schema,
run, unit or target, owner, and sorted duplicate-free selected rows. The
canonical run manifest supplies command, source, system, graph, catalog,
verification, runtime/resource policy, capability-snapshot, and cache-mode
identity. Unit events supply monotonic start, finish, and duration. Every
row-bearing artifact is joined to those authorities by stable IDs and digests;
it MUST NOT duplicate them with an independently computed value. A missing or
mismatched join makes the artifact incompatible, not stale-by-age.

An internal evidence partition MUST retain its own `target_id` while using the command identity of the public semantic command that owns the partition. The `backend-integration-support` partition therefore MUST use `target_id=backend-integration-support` and `command_id=cartulary.harness.command.backend_integration.v1`. This ownership route MUST be explicit catalog-side policy; runtime target-name inference and a new public or private support command identity are forbidden.
Verified by: TH-HARNESS-AC-066

**TH-HARNESS-REQ-153**
`source_snapshot_digest_v2` is SHA-256 over RFC 8785 canonical JSON
describing every Git-tracked and non-ignored untracked executable repository
input, ordered by repository-relative path. Each entry contains path, stable
file-kind or Git-mode identity, and byte digest. Ignored paths, result roots,
cache roots, the complete `docs/` subtree, and every Markdown file
(`.md` or `.markdown`, ASCII case-insensitive) are excluded.

Documentation exclusion MUST use only the normalized repository-relative path
returned by Git and MUST occur before any `lstat`, `stat`, open, read,
`readlink`, or hashing operation for that path. Editing, adding, deleting,
renaming, or rearranging an excluded specification or Markdown guidance file
MUST NOT change the snapshot digest or file count. A retained symlink outside
the excluded documentation surfaces is recorded as a link and its link bytes
are hashed; the digest algorithm MUST NOT follow it. Duplicate, escaping,
unreadable, or unstable executable-input paths MUST fail snapshot
construction.
Verified by: TH-HARNESS-AC-066, TH-HARNESS-AC-067

**TH-HARNESS-REQ-154**
Evidence from different retained roots is compatible only when `owner_id`, source
snapshot digest, test-catalog digest, and verification-routing digest are
identical. Selected-row inventories are target partitions and therefore MAY differ
between gates. Each artifact's selected rows and runtime, resource, and fixture
profile digests MUST exactly equal the current catalog-derived partition for that
target. The union of accepted `(owner_id, target_id, row_id)` records MUST equal the
full applicability set derived by TH-HARNESS-REQ-668. Two candidate artifacts for the
same required target and row are ambiguous and MUST fail. An auditor MUST NOT search
unlisted roots or select a newest candidate.
Verified by: TH-HARNESS-AC-066

**TH-HARNESS-REQ-155**
Evidence freshness has no wall-clock TTL. A matching semantic identity remains fresh regardless of age, subject to the target success, completeness, contamination, and release rules in this NLSpec. Release closure additionally requires a clean worktree. Local dirty-tree evidence is compatible only when its exact source snapshot digest matches the audited worktree.
Verified by: TH-HARNESS-AC-066

**TH-HARNESS-REQ-156**
Owner evidence timestamps MUST be UTC RFC 3339 strings using uppercase `T` and `Z` and exactly three fractional-second digits. `duration_ms` MUST be a non-negative integer measured from a monotonic clock. Wall-clock timestamps MUST NOT be subtracted to determine duration.
Verified by: TH-HARNESS-AC-066

**TH-HARNESS-REQ-157**
Result-root readers MUST apply Section 6 normalization and containment before opening retained evidence. They MUST reject NUL, traversal, symlink escape, non-directory roots, world-writable custom roots without sticky bit, and target artifacts outside the declared run root. Invalid caller path syntax is `usage_error`; a syntactically valid root containing unsafe or incompatible evidence is `artifact_error`.
Verified by: TH-HARNESS-AC-066, TH-HARNESS-AC-067

**TH-HARNESS-REQ-158**
Prepared artifact identity is atomic. A child MAY reuse prepared identity only when
`CARTULARY_HARNESS_IDENTITY_PREPARED=1` and the complete normalized
`{result_root, run_id, target}` tuple is present and validates before any artifact
write. A partial tuple, a marker value other than exact `1`, an undeclared target,
or a target mismatch MUST fail closed with `configuration_error`, exit `2`, before
artifact creation.

A scheduler child that requires an independent artifact identity MUST clear the
complete parent prepared-identity tuple, parent selectors, worker overrides, and
Make command-line override metadata as one operation. The scheduler MUST give that
child a distinct result root contained below the parent-owned diagnostic namespace,
and the child MUST perform ordinary identity preflight. Child diagnostic artifacts
MUST NOT overwrite parent plans, scheduler logs, work-unit results, accounting,
summaries, or cleanup evidence, and MUST NOT substitute for the parent's row-bearing
evidence.
Verified by: TH-HARNESS-AC-003, TH-HARNESS-AC-015, TH-HARNESS-AC-064, TH-HARNESS-AC-071

## 7. Output Modes and Machine Output

**TH-HARNESS-REQ-200**
Public Make targets MUST recognize exactly these output-mode tokens: `quiet`, `summary`, `ci`, `verbose`, `debug`, and `machine`. Unknown output modes MUST fail with `configuration_error` and exit `2` before child work. A recognized mode is accepted only when the target's output class allows it. When a recognized mode is not accepted for that output class, the target MUST fail before child work with the Section 7.2 output-class rejection behavior.
Verified by: TH-HARNESS-AC-002, TH-HARNESS-AC-004, TH-HARNESS-AC-005

### 7.1 Output-Mode Resolution

```text
resolve_output_mode(CARTULARY_OUTPUT_MODE, VERBOSE, CI_VERBOSE, CI, target):
  if CARTULARY_OUTPUT_MODE is present and non-empty:
    return exact token after lower-case validation
  if VERBOSE == "1":
    return "verbose"
  if CI_VERBOSE == "1":
    return "ci"
  if target == "ci" or CI == "1":
    return "ci"
  return "summary"
```

`quiet`, `debug`, and `machine` are selected only by `CARTULARY_OUTPUT_MODE`.

### 7.2 Output Class Matrix

| Output class                       | Public targets                                                                                    | `machine` accepted? | `machine` stdout                                        | `machine` stderr                                                                      | Success artifacts                                         | Failure behavior                                                       |
| ---------------------------------- | ------------------------------------------------------------------------------------------------- | ------------------: | ------------------------------------------------------- | ------------------------------------------------------------------------------------- | --------------------------------------------------------- | ---------------------------------------------------------------------- |
| `summary_with_artifacts`           | Leaf, toolchain, non-aggregate build/lint child, drift, browser-stage, and formatting targets |                 yes | One JSON object using the target registry's stable summary schema plus LF | Empty after wrapper starts; pre-wrapper diagnostics allowed only before JSON emission | Graph-backed targets retain the canonical run artifact set; unchanged wrapper targets retain their declared tool summary | Same declared schema with failure fields and nonzero exit              |
| `service_summary`                  | `db-up`, `db-reset`, `services-up`, `services-down`, `object-store-init`, `object-store-reset` |                 yes | One `cartulary.tool_run_summary.v5` JSON object plus LF | Empty after wrapper starts                                                            | Tool-run summary and service diagnostics for service-owning rows | Same schema with service failure fields                                |
| `aggregate_summary_with_artifacts` | `test-fast`, `test`, `lint`, `ci`, `release-check`, `build`                                       |                 yes | One `cartulary.harness_run_summary.v1` JSON object plus LF | Empty after wrapper starts                                                          | Canonical run manifest, unit events, run summary, target summaries, and unit/row results | Same schema with primary graph or unit failure                         |
| `scheduler_summary_with_artifacts` | `check`, `test-slice`, `service-backed-test-slice`                                                |                 yes | One `cartulary.harness_run_summary.v1` JSON object plus LF | Empty after wrapper starts; no scheduler progress prose                               | Canonical run manifest, unit events, run summary, target summaries, and unit/row results | Same schema with graph, resource, or child failure                      |
| `machine_stdout_json`              | `target-plan-json` and other explicitly declared JSON discovery targets                           |                 yes | One closed target JSON value plus LF                    | Empty on success                                                                      | None unless target declares artifacts                     | Invalid input exits `2`; error JSON only when Section 7.4 declares it |
| `human_summary`                    | `help`, `help-all`, text discovery/explanation targets                                            |                  no | Empty                                                   | Bounded diagnostic allowed on failure                                                 | None unless target row declares diagnostic artifacts      | `machine` rejected as `usage_error`, exit `2`                          |
| `interactive_raw`                  | `dev`                                                                                             |                  no | Empty when `machine` requested                          | Diagnostic allowed                                                                    | None                                                      | `machine` rejected as `usage_error`, exit `2`                          |
| `destructive_human`                | `clean`, `distclean`                                                                              |                  no | Empty when `machine` requested                          | Diagnostic allowed                                                                    | None                                                      | `machine` rejected as `usage_error`, exit `2`                          |

Successful `frontend-fallow-static` `summary` and `quiet` output MUST expose only bounded result and artifact references. Raw Fallow JSON, SARIF, Markdown, stdout, and stderr diagnostics MUST be retained as artifacts rather than emitted in successful summary stdout; `verbose` and `debug` modes MAY expose raw child diagnostics.

### 7.3 Human Output Budgets

| Mode      |             Success stdout budget | Success stderr                                              | Child logs                                       |
| --------- | --------------------------------: | ----------------------------------------------------------- | ------------------------------------------------ |
| `quiet`   |                    At most 1 line | Empty unless failure                                        | No child logs.                                   |
| `summary` |   At most 30 lines and 8192 bytes | Empty                                                       | Retained in artifacts only.                      |
| `ci`      | At most 120 lines and 32768 bytes | Empty unless CI wrapper failure occurs before summary layer | Retained in artifacts; bounded progress allowed. |
| `verbose` |              No fixed line budget | Tool-dependent                                              | May stream child logs.                           |
| `debug`   |              No fixed line budget | Tool-dependent                                              | May stream wrapper telemetry.                    |

### 7.4 Machine Output

**TH-HARNESS-REQ-201**
For every public target whose output class accepts `machine`, stdout MUST be exactly one UTF-8 JSON object followed by one LF, except that `machine_stdout_json` discovery targets MAY emit one closed target JSON value followed by one LF when Section 7.4 defines that value. The JSON payload MAY contain artifact pointers. Stdout MUST NOT be a pointer-only payload, multiple JSON values, scheduler progress prose, child logs, or human summary text.
Verified by: TH-HARNESS-AC-004

**TH-HARNESS-REQ-202**
For every public target whose output class rejects `machine`, setting `CARTULARY_OUTPUT_MODE=machine` MUST fail before child work with `failure_class=config`, `failure_reason=usage_error`, public exit code `2`, empty stdout, and bounded stderr diagnostic.
Verified by: TH-HARNESS-AC-005

For `target-plan-json`, the current closed contract is one
`cartulary.harness_target_plan.v4` JSON object followed by LF. It contains the
selection, graph digest, ordered work units, target projections, and the exact
catalog rows selected by those units, including fixture capability and minimum
tier, service dependencies, and topology-resolved resource claims. IDs and
arrays are ASCII-sorted and duplicate-free. The projection MUST
NOT contain a historical delivery or phase selector, documentation-derived
activation, unresolved executable, ambient shell, inline port allocation, or
implicit fixture path. Unknown `TARGET` is `usage_error`, exit `2`, empty
stdout, and no partial JSON.

## 8. Artifact and Schema Contract

**TH-HARNESS-REQ-250**
A public Make-owned command that declares a stable schema ID MUST emit JSON that validates against the matching normative schema attachment before command success. If required artifact validation fails, the public target MUST fail with `artifact_error` or `scheduler_accounting_error` according to Section 9.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-004, TH-HARNESS-AC-025

Current row producers MUST emit `cartulary.harness_row_result.v2`. Each row
result requires nullable `failure_class`, `failure_reason`, and
`failure_diagnostic`; all three are null on success. A setup failure diagnostic,
when present, MUST validate as `cartulary.harness_test_failure.v1` and MUST agree
with the row's normalized failure class and reason. Duration,
resource, cache, and critical-path projections derive only from canonical unit
events; row results do not invent a parallel timing value. Current target
summaries reference exact row and unit evidence and MUST fail on an absent,
duplicate, historical-schema, or out-of-selection result.

The following table owns harness-public behavioral schema IDs and validation
points only. It is not an exhaustive roster of repository schemas. Schema file
paths are repository attachments, not behavioral owners.
`tools/harness_schema_attachments.json` is the authored machine owner for the
current repository attachment path, classification, producer class, and
validation target of every schema under `tools/schemas`, including schemas
outside the harness-public boundary. It MUST contain exactly the current schema
files; an unregistered file, duplicate ID/path, or missing registered attachment
is invalid. The machine owner supplies attachment mechanics and MUST remain
parity-checked with every harness-public row in this table.

| Schema ID                                       | Repository attachment path                                               | Status            | Producer class           | Required validation point                 |
| ----------------------------------------------- | ------------------------------------------------------------------------- | ----------------- | ------------------------ | ----------------------------------------- |
| `cartulary.harness_artifact_ref.v1`             | `tools/schemas/cartulary.harness_artifact_ref.v1.schema.json`             | present           | Shared schema component  | Whenever a current schema declares structured retained harness artifact references. |
| `cartulary.tool_run_summary.v5`                 | `tools/schemas/cartulary.tool_run_summary.v5.schema.json`                 | present           | Centralized wrappers     | Before wrapper exits.                     |
| `cartulary.harness_work_graph.v5`               | `tools/schemas/cartulary.harness_work_graph.v5.schema.json`               | present           | Graph compiler           | Before any child work.                    |
| `cartulary.harness_capacity_override.v1`        | `tools/schemas/cartulary.harness_capacity_override.v1.schema.json`        | present           | Capability capture       | Before applying a caller override.        |
| `cartulary.harness_capability_snapshot.v1`      | `tools/schemas/cartulary.harness_capability_snapshot.v1.schema.json`      | present           | Capability capture       | Before scheduler admission.               |
| `cartulary.harness_cache_record.v2`             | `tools/schemas/cartulary.harness_cache_record.v2.schema.json`             | present           | Graph cache              | Before reuse or retained publication.     |
| `cartulary.harness_fixture_lease.v4`            | `tools/schemas/cartulary.harness_fixture_lease.v4.schema.json`            | present           | Fixture broker           | Before lease use or cleanup.              |
| `cartulary.harness_run_manifest.v1`             | `tools/schemas/cartulary.harness_run_manifest.v1.schema.json`             | present           | Graph runner             | Before scheduler admission.               |
| `cartulary.harness_unit_event.v2`               | `tools/schemas/cartulary.harness_unit_event.v2.schema.json`               | present           | Graph scheduler          | Before event-stream publication.          |
| `cartulary.harness_unit_result.v1`              | `tools/schemas/cartulary.harness_unit_result.v1.schema.json`              | present           | Graph runner             | Before unit completion is accepted.       |
| `cartulary.harness_row_result.v2`               | `tools/schemas/cartulary.harness_row_result.v2.schema.json`               | present           | Runner adapters          | Before selected-row closure.              |
| `cartulary.harness_test_failure.v1`             | `tools/schemas/cartulary.harness_test_failure.v1.schema.json`             | present           | Go test-support helpers  | Before a setup failure marker is emitted or accepted. |
| `cartulary.harness_run_summary.v1`              | `tools/schemas/cartulary.harness_run_summary.v1.schema.json`              | present           | Event projection         | Before graph command exit.                |
| `cartulary.harness_target_summary.v1`           | `tools/schemas/cartulary.harness_target_summary.v1.schema.json`           | present           | Target projection        | Before graph command exit.                |
| `cartulary.harness_target_plan.v4`              | `tools/schemas/cartulary.harness_target_plan.v4.schema.json`              | present           | Graph diagnostics        | Before machine target-plan output.        |
| `cartulary.harness_evidence_root_manifest.v1`   | `tools/schemas/cartulary.harness_evidence_root_manifest.v1.schema.json`   | present           | Evidence audit caller    | Before canonical owner-evidence audit.    |
| `cartulary.fallow_reachability_owner.v1`        | `tools/schemas/cartulary.fallow_reachability_owner.v1.schema.json`        | present           | Fallow reachability owner | During JSON shape checks and before `frontend-fallow-static` builds its effective config. |
| `cartulary.fallow_static_summary.v2`            | `tools/schemas/cartulary.fallow_static_summary.v2.schema.json`            | present           | Fallow static target     | Before `frontend-fallow-static` success.  |
| `cartulary.vitest_failure_details.v1`           | `tools/schemas/cartulary.vitest_failure_details.v1.schema.json`           | present           | Vitest wrappers          | Before owner summaries consume failure diagnostics. |
| `cartulary.test_target_summary.v4`              | `tools/schemas/cartulary.test_target_summary.v4.schema.json`              | present           | Target summary generator | Before aggregate/run summary consumes it. |
| `cartulary.test_run_summary.v6`                 | `tools/schemas/cartulary.test_run_summary.v6.schema.json`                 | present           | Run summary generator    | Before public aggregate success.          |
| `cartulary.task_surface_owner.v2`               | `tools/schemas/cartulary.task_surface_owner.v2.schema.json`               | present           | Authored task-surface owner validation | Before task-surface, graph, or generated Make projection. |
| `cartulary.contract_family_registry.v3`         | `tools/schemas/cartulary.contract_family_registry.v3.schema.json`         | present           | Contract generator family registry | During JSON shape checks and before `tools/contractgen` emits generated contract roots. |
| `cartulary.fixture_tier_proof.v3`               | `tools/schemas/cartulary.fixture_tier_proof.v3.schema.json`               | present           | Scheduler reporter and fixture-proof validators | Before a retained fixture-tier proof artifact is accepted. |
| `cartulary.verification_registry.v3`            | `tools/schemas/cartulary.verification_registry.v3.schema.json`            | present           | Verification registry    | Before catalog compilation or evidence routing. |
| `cartulary.verification_contract.v3`            | `tools/schemas/cartulary.verification_contract.v3.schema.json`            | present           | Verification owner       | Before catalog compilation or evidence routing. |
| `cartulary.test_owner_registry.v1`              | `tools/schemas/cartulary.test_owner_registry.v1.schema.json`              | present           | Test catalog owner       | Before owner manifest loading. |
| `cartulary.test_family_manifest.v6`             | `tools/schemas/cartulary.test_family_manifest.v6.schema.json`             | present           | Test family owners       | Before selection, topology generation, or audit. |
| `cartulary.test_runner_registry.v1`             | `tools/schemas/cartulary.test_runner_registry.v1.schema.json`             | present           | Runner registry          | Before selector resolution or adapter invocation. |
| `cartulary.test_owner_explanation.v3`           | `tools/schemas/cartulary.test_owner_explanation.v3.schema.json`           | present           | Owner diagnostics        | Before target-local JSON output. |
| `cartulary.task_guide_summary.v3`               | `tools/schemas/cartulary.task_guide_summary.v3.schema.json`               | present           | Task guidance            | Before target-local JSON output. |
| `cartulary.test_catalog_check_summary.v2`       | `tools/schemas/cartulary.test_catalog_check_summary.v2.schema.json`       | present           | Catalog validator        | Before private catalog-check success. |
| `cartulary.executable_input_policy.v1`          | `tools/schemas/cartulary.executable_input_policy.v1.schema.json`          | present           | Executable input boundary owner | Before executable-input policy validation. |
| `cartulary.govulncheck_findings.v1`             | `tools/schemas/cartulary.govulncheck_findings.v1.schema.json`             | present           | Govulncheck wrapper      | Before failure classification or target-summary security rollup consumes findings. |
| `cartulary.test_services.lease.v1`              | `tools/schemas/cartulary.test_services.lease.v1.schema.json`              | present           | Service suite            | Before attach or cleanup relies on lease. |
| `cartulary.test_services.lifecycle.v2`          | `tools/schemas/cartulary.test_services.lifecycle.v2.schema.json`          | present           | Service suite            | During service lifecycle JSONL validation. |
| `cartulary.test_services.local_session.v1`      | `tools/schemas/cartulary.test_services.local_session.v1.schema.json`      | present           | Local session manager    | Before status, attachment, or session cleanup. |
| `cartulary.test_services.local_session_status.v1` | `tools/schemas/cartulary.test_services.local_session_status.v1.schema.json` | present         | Local session manager    | Before redacted status output. |
| `cartulary.test_services.scope.v2`              | `tools/schemas/cartulary.test_services.scope.v2.schema.json`              | present           | Service suite            | Before scheduler failure propagation consumes bounded service-suite diagnostics. |
| `cartulary.object_store_readiness_diagnostic.v1` | `tools/schemas/cartulary.object_store_readiness_diagnostic.v1.schema.json` | present         | Service suite            | Before a known object-store readiness extension is retained or consumed. |
| `cartulary.test_services.start_result.v1`       | `tools/schemas/cartulary.test_services.start_result.v1.schema.json`       | private           | Service suite broker     | Before scheduler fixture acquisition consumes a test-services start result. |
| `cartulary.test_services.journal_event.v1`      | `tools/schemas/cartulary.test_services.journal_event.v1.schema.json`      | present           | Service suite            | Before a completed producer journal record is collated. |
| `cartulary.test_services.resource_ledger.v1`    | `tools/schemas/cartulary.test_services.resource_ledger.v1.schema.json`    | private           | Service suite            | Before exact owned-resource cleanup or stale recovery. |
| `cartulary.test_services.browser_admission.v1`  | `tools/schemas/cartulary.test_services.browser_admission.v1.schema.json`  | present           | Browser session lifecycle | Before browser service admission is accepted. |
| `cartulary.web_e2e_stack.v6`                    | `tools/schemas/cartulary.web_e2e_stack.v6.schema.json`                    | present           | Browser session lifecycle | Before browser target starts Playwright. |
| `cartulary.web_e2e_backend_generation.v1`       | `tools/schemas/cartulary.web_e2e_backend_generation.v1.schema.json`       | present           | Browser reset lifecycle  | Before a replacement backend is attached. |
| `cartulary.browser_startup_event.v1`             | `tools/schemas/cartulary.browser_startup_event.v1.schema.json`             | present           | Browser session lifecycle | For each append-only startup transition. |
| `cartulary.browser_startup_diagnostics.v2`       | `tools/schemas/cartulary.browser_startup_diagnostics.v2.schema.json`       | present           | Browser session lifecycle | Once at terminal ready or failed state. |
| `cartulary.browser_group_result.v5`              | `tools/schemas/cartulary.browser_group_result.v5.schema.json`              | present           | Browser evidence adapter | Before browser group evidence is accepted. |
| `cartulary.browser_target_result.v2`             | `tools/schemas/cartulary.browser_target_result.v2.schema.json`             | present           | Browser evidence finalizer | Before browser target evidence is accepted. |
| `cartulary.local_object_store_proxy_start_attempt.v1` | `tools/schemas/cartulary.local_object_store_proxy_start_attempt.v1.schema.json` | present | Local development proxy lifecycle | Before a startup attempt is recovered or promoted. |
| `cartulary.local_object_store_proxy_lease.v1`    | `tools/schemas/cartulary.local_object_store_proxy_lease.v1.schema.json`    | present           | Local development proxy lifecycle | Before reuse or signaling. |
| `cartulary.local_object_store_proxy_health.v1`   | `tools/schemas/cartulary.local_object_store_proxy_health.v1.schema.json`   | present           | Local development proxy lifecycle | During ownership and configuration proof. |
| `cartulary.test.runtime_identity.v1`             | `tools/schemas/cartulary.test.runtime_identity.v1.schema.json`             | present           | Browser stack            | During backend identity readiness probing. |
| `cartulary.test.database_reset_diagnostic.v1`   | `tools/schemas/cartulary.test.database_reset_diagnostic.v1.schema.json`   | present           | Recovery reset controller | Before database reset success or failure is accepted. |
| `cartulary.browser_reset_attempt.v1`            | `tools/schemas/cartulary.browser_reset_attempt.v1.schema.json`            | present           | Browser reset lifecycle  | Before a browser reset unit reaches a terminal state. |
| `cartulary.test.clock_control.v1`               | `tools/schemas/cartulary.test.clock_control.v1.schema.json`               | present           | Test clock route         | Before a fixed, offset, reset, or state clock-control response is accepted. |
| `cartulary.test.public_error_fault.v1`          | `tools/schemas/cartulary.test.public_error_fault.v1.schema.json`          | present           | Browser stack            | Before an armed public-error fault is accepted. |
| `cartulary.test.network_flow_fault_control.v1`  | `tools/schemas/cartulary.test.network_flow_fault_control.v1.schema.json`  | present           | Network Flow fault-control route | Before an armed Network Flow commit or worker fault is accepted. |
| `cartulary.test.network_flow_randomness_control.v1` | `tools/schemas/cartulary.test.network_flow_randomness_control.v1.schema.json` | present       | Network Flow randomness-control route | Before an armed deterministic Network Flow random stream is accepted. |
| `cartulary.test.network_flow_auth_transition_control.v1` | `tools/schemas/cartulary.test.network_flow_auth_transition_control.v1.schema.json` | present | Network Flow auth-transition control route | Before an armed Network Flow auth-transition control is accepted. |
| `cartulary.test.network_flow_audit_assertion_control.v1` | `tools/schemas/cartulary.test.network_flow_audit_assertion_control.v1.schema.json` | present | Network Flow audit-assertion control route | Before an armed Network Flow audit-count or replay assertion is accepted. |
| `cartulary.fixture_report.v1`                   | `tools/schemas/cartulary.fixture_report.v1.schema.json`                   | present           | Fixture report target    | Before machine JSON is emitted.           |
| `cartulary.network_flow_fixture_manifest.v2`    | `tools/schemas/cartulary.network_flow_fixture_manifest.v2.schema.json`    | present           | Network Flow fixture manifest validator | Before a Network Flow fixture manifest is selected for behavior execution. |
| `cartulary.network_flow_fixture_scenario.v2`    | `tools/schemas/cartulary.network_flow_fixture_scenario.v2.schema.json`    | present           | Network Flow fixture scenario validator | Before a Network Flow fixture scenario is selected for behavior execution. |
| `cartulary.network_flow_timezone_ruleset_provenance.v2` | `tools/schemas/cartulary.network_flow_timezone_ruleset_provenance.v2.schema.json` | present | Network Flow timezone provenance validator | During JSON shape checks and before timestamp fixtures are accepted. |
| `cartulary.agent_finalize_summary.v3`           | `tools/schemas/cartulary.agent_finalize_summary.v3.schema.json`           | present           | Agent finalizer          | Before `agent-finalize` exits.            |
| `cartulary.frontend_visual_fixture_registry.v5` | `tools/schemas/cartulary.frontend_visual_fixture_registry.v5.schema.json` | present           | Semantic frontend visual fixture registry and one-to-one design-contract projection validation | During JSON shape checks, catalog checks, and visual readiness validation. |
| `cartulary.frontend_claim_publication_review.v1` | `tools/schemas/cartulary.frontend_claim_publication_review.v1.schema.json` | present          | Conditional frontend claim-publication review metadata; no default target emits it | Before any future or explicit frontend claim-review artifact is accepted as Core 05-routed release evidence. |
| `cartulary.otel_conformance_summary.v1`         | `tools/schemas/cartulary.otel_conformance_summary.v1.schema.json`         | present           | OpenTelemetry conformance target | Before `otel-conformance` success. |

Adoption and continued conformance for
`cartulary.testing_harness.current.v3` require live repository verification of
every row whose `Status` is `present`. Each declared attachment path MUST exist,
parse as JSON Schema, reject unknown top-level fields unless it declares an
explicit extension container, and validate positive and negative fixtures.
Missing or malformed attachments make v3 nonconforming. A schema absent from
this table is unsupported current input; old retained JSON remains manually
inspectable only.

Owner slices compile directly to `cartulary.harness_work_graph.v5` and accept no
separate retained-plan schema. Validation covers owner and selection mode,
requested and resolved rows, target/command consistency, unit identity and
semantic digests, dependency closure, resource claims, runtime binaries,
fixtures, cache policy, expected evidence, and finalizers. Every input-valid
invocation validates the graph before setup. Rejected input or an empty
selection MUST NOT retain a graph or start child work.

Current conformance MUST be proven from current schema-owned artifacts. Historical artifact inspection is non-conformance troubleshooting and is not an acceptance criterion for Sections 1-17.

**TH-HARNESS-REQ-251**
Schema-owned artifacts MUST be closed by default. Unknown top-level fields are invalid unless the schema declares an explicit extension container.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-025

**TH-HARNESS-REQ-252**
Every retained summary artifact MUST include normalized `result_root`, `run_id`, `run_root`, `target`, `output_mode`, public `exit_code`, primary `failure_class`, primary `failure_reason`, `started_at`, and `completed_at`. Timestamps MUST be RFC3339 UTC strings with non-null values.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-015

Public and machine summaries MUST serialize `run_root` once per summary and MUST
express every structured retained harness artifact reference relative to that
`run_root`. A structured retained harness artifact reference MUST validate against
`cartulary.harness_artifact_ref.v1` and MUST contain `role`, `path_kind`, and `path`.
`path_kind` is exactly `file` or `directory`. A file reference MUST also contain one
`format` from `json`, `jsonl`, `log`, `markdown`, `sarif`, or `text`; a directory
reference MUST NOT contain `format`. Paths use normalized POSIX separators, MUST be
non-empty and relative to `run_root`, and MUST reject absolute paths, empty segments,
`.` segments, and `..` traversal.

Repository source paths are not retained harness artifact references. A schema that
needs both retained evidence and repository sources MUST use separate `artifact_refs`
and `source_refs` or `source_files` fields. Repository source paths MUST be normalized
repo-relative paths and MUST NOT be dereferenced by retained-run diagnostics as though
they were run artifacts.

Before dereferencing a retained harness artifact reference, a consumer MUST prove
run-root containment, reject symlink path components, require the declared filesystem
kind, and apply its owner-defined bounded read policy. A mismatch is
`failure_class=artifact`, `failure_reason=artifact_error`, public exit `11`. Consumers
MUST validate all selected references before emitting referenced content.

The v3 execution schema cut is atomic. Its graph, capability, cache, lease,
manifest, event, unit/row result, run summary, target summary, and target-plan
families are the only current versions of those families. Non-graph public tool
summaries retain `cartulary.tool_run_summary.v5`; service lifecycle retains its
separately listed current schema. Current producers and consumers MUST NOT
dual-emit, dual-read, alias, or translate replaced versions.

**TH-HARNESS-REQ-253**
Every schema-owned artifact MUST include `schema_id`. A schema-owned artifact MAY include `extensions` only when its schema declares that field. When present, `extensions` MUST be an object keyed by reverse-DNS or `cartulary.*` extension keys. Consumers MUST ignore unknown extension keys and MUST NOT derive required behavior from an unknown extension key. Adding a new required top-level member or changing the meaning of an existing member requires a new schema ID. Extension data is supplemental only; any value required for conformance, drift, timing, cleanup, scheduling, or failure classification MUST be a declared schema member.

Supplemental service-backed extension data under `extensions["cartulary.service_backed"]` MUST normalize extension-level `readiness_status`, `teardown_status`, and `leak_status` to `pass`, `fail`, or `unknown`. These extension rollups MUST be derived from canonical service lifecycle artifacts and MUST NOT expose raw lifecycle tokens such as `succeeded`, `cleanup_failed`, or `skipped_no_lease` as pass/fail status fields. The schema-owned scheduler and service artifacts remain authoritative when extension data is absent or `unknown`.

Supplemental security extension data under `extensions["cartulary.security"]` MAY summarize scanner artifacts retained by the current target. When that extension contains a Govulncheck rollup, it MUST contain `govulncheck.status`, `govulncheck.finding_count`, `govulncheck.blocking_count`, and `govulncheck.blocking_vulnerability_ids` derived only from the current target's schema-valid `cartulary.govulncheck_findings.v1` artifacts. Omission is caller-interchangeable because the findings artifact and normalized failure fields remain authoritative.
Verified by: TH-HARNESS-AC-000

**TH-HARNESS-REQ-254**
Generated-drift replay MUST be driven by a declared scratch-input manifest. The drift checker MUST copy every declared generator runtime input into its scratch tree before invoking generation, including shared harness helper scripts used by generator wrappers. A missing declared scratch input MUST fail as an artifact error with a diagnostic naming the missing path, not as a tool-specific module, import, or Make lookup failure.
Verified by: TH-HARNESS-AC-000

**TH-HARNESS-REQ-255**
`browser-e2e-a11y` MUST emit exactly one retained `cartulary.frontend_accessibility_summary.v4` artifact for a completed accessibility target attempt. The artifact MUST contain active catalog accessibility rows only and MUST be a JSON object with `schema_id`, `rows[]`, `scenarios[]`, `keyboard_matrix[]`, `state_communication_checks[]`, `contrast_checks[]`, `violations[]`, and `artifact_refs[]`. Scenario status fields MUST use only `pass`, `fail`, `missing`, or `skipped`; check result fields MUST use only `pass` or `fail`; nested objects MUST be schema-closed. Inactive or unauthorized rows MUST NOT appear in this row-evidence artifact.

`browser-e2e-visual` MUST select active visual catalog rows whose Playwright selector stage is `visual`. Direct execution runs the full target inventory; an owner slice constrains selection to the resolved owner rows. Exact Playwright title patterns MUST come from catalog selectors. Matching screenshots remain implementation-readiness evidence and MUST NOT be inferred from snapshot filenames, deleted ledgers, or visual fixture registry text alone.

The frontend visual fixture registry is exhaustive only for the semantic fixtures and design projections it explicitly declares. It is not an inventory of every active screenshot assertion or committed golden. An active screenshot and golden MAY have no fixture-registry mapping when its catalog row and emitted capture intent establish an exact consumer. Registry absence alone MUST NOT classify an active golden as drift or an orphan, and a harness MUST NOT manufacture a fixture ID from a filename, title, or path. Every declared registry fixture and every path named by that fixture MUST still reconcile exactly.

Each completed `browser-e2e-visual` attempt MUST retain exactly one non-claim-bearing `cartulary.frontend_visual_reconciliation.v1` artifact. The artifact MUST be a schema-closed JSON object containing `schema_id`, `status`, `source_refs`, `capture_intents`, `goldens`, `counts`, and `artifact_refs`. `status` is exactly `pass` or `fail`. `source_refs` MUST identify the authored catalog/family inputs, visual fixture registry, Playwright project/snapshot template, and screenshot-helper source used by the run; it MUST NOT use Markdown as executable input.

Each `capture_intents[]` item MUST contain a stable `capture_id`, exact catalog `row_id`, stable `scenario_id`, `project_id`, screenshot assertion location, semantic `capture_intent`, and the normalized repo-relative `expected_golden_path`. One assertion that intentionally emits multiple captures MUST emit one item per expected path. Each `goldens[]` item MUST contain `golden_path`, `sha256` or `null` when the expected file is absent, exact `consumer_capture_ids[]`, `catalog_row_ids[]`, `scenario_ids[]`, `project_ids[]`, `fixture_ids[]`, non-Playwright `consumer_refs[]`, and `classification`. `fixture_ids[]` MAY be empty for an active nonregistry golden. `classification` is exactly `active`, `orphan`, `missing_golden`, or `ambiguous_mapping`.

Reconciliation MUST derive expected paths from runtime capture intent plus the selected Playwright project's configured snapshot template, then join exact catalog selectors and exact registry paths. It MUST hash every committed PNG with SHA-256 and scan declared non-Playwright consumers. Filename interpretation MUST NOT establish ownership. `counts` MUST include `capture_intents`, `committed_goldens`, `active`, `orphan`, `missing_golden`, `ambiguous_mapping`, `registered_fixtures`, and `unresolved_registered_fixtures`.

Reconciliation MUST fail with `status=fail` when an active intent has no committed golden, an expected or committed path maps ambiguously, a selected assertion/catalog/scenario/project reference does not resolve exactly, or a declared registry fixture/path does not resolve. An `orphan` classification MUST identify a committed PNG with zero Playwright and zero declared non-Playwright consumers; it is a review input and MUST block blind refresh or movement but MAY proceed to an authorized deletion slice. The artifact remains implementation-readiness evidence and MUST NOT satisfy product conformance, design conformance, release, or Core 05 publication gates.

`browser-e2e-visual-update` MUST use the same visual row selection, runtime-profile session grouping, and service lifecycle as direct `browser-e2e-visual`, with Playwright snapshot update mode enabled for every selected group. The current refresh profile is the pinned Linux x86_64 environment and browser/toolchain pins owned by this NLSpec; other host profiles are unsupported for committed refreshes. The target MUST remain helper-only, MUST NOT be selected by `check`, `test`, `ci`, release gates, or either owner-slice command, and MUST NOT emit passing `browser-e2e-visual` target or owner-accounting evidence. Its authored writes are limited to committed Playwright visual goldens under `apps/web/e2e/workbook.visual.spec.ts-snapshots/`. A refresh record MUST name its accepted trigger, affected row and fixture IDs, changed golden paths, capture-contract changes or their explicit absence, reviewer outcome, and the later ordinary visual validation root. A refresh is complete only after changed images are reviewed and a later `browser-e2e-visual` run passes with screenshot comparisons active. Refresh artifacts remain implementation-readiness evidence and MUST NOT satisfy product conformance, design conformance, release, or Core 05 publication gates.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-022

Frontend visual fixture identity MUST be semantic, immutable, and local to the active owner-catalog row and exact Playwright scenario that uses it. When a retained support contract names a fixture, its ID MUST match `visual.fixture.<semantic_name>` and MUST NOT encode a frontend namespace, delivery phase, ordinal, or legacy row ID. Fixture-support metadata MUST declare `capture_scope` as `full_viewport`, `selector`, or `region`; selector captures MUST use a stable data-attribute selector. Selector-only non-grid specimens MAY declare `scroll_normalization.kind="not_applicable"` with a reason instead of a workbook-grid anchor. A fixture with `no_dynamic_regions=true` MUST keep `dynamic_masks=[]`. Missing, retired, placeholder, or ambiguous declared fixtures MUST NOT remain in the active fixture population and MUST NOT close an owner row from generic text, inferred snapshot names, or ad hoc scenario titles. Visual selection and accounting MUST come only from catalog rows; the fixture registry may enrich declared semantic fixtures but is not a second selection inventory.

For semantic-identity validation, fixture identity means values that route, select,
isolate, retain, or name test scenarios and artifacts, including fixture and seed
IDs, transaction and incident-key prefixes, actor/session ownership metadata,
scenario IDs, snapshot and golden paths, and retained artifact names. It does not
mean arbitrary product payload content merely because a test created the value.
Identity-bearing fixture values MUST be behavior-centric and MUST NOT embed a
retired delivery identity or a current owner, family, or row ID. Catalog identities
belong in selectors and retained evidence; they MUST NOT be copied into rendered
workbook content as replacement fixture labels. Product-owned uses of the word
`phase` and constructed negative fixtures that prove obsolete identity rejection
remain valid. Semantic validation MUST inspect declared identity-bearing
constructors and metadata rather than apply a repository-wide word ban.

The `visual.fixture.default_timeline_workbook_shell` fixture is the Default Timeline workbook shell. Its closure evidence MUST come from the exact active `module.workbook` visual catalog row and stable Playwright scenario declared for that fixture; the retained screenshot MUST be Playwright output from the running app-owned workbook shell with browser and operating-system chrome excluded. External concept images, generated mockups, browser-chrome screenshots, or design-reference bitmaps MAY inform review, but they MUST NOT be committed as the frontend golden, cited as row-accounting evidence, or used as an alternate visual fixture source.

Frontend visual support specimens MUST remain separate from default-shell evidence. Grid-adapter and token/theme screenshots classified as implementation or design support MUST NOT close `visual.fixture.default_timeline_workbook_shell`, substitute for the Default Timeline workbook-shell fixture, or become product-conformance evidence. One catalog row MUST NOT close another row unless its verification contract explicitly names an adopted consolidation relationship and the accounting artifact records that relationship.

Browser row accounting MUST close a catalog row only from the exact stable scenario IDs and diagnostic titles in its registered Playwright selector. A shell support row may use a target-level registered command result because its selector is the command ID. No other runner-wide or target-wide pass may close a row whose exact selected cases are absent. Failing or incomplete mapped targets MUST leave the row failed or infrastructure-failed rather than closed.

Public Make-owned verification targets MUST treat `unmapped` as unexpected executed test inventory. If a target exits successfully while its own accounting section has `counts.unmapped > 0`, target-summary generation MUST fail the target with `failure_class="harness"` and `failure_reason="test_accounting_unmapped"`. The retained `unmapped` and `unmapped_failed` count fields remain diagnostic fields, but successful public targets require `unmapped=0`. Canonical evidence MUST map through one active catalog row by the exact registered selector. Intentional residual support coverage MUST be declared in the verification registry and catalog with an owner and support profile; a separate unowned or filename-pattern classification is invalid.

Claim-publication routing is inactive unless a verification contract uses `behavior_class=claim_publication`, the profile and row posture resolve to the same active Core 05 claim, and the required publication predicate and artifact bundle validate. `claim_posture=informative` MAY retain engineering measurement but MUST NOT satisfy claim-bearing publication. The existing `benchmark-claim-check` target remains the Core 05 benchmark-manifest validator; ordinary implementation or informative evidence MUST NOT activate it. When its default manifest is absent, its no-claim pass MUST NOT be cited as Core 05 publication evidence.

**TH-HARNESS-REQ-256**
`explain-run` MUST diagnose retained aggregate run roots that contain `run-summary.json` and retained public tool-run roots that contain at least one `<target>/tool-run-summary.json`. Tool-run diagnostics MUST NOT require a synthetic aggregate `run-summary.json`. When a tool-run target also emits a command-specific summary artifact, such as `agent-finalize/finalize-summary.json`, `explain-run` MUST surface a bounded human summary of that artifact and retain `DETAIL=logs` access to target and child logs when `TARGET=<target>` is supplied. `DETAIL=accounting` MUST group retained test inventory by target, owner, family, evidence class, row ID, and file so module authors can triage unmapped or residual coverage without scraping raw summary JSON.

`explain-run DETAIL=logs` MUST accept only current
`cartulary.tool_run_summary.v5` structured log references. File references are
readable only for `format=log|text`. Directory references enumerate direct regular
`*.log` files in deterministic lexical order. Before emitting any referenced content,
the diagnostic MUST validate every selected reference, reject symlinks and traversal,
and enforce at most 4,096 files, 16 MiB per file, and 256 MiB in aggregate. An unsafe,
missing, or kind-mismatched reference fails with `artifact/artifact_error`, exit `11`,
without partial log replay. Non-current tool-summary schemas are unsupported rather
than translated or scanned for historical schema-specific diagnostics.
Verified by: TH-HARNESS-AC-015, TH-HARNESS-AC-019

**TH-HARNESS-REQ-257**
Cache records and retained cache-state artifacts are schema-owned harness
artifacts only for the registered graph profile that produced them. A
`cartulary.harness_cache_record.v2` identifies the profile, policy, producer,
unit and semantic digest, complete input and semantic-result digests,
vulnerability database revision when required, typed reusable artifact roster,
and creation time. Finalization and
topology rendering do not own separate action-cache schemas.

Cache records MAY be retained in `.cache/cartulary/*` and MAY be referenced by compact retained run-root cache artifacts such as `<target>/*-cache-*.json`. A run-root cache artifact proves only cache behavior for the current target attempt; it MUST NOT substitute for the target's required summary, child log, generated-drift verdict, security-scan output, service lifecycle evidence, or scheduler summary. Investigation tools MAY display cache state, but conformance evidence MUST cite the public target summary and required target artifacts rather than the local cache record.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-015, TH-HARNESS-AC-028

Operator runtime-binary consumers MUST retain a bounded provenance artifact for
graph-selected operator scenario work. That artifact names
`CARTULARY_OPERATOR_BIN` as graph-produced, identifies the `build-operator`
producer unit, records the normalized binary path and file digest, and links to
the producing unit result. A valid registered cache hit MAY satisfy the
producer, but a consumer is not a cache hit merely because its dependency was
reused.
Verified by: TH-HARNESS-AC-037

**TH-HARNESS-REQ-268**
The owner-first schema families are `cartulary.requirement_registry.v1`,
`cartulary.requirement_catalog.v1`, `cartulary.verification_registry.v3`,
`cartulary.verification_contract.v3`, `cartulary.test_owner_registry.v1`,
`cartulary.test_family_manifest.v6`, `cartulary.test_runner_registry.v1`,
`cartulary.harness_work_graph.v5`, `cartulary.harness_target_plan.v4`,
`cartulary.harness_row_result.v2`, `cartulary.harness_test_failure.v1`,
`cartulary.harness_unit_result.v1`,
`cartulary.harness_run_manifest.v1`, `cartulary.harness_unit_event.v2`,
`cartulary.harness_run_summary.v1`, `cartulary.harness_target_summary.v1`,
`cartulary.test_owner_explanation.v3`, `cartulary.task_guide_summary.v3`, and
`cartulary.test_catalog_check_summary.v2`. Each required current attachment
rejects old-family schema IDs.
Verified by: TH-HARNESS-AC-062, TH-HARNESS-AC-071

**TH-HARNESS-REQ-269**
An owner-slice target retains the canonical v3 run quartet, one target
projection, one unit result per selected unit, and one row result per selected
row. `test-evidence-audit` consumes only explicitly selected compatible v3 run
roots. `explain-test-owner` and `task-guide` are read-only and MUST NOT create
retained artifacts.
Verified by: TH-HARNESS-AC-064, TH-HARNESS-AC-071

**TH-HARNESS-REQ-270**
The canonical target plan contains selection kind, owner when applicable,
requested and resolved row IDs, graph digest, normalized work units,
dependencies, resource claims, fixture capabilities, service dependencies,
finalizers, evidence outputs, semantic digests, and target projections. Arrays are sorted and
duplicate-free. Owner selection is `complete_owner` or `exact_rows`; the
service-backed selector filters the complete or exact owner selection by a
nonempty explicit `service_dependencies` declaration.
Verified by: TH-HARNESS-AC-063, TH-HARNESS-AC-064

**TH-HARNESS-REQ-271**
Exactly one `cartulary.harness_row_result.v2` closes every resolved row. It
identifies the row, owning unit, terminal state, and normalized failure fields.
A missing, duplicate, unexpected, or incompatible row result is
`scheduler_accounting_error`.
Verified by: TH-HARNESS-AC-065

**TH-HARNESS-REQ-272**
The target projection aggregates only its resolved unit and row inventory. It
exposes status, exact unit IDs, inclusive and exclusive event-interval unions,
children, and canonical evidence references. A selected subset MUST NOT be
represented as complete-owner closure.
Verified by: TH-HARNESS-AC-065

**TH-HARNESS-REQ-276**
Every successful public target that selects catalog rows MUST finalize exactly
one canonical row result for every selected row before target-projection
success. Go, Vitest, Playwright, and shell adapters prove their exact registered
selector observations. Aggregate process success alone cannot close an absent,
duplicate, failed, or unmapped selected row. Target, aggregate-tier,
complete-owner, and exact-row selection are immutable graph inputs; a partial
or contradictory selection fails before retained row evidence is accepted.
Verified by: TH-HARNESS-AC-065, TH-HARNESS-AC-066, TH-HARNESS-AC-071

**TH-HARNESS-REQ-273**
The evidence audit identifies every required run root and target projection,
the exact compatibility tuple, required rows, accepted artifacts, rejected
artifacts with reasons, and final closure. It fails when any required
`(owner_id, target, row_id)` lacks exactly one compatible successful terminal
result or when a historical schema is supplied.
Verified by: TH-HARNESS-AC-066

**TH-HARNESS-REQ-274**
`cartulary.test_owner_explanation.v3` MUST expose owner, manifest, families,
row counts, runner and evidence distributions, tier and fixture-capability
distributions, and exact commands. `cartulary.task_guide_summary.v3` exposes
`role="module-author"`, owner, focused owner slice, applicable generated/drift
gates, broader gates, and release gate when required. Neither schema may
contain a delivery-phase selector.
Verified by: TH-HARNESS-AC-064, TH-HARNESS-AC-071

**TH-HARNESS-REQ-275**
`cartulary.executable_input_policy.v1` is a required current schema attachment.
It uses Draft 2020-12, a constant `schema_id`, closed keys, required fields, and
`additionalProperties=false`; its policy validator rejects unknown fields,
unsorted inputs, or unsafe restricted-root declarations before scanning
executable behavior.
Verified by: TH-HARNESS-AC-067, TH-HARNESS-AC-071

**TH-HARNESS-REQ-277**
The required current graph and evidence schema attachments are
`cartulary.harness_work_graph.v5`,
`cartulary.harness_capability_snapshot.v1`,
`cartulary.harness_capacity_override.v1`,
`cartulary.harness_cache_record.v2`,
`cartulary.harness_fixture_lease.v4`,
`cartulary.harness_run_manifest.v1`,
`cartulary.harness_unit_event.v2`,
`cartulary.harness_run_summary.v1`,
`cartulary.harness_target_summary.v1`,
`cartulary.harness_public_target_duration_baselines.v3`, and
`cartulary.harness_performance_evidence_roots.v3`. Historical scheduler,
observability, timing, cache, performance, and evidence-root artifacts are
archival bytes only. A current command MUST reject them as input and MUST NOT
provide a compatibility reader. Each current schema uses Draft 2020-12,
requires its exact schema ID, closes unknown fields, and validates before the
corresponding artifact or check succeeds.
Verified by: TH-HARNESS-AC-073, TH-HARNESS-AC-074, TH-HARNESS-AC-079

**TH-HARNESS-REQ-278**
Every graph-backed public invocation retains exactly one immutable
`run-manifest.json`. The manifest binds the public target and command ID,
declared inputs, source, toolchain, system and graph digests, the immutable
capability snapshot, cache mode, and start time. Its graph digest is computed
from the fully resolved work graph before child execution. A missing,
mismatched, mutable, or superseded manifest makes the run artifact-incomplete.
Verified by: TH-HARNESS-AC-073, TH-HARNESS-AC-074

**TH-HARNESS-REQ-279**
`run-summary.json` and `target-summaries/<target>.json` are deterministic
projections of the validated run manifest, unit-event stream, unit results, and
row results. Stable source bytes produce byte-identical normalized projections.
Run and target interval unions, critical path, resource holders, cache
dispositions, selected rows, failures, and artifact references close exactly
once. A projector disagreement fails validation; no wrapper or child summary
may override canonical events.
Verified by: TH-HARNESS-AC-074, TH-HARNESS-AC-075

**TH-HARNESS-REQ-280**
Read-only diagnostic and OTLP export commands consume only a complete canonical
run. Their projections use `service.name=cartulary.harness` and instrumentation
scope `cartulary.harness.execution`, preserve canonical unit identities,
intervals, statuses, and resource attributes, and do not write a retained trace,
metric, index, execution-context, hotspot, or invocation-start artifact. Native
canonical evidence remains authoritative when an export consumer disagrees.
Verified by: TH-HARNESS-AC-075, TH-HARNESS-AC-076

**TH-HARNESS-REQ-281**
`unit-events.ndjson` is the sole current execution event stream. Each record
validates as `cartulary.harness_unit_event.v2`; run-local `seq` starts at one and
has no gaps. Events cover run start, eligibility and waits, unit start and
terminal state, resource and fixture ownership, cache disposition, cancellation,
and run completion. Aggregate target summaries are projections of these events;
they MUST NOT emit a second sequence-local timing authority.

Every production reader MUST consume this stream through the evidence-accounting
owner's validated asynchronous NDJSON iterator. The iterator MUST reject a
missing, empty, non-regular, symlinked, malformed, oversized-line,
non-contiguous, time-reversing, or cancelled input and MUST release its file
handle when a consumer terminates early. Unit interval, terminal-roster,
measurement-quietness, observability, retained-run, drift, and diagnostic
projections MUST retain only the bounded unit state needed by that projection;
they MUST NOT materialize the complete event file, provide a synchronous
whole-file compatibility reader, or rely on a JavaScript string-size or child
`maxBuffer` estimate. A line-size bound protects the closed event schema and is
not permission to omit any valid canonical event from the projection.

The scheduler MUST serialize this single stream through one bounded writer into
a private mode-`0600` staging file and atomically publish
`unit-events.ndjson` only after the terminal run event is durable. A downstream
same-run finalizer that requires already-emitted scheduler proof MUST read that
same writer-owned staging stream through an internal, run-contained path; it
MUST NOT require a prematurely published canonical path, write an event copy,
or treat the incomplete staging stream as retained evidence. External readers
and post-run diagnostics consume only the atomically published canonical path.
Verified by: TH-HARNESS-AC-074, TH-HARNESS-AC-077

**TH-HARNESS-REQ-282**
Harness observability is a read-only projection over canonical run evidence.
Missing, partial, corrupt, secret-bearing, or historical input fails the
explicit observability check without mutating the run root or replacing a
product, harness, cleanup, or artifact result. Ordinary child wrappers may
observe that canonical evidence is not yet finalized, but they MUST NOT create
a second timing authority or a partial placeholder artifact.
Verified by: TH-HARNESS-AC-073, TH-HARNESS-AC-078

**TH-HARNESS-REQ-283**
The top-level invocation boundary begins with the validated immutable run
manifest and the first canonical run event and ends with the terminal canonical
run event and atomically published run summary. Unit parentage comes only from
work-graph dependency edges; it is never inferred from temporal containment or
legacy Make prerequisites. Source identity, externally available capacity,
target-scoped execution policy, cache posture, interruption, terminal status,
and target projections are retained in the canonical artifact set. A required
performance root without that complete same-run closure is artifact-incomplete.
Verified by: TH-HARNESS-AC-073, TH-HARNESS-AC-074, TH-HARNESS-AC-079

### 8.1 Artifact Families

| Artifact family                                      | Producer                                        | Path under run root                                             | Schema policy                                                 | Ordering and nullability                                                              | Retention and cleanup                                        |
| ---------------------------------------------------- | ----------------------------------------------- | --------------------------------------------------------------- | ------------------------------------------------------------- | ------------------------------------------------------------------------------------- | ------------------------------------------------------------ |
| Tool-run summary                                     | Centralized wrappers                            | `<target>/tool-run-summary.json` or the summary directory closed by the target's Section 8.1 row | `cartulary.tool_run_summary.v5`                               | Required non-null timestamps, target, exit code, output mode, artifact refs, failures | Retained; removed by cleanup only under default result root. |
| Fallow static reports                                | `frontend-fallow-static`                        | `frontend-fallow-static/fallow/*` and `frontend-fallow-static/fallow-static-summary.json` | `cartulary.fallow_static_summary.v2` for the normalized summary; raw JSON, SARIF, Markdown, stdout, stderr, and `resolved-fallowrc.json` files are diagnostic-only | Report names, statuses, issue counts, artifact refs, resolved-config refs, baseline state, and enforcement state in schema-defined order | Retained as run-root artifacts; generated/source roots and Fallow config or baseline inputs are not cleanup candidates. |
| Vitest failure details                               | Vitest wrappers                                 | `<target>/raw/<row-id>/vitest-failure-details.json`             | `cartulary.vitest_failure_details.v1`                         | Runner JSON reference, stdout/stderr refs, failed owner path, title, message, source, raw messages, diagnostic tags, and first app frame | Retained when a Vitest runner report exists; absent sidecar uses owner-summary fallback. |
| Run manifest                                         | Graph runner                                    | `run-manifest.json`                                             | `cartulary.harness_run_manifest.v1`                            | Immutable command, source, system, graph, capacity, cache-mode, and target-projection identity | Retained. |
| Unit event stream                                    | Graph scheduler                                 | `unit-events.ndjson`                                            | `cartulary.harness_unit_event.v2`                              | `seq` strictly increases with no gaps; stable unit and resource ordering | Retained; sole current timing authority. |
| Run summary                                          | Canonical evidence projector                    | `run-summary.json`                                              | `cartulary.harness_run_summary.v1`                             | Unit outcomes, interval unions, critical path, resources, cache, and row closure in stable order | Retained. |
| Target summary                                       | Canonical evidence projector                    | `target-summaries/<target>.json`                                | `cartulary.harness_target_summary.v1`                          | One projection per selected public target; unit IDs and rows ordered deterministically | Retained. |
| Unit result                                          | Unit runner                                     | `unit-results/<unit-id>.json`                                   | Schema declared by the unit evidence contract                 | Stable unit identity, terminal state, artifacts, and row references | Retained. |
| Row result                                           | Evidence adapter                                | `row-results/<row-id>.json`                                     | Schema declared by the row runner                              | Exactly one terminal result for each selected row | Retained. |
| Harness public-target duration baselines             | Performance maintenance target                  | `tools/harness_public_target_duration_baselines.v3.json`           | `cartulary.harness_public_target_duration_baselines.v3`        | Exact public-owner roster and bindings, target-local provenance, timing source, policy projection, forced-cold observation, discarded warm-up, five or six measured samples, p50, p90, MAD, structural metrics, gates, and recomputed public portfolio total in target order; internal diagnostics are separate and excluded from public closure | Checked-in maintenance artifact; regenerated only by its Make target. |
| Govulncheck findings                                 | Govulncheck wrapper                             | `<target>/<row-id>/govulncheck-findings.json`                    | `cartulary.govulncheck_findings.v1`                            | Finding IDs and counts in deterministic order; symbol-reachable findings are blocking | Retained with the row; promoted by target summary as artifact refs and supplemental security extension data. |
| Agent finalizer summary                              | Agent finalizer graph unit                      | `unit-artifacts/finalize-summary.json`                          | `cartulary.agent_finalize_summary.v3`                         | Ordered actions, private substeps, skipped work, retained-run checks, rollback state, updated files, `RESULTS_DIR`, and child artifact refs | Retained and referenced by the canonical unit and target summaries. |
| Cache record                                         | Graph cache                                     | `.cache/cartulary/graph-v2/<profile>/<key>/record.json`          | `cartulary.harness_cache_record.v2`                            | Profile, producer, semantic result, key inputs, tool and freshness digests, typed reusable artifacts, and validation state | Records are reusable only under registered policy; the v1 graph root is ignored and `make distclean` removes both. |
| Fixture lease                                        | Run-scoped broker                               | `_shared/fixture-leases/<lease-id>.json`                         | `cartulary.harness_fixture_lease.v4`                           | Capability, ownership, resources, acquisition, quarantine, cleanup outcome, and normalized cleanup failure | Retained as lifecycle and cleanup evidence. |
| Runtime binary provenance                            | Go target runner                                | `_shared/<execution-family>/runtime-binaries.json` when an aggregate declares runtime binaries | diagnostic-only                                               | Runtime binary ID, scheduler-owned consumer env, producer target, normalized path, file digest, build-artifact ref, and output digest | Retained with the shared Go report. |
| Service scope summary                                | Service suite                                   | `_shared/test-services/<suite-id>/service-scope.json`            | `cartulary.test_services.scope.v2`                         | Bounded suite identity, readiness, aggregate service, fixture, failure, and cleanup summaries closed by Section 11 | Retained; atomically replaced during lifecycle.              |
| Service producer journals                            | Service suite                                   | `_shared/test-services/<suite-id>/journals/<producer-id>.ndjson` | `cartulary.test_services.journal_event.v1`                 | Stable producer identity and gap-free local sequence; deterministic collation by event time, producer, and sequence | Retained only through suite finalization.                    |
| Private service resource ledger                      | Service suite                                   | Private suite runtime state                                     | `cartulary.test_services.resource_ledger.v1`               | Exact mode-0600 owned-resource identities and cleanup state | Deleted after successful cleanup; redacted contained recovery copy only on cleanup failure. |
| Service lifecycle event stream                       | Service suite                                   | `_shared/test-services/<suite-id>/lifecycle-events.jsonl`        | `cartulary.test_services.lifecycle.v2`                        | `seq` strictly increases; transitions match Section 11.2                               | Retained; not cleanup proof.                                |
| Browser service admission                            | Browser session lifecycle                       | `_shared/test-services/<suite-id>/browser-sessions/<browser-session-id>/service-admission.json` | `cartulary.test_services.browser_admission.v1` | Suite/session identity, readiness generation, required services, container proof, source digest, and service-scope digest | Retained for the session; contains no exhaustive service inventory. |
| Browser startup events                               | Browser session lifecycle                       | `_shared/test-services/<suite-id>/browser-sessions/<browser-session-id>/startup-events.jsonl` | `cartulary.browser_startup_event.v1` | Exact suite/session/profile identity and validated append-only state transitions | Retained for the session; lifecycle adapter is sole writer. |
| Browser startup diagnostics                          | Browser session lifecycle                       | `_shared/test-services/<suite-id>/browser-sessions/<browser-session-id>/startup-diagnostics.json` | `cartulary.browser_startup_diagnostics.v2` | Immutable terminal state, event reference/digest, classification, redaction-safe message, origins, and artifact references | Retained for the session; group and target evidence consume by reference. |
| Browser stack metadata                               | Browser session lifecycle                       | `_shared/test-services/<suite-id>/browser-sessions/<browser-session-id>/stack-v6.json` | `cartulary.web_e2e_stack.v6` | Immutable suite/session/mode/profile identity, compact service admission, database, object-store namespace, backend/frontend process proofs, build digest, fixture, diagnostic, lease, and readiness bindings | Retained for current-run attach admission. |
| Browser backend generation                           | Browser reset lifecycle                         | `_shared/test-services/<suite-id>/browser-sessions/<browser-session-id>/backend-generations/<reset-id>.json` | `cartulary.web_e2e_backend_generation.v1` | Reset ID, monotonic generation, unchanged runtime/config identity, base stack reference/digest, and replacement backend process proof | Immutable current-run attachment overlay. |
| Browser group result                                 | Browser evidence adapter                        | `<target>/browser-groups/<group-id>/browser-group-result.json` | `cartulary.browser_group_result.v5` | Exact selected rows, terminal observations, lease reference, and ordered session artifact references/digests | Retained for target accounting. |
| Browser target result                                | Browser evidence finalizer                      | `<target>/browser-target-result.json` | `cartulary.browser_target_result.v2` | Ordered group-result references/digests and deduplicated session artifact references/digests | Retained for target accounting. |
| Local object-store proxy attempt, lease, and health  | Local development proxy lifecycle               | owner-only `.cartulary/runtime/object-store-proxy/` state and loopback health endpoint | `cartulary.local_object_store_proxy_start_attempt.v1`, `cartulary.local_object_store_proxy_lease.v1`, `cartulary.local_object_store_proxy_health.v1` | Canonical nonsecret configuration, instance identity, boot-aware process proof, and readiness state | Development-only; never browser or product evidence. |
| Database reset diagnostic                            | Recovery reset controller                       | `<target>/reset-boundary/<label>.database-reset.json`            | `cartulary.test.database_reset_diagnostic.v1`                 | Reset ID, attempt one, closed stage, nullable SQLSTATE, timeout flag, duration, sorted table/count proofs, and normalized failure | Retained; excludes raw SQL, DSNs, database names, credentials, backend IDs, and raw errors. |
| Browser reset attempt                                | Browser reset lifecycle                         | `<target>/reset-boundary/<label>.attempt.json`                   | `cartulary.browser_reset_attempt.v1`                          | Ordered lifecycle outcome, old/new backend generations, database diagnostic reference, persistent/browser reset proof, taint, and terminal classification | Authoritative lifecycle-unit failure evidence. |
| Test clock-control response                          | Test clock route                               | clock-control transcript or target-owned clock-control dir       | `cartulary.test.clock_control.v1`                             | Clock mode, current RFC3339 timestamp, offset seconds, and fixed timestamp when mode is fixed | Retained only by the target or fixture transcript that controls the clock; never production API evidence. |
| Network Flow fault-control response                  | Network Flow fault-control route                | Network Flow fixture transcript or target-owned fault-control dir | `cartulary.test.network_flow_fault_control.v1`                | Fault ID, exact boundary token, fault kind, optional safe error code, optional correlation key, and `consume_once=true` | Retained only by the target or fixture transcript that arms the fault; never production API evidence. |
| Network Flow randomness-control response             | Network Flow randomness-control route           | Network Flow fixture transcript or target-owned randomness-control dir | `cartulary.test.network_flow_randomness_control.v1`           | Control ID, exact stream token, value kind, value count, remaining count, `consume_once=true`, and `exhaustion="fail_closed"` | Retained only by the target or fixture transcript that arms deterministic fixture randomness; never production API evidence. |
| Network Flow auth-transition-control response        | Network Flow auth-transition-control route      | Network Flow fixture transcript or target-owned auth-transition-control dir | `cartulary.test.network_flow_auth_transition_control.v1`      | Control ID, exact boundary token, transition kind, actor ref, incident ref, resource kind/ref, hidden response kind, optional correlation key, `must_not_disclose_resource=true`, and `consume_once=true` | Retained only by the target or fixture transcript that arms route-time authorization or hidden-resource assertions; never production API evidence. |
| Network Flow audit-assertion-control response        | Network Flow audit-assertion-control route      | Network Flow fixture transcript or target-owned audit-assertion-control dir | `cartulary.test.network_flow_audit_assertion_control.v1`     | Assertion ID, assertion kind, event code, operation ref, actor ref, incident ref, resource kind/ref, baseline count, expected final count, expected replay increment, optional correlation key, and `consume_once=true` | Retained only by the target or fixture transcript that arms exact-count or replay-silence assertions; never product audit evidence by itself. |
| Frontend accessibility summary                       | Browser accessibility target                    | `browser-e2e-a11y/accessibility/frontend-accessibility-summary.json` | `cartulary.frontend_accessibility_summary.v4`                  | Active `rows[]`, `scenarios[]`, `keyboard_matrix[]`, `state_communication_checks[]`, `contrast_checks[]`, `violations[]`, and `artifact_refs[]` in schema-defined order | Retained for browser target.                                 |
| Release-readiness projection                         | Unified scheduler                               | `target-summaries/release-readiness-evidence.json`                  | `cartulary.harness_target_summary.v1`                          | Exact producer-unit references, inclusive interval union, status, cache accounting, and evidence refs from the canonical run | Retained as part of the release-check canonical run; no parallel release-evidence format is written. |
| Network Flow fixture manifest                        | Network Flow fixture manifest validator         | `fixtures/network-flow/<fixture_id>/manifest.json`                  | `cartulary.network_flow_fixture_manifest.v2`                   | Fixture identity, source files, expected artifacts, transcript files, per-file SHA-256 values, and aggregate bundle hashes in canonical sorted order | Source fixture roots are committed and immutable after freeze; run-local materializations are retained under the selected target's run root. |
| Generated manifest summaries                         | Generation/drift scripts                        | tool-specific target dirs                                       | JSON schemas declared by generated artifacts                  | Unknown fields rejected where shape tools enforce closure                             | Generated files remain checked in; summaries retained.       |
| Logs                                                 | Shell, Go, scheduler, browser, service wrappers | target log dirs                                                 | diagnostic-only unless producer declares schema               | Logs are text after redaction; empty logs may be omitted                              | Retained unless cleanup removes result root.                 |
| Coverage reports                                     | Go/frontend/test tools                          | tool-specific coverage paths                                    | diagnostic-only                                               | No current schema-owned field contract; retained only as tool diagnostic output       | Removed by `make clean` when under registered paths.         |
| Playwright screenshots, videos, traces, HTML reports | Playwright                                      | Playwright report/test-results dirs                             | diagnostic-only secret-bearing                                | No current schema-owned field contract; retained only as Playwright diagnostic output | Removed by `make clean` when under registered paths.         |
| Visual snapshots and goldens                         | Browser/fixture tools                           | source and tool-specific dirs                                   | validation-only; helper refresh is owned by `browser-e2e-visual-update` | No current schema-owned diagnostic schema contract; helper refresh writes tool-specific committed PNGs only | Refresh is helper-only and is not validation evidence until `browser-e2e-visual` passes. |

**TH-HARNESS-REQ-258**
Every target named by a sequence step's `produces_summary_targets[]` MUST retain `<target>/target-summary.json` in the selected run root before the sequence aggregate emits its run summary or aggregate target summary. A target's `<target>/tool-run-summary.json` remains the wrapper-owned tool-run summary. Command-specific reports retained by the target, such as SeaweedFS compatibility or release-gate reports, MUST NOT substitute for `target-summary.json` when the target is a sequence-produced summary target.

Artifact references follow the ownership direction. A target tool-run summary MUST reference artifacts owned by that target and MAY reference its nested scheduler or owner-partition artifacts. Only the aggregate target whose identity equals the enclosing root tool summary's target MAY reference the root `run-summary.json` and root `tool-run-summary.json`; every leaf or nested target MUST omit those parent artifacts. A child target's artifact inventory therefore MUST be identical whether it finalizes before or after the enclosing aggregate artifacts exist.
Verified by: TH-HARNESS-AC-015, TH-HARNESS-AC-023, TH-HARNESS-AC-027

Graph-backed targets expose scheduler state through the canonical event stream,
run summary, and target projections. There is no nested scheduler artifact
family. Resource pressure is projected from unit claims and canonical events
and includes requested and resolved capacity, peak use, saturation duration,
wait count, blocked units, blocking resources, and actual holders. Readiness,
fixture, cache, browser, and build work is diagnosable by first-class unit ID
and evidence references rather than inferred target-name maps.

**TH-HARNESS-REQ-259**
Same-run reuse is valid only for a registered semantic unit produced earlier by
the same graph invocation. The scheduler retains the producer once, projects
its unit ID into every consuming target, and emits one cache hit or ordinary
dependency event as applicable. It MUST NOT synthesize a second helper-artifact
family, copy timing into consumers, or read another run root.

The producer artifacts referenced by a same-run helper artifact ref MUST resolve under the current run root and MUST exist before the consumer aggregate succeeds. Missing artifacts, paths outside the current run root, malformed refs, digest mismatches, or retained-run references from any prior run MUST fail closed as artifact/configuration errors rather than silently falling back to old evidence. A same-run helper artifact ref MAY reduce duplicate helper/setup work inside one aggregate, but it MUST NOT skip any selected production-conformance row, product test, browser/live-state row, service-backed row, drift verdict, security verdict, cleanup, runtime reset, scratch migration apply, object-store mutation, or destructive-operation safeguard.

Same-run helper artifact refs are not cache records and are not scheduler work reuse. Aggregates MAY expose them in run summaries and `explain-run`, but scheduler summaries and pressure summaries MUST continue to report current-profile scheduler `reused` counts as `0` unless a later NLSpec revision adopts scheduler work reuse separately.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-025, TH-HARNESS-AC-028, TH-HARNESS-AC-048

### 8.2 Network Flow Fixture Manifest Contract

**TH-HARNESS-REQ-262**
Network Flow conformance fixtures MUST use a directory-scoped manifest at `fixtures/network-flow/<fixture_id>/manifest.json`. The `<fixture_id>` directory name and manifest `fixture_id` MUST be identical, MUST use the full `NF-FIX-###-slug` identifier, and MUST NOT be inferred from source filenames, display names, route labels, generated output names, or test titles. Legacy single-file fixture locators are not canonical manifest identity.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-263**
A Network Flow fixture manifest MUST validate as
`cartulary.network_flow_fixture_manifest.v2`, MUST be schema-closed at every
object boundary, and MUST declare `profile_id="network_flow_activity"`,
`manifest_version=2`, `freeze.status`, `freeze.revision`, `source_files[]`,
`expected_artifacts[]`, `transcript_files[]`, `source_bundle_sha256`, and
`expected_bundle_sha256`. It MUST omit acceptance IDs, verification IDs,
copied requirements, execution selectors, phase selectors, and specification
provenance. Each listed scenario file MUST validate as
`cartulary.network_flow_fixture_scenario.v2` and contain only fixture identity
and a human-readable behavior summary. File arrays MUST be ordered by
`logical_path` ascending by Unicode code point, and each listed file MUST
declare exact byte `size_bytes` and lowercase hex SHA-256 of the committed file
bytes.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-264**
The Network Flow source bundle digest algorithm is `network_flow_fixture_bundle_hash_v1`. For each `source_files[]` entry in manifest order, the validator hashes the UTF-8 frame `logical_path`, a NUL byte, lowercase `sha256`, a NUL byte, decimal `size_bytes`, and LF. `source_bundle_sha256` MUST equal the SHA-256 of the concatenated frames. `expected_bundle_sha256` uses the same frame algorithm over `expected_artifacts[]` followed by `transcript_files[]`, preserving each array's manifest order. A missing file, extra unlisted file, digest mismatch, size mismatch, unsorted list, duplicate path, absolute path, symlink, or traversal path MUST fail before product code starts.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-265**
Only `freeze.status="frozen"` Network Flow manifests may close Network Flow
behavior evidence. A frozen manifest is append-only by revision: any byte
change to source files, expected artifacts, transcript files, or aggregate
digests requires a new `freeze.revision` and a tracker entry that names the
changed fixture. Draft manifests MAY exist during fixture authoring, but public
verification targets MUST report them as blocked rather than treating draft
bytes as current evidence.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-266**
Network Flow fixture execution MUST materialize manifest-listed files into a run-local read-only input workspace under the selected result root. Product tests MUST read the run-local materialization, not mutate the committed fixture directory. Expected artifacts and transcript files are read-only comparison inputs. The runner MUST retain a bounded execution summary that names selected fixture IDs, manifest file SHA-256, source and expected bundle SHA-256 values, materialized input root, produced artifact refs, and comparison status. This summary is harness evidence only and does not define Network Flow product behavior.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-267**
Network Flow fixture manifests carry integrity and execution data; they do not
route evidence or own domain semantics. Test-catalog rows route executable
fixture tests through the Network Flow behavior verification. A manifest MUST
NOT redefine import parsing, row identity, cursor behavior, graph behavior,
indicator binding, authorization, retention, audit occurrence rules, or
generated contract shape.
Verified by: TH-HARNESS-AC-049

### 8.3 Agent Finalizer

**TH-HARNESS-REQ-260**
`agent-finalize` is a harness-maintenance finalizer. It refreshes and validates deterministic harness-maintenance artifacts before a caller runs explicit verification. It MUST NOT be described or implemented as a verification gate, test runner, cleanup target, code-generation workflow, migration workflow, release gate, security gate, build gate, browser E2E surface, or benchmark-claim surface.
Verified by: TH-HARNESS-AC-019

**TH-HARNESS-REQ-261**
`agent-finalize` exposes exactly one semantic operation: finalize harness-maintenance artifacts before explicit verification. Its public input surface is `make agent-finalize` with optional `RESULTS_DIR=<run_root>` and the ordinary output-mode controls. Callers MUST NOT select finalizer substeps by child target name.

`agent-finalize` MUST derive its execution plan from the closed finalizer action registry below.

| Action ID | Requires `RESULTS_DIR` | Mutating | Required behavior | Allowed output |
| --- | ---: | ---: | --- | --- |
| `schema_shape_validation` | no | no | Validate current schemas, attachments, and catalog closure. | Finalizer and child summaries. |
| `tier_coverage_validation` | no | no | Prove every active row has one valid tier and executable owner projection. | Finalizer and child summaries. |
| `generated_structure_refresh` | no | yes | Render generated state once into scratch, validate and compare once, and atomically publish only changed outputs. | Finalizer summary, child summary, and updated-file list. |
| `canonical_evidence_validation` | yes | no | Validate manifest, event, run-summary, row/unit-result, and target-projection closure. | Finalizer and child summaries. |
| `scheduler_drift_validation` | yes | no | Validate canonical event order and warm graph health without a fixed timing cap. | Finalizer and child summaries. |

The implementation MAY realize an action by invoking one or more Make targets or scripts. Child target names are not part of the `agent-finalize` public contract unless this NLSpec explicitly promotes them.
Verified by: TH-HARNESS-AC-019, TH-HARNESS-AC-020

Finalizer actions are not a separate cache family. Same-run deduplication is
available only when `agent-finalize` is itself selected in a canonical graph;
standalone finalization executes each selected action exactly once. Historical
action-cache records and disable/force variables are unsupported current input.

When `RESULTS_DIR` is set, `agent-finalize` MUST validate the supplied retained run before any mutating generated-artifact or duration-baseline refresh. A valid finalizer `RESULTS_DIR` is an existing retained full warm `make check` run root that identifies a successful, uncontaminated run and contains schema-valid `run-manifest.json`, `unit-events.ndjson`, `run-summary.json`, `target-summaries/check.json`, and every selected row and unit result. The retained root MUST be the latest successful full warm `make check` root under the same retained-results parent unless `ALLOW_OLDER_RESULTS_DIR=1` is supplied and recorded. Every consumed artifact MUST match the current source, system, semantic, catalog, verification, graph, capacity, and cache-mode identity. Partial, failed, contaminated, mixed-schema, older-without-override, or evidence-incomplete roots MUST be rejected before mutation as `configuration_error` for invalid caller input or `artifact_error` for unsafe retained evidence.

After retained-run preflight succeeds, the finalizer MUST freeze the accepted pre-mutation source-snapshot identity for every retained-evidence child action. A later child that reads owner accounting after an earlier baseline writer intentionally changes tracked maintenance bytes MUST compare that accounting with the frozen identity, not a newly computed mid-transaction snapshot. The frozen identity is private finalizer context: it MUST be scoped to the exact normalized retained root, MUST fail closed when malformed or paired with a different root, and MUST NOT alter standalone public baseline commands, which continue to compare retained evidence with the current source snapshot.

When `agent-finalize/finalize-summary.json` records a normalized child failure, the enclosing shell step and public target summaries MUST promote that failure class, reason, headline, and child target. The generic nonzero shell wrapper failure MUST NOT outrank or duplicate the normalized finalizer failure as `unknown_failure`; raw wrapper stdout and stderr remain retained log artifacts.

Without `RESULTS_DIR`, finalization MUST execute `schema_shape_validation`,
`tier_coverage_validation`, and `generated_structure_refresh` in that order;
retained-only actions remain present and explicitly not selected. With
`RESULTS_DIR`, retained-run preflight and `scheduler_drift_validation` run
first, followed by schema and tier validation, the generated refresh, and
`canonical_evidence_validation`. `generated_structure_refresh` MUST invoke one
Make-owned refresh transaction that renders once, validates and compares once,
and publishes atomically. A retained finalizer MUST NOT invoke the same semantic
unit twice in one prepared run identity.
Verified by: TH-HARNESS-AC-019

Duration-baseline refreshes remain advisory harness planning data and MUST NOT become benchmark claims or product performance conformance evidence. Browser entry baselines MUST use semantic catalog row IDs, and scheduler work-unit baselines MUST use current generated work-unit IDs. Phase-keyed, `E-*`, and `FE-*` duration identities are unsupported. A baseline refresh MUST consume only compatible successful owner and scheduler evidence; obsolete entries MUST be removed or ignored rather than carried forward by compatibility translation.
Verified by: TH-HARNESS-AC-019

Before any selected finalizer action can mutate tracked generated or baseline files, `agent-finalize` MUST snapshot or stage the tracked worktree state it is allowed to change. If any later selected action fails, the finalizer MUST restore tracked files to the start-of-run state before writing `finalize-summary.json`; pre-existing tracked user changes MUST be preserved, and untracked files MUST NOT be deleted as part of rollback. The finalizer summary MUST report rollback status with `mutation_rollback.status`, `restored_file_count`, `restored_files[]`, and `error`.
Verified by: TH-HARNESS-AC-019

When `agent-finalize` mutates tracked generated or baseline artifacts and completes successfully, those promoted mutations MUST be explicit in `finalize-summary.json` through `generated.updated_file_count` and `updated_files[]`. Audit or handoff records that cite an `agent-finalize` run MUST distinguish pre-existing worktree changes from finalizer-caused updates. A finalizer update MUST NOT be treated as silent remediation merely because the command succeeded.
Verified by: TH-HARNESS-AC-019

The `generated_structure_refresh` action MUST use the ordinary generated-drift
ownership surface in refresh mode, which shares the single transactional
renderer with `generate`. `agent-finalize` MUST NOT invoke `format`,
`migration-drift`, `test-fast`, `test`, `check`, `ci`, `release-check`, browser
E2E targets, security scan targets, build targets, `clean`, `distclean`, or
`benchmark-claim-check`.
Verified by: TH-HARNESS-AC-019

`agent-finalize` MUST be fail-fast and resumable. It MUST stop at the first failed substep, preserve completed substeps, mark later selected substeps as skipped-after-failure, retain child logs and child summaries when available, and propagate the normalized child failure class and reason when a failed child summary is readable. Summary-write or cleanup-reporting failures MUST be reported without masking an earlier primary child failure.
Verified by: TH-HARNESS-AC-019

`agent-finalize` MUST retain `agent-finalize/finalize-summary.json` with schema
ID `cartulary.agent_finalize_summary.v3`. The summary MUST include selected and
not-selected actions, private substeps, skipped work, updated files,
`RESULTS_DIR`, retained-run checks, rollback state, failure records, and child
artifact references. Human summary output MUST include one compact `[FINALIZE]`
line before the ordinary `[RESULT]` line. Machine output remains exactly one
`cartulary.tool_run_summary.v5` object; callers use the `finalize_summary`
artifact reference for command-specific details.
Verified by: TH-HARNESS-AC-004, TH-HARNESS-AC-019

Check scheduler summaries that include service sessions MUST report service-suite setup timing separately from child test timing. Each `service_sessions[]` entry MUST include `setup_duration_ms`, `ready_at_monotonic_ms`, `child_work_started_at_monotonic_ms`, and `cleanup_duration_ms`; fields MAY be `null` only when the corresponding lifecycle segment did not run or did not reach readiness. Duration baselines for backend, browser, and service-backed child work MUST be derived from child work-unit timings, not from service-suite setup or cleanup timing.

Scheduler-owned service-session startup MUST retain a minimal redacted service-session environment diagnostic before Docker preflight or managed-service startup begins. If startup fails before a lease can be written, the service-session summary MAY report no lease path only when service cleanup is explicitly recorded as `skipped_no_lease` and the failure artifact points to the current-run `service-scope.json`. Summary artifact references MUST NOT point investigators only at paths that cannot exist for the observed startup stage.

## 9. Failure Classes and Exit Codes

**TH-HARNESS-REQ-300**
Public Make-owned wrappers MUST expose exact normalized public exit codes according to the failure-reason table below in retained summaries and compact failure output. Raw child process exit codes MAY be preserved in summaries but MUST NOT define the normalized public exit code except where `child_target_failure` explicitly delegates to a normalized child failure class.
Verified by: TH-HARNESS-AC-014

Public exit-code selection is reason-based. Wrappers MUST derive the
normalized public exit code from the normalized `failure_reason` and primary-failure
rules in this section, not from the raw process status of a child command. The current GNU Make invocation binding may return GNU Make's executor failure status for a failed recipe; callers that require reason-specific failure codes MUST read the retained `tool-run-summary.json` or compact failure line rather than treating the outer `make` process status as the normalized public exit code.

Scheduler summaries MUST propagate normalized failures from every completed failed work unit whose child target summary is available. For a failed `service_session` work unit, a current-run `cartulary.test_services.scope.v2` artifact is also an authoritative child diagnostic source when startup fails before a child target summary exists. Scheduler failure collection MUST read the schema-valid service scope from the same retained run root, emit the service failure as an ordinary `failures[]` record, and point its artifact reference at `service-scope.json`. The scheduler's own fallback classification is used only when no child summary or current-run service diagnostic exists, the diagnostic is unreadable, or the failure belongs to scheduler orchestration rather than completed child target work. The summary's primary failure still follows Section 9.1 ordering, but `failures[]`, `failure_classes`, and `failure_reasons` MUST retain all completed failed work units. Dependency-skipped completion rows MUST NOT be counted as independent failures when the underlying failed work unit is already represented. A child target assertion failure therefore remains `failure_class=product` and `failure_reason=test_assertion_failure` at the scheduler summary layer, while a concurrent config, harness, artifact, timing, infra, or security failure remains visible in the same scheduler and aggregate summaries.

Every failed retained summary that carries the standard failure fields MUST expose both a non-null `failure_class` and a non-null `failure_reason`. Passing summaries MUST expose no primary failure. A generic shell-wrapper exit such as `command exited with status 1` is diagnostic wrapper evidence when a tool runner has already emitted a classified failure for the same target; it MUST NOT become the primary failure or an independent primary harness failure.

Post-summary scheduler validation failures, including scheduler event, timing, critical-path, summary, or accounting drift detected after child work has completed, MUST be normalized as `failure_class=harness`, `failure_reason=scheduler_accounting_error`, and public exit code `11`. They MUST NOT fall through as caller `configuration_error` merely because they are detected by the scheduler runner.

Failure classification uses two layers:

- `failure_class`: coarse stable grouping for humans and automation.
- `failure_reason`: detailed snake-case reason for diagnosis, exit-code mapping, and handoff.

Go setup helpers MUST emit a test-scoped
`CARTULARY_HARNESS_TEST_FAILURE=` marker followed by exactly one bounded
`cartulary.harness_test_failure.v1` JSON envelope. The envelope contains only
normalized class and reason, setup source, service, readiness stage, attempt
count, and cleanup outcome. Credentials, endpoints, bucket or object identity,
database identity, and transport text are forbidden. The Go adapter associates
the marker only with the selected test or its descendants. Malformed or
conflicting markers are `harness/scheduler_accounting_error`; repeated identical
markers are one failure. When several selected rows fail, ordinary Section 9.1
primary-failure ordering remains deterministic. A Go `fail` event without a
valid setup envelope remains `product/test_assertion_failure`.

| Failure class | Meaning                                                                                 |
| ------------- | --------------------------------------------------------------------------------------- |
| `product`     | The product behavior under test failed after harness setup completed.                   |
| `security`    | A blocking security scanner finding was reported after harness setup completed.         |
| `config`      | Caller input, environment, manifest, or local tool configuration was invalid or missing. |
| `infra`       | Required backing infrastructure failed preflight, startup, readiness, or capacity.      |
| `harness`     | Harness orchestration, fixture, scheduler, child aggregation, or cleanup failed.        |
| `artifact`    | Required retained evidence was missing, malformed, invalid, or unsafe.                 |
| `timing`      | A deadline, timeout, or timing-accounting guard failed.                                |
| `interrupted` | The command was cancelled or interrupted.                                               |
| `unknown`     | The wrapper could not classify the failure.                                             |

| Failure reason                | Default class | Trigger                                                            |                                    Public exit code |
| ----------------------------- | ------------- | ------------------------------------------------------------------ | --------------------------------------------------: |
| success                       | none          | No failure                                                         |                                                 `0` |
| `usage_error`                 | `config`      | Invalid arguments, missing required flags, unsupported output mode |                                                 `2` |
| `configuration_error`         | `config`      | Missing/invalid tool, path, env, config, manifest, resource limit  |                                                 `2` |
| `preflight_error`             | `infra`       | Docker/platform/tool preflight fails before managed services       |                                                 `3` |
| `service_start_error`         | `infra`       | Backing service or browser process fails to start                  |                                                 `3` |
| `service_readiness_timeout`   | `infra`       | Started service fails readiness before deadline                    |                                                 `3` |
| `fixture_error`               | `harness`     | DB/bucket/template/reset/janitor/fixture operation or shape validation fails |                                      `3` |
| `resource_conflict`           | `infra`       | Logical resource, port, lock, DB/bucket name, host conflict, or confirmed filesystem capacity exhaustion (`ENOSPC`) | `4` |
| `test_assertion_failure`      | `product`     | Test runner assertion fails after harness setup                    |                                                `10` |
| `security_finding`            | `security`    | Blocking Govulncheck vulnerability or enforcing Gosec finding after scanner setup; Fallow uses this reason only if a later adopted security-scan profile selects it |                         `1` |
| `child_target_failure`        | `harness`     | Aggregate child exits nonzero                                      |                         normalized child class exit |
| `tool_diagnostic_failure`     | `harness`     | ShellCheck/Biome/Fallow-style static-analysis, formatter, linter, or tool diagnostic failure after setup |                         `1` |
| `scheduler_accounting_error`  | `harness`     | Manifest, summary, timing, event, or accounting mismatch           |                                                `11` |
| `boundary_policy_violation`   | `harness`     | Executable code or validation reads documentation outside the exact Section 15 exception registry |                         `11` |
| `test_accounting_unmapped`    | `harness`     | Public target completed with executed tests that were neither mapped nor intentionally classified |                  `11` |
| `artifact_error`              | `artifact`    | Required artifact missing, invalid, unredacted, or schema-invalid  |                                                `11` |
| `cleanup_error`               | `harness`     | Cleanup command/finalizer/leak check/reaper scheduling fails       |         `12` when no earlier primary failure exists |
| `duration_baseline_drift`     | `timing`      | Explicit duration-baseline or warm scheduler timing drift check fails |                                             `13` |
| `timeout_failure`             | `timing`      | Command, readiness, watchdog, cleanup, or lock exceeds deadline    |                                                `13` |
| `cancelled_or_interrupted`    | `interrupted` | Signal, cancellation, abort                                        | `130` for SIGINT, `143` for SIGTERM, otherwise `15` |
| `unknown_failure`             | `unknown`     | Failure cannot be classified                                       |                                                 `1` |

Default human output MUST expose bounded failure fields for a failed public target before GNU Make's generic recipe failure line can be the only visible diagnostic. It MUST omit full failure records unless verbose output is requested. The canonical compact shape is:

```text
failure_class=infra reason=service_readiness_timeout failed=<unit>
```

Full failure records belong in retained JSON summaries and investigation commands.

For `frontend-fallow-static`, missing Fallow tools, missing install state, invalid `.fallowrc.json`, invalid reachability owner input, missing owner-declared reachability files, or invalid public asset references MUST map to `configuration_error`; enforcing static findings MUST map to `tool_diagnostic_failure`; malformed raw reports or invalid normalized summaries MUST map to `artifact_error`. Non-blocking Fallow findings MAY be retained as warning evidence without failing the target.

**TH-HARNESS-REQ-305**
Missing or unknown owners, malformed `ROWS`, duplicate rows, cross-owner rows, explicit non-service rows on `service-backed-test-slice`, invalid worker bounds, invalid target-local JSON, removed inputs, and zero-row selections MUST use `failure_class=config`, `failure_reason=usage_error`, and exit `2` before setup.
Verified by: TH-HARNESS-AC-064

**TH-HARNESS-REQ-306**
Invalid owner manifests, verification contracts, schema IDs, runner definitions, selector resolution, profile references, duplicate case ownership, and policy registries MUST use `failure_class=config`, `failure_reason=configuration_error`, and exit `2` before child execution.
Verified by: TH-HARNESS-AC-062, TH-HARNESS-AC-063

**TH-HARNESS-REQ-307**
A product assertion after successful setup uses `test_assertion_failure` and exit `10`. Setup, service, fixture, browser, and scheduler infrastructure failures retain their existing Section 9 reasons. Missing, duplicate, or contradictory terminal row records use `scheduler_accounting_error` and exit `11`.
Verified by: TH-HARNESS-AC-065

**TH-HARNESS-REQ-308**
Missing, stale, mixed, ambiguous, unsafe, or semantically incompatible retained evidence uses `artifact_error` and exit `11`. Invalid caller path syntax remains `usage_error`; syntactically valid paths with incompatible contents are artifact failures.
Verified by: TH-HARNESS-AC-066

**TH-HARNESS-REQ-309**
An unauthorized documentation read detected by the Section 15 policy uses `boundary_policy_violation` and exit `11`. An invalid documentation exception or semantic allowlist is `configuration_error` before the scan. The policy failure MUST name the consumer and normalized path without disclosing document contents.
Verified by: TH-HARNESS-AC-067

**TH-HARNESS-REQ-310**
Configuration failures occur before setup and therefore precede no child failure. Once child execution begins, TH-HARNESS-REQ-304 primary-failure precedence remains unchanged. Cleanup failure MUST NOT replace an earlier primary failure; it becomes exit `12` only when no earlier failure exists.
Verified by: TH-HARNESS-AC-065

### 9.1 Primary Failure Selection

```text
select_primary_failure(failures):
  if failures is empty:
    return success
  if any non-cleanup failure exists:
    return non-cleanup failure by:
      1. failure-class precedence from TH-HARNESS-REQ-304,
      2. top-level command lifecycle order from TH-HARNESS-REQ-304,
      3. scheduler event sequence if scheduler-owned,
      4. child target registry order if aggregate-owned,
      5. artifact path lexical order,
      6. failure reason lexical order
  return earliest cleanup failure by cleanup step order
```

**TH-HARNESS-REQ-301**
Cleanup failure after an earlier product or operational failure MUST be recorded but MUST NOT override the public exit code selected for the earlier primary failure.
Verified by: TH-HARNESS-AC-014

**TH-HARNESS-REQ-302**
Harness setup, readiness, fixture, artifact, scheduler, timeout, and cleanup failures MUST NOT use `failure_class=product`. A failing assertion after successful harness setup MUST be classified with `failure_class=product` and `failure_reason=test_assertion_failure`.

Vitest and Playwright per-test timeouts after the test runner has reached product execution are product test failures and MUST be classified as `failure_class=product` with `failure_reason=test_assertion_failure`. Harness-owned watchdogs, command deadlines, lock deadlines, service readiness deadlines, and cleanup deadlines remain operational failures and MUST use `failure_reason=timeout_failure` or `failure_reason=service_readiness_timeout` according to the failure-reason table.

Integration tests for lease-conflict behavior MUST separate successful
precondition acquisition from the deliberately bounded conflicting
acquisition. A success-path database lease acquisition used only to establish
the test precondition MUST allow at least five seconds under declared
scheduler contention; it MUST NOT turn transient pool or database scheduling
latency into a product assertion. A short conflict deadline MAY be used only
after the test has proved that the conflicting lease is already held, and the
assertion MUST still require the exact conflict result.

The `browser-unit` Vitest project MUST declare a 15-second per-test timeout. This project exercises jsdom React composition, virtualization, and bounded multi-case interaction matrices while scheduled `frontend-unit` shares the declared host capacity with other graph work; the runner's generic five-second default is not a stable product-failure boundary for that workload. The harness-owned Vitest watchdog remains the outer hung-process boundary, and a test that exceeds the project timeout remains a product test failure. The `harness-node` project retains the runner default unless a separately adopted owner rule requires another bound. Individual browser-unit tests MUST NOT add a longer timeout merely to mask an unbounded wait or missing readiness condition.

Vitest assertion summaries that contain reporter stack-formatting markers such as `STACK_TRACE_ERROR` MUST preserve actionable assertion context when the runner report provides it. The retained diagnostic MUST keep the assertion title, owner path, raw runner-report reference, reproduce command, and a diagnostic tag such as `vitest_stack_trace_error`; the stack-formatting marker MUST NOT replace an available assertion message as the primary human diagnostic.

Vitest wrappers MAY retain `cartulary.vitest_failure_details.v1` sidecars under their raw artifact directory. When present and schema-valid, owner summaries MUST prefer a matching sidecar assertion message over the runner JSON's stack-formatting marker and MUST retain both the runner JSON and sidecar references in failure details. Missing sidecars remain compatible; summaries MUST continue to use the runner JSON diagnostic fallback when no sidecar entry is available.
Verified by: TH-HARNESS-AC-013, TH-HARNESS-AC-014, TH-HARNESS-AC-021

**TH-HARNESS-REQ-303**
A scheduler MUST preserve the first failed work unit and its retained detail as `failed_work_unit` and `failed_work_unit_detail` even when later sibling work drains and also fails. Scheduler summaries MUST include a bounded `observed_failed_work_units[]` array containing completed nonzero-exit work units in finish order, including later drained sibling failures; that array is diagnostic and MUST NOT rewrite the first failed work unit. Human failure output, scheduler summaries, target summaries, and tool-run summaries MUST choose a primary headline and public exit code from the primary-failure rules without contradicting `failed_work_unit` when the failed work unit has retained classified child evidence.

Scheduler `critical_path_wall_duration_ms` is the scheduler timing envelope and MUST equal the scheduler total duration for every emitted scheduler summary, including failed schedules. When a failed schedule completes no successful work unit, `critical_path_units[]` MUST be empty and `critical_path_terminal_unit` MUST be `null`; the failed path remains represented by `failed_work_unit_detail` and `observed_failed_work_units[]`.
Verified by: TH-HARNESS-AC-014, TH-HARNESS-AC-024

**TH-HARNESS-REQ-304**
Primary-failure precedence is closed. Failure-class precedence is exactly: `product`, `security`, `config`, `infra`, `harness`, `artifact`, `timing`, `interrupted`, `unknown`. Top-level command lifecycle order is exactly: wrapper identity, output-mode resolution, configuration resolution, result-root/run-ID resolution, redaction initialization, semantic target behavior, artifact validation, cleanup or finalizers, public output emission.

When class and lifecycle step tie, scheduler-owned failures order by scheduler event sequence; aggregate-owned child failures order by public child target registry order; artifact failures order by normalized artifact path lexical order; remaining ties order by `failure_reason` lexical order. A cleanup or finalizer failure MUST NOT override an earlier non-cleanup primary failure.
Verified by: TH-HARNESS-AC-014, TH-HARNESS-AC-032

## 10. Scheduler Contract

**TH-HARNESS-REQ-350**
Scheduler manifests are normative scheduler inputs. A scheduler target MUST validate manifest schema, work-unit IDs, dependencies, resource claims, finalizers, output schemas, and timing settings before starting child work.
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-024

**TH-HARNESS-REQ-360**
Owner selection begins after Section 5 input validation. For `test-slice`,
explicit `ROWS` selects exactly those rows; omitted `ROWS` selects every active
executable row owned by `OWNER`. For `service-backed-test-slice`, explicit rows
MUST all declare a non-`none` fixture capability; omitted rows select every
active owner row with such a capability. Tier does not affect either omission
rule. A zero-row result is `usage_error`.
Verified by: TH-HARNESS-AC-064, TH-HARNESS-AC-065

**TH-HARNESS-REQ-361**
Selection MUST set `selection_mode="default_owner"` for omitted `ROWS` and `selection_mode="exact_rows"` otherwise. `completion_scope` is `full_owner` for omitted selection over the command's complete dependency scope and `selected_subset` for explicit rows. `dependency_scope` is `all` for `test-slice` and `service_backed` for `service-backed-test-slice`.
Verified by: TH-HARNESS-AC-064, TH-HARNESS-AC-065

**TH-HARNESS-REQ-362**
Before setup, the planner MUST resolve every verification reference, runner,
selector, runtime profile, resource profile, fixture capability, logical
resource, cache policy, and expected artifact schema. It MUST reject zero or
multiple selector resolution, overlapping active ownership, dependency cycles,
missing producer edges, and incompatible contracts. No child may start after a
preflight failure.
Verified by: TH-HARNESS-AC-063, TH-HARNESS-AC-065

**TH-HARNESS-REQ-363**
The planner MUST sort resolved rows by `row_id`, construct deterministic work
unit and dependency IDs, derive exact runtime-binary prerequisites from the
authored execution topology, and validate an immutable
`cartulary.harness_work_graph.v5` before setup. The same semantic inputs MUST
produce byte-identical graph bytes.
Verified by: TH-HARNESS-AC-063, TH-HARNESS-AC-071

**TH-HARNESS-REQ-364**
Every resolved row MUST end in exactly one state:

| State | Required meaning |
| --- | --- |
| `passed` | Every required selected case executed and passed, and accounting is complete. |
| `failed` | A product assertion failed or a runner emitted an unauthorized skip. |
| `infrastructure_failed` | Setup, service, fixture, browser, or harness infrastructure prevented execution. |
| `skipped_dependency` | A declared failed dependency prevented the row from starting. |
| `cancelled` | A supported signal or explicit cancellation stopped execution. |
| `skipped_authorized` | One exact unexpired verification-level skip authorization applies. |

No other terminal token is current conformance.
Verified by: TH-HARNESS-AC-065

**TH-HARNESS-REQ-365**
Owner-slice success requires every resolved row to be `passed` or validly `skipped_authorized`. Missing records, duplicate records, dependency skips, cancellation, infrastructure failure, and unauthorized skips fail the invocation. A row passes only when every required selected case passes and required artifacts validate.
Verified by: TH-HARNESS-AC-065

**TH-HARNESS-REQ-366**
The current profile performs no automatic product-test retry. Existing service-startup and readiness retry policies in Section 11 remain unchanged. If a runner or service creates more than one attempt, every attempt MUST be retained and only the final row state participates in aggregation.
Verified by: TH-HARNESS-AC-065

**TH-HARNESS-REQ-367**
Both owner-slice commands use `stop_on_first_failure=false`. Running independent work drains after a failure; work whose dependency failed does not start. Cancellation MUST propagate through the existing scheduler contract. Finalizers and cleanup MUST always run, subject to the existing destructive-safety guards.

The owner-slice commands, browser commands, direct selectors, and aggregate
selectors MUST use the one graph compiler and scheduler facade specified by
TH-HARNESS-REQ-802; `test_slice` is not a distinct scheduling family. Owner
plans MUST retain `plan_semantic_digest` and `scheduler_semantic_digest`. These
digests exclude invocation timestamps and run IDs, and identical semantic
selections, profiles, work units, dependencies, resource claims, timeouts, and
finalizers MUST produce identical digests. Browser scheduler units MUST preserve
the catalog-derived browser session group, runtime profile, exact selected group
set, resource profile, and evidence target. A browser evidence finalizer MUST
run after every session reaches a terminal state, and the parent target summary
MUST run after every evidence finalizer reaches a terminal state.

Resource admission uses the ranked resource-fit and starvation-prevention rules
in TH-HARNESS-REQ-802. A ready unit that does not fit holds no resource. Ready
exclusive `host_activity` waiters bar later shared work. FIFO reservation,
high-water drain barriers, preemption, and capacity held for a dependency-blocked
or not-yet-ready unit are unsupported. Product work has one scheduler attempt. A dependency failure emits `skipped_dependency`; an
interrupt emits `cancelled`; an ordinary scheduler watchdog emits
`infrastructure_failed` with `timeout_failure`. Finalizers become ready when
every declared dependency is terminal, whether successful, failed, or
dependency-skipped. They run in deterministic manifest order after ordinary
running work drains. Cleanup failure is primary only when no earlier
higher-precedence failure exists.

The normative command algorithm is:

```text
validate command inputs
resolve OWNER
parse and validate explicit ROWS when present

if command is test-slice:
    selected = explicit rows, otherwise all active owner rows

if command is service-backed-test-slice:
    if explicit:
        reject any row not requiring managed services
        selected = explicit rows
    otherwise:
        selected = all active owner rows requiring managed services

reject empty selected set
resolve selectors and profiles
reject unresolved, multiply resolved, or overlapping cases
sort rows by row_id
emit immutable plan
execute with stop_on_first_failure=false
drain running work
do not start dependency-blocked work
always run finalizers
emit one terminal record per selected row
```
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-065

**TH-HARNESS-REQ-368**
Runner results MUST map exact executed cases back to one catalog row by the preflight-resolved selector inventory. An executed case owned by zero or multiple rows is `scheduler_accounting_error`. A runner-wide pass cannot close a row whose exact cases are absent.
Verified by: TH-HARNESS-AC-063, TH-HARNESS-AC-065

**TH-HARNESS-REQ-369**
The scheduler and evidence-accounting layers MUST preserve owner, family, verification, selector, profile, and evidence-class identity through work-unit expansion, batching, sharding, cancellation, finalization, and summaries. Batching MAY reduce process count but MUST NOT merge or weaken row identity.
Verified by: TH-HARNESS-AC-060, TH-HARNESS-AC-061, TH-HARNESS-AC-065

**TH-HARNESS-REQ-352**
Graph dependency declarations account for readiness work exactly once. A unit
whose command requires pnpm-managed packages depends on `frontend-install`; a
unit requiring a build artifact, runtime binary, browser runtime, service
image, or migrated fixture depends on the exact graph producer for that
artifact or capability. Work that does not require frontend packages MUST NOT
depend on the frontend install stamp.
Verified by: TH-HARNESS-AC-006

If a later profile promotes `frontend-fallow-static` into `check`, it MUST be a
direct graph unit with `frontend-install` readiness and the static Node resource
shape, not a hidden prerequisite of another target.

The public `check` binding MUST NOT run material install, build, service-image,
fixture, or browser readiness outside graph accounting. It may perform only the
minimal Node bootstrap and fail-fast input validation needed to start the graph
runner. All material readiness appears as a unit in canonical events and target
projections.

### 10.1 Scheduler Manifest Fields

The canonical scheduler projection schema is `cartulary.scheduler_manifest.v3`.
`tools/scheduler_manifest.json` contains generated selector-to-graph bindings;
it does not contain family-specific scheduler policies or expanded runtime work
units. Runtime execution accepts only `cartulary.harness_work_graph.v5` from the
single compiler. Non-current scheduler and graph schema IDs are unsupported.

The v3 scheduler manifest contains only:

| Field | Type | Required | Rule |
| --- | --- | ---: | --- |
| `schema_id` | string | yes | Exactly `cartulary.scheduler_manifest.v3`. |
| `generated` | object | yes | Generator and authored work-graph provenance. |
| `schedules` | array | yes | Empty in v3; phase/family runtime schedules are unsupported. |

Public, aggregate, direct, leaf, browser, backend, and owner-slice execution all
compile to `cartulary.harness_work_graph.v5`. The graph fields and validation
rules are closed by TH-HARNESS-REQ-801. Target names and command IDs MUST NOT be
overloaded to infer omitted selection metadata. Every resolved row maps to one
canonical row result and at least one target projection.

Runner concurrency is part of the unit command and claims. A worker pool MUST
keep its worker bound aligned with CPU, memory, process, and runner-specific
claims. Direct and aggregate selection of the same semantic work MUST preserve
that bound and digest. Runtime-binary, readiness, migration, fixture, browser,
reset, evidence, and finalizer relationships are exact graph producer edges;
ambient Make order and recursive prerequisites are not execution topology.

A Make child selected as a graph unit receives prepared run identity and may
skip only prerequisites already represented by exact graph dependencies. Direct
public invocation retains its normal prerequisites. The graph runner owns child
stdout/stderr capture, event timing, failure classification, unit results, and
target projections; a child MUST NOT emit a second scheduler or aggregate
summary for the same interval.

Estimated work is advisory scheduling input only. It is not a resource claim,
timeout, benchmark claim, or product performance threshold.
### 10.2 Logical Resource Registry

The current logical resource registry and its capacity-resolution policies are
closed by TH-HARNESS-REQ-353.
Resource claims distinguish CPU, memory, I/O, processes, volumes, ports,
service capabilities, and browser-stack leases. Fixture isolation is expressed
by fixture capability and broker lease rather than clone/reset scheduler
families. A caller override is accepted only through the v1 capacity-override
schema and only when every graph claim remains feasible.

**TH-HARNESS-REQ-353**
The logical resource registry is closed by the table below.
`tools/scheduler_resource_registry.json` is the v6 typed projection of this
current-conformance registry and MUST NOT independently add resources, change
bounds, change override inputs, or redefine automatic policies. Capacity
resolution MUST interpret that typed projection; hard-coded fallback
capacities in the scheduler are forbidden.
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-030

| Resource | Snapshot source | Unit | Omission behavior |
| --- | --- | --- | --- |
| `cpu` | effective CPU set | token | fail safe to one |
| `memory_mb` | effective memory bound | mebibyte | fail safe to one unit's minimum |
| `io` | host I/O allowance | token | fail safe to one |
| `process` | effective process limit | slot | fail safe to one |
| `volume` | writable-volume proof | capability | unavailable when unproven |
| `port_lane` | broker port-lane inventory | lane | fail safe to one |
| `postgres` | service readiness | capability | unavailable when unready |
| `object_store` | service readiness | capability | unavailable when unready |
| `service_stack` | managed-stack readiness | capability | unavailable when unready |
| `browser_stack` | broker browser-stack capacity | lease | fail safe to one |

The capacity policy projection is closed as follows. Positive integral
overrides remain subject to the v1 override schema and to the host bounds below.

| Snapshot field | Automatic policy | Bound |
| --- | --- | --- |
| `cpu_tokens` | Effective host parallelism. | Minimum `1`. |
| `memory_bytes` | Minimum of the positive cgroup bound and host memory when both exist; otherwise the positive available value. | Minimum `1`. |
| `process_slots` | Minimum of the effective process limit and four times CPU tokens; use four times CPU tokens when the process limit is unavailable. | Minimum `1`. |
| `io_tokens` | Two times CPU tokens. | Minimum `1`. |
| `postgres_lanes` | Eight lanes by default. | Minimum `1`; at most CPU tokens. |
| `object_store_lanes` | Four lanes by default. | Minimum `1`; at most CPU tokens. |
| `port_lanes` | Four lanes by default. | Minimum `1`; at most CPU tokens. |
| `writable_volume` | Writable-root proof. | Boolean capability. |

The single capacity override declaration file accepts positive bounded capacities for
registered resources only. A value below a declared claim or an infeasible
graph is `configuration_error`, exit `2`, before child work. Inline override
JSON, family-specific
capacity profiles and forwarding environment variables are unsupported.

**TH-HARNESS-REQ-354**
Resource capacity MUST be resolved once from the immutable capability snapshot
owned by TH-HARNESS-REQ-803. Every graph unit declares claims using the closed
registry in TH-HARNESS-REQ-353; fixture capabilities are broker leases, not
synthetic clone or reset resource families. The scheduler computes each unit's
feasible lane count as the minimum across its claimed dimensions and MUST reject
an impossible claim before child work. The optional capacity override is one
schema-validated declaration recorded in the run manifest. A target family,
aggregate, or child environment MUST NOT recalculate or forward capacity.

A Go work unit with a PostgreSQL service dependency MUST bound its in-process
Go test parallelism to the minimum of its planned CPU parallelism, captured
PostgreSQL lane capacity, and topology-declared PostgreSQL claim. Its
`GOMAXPROCS`, CPU claim, and PostgreSQL claim MUST all declare that same bounded
parallelism. This rule also applies to raw Go aggregate units with a PostgreSQL
dependency. A package process MUST NOT claim one PostgreSQL lane while retaining
greater implicit Go test fan-out. Non-PostgreSQL Go units retain their ordinary
LPT CPU plan. This is scheduler admission, not fixture sharing: every test still
receives the isolation required by its fixture capability.

Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-018, TH-HARNESS-AC-030

**TH-HARNESS-REQ-355**
Harness performance acceptance is governed exclusively by
TH-HARNESS-REQ-377 through TH-HARNESS-REQ-380. No fixed global wall-time cap is
normative. Every qualifying window MUST preserve command, source, system,
semantic, catalog-row, evidence-route, capacity, cache-mode, artifact, and
cleanup identity. Failed, interrupted, stale, retry-contaminated, or mismatched
runs remain retained diagnostics and MUST NOT enter a qualifying window.
Verified by: TH-HARNESS-AC-059

**TH-HARNESS-REQ-356**
An unsharded pure Go target MAY execute multiple exact-symbol evidence rows in
one execution family only when every row has compatible package selection,
runtime-binary requirements, fixture capability, resource claims, and isolation
policy.
The owner catalog remains the evidence-routing owner. Consolidation MUST preserve
every row ID, verification ID, evidence class, selected symbol, profile, and
minimum tier, and the Go JSON report MUST still prove
each exact selected symbol. Raw package-wide selectors remain separate from
exact-symbol families. A missing, ambiguous, duplicate, or unexpectedly
selected symbol MUST fail closed; it MUST NOT be repaired by inferred row
membership or treated as derived success.

Private execution-family IDs MAY be replaced atomically when no continuing
external consumer is demonstrated. Current producers, consumers, diagnostics,
fixtures, and duration accounting MUST move together; no alias, dual reader,
dual emitter, or forwarding shim is permitted for the retired private IDs.
Verified by: TH-HARNESS-AC-060

**TH-HARNESS-REQ-357**
Service-backed exact-symbol Go evidence items MAY share one scheduler shard only
when they have the same target, package selection, runtime-binary set, fixture
capability, process-isolation policy, resource profile, environment, selector
mode, and authoritative or support evidence class. Raw package selectors,
process-isolated items, and incompatible items remain separate. In the current
capability model `managed_process` is inherently process-isolated. A Go row
whose assertion owns process-global test state MAY declare the explicit
`exclusive` policy; this is an assertion-compatibility boundary, not a resource
capacity or fixture substitute. Dedicated and migration PostgreSQL helpers and
per-test object-store helpers instead create, register, and eagerly retire an
exact resource for each `testing.T`; their rows MAY share a process without
sharing that resource. Within each compatible execution
family the planner MUST use the capacity-derived longest-processing-time
algorithm in TH-HARNESS-REQ-805. Fixed symbol counts, fixed duration ceilings,
and target-specific packing exceptions are unsupported. Package and command
startup overhead is charged once to the shard process and MUST NOT be multiplied
by row count.

Every packed item MUST retain its row ID, scenario ID when declared, symbol,
owner, coverage class, fixture proof, duration key, and runtime-binary
requirements in the shard plan and emitted evidence. The runner MUST use an
anchored exact-symbol selector and MUST fail closed on unexpected, duplicate,
missing, crashed, or incomplete output; unresolved rows MUST remain missing.
Packing and generated shard names MUST be byte-deterministic from authored
inputs, qualified duration owners, and the captured Go-lane capacity. Private
family and shard IDs have no compatibility contract. Public row outcomes,
failure classes, evidence closure, and cleanup behavior remain exact.
Verified by: TH-HARNESS-AC-061

**TH-HARNESS-REQ-358**
Scheduler-manifest growth gates MUST remain meaningful when an optimization
reduces the number of ordinary work units. The report MUST retain total bytes,
unit count, bytes per unit, per-kind p95, and maximum serialized-unit
measurements. Scratch graph compiles MUST be byte-identical for the same
semantic selection. A gate MUST compare like unit kinds and MUST NOT permit a
reduced unit count to conceal growth in a wide unit. Thresholds are
machine-owned policy inputs so future unit kinds can be added without revising
this owner contract.
Verified by: TH-HARNESS-AC-061

**TH-HARNESS-REQ-359**
A `postgres_migration` lease MUST expose a fresh database identity for each test
owner and retain a full replay of the selected migration chain. The broker owns
that database until the test's eager lease cleanup, closes every client before
destruction, and records teardown evidence. Multiple compatible rows in one Go
process MUST still receive distinct database identities and independent cleanup
records. A digest-keyed migrated template MAY accelerate a later lease only
after the required same-chain proof; it MUST NOT replace the full-chain proof or
cross a run without a matching digest. Attached services remain borrowed and
MUST NOT be closed by lease cleanup.
Verified by: TH-HARNESS-AC-061

**TH-HARNESS-REQ-370**
Every graph execution MUST emit `cartulary.harness_unit_event.v2` evidence that
represents each work unit with its stable ID, declared dependency edges,
normalized resource claims, eligibility instant, start instant, and terminal
instant and state.
Eligibility is the first scheduler iteration in which all declared dependencies
are complete or the unit becomes terminally dependency-blocked. Queue wait is
elapsed time from eligibility to start. A unit that never starts records its
terminal dependency or cancellation state without a fabricated execution
duration. All boundaries use the scheduler process's monotonic clock; child
processes MUST NOT synthesize monotonic values from wall-clock timestamps.
Verified by: TH-HARNESS-AC-074, TH-HARNESS-AC-077

**TH-HARNESS-REQ-371**
For every work unit the event stream contains exactly one `eligible` boundary.
Every eligible but unstarted interval contains exactly one `wait_started` and
one matching `wait_ended` boundary with the closed wait reason `resources`,
`earlier_overlapping_ready`, `capacity`, or `scheduler_stop`, plus sorted
blocking logical resources and blocking unit IDs captured at wait start. Holder
changes MUST NOT emit another wait event. The boundaries remain required when
the interval is zero. `wait_ended` precedes `admitted` or an unstarted terminal
event at the same monotonic instant. A duplicate, missing, differently reasoned,
or unmatched boundary is `scheduler_accounting_error`. Resource-blocked
duration is the union of intervals
attributable to one logical resource; overlapping observations MUST NOT be
summed.
Verified by: TH-HARNESS-AC-074, TH-HARNESS-AC-077

**TH-HARNESS-REQ-372**
The actual dependency critical path is computed over observed dependency edges
using each node's queue-wait and execution intervals. For each node, path cost
is its attributable duration plus the maximum predecessor path cost, with
stable unit ID tie-breakers. The run summary MUST name the
ordered path as `critical_path` and its duration as
`actual_dependency_critical_path_ms`. It MUST NOT rewrite the overall run
envelope or project queue time into child execution.
Verified by: TH-HARNESS-AC-074

**TH-HARNESS-REQ-373**
Unattributed envelope time is parent duration minus the union of directly
attributed child intervals clipped to the parent interval. Negative results
clamp to zero and set a bounded clock/accounting warning. Child durations MUST
NOT be summed when they overlap. Graph-unit, broker, runner, projection, and
finalizer intervals participate only under their direct parent.
Verified by: TH-HARNESS-AC-074

**TH-HARNESS-REQ-374**
Every public aggregate MUST compile directly to the union of its selected
policy and row units. Dependencies name exact producer unit IDs; a whole-target
barrier, phase wrapper, child aggregate, nested scheduler, or capacity-forwarded
child invocation is invalid. The task-surface owner retains public command and
summary metadata; the graph owner retains unit dependencies, claims, policies,
and projections. Validation MUST reject duplicate semantic units with different
digests, cycles, unknown dependencies or resources, and infeasible claims before
execution.

Readiness producers appear once and are dependencies only of units that consume
their outputs. Aggregate membership does not add an edge. Duration-baseline
drift consumes terminal retained evidence and remains an explicit post-run
maintenance gate. Source-boundary commands MUST index each repository tree once
per invocation and reuse immutable file text across their independent checks.

The advisory gosec audit owns one physical repository scan, preserving the
stable union of its runtime and support rules. `go-vulncheck` is one semantic
unit in any graph and follows the freshness rules in TH-HARNESS-REQ-807. Direct
and aggregate selection MUST produce the same scan unit ID, findings, artifacts,
exit behavior, and target projection.
Verified by: TH-HARNESS-AC-077, TH-HARNESS-AC-078

**TH-HARNESS-REQ-375**
Browser concurrency is determined by work-unit claims, the captured capability
snapshot, and broker-owned `browser_stack` leases. No target owns a fixed stage
capacity and the retired `release-browser-readiness` target has no v3 successor.
Direct and aggregate browser selectors use the same semantic groups. Independent
read-only groups MAY overlap; stateful affinity chains, measurement quietness,
snapshot write exclusion, quarantine, and cleanup are governed by
TH-HARNESS-REQ-806.
Verified by: TH-HARNESS-AC-077, TH-HARNESS-AC-078

### 10.5 Harness Observability Metrics and Performance Acceptance

**TH-HARNESS-REQ-376**
The current harness metric registry is closed to
`cartulary.harness.invocation.duration`,
`cartulary.harness.dependency.critical_path`,
`cartulary.harness.scheduler.queue_wait`,
`cartulary.harness.scheduler.resource_blocking`, and
`cartulary.harness.invocation.unattributed`. Duration units are milliseconds.
Allowed dimensions are the closed safe
target, command, family, runner, work-unit kind, status, timing bucket, wait
reason, and logical-resource tokens present in the native bundle. Run IDs,
paths, process IDs, test output, and raw symbol names are forbidden metric
dimensions.
Verified by: TH-HARNESS-AC-075, TH-HARNESS-AC-076

**TH-HARNESS-REQ-377**
A qualifying target/provider performance window contains one successful
forced-cold observation, one consecutive successful warm-up observation, and
five consecutive successful measured observations. The cold and warm-up roots
are retained and validated but excluded from warm statistics. A sixth measured
observation MUST be added in matched reference/candidate pairs when either
five-observation window has median absolute deviation greater than ten percent
of its median or when removing any single observation changes the gate result.
No window may contain more than six measured observations. If the six-run
comparison remains unstable, acceptance and baseline publication MUST fail.

Every root in one window MUST match provider, command ID, canonical inputs,
timing source, workload/evidence digest, host profile, externally available
capacity, toolchain profile, target execution policy, commit,
source-snapshot digest, graph digest, and clean source state. Each window MUST
therefore name one clean frozen commit and source snapshot. A reference or
candidate portfolio MAY combine independently qualified windows from more than
one frozen source state; this supports retaining an already qualified target
window when a later source change affects another target. Every target remains
bound to exactly one window, and the baseline artifact MUST close an exact
sorted source-window roster over every target, commit, and source-snapshot
digest. An observation MUST NOT be copied to, relabeled as, or otherwise
attributed to a source state other than the one retained by its execution
context. Across one target's reference and candidate windows, every retained
field above except commit, source snapshot, graph digest, command ID, and a
declared normalized policy projection MUST match. A changed graph, command, or
policy MUST be the reviewed successor declared for that target; comparing only
an opaque changed digest is insufficient.

For each window, `p50` is the median, `p90` is the nearest-rank ninetieth
percentile, and `MAD` is the median absolute deviation from `p50`. For a paired
comparison, `variability_band = 3 * max(reference MAD, candidate MAD, 1ms)`.
A required improvement passes only when candidate p50 is lower than reference
p50 by more than the variability band and candidate p90 does not exceed
reference p90 plus the band. A no-regression gate passes only when candidate
p90 does not exceed reference p90 plus the band.
The normalized execution-policy projection MUST use the Section 3.6 semantic
JSON encoding. Its `execution_policy_sha256` value is the same SHA-256 digest
without the `sha256:` prefix. Producers and validators MUST use that one
recursively key-sorted, I-JSON-safe encoding; formatted bytes, object insertion
order, and a second policy-specific canonicalizer are forbidden digest inputs.
Failed, interrupted, stale, retried, source- or capacity-mismatched observations
are retained as rejected evidence and do not enter either set. Performance
acceptance MUST consume retained execution contexts, record every rejected root
and reason, and MUST NOT derive a historical profile from the current checkout
or trust manifest-declared digests over the root's retained context.
Verified by: TH-HARNESS-AC-079

**TH-HARNESS-REQ-378**
Required improvement targets are `browser-e2e`, its affected public browser
leaves, `test-fast`, `test`, `check`, `ci`, `release-check`,
`agent-finalize`, `harness-contract`, and `go-vulncheck`. Every other retained
required public target uses the no-regression gate. Tier-reduced public timing
MUST be reported separately from unchanged-workload executor improvement.

Exact timing sources are the complete public invocation envelope, an
exact-once canonical graph-unit interval union, and named non-overlapping
canonical setup, fixture, execution, collation, and wrapper interval unions.
An aggregate run MAY provide a leaf sample only when canonical unit events
prove the same command, inputs, workload, capacity contract, and one exact
semantic occurrence; otherwise the command MUST be run directly. Dispatch-only
timing, an aggregate command envelope charged to multiple children, and a
material interval not attributed to one canonical bucket are invalid evidence.
For an aggregate-provided canonical graph-unit observation, the writer MUST
read the retained scheduler event streams, find exactly one successful
`start`/`finish` pair whose `work_unit_id` equals the bound target, and derive
the interval from those events. A missing boundary, duplicate occurrence,
nonzero terminal status, reversed interval, or target-name inference is invalid
evidence. A reconstructed target span is permitted only when no canonical
scheduler unit exists and the retained bundle proves one exact successful
target occurrence.

A v3 evidence-roots manifest deduplicates explicit reference or candidate
windows and binds every target to one window and timing source.
`mode=baseline` contains only reference windows and bindings;
`mode=comparison` also contains candidate windows and bindings. The public
roster, binding roster, observed target roster, declared target count, and
recomputed portfolio sum MUST equal the exact sorted
`observability_policy.required_targets` projection from the task-surface owner.
The baseline's derived `source_windows` collection MUST equal the exact sorted
grouping of public target rows by their retained source commit and snapshot;
missing, duplicate, empty, or incorrectly attributed source-window membership
is invalid evidence.
The current owner roster contains 50 public targets, including
`openapi-compatibility-check`. Internal diagnostics are stored in a separate
collection and MUST NOT affect public roster closure or portfolio totals.
Every public target MUST pass its individual gate. Default `make check` MUST
NOT enforce the performance drift gate.
Verified by: TH-HARNESS-AC-079

**TH-HARNESS-REQ-379**
Compatible pure backend-unit exact-symbol groups MUST be partitioned directly
by normalized package selection, sorted runtime-binary set, complete fixture
profile, fixture policy and budget, isolation policy, and authoritative or
support evidence class, then execute with
`min(group_count,clamp(floor(available_parallelism/4),1,8))` workers. Each Go
child receives scheduler-owned `GOMAXPROCS=max(1,floor(available_parallelism/workers))`.
This child scheduler partition applies only to backend-unit grouped capture.
Other backend targets retain the available host parallelism assigned to their
own scheduler unit; backend finalization MAY reuse the bounded worker-count
formula without rewriting capture-child `GOMAXPROCS`.
All exact symbols in one
compatible package/runtime/fixture/isolation/evidence-class group MAY share one
Go JSON process. Raw package selectors remain separate. Each physical Go report
MUST be parsed once; immutable family-projection requests then execute through
the same host-derived worker pool. Each worker MUST initialize the output
runtime once and process multiple requests sequentially; starting one cold
output runtime per family is forbidden. Target owner-evidence finalization MUST
load one immutable catalog/accounting context and reuse it for every owner
partition in that target. Production MUST NOT read or mutate ambient environment
to control this private worker limit. Stable row ordering,
target-summary ordering, and Section 9 primary-failure ordering MUST be applied
after all bounded work settles. Capture or report-worker failure MUST preserve
any already successful row evidence and select one primary failure through the
existing taxonomy.
Verified by: TH-HARNESS-AC-078

**TH-HARNESS-REQ-380**
`harness-public-target-duration-baselines` is the sole writer of
`tools/harness_public_target_duration_baselines.v3.json`. It MUST accept only an
exact v3 `mode=baseline` manifest with one accepted forced-cold root, one
accepted discarded warm-up root, and five or conditionally six accepted
measured roots per target/provider window. It MUST independently verify each
context, canonical unit-event stream, and bundle; derive observations from the
binding's exact timing source; reject dispatch-only or multiply attributed
timing; compute the Section 10.5 target and portfolio statistics; close the
writer's exact roster, binding, row, count, and total against the task-surface
owner; write deterministic normalized bytes; and retain a bounded maintenance
summary. Internal diagnostics MUST be emitted separately and MUST NOT enter the
public target array, count, or total. The writer MUST reject dirty, failed,
interrupted, retried, duplicate, profile-mismatched, missing-command,
wrong-cardinality, unstable, unattributed, and non-closing roots. Hand editing,
partial refresh, inferred roots, newest-run selection, translation of malformed
retained baselines, and v1 or v2 readers are forbidden.
Verified by: TH-HARNESS-AC-079

Generated source publication MUST preserve the repository-owned mode declared
for each output. A regular generated source file is published as `0644` on
POSIX hosts; owner-only `0600` applies to scratch and retained evidence, not to
the final generated source. A generator that renders bytes identical to the
existing output MUST avoid replacement and MUST normalize an incorrect final
mode before reporting the output unchanged. Temporary siblings MAY remain
owner-only while being written, but their final mode MUST be set before the
atomic rename. A successful generation or finalization command MUST leave the
Section 6 source-snapshot digest unchanged when no generated bytes changed.
Verified by: TH-HARNESS-AC-079

`work_units[].timeout_seconds` is optional and, when present, MUST be an integer from `1` through `3600`. It is a scheduler-owned watchdog around the whole child process group, not a product-performance assertion. Expiry MUST terminate the child group, retain the partial redacted log, record `failure_class=timing` and `failure_reason=timeout_failure`, return `13`, drain already-running independent work, mark dependency-blocked work `skipped_dependency`, and then run finalizers. A finalizer has its own timeout and MUST NOT inherit the aborted work signal. Omitting the field delegates deadlines to the narrower service, browser, runner, or child-target contract. Product assertions MUST have exactly one scheduler attempt; scheduler watchdog expiry MUST NOT create a retry.
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-030, TH-HARNESS-AC-065

### 10.3 Scheduling Algorithm

```text
validate graph and capture immutable capability snapshot
pending = work_units by stable unit ID
running = empty map
terminal = empty map
primary_failure = null
admission_stopped = false
emit run_started

while pending is not empty or running is not empty:
  mark units with a failed required dependency as skipped_dependency
  ready = pending units whose dependencies all succeeded
  compute effective rank = critical-path work + monotonic wait age
  identify the highest-ranked feasible ready unit

  for unit in ready by effective rank DESC, unit ID ASC:
    if admission_stopped:
      break
    if resources_available(unit.resource_claims):
      acquire declared fixture lease
      start unit
      remove unit from pending
      add unit to running
      emit unit_started

  if running is empty:
    fail with scheduler_accounting_error for a deadlock or impossible claim

  wait until one or more running units finish or cancellation arrives
  process simultaneous completions by terminal monotonic time, then unit ID

  for each finished unit:
    release resource claims and clean or return the fixture lease
    record status, logs, evidence, and duration
    if failure:
      select primary failure under Section 9
      stop admission only when the target failure policy requires it

on cancellation, stop admission and terminate owned process groups
run selected graph finalizers after their exact terminal dependencies
release every owned lease
validate events, evidence closure, and target projections
emit run summary and target summaries
exit with selected primary failure
```

Monotonic aging eventually raises a waiting feasible unit ahead of newly ready
work. Dependencies outrank rank and no admission preempts a running unit. A
finalizer failure is primary only when no earlier non-finalizer failure exists.
Dependency skips are propagation records, not additional root failures.

### 10.4 Event Ordering

| Event field             | Rule                                                                         |
| ----------------------- | ---------------------------------------------------------------------------- |
| `schema_id`             | `cartulary.harness_unit_event.v2`.                                           |
| `unit_id`               | Stable semantic work-unit identity.                                          |
| `needs`                 | Stable ASCII-sorted declared dependency IDs.                                 |
| `seq`                   | Starts at `1`, increments by `1`, no gaps.                                   |
| `event`                 | One closed v2 lifecycle, fixture, cache, or run-boundary token.              |
| `monotonic_ms`          | Non-decreasing scheduler-relative monotonic time.                            |
| Work-unit ordering      | Stable unit ID unless completion tie rule applies.                           |
| Completion tie          | Terminal monotonic instant ascending, then stable unit ID ascending.         |
| Artifact ordering       | Lexicographic by normalized artifact path.                                   |
| Resource ordering       | Registry display order, then lexicographic fallback.                         |

## 11. Service and Fixture Lifecycle

**TH-HARNESS-REQ-400**
Service-backed targets MUST run in exactly one service mode: `owned` or `attach`.
Verified by: TH-HARNESS-AC-007

| Mode     | Selection rule                                      | Missing variables                                                      | Ownership                                                                     |
| -------- | --------------------------------------------------- | ---------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| `owned`  | `CARTULARY_TEST_SERVICES_MODE=owned` or omitted.    | Not applicable.                                                        | Harness starts and cleans suite resources.                                    |
| `attach` | `CARTULARY_TEST_SERVICES_MODE=attach`.              | Any missing or invalid session descriptor fails with `configuration_error`. | Harness borrows descriptor-proven containers and deletes only run-owned resources. |

### 11.0 Postgres Fixture Model

**TH-HARNESS-REQ-405**
Fixture selection MUST be intent-based so catalog growth does not turn setup
into hidden critical-path cost. The catalog's explicit `fixture_capability` is
the sole row-level isolation request. Helper code MUST fail closed when its
requested capability differs from the selected row or when no active broker can
satisfy it.
Verified by: TH-HARNESS-AC-007, TH-HARNESS-AC-018

| Capability | Intended use | Required isolation |
| --- | --- | --- |
| `none` | Pure work with no managed fixture. | No broker resource. |
| `postgres_transaction` | Store work whose writes can be fully rolled back. | One transaction; committed-state observation is rejected. |
| `postgres_dedicated` | Committed state, DB identity, process, recovery, or isolation-sensitive work. | One clean database leased to each test owner; compatible rows may share only the Go process. |
| `postgres_migration` | Upgrade, downgrade, replay, boundary, or backfill paths. | One fresh database per test owner and retained full-chain proof. |
| `object_store_namespace` | Isolated object-store effects. | Unique test-owned namespace unless the explicit package-reuse helper owns the scope. |
| `managed_process` | Process lifecycle or binary evidence. | Owned process group and unique ports. |
| `browser_stack` | Browser groups and their application stack. | Broker-owned DB, namespace, processes, and port set. |

The migrated Postgres template key is derived from sorted migration inputs and
the migration-runner identity. Lease events and summaries MUST include schema
digest, capability, ownership, and compatibility key where applicable. Attach
mode MUST fail before child execution when the advertised schema digest does
not match local migration inputs.

Transaction leases MUST prove rollback and reject work that uses a connection
outside the rollback transaction or requires committed cross-connection
visibility. A transaction row MUST be explicitly admitted by the typed
PostgreSQL fixture policy registry; registry omission fails catalog validation.
Dedicated and migration leases MUST be unique to one test owner. Compatible
exact-symbol rows MAY share a Go process, but MUST NOT share the database,
database handle, eager finalizer, or cleanup record. Rows
involving process lifecycle, recovery, schema mutation, DDL,
advisory locks, database identity, background work, WebSocket observers, or any
unproven cleanup surface MUST use dedicated isolation rather than transaction
scope. The registry validator MUST close every active PostgreSQL row
from the current catalog and reject stale or surplus shared-scope approvals.

Performance-fixture builders MUST use an explicit builder policy owned by the
performance-fixture profile. The current builder profile claims bounded CPU,
memory, I/O, and process resources plus one PostgreSQL lane. It MUST NOT inherit
the browser consumer profile and MUST NOT claim a browser stack, port lane, or
object-store lane merely because the resulting fixture is later consumed by a
browser measurement.

The broker maintains an asynchronous supply of clean dedicated databases.
Failed reset, health, cleanup, or contamination validation quarantines the lease
and replaces it; a quarantined lease is never returned to the pool. Cleanup is
idempotent and records whether each resource was returned, destroyed, borrowed,
or already absent. Cancellation and partial startup follow the same ownership
rules. No capacity increase may substitute for assigning the least expensive
capability that still proves the row's isolation contract.

Every per-test dedicated or migration database MUST be deleted eagerly after
its owning application runtime, background work, pool, and handle have closed.
The ordinary path MUST attempt a non-forced drop of the exact harness-issued
database first. One atomic forced drop, including PostgreSQL's connection
termination semantics, is a bounded fallback permitted only after exact run
ownership has been proved and the ordinary drop either reports active
connections or exhausts its own bounded attempt while the parent cleanup
context remains active. The normal attempt and forced fallback MUST have
separate bounded contexts; cancellation of the parent MUST prevent the forced
fallback. Every cooperating harness-owned database create, clone, template
state mutation, exact-owner normal drop, and forced drop against one service
PostgreSQL catalog MUST first acquire a target-specific exclusive
session-advisory lock. A target's create, mutation, and cleanup operations
therefore cannot overlap, while unrelated targets remain independent.
PostgreSQL lock wait ordering supplies same-target admission; polling and
try-lock loops are forbidden. Admission uses one bounded context and release
uses a fresh bounded unlock context. The normal and forced database operations
retain their own separately bounded contexts, and the normal attempt MUST
release target admission before a fallback reacquires it. No outer aggregate
deadline may silently truncate an operation budget after its admission. Exact
owner proof plus target exclusion is the destructive boundary for the atomic
forced drop; service-global quiescence and hash-derived catalog stripes are
neither required nor permitted. Global catalog concurrency belongs exclusively
to the graph's typed PostgreSQL resource claims, including an honest claim for
any contract fixture that deliberately overlaps multiple targets. A split
terminate-then-drop fallback is not current because it can exhaust one cleanup
budget between the two catalog operations. A failed or
interrupted eager cleanup remains in the private
resource ledger for suite teardown or stale recovery; those mechanisms are
idempotent fallbacks and MUST NOT be the successful hot path. Successful
per-test databases MUST NOT emit retained-resource events or appear in an
exhaustive retained summary list. Per-test object-store buckets and prefixes
follow the same owner-finalization rule. Explicit object-store package reuse
remains package-scoped and is cleaned or reset only at its declared package or
suite boundary. A browser session namespace remains stable across stateful
generations: reset MUST empty the exact owned bucket in place and prove its
put/head/delete mutation path before the next generation is admitted. Reset
MUST NOT delete and recreate that namespace. Successful session finalization
deletes the namespace; failed reset or finalization remains recoverable through
the private ledger.

### 11.1 Lifecycle Machine Contract

**TH-HARNESS-REQ-403**
A lifecycle machine is normative only when this NLSpec explicitly labels it normative. A representational lifecycle diagram MUST be labeled non-normative, MUST cite its owning requirements, and MUST NOT add behavior. A normative harness lifecycle machine MUST define scope, instance key, closed state set, closed event set, terminal states, transition table, guard precedence, failure mapping, authoritative state derivation, observable evidence, and conformance criteria. Illegal transitions MUST NOT mutate state, MUST fail closed with Section 9 failure classification, and MUST emit retained evidence. State-advancing artifact writes MUST be atomic and idempotent, or the machine MUST define guardrails that prevent unsafe re-execution. Parent lifecycle logic MUST depend only on child terminal status and retained artifacts, not on child in-memory state.
Verified by: TH-HARNESS-AC-017

Implementations MAY realize a normative lifecycle machine with ordinary control flow, tables, generated code, or a state-machine library. The runtime mechanism is not normative. The closed states, events, transitions, failure mapping, and observable evidence are normative.

Normative lifecycle-machine state and event names MUST be ASCII `lower_snake_case`. A transition table is closed by default: any `(state, event)` pair not listed by the owning machine is illegal.

### 11.2 Normative Service Suite Lifecycle Machine

**TH-HARNESS-REQ-404**
The service suite lifecycle machine is normative for every service-backed suite in `owned` or `attach` mode. The machine ID is `test_services_suite_lifecycle_v1`. The machine instance key is `suite_id`. The authoritative transition record is `_shared/test-services/<suite-id>/lifecycle-events.jsonl`, where every line MUST validate as `cartulary.test_services.lifecycle.v2`. The current state is `requested` before the first lifecycle event and otherwise the `to_state` of the last valid lifecycle event. A missing, malformed, non-sequential, or transition-invalid lifecycle event stream after a suite directory or lease exists MUST fail closed with `failure_class=artifact` and `failure_reason=artifact_error`. The service lease remains cleanup-proof evidence and MUST NOT be interpreted as a transition log.
Verified by: TH-HARNESS-AC-007, TH-HARNESS-AC-017

Lifecycle event `seq` starts at `1`, increments by `1`, and has no gaps. Events MUST be processed in emitted sequence order. When competing conditions are observed before the next event is emitted, guard precedence is:

| State           | Precedence rule                                      |
| --------------- | ---------------------------------------------------- |
| `starting`      | `startup_failed` before `readiness_passed`.          |
| `running_child` | `interrupt_received` before `child_started` before `child_finished` when multiple child signals are observed before the next event is emitted. |
| `cleaning`      | `cleanup_failed` before `cleanup_succeeded`.         |
| all others      | The transition table has at most one allowed event.  |

#### States

| State           | Kind         | Invariants                                                                 | Observable signals                                                                 |
| --------------- | ------------ | -------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| `requested`     | initial      | Suite setup has been requested but no lifecycle event has been emitted.    | No lifecycle event stream exists for the selected `suite_id`.                      |
| `starting`      | intermediate | Suite setup or attach-mode validation is in progress; child work has not started. | Latest lifecycle event has `to_state=starting`; lease or startup diagnostics exist when the suite writes them. |
| `ready`         | intermediate | Required services or supplied attach endpoints have passed readiness.      | Latest lifecycle event has `to_state=ready`; readiness diagnostics are retained when produced. |
| `running_child` | intermediate | One or more child work units are executing under the suite.                 | Latest lifecycle event has `to_state=running_child`; the event references child logs or target artifacts when known and reports `active_child_count`. |
| `interrupted`   | intermediate | Cancellation or interruption was observed while child work was active.     | Latest lifecycle event has `to_state=interrupted` and `failure_reason=cancelled_or_interrupted`. |
| `cleaning`      | intermediate | Owned teardown or attach-mode diagnostic finalization is in progress.      | Latest lifecycle event has `to_state=cleaning`; lease `cleanup_state=in_progress` when a lease exists. |
| `cleaned`       | terminal     | Required cleanup or attach-mode finalization completed.                   | Latest lifecycle event has `to_state=cleaned`; lease `cleanup_state=completed` or `deferred`. |
| `failed_start`  | terminal     | Startup, attach validation, preflight, or readiness failed before child work started. | Latest lifecycle event has `to_state=failed_start`; failure summary records the Section 9 reason. |
| `cleanup_failed` | terminal   | Cleanup or finalization failed and retained proof remains for investigation or stale janitor handling. | Latest lifecycle event has `to_state=cleanup_failed`; lease `cleanup_state=failed` when a lease exists. |

#### Events

| Event                | Definition                                                                 |
| -------------------- | -------------------------------------------------------------------------- |
| `start_services`     | Begin owned suite startup or attach-mode suite validation.                 |
| `readiness_passed`   | All readiness predicates required by Section 11.4 passed before deadline.  |
| `startup_failed`     | Startup, attach validation, preflight, fixture preparation, or readiness failed before child work started. |
| `child_started`      | A child target, child command, or scheduler work unit started under a ready suite. |
| `child_finished`     | An active child target, child command, or scheduler work unit exited and its status was recorded. |
| `interrupt_received` | The wrapper observed cancellation or process interruption while child work was active. |
| `cleanup_started`    | Teardown, cleanup, or attach-mode diagnostic finalization started.         |
| `cleanup_succeeded`  | Teardown, cleanup, or attach-mode diagnostic finalization completed.       |
| `cleanup_failed`     | Teardown, cleanup, or attach-mode diagnostic finalization failed.          |

#### Transition Rules

| From state      | Event                | Guard                                      | To state         | Required actions                                                              | Failure mapping                                                                 | Observable evidence |
| --------------- | -------------------- | ------------------------------------------ | ---------------- | ----------------------------------------------------------------------------- | ------------------------------------------------------------------------------- | ------------------- |
| `requested`     | `start_services`     | Configuration resolved; `suite_id` allocated; no prior lifecycle event exists. | `starting`       | Create suite directory as needed; write initial lease or attach diagnostic when applicable; append lifecycle event. | If setup cannot begin, fail before mutation with `configuration_error`, `preflight_error`, or `fixture_error` according to the failed predicate. | Lifecycle event `seq=1`; lease or diagnostic artifact refs when present. |
| `starting`      | `readiness_passed`   | All required readiness predicates passed.  | `ready`          | Append lifecycle event and retain readiness diagnostics when produced.         | none                                                                            | Lifecycle event with readiness artifact refs. |
| `starting`      | `startup_failed`     | Any startup, attach validation, preflight, fixture, or readiness predicate failed before child start. | `failed_start`   | Append lifecycle event; record failure summary; terminate known partial resources or leave proof for stale janitor. | `preflight_error`, `service_start_error`, `service_readiness_timeout`, `fixture_error`, or `configuration_error` according to Sections 9 and 11.4. | Lifecycle event with failure fields and proof artifact refs. |
| `ready`         | `child_started`      | Child key is non-empty and not already active. | `running_child`  | Append lifecycle event; set `active_child_count=1`; retain child log or target artifact refs when known. | Child start failure before process launch is `child_target_failure` or `fixture_error` according to the wrapper boundary. | Lifecycle event with child key, `active_child_count=1`, and child artifact refs when known. |
| `running_child` | `child_started`      | Child key is non-empty and not already active. | `running_child`  | Append lifecycle event; increment `active_child_count`; retain child log or target artifact refs when known. | Duplicate or missing child key is an illegal transition. | Lifecycle event with child key and incremented `active_child_count`. |
| `running_child` | `child_finished`     | Active child key is known, child status has been recorded, at least two children remain active before this event, and no interruption wins by guard precedence. | `running_child` | Append lifecycle event; decrement `active_child_count`; retain child status and artifacts. | Unknown child key or negative active count is an illegal transition. Child failure is recorded for primary failure selection by Section 9.1; the lifecycle state itself is not terminal. | Lifecycle event with child key, child status artifact refs, and decremented `active_child_count`. |
| `running_child` | `child_finished`     | Active child key is known, child status has been recorded, exactly one child remains active before this event, and no interruption wins by guard precedence. | `ready` | Append lifecycle event; set `active_child_count=0`; retain child status and artifacts. | Unknown child key or negative active count is an illegal transition. Child failure is recorded for primary failure selection by Section 9.1; the lifecycle state itself is not terminal. | Lifecycle event with child key, child status artifact refs, and `active_child_count=0`. |
| `running_child` | `interrupt_received` | Cancellation or signal wins by guard precedence. | `interrupted`    | Append lifecycle event and preserve child/interruption diagnostics when available; report current `active_child_count`. | `failure_class=interrupted`, `failure_reason=cancelled_or_interrupted`.         | Lifecycle event with interruption fields and `active_child_count`. |
| `ready`         | `cleanup_started`    | No child is running and cleanup/finalization is required. | `cleaning`       | Set lease `cleanup_state=in_progress` when a lease exists; append lifecycle event. | Cleanup start failure is recorded as `cleanup_error` without deleting unproven resources. | Lifecycle event and updated lease. |
| `interrupted`   | `cleanup_started`    | Interruption has been recorded and cleanup/finalization is required. | `cleaning`       | Set lease `cleanup_state=in_progress` when a lease exists; append lifecycle event. | Primary interruption failure is preserved by Section 9.1.                       | Lifecycle event and updated lease. |
| `cleaning`      | `cleanup_succeeded`  | Required cleanup or finalization completed. | `cleaned`        | Set lease `cleanup_state=completed` or `deferred`; append lifecycle event.     | Success unless an earlier primary failure exists.                               | Lifecycle event and final lease. |
| `cleaning`      | `cleanup_failed`     | Cleanup or finalization failed.            | `cleanup_failed` | Set lease `cleanup_state=failed`; append lifecycle event; retain proof for stale janitor. | `cleanup_error`; Section 9.1 decides whether it becomes the public exit-code driver. | Lifecycle event, final lease, and cleanup diagnostics. |

Any listed event presented in an unlisted state MUST append a lifecycle event with `transition_status=illegal`, `from_state` equal to `to_state`, `failure_class=harness`, and `failure_reason=scheduler_accounting_error`, then fail without mutating suite state. An unrecognized event token MUST be rejected before lifecycle mutation and MUST NOT be appended to the schema-valid lifecycle event stream. A terminal state MUST reject every later event as illegal.

Lifecycle failure events MUST carry normalized Section 9 failure fields. `startup_failed`, `interrupt_received`, `cleanup_failed`, and illegal-transition lifecycle events MUST set non-null `failure_class` and `failure_reason` in the lifecycle stream. Non-failure lifecycle events MUST preserve `failure_class=null` and `failure_reason=null` unless a later schema revision adds a narrower event-specific diagnostic field.

**TH-HARNESS-REQ-406**
The service-suite lifecycle active-child counter is normative. `child_started`, `child_finished`, and `interrupt_received` lifecycle events MUST include `active_child_count`. `ready + child_started` sets the count to `1`; `running_child + child_started` increments it; `running_child + child_finished` decrements it and remains in `running_child` while the count is greater than `0`; `running_child + child_finished` transitions to `ready` when the count becomes `0`. Negative active counts, missing child identity, duplicate `child_started` for the same active child key, and `child_finished` for an unknown active child key are illegal transitions under Section 11.2.
Verified by: TH-HARNESS-AC-017, TH-HARNESS-AC-033

**TH-HARNESS-REQ-411**
Every current-run service suite MUST retain `_shared/test-services/<suite-id>/service-scope.json` as `cartulary.test_services.scope.v2` after the suite artifact directory exists and before scheduler classification consumes service diagnostics. The summary MUST contain suite identity, target, readiness generation, preflight, cleanup, started-service and fixture aggregate counts, complete failure counts by normalized class and reason, no more than ten deterministic normalized exemplars per failure class, and the ten slowest fixture activities. Fixture totals and the distinct strategy-aggregate count MUST remain exact. A strategy aggregate is identified by the exact ASCII tuple of service, target, operation, preparation strategy, fixture policy, fixture class, and reuse scope; caller-package and test identities are separate hotspot and slowest-event dimensions and MUST NOT change strategy-aggregate identity. The retained strategy diagnostic contains at most 32 aggregates, selected by descending total duration, then descending event count, then that ASCII aggregate identity. It MUST NOT retain exhaustive database, bucket, package, or test-name inventories. When a service suite fails, `failure` MUST include normalized non-null `failure_class` and `failure_reason` fields. Startup preflight failures map to `infra/preflight_error`; service startup failures map to `infra/service_start_error`; readiness deadline failures map to `infra/service_readiness_timeout`; fixture preparation failures map to `harness/fixture_error` unless Section 9 assigns `artifact_error`; cleanup failures map to `harness/cleanup_error`.

Each producer MUST append bounded `cartulary.test_services.journal_event.v1`
records to one owner-only producer journal. Producer identity is stable for the
process and `seq` starts at one and increases without gaps. Collation sorts by
event time, producer identity, and sequence. A reader MAY ignore only an
incomplete trailing crash record; it MUST reject a malformed completed record,
duplicate sequence, or gap. Every scope refresh MUST write and sync an
owner-only sibling, atomically replace the prior summary, and sync the
containing directory, so readers observe one complete old or new document.

Exact owned-resource state belongs only in a mode-`0600`
`cartulary.test_services.resource_ledger.v1`. Cleanup and stale recovery MUST
validate ownership proof from that ledger before mutation. Successful cleanup
deletes the ledger. Cleanup failure MAY retain only a redacted contained copy
needed for recovery. Browser admission MUST consume a compact
`cartulary.test_services.browser_admission.v1` proof and MUST NOT copy the scope
summary; the proof closes suite/session identity, readiness generation,
required services, container proof, source digest, and the admitted scope
digest.

The first executable browser group or reset in an affinity chain MUST acquire
and hold the browser-stack lease while it consumes that admission. The graph
MUST NOT interpose a commandless readiness unit that releases the allocation to
zero references before the first consumer starts. Later reset and group units
MAY retain the same healthy affinity allocation, and the terminal consumer MUST
release it. Successful browser-session startup MUST bind backend and frontend
readiness in the top-level lifecycle, atomically publish both the immutable
stack and compact admission, and verify that both are owner-only regular files
before returning the allocation. Missing readiness identity or either missing
publication is a startup failure; publication MUST NOT return success with a
partial evidence set. Every admission, lease, stack, and final verification
command MUST propagate failure explicitly even when timing instrumentation has
disabled shell fail-fast behavior. A browser-session artifact directory is a
single-use identity: startup MUST reject a pre-existing directory and MUST NOT
clear, replace, or retry publication under that identity.

Verified by: TH-HARNESS-AC-007, TH-HARNESS-AC-014, TH-HARNESS-AC-017, TH-HARNESS-AC-025

**TH-HARNESS-REQ-410**
An authoritative service-backed Go row MUST make its fixture capability and
complete service dependency set explicit in the catalog and at the helper call
site. The scheduler MUST pass the canonical sorted set through
`CARTULARY_HARNESS_SERVICE_DEPENDENCIES`. Before `appsupport`, `pgtest`, or
`s3test` attaches to or creates any resource, a shared acquisition guard MUST
prove every helper-required service is assigned to the work unit. A missing,
unknown, duplicate, unsorted, or runtime-incompatible declaration fails as
`harness/fixture_error` before either service is touched. A generic helper MUST
NOT silently fall back to a dedicated database, transaction,
migration database, object-store namespace, managed process, browser stack, or
implicit scheduler claim.

A direct developer `go test` invocation is noncanonical. When the assignment
environment variable is absent, it MAY proceed only because the helper call
supplies its explicit required-service list; this does not synthesize a graph
claim, canonical evidence, or compatibility fallback. Non-row
implementation-support tests follow the same explicit call-site rule.

Verified by: TH-HARNESS-AC-007, TH-HARNESS-AC-056

### 11.3 Lease Fields

Lease files MUST be written before child work starts, MUST be redacted before retention, and MUST be written atomically as a complete JSON file. A lease is evidence for cleanup only when its resource proof matches the actual resource state; cleanup MUST verify labels, prefixes, generated names, or equivalent proof and MUST NOT trust the lease path alone.

| Field              | Type                                                               |                        Required |
| ------------------ | ------------------------------------------------------------------ | ------------------------------: |
| `schema_id`        | string, `cartulary.test_services.lease.v1`                         |                             yes |
| `lease_id`         | non-empty opaque lease identifier                                  |                             yes |
| `suite_id`         | 24 lowercase hex chars                                             |                             yes |
| `target`           | string                                                             |                             yes |
| `mode`             | `owned` or `attach`                                                |                             yes |
| `ownership_mode`   | `owned` or `attach`                                                |                             yes |
| `result_root`      | normalized path                                                    |                             yes |
| `run_id`           | normalized run ID                                                  |                             yes |
| `run_root`         | normalized run-root path                                           |                             yes |
| `owner_pid`        | integer process ID for the owning wrapper                          |                             yes |
| `created_at`       | RFC3339 UTC                                                        |                             yes |
| `heartbeat_at`     | RFC3339 UTC                                                        |                              no |
| `expires_at`       | RFC3339 UTC                                                        |                              no |
| `resources[]`      | redacted resource records with service kind, logical ID, and proof | yes in owned mode, may be empty |
| `proof_labels`     | object of required labels used to prove container ownership        |           yes for container use |
| `proof_prefixes`   | object of generated DB/bucket/path prefixes used to prove ownership | yes for DB, bucket, or path use |
| `cleanup_state`    | `not_started`, `in_progress`, `completed`, `failed`, or `deferred` |                             yes |

### 11.4 Readiness Deadlines

Browser owned-stack readiness is an ownership predicate, not only an HTTP availability predicate. The backend is ready only when all of the following hold before the deadline:

- the wrapper-started backend process group is still alive;
- the selected backend port is owned by a process in that process group, when the platform exposes listener process metadata;
- the token-protected `GET /api/v1/test/runtime/identity` route returns `cartulary.test.runtime_identity.v1`, `runtime_marker="harness-owned"`, `test_routes_enabled=true`, and a server process ID;
- the backend process group remains alive after the identity probe.

The canonical browser E2E frontend startup mode is a built preview: `build-web` MUST complete before browser stack startup, `apps/web/dist/index.html` MUST exist before the frontend process starts, and the wrapper MUST launch a non-watching preview command rather than the Vite dev server. Missing built frontend artifacts are `configuration_error`, exit `2`; they are not service-readiness failures.

The frontend is ready only when the wrapper-started preview process group is still alive, the selected frontend port is owned by a process in that process group when listener process metadata is available, the frontend HTTP probe succeeds, the process group remains alive after the probe, and stack metadata records `frontend_mode="preview"` and `frontend_command_kind="vite-preview"`. A stale or unrelated listener MUST NOT satisfy browser readiness.

A dynamically allocated browser backend or frontend port lease MUST remain exclusive for the full owned-session lifetime. Before the transient startup controller exits, it MUST transfer each lease to the corresponding live owned process group; the controller exiting MUST NOT make the lease stale while that process group remains alive. A stop controller MAY adopt the exact session's leases, but MUST release them only after it has stopped the recorded process groups and proved the listeners are gone. Allocation MUST reject a lease whose recorded owner is still alive even when the allocating shell that originally created the lease has exited. This ownership transfer and release policy applies across all same-worktree browser schedulers and run roots.

Object-store readiness MUST prove the smallest mutation lifecycle required by
object-backed fixtures, not only port, authentication, or list availability.
The broker MUST create a uniquely named, run-owned probe namespace and perform
one bounded small-object `put`, `head` with exact size verification, `delete`,
and not-found verification sequence. The probe object and namespace MUST be
cleaned on success, failure, cancellation, and deadline expiry. Cleanup failure
is retained as secondary cleanup evidence and MUST NOT replace an earlier
readiness failure.

Authentication, authorization, or required-capability rejection is terminal.
Transient availability failures MAY continue polling only within the same
120-second object-store readiness window. For an owned lane, that window begins
only after container startup and mapped-endpoint resolution have succeeded; it
is separate from every bounded pre-readiness startup attempt. Once polling
starts, expiry MUST NOT start a replacement container or lane. Owned and
attached services MUST use the same probe contract. Attach mode MAY mutate only
its proven run-owned probe namespace and MUST NOT close, delete, or clean
unrelated borrowed resources. Readiness diagnostics disclose only the bounded
normalized fields authorized by TH-HARNESS-REQ-414; credentials and unrelated
object, endpoint, or container identities are forbidden.
Deadline expiry at any mutation stage MUST emit the bounded Go setup-failure
envelope as `infra/service_readiness_timeout` with normalized stage, attempt
count, and cleanup outcome, and MUST normalize to public exit code `3` in row,
unit, target, and run evidence. Capability rejection is terminal
`harness/fixture_error`; cancellation is
`interrupted/cancelled_or_interrupted`. Cleanup remains secondary evidence and
MUST NOT replace the primary readiness class or reason.

To prevent readiness admission itself from overwhelming one brokered service,
the broker MAY retain one run-owned parent probe bucket for the suite lifetime
and export its exact identity to attached child processes. Each attached process
MUST use a unique run-owned prefix inside that bucket, clean its probe object on
every outcome, and MUST NOT create or delete the shared parent bucket. The
broker MUST prove the parent bucket's mutation path before child work and clean
the bucket during owned teardown. An attached process without this broker proof
MUST create and clean its own unique probe namespace instead.

Before returning a newly created or recycled package bucket to a fixture
consumer, the broker MUST run the same `put`, `head` with exact size
verification, `delete`, and not-found verification inside a unique run-owned
probe prefix in that bucket. Product object-store operations remain fail-fast;
fixture admission MUST NOT introduce product-operation retry behavior.
Package-bucket admission MUST expose an error-returning core so timeout,
capability rejection, cancellation, cleanup failure, concurrent distinct
packages, and reuse can be verified without parsing `testing.TB` fatal text.
The `testing.TB` wrapper is responsible only for emitting the validated setup
failure envelope and registering successful lock release.

Scheduler saturation intervals, blocking-resource holders, fixture lifecycle
events, and typed readiness evidence MAY support a contention diagnosis, but a
PUT-path timeout does not by itself prove a provider defect or host-level cause.
A recurrence after complete service claims are enforced requires a separate
provider or host investigation; it MUST NOT authorize timeout expansion,
additional retries, provider-specific behavior, or disclosure of raw transport
errors.

Owned-stack supervision MUST test liveness of the complete process group, not
only the original group-leader PID. A short-lived launcher exiting while its
descendant reporter remains in the group MUST NOT be interpreted as child
completion, success, or permission to tear down the stack. The controller that
started the group MUST collect its exit status with `wait` in that same shell;
command-substitution or another subshell MUST NOT perform the wait. A browser
selector invocation is terminal only after the entire child group has exited
and its declared report is complete.

| Resource                     | Deadline | Poll interval | Failure reason                       |
| ---------------------------- | -------: | ------------: | ------------------------------------ |
| Docker preflight             |    `15s` |          `1s` | `preflight_error`                    |
| Postgres container readiness |   `180s` |       `500ms` | `service_readiness_timeout`          |
| local object-store readiness |   `120s` |       `500ms` | `service_readiness_timeout`          |
| Template DB migration        |   `180s` |           n/a | `fixture_error`                      |
| Browser backend readiness    |   `120s` |       `500ms` | `service_readiness_timeout`          |
| Browser frontend readiness   |   `120s` |       `500ms` | `service_readiness_timeout`          |
| Destructive database reset   |    `30s` |           n/a | `fixture_error` or `timeout_failure` |
| Replacement backend readiness | `120s` |       `500ms` | `service_readiness_timeout` or `timeout_failure` |

### 11.5 Retry and Teardown Rules

**TH-HARNESS-REQ-401**
No hidden startup retry is allowed. Retry is allowed only when a resource row declares `max_attempts`, bounded backoff, retryable failure reasons, and an overall deadline. Readiness polling within a deadline is not a retry.
Verified by: TH-HARNESS-AC-007

**TH-HARNESS-REQ-402**
Owned teardown order MUST be: browser child processes, browser fixtures, reset-tainted runtime roots, test databases, object buckets or prefixes, service containers, lease finalization. Attach mode MUST record diagnostics but MUST NOT delete container-level resources or external services.
Verified by: TH-HARNESS-AC-007, TH-HARNESS-AC-010

Destructive reset, cleanup, attach-mode service mutation, and non-idempotent operations MUST NOT be retried unless a resource row explicitly declares the operation safe to retry.

| Resource operation          | `max_attempts` | Backoff | Retryable failure reasons                                  | Overall deadline | Safe retry scope                                  |
| --------------------------- | -------------: | ------- | ---------------------------------------------------------- | ---------------- | ------------------------------------------------- |
| Docker preflight            |            `1` | none    | none                                                       | `15s`            | none                                              |
| Postgres owned startup      |            `3` | `500ms` | transient Docker startup or transport failure before readiness polling begins | attempt startup only; readiness deadline is Section 11.4 `180s` | Failed attempt container is terminated first.     |
| object-store owned startup  |            `2` | `250ms` | transient Docker startup or transport failure before readiness polling begins | attempt startup only; readiness deadline is Section 11.4 `120s` | Failed attempt container is terminated first.     |
| Template DB migration       |            `1` | none    | none                                                       | `180s`           | none                                              |
| Browser backend startup     |            `1` | none    | none                                                       | `120s`           | readiness polling only                            |
| Browser frontend startup    |            `1` | none    | none                                                       | `120s`            | strict-port conflicts fail as `resource_conflict` |
| Browser reset lifecycle     |            `1` | none    | none                                                       | derived from drain, `30s` database reset, `120s` backend readiness, and bounded evidence publication | no stage or product retry |
| Owned teardown and cleanup  |            `1` | none    | none                                                       | cleanup-specific | cleanup records failure and leaves proof for janitor |

Stale janitor cleanup of previously owned service containers is a proof-gated startup preflight maintenance step, not authoritative product evidence. Once ownership proof and current-suite exclusion pass, Docker `not found` and Docker "removal already in progress" results MUST be accepted as idempotent cleanup outcomes and MUST NOT fail the new suite. Concurrent removal MUST be retained as deferred cleanup diagnostics. Docker daemon/list failures, unsafe ownership proof, and non-idempotent removal failures remain blocking startup-preflight failures.

**TH-HARNESS-REQ-412**
If Docker container deletion for a proven stale owned service container returns `context deadline exceeded`, cancellation, or an equivalent bounded remove timeout, startup preflight MUST perform one bounded post-delete recheck before deciding whether to fail. The recheck MUST use Docker container state for the same container ID after the original ownership proof and current-suite exclusion have already passed. If the recheck proves the container is gone, the outcome is idempotent removal and counts as removed. If the recheck proves the container is in Docker `removing` or `dead` state, the outcome is deferred idempotent cleanup and counts as deferred. If the recheck cannot read Docker state, or proves the same container is still present without `removing` or `dead` state, the original deletion timeout remains a blocking startup-preflight failure with `failure_class=infra` and `failure_reason=preflight_error`.

This timeout acceptance is proof-gated only. A stale-container delete timeout without a successful `not found`, `removing`, or `dead` recheck MUST NOT be accepted.
Verified by: TH-HARNESS-AC-007, TH-HARNESS-AC-010, TH-HARNESS-AC-017

**TH-HARNESS-REQ-414**
An object-store readiness failure MUST retain one bounded
`cartulary.object_store_readiness_diagnostic.v1` value under
`cartulary.test_services.scope.v2`
`extensions["cartulary.object_store_readiness"]`. The diagnostic MUST contain
the probe phase, stage, normalized outcome, attempt and per-attempt timeout
counts, elapsed readiness-window milliseconds, cleanup outcome, and bounded
counts for the closed normalized cause set. An owned lane MAY additionally
retain normalized container state, health state, exit code, and OOM status.
The diagnostic MUST NOT contain endpoints, credentials, bucket, object, or
container identities, raw errors, raw queries, or log tails. The extension is
supplemental diagnosis only; the declared service-scope `failure` fields remain
authoritative for classification.

The service-suite helper MUST atomically publish a private mode-`0600`
`cartulary.test_services.start_result.v1` control record before it exits. The
record contains only ready-or-failed status, run, target, and suite identity,
the run-relative service-scope reference, and normalized nullable failure class
and reason. Its broker consumer MUST prove run-root containment, identity
agreement, schema validity, and agreement with the referenced current-run
service scope before propagating the service failure into row, unit, target,
and run evidence. Missing, malformed, mismatched, or escaping control evidence
is `artifact/artifact_error`; a valid readiness deadline remains
`infra/service_readiness_timeout` and public exit `3` at every evidence layer.
Unknown fixture-provider failures without valid service evidence are
`harness/fixture_error`.
Verified by: TH-HARNESS-AC-007, TH-HARNESS-AC-013, TH-HARNESS-AC-017, TH-HARNESS-AC-025

Attach mode MAY write diagnostic records and lease observations. It MUST NOT delete externally supplied services, containers, databases, buckets, or object prefixes.

For browser owned stacks, a listener conflict detected before process startup or during backend/frontend process bind/startup maps to `resource_conflict`. Backend or frontend process exit before readiness maps to `service_start_error` only when retained startup diagnostics and logs do not identify a listener, port, lock, or other resource conflict. A live owned process that does not satisfy its readiness predicates before the deadline maps to `service_readiness_timeout`. Suite-admin login failures after owned readiness has been proven are no longer treated as readiness failures.

Startup retry windows and readiness deadlines are separate. If a Postgres or
object-store startup attempt reaches readiness polling and then the Section
11.4 readiness deadline expires, the operation MUST NOT retry that service or
start a replacement lane. The failure MUST be `failure_class=infra`,
`failure_reason=service_readiness_timeout`, and public exit `3`. A transient
object-store probe failure MAY cause another probe attempt only inside that
same readiness window and against that same service lane. Browser backend
startup, browser frontend startup, runtime reset, and cleanup have
`max_attempts=1`; their polling or operation deadlines do not create retry
attempts. Browser ports are dynamically allocated before process startup, but
a later strict-port collision is terminal `failure_class=infra`,
`failure_reason=resource_conflict`; silently changing the admitted port or
replacing terminal startup evidence is forbidden.

### 11.6 Duration Baselines

Duration baselines are advisory scheduler planning data only. They MUST NOT become benchmark claims, product performance conformance, timeout policy, or evidence that product behavior is fast enough.

Every work unit MUST carry a positive authored `estimated_work_ms`. Catalog row
units derive that value from the exact evidence-class estimate in the authored
work-graph owner; direct units use their exact target policy; browser group
units sum their selected row estimates; compatible Go units sum their exact
symbols before capacity-derived LPT distribution. Estimates affect rank and
packing only. They MUST NOT change selection, timeout, failure, cache, fixture,
or evidence semantics.

Canonical observations MAY inform a later authored estimate change only when
they come from successful, uncontaminated retained runs whose unit IDs and
semantic digests match the current source state. Concurrent intervals are
unioned, not summed. A missing estimate, stale identity, aggregate target
envelope copied into multiple units, or fixture interval attributed as command
work is invalid planning input and MUST fail graph compilation rather than
silently selecting a legacy default.

The retained public names `go-test-duration-baselines`,
`browser-e2e-duration-baselines`,
`service-backed-make-target-duration-baselines`, and
`harness-smoke-duration-baselines` are read-only canonical observation views;
their matching drift commands provide the same view with the public input
defaults declared by Section 5.3. They do not write private subject baselines.
`harness-public-target-duration-baselines` is the sole baseline writer and
publishes only the complete accepted v3 portfolio under TH-HARNESS-REQ-380.
`agent-finalize` MUST NOT mutate planning estimates or performance baselines.
Until the first accepted v3 portfolio is published,
`go-test-duration-baseline-coverage` validates the exact authored measurement
roster and reports `pending-publication` without reading or translating the
archival v2 baseline. After publication it additionally requires exact v3
baseline roster closure.

Warm graph health checks MAY consume canonical retained evidence from a
successful warm-ready run. Such checks remain harness-maintenance evidence and
MUST NOT be described as claim-bearing product benchmark evidence. They MUST
reject hidden provisioning, measurement work admitted outside its quiet
profile, unexplained queue or resource intervals, missing holders, fixture
contamination, or incomplete event/target projection closure. Performance
regression is decided only by the matched windows and variability rules in
TH-HARNESS-REQ-377 through TH-HARNESS-REQ-380; this diagnostic has no separate
fixed wall-time or peer-balance threshold.

### 11.7 Network Flow Fixture Materialization Lifecycle

**TH-HARNESS-REQ-407**
Network Flow fixture source roots are committed fixture inputs, not service-owned mutable state. A harness runner MUST NOT write into `fixtures/network-flow/**` during validation, preview, apply, graph, cursor, indicator-link, or transcript comparison work. Any generated, staged, normalized, or copied fixture material MUST live under the current run root and MUST be removed only by ordinary result-root cleanup. `make clean`, service teardown, stale janitors, database reset, and object-store reset MUST NOT delete committed Network Flow fixture roots.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-408**
Before a Network Flow fixture is materialized, the runner MUST validate the manifest schema, path safety, per-file byte hashes, aggregate bundle hashes, frozen status when selected for conformance, and owner routing. Materialization MUST use only manifest-listed files, MUST reject symlinks and traversal paths, and MUST make product execution observe a read-only run-local copy. A failed pre-materialization check is `failure_class=artifact`, `failure_reason=artifact_error`; product code MUST NOT start for that fixture.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-409**
Network Flow fixture materialization participates in the existing service lifecycle only as fixture preparation. If materialization fails before child work starts, the service-suite lifecycle records `startup_failed` with `fixture_error` or `artifact_error` according to Section 9 ownership, preserves diagnostics, and performs ordinary owned teardown. If child product work has started, later comparison failures are product or artifact failures according to the owning assertion, but the committed fixture root remains immutable.
Verified by: TH-HARNESS-AC-049

### 11.8 Local Development Object-Store Proxy

**TH-HARNESS-REQ-413**
The local development object-store proxy is development-only and MAY retain
`127.0.0.1:8333`; browser evidence MUST never depend on it. Proxy lifecycle
operations MUST serialize through one OS advisory operation lock. State,
operation metadata, startup attempts, ready leases, and per-instance logs MUST
reside beneath one owner-only repo runtime root as non-symlink regular files.
State publication uses an owner-only temporary file, file `fsync`, atomic
rename, and parent-directory `fsync`.

The launcher MUST publish a secure
`cartulary.local_object_store_proxy_start_attempt.v1` before spawn. The child
MUST synchronously bind the exact loopback listener, then publish its instance
identity and full process proof before serving normal proxy traffic. The proof
contains Linux boot ID, PID, `/proc/<pid>/stat` start-time ticks, effective UID,
executable device and inode, and SHA-256 of `/proc/<pid>/exe`. Promotion to
`cartulary.local_object_store_proxy_lease.v1` is atomic and permitted only after
the five-second bind/identity handshake, closed
`cartulary.local_object_store_proxy_health.v1` identity, and the separately
bounded object-store plus exact-CORS probe all pass.

Recovery or reuse requires matching process, executable, health, listener,
instance, and canonical nonsecret configuration proofs. A fully proven
configuration mismatch is gracefully restarted. A stale or legacy PID file is
untrusted metadata and MUST never authorize signaling. An unproven listener is
an immediate `resource_conflict`. Signaling requires `pidfd_open`, complete
proof revalidation after opening the pidfd, `pidfd_send_signal`, and confirmed
termination through the pidfd. Unsupported pidfd behavior fails closed without
PID-only fallback. A startup attempt abandoned before process proof may be
discarded only when no listener occupies the configured endpoint.

The upstream origin is canonicalized and MUST reject userinfo, query, and
fragment; listener configuration MUST be an explicit loopback IP. Configuration
fingerprints include only canonical nonsecret values. Development proxy health
is loopback-only implementation support and MUST NOT become a product API,
production deployment surface, browser attachment input, or product evidence.
Verified by: TH-HARNESS-AC-007, TH-HARNESS-AC-010

## 12. Test-Only Harness Routes

**TH-HARNESS-REQ-450**
Test-only harness routes are harness routes. Runtime-control routes include `GET /api/v1/test/runtime/identity`, `POST /api/v1/test/clock/set`, `POST /api/v1/test/clock/reset`, `GET /api/v1/test/clock/state`, `POST /api/v1/test/runtime/public-error-faults`, `POST /api/v1/test/runtime/network-flow-faults`, `POST /api/v1/test/runtime/network-flow-randomness`, `POST /api/v1/test/runtime/network-flow-auth-transitions`, and `POST /api/v1/test/runtime/network-flow-audit-assertions`. Fixture routes include `POST /api/v1/test/incidents/{incident_id}/saved-views/system`. Any future `/api/v1/test/*` or `/ws/v1/test/*` route that observes or mutates harness runtime state or fixture state is also a test-only harness route. These routes MUST be unavailable unless every enablement predicate below is satisfied. They MUST NOT be documented as production API behavior.

Runtime-linked implementations of Section 12 routes that are registered by application binaries are owned by the platform harness runtime adapter boundary. Private implementation package paths are not public route contracts. Non-test runtime code MUST NOT import broad test-helper packages to register or execute Section 12 routes once a platform harness runtime adapter exists for that route family.
Verified by: TH-HARNESS-AC-008, TH-HARNESS-AC-013

### 12.1 Enablement

**TH-HARNESS-REQ-480**
An in-process application test server MUST receive an explicit test-route
mode. The closed modes are disabled, harness-owned, and custom environment.
An omitted, empty, or unknown mode MUST fail before runtime construction.
Harness-owned mode is admissible only when the test exercises a guarded
test-route contract or registers an owner test-route contribution; ordinary
product-route tests MUST use disabled mode. Custom-environment mode is limited
to negative configuration, authorization, host, origin, token, and process
composition tests that supply the complete environment under test.

This requirement changes private test-helper setup only. It does not change
the production enablement predicates below.

Verified by: TH-HARNESS-AC-008, TH-HARNESS-AC-056

| Predicate                      | Required value                                                                      |
| ------------------------------ | ----------------------------------------------------------------------------------- |
| `CARTULARY_ENABLE_TEST_ROUTES` | Exact `1`.                                                                          |
| `CARTULARY_TEST_ROUTE_TOKEN`   | Non-empty string with at least 128 bits of entropy, generated by the harness stack. |
| Runtime ownership              | Server started by a Make-owned browser or test harness stack.                       |
| Production/default runtime     | Test-only harness routes are not registered.                                         |

**TH-HARNESS-REQ-453**
The test-route authorization header name is exactly `X-Cartulary-Test-Route-Token`. Harness-generated route tokens MUST be 32 bytes from a cryptographically secure pseudorandom generator encoded as unpadded base64url, producing exactly 43 ASCII characters. Attach-mode supplied tokens MUST be ASCII visible characters, MUST be length `43..512`, MUST contain no whitespace, MUST NOT equal `test`, `token`, `secret`, `password`, or `changeme`, and MUST NOT be a repeated single-character string. Missing, malformed, or weak attach-mode tokens MUST fail before test-route registration with `failure_class=config`, `failure_reason=configuration_error`, and public exit code `2`.
Verified by: TH-HARNESS-AC-008, TH-HARNESS-AC-035

### 12.2 Authorization

| Request condition                                     | Behavior                                                     |
| ----------------------------------------------------- | ------------------------------------------------------------ |
| Test-only harness route not enabled                   | Return ordinary not-found behavior.                          |
| Test-only harness route enabled, missing token header | `403`, `error.code=test_route_forbidden`.                    |
| Test-only harness route enabled, wrong token header   | `403`, `error.code=test_route_forbidden`.                    |
| Test-only harness route enabled, correct token header | Evaluate request after host/origin boundary checks.          |
| Cookie-authenticated request without token            | Forbidden; session auth does not authorize test-only routes. |
| Bearer/session/bootstrap-token request without token  | Forbidden; product auth does not authorize test-only routes. |

CSRF does not apply because cookie authentication is not accepted as authorization for test-only harness routes. Incident roles, session cookies, bearer sessions, bootstrap tokens, and `deployment_admin` authority do not bypass the test-route token requirement.

**TH-HARNESS-REQ-451**
When test routes are enabled and a harness-owned API or browser origin is configured, test-only harness routes MUST reject requests whose request origin is not the harness-declared browser origin or harness-declared API origin, or whose request host does not match the harness-owned API origin. Same-process health and readiness probes MUST use explicitly declared non-destructive health endpoints rather than test-only harness routes. Test-only harness routes MUST NOT enable permissive CORS. A rejected origin or host MUST fail before any runtime-control mutation or fixture mutation with `403`, `error.code=test_route_forbidden`.
Verified by: TH-HARNESS-AC-008

### 12.2.1 Runtime Identity

`GET /api/v1/test/runtime/identity` is the canonical predicate-bearing harness test route. It returns `cartulary.test.runtime_identity.v1`, the harness runtime marker, `test_routes_enabled=true`, and the server process ID. It MUST NOT return the route token or any database, object-store, credential, or session secret. Browser wrappers and Playwright global setup use this route to prove that the selected API origin belongs to the current harness-owned backend before suite-admin login work begins and after backend replacement. There is no runtime reset HTTP route.

### 12.2.2 Test Clock

**TH-HARNESS-REQ-460**
The test clock routes are harness test routes with the same enablement, host/origin, and token authorization predicates as the runtime identity route. The route family consists of `POST /api/v1/test/clock/set`, `POST /api/v1/test/clock/reset`, and `GET /api/v1/test/clock/state`. They MAY control or observe the harness-owned runtime clock for scenarios that need deterministic authentication, cursor TTL, timestamp, timezone, source-retention, or session-expiry timing. Because that clock can feed security-sensitive product decisions in harness-owned runtimes, missing token, wrong token, product session credentials alone, wrong host, missing origin when origins are configured, malformed origin, or unapproved origin MUST fail before clock mutation or state disclosure with `403`, `error.code=test_route_forbidden`.
Verified by: TH-HARNESS-AC-035, TH-HARNESS-AC-051

**TH-HARNESS-REQ-461**
`POST /api/v1/test/clock/set` accepts a JSON object with exactly one of `fixed_now` or `offset_seconds`. `fixed_now` MUST be an RFC3339/RFC3339Nano timestamp and is normalized to UTC. `offset_seconds` MUST set the clock to wall time plus that offset and MUST clear any fixed clock. Unknown members, invalid JSON, non-object JSON, trailing JSON, missing both command fields, or specifying both command fields MUST fail with `400`, `error.code=invalid_mutation_payload`.
Verified by: TH-HARNESS-AC-035, TH-HARNESS-AC-051

**TH-HARNESS-REQ-462**
`POST /api/v1/test/clock/reset` accepts no body or `{}` and restores wall-clock mode by clearing both fixed time and offset. Unknown members, invalid JSON, non-object JSON, or trailing JSON MUST fail with `400`, `error.code=invalid_mutation_payload`. `GET /api/v1/test/clock/state` MUST return the current test-clock state without mutation.
Verified by: TH-HARNESS-AC-051

**TH-HARNESS-REQ-463**
Successful test-clock set, reset, and state responses MUST return `cartulary.test.clock_control.v1` in the standard success envelope. The response MUST include `schema_id`, `mode`, `now`, and `offset_seconds`; `mode` is one of `wall`, `offset`, or `fixed`; `now` is the current harness clock in RFC3339Nano UTC form; `fixed_now` appears only when `mode="fixed"`. The response MUST NOT include the test-route token, cookies, product session credentials, database credentials, object-store credentials, host filesystem paths, or private runtime state.
Verified by: TH-HARNESS-AC-051

**TH-HARNESS-REQ-464**
Browser reset MUST replace the backend process, so registered in-memory test-clock state cannot survive into the replacement backend; its initial state is `mode="wall"` and `offset_seconds=0`. A fixture or target that changes the test clock MUST either run in an owned runtime that is torn down afterward or restore wall mode before subsequent unrelated product work starts. Network Flow fixtures that rely on cursor TTL, safe-digest rotation windows, source-retention expiry, soft-delete timing, timezone fold/gap interpretation, uptime-derived timestamps, or timestamp ordinal boundaries MUST record the selected clock-control response in their transcript and MUST cite the adopted product owner requirement that defines the expected time behavior.
Verified by: TH-HARNESS-AC-051

### 12.2.3 Saved-View System Fixture

`POST /api/v1/test/incidents/{incident_id}/saved-views/system` is a harness fixture route with the same enablement, host/origin, and token authorization predicates as the runtime identity route. The route MAY seed one incident-bound `scope='system'` saved-view fixture per successful request for browser scenarios that must distinguish implementation-owned saved-view configurations from contract-backed system views. It MUST NOT be exposed as production API behavior, MUST NOT accept ordinary session, role, CSRF, bearer, bootstrap-token, or `deployment_admin` authorization as a substitute for the test-route token, and MUST fail host/origin or token checks before decoding the body or creating any saved-view row.

This fixture route MUST create only saved-view fixture rows through the saved-view store path. It MUST NOT expose arbitrary SQL execution, projection mutation, generic fixture mutation, or caller-supplied saved-view identity, owner, timestamps, version, or scope. The route fixes `scope='system'`, fixes `owner_user_id=null`, derives `incident_id` from the path, and returns the normal saved-view resource in the standard success envelope with HTTP `201`.

### 12.2.4 Public-Error Fault Control

`POST /api/v1/test/runtime/public-error-faults` is a harness test route with the same enablement, host/origin, and token authorization predicates as the runtime identity route. The route MAY arm a one-shot public error envelope for the next exact ordinary `/api/v1/` request whose method and path match the request body. It MUST NOT be exposed as production API behavior, MUST NOT be listed in production OpenAPI, MUST NOT accept ordinary session, role, CSRF, bearer, bootstrap-token, or `deployment_admin` authorization as a substitute for the test-route token, and MUST fail host/origin or token checks before decoding the body or arming any fault.

The armed fault is in-memory harness runtime state. It is consumed at the service boundary before the ordinary route handler runs. Matching is exact on uppercase HTTP method and request path; query strings and fragments are not part of the accepted path. Faults MUST apply only to paths that start with `/api/v1/` and MUST NOT apply to paths that start with `/api/v1/test/`. Test-control routes therefore cannot fault themselves or other test-only harness controls.

The request body MUST be a JSON object with exactly the fields below.

| Field          | Required | Behavior                                                                                             |
| -------------- | -------- | ---------------------------------------------------------------------------------------------------- |
| `method`       | yes      | Non-empty HTTP method; normalized to uppercase for exact matching.                                   |
| `path`         | yes      | Exact ordinary public route path beginning `/api/v1/` and not beginning `/api/v1/test/`; no query.   |
| `status`       | yes      | Public error status from `400` through `599`.                                                        |
| `code`         | yes      | Non-empty public error code.                                                                         |
| `message`      | no       | Public error message to place in the public envelope.                                                |
| `retryable`    | no       | Public retryability flag; omitted values default to `false`.                                         |
| `details`      | no       | Public error details object. Consumers MUST render only details keys allowlisted by product UI code. |
| `consume_once` | yes      | Must be `true`; persistent or multi-consume faults are not accepted.                                 |

Unknown members, missing required fields, non-object JSON, invalid JSON, status outside `400..599`, empty `code`, a path outside ordinary `/api/v1/`, a path under `/api/v1/test/`, a path with query or fragment, or `consume_once` other than `true` MUST fail with `400`, `error.code=invalid_public_error_fault_request`.

Successful arming MUST return HTTP `201` with `cartulary.test.public_error_fault.v1` in the standard success envelope. The response MUST include a generated `fault_id`, normalized `method`, exact `path`, `status`, `code`, `retryable`, and `consume_once=true`. The response MUST NOT include the test-route token, configured origins, cookies, product session credentials, database credentials, object-store credentials, or private runtime state.

The next exact ordinary public-route match MUST return a standard public error envelope using the armed `status`, `code`, `message`, `retryable`, and `details`, with the request's public `request_id`. After that response, the fault MUST be consumed and the same request match MUST reach the ordinary route handler unless another fault is armed.

**TH-HARNESS-REQ-454**
At most one public-error fault may be armed per harness-owned runtime. A request to arm a second fault while one is pending MUST fail before replacing the pending fault with HTTP `409`, `error.code=test_public_error_fault_already_armed`. Browser reset MUST replace the backend process, so no pending fault can survive into the replacement backend. A consumed fault MUST be removed before the fault response is written, so a retry of the same ordinary request reaches the ordinary route handler unless another fault has been armed.
Verified by: TH-HARNESS-AC-008, TH-HARNESS-AC-035

### 12.2.5 Network Flow Fault Control

**TH-HARNESS-REQ-455**
`POST /api/v1/test/runtime/network-flow-faults` is a harness test route with the same enablement, host/origin, and token authorization predicates as the runtime identity route. The route MAY arm a one-shot Network Flow commit or worker fault for a named harness boundary consumed by Network Flow tests. It MUST NOT be exposed as production API behavior, MUST NOT be listed in production OpenAPI, MUST NOT accept ordinary session, role, CSRF, bearer, bootstrap-token, or `deployment_admin` authorization as a substitute for the test-route token, and MUST fail host/origin or token checks before decoding the body or arming any fault.
Verified by: TH-HARNESS-AC-050

**TH-HARNESS-REQ-456**
Network Flow fault boundaries are closed tokens owned by the harness. They identify where an opted-in Network Flow test implementation checks for an armed fault; they do not define product state-machine semantics. The supported boundary tokens are exactly:

| Boundary token                                                     | Harness use |
| ------------------------------------------------------------------ | ----------- |
| `network_flow.import.before_owner_prepare`                         | Fault before Network Flow table or diagnostic state is prepared for the shared unit of work. |
| `network_flow.import.after_owner_prepare`                          | Fault after Network Flow owner state has been prepared but before later participants are prepared. |
| `network_flow.import.after_indicator_prepare`                      | Fault after indicator create/dedupe or binding participants have prepared their writes. |
| `network_flow.import.after_audit_prepare`                          | Fault after transactional audit occurrences have been prepared. |
| `network_flow.import.after_idempotency_prepare`                    | Fault after idempotency-success state has been prepared. |
| `network_flow.import.after_terminal_publication_prepare`           | Fault after terminal import-result publication has been prepared. |
| `network_flow.import.before_transaction_commit`                    | Fault immediately before the shared transaction commit. |
| `network_flow.import.after_transaction_commit_before_reply`        | Fault after the shared transaction commits but before the apply caller receives the terminal response. |
| `network_flow.worker.before_handler_start`                         | Fault before a durable Network Flow worker handler starts owner work. |
| `network_flow.worker.before_apply_start`                           | Fault before the worker starts an apply attempt. |
| `network_flow.worker.before_cancellation_check`                    | Fault before the worker observes a cancellation gate. |
| `network_flow.worker.before_final_commit`                          | Fault before the worker reaches the final shared transaction commit. |
| `network_flow.worker.after_final_commit_before_terminal_publication` | Fault after final commit and before terminal-result publication or recovery reconciliation. |
| `network_flow.worker.after_terminal_publication_before_ack`        | Fault after terminal publication and before the worker acknowledges durable completion. |
| `network_flow.worker.before_replay_reconciliation`                 | Fault before worker recovery reconciles an already-committed operation with terminal publication. |

Verified by: TH-HARNESS-AC-050

**TH-HARNESS-REQ-457**
The request body MUST be a JSON object with exactly the fields below.

| Field             | Required | Behavior |
| ----------------- | -------- | -------- |
| `boundary`        | yes      | One of the closed Network Flow boundary tokens in TH-HARNESS-REQ-456. |
| `fault_kind`      | yes      | One of `return_error`, `panic`, `cancel_context`, `worker_crash`, or `worker_cancel`; worker-only kinds are accepted only with `network_flow.worker.*` boundaries. |
| `error_code`      | conditional | Required only for `fault_kind="return_error"`; a lowercase safe diagnostic token matching `^[a-z][a-z0-9_]{1,127}$`. |
| `correlation_key` | no       | Optional ASCII token matching `^[A-Za-z0-9._:-]{1,128}$`; when supplied, consumption requires the same key. |
| `consume_once`    | yes      | Must be `true`; persistent or multi-consume faults are not accepted. |

Unknown members, missing required fields, non-object JSON, invalid JSON, unsupported boundary, unsupported fault kind, worker-only fault kind on an import boundary, invalid or misplaced `error_code`, invalid `correlation_key`, or `consume_once` other than `true` MUST fail with `400`, `error.code=invalid_network_flow_fault_request`. Successful arming MUST return HTTP `201` with `cartulary.test.network_flow_fault_control.v1` in the standard success envelope. The response MUST include a generated `fault_id`, exact `boundary`, exact `fault_kind`, optional `error_code`, optional `correlation_key`, and `consume_once=true`. The response MUST NOT include the test-route token, configured origins, cookies, product session credentials, database credentials, object-store credentials, raw fixture source paths, or private runtime state.
Verified by: TH-HARNESS-AC-050

**TH-HARNESS-REQ-458**
The armed Network Flow fault is in-memory harness runtime state and is consumed only by an opted-in Network Flow test implementation at the exact named boundary. A boundary mismatch MUST leave the fault pending. If the fault has a `correlation_key`, a missing or different consumer key MUST leave the fault pending. A consumed fault MUST be removed before applying its effect so retry, replay, or recovery reaches ordinary behavior unless another fault is armed. If no Network Flow fault registry is registered in the harness-owned runtime, boundary checks MUST be no-ops.
Verified by: TH-HARNESS-AC-050

**TH-HARNESS-REQ-459**
At most one Network Flow fault may be armed per harness-owned runtime. A request to arm a second fault while one is pending MUST fail before replacing the pending fault with HTTP `409`, `error.code=test_network_flow_fault_already_armed`. Browser reset MUST replace the backend process, so no pending Network Flow fault can survive into the replacement backend. Network Flow fault controls MAY be used to prove all-or-nothing final commit, worker crash, cancellation, terminal-publication replay, and recovery behavior only when the executed fixture also cites the adopted product owner requirement that defines the expected state.
Verified by: TH-HARNESS-AC-050

### 12.2.6 Network Flow Deterministic Randomness Control

**TH-HARNESS-REQ-465**
`POST /api/v1/test/runtime/network-flow-randomness` is a harness test route with the same enablement, host/origin, and token authorization predicates as the runtime identity route. The route MAY arm one deterministic random stream for opted-in Network Flow fixture code that needs repeatable IDs, nonces, key IDs, digest salts, or intentional collision values. It MUST NOT be exposed as production API behavior, MUST NOT be listed in production OpenAPI, MUST NOT accept ordinary session, role, CSRF, bearer, bootstrap-token, or `deployment_admin` authorization as a substitute for the test-route token, and MUST fail host/origin or token checks before decoding the body or arming any stream.
Verified by: TH-HARNESS-AC-052

**TH-HARNESS-REQ-466**
Network Flow deterministic-randomness streams are closed harness tokens. They identify fixture-only injection points and do not define product identity semantics, digest algorithms, cursor algorithms, Graph Projection semantics, or public API compatibility. The supported stream tokens are exactly:

| Stream token                             | Harness use |
| ---------------------------------------- | ----------- |
| `network_flow.table_id`                  | Deterministic table identity and table-name collision fixtures. |
| `network_flow.row_id`                    | Deterministic row identity fixtures. |
| `network_flow.diagnostic_id`             | Deterministic diagnostic identity and diagnostic-order fixtures. |
| `network_flow.import_job_id`             | Deterministic import/apply job identity fixtures. |
| `network_flow.import_source_ref`         | Deterministic opaque import-source reference fixtures. |
| `network_flow.cursor_nonce`              | Deterministic cursor nonce, replay, TTL, and rotation fixtures. |
| `network_flow.safe_digest_nonce`         | Deterministic safe-digest salt or key-bound comparison fixtures without exposing production secrets. |
| `network_flow.graph_invocation_id`       | Deterministic ephemeral Graph Projection invocation fixtures. |

Verified by: TH-HARNESS-AC-052

**TH-HARNESS-REQ-467**
The request body MUST be a JSON object with exactly the fields below.

| Field          | Required | Behavior |
| -------------- | -------- | -------- |
| `stream`       | yes      | One of the closed Network Flow stream tokens in TH-HARNESS-REQ-466. |
| `value_kind`   | yes      | One of `uuid`, `token`, or `hex_bytes`. |
| `values`       | yes      | Ordered deterministic values, length `1..256`; duplicate values are allowed only to exercise collision behavior. |
| `consume_once` | yes      | Must be `true`; persistent or multi-consume values are not accepted. |
| `exhaustion`   | yes      | Must be `fail_closed`; an armed stream exhausted by fixture code MUST fail the fixture rather than silently falling back to production randomness. |

For `value_kind="uuid"`, each value MUST be canonical lowercase UUID text. For `value_kind="token"`, each value MUST match `^[A-Za-z0-9._:-]{1,128}$`. For `value_kind="hex_bytes"`, each value MUST be lowercase, even-length hex text no longer than 512 characters.
Verified by: TH-HARNESS-AC-052

**TH-HARNESS-REQ-468**
Unknown members, missing required fields, non-object JSON, invalid JSON, unsupported stream, unsupported value kind, an empty or oversized `values` array, a value that does not match the selected `value_kind`, `consume_once` other than `true`, or `exhaustion` other than `fail_closed` MUST fail with `400`, `error.code=invalid_network_flow_randomness_request`. Successful arming MUST return HTTP `201` with `cartulary.test.network_flow_randomness_control.v1` in the standard success envelope. The response MUST include a generated `control_id`, exact `stream`, exact `value_kind`, `value_count`, `remaining_count`, `consume_once=true`, and `exhaustion="fail_closed"`. The response MUST NOT include deterministic values, the test-route token, configured origins, cookies, product session credentials, database credentials, object-store credentials, production secret material, raw fixture source paths, or private runtime state.
Verified by: TH-HARNESS-AC-052

**TH-HARNESS-REQ-469**
At most one deterministic-randomness sequence may be armed for a stream in a harness-owned runtime. A request to arm a second sequence for the same stream while the stream is registered, including after all values have been consumed, MUST fail before replacing the registered sequence with HTTP `409`, `error.code=test_network_flow_random_stream_already_armed`. Consuming a stream MUST return values in request order exactly once. A missing stream MUST be a no-op for opted-in consumers, but an armed stream that is exhausted or consumed with the wrong `value_kind` MUST fail closed. Browser reset MUST replace the backend process, so registered deterministic-randomness streams cannot survive into the replacement backend.
Verified by: TH-HARNESS-AC-052

### 12.2.7 Network Flow Authorization-Transition Control

**TH-HARNESS-REQ-470**
`POST /api/v1/test/runtime/network-flow-auth-transitions` is a harness test route with the same enablement, host/origin, and token authorization predicates as the runtime identity route. The route MAY arm one fixture-only authorization-transition control for opted-in Network Flow tests that need a route-time authorization change, hidden-resource assertion, cursor authorization recheck, or extension-resource invalidation trigger. It MUST NOT be exposed as production API behavior, MUST NOT be listed in production OpenAPI, MUST NOT accept ordinary session, role, CSRF, bearer, bootstrap-token, or `deployment_admin` authorization as a substitute for the test-route token, and MUST fail host/origin or token checks before decoding the body or arming any transition.
Verified by: TH-HARNESS-AC-053

**TH-HARNESS-REQ-471**
Network Flow authorization-transition boundaries, transition kinds, resource kinds, and hidden-response kinds are closed harness tokens. They identify fixture-only injection and assertion points and do not define product authorization semantics, incident membership rules, hidden-resource status codes, cursor algorithms, WebSocket contracts, or public API compatibility. The supported boundary tokens are exactly `network_flow.route.before_authorization`, `network_flow.route.after_authorization_before_lookup`, `network_flow.route.after_lookup_before_response`, `network_flow.cursor.before_authorization_recheck`, `network_flow.websocket.before_invalidation_publish`, and `network_flow.fixture.after_transition`. The supported `transition_kind` values are exactly `incident_membership_revoked`, `incident_membership_restored`, `incident_soft_deleted`, `network_flow_table_soft_deleted`, `network_flow_table_renamed`, `session_revoked`, and `extension_claim_removed`. The supported `resource_kind` values are exactly `incident`, `network_flow_table`, `network_flow_cursor`, `network_flow_graph`, `network_flow_contributors`, and `network_flow_workspace`. The supported `hidden_response_kind` values are exactly `not_found`, `forbidden_without_resource`, `empty_collection`, `cursor_rejected`, `extension_profile_not_claimed`, and `invalidation_event`.
Verified by: TH-HARNESS-AC-053

**TH-HARNESS-REQ-472**
The request body MUST be a JSON object with exactly the fields below.

| Field                        | Required | Behavior |
| ---------------------------- | -------- | -------- |
| `boundary`                   | yes      | One of the closed Network Flow auth-transition boundary tokens in TH-HARNESS-REQ-471. |
| `transition_kind`            | yes      | One of the closed transition kinds in TH-HARNESS-REQ-471. |
| `actor_ref`                  | yes      | Safe fixture actor reference matching `^[A-Za-z0-9._:-]{1,128}$`. |
| `incident_ref`               | yes      | Safe fixture incident reference matching `^[A-Za-z0-9._:-]{1,128}$`. |
| `resource_kind`              | yes      | One of the closed resource kinds in TH-HARNESS-REQ-471. |
| `resource_ref`               | yes      | Safe fixture resource reference matching `^[A-Za-z0-9._:-]{1,128}$`. |
| `hidden_response_kind`       | yes      | One of the closed hidden-response kinds in TH-HARNESS-REQ-471. |
| `must_not_disclose_resource` | yes      | Must be `true`; controls that allow resource disclosure are not accepted. |
| `correlation_key`            | no       | Optional safe fixture correlation key matching `^[A-Za-z0-9._:-]{1,128}$`; when supplied, consumption requires the same key. |
| `consume_once`               | yes      | Must be `true`; persistent or multi-consume transitions are not accepted. |

Verified by: TH-HARNESS-AC-053

**TH-HARNESS-REQ-473**
Unknown members, missing required fields, non-object JSON, invalid JSON, unsupported boundary, unsupported transition kind, unsupported resource kind, unsupported hidden-response kind, unsafe refs, `must_not_disclose_resource` other than `true`, or `consume_once` other than `true` MUST fail with `400`, `error.code=invalid_network_flow_auth_transition_request`. Successful arming MUST return HTTP `201` with `cartulary.test.network_flow_auth_transition_control.v1` in the standard success envelope. The response MUST include a generated `control_id`, exact boundary, transition kind, actor ref, incident ref, resource kind/ref, hidden response kind, optional correlation key, `must_not_disclose_resource=true`, and `consume_once=true`. The response MUST NOT include product session credentials, membership row IDs, role grants, route-token material, raw hidden resource details, database credentials, object-store credentials, raw fixture source paths, or private runtime state.
Verified by: TH-HARNESS-AC-053

**TH-HARNESS-REQ-474**
An armed Network Flow auth-transition control is in-memory harness runtime state and is consumed only by an opted-in Network Flow test implementation at the exact boundary, actor ref, incident ref, resource ref, and optional correlation key. A mismatch MUST leave the control pending. A consumed control MUST be removed before the fixture applies the transition or hidden-resource assertion, so retry/replay reaches ordinary behavior unless another control has been armed. A request to arm a duplicate exact boundary/actor/incident/resource tuple while one is pending MUST fail before replacement with HTTP `409`, `error.code=test_network_flow_auth_transition_already_armed`; independent tuples MAY be armed concurrently. Browser reset MUST replace the backend process, so registered auth-transition controls cannot survive into the replacement backend.
Verified by: TH-HARNESS-AC-053

### 12.2.8 Network Flow Audit Assertion Control

**TH-HARNESS-REQ-475**
`POST /api/v1/test/runtime/network-flow-audit-assertions` is a harness test route with the same enablement, host/origin, and token authorization predicates as the runtime identity route. The route MAY arm one fixture-only audit assertion for opted-in Network Flow tests that need exact domain audit occurrence counts, zero-occurrence cases, or no-additional-occurrence replay checks. It MUST NOT be exposed as production API behavior, MUST NOT be listed in production OpenAPI, MUST NOT accept ordinary session, role, CSRF, bearer, bootstrap-token, or `deployment_admin` authorization as a substitute for the test-route token, and MUST fail host/origin or token checks before decoding the body or arming any assertion.
Verified by: TH-HARNESS-AC-054

**TH-HARNESS-REQ-476**
Network Flow audit assertion kinds, event codes, and resource kinds are closed harness tokens. They identify fixture-only assertion points and do not define product audit semantics, audit storage, operation authorization, event payload shape, or public API compatibility. The supported `assertion_kind` values are exactly `exact_count`, `zero_occurrences`, and `no_audit_replay`. The supported `event_code` values are exactly `network_flow_table_created`, `network_flow_table_renamed`, `network_flow_table_soft_deleted`, `network_flow_graph_query_executed`, `network_flow_indicator_binding_created`, and `network_flow_indicator_binding_reused`. The supported `resource_kind` values are exactly `network_flow_table`, `network_flow_graph`, `network_flow_indicator_binding`, and `network_flow_import`.
Verified by: TH-HARNESS-AC-054

**TH-HARNESS-REQ-477**
The request body MUST be a JSON object with exactly the fields below.

| Field                       | Required | Behavior |
| --------------------------- | -------- | -------- |
| `assertion_kind`            | yes      | One of the closed assertion kinds in TH-HARNESS-REQ-476. |
| `event_code`                | yes      | One of the closed Network Flow audit event codes in TH-HARNESS-REQ-476. |
| `operation_ref`             | yes      | Safe fixture operation reference matching `^[A-Za-z0-9._:-]{1,128}$`. |
| `actor_ref`                 | yes      | Safe fixture actor reference matching `^[A-Za-z0-9._:-]{1,128}$`. |
| `incident_ref`              | yes      | Safe fixture incident reference matching `^[A-Za-z0-9._:-]{1,128}$`. |
| `resource_kind`             | yes      | One of the closed resource kinds in TH-HARNESS-REQ-476. |
| `resource_ref`              | yes      | Safe fixture resource reference matching `^[A-Za-z0-9._:-]{1,128}$`. |
| `baseline_count`            | yes      | Non-negative count before the exercised product operation, maximum `1000000`. |
| `expected_final_count`      | yes      | Exact count expected after the exercised product operation, maximum `1000000`. |
| `expected_replay_increment` | yes      | Exact additional occurrence count expected from committed replay, maximum `1000000`; no-audit replay assertions require `0`. |
| `correlation_key`           | no       | Optional safe fixture correlation key matching `^[A-Za-z0-9._:-]{1,128}$`; when supplied, consumption requires the same key. |
| `consume_once`              | yes      | Must be `true`; persistent or multi-consume assertions are not accepted. |

`expected_final_count` MUST be greater than or equal to `baseline_count`. For `assertion_kind="zero_occurrences"`, `baseline_count`, `expected_final_count`, and `expected_replay_increment` MUST all be `0`. For `assertion_kind="no_audit_replay"`, `expected_replay_increment` MUST be `0`.
Verified by: TH-HARNESS-AC-054

**TH-HARNESS-REQ-478**
Unknown members, missing required fields, non-object JSON, invalid JSON, unsupported assertion kind, unsupported event code, unsupported resource kind, unsafe refs, negative or oversized counts, `expected_final_count < baseline_count`, assertion-kind count-rule violations, or `consume_once` other than `true` MUST fail with `400`, `error.code=invalid_network_flow_audit_assertion_request`. Successful arming MUST return HTTP `201` with `cartulary.test.network_flow_audit_assertion_control.v1` in the standard success envelope. The response MUST include a generated `assertion_id`, exact assertion kind, event code, operation ref, actor ref, incident ref, resource kind/ref, baseline count, expected final count, expected replay increment, optional correlation key, and `consume_once=true`. The response MUST NOT include product session credentials, raw audit payloads, raw source data, safe-digest key material, cursor tokens, membership row IDs, role grants, database credentials, object-store credentials, raw fixture source paths, or private runtime state.
Verified by: TH-HARNESS-AC-054

**TH-HARNESS-REQ-479**
An armed Network Flow audit assertion is in-memory harness runtime state and is consumed only by an opted-in Network Flow test implementation at the exact event code, operation ref, resource kind/ref, and optional correlation key. A mismatch MUST leave the assertion pending. A consumed assertion MUST be removed before the fixture compares the observed product audit counts, so retry/replay reaches ordinary behavior unless another assertion has been armed. A request to arm a duplicate exact event/operation/resource tuple while one is pending MUST fail before replacement with HTTP `409`, `error.code=test_network_flow_audit_assertion_already_armed`; independent tuples MAY be armed concurrently. Browser reset MUST replace the backend process, so registered audit assertions cannot survive into the replacement backend.
Verified by: TH-HARNESS-AC-054

### 12.3 Retired Runtime Reset Route

`POST /api/v1/test/runtime/reset`, its empty-body request contract,
`cartulary.test.runtime_reset.v1`, direct Recovery injection into HTTP
composition, and the reset-hook contribution API are retired together. An
enabled harness backend returns ordinary not-found behavior for that path. A
browser reset is an owned stack-lifecycle transition outside the product
process; no product HTTP response can close it.

### 12.4 Saved-View System Fixture Request Body

The saved-view system fixture request body MUST be a JSON object with exactly the fixture fields below.

| Field            | Required | Behavior                                                                                                                                 |
| ---------------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| `view_schema_id` | yes      | Stable workbook `view_schema_id`; unknown or empty values fail with `400`, `error.code=invalid_mutation_payload`.                       |
| `display_name`   | yes      | Saved-view display name; normalized with ordinary saved-view display-name normalization.                                                   |
| `query_json`     | yes      | Persisted saved-view query JSON; normalized with ordinary saved-view persisted-query normalization for the selected `view_schema_id`.      |
| `layout_json`    | no       | Saved-view layout JSON; omitted values receive the ordinary normalized default layout for the selected `view_schema_id`.                  |

Unknown members, including `scope`, `owner_user_id`, `saved_view_id`, `incident_id`, `created_at`, `updated_at`, and `saved_view_version`, MUST fail with `400`, `error.code=invalid_mutation_payload`. Missing required fields, non-object JSON, invalid JSON, invalid `view_schema_id`, invalid `display_name`, invalid persisted query shape, or invalid layout shape MUST fail through the ordinary saved-view mutation error envelope with `400`, `error.code=invalid_mutation_payload`, and field/reason details when available.

Successful fixture creation MUST return a saved-view resource with `scope='system'`, `owner_user_id=null`, the path `incident_id`, a generated `saved_view_id`, normalized `display_name`, normalized `query_json`, normalized `layout_json`, and store-managed timestamps/version. The returned resource MAY be visible through ordinary saved-view list behavior because `scope='system'` is visible fixture data; that visibility does not make the fixture route a production create API and does not change the ordinary saved-view create rule that rejects `scope='system'`.

### 12.5 Browser Reset Admission and Deadlines

Only the scheduler-owned lifecycle unit for the currently leased mutable
browser allocation may reset it. The backend has its declared drain deadline;
database mutation has a `30s` destructive-reset deadline; replacement backend
readiness has a `120s` deadline. The reset-unit watchdog is derived from those
stage deadlines plus bounded evidence-publication overhead. Every stage and the
complete reset have `attempt=1`; no reset or product retry is allowed. Immutable
performance-fixture clones reject reset admission. The authored
`browser_reset_policy` in `tools/execution_topology_manifest.json` is the typed
projection of these deadlines and the one-attempt limit.

### 12.6 Browser Stack Reset Algorithm and Partial Failure

**TH-HARNESS-REQ-452**
The browser reset coordinator owns one lifecycle transition for the selected
mutable browser allocation. It MUST validate allocation ownership and create a
reset ID; gracefully terminate the backend within its declared drain deadline;
prove the backend process group and listener are gone; acquire the typed
exclusive Recovery-target admission; reset persistent database state; release
exclusive admission; clear the harness-owned object-store bucket or prefix and
Playwright state; then start exactly one replacement backend. The frontend and
owned fixture allocation remain alive. The replacement uses the same ports,
immutable runtime profile, startup-only extension claims, key rings,
configuration fingerprint, database, bucket, and runtime-only credential. It
MUST reacquire ordinary application Recovery-serving admission and pass runtime
identity/readiness before the coordinator publishes a monotonically increasing
backend generation.

The backend is never paused and resumed in-process. Process replacement clears
HTTP, WebSocket, job, dispatcher, janitor, pool, advisory-lease, test-clock,
fault, randomness, auth-transition, and audit-assertion state without a registry
of reset hooks. Product server composition receives only the runtime credential
and MUST NOT acquire, resolve, or retain the Recovery credential. The Recovery
reset controller runs out of process and shares the exact typed exclusive
Recovery-target primitive used by Recovery assembly. Exclusive admission MUST
be impossible while any application backend holds ordinary serving admission,
and MUST span the complete reset transaction.

The database reset table set is selected by this algorithm:

```text
select_reset_tables(database):
  query information_schema.tables
  keep rows where table_schema = "public"
  keep rows where table_type = "BASE TABLE"
  reject table_name in ("goose_db_version", "schema_migration_lineage")
  order table_name ascending
  return table_name list
```

The Recovery reset controller MUST inventory the selected list in ASCII order and
execute one `TRUNCATE TABLE public.<table> ... CASCADE` statement inside one
database transaction using identifier-safe table quoting. It MUST reset each
owned sequence explicitly with Recovery sequence authority and MUST NOT use
`RESTART IDENTITY`, because Recovery owns no objects. If the list is empty,
truncate is a successful no-op and bootstrap restoration still runs. Before
commit, bootstrap restoration yields exactly one active deployment admin and one
bootstrap marker, `route_idempotency` has zero rows, and the before/after row
counts of both `goose_db_version` and `schema_migration_lineage` are equal and
nonzero. A failed database stage rolls the transaction back.

Database table truncation, sequence reset, bootstrap restoration, metadata
verification, and commit are one Recovery transaction. Object-store reset and
Playwright-state clearing occur only after database commit. The lifecycle MUST
NOT claim rollback across those later surfaces. Failure before backend stop is
proven MUST NOT begin database mutation. Any failure after reset admission,
including backend stop, database, object store, browser-state, replacement
startup, readiness, evidence publication, or provider cleanup, taints and
quarantines the allocation before product work can start.

`cartulary.test.database_reset_diagnostic.v1` is retained at
`reset-boundary/<label>.database-reset.json`. It includes reset ID, terminal
status, `attempt=1`, a closed reset-stage token, nullable five-character
PostgreSQL SQLSTATE, timeout flag, duration, sorted table/count proofs, and
normalized failure class/reason. The typed database failure preserves its
wrapped cause internally but exposes only the exact stage and normalized
`pgconn.PgError.Code`. It MUST NOT retain raw SQL, DSNs, database names,
credentials, raw errors, or backend IDs.

`cartulary.browser_reset_attempt.v1` is retained at
`reset-boundary/<label>.attempt.json` and is the lifecycle unit's authoritative
failure evidence. It records the ordered lifecycle stages, previous and
replacement backend generations, database-diagnostic reference,
persistent/browser reset proofs, taint state, and terminal classification.
Missing or malformed required evidence is `artifact/artifact_error` and public
exit `11`; non-timeout lifecycle failure is `harness/fixture_error` and public
exit `3`; watchdog expiry is `timing/timeout_failure` and public exit `13`.
Lifecycle units MUST NOT infer `product/test_assertion_failure`.

On failure, the broker persists the lease as `quarantined` before provider
cleanup. `cartulary.harness_fixture_lease.v4` records `cleanup_outcome` as one of
`not_required`, `pending`, `completed`, or `failed`, plus nullable normalized
cleanup failure. The broker persists the terminal record after cleanup even
when cleanup fails. Reset failure remains primary and cleanup failure is
secondary; the allocation MUST never return to reusable state.
Verified by: TH-HARNESS-AC-008, TH-HARNESS-AC-034

### 12.7 Browser Stack Reset Success Readiness

Success requires all of the following:

- the old backend process group and listener were proven gone before mutation;
- exclusive Recovery-target admission spanned the database transaction;
- both migration metadata tables were preserved;
- bootstrap state was restored, mutable state and route idempotency were cleared;
- the object-store namespace and Playwright state were cleared;
- the replacement backend has the same runtime/configuration fingerprint, a new
  generation, ordinary serving admission, identity proof, and readiness;
- both reset diagnostics validate and no taint or quarantine is present.

## 13. Cleanup and Destructive Safety

**TH-HARNESS-REQ-500**
Cleanup is destructive. Cleanup commands MUST delete only paths or resources satisfying the exact ownership predicates in this section.
Verified by: TH-HARNESS-AC-009, TH-HARNESS-AC-010

**TH-HARNESS-REQ-501**
External Go build, module, and temporary-work caches are borrowed machine state. No public or
private harness target, including `doctor`, `clean`, and `distclean`, may
delete, rename, quarantine, synthesize markers in, or broadly repair those
caches. A recognized automatic-toolchain corruption MUST fail closed and name
only the computed toolchain-version/platform extraction directory and matching
`.ziphash` entry. Repair requires an explicit operator action after active Go
and Make processes have stopped. Normal checksum-verified Go download into an
absent cache is readiness/bootstrap population, not repair authority.
Verified by: TH-HARNESS-AC-092

`make clean` and `make distclean` are repo-local cleanup commands. They MUST NOT remove caller-supplied external result roots and MUST NOT stop local Compose services. Local service teardown belongs to `make services-down`, not to repo-local cleanup.
Frontend dependency install state is a coupled repo-local artifact set. `make clean` MUST preserve installed dependency roots for local loop speed, while `make distclean` MUST remove the repo-local pnpm store, frontend install stamps, and root/workspace `node_modules` directories together so stale package-manager metadata cannot survive without its store.
Harness cache state is repo-local acceleration state. `make clean` MUST preserve
default `.cache/cartulary/*` roots so ordinary cleanup does not erase valid warm
state. `make distclean` MUST remove the default graph and tool cache roots under
`.cache/cartulary/`. Neither cleanup target may remove caller-supplied cache
directories outside the repository.
Fallow configs, rule packs, schemas, and any future reviewed baselines under repository source/tool roots are harness inputs, not cleanup-owned artifacts. Fallow run-root outputs are ordinary retained artifacts and may be removed only through result-root cleanup predicates.

### 13.1 Path Algorithm

```text
normalize_cleanup_candidate(path):
  reject empty string
  reject NUL
  reject "/"
  reject "."
  reject ".."
  reject any caller-supplied segment equal to ".."
  reject backslash on POSIX conformance hosts
  resolve relative paths against repository root
  reject absolute paths outside repository root
  reject protected repository roots named in the table below when they are named as cleanup candidates
  lstat path
  if path is symlink:
    unlink symlink object only
    MUST NOT follow target
  if path is directory:
    remove directory tree only after every traversed entry remains under the candidate root by lexical path and lstat traversal
```

The protected repository root set is closed in the current profile:

| Protected root | Protection rule |
| --- | --- |
| `.git` | Reject when named directly as a cleanup candidate. |
| `docs` | Reject when named directly as a cleanup candidate. |
| `cmd` | Reject when named directly as a cleanup candidate. |
| `internal` | Reject when named directly as a cleanup candidate. |
| `apps` | Reject when named directly as a cleanup candidate. |
| `packages` | Reject when named directly as a cleanup candidate. |
| `contracts` | Reject when named directly as a cleanup candidate. |
| `db/migrations` | Reject when named directly as a cleanup candidate. |
| `db/queries` | Reject when named directly as a cleanup candidate. |
| `configs` | Reject when named directly as a cleanup candidate. |
| `scripts` | Reject when named directly as a cleanup candidate. |
| `tools` | Reject when named directly as a cleanup candidate. |
| `go.mod` | Reject when named directly as a cleanup candidate. |
| `go.sum` | Reject when named directly as a cleanup candidate. |
| `package.json` | Reject when named directly as a cleanup candidate. |
| `pnpm-lock.yaml` | Reject when named directly as a cleanup candidate. |
| `pnpm-workspace.yaml` | Reject when named directly as a cleanup candidate. |

A child path under a protected root MAY be removed only when Section 13.2 or another adopted cleanup table explicitly lists that exact path or path family as cleanup-owned. Missing cleanup-owned paths are successful no-ops. A path that is both protected and cleanup-owned MUST use the narrower cleanup-owned row; broad ancestor deletion remains rejected.

### 13.2 Cleanup Scope

| Command               |      Removes default result root? | Removes custom `CARTULARY_TEST_RESULTS_DIR`? | Removes default `.cache/cartulary` cache roots? | Removes external Go caches? | Stops Docker/Compose globally? |
| --------------------- | --------------------------------: | -------------------------------------------: | ---------------------------------------------: | --------------------------: | -----------------------------: |
| `make clean`          | yes, only default registered path |                                           no |                                             no |                          no |                             no |
| `make distclean`      | yes, only default registered path |                                           no |                                            yes |                          no |                             no |
| `make services-down`  |                                no |                                           no |                                             no |                          no | no; stops only this repo's local Compose services and preserves named volumes |
| Service-suite cleanup |        only suite-owned artifacts |                                           no |                                             no |                          no |                             no |
| Stale janitor         |        proof-gated resources only |                                           no |                                             no |                          no |                             no |

`make distclean` owns removal of `.pnpm-store`, the repository-root `node_modules` directory, workspace package `node_modules` directories under `apps/web` and `packages/*`, and default repo-local cache roots under `.cache/cartulary/`. It MUST NOT name `.cache` itself as a cleanup candidate. Missing workspace dependency roots or cache roots are not cleanup failures.

After a candidate has passed the Section 13.1 containment and symlink checks,
cleanup MAY add owner read, write, and search permission to traversed real
directories inside that candidate when required to remove read-only generated
or downloaded descendants. It MUST NOT follow a symlink, change a symlink
target, change a preserved child, or change permissions outside the validated
candidate tree.

### 13.2.1 Local Service And Data Reset Scope

`make services-down` MUST stop only the local Compose services declared for repository development and MUST preserve named volumes. It MUST NOT pass a Compose volume-removal flag. `db-down` is not a current public command binding; new and existing automation MUST use `services-down`.

`make db-migrate` MUST apply the repository migration line to the local development database without dropping, recreating, or truncating that database and without resetting, deleting, initializing, or inspecting object storage. It MAY start Postgres to perform the migration. It MUST use the same `migrate up` application surface as deployable migration execution and MUST surface migration-remediation reports without rewriting them. It MUST NOT overwrite an inherited managed Postgres DSN selected by the config's database `service_ref`; for the default local development service it MAY derive a local Compose DSN only when the selected DSN environment value is unset.

`make db-reset` MUST recreate only the local development database and rerun migrations. It MAY start Postgres to perform the reset, but it MUST NOT reset, delete, or inspect object storage. A real `db-reset` MUST reject before Compose, database, migration, or object-store commands unless `CARTULARY_DESTRUCTIVE_CONFIRM=db-reset` was supplied on the Make command line.

`make dev` MUST start the backend process and prove backend readiness before starting the frontend process. If backend readiness fails because the database is behind the current line, the dev-stack diagnostic MUST direct the caller to `make db-migrate`; if the backend log reports `prod_ddl_rebaseline_v2` with `historical_migration_lineage`, the diagnostic MUST emit the exact reset-only remediation hint from Core 01 REQ-01-661. The frontend process MUST NOT start after a backend readiness failure.

`make object-store-reset` MUST clear only objects in the configured local object-store bucket and MUST leave the bucket present afterward. In the current implementation profile the local object store is SeaweedFS S3, and the public command and command ID are provider-neutral. A real `object-store-reset` MUST reject before Compose or object-store commands unless `CARTULARY_DESTRUCTIVE_CONFIRM=object-store-reset` was supplied on the Make command line.

### 13.3 Stale Janitor Thresholds

| Resource        | Completed-run predicate                                         | Uncompleted stale predicate                | Active-resource rule                                                                     |
| --------------- | --------------------------------------------------------------- | ------------------------------------------ | ---------------------------------------------------------------------------------------- |
| Database        | Completed summary or lease cleanup state older than 15 minutes. | Lease or metadata older than 24 hours.     | Active connections may be terminated only after proof predicate passes.                  |
| Bucket          | Completed summary or lease cleanup state older than 15 minutes. | Lease or metadata older than 24 hours.     | Delete only generated bucket/prefix with proof metadata.                                 |
| Container       | Completed summary or lease cleanup state older than 15 minutes. | Harness Docker labels older than 24 hours. | Running container may be stopped only if proof predicate passes and label owner matches. |
| Browser fixture | Completed target summary older than 15 minutes.                 | Fixture metadata older than 6 hours.       | Delete only generated fixture directory with ownership metadata.                         |
| Browser process/session | Completed session lease older than 15 minutes.          | Session lease older than 6 hours with matching runtime-root marker and process command/env proof. | Running processes may be stopped only when PGID, runtime root, command/env proof, and lease identity all match; a port listener alone is never sufficient proof. |

For container cleanup, an already-deleting Docker resource is treated as deferred successful cleanup only after the same proof predicates pass. This compatibility rule exists to make repeated service-backed public targets reproducible under Docker's asynchronous removal lifecycle; it MUST NOT broaden cleanup authority to unlabelled, current-suite, or externally owned containers.

### 13.4 Dry-Run Contract

| Setting                                        | Behavior                                                                                                                    |
| ---------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| `CARTULARY_CLEANUP_DRY_RUN` omitted or not `1` | Cleanup or reset may delete resources satisfying predicates and confirmation rules.                                         |
| `CARTULARY_CLEANUP_DRY_RUN=1`                  | Cleanup MUST emit deletion candidates and reasons and MUST NOT start services, stop services, delete files, delete DBs, delete bucket objects, delete buckets, delete containers, or delete browser fixtures. |

Dry-run output MUST include normalized path or resource identity, proof predicate, action that would be taken, and rejection reason for retained candidates. For human destructive targets, the dry-run line format is:

```text
DRY-RUN <action> <normalized-identity> <proof-or-rejection-reason>
```

`CARTULARY_DESTRUCTIVE_CONFIRM` is ignored for dry-runs. Inherited environment values MUST NOT satisfy the confirmation predicate for public Make targets; only Make command-line values are valid confirmation sources.

### 13.5 Parent-Death and Reaper Rule

Immediate cleanup after parent death is not guaranteed. The conformance guarantee is:

- owned resources carry enough lease or proof metadata for later stale janitor evaluation;
- detached reaper scheduling is optional unless the command declares it;
- if a detached reaper is scheduled, it writes `reaper-scheduled.json` with lease ID, started-at timestamp, target resources, and timeout.

## 14. Platform and CI Support

**TH-HARNESS-REQ-550**
The current conformance support matrix is closed by this section. A target may be run elsewhere, but unsupported environments MUST NOT be used for current harness conformance claims.
Verified by: TH-HARNESS-AC-012

**TH-HARNESS-REQ-551**
Make-owned Go work MUST separate the launcher version from the effective
repository toolchain and select exactly the reviewed `GO_TOOLCHAIN` pin. The
pin MUST agree among `tools/toolchain_pins.json`, `go.mod`, and the authored
Make mirror before child work. The private readiness boundary has two modes:

- `diagnose` is read-only, makes no network request, and is used by `doctor`
  without preventing the remaining doctor checks from running;
- `ensure` runs before Go-consuming Make work, may use Go's ordinary verified
  automatic download when the exact toolchain is absent, and rejects recognized
  corruption or a non-exact effective version before child work.

Both mismatch and readiness failure map to `configuration_error`, exit `2`.
Before either mode invokes selected Go work, the harness MUST resolve the
machine-state root and the build, module, and temporary-work directories using
Section 5.5. `diagnose` validates and reports these paths without creating or
probing them mutably. `ensure` creates missing directories, verifies they are
writable, and passes them to all Make-owned Go work as `GOCACHE`, `GOMODCACHE`,
and `GOTMPDIR`. No helper may restore a literal `/tmp/cartulary-go-*` fallback
or treat native Go variables as a second public configuration interface.
`doctor` MUST report each resolved path, its backing filesystem identity, and
available capacity without enforcing an arbitrary minimum. Existing legacy
`/tmp/cartulary-go-*` directories MAY produce a manual-cleanup advisory but
MUST NOT be migrated, removed, or reused automatically.

Go-tool bootstrap MUST complete readiness before opening a private staging
directory, install only into that unique directory, validate the expected
executable, and atomically replace the versioned destination. Failure and
interruption MUST preserve the previous destination and remove private staging
state.
Verified by: TH-HARNESS-AC-092

| Environment/tool                                   | Current conformance status                 | Required evidence                                                                                                                             |
| -------------------------------------------------- | ------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------- |
| Linux x86_64 with Docker Engine and Docker Compose | required                                   | Full acceptance matrix.                                                                                                                       |
| WSL2 Ubuntu with Docker Desktop integration        | supported compatibility profile            | `doctor`, `test-fast`, `check`, browser E2E smoke.                                                                                            |
| macOS                                              | unsupported for current conformance        | None.                                                                                                                                         |
| Windows native                                     | unsupported                                | None; use WSL2 profile.                                                                                                                       |
| Hosted CI provider                                 | provider-neutral only                      | `make ci`; no annotation/upload claims.                                                                                                       |
| Podman/Podman Compose                              | unsupported                                | None.                                                                                                                                         |
| Docker                                             | required for service-backed targets        | Missing Docker yields `preflight_error`.                                                                                                      |
| Docker Compose                                     | required for local Compose targets         | Missing Compose yields `preflight_error` when Compose is absent and `service_readiness_timeout` when Compose-started service readiness fails. |
| Go/Node/pnpm/Playwright/Bash utilities             | required as pinned by repository procedure | Version mismatch yields `configuration_error`.                                                                                                |
| Linux inotify capacity                             | required only for Vite dev surfaces        | Low watcher limits or exhausted watcher usage MUST be diagnosed before Vite dev startup; release/browser E2E preview paths MUST NOT require this preflight. |

Harness diagnostics MAY report Linux inotify `max_user_watches`, `max_user_instances`, best-effort current watcher usage, and bounded operator diagnostics. The harness MUST NOT mutate host sysctl settings.

## 15. Security and Redaction

**TH-HARNESS-REQ-600**
Centralized summaries, machine output, and retained logs captured by harness wrappers MUST be redacted before retention and before stdout emission.

Retained result roots MUST also be free of secret-capable runtime files by
construction. Every terminal work-graph path MUST run the schema-validated
forbidden-filename, symlink, and bounded streaming content scan before it can
report success. The scan MUST compare against the actual suite-generated
credential values while those values remain available in private memory.
Verified by: TH-HARNESS-AC-011

**TH-HARNESS-REQ-601**
Redaction MUST be applied to captured stdout, stderr, wrapper diagnostics,
machine JSON, bounded retained diagnostic tails, and summary artifacts before
those bytes are written outside a private runtime working file or emitted to
stdout/stderr. Full service environments, DSNs, private leases, key rings,
Playwright state, fixture credential bundles, and raw child captures are
private runtime material and MUST NOT be projected into retained logs or
service env dumps. A redaction failure or retained-boundary scan failure MUST
fail the public target with `failure_class=artifact`,
`failure_reason=artifact_error`, and public exit code `11` unless an earlier
primary failure is preserved by Section 9.1.
Verified by: TH-HARNESS-AC-011

**TH-HARNESS-REQ-602**
The redaction algorithm MUST apply to both keys and values after decoding structured JSON where possible and to raw text otherwise. Matching MUST be case-insensitive for key names and header names. At minimum, the algorithm MUST redact:

- variables, JSON keys, HTTP headers, and CLI arguments whose names match the secret pattern table;
- URL userinfo and DSN password segments;
- bearer-token, session-cookie, JWT, private-key, and object-store credential forms in raw text;
- `Authorization`, `Cookie`, `Set-Cookie`, and `X-Cartulary-Test-Route-Token` values.
Verified by: TH-HARNESS-AC-011

Structured redaction MUST preserve schema-owned container shapes and scalar types unless the value itself is secret. Object and array fields such as `service_sessions`, `browser_stage_sessions`, `session_target`, `cleanup_status`, `lease_file`, and timing fields MUST NOT be replaced merely because their names contain a secret-related substring. Secret key matching MUST use exact or anchored credential-name patterns rather than broad substring matching that can redact structural diagnostics.

Redaction is defense in depth and MUST NOT be used to justify retaining a
secret-capable runtime file. Replacement diagnostics are closed,
schema-validated projections that may retain phase, failure class and reason,
readiness state, connection and resource class, and cleanup outcome without
retaining operational values.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-011

**TH-HARNESS-REQ-604**
Structured secret-key matching is closed. Before comparing a structured key name, the redactor MUST uppercase it and replace every non-ASCII-alphanumeric run with one `_`, then trim leading and trailing `_`. The resulting token is secret-bearing only when it equals or has an anchored credential suffix/prefix equivalent to one of: `PASSWORD`, `PASS`, `PWD`, `TOKEN`, `JWT`, `BEARER`, `API_KEY`, `ACCESS_KEY`, `SECRET_KEY`, `AUTHORIZATION`, `COOKIE`, `SET_COOKIE`, or `X_CARTULARY_TEST_ROUTE_TOKEN`. Substring-only matches such as `session_target` containing `token` across token boundaries MUST NOT redact the field.

Raw-text redaction MUST apply after structured redaction to these closed families: URL userinfo (`scheme://userinfo@host`), PostgreSQL-style DSN password segments (`password=` or `:password@` credential forms), bearer authorization headers, JWT-like three-part base64url tokens, PEM private-key blocks, and S3-compatible access-key or secret-key assignments. Structured redaction MUST preserve object and array shapes and preserve numeric, boolean, and null scalar types unless that scalar value itself is secret. A redaction write or validation failure maps to `failure_class=artifact`, `failure_reason=artifact_error`, and public exit `11` unless Section 9.1 preserves an earlier primary failure.
Verified by: TH-HARNESS-AC-011, TH-HARNESS-AC-036

**TH-HARNESS-REQ-605**
Runtime and support security scans MUST derive their default support-root
exclusions and support package patterns from the schema-validated test-support
inventory required by TH-HARNESS-REQ-067. Every registered Go support root
MUST be included in the support scan. A root MAY be excluded from the runtime
profile only when the inventory classifies it as test support; a package that
is compiled into a production binary, including a module-owned harness-control
package, MUST remain in the runtime profile.

An unknown `internal/**/testsupport` root, a registered root missing from the
support profile, a duplicate path, or a path excluded from both profiles MUST
fail before the security tool starts. Publicly declared security overrides
remain governed by the Section 5 input registry and MUST NOT silently replace
inventory validation.

Verified by: TH-HARNESS-AC-011, TH-HARNESS-AC-056

SeaweedFS strict release evidence MUST derive its redaction scan input set from
the current release-evidence run, the current `seaweedfs-compatibility` report,
and the current backend-process backup/restore row evidence. Strict release
targets MUST run `backend-process` and `seaweedfs-compatibility` as current-run
prerequisites. Stable copies, historical delivery-shaped paths, newest-run
fallback evidence, retained `services-up` reports, and retired source-migration
evidence MUST NOT satisfy the gate. Missing selected child artifacts are
blocking findings and MUST NOT be replaced with fallback evidence.

**TH-HARNESS-REQ-675**
The current profile contains no legacy-source object-storage transition product
behavior or release-support evidence. The authored task surface, verification
catalog, families, topology, occurrence policy, release evidence, and release
gates MUST contain no dedicated transition target, row, or replacement owner.
SeaweedFS release evidence instead covers
current S3 compatibility, backup integrity and restoration, operator-private
artifact redaction, and source-owner storage-reference preservation.
Historical result roots MAY contain retired migration artifacts but MUST NOT be
selected as current conformance or release evidence.
Verified by: TH-HARNESS-AC-011, TH-HARNESS-AC-015

The operational recovery smoke MUST build its uniquely tagged application image exactly once before Compose startup. Startup MUST consume that completed image without requesting a second build. Removing duplicate build orchestration MUST NOT change readiness, backup, restore-verification, route-absence, diagnostic, or cleanup behavior.
Verified by: TH-HARNESS-AC-011, TH-HARNESS-AC-015

**TH-HARNESS-REQ-603**
Retained run roots and target artifact directories MUST be created with owner-only permissions on POSIX conformance hosts unless the caller explicitly supplied a custom result root whose permissions cannot be narrowed without changing ownership. Required summary artifacts and retained logs MUST be written with owner-read/write permissions. A custom result root that is world-writable without the sticky bit, or that cannot protect newly created files from other users on the host, MUST fail preflight with `configuration_error`.

Each scheduler invocation that may create runtime material MUST allocate one
suite-private runtime root below the validated external harness scratch
namespace. The suite root and every owned directory MUST be non-symlink,
current-user-owned `0700`; private files MUST be regular non-symlink `0600`
files. The canonicalized suite root MUST be outside both the repository and
the current result/run roots, and every consumer MUST reject containment
failure, symlink traversal, ownership mismatch, permissive modes, or an
incomplete ownership marker before opening private material. A private runtime
lease carries process and path handles; retained cleanup evidence carries only
opaque lease identity, bounded resource classes, and cleanup proof.

Private files MUST be unlinked after their last consumer closes them and on
setup failure, cancellation, signal, finalizer failure, and suite termination.
The root is removed only after service, process, database, bucket, session, and
open-handle cleanup. Bounded stale cleanup may remove only roots inside the
exact suite-runtime namespace whose private ownership marker validates and is
older than the specified lease age; any unowned, malformed, symlinked, or
out-of-scope entry fails closed. Cleanup claims handle closure, unlink, and
owned-directory removal; it MUST NOT claim physical-media sanitization.

Ephemeral browser state that contains credentials or authentication secrets
MUST be published through one private-state primitive. The producer MUST
serialize the complete value before opening a unique sibling temporary file,
create that file exclusively with owner-read/write permissions, write and
flush the complete bytes, close it, and atomically rename it over the
destination. Publication failure MUST remove the temporary file and preserve
the previously published value. Readers therefore observe a complete old or
complete new value, never a truncated intermediate value.

The worker-admin manifest schema is
`cartulary.playwright_worker_admin_manifest.v1`. It is a closed object with
exact fields `schema_id` and `worker_admins`; every entry is a closed object
with a nonnegative integer `parallel_index` and nonempty string `user_id`,
`email`, and `password`. Parallel indexes, user IDs, and emails MUST each be
unique. A present invalid, malformed, or unsupported manifest fails closed.
Diagnostics MUST NOT include passwords, shared TOTP values, or other secret
contents. The worker-admin manifest and shared TOTP state are ephemeral and
have no legacy reader or persistent migration.
Verified by: TH-HARNESS-AC-003, TH-HARNESS-AC-011, TH-HARNESS-AC-015

Screenshots, videos, traces, visual geometry diagnostics, and Playwright HTML reports are diagnostic secret-bearing artifacts. They MUST NOT be described as safe to upload or publish without separate review. Browser visual targets MAY retain compact geometry diagnostics for workbook screenshot failures, including scroll metrics, visible field keys, required element rectangles, active element identity, and inspector state. Those diagnostics are harness mechanics only; they MUST NOT define product UI behavior or supplement the bounded visual-snapshot refresh authority in TH-HARNESS-REQ-255.

Workbook visual regression tests that capture an outer grid shell while driving an inner grid scrollport MUST normalize and verify both layers before assertion. The screenshot-target shell MUST be reset to `scrollLeft=0` and `scrollTop=0` for left/default viewport captures unless a test explicitly declares a different shell-scroll contract, while the owned grid scrollport MUST be normalized to the test's declared scroll or anchor state. Anchor-based captures that intentionally frame off-screen workbook columns are explicit shell-scroll contracts and MUST still reset stale shell state before computing their deterministic offset. The diagnostic record MUST identify the screenshot target and both shell and scrollport metrics; exact human wording is non-normative. This normalization is harness mechanics only; it does not promote refresh output into product conformance, design conformance, release, or Core 05 publication evidence.

**TH-HARNESS-REQ-606**
`tools/executable_input_policy.json` validates as
`cartulary.executable_input_policy.v1`. It declares closed restricted roots,
standalone documentation-maintenance sources, and machine evidence roots.
Product, catalog, security, release, conformance, and generation commands have
no documentation-read exception mechanism.
Verified by: TH-HARNESS-AC-067

**TH-HARNESS-REQ-607**
The executable-input boundary scans executable and validation sources plus
machine evidence configuration without opening, statting, resolving, hashing,
or enumerating a restricted documentation root. Direct and joined restricted
paths fail the gate. Verification contracts, test-family manifests, generated
artifacts, and production fixtures do not accept documentation references.
Verified by: TH-HARNESS-AC-067

**TH-HARNESS-REQ-608**
Boundary tests use neutral synthetic root names and files. Standalone
documentation maintenance may read documentation only through the explicitly
classified `lint-markdown` command, which is excluded from product, CI,
conformance, release, performance, and claim-publication aggregates.
Verified by: TH-HARNESS-AC-067

**TH-HARNESS-REQ-609**
The semantic-identity scan MUST detect delivery-shaped path components, filenames, symbols, titles, selector values, variables, schema IDs, artifact names, and target IDs case-insensitively. The implementation MAY keep closed detection patterns, but MUST NOT publish or consume a runtime exception registry.
Verified by: TH-HARNESS-AC-067, TH-HARNESS-AC-070

**TH-HARNESS-REQ-610**
Every semantic-identity violation in an executable source, live catalog selector, generated topology identity, public input, schema identity, harness/test path, or active fixture MUST fail the gate. No allowlist, path exception, compatibility classification, or line-number exception is permitted. Product vocabulary that legitimately uses the word “phase” without encoding delivery order remains outside the delivery-shaped patterns and is owned by its product specification.
Verified by: TH-HARNESS-AC-067, TH-HARNESS-AC-070

**TH-HARNESS-REQ-611**
An ambiguous semantic match blocks closure. The scanner MUST report the matched token and normalized location without opening or copying unrelated document content.
Verified by: TH-HARNESS-AC-067

**TH-HARNESS-REQ-612**
Harness observability may emit only the closed attributes and metrics in
Sections 8 and 10. It MUST NOT emit raw commands or arguments, environment
names or values, absolute or source-provided paths, hostnames, process IDs,
headers, URLs, SQL text, runner output, error messages, stack traces, test
symbols, product-authored values, or artifact contents. Failures use only the
existing low-cardinality failure class and failure reason tokens.
The trace bundle MAY identify a digested source artifact by its normalized,
repository-relative retained-artifact path; that reconstruction identity MUST
NOT be copied into OTLP attributes or metrics.
Verified by: TH-HARNESS-AC-075, TH-HARNESS-AC-076

**TH-HARNESS-REQ-613**
Ordinary harness processes MUST ignore inherited variables whose names begin
with `OTEL_` for harness-observability behavior. They MUST NOT initialize an
OTel SDK, autoconfiguration provider, exporter, resource detector, or default
localhost endpoint. The explicit export command receives only its declared
Make inputs through argv; product and browser child processes MUST NOT receive
the export endpoint or header material.
Verified by: TH-HARNESS-AC-076

**TH-HARNESS-REQ-614**
`HARNESS_OTLP_ENDPOINT` MUST be an absolute URL with scheme `https`, except that
`http` is allowed for `localhost`, `127.0.0.0/8`, or `[::1]`. User information,
query, fragment, non-HTTP schemes, encoded authority ambiguity, and redirects
are invalid. The exporter appends `/v1/traces` and `/v1/metrics` to the base
path exactly once and applies a ten-second request timeout with no automatic
retry. Invalid endpoint, selection, or header configuration is
`configuration_error`, exit `2`; collector resolution, connection, timeout,
redirect, or non-success delivery failure is a bounded
`tool_diagnostic_failure`, exit `1`, of the export command only. Neither class
may modify source evidence.
Verified by: TH-HARNESS-AC-076

**TH-HARNESS-REQ-615**
An optional header file MUST be a non-symlink regular file no larger than 64
KiB and owner-readable only; on POSIX, group or other permission bits are
invalid. Its JSON value is a closed object of at most 32 ASCII header names to
string values of at most 4096 bytes each. `host`, `content-length`,
`content-type`, and connection-management headers are forbidden. The path and
all values are redacted from output and retained diagnostics. Header files and
selected retained runs are never modified.
Verified by: TH-HARNESS-AC-076

### 15.1 Secret Pattern Table

| Secret class               | Match rule                                                                          | Redaction token                      |
| -------------------------- | ----------------------------------------------------------------------------------- | ------------------------------------ |
| Passwords                  | Exact or anchored variable/key names for `PASSWORD`, `PASS`, `PWD`, or equivalent credential suffixes | `[REDACTED:password]`                |
| Tokens                     | Exact or anchored variable/key names for `TOKEN`, `JWT`, `BEARER`, cookie headers, or equivalent credential suffixes | `[REDACTED:token]`                   |
| API or access keys         | Exact or anchored `API_KEY`, `ACCESS_KEY`, `SECRET_KEY`, or credential-context key names | `[REDACTED:key]`                     |
| DSNs/URLs with credentials | URL userinfo or DSN password segment present                                        | `[REDACTED:dsn]`                     |
| Object-store credentials   | S3-compatible access key or secret key variables                                    | `[REDACTED:object-store-credential]` |
| Private keys               | PEM private-key block markers                                                       | `[REDACTED:private-key]`             |

### 15.2 Artifact Redaction Table

| Artifact class                      | Redaction requirement                                                                                                      |
| ----------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| Tool/run/target/scheduler summaries | Redact before write.                                                                                                       |
| Machine stdout JSON                 | Redact before stdout.                                                                                                      |
| Captured child stdout/stderr logs   | Redact before retention.                                                                                                   |
| Service env files and env dumps     | Store only redacted credential values unless the file is required for child execution and kept under private runtime root. |
| Browser screenshots/videos/traces   | Diagnostic secret-bearing; not safe for publication.                                                                       |
| Visual geometry diagnostics         | Diagnostic secret-bearing; not safe for publication.                                                                       |
| Playwright HTML reports             | Diagnostic secret-bearing; not safe for publication.                                                                       |
| CI logs                             | Redact using the same token rules before harness-controlled emission.                                                      |

## 16. Integration with Product Specifications

**TH-HARNESS-REQ-650**
Harness verification contracts and catalog rows route behavior classes and
executable selectors only. They MUST NOT copy or resolve product requirement
IDs, acceptance criteria, specification lifecycle status, deleted delivery
registries, or ledgers, and they MUST NOT redefine product behavior under test.
Verified by: TH-HARNESS-AC-013, TH-HARNESS-AC-016

**TH-HARNESS-REQ-651**
Support-profile tests, helper commands, raw aggregate suites, and direct package scripts MUST NOT be counted as authoritative product-conformance evidence unless an adopted verification contract and canonical Make target select them and emit required evidence artifacts.
Verified by: TH-HARNESS-AC-013

**TH-HARNESS-REQ-652**
Load, login bursts, service resets, artificial stress margins, and browser harness stress tests are harness-only unless product specs explicitly adopt them. Browser login bursts MUST NOT be used as the sole evidence for Core 04 session-cap semantics when backend or integration evidence can prove victim selection and revocation delivery directly.
Verified by: TH-HARNESS-AC-013

**TH-HARNESS-REQ-653**
Timing-sensitive browser evidence for asynchronous socket behavior MUST prove the relevant sender readiness, receiver readiness, event identity, and diagnostic capture boundary before starting the measured interaction. A timed assertion MUST measure the product event under test, not page navigation, socket establishment, route cleanup, or waiter attachment.
Verified by: TH-HARNESS-AC-013

**TH-HARNESS-REQ-654**
Harness recovery verification MUST be product-owner-subordinate. A harness target, catalog row, retained artifact, task-surface entry, topology entry, scheduler work unit, or runtime-binary injection rule MUST NOT redefine recovery CLI authorization, logical command grammar, result schema, progress schema, timeout defaults, exit-code mapping, safe-output rules, restore-target preflight, journal behavior, or public route absence.

Any row claiming Core 04 AC-402, AC-427, or AC-428 MUST cite the Core-owned recovery requirements it verifies and MUST treat the recovery CLI as `deployment_admin`-irrelevant. A test that requires `deployment_admin`, session cookies, bearer tokens, CSRF, browser Origin, incident role, common-job authorization, WebSocket authorization, or public HTTP route access for recovery CLI invocation MUST NOT count as recovery conformance evidence.

Legacy tests that characterize a repository implementation requiring `deployment_admin` for recovery commands are negative evidence only. They MUST NOT close product-conformance rows, MUST NOT use blocked or stale status as current closure, and MUST NOT verify AC-402, AC-427, or AC-428.

Verified by: TH-HARNESS-AC-043

**TH-HARNESS-REQ-655**
Operator recovery conformance rows MUST map implementation-owned executable or wrapper behavior to exactly one Core logical command from Core 01 §12.2.1 before product assertions execute.

The current valid conformance command set is closed:

| Logical command | Operation token |
| --- | --- |
| `operator backup inspect latest` | `backup_inspect_latest` |
| `operator backup create` | `backup_create` |
| `operator restore latest` | `restore_latest` |
| `operator restore-verify latest` | `restore_verify_latest` |
| `operator restore-verify due` | `restore_verify_due` |

Compatibility aliases are not a second public recovery contract. A recovery invocation that is not one of the five canonical logical commands is negative evidence only and MUST fail current conformance routing.

Verified by: TH-HARNESS-AC-044

**TH-HARNESS-REQ-656**
Restore-workbook-probe evidence MUST cite a product owner for the probe behavior. A recovery catalog row MAY use the fixture only when Core 01 or an adopted recovery NLSpec owns the probe and the verification contract cites that owner.

The harness MAY route, schedule, and retain evidence for the probe, but it MUST NOT define workbook query semantics, selected workbook surfaces, source-row eligibility, pass/fail reason codes, or operator error mapping by itself. If no owner-defined probe contract exists for a claimed row, the harness MUST report the row as blocked or unsupported rather than inventing semantics from test fixtures, filenames, or package names.

Verified by: TH-HARNESS-AC-047

**TH-HARNESS-REQ-657**
Network Flow fixture manifests are byte-freeze and execution-input artifacts.
They MUST omit product requirement IDs, acceptance IDs, verification IDs,
copied owner text, and test or phase selectors. Product semantics remain in the
adopted Network Flow, Core, and Graph Projection specifications. The test
catalog routes fixture execution through the Network Flow behavior
verification without copying specification identities into the fixture.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-658**
Network Flow fault controls are harness mechanics for exercising adopted
product-owned commit, worker, cancellation, replay, and recovery behavior. A
fixture or target MAY use `cartulary.test.network_flow_fault_control.v1` only
from a row routed through the Network Flow behavior verification. Fault-control
boundary names, fault kinds, correlation keys, and route responses MUST NOT be
cited as independent product semantics, public API compatibility, Core 05
publication evidence, or performance evidence.
Verified by: TH-HARNESS-AC-050

**TH-HARNESS-REQ-659**
Network Flow clock-control evidence is harness mechanics for exercising adopted
product-owned time behavior. A fixture or target MAY use
`cartulary.test.clock_control.v1` only from a row routed through the Network
Flow behavior verification. Clock route responses, wall-clock mode names,
offsets, and fixed timestamps MUST NOT be cited as independent product
semantics, public API compatibility, Core 05 publication evidence, or
performance evidence.
Verified by: TH-HARNESS-AC-051

**TH-HARNESS-REQ-660**
Network Flow deterministic-randomness evidence is harness mechanics for
exercising adopted product-owned identity, nonce, digest, collision, ordering,
and replay behavior. A fixture or target MAY use
`cartulary.test.network_flow_randomness_control.v1` only from a row routed
through the Network Flow behavior verification. Stream names, value kinds,
collision values, response counts, and fail-closed exhaustion behavior MUST
NOT be cited as independent product semantics, public API compatibility, Core
05 publication evidence, production secret management, or performance
evidence.
Verified by: TH-HARNESS-AC-052

**TH-HARNESS-REQ-661**
Network Flow auth-transition evidence is harness mechanics for exercising
adopted product-owned route authorization, hidden-resource, cursor recheck, and
extension-resource invalidation behavior. A fixture or target MAY use
`cartulary.test.network_flow_auth_transition_control.v1` only from a row
routed through the Network Flow behavior verification. Boundary names,
transition kinds, safe fixture refs, hidden-response kinds, correlation keys,
and route responses MUST NOT be cited as independent product semantics, public
API compatibility, production authorization policy, Core 05 publication
evidence, or performance evidence.
Verified by: TH-HARNESS-AC-053

**TH-HARNESS-REQ-662**
Network Flow audit-assertion evidence is harness mechanics for exercising
adopted product-owned domain audit occurrence counts, transactional audit
boundaries, exact idempotency replay behavior, and no-audit failure cases. A
fixture or target MAY use
`cartulary.test.network_flow_audit_assertion_control.v1` only from a row routed
through the Network Flow behavior verification. Assertion kinds, event codes,
safe fixture refs, baseline counts, expected final counts, replay increments,
correlation keys, and route responses MUST NOT be cited as independent product
semantics, public API compatibility, audit storage design, Core 05 publication
evidence, or performance evidence.
Verified by: TH-HARNESS-AC-054

**TH-HARNESS-REQ-663**
Network Flow generation and fixture integrity MUST be checked by their owning
validators without an acceptance-accounting manifest. Contract-family
validation owns generation lifecycle and generated-output closure; the
generated-drift manifest owns scratch-copy coverage; task-surface validation
owns public targets; fixture validation owns containment, regular-file and
symlink policy, resource bounds, exact sizes, per-file digests, bundle digests,
and frozen revision state. None of those validators may parse a specification,
count acceptance IDs, resolve prose locators, or infer product completeness
from static metadata.
Verified by: TH-HARNESS-AC-055

**TH-HARNESS-REQ-664**
Platform harness runtime owns generic guarded-route authorization, control
contribution registration, reset-hook orchestration, and centralized
redaction. Product- or extension-specific control registries, request
validation, pending state, consume semantics, and product dependency adapters
MUST be owned by the corresponding module in a runtime-scanned package.

A module control contribution MUST provide only its guarded route registrars,
reset hook, and explicitly typed dependency adapters to the generic platform
boundary. Binary composition MAY register that contribution only when test
routes are enabled. Moving a control between implementation packages MUST
preserve its route path, schema ID, guard ordering, disclosure behavior,
redaction, pending-state conflict rules, one-shot consumption, and runtime
reset behavior. Test-support scan exclusions MUST NOT hide a package compiled
into the server binary.

Verified by: TH-HARNESS-AC-050, TH-HARNESS-AC-052, TH-HARNESS-AC-053,
TH-HARNESS-AC-054, TH-HARNESS-AC-056

**TH-HARNESS-REQ-666**
Harness tests and catalog rows MUST NOT carry product requirements, acceptance
criteria, deleted delivery-phase maps, or ledgers, and they MUST NOT redefine
product behavior. Support-only rows and direct package scripts cannot close
product conformance merely because a verification contract or canonical Make
target selects them; retained execution evidence remains subject to human
review against the adopted product specification.
Verified by: TH-HARNESS-AC-013, TH-HARNESS-AC-062, TH-HARNESS-AC-069

**TH-HARNESS-REQ-667**
Catalog ownership is determined in this order: owner of the normative verification postcondition; for a cross-module integration, owner of the externally visible postcondition or primary durable mutation; for mechanism-only evidence, owner of the platform or harness mechanism. Other participants are collaborators. If these rules do not produce one owner, catalog adoption is blocked. Filename, package, runner, and maintainer identity MUST NOT break a tie.
Verified by: TH-HARNESS-AC-062, TH-HARNESS-AC-068

**TH-HARNESS-REQ-668**
Evidence-class gate applicability is closed:

Every active row requires one full-owner `test-slice` partition in addition to the
evidence-class gate below. An explicit selected-subset slice remains valid execution
evidence but cannot close a full-owner audit.

| Evidence class | Required gate family |
| --- | --- |
| `unit` | Owner slice and the unit or fast gate named by the verification contract. |
| `integration` | Owner slice and the integration gate named by the verification contract. |
| `browser` | The Playwright stage's webserver-backed, stateful, or support public gate. |
| `accessibility` | `browser-e2e-a11y`. |
| `visual` | Visual golden digest validation and `browser-e2e-visual`. |
| `measurement` | `browser-e2e-measurement`; informative unless an active claim authorizes publication. |
| `static` | The exact static public target named by the verification contract. |
| `security` | The exact security public target named by the verification contract. |
| `release` | `release-check`. |

Every static or security verification contract MUST name one current public target. An unknown or private target is invalid.
Verified by: TH-HARNESS-AC-069

**TH-HARNESS-REQ-669**
A gate is `not_applicable_zero_rows` only when the selected owner has zero active rows of the corresponding evidence class or Playwright stage. A required row cannot be declared inapplicable by a caller, tracker, guide, or implementation heuristic. Exact applicability MUST be generated from the catalog and verification contracts.
Verified by: TH-HARNESS-AC-069

**TH-HARNESS-REQ-670**
An owner evidence audit MUST require one compatible successful terminal record for every active owner row and every gate derived by TH-HARNESS-REQ-668. A broad passing `check` root does not prove an explicit support, visual, accessibility, or measurement target ran. Missing explicit roots are blockers when such rows exist.
Verified by: TH-HARNESS-AC-066, TH-HARNESS-AC-069

**TH-HARNESS-REQ-671**
`minimum_tier` controls only aggregate row admission through
TH-HARNESS-REQ-800. It does not change owner-slice omission, evidence-class
meaning, claim posture, or release applicability. Generated topology MUST fail
when it omits an admitted row or selects a row above the entry point's tier
without an explicit aggregate policy contract.
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-069

**TH-HARNESS-REQ-672**
Release-readiness aggregation MUST consume current v3 target projections and
owner evidence artifacts only. It MUST preserve owner and evidence class,
reject historical artifact schemas, and keep implementation, informative,
conformance, and Core 05 publication effects distinct.
Verified by: TH-HARNESS-AC-066, TH-HARNESS-AC-070

**TH-HARNESS-REQ-673**
The migration crosswalk is temporary implementation evidence and MUST NOT be a runtime alias, selector source, compatibility reader, or permanent catalog dependency. After its final reconciliation report is retained, the crosswalk is removed; the immutable owner catalog remains the only executable ownership source.
Verified by: TH-HARNESS-AC-068, TH-HARNESS-AC-070

**TH-HARNESS-REQ-674**
R01 through R09 and external research sources in Section 18 are rationale and supporting evidence only. Current behavior is adopted only through the requirements and tables in Sections 1 through 17. An implementation MUST NOT infer a missing default or policy from a research report.
Verified by: TH-HARNESS-AC-016, TH-HARNESS-AC-071

**TH-HARNESS-REQ-675**
An implementation workstream MAY close validation from an accumulated ledger
of retained attempts instead of rerunning every previously passing target after
each source version. Every ledger entry MUST identify the exact source-snapshot
digest, changed authored and generated inputs, directly affected rows and
public targets, retained run root, terminal result, and superseded entry when
applicable. The affected set is the deterministic closure of changed command
contracts, selectors, implementation owners, declared canonical inputs,
generated projections, schemas, fixtures, and producer dependencies; target
name similarity, elapsed time, or an undocumented human assumption MUST NOT
define impact.

For each new source version, every directly affected row and target MUST have a
new successful result before the ledger can close. A prior successful result
MAY carry forward only when its command ID, selector and semantic digest,
execution profiles, evidence contract, toolchain contract, and every input in
its impact closure remain unchanged. A failed or interrupted attempt remains
retained and invalidates only its affected entries; it MUST NOT revoke an
unaffected passing entry, and it MUST NOT be replaced by another attempt at the
same source version. A later source version MAY supersede that failure only
when the ledger records the relevant change and a passing rerun of its complete
affected closure.

The combined result set closes only when every required current row, public
target, schema/generation gate, security freshness boundary, lifecycle check,
and cleanup obligation resolves exactly once to the newest applicable passing
entry. Missing, duplicate, conflicting, unattributed, or provenance-mismatched
entries block progression. This rule is acceptance-evidence composition only:
it does not authorize scheduler reuse, skip selected work inside an invocation,
or relabel one attempt as another source version. Performance observations
remain governed by TH-HARNESS-REQ-377: samples from different source versions
MUST NOT be combined inside one target window, although independently
qualified target windows may retain their own exact source provenance.
Verified by: TH-HARNESS-AC-066, TH-HARNESS-AC-071, TH-HARNESS-AC-081

## 17. Acceptance Criteria / Definition of Done

The acceptance matrix is the harness Definition of Done: it defines what must
be true at the observable boundary. It is not a mandatory one-row-per-test,
one-row-per-verification, or machine completeness map. Implementations MAY
combine compatible scenarios in one test or use multiple tests for one
scenario, provided the resulting behavior evidence is unambiguous.

Owner-slice acceptance under TH-HARNESS-AC-063 through TH-HARNESS-AC-065 MUST cover omitted and exact-row plans, complete owner and service-backed dependency scopes, and target-wide checks whose closure remains limited to resolved rows. Negative fixtures MUST cover missing and unknown owners, blank tokens, normalized duplicates, cross-owner IDs, inactive rows, unmapped rows, zero-row owners, and a non-service-backed row requested by `service-backed-test-slice`. Each rejected selection MUST exit `2` before setup or child launch. The retained plan, retained scheduler summary, and target-local JSON scheduler summary MUST contain identical selection identity, and every selected-subset fixture MUST prove that row rollup is not treated as full-owner completion.

Cutover acceptance MUST execute in this order:

1. Review specification ownership, internal references, and open decisions
   editorially; executable tools do not inspect specification content.
2. Run `make lint-markdown` as documentation maintenance only.
3. Run schema and negative-fixture validation through `make json-shape-check`.
4. Run `make generated-artifact-policy-check`.
5. Run `make harness-contract` and inspect the public command/input surface.
6. Regenerate only through `make generate`, then run `make generate-drift`.
7. Run focused owner slices and catalog, selector, accounting, and audit fixtures.
8. Run every gate generated by TH-HARNESS-REQ-668, including frontend, browser, accessibility, visual, measurement, security, and release row evidence.
9. Run `make agent-finalize` without retained-run maintenance when no eligible full warm root exists.
10. Produce one successful full warm `make check` root, run `make agent-finalize RESULTS_DIR=<that-exact-root>`, repeat the broad verification affected by finalization, and run `make release-check`.
11. Confirm zero active v1 targets, inputs, schema readers, catalog readers, artifact writers, generated drift, and unresolved migration items.

A later step at the same source version MUST NOT compensate for an earlier
failure. Across source versions, TH-HARNESS-REQ-675 controls exact affected-test
reruns, unaffected pass carry-forward, failure supersession, and combined
closure. Every historical failure remains visible in the accumulated ledger.

| ID                | Requirement owner  | Scope                            | Setup fixture                                                                | Invocation                                                                                | Expected exit/status                                           | Stdout                                                 | Stderr                                                       | Required artifacts                                                                                 | Negative case                                                                | Cleanup expectation                                     |
| ----------------- | ------------------ | -------------------------------- | ---------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- | -------------------------------------------------------------- | ------------------------------------------------------ | ------------------------------------------------------------ | -------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- | ------------------------------------------------------- |
| TH-HARNESS-AC-000 | Section 8          | Schema validation                | Any public target that emits required JSON                                   | Target named by the fixture                                                               | Success only if JSON validates                                 | Per Section 7                                          | Per Section 7                                                | Every emitted required JSON artifact validates against Section 8 schema attachments                | Inject schema-invalid required summary                                       | No extra cleanup beyond target contract                 |
| TH-HARNESS-AC-001 | Sections 1, 4      | Command registry                 | Current tree                                                                 | `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all` plus registry parity checker    | `0` when registry matches exactly                              | Bounded report                                         | Empty on success                                             | Public target registry parity report                                                               | Extra/missing public target fails                                            | none                                                    |
| TH-HARNESS-AC-002 | Section 5          | Config precedence                | Fixture target with CLI, Make var, env, manifest, config, default candidates | Dedicated config resolver test target or unit harness                                     | `0`                                                            | Machine or bounded summary                             | Empty on success                                             | Resolver summary showing CLI > Make var > env > manifest > config file > default                   | Non-positive scheduler limit exits `2` with `configuration_error`            | no child work                                           |
| TH-HARNESS-AC-003 | Sections 5, 6      | Result root, run ID, prepared identity, and private suite-runtime boundary | No child work required                                                       | Invalid result root, invalid run ID, unsafe custom result root, complete/partial prepared identity, runtime containment, symlink, ownership, and permission fixtures | `2` for invalid or partial identity; `0` for complete prepared reuse | Empty or failure JSON according to target output class | Bounded config diagnostic                                    | Failure summary when wrapper starts; retained root and private runtime preflight reject unsafe permissions, containment, symlinks, and partial prepared identity before private-file use | Slash, backslash, whitespace, `.`, `..`, existing non-empty unprepared run dir, world-writable custom root, repository/result-contained runtime, symlink traversal, permissive runtime root, partial prepared tuple, non-`1` marker, and target mismatch all fail | no child work and no artifact creation for rejected identity |
| TH-HARNESS-AC-004 | Sections 7, 8      | Machine output accepted          | Toolchain ready; explicit result root/run ID                                 | `CARTULARY_OUTPUT_MODE=machine make backend-unit`; `... make test-fast`; `... make check` | Target status                                                  | Exactly one JSON object plus LF                        | Empty after wrapper starts                                   | `cartulary.tool_run_summary.v5` and target artifacts                                               | Progress prose or duplicate JSON fails                                       | normal target cleanup                                   |
| TH-HARNESS-AC-005 | Section 7          | Machine output rejected          | No child work                                                                | `CARTULARY_OUTPUT_MODE=machine make clean`; `... make dev`; `... make help`               | `2`                                                            | Empty                                                  | Bounded `usage_error` diagnostic                             | None required                                                                                      | Child work starts despite rejection                                          | no deletion or service start                            |
| TH-HARNESS-AC-006 | Section 10         | Scheduler determinism            | Controlled manifest with simultaneous child completions and scheduled browser groups | Run scheduler fixture twice with same manifest; validate generated browser worker-admin slot ranges | `0`                                                            | Bounded summary or machine object                      | Empty on success                                             | Byte-identical scheduler events after dynamic timestamp normalization allowed only by schema rules; browser group worker slots are explicit, contiguous, and non-overlapping | Event sequence differs; browser worker slot env is missing or overlaps       | finalizers run                                          |
| TH-HARNESS-AC-007 | Section 11         | Service modes and eager resource lifetime | Owned and attach fixtures, object-store mutation-probe faults, dedicated and migration database finalizers, active-connection and drop-failure injection, interruption, stale recovery, and package-scoped object-store reuse | Owned service target; attach target missing one required var; controlled owner slice before/after; focused normal-drop, forced-fallback, cleanup-failure, interruption, stale-recovery, bucket, and prefix fixtures | owned success; attach failure `2`; readiness expiry `3`; zero successful per-test databases at suite teardown; peak live databases reduced by at least 75%; comparable wall time no more than 5% slower | Bounded normalized stage, cause, attempt, timeout, cleanup, lane-state, peak-live, and exact success/failure count summary | Empty on owned success; config, readiness, ownership, or cleanup diagnostic | Owned lease before child work; ordinary exact-owner deletion before bounded connection termination; failed cleanup remains in the private ledger; same-lane object probe succeeds with zero residue; package bucket reuse remains active | Attach mode deletes unrelated resources, successful per-test resources survive to terminal sweep, forced cleanup runs without exact ownership or an ordinary attempt, capability rejection polls, readiness expiry replaces lane, raw service data is retained, or package-scoped object-store reuse is removed | owner finalizers close runtimes, jobs, pools, handles, buckets, and prefixes before eager deletion; suite teardown and stale recovery remain idempotent fallback |
| TH-HARNESS-AC-008 | Section 12         | Test-only harness routes         | Browser test runtime with test route token and saved-view fixture inputs     | Runtime identity, retired reset-path not-found, saved-view fixture success, auth rejection, and origin/host rejection fixtures | Expected HTTP statuses from Section 12 | HTTP JSON response | n/a | Identity validates schema; reset path is absent; saved-view response is a normal system-scoped resource; no permissive CORS | Default runtime exposes a test route, reset endpoint exists, wrong host/origin reaches mutation, product auth bypasses token, saved-view accepts owned fields, or wildcard CORS appears | owned backend terminated on fixture completion |
| TH-HARNESS-AC-009 | Section 13         | Cleanup and destructive reset guard | Synthetic registry with safe and unsafe paths; fake Compose, database, migration, and object-store commands | Cleanup guard unit; `CARTULARY_CLEANUP_DRY_RUN=1 make clean`; dry-run and missing-confirmation invocations for `services-down`, `db-reset`, and `object-store-reset` | `0` for safe dry-run; nonzero for unsafe synthetic path or missing destructive confirmation | Dry-run lines match format                             | Bounded guard or confirmation diagnostic before mutation      | Candidate list, guard evidence, and command-shape evidence for confirmed local resets                              | Empty path, `/`, `.`, `..`, traversal, protected root, outside-repo path, symlink-following, inherited-env-only destructive confirmation, object-store reset touching another bucket, or `services-down` removing volumes accepted | no deletion, service start, or service stop in dry-run  |
| TH-HARNESS-AC-010 | Section 13         | Stale janitor proof gates        | Fake DB, bucket, container, and browser fixtures with/without proof          | Focused stale-janitor tests                                                               | `0`                                                            | Bounded summary                                        | Empty on success                                             | Evidence that unproven resources retained and proven stale fixtures deleted only outside dry-run   | Resource lacking generated name/proof deleted                                | unproven resources retained                             |
| TH-HARNESS-AC-011 | Section 15         | Redaction and secret-free retention | Fake DSN, object-store secret, token, header, cookie, CLI arg, nested JSON, structural session fields, private key, forbidden filename, symlink, injected-value, early-delete, permission, and process-crash fixtures | Redaction unit, wrapper capture, private-runtime lifecycle, and terminal retained-root scan | `0`; redaction/write/scan failure exits `11` unless Section 9.1 preserves an earlier primary failure | No unredacted secret in machine JSON                   | No unredacted secret in captured stderr                      | Summaries and bounded tails contain required redaction tokens; retained roots contain no secret-capable filename, symlink, or injected value; private files and roots are removed after cleanup | Any runtime file is retained, any injected value or secret syntax survives, required structural fields are redacted, cleanup removes an unproven root, or a retained file is group/world-readable | exact private files and directories removed; no physical-media sanitization claim |
| TH-HARNESS-AC-012 | Section 14         | Platform matrix                  | Platform claim checker fixture                                               | Platform matrix checker                                                                   | `0` for allowed profiles; nonzero for unsupported claim        | Bounded summary                                        | Diagnostic on unsupported claim                              | Matrix report                                                                                      | macOS/Windows-native/Podman claimed as current conformance                   | none                                                    |
| TH-HARNESS-AC-013 | Sections 9, 16     | Product versus harness failure   | One known failing assertion, one harness setup failure, one object-store readiness timeout, and one browser strict-port conflict fixture | Canonical test target under each fixture                                                  | Product failure exits `10`; setup failure exits Section 9 code; readiness timeout exits `3`; strict-port conflict retains its existing resource-conflict exit | Failure headline names class and reason                | Bounded diagnostic                                           | Service scope, target/tool, owner, scheduler, and run summaries preserve the exact normalized failure and reference the authoritative diagnostic | Setup failure classified as product, readiness timeout collapsed to fixture error, or strict-port conflict classified as `service_start_error` | harness cleanup attempted                               |
| TH-HARNESS-AC-014 | Section 9          | Exit-code matrix                 | Controlled failure fixtures, including Playwright assertion failure, per-test timeout, interrupted test, selector-accounting failure, nonzero child/report contradiction, mixed product/accounting failure, and outer watchdog expiry | Exit matrix test target                                                                   | Exact Section 9 code for every class; Playwright assertion and per-test timeout exit `10`, selector accounting and child/report contradiction exit `11`, and outer watchdog expiry exits `13` | Per output mode                                        | Per output mode                                              | Failure summaries with primary failure selection preserve the exact row, group, scheduler, target, and aggregate class/reason | Cleanup failure overrides earlier product failure; a Playwright per-test timeout is classified as harness timeout; an outer watchdog is classified as product | cleanup failure recorded but primary exit preserved     |
| TH-HARNESS-AC-015 | Sections 6, 8      | Retained artifact identity       | Explicit result root/run ID plus generated default identity fixtures for public node-tool and owner-slice targets | `CARTULARY_TEST_RESULTS_DIR=<dir> CARTULARY_TEST_RUN_ID=<id> make backend-unit`; direct generated-ID public targets | `0`                                                            | Summary names run root                                 | Empty                                                        | Artifacts under one `<dir>/<id>` with target, run ID, run root, invocation marker, terminal summary, and passing retained-secret scan; retained run roots and target dirs are owner-only on POSIX hosts; external suite runtime is absent after exit | Preflight marker and summary use sibling generated IDs, newest-run fallback is accepted as proof, retained directories are group/world-accessible, or a secret-capable runtime file exists below the run root | custom absolute result root not removed by `make clean`; owned external suite runtime removed |
| TH-HARNESS-AC-016 | Sections 1, 2, 18, 19 | Editorial and boundary closure | Revised document                                                             | Human owner review; `make lint-markdown` checks Markdown quality only                      | Review complete; Markdown lint `0`                             | Existing Markdown-lint output                         | Existing Markdown-lint diagnostic                            | Human review records owner conflicts or open decisions; no executable artifact consumes this document | A specification conflict is silently resolved by a machine projection or Markdown lint is cited as product conformance | none                                                    |
| TH-HARNESS-AC-017 | Section 11         | Lifecycle-machine conformance    | Service-suite fixtures for happy path, startup failure, interrupted child, cleanup failure, illegal transition, and crash/rerun | Lifecycle-machine conformance target or unit harness                                      | Happy path `0`; failure fixtures use exact Section 9 code      | Bounded summary or machine object                      | Empty on happy path; bounded diagnostic on failure fixture | `cartulary.test_services.lifecycle.v2` stream with sequential events, valid transitions, terminal state, Section 9 failure mapping, and cleanup proof behavior | Unlisted `(state,event)` mutates state, terminal state accepts later event, or lifecycle stream validates with a sequence gap | normal suite cleanup; unproven resources retained       |
| TH-HARNESS-AC-018 | Sections 4, 10, 11 | Warm graph health | Retained warm-ready `check` fixture plus cold-provisioning, measurement-quietness, holder, contamination, cache, and event-closure fixtures | `make scheduler-summary-timing-drift RESULTS_DIR=<dir> TARGET=check` | Success only for warm-eligible canonical evidence with exact interval, holder, lease, cache, row, and target closure | Bounded summary | Bounded diagnostic on failure fixture | Run manifest, unit events, run summary, and `check` target summary identify readiness, queue, holders, leases, cache accounting, and projection closure | Hidden provisioning, measurement overlap, missing holder, contaminated lease, unexplained cache state, or incomplete event projection passes unnoticed | none |
| TH-HARNESS-AC-019 | Section 8.2        | Agent finalizer                  | Fake Make fixture plus valid, missing, failed, incomplete, contaminated, non-warm retained run, action-cache hit/miss/disabled/corrupt/input-change/output-change fixtures | `make agent-finalize`; `RESULTS_DIR=<dir> make agent-finalize`; `CARTULARY_OUTPUT_MODE=machine make agent-finalize` | Success for coherent maintenance inputs; fail-fast for first failed action substep or invalid retained run; cache hits only for eligible closed-profile actions | One `[FINALIZE]` line then bounded result/artifact lines; machine emits one JSON object | Bounded failure diagnostic naming failed action/substep | `agent-finalize/finalize-summary.json`, per-action `execution_state` and cache state, child summaries/logs when executed, and `finalize_summary` artifact ref | Excluded targets run, mutation starts after invalid `RESULTS_DIR`, cache hit bypasses retained-run validation, corrupt cache produces success, machine output requires log parsing, semantic action IDs are absent, or skipped-after-failure work is absent | No cleanup or destructive command is run               |
| TH-HARNESS-AC-020 | Section 4          | Public target semantic value     | Current target registry plus one synthetic shallow-wrapper fixture           | Registry semantic-value checker                                                           | Success only when every public target declares at least one semantic behavior and every declared behavior has an owner section | Bounded report                                         | Empty on success                                             | Semantic-value parity report                                                                    | Target with only child command aliases and no semantic behavior passes        | none                                                    |
| TH-HARNESS-AC-021 | Sections 5, 8, 10 | Frontend-unit harness stability | Current graph plus delayed jsdom workbook-row, inspector-subject, row-history selector, controlled-input helper, worker-limit, and Vitest failure-sidecar fixtures | Topology contract tests; `make frontend-unit`; capacity-override graph fixture | Success when Vitest workers and graph claims match, shared helpers tolerate bounded hydration, inspector readiness uses stable identity, and sidecar diagnostics are preferred | Bounded summary | Empty on success | Work graph records the worker effect in CPU claims; frontend helper and selector-policy tests pass; failure sidecar remains linked when runner JSON exists | Runner workers exceed claims, helper waits are unbounded, identity is unstable, or sidecar evidence is discarded | none |
| TH-HARNESS-AC-022 | Sections 1, 4, 5, 8, 10 | Frontend readiness harness metadata | Active frontend catalog rows plus missing accessibility target, raw-only artifact, missing visual fixture, stale accounting, support-specimen substitution, concept-image substitution, cross-owner selection, and removed-preflight fixtures | Task-surface report, catalog/JSON validation, `explain-test-owner`, accessibility/visual fixture checks, and evidence-accounting validation | Success only when active owner rows, exact selectors, profiles, semantic fixture support, normalized accessibility summary, and selected-row accounting closure are valid | Bounded report or explanation | Empty on success; bounded configuration diagnostic | Public registry includes `browser-e2e-a11y`; catalog owns frontend rows; semantic fixture metadata owns capture scope; `visual.fixture.default_timeline_workbook_shell` closes only from its exact active catalog scenario and app-owned screenshot; support specimens and concept images cannot close it; v4 accessibility and v1 evidence-accounting schemas validate | Filename/title inference, cross-owner closure, stale accounting, support substitution, raw third-party output, or removed preflight target passes | no child browser work required for metadata fixtures |
| TH-HARNESS-AC-023 | Sections 4, 7      | Registry output-class and side-effect parity | Current NLSpec registry, `tools/task_surface_manifest.json`, rendered task-surface report, and output-class matrix | Registry parity checker and malformed registry fixtures | Success only when target membership, output class, artifact policy, schema policy, and side effects agree exactly | Bounded parity report | Empty on success; bounded drift diagnostic on failure | Public registry parity report with side-effect and output-class comparison | Missing `side_effects[]`, `none` plus another class, undeclared side effect, or output-class drift passes | none |
| TH-HARNESS-AC-024 | Section 10         | Scheduler command-shape closure | One live generated scheduler fixture for each required command type plus malformed command descriptors, prerequisite-policy fixtures, and artifact-consumer dependency fixtures, with optional command types validated when present | Scheduler manifest schema and shape checks | Success only for closed command shapes, required and closed `make_prerequisite_policy` values for `make_target`, and declared artifact producer/consumer DAG edges | Bounded summary | Empty on success; bounded shape diagnostic on failure | Shape-check evidence naming command type, required fields, forbidden fields, wrong-type failures, prerequisite policy, and artifact dependency parity | Missing required field, forbidden field, unknown command type, wrong field type, omitted or invalid prerequisite policy, or missing build-artifact producer edge passes | no child work |
| TH-HARNESS-AC-025 | Section 8          | Schema attachment closure | Every Section 8 `present` schema attachment plus positive and negative fixtures | Schema attachment policy check and schema fixture validation | Success only when every present attachment exists, parses, is top-level closed or extension-container closed, and validates fixtures | Bounded summary | Empty on success; bounded schema diagnostic on failure | Schema attachment report | Missing schema, malformed schema, open top-level schema without extension container, or fixture-blind schema passes | none |
| TH-HARNESS-AC-026 | Sections 1, 16     | Core test-route traceability | Core 04, verification contracts, and active catalog rows | Requirement-ID uniqueness check and `REQ-04-109` citation classifier | Success only when `REQ-04-109` means test-only runtime-control route security and public-origin behavior cites `REQ-04-110` | Bounded traceability report | Empty on success; bounded traceability diagnostic on failure | Duplicate-ID report and citation classification report | Public route, WebSocket origin, evidence-handle, or deployment-origin row cites `REQ-04-109` | none |
| TH-HARNESS-AC-027 | Section 4 | Public registry source-of-truth parity | Current public Make surface, task-surface owner/projection, work-graph owner, generated Make include, aggregate policies, and prose registry | Registry source-of-truth parity checker | Success only when every public target is owner-backed, all 98 public names are present, the five retired internal targets are absent, and tier/policy selections are graph-reachable | Bounded parity report | Empty on success; bounded drift diagnostic on failure | Public-source parity report with minimum-tier and aggregate-policy projections | A public target exists in only one projection, a retired target remains live, or policy selection is unreachable | none |
| TH-HARNESS-AC-028 | Sections 4, 5, 8, 13 | Local cache profiles | Cache helper fixture plus representative readiness, build, finalizer, and render-cache fixtures | Cache-helper smoke tests; cold/hot readiness and build target runs; `make agent-finalize`; render drift checks; cleanup fixture | Success only when first run misses, second run hits, disable/force/corrupt/missing-output cases execute or fail safely, summaries remain emitted, and scheduler accounting reports no undeclared reuse | Bounded summary | Empty on success; bounded diagnostic on invalid cache state | Cache records validate against Section 8 schemas; run-root cache artifacts show state/reason/record path; public summaries remain present | Security, drift, service readiness, runtime reset, cleanup, destructive guard, browser/service-backed live-state work, or aggregate success is accepted by cache reuse; missing output succeeds by reuse; cache artifact is cited as product evidence | `make clean` preserves default cache roots; `make distclean` removes default cache roots |
| TH-HARNESS-AC-029 | Sections 1, 5     | Public input matrix closure | Current NLSpec input matrix and task-surface metadata | Input-contract parity checker | Success only when every public target input in task-surface metadata appears in the NLSpec matrix and every NLSpec matrix row mirrors metadata or an explicitly documented NLSpec default override | Bounded parity report | Empty on success; bounded drift diagnostic on mismatch | Public input matrix parity report | Public target accepts an input absent from the NLSpec matrix, or NLSpec row omits default, bound, empty-string, invalid, summary, or forwarding behavior | none |
| TH-HARNESS-AC-030 | Section 10        | Scheduler defaults and auto policies | Fixture graphs for each resource and auto policy, plus override and impossible-resource fixtures | Scheduler resource-resolution tests | Success only when every declared default, override bound, auto formula, omission rule, and impossible-resource failure matches Section 10 | Bounded summary | Empty on success; bounded diagnostic on mismatch | Capability snapshot, override source, and resolved-limit evidence in the run manifest | `auto` resolves differently, override above max passes, omission lacks declared behavior, or impossible resources start child work | no child work beyond fixture scheduler |
| TH-HARNESS-AC-031 | Section 8 | Canonical diagnostic artifact closure | Positive and negative run-manifest, unit-event, run-summary, target-summary, fixture-lease, and cache-record fixtures | Schema/artifact policy checker and canonical evidence validation | Success only when every canonical artifact validates, unit intervals and resource holders close, and all old scheduler/pressure schemas are rejected as current input | Bounded report | Empty on success; bounded diagnostic on mismatch | Artifact-policy and canonical-evidence closure report | Missing or open canonical fields, broken interval union, absent holder, invalid lease/cache record, or legacy artifact acceptance passes | none |
| TH-HARNESS-AC-032 | Section 9         | Primary failure determinism | Simultaneous failure fixtures covering class, lifecycle, event, target, path, and reason ties | Exit matrix and primary-failure unit tests | Success only when selected primary failure and public exit follow Section 9.1 and TH-HARNESS-REQ-304 exactly | Bounded summary | Bounded diagnostic on mismatch | Failure summary with ordered candidate failures and selected primary | Cleanup overrides earlier non-cleanup failure, or tie order differs across runs | cleanup failure recorded when fixture creates one |
| TH-HARNESS-AC-033 | Section 11        | Concurrent lifecycle | Service-suite lifecycle fixtures with overlapping child work, duplicate child start, unknown child finish, and interruption | Lifecycle-machine conformance target or unit harness | Success only when active-child counts transition legally and illegal duplicate/unknown events fail closed | Bounded summary or machine object | Empty on happy path; bounded diagnostic for illegal fixtures | Lifecycle stream with `active_child_count`, child keys, legal transitions, and terminal state | Concurrent child start is rejected, unknown finish mutates state, duplicate start passes, or active count becomes negative | normal suite cleanup; unproven resources retained |
| TH-HARNESS-AC-034 | Sections 10, 12 | Browser reset lifecycle closure | Leading, middle, and incompatible-affinity selections; mutable tables; both migration metadata tables; bootstrap/idempotency/object/browser state; conflicting table lock; serving/exclusive admission; active request drain; backend-stop, pre/post-commit, object, restart, and cleanup faults; armed in-memory controls | Graph compiler, typed Recovery reset, broker quarantine, backend-replacement lifecycle, schema/redaction/classification tests, Network Flow owner slice, and bounded saturated probe | A first selected group has no reset; a selected predecessor/successor has exactly one derived reset; exclusive admission begins only after proven backend termination; one transaction resets persistent state; replacement generation is ready with the same fingerprint; every injected failure quarantines and prevents product work | `cartulary.test.database_reset_diagnostic.v1`, `cartulary.browser_reset_attempt.v1`, backend-generation evidence, and fixture lease v4 | Bounded redacted diagnostic; reset and product attempts remain one | Sorted table/count proofs preserve `goose_db_version` and `schema_migration_lineage`; stage and normalized SQLSTATE are exact; failed reset exits `3` as fixture/operational evidence; cleanup outcome is terminal | Leading reset, affinity crossover, mutation before stop, serving/exclusive overlap, blind retry, raw PostgreSQL detail leak, stale in-memory control, old backend generation, missing evidence, product assertion classification, or reusable tainted lease passes | Every failed reset ends quarantined/destroyed; ten saturated cycles have no reset failure, retry, unclassified failure, or leaked lease |
| TH-HARNESS-AC-035 | Section 12        | Test-route edge closure | Weak token, malformed token, missing/wrong header, pending-fault conflict, consumed-fault retry, and retired reset-route fixture | Test-only route unit/integration tests | Expected HTTP status or startup/config failure for every Section 12 token and fault edge case; the retired reset path is not found | HTTP JSON response where route starts | n/a | Route guard failures use `test_route_forbidden`; duplicate pending faults use `test_public_error_fault_already_armed`; startup configuration failures use `configuration_error` | Weak token starts, product auth bypasses token, second fault replaces pending fault, consumed fault remains armed, or runtime reset route exists | process replacement owns browser reset state |
| TH-HARNESS-AC-036 | Sections 13, 15   | Cleanup and redaction closure | Protected-root cleanup fixtures, cleanup-owned child paths, structured secret keys, raw secret text, and structural field names | Cleanup guard and redaction tests | Success only when protected-root attempts fail, cleanup-owned paths succeed or no-op when missing, exact key/raw-text redaction applies, and schema-owned structures are preserved | Bounded summary | Empty on success; bounded diagnostic on mismatch | Cleanup guard report and redacted summary/log fixtures | Protected root deletion passes, missing cleanup-owned path fails, secret leaks, or structural fields are over-redacted | no deletion outside cleanup-owned fixtures |
| TH-HARNESS-AC-037 | Sections 5, 8, 10 | Operator runtime binary injection | Current topology plus missing, non-executable, digest-mismatch, undeclared-input, and raw-Go fallback fixtures | `make build-operator`; scheduler-selected operator Go work; `make check OPERATOR_BIN=/tmp/x`; `make check CARTULARY_OPERATOR_BIN=/tmp/x` | Producer succeeds with declared output path; consumer fails before product assertions for invalid injection; undeclared public inputs exit `2`; digest mismatch exits `11` | Bounded summary | Empty on success; bounded config/artifact diagnostic on mismatch | `build-operator` tool summary and build-artifact cache artifact; operator aggregate `runtime-binaries.json`; Go runner logs contain no nested `make build-operator` for scheduler-selected operator work | Hidden nested operator builds, arbitrary child forwarding, missing producer dependency, missing runtime-binary provenance, or binary cache hit marked as scheduler `reused` passes | no extra cleanup beyond build output contract |
| TH-HARNESS-AC-038 | Section 4.1A      | Helper ownership registry | Current helper registry plus allowed facade and forbidden catalog/execution/accounting/scheduler-internal import fixtures | Harness import-boundary contract tests | Success only when every helper family is classified exactly once and no unclassified private catalog, execution, accounting, backend, browser, scheduler, duration, diagnostics, generated-artifact, or finalization helper is imported by non-owner code | Bounded report | Empty on success; bounded import-boundary diagnostic on mismatch | Import-boundary report naming helper family, facade key, source, and target | Non-owner code imports private helpers after a facade exists, or a family is unclassified or duplicated | none |
| TH-HARNESS-AC-039 | Section 4.1A      | Compatibility paths | Current tree plus semantic-boundary and unknown-owner fixtures | Caller inventory, import-boundary contract tests, and relevant public target characterization | Success only when historical backend, frontend, scheduler, and execution paths are absent as compatibility paths, no temporary redirects are accepted, current facades are owned exactly once, semantic boundary rules reject bypasses, generated/task-surface metadata does not reference removed paths, and public Make behavior remains unchanged | Bounded report | Empty on success; bounded semantic ownership diagnostic on mismatch | Authored helper-owner validation and import-boundary report naming current facade key, source, and target | A forwarding shim is added without a demonstrated consumer, a facade is missing or duplicated, unknown owner roots or cross-owner private imports pass, removed paths return in generated metadata, or public behavior changes after deletion | none |
| TH-HARNESS-AC-040 | Section 4.1A      | Govulncheck findings ownership | Govulncheck JSON stream fixtures covering no findings, package/module findings, symbol findings, redaction, and malformed JSON | Static-analysis security findings tests and `go-vulncheck` when toolchain is ready | Success only when static-analysis/security ownership proves identical normalized findings, redaction, exit mapping, and artifact behavior after helper movement | Bounded summary | Empty on success; bounded parse/security diagnostic on failure | `govulncheck-findings.json` validates `cartulary.govulncheck_findings.v1` with deterministic finding order | Backend path remains a supported import, redaction changes, symbol findings stop blocking, or exit mapping drifts | temp files removed |
| TH-HARNESS-AC-041 | Section 4.1A      | Migration/schema validator ownership | Migration history, schema object ownership, scratch migration, and JSON-shape fixtures | `json-shape-check`, `migration-drift`, and database-contract drift tests | Success only when `json-shape-check` and `migration-drift` retain manifest schema validation, scratch apply behavior, diagnostics, and failure classification after helper movement | Bounded summary | Empty on success; bounded drift diagnostic on failure | Migration/schema manifest validation summaries and migration drift retained artifacts | Schema IDs change, diagnostics drift unexpectedly, scratch DB cleanup changes, or validators become backend execution behavior owners | scratch DB cleanup per migration-drift contract |
| TH-HARNESS-AC-042 | Section 4.1A      | Duration retained-run safety | Coverage, drift, update, and finalizer retained-run fixtures including failed, partial, missing-artifact, stale, contaminated, ambiguous, valid drift, and valid full warm check roots | Duration baseline and finalizer tests | Success only when coverage is read-only, drift is read-only, update rejects invalid retained evidence before mutation, and `agent-finalize` accepts only valid retained full warm `make check` roots for mutating duration refresh | Bounded summary | Empty on success; bounded retained-run diagnostic on failure | Baseline files unchanged for invalid evidence; finalizer summary records retained-run validation before mutation | Mutating update starts before retained-run validation, partial run is accepted for finalizer refresh, invalid evidence writes baselines, or drift mutates files | invalid fixtures leave tracked baselines unchanged |
| TH-HARNESS-AC-043 | Section 16        | Recovery evidence routing | Core recovery verification contracts plus legacy `deployment_admin` negative fixtures | Catalog evidence-routing validation and affected recovery target planning | Success only when recovery CLI invocation is `deployment_admin`-irrelevant and legacy gated tests are negative-only and cannot close AC-402, AC-427, or AC-428 | Bounded report | Empty on success; bounded routing diagnostic | Report names row IDs, verification IDs, evidence class, and negative status | A gated recovery CLI test counts as product conformance or stale status closes a current row | none |
| TH-HARNESS-AC-044 | Section 16        | Canonical operator command mapping | Operator scenario rows, runtime-binary rows, and compatibility-alias fixtures | `make build-operator` plus scheduler-selected operator work or command-mapping validation | Success only when implementation-owned executable behavior maps to exactly the five Core logical commands, compatibility aliases are negative-only, and Core-owned final stdout, optional progress, timeout/default, target-config, and exit-code behavior validate | Bounded summary or machine result | Empty on success; bounded command-mapping diagnostic on mismatch | Operator runtime-binary provenance and retained operator result/progress artifacts for selected work | Old command names, compatibility aliases, or alternate JSON envelopes pass as a second public contract | no extra cleanup beyond build output contract |
| TH-HARNESS-AC-045 | Section 4.1A      | Migration-evidence classification | Migration manifest, embedded SQL source, goose ledger, schema-object ownership, and recovery catalog fixtures | `json-shape-check`, database-contract drift validation, and catalog classification checks | Success only when migration-history evidence routes through database-contract or migration-evidence ownership and cannot close operator-recovery conformance | Bounded report | Empty on success; bounded classification diagnostic | Migration/schema summaries and catalog-row owner classification | A migration-evidence operator wrapper counts as AC-428 recovery proof | none |
| TH-HARNESS-AC-046 | Section 4.1A      | Moved-test and support-package accounting | Test-path movement fixtures covering catalog, task-surface, topology, generated schedules, runtime-binary profiles, and the registered `processtest` and `suiteservices` raw support-package bindings | JSON/catalog validation, task-surface report, generated drift, `make backend-process`, and `make backend-unit` | Success only when owner inputs change before generated artifacts, outputs regenerate through Make, `processtest` runs as self-contained `managed_process` work under `backend-process`, and `suiteservices` runs as service-free work under `backend-unit` | Bounded report | Empty on success; bounded accounting diagnostic | Owner-input diff plus drift/schema summaries and ordinary raw-aggregate target summaries | A moved test hand-edits generated output, changes profiles/scheduling without NLSpec coverage, omits either support package from its public aggregate, or introduces a hidden product-service prerequisite | none |
| TH-HARNESS-AC-047 | Section 16        | Restore-workbook-probe owner routing | Recovery probe catalog rows with owner-cited and owner-missing fixtures | Verification routing and affected owner planning | Success only when probe evidence cites Core 01 or an adopted recovery owner and harness fixtures do not define semantics | Bounded report | Empty on success; bounded owner-routing diagnostic | Probe routing report with verification ID or blocked status | Fixture, filename, or package inference closes recovery proof | none |
| TH-HARNESS-AC-048 | Section 8 | Same-run unit deduplication closure | Union-graph fixtures with duplicate producers, digest mismatch, old-run cache, and missing output | Graph compiler, cache fixtures, `explain-run`, and `make json-shape-check` | Success only when identical semantic units execute once, every target projection names the shared unit, digest conflicts fail before work, and prior-run bytes never satisfy `same_run` | Bounded run summary and `explain-run` lines | Bounded graph/cache diagnostic on mismatch | Canonical unit events and target summaries prove exact-once execution and projection membership | A helper reruns, a conflicting digest merges, an old run is reused as same-run, or selected rows disappear | none |
| TH-HARNESS-AC-049 | Sections 8, 11, 16 | Network Flow fixture manifests | Positive frozen fixture manifest and scenario with committed source, expected, and transcript bytes plus negative fixtures for forbidden provenance fields, unsorted paths, missing files, symlink/traversal paths, digest mismatch, and draft selection | Network Flow fixture/scenario validators, schema validation, `make json-shape-check`, and the Network Flow behavior tests that execute fixtures | Success only when every selected manifest and scenario validates, hashes match exact bytes, committed fixture roots are not mutated, run-local materialization is used, frozen-only selection is enforced, and no acceptance, verification, requirement, selector, or specification-provenance field is accepted | Bounded summary naming fixture IDs and bundle digests | Bounded schema/artifact diagnostic on mismatch | `cartulary.network_flow_fixture_manifest.v2` and `cartulary.network_flow_fixture_scenario.v2` validation evidence plus retained execution summary with manifest SHA-256, source and expected bundle digests, materialized input root, produced artifact refs, and comparison status | Fixture identity inferred from an untyped filename, unlisted file accepted, committed bytes mutated, forbidden provenance accepted, draft fixture selected as current evidence, or digest/order/path mismatch reaches product code | Run-local materialization removed by result-root cleanup; committed fixture bytes retained |
| TH-HARNESS-AC-050 | Sections 12, 16    | Network Flow fault controls | Disabled route, token/host/origin edge cases, invalid boundary/kind/correlation/error-code bodies, pending conflict, exact boundary/correlation consume, registry clear, process replacement, and ownerless selector fixtures | Network Flow fault-control route tests, registry tests, backend-replacement reset test, schema validation, and owner fixtures | Route guard and mutation rules remain closed; registry clear is exact; armed state is absent from the replacement process; evidence is owner-routed | HTTP JSON response and bounded fixture/reset evidence | Bounded error envelope on mismatch | Route response plus fixture transcript and reset backend-generation proof | Authorization bypass, pending replacement, incorrect consume, state surviving replacement, or control-only product closure | failed reset quarantines allocation |
| TH-HARNESS-AC-051 | Sections 12, 16    | Test clock controls | Disabled route, token/host/origin edge cases, fixed/offset/reset/state payloads, process replacement, and ownerless time selectors | Test clock route tests, backend-replacement reset test, schema validation, and owner fixtures | Route behavior remains closed; explicit clock reset restores wall mode; replacement backend begins in wall mode; evidence is owner-routed | HTTP JSON response and bounded fixture/reset evidence | Bounded error envelope on mismatch | Clock response plus fixture transcript and replacement backend proof | Authorization bypass, state mutation by read, invalid mutation, fixed clock in replacement, or control-only product closure | failed reset quarantines allocation |
| TH-HARNESS-AC-052 | Sections 12, 16    | Network Flow deterministic randomness controls | Route edges, invalid stream/value cases, collision/exhaustion, registry clear, process replacement, and ownerless selectors | Route/registry tests, backend-replacement reset test, schema validation, and owner fixtures | Ordered values and fail-closed exhaustion hold; secrets never leak; registered streams are absent from the replacement backend | HTTP JSON response and bounded fixture/reset evidence | Bounded error envelope on mismatch | Randomness response, transcript, and replacement backend proof | Value leak, replacement of pending stream, fallback randomness, state surviving replacement, or control-only product closure | failed reset quarantines allocation |
| TH-HARNESS-AC-053 | Sections 12, 16    | Network Flow auth-transition controls | Route edges, matching/correlation/duplicate cases, hidden-response safety, registry clear, process replacement, and ownerless selectors | Route/registry tests, backend-replacement reset test, schema validation, and owner fixtures | Exact matching and hidden-resource rules hold; registered transitions are absent from the replacement backend | HTTP JSON response and bounded fixture/reset evidence | Bounded error envelope on mismatch | Auth-transition response, transcript, and replacement backend proof | Incorrect consume, replacement of pending tuple, disclosure, state surviving replacement, or control-only product closure | failed reset quarantines allocation |
| TH-HARNESS-AC-054 | Sections 12, 16    | Network Flow audit assertion controls | Route edges, exact/zero/replay counts, correlation/duplicate cases, registry clear, process replacement, and ownerless selectors | Route/registry tests, backend-replacement reset test, schema validation, and owner fixtures | Count and matching rules hold; payloads remain redacted; registered assertions are absent from the replacement backend | HTTP JSON response and bounded fixture/reset evidence | Bounded error envelope on mismatch | Audit-assertion response, transcript, and replacement backend proof | Incorrect consume/count, pending replacement, disclosure, state surviving replacement, or control-only product closure | failed reset quarantines allocation |
| TH-HARNESS-AC-055 | Sections 8, 16     | Network Flow generation and fixture integrity | Contract-family registry, generated outputs, generated-drift scratch manifest, task surface, fixture manifests, and scenario files, plus negative cases for invalid lifecycle, missing generated output, missing scratch input, missing public target, forbidden provenance, and unsafe fixture path | Contract-family, generated-artifact, task-surface, fixture/scenario, and generated-drift validators; `make json-shape-check`; `make harness-contract`; `make generate-drift` | Success only when each owner validates its own closed input, the active Network Flow family has no activation dependency, official generation reproduces outputs, scratch replay contains required typed inputs, public targets resolve, and fixture integrity rejects unsafe or forbidden metadata without parsing specifications | Bounded owner-specific summaries | Bounded schema/artifact diagnostic on mismatch | Retained JSON-shape, harness-contract, generated-policy, and generate-drift summaries | Static acceptance accounting, specification parsing, missing typed input, unsafe path, private raw target, or hand-authored generated output closes evidence | no child product work beyond the separately routed behavior tests |
| TH-HARNESS-AC-056 | Sections 4, 11, 12, 15, 16 | Test-support ownership and explicit policy | Registered shared and owner-local support roots, unknown-root fixtures, in-process server mode fixtures, fixture-policy agreement fixtures, security-profile rendering fixtures, and module control contributions | Support-inventory validator, service-backed guard tests, server-helper tests, PostgreSQL helper tests, static-analysis wrapper tests, backend boundary check, and affected product-control route tests | Success only when support roots are owner-named and registered exactly once, phase-shaped active helpers are absent, route mode and database policy are explicit, manifest and call-site fixture policy agree, every support root is security-scanned, runtime-compiled controls stay in runtime scans, and module contributions preserve guarded behavior | Bounded ownership/policy summary | Bounded configuration, boundary, or security diagnostic | Validated test-support inventory plus ordinary target summaries from the affected checks | Unknown/duplicate support root, zero-value route mode, fixture-policy mismatch, unscanned support root, phase-shaped active helper, or runtime control hidden by support exclusions passes | no cleanup beyond the selected test/service contract |
| TH-HARNESS-AC-057 | Section 4.1B | Runtime-binary and private-runner closure | Production and harness server profiles, runtime-binary registry fixtures, invalid harness-only environment, missing/nonregular/nonexecutable/digest-mismatched injection, nested-build attempts, legacy runner aliases, and static shell-import fixtures | Server build-profile tests, runtime-binary validator, harness import-boundary checks, and private runner usage tests | Success only when the repository exposes exactly the three deployable identities, the harness server remains a non-deployable build profile, production rejects harness-only inputs, every black-box row receives an exact validated scheduler artifact, and every legacy or contextless runner path fails closed | Bounded build and contract summaries | Bounded configuration, boundary, or usage diagnostic | Runtime-binary registry validation, build summaries, injection provenance, and import-boundary report | A fourth deployable appears, production consumes a harness-only input, a nested build fallback executes, an injected binary lacks identity proof, or a legacy alias succeeds | build outputs and run-local injected binaries follow their target cleanup contracts |
| TH-HARNESS-AC-058 | Sections 4, 5, 8 | Task-surface ownership and generated Make density | Authored task-surface owner, execution topology, generated task-surface projection, shared Make runtime, thin Make bindings, and synthetic-growth fixtures | Task-surface owner/schema validation, public registry parity, generation drift, density validation, and public wrapper characterization | Success only when public command metadata has one authored machine owner, topology cannot redefine it, generated v15 parity holds, unsupported `alias` profiles and per-variable origin transport are rejected, every size/line/growth budget passes, and public target behavior is unchanged | Bounded ownership and density report | Bounded configuration or generated-drift diagnostic | Owner/projection digest relationship plus byte, line-length, repeated-expansion, and synthetic-growth metrics | A generated projection becomes an owner, topology embeds task metadata, dense global input plumbing returns, budgets regress, or public behavior changes | generated scratch outputs removed |
| TH-HARNESS-AC-059 | Sections 10, 17 | Public-target performance acceptance | Documented host/capability context, unchanged semantic window, one forced-cold observation, one discarded warm-up, five measured observations, and matched-sixth instability fixtures | Canonical selected target for every run plus retained-run inspection and external process-wall measurement | Success only for a stable matched window with material p50 improvement beyond variability, no p90 regression, predicted structural reduction, and exact target/row/evidence/artifact/cleanup parity | Existing bounded target output | Existing bounded failure output; contamination is classified separately | Canonical cold, warm-up, and measured run roots with exact identity and structural metrics | An unstable, failed, stale, mismatched, retried, reduced-coverage, or mixed-semantics run enters the window or refresh occurs before acceptance | ordinary target cleanup succeeds for every run |
| TH-HARNESS-AC-060 | Sections 10, 16 | Pure Go execution-family consolidation | Backend-unit catalog rows plus compatible, incompatible, duplicate-symbol, over-selection, missing-symbol, partial-output, and attribution fixtures | Target-plan validation, Go runner fixtures, `make backend-unit`, exact owner slices, and generated drift | Success only when compatible exact-symbol rows share execution without changing identity/routing and every symbol is proven exactly once | Existing bounded target summary | Bounded row/symbol or artifact diagnostic | Per-row evidence plus before/after execution-family count | A row drops, an unowned test selects, raw/exact selectors merge, incompatible profiles merge, or a shim remains | ordinary target cleanup |
| TH-HARNESS-AC-061 | Sections 10, 16 | Deterministic service Go batching | Service-backed catalog rows and duration owners plus compatible, incompatible, isolated, oversized, partial-output, exact-failure, and growth fixtures | Shard-plan smoke, scheduler matrix, owner slices, Go runners, profile agreement, generated drift, and `make check` timing | Success only when compatible items share deterministic shards without changing row/scenario/owner coverage, profiles, attribution, artifacts, or cleanup | Existing bounded shard and target summaries | Bounded planning, resource, row/symbol, or artifact diagnostic | Shard metadata, process counts, plan digest, profile parity, and current evidence | Packing crosses a boundary, selects unowned work, loses a row, weakens a claim, accepts partial output, or is nondeterministic | ordinary target and service cleanup |
| TH-HARNESS-AC-062 | Sections 3, 8 | Owner, verification, and identifier closure | Valid registries plus missing, extra, null, duplicate, unordered, unknown, Unicode, overlength, and phase-token fixtures | Catalog schema and reference validator | `0` only for a closed, unique, fully resolved active catalog | Bounded catalog summary | Bounded configuration diagnostic | Catalog and verification semantic digests plus schema summaries | Unknown field, unresolved reference, zero-row owner, recycled ID, or invalid ordering passes | no child work |
| TH-HARNESS-AC-063 | Sections 3, 10 | Runner selectors and profiles | Exact Go, Vitest, Playwright, and shell selectors plus zero, multiple, overlap, traversal, symlink, glob, regex, and unknown-profile fixtures | `test-catalog-check` selector/profile validation | `0` only when every case resolves exactly once and every profile is compatible | Bounded catalog summary | Bounded selector/profile diagnostic | Selector-resolution and profile-resolution report | Arbitrary executable, ambiguous selector, inline resource override, or symlink escape passes | no child work |
| TH-HARNESS-AC-064 | Sections 4, 5, 7 | Owner command inputs and output | Missing/valid owners; omitted/blank/duplicate/cross-owner rows; worker `0`, `1`, `16`, `17`; JSON omitted, empty, `1`, invalid; retired inputs | Public command contract fixtures for all five owner commands | Exact Section 5 success or exit `2` before setup | Human output or exact command JSON plus LF | Bounded usage diagnostic | Plan, explanation, guide, or audit artifacts exactly where required | Default-check narrows omitted rows, JSON and machine combine, or old input is ignored | no setup for rejected inputs |
| TH-HARNESS-AC-065 | Sections 9, 10 | Terminal row accounting | Pass, assertion, setup, dependency, cancellation, authorized/expired skip, duplicate/missing record, concurrent failure, and cleanup fixtures | Owner-slice scheduler matrix | Success only for passed or authorized-skipped rows; exact Section 9 exits otherwise | Bounded scheduler summary | Bounded primary diagnostic | One terminal record per resolved row, all attempts, finalizer evidence | Unauthorized skip passes, cleanup masks primary, retry hides failure, or selected row lacks record | finalizers always run |
| TH-HARNESS-AC-066 | Sections 6, 8, 16 | Evidence compatibility and freshness | Matching and mismatched source/catalog/verification/profile digests; missing, duplicate, extra, mixed, old, dirty, and unsafe roots | `test-evidence-audit` fixtures | `0` only for one compatible complete owner evidence set | Bounded audit result | Bounded usage or artifact diagnostic | `cartulary.harness_evidence_root_manifest.v1` plus canonical row, unit, run, and target closure with used/unused/rejected roots | TTL, newest fallback, mixed snapshot, duplicate candidate, or broad-check inference passes | retained inputs unchanged |
| TH-HARNESS-AC-067 | Sections 3, 6, 15 | Executable-input and semantic boundaries | Neutral direct/joined restricted-root fixtures plus product-phase, execution-step, delivery-phase, and ambiguous fixtures | Executable-input policy and semantic identity scan | `0` only when executable sources and machine evidence contain no restricted input | Bounded policy summary | Bounded boundary or configuration diagnostic | Closed policy validation summary | Validator consumes a restricted root, machine evidence contains a documentation path, or a live selector retains delivery identity | no document inspection or mutation |
| TH-HARNESS-AC-068 | Sections 3, 4, 16 | Migration reconciliation | Frozen 456 backend, 87 frontend, and 5 Graph identities plus duplicate, missing, consolidation, deletion, and new-row fixtures | Baseline/crosswalk validator and final totals report | `0` only when all 548 identities have one terminal disposition and every new row has authorization | Bounded totals by source, disposition, and owner | Bounded reconciliation diagnostic | Baseline digest, crosswalk digest, assertion-preservation and owner-review evidence | Graph row omitted, old identity repeated, unauthorized row added, or deletion lacks owner proof | temporary crosswalk removed only after report retention |
| TH-HARNESS-AC-069 | Sections 3, 16 | Evidence-to-gate applicability | Owners with each evidence class, zero-row classes, informative measurement, claim measurement, and unknown target fixtures | Generated applicability matrix and owner audit | `0` only when every active row maps to exact required gates | Bounded applicability summary | Bounded routing diagnostic | Generated owner/evidence/gate matrix | Human applicability skip, informative evidence closes release, or private target satisfies gate | none |
| TH-HARNESS-AC-070 | Sections 1, 4, 15, 16 | Atomic predecessor retirement | Current tree plus old target, variable, schema, artifact, reader, alias, dual-write, and delivery-identity fixtures | Task-surface parity, semantic scan, schema attachment validation, and repository reference scan | `0` only when no superseded live surface remains | Bounded retirement summary | Bounded compatibility diagnostic | Removed-identity report and current registry parity | Any predecessor public command, reader, fallback, phase catalog, or ledger remains live | generated outputs regenerated through Make |
| TH-HARNESS-AC-071 | Sections 1-17 | Harness v3 parity and adoption | Revised NLSpec, schemas, catalog, topology, task surface, generated outputs, focused owner evidence, warm check, release evidence, and accumulated validation ledger | Editorial lint, `json-shape-check`, generated-policy/drift, harness contract, owner slices, `agent-finalize`, affected aggregate gates, and `release-check` | Success only when one exact TH-HARNESS-REQ-675 ledger closes every current requirement from newest affected passes and unchanged carried-forward passes | Existing bounded summaries | Exact failing target or ledger-closure diagnostic | Current v3 schemas, task surface, catalog, topology, canonical run roots, finalizer, release artifacts, source versions, impact sets, and supersession records | Partial adoption, missing schema, generated drift, unresolved requirement, same-version retry, undocumented impact omission, or incompatible historical evidence closes a gate | ordinary target cleanup succeeds |
| TH-HARNESS-AC-072 | Sections 2, 4 | Observability ownership and coverage | Complete public task surface; required, excluded, out-of-scope, omission, overlap, unknown, duplicate-sequence, unowned-reason, parameterized-profile, and application-boundary fixtures | Task-surface validation, harness contract, and OTel conformance | Every public target has exactly one explicit disposition, every required target has one canonical measurement profile, and every public testing entry point is required | Bounded coverage summary | Bounded owner or boundary diagnostic | Authored policy and generated projection digests | Default or target-name inference, omission, overlap, ownerless disposition, unstable input profile, duplicate sequence occurrence, or application scope widening passes | none |
| TH-HARNESS-AC-073 | Sections 4, 8 | Canonical evidence lifecycle | Direct, aggregate, external-root, source-change, dirty, failure, interruption, contamination, tamper, and partial-generation fixtures | Run-manifest capture, canonical event finalization, and explicit evidence audit | One canonical run manifest, event stream, run summary, and complete target projection set per top-level invocation; artifacts validate independently of the checkout | Bounded result and artifact refs | Bounded artifact diagnostic | Canonical run and target artifacts with source/system/graph digests | Current checkout changes historical identity, absolute path leaks, normal result changes, tampered evidence is rewritten, or partial output passes explicit validation | selected retained run byte-for-byte unchanged by the check |
| TH-HARNESS-AC-074 | Sections 8, 10 | Work graph and timing accounting | Direct/aggregate graphs, every wait reason, overlapping units, blockers, services, runners, clock skew, finalizers, failure, and interruption fixtures | Golden event-stream reconstruction and schema checks | Exact dependencies, interval unions, queue waits, resource blocking, critical path, and unattributed time match expected values | Bounded run summary | Bounded graph/accounting diagnostic | Canonical unit events, run summary, and target summaries | Temporal containment invents parentage, overlapping intervals are summed, dependency edges disappear, or direct and aggregate projections diverge | scratch output removed |
| TH-HARNESS-AC-075 | Sections 8, 10, 15 | OTLP shape and privacy | Valid bundle plus hostile paths, commands, environment, credentials, SQL-like text, symbols, output, and error strings | OTLP decoding, allowlist validation, and redaction scan | Trace and metric payloads decode, match native identity, and contain only registered names/attributes | Bounded shape summary | Bounded privacy or shape diagnostic | OTLP trace/metric payload digests | Forbidden literal or unknown attribute reaches payload | scratch capture removed |
| TH-HARNESS-AC-076 | Sections 2, 4, 15 | Explicit exporter containment | Disabled ordinary runs, hostile `OTEL_*`, exact and ambiguous selection, valid HTTPS/loopback endpoints, invalid URLs, header permissions/shapes, redirect, timeout, and receiver failure | Fake collector and process-network fixtures | Ordinary runs make no telemetry request; explicit export sends exact payloads; configuration exits `2`; delivery exits `1` | Bounded count and endpoint class | Bounded configuration or exporter diagnostic with secrets absent | Fake collector request digests | Newest-run fallback, inherited env egress, redirect follow, timeout drift, header leak, failure-code collapse, or source mutation passes | fake collector stopped; source run unchanged |
| TH-HARNESS-AC-077 | Sections 8, 10 | Graph and browser scheduling | Serial, parallel, simultaneous failure, dependency failure, interruption, generic resource contention, cycles, unknown dependencies, release dependency, and one/two-stack capability fixtures | Unified scheduler matrix, generated topology, broker, and browser lifecycle tests | DAG order, deterministic event bytes and failure order, capacity, summary projection, process-group cancellation, resets, isolated leaves, finalizers, and cleanup match the adopted contract | Bounded run/browser summaries | Bounded scheduler diagnostic | Canonical unit events, run/target summaries, fixture leases, and browser group results | Shell owns scheduler policy, release dependency starts early, more than declared stacks overlap, sibling leaks, or direct leaf shares a stack | all started stacks cleaned |
| TH-HARNESS-AC-078 | Sections 9, 10 | Backend and finalizer optimization parity | Compatible/incompatible direct compatibility keys, raw and isolated selectors, missing/extra/duplicate/partial Go JSON, capture and worker exceptions, concurrent failure, and emission fixtures | Backend target plan, runner fixtures, backend-unit, and artifact comparison | Process grouping and bounded concurrency reduce work while every row, partial success, artifact, and primary failure remains exact | Existing bounded target summary | Bounded row, artifact, or failure diagnostic | Before/after process and interval-union finalizer accounting | Family-level inference merges rows, raw selector merges, partial success drops, artifact order changes, or race masks failure | ordinary target cleanup |
| TH-HARNESS-AC-079 | Sections 4, 8, 10, 17 | Public-target performance acceptance | Target/provider windows with one forced-cold root, one discarded warm-up, and five or conditionally six measured roots; p50, nearest-rank p90, MAD, variability-band and leave-one-out stability; canonical non-overlapping timing; ordering, duplicate, failed, cold, retried, dirty, provenance mismatch, missing-command, exact reviewed policy transition, required-improvement targets, 47-row public closure including `openapi-compatibility-check`, separate internal diagnostics, and independent check-window fixtures | Baseline writer, performance checker, retained-run inspection, canonical unit-event accounting, and task-surface owner closure | Exact Section 10.5 formulas and cardinality pass for every required public command, required improvement clears the variability band without p90 regression, every other p90 is non-regressing, all material intervals are singly attributed, and the public roster/count/bindings/rows/sum close exactly | Bounded performance summary | `duration_baseline_drift`, exit `13`, with every rejected-root reason | Verified root refs and bindings, timing sources, policy projections, samples, p50, p90, deviations, variability bands, structural metrics, roster arithmetic, and verdicts | Hand-edited baseline, inferred/newest root, mismatched or failed run, dispatch-only timing, multiply attributed intervals, opaque policy change, blanket percentage, archival-reader translation, or moved required work passes | retained inputs unchanged; baseline writes only after complete validation |
| TH-HARNESS-AC-080 | Sections 5, 9, 17 | Versioned focus-continuity observation | Delayed mention mutation whose focused control unmounts, whose semantic Timeline row is initially offscreen, and whose projection delivers `N-1`, then response version `N`; instrumented postcondition helper | Focus-continuity model and React fixtures plus the exact live browser row through its canonical owner slice | One attempt succeeds only after the stable source row reaches at least `N` and the application restores the row-local focus target; helper interaction counters remain zero | Existing bounded unit/browser summary | Product assertion diagnostic with expected/rendered versions, target presence, active-element identity, mounted row IDs, and scroll geometry | Ordered lifecycle generations, response source identity/version, rendered version observations, and focus/viewport observation | `N-1` settles continuity, the helper focuses or scrolls after mutation, user interruption still permits focus stealing, or a later pass replaces the failed invocation | ordinary target cleanup; historical failed evidence remains unchanged |
| TH-HARNESS-AC-081 | Sections 8, 17 | Incremental accumulated validation | At least three source versions containing an unaffected passing target, an affected passing target, a failed affected attempt, a relevant repair, an unrelated repair, a selector/profile change, and malformed impact/provenance entries | Build and validate the accumulated workstream ledger; execute only each version's deterministic affected closure | Success only when all current requirements resolve exactly once to the newest applicable pass, unaffected passes carry forward, the repaired affected failure is superseded by a later-version pass, and unrelated or same-version attempts cannot erase a failure | Bounded closure summary by source version and carried/rerun/superseded count | Bounded impact, provenance, duplicate, missing, conflict, or same-version-retry diagnostic | Source-snapshot roster, changed-input digests, affected row/target closure, retained root refs, pass provenance, failure history, and supersession edges | Filename heuristic omits impact, changed contract carries forward, failed same-version result is retried away, historical bytes are relabeled, performance samples cross source within one window, or missing current work closes | retained attempts remain immutable; no scheduler or cache reuse is inferred |
| TH-HARNESS-AC-082 | Section 2.1 | Tier selection | Complete catalog plus invalid, duplicate, and redundancy fixtures | Cutover disposition and tier-reachability validation | Every active row has one minimum tier and reaches its declared aggregate union | Bounded tier counts | Exact row/tier diagnostic | Tier report and evidence-union closure | Boolean fallback, missing tier, or silent row removal passes | no child work |
| TH-HARNESS-AC-083 | Section 2.1 | Canonical graph identity | Equivalent direct, aggregate, leaf, and owner-slice selectors plus semantic mutations | Compile and compare graphs | Equivalent selections have identical units/digests and changed semantics change the digest | Bounded graph summary | Graph validation diagnostic | Canonical graph and digest report | Duplicate execution, cycle, or invalid claim reaches child work | scratch graph removed |
| TH-HARNESS-AC-084 | Section 2.1 | Unified scheduling | Contention, starvation, failure, cancellation, and simultaneous-completion fixtures | Deterministic scheduler simulation and process tests | Rank, fit, backfill, aging, failure propagation, and cancellation match REQ-802 | Bounded scheduler summary | Scheduler diagnostic | Unit events and capability snapshot | Unheld reservation, starvation, or leaked process passes | owned work cleaned |
| TH-HARNESS-AC-085 | Section 2.1 | Capacity closure | CPU, memory, process, IO, port, service, missing-data, and override fixtures | Capability resolver validation | Resolved capacity is the safe multidimensional minimum and every override is declared/recorded | Bounded capability summary | Configuration diagnostic | Capability snapshot and sources | CPU-only inference or undeclared override passes | no child work |
| TH-HARNESS-AC-086 | Section 2.1 | Fixture broker | Transaction, dedicated, migration, object, process, browser, contamination, and borrowed-resource fixtures | Broker lifecycle, object-store put/head/delete admission, and affected owner slices | Explicit leases preserve isolation; object probes and package probes leave zero residue; owned cleanup does not close borrowed resources | Bounded lease and safe object-stage summary | Fixture/cleanup diagnostic, with earlier readiness failure retained as primary | Lease, probe-attempt, cleanup, and lifecycle events | Implicit clone fallback, contamination, probe residue, post-readiness lane replacement, or borrowed deletion passes | unhealthy leases destroyed; exact package bucket re-probed |
| TH-HARNESS-AC-087 | Section 2.1 | Go grouping | Compatible, incompatible, oversized, raw, exact-symbol, and missing-output fixtures | Deterministic shard planning and Go JSON reconciliation | LPT plans are capacity-derived and every selected row resolves exactly once | Bounded shard summary | Row/shard diagnostic | Plan digest and row outcomes | Fixed packing, incompatible merge, or missing row passes | ordinary test cleanup |
| TH-HARNESS-AC-088 | Section 2.1 | Browser group scheduling | Parallel isolated, stateful, reset, measurement, visual-read/update, failed-lane, and cancellation fixtures | Browser graph and lifecycle tests | Groups share only declared state, conflicting writes never overlap, and every lane cleans | Bounded browser summary | Lifecycle diagnostic | Group events, leases, and target projections | Batch loop, stage lock, reset drift, or lane leak passes | all lanes cleaned |
| TH-HARNESS-AC-089 | Section 2.1 | Aggregate and cache closure | Five aggregate graphs plus hit, miss, cold, off, mutation, corruption, and stale-security fixtures | Aggregate graph/cache validation | No phase barrier or nested scheduler remains; units deduplicate and reuse only closed work | Bounded aggregate/cache summary | Cache or graph diagnostic | Current-run unit and target evidence | Stateful reuse, stale finding, or duplicate unit passes | cache scratch proof-gated |
| TH-HARNESS-AC-090 | Section 2.1 | Canonical retained artifacts | Direct and aggregate runs plus old, malformed, overlapping, and unattributed fixtures | Schema and event-projection validation | Event unions close material wall time once and all projections agree | Bounded evidence summary | Artifact/accounting diagnostic | V3 run manifest, events, run summary, and target summaries | Old reader, dual write, dispatch timing, or duplicate attribution passes | retained v3 artifacts follow cleanup policy |
| TH-HARNESS-AC-091 | Section 2.1 | Atomic v3 cutover | Current tree plus old target, variable, schema, reader, writer, alias, and wrapper fixtures | Task-surface, source-reference, schema, generation, focused, aggregate, and release gates | Public names remain, removed internal surfaces are absent, changed IDs version once, and no compatibility path survives | Existing bounded summaries | Exact compatibility diagnostic | V3 owner/projection and validation ledger | Partial cutover, alias, dual reader/writer, or fixed global timing gate passes | generated outputs refreshed through Make |
| TH-HARNESS-AC-092 | Sections 4.1A, 4.5, 5, 13, 14 | Exact bootstrap and durable machine-state readiness | Clean checkout without Node or frontend dependencies; controlled `HOME`/`XDG_CACHE_HOME`; exact or older Go launcher; valid, absent, overlapping, repo-contained, read-only, capacity-exhausted, and legacy `/tmp` cache fixtures; failed and successful tool installs; read-only cleanup trees and external symlinks | Scratch-only public-wrapper, cleanup, fake Go, isolated cache/install, foundational-schema, and pin-drift validation | Public preflight obtains pinned Node first; Node bootstrap summary validates without `node_modules`; defaults resolve outside `/tmp` and the repository; diagnose is read-only; ensure passes exact `GOCACHE`, `GOMODCACHE`, and `GOTMPDIR`; cleanup removes only owned paths and handles read-only descendants; exact pin selection and failure-atomic install hold | Bounded readiness or cleanup summary | Bounded configuration or resource-conflict diagnostic with exact path, filesystem capacity, or repair scope | Global-input projection, generated validator drift, pin projection, normalized paths, and isolated filesystem observations | Bootstrap depends on frontend AJV, global inputs are implementation-only, literal `/tmp` fallback survives, paths overlap or enter the repo, doctor mutates, symlink target changes, unrelated `.cache` content is removed, `ENOSPC` is unknown, corruption reaches child work, or failed install removes the old executable | isolated scratch removed; borrowed machine state and symlink targets unchanged |
| TH-HARNESS-AC-093 | Sections 4, 5 | Negative-fixture determinism | Prewarmed successful graph-cache entry followed by a fake failing tool; empty and valid artifact fixtures behind ordinary and phony producer prerequisites; repeated and reordered invocations | Owner-controlled lint-shell and release-task-surface smoke fixtures | Success only when the nested failing tool executes with cache reuse disabled at its public invocation, every injected artifact reaches validation byte-identically without producer execution, and warm or reordered execution cannot change the fixture verdict | Existing bounded smoke summary | Exact fake-tool, cache-mode, artifact-mutation, or validation diagnostic | Fake-tool invocation log, nested run manifest, seeded artifact bytes, and ordinary target summaries | Cached success bypasses the fake tool, producer regeneration replaces an injected artifact, parent-only cache control is treated as sufficient, prerequisite enumeration substitutes for artifact isolation, or repeated execution changes the verdict | isolated scratch removed; shared production cache and generated artifacts unchanged |
| TH-HARNESS-AC-094 | Sections 2.1, 9 | Disposable targeted migration capability | Harness-issued migration scratch databases; arbitrary database/source construction attempts; apply targets `-1`, `0`, and positive versions; rollback targets `-1`, `0`, and positive versions; preparation success/failure/cleanup | pgtest capability unit and service-backed fixtures plus affected source-owner slices | Only the opaque harness-issued capability performs canonical-source targeted execution; invalid targets fail before source/database access; preparation events exactly describe the outcome | Existing bounded row and lease summaries | Exact capability, target-validation, fixture, or cleanup diagnostic | Migration lease identity, preparation lifecycle events, and selected row outcomes | A free production helper survives, an arbitrary handle/source is accepted, invalid input touches source/database state, duplicate status conflicts with events, or borrowed state is closed | owned scratch database destroyed; borrowed database unchanged |
| TH-HARNESS-AC-095 | Sections 2.1, 9 | Production DDL Rebaseline v2 isolation and residue | Pristine, contaminated, prerequisite, lineage, purpose, role, ACL/default, recycled-connection, profile-claim, and rollback-through-zero PostgreSQL 16 fixtures | Database Migrations, PostgreSQL, Recovery, pgtest, testservices, dev-stack, and owner-routed unit/service-backed slices | Only versions 1..29 apply; incompatible state fails before v2 DDL; exact roles and purpose credentials are isolated; runtime and Recovery positive/negative operations match the object manifest; rollback residue is exact; physical extension state does not claim a profile | Existing bounded row, run, and lease summaries | Closed migration, binding, role, ACL, prerequisite, claim-state, or residue diagnostic | Manifest parity, PostgreSQL catalog facts, role identity, allow/deny/default matrices, remediation object, and cleanup evidence | Legacy SQL, compatibility credential, wrong extension, contamination, mixed role, excess privilege, incomplete Recovery, claimed-by-table profile, or undeclared rollback residue passes | owned scratch database destroyed; borrowed database unchanged |

| TH-HARNESS-AC-096 | Sections 2.1, 3.2 | Populated fixture profile and key closure | Valid v2 registry and catalog plus a synthetic second profile and missing, unknown, inactive, duplicate, incompatible, divergent, malformed-digest, unsupported-version, contract-drift, and reordered fixtures | Registry, catalog, schema, generator, direct-plan, aggregate-plan, and generated Go/JavaScript key-vector validation | Every bound row resolves its explicit profile; every route resolves the same raw key and builder identity; a second profile renders and routes without an AC-043 generic branch; invalid routing fails before child work | Bounded profile/key summary | Exact registry, row, field, digest, version, or compatibility diagnostic | Registry digest, defensive generated descriptor, key vector, generated profile group, graph unit, and target-plan projection | Filename inference, default profile, handwritten profile mirror, prefixed raw key, duplicated binding, route divergence, unsupported current version, or cross-language mismatch passes | no child work; scratch registry removed |
| TH-HARNESS-AC-097 | Section 2.1 | Source-owner contribution closure | Registered contribution graph plus missing, duplicate, reordered, cyclic, incompatible, authorization, descriptor mutation, entropy, session, and redaction fixtures | Direct contribution and assembler contracts plus one canonical production snapshot build and read-back | The selected profile adapter proves its exact source-owner counts and conditions, production invariants, inactive sessions, stable semantic digest, and safe structural receipts without an independent repeated full build | Bounded contribution receipts | Owner-qualified contribution, descriptor, or validation diagnostic | Ordered structural receipts, selected profile descriptor, semantic validation digest, and owner-backed retired-row mapping | A generic schema embeds profile facts, raw browser SQL, cross-owner loader, secret field, unstable ID, redundant production build, or partial seal passes | partial database and secret bundle removed |
| TH-HARNESS-AC-098 | Sections 2.1, 5, 9 | Snapshot builder, clone, private runtime, and cleanup lifecycle | Same-key concurrency, different-profile and different-key, bounded four-clone lifecycle, connection race, corruption, partial build, clone failure, cancellation, finalizer failure, credential cleanup, stale-resource, ownership-marker, permission, symlink, and early-delete fixtures | Graph, broker, testservices, PostgreSQL support, browser lifecycle, generated descriptor, private-runtime boundary, and janitor tests | One shared production builder per profile and key precedes quiet sessions; each nonempty Timeline measurement closure records construction count one, a non-measurement closure records zero, clones are isolated, and all proved owned resources clean idempotently on every terminal path | Bounded build/lease/cleanup summary | Builder, descriptor, construction-count, private lease, isolation, or cleanup diagnostic | Active-policy immutable build artifact, separate build diagnostic, opaque retained lease evidence, lifecycle events, cleanup proof, and absent external runtime root | Generic lifecycle code branches on AC-043, lifecycle tests construct production volume, builder enters quiet session, live fallback, unknown connection termination, clone aliasing, early private-file deletion, retained private handle, or unproven janitor deletion passes | zero owned database, bucket, runtime-bundle, private file, session, process, or directory residue |
| TH-HARNESS-AC-099 | Sections 2.1, 8, 15 | Immutable post-cleanup measurement provenance | Current-policy observations plus valid, missing, stale, inconsistent, unredacted, superseded-generation-only, digest-tampered, overlap, cleanup-failed, secret-bearing, forbidden-path, and injected-value fixtures | Row and target finalizers, schema validation, retained-root scan, direct and aggregate measurement routing | Every current summary links immutable observation, build, and retained cleanup-evidence bytes after private cleanup; independent profile groups prove builder and clone closure, exact policy, zero overlap, cleanup, redaction, and secret-free retention | Bounded active-generation qualification summary | Evidence-integrity, environment, cleanup, compatibility, or security diagnostic | Active registry policy, observation, build, opaque lease evidence, summary, aggregate, digested references, and retained-secret scan | Embedded observation qualifies as a summary, referenced artifact mutates, historical evidence translates or qualifies, cleanup is claimed early, a private lease is retained, or secret material/value is retained | secret copies, private runtime tree, and owned resources removed before retained lease finalization |
| TH-HARNESS-AC-100 | Sections 2.1, 4, 11 | Explicit local service sessions | Missing, malformed, stale, expired, wrong-owner, wrong-digest, dead-container, symlink, concurrent attach, live-borrower, repeated-down, and owned-only target fixtures | Session lifecycle commands, task-surface validation, graph runner, testservices, and one small attached service-backed owner slice | Attachment is explicit and redacted; borrowers and run resources are unique; down refuses live borrowers and deletes only exact proven session resources; owned-only targets reject attach before setup | Redacted typed session status | Usage, configuration, ownership, expiry, borrower, or cleanup diagnostic | Session descriptor, borrower leases, per-run service scope, lifecycle events, and exact cleanup proof | Descriptor presence auto-attaches, a caller sets internal active state, CI/release borrows, unrelated containers are removed, secrets print, or a live borrower loses its session | Run-owned resources clean after each borrower; session resources clean only after exclusive down |
| TH-HARNESS-AC-101 | Sections 2.1, 8, 10 | Deterministic cached release inventory | Cold and warm paired-producer runs plus source, manifest, lockfile, package, toolchain, scanner, schema, container, generator, partial, surplus, corruption, traversal, link, mode, producer, destination, concurrency, and missing-output fixtures | Release inventory generator and validators, graph compiler, cache-v2 contract fixtures, schema validation, and downstream release-readiness consumption | The real cold producer executes once and publishes one deterministic current pair; a warm hit invokes no producer scanner, restores both files transactionally, and both non-cacheable consumers emit fresh evidence | Existing bounded release summaries | Exact generation, validation, schema, cache, or artifact diagnostic | CycloneDX SBOM, `cartulary.license_report.v2`, cache-v2 record with the exact typed pair, and fresh validator unit evidence | Either consumer regenerates, canonical bytes contain volatile identity, v1 is accepted, the output set is incomplete, an unsafe cache entry publishes bytes, or downstream readiness consumes stale/unvalidated inventory | temporary scanner work is removed; invalid cache entry quarantined; canonical pair follows release-artifact cleanup policy |

### 17.1 Requirement-to-Acceptance Traceability

| Requirement range         | Owner section                      | Acceptance criteria                                     |
| ------------------------- | ---------------------------------- | ------------------------------------------------------- |
| `TH-HARNESS-REQ-001..049` | Status, scope, authority, owner model | TH-HARNESS-AC-013, TH-HARNESS-AC-015, TH-HARNESS-AC-016, TH-HARNESS-AC-026, TH-HARNESS-AC-029, TH-HARNESS-AC-062, TH-HARNESS-AC-063, TH-HARNESS-AC-066, TH-HARNESS-AC-071, TH-HARNESS-AC-072 |
| `TH-HARNESS-REQ-050..099` | Public command surface             | TH-HARNESS-AC-001, TH-HARNESS-AC-004, TH-HARNESS-AC-005, TH-HARNESS-AC-018, TH-HARNESS-AC-020, TH-HARNESS-AC-023, TH-HARNESS-AC-027, TH-HARNESS-AC-028, TH-HARNESS-AC-038, TH-HARNESS-AC-039, TH-HARNESS-AC-040, TH-HARNESS-AC-041, TH-HARNESS-AC-042, TH-HARNESS-AC-045, TH-HARNESS-AC-046, TH-HARNESS-AC-056, TH-HARNESS-AC-058, TH-HARNESS-AC-064, TH-HARNESS-AC-070, TH-HARNESS-AC-071, TH-HARNESS-AC-072, TH-HARNESS-AC-073, TH-HARNESS-AC-076, TH-HARNESS-AC-079, TH-HARNESS-AC-093 |
| `TH-HARNESS-REQ-100..149` | Configuration                      | TH-HARNESS-AC-002, TH-HARNESS-AC-003, TH-HARNESS-AC-021, TH-HARNESS-AC-028, TH-HARNESS-AC-029, TH-HARNESS-AC-064 |
| `TH-HARNESS-REQ-150..199` | Result roots and artifact identity | TH-HARNESS-AC-003, TH-HARNESS-AC-015, TH-HARNESS-AC-064, TH-HARNESS-AC-066, TH-HARNESS-AC-067, TH-HARNESS-AC-071 |
| `TH-HARNESS-REQ-200..249` | Output modes                       | TH-HARNESS-AC-004, TH-HARNESS-AC-005, TH-HARNESS-AC-023 |
| `TH-HARNESS-REQ-250..299` | Artifacts and schemas              | TH-HARNESS-AC-000, TH-HARNESS-AC-004, TH-HARNESS-AC-015, TH-HARNESS-AC-019, TH-HARNESS-AC-025, TH-HARNESS-AC-028, TH-HARNESS-AC-031, TH-HARNESS-AC-048, TH-HARNESS-AC-049, TH-HARNESS-AC-062, TH-HARNESS-AC-064, TH-HARNESS-AC-065, TH-HARNESS-AC-066, TH-HARNESS-AC-071, TH-HARNESS-AC-073, TH-HARNESS-AC-074, TH-HARNESS-AC-075, TH-HARNESS-AC-079 |
| `TH-HARNESS-REQ-300..349` | Failure and exit codes             | TH-HARNESS-AC-013, TH-HARNESS-AC-014, TH-HARNESS-AC-032, TH-HARNESS-AC-064, TH-HARNESS-AC-065, TH-HARNESS-AC-066, TH-HARNESS-AC-067 |
| `TH-HARNESS-REQ-350..399` | Scheduler                          | TH-HARNESS-AC-006, TH-HARNESS-AC-018, TH-HARNESS-AC-021, TH-HARNESS-AC-024, TH-HARNESS-AC-030, TH-HARNESS-AC-059, TH-HARNESS-AC-060, TH-HARNESS-AC-061, TH-HARNESS-AC-063, TH-HARNESS-AC-064, TH-HARNESS-AC-065, TH-HARNESS-AC-074, TH-HARNESS-AC-077, TH-HARNESS-AC-078, TH-HARNESS-AC-079 |
| `TH-HARNESS-REQ-400..449` | Services                           | TH-HARNESS-AC-007, TH-HARNESS-AC-010, TH-HARNESS-AC-013, TH-HARNESS-AC-017, TH-HARNESS-AC-025, TH-HARNESS-AC-033, TH-HARNESS-AC-049, TH-HARNESS-AC-056 |
| `TH-HARNESS-REQ-450..499` | Test controls and browser reset lifecycle | TH-HARNESS-AC-008, TH-HARNESS-AC-034, TH-HARNESS-AC-035, TH-HARNESS-AC-050, TH-HARNESS-AC-051, TH-HARNESS-AC-052, TH-HARNESS-AC-053, TH-HARNESS-AC-054, TH-HARNESS-AC-056 |
| `TH-HARNESS-REQ-500..549` | Cleanup                            | TH-HARNESS-AC-009, TH-HARNESS-AC-010, TH-HARNESS-AC-028, TH-HARNESS-AC-036, TH-HARNESS-AC-092 |
| `TH-HARNESS-REQ-550..599` | Platform                           | TH-HARNESS-AC-012, TH-HARNESS-AC-092                    |
| `TH-HARNESS-REQ-600..649` | Security and redaction             | TH-HARNESS-AC-003, TH-HARNESS-AC-011, TH-HARNESS-AC-015, TH-HARNESS-AC-036, TH-HARNESS-AC-056, TH-HARNESS-AC-067, TH-HARNESS-AC-075, TH-HARNESS-AC-076 |
| `TH-HARNESS-REQ-650..699` | Product integration                | TH-HARNESS-AC-013, TH-HARNESS-AC-016, TH-HARNESS-AC-026, TH-HARNESS-AC-039, TH-HARNESS-AC-043, TH-HARNESS-AC-044, TH-HARNESS-AC-047, TH-HARNESS-AC-049, TH-HARNESS-AC-050, TH-HARNESS-AC-051, TH-HARNESS-AC-052, TH-HARNESS-AC-053, TH-HARNESS-AC-054, TH-HARNESS-AC-055, TH-HARNESS-AC-056, TH-HARNESS-AC-062, TH-HARNESS-AC-066, TH-HARNESS-AC-068, TH-HARNESS-AC-069, TH-HARNESS-AC-070, TH-HARNESS-AC-071, TH-HARNESS-AC-080, TH-HARNESS-AC-081, TH-HARNESS-AC-082 |
| `TH-HARNESS-REQ-800..817` | V3 execution control               | TH-HARNESS-AC-082, TH-HARNESS-AC-083, TH-HARNESS-AC-084, TH-HARNESS-AC-085, TH-HARNESS-AC-086, TH-HARNESS-AC-087, TH-HARNESS-AC-088, TH-HARNESS-AC-089, TH-HARNESS-AC-090, TH-HARNESS-AC-091, TH-HARNESS-AC-094, TH-HARNESS-AC-095, TH-HARNESS-AC-096, TH-HARNESS-AC-097, TH-HARNESS-AC-098, TH-HARNESS-AC-099, TH-HARNESS-AC-100, TH-HARNESS-AC-101 |

## 18. Sources and Evidence Limits

This section is traceability and evidence posture. It does not add current conformance behavior.

Primary repository evidence used to shape this NLSpec includes:

- `testing-harness-nlspec.md`, prior draft;
- `nlspec-spec.md`, NLSpec standard;
- Core 00 through Core 04 for product-conformance authority;
- Core 05 for claim-publication separation;
- `docs/domain.md` for vocabulary and owner navigation;
- implementation and testing guides for repository command-surface and harness context;
- `Makefile`, `tools/task_surface_manifest.json`, generated task-surface includes, scheduler manifests, and schema files when present in the repository.
- research reports R01 through R09 under `docs/research/` for state-boundary, handoff, stable-identity, cancellation, integrity, governance, extension-registry, and public-contract rationale only.

The imported standards named by TH-HARNESS-REQ-011 were evaluated for current status as of 2026-07-16. Their primary references are JSON Schema Draft 2020-12, RFC 8785 plus verified errata, RFC 3339, FIPS 180-4, NIST SP 800-218, OWASP Path Traversal, and CWE-22. Research and security guidance explain the chosen boundaries; only the exact imported algorithms and constraints stated in Sections 1 through 17 are normative.

The following evidence categories remain non-normative in this document unless promoted by a requirement above:

| Evidence category                                                | Current role                                                                      |
| ---------------------------------------------------------------- | --------------------------------------------------------------------------------- |
| Recovery docs under `docs/testing-harness-spec-recovery-docs/**` | Historical traceability and diagnostic context.                                   |
| Raw package scripts and tool output                              | Developer convenience or child-command evidence.                                  |
| Script paths, generated Make include names, helper binaries, `internal_helper`, `check_internal`, priority-band names, and generator-only constants | Implementation details unless a requirement above explicitly promotes one. |
| Playwright screenshots, videos, traces, HTML reports             | Diagnostic secret-bearing artifacts.                                              |
| Hosted CI provider workflows                                     | Outside current conformance unless provider source is supplied and later adopted. |
| Visual snapshot refresh diagnostics                              | The helper workflow is current under Sections 6, 8, 11, 15, and 17; raw Playwright image/report internals remain non-normative. |

Exact numeric constants are normative only when they protect security, cleanup safety, bounded output, or deterministic scheduling. Other numeric values in generated manifests, helper names, priority bands, and generator-only constants are implementation details unless this NLSpec gives them a requirement.

The editorial lint for TH-HARNESS-AC-016 rejects the forbidden evidence markers listed in this non-normative section when they appear in Sections 1 through 17. The forbidden markers are: `TODO`, `source_limited`, `source-limited`, `source-observed`, `current code`, `selected evidence`, `recovery evidence`, and `maintainer_decision_required`.

## 19. Future Decisions Outside Current Conformance

The items below are explicitly outside the current conformance profile. They do not block implementation of the current harness contract.

Adoption of `cartulary.testing_harness.current.v3` adopts only Sections 1 through 17 and the current conformance rules explicitly listed there. It MUST NOT adopt any Section 19 future area as current harness conformance, product conformance, provider-specific hosted CI behavior, Playwright diagnostic schema stability, or Core 05 claim-publication evidence. The helper-only visual refresh contract is current only to the extent defined in Sections 6, 8, 11, 15, and 17.

| Future area                                                 | Current treatment                    | Future adoption requirement                                                                  |
| ----------------------------------------------------------- | ------------------------------------ | -------------------------------------------------------------------------------------------- |
| macOS certification                                         | Unsupported for current conformance. | Add platform profile, exact toolchain matrix, and acceptance evidence.                       |
| Windows-native support                                      | Unsupported for current conformance. | Add platform profile separate from WSL2.                                                     |
| Podman/Podman Compose                                       | Unsupported for current conformance. | Add service fixture compatibility profile and cleanup proof.                                 |
| Hosted CI annotations/uploads/artifact-retention dashboards | Provider-neutral `make ci` only.     | Add provider workflow source and provider-specific contract.                                 |
| Playwright report/trace/video/screenshot and visual-geometry diagnostic schemas | Diagnostic-only.                     | Adopt exact Playwright version/schema family or wrapper schema.                              |
| Benchmark-publication harness integration                   | Not part of harness conformance.     | Add Core 05-compatible benchmark manifest and claim-publication profile.                     |
