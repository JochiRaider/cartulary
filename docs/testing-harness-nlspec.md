# Testing Harness NLSpec

## 1. Status, Scope, and Authority

Status: `draft/proposed`.

This NLSpec defines the Cartulary testing harness subsystem. It is not adopted evidence of product conformance until the repository authority process adopts it.

**TH-HARNESS-REQ-001**
This NLSpec owns only harness mechanics: command invocation, target selection, scheduling, fixture lifecycle, service ownership, artifact emission, summary emission, cleanup, and harness verification gates. It MUST NOT define product behavior owned by Core 00 through Core 04. It MUST NOT define claim-publication or benchmark-publication behavior owned by Core 05.
Verified by: TH-HARNESS-AC-013, TH-HARNESS-AC-016

**TH-HARNESS-REQ-002**
A harness conformance claim MUST identify this NLSpec version, the exact public Make target or target set under evaluation, the conformance environment from Section 14, and the retained result root/run ID/run root when retained harness artifacts are used as evidence.
Verified by: TH-HARNESS-AC-015, TH-HARNESS-AC-016

**TH-HARNESS-REQ-003**
The canonical public command surface is Make. A public command is canonical only when invoked as `make <target>` from the repository root or through a Make-owned wrapper that preserves the target identity.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-004, TH-HARNESS-AC-005

**TH-HARNESS-REQ-004**
Generated files under `internal/gen/**`, `packages/protocol-ts/src/generated/**`, generated task/schedule artifacts, and generated Make includes are downstream generated artifacts. They MUST NOT be hand-edited and MUST NOT become behavior owners unless a later adopted NLSpec explicitly promotes one of them.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-016

**TH-HARNESS-REQ-005**
Direct package scripts, raw scripts, raw Go/Vitest/Playwright/Biome/Vite/pnpm commands, and tool-specific reports are developer conveniences or child commands unless a public Make target invokes them. Direct invocation of those surfaces MUST NOT be treated as equivalent to a canonical harness run.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-005

## 2. Purpose, Non-Goals, and Conformance Boundary

The testing harness exists to provide a reproducible repository command surface for local developers, CI entrypoints, coding agents, and release verification. It provides deterministic target selection, bounded output, structured artifacts, explicit service ownership, controlled fixture lifecycle, stable failure classification, and destructive cleanup gates.

**TH-HARNESS-REQ-006**
The harness MUST provide all of the following for public Make targets:

- deterministic target classification and target selection;
- declared configuration resolution;
- output-mode behavior;
- exit-code mapping;
- retained artifact identity when a target declares artifacts;
- failure classification that separates product assertion failures from harness operational failures;
- cleanup predicates for every destructive operation.
Verified by: TH-HARNESS-AC-001..TH-HARNESS-AC-016

**TH-HARNESS-REQ-007**
The harness MUST NOT claim provider-specific hosted CI behavior, benchmark publication, release publication readiness, visual-snapshot refresh authority, macOS support, Windows-native support, Podman support, or Playwright artifact schema stability unless those areas are explicitly included in this NLSpec's current conformance profile or in a later adopted NLSpec.
Verified by: TH-HARNESS-AC-012, TH-HARNESS-AC-016

**TH-HARNESS-REQ-008**
Logical scheduler resources are execution constraints inside the harness. They MUST NOT be represented as guarantees about physical CPU, I/O, Docker, database, object-store, browser, or network capacity.
Verified by: TH-HARNESS-AC-006

## 3. Terminology

| Term                     | Meaning                                                                                                                                                    |
| ------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| harness run              | One invocation of a canonical harness command or one child invocation explicitly tied to a result root and run ID.                                         |
| target                   | A named Make target or scheduler target selected by the harness.                                                                                           |
| public target            | A target classified as public in the public command registry and canonical only through Make.                                                              |
| child target             | A target invoked by an aggregate, sequence, scheduler, or wrapper target.                                                                                  |
| work unit                | One scheduler-visible executable unit with an identity, dependencies, resource claims, logs, status, and optional completion keys.                         |
| scheduler                | A harness runner that executes a manifest-defined DAG using logical resource claims and emits scheduler events and summaries.                              |
| result root              | The root directory that contains run artifacts. The default is `.cartulary/test-results`.                                                                  |
| run ID                   | The run directory name under the result root. The default format is defined in Section 6.                                                                  |
| run root                 | The directory `normalize_result_root(CARTULARY_TEST_RESULTS_DIR) / normalize_run_id(CARTULARY_TEST_RUN_ID)`.                                               |
| harness artifact         | A file or directory produced by a harness run, child command, service, scheduler, test runner, or diagnostic tool.                                         |
| retained artifact        | An artifact preserved after command exit for a specific result root, run ID, and target.                                                                   |
| generated artifact       | A file produced from owner inputs by a generator and checked for drift.                                                                                    |
| fixture                  | Test setup state created for a test, package, target, scheduler group, browser stack, or service suite.                                                    |
| service-backed fixture   | A fixture that uses Postgres, MinIO, Docker/testcontainers, browser processes, or Compose-backed services.                                                 |
| backing services         | Postgres, MinIO, Docker/testcontainers, Compose services, backend processes, frontend processes, and browser runtime dependencies used by harness targets. |
| output mode              | The resolved mode from Section 7 that controls stdout, stderr, and artifact summary behavior.                                                              |
| machine output           | The `machine` output mode defined in Section 7. For public Make targets that accept it, stdout is exactly one UTF-8 JSON object followed by LF.            |
| failure class            | A coarse normalized grouping for failed harness commands: `product`, `config`, `infra`, `harness`, `artifact`, `timing`, `interrupted`, or `unknown`.      |
| failure reason           | A detailed snake-case reason code used for diagnostics, exit-code mapping, automation, and handoff.                                                        |
| cleanup tier             | A named cleanup scope such as repo-local clean, repo-local distclean, service-suite cleanup, browser-stack cleanup, or stale janitor cleanup.              |
| stale janitor            | A cleanup routine that removes previously generated DBs, buckets, containers, or browser fixtures only when proof predicates match.                        |
| diagnostic-only artifact | An artifact retained for human investigation whose internal shape is not a machine-readable harness conformance contract.                                  |

Domain and product terms keep their meanings from the product specs and `docs/domain.md`.

## 4. Public Command Surface

**TH-HARNESS-REQ-050**
The public command registry MUST be closed over public targets in `tools/task_surface_manifest.json` with `classification="public"`. The implementation MUST provide exactly the public targets listed in the target registry below unless the manifest and this NLSpec are revised together.
Verified by: TH-HARNESS-AC-001

**TH-HARNESS-REQ-051**
Every public target MUST declare exactly one output class, exactly one stable-summary schema policy, and exactly one artifact policy. The output-class behavior is owned by Section 7. The schema policy is owned by Section 8.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-004, TH-HARNESS-AC-005

**TH-HARNESS-REQ-052**
A Make-owned wrapper MAY invoke package scripts, raw scripts, or external tools as implementation mechanisms. The wrapper remains responsible for the public target's configuration, output, artifact, failure, exit-code, and cleanup contract.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-004

### 4.1 Mechanism Boundary

| Surface                                                  |                                  Normative? | Required contract                                                                      |
| -------------------------------------------------------- | ------------------------------------------: | -------------------------------------------------------------------------------------- |
| Public Make target name                                  |                                         yes | Stable command surface invoked as `make <target>` from the repository root.            |
| `tools/task_surface_manifest.json` public classification |                                         yes | Source input for public target registry completeness.                                  |
| Root/package `pnpm` scripts                              |                                          no | Developer convenience unless invoked by a Make-owned public target.                    |
| Raw `scripts/*.mjs` and `scripts/*.sh`                   | no, except wrapper-owned observable effects | May change when public Make behavior remains unchanged.                                |
| `tools/testservices` binary path                         |                                          no | Service lifecycle behavior is normative; binary path is an implementation realization. |
| Public output classes and schema IDs listed in Section 8 |                                         yes | Required machine-output and artifact validation contracts.                              |
| Docker image tag for Postgres or MinIO                   | no unless declared in a service fixture row | Exact tag is not normative unless it defines fixture semantics in Section 11.          |
| Generated Make include names, helper binaries, helper classifications, priority-band names, and generator constants | no | Implementation detail unless promoted by an explicit requirement.                      |

### 4.2 Command Family Defaults

