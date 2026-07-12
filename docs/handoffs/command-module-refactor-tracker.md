# command Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

- Target path: `cmd`.
- Normalized target label: `command`.
- Output path: `docs/handoffs/command-module-refactor-tracker.md`.
- Status: first remediation implemented and validated; forward-design iteration planned in Section 13.
- Authorized change: the approved command-root and runtime remediation plan,
  including production, test, harness, generated-downstream, and documentation edits.
- Non-goals: no domain vocabulary, public HTTP/WS contract, database schema,
  generated protocol contract, or deployable identity change.
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

- `find cmd -type f -name '*.go'` reports 14 authored Go files after remediation.
- `rg --files cmd` discovers all three command roots because root build-artifact
  ignores are anchored and redundant `.gitkeep` files are gone.
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
| `cmd/migrate/main.go` | Process signal context, argument forwarding, stderr binding, and exit propagation for migrations. | Executable `migrate`; package `main`; no reusable API. | `make build-migrate`, dev-service migration flows, browser stack setup, migration drift helpers. | `internal/app.RunMigrateCLIContext`; standard signal and process packages. | `internal/app/migrate_test.go`; non-CWD execution is covered by `migration-scratch-apply`. | Build inputs and deployable-shape evidence; no generated contract directly. | `cmd` thin root plus `internal/app` migrate runner. | low | Legitimate thin composition root. |
| `cmd/operator/main.go` | Process signal context, CLI argument forwarding, stdout/stderr binding, and exit propagation. | Executable `operator`; package `main`; no reusable API. | `make build-operator`, release and operational recovery surfaces, operator process tests. | `internal/app.RunOperatorCLIContext`; standard signal and process packages. | `cmd/operator/operator_phase10_test.go`; `internal/app/operator*_test.go`. | Operator result/progress schema IDs are owned in Core 01 and `internal/app`, not generated here. | `cmd` thin root plus `internal/app` operator runner. | low | Legitimate thin composition root. |
| `cmd/operator/operator_phase10_test.go` | Canonical operator CLI process evidence and object-store initialization smoke. | Process test symbols; no production API. | Phase 10 operator rows and Phase 0 supplemental `E-0-SUPPORT-03`. | `internal/app`, recovery operator operations, config, object store, Postgres/S3 test harnesses, direct SQL and process execution. | The test file and Phase 0/10 map selectors. | Operator JSON/JSONL contract schemas and duration accounting. | Operator executable boundary. | high | Phase F preservation scenarios and helpers moved to recovery-owned release support. |
| `cmd/server/main.go` | Process signal context, standard-stream binding, and final exit propagation. | Executable `server`; package `main`; no reusable API. | `make build-server`, dev stack, browser stack, `processtest`, packaged deployment. | Standard process packages and `internal/app.RunServerContext`. | All `cmd/server` process tests plus app/platform tests. | Build inputs and deployable-shape evidence. | Thin executable root. | low | Import allowlist enforces the remediated boundary. |
| `cmd/server/main_embedded_frontend_process_test.go` | Real-process root HTML and embedded asset smoke. | Process test and asset-path matcher. | Not present in current backend-process phase rows; release deployable-shape check covers adjacent packaging behavior. | HTTP client and Phase 1 process fixture. | The test itself. | Embedded asset archive/build contract; no generated protocol contract. | Server executable-boundary test. | medium | Keep at process boundary; accounting owner is `TODO`. |
| `cmd/server/main_phase0_e2e_test.go` | Startup readiness, invalid diagnostics, first-admin bootstrap, skip/recovery, database, object-store, audit, and secret-safety process evidence. | Phase 0 process tests and shared helpers. | Authoritative Phase 0 `E-0-01` through `E-0-05` backend-process rows. | Auth bootstrap test support, object store, process/config fixtures, Postgres/S3, audit/security assertions, direct SQL. | Selected by Phase 0 map; helpers are reused by later server tests. | Phase 0 accounting and startup diagnostic goldens; no generated protocol contract. | Server process boundary plus owner-local auth/platform assertions. | high | Legitimate black-box evidence; helper ownership should remain explicit. |
| `cmd/server/main_phase10_config_test.go` | Real-process fail-closed validation for `roots.backup_storage`. | Phase 10 process test and environment helper. | Authoritative Phase 10 `E-10-04` backend-process row. | Platform config and shared process/config fixtures. | Selected by Phase 10 map. | Phase 10 accounting; deployment-config contract only. | Server process boundary and config owner. | medium | Correct process-level characterization. |
| `cmd/server/main_phase10_recovery_sentinel_test.go` | Fresh-environment restore followed by real-server workbook access, recovery-route absence, and process fixture support. | Phase 10 process evidence. | Phase 10 `I-10-02` and `E-10-03` backend-process rows. | Recovery/projections APIs for setup plus the real server process and HTTP helpers. | Selected by the Phase 10 map. | Phase E process artifacts and Phase 10 accounting. | Server executable boundary. | high | Direct recovery-only rows moved to `internal/modules/recovery`. |
| `cmd/server/main_phase1_process_test.go` | Login, session, CSRF, WebSocket revocation, TOTP enrollment, credential, and user-administration process smoke. | Supplemental Phase 1 process tests and reusable HTTP/auth helpers. | Supplemental `E-1-SMOKE-01` backend-process row. | Auth test support, authn constants, WebSocket client, process and HTTP fixtures. | Selected by Phase 1 map; helpers reused by Phase 2/5/10 tests. | Phase 1 accounting and duration baselines. | Server process boundary; auth owns behavior. | medium | Evidence only; does not confer auth ownership on `cmd`. |
| `cmd/server/main_phase2_smoke_test.go` | Incident creation/list/patch, workbook preferences, membership administration, extension discovery, and hidden-incident authorization smoke. | Supplemental Phase 2 process tests and incident-create helper. | Supplemental `E-2-SMOKE-01` backend-process row. | Authn, HTTP assertions, process fixture, Phase 1 helpers. | Selected by Phase 2 map. | Phase 2 accounting and duration baselines. | Server process boundary; incidents/auth/extensions own behavior. | medium | Evidence only; no command-domain behavior. |
| `cmd/server/main_phase5_smoke_test.go` | Evidence row creation, blob upload/attach, record patch, Timeline projection query, and no-blob preview failure smoke. | Supplemental Phase 5 process test and HTTP helpers. | Supplemental `E-5-SMOKE-01` backend-process row. | Authn, HTTP/process fixtures, Phase 1/2 helpers. | Selected by Phase 5 map. | Phase 5 accounting and duration baselines; view-schema IDs are consumed, not owned. | Server process boundary; evidence/timeline/projections own behavior. | high | Cross-module black-box evidence; preserve envelopes and projection effects. |
| `cmd/server/main_networkflow_test_runtime_routes_process_test.go` | Real-process composition of the Network Flow harness-control contribution and reset behavior. | Supplemental process test. | Phase 12 supplemental `E-12-SUPPORT-01`, explicit/support-only. | Network Flow harness controls, shared guarded-route process helpers. | Selected explicitly through the Phase 12 map. | Harness support accounting only. | Network Flow process contribution. | medium | Split from generic guarded-route security evidence. |
| `cmd/server/main_test_runtime_routes_process_test.go` | Default-disabled, fail-closed, token/host/origin security, reset, and clock process characterization. | Process tests and harness-route request helpers. | Phase 0 supplemental `E-0-SUPPORT-02`, selected by default for its lower-layer security gap. | Platform harness runtime via server, process fixture, test-route token helpers. | Selected by the Phase 0 map. | Harness route contract and security traceability. | Server process boundary plus harness/platform owner. | high | Network Flow-specific composition is split into the Phase 12 support file. |
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
| T-005 | Resolve unmapped process-test accounting | tests/harness | DONE | T-003 | RB-001; Phase 0/12 supplemental rows and migration target | Every retained process behavior has an authored owner. |
| T-006 | Resolve Phase F migration-support owner | tests/harness | DONE | T-002, T-003 | RB-002; `seaweedfs-migration-preservation` | Recovery owns preservation semantics and release owns publication. |
| T-007 | Implement server runner facade and platform transport seam | implementation | DONE | T-004, T-005 | `internal/app/server.go`; `internal/platform/httpruntime` | Command root is thin and focused unit/process checks pass. |
| T-008 | Split recovery and operator support tests | implementation | DONE | T-005, T-006 | Recovery owner tests and dedicated release-support target | Owner-local tests retain row predicates and artifacts. |
| T-009 | Apply command-tree discovery cleanup | cleanup | DONE | T-007, T-008 | Removed placeholders; anchored root binary ignores | Discovery exposes all 14 Go files and three roots remain. |
| T-010 | Update owner accounting and regenerate outputs | harness | DONE | T-008, T-009 | Authored maps/manifests and Make-generated ledgers/schedules | Shape and drift validation records are in the final handoff. |
| T-011 | Execute final verification and handoff | validation | DONE | T-007, T-010 | S-08 and retained run roots | Required checks pass and intermediate failures are classified below. |
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

