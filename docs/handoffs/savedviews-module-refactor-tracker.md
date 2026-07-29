# savedviews Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

- **Target path:** `internal/modules/savedviews`
- **Normalized target label:** `savedviews`
- **Output path:** `docs/handoffs/savedviews-module-refactor-tracker.md`
- **Status:** Planning and documentation only.
- **Allowed change in this session:** This tracker file only.
- **Non-goals:** This task MUST NOT modify production code, tests, Core
  documents, contracts, generated artifacts, package configuration, migrations,
  or harness inputs. It MUST NOT establish product behavior independently of an
  adopted owner.
- **Implementation authorization:** Every implementation slice in Section 7
  requires a later authorized task. Slices that change accepted Incident Bundle
  input additionally require the Core adoption gate in SV-00.

The label is derived from `internal/modules/savedviews`, is already lowercase
kebab case, and contains no path separator, space, shell metacharacter, or
unsafe filename character.

### 1.1 Normative status

| Statement class | Meaning in this tracker | Authority |
| --- | --- | --- |
| Adopted `REQ-*` or `AC-*` statement | The implementation MUST conform now. | Its named Core or adopted subsystem owner |
| `SV-DEC-*` decision | A later savedviews refactor MUST satisfy the decision before the corresponding tracker item can become `DONE`. | This planning handoff only |
| Proposed `REQ-01-644` through `REQ-01-646` or proposed `AC-508` | The content is decision-complete for owner review, but it MUST NOT be implemented as product behavior until adopted. | Pending Core 01/Core 04 adoption |
| Current repository observation | Describes the inspected commit and MUST NOT be treated as a timeless requirement. | Current code, tests, and authored inputs |
| Analysis-note recommendation | Advisory evidence only; it MUST be reconciled against owners and live repository state. | `temp/analysis-notes.md` |

`MUST`, `MUST NOT`, and `MAY` in this tracker bind future slice completion.
They do not promote this tracker above the authority order below.

### 1.2 Adopted planning decisions

| ID | Decision |
| --- | --- |
| SV-DEC-001 | `savedviews` remains the Core-recognized owner of saved-view persistence and route behavior. Its internal policy, application, transport, persistence, startup, and portability responsibilities MUST become cohesive without creating another deployable. |
| SV-DEC-002 | Contract-major-1 `data/saved_views.ndjson` MUST freeze the live member names `saved_view_id` and `scope`. The aliases `id` and `view_scope` MUST be rejected. |
| SV-DEC-003 | After Core adoption, import MUST reject semantically noncanonical display-name, query, layout, and timestamp values. JSON member order and insignificant whitespace MAY vary. |
| SV-DEC-004 | Export MUST include all incident-owned private, shared, and system saved views and MUST emit the required zero-row member when none exist. |
| SV-DEC-005 | Imported private/shared saved views MUST use the dual-identity mapping in Section 4.5: target-local runtime ownership plus preserved portable source ownership. |
| SV-DEC-006 | The savedviews source port MUST perform deterministic explicit export, side-effect-free typed preparation, fixed explicit application, and transaction-scoped invariant validation. |
| SV-DEC-007 | `saved_views.row_shape_exact` MUST become the first of seven savedviews portability invariants after Core adoption, with the failure mapping in Section 4.7. |
| SV-DEC-008 | Saved-view portability MUST NOT carry or infer custom sheets, custom `view_schema` resources, new `sheet_ref` variants, grid state, permissions, workbook preferences, or session-local state. |
| SV-DEC-009 | Ordinary HTTP, startup, storage, authorization, no-op, and negative side-effect behavior MUST remain unchanged by the structural refactor. |
| SV-DEC-010 | Every normative decision MUST map to owner-selected executable evidence before the implementation is complete. Final hashed row IDs MUST be generator-derived, never hand-authored. |

### 1.3 Authority order and evidence reconciliation

The authority order is:

1. Adopted subsystem NLSpecs for their named subsystem. No savedviews-specific
   adopted subsystem NLSpec was found. The adopted Testing Harness NLSpec
   applies only to harness mechanics and verification accounting.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication. Core
   05 is not applicable to this refactor.
4. Domain vocabulary and implementation-support guides.
5. Current repository code and tests for current implementation state.
6. The planning framework, analysis notes, and prior handoffs as evidence only.

No owner contradiction was found. Core 01 recognizes `savedviews` as the owner
for saved-view persistence and route behavior and requires each Incident Bundle
source owner to construct its own portability port.

The following advisory-note mismatches are resolved by live evidence and
SV-DEC-002 through SV-DEC-005:

| Advisory statement | Live repository evidence | Tracker resolution |
| --- | --- | --- |
| Contract-major-1 uses `id` and `view_scope`. | Migration, exporter, importer, descriptor, and tests use `saved_view_id` and `scope`. | Freeze `saved_view_id` and `scope`; reject both advisory aliases. |
| Portability verification uses profile `extension.incident_portability`. | `cartulary.verification_contract.v3` owner files use `profile="base"` for current Incident Portability verification contracts. | Retain the supported machine profile representation; requirements continue to declare `Profiles: incident_portability` in Core. |
| Import rejects noncanonical saved-view values now. | The current local test proves normalization of padded display name, `{}` query, and `{}` layout. | Record that test as pre-change evidence; strict rejection begins only after Core adoption and authorization. |
| Source ownership can be stored unchanged as runtime `owner_user_id`. | `saved_views.owner_user_id` references deployment-local `users`; imported actors are separate descriptors and current fixed-row import maps user fields to the submitter while recording attribution. | Preserve portable source ownership in imported attribution, use a target-local runtime owner, and reconstruct source ownership during re-export. |