| Family                          | Commands                                                                                                                                                                                                                                               | Required inputs                                                                                              | Optional inputs and defaults                                                                  | Output class family                                     | Scheduler use                                                   | Backing services                                                              | Artifact behavior                                                                   | Failure contract                                                                         |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------- | ------------------------------------------------------- | --------------------------------------------------------------- | ----------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| help and discovery              | `help`, `help-all`, `task-surface-report`, `task-guide`, `target-plan`, `target-plan-json`, `fixture-report`, `explain-run`, `explain-phase`, `explain-target`                                                                                         | Command-specific Make variables such as `ROLE`, `PHASE`, `TARGET`, `RESULTS_DIR`, `RUN_ID`, `DETAIL`, `JSON` | Omitted optional inputs select documented summary views.                                      | `human_summary` or `machine_stdout_json` where declared | None                                                            | None                                                                          | Does not create central run evidence unless a target row declares a summary schema. | Usage/config errors use Section 9.                                                       |
| bootstrap and toolchain         | `doctor`, `bootstrap`, `bootstrap-node-runtime`, `frontend-toolchain`, `frontend-install`, `playwright-install`                                                                                                                                        | Required local tools according to target                                                                     | Tool paths default from Section 5.                                                            | `summary_with_artifacts`                                | None                                                            | May download/install repo-local tools.                                        | Tool-run summary required.                                                          | Tool/config failures are `configuration_error` or `preflight_error`.                     |
| local services and dev          | `db-up`, `db-reset`, `services-up`, `minio-init`, `dev`                                                                                                                                                                                                | Docker Compose and local config where required                                                               | `CONFIG_FILE=configs/dev/config.toml`, `MINIO_BUCKET=cartulary` where target reads them.      | `service_summary_with_artifacts` or `interactive_raw`   | None                                                            | Compose Postgres/MinIO and local processes.                                   | Service summaries where declared; `dev` has no verification artifact contract.      | Startup/readiness/config failures are harness operational failures.                      |
| generated and drift             | `generate`, `generate-drift`, `generated-artifact-policy-check`, `json-shape-check`, `toolchain-drift`, `migration-drift`, `phase-ledgers`, `phase-ledger-drift`, `phase-schedules`, `phase-schedule-drift`, `agent-finalize`, `benchmark-claim-check` | Owner inputs and manifests                                                                                   | `RESULTS_DIR` only where target reads retained evidence.                                      | `summary_with_artifacts`                                | Only where child target declares scheduling.                    | Migration drift may use scratch Postgres when scheduled.                      | Tool summary and command-specific drift files.                                      | Drift mismatch is `artifact_error` or `scheduler_accounting_error`.                      |
| phase and service slices        | `phase-slice`, `service-backed-slice`                                                                                                                                                                                                                  | `PHASE=phaseN` where required                                                                                | `JSON` where the target declares JSON planning output.                                        | `scheduler_summary_with_artifacts`                      | Uses phase selection and service-backed scheduler.              | Service-backed slice requires backing services when phase work does.          | Target, run, scheduler, and phase artifacts.                                        | Missing/invalid phase is `usage_error`; child failures retain child class.               |
| backend and frontend leaf tests | `backend-unit`, `backend-store`, `backend-integration`, `backend-process`, `frontend-typecheck`, `frontend-unit`, `frontend-import-boundary-check`                                                                                                     | Toolchain, manifests, package inputs                                                                         | Parallelism and worker variables from Section 5.                                              | `summary_with_artifacts`                                | Service-backed targets may use a scheduler or testservices.     | Store/integration/process targets require Postgres/MinIO when service-backed. | Phase, target, tool, logs, reports.                                                 | Product assertion failures are `test_assertion_failure`; setup failures are operational. |
| browser E2E                     | `browser-e2e`, `browser-e2e-webserver-backed`, `browser-e2e-stateful`, `browser-e2e-measurement`, `browser-e2e-visual`                                                                                                                                 | Node/pnpm, Playwright browser, backend/migrate/server support, services                                      | `PLAYWRIGHT_WORKERS=3`, `BROWSER_E2E_FUNCTIONAL_SHARDS=auto` unless overridden by Section 5.  | `summary_with_artifacts`                                | Uses browser batch and service-backed scheduler where declared. | Postgres, MinIO, backend, frontend, browser runtime.                          | Browser stack, Playwright, reset, target, scheduler artifacts.                      | Product assertions are product failures; stack/readiness/reset failures are operational. |
| aggregates and gates            | `test-fast`, `test`, `lint`, `check`, `ci`, `release-check`, `build`                                                                                                                                                                                   | Toolchain and child inputs                                                                                   | `summary` output mode by default; `ci` may default to `ci` mode under `scripts/ci/verify.sh`. | aggregate or scheduler output classes                   | `check` uses the check scheduler.                               | Service-backed and browser children require backing services.                 | Aggregate run, child target, scheduler, and tool summaries.                         | Exit nonzero if any required child fails or artifact validation fails.                   |
| static analysis and security    | `lint-biome`, `lint-scripts`, `lint-shell`, `go-vulncheck`, `go-gosec-targeted`, `go-gosec-audit`                                                                                                                                                      | Toolchain and source roots                                                                                   | Rule variables from Make; `LINT_SHELL_STRICT=1` makes shell lint blocking.                    | `summary_with_artifacts`                                | Scheduled inside `check` where declared.                        | None                                                                          | Tool summary and logs.                                                              | Findings are gate failures unless target definition makes them warning-only.             |
| builds                          | `build-server`, `build-migrate`, `build-web`                                                                                                                                                                                                           | Build inputs and toolchain                                                                                   | Output paths from Make variables.                                                             | `summary_with_artifacts`                                | Scheduled as readiness work where declared.                     | None                                                                          | Tool summary and build logs.                                                        | Build failures are gate failures.                                                        |
| cleanup                         | `clean`, `distclean`                                                                                                                                                                                                                                   | None                                                                                                         | Uses Make path registries.                                                                    | `destructive_human`                                     | None                                                            | Does not stop Docker Compose services.                                        | No central summary contract.                                                        | Unsafe path guard failure exits nonzero; missing paths are not failures.                 |
| formatting                      | `format`                                                                                                                                                                                                                                               | Toolchain                                                                                                    | None                                                                                          | `summary_with_artifacts`                                | None                                                            | None                                                                          | Tool summary and formatter logs.                                                    | Formatter failure is operational; formatter rewrites are mutating.                       |

### 4.3 Public Target Registry

Every command below inherits the matching family defaults. `Included in` is the manifest-declared inclusion set.

| Target                                               | Included in       | Output class                     | Stable summary schema                              | Notes                                                                       |
| ---------------------------------------------------- | ----------------- | -------------------------------- | -------------------------------------------------- | --------------------------------------------------------------------------- |
| `help`                                               | helper_only       | human_summary                    | none                                               | Compact public command help.                                                |
| `help-all`                                           | helper_only       | human_summary                    | none                                               | Exhaustive public help.                                                     |
| `doctor`                                             | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Tool readiness check.                                                       |
| `bootstrap`                                          | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Installs pinned tooling and dependencies.                                   |
| `bootstrap-node-runtime`                             | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Installs repo-local Node runtime.                                           |
| `frontend-toolchain`                                 | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Prepares Node/pnpm toolchain.                                               |
| `frontend-install`                                   | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Installs workspace dependencies.                                            |
| `playwright-install`                                 | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Installs Playwright browser runtime.                                        |
| `db-up`                                              | helper_only       | service_summary_with_artifacts   | `cartulary.tool_run_summary.v2`                    | Starts local Postgres.                                                      |
| `db-reset`                                           | helper_only       | service_summary_with_artifacts   | `cartulary.tool_run_summary.v2`                    | Resets local database only; object storage is not reset by this target.     |
| `services-up`                                        | helper_only       | service_summary_with_artifacts   | `cartulary.tool_run_summary.v2`                    | Starts local backing services.                                              |
| `minio-init`                                         | helper_only       | service_summary_with_artifacts   | `cartulary.tool_run_summary.v2`                    | Initializes local MinIO bucket.                                             |
| `dev`                                                | helper_only       | interactive_raw                  | none                                               | Local interactive dev stack.                                                |
| `generate`                                           | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Runs generation commands.                                                   |
| `generate-drift`                                     | check             | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Checks generated drift.                                                     |
| `generated-artifact-policy-check`                    | check             | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Checks generated artifact policy.                                           |
| `json-shape-check`                                   | check             | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Checks JSON manifest shapes.                                                |
| `toolchain-drift`                                    | check             | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Checks toolchain pins.                                                      |
| `migration-drift`                                    | check             | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Checks migrations.                                                          |
| `phase-ledgers`                                      | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Regenerates phase ledgers.                                                  |
| `phase-ledger-drift`                                 | check             | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Checks phase ledger drift.                                                  |
| `phase-schedules`                                    | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Regenerates phase schedules.                                                |
| `phase-schedule-drift`                               | check             | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Checks schedule drift.                                                      |
| `agent-finalize`                                     | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | End-of-run maintenance surface.                                             |
| `benchmark-claim-check`                              | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Checks benchmark claim artifacts.                                           |
| `task-surface-report`                                | helper_only       | human_summary                    | none                                               | Prints target surface report.                                               |
| `task-guide`                                         | helper_only       | human_summary                    | none                                               | Prints role/phase guidance.                                                 |
| `phase-slice`                                        | helper_only       | scheduler_summary_with_artifacts | `cartulary.tool_run_summary.v2`                    | Runs selected phase target slice.                                           |
| `service-backed-slice`                               | helper_only       | scheduler_summary_with_artifacts | `cartulary.tool_run_summary.v2`                    | Runs selected phase service-backed slice or explicit no-op.                 |
| `backend-unit`                                       | test,check,ci     | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Pure backend unit evidence.                                                 |
| `backend-store`                                      | test,check,ci     | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Service-backed store slice.                                                 |
| `backend-integration`                                | test,check,ci     | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Service-backed integration slice.                                           |
| `backend-process`                                    | test,check,ci     | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Service-backed process slice.                                               |
| `target-plan`                                        | helper_only       | human_summary                    | none                                               | Prints Go target plan.                                                      |
| `target-plan-json`                                   | helper_only       | machine_stdout_json              | command-specific JSON                              | Prints Go target plan JSON.                                                 |
| `fixture-report`                                     | helper_only       | human_summary                    | `cartulary.fixture_report.v1` where JSON requested | Reads retained fixture artifacts.                                           |
| `explain-run`                                        | helper_only       | human_summary                    | command-specific JSON where JSON requested         | Reads retained run artifacts.                                               |
| `explain-phase`                                      | helper_only       | human_summary                    | none                                               | Explains phase manifest.                                                    |
| `explain-target`                                     | helper_only       | human_summary                    | none                                               | Explains target plan/artifacts/logs.                                        |
| `go-test-duration-baselines`                         | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Refreshes Go duration baselines from a selected run.                        |
| `go-test-duration-baseline-coverage`                 | check             | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Checks baseline coverage.                                                   |
| `go-test-duration-baseline-drift`                    | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Checks Go duration baseline drift against selected run.                     |
| `browser-e2e-duration-baselines`                     | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Refreshes browser duration baselines.                                       |
| `browser-e2e-duration-baseline-drift`                | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Checks browser duration drift.                                              |
| `service-backed-make-target-duration-baselines`      | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Refreshes service-backed target durations.                                  |
| `service-backed-make-target-duration-baseline-drift` | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Checks service-backed target duration drift.                                |
| `harness-smoke-duration-baselines`                   | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Refreshes harness smoke durations.                                          |
| `harness-smoke-duration-baseline-drift`              | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Checks harness smoke duration drift.                                        |
| `scheduler-event-order-drift`                        | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Checks scheduler event ordering.                                            |
| `scheduler-summary-timing-drift`                     | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Checks scheduler summary timing.                                            |
| `frontend-typecheck`                                 | test,check,ci     | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Frontend TypeScript check.                                                  |
| `frontend-unit`                                      | test,check,ci     | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Frontend unit suite.                                                        |
| `frontend-import-boundary-check`                     | check             | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Frontend import boundary check.                                             |
| `lint-biome`                                         | check             | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Frontend authored-source lint.                                              |
| `lint-scripts`                                       | check             | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Script lint.                                                                |
| `lint-shell`                                         | check             | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | ShellCheck wrapper; direct target is warning-only unless strict env is set. |
| `format`                                             | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Mutating formatter target.                                                  |
| `browser-e2e`                                        | test,check,ci     | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Isolated browser E2E aggregate.                                             |
| `browser-e2e-webserver-backed`                       | test,check,ci     | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Shared-stack browser stage.                                                 |
| `browser-e2e-stateful`                               | test,check,ci     | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Stateful browser stage.                                                     |
| `browser-e2e-measurement`                            | test,check,ci     | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Measurement browser stage.                                                  |
| `browser-e2e-visual`                                 | test,check,ci     | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Visual validation stage.                                                    |
| `test-fast`                                          | test,check,ci     | aggregate_summary_with_artifacts | `cartulary.tool_run_summary.v2`                    | Fast local verification aggregate.                                          |
| `test`                                               | test,check,ci     | aggregate_summary_with_artifacts | `cartulary.tool_run_summary.v2`                    | Full-corpus test aggregate.                                                 |
| `lint`                                               | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Aggregate lint wrapper.                                                     |
| `go-vulncheck`                                       | check             | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Govulncheck wrapper.                                                        |
| `go-gosec-targeted`                                  | check             | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Blocking targeted Gosec wrapper.                                            |
| `go-gosec-audit`                                     | check             | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Warning-only audit profile wrapper.                                         |
| `check`                                              | check             | scheduler_summary_with_artifacts | `cartulary.tool_run_summary.v2`                    | Developer gate through check scheduler.                                     |
| `ci`                                                 | ci                | aggregate_summary_with_artifacts | `cartulary.tool_run_summary.v2`                    | Provider-neutral CI entrypoint.                                             |
| `release-check`                                      | release-check     | aggregate_summary_with_artifacts | `cartulary.tool_run_summary.v2`                    | Release gate aggregate.                                                     |
| `build`                                              | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Build aggregate.                                                            |
| `build-server`                                       | check,helper_only | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Server build.                                                               |
| `build-migrate`                                      | check,helper_only | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Migration binary build.                                                     |
| `build-web`                                          | helper_only       | summary_with_artifacts           | `cartulary.tool_run_summary.v2`                    | Web build.                                                                  |
| `clean`                                              | helper_only       | destructive_human                | none                                               | Repo-local cleanup.                                                         |
| `distclean`                                          | helper_only       | destructive_human                | none                                               | Repo-local cleanup plus repo-local tool/runtime caches.                     |