### Remediation implementation

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-11 | Codex implementation session | Server assembly is app-owned; HTTP lifecycle is platform-owned; `cmd/server` is a signal/exit root. | `cmd/server/main.go`; `internal/app/server.go`; `internal/platform/httpruntime/**`; boundary manifest/tests. | `make backend-unit`; `make backend-module-boundary-check`; `make backend-process`. | Passed at `.cartulary/test-results/20260711T232059Z-p97008`, `.cartulary/test-results/20260711T231922Z-p76718`, and `.cartulary/test-results/20260711T232616Z-p70476`. | None. | Complete broad validation. |
| 2026-07-11 | Codex implementation session | Direct recovery tests and Phase F preservation support are recovery-owned; only restored real-server evidence remains command-owned. | Recovery owner tests; `cmd/server/main_phase10_recovery_sentinel_test.go`; `cmd/operator/operator_phase10_test.go`; release target/evidence tooling. | `make backend-integration`; `make seaweedfs-migration-preservation`; authored-map generation. | Integration passed at `.cartulary/test-results/20260711T232234Z-p2615`; preservation passed at `.cartulary/test-results/20260711T230737Z-p25405`; ledgers/schedules passed at `.cartulary/test-results/20260711T232214Z-p1988` and `.cartulary/test-results/20260711T232217Z-p2304`. | None. | Validate release gate and generated drift. |
| 2026-07-11 | Codex implementation session | Migration non-CWD evidence and supplemental Phase 0/12 process accounting are authored and discoverable. | Migration drift script/test; Phase 0/10/12 maps; Testing Harness NLSpec; task/topology manifests. | `make migration-scratch-apply`; `make explain-target TARGET=backend-process DETAIL=rows`. | Scratch apply passed; target explanation lists `E-0-SUPPORT-01` through `03` and `E-12-SUPPORT-01`; backend-process contains no Phase F row. | None. | Complete harness and broad validation. |
| 2026-07-12 | Codex implementation session | Final warm validation and retained-run maintenance completed. | Entire remediation diff; generated ledgers/schedules and duration baselines refreshed only through Make-owned tooling. | `make test-fast`; `make check`; `make agent-finalize RESULTS_DIR=...`; final shape, drift, duration, boundary, and Markdown checks. | `test-fast` passed at `.cartulary/test-results/20260711T235351Z-p81047`; full `check` passed 345/345 work units and 1,074 tests at `.cartulary/test-results/20260712T001311Z-p74222`; retained finalization passed at `.cartulary/test-results/20260712T001530Z-p78596`; final drift/boundary checks passed under `.cartulary/test-results/20260712T001608Z-p80750` through `...T001618Z-p82768`. | None. | Handoff complete. |

