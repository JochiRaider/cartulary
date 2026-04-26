# Cartulary Repository Bootstrap Guide

## 1. Purpose

This guide defines the first concrete steps for creating a new Cartulary repository in an empty folder and making it usable for TDD from day one. It assumes a greenfield start, uses the normative core as the behavior owner, and uses the development guide plus the progressive implementation and testing guide as implementation-support inputs rather than as independent authorities.[^1]

The target outcome is not “a repo that compiles.” The target outcome is a repo that already has the right architectural boundaries, command surface, local services, contract/codegen plumbing, and reusable test harnesses so that Codex can work phase by phase without inventing architecture midstream.[^2]

## 2. What the bootstrap must produce

At bootstrap completion, the repository should satisfy all of the following:

- one modular-monolith application skeleton, not a set of future microservices.[^3]
- one root Go module and one top-level pnpm workspace with the baseline monorepo layout already in place.[^4]
- local PostgreSQL and S3-compatible object storage available through `docker-compose.dev.yml` for development and integration tests.[^3][^5]
- a root `Makefile` with the canonical developer task surface, including `help`, `doctor`, `bootstrap`, `db-up`, `db-reset`, `dev`, `generate`, `test`, `lint`, `check`, `build`, `clean`, and `distclean`.[^6]
- a live contract derivation path from owner sections to `/contracts/*` to generated code, with drift detection wired into `make check`.[^7]
- reusable unit, integration, and black-box end-to-end harnesses, plus the shared cross-cutting harnesses that must apply across phases.[^8]
- the first red tests for Phase 0 checked in before any feature code beyond the bootstrap shell is treated as complete.[^9]

## 3. Scope boundaries for the first implementation window

For a greenfield backend-first start, the implementation window should cover bootstrap infrastructure, then Phase 0, then Phase 1, then the incident shell in Phase 2, and only then the first real record-row mutation path in Phase 3.[^10]

Do not start with file-based import, reporting, reference packs, portability, or enterprise auth. Those are extension-profile concerns or later-phase work, even though the repo should reserve the internal module boundaries and directories for them from the beginning.[^10][^11]

Do not start by building forms-first CRUD. Cartulary is explicitly grid first, preserves the spreadsheet mental model at the view layer, and expects the record mutation substrate to be present as soon as the first hot-path workbook mutation exists.[^12][^13]

## 4. Step 1: create the repository control surface

Start in an empty folder and create the control files first. For Cartulary, the repo-control facts are already fixed and should be used directly.

```text
Repository remote: https://github.com/JochiRaider/cartulary.git
Go module path: github.com/JochiRaider/cartulary
Supported toolchains: Go 1.26 with toolchain go1.26.2; Node 24.15.0; pnpm 10.33.0
```

Create these root files immediately:

- `README.md`
- `AGENTS.md`
- `Makefile`
- `docker-compose.dev.yml`
- `.editorconfig`
- `.env.example`
- `go.mod`
- `package.json`
- `pnpm-workspace.yaml`
- `biome.json`
- `tsconfig.base.json`

`AGENTS.md` should exist from the beginning because the development guide treats it as the intended owner for contributor and coding-agent procedure, including the repo map, canonical command surface, generated-file edit prohibitions, and local execution procedure.[^14]

Recommended first commands:

```bash
git init
go mod init github.com/JochiRaider/cartulary
pnpm init
```

Recommended first directory command:

```bash
mkdir -p \
  cmd/server cmd/migrate \
  internal/app \
  internal/platform/httpapi internal/platform/ws internal/platform/jobs \
  internal/platform/postgres internal/platform/objectstore internal/platform/authn internal/platform/config \
  internal/modules/auth internal/modules/incidents internal/modules/timeline internal/modules/entities \
  internal/modules/evidence internal/modules/imports internal/modules/links internal/modules/revisions \
  internal/modules/projections internal/modules/reference_data internal/modules/reporting internal/modules/collaboration \
  internal/gen/contracts internal/gen/sql \
  db/migrations db/queries \
  contracts/openapi contracts/ws contracts/view-schemas contracts/errors \
  apps/web packages/ui packages/grid-adapter packages/protocol-ts/src/generated packages/view-contracts packages/test-utils \
  scripts tools docs configs/dev internal/testutil/configtest internal/testutil/pgtest \
  internal/testutil/s3test internal/testutil/httptestx internal/testutil/wstest \
  internal/testutil/fixtures internal/testutil/golden
```

