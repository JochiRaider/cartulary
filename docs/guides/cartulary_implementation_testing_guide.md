# Cartulary Implementation and Testing Guide

**Status**: Implementation-support guide
**Authority**: Core 00-04 own product behavior. `docs/testing-harness-nlspec.md`
owns harness mechanics. This guide does not create product requirements, catalog
rows, verification claims, or command contracts.

## Purpose

This guide describes the durable workflow for implementing Cartulary changes and
collecting repository-owned test evidence. It intentionally does not copy the
catalog, target inventory, requirement tables, or verification contracts. Those
inputs evolve independently and must be read from their owners.

## Read the Owners First

Before changing behavior:

1. Read `docs/spec/00_document_set_status_and_precedence.md` and the applicable
   Core 01-04 owner documents.
2. Read `docs/domain.md` for domain-facing terminology, record models, workbook
   surfaces, and concept boundaries.
3. Read `docs/testing-harness-nlspec.md` for selection, scheduling, fixtures,
   artifacts, cleanup, and audit behavior.
4. Inspect `tools/test_catalog_owner.json` and the applicable authored family in
   `tools/test_families/`; do not infer ownership from test names or directories.
5. Use `make task-guide ROLE=module-author OWNER=<owner-id>` to obtain the
   current narrow workflow for that owner.

Supporting guides under `docs/guides/` help implement an adopted contract but do
not supersede an owner specification or close conformance by themselves.

## The Executable Testing Model

The harness reads one owner-first model:

```text
owner -> family -> semantic row -> work unit -> owner artifact shard
```

- An owner identifies the subsystem responsible for the behavior.
- A family groups related semantic evidence obligations.
- A row is the stable unit of selection and accounting.
- A work unit is a scheduler operation selected from one or more rows only when
  its adapter can report each selected row independently.
- An owner artifact shard records exact expected and observed terminal evidence
  for one target partition.

Catalog rows are semantic identities, not delivery milestones. Never add a
parallel registry, title-prefix convention, handwritten target map, or runtime
translation table. Add or change behavior at the owning specification and
catalog inputs, then regenerate through Make.

## Plan a Change

Start with read-only discovery:

```sh
make task-guide ROLE=module-author OWNER=<owner-id>
make explain-test-owner OWNER=<owner-id>
make explain-target TARGET=<target-id> DETAIL=summary
```

Choose the smallest owner and row set that proves the intended change. If the
change crosses owners, keep each owner's behavior and fixtures cohesive, then
run the broader derived target only after the narrow owner checks pass.

For a new or changed row, define:

- the owning product rule or support boundary;
- the semantic family and stable row identity;
- the exact selector and adapter able to report it independently;
- the runtime profile, fixture class, isolation, and reset policy;
- dependencies and resource claims;
- applicable direct-evidence targets;
- expected owner-shard artifacts and any authorized redactions.

Do not create a row solely to mirror a file, a test framework invocation, or a
temporary delivery sequence.

## Implement with Narrow Feedback

Use this loop:

1. Add or update the exact test at the owning boundary.
2. Confirm that the catalog selector resolves once and only once.
3. Run the narrow owner slice:

   ```sh
   make test-slice OWNER=<owner-id> ROWS=<row-id,...>
   ```

4. When the runtime profile requires managed services, run:

   ```sh
   make service-backed-test-slice OWNER=<owner-id> ROWS=<row-id,...>
   ```

5. Inspect failures and retained artifacts with:

   ```sh
   make explain-run RESULTS_DIR=<run-root-or-run-directory>
   ```

6. Broaden to the complete owner slice and the derived targets named by the task
   guide.

Product assertion failures must remain visible as product failures. Do not add
automatic retries, aggregate-success inference, or exception paths that turn a
missing selector result into a pass.

### Browser-runtime mutation recovery

The Workbook pending-mutation runtime is owned above authenticated presentation
by the App's browser-runtime registry. Authenticated Workbook and Collaboration
shells borrow that incident- and client-instance-scoped authority; their
unmount during an HTTP authentication failure or `session_revoked` transition
must detach queries, surfaces, presence, and sockets without disposing queued
unsent writes. In-page re-authentication must commit a new session, rederive
incident authorization, remount the same scope, and resume the existing FIFO.

Entering a different incident retires the prior mutation runtime, App teardown
disposes it, and a reload creates a new registry. Tests must therefore exercise
the real Auth-shell transition and exact post-login replay in one JavaScript
runtime. Do not use reload, API-only cookie replacement, cross-tab persistence,
or Auth-shell copies of Workbook queue controls as recovery shortcuts.

## Generated Contracts and Repository Boundaries

Edit authored inputs only. Run the public generation surface after catalog,
schema, task-surface, or topology changes:

```sh
make generate
make generate-drift
make generated-artifact-policy-check
make json-shape-check
```

Never hand-edit generated roots, generated topology or task-surface outputs,
lockfiles, or tool-managed install artifacts. Keep binary composition in
`cmd/*`, platform mechanics in `internal/platform/*`, domain logic in
`internal/modules/*`, and reusable application test composition in
`internal/testutil/*`.

## Extension Profile Changes

Treat extension adoption as one owner-to-runtime chain:

```text
owner specification and authored contracts
  -> generated plan rows and digests
  -> exact application contribution/worker/participant catalogs
  -> copied serving-epoch projections
  -> production routes, workspaces, workers, and participants
```

Do not add a profile switch, fixed registrar list, broad epoch provider, default
registry, or test-only production setter alongside that chain. A claimed
profile implementation is prepared quiescently, installed by commit,
acknowledged by the exact HTTP, WebSocket, dequeue, and claimed-worker
components, and made available only by the single `Serve` transition. Recovery
and dequeue evidence must prove that no handler runs before serving.

Extension jobs use their generated versioned job kind and canonical worker kind.
Terminal success evidence is complete only when the owner mutation, public
terminal job result, immutable proof, route-scoped idempotency outcome, and
applicable audit/resource references share one final transaction. Cancellation
evidence must cover the durable precommit observation and the success-wins-race
case. Public job assertions must also prove that internal profile/job ownership
metadata is absent from the resource.

An extension participant row is not evidence of invocation by itself. Its
semantic owner must provide a production-path selector showing exact catalog
admission, bounded invocation, closed result validation, owner-controlled output
admission, and no effect when inactive, absent, malformed, oversized, timed out,
cancelled, or failed. For Snapshot/Reporting, Report Composition owns the
immutable preview source while Reporting owns participant invocation, result
admission, redaction, rendering, and the terminal job/proof transaction.

For an ordered multi-workstream remediation, the named handoff tracker is the
controlling execution artifact. Mark exactly one slice `IN_PROGRESS` before
changing its surfaces. After implementation, record files, decisions, commands,
run roots, failures, risks, rollback and domain posture; mark the slice `DONE`
only after its exit criteria pass; commit that implementation and checkpoint
together; then activate the next slice. The final validation slice changes only
validation, accounting, or handoff defects.

## Reuse Local Test Services Explicitly

Ordinary harness commands own and remove their managed PostgreSQL and object-store
containers. For repeated local work on a small service-backed slice, create one
explicit expiring session, inspect only its redacted status, attach the eligible
target, and stop the session when finished:

```sh
make test-services-session-up
make test-services-session-status
make service-backed-test-slice \
  OWNER=<owner-id> \
  ROWS=<row-id,...> \
  CARTULARY_TEST_SERVICES_MODE=attach
make test-services-session-down
```

The default descriptor is
`${CARTULARY_MACHINE_CACHE_DIR}/test-services/session.json`. Set
`CARTULARY_TEST_SERVICES_SESSION_FILE=/absolute/path/session.json` on every
command only when an explicit alternate location is required. A descriptor is
never attached merely because it exists. `attach` is accepted only by
`service-backed-test-slice`, `browser-e2e-webserver-backed`, and
`browser-e2e-stateful`; aggregate, measurement, accessibility, visual, CI, and
release targets always own fresh services.

Each attached invocation still receives unique databases, buckets, ports,
browser contexts, runtime roots, and cleanup ownership. Session shutdown refuses
live borrowers. If status reports `stale`, `expired`, or `invalid`, do not edit
the descriptor or remove containers by name; use the session-down command so
cleanup can verify exact ownership. Status output intentionally omits credentials,
container identities, runtime paths, and administrative endpoints.

The retired ambient service-active boolean is no longer produced or consumed.
Child authority comes from the exact suite identity, runtime or borrower lease,
service metadata, readiness generation, and container proof. The generic
reserved `CARTULARY_TEST_SERVICES_*` input boundary rejects caller-supplied
internal state; use `CARTULARY_TEST_SERVICES_MODE=attach`. There is no selector
alias or automatic legacy attachment path.

Service evidence uses a hard-cut current model. Each producer appends bounded
records to one owner-only NDJSON journal; `service-scope.json` is the bounded
`cartulary.test_services.scope.v2` diagnostic and never contains exhaustive
database, bucket, package, or test inventories. Exact live resource identities
remain in a mode-0600 cleanup ledger that is deleted after successful cleanup.
Fixture totals and the distinct strategy count remain exact, while the retained
strategy diagnostic is capped at the deterministic 32 most costly aggregates;
the ten slowest individual fixture activities remain separately bounded.
Browser sessions retain only a compact
`cartulary.test_services.browser_admission.v1` proof and publish
`cartulary.web_e2e_stack.v6`; scope v1, per-event JSON directories, copied scope
admissions, and stack v5 are historical archives rather than current inputs.