Resolved validation findings: initial integration compilation exposed a duplicate
test helper name; initial process runs exposed a test-only constant and an operator
fixture dependency; SBOM generation exposed app-import cycles in platform/module
harness tests; the strict release gate exposed stale deleted-file scanning and the
need to retain Phase E verification publication in the real-server row; finalization
required a clean service-backed timing run; staticcheck found one obsolete helper;
and the first full check exposed a shared execution-family manifest collision for
supplemental Phase 0 rows. Each cause was corrected structurally. The final
SeaweedFS release gate passed at
`.cartulary/test-results/20260711T234936Z-p59982`, and the final warm check is green.

Session identity remains branch `main`, starting commit
`f7d69a1d9eb9977d7137d91a47e5a8f29f132d3c`. The tracker was the only pre-existing
worktree addition at implementation start; all other remediation edits belong to this
session. Retained-run maintenance will use a qualifying successful full warm `check`
root if that gate completes.

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Which authored harness owner input should account for the migrate, embedded-frontend, guarded-runtime-route, and object-store-init process tests? | Choosing a phase from filenames or helper reuse would violate evidence-accounting posture and may alter default selection. | Testing Harness owner decision using behavior and target-cost evidence. | RESOLVED: migration non-CWD coverage is consolidated in `migration-scratch-apply`; Phase 0 owns embedded assets, generic guarded routes, and object-store init as supplemental process support; Phase 12 owns the Network Flow contribution. |
| RB-002 | Which semantic owner should receive the Phase F SeaweedFS migration-support tests now mixed into `cmd/operator`? | Recovery, object store, app operator support, and release harness are plausible; the choice changes imports and row accounting. | Review existing operator migration evidence, recovery/object-store APIs, and release artifact consumers. | RESOLVED: recovery owns preservation semantics under `internal/modules/recovery`; release tooling selects the explicit `seaweedfs-migration-preservation` target and its current-run artifact root. |
| RB-003 | What approved internal platform package should own inherited-listener and graceful-shutdown plumbing? | Creating an arbitrary helper package could produce another shallow boundary. | Repository architecture review of existing platform runtime/HTTP packages and intended public facade. | RESOLVED: `internal/platform/httpruntime` owns listener acquisition, inherited-FD conversion, serving, and bounded shutdown. |
| RB-004 | Do current tests fully characterize invalid inherited FD, listener conversion failure, cancellation timing, and diagnostics writer behavior? | An untested move could change exit status, logging, or listener exposure. | S-01 focused characterization and baseline `make backend-process`. | RESOLVED: focused app/platform tests now cover malformed, closed, and non-listener FDs; effective addresses; pre-start cancellation; in-flight and forced shutdown; runtime cleanup; exact diagnostics; writer failures; and exit mapping. |

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
- [x] The authorized remediation preserves public behavior while correcting the
  private cancellation/shutdown race and owner boundaries.

