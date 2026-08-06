# app-server Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Item | Current posture |
| --- | --- |
| Target paths | `internal/app/server` and `internal/app/serverprocess` |
| Target label | `app-server` (normalized lowercase kebab case from the target paths) |
| Output path | `docs/handoffs/app-server-module-refactor-tracker.md` |
| Status | Go cache blocker `RB-005` is resolved; `ASW-02` remains rolled back and is ready to restart from the ASW-01 checkpoint; `ASW-03` has not begun |
| Allowed change in `ASW-00` | This tracker file only; later files are admitted exclusively by the next checkpointed workstream |
| Non-goals | No workstream overlap; no hand-edited generated artifact or lockfile; no domain vocabulary change unless owner navigation actually changes; no compatibility layer for unsupported replicated/shared-publication behavior |
| Default posture | Execute the strict sequence in Section 6; prefer the clean future-state contract over preserving unsupported current behavior |
| Repository state at execution start | `main` at `43e74d1f`; tracker already modified by the user, with no other worktree change; the pre-existing tracker diff is preserved and extended in place |

The target exists. `internal/app/server` is the application facade and composition edge used by `cmd/server`; `internal/app/serverprocess` contains retained process-level test evidence and no production package files. The package names are current implementation facts, not proof that `app-server` is a permanent domain-module boundary. A later implementation task must preserve the facade role while deciding, owner by owner, which assembly and adapter responsibilities remain at the application edge.

The source hierarchy used by this tracker is:

1. Adopted subsystem NLSpecs for their named subsystem only.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. Domain vocabulary and implementation-support guides.
5. Current repository code and tests.
6. The planning framework and prior handoffs as evidence only.

Core 05 is not applicable to this tracker because no timed or fixture-sensitive claim is being published. Harness rows and phase maps are evidence accounting only and do not establish runtime architecture.

### NLSpec posture and normative language

**ASR-REQ-001 — authority boundary**

This tracker MUST govern the sequencing, evidence, checkpoints, and exit criteria of the later app-server remediation. It MUST NOT create, replace, widen, or narrow product behavior owned by an adopted subsystem NLSpec or Core 00 through Core 04. An owner-backed behavior restatement in this tracker is binding only because of its cited owner. A remediation decision is binding on later authorized implementation work but does not become implementation-conformance authority until its required owner amendment is adopted.

`docs/research/nlspec-spec.md` supplies writing and completeness discipline only. `temp/analysis-notes.md` supplies recommendations and rationale only. Neither document supersedes the authority hierarchy above.

**ASR-REQ-002 — statement classes and key words**

The tracker uses the following closed statement classes and key words:

| Term | Exact meaning in this tracker |
| --- | --- |
| **Current state** | Descriptive repository evidence observed at the revision commit. It is not authority and MUST be rechecked before implementation. |
| **Owner-backed behavior** | A restatement whose cited adopted owner defines product behavior. A conflict MUST be reported as `BLOCKED: owner contradiction`; implementation MUST NOT choose a side. |
| **Remediation requirement** | A mandatory planning or execution constraint identified by `ASR-REQ-*`. It uses **MUST** or **MUST NOT** and is authorized only inside its checkpointed workstream. |
| **SHOULD / SHOULD NOT** | The named course is required unless the handoff records specific owner evidence or a safety reason that makes it inapplicable. |
| **MAY** | The implementation is permitted, but not required, to select the named option without changing the contract. |
| **Future-only** | Excluded from the current remediation and current conformance. It MUST NOT be activated through a compatibility flag, fallback, or undocumented configuration. |

**ASR-REQ-003 — workstream checkpoints**

Each authoritative remediation workstream in Section 6 MUST be executed as an independent slice. After its validation completes, the implementer MUST update Sections 9 through 12 and every applicable handoff table before beginning a dependent workstream. A workstream MUST NOT be marked `DONE` from planned evidence, an unexecuted command, or a downstream row owned by another workstream.

**ASR-REQ-004 — generated and harness boundaries**

Generated artifacts MUST be changed only through their authored owner inputs and Make-owned generators. Harness rows, selectors, phase maps, and execution topology MUST remain evidence accounting; they MUST NOT be used to infer runtime ownership. Path or selector accounting MUST change only when an authorized implementation slice actually moves the corresponding source or test.

### Owner and support documents inspected

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md` (template and doctrine only).
- `docs/domain.md` (vocabulary and owner navigation only).
- `docs/spec/00_document_set_status_and_precedence.md` through `docs/spec/05_claim_publication_and_benchmark_reproducibility.md`.
- `docs/extension-subsystem-nlspec.md`, `docs/network-flow-activity-nlspec.md`, `docs/reporting-subsystem-nlspec.md`, `docs/report-composition-nlspec.md`, `docs/graph_projection_nlspec.md`, and `docs/testing-harness-nlspec.md`.
- `docs/reference-pack-subsystem-nlspec.md`, inspected as a draft and therefore evidence rather than adopted authority.
- `docs/guides/cartulary-dev-guide.md` for current application-facade, binary-root, server-process-test, and test-support boundaries.
- `docs/research/nlspec-spec.md` for NLSpec writing discipline only and `temp/analysis-notes.md` for recommendation evidence only.

### Repository evidence inspected

- Every file listed in Section 2 under `internal/app/server` and `internal/app/serverprocess`.
- Direct inbound composition and support callers: `cmd/server/main.go`, `internal/testutil/httptestx/httptestx.go`, `internal/testutil/appsupport/commit_fault.go`, and `tools/recoverybrowserrestore/main.go`.
- The Network Flow source-boundary assertion in `internal/modules/networkflow/network_flow_contract_test.go` and server field use in module integration tests.
- Generated-contract consumers and route/OpenAPI parity support under `internal/platform/httpapi`, `internal/platform/contracttest`, `internal/gen`, `internal/app/extensionassembly`, and `internal/platform/webassets` as reached from the target.
- Boundary and harness projections: `tools/backend_module_boundaries.json`, `tools/test_families/app.server.json`, `tools/test_support_inventory.json`, `tools/task_surface_owner.json`, `tools/execution_topology_manifest.json`, and `tools/browser_e2e_batch_manifest.json`.
- Configuration and process-model evidence: `internal/platform/config/config.go`, `internal/platform/config/validation.go`, `internal/platform/config/core_config_test.go`, `configs/dev/config.toml`, `contracts/extensions/specification/shared-protocols.json`, `contracts/extensions/validation/surfaces.json`, and their generated consumer under `internal/gen/contractextensions`.
- Reference Data storage evidence: `internal/modules/reference_data/storage_port.go`, `internal/platform/rootedfs`, the current root/object-store adapters and tests, and the `module.reference_data` owner and verification manifests.
- Runtime caller and harness evidence: `internal/testutil/httptestx/httptestx.go`, `internal/testutil/appsupport/runtime.go`, `internal/testutil/appsupport/commit_fault.go`, `tools/recoverybrowserrestore/main.go`, and all direct `Runtime` field uses found in module integration tests.

The user authorized the complete remediation on 2026-08-05. `ASW-00` changes only this ledger. Every later change remains gated by the immediately preceding workstream's targeted validation, Sections 9–12 checkpoint, applicable handoff-table update, and a passing `make lint-markdown` run.

## 2. Current-State Repository Inventory

All 37 files in the two target directories are inventoried below; none is excluded.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/app/server/collaboration_intents.go` | Translates Jobs and Network Flow transaction intents into Collaboration event intents. | Private adapter methods. | `runtime.go`; owner transaction assembly. | `modules/collaboration`, `modules/networkflow`, platform jobs, `pgx`. | Runtime and module integration coverage through composed mutations. | Collaboration WebSocket event contracts indirectly. | Keep at application composition edge; semantic events remain source-owner/Collaboration owned. | High | Translation must remain in the authoritative transaction; no duplicate or pre-commit publication. |
| `internal/app/server/collaboration_socket.go` | Adapts the platform WebSocket connection to the Collaboration socket port, including frame and origin handling. | Private `collaborationSocket`. | `runtime.go` HTTP dependency composition. | `modules/collaboration`, platform WebSocket transport. | Server process auth/socket coverage. | WebSocket v1 message shapes indirectly. | Keep as transport-adjacent application adapter. | High | Canonical path, authorization recheck, close/error, and event semantics are frozen. |
| `internal/app/server/component_leader.go` | Runs a process-lease-backed component leader loop for replicated staged-object cleanup. | Private `componentLeader`. | `runtime.go`. | `platform/processlease`, contexts, synchronization. | Replicated runtime integration and staged-object tests. | Extension shared-protocol lease projection indirectly. | Retire distributed leadership; retain only an ordinary application lifecycle worker if needed. | High | ASW-05 is blocked by `RB-001` until owner adoption. |
| `internal/app/server/evidence_composition_characterization_test.go` | Characterizes that application composition exposes the narrow Evidence owner runtime rather than its store. | Test function only. | Go test catalog row. | Server runtime and Evidence composition. | Self. | None directly. | `app.server` characterization evidence. | Medium | Preserve until an equivalent narrow-composition assertion exists. |
| `internal/app/server/extensions_publication_characterization_test.go` | Characterizes one immutable publication plan, atomic gate, claim filtering, acknowledgements, and component-loss behavior. | Test functions only. | `app.server` unit rows. | Extensions coordinator, process lifecycle, runtime. | Self. | Generated extension registry and publication projections indirectly. | `app.server` with Extensions collaboration. | Critical | Primary evidence for Stage 6 behavior; structural moves need parity before removal. |
| `internal/app/server/incident_bundle_storage.go` | Implements filesystem-root and shared-object-store bundle staging, publication, read, and removal adapters. | Private storage implementations. | `runtime.go`. | `modules/incidentbundles`, rooted filesystem, object store. | `incident_bundle_storage_test.go`; process portability coverage indirectly. | Incident portability storage/reference contracts indirectly. | Semantic port: `incidentbundles`; byte/path adapter: platform storage; binding: application composition. | High | Persistence-adjacent behavior currently lives in the facade package. |
| `internal/app/server/incident_bundle_storage_test.go` | Characterizes bundle reference rules, lifecycle, cancellation, symlink, root-replacement, and race safety. | Test functions only. | Go test catalog. | Bundle storage adapters, filesystem test support. | Self. | None directly. | Follow the bundle storage adapter. | High | Must move with or continue testing any extracted adapter. |
| `internal/app/server/module_settings.go` | Projects admitted deployment configuration into Collaboration, Evidence, import, portability, reference-data, and HTTP runtime settings. | Private projection helpers. | `runtime.go`. | Config plus affected module option types. | `module_settings_test.go`. | Configuration and resource-limit projections indirectly. | Keep at application composition edge. | Medium | This is legitimate cross-owner configuration translation, not source behavior ownership. |
| `internal/app/server/module_settings_test.go` | Verifies exact module-setting projections and bootstrap reset behavior. | Test function only. | Go test catalog. | `module_settings.go`, config fixtures. | Self. | None directly. | `app.server` unit evidence. | Medium | Add assertions when a newly separated assembly collaborator consumes settings. |
| `internal/app/server/openapi_contract_test.go` | Compares runtime route operations with the generated OpenAPI contract. | Test function only. | `app.server` unit evidence. | `httpapi`, `contracttest`, runtime route diagnostics. | Self. | Generated OpenAPI and route catalog. | `app.server` plus platform OpenAPI owner. | Critical | No structural slice may drop, add, or remap an operation silently. |
| `internal/app/server/publication_controller.go` | Owns the in-memory prepare/acknowledge/commit/serve gate and published projections for one extension serving epoch. | `PublicationState`, `PublicationController`, constructor, lifecycle and projection methods. | `runtime.go`, server tests, publication characterization tests. | Extensions publication plan and process lifecycle. | Extensive publication characterization and runner tests. | Generated extension registry facts indirectly. | Application process/publication orchestration; exact package placement deferred. | Critical | Keep behavior; do not rederive claims, routes, workspaces, workers, or contracts independently. |
| `internal/app/server/publication_plan_agreement.go` | Uses PostgreSQL coordination and object storage to admit one replicated publication-plan digest and renew its marker. | Private agreement and `Close`. | `runtime.go`. | PostgreSQL advisory locking, object store, time. | Replicated runtime integration. | Extension lease/publication projections. | Retire from the current-major serving path. | Critical | ASW-05 MUST ensure legacy markers cannot authorize or block a current process; work remains blocked by `RB-001`. |
| `internal/app/server/reference_pack_storage.go` | Implements filesystem-root and shared-object-store Reference Data staging, publication, read, and removal adapters. | Private storage implementations. | `runtime.go`. | `modules/reference_data`, rooted filesystem, object store. | `reference_pack_storage_test.go`. | Reference-pack contracts indirectly. | Move the filesystem adapter to `internal/app/referenceassembly`; delete the unsupported object-store adapter. | High | ASW-05 deletes shared publication; ASW-06 moves the supported adapter without changing the port. |
| `internal/app/server/reference_pack_storage_test.go` | Characterizes reference validation, lifecycle, cancellation, symlink, and root-replacement behavior. | Test functions only. | Go test catalog. | Reference-pack storage adapters and filesystem support. | Self. | None directly. | Follow the filesystem adapter into `referenceassembly`; delete unsupported object-store cases. | High | Preserve unique immutable publication, containment, cancellation, and root-free references under ASR-AC-201..ASR-AC-206. |
| `internal/app/server/runtime.go` | Admits configuration and assembles storage, leases, telemetry, extensions, jobs, collaboration, revision/timeline/evidence/import/entity/indicator/network-flow/reporting/workbook services, routes, publication, and cleanup. | `Options`, `Runtime`, `RuntimeSettings`, `NewRuntime`, `ActivatePublication`, `PublishedComponentLost`, `FatalEvents`, `Close`; many exported `Runtime` fields. | `server.go`, `httptestx`, recovery browser tool, module integration tests. | Most application assemblies, domain modules, and platform runtime/storage/transport packages. | `runtime_test.go`, `runtime_integration_test.go`, route/publication/staged tests, serverprocess suites. | Generated extension registry, route/OpenAPI, view schema, web asset, and protocol contracts indirectly. | Keep a narrow `internal/app/server` facade; split private assembly collaborators by owner boundary. | Critical | Approximately 1,600 lines and the principal mixed-responsibility seam. Public fields are an internal-repo compatibility surface. |
| `internal/app/server/runtime_integration_test.go` | Exercises real PostgreSQL/object storage config admission, bootstrap, serving/extension leases, replicated fencing, and recovery behavior. | Test-only exported helpers/types: `StartupCounters`, `AuditEvent`, `IntegrationEnv`, `BindPostgres`, `BucketName`. | `app.server` integration rows. | PostgreSQL and S3 test harnesses, config/bootstrap, runtime. | Self. | Extension lease/publication projections indirectly. | `app.server` integration evidence; reusable fixtures may move to `internal/testutil/appsupport`. | Critical | Contains the overlapping-runtime assertion implicated by `RB-001`. |
| `internal/app/server/runtime_routes.go` | Defines fixed built-in route contributions and validates extension route bindings against the publication catalog. | Private route helpers/constants. | `runtime.go`; source-boundary test reads the file text. | `httpapi`, Extensions route publication, owner route registrars. | `runtime_routes_test.go`; Network Flow source-boundary assertion. | Generated HTTP contract catalog and extension route descriptors indirectly. | Keep at application route-composition edge, possibly as a cohesive collaborator. | Critical | File and symbol placement is coupled to a brittle source-text boundary test. |
| `internal/app/server/runtime_routes_test.go` | Verifies route membership/order, exact catalog projection, contribution validation, and reverse/idempotent cleanup. | Test functions only. | `app.server` unit evidence. | Runtime routes, HTTP contract catalog, cleanup helpers. | Self. | Generated route catalog indirectly. | `app.server` route-composition evidence. | Critical | Characterizes exact route composition rather than owner route behavior. |
| `internal/app/server/runtime_test.go` | Verifies fail-closed ordering, secret/manifest preflight, borrowed resources, configuration, and cleanup. | Test-only `CloseTrackingStore`, `RuntimeConfig`. | `app.server` unit rows. | Runtime seams, config fixtures, storage fakes. | Self. | Generated owner manifests indirectly. | `app.server` unit evidence; reusable fixtures may move to test support. | High | Preserve startup order and borrowed-resource ownership invariants. |
| `internal/app/server/server.go` | Binary-facing process facade: config load, runtime construction, listener serving, publication activation, diagnostics, shutdown, and exit-code mapping. | `RunServerContext`. | `cmd/server/main.go`; server unit tests. | `configassembly`, `extensionassembly`, `httpruntime`, process lifecycle/lease, HTTP diagnostics. | `server_test.go`, process suites. | Public route diagnostics and generated inactive-policy facts indirectly. | Keep as the thin application facade for `cmd/server`. | Critical | `cmd/server` must remain a composition root with no feature logic. |
| `internal/app/server/server_profile_harness.go` | Build-tagged harness server profile that contributes authorized test routes and inherited-listener behavior. | Private profile implementation. | `server.go` under `cartulary_harness`. | Auth, Saved Views, Network Flow harness control, harness runtime, HTTP runtime. | Harness profile test and serverprocess harness-route suites. | Harness route contracts. | Keep as the sole allowed harness contributor in production package. | Critical | Must remain build-tagged and unavailable in the production profile. |
| `internal/app/server/server_profile_harness_test.go` | Verifies the exact harness-route enablement value. | Test function only. | Harness-tagged unit execution. | Harness server profile. | Self. | Harness contract indirectly. | `app.server` harness-profile evidence. | High | Preserve exact enablement and build-tag isolation. |
| `internal/app/server/server_profile_production.go` | Production server profile; rejects harness-only environment and delegates ordinary serving. | Private profile implementation. | `server.go` in production builds. | Platform config and HTTP runtime. | `server_test.go`; serverprocess harness isolation. | None directly. | Keep at application process facade. | High | Production must fail closed when harness-only variables are present. |
| `internal/app/server/server_test.go` | Verifies runner diagnostics, exit mapping, cancellation, listener/publication failure, lease loss, cleanup, and production harness-env rejection. | Test functions and private writer fake. | `app.server.support_unit` and unit rows. | Server runner fakes, process lifecycle, config diagnostics. | Self. | Route diagnostics indirectly. | `app.server` process-facade evidence. | Critical | Primary characterization for process outcomes without spawning a process. |
| `internal/app/server/shared_publication_storage.go` | Shared bounded object-store byte helper used by bundle and reference-pack adapters. | Private helper. | Two storage adapter files. | Platform object store and byte limits. | Storage adapter tests indirectly. | None directly. | Delete with the unsupported shared adapters in ASW-05. | Medium | It has no supported current-major consumer after filesystem-only admission. |
| `internal/app/server/staged_objects_test.go` | Characterizes staged-cleanup readiness degradation/recovery and fatal integrity handling. | Test functions only. | `app.server` unit evidence. | Staged Objects health, runtime lifecycle. | Self. | Extensions staged-object contracts indirectly. | `app.server` composition evidence with `stagedobjects` ownership. | High | Does not make server the staged-object behavior owner. |
| `internal/app/server/test_dependencies.go` | Exposes a test-only PostgreSQL DB decorator through the HTTP dependency set, guarded by test-runtime detection. | `DependencySetWithPostgresDBDecoratorForTesting`. | `internal/testutil/appsupport/commit_fault.go`. | `httpapi`, platform Postgres, test-runtime guard. | Commit-fault integration paths. | None directly. | Replace with a narrower test-support seam when feasible; keep until callers migrate. | High | Test-only assumption currently leaks through an exported production-package symbol. |
| `internal/app/serverprocess/config_test.go` | Black-box process tests for effective configuration failure and secret-safe diagnostics. | Test functions only. | `app.server.process` rows. | Shared process harness, config fixtures. | Self. | Deployment config contract indirectly. | Retain as process evidence. | High | No production code belongs in this package. |
| `internal/app/serverprocess/e2e_test.go` | Broad process smoke coverage for health, login/session/CSRF, incidents, memberships, workbook preferences, extensions, and route isolation. | Test functions and local helpers only. | Multiple `app.server.process` rows. | Built server harness, HTTP client, PostgreSQL fixtures. | Self and shared process harness. | HTTP, WebSocket, view, extension, and auth contracts indirectly. | Retain as process evidence; extract helpers only if semantically reusable. | Critical | Broad behavior freeze; not a runtime module design. |
| `internal/app/serverprocess/embedded_frontend_process_test.go` | Verifies the packaged standalone server serves embedded frontend assets. | Test function only. | Full-tier process row. | Process harness and HTTP client. | Self. | Packaged web assets and client-support registry indirectly. | Retain as application packaging evidence. | High | Frontend implementation remains out of scope. |
| `internal/app/serverprocess/evidence_process_test.go` | Verifies evidence creation/upload/attach and timeline projection refresh through the real process. | Test function only. | Full-tier process row. | Process harness, Evidence/Timeline public routes, object storage. | Self. | Evidence, route, and view contracts indirectly. | Retain as cross-owner process evidence. | Critical | Freezes post-commit projection behavior without assigning it to server. |
| `internal/app/serverprocess/incident_membership_process_test.go` | Verifies membership lifecycle and authorization through the real server. | Test functions and local request helpers only. | Process smoke rows. | Process harness, incident/auth routes. | Self. | HTTP authorization contracts indirectly. | Retain as process/security evidence. | Critical | Authorization remains owner-route behavior composed by server. |
| `internal/app/serverprocess/networkflow_runtime_routes_process_test.go` | Verifies standalone-server Network Flow route composition and behavior. | Test function only. | Full-tier process row. | Process harness and Network Flow routes. | Self. | Generated extension and Network Flow route contracts indirectly. | Retain as application route-composition evidence. | Critical | Network Flow semantics remain owned by its adopted NLSpec/module. |
| `internal/app/serverprocess/process_test.go` | Shared process harness for binaries, environment, PostgreSQL/object-store fixtures, readiness, requests, and cleanup. | Test-local harness surface within package. | All serverprocess test files. | Built server/migrate artifacts, service fixtures, HTTP/SQL clients. | All serverprocess suites. | Harness execution topology indirectly. | Retain process-specific support; move only reusable composition to `internal/testutil/appsupport`. | High | Central test support file; helper movement may require test-catalog selector review. |
| `internal/app/serverprocess/recovery_sentinel_test.go` | End-to-end retained backup/restore sentinel and absence-of-public-recovery-route evidence. | Test functions and specialized recovery fixture helpers. | Recovery process row. | Server/migrate/operator tooling, backup/restore, SQL/object storage. | Self. | Recovery and incident portability contracts indirectly. | Retain specialized process evidence; defer helper extraction. | Critical | Large test-only orchestration is legitimate evidence unless a reusable seam is proven. |
| `internal/app/serverprocess/runtime_routes_process_test.go` | Verifies public and harness-only runtime route availability, authorization, reset, and fault controls. | Test functions only. | Process rows. | Production and harness binaries, HTTP client. | Self and shared process harness. | Harness and public route catalogs indirectly. | Retain as route/profile process evidence. | Critical | Must preserve separation between production and harness profiles. |
| `internal/app/serverprocess/shared_process_harness_test.go` | Supplies shared request/response helpers for process suites. | Test-local helpers only. | Other serverprocess tests. | HTTP client and test assertions. | Package process suites. | None directly. | Retain or move reusable portions to `internal/testutil/appsupport`. | Medium | Do not create a generic helper package without semantic reuse. |

