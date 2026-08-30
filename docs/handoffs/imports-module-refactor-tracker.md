# imports Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Field | Value |
| --- | --- |
| Target path | `internal/modules/imports` |
| Target label | `imports`, derived from the final path segment and already safe lowercase kebab case |
| Output path | `docs/handoffs/imports-module-refactor-tracker.md` |
| Repository baseline | Branch `main`; commit `8ee25792ff11032923d49f71b495f99f17f870b7`; clean worktree observed at `2026-08-30T09:51:22-04:00` |
| Current documentation posture | The staged user-owned tracker was evolved in place. The earlier clean-worktree observation is preserved as historical baseline evidence and is not a claim about the current worktree. |
| Status | Implementation remediation complete; WS-00 through WS-18 and RB-001 through RB-006 are closed |
| Allowed change in this task | Adopted owner specifications, authored contracts and harness inputs, Imports and source-owner implementation, tests, and this controlling tracker, strictly through WS-00 through WS-18 |
| Non-goals | New external product interfaces, database migrations, frontend redesign, grid-vendor changes, generic persistence facades, Network Flow mapping generalization, recovery redesign, or a replacement aggregate finalization subsystem |
| Current authority | The approved Imports Specification and Implementation Remediation plan authorizes the serial workstreams in Section 7. Generated roots, lockfiles, and generated topology remain generator-owned. |

### Normative language convention

This tracker uses `MUST`, `MUST NOT`, `SHOULD`, and `MAY` to make later work
unambiguous. Those words bind an implementation task only after that task has
the required authority. They do not promote this tracker, the planning
framework, `temp/analysis-notes.md`, or `docs/research/nlspec-spec.md` above an
adopted owner. The following states are distinct and MUST NOT be collapsed:

| Statement class | Meaning in this tracker | May authorize implementation? | Closure evidence |
| --- | --- | --- | --- |
| Adopted requirement | A requirement already owned by an adopted Core section, subsystem NLSpec, or adopted architecture decision | Yes, within that owner's exact scope and a later authorized task | Owner text plus conforming implementation and owner-routed evidence |
| Required owner amendment | Decision-complete replacement text or behavior that repairs an owner contradiction or incomplete owner contract | No; the affected owners MUST adopt it first | Atomically consistent adopted owners and downstream projections |
| Planned implementation obligation | Exact behavior, interface, ordering, failure, or compatibility condition for a later authorized slice | No in this documentation task | Passing characterization, implementation, static, drift, and owner evidence named by the slice |
| Supporting evidence | Repository code, tests, generated consumers, harness maps, guides, research, and prior handoffs | No | It may demonstrate current state or corroborate a decision but cannot create behavior |
| Closure criterion | A binary fact that MUST be true before an RB item or slice is reported complete | No by itself | Every listed fact passes; partial success is not closure |

The tracker is the controlling execution and handoff artifact for the approved
remediation. Adopted owners continue to govern their scopes. The approved plan
authorizes the owner amendments named below; each amendment and its downstream
implementation must remain atomically consistent before its workstream closes.

The planning source hierarchy is:

1. Adopted subsystem NLSpecs for their named subsystem only.
2. Core 00 through Core 04 for current implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. Domain vocabulary and implementation-support guides for terminology, package boundaries, harness mechanics, and execution support.
5. Current repository code and tests for current implementation state.
6. Prior plans, handoffs, and the planning framework as evidence only.

The local framework was read first and is used as planning doctrine, not as evidence that a path, symbol, test, command, or boundary exists. The live repository was then inspected before recording the findings below. Where adopted owners conflict, this tracker records `BLOCKED: owner contradiction` and does not select a side.

Owner documents inspected:

- `docs/spec/00_document_set_status_and_precedence.md`, including the adopted profile and owner allocation for Imports and Network Flow Activity.
- `docs/spec/01_architecture_storage_and_view_contracts.md`, especially the Import Extension Profile routes, dispatcher, registry, unit commit, resource, mapping, error, and owner-facade contracts.
- `docs/spec/02_domain_model_schema_and_history.md`, for import provenance and source-owner mutation consequences.
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`, for file import, clipboard tabular-ingest reuse, discovery, selection, mapping, unit atomicity, and apply workflow.
- `docs/spec/04_security_deployment_and_conformance.md`, for authentication, cookie CSRF, authorization, and import acceptance evidence.
- `docs/extension-subsystem-nlspec.md`, adopted/current version `0.10.0`, for import-target contribution and analytical facade admission.
- `docs/network-flow-activity-nlspec.md`, adopted/current version `5.0.0`, for Network Flow target-owned mapping, preview, apply, errors, and atomic publication.
- `docs/testing-harness-nlspec.md`, for owner-first Make commands, row selection, generated topology, retained evidence, and harness ownership.
- `docs/domain.md`, especially Imports and Tabular Ingest and the Imports-to-Network-Flow anti-corruption boundary.
- `docs/guides/cartulary-dev-guide.md` and `docs/guides/cartulary_implementation_testing_guide.md`, as implementation support only.
- Core 05 was checked for applicability and is not used: this planning task publishes no timed, fixture-sensitive, benchmark, or claim-bearing result.

Repository evidence inspected:

- Every Go file under `internal/modules/imports`, including `ownerfacade`, as inventoried in Section 2.
- `internal/modules/tabularingest/tabularingest.go` and its tests.
- `internal/app/importassembly/owner_registry.go`, server module settings and runtime assembly, `internal/app/extensionassembly/import_jobs.go`, and `internal/app/recoveryassembly/state_catalog.go`.
- Network Flow import facade, owner-error translation, module composition, and transaction participant files.
- Source-owner Imports facades in Timeline, Entities, Evidence, Indicators, Assessments, Parties, Tasks/Decisions, and Artifacts.
- Frontend import coordination in `apps/web/src/imports/importCoordinator.ts`, the Workbook Import Assistant, the Network Flow import controller, their tests, generated-contract adapters, UI selector, and `apps/web/e2e/import-assistant.spec.ts`.
- `contracts/imports/**`, the Imports OpenAPI owner input, Network Flow import binding and schemas, relevant view-schema and recovery inputs, generated Go and TypeScript consumers, `contracts/verification/owners/module.imports.json`, `tools/test_families/module.imports.json`, and `tools/backend_module_boundaries.json`.

Implementation is authorized only for the workstreams and compatibility choices
recorded in this tracker. Test rows and phase maps are verification accounting,
not runtime architecture.

### Source allocation

| Topic | Normative home | Supporting home | Rule for this tracker |
| --- | --- | --- | --- |
| Network Flow version, Extensions dependency, client major, and compatibility | Core 03, Network Flow Activity NLSpec, and affected Extensions owner facts | Authored fragments, descriptors, client-support inputs, generated consumers, and drift evidence | WS-01 adopts the major-5 bundle atomically; major-4 compatibility is intentionally unsupported. |
| Imports route semantics and CSRF classification | Core 01 for the route contract; Core 04 for security, precedence, errors, and no-side-effect conformance | Imports handlers and characterization/integration tests | The tracker defines the required owner amendment and later correction but does not itself change authorization behavior. |
| Shared view-target mapping semantics | Core 01 and Core 03; Core 04 for binary conformance and hostile-input evidence | Mapping/parser fixtures, byte vectors, owner-request fixtures, and provenance fixtures | Defaults, ordering, canonical bytes, and exclusions MUST appear in the owners before extraction. |
| Backend boundary scan mechanics | Testing Harness only if its public contract changes | Authored boundary-policy input and Make-owned evidence | No Core behavior change is required for RB-004. Generated topology MUST NOT be hand-edited. |
| Source-owner finalization topology | Core 00/Core 04 and the adopted Collaboration boundary decision | Caller/effect matrices, failure injection, package graphs, and handoff rationale | The decision is already adopted. No new decision or combined Revisions-and-Collaboration owner is required. |
| External research and standards | None unless an adopted owner restates the behavior | Rationale and corroboration only | OWASP, RFCs, vendor research, and `temp/analysis-notes.md` MUST NOT establish Cartulary behavior independently. |

## 2. Current-State Repository Inventory

The target exists and contains 36 Go files totaling 12,759 lines at the recorded baseline. Every file is in scope for inventory; no target file is omitted.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| ---- | ---------------------- | ------------------------------------------ | --------------- | --------------------- | ----------------- | ---------------------------------------- | ----------------------------- | ---------- | ----- |
| `internal/modules/imports/api.go` | Import request DTOs, strict JSON and multipart metadata decoders, approved mapping normalization, and fingerprints | `CreateSessionRequest`, `SourceColumnMapping`, `ApprovedMapping`, request types, decoder functions, `ImportSessionFileContentTypes` | Imports HTTP handlers; Network Flow and frontend contracts indirectly through HTTP | Platform HTTP API; JSON, hashing, and view/mapping validation helpers | `owner_errors_test.go`, integration tests, frontend coordinator tests through wire behavior | `contracts/imports/schemas.v1.json`, Imports OpenAPI, generated core HTTP types and validators | Imports | high | Any decoder or fingerprint drift changes replay, persisted mappings, or public errors. |
| `internal/modules/imports/apply_coordination.go` | Session apply sequencing, cancellation checks, per-unit transaction coordination, and terminal coordination | Private `Service` methods | `apply_jobs.go`, job handler path | Jobs, Postgres transactions, Imports store and owner dispatch | Apply/current-state, rollback, crash recovery, cancellation, and evidence parity integration tests | Imports unit-commit and job contracts | Imports | high | Coordinates multiple observable side effects but does not own target semantics. |
| `internal/modules/imports/apply_jobs.go` | Loads and applies frozen units, revalidates current state, records outcomes, and coordinates final success/failure | Private `Service` methods | `jobs.go`, apply coordination | Jobs, Postgres, Imports store, owner facade and finalizer | Apply, rollback, crash recovery, partial cancellation, Network Flow integration tests | Imports job/result and unit-commit contracts | Imports | high | Ordering and recovery behavior are contract-sensitive. |
| `internal/modules/imports/boundary_guard_test.go` | Static guards for generated target use, injected owner facades, peer-store and peer-table bans, and responsibility separation | Test surface only | `module.imports` unit harness row | Filesystem source inspection and Go testing | Self; `module.imports.unit.boundary_complete_52a2a59837` | `tools/test_families/module.imports.json`, backend boundary policy as supporting accounting | Imports test evidence | medium | Current evidence explicitly requires the split production files and rejects peer stores. |
| `internal/modules/imports/boundary_test.go` | Guards owner-facade use instead of workbook-row mutation shortcuts | Test surface only | `module.imports` unit harness row | Filesystem source inspection and Go testing | Self; boundary-complete row | Harness row only | Imports test evidence | medium | Evidence, not architecture authority. |
| `internal/modules/imports/characterization_test.go` | Characterizes member route parsing, generated target inventory, safe internal errors, and hidden-sheet discovery | Test surface only | `module.imports` characterization row | Imports private parsers and registry | Self; `module.imports.unit.characterization_complete_180f923ec3` | Generated target registry consumed by the test | Imports test evidence | high | Freezes current observable inventory but has no CSRF matrix. |
| `internal/modules/imports/discovery.go` | CSV and XLSX unit discovery, bounded previews, source dimensions, warnings, and initial unit materialization | Private `Service` and discovery types | Upload handler and discovery workflow | `tabularingest` CSV parser, XLSX parser, Imports store types | CSV/XLSX discovery and upload integration tests | Import resource, locator, warning, and preview contracts | Imports | high | Imports reuses `tabularingest.ParseTableWithMaxColumns`; the existing clipboard `TabularRowPlanV1` and its fingerprint are not automatically the file-import approved-mapping or persistence contract. |
| `internal/modules/imports/extension_facade.go` | Stable analytical target mapping/preview/apply facade and application override | `ExtensionImportApplyRequest`, mapping request/result, apply result, `ExtensionImportFacade`, facade key | Network Flow module and app/server composition | Context, JSON, Postgres transaction capability, Imports source stream | Network Flow integration and contract tests | Network Flow import binding and owner schemas; Imports analytical binding contract | Imports published analytical seam | high | Core owns the binding slots; Network Flow owns exact target payload semantics. |
| `internal/modules/imports/http_handlers.go` | Binds 11 OpenAPI operations, parses paths, authenticates, authorizes, decodes, dispatches, and writes responses | Package route behavior through `RegisterRoutes`; no new exported DTO family | Server route assembly; generated OpenAPI operation binder | Platform HTTP API/auth/pagination, Incidents admission, Imports store and application methods | Characterization and all HTTP integration tests; frontend and E2E consumers | Imports OpenAPI owner input, generated HTTP bindings/types/validators | Imports transport-adjacent adapter | critical | Six operations are state-changing for cookie CSRF: create, mapping approval, select, skip, regions, and apply. Four GET operations and mapping preview are read-only. The live member handler currently omits state-changing mode for mapping, select, skip, and apply. |
| `internal/modules/imports/imports_integration_test.go` | Service-backed end-to-end Imports behavior, replay, storage, selection, regions, owner dispatch, atomicity, recovery, Network Flow, and cancellation | Test surface only | 15 active Imports integration rows | Test server/app support, Postgres, object store, owner modules | Self | Verification and test-family manifests | Imports integration evidence | high | Existing helpers send valid CSRF; negative member-route CSRF coverage was not found. |
| `internal/modules/imports/job_finalization.go` | Typed success, failure, and cancellation finalizer port for common Jobs | Finalization structs and `JobSuccessFinalizer` | Imports job orchestration; extension assembly adapter | Jobs, Postgres transaction function | Discovery/apply integration and recovery tests through behavior | Common Jobs and Imports job contracts | Imports-to-Jobs port | high | Preserves final commit and result publication boundaries. |
| `internal/modules/imports/jobs.go` | Registers and executes `imports.discovery` and `imports.apply` handlers and job progress transitions | Private handler registration and execution | `Service.RegisterRoutes`; Jobs runner | Platform Jobs, JSON payloads, Imports finalizer/store | Upload, apply, cancellation, crash recovery integration tests | Job handler identity and result/resource-ref contracts | Imports | high | Handler identities are observable through durable jobs and recovery. |
| `internal/modules/imports/limits.go` | Imports resource limits and archive-policy alias | `Limits`, `ArchiveLimits` | Server settings and Imports parser/discovery | Platform archive policy | XLSX security tests and server module settings tests | Import limit and configuration projections | Imports | medium | Limits are injected from composition; no runtime default should move into target owners. |
| `internal/modules/imports/mapping.go` | Validates and prepares view or analytical mappings, invokes owner preview, and enforces claim/facade availability | Private `Service` methods | Mapping and mapping-preview handlers | Generated target registry, view schema, extension facade, Imports store/source stream | Mapping integration, owner error, Network Flow preview/apply tests | Import target registry, view schemas, Network Flow binding/schemas | Imports orchestration | critical | Must retain server-derived source context and safe target-owner translation. |
| `internal/modules/imports/owner_apply.go` | Creates one unit change set, maps rows into source-owner requests, dispatches owner facades, and records journal/result facts | Private apply helpers | Apply coordination | Revisions append port, ownerfacade registry, Imports store and source rows | Timeline/entity/evidence parity, rollback and crash recovery integration tests; boundary guards | View-target and owner-facade contracts | Imports mutation coordinator | critical | Legitimate coordination; must not acquire peer source or projection semantics. |
| `internal/modules/imports/owner_errors.go` | Closed analytical owner error type, validation, safe translation, and canonical token handling | `ExtensionImportOwnerError`, `ExtensionImportErrorTranslation` | Network Flow facade, mapping/apply paths | Imports API/errors and schema checks | `owner_errors_test.go`, Network Flow safe-error integration and contract tests | Network Flow owner-error schema and translation ID | Imports translation boundary | high | Unknown target tokens fail closed and must never echo unsafe detail. |
| `internal/modules/imports/owner_errors_test.go` | Characterizes canonical `write_null` fingerprinting and safe owner-error translation | Test surface only | Imports unit harness rows | Imports decoders and error translation | Self | Test-family row; mapping/error contracts consumed | Imports test evidence | medium | Does not establish shared clipboard/file mapping parity. |
| `internal/modules/imports/owner_registry.go` | Constructs the exact owner-create registry from contributed facades | `NewOwnerCreateRegistry` | App import assembly | Generated target registry and `ownerfacade` | Import assembly registry tests and Imports boundary tests | Import target registry and owner-facade catalog | Imports composition seam | medium | Application assembly supplies implementations; Imports validates completeness. |
| `internal/modules/imports/recovery_inventory.go` | Declares Imports vNext recovery object inventory | `VNextRecoveryObjectInventory` | Recovery assembly | Recovery contracts and Imports table inventory | Recovery catalog tests indirectly | Recovery inventory projections | Imports source-state owner | medium | Inventory describes import-owned durable state only. |
| `internal/modules/imports/recovery_state.go` | Declares Imports recovery-state contribution | `RecoveryStateContribution` | Recovery assembly | Recovery-state contract | Recovery assembly/catalog tests indirectly | Recovery state catalog projections | Imports source-state owner | medium | Does not transfer restore orchestration into Imports. |
| `internal/modules/imports/regions.go` | Validates operator-selected rectangles, reparses the original XLSX source, creates/replays a durable region unit | Region parameter/result types and `ErrInvalidSourceRect` | Region HTTP operation | XLSX parser, hashing, Imports store | Operator-region Go integration, frontend coordinator/assistant, browser E2E | Region request and import-unit contracts | Imports | high | Must preserve containment, limits, exact replay, and source identity. |
| `internal/modules/imports/revision_append_port.go` | Narrows the Revisions appender to change-set creation | Private port and adapter | Owner apply path | Revisions, Postgres | Atomicity and owner apply integration tests | Revisions contract indirectly | Imports coordination port | low | Appropriate dependency inversion; no generic transaction facade is introduced. |
| `internal/modules/imports/selection.go` | Selects, skips, reselects, rejects overlap, and preserves approved mappings | Private route/store helpers | Select/skip handlers and apply admission | Imports store and geometry/state checks | Selection lifecycle integration, frontend assistant tests, E2E | Import-unit selection/status contracts | Imports | high | Mapping retention and deterministic selected-unit order are observable. |
| `internal/modules/imports/service.go` | Public Imports composition facade, dependency options, route registrar, and service wiring | `Service`, `RouteOption`, `WithJobs`, `WithOwnerCreateRegistry`, `WithRevisionAppender`, `WithLimits`, `WithExtensionProfileAdmission`, `WithJobSuccessFinalizer`, `RegisterRoutes` | Server runtime assembly and tests | HTTP API, auth keys, store, Jobs, revisions, ownerfacade | Server composition, module settings, integration tests | OpenAPI owner binding and module composition contracts | Imports application facade | high | Legitimate thin public edge over a non-thin internal package. |
| `internal/modules/imports/source_streams.go` | Issues and opens opaque unit-scoped read-only source capabilities | `ImportSourceCapability`, `ImportSourceStream` | Imports mapping/apply and Network Flow facade | Imports store/source bytes, hashing and bounded readers | Network Flow preview/apply and source-changed tests | Analytical facade source capability contract | Imports | high | Filesystem paths, object-store keys, and bearer credentials must not cross the facade. |
| `internal/modules/imports/store.go` | Direct SQL for import-owned sessions, units, streams, plans, outcomes, journals, replay, and apply admission | `Store`, state errors, session/unit/apply data types, `NewStore`, `SourceRect` | Imports handlers and application orchestration | Platform Postgres, Jobs transaction ports, Incidents admission, Revisions appender | Most integration tests and peer-table boundary guard | Imports migration/schema, resource and persistence contracts | Imports persistence adapter | critical | No peer owner table SQL was found; imported resource semantics remain owner-dispatched. |
| `internal/modules/imports/targets.go` | Converts generated target rows into runtime lookup maps and availability decisions | Private runtime target surface | Mapping, owner registry, apply path | `internal/gen/importtargetregistry` | Characterization and boundary tests; frontend generated-target tests indirectly | `contracts/imports/index.json`, `view-targets.v1.json`, generated Go/TS registries | Imports | high | Generated registry is correctly consumed; do not replace it with a handwritten registry. |
| `internal/modules/imports/unit_outcomes.go` | Freezes unit plans, revalidates source/mapping/registry digests, persists outcomes, and derives session terminal state | Private store and outcome types/methods | Apply orchestration and recovery | Postgres, JSON/hashing, Imports store | Apply current-state, rollback, crash recovery, cancellation tests | Unit-commit, terminal result, and recovery contracts | Imports | critical | Protects against replaying owner effects after an ambiguous worker outcome. |
| `internal/modules/imports/xlsx.go` | Bounded hostile ZIP/XML parsing for sheets, tables, named ranges, used ranges, formula cache diagnostics, and source cells | Private parser/index types | Discovery and operator regions | ZIP, XML, archive policy, Imports limits | `xlsx_test.go`, characterization and XLSX integration tests | XLSX locator, warning, preview, and resource-limit contracts | Imports source adapter | critical | Macros, formulas, external links, dynamic ranges, and archive abuse must remain inert or rejected. |
| `internal/modules/imports/xlsx_test.go` | Characterizes XLSX locators, formula diagnostics, archive abuse, and mapping-dependent formula blockers | Test surface only | Imports unit harness security row | XLSX parser and approved mapping types | Self; `module.imports.unit.xlsx_locator_security_and_formula_index_41d0b577ac` | XLSX import contracts consumed | Imports test evidence | high | Security-sensitive evidence. |
| `internal/modules/imports/ownerfacade/characterization_test.go` | Characterizes empty-value normalization, nullability, and safe Party match-conflict details | Test surface only | Ownerfacade unit harness row | Ownerfacade scalar/error implementation | Self | Owner-create contract evidence | Imports published-language evidence | medium | Covers owner-neutral input shape, not source-owner business policy. |
| Retired in WS-17: `internal/modules/imports/ownerfacade/finalize.go` | No live responsibility; the combined historical/live operation, generic whole-row fact derivation, row/version helpers, and aggregate ports were deleted | None | None | None | Negative source-owner boundary and affected owner slices | Core 00 REQ-00-075 and Core 04 AC-564 | No replacement owner; source owners coordinate separate consumer ports | closed | Imports session/job finalization remains independently owned in `job_finalization.go`. |
| `internal/modules/imports/ownerfacade/owner_create.go` | Closed scalar union, field values, provenance, owner request/result, safe validation errors, and default normalization | Public owner-create types, scalar constructors, indexing, normalization, safe-error functions | All source-owner import facades and app import assembly | View schema and field normalization, UUID/time | Ownerfacade tests plus source-owner unit/integration tests | Approved view mapping and owner-create facade contracts | Imports published language | critical | This is a cross-module API even though it is internal to the Go module. |
| `internal/modules/imports/ownerfacade/registry.go` | Owner facade binding, validation/normalization wrapper, exact registry, and contiguous mutation sequencing | Public facade/command/binding/registry types and constructors | App import assembly and source-owner contributors | Postgres transaction, owner-create types | `registry_test.go`, import assembly tests, Imports boundary/integration tests | Generated target/facade registry contracts | Imports published language | critical | Exact binding validation keeps source-owner implementations out of Imports discovery. |
| `internal/modules/imports/ownerfacade/registry_test.go` | Characterizes registry completeness, wrong-target rejection, field validation, and sequencing | Test surface only | Imports ownerfacade unit row | Ownerfacade registry | Self | Owner-facade harness row | Imports test evidence | medium | Preserves the published cross-owner seam. |
| `internal/modules/imports/ownerfacade/scalar_contract_test.go` | Characterizes the scalar closed union and invalid field structures | Test surface only | Imports ownerfacade unit row | Ownerfacade scalar types | Self | Owner-create nullability/scalar row | Imports test evidence | medium | Guards DTO shape, not durable target semantics. |

## 3. Module Boundary Diagnosis

The directory name is not the ownership proof. Adopted Imports owner sections and the observed dependencies support a permanent Imports boundary, while the live package still contains several internal responsibility classes that should remain explicit.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| -------------------- | ---------------- | ----------------------- | --------------------------- | -------- | ----- |
| CSV/XLSX source compatibility, hostile parsing, locators, warnings, and regions | `discovery.go`, `xlsx.go`, `regions.go` | Imports | keep | Core 01/Core 03, domain §13.12, live parser tests | Worksheet terms remain source locators, never runtime workbook identity. |
| Import session/unit resources, mapping/source integrity, selection, provenance, and opaque streams | Imports root package and store | Imports | keep | Core 01 Imports contract and live SQL/resource implementation | Legitimate deep module complexity. |
| HTTP route binding, authentication, role checks, decoding, and response mapping | `http_handlers.go` | Imports transport-adjacent adapter over Imports application behavior | split | Live handler combines transport and orchestration; route surface is Imports-owned | Split means internal delegation, not a route or envelope change. |
| Discovery/apply Jobs and unit-finalizer coordination | `jobs.go`, apply files, finalizer port | Imports with common Jobs ports | keep | Core 01/Common Jobs contracts and live extensionassembly adapter | Job attempts remain common Jobs state; Imports owns its handler semantics. |
| View-target row semantics, defaults, revision details, projections, and live publication | Source-owner modules behind `ownerfacade` | Timeline, Entities, Evidence, Indicators, Assessments, Parties, Tasks/Decisions, Artifacts | keep behind owner facade | Generated registry, app import assembly, peer-store boundary tests | Imports coordinates one unit but must not absorb target business policy. |
| Network Flow mapping, row validation, table/row storage, diagnostics, and atomic publication | `internal/modules/networkflow` behind `ExtensionImportFacade` | Network Flow Activity | keep behind analytical facade | Network Flow NLSpec §10 and live facade/transaction participant | Imports owns upload and dispatch only. |
| Shared mapping/parser kernel for clipboard and file import | `internal/modules/tabularingest`; partial reuse from `discovery.go`; separate Imports mapping in `api.go` | Imports and clipboard owners consuming one stable Tabular Ingest contract | split | Core 03 shared-engine requirement and live implementation difference | Exact shared overlap and fingerprint parity require characterization. |
| Combined source-owner record finalization | Removed in WS-17 | No combined permanent owner; each source owner coordinates separate Revisions, Projections, and Collaboration ports | removed | Core 00 REQ-00-075, Core 04 AC-564, per-owner cutovers, negative boundary, and broad fast suite | No alias or replacement aggregate exists. The separate Imports session/job finalizer remains Imports-owned. |
| Direct Imports persistence | `store.go`, `unit_outcomes.go` and receiver-specific files | Imports persistence adapter | keep, then isolate behind private application ports | SQL inspection and peer-table guard | Current direct SQL is import-owned; no generic transaction facade is proposed. |
| Frontend import workflow and controller state | `apps/web/src/imports`, Workbook Import Assistant, Network Flow controller | Web application and Network Flow web owner | keep outside backend target | Live TypeScript imports and frontend tests | Frontend is contract consumer, not a backend Imports subpackage. |
| Grid-vendor integration | `packages/grid-adapter`; Network Flow presentation consumes its public API | Grid adapter | intentional/no_action | No `react-data-grid` import under Imports or its frontend coordinator/assistant | No grid refactor is supported by this target. |
| Recovery inventory/state classification | Imports recovery contribution files | Imports source-state owner; Recovery orchestrates | keep | Recovery assembly composition | Does not make Imports a restore coordinator. |

Diagnosis: the target is a **legitimate application/service facade**, **transport-adjacent adapter**, **persistence-adjacent adapter**, **mutation coordinator**, and **mixed-responsibility package**. It is not an accidental catch-all in the sense of owning unrelated durable semantics, and it is not a frontend shell or grid-vendor layer. It coordinates view/projection consequences through target owners but is not itself a projection owner.

Planning findings relative to the generic framework:

- The live package is already decomposed by responsibility and has static boundary tests; a wholesale catch-all extraction is unsupported.
- The implementation-support guide describes Imports and shared tabular ingest together, but the live repository has a separate `internal/modules/tabularingest` package and Imports only reuses its CSV parser.
- `internal/modules/tabularingest.TabularRowPlanV1` includes clipboard identity and its own mapping fingerprint. It is current implementation evidence, not proof that file imports should adopt those complete bytes or persist clipboard identities.
- The adopted Collaboration boundary already decides the permanent topology for source mutation consequences: source owners derive private Revisions facts and public Collaboration effects independently. The unresolved work is caller characterization, an explicit removal decision, and implementation evidence, not selection of a destination for the combined helper.
- Test rows, target counts, and delivery history are evidence accounting only and do not justify moving source-owner behavior into Imports.

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| -------- | ------------- | -------- | -------------- | ------------------------------- | ------------- | ----- |
| 11 Imports HTTP operations: create session; get session; list units; get unit; get preview; mapping preview; put mapping; select; skip; create region; apply | Core 01 Imports route owner; Imports adapter implementation | Imports OpenAPI owner input and `http_handlers.go` binding map | Route characterization, integration suite, frontend coordinator, browser E2E | Preserve exact paths/methods/operation IDs and add an auth/CSRF matrix for member mutations | critical | No route move, alias, or v2 path is planned. |
| Session and unit resource envelopes, status tokens, arrays, absent mapping fields, and cursor pagination | Core 01 | `contracts/imports/schemas.v1.json`, store serializers and handlers | Upload/replay/read integration and frontend coordinator paging tests | Add focused resource parity only if handler delegation changes serialization | high | Singleton reads reject paging; list uses common cursor behavior. |
| Preview shape, first-50 cap, cells/columns, warnings, and locator vocabulary | Core 01/Core 03 Imports sections | Discovery/XLSX implementation and schemas | XLSX/CSV discovery, hidden-sheet characterization, frontend assistant | Preserve CSV/XLSX preview parity during source-adapter movement | high | Preview is read-only. |
| Strict upload/mapping/action/apply/region request decoding and safe error vocabulary | Core 01; Core 04 for disclosure order | `api.go`, OpenAPI, error registries | Metadata, mapping, region, safe owner error integration tests | Add authorization/CSRF-before-body cases for mutating member routes | critical | Unknown or unsafe target details must not escape. |
| Mapping fingerprint, exhaustive source columns, transforms, empty-value policy, and approved mapping readback | Core 01/Core 03 | Imports API/mapping code and contracts | Mapping, `write_null`, selection retention, Network Flow fingerprint tests | Cross-path parity for the stable portion shared with clipboard Tabular Ingest | critical | Persisted fingerprint bytes are a compatibility surface. |
| Select, skip, reselect, overlap prevention, and operator regions | Core 01/Core 03 | `selection.go`, `regions.go`, store | Selection/region integration, frontend unit, browser E2E | Retain overlap and mapping-retention cases across handler separation | high | Skipping does not erase an approved mapping. |
| Discovery/apply job identity, progress, cancellation, result code, and import-session resource refs | Common Jobs plus Core 01 Imports result semantics | `jobs.go`, finalizer port, extensionassembly adapter | Upload, apply, partial cancellation, E2E polling | Preserve recovery of committed unit results after job interruption | critical | Durable session/unit statuses do not become job-status tokens. |
| Per-unit transaction, one change set, owner journal, idempotency, crash recovery, and partial terminal outcomes | Core 01 Imports unit commit; Revisions for history substrate | apply, owner apply, store, unit outcome code | Current-state revalidation, rollback, crash recovery, partial cancellation | Add failure injection only if orchestration boundaries are physically moved | critical | No all-session atomicity is inferred. |
| View-owner row creation for 17 generated view targets | Core 01 dispatcher slots; each source owner for exact mutation semantics | Authored target catalog, generated registry, importassembly | Timeline, entity/write-null, Evidence parity, source-owner tests | Preserve exact registry/facade admission; add rows only for real uncovered owner behavior | critical | 14 enabled/selectable and 3 reserved/hidden view targets are authored at baseline. |
| Network Flow mapping preview, durable approval, opaque source, owner errors, and atomic table publication | Core Imports for orchestration; Network Flow NLSpec for exact payload/target behavior | Analytical binding and schemas, live facade and transaction participant | Network Flow atomic apply, safe owner errors, frontend controller tests | No binding/schema change until RB-001 is resolved; preserve current behavior meanwhile | critical | Owner contradiction blocks target contract redesign, not ordinary freeze testing. |
| Authorization and incident roles | Core 04; Core 01 route-specific role sets | `http_handlers.go`, `httpauth`, Incidents admission | Viewer denial and editor/admin success within integration cases | Full cookie missing/invalid/valid CSRF and bearer matrix for create, mapping, select, skip, regions, and apply | critical | Current member-route state-changing classification is incomplete. |
| Collaboration and projection consequences | Source owners, Collaboration, Revisions, Projections | Owner facade implementations and finalization helper | Owner integration/atomicity tests | Preserve publication only after commit and no effect on failure/replay | high | Imports owns no WebSocket route. |
| Saved views and view-schema behavior | Saved Views and View Schema owners; Imports only consumes generated target/view contracts | Target registry and frontend target adapter | Generated target frontend tests | No new test unless target registry consumption changes | medium | Saved-view lifecycle is not directly affected; `intentional/no_action`. |
| Generated Go/TypeScript import registry and HTTP surfaces | Contract owners and generators | `internal/gen/importtargetregistry/registry_gen.go`; generated protocol files | Generated target tests and drift targets | Run drift checks if an owner input changes; never hand-edit | high | No generated change is planned for a behavior-preserving refactor. |
| UI selector and grid-adapter contract | UI Contracts and Grid Adapter | `workbookImportAssistantTestId`; no vendor import in Imports UI | UI-contract unit, assistant unit, browser E2E, Network Flow boundary test | Preserve selector and adapter isolation if frontend ever enters scope | medium | Direct grid behavior is `not directly affected/no action`. |
| Recovery contribution and harness accounting | Imports for source-state contribution; Testing Harness for rows/evidence | Recovery catalog, verification owner, 29-row test-family manifest | Recovery catalog tests and all owner rows | Update authored row/boundary inputs only when an actual selector/path changes | high | Phase maps and rows are not runtime architecture. |

### 4.1 Gap-resolution state

Planning resolution and implementation closure are separate states. The current
workstream and gap states below are authoritative for this execution. A gap
remains open until all named owner, implementation, and evidence conditions pass.

| ID | Required resolution | Normative location | Supporting location | Current closure state |
| --- | --- | --- | --- | --- |
| RB-001 | Reconcile Extensions `0.10.0`, Network Flow major `5`, durable state `4`, and no major-4 compatibility as one owner bundle | Core 03 and Network Flow Activity NLSpec; audit Extensions companion facts | Authored owner inputs, generated registries, descriptors, bindings, client support, drift and owner evidence | DONE in WS-01; no machine projection rewrite was necessary |
| RB-002 | Classify exactly six Imports operations as state-changing for cookie CSRF and keep mapping preview read-only | Core 01 route inventory and Core 04 security/conformance | Imports characterization and service-backed tests | DONE in WS-03; closed route catalog and exhaustive security evidence pass |
| RB-003 | Use one shared view-target mapping kernel while retaining distinct file and clipboard lifecycle envelopes | Core 01/Core 03; Core 04 acceptance | Golden mapping, fingerprint, parser, owner-request, provenance, and hostile-input fixtures | DONE in WS-08; both adapters invoke the pure kernel, duplicate execution logic is removed, and all compatibility/hostile-input evidence passes |
| RB-004 | Replace boundary schema v2 with v3 exact file sets that prove the live transport set was scanned | Testing Harness public-input contract | Authored schema, checker, backend-boundary policy, fail-closed fixtures, and Make-owned evidence | DONE in WS-05; v2 removed and exact Imports transport accounting passes |
| RB-005 | Remove the combined helper; each source owner coordinates separate Revisions, Projections, and Collaboration ports | Core 00/Core 04 and the already-adopted Collaboration boundary decision | Caller/effect matrix, failure injection, package graph, and owner slices | DONE in WS-17; all owner cutovers, aggregate removal, negative boundaries, package compilation, and fast-suite evidence pass |
| RB-006 | Keep `http_handlers.go` as the sole transport-binding file while moving application orchestration behind private typed methods | Existing Imports route and behavior owners; no external contract change | Exact route catalog, transport review, application tests, service-backed tests, and browser evidence | DONE in WS-04; thin transport, application boundaries, owner tests, and browser parity pass |

### 4.2 RB-001 required owner reconciliation

The least-disruptive internally consistent line supported by the adopted corpus
is the following. This table is a required owner amendment, not a local choice.

| Fact | Required value |
| --- | --- |
| Extensions NLSpec | `0.10.0` |
| Extensions contract major | `2` |
| Network Flow document version | `5.0.x` after the corrective revision |
| Network Flow contract major | `5` |
| Network Flow durable state | `4` |
| Minimum migratable Network Flow state | `3` |
| Network Flow runtime dependency | Exactly `import@1` |
| Client-supported Network Flow major | `5` |
| Major-4 compatibility | None |

The owner bundle MUST change Core 03 `REQ-03-011A` from Network Flow major
`4` to major `5`. Network Flow Table 1-B MUST name Extensions `0.10.0`, the
exact adopted digest used by the owner process, and only the current imported
schemas, algorithms, and requirement locators. A blanket withdrawn range such
as `EXT-REQ-001..236` MUST NOT stand in for the exact dependency boundary.

The same owner change MUST audit Core 00 recognition, Core 01 discovery and
analytical binding, Core 04 conformance, Extensions fragments and dependency
facts, packaged client support, implementation bindings, worker runtime,
durable state and migrations, Recovery v4, Reporting, and generated Go and
TypeScript consumers. An already-correct artifact receives a recorded
no-change parity result rather than an unnecessary rewrite.

No major-4 decoder, translator, alias, fallback support row, hidden flag,
second discovery item, state downgrade, or binding adapter may be introduced.
A major-4-only client MUST omit the Network Analysis workspace. Existing
Network Flow import preview and apply bytes MUST remain unchanged unless a
separately adopted owner revision versions them.

### 4.3 RB-002 exhaustive Imports CSRF contract

The route's semantic effect, not the HTTP method alone, determines the cookie
CSRF requirement. The effective authentication mode comes from Platform;
Imports MUST NOT define cookie-versus-bearer precedence locally.

| Method and path | OpenAPI operation ID | Semantic class | Cookie authentication | Bearer authentication |
| --- | --- | --- | --- | --- |
| `POST /api/v1/import-sessions` | `createImportSession` | State-changing | CSRF required | No CSRF token required |
| `GET /api/v1/import-sessions/{import_session_id}` | `getImportSession` | Read-only | No CSRF required | No CSRF required |
| `GET /api/v1/import-sessions/{import_session_id}/units` | `listImportUnits` | Read-only | No CSRF required | No CSRF required |
| `GET /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}` | `getImportUnit` | Read-only | No CSRF required | No CSRF required |
| `GET /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/preview` | `getImportUnitPreview` | Read-only | No CSRF required | No CSRF required |
| `POST /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/mapping-preview` | `previewImportUnitExtensionMapping` | Read-only owner preview | No CSRF required | No CSRF required |
| `PUT /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/mapping` | `putImportUnitMapping` | State-changing | CSRF required | No CSRF token required |
| `POST /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/select` | `selectImportUnit` | State-changing | CSRF required | No CSRF token required |
| `POST /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/skip` | `skipImportUnit` | State-changing | CSRF required | No CSRF token required |
| `POST /api/v1/import-sessions/{import_session_id}/units/{base_unit_id}/regions` | `createImportUnitRegion` | State-changing | CSRF required | No CSRF token required |
| `POST /api/v1/import-sessions/{import_session_id}/apply` | `applyImportSession` | State-changing | CSRF required | No CSRF token required |

Every Imports route MUST stop at the first failed gate in this order:

1. Validate the method and only the basic framing needed to select the route.
2. Authenticate through the platform owner.
3. Consume the platform-derived effective credential mode.
4. For cookie authentication on one of the six state-changing operations,
   validate CSRF before path-specific, incident-specific, role-specific, or
   body-specific diagnostics.
5. Decode only the bounded scope required to identify the incident and
   addressed resource.
6. Apply visibility and concealment for absent, foreign, or hidden resources.
7. Apply the route's existing minimum role.
8. Decode and validate the complete JSON or multipart request.
9. Validate lifecycle and current state.
10. Resolve idempotency or exact replay.
11. Perform the mutation, job admission, or target-owner dispatch.

For session creation, CSRF MUST precede malformed metadata, duplicate parts,
unsupported media types, and other import-envelope diagnostics. After CSRF
succeeds, the route MAY decode only bounded metadata needed to establish
`incident_id`, visibility, and role before full multipart validation.

An unauthenticated request MUST return the existing session-required result
before CSRF evaluation. Missing or invalid cookie CSRF MUST return HTTP `403`
with `error.code='csrf_verification_failed'`. The rejection MUST create no
session, unit, approved mapping, selection, operator region, apply job,
idempotency result, owner mutation, revision, projection effect,
Collaboration intent, or domain audit occurrence. Safe security telemetry MAY
record the failure without the token or incident content. Valid-cookie and
bearer requests MUST preserve every later authorization, replay, response, and
error rule. Mapping preview remains authenticated and authorized but read-only
for this classification; a bounded preview cache is not mapping approval or a
domain mutation.

### 4.4 RB-003 shared view-target mapping kernel

The shared contract covers view-schema mapping semantics, not the complete
file-import or clipboard lifecycle. A logical name such as
`cartulary.tabular_mapping_kernel.v1` MAY be used during design, but only the
Core owner may adopt its permanent identifier.

| Boundary | Required contents | Explicit exclusions |
| --- | --- | --- |
| Shared mapping kernel | Ordered source-column rules, stable `field_key`, nullable `entity_binding_mode`, transform/options, empty-value policy, unknown-column disposition, target normalization, deterministic row plan, common provenance, and a closed safe failure | Sessions, units, file locations, workbook objects, UI identity, actors, transactions, and consumer ports |
| File Imports envelope | Source kind/hash, parser profile/version, session/unit IDs, locator, header/data row references, warnings, approval, persisted fingerprint, duplicate-apply tuple, jobs, and cancellation | Clipboard target anchors and row-version conflicts |
| Clipboard envelope | Dispatch decision, active `view_schema_id`, target anchor or existing record IDs, row versions/conflicts, one paste change set, and clipboard origin | Import session/unit, file locator/hash, durable mapping approval, and file mapping fingerprint |
| Source-owner boundary | Exact field validation, entity reuse/stub behavior, record creation, private revision facts, projection request, and public Collaboration effect | File parsers, workbook heuristics, grid callbacks, and generic whole-row semantic inference |
| Analytical target | Exact target-owned mapping, preview, apply, errors, and publication | Generic view-target kernel while RB-001 is open; Network Flow remains target-owned |

#### Shared interface

Every kernel invocation MUST provide all of the following; none has an
implicit default:

| Input | Required shape and rule |
| --- | --- |
| Target identity | One immutable `view_schema_id` admitted by the current generated registry |
| Field registry | One immutable validated target-field snapshot, including field order, writability, clearability, normalization capability, and exact entity-binding mode |
| Source columns | One exhaustive ordered plan with contiguous one-based ordinals |
| Source rows | One ordered scalar matrix supplied through an owner-neutral tabular source abstraction |
| Unknown policy | Exactly one explicit current policy: `preserve_raw_capture`, `preserve_custom_attrs`, or `reject_if_unmapped` |
| Transform registry | The exact closed current transform set and option validators |
| Empty-value registry | The exact closed current empty-value set |
| Cancellation | One cancellation context checked before producing the final plan |

Each source-column plan entry MUST carry exactly
`source_column_ordinal`, `source_header_text`, `field_key`,
`entity_binding_mode`, `transform_id`, `transform_options`, and
`empty_value_policy`. `field_key=null` means intentionally unmapped;
`entity_binding_mode` MUST be null for an unmapped or non-entity target. No
`import_session_id`, `import_unit_id`, path, object key, sheet/parser object,
UI label as identity, grid object, actor, transaction, Revisions port,
Projections port, or Collaboration port may cross this boundary.

The closed mapping registries are:

| Contract | Exact current values and defaults |
| --- | --- |
| `unknown_column_policy` | `preserve_raw_capture`, `preserve_custom_attrs`, or `reject_if_unmapped`; the caller MUST supply one, so there is no kernel default |
| `transform_id` | `null`, `trim_v1`, `collapse_whitespace_v1`, `lowercase_v1`, or `split_delimited_v1`; `null` means no mapping transform |
| `transform_options` | `{}` for `null`, `trim_v1`, `collapse_whitespace_v1`, and `lowercase_v1`; only `split_delimited_v1` may use options |
| `split_delimited_v1.delimiter` | Exactly `,`, `;`, `|`, `\n`, or `\t` |
| `split_delimited_v1` Boolean options | `trim_items` and `drop_empty_items`; omission/default bytes MUST be characterized and then stated by the Core owner before extraction |
| `empty_value_policy` | Exactly `omit_field` or `write_null`; the caller MUST supply one for mapped columns |
| `entity_binding_mode` | The exact nullable mode from the immutable field registry; a caller cannot override the target field contract |

For each source row, in source order, the kernel MUST:

1. Validate that the source-column plan is exhaustive, ordered, contiguous,
   and contains no duplicate non-null destination `field_key`.
2. Read the source scalar through the supplied tabular source abstraction.
3. Apply parser extraction.
4. Apply the declared optional mapping transform.
5. Apply target-field normalization.
6. Apply `empty_value_policy`.
7. Apply the target's explicit unknown-column policy to intentionally unmapped
   values.
8. Materialize ordered owner-neutral field values and common provenance.
9. Return either one complete immutable row plan or one closed safe failure.
10. Emit no partial row plan, owner request, or durable effect on failure.

Output ordering MUST be source-row order, destination-field registry order
when required by the owner request, source-column ordinal for unknown and
provenance values, and stable identifier order for set-like secondary facts.
The successful result MUST contain all ordered row plans, owner-neutral field
values, unknown dispositions, and common provenance. Internal failure reasons
MUST map to the existing route- or owner-owned public error without adding a
new public error code in this behavior-preserving refactor.

#### Parity and canonical bytes

| Parity boundary | Required equality |
| --- | --- |
| Cross-path semantic parity | Equivalent target registry, column plan, source scalar matrix, transforms, entity modes, and unknown policy produce byte-identical shared row-plan bytes, equal owner field sequences, equal unknown dispositions, equal failures, equal entity consequences, and equal shared owner-request members |
| File persistence parity | Before and after extraction, approved-mapping read bytes, `mapping_fingerprint`, duplicate comparison, durable provenance, owner requests, and replay results are byte-identical |
| Common provenance | Source row/column ordinal, raw header, raw scalar, shared cell classification, mapped field, transform identity, and unknown disposition are equal |
| Source-specific provenance | File paths retain session, unit, source hash, parser, locator, and mapping fingerprint; clipboard paths retain clipboard/paste origin and fabricate none of those file identities |

Before extraction, the Core owner MUST close the canonicalization contract for
object-key comparison, UTF-8 and BOM handling, string escaping, number
serialization, omission versus explicit `null`, array ordering, insignificant
whitespace, duplicate-member rejection, and lowercase hexadecimal output. The
existing approved-mapping bytes are the compatibility baseline. RFC 8785 or
another serializer MUST NOT be adopted retroactively when it changes any
stored fingerprint; a canonical-byte change requires a new owner-adopted
algorithm/version and migration or compatibility decision.

The file parser profile MUST likewise define encoding, BOM behavior, line
endings, quoting, escaped quotes, embedded line breaks, blank records,
trailing empty fields, inconsistent field counts, duplicate headers, and all
bounds. An open-ended “latest CSV” dependency is invalid.

The parity fixture matrix MUST include:

| Fixture family | Required cases |
| --- | --- |
| Target owners | Timeline `mention_origin`, Hosts `entity_origin`, Identities `entity_origin`, and at least one non-entity view target |
| Mapping registries | Every transform, every legal split delimiter, both empty-value policies, every target-admitted unknown policy, mapped and intentionally unmapped columns |
| Plan validity | Blank/nonblank headers, duplicate field rejection, noncontiguous ordinals, clearable/non-clearable fields, and maximum/one-over column count |
| Scalar/canonical bytes | Empty, null, string, numeric, Boolean, timestamp-like, Unicode combining forms, JSON escapes, omission/null, arrays, and stable object-key ordering |
| Parity products | File fingerprint before/after goldens, file/clipboard row-plan bytes, owner-request bytes, common/source-specific provenance, and exact failure precedence |
| File-only compatibility | Formula-cache diagnostics, merged cells, named ranges, hidden sheets, workbook warnings, parser bounds, and hostile archive/XML input |

Any incompatible approved mapping, fingerprint, replay result, owner request,
or shared clipboard output MUST block SL-03B. It MUST NOT be silently
normalized as a behavioral correction.

### 4.5 RB-004 exact backend-boundary repair

The live authored defect is rule `retired-timeline-facade-calls` in
`tools/backend_module_boundaries.json`: it scans removed
`internal/modules/imports/routes.go`. The current exact Imports transport path
is `internal/modules/imports/http_handlers.go`. After SL-02, every exact
transport-binding file MUST be enumerated; a broad glob is forbidden.

Every declared path MUST be repository-relative, normalized, unique, contained
inside the repository, non-symlink, and a regular file that resolves exactly
once. Absolute paths, backslashes, `.` or `..` segments, duplicate normalized
paths, paths outside the repository, symlinks, missing paths, and globs MUST
fail. Zero live paths MUST fail rather than produce an empty successful scan.
An additional route-binding file absent from the policy MUST fail or trigger an
explicit manifest-update requirement.

Negative evidence MUST prove that missing, stale, duplicate, escaping, and
symlink paths fail; a forbidden retired Timeline facade call added to
`http_handlers.go` fails; the unchanged conforming handler passes; and an
omitted new route-binding file is detected. The public
`make backend-module-boundary-check` identity, outputs, and failure mapping
MUST remain unchanged unless separately specified. Testing Harness normative
text changes only if that public contract changes.

### 4.6 RB-005 independent source-owner finalization

There is no valid permanent owner for the current combined helper. Moving
`ownerfacade/finalize.go` unchanged to Imports, Revisions, Projections,
Collaboration, a source owner, or a new “finalization” module would retain the
defect. This conclusion applies only to record finalization; Imports continues
to own import-session and common-job finalization.

| Behavior | Permanent owner |
| --- | --- |
| Source-record current-state semantics | Named source owner |
| Record-envelope mutation port | Records/current-envelope owner |
| Change-set, mutation-entry, and revision append | Revisions |
| Projection refresh mechanics | Projections |
| Projection source meaning | Named source owner |
| Public record-change effect derivation | Named source owner |
| Durable Collaboration intent append and delivery | Collaboration |
| Source transaction begin/commit/rollback | Named source owner or its application service |
| Import-session and common-job finalization | Imports |
| HTTP admission and error translation | Route owner/adapter |

For every live source mutation, the source owner MUST:

1. Authenticate, authorize, validate, and establish idempotency through its
   owning route or application path.
2. Begin or borrow the source-owner transaction.
3. Apply and validate the authoritative source mutation.
4. Derive private Revisions facts directly from source semantics.
5. Derive the public Collaboration effect independently from the same source
   semantics and the catalog-admitted public-field contract.
6. Derive the projection refresh request independently.
7. Invoke the separate consumer-owned ports in the characterized deterministic
   order.
8. Commit source state, envelope, history, projection, durable Collaboration
   intent, and idempotency state atomically, or commit none.
9. Permit Collaboration delivery only from the committed durable intent.
10. On exact committed replay, return the original result without adding a
    revision, projection effect, or Collaboration intent.

Revisions MUST NOT produce the Collaboration effect. Collaboration MUST NOT
inspect or translate a Revisions representation. Projection rows MUST NOT be
used as source truth for either representation.

Shared code MAY validate stable record IDs and row versions, canonicalize an
owner-neutral field-key set, defensively copy immutable inputs, or provide pure
ordering helpers. It MUST NOT accept both Revisions and Collaboration facts in
one aggregate command, invoke multiple consumer ports, own a transaction,
authorize, inspect import sessions/units, map HTTP errors, infer owner semantics
from a whole-row diff or projection, or expose a forwarding alias to the
retired operation.

WS-09 completed the caller characterization below. Every caller is reached
through the exact Imports owner-facade registry and receives the unit-level
borrowed `pgx.Tx`; none begins or commits a nested transaction. Imports owns
the unit transaction and owner-outcome replay guard. Therefore an error from
any characterized source, projection, Revisions, Collaboration, or link step
rolls back the entire unit, while an exact committed replay returns the stored
owner outcome before reinvoking the source owner.

| Source owner | Authoritative mutation and operation family | Projection and current finalization order | Existing owner-semantic helper and required cutover posture |
| --- | --- | --- | --- |
| Artifacts | Creates the record envelope and Artifact source row; import has create only | Source create includes projection refresh/load; then capture envelope snapshot, append `record` create mutation, live revision, and public change intent | Artifacts already owns changed-field, revision-fact, version, and publication derivation used by ordinary create. Reuse those helpers without changing artifact type/view semantics. |
| Parties | Runs Party match-conflict validation, then creates or exactly reuses one active same-incident Party | Refresh/load Party projection; both paths append a record mutation, while only create appends live revision and public intent | Party already owns match-conflict, fact, version, projection, and publication behavior. Preserve the create/reuse distinction and do not publish on reuse. |
| Indicators | Captures pre-upsert snapshot, then creates, exactly reuses, or updates canonical Indicator source state | Upsert returns before row; refresh/load projection; capture after snapshot and append a record mutation; create/update append live revision and public intent, reuse does not | Ordinary Indicator mutation owns `indicator` target/version identity and changed-field-only facts/effects. The aggregate import path's generic `record:` identity and union-of-projection-cells facts are a defect to correct, not compatibility to preserve. |
| Assessments | Validates subject and assessor membership, resolves assessor, and creates Assessment source state | Insert; refresh/load Assessment projection; capture snapshot; append record mutation, live revision, and public intent | No complete Assessment-owned fact/effect helper exists. WS-13 must add derivation from Assessment semantics and the admitted public-field contract, not copy whole projection rows. |
| Entities Host | Upserts Host source state with exact-match create/reuse/update behavior and captures pre-upsert snapshot | Refresh/load Host projection; append record mutation on all paths; append live revision and public intent only for create/update | Ordinary Host mutation owns `host` target/version identity, changed-field facts, and public effects. Correct the aggregate path's generic `record:` identity and projection-row union. Keep Host logic independent from Identity. |
| Entities Identity | Upserts Identity source state with exact-match create/reuse/update behavior and captures pre-upsert snapshot | Refresh/load Identity projection; append record mutation on all paths; append live revision and public intent only for create/update | Ordinary Identity mutation owns `identity` target/version identity, changed-field facts, and public effects. Correct the aggregate path's generic `record:` identity and projection-row union. Keep Identity logic independent from Host. |
| Tasks/Decisions | Creates Task or Decision source state; Task additionally validates and synchronizes ordered link values | Refresh/load projection; allocate one mutation sequence plus one per link; append record mutation, live revision, and public intent first; then append link non-row mutations in source order | Task/Decision already owns fact, version, publication, and link semantics. Preserve contiguous sequence allocation, record-before-link mutation order, ordered links, and Decision's no-link path. Any later link append failure must roll back the earlier durable intent in the borrowed transaction. |

The aggregate helper itself is now pinned by a direct characterization test:

| Input/path | Exact current behavior retained only until its owner cutover |
| --- | --- |
| Defaults | Empty operation becomes `create`; empty created/result values become `created`; row versions accept `int`, `int32`, `int64`, or `float64`. |
| Common ordering | Read projected `row_version`; capture after snapshot; append `record` mutation with `record:<id>:<version>`; construct owner response; conditionally append live revision; conditionally append Collaboration intent. |
| Create | Appends record mutation, live revision, and Collaboration intent. Conflict/public keys are the sorted union of projection cell keys. |
| Update | Requires a before snapshot to publish; appends record mutation, live revision, and Collaboration intent. The sorted union includes unchanged and removed projection cells, demonstrating why it cannot be the target design. |
| Reuse | Appends the record mutation but no live revision or Collaboration intent, regardless of whether a before snapshot is supplied. |
| Public view effect | Builds a patch only when the Collaboration protocol admits the projected row for the target view; otherwise emits an invalidate effect. |
| Failure order | Capture failure stops all later calls; record failure stops live/publication; live failure stops publication; publication failure returns an empty response. The borrowed outer transaction supplies rollback at every point. |
| Replay | The Imports owner-outcome journal prevents the helper and source mutation from being called again after committed replay; there is no helper-local replay mechanism. |

The projection row currently supplies both generic revision facts and public
field keys. That is recorded evidence of the defect, not a behavior-freeze
requirement. Each cutover must derive its private and public representations
independently from source semantics, retain the caller-visible create/reuse/
update and link differences above, and prove parity with the corresponding
ordinary source mutation where one exists.

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| ------- | -------- | ---- | -------------- | -------------- | ------------------------ |
| Mapping `PUT`, select `POST`, skip `POST`, and apply `POST` authenticate with `StateChanging=false`; only regions is classified state-changing in the member handler | `http_handlers.go` compares only `route.Kind == "regions"`; `httpauth` gates cookie CSRF on `StateChanging`; Core 04 REQ-04-003 requires fail-closed CSRF | critical security and authorization-outcome drift | `must_fix` | Imports HTTP adapter under Core 04 | Add negative/positive characterization, then correct in a separate `requires later authorization` slice. Mapping preview remains read-only. |
| File Imports and clipboard ingest do not currently share the full stable mapping engine | Imports calls `tabularingest.ParseTableWithMaxColumns` but owns separate DTOs, mapping validation, transforms, and fingerprinting; clipboard uses `BuildTabularRowPlanV1` | conformance and persisted-fingerprint risk | `must_fix` | Imports plus stable Tabular Ingest contract | Characterize exact overlap and bytes before extracting shared primitives; stop if parity needs behavior change. |
| Boundary rule `retired-timeline-facade-calls` scans removed `internal/modules/imports/routes.go` | Live file is `http_handlers.go`; authored policy still names `routes.go` | static guard may miss the current handler and may accept an empty scan | `should_fix` | Testing Harness/backend boundary owner | Repair the exact authored path set after SL-02; make invalid/missing paths and omitted handler files fail; run the public Make check. |
| Retired `ownerfacade/finalize.go` combined source-owner, Revisions, projection-row, and Collaboration consequences | WS-09 characterization, WS-10 through WS-16 cutovers, WS-17 deletion and negative boundary | resolved; source owners now derive and coordinate separate consequences | `fixed` | Source-owner orchestration with separate Revisions, Projections, and Collaboration ports | No aggregate command, generic whole-row derivation, alias, or replacement subsystem remains. |
| Imports store contains substantial direct SQL | SQL inspection shows import-owned sessions, units, streams, plans, outcomes, journals, and route idempotency; boundary guard finds no peer tables | persistence adjacency and large-file review risk | `intentional/no_action` | Imports persistence adapter | Introduce only private application ports if useful; do not replace valid import-owned SQL with a generic storage facade. |
| View-owner mutations are injected through exact facades rather than peer stores | `owner_apply.go`, generated target registry, importassembly, boundary tests | low if seam remains exact; high if bypassed | `intentional/no_action` | Source owners behind Imports dispatcher | Preserve facade validation, one-unit transaction, and owner-controlled durable semantics. |
| Network Flow uses an analytical owner facade and transaction capability | Extension facade, Network Flow facade, binding contract, transaction participant | cross-owner atomicity | `intentional/no_action` | Imports orchestration and Network Flow target owner | Preserve opaque source and exact binding; make no schema decision while RB-001 is blocked. |
| Generated target registry is consumed by backend and frontend adapters | Generated Go and TypeScript registries; static tests reject handwritten targets | registry drift if duplicated | `intentional/no_action` | Imports contract owner and generators | Change authored inputs first and regenerate only through Make if a future owner change requires it. |
| No direct grid-vendor dependency exists in Imports backend or import assistant/controller code | Targeted import search; vendor dependency confined to `packages/grid-adapter` | low | `intentional/no_action` | Grid Adapter | Keep grid-vendor integration out of this refactor. |
| Recovery contributions remain small source-owner declarations | Recovery contribution files and recoveryassembly catalog | low | `intentional/no_action` | Imports source state plus Recovery orchestration | Preserve contribution identities and do not relocate restore mechanics. |
| Network Flow Table 1-B names Extensions 0.7.1 and Core 03 advertises client major 4, while the adopted current line otherwise declares Extensions 0.10.0 and Network Flow major 5 | Network Flow NLSpec Table 1-B, Extensions front matter/status, Core 03 REQ-03-011A, and current generated major-5 consumers | contract corpus cannot safely select one dependency/client revision | `defer` | Specification owners | `BLOCKED: owner contradiction`; adopt the complete major-5 bundle atomically before changing Network Flow binding/schema behavior. |

No production test-only assumption was found leaking into runtime Imports code. The static tests do duplicate some expected file/target facts as evidence; those tests remain downstream evidence and do not become ownership authority.

## 6. Refactor Workstreams

The approved execution is strictly serial. No dependent workstream begins until
the preceding row is `DONE`, its handoff entry is current, and
`make lint-markdown` has passed. No more than one row may be `IN_PROGRESS`.

| Workstream | Depends on | Scope and exit criterion |
| --- | --- | --- |
| WS-00 — Rebaseline controlling tracker | none | Replace the planning-only posture, add RB-006, correct RB-004/RB-005 ownership, install this serial plan, and preserve historical logs. Exit when the baseline and all open rows are accurate and Markdown lint passes. |
| WS-01 — Reconcile Network Flow owners | WS-00 | Amend Core 03 and publish Network Flow 5.0.1 against the final Extensions 0.10.0 bytes and exact interfaces. Exit when no current owner or typed input names Extensions 0.7.1 or Network Flow major 4 and owner/drift tests pass. |
| WS-02 — Adopt Imports security matrix | WS-01 | Add the exhaustive six-mutating/five-read-only classification, precedence, error, and no-effect requirements to Core 01/Core 04. Exit on contradiction-free owner text and documentation validation. |
| WS-03 — Implement CSRF correction | WS-02 | Add an exact route catalog and security matrix tests, then enforce cookie CSRF for all six mutations without changing bearer or valid-client behavior. Exit when RB-002 passes. |
| WS-04 — Separate transport/application behavior | WS-03 | Retain one `http_handlers.go` binding adapter and extract private typed session, read, mapping, selection, region, and apply use cases. Exit on exact 11-operation and frontend/browser parity. |
| WS-05 — Harden boundary-file accounting | WS-04 | Adopt Testing Harness boundary schema v3, exact file sets, fail-closed resolution and discovery equality; remove v2 support. Exit when RB-004, harness, JSON-shape, boundary, and drift checks pass. |
| WS-06 — Characterize mapping compatibility | WS-05 | Add stable file/clipboard mapping, fingerprint, owner-request, transform, Unicode/JSON, failure, and provenance goldens. Exit with reproducible baselines and no production change. |
| WS-07 — Adopt shared mapping contract | WS-06 | Specify the pure parser-neutral kernel, exclusions, registries, ordering, errors, parser boundary, and preserved fingerprint envelopes in Core 01/Core 03/Core 04. Exit when characterized bytes follow unambiguously. |
| WS-08 — Implement mapping kernel | WS-07 | Route file view-target and Timeline/Entities clipboard mapping through one kernel, retain source-owned normalization and target-owned analytical mapping, and delete duplicate logic. Exit when RB-003 and hostile-input suites pass. |
| WS-09 — Characterize finalizer callers | WS-08 | Complete the source/effect/transaction/failure/replay matrix for Artifacts, Parties, Indicators, Assessments, Host, Identity, and Tasks/Decisions. Exit with no semantics inferred solely from the aggregate helper. |
| WS-10 — Artifacts cutover | WS-09 | Use Artifacts-owned fact and publication helpers and remove its aggregate call. Exit on owner/import/failure/replay evidence. |
| WS-11 — Parties cutover | WS-10 | Use Party-owned match-conflict, fact, projection, and publication behavior. Exit on owner and Imports integration evidence. |
| WS-12 — Indicators cutover | WS-11 | Use Indicator-owned fields, identity, reuse/update, revision, and publication behavior. Exit on import and ordinary-mutation parity. |
| WS-13 — Assessments cutover | WS-12 | Add Assessment-owned fact/effect derivation and remove its aggregate call. Exit on validation, projection, failure-injection, and replay evidence. |
| WS-14 — Host cutover | WS-13 | Use Host-owned match/reuse/update, version identity, facts, and publication. Exit on Host parity and rollback evidence. |
| WS-15 — Identity cutover | WS-14 | Apply distinct Identity-owned semantics without a shared host/identity switch. Exit on Identity parity and rollback evidence. |
| WS-16 — Tasks/Decisions cutover | WS-15 | Preserve mutation-sequence allocation, ordered links, per-mutation effects, rollback, and replay for both variants. Exit when every link/consumer failure point is atomic. |
| WS-17 — Remove aggregate finalization | WS-16 | Delete historical/live combined functions, `FinalizeCommand`, generic whole-row derivation, illegitimate helpers, and aliases; add negative boundaries. Exit when RB-005 passes and the package graph is acyclic. |
| WS-18 — Validation and handoff completion | WS-17 | Run the full narrow-to-broad validation ladder, classify failures, close every RB/workstream row, and record run roots and residual risk. Exit only when RB-001 through RB-006 are `DONE`. |

### 6.1 Mandatory tracker checkpoint

At the end of every workstream:

1. Mark the workstream `DONE` or `BLOCKED`; leave no dependent workstream active
   while blocked.
2. Update Section 4.1, Section 9, the top-level status, and the relevant binary
   acceptance table.
3. Record changed files, substantive behavior, compatibility or migration
   outcome, validation commands and run roots, failures, rollback point, and
   residual risk in Section 10.
4. Run `make lint-markdown` and record its result before starting the next row.

### 6.2 Superseded planning workflows

The historical WF rows below are retained as provenance only. They do not
control sequencing or authorization for this execution.

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| ----------- | ---- | -------------------------- | --------------------------- | ----------------------------- | ---- | --------------------- | ---------- | ------------------ |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Record baseline, authority, constraints, and the sole permitted documentation edit | This tracker; owner docs and framework read-only | `git status --short`; source discovery only | Scope, branch/commit, dirty state, and non-goals are explicit. |
| WF-01 | Target inventory | chain | WF-00 | WF-02 | Inventory all 36 target files, callers, dependencies, tests, and generated consumers | `internal/modules/imports/**` plus composition/caller evidence | Read/search only | Every target file has one Section 2 row. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-04 | Map routes, storage, jobs, mutations, frontend, generated and harness surfaces to owners | Owner docs, contracts, target code, frontend adapters | Read/search only | Every discovered contract has an owner and test posture; RB-001 is explicit. |
| WF-03 | Characterization test gap analysis | parallel | WF-02 | WF-05 | Identify missing evidence before movement or normative correction | Imports tests, source-owner tests, frontend tests, harness manifest | `make explain-test-owner OWNER=module.imports` for discovery; later owner slices | CSRF, mapping byte parity, and all seven finalizer caller/effect gaps have exact scenarios. |
| WF-04 | Boundary/coupling scan | parallel | WF-02 | WF-05 | Classify ownership leaks, intentional dependencies, static drift, and prohibited aggregate seams | Imports, `tabularingest`, owner facades, adopted Collaboration boundary, boundary policy | Later `make backend-module-boundary-check` | Each finding is classified without filename-derived ownership; the combined finalizer has no proposed destination. |
| WF-05 | Facade and ownership redesign plan | chain | WF-03, WF-04 | WF-06 | Retain the legitimate Imports boundary, define the shared mapping kernel, separate transport/application seams, and require independent source-owner fact ports | `http_handlers.go`, application operation files, `tabularingest`, `ownerfacade`, source-owner callers | Review against Sections 4.3-4.6 | Public HTTP behavior remains frozen except authorized RB-002; no combined finalizer survives the completed cutover. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07 | Sequence Generic Imports, Network Flow owner reconciliation, and Finalization Architecture as explicit dependency tracks | Files named by SL-00 through SL-06, SL-NF-01, and SL-04 sub-slices | Per-slice owner checks | Each slice has authority, rollback, binary completion, and stop conditions. |
| WF-07 | Harness/test/accounting update plan | chain | WF-06 | WF-08 | Preserve 29 active owner rows and repair authored accounting only when selectors or paths change | Test-family manifest and backend boundary input in a later authorized task | Catalog/boundary checks selected by Make | No generated topology is hand-edited; row changes are conditional and owner-backed. |
| WF-08 | Validation and final handoff | chain | WF-07 | none | Run narrow-to-broad proof for every applicable completed track, record artifacts/failures, and hand off residual risk | Changed implementation/test/owner inputs from later slices | Imports and affected source-owner slices, boundary/drift, browser when applicable, `make agent-finalize`, `make check` | Exact commands, results, run roots, blockers, and any intentionally incomplete parallel track are recorded. |

### Historical execution tracks

| Track | Work | Dependency rule | Completion boundary |
| --- | --- | --- | --- |
| A — Generic Imports | Characterize; correct CSRF; separate transport/application behavior; adopt and implement the shared view-target kernel; repair boundary accounting | Runs from SL-00. Generic characterization and view-target work do not wait for RB-001. | RB-002, RB-003, and RB-004 meet their binary criteria; all frozen Imports behavior is preserved except the authorized CSRF correction. |
| B — Network Flow owner reconciliation | Atomically repair Core 03, Network Flow, Extensions companion facts, projections, bindings, and client support | Parallel specification-owner track; blocked by RB-001. No Network Flow binding/schema redesign may precede adoption. | The coordinated major-5 owner bundle and all dependent evidence pass atomically. |
| C — Finalization architecture | Characterize every caller, adopt the no-combined-operation decision, migrate source owners independently, remove the aggregate helper, and add negative boundaries | Separate from handler and mapping work. Characterization precedes adoption; adoption precedes caller migration; aggregate removal follows the last caller. | RB-005 criteria pass with no forwarding alias or aggregate finalization subsystem. |

## 7. Implementation Slice Plan

The WS rows in Section 6 are the authorized implementation slices. Each row is
separate, serial, and subject to the mandatory checkpoint. The former SL rows
below are retained only as planning provenance and are superseded wherever they
conflict with WS-00 through WS-18.

### 7.1 Superseded planning slices

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| -------- | ---------- | --------------- | ------------------------------ | -------------- | ------------------------ | ------------------ | ------------- | -------------------- |
| SL-00 | none | Characterize the complete RB-002 security matrix, RB-003 canonical/file/clipboard byte baselines, and RB-005 caller/effect behavior before production movement | Imports and `tabularingest` tests; seven source-owner callers and their tests; no production change | Tests could accidentally bless owner-nonconformant CSRF acceptance or infer semantics from whole-row output | Eleven route cases; approved mapping/fingerprint/owner-request/provenance goldens; caller transaction, fact, effect, order, replay, and failure matrix | `make test-slice OWNER=module.imports`; `make service-backed-test-slice OWNER=module.imports`; affected source-owner slices from Section 8 | Remove only incorrect new characterization; retain discovered current-state evidence and report any owner mismatch | Current valid behavior, current nonconformance, all compatibility bytes, and every source-owner difference are independently demonstrated. |
| SL-01 | SL-00 | `requires later authorization`: adopt the exhaustive route classification and Core 04 criterion, then classify mapping, select, skip, and apply as state-changing; preserve create/regions and read-only mapping preview | Core 01/Core 04 owner text, `internal/modules/imports/http_handlers.go`, Imports tests, affected projections only through owners | Missing/invalid cookie CSRF changes from acceptance to `403 csrf_verification_failed`; precedence or valid-client behavior could drift | All Section 4.3 cases, exact no-side-effect checks, and existing route/integration rows | Focused service-backed rows; `make service-backed-test-slice OWNER=module.imports`; owner-required doc/drift checks | Revert the complete owner-and-code correction together; do not mix transport movement | RB-002 binary criteria pass and all valid cookie/bearer route behavior remains unchanged. |
| SL-02 | SL-01 | Make each handler a transport/admission/translation adapter that delegates to session, mapping, selection, region, and apply application seams | `http_handlers.go`, `service.go`, existing operation files, and narrowly named private application files if required | Route, method, operation ID, auth/CSRF order, concealment, body precedence, replay, and bytes | All 11 route, resource, mapping, selection, region, apply, owner, frontend, and E2E contract tests | `make test-slice OWNER=module.imports`; `make service-backed-test-slice OWNER=module.imports`; `make backend-module-boundary-check` | Move one operation family at a time and revert only the last delegation; no storage or schema change | Transport files contain only binding, admission, decoding/encoding, and error translation; frozen behavior is equal. |
| SL-03A | SL-00 | `requires later authorization`: adopt the exact shared-kernel interface, exclusions, registries, ordering, canonicalization, CSV profile, failure mapping, and parity criteria | Core 01/Core 03/Core 04 owner text and supporting canonical/parser fixtures | An incomplete owner contract would let file and clipboard paths diverge or silently change fingerprint bytes | Section 4.4 fixture matrix and existing approved-mapping/clipboard goldens | Owner-document checks; `make lint-markdown`; applicable drift checks only if owner inputs change | Revert the owner amendment if it cannot describe existing bytes without contradiction; do not implement first | Two independent implementers can produce interchangeable shared row plans and preserve file-only bytes from the adopted contract. |
| SL-03B | SL-03A | Implement the adopted shared view-target mapping kernel and route both file Imports and clipboard ingest through it while preserving separate envelopes | `internal/modules/tabularingest`; Imports `api.go`, `mapping.go`, `discovery.go`, `owner_apply.go`; clipboard consumers; tests | Approved-mapping, fingerprint, replay, provenance, owner-request, clipboard, normalization, and failure drift | Every Section 4.4 fixture plus existing file, clipboard, source-owner, and hostile XLSX suites | Imports owner slices, affected source-owner slices, `make test-fast` after narrow success, and applicable drift/static checks | Stop and do not merge at the first incompatible golden; preserve the current implementation and open an owner-authorized version/migration decision | One shared kernel serves view targets, no excluded identity/port crosses it, and every required byte and semantic parity assertion passes. |
| SL-NF-01 | RB-001 owner authority | `BLOCKED: owner contradiction`: atomically reconcile Core 03, Network Flow Table 1-B, Extensions companion facts, client support, bindings, state/recovery facts, and generated projections on the major-5 line | Owner documents, authored extension inputs, Network Flow contracts, client support, bindings, migrations/recovery declarations, generated Go/TypeScript consumers | Partial adoption, accidental major-4 compatibility, workspace availability, binding/schema drift | Major-4 omission; claimed/unclaimed major-5 workspace; existing import preview/apply bytes; registry, binding, state, recovery, and client parity | Affected owner slices, registry integrity, generated drift/policy, frontend support tests, then repository gates | Revert the entire atomic owner bundle; an intermediate commit MUST NOT claim coordinated current status | All RB-001 criteria pass in one adopted bundle; no current `0.7.1` dependency or major-4 support remains. |
| SL-04A | SL-00 | Complete the exact per-operation caller/effect matrix for Indicators, Parties, Tasks/Decisions, Assessments, Artifacts, Host, and Identity | Seven caller operations and their owner tests; Revisions, Projections, and Collaboration test support | Convenience normalization could erase caller-visible differences or miss non-import behavior | Transaction ownership, source/envelope rows, private facts, projection request, public keys/views, order, failure at each step, replay, import/non-import parity | Owner slices listed in Section 8; no production mutation required | Remove only an incorrect characterization assertion; preserve all observed differences | Every Section 4.6 caller row has complete causal and failure evidence. |
| SL-04B | SL-04A | `requires later authorization`: adopt an implementation architecture decision naming source-owner orchestration, separate consumer ports, import directions, transition, no-compatibility rule, and complete caller set | Collaboration/Revisions decisions and exact Core adoption/traceability locations selected by their owners | A tracker-only decision cannot retire the aggregate operation; duplicate normative text could drift | Owner consistency and traceability review; Markdown validation | Owner-prescribed documentation checks and `make lint-markdown` | Revert the complete adoption bundle if owners conflict; production remains unchanged | The adopted architecture explicitly forbids a combined revision/publication operation and binds all callers/evidence. |
| SL-04C-* | SL-04B | Migrate each named source owner independently to explicit source-semantic Revisions facts, projection requests, and Collaboration effects through separate ports in its borrowed transaction | `module.indicators`, `module.parties`, `module.tasksdecisions`, `module.assessments`, `module.artifacts`, and `module.entities` source-owner/application assembly | Atomicity, field disclosure, conflict reconstruction, projection refresh, call order, replay, and import/non-import parity | Per-owner Section 4.6 matrix, failure injection at every port, exact replay, and existing owner/import tests | The corresponding owner slice and service-backed slice when available; Revisions, Projections, Collaboration, and Imports slices as affected | Revert one complete source-owner cutover; unmigrated callers may temporarily retain the old helper, but no new caller may adopt it | Every named owner has an independently passing cutover and no migrated path invokes the combined helper. |
| SL-04D | All SL-04C-* | Remove `FinalizeCommand`, combined finalization functions, semantic whole-row diff, aggregate port imports, and every compatibility alias; retain only permitted pure helpers or move them to a neutral owner | `ownerfacade/finalize.go`, remaining ownerfacade tests, boundary policies, app assembly, affected owners | Dead compatibility surface, import cycles, hidden combined behavior, job-finalizer confusion | Negative import/symbol checks, full failure-injection matrix, replay, package graph, and unchanged Imports job finalization | All affected owner slices; `make backend-module-boundary-check`; Imports/Revisions/Projections/Collaboration slices | Revert the complete final removal if any caller remains; do not add a forwarding alias | No aggregate operation or alias exists, package graph is acyclic, independent ports remain, and `job_finalization.go` behavior is unchanged. |
| SL-05 | SL-02 | Repair rule `retired-timeline-facade-calls` to enumerate the final exact Imports transport file set and enforce path-integrity/omitted-handler failures; update rows only for real selector changes | `tools/backend_module_boundaries.json`; authored test-family inputs only if selectors change | Empty/stale scan success, path escape, incomplete handler coverage, or generated accounting drift | Missing, stale, duplicate, escape, symlink, forbidden-token, conforming-handler, and omitted-new-handler cases | `make backend-module-boundary-check`; catalog checks through `make check`; `make generate-drift` only when generated inputs change | Revert authored input and generated refresh together; never hand-edit generated topology | RB-004 binary criteria pass and every exact production transport path is covered. |
| SL-06 | Track A complete; include Track B or C only when those tracks changed in the authorized task | Run narrow-to-broad owner, service-backed, static, drift, browser-when-applicable, finalizer, and full checks; append exact implementation handoff | Every file changed by the applicable authorized tracks | Incomplete evidence, stale retained roots, or unrelated broad-suite failure | Preserve all 29 active Imports rows, every affected source-owner row, and every newly admitted row | Commands in Section 8 | Revert the smallest failing slice; retain artifacts and classify product, harness, or unrelated failures | Required commands pass, or every failure has a run root, relation assessment, blocker, and safe next action; incomplete parallel tracks remain explicit. |

No external public route, operation ID, JSON envelope, database schema,
migration, generated wire type, frontend workflow, or external API change is
planned. RB-002 is an explicitly authorized-later security outcome correction.
RB-005 intentionally removes a repository-internal exported Go seam only after
all callers migrate; no compatibility wrapper is allowed. Discovery of any
other public change MUST stop the slice and require a new owner-approved task.

## 8. Validation Plan

These commands were discovered from the public Make surface. They are planned commands, not results from this documentation session.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| ---------------- | ------- | ----- | ------------------------------- | ----- |
| unit | `make test-slice OWNER=module.imports` | All 29 active owner rows when `ROWS` is omitted, including unit/frontend and other non-service selections according to the harness | yes | Establish the owner baseline and rerun after safe slices. Use exact `ROWS` for the narrowest feedback when tests are added. |
| integration | `make service-backed-test-slice OWNER=module.imports` | The 17 service-backed Imports rows reported by `make explain-test-owner` | yes | Required before mutation, transaction, persistence, recovery, or owner-dispatch slices. |
| e2e/browser | `make browser-e2e-webserver-backed` | Claimed/unclaimed production import assistant workflow | no | Required before final handoff only if routes, claim gating, frontend coordination, or end-to-end workflow code changes. |
| generated drift | `make generate-drift` and `make generated-artifact-policy-check` | Regeneration equivalence and generated-root policy | no | Required after owner-input/contract changes and at final handoff when generated consumers are implicated. Do not hand-edit generated roots. |
| import-boundary/static | `make backend-module-boundary-check`; if frontend changes, also `make frontend-import-boundary-check` | Backend module policy and conditional frontend import isolation | yes | The current backend policy contains the stale `routes.go` finding; running it before repair may not prove coverage of `http_handlers.go`. |
| full check | `make agent-finalize` followed by `make check` | Final generated state, complete local correctness and policy gate | no | Run after narrow checks. Omit `RESULTS_DIR` unless a qualifying successful full warm `check` root exists. |

Affected source-owner validation for Track C is explicit. Each command remains
a planned command until a later authorized task actually runs it.

| Owner | Narrow command | When required |
| --- | --- | --- |
| Indicators | `make test-slice OWNER=module.indicators` | SL-04A characterization and the Indicators SL-04C cutover |
| Parties | `make test-slice OWNER=module.parties` | SL-04A characterization and the Parties SL-04C cutover |
| Tasks/Decisions | `make test-slice OWNER=module.tasksdecisions` | SL-04A characterization and the Tasks/Decisions SL-04C cutover |
| Assessments | `make test-slice OWNER=module.assessments` | SL-04A characterization and the Assessments SL-04C cutover |
| Artifacts | `make test-slice OWNER=module.artifacts` | SL-04A characterization and the Artifacts SL-04C cutover |
| Entities Host and Identity | `make test-slice OWNER=module.entities` | SL-04A characterization and both entity-operation SL-04C cutovers |
| Revisions | `make test-slice OWNER=module.revisions` | Every SL-04C cutover and SL-04D aggregate removal |
| Projections | `make test-slice OWNER=module.projections` | Every SL-04C cutover that changes projection request coordination and SL-04D |
| Collaboration | `make test-slice OWNER=module.collaboration` | Every live-owner SL-04C cutover and SL-04D aggregate removal |

Before running a service-backed variant for any owner, the implementation
session MUST use `make explain-test-owner OWNER=<owner-id>` and the public task
guide to confirm that the owner has applicable service-backed rows. This
discovery rule prevents invention of unsupported targets while keeping the
owner list decision-complete.

Required scenarios for a later implementation task:

- Exact 11-route inventory, methods, operation IDs, success envelopes, common errors, and path parsing.
- Cookie CSRF missing, invalid, and valid cases; bearer behavior; unauthenticated behavior; incident visibility and role concealment; body-validation precedence.
- Upload and action exact replay, `client_txn_conflict`, mapping fingerprint stability, approved mapping readback, selection retention, overlap, and operator-region replay.
- CSV/XLSX discovery, first-50 preview, hidden sheets, formula-cache blockers, dynamic named-range rejection, archive abuse, and inert hostile content.
- Shared Tabular Ingest behavior for the subset common to clipboard and file import, with exact fingerprint and provenance parity.
- View-owner transaction atomicity, one unit change set, revisions, projection refresh, Collaboration publication, safe errors, crash recovery, and cancellation outcomes.
- Per-caller independent fact/effect derivation for Indicators, Parties, Tasks/Decisions, Assessments, Artifacts, Host, and Identity; failure injection at each Revisions, Projections, and Collaboration port; exact replay without duplicate effects; no aggregate finalizer or alias.
- Backend boundary negative cases for missing, stale, duplicate, escaping, absolute, backslash, dot-segment, symlink, glob, non-regular, empty, and omitted-handler path sets, plus a forbidden retired Timeline facade call in the live handler.
- Network Flow preview side-effect freedom, approved mapping parity, opaque source capability, safe owner errors, atomic table publication, and no binding/schema change while RB-001 is blocked.
- Generated target registry consumption, reserved/claim-gated targets, frontend assistant availability/workflow, stable selector, and no direct grid-vendor import.
- Imports recovery contribution identities and all 29 active `module.imports` harness rows.

Commands actually run during planning and tracker authoring:

- Read/search discovery: `git status`, `git rev-parse`, `find`, `wc`, `rg`, `sed`, and `jq` over the files and contracts recorded above.
- `make task-guide ROLE=module-author OWNER=module.imports`.
- `make explain-test-owner OWNER=module.imports`.
- `make explain-target TARGET=<target> DETAIL=summary` for `test-slice`, `service-backed-test-slice`, `backend-module-boundary-check`, `frontend-import-boundary-check`, `generate-drift`, and `check`.
- The framework and worktree baseline were re-read immediately before creating this tracker.
- NLSpec-style revision discovery: `date`, `git status --short`, `rg`, `sed`, `jq`, and `awk` over `temp/analysis-notes.md`, `docs/research/nlspec-spec.md`, the tracker, exact owner clauses, OpenAPI operations, the live handler, `tabularingest`, finalizer callers, verification owners, and the boundary policy.
- Documentation validation: `make lint-markdown` passed after tracker creation and was run again during this revision. The final revision result is recorded in Section 10.

No product test, build, drift, boundary, browser, service-backed, or full-check target was executed during this planning-only task. The Markdown-lint result is documentation validation only and is not product, contract, generated-artifact, or refactor validation.

## 9. Top-Level Work Tracker

| Workstream | Status | Depends on | Gap impact | Checkpoint evidence |
| --- | --- | --- | --- | --- |
| WS-00 | DONE | none | Adds RB-006; corrects RB-004/RB-005 posture | Tracker rebaseline and Markdown lint recorded in Section 10 |
| WS-01 | DONE | WS-00 | RB-001 | Network Flow 5.0.1, Core 03 major 5, final Extensions 0.10.0 digest/interfaces, owner and drift evidence |
| WS-02 | DONE | WS-01 | RB-002 specification | Core 01/Core 04 exhaustive matrix, precedence, exact error, and no-effect rules |
| WS-03 | DONE | WS-02 | RB-002 implementation/tests | Closed 11-row catalog, corrected admission, exhaustive security matrix, owner and drift evidence |
| WS-04 | DONE | WS-03 | RB-006 | Sole transport-binding file, private typed operations, owner and browser parity evidence |
| WS-05 | DONE | WS-04 | RB-004 | Boundary schema v3, exact file sets, fail-closed fixtures, and public harness gates |
| WS-06 | DONE | WS-05 | RB-003 characterization | Stable file/clipboard bytes, fingerprints, owner requests, transforms, failures, and provenance |
| WS-07 | DONE | WS-06 | RB-003 specification | Pure kernel contract, closed/defaulted registries, parser boundary, deterministic outputs, exact exclusions, and preserved v1 envelopes adopted |
| WS-08 | DONE | WS-07 | RB-003 implementation/tests | Pure mapping kernel, file/clipboard adapter cutover, duplicate transform removal, compatibility parity, and hostile-input evidence |
| WS-09 | DONE | WS-08 | RB-005 characterization | Seven-owner caller/effect/transaction/failure/replay matrix and direct aggregate characterization test |
| WS-10 | DONE | WS-09 | RB-005 | Artifacts-owned fact/version/publication derivation; no aggregate call or aggregate port type |
| WS-11 | DONE | WS-10 | RB-005 | Party-owned create/reuse facts, versioning, projection, and public effects; no aggregate dependency |
| WS-12 | DONE | WS-11 | RB-005 | Canonical Indicator identity, changed-field facts/effects, and independent ports replace aggregate behavior |
| WS-13 | DONE | WS-12 | RB-005 | Assessment-owned closed create facts/effects and canonical identity replace aggregate behavior |
| WS-14 | DONE | WS-13 | RB-005 | Host-specific identity, facts/effects, rollback, and reuse/update behavior; Identity deliberately unchanged |
| WS-15 | DONE | WS-14 | RB-005 | Separate Identity identity, facts/effects, rollback, and reuse/update behavior; no Host/Identity aggregate branch |
| WS-16 | DONE | WS-15 | RB-005 | Owner-local Task/Decision record effects preserve sequence allocation, link order, atomicity, and replay |
| WS-17 | DONE | WS-16 | RB-005 | Aggregate source/tests removed; symbols and combined aliases prohibited; boundary, drift, package graph, and fast suite pass |
| WS-18 | DONE | WS-17 | RB-001 through RB-006 | Full Imports service-backed, browser-backed, fast, finalization, lint, and 671-unit repository validation; classified failure evidence and final handoff closure |

All workstreams are `DONE`; no workstream is `IN_PROGRESS` or eligible to begin.

### 9.1 Superseded planning tracker

The historical T rows below are retained as provenance and do not control the
authorized implementation.

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| -- | --------- | ---------- | ------ | ---------- | -------------------- | -------------- |
| T-001 | Establish target, safe label, source hierarchy, baseline and one-file constraint | WF-00 | DONE | none | Section 1; branch/commit/status commands | Scope and later-authorization boundary are explicit. |
| T-002 | Inventory all 36 target files and external callers | WF-01 | DONE | T-001 | Section 2; file/symbol/import searches | Every target file has a responsibility, surface, dependency, test and risk row. |
| T-003 | Map generic Imports contracts to owners and evidence | WF-02 | DONE | T-002 | Section 4; owner docs, contracts, code and test rows | Every discovered contract risk has an owner and test posture. |
| T-004 | Adopt the atomic Network Flow/Extensions/Core major-5 reconciliation | WF-02 | BLOCKED | T-003 | RB-001; SL-NF-01 | Specification owners adopt the complete bundle; no current 0.7.1 dependency or major-4 client row remains. |
| T-005 | Add security, mapping-byte, and finalizer caller/effect characterization | WF-03 | TODO | T-003 | SL-00, SL-04A | Current nonconformance, compatibility bytes, and all caller differences are independently demonstrated. |
| T-006 | Correct mutating member-route CSRF admission | WF-05 | TODO | T-005 | RB-002; SL-01 | Later-authorized correction and security matrix pass. |
| T-007 | Adopt and implement the shared view-target mapping kernel | WF-05 | TODO | T-005 | RB-003; SL-03A, SL-03B | Exact owners are complete and one kernel passes all semantic and byte-parity criteria. |
| T-008 | Separate HTTP transport from application orchestration inside Imports | WF-05 | TODO | T-005, T-006 | SL-02 | Public surface is unchanged and handler/application responsibilities are explicit. |
| T-009 | Decide permanent ownership of the combined finalizer | WF-05 | DONE | T-003 | RB-005; Core 00 REQ-00-075; Core 04 AC-564; Section 4.6 | The decision is “no combined permanent owner”; the source-owner independent-port topology is explicit. |
| T-010 | Repair stale backend boundary policy and conditional harness accounting | WF-07 | TODO | T-008 | RB-004; SL-05 | Current handler path is covered and selectors/topology are synchronized through owners. |
| T-011 | Execute narrow-to-broad implementation validation | WF-08 | TODO | T-006, T-007, T-008, T-010; Track C rows when changed | Section 8 commands and future run roots | Commands pass or failures are classified with retained evidence. |
| T-012 | Complete the NLSpec-style planning tracker and session handoff | WF-08 | DONE | T-001, T-002, T-003, T-009 | This file, Sections 10-12 | Another agent can start SL-00 without repository rediscovery or an unresolved design choice. |
| T-013 | Move `ownerfacade/finalize.go` unchanged to another permanent package | WF-05 | DROPPED | T-009 | Core independent-port boundary and Section 4.6 | No later slice proposes a destination or compatibility alias for the combined operation. |
| T-014 | Adopt the no-combined-finalizer implementation architecture decision | WF-05 | TODO | T-005, T-009 | SL-04B | Adopted owner text names ports, callers, transition, import directions, and no-compatibility rule. |
| T-015 | Migrate every source-owner caller to separate consumer ports | WF-05 | TODO | T-014 | SL-04C-* | Each named caller passes its causal/failure matrix and no migrated caller invokes the aggregate helper. |
| T-016 | Remove the combined finalizer and prohibit reintroduction | WF-07 | TODO | T-015 | SL-04D | Aggregate symbols and aliases are absent, negative boundaries pass, and Imports job finalization is unchanged. |

No implementation work item is `IN_PROGRESS`.

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| `2026-08-30T09:51:22-04:00` | `/root` planning discovery | Target exists; tracker absent; planning-only scope established | Inspected framework, Core 00-04, applicable NLSpecs, domain and guides; touched none | `sed`, `rg`, `git status`, `git rev-parse` | Authority and clean baseline established | RB-001 owner contradiction | Create tracker only. |
| `2026-08-30T09:57:55-04:00` | `/root` documentation implementation | Approved plan recorded; no refactor authorized | Re-read framework and baseline; touched only `docs/handoffs/imports-module-refactor-tracker.md` | `sed`, `git status`, `git rev-parse`, `date`, `make lint-markdown`; documentation patch | Tracker created from live findings; Markdown lint passed | Implementation remains unauthorized | Begin SL-00 only under a later authorized task. |
| `2026-08-30T11:03:42-04:00` | `/root` NLSpec-style tracker revision | Analysis recommendations converted into decision-complete planning requirements without changing owner authority | Inspected analysis notes, NLSpec doctrine, exact owner clauses, live handlers/mapping/finalizer/policy; touched only `docs/handoffs/imports-module-refactor-tracker.md` | `date`, `git status --short`, `rg`, `sed`, `jq`, `awk`, documentation patches, `make lint-markdown` | Twelve-section tracker preserved; 36 inventory rows verified; Markdown lint passed at `.cartulary/test-results/20260830T151310Z-p234196` | RB-001 remains owner-blocked; implementation is unauthorized | A later task may begin SL-00; no owner or product file changed here. |
| `2026-08-30T11:37:12-04:00` | `/root` WS-00 | Approved remediation authority installed; historical planning evidence preserved as superseded provenance | Changed this tracker only; re-read the approved plan and repository status | `git status --short`, `date`, documentation patches, `make lint-markdown` | WS-00 `DONE`; RB-006 added; RB-004 classified as a Testing Harness public-input change; RB-005 linked to the already-adopted decision; Markdown lint passed at `.cartulary/test-results/20260830T153837Z-p245641` | None; no compatibility or migration change in this tracker-only slice | Start WS-01; rollback is the WS-00 tracker patch; residual risk is confined to still-open RB rows. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| `2026-08-30T09:51:22-04:00` | `/root` planning discovery | 36-file package is legitimate Imports boundary plus mixed transport/persistence/coordinator responsibilities | Inspected every target Go file, `tabularingest`, app assembly, source-owner facades and Network Flow facade; touched none | `find`, `wc`, `rg`, `sed` | File inventory and caller/dependency map complete | RB-003, RB-005 | Characterize mapping overlap; preserve deferred finalizer seam. |
| `2026-08-30T09:57:55-04:00` | `/root` documentation implementation | Boundary diagnosis recorded without code changes | Touched tracker only | Documentation patch | Keep/split/defer findings and slice order recorded | No authority for production edits | Later task starts with SL-00, not package movement. |
| `2026-08-30T11:03:42-04:00` | `/root` NLSpec-style tracker revision | Shared view-target kernel and independent source-owner finalization boundaries are decision-complete for planning | Inspected `tabularingest`, Imports mapping/handlers, `ownerfacade/finalize.go`, seven caller operations, Collaboration boundary, and backend policy; touched tracker only | `rg`, `sed`, `jq`, `awk`, documentation patches | Track A and Track C interfaces, algorithms, stop rules, caller matrix, and exact boundary repair recorded | RB-003 and RB-005 remain open for adoption/implementation | Authorize SL-00 and SL-04A separately before any production movement. |
| `2026-08-30T12:18:43-04:00` | `/root` WS-04 | Imports transport delegates all session, read, mapping, selection, region, and apply orchestration through private typed application methods | Changed `internal/modules/imports/application_operations.go`, `http_handlers.go`, `mapping.go`, `regions.go`, `boundary_guard_test.go`, and this tracker | `make format`; Imports task guide and owner explanation; focused boundary/characterization rows; focused service-backed upload, selection, region, and apply rows; full Imports owner slice; `make browser-e2e-webserver-backed`; `make lint-markdown` | Formatting passed at `.cartulary/test-results/20260830T160939Z-p461915`; focused units passed at `.cartulary/test-results/20260830T161013Z-p466789`; service-backed parity passed at `.cartulary/test-results/20260830T161028Z-p467771`; all 23 full-owner units passed at `.cartulary/test-results/20260830T161114Z-p485463`; all 60 browser-backed units passed at `.cartulary/test-results/20260830T161229Z-p529930` | None; internal-only refactor changes no route, DTO, error, persistence, job, replay, frontend, or migration contract | Start WS-05; rollback is the five Imports code/test files; residual risk is the known stale boundary path until v3 exact-file accounting closes RB-004. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| `2026-08-30T09:51:22-04:00` | `/root` planning discovery | Frontend is an external contract consumer, not part of the backend target; grid vendor remains isolated | Inspected import coordinator/tests, Workbook Import Assistant/tests, Network Flow import controller, service adapters, UI selector, E2E, and grid import boundaries; touched none | `rg`, `sed`, `wc` | Frontend freeze/test posture mapped; no direct vendor coupling found | None for planning | Include frontend only if a later route or workflow slice changes it. |
| `2026-08-30T09:57:55-04:00` | `/root` documentation implementation | Frontend is explicitly out of implementation scope | Touched tracker only | Documentation patch | `intentional/no_action` recorded for grid and saved-view behavior | No frontend mutation authorized | Preserve existing frontend tests as downstream evidence. |
| `2026-08-30T11:03:42-04:00` | `/root` NLSpec-style tracker revision | Frontend remains a frozen downstream consumer; no UI or grid behavior is assigned to the shared kernel | Reused inspected frontend/selector/grid evidence; touched tracker only | `rg`, `sed`, documentation patches | Kernel exclusions and conditional browser/frontend validation are explicit | None beyond later route/workflow change scope | Run frontend/browser evidence only when a later authorized slice changes those consumers. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| `2026-08-30T09:51:22-04:00` | `/root` planning discovery | 11 OpenAPI operations, 17 view inputs plus one analytical target, and generated Go/TS consumers identified | Inspected Imports catalog/schemas, OpenAPI owner, Network Flow binding/schemas, view/recovery inputs, generated registry and HTTP files; touched none | `jq`, `rg`, `sed` | Contract freeze map complete | RB-001 | Do not change Network Flow binding/schema until owner reconciliation. |
| `2026-08-30T09:57:55-04:00` | `/root` documentation implementation | No contract or generated change planned | Touched tracker only | Documentation patch | Generated hand-edit prohibition and drift routing recorded | RB-001 | Use authored owner inputs and Make generators only in a later authorized change. |
| `2026-08-30T11:03:42-04:00` | `/root` NLSpec-style tracker revision | Exact route operation IDs, major-5 owner reconciliation, mapping contract allocation, and no-public-change posture recorded | Inspected Core 03, Network Flow/Extensions versions, Imports OpenAPI owner, and generated major-5 evidence; touched tracker only | `jq`, `rg`, `sed`, documentation patches | RB-001 bundle and RB-002/RB-003 owner placements are unambiguous; no contract/codegen file changed | RB-001 remains `BLOCKED: owner contradiction` | Specification owners must adopt SL-NF-01 atomically before binding/schema redesign. |
| `2026-08-30T11:49:05-04:00` | `/root` WS-01 | Network Flow owner corpus is consistently major 5/state 4 against Extensions 0.10.0 | Changed Core 03 REQ-03-011A and `docs/network-flow-activity-nlspec.md`; audited Extensions, client support, owner fragment, bindings, and Network Flow typed inputs without rewriting them | `make task-guide` and `make explain-test-owner` for `module.extensions`, `module.networkflow`, and `web.networkflow`; their routed slices; focused Network Flow service-backed retry; `make generate-drift`; `make generated-artifact-policy-check`; `make lint-markdown` | Network Flow 5.0.1 pins Extensions digest `f7a78f...e42d3` and exact interfaces; Core 03 digest is `311e30...4231`; Extensions passed at `.cartulary/test-results/20260830T154215Z-p250575`, web passed at `.cartulary/test-results/20260830T154312Z-p292379`, failed Network Flow root `.cartulary/test-results/20260830T154312Z-p292373` was unrelated object-store readiness infrastructure, its five affected rows passed at `.cartulary/test-results/20260830T154759Z-p351740`; drift passed at `.cartulary/test-results/20260830T154847Z-p368786`; generated policy passed at `.cartulary/test-results/20260830T154847Z-p368796`; Markdown passed at `.cartulary/test-results/20260830T154944Z-p372560` | None; compatibility intentionally removes major 4, changes no public major-5 bytes, and needs no database migration | Start WS-02; rollback is the atomic two-owner-document patch; residual risk moves to RB-002 through RB-006. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| `2026-08-30T09:51:22-04:00` | `/root` planning discovery | Owner manifest reports 29 rows: 2 browser, 5 frontend, 15 integration, 7 unit; 17 service-backed | Inspected target tests, verification owner, test-family manifest and backend boundary input; touched none | `make task-guide ROLE=module-author OWNER=module.imports`; `make explain-test-owner OWNER=module.imports`; `make explain-target ... DETAIL=summary`; `jq`, `rg`, `sed` | Narrow and broad commands discovered; no test target executed | RB-004 stale boundary path | Add characterization through authored rows only when later authorized; repair boundary input separately. |
| `2026-08-30T09:57:55-04:00` | `/root` documentation implementation | Validation plan recorded without claiming product success | Touched tracker only | Documentation patch; `make lint-markdown` | Required commands and scenarios documented; Markdown lint passed | No product/refactor validation run in this task | Later implementation establishes baseline before mutation. |
| `2026-08-30T11:03:42-04:00` | `/root` NLSpec-style tracker revision | Binary evidence exists on paper for all five RB items; affected owner commands and RB-004 negative cases are explicit | Inspected verification owner IDs and boundary-policy rule; touched tracker only | `rg`, `jq`, `awk`, documentation patches, `make lint-markdown` | Markdown lint passed at `.cartulary/test-results/20260830T151310Z-p234196`; inventory/route/caller counts are 36/11/7 | No product, service-backed, drift, boundary, browser, or full-check target ran | A later task establishes SL-00 baselines before mutation and records exact run roots. |
| `2026-08-30T12:29:19-04:00` | `/root` WS-05 | Boundary-file accounting is fail-closed and the retired Timeline-call rule scans exactly the token-discovered Imports transport set | Changed the Testing Harness owner, boundary policy, checker, new exact-file-set module, v3 schema/attachment, harness contract fixtures, task-surface owner and generated projections; removed the v2 schema; adjusted the Imports CSRF test to use Collaboration-owned intent test support; changed this tracker | `make generate`; `make format`; `make harness-contract`; `make backend-module-boundary-check`; `make json-shape-check`; `make generate-drift`; `make generated-artifact-policy-check`; `make lint-scripts`; Imports task guide/explanation and focused service-backed security row; `make lint-markdown` | Generation passed at `.cartulary/test-results/20260830T162656Z-p592414`; formatting at `.cartulary/test-results/20260830T162713Z-p595397`; harness contract at `.cartulary/test-results/20260830T162722Z-p599662`; boundary at `.cartulary/test-results/20260830T162741Z-p600283`; JSON shape at `.cartulary/test-results/20260830T162750Z-p600691`; drift at `.cartulary/test-results/20260830T162801Z-p601164`; generated policy at `.cartulary/test-results/20260830T162814Z-p604165`; script lint at `.cartulary/test-results/20260830T162824Z-p604699`; Imports security integration at `.cartulary/test-results/20260830T162833Z-p605521` | Internal harness input migration only; public Make target, command ID, summary schema, output shape, artifact contract, and failure mapping are unchanged; no product or data migration | Start WS-06; rollback is the Testing Harness owner/input/checker/schema/test/task-metadata set plus the Imports test-support adjustment; residual risk moves to RB-003 and RB-005. |
| `2026-08-30T12:40:22-04:00` | `/root` WS-06 | Pre-extraction file and clipboard mapping behavior is locked by exact reproducible characterization | Added Imports mapping compatibility characterization and extended Tabular Ingest, Timeline, Host, and Identity owner tests; changed no production source or owner contract in this workstream; updated this tracker | `make format`; focused `make test-slice` rows for `module.imports`, `module.tabularingest`, `module.entities`, and `module.timeline`; `make lint-markdown` | File approved-mapping bytes, normalized request bytes, fingerprint `8b046f...6234`, all transforms/delimiters, omitted split Boolean behavior, owner-request ordering/provenance, exact failures, four clipboard target fingerprints, exact Timeline row-plan bytes, Timeline fingerprint `8fa416...4c48`, Host request hash `ad66b0...bcfb`, and Identity request hash `e3670f...a7a0` are fixed; final focused rows passed at `.cartulary/test-results/20260830T163949Z-p641994`, `.cartulary/test-results/20260830T163954Z-p642551`, `.cartulary/test-results/20260830T163956Z-p642801`, and `.cartulary/test-results/20260830T164000Z-p643193` | No compatibility or migration impact; this is test-only characterization and intentionally preserves all current bytes, defaults, failures, and source-specific provenance | Start WS-07; rollback is the four owner test edits plus the new Imports characterization file; residual risk is specification ambiguity until the shared kernel contract is adopted. |
| `2026-08-30T12:44:19-04:00` | `/root` WS-07 | The shared mapping behavior is now an adopted, parser-neutral Core contract rather than an implementation coincidence | Amended Core 01 REQ-01-474, Core 03 REQ-03-156/188/190, Core 04 AC-028/065, and added AC-264C; changed this tracker; no production, machine projection, or domain vocabulary file changed | Exact owner-text audit; `git diff --check`; `make lint-markdown` | Owners now define immutable target/source/cell inputs, deterministic row/value/unknown/provenance output, closed transforms/delimiters/policies, omitted split booleans as `false`, cancellation/failure atomicity, parser and source-owner boundaries, exact forbidden dependencies, target-owned analytical mappings, and byte-preserved independent file/clipboard v1 fingerprint envelopes; Markdown passed at `.cartulary/test-results/20260830T164446Z-p646845` | No public interface, data, or compatibility migration; the adopted contract explicitly preserves all WS-06 file and clipboard canonical bytes and forbids a cross-lifecycle fingerprint | Start WS-08; rollback is the three Core owner amendments plus this tracker row; residual risk is implementation divergence until both adapters invoke the kernel and duplicate logic is removed. |
| `2026-08-30T13:04:06-04:00` | `/root` WS-08 | File view-target mapping and Timeline/Entities clipboard planning now invoke one pure `cartulary.tabular_mapping_kernel.v1`; Network Flow analytical mapping remains target-owned | Added `mapping_kernel.go` and exhaustive unit/boundary tests; adapted `BuildTabularRowPlanV1` and Imports owner-request construction; removed `transformImportValue`; added mapping-time non-clearable `write_null` rejection and cancellation translation; extended cross-path characterization and static boundaries; made the WS-07 ordering sentence explicitly preserve source-column owner-request order; changed this tracker | `make format`; focused/full `module.tabularingest` and `module.imports` unit slices; focused Timeline/Entities clipboard slices; focused and full Imports service-backed slices; `make backend-module-boundary-check`; `make generate-drift`; `make generated-artifact-policy-check`; `git diff --check`; `make lint-markdown` | Final format `.cartulary/test-results/20260830T170256Z-p805274`; kernel unit `.cartulary/test-results/20260830T170217Z-p803720`; final focused Imports `.cartulary/test-results/20260830T170300Z-p809415`; full Imports unit 23/23 `.cartulary/test-results/20260830T165856Z-p709576`; full Imports service-backed 14/14 `.cartulary/test-results/20260830T170019Z-p754914`; Timeline clipboard `.cartulary/test-results/20260830T165734Z-p702561`; Entities clipboard `.cartulary/test-results/20260830T165740Z-p703690`; boundary `.cartulary/test-results/20260830T170141Z-p798881`; drift `.cartulary/test-results/20260830T170350Z-p827751`; generated policy `.cartulary/test-results/20260830T170358Z-p830710`; Markdown `.cartulary/test-results/20260830T170443Z-p831465`. An initial format failed at `.cartulary/test-results/20260830T165313Z-p661745` because the new test had one missing composite-literal delimiter; the corrected rerun passed. An initial focused Imports run failed at `.cartulary/test-results/20260830T165415Z-p671673` because characterization still named the deliberately removed transform helper, and the next run at `.cartulary/test-results/20260830T165454Z-p677209` exposed its stale boundary marker; both were directly related and corrected by routing characterization through the kernel and updating the guard. | No route, DTO, storage, frontend, or data migration; file approved-mapping bytes/fingerprint `8b046f...6234`, clipboard fingerprints and exact Timeline plan bytes, owner-request ordering, replay results, and source-specific provenance remain unchanged. The internal apply path now rejects non-clearable `write_null` during mapping validation and translates kernel cancellation to the existing unit-canceled outcome. | Start WS-09; rollback is the new kernel/tests plus the two adapter cutovers and boundary updates, restoring the deleted duplicate helper only as part of a full rollback. Residual RB-003 risk is closed; remaining risk is RB-005 aggregate finalization. |
| `2026-08-30T13:16:22-04:00` | `/root` WS-09 | All seven finalizer callers have an explicit source/effect/transaction/failure/replay baseline; no target behavior is inferred solely from the aggregate helper | Added `ownerfacade/finalizer_characterization_test.go`, invoked it through the existing ownerfacade characterization row, inspected all seven callers and their ordinary owner helpers, completed Section 4.6, and changed no production source | Task guide and owner explanation for Artifacts, Parties, Indicators, Assessments, Entities, and Tasks/Decisions; focused ownerfacade, import, composition, transaction, rollback, reuse, and atomicity slices; source/helper audits; `make lint-markdown` | Aggregate characterization passed at `.cartulary/test-results/20260830T171125Z-p845665`; Artifacts import at `.cartulary/test-results/20260830T171227Z-p849140`; Parties composition/reuse at `.cartulary/test-results/20260830T171315Z-p866706` and `.cartulary/test-results/20260830T171320Z-p867555`; Indicator rollback at `.cartulary/test-results/20260830T171403Z-p884745`; Assessment import/rollback/replay at `.cartulary/test-results/20260830T171448Z-p902063`; Entities Host/Identity rollback at `.cartulary/test-results/20260830T171532Z-p919416`; Tasks/Decisions import atomicity at `.cartulary/test-results/20260830T171535Z-p919782`; Markdown at `.cartulary/test-results/20260830T171737Z-p937564`. The first characterization build failed at `.cartulary/test-results/20260830T170918Z-p839616` because the test treated opaque snapshots as field-addressable and used `testing.TB.Run`; the corrected build then failed at `.cartulary/test-results/20260830T170959Z-p844602` because an unknown characterization view correctly emits `invalidate`, not `patch`; both failures were test-only and directly corrected. | No production, external contract, database, replay, or migration change; the test freezes the aggregate only as removal evidence and explicitly does not bless its generic identity or projection-derived facts | Start WS-10; rollback is the new characterization test plus this tracker checkpoint. Residual risk is confined to the still-live aggregate callers, beginning with Artifacts. |
| `2026-08-30T13:22:35-04:00` | `/root` WS-10 | Artifacts import independently derives its Revisions facts and Collaboration effect through separate owner-local ports | Changed `artifacts/import_create.go`, `publication.go`, and its composition test; reused Artifacts' ordinary create fact, `record:` version, changed-field, patch/invalidate, and publication helpers; removed every Artifacts aggregate-finalizer reference | `make format`; Artifacts import service-backed row; full Artifacts unit slice; exact symbol audit; `git diff --check`; `make lint-markdown` | Final format passed at `.cartulary/test-results/20260830T172039Z-p962975`; import create/failure/rollback evidence passed at `.cartulary/test-results/20260830T172043Z-p967123`; all 7 unit rows passed at `.cartulary/test-results/20260830T172151Z-p984790`. An attempted nonexistent focused row failed before execution at `.cartulary/test-results/20260830T171931Z-p940405`; the first import rerun then failed to build at `.cartulary/test-results/20260830T171945Z-p944934` because the cutover made one test import unused; removal fixed the directly related test-only failure. | Repository-internal Go wiring only; create response, record mutation identity, exact revision/public fields, projection row, durable intent, rollback, and Imports replay behavior are unchanged; no data migration | Start WS-11; rollback is the complete three-file Artifacts cutover. Residual RB-005 risk is the six remaining owner operation families and the still-live aggregate helper. |
| `2026-08-30T13:26:33-04:00` | `/root` WS-11 | Party import independently coordinates owner-local Revisions and Collaboration ports and preserves create versus exact-reuse behavior | Changed `parties/import_create.go` and `publication.go`; reused Party row-version, `record:` version, changed-field, fact, and patch/invalidate derivation; retained match-conflict translation and reuse's record-mutation-only consequence; removed every Party aggregate-finalizer reference | `make format`; Party composition row; Party exact-reuse service-backed row; full Party unit slice; exact symbol audit; `git diff --check`; `make lint-markdown` | Format passed at `.cartulary/test-results/20260830T172429Z-p1004235`; composition passed at `.cartulary/test-results/20260830T172432Z-p1008386`; exact create/reuse persistence passed at `.cartulary/test-results/20260830T172457Z-p1009502`; all 20 unit/security rows passed at `.cartulary/test-results/20260830T172537Z-p1027165` | Repository-internal Go wiring only; Party match conflicts, create/reuse results, sequence/ordinal, projection rows, record mutations, create-only live/public effects, and Imports replay remain compatible; no migration | Start WS-12; rollback is the two-file Party cutover. Residual risk is Indicators' known generic identity/fact mismatch plus the four later owner operation families. |
| `2026-08-30T13:32:05-04:00` | `/root` WS-12 | Indicator import now uses the same canonical owner identity and changed-field derivation as ordinary Indicator creation/upsert | Changed `indicators/import_create.go` and its exact sibling-import boundary allowlist; removed the aggregate call and `ownerfacade.VersionID`; directly coordinated the existing Revisions and Collaboration ports using `indicator:<id>:<version>`, target kind `indicator`, owner changed-field facts, and create/update-only publication | `make format`; Indicator transaction service-backed row; full Indicator unit slice; focused structure/boundary rerun; exact symbol audit; `git diff --check`; `make lint-markdown` | Format passed at `.cartulary/test-results/20260830T172808Z-p1071221`; transaction rollback evidence passed at `.cartulary/test-results/20260830T172812Z-p1075375`; the first full unit run had 19 passing units and one directly related boundary failure at `.cartulary/test-results/20260830T172914Z-p1093506`, isolated again at `.cartulary/test-results/20260830T173116Z-p1112714`; admitting `import_create.go` as an owner-local Revisions consumer made the focused boundary/composition row pass at `.cartulary/test-results/20260830T173154Z-p1113518` | Intentional repository-internal semantic correction: imported Indicator mutation identity changes from generic `record:`/`record` to `indicator:`/`indicator`, update facts/public fields now include only owner-detected changes, and reuse carries the canonical before version; no route, schema, database, or frontend migration | Start WS-13; rollback is the two-file Indicator cutover, but do not preserve the retired generic identity through an adapter. Residual risk is Assessment's missing owner helper and the remaining Entity/Task/Decision callers. |
| `2026-08-30T13:37:26-04:00` | `/root` WS-13 | Assessment import owns a closed create fact/publication contract and coordinates separate Revisions and Collaboration ports | Changed `assessments/import_create.go`, added owner-local `publication.go`, and updated constructor test stubs; removed the aggregate call/type, used strict owner row-version admission, canonical `assessment:`/`assessment` identity, an immutable eight-field create registry, independently copied private/public field lists, and existing patch/invalidate protocol | `make format`; Assessment import identity unit row; Assessment import plus facade rollback/replay service-backed rows; full Assessment unit/frontend slice; exact symbol audit; `git diff --check`; `make lint-markdown` | Format passed at `.cartulary/test-results/20260830T173513Z-p1117349`; import identity passed at `.cartulary/test-results/20260830T173517Z-p1121502`; import/facade rollback and replay passed at `.cartulary/test-results/20260830T173526Z-p1122534`; all 28 unit/frontend rows passed at `.cartulary/test-results/20260830T173616Z-p1140455` | Intentional repository-internal correction from generic `record:`/`record` to canonical `assessment:`/`assessment`; subject/assessor validation, projection refresh, create response, public view, rollback, and replay remain unchanged; no data migration | Start WS-14; rollback is the Assessment import/publication/test-stub set. Residual risk is the shared Host/Identity caller and Tasks/Decisions. |
| `2026-08-30T13:42:28-04:00` | `/root` WS-14 | Host import is independently finalized with Host-owned identity and semantics; the Identity branch remains on the aggregate helper for WS-15 | Changed only `entities/hostidentity/import_create.go`; split the Host branch into a typed owner-local finalizer, reused exact Host version, changed-field, revision-fact, and publication helpers, and preserved create/reuse/update response behavior | `make format`; Host/Identity rollback unit row; Imports entity-owner service-backed row; full Entities unit slice; branch-specific symbol audit; `git diff --check`; `make lint-markdown` | Format passed at `.cartulary/test-results/20260830T173927Z-p1185968`; rollback contract passed at `.cartulary/test-results/20260830T173931Z-p1190121`; Imports entity-owner application passed at `.cartulary/test-results/20260830T173938Z-p1190437`; all 42 Entities unit/frontend rows passed at `.cartulary/test-results/20260830T174024Z-p1208515` | Intentional repository-internal correction for Host from generic `record:`/`record` to `host:`/`host`, owner changed fields, and canonical before-version handling; routes, rows, create/reuse/update result codes, projection refresh, rollback, and replay remain unchanged; no migration | Start WS-15; rollback is the Host-only helper/branch change without touching Identity. Residual risk is exactly the still-visible Identity aggregate call plus Tasks/Decisions. |
| `2026-08-30T13:45:06-04:00` | `/root` WS-15 | Identity import is independently finalized and no shared Host/Identity switch hides owner-specific version/fact behavior | Changed only `entities/hostidentity/import_create.go`; replaced the remaining aggregate branch with a separate Identity finalizer using `identity:`/`identity`, owner changed fields, exact before/after snapshots, and Identity view publication | `make format`; Host/Identity rollback unit row; Imports entity-owner service-backed row; exact aggregate-symbol audit; `git diff --check`; `make lint-markdown` | Format passed at `.cartulary/test-results/20260830T174405Z-p1262123`; rollback contract passed at `.cartulary/test-results/20260830T174409Z-p1266281`; Imports entity-owner application passed at `.cartulary/test-results/20260830T174415Z-p1266595` | Intentional repository-internal correction for Identity from generic `record:`/`record` to `identity:`/`identity`, owner changed fields, and canonical before-version handling; routes, rows, create/reuse/update result codes, projection refresh, rollback, and replay remain unchanged; no migration | Start WS-16; rollback is the Identity-only helper/branch change while retaining the already-checkpointed Host path. Residual risk is Tasks/Decisions and the now-unused aggregate surface pending WS-17. |
| `2026-08-30T13:48:30-04:00` | `/root` WS-16 | Task and Decision imports independently derive record revision/publication effects while retaining their owner-specific link sequence | Changed `tasksdecisions/import_create.go`, `publication.go`, and the exported-surface lock; replaced the embedded aggregate port with explicit Revisions methods, preserved one contiguous sequence reservation for record plus link mutations, emitted record/live/public effects before ordered link mutation rows, and kept Decision's zero-link path | `make format`; Tasks/Decisions import atomicity service-backed row; full Tasks/Decisions unit/frontend slice; exact production-caller symbol audit; `git diff --check`; `make lint-markdown` | Format passed at `.cartulary/test-results/20260830T174635Z-p1286397`; owner-reference and atomicity row passed at `.cartulary/test-results/20260830T174638Z-p1290552`; all 21 unit/frontend rows passed at `.cartulary/test-results/20260830T174736Z-p1308553` | Repository-internal Go interface simplification only; `record:` identity, create response, row version, record-before-link sequence, ordered links, public ordinal, rollback at later link failure, and Imports replay behavior are unchanged; no migration | Start WS-17; rollback is the three-file Tasks/Decisions cutover. Residual RB-005 risk is only dead aggregate code/tests and reintroduction prevention. |
| `2026-08-30T13:55:38-04:00` | `/root` WS-17 | The combined historical/live finalizer, aggregate port types, generic row/version/fact helpers, aliases, and temporary characterization fixture are absent; RB-005 is closed | Deleted `ownerfacade/finalize.go` and `finalizer_characterization_test.go`; removed the temporary caller from ownerfacade characterization; added a fail-closed source-owner token rule and Imports source-tree assertion; consolidated Assessment's owner helper into its allowed import seam; updated current tracker inventory/findings | `make format`; focused Imports boundary/ownerfacade rows; `make backend-module-boundary-check`; `make harness-contract`; `make json-shape-check`; `make generate-drift`; `make generated-artifact-policy-check`; task guide/explanation for Revisions, Projections, and Collaboration; exact source symbol audit; `git diff --check`; `make test-fast`; `make lint-markdown` | Final format `.cartulary/test-results/20260830T175121Z-p1359862`; focused Imports `.cartulary/test-results/20260830T175036Z-p1357624`; boundary `.cartulary/test-results/20260830T175125Z-p1364045`; harness contract `.cartulary/test-results/20260830T175127Z-p1364450`; JSON shape `.cartulary/test-results/20260830T175141Z-p1365017`; drift `.cartulary/test-results/20260830T175219Z-p1366939`; generated policy `.cartulary/test-results/20260830T175227Z-p1369901`; all 441 fast units `.cartulary/test-results/20260830T175232Z-p1370455`. The first boundary run failed at `.cartulary/test-results/20260830T175045Z-p1359029` because the new Assessment helper file was outside its exact allowed Collaboration seam; consolidating it into `import_create.go` corrected the directly related boundary failure. | Breaking repository-internal aggregate API removal with no wrapper or data migration; external routes, schemas, jobs, Imports session/job finalization, and valid client behavior are unchanged | Start WS-18. Rollback requires restoring the entire old aggregate and every owner caller, never a forwarding alias. Residual implementation risk is limited to broad validation; RB-005 structural risk is closed. |
| `2026-08-30T14:20:54-04:00` | `/root` WS-18 | All six remediation gaps and all nineteen serial workstreams are complete; the final product tree passes the full validation ladder | Inspected the final specification, implementation, authored harness, generated-index, test, and tracker changes; changed `internal/modules/tabularingest/tabularingest.go` only to use the Staticcheck-required named warning conversion, and changed this tracker for result-only closure; `docs/domain.md` remained unchanged because no vocabulary conflict was found | Imports task guide and owner explanation; full Imports service-backed slice; `make browser-e2e-webserver-backed`; `make test-fast`; `make agent-finalize`; `make check`; `make lint-go`; Incident Bundles task guide and owner explanation; exact service-backed failure rerun; refreshed `make test-fast`; refreshed `make agent-finalize`; final `make check`; `git diff --check`; `make lint-markdown` | Imports passed 14/14 at `.cartulary/test-results/20260830T175638Z-p1417244`; browser-backed passed 60/60 at `.cartulary/test-results/20260830T175757Z-p1462105`; initial final fast/finalize passed at `.cartulary/test-results/20260830T180402Z-p1519105` and `.cartulary/test-results/20260830T180418Z-p1519836`. The first check failed 669/671 at `.cartulary/test-results/20260830T180436Z-p1522789`: one directly related Staticcheck S1016 finding, reproduced at `.cartulary/test-results/20260830T181200Z-p1663740`, was corrected by the behavior-neutral conversion; one unrelated Incident Bundles assertion passed its exact 3-unit owner rerun at `.cartulary/test-results/20260830T181303Z-p1679499`. Go format, vet, and Staticcheck then passed at `.cartulary/test-results/20260830T181236Z-p1664939`, `.cartulary/test-results/20260830T181239Z-p1668851`, and `.cartulary/test-results/20260830T181247Z-p1677521`; refreshed fast/finalize passed at `.cartulary/test-results/20260830T181358Z-p1697214` and `.cartulary/test-results/20260830T181452Z-p1706292`; final `make check` passed 671/671 at `.cartulary/test-results/20260830T181512Z-p1709285`; the mandatory WS-18 Markdown checkpoint passed at `.cartulary/test-results/20260830T182205Z-p1826723`. Retained-run maintenance was skipped in both finalization runs because `RESULTS_DIR` was unset. | No route, method, operation ID, envelope, database schema, job identity, WebSocket message, or frontend workflow migration; invalid cookie-authenticated mutations now intentionally fail closed; major-4 Network Flow and the aggregate Go API are intentionally unsupported. No data translator/backfill was introduced; reset pre-production local state if old generic import-mutation history must be discarded. | No open blocker or known remediation gap. Rollback is by completed workstream boundary, never by adding a major-4 shim, aggregate alias, or second mapping path. Residual risk is limited to the intentional compatibility removals and ordinary future-change risk covered by the new owner, kernel, route-catalog, and boundary checks. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| `2026-08-30T09:51:22-04:00` | `/root` planning discovery | Four member mutations omit state-changing CSRF mode; create and regions enable it; mapping preview is read-only | Inspected Imports handlers, platform HTTP auth and Core 04 CSRF owner; touched none | `rg`, `sed` | Owner-required correction and missing negative coverage identified | RB-002 | Add SL-00 security matrix, then request authority for SL-01. |
| `2026-08-30T09:57:55-04:00` | `/root` documentation implementation | Security mismatch classified separately from structural refactor | Touched tracker only | Documentation patch | `must_fix` finding and behavior-change marker recorded | Production correction not authorized | Do not combine SL-01 with handler movement. |
| `2026-08-30T11:03:42-04:00` | `/root` NLSpec-style tracker revision | Six state-changing and five read-only operations, authentication modes, precedence, error, and no-effect behavior are exhaustive | Inspected live handler state-changing flags, OpenAPI operation IDs, and Core security clauses; touched tracker only | `jq`, `rg`, `sed`, documentation patches | RB-002 is decision-complete and has ten binary criteria | Correction changes invalid-cookie outcomes and remains unauthorized | Run SL-00 security characterization, then authorize isolated SL-01 owner/code correction. |
| `2026-08-30T11:52:41-04:00` | `/root` WS-02 | The exhaustive Imports security matrix is adopted by its route and security owners | Changed Core 01 Table 17.2-A and Core 04 §9.2; audited the Imports OpenAPI owner input without changing it | OpenAPI `jq` audit, `git diff --check`, documentation patches, `make lint-markdown` | Core 01 now owns the exact operation IDs and six-mutating/five-read-only classification; Core 04 owns credential mode, gate precedence, `403 csrf_verification_failed`, role/body ordering, and exhaustive no-effect behavior; the existing OpenAPI security rows already match; Markdown passed at `.cartulary/test-results/20260830T155219Z-p375425` | None; no product behavior or migration changed in this specification-only slice | Start WS-03; rollback is the atomic Core 01/Core 04 amendment; residual risk is the live handler's four missing state-changing flags until WS-03 closes. |
| `2026-08-30T12:02:47-04:00` | `/root` WS-03 | All Imports routes derive method, operation ID, and cookie-CSRF classification from one closed catalog | Changed `http_handlers.go`, characterization and security integration tests, the authored Imports test-family row, generated execution-topology index through `make generate`, and this tracker | Imports task guide/explanation; focused unit and service-backed rows; full Imports owner slice; `make format`; `make json-shape-check`; `make generate`; `make generate-drift`; `make lint-markdown` | Catalog unit passed at `.cartulary/test-results/20260830T155907Z-p385070`; security matrix passed at `.cartulary/test-results/20260830T155919Z-p385820`; full Imports owner passed at `.cartulary/test-results/20260830T160013Z-p403351`; initial JSON/drift roots `.cartulary/test-results/20260830T160134Z-p448293` and `.cartulary/test-results/20260830T160134Z-p448283` correctly reported stale generated topology, `make generate` passed at `.cartulary/test-results/20260830T160208Z-p452213`, then JSON and drift passed at `.cartulary/test-results/20260830T160228Z-p455156` and `.cartulary/test-results/20260830T160228Z-p455155` | None; missing/invalid cookie CSRF on mapping, select, skip, and apply now changes to exact `403 csrf_verification_failed`; bearer, read-only, valid-cookie, routes, envelopes, and durable bytes are unchanged; no migration | Start WS-04; rollback is the WS-03 code/test/authored-manifest/generated-index set; residual structural risk is RB-006 until transport extraction completes. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| `2026-08-30T09:51:22-04:00` | `/root` planning discovery | Owner contradiction, CSRF gap, mapping-engine mismatch, stale boundary path and deferred finalizer home are known | Inspected exact evidence named by RB-001 through RB-005; touched none | Read/search and Make discovery commands above | Planning unknowns are explicit; no behavior guessed | RB-001 through RB-005 | Resolve only the blocker required by the selected later slice. |
| `2026-08-30T09:57:55-04:00` | `/root` documentation implementation | Tracker is sufficient for restart without rediscovery | Touched tracker only | Documentation patch | WF-00 through WF-07 planning complete; WF-08 implementation validation remains future work | RB-001 is specification-blocking for Network Flow contract redesign | Safe next task: authorize and execute SL-00 characterization only. |
| `2026-08-30T11:03:42-04:00` | `/root` NLSpec-style tracker revision | RB-002 through RB-005 are design-complete but implementation-open; RB-001 remains blocked; no implementation row is active | Inspected all evidence referenced by the revised RB entries; touched tracker only | Read/search and documentation commands above | Tracks A-C, slice dependencies, stop rules, binary closure, and owner boundaries are sufficient for handoff | RB-001 owner contradiction; later authority required for every owner/code/harness change | Start SL-00 for Generic Imports or SL-04A for finalizer characterization; do not start SL-NF-01 without owner authority. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| -- | ------------------- | -------------- | ---------------------------- | -------------- |
| RB-001 | Network Flow Activity 5.0.0 pinned Extensions 0.7.1 and Core 03 advertised client major 4, while Extensions 0.10.0 and the remainder declared major 5/state 4 | The corpus could not safely select one dependency and client line until amended atomically | WS-01 published Network Flow 5.0.1, amended Core 03 to major 5, pinned the final Extensions 0.10.0 digest and exact imported interfaces, and confirmed major-5/state-4 projections | DONE; no major-4 adapter or data migration |
| RB-002 | Four state-changing member operations omitted cookie-CSRF admission | Structural preservation could not freeze owner-nonconformant security behavior; correction changes invalid-cookie outcomes | Core 01 Table 17.2-A, Core 04 §9.2, closed route catalog, full cookie/bearer/read-only matrix, precedence tests, and durable/publication no-effect assertions | DONE; valid clients unchanged, invalid cookie mutations now fail with exact `403 csrf_verification_failed`, no migration |
| RB-003 | File and clipboard previously used divergent mapping engines | Divergence could change stored fingerprints, replay equality, provenance, owner requests, or clipboard semantics | WS-06 fixed current bytes, WS-07 adopted the exact contract, and WS-08 cut both adapters over to one kernel with full parity and boundary evidence | DONE; no cross-lifecycle fingerprint or analytical mapping generalization |
| RB-004 | Rule `retired-timeline-facade-calls` named removed `internal/modules/imports/routes.go` and could pass after scanning no live Imports transport | The static check could miss the current transport adapter or pass after scanning zero files | Testing Harness boundary schema v3, exact-file-set implementation, negative fixtures, authored-policy migration, and public checks | DONE; WS-05 |
| RB-005 | The former aggregate helper violated the adopted independent-port boundary | Preserving it would couple private revision semantics to public disclosure and risk atomicity/replay defects | Complete caller matrix, the already-adopted Collaboration decision, serial per-owner cutovers, aggregate removal, negative boundary, and broad fast evidence | DONE; no aggregate symbols, aliases, or replacement subsystem remain |
| RB-006 | Imports HTTP transport and application orchestration were mixed in `http_handlers.go` | Security admission, serialization, transactions, and lifecycle changes were unnecessarily coupled | Private typed application methods, exact route-catalog parity, application tests, service-backed Imports evidence, and browser workflow evidence | DONE; WS-04 |

### RB-001 binary acceptance

| Criterion | Required evidence |
| --- | --- |
| No current normative text names Extensions 0.7.1 as Network Flow's dependency | Exact owner-text audit |
| No current normative or machine client-support input assigns Network Flow major 4 | Owner and authored-input audit |
| Network Flow resolves the exact adopted Extensions 0.10.0 bytes and current locators | Dependency resolver/owner evidence |
| Discovery, descriptors, implementation bindings, and client support all report major 5 | Registry, binding, generated, and client tests |
| Major-4 discovery produces workspace omission, never translation | Negative client/discovery fixture |
| Claimed major 5 exposes exactly `network_analysis`; unclaimed major 5 does not | Browser/client availability evidence |
| Existing Network Flow import preview and apply bytes remain unchanged unless separately versioned | Golden request/response parity |
| Every affected owner slice, registry check, drift check, and client-support test passes | Retained Make-owned run evidence |
| The owner bundle is adopted atomically | Commit/owner review with no intermediate coordinated-current claim |

### RB-002 binary acceptance

| Criterion | Required evidence |
| --- | --- |
| All six state-changing cookie operations reject missing and invalid CSRF with the exact common error | Six-operation negative matrix |
| All six accept valid CSRF when later gates succeed | Six-operation positive matrix |
| All six preserve bearer behavior without a CSRF token | Bearer matrix |
| Four GET operations and mapping preview require no CSRF | Five read-only cases |
| Unauthenticated failure precedes CSRF | Precedence test for every route family |
| CSRF failure precedes path, visibility, role, body, lifecycle, and idempotency diagnostics | Causal precedence tests |
| A visible underprivileged caller with valid CSRF receives role denial before malformed-body detail | Role/body test |
| An authorized caller with valid CSRF can receive the route-owned malformed-body error | Positive admission/negative body test |
| Every CSRF rejection leaves all durable and publication state unchanged | Database, job, revision, projection, Collaboration, idempotency, and audit assertions |
| Paths, methods, operation IDs, success envelopes, and valid-client behavior remain unchanged | OpenAPI/generated parity and existing integration/browser evidence |

### RB-003 binary acceptance

| Criterion | Required evidence |
| --- | --- |
| File and clipboard view targets invoke one shared mapping kernel | Static call graph and positive cross-path tests |
| No parser object, workbook object, grid object, HTTP DTO, file identity, actor, transaction, or consumer port crosses the kernel | Interface review and negative import/type evidence |
| Equivalent shared inputs produce byte-identical row plans and owner requests | Golden cross-path matrix |
| Existing approved-mapping bytes and fingerprints remain byte-identical | Before/after file goldens |
| Clipboard creates no file session, unit, locator, approval, or file mapping fingerprint | Durable-state negative assertions |
| Common provenance matches and source-specific provenance remains distinct | Provenance fixture matrix |
| Every transform, delimiter, empty policy, entity mode, and admitted unknown policy is covered | Closed-registry fixture accounting |
| Invalid plans fail before an owner request or durable effect | Failure-precedence and no-effect tests |
| Network Flow exact mapping remains target-owned | Static boundary and Network Flow contract tests |
| Any incompatible fixture blocks extraction instead of changing behavior silently | Slice stop record with the exact incompatible bytes, if encountered |

### RB-004 binary acceptance

| Criterion | Required evidence |
| --- | --- |
| The live handler and final post-SL-02 transport set are inspected | Exact path inventory |
| A deliberate forbidden call in a live transport file fails the public check | Negative token fixture |
| Missing and stale declared paths fail | Negative path fixtures |
| Duplicate, escaping, absolute, backslash, dot-segment, symlink, glob, and non-regular paths fail | Path-integrity fixture matrix |
| Every declared path resolves exactly once | Positive resolver evidence |
| An unlisted new route-binding file fails or requires explicit policy update | Omitted-handler fixture |
| Catalog selectors still resolve exactly once and generated drift is clean when affected | Catalog/drift evidence |
| `make backend-module-boundary-check` and the applicable broader gate pass | Retained Make-owned run roots |

### RB-005 binary acceptance

| Criterion | Required evidence |
| --- | --- |
| An adopted architecture decision states that no combined revision/publication finalizer is current | Adopted owner text and traceability |
| Every production caller and operation family is characterized | Complete seven-row Section 4.6 matrix |
| `FinalizeCommand` and every equivalent aggregate are removed | Exact symbol/source audit |
| No forwarding alias or compatibility wrapper remains | Negative static/runtime evidence |
| Each source owner derives Revisions facts and Collaboration effects independently | Per-owner causal tests and code review |
| Revisions, Projections, and Collaboration use distinct narrow ports | Interface/import graph evidence |
| Revisions and Collaboration neither import, call, derive, nor translate one another's representation | Negative boundary evidence |
| Projection state is not used as source truth | Failure/semantic tests and source audit |
| Each source-owner transaction is atomic under failure at every port | Per-owner database failure injection |
| Exact replay creates no duplicate revision, projection effect, or Collaboration intent | Replay assertions |
| Imports session/job finalization remains unchanged and separately owned | Imports job/recovery evidence |
| Boundary checks reject a combined operation or broad publication capability | Negative static fixtures |
| The final package graph is acyclic and no new aggregate finalization subsystem exists | Package graph and boundary check |

### RB-006 binary acceptance

| Criterion | Required evidence |
| --- | --- |
| `http_handlers.go` is the only Imports route-binding file | Exact token-based route-binding discovery and boundary policy |
| The transport adapter performs only route binding, admission, decoding, delegation, encoding, and safe error translation | Thin-transport review and static dependency checks |
| Session, read, mapping, selection, region, and apply orchestration use private typed application methods | Application interface and focused unit evidence |
| All 11 paths, methods, operation IDs, DTOs, errors, and success envelopes are unchanged | Exact route-catalog and generated/OpenAPI parity tests |
| Security and diagnostic precedence remain as adopted in WS-02/WS-03 | Full cookie, bearer, role, visibility, and body matrix |
| Persistence, transactions, jobs, owner dispatch, and replay semantics remain unchanged | Service-backed Imports rows and failure/replay evidence |
| Frontend and browser workflows remain equivalent | Frontend tests and `make browser-e2e-webserver-backed` |

## 12. Binary Completion Criteria

### 12.1 Current implementation completion

| Criterion | Status | Evidence |
| --- | --- | --- |
| WS-00 tracker rebaseline is complete | PASS | Sections 1, 4.1, 6, 7, 9, 10, and 11; Markdown run `.cartulary/test-results/20260830T153726Z-p244228` |
| RB-001 owner contradiction is remediated | PASS | Network Flow 5.0.1, Core 03 major 5, Extensions 0.10.0 digest/interface pin, owner runs, and drift runs recorded in Section 10 |
| RB-002 exhaustive security contract and implementation pass | PASS | WS-02 owner amendment and WS-03 route catalog/security evidence recorded in Section 10 |
| RB-003 shared mapping kernel preserves compatibility bytes | PASS | Pure kernel and two adapter calls; exact file/clipboard goldens, cross-path plans, owner requests, parser hostility, cancellation, boundary, full Imports, Timeline, and Entities evidence recorded in Section 10 |
| RB-004 boundary exact-file accounting fails closed | PASS | V3 owner/schema/checker, full missing/stale/path/symlink/non-regular/omission fixtures, public boundary, harness-contract, JSON-shape, and drift evidence recorded in Section 10 |
| RB-005 aggregate finalization is removed after all owner cutovers | PASS | WS-09 through WS-17; negative boundary, affected owner evidence, and 441-unit fast run recorded in Section 10 |
| RB-006 transport and application concerns are separated | PASS | Private typed application operations, thin-adapter static checks, exact route parity, full Imports owner slice, and 60-unit browser-backed run recorded in Section 10 |
| Complete validation and handoff are recorded | PASS | WS-18; full Imports service-backed 14/14, browser-backed 60/60, refreshed fast 441/441, finalization 1/1, and repository check 671/671 evidence recorded in Section 10 |

### 12.2 Historical planning completion

| Criterion | Status | Evidence |
| --- | --- | --- |
| Every file in `internal/modules/imports` is inventoried or explicitly out of scope | PASS | Section 2 contains one row for each of the 36 Go files; none is omitted. |
| Every discovered public contract risk has an owner, interface, default/omission posture, and test posture | PASS | Section 4 maps HTTP, mapping, jobs, transactions, owners, Network Flow, frontend, generated, recovery and harness surfaces; Sections 4.2-4.6 close the five gap designs. |
| Every proposed workflow has dependencies, an execution track, and an exit criterion | PASS | Section 6 defines WF-00 through WF-08 and Tracks A-C with previous/subsequent workflows and handoff checkpoints. |
| Every implementation slice names authority, exact change, risks, tests, rollback, stop conditions, and completion | PASS | Section 7 defines SL-00 through SL-06, SL-NF-01, and SL-04 sub-slices; behavior corrections and owner adoptions are explicitly `requires later authorization` or blocked. |
| Validation commands are discovered or have an explicit discovery rule | PASS | Section 8 uses Make-owned Imports commands, names all affected source-owner slices, and requires owner explanation before any service-backed variant is selected. |
| Each RB item has testable binary acceptance criteria | PASS | Section 11 traces RB-001 through RB-005 to owner, implementation, byte, failure-injection, boundary, drift, and retained-run evidence. |
| Contradictions are marked `BLOCKED: owner contradiction` and no side is locally selected | PASS | RB-001, SL-NF-01, and the corresponding contract/coupling entries preserve the block until atomic owner adoption. |
| Repository/framework mismatches are recorded as planning findings | PASS | Sections 3 and 5 record live decomposition, separate `tabularingest`, incomplete CSRF admission, the prohibited combined finalizer, and stale `routes.go` accounting. |
| Handoff sections preserve history and are current enough for another agent to continue without rediscovery | PASS | Section 10 retains prior discovery/authoring rows and appends the NLSpec-style revision session with exact commands and outcomes. |
| Planning completion is not confused with owner or implementation closure | PASS | Section 4.1 distinguishes decision-complete from closed; Section 9 leaves implementation rows TODO/BLOCKED/DROPPED and has no IN_PROGRESS row. |

The historical table remains evidence of planning completeness only. Section
12.1 records the completed implementation state: WS-00 through WS-18 and
RB-001 through RB-006 are closed under the authority recorded in Section 1.