### 4.4 Direct Script and Package Boundary

| Surface                                                  | Classification                           | Contract                                                                                                  |
| -------------------------------------------------------- | ---------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| Root `package.json` scripts `build`, `test`, `typecheck` | Developer convenience                    | They do not promise Make result roots, run IDs, scheduler summaries, cleanup, or machine output.          |
| `apps/web/package.json` scripts                          | Developer convenience or child command   | Browser and unit scripts become harness child work only when invoked through Make wrappers.               |
| Raw `scripts/run-*.sh` and `scripts/*.mjs`               | Tool-owned diagnostics or child commands | Their direct usage and exit codes are not public harness contracts unless a Make target adopts them.      |
| Raw Go, Vitest, Playwright, Biome, Vite, pnpm commands   | Tool-owned                               | Tool output schemas remain external or diagnostic unless consumed and normalized by a Make-owned wrapper. |

## 5. Configuration Resolution Contract

**TH-HARNESS-REQ-100**
Every public Make target MUST resolve harness configuration through `resolve_harness_config()` before child work begins. A target that cannot resolve or validate configuration MUST fail with `failure_class=config`, `failure_reason=configuration_error`, and public exit code `2`.
Verified by: TH-HARNESS-AC-002, TH-HARNESS-AC-003, TH-HARNESS-AC-014

`resolve_harness_config()` is the normative configuration-resolution contract. Repository implementation entrypoints such as preflight helpers MAY wrap this resolver, but MUST NOT define a narrower public-target configuration contract.

**TH-HARNESS-REQ-101**
Generated manifests are execution inputs, not caller configuration. A caller-supplied variable that attempts to override a non-overridable manifest field MUST fail with `configuration_error` before child work.
Verified by: TH-HARNESS-AC-002

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
  declared = global_configuration_table + configuration_table[target]
  resolved = empty map

  reject undeclared wrapper CLI flags

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
  emit resolved values required by Section 8 summaries
```

### 5.3 Empty-String Rules

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

### 5.4 Configuration Variable Table

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
| `GO`, `GO_CACHE_DIR`, `GO_MOD_CACHE_DIR`, `GOCACHE`, `GOMODCACHE`                               | toolchain               | executable path or filesystem path                                                                                    | Go auto-discovered; external caches `/tmp/cartulary-go-build` and `/tmp/cartulary-go-mod` | Make variable, env, default                     | invalid for paths, omitted for executable               | path normalization by target helper                                                                           | `configuration_error`, exit `2`                                                    | executable path and cache path unless redacted     |
| `NODE_VERSION`, `PNPM_VERSION`, `NODE_RUNTIME_DIR`, `NODE_BIN`, `PNPM`, `COREPACK_HOME`, `PATH` | toolchain               | version token or filesystem path                                                                                      | Node `24.15.0`, pnpm `10.33.0`, repo-local `tmp/node-runtime`                             | Make variable, env, default                     | invalid for paths/versions unless row-specific optional | exact version token; path normalization                                                                       | `configuration_error`, exit `2`                                                    | version and runtime path                           |
| `CONFIG_FILE`, `CARTULARY_CONFIG_FILE`                                                          | app runtime             | config file path                                                                                                      | `configs/dev/config.toml` for local/dev/browser targets                                   | Make variable, env, config binding, default     | omitted                                                 | path normalization; `CARTULARY_CONFIG_FILE` wins only inside application runtime when both are passed through | harness invalid path: `configuration_error`; app invalid config: target failure    | path, not file contents                            |
| `TEST_SERVICES_BIN`, `CARTULARY_TEST_SERVICES_BIN`                                              | service suite           | executable path                                                                                                       | `tmp/toolbin/cartulary-test-services`                                                     | Make variable, env, default                     | invalid                                                 | path normalization                                                                                            | `configuration_error`, exit `2`                                                    | normalized path                                    |
| `CARTULARY_TEST_SERVICES_ACTIVE`                                                                | service suite           | exact `1` selects attach mode                                                                                         | owned mode                                                                                | env, Make variable, default                     | false                                                   | exact string compare                                                                                          | non-`1` selects owned mode                                                         | mode only                                          |
| `CARTULARY_TEST_SUITE_ID`, `CARTULARY_TEST_TARGET`                                              | service suite           | non-empty ASCII token; suite ID is 24 lowercase hex in owned mode                                                     | generated in owned mode                                                                   | service manifest, env in attach mode            | invalid in attach mode                                  | exact grammar validation                                                                                      | `configuration_error`, exit `2`                                                    | suite ID, target                                   |
| Postgres attach set                                                                             | service suite           | `CARTULARY_PGTEST_ADMIN_DSN`, `CARTULARY_PGTEST_DSN_TEMPLATE` containing `{database}`, `CARTULARY_PGTEST_TEMPLATE_DB` | none                                                                                      | env, Make variable                              | invalid                                                 | DSN redacted; template exact placeholder validation                                                           | partial or malformed set: `configuration_error`, exit `2`                          | redacted DSN and attach mode                       |
| Postgres fixture policy                                                                         | service suite           | `template_clone`, `transaction`, `package_reset`; table lists use comma-separated SQL identifiers                     | default policy `template_clone`                                                           | Make variable, env, manifest, default           | omitted                                                 | lower-case exact token; table identifiers sorted for summaries                                                | unknown policy or bad identifier: `configuration_error`, exit `2`                  | policy and table count                             |
| MinIO attach set                                                                                | service suite           | endpoint, access key, secret key, secure bool through `CARTULARY_S3TEST_*`                                            | none                                                                                      | env, Make variable                              | invalid for required members                            | endpoint normalized; credentials redacted; secure bool exact `true`/`false` or `1`/`0`                        | partial set or invalid bool: `configuration_error`, exit `2`                       | endpoint, secure flag, credential redaction marker |
| `CARTULARY_TEST_SERVICES_WEB_E2E_CLEANUP_WORKERS`                                               | browser/service cleanup | integer `1..16`                                                                                                       | `4`                                                                                       | Make variable, env, default                     | omitted                                                 | decimal parse                                                                                                 | invalid value falls back to `4` and records warning                                | resolved value and warning when fallback used      |
| Compose env                                                                                     | local services          | `CARTULARY_COMPOSE_FILE`, ready timeouts, `MINIO_BUCKET`                                                              | `docker-compose.dev.yml`, Postgres `180s`, MinIO `120s`, bucket `cartulary`               | Make variable, env, default                     | omitted for optional values                             | path and duration normalization                                                                               | missing Docker/Compose: Section 9 class                                            | non-secret values                                  |
| Browser owned-stack env                                                                         | browser                 | runtime roots, origins, backend/frontend port overrides                                                               | dynamic ports; public origin `http://127.0.0.1:4173` when not overridden                  | Make variable, env, manifest, default           | invalid for required values                             | origins lower-case scheme/host where applicable; ports decimal                                                | config or port collision: `resource_conflict` or `configuration_error`             | origins, ports, runtime root                       |
| `PLAYWRIGHT_WORKERS`, worker index/offset envs                                                  | browser                 | positive integers; worker offset `0..1024`                                                                            | Make `3`; shared config fallback `2`; offset `0`                                          | Make variable, env, default                     | omitted                                                 | decimal parse                                                                                                 | `configuration_error`, exit `2`                                                    | worker counts                                      |
| Webserver-backed shard env                                                                      | browser                 | required grep/file values declared by target                                                                          | none                                                                                      | Make variable, manifest                         | invalid                                                 | exact string after JSON/shell decoding                                                                        | missing required value: `configuration_error`, exit `2`                            | declared shard IDs only                            |
| `CARTULARY_ENABLE_TEST_ROUTES`                                                                  | reset/browser           | exact `1` enables test routes                                                                                         | disabled                                                                                  | Make variable, env, default                     | false                                                   | exact string compare                                                                                          | non-`1` means disabled                                                             | enabled boolean                                    |
| `CARTULARY_TEST_ROUTE_TOKEN`                                                                    | reset/browser           | non-empty opaque string with at least 128 bits entropy                                                                | generated by harness stack when reset route enabled                                       | harness generated, env for attach mode          | invalid                                                 | not normalized                                                                                                | missing/low-entropy token: `configuration_error`, exit `2` before stack use        | redaction token only                               |
| Object-store runtime env                                                                        | app runtime             | `CARTULARY_S3_OBJECT_PRIMARY_*` endpoint, credentials, secure bool, bucket                                            | browser/dev MinIO local values                                                            | Make variable, env, config binding, default     | invalid for required members                            | endpoint normalized; credentials redacted                                                                     | app startup/reset failure according to Section 12                                  | redacted credential fields                         |
| Runtime root envs                                                                               | app runtime             | `CARTULARY__ROOTS__*__PATH` filesystem paths                                                                          | browser stack creates under runtime root                                                  | Make variable, env, config binding, default     | invalid                                                 | path normalization                                                                                            | invalid/unwritable path: `configuration_error` or app startup failure              | normalized path                                    |
| `CARTULARY_HARNESS_REPO_ROOT`, `CARTULARY_HARNESS_SCRATCH_ROOT`, `TMPDIR`                       | harness scratch         | filesystem path                                                                                                       | `${TMPDIR:-/tmp}/cartulary-harness-scratch`                                               | env, default                                    | invalid for explicit scratch                            | path normalization; scratch root must be outside repo                                                         | in-repo scratch root: `configuration_error`, exit `2`                              | normalized scratch root                            |
| `CARTULARY_CLEANUP_DRY_RUN`                                                                     | cleanup                 | exact `1`                                                                                                             | false                                                                                     | Make variable, env, default                     | false                                                   | exact string compare                                                                                          | non-`1` false                                                                      | dry-run boolean                                    |
| `LINT_SHELL_STRICT`                                                                             | lint                    | exact `1` makes shell lint blocking                                                                                   | false                                                                                     | Make variable, env, default                     | false                                                   | exact string compare                                                                                          | non-`1` false                                                                      | boolean when true                                  |

