# evidence Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Item | Current posture |
| --- | --- |
| Target path | `internal/modules/evidence` |
| Target label | `evidence` |
| Output path | `docs/handoffs/evidence-module-refactor-tracker.md` |
| Baseline | Commit `367bc76c124a`; worktree clean before this tracker was created; target contains 40 tracked entries: 28 production Go files, 11 Go test files, and one non-behavioral `.gitkeep`. |
| Status | Planning and documentation only. |
| Allowed changes | This tracker file only. |
| Non-goals | No production code, tests, contracts, generated artifacts, package configuration, migrations, harness inputs, or runtime behavior are changed. |
| Later authority | Any production refactor, characterization-test addition, contract correction, code generation, harness edit, or migration requires a later authorized task. |

The target label was derived from `internal/modules/evidence` and normalized to
the safe lowercase kebab-case value `evidence`. The target exists. This tracker
does not assume that the directory name is the correct permanent implementation
boundary.

The authority order used here is:

1. adopted subsystem NLSpecs, within their named subsystem;
2. Core 00 through Core 04 for implementation conformance;
3. Core 05 only for claim-bearing timed or fixture-sensitive publication;
4. domain vocabulary and implementation-support guidance;
5. current repository code and tests;
6. prior plans, handoffs, and the planning framework as evidence only.

No Core 00 through Core 04 owner contradiction was found. Core 05 is not
applicable because this plan makes no timed or fixture-sensitive publication
claim. Core 01 currently requires the normal Evidence create path to accept an
uploaded `object_blob_id`, but its closed generic row-create namespace and
view-schema discovery contract declare no corresponding input. RB-001 therefore
records both an incomplete owner interface and a current
implementation-conformance gap. It is not an owner contradiction.

In this tracker, `MUST`, `MUST NOT`, `SHOULD`, and `MAY` state requirements for
the proposed adoption package or for later authorized refactor execution. They
do not amend an adopted owner. Proposed behavior becomes implementation
authority only after the named Core owner edits are adopted. Until then,
structural slices MUST preserve current observable behavior and MUST keep the
known non-conformance visible.

### Owner documents inspected

- `docs/spec/00_document_set_status_and_precedence.md`, especially status,
  precedence, conformance, and the contract-owner matrix.
- `docs/spec/01_architecture_storage_and_view_contracts.md`, especially the
  Evidence routes, Evidence view schema, projection model, backup/recovery,
  incident portability, and access-handle contract.
- `docs/spec/02_domain_model_schema_and_history.md`, especially Evidence and
  object metadata and both lifecycle machines.
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`, especially
  collaboration, two-step attachment, lifecycle bridge, access behavior, and
  Evidence workbook surfaces.
- `docs/spec/04_security_deployment_and_conformance.md`, especially current
  authorization, Evidence uploads, hostile-content controls, lifecycle
  conformance, and public-interface criteria.
- `docs/reporting-subsystem-nlspec.md`, within its adopted Reporting-only source
  and immutable export-model boundary.
- `docs/testing-harness-nlspec.md`, within its adopted harness catalog,
  selection, and evidence-accounting boundary.
- `docs/domain.md`, especially the Evidence bounded context, Evidence versus
  object-blob distinction, context map, workflow vocabulary, and agent rules.
- `docs/research/nlspec-spec.md` for specification discipline; it is grounding
  material, not Cartulary behavior authority.

### Planning and repository evidence inspected

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md` was read
  first and used as planning doctrine, not repository-state proof.
- `temp/analysis-notes.md` was inspected as planning evidence. Its recommended
  resolutions are restated here with explicit adoption gates; it is not product
  authority.
- Every file listed in Section 2 was opened directly.
- Application composition and inbound use were inspected in Server, Timeline,
  Workbook, Projection, Import, Incident Portability, Recovery, Reporting, and
  Revision assembly or module code.
- Authored and generated surfaces were inspected under
  `contracts/openapi-source`, `contracts/view-schemas`,
  `contracts/projection-providers`, `contracts/imports`,
  `contracts/verification`, `internal/gen/openapioperations`,
  `packages/protocol-ts/src/generated`, and
  `packages/ui-contracts/src/generated`.
- Harness routing was inspected in `tools/test_families/module.evidence.json`,
  `tools/test_catalog_owner.json`, `tools/backend_module_boundaries.json`, and
  the generated execution topology.

## 2. Current-State Repository Inventory

