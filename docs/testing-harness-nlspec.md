# Testing Harness NLSpec

## 1. Status, Authority, and Evidence Basis

Status: draft/proposed. This document is not adopted evidence of product conformance until the repository authority process adopts it.

This NLSpec defines the Cartulary testing harness subsystem. Product behavior remains owned by `docs/spec/00_document_set_status_and_precedence.md` through Core 04. Core 05 owns only claim publication and benchmark reproducibility for claim-bearing timed or fixture-sensitive statements. This NLSpec owns harness mechanics only: command invocation, orchestration, fixtures, services, artifacts, summaries, cleanup, and verification gates.

Authority and evidence are separated as follows.

| Evidence store | Contract role in this NLSpec |
|---|---|
| Core product specs | Own product behavior, conformance scope, and claim publication boundaries. |
| `docs/domain.md` | Defines vocabulary and owner navigation. It does not add harness runtime behavior. |
| Implementation and testing guides | Provide implementation-support evidence. They do not independently own behavior. |
| Recovery docs under `docs/testing-harness-spec-recovery-docs/**` | Provide recovery evidence, source limits, selected runtime evidence, and maintainer decisions. They are not a standalone behavior owner. |
| Root `Makefile` and `tools/task_surface_manifest.json` | Define the observed public Make command surface. Make is the canonical harness command surface. |
| Generated task, schedule, and code artifacts | Provide downstream execution evidence. They do not own behavior unless this repository explicitly marks them as source inputs. |
| Code and tests | Provide current implementation and expectation evidence. Code behavior does not override product specs or unresolved maintainer decisions. |
| CI provider configuration | Source-limited. No `.github/**` workflow source is present in the inspected tree. |
| Direct package scripts | Developer conveniences or child commands unless invoked through Make-owned wrappers. They do not inherit Make-level result-root, run-ID, scheduler, cleanup, or machine-output guarantees. |

Generated files under `internal/gen/**`, `packages/protocol-ts/src/generated/**`, and generated task/schedule artifacts must not be hand-edited. Product-facing tests may produce evidence for product conformance, but this NLSpec defines only how that evidence is run, classified, retained, and reported.

## 2. Purpose and Non-Goals

The testing harness exists to guarantee a reproducible repository command surface for local developers, CI entrypoints, coding agents, and release verification. It must provide deterministic target selection, bounded output, structured artifacts, explicit service ownership, controlled fixture lifecycle, stable failure classification, and destructive cleanup gates.

The harness must:

- expose the canonical public test and verification surface through Make;
- isolate harness operational failures from product assertion failures;
- retain artifacts under explicit result roots and run IDs;
- preserve scheduler, service, fixture, and cleanup provenance;
- provide machine-readable summaries where the command declares such output;
- prevent generated artifacts and retained artifacts from silently becoming behavior owners.

Non-goals:

- The harness must not define product behavior owned by Core 00 through Core 04.
- The harness must not define provider-specific CI behavior without provider source.
- Direct package scripts must not be treated as equivalent to canonical Make targets unless they re-enter a Make-owned wrapper.
- Retained artifacts from an unspecified or newest run must not be treated as current truth.
- Visual snapshot refresh, benchmark publication, public timed claims, and release readiness must not be claimed unless the evidence and owner decisions close those contracts.
- Logical scheduler lanes must not be represented as physical host, Docker, database, object-store, browser, CPU, or network guarantees.

## 3. Terminology

| Term | Meaning |
|---|---|
| harness run | One invocation of a canonical harness command or one child invocation explicitly tied to a result root and run ID. |
| target | A named Make target or scheduler target selected by the harness. |
| child target | A target invoked by an aggregate, sequence, scheduler, or wrapper target. |
| work unit | One scheduler-visible executable unit with an identity, dependencies, resource claims, logs, status, and optional completion keys. |
| scheduler | A harness runner that executes a manifest-defined DAG using logical resource claims and emits scheduler events and summaries. |
| result root | The root directory that contains run artifacts. The Make default is `.cartulary/test-results`. |
| run ID | The run directory name under the result root. The Make default is a UTC timestamp plus PID. |
| artifact | A file or directory produced by a harness run, child command, service, scheduler, test runner, or diagnostic tool. |
| retained artifact | An artifact preserved after command exit for a specific result root and run ID. |
| generated artifact | A file produced from owner inputs by a generator and checked for drift. |
| fixture | Test setup state created for a test, package, target, scheduler group, browser stack, or service suite. |
| service-backed fixture | A fixture that uses Postgres, MinIO, Docker/testcontainers, browser processes, or Compose-backed services. |
| backing services | Postgres, MinIO, Docker/testcontainers, Compose services, backend processes, frontend processes, and browser runtime dependencies used by harness targets. |
| output mode | The mode selected by `CARTULARY_OUTPUT_MODE`, `VERBOSE`, or `CI_VERBOSE` that controls stdout/stderr and artifact summaries. |
| machine output | A machine-readable output mode that emits one parseable top-level JSON object or one parseable JSON pointer where the command supports that contract. |
| failure class | A normalized classification of why a harness command failed. |
| failure origin | The owner area of a failure: product, infrastructure, helper, timing, artifact, source-limited, or unknown. |
| cleanup tier | A named cleanup scope such as repo-local clean, repo-local distclean, service-suite cleanup, browser-stack cleanup, or stale janitor cleanup. |
| stale janitor | A cleanup routine that removes previously generated DBs, buckets, containers, or browser fixtures only when proof predicates match. |
| source-limited claim | A statement where source or selected runtime evidence is insufficient for a final stable contract. |
| owner-required decision | A behavior that needs a maintainer or spec-owner decision before becoming normative. |

Domain and product terms keep their meanings from the product specs and `docs/domain.md`.

## 4. Subsystem Boundaries

The harness owns command mechanics, execution orchestration, structured evidence, runtime fixture lifecycle, and cleanup predicates. It does not own workbook, record, security, auth, collaboration, or view behavior except as test evidence.

| Surface | In scope | Owner evidence | Canonicality | Stability |
|---|---|---|---|---|
| Command invocation | Public Make targets, Make variables, wrapper usage, exit status, output mode | `Makefile`, `tools/task_surface_manifest.json`, generated Make include | Canonical only through Make | Stable for public Make targets; helper internals diagnostic |
| Orchestration and scheduler | Check scheduler, service-backed scheduler, sequence runner, phase-slice runners | scheduler scripts, manifests, scheduler tests | Canonical when invoked by public Make targets | Stable for declared schemas; source-limited for unobserved runtime failures |
| Service lifecycle | testservices, Postgres, MinIO, Compose local services, browser owned stack | service helpers, browser scripts, recovery runtime evidence | Canonical for Make targets using them | Stable where source and selected evidence agree; otherwise source-limited |
| Artifacts and diagnostics | tool/run/target/phase/scheduler summaries, logs, service files, reports | output libraries, schemas, manifest notes | Canonical for Make-owned summaries | Stable for declared schema IDs; diagnostic for tool-defined reports |
| Cleanup and destructive safety | `clean`, `distclean`, suite cleanup, stale janitors, browser cleanup | Make cleanup macros, testservices cleanup, cleanup evidence | Canonical only through Make or owned helper paths | Stable for repo-local path guards; source-limited for parent-death and active DB cleanup |
| CI and release-check | `make check`, `make ci`, `make release-check`, `scripts/ci/verify.sh` | Make sequences, check schedule, CI script | Provider-neutral canonical Make surface | Provider-specific behavior source-limited |
| Direct scripts and package tools | Raw scripts, pnpm scripts, Playwright/Vitest/Biome/Vite direct commands | package manifests, script source | Non-canonical unless called by Make wrappers | Tool-owned or developer convenience |

## 5. Canonical Command Registry

The canonical command registry is closed over public targets in `tools/task_surface_manifest.json` with `classification="public"`. A public target is canonical only when invoked as `make <target>` from the repository root or through a Make-owned wrapper that preserves the target identity.

### Command Family Defaults