## 13. Next Iteration: Legacy Contraction and Production Profiles

### 13.1 Iteration objective and authority

This section is the current plan for the next command refactor iteration. Sections 3
through 8 remain the historical plan for the completed first remediation and are not
the current implementation queue where they conflict with this section.

The next iteration will make the three executable roots smaller by contracting
unsupported compatibility surfaces, separating production server capabilities from
harness-only capabilities, moving semantic process evidence to its owners, and making
future operator commands additive through one explicit dispatcher. It will not create
a `command` domain module or a fourth deployable identity.

Owner changes required before implementation:

- Core 01 and Core 04 already define exactly five recovery commands and state that
  older recovery aliases are negative-only; implementation must stop treating retired
  top-level names as a second recovery protocol.
- The Testing Harness NLSpec must define a harness-profile build of the existing
  `server` identity before test routes or inherited listeners are removed from the
  production build.
- Migration CLI grammar is implementation-owned. The development and bootstrap
  guides plus harness scripts must adopt the narrowed grammar before code removes the
  broad Goose passthrough.
- No domain vocabulary, public HTTP/WS route, database schema, generated protocol,
  recovery result schema, or deployable identity change is planned.

### 13.2 Current evidence and legacy disposition

Current evidence for this iteration:

- The three `cmd/*/main.go` files are thin, but command process tests still contain
  3,594 lines and directly assemble substantial owner fixtures.
- `RunMigrateCLI` and `RunOperatorCLI` are contextless compatibility wrappers with no
  production callers; the binaries use only their context-aware forms.
- Migration parsing accepts implicit `up`, positional Goose commands, `-command`, and
  arbitrary Goose arguments even though repository consumers invoke explicit
  positional `up`; only database-contract support uses `up-to`.
- Operator dispatch first asks the recovery runner to claim an argument vector and
  then runs a second parser for migration evidence and object-store initialization.
  Help, error output, and command matching therefore have two owners.
- Recovery parsing explicitly recognizes the retired `backup-metadata` top-level name
  only to emit a legacy-shaped rejection. Core 01 gives that name no continuing
  contract value.
- `cmd/operator` retains a raw-Go fallback that invokes nested `make build-operator`;
  canonical harness work already injects and validates the built binary.
- The default production application imports auth, saved-view, Timeline, generic
  harness-runtime, and Network Flow harness-control test route contributors. This
  linkage caused app/module test import cycles during the first remediation.
- `CARTULARY_HTTP_LISTEN_FD` is used by repository process/browser harnesses to avoid
  port-allocation races. No production deployment consumer was found.