## 3. Module Boundary Diagnosis

The current target is a legitimate application/service facade and a mixed-responsibility package. It is also transport-adjacent, persistence-adjacent, a view/projection orchestration edge, and a mutation/publication coordinator. It is not evidence for a new `app-server` domain owner. Source behavior remains with its named modules and adopted owners.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Binary startup, diagnostics, exit mapping, and shutdown | `server.go` and profile files | `internal/app/server` | keep | `cmd/server/main.go`, developer guide, server tests | Legitimate thin facade behavior. |
| Full runtime composition | `runtime.go` | `internal/app/server` plus existing assembly packages | split | 1,600-line runtime and broad dependency fan-out | Keep `NewRuntime` facade initially; extract private cohesive assembly steps without changing order. |
| Module settings projection | `module_settings.go` | Application composition edge | keep | Exact config-to-owner DTO translation and tests | Do not move owner behavior into config/platform code. |
| HTTP route catalog and extension route binding | `runtime_routes.go` | Application route composition; route semantics remain owner-owned | keep / split | OpenAPI parity, exact catalog tests, Network Flow assertion | Preserve catalog order, IDs, reserved routes, and registration. |
| WebSocket connection adaptation | `collaboration_socket.go` | Application transport adapter; Collaboration owns semantics | keep | Developer guide and process socket tests | No second event or mutation API. |
| Transactional intent translation | `collaboration_intents.go` | Application adapter between source owners and Collaboration | keep / split | Source-owner transaction injection in runtime | Must remain post-commit sequencer input through the authoritative transaction. |
| Extension publication gate | `publication_controller.go` | Application process orchestration consuming Extensions plan | keep / split | Extensions Stage 6 characterization | Exact package move is unnecessary until runtime decomposition proves a seam. |
| Replicated publication agreement | `publication_plan_agreement.go` | No current-major owner; retire after owner adoption | move / defer | `EXT-REQ-213` and `REQ-04-145` permit one active process; current replicated implementation widens them | ASW-05 MUST retire current-major agreement semantics after ASW-01 and ASW-02. |
| Replicated component leadership | `component_leader.go` | Ordinary application lifecycle worker, using generic platform lease mechanics only where owner-backed | move / defer | One active process makes distributed staged-cleanup leadership unnecessary | ASW-05 MUST replace leader semantics; no change may precede owner adoption. |
| Incident bundle storage adapters | `incident_bundle_storage.go` | `incidentbundles` port + platform storage adapter + app binding | move / split | Implements owner storage port using rootedfs/objectstore | Characterize paths/references before moving. |
| Reference Data storage adapters | `reference_pack_storage.go` | `reference_data` semantic port + `internal/app/referenceassembly` concrete adapters + app binding | move | Core 01 owns Reference Pack semantics; Core 04 owns roots/configuration; the draft subsystem NLSpec is unnecessary for a behavior-preserving move | ASW-06 MUST preserve the exact port and storage outcomes while moving the adapters. |
| Shared publication byte helper | `shared_publication_storage.go` | Platform only for wholly owner-neutral bounded-byte mechanics; otherwise owner assembly | split / defer | Shared only by two unrelated owner adapters | ASW-06 MUST apply the closed neutrality test in ASR-REQ-204; no owner grammar or lifecycle may move to platform. |
| Timeline, entities, indicators, evidence, imports, links, revisions, projections, saved views, reporting, and Network Flow behavior | Assembly calls in `runtime.go` | Existing named modules and assembly packages | move only if misplaced logic is found; otherwise keep delegation | Runtime constructs their facades/providers; owner modules implement behavior | No evidence supports transferring these responsibilities to `app-server`. |
| Test-only DB decorator | `test_dependencies.go` | Typed incident-create fault capability in test support | move / split | One exported caller in `appsupport`; string-keyed untyped override crosses production package API | ASW-07 MUST preserve commit-failure evidence while removing the exported helper, string key, SQL-text observation, and generic override path. |
| Black-box process verification | `internal/app/serverprocess` | Retained application process evidence | keep | AGENTS/developer guide and test manifest | Reusable helpers may move; process assertions stay. |
| Frontend shell/controller state | Not present in target | `apps/web` | defer / intentional/no_action | Only embedded asset serving and extension workspace projection are composed | Freeze indirect contracts; no frontend refactor is planned. |
| Grid-vendor integration | Not present in target | `packages/grid-adapter` | intentional/no_action | No direct vendor import found | Preserve import boundary; no grid-adapter slice is supported. |

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `RunServerContext` process facade, diagnostics, and exit codes | Core 04 and application facade | `server.go` | `server_test.go`, config/process suites | Preserve existing runner tests; add only if a new startup stage is introduced | Critical | Startup failures remain exit `2`; fatal integrity loss remains exit `70`; cancellation semantics stay fixed. |
| `NewRuntime`, `Options`, `Runtime`, `RuntimeSettings` | Application composition | `runtime.go` and inbound callers | Runtime unit/integration and `httptestx` users | ASW-04 preserves the entire surface; ASW-07 installs the exact ASR-REQ-302 lifecycle API only after caller characterization | Critical | Two-stage migration; borrowed ownership and idempotent close remain frozen. |
| Runtime cleanup order and resource ownership | Application process lifecycle | `Runtime.Close`, cleanup stack | Runtime and route cleanup tests | Preserve reverse order, borrowed-resource non-closure, and idempotence | Critical | No resource leak or double close. |
| HTTP `/api/v1/*` route shapes, envelopes, IDs, order, and reserved families | Core 01 and route owners | `httpapi.ContractOperations`, built-in contributions | OpenAPI parity, route unit/process tests | Exact catalog parity before and after every route-composition move | Critical | No route semantics may be inferred from package names. |
| OpenAPI/generated HTTP contract parity | Core 01; generated projections downstream | Generated OpenAPI consumed by `contracttest` | `openapi_contract_test.go` | Existing parity is required; no generated hand edit | Critical | Owner inputs precede generation if a later authorized behavior change occurs. |
| WebSocket `GET /ws/v1/incidents/{incident_id}` and event semantics | Core 01/Core 03/Collaboration | Socket adapter, hub/dispatcher/intents | Process auth/socket and collaboration tests | Preserve origin, session/membership recheck, resume, sequencing, close, and event shapes | Critical | HTTP remains the query/mutation path; WS is not a second mutation API. |
| Transactional `record_changed`, job progress, and extension-resource intents | Source owners and Collaboration | `collaboration_intents.go` | Module integration and process smoke tests | Add adapter-focused tests if extraction changes transaction injection | Critical | Rejected/rolled-back work emits no durable intent. |
| Workbook row/query/mutation and conflict behavior | Core 01–03 and owner modules | Route dependency assembly and workbook/revision facades | Process smoke, module integration, browser support row | Existing owner tests remain primary; server needs only composition characterization | Critical | Server does not become the current-envelope or mutation owner. |
| Saved-view, view-schema, and workbook preference behavior | Core 01–03, Saved Views, platform View Schema | Runtime route/dependency composition | Process smoke, harness routes, owner tests | Preserve route registration and generated schema consumption | High | No duplicate app-server row or view-schema logic was found. |
| Revision/change-set, rollback, conflict-token, and current-envelope behavior | Core 02–03 and Revisions/source owners | `revisionassembly.Runtime` injection | Runtime manifest preflight, process and module tests | Characterize only altered composition seams | Critical | Revisions coordinates history but does not absorb source-owner authority. |
| Projection refresh and timeline/evidence visibility | Core 01–03 and Projections/source owners | Timeline assembly and projection catalog wiring | Evidence process test and module integration | Preserve post-commit refresh assertions | Critical | Projection tables remain disposable derived state. |
| Extension discovery, claims, routes, workspaces, workers, and Stage 6 gate | Extensions NLSpec with Core 00/01/04 | Generated coordinator and publication controller | Publication characterization, route/process/browser rows | Existing one-epoch and atomic-gate evidence plus ASR-AC-109 | Critical | ASW-05 may change only the unsupported replicated path after ASW-01 adoption. |
| Application-process and serving leases | Core 04 and Extensions owner | Runtime lease branches and processlease | Current single/replicated integration and recovery sentinel | ASR-AC-101..ASR-AC-108 and ASR-AC-110 | Critical | Planning selects single-active; current replicated code remains evidence of a mismatch, not a behavior to preserve. |
| Incident bundle storage and publication | Core owners and Incident Bundles module | Root adapter plus unsupported shared adapter | Storage unit and recovery/portability process tests | Preserve filesystem behavior under ASR-AC-207..ASR-AC-210; delete shared publication in ASW-05/ASW-06B | High | ASW-06B moves only the supported filesystem adapter. |
| Reference Data storage and publication | Core 01 semantic behavior; Core 04 roots/configuration; draft Reference Pack evidence only | Root adapter plus unsupported shared adapter and `reference_data.Storage` | Storage unit tests and `module.reference_data` owner rows | ASR-AC-201..ASR-AC-206 on the filesystem adapter | High | ASW-05 deletes the unsupported adapter; ASW-06 moves the filesystem adapter without promoting the draft NLSpec. |
| Authorization and route admission | Core 04 and route owners | HTTP dependency composition; reserved-route catalog | Membership, auth/session/CSRF, extension discovery process tests | Preserve hidden `404`, `401`, `403`, CSRF, and deployment-admin isolation through owner tests | Critical | No new application-layer authorization policy is planned. |
| Embedded frontend assets and client support registry | Core 01 and web packaging | Server handler composition and webassets registry | Embedded frontend process test | Preserve packaged asset and fallback behavior | High | No frontend-controller change is in scope. |
| Harness-only routes, test clock/reset/fault controls | Testing Harness NLSpec and named owners | Build-tagged harness profile | Harness profile and runtime-route process tests | Preserve build-tag and exact enablement isolation | Critical | Production profile must expose none of these routes. |
| Network Flow source boundary assertion | Network Flow owner and backend boundary policy | `network_flow_contract_test.go` reads two server source files | Source-text authorization-boundary test; process route test | In ASW-04 replace the server filename assertion with an app-server catalog assertion that `networkflow.RouteContributionID` is admitted exactly once; retain Network Flow owner-local authorization evidence | High | Semantic evidence survives structural movement and tests the real ownership relationship. |
| Harness owner/row accounting | Testing Harness NLSpec | `tools/test_families/app.server.json` with 32 rows | Catalog checks and Make target selection | Update owner inputs only if selectors/packages move; never use rows as architecture | High | Current families: 1 browser support, 9 integration, 14 process, 1 support unit, 7 unit. |