| Family | Commands | Required inputs | Optional inputs and defaults | Output modes | Scheduler use | Backing services | Result-root behavior | Artifacts | Exit and failure contract |
|---|---|---|---|---|---|---|---|---|---|
| help and discovery | `help`, `help-all`, `task-surface-report`, `task-guide`, `target-plan`, `target-plan-json`, `fixture-report`, `explain-run`, `explain-phase`, `explain-target` | Command-specific Make variables such as `ROLE`, `PHASE`, `TARGET`, `RESULTS_DIR`, `RUN_ID`, `DETAIL`, `JSON` | Omitted optional inputs select documented summary views; invalid required inputs fail before child work | Human summary; `target-plan-json` is machine stdout | No scheduler execution | No service startup | May read existing artifacts; does not create central run evidence unless wrapped by a tool summary | Human output or command-specific JSON | Usage/config errors exit non-zero; no product assertion class |
| bootstrap and toolchain | `doctor`, `bootstrap`, `bootstrap-node-runtime`, `frontend-toolchain`, `frontend-install`, `playwright-install` | Required local tools according to target | Tool paths default from Make variables | Summary, ci, verbose, debug, machine where wrapper supports tool summary | No scheduler execution | May download/install repo-local tools; no test backing service | Creates target artifacts under run ID when centralized wrapper is used | `cartulary.tool_run_summary.v1` for summary targets | Tool/config failures are `configuration_error` or `preflight_error` |
| local services and dev | `db-up`, `db-reset`, `services-up`, `minio-init`, `dev` | Docker Compose and local config | `CONFIG_FILE=configs/dev/config.toml`, `MINIO_BUCKET=cartulary`, local ports from scripts | Service summary; `dev` is interactive raw | No scheduler execution | Compose Postgres/MinIO and local backend/frontend processes | Service commands may produce summaries; `dev` is interactive and not a verification artifact contract | Service logs and summaries where wrapper emits them | Service startup/readiness/config failures are harness operational failures |
| generated and drift | `generate`, `generate-drift`, `generated-artifact-policy-check`, `json-shape-check`, `toolchain-drift`, `migration-drift`, `phase-ledgers`, `phase-ledger-drift`, `phase-schedules`, `phase-schedule-drift`, `agent-finalize`, `benchmark-claim-check` | Owner inputs and manifests | `RESULTS_DIR` only where the command reads retained evidence; `agent-finalize` may receive `RESULTS_DIR` | Summary/machine where wrapper supports tool summary | No general scheduler, except child targets declared by Make | Migration drift may use scratch Postgres where scheduled by check | Produces summaries under result root when centralized | Tool summary and command-specific drift files | Drift mismatch is `manifest_or_accounting_mismatch` or `artifact` |
| phase and service slices | `phase-slice`, `service-backed-slice` | `PHASE=phaseN` where required | `JSON` for machine-style planning output where supported | Scheduler summary with artifacts | Uses phase selection and service-backed scheduler when applicable | Service-backed slice requires backing services when phase has service work | Creates target-owned summaries and scheduler artifacts | Tool, target, run, scheduler, phase artifacts | Missing/invalid phase is `usage_error`; child failures retain child class |
| backend and frontend leaf tests | `backend-unit`, `backend-store`, `backend-integration`, `backend-process`, `frontend-typecheck`, `frontend-unit`, `frontend-import-boundary-check` | Toolchain, manifests, package inputs | Parallelism and worker variables from Make | Summary/machine through wrappers | Service-backed backend targets may use service-backed scheduler or testservices | Store/integration/process require Postgres/MinIO when service-backed | Creates target artifacts under run ID | Phase, target, tool, logs, reports | Test assertion failures are `test_assertion_failure`; harness setup failures are operational |
| browser E2E | `browser-e2e`, `browser-e2e-webserver-backed`, `browser-e2e-stateful`, `browser-e2e-measurement`, `browser-e2e-visual` | Node/pnpm, Playwright browser, backend/migrate/server support, services | `PLAYWRIGHT_WORKERS=3`, `BROWSER_E2E_FUNCTIONAL_SHARDS=auto` from Make | Summary/machine through wrappers | Uses browser batch and service-backed scheduler where declared | Requires Postgres, MinIO, backend, frontend, browser runtime | Creates browser stack, Playwright, reset, target, and scheduler artifacts | Stack JSON/env, logs, reports, traces/screenshots/videos where Playwright emits them | Product assertions are `test`; stack/readiness/reset failures are operational |
| aggregates and gates | `test-fast`, `test`, `lint`, `check`, `ci`, `release-check`, `build` | Toolchain and all child inputs | Output mode default `summary`; `ci` defaults to CI output when called through `scripts/ci/verify.sh` unless caller sets mode | Summary/ci/verbose/debug/machine where supported | `check` uses check scheduler; `test` and `test-fast` use sequences and service-backed scheduler | Service-backed and browser children require backing services | Creates aggregate run summaries and child artifacts | Run summary, target summaries, scheduler summaries, tool summary | Exit non-zero if any required child fails or required artifact is missing |
| static analysis and security | `lint-biome`, `lint-scripts`, `lint-shell`, `go-vulncheck`, `go-gosec-targeted`, `go-gosec-audit` | Toolchain and source roots | Rule and pattern variables from Make; `LINT_SHELL_STRICT=1` makes direct shell lint blocking | Summary/machine through wrappers | Scheduled inside `check`; direct target is non-scheduler | No backing services | Creates tool summaries and logs | Tool summary and tool logs | Tool findings are harness gate failures unless warning-only by target definition |
| builds | `build-server`, `build-migrate`, `build-web` | Build inputs and toolchain | Output paths from Make variables | Summary/machine through wrappers | Scheduled as readiness work inside `check` when needed | No backing services | Produces binaries or web dist artifacts | Tool summary and build logs | Build failures are operational or product-code compile failures surfaced as gate failures |
| cleanup | `clean`, `distclean` | None | Uses Make path registries | Human destructive output | No scheduler execution | Does not stop Docker Compose services | Deletes only registered repo-local paths | No central summary contract | Unsafe path guard failure exits non-zero; missing paths are not failures |
| formatting | `format` | Toolchain | None | Summary/machine through wrapper | No scheduler execution | No backing services | May rewrite authored files | Tool summary and formatter logs | Formatter failure is operational; formatter rewrites are mutating |

### Public Target Registry

Every command below inherits the matching family defaults. `Included in` is the manifest-declared inclusion set.

| Target | Included in | Output class | Stable summary schema | Notes |
|---|---|---|---|---|
| `help` | helper_only | human_summary | none | Compact public command help. |
| `help-all` | helper_only | human_summary | none | Exhaustive public help. |
| `doctor` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Tool readiness check. |
| `bootstrap` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Installs pinned tooling and dependencies. |
| `bootstrap-node-runtime` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Installs repo-local Node runtime. |
| `frontend-toolchain` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Prepares Node/pnpm toolchain. |
| `frontend-install` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Installs workspace dependencies. |
| `playwright-install` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Installs Playwright browser runtime. |
| `db-up` | helper_only | service_summary | `cartulary.tool_run_summary.v1` | Starts local Postgres. |
| `db-reset` | helper_only | service_summary | `cartulary.tool_run_summary.v1` | Resets local database only; object storage is not reset by this target. |
| `services-up` | helper_only | service_summary | `cartulary.tool_run_summary.v1` | Starts local backing services. |
| `minio-init` | helper_only | service_summary | `cartulary.tool_run_summary.v1` | Initializes local MinIO bucket. |
| `dev` | helper_only | interactive_raw | none | Local interactive dev stack. |
| `generate` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Runs generation commands. |
| `generate-drift` | check | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Checks generated drift. |
| `generated-artifact-policy-check` | check | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Checks generated artifact policy. |
| `json-shape-check` | check | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Checks JSON manifest shapes. |
| `toolchain-drift` | check | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Checks toolchain pins. |
| `migration-drift` | check | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Checks migrations. |
| `phase-ledgers` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Regenerates phase ledgers. |
| `phase-ledger-drift` | check | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Checks phase ledger drift. |
| `phase-schedules` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Regenerates phase schedules. |
| `phase-schedule-drift` | check | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Checks schedule drift. |
| `agent-finalize` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | End-of-run maintenance surface. |
| `benchmark-claim-check` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Checks benchmark claim artifacts. |
| `task-surface-report` | helper_only | human_summary | none | Prints target surface report. |
| `task-guide` | helper_only | human_summary | none | Prints role/phase guidance. |
| `phase-slice` | helper_only | scheduler_summary_with_artifacts | `cartulary.tool_run_summary.v1` | Runs selected phase target slice. |
| `service-backed-slice` | helper_only | scheduler_summary_with_artifacts | `cartulary.tool_run_summary.v1` | Runs selected phase service-backed slice or explicit no-op. |
| `backend-unit` | test,check,ci | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Pure backend unit evidence. |
| `backend-store` | test,check,ci | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Service-backed store slice. |
| `backend-integration` | test,check,ci | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Service-backed integration slice. |
| `backend-process` | test,check,ci | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Service-backed process slice. |
| `target-plan` | helper_only | human_summary | none | Prints Go target plan. |
| `target-plan-json` | helper_only | machine_stdout | command-specific JSON | Prints Go target plan JSON. |
| `fixture-report` | helper_only | human_summary | `cartulary.fixture_report.v1` where JSON requested | Reads retained fixture artifacts. |
| `explain-run` | helper_only | human_summary | command-specific JSON where JSON requested | Reads retained run artifacts. |
| `explain-phase` | helper_only | human_summary | none | Explains phase manifest. |
| `explain-target` | helper_only | human_summary | none | Explains target plan/artifacts/logs. |
| `go-test-duration-baselines` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Refreshes Go duration baselines from a selected run. |
| `go-test-duration-baseline-coverage` | check | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Checks baseline coverage. |
| `go-test-duration-baseline-drift` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Checks Go duration baseline drift against selected run. |
| `browser-e2e-duration-baselines` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Refreshes browser duration baselines. |
| `browser-e2e-duration-baseline-drift` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Checks browser duration drift. |
| `service-backed-make-target-duration-baselines` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Refreshes service-backed target durations. |
| `service-backed-make-target-duration-baseline-drift` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Checks service-backed target duration drift. |
| `harness-smoke-duration-baselines` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Refreshes harness smoke durations. |
| `harness-smoke-duration-baseline-drift` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Checks harness smoke duration drift. |
| `scheduler-event-order-drift` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Checks scheduler event ordering. |
| `scheduler-summary-timing-drift` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Checks scheduler summary timing. |
| `frontend-typecheck` | test,check,ci | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Frontend TypeScript check. |
| `frontend-unit` | test,check,ci | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Frontend unit suite. |
| `frontend-import-boundary-check` | check | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Frontend import boundary check. |
| `lint-biome` | check | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Frontend authored-source lint. |
| `lint-scripts` | check | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Script lint. |
| `lint-shell` | check | summary_with_artifacts | `cartulary.tool_run_summary.v1` | ShellCheck wrapper; direct target is warning-only unless strict env is set. |
| `format` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Mutating formatter target. |
| `browser-e2e` | test,check,ci | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Isolated browser E2E aggregate. |
| `browser-e2e-webserver-backed` | test,check,ci | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Shared-stack browser stage. |
| `browser-e2e-stateful` | test,check,ci | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Stateful browser stage. |
| `browser-e2e-measurement` | test,check,ci | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Measurement browser stage. |
| `browser-e2e-visual` | test,check,ci | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Visual validation stage. |
| `test-fast` | test,check,ci | aggregate_summary_with_artifacts | `cartulary.tool_run_summary.v1` | Fast local verification aggregate. |
| `test` | test,check,ci | aggregate_summary_with_artifacts | `cartulary.tool_run_summary.v1` | Full-corpus test aggregate. |
| `lint` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Aggregate lint wrapper. |
| `go-vulncheck` | check | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Govulncheck wrapper. |
| `go-gosec-targeted` | check | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Blocking targeted Gosec wrapper. |
| `go-gosec-audit` | check | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Warning-only audit profile wrapper. |
| `check` | check | scheduler_summary_with_artifacts | `cartulary.tool_run_summary.v1` | Developer gate through check scheduler. |
| `ci` | ci | aggregate_summary_with_artifacts | `cartulary.tool_run_summary.v1` | Provider-neutral CI entrypoint. |
| `release-check` | release-check | aggregate_summary_with_artifacts | `cartulary.tool_run_summary.v1` | Release gate aggregate. |
| `build` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Build aggregate. |
| `build-server` | check,helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Server build. |
| `build-migrate` | check,helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Migration binary build. |
| `build-web` | helper_only | summary_with_artifacts | `cartulary.tool_run_summary.v1` | Web build. |
| `clean` | helper_only | destructive_human | none | Repo-local cleanup. |
| `distclean` | helper_only | destructive_human | none | Repo-local cleanup plus repo-local tool/runtime caches. |