Every current file under `internal/modules/evidence` is inventoried below.
Tests are in scope as current behavior and verification evidence; they are not
production package ownership.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/evidence/.gitkeep` | Retains the directory in Git; contains no production or test behavior | None | None | None | None | None | Out of scope: repository placeholder | none | Explicitly inventoried and excluded from refactor slices because it is non-behavioral. |
| `internal/modules/evidence/api.go` | Strict JSON decoding and request hashing for blob create, attach, and handle issue | `BlobCreateRequest`, `AttachBlobRequest`, decode and hash functions | Evidence route registrar | `platform/httpapi` | Blob-create, attach, handle, and integration tests | Evidence OpenAPI request schemas and error contracts | Evidence transport adapter | high | Transport-adjacent code in the module root. |
| `internal/modules/evidence/attach_test.go` | Attach validation, lifecycle, replay, and durable-side-effect tests plus shared attach fixtures | Test functions and test-only exported fixture helpers | Go test runner; Evidence catalog rows | Evidence store, auth test support, Postgres, app test support | This file | Error, lifecycle, revision, and row-refresh contracts | Evidence test evidence | high | Test-only helpers must not become production APIs. |
| `internal/modules/evidence/blob_create_test.go` | Blob-create request, authorization, idempotency, and size-limit tests | Test functions | Go test runner; Evidence catalog rows | Evidence routes/store, auth and incident fixtures | This file | Blob-create OpenAPI and limit/error contracts | Evidence test evidence | high | Covers route and store behavior. |
| `internal/modules/evidence/blobref/blobref.go` | Canonical physical object key and logical `object://` reference parsing/formatting | `StorageKeyParts` and key/ref functions | Evidence routes, store, workbook decoder, portability, tests | UUID and standard library only | `blobref_test.go`, route, integration, and object-store tests | Core object-key and storage-reference contracts | Evidence anti-corruption helper | medium | Legitimate Evidence-owned boundary; not an object-store adapter. |
| `internal/modules/evidence/blobref/blobref_test.go` | Key/ref grammar and round-trip characterization | Test functions | Go test runner | `blobref` | This file | Object-key and logical-ref contracts | Evidence test evidence | low | Keep with the owner helper. |
| `internal/modules/evidence/collaboration_intents.go` | Builds canonical record-change intents for primary and affected Evidence changes | Private package surface | Store and workbook facade | Collaboration intent API | Attach and integration tests | WebSocket `record_changed` contract | Evidence commit adapter | high | Durable intent creation is a post-mutation effect, not transport publication. |
| `internal/modules/evidence/conformance_test.go` | Checks reason registries and Evidence/Timeline derived field-key closure | Test functions and test-only helpers | Go test runner; Evidence unit rows | Contract test and view-schema loaders | This file | Error registry and view-schema contracts | Evidence conformance evidence | medium | Downstream verification, not behavior authority. |
| `internal/modules/evidence/create_validation.go` | Evidence workbook create-signal and direct lifecycle patch validation | `ValidateWorkbookCreateParams`, `ValidateWorkbookDirectPatchChange` | Workbook facade and import create path | Evidence workbook value types | Workbook mutation and import integration coverage | Evidence view-schema create contract | Evidence domain/application validation | high | Current create-signal implementation differs from Core 01; see RB-001. |
| `internal/modules/evidence/deleterestore/provider.go` | Evidence source snapshot and delete/restore precondition contribution | `Source`, `NewSource`, source-contract methods | Evidence revision contribution; Revisions coordinator | Revisions delete/restore contract, SQL transaction | Revision assembly and delete/restore suites | Revision provider contract | Evidence source-owner contribution | medium | Correctly Evidence-owned and consumer-facing through a narrow contract. |
| `internal/modules/evidence/evidence_integration_test.go` | End-to-end Evidence routes, handle states, authorization, and OpenAPI presence | Test functions and integration helpers | Go test runner; Evidence integration rows | Server harness, Evidence, auth/http test support | This file | OpenAPI routes, access errors, lifecycle and handle contracts | Evidence integration evidence | high | Direct filesystem read is test evidence only. |
| `internal/modules/evidence/handles_test.go` | Handle issuance, current-state recheck, filename, disposition, and consumption tests | Test functions and test fixtures | Go test runner; Evidence store rows | Evidence store/routes and test support | This file | Evidence handle response and header contracts | Evidence test evidence | high | Freezes redeem-time state binding. |
| `internal/modules/evidence/import_create.go` | Evidence-owned import create facade and transaction-bound row creation | `ImportCreateCommand`, `NewImportCreateFacade`, `Store.CreateImportRowTx` | Import assembly/dispatcher | Imports owner facade, Records, Revisions, Collaboration, Postgres | Imports integration and owner-registry tests | Import target `cartulary.view.evidence.v1` / `evidence.import_create` | Evidence source-owner contribution | high | Imports owns orchestration; Evidence owns durable row semantics. |
| `internal/modules/evidence/import_projection.go` | Refreshes and returns an imported Evidence row inside the caller transaction | `Store.RefreshImportRowTx` | Evidence import facade | Evidence projection provider and Projections row reader | Imports integration tests | Import owner response and Evidence view-row contract | Evidence source-owner contribution | medium | Keep transaction ownership and row shape stable. |
| `internal/modules/evidence/incident_bundle_blob_portability.go` | Exports blob bytes, validates/re-writes object metadata, stages imports, and cleans staging | `IncidentBundleBlobPortability` methods | Incident Bundle coordinator/application assembly | Evidence blob refs, Incident Portability types, object store | Incident Bundle suites and Evidence portability catalog row | Incident Bundle manifest/member contracts | Evidence portability adapter | high | Evidence owns its bytes and metadata; coordinator owns bundle flow. |
| `internal/modules/evidence/incident_bundle_portability.go` | Exports and imports Evidence database member files | Export/import functions | Evidence source port and Incident Bundle coordinator | Incident Portability query/file contracts | Incident Bundle integration tests | `data/evidence_records.ndjson`, custody, and object-blob members | Evidence source-owner contribution | high | Must preserve FK order, lifecycle, attribution, and hidden staging. |
| `internal/modules/evidence/incident_bundle_source_port.go` | Declares Evidence Incident Bundle source descriptor and validation | `NewIncidentBundleSourcePort` | Incident portability assembly | Incident Bundles source-port contract | Incident Bundle catalog and import tests | Source catalog and traceability contracts | Evidence source-owner contribution | medium | Source owner constructs the provider by repository policy. |
| `internal/modules/evidence/incident_bundle_subtype_presence.go` | Enumerates Evidence subtype presence for aggregate bundle validation | `IncidentBundleSubtypeContribution` and source methods | Records subtype-presence catalog | Records subtype-presence contract, Postgres | Incident Bundle and Records tests | Incident Bundle subtype accounting | Evidence source-owner contribution | medium | Fixed-query source-owner adapter. |
| `internal/modules/evidence/integration_test.go` | Upload capability, attach, projection rebuild, collaboration refresh, expiry, quarantine, and cleanup integration coverage | Test functions and test helpers | Go test runner; Evidence integration catalog rows | Full server/test composition, Evidence, Collaboration, object store, Workbook support | This file | HTTP, WebSocket, projection, lifecycle, and cleanup contracts | Evidence integration evidence | high | Primary cross-owner backend characterization. |
| `internal/modules/evidence/lifecycle_test.go` | Evidence/blob lifecycle separation and attributed history tests | Test functions and test-only query interface | Go test runner; Evidence store catalog row | Evidence, Workbook, auth and revision test support | This file | Core 02 lifecycle and revision contracts | Evidence lifecycle evidence | high | Protects the two-machine distinction. |
| `internal/modules/evidence/objectstore_dependency_test.go` | Public object-store dependency error mapping and secret/backend-detail redaction | Test functions and faulting object-store fixture | Go test runner; Evidence route integration rows | Object-store contract, route test support | This file | Public error and reason registries | Evidence security evidence | high | Freezes fail-closed redaction and stable codes. |
| `internal/modules/evidence/projectionprovider/provider.go` | Evidence projection row refresh and incident rebuild SQL | Refresh and rebuild functions | Projection catalog and Evidence store adapter | Postgres transaction | Evidence integration and Projections tests | Projection-provider descriptor and Evidence grid table | Evidence source projection provider | high | Source owner supplies derivation; Projections owns storage/query coordination. |
| `internal/modules/evidence/projectionprovider/query_surfaces.go` | Declares Evidence query fields and SQL expressions | `QuerySurfaces` | Projection assembly/catalog | Projections provider contract | Projections query-surface tests | Evidence view schema and provider descriptor | Evidence source projection provider | high | Public query semantics must remain aligned with the view schema. |
| `internal/modules/evidence/recovery_inventory.go` | Enumerates available immutable Evidence objects for backup | `VNextRecoveryObjectInventory` | Recovery object-family composition | Recovery inventory contract, Postgres query | Recovery provider and operator suites | Recovery catalog/object manifest | Evidence recovery contribution | high | Owner inventory must agree with authoritative blob state. |
| `internal/modules/evidence/recovery_state.go` | Declares authoritative, rebuildable, and ephemeral Evidence recovery state | `RecoveryStateContribution` | Recovery state catalog | Recovery state contract | Recovery assembly catalog tests | Recovery-state catalog | Evidence recovery contribution | medium | Handles are security-ephemeral; projections are rebuildable. |
| `internal/modules/evidence/recoveryprovider/provider.go` | Reads available Evidence object state and row counts for recovery | `Provider`, `New`, list/count methods | Operator/recovery composition and recovery browser tool | Recovery contract and Postgres | Recovery integration suites | Recovery object and verification contracts | Evidence recovery adapter | high | Read-only recovery view over Evidence-owned state. |
| `internal/modules/evidence/reportingprovider/provider.go` | Materializes redacted Evidence fields/facts for immutable reporting export models | `CollectFieldsTx`, `CollectFactsTx` | Reporting export materializer | Reporting export-provider contract | Reporting boundary and materialization tests | Reporting derivation/export-model contract | Evidence reporting contribution | high | Excludes blob hash, storage ref, and object-blob ID from emitted facts. |
| `internal/modules/evidence/revision_append_port.go` | Private adapter over generic Revisions append operations | Private interface and adapter methods | Store, workbook, and import mutation paths | Revisions appender | Mutation and lifecycle tests | Change-set, mutation, and record-revision contracts | Evidence mutation coordination | high | Makes revision effects replaceable but remains coupled to Revisions DTOs. |
| `internal/modules/evidence/revision_provider_contribution.go` | Registers Evidence delete/restore, rollback, live-change, and view routing | `RevisionProviderContribution` | Revision assembly | Evidence subproviders and Revisions contracts | Revision assembly and conflict/history tests | Revision provider catalog | Evidence source-owner contribution | medium | Legitimate source-owner contribution. |
| `internal/modules/evidence/rollbackprovider/provider.go` | Validates and restores Evidence row values and associations during rollback | `Provider`, `NewProvider`, validation/restore methods | Evidence revision contribution; Revisions rollback coordinator | Revisions rollback contract, SQL transaction | `provider_test.go`, Revisions tests | Rollback source contract | Evidence source-owner contribution | high | Must preserve lifecycle and reference guards. |
| `internal/modules/evidence/rollbackprovider/provider_test.go` | Rollback value-presence, association, lifecycle, and reference validation | Test functions | Go test runner | Evidence rollback provider and contract | This file | Rollback contract | Evidence test evidence | medium | Selected through Revisions/Evidence ownership as applicable. |
| `internal/modules/evidence/routes.go` | Registers six Evidence HTTP operations and coordinates auth, upload, attach, handles, object-store access, and public errors | `Service`, `Settings`, `RouteOption`, `WithStore`, `RegisterRoutes` | Server runtime | Incidents, Evidence blob refs/store, auth/http/object-store platform packages | Route, handle, object-store, process, and browser tests | Evidence OpenAPI operations and generated protocol clients | Evidence transport/application adapter | high | Broad transport-adjacent file; preserve exact public routes and envelopes. |
| `internal/modules/evidence/store.go` | Blob slots, attachment, lifecycle bridge, quarantine, cleanup, access records, handles, idempotency, revisions, projections, and intents | `Store`, options, request/result records, lifecycle/access methods, stable errors | Routes, Workbook, Timeline adapter, Import, tests | Collaboration, Incidents, Projections, Records, Revisions, auth, object store, Postgres | Most Evidence backend tests plus Timeline tests | Lifecycle, projection, revision, WS, and access contracts | Evidence application facade over private persistence | critical | Legitimate Evidence behavior but currently a monolithic mutation/persistence coordinator. |
| `internal/modules/evidence/store_test.go` | Evidence projection-row mapping test | Test function | Go test runner | Evidence row mapping | This file | Full `view_row_v1` Evidence cell contract | Evidence unit evidence | medium | Narrow mapping coverage only. |
| `internal/modules/evidence/timeline_facts.go` | Loads Evidence facts used to derive Timeline attached-evidence cells | `TimelineFactReader`, `LoadTx` | Timeline application collection-read adapter | Timeline `workbookprojection.EvidenceFact`, Postgres | Timeline attached-evidence and Evidence integration tests | Timeline collection/projection contract | Evidence read port with application mapping | high | Direct import of a Timeline implementation DTO is a proven reverse coupling. |
| `internal/modules/evidence/timeline_port.go` | Validates same-incident active Evidence IDs in a caller transaction | `Store.ValidateTimelineAttachmentsTx` | Timeline Evidence collaborator adapter | Postgres and Evidence source tables | Timeline collection and integration tests | Timeline attached-evidence mutation contract | Evidence transaction port | high | Legitimate source-owner validation; keep caller transaction. |
| `internal/modules/evidence/upload_token.go` | Signs and verifies opaque same-origin upload capabilities | Private token claims and helpers | Evidence routes | Auth signing key and standard crypto/JSON | Blob-create and integration tests | Upload target capability contract | Evidence transport security adapter | high | Private implementation; token remains opaque to clients. |
| `internal/modules/evidence/workbook_conflict.go` | Resolves Evidence workbook conflicts through Revisions mechanics and returns authoritative rows | `WorkbookConflictCommand`, `WorkbookFacade.ResolveConflict` | Workbook mutation contribution | Revisions conflict resolution and view-schema types | Workbook conflict-route tests | Generic conflict route and Evidence row contract | Evidence workbook facade | high | Evidence owns row application; Revisions owns generic conflict mechanics. |
| `internal/modules/evidence/workbook_facade.go` | Evidence workbook create/patch, idempotency, concurrency, history, projection, and collaboration choreography | `WorkbookFacade`, request/command/result/error types, `NewWorkbookFacade`, create/patch methods | Workbook assembly, import construction, tests | Collaboration, Incidents, Evidence projection provider, Projections, Records, Revisions, auth, Postgres | Workbook mutation, row-wire, party, conflict, and Evidence tests | Generic workbook mutation routes and Evidence view schema | Evidence workbook application facade | critical | Public facade is legitimate; internal orchestration has broad cross-owner coupling. |
| `internal/modules/evidence/workbook_mutations.go` | Evidence row insert, lifecycle validation, scalar patch application, and touch behavior | Value/validation types and `Store` mutation methods | Workbook facade, import facade, rollback-related flows | Postgres and Evidence lifecycle rules | Workbook and lifecycle tests | Evidence writable-field and lifecycle contracts | Evidence domain persistence | high | Keep Evidence-owned semantics behind a narrower repository/service boundary. |

## 3. Module Boundary Diagnosis

