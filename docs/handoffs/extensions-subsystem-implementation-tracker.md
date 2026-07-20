# Extensions Subsystem Implementation Tracker

## 1. Scope and Source Posture

| Field | Recorded value |
| --- | --- |
| Target label | `extensions-subsystem` |
| Current implementation root | `internal/modules/extensions` |
| Proposed target contract | `docs/extension-subsystem-nlspec.md` |
| Planning framework | `docs/handoffs/cartulary_modular_refactor_planning_framework.md` |
| Boundary-completeness input | `temp/analysis-notes.md` (informative and non-normative) |
| Tracker output | `docs/handoffs/extensions-subsystem-implementation-tracker.md` |
| Repository baseline | Branch `revision/grid-adapter`, commit `200f631152b76cf102bb6e3f81953de820978075` |
| Initial dirty state | Pre-existing modification to `docs/extension-subsystem-nlspec.md`; initial SHA-256 `18fbd7f8c83e4a92ceec5bb0913a5443f454d8be4f96a2397df8627ec0915ee8` |
| Planning status | Planning only. Boundary decisions selected; normative owner adoption pending. This tracker does not adopt the draft or establish implementation conformance. |
| Boundary status | Decision state `SELECTED`; normative adoption state `BLOCKED`; implementation state `NOT_STARTED`. |
| Allowed change in this task | This tracker file only. |
| Required later authority | Implementation, tests, owner-document amendments, contracts, generation, migrations, configuration, Harness v2 inputs, and adoption require a later authorized task. |

The Extensions Subsystem NLSpec has `status: draft`. It is the proposed target
contract, not current implementation-conformance authority. Its mandatory
coordinated-adoption gates therefore describe future closure work. Current behavior
remains owned by adopted Core 00 through Core 04 and the adopted subsystem owners.
The local planning framework supplies planning doctrine and table structure only; it
is not evidence that a package, contract, test, or command exists.

`temp/analysis-notes.md` is informative source material for this revision. The
boundary rules selected from it are recorded canonically in Section 6. Tracker
decision selection is complete; target-contract boundary closure remains blocked.
Uppercase normative language in this tracker governs mandatory execution of this
plan only. Product-conformance authority arises only after each rule is incorporated
into and adopted by its owning Core or subsystem document.

Authority is applied in this order:

1. adopted named-subsystem NLSpecs within their allocated behavior;
2. Core 00 through Core 04 for current implementation conformance;
3. Core 05 only if a genuinely claim-bearing timed, benchmark, fixture-sensitive,
   or publication boundary is later introduced;
4. `docs/domain.md`, the Testing Harness NLSpec, and implementation-support guides
   for terminology and mechanics;
5. current code, authored contracts, tests, and generated outputs for current state;
6. this tracker and the planning framework as non-authoritative handoff material.

No adopted-owner contradiction was selected or resolved during this planning pass.
The differences between the draft and adopted Network Flow/Core discovery contracts
are target-state adoption work, not a current-owner contradiction while the draft
remains non-adopted. Any contradiction discovered during companion amendment review
must be recorded exactly as `BLOCKED: owner contradiction` and must not be resolved by
implementation convention.

### Owner documents inspected

- `AGENTS.md`; the `refactor-tracker` skill, its tracker-format reference, and the
  modular refactor planning framework.
- `docs/extension-subsystem-nlspec.md`, including the current 236 `EXT-REQ-*` IDs,
  current 141 `EXT-AC-*` rows, Section 27 artifacts, and `EXT-GATE-001` through
  `EXT-GATE-028`. Planned `EXT-AC-142` through `EXT-AC-158` do not exist
  normatively until ES-01A amends the draft.
- Core 00 through Core 04: `docs/spec/00_document_set_status_and_precedence.md`,
  `01_architecture_storage_and_view_contracts.md`,
  `02_domain_model_schema_and_history.md`,
  `03_workbook_interaction_collaboration_and_workflows.md`, and
  `04_security_deployment_and_conformance.md`.
- `docs/testing-harness-nlspec.md`, `docs/domain.md`,
  `docs/network-flow-activity-nlspec.md`,
  `docs/reporting-subsystem-nlspec.md`,
  `docs/report-composition-nlspec.md`, and
  `docs/opentelemetry-instrumentation-nlspec.md`.
- Core 00/01/02/03/04 Incident Portability and physical-backup sections. There is no
  separate adopted Incident Portability NLSpec in the inspected repository; its
  current behavior is distributed across Core owners and `internal/modules/incidentbundles`.

### Repository evidence inspected

- `temp/analysis-notes.md` as a non-normative boundary-completeness input; it is not
  an owner document, contract, or implementation-conformance source.

- Every file under `internal/modules/extensions`, including its testsupport package.
- `internal/platform/httpapi/extensions.go`, `httpapi.go`, and their tests;
  `internal/app/server/runtime.go`, `runtime_routes.go`, and route-composition tests;
  platform configuration, jobs, object-store, PostgreSQL, and telemetry boundaries.
- Workbook startup registry/routes/tests; `apps/web/src/app/App.tsx`,
  `app/api/appShellClient.ts`, the incident-directory debug harness, relevant browser
  support tests, and Network Flow browser/workbook consumers.
- `contracts/extensions/index.json`, `contracts/index.json`, the OpenAPI extension
  route/schema, `tools/contractgen`, generated Go/TypeScript projections, and
  `tools/generated_artifact_policy.json`.
- Network Flow module facade, routes, stores, import facade, transaction participants,
  configuration, migrations `00028`, `00029`, `00030`, and `00033`, and its current
  verification owner/family rows.
- Reporting, Report Composition, Incident Bundle, jobs, backup/configuration, and
  telemetry implementation surfaces named by searches and opened before conclusions.
- Harness v2 verification registry, test-owner registry, family manifests, runner
  registry, runtime/resource/fixture profiles, execution topology, render index,
  schema attachments, and evidence-audit command surface.

### Fixed planning assumptions and dependencies

- BC-001 through BC-017 are selected and are not reopened by this tracker revision.
- Network Flow's empty authoritative state is valid and therefore selects
  `empty_state_policy=allowed`.
- Portability import mutates only through the shared transaction protocol.
- Restore v1 targets a stopped empty deployment.
- Browser contract v1 has exactly one `standard` build class.
- Capability advertisement remains entirely disabled in v1.
- PostgreSQL advisory locking is permitted supporting implementation guidance, not a
  normative mechanism or required storage coupling.
- The system remains statically packaged; runtime-downloadable executable extensions
  remain deferred.
- Existing requirements are amended in later normative work. No new `EXT-REQ` or
  `EXT-GATE` ID is allocated; exactly 17 acceptance IDs are planned.

### Explicit non-goals

- No production, test, owner-specification, contract, schema, generated-artifact,
  dependency, migration, configuration, harness, fixture, or lockfile change.
- No runtime package installation, marketplace, arbitrary callback bus, separate
  extension host, or independently distributed executable extension format.
- No transfer of Core 00 recognition/claimability, Core 01 public discovery/dispatch,
  Reporting, Report Composition, Incident Portability, backup, or OpenTelemetry
  ownership to `internal/modules/extensions`.
- No phase identity, delivery phase, test row, fixture family, or adoption gate used
  as runtime architecture.
- No `EXT-FIX-*`, extension fixture-result schema, v1 alias, compatibility reader,
  newest-run lookup, or historical retained-run fallback.

## 2. Current-State Repository Inventory