Owner and planning documents inspected:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `docs/research/nlspec-spec.md`
- `temp/analysis-notes.md`
- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/domain.md`
- `docs/testing-harness-nlspec.md`

Repository evidence inspected:

- All eight files under `internal/modules/savedviews`
- `internal/modules/incidentportability/portability.go`
- `internal/modules/incidentbundles/source.go`
- `internal/modules/incidentbundles/sourceport/sourceport.go`
- `internal/modules/incidentbundles/sourceport/adapter.go`
- `internal/modules/incidentbundles/worker_service.go`
- `internal/app/server/runtime.go`
- `internal/app/server/runtime_routes.go`
- `internal/app/server/server_profile_harness.go`
- `internal/app/workbookassembly/startup.go`
- `internal/app/incidentportabilityassembly/catalog.go`
- `internal/modules/workbook/coordination_surfaces_test.go`
- `internal/modules/workbook/workbook_startup_test.go`
- `internal/modules/incidentbundles/api_test.go`
- `internal/modules/incidentbundles/routes_integration_test.go`
- `internal/modules/incidentbundles/source_catalog_test.go`
- `internal/modules/codegenboundary/sqlc_boundary_test.go`
- `apps/web/src/workbook/hooks/useWorkbookSavedViewController.ts`
- `apps/web/src/workbook/hooks/useWorkbookSavedViewController.test.tsx`
- `apps/web/src/workbook/models/workbookSavedViews.ts`
- `apps/web/src/workbook/models/workbookSavedViews.test.ts`
- `contracts/openapi-source/owners/module.savedviews/openapi.json`
- `contracts/openapi-source/manifest.json`
- `contracts/verification/owners/module.savedviews.json`
- `contracts/incident-bundles/source_catalog.json`
- `contracts/incident-bundles/traceability.json`
- `db/queries/savedviews.sql`
- `db/migrations/00002_auth_accounts_and_enterprise.sql`
- `db/migrations/00017_saved_views.sql`
- `db/migrations/00023_incident_bundles.sql`
- `tools/test_families/module.savedviews.json`
- `tools/test_catalog_owner.json`
- `tools/backend_module_boundaries.json`
- `tools/schema_object_ownership_manifest.json`
- `tools/schemas/cartulary.incident_bundle_traceability.v1.schema.json`
- Generated SQLC, OpenAPI, catalog, schedule, and topology projections, read
  only to identify downstream surfaces

## 2. Current-State Repository Inventory

Every file currently under `internal/modules/savedviews` is included below.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/savedviews/api.go` | Decodes create, fixture, and patch requests; normalizes query/layout state; applies patches; builds API resources; maps module errors to HTTP errors. | `CreateRequest`, `PatchRequest`, optional patch DTOs, decoders, `ApplyPatch`, `BuildResource`, and error helpers. | Route handlers; workbook coordination test directly uses the create decoder. | `viewquery`, `viewschema`, `httpapi`, JSON, and time. | `savedviews_test.go`; workbook coordination characterization. | Consumes view-schema facts and implements authored savedviews OpenAPI shapes. | `module.savedviews`, split into private domain/application policy and HTTP mapping. | High | Pure state transition and transport mapping share one broad package surface. |
| `internal/modules/savedviews/scope.go` | Owns scope vocabulary, ordinary-create and mutability policy, and display-name normalization. | `Scope`, its three values, parsing/default/mutability helpers, and `NormalizeDisplayName`. | Request decoding, patch/store authorization, and package tests. | `fieldnorm`. | Exact vocabulary, defaulting, authorization, create, and patch tests. | Implements scope enums and constraints represented in authored OpenAPI. | `module.savedviews` domain policy. | Medium | Ownership is correct; later privatization MUST preserve exact wire vocabulary. |
| `internal/modules/savedviews/store.go` | SQLC persistence, visibility, optimistic PATCH transaction, DELETE transaction, conflicts, and workbook-startup lookup. | `Store`, `NewStore`, errors, records/page types, CRUD methods, and startup resolver/reasons. | Route service; workbook-startup assembly; workbook coordination tests. | SQLC, Postgres, `authn.UserRecord`, UUID, and time. | Store and lifecycle tests; workbook startup tests. | Executes authored savedviews SQL through generated SQLC. | `module.savedviews` private persistence adapter. | High | PATCH has a version precondition. DELETE has no version request parameter and MUST remain so unless an owner later changes the route. |
| `internal/modules/savedviews/routes.go` | Registers collection/item routes and a guarded harness-only fixture route; performs authentication, membership, paging, session sliding, envelopes, and error mapping. | `RegisterRoutes` and `RegisterTestRoutes`; current `Service` is private but transport-heavy. | Server runtime and server profile harness. | Platform HTTP/auth/paging/session helpers, incident access, and savedviews store. | Route, authorization, envelope, fixture, conflict, and browser tests. | Must conform to authored savedviews OpenAPI and the harness route contract. | `module.savedviews` transport-adjacent adapter. | High | Route evaluation order is observable and is frozen in Section 4.2. |
| `internal/modules/savedviews/incident_bundle_source_port.go` | Constructs the source-owner descriptor and adapts export, prepare, apply, and validate functions. | `NewIncidentBundleSourcePort`. | Incident Portability assembly and catalog tests. | Incident Bundle source-port types and PostgreSQL transaction. | Local portability tests plus catalog and round-trip integration. | Declares path versions, stable identity `saved_view_id`, dependencies, and six current invariants. | `module.savedviews` source-owner portability contribution. | Critical | The adopted catalog has six invariants; the proposed seventh MUST NOT be declared until Core adoption and matching evidence. |
| `internal/modules/savedviews/incident_bundle_portability.go` | Whole-relation NDJSON export; map-based decode/normalization; generic fixed-row import. | Export/import functions are package-exported; other helpers are private. | Source-port adapter and local tests. | Incident portability codec, `viewquery`, `viewschema`, PostgreSQL, UUID, JSON. | Two local portability tests and broader Incident Bundle integration. | Owns the current `data/saved_views.ndjson` implementation. | `module.savedviews` persistence-adjacent portability adapter. | Critical | Current export uses `to_jsonb(t)` and current apply uses `jsonb_populate_record`; current validation accepts and normalizes noncanonical values. |
| `internal/modules/savedviews/savedviews_test.go` | Characterizes HTTP, storage, OpenAPI, scope, authorization, conflicts, duplication, no-op, fixture, and lifecycle behavior. | Test-only package surface. | Owner test-family rows and broad Go execution. | Savedviews routes/store and backend test support. | This file is the primary ordinary-behavior evidence. | Verifies authored OpenAPI and SQL-backed behavior. | `module.savedviews` test evidence. | High | Direct constructor tests constrain facade narrowing until migrated. |
| `internal/modules/savedviews/incident_bundle_portability_test.go` | Proves current normalization and a small malformed-row set. | Test-only package surface. | Broad Go execution; no dedicated owner row selects either test. | Savedviews portability helpers and Incident Portability errors. | `TestIncidentBundleSavedViewImportValidationNormalizesPortableRows` and `TestIncidentBundleSavedViewImportValidationRejectsMalformedRows`. | Exercises the current row implementation. | `module.savedviews` portability evidence. | Critical | The normalization test is pre-change evidence and conflicts with SV-DEC-003 after adoption; both tests first require explicit owner accounting. |

## 3. Module Boundary Diagnosis

`savedviews` is a Core-recognized owner, not an accidental directory boundary.
The current root package is a mixed-responsibility mutation coordinator with
transport-adjacent, persistence-adjacent, startup, and portability adapters. It
is not a projection owner, frontend controller, custom-sheet owner, or
grid-vendor integration layer.

The durable production seams MUST be:

1. `RegisterRoutes` and `RegisterTestRoutes`;
2. the workbook-startup saved-view resolver;
3. `NewIncidentBundleSourcePort`.

All other exports MUST be privatized after every live caller is migrated or
explicitly justified.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Normative disposition |
| --- | --- | --- | --- | --- | --- |
| Scope, name, identity, mutability, version, and canonical state | `scope.go`, `api.go`, `store.go` | `module.savedviews` domain/application layer | split | Core 01/Core 03 and route/store tests | MUST remain savedviews-owned and MUST be independent of HTTP and SQL types. |
| Request decoding, envelopes, paging, and HTTP errors | `api.go`, `routes.go` | Savedviews transport adapter using platform primitives | split | OpenAPI operations and server composition | MUST remain a thin adapter and preserve Section 4.2 ordering. |
| SQLC persistence, visibility, PATCH, and DELETE | `store.go` | Savedviews private repository | split | Authored SQL, schema ownership, SQLC boundary test | MUST accept stable actor IDs and owner commands, not `authn.UserRecord`. |
| Workbook-startup lookup and reason codes | `store.go`; workbook assembly caller | Savedviews lookup port consumed by workbook assembly | keep | `sheet_ref.kind='saved_view'` startup path | MUST preserve locking, result shape, `saved_view_not_visible`, and `saved_view_not_found`. |
| Incident Bundle contribution | Source-port and portability files | `module.savedviews` source owner | keep | Core 01 source-owner-port model | MUST remain under savedviews; coordinator integration MAY change only through its public source-port interface. |
| Imported actor descriptor and attribution coordination | Incident Bundles source/coordinator | Incident Bundles owns generic actor catalog; savedviews owns its owner mapping | split | Current imported-actor and attribution tables | Savedviews MUST declare the mapping; the coordinator MUST provide preserved actor descriptors without learning saved-view semantics. |
| Harness-only system fixture creation | Routes/API/store | Savedviews adapter under harness guard | keep | Testing Harness NLSpec and route tests | MUST remain unavailable unless the harness profile and token guard authorize it. |
| View-query and view-schema interpretation | Platform owner dependencies | Existing viewquery/viewschema owners | keep | REQ-01-142 and REQ-01-143 | Savedviews MUST reuse these owners and MUST NOT define a parallel grammar. |
| Projection, revision, row mutation, or WebSocket publication | No ordinary CRUD implementation found | Existing projection/revision/entity/collaboration owners | defer | Source inspection and lifecycle tests | Savedviews MUST NOT acquire these side effects without later owner requirements. |
| Frontend saved-view state | Web workbook controller/model | Web workbook shell/controller | keep | Hook/model and browser evidence | The backend refactor MUST NOT move frontend state into the Go module. |
| Custom sheet or grid-vendor state | No target implementation found | Future custom-sheet owner or existing grid adapter | defer | Domain/Core boundary and import scan | This refactor MUST NOT invent either responsibility. |

## 4. Public Contract and Behavior Freeze Map

### 4.1 Observable contract inventory