The directory is a legitimate Evidence bounded context and a mixed-responsibility
implementation package. It is simultaneously a persistence-adjacent adapter,
mutation coordinator, transport-adjacent adapter, workbook facade, and
view/projection source contribution. It is not a frontend shell, grid-vendor
integration, or accidental catch-all whose entire contents should move.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Evidence records, blob metadata, upload slots, attachment, custody, lifecycle, quarantine, and cleanup | Root Store and workbook mutation files | Evidence | keep and split internally | Core 01–03; domain Evidence context; source SQL | Core Evidence behavior, currently combined in one store. |
| Six public Evidence route operations | `api.go`, `routes.go`, `upload_token.go` | Evidence route adapter over platform HTTP/auth/object store | split | OpenAPI owner, route registration, route tests | Keep route identities; separate decoding/admission/adapters from service behavior. |
| Workbook create, patch, and conflict behavior | Workbook facade, validation, mutation, and conflict files | Evidence source owner using Workbook/Revisions contracts | keep and split | Workbook assembly and generic route tests | Do not move Evidence validation or persistence into Workbook. |
| Collaboration, revision, and projection effects | Store, workbook facade, collaboration/revision adapters | Evidence mutation coordinator using narrow owner ports | split | Transaction paths and integration tests | Preserve atomic ordering, replay, and durable intent semantics. |
| Evidence grid projection derivation/query declaration | `projectionprovider` and import projection | Evidence provider; Projections storage/query coordinator | keep | Core 01 §8, provider catalog, query tests | Source-owner contribution is intentional. |
| Timeline attachment validation | `timeline_port.go` | Evidence transaction port | keep | Timeline collaborator composition and tests | Timeline owns the action; Evidence owns target validation. |
| Timeline display facts | `timeline_facts.go` | Evidence-owned facts mapped by application composition | move mapping / split DTO | Direct import of Timeline workbook-projection DTO | Remove the reverse implementation dependency without changing facts. |
| Import create | `import_create.go`, `import_projection.go` | Evidence source-owner facade | keep | Import registry and integration tests | Imports retains orchestration and unit state. |
| Incident portability | Four Incident Bundle contribution files | Evidence source-owner portability adapters | keep and isolate | Core portability catalog/order and assembly | Do not move Evidence byte or row invariants into the coordinator. |
| Recovery | Recovery state, inventory, and provider files | Evidence source-owner recovery contribution | keep and isolate | Recovery catalog and operator composition | Recovery owns workflow; Evidence owns enumeration and invariants. |
| Reporting | Reporting provider | Evidence source-owner contribution under Reporting NLSpec | keep and isolate | Export materializer and redaction SQL | Extension-scoped; must not redefine live Evidence access. |
| Delete, restore, rollback, and revision registration | Provider subpackages and contribution | Evidence source owner with Revisions coordination | keep | Core 00 owner matrix and revision assembly | Existing subpackage boundaries are appropriate. |
| Evidence runtime construction under `timelineassembly.Bundle.EvidenceStore` | Timeline application assembly | `internal/app/server` composition using Evidence-owned constructors | move in SL-06 | The developer guide assigns configuration-driven Server assembly to `internal/app/server`; live Server retrieves the Evidence store from Timeline | Decision closed at planning level: Server owns composition, not Evidence behavior. No replacement broad bundle or store accessor is permitted. |
| Frontend shell/controller, saved-view implementation, and grid adapter | No production file in target | Workbook/Web/Saved Views/Grid Adapter owners | defer / no move | Target import scan and contract/test catalog | Indirect Evidence contracts remain frozen. |

### Application-composition boundary

`internal/app/server` MUST own Evidence application composition. Evidence MUST
continue to own its constructors, source behavior, private persistence,
route-facing service, Workbook facade, Timeline capabilities, and source-owner
contributions. Server MUST construct the Evidence runtime exactly once, map
owner facts at application seams, inject only narrow capabilities, validate
required contributions before use, and register routes only after composition
is complete.

The replacement MUST NOT expose `server.Bundle.EvidenceStore`,
`application.Bundle.EvidenceStore`, another public aggregate store accessor, a
service locator, a mutable global registry, or package-initialization
registration. A private Server wiring helper MAY organize construction, but it
MUST remain subordinate to the Server composition root.

| Consumer | Capability Server MUST inject | Consumer MUST NOT receive |
| --- | --- | --- |
| Evidence routes | Narrow route/application service | `Store`, SQL handles, or provider catalogs |
| Workbook | Evidence `WorkbookFacade` or equivalent owner port | Evidence repository or object-store adapter |
| Timeline | Attachment validator plus application-mapped fact source | Evidence `Store` or Timeline DTO returned by Evidence |
| Projections | Evidence projection contribution | Callback registry owned by Evidence |
| Imports | Evidence import owner facade | Evidence repository or import parser |
| Incident Portability | Evidence row, blob, source, and subtype contributions | Evidence runtime or private persistence |
| Recovery | Evidence state, inventory, and read-only provider contributions | Route or Workbook facade |
| Reporting | Deterministic redacted Evidence contribution | Live-access service or raw/private blob metadata |
| Revisions | Explicit delete/restore, rollback, live-change, and routing contributions | Evidence construction authority |
| Server tests | Application-support composition or public routes | Production `Runtime.EvidenceStore` accessor |

Timeline MUST own its consumer interfaces and Timeline-facing DTO. Evidence MUST
return an Evidence-owned fact DTO. An application adapter MUST map between the
two without loss of identifiers, order, labels, lifecycle state, upload state,
or attached counts. The initiating Timeline transaction MUST remain the caller
transaction for validation and fact loading; neither module may import the
other's implementation package.

| Construction order | Required result |
| --- | --- |
| 1 | Validate configuration and classify supplied Postgres and object-store dependencies as borrowed or runtime-owned. |
| 2 | Construct Records, Revisions, Collaboration, Incidents, authorization, and platform object-store adapters required by Evidence. |
| 3 | Construct the Evidence projection contribution. |
| 4 | Validate the complete projection catalog before any dependent Evidence facade is usable. |
| 5 | Construct one Evidence core facade from the validated narrow ports. |
| 6 | Construct Evidence route, Workbook, Timeline, import, portability, recovery, reporting, and revision capabilities from that owner runtime. |
| 7 | Register and validate those contributions with their consumer owners. |
| 8 | Construct Workbook and Timeline application adapters from the registered capabilities. |
| 9 | Register HTTP and WebSocket routes. |
| 10 | Acquire listeners and publish readiness only after every required component is complete. |

Failure at any construction step MUST leave readiness false and MUST acquire no
later listener. Shutdown MUST close only runtime-owned resources in reverse
acquisition order, MUST leave borrowed Postgres and object-store dependencies
open, and MUST remain idempotent.

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `POST /api/v1/object-blobs` | Core 01 Evidence routes; Evidence implementation | OpenAPI owner, `api.go`, `routes.go`, Store | Blob-create unit/integration and browser tests | Preserve exact body, idempotency, limits, accepted contract, upload target, auth, and error shapes | critical | No raw object-store identity may escape. |
| `PUT /api/v1/object-uploads/{upload_token}` | Core 01/Core 03; Evidence route adapter | OpenAPI, upload token, object-store calls | Upload/attach integration and browser tests | Preserve expiry, one-slot scope, body/size behavior, dependency errors, and opacity | critical | Capability is not an Evidence identifier. |
| `POST /api/v1/evidence-records/{record_id}/attach-blob` | Core 01/Core 03; Evidence | OpenAPI, routes, Store | Attach validation, replay, projection, WS, browser tests | Preserve version check, idempotency, visible-blob concealment, lifecycle bridge, row refresh, and side effects | critical | Same request replay must not duplicate durable effects. |
| Preview/download handle issuance routes | Core 01 §16; Evidence | OpenAPI, routes, Store | Handle unit/integration, frontend, browser, accessibility tests | Preserve `{}`-only issue body, fresh-handle behavior, exact expiry/kind/disposition, preview policy, and errors | critical | Issuance is deliberately non-idempotent. |
| `GET /api/v1/evidence-handles/{handle_token}` | Core 01 §16/Core 04; Evidence | Route and handle store | Redemption, membership/session loss, missing blob, range/browser tests | Preserve session binding, current-state recheck, single-use consumption point, filename sanitization, and concealment | critical | Authentication failure precedes handle lookup. |
| Generic Evidence workbook create | Core 01 §7.4.4; Workbook route with Evidence facade | View schema, Workbook assembly, facade and validator | Workbook mutation, import, browser tests | Add the complete Section 4 input, signal, lifecycle, atomicity, replay, concurrency, error, and browser-flow matrices | critical | RB-001 is both an incomplete owner interface and an implementation gap; correction requires owner adoption and later SL-07 authorization. |
| Generic Evidence workbook patch and conflict resolution | Core 01/Core 02; Evidence facade with Workbook/Revisions | View schema and generic routes | Row-wire, party, lifecycle, conflict, replay tests | Preserve null/omission, direct refs, lifecycle guards, conflict classes, no-op/replay, history, and sparse row changes | critical | Projection rows are never source authority. |
| Evidence workbook query, sort, filter, grouping, and cursors | Core 01 §3.3.4/§8; Projections using Evidence provider | View schema and query surfaces | Projections and Workbook query tests | Preserve all declared fields, deterministic sort/cursor binding, full row cells, and no blob hot-path reads | high | Query orchestration is not owned by Evidence. |
| Evidence and blob lifecycle machines | Core 02 §13; Evidence | Store, workbook mutations, cleanup | Lifecycle, attach, quarantine, cleanup, process tests | Preserve every legal/illegal transition, bridge guard, retry/timeout, cleanup, and inconsistent-state outcome | critical | Machines remain separate. |
| Object-store key/ref and dependency mapping | Core 01/Core 04; Evidence plus platform object store | `blobref`, routes, Store | Blobref and dependency tests | Preserve grammar, identity validation before calls, stable public codes, retryability, and secret/backend redaction | critical | External locators and server-managed refs remain distinct. |
| Projection refresh/rebuild | Core 01 §8; Projections with Evidence provider | Provider catalog and Store ports | Evidence projection and Projections tests | Preserve row-version parity, derived fields, linked counts, affected Timeline/Host/Identity refresh, and rebuild determinism | critical | Physical phase/test maps do not define runtime ownership. |
| `record_changed` WebSocket effects | Core 01 collaboration payload; Collaboration durable stream | Intent builder and commit paths | Attach WS and browser multi-session tests | Preserve changed field keys, affected views, row versions, replay, authorization filtering, and exactly-once durable intent | critical | Routes must not become the durable publication owner. |
| Revisions, delete/restore, history, conflict, and rollback | Core 00 matrix/Core 01–02; Revisions plus Evidence source contributions | Revision contribution and subproviders | Revision, conflict, lifecycle, rollback tests | Preserve current-envelope ownership, change sets, mutation ordering, snapshots, live-change policy, and row restoration guards | critical | Revisions coordinates history; Evidence retains current source authority. |
| Import target `cartulary.view.evidence.v1` / `evidence.import_create` | Core Import profile; Imports plus Evidence owner facade | Import target contract and registry | Import registry and integration tests | Preserve normalized field plan, caller transaction/change set, validation, row refresh, replay, and rollback | high | No parser or worksheet DTO belongs in Evidence. |
| Incident Bundle Evidence members and blob objects | Core Incident Portability profile; Incident Bundles plus Evidence source port | Source catalog, compatibility, portability code | Incident Bundle and recovery tests | Preserve exact members, staging invisibility, digest/key rewrite, FK order, validation, attribution, cleanup, and final transaction | critical | Generated or authored bundle contracts are not hand-edited during code movement. |
| Backup/recovery Evidence state | Core 01/Core 04; Recovery plus Evidence contributions | Recovery catalog and object inventory | Recovery catalog, operator, restore tests | Preserve authoritative/rebuildable/ephemeral classification, complete blob inventory, digest metadata, and restore invariant checks | critical | Handles remain invalidated after restore. |
| Reporting Evidence facts | Reporting NLSpec within Snapshot/Reporting profile | Reporting provider and materializer | Reporting boundary/materialization tests | Preserve immutable snapshot boundary, support refs, deterministic facts, and exclusion of storage/blob secrets | high | No effect on live Evidence authorization. |
| Evidence view schema, inspector, and saved-view use | Core 01/Core 03/Core 04; View Schema, Workbook, Saved Views | Authored schema and generated UI/protocol projections | Frontend unit, browser, visual, accessibility, startup/saved-view tests | Preserve fields, visibility, inspector route bindings, disabled states, selector identities, and saved-view scope | critical | No frontend implementation is inside the target. |
| Authorization and security | Core 04; route/application admission | Routes, auth calls, hostile-content and object-store behavior | Role, membership-loss, CSRF, handle, browser tests | Preserve current-membership checks, editor/reviewer/admin gates, hidden-resource behavior, CSRF, hostile-content preview limits, and no admin bypass | critical | UI visibility is not authorization. |
| Generated OpenAPI/protocol/view contracts | Applicable Core owner through authored contract inputs | OpenAPI manifest, view schema, generated catalogs/packages | OpenAPI, protocol, view-contract, frontend tests | Drift checks after any authored contract change; never hand-edit generated roots | high | Structural slices should require no wire regeneration. |
| Harness/test accounting | Testing Harness NLSpec and verification owners | Evidence verification contract and 35-row family | Catalog validation and owner diagnostics | Update only when tests move/add/retire; preserve one active owner row per executable identity | high | Routing proves execution accounting, not product behavior. |