## 6. Result Roots, Run IDs, and Artifact Identity

**TH-HARNESS-REQ-150**
A public Make target that emits retained artifacts MUST compute artifact identity as:

```text
run_root = normalize_result_root(CARTULARY_TEST_RESULTS_DIR) / normalize_run_id(CARTULARY_TEST_RUN_ID)
```
Verified by: TH-HARNESS-AC-003, TH-HARNESS-AC-015

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
  fail with configuration_error if parent is not writable
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

## 7. Output Modes and Machine Output

**TH-HARNESS-REQ-200**
Public Make targets MUST support exactly these output modes: `quiet`, `summary`, `ci`, `verbose`, `debug`, and `machine`. Unknown output modes MUST fail with `configuration_error` and exit `2` before child work.
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
| `summary_with_artifacts`           | Leaf, toolchain, build, lint, drift, browser-stage, and formatting targets with wrapper summaries |                 yes | One `cartulary.tool_run_summary.v2` JSON object plus LF | Empty after wrapper starts; pre-wrapper diagnostics allowed only before JSON emission | Tool-run summary required                                 | Same schema with failure fields and nonzero exit                       |
| `service_summary_with_artifacts`   | `db-up`, `db-reset`, `services-up`, `minio-init`                                                  |                 yes | One `cartulary.tool_run_summary.v2` JSON object plus LF | Empty after wrapper starts                                                            | Tool-run summary and service diagnostics where applicable | Same schema with service failure fields                                |
| `aggregate_summary_with_artifacts` | `test-fast`, `test`, `lint`, `ci`, `release-check`, `build`                                       |                 yes | One `cartulary.tool_run_summary.v2` JSON object plus LF | Empty after wrapper starts                                                            | Aggregate summary plus child references                   | Same schema with primary failure                                       |
| `scheduler_summary_with_artifacts` | `check`, `phase-slice`, `service-backed-slice`                                                    |                 yes | One `cartulary.tool_run_summary.v2` JSON object plus LF | Empty after wrapper starts; no scheduler progress prose                               | Scheduler summary/events and run summary                  | Same schema with scheduler or child failure                            |
| `machine_stdout_json`              | `target-plan-json` and other explicitly declared JSON discovery targets                           |                 yes | One target-specific JSON object plus LF                 | Empty on success                                                                      | None unless target declares artifacts                     | Invalid input exits `2`; target-specific error JSON only when declared |
| `human_summary`                    | `help`, `help-all`, text discovery/explanation targets                                            |                  no | Empty                                                   | Bounded diagnostic allowed on failure                                                 | None unless target row declares diagnostic artifacts      | `machine` rejected as `usage_error`, exit `2`                          |
| `interactive_raw`                  | `dev`                                                                                             |                  no | Empty when `machine` requested                          | Diagnostic allowed                                                                    | None                                                      | `machine` rejected as `usage_error`, exit `2`                          |
| `destructive_human`                | `clean`, `distclean`                                                                              |                  no | Empty when `machine` requested                          | Diagnostic allowed                                                                    | None                                                      | `machine` rejected as `usage_error`, exit `2`                          |

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
For every public target whose output class accepts `machine`, stdout MUST be exactly one UTF-8 JSON object followed by one LF. The object MAY contain artifact pointers. Stdout MUST NOT be a pointer-only payload, multiple JSON objects, scheduler progress prose, child logs, or human summary text.
Verified by: TH-HARNESS-AC-004

**TH-HARNESS-REQ-202**
For every public target whose output class rejects `machine`, setting `CARTULARY_OUTPUT_MODE=machine` MUST fail before child work with `failure_class=config`, `failure_reason=usage_error`, public exit code `2`, empty stdout, and bounded stderr diagnostic.
Verified by: TH-HARNESS-AC-005

## 8. Artifact and Schema Contract

**TH-HARNESS-REQ-250**
A public Make-owned command that declares a stable schema ID MUST emit JSON that validates against the matching normative schema attachment before command success. If required artifact validation fails, the public target MUST fail with `artifact_error` or `scheduler_accounting_error` according to Section 9.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-004

The following schema IDs are public contracts. Schema file paths are repository attachments, not behavioral owners. A missing file in the current repository snapshot MUST be marked as a future attachment here rather than implied by prose.

| Schema ID                                       | Repository attachment path                                               | Status            | Producer class           | Required validation point                 |
| ----------------------------------------------- | ------------------------------------------------------------------------- | ----------------- | ------------------------ | ----------------------------------------- |
| `cartulary.tool_run_summary.v2`                 | `tools/schemas/cartulary.tool_run_summary.v2.schema.json`                 | present           | Centralized wrappers     | Before wrapper exits.                     |
| `cartulary.test_phase_summary.v3`               | `tools/schemas/cartulary.test_phase_summary.v3.schema.json`               | present           | Phase handlers           | Before target summary consumes it.        |
| `cartulary.test_target_summary.v4`              | `tools/schemas/cartulary.test_target_summary.v4.schema.json`              | present           | Target summary generator | Before aggregate/run summary consumes it. |
| `cartulary.test_run_summary.v6`                 | `tools/schemas/cartulary.test_run_summary.v6.schema.json`                 | present           | Run summary generator    | Before public aggregate success.          |
| `cartulary.check_scheduler_summary.v9`          | `tools/schemas/cartulary.check_scheduler_summary.v9.schema.json`          | present           | Check scheduler          | Before scheduler target success.          |
| `cartulary.service_backed_scheduler_summary.v9` | `tools/schemas/cartulary.service_backed_scheduler_summary.v9.schema.json` | present           | Service-backed scheduler | Before scheduler target success.          |
| `cartulary.scheduler_event.v6`                  | `tools/schemas/cartulary.scheduler_event.v6.schema.json`                  | present           | Scheduler                | During scheduler JSONL validation.        |
| `cartulary.test_services.lease.v1`              | `tools/schemas/cartulary.test_services.lease.v1.schema.json`              | present           | Service suite            | Before attach or cleanup relies on lease. |
| `cartulary.web_e2e_stack.v1`                    | `tools/schemas/cartulary.web_e2e_stack.v1.schema.json`                    | present           | Browser stack            | Before browser target starts Playwright.  |
| `cartulary.test.runtime_reset.v1`               | `tools/schemas/cartulary.test.runtime_reset.v1.schema.json`               | present           | Reset route/wrapper      | Before browser reset success is accepted. |
| `cartulary.fixture_report.v1`                   | `tools/schemas/cartulary.fixture_report.v1.schema.json`                   | present           | Fixture report target    | Before machine JSON is emitted.           |

**TH-HARNESS-REQ-251**
Schema-owned artifacts MUST be closed by default. Unknown top-level fields are invalid unless the schema declares an explicit extension container.
Verified by: TH-HARNESS-AC-000

**TH-HARNESS-REQ-252**
Every retained summary artifact MUST include normalized `result_root`, `run_id`, `run_root`, `target`, `output_mode`, public `exit_code`, primary `failure_class`, primary `failure_reason`, `started_at`, and `completed_at`. Timestamps MUST be RFC3339 UTC strings with non-null values.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-015

Public and machine summaries SHOULD print or serialize `run_root` once per summary and SHOULD express retained artifact references relative to `run_root` when the referenced path is under that directory. Absolute paths remain allowed only for external diagnostics that are not under the retained run root.

### 8.1 Artifact Families