### Direct Script and Package Boundary

| Surface | Classification | Contract |
|---|---|---|
| Root `package.json` scripts `build`, `test`, `typecheck` | Developer convenience | May be used directly by developers. They do not promise Make result roots, run IDs, scheduler summaries, cleanup, or machine output. |
| `apps/web/package.json` scripts | Developer convenience or child command | Browser and unit scripts become harness child work only when invoked through Make wrappers. |
| Raw `scripts/run-*.sh` and `scripts/*.mjs` | Tool-owned diagnostics or child commands | Their usage and exit codes matter to Make wrappers. Direct invocation is not canonical unless a public Make target defines it as the surface. |
| Raw Go, Vitest, Playwright, Biome, Vite, pnpm commands | Tool-owned | Tool output schemas are external or partial. They do not become stable harness schemas by direct use. |

## 6. Configuration and Environment Contract

The harness must document source-observed variables and defaults. Cross-layer precedence is unresolved unless a row names a specific source-proved rule.

| Name or group | Type and valid values | Default when omitted | Reads direct package scripts | Invalid input behavior | Stability |
|---|---|---|---|---|---|
| `CARTULARY_TEST_RESULTS_DIR` | Path | Make: `.cartulary/test-results`; runner context may use same root or command-specific fallback | No Make guarantee | Invalid/unwritable paths fail the command that writes artifacts | Stable for Make; direct scripts source-limited |
| `CARTULARY_TEST_RUN_ID` | Non-empty path segment generated by caller | Make: UTC timestamp plus PID; runner context may use `adhoc` when absent | No Make guarantee | Invalid path behavior is source-limited | Stable for Make; direct scripts source-limited |
| `CARTULARY_OUTPUT_MODE` | `quiet`, `summary`, `ci`, `verbose`, `debug`, `machine`, plus aliases normalized by output library | `summary` | Some direct scripts read it, but no Make guarantee | Unknown modes normalize to summary in tool-output library | Stable where centralized output library is used |
| `VERBOSE`, `CI_VERBOSE` | `1` enables verbose/normal Make behavior | unset | Some wrappers read | Non-`1` is ignored where source checks exact `1` | Stable in Make |
| `CI` | Environment marker | unset; `scripts/ci/verify.sh` sets CI mode behavior through output mode when caller did not set it | Tool-dependent | Source-limited | Source-limited |
| Scheduler capacity envs | Positive integers: `CHECK_HOST_CPU_JOBS`, `CHECK_HOST_IO_JOBS`, `CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT`, `CARTULARY_SERVICE_BACKED_GO_IO_LIMIT`, `CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT` | Registry auto policies or default limits | No | Non-positive or non-integer values fail scheduler config before child work | Stable for scheduler runners |
| CLI scheduler `--resource-limit name=value` | Declared resource and positive integer or allowed bounded value | No CLI override | No | Unknown resource, undeclared resource, or invalid value exits usage/config non-zero | Stable for scheduler runners |
| `GO`, `GO_CACHE_DIR`, `GO_MOD_CACHE_DIR`, `GOCACHE`, `GOMODCACHE` | Paths and Go executable | Go auto-discovered; external caches `/tmp/cartulary-go-build` and `/tmp/cartulary-go-mod` | Direct tool-dependent | Missing Go is configuration failure | Stable for Make; cleanup exclusion stable |
| `NODE_VERSION`, `PNPM_VERSION`, `NODE_RUNTIME_DIR`, `NODE_BIN`, `PNPM`, `COREPACK_HOME`, `PATH` | Versions/paths | Node `24.15.0`, pnpm `10.33.0`, repo-local `tmp/node-runtime` | Direct package scripts may bypass repo-local runtime | Missing/mismatched runtime fails toolchain targets | Stable for Make |
| `TEST_SERVICES_BIN`, `CARTULARY_TEST_SERVICES_BIN` | Executable path | `tmp/toolbin/cartulary-test-services` | No | Missing executable fails wrapper or browser stack config | Stable for Make-owned service targets |
| Testservices suite env | `CARTULARY_TEST_SERVICES_ACTIVE=1`, `CARTULARY_TEST_SUITE_ID`, `CARTULARY_TEST_TARGET` | Suite wrapper generates active env and 12-byte-hex suite ID | No | Missing active attach values cause owned mode or config failure depending helper | Stable for testservices |
| Postgres attach env | `CARTULARY_PGTEST_ADMIN_DSN`, `CARTULARY_PGTEST_DSN_TEMPLATE` containing `{database}`, `CARTULARY_PGTEST_TEMPLATE_DB` | none | No | Missing pair or malformed template fails fixture setup | Stable for pgtest attach |
| Postgres fixture policy env | `CARTULARY_POSTGRES_FIXTURE_POLICY_TESTS`, `...PACKAGES`, `...DEFAULT`, reset table lists | Helper defaults | No | Unknown policies or table closure errors fail fixture setup | TODO: maintainer_decision_required - exact public precedence and allowed policy matrix |
| MinIO attach env | `CARTULARY_S3TEST_ENDPOINT`, `CARTULARY_S3TEST_ACCESS_KEY_ID`, `CARTULARY_S3TEST_SECRET_ACCESS_KEY`, `CARTULARY_S3TEST_SECURE` bool | none; secure blank means false | No | Missing partial set or invalid bool fails fixture setup | Stable for s3test attach |
| `CARTULARY_TEST_SERVICES_WEB_E2E_CLEANUP_WORKERS` | Integer 1..16 | 4 | No | Invalid or `<1` falls back to 4; values over 16 clamp to 16 | Stable source-observed |
| Compose env | `CARTULARY_COMPOSE_FILE`, ready timeouts, `MINIO_BUCKET` | `docker-compose.dev.yml`, Postgres 180s, MinIO 120s, bucket `cartulary` | No | Missing Docker/Compose or failed readiness exits non-zero | Stable for local service scripts |
| `CONFIG_FILE`, `CARTULARY_CONFIG_FILE` | Config file path | `configs/dev/config.toml` | Direct app-dependent | Missing/invalid config fails startup/migration | Source-observed; precedence unresolved |
| Browser owned-stack env | Stack artifact/runtime roots, API/public origins, backend/frontend port overrides | Dynamic ports; default public origin `http://127.0.0.1:4173` in shared Playwright config | Playwright reads subset | Configured port in use fails before stack start | Stable source-observed |
| Browser stale-binary guard | `CARTULARY_WEB_E2E_USE_REPO_ROOT_BINARIES`, `CARTULARY_SERVER_BIN`, `CARTULARY_MIGRATE_BIN` | Make sets opt-in for owned stack | No | Missing binary or build prerequisite failure exits non-zero | Stable source-observed |
| Playwright shared state env | `CARTULARY_PLAYWRIGHT_EXTERNAL_SERVER=1`, state dir, worker count/index offset, `PLAYWRIGHT_WORKERS` | Make workers 3; shared config fallback 2; offset 0 | Yes | Invalid worker offset throws config error | Stable source-observed |
| Webserver-backed shard env | Functional/support grep and file envs | none | Playwright config reads | Missing any required value throws config error | Stable source-observed |
| `CARTULARY_ENABLE_TEST_ROUTES` | Exact `1` enables test routes | disabled | No | Other values do not enable routes | Stable source-observed |
| Object store runtime env | `CARTULARY_S3_OBJECT_PRIMARY_*` endpoint, credentials, secure, bucket | Browser/dev local defaults to MinIO local credentials and `cartulary` bucket | App runtime reads | Config/connection failure exits runtime setup/reset | TODO: maintainer_decision_required - precedence with config file |
| Runtime root envs | `CARTULARY__ROOTS__*__PATH` | Browser stack creates paths under runtime root | App runtime reads | Invalid/unwritable paths fail runtime startup | Stable source-observed |
| Harness scratch env | `CARTULARY_HARNESS_REPO_ROOT`, `CARTULARY_HARNESS_SCRATCH_ROOT`, `TMPDIR` | `${TMPDIR:-/tmp}/cartulary-harness-scratch` outside repo | No | In-repo scratch root is rejected | Stable source-observed |
| Cleanup controls | `LINT_SHELL_STRICT`, Make cleanup path registries | Strict shell lint unset; cleanup paths from Make | No | Rejected cleanup path fails before deletion | Stable for Make cleanup |

