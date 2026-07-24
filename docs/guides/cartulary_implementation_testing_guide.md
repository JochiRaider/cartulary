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