Do not try to finalize dependency versions in the guide itself. Pin them in the repo-control files once chosen, because those files become the source of exact toolchain truth for the repository.[^15]

## 5. Step 2: lay down the baseline monorepo tree

Create the baseline directory structure before writing business logic. The intended baseline is a polyglot modular-monolith monorepo with a Go application layer, a TypeScript browser layer, and a repo-local contract layer that keeps them aligned.[^4]

Use this tree as the initial repository shape:

```text
/
  README.md
  AGENTS.md
  Makefile
  docker-compose.dev.yml
  .editorconfig
  .env.example
  go.mod
  package.json
  pnpm-workspace.yaml
  biome.json
  tsconfig.base.json

  /cmd
    /server
    /migrate

  /internal
    /app
    /platform
      /httpapi
      /ws
      /jobs
      /postgres
      /objectstore
      /authn
      /config
    /modules
      /auth
      /incidents
      /timeline
      /entities
      /evidence
      /imports
      /links
      /revisions
      /projections
      /reference_data
      /reporting
      /collaboration
    /gen
      /contracts
      /sql

  /db
    /migrations
    /queries

  /contracts
    /openapi
    /ws
    /view-schemas
    /errors

  /apps
    /web

  /packages
    /ui
    /grid-adapter
    /protocol-ts
    /view-contracts
    /test-utils

  /scripts
  /tools
  /docs
```

This structure matches the intended baseline in the development guide and keeps transport, runtime plumbing, and storage adapters under `internal/platform`, while domain and application logic live under `internal/modules`.[^4][^16]

`/packages/grid-adapter` owns the direct `react-data-grid` integration. `/apps/web` must consume Cartulary adapter components and types from `/packages/grid-adapter` rather than importing `react-data-grid` directly. `/packages/ui` remains presentational and must not own workbook state, grid mutation semantics, or vendor-coordinate translation.[^25]

`/packages/view-contracts` and `/packages/test-utils` must also be real workspace packages with their own package manifests, TS config, and source exports. They should follow the same source-export pattern used by `/packages/protocol-ts`: `/packages/view-contracts` stays a thin parser or adapter over generated contract artifacts, while `/packages/test-utils` owns shared selector factories and browser helpers reused by both functional and visual workbook suites.

Two rules matter immediately:

- keep the backend as one root Go module and the frontend as one top-level pnpm workspace.[^4]
- create the module directories now, even when many are still empty, so Codex does not invent alternate module boundaries later.[^3][^16]

## 6. Step 3: scaffold the backend composition root before handlers

Create `/cmd/server` and `/cmd/migrate` first. `/cmd/server` is the application composition root. `/cmd/migrate` is the migration entry point. Schema DDL changes belong in numbered migrations under `/db/migrations`, not in startup side effects.[^16]

Then create the platform packages with empty or stubbed interfaces:

- `internal/platform/httpapi`
- `internal/platform/ws`
- `internal/platform/jobs`
- `internal/platform/postgres`
- `internal/platform/objectstore`
- `internal/platform/authn`
- `internal/platform/config`

Their initial responsibilities should match the development-guide split: HTTP envelopes and middleware in `httpapi`, WebSocket lifecycle in `ws`, job shell in `jobs`, `pgx` pool and transaction helpers in `postgres`, S3-compatible storage access in `objectstore`, password and session primitives in `authn`, and deployment-config plus runtime-root validation in `config`.[^16]

Recommended first compile target:

- `cmd/server/main.go` boots config loading, logger creation, Postgres wiring, object-store wiring, and HTTP route registration.
- `cmd/migrate/main.go` loads config, opens Postgres, and applies `goose` migrations.
- both binaries compile before any domain module is implemented.

## 7. Step 4: pin the backend dependency baseline

Start with the backend runtime dependency family already called out in the development guide:

- `github.com/jackc/pgx/v5`
- `github.com/coder/websocket`
- `github.com/minio/minio-go/v7`
- `github.com/BurntSushi/toml`
- `golang.org/x/crypto`
- `github.com/pquerna/otp`

Use the Go standard library for HTTP routing and most core plumbing, especially `net/http`, `encoding/json`, `log/slog`, `context`, and `net/http/httptest`.[^17]

For development tools, add:

- `sqlc`
- `goose`
- `testcontainers-go`

Do not introduce a Go web framework or ORM into the bootstrap. The baseline explicitly excludes that stack choice.[^17]

## 8. Step 5: create local development services and configuration

`docker-compose.dev.yml` should stand up only the local backing services needed by the modular monolith during development: PostgreSQL and MinIO.[^6][^18]

Create a repo-local sample config such as `configs/dev/config.toml`, but keep the runtime contract aligned with the normative deployment-config rules:

- canonical runtime path defaults to `/etc/cartulary/config.toml`.
- `CARTULARY_CONFIG_FILE` may point the server to an alternate absolute path.
- `CARTULARY__...` overlays nested config keys and must fail on unknown keys.[^19]

Your initial `config.toml` should already include the required runtime roots and bootstrap path keys, even if some are only local-dev paths at first:

- `roots.database_storage`
- `roots.object_storage`
- `roots.backup_storage`
- `roots.reference_pack_storage`
- `roots.temporary_work`
- `roots.export_outputs`
- `bootstrap.first_admin_manifest_path`
- `limits.*` resource-limit keys used by the effective configuration.[^5][^19]

The configuration loader must fail closed on invalid schema ID, invalid deployment profile, missing required roots, invalid path shapes, or out-of-range resource limits, and startup validation must complete before any HTTP listener, WebSocket listener, or background-job runner starts.[^5][^19]

## 9. Step 6: establish the contract layer and code-generation plumbing

Create the contract directories at bootstrap even if they only contain minimal artifacts at first:

- `/contracts/openapi`
- `/contracts/ws`
- `/contracts/view-schemas`
- `/contracts/errors`

The repository derivation chain is owner section or adopted NLSpec to `/contracts/*` to generated Go and TypeScript code to runtime consumers. `/contracts/*` is therefore a repo-local derived artifact layer, not the behavior owner, and generated paths must never become hand-edited source files.[^7]

Recommended bootstrap rule:

- create minimal placeholder artifacts only for what the repository can actually validate today.
- keep `make generate` live from day one.
- make generated outputs explicit and mark them `DO NOT EDIT`.
- fail `make check` if `make generate` changes tracked files.[^7]

For the initial greenfield phase, it is enough for `make generate` to prove the pipeline works on empty or skeletal contract inputs. The important point is that the repo already has one deterministic path for generated code before Codex starts filling in routes and schemas.

## 10. Step 7: define the command surface before feature work

Implement the root `Makefile` next. It should expose the baseline human-facing tasks from the development guide:[^6]

- `make help`
- `make doctor`
- `make bootstrap`
- `make db-up`
- `make db-reset`
- `make dev`
- `make generate`
- `make test-fast`
- `make test`
- `make lint`
- `make check`
- `make ci`
- `make release-check`
- `make build`
- `make clean`
- `make distclean`

Repository-local recommended meanings:

- `make help`: print the grouped root task surface without bootstrapping local toolchains.
- `make doctor`: verify required local tools and pinned toolchain versions without installing them.
- `make bootstrap`: install Go tools, install pnpm dependencies, and prepare local service prerequisites.
- `make db-up`: start PostgreSQL and MinIO through Compose.
- `make db-reset`: recreate the local database and apply migrations.
- `make dev`: run the Go server and, once present, the Vite dev server.
- `make generate`: regenerate `sqlc` outputs and contract-derived outputs.
- `make backend-store`: run the service-backed store-domain `U-*` backend slice that preserves unit-layer phase IDs while using real Postgres.
- `make test-fast`: run the pure backend unit slice, the service-backed backend store and integration slices, explicit support integration and process-smoke coverage, frontend type-checking, and frontend unit tests.
- `make test`: run the authoritative full corpus, including manifest-verified browser E2E and explicit supplemental support suites. Backend service-backed work should run through one service-backed stage scheduled from `tools/service_backed_schedule_manifest.json` by declared resource capacity, with aggregate target summaries finalized after the `cartulary-test-services` wrapper has completed teardown so leak-check, janitor, and service-termination spans remain visible; all browser evidence should then run through one manifest-driven `browser-e2e` owned stack using the `all` batch and explicit reset boundaries. Summary artifacts report wall, critical-path wall, executed, logical, reused, derived, and teardown durations as separate fields, with backend service-backed and aggregate browser duration groups reported separately.
- `make lint`: run Go vet or lint plus frontend lint and type-check.
- `make check`: run the full developer gate and fail if any authoritative phase-manifest row is absent from execution. Keep toolchain drift and frontend install as early setup blockers, prepare build artifacts and service images as the readiness gate, then start service-backed backend phases through the weighted capacity-aware scheduler while parallel-safe static validation, harness smoke, and product checks continue concurrently. Finalize the service-backed aggregate summary after wrapper teardown, fail on hidden lifecycle teardown failures as non-test failures, and run all browser suites afterward through one `browser-e2e` owned-stack batch. Browser phase expansion belongs in `tools/browser_e2e_batch_manifest.json`, not in backend service-backed schedules. Service-backed schedule v3 expands manifest-declared work-unit sources into Go shards and explicit Make targets, and schedule-level `resource_limits` are the concurrency authority.
- `make ci`: run the provider-neutral CI gate that composes the canonical task surface and enforces execution truth, codegen drift, migration verification, and deployable-shape checks.
- `make release-check`: run the release verification tier by composing the developer gate, dependency license report verification, SBOM verification, and release build verification.
- `make build`: build the application artifact with embedded frontend assets.
- `make clean`: remove reproducible repo-local build and report artifacts while preserving checked-in files and external Go caches.
- `make distclean`: additionally remove repo-local tool/runtime caches after printing the removal list.

`make check` is not optional. It is the required developer verification gate and must include codegen drift detection and migration verification.[^6] `make release-check` is the release verification gate and must fail if the required license or SBOM artifacts are missing or empty.

Repo-control helper targets SHOULD also include at least `make frontend-unit`, `make browser-e2e-support`, and `make browser-e2e-visual` so the frontend workspace packages, support browser helpers, and workbook screenshot fixtures can run independently. `make browser-e2e-visual` remains a Playwright screenshot suite under the owned-stack harness rather than a second visual runner.

By the end of bootstrap, `make check` must include a frontend smoke path that proves the browser bundle can import `react-data-grid/lib/styles.css`, render a minimal Cartulary fixture grid through `/packages/grid-adapter`, and key rows by `record_id`. The smoke fixture must include at least two rows with distinct `record_id` values. This smoke path must not assert feature-complete workbook behavior. It exists only to fail early on package-format, CSS-export, peer-dependency, and stable-row-key integration errors.[^25]

## 11. Step 8: bootstrap the TDD harness before the first feature slice

Cartulary’s testing guide is explicitly phase-shaped and expects unit, integration, and E2E coverage, with the shared harnesses in Section 14 implemented once and reused across phases.[^8][^20]

Create the test harness structure before Phase 0 implementation. A practical backend-first layout is:

```text
/internal/testutil
  /configtest
  /pgtest
  /processtest
  /s3test
  /httptestx
  /wstest
  /fixtures
  /golden
```

Recommended responsibilities:

- `configtest`: effective-config fixture loader, overlay helper, invalid-config golden files.
- `pgtest`: Postgres testcontainer startup plus fresh migrated database-per-test helpers.
- `processtest`: real `cmd/server` lifecycle, readiness and health polling, fail-closed connection probes, and startup diagnostics parsing.
- `s3test`: MinIO testcontainer startup, bucket bootstrap, round-trip helper.
- `httptestx`: in-process runtime or HTTP server boot helper, authenticated request helper, and JSON envelope assertions.
- `wstest`: WebSocket connect, handshake, receive, revoke, and close assertions.
- `fixtures`: canonical bootstrap manifests, config artifacts, and payload fixtures.
- `golden`: deterministic expected JSON or diagnostics outputs.

