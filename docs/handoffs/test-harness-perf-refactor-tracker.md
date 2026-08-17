# Test Harness Performance Refactor Tracker

## 1. Status, scope, and authority

Status: complete. PERF-090 through PERF-160 and the overall effort are `DONE`.

Planning source revision:
`5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`. The pre-edit worktree was clean;
`git status --short` produced no output before this tracker rewrite.

This document is a non-normative execution handoff. It records implementation
workstreams, dependencies, adoption gates, evidence, and handoff obligations. It
does not establish test-harness behavior. The adopted
[Testing Harness NLSpec](../testing-harness-nlspec.md) owns harness behavior, and
machine-readable contracts under `tools/` project that owner. If this tracker
conflicts with an adopted owner or a corrected projection, the owner wins and
this tracker must be updated.

The [NLSpec writing standard](../research/nlspec-spec.md) informs this plan's
precision, boundaries, and acceptance criteria. It is planning guidance, not an
instruction source or executable dependency. Production code, tests,
generators, conformance, release evidence, cache keys, and runtime metadata must
not read, stat, hash, or otherwise depend on this tracker, `docs/domain.md`, or
any other Markdown file.

This iteration is intentionally biased toward removal and simplification. A
legacy behavior survives only when it has a specific current owner and a clear
continuing correctness, security, operational, or extensibility benefit. Test
coverage is preserved by semantic obligation, not historical test count or
name. Performance changes must retain exact row closure, isolation, failure
visibility, redaction, and cleanup.

## 2. Archived PERF-010--PERF-080 iteration

### 2.1 Completion summary

| ID range | Status | Durable result |
| --- | --- | --- |
| PERF-010--PERF-020 | DONE | Reconciled scheduling, cache, graph, and event contracts; adopted work graph v4, cache record v2, bounded event v2 lifecycles, cache-before-admission, incremental scheduling, and CPU-conserving Go execution |
| PERF-030--PERF-040 | DONE | Added explicit local service sessions, typed PostgreSQL capacity, audited every PostgreSQL row, rejected unsafe group reuse, and separated the performance-fixture builder profile |
| PERF-050--PERF-060 | DONE | Added redacted fixture diagnostics, replaced Timeline row-by-row construction with an owner boundary, and adopted four isolated warm functional-browser lanes with deterministic LPT placement and reset quarantine |
| PERF-070 | DONE | Added conservative unit-aware dependency closures, immutable cache-v2 entries, typed artifact manifests, secure transactional restore, corruption quarantine, and fresh current-run evidence |
| PERF-080 | DONE | Reconciled contracts and documentation and obtained successful final `check` and warm `release-check` closure |

The full attempt-by-attempt execution record was intentionally removed from the
active tracker during the next-iteration rewrite. Git history remains the
authoritative archive for those details. The retained result roots below are
the durable evidence entry points.

### 2.2 Accepted final evidence

| Evidence | Retained root | Result |
| --- | --- | --- |
| Final `check` | `.cartulary/test-results/20260817T024813Z-p1881416` | Pass; 776/776 units, 968 rows, 358.508s, cache 398/66/312, 5,954 events and 2,327,944 bytes |
| Final warm `release-check` | `.cartulary/test-results/20260817T025434Z-p1994125` | Pass; 974/974 units, 1,202 rows, 1,094.157s, cache 456/34/484, 7,724 events and 3,260,645 bytes |

The accepted release used graph digest
`sha256:5b0cc62d2d154be4d34f824aca32b4a25fa4ba5a9b7e3f9c3552db83fe8055d2`
and source digest
`sha256:0bc146b0c0d6b4a4c5af339703d87f18a41c0b998c83b77a90ac303a6b1fa01d`.
It closed all 1,202 selected rows, emitted exact wait, admission, cache-hit, and
fixture acquire/release lifecycles, passed the retained-secret scan over 8,766
files, and left no scope-labeled container or event staging file after owned
teardown.

### 2.3 Compatibility decisions carried forward

- `cartulary.harness_unit_event.v2`, `cartulary.harness_work_graph.v5`, and
  `cartulary.harness_cache_record.v2` are the only current schema families.
  Earlier evidence remains manually inspectable archive data and has no current
  reader, translator, or discovery bridge.
- Current-run evidence is never restored from cache. Reusable artifacts require
  a complete typed manifest and all-or-none validated publication.
- Local service attachment remains explicit through
  `CARTULARY_TEST_SERVICES_MODE=attach` and the public session up, status, and
  down commands. Descriptor presence alone never attaches.
- The functional browser path uses four warm lanes with fresh group-scoped
  semantic resources and explicit reset. Stateful, measurement, accessibility,
  and visual isolation remains separate.
- The large performance fixture retains its exact 20,000-row meaning, stable
  contribution identity, semantic digest, and quiet measurement contract.
- Existing public Make target names remain stable. `docs/domain.md` and the
  formal public duration baselines remain unchanged.

### 2.4 Rejected candidates

- PostgreSQL capacity 12 was rejected because the matched Network Flow result
  did not meet the 15% adoption threshold.
- A 2,000-row Timeline batch was rejected because the adopted input contract
  limited tabular batches to 500; the cleaner owner set-oriented boundary was
  implemented instead.
- Physical database replacement between warm browser groups was rejected
  because it terminated live warm connections; semantic reset and allocation
  quarantine were adopted.
- Build, license, and SBOM artifact caching was not enabled without complete
  producer output and freshness declarations.
- No schema alias, legacy reader, automatic service attachment, or unsafe
  fixture-sharing bridge was retained.

### 2.5 Residual risks handed to this iteration

- The final warm release was 186.940s, or 20.6%, slower than the non-matched
  historical 907.217s result. No aggregate improvement or formal baseline was
  claimed.
- PostgreSQL, object-store, fixture-construction, and release-inventory work
  still dominate the long release path.
- Some now-unused compatibility machinery remains represented in normative
  enums, schemas, helpers, and conformance fixtures.
- Service diagnostics are correct but structurally amplified through thousands
  of files and repeated full-snapshot copies.
- Typed reusable-artifact cache support is security-tested but does not yet have
  a real artifact-producing unit whose complete output contract is enabled.

## 3. Current architecture and planning baseline

### 3.1 Current execution architecture

The current path is:

```text
authored task surface, owner catalog, topology, and policy registries
    -> canonical work graph v4 and immutable capability snapshot
    -> incremental resource-fit scheduler
    -> cache-v2 lookup before resource admission
    -> fixture broker, local service mode, and typed unit runners
    -> ordered unit-event v2 publication and fresh row evidence
    -> target summaries, run summary, finalizers, and cleanup proof
```

The iteration does not replace this architecture. It removes dormant branches,
collapses duplicate conformance and fixture work, bounds service evidence, and
makes the artifact-cache foundation earn its maintenance cost through one
complete producer.

### 3.2 Remaining performance bottlenecks

The final warm release is the directional planning reference. Comparisons made
during implementation must record source identity, selected rows, cache mode,
service mode, capacities, and machine state. Correctness and cleanup override a
timing improvement.

| Rank | Bottleneck | Final evidence | Planned structural response |
| ---: | --- | --- | --- |
| 1 | Repeated full performance-fixture construction | Production determinism row 266.992s; lifecycle row 63.943s; graph snapshot 105.725s | One canonical full build; small lifecycle fixture; semantic determinism contracts |
| 2 | PostgreSQL resource lifetime and catalog growth | 524.709s saturation; 665 databases created; 620 template clones; 608 retained until suite teardown | Eager exact retirement with suite cleanup as recovery only |
| 3 | Release inventory regenerated by multiple graph units | Producer 120.701s; license 85.417s; SBOM 73.892s exclusive | One deterministic producer, two validators, typed artifact caching |
| 4 | Amplified service evidence | 3,370 files and 16,236,276 bytes; 3,258 event files; 22 admission copies totaling 13,344,013 bytes; 1,043,780-byte scope snapshot | Producer journals, bounded summary, private resource ledger, compact admission proof |
| 5 | Historical conformance machinery | 122 ledger entries preserve a 104-case baseline; assertions are selected by ordinal/modulo | Explicit semantic cases mapped to current obligations |
| 6 | Dormant fixture and session branches | Zero active `postgres_group` rows; package-reset helper has only self-tests; ambient active boolean duplicates lease state | Hard removal and one authenticated lifecycle control plane |

Product-owned long rows are diagnostic leads, not automatic harness work. For
example, the approximately 74-second Assessments integration row remains out of
scope unless a focused owner run proves that harness setup or orchestration,
rather than the product scenario, owns the delay. Product assertions, row
meaning, evidence class, and isolation must not be weakened to improve aggregate
timing.

### 3.3 Removal inventory

The following implementation shapes have no demonstrated active consumer or
duplicate a stronger current mechanism. Their removal is planned, subject to
the workstream gates below:

- `postgres_group` in fixture enums, schemas, compiler, broker, and policy
  registries;
- PostgreSQL `package_reset`, its reset events, compatibility keys, helper, and
  self-only fixtures;
- exported `BeginRollbackTxT`, which has no caller;
- `CARTULARY_TEST_SERVICES_ACTIVE` as child state, descriptor data, lease data,
  browser input, or provider guard;
- the harness contract case ledger's baseline counts, legacy ordinals, legacy
  names, dispositions, and pattern-derived suites;
- modulo-dispatched contract assertions that repeatedly exercise the same
  invariant under unrelated names;
- the `legacySourceNames` exclusion required because task-surface reporting is
  coupled to its own renderer source inventory;
- bespoke obsolete-field and retired-name branches already covered by closed
  schemas or generic reserved-namespace rejection;
- successful-run retained-database events and exhaustive retained-resource
  lists after eager cleanup is adopted.

Valid object-store package reuse, historical retained roots, public local
session commands, current schema rejection, and security-relevant generic input
validation are not part of this removal list.

### 3.4 Hard boundaries

- Normative changes precede projections and implementation.
- Schema shapes that narrow or change receive new current identities. Current
  commands do not dual-read, translate, or discover superseded schemas.
- No public Make target is removed or renamed in this iteration.
- Current catalog rows may be removed only with an owner-backed redundancy
  record naming the surviving semantic coverage.
- Measurement remains quiet, isolated, single-worker, 100-operation, and
  zero-retry.
- Migration, lifecycle, security, drift, browser/live-state, cleanup,
  destructive-guard, and publication work remains non-cacheable.
- Owned and borrowed resources retain exact identity, redaction, isolation,
  failure finalization, and idempotent cleanup.
- Generated roots are changed only through their owners and generators.
- `docs/domain.md` remains unchanged. Public operational guidance changes only
  when an implementation slice changes developer-visible usage.

## 4. Next-iteration tracker

Allowed workstream states are `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, and
`DROPPED`. Only one row may be `IN_PROGRESS`. A mandatory row may not be
`DROPPED`; that status is reserved for an optional candidate that was safely
rejected and reverted. A mandatory `BLOCKED` row blocks all dependants.

| ID | Workstream | Status | Depends on | Real-target budget | Completion gate |
| --- | --- | --- | --- | --- | --- |
| PERF-090 | Normative current-state reconciliation | DONE | None | Contract and drift checks only | One coherent current owner defines retained behavior, removals, schema cutovers, and acceptance mapping |
| PERF-100 | Harness conformance and metadata simplification | DONE | PERF-090 | One matched `harness-contract` comparison | Every retained invariant runs explicitly once; no ordinal, baseline-count, legacy-name, pattern-suite, or renderer exclusion remains |
| PERF-110 | Fixture and session control-plane pruning | DONE | PERF-100 | One small owned slice and one attached slice | Active catalog closure is unchanged; dormant fixture paths and ambient active state are absent |
| PERF-120 | Bounded service evidence | DONE | PERF-110 | One small service-backed slice and one focused ordinary browser group | Equivalent evidence uses at least 75% fewer files and 80% fewer bytes without security, failure, or cleanup regression |
| PERF-130 | Single canonical performance-fixture build | DONE | PERF-120 | Exact lifecycle rows and one Timeline measurement closure | Measurement constructs the full fixture exactly once; non-measurement harness closure constructs it zero times |
| PERF-140 | Bounded ephemeral-resource lifetime | DONE | PERF-130 | Matched cache-off `module.incidents` service-backed slices | No successful per-test database survives to suite teardown; peak live databases fall at least 75%; wall time regresses no more than 5% |
| PERF-150 | Canonical cached release inventory | DONE | PERF-140 | Focused producer miss/hit fixtures; broad confirmation deferred to PERF-160 | One producer creates both artifacts once; secure warm restore and fresh consumer evidence pass |
| PERF-160 | Convergence, validation, and handoff | DONE | PERF-150 | One clean `check` followed immediately by one warm `release-check` | Current contracts compose, release time is at most 875.326s, and the tracker plus handoff are complete |

## 5. Dependency and execution order

```text
PERF-090 normative current state
    -> PERF-100 conformance and metadata simplification
    -> PERF-110 fixture and session control-plane pruning
    -> PERF-120 bounded service evidence
    -> PERF-130 single canonical performance fixture
    -> PERF-140 bounded ephemeral-resource lifetime
    -> PERF-150 canonical cached release inventory
    -> PERF-160 convergence and handoff