| Contract | Current owner | Evidence | Existing tests | Required characterization | Refactor risk | Normative freeze |
| --- | --- | --- | --- | --- | --- | --- |
| Saved-view collection GET | `module.savedviews` | REQ-01-144, routes, OpenAPI | Route/store/browser rows | Visibility, ordering, paging, auth, membership, session sliding | High | Route, envelope, default `limit=100`, maximum `500`, and cursor binding MUST remain unchanged. |
| Saved-view collection POST | `module.savedviews` | REQ-01-145, API/store | Create/default/OpenAPI/lifecycle rows | Omitted scope/layout, query limits, normalization, owner, status | High | Ordinary create MUST default omitted scope to private and MUST reject system scope. |
| Saved-view item PATCH | `module.savedviews` | REQ-01-146, API/store | Patch/conflict/no-op/integration rows | Identity immutability, error precedence, version, no-op | Critical | Required `base_saved_view_version`, structural no-op, and one-step version/timestamp advance MUST remain unchanged. |
| Saved-view item DELETE | `module.savedviews` | REQ-01-147, route/store | Authorization/lifecycle/data-preservation rows | Visibility, mutability, response, configuration-only delete | Critical | DELETE has no request-body version precondition; it MUST delete only the configuration object. |
| Scope and authorization | Savedviews under Core 04/Core 03 | Scope/store/routes | Vocabulary and authorization rows | Owner/member/admin matrix for private/shared/system | Critical | Scope controls only saved-view discoverability and mutability. |
| Query/layout state | Viewquery/viewschema owners consumed by savedviews | REQ-01-142/143 | Create/patch/store/portability rows | Direct and portability parity | High | Ordinary routes retain normalization; strict portability import uses canonical-equality validation after adoption. |
| Workbook startup | Workbook owner consuming savedviews resolver | Startup assembly/tests | Startup and stateful browser rows | Found/not-found/not-visible, lock, selected state | Critical | `sheet_ref.kind='saved_view'`, reason codes, and collaboration identity MUST remain unchanged. |
| Harness system fixture | Testing Harness plus savedviews | Test route and harness NLSpec | Unavailable-by-default and token cases | Profile/token/path/system tuple | High | The route MUST NOT become an ordinary production create path. |
| Incident Bundle saved views | Savedviews source owner | REQ-01-639/640, port descriptor | Local tests and aggregate round trip | Sections 4.3 through 4.7 | Critical | Proposed strict behavior MUST wait for SV-00 adoption. |
| OpenAPI and SQLC projections | Authored savedviews inputs and generators | OpenAPI owner JSON, savedviews SQL | Compatibility and store tests | No unexpected source/generated drift | High | Generated outputs MUST be changed only through their owners. |
| Frontend controller/model | Web workbook owner | Hook/model/browser rows | Unit and browser rows | Request timing and selected/default/startup state | High | No public frontend or grid-adapter interface change is planned. |
| Negative side effects | Projection/revision/entity/collaboration owners | No target publisher/mutator | Lifecycle and delete evidence | Service-port spies after extraction | High | Ordinary CRUD MUST NOT refresh projections, create revisions/change sets, mutate rows, or publish WebSocket events. |

### 4.2 Route evaluation order

The structural refactor MUST preserve this observable precedence:

| Operation | Required order |
| --- | --- |
| Collection GET | Parse incident path; authenticate; resolve and validate cursor/paging; require membership; list visible rows; slide session; write paged response. |
| Collection POST | Parse incident path; authenticate with state-changing policy; require membership; decode and normalize request; create; slide session; write `201`. |
| Item PATCH | Parse incident and saved-view paths; authenticate with state-changing policy; require membership; resolve current visible row to obtain its immutable schema; decode request; lock and re-read; check mutability; check base version; apply normalized transition; commit only a material change; slide session; write `200`. |
| Item DELETE | Parse incident and saved-view paths; authenticate with state-changing policy; require membership; lock visible row; check mutability; delete only the saved-view row; slide session; write `200`. |
| Harness system POST | Enforce test-route enablement and guard; parse path; decode fixture request; create system tuple; write `201`. |

Malformed path identifiers continue to produce the existing not-found behavior.
A visible shared view submitted by a non-owner member continues to reach request
decoding before mutation authorization is decided. A valid PATCH MUST decide
mutation denial before version conflict and version conflict before no-op.

### 4.3 Proposed exact `data/saved_views.ndjson` row

This subsection implements SV-DEC-002 through SV-DEC-004 as proposed owner
content. It is not active product authority until SV-00 is adopted.

For bundle versions `1` and `2`, source-family contract major `1`, every line
MUST contain exactly the following eleven required members:

| Member | JSON type and required contract |
| --- | --- |
| `saved_view_id` | String containing a valid UUID; stable identity; unique within the file and target deployment. |
| `incident_id` | String containing a valid UUID equal to the manifest and immutable import-context incident ID. |
| `view_schema_id` | Non-empty string naming an admitted registered schema; immutable for the saved view. |
| `scope` | String equal to `private`, `shared`, or `system`. |
| `display_name` | String already equal to its `display_name_line_v1` canonical result. |
| `query_json` | Non-null object already structurally equal to the REQ-01-142 canonical result for `view_schema_id`. |
| `layout_json` | Non-null canonical `cartulary.layout.v1` object satisfying REQ-01-143; `{}` is invalid. |
| `owner_user_id` | UUID string for private/shared rows; JSON `null` for system rows. |
| `created_at` | Canonical UTC RFC3339Nano string using `Z`. |
| `updated_at` | Canonical UTC RFC3339Nano string using `Z`; instant MUST be greater than or equal to `created_at`. |
| `saved_view_version` | JSON integer greater than or equal to `1`; fractional and string encodings are invalid. |

The importer MUST reject `id`, `view_scope`, every other alias, every unknown
member, every missing member, every duplicate member at any object depth, every
wrong JSON type, and every prohibited `null`.

### 4.4 Defaults, canonicality, and boundaries

| Boundary | Required behavior |
| --- | --- |
| Export bundle version | Newly generated bundles use version `2`; version `1` remains import-only under REQ-01-635/636. No version fallback exists. |
| Required path | `data/saved_views.ndjson` MUST be present exactly once for admitted versions `1` and `2`. |
| Zero rows | Export MUST emit the required path with a zero-byte payload. Import MUST interpret that payload as zero rows. |
| Row order | Export MUST order rows by `saved_view_id` ascending. |
| JSON output | Each exported row MUST use repository canonical JSON member ordering and exactly one trailing LF. Two exports from identical state and inputs MUST be byte-identical. |
| JSON input formatting | Member order and insignificant JSON whitespace MAY vary. The importer MUST reject blank logical rows, more than one JSON value on a line, trailing non-whitespace content, and malformed JSON. |
| Row bound | Preparation MUST use the established bounded accessor and MUST reject a logical line larger than `16 MiB`; global archive and extracted-byte limits remain owned by Incident Bundles. |
| Display name | Import MUST compare the submitted string with `NormalizeDisplayName` output and reject inequality; it MUST NOT trim or normalize into acceptance. |
| Query | `sort` and `filters` MUST be present arrays; inactive `group_by` MUST be omitted; raw sort count MUST NOT exceed `8`; raw filter count MUST NOT exceed `16`; canonical value MUST be structurally equal to REQ-01-142 output. |
| Layout | All four REQ-01-143 members MUST be present; widths remain in `40..4096`; canonical value MUST be structurally equal to REQ-01-143 output. |
| Timestamp | Input MUST parse as RFC3339Nano, use UTC `Z`, and equal its canonical UTC RFC3339Nano serialization. |
| Invalid row | The source family MUST fail as a whole. No row may be skipped, repaired, or partially applied. |

The row MUST NOT contain custom-sheet definitions, custom `view_schema`
definitions, new `sheet_ref` variants, membership or permission objects,
workbook preferences, arbitrary columns, grid-vendor state, selection, scroll,
focused cell, inspector state, presence, or another session-local value.

A registered base/current-profile schema is valid. A claimed extension schema
is valid only when its adopted owner contract admits it for the target profile.
An unknown schema is invalid. The importer MUST NOT infer a schema or custom
sheet from query state, layout state, a display name, visible labels, or unknown
members.

Missing optional Reference Packs MAY degrade only their admitted dependent
overlay behavior. The saved-view row, schema identity, scope, ownership, query,
and layout MUST remain unchanged. An unknown base/current-profile schema,
malformed query, or malformed layout is not bounded degradation and MUST fail.

### 4.5 Portable and runtime owner mapping