Then implement the shared cross-cutting harnesses as reusable assertions rather than one-off tests:

- envelope consistency.
- authorization re-derivation.
- mutation attribution and history emission.
- idempotent replay and divergent replay.
- closed-vocabulary rejection.
- writable-string normalization.
- view-schema field-key conformance.
- projection determinism and rebuild.
- WebSocket lifecycle behavior.[^8]

The development guide adds one more concrete boundary: backend integration tests must use real Postgres and MinIO through `testcontainers-go` or equivalent real-service harnesses, and they must exercise HTTP routes, WebSocket behavior, object-store lifecycle, projection maintenance, and migration application.[^21]

## 12. Step 9: define the TDD workflow as repository law

Once the harness exists, make the TDD loop mechanical.

For every slice of work:

1. choose one phase row or a very small cluster of rows from the progressive implementation and testing guide.[^9]
2. create the failing unit tests first and include the phase test IDs in the test names or comments.
3. implement the smallest code needed to make those tests pass.
4. add or activate the matching integration tests against real backing services.
5. run `make check`.
6. refactor only behind a green test suite.

Recommended test naming convention:

- `TestPhase0_ConfigDiscovery_U_0_01`
- `TestPhase0_RuntimeRoots_U_0_02`
- `TestPhase1_LoginRequestShape_U_1_01`
- `TestPhase3_TimelinePatchReplay_U_3_07`

That keeps the implementation plan, the test corpus, and Codex work orders aligned to the same identifiers.

Use the phase test IDs as the unit of planning for Codex. “Make U-0-01 through U-0-05 pass” is a good work order. “Implement config” is too vague and tends to drift.[^9][^20]

Also make two failure classes impossible to ignore:

- codegen drift after `make generate`.
- migration drift when schema-affecting changes are not represented in `/db/migrations/*` or migrations do not apply cleanly in CI.[^6][^7]

## 13. Step 10: execute the first implementation slices in order

### Slice A: repository bootstrap only

Goal: create the tree, dependencies, Compose services, `Makefile`, minimal config loader, migration runner, and reusable test harnesses.

Definition of done:

- `go test ./...` compiles and runs the empty or stubbed backend packages.
- `make db-up` starts Postgres and MinIO.
- `make db-reset` can connect and apply an initial migration set.
- `make generate` runs successfully, even if the generated outputs are skeletal.
- `make check` passes on the bootstrap baseline.[^6]

### Slice B: Phase 0

Begin by writing the failing tests for the Phase 0 matrix:

- U-0-01 through U-0-09.
- I-0-01 through I-0-06.
- E-0-01 through E-0-05.[^9]

Implement only enough code to satisfy Phase 0 scope:

- deployment-config artifact loading and overlay.
- runtime-root registry and path validation.
- resource-limit registry validation.
- schema bootstrap and migration idempotency.
- object-store reachability.
- first-admin bootstrap preflight and manifest validation.
- fail-closed startup gating.

Do not treat domain routes as complete at the end of this slice. Phase 0 ends with health or startup diagnostics plus a trustworthy process shell, not incident behavior.[^9]

Recommended package targets for Slice B:

- `internal/platform/config`
- `internal/platform/postgres`
- `internal/platform/objectstore`
- `internal/app`
- `cmd/server`
- `cmd/migrate`
- the minimal administrative schema and audit tables needed for bootstrap state

### Slice C: Phase 1

After Phase 0 is green, write the failing Phase 1 tests and implement the authenticated shell:

- login, logout, session inspection.
- session lifecycle and concurrency cap.
- credential-state inspection.
- password change.
- TOTP begin and complete.
- deployment-local user create and patch.
- admin password reset, TOTP reset, revoke-all.
- `session_revoked` behavior on connected sockets.[^22][^23]

This slice should primarily land in:

- `internal/platform/authn`
- `internal/modules/auth`
- `internal/platform/ws`
- the session and deployment-local admin tables plus audit substrate