### Proposed owner-adoption package for Evidence create

The following package closes the interface ambiguity identified by RB-001. It
is a proposed owner amendment, not current conformance. SL-07 MUST NOT implement
or claim this behavior until the applicable owner changes are adopted and the
behavior slice receives separate authorization.

| Owner location | Required amendment |
| --- | --- |
| Core 01 REQ-01-057 | Admit `client_txn_id`, allowed create-writable field keys, and exact create-input keys declared by the addressed view. Keep the top-level namespace closed. |
| Core 01 REQ-01-061 | Include normalized create inputs in replay comparison. Compare `evidence.initial_object_blob_id` as one exact opaque identifier. |
| Core 01 REQ-01-069 | Validate create inputs separately from fields and reject null, malformed, duplicate, unknown, or foreign-view inputs. |
| Core 01 REQ-01-070 | Preserve the generic create success and replay envelope. Do not add a bespoke Evidence response or duplicate the blob ID outside `data.row`. |
| Core 01 REQ-01-245 | Name `evidence.initial_object_blob_id` as the declared normal-row-create input for new-record blob finalization. |
| Core 01 REQ-01-288 | Add required `create_inputs[]` to `view_schema_resource_v2`; the explicit default for a view with no inputs is `[]`. |
| Core 01 REQ-01-328 | Declare the Evidence input and the signal, default, lifecycle, atomicity, replay, and failure semantics below. |
| Core 01 error registry | Add `minimum_create_signal_missing` and bind already-associated or foreign blob concealment to `evidence_attach_rejected/blob_not_visible`. |
| Core 02 §13 | Enumerate legal initial Evidence lifecycle and finalized-blob combinations; preserve the separate Evidence and blob state machines. |
| Core 03 §8 | Define the client-local provisional draft and the exact same-visible-flow sequence. |
| Core 04 acceptance criteria | Add binary interface, authorization, concealment, lifecycle, atomicity, replay, concurrency, hostile-content, and no-leakage cases. |

Every current-profile view MUST declare `create_inputs`. Every view other than
`cartulary.view.evidence.v1` MUST declare `create_inputs=[]`. Evidence MUST
declare exactly:

```json
{
  "input_key": "evidence.initial_object_blob_id",
  "value_contract_id": "object_blob_id_v1",
  "required": false,
  "nullable": false
}
```

`object_blob_id_v1` MUST be one non-null opaque public identifier. A client MUST
NOT parse or synthesize it. It is not a bucket, physical object key, URL,
upload capability, storage reference, or authorization claim.

A create input MUST be accepted only by
`POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows`, MUST be
declared by that view, and MUST be dispatched only to its source owner. It MUST
NOT appear in `view_row_v1.cells`, record patch, sort, filter, group, saved
layout, clipboard, or import behavior unless a future owner independently
adopts that behavior. Unknown create inputs MUST fail closed. No new
Evidence-specific row-create route is permitted.

### Same-visible-flow and provisional-state mapping

| Step | Client or server action | Authoritative effect permitted |
| --- | --- | --- |
| 1 | Client opens a local Evidence draft. | None; no `record_id` or `row_version` exists. |
| 2 | Client calls `POST /api/v1/object-blobs`. | One hidden pending blob slot may be created under the existing blob-create idempotency contract. |
| 3 | Client sends bytes to the returned opaque same-origin `PUT` capability. | Byte transfer only; no row, association, projection, available state, or `record_changed` event. |
| 4 | Client calls the generic Evidence row-create route with authored fields and optional `evidence.initial_object_blob_id`. | Request validation and replay resolution only until the final transaction. |
| 5 | The generic create operation finalizes the blob when supplied and commits the first Evidence row. | The complete structured effect set below commits atomically, or no Evidence row commits. |

Before step 5 succeeds, the draft MUST remain client-local and retryable. It
MUST NOT be queryable, collaborative, searchable, saved-view state, history,
export, backup, portability, or other authoritative incident state.

### Minimum-create-signal and default mapping

Presence of `evidence.initial_object_blob_id` does not itself satisfy the
minimum signal. Only successful finalization and association before first
commit qualifies.

| Supplied create content | Minimum signal | Required disposition |
| --- | --- | --- |
| Non-empty normalized `evidence.title` | yes | A no-blob create may commit. |
| Explicit valid `evidence.lifecycle_state`, including explicit `requested` | yes | The user-supplied semantic choice qualifies. |
| Non-null valid `evidence.requested_at` | yes | A no-blob create may commit. |
| Non-null valid `evidence.received_at` | yes | A no-blob create may commit. |
| Non-empty external `evidence.storage_ref` | yes | A no-blob create may commit; user input matching reserved `object://...` remains forbidden. |
| Non-empty normalized `evidence.collector_party_text` | yes | A stable Party link is not required. |
| Non-empty normalized `evidence.source_party_text` | yes | A stable Party link is not required. |
| `evidence.collector_party_id` only | no | Reject unless another qualifying value or finalized blob exists. |
| `evidence.source_party_id` only | no | Reject unless another qualifying value or finalized blob exists. |
| Party ID plus qualifying Party text | yes | Validate and persist both; do not erase the text. |
| Server-filled lifecycle or timestamp default | no | A default MUST NOT make a blank create valid. |
| Read-only or derived field | no | Reject as an unsupported client write. |
| Text normalized to empty, including whitespace/control-only input | no | Treat as absent after the applicable string contract. |
| Explicit JSON `null` for a clearable field | no | A clear is not a non-empty create signal. |
| Pending initial blob with no observed successful upload | no | Return `blob_pending`; commit no Evidence row. |
| Initial blob successfully finalized in this flow | yes | Create exactly one Evidence row. |
| Qualifying field plus a supplied blob whose finalization fails | no commit | Blob finalization is mandatory when the input is supplied; do not fall back to field-only create. |
| Preseeded related-record context | no | Context alone MUST NOT create an Evidence row. |

Omitted lifecycle MUST default to `requested`, but that default MUST NOT
qualify as a signal. Explicit `requested` does qualify. For persisted
normalization and row-create replay comparison, omitted lifecycle and explicit
`requested` MUST compare equal. In either case, omitted `requested_at` MUST
default to the commit timestamp. Explicit `requested_at:null` MUST be an
explicit clear, MUST prevent that timestamp default, and MUST NOT qualify as a
signal.

### Initial lifecycle mapping

Creation is an initial-state operation, not a transition from an invented prior
state.