| Row case | Portable `owner_user_id` | Runtime `saved_views.owner_user_id` | Attribution and re-export | Authorization effects |
| --- | --- | --- | --- | --- |
| Native private/shared | Native deployment-local owner UUID | Same UUID | No imported-owner attribution is required; export uses runtime owner. | Existing saved-view scope rules apply. |
| Imported private/shared | Source owner UUID from the bundle | Target-local import submitter UUID | Apply MUST record `(saved_views, saved_view_id, owner_user_id, source owner UUID, local submitter UUID)`. Re-export MUST prefer that preserved source UUID and its imported-actor descriptor. | Import MUST create no membership, login capability, credential, provider binding, deployment role, or session. |
| Native or imported system | JSON `null` | SQL `NULL` | No owner attribution is emitted. | Ordinary routes remain unable to mutate the system row. |
| Duplicate created after import | Duplicating actor UUID on later export | Duplicating actor UUID | The new identity MUST NOT inherit imported-owner attribution from its source. | Ordinary create/duplicate scope rules apply. |

Every private/shared source owner MUST have exactly one valid actor descriptor.
Missing, malformed, or duplicate descriptors MUST fail
`saved_views.owner_tuple_legal`. Re-export actor collection MUST resolve native
users and imported descriptors without duplicate actor identities. It MUST NOT
replace the preserved source actor with the import submitter in the portable
row.

### 4.6 Source-port behavioral interface

| Operation | Input | Output | Required behavior |
| --- | --- | --- | --- |
| Descriptor | None | Immutable descriptor | MUST declare family `saved_views`, contract major `1`, path for versions `1` and `2`, stable identity `saved_view_id`, dependency `revisions`, owner `module.savedviews`, owned relation ID, and all seven adopted invariant IDs in the declared order. |
| Export | Read-only query capability and `incident_id` | Exactly one deterministic file | MUST select the eleven fields explicitly, resolve portable owner identity, include all scopes, canonicalize, sort, and emit the zero-row file. MUST NOT use whole-relation serialization. |
| Prepare import | Bounded bundle capability and immutable import context | Port-bound opaque prepared saved-view value or typed failure | MUST perform strict duplicate-aware decoding, exact-shape validation, canonical-equality validation, actor-descriptor resolution, and row-local invariant checks. MUST perform no database, object-store, or visible-state mutation. |
| Apply import transaction | Supplied transaction, matching prepared value, immutable import context | Success or typed failure | MUST use fixed owner-controlled SQL/SQLC with an explicit column list, write only savedviews-owned runtime state and admitted attribution through the supplied recorder, map runtime owners exactly once, and require one affected row per prepared row. |
| Validate import transaction | Supplied transaction and immutable import context | Success or typed failure | MUST prove all seven savedviews invariants over the imported incident before publication and MUST NOT treat successful conversion as sufficient validation. |

Prepare MUST own decoding. Apply MUST NOT decode source bytes. Apply MUST NOT
use `jsonb_populate_record`, descriptor-derived SQL, generic
`ON CONFLICT DO NOTHING`, owner reassignment beyond Section 4.5, scope repair,
or silent row skipping.

### 4.7 Invariant and public-failure mapping

| Invariant ID | Exact proof obligation | Representative failures |
| --- | --- | --- |
| `saved_views.row_shape_exact` | Exactly eleven required members; legal JSON types/nullability; no alias, unknown, missing, or duplicate member; one bounded object per logical row. | `id`, `view_scope`, unknown nested key, missing key, duplicate key, blank line, multi-value line |
| `saved_views.identity_scope_legal` | Valid unique `saved_view_id`; incident match; admitted immutable schema; closed scope. | Duplicate ID, foreign incident, unknown schema, invalid scope |
| `saved_views.owner_tuple_legal` | Private/shared portable owner UUID and descriptor exist; runtime mapping/attribution agree; system owner is null; no authorization state is synthesized. | Missing actor, non-UUID owner, non-null system owner, attribution mismatch |
| `saved_views.display_name_normalized` | Submitted value already equals `display_name_line_v1` canonical output. | Padded, non-NFC, empty-after-normalization, control character, over-limit name |
| `saved_views.query_layout_legal` | Query and layout are closed, canonical, schema-valid, and contain no technical or transient state. | `{}` query/layout, omitted canonical arrays, null grouping, unknown field, invalid width |
| `saved_views.version_timestamps_legal` | Version is a positive integer; timestamps are canonical UTC RFC3339Nano; `updated_at >= created_at`; deterministic export normalization holds. | Zero/fractional version, offset timestamp, noncanonical timestamp, reversed time |
| `saved_views.reference_pack_degradation_bounded` | Only optional pack-dependent behavior degrades; core row and other incident state are preserved unchanged. | Schema substitution, filter deletion, row skip, unrelated-state mutation |

Every invariant failure MUST surface publicly as:

```json
{
  "error": {
    "code": "incident_bundle_import_rejected",
    "details": {
      "invariant_id": "<exact saved_views invariant>",
      "source_family_id": "saved_views"
    },
    "reason_code": "source_family_invalid",
    "retryable": false
  }
}
```

Public errors, job results, logs, telemetry, readiness, administrative
summaries, and operator output MUST NOT disclose raw row values, hostile member
names, SQL, relation names, credentials, provider subjects, object keys,
staging identifiers, paths, or topology. A retained internal diagnostic MAY
contain only the fixed logical path, one-based row ordinal, and a closed token
such as `unknown_member`; it MUST NOT retain the hostile name or value.

### 4.8 Proposed private application interface

No public HTTP interface changes. The later implementation MAY choose private
Go type names, but it MUST implement these behavioral seams:

| Operation | Inputs | Output | Required errors and effects |
| --- | --- | --- | --- |
| List visible | Context, incident UUID, actor UUID, validated page request | Ordered page of records | Repository failure only; no mutation or session behavior. |
| Create | Context, incident UUID, actor UUID, normalized create command, commit time | Created record | Actor UUID becomes owner for ordinary private/shared create; no projection/revision/WebSocket effect. |
| Patch | Context, incident UUID, saved-view UUID, actor UUID, membership role, normalized patch, commit time | Current or updated record | Not found/not visible, mutation denied, version conflict, or repository failure in Section 4.2 order; no-op returns current record without timestamp/version change. |
| Delete | Context, incident UUID, saved-view UUID, actor UUID, membership role | Deleted identity | Not found/not visible, mutation denied, or repository failure; no version input and no underlying-row effect. |
| Resolve startup | Context, supplied transaction, incident UUID, saved-view UUID, actor UUID | Record, closed reason, or failure | Exact reasons are empty on success, `saved_view_not_visible`, or `saved_view_not_found`. |