TODO: maintainer_decision_required - Define precedence across CLI flags, Make variables, exported environment variables, generated manifests, config files, package scripts, and hardcoded defaults for public harness configuration.

## 7. Run Lifecycle

A harness run starts when a canonical Make target or Make-owned wrapper resolves a target name and run context. Direct package scripts do not start a canonical harness run unless they are invoked as child work by a canonical Make target.

Lifecycle:

1. Resolve command inputs, Make variables, output mode, result root, and run ID.
2. Validate required command arguments and manifest/schema inputs.
3. Run preflight checks for tools, generated artifacts, Docker, Compose, browser, or service dependencies as required by the target.
4. Provision backing services or attach to existing harness-owned suite services where the target requires them.
5. Run generation or drift checks when the target is a generation/drift command.
6. Execute sequence, scheduler, or leaf child work.
7. Emit phase, target, run, scheduler, service, and tool artifacts according to the target family.
8. Attempt required teardown and cleanup for resources owned by the target.
9. Emit final human or machine summary.
10. Exit with the command status after accounting for child status, finalizer failures, missing artifacts, and cleanup failures.

| Status kind | Values | Transition rules |
|---|---|---|
| Run status | `pass`, `fail`, `cancelled`, `source_limited` | `pass` only when every required child and finalizer passes and required artifacts exist. `fail` when any required child, scheduler, setup, artifact, or cleanup step fails. `cancelled` when signal/cancellation evidence is explicit. `source_limited` is documentation status, not a runtime success. |
| Target status | `pass`, `fail`, `skipped` | Child targets skipped by dependency failure retain skip provenance and must not be reported as missing artifacts. |
| Work-unit status | `pending`, `running`, `passed`, `failed`, `skipped` | Scheduler owns transitions. A skipped unit records `dependency_failure` or `schedule_stopped_after_failure`. |
| Failure class | See Section 11 | Classification must use the most specific available class. |
| Failure origin | `product`, `infrastructure`, `helper`, `timing`, `artifact`, `unknown` | Product assertion failures must remain separate from harness operational failures. |
| Retryability | `no`, `conditional`, `unknown`, `not_applicable` | Retryability is diagnostic only unless a command defines a retry policy. |
| Cleanup outcome | `not_needed`, `completed`, `failed`, `deferred`, `source_limited` | Deferred cleanup must include lease or artifact evidence when source provides it. |

Interruption and timeout handling are source-limited except for selected paths recorded in recovery evidence. Parent-death cleanup is not a stable contract.

## 8. Scheduler and Resource Model

The check scheduler consumes `cartulary.check_schedule.v11`. The service-backed scheduler consumes `cartulary.service_backed_schedule.v10`. Scheduler resources are declared by `cartulary.scheduler_resource_registry.v3`.

Work-unit identity must be stable within one schedule. Work-unit dependency keys must refer to completion keys. Unknown schemas, duplicate IDs, invalid dependency shapes, invalid resource claims, and undeclared resource limits are configuration failures.

Default logical resources:

| Resource | Scheduler | Default or auto policy |
|---|---|---|
| `host_cpu` | check | Auto check host CPU; env override `CHECK_HOST_CPU_JOBS`. |
| `host_io` | check | Auto check host IO; env override `CHECK_HOST_IO_JOBS`. |
| `suite_service_stack` | check | 1. |
| `migration_scratch_postgres` | check | 1. |
| `go_cpu` | service-backed | Auto service-backed Go CPU; env override `CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT`. |
| `go_io` | service-backed | Auto service-backed Go IO; env override `CARTULARY_SERVICE_BACKED_GO_IO_LIMIT`. |
| `browser_stack` | check, service-backed | Auto browser stack; env override `CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT`. |
| `process` | check, service-backed | 6. |
| `postgres` | check, service-backed | 32. |
| `minio` | check, service-backed | 32. |
| `postgres_reset` | check, service-backed | 1. |
| `browser_stage_*` | check, service-backed | Generated per browser stage, usually 1. |

Scheduler events must be JSONL with monotonically increasing sequence numbers. Event monotonic time must not move backward; wall-clock regressions may appear only with explicit clock-skew diagnostics.

Pseudocode:

```text
validate manifest schema, resources, work units, dependencies, and limits
pending = work units in manifest order
running = empty
completedKeys = empty
failedKeys = empty
activeClaims = empty
emit scheduler-start

while pending or running:
  while scheduler not stopped:
    candidate = first pending unit that:
      has no failed dependency
      has all dependencies in completedKeys
      has enough resource capacity
      does not claim any resource reserved by an earlier ready but resource-blocked unit
    if no candidate:
      break
    remove candidate from pending
    add candidate resource claims to activeClaims
    start candidate, emit unit-start
    add candidate to running

  if running is empty and pending is not empty and scheduler not stopped:
    fail scheduler as deadlock/configuration error

  wait for the next running unit to finish or for a progress tick
  on progress tick:
    emit bounded progress and blocked-work diagnostics

  on unit finish:
    remove resource claims unless retained by the unit
    record log path, duration, status, completion keys, and failure keys
    if status is success or unit has complete-on-failure:
      add completion keys
    if status is failure:
      add failure keys
      record first failure
      if stopOnFirstFailure:
        stop scheduling new work
      else:
        skip pending units with failed dependencies as dependency_failure

  if stop scheduling and running is empty:
    skip remaining pending units as dependency_failure or schedule_stopped_after_failure

run finalizers in manifest-defined scheduler order
release retained resource claims
emit scheduler summary, progress summary, and scheduler-complete
validate summary timing unless explicitly disabled
exit with first failure or finalizer failure status
```

The check scheduler stops scheduling after the first failure and drains running work. The service-backed scheduler continues independent work and skips dependency-failed units. Summary ordering must be deterministic: work units by scheduler record order where source defines it, artifact lists by stable sorting, resources by registry display order or lexicographic fallback.

## 9. Services and Fixture Lifecycle

Service-backed targets must use `tools/testservices` or a Make-owned service wrapper unless explicitly classified as local-dev services. The harness may start Docker/testcontainers containers, Docker Compose services, backend processes, frontend processes, and browser runtimes. Docker itself is an external service; the harness does not own the Docker daemon.