| Initial Evidence lifecycle | No blob | Same-flow blob finalized as `available` |
| --- | --- | --- |
| `requested` | allowed | allowed |
| `pending_receipt` | allowed | allowed |
| `received` | allowed | allowed |
| `available` | rejected | allowed |
| `quarantined` | allowed | rejected |
| `released` | rejected | rejected |

Omitted lifecycle MUST remain `requested` even for a blob-backed create.
Finalizing a blob MUST NOT auto-promote Evidence to `available`. An available
blob associated with `requested`, `pending_receipt`, or `received` Evidence
MUST remain non-previewable until the Evidence lifecycle reaches `available` or
`released`. Entry to `released` still requires a later explicit
`available -> released` transition.

### Transaction, replay, concurrency, and failure mapping

The final authoritative Postgres transaction MUST serialize the blob slot and
association, revalidate that the incident is open, transition the pending blob
to `available`, create the record envelope and Evidence source row, associate
the blob, generate the server-managed logical `object://...` reference, apply
defaults, append custody and revision history, refresh projections, append the
durable Collaboration intent, and record the successful idempotency result.
These structured visibility effects MUST commit together or not at all.

Uploaded bytes exist outside that transaction. If structured commit fails, no
row, association, successful idempotency record, projection, or collaboration
publication may become visible. The hidden pending slot remains retryable and
subject to its existing timeout and cleanup rules. An observed accepted-contract
mismatch MAY durably mark the blob `failed`, as Core 02 requires, but MUST
still create no Evidence row or projection.

The row-create idempotency key remains
`(actor_user_id, incident_id, view_schema_id, client_txn_id)`. The normalized
request MUST include all recognized normalized fields, the initial blob ID when
present, and defaults under the owner equivalence rules.

| Case | Required result |
| --- | --- |
| First successful create | `201 Created`; exactly one row, change set, revision result, projection refresh, durable intent, and idempotency result. |
| Exact replay after response uncertainty | `200 OK` with the original committed result, not current mutable row state; no new effect. |
| Same key with different fields or blob ID | `409 client_txn_conflict`; no new effect. |
| Two transactions race for one initial blob | At most one succeeds; the other creates no row or association. |
| Fresh key reuses a blob already associated elsewhere | `409 evidence_attach_rejected` with `blob_not_visible`; do not disclose the other association. |
| Commit succeeds but response delivery fails | Exact replay returns the original committed result. |
| Collaboration delivery fails after commit | Durable intent remains replayable; source mutation MUST NOT be repeated. |

The live migration defines only a non-unique
`evidence_object_blob_idx`. SL-07 therefore MUST adopt and enforce the invariant
that one blob is authoritatively associated with at most one current Evidence
record, using transaction locking and a forward migration for the required
uniqueness constraint. This is a later schema change and MUST NOT be inferred
from current storage.

For a non-replay Evidence create, evaluation precedence MUST be:

1. authenticate the current session;
2. enforce CSRF for cookie-authenticated mutation;
3. resolve incident visibility and the ordinary create role;
4. resolve the addressed view schema;
5. strictly decode the closed request shape;
6. normalize recognized fields and create inputs;
7. resolve committed replay or divergent-key conflict;
8. evaluate the minimum signal;
9. resolve and conceal blob visibility and structured state;
10. perform required object-store verification;
11. recheck incident lifecycle and mutable preconditions in the final
    transaction;
12. commit source, history, projection, durable intent, and idempotency effects;
13. deliver collaboration or the HTTP response only after commit.

| Condition | Public result | Durable result |
| --- | --- | --- |
| No qualifying field or successfully finalized blob | `400 invalid_mutation_payload`, `reason_code=minimum_create_signal_missing` | Nothing commits. |
| Missing, null, or malformed create input | `400 invalid_mutation_payload`; `error.details.field=evidence.initial_object_blob_id` | Nothing commits. |
| Missing, cross-incident, or already-associated blob | `409 evidence_attach_rejected`, `reason_code=blob_not_visible` | No Evidence row commits. |
| No successfully observed upload | `409 evidence_attach_rejected`, `reason_code=blob_pending` | No row; slot remains pending. |
| Failed blob | `409 evidence_attach_rejected`, `reason_code=blob_failed` | No Evidence row commits. |
| Quarantined blob | `409 evidence_attach_rejected`, `reason_code=blob_quarantined` | No Evidence row commits. |
| Size or expected-hash mismatch | `409 evidence_attach_rejected`, `reason_code=accepted_contract_mismatch` | Blob may become terminal `failed`; no Evidence row or projection commits. |
| Illegal requested lifecycle | Existing `illegal_transition` result | No row or association commits. |
| Object-store dependency failure | Existing stable redacted object-store result after visibility/state checks | No row; retry state follows existing finalization rules. |
| Incident closes before commit | Existing `incident_closed` result | No fresh source mutation commits. |
| Exact successful replay | `200 OK` with the original result | No new effect. |

No failure may expose a bucket, physical object key, backend endpoint,
credential, raw provider error, foreign incident identifier, or foreign
Evidence association.

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| `Store` combines persistence, lifecycle, handles, cleanup, idempotency, revisions, projections, and collaboration | `store.go` imports and method surface | Transaction/order regressions from large edits | `must_fix` | Evidence facade over private Evidence repositories/services | Characterize transactions and split behind the existing facade. |
| Route registration mixes decoding, authorization, capability issuance, object-store adaptation, service calls, and public error mapping | `api.go`, `routes.go`, `upload_token.go` | Wire, concealment, and secret-redaction regression | `should_fix` | Evidence transport adapter plus platform contracts | Split internally while freezing six operations and all envelopes. |
| Evidence Timeline fact reader returns a Timeline-owned projection DTO | Direct import of `internal/modules/timeline/workbookprojection` | Reverse module dependency and model leakage | `must_fix` | Evidence-owned fact DTO; application mapper | Add parity tests, then move only the DTO translation. |
| Evidence construction and public route store are exposed through `timelineassembly.Bundle.EvidenceStore` | Timeline assembly and Server route registration | Misleading ownership, broad store leakage, and future composition cycles | `must_fix` | `internal/app/server` composition using Evidence-owned constructors | Complete SL-00 and SL-05, then execute SL-06 without creating another broad bundle or store accessor. |
| Workbook facade repeats generic revision/conflict/projection/collaboration choreography | Facade imports and generic Workbook tests | Divergent replay/conflict semantics across source owners | `should_fix` | Evidence validation/persistence plus shared owner contracts | Separate owner semantics from generic mechanics without inventing a new runtime abstraction in this task. |
| Source-owner projection/import/portability/recovery/reporting/revision providers live under Evidence | Provider construction and assembly catalogs | Moving them to consumers would invert source authority | `intentional/no_action` | Evidence source owner | Keep contributions; isolate interfaces and one-way imports. |
| Direct SQL is widespread in the root facade and owner adapters | Store, workbook mutation, provider files | Hard-to-characterize transaction and table ownership | `should_fix` | Evidence private persistence; source-owner adapters where required | Centralize root persistence without moving fixed-query owner contributions to consumers. |
| Revision/projection/collaboration side effects are embedded in mutation methods | Store and workbook facade commit paths | Partial success or duplicate effects | `must_fix` | Evidence mutation coordinator using explicit ports | Freeze ordering, rollback, replay, and exactly-once intent outcomes before extraction. |
| Authorization is enforced in route/application admission and rechecked on handle redemption | Route handlers and tests | Moving checks inward/outward could widen access | `intentional/no_action` | Core 04 route/application security boundary | Preserve current membership/role derivation and no `deployment_admin` bypass. |
| Machine boundary policy protects Evidence delete/restore and rollback subpackages but not the mixed root | `tools/backend_module_boundaries.json` | Future peer coupling can reaccumulate | `should_fix` | Backend boundary owner with Evidence collaborators | Add rules only after the target private facade/port shape exists. |
| No production test-util imports exist | Production/test import scan | None currently | `intentional/no_action` | Test support owner | Keep all app/test support imports confined to tests. |
| No direct grid-vendor or production frontend dependency exists | Target file/import scan | None currently | `intentional/no_action` | Grid Adapter/Web owners | Retain indirect generated contract and selector validation. |
| Core 01 requires a blob-backed normal create but declares no create-input interface; current validation also rejects some owner-permitted field signals | REQ-01-057, REQ-01-245, REQ-01-288, REQ-01-328 versus authored schema, `create_validation.go`, and routes | An owner-interface omission and implementation gap can be hidden or inconsistently repaired | `must_fix` | Core 01–04 adoption package, then Evidence through the generic Workbook route | Keep current behavior visible during structural work; implement only in separately authorized SL-07 after owner adoption. |
| Evidence blob association has a non-unique index and no database uniqueness constraint | `db/migrations/00015_evidence_and_object_blobs.sql` | Concurrent initial creates could associate one blob with multiple Evidence rows unless serialized and constrained | `must_fix` | Evidence storage invariant adopted by Core owner; forward migration in SL-07 | Make one-current-association, locking, and migration explicit prerequisites; do not claim the invariant exists today. |
| Generated outputs reflect Evidence contracts and must not be edited during a structural refactor | Generated policy and generated packages | Hand-edit/drift risk | `intentional/no_action` | Contract generators | Change authored inputs only when behavior changes, then run Make generation. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | None | WF-01 | Fix baseline, authority, scope, and allowed writes | Tracker and read-only owners | `make task-guide ROLE=module-author OWNER=module.evidence` | Baseline and authority recorded. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Account for all 40 tracked entries and live callers/dependencies | Evidence tree and inbound composition | Inventory count and repository searches | All 39 Go files have behavior/evidence rows and `.gitkeep` is explicitly non-behavioral. |
| WF-02 | Contract-owner mapping | parallel | WF-01 | WF-03, WF-05 | Map every observable contract and define the RB-001 Core 01–04 adoption package without claiming adoption | Core owners, contracts, routes, providers | `make explain-test-owner OWNER=module.evidence` | Freeze map has owner, current/proposed posture, test posture, and adoption gate for each contract. |
| WF-03 | Characterization test gap analysis | parallel | WF-01, WF-02 | WF-05, WF-06, WF-07 | Specify the four-part SL-00 package: posture split, create/lifecycle matrix, durable-effect faults, and provider/composition seams | Evidence and cross-owner tests/catalog | Evidence owner diagnostics and exact selector review | Every risky behavior has retained or required coverage and current non-conformance is informative only. |
| WF-04 | Boundary/coupling scan | parallel | WF-01 | WF-05 | Classify imports, SQL, side effects, adapters, frontend/grid absence, and test leakage | Evidence imports, app composition, boundary policy | `make backend-module-boundary-check` after later edits | Findings classified without inferring behavior from names. |
| WF-05 | Facade and ownership redesign plan | chain | WF-02, WF-03, WF-04 | WF-06 | Define behavior-preserving private seams and the adopted Server-composition direction | Evidence facade/store/routes/providers; Server and application adapters | Characterization matrix review | Public facade and owner contributions remain stable; no broad replacement bundle is introduced. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07, WF-08 | Order reversible structural slices and isolate behavior correction | Packages named in Section 7 | Per-slice owner and boundary commands | Every slice has dependency, rollback, and completion criteria. |
| WF-07 | Harness/test/accounting update plan | parallel | WF-03, WF-06 | WF-08 | Keep active tests selected exactly once after authorized moves/additions | Authored verification/test-family inputs and tests | Owner slices, drift, harness checks | Every executable identity has an explicit disposition. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | None | Record narrow-to-broad results, failures, run roots, and residual blockers | Later changed files and tracker | `make agent-finalize` plus risk-appropriate gates | Another agent can resume without rediscovery. |