Transport code owns HTTP decoding, envelopes, authentication, membership
lookup, cursor tokens, test-route guarding, and session sliding. The application
service owns saved-view policy ordering. The private repository owns SQLC and
transaction mechanics and MUST accept actor UUIDs rather than
`authn.UserRecord`.

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| `api.go` mixes pure state policy with HTTP decoding and errors. | Source inspection | Wire behavior can change during extraction. | `should_fix` | Savedviews application and transport layers | Characterize equivalence, then extract pure policy without changing route mapping. |
| Store methods accept `authn.UserRecord` although persistence requires stable actor identity. | `store.go` | Platform auth representation leaks into persistence. | `should_fix` | Savedviews service/repository boundary | Pass actor UUIDs through the private repository interface. |
| Authorization, visibility, version, and no-op decisions span handler, helper, and transaction code. | Routes/API/store | Reordering can disclose existence or change errors. | `should_fix` | Savedviews application service | Preserve Section 4.2 precedence with focused tests before consolidation. |
| The root exports implementation-only DTOs, constructors, records, and helpers. | Export scan and workbook coordination test | Internals are coupled across package boundaries. | `should_fix` | Savedviews facade | Migrate callers before removing any export; retain the three production seams. |
| Whole-relation `to_jsonb(t)` defines current bundle shape accidentally. | Portability exporter | A storage migration can change the wire contract. | `must_fix` | Savedviews portability adapter | After SV-00/SV-02, replace with explicit field selection and DTO mapping. |
| Map decode plus `jsonb_populate_record` cannot prove the proposed exact contract. | Portability importer and shared codec | Unknown/duplicate/noncanonical content can be accepted or lost. | `must_fix` | Savedviews port using admitted shared bounded codec | Move strict decoding to Prepare and fixed explicit persistence to Apply. |
| Current import normalizes values that SV-DEC-003 requires rejecting. | Existing normalization test and implementation | This is a deliberate compatibility change. | `must_fix` | Core 01/Core 04, then savedviews | Preserve pre-change evidence; do not implement until owner adoption and authorization. |
| Current runtime owner remap is not reversed on re-export. | Attribution buffer, fixed-row import, actor export | Portable source ownership can drift after a round trip. | `must_fix` | Savedviews mapping plus Incident Bundles actor collection | Implement Section 4.5 and prove portable/runtime identities separately. |
| The descriptor has six invariants and aggregate validation directly proves only an owner tuple subset. | Source-port descriptor and Validate function | Declared claims lack one-to-one proof. | `must_fix` | Savedviews source owner | Adopt the seventh invariant and map all seven to code and selected evidence. |
| Both local portability tests are absent from the owner family. | Exact names versus test-family manifest | Focused owner validation omits them. | `must_fix` | Authored savedviews harness input | Add distinct active rows before structural portability changes. |
| Verification owner files do not carry requirement IDs directly. | Verification schema v3 and Incident Bundle traceability projection | Adding unsupported fields would make invalid authored input. | `intentional/no_action` | Traceability projection owner | Put requirement-to-AC-to-verification mappings in the Incident Bundle traceability projection. |
| Platform HTTP/auth/paging imports are confined to the route adapter. | Route imports and repository pattern | Moving them creates a forwarding facade. | `intentional/no_action` | Savedviews transport adapter | Keep them out of pure policy; do not relocate the adapter. |
| SQLC is confined to savedviews persistence. | Store/import files and SQLC boundary test | Moving SQL into policy erodes ownership. | `intentional/no_action` | Savedviews persistence adapters | Preserve confinement and run the boundary check. |
| The source port remains under savedviews as Core requires. | REQ-01-639 and assembly catalog | Relocation would invert ownership. | `intentional/no_action` | `module.savedviews` | Keep `NewIncidentBundleSourcePort` as the aggregate seam. |
| No direct grid-vendor, projection, revision, entity mutation, or WebSocket publisher was found. | Import/source scan | Speculative ports would expand scope. | `intentional/no_action` | Existing owners | Freeze absence of those dependencies and effects. |
| Generated files are downstream of authored owner inputs. | Generated-artifact policy | Hand edits drift and are overwritten. | `intentional/no_action` | Owning generators | Change authored inputs and use Make-owned generation only. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved in later authorized work | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Source bootstrap and tracker revision | root | None | WF-01 | Establish authority, decisions, scope, and current evidence. | This tracker only in the current task | `make lint-markdown` | All prior history retained; current revision recorded. |
| WF-01 | Live inventory and pre-change baseline | chain | WF-00 | WF-02, WF-03, WF-04 | Preserve the eight-file inventory and record current wire, normalization, actor, and test behavior. | Target package, callers, current tests | Existing focused baseline and read-only discovery | No current behavior is inferred from advisory prose. |
| WF-02 | Owner-contract adoption | chain | WF-01 | WF-03, WF-06, WF-07 | Adopt the exact row, strict canonical policy, dual identity, seventh invariant, and binary AC. | Core 00/01/02/04 and typed traceability inputs | Markdown, JSON shape, traceability, generated drift | `REQ-01-644..646` and `AC-508` are adopted or implementation remains blocked. |
| WF-03 | Characterization and owner accounting | chain | WF-01, WF-02 | WF-06, WF-07 | Make pre-change and new conformance evidence owner-selectable. | Savedviews tests, authored owner family, verification/traceability inputs | Focused and service-backed owner slices; harness contract | Generated row IDs and baseline run roots are recorded. |
| WF-04 | Boundary and coupling scan | parallel | WF-01 | WF-05, WF-06 | Preserve legitimate adapters and identify seams that require extraction. | Target, assembly, boundary manifests, frontend imports | Backend boundary; frontend boundary only if touched | Every finding has one classification and owner. |
| WF-05 | Ordinary CRUD cohesion | chain | WF-03, WF-04 | WF-08 | Extract policy/service/repository seams without changing behavior. | API/scope/store/routes and direct tests/callers | Exact route/store/startup rows | Public seams and evaluation order remain unchanged. |
| WF-06 | Strict portability implementation | chain | WF-02, WF-03, WF-04 | WF-07, WF-08 | Implement explicit export, strict Prepare, fixed Apply, Validate, and dual identity. | Savedviews portability, source port, admitted shared actor/codec integration | Portability unit/integration/invariant matrix | All seven invariants and round-trip identities are proven. |
| WF-07 | Typed projections and generated accounting | chain | WF-02, WF-03, WF-06 | WF-08 | Update source catalog, traceability, schema projection, and generated harness artifacts. | Authored contracts/manifests and generated outputs via Make | JSON shape, harness contract, generation drift, artifact policy | No orphan requirement, AC, invariant, verification, or row. |
| WF-08 | Final validation and handoff | chain | WF-05, WF-06, WF-07 | None | Run narrow-to-broad gates and record exact evidence. | Later implementation diff and tracker | `make agent-finalize` before `make check` | Exact commit, row IDs, run roots, failures, and skipped checks are recorded. |

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| SV-00 | WF-02 | **Requires later authorization:** adopt proposed `REQ-01-644..646`, extend REQ-01-640, add `AC-508`, clarify REQ-02-152, and extend typed traceability schemas/mappings. | Core 00/01/02/04; Incident Bundle traceability authored inputs | Owner contradiction or accidental authority inversion | Binary owner review and traceability validation | `make lint-markdown`; `make json-shape-check`; `make generate-drift` | Revert the owner revision and all downstream projections as one change. | Owners explicitly adopt the live-key exact row, strict rejection, dual identity, seven invariants, and AC; otherwise SV-03 through SV-06 remain blocked. |
| SV-01 | SV-00 | Add distinct active owner rows for both current portability tests and record the pre-change normalization baseline. | Authored savedviews test-family input; existing tests; generated harness outputs via Make | A selector can silently miss a renamed test. | Select exactly `TestIncidentBundleSavedViewImportValidationNormalizesPortableRows` and `TestIncidentBundleSavedViewImportValidationRejectsMalformedRows`. | `make test-slice OWNER=module.savedviews ROWS=TODO:generator-derived-current-portability-row-ids`; `make harness-contract` | Revert authored rows and regenerate; do not hand-edit generated outputs. | `make explain-test-owner` lists two distinct rows, exact selection passes, and deletion/rename fails catalog validation. |
| SV-02 | SV-01 | Capture pre-change export evidence for the live eleven keys, private/shared/system rows, order, and zero-row path. | Savedviews portability tests and Incident Bundle test support | Golden data can accidentally encode unstable storage or timestamps. | Add deterministic semantic/golden export tests with fixed IDs/times and no advisory aliases. | `make test-slice OWNER=module.savedviews ROWS=TODO:generator-derived-export-baseline-row-ids` | Remove only the new characterization fixtures/tests; runtime remains unchanged. | Current export is reproducibly characterized before the behavior change. |
| SV-03 | SV-02 | **Requires later authorization:** replace the normalization test with strict noncanonical-rejection evidence and add exact-shape, alias, unknown/duplicate, scope, schema, pack, determinism, ownership, and atomicity cases. | Savedviews unit/integration tests, owner-family input, verification/traceability input | Tests can encode behavior not adopted by Core. | Exact matrix in Sections 4.3 through 4.7; retain the pre-change run root in handoff history. | `make test-slice OWNER=module.savedviews ROWS=TODO:generator-derived-strict-unit-row-ids`; `make service-backed-test-slice OWNER=module.savedviews ROWS=TODO:generator-derived-strict-integration-row-ids` | Revert new expectations independently of implementation; never erase baseline evidence. | Every SV-DEC requirement has at least one active row and every row selects an executable test. |
| SV-04 | SV-03 | **Requires later authorization:** replace whole-relation export with explicit field selection, canonical DTO mapping, deterministic zero-row output, and source-owner reconstruction. | Savedviews portability; Incident Bundles actor collection through an owner-approved interface | Wire drift, actor descriptor loss, nondeterminism | Export shape, byte determinism, native/imported/system ownership, round trip | Focused export rows plus owner-selected Incident Bundle integration | Retain old exporter behind test-only equivalence until valid current bundles match; revert as one slice. | No `to_jsonb(saved_views)` equivalent remains; valid export and re-export satisfy Sections 4.3 through 4.5. |
| SV-05 | SV-03, SV-04 | **Requires later authorization:** move strict bounded duplicate-aware decode and canonical-equality validation into Prepare and return a port-bound typed prepared value. | Savedviews port/portability; admitted shared bounded codec only where generic | Previously accepted malformed or noncanonical bundles fail. | Shape, duplicate, alias, canonicality, actor, schema, safe-error rows | Focused strict unit rows and `make service-backed-test-slice` for coordinator mapping | Keep the old decoder unavailable to Apply but revert the Prepare replacement independently. | Prepare is side-effect-free, rejects every mapped defect, and returns no raw map-based contract. |
| SV-06 | SV-05 | **Requires later authorization:** replace generic record population with explicit Apply; implement seven-invariant Validate and dual-identity attribution. | Savedviews SQL/portability/source port; actor attribution integration | Transactionality, affected-row equality, owner remap, invariant-to-error mapping | Round trip, no synthesized authorization, affected-row/duplicate failure, all invariants, rollback | Service-backed savedviews rows plus owner-selected Incident Bundle atomicity/safe-failure rows | Revert Apply/Validate together; no migration is planned for the selected dual-identity model. | No generic population/conflict-ignore path remains; all seven invariants pass or fail safely with no visible state. |
| SV-10 | SV-01, WF-04 | Separate pure scope/name/query/layout/patch policy from HTTP decoding/resource/error mapping. | `api.go`, `scope.go`, private savedviews files, ordinary tests | HTTP field/reason mapping and no-op semantics | Preserve scope, create, patch, OpenAPI, and no-op rows | Existing exact unit/store owner rows | Keep compatibility wrappers until all package callers migrate. | Pure policy imports no HTTP, SQLC, Postgres, or auth record type and produces equivalent results. |
| SV-11 | SV-10 | Introduce the private application service and repository interface; replace persistence `authn.UserRecord` inputs with UUIDs. | `routes.go`, `store.go`, private service/repository files | Authorization/error precedence and transaction ordering | Service fakes plus existing store/lifecycle rows | Existing exact store and lifecycle rows through `make service-backed-test-slice` | Adapt one handler at a time and retain the old store path until evidence passes. | Section 4.8 seams exist; routes delegate; SQLC remains private; behavior is unchanged. |
| SV-12 | SV-11 | Preserve the startup resolver seam while hiding raw lookup/record mapping behind the repository. | Store/repository; workbook startup assembly only if call adaptation is required | Startup identity, locking, reason codes | Workbook startup and stateful browser evidence | Owner-selected startup rows; broaden to `make browser-e2e-stateful` when risk requires | Retain a wrapper with the current signature until startup evidence passes. | Startup results and reasons are unchanged and assembly imports only the justified seam. |
| SV-13 | SV-11, SV-12 | Replace workbook coordination test use of `NewStore`/decoder with public-route characterization and privatize unjustified exports. | Workbook coordination test; savedviews tests/API/store/scope | Coverage loss and package compatibility | Route/OpenAPI/lifecycle/workbook coordination evidence | Focused/service-backed savedviews rows and `make backend-module-boundary-check` | Migrate tests before removing exports; restore a wrapper if a live caller remains. | Only the three durable production seams and explicitly justified types remain exported. |
| SV-90 | SV-06, SV-13 | Update authored source catalog, exact-row machine schema, traceability, and harness inputs; regenerate downstream outputs only through Make. | `contracts/incident-bundles`, `contracts/verification`, `tools/test_families`, schema inputs, generated outputs | Orphan mapping, manual generated edit, unsupported schema field | JSON/schema, traceability, catalog, selector, generated policy evidence | `make json-shape-check`; `make harness-contract`; `make generate-drift`; `make generated-artifact-policy-check` | Revert authored inputs and rerun the owning generator. | Every owner requirement maps to AC-508 and selected verification; every invariant maps to active evidence; generated drift is clean. |
| SV-99 | SV-90 | Execute final narrow-to-broad validation and update every handoff table with exact evidence. | Entire later authorized diff and tracker | Stale evidence or hidden skipped checks | All applicable owner rows and broad gates | `make agent-finalize`; then `make check`; conditional browser targets | Each preceding slice remains independently revertible. | Exact commit, generated row IDs, run roots, outcomes, failures, and skips are recorded. |

