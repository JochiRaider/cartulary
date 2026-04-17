# Cartulary Repository Procedure

## Authority and placeholders

- Normative behavior is owned by the Cartulary normative core under `docs/spec/00_document_set_status_and_precedence.md` through Core 04. The guides under `docs/guides/` are implementation-support inputs, not independent behavior owners.
- repository remote: `github.com/JochiRaider/cartulary`
- TODO: replace the temporary Go module path `example.com/todo/cartulary`.
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

- `make bootstrap`
- `make db-up`
- `make db-reset`
- `make dev`
- `make generate`
- `make test`
- `make lint`
- `make check`
- `make build`

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
- `make bootstrap` installs the pinned Go CLI tools and workspace dependencies.
- `make check` is the developer verification gate and runs frozen frontend install, generated-artifact drift detection, migration verification against a scratch local Postgres database, backend lint and tests, frontend lint, type-check, and tests, plus backend and frontend builds.