### SL-00 characterization package

SL-00 MUST complete all four groups below before any production movement.

| Group | Required evidence | Binary exit |
| --- | --- | --- |
| Current behavior versus conformance posture | Tests for current non-conforming create outcomes are visibly named and cataloged with `claim_posture=informative`; they cite no Core acceptance criterion as satisfied. Corrected tests remain inactive or otherwise non-claiming until owner adoption and SL-07. | Every current-behavior test has an SL-07 retirement or conversion disposition, and no wrong outcome closes Core conformance. |
| Create-signal and lifecycle matrix | Parameterized owner and public-route cases cover blank input, every writable field, string normalization boundaries and controls, all initial lifecycle states, timestamps, Party IDs/text, read-only and derived fields, external/reserved storage refs, each blob state, replay, concurrency, publication, and provisional frontend behavior. | Every row in the Section 4 signal, lifecycle, replay, and error tables is selected exactly once under an accountable owner. |
| Facade durable-effect faults | Faults are injected before envelope creation; after envelope, Evidence row, blob link, revision, projection, durable intent, and idempotency writes; at commit failure, HTTP response loss, dispatcher failure, object-store unavailability, and accepted-contract mismatch. | Each fault proves the allowed durable result from Section 4; tests assert persisted/public effects, not private helper call order. |
| Provider and application composition seams | Direct contribution tests and consumer catalog tests cover every mapping below; Server seam tests cover construction, registration, resource ownership, readiness, and static dependency direction. | Every contribution has an owner-local constructor test and an accountable consumer integration test, and all composition invariants below pass. |

| Contribution or behavior | Accountable test owner | Required characterization |
| --- | --- | --- |
| Public generic Evidence row create | Workbook/view-row-create owner; Evidence collaborator | Exact generic request, response, replay, validation, and current-versus-proposed behavior posture. |
| Evidence normalization and blob finalization | Evidence | Field normalization, minimum signal, lifecycle, storage-ref reservation, blob finalization, atomicity, and concealment. |
| Projection provider and query surfaces | Projections | Exact provider identity, query/field closure, duplicate rejection, refresh, rebuild, and deterministic output. |
| Import create and row refresh | Imports | Target identity, caller-owned transaction, provenance, replay, rollback, and returned Evidence row. |
| Incident Bundle source port | Incident Portability | Exact descriptor and member set without private implementation dependency. |
| Evidence row portability | Incident Portability | FK order, lifecycle, attribution, staging invisibility, validation, and final transaction. |
| Blob portability | Incident Portability | Digest verification, logical-key rewrite, hidden staging, cleanup, and no physical locator leakage. |
| Evidence subtype presence | Records with Incident Portability collaborator | Exact subtype counts and retained-state semantics. |
| Recovery state and inventory | Recovery | Authoritative/rebuildable/ephemeral classification, complete available-object inventory, and no handles. |
| Recovery provider | Recovery | Read-only row/object counts, digest metadata, and failure propagation. |
| Reporting provider | Reporting | Deterministic redacted facts with no object-blob ID, physical key, secret, or live-access semantics. |
| Revision contribution | Revisions | Exact delete/restore, rollback, live-change, and view-routing registration. |
| Delete/restore and rollback providers | Revisions | Source snapshots, preconditions, lifecycle/reference/association guards, and row restoration. |
| Timeline attachment validation | Timeline; Evidence collaborator | Same-incident active Evidence validation inside the initiating transaction. |
| Timeline fact mapping | Timeline; Evidence collaborator | Lossless identifier, ordering, label, lifecycle, upload-state, and count parity before and after DTO decoupling. |
| Evidence construction | Server/application owner | Single runtime, registration completeness, construction order, readiness failure, and resource ownership. |

Application-composition tests MUST establish:

1. exactly one Evidence owner runtime exists per Server runtime;
2. routes, Workbook, Timeline adapters, and owner contributions derive from that
   runtime where state is shared;
3. missing or duplicate required contributions fail before listener startup;
4. projection-catalog validation precedes dependent Evidence facade use;
5. production Server exposes no Evidence store accessor;
6. Evidence route registration does not require `timelineassembly.Bundle`;
7. the Evidence-to-Timeline mapping is lossless;
8. assembly failure closes owned resources only;
9. borrowed Postgres and object-store dependencies remain open;
10. repeated `Runtime.Close` is harmless;
11. no mutable global or package-initialization registry participates;
12. reusable tests compose through application support, not a production store
    accessor;
13. static boundary checks reject cross-module implementation imports, private
    persistence access, and Server bypass of Evidence-owned constructors.

Every new or moved test MUST have exactly one active owner row, one executable
identity, and separately declared collaborators. Authored catalog and family
inputs MUST be changed before regeneration; generated topology MUST NOT be
hand-edited. Omitting `ROWS` MUST continue to select the full owner slice. A
refactor-phase owner such as `evidence-refactor` MUST NOT be created.

## 7. Proposed Refactor Slice Plan

