# Cartulary Repository Bootstrap Guide

**Status**: Implementation-support guide
**Authority**: The normative Core documents own product behavior;
`docs/testing-harness-nlspec.md` owns harness mechanics; `docs/domain.md` owns
domain vocabulary and concept boundaries.

## Goal

A bootstrapped Cartulary repository has stable architectural boundaries,
deterministic generation, managed local services, an owner-first test catalog,
and reusable test composition before feature behavior expands. Compilation alone
is not closure.

This guide describes the current repository shape. It does not preserve a
delivery-stage implementation plan, copied test inventory, or alternate command
surface.

## Required Architecture

Keep the modular-monolith boundaries explicit:

- `cmd/*` packages are binary composition roots only.
- `internal/app/server`, `internal/app/migrate`, and `internal/app/operator` are
  the application facades for their commands.
- `internal/platform/*` owns transport, configuration, storage adapters, auth
  primitives, runtime plumbing, and job shells.
- `internal/modules/*` owns domain and application behavior.
- `internal/testutil/*` owns reusable backend harnesses and fixtures.
- `contracts/*` is the derived repo-local contract layer.
- `db/migrations` and `db/queries` are authored SQL inputs.
- `apps/web` is the web application; `packages/*` contains shared TypeScript
  packages.
- `tools`, `scripts`, and `configs/dev` own repository automation and local
  configuration inputs.

Source owners construct their revision providers. Generic revision coordination
validates and runs the complete catalog without absorbing source-specific
behavior.

## Toolchain and Dependency Bootstrap

Use the repository's pinned toolchain and public Make surface:

```sh
make doctor
make bootstrap
make toolchain-drift
```

Do not invoke package managers or language tools as substitute automation when a
Make-owned wrapper exists. Do not hand-edit lockfiles or tool-managed install
artifacts.

The canonical Go module path is `github.com/JochiRaider/cartulary`. The pnpm
workspace owns `apps/web` and shared packages. Tool versions live in
`tools/toolchain_pins.json`; mirrored version text must pass drift validation.

## Local Services

The development environment provides real PostgreSQL and an S3-compatible object
store through repository-owned service targets:

```sh
make services-up
make db-up
make db-migrate
make object-store-init
```

Local development keeps its object-store CORS proxy on
`127.0.0.1:8333`. Its lease is proof-gated and browser tests never attach to it.
If startup reports `resource_conflict`, inspect and stop the unrelated listener
yourself, then rerun the Make target; the lifecycle deliberately will not signal
a listener it cannot prove it owns. Legacy proxy PID files are not ownership
evidence. `make services-down` stops a proven local proxy and preserves named
development volumes.

Use `make db-reset` only when resetting repo-local development data is intended.
Application tests that claim storage, migration, transaction, object-lifecycle,
or service-integration behavior must use the applicable managed-service harness,
not mocks that cannot exercise those boundaries.

The server runtime owns only resources it creates. Borrowed Postgres and object
store dependencies are never closed by the runtime. Cleanup is reverse-order and
idempotent.

## Generation Bootstrap

Establish one authored-input-to-generated-output path before adding derived
contracts:

```sh
make generate
make generate-drift
make generated-artifact-policy-check
make json-shape-check
```

Generated roots declared by `tools/generated_artifact_policy.json`, generated
topology, generated schedules, and the generated task surface are read-only.
Change their owner inputs and regenerate through Make.

Bootstrap generation should prove:

- authored contracts validate before code generation;
- output ordering and bytes are deterministic;
- stale, missing, and unexpected outputs are detected;
- ordinary generation owns topology and schedule outputs;
- a clean generated tree is reproducible from committed inputs.

## Harness Bootstrap

The executable model is:

```text
owner -> family -> semantic row -> work unit -> owner artifact shard
```

`tools/test_catalog_owner.json` registers owners and authored family manifests.
`tools/test_families/*.json` owns semantic rows and exact selectors.
Verification contracts own the postconditions those rows support. Authored
topology owns runtime, fixture, isolation, dependency, and resource policy.

Create reusable harnesses before adding rows that depend on them. The baseline
includes:

- in-process HTTP envelopes and public error checks;
- process readiness and diagnostics;
- real Postgres, migration, and object-store fixtures;
- authorization re-derivation and tenant isolation;
- idempotent replay and divergent-replay checks;
- projection determinism and rebuild support;
- WebSocket connection, reconnect, and cleanup behavior;
- exact Go, Vitest, shell, and Playwright selector-result adapters;
- owner accounting, summary, and evidence-audit support.

Every selected row must resolve exactly once and emit an independently
identifiable terminal result. Aggregate command success cannot stand in for
selector evidence.

## Initial Implementation Workflow

Discover the current owner surface instead of copying row IDs into a guide:

```sh
make help
make help-all
make task-guide ROLE=module-author OWNER=<owner-id>
make explain-test-owner OWNER=<owner-id>
```

For each small owner change:

1. Select one semantic row or a cohesive family.
2. Write the exact failing test and catalog selector.
3. Implement the smallest owner-cohesive behavior.
4. Run `make test-slice OWNER=<owner-id> ROWS=<row-id,...>`.
5. Run the service-backed slice when the runtime profile requires managed
   services.
6. Broaden to the complete owner, generated drift, and `make check`.
7. Refactor only behind green owner evidence.

Use descriptive language-level test names. Stable selection identity belongs in
the catalog, not in encoded delivery labels inside symbols, fixture data, or
comments.

## CI and Release Shape

Wire CI around `make ci`, not a handwritten subset of local commands. CI must
prove deterministic generation, migration applicability, owner-row execution,
cleanup, deployable composition, and artifact schema validity.

Release readiness uses the same owner accounting and finalizer contract as local
verification. A separate CI-only registry or reader is not permitted.

## Bootstrap Definition of Done

Bootstrap is complete when:

- the repository boundaries above exist and compile;
- `make doctor`, `make generate-drift`, and the generated-artifact checks pass;
- managed PostgreSQL and object-store services start, report readiness, and
  clean up through public targets;
- migrations apply to an empty database and expose current history evidence;
- reusable process, storage, auth, projection, and browser harnesses exist;
- the owner catalog and verification contracts validate;
- representative unit, service-backed, browser, and failure-path rows emit valid
  per-owner accounting;
- `make test-fast` and `make check` pass on the bootstrapped tree;
- generated outputs and retained result roots do not leave untracked authority;
- the handoff records exact successful commands, result roots, skipped checks,
  and the rollback commit.

From that state, feature work proceeds by owner and semantic evidence obligation.
New capabilities extend the catalog and generic scheduler rather than creating a
new execution model.