| Artifact family                                      | Producer                                        | Path under run root                                             | Schema policy                                                 | Ordering and nullability                                                              | Retention and cleanup                                        |
| ---------------------------------------------------- | ----------------------------------------------- | --------------------------------------------------------------- | ------------------------------------------------------------- | ------------------------------------------------------------------------------------- | ------------------------------------------------------------ |
| Tool-run summary                                     | Centralized wrappers                            | `<target>/tool-run-summary.json` or target-specific summary dir | `cartulary.tool_run_summary.v2`                               | Required non-null timestamps, target, exit code, output mode, artifact refs, failures | Retained; removed by cleanup only under default result root. |
| Phase summary                                        | Phase handlers                                  | target phase dirs                                               | `cartulary.test_phase_summary.v3`                             | Stable phase/target/runner/status/count fields                                        | Retained.                                                    |
| Target summary                                       | Target summary generator                        | `<target>/target-summary.json`                                  | `cartulary.test_target_summary.v4`                            | Child/totals rollups ordered by registry order                                        | Retained.                                                    |
| Run summary                                          | Run summary generator                           | `run-summary.json` or aggregate dir                             | `cartulary.test_run_summary.v6`                               | Work units and artifact dirs ordered deterministically                                | Retained.                                                    |
| Scheduler summary                                    | Scheduler                                       | `<target>/scheduler-summary.json`                               | Scheduler summary schema by scheduler type                    | Work units by manifest ordinal; resources by registry order                           | Retained.                                                    |
| Scheduler event stream                               | Scheduler                                       | `<target>/scheduler-events.jsonl`                               | `cartulary.scheduler_event.v6`                                | `seq` strictly increases with no gaps                                                 | Retained.                                                    |
| Scheduler progress summary                           | Scheduler reporter                              | `<target>/progress-summary.log`                                 | diagnostic-only                                               | Bounded progress snapshots                                                            | Retained.                                                    |
| Service scope artifacts                              | Service suite                                   | `_shared/test-services/<suite-id>/...`                          | lease schema required; other service logs diagnostic-only     | Lease fields closed by Section 11                                                     | Retained; cleanup may append diagnostics.                    |
| Browser stack metadata                               | Browser stack                                   | browser target support dir                                      | `cartulary.web_e2e_stack.v1`                                  | Origins, ports, runtime root, log paths required                                      | Retained for browser target.                                 |
| Reset response/status/state                          | Reset route/wrapper                             | `reset-boundary/*.json`, `*.status`, `*.state-reset`            | `cartulary.test.runtime_reset.v1` for reset data              | Reset ID, table list, migration/admin flags, object count required                    | Retained for browser target.                                 |
| Generated manifest summaries                         | Generation/drift scripts                        | tool-specific target dirs                                       | JSON schemas declared by generated artifacts                  | Unknown fields rejected where shape tools enforce closure                             | Generated files remain checked in; summaries retained.       |
| Logs                                                 | Shell, Go, scheduler, browser, service wrappers | target log dirs                                                 | diagnostic-only unless producer declares schema               | Logs are text after redaction; empty logs may be omitted                              | Retained unless cleanup removes result root.                 |
| Coverage reports                                     | Go/frontend/test tools                          | tool-specific coverage paths                                    | diagnostic-only                                               | Tool-defined                                                                          | Removed by `make clean` when under registered paths.         |
| Playwright screenshots, videos, traces, HTML reports | Playwright                                      | Playwright report/test-results dirs                             | diagnostic-only secret-bearing                                | Tool-defined                                                                          | Removed by `make clean` when under registered paths.         |
| Visual snapshots and goldens                         | Browser/fixture tools                           | source and tool-specific dirs                                   | validation-only unless future profile adopts refresh contract | Tool-defined                                                                          | Refresh is outside current conformance.                      |

## 9. Failure Classes and Exit Codes

**TH-HARNESS-REQ-300**
Public Make-owned wrappers MUST expose exact public exit codes according to the failure-reason table below. Raw child process exit codes MAY be preserved in summaries but MUST NOT define the public wrapper exit code except where `child_target_failure` explicitly delegates to a normalized child failure class.
Verified by: TH-HARNESS-AC-014

Failure classification uses two layers:

- `failure_class`: coarse stable grouping for humans and automation.
- `failure_reason`: detailed snake-case reason for diagnosis, exit-code mapping, and handoff.

| Failure class | Meaning                                                                                 |
| ------------- | --------------------------------------------------------------------------------------- |
| `product`     | The product behavior under test failed after harness setup completed.                   |
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
| `fixture_error`               | `harness`     | DB/bucket/template/reset/janitor/fixture operation fails           |                                                 `3` |
| `resource_conflict`           | `infra`       | Logical resource, port, lock, DB/bucket name, or host conflict     |                                                 `4` |
| `test_assertion_failure`      | `product`     | Test runner assertion fails after harness setup                    |                                                `10` |
| `child_target_failure`        | `harness`     | Aggregate child exits nonzero                                      |                         normalized child class exit |
| `scheduler_accounting_error`  | `harness`     | Manifest, summary, timing, event, or accounting mismatch           |                                                `11` |
| `artifact_error`              | `artifact`    | Required artifact missing, invalid, unredacted, or schema-invalid  |                                                `11` |
| `cleanup_error`               | `harness`     | Cleanup command/finalizer/leak check/reaper scheduling fails       |         `12` when no earlier primary failure exists |
| `timeout_failure`             | `timing`      | Command, readiness, watchdog, cleanup, or lock exceeds deadline    |                                                `13` |
| `cancelled_or_interrupted`    | `interrupted` | Signal, cancellation, abort                                        | `130` for SIGINT, `143` for SIGTERM, otherwise `15` |
| `unknown_failure`             | `unknown`     | Failure cannot be classified                                       |                                                 `1` |

Default human output SHOULD expose dense failure fields and avoid full failure records unless verbose output is requested. The canonical compact shape is:

```text
failure_class=infra reason=service_readiness_timeout failed=<unit>
```

Full failure records belong in retained JSON summaries and investigation commands.

### 9.1 Primary Failure Selection

```text
select_primary_failure(failures):
  if failures is empty:
    return success
  if any non-cleanup failure exists:
    return earliest non-cleanup failure by:
      1. top-level command lifecycle order,
      2. scheduler event sequence if scheduler-owned,
      3. child target registry order if aggregate-owned,
      4. artifact path lexical order
  return earliest cleanup failure by cleanup step order
```

**TH-HARNESS-REQ-301**
Cleanup failure after an earlier product or operational failure MUST be recorded but MUST NOT override the public exit code selected for the earlier primary failure.
Verified by: TH-HARNESS-AC-014

**TH-HARNESS-REQ-302**
Harness setup, readiness, fixture, artifact, scheduler, timeout, and cleanup failures MUST NOT use `failure_class=product`. A failing assertion after successful harness setup MUST be classified with `failure_class=product` and `failure_reason=test_assertion_failure`.
Verified by: TH-HARNESS-AC-013, TH-HARNESS-AC-014

## 10. Scheduler Contract

**TH-HARNESS-REQ-350**
Scheduler manifests are normative scheduler inputs. A scheduler target MUST validate manifest schema, work-unit IDs, dependencies, resource claims, finalizers, output schemas, and timing settings before starting child work.
Verified by: TH-HARNESS-AC-006

### 10.1 Scheduler Manifest Fields

The canonical scheduler input schema is `cartulary.scheduler_manifest.v1`.
`tools/scheduler_manifest.json` is the committed generated scheduler input for
check, service-backed, phase-slice, and future scheduler families. Family-
specific source forms such as check schedule metadata, service-backed
`work_unit_sources[]`, Go shard expansion, and browser group expansion MAY exist
only as upstream authoring inputs. Scheduler runners MUST NOT accept those
family-specific source forms as runtime scheduler manifests.

| Field                              | Type                 | Required | Default                      | Rule                                                        |
| ---------------------------------- | -------------------- | -------: | ---------------------------- | ----------------------------------------------------------- |
| `schema_id`                        | string               |      yes | none                         | Must be `cartulary.scheduler_manifest.v1`.                  |
| `generated`                        | object               |      yes | none                         | Generator and authoring-input provenance.                   |
| `schedules[]`                      | array                |      yes | none                         | Normalized scheduler inputs.                                |
| `schedules[].target`               | string               |      yes | none                         | Public target or scheduler target identity.                 |
| `schedules[].scheduler_kind`       | string               |      yes | target-specific              | `check`, `service_backed`, `phase_slice`, or future family. |
| `schedules[].capacity_profile`     | string               |      yes | none                         | Registry-backed capacity profile name.                      |
| `schedules[].resource_limits`      | object               |      yes | none                         | Logical resource limits or `auto` policies.                 |
| `schedules[].stop_on_first_failure` | boolean             |      yes | target-specific              | Check scheduler: `true`; service-backed scheduler: `false`. |
| `schedules[].progress_tick_seconds` | integer             |      yes | `30`                         | Must be `5..300`; affects reporting only.                   |
| `schedules[].validate_timing`      | boolean              |      yes | `true`                       | Must be `true` for conformance runs.                        |
| `schedules[].summary_groups`       | array                |       no | `[]`                         | Summary grouping policy for scheduler output.               |
| `schedules[].work_units[]`         | array                |      yes | none                         | Ordered by normalized manifest ordinal.                     |
| `work_units[].id`                  | string               |       no | `target`                     | Unique within schedule when defaulted.                      |
| `work_units[].command`             | object               |      yes | none                         | Structured descriptor resolved by the scheduler runner.     |
| `work_units[].priority`            | integer              |       no | `0`                          | Higher integer wins among ready work.                       |
| `work_units[].weight_ms`           | positive integer     |      yes | none                         | Advisory duration estimate only.                            |
| `work_units[].needs[]`             | string array         |       no | `[]`                         | Completion keys required before start.                      |
| `work_units[].completion_keys[]`   | string array         |       no | `[work_unit.id]`             | Added on success.                                           |
| `work_units[].failure_keys[]`      | string array         |       no | completion keys              | Added on failure.                                           |
| `work_units[].complete_on_failure` | boolean              |       no | `false`                      | Adds completion keys even when command fails.               |
| `work_units[].resource_claims`     | object               |      yes | `{}`                         | Logical claims only.                                        |
| `work_units[].timeout_seconds`     | integer              |       no | target-family default        | Must be positive and bounded by registry.                   |
| `work_units[].retained_resource_claims` | object         |       no | `{}`                         | Claims kept after work-unit exit until explicit release.    |
| `work_units[].release_retained_resource_claims` | object |       no | `{}`                         | Retained claims to release after work-unit exit.            |
| `schedules[].finalizers[]`         | array                |      yes | `[]`                         | Always run after scheduler drains or stops.                 |

Supported `work_units[].command.type` values are `make_target`,
`service_session_start`, `browser_stage_session_start`, `browser_group`,
`browser_stage_complete`, `go_shard`, `go_shard_finalize`, and
`service_complete`. Dependency-gated aggregate work, including Go shard
aggregation, MUST remain in `work_units[]` and MUST NOT be modeled as an
unconditional scheduler finalizer.

