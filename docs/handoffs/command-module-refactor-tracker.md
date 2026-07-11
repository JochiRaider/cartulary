# command Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

- Target path: `cmd`.
- Normalized target label: `command`.
- Output path: `docs/handoffs/command-module-refactor-tracker.md`.
- Status: planning and documentation only.
- Allowed change for this session: this tracker file only.
- Non-goals: no production refactor, test edit, contract edit, generated-file edit,
  package change, migration change, or harness/accounting edit.
- Architectural posture: `command` is a planning label, not an accepted permanent
  module boundary. The live target is three executable composition roots:
  `server`, `migrate`, and `operator`.
- Later implementation requires a separately authorized task. Any behavior change
  to routes, envelopes, WebSockets, authorization, recovery, lifecycle, CLI output,
  or harness behavior requires explicit later authorization.

Source hierarchy used:

1. Adopted subsystem NLSpecs for their named subsystem only.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. `docs/domain.md` and implementation-support guides for terminology and mechanics.
5. Current code, tests, manifests, and Make targets for repository state.
6. Prior trackers and the planning framework as evidence only.

Owner and support documents inspected:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md` was read first.
- `docs/spec/00_document_set_status_and_precedence.md`, including authority and
  contract-owner posture.
- Relevant module, interface, recovery, packaging, and invariant sections of Core 01.
- Relevant record, saved-view, evidence, projection/source, and history sections of
  Core 02.
- Relevant workbook, collaboration, conflict, evidence, history, and query sections
  of Core 03.
- Relevant authorization, test-route, deployment-configuration, startup, backup, and
  restore sections of Core 04.
- Scope and authority sections of the adopted Graph Projection and Network Flow
  Activity NLSpecs.
- `docs/testing-harness-nlspec.md`, especially canonical Make invocation, generated
  artifacts, backend-process accounting, moved-test accounting, and test-route
  traceability.
- `docs/domain.md` for vocabulary and bounded-context interpretation.

Repository evidence inspected includes every tracked file under `cmd`, relevant
`internal/app` runners and runtime assembly, `internal/testutil/processtest`, platform
harness-runtime controls, the Make build and validation surfaces, phase maps that
select `cmd` tests, backend module-boundary configuration, duration baselines, and
the deployable-shape release check.

Planning findings about evidence discovery:

- `git ls-files cmd` is the inventory source of truth and reports 16 tracked files.
- `rg --files cmd` omitted `cmd/operator/**` because `.gitignore` contains the
  unanchored pattern `operator`; it also omitted tracked `.gitkeep` files.
- The reusable framework does not define a `command` domain module. The live
  repository instead requires exactly three `cmd/*/main.go` entrypoints.
- Phase-map `unit`, `integration`, and `e2e` labels are evidence accounting. They do
  not establish runtime ownership; several Phase 10 unit/integration rows currently
  execute through the `backend-process` family from `cmd/server`.

No owner contradiction was found. Unknown implementation-owner or accounting choices
are recorded as `TODO:` or blockers rather than guessed.

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `cmd/migrate/.gitkeep` | Empty directory placeholder. | None. | Repository layout only. | None. | None. | None. | Repository hygiene. | low | Redundant because the directory contains tracked Go files. |
| `cmd/migrate/main.go` | Process signal context, argument forwarding, stderr binding, and exit propagation for migrations. | Executable `migrate`; package `main`; no reusable API. | `make build-migrate`, dev-service migration flows, browser stack setup, migration drift helpers. | `internal/app.RunMigrateCLIContext`; standard signal and process packages. | `cmd/migrate/main_test.go`; `internal/app/migrate_test.go`. | Build inputs and deployable-shape evidence; no generated contract directly. | `cmd` thin root plus `internal/app` migrate runner. | low | Legitimate thin composition root. |
| `cmd/migrate/main_test.go` | Builds and runs the migrate binary outside the repository working directory and verifies schema application. | Process test only. | Not present in current phase-map/backend-process row accounting. | Postgres test harness, suite-service repo lookup, `os/exec`, direct SQL verification. | The test itself. | Migration/build evidence only; no generated contract. | Executable-boundary test, with harness accounting owner `TODO`. | medium | Contract is valuable but lacks a discovered canonical Make row selection. |
| `cmd/operator/main.go` | Process signal context, CLI argument forwarding, stdout/stderr binding, and exit propagation. | Executable `operator`; package `main`; no reusable API. | `make build-operator`, release and operational recovery surfaces, operator process tests. | `internal/app.RunOperatorCLIContext`; standard signal and process packages. | `cmd/operator/operator_phase10_test.go`; `internal/app/operator*_test.go`. | Operator result/progress schema IDs are owned in Core 01 and `internal/app`, not generated here. | `cmd` thin root plus `internal/app` operator runner. | low | Legitimate thin composition root. |
| `cmd/operator/operator_phase10_test.go` | Canonical operator CLI process evidence, object-store initialization smoke, recovery helpers, and Phase F SeaweedFS migration-support evidence. | Process and support test symbols; no production API. | Phase 10 `backend-process` rows, release evidence, duration baselines. | `internal/app`, recovery, recovery operator operations, evidence blob references, config, object store, Postgres/S3 test harnesses, direct SQL and process execution. | The test file and Phase 10 map selectors. | Operator JSON/JSONL contract schemas, Phase 10 map, duration baselines, retained migration evidence. | Keep CLI process checks in `cmd/operator`; review support logic for `internal/app`, recovery, object store, or owner-local test support. | high | Mixed executable characterization and non-CLI migration-support implementation. |
| `cmd/server/.gitkeep` | Empty directory placeholder. | None. | Repository layout only. | None. | None. | None. | Repository hygiene. | low | Redundant because the directory contains tracked Go files. |
| `cmd/server/main.go` | Server startup, logging, configuration load, guarded test-route composition, app runtime construction, HTTP listener selection, inherited-FD serving, graceful shutdown, and diagnostics. | Executable `server`; package-local `serveHTTP` and `writeStartupError`; three environment keys. | `make build-server`, dev stack, browser stack, `processtest`, packaged deployment. | `internal/app`; auth, savedviews, timeline, and Network Flow harness controls; platform config, harness runtime, and HTTP API; `net/http`. | All `cmd/server` process tests plus app/platform integration tests. | Public HTTP/WS behavior indirectly; build inputs, deployable shape, phase maps, runtime security accounting. | `cmd` signal/exit root; `internal/app` assembly; platform runtime/listener adapter. | high | Not currently a thin composition root; directly composes module and platform test routes. |
| `cmd/server/main_embedded_frontend_process_test.go` | Real-process root HTML and embedded asset smoke. | Process test and asset-path matcher. | Not present in current backend-process phase rows; release deployable-shape check covers adjacent packaging behavior. | HTTP client and Phase 1 process fixture. | The test itself. | Embedded asset archive/build contract; no generated protocol contract. | Server executable-boundary test. | medium | Keep at process boundary; accounting owner is `TODO`. |
| `cmd/server/main_phase0_e2e_test.go` | Startup readiness, invalid diagnostics, first-admin bootstrap, skip/recovery, database, object-store, audit, and secret-safety process evidence. | Phase 0 process tests and shared helpers. | Authoritative Phase 0 `E-0-01` through `E-0-05` backend-process rows. | Auth bootstrap test support, object store, process/config fixtures, Postgres/S3, audit/security assertions, direct SQL. | Selected by Phase 0 map; helpers are reused by later server tests. | Phase 0 accounting and startup diagnostic goldens; no generated protocol contract. | Server process boundary plus owner-local auth/platform assertions. | high | Legitimate black-box evidence; helper ownership should remain explicit. |
| `cmd/server/main_phase10_config_test.go` | Real-process fail-closed validation for `roots.backup_storage`. | Phase 10 process test and environment helper. | Authoritative Phase 10 `E-10-04` backend-process row. | Platform config and shared process/config fixtures. | Selected by Phase 10 map. | Phase 10 accounting; deployment-config contract only. | Server process boundary and config owner. | medium | Correct process-level characterization. |
| `cmd/server/main_phase10_recovery_sentinel_test.go` | Restore orchestration, integrity failures, projection rebuild, restored workbook consistency, evidence artifact emission, and absence of public recovery routes. | Phase 10 unit, integration, and process tests plus extensive recovery fixtures. | Phase 10 `U-10-02`, `U-10-03`, `I-10-02`, `I-10-03`, and `E-10-03` backend-process rows. | Recovery and projections modules, restore contract, object store, Postgres, process fixture, HTTP helpers, filesystem artifact writes. | Selected by Phase 10 map; reuses Phase 0/1/2/5 helpers. | Restore-verification artifacts, Phase 10 accounting, duration baselines. | Split recovery/projections owner tests from true `cmd/server` process checks. | high | Principal misplaced owner-specific test logic under `cmd`. |
| `cmd/server/main_phase1_process_test.go` | Login, session, CSRF, WebSocket revocation, TOTP enrollment, credential, and user-administration process smoke. | Supplemental Phase 1 process tests and reusable HTTP/auth helpers. | Supplemental `E-1-SMOKE-01` backend-process row. | Auth test support, authn constants, WebSocket client, process and HTTP fixtures. | Selected by Phase 1 map; helpers reused by Phase 2/5/10 tests. | Phase 1 accounting and duration baselines. | Server process boundary; auth owns behavior. | medium | Evidence only; does not confer auth ownership on `cmd`. |
| `cmd/server/main_phase2_smoke_test.go` | Incident creation/list/patch, workbook preferences, membership administration, extension discovery, and hidden-incident authorization smoke. | Supplemental Phase 2 process tests and incident-create helper. | Supplemental `E-2-SMOKE-01` backend-process row. | Authn, HTTP assertions, process fixture, Phase 1 helpers. | Selected by Phase 2 map. | Phase 2 accounting and duration baselines. | Server process boundary; incidents/auth/extensions own behavior. | medium | Evidence only; no command-domain behavior. |
| `cmd/server/main_phase5_smoke_test.go` | Evidence row creation, blob upload/attach, record patch, Timeline projection query, and no-blob preview failure smoke. | Supplemental Phase 5 process test and HTTP helpers. | Supplemental `E-5-SMOKE-01` backend-process row. | Authn, HTTP/process fixtures, Phase 1/2 helpers. | Selected by Phase 5 map. | Phase 5 accounting and duration baselines; view-schema IDs are consumed, not owned. | Server process boundary; evidence/timeline/projections own behavior. | high | Cross-module black-box evidence; preserve envelopes and projection effects. |
| `cmd/server/main_test_runtime_routes_process_test.go` | Default-disabled, fail-closed, token/host/origin security, reset, clock, and Network Flow test-control process characterization. | Process tests and harness-route request helpers. | No current backend-process phase row discovered; prior test-util handoff cites direct validation. | Platform harness runtime via server, Network Flow harness controls, process fixture, test-route token helpers. | The test itself. | Harness route contract and security traceability; no public generated contract. | Server process boundary plus harness/platform owner. | high | Important characterization currently missing explicit row accounting. |
| `cmd/server/shared_process_harness_test.go` | Package-level `TestMain` startup and cleanup for shared Postgres and S3 test services. | `TestMain`, shared harness accessors. | Automatically runs for selected `cmd/server` package tests. | `internal/testutil/pgtest`, `s3test`, process exit and diagnostics. | All selected `cmd/server` tests. | Backend-process fixture lifecycle only. | Thin package wrapper over test-util service facades. | medium | Keep required wrapper; do not add product behavior. |

## 3. Module Boundary Diagnosis

The target is not a domain `command` module. It is a mixed package tree containing
legitimate executable boundaries, server transport/application assembly, and broad
process or owner-specific tests. Production `migrate` and `operator` entrypoints are
already thin. Production `server` is a mixed-responsibility composition root.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| OS signal context and final process exit | Three `main.go` files | `cmd/*` | keep | All entrypoints construct signal contexts and call `os.Exit`. | Appropriate binary-root behavior. |
| Migration CLI behavior | `internal/app/migrate.go`, invoked by `cmd/migrate` | `internal/app` plus platform Postgres | keep | `RunMigrateCLIContext` owns parsing and migration invocation. | `cmd/migrate` is already thin. |
| Operator CLI behavior | `internal/app/operator*.go`, invoked by `cmd/operator` | `internal/app` plus recovery/platform owners | keep | `RunOperatorCLIContext` owns the CLI facade. | `cmd/operator` is already thin. |
| Server application assembly | `cmd/server/main.go` and `internal/app/runtime.go` | `internal/app` | split | Server directly loads config, builds test-route options, and calls `app.NewRuntime`. | Add a later server-runner facade consistent with other binaries. |
| HTTP listener, inherited FD, timeout, and shutdown | `cmd/server/main.go` | `internal/platform` runtime adapter | move | `serveHTTP` handles `CARTULARY_HTTP_LISTEN_FD`, `net.FileListener`, and `http.Server`. | Exact platform package name is an implementation blocker. |
| Guarded test-route composition | `cmd/server/main.go` | `internal/app` assembly using platform/module contributions | move | Direct imports of auth, savedviews, timeline, harnessruntime, and Network Flow harness controls. | Preserve Core 04 and harness predicates exactly. |
| Server process characterization | `cmd/server/*process_test.go` and phase smoke tests | Executable boundary with owner-local helpers | keep / split | Tests launch the real process and assert observable behavior. | Phase labels remain evidence accounting only. |
| Recovery orchestration tests and fixtures | `main_phase10_recovery_sentinel_test.go` | Recovery and projections module testsupport | move / split | Most tests invoke recovery APIs directly and only some start `cmd/server`. | Keep route-absence and true restored-process checks at command boundary. |
| Operator CLI process evidence | `operator_phase10_test.go` | Executable boundary | keep | Tests build/run the binary and validate CLI wire output. | Core 01 owns logical commands; filename is implementation-owned. |
| Phase F object-store migration support | `operator_phase10_test.go` | Recovery, object-store, or `internal/app` test support | defer | Support tests call internal migration helpers and emit retained artifacts. | `TODO:` owner decision required before movement. |
| Shared service lifecycle wrapper | `shared_process_harness_test.go` | `cmd/server` wrapper over `internal/testutil` | keep / split | Go requires package `TestMain` in the tested package. | Keep wrapper thin; service behavior belongs to test-util. |
| Embedded frontend process smoke | `main_embedded_frontend_process_test.go` | Server executable boundary | keep | Verifies packaged UI from the real server. | No frontend controller or grid ownership is present. |
| Grid vendor integration | None in `cmd` | `/packages/grid-adapter` | defer | No direct imports or runtime use found. | Not applicable to this target. |

Diagnosis categories:

- Legitimate thin application/service facade: `cmd/migrate`, `cmd/operator`.
- Transport-adjacent adapter and mixed-responsibility package: `cmd/server`.
- Misplaced home for owner-specific logic: recovery and migration-support test bodies.
- View/projection orchestration: only in tests; production ownership remains in modules.
- Frontend shell/controller or grid-vendor integration: not present, apart from a
  packaged-asset process smoke.

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Exactly three executable identities and build roots | Core 01 deployable topology plus Make/release support | `cmd/*/main.go`, Make build inputs, deployable-shape check | Build/release checks | Preserve exact server, migrate, and operator artifacts | high | Do not create a permanent `command` deployable/module. |
| Server config load and structured startup diagnostics | Core 04 §12 | `config.Load`, `writeStartupError`, process stderr | Phase 0 and Phase 10 config process tests | Add runner-level writer/exit-code characterization before movement | high | Invalid config must expose no HTTP or WS listener. |
| HTTP address, inherited listener FD, and graceful shutdown | Platform runtime support; repository process contract | `CARTULARY_HTTP_ADDR`, `CARTULARY_HTTP_LISTEN_FD`, `serveHTTP` | Indirectly exercised by `processtest` | Add invalid-FD, inherited-listener, cancellation, and shutdown characterization | high | Preserve listener ownership and address logging. |
| Packaged root HTML and embedded assets | Core 01 packaged deployable boundary | embedded process smoke and deployable-shape check | `TestServerEmbeddedFrontendAssets_ProcessSmoke` | Account the test in a canonical owner-selected row | medium | No frontend runtime redesign is planned. |
| Public HTTP routes and success/error envelopes | Core 01 route and envelope sections | `app.NewRuntime` route assembly | Phase 1/2/5 process smoke and module tests | Preserve route inventories; no command-local route definitions | high | `cmd` consumes route registrars but does not own route semantics. |
| WebSocket paths, authorization, session revocation, and event behavior | Core 01 and Core 03; Core 04 authorization | process WebSocket helpers and runtime hub | Phase 0 refusal and Phase 1 revocation tests | Preserve paths, auth rejection, and revocation reasons | high | No WebSocket contract change is authorized. |
| Test-only runtime routes | Core 04 `REQ-04-109` and Testing Harness NLSpec | server guarded-route wiring and harnessruntime controls | Platform integration tests and server process characterization | Add explicit canonical accounting for server process tests | high | Must remain disabled by default and token/marker/host/origin gated. |
| Auth, session, CSRF, enrollment, credential, and admin behavior | Core 01 public contracts; Core 04 security | Phase 1 process tests | Supplemental Phase 1 smoke plus auth owner tests | Preserve cookies, CSRF header, status/error envelopes, revocation | high | Auth owns behavior. |
| Incident, membership, preferences, extension discovery, and hidden-resource behavior | Core 01, Core 02, Core 03, Core 04 | Phase 2 process tests | Supplemental Phase 2 smoke and owner tests | Preserve authorization order and route envelopes | high | Deployment admin does not bypass incident membership. |
| Entity/evidence rows, row query/mutation, revision, and projection refresh | Core 01-03 | Phase 5 process test and recovery workbook probe | Phase 5 smoke, module tests, Phase 10 restored query | Preserve `row_version`, change-set, projection, and response-row effects | high | View-schema IDs and fields are consumed, not inferred from labels. |
| Saved-view and view-schema behavior | Core 01-03 | production test-route registrar and app route assembly | Module/browser tests; no dedicated command test found | Preserve route registration and harness-only fixture routes | medium | Generated view contracts are not owned by `cmd`. |
| Recovery ordering, integrity, projection rebuild, readiness, and restored workbook | Core 01 §12, Core 04 recovery criteria | Phase 10 recovery sentinel tests | U/I/E Phase 10 rows under backend-process | Preserve pre-mutation failure, deterministic steps, and readiness gating | high | Graph Projection NLSpec does not own workbook restore rebuild behavior. |
| Absence of public backup/restore HTTP and WS routes | Core 01 operator recovery boundary | `TestPhase10_E_10_03_PublicRouteInventoryAbsence` | Authoritative E-10-03 | Keep a real-process route inventory check | high | Recovery remains deployment-local CLI behavior. |
| Operator logical command grammar, output, progress, errors, exit codes, and redaction | Core 01 §12.2.1; Core 04 recovery security | operator process tests and app runner | Authoritative E-10-01 | Preserve exact JSON/JSONL schemas and no-extra-output rule | high | Behavior change requires later authorization. |
| Operator object-store initialization | Repository operational contract in `internal/app` and Make/service flows | operator process smoke | `TestMVPObjectStoreInitOperatorCreatesConfiguredBucket` | Discover canonical accounting or mark explicit support | medium | Output must not reveal storage details. |
| Migration execution outside repository CWD | Core 01 packaging/runtime-root boundary and migration policy | migrate process test | `TestMigrateBinaryRunsFromNonRepoWorkingDirectory` | Add canonical Make-owned selection before movement | high | Must continue using embedded/authored migration source without CWD reliance. |
| Harness/test accounting and retained evidence | Testing Harness NLSpec | phase maps, target plan, ledgers, schedules, duration baselines | `backend-process` target diagnostics | Add owner-selected rows for unmapped process tests; update owner inputs first | high | Accounting is evidence routing, not runtime architecture. |
| Generated protocol/view contracts | Core 01 and derived contract owners | No direct generated import under `cmd` found | Drift checks elsewhere | No hand edits; run drift checks if accounting or build inputs change | medium | Test moves may affect generated harness outputs, not product codegen. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| `cmd/server` directly composes module and platform test routes. | Direct imports in `cmd/server/main.go`. | Server root continues to accumulate application behavior. | `must_fix` | `internal/app` assembly with owner contributions. | Characterize predicates, then move composition behind a server runner. |
| Listener inheritance and transport shutdown live beside module assembly. | `serveHTTP` handles FD parsing, listener creation, server address, and serving. | Transport errors or shutdown behavior may drift during thinning. | `should_fix` | Internal platform runtime adapter. | Add focused characterization and define the platform seam. |
| Recovery unit/integration behavior is implemented in a `cmd/server` test file. | Direct recovery/projections calls and large fixtures in Phase 10 sentinel tests. | Executable tests become a catch-all and obscure module ownership. | `should_fix` | Recovery/projections owner testsupport. | Split direct-module tests from true process checks. |
| Operator process evidence and Phase F migration-support logic share one file. | CLI process invocations coexist with direct migration helpers and retained artifact writers. | Moving one concern can disturb unrelated release evidence. | `defer` | Operator boundary plus recovery/object-store/app test support. | Resolve RB-002 before splitting. |
| Several process tests have no discovered backend-process row. | Target-plan rows omit migrate, embedded-frontend, and guarded-runtime tests. | Useful characterization can silently fall outside canonical Make evidence. | `must_fix` | Testing Harness accounting owner. | Resolve RB-001 and add owner-input rows without inferring a phase. |
| Unanchored `.gitignore` pattern hides tracked `cmd/operator` from normal `rg` discovery. | `.gitignore` contains `operator`; `git ls-files` reveals omitted files. | Future inventories may miss a required executable. | `should_fix` | Repository hygiene. | Later anchor the build artifact rule to `/operator`. |
| Redundant `.gitkeep` files remain in non-empty command directories. | Both placeholders are tracked beside Go files. | Minor review and inventory noise. | `should_fix` | Repository hygiene. | Remove in an isolated behavior-preserving cleanup slice. |
| Process tests use direct SQL and storage checks. | Phase 0/10 and operator tests inspect durable effects. | Tests can couple to persistence details. | `intentional/no_action` | Executable evidence plus owner assertions. | Retain only checks needed to prove observable durability/security; move owner mechanics with owner tests. |
| Test-only controls are linked into the server binary. | Guarded runtime routes are composed only when the enable flag equals `1`. | Incorrect movement could expose privileged controls. | `intentional/no_action` | Platform harness runtime and module-owned contributions. | Preserve fail-closed Core 04 predicates and process characterization. |
| No direct grid-vendor coupling exists under `cmd`. | Imports and source inspection found no frontend grid dependency. | None for this target. | `intentional/no_action` | `/packages/grid-adapter`. | Keep out of command workstreams. |
| Generated files could be edited accidentally when test rows move. | Phase maps feed generated ledgers, schedules, and task topology. | Hand edits would create drift and violate owner policy. | `must_fix` | Harness owner inputs and Make generators. | Update maps/manifests first and regenerate through Make only. |
| Current backend boundary check passes but does not enforce thin command roots. | `make backend-module-boundary-check` passed while `cmd/server` imports several modules. | A passing guard can be overread as architectural approval. | `should_fix` | Boundary manifest/check owner. | After the facade move, consider a production `cmd/*` import guard. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Establish authority, branch, commit, write limits, and target posture. | Tracker and owner documents. | Read-only repository inspection. | Scope says tracker-only and later authorization required. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Inventory all 16 tracked files and discover callers/dependencies/accounting. | `cmd/**`, Make/build inputs, phase maps. | `git ls-files`, target diagnostics. | Section 2 contains every tracked file. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-05 | Map every observable surface to its owner and current evidence. | Core/adopted specs, app/platform/module sources. | Owner-section reads and route/test searches. | Section 4 has an owner and test posture for each risk. |
| WF-03 | Characterization test gap analysis | chain | WF-01, WF-02 | WF-05, WF-06 | Identify missing pre-move evidence and unmapped tests. | Server runner/listener tests, phase-map owner inputs. | `make explain-target`, `make target-plan`. | RB-001 and required characterization are explicit. |
| WF-04 | Boundary/coupling scan | chain | WF-01 | WF-05, WF-06 | Classify server, recovery-test, operator-support, and hygiene findings. | `cmd`, `internal/app`, platform/runtime boundaries. | `make backend-module-boundary-check`. | Section 5 classifications are complete. |
| WF-05 | Facade or ownership redesign plan | chain | WF-02, WF-03, WF-04 | WF-06 | Keep `cmd` as thin roots and assign assembly/runtime/test responsibilities. | `cmd/server`, `internal/app`, internal platform runtime, owner testsupport. | Design review against Core 01 and repository boundaries. | Server facade and split candidates are recorded without behavior change. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07, WF-08 | Order behavior-preserving characterization, moves, cleanup, and validation. | Likely files in Section 7. | Per-slice Make targets. | Each slice has dependency, rollback, and completion rules. |
| WF-07 | Harness/test/accounting update plan | parallel | WF-03, WF-06 | WF-08 | Preserve row selection, ledgers, schedules, baselines, and build/security scans. | Owner phase maps, generated downstream outputs, duration inputs. | Shape, drift, and harness targets. | RB-001/RB-002 resolved and owner inputs named. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | none | Run narrow checks, finalization, broad gates, and update handoff. | Tracker plus implementation diff in a later task. | Section 8 command ladder. | Exact run roots, failures, skips, and residual debt recorded. |

WF-00, WF-01, WF-02, WF-04, WF-05, and WF-06 planning are complete in this
tracker. WF-03 and WF-07 retain explicit accounting/characterization work. WF-08 is
reserved for a later authorized implementation session.

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-01 | none | Add characterization for server runner exits/diagnostics, inherited listener FD, cancellation/shutdown, guarded test routes, and currently unmapped process contracts. | `cmd/server` tests, `cmd/migrate` tests, owner-selected accounting inputs. | Startup, listener, diagnostics, test-route security, migration CWD. | Existing process tests plus new focused runner/listener tests. | `make backend-unit`; `make backend-process` | Revert only added tests/accounting inputs if they mischaracterize owners. | Risky server/migrate behavior has canonical pre-move evidence or an explicit blocker. |
| S-02 | S-01 | Introduce an `internal/app` server execution facade consistent with `RunMigrateCLIContext` and `RunOperatorCLIContext`; leave signal setup and exit propagation in `cmd/server`. | `cmd/server/main.go`, new or existing `internal/app` server runner/tests. | All server startup and route-registration behavior. | Preserve Phase 0/1/2/5/10 and harness-runtime process behavior. | `make backend-unit`; `make backend-integration`; `make backend-process`; `make build-server` | Restore old `main.go` orchestration as one unit. | `cmd/server/main.go` imports only standard process support and the app facade, with identical behavior. |
| S-03 | S-02 | Move HTTP server/listener FD and graceful-shutdown plumbing behind an internal platform runtime boundary while `internal/app` owns composition. | `internal/app` server runner and a `TODO:` approved platform runtime package. | Inherited FD, server address, timeouts, shutdown races, logging. | Focused platform tests and existing real-process tests. | `make backend-unit`; `make backend-process`; `make build-server` | Keep the app facade but restore private listener implementation if the platform split regresses. | Platform owns transport plumbing and the process contract remains byte/status compatible. |
| S-04 | S-01 | Split direct recovery/projection tests and fixtures from true server-process recovery checks. | `cmd/server/main_phase10_recovery_sentinel_test.go`, recovery/projections owner testsupport, Phase 10 map. | Restore order, integrity, projection rebuild, readiness, workbook probe, route absence. | Preserve all five Phase 10 row predicates and artifact behavior. | `make backend-integration`; `make backend-process`; `make service-backed-slice PHASE=phase10` | Restore original test placement and map rows together. | Direct-module tests execute under owner packages; only real executable checks remain under `cmd/server`. |
| S-05 | S-01 | Separate canonical operator CLI process evidence from non-CLI Phase F migration-support evidence. | `cmd/operator/operator_phase10_test.go`, approved recovery/object-store/app testsupport, Phase 10 support rows. | CLI output, migration preservation, retained artifacts, release accounting. | Preserve E-10-01 and both Phase F support predicates. | `make backend-process`; `make service-backed-slice PHASE=phase10`; `make build-operator` | Move code and accounting back as one unit. | CLI tests remain command-local; support evidence has one approved semantic owner. |
| S-06 | S-02, S-05 | Remove redundant `.gitkeep` files and anchor the root operator build-artifact ignore rule. | `cmd/*/.gitkeep`, `.gitignore`. | File discovery and local build artifacts only. | Add or preserve build-input/deployable-shape discovery checks. | `make build-server`; `make build-migrate`; `make build-operator` | Restore placeholders/pattern if discovery checks identify an undocumented dependency. | `rg --files cmd` discovers tracked operator Go files and all three binaries still build. |
| S-07 | S-04, S-05, S-06 | Update owner phase/support maps and other authored accounting inputs, then regenerate downstream ledgers/schedules through Make. | Phase maps, harness owner manifests, generated ledgers/schedules. | Row selection, default-check reachability, duration/shard accounting. | Preserve authoritative/supplemental/support classification by owner decision. | `make phase-ledgers`; `make phase-schedules`; drift and shape targets | Revert owner inputs and regenerated outputs together. | No moved/unmapped command test is lost and all generated drift checks pass. |
| S-08 | S-03, S-07 | Run final narrow-to-broad validation and update this tracker handoff. | All later authorized diff files and tracker. | Any observable behavior or harness regression. | Entire relevant existing corpus. | `make agent-finalize`; `make test-fast`; `make check` | Roll back to the last passing slice, not a mixed partial state. | Required targets pass or failures are recorded with run roots and classification. |

Every slice above is behavior-preserving. Any discovered need to change a route, CLI,
envelope, authorization result, recovery/lifecycle rule, generated contract, or harness
public surface is `requires later authorization` and must be separated from these slices.

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make backend-unit` | App runner and platform/module unit characterization. | yes | Run before and after server facade/runtime work. |
| integration | `make backend-integration`; `make backend-process` | Service-backed app/module checks and real executable boundaries. | yes | `backend-process` currently covers phases 0, 1, 2, 5, and 10. |
| e2e/browser | `make browser-e2e-webserver-backed` | Browser use of the packaged server and guarded test routes. | no | Required after server/test-route or embedded-frontend movement, not for tracker-only work. |
| generated drift | `make generated-artifact-policy-check`; `make json-shape-check`; `make generate-drift`; `make phase-ledger-drift`; `make phase-schedule-drift` | Generated policy, manifests, ledgers, schedules, and codegen drift. | no | Required when test/accounting owner inputs change; never hand-edit generated outputs. |
| import-boundary/static | `make backend-module-boundary-check` | Backend owner/import boundary. | yes | Current planning baseline passed; add a thin-command guard only in a later authorized slice. |
| build | `make build-server`; `make build-migrate`; `make build-operator` | Three required executable artifacts. | no | Required after source or discovery/ignore changes. |
| harness | `make harness-contract` | Moved-test accounting and public harness behavior. | no | Required after owner map, task surface, or schedule changes. |
| full check | `make agent-finalize`; `make test-fast`; `make check` | Finalization, narrow local loop, and developer gate. | no | Follow repository order; retain and report exact run roots. |
| documentation | `make lint-markdown` | Tracker Markdown structure. | no | Required for this tracker-only creation session. |

