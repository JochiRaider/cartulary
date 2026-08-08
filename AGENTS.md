# Cartulary Repository Procedure

## Authority

- Adopted subsystem NLSpecs and normative Core owner sections define required behavior for their named scopes. When an adopted specification and a downstream artifact disagree, repair the projection or implementation.
- Typed limits, enums, schemas, mappings, algorithms, and fixtures under `contracts/<subsystem>/**` are versioned machine projections of their adopted owners. They contain executable facts but do not supersede those owners.
- Verification routing is owned by `contracts/verification/**`, `tools/test_catalog_owner.json`, and `tools/test_families/**`. Routing and execution evidence do not define requirements or prove specification completeness.
- Harness mechanics are specified by the adopted Testing Harness NLSpec and projected through machine-readable inputs under `tools/`, including the authored task surface, execution topology, schemas, and policy registries.
- Tests, generators, runtime metadata, conformance, and release evidence must not read, stat, hash, or otherwise depend on files under `docs/`, README, or other Markdown. Human review establishes that typed projections faithfully implement their adopted owners.
- `docs/domain.md` owns vocabulary and owner navigation within its stated boundary; `docs/design.md` supplies design direction within its stated boundary.
- The canonical Go module path is `github.com/JochiRaider/cartulary`.

## Repository Boundaries

- `cmd/server`, `cmd/migrate`, and other `cmd/*` packages are binary composition roots only. Keep feature behavior out of `cmd/`.
- `internal/app/server`, `internal/app/migrate`, and `internal/app/operator` are the exact application facades for their matching `cmd/*` roots; the `internal/app` root is not a Go package.
- `internal/app/revisionassembly` aggregates source-owner revision contributions and injects platform dependencies. Source owners construct their providers; Revisions validates the complete catalog and owns generic coordination.
- `internal/app/serverprocess` is retained process-level test evidence, not production assembly. Reusable application test composition belongs in `internal/testutil/appsupport`.
- Non-nil `server.Options.Postgres` and `server.Options.ObjectStore` values are borrowed. The server runtime closes only resources it creates, in reverse acquisition order, and `Runtime.Close` is idempotent.
- `internal/platform/*` owns transport, runtime plumbing, configuration, storage adapters, auth primitives, and job shells.
- `internal/modules/*` owns domain, application, and owner-specific supporting-subsystem logic inside the modular monolith. `internal/modules/database_migrations` owns migration lifecycle mechanics and receives already-opened database handles; it does not own PostgreSQL connectivity, secrets, generic query/transaction ports, source-owner schema meaning, or authored SQL.
- `contracts/*` contains versioned machine projections and production contracts downstream of adopted specifications and upstream of generated code.
- `db/migrations` and `db/queries` are authored SQL inputs.
- `apps/web` is the top-level web app in the pnpm workspace.
- `packages/*` contains shared TypeScript packages.
- `scripts` and `tools` contain repo-local automation, manifests, schemas, and pinned helper tooling.
- `configs/dev` contains local development configuration inputs.
- `internal/testutil/*` contains reusable backend test harnesses and fixtures.

## Generated Files And Artifacts

- Do not hand-edit generated roots declared by `tools/generated_artifact_policy.json`: `internal/gen/**`, `packages/protocol-ts/src/generated/**`, and `packages/ui-contracts/src/generated/**`.
- Do not hand-edit generated harness/topology outputs such as `tools/task_surface.generated.mk` or generated outputs listed in `tools/execution_topology_manifest.json`.
- Generated topology and schedules are downstream of the owner catalog and authored execution topology. Update owner inputs, then run the relevant Make generator or drift target.
- Do not hand-edit `go.sum`, `pnpm-lock.yaml`, or tool-managed dependency/install artifacts.
- For visual golden changes, follow `docs/guides/cartulary_visual_golden_maintenance.md`. Visual and accessibility artifacts remain design-direction or implementation-support evidence unless a Core 05 claim-publication boundary is separately satisfied.