| Service or fixture | Provision | Ready condition | Reset | Teardown and cleanup | Failure behavior |
|---|---|---|---|---|---|
| testservices suite | `testservices run` creates a suite unless `CARTULARY_TEST_SERVICES_ACTIVE=1`; `start-suite` writes env and lease files | Docker preflight, Postgres ready, template DB migrated, MinIO ready, lease written, stale browser janitor complete | Delegated to DB/bucket/browser/reset helpers | Deferred cleanup after child exit; detached reaper may handle containers through lease | Usage errors exit 2; setup/cleanup errors exit non-zero and record service diagnostics |
| Suite Postgres | Testcontainers `postgres:16-alpine` with local test creds and Cartulary labels | Mapped port then admin DSN ping | DB clone, reset table truncation, transaction rollback, or app reset route | Container termination by suite cleanup or reaper | Startup attempts may retry source-observed transient failures |
| Template DB | `ct_tpl_<suite-hash>` migrated by suite setup | Migrations applied, connections terminated, `ALLOW_CONNECTIONS=false`, `IS_TEMPLATE=true` | None; children clone from it | Removed with owning container | Template failure prevents child execution |
| Go DB fixtures | `ct_<suite>_<process>_<counter>_<suffix>` | Clone/create/migrate/reset succeeds | Per policy: clone, package reset, transaction, group clone | `t.Cleanup` drops or retains according to active suite mode | Fixture setup/reset/drop failure fails the test or helper |
| Suite MinIO | Testcontainers MinIO image with local test creds | Authenticated `ListBuckets` succeeds | Bucket/prefix clear or app reset route | Container termination by suite cleanup or reaper | Auth/config errors are non-retryable; readiness timeouts may retry |
| Buckets/prefixes | Generated bucket names `ct-...` | Create or existing bucket verified | Object list/delete by prefix or full bucket where owned | Per-test/browser bucket cleanup; package buckets may persist for reuse | Partial cleanup records diagnostics when surfaced |
| Compose local services | `scripts/dev-services.sh` starts Postgres/MinIO from `docker-compose.dev.yml` | `pg_isready` and MinIO readiness probes | `db-reset` recreates only local DB; object store is not reset | No canonical Compose down target in this NLSpec | Compose/readiness failure exits non-zero and may print recent logs |
| Browser backend | `scripts/start-web-e2e.sh` starts server with test routes enabled | `/readyz` responds within configured loop | Reset route | Process group TERM, wait, KILL, port-release wait | Early exit/readiness failure retains backend log |
| Browser frontend | Vite dev server through repo-local pnpm | Public origin responds | Playwright state reset clears state dir where configured | Process group TERM, wait, KILL, port-release wait | Early exit/readiness failure retains web log |

### Test-Only Reset Route

The current harness-owned reset route is `POST /api/v1/test/runtime/reset`. It exists only when the server is started with `CARTULARY_ENABLE_TEST_ROUTES=1` and the server wires test route registrars. The default runtime must not expose the route.

Request schema: the route accepts a POST request with no required body. A JSON body may be sent by callers, but no request fields are required by current code.

Success response: HTTP 200 with a standard success envelope whose `data.schema_id` is `cartulary.test.runtime_reset.v1`. Required `data` fields are `reset_id`, `tables_reset`, `mutable_table_count`, `object_count_removed`, `object_count_after`, `migration_metadata_preserved`, `bootstrap_admin_restored`, and `post_reset_counts`. `post_reset_counts` contains `active_deployment_admins`, `bootstrap_markers`, `incidents`, `records`, `user_sessions`, and `route_idempotency`.

Failure response: HTTP 500 with error code `test_runtime_reset_failed`, the failed action, and an `error` string in metadata.

Algorithm:

1. Count `goose_db_version`.
2. List public base tables except `goose_db_version` ordered by table name.
3. Truncate those tables with `RESTART IDENTITY CASCADE`.
4. Run bootstrap preflight to restore the deployment admin.
5. List and delete all objects in the configured object bucket.
6. Count remaining objects and migration metadata.
7. Count post-reset state fields.
8. Return success with a new reset ID.

Security boundary: test routes are disabled unless `CARTULARY_ENABLE_TEST_ROUTES=1`. The browser owned stack sets that variable for harness runs; the local dev stack tests assert the variable is absent from ordinary dev backend env.

TODO: maintainer_decision_required - Define reset route concurrency semantics, request timeout, partial-failure rollback expectations, and readiness guarantees.

## 10. Artifacts, Results, and Schemas

Artifacts must be tied to a result root, run ID, target, and producer whenever the target declares centralized artifacts. Retained artifacts prove only the specific run identified by explicit result root and run ID.

| Artifact family | Producer | Path under result root | Schema/status | Required ordering and nullability | Retention and cleanup |
|---|---|---|---|---|---|
| Tool-run summary | `scripts/lib/tool-output.mjs` and centralized wrappers | `<run>/<target>/tool-run-summary.json` or command-specific target dir | Stable `cartulary.tool_run_summary.v1` | Required fields include schema ID, target, command, status, exit code, timestamps, output mode, artifact root, artifacts, work/evidence/helper accounting, counts, failure class/origin, failures, slowest, warnings, rerun commands, extensions | Retained under result root; removed by `make clean` when under default result root |
| Phase summary | `scripts/lib/test-output/cli.mjs` phase handlers | Target phase dirs | Stable `cartulary.test_phase_summary.v3` | Includes label, target, runner, status, command, timing, counts, artifacts, owners, inventory, and failure fields | Retained under result root |
| Target summary | `test-output.mjs target-summary` | `<run>/<target>/target-summary.json` | Stable `cartulary.test_target_summary.v4` | Includes own/child/totals rollups, artifacts, failures, accounting, scheduler timing, and status | Retained under result root |
| Run summary | `test-output.mjs run-summary`, sequence/check wrappers | `<run>/run-summary.json` or aggregate dir | Stable `cartulary.test_run_summary.v6` | Includes work units, evidence targets, summary groups, helper units, shared groups, failures, fixture summary, artifact dirs | Retained under result root |
| Scheduler summary | Check/service-backed scheduler | `<run>/<target>/scheduler-summary.json` | Stable `cartulary.check_scheduler_summary.v9` or `cartulary.service_backed_scheduler_summary.v9` | Work and artifact ordering must be deterministic; summary timing must validate unless disabled | Retained under result root |
| Scheduler event stream | Check/service-backed scheduler | `<run>/<target>/scheduler-events.jsonl` | Stable event schema IDs v5 | Sequence strictly increases; monotonic time non-decreasing; clock skew explicit | Retained under result root |
| Scheduler progress summary | Scheduler reporter | `<run>/<target>/progress-summary.log` | Human diagnostic | Bounded progress snapshots | Retained under result root |
| Service scope artifacts | suiteservices and testservices | `_shared/test-services/<suite-id>/...` | Lease has `cartulary.test_services.lease.v1`; service events/scope summary source-observed without full schema ID | Service events record typed fields; exact closure source-limited | Retained under result root; suite cleanup may add final diagnostics |
| Generated manifest summaries | Generation/drift scripts | Tool-specific target dirs | Manifest schema IDs declared in generated JSON | Unknown fields rejected where JSON-shape tools enforce closure | Generated files remain checked-in; summaries retained |
| Browser stack metadata | `start-web-e2e.sh` | Browser target support dir | `cartulary.web_e2e_stack.v1`; stack env file | Contains origins, ports, runtime root, server log, web log paths | Retained for browser target |
| Browser session lease | Browser stack/session scripts | Browser support dir | `cartulary.web_e2e_session_lease.v1` | Source-observed | Retained for browser target |
| Reset response/status/state | `reset-web-e2e-stack.sh` and reset route | `reset-boundary/*.json`, `*.status`, `*.state-reset` | Reset data `cartulary.test.runtime_reset.v1` inside success envelope | Wrapper validates schema ID, reset ID, table list, migration/admin flags, object count, post-reset counts | Retained for browser target |
| Logs | Shell, Go, scheduler, browser, service wrappers | Target log dirs and tool-specific report dirs | Contents schema unknown unless a producer declares JSON | Logs are diagnostic; empty logs may be omitted by source | Retained unless cleanup removes result root |
| Coverage reports | Go/frontend/test tools | Tool-specific coverage paths | Tool-defined | Source-limited unless command declares exact schema | Removed by `make clean` when under registered paths |
| Screenshots, videos, traces, Playwright reports | Playwright | Playwright report/test-results dirs | TODO: source_limited - Playwright tool schemas are not adopted by this NLSpec | Tool-owned ordering/nullability | Removed by `make clean` when under registered paths |
| Visual snapshots and goldens | Browser/fixture tools | Source and tool-specific locations | TODO: maintainer_decision_required - refresh authority and OS/browser/version bounds | Validation-only until owner decision | Must not be refreshed by undocumented commands |

Machine output: centralized top-level commands that support `CARTULARY_OUTPUT_MODE=machine` must emit exactly one parseable `cartulary.tool_run_summary.v1` JSON object or one parseable JSON pointer, and must not stream child logs, progress prose, or duplicate run/target JSON objects. TODO: source_limited - Commands outside the centralized output library and direct package scripts do not have a complete machine-output contract in this NLSpec.

## 11. Failure, Error, and Exit-Code Contract

The harness uses a specific taxonomy for specification and maps it to the current summary classes where required by existing schemas.