## 8. Validation Plan

Except for the recorded historical baseline and Markdown lint runs, these
commands are future gates and have not been claimed as passing.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| owner adoption | `make lint-markdown`; `make json-shape-check` | Core prose and typed traceability/schema inputs | yes | Required by SV-00 before behavior-changing tests or code. |
| unit | `make test-slice OWNER=module.savedviews ROWS=<generator-derived-row-ids>` | Pure policy, shape, canonicality, and descriptor tests | yes | Exact IDs MUST be recorded after SV-01/SV-03 generation. |
| integration | `make service-backed-test-slice OWNER=module.savedviews ROWS=<generator-derived-row-ids>` | SQL-backed CRUD, owner mapping, import/apply/validate, startup | yes | Use only rows relevant to the slice. |
| Incident Bundle/harness | `make harness-contract` plus owner-selected Incident Bundle rows | Catalog, selectors, traceability, safe errors, transactionality | yes | Required when owner rows, traceability, source catalog, or coordinator behavior changes. |
| browser | Owner-selected savedviews browser rows; then `make browser-e2e-webserver-backed` and `make browser-e2e-stateful` when risk requires | Route, controller, persistence, default/startup, replay | no | Required for route wiring, startup, or frontend-visible changes; not required for isolated portability internals without browser effects. |
| generated drift | `make generate-drift`; `make generated-artifact-policy-check` | Generated code/contracts/harness projections | no | Required whenever an authored generator input changes. |
| contract compatibility | `make openapi-compatibility-check` | Public HTTP contract | no | Conditional because no OpenAPI change is planned. |
| import-boundary/static | `make backend-module-boundary-check`; add `make frontend-import-boundary-check` only if frontend files change | Backend and conditional frontend imports | yes | Run before and after boundary work. |
| full check | `make agent-finalize` followed by `make check` | Final implementation gate | no | Run only after focused checks; report run roots and unrelated failures. |
| tracker documentation | `make lint-markdown` | This tracker-only revision | no | This is the only execution validation required in the current task. |

Historical evidence retained:

- `make test-slice OWNER=module.savedviews
  ROWS=module.savedviews.unit.saved_view_scope_vocabulary_is_exactly_private_s_aa8991ad84`
  passed at commit `4e42f709d875ead7f977f95c6e3433e556ad8388`;
  run root `.cartulary/test-results/20260729T060105Z-p542483`.
- The original tracker revision passed `make lint-markdown`; recorded run root
  `.cartulary/test-results/20260729T061246Z-p548044`.

Command discovery used `make task-guide ROLE=module-author
OWNER=module.savedviews`, `make explain-test-owner OWNER=module.savedviews`,
`make explain-target TARGET=<target> DETAIL=summary`, and `make help-all`.

The revised tracker passed `make lint-markdown`; run root:
`.cartulary/test-results/20260729T071524Z-p569244`.

## 9. Top-Level Work Tracker

