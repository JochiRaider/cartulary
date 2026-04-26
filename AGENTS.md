# Cartulary Repository Procedure

## Authority and placeholders

- Normative behavior is owned by the Cartulary normative core under `docs/spec/00_document_set_status_and_precedence.md` through Core 04. The guides under `docs/guides/` are implementation-support inputs, not independent behavior owners.
- The canonical Go module path is `github.com/JochiRaider/cartulary`.
- Supported toolchain baseline: `Go 1.26` with `toolchain go1.26.2`, `Node 24.15.0`, and `pnpm 10.33.0`.
- Pinned bootstrap tools: `github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0`, `github.com/pressly/goose/v3/cmd/goose@v3.27.0`, and `github.com/testcontainers/testcontainers-go v0.42.0`.

## Repo map and path conventions

- `cmd/server` and `cmd/migrate` are binary entrypoints only. Keep composition roots here; do not place domain logic in `cmd/`.
- `internal/app` is reserved for application assembly shared by the binaries.
- `internal/platform/*` owns transport, runtime plumbing, configuration, storage adapters, auth primitives, and job shells.
- `internal/modules/*` owns domain and application logic inside the modular monolith.
- `internal/gen/**` is generated Go code derived from `/contracts/**` or `/db/queries/**`. Do not hand-edit it.
- `db/migrations` and `db/queries` are authored SQL inputs.
- `contracts/*` is the repo-local derived contract layer. It is downstream of the normative core and upstream of generated code.
- `apps/web` is the single top-level web app in the pnpm workspace.
- `packages/*` is for shared TypeScript packages. `packages/protocol-ts/src/generated/**` is generated and must not be hand-edited.
- `scripts` and `tools` hold repo-local automation and pinned helper tooling.
- `configs/dev` is reserved for local development configuration inputs.
- `internal/testutil/*` is reserved for reusable backend test harnesses and fixtures.
- `cmd/server` and `cmd/migrate` remain bootstrap wiring only in this step. Keep feature behavior out of them.

## Canonical command surface

- `make help`
- `make doctor`
- `make bootstrap`
- `make db-up`
- `make db-reset`
- `make dev`
- `make generate`
- `make format`
- `make phase-ledgers`
- `make phase-ledger-drift`
- `make test-fast`
- `make backend-store`
- `make target-plan`
- `make target-plan-json`
- `make explain-target TARGET=<backend target>`
- `make test`
- `make lint`
- `make check`
- `make ci`
- `make release-check`
- `make build`
- `make clean`
- `make distclean`

## Artifact ownership and edit rules

- The normative core owns behavior. Change owner text first when behavior changes.
- `/contracts/*` contains repo-local derived contract artifacts. Hand-edit them only as owner-driven contract updates, and do not treat them as the behavioral owner.
- `/internal/gen/**` and `/packages/protocol-ts/src/generated/**` are generated outputs. Do not hand-edit either path.
- `pnpm-lock.yaml` and `go.sum` are tool-managed. Do not hand-edit either file.
- Keep codegen drift and migration drift separate.
- Codegen drift means generated outputs change after `make generate`.
- Migration drift means schema-affecting changes are missing from numbered migrations or migrations do not apply cleanly.

## Local execution procedure

- Start local backing services with `make db-up`.
- Then run `make dev`.
- The local bootstrap server uses `configs/dev/config.toml` through `CARTULARY_CONFIG_FILE`.
- `make help` prints the grouped root task surface without bootstrapping Node or pnpm.
- `make doctor` verifies required local tools and pinned toolchain versions without installing them.
- `make bootstrap` installs the pinned Go CLI tools and workspace dependencies.
- `make phase-ledgers` regenerates the committed phase coverage ledgers from `tools/phase*_test_map.json`.
- `make phase-ledger-drift` verifies committed phase coverage ledgers match the phase manifests without requiring Docker or service-backed tests.
- `make backend-store` runs the service-backed store-domain `U-*` backend slice that keeps unit-layer phase IDs while using real Postgres.
- `make target-plan`, `make target-plan-json`, and `make explain-target TARGET=<backend target>` inspect the backend Go target execution plan without running tests or starting services.
- `make test-fast` runs the pure backend unit slice, the service-backed backend store and integration slices, the backend process or E2E slice, frontend type-checking, and the frontend unit suite for the narrower local loop.
- `make test` is the authoritative full-corpus test surface. It runs pure backend unit, frontend type-check, and frontend unit work first, then shares one service-backed stage across backend service-backed work and `browser-e2e-webserver-backed`, and finally runs isolated browser suites. The Phase 0 process evidence under `cmd/server` is part of this surface and is not a direct-only command.
- `make check` is the developer verification gate. Toolchain drift and frontend install are early setup blockers; authored-frontend Biome, phase coverage ledger drift detection, generated-artifact drift detection, harness smoke checks, migration verification against a scratch local Postgres database, backend lint and tests, frontend type-check and tests, plus backend and frontend builds run in the parallel-safe block after setup passes. The gate then runs service-backed backend and shared-stack browser verification through one capacity-aware service-backed scheduler stage, and finally runs isolated browser suites.
- Apply authored frontend formatting with `make format`.
- `make ci` is the provider-neutral CI enforcement entrypoint. It composes the canonical repo task surface and fails on codegen drift, phase coverage ledger drift, migration failures, and deployable-shape drift.
- `make release-check` is the release verification gate. It runs the developer gate plus license report verification, SBOM verification, and build verification; until generators are chosen, the license and SBOM targets fail if their configured artifacts are missing or empty.
- `make clean` removes reproducible repo-local build and report artifacts while preserving checked-in files and external Go caches.
- `make distclean` additionally removes repo-local tool/runtime caches after printing the removal list.
- `make check` and `make ci` quiet output is the default. Use `VERBOSE=1` when you need the full streaming logs for investigation.
- If a `check-heavy` child fails under parallel execution, GNU Make may still print `Waiting for unfinished jobs....` while sibling jobs drain; that line is expected orchestration output, not a second root cause.
- Shell-backed verification wrappers such as frontend lint or migration drift report as non-test failures rather than unmapped test inventory when they exit non-zero.
- From PowerShell, prefer repo commands through `wsl.exe -d Ubuntu-24.04 --cd /home/askahn/code/cartulary ...`; for Node/pnpm, prepend `/home/askahn/code/cartulary/tmp/node-runtime/bin` to `PATH` and use `corepack pnpm`.
- If Git on the UNC WSL path reports dubious ownership, retry with `git -c safe.directory=//wsl.localhost/Ubuntu-24.04/home/askahn/code/cartulary ...`.