All structural slices below preserve observable behavior. SL-07 is the only
planned behavior correction and is explicitly marked `requires later
authorization`.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| SL-00 | WF-03 | Implement the complete Section 6 characterization package: posture separation, create/lifecycle matrix, durable-effect faults, direct contributions, and Server composition seams | Evidence tests; accountable Workbook, Projection, Import, Portability, Records, Recovery, Reporting, Revision, Timeline, and Server tests; authored harness inputs | Tests could encode current non-conformance as desired behavior or overclaim Core acceptance | Current wrong outcomes MUST be informative with retirement/conversion dispositions; corrected conformance activates only after adoption and SL-07 | `make test-slice OWNER=module.evidence`; `make service-backed-test-slice OWNER=module.evidence`; affected owner slices | Revert only new tests and authored catalog inputs; no production rollback | All four groups and 13 composition invariants pass in their declared posture; every executable identity is selected once. |
| SL-01 | SL-00 | Return an Evidence-owned fact DTO and map it to Timeline's consumer DTO in application composition while retaining the initiating Timeline transaction | `timeline_facts.go`, Timeline consumer port, application collection-read adapter, focused tests | Identifiers, ordering, labels, lifecycle/upload states, attached counts, and transaction scope | Add exact DTO parity and caller-transaction tests; preserve Timeline attached-evidence tests | Evidence and Timeline owner slices; `make backend-module-boundary-check` | Restore the old DTO return and adapter atomically | Evidence and Timeline no longer import each other's implementation DTOs; application mapping is lossless and transaction ownership is unchanged. |
| SL-02 | SL-00, SL-01 | Decompose blob-slot, attachment/lifecycle, access-handle, and cleanup persistence behind the existing `Store` facade | Evidence Store and new private Evidence repository/service files | Transactions, locks, idempotency, lifecycle bridges, cleanup, handle consumption | Preserve all Store/integration/process tests; add transaction-failure parity where missing | Evidence focused and service-backed slices; `make test-fast` | Revert the private extraction as one slice; public Store callers remain unchanged | `Store` remains the facade, private components have one semantic owner each, and behavior/effects are byte- or value-equivalent. |
| SL-03 | SL-00, SL-02 | Separate request decoding, authorization admission, upload capability handling, object-store adaptation, and error translation behind `RegisterRoutes` | Evidence API/routes/upload-token internals | Route paths, JSON strictness, concealment, CSRF, opaque tokens, dependency redaction | Preserve route, object-store, handle, OpenAPI, browser, and auth tests | Evidence slices; `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | Restore prior handler wiring without data migration | Six operations and all public envelopes remain unchanged; route layer delegates to narrow services. |
| SL-04 | SL-00, SL-02 | Separate Evidence field/lifecycle validation and persistence from generic idempotency, conflict, revision, projection, and collaboration mechanics behind `WorkbookFacade` | Workbook facade/conflict/mutation/validation files and assembly adapters | Create/patch/replay/no-op/conflict, sparse cells, history, transaction order | Preserve Workbook mutation, conflict, row-wire, party, import, WS tests | Evidence, Workbook, Revisions, and Projections owner slices | Reconnect the prior facade internals; no schema rollback | Evidence semantics remain owner-local, generic mechanics use explicit ports, and public generic routes are unchanged. |
| SL-05 | SL-01, SL-02, SL-04 | Isolate import, projection, portability, recovery, reporting, and revision contributions behind their existing consumer contracts and strengthen one-way import rules | Evidence provider/contribution subpackages, application catalogs, boundary policy | Provider IDs, catalog order, transaction ownership, redaction, restore and bundle invariants | Preserve each consumer-owner integration suite and add direct contribution tests only where Section 4 identifies a gap | Evidence plus affected owner slices; `make backend-module-boundary-check`; drift checks if authored inputs change | Revert provider moves and matching boundary rules together | Each contribution is constructed by Evidence, consumed only through the owner contract, and imports no consumer implementation internals. |
| SL-06 | SL-00, SL-05 | Move Evidence construction to `internal/app/server`, remove `timelineassembly.Bundle.EvidenceStore`, and inject only the Section 3 narrow capabilities | Server and Timeline assembly, Evidence constructors, application adapters, Server process/application-support tests | Runtime identity, projection validation order, route wiring, readiness, owned/borrowed resources, and shutdown | Preserve all Section 6 composition invariants and Server lifecycle tests | Evidence, Timeline, and Server owner checks; `make backend-module-boundary-check`; `make test-fast` | Restore the prior Timeline bundle field and wiring atomically; introduce no compatibility store accessor | Server constructs one Evidence runtime, Timeline bundle exposes no Evidence implementation, readiness remains fail-closed, and resource ownership is unchanged. |
| SL-07 | SL-00; RB-001 owner-adoption exit; separate behavior authorization | `requires later authorization`: implement the adopted `create_inputs[]` interface, complete Evidence first-create behavior, and enforce one-current-Evidence-association per blob with locking and a forward migration | Core 01–04 owner documents in an owner task; authored view-schema/OpenAPI/error contracts; Evidence validation/facade/routes; Workbook create coordination; frontend controller; migration; generated outputs through Make | Intentional public behavior and discovery change, schema migration, partial row/blob commit, lifecycle, replay, concurrency, projection, UI draft, and error precedence | Convert informative cases to owner conformance; add full signal/lifecycle/error matrix, fault atomicity, concurrent blob consumption, replay, and browser same-flow tests | Evidence, Workbook, View Contracts, and affected owner slices; browser targets; `make generate`; drift/policy/migration checks | Revert owner implementation, migration, authored inputs, and generated outputs as one authorized slice; no partial compatibility path | Adopted Core outcomes and every Section 4 matrix row pass; one blob cannot bind two current Evidence rows; failed flow creates no row, false projection, duplicate effect, or leakage. |
| SL-08 | SL-03, SL-04, SL-05, SL-06; SL-07 only when authorized | Update authored harness ownership for moved/added tests, regenerate topology through Make, run final verification, and refresh handoff | Verification owner/test-family inputs, generated topology through Make, tracker | Missing/duplicate selection, stale informative rows, or overclaimed evidence | Every retained/new test resolves to one active row; SL-07 conversions occur only if that slice ran | `make generate-drift`; `make generated-artifact-policy-check`; `make test-fast`; `make agent-finalize`; risk-appropriate `make check` | Revert authored accounting and generated refresh together | Required checks and run roots are recorded; no unexplained failure, stale posture, or unaccounted executable test remains. |

## 8. Validation Plan

Current harness discovery reports 35 `module.evidence` rows: 20 Go, 12
Playwright, and 3 Vitest. Thirty rows are service-backed. The families cover
accessibility, browser, stateful browser, frontend, frontend unit, integration,
store, unit, and visual evidence. Current-behavior tests for RB-001 MUST use
`claim_posture=informative`, MUST NOT claim Core acceptance closure, and MUST
name their SL-07 retirement or conversion disposition. Corrected conformance
rows MUST activate only after owner adoption and implementation.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make test-slice OWNER=module.evidence` | Focused owner-selected Evidence rows | yes | Use `ROWS=<row-id,...>` only for an exact diagnostic subset; an omitted `ROWS` must remain the full owner selection. |
| integration | `make service-backed-test-slice OWNER=module.evidence` | Postgres/object-store-backed Evidence rows | yes | Run affected Timeline, Workbook, Projections, Imports, Incident Bundles, Recovery, Reporting, and Revisions owner slices when their boundary is touched. |
| e2e/browser | `make browser-e2e-webserver-backed` and `make browser-e2e-stateful` | Public upload, attach, preview/download, auth, live refresh, and state-loss paths | yes for route, access, composition, or frontend-affecting slices | Run `make browser-e2e-a11y` or `make browser-e2e-visual` only when those owned contracts change. |
| generated drift | `make generate-drift` and `make generated-artifact-policy-check` | Authored-to-generated parity and hand-edit protection | yes when contracts, providers, or harness inputs change | Structural code-only slices should produce no generated diff. |
| import-boundary/static | `make backend-module-boundary-check` | Production package imports and owner-port policy | yes | Add new Evidence rules only after the intended private facade exists. Frontend boundary checking is required only if later authorization touches frontend files. |
| full check | `make test-fast`, then `make agent-finalize`, then risk-appropriate `make check` | Cross-owner and repository closure | no; required at final closure according to risk | `make check` needs Postgres, object store, and browser stack. Record run roots and related/unrelated failure disposition. |

For this tracker-only change, use `make lint-markdown` and
`git diff --check`, then inspect `git status --short`. Because this tracker is
currently untracked, also use
`git diff --no-index --check /dev/null docs/handoffs/evidence-module-refactor-tracker.md`.
For that last diagnostic, exit `1` with no whitespace output is the expected
content-difference result; whitespace diagnostics or exit `3` fail the check.
No product test result is claimed by this planning document.

This revision ran `make lint-markdown` successfully with run root
`.cartulary/test-results/20260730T220244Z-p2139397`. `git diff --check`
produced no diagnostics. The untracked-file check exited `1` with no output,
which is the expected clean content-difference result. `git status --short`
listed only this tracker.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| EV-001 | Fix scope, authority, baseline, and write boundary | WF-00 | DONE | None | Section 1 | Planning-only posture and later authorization boundary are explicit. |
| EV-002 | Inventory all 40 tracked target entries | WF-01 | DONE | EV-001 | Section 2 | All 39 Go files have populated rows and `.gitkeep` is explicitly non-behavioral. |
| EV-003 | Diagnose the current module boundary | WF-04 | DONE | EV-002 | Sections 3 and 5 | Legitimate owner behavior, mixed responsibilities, moves, keeps, and deferrals are distinguished. |
| EV-004 | Freeze current contracts and specify the proposed owner-adoption closure | WF-02 | DONE | EV-002 | Section 4 | Every contract has current/proposed posture, owner, interface/default mapping, tests, and risk without falsely claiming adoption. |
| EV-005 | Add the four-part characterization baseline | WF-03 | TODO | EV-004 | SL-00 and Section 6 | Every structural slice has executable invariants; current wrong behavior is informative only. |
| EV-006 | Adopt and implement the Evidence create-input contract | WF-02, WF-03 | BLOCKED | EV-005; Core 01–04 adoption; separate behavior authorization | RB-001 and SL-07 | Owner edits and authored projections are adopted, migration/locking enforce one association, and all corrected conformance cases pass. |
| EV-007 | Remove the Timeline projection DTO dependency | WF-05, WF-06 | TODO | EV-005 | SL-01 | Evidence exposes owner facts and application composition maps them losslessly. |
| EV-008 | Decompose the Evidence Store internally | WF-05, WF-06 | TODO | EV-005, EV-007 | SL-02 | Private responsibilities are narrow and facade behavior is unchanged. |
| EV-009 | Separate route transport/admission concerns | WF-05, WF-06 | TODO | EV-008 | SL-03 | Six public operations remain exact and handlers delegate through narrow services. |
| EV-010 | Narrow the workbook facade internals | WF-05, WF-06 | TODO | EV-008 | SL-04 | Evidence semantics and generic mutation mechanics have explicit seams. |
| EV-011 | Isolate source-owner contributions and add boundary rules | WF-05, WF-06 | TODO | EV-007, EV-008, EV-010 | SL-05 | Contributions remain Evidence-owned and one-way. |
| EV-012 | Move Evidence application composition to Server | WF-05, WF-06 | TODO | EV-005, EV-011 | RB-002 decision and SL-06 | Server constructs one Evidence runtime, Timeline exposes no Evidence store, seam tests pass, and resource behavior is preserved. |
| EV-013 | Update exact harness accounting | WF-07 | TODO | EV-005, applicable implementation slices | Authored catalog inputs and SL-08 | Every executable test is selected exactly once. |
| EV-014 | Run closure validation and refresh handoff | WF-08 | TODO | EV-009, EV-010, EV-011, EV-012, EV-013; EV-006 only if SL-07 is authorized | Section 8 and future run roots | Required commands, results, failures, skips, posture conversions, and blockers are current. |

## 10. Session Handoff Log