### 4.1 Application-process and lease remediation contract

**ASR-REQ-101 — current-major process cardinality**

Extensions contract major `1` MUST permit exactly one active, ready, serving Cartulary application process per deployment. Infrastructure MAY provision additional instances for replacement, restart, or standby, but only the holder of the deployment-global exclusive `application_process_lease` may enter Stage 1 or any later migration, claim resolution, publication, listener, readiness, workspace advertisement, worker activation, or job-dequeue stage. Concurrent active serving and mixed-build serving are prohibited.

**ASR-REQ-102 — lease state, defaults, and terminal outcomes**

The application lease state machine MUST remain exactly:

```text
unacquired -> acquiring -> held -> uncertain -> held
                                      \-> lost
held -> released
```

| Input or transition | Default and bounds | Required observable outcome |
| --- | --- | --- |
| Initial acquisition | `timeouts.extensions.process_lease_acquire_seconds`; default `30`; integer `1..300` | Timeout performs no Stage 1+ side effect, emits `extension_application_process_active`, and exits `2`. |
| Loss of ownership proof | Immediate | Transition `held -> uncertain`; close readiness and every HTTP, WebSocket, publication, workspace, worker, and job admission gate. |
| Continuous ownership proof | Before the loss-detection deadline | Only the original underlying session may transition `uncertain -> held`; a recreated or replacement session MUST NOT do so. |
| Confirmed loss or deadline expiry | `timeouts.extensions.process_lease_loss_detection_seconds`; default `5`; integer `1..30` | Transition to irreversible `lost`, emit `application_process_lease_lost`, execute fatal drain, and exit `70`; same-process reacquisition is forbidden. |
| Fatal drain | `timeouts.extensions.shutdown_drain_seconds`; default `30`; integer `1..300` | Preserve committed state and durable queued work, reject new work, then force remaining termination at expiry. |
| Orderly shutdown | From `held` only | Release the underlying lease and owned resources without a false fatal signal. |

**ASR-REQ-103 — recovery serving-lease separation**

The application-process lease and recovery serving lease MUST retain separate typed identities and owner-visible outcomes. Their typed wrappers MAY share owner-neutral session, advisory-lock, renewal, and continuity mechanics from `internal/platform/processlease`:

| Lease | Purpose | Mode and holder | Failure effect |
| --- | --- | --- | --- |
| `application_process_lease` | Select the one process permitted to start and serve a deployment. | Deployment-global and exclusive; one application process for its full serving lifetime. | Uncertainty closes all admission; confirmed loss is whole-process fatal. |
| Recovery serving lease | Exclude application traffic while an admitted recovery operation mutates or verifies a restore target. | Application startup holds the shared side; Recovery holds the exclusive side. | Active exclusivity reports `recovery_serving_lease_active`, keeps listener startup closed, and exits `2`; loss while serving reports `recovery_serving_lease_lost`, closes readiness/admission immediately, performs fatal drain, and exits `70`. |

The wrappers MUST NOT share a lock identifier, semantic owner, mode, or failure vocabulary. Recovery's exclusive target lease remains a third owner-specific contract and MUST retain its existing Recovery failure mapping.

**ASR-REQ-104 — adopted-owner correction**

ASW-01 MUST preserve the substance of Extensions `EXT-REQ-213` and Core 04 `REQ-04-145`. It MUST amend Core 04 `AC-404` so one application deployable may have multiple provisioned instances but no more than one active application process, and it MUST cross-reference the application-process lease contract. The development guide MUST cease advertising replicated serving for the current major. No implementation remediation under ASR-REQ-105 or ASR-REQ-106 may begin before this owner correction is adopted.

**ASR-REQ-105 — projection and configuration closure**

The authored extension projections MUST express the following closed current-major values and MUST remove any emitted active alternative:

| Projected fact | Required value |
| --- | --- |
| Process model | `single_active` |
| Lease scope and mode | Deployment-global, exclusive |
| Holder cardinality | Exactly one |
| Loss scope | Whole process |
| Recovery from `uncertain` | Original continuous underlying session only |
| Recovery from `lost` | Prohibited |
| Non-holder behavior | No Stage 1+ side effects; bounded startup failure |
| Concurrent serving | Prohibited |
| Component-scoped replicated serving lease | Prohibited |

No adopted owner defines `application.process_model` in the closed deployment-configuration registry. ASW-02 and ASW-05 therefore MUST remove the key from configuration and runtime code. A supplied file or environment-overlay key MUST fail with `error.code='invalid_deployment_config'`, item `path='application.process_model'`, and `reason_code='unknown_key'`. There MUST be no alias, compatibility default, silent ignore, or fallback. Replicated projection values have no identified persisted or external compatibility obligation; they MUST be removed rather than recognized as an active legacy mode. Generated artifacts MUST be regenerated from authored inputs and MUST NOT be hand-edited.

**ASR-REQ-106 — current implementation disposition**

| Current element | Required disposition after ASW-01 through ASW-05 |
| --- | --- |
| Replicated branch in `runtime.go` | Remove or make unreachable through the closed admitted configuration shape. |
| `publication_plan_agreement.go` | Retire from current-major serving. Legacy marker state, if retained for cleanup, MUST neither authorize nor block a current process. |
| `component_leader.go` | Remove distributed-leader semantics. Staged-object cleanup becomes an ordinary lifecycle-owned worker in the single active process. |
| `RuntimeSettings.ProcessModel` | Remove after caller migration; it MUST NOT remain an extension point. |
| Overlapping-runtime integration test | Replace the success assertion with the negative single-active acceptance cases in Section 12. |
| Required published component loss | Retain whole-process fatal handling; no component continues serving independently. |

**ASR-REQ-107 — future-only active-active boundary**

Active-active serving, component-scoped serving leases, publication epochs, mixed-version admission, mixed-build rolling upgrades, stale-writer fencing, and independent component failover are future-only. A future design MUST use a new contract major and define mutation/job/publication fencing, singleton ownership, failover, traffic admission, compatibility, data migration, and partition/stale-holder conformance before activation. Current code MUST NOT expose these behaviors behind an undocumented option.