`weight_ms` is an advisory scheduling estimate. It MUST NOT be treated as a logical resource claim, timeout, benchmark claim, pass/fail threshold, or product performance conformance statement.

### 10.2 Logical Resource Registry

| Resource                     | Scheduler             | Default or auto policy                                                                   |    Bound |
| ---------------------------- | --------------------- | ---------------------------------------------------------------------------------------- | -------: |
| `host_cpu`                   | check                 | auto host CPU, override `CHECK_HOST_CPU_JOBS`                                            | `1..256` |
| `host_io`                    | check                 | auto host IO, override `CHECK_HOST_IO_JOBS`                                              | `1..256` |
| `suite_service_stack`        | check                 | `1`                                                                                      |   `1..4` |
| `migration_scratch_postgres` | check                 | `1`                                                                                      |   `1..4` |
| `go_cpu`                     | service-backed        | auto Go CPU, override `CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT`                            | `1..256` |
| `go_io`                      | service-backed        | auto Go IO, override `CARTULARY_SERVICE_BACKED_GO_IO_LIMIT`                              | `1..256` |
| `browser_stack`              | check, service-backed | auto browser stack, override `CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT`              |   `1..8` |
| `process`                    | check, service-backed | `6`                                                                                      | `1..256` |
| `postgres`                   | check, service-backed | `32`                                                                                     | `1..256` |
| `minio`                      | check, service-backed | `32`                                                                                     | `1..256` |
| `postgres_reset`             | check, service-backed | `1`                                                                                      |   `1..8` |
| `browser_stage_*`            | check, service-backed | generated per browser stage; default `1` unless manifest declares another positive value |   `1..8` |

### 10.3 Scheduling Algorithm

```text
pending = work_units in manifest order
running = empty map
completed_keys = empty set
failed_keys = empty set
primary_failure = null
scheduler_stopped = false
emit scheduler_started

while pending is not empty or running is not empty:
  ready = pending units whose dependencies are all in completed_keys
          and whose dependencies are not in failed_keys

  for unit in ready by priority DESC, weight_ms DESC, manifest ordinal ASC, id ASC:
    if scheduler_stopped:
      break
    if any earlier ready unit is resource-blocked and overlaps unit.resource_claims:
      continue
    if resources_available(unit.resource_claims):
      start unit
      remove unit from pending
      add unit to running
      emit unit_started

  if running is empty:
    if pending contains units with failed dependencies:
      mark them skipped in manifest order
      remove them from pending
      continue
    fail with scheduler_accounting_error for deadlock or impossible resources

  wait until one or more running units finish or the progress tick fires

  if progress tick fires:
    emit progress event
    continue

  process finished units by:
    1. observed_monotonic_finished_at ascending,
    2. manifest ordinal ascending

  for each finished unit:
    release non-retained resources
    record status, logs, duration, completion keys, failure keys
    if success or unit.complete_on_failure:
      add completion_keys
    if failure:
      add failure_keys
      set primary_failure if null
      if stop_on_first_failure:
        scheduler_stopped = true

run finalizers in manifest order after all running units drain
release retained resource claims
emit scheduler summary, progress summary, and scheduler_complete
validate timing when validate_timing is true
exit with selected primary failure
```

Finalizer failure becomes primary only when no earlier non-finalizer failure exists.

Dependencies outrank priority: a work unit is not ready until its dependencies are satisfied. Priority affects only ready work and MUST NOT preempt work that is already running.

### 10.4 Event Ordering

| Event field             | Rule                                                                         |
| ----------------------- | ---------------------------------------------------------------------------- |
| `schema_id`             | `cartulary.scheduler_event.v6`.                                              |
| `target`                | Public target or scheduler target identity.                                  |
| `scheduler_kind`        | Scheduler family such as `check`, `service_backed`, or `phase_slice`.        |
| `seq`                   | Starts at `1`, increments by `1`, no gaps.                                   |
| `event`                 | Compact event token such as `scheduler-started`, `unit-started`, or `progress`. |
| `monotonic_ms`          | Non-decreasing scheduler-relative monotonic time.                            |
| `emitted_at`            | RFC3339 UTC. Wall-clock regressions require a `clock-skew` marker event.     |
| Work-unit ordering      | Manifest ordinal unless completion tie rule applies.                         |
| Completion tie          | `observed_monotonic_finished_at` ascending, then manifest ordinal ascending. |
| Artifact ordering       | Lexicographic by normalized artifact path.                                   |
| Resource ordering       | Registry display order, then lexicographic fallback.                         |

## 11. Service and Fixture Lifecycle

**TH-HARNESS-REQ-400**
Service-backed targets MUST run in exactly one service mode: `owned` or `attach`.
Verified by: TH-HARNESS-AC-007

| Mode     | Selection rule                                       | Missing variables                                                      | Ownership                                                                     |
| -------- | ---------------------------------------------------- | ---------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| `owned`  | `CARTULARY_TEST_SERVICES_ACTIVE` omitted or not `1`. | Not applicable.                                                        | Harness starts and cleans suite resources.                                    |
| `attach` | `CARTULARY_TEST_SERVICES_ACTIVE=1`.                  | Any missing required attach variable fails with `configuration_error`. | Harness uses supplied services and MUST NOT delete container-level resources. |

### 11.1 Suite State Machine

```text
requested -> starting -> ready -> running_child -> cleaning -> cleaned
requested -> starting -> failed_start
ready -> running_child -> interrupted -> cleaning
cleaning -> cleanup_failed
```

### 11.2 Lease Fields

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

### 11.3 Readiness Deadlines

| Resource                     | Deadline | Poll interval | Failure reason                       |
| ---------------------------- | -------: | ------------: | ------------------------------------ |
| Docker preflight             |    `15s` |          `1s` | `preflight_error`                    |
| Postgres container readiness |   `180s` |       `500ms` | `service_readiness_timeout`          |
| MinIO readiness              |   `120s` |       `500ms` | `service_readiness_timeout`          |
| Template DB migration        |   `180s` |           n/a | `fixture_error`                      |
| Browser backend readiness    |   `120s` |       `500ms` | `service_readiness_timeout`          |
| Browser frontend readiness   |   `120s` |       `500ms` | `service_readiness_timeout`          |
| Reset route success          |    `30s` |           n/a | `fixture_error` or `timeout_failure` |

### 11.4 Retry and Teardown Rules

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
| Postgres owned startup      |            `3` | `500ms` | transient Docker startup/readiness transport timeout only  | `120s`           | Failed attempt container is terminated first.     |
| MinIO owned startup         |            `2` | `250ms` | transient Docker startup/readiness transport timeout only  | `300s`           | Failed attempt container is terminated first.     |
| Template DB migration       |            `1` | none    | none                                                       | `180s`           | none                                              |
| Browser backend startup     |            `1` | none    | none                                                       | `120s`           | readiness polling only                            |
| Browser frontend startup    |            `1` | none    | none                                                       | `120s`           | readiness polling only                            |
| Runtime reset route         |            `1` | none    | none                                                       | `30s`            | none                                              |
| Owned teardown and cleanup  |            `1` | none    | none                                                       | cleanup-specific | cleanup records failure and leaves proof for janitor |

Attach mode MAY write diagnostic records and lease observations. It MUST NOT delete externally supplied services, containers, databases, buckets, or object prefixes.

### 11.5 Duration Baselines

Duration baselines are advisory scheduler planning data only. They MUST NOT become benchmark claims, product performance conformance, timeout policy, or evidence that product behavior is fast enough.

Baseline values MUST be positive integer `weight_ms` values derived only from successful, uncontaminated retained runs. Missing entries MUST use explicit default weights and MUST be reported as defaulted, not silently ignored.

Baseline refresh MUST reject contaminated evidence, including failed scheduler runs, service startup retries, service failures, reset taint, missing timing events, or interrupted runs.

Duration-baseline drift checks MAY fail only for severe stale planning. Compact drift diagnostics MUST include `subject`, `planned_ms`, `actual_ms`, `ratio`, and `kind`.

## 12. Test-Only Reset Route

**TH-HARNESS-REQ-450**
`POST /api/v1/test/runtime/reset` is a harness test route. It MUST be unavailable unless every enablement predicate below is satisfied. It MUST NOT be documented as production API behavior.
Verified by: TH-HARNESS-AC-008, TH-HARNESS-AC-013

### 12.1 Enablement

| Predicate                      | Required value                                                                      |
| ------------------------------ | ----------------------------------------------------------------------------------- |
| `CARTULARY_ENABLE_TEST_ROUTES` | Exact `1`.                                                                          |
| `CARTULARY_TEST_ROUTE_TOKEN`   | Non-empty string with at least 128 bits of entropy, generated by the harness stack. |
| Runtime ownership              | Server started by a Make-owned browser or test harness stack.                       |
| Production/default runtime     | Route is not registered.                                                            |

### 12.2 Authorization

| Request condition                                     | Behavior                                               |
| ----------------------------------------------------- | ------------------------------------------------------ |
| Route not enabled                                     | Return ordinary not-found behavior.                    |
| Route enabled, missing token header                   | `403`, `error.code=test_route_forbidden`.              |
| Route enabled, wrong token header                     | `403`, `error.code=test_route_forbidden`.              |
| Route enabled, correct `X-Cartulary-Test-Route-Token` | Evaluate request.                                      |
| Cookie-authenticated request without token            | Forbidden; session auth does not authorize this route. |

CSRF does not apply because cookie authentication is not accepted as authorization for this route.

### 12.3 Request Body

| Body                                | Behavior                                        |
| ----------------------------------- | ----------------------------------------------- |
| No body                             | Accepted.                                       |
| `{}` JSON object                    | Accepted.                                       |
| JSON object with any members        | `400`, `error.code=invalid_test_reset_request`. |
| Non-object JSON                     | `400`, `error.code=invalid_test_reset_request`. |
| Invalid JSON with JSON content type | `400`, `error.code=invalid_test_reset_request`. |

### 12.4 Concurrency and Timeout

| Condition            | Behavior                                                                                    |
| -------------------- | ------------------------------------------------------------------------------------------- |
| No reset active      | Acquire reset lock and run reset.                                                           |
| Reset already active | `409`, `error.code=test_runtime_reset_in_progress`.                                         |
| Reset exceeds `30s`  | `503`, `error.code=test_runtime_reset_timeout`; response includes failed action when known. |