Per-test Postgres databases are eager-cleanup resources. Create unmigrated
scratch databases with `NewDatabaseT`, current-head clones with
`PrepareIsolatedDatabaseT`, and migration-history scratch databases with the
migration capability; each helper registers deletion before callers register
their runtime, job, pool, and handle finalizers, so those borrowers close first.
Cleanup attempts an ordinary exact-owner drop and uses one bounded PostgreSQL
`DROP DATABASE ... WITH (FORCE)` only after ownership proof when that attempt
reports active connections or exhausts its separate bounded budget while parent
cleanup is still active. Parent cancellation prevents fallback. Keeping forced
termination and deletion in one database operation prevents catalog pressure
from exhausting a shared terminate-then-drop budget. Every harness-owned create,
clone, template-state mutation, ordinary exact-owner drop, and forced drop first
holds a target-specific PostgreSQL session-advisory lock. Same-target mutations
cannot overlap, while a slow target cannot stop unrelated catalog work.
Admission uses PostgreSQL's lock queue without polling and releases through a
fresh bounded context. Normal and forced operations keep independent budgets,
and fallback reacquires admission only after the ordinary attempt has released
it. Exact owner proof and target exclusion make service-global quiescence
unnecessary. Global catalog concurrency is bounded only by typed scheduler
PostgreSQL claims; hash stripes are not a second, collision-prone admission
layer. A fixture that deliberately overlaps targets must reserve the matching
PostgreSQL capacity.
Service-backed Go work also reserves PostgreSQL capacity honestly: its
`GOMAXPROCS`, CPU claim, and PostgreSQL claim use the same captured-capacity-
and topology-claim-bounded value. The ordinary service minimum declares one
lane, so each package process runs Go tests one at a time while independently
compatible package processes may overlap up to the admitted PostgreSQL
capacity; browser reservations reduce that admission further. Exact-symbol rows
with identical target, package, profiles, dependencies, evidence class, and
fixture capability share one deterministic process. Dedicated and migration
helpers still create and eagerly retire a distinct database for every
`testing.T`, and per-test object-store helpers retain the same ownership rule;
only the process startup is shared. `managed_process` rows remain separate
because process lifecycle is their subject. The exceptional Jobs telemetry
capture row declares typed exclusive process isolation because it installs and
observes process-global OpenTelemetry providers; future exceptions require the
same concrete process-global proof and must not use isolation as a scheduling
tuning knob. This prevents hidden package-level
`t.Parallel` fan-out from overwhelming catalog cleanup without multiplying Go
startup and package initialization cost by row count. A specialized topology
claim may request greater parallelism only by reserving the same number of
PostgreSQL lanes.
Suite teardown and stale recovery handle only failed, interrupted, or abandoned
cleanup. Successful
database-retention events no longer exist.

Use `BootstrapBucketT` for a per-test object-store bucket so the bucket is
deleted after its stores and runtimes close. `PreparePackageBucketT` is the
intentional exception: it serializes and resets one package-scoped bucket for
reuse, and callers must not reinterpret it as a per-test ownership boundary.
Broker-owned browser and performance-fixture buckets retain their explicit
lease finalizers. Stateful browser renewal empties its stable session bucket in
place and completes a put/head/delete readiness proof before admitting the next
generation; it does not delete and recreate the namespace. The session
finalizer still deletes the bucket exactly once.

The first executable browser group owns its stack lease directly; there is no
synthetic readiness borrower between stack publication and attachment. Reset
and later group units retain the same healthy affinity allocation, and the
terminal group releases it. This keeps immutable stack-v6 and compact admission
evidence available throughout the consuming command while preserving exact
process, port, database, bucket, and private-runtime cleanup. Session startup
binds both readiness timestamps after their probes and fails unless the broker
can verify the complete owner-only stack-v6/admission pair before returning the
allocation; a diagnostics-only session is never attachable. Publication steps
propagate their status explicitly through timing wrappers, and retained session
directories are single-use identities: a collision is rejected rather than
cleared or republished.

## Harness Evidence and Cache Cutover

Current harness runs emit `cartulary.harness_unit_event.v2`, compile
`cartulary.harness_work_graph.v5`, and read or write only
`cartulary.harness_cache_record.v2`. There are no compatibility readers or dual
writes for the retired event v1, graph v3, or cache-record v1 formats. Historical
result roots remain manually inspectable archives, but cannot be retained as
current-run evidence or supplied as a cache input.