| Surface | Decision | Continuing value and removal posture |
| --- | --- | --- |
| Three executable identities and signal-aware context entrypoints | Retain | They are the deployment boundary and already have high cohesion. |
| Contextless `RunMigrateCLI` and `RunOperatorCLI` wrappers | Remove now | No production caller exists; background-context wrappers weaken cancellation guarantees and expand internal API surface. |
| Implicit migrate `up`, `-command`, arbitrary Goose verbs/arguments | Remove now | Repository production consumers use explicit `migrate up`; broad passthrough exposes destructive or accidental behavior without an owner contract. |
| `up-to` through the production migrate binary | Move to database-contract support | Penultimate-version testing is valuable, but it is test mechanics rather than a production command. |
| Retired recovery aliases and alias-specific output assertions | Remove now | Core 01 says they are negative-only and should not become an alternate protocol. Generic unknown-command rejection remains. |
| `migration-evidence capture` | Retain and move to its Postgres migration-evidence owner adapter | It provides current implementation-support evidence and has a distinct semantic owner. |
| `object-store init` | Retain and move to an object-store owner adapter | Stand-up packaging needs idempotent configured-bucket initialization. |
| SeaweedFS legacy-source migration preservation | Retain as time-bounded release support | Existing deployments may still require byte-preserving migration. Keep it quarantined from Phase 10 and add explicit retirement criteria rather than an indefinite compatibility promise. |
| `CARTULARY_HTTP_LISTEN_FD` | Retain only in the harness server profile | It eliminates allocation races in process/browser tests; no continuing production contract was found. |
| Guarded test routes in the production server build | Remove from production; retain in harness profile | Harness routes materially improve verification, but compiling their owners into the deployable server increases coupling and security surface. |
| Embedded frontend and real-process startup/restore checks | Retain, move evidence to semantic assembly owners | They prove deployable integration and cannot be replaced by lower-layer tests. |
| Hidden nested binary builds from Go tests | Remove now | Canonical runtime-binary injection already provides deterministic provenance and avoids recursive Make behavior. |

SeaweedFS migration preservation may be removed in a later iteration only when an
owner document declares the supported source-deployment window closed, release
readiness no longer consumes its artifacts, and retained release evidence proves no
supported upgrade path depends on it. Until then, it remains explicit release-only
support and must not leak back into operator or default local-check ownership.

### 13.3 Gap remediation plan