### 12.5 Reset Algorithm and Partial Failure

The reset route MUST preserve migration metadata, restore the active deployment admin, truncate mutable incident/product state, clear route idempotency state, and clear the configured object store bucket or prefix for the harness-owned runtime.

Database table truncation and bootstrap restoration MUST execute in one database transaction when the database supports that transaction shape. Object-store deletion occurs after the database transaction commits. The route MUST NOT claim rollback across object-store deletion.

| Failure point                                      | Required response                                                                                                  |
| -------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| Before DB transaction commit                       | `500`, `error.code=test_runtime_reset_failed`, `partial_failure=false` unless prior mutation occurred.             |
| After DB commit and before object cleanup complete | `500`, `error.code=test_runtime_reset_failed`, `partial_failure=true`.                                             |
| Object cleanup partial deletion                    | `500`, `error.code=test_runtime_reset_failed`, `partial_failure=true`, include `object_count_after` if measurable. |
| Bootstrap admin not restored                       | `500`, `error.code=test_runtime_reset_failed`, `partial_failure=true`.                                             |

A browser wrapper receiving `partial_failure=true` MUST mark the owned stack tainted and restart it before further browser child work.

### 12.6 Success Readiness

Success requires all of the following:

- migration metadata preserved;
- active deployment admin restored;
- mutable incident/product tables empty;
- route idempotency table empty;
- object count after reset equals `0`;
- response validates against `cartulary.test.runtime_reset.v1`.

## 13. Cleanup and Destructive Safety

**TH-HARNESS-REQ-500**
Cleanup is destructive. Cleanup commands MUST delete only paths or resources satisfying the exact ownership predicates in this section.
Verified by: TH-HARNESS-AC-009, TH-HARNESS-AC-010

`make clean` and `make distclean` are repo-local cleanup commands. They MUST NOT remove caller-supplied external result roots.

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
  lstat path
  if path is symlink:
    unlink symlink object only
    MUST NOT follow target
  if path is directory:
    remove directory tree only after every traversed entry remains under the candidate root by lexical path and lstat traversal