Only `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`, and `DROPPED` are
valid statuses in this table.

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| SVT-001 | Establish authority, scope, and planning-only posture | WF-00 | DONE | None | Section 1 | Authority, Core 05 disposition, and single-file scope are explicit. |
| SVT-002 | Inventory all eight target files and live relationships | WF-01 | DONE | SVT-001 | Section 2 | No target file is omitted. |
| SVT-003 | Diagnose the durable owner and mixed internal responsibilities | WF-04 | DONE | SVT-002 | Sections 3 and 5 | Legitimate adapters and required splits have one owner/disposition. |
| SVT-004 | Freeze ordinary HTTP/startup/storage/side-effect behavior | WF-01 | DONE | SVT-002 | Sections 4.1, 4.2, 4.8 | Route precedence, defaults, errors, and negative effects are explicit. |
| SVT-005 | Resolve exact keys, strict canonical policy, dual identity, and invariant plan | WF-00, WF-02 | DONE | SVT-001 | SV-DEC-002 through SV-DEC-007; RB-001 | Planning choices are binary and repository mismatches are reconciled. |
| SVT-006 | Adopt proposed Core and acceptance requirements | WF-02 | BLOCKED | SVT-005 | Planned SV-00 | Later owner-document authorization adopts REQ-01-644..646 and AC-508. |
| SVT-007 | Add owner rows for both current portability tests | WF-03 | TODO | SVT-006 | Planned SV-01; RB-002 | Two distinct generated row IDs select and pass the exact existing tests. |
| SVT-008 | Capture current live eleven-key export baseline | WF-03 | TODO | SVT-007 | Planned SV-02 | Private/shared/system/zero-row current output is retained deterministically. |
| SVT-009 | Add strict contract and invariant evidence | WF-03 | TODO | SVT-008 | Planned SV-03 | Every SV-DEC and all seven invariants map to active executable rows. |
| SVT-010 | Implement explicit deterministic export and portable-owner reconstruction | WF-06 | BLOCKED | SVT-009, SVT-006 | Planned SV-04 | No whole-relation export remains and portable identity round-trips. |
| SVT-011 | Implement strict typed Prepare | WF-06 | BLOCKED | SVT-010, SVT-006 | Planned SV-05 | Prepare is bounded, strict, typed, side-effect-free, and owns decode. |
| SVT-012 | Implement explicit Apply and seven-invariant Validate | WF-06 | BLOCKED | SVT-011, SVT-006 | Planned SV-06 | Fixed persistence and invariant mapping pass with no visible failure state. |
| SVT-013 | Extract ordinary saved-view policy from HTTP mapping | WF-05 | TODO | SVT-007 | Planned SV-10 | Pure policy produces equivalent route/store outcomes. |
| SVT-014 | Introduce application service and private repository | WF-05 | TODO | SVT-013 | Planned SV-11 | Routes delegate and persistence accepts UUIDs, not auth records. |
| SVT-015 | Preserve/narrow startup seam and root facade | WF-05 | TODO | SVT-014 | Planned SV-12/SV-13 | Startup evidence passes and unjustified exports are removed after caller migration. |
| SVT-016 | Update typed projections and generated accounting | WF-07 | TODO | SVT-012, SVT-015 | Planned SV-90 | Traceability, catalog, schema, rows, and generated outputs are complete and clean. |
| SVT-017 | Execute final gates and implementation handoff | WF-08 | TODO | SVT-016 | Planned SV-99 | Exact evidence for the final commit is recorded without rediscovery. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-29T02:04:57-04:00 | Codex planning session; `main` at `4e42f709d875ead7f977f95c6e3433e556ad8388` | Tracker created from live repository evidence; planning-only scope enforced. | Inspected planning framework, Core 00-04, domain guide, Testing Harness NLSpec, target/callers; touched only this tracker. | `git status --short --branch`; `git rev-parse HEAD`; read-only `rg`, `sed`, `wc`, and `jq`; `make lint-markdown`. | Target exists; tracker was absent; eight files found; label/output safe; no owner contradiction; Core 05 not applicable; lint passed at `.cartulary/test-results/20260729T061246Z-p548044`. | Later implementation is outside that task's authority. | Authorize an implementation session beginning with owner accounting. |
| 2026-07-29T03:07:46-04:00 | Codex NLSpec-style revision; `main` at `4e42f709d875ead7f977f95c6e3433e556ad8388` | Tracker decisions are precise; product behavior remains pending owner adoption. | Inspected NLSpec guidance, advisory analysis, Core 00-04, domain, target and portability/actor/harness surfaces; touched only this tracker. | Read-only `git`, `date`, `sha256sum`, `rg`, `sed`, `jq`, `find`, Make help/task-guide/explain commands; final lint recorded below. | Live-key, strict-canonical, dual-identity, invariant, and authority decisions are reconciled without elevating advisory prose. | SV-00 and later behavior-changing work require separate authorization. | Seek Core owner adoption of proposed REQ-01-644..646 and AC-508. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-29T02:04:57-04:00 | Codex planning session; `main` at `4e42f709d875ead7f977f95c6e3433e556ad8388` | Legitimate savedviews owner diagnosed as a mixed-responsibility root with transport and persistence adapters. | Inspected all target files, server/workbook/portability assembly callers, codegen boundary, backend boundary manifest; touched only this tracker. | Read-only symbol/import/caller searches; `make explain-target TARGET=backend-module-boundary-check DETAIL=summary`. | Keep route, startup, and source-port seams; split policy/service/persistence internally; no projection/revision/WebSocket responsibility found. | No planning blocker; implementation authorization required. | Characterize, then execute the internal cohesion slices. |
| 2026-07-29T03:07:46-04:00 | Codex NLSpec-style revision; `main` at `4e42f709d875ead7f977f95c6e3433e556ad8388` | Private application/repository and source-port interfaces are behaviorally specified. | Inspected routes, API, scope, store, source port, portability code, incidentportability codec, sourceport adapter, actor coordinator; touched only this tracker. | Read-only source/import/interface searches. | Three durable seams retained; route precedence and DELETE's absence of a version input are now explicit. | Implementation remains outside current scope. | Execute SV-10 through SV-13 only after owner-selected baseline evidence exists. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-29T02:04:57-04:00 | Codex planning session; `main` at `4e42f709d875ead7f977f95c6e3433e556ad8388` | Frontend remains outside the Go module refactor. | Inspected saved-view controller/model and tests plus browser-row selectors; touched only this tracker. | Read-only import/usage searches; `make explain-target TARGET=frontend-import-boundary-check DETAIL=summary`. | Web workbook shell owns controller state; no direct grid-vendor coupling found. | None unless a later slice changes frontend files. | Preserve browser/controller evidence; run the frontend boundary check only if frontend changes. |
| 2026-07-29T03:07:46-04:00 | Codex NLSpec-style revision; `main` at `4e42f709d875ead7f977f95c6e3433e556ad8388` | Frontend and grid-vendor exclusions are normative tracker boundaries. | Rechecked controller/model paths and current browser owner rows; touched only this tracker. | Read-only `rg`/`jq` inspection and Make task-guide discovery. | No frontend interface change is part of the strict portability or backend cohesion plan. | None for tracker revision. | Add browser/frontend checks only when a later diff touches those surfaces. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-29T02:04:57-04:00 | Codex planning session; `main` at `4e42f709d875ead7f977f95c6e3433e556ad8388` | Authored OpenAPI, SQL, portability descriptor, and generated downstream surfaces mapped. | Inspected savedviews OpenAPI owner source/manifest, SQL query/migration, generated SQLC/OpenAPI projections, source port; touched only this tracker. | Read-only searches; `make explain-target` for generation/policy/OpenAPI targets. | Generated outputs remain downstream; portability exact-shape conformance is behavior-sensitive. | RB-001 blocked strict acceptance pending authorization and owner review. | Implement explicit behavior-preserving shape before any separately authorized acceptance change. |
| 2026-07-29T03:07:46-04:00 | Codex NLSpec-style revision; `main` at `4e42f709d875ead7f977f95c6e3433e556ad8388` | Proposed exact row, seven invariants, owner AC, and traceability path are decision-complete. | Inspected Core portability registry, current six-invariant catalog, traceability JSON/schema, migrations, generated policy; touched only this tracker. | Read-only `rg`/`sed`/`jq`; Make explain targets for JSON/generation/artifact policy. | Advisory `id`/`view_scope` mapping rejected; live `saved_view_id`/`scope` preserved; proposed IDs are REQ-01-644..646 and AC-508. | These IDs/content have no product authority until Core adoption. | Execute SV-00 as a separately reviewed owner change. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-29T02:04:57-04:00 | Codex planning session; `main` at `4e42f709d875ead7f977f95c6e3433e556ad8388` | Eleven owner rows mapped; two portability tests are not owner-selected. | Inspected savedviews tests, portability tests, verification owner, test-family/catalog, workbook and Incident Bundle tests; touched only this tracker. | `make task-guide ROLE=module-author OWNER=module.savedviews`; `make explain-test-owner OWNER=module.savedviews`; focused baseline slice. | Baseline passed 1/1 at `.cartulary/test-results/20260729T060105Z-p542483`; test-family gap recorded. | RB-002 omits two portability tests until authored rows are added. | Make owner accounting the first authorized harness slice. |
| 2026-07-29T03:07:46-04:00 | Codex NLSpec-style revision; `main` at `4e42f709d875ead7f977f95c6e3433e556ad8388` | Pre-change evidence and post-adoption conformance evidence are separated. | Inspected exact local tests, savedviews owner rows/profiles, verification schema, Incident Bundle traceability; touched only this tracker. | `jq` owner-row/profile inspection; `make task-guide`; `make help-all`; `make explain-target TARGET=harness-contract DETAIL=summary`. | Existing normalization test MUST first receive a row/baseline, then be replaced by strict rejection evidence after adoption; final row IDs remain generator-owned. | SV-01 requires authorized harness-input changes. | Add two current rows, record run roots, then add the Section 4 invariant matrix. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-29T02:04:57-04:00 | Codex planning session; `main` at `4e42f709d875ead7f977f95c6e3433e556ad8388` | Authentication, membership, scope, owner, version, and harness-token checks mapped as frozen behavior. | Inspected routes, scope, store, route tests, Core 04, harness fixture requirements; touched only this tracker. | Read-only source/test search and owner review. | Policy is split across layers; consolidation requires error-precedence characterization. | No owner contradiction; later implementation authorization required. | Add service-level authorization/error-precedence cases before moving checks. |
| 2026-07-29T03:07:46-04:00 | Codex NLSpec-style revision; `main` at `4e42f709d875ead7f977f95c6e3433e556ad8388` | Dual portable/runtime ownership and safe failure behavior are explicit. | Inspected user/saved-view/imported-actor schemas, attribution mapping, actor export, Core 01/03/04 security rules; touched only this tracker. | Read-only schema/source/owner searches. | Imported source owner is retained in attribution and restored on export; runtime owner defaults to submitter; no authorization state may be synthesized. | Core owner adoption required before changing round-trip behavior. | Add native/imported/system owner and no-membership/login conformance cases. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-29T02:04:57-04:00 | Codex planning session; `main` at `4e42f709d875ead7f977f95c6e3433e556ad8388` | Structural sequence ready; portability behavior correction isolated. | Inspected all evidence then listed in Section 1; touched only this tracker. | Repository discovery, Make target discovery, Markdown lint, final status. | Another agent can begin with owner accounting without redoing boundary discovery. | RB-001 behavior authorization; RB-002 accounting; current task does not authorize implementation. | Add missing rows, characterize, then take small reversible slices. |
| 2026-07-29T03:07:46-04:00 | Codex NLSpec-style revision; `main` at `4e42f709d875ead7f977f95c6e3433e556ad8388` | Both identified planning questions have adopted tracker decisions and explicit execution gates. | Inspected all Section 1 evidence; touched only this tracker. | Read-only repository/Make discovery; final lint and status recorded below. | Another agent can execute SV-00 without choosing wire keys, canonicality, owner mapping, invariant mapping, or test posture. | Product implementation remains blocked by owner adoption and explicit authorization. | Submit SV-00 for Core owner review; do not start strict implementation first. |

