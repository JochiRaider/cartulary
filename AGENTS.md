# Cartulary Repository Procedure

## Authority

- Product behavior is owned by the normative core: `docs/spec/00_document_set_status_and_precedence.md` through Core 04. Core 05 is only for claim-bearing timed, benchmark, fixture-sensitive, or publication evidence.
- `docs/testing-harness-nlspec.md` owns harness mechanics: command invocation, target selection, scheduling, fixture lifecycle, artifact emission, cleanup, and verification gates.
- `docs/domain.md` is the domain vocabulary and concept-boundary reference. Consult it before domain-facing changes, terminology-sensitive changes, view-schema work, record-model work, workbook-surface work, or docs that touch project vocabulary.
- `docs/design.md` owns frontend design-direction constraints and token definitions. It is not product-conformance evidence by itself, and design evidence must not be represented as Base Profile or extension-profile conformance.
- Guides under `docs/guides/` are implementation-support inputs unless an adopted owner document explicitly promotes a narrower rule.
- The canonical Go module path is `github.com/JochiRaider/cartulary`.

## Repository Boundaries

- `cmd/server`, `cmd/migrate`, and other `cmd/*` packages are binary composition roots only. Keep feature behavior out of `cmd/`.
- `internal/app` owns application assembly shared by binaries.
- `internal/platform/*` owns transport, runtime plumbing, configuration, storage adapters, auth primitives, and job shells.
- `internal/modules/*` owns domain and application logic inside the modular monolith.
- `contracts/*` is the repo-local derived contract layer, downstream of owner specs and upstream of generated code.
- `db/migrations` and `db/queries` are authored SQL inputs.
- `apps/web` is the top-level web app in the pnpm workspace.
- `packages/*` contains shared TypeScript packages.
- `scripts` and `tools` contain repo-local automation, manifests, schemas, and pinned helper tooling.
- `configs/dev` contains local development configuration inputs.
- `internal/testutil/*` contains reusable backend test harnesses and fixtures.

## Generated Files And Artifacts

- Do not hand-edit generated roots declared by `tools/generated_artifact_policy.json`: `internal/gen/**`, `packages/protocol-ts/src/generated/**`, and `packages/ui-contracts/src/generated/**`.
- Do not hand-edit generated harness/topology outputs such as `tools/task_surface.generated.mk` or generated outputs listed in `tools/execution_topology_manifest.json`.
- Generated phase ledgers and schedules are downstream of phase manifests and frontend phase maps. Update the owner inputs, then run the relevant Make generator or drift target.
- Do not hand-edit `go.sum`, `pnpm-lock.yaml`, or tool-managed dependency/install artifacts.
- For visual golden changes, follow `docs/guides/cartulary_visual_golden_maintenance.md`. Visual and accessibility artifacts remain design-direction or implementation-support evidence unless a Core 05 claim-publication boundary is separately satisfied.

## Commands

- Run repository commands from the repository root through public Make targets. Direct `go`, `pnpm`, Vitest, Playwright, Biome, and raw script commands are developer conveniences unless a Make-owned wrapper invokes them.
- Canonical repo-control pin values live in `tools/toolchain_pins.json`; mirrored toolchain text is checked by `make toolchain-drift`.
- Pinned bootstrap tools: `github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0`, `github.com/pressly/goose/v3/cmd/goose@v3.27.0`, `honnef.co/go/tools/cmd/staticcheck@v0.7.0`, `golang.org/x/vuln/cmd/govulncheck@v1.3.0`, `github.com/securego/gosec/v2/cmd/gosec@v2.26.1`, `github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@v1.10.0`, `github.com/anchore/syft/cmd/syft@v1.44.0`, ShellCheck `0.11.0`, and `github.com/testcontainers/testcontainers-go v0.42.0`.
- Use `make help` for the compact task surface and `make help-all` for the current public target inventory. Do not copy target lists into new docs when a pointer is sufficient.
- Use `make task-guide ROLE=<role> PHASE=phaseN` to choose narrow verification by role and phase. Useful roles include `local-dev`, `feature-dev`, `phase-author`, `ci-investigator`, and `release`.
- Use investigation targets such as `make explain-phase PHASE=phaseN`, `make explain-target TARGET=<target> DETAIL=summary|rows|artifacts`, `make explain-run RESULTS_DIR=<root|run-dir>`, `make target-plan`, and `make target-plan-json` before rerunning broad suites.
- Local setup and dev: `make doctor`, `make bootstrap`, `make db-up`, `make db-reset`, `make services-up`, `make object-store-init`, and `make dev`.
- Generation and drift: `make generate`, `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check`, `make toolchain-drift`, `make migration-drift`, `make phase-ledger-drift`, and `make phase-schedule-drift`.
- Common verification: `make test-fast`, `make test`, `make check`, `make lint`, `make frontend-typecheck`, `make frontend-unit`, `make frontend-import-boundary-check`, `make lint-biome`, `make lint-scripts`, `make lint-shell`, `make go-vulncheck`, `make go-gosec-targeted`, and `make go-gosec-audit`.
- Browser and frontend readiness targets include `make browser-e2e`, `make browser-e2e-webserver-backed`, `make browser-e2e-stateful`, `make browser-e2e-measurement`, `make browser-e2e-a11y`, `make browser-e2e-a11y-preflight`, and `make browser-e2e-visual`.
- Release/build targets include `make ci`, `make release-check`, `make harness-contract`, `make build`, `make build-server`, `make build-migrate`, `make build-operator`, and `make build-web`.
- Cleanup targets are destructive within repo-local outputs: `make clean` removes reproducible build/report artifacts, and `make distclean` also removes repo-local tool/runtime caches and dependency installs.

## Verification And Handoff

- Choose the narrowest target that covers the change, then broaden only when risk or ownership requires it. For phase work, prefer `make phase-slice PHASE=phaseN` and `make service-backed-slice PHASE=phaseN` after checking `make task-guide`.
- Run `make agent-finalize` before broader end-of-run verification. If using retained successful run evidence, pass `RESULTS_DIR=<successful full warm check run root>`; otherwise report that retained-run maintenance was skipped because `RESULTS_DIR` was unset.
- For docs-only changes, use documentation/harness validation targets such as `make generated-artifact-policy-check`, `make json-shape-check`, and command-surface inspection before considering broad `make check`.
- `make format` rewrites authored Go and frontend sources; do not run it solely for Markdown-only edits unless another touched file needs that formatter.
- When a command fails, report the failing target, the relevant summary artifact or run root when available, and whether the failure appears related to the change.
- Final reports should state the planning summary, files inspected or changed, substantive edits, verification commands and results, and any skipped checks with the reason.