All three target files are accounted for below. Adjacent rows are included because
the current responsibility is materially distributed outside the target path.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/extensions/api.go` | Constructs the current three-member discovery item and `{extensions: [...]}` response data. | `BuildResource`, `BuildResponseData` | `routes.go`; response behavior is exercised through server and browser tests. | `internal/platform/httpapi.ExtensionProfile` | Incident HTTP/integration/request suites and `apps/web/e2e/incident.support.spec.ts` through the route. | Current OpenAPI resource and generated Go/TS types. | Split: Extensions normalization facade plus Core 01 discovery producer. | high | Current shape is `profile_id`, `claimed`, `route_families`; draft target is the coordinated generic seven-member item. |
| `internal/modules/extensions/routes.go` | Registers authenticated `GET /api/v1/extensions`, validates singleton query shape, slides the session, and writes common envelopes/errors. | `Service`, `RegisterRoutes` | `internal/app/server/runtime.go` built-in route contribution. | `authn`, `httpauth`, `httpapi`, PostgreSQL handle, master keys, clock. | Process/integration HTTP suites, browser support route checks, and common auth/envelope tests. | OpenAPI path and generated clients/types. | Split: Core 01 route contract; platform auth/session adapter; small Extensions query facade. | high | The package directly composes platform authentication and transport concerns; no registry generation or claim-resolution logic lives here. |
| `internal/modules/extensions/testsupport/routetest/routes.go` | Contributes the discovery route to shared route-inventory assertions. | `PublicDiscovery` | `internal/modules/incidents/http_conformance_test.go` | `internal/testutil/routeinventory` | Shared HTTP conformance suite. | None. | Keep as owner-aligned testsupport only if `module.extensions` owns the row; otherwise move to Core 01/app-server testsupport. | medium | Current entry freezes method/path/envelope only, not the target seven-member response. |
| `internal/platform/httpapi/extensions.go` | Holds the process-global recognized profile list, claim flags, route families, workspace metadata, reservation matching, cloning, and testing override. | `ExtensionProfile`, `ExtensionWorkspace`, profile resolution/claim/reservation functions | Server assembly, route wrapper, Network Flow, workbook startup, tests. | Standard library only. | `internal/platform/httpapi/httpapi_test.go`, workbook and module tests. | Parity-tested against `contracts/extensions/index.json`. | Split: Core 00 registry facts, Core 01 route reservation, Core 03 workspace declarations; generated Extensions consumers. | critical | Current code creates recognition from implementation data, which the draft forbids after adoption. Core ownership must not be transferred. |
| `internal/platform/httpapi/httpapi.go` | Builds the HTTP mux and wraps it with unclaimed reserved-family dispatch before common handling. | `DependencySet`, `RouteRegistrar`, `NewHandler`, envelope/error APIs | Server runtime and test harness composition. | Web assets, config, jobs, object store, PostgreSQL, telemetry, WebSocket. | HTTP API tests cover reserved-unclaimed outcomes. | OpenAPI/common error contracts. | Core 01/platform HTTP runtime. | critical | Current reserved-unclaimed wrapper returns `404 extension_profile_not_claimed` before the inner handler. Exact precedence must be characterized before replacement. |
| `internal/app/server/runtime.go` and `runtime_routes.go` | Resolves profile claims, feeds telemetry, constructs Network Flow, registers the Extensions route in fixed built-in order, and injects profiles into HTTP/workbook dependencies. | `NewRuntime`, `Runtime`, route composition helpers | `cmd/server`, server harness/tests. | Modules and platform runtime dependencies. | Runtime route-order, profile, process, and integration tests. | None directly. | Application assembly; future Extensions coordinator is injected here, not implemented here. | critical | `network_flow_activity` and enterprise-auth claims are applied from config; current assembly is not the draft's six-stage atomic publication coordinator. |
| `contracts/extensions/index.json` | Authored derived registry input with recognized profile IDs and route families. | `registry_id=cartulary.extensions.phase2.v1` | Contract generator and parity tests. | Core 00 owner sections declared in `contracts/index.json`. | Contractgen validation and `httpapi` parity test. | Generates `internal/gen/contracts/**` and `packages/protocol-ts/src/generated/**`. | Replace through owner inputs plus Extensions generator after owner adoption. | critical | Historical phase-shaped identity must not survive adoption or become a compatibility alias. |
| `contracts/openapi/cartulary.openapi.yaml` | Owns the current discovery route and three-member resource schema. | `/api/v1/extensions` and generated public types | Server/client generated consumers. | Core 01 contract ownership. | OpenAPI, backend route, browser support, reporting boundary tests. | Go/TS contracts and validators. | Core 01. | critical | Target producer and decoder changes require coordinated contract-major adoption. |
| `tools/contractgen/**`, `contracts/index.json`, generated roots | Validates the single current extension index and emits Go/TS embedded contracts. | Contract-family `extensions` | `make generate`, drift checks, builds. | Authored contracts and family registry. | Contractgen and generated-artifact checks. | `internal/gen`, protocol TS generated files. | Extensions generator plus contract-generation platform. | critical | Current validation admits only `contracts/extensions/index.json`; Section 27 requires a much larger generated artifact family. Generated roots must never be hand-edited. |
| `apps/web/src/app/api/appShellClient.ts`, `apps/web/src/app/App.tsx`, debug harness | Fetches and consumes current discovery during app startup and uses claim state for feature availability. | `ExtensionProfileResource`, `fetchExtensions` and application state | Web application startup and debug surface. | Generated protocol types and HTTP client helpers. | App landing/unit tests and browser support. | Generated TS discovery type. | Core 03/web application, consuming a generated client support registry. | critical | Current browser does not gate on digest-bound support registry plus availability epoch/generation. |
| Workbook startup backend and web models | Validates `extension_workspace` sheet refs, declared workspace, claim state, and minimum role; renders Network Analysis when eligible. | Startup DTOs/registry and URL state | Workbook routes/controllers/shell. | `httpapi.ExtensionProfile`, auth/membership, generated startup contracts. | Backend startup tests, workbook unit/browser tests. | OpenAPI startup schemas. | Core 01/03 and web application. | high | Must preserve Base startup, drafts, pending queue, WebSocket identity, and fallback semantics. |
| `internal/modules/networkflow/**` | Owns Network Flow resources, routes, import facade, stores, transaction participants, graph adapter, security, and claimed-route registration. | `Module`, `NewModule`, `ImportOwner`, `RegisterRoutes`, profile constants | Server assembly, imports, browser workspace. | PostgreSQL, Imports port, Incidents/Indicators participants, key rings. | Unit/store/integration/browser/accessibility/visual/measurement rows. | Network Flow contracts and generated consumers. | `module.networkflow` named profile owner. | critical | Adopted v1.2.0/major 1 currently competes with the draft's proposed generic discovery and major-2 action. Profile behavior stays owned here. |
| `db/migrations/00028`, `00029`, `00030`, `00033` and Network Flow queries | Current Network Flow import linkage, authoritative tables/rows/diagnostics/bindings, and indexes. | SQL schema | Network Flow/Imports stores. | PostgreSQL. | Store and integration tests; migration drift. | sqlc outputs where applicable. | Network Flow state owner plus migration application facade. | critical | No generic extension metadata, migration ledger, state-presence, or physical-binding schema was found. Future authored migrations require later authorization. |
| `internal/modules/incidentbundles/**` and Core Incident Portability sections | Whole-incident export/import, bundle files, attribution, jobs, and publication. | Route/store/worker facades | Server assembly and revision assembly. | Object store, jobs, incidents, revisions. | API/integration/worker tests. | Incident-bundle OpenAPI/contracts. | Incident Portability/Core 01 shared owner. | high | Draft participation must be typed; inactive blocking must be declarative and must not execute profile code. |
| `internal/modules/reporting/**` and `internal/modules/reportcomposition/**` | Snapshot/report render/release and authored composition behavior. | Reporting and composition routes/services/providers | Server, jobs, browser reporting surfaces. | PostgreSQL, graph projection, object store, jobs. | Unit/integration/traceability/OpenAPI tests. | Reporting/composition contracts and generated outputs. | Existing named owners. | high | Generic participation may be imported, but ownership cannot move to Extensions. Record no-change parity when interfaces do not change. |
| Platform config, jobs, object store, PostgreSQL, telemetry | Current claim keys, job shells, storage adapters, transactions, and claimed-profile telemetry serialization. | Platform facades | Server and modules. | External services and standard libraries. | Platform unit/integration/conformance tests. | Telemetry contracts and config examples. | Core 01/Core 04 and platform owners. | critical | `cartulary.profile.claims` currently uses a hard-coded known-profile set; target derives it from the canonical resolved claim set and digest. |
| Harness v2 registries and topology | Routes exact test evidence by owner/family/verification/runner/profile. | Verification registry, owner registry, family manifests, runner registry, execution topology. | Public Make targets and owner-slice scheduler. | Authored JSON plus generated topology projections. | Harness contract/json-shape/generated-drift suites. | Generated task surface, render index, browser batch manifests. | Testing Harness v2. | critical | No `module.extensions` owner exists. `task-guide` and `explain-test-owner` fail for it today. |

## 3. Module Boundary Diagnosis

The current target is a shallow transport facade, not a durable subsystem boundary.
The proposed subsystem needs a small coordination facade around generated immutable
inputs while leaving fact ownership, transport/runtime plumbing, named-profile state,
and shared-owner behavior with their existing authorities.

### Current-state versus proposed-target boundary map

| Concern | Current state | Proposed target boundary | Decision |
| --- | --- | --- | --- |
| Recognition and claimability | Hard-coded in `internal/platform/httpapi/extensions.go`. | Core 00 remains the sole fact owner; generated owner inputs are consumed by Extensions. | split |
| Discovery | Target package builds three fields; Core OpenAPI owns the route. | Core 01 owns producer/decoder/route; Extensions supplies validated descriptor-derived data. | split |
| Claim resolution | Server applies selected config claims; modules self-check claim flags. | Extensions coordinator executes the closed dependency/admission algorithm using Core 04 configuration inputs. | move |
| Route dispatch | HTTP wrapper matches hard-coded reserved families; claimed modules register directly. | Core 01 owns Base reservation registry, overlap rules, and public dispatch precedence; Extensions supplies registry facts. | split |
| Registry generation | One phase-shaped authored index plus narrow contractgen validator. | Extensions generator consumes digest-bound owner manifests/fragments and emits canonical registry/integrity artifacts. | move |
| Implementation bindings | Go assembly and constants imply availability. | Build-owned packaged bindings must match descriptors and integrity digests; runtime consumes bindings only. | split |
| State ownership | Network Flow owns its SQL; no generic state metadata/ledger exists. | Named profiles own authoritative families; Extensions owns generic metadata/ledger contracts and scoped coordination. | split |
| Migration | Ordinary DB migrations plus profile code; no generic extension migration coordinator. | Profile owners author definitions; Extensions coordinates locks, scoped contexts, pending validation, ledger, and exact-once final validation. | split |
| Jobs | Platform common manager; modules own job behavior. | Core 01 retains common shell; profile owners define job-kind contracts; Extensions coordinates proof/reconciliation rules. | split |
| Transactions | Network Flow composes PostgreSQL/incident/audit/indicator participants directly. | Core 01 owns bounded cross-owner protocol/final commit; participants remain with semantic owners. | split |
| Staged objects | Object-store adapters and job/module code. | Core 01 object-storage owner controls access cutoff/cleanup; Extensions supplies typed staged-object contract and fatal conditions. | split |
| Backup/restore | Core configuration and incident-bundle paths; profile SQL is physically included by deployment backup. | Physical backup owner orchestrates codecs/bindings; profile owners declare state; Extensions validates parity and ordering. | split |
| Portability | Incident Bundle owns public export/import. | Incident Portability owns operation; Extensions supplies closed participant interfaces and declarative inactive blocker. | split |
| Reporting | Reporting/Composition modules own behavior. | Existing owners retain behavior and import generic participant interfaces when needed. | keep |
| Browser | App startup consumes discovery; workspace shell uses claim/workspace facts. | Core 03/web app intersects discovery, client-support registry, authorization, and no-store availability epoch/generation. | split |
| Security/configuration | Core config/platform secret handling plus module-specific checks. | Core 04 owns claim keys, syntax-only inactive policy, timeouts, lease, readiness, fatal lifecycle; profiles own local schemas. | split |
| Diagnostics | Common API errors plus module diagnostics; no canonical extension condition registry. | Extensions generator emits validation-condition registry; Core 04 owns startup/fatal presentation and exit behavior. | split |
| Observability | Telemetry receives a server-derived claimed-profile string. | OpenTelemetry owner derives the signal from the canonical resolved claim set/digest; Extensions exposes no secrets. | split |
| Conformance accounting | Existing owners only; no `module.extensions`. | Static Extensions accounting plus Harness v2 owner/selector/evidence accounting, without runtime coupling. | move |

### Responsibility map

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Authenticated collection route | `extensions/routes.go` | Core 01 route plus platform auth adapter | split | Core 01 owns the public route/envelope; current code imports auth/platform directly. | Preserve method, query rejection, auth, session sliding, envelopes until coordinated change. |
| Discovery data normalization | `extensions/api.go` | Extensions facade | keep | Target draft allocates descriptor consumption to Extensions. | Input must become generated descriptors, not hard-coded profile structs. |
| Recognized profile vocabulary | `httpapi/extensions.go` | Core 00 | move | Core 00 current owner and draft Table 1-A. | Remove implementation-created recognition only after generated parity exists. |
| Route reservation matching | `httpapi/extensions.go` and HTTP wrapper | Core 01/platform HTTP | move | Current public dispatch behavior and Core 01 ownership. | Must consume generated Base and extension registries. |
| Workspace declarations | `httpapi.ExtensionProfile.Workspaces` | Core 01 identity plus Core 03 behavior | split | Workbook registry consumes the field; draft separates identity and rendering. | Browser support registry adds build compatibility, not ownership. |
| Claim config application | `server.applyConfigExtensionClaims` | Core 04 plus Extensions coordinator | move | Only two profiles currently read explicit config here. | No partial serving during transition. |
| Named-profile implementation | Network Flow, Imports, Incident Bundle, Reporting, etc. | Existing named module owners | keep | Current adopted owners and repository module boundaries. | Extensions coordinates closed interfaces only. |
| Application composition | `internal/app/server` | Application facade | keep | Repository procedure designates exact composition root. | Inject the coordinator and bindings; do not add product logic. |
| State metadata/migration ledger | absent | Extensions module with profile-owned definitions | split | Draft Sections 19/21/27; no current generic schema found. | Authored SQL/migrations are later behavior-changing work. |
| Cross-owner transaction engine | PostgreSQL/module-specific composition | Core 01/platform transaction coordinator | move | Draft explicitly leaves final-commit protocol with Core 01. | Extensions supplies participant contracts, not a second transaction engine. |
| Backup codec/physical binding | deployment backup and module stores | Backup platform plus durable profile owners | split | Draft Table 27-A and Core physical-backup ownership. | Filesystem remains derived-only for extension state. |
| Client lifecycle state | `apps/web` controllers/shell | Core 03/web application | keep | Browser state is application behavior. | Extensions provides generated support facts only. |
| Telemetry claim serialization | `internal/platform/telemetry` | OpenTelemetry/platform telemetry | keep | Adopted OTel NLSpec owns signal shape. | Replace hard-coded known set only in coordinated adoption. |
| Contract generation | `tools/contractgen` | Contract tooling plus Extensions generator | split | Existing family generation and Section 27 target artifacts. | Owner inputs authored; projections generated. |
| Harness evidence routing | Harness v2 registries/topology | Harness owners and `module.extensions` test owner | split | Current registries lack the owner; draft defines two verification IDs. | Test rows are evidence accounting, never runtime architecture. |
| Independent executable package support | absent | Future NLSpec | defer | Draft Section 30 explicitly makes it future-only. | Do not reserve a format or compatibility reader. |

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `GET /api/v1/extensions` method/path | Core 01; target route facade | OpenAPI and `RegisterRoutes` | Shared route inventory; browser support request | Exact GET success, non-GET 405, unknown query rejection, request ID, auth/session sliding | high | Behavior-preserving until the coordinated discovery-major change. |
| Current discovery envelope | Core 01 | `{data:{extensions:[...]},meta}` with three fields | Incident/browser support tests | Snapshot every recognized/claimable/claimed ordering and omitted forbidden fields | critical | Target seven-member shape is `requires later authorization`. |
| Tolerant discovery decoder | Core 01/web | Current generated type and app casts | App landing/unit tests | Known-member malformation rejection; additive unknown-member ignore; no unknown execution | critical | Producer stays strict while decoder becomes tolerant only with owner amendment. |
| Reserved-unclaimed routes | Core 01/platform HTTP | `withUnclaimedReservedExtensionFamilies` | `httpapi_test.go` and Network Flow route tests | Every exact/ancestor/descendant route; verify precedence before auth, incident, payload, cursor, job, resource lookup | critical | Preserve `404 extension_profile_not_claimed` until adopted owner changes exact token/status. |
| Claimed-family dispatch | Core 01 plus named profile | Network Flow registers only when claimed | Network Flow integration/browser tests | Authorization denial and empty-resource outcomes never collapse to unclaimed | critical | Profile-local errors begin after Core dispatch. |
| Recognition/claimability | Core 00 | Hard-coded six-profile list plus Core registry | Contract parity and config tests | Recognized-unclaimable, zero-profile generation, no implementation-created profile | critical | Core 00 remains sole owner. |
| Claim configuration and resolution | Core 04/Extensions target | Server config application | Core config and startup tests | Omitted/false/null/type cases, dependency order/cycles, preflight, no listener/worker on failure | critical | Characterize exit `2` and no partial profile set. |
| Startup publication and fatal lifecycle | Core 04 | Current runtime construction/cleanup | Process/bootstrap tests | Second process, crash release, lease loss at three stages, Stage 6 atomic serving, repeated fatal exit `70` | critical | No current complete implementation was found. |
| Implementation binding admission | Build/Extensions target | Current Go composition only | Module construction tests | Missing/extra/major/digest/capability/state/job/participant mismatch before migration | critical | Packaged binding artifact is new. |
| Migration state and results | Profile owners/Extensions target | Network Flow SQL migrations only | Migration/store tests | Missing path, too new/old, immutable digest, lock/step/profile timeout, resumability, exact-once final validation | critical | Closed scoped contexts; no cross-owner access. |
| Cross-owner transaction result | Core 01 | Current Network Flow transaction participants | Store/integration tests | Every ordered failure/cancel position, bounds, lock order, conflict/timeout/commit/no-partial-effects | critical | No automatic retry. |
| Staged-object lifecycle | Core 01 object storage | Object store/job paths | Object-store and job tests | Access cutoff independent of cleanup, batching, every deletion outcome, retry, readiness degradation, fatal contradiction | critical | Publication must be atomic with final DB commit. |
| Job proof/reconciliation | Core 01 jobs plus profile owner | Current common manager/runner | Jobs and module tests | Proof required/prohibited, replay, cancellation precedence, contradictory proof, inactive reconciliation | high | Internal owner profile identity is mandatory. |
| Backup/restore | Physical backup owner/profile owner | Config backup tests and module storage | Backup/config/store tests | Claimed/unclaimed state, codec ID/digest, empty binding, historical codec, unsupported codec, restore ordering | critical | Inactive restore executes no profile code. |
| Incident portability | Core 01/Incident Bundle | Incident-bundle module | Bundle API/integration/worker tests | Every claim/state matrix, declarative inactive blocker, pre-publication failure, no profile execution inactive | high | Preserve public bundle ownership. |
| Reporting participation | Reporting/Report Composition | Adopted NLSpecs/modules | Reporting/composition tests | No-participation omission and typed participant matrices | high | Record no-change parity if generic interfaces do not change their owner contracts. |
| Workbook extension workspace | Core 03/web | Startup registry and shell | Backend/web workbook tests | Discovery/support/auth intersection, lazy load, fallback, unsupported major, stale generation, authorization loss | critical | Preserve Base tabs, caches, requests, queue, optimistic state, drafts, and stable client/WebSocket identity. |
| WebSocket behavior | Core 03/platform WS | Base collaboration implementation | Collaboration/workbook tests | Unknown extension values, authorization/session loss, epoch rollover without Base identity reset | high | Draft adds no arbitrary extension event bus. |
| Security and egress | Core 04/profile owners | Config/secret/Network Flow security | Config/security/telemetry tests | Inactive syntax-only no resolution/DNS/connect/profile code; egress-none across all call sites; secret-negative assertions | critical | Diagnostics and telemetry must remain content-free. |
| Telemetry claim identity | OpenTelemetry | `SerializeProfileClaims` and OTel NLSpec | Telemetry resource/privacy/conformance tests | Canonical resolved set/digest, unknown profile rejection, ordering, no secret or incident content | high | OTel remains signal owner. |
| Generated contracts | Owner specs/contracts/tooling | Contract family and generated policy | Generate/json-shape/drift tests | Full Section 27 schema/identity/digest/parity/limit vectors | critical | Never hand-edit generated roots. |
| Harness owner accounting | Testing Harness v2 | Current registries/topology; absent owner | Harness contract tests | Exact row resolution, full-owner versus subset, paired shards, evidence-class gates, exact-root audit, stale-root rejection | critical | Aggregate `check` success is insufficient. |
| Core 05 publication | Core 05 | No draft publication behavior | Existing benchmark claim check | None unless a later owner introduces a claim-bearing timed/fixture-sensitive boundary | low | Out of scope for current adoption. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Profile recognition, claim state, reservations, and workspaces share one platform struct. | `internal/platform/httpapi/extensions.go` | Ownership drift and implementation-created facts. | must_fix | Core 00/Core 01/Core 03 plus generated Extensions inputs | Split facts by owner and consume an immutable generated registry. |
| The target route facade imports auth store, master keys, HTTP auth, and transport response helpers directly. | `internal/modules/extensions/routes.go` | A future deep module could become transport/platform coupled. | should_fix | Platform HTTP/auth adapter plus Extensions query facade | Keep authentication/session behavior in route adapter and expose a narrow discovery query port. |
| Server assembly performs partial claim application and telemetry derivation before any generic coordinator exists. | `internal/app/server/runtime.go` | Partial publication and inconsistent claim consumers. | must_fix | Application assembly injecting Core 04/Extensions coordinator | Introduce one resolved claim-set result and atomic publication plan in a later slice. |
| Contract identity is phase-shaped. | `registry_id=cartulary.extensions.phase2.v1` | Historical delivery identity could leak into runtime and compatibility. | must_fix | Extensions generator/Core 00 contract family | Replace after owner inputs are adopted; do not alias or compatibility-read it. |
| Generated contract validation permits only one narrow extension artifact. | `tools/contractgen/validation.go` | Section 27 artifacts cannot be represented or drift-checked. | must_fix | Contract tooling/Extensions generator | Add authored schemas/owner inputs first, then generator outputs and drift checks. |
| Network Flow currently owns a competing discovery shape/major 1 while the selected target requires major 2. | Adopted Network Flow NLSpec and current contracts | Premature target implementation would break adopted behavior. | must_fix | Core 00/01/03/04 plus Network Flow owner | Follow the ten-step coordinated amendment and promotion order; no intermediate adoption. |
| Current browser trusts discovery and local code support without a digest-bound support registry/availability generation. | App startup and workbook shell | Stale or unsupported UI can render. | must_fix | Core 03/web application/client build | Add generated support registry and no-store availability epoch/generation atomically. |
| Generic extension state metadata, migration ledger, presence manifest, physical binding, and codec artifacts are absent. | SQL/config/contract searches | State adoption cannot be proven by code alone. | must_fix | Extensions generator, profile owners, backup owner | Author schemas/migrations only after companion owner authorization. |
| Cross-owner transaction semantics are implemented profile-by-profile. | Network Flow module/store participant composition | Inconsistent deadlines, ordering, replay, and final commit. | must_fix | Core 01 transaction coordinator | Characterize current behavior, then adopt one bounded protocol with typed participants. |
| Incident portability/reporting/backup interfaces are not generic Extensions participant contracts. | Current modules and owner docs | Ownership transfer or implicit participation risk. | should_fix | Existing shared owners | Amend only imported interfaces; otherwise record exact no-change parity. |
| Telemetry validates a hard-coded profile vocabulary. | `internal/platform/telemetry/resource.go` and privacy registry | Drift from Core 00 registry. | must_fix | OpenTelemetry/platform telemetry | Derive from canonical claim set and digest after adoption. |
| `module.extensions` is absent from every owner-first Harness input. | Verification/owner/family registries and topology; Make failures | No owner slice or evidence audit can close adoption. | must_fix | Harness v2 plus `module.extensions` | Add both verification IDs, nonempty exact-selector family, topology, and projections together. |
| Framework catalog does not list an `extensions` module. | Planning framework module table versus live target path | Treating framework as repository truth would erase the target seam. | intentional/no_action | Tracker | Record the mismatch and use live repository evidence. |
| No dynamic executable extension system exists. | Code/config/route searches and draft non-goals | Accidental scope expansion. | intentional/no_action | Future owner | Preserve absence; defer any package format to a future NLSpec. |
| No direct vendor-grid import was found in the Extensions backend target. | Target file inspection | None for this seam. | intentional/no_action | Grid adapter/web owners | Keep vendor semantics out of Extensions work. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Authority and baseline freeze | root | none | WF-01 | Freeze exact draft, adopted owners, worktree, public contracts, and source digests. | Owner docs, tracker | Markdown/source lint; recorded digests | No source ambiguity and no owner contradiction selected. |
| WF-01 | Characterization closure | chain | WF-00 | WF-02, WF-03, WF-04 | Add exact pre-change evidence for every observable current boundary. | Existing backend/web tests and future owner rows | Narrow current owner slices | Every risky move has pre-change evidence. |
| WF-02 | Companion-owner coordination | chain | WF-00, WF-01 | WF-02A | Amend Core/profile/shared owners or record exact permitted no-change parity. | Core 00-04, Network Flow, Reporting, Composition, OTel, domain, Harness | Normative-source and documentation checks | All owner versions/digests/anchors are adoption-ready; draft remains draft. |
| WF-02A | Normative boundary closure | chain | WF-02 | WF-03, WF-04, WF-05, WF-06, WF-07, WF-08 | Apply BC-001 through BC-017 to the draft and companion-owner amendment set; assign exact owner anchors, schemas, algorithms, acceptance criteria, and selectors while the Extensions NLSpec remains draft. | Draft, Core/profile/shared owner documents, traceability inputs | Normative-source, traceability, and documentation checks | Every BC row has exact target anchors and an active acceptance/verification mapping; no generation or implementation starts from the pre-closure draft. |
| WF-03 | Owner manifests and canonical generation | chain | WF-02A | WF-04, WF-08 | Produce digest-bound inputs, descriptors, registry, integrity, closure, schemas, and projections. | `contracts/**`, generator inputs, generated roots | `make generate-drift`, JSON shape, artifact policy | All artifacts are generator-owned and byte-stable. |
| WF-04 | Runtime coordination and publication | chain | WF-01, WF-02A, WF-03 | WF-05, WF-06, WF-07 | Implement claim resolution, admission, bindings, dependencies, lease, publication, diagnostics, and fatal lifecycle. | Extensions facade, app server, config/platform adapters | Backend unit/process/integration; security | No listener/worker/route is partially published. |
| WF-05 | Core discovery and dispatch | parallel | WF-02A, WF-03, WF-04 | WF-07, WF-09 | Move producer/decoder/reservation/dispatch to generated generic contracts without transferring Core ownership. | Core OpenAPI, HTTP runtime, Extensions query facade | Go integration/browser support | One seven-member producer and one dispatch precedence. |
| WF-06 | State, migration, jobs, transactions, and storage | parallel | WF-02A, WF-03, WF-04 | WF-08, WF-09 | Add scoped generic coordination while preserving profile-owned state and Core shared protocols. | Extensions, Core 01 transaction/storage/jobs, profile stores/migrations | Service-backed owner slices, migration drift | State and commit outcomes satisfy every closed matrix. |
| WF-07 | Browser support and lifecycle | parallel | WF-02A, WF-03, WF-04, WF-05 | WF-09 | Add support registry, availability generation, eligibility intersection, fallback, and cleanup. | Web app, generated protocol/UI contracts, workbook startup | Typecheck, unit, browser/stateful/a11y/visual as allocated | Base state remains stable and stale extension state cannot render. |
| WF-08 | Named-profile and shared-owner adoption | parallel | WF-02A, WF-03, WF-06 | WF-09 | Adopt Network Flow major 2 and typed portability/reporting/backup participation without ownership transfer. | Network Flow, Incident Bundle, Reporting, Composition, backup | Affected full-owner slices | Every participant and parity row resolves to its primary owner. |
| WF-09 | Harness v2 onboarding and traceability | chain | WF-03, WF-05, WF-06, WF-07, WF-08 | WF-10 | Register owner-first contracts, exact selectors/profiles, clause traceability, paired shards, and evidence audit. | Verification/owner/family/topology authored inputs and generated projections | Harness contract, task guide, full-owner slices, evidence audit | Both Extensions verification IDs, all planned boundary acceptance criteria, and every imported-owner obligation exist and resolve. |
| WF-10 | Atomic adoption and final handoff | chain | WF-09 | none | Execute all gates, audit retained roots, and promote all companions together. | Owner docs and status metadata only after evidence | Full drift/check/release gates plus exact-root audit | All 28 gates are `DONE`; no intermediate artifact claimed adoption. |

No schema generation, implementation, browser conversion, or participant work may
begin from the pre-closure draft. WF-03 through WF-08 MUST consume the WF-02A
closure set, and WF-09 MUST reject any planned boundary criterion that is absent,
unresolved, or lacks an exact selector.

### Canonical boundary-closure ledger

This table is the sole tracker definition of the selected boundary rules. Other
sections reference BC IDs rather than restating the rules. `SELECTED` records a
planning decision, not normative adoption. Owner requirement anchors become exact
under ES-01A through ES-04; verification IDs and selectors become exact under ES-05.

| BC ID | Planned acceptance | Required target rule | Normative owners | Main slices | Decision | Normative adoption | Implementation | Required future bindings |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| BC-001 | EXT-AC-142 | Add `empty_state_policy=allowed\|forbidden`; metadata never makes state present; empty initialization is permitted only when allowed; metadata-present/empty state is valid only under `allowed`; Network Flow selects `allowed`. | Extensions, Core 02, Network Flow | ES-01A, ES-02, ES-03, ES-11 | SELECTED | BLOCKED | NOT_STARTED | Owner anchors: ES-01A through ES-03; verification/selectors: ES-05. |
| BC-002 | EXT-AC-143 | Establish one validation precedence: invocation failure, structural invalidity, overflow, remaining schema defects, valid findings, valid empty result. Counts `257..4096` violate 256-bound schemas; `4097+` selects overflow. | Extensions, Core 04 | ES-01A, ES-02, ES-10 | SELECTED | BLOCKED | NOT_STARTED | Owner anchors: ES-01A and ES-02; verification/selectors: ES-05. |
| BC-003 | EXT-AC-144 | Split portability export/import results; make import side-effect-free preparation followed by shared transaction participation; add scoped staged-output capability; set participant and aggregate import input ceilings to 64 MiB. | Extensions, Core 01, Incident Portability, profile owners | ES-01A, ES-02, ES-04, ES-11 | SELECTED | BLOCKED | NOT_STARTED | Owner anchors: ES-01A, ES-02, and ES-04; verification/selectors: ES-05. |
| BC-004 | EXT-AC-145 | Add authored `cartulary.extension_dependency_declaration_set.v1`; arrays are always present; null is invalid; manifests supply versions/digests; generation emits the snapshot. | Extensions and dependency owners | ES-01A, ES-06, ES-07 | SELECTED | BLOCKED | NOT_STARTED | Owner anchors: ES-01A; declaration/schema bindings: ES-06; verification/selectors: ES-05. |
| BC-005 | EXT-AC-146 | Add `recognized_profile.primary_owner_contract_ref`; define the sole source of every descriptor member; reject missing/multiple scalar sources and duplicate set members; prohibit prose or code inference. | Core 00, Extensions generator | ES-01A, ES-02, ES-06, ES-07 | SELECTED | BLOCKED | NOT_STARTED | Owner anchors: ES-01A and ES-02; generator bindings: ES-06; verification/selectors: ES-05. |
| BC-006 | EXT-AC-147 | Add condition annotations for generated schema rules and closed decision tables for procedural validators; every admitted validation surface must have a complete condition inventory; unregistered emitted conditions fail conformance. | Extensions plus every validation owner | ES-01A, ES-02 through ES-04, ES-06, ES-07, ES-10 | SELECTED | BLOCKED | NOT_STARTED | Owner anchors: ES-01A through ES-04; registry bindings: ES-06; verification/selectors: ES-05. |
| BC-007 | EXT-AC-148 | Replace every “at least” closure-category derivation with an exact subject/contribution-kind mapping; generated subject rows permit no owner-authored not-applicable reason; only fixed baseline rows retain their enumerated reasons. | Extensions generator | ES-01A, ES-06, ES-07 | SELECTED | BLOCKED | NOT_STARTED | Owner anchors: ES-01A; generator bindings: ES-06; verification/selectors: ES-05. |
| BC-008 | EXT-AC-149 | Close clause kinds, parent kinds, zero-based ordinals, half-open byte ranges, parent scope, clause-ID digest input, and document-clause mapping; add authored `cartulary.extension_traceability_mapping_source.v1`. | Extensions documentation tooling, Harness | ES-01A, ES-05, ES-06, ES-07, ES-09 | SELECTED | BLOCKED | NOT_STARTED | Owner anchors: ES-01A; Harness and selector bindings: ES-05; schema bindings: ES-06. |
| BC-009 | EXT-AC-150 | Add non-null inactive schema reference exactly for `syntax_only`; restrict its vocabulary to inert structural validation; while inactive apply no required/default omission policy, create no configuration view, perform no resolution, and discard accepted values. | Extensions, Core 04, profile owners | ES-01A, ES-02, ES-03, ES-10 | SELECTED | BLOCKED | NOT_STARTED | Owner anchors: ES-01A through ES-03; verification/selectors: ES-05. |
| BC-010 | EXT-AC-151 | Define `unacquired -> acquiring -> held -> uncertain -> held/lost`, plus release; uncertainty closes admission immediately; proof must come from the original lease session; loss is irreversible and exits 70; initial acquisition timeout exits 2. | Core 04, platform lease adapter | ES-01A, ES-02, ES-10 | SELECTED | BLOCKED | NOT_STARTED | Owner anchors: ES-01A and ES-02; verification/selectors: ES-05. |
| BC-011 | EXT-AC-152 | Define checked/saturating local deadline calculation, inherited deadline minimum, `now >= deadline` expiry, commit/cancellation/timeout precedence, equal-deadline tiebreak, and zero-grace behavior. | Core 04, Core 01 | ES-01A, ES-02, ES-10, ES-11 | SELECTED | BLOCKED | NOT_STARTED | Owner anchors: ES-01A and ES-02; verification/selectors: ES-05. |
| BC-012 | EXT-AC-153 | Bound participants to `1..16384`; bound per-participant and aggregate input to 64 MiB; bound aggregate prepare results to 64 MiB; stop at the first invalid validator in participant order; sample cancellation around every step and invocation. | Core 01, Extensions | ES-01A, ES-02, ES-04, ES-11 | SELECTED | BLOCKED | NOT_STARTED | Owner anchors: ES-01A, ES-02, and ES-04; verification/selectors: ES-05. |
| BC-013 | EXT-AC-154 | Define every allocated staged-object default; define the exact expiry/retry eligibility predicate and ordering; abandon upload failures immediately; prohibit holding a database transaction during deletion; serialize sweeps and coalesce missed intervals. | Core 01 object storage | ES-01A, ES-02, ES-11 | SELECTED | BLOCKED | NOT_STARTED | Owner anchors: ES-01A and ES-02; verification/selectors: ES-05. |
| BC-014 | EXT-AC-155 | Restore v1 only into a stopped empty target; process groups numerically and bindings sequentially; validate before advancing; failed targets never serve; no inactive profile code; rebuild derived state only after successful claim. | Core 01 backup, profile owners | ES-01A, ES-02 through ES-04, ES-11 | SELECTED | BLOCKED | NOT_STARTED | Owner anchors: ES-01A through ES-04; verification/selectors: ES-05. |
| BC-015 | EXT-AC-156 | Add `client_build_class='standard'`; require one support row for every claimable profile with workspaces; require Network Flow major 2 and `network_analysis`; linearize generation reservation and epoch rollover. | Core 01, Core 03, web | ES-01A, ES-02, ES-12 | SELECTED | BLOCKED | NOT_STARTED | Owner anchors: ES-01A and ES-02; verification/selectors: ES-05. |
| BC-016 | EXT-AC-157 | Prohibit capability facts and all nonempty capability arrays in v1; use `extension_capability_not_supported`; retain empty wire arrays for future compatibility; require a later capability contract before activation. | Core 00, Extensions | ES-01A, ES-02, ES-06, ES-09, ES-12 | SELECTED | BLOCKED | NOT_STARTED | Owner anchors: ES-01A and ES-02; registry bindings: ES-06; verification/selectors: ES-05. |
| BC-017 | EXT-AC-158 | Add fatal `published_component_lost` for unexpected termination of a publication-plan listener, dequeue gate, or worker; distinguish individual operation failures; close readiness/admission, drain, preserve committed state, and exit 70; no in-process restart. | Core 04, Core 01 jobs/runtime | ES-01A, ES-02, ES-10, ES-11 | SELECTED | BLOCKED | NOT_STARTED | Owner anchors: ES-01A and ES-02; verification/selectors: ES-05. |

ES-01A through ES-05 MUST replace each future-binding statement with exact owner
requirement anchors, verification IDs, and selector-family identifiers. The adopted
closure MUST contain no `TODO`, “appropriate,” “as needed,” “at least,” or
implementer-selected fallback for any BC rule.

### Companion-owner amendment plan

| Owner | Required amendment or parity decision | Boundary closures | Ownership guardrail | Gate dependencies | Completion evidence | Status |
| --- | --- | --- | --- | --- | --- | --- |
| Core 00 | Adopt the shared subsystem, manifest association, current majors, `network_flow_activity@2`, and `network_flow_activity -> import@1`. | BC-005, BC-016 | Recognition, claimability, retirement, and current major remain Core 00 only. | EXT-GATE-001, 013, 017 | Adopted Core revision, owner manifest, generated parity | TODO |
| Core 01 | Adopt strict seven-member producer/tolerant decoder, Base reservation registry, availability/no-store member, dispatch precedence, bounded transactions, staged objects, jobs, backup/participant interfaces, final commit, and recovery. | BC-003, BC-011 through BC-015, BC-017 | Public routes, envelopes, errors, transaction shell, object storage, and physical backup orchestration remain Core 01. | EXT-GATE-002, 003, 017, 022, 025, 026 | OpenAPI/owner anchors, route parity, participant and failure-injection evidence | TODO |
| Core 02 | Adopt or confirm extension-resource, authoritative/derived family, state-presence exclusion, and no cross-owner authoritative-write boundaries. | BC-001 | No implicit Core record/view/saved-view promotion. | EXT-GATE-004 | Owner manifest and Core-record boundary tests | TODO |
| Core 03 | Adopt epoch/generation, stable client/WebSocket identity, eligibility intersection, lazy loading, fallback, unsupported-major and authorization-loss consequences. | BC-015 | Browser/workbook behavior remains Core 03/web owned. | EXT-GATE-005, 017, 021 | Web owner rows, browser/stateful evidence, support-registry parity | TODO |
| Core 04 | Adopt inactive syntax-only processing, all deadlines, process lease, Stage 6, readiness degradation, diagnostics, fatal lifecycle, and exit codes. | BC-002, BC-009 through BC-011, BC-017 | Deployment config, authorization, readiness, and process exit remain Core 04. | EXT-GATE-006, 015, 023 | Config/process/security owner rows | TODO |
| Network Flow Activity | Publish owner manifest/fragments; remove competing discovery; adopt major 2; declare dependency, state, initialization, migrations, bindings/codecs, jobs, participants, rebuilds, and blocker. | BC-001, BC-009, BC-014, BC-015 | Network Flow resources/routes/algorithms remain `module.networkflow`. | EXT-GATE-007, 016, 017, 024 | Full `module.networkflow` and `web.networkflow` evidence | TODO |
| Incident Portability | Import closed export/import results, scoped staged-output capability, participant context/result/finding, logical refs, declarative inactive blocker, and invocation matrix when the interface changes. | BC-003, BC-012 | Public incident-bundle workflow remains Core/Incident Bundle owned. | EXT-GATE-018 | Amended Core anchors or exact no-change parity accounting | TODO |
| Physical backup owner | Import binding, codec, ordering, restore-target, and invocation contracts when changed. | BC-014 | Physical backup orchestration remains platform/Core owned. | EXT-GATE-003, 016, 018 | Codec vectors, restore tests, or exact no-change parity | TODO |
| Reporting | Import generic descriptor/claim/compatibility/state-presence/participant lifecycle only if its interface changes. | BC-003, BC-012 | Snapshot, render, release, and report model remain Reporting owned. | EXT-GATE-008, 018 | Amended owner anchors or exact no-change parity | TODO |
| Report Composition | Import only the generic interfaces it actually consumes. | BC-003, BC-012 | Authoring, validation, preview, and composition schema remain Composition owned. | EXT-GATE-008, 018 | Amended owner anchors or exact no-change parity | TODO |
| OpenTelemetry | Derive `cartulary.profile.claims` from canonical resolved claim set and digest; prohibit profile secrets/content. | none | Telemetry signal shape remains OpenTelemetry owned. | EXT-GATE-010 | OTel conformance/golden corpus and privacy tests | TODO |
| Domain vocabulary | Add adopted extension terms and remove stale discovery/unclaim/migration/multiprocess/client-support language after owners adopt. | BC-001 through BC-017 (terminology only) | Domain document remains terminology reference, never behavior owner. | EXT-GATE-011 | Markdown/domain review | TODO |
| Testing Harness v2 | Add authored owner contracts/rows/profiles/topology; amend NLSpec only if a new public command, runner, schema family, or execution-profile contract is needed. | BC-006, BC-008 | Harness owns mechanics only and cannot derive product behavior from the draft. | EXT-GATE-009, 020, 027, 028 | Harness contract, full-owner shards, exact-root audit | TODO |

### Section 27 artifact and generation plan

Authored owner manifests, fragments, schemas, contracts, catalog rows, and topology
inputs are the only hand-authored sources. All canonical registries, descriptors,
digests, indexes, embedded constants, TypeScript projections, task surfaces, schedules,
and topology render outputs are generated through repository-owned generators.

| Artifact or schema family | Authored owner/input | Generated or runtime consumers | Dependencies and validation | Status |
| --- | --- | --- | --- | --- |
| `cartulary.extension_dependency_declaration_set.v1` | Extensions specification owner; dependency owners supply manifest versions and digests | Exact input to dependency snapshot generation | Arrays always present; null invalid; complete owner/version/digest vectors; BC-004 | TODO |
| `cartulary.extension_dependency_snapshot.v1` | Adopted dependency identities/versions/digests | Extensions generator and adoption accounting | Owner manifests; canonical/digest/drift vectors | TODO |
| `cartulary.extension_owner_contract_manifest.v1` | One manifest per dependency owner document | Locator validator, snapshot, integrity, runtime admission | Exact document digest/anchors/fragments; symlink-safe path tests | TODO |
| `cartulary.extension_owner_fragment.v1` | Contributing named owner documents | Owner-input registry and descriptor derivation | Adopted fragment IDs/paths/digests only | TODO |
| `cartulary.extension_owner_input_registry.v1` | Generator input from validated manifests/fragments | Descriptor and registry generation | Fact identity, duplicate/order/collision vectors | TODO |
| `cartulary.extension_owner_fact_identity.v1` | Extensions schema and derivation vectors | Generator ordering/collision/accounting | EXT-REQ-205 vectors | TODO |
| Profile configuration contract/view | Extensions shape; named profile content; Core 04 view | Claim resolution and diagnostics | Inactive policy, bounds, secret refs, omission/null tests; BC-009 | TODO |
| Inactive-value schema family | Profile configuration owners under the closed Extensions vocabulary | Inert validation of unclaimed `syntax_only` values | Non-null schema reference; structural vocabulary only; no defaulting, view, resolution, retention, or egress; BC-009 | TODO |
| Descriptor-source schema | Extensions schema; ephemeral construction only | Generator in-memory normalization | Sole source for each member; missing/multiple scalar and duplicate-set rejection; never persist, package, hash, log, or drift-check instances; BC-005 | TODO |
| Descriptor schema and per-profile descriptors | Generated from owner input, including `recognized_profile.primary_owner_contract_ref` | Discovery, bindings, client support, accounting | One per recognized profile; exact source mapping; canonical/digest/parity; BC-005 | TODO |
| Profile registry schema and canonical registry | Extensions generator | Runtime admission, discovery input, build package | Zero-profile and current-profile vectors; root digest | TODO |
| Registry integrity object | Extensions generator/build packaging | Runtime admission and static accounting | Exact artifact/static-support identity/digest sets | TODO |
| Base route reservation registry | Core 01 authored route ownership, generated registry | HTTP dispatch overlap validator | Packaged handler parity and no extension capture | TODO |
| Client support registry and client asset-set manifest | Web build inputs and generated asset manifest | Browser eligibility intersection | `client_build_class='standard'`; one row per claimable profile with workspaces; descriptor major/workspace/capability/schema and asset digest parity; BC-015, BC-016 | TODO |
| Workspace availability | Core 01 plus Extensions result | Workbook startup and web controller | No-store response, epoch/generation, auth intersection | TODO |
| Publication plan and six component schemas | Extensions/Core 04 | Application startup Stage 6 coordinator | Atomic serving, deadlines, lease/readiness, `published_component_lost`, drain, and fatal tests; BC-010, BC-011, BC-017 | TODO |
| Implementation binding schema and packaged bindings | Build plus profile implementations | Runtime admission | Descriptor/contribution/state/job/participant parity | TODO |
| Admission validation/context/result schemas | Extensions | Profile preflight/post-migration calls | Closed scoped context, result precedence, timeout tests | TODO |
| Migration context/apply/validation/final-state schemas | Extensions shape; profile definitions | Migration coordinator | Pending-state validation, exact-once final validation, bounds | TODO |
| State-presence manifest/digest/vectors | Each versioned profile plus generator | Migration, backup, portability, accounting | `empty_state_policy`, authoritative/derived classification, presence exclusion, and all metadata/state combinations; BC-001 | TODO |
| State-initialization definition/context/result | Extensions shape; each versioned profile | Fresh-state coordinator | Every initialization default, empty/algorithm variants, and final validation; BC-001 | TODO |
| Physical state binding | Build plus durable profile owner | Backup/restore and integrity | Storage kind, state family, codec/rebuild parity, stopped-empty restore target and numeric/sequential ordering; BC-014 | TODO |
| Backup binding codec and vectors | Build, backup owner, durable profile | Physical backup/restore | Framing, limits, order, empty/historical/unsupported cases | TODO |
| State metadata and migration-ledger logical schemas | Extensions | PostgreSQL/profile state coordinator | Authored migrations, immutability, replay, restore | TODO |
| Job-kind contracts | Extensions shape; profile content | Core jobs and profile workers | Proof/cancel/idempotency/resource-result parity | TODO |
| Job commit proof and cancellation observation | Core 01 jobs plus Extensions | Reconciliation/replay | Proof precedence, contradiction, bounds | TODO |
| Transaction participant contract/context/result/finding | Core 01 plus Extensions | Cross-owner coordinator and participants | Participant count `1..16384`; 64 MiB per-participant/aggregate input and aggregate-result limits; ordered validation, cancellation sampling, deadlines, conflict/commit matrices; BC-011, BC-012 | TODO |
| Participant specialization and portability/reporting/backup contexts/results | Applicable shared and profile owners | Incident Bundle, Reporting, Composition, backup | Closed invocation matrix, transaction binding limits, and no implicit participation; BC-003, BC-012, BC-014 | TODO |
| Operation-specific portability export/import results | Extensions plus Incident Portability | Side-effect-free import preparation, shared transaction participation, and export publication | Absent/malformed/incompatible payloads; 64 MiB ceilings; scoped staged-output capability; rollback/commit/indeterminate outcomes; BC-003 | TODO |
| State blocking predicate | Extensions plus applicable profile | Inactive portability checks | Declarative state only; no inactive profile code | TODO |
| Staged-object logical schema | Core 01 object storage plus Extensions | Transaction publication and cleanup | Every allocated default; exact expiry/retry predicate/order; immediate upload abandonment; transaction-free deletion; serialized/coalesced sweeps; BC-013 | TODO |
| `cartulary.extension_validation_surface_declaration.v1` | Each schema or procedural-algorithm owner | Complete input to validation-condition registry generation | Schema condition annotations; closed procedural decision tables; emitted-condition completeness; BC-006 | TODO |
| Validation-condition registry | Generator plus every validation owner | Startup/runtime diagnostics and accounting | One row per reachable invalid condition; unregistered emitted conditions fail; safe formatter parity; BC-002, BC-006 | TODO |
| Startup finding schema | Core 04 plus Extensions | Startup diagnostics/readiness | Closed message/details/path/order/overflow | TODO |
| Fatal-condition and process-lifecycle registries | Core 04 plus Core 01 jobs/runtime | Readiness, admission, drain, process supervision, and exit handling | Closed lease and component-loss states; operation-failure distinction; no in-process restart; exit `2`/`70`; BC-010, BC-017 | TODO |
| Per-profile contract closure catalog | Extensions generator | Conformance manifest/accounting | Complete baseline and owner obligations; no owner reduction | TODO |
| Conformance manifest and index | Named profile owners plus generator | Static adoption accounting | One per claimable profile; unclaimable omission | TODO |
| Registry accounting object | Extensions generator | Adoption gate only | Named deterministic predicates/current digests; no run result as input | TODO |
| `cartulary.extension_traceability_mapping_source.v1` | Extensions specification owner | Exact input to acceptance/verification and document/clause mappings | Closed kinds, parent kinds/scope, zero-based ordinals, half-open byte ranges, clause-ID digest input; BC-008 | TODO |
| Clause traceability object | Specification owner/document tooling | Static adoption accounting | Exact source digest, all clauses/requirements/criteria/verifications, exact acceptance-to-selector mappings; BC-008 | TODO |
| Canonicalization and normative-source-lint vectors | Extensions/document tooling | Generator/linter tests | Exact Markdown subset and boundary/overflow vectors | TODO |
| `extension_safe_logical_ref_v1` and vectors | Extensions | Diagnostics, indexes, findings | Grammar, secrecy, traversal and overflow tests | TODO |
| Verification registry/owner contracts | Harness verification owners | Catalog validator and evidence accounting | Add both immutable `module.extensions` IDs atomically | TODO |
| Test-owner registry/family manifest | Harness test owners | Owner-slice planner/scheduler | Active nonempty exact-selector rows | TODO |
| Runner registry and runtime/resource/fixture profiles | Testing Harness v2 | Runner adapters/scheduler | Reuse existing runner contracts; add profiles only as authored topology inputs | TODO |
| Slice plan and scheduler summary | Harness-generated per invocation | Owner execution/finalization | Full owner versus subset; semantic digests | TODO |
| Evidence accounting and owner summary | Harness target finalizers | Evidence auditor/adoption gate | Paired shard for every owner/target partition | TODO |
| Evidence-root manifest and audit summary | Evidence-audit caller/auditor | Final adoption gate | Explicit compatible roots only; no newest/historical fallback | TODO |
| Generated Go/TS/schema/topology/task projections | Generators from the authored inputs above | Backend, web, Make, scheduler | `make generate-drift`, artifact policy, JSON shape, harness contract | TODO |

### Harness v2 onboarding plan

The Extensions owner must be added atomically to
`contracts/verification/owners/module.extensions.json`,
`contracts/verification/registry.json`, `tools/test_catalog_owner.json`, and
`tools/test_families/module.extensions.json`, with authored execution-topology/profile
inputs and generator-owned projections. The two immutable product verification IDs are:

- `module.extensions.verification.behavior_contract` for shared Base Profile
  extension runtime postconditions;
- `module.extensions.verification.contract_accounting` for owner input, generated
  contracts, registry integrity, closure, diagnostics, traceability, and static routing.

Imported Core, web, platform, application, and named-profile behavior remains routed
through those primary owners' verification IDs; `module.extensions` must not duplicate
it. The initial exact selectors below are selected planned identities. If an
adopted owner later reallocates a postcondition, change the owning row and traceability
together; do not duplicate the selector.

| Planned owner/family | Verification ID | Exact selector | Profiles | Evidence class | Required role |
| --- | --- | --- | --- | --- | --- |
| `module.extensions.unit` | behavior contract | Go package `./internal/modules/extensions`; tests `TestExtensionsClaimResolutionContract`, `TestExtensionsDependencyAndCollisionContract`, `TestExtensionsBindingAdmissionContract` | runtime/resource/fixture `none` | unit | Default owner row |
| `module.extensions.integration` | behavior contract | Go package `./internal/modules/extensions`; tests `TestExtensionsDiscoveryAndDispatchContract_Integration`, `TestExtensionsStateInitializationAndMigrationContract_Integration`, `TestExtensionsTransactionAndJobContract_Integration`, `TestExtensionsBackupPortabilityReportingContract_Integration` | `default` / `go_transaction_heavy` / `postgres_transaction` | integration | Service-backed full-owner row |
| `module.extensions.process` | behavior contract | Go package `./internal/app/serverprocess`; tests `TestExtensionsApplicationLeaseAndFatalLifecycle_Process`, `TestExtensionsAtomicPublication_Process`, `TestExtensionsStagedObjectExpiryCleanup_Process` | authored unclaimed and claimed profiles / `go_process` / service stack | process | Process/evidence-class gate |
| `module.extensions.static` | contract accounting | Go package `./internal/modules/extensions`; tests `TestExtensionsGeneratedArtifactIntegrity`, `TestExtensionsValidationConditionRegistry`, `TestExtensionsContractClosureAndAccounting`, `TestExtensionsClauseTraceabilityAndLimits` | `none` / `none` / `none` | static | Default owner row |
| `web.application.extensions` | existing/new primary web verification | Vitest file `apps/web/src/app/extensions/extensionLifecycle.test.tsx`; title `Verify extension discovery, client support, authorization, availability generation, fallback, and Base state preservation.` | `none` / `none` / `none` | unit | Affected web full-owner row |
| `web.application.extensions_browser` | existing/new primary web verification | Playwright file `apps/web/e2e/extensions.spec.ts`, project `chromium`, stage `webserver_backed`; title `Extensions eligibility uses discovery, client support, authorization, and current availability generation.` | claimed authored runtime / `browser_exclusive` / `service_stack` | browser | Browser evidence gate |
| `web.application.extensions_stateful` | existing/new primary web verification | Playwright file `apps/web/e2e/extensions.stateful.spec.ts`, project `chromium`, stage `stateful`; title `Extension authorization loss and generation rollover dispose extension state while preserving Base state and client identity.` | claimed authored runtime / `browser_exclusive` / `service_stack` | browser | Stateful evidence gate |
| `module.networkflow` and `web.networkflow` | existing Network Flow IDs plus any owner-amended ID | Existing exact selectors updated for generic seven-member discovery and major 2; add exact title `Network Flow major 2 uses only generic extension discovery and reserved dispatch.` if no existing selector proves it | `network_flow_claimed` and `default` as applicable | unit/integration/browser | Affected full-owner closure |
| Harness generated-artifact owner | Harness-owned verification | Existing shell command IDs for generation drift, artifact policy, JSON shape, and catalog/topology checks | `none` | static | Evidence-class gate, not product behavior |

Exact row IDs are generated from these owner/family/selector identities by the existing
catalog tooling; the generated row ID is not authored as product identity. A family
name, filename, target exit, title prefix, aggregate `make check`, or support-only row
cannot substitute for the exact selectors. Full adoption requires compatible
successful full-owner slices, every required evidence-class target, paired
`cartulary.test_evidence_accounting.v1` and `cartulary.test_owner_summary.v1` shards,
one terminal record per row, and one exact-root evidence audit.

### Compact requirement traceability ledger

Every requirement ID appears exactly once in the section-grouped rows below. The
grouping references IDs rather than reproducing normative prose. Primary ownership is
the draft allocation to be confirmed by companion adoption; current adopted ownership
continues until then.

| Requirement IDs | Workstream | Primary owner | Affected artifacts/code | Dependencies | Validation posture | Completion state |
| --- | --- | --- | --- | --- | --- | --- | --- |
| EXT-REQ-001, 002, 003, 004, 005, 174 | WF-00, WF-02, WF-02A | Extensions spec plus imported owners | Dependency declarations/snapshot/manifests and source posture; BC-004 | Core 00-04 and named owners | Source/anchor/dependency validation | TODO |
| EXT-REQ-006 through 014 | WF-00, WF-02A, WF-09 | Extensions spec/document tooling | Normative source and identifier traceability; BC-008 | Exact draft bytes and Harness IDs | Linter, uniqueness, bidirectional mapping | TODO |
| EXT-REQ-015 through 019 | WF-00, WF-04 | Extensions spec/Core boundaries | Runtime non-goals and subsystem facade | Owner allocation | Negative architecture/security tests | TODO |
| EXT-REQ-020, 021, 022, 023, 175, 203 | WF-02A, WF-03 | Extensions generator | Scalars, locators, manifests, safe refs; BC-004 | Owner document digests | Grammar, traversal, symlink, digest vectors | TODO |
| EXT-REQ-024, 025, 026, 027, 176, 177, 204, 205, 206 | WF-02, WF-02A, WF-03 | Named owners plus Extensions generator | Owner fragments/input/fact identities; BC-005, BC-006 | Adopted manifests | Determinism, omission, identity/collision tests | TODO |
| EXT-REQ-028 through 033 | WF-02, WF-02A | Core 00 | Recognized profile facts and adoption state; BC-005, BC-016 | Companion owner revisions | Core/registry/discovery parity | TODO |
| EXT-REQ-034 through 040, 178, 209 | WF-02A, WF-03 | Extensions generator/profile owners | Descriptor-source/configuration/descriptor schemas; BC-005, BC-009 | Owner input registry | Closed-shape/default/ephemeral/digest tests | TODO |
| EXT-REQ-041 through 046, 179, 180 | WF-02A, WF-03 | Extensions generator/build | Registry, dependency snapshot, canonical JSON, integrity and package roots; BC-004, BC-005 | Descriptors/bindings/static support | Canonical, bound, integrity, zero-profile vectors | TODO |
| EXT-REQ-047 through 053, 184, 207, 208, 213, 214, 215 | WF-02, WF-02A, WF-04 | Core 04 plus Extensions | Claim config/view, lease, publication, deadlines; BC-009 through BC-011, BC-017 | Registry, bindings, platform config | Config/process/readiness/fatal tests | TODO |
| EXT-REQ-054 through 058, 187 | WF-04 | Extensions coordinator | Claim/dependency/admission algorithm | Config, descriptors, bindings | Order, preflight, no-side-effect, no-listener tests | TODO |
| EXT-REQ-059 through 064, 185 | WF-04 | Extensions coordinator/profile owners | Runtime dependency graph/probes | Resolved claims and owner contracts | Cycle/order/probe/timeout tests | TODO |
| EXT-REQ-065 through 067, 181, 182 | WF-02A, WF-03, WF-04 | Build plus Extensions | Implementation binding and parity; BC-012, BC-016 | Registry/integrity/profile implementation | Missing/extra/mismatch admission tests | TODO |
| EXT-REQ-068 through 072, 210 | WF-03 | Extensions generator/Core 01 | Collision and Base reservation registries | Canonical facts/routes/dependencies | Every collision/multiplicity/route-overlap case | TODO |
| EXT-REQ-073 through 079, 196 | WF-02A, WF-03, WF-04, WF-07 | Extensions/Core 03/profile owners | Compatibility matrices, capability prohibition, and support registry; BC-015, BC-016 | Descriptor/binding/state/schema versions | Matrix and unsupported-value tests | TODO |
| EXT-REQ-080 through 086, 194, 195, 231 | WF-02, WF-02A, WF-05 | Core 01 | OpenAPI discovery producer/decoder; BC-005, BC-016 | Core 00 facts and registry | Strict seven-member producer/tolerant decoder/parity | TODO |
| EXT-REQ-087 through 094 | WF-02, WF-02A, WF-03, WF-08 | Named owners plus generator | Contribution registry/participant bindings; BC-003, BC-007, BC-012 | Adopted fragments/descriptors | Closed-kind/duplicate/parity tests | TODO |
| EXT-REQ-095 through 099 | WF-05 | Core 01/platform HTTP | Reservation and dispatch | Base/extension registries | Exact precedence and claimed/unclaimed outcomes | TODO |
| EXT-REQ-100 through 109, 197, 201, 211, 212, 220 | WF-02A, WF-07 | Core 03/web application | Client support, availability, workspace lifecycle; BC-015, BC-016 | Discovery/auth/browser assets | Unit/browser/stateful selectors and Base preservation | TODO |
| EXT-REQ-110 through 114, 200, 226 | WF-02A, WF-06 | Core 02/profile state owners/Extensions | Resource/state ownership, closure catalog; BC-001, BC-007 | Profile descriptors/state families | No-promotion/no-cross-write/closure tests | TODO |
| EXT-REQ-115 through 119, 192, 219 | WF-02A, WF-06 | Core 01 transaction owner | Participant protocol and staged publication; BC-011 through BC-013 | Jobs/storage/owner participants | Ordered failure/cancel/deadline/commit tests | TODO |
| EXT-REQ-120 through 129, 188, 189, 190, 216, 217, 234 | WF-02A, WF-06 | Extensions plus profile state owners | State metadata, initialization, migration, ledger; BC-001 | Bindings/state presence/locks | Fresh/current/migrated/restored and resumability tests | TODO |
| EXT-REQ-130 through 135, 191, 193, 218 | WF-02A, WF-06 | Core 01 jobs plus profile owners | Job contracts/proof/reconciliation/failure isolation; BC-017 | Resolved claim and transaction results | Proof/cancel/replay/fatal tests | TODO |
| EXT-REQ-136 through 145, 198, 199, 221, 222, 223, 232, 235 | WF-02A, WF-06, WF-08 | Backup, Incident Portability, Reporting, profile owners | Presence/bindings/codecs/participant specializations; BC-003, BC-012, BC-014 | State metadata and owner amendments/parity | Backup/restore/portability/reporting matrices | TODO |
| EXT-REQ-146 through 153 | WF-02, WF-02A, WF-04 | Core 04 plus profile security owners | Config, secret refs, authorization, egress; BC-009 | Owner configuration contracts | Syntax-only/egress/secret-negative/security tests | TODO |
| EXT-REQ-154 through 158 | WF-02, WF-04 | OpenTelemetry/audit owners | Claim-set telemetry and audit fields | Canonical resolved claims | OTel conformance/privacy and audit tests | TODO |
| EXT-REQ-159 through 163, 186, 224, 225, 233 | WF-02A, WF-03, WF-04 | Extensions generator plus Core 04 | Validation registry, startup findings, messages, paths, formatters, exits; BC-002, BC-006, BC-011, BC-017 | All validation owners | Exact precedence/path/order/overflow/fatal tests | TODO |
| EXT-REQ-164 through 167, 183, 202, 227, 228, 229, 230, 236 | WF-02A, WF-03, WF-09 | Extensions spec/generator plus Harness v2 | Section 27 artifacts, accounting, traceability, selectors/evidence; BC-004, BC-006 through BC-008 | All owners and generated inputs | Static accounting, limits, full-owner shards, evidence audit | TODO |
| EXT-REQ-168 through 171 | WF-10 | Coordinated document owners | Adoption statuses and companion revisions | All gates/evidence | Atomic promotion audit | TODO |
| EXT-REQ-172, 173 | WF-00 | Future owner | No current executable package surface | Future NLSpec | Negative upload/installation/execution tests | DEFERRED |

### Acceptance-criterion traceability ledger

The first eleven rows inventory the 141 acceptance criteria currently present in the
draft. The final row reserves 17 planned criteria. The BC ledger supplies their exact
one-to-one allocation; they become normative criteria only when ES-01A adds them to
the draft. After ES-01A, the fixed acceptance count is 158.

| Acceptance IDs | Workstream | Primary owner set | Affected artifacts | Dependencies | Validation posture | Completion state |
| --- | --- | --- | --- | --- | --- | --- |
| EXT-AC-001 through 014 | WF-00, WF-03, WF-04 | Extensions/Core 00/Core 04 | Owner inputs, descriptors, registry, claims, bindings | Adopted owners/config | Static plus startup unit/process | TODO |
| EXT-AC-015 through 023 | WF-03, WF-04, WF-05 | Extensions/Core 00/Core 01 | Compatibility, dependency graph, collisions/reservations | Registry and bindings | Exact matrix/collision/route tests | TODO |
| EXT-AC-024 through 033 | WF-05 | Core 01 and Network Flow | Discovery and dispatch | Seven-member contract and major action | HTTP integration/browser support | TODO |
| EXT-AC-034 through 040 | WF-07 | Core 03/web application | Workspace, Base fallback, identity, lazy load | Support/availability registries | Unit/browser/stateful | TODO |
| EXT-AC-041 through 050 | WF-06 | Extensions/profile state owners/Core 00 | Unclaim/reclaim/retirement/migration | State metadata and ledger | Service-backed migration/process | TODO |
| EXT-AC-051 through 060 | WF-06 | Core 01 jobs/transactions plus profile/Core 02 | Jobs, commit proof, participants, ownership | Typed contracts | Failure-injection integration | TODO |
| EXT-AC-061 through 071 | WF-06, WF-08 | Backup/Portability/Reporting/Core 04/OTel | Codec, state, participants, security, telemetry | Owner amendments/parity | Restore/matrix/security/OTel | TODO |
| EXT-AC-072 through 084 | WF-02, WF-03, WF-09 | All owners/generator/Harness | Drift, manifests, limits, diagnostics, accounting | Current digests and catalog | Static drift/shape/limit/diagnostic | TODO |
| EXT-AC-085 through 096 | WF-04, WF-06, WF-07, WF-08 | Core 04/Extensions/shared owners/web | Preflight, state, jobs, objects, discovery, browser outcomes | Publication and participant contracts | Process/service-backed/browser | TODO |
| EXT-AC-097 through 110 | WF-03, WF-04, WF-05, WF-07, WF-09 | Spec/generator/Core 01/03/04/Network Flow | Traceability, owner facts, config, lease, client support | Manifests and exact selectors | Static/process/browser/full-owner | TODO |
| EXT-AC-111 through 121 | WF-03, WF-06, WF-08 | Core 01/Extensions/profile/shared owners | Transactions, staged objects, migrations, jobs, codecs, closure | State/participant contracts | Failure injection and canonical vectors | TODO |
| EXT-AC-122 through 141 | WF-03, WF-04, WF-07, WF-09, WF-10 | Generator/Harness/Core 04/web/all affected owners | Accounting, selectors, lint, artifacts, fatal/browser/evidence | All current digests and retained roots | Static/full-owner/evidence audit/docs | TODO |
| EXT-AC-142 through 158 (planned) | WF-02A, WF-03 through WF-09 | Owners identified by BC-001 through BC-017 | Boundary schemas, algorithms, tables, registries, and runtime matrices | ES-01A owner anchors; ES-05 verification IDs and exact selector families | One planned criterion per BC row; boundary scenario inventory; primary-owner executable evidence | BLOCKED |

### Adoption-gate traceability ledger

| Gate IDs | Workstream | Primary owner set | Affected artifacts | BC prerequisites | Dependencies | Validation posture | Completion state |
| --- | --- | --- | --- | --- | --- | --- | --- |
| EXT-GATE-001 through 006 | WF-02A, WF-04, WF-05, WF-06, WF-07 | Core 00 through Core 04 | Core revisions, registry facts, discovery, transactions, client lifecycle, config/process | BC-001 through BC-003, BC-005, BC-009 through BC-017 | Characterization and boundary-closed draft retained as draft | Core owner evidence and parity | TODO |
| EXT-GATE-007 through 011 | WF-02A, WF-07, WF-08 | Network Flow, Reporting, Composition, Harness, OTel, domain | Owner manifests/fragments, major action, participants, telemetry, vocabulary | BC-001, BC-003, BC-008, BC-014, BC-015 | Core companion amendments | Affected full-owner/parity evidence | TODO |
| EXT-GATE-012 through 016 | WF-02A, WF-03, WF-04, WF-06, WF-09 | Generator/build/Core 04/state owners | Complete artifacts, manifests, integrity, diagnostics, state/migration/codec | BC-001, BC-004 through BC-007, BC-009, BC-013, BC-014 | All authored owners | Generation/drift/limit/process/service-backed | TODO |
| EXT-GATE-017 through 022 | WF-02A, WF-03, WF-05, WF-07, WF-08, WF-09 | Core 01, Network Flow, shared owners, web, generator | Generic discovery, participants, closure, traceability, client support, Base reservations | BC-003, BC-005, BC-007, BC-008, BC-015, BC-016 | Companion adoption | Static parity plus exact owner rows | TODO |
| EXT-GATE-023 through 026 | WF-02A, WF-04, WF-06, WF-09 | Core 04/Core 01/Extensions | Lease/fatal, initialization/migration, transaction, staged-object evidence | BC-001, BC-010 through BC-014, BC-017 | Implemented failure injection | Exact v2 selectors and terminal row records | TODO |
| EXT-GATE-027, EXT-GATE-028 | WF-02A, WF-09, WF-10 | Harness v2/document tooling/all owners | Full-owner shards, evidence audit, source lint, acceptance continuity, limit selectors | BC-001 through BC-017 | Every prior gate and all 158 acceptance criteria | Explicit-root audit and documentation/static gates | TODO |

## 7. Proposed Refactor Slice Plan

Every slice below is future work. A slice may be prepared on a branch but must not
mark the Extensions NLSpec or a dependent owner adopted/current until Slice ES-14.
Every code, test, contract, schema, migration, configuration, harness, or owner-document
slice requires later authorization. Rollback means reverting only that slice to its
last passing checkpoint while retaining earlier adopted-input work; committed profile
migration steps are never rolled back or reinterpreted by a down migration.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| ES-00 | none | Freeze current route, envelopes, reservations, claim/config, workbook, telemetry, Network Flow, state, jobs, and accounting behavior. `requires later authorization` | Existing owner tests only | Missing characterization can hide drift. | Exact tests named in Section 4 before movement. | `make task-guide ROLE=module-author OWNER=<current-owner>` then narrow owner slices | Remove only new characterization tests if invalid; no production change. | Every critical freeze row has passing owner-aligned evidence or `BLOCKED: missing characterization`. |
| ES-01 | ES-00 | Revise the draft while retaining `status: draft`; close source lint, IDs, anchors, imports, and manifest model. `requires later authorization` | Draft and document tooling | Draft changes could invalidate all locators/digests. | Normative-source golden vectors and traceability extraction. | `make lint-markdown` plus future normative-source target | Revert the draft revision; preserve current owners. | Exact draft digest is accepted and contains no open delegation or required placeholder. |
| ES-01A | ES-01 | Apply the exact BC-001 through BC-017 rules to the draft; amend existing owning requirements; add EXT-AC-142 through EXT-AC-158; update Section 27 artifacts, clause traceability, canonical limits, and gate mappings; retain `status: draft`. `requires later authorization` | Draft and document tooling | A rule could remain split, contradictory, unbounded, or allocated to the wrong owner. | Every BC boundary vector and source-lint/traceability check. | `make lint-markdown` plus future normative-source and traceability targets | Revert the complete boundary-closure edit; preserve ES-01 source foundations. | Every BC row has exact draft anchors, one acceptance criterion, complete omission/bound/error behavior, and no conflicting normative sentence. |
| ES-02 | ES-01A | Amend Core 00, 01, 03, and 04 together; Core 02 confirms/adopts BC-001; allocate Core anchors for BC-001 through BC-003, BC-005, and BC-009 through BC-017. `requires later authorization` | Core owner documents | Accidental ownership transfer or partial generic discovery. | Owner-section traceability and current contract characterization. | Documentation/owner contract validation targets | Revert the complete companion-doc set together. | Core manifests/anchors are digest-bound; every allocated BC has an exact Core anchor; draft still not adopted. |
| ES-03 | ES-02 | Amend Network Flow and state-owning profile owners for BC-001, BC-009, BC-014, and BC-015; require Network Flow major 2; declare dependencies/state/init/migrations/bindings/codecs/jobs/participants/rebuilds. `requires later authorization` | Network Flow owner and profile owner inputs | Major/version mismatch; competing discovery remains. | Existing Network Flow full-owner rows plus planned major-2 selectors. | `make test-slice OWNER=module.networkflow`; affected browser slices | Revert owner inputs as a unit; retain current major 1 behavior. | No competing discovery item remains in proposed owner set; every allocated BC has an exact profile-owner anchor; no runtime change yet. |
| ES-04 | ES-02, ES-03 | Amend Incident Portability, physical backup, Reporting, and Composition for BC-003, BC-006, BC-012, and BC-014 where imported interfaces change; otherwise emit exact no-change parity. `requires later authorization` | Core shared sections and named NLSpecs | Generic interface could usurp shared-owner behavior. | Participant matrix and no-participation/no-change assertions. | Affected owner task guides/slices | Revert each owner amendment and its parity row together. | Every allocated BC resolves to an amended shared-owner anchor or exact digest-bound no-change parity result. |
| ES-05 | ES-02, ES-03, ES-04 | Register Harness v2 contracts/owner/families/selectors/topology, including exact selectors for EXT-AC-142 through EXT-AC-158, and revise OTel; amend Harness NLSpec only for a new public mechanic. `requires later authorization` | Verification registry, test-owner/family manifests, topology profiles, OTel owner | Zero-row owner, duplicate ownership, unregistered runner/profile, hard-coded OTel set. | Exact planned selectors, every BC scenario family, and OTel golden/privacy tests. | `make harness-contract`; owner task guide; `make otel-conformance` | Revert all `module.extensions` authored rows/profiles and OTel change together. | Both IDs resolve, family is nonempty, every planned BC criterion has a primary-owner verification ID and exact selector family, profiles validate, and OTel consumes canonical claim identity. |
| ES-06 | ES-01A, ES-05 | Author owner manifests/fragments and every Section 27 schema/input, including dependency declarations, validation surfaces, traceability mappings, inactive-value schemas, and operation-specific portability results; extend generator and generation manifests. `requires later authorization` | `contracts/**`, generator code, authored schemas/manifests | Hand-editing outputs, stale locators, incomplete condition inventories, inconsistent byte limits. | Canonicalization, locator, identity, declaration, validation-condition, traceability, bounds, zero-profile and overflow vectors. | `make generate`; `make generate-drift`; JSON shape and artifact policy | Revert authored inputs/generator; regenerate old outputs. | Every required authored input is owner-bound and generator consumes no prose search or implementation-derived fact. |
| ES-07 | ES-06 | Generate dependency snapshot, owner input, descriptors, registry/integrity, validation-condition registry, closure/manifests/accounting/traceability, bindings, state, codec, client, portability, and Base reservation projections. `requires later authorization` | Generated roots and packaged assets via generators only | Partial or stale generated set; phase-shaped alias survives. | Artifact identity/digest/parity and package-root tests, including BC-004 through BC-008. | `make generate-drift`; `make harness-contract` | Regenerate from last passing authored input commit; never edit outputs. | Full artifact set is byte-stable, current, packaged, and contains no phase/v1 compatibility alias. |
| ES-08 | ES-00, ES-02, ES-06, ES-07 | Introduce a deep Extensions coordination facade for immutable registry queries, claim resolution, bindings, dependencies, admission, and publication-plan output. `requires later authorization` | `internal/modules/extensions`, app/server ports, platform adapters | Facade could absorb Core/platform/profile ownership. | Unit selectors for claims/dependencies/collisions/bindings. | `make test-slice OWNER=module.extensions` | Keep current profile structs/route path behind adapter until parity passes. | Callers depend on narrow DTOs/ports; coordinator owns no profile behavior or transport/storage adapter. |
| ES-09 | ES-08 | Switch Core 01 discovery/reservation/dispatch and application assembly to generated registry facts and atomic Stage 6 publication. `requires later authorization` | HTTP runtime, server assembly, OpenAPI, Extensions query facade | Public response major change and partial routing. | Discovery/dispatch integration, process atomic publication, Base reservation parity. | Module/app/platform owner slices and `make backend-process` | Feature-branch rollback to characterized three-field/current dispatch implementation; no mixed producer. | One generic producer/decoder, one reservation registry, no listener/worker before serving, and no competing Network Flow discovery. |
| ES-10 | ES-08, ES-09 | Implement the BC-002, BC-006, BC-009, BC-010, BC-011, and BC-017 runtime matrices: Core 04 config views, validation precedence, process lease, deadlines, readiness/fatal lifecycle, validation registry diagnostics, and safe errors. `requires later authorization` | Config, app server, Extensions validation, telemetry/log adapters | Secret disclosure, wrong exit, partial startup, lease split brain. | Config, process lease/fatal, path/message/details/overflow, secret-negative tests. | Owner slices; `make go-gosec-targeted`; OTel conformance | Disable new coordinator before serving and fall back only to the fully characterized current runtime on the branch. | Exit 2/70, lease/readiness/timeout, component-loss, and diagnostic matrices all close. |
| ES-11 | ES-08, ES-10 | Implement the BC-001, BC-003, BC-011, BC-012, BC-013, BC-014, and BC-017 runtime matrices: state presence, metadata/ledger, initialization/migration, jobs/proof, Core 01 transactions, staged objects, backup codecs, portability/reporting participation. `requires later authorization` | Extensions, Core platform, profile modules, authored migrations | Durable partial effects, cross-owner access, data loss, unsupported restore. | All fresh/current/migrated/restored, ordered failure, cleanup and matrix selectors. | Service-backed full-owner slices; `make migration-drift` | Stop before schema publication where possible; never down-migrate committed profile state. | Every state/transaction/job/storage/participant result is closed, resumable, and owner-correct. |
| ES-12 | ES-07, ES-09, ES-10 | Implement BC-015 and BC-016: move browser to the standard support-registry/availability/auth intersection, prohibit capability activation, add lazy loading, fallback, stable identity, and exact cleanup. `requires later authorization` | `apps/web`, generated TS/UI contracts, workbook startup | Base cache/draft/queue loss, stale UI render, or unsupported capability activation. | Planned Vitest/Playwright browser/stateful selectors; every nonempty capability rejection; existing a11y/visual allocations. | `make frontend-typecheck`; unit and browser targets | Restore current discovery consumer as one complete version; never dual-read. | Unsupported/stale/unauthorized extensions and nonempty capabilities cannot render; Base state survives every transition. |
| ES-13 | ES-09, ES-10, ES-11, ES-12 | Execute full-owner slices and all required evidence-class gates; finalize paired shards and exact-root audit. `requires later authorization` | Harness artifacts only | Selected subset or broad target falsely claimed as closure. | Every active exact selector once; no unmapped/skipped/unexpected result. | Full-owner `test-slice`, service-backed slices, browser gates, `test-evidence-audit` | Retain failed roots for diagnosis; do not promote. | Compatible successful roots close every affected owner/target/row partition. |
| ES-14 | ES-13 | Run static registry/traceability/drift accounting and promote the draft and all companions together. `requires later authorization` | Owner statuses/manifests and generated projections | Intermediate adopted/current state. | All 158 acceptance criteria and 28 gates. | `make check`, `make harness-contract`, release-required gates, exact-root audit | Revert the entire promotion-status commit if any post-promotion gate fails; preserve implementation/data. | All 158 criteria and all gates are `DONE`, all digests/evidence current, and no older competing contract remains current. |

### Mandatory atomic-adoption sequence

The following order is mandatory and maps directly to ES-00 through ES-14. It may be
implemented as an ordered change series, but no intermediate change may mark the
Extensions NLSpec adopted/current or expose partially coordinated behavior:

1. characterize current behavior;
2. revise source/lint foundations and retain `status: draft`;
3. apply BC-001 through BC-017 to the draft;
4. amend Core owners;
5. amend profile and shared owners, recording exact no-change parity only where the
   imported interface does not change;
6. register Harness/OpenTelemetry inputs and exact boundary selectors;
7. author contracts and generate every Section 27 extension artifact,
   clause-traceability object, and v2 projection;
8. implement runtime, participant, and browser behavior;
9. execute owner evidence, explicit-root audit, and static registry,
   clause-traceability, and drift accounting;
10. promote the Extensions NLSpec and every required companion revision atomically.

## 8. Validation Plan

Commands are public Make targets discovered through `make help`, `make help-all`, and
`make explain-target`. Direct `go`, `pnpm`, Vitest, Playwright, and raw scripts are not
conformance commands. Run `make task-guide` before owner slices and
`make agent-finalize` before broad end-of-run verification in the later authorized
implementation task. `make agent-finalize` is intentionally not run for this
tracker-only task because it may refresh harness-maintenance artifacts.

### Boundary acceptance scenario inventory

Each planned criterion MUST cover every applicable enum token and state transition,
plus minimum, maximum, maximum-plus-one, empty, omitted, and explicit-null inputs.
The cases below are the minimum executable inventory; owner decision tables MAY add
cases but MUST NOT remove or merge outcomes that have different normative effects.

| BC ID | Minimum executable scenarios for its planned acceptance criterion |
| --- | --- |
| BC-001 | Every metadata/state-presence combination under `allowed` and `forbidden`; empty initialization; metadata-only input; omitted and explicit-null policy. |
| BC-002 | Invocation failure, structural invalidity, overflow, remaining schema defect, valid findings, and valid empty result; counts 0, 256, 257, 4096, and 4097. |
| BC-003 | Absent, malformed, incompatible, minimum, 64 MiB, and 64 MiB-plus-one payloads; prepared import, rollback, committed publication, and indeterminate commit; scoped staged-output denial. |
| BC-004 | Present empty arrays; omitted arrays; explicit null; missing, duplicate, extra, and stale dependency declarations; manifest version/digest mismatch. |
| BC-005 | Exactly one scalar source; missing and multiple scalar sources; duplicate set members; stale owner reference; attempted prose/code inference. |
| BC-006 | Complete schema annotations and procedural decision tables; missing, duplicate, stale, extra, and unregistered emitted conditions. |
| BC-007 | Every subject/contribution-kind mapping; generated subject with a not-applicable reason; every permitted and unrecognized fixed-baseline reason. |
| BC-008 | Every clause/parent kind; root and child scope; ordinal zero and maximum-plus-one; empty and adjacent half-open ranges; invalid overlap/out-of-bounds; digest and document-clause mismatch. |
| BC-009 | Syntax-only present, omitted, defaulted, explicit-null, reference-shaped, and structurally invalid values; prohibited view creation, resolution, retention, and egress attempts. |
| BC-010 | Initial acquisition success/timeout; every lease transition; uncertainty recovery on the original session; different-session proof; session loss; detection timeout; release; forbidden reacquisition after loss. |
| BC-011 | Checked and saturated local deadlines; inherited earlier/equal/later deadlines; cancellation before, at, and after deadline; cancellation before, during, and after proven commit; equal-deadline tiebreak and zero grace. |
| BC-012 | Participant counts 1 and 16384; count 0 and 16385; per-participant and aggregate bytes at 0, 64 MiB, and 64 MiB-plus-one; aggregate results at the same boundary; first-invalid ordering; malformed result; cancellation around every step/invocation. |
| BC-013 | Every staged-object initial default; noncandidate, expiry candidate, and retry candidate; upload failure; each deletion result; transaction-held deletion attempt; sweep overlap, missed interval coalescing, startup failure, and serving-dependency failure. |
| BC-014 | Stopped-empty, running, and nonempty targets; numeric group order and sequential binding order; validation failure isolation; inactive restore attempt; failed target serving attempt; derived rebuild before and after successful claim. |
| BC-015 | Required standard support row; omitted API-only profile; Network Flow major/workspace mismatch; stale/current generation; concurrent reservation; epoch rollover; minimum and overflow generation values. |
| BC-016 | Empty wire arrays; every nonempty capability fact/array surface; capability activation attempt; exact `extension_capability_not_supported` outcome. |
| BC-017 | Individual operation failure versus termination of each publication-plan listener, dequeue gate, and worker; readiness/admission closure, drain completion/timeout, committed-state preservation, exit 70, and prohibited restart. |

For this tracker-only revision, validation is limited to `git diff --check`,
`make lint-markdown`, `make json-shape-check`, and
`make generated-artifact-policy-check`, plus structural searches for all BC and
planned acceptance IDs, their owner/slice/verification/completion references, and
stale completion claims. Generation, formatting, implementation suites, and
`make agent-finalize` are out of scope because no production, contract, generated,
or Harness input changes in this revision.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| Current owner discovery | `make task-guide ROLE=module-author OWNER=<owner-id>` | Select narrow authoritative rows for each existing owner. | yes | `OWNER=module.extensions` currently fails until ES-05. |
| Owner explanation | `make explain-test-owner OWNER=<owner-id>` | Inspect exact owner rows/targets/profiles. | yes | Current Extensions owner failure is recorded evidence. |
| Extensions unit/full owner | `make test-slice OWNER=module.extensions` | Full default-owner unit/static rows. | no | Required after ES-05/ES-08; selected `ROWS=` is focus evidence only. |
| Extensions service-backed | `make service-backed-test-slice OWNER=module.extensions` | Integration/process/state/transaction partitions. | no | Must finish as full-owner closure, not a subset. |
| Backend unit | `make backend-unit` | Pure Go behavior and exact Go selectors. | no | Characterization may use current-owner projections before Extensions onboarding. |
| Backend store/integration/process | `make backend-store`, `make backend-integration`, `make backend-process` | PostgreSQL, HTTP, process lease/publication/fatal boundaries. | no | Required according to row evidence classes. |
| Migration | `make migration-drift` | Authored SQL against scratch database. | no | Required for state metadata/ledger or profile schema changes. |
| Frontend types/unit | `make frontend-typecheck`, `make frontend-unit` | Generated types, app controller, workbook state. | no | Both required for browser client changes. |
| Frontend import boundary | `make frontend-import-boundary-check` | Web/package dependency ownership. | no | Prevent generated/vendor/platform leakage. |
| Backend module boundary | `make backend-module-boundary-check` | Module/platform/application imports. | no | Enforce coordinator ports and prevent peer-internal imports. |
| Browser ordinary | `make browser-e2e-webserver-backed` | Discovery/support/auth and claimed/unclaimed routes. | no | Broad browser target closes only its exact selected rows. |
| Browser stateful | `make browser-e2e-stateful` | Authorization loss, epoch/generation rollover, state preservation. | no | Required for Core 03 cleanup/identity behavior. |
| Browser accessibility/visual/measurement | `make browser-e2e-a11y`, `make browser-e2e-visual`, `make browser-e2e-measurement` | Only rows whose primary owners require these evidence classes. | no | Do not promote design/support evidence into Base/extension conformance. |
| OpenTelemetry | `make otel-conformance` | Canonical claim signal, privacy, golden corpus. | no | Required after OTel amendment. |
| Security | `make go-gosec-targeted`, `make go-vulncheck` | New coordinator, config, storage, secret and egress surfaces. | no | Gosec audit remains advisory unless owner gates say otherwise. |
| Generation | `make generate`, then `make generate-drift` | Authored inputs to generated Go/TS/topology projections. | no | Run generation only in an authorized implementation task. |
| Generated policy | `make generated-artifact-policy-check` | Markers, roots, and lint-scope protection. | yes | Also required for this tracker handoff. |
| JSON/bootstrap shapes | `make json-shape-check` | Contracts, registries, manifests, topology inputs. | yes | Also required for this tracker handoff. |
| Harness contracts | `make harness-contract` | Verification, owner, family, runner, selector, topology, artifact schemas. | no | Required after every ES-05/ES-06/ES-07/ES-09 change. |
| Evidence audit | `make test-evidence-audit OWNER=<owner-id> EVIDENCE_ROOTS_FILE=<path>` | Explicit compatible retained roots for one owner. | no | No absent roots, newest-run lookup, or historical fallback. |
| Documentation | `make lint-markdown` | Authored Markdown structure. | yes | Required for this tracker and every owner-document slice. |
| Narrow developer gate | `make test-fast` | Fast combined loop after coherent code slices. | no | Does not replace service/browser/evidence-class gates. |
| Full developer gate | `make check` | Default full correctness projection. | no | Run after `make agent-finalize`; broad success does not replace omitted evidence classes. |
| Full/release | `make test`, `make release-check` | Full corpus and release aggregation when promotion requires it. | no | Core 05 remains excluded unless separately activated by a claim owner. |

## 9. Top-Level Work Tracker

### Planning and implementation work

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| T-001 | Define target, authority, and exclusive write scope | WF-00 | DONE | none | Sections 1 and 2 | Target and non-goals are explicit. |
| T-002 | Inventory target and adjacent current state | WF-00 | DONE | T-001 | Sections 2 through 5 | All target files and material adjacent owners are accounted for. |
| T-003 | Map proposed requirements, criteria, gates, artifacts, and owners | WF-00 | DONE | T-002 | Section 6 ledgers | Every EXT ID and Section 27 artifact is mapped. |
| T-004 | Freeze current observable behavior | WF-01 | TODO | T-003 | Section 4 characterization matrix | Every critical boundary has pre-change owner evidence. |
| T-005 | Coordinate draft and companion owners | WF-02 | BLOCKED | T-004 | ES-01 through ES-05 | Requires later authorization and owner adoption decisions. |
| T-005A | Close selected boundary decisions in normative sources | WF-02A | BLOCKED | T-005 | BC-001 through BC-017; ES-01A through ES-05 | Requires later authorization; every BC has exact owner anchors, one active acceptance criterion, and exact primary-owner verification mappings. |
| T-006 | Build owner-manifest and canonical generation pipeline | WF-03 | BLOCKED | T-005A | Section 27 artifact set | Requires later authorization; generated outputs are drift-free. |
| T-007 | Implement runtime coordinator and atomic publication | WF-04 | BLOCKED | T-006 | Extensions facade/application ports | Requires later authorization; no partial serving. |
| T-008 | Adopt Core discovery/reservation/dispatch | WF-05 | BLOCKED | T-007 | OpenAPI/route parity | Requires later authorization and coordinated major action. |
| T-009 | Implement state/migration/jobs/transaction/storage coordination | WF-06 | BLOCKED | T-007 | State/participant artifacts and migrations | Requires later authorization; all matrices pass. |
| T-010 | Adopt browser support/lifecycle behavior | WF-07 | BLOCKED | T-006, T-008 | Support registry and browser rows | Requires later authorization; Base state preserved. |
| T-011 | Adopt Network Flow and shared-owner participation | WF-08 | BLOCKED | T-005A, T-009 | Owner amendments/parity | Requires later authorization; no ownership transfer. |
| T-012 | Onboard Harness v2 owner and evidence | WF-09 | BLOCKED | T-006, T-008, T-009, T-010, T-011 | Two verification IDs, owner/family/topology, paired shards | Requires later authorization and implemented exact selectors. |
| T-013 | Audit and atomically promote | WF-10 | BLOCKED | T-012 | All 28 gates and exact-root audit | Every binary criterion is `DONE`; companions promote together. |
| T-014 | Independently distributed executable extensions | future | DEFERRED | none | Draft Section 30 | A later adopted NLSpec explicitly authorizes the surface. |

### Gate closure tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| G-000 | Establish normative boundary prerequisites | WF-02A | BLOCKED | T-005A | Adopted owner anchors plus active EXT-AC-142 through EXT-AC-158 verification mappings | Every BC row has adopted owner anchors and active acceptance/verification mappings before G-001. |
| G-001 | Close EXT-GATE-001 through EXT-GATE-006 | WF-02A through WF-07 | BLOCKED | G-000 | Core owner revisions and owner evidence | Requires later authorization; all six Core gates pass. |
| G-002 | Close EXT-GATE-007 through EXT-GATE-011 | WF-02, WF-07, WF-08, WF-09 | BLOCKED | G-001 | Profile/shared/Harness/OTel/domain amendments or parity | Requires later authorization; all five companion gates pass. |
| G-003 | Close EXT-GATE-012 through EXT-GATE-016 | WF-03, WF-04, WF-06, WF-09 | BLOCKED | G-001, G-002 | Complete generated/runtime/state artifacts | Requires later authorization; all five artifact/runtime gates pass. |
| G-004 | Close EXT-GATE-017 through EXT-GATE-022 | WF-03, WF-05, WF-07, WF-08, WF-09 | BLOCKED | G-003 | Discovery/participant/closure/client/reservation parity | Requires later authorization; all six integration gates pass. |
| G-005 | Close EXT-GATE-023 through EXT-GATE-026 | WF-04, WF-06, WF-09 | BLOCKED | G-003 | Exact process/migration/transaction/cleanup row evidence | Requires later authorization; all four failure-injection gates pass. |
| G-006 | Close EXT-GATE-027 and EXT-GATE-028 | WF-09, WF-10 | BLOCKED | G-004, G-005 | Full-owner paired shards, audit, source lint, limit selectors | Requires later authorization; exact current roots close all partitions. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-19 | Codex planning session | Planning complete; draft remains non-authoritative and pre-existing edit preserved. | `AGENTS.md`, skill/reference, framework, draft, Core 00-04, domain and named owners; touched tracker only. | `git status`, `git rev-parse`, `sha256sum`, `sed`, `rg`, `awk`, `wc` | Authority, 236 requirements, 141 criteria, and 28 gates indexed. | Later authorization for every implementation/adoption slice. | Begin ES-00 characterization in a separately authorized task. |

### Backend boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-19 | Codex planning session | Current target is a three-file thin HTTP facade; profile/reservation/claim composition is adjacent. | Target files, `httpapi/extensions.go`, `httpapi.go`, server runtime/routes, Network Flow and shared-owner modules. | `find`, `rg`, `sed` | Boundary split and current call paths recorded. | No generic coordinator/state/ledger/binding implementation exists. | Add owner-aligned characterization before moving responsibilities. |

### Frontend boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-19 | Codex planning session | Web consumes current discovery and claim state; no support-registry/availability generation exists. | App shell client, App, debug harness, workbook startup/shell tests, browser support and Network Flow surfaces. | `rg`, `sed` | Core 03/web ownership and browser characterization plan recorded. | Target behavior requires coordinated client/server contract changes. | Implement only after support and availability artifacts generate. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-19 | Codex planning session | One phase-shaped authored extension index feeds current Go/TS projections. | Extension/OpenAPI contracts, family registry, contractgen validator, generated policy and generated consumers. | `rg`, `sed` | Full Section 27 authored/generated ownership plan recorded. | Owner manifests/fragments and generator do not exist. | ES-06 authors inputs/generator, then ES-07 generates outputs. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-19 | Codex planning session | Harness v2 is adopted/current; `module.extensions` is absent. | Verification registry/owners, test-owner registry, Network Flow family, runner registry, execution topology/profiles/schemas. | `make help`; `make help-all`; `make task-guide ROLE=module-author OWNER=module.extensions`; `make explain-test-owner OWNER=module.extensions`; `make explain-target` for required doc/policy targets | Help/explain-target commands succeeded; both Extensions owner commands failed with `unknown active test owner module.extensions`. | Owner/family/topology cannot close until ES-05 and executable selectors exist. | Add both verification IDs and nonempty exact-selector rows atomically. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-19 | Codex planning session | Current auth/session handling is in route adapter; config and OTel hard-code parts of profile handling. | Extensions route, platform auth/config/telemetry, Network Flow security/key-ring code, Core 04 and OTel owner. | `rg`, `sed` | Secret, egress, syntax-only inactive, lease/fatal and telemetry risks mapped. | Core 04/OTel companion amendments and implementation authorization. | Preserve current outcomes until ES-10 passes negative/security evidence. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-19 | Codex planning session | Tracker decision selection is complete; target-contract boundary closure remains blocked. | Tracker plus recorded inspected corpus and informative `temp/analysis-notes.md`. | Required validation commands and final worktree/hash audit. | Selected BC-001 through BC-017 recorded; no normative or production implementation performed. | Normative adoption, implementation, and evidence remain blocked on later authority. | Next authorized session starts at ES-00 and proceeds through ES-01A before generation or production edits. |

### Tracker validation

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-19 | Codex planning session | Final tracker validation | Tracker; retained summaries under `.cartulary/test-results` are ignored runtime evidence, not authored changes. | `git diff --check`; `make lint-markdown`; `make json-shape-check` | All passed. JSON-shape run root: `.cartulary/test-results/20260719T223232Z-p9`. | none | Preserve results in handoff. |
| 2026-07-19 | Codex planning session | Generated policy validation | Generated-artifact policy and retained target summary | `make generated-artifact-policy-check`; `make explain-run RESULTS_DIR=.cartulary/test-results/20260719T223232Z-p9 TARGET=generated-artifact-policy-check DETAIL=summary`; escalated rerun of `make generated-artifact-policy-check` | Initial sandboxed run failed as harness `unknown_failure` because `spawnSync git` returned `EPERM`; retained summary identified the sandbox cause. Escalated exact rerun passed at `.cartulary/test-results/20260719T223316Z-p21698`. | none | Cite the successful rerun and the diagnosed initial failure. |
| 2026-07-19 | Codex boundary-closure revision | Tracker-only boundary validation | Tracker only; no normative, contract, generated, Harness, or production input changed. | `git diff --check`; `make lint-markdown`; `make json-shape-check`; `make generated-artifact-policy-check`; structural BC/AC and stale-claim searches | All passed. JSON-shape root: `.cartulary/test-results/20260720T034056Z-p18996`; generated-policy root: `.cartulary/test-results/20260720T034110Z-p19402`. | Normative adoption remains blocked. | Skip generation, formatting, implementation suites, and `make agent-finalize` because this is a tracker-only revision. |

## 11. Open Questions and Blockers

Only blockers to safe implementation or adoption are listed. Absence of current
`module.extensions` rows or generic runtime code is established repository state, not
a discovery `TODO:`.

No design question remains open for BC-001 through BC-017. Their selected rules are
fixed by the canonical ledger. Remaining states are authority, implementation, and
evidence blockers; reopening a selected rule requires a separately authorized
decision change and corresponding tracker revision.

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | The target NLSpec is draft and none of its coordinated adoption gates may be treated as current behavior. | Implementing the target over adopted owners would create competing contracts. | Later authorized coordinated owner-document task and all gate evidence. | BLOCKED |
| RB-002 | Network Flow is currently adopted at document version 1.2.0/contract major 1 while BC-015 requires major 2. | Discovery and client compatibility cannot be changed piecemeal. | Core 00 status/major action plus Network Flow/Core 01/03/04 companion adoption and major-2 owner evidence. | BLOCKED |
| RB-003 | No `module.extensions` verification contract, owner row, family manifest, or topology profile exists. | Owner slices, paired shards, and evidence audit cannot close. | ES-05 Harness-authored inputs and implemented exact selectors. | BLOCKED |
| RB-004 | Generic extension owner manifests, generator, registry integrity, bindings, state metadata/ledger, codecs, closure, and accounting artifacts do not exist. | Runtime admission or adoption cannot be inferred from code/routes/database. | ES-06/ES-07 later authorized contract/generator work. | BLOCKED |
| RB-005 | Current code lacks the process lease, Stage 6 atomic publication, generic migration/final-validation coordinator, and complete transaction/staged-object protocol. | Failure cases could publish partial behavior or corrupt durable state. | ES-08 through ES-11 implementation plus exact process/service-backed evidence. | BLOCKED |
| RB-006 | No compatible successful retained roots exist for future Extensions rows. | Static registry success or broad target success cannot prove execution. | ES-13 full-owner/evidence-class runs and explicit evidence-root manifest. | BLOCKED |
| RB-007 | Every production, test, owner, contract, migration, config, generated, and harness change is outside this tracker-only task. | The planning write boundary must not be mistaken for implementation authority. | A later explicit implementation authorization. | BLOCKED |
| RB-008 | The 17 selected boundary rules are not yet present in adopted owner documents. | Generation or implementation against the pre-closure draft would preserve ambiguity and could create divergent contracts. | ES-01A through ES-04 owner amendments, ES-05 active verification mappings, and G-000 closure. | BLOCKED |

If companion review reveals two adopted primary owners claiming incompatible behavior,
add a new blocker with the exact text `BLOCKED: owner contradiction`, cite both exact
owner anchors/digests, and stop the affected workflow. No such contradiction was
established by this planning pass.

## 12. Binary Completion Criteria

Planning is complete only when every tracker criterion below is true; implementation
is complete only when every implementation/adoption criterion is `DONE`. Code existing
or a broad test target passing is never sufficient by itself.

### Tracker handoff criteria

- [x] Every file under `internal/modules/extensions`, including testsupport, is
  inventoried.
- [x] Current callers, assembly, platform dependencies, route registration, tests,
  frontend consumers, contracts, migrations, generated outputs, configuration, and
  Harness v2 mappings are represented by opened repository evidence.
- [x] Authored inputs, generated outputs, runtime state, retained evidence, and
  normative owner documents are separated.
- [x] Current and target boundaries cover discovery, claims, dispatch, generation,
  bindings, state, migrations, jobs, transactions, backup/restore, portability,
  reporting, browser, security, diagnostics, and conformance accounting.
- [x] Every current/adjoining responsibility has a keep/move/split/defer decision.
- [x] All current 236 `EXT-REQ-*` IDs, current 141 `EXT-AC-*` IDs, and 28
  `EXT-GATE-*` IDs are present in traceability ledgers with workstream, owner,
  artifacts, dependencies, validation, and state.
- [x] BC-001 through BC-017 occur exactly once as target-rule definitions in the
  canonical ledger, preserve `SELECTED`/`BLOCKED`/`NOT_STARTED` state, identify
  owners and slices, and map one-to-one to planned EXT-AC-142 through EXT-AC-158.
- [x] Every planned boundary criterion has a required owner-anchor, verification-ID,
  exact-selector, scenario-inventory, gate-prerequisite, and completion path through
  ES-01A through ES-05 and G-000.
- [x] Every Section 27 artifact/schema family has an authored owner, generated/runtime
  consumer, dependency, and validation posture.
- [x] Every workflow and slice has dependencies, validation, rollback, and an exact
  exit condition; behavior-changing slices say `requires later authorization`.
- [x] Public validation commands were discovered through Make; current owner-command
  failures are recorded rather than hidden.
- [x] No generated-file hand edit, phase-shaped runtime boundary, `EXT-FIX-*`, fixture
  result schema, v1 alias, compatibility reader, or historical evidence fallback is
  planned.

### Implementation and coordinated-adoption criteria

- [ ] ES-00 characterization closes every route, envelope, claim, reserved-route,
  startup/fatal, migration, transaction, browser, and accounting boundary.
- [ ] ES-01A incorporates every selected BC rule into the draft, amends existing
  requirements, creates EXT-AC-142 through EXT-AC-158, and leaves no conflicting,
  omitted, unbounded, or implementer-selected behavior.
- [ ] All required Core, profile, shared-owner, OpenTelemetry, domain, and Harness
  companion amendments or exact permitted no-change parity records are adopted.
- [ ] Every BC row has exact adopted owner requirement anchors plus active
  primary-owner verification IDs and exact selector families; G-000 is `DONE`.
- [ ] Every owner locator/manifest/fragment resolves against exact current digests;
  no required placeholder or open delegation remains.
- [ ] Every Section 27 authored input exists, including dependency declarations,
  validation-surface declarations, traceability mappings, inactive-value schemas,
  and operation-specific portability results; every generated artifact/projection
  regenerates byte-identically with no drift and generated roots were not hand-edited.
- [ ] Core 00 alone controls recognition, claimability, retirement, and current major;
  Core 01 alone owns public discovery/reservation/dispatch and shared protocols.
- [ ] The Extensions coordinator exposes a small facade and owns only generic
  validation/coordination; profile, transport, storage, browser, reporting,
  portability, backup, and telemetry behavior stays with primary owners.
- [ ] No route, workspace, job dequeue, WebSocket subscription, or readiness success is
  visible before atomic Stage 6 serving; admission exits `2` and fatal integrity exits
  `70` under the adopted Core 04 lifecycle.
- [ ] Every state-owning profile has exact state presence, initialization, binding,
  codec, migration, ledger, lock, rebuild, validation, and backup/restore parity.
- [ ] Every cross-owner transaction, job, staged-object, portability, reporting, and
  backup matrix passes exact failure/cancellation/boundary selectors with no partial
  effects or inactive profile-code execution.
- [ ] Browser eligibility is the exact discovery/support/authorization/availability
  intersection; stale responses cannot render; Base cache/request/queue/optimistic
  state/drafts and stable client/WebSocket identity survive required transitions.
- [ ] Security, egress, diagnostics, audit, telemetry, readiness, jobs, browser state,
  and WebSocket payloads contain no prohibited secret or incident content.
- [ ] Both `module.extensions` verification IDs, every affected primary-owner
  verification, every active exact selector, runner, and execution profile validate
  with no unmapped, duplicate, unauthorized-skipped, or unexpected executed case.
- [ ] Every affected owner has compatible successful full-owner evidence and every
  required evidence-class target has paired accounting/owner-summary shards.
- [ ] One exact-root Harness v2 evidence audit closes every required owner/target/row
  partition; subset, broad-target, stale-root, newest-run, and historical fallback are
  rejected.
- [ ] Static registry accounting, clause traceability, source lint, acceptance
  continuity, canonical limits, generated drift, owner evidence, documentation checks,
  all 158 acceptance criteria, and all `EXT-GATE-001` through `EXT-GATE-028` are
  `DONE`.
- [ ] The mandatory ten-step sequence completed and the Extensions NLSpec plus every
  required companion revision was promoted atomically; no intermediate artifact ever
  claimed the generic subsystem current.
- [ ] Core 05 remains outside implementation conformance unless a separately adopted,
  genuinely claim-bearing timed or fixture-sensitive publication boundary explicitly
  activates it.

Implementation requires a later authorized task.