| Failure class | Trigger | Detection rule | Retryability | Summary mapping | Exit-code mapping | Cleanup expectation | Accepted conformance evidence |
|---|---|---|---|---|---|---|---|
| `usage_error` | Invalid args or missing required flags | CLI parser rejects invocation before child work | no | helper/infra | Usually 2 where source observes usage exits | No owned resources started | May appear in negative acceptance criteria |
| `configuration_error` | Missing/invalid tool, path, env, config, manifest, resource limit | Setup or schema validation fails | no until fixed | helper/infra/artifact | Usually 1 or 2 depending wrapper | Cleanup any started resources | May appear in negative criteria |
| `preflight_error` | Docker/platform/tool preflight fails before managed services | Preflight command returns non-zero | conditional | infra | Non-zero, often 1 or 2 at Make surface | Record diagnostics; cleanup partial setup | May appear in negative criteria |
| `service_start_error` | Postgres, MinIO, Compose, backend, frontend, browser process fails to start | Start command fails or process exits early | conditional | infra | Non-zero | Stop partial services/processes where source owns them | May appear in operational criteria |
| `service_readiness_timeout` | Started service fails readiness before deadline | Readiness loop exhausts | conditional | infra/timing | Non-zero | Stop started service/process | May appear in operational criteria |
| `fixture_error` | DB/bucket/template/reset/janitor/fixture operation fails | Fixture helper or reset validation fails | conditional | infra/helper/artifact | Non-zero | Record partial fixture state where available | May appear in operational criteria |
| `test_assertion_failure` | Test runner assertion fails after harness setup | Go/Vitest/Playwright/runner non-zero classified as test | no by default | test/product | Child status propagated or summarized | Cleanup harness resources | Accepted as product failure evidence, not harness failure |
| `child_target_failure` | Aggregate child exits non-zero | Sequence/scheduler child status | depends on child class | child class | Aggregate non-zero | Aggregate cleanup/finalizers run | May appear in aggregate negative criteria |
| `scheduler_accounting_error` | Manifest, summary, timing, event, or accounting mismatch | Drift/shape/accounting validator fails | no until fixed | artifact/helper | Non-zero | No service cleanup unless child resources started | May appear in harness criteria |
| `resource_conflict` | Logical lane, port, lock, DB/bucket name, or host resource conflict | Scheduler capacity, port check, lock, or setup check fails | conditional | infra/helper | Non-zero | Cleanup partial resources | May appear in operational criteria |
| `timeout_failure` | Command, readiness, watchdog, cleanup, or lock exceeds deadline | Timeout/watchdog status or loop exhaustion | conditional | timing/infra/helper | Non-zero or timeout wrapper status | Cleanup started resources where possible | May appear in negative criteria |
| `cleanup_error` | Cleanup command/finalizer/leak check/reaper scheduling fails | Cleanup step returns non-zero or required proof missing | conditional | helper/infra/artifact | Non-zero or failed summary | Report remaining resource evidence | May appear in operational criteria |
| `cancelled_or_interrupted` | Signal, cancellation, parent death, abort | Signal/trap/status evidence | unknown | helper/infra | Source-limited and command-specific | Attempt source-defined cleanup | Source-limited except selected evidence |
| `authority_required` | Behavior cannot be normative without owner decision | NLSpec/open-decision review | n/a | n/a | n/a | n/a | Not a runtime pass/fail class |
| `source_limited` | Evidence insufficient for final claim | NLSpec/open-decision review | n/a | n/a | n/a | n/a | May appear only as documented limit |
| `unknown_failure` | Failure cannot be classified | Catch-all exception or unselected artifact | unknown | helper/infra | Non-zero | Record diagnostics if available | Must not appear in accepted conformance evidence except criteria about source-limited diagnostics |

Source summary classes are currently `test`, `infra`, `helper`, `timing`, and `artifact`; failure origins are `product`, `infrastructure`, `helper`, `timing`, and `artifact`.

## 12. Cleanup and Destructive Safety

Cleanup is a destructive contract. A cleanup command must delete only paths or resources that satisfy its exact ownership predicates.

### Repo-Local Cleanup

`make clean` removes registered repo-local artifacts when present:

- `server`, `migrate`;
- `apps/web/dist`;
- embedded web asset stamp and generated embedded web assets under the embedded asset directory while preserving `.keep`;
- `$(CARTULARY_TEST_RESULTS_DIR)`;
- release artifacts;
- `test-results`, `apps/web/test-results`;
- Playwright report dirs;
- coverage dirs;
- `.vite` dirs and Vite caches under root, app, and packages;
- repo-local scratch directories under `tmp/` except preserved names.

`make distclean` removes all `clean` paths plus:

- `tmp/node-runtime`;
- `tmp/node-archives`;
- `tmp/toolbin`;
- shellcheck archives;
- frontend install/toolchain stamps;
- Playwright install cache;
- frontend embed cache;
- `.cache`;
- `.pnpm-store`.

External Go caches `/tmp/cartulary-go-build` and `/tmp/cartulary-go-mod` are outside default `clean` and `distclean`.

Deletion predicates:

- Empty path, `/`, and `.` must be rejected.
- Paths outside the repository prefix must be rejected.
- Existing paths and symlinks may be removed if they pass the repo-prefix guard.
- Missing paths are ignored.
- Registered path iteration order is the Make registry order.
- Shell traversal or path-normalization behavior beyond the current guard is TODO: source_limited.

### Service and Stale Cleanup

Service-suite cleanup may terminate owned containers, remove generated DBs and buckets, record service diagnostics, and schedule a detached lease-backed reaper. The Docker daemon and Compose services are not globally stopped by `clean` or `distclean`.

Stale janitor proof gates:

| Resource | Required proof before deletion |
|---|---|
| Database | Generated Cartulary test DB naming pattern plus suite metadata, lease evidence, or completed-summary evidence, and a conservative age or completed-run predicate. |
| Bucket | Generated Cartulary test bucket naming pattern plus harness metadata/lease evidence, and a conservative age or completed-run predicate. |
| Container | Cartulary-managed Docker labels plus suite/run/target/service metadata, and a conservative age or completed-run predicate. |
| Browser fixture | Generated browser fixture metadata plus suite or target ownership evidence. |

TODO: maintainer_decision_required - Define parent-death cleanup guarantees, active DB cleanup guarantees, detached reaper completion proof, exact stale age thresholds, and dry-run requirements for stale janitors.

## 13. Generated Artifacts and Drift Detection

Generated artifacts are downstream execution inputs. They must not be treated as behavioral owners.

| Generated artifact | Owner inputs | Generation command | Drift command | Hand-edit rule |
|---|---|---|---|---|
| Go generated code under `internal/gen/**` | Contracts, SQL, generators | `make generate` | `make generate-drift` | Do not hand-edit. |
| TypeScript protocol output under `packages/protocol-ts/src/generated/**` | Contracts and protocol generator | `make generate` | `make generate-drift` | Do not hand-edit. |
| `tools/task_surface.generated.mk` | Execution topology/task surface inputs | `make generate` or task topology renderer | `make generate-drift`, `make generated-artifact-policy-check` | Do not hand-edit. |
| Task/schedule/browser manifests under `tools/*manifest*.json` | `tools/execution_topology_manifest.json` and related owner manifests | `make generate` or named render scripts | `make generate-drift`, `make json-shape-check`, schedule drift checks | Generated artifacts are downstream unless explicitly declared owner inputs. |
| Phase ledgers and schedules | Phase test maps/manifests | `make phase-ledgers`, `make phase-schedules` | `make phase-ledger-drift`, `make phase-schedule-drift` | Treat as generated verification artifacts. |
| Duration baselines | Selected successful run artifacts | `make *-duration-baselines RESULTS_DIR=<dir>` | `make *-duration-baseline-drift RESULTS_DIR=<dir>` | Refresh only from explicit selected successful runs. |

Drift commands must fail when committed generated outputs do not match owner inputs or when required generated artifacts are missing, empty, or policy-inconsistent.

## 14. CI, Check, and Release Verification

`make check` is the developer verification gate. Passing `check` means the check scheduler completed its declared work for the current repository and environment: toolchain readiness, generated/drift checks, lint/static/security checks, frontend checks, service-backed backend and browser verification, migration verification, harness smoke, duration-baseline checks, and build/deployable-shape readiness as declared by the current manifests.

`make ci` is provider-neutral. It composes the canonical repo task surface through `check` according to the task-surface sequence. `scripts/ci/verify.sh` sets `CARTULARY_OUTPUT_MODE=ci` only when the caller has not selected an output mode.

`make release-check` composes `check`, license report verification, SBOM verification, and build verification. Passing `release-check` means those repository targets passed in the current environment. It does not by itself publish release claims, hosted CI claims, benchmark claims, or provider upload guarantees.

| Command | Passing means | Passing does not mean |
|---|---|---|
| `check` | Developer gate passed for selected current-host work. | Provider CI annotations/upload behavior; public benchmark publication; non-observed platform support. |
| `ci` | Provider-neutral CI entrypoint passed its Make sequence. | `.github/**` workflow behavior, hosted runner retention, annotations, or uploads. |
| `release-check` | Developer gate, release evidence checks, and build verification passed. | Release publication readiness beyond configured artifacts; external distribution claims. |