Read-only planning commands run before tracker creation:

- Repository searches and exact source/document reads with `git ls-files`, `find`,
  `rg`, `sed`, `wc`, `jq`, and Git status/revision commands.
- `make help` and filtered `make help-all`.
- `make explain-target TARGET=backend-process DETAIL=rows`.
- `make target-plan TARGET=backend-process` and diagnostic target-plan JSON inspection.
- `make backend-module-boundary-check`, which passed with run root
  `.cartulary/test-results/20260711T223620Z-p79425`.

No unit, integration, browser, build, generated-drift, or full-check suite was run in
the planning pass.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| T-001 | Define `cmd` target and reject `command` as a domain module | scope | DONE | none | Section 1 posture | Three executable roots and exclusions are explicit. |
| T-002 | Inventory every tracked file under `cmd` | discovery | DONE | T-001 | Section 2, `git ls-files cmd` | All 16 tracked files have target-specific rows. |
| T-003 | Map public contracts and owner authority | contracts | DONE | T-002 | Section 4 | Every discovered contract risk has an owner/test posture. |
| T-004 | Record server composition and transport split | architecture | DONE | T-003 | Sections 3, 5, and 7 | App/platform candidates and characterization dependencies are explicit. |
| T-005 | Resolve unmapped process-test accounting | tests/harness | BLOCKED | T-003 | RB-001 | Harness owner selects canonical rows without phase-name inference. |
| T-006 | Resolve Phase F migration-support owner | tests/harness | BLOCKED | T-002, T-003 | RB-002 | One semantic owner and target/accounting family are approved. |
| T-007 | Implement server runner facade and platform transport seam | implementation | TODO | T-004, T-005 | S-01 through S-03 | Command root is thin and behavior-preserving checks pass. |
| T-008 | Split recovery and operator support tests | implementation | TODO | T-005, T-006 | S-04 and S-05 | Owner-local tests retain all row predicates. |
| T-009 | Apply command-tree discovery cleanup | cleanup | TODO | T-007, T-008 | S-06 | Ignore discovery and three builds remain correct. |
| T-010 | Update owner accounting and regenerate outputs | harness | TODO | T-008, T-009 | S-07 | Shape and drift checks pass with no hand edits. |
| T-011 | Execute final verification and handoff | validation | TODO | T-007, T-010 | S-08 and retained run roots | Required checks pass or failures are classified. |
| T-012 | Create this planning tracker only | documentation | DONE | T-001, T-002, T-003 | This file | Tracker contains all required sections and no production edit. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-11T18:36:58-04:00 | Codex planning/documentation session | `command` is a label for three executable roots, not a module. | Inspected framework, domain, relevant Core/adopted specs, all tracked `cmd` files; touched only this tracker. | Source reads/searches, Git inventory/status. | Authority and write limits recorded; no contradiction found. | None for tracker creation. | Obtain later implementation authorization before S-01. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-11T18:36:58-04:00 | Codex planning/documentation session | Migrate/operator roots are thin; server root mixes app and transport assembly; recovery test logic is misplaced. | Inspected `cmd/**`, `internal/app/runtime.go`, migrate/operator runners, process harness; touched only tracker. | `make backend-module-boundary-check`. | Passed at `.cartulary/test-results/20260711T223620Z-p79425`; pass does not prove thin roots. | RB-003 before transport extraction. | Start S-01 characterization, then S-02. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-11T18:36:58-04:00 | Codex planning/documentation session | No frontend controller or grid-vendor logic exists in `cmd`; only embedded-asset process evidence applies. | Inspected embedded frontend process test and deployable-shape check; touched only tracker. | Source search/read. | Frontend refactor is out of scope; packaged asset behavior is frozen. | RB-001 for test accounting. | Preserve process smoke and use browser validation only when server packaging moves. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-11T18:36:58-04:00 | Codex planning/documentation session | `cmd` directly owns no generated protocol/view artifact; tests consume public contracts and harness outputs. | Inspected contract-owner sections, generated policy, phase maps, build inputs; touched only tracker. | Source searches, `make explain-target`, `make target-plan`. | Owner-input-before-generation rule recorded. | RB-001 and RB-002. | Update authored maps first in S-07; regenerate only through Make. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-11T18:36:58-04:00 | Codex planning/documentation session | Backend-process selects 17 rows across phases 0/1/2/5/10; several command tests lack explicit rows. | Inspected phase maps, target plan, duration baselines, `cmd` tests; touched only tracker. | `make explain-target TARGET=backend-process DETAIL=rows`; `make target-plan TARGET=backend-process`. | Authoritative, supplemental, and support rows mapped; gaps recorded without phase inference. | RB-001, RB-002. | Harness owner selects rows before any test movement. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-11T18:36:58-04:00 | Codex planning/documentation session | Test routes are guarded; process tests cover disabled/fail-closed/token/host/origin behavior but lack explicit accounting. | Inspected Core 04 `REQ-04-109`, harness rules, server wiring, harnessruntime controls, process tests; touched only tracker. | Targeted searches/source reads. | Security freeze map records exact no-exposure posture. | RB-001. | Add canonical characterization/accounting before moving wiring. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-11T18:36:58-04:00 | Codex planning/documentation session | Tracker is ready for a later authorized refactor session. | Touched only `docs/handoffs/command-module-refactor-tracker.md`. | No implementation or broad validation commands. | No production refactor performed. | RB-001 through RB-004. | Resolve accounting/owner blockers, then implement S-01 only. |

