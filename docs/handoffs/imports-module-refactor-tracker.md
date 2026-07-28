# imports Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

### 1.1 Target and change boundary

- **Target path:** `internal/modules/imports`
- **Target label:** `imports`
- **Output path:** `docs/handoffs/imports-module-refactor-tracker.md`
- **Status:** Planning and documentation only.
- **Allowed change in this revision:** This tracker file only.
- **Non-goals:** No production refactor; no changes to tests, adopted owners, contracts,
  generated artifacts, package configuration, migrations, harness inputs, or dependency locks.
- **Implementation authorization:** Every production, test, owner, contract, generator, migration,
  or harness change described below requires a later explicitly authorized task.
- **Behavior posture:** A structural slice MUST preserve current observable behavior. A
  conformance correction MUST be isolated, authorized, migrated when necessary, and verified
  against its adopted owner.

### 1.2 Authority and normative language

This tracker uses NLSpec-style normative language to make the future refactor handoff
decision-complete. It is not an adopted behavior owner and MUST NOT supersede an adopted Core or
subsystem NLSpec.

- **MUST** and **MUST NOT** state requirements for an authorized future refactor slice after its
  owner and execution gates pass.
- **SHOULD** states a preferred choice that may be replaced only when the implementation handoff
  records an equivalent owner-conformant result and its verification.
- **MAY** grants intentional implementation freedom whose alternatives do not change caller-visible
  behavior or interoperability.
- **RESOLVED_IN_TRACKER** means the planning decision is complete. It does not mean the relevant
  owner amendment is adopted, code is implemented, or conformance evidence passes.
- A proposed requirement that conflicts with current adopted owner text MUST remain
  non-implementable until the coordinated owner-repair gate passes.

The source hierarchy is:

1. Adopted subsystem NLSpecs for their named scopes.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. Domain vocabulary and implementation-support guides.
5. Current repository code and tests.
6. Prior plans, handoffs, analysis notes, and the planning framework as evidence only.

Core 05 is not applicable because this tracker publishes no timed or fixture-sensitive claim.
`temp/analysis-notes.md` supplies recommendations and rationale only; local adopted owners supply
every normative behavior reference. External-source freshness prose and `sandbox:` links from the
notes are not imported into this tracker.

### 1.3 Documents and repository evidence inspected

Owner and support documents:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `docs/domain.md`
- `docs/research/nlspec-spec.md`
- Core 00 through Core 04 under `docs/spec/`
- `docs/network-flow-activity-nlspec.md`
- `docs/extension-subsystem-nlspec.md`
- `docs/testing-harness-nlspec.md`
- `contracts/verification/owners/module.imports.json`
- `tools/test_families/module.imports.json`
- `tools/backend_module_boundaries.json`

Repository evidence:

- Every tracked file under `internal/modules/imports`.
- The shared `internal/modules/tabularingest` package.
- Server and extension composition under `internal/app/server` and
  `internal/app/extensionassembly`.
- Network Flow import-facade and transaction-participant code.
- Import-create facades in artifacts, assessments, entities/hostidentity, evidence, indicators,
  parties, and tasksdecisions.
- Import migrations `db/migrations/00018_imports.sql` and
  `db/migrations/00028_import_source_streams_and_targets.sql`.
- The authored imports OpenAPI owner and generated Go and TypeScript consumer locations.
- The web import coordinator, Import Assistant, Network Flow import controller, unit tests, and
  browser tests.
- Verification-owner, test-family, test-catalog, and module-boundary mappings.