### Slice D: Phase 2

Only after the auth shell is stable should you add incident create, list, get, patch, membership routes, extension discovery, and workbook-preference bootstrap.[^10]

This slice primarily lands in:

- `internal/modules/incidents`
- `internal/platform/httpapi`
- contract artifacts for incident, membership, extension, and saved-view families

### Slice E: Phase 3

Only after incident control exists should you implement the first hot-path record-row mutation substrate for Timeline. Phase 3 is where the system first proves that row creation, autosave, projection maintenance, attributed mutation history, lifecycle state, and idempotent patch replay all work together.[^13]

Do not implement Timeline as temporary CRUD. The testing guide is explicit that once the first record-row mutation path exists, it must already emit attributed mutations, maintain projections, and satisfy normalized idempotency and optimistic-concurrency contracts.[^13]

## 14. Step 11: wire CI at the same time as local development

As soon as the root task surface exists, wire CI around `make ci` rather than a handwritten subset of local commands. The provider-neutral entrypoint must already compose generation drift detection, authoritative and support test execution, migration verification, and deployable-shape checks.

The green condition is not just passing tests. CI must also prove that:

- `make generate` leaves a clean diff.
- every authoritative phase-manifest row actually executed in the intended layer.
- migrations apply successfully on an empty database and on an upgrade path.
- the repository still builds as one application artifact rather than drifting toward separate deployables.[^6][^24]

## 15. Recommended Codex work order

Use Codex in this order:

1. create the repository tree and control files.
2. wire the root `Makefile` and `docker-compose.dev.yml`.
3. add the Go composition root, platform stubs, and migration runner.
4. add the contract directories and generation pipeline.
5. add shared test harness packages.
6. write Phase 0 unit tests first.
7. implement until Phase 0 unit tests pass.
8. write and pass the Phase 0 integration tests.
9. move to Phase 1 only when `make check` is green again.

A useful initial Codex prompt is:

```text
Bootstrap a greenfield Cartulary repo as a modular monolith in Go with a pnpm workspace. Create the baseline tree, root Makefile, docker-compose.dev.yml for Postgres and MinIO, cmd/server, cmd/migrate, internal/platform/*, internal/modules/*, db/migrations, db/queries, contracts/*, and a reusable backend test harness. Then write failing Phase 0 tests U-0-01 through U-0-05 before implementing config loading, runtime-root validation, and fail-closed startup.
```

## 16. Bootstrap definition of done

The repository bootstrap is complete when all of the following are true:

- the baseline monorepo tree exists and is committed.[^4]
- the command surface exists and `make check` is the enforced developer gate.[^6]
- PostgreSQL and MinIO can be started locally through Compose.[^18]
- contract/codegen directories exist and generated outputs are treated as read-only.[^7]
- the frontend smoke path renders a minimal `react-data-grid` fixture through `/packages/grid-adapter`, imports `react-data-grid/lib/styles.css`, and keys fixture rows by distinct `record_id` values.[^25]
- reusable shared harnesses exist for in-process HTTP envelopes, real process readiness or diagnostics, authorization re-derivation where applicable, idempotency, projection determinism where applicable, and WebSocket lifecycle.[^8]
- the first failing Phase 0 tests are checked in and can be run repeatedly.[^9]
- no feature work beyond the bootstrap shell has bypassed migrations, config validation, or the TDD loop.[^16][^20]

That state is the correct handoff point for Codex. From there, implementation should proceed phase by phase, with the repository already structured to prevent architectural drift.