The tracker did not exist before the original planning session. Its rows are
preserved below, followed by this NLSpec-style revision session. Across both
sessions, only this tracker was touched.

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30T17:07:49-04:00 | Codex planning session | Target exists; authority and planning-only boundary fixed at `367bc76c124a` | Inspected framework, Core 00–04, Domain, NLSpec guidance, Reporting and Testing Harness NLSpecs; touched this tracker only | `sed`, `rg`, `git status`, `git rev-parse` | No owner contradiction; Core 05 inapplicable | RB-001 and RB-002 | Begin SL-00 only in a later authorized task. |
| 2026-07-30T17:57:24-04:00 | Codex NLSpec-style revision session | Authority boundary preserved; RB-001 split into owner-interface and implementation gaps; all 40 tracked entries accounted | Inspected analysis notes, NLSpec guidance, Core 01–04, live contracts/migration, and current tracker; touched this tracker only | `sed`, `rg`, `jq`, `find`, `git status`, `git rev-parse`, `make lint-markdown`, Git whitespace checks | Proposed adoption requirements are explicit without claiming owner adoption; Core 05 remains inapplicable; tracker-only checks passed | RB-001 adoption/authorization gate; RB-003 execution gate | Implement SL-00 in a later authorized task; do not implement proposed behavior before adoption. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30T17:07:49-04:00 | Codex planning session | Legitimate Evidence owner with mixed transport, persistence, mutation, workbook, and provider responsibilities | Inspected all 28 production files, inbound application composition, backend boundary policy; touched this tracker only | `rg --files`, targeted `rg`, `sed` | Store/route decomposition and Timeline DTO coupling are supported; assembly relocation is not yet owner-decided | RB-002 | Add characterization, then implement SL-01 through SL-05 in order. |
| 2026-07-30T17:57:24-04:00 | Codex NLSpec-style revision session | Server selected as application-composition owner; Evidence retains constructors and behavior; Timeline retains consumer transaction and DTO | Inspected Server runtime, Timeline assembly, developer-guide composition rules, Evidence Timeline ports, and boundary policy; touched this tracker only | `sed`, `rg` | RB-002 design decision closed; narrow injections, construction order, prohibited bundle shapes, and 13 seam invariants specified | SL-00 and SL-05 must precede SL-06 | Characterize composition, then remove `timelineassembly.Bundle.EvidenceStore` in SL-06. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30T17:07:49-04:00 | Codex planning session | No production frontend or grid-vendor code in target; indirect Evidence UI contracts remain material | Inspected Evidence view schema, generated UI/protocol surfaces, frontend/browser catalog selectors; touched this tracker only | `jq`, targeted `rg` | No frontend move supported; preserve schema, inspector, selector, accessibility, visual, and browser behavior | RB-001 if same-flow create requires frontend work | Reassess frontend only in separately authorized SL-07 or when a structural slice changes generated consumption. |
| 2026-07-30T17:57:24-04:00 | Codex NLSpec-style revision session | No frontend implementation moves; proposed same-flow create keeps a client-local provisional draft and excludes grid-vendor ownership | Inspected Core 03 upload/draft behavior, view-schema contracts, generated consumers, and analysis recommendations; touched this tracker only | `sed`, `rg`, `jq` | Five-step flow and authoritative/provisional boundary are specified; no durable draft family or vendor callback is introduced | RB-001 adoption and later SL-07 authorization | Add browser conformance only with SL-07; preserve existing frontend contracts during structural slices. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30T17:07:49-04:00 | Codex planning session | Six routes, Evidence view schema, provider/import/recovery/portability contracts, and generated projections mapped | Inspected authored contract inputs, generated operation/type outputs, and generated-artifact policy; touched this tracker only | `jq`, `rg`, `sed`, Make target explanations | Structural slices should not alter wire contracts or generated output | RB-001 may require coordinated authored projection review | Use owner inputs and `make generate`; never hand-edit generated roots. |
| 2026-07-30T17:57:24-04:00 | Codex NLSpec-style revision session | Live `view_schema_resource_v2` and authored schemas lack `create_inputs[]`; proposed exact interface and owner amendment map recorded | Inspected REQ-01-057/061/069/070/245/288/328, authored Evidence view schema, source schema, OpenAPI owner, and generated types; touched this tracker only | `sed`, `rg`, `jq` | RB-001 is confirmed as owner-interface incompleteness plus implementation mismatch; generated files remain downstream only | Core 01–04 adoption and SL-07 authorization | Amend owners first, then authored inputs, then regenerate through Make in the authorized behavior task. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30T17:07:49-04:00 | Codex planning session | 35 owner rows: 20 Go, 12 Playwright, 3 Vitest; 30 service-backed | Inspected all 11 target tests, cross-owner tests, verification owner, family manifest, catalog, topology, and boundary policy; touched this tracker only | `make task-guide ROLE=module-author OWNER=module.evidence`; `make explain-test-owner OWNER=module.evidence`; `make explain-target ...` | Canonical narrow and broad commands discovered; no product suite was run | RB-003 characterization gaps | Implement SL-00 and account for every new or moved test exactly once. |
| 2026-07-30T17:57:24-04:00 | Codex NLSpec-style revision session | Existing 35-row accounting retained; exact four-part SL-00 package, owner mapping, posture rules, and retirement/conversion requirements specified | Inspected Evidence family rows, informative posture schema, all 11 target tests, and cross-owner test responsibilities; touched this tracker only | `sed`, `rg`, `jq` | Current wrong outcomes must be informative; corrected conformance activates only after adoption and SL-07 | RB-003 remains unexecuted | Implement all four SL-00 groups and prove each executable identity is selected once. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30T17:07:49-04:00 | Codex planning session | Route admission and handle redemption implement current membership/session checks and no admin bypass | Inspected Core 04, Evidence routes/store, handle and object-store dependency tests; touched this tracker only | `sed`, targeted `rg` | Security placement is intentional; error redaction and redeem-time revalidation are frozen | None beyond characterization required by SL-00 | Preserve all auth order, concealment, CSRF, hostile-content, and secret-redaction cases. |
| 2026-07-30T17:57:24-04:00 | Codex NLSpec-style revision session | Security boundary remains route/application admission with redeem-time revalidation; create evaluation and concealment precedence is now explicit | Inspected Core error registries, Core 04, Evidence routes/store, and object-store dependency tests; touched this tracker only | `sed`, `rg` | Thirteen-step precedence and public/durable error matrix specified; no storage or foreign-association leakage is permitted | RB-001 before corrected create behavior | Characterize precedence in SL-00 and bind corrected cases to Core only after adoption. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-30T17:07:49-04:00 | Codex planning session | Planning is decision-complete for behavior-preserving slices; behavior correction and assembly relocation remain gated | Inspected this tracker against live target and owner evidence; touched this tracker only | Inventory and owner diagnostics; no implementation or product validation | Refactor sequence, rollback, validation, and handoff checkpoints defined | RB-001, RB-002, RB-003 | Authorize SL-00; do not begin SL-06 or SL-07 without their named decisions/authority. |
| 2026-07-30T17:57:24-04:00 | Codex NLSpec-style revision session | RB-001 and RB-002 design decisions closed; RB-001 remains adoption/authorization-blocked, RB-002 and RB-003 are executable prerequisites | Inspected tracker against analysis notes and live owner, contract, migration, assembly, and harness evidence; touched this tracker only | Read-only repository diagnostics, `make lint-markdown`, and Git whitespace/status checks | Interfaces, defaults, matrices, dependencies, rollback, and binary exits are specified; tracker-only checks passed; no production validation ran | RB-001 owner adoption; RB-003 execution | Run SL-00, proceed through SL-01–SL-06, and run SL-07 only after separate owner adoption and authorization. |

## 11. Open Questions and Blockers

No design question remains open for RB-001 or RB-002. Their stable IDs are
retained for handoff continuity and now identify adoption or execution gates.

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Core 01 requires blob-backed normal Evidence create but does not declare how `object_blob_id` enters the closed generic row-create interface; current validation also rejects qualifying field signals and implements no same-flow blob finalization. | Implementing from the current text would require invention, while structural work could conceal both owner incompleteness and current non-conformance. | Adopt the Section 4 Core 01–04 package, `create_inputs[]` authored projections, one-blob/one-current-Evidence locking and forward migration; separately authorize SL-07; pass corrected owner and browser conformance | Decision closed; `BLOCKED` on owner adoption and later SL-07 authorization. Structural slices MUST preserve and visibly classify current behavior. |
| RB-002 | Evidence construction currently leaks through `timelineassembly.Bundle.EvidenceStore`. The selected replacement owner is `internal/app/server`, using Evidence-owned constructors and narrow consumer ports. | Relocation without seam tests could change instance sharing, projection order, readiness, route registration, or owned/borrowed resource behavior. | Complete SL-00 and SL-05; execute SL-06; pass the 13 composition invariants and static boundary checks | Decision closed; `TODO` after SL-00 and SL-05. No further owner-selection analysis is required. |
| RB-003 | The complete characterization prerequisite is not implemented. | Code movement without the posture split, create/lifecycle cases, durable-effect faults, direct contributions, and composition seams could change behavior or overclaim conformance. | Implement all four SL-00 groups with exact accountable-owner catalog dispositions and retirement/conversion plans | `TODO` executable prerequisite before any production refactor. |

There is no `BLOCKED: owner contradiction` entry because no contradiction
between applicable adopted owners was found.

## 12. Binary Completion Criteria

- [x] All 40 tracked entries under `internal/modules/evidence` are inventoried:
  28 production Go files, 11 Go tests, and the explicitly non-behavioral
  `.gitkeep`.
- [x] Every discovered public or cross-owner contract risk has an owner,
  existing-test posture, required characterization posture, and risk.
- [x] Current adopted behavior, proposed owner amendments, current
  implementation, and later authorized refactor requirements are explicitly
  distinguished.
- [x] The proposed create interface, default rules, visible-flow boundary,
  lifecycle matrix, atomic effect set, idempotency/concurrency outcomes,
  evaluation order, errors, and non-disclosure outcomes are unambiguous and
  testable.
- [x] Every workflow has explicit predecessors, successors, validation, and a
  handoff checkpoint.
- [x] Every implementation slice is behavior-preserving unless explicitly
  marked `requires later authorization`.
- [x] SL-00 defines the posture split, complete create/lifecycle matrix,
  durable-effect fault matrix, direct provider coverage, 13 Server-composition
  invariants, and exact harness-accounting rules.
- [x] Validation commands were discovered from public Make diagnostics; no
  command was invented.
- [x] Applicable owner differences are classified correctly. No owner
  contradiction was found; the Core 01 create gap is recorded separately as an
  incomplete owner interface and a current implementation mismatch.
- [x] Repository/framework mismatches are recorded: the live target is a real
  core Evidence owner, not merely a generic refactor target, and it contains
  more source-owner contributions and cross-owner coordination than a thin
  module template would imply; the live blob-association index is non-unique.
- [x] Handoff sections contain the baseline, inspected surfaces, commands,
  results, blockers, and next actions needed to continue without rediscovery.
- [x] RB-001 through RB-003 are concrete adoption or execution gates with
  binary exit conditions; no unsupported owner adoption or implementation
  completion is claimed.

The tracker is complete as planning documentation. Production completion is
not claimed: RB-001 remains blocked on owner adoption and separate behavior
authorization, while RB-002 and RB-003 remain executable `TODO` gates.