```

### 13.2 Cleanup Scope

| Command               |      Removes default result root? | Removes custom `CARTULARY_TEST_RESULTS_DIR`? | Removes external Go caches? | Stops Docker/Compose globally? |
| --------------------- | --------------------------------: | -------------------------------------------: | --------------------------: | -----------------------------: |
| `make clean`          | yes, only default registered path |                                           no |                          no |                             no |
| `make distclean`      | yes, only default registered path |                                           no |                          no |                             no |
| Service-suite cleanup |        only suite-owned artifacts |                                           no |                          no |                             no |
| Stale janitor         |        proof-gated resources only |                                           no |                          no |                             no |

### 13.3 Stale Janitor Thresholds

| Resource        | Completed-run predicate                                         | Uncompleted stale predicate                | Active-resource rule                                                                     |
| --------------- | --------------------------------------------------------------- | ------------------------------------------ | ---------------------------------------------------------------------------------------- |
| Database        | Completed summary or lease cleanup state older than 15 minutes. | Lease or metadata older than 24 hours.     | Active connections may be terminated only after proof predicate passes.                  |
| Bucket          | Completed summary or lease cleanup state older than 15 minutes. | Lease or metadata older than 24 hours.     | Delete only generated bucket/prefix with proof metadata.                                 |
| Container       | Completed summary or lease cleanup state older than 15 minutes. | Harness Docker labels older than 24 hours. | Running container may be stopped only if proof predicate passes and label owner matches. |
| Browser fixture | Completed target summary older than 15 minutes.                 | Fixture metadata older than 6 hours.       | Delete only generated fixture directory with ownership metadata.                         |

### 13.4 Dry-Run Contract

| Setting                                        | Behavior                                                                                                                    |
| ---------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| `CARTULARY_CLEANUP_DRY_RUN` omitted or not `1` | Cleanup may delete resources satisfying predicates.                                                                         |
| `CARTULARY_CLEANUP_DRY_RUN=1`                  | Cleanup MUST emit deletion candidates and reasons and MUST NOT delete files, DBs, buckets, containers, or browser fixtures. |

Dry-run output MUST include normalized path or resource identity, proof predicate, action that would be taken, and rejection reason for retained candidates. For human destructive targets, the dry-run line format is:

```text
DRY-RUN <action> <normalized-identity> <proof-or-rejection-reason>
```

### 13.5 Parent-Death and Reaper Rule

Immediate cleanup after parent death is not guaranteed. The conformance guarantee is:

- owned resources carry enough lease or proof metadata for later stale janitor evaluation;
- detached reaper scheduling is optional unless the command declares it;
- if a detached reaper is scheduled, it writes `reaper-scheduled.json` with lease ID, started-at timestamp, target resources, and timeout.

## 14. Platform and CI Support

**TH-HARNESS-REQ-550**
The current conformance support matrix is closed by this section. A target may be run elsewhere, but unsupported environments MUST NOT be used for current harness conformance claims.
Verified by: TH-HARNESS-AC-012

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

## 15. Security and Redaction

**TH-HARNESS-REQ-600**
Centralized summaries, machine output, and retained logs captured by harness wrappers MUST be redacted before retention and before stdout emission.
Verified by: TH-HARNESS-AC-011

Screenshots, videos, traces, and Playwright HTML reports are diagnostic secret-bearing artifacts. They MUST NOT be described as safe to upload or publish without separate review.

### 15.1 Secret Pattern Table

| Secret class               | Match rule                                                                          | Redaction token                      |
| -------------------------- | ----------------------------------------------------------------------------------- | ------------------------------------ |
| Passwords                  | Variable or key contains `PASSWORD`, `PASS`, or `PWD`                               | `[REDACTED:password]`                |
| Tokens                     | Contains `TOKEN`, `JWT`, `BEARER`, `SESSION`, or `COOKIE`                           | `[REDACTED:token]`                   |
| API or access keys         | Contains `API_KEY`, `ACCESS_KEY`, `SECRET_KEY`, or `KEY_ID` with credential context | `[REDACTED:key]`                     |
| DSNs/URLs with credentials | URL userinfo or DSN password segment present                                        | `[REDACTED:dsn]`                     |
| Object-store credentials   | MinIO/S3 access key or secret key variables                                         | `[REDACTED:object-store-credential]` |
| Private keys               | PEM private-key block markers                                                       | `[REDACTED:private-key]`             |

### 15.2 Artifact Redaction Table

| Artifact class                      | Redaction requirement                                                                                                      |
| ----------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| Tool/run/target/scheduler summaries | Redact before write.                                                                                                       |
| Machine stdout JSON                 | Redact before stdout.                                                                                                      |
| Captured child stdout/stderr logs   | Redact before retention.                                                                                                   |
| Service env files                   | Store only redacted credential values unless the file is required for child execution and kept under private runtime root. |
| Browser screenshots/videos/traces   | Diagnostic secret-bearing; not safe for publication.                                                                       |
| Playwright HTML reports             | Diagnostic secret-bearing; not safe for publication.                                                                       |
| CI logs                             | Redact using the same token rules before harness-controlled emission.                                                      |

## 16. Integration with Product Specifications

**TH-HARNESS-REQ-650**
Harness tests MAY reference product requirement IDs, phase ledgers, coverage ledgers, phase maps, and acceptance criteria. Those references provide evidence routing only. They MUST NOT redefine product behavior under test.
Verified by: TH-HARNESS-AC-013, TH-HARNESS-AC-016

**TH-HARNESS-REQ-651**
Support-only tests, helper phases, raw aggregate suites, and direct package scripts MUST NOT be counted as authoritative product conformance evidence unless a canonical Make target adopts them and emits required evidence artifacts.
Verified by: TH-HARNESS-AC-013

**TH-HARNESS-REQ-652**
Load, login bursts, service resets, artificial stress margins, and browser harness stress tests are harness-only unless product specs explicitly adopt them. The Phase 6 browser login burst remains a harness stress margin and does not change Core 04 session-cap semantics.
Verified by: TH-HARNESS-AC-013

## 17. Acceptance Criteria / Definition of Done

The acceptance matrix is the harness Definition of Done. Each row is binary. A row passes only when its setup, invocation, exit/status, stdout/stderr, artifact, and cleanup expectations all match.

| ID                | Requirement owner  | Scope                            | Setup fixture                                                                | Invocation                                                                                | Expected exit/status                                           | Stdout                                                 | Stderr                                                       | Required artifacts                                                                                 | Negative case                                                                | Cleanup expectation                                     |
| ----------------- | ------------------ | -------------------------------- | ---------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- | -------------------------------------------------------------- | ------------------------------------------------------ | ------------------------------------------------------------ | -------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- | ------------------------------------------------------- |
| TH-HARNESS-AC-000 | Section 8          | Schema validation                | Any public target that emits required JSON                                   | Target-specific                                                                           | Success only if JSON validates                                 | Per Section 7                                          | Per Section 7                                                | Every emitted required JSON artifact validates against Section 8 schema attachments                | Inject schema-invalid required summary                                       | No extra cleanup beyond target contract                 |
| TH-HARNESS-AC-001 | Sections 1, 4      | Command registry                 | Current tree                                                                 | `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all` plus registry parity checker    | `0` when registry matches exactly                              | Bounded report                                         | Empty on success                                             | Public target registry parity report                                                               | Extra/missing public target fails                                            | none                                                    |
| TH-HARNESS-AC-002 | Section 5          | Config precedence                | Fixture target with CLI, Make var, env, manifest, config, default candidates | Dedicated config resolver test target or unit harness                                     | `0`                                                            | Machine or bounded summary                             | Empty on success                                             | Resolver summary showing CLI > Make var > env > manifest > config file > default                   | Non-positive scheduler limit exits `2` with `configuration_error`            | no child work                                           |
| TH-HARNESS-AC-003 | Sections 5, 6      | Result root and run ID           | No child work required                                                       | Invalid result root and invalid run ID fixtures                                           | `2`                                                            | Empty or failure JSON according to target output class | Bounded config diagnostic                                    | Failure summary when wrapper starts                                                                | Slash, backslash, whitespace, `.`, `..`, existing non-empty run dir all fail | no child work                                           |
| TH-HARNESS-AC-004 | Sections 7, 8      | Machine output accepted          | Toolchain ready; explicit result root/run ID                                 | `CARTULARY_OUTPUT_MODE=machine make backend-unit`; `... make test-fast`; `... make check` | Target status                                                  | Exactly one JSON object plus LF                        | Empty after wrapper starts                                   | `cartulary.tool_run_summary.v2` and target artifacts                                               | Progress prose or duplicate JSON fails                                       | normal target cleanup                                   |
| TH-HARNESS-AC-005 | Section 7          | Machine output rejected          | No child work                                                                | `CARTULARY_OUTPUT_MODE=machine make clean`; `... make dev`; `... make help`               | `2`                                                            | Empty                                                  | Bounded `usage_error` diagnostic                             | None required                                                                                      | Child work starts despite rejection                                          | no deletion or service start                            |
| TH-HARNESS-AC-006 | Section 10         | Scheduler determinism            | Controlled manifest with simultaneous child completions                      | Run scheduler fixture twice with same manifest                                            | `0`                                                            | Bounded summary or machine object                      | Empty on success                                             | Byte-identical scheduler events after dynamic timestamp normalization allowed only by schema rules | Event sequence differs                                                       | finalizers run                                          |
| TH-HARNESS-AC-007 | Section 11         | Service modes                    | Owned and attach fixtures                                                    | Owned service target; attach target missing one required var                              | owned success; attach failure `2`                              | Bounded summary                                        | Empty on owned success; config diagnostic for attach failure | Owned lease before child work; attach failure summary                                              | Attach mode deletes container-level resource                                 | owned teardown recorded                                 |
| TH-HARNESS-AC-008 | Section 12         | Reset route                      | Browser test runtime with test route token                                   | Reset route success, auth rejection, concurrent reset, timeout, partial failure fixtures  | Expected HTTP statuses from Section 12                         | HTTP JSON response                                     | n/a                                                          | Reset response validates schema; tainted stack marker on partial failure                           | Default runtime exposes route                                                | tainted stack restarted before further work             |
| TH-HARNESS-AC-009 | Section 13         | Cleanup path guard               | Synthetic registry with safe and unsafe paths                                | Cleanup guard unit and `CARTULARY_CLEANUP_DRY_RUN=1 make clean`                           | `0` for safe dry-run; nonzero for unsafe synthetic path        | Dry-run lines match format                             | Bounded guard diagnostic on unsafe path                      | Candidate list or guard evidence                                                                   | Empty path, `/`, `.`, `..`, traversal, outside-repo path accepted            | no deletion in dry-run                                  |
| TH-HARNESS-AC-010 | Section 13         | Stale janitor proof gates        | Fake DB, bucket, container, and browser fixtures with/without proof          | Focused stale-janitor tests                                                               | `0`                                                            | Bounded summary                                        | Empty on success                                             | Evidence that unproven resources retained and proven stale fixtures deleted only outside dry-run   | Resource lacking generated name/proof deleted                                | unproven resources retained                             |
| TH-HARNESS-AC-011 | Section 15         | Redaction                        | Fake DSN, object-store secret, token, and private-key fixtures               | Redaction unit plus one wrapper log capture                                               | `0`                                                            | No unredacted secret in machine JSON                   | No unredacted secret in captured stderr                      | Summaries/logs contain required redaction tokens                                                   | Any secret pattern appears unredacted                                        | none                                                    |
| TH-HARNESS-AC-012 | Section 14         | Platform matrix                  | Platform claim checker fixture                                               | Platform matrix checker                                                                   | `0` for allowed profiles; nonzero for unsupported claim        | Bounded summary                                        | Diagnostic on unsupported claim                              | Matrix report                                                                                      | macOS/Windows-native/Podman claimed as current conformance                   | none                                                    |
| TH-HARNESS-AC-013 | Sections 9, 16     | Product versus harness failure   | One known failing assertion and one harness setup failure                    | Canonical test target under each fixture                                                  | Product failure exits `10`; setup failure exits Section 9 code | Failure headline names class and reason                | Bounded diagnostic                                           | Target/tool summary with failure class and reason                                                  | Setup failure classified as product                                          | harness cleanup attempted                               |
| TH-HARNESS-AC-014 | Section 9          | Exit-code matrix                 | Controlled failure fixtures                                                  | Exit matrix test target                                                                   | Exact Section 9 code for every class                           | Per output mode                                        | Per output mode                                              | Failure summaries with primary failure selection                                                   | Cleanup failure overrides earlier product failure                            | cleanup failure recorded but primary exit preserved     |
| TH-HARNESS-AC-015 | Sections 6, 8      | Retained artifact identity       | Explicit result root/run ID                                                  | `CARTULARY_TEST_RESULTS_DIR=<dir> CARTULARY_TEST_RUN_ID=<id> make backend-unit`           | `0`                                                            | Summary names run root                                 | Empty                                                        | Artifacts under `<dir>/<id>` with target, run ID, run root                                         | Newest-run fallback accepted as proof                                        | custom absolute result root not removed by `make clean` |
| TH-HARNESS-AC-016 | Sections 1, 18, 19 | Editorial and adoption readiness | Revised document                                                             | Editorial lint and future-decision scanner                                                | `0`                                                            | Bounded summary                                        | Empty on success                                             | No prohibited evidence markers in Sections 1-17; no current blockers in Section 19                 | Current-profile blocker appears in future section                            | none                                                    |

### 17.1 Requirement-to-Acceptance Traceability

| Requirement range         | Owner section                      | Acceptance criteria                                     |
| ------------------------- | ---------------------------------- | ------------------------------------------------------- |
| `TH-HARNESS-REQ-001..049` | Status, scope, authority, purpose  | TH-HARNESS-AC-013, TH-HARNESS-AC-015, TH-HARNESS-AC-016 |
| `TH-HARNESS-REQ-050..099` | Public command surface             | TH-HARNESS-AC-001, TH-HARNESS-AC-004, TH-HARNESS-AC-005 |
| `TH-HARNESS-REQ-100..149` | Configuration                      | TH-HARNESS-AC-002, TH-HARNESS-AC-003                    |
| `TH-HARNESS-REQ-150..199` | Result roots and artifact identity | TH-HARNESS-AC-003, TH-HARNESS-AC-015                    |
| `TH-HARNESS-REQ-200..249` | Output modes                       | TH-HARNESS-AC-004, TH-HARNESS-AC-005                    |
| `TH-HARNESS-REQ-250..299` | Artifacts and schemas              | TH-HARNESS-AC-000, TH-HARNESS-AC-004, TH-HARNESS-AC-015 |
| `TH-HARNESS-REQ-300..349` | Failure and exit codes             | TH-HARNESS-AC-013, TH-HARNESS-AC-014                    |
| `TH-HARNESS-REQ-350..399` | Scheduler                          | TH-HARNESS-AC-006                                       |
| `TH-HARNESS-REQ-400..449` | Services                           | TH-HARNESS-AC-007, TH-HARNESS-AC-010                    |
| `TH-HARNESS-REQ-450..499` | Reset route                        | TH-HARNESS-AC-008                                       |
| `TH-HARNESS-REQ-500..549` | Cleanup                            | TH-HARNESS-AC-009, TH-HARNESS-AC-010                    |
| `TH-HARNESS-REQ-550..599` | Platform                           | TH-HARNESS-AC-012                                       |
| `TH-HARNESS-REQ-600..649` | Security and redaction             | TH-HARNESS-AC-011                                       |
| `TH-HARNESS-REQ-650..699` | Product integration                | TH-HARNESS-AC-013, TH-HARNESS-AC-016                    |

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

The following evidence categories remain non-normative in this document unless promoted by a requirement above:

| Evidence category                                                | Current role                                                                      |
| ---------------------------------------------------------------- | --------------------------------------------------------------------------------- |
| Recovery docs under `docs/testing-harness-spec-recovery-docs/**` | Historical traceability and diagnostic context.                                   |
| Raw package scripts and tool output                              | Developer convenience or child-command evidence.                                  |
| Script paths, generated Make include names, helper binaries, `helper_only`, `check_internal`, priority-band names, and generator-only constants | Implementation details unless a requirement above explicitly promotes one. |
| Playwright screenshots, videos, traces, HTML reports             | Diagnostic secret-bearing artifacts.                                              |
| Hosted CI provider workflows                                     | Outside current conformance unless provider source is supplied and later adopted. |
| Visual snapshot refresh process                                  | Future-only unless refresh authority is later adopted.                            |

Exact numeric constants are normative only when they protect security, cleanup safety, bounded output, or deterministic scheduling. Other numeric values in generated manifests, helper names, priority bands, and generator-only constants are implementation details unless this NLSpec gives them a requirement.

The editorial lint for TH-HARNESS-AC-016 rejects the forbidden evidence markers listed in this non-normative section when they appear in Sections 1 through 17. The forbidden markers are: `TODO`, `source_limited`, `source-limited`, `source-observed`, `current code`, `selected evidence`, `recovery evidence`, and `maintainer_decision_required`.

## 19. Future Decisions Outside Current Conformance

The items below are explicitly outside the current conformance profile. They do not block implementation of the current harness contract.

| Future area                                                 | Current treatment                    | Future adoption requirement                                                                  |
| ----------------------------------------------------------- | ------------------------------------ | -------------------------------------------------------------------------------------------- |
| macOS certification                                         | Unsupported for current conformance. | Add platform profile, exact toolchain matrix, and acceptance evidence.                       |
| Windows-native support                                      | Unsupported for current conformance. | Add platform profile separate from WSL2.                                                     |
| Podman/Podman Compose                                       | Unsupported for current conformance. | Add service fixture compatibility profile and cleanup proof.                                 |
| Hosted CI annotations/uploads/artifact-retention dashboards | Provider-neutral `make ci` only.     | Add provider workflow source and provider-specific contract.                                 |
| Visual snapshot refresh authority                           | Validation-only.                     | Add exact refresh command, platform/browser bounds, review path, and golden update criteria. |
| Playwright report/trace/video/screenshot schemas            | Diagnostic-only.                     | Adopt exact Playwright version/schema family or wrapper schema.                              |
| Benchmark-publication harness integration                   | Not part of harness conformance.     | Add Core 05-compatible benchmark manifest and claim-publication profile.                     |