TODO: source_limited - CI provider workflows, annotations, artifact uploads, and dashboard behavior are not specified because provider source is absent.

## 15. Integration with Product Specifications

Harness tests may reference product requirement IDs, phase ledgers, coverage ledgers, phase maps, and acceptance criteria. Those references provide evidence routing only. They must not redefine the product behavior under test.

Product assertion failures and harness operational failures must remain separate:

- A failing assertion after setup succeeds is `test_assertion_failure` with product origin.
- Setup, readiness, fixture, artifact, scheduler, timeout, or cleanup failure is a harness operational failure.
- Support-only tests and helper phases are support evidence unless the schedule explicitly lists them as produced summary targets.
- Raw aggregate suites and direct package scripts are not authoritative product conformance evidence unless a canonical Make target adopts them.
- Unowned regressions and unmapped tests must be classified as unowned/unmapped rather than silently counted as authoritative coverage.

Phase/test maps participate in acceptance evidence by declaring selected tests and owners. Retained artifacts prove only the run identified by explicit result root and run ID. Newest-run fallback is human investigation only.

Load, login bursts, service resets, artificial stress margins, and browser harness stress tests are harness-only unless product specs explicitly adopt them. The Phase 6 browser login burst remains a harness stress margin and does not change Core 04 session-cap semantics.

## 16. Platform and Toolchain Support

The supported baseline for repository work is Go `1.26` with toolchain `go1.26.3`, Node `24.15.0`, pnpm `10.33.0`, and the pinned tools listed in repository procedure. Platform support is evidence-scoped.

| Environment/tool | Status | Contract |
|---|---|---|
| WSL/Linux current host profile | supported for selected evidence | Recovery runtime evidence observed WSL/Linux with the pinned toolchain and Docker/Compose versions. |
| General Linux | source-limited supported intent | The harness aims to remain portable across Linux where practical, but this NLSpec does not certify all distributions. |
| macOS | TODO: source_limited | No complete platform evidence was inspected. |
| Windows native | TODO: source_limited | Use WSL command guidance where repository procedure says so; native behavior is not specified. |
| CI hosted runners | TODO: source_limited | Provider source absent. |
| Docker | required for service-backed targets | Missing Docker is a preflight/configuration failure. |
| Docker Compose | required for local services and standalone browser backing services | Missing Compose or failed readiness is a service/configuration failure. |
| Podman | TODO: source_limited | No evidence adopts Podman as compatible. |
| Go | required | Version/pin mismatch fails `doctor` or toolchain checks. |
| Node/pnpm/Corepack | required for frontend/browser targets | Repo-local runtime is canonical through Make. |
| Playwright browser runtime | required for browser targets | Missing runtime fails browser/toolchain target. |
| Bash/core utilities | required where scripts use them | Missing `bash`, `curl`, `realpath`, `setsid`, or `ss` behavior is command-specific; `ss` absence weakens port-release proof. |

## 17. Security and Test-Only Boundaries

The harness uses local test credentials for Postgres, MinIO, and browser fixtures. These credentials are not production secret policy.

Test-only routes:

- Must be disabled by default.
- May be enabled only with `CARTULARY_ENABLE_TEST_ROUTES=1`.
- Must be treated as harness-owned support routes.
- Must not be documented as production API behavior.

Secrets and redaction:

- Service diagnostics must avoid exposing secrets where redaction helpers cover the surface.
- TODO: maintainer_decision_required - Define complete redaction guarantees for environment variables, service credentials, object-store credentials, artifact logs, and CI secrets.

Destructive boundaries:

- Cleanup must remain inside exact repo-local path registries or proof-gated generated service resources.
- Stale janitors must not delete resources lacking generated names and harness proof metadata.
- Browser reset route must operate only on the configured test database/object bucket for the harness-owned runtime that exposes the route.

Production exclusion proof is source-observed through disabled default server wiring and browser-owned explicit env assignment. A stronger production-runtime exclusion claim requires deployment-owner review.

## 18. Acceptance Criteria / Definition of Done