## Commands

- Run repository commands from the repository root through public Make targets. Direct `go`, `pnpm`, Vitest, Playwright, Biome, and raw script commands are developer conveniences unless a Make-owned wrapper invokes them.
- Canonical repo-control pin values live in `tools/toolchain_pins.json`; `make toolchain-drift` checks production and bootstrap inputs, not prose mirrors.
- Pinned bootstrap tools: `github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0`, `github.com/pressly/goose/v3/cmd/goose@v3.27.0`, `honnef.co/go/tools/cmd/staticcheck@v0.7.0`, `golang.org/x/vuln/cmd/govulncheck@v1.3.0`, `github.com/securego/gosec/v2/cmd/gosec@v2.26.1`, `github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@v1.10.0`, `github.com/anchore/syft/cmd/syft@v1.44.0`, ShellCheck `0.11.0`, and `github.com/testcontainers/testcontainers-go v0.42.0`.
- Use `make help` for the compact task surface and `make help-all` for the current public target inventory. Do not copy target lists into new docs when a pointer is sufficient.
- Use `make task-guide ROLE=module-author OWNER=<owner-id>` to choose narrow owner verification. Use `make help` and `make help-all` for the current role and target inventory.
- Use investigation targets such as `make explain-test-owner OWNER=<owner-id>`, `make explain-target TARGET=<target> DETAIL=summary|rows|artifacts`, `make explain-run RESULTS_DIR=<root|run-dir>`, `make target-plan`, and `make target-plan-json` before rerunning broad suites.
- Local setup and dev: `make doctor`, `make bootstrap`, `make db-up`, `make db-reset`, `make services-up`, `make object-store-init`, and `make dev`.
- Generation and drift: `make generate`, `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check`, `make toolchain-drift`, and `make migration-drift`.
- Common verification: `make test-fast`, `make test`, `make check`, `make lint`, `make frontend-typecheck`, `make frontend-unit`, `make frontend-import-boundary-check`, `make lint-biome`, `make lint-scripts`, `make lint-markdown`, `make lint-shell`, `make go-vulncheck`, `make go-gosec-targeted`, and `make go-gosec-audit`.
- Browser and frontend readiness targets include `make browser-e2e`, `make browser-e2e-webserver-backed`, `make browser-e2e-stateful`, `make browser-e2e-measurement`, `make browser-e2e-a11y`, and `make browser-e2e-visual`.
- Release/build targets include `make ci`, `make release-check`, `make harness-contract`, `make build`, `make build-server`, `make build-migrate`, `make build-operator`, and `make build-web`.
- Cleanup targets are destructive within repo-local outputs: `make clean` removes reproducible build/report artifacts, and `make distclean` also removes repo-local tool/runtime caches and dependency installs.

## Verification And Handoff

- Choose the narrowest target that covers the change, then broaden only when risk or ownership requires it. Prefer `make test-slice OWNER=<owner-id> [ROWS=<row-id,...>]` and `make service-backed-test-slice OWNER=<owner-id> [ROWS=<row-id,...>]` after checking `make task-guide ROLE=module-author OWNER=<owner-id>`.
- Run `make agent-finalize` before broader end-of-run verification. If using retained successful run evidence, pass `RESULTS_DIR=<successful full warm check run root>`; otherwise report that retained-run maintenance was skipped because `RESULTS_DIR` was unset.
- For docs-only changes, use `make lint-markdown` when documentation maintenance is desired. Documentation-only edits must not change product checks, generated artifacts, conformance, or release evidence.
- `make format` rewrites authored Go and frontend sources; do not run it solely for Markdown-only edits unless another touched file needs that formatter.
- When a command fails, report the failing target, the relevant summary artifact or run root when available, and whether the failure appears related to the change.
- Final reports should state the planning summary, files inspected or changed, substantive edits, verification commands and results, and any skipped checks with the reason.