Session identity: branch `main`, commit
`f7d69a1d9eb9977d7137d91a47e5a8f29f132d3c`. The worktree was clean before tracker
creation. Retained-run maintenance was not requested, and no `RESULTS_DIR` was supplied.

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Which authored harness owner input should account for the migrate, embedded-frontend, guarded-runtime-route, and object-store-init process tests? | Choosing a phase from filenames or helper reuse would violate evidence-accounting posture and may alter default selection. | Testing Harness owner decision using behavior and target-cost evidence. | BLOCKED: owner/accounting decision required before S-01/S-07. |
| RB-002 | Which semantic owner should receive the Phase F SeaweedFS migration-support tests now mixed into `cmd/operator`? | Recovery, object store, app operator support, and release harness are plausible; the choice changes imports and row accounting. | Review existing operator migration evidence, recovery/object-store APIs, and release artifact consumers. | BLOCKED: owner decision required before S-05. |
| RB-003 | What approved internal platform package should own inherited-listener and graceful-shutdown plumbing? | Creating an arbitrary helper package could produce another shallow boundary. | Repository architecture review of existing platform runtime/HTTP packages and intended public facade. | TODO: required before S-03; does not block S-01/S-02 planning. |
| RB-004 | Do current tests fully characterize invalid inherited FD, listener conversion failure, cancellation timing, and diagnostics writer behavior? | An untested move could change exit status, logging, or listener exposure. | S-01 focused characterization and baseline `make backend-process`. | TODO: characterization required before S-02/S-03. |

No `BLOCKED: owner contradiction` entry is present because no contradiction was found
between inspected owner documents.

## 12. Binary Completion Criteria

- [x] Every tracked file in `cmd` is inventoried, including ignored tracked files and
  redundant placeholders.
- [x] Every discovered public contract risk has an owner and current or required test
  posture.
- [x] Every workflow has dependencies, validation, and a handoff checkpoint.
- [x] Every proposed implementation slice is behavior-preserving; changes require
  later authorization.
- [x] Canonical validation commands were discovered from Make and no command was
  invented.
- [x] No owner contradiction was found; the required contradiction marker policy is
  recorded.
- [x] Framework/repository mismatches are recorded: no `command` module exists, the
  repo requires three executable roots, ignored files affect discovery, and phase rows
  are evidence rather than architecture.
- [x] Handoff tables identify inspected/touched files, commands, results, blockers, and
  the next safe action.
- [x] Generated files are downstream only; no hand edit is proposed.
- [x] This session changes only the tracker and performs no production refactor.