The inspected repository baseline is clean `main` at
`effce7c1fb34e7e6eabe7c2e80017d38c539f123`, apart from this staged tracker.

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/imports/.gitkeep` | Historical directory marker; no runtime responsibility | None | None | None | None | None | imports | low | Obsolete in a non-empty package; removal is cleanup only and is deferred. |
| `internal/modules/imports/api.go` | Import request/response DTOs, profile and state vocabulary, strict decoding, mapping validation and fingerprinting, normalization, and API errors | Import session, unit, mapping, preview, apply, warning, error, and profile types and constants | Routes, Store, tests, server composition, Network Flow facade | Standard library plus import contract/value helpers | Import integration and frontend contract-adapter tests indirectly | Authored imports OpenAPI and generated protocol models mirror this surface | imports | high | Public HTTP envelope and error compatibility MUST be frozen before movement. |
| `internal/modules/imports/boundary_guard_test.go` | Source-level dependency and ownership guards for the imports package | Test-only package surface | Go test runner and verification catalog when routed | Filesystem/source inspection | Self | Verification owner and family mappings | imports test ownership | medium | Current harness rows do not explicitly account for every boundary test. |
| `internal/modules/imports/boundary_test.go` | Guards against workbook/projection ownership leakage and selected peer-owner coupling | Test-only package surface | Go test runner | Filesystem/source inspection | Self | Verification owner and family mappings | imports test ownership | medium | Source-text guards are boundary evidence, not runtime-conformance proof. |
| `internal/modules/imports/extension_facade.go` | Analytical import preview/apply contract, request/result types, facade registry, and dependency override lookup | Extension facade interfaces, preview/apply request and result DTOs, registry construction and errors | Network Flow module, routes, tests, server dependency override | `internal/modules/extensions` dependency set and shared import values | Network Flow and imports integration tests | Core import profile and Network Flow owner contract | imports orchestration with target-owner implementations | critical | Current Core and Network Flow exact facade shapes conflict; the planning resolution is in IRT-REQ-003 through IRT-REQ-006. |
| `internal/modules/imports/imports_integration_test.go` | Service-backed HTTP, session, mapping, replay, bounded XLSX, generic-owner, Network Flow, evidence, journal, and failure-path coverage | Test-only fixtures and cases | Go test runner and module.imports verification rows | Server test support, PostgreSQL, object store, HTTP, module stores | Self | Verification owner and test-family rows | imports test ownership | high | Coverage does not exercise every registered owner or all normative state and authorization cases. |
| `internal/modules/imports/job_finalization.go` | Narrow port for publishing successful import-job terminal state | `JobSuccessFinalizer` interface | Route/job execution and extension assembly adapter | Job finalization owner through injected application adapter | Imports integration tests indirectly | Job and import profile contracts | imports application coordination | high | Finalization MUST become a recoverable derivation separate from owner-resource creation. |
| `internal/modules/imports/limits.go` | Imports configuration DTO and archive-limit alias | `Limits`, `ArchiveLimits` | Server module settings, parser/session code | Platform archive policy | XLSX and resource-limit integration coverage | Core resource-limit contract | imports policy projection; platform owns archive mechanics | medium | Platform dependency is a typed policy adapter, not source-domain logic. |
| `internal/modules/imports/owner_apply.go` | Generic target dispatch, transaction opening, change-set creation, source-row mapping, source-owner validation, peer-store construction, owner mutation, projection/revision coordination, and apply journal writes | Primarily unexported apply coordinator; contributes Store apply behavior | Routes/jobs, Store, integration tests | Revisions, collaboration intents, Timeline, seven peer source-owner stores, and direct SQL reference checks | Timeline, host identity, evidence, replay, and failure integration cases | Owner-create facade and Core import dispatcher contract | imports dispatcher plus source-owner facades and application composition | critical | Cross-owner validation and concrete store construction MUST move behind owner facades and application composition. |
| `internal/modules/imports/ownerfacade/finalize.go` | Shared normalized field-plan finalization and revision metadata used by source-owner creates | Finalization request/result types and helper functions | Seven peer source-owner import-create implementations | Revision/value helpers | Peer module tests and imports integration tests indirectly | Internal owner-create contract | source-owner-neutral import facade | high | It MAY remain a narrow contract after owner semantics are characterized. |
| `internal/modules/imports/ownerfacade/owner_create.go` | Owner-create request/result DTOs, normalized scalar conversion, action/value vocabulary, and validation helpers | `CreateImportRowRequest`, result/value types, scalar normalization and related constants/helpers | Artifacts, assessments, host identity, evidence, indicators, parties, tasksdecisions, and imports dispatcher | Standard/value parsing helpers | Peer tests and imports integration tests indirectly | Core owner-create facade contract | imports shared facade with source-owner implementations | critical | `use_null` is a live internal mismatch against the accepted and persisted `write_null` token. |
| `internal/modules/imports/revision_append_port.go` | Narrow adapter for appending revision contributions during import-owned coordination | Internal revision-appender wrapper surface | Owner apply and route/job assembly | Revisions module | Imports integration tests indirectly | Revision/change-set contracts | revisions owner behind imports port | high | Revision order and idempotency MUST remain owner-defined. |
| `internal/modules/imports/routes.go` | Route registration and handlers, authentication/CSRF/role checks, job registration and recovery, source discovery, mapping preview/approval, unit selection/skip/apply, Timeline dispatch, transformations, path parsing, and HTTP error mapping | `RegisterRoutes`, registration options, service composition surface | Server runtime | HTTP/platform auth/jobs/object storage, Store, tabular ingest, revisions, Timeline, extension facade | Imports integration tests and web browser tests | Ten authored OpenAPI operations and generated route/model consumers | imports HTTP adapter plus application service/job coordinator | critical | Admission authorization exists; transaction-current authorization immediately before mutation is not visible. |
| `internal/modules/imports/source_streams.go` | Import-owned opaque source-stream references, capability loading, and content-digest validation | Source capability/stream types and Store methods consumed by analytical facades | Network Flow import facade and transaction participants | Import Store and hashing/value helpers | Network Flow behavior and imports integration tests | Analytical import facade contract | imports | high | Capability binding MUST satisfy IRT-REQ-004 and IRT-REQ-007 before interface movement. |
| `internal/modules/imports/store.go` | Direct SQL persistence for import sessions, units, source rows, previews, idempotency, admission/state transitions, source streams, and apply journal | `Store`, constructor, session/unit/query/mutation methods, sentinel errors | Routes, owner apply, Network Flow through a restricted interface, tests, server assembly | PostgreSQL and import-owned schema | Imports and Network Flow tests | Import database schema and HTTP resources | imports persistence adapter | critical | Direct SQL is intentional for import-owned state; cross-owner SQL MUST NOT migrate here. |
| `internal/modules/imports/targets.go` | Registry of supported Core view-schema targets and Network Flow analytical targets | Target registry/lookup surface | Routes, mapping validation, owner apply, tests | View-schema identifiers and target metadata | Target-registry and owner-apply integration cases | View contracts and imports OpenAPI target fields | imports contract-owner mapping | high | Backend registry and frontend importability IDs duplicate one Core-owned fact. |
| `internal/modules/imports/xlsx.go` | Bounded ZIP/OpenXML XLSX parsing, shared strings, worksheet used ranges, cell kinds, and archive limits | Internal XLSX discovery/row-loading surface | Route discovery and preview/application flow | ZIP/XML parsing and archive policy | Bounded used-range integration test | Core import profile and limit contract | imports source adapter using shared tabular-ingest semantics | high | Tables, named ranges, operator regions, and presentation neutrality lack implementation evidence. |

The filesystem also contains an empty, untracked
`internal/modules/imports/tabularingest` directory. Git does not track empty directories. The live
shared parser package is `internal/modules/tabularingest`; no module move may be inferred from the
empty directory.

## 3. Module Boundary Diagnosis

The dedicated imports boundary is legitimate. Core 01, Core 03, and the domain vocabulary assign it
source adapters, sessions, units, mappings, warnings, provenance, target dispatch, and import
orchestration. That valid boundary does not validate every responsibility currently placed inside
the package.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Normative disposition |
| --- | --- | --- | --- | --- | --- |
| Import session, unit, mapping, source-stream, warning, and journal state | `api.go`, `store.go`, `source_streams.go` | imports | keep | Core 01/03 and imports migrations | Imports MUST retain this state behind its facade. |
| CSV and XLSX source adaptation and bounded discovery | `routes.go`, `xlsx.go`, shared tabular ingest | imports using owner-neutral tabular-ingest primitives | split | Core 01/03 and live call graph | Parsing/discovery MUST separate from HTTP and jobs without changing source semantics. |
| HTTP request parsing and response/error mapping | `routes.go` | imports transport adapter | split | Authored OpenAPI and route registration | All ten route shapes and envelopes MUST remain stable. |
| Job registration, recovery, execution, and terminal publication | `routes.go`, `job_finalization.go` | imports coordinator plus platform job shell | split | Core job rules and extension assembly | Finalization MUST follow the two-level model in IRT-REQ-005 and IRT-REQ-006. |
| Admission and apply-time authorization | `routes.go`, transaction paths | imports service using platform auth | split | Core 04 | Transaction-current authorization MUST be added only through an authorized correction. |
| Generic view-schema target dispatch | `targets.go`, `owner_apply.go` | imports dispatcher | keep | Core 01 dispatcher requirements | Dispatcher MUST select exactly one owner facade and MUST NOT know owner internals. |
| Concrete peer-store construction | `owner_apply.go` | application composition root | move | Application facade doctrine and live constructors | Construction MUST move to application composition. |
| Source-owner validation, defaults, references, and create rules | `owner_apply.go`, peer import-create files | each source-owner module | move | Core 01 owner-facade requirements | Imports MUST delegate these semantics to the selected source owner. |
| Timeline import mutation | Special apply path | Timeline behind the owner-create facade | move | Dispatcher rule and Timeline ownership | Timeline MUST use the same registry selection boundary while preserving effects. |
| Revision/change-set coordination | `owner_apply.go`, revision port | revisions owner behind coordination port | split | Core revision semantics | Imports MAY coordinate but MUST NOT redefine revision semantics. |
| Projection refresh and collaboration effects | Peer facades | corresponding source owner | move | Core owner-facade rule and live peer stores | Owner effects MUST remain behind the owner facade. |
| Network Flow normalization and table mutation | Network Flow implementation | Network Flow | keep outside imports | Adopted Network Flow NLSpec | Imports MUST retain only orchestration and Core capabilities. |
| Frontend import workflow state | Web coordinator and Import Assistant | web imports controller/feature | keep | Live frontend flow | Frontend MUST consume the generated semantic target projection after GATE-04. |
| Grid-vendor integration | Not found | grid adapter if later discovered | defer | Targeted search | No action is authorized on current evidence. |
| Saved views | Not directly owned or mutated | saved-view owner | defer | Live code and contract search | View-schema targeting MUST NOT be described as saved-view mutation. |
| Test-util mutation paths | No target-specific path found | test-util owner if later discovered | defer | Integration tests use reusable app/HTTP support | No test-util move is planned. |

Overall classification: **legitimate public imports facade over a mixed-responsibility package**,
including import/view orchestration, transport-adjacent and persistence-adjacent adapters, and
mutation coordination. The refactor MUST retain the imports boundary and make its dependencies
shallower; it MUST NOT delete or rename the boundary merely because its internals are mixed.

## 4. Public Contract and Behavior Freeze Map

### 4.1 Normative refactor requirements

| ID | Requirement | Current owner or required owner repair |
| --- | --- | --- |
| IRT-REQ-001 | The refactor MUST preserve the authority hierarchy in §1.2 and MUST NOT implement a proposed owner repair before coordinated adoption. | Repository procedure and Core 00 |
| IRT-REQ-002 | Structural slices MUST preserve all ten public import operations, operation IDs, request and response envelopes, WebSocket non-surface, and existing generated public types. | Core 01 and authored imports OpenAPI |
| IRT-REQ-003 | Core MUST own analytical dispatch, target selection, lifecycle, Core-issued capabilities, authorization, idempotency, cancellation, unit/session outcomes, and job publication. Each target NLSpec MUST exclusively own exact target request/result members, mapping, diagnostics, target errors, and owner resource mutation. | Proposed coordinated Core 00/Core 01/target-owner repair |
| IRT-REQ-004 | Every analytical target MUST register one `cartulary.imports.analytical_facade_binding.v1`; the binding MUST resolve each referenced schema exactly once and convey the semantic slots in §4.4. | Proposed Core 01 repair and machine projection |
| IRT-REQ-005 | One import-unit commit MUST atomically contain the selected owner effects, owner-required audit/revision/projection effect or durable obligation, unit outcome, apply journal, idempotency success, immutable owner result, transaction participants, and a recoverable completion fact. | Proposed clarification of Core 01 REQ-01-620c with Core 03 unit atomicity |
| IRT-REQ-006 | Session/job finalization MUST be idempotent, MUST derive terminal state and ordered resource references from durable unit outcomes, and MUST NOT create owner resources. | Proposed Core 01 clarification |
| IRT-REQ-007 | Immediately before authoritative unit mutation, apply MUST rederive current incident visibility, membership, role, lifecycle, extension claim, target availability, and facade availability. | Core 04 and Core 01 |
| IRT-REQ-008 | Authorization and incident lifecycle checks MUST share the commit serialization boundary: whichever valid state change commits first determines the other operation's outcome. | Core 04 acceptance behavior |
| IRT-REQ-009 | Selection MUST reject a proposed overlapping persisted unit set atomically; apply MUST independently recheck overlap; concurrent conflicting selections MUST serialize so at most one succeeds. | Core 03 |
| IRT-REQ-010 | A skipped unit MUST retain its approved mapping. Reselection MUST recompute its state and the session state without creating a new mapping or fingerprint. | Core 03 |
| IRT-REQ-011 | `write_null` MUST be the only public and newly persisted null token. `omit_field` MUST omit the value and permit owner create defaults. `write_null` MUST create explicit null only for nullable create fields. | Core 01/Core 03 |
| IRT-REQ-012 | Core import top-level errors MUST remain closed. Target failures MUST be schema-validated nested owner details translated through the registered analytical binding. Unknown owner errors MUST fail closed. | Core 01 and target owner |
| IRT-REQ-013 | Discovery MUST cover CSV file, worksheet used range, table, static single-sheet single-rectangle named range, and operator-selected region. Presentation metadata MUST NOT change discovered values or rectangles. Formulas MUST NOT execute. | Core 03 and Core 01 |
| IRT-REQ-014 | One governed `cartulary.import_target_registry.v1` projection MUST generate backend, frontend, adapter, verification, and integrity outputs deterministically. Tests or source scans MUST NOT become target authority. | Proposed Core 01 projection |
| IRT-REQ-015 | The Import Assistant MUST consume only the generated public target projection, MUST hide reserved targets, MUST claim-gate Network Flow, and MUST NOT fall back to a manually maintained ID set. | Proposed generated web contract |
| IRT-REQ-016 | Every active test MUST have exactly one owner, semantic family, row ID, exact selector, evidence class, and verification set. Generated topology MUST NOT be hand-edited. | Testing Harness NLSpec |
| IRT-REQ-017 | Imports MUST construct no concrete peer store, perform no cross-owner reference SQL, and directly write no source-owner, analytical, projection, workbook, or grid-vendor state. | Core 01 dispatcher/owner-facade boundary |
| IRT-REQ-018 | Structural authorization and observable-correction authorization MUST remain separate. Passing characterization does not authorize a correction, and adopting a behavior owner does not authorize production mutation. | This handoff's execution discipline |

### 4.2 Public HTTP freeze

| Operation | Frozen request behavior | Frozen success behavior | Refactor rule |
| --- | --- | --- | --- |
| `POST /api/v1/import-sessions` | Existing upload envelope, metadata, media, idempotency, auth, and limits | Common job and import-session discovery result | Structural work MUST NOT change the route, operation ID, envelope, or source limits. |
| `GET /api/v1/import-sessions/{import_session_id}` | Existing visible session lookup | Existing import-session resource | State names and required members MUST remain stable. |
| `GET /api/v1/import-sessions/{import_session_id}/units` | Existing paging and visibility | Existing ordered import-unit list | Discovery order and paging MUST remain stable during structural work. |
| `GET /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}` | Existing session/unit lookup | Existing import-unit resource | Mapping, warning, and state serialization MUST remain stable. |
| `GET /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/preview` | Existing bounded preview request | Existing columns, rows, and truncation | Parser corrections require separate authorization. |
| `POST .../mapping-preview` | Existing analytical target/mapping candidate | Existing extension preview wrapper | Internal facade repair MUST NOT change this public request. |
| `PUT .../mapping` | Existing client transaction and exhaustive mapping | Existing import-unit resource | `write_null` correction MUST preserve the public token and fingerprint contract. |
| `POST .../select` | Existing client transaction | Existing session/selection/unit result | Backend overlap and reselection are authorized corrections, not structural side effects. |
| `POST .../skip` | Existing client transaction and optional reason | Existing session/selection/unit result | Skip/reselect correction MUST preserve the envelope. |
| `POST /api/v1/import-sessions/{import_session_id}/apply` | Existing client transaction and optional selected IDs | Common job; terminal import session and owner resource references | Unit/finalizer repair MUST preserve the public request and operation ID. |

No import-specific WebSocket route or event exists. The refactor MUST NOT invent one. Source-owner
collaboration or projection effects remain owner-specific postconditions behind owner facades.
Saved-view behavior remains outside imports. No public `view_schema_resource_v2.import_target` field
or new target-discovery route is required while frontend and backend ship together.

### 4.3 Analytical ownership split

| Contract concern | Required primary owner | Secondary-owner boundary |
| --- | --- | --- |
| Dispatch, target selection, opaque source/actor/mapping/idempotency capabilities, public wrapper, unit/session outcomes, cancellation, and job publication | Core 01 | A target MAY consume these facilities but MUST NOT redefine Core lifecycle or public import behavior. |
| Exact target preview/apply request and result members, target mapping, target validation, diagnostics, owner errors, and resource mutation | Adopted target NLSpec | Core MUST register schema IDs and common semantic obligations but MUST NOT duplicate target member lists. |
| Target recognition and contract-owner precedence | Core 00 | The owner matrix MUST identify the split before the repaired interface is implementable. |
| Network Flow import request/result schemas and table commit semantics | Network Flow Activity NLSpec | Network Flow MUST NOT redefine Core authorization, session lifecycle, idempotency, or job publication. |

### 4.4 `cartulary.imports.analytical_facade_binding.v1`

Every member is required. The binding is internal and machine-readable.

| Member | Type or closed value | Rule |
| --- | --- | --- |
| `schema_id` | exact string | MUST equal `cartulary.imports.analytical_facade_binding.v1`. |
| `target_kind` | registered token | Identifies the analytical target family. |
| `extension_profile_id` | registered profile ID | Identifies the owning claimed extension. |
| `owner_contract_ref` | owner ID plus major | MUST resolve to one adopted target owner. |
| `facade_id` | registered internal facade ID | MUST resolve to exactly one application binding. |
| `contract_major` | positive integer | MUST match the binding and referenced schemas. |
| `mapping_schema_id` | governed schema ID | Exact target-owned mapping schema. |
| `preview_request_schema_id` | governed schema ID | Exact target-owned preview request schema. |
| `preview_result_schema_id` | governed schema ID | Exact target-owned preview result schema. |
| `apply_request_schema_id` | governed schema ID | Exact target-owned apply request schema. |
| `apply_result_schema_id` | governed schema ID | Exact target-owned apply result schema. |
| `commit_protocol_id` | governed protocol ID | MUST identify the Core unit-commit protocol. |

The binding MUST convey these semantic slots without requiring common target member names:

| Semantic slot | Core obligation | Target obligation |
| --- | --- | --- |
| Session and unit identity | Supply authorized immutable identities | Validate request-schema placement and target use |
| Actor authorization | Supply transaction-current opaque actor context | Consume it only through the Core authorization boundary |
| Source bytes | Supply read-only capability bound to exact bytes, incident, unit, target, and digest | Read only through the capability and revalidate digest |
| Mapping preview | Supply proposed target mapping candidate | Validate and return target-owned materialized mapping |
| Apply mapping | Supply immutable approved mapping, fingerprint, and approval reference | Revalidate and consume the approved mapping |
| Idempotency | Supply opaque context bound to admitted request | Join the unit commit and return the immutable prior result on replay |
| Operation schema | Select the binding's exact operation schema | Validate exact target request/result shape |
| Owner result | Validate registered schema and translate errors | Return only target-owned result and safe diagnostics |

For Network Flow, the exact schema IDs remain:

- `cartulary.network_flow.import_preview_request.v1`
- `cartulary.network_flow.import_preview_result.v1`
- `cartulary.network_flow.import_apply_request.v1`
- `cartulary.network_flow.import_unit_result.v1`

The following compatibility map is migration guidance, not normative duplicate payload ownership:

| Former Core Table 17.2-F concept | Network Flow representation |
| --- | --- |
| `incident_id`, `actor_user_id` | `actor_context_ref`, Core-bound to actor, incident, operation, role, and target |
| `import_source_capability` | `source_stream_ref` |
| Source content digest | `source_content_sha256` |
| Parser, locator, source revision, descriptors | Bound into source capability and approved mapping; not separately trusted owner input |
| `proposed_owner_mapping` | `mapping_candidate` |
| `canonical_owner_mapping` | `approved_mapping` |
| `mapping_fingerprint` | `mapping_fingerprint` |
| Core `client_txn_id` | `idempotency_context_ref` bound to admitted Core transaction |
| Generic response fields | Exact target-owned preview or apply result |

An interim adapter MAY translate current internal objects. It MUST remain private to imports, MUST
NOT log capabilities or source locations, and MUST be removed when the binding is canonical.

### 4.5 Unit commit and session/job finalization

The two-level model is mandatory after coordinated owner adoption:

1. **Unit commit.** The owner resource or rows, owner audit, revisions/projections or durable
   publication obligation, import-unit result, apply journal, idempotency success, immutable owner
   result, required transaction participants, and recovery fact MUST commit atomically. A
   precommit failure or observed cancellation MUST leave none of those authoritative effects.
2. **Session/job finalization.** After the frozen selected units have durable outcomes, an
   idempotent finalizer MUST derive `applied`, `partially_applied`, `failed`, or `canceled`, publish
   the terminal common-job result, and return ordered committed resource references. It MUST NOT
   create an owner resource.

A crash after unit commit and before finalization MUST recover the committed unit result without
recreating the resource. A single-unit implementation MAY combine the unit and terminal job commit
only when it has identical recovery and caller-visible behavior.

### 4.6 Conformance-correction decisions

| Area | Required behavior | Default or omitted case | Compatibility class |
| --- | --- | --- | --- |
| Apply authorization | Revalidate current session, incident, membership, role, lifecycle, extension claim, target, and facade immediately before mutation. | Admission authorization alone is insufficient. | Behavior correction |
| Authorization race | Bind revalidation to the unit-commit serialization boundary. The first committed state change governs the other operation. | No best-effort or cached-role fallback. | Behavior correction |
| Unit atomicity | Commit all owner and import-unit effects or none. | A later unit failure does not roll back earlier committed units. | Behavior correction/owner repair |
| Partial session | At least one committed plus at least one failed/canceled unit derives `partially_applied`. No committed units derives `failed` unless cancellation owns the outcome. | Session-wide rollback is forbidden. | Behavior correction |
| Overlap | Selection atomically validates the proposed persisted set; apply rechecks it; concurrent conflicts serialize. | Use existing `import_apply_blocked` / `overlapping_units`. | Behavior correction |
| Reselection | Preserve mapping and fingerprint; recompute unit and session readiness. | Applying and terminal non-skipped units retain current conflict behavior. | Behavior correction |
| Empty value | `omit_field` omits the owner input; `write_null` sends explicit null only to nullable create fields. | Omission applies ordinary owner defaults. | Internal correction unless retained legacy data exists |
| Error ownership | Core top-level codes remain closed; owner detail is nested and schema-validated. | Unknown/unregistered owner tokens fail closed. | Contract-compatible behavior correction |
| XLSX locators | Admit used range, table, static single-sheet/single-rectangle name, and operator region in addition to CSV file. | Dynamic, external, and multi-area names fail as unsupported. | Additive behavior correction |
| Presentation neutrality | Hidden state, filters, sorting, and styles do not alter source rectangles or values; formulas never execute. | Only cached formula values may be read; required missing cache uses `formula_cached_value_missing`. | Behavior correction |

Repository search finds `use_null` only in `NormalizeImportScalar`; public validation and persisted
approved mappings use `write_null`. Therefore the default correction is an internal atomic
replacement with no migration. Before rollout, the authorized task MUST inspect retained
`approved_mapping_json` data. If any `use_null` value exists, it MUST use a bounded compatibility
decoder or explicit migration, MUST canonicalize fingerprints to `write_null`, and MUST never emit
new `use_null`.

### 4.7 Authorization, cancellation, replay, and error translation

| Condition | Unit outcome | Session/job outcome | Forbidden effect |
| --- | --- | --- | --- |
| Authorization/lifecycle failure before any unit commits | Current and remaining selected units fail deterministically | Session `failed`; common authorization or `incident_closed` failure | Owner mutation, journal success, resource ref, idempotency success |
| Authorization/lifecycle failure after earlier units commit | Earlier units remain applied; current and remaining units fail | Session `partially_applied` | Relabeling or rolling back earlier success |
| Cancellation before current unit commit | Current unit rolls back | Derived canceled/partial result from durable outcomes | Partial owner or import success |
| Cancellation after unit commit | Unit remains committed | Finalizer publishes/replays committed outcome | Deleting or relabeling committed success |
| Exact replay | Return immutable committed unit/session result | No new finalization side effect | Reauthorization that creates a second mutation |
| Replay detail read after role change | Recheck visibility before returning sensitive referenced resources | Return only currently visible detail | Leaking hidden owner resources |

Opaque actor, source, mapping-approval, and idempotency references MUST be bound to the intended
actor, incident, operation, session, unit, target, and invocation or authorization epoch. They MUST
NOT be reusable public bearer credentials and MUST NOT appear in logs, job results, diagnostics, or
audit payload values.

| Owner condition | Core top-level result | Nested owner detail |
| --- | --- | --- |
| Source digest, descriptor, or fingerprint changed | `import_apply_blocked` / `source_changed` | Optional registered source-change detail without bytes, paths, or capabilities |
| Network Flow has no data rows | `import_apply_blocked` / `owner_apply_validation_failed` | `network_flow_no_data_rows` plus registered reason |
| Network Flow rejects all rows | `import_apply_blocked` / `owner_apply_validation_failed` | `network_flow_all_rows_rejected` plus bounded diagnostics |
| Facade or schema missing/incompatible | `import_apply_blocked` / `owner_apply_contract_unavailable` | No internal stack, package, or registration path |
| Owner result is structurally invalid | `import_apply_blocked` / `owner_apply_validation_failed` | Safe schema ID and validation class only |
| Owner code is unknown | `import_apply_blocked` / `owner_apply_validation_failed` | Unregistered token MUST NOT be echoed |
| Mapping preview owner validation fails | `invalid_import_request` / `owner_preview_validation_failed` | Registered bounded preview detail |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action and requirement |
| --- | --- | --- | --- | --- | --- |
| Core and Network Flow duplicate exact analytical facade authority | Core 01 Table 17.2-F; Network Flow Tables 10-A0/A1/A1a | Interface incompatibility | `must_fix` | Coordinated Core/Network Flow owners | Adopt IRT-REQ-003 through IRT-REQ-006 through GATE-02. |
| Imports constructs seven concrete peer stores | `owner_apply.go` | Cross-owner composition and change blast radius | `must_fix` | application composition root | Inject one owner registry under IRT-REQ-017 after characterization. |
| Source-owner validation/default/reference logic is duplicated in imports | `owner_apply.go` and peer import-create files | Drift and inconsistent errors | `must_fix` | each source owner | Move semantics behind owner facades while preserving characterized outcomes. |
| Store writes import-owned tables directly | `store.go` and imports migrations | Low when encapsulated | `intentional/no_action` | imports | Keep behind Store; do not generalize it. |
| Apply coordination performs cross-owner reference SQL | `owner_apply.go` | Bypasses source-owner facade | `must_fix` | relevant source owner | Move lookup and error semantics behind selected owner facade. |
| HTTP, jobs, discovery, mapping, and apply share `routes.go` | Live file responsibilities | High structural blast radius | `should_fix` | imports internal service/adapters | Split by file responsibility only after RS-00; retain public package facade. |
| Unit publication and session/job finalization are conflated | Core/live transaction and finalization sequence | Partial result or duplicate recovery | `must_fix` | imports, target owner, jobs | Adopt and implement IRT-REQ-005/006. |
| Current role is not visibly rechecked at mutation commit | Route and transaction paths | Unauthorized delayed mutation | `must_fix` | imports using platform auth | Implement IRT-REQ-007/008 as a separately authorized correction. |
| Overlap is blocked only in the frontend | Import Assistant and backend paths | Alternate-client conflict | `must_fix` | imports | Implement IRT-REQ-009. |
| Skipped units cannot visibly be reselected | Store state guards | Incorrect workflow state | `must_fix` | imports | Implement IRT-REQ-010. |
| XLSX discovery lacks owned locator kinds and neutrality evidence | `xlsx.go` and Core 03 | Missing or presentation-dependent units | `must_fix` | imports source adapter | Implement IRT-REQ-013 with fixtures. |
| Public `write_null` reaches an internal `use_null` switch | `api.go`, persisted approved mapping, ownerfacade | Null silently rejected | `must_fix` | imports contract/facade | Apply the §4.6 preflight and IRT-REQ-011. |
| Core and target error vocabularies need explicit translation | Live mapping plus owners | Client drift and unsafe leakage | `must_fix` | Core plus target owner | Register §4.7 translation under IRT-REQ-012. |
| Frontend hardcodes 14 importable IDs | Import Assistant and backend registry | Frontend/backend drift | `must_fix` | Core import-target projection | Generate both consumers under IRT-REQ-014/015. |
| Verification rows omit targets and boundary/security cases | Imports owner/family inputs | Incomplete evidence accounting | `must_fix` | imports/source owners/harness | Satisfy IRT-REQ-016 and GATE-01/GATE-07. |
| Platform archive/auth/job dependencies occur at adapter boundaries | `limits.go`, `routes.go`, job finalizer port | Acceptable when isolated | `intentional/no_action` | platform mechanics plus imports adapter | Preserve ports; keep platform mechanics out of owner rules. |
| No direct grid-vendor coupling was found | Targeted searches | None on current evidence | `intentional/no_action` | grid adapter if later applicable | Retain boundary checks only. |
| No target-specific test-util mutation path was found | Integration test composition | None on current evidence | `intentional/no_action` | test-util owner | Reassess only if a later authorized slice touches test utilities. |
| Generated surfaces are downstream projections | Generated artifact policy | Hand-edit/drift risk | `intentional/no_action` | contract owners and generators | Change authored inputs and regenerate; never hand-edit generated roots. |

## 6. Refactor Workstreams and Execution Gates

### 6.1 Binary execution gates

These gates replace unresolved planning blockers. A gate is a prerequisite with a binary exit
condition, not a design choice left to an implementer.

| Gate | Required action | Status | Exit condition | Blocks |
| --- | --- | --- | --- | --- |
| GATE-01 Characterization | Add route, target, state, owner-effect, security, boundary, and frontend characterization from registry-driven inputs. | TODO | Current behavior is frozen; known non-conformance is labeled; required narrow baseline is green. | All structural slices |
| GATE-02 Coordinated owner repair | Adopt the IRT-REQ-003 through IRT-REQ-006 ownership split in Core 00, Core 01, Network Flow, and governed schemas as one logical change. | TODO | No exact payload has two owners; unit commit and session finalization are coherent. | Analytical facade and transaction changes |
| GATE-03 Correction authorization | Adopt one Imports Conformance Corrections Authorization row per observable correction. | TODO | Each correction has owner requirements, compatibility class, tests, migration posture, rollback boundary, approver, and adopted status. | RS-06 through RS-10 |
| GATE-04 Registry projection | Adopt and generate `cartulary.import_target_registry.v1`. | TODO | Backend, frontend, adapter, verification, and integrity outputs validate and share one source digest. | Registry injection and frontend cleanup |
| GATE-05 Structural authorization | Explicitly authorize behavior-preserving registry injection and file-level decomposition. | TODO | GATE-01, GATE-02, and GATE-04 pass; authorized files and rollback boundaries are recorded. | RS-03 through RS-05 |
| GATE-06 Correction implementation | Authorize and implement each observable correction independently. | TODO | Each selected correction's current and normative tests pass with required migration/compatibility evidence. | Final validation |
| GATE-07 Harness closure | Add exact authored owner/family rows, regenerate topology, and run narrow-to-broad checks. | TODO | No uncovered target, duplicate selector, unowned test, or generated drift remains. | Final handoff |

The conformance authorization artifact required by GATE-03 MUST contain:

| Field | Required meaning |
| --- | --- |
| `correction_id` | Stable identity for authorization, atomicity, overlap, reselection, null, errors, or discovery |
| `normative_owner_requirements[]` | Exact Core/subsystem requirement and acceptance IDs |
| `observable_change` | HTTP, state, stored-data, job, or frontend behavior affected |
| `compatibility_class` | `behavior_correction`, `additive_contract`, `data_migration`, or `internal_only` |
| `characterization_prerequisites[]` | Current-behavior tests required before correction |
| `normative_acceptance_tests[]` | Tests that prove the adopted behavior |
| `migration_requirement` | `none`, `decoder`, `data_migration`, or `regenerate_contracts` |
| `rollback_boundary` | Exact code, contract, data, and generated artifacts that revert together |
| `authorization_owner` | Person or process permitted to approve the behavior class |
| `status` | `proposed`, `adopted`, `implemented`, or `verified` |

### 6.2 Workstreams

| Workflow ID | Name | Class | Required previous workflows/gates | Required subsequent workflows | Goal | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Scope, source, and tracker posture | root | None | WF-01 through WF-04 | Fix authority, baseline, language, and write boundary | `make lint-markdown` | Tracker is internally coherent and only it changed. |
| WF-01 | Live target inventory | chain | WF-00 | WF-02 through WF-04 | Account for every file, caller, dependency, and test | Repository inspection | All 17 tracked files have target-specific rows. |
| WF-02 | Contract and owner resolution | chain | WF-01 | WF-05 | Make ownership and interfaces decision-complete without claiming adoption | Human owner review; later contract checks | RB-001 and RB-002 are resolved in tracker; GATE-02/03 remain explicit. |
| WF-03 | Characterization and evidence design | parallel | WF-01 | WF-05, WF-07 | Define exact coverage and accounting for all 18 targets | Owner slices and frontend checks | GATE-01 plan is complete. |
| WF-04 | Boundary and projection design | parallel | WF-01 | WF-05 | Specify owner registry, composition, and generated target projection | Boundary and generation checks | RB-003/RB-004 are resolved in tracker. |
| WF-05 | Structural sequence | chain | WF-02 through WF-04; GATE-01/02/04/05 | WF-06 | Inject facades and split responsibilities without behavior changes | Narrow owner and boundary targets | Imports is a thin coordinator with characterized behavior intact. |
| WF-06 | Conformance corrections | chain/parallel | WF-05; GATE-03 | WF-07 | Implement independently authorized observable corrections | Per-correction tests | Each correction is implemented and verified independently. |
| WF-07 | Harness and final handoff | chain | Applicable WF-05/06; GATE-07 | None | Complete accounting and narrow-to-broad evidence | `make agent-finalize`, owner slices, `make check` | Results, failures, run roots, and skips are recorded. |

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation | Rollback and completion |
| --- | --- | --- | --- | --- | --- | --- | --- |
| RS-00 | None | Add behavior-freeze characterization for all routes, targets, states, security transitions, owner effects, boundaries, and frontend selectors. | Imports/source-owner tests, web tests, authored harness inputs later | Characterization MUST label known non-conformance and MUST NOT become normative authority. | §8 matrices | Owner slices and frontend checks | Revert only new tests; complete when GATE-01 passes. |
| RS-01 | RS-00 | Adopt coordinated owner repair for analytical ownership and two-level commit; no production code. | Core 00, Core 01, Network Flow owner and governed schemas in a separate authorized owner task | An intermediate adoption could remain self-contradictory. | Owner contract/fixture validation | Owner-selected documentation and contract targets | Revert the logical owner repair together; complete when GATE-02 passes. |
| RS-02 | RS-01 | Author and generate `cartulary.import_target_registry.v1` and integrity projections; no hand edits to outputs. | Core-owned contract input, generators, generated Go/TS/verification projections | Target availability or schema drift | Generator validation and 18-row parity | `make generate-drift`; generated policy checks | Revert authored input and regenerate; complete when GATE-04 passes. |
| RS-03 | RS-00 through RS-02 | Inject one application-composed owner registry; remove concrete peer-store construction and cross-owner SQL from imports. | `owner_apply.go`, ownerfacade, application assembly, peer facades | Owner order, validation, defaults, revisions, projection effects | Owner matrix, missing/duplicate bindings, rollback | Imports owner slice; backend boundary check | Revert registry injection as one slice; complete when exactly one owner is selected and imports has no concrete stores. |
| RS-04 | RS-03 | Route Timeline through the same owner-create registry. | Imports apply path, Timeline facade, application composition | Timeline defaults, event/revision order | Existing Timeline case plus rollback/replay/projection/effect cases | Service-backed imports and Timeline owner slices | Restore Timeline registration and adapter together; complete on effect parity. |
| RS-05 | RS-03 | Split HTTP handlers, service, job coordination, discovery/mapping, apply coordination, and Store by file responsibility inside the public imports package. | Primarily routes, store, XLSX, API, job/facade files | Accidental route, error, transaction, or parser change | Entire RS-00 suite | Imports slice and boundary checks | Revert each extraction independently; complete when public symbols and behavior are unchanged. |
| RS-06 | RS-01, RS-03, RS-05; GATE-03 | Add transaction-current authorization and two-level unit/finalizer commit — **requires later authorization**. | Imports service/Store/jobs, auth port, revisions, target participants | Observable auth, transaction, and terminal-state changes | Role/membership/close/claim races, failure injection, replay | Service-backed owner slices and `make test-fast` | Revert complete auth/transaction slice; complete when IRT-REQ-005 through 008 pass. |
| RS-07 | RS-00, RS-05; GATE-03 | Enforce backend overlap and skipped-unit reselection — **requires later authorization**. | Store/service state machine and frontend tests | Observable state/error changes | Raw HTTP overlap, concurrent select, skip/reselect, mapping retention | Service-backed imports and frontend unit | Revert state-machine slice; complete when IRT-REQ-009/010 pass. |
| RS-08 | RS-00, RS-05; GATE-03 | Replace internal `use_null`, implement registered error translation, and preserve canonical fingerprints — **requires later authorization**. | API, ownerfacade, route error translation, owner binding | Stored mapping and client error compatibility | Empty policies, retained-data preflight, every translation row, unknown owner error | Imports slice; compatibility/generation checks if contract changes | Revert implementation and any migration together; complete when IRT-REQ-011/012 pass. |
| RS-09 | RS-00, RS-05; GATE-03 | Add required XLSX locator kinds and presentation neutrality — **requires later authorization**. | XLSX adapter, owner-neutral tabular-ingest primitives if needed, fixtures | Discovery order, locator, warnings, source bounds | Used range, table, named range variants, region, hidden state, formula cache, limits | Service-backed imports slice | Revert by locator kind while retaining bounded reader; complete when IRT-REQ-013 passes. |
| RS-10 | RS-02, RS-08; GATE-03/04 | Replace frontend hardcoded target IDs with generated semantic catalog — **requires later authorization**. | Generated UI contracts, Import Assistant, coordinator/tests | Target visibility and claimed/unclaimed workflow | 18-row disposition, selector stability, no fallback, server revalidation | Frontend unit/typecheck/boundary/browser checks | Revert authored projection and consumer together; complete when IRT-REQ-014/015 pass. |
| RS-11 | All applicable slices | Add exact harness rows, regenerate topology, finalize, and run risk-appropriate broad checks. | Authored verification/test-family/catalog inputs and generated topology | Missing/duplicate accounting and hand-edit risk | Every active test mapped exactly once | Owner slices, generate drift, `make agent-finalize`, `make check` | Revert authored inputs and regenerate; complete when GATE-07 and IRT-REQ-016 pass. |

## 8. Validation and Evidence Plan

### 8.1 Current 18-target evidence matrix

| Target selector | Availability | Facade | Required minimum evidence |
| --- | --- | --- | --- |
| `cartulary.view.timeline.v2` | enabled/selectable | owner create | Valid apply, Timeline defaults, revision, projection, collaboration, rollback |
| `cartulary.view.hosts.v1` | enabled/selectable | owner create | Valid apply, authoritative row/version, host validation, projection, rollback |
| `cartulary.view.identities.v1` | enabled/selectable | owner create | Valid apply, identity validation/defaults, projection, rollback |
| `cartulary.view.evidence.v1` | enabled/selectable | owner create | Valid apply, evidence owner effects, revision/journal, rollback |
| `cartulary.view.notes.v1` | enabled/selectable | owner create | Valid note variant, owner defaults/references, projection, rollback |
| `cartulary.view.indicators.v1` | enabled/selectable | owner create | Valid indicator normalization/deduplication, projection, rollback |
| `cartulary.view.assessments.v1` | enabled/selectable | owner create | Valid assessment, direct references, projection, rollback |
| `cartulary.view.task_requests.v1` | enabled/selectable | owner create | Valid task request, references/defaults, projection, rollback |
| `cartulary.view.decisions.v1` | enabled/selectable | owner create | Valid decision, references/defaults, projection, rollback |
| `cartulary.view.parties.v1` | enabled/selectable | owner create | Valid party, incident/reference validation, projection, rollback |
| `cartulary.view.comm_log.v1` | enabled/selectable | owner create | Valid artifact variant, party references, projection, rollback |
| `cartulary.view.handoff.v1` | enabled/selectable | owner create | Valid artifact variant, task/decision/risk references, rollback |
| `cartulary.view.status_review.v1` | enabled/selectable | owner create | Valid artifact variant, task/evidence/decision references, rollback |
| `cartulary.view.lesson.v1` | enabled/selectable | owner create | Valid artifact variant, task/evidence references, rollback |
| `cartulary.view.findings.v1` | reserved/hidden | no enabled binding | Mapping/apply unavailable; absent from frontend selection |
| `cartulary.view.investigative_queries.v1` | reserved/hidden | no enabled binding | Mapping/apply unavailable; absent from frontend selection |
| `cartulary.view.forensic_keywords.v1` | reserved/hidden | no enabled binding | Mapping/apply unavailable; absent from frontend selection |
| `target_kind=network_flow_table; extension_profile_id=network_flow_activity` | claim-gated | owner preview/apply | Unclaimed unavailability; claimed preview/apply; no-data/all-rejected/partial/source-change/cancel/recovery/replay |

Every enabled view-schema target MUST prove correct owner selection, authoritative record and
`row_version`, one unit change set, committed journal, expected projection, and rollback. Failure
injection MAY be parameterized by distinct facade/transaction implementation, but a target with
different defaults or projections requires target-specific assertions.

### 8.2 Route, state, security, and owner-effect coverage

The ten operations in §4.2 require exact success, malformed request, authorization, visibility,
state-conflict, and owner-failure coverage. The state suite MUST cover every legal and illegal
session/unit transition, idempotent select/skip, skip/reselection, mapping retention, overlap at
selection/apply, concurrent selection, frozen selected set, deterministic order, duplicate apply,
explicit re-import, and all four terminal session outcomes.

Security coverage MUST include role revocation, membership removal, incident close race, claim
removal, facade loss, cross-incident actor context, cross-session/unit source capability, target
substitution, source and fingerprint change, unauthorized job polling/cancellation, and replay
after a role change with safe visibility.

For each applicable owner family, evidence MUST assert exactly one change set, correct actor/reason/
session/unit/provenance, no effect on rollback, projection visibility after commit, authoritative
row refresh, and owner-required collaboration or outbox effects. Tests MUST NOT invent an
imports-specific WebSocket event.

### 8.3 Registry projection and generator contract

`cartulary.import_target_registry.v1` MUST use these closed fields:

| Field | Rule |
| --- | --- |
| `target_kind` | `view_schema` or `network_flow_table` |
| Selector | Exactly one of `target_view_schema_id` or `extension_profile_id` |
| `owner_contract_ref` | Exact adopted owner and major |
| `facade_kind` | `owner_create` or `owner_preview_apply` |
| `availability_kind` | `enabled`, `reserved`, or `claim_gated` |
| `mapping_contract_schema_id` | Exact governed mapping schema |
| `default_unknown_column_policy` | Exact policy or target-owned marker |
| `entity_bearing_default` | Exact Core value, target-owned marker, or explicit none |
| `facade_id` | Exact internal binding identity |
| Analytical schema IDs | Required only for `owner_preview_apply` |
| `public_projection_disposition` | `selectable`, `hidden_reserved`, or `extension_claim_gated` |

The authored source MUST be explicitly cataloged, schema-validated, canonically ordered, and free of
timestamps, host paths, branches, random values, process IDs, modification times, or unordered-map
output. It MUST NOT be generated by parsing Markdown, scanning directories, inspecting filenames,
reflecting Go packages, probing routes, reading frontend components, or treating tests as authority.

Generation MUST produce:

1. Backend target bindings used by mapping and dispatch.
2. A frontend semantic catalog containing no package, SQL, store, route-internal, or grid-vendor
   details.
3. Typed contract adapters.
4. A verification parameter projection that does not restate product behavior.
5. An integrity manifest proving one reviewed source digest.

Generation MUST fail for duplicate selectors; both/neither selector variants; unknown owner,
facade, or disposition; enabled target without binding; reserved target projected selectable;
analytical target missing a schema ID; unknown view schema; projection digest mismatch; or
nondeterministic output.

The Import Assistant MUST intersect the generated catalog with discovered server target IDs,
present only `selectable`, omit `hidden_reserved`, and gate `extension_claim_gated` against current
extension claims. The server MUST independently revalidate all target and authorization conditions.

### 8.4 Harness families

| Family | Scope |
| --- | --- |
| `imports.http_contract` | Ten public route contracts and errors |
| `imports.state_machine` | Session and unit transitions |
| `imports.discovery` | CSV/XLSX discovery, ordering, limits, and neutrality |
| `imports.owner_dispatch` | Registry selection and owner facade routing |
| `imports.unit_of_work` | Commit, rollback, replay, and finalization |
| `imports.security_transitions` | Current role, lifecycle, claim, and capability changes |
| `imports.owner_effects` | Revisions, projections, row refresh, and collaboration |
| `imports.boundaries` | Static dependency and ownership guards |
| `imports.frontend_workflow` | Coordinator, assistant, generated target catalog, and selectors |

Behavior-freeze characterization and normative conformance MUST remain distinct. When a correction
lands, a legacy test encoding non-conforming behavior MUST be retired or replaced. No test or tool
may read documentation text, line numbers, headings, hashes, or formatting to establish product
conformance.

### 8.5 Make-owned validation

Only Markdown validation is appropriate for this tracker-only revision. Implementation commands are
recorded for later authorized work and MUST NOT be claimed as passing here.

| Layer | Command | Required phase |
| --- | --- | --- |
| Tracker documentation | `make lint-markdown` | This revision |
| Fast baseline | `make test-fast`; `make frontend-unit` | GATE-01 |
| Imports owner slice | `make test-slice OWNER=module.imports` | GATE-01 and every backend slice |
| Service-backed owner slice | `make service-backed-test-slice OWNER=module.imports` | GATE-01 and persistence/transaction slices |
| Browser | `make browser-e2e-webserver-backed` | Frontend/public workflow changes |
| Generated drift | `make generate-drift` | GATE-04 and generated changes |
| Generated policy | `make generated-artifact-policy-check` | GATE-04 and generated changes |
| OpenAPI compatibility | `make openapi-compatibility-check` | Authorized public contract change only |
| Backend boundary | `make backend-module-boundary-check` | RS-03 through RS-05 |
| Frontend boundary/type | `make frontend-import-boundary-check`; `make frontend-typecheck` | RS-10 |
| Finalization | `make agent-finalize` | Before broad implementation verification |
| Full check | `make check` | Final risk-appropriate implementation handoff |

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| IMP-001 | Fix target scope, authority, baseline, and write boundary | WF-00 | DONE | None | §1 | Exact scope and non-goals are normative. |
| IMP-002 | Inventory every tracked target file and direct boundary | WF-01 | DONE | IMP-001 | §2 | All 17 tracked files have live target-specific rows. |
| IMP-003 | Resolve analytical ownership and commit semantics in the tracker | WF-02 | DONE | IMP-002 | IRT-REQ-003 through 006; RB-001 | Planning decision is complete without claiming owner adoption. |
| IMP-004 | Resolve conformance behavior decisions in the tracker | WF-02 | DONE | IMP-002 | §4.6/4.7; RB-002 | Every correction has exact behavior, default, and compatibility class. |
| IMP-005 | Define complete characterization and harness posture | WF-03 | DONE | IMP-002 | §8; RB-003 | All 18 targets and cross-cutting invariants have binary evidence obligations. |
| IMP-006 | Define one generated target registry | WF-04 | DONE | IMP-002 | §8.3; RB-004 | Fields, outputs, failures, and frontend rules are decision-complete. |
| IMP-007 | Complete characterization baseline | WF-03 | TODO | IMP-005; GATE-01 | RS-00 | Required narrow suites pass with known non-conformance labeled. |
| IMP-008 | Adopt coordinated owner repair | WF-02 | TODO | IMP-003; GATE-02 | RS-01 | Core/Network Flow/machine owners are coherent and adopted together. |
| IMP-009 | Adopt and generate target registry | WF-04 | TODO | IMP-006, IMP-008; GATE-04 | RS-02 | All projections share one source digest. |
| IMP-010 | Inject owner registry and remove ownership leakage | WF-05 | TODO | IMP-007 through IMP-009; GATE-05 | RS-03/04 | Exactly one owner selected; no concrete peer stores or cross-owner SQL in imports. |
| IMP-011 | Split imports internal responsibilities | WF-05 | TODO | IMP-010; GATE-05 | RS-05 | Characterized public behavior is unchanged. |
| IMP-012 | Adopt conformance correction authorization | WF-06 | TODO | IMP-004, IMP-007; GATE-03 | Authorization artifact | Every correction row is adopted before implementation. |
| IMP-013 | Implement transaction/auth correction | WF-06 | TODO | IMP-008, IMP-010 through IMP-012; GATE-06 | RS-06 | IRT-REQ-005 through 008 pass. |
| IMP-014 | Implement state/null/error/discovery corrections | WF-06 | TODO | IMP-011/012; GATE-06 | RS-07 through RS-09 | IRT-REQ-009 through 013 pass independently. |
| IMP-015 | Replace frontend hardcoded registry | WF-06 | TODO | IMP-009, IMP-012; GATE-06 | RS-10 | IRT-REQ-014/015 pass with selector parity. |
| IMP-016 | Complete harness accounting and final evidence | WF-07 | TODO | Applicable implementation rows; GATE-07 | RS-11 | Every active test is accounted for once; required checks are recorded. |
| IMP-017 | Validate this NLSpec-style tracker revision | WF-00 | DONE | IMP-001 through IMP-006 | `make lint-markdown` | Markdown and structural checks passed with only this file changed. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-28 17:48 EDT | Codex tracker-planning session | Scope, hierarchy, baseline, and non-goals fixed | Inspected planning framework, domain/Core/subsystem owners; touched this tracker only | Repository reads/searches; `git status --short --branch`; `git rev-parse HEAD`; `make lint-markdown` | Planning baseline complete; Markdown passed at `.cartulary/test-results/20260728T215218Z-p1253262` | None for tracker publication | Preserve the one-file diff; implementation needs a new task. |
| 2026-07-28 18:29 EDT | Codex NLSpec-style tracker revision | Normative planning posture and decision/gate distinction added | Read analysis notes, NLSpec guide, local owners, current registry/code; touched this tracker only | `rg`, `sed`, `find`, `git status`, exact owner/code reads; `make lint-markdown`; structural checks | RB decisions resolved; validation passed at `.cartulary/test-results/20260728T223545Z-p1268589` | No open design question; GATE-01 through GATE-07 remain | Preserve the one-file write boundary; implementation needs a later task. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-28 17:48 EDT | Codex tracker-planning session | Valid imports boundary diagnosed as internally mixed; all target files inventoried | Inspected all `internal/modules/imports` files, app composition, Network Flow, seven peer owner facades, and migrations; touched this tracker only | `rg`/`git ls-files`/file reads; `make task-guide ROLE=module-author OWNER=module.imports`; target explanations | Inventory and coupling scan complete; tracker Markdown passed | Historical RB-001/RB-003 | Characterize behavior, resolve owner contract, then inject owner registry. |
| 2026-07-28 18:29 EDT | Codex NLSpec-style tracker revision | Owner registry, composition, and two-level commit are specified as gated future requirements | Same live backend evidence plus owner tables; touched this tracker only | Exact Core 01/Network Flow reads and target-registry inspection | Structural path is decision-complete; tracker validation passed | GATE-01, GATE-02, GATE-04, GATE-05 | Complete characterization and coordinated adoption before code movement. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-28 17:48 EDT | Codex tracker-planning session | Controller/assistant surfaces mapped; registry duplication found; no direct grid vendor coupling found | Inspected import coordinator/tests, Import Assistant, Network Flow controller, browser spec, generated contract adapter locations; touched this tracker only | Targeted searches and target explanations | Frontend risk and test posture recorded; tracker Markdown passed | Historical RB-004 | Preserve selectors; wait for an authoritative projection. |
| 2026-07-28 18:29 EDT | Codex NLSpec-style tracker revision | Generated catalog fields, availability rules, and no-fallback consumption are specified | Import Assistant hardcoded set, view-schema index, generated roots; touched this tracker only | Exact file reads and projection-pattern searches | RB-004 design is resolved; tracker validation passed | GATE-04 and GATE-06 | Generate catalog before removing the manual set. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-28 17:48 EDT | Codex tracker-planning session | Ten HTTP operations and generated consumers mapped; analytical facade contradiction identified | Inspected authored imports OpenAPI, Core/Network Flow owners, generated consumer locations, target registry; touched this tracker only | `jq`, searches, file reads, generation/compatibility target explanations | Contract freeze map complete except contradicted facade; tracker Markdown passed | Historical RB-001 | Obtain adopted owner resolution before interface changes. |
| 2026-07-28 18:29 EDT | Codex NLSpec-style tracker revision | Binding schema, ownership split, projection schema, generated outputs, and failure rules are specified | Core/Network Flow exact facade tables, extension projection pattern, generated policy; touched this tracker only | Exact owner and contract-pattern reads | Contract design is complete without claiming adoption; tracker validation passed | GATE-02 and GATE-04 | Adopt owners and authored machine projection as one coordinated change. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-28 17:48 EDT | Codex tracker-planning session | Existing tests and five routed owner rows mapped; missing owner/boundary/accounting coverage identified | Inspected imports, frontend, Network Flow tests, and verification inputs; touched this tracker only | Owner/test/target explanation commands | Gap analysis recorded; no test suite run; tracker Markdown passed | Historical RB-003 | Add characterization before structural movement. |
| 2026-07-28 18:29 EDT | Codex NLSpec-style tracker revision | Eighteen-target, route/state/security/effect, boundary, family, and exact-accounting obligations specified | Same test/harness evidence plus notes recommendations; touched this tracker only | Repository searches and exact harness reads | RB-003 design is resolved; no implementation suite run; tracker validation passed | GATE-01 and GATE-07 | Implement RS-00 in a later authorized task. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-28 17:48 EDT | Codex tracker-planning session | Admission checks found; immediate pre-mutation current-role recheck not visible | Inspected route authorization, Store guards, Network Flow participants, Core 04; touched this tracker only | Targeted searches and exact file reads | Security gap recorded; no behavior changed; tracker Markdown passed | Historical RB-002 | Add role-revocation characterization and authorize correction. |
| 2026-07-28 18:29 EDT | Codex NLSpec-style tracker revision | Transaction-current authorization, race, cancellation, replay, and capability rules specified | Core 04 acceptance behavior, routes, transaction participants; touched this tracker only | Exact owner and code reads | RB-002 security decisions are complete; tracker validation passed | GATE-03 and GATE-06 | Adopt correction authorization before RS-06. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-28 17:48 EDT | Codex tracker-planning session | Workstreams and reversible slices sequenced; implementation unauthorized | This tracker only was touched; evidence read-only | `make help-all` plus targeted explanations; `make lint-markdown` | Handoff decision-complete subject to historical blockers; tracker passed | Historical RB-001 through RB-004 | Resolve RB-001 and begin characterization in a new task. |
| 2026-07-28 18:29 EDT | Codex NLSpec-style tracker revision | All four RB decisions resolved; seven binary gates replace ambiguous blockers | This tracker only touched; analysis notes and authorities read-only | Repository inspection; `make lint-markdown`; diff/structural/one-file checks | No open planning choice remains; tracker validation passed | GATE-01 through GATE-07 are required prerequisites, not open questions | Start with GATE-01 only after a later task authorizes test changes. |

## 11. Resolved Questions and Remaining Gates

There are no remaining design questions for the refactor sequence described by this tracker. The
former RB items are closed at the planning layer as follows:

| ID | Resolution | Decision status | What is not claimed | Required next gate |
| --- | --- | --- | --- | --- |
| RB-001 | Core owns generic orchestration and semantic obligations; the target NLSpec owns exact target payloads; the binding and two-level commit model join them. | RESOLVED_IN_TRACKER | Current Core 00/Core 01/Network Flow text is not yet repaired or adopted. | GATE-02 |
| RB-002 | Authorization, races, unit/partial outcomes, overlap, reselection, nulls, errors, XLSX, cancellation, replay, and capability behavior are exact in §4. | RESOLVED_IN_TRACKER | Observable corrections are not authorized or implemented. | GATE-03 and GATE-06 |
| RB-003 | The 18-target matrix, cross-cutting suites, boundary rules, harness families, and exact accounting define a complete evidence baseline. | RESOLVED_IN_TRACKER | Tests and harness rows do not yet exist or pass. | GATE-01 and GATE-07 |
| RB-004 | One Core-backed deterministic registry supplies backend, frontend, adapter, verification, and integrity projections; no public field is required for same-release deployment. | RESOLVED_IN_TRACKER | The authored registry and generated outputs do not yet exist. | GATE-04 |

An implementation agent MUST treat any unmet gate in §6.1 as a hard dependency. It MUST NOT reinterpret
`RESOLVED_IN_TRACKER` as permission to edit an owner, production code, contracts, migrations,
generated files, tests, or harness inputs.

## 12. Binary Completion Criteria

### 12.1 Requirement-to-acceptance mapping

| Acceptance ID | Requirements | Pass condition |
| --- | --- | --- |
| IRT-AC-001 | IRT-REQ-001, IRT-REQ-018 | No production slice begins before its owner, characterization, and authorization gates; structural and correction changes remain separate. |
| IRT-AC-002 | IRT-REQ-002 | All ten HTTP operation IDs, paths, requests, responses, errors, and generated public types are unchanged by structural slices; no imports WebSocket event is added. |
| IRT-AC-003 | IRT-REQ-003, IRT-REQ-004 | Core and target ownership are adopted without duplicated exact payload authority; every analytical binding and referenced schema resolves exactly once. |
| IRT-AC-004 | IRT-REQ-005, IRT-REQ-006 | Precommit failure/cancellation leaves no authoritative unit effects; postcommit crash recovers one result; finalization creates no owner resource. |
| IRT-AC-005 | IRT-REQ-007, IRT-REQ-008 | Role/membership/claim/lifecycle removal before commit prevents mutation; a mutation committed first is observed by the administrative change. |
| IRT-AC-006 | IRT-REQ-009 | Overlap fails atomically at selection and apply; concurrent conflicting selection yields at most one success. |
| IRT-AC-007 | IRT-REQ-010 | Skip then select retains the mapping/fingerprint and recomputes unit/session readiness without remapping. |
| IRT-AC-008 | IRT-REQ-011 | `write_null` succeeds only for nullable create fields, `omit_field` applies defaults, new state never emits `use_null`, and retained-data handling matches the preflight. |
| IRT-AC-009 | IRT-REQ-012 | Every registered error translation produces the exact Core result and safe nested owner detail; unknown tokens fail closed without leakage. |
| IRT-AC-010 | IRT-REQ-013 | CSV and all four XLSX locator kinds are deterministic and bounded; hidden/style/filter/sort metadata has no semantic effect; formulas do not execute. |
| IRT-AC-011 | IRT-REQ-014 | All five registry projections are byte-deterministic, share one source digest, cover all 18 rows, and fail every invalid-generation case in §8.3. |
| IRT-AC-012 | IRT-REQ-015 | Frontend has no manual target-ID set; selectable/reserved/claim-gated behavior and selectors pass unit and browser evidence; server revalidation remains mandatory. |
| IRT-AC-013 | IRT-REQ-016 | Every active test resolves to exactly one owner/family/row/selector/evidence/verification set, with zero uncovered targets and duplicate selectors. |
| IRT-AC-014 | IRT-REQ-017 | Imports constructs no concrete peer stores, performs no cross-owner SQL, and writes no owner, analytical, projection, workbook, or grid state directly. |
| IRT-AC-015 | All | Narrow owner and frontend checks pass before `make agent-finalize`; risk-appropriate broader checks and any skips are recorded with run roots. |

### 12.2 Tracker readiness

- [x] Every tracked file in `internal/modules/imports` is inventoried; the empty untracked directory
  is explicitly explained.
- [x] All ten public HTTP operations and every discovered cross-boundary contract are frozen.
- [x] Core-versus-target analytical ownership and the binding interface are exact.
- [x] Unit commit and session/job finalization are specified separately.
- [x] Authorization, race, overlap, reselection, null, errors, XLSX, cancellation, replay, and
  capability defaults are explicit.
- [x] All 18 target rows have availability and evidence obligations.
- [x] Registry fields, outputs, frontend behavior, and generator failures are explicit.
- [x] Every workflow and implementation slice has dependencies, validation, rollback, and a binary
  completion condition.
- [x] Former RB items are `RESOLVED_IN_TRACKER`; unmet adoption/execution work is represented by
  binary gates without false completion claims.
- [x] Prior session history is preserved and the current revision session is appended.
- [x] Implementation remains a later authorized task.
- [x] `make lint-markdown`, `git diff --check`, structural checks, and the one-file status check pass
  for this revision.