| ID | Remediation and areas | Rationale and long-term benefit | Compatibility or migration impact | Risk if unresolved | Validation criteria |
| --- | --- | --- | --- | --- | --- |
| N-01 | **Specification, harness, documentation:** add a legacy-surface disposition table to the Testing Harness NLSpec and guides. Record owner, supported caller, retirement trigger, and last allowed execution surface for every retained compatibility mechanism. Delete wording that suggests old recovery aliases or raw-Go nested builds are supported developer workflows. | Compatibility must be an explicit, expiring product decision rather than an accidental property of parsers and tests. A single inventory prevents future phases from reviving removed aliases or adding unowned shims. | Documentation and harness contracts change first; no runtime change in this slice. Historical archives remain untouched. | Legacy branches will continue accumulating because tests make them appear supported, and removal decisions will be repeatedly reopened without evidence. | Harness contract tests reject an unregistered alias/fallback; active docs contain one current grammar; every retained legacy surface has an owner and retirement criterion. |
| N-02 | **Implementation, tests, harness, security:** create production and harness build profiles for the existing `cmd/server` package. The default `build-server` compiles no harness route contributor and rejects harness-only environment keys before listener acquisition. A new Make-owned `build-server-harness` builds the same executable identity with the minimal harness assembly contribution and is the only profile that accepts test-route keys or inherited listeners. Share all product runtime assembly between profiles; only the harness contribution and listener source differ. | Runtime guards are useful, but they do not justify linking test route owners into the production artifact. A build-profile boundary removes app-to-harness/module coupling, reduces production attack surface, and lets future modules contribute harness controls without modifying production assembly. | Browser/process targets and runtime-binary provenance move to the harness artifact. Release, stand-up, development, and deployable-shape targets continue using the production artifact. Setting test-route or inherited-FD keys on the production binary becomes a startup configuration error instead of enabling or silently ignoring behavior. | Test-only routes remain compiled into production; every future module control expands `internal/app`; import cycles and accidental route exposure remain recurring risks. | Binary inspection and negative process tests prove production contains/exposes no harness route family and rejects harness-only keys before binding. Harness-profile tests preserve marker/token/host/origin/reset behavior and inherited-listener race avoidance. Both profiles share identical public route inventories and application behavior. |
| N-03 | **Implementation and tests:** replace the two-stage operator parsing flow with one exact command registry assembled in `internal/app`. Command descriptors use token paths, owner, usage, and a handler interface; startup rejects duplicate or prefix-ambiguous registrations. Recovery supplies its five exact descriptors, Postgres migration evidence supplies one, and object store supplies one. Generate help from the registered descriptors. Remove app-level aliases of recovery result/progress types; process tests decode owner contracts directly. | One dispatcher makes command growth additive and reviewable. Exact registration prevents a broad namespace or parser-order change from capturing future commands, while owner adapters keep semantics outside application assembly. Removing type re-exports narrows coupling between command tests and `internal/app`. | The seven retained commands keep their current semantic behavior. Unknown top-level legacy names become ordinary usage failures and no longer emit a recovery envelope. Unknown subcommands within a canonical recovery namespace continue to use the owner-defined closed invalid-request result where Core 01 requires it. Help formatting becomes registry-derived. | Parser order, duplicated usage text, mixed output behavior, and app type aliases will become more brittle with every future operational command. Retired names may accidentally remain observable contracts. | Registry unit tests cover exact matches, duplicates, prefix ambiguity, unknown top-level commands, canonical-namespace invalid subcommands, stream ownership, and exit codes. The five Phase 10 process commands, migration evidence, and object-store init pass through the same dispatcher. Searches find no `backup-metadata` special case or app recovery type alias. |
| N-04 | **Implementation, tests, documentation:** narrow `migrate` to the exact production grammar `migrate up`. Require the explicit command, reject flags, extra arguments, implicit defaults, and every destructive or arbitrary Goose command with exit `2` before config/database access. Remove `RunMigrateCLI`, the mutable `newMigrateRunnerForCLI` factory, and `migrateCommandFlag`. Move penultimate `up-to` execution into a database-contract-owned test driver used only by migration drift/scratch targets. | Production startup needs forward application, not a general-purpose embedded Goose console. An exact grammar is safer, easier to document, and stable as migration policy grows. Removing global test factories eliminates order-dependent injection. | Existing repository invocations already use explicit positional `up`. Any external use of no-argument migration, `-command`, `up-to`, `down`, `reset`, or arbitrary Goose flags is intentionally unsupported and must migrate to owner-approved tooling. | A production deployable continues exposing destructive implementation commands and compatibility modes with no normative owner; parser behavior remains hard to secure and extend. | Parser tests prove only exact `up` reaches config/database setup. Migration scratch and drift still prove empty and penultimate application outside repository CWD. Searches find no contextless wrapper, command flag, implicit default, or production `up-to`. |
| N-05 | **Tests, harness, ownership:** remove every nested Make/build fallback from Go tests and extend the runtime-binary registry to `server`, `server-harness`, `migrate`, and `operator`. Consumers must receive a normalized, regular, executable, digest-matched binary produced by the declared target. Direct raw `go test` without injection fails fast with a short Make-target instruction rather than building. | Deterministic binary provenance is more valuable than non-canonical convenience. One mechanism removes recursive builds, cache ambiguity, and differences between local and scheduled behavior. | Developers must run public Make targets. Canonical targets remain unchanged except for selecting the proper server profile. | Tests can exercise a different binary than the scheduler recorded, hidden builds can deadlock or bypass cache/provenance policy, and future binaries will copy ad hoc fallback logic. | Harness fixtures cover missing, symlinked, non-executable, stale-digest, wrong-profile, and caller-overridden binaries. Runner logs contain no nested `make build-*` or direct `go run ./cmd/*`. |
| N-06 | **Tests, test support, phase accounting:** move black-box process evidence out of `cmd`. Recovery operator scenarios move to `internal/modules/recovery` using owner-local `testsupport/operatortest`; object-store init moves to `internal/platform/objectstore`; migration evidence stays with Postgres migration evidence; server startup/packaging/guard checks move to `internal/app` or the owning platform package; module smoke assertions move to their semantic modules. Cross-owner fixture mechanics may use registered test-support facades, but product assertions must not move into `internal/testutil`. End state: `cmd` contains only three `main.go` files. | Package placement should communicate ownership even for process tests. Empty command test packages prevent phase growth from turning `cmd/server` into the shared fixture and HTTP-client layer for the monolith. | Phase row IDs and observable assertions remain. Package/file selectors, runtime-binary declarations, fixture budgets, duration identities, and generated ledgers/schedules change. No forwarding test files or phase-named testsupport packages remain. | Another phase will add helpers to the existing 3,594-line command test tree, increasing cross-phase coupling and making owner changes require command-package edits. | Every moved row has one owner and one injected binary. `rg --files cmd` returns exactly three `main.go` files. Phase-map predicates remain one-to-one; backend integration/process and affected Phase 0/1/2/5/10/12 slices pass. Test-support inventory and boundary checks reject product assertions under generic support packages. |
| N-07 | **Static policy, documentation, generated accounting:** enforce the target architecture: no `_test.go` under `cmd`; production app/server profile cannot import harness packages or module harness controls; contextless runner wrappers and legacy CLI tokens are forbidden; command tests cannot execute Make/Go builds; command registries cannot contain duplicate paths. Update active guides and this tracker, regenerate only from authored inputs, and refresh duration baselines only from a qualifying final check. | Architectural cleanup is durable only when the old shape cannot silently return. These checks turn the iteration into a maintained invariant and lower review cost for future command additions. | New static failures are intentional for code that recreates removed behavior. Archives remain historical and are excluded from active-token checks. | The code will regress toward hard-wired app imports, catch-all command tests, and compatibility branches as soon as a new phase or operational command arrives. | Positive/negative fixtures cover each policy. Shape, generated, phase, duration, static, security, release, and Markdown checks pass after Make-owned generation and retained-run finalization. |