The v5 graph distinguishes current-run evidence from reusable artifacts. A
cache hit always regenerates row and unit evidence with current timestamps;
artifact bytes are reusable only after the producer declares their complete
typed output and freshness contract. `release-inventory-artifacts` is the sole
producer of the canonical SBOM and license-report pair and is the only release
inventory unit with content-addressed caching. `license-report` and `sbom` are
non-cacheable validators: their direct Make targets require the producer, while
graph children consume the explicit producer edge without running local
generation prerequisites and still retain fresh validation evidence.

The current deterministic pair is
`.cartulary/release-artifacts/sbom.cyclonedx.json` and
`.cartulary/release-artifacts/license-report.json`. The latter validates only as
`cartulary.license_report.v2`; the v1 wrapper is not translated. Canonical bytes
contain semantic dependency and license facts but no timestamp, checkout path,
commit, run root, or command transcript. Scanner captures and copied license
texts are temporary producer work and are removed before exit. Cache-v2 records
declare exactly both mode-`0644` repository files; invalid or incomplete entries
are quarantined and execute as misses.

Go test-dependency inventory is resolved only from the explicit first-party
package roots `./cmd/...`, `./db/...`, `./internal/...`, and `./tools/...`.
Repository-local `tmp`, result, cache, and dependency-install trees are never
package discovery inputs, so test-owned create/remove activity cannot race the
release producer. A future top-level Go package root must be added explicitly to
both this producer contract and its cache inputs.

Cache v2 entries live under `.cache/cartulary/graph-v2`. The former graph cache
root is ignored, not migrated. Use `make distclean` when the old local cache
bytes should be removed; do not copy or translate entries between roots.

## Performance Fixture Construction

The graph-owned `fixture_snapshot` unit is the only path that constructs the
exact 20,000-row production fixture. Any nonempty selection of its four
Timeline measurement rows shares one builder, while a non-measurement closure
constructs none. Confirm the current-run `build-diagnostics.json` reports
`cartulary.performance_fixture_build_diagnostics.v2`,
`construction_count=1`, `failure_stage=none`, and completed contribution
states; that diagnostic is deliberately separate from the immutable semantic
receipt. Failed diagnostics retain only a bounded stage and safe ordered
contribution identities, states, and durations. They never retain error text or
physical resource identities.

The builder records every PostgreSQL backend process ID opened by its pool.
After pool closure, only those exact connections may use the bounded server-side
drain window before sealing. An unrecognized connection fails immediately and
is never terminated by the builder. Drain exhaustion leaves the template
unsealed and invokes failed-build cleanup.

The snapshot lifecycle integration uses a bounded deterministic profile to
exercise four-clone isolation, corruption rejection, credential handling, and
cleanup without rebuilding production volume. The retired independent
source-owner production assembly row is recorded in
`tools/test_catalog_row_migrations.json`; direct assembler contracts plus the
canonical production snapshot read-back own its continuing coverage.

## Browser Work

Browser rows are grouped by semantic stage, runtime profile, fixture and
isolation policy, and exact Playwright selector. Do not introduce ordering or
selection environment variables outside those contracts.

Use the derived targets selected by the catalog, including the ordinary,
webserver-backed, stateful, accessibility, visual, and measurement surfaces.
Every selector must resolve exactly once and produce per-row evidence.

For committed visual snapshots, follow
`docs/guides/cartulary_visual_golden_maintenance.md`. A golden update is an
explicit maintenance operation, not a way to make a failed comparison pass.

## Evidence and Audit

Successful process exit is not row evidence. Each applicable `(owner, target,
row)` partition must have exactly one expected record and exactly one observed
terminal record in the uniform owner layout:

```text
<target>/owners/<owner-id>/test-evidence-accounting.json
<target>/owners/<owner-id>/test-owner-summary.json
```

Audit exact retained roots with an ignored, schema-valid root manifest:

```sh
make test-evidence-audit \
  OWNER=<owner-id> \
  EVIDENCE_ROOTS_FILE=<manifest-path>
```

The manifest maps each target ID to an exact run root. Do not discover the
newest run, reuse stale evidence after source changes, or ingest historical
artifact formats automatically.

## Broad Verification and Handoff

After narrow checks pass, select broader validation in proportion to risk. The
common progression is:

```sh
make test-fast
make check
make agent-finalize
```

Run `make agent-finalize RESULTS_DIR=<compatible-successful-run-root>` only when
the retained run matches the current source, catalog, verification contract,
selected-row, and applicable-profile identities.

A handoff must record:

- the owner and semantic rows changed;
- authored and generated files changed;
- exact commands and successful retained run roots;
- failures encountered and how they were resolved;
- skipped checks with concrete reasons;
- compatibility or migration effects;
- the clean rollback commit.

Do not declare closure after modifying tracked bytes used by the validation.
Final release and audit evidence must be collected from the exact committed tree
named by the handoff.