PostgreSQL advisory locking is a non-normative implementation option when a dedicated session preserves the owner state machine. Kubernetes rollout strategies and generic leader election are deployment rationale only; they do not alter the lease contract. See [PostgreSQL advisory locks](https://www.postgresql.org/docs/18/functions-admin.html), [client-go leader election](https://pkg.go.dev/k8s.io/client-go/tools/leaderelection), and [Kubernetes Deployment strategies](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/).

### 4.2 Reference Data storage remediation contract

**ASR-REQ-201 — semantic and adapter ownership**

The Reference Data boundary MUST be:

| Concern | Required owner or package | Prohibited responsibility |
| --- | --- | --- |
| Opaque logical references and semantic storage port | `internal/modules/reference_data` | Filesystem roots, service endpoints, application configuration, or adapter selection. |
| Rooted-filesystem realization | `internal/app/referenceassembly` | Reference Pack lifecycle, activation, verification, refresh, authorization, job, audit, or public error ownership. |
| Generic root containment and bounded byte I/O | Existing `internal/platform/rootedfs` package | Reference Pack reference grammar, digest meaning, staging, publication, or lifecycle. |
| Configuration projection, resource ownership, construction, and injection | `internal/app/server` | Storage mechanics or Reference Pack domain behavior. |

The package name MUST be `referenceassembly`; no `referencepackassembly` alias or compatibility wrapper may be introduced. The draft Reference Pack subsystem NLSpec MUST NOT be promoted merely to authorize this behavior-preserving package move.

Core 04 permits only filesystem roots for Reference Pack storage, temporary work, and export outputs in the current contract major. `sharedReferencePackStorage`, its object-store tests, and any shared-publication helper used only by that unsupported path MUST be deleted, not moved or deprecated. `referenceassembly.NewRootStorage(temporaryRoot, publishedRoot)` MUST return a closable concrete adapter implementing the unchanged port.

**ASR-REQ-202 — preserved semantic port**

ASW-06 MUST preserve this interface exactly unless a separate behavior-changing task is authorized:

```go
type Storage interface {
    Stage(context.Context, string, []byte) (StagingRef, error)
    Publish(context.Context, string, []byte) (StorageRef, error)
    ReadStaged(StagingRef, int64) ([]byte, error)
    ReadPublished(StorageRef, int64) ([]byte, error)
    RemoveStaged(StagingRef) error
    RemovePublished(StorageRef) error
}
```

**ASR-REQ-203 — method and reference behavior**

| Operation or invariant | Required behavior |
| --- | --- |
| `Stage` | Honor cancellation before or during the write, publish no partial object, and return an opaque root-free `StagingRef`. The owner layer remains responsible for admitted input-size and verification policy. |
| `Publish` | Require a lowercase 64-character SHA-256 token, honor cancellation, publish immutable bytes atomically, and return a new opaque root-free `StorageRef`. Lifecycle verification and byte-to-digest proof remain owner-service responsibilities. |
| Repeated publication | Each successful call creates a distinct immutable reference and MUST NOT overwrite or reuse a mutable destination, including when the supplied digest token repeats. |
| `ReadStaged` / `ReadPublished` | Return the addressed bytes only when the reference is valid and the result is within the supplied positive maximum; over-limit or unsafe targets fail closed. |
| `RemoveStaged` / `RemovePublished` | Remove only the addressed regular object and succeed when it is already absent. |
| Logical references | Expose no absolute host path, storage root, raw service URL, endpoint, credential, secret, or source-tree location. |
| Hostile filesystem changes | Traversal, symlink escape, non-regular targets, and root replacement fail closed without outside-root access or secret-bearing diagnostics. |
| Resource ownership | Owned rooted capabilities close once in reverse acquisition order. `Options.ObjectStore` remains borrowed for unrelated supported product behavior. |

The same-digest idempotent-reuse recommendation in `temp/analysis-notes.md` is rejected for ASW-06 because the live contract creates unique immutable destinations and Core 01 does not require reuse. Deduplication would be a behavior change and requires separate authorization.

**ASR-REQ-204 — filesystem-only closure**

ASW-05 MUST close configuration admission for managed/object-store Reference Pack publication and ASW-06 MUST leave no Reference Pack storage mechanics in `internal/app/server`. No platform helper may acquire Reference Pack reference grammar, digest meaning, lifecycle, errors, or authorization. Validation MUST cover cancellation, exact/over-limit reads, digest validation, distinct immutable publication, traversal and malformed references, symlinks, root replacement, byte equality, idempotent removal, root-free diagnostics, and reverse-order capability closure on the filesystem adapter.

### 4.2B Incident Bundle storage remediation contract

**ASR-REQ-205 — Incident Bundle adapter ownership**

`internal/modules/incidentbundles` MUST retain semantic bundle references and the unchanged `BundleStorage` port. The only current-major realization MUST be a filesystem adapter in `internal/app/incidentportabilityassembly`, constructed by `incidentportabilityassembly.NewBundleRootStorage(temporaryRoot, exportRoot)`. Server MUST only construct, own, close, and inject it.

**ASR-REQ-206 — unsupported Incident Bundle object storage removal**

`sharedIncidentBundleStorage` and `sharedPublicationBytes` MUST be deleted. No runtime alias, dual-read, object-marker fallback, or compatibility mode may remain. An unsupported deployment MUST stop all application processes, restore verified filesystem-root contents, remove unsupported configuration, and only then upgrade.

**ASR-REQ-207 — Incident Bundle behavior preservation**

ASW-06B MUST preserve cancellation without partial files, unique immutable publication, bundle-ID and logical-reference validation, exact/over-limit reads, traversal and symlink rejection, root-replacement detection, idempotent removal, root-free diagnostics, containment, and reverse-order closure. Selector/accounting changes are permitted only if the tests actually move.

### 4.3 Runtime façade and test-seam remediation contract

**ASR-REQ-301 — mandatory two-stage migration**

ASW-04 MUST first decompose `runtime.go` within package `server` while preserving `Options`, `RuntimeSettings`, every exported `Runtime` field, `NewRuntime`, startup order, cleanup order, error mapping, and all callers. Only after ASW-04 passes may ASW-07 migrate callers and replace the broad structure with the lifecycle-only façade. The two stages MUST NOT be combined into one uncharacterized change.

**ASR-REQ-302 — final lifecycle interface**

The final exported runtime surface MUST be exactly the following constructor and methods; `Runtime` fields MUST be unexported:

```go
func NewRuntime(context.Context, configassembly.Deployment, Options) (*Runtime, error)
func (r *Runtime) HTTPHandler() http.Handler
func (r *Runtime) ActivatePublication() error
func (r *Runtime) FatalEvents() <-chan processlifecycle.FatalSignal
func (r *Runtime) PublishedComponentLost(reason string) bool
func (r *Runtime) ShutdownDrainTimeout() time.Duration
func (r *Runtime) PublicHTTPDiagnostics() httpapi.RouteDiagnostics
func (r *Runtime) Close()
```

The implementation MUST NOT add a broad `RuntimeInterface`, `RuntimeTestSupport`, generic service locator, or field-equivalent getter set. A consumer that requires an interface MUST own the smallest interface containing only the methods it calls.

**ASR-REQ-303 — caller and field migration**

| Current access | Required final access |
| --- | --- |
| `Runtime.Handler` | `HTTPHandler()` for production and recovery callers; `httptestx` uses it internally to create `HTTP`. |
| `Runtime.Settings.ShutdownDrainSeconds` | `ShutdownDrainTimeout()`. |
| `Runtime.PublicHTTP` | `PublicHTTPDiagnostics()`. |
| `Runtime.Postgres` | `appsupport.ServerHarness.DB` or `.Pool`, selected by the consumer's API. |
| `Runtime.ObjectStore` | Explicitly borrowed `appsupport.ServerHarness.ObjectStore`. |
| Jobs manager/runner/transactions | Jobs public/store evidence or an owner-named probe exposing only the required create, dequeue-gate, recovery, or activation capability. |
| Collaboration hub/dispatcher/intents | Collaboration-owned test support exposing only subscription, revocation, dispatcher shutdown, or intent append required by the test. |
| Timeline projection bundle | Owner façade or named rebuild capability for the exact projection family. |
| Revisions runtime | `revisionassembly` appender or another owner-specific façade. |
| Extension, staged-object, publication, and lease internals | Package-local tests or a typed owner-specific fault/probe; no cross-package runtime field. |
| `httptestx.Server.Runtime` | Remove; final exported fields are exactly `Config`, `HTTP`, and `Clock`; cleanup uses a private runtime. |
| Recovery browser tooling | Constructor plus lifecycle accessors only. |

`appsupport.ServerHarness` MUST expose `DB`, `Pool`, and the explicitly borrowed `ObjectStore`; additional capabilities MUST be owner-named and justified by a current caller. It MUST NOT expose `*server.Runtime` directly or indirectly.

**ASR-REQ-304 — typed incident-create commit fault**

ASW-07 MUST replace `DependencySetWithPostgresDBDecoratorForTesting`, key `app.server.postgres_db_decorator`, its `map[string]any` entry, and SQL-text observation with a typed incident-create commit-fault capability owned by test support. During migration, the old seam MAY remain only while all of these conditions hold: test-runtime admission guards it, exactly the known `appsupport` caller uses it, a static assertion enforces that caller set, production admission rejects it, and the same workstream removes it after caller migration. No generic override registry may replace it.

**ASR-REQ-305 — lifecycle and ownership invariants**

The final migration MUST preserve idempotent `Runtime.Close`, reverse-order closure of owned resources, non-closure of borrowed `Options.Postgres` and `Options.ObjectStore`, startup-error cleanup, publication activation ordering, fatal event delivery, and process exit mapping. These invariants MUST be tested independently of the removed exported fields.

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| `runtime.go` centralizes unrelated owner assembly and lifecycle stages. | Approximately 1,600 lines and imports most application/module/platform layers. | Critical | `must_fix` | Narrow `internal/app/server` facade with cohesive private assembly collaborators. | Characterize startup/cleanup order, then split without changing public surface or stage order. |
| Process-model behavior conflicts with adopted owner text. | `EXT-REQ-213`, Core 04 `REQ-04-145` and `AC-404`, typed projections, runtime branch, replicated integration test, and non-owner guide. | Critical | `must_fix` | Adopted owners first; application and platform configuration implementation afterward. | Execute ASW-01 before ASW-02 or ASW-05; the planning decision is closed but implementation remains blocked until adoption. |
| Bundle and Reference Data persistence adapters live in the app facade package. | Adapter files implement module ports with rootedfs/objectstore. | High | `should_fix` | Named module port, owner assembly adapter, platform mechanics, application binding. | Keep Incident Bundles separate; move Reference Data adapters under ASW-06 according to ASR-REQ-201..ASR-REQ-204. |
| Shared object-store helper has no supported current-major consumer. | Used only by the two unsupported publication storage adapters. | High | `must_fix` | Delete with those adapters; do not promote it into platform. | ASW-05 removes the helper and every activation surface. |
| Generic component leadership lives beside server assembly. | `component_leader.go` depends on processlease and one staged-object use. | High | `should_fix` | Ordinary application lifecycle worker after owner adoption. | ASW-05 removes distributed leadership; generic lease mechanics remain platform-owned only where still used. |
| Collaboration translations cross platform jobs, Network Flow, and Collaboration. | `collaboration_intents.go` maps two owner intent types. | High | `intentional/no_action` | Application composition adapter. | Retain transaction boundary; split file only as part of cohesive runtime decomposition. |
| WebSocket adapter is transport-adjacent but not domain logic. | `collaboration_socket.go` maps platform frames to a Collaboration port. | High | `intentional/no_action` | Application transport edge. | Preserve origin, frame, close, and authorization semantics. |
| Exported test hook leaks a test-only assumption through production package API. | One caller in `internal/testutil/appsupport`; string key and `map[string]any`; runtime guard. | High | `must_fix` | Typed incident-create fault support under `internal/testutil/appsupport` or the Incidents owner test support. | ASW-07 must replace the seam according to ASR-REQ-304 and prove production injection is impossible. |
| Serverprocess contains repeated fixture helpers. | Shared/local SQL, environment, bucket, request, and recovery setup across test files. | Medium | `should_fix` | `internal/testutil/appsupport` only for proven reusable application composition. | Consolidate after semantic reuse review; retain process-specific orchestration locally. |
| A Network Flow test depends on source file text. | Test reads `runtime.go` and `runtime_routes.go` and searches symbols. | High | `must_fix` | Backend module-boundary test owned with Network Flow collaborators. | Update assertion in the same authorized slice that moves the source; preserve the boundary, not the filename. |
| Owner modules are assembled but their behavior is not duplicated in server. | Runtime calls existing timeline, revision, evidence, import, entity, indicator, saved-view, reporting, and Network Flow facades/providers. | Medium | `intentional/no_action` | Existing modules. | Preserve dependency direction and prevent logic migration into application assembly. |
| Generated contracts are consumed indirectly by assembly and tests. | Generated extension coordinator, HTTP/OpenAPI catalog, view schemas, web asset registry. | Critical | `intentional/no_action` | Adopted owner inputs and generators. | Never hand-edit generated files; run drift checks after any authorized owner-input change. |
| No direct grid-vendor or frontend-controller implementation exists in the target. | No target imports from grid vendor; frontend is served as packaged assets. | Low | `intentional/no_action` | `packages/grid-adapter` and `apps/web`. | Freeze indirect contracts only. |

## 6. Refactor Workstreams

WF-00 through WF-08 preserve the completed discovery and planning history. They are not the authoritative execution sequence and their stable IDs MUST NOT be repurposed.

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Lock scope, authority, dirty-tree state, and planning-only posture. | Tracker only. | `make lint-markdown` after tracker edit. | Scope, commit, source list, and allowed write recorded. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Inventory every target file, caller, dependency, and evidence row. | Both target directories and inbound callers. | Repository searches; later `make backend-module-boundary-check`. | All 37 files have owner/risk/test posture. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-05 | Bind observable surfaces to adopted owners without transferring authority. | Owner docs, runtime/routes/publication/storage seams. | Existing contract parity tests; `make generate-drift` later. | Every contract risk has owner and evidence. |
| WF-03 | Characterization test gap analysis | parallel | WF-01, WF-02 | WF-05, WF-06 | Identify sufficient existing evidence and gaps before movement. | Server unit/integration/process tests and owner test families. | `make test-slice OWNER=app.server`; service-backed slice. | Each risky slice has pre-move evidence or a blocker. |
| WF-04 | Boundary/coupling scan | parallel | WF-01 | WF-05 | Separate facade, transport, persistence, owner logic, test support, and generated consumers. | Runtime, adapters, boundary policies, test support. | `make backend-module-boundary-check`. | Findings classified with candidate owners. |
| WF-05 | Facade or ownership redesign plan | chain | WF-02, WF-03, WF-04 | WF-06 | Retain public facade while defining private assembly and adapter seams. | `runtime.go`, server/publication/routes/settings/adapters. | Slice-specific unit/integration targets. | Design preserves public contracts; lease branch remains blocked. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07, WF-08 | Order smallest behavior-preserving moves and rollback points. | Files named per Section 7. | Validation attached to each slice. | Every slice has dependency, rollback, and completion criterion. |
| WF-07 | Harness/test/accounting update plan | parallel | WF-03, WF-06 | WF-08 | Update test support and owner selectors only when paths actually move. | Test files, `internal/testutil/appsupport`, authored test-family inputs if required. | Catalog/drift and selected owner rows. | No harness change justified by runtime architecture alone. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | none | Run narrow-to-broad checks and record residual risk. | Changed implementation/test/support files in later task; tracker handoff. | Section 8 ladder. | Commands/results/artifacts and next action recorded. |

### Authoritative remediation execution sequence

The user has authorized the complete sequence. Execution MUST be strictly linear:

```text
ASW-00 -> ASW-01 -> ASW-02 -> ASW-03 -> ASW-04 -> ASW-05 -> ASW-06 -> ASW-06B -> ASW-07 -> ASW-08
```

After each workstream, targeted validation MUST complete, Sections 9–12 and every applicable handoff table MUST record actual evidence, and `make lint-markdown` MUST pass before the successor begins. A failed workstream MUST be recorded and execution MUST stop without marking it `DONE`.

| ID | Workstream | Depends on | Affected areas | Remediation and rationale | Long-term benefit | Compatibility or migration impact | Risk if unresolved | Rollback boundary | Validation | Binary exit criterion and tracker checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| ASW-00 | Normalize this controlling tracker | none | Tracker only | Reconcile strict dependencies, filesystem-only storage, the separate Incident Bundle move, shared lease mechanics with typed identities, recovery-serving failures, acceptance rows, and traceability. | A decision-complete ledger that prevents slice overlap and rediscovery. | Preserves historical workflow/slice IDs and the user's existing tracker diff; adds `ASW-06B`. | Execution could follow stale dependencies or preserve unsupported adapters. | Revert only the ASW-00 additions before any owner file changes. | `make lint-markdown`; `git diff --check`. | This tracker records branch/revision/dirty posture, all decisions, actual validation, rollback posture, and ASW-01 as the sole next action. |
| ASW-01 | Adopt specification and guidance corrections | ASW-00 | Core 04, Extensions lifecycle/fatal registries, development guide | Preserve `EXT-REQ-213` and `REQ-04-145`; amend `AC-404`, `REQ-04-113`, `REQ-04-145`, `REQ-04-146`, and acceptance/fatal entries; define exact recovery-serving outcomes; remove current-major replicated instructions. | One authoritative current-major lifecycle with active-active reserved for a future major. | Unsupported replicated deployments require an offline cutover; no runtime or data migration occurs in this slice. | Owners continue to contradict each other and cannot safely govern downstream code. | Revert the coordinated specification transaction before any downstream work. | `make lint-markdown`; human authority and contradiction review. | No owner contradiction remains; the offline cutover posture and exact recovery-serving failures are adopted and checkpointed. |
| ASW-02 | Correct authored and generated projections | ASW-01 | Extensions authored contracts, catalogs, generated projections | Advance the shared protocol schema and emit only the closed single-active facts plus recovery-serving outcomes; defer Go field removal to ASW-05. | Generated consumers cannot widen or reactivate unsupported modes. | No persisted migration; unsupported projection alternatives disappear. | Generated consumers can preserve contradictory behavior. | Revert authored inputs and generated artifacts as one transaction. | `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; Extensions owner unit slice. | No authored/generated artifact emits active replicated serving and all generation evidence is checkpointed. |
| ASW-03 | Establish characterization baseline | ASW-02 | Tests and evidence | Add or tighten API-shape, lifecycle/order, borrowed-resource, storage, lease, recovery, and current fault-seam evidence; label replicated-overlap success as negative evidence, not a freeze. | Later behavioral and structural regressions are attributable. | Test-only additions may expose current owner mismatches; production remains unchanged. | Later movement could erase undocumented invariants. | Revert independent test additions that lack owner or preservation justification. | Unit/service-backed slices for `app.server`, `platform.config`, `module.reference_data`, `module.incidentbundles`, and `module.incidents`; boundary check; build-server. | Each later slice cites passing baseline evidence or an explicit owner-backed behavior change. |
| ASW-04 | Same-package runtime decomposition | ASW-03 | App-server implementation/tests and Network Flow boundary evidence | Extract private phases for preflight, resources, lifecycle gates, extension coordination, modules, and HTTP/publication while preserving the full API; replace the filename assertion with exact semantic route-catalog evidence. | A thin coordinator with cohesive, independently reviewable assembly phases. | No caller, API, route, error, startup, cleanup, or publication behavior change. | The monolith and brittle source test make later remediation unsafe. | Revert one extraction at a time while retaining the original coordinator until parity. | App-server unit/service-backed slices; Network Flow unit slice; route/OpenAPI parity; boundary check; build-server. | `runtime.go` is a thin coordinator and the pre-ASW-04 API snapshot is identical. |
| ASW-05 | Implement single-active lifecycle and remove unsupported branches | ASW-04 | Config, runtime, process/recovery leases, storage selection, tests | Remove `application.process_model`, replicated serving, publication-plan agreement, component leadership, object-store publication adapters/helper, and all reactivation surfaces; always acquire the app lease before Stage 1; use a distinct typed recovery-serving admission; run one lifecycle janitor. | One comprehensible fencing, readiness, publication, job-admission, and fatal model. | Legacy keys fail `unknown_key`; unsupported deployments must cut over offline; legacy markers are ignored. | Split-brain mutation, stale publication, multiple dequeuers, and ambiguous recovery overlap remain possible. | Roll back the whole lifecycle transaction to ASW-04; never partially restore replicated behavior. | Affected owner slices; exact lease/process cases; drift; build-server; test-fast; check. | Single-active and recovery acceptance pass, filesystem roots are enforced, and no replicated/shared-publication code remains. |
| ASW-06 | Move Reference Data filesystem adapter | ASW-05 | `referenceassembly`, server binding, Reference Data tests/accounting | Move the supported rooted adapter behind the unchanged six-method port; server only constructs, owns, and injects it. | Reference semantics, filesystem mechanics, and application binding have clear owners. | Package/test selectors may move; logical references and persisted shapes do not. | Persistence mechanics remain in the facade and semantics can leak into platform code. | Keep the old binding until identical tests pass; revert path/selector changes together. | Reference Data unit/service-backed slices; app-server slice; boundary check; build-server; accounting drift if needed. | `referenceassembly.NewRootStorage` passes all filesystem criteria and no Reference Pack storage mechanics remain in server. |
| ASW-06B | Move Incident Bundle filesystem adapter | ASW-06 | `incidentportabilityassembly`, server binding, Incident Bundle tests/accounting | Move the supported rooted adapter behind the unchanged owner port while leaving semantic references in `incidentbundles`. | Portability storage becomes cohesive without coupling the server facade to path mechanics. | Package/test selectors may move; references, limits, and persisted shapes do not. | Incident portability mechanics remain misplaced and difficult to evolve. | Keep the old binding until the new adapter passes the same suite; revert path/selector changes together. | Incident Bundles unit/service-backed slices; app-server service-backed slice; boundary check; build-server; accounting drift if needed. | `NewBundleRootStorage` passes all ASR-REQ-207 outcomes and no bundle storage mechanics remain in server. |
| ASW-07 | Narrow runtime and replace test seams | ASW-06B | Runtime API, harnesses, recovery tooling, module tests, Incidents owner | Migrate capability families, make runtime/settings/publication types private, enforce exact lifecycle accessors and harness fields, and replace the generic SQL-observing decorator with an owner-local typed incident-create committer seam. | Minimal stable façade, explicit test ownership, and production/test isolation. | Internal callers migrate; no external Go compatibility is owed; public HTTP/WS and resource ownership remain unchanged. | Runtime internals and production fault injection remain broadly coupled. | Migrate/validate one family at a time, but remove every old field/decorator before completion. | App-server and all affected owner slices; boundary check; build-server; test-fast; check. | Static searches and tests prove the exact final APIs and absence of broad getters, override keys, SQL observation, or production fault injection. |
| ASW-08 | Final accounting, validation, and handoff | ASW-07 | Authored accounting, generated topology, validation evidence, tracker | Audit actual moves/selectors, regenerate only through owner inputs, remove obsolete wrappers/imports, and execute the final ladder. | Durable evidence routing and a complete future-maintainer handoff. | Only genuinely moved selectors/paths change; no wire/data migration. | Passing behavior may be unaccounted or handoff claims unverifiable. | Revert each owner-input/generated transaction together and the last independently validated slice on failure. | Section 8 in order, including `make agent-finalize` and `make check`. | Every acceptance row has actual evidence or justified N/A; no stale selector, drift, compatibility path, or blocker remains. |

## 7. Proposed Refactor Slice Plan

S-00 through S-08 preserve the original planning baseline and stable identifiers. They require later implementation authorization but are historical after this revision. Where a row conflicts with ASR-REQ-* or ASW-01 through ASW-08, the authoritative remediation workstream governs; the historical row MUST NOT be edited into a new meaning.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 | none | Freeze the current exported facade, startup/cleanup ordering, route catalog, publication gate, and adapter storage outcomes before movement. | `internal/app/server` tests and existing owner tests. | Missing characterization could conceal drift. | Preserve all current unit/integration/process rows; add only owner-aligned gaps found by WF-03. | `make test-slice OWNER=app.server`; `make service-backed-test-slice OWNER=app.server` | Test-only change can be reverted independently. | Every later slice cites passing pre-move evidence or is blocked. |
| S-01 | S-00 | Split `runtime.go` into cohesive same-package private assembly steps while keeping `NewRuntime`, exported types/fields, ordering, errors, cleanup, and behavior unchanged. | `internal/app/server`; existing assembly packages remain dependencies. | Startup order, cleanup, fatal lifecycle, borrowed resource ownership, public test/tool fields. | Runtime, route, publication, server, and process tests. | `make test-slice OWNER=app.server`; relevant service-backed rows; `make build-server` | Revert one extracted collaborator at a time; retain original facade until parity passes. | `runtime.go` is a readable coordinator and no public contract or dependency direction changes. |
| S-02 | S-00, S-01 | Isolate incident-bundle storage port implementations from the facade while leaving construction/binding at the application edge. | `internal/app/server/incident_bundle_storage.go`, `shared_publication_storage.go`, `modules/incidentbundles`, existing platform storage packages. | Reference/path grammar, cancellation, limits, atomic publication/removal, root/symlink safety. | Move/preserve storage unit tests and process portability coverage. | `make test-slice OWNER=module.incidentbundles`; `make service-backed-test-slice OWNER=app.server`; `make backend-module-boundary-check` | Keep old adapter binding until new adapter passes identical tests. | Server depends on the owner port and a clearly owned adapter without observable storage drift. |
| S-03 | S-00, S-01, authority resolution | Reassess Reference Data storage placement without changing behavior. | `reference_pack_storage.go`, `shared_publication_storage.go`, `modules/reference_data`, platform storage. | Draft specification, reference grammar, digest/publication/read semantics. | Existing storage unit tests plus Reference Data owner rows. | `make test-slice OWNER=module.reference_data`; `make backend-module-boundary-check` | Leave current adapter in place if authority remains draft. | DEFERRED until an adopted owner or explicit implementation authority selects the boundary. |
| S-04 | S-00, S-01 | Isolate route composition and Collaboration transport/intent adapters as cohesive application-edge collaborators. | `runtime_routes.go`, `collaboration_socket.go`, `collaboration_intents.go`, runtime HTTP composition. | Routes/order, WS origin/auth/events, transaction publication timing, generated catalog parity. | Route/OpenAPI tests, collaboration/module integration, process socket and Network Flow route tests. | `make test-slice OWNER=app.server`; selected service-backed rows; `make backend-module-boundary-check` | Keep original registration/adapter path until exact parity passes. | Runtime assembly consumes narrow collaborators and all route/WS behavior is unchanged. |
| S-05 | S-00, S-01 | Replace the exported DB decorator hook with a narrow test injection seam and consolidate only proven reusable process helpers. | `test_dependencies.go`, `internal/testutil/appsupport`, `internal/testutil/httptestx`, selected serverprocess helpers. | Commit-fault coverage, harness/production separation, test selector paths. | Commit-fault integration and affected process tests. | `make test-slice OWNER=app.server`; `make service-backed-test-slice OWNER=app.server` | Migrate one caller at a time; retain guarded hook until no caller remains. | Production API no longer exposes a broad test helper and process-specific evidence stays local. |
| S-06 | S-01, `RB-001` resolved | Rework process-model publication agreement, application/component leases, or component leadership only according to reconciled adopted owners. | `runtime.go`, `publication_plan_agreement.go`, `component_leader.go`, processlease/config projections and tests. | Whole-process versus component loss, readiness, exit codes, concurrent serving, restore exclusion. | New owner-directed characterization plus all single/replicated/recovery tests. | `make service-backed-test-slice OWNER=app.server`; `make check` | Keep current semantics until owner correction and migration are separately authorized. | `requires later authorization`; completion criteria come from the reconciled owner amendment. |
| S-07 | Any slice that moves packages/tests | Update authored boundary/test-family inputs and regenerate only through Make-owned generators when selectors or paths change. | Authored boundary and test-catalog inputs; generated outputs only through generator. | Lost evidence accounting, hand-edited generated files, source-text boundary failure. | Catalog and boundary checks; replace Network Flow filename assertion semantically. | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make backend-module-boundary-check` | Revert owner-input edit and regenerated transaction together. | Generated outputs are clean and all active rows resolve to the moved tests. |
| S-08 | S-01 through applicable S-07 | Remove obsolete wrappers/imports after callers move and complete narrow-to-broad verification/handoff. | Only later authorized implementation files and tracker. | Orphaned paths, broad behavior drift, unreported skipped checks. | All affected owner rows and process evidence. | `make test-fast`, then `make check` when risk warrants it | Roll back the last independently validated slice, not the whole chain. | No obsolete path, out-of-scope diff, or unrecorded blocker remains. |

### Historical-slice disposition

| Historical slice | Authoritative disposition |
| --- | --- |
| S-00 | Absorbed by ASW-03 characterization baseline. |
| S-01 | Implemented only as ASW-04; its full exported-surface preservation remains mandatory. |
| S-02 | Absorbed by the separate ASW-06B workstream and MUST NOT be bundled into ASW-06. |
| S-03 | Superseded by the resolved ASR-REQ-201 boundary and ASW-06; it is no longer authority-deferred. |
| S-04 | May be performed as a private collaborator extraction within ASW-04 or caller migration within ASW-07, while its transport and route freeze remains binding. |
| S-05 | The generic narrow-seam wording is superseded by the exact typed capability in ASR-REQ-304 and ASW-07. |
| S-06 | Superseded by ASW-01, ASW-02, and ASW-05; implementation remains blocked until ASW-01 adoption. |
| S-07 | Absorbed by ASW-08 and applies only when actual paths or selectors move. |
| S-08 | Absorbed by ASW-08 final validation and handoff. |

## 8. Validation Plan

The commands below were discovered through the public Make task surface and owner guidance. A command is planned evidence until its exact result and run root are recorded. The 2026-08-04 Markdown result is historical evidence for the original tracker, not validation of this revision.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| documentation | `make lint-markdown` | Tracker and authored Markdown | after every workstream | This is the final gate before a successor begins. |
| unit | `make test-slice OWNER=app.server` | Non-service-backed `app.server` owner rows | yes | Current manifest includes unit and support-unit evidence. Run as the characterization baseline. |
| Reference Data unit | `make test-slice OWNER=module.reference_data` | Non-service-backed Reference Data owner rows | yes before ASW-06 | Includes logical-reference and storage conformance evidence. |
| integration | `make service-backed-test-slice OWNER=app.server` | Current service-backed integration, process, and browser rows for the owner | yes | Requires declared fixtures; record run root and classify infrastructure failures separately. |
| Reference Data integration | `make service-backed-test-slice OWNER=module.reference_data` | Service-backed Reference Data rows | yes before and after ASW-06 | Use the owner selector rather than package-name inference. |
| Incident Bundles unit | `make test-slice OWNER=module.incidentbundles` | Non-service-backed Incident Bundles rows | yes before and after ASW-06B | Preserve owner row IDs unless a real test move requires selector maintenance. |
| Incident Bundles integration | `make service-backed-test-slice OWNER=module.incidentbundles` | Service-backed Incident Bundles rows | yes before and after ASW-06B | Record declared fixture and run-root evidence. |
| e2e/browser | `make service-backed-test-slice OWNER=app.server ROWS=app.server.browser_support.supports_zero_membership_extension_discovery_and_2d096f2dcd` | Narrow browser-authenticated support scenario | yes for route/discovery/auth composition changes | `make browser-e2e-support` was discovered as an internal helper, so the public owner-slice command is preferred. Use public `make browser-e2e` only when broader browser risk warrants it. |
| generated drift | `make generate-drift && make generated-artifact-policy-check && make json-shape-check` | Generated contracts, policy, JSON shape | no for a same-package structural baseline; yes before completing any owner-input, selector, or generated-consumer move | Never hand-edit generated outputs. Execute as separate commands when recording exact results. |
| import-boundary/static | `make backend-module-boundary-check` | Backend import and source boundary policies | yes | The Network Flow source-text assertion is a known structural coupling. |
| build | `make build-server` | Production deployable composition | yes after ASW-04, ASW-05, and ASW-07 | Proves the binary composition root still consumes the intended facade. |
| finalization | `make agent-finalize` | Harness-maintenance artifacts and retained-run handling | yes before final broader verification | Supply `RESULTS_DIR` only for intentionally retained successful full warm-check evidence; otherwise record that it was unset. |
| full check | `make test-fast`, followed by `make check` when implementation risk warrants it | Broader repository verification | no before the first slice; yes before final handoff for high-risk runtime changes | Run `make agent-finalize` before broader end-of-run verification in a later implementation task. Supply `RESULTS_DIR` when retaining a successful run, otherwise explicitly report it unset. |

ASW-08 MUST use this order, skipping a row only when the handoff records why it is not applicable:

1. `make task-guide ROLE=module-author OWNER=<owner-id>` for every affected owner.
2. Every affected unit owner slice.
3. Every affected service-backed owner slice.
4. The exact app-server browser-support row when route, discovery, authorization, or packaged-browser composition changed.
5. `make backend-module-boundary-check`.
6. `make generate-drift`.
7. `make generated-artifact-policy-check`.
8. `make json-shape-check`.
9. `make build-server`.
10. `make test-fast`.
11. `make agent-finalize`, with `RESULTS_DIR` unset unless intentionally retaining a successful full warm-check run.
12. `make check`.
13. Update this tracker and run `make lint-markdown`.

For ASW-00, the applicable validation is `git diff --check` and `make lint-markdown`. Historical Markdown results do not validate this normalization, and no product validation may be inferred from it.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| AST-001 | Normalize and lock `app-server` scope and write boundary | WF-00 | DONE | none | Section 1 | Both target paths, output, allowed write, and non-goals are explicit. |
| AST-002 | Inventory all target files and inbound callers | WF-01 | DONE | AST-001 | Section 2; 37 file rows | Every file is inventoried with responsibility, coupling, tests, owner candidate, and risk. |
| AST-003 | Map adopted owners and support documents | WF-02 | DONE | AST-002 | Sections 1, 3, and 4 | Observable behavior is assigned without promoting code or tests to authority. |
| AST-004 | Record public contract freeze and evidence posture | WF-03 | DONE | AST-003 | Section 4 | Every discovered contract risk has owner, tests, and characterization posture. |
| AST-005 | Classify facade and coupling findings | WF-04 | DONE | AST-002, AST-003 | Sections 3 and 5 | Findings use only the required classifications and name planning actions. |
| AST-006 | Adopt the selected single-active owner correction | ASW-01 | DONE | AST-003, AST-015, AST-025 | Core 04 `AC-404`; Extensions lifecycle/fatal registry; ASW-01 handoff | Core 04 is adopted without contradiction with `EXT-REQ-213` or `REQ-04-145`. |
| AST-007 | Define behavior-preserving facade decomposition | WF-05 | TODO | AST-004, AST-005 | S-01 and S-04 | Later implementation design fixes private seams without changing the public facade. |
| AST-008 | Characterize storage adapters before ownership movement | WF-03 | TODO | AST-004 | S-02 and S-03 | Exact storage outcomes pass under candidate adapter boundaries. |
| AST-009 | Resolve Reference Data storage owner | WF-05 | DONE | AST-003 | `RB-002`; ASR-REQ-201 | Tracker selects the behavior-preserving `referenceassembly` boundary without promoting the draft NLSpec. |
| AST-010 | Replace exported test-only dependency seam | WF-05 | TODO | AST-004, AST-005 | S-05 | Commit-fault coverage passes with a narrower seam and no caller remains. |
| AST-011 | Replace filename-sensitive Network Flow boundary assertion when files move | WF-07 | TODO | AST-007 | S-07; Network Flow contract test | Semantic boundary remains enforced without depending on obsolete file placement. |
| AST-012 | Update harness/test accounting only for actual path or selector movement | WF-07 | TODO | AST-007 through AST-011 as applicable | Authored test-family inputs and drift results | All active rows resolve and no phase identity becomes architecture. |
| AST-013 | Execute narrow-to-broad validation in authorized implementation task | WF-08 | TODO | Applicable implementation slices | Section 8 run artifacts | Exact commands, results, failures, and skipped checks are recorded. |
| AST-014 | Maintain tracker and final handoff | WF-08 | DONE | AST-001 through AST-005 | Sections 9–12; Markdown lint run root | Markdown validation result and current session handoff are recorded. |
| AST-015 | Convert analysis into the NLSpec-style remediation contract | Tracker revision | DONE | AST-001 through AST-005 | ASR-REQ-001..ASR-REQ-305; Sections 4–12 | The original three planning decisions, interfaces, defaults, workstreams, and binary criteria are decision-complete. |
| AST-025 | Normalize the authorized execution ledger | ASW-00 | DONE | AST-015 | Sections 4, 6, 9–12; `.cartulary/test-results/20260805T230501Z-p1887906` | Tracker validation passed and ASW-01 is the sole next action. |
| AST-016 | Correct adopted owner wording | ASW-01 | DONE | AST-025 | `.cartulary/test-results/20260805T230947Z-p1892758`; ASR-AC-112 | Owner amendment is adopted and the contradiction is closed. |
| AST-017 | Close typed projections and unsupported configuration | ASW-02 | TODO | AST-016, AST-026 | Prior failed run `.cartulary/test-results/20260805T231306Z-p1897061`; resolved `RB-005` | Restart the authored/generated transaction from the ASW-01 checkpoint; no ASW-02 completion is yet claimed. |
| AST-026 | Restore exact Go toolchain readiness | Harness prerequisite remediation | DONE | AST-017 failure evidence | Effective `go1.26.5`; final generation root `.cartulary/test-results/20260806T023137Z-p1977974`; full-check root `.cartulary/test-results/20260806T023310Z-p1990254`; `RB-005` | The same-cache toolchain is healthy, generation passes beyond `codegen-toolchain`, and the corrupt entries remain quarantined for operator-controlled cleanup. |
| AST-018 | Establish the remediation characterization baseline | ASW-03 | TODO | AST-017 | ASR-AC-101..ASR-AC-310 as applicable | Required pre-change evidence passes and is recorded with run roots. |
| AST-019 | Decompose runtime within the existing package | ASW-04 | TODO | AST-018 | ASR-REQ-301; ASR-AC-301 | The coordinator is thin and the complete exported surface remains unchanged. |
| AST-020 | Implement single-active process remediation | ASW-05 | TODO | AST-019 | ASR-REQ-101..ASR-REQ-107 | All process acceptance rows pass and no replicated/shared-publication path remains. |
| AST-021 | Move Reference Data filesystem adapter to `referenceassembly` | ASW-06 | TODO | AST-020 | ASR-REQ-201..ASR-REQ-204 | The filesystem adapter passes the unchanged port and no Reference Pack mechanics remain in server. |
| AST-021B | Move Incident Bundle filesystem adapter to `incidentportabilityassembly` | ASW-06B | TODO | AST-021 | ASR-REQ-205..ASR-REQ-207 | The filesystem adapter passes its unchanged port and no bundle mechanics remain in server. |
| AST-022 | Narrow runtime and replace test seams | ASW-07 | TODO | AST-021B | ASR-REQ-301..ASR-REQ-305 | Lifecycle-only API and named test capabilities replace broad field and override access. |
| AST-023 | Update actual boundary and harness accounting | ASW-08 | TODO | AST-022 | Authored inputs and generated drift evidence | Only genuinely moved selectors/paths change and all rows resolve. |
| AST-024 | Complete final validation and handoff | ASW-08 | TODO | AST-023 | Section 8 ladder and Section 12 traceability | Every applicable acceptance row has actual evidence and no unrecorded blocker remains. |

## 10. Session Handoff Log

Time for this planning session: `2026-08-04T19:41:55-04:00`. Branch/commit inspected: `main` at `7c3168c3`. The pre-existing untracked `docs/handoffs/web-e2e-module-refactor-tracker.md` is unrelated and was not touched.

NLSpec remediation revision session: `2026-08-05T18:39:24-04:00`. Branch/commit inspected: `main` at `43e74d1f`; the worktree was clean before this tracker-only edit. The only touched repository file is this tracker.

Authorized execution session began at `2026-08-05T19:05:19-04:00`. Branch/commit: `main` at `43e74d1f`. The tracker was already modified when execution began and no other file was dirty; ASW-00 preserved and extended that user-owned diff.

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `2026-08-04T19:41:55-04:00` | Codex app-server planning | Scope and authority mapped; tracker created. | Inspected framework, domain, Core 00–05, adopted subsystem owners, draft Reference Pack NLSpec, guide; touched only this tracker. | Targeted `sed`/`rg`; `git status --short --branch`; `git rev-parse --short HEAD`. | Target exists; safe label is `app-server`; Core 05 is inapplicable. | `RB-001`; `RB-002`. | Obtain owner resolution before any lease/process-model implementation. |
| `2026-08-05T18:39:24-04:00` | Codex NLSpec remediation revision | Authority classes and ASR requirement families added; analysis recommendations reconciled with adopted owners and live code. | Inspected `nlspec-spec.md`, `analysis-notes.md`, tracker, Core 00/01/04, Extensions NLSpec, guide, configuration/projection evidence; touched only this tracker. | Targeted `sed`/`rg`; `git status --short --branch`; `git rev-parse --short HEAD`; `make lint-markdown`. | Tracker is decision-complete; Markdown lint passed at `.cartulary/test-results/20260805T224327Z-p1876699`. | ASW-01 owner adoption remains blocked. | Obtain authorization and owner review for ASW-01 only; do not begin ASW-02 or ASW-05. |
| `2026-08-05T19:05:19-04:00` | Codex ASW-00 | Authorized ledger normalized without touching an owner, implementation, test, generated, or domain file. | Extended this pre-existing modified tracker only; recorded `main` at `43e74d1f` and the dirty-tree posture. | `git status --short --branch`; `git rev-parse --short HEAD`; `git diff --check`; `make lint-markdown`. | Strict order, exact ownership decisions, rollback gates, and acceptance/traceability rows reconciled; Markdown lint passed at `.cartulary/test-results/20260805T230501Z-p1887906`; diff check passed. | `RB-001` blocks ASW-02+, not the ASW-01 owner correction. | Execute ASW-01 only; if its coordinated owner transaction fails, record and stop. |
| `2026-08-05T19:10:08-04:00` | Codex ASW-01 | Core 04 and Extensions now agree on one active process, typed lease identities, and exact recovery-serving outcomes. | Changed Core 04, Extensions NLSpec, development guide, and this tracker; reviewed `docs/domain.md` navigation posture and left it unchanged. | Targeted owner/contradiction `rg`; `git diff --check`; `make lint-markdown`. | ASR-AC-112 and owner review pass; Markdown lint passed at `.cartulary/test-results/20260805T230947Z-p1892758`. | None; `RB-001` is closed. | Execute authored/generated projection correction in ASW-02 only. |
| `2026-08-05T19:14:04-04:00` | Codex ASW-02 | FAILED and stopped before generation; the partial authored projection edit was rolled back to the ASW-01 checkpoint. | Temporarily edited four authored Extensions JSON inputs; restored all four with `apply_patch`; no generated artifact remains changed. | `make generate`; failure-artifact inspection; `git diff --exit-code -- contracts/extensions/...`; `git diff --check`. | `make generate` failed in `codegen-toolchain` while bootstrapping SQLC; run `.cartulary/test-results/20260805T231306Z-p1897061`. | `RB-005`. | Do not start ASW-03. Repair or clear the pinned Go-tool bootstrap state, then rerun ASW-02 from the ASW-01 checkpoint. |
| `2026-08-05T22:41:05-04:00` | Codex Go toolchain remediation | The corrupt automatic-toolchain cache is repaired and the shared readiness boundary is hardened; the failed ASW-02 content was not reapplied. | Updated Make/harness readiness, Testing Harness ownership/specification, regression tests, generated task-surface projections, bootstrap guidance, and this tracker; left `go.mod`, `go.sum`, and `docs/domain.md` unchanged. | Same-cache `go version`; `make generate`; drift/policy/shape/pin/lint/harness/doctor/finalize gates; `make check`. | Effective Go is exactly `go1.26.5`; generation passed at `.cartulary/test-results/20260806T023137Z-p1977974`; full check passed 715/715 units at `.cartulary/test-results/20260806T023310Z-p1990254`. | None for restarting ASW-02. | Restart ASW-02 only from the ASW-01 checkpoint; do not begin ASW-03 until the authored/generated transaction passes. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `2026-08-04T19:41:55-04:00` | Codex app-server planning | Facade diagnosed as legitimate but mixed-responsibility; all 37 files inventoried. | Inspected every file in both target directories plus direct callers; touched only this tracker. | `find internal/app/server internal/app/serverprocess -maxdepth 1 -type f`; targeted symbol/import searches. | `server.go` is the thin binary facade; `runtime.go` is the primary decomposition seam; `serverprocess` is test-only evidence. | Public `Runtime` fields and test/tool callers constrain early movement. | Start later implementation with S-00, then same-package S-01. |
| `2026-08-05T18:39:24-04:00` | Codex NLSpec remediation revision | Planning boundaries closed for process semantics, Reference Data adapters, runtime API, callers, and test fault injection. | Reinspected runtime/config/storage surfaces, `httptestx`, `appsupport`, recovery tool, and direct field callers; touched only this tracker. | Targeted `sed`/`rg`; 27/10 file count check. | Canonical execution order is ASW-01..ASW-08; ASW-04 must preserve the complete surface before ASW-07 narrows it. | ASW-01 owner adoption. | After adoption, characterize through ASW-03 before any structural or behavioral change. |
| `2026-08-05T19:05:19-04:00` | Codex ASW-00 | Backend ownership decisions are execution-closed but no package changed. | Tracker inventory/findings only. | Tracker review; `git diff --check`; `make lint-markdown`. | Reference and Incident Bundle filesystem moves are separate; unsupported adapters/helper are deleted in ASW-05; Network Flow semantic evidence moves to ASW-04. | None for ASW-01. | Do not touch backend source until ASW-03; follow the strict gate. |
| `2026-08-05T19:10:08-04:00` | Codex ASW-01 | No backend source changed; the adopted contract now governs later implementation. | Specification/guide/tracker only. | Human authority review; `git diff --check`. | Multiple provisioned instances are distinct from exactly one active process; process and recovery leases may share only owner-neutral mechanics. | None. | Keep implementation untouched until ASW-03 after ASW-02 generation. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `2026-08-04T19:41:55-04:00` | Codex app-server planning | No frontend shell/controller or grid-vendor implementation is in scope. | Inspected target imports and embedded-frontend process evidence; touched only this tracker. | Targeted `rg` for web assets, view schemas, grid/vendor imports, and browser rows. | Only packaged assets, client support, discovery, and workspace availability are indirect contracts. | None for tracker planning. | Preserve indirect contracts; do not create a frontend refactor slice. |
| `2026-08-05T18:39:24-04:00` | Codex NLSpec remediation revision | Frontend and grid scope remains unchanged; indirect contracts are traced through ASW-03 and ASW-08 only. | Reused verified target inventory and browser-support owner row; touched only this tracker. | Targeted `rg`; `make explain-test-owner OWNER=app.server`. | No frontend implementation workstream is justified. | None. | Run the exact browser-support row only when affected by a later route/discovery/auth composition change. |
| `2026-08-05T19:05:19-04:00` | Codex ASW-00 | Frontend remains non-applicable; public HTTP/WS and packaged-browser behavior stay frozen. | Tracker only. | `make lint-markdown`. | No frontend file, contract, or selector is admitted by ASW-00. | None. | Reassess the browser-support row only if a later composition change triggers it. |
| `2026-08-05T19:10:08-04:00` | Codex ASW-01 | Frontend remains non-applicable. | No frontend file changed. | Scope review. | HTTP/WS behavior remains frozen; only process admission/lifecycle ownership changed. | None. | No frontend action in ASW-02. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `2026-08-04T19:41:55-04:00` | Codex app-server planning | Generated consumers and OpenAPI/route/extension surfaces mapped. | Inspected OpenAPI parity test, generated-consumer call sites, generated-artifact policy, boundary projections; touched only this tracker. | Targeted `rg`; `make explain-target TARGET=generate-drift DETAIL=summary`. | Generated artifacts are downstream and need no tracker-task change. | `RB-001` affects lease projections; Reference Pack authority remains draft. | Use owner input then Make generation only in a later authorized behavior change. |
| `2026-08-05T18:39:24-04:00` | Codex NLSpec remediation revision | Exact single-active projection and unsupported-configuration disposition defined; no projection changed. | Inspected authored extension shared protocols/validation surfaces, generated consumer, Core 04 closed config registry, config implementation/tests; touched only this tracker. | Targeted `sed`/`rg`; `make help-all`. | ASW-02 must remove active replicated facts through authored inputs; `application.process_model` must become `unknown_key`. | ASW-01 adoption precedes projection work. | Update authored owners first, then authored projections and Make-generated outputs; never hand-edit generated files. |
| `2026-08-05T19:05:19-04:00` | Codex ASW-00 | Projection work is strictly deferred until ASW-02. | Tracker only; no authored or generated contract changed. | `git diff --check`; `make lint-markdown`. | Schema-version advance, exact single-active facts, recovery-serving outcomes, and Make-only generation are checkpointed. | ASW-01. | Adopt owners first; then edit authored inputs and regenerate as one ASW-02 transaction. |
| `2026-08-05T19:10:08-04:00` | Codex ASW-01 | Owner correction adopted; projections intentionally unchanged. | Core 04 and Extensions owner Markdown; no `contracts/**` or generated file changed. | Owner token/contradiction searches; `make lint-markdown`. | Downstream authority now requires single-active and the two recovery-serving outcomes. | None. | Advance authored shared-protocol inputs and generate in ASW-02. |
| `2026-08-05T19:14:04-04:00` | Codex ASW-02 | Projection transaction failed and was rolled back. | Restored `contracts/extensions/index.json`, `specification/contract-definitions.json`, `specification/shared-protocols.json`, and `validation/surfaces.json`; no generated file changed. | `make generate`; inspected `codegen-toolchain/tool-run-summary.json`; exact four-file `git diff --exit-code`. | Harness failure: Go toolchain cache lacked `/tmp/cartulary-go-mod/golang.org/toolchain@v0.0.1-go1.26.5.linux-amd64/src/_go.mod` during SQLC bootstrap. | `RB-005`. | Repair tool bootstrap outside this failed slice, then reapply authored v4 and generate as one transaction. |
| `2026-08-05T22:41:05-04:00` | Codex Go toolchain remediation | Code generation is unblocked; no Extensions projection input was changed by the repair. | Added exact Go readiness and atomic Go-tool bootstrap support, updated authored harness ownership, and regenerated task-surface/topology projections through Make. | Same-cache `go version`; `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make toolchain-drift`. | All commands passed; final generation root `.cartulary/test-results/20260806T023137Z-p1977974`, drift root `.cartulary/test-results/20260806T023147Z-p1980229`, and JSON-shape root `.cartulary/test-results/20260806T023202Z-p1983384`. | None. | Reapply the ASW-02 authored projection transaction and regenerate; preserve the strict checkpoint. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `2026-08-04T19:41:55-04:00` | Codex app-server planning | Owner rows and public validation ladder discovered; no product test suite run. | Inspected `tools/test_families/app.server.json`, topology/task/batch data, all target tests; touched only this tracker. | `make task-guide ROLE=module-author OWNER=app.server`; `make explain-test-owner OWNER=app.server`; `make help-all`; selected `make explain-target ... DETAIL=summary`; `make lint-markdown`. | 32 rows: 1 browser, 9 integration, 14 process, 1 support unit, 7 unit; 24 are service-backed. Markdown lint passed at `.cartulary/test-results/20260804T234731Z-p205315`. | Browser support target itself is an internal helper; public owner-slice command selected. | Run owner slices only during a later authorized implementation task. |
| `2026-08-05T18:39:24-04:00` | Codex NLSpec remediation revision | Binary acceptance and requirement-to-evidence traceability added; no product test target run. | Inspected app-server and Reference Data owner manifests and exact browser/process rows; touched only this tracker. | `make task-guide ROLE=module-author OWNER=app.server`; `make task-guide ROLE=module-author OWNER=module.reference_data`; both `make explain-test-owner` commands; `make lint-markdown`. | Tracker-only Markdown lint passed at `.cartulary/test-results/20260805T224327Z-p1876699`; no unit, integration, browser, build, drift, or full-check success is claimed. | ASW-01 blocks implementation validation. | Execute ASW-03 baseline after owner adoption and record every run root before ASW-04. |
| `2026-08-05T19:05:19-04:00` | Codex ASW-00 | Validation routing now includes all five ASW-03 baseline owners and the distinct ASW-06B owner. | Tracker only. | `git diff --check`; `make lint-markdown`. | ASW-00 documentation validation passed at `.cartulary/test-results/20260805T230501Z-p1887906`; no product-test success is claimed. | Product validation waits for its workstream. | Run only ASW-01 Markdown/human review next. |
| `2026-08-05T19:10:08-04:00` | Codex ASW-01 | Specification validation complete; no product test was applicable. | Three authored Markdown files plus tracker. | `git diff --check`; `make lint-markdown`. | Passed at `.cartulary/test-results/20260805T230947Z-p1892758`; no unit, integration, drift, build, or runtime result is claimed. | None. | Run ASW-02 generation/drift and Extensions owner unit evidence next. |
| `2026-08-05T19:14:04-04:00` | Codex ASW-02 | Required generation and validation ladder did not run past the first generation prerequisite. | Failure evidence only; no test or generated output retained. | `make generate`; read-only failure artifact inspection. | `codegen-toolchain` failed before Extensions owner unit, drift, artifact-policy, or JSON-shape validation; none is claimed. | `RB-005`. | Stop execution after checkpoint Markdown lint. |
| `2026-08-05T22:41:05-04:00` | Codex Go toolchain remediation | Harness readiness and recovery regression coverage are complete; product behavior remains unchanged. | Added scratch-only fake-launcher/cache and atomic-install tests; registered them in extended/full harness tiers. | `make lint-shell`; `make lint-markdown`; `make harness-contract`; `make doctor`; `make agent-finalize`; `make check`. | All passed; doctor identified launcher `go1.26.4` and effective cached `go1.26.5`; full check passed 715/715 units at `.cartulary/test-results/20260806T023310Z-p1990254`. Retained-run maintenance was skipped because `RESULTS_DIR` was unset. | None for ASW-02. | Run ASW-02's own Extensions evidence after its authored/generated transaction is reapplied. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `2026-08-04T19:41:55-04:00` | Codex app-server planning | Authorization remains with route/source owners; server composes dependencies and transport. | Inspected Core 04, route/profile assembly, membership/auth/socket/process tests; touched only this tracker. | Targeted `sed`/`rg` for authorization, leases, harness routes, and WebSocket behavior. | No app-local domain authorization policy or public recovery route found. | `RB-001` blocks lifecycle/lease corrections. | Preserve current authorization precedence and production/harness isolation in every slice. |
| `2026-08-05T18:39:24-04:00` | Codex NLSpec remediation revision | Lease identity, fatal outcomes, recovery exclusion, and production test-fault isolation are explicit requirements. | Inspected Extensions lease owner, Core 04 lease/config/security clauses, production/harness profiles, and commit-fault seam; touched only this tracker. | Targeted `sed`/`rg`; `make lint-markdown`. | Single-active decision is closed in planning; production fault injection and lease/recovery separation have binary criteria. | Owner contradiction remains until ASW-01. | Preserve fail-closed gates and authorize owner correction before runtime changes. |
| `2026-08-05T19:05:19-04:00` | Codex ASW-00 | Typed lease identities and exact recovery-serving failure vocabulary are fixed for adoption. | Tracker only. | Authority review; `make lint-markdown`. | `recovery_serving_lease_active` maps to closed startup/exit `2`; `recovery_serving_lease_lost` maps to immediate closure/fatal drain/exit `70`; Recovery target-lease mapping stays distinct. | ASW-01 owner adoption. | Amend the coordinated owner registries; do not partially adopt them. |
| `2026-08-05T19:10:08-04:00` | Codex ASW-01 | Recovery-serving startup/loss and application-process loss are distinct adopted outcomes. | Core 04 `REQ-04-113`, `REQ-04-145`, `REQ-04-146`, `AC-404`, `AC-534`; Extensions lifecycle/fatal registries and criteria. | Targeted `rg`; human contradiction review; `make lint-markdown`. | Startup denial is `recovery_serving_lease_active`/`2`; application-side loss is `recovery_serving_lease_lost`/fatal `70`; Recovery target loss retains its owner mapping. | None. | Project exact outcomes without widening them in ASW-02. |
| `2026-08-05T19:14:04-04:00` | Codex ASW-02 | No projection/security claim is retained from the failed slice. | Authored projection edits were rolled back; adopted ASW-01 owner documents remain. | Exact authored-file diff check. | Security/lifecycle authority is unchanged from ASW-01; machine projection remains knowingly stale until ASW-02 can complete. | `RB-005`. | Do not implement against the stale projection. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `2026-08-04T19:41:55-04:00` | Codex app-server planning | Tracker is implementation-ready except for explicitly blocked/deferred owner decisions. | Touched only this tracker; unrelated dirty file preserved. | `git status --short`; `make lint-markdown`. | Markdown lint passed at `.cartulary/test-results/20260804T234731Z-p205315`. Safe next implementation order is S-00 then S-01; S-06 is prohibited pending authority. | `RB-001`, `RB-002`, `RB-003`. | Authorize a behavior-preserving characterization/decomposition task or resolve the owner blockers. |
| `2026-08-05T18:39:24-04:00` | Codex NLSpec remediation revision | RB-002 and RB-003 are closed; RB-001 has a selected decision but remains implementation-blocked. | Touched only this tracker. | `git status --short --branch`; `git diff --check`; `make lint-markdown`. | Authoritative next step is ASW-01. No implementation row is complete and no product validation is claimed. | `RB-001` / ASW-01 owner adoption. | Adopt ASW-01, update the tracker checkpoint, then run ASW-02 and ASW-03 as separate workstreams. |
| `2026-08-05T19:05:19-04:00` | Codex ASW-00 | Workstream checkpoint complete; rollback is the ASW-00-only tracker delta before any owner edit. | Changed only this tracker and preserved its pre-existing user-owned diff. | `git status --short --branch`; `git rev-parse --short HEAD`; `git diff --check`; `make lint-markdown`. | ASW-00 passed; run root `.cartulary/test-results/20260805T230501Z-p1887906`. No failure or migration exists; unsupported-mode offline cutover is documented for later adoption. | `RB-001` is intentionally resolved by ASW-01. | Begin ASW-01 and no other workstream. |
| `2026-08-05T19:10:08-04:00` | Codex ASW-01 | Coordinated owner transaction passed and is the rollback unit. | Changed `docs/spec/04_security_deployment_and_conformance.md`, `docs/extension-subsystem-nlspec.md`, `docs/guides/cartulary-dev-guide.md`, and this tracker. | `git diff --check`; contradiction search; `make lint-markdown`. | No database/wire migration; unsupported replicated deployments must stop, remove the key, and restore verified filesystem publication roots. Run root `.cartulary/test-results/20260805T230947Z-p1892758`. | None. | Begin ASW-02 only; revert all ASW-01 owner/guide edits together if rollback is required. |
| `2026-08-05T19:14:04-04:00` | Codex ASW-02 | FAILED; strict execution stopped and rollback completed. | No ASW-02 authored/generated diff remains; ASW-01 files and tracker remain changed. | `make generate`; failure artifact inspection; authored-input rollback verification; `git diff --check`; `make lint-markdown`. | Failure is classified `harness/unknown_failure`, not a product assertion; run `.cartulary/test-results/20260805T231306Z-p1897061`; failed-slice checkpoint lint passed at `.cartulary/test-results/20260805T231510Z-p1898262`. | `RB-005`. | Resolve the missing pinned Go-toolchain cache/bootstrap prerequisite before resuming ASW-02; ASW-03 is prohibited. |
| `2026-08-05T22:41:05-04:00` | Codex Go toolchain remediation | `RB-005` is closed; the ASW-01 checkpoint and ASW-02 rollback posture are unchanged. | Harness/toolchain implementation, tests, projections, guide, and tracker only; no ASW-02 Extensions input was reapplied. | Exact-version and final verification ladder through `make check`. | The live cache contains a complete verified Go 1.26.5 toolchain; the prior corrupt extraction and `.ziphash` remain recoverable at `/tmp/cartulary-go-toolchain-quarantine.xn9mKA`. | None for restarting ASW-02. | Restart ASW-02 only; keep ASW-03 prohibited until ASW-02 passes and is recorded. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | **Resolved by ASW-01.** Core 04 `AC-404`, `REQ-04-113`, `REQ-04-145`, and `REQ-04-146` now agree with `EXT-REQ-213`; the Extensions lifecycle/fatal registries contain the distinct recovery-serving outcomes and the guide contains no current-major replicated-serving instructions. | Downstream projections and code now have one unambiguous owner contract. | ASW-01 owner review and `.cartulary/test-results/20260805T230947Z-p1892758`. | DONE |
| RB-002 | **Resolved execution decision.** Keep the semantic port in `reference_data`, delete the unsupported object-store adapter, move only the rooted-filesystem adapter to `internal/app/referenceassembly`, and bind from server. | This closes package ownership without preserving unsupported compatibility or promoting the draft NLSpec. | ASR-REQ-201..ASR-REQ-204 and ASW-05/ASW-06 evidence. | DONE |
| RB-004 | **Resolved execution decision.** Keep the semantic port in `incidentbundles`, delete the unsupported object-store adapter/helper, and move only the rooted-filesystem adapter to `internal/app/incidentportabilityassembly` in ASW-06B. | This prevents the Reference move from absorbing a distinct owner and keeps the facade free of path mechanics. | ASR-REQ-205..ASR-REQ-207 and ASW-05/ASW-06B evidence. | DONE |
| RB-003 | **Resolved planning decision.** Use the mandatory two-stage migration: ASW-04 preserves the complete surface; ASW-07 migrates callers and installs the exact lifecycle-only API plus typed test capabilities. | It separates safe decomposition from API and test-architecture change. | ASR-REQ-301..ASR-REQ-305 and ASW-03 characterization evidence. | DONE |
| RB-005 | **Resolved.** The incomplete Go 1.26.5 automatic-toolchain extraction and matching `.ziphash` were quarantined, Go's verified downloader restored the exact effective toolchain, and shared readiness now fails early on future corruption. SQLC was not the cause. | Generated work is no longer blocked by toolchain selection; ASW-02 still must restart its authored/generated transaction from the ASW-01 checkpoint. | Same-cache `go version` reports `go1.26.5`; `make generate` passed at `.cartulary/test-results/20260806T023137Z-p1977974`; `make check` passed 715/715 units at `.cartulary/test-results/20260806T023310Z-p1990254`; quarantine `/tmp/cartulary-go-toolchain-quarantine.xn9mKA` remains recoverable. | DONE |

## 12. Binary Completion Criteria

### 12.1 Tracker decision-completeness criteria

| Criterion | Status | Evidence |
| --- | --- | --- |
| Every file in `internal/app/server` and `internal/app/serverprocess` is inventoried or explicitly out of scope. | PASS | Section 2 contains all 37 files; none is excluded. |
| Every discovered public contract risk has an owner and test posture. | PASS | Section 4 records owners, evidence, existing tests, characterization posture, and risk. |
| Every proposed workflow has dependencies and exit criteria. | PASS | Section 6 preserves WF-00 through WF-08 and defines the strict ASW-00 through ASW-08 sequence, including ASW-06B. |
| Every proposed implementation slice is behavior-preserving unless explicitly marked `requires later authorization`. | PASS | Section 7 defaults to preservation; S-06 explicitly requires later authorization and authority resolution. |
| Validation commands are discovered or marked `TODO` with a reason. | PASS | Section 8 contains Make-owned commands and applicability. |
| Contradictions are marked `BLOCKED: owner contradiction`. | PASS | `RB-001` records the selected planning decision and the still-required adopted-owner correction without treating the selection as authority. |
| Repository/framework mismatches are recorded as planning findings. | PASS | The framework's generic module shape is adapted to the live application facade; replicated code/projection versus owner text is recorded; `serverprocess` is retained test evidence. |
| Handoff sections are current enough for another agent to continue without rediscovery. | PASS | Section 10 records both sessions, authority, exact decisions, files, commands, findings, blockers, and next workstream. |

### 12.2 Binary remediation acceptance criteria

#### Governance and execution

| Acceptance ID | Scenario | Binary pass condition | Planned evidence route |
| --- | --- | --- | --- |
| ASR-AC-001 | Authority and statement audit | Every normative tracker statement is classified as owner-backed behavior, remediation requirement, or future-only; analysis/research evidence is not represented as adopted behavior; owner conflicts remain blocked. | Human review against Section 1 and the authority sources. |
| ASR-AC-002 | Workstream transition | Before ASW-N+1 starts, ASW-N has actual validation results, status, changed files, blockers, rollback posture, and next action recorded in Sections 9–12 and applicable handoff tables. | Tracker diff at every workstream boundary. |
| ASR-AC-003 | Generated and harness boundary | No generated file is hand-edited; every generated change traces to an authored input and passing drift check; every harness/accounting change traces to an actual moved path or selector and carries no architecture claim. | Generated-artifact policy, drift checks, boundary review, and tracker traceability. |

#### Process, lease, projection, and configuration

| Acceptance ID | Scenario | Binary pass condition | Planned evidence route |
| --- | --- | --- | --- |
| ASR-AC-101 | First process starts | It acquires `application_process_lease` before Stage 1 and only then may migrate, resolve claims, prepare publication, bind listeners, become ready, advertise workspaces, activate workers, and dequeue jobs. | App-server service-backed process/integration row. |
| ASR-AC-102 | Second process overlaps | It performs none of the ASR-AC-101 effects, reports `extension_application_process_active`, and exits `2` after the configured acquisition timeout. | Replace the current replicated-overlap success row with a negative app-server integration row. |
| ASR-AC-103 | Holder exits or its lease session ends | The lease releases and a new process may subsequently acquire it; no two processes overlap in `held`. | Process-lease integration row. |
| ASR-AC-104 | Ownership proof becomes uncertain | Readiness and every new-work admission gate close immediately. | Lease fault and publication-controller characterization. |
| ASR-AC-105 | Original session proves continuity | `uncertain -> held` succeeds only before the deadline and only for the identical underlying session. | Process-lease unit/integration row. |
| ASR-AC-106 | Session is recreated | The process cannot treat the new session as continuous ownership and cannot resume serving. | Process-lease fault row. |
| ASR-AC-107 | Loss is confirmed or deadline expires | The process emits `application_process_lease_lost`, drains no longer than configured, exits `70`, and never reacquires in process. | Server runner plus service-backed lease-loss row. |
| ASR-AC-108 | Recovery holds exclusive serving lease at startup | Application publication/listener startup remains closed, the process reports `recovery_serving_lease_active`, and exits `2`; the application-process lease and Recovery target lease retain distinct identities and owner mappings. | Recovery sentinel and app-server service-backed rows. |
| ASR-AC-108B | Application loses its recovery serving lease | Readiness and all admission close immediately, fatal drain runs, the process reports `recovery_serving_lease_lost`, and exits `70`. | App-server lease-fault and server-runner rows. |
| ASR-AC-109 | Publication activation fails | No partial readiness, route publication, extension workspace, worker activation, listener admission, or job dequeue becomes visible. | Publication characterization and server runner rows. |
| ASR-AC-110 | Startup is canceled before serving | No false fatal event, Stage 1+ side effect, lease leak, owned-resource leak, or borrowed-resource close occurs. | Runtime/server cancellation and resource-ownership rows. |
| ASR-AC-111 | Legacy `application.process_model` is supplied | Configuration fails with `invalid_deployment_config`, path `application.process_model`, reason `unknown_key`; omission requires no default because single-active is invariant. | Platform configuration unit row. |
| ASR-AC-112 | Owner correction is reviewed | `AC-404` permits multiple provisioned instances but exactly one active process and does not contradict `EXT-REQ-213` or `REQ-04-145`; the guide contains no current-major replicated-serving instructions. | Human owner review plus Markdown/spec consistency validation. |
| ASR-AC-113 | Authored and generated extension projections are inspected | They encode `single_active` only, no component-scoped replicated serving fact is emitted, and generation/drift checks pass. | Extension owner-input tests and Make-owned drift targets. |
| ASR-AC-114 | Current-major replicated implementation is absent | No admitted configuration, runtime branch, agreement marker, component leader, exported setting, guide, or generated consumer can activate concurrent serving; future-only prerequisites remain non-executable. | Static search, boundary checks, app-server slices, and generated drift. |

#### Reference Data storage

| Acceptance ID | Scenario | Binary pass condition | Planned evidence route |
| --- | --- | --- | --- |
| ASR-AC-201 | Package boundary and port | `reference_data.Storage` retains the exact six methods; the filesystem adapter resides in `referenceassembly`; server only constructs/owns/injects it; platform imports no Reference Pack semantics. | Compile, Reference Data unit slice, and backend boundary check. |
| ASR-AC-202 | Cancellable immutable publication | Canceled writes leave no partial object; valid publication returns a new reference; invalid digest token fails before publication; repeated or concurrent same-digest successes have distinct references and preserve earlier bytes. | Filesystem adapter contract suite. |
| ASR-AC-203 | Bounded reads and logical references | Exact-limit reads return identical bytes, over-limit reads fail closed, and all returned/persisted/diagnostic references remain canonical and root-free. | Filesystem adapter suite plus logical-reference owner row. |
| ASR-AC-204 | Filesystem hostile-path behavior | Traversal, malformed references, symlink components, non-regular targets, and root replacement fail closed without outside-root access or host-path disclosure. | Rooted filesystem and Reference Data unit rows. |
| ASR-AC-205 | Root lifecycle and ownership | Repeated removal is idempotent, incomplete operations do not publish authority, root diagnostics stay private, and owned capabilities close once in reverse order. | Filesystem adapter and runtime ownership rows. |
| ASR-AC-206 | Unsupported Reference shared storage absent | No Reference object-store adapter, shared byte helper, selector, configuration branch, or runtime compatibility path remains. | Static search, boundary review, configuration rows, and app-server/Reference slices. |

#### Incident Bundle storage

| Acceptance ID | Scenario | Binary pass condition | Planned evidence route |
| --- | --- | --- | --- |
| ASR-AC-207 | Package boundary and port | `incidentbundles.BundleStorage` is unchanged; the filesystem adapter resides in `incidentportabilityassembly`; server only constructs/owns/injects it. | Compile, Incident Bundles unit slice, and boundary check. |
| ASR-AC-208 | Bundle filesystem behavior | Cancellation, unique immutable publication, bundle-ID validation, exact/over-limit reads, removal, and read-after-write equality pass without partial files. | Moved bundle storage contract suite. |
| ASR-AC-209 | Bundle hostile-path behavior | Traversal, malformed references, symlinks, non-regular targets, and root replacement fail closed without outside-root access or root disclosure. | Rooted filesystem and Incident Bundles rows. |
| ASR-AC-210 | Unsupported Bundle shared storage absent | No Incident Bundle object-store adapter, shared byte helper, selector, configuration branch, or compatibility path remains. | Static search, boundary review, and app-server/Incident Bundles slices. |

#### Runtime façade and test seams

| Acceptance ID | Scenario | Binary pass condition | Planned evidence route |
| --- | --- | --- | --- |
| ASR-AC-301 | ASW-04 same-package decomposition | `Options`, `RuntimeSettings`, all exported `Runtime` fields, constructor/method signatures, ordering, errors, and callers are byte-for-byte or API-snapshot equivalent to the baseline while `runtime.go` becomes a thin coordinator. | App-server API/static snapshot and full owner baseline. |
| ASR-AC-302 | Final lifecycle surface | Cross-package production and recovery callers compile against exactly the ASR-REQ-302 constructor/method set; `Runtime` has no exported fields and no broad runtime interface/getter set exists. | Compile, static API test, `make build-server`, recovery row. |
| ASR-AC-303 | HTTP test harness | `httptestx.Server` exports exactly `Config`, `HTTP`, and `Clock`; its runtime reference is private and used only for lifecycle cleanup. | Static harness API test and app-server slices. |
| ASR-AC-304 | Owner-specific test capabilities | No package outside `internal/app/server` accesses a `Runtime` field; each affected test uses `DB`, `Pool`, `ObjectStore`, a public owner façade, or a named minimal owner probe. | Repository static search, backend boundary check, affected owner slices. |
| ASR-AC-305 | Generic decorator removed | The exported decorator helper, `app.server.postgres_db_decorator`, its `map[string]any` value, and SQL-text observation are absent; incident-create commit failure remains covered by a typed capability. | Static assertion and incident-create failure integration row. |
| ASR-AC-306 | Production test-fault isolation | Production configuration, HTTP dependencies, environment, and application APIs cannot enable the typed commit fault; an attempted production injection fails before mutation. | Production profile and commit-fault isolation rows. |
| ASR-AC-307 | Resource and lifecycle invariants | `Close` is idempotent; owned resources close once in reverse acquisition order; borrowed Postgres/object store remain open; startup failure cleans owned state; publication/fatal/exit ordering remains unchanged. | Runtime and server unit/service-backed ownership rows. |

### 12.3 Requirement-to-evidence traceability

| Requirement | Owner or decision source | Workstream | Acceptance criteria |
| --- | --- | --- | --- |
| ASR-REQ-001 | Core 00 authority order; tracker decision | Tracker revision, ASW-01 | ASR-AC-001, ASR-AC-112 |
| ASR-REQ-002 | `nlspec-spec.md` writing discipline; tracker decision | Every workstream | ASR-AC-001 |
| ASR-REQ-003 | Tracker execution decision | ASW-00..ASW-08 | ASR-AC-002 |
| ASR-REQ-004 | Core 00 and Testing Harness authority | ASW-02, ASW-08 | ASR-AC-003, ASR-AC-113, ASR-AC-114, ASR-AC-201, ASR-AC-304 |
| ASR-REQ-101 | Extensions `EXT-REQ-213` | ASW-01, ASW-05 | ASR-AC-101, ASR-AC-102, ASR-AC-114 |
| ASR-REQ-102 | Extensions `EXT-REQ-213`, `EXT-REQ-193`; Core 04 `REQ-04-145` | ASW-05 | ASR-AC-102..ASR-AC-107, ASR-AC-110 |
| ASR-REQ-103 | Core 04 and Recovery lease contracts | ASW-01, ASW-03, ASW-05 | ASR-AC-108, ASR-AC-108B |
| ASR-REQ-104 | Selected correction to Core 04 `AC-404` | ASW-01 | ASR-AC-112 |
| ASR-REQ-105 | Core 04 closed configuration registry; Extensions owner | ASW-02, ASW-05 | ASR-AC-111, ASR-AC-113 |
| ASR-REQ-106 | Tracker implementation decision | ASW-05 | ASR-AC-102, ASR-AC-109, ASR-AC-114 |
| ASR-REQ-107 | Extensions future-only boundary; tracker decision | ASW-01, ASW-05 | ASR-AC-114 |
| ASR-REQ-201 | Core 01 semantic owner; Core 04 configuration/root owner; repository boundaries | ASW-06 | ASR-AC-201 |
| ASR-REQ-202 | Live `reference_data.Storage` surface | ASW-03, ASW-06 | ASR-AC-201 |
| ASR-REQ-203 | Live adapter behavior and adopted Core safety/lifecycle consequences | ASW-03, ASW-06 | ASR-AC-202..ASR-AC-205 |
| ASR-REQ-204 | Repository boundary decision | ASW-06 | ASR-AC-206 |
| ASR-REQ-205 | Incident Bundles port and assembly boundary | ASW-06B | ASR-AC-207 |
| ASR-REQ-206 | Core 04 filesystem-only roots; tracker compatibility decision | ASW-05, ASW-06B | ASR-AC-210 |
| ASR-REQ-207 | Live Incident Bundle filesystem behavior | ASW-03, ASW-06B | ASR-AC-208, ASR-AC-209 |
| ASR-REQ-301 | Tracker migration decision | ASW-03, ASW-04, ASW-07 | ASR-AC-301 |
| ASR-REQ-302 | Tracker lifecycle-interface decision | ASW-07 | ASR-AC-302 |
| ASR-REQ-303 | Live caller inventory and tracker decision | ASW-07 | ASR-AC-303, ASR-AC-304 |
| ASR-REQ-304 | Live test seam and tracker security decision | ASW-03, ASW-07 | ASR-AC-305, ASR-AC-306 |
| ASR-REQ-305 | Repository resource-ownership rule and current lifecycle contract | ASW-03, ASW-07 | ASR-AC-307 |

### 12.4 Completion state

| Completion boundary | Current state | Exact meaning |
| --- | --- | --- |
| Tracker decision-complete | PASS | The inventory and history are preserved; all three gaps have closed planning decisions, exact interfaces/defaults, workstreams, risks, migration posture, and binary traceability. |
| ASW-00 controlling ledger normalized | PASS | Strict sequencing and decisions are reconciled; `git diff --check` and Markdown lint passed at `.cartulary/test-results/20260805T230501Z-p1887906`. |
| Adopted-owner contradiction closed | PASS | ASW-01 changed Core 04, Extensions, and deployment guidance as one transaction; `RB-001` is done and Markdown lint passed at `.cartulary/test-results/20260805T230947Z-p1892758`. |
| Remediation implemented | OPEN | The Go prerequisite blocker is fixed, but the rolled-back ASW-02 authored/generated transaction has not been restarted and ASW-03 through ASW-07 have not run. ASW-01 specification/guidance changes remain; no ASW-02 production, test, authored projection, generated, configuration, data, or wire change is complete. |
| Remediation validated and handed off | OPEN | ASW-08 and the ordered validation ladder have not run. No implementation validation success is claimed. |

Tracker completeness does not mean the refactor is implemented. `AST-025`, `AST-016`, and the Go-readiness repair `AST-026` are complete. `AST-017` is ready to restart now that `RB-005` is resolved; all later work remains open and prohibited until ASW-02 passes.

### 12.5 Workstream acceptance evidence

| Workstream | Status | Applicable criteria | Actual evidence | Migration and rollback posture | Next action |
| --- | --- | --- | --- | --- | --- |
| ASW-00 | PASS | ASR-AC-001..ASR-AC-003 at ledger scope | `git diff --check`; Markdown run `.cartulary/test-results/20260805T230501Z-p1887906`; final-state lint `.cartulary/test-results/20260805T230624Z-p1889868` | Tracker-only delta; preserve the pre-existing user-owned tracker diff. | ASW-01. |
| ASW-01 | PASS | ASR-AC-108, ASR-AC-108B, ASR-AC-112 | Human authority/contradiction review; `git diff --check`; Markdown run `.cartulary/test-results/20260805T230947Z-p1892758` | No data or wire migration. Revert Core 04, Extensions, guide, and tracker checkpoint together before ASW-02 if rollback is required. | ASW-02 only. |
| ASW-02 | READY TO RETRY | ASR-AC-003 and ASR-AC-113 remain unsatisfied | Prior attempt failed in `codegen-toolchain` at `.cartulary/test-results/20260805T231306Z-p1897061` and was rolled back; `RB-005` is now resolved by generation root `.cartulary/test-results/20260806T023137Z-p1977974` and full-check root `.cartulary/test-results/20260806T023310Z-p1990254` | Partial v4 authored edits remain reverted; no ASW-02 generated artifact changed. ASW-01 remains the last complete refactor checkpoint. | Restart ASW-02 from the ASW-01 checkpoint; do not begin ASW-03 until it passes. |