## Sources
[^1]: [00_document_set_status_and_precedence.md](sandbox:/mnt/data/00_document_set_status_and_precedence.md), lines 5-13 and 22-33; [cartulary-dev-guide.md](sandbox:/mnt/data/cartulary-dev-guide.md), lines 9-13; [cartulary_implementation_testing_guide.md](sandbox:/mnt/data/cartulary_implementation_testing_guide.md), lines 10-18.
[^2]: [cartulary-dev-guide.md](sandbox:/mnt/data/cartulary-dev-guide.md), lines 79-85, 241-247, 365-371, and 618-643.
[^3]: [01_architecture_storage_and_view_contracts.md](sandbox:/mnt/data/01_architecture_storage_and_view_contracts.md), lines 5-20 and 28-49; [cartulary-dev-guide.md](sandbox:/mnt/data/cartulary-dev-guide.md), lines 79-85.
[^4]: [cartulary-dev-guide.md](sandbox:/mnt/data/cartulary-dev-guide.md), lines 231-347.
[^5]: [04_security_deployment_and_conformance.md](sandbox:/mnt/data/04_security_deployment_and_conformance.md), lines 459-478 and 1629-1820.
[^6]: [cartulary-dev-guide.md](sandbox:/mnt/data/cartulary-dev-guide.md), lines 618-695.
[^7]: [cartulary-dev-guide.md](sandbox:/mnt/data/cartulary-dev-guide.md), lines 363-414.
[^8]: [cartulary_implementation_testing_guide.md](sandbox:/mnt/data/cartulary_implementation_testing_guide.md), lines 869-953.
[^9]: [cartulary_implementation_testing_guide.md](sandbox:/mnt/data/cartulary_implementation_testing_guide.md), lines 32-93 and 958-978.
[^10]: [cartulary_implementation_testing_guide.md](sandbox:/mnt/data/cartulary_implementation_testing_guide.md), lines 26-28, 97-117, 172-188, and 240-255.
[^11]: [01_architecture_storage_and_view_contracts.md](sandbox:/mnt/data/01_architecture_storage_and_view_contracts.md), lines 51-69; [cartulary-dev-guide.md](sandbox:/mnt/data/cartulary-dev-guide.md), lines 487-504.
[^12]: [03_workbook_interaction_collaboration_and_workflows.md](sandbox:/mnt/data/03_workbook_interaction_collaboration_and_workflows.md), lines 5-23.
[^13]: [cartulary_implementation_testing_guide.md](sandbox:/mnt/data/cartulary_implementation_testing_guide.md), lines 18 and 240-299.
[^14]: [cartulary-dev-guide.md](sandbox:/mnt/data/cartulary-dev-guide.md), lines 1058-1065 and 1071-1073.
[^15]: [cartulary-dev-guide.md](sandbox:/mnt/data/cartulary-dev-guide.md), lines 122-133 and 239-249.
[^16]: [cartulary-dev-guide.md](sandbox:/mnt/data/cartulary-dev-guide.md), lines 424-512.
[^17]: [cartulary-dev-guide.md](sandbox:/mnt/data/cartulary-dev-guide.md), lines 135-170 and 203-212.
[^18]: [04_security_deployment_and_conformance.md](sandbox:/mnt/data/04_security_deployment_and_conformance.md), lines 428-434; [cartulary-dev-guide.md](sandbox:/mnt/data/cartulary-dev-guide.md), lines 255-256 and 624-627.
[^19]: [04_security_deployment_and_conformance.md](sandbox:/mnt/data/04_security_deployment_and_conformance.md), lines 1629-1646 and 1654-1715.
[^20]: [cartulary_implementation_testing_guide.md](sandbox:/mnt/data/cartulary_implementation_testing_guide.md), lines 20-28, 891-900, and 958-974.
[^21]: [cartulary-dev-guide.md](sandbox:/mnt/data/cartulary-dev-guide.md), lines 662-688.
[^22]: [cartulary_implementation_testing_guide.md](sandbox:/mnt/data/cartulary_implementation_testing_guide.md), lines 97-168.
[^23]: [04_security_deployment_and_conformance.md](sandbox:/mnt/data/04_security_deployment_and_conformance.md), lines 7-39 and 43-94.
[^24]: [00_document_set_status_and_precedence.md](sandbox:/mnt/data/00_document_set_status_and_precedence.md), lines 128-132; [cartulary-dev-guide.md](sandbox:/mnt/data/cartulary-dev-guide.md), lines 692-695.
[^25]: [R09-react-data-grid-research-report.md](sandbox:/mnt/data/R09-react-data-grid-research-report.md), especially §§1, 3, 18, and 20 on the inspected `react-data-grid` package shape, controlled grid surface, CSS export, and build/package constraints.