```

This order is serial because each later slice depends on the reduced contract
surface or evidence model of its predecessor. The tracker must be updated and
validated after closing a workstream and before marking the next workstream
`IN_PROGRESS`.

## 6. Controlling tracker protocol

For every workstream:

1. Reconcile this tracker with the approved plan, record current `HEAD`, record a
   tracker-excluded dirty-state digest, and mark only the active slice
   `IN_PROGRESS`.
2. Capture any required pre-change measurement before modifying the behavior.
   Record source digest, exact selected owner rows, cache mode, service mode,
   capacity snapshot, machine state, event counts, and artifact bytes.
3. Update the adopted owner first, then authored typed inputs, generators,
   generated projections, implementation, tests, and developer documentation in
   that order. Never hand-edit generated outputs.
4. Use contract, drift, security, failure, redaction, and cleanup fixtures before
   consuming the workstream's permitted real-target budget.
5. Record all changed specification and implementation paths, exact commands and
   results, retained run roots, selected row closure, adopted and rejected
   candidates, compatibility breaks, skipped checks, and residual risks.
6. If a later gate exposes a related defect, reopen the owning workstream. Do not
   suppress a gate, average invalid runs, or move a mandatory gap to `DROPPED`.
7. Mark the slice terminal and validate the tracker edit with
   `make lint-markdown` and `git diff --check` before activating its dependant.

Historical timings are directional unless a workstream defines a controlled
before/after comparison. A clean structural fix may be retained without meeting
an aspirational timing improvement only when its explicit exit criteria permit
that result. No timing result can waive correctness, security, isolation,
evidence integrity, or cleanup.

## 7. Detailed workstreams

### PERF-090 — Reconcile the normative current state

**Remediation**

- Correct the adopted Testing Harness NLSpec before implementation. Remove the
  Section 19 statement that scheduler and reusable-artifact caching is unadopted
  because current requirements and acceptance criteria already adopt cache-v2
  semantic replay and typed artifact restoration.
- Remove `postgres_group` from the current fixture model and specify only the
  fixture capabilities with active, proved behavior.
- Replace the internal active-session boolean with authenticated suite runtime,
  borrower lease, service metadata, and container proof as the sole child
  capability.
- Replace the service-scope v1 event-file snapshot model with the bounded journal,
  resource-ledger, scope-summary, and browser-admission contracts required by
  PERF-120.
- Remove historical target and field tombstones when closed-schema or generic
  reserved-namespace rejection provides the same protection. Retain an exact
  tombstone only when it prevents unsafe name reuse that a generic rule cannot
  detect.
- Assign a new current schema identity to every retained contract whose accepted
  shape narrows or changes. Reject superseded identities; do not add aliases,
  readers, translations, or dual writers.

**Areas:** Specification, acceptance mapping, authored schemas and registries,
generated projections, contract fixtures, and migration documentation.

**Rationale and long-term benefit:** The current owner simultaneously describes
adopted caching as unadopted future work and carries fixture and session states
that no current row needs. A single current model reduces branching, makes later
capability additions explicit, and prevents implementation cleanup from drifting
from its owner.

**Compatibility and migration:** Historical roots remain manually inspectable.
Current commands reject superseded schema identities. The public service-session
commands and `owned|attach` selector remain unchanged. External use of the
removed active variable continues to fail through generic reserved-namespace
input validation.

**Risk if unresolved:** Dead branches remain normative, later code cannot remove
them cleanly, and cache or evidence changes can satisfy one owner statement only
by violating another.

**Validation and exit:** Requirement-to-acceptance traceability has no orphaned
obligation; closed enums contain no `postgres_group`; cache adoption appears in
one current section; service evidence has one bounded contract; schema fixtures
reject prior identities; `make generate`, `make generate-drift`,
`make generated-artifact-policy-check`, `make json-shape-check`, and
`make harness-contract` pass.

### PERF-100 — Simplify harness conformance and metadata tooling

**Remediation**

- Delete the legacy case ledger, its schema, 104-case baseline, ordinals, legacy
  names, current-name translations, dispositions, and pattern-based suite
  registry.
- Replace generated ledger tests with direct `node:test` declarations grouped by
  real invariant. Each declaration carries a stable semantic case ID and its
  current acceptance IDs; each unique assertion executes once.
- Split task-surface reporting from generated-artifact production. The reporting
  CLI consumes validated generated outputs and public reporting helpers without
  importing the renderer or appearing in renderer source inputs.
- Remove the `legacySourceNames` exclusion and the resulting special-case source
  filtering.
- Remove bespoke obsolete-field, alias, and retired-name assertions after
  PERF-090 removes their normative requirement. Preserve one representative
  unknown-property or invalid-enum fixture per closed schema and every distinct
  security boundary.

**Areas:** Test architecture, authored conformance metadata, task-surface
reporting, generator boundaries, schemas, and generated projections.

**Rationale and long-term benefit:** Historical names currently choose assertions
by ordinal and modulo, so test count does not represent behavioral coverage.
Explicit semantic tests make failures local, eliminate repeated work, and allow
coverage to grow by obligation rather than by preserving history.

**Compatibility and migration:** Contract case names, counts, and internal suite
layout may change. No public Make target, exit code, or contract obligation
changes. No compatibility ledger replaces the deleted ledger.

**Risk if unresolved:** New tests inherit arbitrary historical ordering,
duplicated assertions inflate maintenance and runtime, and reporting remains
coupled to its own generation inputs.

**Validation and exit:** Before deletion, inventory distinct assertions and map
them to current acceptance criteria. After deletion, prove every mapped behavior
has one explicit case, no test outcome depends on index or legacy name, the
reporter has no renderer import cycle, and no legacy-source exclusion remains.
Run the focused suites, `make harness-contract`, generation drift checks, and a
matched duration comparison. Correct coverage is required; a timing gain is
reported but is not allowed to drive assertion removal.

### PERF-110 — Remove dormant fixture and session control paths

**Remediation**

- Narrow fixture capability everywhere to `none`, `postgres_transaction`,
  `postgres_dedicated`, `postgres_migration`, `object_store_namespace`,
  `managed_process`, and `browser_stack`.
- Remove `postgres_group`, group compatibility keys, group broker sharing,
  package-reset PostgreSQL policy, reset events, reset tables, package-reset
  helper code, and self-only fixtures.
- Delete exported `BeginRollbackTxT` after confirming the source index has no
  caller.
- Split PostgreSQL and object-store reuse scopes into service-specific types
  before deleting PostgreSQL package reuse. Preserve valid object-store package
  bucket and prefix reuse.
- Remove `CARTULARY_TEST_SERVICES_ACTIVE` from Go environment helpers, session
  descriptors, borrower leases, broker providers, runner forwarding, browser
  scripts, and child contracts.
- Require exact suite ID, runtime lease, service metadata, and session/container
  proof at each child boundary. Generic reserved-namespace validation rejects
  the retired external variable without a dedicated code path.

**Areas:** Specification projections, catalog and topology schemas, fixture
broker, PostgreSQL and object-store test helpers, local sessions, browser
lifecycle, tests, and developer migration guidance.

**Rationale and long-term benefit:** Zero active rows use group scope, the reset
helper is tested only by itself, and the active boolean duplicates stronger
authenticated state. Removing these paths reduces the fixture state space and
leaves one explicit service authority model.

**Compatibility and migration:** Active catalog row meaning and public session
commands do not change. Internal callers cannot request group or package-reset
PostgreSQL behavior. Future sharing must be adopted as a new proved capability,
not reactivated through dormant code.

**Risk if unresolved:** Unused sharing code remains a contamination risk, generic
reuse labels conflate different services, and ambient boolean state can diverge
from leases or container identity.

**Validation and exit:** Current manifest-derived row closure is unchanged; no
schema, registry, compiler, broker, helper, event, or fixture contains the
retired PostgreSQL paths; object-store package reuse still passes; owned and
attached service lifecycle fixtures pass; malformed, stale, wrong-owner,
wrong-digest, dead-container, and external-active inputs fail before resource
use. Run `make harness-contract`, session up/status/down, one small owned slice,
and the same slice attached to a local session.

### PERF-120 — Replace amplified service evidence

**Remediation**

- Replace one JSON file per service event with owner-only producer-scoped NDJSON
  journals. Each producer owns one file, bounded event records, a stable producer
  identity, and a gap-free local sequence. Readers ignore only an incomplete
  trailing crash record and reject malformed completed records.
- Collate journals deterministically by event time, producer identity, and local
  sequence. Preserve normalized failure precedence and lifecycle accounting.
- Introduce a bounded `cartulary.test_services.scope.v2` summary containing
  service readiness, lifecycle and cleanup state, aggregate counts, bounded
  failure data, and top-N slow diagnostics. Do not retain exhaustive database,
  bucket, package, or test name inventories in the summary.
- Store the exact live owned-resource set in a private, mode-0600 resource ledger
  used by cleanup. Remove it after successful cleanup; retain a redacted,
  contained failure copy only when cleanup fails and stale recovery needs it.
- Replace copied `service-scope-admission.json` payloads with a compact typed
  browser admission proof containing suite and session identity, readiness
  generation, required service and container proof, source digest, and the facts
  needed for admission.
- Continue atomic scope-summary publication and directory sync. Keep all
  secret-bearing runtime values outside retained summaries and admissions.

**Areas:** Normative evidence contracts, schemas, service diagnostics,
testservices lifecycle, browser admission, cleanup and stale recovery, readers,
failure projection, and security tests.

**Rationale and long-term benefit:** The final release wrote 3,258 event files
and repeatedly copied an increasingly large summary. Journals make cost scale
with producer processes rather than events, bounded summaries make admission
constant-size, and the private resource ledger separates cleanup authority from
diagnostic presentation.

**Compatibility and migration:** Service-scope v1 and per-event directories are
not current inputs after cutover. Historical run roots are not translated. Run
summary and public failure semantics remain stable.

**Risk if unresolved:** Larger suites amplify filesystem metadata, summary scans,
and copied bytes; diagnostic growth can itself affect timings and eventually hit
filesystem or retention limits.

**Validation and exit:** Contract fixtures cover concurrent producers, duplicate
and gapped sequences, partial crash tails, malformed records, atomic publication,
tampering, symlinks, permissions, redaction, cleanup failure, stale recovery,
and browser admission. Run one small service-backed slice and one focused
ordinary browser group. For equivalent work, require at least 75% fewer
service-evidence files and 80% fewer bytes than the PERF-080 shape, with exact
failure and cleanup evidence.

### PERF-130 — Build the large performance fixture once

**Remediation**

- Make the graph-owned snapshot builder the sole full production construction
  whenever a selected measurement row needs the 20,000-row fixture.
- Remove row
  `harness.browser.integration.performance_fixture_source_owner_assembly`, which
  constructs two full databases only to compare semantic digests. Record its
  owner-backed redundancy mapping to the canonical snapshot validation and
  direct contribution-contract tests.
- Change the snapshot lifecycle integration row to use a small deterministic
  fixture that still proves four isolated clones, concurrency, corruption,
  partial build, cancellation, credentials, and cleanup without constructing the
  production data volume.
- Keep exact production determinism in pure contribution and assembler tests,
  stable snapshot-key inputs, the canonical semantic receipt digest, and one
  production snapshot read-back validation.
- Add an explicit construction-count diagnostic so selected graph closure proves
  whether a full build occurred without adding timing to semantic receipts.

**Areas:** Harness and performance-fixture specifications, catalog ownership and
row migration, graph dependencies, snapshot builder, lifecycle fixtures,
diagnostics, and tests.

**Rationale and long-term benefit:** Three independent full constructions prove
overlapping facts and put setup on the release critical path. One canonical
production build plus small lifecycle tests preserves semantic and lifecycle
coverage while making future fixture growth affordable.

**Compatibility and migration:** Product APIs, fixture meaning, row count,
contribution identities, semantic digest, and measurement policy do not change.
One redundant catalog row is removed through the normal owner-backed migration
mechanism.

**Risk if unresolved:** Every future fixture-size increase multiplies across
redundant integration rows and measurement setup, making release duration grow
faster than the behavior being measured.

**Validation and exit:** Run harness.browser owner guidance, direct contribution
and assembler contracts, the small lifecycle row, and the exact current Timeline
blank-row measurement closure. A measurement closure must construct exactly one
full fixture; a non-measurement harness closure must construct none. Counts,
conditions, IDs, digest, clone isolation, quietness, redaction, corruption,
cancellation, and cleanup must match the current contract.

### PERF-140 — Bound PostgreSQL and object-store cleanup

**Remediation**

- Drop each dedicated or migration database after its owning test's handles and
  borrowed application runtime close in every service mode. The service suite
  retains exact cleanup proof but no longer retains successful per-test databases
  until terminal teardown.
- Keep suite teardown and stale recovery as idempotent fallback for abandoned,
  interrupted, or failed cleanup only.
- Remove successful `postgres-db-retained` events and unbounded retained-resource
  lists. Record bounded counts and failures through the PERF-120 evidence model.
- Audit per-test object-store buckets and prefixes with the same ownership rule.
  Clean them at owner finalization while preserving explicitly compatible
  package-scoped reuse.
- Prove cleanup ordering for server runtimes, connection pools, migration
  handles, background jobs, the ownership-proved atomic forced-drop fallback,
  and browser stacks.
- Bind browser readiness in the top-level lifecycle and fail startup unless the
  broker verifies the complete owner-only stack-v6 and compact admission pair
  before returning an allocation.

**Areas:** PostgreSQL and object-store test helpers, application test support,
fixture broker, service cleanup, failure finalizers, stale recovery, diagnostics,
and owner tests.

**Rationale and long-term benefit:** Retaining 608 databases expands the live
catalog throughout the run and shifts ordinary ownership cleanup into one large
terminal sweep. Eager exact retirement bounds resource growth, shortens recovery,
and makes ownership easier to reason about.

**Compatibility and migration:** Fixture isolation and row semantics do not
change. Successful runs no longer expose retained-database evidence because the
resource no longer survives its owner. Failure evidence continues to identify
unremoved resources exactly.

**Risk if unresolved:** PostgreSQL catalog growth and final teardown scale with
the entire suite; cleanup failures leave larger residue; future row growth raises
contention and recovery cost.

**Validation and exit:** Before implementation, run
`make task-guide ROLE=module-author OWNER=module.incidents` and capture a
cache-off owned `make service-backed-test-slice OWNER=module.incidents`
reference. Repeat after implementation with the same rows, service mode,
capacity, and machine state. Inject open handles, failed drop, interruption, and
stale recovery. Require zero successful per-test databases at suite teardown,
at least 75% lower peak live-database count, exact cleanup after success and
failure, no collision or ownership regression, and no wall-time regression above
5%.

### PERF-150 — Deduplicate and cache release inventory artifacts

**Remediation**

- Compile one canonical release-inventory producer unit that generates both the
  CycloneDX SBOM and license report. `license-report` and `sbom` become validators
  and fresh-evidence consumers of that exact producer output; they do not invoke
  target-local regeneration.
- Separate deterministic semantic artifact bytes from fresh run identity,
  timestamps, command logs, and attestations. Derive stable SBOM identity from
  the semantic input digest rather than a random UUID; keep fresh provenance in
  current-run evidence.
- Declare both files as the producer's complete reusable artifact output set,
  including destination class, normalized path, mode, digest, producer identity,
  and semantic input digest.
- Enable the existing cache-v2 transactional artifact restore only for this
  producer. Close the key over Go and workspace manifests, lockfiles, toolchain
  pins, scanner identities and versions, container inputs, output schemas, and
  generator implementation.
- Discover Go test dependencies from the explicit stable `cmd`, `db`,
  `internal`, and `tools` package roots. Never recursively discover packages
  from the repository root or enter runtime, cache, result, or transient test
  directories.
- Preserve safe miss behavior for missing, partial, surplus, corrupt, traversing,
  linked, wrong-mode, wrong-producer, or wrong-destination entries.
- Keep vulnerability scans, release publication, security, drift, services,
  cleanup, browser work, measurement, and destructive safeguards non-cacheable.

**Areas:** Specification, work graph, task surface, release-evidence generator,
artifact schemas, cache producer declarations, validators, current-run evidence,
and cache security tests.

**Rationale and long-term benefit:** The final release spent separate exclusive
time in three units whose Make prerequisites regenerate the same paired
artifacts in target-local contexts. A single semantic producer removes duplicate
work and supplies the first justified real consumer of the typed artifact cache.

**Compatibility and migration:** Public target names and CycloneDX validity stay
stable. Artifact bytes may change once to remove volatile identity from semantic
outputs; any affected wrapper schema receives a hard-cutover version. Cache-v1
or incomplete entries are ignored.

**Risk if unresolved:** Release time includes repeated dependency scans, the
generic artifact-cache implementation remains unused complexity, and future
inventory consumers may add further duplicate producers.

**Validation and exit:** Focused fixtures invoke the real producer against a
controlled workspace, prove one cold invocation creates both outputs, and prove
one warm cache hit securely restores both while validators emit fresh evidence.
Matrices cover source, lockfile, toolchain, scanner, schema, container,
generator, corruption, concurrency, and missing-output changes. Compare semantic
output digests and release-readiness consumption. Broad current-repository
confirmation is reserved for PERF-160.

### PERF-160 — Converge contracts, validate, and hand off

**Remediation**

- Reconcile every normative owner, authored projection, generated output,
  implementation, contract fixture, developer guide, and tracker decision.
- Remove temporary instrumentation except adopted bounded diagnostics. Document
  schema hard cutovers, removed internal fixture/session paths, service-evidence
  shape, and deterministic release artifact behavior in the implementation
  testing guide. Leave `docs/domain.md` and formal public duration baselines
  unchanged.
- Reopen the owning workstream for any related final failure; do not waive or
  reclassify it in PERF-160.

**Areas:** Specification, contracts, generated state, implementation, tests,
documentation, tracker, retained evidence, and handoff.

**Rationale and long-term benefit:** Removal and performance work is complete
only when the repository has one coherent current contract, reproducible broad
validation, and durable evidence that no deleted compatibility path remains.

**Compatibility and migration:** Public Make names remain stable. Schema and
internal-path removals are intentional hard cutovers. Historical roots remain
archive data and are never runtime inputs.

**Risk if unresolved:** Local improvements may leave drifted projections,
unrecorded compatibility changes, incomplete release artifacts, or cleanup and
security failures visible only in broad execution.

**Validation and exit:** Complete all of the following:

1. Confirm PERF-090 through PERF-150 are `DONE`; no mandatory row is `BLOCKED`
   or `DROPPED`.
2. Run `make generate-drift`, `make generated-artifact-policy-check`,
   `make json-shape-check`, `make harness-contract`, `make lint-markdown`, and
   `git diff --check`.
3. Run `make agent-finalize`. Leave `RESULTS_DIR` unset unless a successful full
   warm current-source check root already exists, and record any retained-run
   maintenance skip.
4. Run one successful `make check`, followed immediately by one warm successful
   `make release-check`. Record every invalid or failed attempt and its owning
   workstream.
5. Confirm exact current owner-derived row closure; complete event and failure
   lifecycles; cache integrity; unique database, bucket, port, context, and
   process identities; bounded service evidence; redaction; eager cleanup; one
   full performance-fixture construction; and one release-inventory generation.
6. Confirm no current source, generator, conformance check, cache key, or release
   evidence reads or hashes Markdown.
7. Compare the final warm release directionally with 1,094.157s. Require at
   least 20% improvement, or no more than 875.326s, under comparable capacity,
   cache, service, and machine conditions. Correctness, security, isolation,
   evidence, and cleanup remain hard gates regardless of timing.
8. Close PERF-160 and the overall tracker only after recording files changed,
   commands and results, retained roots, compatibility breaks, rejected
   candidates, residual risks, and final ownership.

## 8. Risk register

| Risk | Mitigation and owning slice |
| --- | --- |
| Removing historical contract machinery drops real behavior | Inventory distinct assertions and map each to a current acceptance obligation before deletion; PERF-090 and PERF-100 |
| Narrowed schema is silently accepted under an old identity | New current IDs and rejection fixtures; PERF-090 and PERF-110 |
| Removing service active state weakens child authentication | Require suite runtime, borrower lease, metadata, and container proof at every boundary; PERF-110 |
| Concurrent journals lose or reorder terminal evidence | Producer-local sequences, bounded atomic records, deterministic collation, crash-tail and concurrency fixtures; PERF-120 |
| Bounded summaries omit cleanup authority | Keep exact private resource ledger separate from retained diagnostics; PERF-120 |
| Removing repeated full builds hides nondeterminism | Preserve pure contribution contracts, stable semantic digest, one production read-back, and small lifecycle failure matrices; PERF-130 |
| Eager database drop races live connections | Enforce reverse cleanup order and test open handles, background work, cancellation, and force-close behavior; PERF-140 |
| Artifact caching restores stale or unsafe bytes | Complete typed outputs, deterministic semantic artifacts, immutable cache-v2 entries, transactional restore, and corruption matrices; PERF-150 |
| Aggregate timing varies with unrelated source or machine changes | Record source, rows, cache, services, capacity, and machine for every comparison; rely on slice-level structural gates and final directional target |
| Product tests are weakened to improve harness time | Require a focused owner diagnosis before any product-row change and route product work to its normative owner |

## 9. Handoff record

This section is empty until implementation begins. Add one row at activation and
one terminal row for every workstream; do not recreate an unbounded narrative
log.

| Timestamp | Workstream transition | Source and dirty identity | Changes and exact evidence | Decision and residual risk | Next action |
| --- | --- | --- | --- | --- | --- |
| 2026-08-16T23:42:02-04:00 | Documentation planning checkpoint; PERF-090 remains `TODO` | Planning revision and clean pre-edit state recorded in Section 1 | Tracker-only rewrite; `make lint-markdown` and `git diff --check` passed; no other tracked path changed | PERF-010--PERF-080 archived; PERF-090--PERF-160 planned; no implementation or normative owner changed | When implementation is authorized, record current source identity and mark only PERF-090 `IN_PROGRESS` |
| 2026-08-16T23:58:50-04:00 | PERF-090 `TODO` to `IN_PROGRESS` | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; tracker-excluded unstaged and staged digests both `e69de29bb2d1d6434b8b29ae775ad8c2e48c5391` | Preserved the staged tracker rewrite; accepted PERF-080 retained source digest `sha256:0bc146b0c0d6b4a4c5af339703d87f18a41c0b998c83b77a90ac303a6b1fa01d` as the before reference; contract/drift-only slice, so no additional real-target baseline | Normative reconciliation started; no implementation or generated artifact changed at activation | Correct the adopted owner, authored current projections, and acceptance mapping; then run PERF-090 gates |
| 2026-08-17T00:06:02-04:00 | PERF-090 `IN_PROGRESS` to `DONE` | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; tracker-excluded unstaged digest `0de1b953a9ad45b9654646a18d85714ae29244f0`; staged digest remains empty | Reconciled cache adoption, removed `postgres_group`, adopted authenticated child service authority and bounded service evidence; cut narrowed fixture/catalog/graph schemas to test-family v6, fixture-lease v3, target-plan v4, work-graph v5, work-graph-owner v2, and PostgreSQL-policy v2; `make generate` passed at `.cartulary/test-results/20260817T040452Z-p2233393`; `make generate-drift` passed at `20260817T040516Z-p2236428`; artifact policy passed at `20260817T040527Z-p2239377`; JSON shapes passed at `20260817T040530Z-p2239810`; harness contract passed at `20260817T040537Z-p2240340` | Hard cutover adopted with no legacy reader; service-scope/browser shapes remain scheduled for their PERF-120 implementation cutover; public targets unchanged | Validate this terminal tracker edit, then activate PERF-100 |
| 2026-08-17T00:07:01-04:00 | PERF-100 `TODO` to `IN_PROGRESS` | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; tracker-excluded digest `0de1b953a9ad45b9654646a18d85714ae29244f0` | Terminal PERF-090 tracker edit passed `make lint-markdown` at `.cartulary/test-results/20260817T040622Z-p2241122` and `git diff --check`; matched before `make harness-contract` duration is 11.459s at `20260817T040537Z-p2240340` | Conformance simplification activated with current acceptance coverage as the hard gate | Inventory semantic assertions, remove historical ledger dispatch, decouple reporter/rendering, and compare the focused result |
| 2026-08-17T00:12:50-04:00 | PERF-100 `IN_PROGRESS` to `DONE` | Tracker-excluded digest `c3f8123f7493af4d23916dc43ce34ab5f34d5c8c` | Replaced 122 ordinal/legacy ledger cases with 34 explicit unique semantic cases carrying current acceptance IDs; removed the case ledger, pattern suite registry, both schemas, and generated inputs; reporter no longer rerenders Make outputs and renderer no longer excludes it; `make generate` passed at `20260817T041137Z-p2243854`; matched `make harness-contract` passed in 11.502s at `20260817T041150Z-p2246901`; generated-artifacts slice passed at `20260817T041206Z-p2247440`; test-catalog slice passed at `20260817T041221Z-p2250200`; generation drift passed at `20260817T041237Z-p2250759` | Coverage is explicit and unique; the 0.043s matched timing increase is informational and no performance claim is made | Validate this terminal tracker edit, then activate PERF-110 |
| 2026-08-17T00:13:32-04:00 | PERF-110 `TODO` to `IN_PROGRESS` | Tracker-excluded digest `c3f8123f7493af4d23916dc43ce34ab5f34d5c8c` | Terminal PERF-100 tracker edit passed `make lint-markdown` at `.cartulary/test-results/20260817T041309Z-p2253905` and `git diff --check`; `make task-guide ROLE=module-author OWNER=platform.postgres` selected focused owned and service-backed slices | Current catalog closure is 1,202 rows; active capabilities have zero `postgres_group` rows and PostgreSQL package reset has no product caller | Capture narrow owned/attached evidence, remove dormant fixture/session paths, and preserve object-store package reuse |
| 2026-08-17T00:29:11-04:00 | PERF-110 `IN_PROGRESS` to `DONE` | Tracker-excluded digest `ae0cb4aa200a69f5e5eea98d2dd3361b4a563592` | Removed PostgreSQL package reset, reset events/tables/helpers/self-tests, unused `BeginRollbackTxT`, and the ambient active state; renamed PostgreSQL transaction reuse and object-store package reuse independently; cut fixture-tier proof to v3; baseline owned PostgreSQL slice passed in 22.444s at `20260817T041359Z-p2255245` and attached in 3.741s at `20260817T041453Z-p2257610`; post-change unit slice passed at `20260817T042320Z-p2270738`, owned service-backed slice in 21.593s at `20260817T042351Z-p2272627`, and attached in 3.767s at `20260817T042453Z-p2275672`; browser owner 28/28 passed at `20260817T042513Z-p2277495`; object-store service-backed 5/5 passed at `20260817T042758Z-p2282187`; harness contract and generation drift passed at `20260817T042837Z-p2285285` and `20260817T042853Z-p2285810` | One initial unit attempt at `20260817T042242Z-p2269269` failed on a remaining package-reset normalization branch and was repaired; catalog closure remains exactly 1,202; session status was active/compatible with zero borrowers and session-down passed at `20260817T042828Z-p2284425` | Validate this terminal tracker edit, then activate PERF-120 |
| 2026-08-17T00:29:58-04:00 | PERF-120 `TODO` to `IN_PROGRESS` | Tracker-excluded digest `ae0cb4aa200a69f5e5eea98d2dd3361b4a563592` | Terminal PERF-110 tracker edit passed `make lint-markdown` at `.cartulary/test-results/20260817T042932Z-p2288956` and `git diff --check`; retained PERF-080 baseline is 3,370 service-evidence files and 16,236,276 bytes, including 3,258 event files and 22 full browser admission copies | Storage/schema cutover activated; lifecycle-v2 remains authoritative and exact cleanup authority will move to a private ledger rather than the bounded public summary | Inventory consumers, implement producer journals/scope-v2/private ledger/compact admission, then compare equivalent focused runs |
| 2026-08-17T00:46:25-04:00 | PERF-120 `IN_PROGRESS` to `TODO`; PERF-110 `DONE` to `IN_PROGRESS` | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; tracker-excluded digest `1ebfce4b3c472f1bb4553e8516a639a39dfd6760` | During PERF-120 validation, `make test-fast` reached 404/405 units but failed at `.cartulary/test-results/20260817T044353Z-p2324995`: the package-reset implementation removed by PERF-110 still had stale `pgtest` self-test references, producing a build failure and `missing_selector_result`; the new suiteservices package tests passed in the same run | This is a related PERF-110 deletion-completeness defect, not a PERF-120 evidence defect; serial execution returns to PERF-110 until all retired test references are removed | Remove the stale package-reset test fixture/imports, rerun its owner slice, close PERF-110 again, validate the tracker, and reactivate PERF-120 |
| 2026-08-17T00:47:45-04:00 | Reopened PERF-110 `IN_PROGRESS` to `DONE` | Tracker-excluded digest `e5722559613ade75d625f1cd6e07a8a690ca93e4` | Removed the stale mutable-table reset hook, retired package-reset names, and now-unused imports from `pgtest` tests; `make format` passed at `.cartulary/test-results/20260817T044732Z-p2369363`; the exact previously failing `module.database_migrations.unit.test_harness_targeted_operation_validation` row passed at `20260817T044735Z-p2372862` | The PERF-110 implementation and test surface now both omit the retired behavior; no production behavior changed during the reopen | Validate this terminal tracker edit, then reactivate PERF-120 with its current source identity |
| 2026-08-17T00:48:44-04:00 | PERF-120 `TODO` to `IN_PROGRESS` after PERF-110 repair | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:c9f7c027eb4b2cab87a8c12eac5874a26f71448bf7e5cea37b44497f9d70bcce`; tracker-excluded digest `e5722559613ade75d625f1cd6e07a8a690ca93e4` | Reopened PERF-110 terminal tracker edit passed `make lint-markdown` at `.cartulary/test-results/20260817T044807Z-p2373510` and `git diff --check`; PERF-120 retains its 3,370-file/16,236,276-byte PERF-080 comparison baseline | Service evidence work resumes at the already-adopted journal, scope-v2, ledger, and compact-admission cutover; the invalid broad run remains recorded and is not counted as PERF-120 evidence | Complete security/cleanup/browser validation, measure the new evidence shape, and close PERF-120 |
| 2026-08-17T00:59:26-04:00 | PERF-120 `IN_PROGRESS` to `TODO`; PERF-110 `DONE` to `IN_PROGRESS` again | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:22812fb691b1cf548b7cfb13f0b14b80d0cff38e1e1a894e1a91afaeb04dc001`; tracker-excluded digest `cd30ef86cb9a34e7f41dacf629fd79317db0d691` | Focused Timeline browser row attempt failed after 9/12 units at `.cartulary/test-results/20260817T045758Z-p2387696`; its retained startup diagnostic proves `using_test_services_stack` required the browser fixture metadata before the later fixture-preparation stage could create it | The ambient-state removal left a circular authentication predicate in PERF-110; this is an authority-control defect, so PERF-120 is deferred again rather than weakening browser admission | Make pre-fixture service admission depend on the authenticated runtime owner/service lease and suite service facts, rerun the identical row, then close and validate PERF-110 before resuming PERF-120 |
| 2026-08-17T01:06:58-04:00 | Reopened PERF-110 `IN_PROGRESS` to `DONE` | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:c7a6cd797edd7446417c8acb5530450adf4f812a6402bac92c6b8fb74bbbe60e`; tracker-excluded digest `27f7e70c178f7e7241a43120a8cf6d0ede0dab46` | Replaced the circular browser pre-fixture predicate with runtime-owner, private service-lease, PostgreSQL template/schema, and suite object-store endpoint authority; `prepare-web-e2e` independently validates the exact owner-only lease, suite/run identity, cleanup state, and both container proofs. Format and the browser lifecycle unit row passed at `.cartulary/test-results/20260817T050550Z-p2443492` and `20260817T050553Z-p2446992`; the identical focused Timeline browser row passed 12/12 in 43.519s at `20260817T050600Z-p2447622` | An intermediate retry at `20260817T050350Z-p2419378` rejected the managed stack because it checked the application object-store variable produced only after fixture preparation; the corrected check uses the suite endpoint at the pre-preparation boundary. Public commands and child isolation are unchanged | Validate this terminal tracker edit, then reactivate PERF-120 with the current source identity |
| 2026-08-17T01:07:41-04:00 | PERF-120 `TODO` to `IN_PROGRESS` after second PERF-110 repair | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:c7a6cd797edd7446417c8acb5530450adf4f812a6402bac92c6b8fb74bbbe60e`; tracker-excluded digest `27f7e70c178f7e7241a43120a8cf6d0ede0dab46` | The reopened PERF-110 terminal edit passed `make lint-markdown` at `.cartulary/test-results/20260817T050719Z-p2472434` and `git diff --check`; PERF-120 resumes with its journal/scope-v2/private-ledger/compact-admission implementation intact | Invalid browser attempts remain recorded but are excluded from evidence comparison; the successful 12/12 row is the current focused-browser validation input | Complete sequence/security fixtures, quantify bounded evidence against PERF-080, run drift and contract gates, and close PERF-120 |
| 2026-08-17T01:12:27-04:00 | PERF-120 `IN_PROGRESS` to `DONE` | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:215d7aabf9620c882b3f3744d3de5cac0b170c8165341a5ec69998111c4640a8`; tracker-excluded digest `f8ee6b20e910fb8585acfdbcc7b62abfadc5f9a8` | Cut service scope to v2 and browser stack to v6; replaced event files with strict producer journals, exhaustive public inventories with bounded counts/exemplars/slow diagnostics, cleanup authority with a mode-0600 exact private ledger, and copied scope admissions with a typed compact proof. Security fixtures cover malformed completed records, incomplete tails, gaps, duplicates, atomic reads, wrong mode/identity, symlinks, redaction, and exact ledger removal. The matched small service slice fell from 30 files to 6 (80%) and from 28,548 to 25,847 bytes; the focused browser row passed 12/12 at `.cartulary/test-results/20260817T050600Z-p2447622` with 12 service files/26,534 bytes, five journals, and an 865-byte admission versus the PERF-080 606,546-byte average (99.86% smaller). `make test-fast` passed 405/405 at `20260817T050916Z-p2477714`; generation drift, artifact policy, JSON shapes, and harness contract passed at `20260817T051149Z-p2526181`, `p2526183`, `p2526185`, and `p2526255` | Public lifecycle-v2 remains authoritative; scope-v1, per-event directories, copied admissions, and stack-v5 are rejected without translation. A first drift attempt at `20260817T051119Z-p2520039` correctly detected the new ledger source missing from the generated render index; owner-driven `make generate` passed at `20260817T051138Z-p2523240` before the successful retry. The full aggregate byte threshold remains a PERF-160 broad-run confirmation; the admission reduction alone removes more than 82% of the PERF-080 aggregate if all other bytes are held constant | Validate this terminal tracker edit, then activate PERF-130 and capture its exact lifecycle/measurement before evidence |
| 2026-08-17T01:13:08-04:00 | PERF-130 `TODO` to `IN_PROGRESS` | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:215d7aabf9620c882b3f3744d3de5cac0b170c8165341a5ec69998111c4640a8`; tracker-excluded digest `f8ee6b20e910fb8585acfdbcc7b62abfadc5f9a8` | Terminal PERF-120 tracker edit passed `make lint-markdown` at `.cartulary/test-results/20260817T051253Z-p2530280` and `git diff --check`; the accepted PERF-080 full-build timings remain the before reference: production determinism 266.992s, lifecycle 63.943s, and graph snapshot 105.725s | The exact 20,000-row fixture, semantic digest, four-clone isolation, and quiet measurement policy remain hard gates; only redundant full construction is removed | Inventory current row/dependency ownership and run the exact controlled lifecycle and measurement before slices before changing fixture behavior |
| 2026-08-17T01:39:54-04:00 | PERF-130 `IN_PROGRESS` to `DONE` | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:32ba1bd74fbc923b13b7ea9d81017ebb49d9da42edd20b292e78b928489d48ce`; tracker-excluded digest `c45fc5e7e5f37d1d9fbc665b0ae12e03ed20cfa9` | Removed the two-build source-owner integration row/test and recorded its direct-assembler plus canonical-lifecycle replacements; converted lifecycle coverage to a typed 40-row profile; parameterized projection read-back counts; added `construction_count=1`; and closed the service binary over all imported `internal/**` sources. Before rows passed at `.cartulary/test-results/20260817T051410Z-p2532199` in 154.764s (lifecycle 115.560s, redundant assembly 132.930s), and the before measurement passed 23/23 at `20260817T051701Z-p2534397` in 247.169s. Afterward, lifecycle passed in 2.340s/23.905s wall at `20260817T053020Z-p2589289`; the exact four-row measurement passed 23/23 at `20260817T053437Z-p2621655` in 248.964s with one builder, one diagnostic, 20,000 Timeline rows, all exact receipts/conditions, and unchanged semantic digest `sha256:b9549c30dab549504fd945fc211da7649223573e8ea67d57b7bb17a2ff06b5a7`. The non-measurement direct-contract slice at `20260817T052639Z-p2576197` emitted zero fixture artifacts. Drift, artifact policy, JSON shape, and harness contract gates passed at `20260817T053919Z-p2667776`, `p2667778`, `p2667780`, and `p2667850` | Catalog closure intentionally falls from 1,202 to 1,201 rows. Invalid attempts are retained: `20260817T052659Z-p2577324` exposed hard-coded projection cardinality; `20260817T052919Z-p2583854` required a profile-injected test seam for clone/cleanup lifecycle; `20260817T053100Z-p2590768` exposed the incomplete service-binary build-input closure after completing the correct production build. All were structurally repaired; the matched measurement timing change is informational (+0.7%) and product semantics did not change | Validate this terminal tracker edit, then activate PERF-140 and capture its cache-off Incidents baseline |
| 2026-08-17T01:40:44-04:00 | PERF-140 `TODO` to `IN_PROGRESS` | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:32ba1bd74fbc923b13b7ea9d81017ebb49d9da42edd20b292e78b928489d48ce`; tracker-excluded digest `c45fc5e7e5f37d1d9fbc665b0ae12e03ed20cfa9` | Terminal PERF-130 tracker edit passed `make lint-markdown` at `.cartulary/test-results/20260817T054028Z-p2672087` and `git diff --check`; `make task-guide ROLE=module-author OWNER=module.incidents` selected the focused service-backed slice, with `CARTULARY_HARNESS_CACHE_MODE=off` as the controlled cache setting | Eager exact cleanup must follow application/background/pool closure; forced connection termination remains a bounded ownership-proved fallback and suite teardown remains recovery-only | Capture the full module.incidents service-backed baseline, inventory resource finalizers, then implement eager database and object-store retirement |
| 2026-08-17T02:03:41-04:00 | PERF-140 `IN_PROGRESS` to `DONE` | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:52c2abbe0611898e58d2e0df0f7d5895f5df8c2a74c539f36b5a6ab6e86065be`; tracker-excluded digest `5c5e84fed772b70fe6acaf4458bcbbd74889feba` | Adopted eager exact-owner cleanup; removed attached-suite retention events/lists; replaced unowned `NewDatabase` with cleanup-owning `NewDatabaseT`; added per-test `BootstrapBucketT`; retained explicit object-store package reuse; made normal PostgreSQL drop the primary path with a bounded, active-connection-only, ownership-proved termination fallback across ordinary, browser, and performance fixtures. The cache-off baseline passed 32/32 in 73.650s at `.cartulary/test-results/20260817T054117Z-p2673438`, with 78 peak non-template databases, 74 retention events, and zero hot-path drops. The identical after slice passed 32/32 in 75.144s at `20260817T055551Z-p2750251`: peak fell to 6 (92.3%), all 74 dedicated per-test databases used normal eager drops, zero retention events remained, and the four terminal identities were one transaction database plus three owned browser-stack fixtures. Migration recurrence passed 3/3 at `20260817T055959Z-p2814483` with all 7/7 databases normally dropped and an empty terminal ledger. Incident Bundles passed 3/3 at `20260817T060139Z-p2830406`, proving the per-test bucket and both databases cleaned while its package bucket remained reusable. Performance-fixture lifecycle passed 3/3 at `20260817T060229Z-p2845440`. Focused cleanup contracts passed at `20260817T055930Z-p2798817` and `20260817T055934Z-p2799265`; harness contract passed at `20260817T060106Z-p2829469`; drift, artifact policy, and JSON shapes passed at `20260817T060315Z-p2860167`, `p2860183`, and `p2860198` | The 2.0% matched wall-time regression is below the 5% ceiling. Rejected retaining successful resources, unconditional `DROP ... WITH (FORCE)`, and removing package-scoped object-store reuse. The first newly cataloged stub run at `20260817T055427Z-p2729512` failed because simulated dependency/policy inputs were omitted; the fixture was corrected without weakening admission. Internal callers must use `NewDatabaseT`; the retained-event token is removed without alias. Transaction and browser-stack identities remain intentionally longer-lived under their separate declared lifecycles, and all disappear with successful owned teardown; the private resource ledger was deleted in every successful evidence root | Run terminal tracker Markdown/diff validation, then activate PERF-150 and capture controlled release-inventory generation evidence |
| 2026-08-17T02:04:50-04:00 | PERF-150 `TODO` to `IN_PROGRESS` | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:52c2abbe0611898e58d2e0df0f7d5895f5df8c2a74c539f36b5a6ab6e86065be`; tracker-excluded digest `5c5e84fed772b70fe6acaf4458bcbbd74889feba` | Terminal PERF-140 tracker edit passed `make lint-markdown` at `.cartulary/test-results/20260817T060410Z-p2864263` and `git diff --check`; the accepted PERF-080 release timings are the before reference: inventory producer 120.701s, license 85.417s, and SBOM 73.892s exclusive | Public target names and canonical artifact paths remain fixed; semantic artifacts must become deterministic, only the paired producer may use cache v2, and every validator must still emit fresh current-run evidence | Inventory the three current graph/Make paths, capture one controlled cold real-producer run, then adopt and implement the paired deterministic producer contract |
| 2026-08-17T02:55:40-04:00 | PERF-150 `IN_PROGRESS` to `DONE` | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:7fb8759ae95393951376da9e7ceb4fa3ca80d0f7e4a5db622084eb56c070d15f`; tracker-excluded dirty-state digest `e235e03e459c5caf7f7531d422fc594cb1df3365` | Adopted `TH-HARNESS-REQ-817`/`AC-101`; made `release-inventory-artifacts` the sole paired producer; cut license reports to deterministic v2; derived the timestamp-free CycloneDX serial from the shared semantic input digest; removed the retained raw/Markdown/provenance side bundle; declared the exact two-file, mode-0644 repository-artifact output set; made `license-report` and `sbom` fresh non-cacheable validators; and enabled cache-v2 only for the producer. The controlled before generator used 113.24s, 526 files, and 10,139,830 bytes at `/tmp/cartulary-perf150-before.gEI6r9`. The final cold graph run passed in 4.64s at `.cartulary/test-results/20260817T065229Z-p3041073`, invoked the real producer once, and emitted only `license-report.json` (81,943 bytes, SHA-256 `a7e2bba27bb21d7e64072898f40ba6fee3542970a4f1e337dfb7b280d694449f`) and `sbom.cyclonedx.json` (55,748 bytes, SHA-256 `7a4261982a5b4180f63b7530e48bb3f1908778c85bcbb0271fee9a7242d23bb7`) for a 99.6% file and 98.6% byte reduction. After both destinations were moved aside, the 0.98s warm run at `20260817T065244Z-p3042157` restored byte-identical files transactionally, recorded `cache_hit/record_valid`, and emitted zero producer stdout/stderr. Deliberately changing the current cache payload from mode 0600 to 0644 produced `cache_miss/record_invalid`, quarantined the exact entry, reran the producer, and passed at `20260817T064657Z-p2923731`; the rebuilt entry then hit with zero producer output at `20260817T064727Z-p2924911`. Fresh v2/CycloneDX consumers passed at `20260817T065257Z-p3042842` and `20260817T065257Z-p3043008`; current-source release readiness passed 32/32 in 83.218s at `20260817T065326Z-p3043580`, with one producer cache hit and 31 explicitly non-cacheable units. `make generate` passed at `20260817T065153Z-p3036462`; the release-owner slice, Biome, final generation drift, artifact policy, JSON shapes, and harness contract passed at `20260817T065207Z-p3039572`, `p3039692`, `20260817T065455Z-p3147202`, `p3147221`, `p3147240`, and `p3147543` | Public targets and canonical paths are unchanged. `cartulary.license_report.v1`, incomplete outputs, separate producers, and old current readers are rejected without aliases; historical roots remain manual archives. Contract matrices cover manifest/lock/package/toolchain/scanner/schema/container/generator invalidation plus missing, surplus, corrupt, traversing, linked, wrong-mode, wrong-producer, wrong-destination, concurrent, and rollback cases. A dangling `LICENSE.md`/`README.md` fixture proves Markdown is neither read nor stated; the production snapshot also excludes Markdown before file access. Rejected dual readers/writers, cacheable consumers, retained raw/Markdown diagnostics, volatile UUID/timestamp/path fields, and host-dependent local-image scans. Invalid attempts are retained: `20260817T062833Z-p2885793` exposed the initial font-object schema mismatch; a first graph attempt had no unit root because the internal producer lacked a command ID; `20260817T063348Z-p2898241` exposed missing cache-mode forwarding. All were structurally repaired. Residual scanner availability/version variance is bounded by explicit identities and cache inputs; no broad repository timing claim is made until PERF-160 | Validate this terminal tracker edit with Markdown lint and diff checks, then activate PERF-160 with a new source/dirty checkpoint |
| 2026-08-17T02:56:50-04:00 | PERF-160 `TODO` to `IN_PROGRESS` | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:7fb8759ae95393951376da9e7ceb4fa3ca80d0f7e4a5db622084eb56c070d15f`; tracker-excluded dirty-state digest `e235e03e459c5caf7f7531d422fc594cb1df3365` | Terminal PERF-150 tracker edit passed `make lint-markdown` at `.cartulary/test-results/20260817T065628Z-p3152062` and `git diff --check`; PERF-090 through PERF-150 are all `DONE`, with no mandatory `BLOCKED` or `DROPPED` slice | Broad verification must preserve exact 1,201-row closure, bounded/redacted evidence, eager cleanup, secure deterministic inventory restore, one production fixture construction, and the unchanged public target surface. No successful full warm current-source `check` root exists yet, so `make agent-finalize` will run with `RESULTS_DIR` unset | Reconcile current source/generated/documentation state, run final focused gates and finalization, then run `make check` followed immediately by warm `make release-check` |
| 2026-08-17T03:07:16-04:00 | PERF-160 `IN_PROGRESS` to `TODO`; PERF-140 `DONE` to `IN_PROGRESS` | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:7fb8759ae95393951376da9e7ceb4fa3ca80d0f7e4a5db622084eb56c070d15f`; tracker-excluded dirty-state digest `e235e03e459c5caf7f7531d422fc594cb1df3365` | PERF-160 preflight drift, artifact-policy, JSON-shape, harness-contract, Markdown, and diff gates passed at `.cartulary/test-results/20260817T065751Z-p3153580`, `p3153586`, `p3153609`, `p3154023`, and `p3153983`; `make agent-finalize` passed at `20260817T065819Z-p3158734` and recorded `results-dir-not-provided` for retained-run maintenance. The first current-source `make check` failed 772/776 in 478.109s at `.cartulary/test-results/20260817T065841Z-p3161699`. Its four terminal failures reduce to one compile defect: `internal/testutil/httptestx/httptestx_test.go` retains an unused `context` import after PERF-140 replaced manual bucket cleanup with `BootstrapBucketT`; the same compile error caused `go-vulncheck`, `lint-go`, one raw Go unit, and one selector result to fail | This is a deletion-completeness defect in PERF-140, not an acceptable broad-run variance. All service cleanup completed and no evidence suggests a PERF-150 cache or release-inventory defect. The failed broad run is retained and excluded from acceptance | Remove the stale import, run the exact affected testutil owner/build/lint checks, reclose PERF-140, validate its tracker edit, then reactivate PERF-160 with the new source identity |
| 2026-08-17T03:10:01-04:00 | PERF-140 `IN_PROGRESS` to `TODO`; PERF-120 `DONE` to `IN_PROGRESS` | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:3a12cf2f937b02e7f90c9fb767e793827fe9f61c0d85ac3bd05b6999c276d5e1`; tracker-excluded dirty-state digest `94881e7f6d84e2c27fe7b8b3713e70b3291cd7a2` | Removed the stale PERF-140 import; the exact previously missing Projections selector passed 3/3 at `.cartulary/test-results/20260817T070815Z-p3333234`, and `make go-vulncheck` passed 4/4 at `20260817T070903Z-p3348470`. The paired `make lint-go` rerun then found `internal/testutil/suiteservices/diagnostics.go:sanitizeFileComponent` unused after PERF-120 removed one-file-per-event naming; its retained target-local root is `.cartulary/test-results/20260817T070907Z-p3352826` | The staticcheck failure is a PERF-120 deletion-completeness defect and must be repaired before PERF-140 can reclose. The vulnerability findings were policy-accepted and the target passed; they are unrelated to this refactor | Remove the obsolete filename sanitizer, rerun suiteservices tests and Go lint, reclose PERF-120, then resume the already-focused PERF-140 repair |
| 2026-08-17T03:12:39-04:00 | Reopened PERF-120 `IN_PROGRESS` to `DONE` | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:e853c9f0ce1206f286215de87bfdf28252099ac8a0637bbe48f6e29789be9b6c`; tracker-excluded dirty-state digest `dae365810d555b43d51719277431082d6a75d463` | Deleted the obsolete per-event filename sanitizer. `make lint-go` then passed format, vet, and staticcheck at `.cartulary/test-results/20260817T071109Z-p3357579`, `20260817T071111Z-p3360891`, and `20260817T071115Z-p3364522`; the complete `make backend-unit` closure passed 107/107 in 42.330s at `20260817T071146Z-p3366342`. An attempted nonexistent internal target, `make backend-unit-suiteservices`, failed before execution and produced no retained run; the public `backend-unit` graph supplied the valid package-level closure | Journal/scope behavior, evidence size, schemas, and compatibility are unchanged; this is deletion-only completion. No obsolete event filename helper remains | Validate this terminal tracker edit, then reactivate PERF-140 and finish its already-applied import repair |
| 2026-08-17T03:13:18-04:00 | PERF-140 `TODO` to `IN_PROGRESS` after PERF-120 repair | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:e853c9f0ce1206f286215de87bfdf28252099ac8a0637bbe48f6e29789be9b6c`; tracker-excluded dirty-state digest `dae365810d555b43d51719277431082d6a75d463` | Reopened PERF-120 terminal tracker edit passed `make lint-markdown` at `.cartulary/test-results/20260817T071303Z-p3388314` and `git diff --check`; the PERF-140 stale-import repair remains applied, and its exact selector plus vulnerability scan already pass | No eager-cleanup behavior changed; this reopen exists to close the compile surface and preserve serial tracker authority | Rerun Go lint on the combined current source, record the repair terminally, and validate PERF-140 before resuming PERF-160 |
| 2026-08-17T03:13:52-04:00 | Reopened PERF-140 `IN_PROGRESS` to `DONE` | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:e853c9f0ce1206f286215de87bfdf28252099ac8a0637bbe48f6e29789be9b6c`; tracker-excluded dirty-state digest `dae365810d555b43d51719277431082d6a75d463` | Removed the unused `context` import left by conversion to `BootstrapBucketT`. The exact affected selector passed 3/3 at `.cartulary/test-results/20260817T070815Z-p3333234`; `make go-vulncheck` passed 4/4 at `20260817T070903Z-p3348470`; combined-source `make backend-unit` passed 107/107 at `20260817T071146Z-p3366342`; and the final format, vet, and staticcheck targets passed at `20260817T071341Z-p3389590`, `20260817T071343Z-p3392906`, and `20260817T071344Z-p3393239` | Eager resource behavior, matched performance, cleanup authority, and compatibility remain unchanged; this is deletion-only compile closure | Validate this terminal tracker edit, then reactivate PERF-160 with the repaired source identity and repeat its preflight/finalize/full-run sequence |
| 2026-08-17T03:14:29-04:00 | PERF-160 `TODO` to `IN_PROGRESS` after reopened-slice repairs | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:e853c9f0ce1206f286215de87bfdf28252099ac8a0637bbe48f6e29789be9b6c`; tracker-excluded dirty-state digest `dae365810d555b43d51719277431082d6a75d463` | Reopened PERF-140 terminal tracker edit passed `make lint-markdown` at `.cartulary/test-results/20260817T071414Z-p3394047` and `git diff --check`; PERF-090 through PERF-150 are again all `DONE`, with no mandatory `BLOCKED` or `DROPPED` slice | The failed 772/776 broad root remains an invalid attempt. Acceptance requires a new clean `check`, not retained partial work | Repeat required preflight and `agent-finalize`, then run a new `make check` followed immediately by warm `make release-check` |
| 2026-08-17T03:39:21-04:00 | PERF-160 `IN_PROGRESS` to `TODO`; PERF-140 `DONE` to `IN_PROGRESS` again | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:e853c9f0ce1206f286215de87bfdf28252099ac8a0637bbe48f6e29789be9b6c`; tracker-excluded dirty-state digest `dae365810d555b43d51719277431082d6a75d463` | Repeated preflight gates passed at `.cartulary/test-results/20260817T071455Z-p3395285`, `p3395307`, `p3395342`, `p3395720`, and `p3395739`; `make agent-finalize` passed at `20260817T071516Z-p3400451`. The repaired `make check` then passed 776/776 in 330.614s at `.cartulary/test-results/20260817T071534Z-p3403346`. The immediately warm `make release-check` completed inside the 875.326s ceiling but failed 945/973 in 819.728s at `.cartulary/test-results/20260817T072114Z-p3520318` | Two PERF-140 cleanup defects own the failure. `TestStateStore_Integration_LogicalPortAndScope` waited the full 15-second normal-drop context on a live connection, so the fallback inherited an already-expired context. All terminal stateful, accessibility, support, and visual browser groups then failed immediate stack attachment because eager browser cleanup removed the newly published `stack-v6.json` and compact admission before the runner consumed them; measurement and webserver-backed stacks survived. Timing is not accepted without correctness. The failed release root is retained and excluded | Give normal and forced PostgreSQL cleanup separate bounded budgets after exact ownership proof; preserve immutable retained browser evidence while eagerly deleting only live database, bucket, process, port, and private runtime authority; then rerun focused cleanup, browser, and exact failed database cases before reclosing PERF-140 |
| 2026-08-17T04:04:33-04:00 | Reopened PERF-140 `IN_PROGRESS` to `DONE` | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:aa6f0732f9246c17c3fe7240d040b7740517426acad7078c80e9e369260658a5`; tracker-excluded dirty-state digest `e4d12bf85c30ff3dd8f5359d7d03d7ad082b928a` | Gave the ordinary PostgreSQL drop a five-second child budget and the forced fallback a fresh 15-second child budget inside a 20-second owner finalizer; fallback is allowed only after exact ownership and an active-connection error or bounded normal-attempt expiry while the parent remains active, and parent cancellation still prevents forcing. Removed commandless `browser_lifecycle:*` units so the first executable group owns the stack lease during stack-v6/admission consumption; reset and later groups preserve healthy affinity reuse and the terminal group still releases it. The exact PostgreSQL catalog row passed at `.cartulary/test-results/20260817T080208Z-p3972024`; the pre-change browser-only aggregate passed 67/67 at `20260817T074857Z-p3809129`, proving the missing-file defect required broader overlap; after the structural handoff removal, accessibility passed 12/12 at `20260817T075743Z-p3882476` and stateful reuse passed 34/34 at `20260817T075913Z-p3922487`. `make generate` passed at `20260817T080135Z-p3964721`; drift, artifact policy, JSON shapes, Biome, and the final harness contract passed at `20260817T080148Z-p3967658`, `20260817T080200Z-p3970705`, `20260817T080200Z-p3970689`, `20260817T080200Z-p3970875`, and `20260817T080416Z-p3974441` | Public Make targets, selected rows, browser runtime/session identity, retained stack-v6/admission schemas, isolation, reset semantics, and terminal cleanup remain stable. Internal synthetic `browser_lifecycle:*` unit identities are intentionally removed, reducing accessibility by two units and eliminating the zero-reference pre-consumer interval. Rejected file-read retries, accepting missing evidence, and keeping the synthetic unit behind a health-check patch because each would preserve an unnecessary ownership transfer. The attempted undocumented `make harness-smoke-browser-work-graph` target failed before execution and produced no retained run; `make harness-contract` is the valid public gate. Broad release composition remains the PERF-160 acceptance gate | Validate this terminal tracker edit with Markdown lint and diff checks, then reactivate PERF-160 with the current source/dirty checkpoint and repeat preflight, finalization, `check`, and immediate warm `release-check` |
| 2026-08-17T04:05:31-04:00 | PERF-160 `TODO` to `IN_PROGRESS` after second PERF-140 repair | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:aa6f0732f9246c17c3fe7240d040b7740517426acad7078c80e9e369260658a5`; tracker-excluded dirty-state digest `e4d12bf85c30ff3dd8f5359d7d03d7ad082b928a` | Reopened PERF-140 terminal tracker edit passed `make lint-markdown` at `.cartulary/test-results/20260817T080515Z-p3975294` and `git diff --check`; PERF-090 through PERF-150 are again all `DONE`, with no mandatory `BLOCKED` or `DROPPED` workstream | Prior broad failures remain retained invalid attempts. Acceptance requires a fresh clean pair on this exact source; browser unit cardinality is expected to fall because synthetic readiness units were intentionally removed, while the 1,201-row owner closure remains exact | Repeat all PERF-160 preflight gates, run `make agent-finalize` with `RESULTS_DIR` unset, then run `make check` followed immediately by warm `make release-check` |
| 2026-08-17T04:30:18-04:00 | PERF-160 `IN_PROGRESS` to `TODO`; PERF-140 `DONE` to `IN_PROGRESS` a third time | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:aa6f0732f9246c17c3fe7240d040b7740517426acad7078c80e9e369260658a5`; tracker-excluded dirty-state digest `e4d12bf85c30ff3dd8f5359d7d03d7ad082b928a` | PERF-160 preflight passed at `.cartulary/test-results/20260817T080617Z-p3976577`, `20260817T080628Z-p3979482`, `20260817T080633Z-p3979937`, `20260817T080639Z-p3980466`, and `20260817T080657Z-p3981082`; `make agent-finalize` passed with retained-run maintenance skipped at `20260817T080709Z-p3981985`. `make check` then passed 776/776 in 476.901s at `.cartulary/test-results/20260817T080728Z-p3984871`. The immediately warm release finished inside the ceiling but failed 906/951 in 845.720s at `.cartulary/test-results/20260817T081534Z-p4153658` | Four owned-database finalizers exhausted both normal and terminate-then-drop budgets under concurrent catalog pressure. Browser stack publication returned success without producing `stack-v6.json` whenever required readiness timestamps were absent, so 31 groups failed closed on nonexistent attachment evidence even after direct first-consumer ownership. Separately, the sole release-inventory producer raced recursive Go package discovery against a transient `tmp/work-graph-cache-contract.*` test directory; that PERF-150 defect will be reopened only after PERF-140 is repaired and terminally validated. The release timing is invalid despite meeting the ceiling | Make the ownership-proved PostgreSQL forced fallback one bounded atomic operation; make browser publication fail unless the complete stack/admission pair exists before provider return; validate exact cleanup and broad-overlap browser paths, close PERF-140, then serially reopen PERF-150 for repository input discovery closure |
| 2026-08-17T04:46:13-04:00 | Reopened PERF-140 `IN_PROGRESS` to `DONE` a third time | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:08aab2df735653a7416e15e8dead536cfca4a84ffcc0541f523aa38ad645d3bf`; tracker-excluded dirty-state digest `2c9be3f2c5a082d8277f1a77a7b55e1c223b6604` | Replaced split terminate-then-drop fallbacks with one ownership-proved PostgreSQL `DROP DATABASE ... WITH (FORCE)` operation under the fresh bounded fallback context across ordinary, browser, and performance-fixture cleanup. Browser readiness is now bound after each successful top-level probe; missing bindings fail publication, and the broker verifies the complete owner-only mode-0600 stack-v6/admission pair before returning or cleans the rejected session. The exact pgtest contract passed at `.cartulary/test-results/20260817T083913Z-p216780`; all four release-failing live rows passed at `20260817T083942Z-p217882`, `20260817T084021Z-p232910`, and `20260817T084100Z-p247771`; harness contract passed at `20260817T083924Z-p217325`; accessibility passed 12/12 with two complete five-file session sets at `20260817T084143Z-p262641`; stateful affinity/reset coverage passed 34/34 at `20260817T084314Z-p302713`; Go format, vet, and staticcheck passed at `20260817T084534Z-p345175`, `20260817T084536Z-p348493`, and `20260817T084541Z-p352752`; Biome passed at `20260817T084550Z-p354607` | Public targets, schema identities, row closure, normal-first cleanup, exact ownership proof, cancellation behavior, affinity reuse, and retained evidence stay unchanged. Atomic forced drop is the supported PostgreSQL operation; split terminate-then-drop is intentionally removed. Diagnostics-only browser sessions can no longer be treated as attachable, and failed postcondition cleanup preserves both publication and cleanup errors. The prior release failure remains an invalid attempt | Validate this terminal tracker edit with Markdown lint and diff checks, then reopen PERF-150 before changing release-inventory discovery |
| 2026-08-17T04:47:15-04:00 | PERF-150 `DONE` to `IN_PROGRESS` for repository input discovery closure | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:08aab2df735653a7416e15e8dead536cfca4a84ffcc0541f523aa38ad645d3bf`; tracker-excluded dirty-state digest `2c9be3f2c5a082d8277f1a77a7b55e1c223b6604` | Reopened PERF-140 terminal tracker maintenance passed `make lint-markdown` at `.cartulary/test-results/20260817T084655Z-p355570` and `git diff --check`. The controlled before case is retained in the failed release root `.cartulary/test-results/20260817T081534Z-p4153658`: the sole inventory producer failed when recursive `go list ... ./...` traversed `tmp/work-graph-cache-contract.uzjiUn` while its owner test removed that transient directory | Canonical inventory must cover the stable authored Go package closure without reading repository-local runtime, cache, or transient test roots. This is producer input-discovery closure, not authorization for retrying filesystem races or weakening inventory completeness | Adopt stable package discovery, add a deterministic transient-root race fixture, rerun cold/miss and warm/hit producer contracts plus consumers, then close and validate PERF-150 before resuming PERF-160 |
| 2026-08-17T04:51:05-04:00 | Reopened PERF-150 `IN_PROGRESS` to `DONE` | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:7c8e89fae660f14899bf3908e9709787c484360090dc392f8a1def774e924700`; tracker-excluded dirty-state digest `73c2885f3d0e065e620b1a23f8f9f0dc7c53d709` | Replaced recursive repository-root test dependency discovery with the explicit stable `./cmd/...`, `./db/...`, `./internal/...`, and `./tools/...` closure and updated the canonical evidence label. A deterministic release-owner fixture placed an invalid Go source beneath `tmp/work-graph-cache-contract.*` and proved the exact production arguments neither discover nor stat it. The release-owner slice passed at `.cartulary/test-results/20260817T084923Z-p357675`. A real current-source cache miss invoked the paired producer once and passed at `20260817T084935Z-p358277`; semantic input remained `sha256:deb4773c1a2726fe062b9fd972f4ca319fba6bec601615a83f36d7b9e06f199b`. The immediate warm run passed in 322ms with `cache_hit/record_valid`, empty producer logs, and output digest `sha256:35ccda29d5d7cb2aeb2ee58161d99e29da5be07f9f173ee87c581eb7b0db1c45` at `20260817T084951Z-p359364`. Fresh license and SBOM consumers passed at `20260817T085001Z-p359985` and `20260817T085011Z-p360601`; release-evidence contract, Biome, and harness contract passed at `20260817T085026Z-p361240`, `20260817T085039Z-p361613`, and `20260817T085047Z-p362140` | Canonical paths, v2/CycloneDX bytes, semantic digest, deterministic UUID, cache policy, public targets, and dependency completeness remain stable. Runtime/transient roots are intentionally outside package discovery; any future first-party top-level Go package root requires an explicit producer and cache-input change. Rejected retries and filesystem-wide discovery because both preserve the race. The failed broad release remains an invalid attempt | Validate this terminal tracker edit with Markdown lint and diff checks, then activate PERF-160 with the new source and dirty-state checkpoint |
| 2026-08-17T04:51:52-04:00 | PERF-160 `TODO` to `IN_PROGRESS` after serial broad-run repairs | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:7c8e89fae660f14899bf3908e9709787c484360090dc392f8a1def774e924700`; tracker-excluded dirty-state digest `73c2885f3d0e065e620b1a23f8f9f0dc7c53d709` | Reopened PERF-150 terminal tracker maintenance passed `make lint-markdown` at `.cartulary/test-results/20260817T085133Z-p362944` and `git diff --check`; PERF-090 through PERF-150 are `DONE`, with no mandatory `BLOCKED` or `DROPPED` slice | Prior broad failures remain retained invalid attempts. No successful full `check` root exists for this exact source, so `make agent-finalize` will run with `RESULTS_DIR` unset before the acceptance pair | Run all required preflight gates, finalization, one fresh successful `make check`, and an immediately warm successful `make release-check`; reopen the owning slice for any related defect |
| 2026-08-17T04:52:38-04:00 | PERF-160 `IN_PROGRESS` to `TODO`; PERF-150 `DONE` to `IN_PROGRESS` again | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:7c8e89fae660f14899bf3908e9709787c484360090dc392f8a1def774e924700`; tracker-excluded dirty-state digest `73c2885f3d0e065e620b1a23f8f9f0dc7c53d709` | The first PERF-160 preflight target failed 3/4 at `.cartulary/test-results/20260817T085219Z-p364097` because the changed release-inventory generator correctly invalidated `tools/execution_topology_render_index.json` | This is a PERF-150 generated-projection completeness defect. No product, cache, cleanup, or timing assertion ran, and the failed preflight root is excluded from acceptance | Run owner-driven `make generate`, then generation drift, generated-artifact policy, JSON shape, and harness-contract gates; terminally validate PERF-150 before reactivating PERF-160 |
| 2026-08-17T04:53:51-04:00 | Reopened PERF-150 `IN_PROGRESS` to `DONE` after projection repair | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:9e9ea9a9515cf4d16520f5ae53be8d8b82f6d4612c3f25eaa16f7cac33c0ee67`; tracker-excluded dirty-state digest `ee2eae42b1b214089b9ad1349f028955fde89b03` | Owner-driven `make generate` refreshed the topology render index at `.cartulary/test-results/20260817T085301Z-p367264`. Generation drift then passed 4/4 at `20260817T085313Z-p370220`; generated-artifact policy passed 3/3 at `20260817T085323Z-p373177`; JSON shapes passed 3/3 at `20260817T085328Z-p373632`; and harness contract passed 2/2 at `20260817T085334Z-p374173` | Only the owner-generated source inventory/digests changed. Release semantics, canonical bytes, cache behavior, public targets, and row closure remain as recorded in the preceding PERF-150 terminal entry. The failed drift root remains an invalid attempt | Validate this terminal tracker edit with Markdown lint and diff checks, then reactivate PERF-160 with the generated projection included in source identity |
| 2026-08-17T04:54:29-04:00 | PERF-160 `TODO` to `IN_PROGRESS` after generated projection repair | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:9e9ea9a9515cf4d16520f5ae53be8d8b82f6d4612c3f25eaa16f7cac33c0ee67`; tracker-excluded dirty-state digest `ee2eae42b1b214089b9ad1349f028955fde89b03` | The PERF-150 projection terminal checkpoint passed `make lint-markdown` at `.cartulary/test-results/20260817T085411Z-p374951` and `git diff --check`; all prerequisite slices are `DONE` | The earlier PERF-160 drift failure is retained and excluded. Acceptance requires a complete fresh preflight plus current-source full pair | Repeat all required preflight gates, run `make agent-finalize` with `RESULTS_DIR` unset, then run `make check` followed immediately by warm `make release-check` |
| 2026-08-17T05:04:32-04:00 | PERF-160 `IN_PROGRESS` to `TODO`; PERF-140 `DONE` to `IN_PROGRESS` for fixture lint closure | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:9e9ea9a9515cf4d16520f5ae53be8d8b82f6d4612c3f25eaa16f7cac33c0ee67`; tracker-excluded dirty-state digest `ee2eae42b1b214089b9ad1349f028955fde89b03` | PERF-160 preflight passed at `.cartulary/test-results/20260817T085451Z-p376100`, `20260817T085503Z-p379044`, `20260817T085506Z-p379499`, `20260817T085513Z-p380045`, and `20260817T085529Z-p380640`; `make agent-finalize` passed at `20260817T085546Z-p381620`. The fresh `make check` then failed 775/776 in 477.422s at `.cartulary/test-results/20260817T085602Z-p384513` solely because ShellCheck reported `SC2016` on the intentionally literal multiline browser-publication assertion | All product rows passed; no release run was started. This is a PERF-140 test-fixture lint completeness defect, not a product, cleanup, evidence, or timing failure. The failed check root is retained and excluded | Mark the literal assertion for ShellCheck, run `make lint-shell` and harness contract, reclose and validate PERF-140, then reactivate PERF-160 with a fresh preflight and full pair |
| 2026-08-17T05:05:30-04:00 | Reopened PERF-140 `IN_PROGRESS` to `DONE` after fixture lint repair | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:6626c351f7196dbdee1645a57834e0d44248809d22fd1d319273fefd5e9c92d1`; tracker-excluded dirty-state digest `7918b6d93fb4ff3dc4b36d97763db7dc9c683452` | Marked the intentionally non-expanding source-pattern assertion with the exact ShellCheck suppression. `make lint-shell` passed 4/4 at `.cartulary/test-results/20260817T090500Z-p548154`, and harness contract passed 2/2 at `20260817T090513Z-p549154` | Browser runtime behavior, publication validation, schemas, and evidence are unchanged; this is fixture-lint completion only | Validate this terminal tracker edit, then reactivate PERF-160 and repeat the full required sequence |
| 2026-08-17T05:06:09-04:00 | PERF-160 `TODO` to `IN_PROGRESS` after fixture lint repair | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:6626c351f7196dbdee1645a57834e0d44248809d22fd1d319273fefd5e9c92d1`; tracker-excluded dirty-state digest `7918b6d93fb4ff3dc4b36d97763db7dc9c683452` | PERF-140 terminal tracker maintenance passed `make lint-markdown` at `.cartulary/test-results/20260817T090551Z-p549911` and `git diff --check`; all prerequisite slices are `DONE` | The 775/776 root is retained and excluded. A fully repeated preflight and broad pair remain mandatory | Repeat preflight, `agent-finalize` without retained check evidence, one fresh `make check`, and its immediate warm `make release-check` |
| 2026-08-17T05:32:07-04:00 | PERF-160 `IN_PROGRESS` to `TODO`; PERF-140 `DONE` to `IN_PROGRESS` after the warm release | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:6626c351f7196dbdee1645a57834e0d44248809d22fd1d319273fefd5e9c92d1`; tracker-excluded dirty-state digest `7918b6d93fb4ff3dc4b36d97763db7dc9c683452` | Repeated preflight passed at `.cartulary/test-results/20260817T090630Z-p551051`, `20260817T090641Z-p553973`, `20260817T090645Z-p554428`, `20260817T090651Z-p554968`, and `20260817T090706Z-p555571`; `make agent-finalize` passed at `20260817T090724Z-p556550`; `make check` passed 776/776 in 460.924s at `20260817T090740Z-p559426`. The immediate warm release then failed 875/951 in 971.850s at `.cartulary/test-results/20260817T091529Z-p720191` | Four exact-owner database finalizers still exhausted the atomic forced-drop context under extreme catalog overlap. Browser publication postconditions passed initially, but later exact session starts replaced the same retained directories with diagnostics-only evidence, causing stack-v6 attachment failures across webserver-backed, stateful, accessibility, support, and visual work. The result exceeds the 875.326s ceiling and is invalid | Identify and remove the duplicate browser-session start path; reduce PostgreSQL catalog-operation overlap structurally rather than extending timeouts; rerun controlled overlap evidence before reclosing PERF-140 |
| 2026-08-17T05:56:17-04:00 | Reopened PERF-140 `IN_PROGRESS` to `DONE` after scale-failure diagnosis | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:f5de3c347cc152c6896a3d8e81ec44177c5f01b31e680d8777822bff59c270b3`; tracker-excluded dirty-state digest `a5c5093ab585feccf2c1fcf3ecf78a62baa3ff9d` | Added one service-local PostgreSQL session advisory lock around only the exact-owner atomic forced-drop fallback, keeping ordinary drops and borrower execution concurrent. Browser admission, lease, stack, and terminal checks now propagate failure explicitly through timing wrappers; session artifact identities are single-use and cannot clear or overwrite retained evidence. The platform PostgreSQL slices passed at `.cartulary/test-results/20260817T094446Z-p1025678` and `20260817T094556Z-p1041591`; accessibility passed 12/12 at `20260817T094641Z-p1056279`; stateful reuse passed 34/34 at `20260817T094808Z-p1096297`; the four previously failing rows passed at `20260817T095144Z-p1139547`, `20260817T095224Z-p1154450`, and `20260817T095302Z-p1169299`. Owner generation passed at `20260817T095359Z-p1184086`; generation drift, artifact policy, JSON shapes, and the terminal harness contract passed at `20260817T095517Z-p1199710`, `20260817T095531Z-p1202629`, `20260817T095535Z-p1203090`, and `20260817T095543Z-p1203613`; `make lint-go` passed. An initial `make lint-shell` exposed only the new source-fixture warnings at `20260817T095430Z-p1197467`; after removing the unused variable and marking literal assertions, it passed 4/4 at `20260817T095502Z-p1198786` | Exact ownership, normal-first behavior, cancellation, schema identities, public targets, browser affinity, and retained evidence remain stable. Extending cleanup timeouts, retrying missing files, and accepting partial evidence remain rejected. Investigation corrected the duplicate-start hypothesis: a non-gate direct schema diagnostic showed the late scope-v2 summary had 40 strategy aggregates against its declared maximum of 32; timing wrappers swallowed the resulting admission error. That is a distinct PERF-120 bounded-summary defect and the direct invocation is diagnostic only, not accepted validation evidence | Validate this terminal tracker edit with Markdown lint and diff checks, then reopen PERF-120 with the failed release scope as its controlled before case; cap aggregate diagnostics deterministically while preserving total counts, and use Make-owned gates before resuming PERF-160 |
| 2026-08-17T05:57:38-04:00 | PERF-120 `DONE` to `IN_PROGRESS` for bounded strategy-summary closure | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:f5de3c347cc152c6896a3d8e81ec44177c5f01b31e680d8777822bff59c270b3`; tracker-excluded dirty-state digest `a5c5093ab585feccf2c1fcf3ecf78a62baa3ff9d` | The PERF-140 terminal tracker edit passed `make lint-markdown` at `.cartulary/test-results/20260817T095654Z-p1204644` and `git diff --check`. Controlled before evidence is the invalid late summary in `.cartulary/test-results/20260817T091529Z-p720191`: it retained the correct `fixture.total_count=2042` and ten slowest diagnostics but emitted 40 `fixture.by_strategy` entries against scope-v2's declared maximum of 32. The first browser admission after that transition failed, and explicit propagation added in PERF-140 now prevents recurrence from appearing as a successful diagnostics-only session | Aggregate diagnostics must remain bounded independently of suite diversity. Full total counts remain exact; retained strategy examples may be capped only by a deterministic, documented ordering. Increasing the schema bound, dropping total counts, accepting invalid scope, or removing admission validation are rejected | Specify and implement a deterministic top-32 strategy projection with full aggregate totals, add an over-bound fixture proving stable selection and schema validity, rerun service diagnostics/security and focused browser gates, then close and validate PERF-120 before resuming PERF-160 |
| 2026-08-17T06:07:04-04:00 | Reopened PERF-120 `IN_PROGRESS` to `DONE` | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:5003cbedd4ddf5f6def40646e2a206b6d5940757339de8b2ecb8bfec8ad65585`; tracker-excluded dirty-state digest `98d8b53b2c709bd1d61b64fdb1165d752c0c37e1` | Scope-v2 now carries exact fixture totals and `strategy_aggregate_count`, while `by_strategy` retains at most 32 aggregates ordered by descending duration, descending count, and the closed ASCII identity tuple. A 40-strategy producer fixture proves totals of 40 events/780ms, exact distinct count 40, and stable retained endpoints `strategy-39` through `strategy-08`. `make test-fast` passed 405/405 at `.cartulary/test-results/20260817T100104Z-p1211362`; owner generation passed at `20260817T100313Z-p1253667`; a real accessibility browser admission passed 12/12 with two complete five-file session sets at `20260817T100328Z-p1256742`; and the small service-backed PostgreSQL slice passed at `20260817T100501Z-p1297272`. Generation drift, artifact policy, JSON shapes, harness contract, and shell lint passed at `20260817T100550Z-p1312110`, `20260817T100601Z-p1315018`, `20260817T100607Z-p1315479`, `20260817T100614Z-p1316002`, and `20260817T100644Z-p1325053`; `make lint-go` passed | Scope-v2 remains the current identity because this required field completes the still-unreleased cutover; there is no compatibility reader, alias, or historical-root input. Total counts, failure counts, ten slowest activities, journal integrity, admission validation, and security posture are unchanged. Raising the bound, schema-invalid publication, nondeterministic map truncation, and exhaustive retained strategy inventories remain rejected. The broad over-bound composition check remains mandatory in PERF-160 | Validate this terminal tracker edit with Markdown lint and diff checks, then activate PERF-160 with the new source identity and repeat every preflight, finalization, full-check, and immediate warm-release gate |
| 2026-08-17T06:07:51-04:00 | PERF-160 `TODO` to `IN_PROGRESS` after reopened evidence and cleanup repairs | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:5003cbedd4ddf5f6def40646e2a206b6d5940757339de8b2ecb8bfec8ad65585`; tracker-excluded dirty-state digest `98d8b53b2c709bd1d61b64fdb1165d752c0c37e1` | The reopened PERF-120 terminal edit passed `make lint-markdown` at `.cartulary/test-results/20260817T100731Z-p1326280` and `git diff --check`; PERF-090 through PERF-150 are `DONE`, with no mandatory `BLOCKED` or `DROPPED` slice. The preceding successful full check belongs to the superseded source, so retained-run maintenance cannot use it for this current-source finalization | Earlier broad failures and the direct schema diagnostic remain recorded invalid attempts. Acceptance requires a fresh Make-owned over-bound release composition, one complete current-source check, exact 1,201-row closure, and the 875.326s warm release ceiling without weaker correctness, cleanup, isolation, security, or evidence assertions | Run all required preflight gates, `make agent-finalize` with `RESULTS_DIR` unset, a fresh `make check`, and its immediate warm `make release-check`; reopen the exact owning slice for any related defect |
| 2026-08-17T06:33:03-04:00 | PERF-160 `IN_PROGRESS` to `TODO`; PERF-140 `DONE` to `IN_PROGRESS` for catalog-drop coordination | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:5003cbedd4ddf5f6def40646e2a206b6d5940757339de8b2ecb8bfec8ad65585`; tracker-excluded dirty-state digest `98d8b53b2c709bd1d61b64fdb1165d752c0c37e1` | PERF-160 preflight passed at `.cartulary/test-results/20260817T100820Z-p1327411`, `20260817T100836Z-p1330340`, `20260817T100841Z-p1330795`, `20260817T100847Z-p1331319`, and `20260817T100902Z-p1331898`; `make agent-finalize` passed with `RESULTS_DIR` unset at `20260817T100919Z-p1332863`; and `make check` passed 776/776 in 350.508s at `20260817T100942Z-p1335769`. The immediate warm release failed 947/951 in 934.479s at `.cartulary/test-results/20260817T101539Z-p1458957`, 59.153s above the ceiling | Browser evidence is complete and valid. Four exact-owner finalizers exhausted their 15-second forced-drop contexts while successful forced fallbacks queued for 8.929s, 9.099s, and 17.636s. The existing advisory lock serializes only fallback operations, leaving ordinary database drops to sustain the same catalog contention and making later fallbacks spend their budgets waiting. The failed release root is the controlled before case and is excluded from acceptance | Coordinate every cooperating normal and forced database drop with the same service-local advisory lock, release the lock explicitly with a fresh bounded context, preserve normal-first behavior and independent operation budgets, then rerun focused overlap, cleanup, and broad timing evidence before reclosing PERF-140 |
| 2026-08-17T07:11:32-04:00 | Reopened PERF-140 `IN_PROGRESS` to `DONE` after bounded catalog-concurrency repair | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:a5f57074019f762435599b1da55dc6e8c29292a7c459d321539d7e4f552b5e13`; tracker-excluded dirty-state digest `d0c45f8747f780d148c4a2da6d1a5373d2135cd6` | Centralized exact-owner normal-first cleanup in `internal/testutil/postgrescleanup`: ordinary drops retain their separate five-second budget and remain concurrent; active or timed-out cases use one atomic `DROP DATABASE ... WITH (FORCE)` under a service-local two-slot advisory semaphore; available-slot selection does not poll, a saturated request waits on one deterministic slot, the forced budget starts after admission, and unlock uses a fresh bounded context. The platform service-backed slice passed at `.cartulary/test-results/20260817T110408Z-p1839227`; unit contracts passed 405/405 at `20260817T110335Z-p1833929`. The controlled 221-unit backend overlap passed at `20260817T110458Z-p1854320` in 239.767s with all 418 database drops, 11 forced fallbacks, 351.170s aggregate drop time, 13.134s maximum drop time, and no terminal recovery ledger. This is 39.7% faster than the first 397.420s serialized-operation candidate and 44.2% faster than the 429.719s single-sequence-lock candidate. Go format, vet, and staticcheck passed at `20260817T111007Z-p1899227`, `20260817T111009Z-p1902549`, and `20260817T111014Z-p1906888`; generation drift, artifact policy, JSON shapes, and harness contract passed at `20260817T111030Z-p1908732`, `20260817T111040Z-p1911642`, `20260817T111044Z-p1912097`, and `20260817T111053Z-p1912649` | Exact ownership proof, parent cancellation, normal-first behavior, independent child budgets, atomic forced deletion, eager resource lifetime, public targets, schema identities, and stale recovery remain unchanged. The first service-backed attempt at `20260817T105304Z-p1772628` stopped before rows because the refactor left one unused import; the first staticcheck attempt at `20260817T110921Z-p1892765` found one superseded identifier validator. Both were repaired without behavioral changes. Rejected the one-lock-per-operation candidate because 66 attempts lost their queue position and amplified fallback work, and rejected the single global cleanup lease because it serialized the ordinary hot path. Extending timeouts and weakening cleanup remain rejected | Validate this terminal tracker edit with Markdown lint and diff checks, then reactivate PERF-160 and repeat the complete preflight, finalization, full-check, and immediate warm-release sequence on the new source identity |
| 2026-08-17T07:12:23-04:00 | PERF-160 `TODO` to `IN_PROGRESS` after bounded cleanup repair | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:a5f57074019f762435599b1da55dc6e8c29292a7c459d321539d7e4f552b5e13`; tracker-excluded dirty-state digest `d0c45f8747f780d148c4a2da6d1a5373d2135cd6` | The reopened PERF-140 terminal edit passed `make lint-markdown` at `.cartulary/test-results/20260817T111206Z-p1913679` and `git diff --check`; PERF-090 through PERF-150 are `DONE`, with no mandatory `BLOCKED` or `DROPPED` slice. The prior 776/776 check belongs to superseded source, and the 947/951 release remains a retained invalid attempt | Acceptance still requires exact 1,201-row closure, all browser/session evidence, zero leaked per-test resources, bounded two-slot fallback behavior, current-source cache and release inventory evidence, and a successful release no slower than 875.326s. No successful full check exists for this exact source, so retained-run maintenance remains unavailable | Repeat the five preflight gates, run `make agent-finalize` with `RESULTS_DIR` unset, then run a fresh `make check` followed immediately by warm `make release-check`; reopen the owning slice for any defect or unexplained timing miss |
| 2026-08-17T07:38:06-04:00 | PERF-160 `IN_PROGRESS` to `TODO`; PERF-140 `DONE` to `IN_PROGRESS` for finalizer admission and object-store reset readiness | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:a5f57074019f762435599b1da55dc6e8c29292a7c459d321539d7e4f552b5e13`; tracker-excluded dirty-state digest `d0c45f8747f780d148c4a2da6d1a5373d2135cd6` | All PERF-160 preflight gates passed at `.cartulary/test-results/20260817T111251Z-p1914830`, `20260817T111301Z-p1917749`, `20260817T111305Z-p1918204`, `20260817T111312Z-p1918744`, and `20260817T111326Z-p1919331`; `make agent-finalize` passed with `RESULTS_DIR` unset at `20260817T111345Z-p1920325`; and `make check` passed 776/776 in 384.832s at `20260817T111401Z-p1923205`. The immediate warm release failed 948/951 in 885.956s at `.cartulary/test-results/20260817T112032Z-p2046548`, 10.630s above the ceiling | One app.server exact-owner finalizer exhausted its actual forced-drop operation context after admission, proving two fallback slots alone do not bound concurrent ordinary catalog mutations. Separately, the stateful Evidence browser group received HTTP 503 while uploading immediately after its session reset; the reset currently deletes and recreates the stable session bucket, introducing an unnecessary availability boundary. The derived browser target failure is not independent. The failed release root is retained as the controlled before case and excluded from acceptance | Bound all cooperating exact-owner drops without globally serializing the hot path; preserve the stable browser bucket while emptying it between stateful generations and prove mutation readiness before admission; then rerun focused overlap, reset, browser, cleanup, and timing evidence before reclosing PERF-140 |
| 2026-08-17T07:53:16-04:00 | Reopened PERF-140 `IN_PROGRESS` to `DONE` after catalog admission and stable-namespace repair | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:bad1d93254c79e9a0695c0d7dca93d4eb963cfa76e50b99f08556207f3d536aa`; tracker-excluded dirty-state digest `4a3b11188c442c292d3db83bd5822daea3ad143a` | Every cooperating exact-owner database drop now acquires one service-local four-slot advisory lease before its normal attempt, retains it through any atomic forced fallback, and starts the unchanged five- and fifteen-second operation budgets only after admission; fresh bounded unlock and parent cancellation remain mandatory. Stateful browser renewal now empties its stable owned bucket in place and requires a put/head/delete readiness proof before admitting the next generation; only terminal finalization deletes the namespace. `make test-fast` passed 405/405 at `.cartulary/test-results/20260817T114231Z-p2272717`; the real object-store reset row passed 3/3 at `20260817T114403Z-p2294630`; and owner generation passed at `20260817T114908Z-p2339672`. The controlled 221-unit backend overlap passed at `20260817T114455Z-p2309578` in 229.953s: all 418 drops completed, none required forced fallback, aggregate observed cleanup time was 302.595s, maximum observed cleanup including lease wait was 7.260s, and no terminal recovery ledger remained. This is 4.1% faster than the prior 239.767s candidate and 42.1%/46.5% faster than the rejected 397.420s/429.719s serialized candidates. The full stateful browser chain passed 34/34 at `20260817T114922Z-p2342724` in 132.813s, including the previously failing Evidence upload after renewal. `make lint-go` passed; generation drift, artifact policy, JSON shapes, harness contract, and shell lint passed at `20260817T115256Z-p2394068`, `p2394110`, `p2394164`, `p2394456`, and `p2394558` | Exact ownership, eager cleanup, independent operation limits, atomic fallback, database and bucket identity, reset contamination checks, redaction, recovery fallback, schema identities, and public targets remain unchanged. The two-slot fallback-only semaphore is retired because ordinary mutations could still starve an admitted fallback; global serialization remains rejected for throughput, and timeout increases remain rejected. Deleting and recreating a stable stateful namespace is retired because identity replacement adds no isolation value and creates an availability boundary. The full release-scale composition remains a mandatory PERF-160 gate; any recurrence reopens this slice | Validate this terminal tracker edit with Markdown lint and diff checks, then reactivate PERF-160 and repeat the entire preflight, finalization, full-check, and immediate warm-release sequence on this source identity |
| 2026-08-17T07:54:07-04:00 | PERF-160 `TODO` to `IN_PROGRESS` after terminal resource-lifetime repair | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:bad1d93254c79e9a0695c0d7dca93d4eb963cfa76e50b99f08556207f3d536aa`; tracker-excluded dirty-state digest `4a3b11188c442c292d3db83bd5822daea3ad143a` | The reopened PERF-140 terminal checkpoint passed `make lint-markdown` at `.cartulary/test-results/20260817T115352Z-p2399449` and `git diff --check`; PERF-090 through PERF-150 are `DONE`, with no mandatory `BLOCKED` or `DROPPED` slice | All preceding broad release failures remain retained invalid attempts. Acceptance requires a complete new preflight, a current-source 776-unit check, exact 1,201-row closure, stable resource and evidence lifecycles under full overlap, and the 875.326s warm release ceiling | Run all five preflight gates, `make agent-finalize` with `RESULTS_DIR` unset, then one fresh `make check` followed immediately by its warm `make release-check`; reopen the exact owning slice for any failure or unexplained timing miss |
| 2026-08-17T08:22:39-04:00 | PERF-160 `IN_PROGRESS` to `TODO`; PERF-130 `DONE` to `IN_PROGRESS` for measurement observation closure | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:bad1d93254c79e9a0695c0d7dca93d4eb963cfa76e50b99f08556207f3d536aa`; tracker-excluded dirty-state digest `4a3b11188c442c292d3db83bd5822daea3ad143a` | PERF-160 preflight passed at `.cartulary/test-results/20260817T115438Z-p2400704`, `p2400745`, `p2400708`, `p2401115`, and `p2401155`; `make agent-finalize` passed with `RESULTS_DIR` unset at `20260817T115455Z-p2405803`; and `make check` passed 776/776 in 532.494s at `20260817T115511Z-p2408675`. The immediate release failed 945/951 in 1,021.020s at `.cartulary/test-results/20260817T120411Z-p2570178`, 145.694s above the ceiling | One of four serialized Timeline measurements completed 39 measured samples but its final blank-row commit never satisfied any accepted summary variant despite the accepted action mark and mounted grid remaining present; the other three Timeline measurements and all three Network Flow measurements passed. This is a PERF-130 measurement-observation completeness defect and is the active slice. Separately, four exact-owner database finalizers all exhausted actual forced-drop operation contexts at the same time; that PERF-140 defect is recorded but remains queued so only one slice is active. The derived measurement target summary is not independent. The invalid release had 456 cache hits, 35 misses, and 460 bypasses; its 186.572s dependency critical path ended in the webserver-backed functional-support lane, while PostgreSQL, object-store, and CPU saturation were 301.384s, 451.483s, and 186.390s | Prove whether the accepted mutation lacked a current summary signal or the observer rejected an equivalent committed row; repair the current semantic observation without retry or weakened criteria, run the exact four-row Timeline closure, then terminally validate PERF-130 before serially reopening PERF-140 for forced-drop quiescence |
| 2026-08-17T08:40:01-04:00 | Reopened PERF-130 `IN_PROGRESS` to `DONE` after query-scoped created-row presentation repair | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:fb5d35119f2e7d6b7631e50b2ffdb6a91695039957d92609de437618776da075`; tracker-excluded dirty-state digest `b2bdbb399ab96265f8f72f273bf24b7f934b22b7` | Core 03 REQ-03-224 and its traceability mapping now require the most recent locally created committed row to remain a same-query presentation pin when an otherwise authoritative bounded refresh omits it; the next create replaces the pin and an incident or normalized query-identity change clears it. The Timeline coordinator owns the single pinned `record_id`, and row reconciliation resolves that identity through the committed-version high-water source before adding it only when absent. The strict observer, 10-second bound, 25-session traffic, sample count, thresholds, and zero-retry policy are unchanged. The focused catalog row passed 5/5 at `.cartulary/test-results/20260817T123252Z-p2807441`; `make frontend-typecheck` passed at `20260817T123131Z-p2795326`; `make frontend-unit` passed 390/390 at `20260817T123252Z-p2807544`; Biome passed at `20260817T123253Z-p2807587`; and owner generation passed at `20260817T123231Z-p2803971` | The exact four-row Timeline measurement passed 19/19 in 258.008s at `.cartulary/test-results/20260817T123453Z-p2845515`. It contains one sealed construction of the exact 20,000-row fixture, six ordered contributions, unchanged semantic validation digest `sha256:b9549c30dab549504fd945fc211da7649223573e8ea67d57b7bb17a2ff06b5a7`, four isolated leases, four qualified observations with one warm-up and 100 measured samples each, zero Playwright retries, p95 values 77.1/31.4/32.3/28.8ms under unchanged 150/100/100/100ms limits, complete five-class cleanup for every clone, and a 187-file retained secret-scan pass. Earlier Biome and format attempts at `20260817T123131Z-p2795358` and `20260817T123200Z-p2796594` correctly rejected formatting and one missing hook dependency before the terminal passes. Increasing observation time, adding retries, suppressing traffic, unioning all local creates, and carrying a pin across a query change remain rejected | Validate this terminal tracker edit with `make lint-markdown` and `git diff --check`; only then reopen PERF-140 for the queued release-scale forced-drop quiescence defect |
| 2026-08-17T08:41:24-04:00 | PERF-140 `DONE` to `IN_PROGRESS` for forced-drop reader/writer coordination | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:fb5d35119f2e7d6b7631e50b2ffdb6a91695039957d92609de437618776da075`; tracker-excluded dirty-state digest `b2bdbb399ab96265f8f72f273bf24b7f934b22b7` | The terminal PERF-130 tracker edit passed `make lint-markdown` at `.cartulary/test-results/20260817T124043Z-p2893899` and `git diff --check`. Controlled before evidence remains the invalid release `.cartulary/test-results/20260817T120411Z-p2570178`, where four exact-owner finalizers entered forced fallback together and all four actual `DROP DATABASE ... WITH (FORCE)` operations exhausted their independent 15-second contexts | The four-slot all-drop semaphore admitted four mutually disruptive forced catalog operations concurrently. Ordinary drops should remain concurrent, but a forced fallback must begin only after cooperating ordinary drops drain and must exclude every other normal or forced drop until it completes. Extending timeouts, retrying cleanup, or globally serializing all ordinary drops remain rejected | Replace slot admission with service-local shared/exclusive advisory coordination; prove normal overlap, exclusive drain/exclusion, parent cancellation, fresh unlock, and normal-first independent budgets; then rerun the focused cleanup contracts and real backend overlap before terminally validating PERF-140 |
| 2026-08-17T08:57:17-04:00 | Reopened PERF-140 `IN_PROGRESS` to `DONE` after shared/exclusive catalog coordination | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:03b17d69e430b0888430ed13400c1a52b2f9261d7f8a09a955f20c5d722dc95e`; tracker-excluded dirty-state digest `3c44913fe1b14db42bc982dc541362869a16714b` | Replaced numbered drop slots with one service-local reader/writer advisory identity: normal exact-owner drops use shared session locks and remain concurrent; an active or timed-out normal attempt releases shared admission before waiting for the exclusive counterpart; `ForceDropDatabase` also uses exclusive admission. The forced operation starts only after all cooperating readers drain and excludes all other normal or forced drops. Normal and forced budgets still begin independently after their own admission; parent cancellation blocks fallback; shared/exclusive unlocks use fresh bounded contexts; and coordination failure is never masked. The focused cataloged service-backed row passed 3/3 at `.cartulary/test-results/20260817T125113Z-p2921828`, including four simultaneous real active-handle databases whose cleanups all required fallback and left every database unreachable. Its pure concurrency fixture proves ordinary readers overlap and the forced writer cannot enter before the last reader exits | The 222-unit backend overlap passed in 216.305s at `.cartulary/test-results/20260817T125207Z-p2936925`, 5.9% faster than the prior 229.953s four-slot candidate despite the added explicit row. It recorded 422 ordinary harness finalizers, seven forced fallbacks in the ordinary workload, 223.682s aggregate observed drop time, 6.019s maximum observed drop time, zero failures, zero private recovery ledgers, and a 1,331-file retained secret-scan pass. Current owner-derived catalog closure is 1,202 rows: PERF-130's redundant production-build row remains removed, while this release-exposed cleanup mode now has one explicit semantic row. `make generate` passed at `20260817T125059Z-p2918657`; drift, artifact policy, and JSON shapes passed at `20260817T125624Z-p2967243`, `p2967319`, and `p2967302`; Go lint passed; and the terminal harness contract passed at `20260817T125656Z-p2976164`. The first contract attempt at `20260817T125625Z-p2967630` correctly rejected the stale dedicated-row count before its exact current expectation changed from 251 to 252. Numbered slots, fallback-only semaphores, global ordinary-drop serialization, lock polling/upgrades, timeout increases, and cleanup retries remain rejected | Validate this terminal tracker edit with `make lint-markdown` and `git diff --check`; then activate PERF-160 and repeat every preflight, finalization, full-check, and immediate warm-release gate on the new source identity |
| 2026-08-17T08:59:17-04:00 | PERF-160 `TODO` to `IN_PROGRESS` after terminal reader/writer cleanup repair | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:03b17d69e430b0888430ed13400c1a52b2f9261d7f8a09a955f20c5d722dc95e`; tracker-excluded dirty-state digest `3c44913fe1b14db42bc982dc541362869a16714b` | The reopened PERF-140 terminal checkpoint passed `make lint-markdown` at `.cartulary/test-results/20260817T125857Z-p2977271` and `git diff --check`; PERF-090 through PERF-150 are `DONE`, with no mandatory `BLOCKED` or `DROPPED` slice. Current exact owner-derived closure is 1,202 rows: the redundant production-fixture row remains removed and the release-exposed cleanup coordination mode has one direct semantic row | All preceding broad release failures remain retained invalid attempts. Acceptance requires a complete new preflight, `make agent-finalize` with `RESULTS_DIR` unset because no successful full check exists for this exact source, one fresh `make check`, and its immediate warm `make release-check` no slower than 875.326s, with complete secure lifecycle, cache, fixture, cleanup, evidence, and no-Markdown-runtime-dependency proofs | Run all final preflight gates and finalization, then the required full pair; reopen the exact owning slice for any defect or unexplained timing miss |
| 2026-08-17T09:29:03-04:00 | PERF-160 `IN_PROGRESS` to `TODO`; PERF-140 `DONE` to `IN_PROGRESS` for complete catalog-mutation coordination | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:03b17d69e430b0888430ed13400c1a52b2f9261d7f8a09a955f20c5d722dc95e`; tracker-excluded dirty-state digest `3c44913fe1b14db42bc982dc541362869a16714b` | PERF-160 preflight passed at `.cartulary/test-results/20260817T130005Z-p2978661`, `p2978662`, `p2978685`, `p2979090`, and `p2979058`; `make agent-finalize` passed with `RESULTS_DIR` unset at `20260817T130022Z-p2983682`; and `make check` passed 777/777 in 479.318s at `20260817T130057Z-p2986638`. The immediate warm release failed 940/952 in 1,006.291s at `.cartulary/test-results/20260817T130903Z-p3147999`, 130.965s above the ceiling. Four Go rows failed only during eager database retirement: three actual forced drops exhausted their operation contexts and one waiter exhausted its parent cleanup context before exclusive admission | The shared/exclusive protocol coordinates repository-managed drops but not repository-managed database creation and clone mutations. At release-scale overlap, uncoordinated catalog writers can continue while the forced writer holds its drop-only lock, so exclusive admission does not establish actual catalog quiescence. The same failed release also contains four direct browser row failures and one derived accessibility timeout; those are recorded but remain queued until this serial PERF-140 repair is terminal. The release root is excluded from acceptance | Move the service-local reader/writer identity into a cohesive catalog coordinator, route every harness-owned create/clone and normal drop through shared admission, keep forced drop exclusive, prove cross-operation exclusion with real service evidence, then reclose PERF-140 before classifying the queued browser failures |
| 2026-08-17T09:45:28-04:00 | Reopened PERF-140 `IN_PROGRESS` to `DONE` after complete catalog-mutation coordination | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:146504fe8258697be2fa868612de88fcfdc4a36e75acc2902dc9deeac15cbcf0`; tracker-excluded dirty-state digest `92eb49a73582b5b29495e8465268293ed8b8b528` | The adopted owner now places every cooperating harness-owned database create, clone, template-state mutation, and ordinary exact-owner drop under the shared side of one service-local catalog coordinator; forced deletion is the exclusive writer. Admission waits, normal operations, and forced operations are independently bounded, and the removed 20-second aggregate finalizer deadline can no longer truncate an admitted forced operation. Pgtest, performance-fixture lifecycle, testservices/browser construction, Recovery clone construction, and template corruption fixtures all use the common boundary. The focused cataloged row passed 3/3 at `.cartulary/test-results/20260817T133629Z-p3389296`, including a real PostgreSQL proof that database creation cannot enter during exclusive admission. `make format` and owner generation passed at `20260817T133604Z-p3382771` and `20260817T133616Z-p3386366` | The 222-unit backend overlap passed in 243.488s at `.cartulary/test-results/20260817T133744Z-p3404784`: all 422 recorded drops completed, including 11 bounded forced fallbacks; maximum observed cleanup was 9.444s; aggregate observed cleanup was 308.587s; failures and recovery ledgers were zero. This run is 12.6% slower than the immediately preceding 216.305s candidate, while the original matched Incidents before/after gate remains the governing PERF-140 ceiling and passed at +2.0%. The additional admission round trips and writer quiescence are retained because they close the release-scale correctness gap; capacity increases, timeout extensions, retries, and uncoordinated construction remain rejected. `make test-fast` passed 405/405 at `20260817T134237Z-p3437836`; Go lint passed; drift, artifact policy, JSON shapes, and harness contract passed at `20260817T134451Z-p3487948`, `p3487934`, `p3487981`, and `p3488228`; `git diff --check` passed and `docs/domain.md` remains unchanged | Validate this terminal tracker edit with Markdown lint and diff checks; then classify the queued browser failures from the invalid release one owning slice at a time before resuming PERF-160 |
| 2026-08-17T09:48:56-04:00 | PERF-160 `TODO` to `IN_PROGRESS` for browser convergence after classification | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:146504fe8258697be2fa868612de88fcfdc4a36e75acc2902dc9deeac15cbcf0`; tracker-excluded dirty-state digest `92eb49a73582b5b29495e8465268293ed8b8b528` | PERF-140's terminal edit passed `make lint-markdown` at `.cartulary/test-results/20260817T134650Z-p3493800` and `git diff --check`. The invalid release's three stateful failures all waited for workbook-only `pending-queue-count` after the client had correctly transitioned to the sign-in shell; its fourth replay failure applied new session cookies through an API context but never notified the live SPA to leave that shell. The accessibility timeout likewise clicked a current-user control while the same automatic auth transition detached it | Core 03 requires the client-memory queue to survive same-runtime revocation and replay FIFO after re-authentication; it does not require workbook queue controls to remain mounted in the auth shell. Adding those controls would couple authentication presentation to workbook state. Full navigation/reload is also invalid because the base profile does not promise queue survival across reload | Make tracked re-authentication use the visible in-page login surface when present, retain API-context login for initial setup, remove only auth-shell-inapplicable queue-markup assertions while preserving exact post-login replay assertions, and make the accessibility case await automatic revocation routing before re-authentication; then run the exact failed rows and repeat PERF-160 from preflight |
| 2026-08-17T09:55:15-04:00 | PERF-160 remains `IN_PROGRESS`; browser convergence diagnosis refined after the first exact rerun | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:0ac65b52e2c1d5cc0cd7d8003e409544786a3d885f112b619b506391f1e178e3`; tracker-excluded dirty-state digest `5feac74244a5b17087747eff2a7e6d15cf85e7a7` | Form-based re-authentication, automatic revoked-shell waiting, and removal of only auth-shell-inapplicable markup checks passed formatting, typecheck, and Biome at `.cartulary/test-results/20260817T135005Z-p3496051`, `20260817T135013Z-p3499761`, and `p3499779`. The exact three-row Collaboration rerun failed 11/16 at `.cartulary/test-results/20260817T135030Z-p3500645`: both stateful rows reached successful in-page authentication but observed zero replayed patches, while the HTTP-auth case returned before the session cookie was committed | The exact replay assertions correctly exposed a product lifetime defect: `WorkbookMutationRuntime` is constructed inside the authenticated Workbook shell, and collaboration-coordinator disposal invalidates it when revocation routes the same JavaScript runtime to the Auth shell. This contradicts Core 03 REQ-03-099 and REQ-03-100. Removing replay assertions, adding reload, persisting across browser runtimes, or mounting Workbook controls in Auth remain rejected | Make mutation-runtime ownership browser-runtime-scoped and bounded to one `(incident_id, client_instance_id)` authority across authenticated-shell unmount/remount; treat coordinators as borrowers; invalidate on a different incident or app disposal; await the terminal login response and cookie before returning; then prove unit lifetime boundaries and rerun the same exact rows |
| 2026-08-17T10:08:41-04:00 | PERF-160 browser convergence repair complete; slice remains `IN_PROGRESS` for broad acceptance | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:1996821d1ab0ba4ce6bea35857ffa0a029309d48161825151d7edfc7fba8552b`; tracker-excluded dirty-state digest `96511247b66afe50686b233e516e8ea099424f7e` | Added one App-owned, browser-runtime-lifetime mutation registry bounded to the active `(incident_id, client_instance_id)`; authenticated Workbook and Collaboration shells now borrow it, while a different incident and App disposal retire it. In-page login now requires a successful terminal response, a newly issued session plus CSRF cookie, request-context synchronization, and Auth-shell exit. Unit coverage proves same-scope retention, incident replacement, App disposal, wrong-scope rejection, and borrowed coordinator lifetime. Frontend units passed 390/390 at `.cartulary/test-results/20260817T140004Z-p3589548`; the import boundary passed at `20260817T140005Z-p3589749`; format and typecheck passed at `20260817T140308Z-p3635285` and `20260817T140309Z-p3635465` | The first post-lifetime combined rerun at `.cartulary/test-results/20260817T140021Z-p3590525` proved both stateful queues replayed but exposed login completion races; the next run at `20260817T140324Z-p3639341` passed both stateful rows and isolated the HTTP test beginning login before the required revoked prompt. After making that transition explicit, the exact HTTP-auth row passed 11/11 at `20260817T140606Z-p3684664`; the adjacent Timeline remount row passed 12/12 at `20260817T140715Z-p3723004`; and the session/MFA accessibility row passed 11/11 at `20260817T140715Z-p3723009`. Exact FIFO, server state, no-reload, scope isolation, session tracking, and accessibility assertions remain intact | Validate the convergence tracker and guide with Markdown lint and diff checks, refresh the remaining final gates, run `make agent-finalize` without retained evidence for this source, then execute a fresh `make check` followed immediately by the required warm `make release-check` |
| 2026-08-17T10:37:54-04:00 | PERF-160 `IN_PROGRESS` to `TODO`; PERF-140 `DONE` to `IN_PROGRESS` for release-scale cleanup contention | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:13fa67956a8ffe4dcc6750f9092d383281565d0ba1c7740e13ce2faba3f2afcf`; tracker-excluded dirty-state digest `a6a8b2441b6335b6b671dec2b59c9b6869e4937c` | Final preflight and `make agent-finalize` passed, and `make check` passed 777/777 in 458.195s at `.cartulary/test-results/20260817T141128Z-p3813756`. The immediately warm `make release-check` failed 948/952 in 965.175s at `.cartulary/test-results/20260817T141912Z-p3968214`: two Extension State and two Imports rows timed out acquiring shared/exclusive catalog coordination or completing a forced drop. The run recorded 373 backend-integration drops consuming 274.774s, with forced drops as slow as 26.816s, while eight scheduler PostgreSQL lanes admitted package processes that internally fan out many parallel database lifecycles | The current single global reader/writer advisory lock prevents conflicting forced catalog operations but creates a convoy: a queued or executing target-specific forced drop blocks unrelated creates and ordinary drops, and one package-level scheduler token materially undercounts its in-process database fan-out. Increasing timeouts, retrying DDL, weakening cleanup, accepting four failed rows, or waiving the 875.326s ceiling are rejected. The failure is owned by PERF-140 and the release root is retained as invalid evidence | Adopt a bounded PostgreSQL admission model that reflects service-backed package fan-out, preserve explicit override validation and browser claims, prove ordinary and forced cleanup under overlapping owners without catalog-lock timeouts, reclose PERF-140, and only then reactivate PERF-160 and repeat the complete acceptance sequence |
| 2026-08-17T10:57:50-04:00 | Reopened PERF-140 `IN_PROGRESS` to `DONE` after package-fan-out admission repair | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:cfafa2b226dcf6c0ef263afcd07a98cd1b81d8576b6fd6da849cb7357508c627`; tracker-excluded dirty-state digest `06e34201f9416ff6037dce7dd79aa089fc4a9499` | Adopted one-to-one PostgreSQL Go admission: every PostgreSQL-backed sharded or raw Go process now bounds `GOMAXPROCS`, its scheduler CPU claim, and its PostgreSQL claim to the minimum of planned CPU parallelism, captured PostgreSQL capacity, and the topology-declared PostgreSQL claim. Ordinary service rows therefore keep one lane and execute one Go test at a time; up to eight independent package processes still overlap by default, while existing browser claims remain two or four. The 222-unit backend overlap passed in 247.154s at `.cartulary/test-results/20260817T145238Z-p32933`, only 1.5% slower than the preceding 243.488s comparison. It used all eight PostgreSQL lanes, completed 436 database lifecycles with zero failures or recovery ledger, and retained complete bounded diagnostics | A four-lane-per-process candidate was rejected and interrupted after only 210/222 units in 414.536s at `.cartulary/test-results/20260817T144435Z-p1796`; increasing claims serialized independent single-row packages without reducing aggregate service parallelism further. The accepted model removes the hidden package-level multiplier instead of reducing service capacity, increasing timeouts, polling, retrying, or weakening the shared/exclusive cleanup boundary. `make generate` passed at `.cartulary/test-results/20260817T145716Z-p58568`; Biome and semantic harness contracts passed at `20260817T145217Z-p31783` and `p31802`; final drift, artifact-policy, JSON-shape, and harness-contract gates passed at `20260817T145731Z-p61623`, `p61639`, `p61659`, and `20260817T145732Z-p61950`; `git diff --check` passed and `docs/domain.md` remains unchanged | Validate this terminal tracker edit with `make lint-markdown` and `git diff --check`; then activate PERF-160 with a new source and tracker-excluded dirty-state checkpoint and repeat the full acceptance sequence |
| 2026-08-17T10:59:05-04:00 | PERF-160 `TODO` to `IN_PROGRESS` after terminal release-scale admission repair | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:cfafa2b226dcf6c0ef263afcd07a98cd1b81d8576b6fd6da849cb7357508c627`; tracker-excluded dirty-state digest `06e34201f9416ff6037dce7dd79aa089fc4a9499` | The terminal PERF-140 tracker edit passed `make lint-markdown` at `.cartulary/test-results/20260817T145825Z-p66212` and `git diff --check`. PERF-090 through PERF-150 are `DONE`, with no mandatory `BLOCKED` or `DROPPED` slice. The immediately preceding successful full check belongs to the superseded source, so retained-run maintenance cannot use it for this source | Broad acceptance must prove exact 1,202-row closure, one-to-one PostgreSQL Go admission under browser overlap, bounded/redacted evidence, eager cleanup, one production fixture build, deterministic paired release inventory, complete browser/auth lifecycles, no executable Markdown dependency, and the 875.326s warm release ceiling. Earlier failed and interrupted roots remain diagnostics only | Run the required drift, artifact-policy, JSON-shape, harness-contract, Markdown, and diff gates; run `make agent-finalize` with `RESULTS_DIR` unset; then run one fresh `make check` followed immediately by one warm `make release-check` |
| 2026-08-17T11:24:59-04:00 | PERF-160 `IN_PROGRESS` to `TODO`; PERF-140 `DONE` to `IN_PROGRESS` for the remaining global-lock convoy | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:cfafa2b226dcf6c0ef263afcd07a98cd1b81d8576b6fd6da849cb7357508c627`; tracker-excluded dirty-state digest `06e34201f9416ff6037dce7dd79aa089fc4a9499` | PERF-160 preflight passed at `.cartulary/test-results/20260817T145915Z-p67464`, `p67488`, `p67560`, `20260817T145916Z-p67882`, and `p67895`; `make agent-finalize` passed with `RESULTS_DIR` unset at `20260817T145931Z-p72553`. The fresh check passed 777/777 in 506.397s at `.cartulary/test-results/20260817T145951Z-p75430`. Its immediate warm release failed 951/952 in 964.842s at `.cartulary/test-results/20260817T150822Z-p230714`, 89.516s above the ceiling. The sole row failure, `module.extensions.integration.state_concurrent_admission`, spent its complete 15-second preparation budget waiting for the global shared catalog advisory lock | One-to-one package admission removed hidden Go fan-out and passed the focused overlap, but cannot prevent a target-specific forced drop from holding or queuing the global exclusive identity while unrelated creates wait. The release critical path moved to the canonical Timeline measurement closure; PostgreSQL remained saturated for 396.479s and object storage for 520.923s. Increasing admission budgets, retrying the failed row, accepting the 951/952 result, or closing on an unexplained timing miss remain rejected | Replace the service-global reader/writer lock with target-scoped ownership exclusion plus a bounded fair catalog-mutation admission mechanism that cannot let one slow target stop unrelated targets; preserve normal-first/atomic-force semantics and exact owner proof; then remeasure overlap and broad critical-path timing before reclosing PERF-140 |
| 2026-08-17T11:43:29-04:00 | Reopened PERF-140 `IN_PROGRESS` to `DONE` after target-striped catalog and high-water admission repair | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:88b2cf87cccdf749c1f34d63ad697a001afe3d28659fa522f51b7b4fb0164e2c`; tracker-excluded dirty-state digest `3b6a64117359b6dacf117d61321564c1517461a9` | Replaced the service-global reader/writer identity with an exact target lock followed by one of four deterministic service-local catalog stripes for every create, clone, template mutation, normal drop, and forced drop. Normal and forced attempts reacquire independently, unlock in reverse order through fresh bounded contexts, and retain exact ownership plus atomic force semantics. The renamed semantic row `harness.browser.integration.postgres_cleanup_target_stripe_coordination` proves same-target exclusion, overlap on a different stripe, and four simultaneous active-handle forced cleanups; it passed at `.cartulary/test-results/20260817T153018Z-p453130` and on the combined source at `20260817T153752Z-p498815`. The former reader/writer row has an owner-backed hard-cut migration record | Added a ranked high-water drain barrier: when the highest-priority ready unit is temporarily short of a resource, lower-ranked work cannot refill only that resource; unrelated-resource backfill continues and no unstarted unit owns capacity. This prevents one-lane Go packages from indefinitely fragmenting four-lane browser/measurement admission. The semantic scheduler fixture and `make harness-contract` passed at `.cartulary/test-results/20260817T153752Z-p498919`. The combined 222-unit backend overlap passed 222/222 in 241.752s at `20260817T153847Z-p513860`, faster than both the 243.934s stripe-only and 247.154s global-lock comparisons, with zero cleanup failure or recovery ledger. Polling, random/caller-selected stripes, global quiescence, timeout increases, retry, preemption, and global backfill suspension remain rejected | `make format` and `make generate` passed at `.cartulary/test-results/20260817T153744Z-p495182` and `20260817T154255Z-p539241`; final drift, artifact-policy, JSON-shape, and harness-contract gates passed at `20260817T154311Z-p542290`, `p542302`, `p542337`, and `p542602`; `git diff --check` passed and `docs/domain.md` remains unchanged. Validate this terminal tracker edit, then reactivate PERF-160 and repeat the full acceptance sequence |
| 2026-08-17T11:44:30-04:00 | PERF-160 `TODO` to `IN_PROGRESS` after terminal target-striped cleanup repair | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:88b2cf87cccdf749c1f34d63ad697a001afe3d28659fa522f51b7b4fb0164e2c`; tracker-excluded dirty-state digest `3b6a64117359b6dacf117d61321564c1517461a9` | The terminal PERF-140 tracker edit passed `make lint-markdown` at `.cartulary/test-results/20260817T154404Z-p546886` and `git diff --check`. PERF-090 through PERF-150 are `DONE`, with no mandatory `BLOCKED` or `DROPPED` slice. No successful full check exists for this current source, so `make agent-finalize` must again omit `RESULTS_DIR` | Acceptance requires exact 1,202-row closure, no catalog timeout, measurement admission without resource fragmentation, full browser/auth lifecycle, bounded/redacted evidence, secure deterministic release-inventory caching, one production fixture build, eager cleanup, no executable Markdown dependency, and a warm release at or below 875.326s | Repeat every preflight gate and `make agent-finalize`; then run one current-source `make check` followed immediately by its warm `make release-check`, reopening the exact owning slice for any remaining failure |
| 2026-08-17T12:13:21-04:00 | PERF-160 `IN_PROGRESS` to `TODO`; PERF-140 `DONE` to `IN_PROGRESS` for target-stripe saturation | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:88b2cf87cccdf749c1f34d63ad697a001afe3d28659fa522f51b7b4fb0164e2c`; tracker-excluded dirty-state digest `3b6a64117359b6dacf117d61321564c1517461a9` | The current-source preflight and `make agent-finalize` passed; `make check` then passed 777/777 in 540.336s at `.cartulary/test-results/20260817T154539Z-p556156`. Its immediate warm `make release-check` failed 949/952 in 1,005.563s at `.cartulary/test-results/20260817T155445Z-p713947`, 130.237s above the ceiling. Imports timed out after 15.020s acquiring a catalog stripe, while two Extension State rows exhausted forced-drop contexts after 20.25s and 20.38s. The bounded summary records 374 backend-integration drops consuming 492.084s and several individual drops above ten seconds | Four deterministic stripes eliminated global unrelated-target exclusion but underrepresent the scheduler's eight admitted PostgreSQL lanes. Release-scale hash collisions still form queues long enough to exhaust operation contexts. The high-water scheduler barrier also failed its broad timing objective: although the Timeline dependency path shortened, total release time regressed, so it remains a candidate to remove rather than an accepted compatibility burden. Retrying rows, increasing operation timeouts, weakening eager cleanup, accepting three failures, or waiving the timing ceiling remain rejected | Align bounded catalog admission with the typed eight-lane PostgreSQL capacity, remove any scheduler reservation that does not improve total convergence, rerun the exact cleanup coordination row and the 222-unit backend overlap, then reclose PERF-140 before repeating PERF-160 |
| 2026-08-17T12:25:13-04:00 | Reopened PERF-140 `IN_PROGRESS` to `DONE` after removing redundant catalog admission | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:2290497da69cd79747e5be897fa42dae28dec042fd05495ae52876ed6491fc62`; tracker-excluded dirty-state digest `4a47bdaf33ca24e291f2849f4ad151a09df18e79` | Removed the hash-stripe layer from the cohesive PostgreSQL catalog coordinator. Every cooperating mutation retains an exact-target session lock, normal and forced attempts remain separately bounded, and exact ownership still gates atomic force; aggregate catalog concurrency is now owned solely by graph PostgreSQL claims. Renamed the semantic row to `harness.browser.integration.postgres_cleanup_target_scoped_coordination`, recorded the reader/writer hard-cut replacement, and added `postgres_catalog_isolated` so its intentional four-target cleanup overlap honestly reserves four CPU and PostgreSQL lanes. The focused row passed 3/3 in 39.741s at `.cartulary/test-results/20260817T161940Z-p951832`, with peak claims exactly matching the fixture | Removed the high-water drain barrier because its broad run increased total release time and restored ordinary ranked backfill plus monotonic aging. The release-shaped `make backend-integration` overlap passed 222/222 in 217.329s at `.cartulary/test-results/20260817T162047Z-p967178`, 10.1% faster than the 241.752s target-striped comparison. It saturated all eight PostgreSQL lanes for 109.964s, completed 377 eager per-test drops, retained no cleanup failure or recovery ledger, and reduced the worst retained cleanup diagnostic to 6.899s. Doubling the stripe count, random/caller-selected stripes, timeout increases, retries, weakened cleanup, global quiescence, and scheduler reservation remain rejected as unnecessary or collision-prone | `make format` passed at `.cartulary/test-results/20260817T161921Z-p945304`; `make generate` passed at `20260817T161927Z-p948853`; final generation drift, artifact policy, JSON shapes, and harness contract passed at `20260817T162443Z-p993203`, `p993240`, `p993254`, and `p993504`; `git diff --check` passed and `docs/domain.md` remains unchanged. Validate this terminal tracker edit with Markdown lint and diff checks, then reactivate PERF-160 on this exact source and repeat the complete acceptance sequence |
| 2026-08-17T12:26:00-04:00 | PERF-160 `TODO` to `IN_PROGRESS` after terminal target-scoped cleanup repair | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:2290497da69cd79747e5be897fa42dae28dec042fd05495ae52876ed6491fc62`; tracker-excluded dirty-state digest `4a47bdaf33ca24e291f2849f4ad151a09df18e79` | The terminal PERF-140 tracker edit passed `make lint-markdown` at `.cartulary/test-results/20260817T162545Z-p997854` and `git diff --check`. PERF-090 through PERF-150 are again `DONE`, with no mandatory `BLOCKED` or `DROPPED` slice. The prior 777/777 check belongs to the superseded striped source, so retained-run maintenance cannot use it for this source | Acceptance requires exact 1,202-row owner closure; complete event and failure lifecycles; target-scoped eager database cleanup; bounded/redacted evidence; secure deterministic release-inventory restore; one production fixture construction; unique resource identities; no executable Markdown dependency; and a successful warm release at or below 875.326s | Repeat every final preflight gate; run `make agent-finalize` with `RESULTS_DIR` unset; then run one current-source `make check` followed immediately by its warm `make release-check`, reopening the owning slice for any related defect |
| 2026-08-17T12:54:57-04:00 | PERF-160 `IN_PROGRESS` to `TODO`; PERF-110 `DONE` to `IN_PROGRESS` for object-store admission closure | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:2290497da69cd79747e5be897fa42dae28dec042fd05495ae52876ed6491fc62`; tracker-excluded dirty-state digest `4a47bdaf33ca24e291f2849f4ad151a09df18e79` | Final preflight and `make agent-finalize` passed; `make check` passed 777/777 in 510.492s at `.cartulary/test-results/20260817T162717Z-p1007139`; and the immediately warm `make release-check` passed all 952/952 units in 974.365s at `.cartulary/test-results/20260817T163656Z-p1165012`. Correctness, cleanup, and lifecycle closure are accepted, but the release missed the 875.326s ceiling by 99.039s and cannot close PERF-160 | The production fixture constructed once and its measurement target finished by 336.304s; release inventory recorded 456 cache hits and remained off the terminal path. The dominant constraint was the four-lane object-store claim: it was saturated for 538.791s although the bounded service scope recorded only about 1.3s of object-store create/clean activity. Webserver-backed browser work could not begin until 627.740s, and the dependent stateful, accessibility, support, and visual tail ended at 969.510s. PostgreSQL saturation was lower at 373.554s and exact cleanup was complete. This is a service-specific fixture-admission defect owned by PERF-110, not authority to weaken browser stages, evidence, isolation, or duration assertions | Run one controlled warm release with an eight-lane object-store capacity override and otherwise identical source, PostgreSQL capacity, service mode, and host. Adopt the policy only if the complete release passes inside the ceiling with exact closure and no object-store failure; otherwise reject it and continue critical-path diagnosis |
| 2026-08-17T16:15:31-04:00 | Reopened PERF-110 `IN_PROGRESS` to `DONE` after service-Go process-boundary repair | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:b2bd8547edcc26f323e74385525f8ab0f73f0c739d55fcb78158d54d70efa8a9`; tracker-excluded dirty-state digest `95a04087a8ba123d6d11258c631a13c6602d1473` | Controlled release experiments rejected every capacity or packing workaround: object-store capacity eight failed 950/952 in 960.079s at `.cartulary/test-results/20260817T165544Z-p1380694`; bounded scheduler affinity passed but regressed to 1,081.829s cold and 995.680s warm at `20260817T171830Z-p1603181` and `20260817T173657Z-p1861534`; browser PostgreSQL claim two passed its 54/54 aggregate at `20260817T175852Z-p2080721` but both release attempts failed at `20260817T180438Z-p2151726` and `20260817T182152Z-p2409264`; object-store capacity six failed 949/952 in 1,004.192s at `20260817T184139Z-p2626846`; and balanced service packing failed 945/952 in 962.739s at `20260817T193821Z-p3362919` after the schema-drift-invalid diagnostic attempt at `20260817T192028Z-p3099470`. Every candidate was fully reverted. The restored fully warm source passed 952/952 in 957.570s at `20260817T190026Z-p2884888`, confirming repeated Go process startup and package initialization, rather than a safe capacity increase, as the remaining structural cost | Adopted the already-owned deterministic service-Go batching path for compatible `postgres_dedicated`, `postgres_migration`, and per-test object-store rows. Their exact resources remain distinct per `testing.T`, eagerly finalized, and one-to-one admitted; only process startup is shared. `managed_process` remains inherently isolated. Added the typed optional `process_isolation="exclusive"` catalog input for assertions with proven process-global state and applied it to the Jobs OpenTelemetry capture row. A first 89-unit backend run failed only that row at `.cartulary/test-results/20260817T200110Z-p3586309`; the row passed alone at `20260817T200506Z-p3606255` and failed deterministically after any prior Jobs manager at `20260817T200549Z-p3620819` and `20260817T200749Z-p3636267`, proving the boundary. After projection, the exact pair passed at `20260817T201100Z-p3661560` and the complete Jobs owner passed 5/5 at `20260817T201145Z-p3676201` | `make format`, `make generate`, `make harness-contract`, and `make generate-drift` passed at `.cartulary/test-results/20260817T201018Z-p3651661`, `20260817T201021Z-p3655159`, `20260817T201030Z-p3658097`, and `20260817T201043Z-p3658596`. The final `make backend-integration` passed all 231 rows in 90/90 units and 159.161s at `.cartulary/test-results/20260817T201226Z-p3690846`, 26.8% faster than the matched 217.329s pre-change run while completing 377 dedicated and 33 migration database lifecycles with no terminal recovery ledger. Public targets, row meaning, failure attribution, fixture identity, object-store package reuse, and cleanup authority are unchanged. Validate this terminal tracker edit with `make lint-markdown` and `git diff --check`, then activate PERF-160 on this exact source and repeat the complete acceptance sequence |
| 2026-08-17T16:17:32-04:00 | PERF-160 `TODO` to `IN_PROGRESS` after terminal service-Go batching repair | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:b2bd8547edcc26f323e74385525f8ab0f73f0c739d55fcb78158d54d70efa8a9`; tracker-excluded dirty-state digest `95a04087a8ba123d6d11258c631a13c6602d1473` | The reopened PERF-110 terminal edit passed `make lint-markdown` at `.cartulary/test-results/20260817T201652Z-p3710790` and `git diff --check`. PERF-090 through PERF-150 are `DONE`, with no mandatory `BLOCKED` or `DROPPED` slice. The earlier 777/777 check belongs to a superseded source, so retained-run maintenance cannot consume it | Acceptance requires exact 1,202-row closure; explicit process-global isolation and otherwise deterministic service-Go batching; complete event/failure and browser/auth lifecycles; eager exact resource cleanup; bounded redacted evidence; one production fixture construction; one deterministic paired release-inventory generation or secure cache restore; no executable Markdown dependency; and a successful warm release no slower than 875.326s | Run the required preflight gates, run `make agent-finalize` with `RESULTS_DIR` unset, then run one fresh `make check` followed immediately by its warm `make release-check`; reopen the owning slice for any related defect or unexplained timing miss |
| 2026-08-17T16:38:15-04:00 | PERF-160 `IN_PROGRESS` to `TODO`; PERF-130 `DONE` to `IN_PROGRESS` for release-scale fixture construction | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:b2bd8547edcc26f323e74385525f8ab0f73f0c739d55fcb78158d54d70efa8a9`; tracker-excluded dirty-state digest `95a04087a8ba123d6d11258c631a13c6602d1473` | Final preflight and `make agent-finalize` passed; the fresh `make check` passed 622/622 in 348.717s at `.cartulary/test-results/20260817T201921Z-p3720502`. Its immediate warm `make release-check` failed 768/778 in 696.000s at `.cartulary/test-results/20260817T202517Z-p3868341`. The sole primary failure was the canonical fixture snapshot unit after one 225.049s construction; its failed bounded artifact reports `contribution_invalid`, and the four Timeline measurement groups plus their summaries were correctly dependency-blocked. All other 768 units passed, cache restore was healthy, and the invalid run remains below the timing ceiling but is excluded from acceptance | This is a PERF-130 construction-composition defect, not permission to retry, accept skipped measurement, build the production dataset twice, or weaken semantic validation. The retained diagnostics prove exactly one construction but do not identify the failing contribution, so failure observability must also be made bounded and actionable before another broad attempt | Reproduce through the focused measurement target, identify the exact contribution or semantic boundary, preserve one full build and every 20,000-row expectation, add bounded safe failure-stage diagnostics, reclose PERF-130 with its exact lifecycle and measurement gates, then repeat the complete PERF-160 sequence on the repaired source |
| 2026-08-17T16:58:34-04:00 | Reopened PERF-130 `IN_PROGRESS` to `DONE` after owned-connection finalization repair | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:b879471834b245a2ca5fb32c43241fd42b06b488c782b54456ed545c1d963f5a`; tracker-excluded dirty-state digest `3f34d5f7d5254ca29f843f0b6ee7515e2ee7c3a2` | The failed release unit log identified the exact boundary: all construction completed, but sealing observed one PostgreSQL connection after client pool closure. Cut bounded build diagnostics from v1 to v2, retaining only failure stage and ordered safe contribution identities, states, durations, and batch shape; raw errors and physical identities remain excluded. The builder now records every server-assigned backend PID opened by its pool, closes the pool, permits a five-second drain only for those exact PIDs, rejects any unowned connection immediately, never terminates it, and cleans an unsealed template on failure. Deterministic drain, unowned-connection, deadline, assembler-failure, and redaction tests passed in the two-row unit slice at `.cartulary/test-results/20260817T205204Z-p4132753`; the small four-clone lifecycle passed 3/3 at `20260817T205211Z-p4133192` | The exact four-row production measurement closure passed 22/22 in 259.762s at `.cartulary/test-results/20260817T205257Z-p4148268`. It emitted one sealed v2 diagnostic and one semantic build artifact; construction took 69.851s, recorded all six contributions completed, preserved the 20,000-row one-batch Timeline construction and ten-millisecond semantic validation, and admitted all four quiet isolated measurement groups only after the shared builder completed. Increasing the builder's resource claim, retrying the build, force-terminating an unknown connection, weakening the zero-connection seal, or constructing a second snapshot are rejected because retained evidence proved a finalization observation race, not insufficient construction capacity | `make format` passed at `.cartulary/test-results/20260817T205155Z-p4129164`; `make generate` and JSON shapes passed at `20260817T204943Z-p4124619` and `20260817T204957Z-p4127599`; final generation drift, artifact policy, and harness contract passed at `20260817T205731Z-p7434`, `20260817T205742Z-p10379`, and `20260817T205748Z-p10927`. Broad release-scale composition remains intentionally owned by PERF-160. Validate this terminal tracker edit with Markdown lint and diff checks, then reactivate PERF-160 on the exact repaired source |
| 2026-08-17T16:59:25-04:00 | PERF-160 `TODO` to `IN_PROGRESS` after terminal fixture finalization repair | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:b879471834b245a2ca5fb32c43241fd42b06b488c782b54456ed545c1d963f5a`; tracker-excluded dirty-state digest `3f34d5f7d5254ca29f843f0b6ee7515e2ee7c3a2` | The terminal PERF-130 tracker edit passed `make lint-markdown` at `.cartulary/test-results/20260817T205906Z-p12139` and `git diff --check`. PERF-090 through PERF-150 are `DONE`, with no mandatory `BLOCKED` or `DROPPED` slice. The earlier 622/622 check predates the v2 diagnostic and owned-connection barrier, so retained-run maintenance cannot consume it | Acceptance requires exact current owner-derived row closure; complete service, browser, failure, and cleanup lifecycles; secure deterministic release-inventory restore; one sealed production fixture construction; bounded and redacted evidence; unique resource identities; no executable Markdown dependency; and a successful warm release at or below 875.326s | Run every final preflight gate and `make agent-finalize` with `RESULTS_DIR` unset, then run one fresh `make check` followed immediately by its warm `make release-check`; reopen the owning slice for any related defect or unexplained timing miss |
| 2026-08-17T17:23:53-04:00 | PERF-160 `IN_PROGRESS` to `DONE`; overall effort `DONE` | `HEAD=5413bb0a3d09e22bbd5b9e9b4b0e28e217533cc8`; source digest `sha256:b879471834b245a2ca5fb32c43241fd42b06b488c782b54456ed545c1d963f5a`; graph digest `sha256:98794f8b421314e6a4cb1bf9736d606f5571f216510dac4d8ea18ea7af37b8a1`; tracker-excluded dirty-state digest `3f34d5f7d5254ca29f843f0b6ee7515e2ee7c3a2` | Final generation drift, artifact policy, JSON shapes, harness contract, Markdown lint, and diff checks passed at `.cartulary/test-results/20260817T210000Z-p13361`, `20260817T210011Z-p16292`, `20260817T210014Z-p16741`, `20260817T210021Z-p17287`, and `20260817T210037Z-p17896`; `make agent-finalize` passed with `RESULTS_DIR` unset at `20260817T210056Z-p18907`, recording the required retained-run maintenance skip because no successful current-source full check then existed. The fresh `make check` passed 622/622 in 332.083s at `.cartulary/test-results/20260817T210119Z-p21824`; its immediately warm `make release-check` passed 778/778 units and all 1,202 owner-derived rows in 850.258s at `.cartulary/test-results/20260817T210703Z-p170525` | The accepted release is 243.899s and 22.3% faster than the 1,094.157s starting reference and 25.068s inside the 875.326s ceiling. It recorded 456 cache hits, 35 misses, and 287 intentional bypasses; one release-inventory producer invocation followed by fresh non-cacheable license and SBOM validators; one sealed production fixture construction with all six contributions, 20,000 Timeline rows, unchanged semantic digest, and closed connections; 1,202/1,202 passed row artifacts; 224/224 released fixture leases; no failed, skipped, or cancelled unit; and lifecycle-v2 `cleanup_succeeded`. Service evidence is 302 files and 1,944,855 bytes, 91.0% fewer files and 88.0% fewer bytes than the PERF-080 baseline. The retained-secret scan passed over 4,914 files; no private recovery ledger, legacy service artifact, or executable Markdown read/stat/hash reference remains | Final ownership and compatibility are recorded below. The primary residual operational risk is only timing margin under materially different hosts: PostgreSQL was saturated for 409.860s and object storage for 299.650s, while the production fixture took 191.870s under release contention. These are measured current bottlenecks, not unexplained failures, and no correctness, isolation, redaction, evidence, or cleanup gate was weakened. The post-terminal tracker `make lint-markdown` passed at `.cartulary/test-results/20260817T212532Z-p378774` and `git diff --check` passed; repeat both after this evidence-only record and hand off the accepted roots |

## 10. Final handoff

### 10.1 Accepted evidence and ownership

| Boundary | Final owner and accepted state |
| --- | --- |
| Normative harness behavior | `docs/testing-harness-nlspec.md`; all adopted requirements and acceptance mappings are current |
| Vocabulary and owner navigation | `docs/domain.md`; intentionally unchanged |
| Authored machine contracts | The task-surface owner, work-graph owner, execution topology, cache and PostgreSQL policy registries, schema attachments, row migrations, and all 64 `tools/test_families/*.json` manifests |
| Generated projections | `tools/task_surface.generated.mk`, `tools/task_surface_manifest.json`, `tools/execution_topology_render_index.json`, and `internal/gen/contractextensions/artifacts_gen.go`; generated only from owners |
| Harness implementation | `tools/harness/**`, `tools/testservices/**`, and the focused release-evidence validators/generator |
| Resource and fixture implementation | `internal/testutil/**` plus source-owner performance-fixture adapters under `internal/modules/**` |
| Browser integration | The changed `apps/web/**` runtime, collaboration, Timeline, Playwright, and browser-test boundaries |
| Developer guidance | `docs/guides/cartulary_implementation_testing_guide.md` |
| Execution and handoff state | This tracker; final accepted check and release roots are recorded in the terminal PERF-160 row |

### 10.2 Compatibility and migration result

- Every public Make target and the public `owned|attach` service session surface
  remains current.
- Internal `postgres_group`, PostgreSQL package reset, `BeginRollbackTxT`, the
  ambient active variable, historical contract ledgers, renderer-cycle
  exclusions, repeated production fixture assembly, retained successful
  resources, and duplicate release-inventory production are removed without
  compatibility readers or writers.
- Scope v1, web stack v5, license report v1, build diagnostics v1, family
  manifest v5, fixture lease v2, work graph v4, work-graph owner v1, target plan
  v3, fixture-tier proof v2, and PostgreSQL fixture-policy v1 are rejected
  superseded identities. Their current replacements are attached directly and
  historical evidence remains archive-only.
- Object-store package reuse, lifecycle-v2, exact product and fixture semantics,
  public command names, CycloneDX validity, and canonical release artifact paths
  retain clear current value and remain supported.
- Exact-symbol Go rows may declare `process_isolation="exclusive"` only for
  proven process-global state. All other compatible service-backed Go rows share
  process startup while retaining distinct per-test resources and cleanup.

### 10.3 Complete changed-path inventory

The terminal worktree contains 208 changed or new files. The inventory below is
lossless by repository path set; a directory glob is used only when every file
in that set changed.

- Root: `Makefile`.
- Web, 15 files: the changed files under `apps/web/e2e/**`,
  `apps/web/playwright.shared.config.ts`, `apps/web/src/app/App.tsx`, and the
  changed Workbook shell, collaboration coordinator, mutation-runtime,
  Timeline composition/loader, and row-mutation coordinator files under
  `apps/web/src/workbook/**`, including the new
  `WorkbookMutationRuntimeRegistry.ts`.
- Contracts, one file: `contracts/extensions/dependencies.json`.
- Documentation, five files: the implementation testing guide, this tracker,
  the Testing Harness NLSpec, Workbook Interaction specification, and source
  traceability matrix. `docs/domain.md` is not changed.
- Backend, 36 files: the changed server runtime, generated contract-extension,
  database-migration, Extensions, Incident Bundles, Network Flow, Projections,
  harness-runtime, HTTP, WebSocket, application performance-fixture,
  performance-fixture lifecycle, PostgreSQL test, object-store test, and suite
  service files under `internal/**`; the new `postgrescatalog/coordinator.go`,
  three `postgrescleanup/*` files, lifecycle `postgres_test.go`, and private
  `resource_ledger.go`; and the removed repeated-production integration test.
- Harness family projections, 64 files: every current
  `tools/test_families/*.json` owner manifest.
- Schemas, 26 paths: removal of the eleven superseded ledger, scope, stack,
  diagnostics, policy, family, fixture, plan, graph, and graph-owner schemas;
  addition of their current v2--v6 replacements plus license-report v2 and the
  journal, private-ledger, and compact browser-admission schemas.
- Harness implementation, 33 files: the changed browser, contract,
  diagnostics, generated-artifact, output, performance-fixture, scheduler,
  service-session, test-catalog, and semantic contract-test files under
  `tools/harness/**`, including the new journal reader.
- Remaining tools, 27 paths: the execution topology and render index, frontend
  source ownership, drift scratch inputs, cache registry, removed historical
  contract ledgers, schema attachments, work-graph owner, PostgreSQL policy,
  recovery browser restore tool, seven release-evidence generator/validator/test
  files, task-surface projection and owners, row migrations, and five
  `tools/testservices/**` files.

### 10.4 Rejected candidates and residual risks

Every failed or invalid attempt, retained root, and rejected candidate is
recorded in the workstream transitions above. The durable rejected classes are
legacy schema bridges, assertion weakening, repeated production builds, retrying
failed rows, retaining successful resources, unconditional forced cleanup,
ambient service authority, increased service capacity without full closure,
scheduler affinity/high-water reservations that regressed broad time, hash
striping that formed catalog convoys, and cacheability outside the paired release
producer.

Current residual risks are bounded and assigned:

- Testing Harness owns future scheduling work if PostgreSQL or object-store
  saturation consumes the 25.068-second release margin on comparable hosts.
- Entities and Timeline owners retain their production fixture contribution
  performance; their exact semantics and one-build policy must not be weakened.
- Release tooling owns scanner availability and version variance through typed
  identities and cache inputs.
- Resource owners retain eager normal cleanup, exact ownership proof, and bounded
  force fallback; suite teardown remains recovery-only.