### 13.4 Workstreams, sequencing, and exit criteria

#### Phase A — Owner cutover and removal ledger

**Depends on:** none.

- Amend the Testing Harness NLSpec for production/harness server profiles, mandatory
  runtime-binary injection, and removal of the raw-Go fallback.
- Update active guides to the exact migrate and operator grammar.
- Record the time-bounded SeaweedFS migration-support retirement condition.
- Add characterization only for retained outcomes; do not add alias-specific golden
  behavior that would prolong the retired surface.

**Primary risk:** describing a harness server as a fourth deployable or accidentally
promoting implementation-support behavior into product conformance.

**Exit:** exactly three deployable identities remain; both server build profiles and
all retained commands have explicit owners; removed surfaces have migration notes and
no open authority question.

#### Phase B — Deterministic server profiles

**Depends on:** Phase A.

- Add the production/harness assembly split and harness server build target.
- Move inherited-FD parsing into the harness-only listener source while retaining the
  shared `httpruntime` serving/shutdown lifecycle.
- Register the harness server in runtime-binary provenance and migrate process/browser
  consumers atomically.
- Make production startup reject all harness-only keys before listener acquisition.

**Primary risks:** profile drift, tests accidentally exercising harness-only product
behavior, or the production binary binding before rejecting forbidden configuration.

**Exit:** production binary has no harness imports or routes; harness profile preserves
all current test capabilities; public product behavior is identical between profiles.

#### Phase C — CLI contraction and owner command registry

**Depends on:** Phase A; may run in parallel with Phase B.

- Introduce the exact operator registry and owner adapters.
- Remove legacy recovery-name recognition, app contract aliases, contextless wrappers,
  broad migrate parsing, and mutable CLI factories.
- Move penultimate migration execution to database-contract support.
- Standardize retained non-recovery operator commands on one JSON-object-plus-LF
  stdout policy and bounded stderr diagnostics unless an owner contract requires a
  different stream.

**Primary risks:** changing Core-owned recovery envelopes/exit codes, allowing an
unknown command to reach an owner handler, or breaking stand-up object-store init.

**Exit:** one operator dispatcher owns matching/help, only exact retained commands are
registered, production migrate accepts only `up`, and retained process contracts pass.

#### Phase D — Process evidence ownership

**Depends on:** Phases B and C.

- Move process tests and semantic fixtures to the owners listed in N-06.
- Split shared command-package helpers into owner-local clients/fixtures or narrow
  registered facades.
- Delete `cmd` test files rather than leaving forwarding wrappers.
- Update phase maps and runtime-binary declarations before regenerating ledgers and
  schedules.

**Primary risks:** owner test import cycles, lost artifact publication, duplicate row
selection, and duration/shard identity churn.

**Exit:** `cmd` contains three `main.go` files only; every retained process row and
release artifact has one semantic owner and canonical binary producer.

#### Phase E — Guardrails, generation, and final handoff

**Depends on:** Phases B through D.

- Add the static negative fixtures and active-token scans from N-07.
- Regenerate task surfaces, topology, ledgers, and schedules through Make.
- Run narrow owner and profile validation, then the full warm check.
- Run `agent-finalize RESULTS_DIR=<successful-check-root>` and repeat all generated,
  duration, boundary, and shape drift checks.
- Update this tracker with exact final paths, removed surfaces, run roots, failures,
  and retirement status.

**Primary risk:** refreshing duration/generated artifacts from a partial or
contaminated run after large package-selector changes.

**Exit:** all new invariants are enforced, no compatibility shim or stale selector
remains, retained-run maintenance passes, and the tracker has no unresolved blocker.

### 13.5 Validation and acceptance plan

Run from the repository root, narrowing first:

1. `make backend-unit`
2. `make backend-module-boundary-check`
3. `make migration-scratch-apply`
4. `make backend-integration`
5. `make backend-integration-support`
6. `make backend-process`
7. `make build-server`
8. `make build-server-harness` after Phase A registers the target
9. `make build-migrate`
10. `make build-operator`
11. `make phase-slice PHASE=phase0`
12. `make service-backed-slice PHASE=phase10`
13. `make service-backed-slice PHASE=phase12`
14. `make browser-e2e-webserver-backed`
15. `make standup-package-smoke`
16. `make standup-operational-recovery-smoke`
17. `make seaweedfs-migration-preservation`
18. `make seaweedfs-release-gate`
19. `make harness-contract`
20. `make lint-markdown`
21. `make phase-ledgers && make phase-schedules`
22. `make json-shape-check`
23. `make generated-artifact-policy-check`
24. `make generate-drift`
25. `make phase-ledger-drift`
26. `make phase-schedule-drift`
27. `make go-test-duration-baseline-coverage`
28. `make agent-finalize`
29. `make test-fast`
30. `make check`
31. `make agent-finalize RESULTS_DIR=<successful-full-warm-check-root>`
32. Repeat JSON shape, generated policy, phase drift, boundary, and all duration
    coverage/drift checks against the retained check root.

Completion requires:

- `cmd` contains exactly three production `main.go` files and no tests or helpers.
- The release/development production server cannot contain or enable harness routes or
  inherited-listener behavior.
- The harness server is the same product assembly plus one explicit harness
  contribution, not a separate deployable architecture.
- Operator dispatch contains only the seven retained exact commands and has one
  duplicate-safe registry/help owner.
- Production migrate accepts only explicit `up`; penultimate application remains
  covered outside the deployable CLI.
- No contextless runner, mutable CLI factory, legacy recovery top-level name, app
  recovery contract alias, or nested test build remains.
- All public HTTP, WebSocket, recovery, authorization, diagnostic, and generated
  contracts remain unchanged unless Phase A owner amendments explicitly authorize a
  narrower command-only compatibility break.
- SeaweedFS legacy-source support remains release-only with a recorded retirement
  trigger, and no fallback to retired artifact paths is introduced.
- Every generated or duration artifact is refreshed only through Make-owned tooling
  from qualifying evidence.

### 13.6 Next-iteration work tracker

| ID | Work item | Status | Depends on | Exit evidence |
| --- | --- | --- | --- | --- |
| NXT-001 | Adopt legacy disposition and production/harness profile rules | TODO | none | NLSpec/guides validate and no authority question remains. |
| NXT-002 | Add deterministic production and harness server builds | TODO | NXT-001 | Binary provenance and profile-separation tests pass. |
| NXT-003 | Replace two-stage operator parsing with exact owner registry | TODO | NXT-001 | Seven retained commands pass; legacy-specific branches are absent. |
| NXT-004 | Contract migrate grammar and move `up-to` support | TODO | NXT-001 | Production accepts only explicit `up`; scratch/drift remain green. |
| NXT-005 | Remove contextless wrappers, app aliases, global CLI factories, and nested builds | TODO | NXT-003, NXT-004 | Static searches and negative fixtures pass. |
| NXT-006 | Move process evidence out of `cmd` and update owner maps | TODO | NXT-002, NXT-003, NXT-004 | `cmd` contains only three `main.go` files; row accounting is one-to-one. |
| NXT-007 | Add architectural guardrails and regenerate downstream artifacts | TODO | NXT-005, NXT-006 | Shape, generated, phase, boundary, and harness checks pass. |
| NXT-008 | Complete warm validation, retained-run refresh, and handoff | TODO | NXT-007 | Full check and retained finalization pass with recorded roots. |

No implementation blocker is open for this plan. Phase A owner-document amendments
are sequencing requirements, not unresolved design choices.

### 13.7 Plan-authoring validation

| Command | Result | Evidence or note |
| --- | --- | --- |
| `git diff --check -- docs/handoffs/command-module-refactor-tracker.md` | PASS | No whitespace errors in the authored tracker change. |
| `make lint-markdown` | PASS | Active Markdown, including this iteration, satisfies repository lint policy. |
| `make generated-artifact-policy-check` | PASS | Run root `.cartulary/test-results/20260712T002351Z-p90714`; no generated output was hand-edited. |
| `make json-shape-check` | PASS | Run root `.cartulary/test-results/20260712T002351Z-p90710`; authored JSON-shape policy remains valid. |

Implementation, build, phase-slice, release, browser, and broad-check commands were
not run for this planning-only update. They are the ordered acceptance surface in
Section 13.5 and become required as the corresponding workstreams modify code or
harness owner inputs.