## 11. Open Questions and Blockers

`RESOLVED` in this table means the planning choice is closed. It does not mean
the corresponding owner or implementation work has occurred.

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Exact saved-view portability row, strictness, and invariant ownership | A storage-derived shape and permissive normalization can silently drift or discard producer intent. | Proposed REQ-01-644..646, extended REQ-01-640, and AC-508 must be adopted before implementation. | RESOLVED — use exact live eleven-member `saved_view_id`/`scope` shape, strict semantic canonicality, and `saved_views.row_shape_exact`; pending Core adoption and authorized implementation. |
| RB-002 | Focused owner accounting omits both local portability tests and all proposed strict conformance cases. | An owner slice can pass without executing portability evidence. | Authorized authored row changes, generator-derived IDs, harness validation, focused passing runs. | RESOLVED — execute SV-01 before structural portability work and SV-03 after owner adoption. |
| RB-003 | Production and owner-document implementation is outside this tracker-only task. | Any such edit would violate the permitted write scope. | A later task authorizing Core, contract, harness, test, and production changes as applicable. | BLOCKED for implementation; tracker revision may complete. |
| RB-004 | Portable source ownership and runtime ownership use different identities after import. | Without an explicit mapping, round trip either loses source identity or creates unauthorized local capability. | Section 4.5 plus Core adoption and native/imported/system integration evidence. | RESOLVED — runtime owner is the target-local submitter, portable source owner remains in attribution and is restored on re-export. |

No `BLOCKED: owner contradiction` entry is present because no owner
contradiction was found. The advisory/live-repository mismatches are not owner
contradictions and are resolved in Section 1.3.

## 12. Binary Completion Criteria

### 12.1 Tracker revision

| Criterion | Status | Evidence |
| --- | --- | --- |
| Every file in `internal/modules/savedviews` is inventoried. | PASS | Section 2 contains exactly eight target rows. |
| Authority and normative status are unambiguous. | PASS | Sections 1.1 through 1.3 distinguish owners, tracker decisions, repository observations, and advisory evidence. |
| Advisory recommendations are reconciled with the live repository. | PASS | Section 1.3 resolves key names, verification profile, current normalization, and owner representation. |
| Observable ordinary behavior is complete and testable. | PASS | Sections 4.1, 4.2, and 4.8 define routes, defaults, precedence, seams, errors, and negative effects. |
| The proposed portability row has exact members, types, defaults, bounds, and exclusions. | PASS | Sections 4.3 and 4.4 define all eleven members and every admitted omitted/zero/error condition. |
| Portable/runtime ownership is deterministic. | PASS | Section 4.5 defines native, imported, system, and duplicate mappings. |
| Source-port inputs, outputs, side effects, and failure behavior are unambiguous. | PASS | Sections 4.6 and 4.7 define all four operations and seven invariant mappings. |
| Every workstream and slice has dependencies, validation, rollback, and a binary exit. | PASS | Sections 6 and 7. |
| Behavior-changing slices are gated by later authorization and owner adoption. | PASS | SV-00 and SV-03 through SV-06 are explicitly gated; Section 9 marks implementation blocked. |
| Validation commands are Make-owned and no generated row ID is invented. | PASS | Section 8 and SV-01/SV-03 use generator-derived IDs with an explicit pre-generation `TODO` reason. |
| Prior handoff history is preserved and a new row exists in every required table. | PASS | Section 10 retains each original row and appends seven current-session rows. |
| Contradictions use the required blocker posture. | PASS | No owner contradiction exists; advisory mismatches are explicitly non-authoritative and reconciled. |
| Only the tracker changed and its final content passed Markdown lint. | PASS | Worktree inspection identified only this tracker; `make lint-markdown` passed at `.cartulary/test-results/20260729T071524Z-p569244`. |

### 12.2 Future implementation definition of done

The refactor is complete only when every statement below is true:

1. REQ-01-644 through REQ-01-646, the extended seven-invariant REQ-01-640 row,
   AC-508, and the REQ-02-152 clarification are adopted.
2. The exact eleven-member contract has a downstream closed machine-readable
   schema with no additional properties and generator provenance.
3. Both original portability tests have retained pre-change owner-selected
   evidence, and the obsolete normalization expectation is replaced after
   adoption rather than silently rewritten.
4. Exact shape, alias, unknown/duplicate member, strict canonicality, scope,
   ownership, schema, Reference Pack, determinism, zero-row, safe-failure, and
   transactionality cases are active owner-selected evidence.
5. Export uses explicit fields, emits deterministic bytes, includes all scopes,
   and reconstructs preserved source owners.
6. Prepare is bounded, strict, duplicate-aware, typed, and side-effect-free.
7. Apply uses fixed explicit persistence, writes only admitted state, requires
   affected-row equality, and performs the Section 4.5 owner mapping exactly.
8. Validate proves all seven invariants inside the final transaction.
9. Ordinary route, startup, authorization, version, no-op, deletion, OpenAPI,
   frontend, and negative side-effect behavior remains unchanged.
10. The root exposes only the three justified production seams and any type
    whose live caller is documented.
11. Traceability has no orphan requirement, AC, invariant, verification, or
    active row, and generated outputs equal authored inputs.
12. Focused, service-backed, applicable browser, harness, drift, boundary,
    `make agent-finalize`, and `make check` results are recorded for the exact
    final commit.

No production refactor is performed or authorized by this tracker revision.