| ID | Scope | Setup | Command or invocation | Required inputs | Expected exit | Stdout/stderr | Required artifacts/schema fields | Cleanup expectation | Negative case |
|---|---|---|---|---|---|---|---|---|---|
| TH-HARNESS-AC-001 | Command registry completeness | Current tree | Compare this Section 5 registry with `tools/task_surface_manifest.json` public targets | none | 0 for comparison tool or manual review pass | No missing/extra public targets | Registry lists all 77 public targets and classifies direct scripts | none | Extra or missing target fails review |
| TH-HARNESS-AC-002 | Help/task surface | Toolchain not required | `make help`; `make help-all`; `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all` | none | 0 | Bounded human help, successful stderr empty | No central artifact required for help targets | none | Unknown help flags fail usage |
| TH-HARNESS-AC-003 | Generated drift | Clean checkout | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check` | pinned tools present | 0 | Bounded summary | Tool summaries include status, target, failure fields | none | Edited generated file causes artifact/drift failure |
| TH-HARNESS-AC-004 | Leaf target success | Toolchain ready | `make backend-unit` or `make frontend-typecheck` | result root/run ID optional | 0 | Summary lines in summary mode, stderr empty on success | `cartulary.tool_run_summary.v1`, target/phase summaries where target emits them | no service cleanup | Product assertion failure exits non-zero with `test` class |
| TH-HARNESS-AC-005 | Aggregate success | Toolchain and services ready | `make test-fast` | backing services available | 0 | Aggregate summary with artifact refs | Run summary and child target summaries | Service cleanup for service-backed children | Child failure marks aggregate failed |
| TH-HARNESS-AC-006 | Scheduler success | Toolchain and services ready | `make check` | Docker/Node/Go/pnpm/Playwright available | 0 | Bounded progress and summary; no raw child logs in summary mode | Check scheduler summary/events, run summary, tool summary | Suite/browser cleanup attempts recorded | Invalid scheduler config exits before child work |
| TH-HARNESS-AC-007 | Invalid usage | No setup | `make service-backed-slice` without `PHASE` | none | non-zero, observed 2 | Usage/config diagnostic; no child test logs | Tool summary if wrapper reaches summary layer | no services started | Supplying valid `PHASE` must not fail usage |
| TH-HARNESS-AC-008 | Invalid configuration | No child work | `CHECK_HOST_CPU_JOBS=0 make check` | invalid env | non-zero, observed 2 | Config diagnostic | No scheduler child artifacts required beyond failure summary | no service startup | Positive integer must pass config validation |
| TH-HARNESS-AC-009 | Service readiness failure | Docker made unavailable or service binary invalid by controlled setup | Service-backed target | controlled failing service config | non-zero | Preflight/readiness diagnostic | Service scope or target failure artifact where wrapper reaches reporting | No managed resources remain if startup did not occur | Normal Docker config must not fail preflight |
| TH-HARNESS-AC-010 | Product test failure | Inject known failing test in disposable branch | Canonical test target | failing assertion | non-zero | Failure headline with `failure_class=test` | Test logs and target/tool summary | Harness resources cleaned | Harness setup failure must not be classified as product failure |
| TH-HARNESS-AC-011 | Child target failure | Use aggregate with one controlled failing child | Aggregate or scheduler target | failing child | non-zero | Child target/work unit named | Run/scheduler summary records failed child and skipped descendants | Finalizers run | Missing child artifact without child failure is artifact failure |
| TH-HARNESS-AC-012 | Cleanup path guard | Dry inspection or controlled temp registry test | `make -n clean`; guarded cleanup unit tests | none | 0 for safe registry; non-zero for unsafe synthetic path | Deletion list or guard error | No central artifact required | Empty, `/`, `.`, and outside-repo paths rejected | Missing registered paths are ignored |
| TH-HARNESS-AC-013 | Stale janitor proof gate | Unit tests with fake resources | Focused testservices stale janitor tests | generated and unproven resource fixtures | 0 | Test output summary | Test evidence shows unproven resources refused | No real unproven resource deleted | Resource lacking generated name/proof is retained |
| TH-HARNESS-AC-014 | Machine output | Centralized top-level target | `CARTULARY_OUTPUT_MODE=machine make <target>` for a supported target | result root/run ID optional | target status | Exactly one JSON object or pointer; no child logs/progress prose | `schema_id=cartulary.tool_run_summary.v1` or valid pointer | normal target cleanup | Duplicate JSON or progress prose fails |
| TH-HARNESS-AC-015 | Retained artifact identity | Run any summary target with explicit result root/run ID | `CARTULARY_TEST_RESULTS_DIR=<dir> CARTULARY_TEST_RUN_ID=<id> make backend-unit` | explicit dir/id | 0 | Summary names artifact root | Artifacts exist under `<dir>/<id>` | `make clean` removes default result root only when registered | Newest-run fallback not accepted as proof |
| TH-HARNESS-AC-016 | Provider-neutral CI | Toolchain/services ready | `make ci` | current repository config | Same as child `check` status | CI bounded output | CI aggregate/tool summaries | Child cleanup semantics | Hosted-provider annotation/upload claims remain source-limited |
| TH-HARNESS-AC-017 | Release-check | Toolchain/services ready and release artifacts configured | `make release-check` | license/SBOM/build inputs | 0 only if all child checks and artifacts pass | Release summary | Check, license/SBOM, build summaries | Child cleanup semantics | Missing/empty license or SBOM artifact fails |

Definition of done for this NLSpec:

- All public Make targets are listed or classified.
- Every required section exists and uses present-tense contract language.
- Every known contradiction, ambiguity, or missing decision is in Section 19.
- Recovery row IDs are traceability only.
- No claim relies on unspecified retained artifacts.

## 19. Open Decisions, Ambiguities, and Source Limits

| ID | Affected section | Conflicting or missing evidence | Why it matters | Decision owner | Exact decision needed | Blocked requirements | Interim treatment |
|---|---|---|---|---|---|---|---|
| TH-HARNESS-OD-001 | 6 | Env reads exist across Make, scripts, config, manifests, and direct tools; no full override matrix | Public callers need predictable config precedence | Harness maintainer | Define precedence across CLI, Make variables, env, manifests, config files, direct package scripts, and hardcoded defaults | Full config contract | TODO: maintainer_decision_required |
| TH-HARNESS-OD-002 | 10, 18 | Visual snapshots exist but refresh OS/browser/version/command is not adopted | Prevents unstable golden updates | Harness/product test owners | Name exact refresh command, platform bounds, review evidence, and acceptance path | Snapshot refresh contract | TODO: maintainer_decision_required |
| TH-HARNESS-OD-003 | 12 | Parent-death cleanup not fully evidenced | Determines leak guarantees after abrupt agent/CI death | Harness maintainer | Define whether parent-death cleanup is guaranteed, diagnostic-only, or unsupported | Cleanup guarantee | TODO: maintainer_decision_required |
| TH-HARNESS-OD-004 | 12 | Active DB cleanup source-limited | Determines stale DB deletion after interrupted tests | Harness maintainer | Define active-connection cleanup predicates and proof | Stale janitor DB cleanup | TODO: maintainer_decision_required |
| TH-HARNESS-OD-005 | 12 | Detached reaper scheduling is observed; completion proof incomplete | Determines whether cleanup can be claimed complete after process exit | Harness maintainer | Define required reaper proof artifacts and timeout | Cleanup acceptance | TODO: maintainer_decision_required |
| TH-HARNESS-OD-006 | 14 | `.github/**` absent | Hosted CI behavior cannot be specified | CI owner | Provide provider workflow source or state provider behavior is out of scope | CI annotations/uploads | TODO: source_limited |
| TH-HARNESS-OD-007 | 10 | Playwright reports/traces/videos/screenshots are tool-defined | Machine consumers cannot rely on field schema | Browser test owner | Adopt specific Playwright schema/version or keep diagnostic-only | Artifact schemas | TODO: source_limited |
| TH-HARNESS-OD-008 | 16 | Only WSL/Linux selected profile evidenced | Platform claims could overstate portability | Repo/tooling owner | Define support matrix for Linux/macOS/Windows/CI/Docker variants | Platform contract | TODO: source_limited |
| TH-HARNESS-OD-009 | 9 | Reset route concurrency, timeout, readiness, and partial failure semantics not closed | Browser reset may race or partially mutate state | Harness and server test owners | Define concurrency behavior, deadline, idempotence, and partial-failure handling | Reset route contract | TODO: maintainer_decision_required |
| TH-HARNESS-OD-010 | 17 | Redaction helpers exist but full secret/log/artifact redaction scope not closed | Prevents accidental secret leakage claims | Security/harness owner | Define required env, log, artifact, CI secret redaction | Security boundary | TODO: maintainer_decision_required |
| TH-HARNESS-OD-011 | 14 | Historical selected evidence showed CI/release-check failures in recovery; current state may differ | Prevents stale readiness claims | Harness maintainer | Decide whether to require fresh passing runs before adoption | CI/release readiness | TODO: source_limited until current validation run is cited |
| TH-HARNESS-OD-012 | 10 | Some machine-mode behavior is centralized, some helper/direct surfaces are partial | Prevents exact JSON contract for all commands | Harness maintainer | Define per-command machine-output support matrix beyond centralized wrappers | Machine output for all surfaces | TODO: source_limited |

## 20. Traceability Appendix

Primary repository evidence:

- `Makefile`
- `tools/task_surface_manifest.json`
- `tools/task_surface.generated.mk`
- `tools/execution_topology_manifest.json`
- `tools/check_schedule_manifest.json`
- `tools/service_backed_schedule_manifest.json`
- `tools/scheduler_resource_registry.json`
- `tools/generated_artifact_policy.json`
- `tools/browser_e2e_batch_manifest.json`
- `tools/schemas/cartulary.tool_run_summary.v1.schema.json`
- `scripts/cartulary-runner.mjs`
- `scripts/run-make-sequence.sh`
- `scripts/run-check-schedule.mjs`
- `scripts/run-service-backed-schedule.mjs`
- `scripts/lib/scheduler/**`
- `scripts/lib/scheduler-resources.mjs`
- `scripts/lib/tool-output.mjs`
- `scripts/lib/test-output/**`
- `scripts/start-web-e2e.sh`
- `scripts/reset-web-e2e-stack.sh`
- `scripts/dev-services.sh`
- `scripts/dev-stack.sh`
- `tools/testservices/**`
- `internal/testutil/pgtest/**`
- `internal/testutil/s3test/**`
- `internal/testutil/suiteservices/**`
- `internal/testutil/testruntime/reset.go`
- `internal/testutil/testruntime/reset_integration_test.go`
- `cmd/server/main.go`
- `apps/web/playwright*.config.ts`
- `apps/web/e2e/**`
- root `package.json`
- `apps/web/package.json`
- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/domain.md`
- `docs/guides/cartulary_implementation_testing_guide.md`
- `docs/guides/cartulary-dev-guide.md`
- `docs/guides/cartulary_repository_bootstrap_guide.md`

Recovery evidence used for traceability only:

- `docs/testing-harness-spec-recovery-docs/harness-nlspec.md`
- `docs/testing-harness-spec-recovery-docs/harness-acceptance-matrix.md`
- `docs/testing-harness-spec-recovery-docs/maintainer-decision-summary-2026-05-09.md`
- `docs/testing-harness-spec-recovery-docs/environment-contract-observations.md`
- `docs/testing-harness-spec-recovery-docs/service-lifecycle-map.md`
- `docs/testing-harness-spec-recovery-docs/structured-output-schema-notes.md`
- `docs/testing-harness-spec-recovery-docs/failure-class-taxonomy.md`
- `docs/testing-harness-spec-recovery-docs/cleanup-lifecycle-matrix.md`
- `docs/testing-harness-spec-recovery-docs/cleanup-signal-evidence-register.md`
- `docs/testing-harness-spec-recovery-docs/runtime-evidence-register.md`
- `docs/testing-harness-spec-recovery-docs/source-limit-log.md`
- `docs/testing-harness-spec-recovery-docs/entrypoint-command-map.md`
- `docs/testing-harness-spec-recovery-docs/observable-interface-map.md`
- `docs/testing-harness-spec-recovery-docs/resource-allocation-register.md`
- `docs/testing-harness-spec-recovery-docs/race-timing-resource-register.md`
- `docs/testing-harness-spec-recovery-docs/timeout-retry-register.md`

Key recovery decision IDs referenced by behavior:

- `MD-S7-0001`: Make is the sole canonical harness command surface.
- `MD-S7-0003`: Generated artifacts are downstream execution inputs only.
- `MD-S7-0004`: Reset route is harness-owned and outside production app ownership.
- `MD-S7-0005`: Local-dev Compose and `make dev` are local verification behavior.
- `MD-S7-0006`: Scheduler lanes are logical constraints only.
- `MD-S7-0007`: Stale janitors require generated names plus metadata or lease evidence and conservative age or completed-summary checks.
- `MD-S7-0008`: Cleanup strength is evidence-scoped; parent-death and active DB cleanup remain source-limited.
- `MD-S7-0009`: External Go caches are outside default cleanup.
- `MD-S7-0010`: Environment variables/defaults are source-observed; precedence is unresolved.
- `MD-S7-0011`: WSL/Linux is the primary observed environment; no full support matrix is closed.
- `MD-S7-0012`: Retained artifact claims require explicit run identity.
- `MD-S7-0013`: Visual snapshots remain validation-only until refresh bounds are adopted.
- `MD-S7-0014`: Known harness machine schemas are stabilized; unknown tool schemas remain partial/unknown.
- `MD-S7-0015`: CI remains provider-neutral.
- `MD-S7-0018`: Phase 6 browser login burst is harness-only stress margin.
