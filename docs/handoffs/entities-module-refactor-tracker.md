# entities Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

- **Target path:** `internal/modules/entities`
- **Target label:** `entities`, derived from the target path and normalized to lowercase kebab case.
- **Output path:** `docs/handoffs/entities-module-refactor-tracker.md`
- **Status:** The original remediation is complete through S-06, and the
  Entities Production-Readiness iteration in Sections 14 through 22 is
  complete through EPR-S06. The Entities Source Integrity iteration in
  Sections 23 through 31 is complete through ESI-S07. Sections 32 through 38
  are the controlling plan for the Entities Legacy Removal iteration.
  ELR-S06 final validation and handoff is complete. The Entities Legacy
  Removal iteration has no active or pending successor.
- **Allowed change in this effort:** Execute ELR-S01 through ELR-S06, including
  owner amendments, implementation, tests, authored contracts, Make-generated
  projections, verification evidence, and this tracker. Every successor
  remains subject to the serial checkpoint gate in Sections 35 through 37.
- **Non-goals for ELR-S00:** No product, test, contract, generated, migration,
  domain, application-composition, route, operation, schema, field-key, event,
  authorization, idempotency, or visual-golden change.
- **Implementation authorization:** The 2026-08-21 and 2026-08-22 directives
  for S-01 through S-06 and EPR-S00 through EPR-S06 are complete historical
  authority. The 2026-08-22 ESI implementation directive authorizes ESI-S01
  through ESI-S07 as completed historical work. The 2026-08-29 document-update
  directive authorized ELR-S00. The later 2026-08-29 implementation directive
  authorizes ELR-S01 through ELR-S06 subject to their serial gates.

### Normative language and artifact role

Within this tracker, `MUST` and `MUST NOT` define mandatory conditions for the
authorized Entities refactor. `MAY` identifies intentional implementation
latitude whose alternatives are equivalent at the owned boundary. Descriptive
current-state statements do not create requirements.

This tracker is an implementation-planning and handoff artifact. It MUST NOT
supersede an adopted subsystem NLSpec, Core owner section, or adopted
implementation decision. A requirement in this tracker becomes implementation
authority only through the owner-adoption process named by that requirement.
`docs/research/nlspec-spec.md` supplies writing discipline only, and
`temp/analysis-notes.md` is advisory input only.

The source hierarchy used for this tracker is:

1. Adopted subsystem NLSpecs, within each named subsystem.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. Domain vocabulary and implementation-support guides.
5. Current repository code and tests.
6. Prior plans, handoffs, and the planning framework as evidence only.

No adopted Entities-specific subsystem NLSpec was found. Adopted subsystem owners such as the Testing Harness govern only their corresponding seams. No owner contradiction was found during this inspection. Any later contradiction must be recorded as `BLOCKED: owner contradiction` without selecting a side.

### Owner documents inspected

- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/spec/05_claim_publication_and_benchmark_reproducibility.md`
- `docs/domain.md`
- `docs/decisions/workbook-module-boundary.md`
- `docs/decisions/revisions-module-boundary.md`
- `docs/decisions/projections-module-boundary.md`
- `docs/decisions/entities-module-boundary.md`, adopted during S-01.
- `docs/testing-harness-nlspec.md`
- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `docs/research/nlspec-spec.md` as writing guidance, not behavioral authority.
- `temp/analysis-notes.md` as advisory analysis, not repository authority.

### Repository evidence inspected

- Every file under `internal/modules/entities`, individually inventoried in Section 2.
- Composition in `internal/app/server`, `timelineassembly`, `workbookassembly`, `importassembly`, `projectionassembly`, `revisionassembly`, `recoveryassembly`, and `incidentportabilityassembly`.
- Timeline auto-resolution and mention-collection storage under `internal/modules/timeline`.
- Entity and mention surfaces under `apps/web/src` and their use of `packages/grid-adapter`.
- Authored OpenAPI, view-schema, projection-provider, import, revision, recovery, incident-bundle, inspector, and verification contracts.
- Generated protocol, view-contract, UI-contract, and OpenAPI projections as read-only evidence.
- `tools/test_families/module.entities.json`, `tools/test_support_inventory.json`, and related harness ownership inputs.

## 2. Current-State Repository Inventory

The target contains 70 Go files: the 68-file planning baseline plus the
production registry and its conformance test added in S-05. The current split
is 51 production/support files and 19 test files. Every file is in scope for
inventory; none is excluded.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/entities/boundary_guard_test.go` | Static boundary assertions for the Entities tree. | Seven `TestEntities*`/boundary tests. | Go test discovery. | Filesystem source inspection. | This file. | Test-family accounting only. | Entities / Testing Harness. | high | All seven top-level tests are absent from the focused owner manifest. |
| `internal/modules/entities/entity_alias_migration_test.go` | Service-backed alias head-schema contract. | `TestEntityAliasHeadSchemaContract_Integration`. | Go test discovery. | PostgreSQL test support and migrations. | This file. | Alias database schema. | Entities. | high | Selected by the owner manifest. |
| `internal/modules/entities/http_helpers.go` | Shared HTTP parsing and API-error mapping for entity routes. | Package-private helpers. | `routes.go`. | `httpapi` and merge error contracts. | Root route/unit/integration tests. | OpenAPI error envelopes indirectly. | Entities transport facade. | high | Error precedence is observable behavior. |
| `internal/modules/entities/incident_bundle_portability.go` | Exports and imports entity-owned incident-bundle rows. | `ExportIncidentBundleFiles`, `ImportIncidentBundleFilesTx`. | Incident portability source port and assembly. | PostgreSQL, incident portability contracts, entity tables. | Incident portability verification through collaborating suites. | Incident-bundle source catalog and row files. | Entities source owner. | high | Owner-local SQL is intentional. |
| `internal/modules/entities/incident_bundle_source_port.go` | Adapts entity portability functions to the generic source port. | `NewIncidentBundleSourcePort`. | `incidentportabilityassembly`. | Incident portability source-port contract. | Collaborating incident-bundle tests. | `contracts/incident-bundles/source_catalog.json`. | Entities source owner. | medium | Thin source-owned adapter. |
| `internal/modules/entities/incident_bundle_subtype_presence.go` | Declares subtype-presence bindings for hosts and identities. | `IncidentBundleSubtypeContribution`. | `incidentportabilityassembly`. | Subtype-presence contract and PostgreSQL. | Collaborating portability tests. | Incident-bundle subtype contracts. | Entities source owner. | medium | Keep semantic subtype ownership here. |
| `internal/modules/entities/openapi_contract_test.go` | Characterizes entity mutation OpenAPI shapes. | `TestOpenAPIMutationContractShape`. | Go test discovery. | Authored and generated OpenAPI. | This file. | Entity OpenAPI owner projection. | Entities / API contracts. | high | Selected by the owner manifest. |
| `internal/modules/entities/recovery_state.go` | Contributes entity tables to recovery state coverage. | `RecoveryStateContribution`. | `recoveryassembly`. | Recovery state contract. | Recovery catalog/coverage tests. | Recovery state catalog projections. | Entities source owner. | medium | Source-owned contribution; Recovery coordinates. |
| `internal/modules/entities/resolution_integration_test.go` | End-to-end backend tests for create, mention resolution, merge, auth, CSRF, and idempotency. | Five integration tests. | Go test discovery. | Server test support, PostgreSQL, routes. | This file. | HTTP and storage contracts. | Entities. | high | All five tests are selected by the owner manifest. |
| `internal/modules/entities/revision_provider_contribution.go` | Aggregates host, identity, alias, preserved-identifier, and mention revision providers. | `RevisionProviderContribution`. | `revisionassembly`. | Revisions and owner-local rollback/delete-restore providers. | Contribution unit test and revision integration suites. | Revision snapshot and target registries. | Entities source owner. | high | Revisions retains generic coordination. |
| `internal/modules/entities/revision_provider_contribution_test.go` | Verifies the entity history-facet contribution. | `TestRevisionProviderContributionOwnsEntityHistoryFacet_Unit`. | Go test discovery. | Revision contribution. | This file. | Revision target semantics. | Entities / Revisions. | medium | Selected by the owner manifest. |
| `internal/modules/entities/routes.go` | Registers merge and mention-action HTTP routes and enforces authentication, CSRF, authorization, and error ordering. | `Service`, `RouteOptions`, `RegisterRoutes`. | `internal/app/server/runtime_assembly.go`. | Mentions, merge, incident admission, authn, HTTP platform. | Root unit/integration/support/OpenAPI tests. | `module.entities` OpenAPI owner document. | Entities transport facade. | high | Legitimate thin transport-adjacent facade. |
| `internal/modules/entities/support_integration_test.go` | Reusable conformance scenarios for envelopes, CSRF, replay, authorization, query metadata, projections, WebSockets, and schema. | Seven tests plus `SupportScenario`/`SupportSuite`. | Go test discovery. | Server test support and generated contracts. | This file. | Record, view, and WebSocket contracts. | Entities / test support. | high | All seven tests are selected. |
| `internal/modules/entities/unit_test.go` | In-memory transaction characterization for create-from-mention, lifecycle, exact matching, and merge. | Four unit tests. | Go test discovery. | Hostidentity, mentions, merge, test doubles. | This file. | Core entity semantics. | Entities. | high | All four tests are selected. |
| `internal/modules/entities/entitycontract/contract.go` | Stable host/identity record-type and view-schema identifiers. | Exported constants for host and identity contracts. | Hostidentity and merge collaboration code. | No owner-external implementation dependency. | Entity unit/integration and contract tests indirectly. | Hosts/Identities view schemas. | Entities shared kernel. | medium | Small owner-local semantic contract. |
| `internal/modules/entities/hostidentity/api.go` | Host/identity request decoding, DTOs, registry-backed row materialization entry points, alias actions, hashes, and mutation envelopes. | Create/record/alias DTOs and build/decode helpers. | Workbook adapters, stores, merge aliases, tests. | Entity contract, field registry, HTTP API, UUID, normalization helpers. | Admission, registry, route, unit, and integration tests. | View schemas and protocol row envelopes. | Entities host/identity. | high | S-05 removed the manual 15-field row maps. |
| `internal/modules/entities/hostidentity/clipboard_paste_api.go` | Decodes clipboard requests and creates typed tabular-ingest plans. | `ClipboardPasteRequest` and decode/plan/hash helpers. | Workbook entity adapters. | Tabular ingest and HTTP API contracts. | Clipboard API tests and browser flows. | Workbook clipboard request contract. | Entities host/identity. | high | Source owner retains normalization/admission. |
| `internal/modules/entities/hostidentity/clipboard_paste_api_test.go` | Characterizes clipboard decode, plan, hash, and invalid payloads. | Two tests. | Go test discovery. | Clipboard API. | This file. | Workbook clipboard contract. | Entities. | medium | Both tests are selected. |
| `internal/modules/entities/hostidentity/clipboard_paste_store.go` | Applies clipboard plans transactionally and emits row results/revisions/projection effects. | `ClipboardPasteResult`, `ClipboardPasteRowResult`, store method, hash helper. | Workbook entity adapters. | PostgreSQL, records, revisions, ingest, projections. | Backend integration and browser ingest paths. | Row mutation and projection contracts. | Entities host/identity. | high | Mutation side effects must remain explicit. |
| `internal/modules/entities/hostidentity/deleterestore/provider.go` | Supplies host/identity snapshots and source delete/restore operations. | `HostSource`, `IdentitySource` and constructors. | Root revision contribution. | PostgreSQL and revision delete/restore contracts. | Revision and delete/restore suites. | Revision snapshot schemas. | Entities host/identity. | high | Source-owned revision seam. |
| `internal/modules/entities/hostidentity/field_registry.go` | Defines the immutable owner-local Host/Identity field behavior registry and validates it against generated view-schema owners. | Package-private descriptors, registry construction, lookup, row materialization, and typed patch application. | Host/Identity create and patch admission, row builders, conflict admission, and patch stores. | Generated view-schema projection, PostgreSQL transaction type, UUID, owner-local alias behavior. | Registry conformance plus existing admission/integration/browser suites. | Hosts/Identities view schemas. | Entities host/identity. | high | No registration API, service location, duplicated normative metadata, or cross-owner export. |
| `internal/modules/entities/hostidentity/field_registry_test.go` | Proves registry closure, metadata parity, partitions, stable row hashes, and deterministic construction failures. | `TestEntityFieldRegistryMatchesOwnerProjection_Unit`. | Go test discovery. | Field registry and generated view-schema projection. | This file. | Exact `module.entities` selector and generated harness topology. | Entities / Testing Harness. | high | Added and exactly accounted in S-05. |
| `internal/modules/entities/hostidentity/import_create.go` | Implements import-owner create facade for hosts and identities. | `ImportCreateCommand`, `NewImportCreateFacade`, `CreateImportRowTx`. | `importassembly/owner_registry.go`. | Imports owner facade, PostgreSQL, workbook projection. | Import integration tests through assembly. | `contracts/imports/view-targets.v1.json`. | Entities host/identity. | high | Imports remains anti-corruption coordinator. |
| `internal/modules/entities/hostidentity/match.go` | Exact-match precedence, normalized identifiers, aliases, reusable identifiers, and synchronization. | Conflict error, precedence/canonicalization and sync helpers. | Host/identity create/patch/import and merge. | PostgreSQL, records, normalization, UUID. | Exact-match unit/integration/browser tests. | Core entity matching semantics. | Entities host/identity. | high | Large source-semantic file; behavior freeze required. |
| `internal/modules/entities/hostidentity/patch_api.go` | Decodes and hashes entity patch requests through registry-backed field admission. | `DecodePatchRequest`, `PatchRequestHash`. | Workbook patch adapter/store. | Field registry, HTTP API, and JSON helpers. | Patch API and registry tests. | Workbook patch contract. | Entities host/identity. | high | S-05 removed the explicit direct-field allowlist. |
| `internal/modules/entities/hostidentity/patch_api_test.go` | Characterizes patch ordering, hashing, alias actions, and invalid fields. | Three tests. | Go test discovery. | Patch API. | This file. | Workbook patch contract. | Entities. | medium | All tests are selected. |
| `internal/modules/entities/hostidentity/patch_store.go` | Coordinates registry-backed host/identity patches with conflicts, revisions, aliases, and projection effects. | Patch DTOs/errors and `PatchEntityRow`. | Workbook entity adapters. | Field registry, PostgreSQL, authn, records, revisions, projections. | Patch, conflict, registry, integration, and browser tests. | Row mutation and conflict contracts. | Entities host/identity. | high | S-05 removed both per-type field switches while preserving row-version and no-effective-change behavior. |
| `internal/modules/entities/hostidentity/ports.go` | Owner-local adapters to Records and Revisions ports. | Package-private adapter methods. | Host/identity stores. | Records, Revisions, PostgreSQL. | Store and integration tests indirectly. | Change-set/revision contracts. | Entities host/identity. | medium | Typed adapters are intentional. |
| `internal/modules/entities/hostidentity/projectionprovider/host_source.go` | Loads and lists authoritative host inputs for derived projections. | `Source`, `NewSource`, host load/list methods. | Projection contribution assembly. | PostgreSQL plus Records, Links, Evidence source tables. | Projection assembly and entity integration tests. | Host projection-provider descriptor. | Entities source contributor. | high | Cross-source reads are declared by the descriptor. |
| `internal/modules/entities/hostidentity/projectionprovider/identity_source.go` | Loads and lists authoritative identity inputs for derived projections. | Identity load/list methods on `Source`. | Projection contribution assembly. | PostgreSQL plus Records, Links, Evidence source tables. | Projection assembly and entity integration tests. | Identity projection-provider descriptor. | Entities source contributor. | high | Physical projection ownership remains Projections. |
| `internal/modules/entities/hostidentity/reportingprovider/provider.go` | Produces reporting fields and facts from entity projections. | `Provider`, `New`, provider collection methods. | Server reporting contribution. | Reporting export contract and projection reader. | Reporting integration through composition. | Reporting support/fact contracts. | Entities source contributor. | medium | Typed reporting seam. |
| `internal/modules/entities/hostidentity/rollbackprovider/collections.go` | Describes and reverses alias and reusable-identifier collection mutations. | `CollectionProvider` and rollback methods. | Root revision contribution. | PostgreSQL and rollback contract. | Collection rollback tests. | Revision target semantics. | Entities host/identity. | high | Four tests are absent from the focused owner manifest. |
| `internal/modules/entities/hostidentity/rollbackprovider/collections_test.go` | Characterizes collection target identities and owner-defined changed fields. | Four tests. | Go test discovery. | Collection rollback provider. | This file. | Revision target semantics. | Entities / Revisions. | high | All four top-level tests are currently unaccounted. |
| `internal/modules/entities/hostidentity/rollbackprovider/host.go` | Validates and restores host snapshot values. | `HostProvider` and rollback methods. | Root revision contribution. | PostgreSQL and rollback contract. | Host rollback test. | Host revision snapshot schema. | Entities host/identity. | high | Source-owned inverse behavior. |
| `internal/modules/entities/hostidentity/rollbackprovider/host_test.go` | Characterizes host rollback source conversion. | `TestHostSourceForRollbackValue`. | Go test discovery. | Host rollback provider. | This file. | Host revision snapshot schema. | Entities / Revisions. | high | Test is absent from the focused owner manifest. |
| `internal/modules/entities/hostidentity/rollbackprovider/identity.go` | Validates and restores identity snapshot values. | `IdentityProvider` and rollback methods. | Root revision contribution. | PostgreSQL and rollback contract. | Identity rollback tests. | Identity revision snapshot schema. | Entities host/identity. | high | Preserves nullable identifiers. |
| `internal/modules/entities/hostidentity/rollbackprovider/identity_test.go` | Characterizes identity rollback values and invalid state. | Two tests. | Go test discovery. | Identity rollback provider. | This file. | Identity revision snapshot schema. | Entities / Revisions. | high | Both tests are absent from the focused owner manifest. |
| `internal/modules/entities/hostidentity/store.go` | Queries projection rows, hydrates source facts, and creates/updates hosts and identities. | `Store`, constructor/options, query/create/update/load methods. | Workbook, Imports, Timeline assembly, performance fixtures. | PostgreSQL, Records, Revisions, Projections, query/view contracts. | Unit, integration, support, browser, import, and projection tests. | Hosts/Identities view and projection contracts. | Entities host/identity. | high | Central persistence-adjacent application surface. |
| `internal/modules/entities/hostidentity/testsupport/routetest/inventory.go` | Declares control queries for route tests. | `ControlQueries`. | Reusable route-test support. | Hostidentity store/query contracts. | Test-support consumers. | `tools/test_support_inventory.json`. | Entities owner-local test support. | low | Runtime and support eligibility are explicitly inventoried. |
| `internal/modules/entities/hostidentity/timeline_port.go` | Exposes eligible aliases and target validation to Timeline. | `EligibleAlias` and two store methods. | `timelineassembly` and Timeline auto-resolution. | PostgreSQL and entity source tables. | Auto-resolution unit/browser/integration tests. | Timeline/Entities port contract. | Entities host/identity. | high | Entities provides facts; Timeline owns current policy/transaction. |
| `internal/modules/entities/hostidentity/workbook_admission_test.go` | Characterizes entity create and conflict admission/hash compatibility. | Three tests. | Go test discovery. | Hostidentity API and conflict admission. | This file. | Workbook mutation contracts. | Entities / Workbook. | medium | All tests are selected. |
| `internal/modules/entities/hostidentity/workbook_conflict.go` | Resolves Workbook conflicts through source-owned host/identity mutation rules. | `ConflictCommand`, `ResolveWorkbookConflict`. | Workbook entity adapters. | PostgreSQL, revisions, records, projections. | Admission, integration, and browser conflict tests. | Workbook conflict contract. | Entities host/identity. | high | Workbook coordinates; Entities owns source mutation. |
| `internal/modules/entities/hostidentity/workbook_conflict_admission.go` | Decodes and hashes conflict-resolution requests and claims. | Claims/request DTOs plus decode/hash helpers. | Workbook conflict adapter. | HTTP API and conflict-token contracts. | Workbook admission tests. | Workbook conflict request contract. | Entities host/identity. | high | Preserve exact admission and hash semantics. |
| `internal/modules/entities/mentions/collaboration_intents.go` | Builds collaboration intents for mention mutations. | Package-private intent helpers. | Mention lifecycle operations. | Collaboration intent contract. | Mention unit/integration/browser tests. | Collaboration event semantics. | Entities mentions. | medium | Typed cross-owner effect is intentional. |
| `internal/modules/entities/mentions/mention_api.go` | Decodes mention actions and builds mutation/invalidation envelopes. | Mention request/result DTOs and decode/hash helpers. | Root routes and mention store. | HTTP API, UUID, mention effects. | Root route/unit/integration tests. | Mention OpenAPI/protocol contracts. | Entities mentions. | high | Wire hashes and statuses are frozen. |
| `internal/modules/entities/mentions/mention_lifecycle.go` | Coordinates explicit resolve, dismiss, restore, row versions, revisions, links, projections, Timeline effects, and collaboration. | Errors, conflict/transition types, access and lifecycle methods. | Root route and merge/Timeline workflows. | PostgreSQL, authn, Records, Revisions, Links, Timeline, Projections, Collaboration. | Unit, integration, support, and browser tests. | Mention lifecycle and mutation contracts. | Entities mentions. | high | Legitimate source mutation coordinator using ports. |
| `internal/modules/entities/mentions/mention_resolution.go` | Resolves an existing mention to an entity within a transaction. | Error/result and `ResolveExistingFromMentionTx`. | Mention lifecycle and Timeline workflows. | PostgreSQL, Records, Links, projection writer. | Mention resolution unit/integration/browser tests. | Mention resolution behavior. | Entities mentions. | high | Explicit resolution remains Entities-owned. |
| `internal/modules/entities/mentions/merge_repoint.go` | Repoints mentions after explicit entity merges. | Merge command/result/mutation DTOs and repoint method. | Merge store through `MentionStore`. | PostgreSQL and mention source rows. | Merge unit/integration tests. | Merge/revision consequences. | Entities mentions. | high | Called through a typed port. |
| `internal/modules/entities/mentions/ports.go` | Constructs the mention store and defines Links, Timeline, Collaboration, and Projection ports plus API error mapping. | `Store`, options, port types, command DTOs, constructor, API-error helpers. | Timeline assembly, root routes, merge composition. | PostgreSQL, Revisions, Links, Collaboration, Projections, HTTP API. | Port, unit, integration, and browser tests. | Mutation/error contracts. | Entities mentions. | high | Cross-owner dependencies are injected. |
| `internal/modules/entities/mentions/ports_test.go` | Verifies required link-operation injection. | `TestNewStoreRequiresLinkOperations`. | Go test discovery. | Mention store constructor. | This file. | Harness accounting. | Entities. | medium | Test is absent from the focused owner manifest. |
| `internal/modules/entities/mentions/reportingprovider/provider.go` | Produces mention reporting fields and facts. | `CollectFieldsTx`, `CollectFactsTx`. | Server reporting contribution. | PostgreSQL and reporting export contract. | Reporting integration through composition. | Reporting support/fact contracts. | Entities mentions. | medium | Owner-local reporting provider. |
| `internal/modules/entities/mentions/rollbackprovider/mention.go` | Describes and reverses entity-mention mutations. | `MentionProvider` and rollback methods. | Root revision contribution. | PostgreSQL and rollback contract. | Mention rollback tests. | Revision target semantics. | Entities mentions. | high | Source-owned inverse behavior. |
| `internal/modules/entities/mentions/rollbackprovider/mention_test.go` | Characterizes mention target identity and changed-key rules. | Three tests. | Go test discovery. | Mention rollback provider. | This file. | Revision target semantics. | Entities / Revisions. | high | All three tests are absent from the focused owner manifest. |
| `internal/modules/entities/mentions/timeline_history_facts.go` | Supplies mention collection-field history facts to Timeline. | `LoadTimelineCollectionFieldsChangedTx`. | Timeline history composition. | PostgreSQL and entity mention rows. | Timeline history and entity integration tests. | Timeline projection disclosure. | Entities mentions. | high | Entities supplies source facts without owning Timeline projection. |
| `internal/modules/entities/mentions/write_port.go` | Provides transactional mention ordinal and insert operations to Timeline. | `CreateParams`, `NextOrdinalTx`, `InsertTx`. | Timeline mention collection store. | PostgreSQL and entity mention rows. | Timeline auto-resolution and entity integration tests. | Mention source-row contract. | Entities mentions. | high | Source table write remains behind an Entities port. |
| `internal/modules/entities/merge/collaboration_intents.go` | Builds collaboration intents for merge mutations. | Package-private helpers. | Merge store. | Collaboration intent contract. | Merge unit/integration/browser tests. | Collaboration event semantics. | Entities merge. | medium | Typed cross-owner effect is intentional. |
| `internal/modules/entities/merge/hostidentity_boundary.go` | Narrows hostidentity types and operations used by merge. | Type aliases and package-private boundary helpers. | Merge store. | Entities hostidentity. | Merge tests. | Host/identity row semantics. | Entities merge. | medium | Owner-internal dependency, not a cross-module leak. |
| `internal/modules/entities/merge/link_effects.go` | Defines typed link/tag repoint effects. | Link mutation commands/results and `LinkEffectsPort`. | Merge store and Timeline assembly. | Links implementation through injected adapter. | Merge composition/unit/integration tests. | Link consequences of merge. | Entities merge / Links seam. | high | No direct Links storage write. |
| `internal/modules/entities/merge/merge_api.go` | Decodes merge requests, hashes them, and builds response payloads. | Merge request/result DTOs and decode/hash/payload helpers. | Root route and merge store. | HTTP API and Timeline invalidation contract. | Merge unit/integration/OpenAPI/browser tests. | Merge OpenAPI/protocol contract. | Entities merge. | high | Wire behavior is frozen. |
| `internal/modules/entities/merge/merge_protected_set_composition_test.go` | Verifies assessment subjects are included in protected merge records. | One test and port stubs. | Go test discovery. | Merge composition ports. | This file. | Merge safety behavior. | Entities / Assessments seam. | high | Selected by the owner manifest. |
| `internal/modules/entities/merge/merge_protected_set_test.go` | Verifies mention-store injection and rejects unprotected assessment repoints. | Two tests and stubs. | Go test discovery. | Merge store and ports. | This file. | Merge safety behavior. | Entities. | high | Assessment test is selected; injection test is unaccounted. |
| `internal/modules/entities/merge/merge_store.go` | Validates and executes explicit host/identity merge transactions. | Merge errors and route/store methods. | Root merge route. | PostgreSQL, authn, hostidentity, mentions, Records, Revisions, injected effects. | Merge unit/integration/browser tests. | Merge, revision, projection, and collaboration contracts. | Entities merge. | high | Legitimate explicit mutation coordinator. |
| `internal/modules/entities/merge/ports.go` | Adapts Records, Revisions, Assessments, and Mentions to merge-local ports. | Package-private adapters. | Merge store. | Records, Revisions, Assessments, Mentions, PostgreSQL. | Merge unit/composition/integration tests. | Change-set and protected-set behavior. | Entities merge. | high | Boundary tests forbid direct mention/projection SQL. |
| `internal/modules/entities/merge/store.go` | Constructs merge store and declares Timeline, Links, Collaboration, Assessments, Projections, and Mentions dependencies. | `Store`, options, `MentionStore`, `NewStore`. | `timelineassembly` and root routes. | PostgreSQL, Revisions, typed cross-owner ports. | Merge port/unit/integration tests. | Merge effect contract. | Entities merge. | high | All cross-owner effects are injected. |
| `internal/modules/entities/testsupport/entities.go` | Seeds and inspects hosts, identities, aliases, and mentions for tests. | Payload, seed, lookup, assertion, and route-payload helpers. | Backend integration and browser-support suites. | PostgreSQL and entity wire/storage semantics. | Test consumers only. | `tools/test_support_inventory.json`. | Entities owner-local test support. | medium | Test-only assumptions must not enter production. |
| `internal/modules/entities/testsupport/performancefixture/production.go` | Adapts production hostidentity creation to performance fixtures. | `ProductionApplication` and create methods. | Shared performance fixture assembly. | Hostidentity store and authn. | AC-043 fixture suites. | Performance fixture behavior. | Entities owner-local test support. | medium | Core 05 applies only to published claims. |
| `internal/modules/entities/testsupport/performancefixture/provider.go` | Defines entity fixture inputs and contributes them to fixture construction. | Host/Identity DTOs, `Application`, `Provider`. | Shared performance fixture assembly. | Fixture contracts. | AC-043 fixture suites. | Fixture descriptor and receipt. | Entities owner-local test support. | medium | No service-starting entrypoint. |
| `internal/modules/entities/timelinefacts/reader.go` | Reads entity mentions and alias facts for Timeline projection disclosure. | `Reader`, `LoadMentionsTx`. | Timeline projection assembly. | PostgreSQL and Timeline projection fact contract. | Timeline projection and entity integration tests. | Timeline mention/alias disclosure. | Entities source contributor. | high | Facts remain source-owned; Timeline derives presentation. |
| `internal/modules/entities/workbookprojection/contribution.go` | Declares typed host/identity projection ports, inputs, descriptors, and surface intents. | Writer/Rebuilder/Reader/SourceReader interfaces, DTOs, `Contribution`, descriptors and intents. | Projection assembly, hostidentity, mentions, merge, frontend-generated contracts indirectly. | Projection provider contract and view schemas. | Contribution, projection assembly, integration, and browser tests. | `contracts/projection-providers/index.json` and view-contract generation. | Entities source contributor. | high | Projections owns physical storage and query implementation. |
| `internal/modules/entities/workbookprojection/contribution_test.go` | Verifies source requirements and typed descriptor facts. | Two tests. | Go test discovery. | Workbook projection contribution. | This file. | Projection-provider contract. | Entities / Projections. | high | Both tests are absent from the focused owner manifest. |

## 3. Module Boundary Diagnosis

The target is a mixed-responsibility source-owner module with a bounded root
package. The root package MUST remain the Entities transport facade and
source-contribution publication surface. No physical split is authorized by
this tracker. The target MUST NOT become a generic application facade,
frontend controller, grid-vendor integration layer, or cross-owner transaction
coordinator.

The inspected responsibilities do not establish an accidental catch-all. They
establish a legitimate thin HTTP facade, persistence-adjacent host/identity
application behavior, explicit mention and merge mutation coordination,
source-owned contributions, and owner-local test support.

This boundary is a binding plan for later authorized work, but it is not yet an
adopted repository boundary. Before topology implementation begins, the owner
MUST add `docs/decisions/entities-module-boundary.md`, add the corresponding
Core 00 owner/adoption entry, and add a matching Core 04 conformance criterion.
Until all three exist, the current physical topology MUST remain unchanged.

### Bounded-root responsibility whitelist

The root `entities` package MUST contain only the following responsibility
classes:

| Allowed root responsibility | Required interface boundary | Prohibited expansion |
| --- | --- | --- |
| HTTP transport facade | Register the two Entities-owned mutation operations; decode transport input; enforce the current authentication, CSRF, authorization, visibility, error-precedence, and response rules; delegate to owner services. | Workbook row-create routes, clipboard planning, policy engines, persistence algorithms, or application-wide dependency injection. |
| Source-contribution publication | Export typed Entities contributions for Revisions, Recovery, Incident Portability, and subtype presence. | Generic catalog aggregation, coordinator-owned validation, physical projection storage, reporting coordination, or dynamic discovery. |
| Entity portability operations | Export and import Entities-owned incident-bundle rows through the coordinator's typed source port. | Cross-owner bundle writes, transaction ownership, or coordinator policy. |
| Entity subtype declarations | Publish stable host/identity subtype presence and owner-local identifiers. | Owning adjacent Parties, Indicators, Assessments, Evidence, or Links behavior. |
| Owner-internal composition | Construct or expose typed owner-local providers from direct child packages when a source contribution requires them. | Service locators, package `init` registration, mutable global registries, or hidden dependency lookup. |

All root exports MUST belong to one whitelist row. A later export that cannot be
mapped to exactly one row MUST be rejected until the adopted boundary decision
is amended. The root MUST NOT own Timeline automatic-resolution policy, the
Timeline source-row transaction, generic Revisions/Recovery/Portability/
Projections coordination, direct physical projection SQL, frontend state, grid
objects, or cross-owner source mutation.

### Import and construction rules

The permitted dependency direction is:

```text
cmd/* -> internal/app/*assembly -> internal/modules/entities
                                   -> entities/<direct-child>
entities root -> entities/<direct-child> -> owner dependencies through typed ports
coordinating module/app assembly -> an explicitly published Entities child surface
frontend -> generated contracts -> @cartulary/grid-adapter
```

An Entities child package MUST NOT import the Entities root, any
`internal/app/*assembly` package, or a concrete cross-owner coordinator merely
to obtain dependencies. Root and child constructors MUST receive mandatory
dependencies explicitly. They MUST NOT self-register through `init`, a mutable
global registry, a process singleton, or a service locator.

The exported `mentions.TimelineEffectsPort` and
`merge.TimelineEffectsPort`, injected through their respective
`WithTimelineEffects` options, are authorized child-package boundaries. They
MUST remain typed constructor injection surfaces. Boundary guards MUST permit
these interfaces and the root's whitelisted contribution imports while still
prohibiting concrete Timeline-store imports, reverse root imports, and direct
cross-owner writes.

A caller-supplied `pgx.Tx` is borrowed. An Entities operation receiving it MUST
NOT begin, commit, roll back, nest, or detach the transaction. The caller owns
its lifecycle. An operation MAY return owner facts or typed effect inputs; it
MUST NOT publish collaboration events or execute another owner's private SQL.

### Authorized production consumers

The following direct consumers form the bounded import allowlist. A new direct
consumer MUST be justified by, and added to, the adopted Entities boundary
decision and its conformance projection before introduction.

| Entities surface | Authorized direct production consumers | Default rule |
| --- | --- | --- |
| Root `entities` | Server assembly, Revision assembly, Recovery assembly, Incident Portability assembly. | No other direct root consumer is authorized. |
| `hostidentity` | Workbook, Imports, Timeline, and Assessment composition. | Consume typed APIs; do not reach through to owner-private storage. |
| `hostidentity/projectionprovider` | Projection assembly. | Read authoritative source facts only; Projections owns physical derived storage. |
| `hostidentity/reportingprovider` | Server reporting composition. | Publish typed reporting facts only. |
| `mentions` | Timeline and Workbook composition. | Use caller-transaction commands and explicit effect ports. |
| `mentions/reportingprovider` | Reporting module. | Publish typed reporting facts only. |
| `merge` | Timeline/entity-merge composition and the root route facade. | Inject cross-owner effects; do not write other owners' tables. |
| `timelinefacts` | Timeline projection composition. | Return Entities-owned facts; Timeline owns presentation. |
| `workbookprojection` | Projections, Workbook, Imports, and Timeline composition. | Publish typed descriptors and ports; do not own physical projection storage. |
| `entitycontract` | Entities packages only. | Stable owner-local identifiers; no generic shared-kernel expansion. |
| `testsupport` and `hostidentity/testsupport` | Tests and registered test-support composition only. | Production imports are prohibited. |

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Entity HTTP registration and HTTP error/auth ordering | Root `routes.go` and `http_helpers.go` | Entities transport facade with platform auth/HTTP primitives | keep | Server assembly registers exactly two owner routes; boundary tests keep Workbook row routes out. | Thin transport-adjacent surface. |
| Host/identity source semantics and persistence | `hostidentity` | Entities | keep | Core 02 vocabulary; create/query/patch/import/conflict stores. | Persistence-adjacent but source-owned. |
| Alias and reusable-identifier matching | `hostidentity/match.go` | Entities | keep | Exact-match precedence and provenance are entity semantics. | Characterize before internal decomposition. |
| Explicit mention lifecycle | `mentions` | Entities | keep | Entity mention rows, explicit resolve/dismiss/revert, revisions, and source facts. | Cross-owner effects use typed ports. |
| Timeline automatic resolution policy and transaction | `internal/modules/timeline` | Timeline, consuming Entities ports | keep | `auto_resolution.go` and `mentions_collections_store.go` implement policy and transaction. | This ownership split is fixed; no relocation slice is planned. |
| Explicit entity merge coordination | `merge` | Entities | keep | Core entity merge behavior; injected Links, Assessments, Timeline, Projection, Collaboration effects. | Mutation coordinator is legitimate. |
| Projection source contribution | `workbookprojection` and `hostidentity/projectionprovider` | Entities source facts; Projections physical storage | keep | Adopted Projections boundary and provider descriptors. | Do not move physical projection ownership into Entities. |
| Revision, recovery, reporting, and portability contributions | Root/provider subpackages | Entities supplies contributions; coordinating modules aggregate | keep | App assembly and adopted Revisions boundary. | Source owners construct providers. |
| Workbook/import adapters and coordination | `internal/app/workbookassembly` and `importassembly` | Workbook and Imports | keep | App composition adapts Entities facades; generic modules do not own source meaning. | No direct move into target proposed. |
| Owner-local test fixtures | `testsupport` and `hostidentity/testsupport` | Entities test support | keep | `tools/test_support_inventory.json`. | Runtime exclusion is explicit. |
| Frontend shell/controller state | `apps/web/src/workbook` and `features/entities` | Web/Workbook frontend | keep | Entity surface/query/merge controllers are outside the backend target. | Contract impact is frozen, not relocated. |
| Grid vendor integration | `packages/grid-adapter` | Grid adapter | keep | App entity surface imports `@cartulary/grid-adapter`, not `react-data-grid`. | No violation found. |
| Root transport and source-contribution publication | Root `entities` package | Bounded Entities root | keep | Routes coexist with source-owned revision, recovery, and portability publications. | Responsibilities MUST remain within the whitelist; no physical split is authorized. |

## 4. Public Contract and Behavior Freeze Map

Unless a later task is explicitly authorized to change behavior, every contract
in this section MUST remain observationally equivalent. Equivalent means the
same route and operation identity, request and response envelope, status and
error precedence, authorization outcome, transaction result, row shape,
revision/change-set effect, projection refresh, collaboration event, saved-view
behavior, generated surface, and harness accounting.

### HTTP operation invariants

Both operations MUST authenticate as state-changing requests before path
parsing, authorization, or body decoding. This preserves session and CSRF
handling owned by the HTTP-auth platform. A successful mutation MUST perform
session sliding only after the domain operation succeeds, then return the
existing success envelope. Errors MUST use the established error envelope and
the precedence below.

| Route and operation ID | Authorization and error precedence | Success |
| --- | --- | --- |
| `POST /api/v1/records/{survivor_record_id}/merge`; `mergeEntityRecord` | (1) platform authentication/CSRF; (2) malformed or non-visible survivor as `404 incident_not_found`; (3) incident visibility before role disclosure; (4) require `reviewer` or `admin`, otherwise `403 authorization_denied` with `required_role=reviewer\|admin`; (5) decode body, hiding invalid `loser_record_id` as `404 incident_not_found` and returning existing `400 invalid_mutation_payload` for other invalid input; (6) map, in current precedence, client-transaction conflict, closed incident, hidden target, merge precondition, row-version conflict, record lock, then internal failure. The 409 codes MUST remain `client_txn_conflict`, `incident_closed`, `merge_precondition_failed`, `row_version_conflict`, and `record_locked` as applicable. | Existing `200` `RecordMergeEnvelope`; idempotent replay, payload ordering, changed records, revisions, projections, and ordinary collaboration behavior MUST remain unchanged. |
| `POST /api/v1/entity-mentions/{entity_mention_id}/resolve`; `resolveEntityMention` | (1) platform authentication/CSRF; (2) malformed, missing, or non-visible mention as `404 entity_mention_not_found`; (3) require incident role `editor`, `reviewer`, or `admin`, otherwise `403 authorization_denied` with `required_role=editor\|reviewer\|admin`; (4) decode body as existing `400 invalid_mutation_payload`; (5) map, in current precedence, client-transaction conflict, closed incident, missing mention, non-visible target, row-version conflict, illegal transition, invalid resolved target, deleted source, then internal failure. Invalid resolved target MUST be `400 invalid_mutation_payload` with `field=resolved_record_id` and `reason_code=invalid_value`; other established codes MUST remain `client_txn_conflict`, `incident_closed`, `entity_mention_not_found`, `resolved_record_not_found`, `row_version_conflict`, `illegal_transition`, and `record_deleted_use_restore`. | Existing `200` `MentionActionEnvelope`; lifecycle, idempotent replay, links, revision/change set, source row version, projection refresh, and ordinary collaboration behavior MUST remain unchanged. |

### Timeline-to-Entities automatic-resolution port

Timeline MUST retain automatic-resolution eligibility, suppression and
ambiguity policy, source mutation, transaction lifecycle, Links/History/
Projection/Collaboration coordination, disclosure, and Undo. Entities MUST
provide alias facts, incident/type/lifecycle target validation, mention
persistence, and Entities-owned history facts through operations borrowing the
caller's transaction.

| Operation | Inputs | Output | Mandatory error/result obligations |
| --- | --- | --- | --- |
| `EntityPort.ListEligibleAliasesTx` | Context, borrowed `pgx.Tx`, incident UUID, entity type. | Ordered `[]EntityAlias`, each containing record UUID and raw alias text. | MUST scope to the incident and requested `host` or `identity` type; MUST include only `stub` or `canonical` lifecycle states; an unsupported type MUST return an empty result and no invented match. |
| `EntityPort.ValidateResolvedTargetTx` | Context, borrowed `pgx.Tx`, incident UUID, entity type, record UUID. | No payload on success. | MUST reject unsupported type, cross-incident target, missing target, inactive target, and wrong-type target through the established not-found/validation error contract. |
| `MentionPort.ResolveExistingFromMentionTx` | Context, borrowed `pgx.Tx`, actor, source record UUID, source field key, mention UUID, resolved record UUID, timestamp. | Typed record-link mutation facts plus error. | MUST validate mention/source identity and target; MUST apply only the established mention transition; MUST return typed facts needed by the caller. |
| `MentionPort.ApplyMentionLifecycleTx` | Context, borrowed `pgx.Tx`, actor, source record UUID, field key, mention UUID, action, optional resolved record UUID, timestamp. | Typed record-link mutation facts plus error. | MUST enforce the existing lifecycle transition and target guards; MUST NOT convert invalid actions into no-ops. |
| `MentionPort.NextOrdinalTx` | Context, borrowed `pgx.Tx`, source record UUID, source field key. | Next integer ordinal. | MUST derive the next ordinal in the caller's transaction and MUST propagate storage failure. |
| `MentionPort.InsertTx` | Context, borrowed `pgx.Tx`, typed mention-create parameters. | Error only. | MUST insert exactly the supplied source/type/origin/text/status/ordinal/actor/time/resolution facts and MUST propagate storage failure. |
| `MentionPort.LoadTimelineCollectionFieldsChangedTx` | Context, borrowed `pgx.Tx`, source record UUID, change timestamp. | Ordered `[]string` of owner-defined changed field keys. | MUST return Entities-owned mention history facts for Timeline revision composition and MUST propagate storage failure. |

Every operation in this port MUST NOT begin, commit, roll back, nest, or detach
the borrowed transaction. It MUST NOT publish events, execute private SQL for a
different source owner, return HTTP response objects, or return frontend/grid
objects. Error values MUST remain machine-classifiable by existing callers; an
implementation MUST NOT replace a classified error with an untyped string.

The existing injected Timeline-effect surfaces MUST remain command/fact ports:

| Surface | Required methods and result boundary | Prohibitions |
| --- | --- | --- |
| `mentions.TimelineEffectsPort` | `PrepareMentionActionTx` returns typed pre-action state; `ApplyMentionActionEffectsTx` applies a typed command and returns a typed action result in the borrowed transaction. | Entities MUST NOT reconstruct Timeline source rows, disclosure, or projection policy privately. |
| `merge.TimelineEffectsPort` | Load timeline invalidations, load relationship invalidations, and refresh named Timeline projection rows in the borrowed transaction. | Entities MUST NOT import a concrete Timeline store or publish Timeline events directly. |

### Host and Identity field partitions

Each materializer MUST emit exactly 15 semantic cells. `record_id` and
`row_version` MUST remain top-level technical row members and MUST NOT be
counted as cells. A field not listed in a partition below is unsupported.

| Record type | Patchable semantic cell keys: exactly 8 | Non-patchable semantic cell keys: exactly 7 | Create-only seed cell keys |
| --- | --- | --- | --- |
| Host | `host.display_name`, `host.hostname`, `host.aliases`, `host.location`, `host.os_platform`, `host.business_owner`, `host.criticality`, `host.containment_status` | `host.aad_device_id`, `host.fqdn`, `host.reusable_identifiers`, `host.host_state`, `host.linked_event_count`, `host.evidence_count`, `host.edited_at` | `host.aad_device_id`, `host.fqdn` |
| Identity | `identity.display_name`, `identity.upn`, `identity.email`, `identity.sam_account_name`, `identity.aliases`, `identity.privilege_level`, `identity.mfa_state`, `identity.reset_status` | `identity.aad_object_id`, `identity.sid`, `identity.reusable_identifiers`, `identity.identity_state`, `identity.linked_event_count`, `identity.evidence_count`, `identity.edited_at` | `identity.aad_object_id`, `identity.sid` |

Patch omission and null behavior MUST use these closed defaults:

| Patch field class | Members | Omission | Explicit `null` or normalized-empty input |
| --- | --- | --- | --- |
| Non-clearable direct text | `host.display_name`, `host.hostname`, `identity.display_name`, `identity.upn`, `identity.email`, `identity.sam_account_name` | Leaves the field unchanged. | MUST return the established invalid-payload/nullability result; MUST NOT clear the source value. |
| Clearable direct text | `host.location`, `host.os_platform`, `host.business_owner`, `host.criticality`, `host.containment_status`, `identity.privilege_level`, `identity.mfa_state`, `identity.reset_status` | Leaves the field unchanged. | Explicit JSON `null` clears to source `null`; other input that fails the bound normalization MUST return the established invalid-payload result. |
| Alias collections | `host.aliases`, `identity.aliases` | Leaves the collection unchanged. | Raw JSON `null`, raw arrays, and direct values are invalid; patch MUST use non-empty `collection_actions_v1` with the existing bounded action grammar. |
| Non-patchable or unknown | The 7-field sets above, `record_id`, `row_version`, and undeclared keys. | No request change exists. | Admission MUST return `unsupported_field_key` before value interpretation or effects. |

Create-only identifiers MUST remain accepted by the existing create path and
visible in query rows. Ordinary patch admission MUST reject them. In the
authored machine view-schema input, each create-only identifier MUST project as
`writable=false`, `create_writable=true`, `grid_editable=false`, and
`write_kind=direct_value`. A declared non-patchable key MUST return
`400 invalid_mutation_payload` with its canonical key in `error.details.field`
and `reason_code=unsupported_field_key`; an unknown key MUST use
`error.details.field=field_key` with the same reason code. Rejection MUST occur
before any source state, row version, change set, history, projection,
idempotency-success, or collaboration effect. The refactor MUST NOT introduce
a route, schema ID, field key, data migration, or compatibility shim.

One Entities-owned executable field descriptor registry MUST become the sole
source used by Host and Identity materialization and patch admission. Each
descriptor MUST define the following values; omission is prohibited.

| Descriptor member | Allowed/default semantics |
| --- | --- |
| `view_schema_id` and `field_key` | MUST use Core-owned identifiers exactly; no aliases or inferred keys. |
| materializer and cell kind | MUST deterministically emit the existing value and cell representation; every declared field MUST materialize once. |
| default visibility | MUST preserve the authored view-schema value. |
| patch write kind | Exactly one of `none`, `direct_value`, or `collection_action`; default is `none`. |
| create seed | Boolean; default is `false`; `true` does not imply ordinary patchability. |
| null/clear policy | MUST explicitly distinguish omission, explicit null, empty collection, and permitted clear behavior. |
| conflict class | MUST preserve existing conflict and hash behavior for the field. |
| entity binding mode | MUST identify the existing source/projection binding; no inferred cross-owner binding. |

Frontend code MUST consume generated contracts derived from authored owner
inputs. It MUST NOT import this backend registry.

### Automatic-resolution characterization matrix

Before implementation changes to either side of the port, the following cases
MUST be executable. Each failure case MUST prove that the Timeline source
mutation and every participating mention, link, history, projection, and
collaboration effect are atomic.

| Case | Required stimulus | Required observation |
| --- | --- | --- |
| Exact success | One eligible alias equals the JSON-decoded raw token after `mention_token_text_v1` normalization and locale-independent Unicode case folding, for an eligible signed-in interactive workbook writer. | Source text is unchanged; one resolved mention names the chosen record; one active link has `provenance='auto_match'` and `confidence=100`; history, projection, disclosure, and ordinary collaboration outcomes match existing contracts. |
| No match | No eligible alias exactly matches under the required comparison. | The mention is created or preserved as `resolution_status='unresolved'` with null resolved record; no active resolved link, auto-match provenance, or auto-resolution disclosure exists. |
| Ambiguity | More than one eligible target matches exactly. | The unresolved path is retained and no arbitrary target or active resolved link is chosen. |
| Suppressors and forbidden rewrites | Exercise ASCII `?` and `~` anywhere; each whole whitespace-delimited token `maybe`, `prob`, `probably`, `approx`, and `approximately`; and matches requiring punctuation/parenthetical stripping, punctuation collapse, token deletion, transliteration, stemming, fuzzy matching, or locale-specific lexicons. | The ordinary unresolved result is retained; ranking MAY remain non-mutating, but no active resolved link, auto-match provenance, or disclosure exists. |
| Ineligible writer or workflow | File import, machine-applied create/patch/paste, background job, service-to-service writer, imported/machine replay, explicit inspector resolve, alias-edit cell, merge/dedupe, asynchronous enrichment/cleanup, or implicit canonical-record creation. | Automatic resolution does not run; the workflow's otherwise-owned source behavior remains unchanged. |
| Target rejection | Cross-incident, inactive, missing, unsupported-type, and wrong-type targets. | Existing classified rejection occurs before mutation effects. |
| Participant failure | Inject failure independently at mention, link, history, projection, and collaboration pre-commit participants. | The complete transaction rolls back; no participant's durable effect remains. |
| Exact replay | Same idempotency identity and same request hash. | Existing response/result is replayed without duplicate effects. |
| Conflict | Same idempotency identity with a different hash or current conflicting version. | Existing conflict classification and precedence are preserved. |
| Batch ordering | Batch paste with multiple automatic resolutions and stable ties. | Disclosure includes the visible change-set count and deterministic item order; no unstable map/database iteration leaks into output. |
| Disclosure and Undo | Resolve, inspect disclosure, activate Review, execute immediate Undo, and execute later Revert to unresolved after disclosure expiry. | Disclosure contains raw token, canonical target, matched alias, direct Undo, and direct Review; Review does not dismiss it; Undo restores the raw unresolved token, removes the auto-created link, and preserves focus/scroll; later revert remains reachable in at most two actions and creates a new attributed revision. |
| Authorization and lifecycle | Unauthorized actor and illegal mention/source lifecycle transitions. | Authorization and lifecycle rejection occur in the existing order with zero side effects. |

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `POST /api/v1/records/{survivor_record_id}/merge` | Entities | Root route, merge API/store, authored OpenAPI. | Merge unit, integration, OpenAPI, browser. | Preserve auth/CSRF, hidden-not-found, role, conflict, lock, idempotency, and payload ordering. | high | Route shape and envelopes are frozen. |
| `POST /api/v1/entity-mentions/{entity_mention_id}/resolve` | Entities | Root route, mention API/lifecycle, authored OpenAPI. | Mention unit, integration, support, browser. | Preserve legal transitions, role checks, conflict ordering, target validation, and restore rules. | high | Route supports explicit action semantics. |
| Host and identity query rows | Entities source / Workbook query / Projections storage | Hostidentity store, contribution descriptors, view schemas. | Support integration, frontend unit, browser provenance. | Prove the exact key set, 15/15 sentinel values, null/omission behavior, technical-member separation, and projection parity for each record type. | high | Projection IDs and query metadata are observable. |
| Host and identity create/patch/clipboard behavior | Entities source / Workbook coordination | Hostidentity APIs/stores and Workbook adapters. | Admission, integration, support, browser. | Prove 8/8 positive patch admission, 7/7 negative admission, create-only seed acceptance, hash stability, and zero effects on rejection. | high | Preserve hashes, idempotency, provenance, and conflicts. |
| Exact-match, aliases, and reusable identifiers | Entities | `hostidentity/match.go` and Core 02. | Exact-match unit, alias migration, integration, browser auto-resolution. | Preserve precedence and normalization fixtures before decomposition. | high | Do not infer behavior from labels alone. |
| Mention source rows and lifecycle | Entities | `mentions` package and source table operations. | Unit, integration, support, browser lifecycle. | Account port and rollback tests in focused harness rows. | high | Raw text, status, ordinals, and history are frozen. |
| Automatic mention resolution | Timeline policy / Entities facts | Timeline auto-resolution/store plus Entities alias and mention ports. | Browser auto-resolution and backend integration. | Implement every row of the automatic-resolution characterization matrix above. | high | Ownership is fixed at the typed port; relocation is not a planned slice. |
| Explicit entity merge and side effects | Entities with injected owner effects | `merge` package and Timeline assembly. | Unit, integration, composition, browser. | Account injection test; preserve protected-set and effect ordering. | high | No direct mention/projection SQL is allowed. |
| Saved-view and view-schema behavior | Workbook/views / authored view contracts | Hosts/Identities view-schema JSON and generated view contracts. | Support, frontend unit, browser grid/provenance. | Prove exact field keys, default visibility, write attributes, saved-view compatibility, and generated projection parity. | high | No saved-view implementation resides in the target. |
| Projection refresh/rebuild/query | Projections physical owner; Entities source contributor | Adopted projection boundary and provider descriptors. | Projection assembly, support integration, browser. | Preserve source descriptor facts and refresh/delete effects. | high | Generated projections must not be hand-edited. |
| Authorization and CSRF | Platform auth plus Entities route policy | `routes.go` and Core 04. | Integration/support authorization and CSRF tests. | Preserve authenticate-before-body and role re-derivation order. | high | Error precedence is part of the contract. |
| Revisions and change sets | Revisions coordinator / Entities source providers | Root contribution and rollback/delete-restore providers. | Contribution test plus broader revision suites. | Account ten rollback/provider tests before moving code. | high | Snapshot and target identity semantics are owner-defined. |
| Incident portability and recovery | Coordinating modules / Entities source contributions | Portability and recovery contribution files/catalogs. | Collaborating portability and recovery suites. | Map exact owner rows before changing contribution paths. | high | Source semantics remain Entities-owned. |
| Import facade | Imports coordinator / Entities create semantics | Import owner registry and hostidentity facade. | Import assembly integration tests. | Preserve target IDs, provenance, unmapped-column policy, and transaction ownership. | high | `module.entities@1` is a public authored target. |
| Reporting provider | Reporting coordinator / Entities fact provider | Hostidentity and mention reporting providers. | Reporting integration through server composition. | Characterize provider-key and field/fact output if reorganized. | medium | Typed contribution seam. |
| WebSocket behavior | Collaboration/runtime transport | Support integration and frontend query subscription. | `TestProjectionAndWebsocketConsequences` and browser flows. | Preserve ordinary `record_changed` invalidation; do not invent an entity event family. | high | No mention-specific WebSocket family was found. |
| Frontend entity surface and selectors | Web/Workbook frontend | Entity surface, query, model, merge controller, mention adapters. | Vitest, Playwright, accessibility, visual. | Run affected browser layers after contract-facing changes. | high | Grid vendor access remains behind the adapter. |
| Generated protocol/view/UI contracts | Authored contract owners and generators | OpenAPI, view schemas, projection providers, generated roots. | Drift, shape, frontend type/unit/browser checks. | No new characterization beyond owner-input parity unless a contract changes. | high | Update owner inputs, then run Make generation. |
| Harness accounting and evidence mapping | Testing Harness | Owner verification contract and test-family manifest. | Harness contract/explain targets. | Add the six exact-selector rows in Section 5 and an owner-local reconciliation subtest; prove a 50/50 baseline with all anomaly counts zero. | high | Evidence routing does not define runtime architecture. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Focused owner manifest selects 29 of 50 target-local Go test functions. | Exact comparison of Go `Test*` declarations with `tools/test_families/module.entities.json`. | Boundary and rollback regressions can be omitted from focused slices. | `must_fix` | Testing Harness with Entities ownership input. | Reconcile rows before production movement; regenerate schedules through Make if required. |
| Seven static boundary guards are absent from the focused owner manifest. | `boundary_guard_test.go` versus manifest selectors. | A refactor can violate imports or direct-write boundaries without focused evidence. | `must_fix` | Entities / Testing Harness. | Add explicit owner accounting for all seven guards. |
| Four collection rollback, one host rollback, two identity rollback, three mention rollback, one mention port, one merge injection, and two projection contribution tests are unaccounted. | Test declarations versus manifest selectors. | Source history, dependency injection, and projection boundaries can regress. | `must_fix` | Entities / Testing Harness. | Add the six exact-selector rows specified below; alternate implicit reachability is not sufficient. |
| Host/identity schemas, row builders, and patch allowlists duplicate field knowledge. | Two 15-field view schemas, `BuildHostRow`/`BuildIdentityRow`, patch admission switches. | Silent field drift across query and mutation surfaces. | `must_fix` | Entities with View Contracts. | Characterize the exact partitions, then introduce the executable descriptor registry and correct authored projections. |
| Root package contains HTTP registration and source-contribution publications. | Root file inventory and app assembly imports. | Unbounded additions could obscure ownership or introduce invalid import direction. | `must_fix` | Bounded Entities root. | Adopt the whitelist, import matrix, decision record, Core 00 owner entry, and Core 04 conformance criterion before topology work. |
| Timeline owns automatic-resolution policy and the source transaction while Entities supplies facts and caller-transaction operations. | Timeline `auto_resolution.go`/`mentions_collections_store.go` and Entities ports. | Moving ownership could change transaction, event, Undo, or disclosure semantics. | `intentional/no_action` | Timeline policy; Entities source facts. | Preserve the split, enforce the port prohibitions, and add the complete characterization matrix. |
| Merge and mention side effects cross module boundaries through typed ports. | Store options and Timeline assembly injection. | Low if ports remain stable; high if bypassed. | `intentional/no_action` | Respective source owners. | Preserve ports and boundary tests. |
| Projection source readers access Records, Links, and Evidence facts declared by descriptors. | Projection provider source files and `contracts/projection-providers/index.json`. | A physical move could conflate source and derived storage. | `intentional/no_action` | Entities source contribution / Projections storage. | Preserve the adopted contribution boundary. |
| Owner-local stores use PostgreSQL directly for entity-owned source rows. | Hostidentity, mentions, merge, rollback, and portability code. | Replacing it without a proven seam can add indirection or change transaction scope. | `intentional/no_action` | Entities. | Keep source persistence local; review only concrete cross-owner writes. |
| Platform auth/HTTP imports are confined to transport/application-facing code. | Root routes/helpers and API decode/error helpers. | Low under current structure. | `intentional/no_action` | Platform primitives / Entities route policy. | Preserve the direction; do not move domain semantics into platform. |
| Generated contract projections are downstream of authored owners. | Generated artifact policy and contract trees. | Hand edits would drift or be overwritten. | `must_fix` | Contract owners and generators. | Never hand-edit; use owner inputs and Make-owned generation. |
| Testsupport is explicitly owner-local and runtime-excluded. | `tools/test_support_inventory.json` and package imports. | Test assumptions could leak if reused by production. | `intentional/no_action` | Entities test support. | Preserve runtime exclusion. |
| Frontend entity code accesses the grid vendor only through the adapter. | Entity workbook surface imports `@cartulary/grid-adapter`; vendor imports reside in `packages/grid-adapter`. | Low under current boundary. | `intentional/no_action` | Grid adapter. | Preserve frontend import-boundary checks. |

### Exact harness-accounting closure

The authored harness input MUST add six compatible unit rows for the following
21 symbols. Every row MUST use primary owner `module.entities`, the ordinary Go
unit runner, no service runtime, standard resources, and no fixture unless the
adopted Testing Harness owner requires a stricter value. Collaborators MUST be
limited to those shown. Row identifiers MUST be produced by the harness's
public semantic-ID rule; this tracker does not authorize a parallel naming
scheme.

| Compatible row purpose | Package | Exact top-level selectors | Collaborators |
| --- | --- | --- | --- |
| Root boundary guards | `./internal/modules/entities` | `TestEntitiesProductionImportBoundaries`, `TestEntitiesDoNotRegisterWorkbookRowCreateRoutes`, `TestEntitiesDoNotBuildClipboardPastePlans`, `TestEntitiesRoutesUseCollaborationPublisher`, `TestEntitiesRootDoesNotImportHostIdentity`, `TestMergeDoesNotWriteMentionOrProjectionTablesDirectly`, `TestMentionsUseCommandLevelTimelineEffectsPort` | none |
| Host/identity rollback | `./internal/modules/entities/hostidentity/rollbackprovider` | `TestParseCollectionTargetAcceptsCanonicalOwnerIdentities`, `TestParseCollectionTargetRejectsMalformedOrMismatchedIdentity`, `TestParseCollectionTargetAcceptsPhysicalAliasDeleteIdentity`, `TestCollectionChangedFieldsRemainOwnerDefined`, `TestHostSourceForRollbackValue`, `TestIdentitySourceForRollbackValuePreservesNullableIdentifiers`, `TestIdentityProviderRejectsMalformedStateAndMergeID` | `module.revisions` |
| Mention rollback | `./internal/modules/entities/mentions/rollbackprovider` | `TestParseMentionTargetRejectsMalformedIdentity`, `TestParseMentionTargetAcceptsRetainedMention`, `TestMentionChangedKeysAreOwnerDefined` | `module.revisions` |
| Mention-store construction | `./internal/modules/entities/mentions` | `TestNewStoreRequiresLinkOperations` | `module.links` |
| Merge dependency injection | `./internal/modules/entities/merge` | `TestWithMentionStoreRetainsInjectedInstance` | none |
| Projection contribution | `./internal/modules/entities/workbookprojection` | `TestNewContributionRequiresSource`, `TestRuntimeContributionOwnsTypedEntityDescriptorFacts` | `module.projections` |

Every selector MUST be an exact top-level Go test symbol. Prefix, substring,
glob, and regular-expression selectors are prohibited. After reconciliation,
the ordinary build-context baseline MUST report 50 discovered target-local
top-level tests and 50 selected exactly once, with zero missing, duplicate,
overlapping, stale, or unresolved selectors.

`TestEntitiesProductionImportBoundaries` MUST gain an Entities-owned
reconciliation subtest or helper that AST-discovers active top-level Go tests
under `internal/modules/entities` and compares them with active exact Go
selectors. It MUST NOT add another top-level test. Build-tag-specific tests, if
introduced, MUST be reconciled in their explicit harness profile rather than
silently added to the ordinary-build baseline.

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Establish authority, planning scope, baseline, and the only allowed tracker write. | This tracker; owner documents read-only. | `make lint-markdown` for tracker revisions. | Authority order, normative terms, source posture, and later-authorization rule are recorded. |
| WF-01 | Target inventory | chain | WF-00 | WF-02 | Account for the 68-file baseline and every authorized new file, including the two S-05 registry files, with callers, dependencies, tests, and contracts. | `internal/modules/entities/**` and composition callers. | Exact live-path-to-inventory comparison. | Each live target path appears exactly once and no stale path appears. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-04 | Freeze every observable contract and assign its current/adopted owner and test posture. | Core owners, OpenAPI, view/projection/import/revision contracts, routes, generated projections. | Read-only owner, route, and contract comparison. | Every discovered contract has an owner, evidence, required characterization, and risk. |
| WF-03 | Characterization and field-parity plan | parallel | WF-02 | WF-05 | Define executable automatic-resolution, transaction-atomicity, and Host/Identity field matrices. | Entity and Timeline tests, view schemas, materializers, patch admission, frontend contract consumers. | Proposed focused, service-backed, frontend, and browser gates. | Every matrix row has a deterministic stimulus and observation; 15/15, 8/8, and 7/7 counts are explicit. |
| WF-04 | Boundary and coupling scan | parallel | WF-02 | WF-05 | Classify dependencies and define the bounded-root import and responsibility rules. | Entities tree, app assembly, Timeline, generated-contract and frontend adapter boundaries. | Existing guards plus `make backend-module-boundary-check`. | Every finding has one allowed classification, one owner, and one required action. |
| WF-05 | Bounded-root owner adoption | chain | WF-03, WF-04 | WF-06 | Adopt the no-split bounded-root decision and its owner/conformance projections. | New Entities boundary decision, Core 00 owner/adoption entry, Core 04 conformance criterion. | Owner review and `make lint-markdown`; executable enforcement follows in S-02. | All three owner artifacts adopt the same whitelist, consumer matrix, import DAG, borrowed-transaction rule, and no-registration rule. |
| WF-06 | Behavior-preserving slice sequencing | chain | WF-05 | WF-07 | Execute the ordered characterization, registry, projection, and conformance slices without relocating Timeline behavior. | Files named by the authorized slices in Section 7. | Narrow gates named per slice. | Each slice passes its binary criterion before the next slice begins. |
| WF-07 | Exact harness accounting | chain | WF-06 | WF-08 | Add the six exact-selector rows, account for all new tests, and regenerate derived harness outputs. | Authored test-family/catalog inputs and Entities tests; generated outputs only through Make. | `make test-catalog-check`; `make harness-contract`; focused owner slices. | Every active target-local top-level test has exactly one compatible selector and all anomaly counts are zero. |
| WF-08 | Validation and final handoff | chain | WF-07 | none | Run all affected narrow gates, generated checks, finalization, and the full check; record retained evidence. | No additional implementation surface. | Section 8 gate sequence. | Every required gate passes or is reported with its result artifact; handoff and rollback state are current. |

## 7. Authorized Refactor Slice Plan

The slices MUST execute in the listed order; a slice MUST NOT start until every
dependency meets its completion criterion, this tracker is updated, and the
updated tracker passes Markdown lint. The authorized outcome preserves owner
contracts except for the explicit create-only patch correction required by Core
01. Timeline automatic-resolution policy and transaction ownership MUST NOT
move in any slice. Any new top-level test MUST receive one compatible exact
selector in the same slice rather than being deferred to final accounting.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-01 | none | Adopt the bounded-root whitelist, prohibited responsibilities, consumer matrix, import DAG, borrowed-transaction rule, and no-global-registration rule in the required Entities decision, Core 00 entry, and Core 04 criterion. Rebaseline S-04 as the contract-first field cutover and make selector accounting continuous. No package move is authorized. | `docs/decisions/entities-module-boundary.md`, Core 00, Core 04, and this tracker. | Ambiguous ownership or inconsistent owner text could authorize incompatible topology. | Preserve current boundary guards pending their S-02 rebase. | `make lint-markdown`. | Revert the owner edits together before S-02; retain the current physical package. | The three owner artifacts state one identical bounded-root model, this tracker records the executable sequence, and Markdown lint passes. |
| S-02 | S-01 | Rebase root and child-package boundary guards onto the adopted model; add the six exact-selector unit rows for all 21 omitted tests; add the AST reconciliation subtest. | `internal/modules/entities/boundary_guard_test.go`, existing target tests, authored verification/test-family/catalog inputs; generated schedules only through Make. | An overly broad guard could reject valid contribution imports; an inexact selector could hide or duplicate evidence. | Preserve all 50 baseline tests; add no new top-level test for reconciliation. | `make test-catalog-check`; `make backend-module-boundary-check`; `make test-slice OWNER=module.entities`; `make harness-contract`. | Revert guard and authored harness inputs together, regenerate derived outputs, and retain the prior selectors. | Baseline is 50 discovered and 50 selected exactly once; missing, duplicate, overlap, stale, and unresolved counts are all zero. |
| S-03 | S-02 | Add the complete automatic-resolution matrix from Section 4 without moving policy or transaction ownership. Add deterministic failure injection at every transaction participant. | Timeline automatic-resolution tests and test ports; Entities hostidentity/mention port tests; authorized test-support only. | Transactions, idempotency, authorization/error order, disclosure, Undo, WebSocket invalidation, and batch order. | Success, no match, ambiguity, suppressors, ineligible writers, invalid targets, each participant failure, replay, conflict, batch ordering, disclosure, Undo, authorization, and lifecycle rejection. | `make test-slice OWNER=module.entities`; affected Timeline owner slice discovered through its task guide; `make service-backed-test-slice OWNER=module.entities`. | Remove only the new test cases/test doubles; do not change production behavior to make an invalid premise pass. | Every matrix row passes, and each injected participant failure proves zero durable partial effects. |
| S-04 | S-03 | Correct the four create-only fields in the authored Host/Identity view schemas, regenerate downstream projections through Make, and add exact field/effect characterization. | Authored Host/Identity view schemas, generated projections through Make, Host/Identity tests, and affected frontend contract tests. | Intentional correction of legacy patch admission, row envelopes, saved views, conflicts, idempotency hashes, projections, history, and collaboration. | For each record type: 15/15 cells, 8/8 accepted patch fields, 7/7 rejected non-patchable fields, 2 create-only seeds, generated editability, and zero effects on rejection. | Generation drift/policy/shape checks; focused/service-backed Entities; affected frontend and browser gates. | Do not restore the out-of-contract patch behavior after release; repair forward from authored owner inputs. | Authoritative and generated projections agree, every specified count and effect assertion passes, and the four create-only identifiers reject ordinary patch before effects. |
| S-05 | S-04 | Introduce one executable Entities field descriptor registry and make materialization and ordinary patch admission consume it; preserve Core-owned IDs and values. Correct implementation behavior only where required by adopted owners. | `internal/modules/entities/hostidentity` and owner-local tests; no generated output is edited directly. | Patch admission, cell values, omission/null semantics, hashes, row versions, history, projections, collaboration, and saved views. | Preserve all S-04 characterization and existing create/query/patch/clipboard/conflict suites. | Focused and service-backed Entities slices; affected frontend unit/typecheck/import-boundary/browser gates. | Keep the prior explicit maps/switches until registry-backed behavior passes; revert the slice atomically if parity fails. | Both materializers use the registry and prove 15/15 cells; patch admission proves 8/8 accepted and 7/7 rejected fields; any behavior correction was explicitly authorized. |
| S-06 | S-05 | Reconcile every live test and generated projection, run the complete affected validation sequence, and publish the final implementation handoff. | Verification inputs, retained run evidence, and this tracker; no new behavior surface. | Harness completeness, frontend consumers, generated drift, broad regressions, and retained evidence. | Preserve all prior matrices, OpenAPI operation identities, route/security behavior, WebSocket semantics, saved views, and grid-adapter boundary. | Section 8 gate sequence, ending with `make agent-finalize` and `make check`. | Prefer a forward fix; any pre-merge revert removes a complete slice and regenerates derived outputs from restored owner inputs. | Every active test has one compatible selector, all required gates pass, and the handoff records retained evidence and residual risk. |

## 8. Validation Plan

These commands are public Make targets discovered from the repository. In this
table, `yes` means the gate is mandatory before the affected implementation
slice is complete; it does not authorize implementation. A command marked
`as affected` MUST run when the slice changes or can change that surface. The
tracker-only revision runs only the Markdown gate.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| harness catalog | `make test-catalog-check` | Authored test catalog, exact selectors, compatibility, ownership, and stale/unresolved accounting. | yes | S-02 passed with the 50/50 baseline and zero selector anomalies; rerun after each new top-level test. |
| backend boundary | `make backend-module-boundary-check` | Bounded root, child imports, source-contribution imports, and prohibited dependency directions. | yes | MUST run after S-01, S-02, and any later import change. |
| unit | `make test-slice OWNER=module.entities` | Focused Entities owner rows, including all compatible non-service tests. | yes | The result is not a complete owner signal until exact accounting closes. |
| integration | `make service-backed-test-slice OWNER=module.entities` | Service-backed Entities rows for storage, routes, revisions, projections, and transactional effects. | yes | MUST cover atomic participant failures and persistence-facing field behavior. |
| frontend | `make frontend-unit`; `make frontend-typecheck`; `make frontend-import-boundary-check` | Generated-contract consumers, row/view behavior, types, and grid-vendor boundary. | no | All three become mandatory when a slice changes a generated frontend contract or its consumer behavior. |
| e2e/browser | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful`; `make browser-e2e-a11y`; `make browser-e2e-visual` | Contract-facing entity, mention, saved-view, inspector, authorization, and grid workflows. | no | Every affected layer becomes mandatory; visual updates MUST follow the visual-golden guide. |
| generated drift | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check` | Authored inputs, generated outputs, generated-file policy, and JSON shapes. | yes | Mandatory after S-06; generated files MUST NOT be hand-edited. Run `make openapi-compatibility-check` if an authorized OpenAPI owner input changes. |
| harness contract | `make harness-contract` | Harness schemas, generated topology, schedules, and execution contracts. | yes | Mandatory after authored harness input or selector changes. |
| finalization | `make agent-finalize` | End-of-run harness maintenance and retained-result preparation. | yes | Run before the full gate; record whether `RESULTS_DIR` was supplied. |
| full check | `make check` | Repository developer verification gate. | yes | Final implementation acceptance requires a passing result or an explicitly reported failure artifact; failure is not success. |
| tracker checkpoint | `make lint-markdown` | This authored Markdown tracker and all touched Markdown owner artifacts. | yes | Run after every S-01 through S-06 tracker update and before the next slice. |

Command discovery performed:

- `make help`
- `make help-all`
- `make task-guide ROLE=module-author OWNER=module.entities`
- `make explain-test-owner OWNER=module.entities`
- `make explain-target TARGET=test-slice DETAIL=summary`
- `make explain-target TARGET=test-catalog-check DETAIL=summary`
- `make explain-target TARGET=backend-module-boundary-check DETAIL=summary`
- `make explain-target TARGET=frontend-import-boundary-check DETAIL=summary`
- `make explain-target TARGET=generate-drift DETAIL=summary`
- `make explain-target TARGET=harness-contract DETAIL=summary`
- `make explain-target TARGET=check DETAIL=summary`

Read-only `rg`, `find`, `sed`, `jq`, `git status`, and `git rev-parse` commands supported file, symbol, selector, caller, and baseline discovery.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| ENT-001 | Establish scope, authority, normative terms, and planning-only posture. | WF-00 | DONE | none | Section 1. | Only the tracker is changed and later authorization is explicit. |
| ENT-002 | Inventory every baseline and authorized new target file exactly once. | WF-01 | DONE | ENT-001 | Section 2 and live path comparison; S-05 added two explicit rows for a current total of 70. | The live/inventory path sets are equal and duplicate-free. |
| ENT-003 | Map owners and freeze observable contracts. | WF-02 | DONE | ENT-002 | Sections 3 and 4. | Every discovered contract has an owner, evidence, tests, required characterization, and risk. |
| ENT-004 | Define the bounded-root topology, Timeline/Entities port, exact field partitions, and coupling actions. | WF-03, WF-04 | DONE | ENT-003 | Sections 3 through 5. | Interfaces, defaults, prohibitions, mappings, and binary observations are unambiguous. |
| ENT-005 | Adopt the bounded-root owner decision and matching Core ownership/conformance entries. | WF-05 | DONE | ENT-004 | S-01; Markdown run `.cartulary/test-results/20260822T024626Z-p411947`. | Decision, Core 00, and Core 04 state the same no-split model; executable enforcement follows in S-02. |
| ENT-006 | Rebase boundary guards and account for all 21 omitted tests with six exact-selector rows. | WF-06, WF-07 | DONE | ENT-005 | S-02; focused run `.cartulary/test-results/20260822T025618Z-p484473`; harness run `.cartulary/test-results/20260822T025805Z-p537665`. | Baseline is 50/50 with zero missing, duplicate, overlap, stale, or unresolved selectors. |
| ENT-007 | Add the complete automatic-resolution and atomic-failure matrix. | WF-03, WF-06 | DONE | ENT-006 | S-03; Timeline focused `.cartulary/test-results/20260822T031925Z-p663519`, Timeline service-backed `.cartulary/test-results/20260822T032400Z-p720139`, Entities focused `.cartulary/test-results/20260822T032835Z-p776534`, and Entities service-backed `.cartulary/test-results/20260822T033020Z-p829396`. | Every policy row passes; all seven post-write participant failures leave source, mention, link, history, both projections, collaboration, and idempotency state unchanged. |
| ENT-008 | Move Timeline automatic-resolution behavior into Entities. | WF-06 | DROPPED | none | Fixed ownership split in Sections 3 and 4. | This work item remains absent from the refactor; a future owner change requires a separate authority process. |
| ENT-009 | Add Host/Identity field and side-effect characterization. | WF-03, WF-06 | DONE | ENT-007 | S-04; Entities focused `.cartulary/test-results/20260822T034519Z-p923575` and service-backed `.cartulary/test-results/20260822T034706Z-p978317`. | Each type proves 15/15 cells, 8/8 accepted patches, 7/7 rejected fields, and create-only behavior. |
| ENT-010 | Introduce the Entities field descriptor registry. | WF-06 | DONE | ENT-009, ENT-011 | S-05; registry row `.cartulary/test-results/20260822T041419Z-p1386558`, focused `.cartulary/test-results/20260822T041112Z-p1273817`, and service-backed `.cartulary/test-results/20260822T041112Z-p1273824`. | Materialization, create/patch admission, conflict admission, and patch application use one registry without contract drift. |
| ENT-011 | Correct authored machine projections and regenerate derived outputs. | WF-06 | DONE | ENT-007 | S-04; generation `.cartulary/test-results/20260822T033602Z-p893686` and drift `.cartulary/test-results/20260822T034506Z-p919817`. | Create-only attributes are exact and generated outputs derive only from authored inputs. |
| ENT-012 | Account for every newly added top-level test continuously and reconcile the final inventory. | WF-07 | DONE | ENT-006 | S-06 static and AST-backed reconciliation passed 51/51 at `.cartulary/test-results/20260822T041739Z-p1392676`; final catalog and harness roots are recorded in Section 13. | Each active ordinary-build target-local top-level test has one compatible exact selector. |
| ENT-013 | Run final validation and publish the implementation handoff. | WF-08 | DONE | ENT-011, ENT-012 | Section 13; final `make check` passed 642/642 at `.cartulary/test-results/20260822T044053Z-p1929849`. | Every required gate passes and command/result evidence is recorded. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-21T21:08:52-04:00 | Codex planning session | Scope and authority mapped; planning-only constraint active. | Inspected owner documents and framework; touched only this tracker. | `sed`, `rg`, `make help`, `make help-all`. | No owner contradiction found; framework treated as evidence only. | None for tracker creation. | Preserve authority order in every later slice. |
| 2026-08-21T21:54:12-04:00 | Codex NLSpec-style tracker revision | Normative terms, defaults, adoption limits, and later-authorization rules are explicit. | Inspected the tracker, NLSpec writing guide, advisory notes, Core owners, domain guide, and boundary ADRs; touched only this tracker. | `sed`, `rg`, `git status`, `apply_patch`, `make lint-markdown`. | Advisory recommendations were converted into testable planning requirements without elevating the tracker above adopted owners; Markdown lint passed. | Implementation is not authorized. | Begin S-01 only in a later authorized task. |
| 2026-08-21T22:44:05-04:00 | Codex S-01 implementation | S-01 complete; contract-first sequence recorded. | Added the Entities boundary decision; updated Core 00, Core 04, and this tracker. | `rg`, `sed`, `git status`, `apply_patch`, `make lint-markdown`. | REQ-00-073 and AC-558 adopt the same bounded-root model; Markdown passed at `.cartulary/test-results/20260822T024626Z-p411947`. | Executable boundary enforcement remains deliberately sequenced in S-02. | Begin S-02 boundary and harness reconciliation. |
| 2026-08-22T00:45:28-04:00 | Codex S-06 implementation | All authorized slices and final developer gates are complete. | Reconciled final repository state and updated this tracker; `docs/domain.md` was inspected but not edited because vocabulary is unchanged. | The ordered S-06 sequence in Section 13, ending with `make check` and `make lint-markdown`. | No owner contradiction, generated hand edit, unaccounted test, or incomplete mandatory gate remains. | None. | Use Section 13 to audit or reproduce the result; release certification remains a separate task. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-21T21:08:52-04:00 | Codex planning session | All 68 target files mapped; target diagnosed as a mixed source-owner module with legitimate facades, stores, coordinators, and contributions. | Inspected `internal/modules/entities/**` and relevant app/Timeline composition; touched only this tracker. | `find`, `rg`, `sed`. | No accidental catch-all or concrete misplaced owner logic proven. | RB-001, RB-003. | Reconcile evidence, then seek an owner-approved physical topology. |
| 2026-08-21T21:54:12-04:00 | Codex NLSpec-style tracker revision | Bounded-root no-split model, whitelist, prohibited responsibilities, import DAG, consumer allowlist, borrowed-transaction rule, and no-registration rule are specified. | Inspected all Entities packages and current server, Revision, Recovery, Portability, Workbook, Imports, Timeline, Projection, Assessment, and Reporting composition; touched only this tracker. | `rg`, `sed`, `find`, `apply_patch`. | RB-001 and RB-003 are closed at decision level; production topology remains unchanged pending adoption. | Decision/Core adoption and implementation authorization are pending. | Execute S-01, then rebase guards in S-02. |
| 2026-08-21T22:44:05-04:00 | Codex S-01 implementation | Bounded-root decision adopted without moving packages. | Added `docs/decisions/entities-module-boundary.md`; updated Core 00 REQ-00-073 and owner matrix plus Core 04 AC-558/Base manifest. | `rg`, `sed`, `apply_patch`, `make lint-markdown`. | Root responsibilities, consumers, import direction, borrowed transactions, Timeline split, and no-registration rule agree; Markdown passed at `.cartulary/test-results/20260822T024626Z-p411947`. | S-02 machine enforcement remains pending by design. | Rebase executable guards and policy in S-02. |
| 2026-08-21T22:58:42-04:00 | Codex S-02 implementation | Adopted boundary is executable and the focused baseline is complete. | Updated `tools/backend_module_boundaries.json` and `internal/modules/entities/boundary_guard_test.go`. | `make format`, `make generate`, `make generate-drift`, `make backend-module-boundary-check`, `make test-slice OWNER=module.entities`. | Boundary passed at `.cartulary/test-results/20260822T025350Z-p429622`; focused slice passed at `.cartulary/test-results/20260822T025618Z-p484473`. Initial failures identified and repaired overbroad Timeline-contract matching and a stale port-name assertion. | None. | Begin S-03 without moving Timeline policy or transaction ownership. |
| 2026-08-22T00:14:39-04:00 | Codex S-05 implementation | One static owner-local registry now owns field materialization and patch strategy/application. | Added `hostidentity/field_registry.go`; migrated `api.go`, `patch_api.go`, `patch_store.go`, and `workbook_conflict_admission.go`; removed the two row maps, direct-field allowlist, alias-field helper, and per-type patch switches. | `make format`, `make backend-module-boundary-check`, focused and service-backed Entities slices. | Boundary passed at `.cartulary/test-results/20260822T041037Z-p1272446`; Entities passed 38/38 at `.cartulary/test-results/20260822T041112Z-p1273817` and 29/29 service-backed at `.cartulary/test-results/20260822T041112Z-p1273824`. | None. | Run the ordered S-06 final gates. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-21T21:08:52-04:00 | Codex planning session | Entity surface, query, merge, mention, selector, and grid boundaries mapped as contract consumers outside the backend target. | Inspected entity/mention Workbook frontend and `packages/grid-adapter`; touched only this tracker. | `rg`, `sed`, `make explain-target TARGET=frontend-import-boundary-check DETAIL=summary`. | No direct grid-vendor coupling outside the adapter found. | None unless a later slice changes frontend contracts. | Preserve query, row continuity, inspector, selector, accessibility, and visual behavior. |
| 2026-08-21T21:54:12-04:00 | Codex NLSpec-style tracker revision | Frontend remains a generated-contract consumer and grid access remains adapter-only. | Inspected entity/mention frontend surfaces, selectors, saved-view consumers, and `packages/grid-adapter`; touched only this tracker. | `rg`, `sed`, targeted Make target discovery, `apply_patch`. | No backend registry import or direct grid-vendor dependency is authorized. | Frontend gates are required only for affected later slices. | Run unit, typecheck, import-boundary, and affected browser gates after contract-facing work. |
| 2026-08-21T23:56:49-04:00 | Codex S-04 implementation | Frontend discovery consumes the corrected generated field metadata. | Extended `apps/web/src/workbook/models/entityWorkbookModel.test.ts`; generated TypeScript changed only through `make generate`. | `make frontend-unit`, `make frontend-typecheck`, `make frontend-import-boundary-check`, and four affected browser targets. | The four stable field IDs advertise `createWritable=true`, `gridEditable=false`, and `writeKind=direct_value`; all frontend and browser gates passed, including accessibility and visual with no golden change. | Clients that submitted ordinary patches to these four fields receive the canonical unsupported-field response and must use sanctioned create/upsert workflows. | Preserve these generated-contract assertions through S-05. |
| 2026-08-22T00:14:39-04:00 | Codex S-05 implementation | The backend-only registry remains owner-local and does not create a frontend or grid dependency. | No additional frontend source or generated contract changed in S-05. | `make frontend-unit`, `make frontend-typecheck`, `make frontend-import-boundary-check`. | Frontend passed 390/390 at `.cartulary/test-results/20260822T041310Z-p1381167`; typecheck and import boundary passed at `.cartulary/test-results/20260822T041310Z-p1381163` and `.cartulary/test-results/20260822T041310Z-p1381181`. | None. | Rerun the same checks in S-06. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-21T21:08:52-04:00 | Codex planning session | OpenAPI, view, projection, import, revision, recovery, portability, inspector, and generated surfaces mapped read-only. | Inspected authored and generated contract trees; touched only this tracker. | `rg`, `sed`, `make explain-target TARGET=generate-drift DETAIL=summary`. | Owner inputs and generated projections distinguished; no generated edit proposed. | RB-004 before field-map consolidation. | Change owner inputs only under later authority, then use Make generation and drift checks. |
| 2026-08-21T21:54:12-04:00 | Codex NLSpec-style tracker revision | Exact Host/Identity field partitions, create-only projection attributes, registry members, and generation rules are specified. | Inspected authored view schemas, OpenAPI and provider inputs, generated projections, materializers, and patch admission; touched only this tracker. | `rg`, `sed`, `jq`, targeted Make target discovery, `apply_patch`. | RB-004 is closed at decision level; generated outputs remain read-only and no route, schema ID, field key, migration, or shim is planned. | Characterization and later owner-input implementation remain unauthorized. | Execute S-04 through S-06 in order after prerequisites pass. |
| 2026-08-21T23:56:49-04:00 | Codex S-04 implementation | The four create-only fields now project their Core-01 asymmetry and generated consumers are synchronized. | Updated the authored Host and Identity view schemas; regenerated `internal/gen/contractviewschemas/artifacts_gen.go` and `packages/view-contracts/src/generated/view-contract-projection.ts` only through `make generate`. | `make generate`, `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check`. | Generation passed at `.cartulary/test-results/20260822T033602Z-p893686`; drift, policy, and shape passed at `.cartulary/test-results/20260822T034506Z-p919817`, `.cartulary/test-results/20260822T034506Z-p919847`, and `.cartulary/test-results/20260822T034506Z-p919842`. | None. | Begin S-05 registry work without duplicating the normative projection. |
| 2026-08-22T00:14:39-04:00 | Codex S-05 implementation | Registry construction binds implementation descriptors to generated owner fields and rejects closure or strategy drift deterministically. | Added the registry and its conformance test; added one authored exact test row; regenerated only the harness render index through Make. | `make generate`, `make generate-drift`, `make test-catalog-check`, `make harness-contract`. | Generation passed at `.cartulary/test-results/20260822T040924Z-p1264326`; drift passed at `.cartulary/test-results/20260822T041419Z-p1386516`; catalog passed at `.cartulary/test-results/20260822T041037Z-p1272584`; harness passed at `.cartulary/test-results/20260822T041053Z-p1273195`. The first generation attempt `.cartulary/test-results/20260822T040838Z-p1260977` stopped on non-ASCII-sorted row order; the authored row was moved to canonical order before any dependent gate ran. | None. | Reconfirm all generation and harness policies in S-06. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-21T21:08:52-04:00 | Codex planning session | Target has 50 Go tests; focused owner manifest selects 29 and omits 21. | Inspected all target tests, owner verification, test-family manifest, and support inventory; touched only this tracker. | `rg`, `jq`, `make task-guide ROLE=module-author OWNER=module.entities`, `make explain-test-owner OWNER=module.entities`. | Exact omitted categories recorded; no test suite run. | RB-002. | Complete S-01 before production movement. |
| 2026-08-21T21:54:12-04:00 | Codex NLSpec-style tracker revision | All 21 omitted symbols map to six compatible exact-selector unit rows; the 50/50 closure invariant and reconciliation subtest are specified. | Inspected target tests, authored owner manifest, verification contracts, test catalog, and test-support inventory; touched only this tracker. | `rg`, `jq`, `make task-guide ROLE=module-author OWNER=module.entities`, `make explain-test-owner OWNER=module.entities`, `apply_patch`. | RB-002 is closed at decision level; no harness input or test was changed. | Harness reconciliation requires later authorization. | Execute S-02 and rerun catalog/harness gates after every new test. |
| 2026-08-21T22:58:42-04:00 | Codex S-02 implementation | Six exact rows account for all omitted tests; reconciliation is self-enforcing. | Updated `tools/test_families/module.entities.json`, the boundary test, and generated `tools/execution_topology_render_index.json` through `make generate`. | `make author-test-row-id`, `make generate`, `make generate-drift`, `make test-catalog-check`, `make test-slice OWNER=module.entities`, `make harness-contract`. | AST and static counts both report 50 discovered and 50 selected; catalog passed, generation passed at `.cartulary/test-results/20260822T025236Z-p422374`, drift passed at `.cartulary/test-results/20260822T025254Z-p425366`, focused passed at `.cartulary/test-results/20260822T025618Z-p484473`, and harness passed at `.cartulary/test-results/20260822T025805Z-p537665`. | Continuous accounting remains active for the planned S-05 test. | Add S-03 characterization only under already selected Timeline top-level tests. |
| 2026-08-21T23:33:03-04:00 | Codex S-03 implementation | Automatic-resolution policy and atomicity evidence are complete without moving policy, transaction ownership, or production behavior. | Updated `internal/modules/timeline/request_test.go`, `resolution_integration_test.go`, and the test-only facade composer in `timeline_event_integration_test.go`. | `make format`; focused Timeline integration and unit rows; focused and service-backed owner slices for `module.timeline` and `module.entities`; `make lint-markdown`. | Exact integration passed at `.cartulary/test-results/20260822T031836Z-p648119`; unit passed at `.cartulary/test-results/20260822T033303Z-p890395`; full Timeline passed 51/51 at `.cartulary/test-results/20260822T031925Z-p663519`; Timeline service-backed passed 29/29 at `.cartulary/test-results/20260822T032400Z-p720139`; Entities focused passed 38/38 at `.cartulary/test-results/20260822T032835Z-p776534`; Entities service-backed passed 29/29 at `.cartulary/test-results/20260822T033020Z-p829396`; Markdown passed at `.cartulary/test-results/20260822T033347Z-p891048`. Two initial test-only assertion defects were diagnosed from `.cartulary/test-results/20260822T030955Z-p551932` and `.cartulary/test-results/20260822T031728Z-p628927` and corrected without production changes. | No owner contradiction or product defect. Compatibility impact is none. Rollback removes only the new subtests, transaction decorators, and generalized test composer; no production rollback is needed. Residual risk is limited to the intentional create-only cutover not yet begun. | Begin S-04 by changing authored Host/Identity view schemas, then regenerate through Make. |
| 2026-08-21T23:56:49-04:00 | Codex S-04 implementation | Exact field partitions, create visibility, patch rejection precedence, and side-effect behavior are executable. | Extended existing Host/Identity admission, patch API, integration, and frontend contract tests; no new top-level Go test was added. | `make format`; exact admission and create-route rows; focused and service-backed Entities slices; frontend unit, typecheck, and import-boundary checks; affected webserver-backed, stateful, accessibility, and visual browser suites; `make lint-markdown`. | Entities passed 38/38 at `.cartulary/test-results/20260822T034519Z-p923575` and 29/29 service-backed at `.cartulary/test-results/20260822T034706Z-p978317`; frontend passed 390/390 at `.cartulary/test-results/20260822T034853Z-p1031287`, with typecheck `.cartulary/test-results/20260822T034853Z-p1031284` and import boundary `.cartulary/test-results/20260822T034853Z-p1031303`; browser runs passed at `.cartulary/test-results/20260822T035033Z-p1069360`, `.cartulary/test-results/20260822T035033Z-p1069359`, `.cartulary/test-results/20260822T035450Z-p1164691`, and `.cartulary/test-results/20260822T035450Z-p1164605`; Markdown passed at `.cartulary/test-results/20260822T035826Z-p1247313`. No visual golden changed. | None. The only compatibility change is rejection of ordinary patches to the four canonical create-only keys; existing records, history, create/upsert, query visibility, routes, IDs, and envelopes remain valid. | Replace duplicate field behavior with the owner-local registry in S-05. |
| 2026-08-22T00:14:39-04:00 | Codex S-05 implementation | Registry conformance and exact-selector accounting are complete. | Added `hostidentity/field_registry_test.go`; added its one exact authored selector and regenerated harness topology through Make. | `make author-test-row-id`, exact registry row, static selector recount, `make test-catalog-check`, `make harness-contract`, full focused/service-backed Entities slices. | Registry row passed at `.cartulary/test-results/20260822T041419Z-p1386558`; construction rejects duplicates, omissions, unknown keys, and patch-strategy mismatch; Host/Identity each prove 15 descriptors, 7 direct, 1 collection, 7 none, 8/7/2 partitions, exact owner metadata, grouping parity, and stable row hashes. Static and AST-backed accounting are 51 discovered and 51 selected with no duplicate selector. | None. The first exact test run `.cartulary/test-results/20260822T040936Z-p1267251` captured the two intended row snapshot hashes; the asserted snapshots then passed. | Repeat 51/51 reconciliation as S-06 gate 1. |
| 2026-08-22T00:45:28-04:00 | Codex S-06 implementation | Final exact accounting and all repository gates are complete. | No new production or test behavior was added in S-06; only two lint-message strings and this handoff changed. | Final reconciliation, catalog, harness, boundary, generated checks, four owner slices, three frontend gates, four browser gates, finalization, narrow Go lint, full check, and Markdown lint. | Final accounting is 51/51. The first full check `.cartulary/test-results/20260822T043505Z-p1799078` failed 641/642 only on two related ST1005 messages; the narrow lint fix passed and the complete rerun passed 642/642 at `.cartulary/test-results/20260822T044053Z-p1929849`. | None. | Handoff is complete. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-21T21:08:52-04:00 | Codex planning session | Route authentication, CSRF, incident visibility, role re-derivation, and error precedence frozen. | Inspected root routes/helpers, Core 04, and integration/support tests; touched only this tracker. | `sed`, `rg`. | Existing tests cover primary orderings; behavior changes are excluded. | None for planning. | Preserve ordering in characterization and every later slice. |
| 2026-08-21T21:54:12-04:00 | Codex NLSpec-style tracker revision | Security ordering, target validation, lifecycle rejection, transaction atomicity, replay, conflict, disclosure, and Undo observations are mandatory. | Inspected routes/helpers, Core 03/Core 04, Timeline auto-resolution, Entities validation/lifecycle ports, and relevant tests; touched only this tracker. | `sed`, `rg`, `apply_patch`. | The characterization matrix prohibits partial effects and preserves existing authorization/error order. | Failure-injection evidence does not yet exist for every participant. | Add and pass S-03 characterization before registry implementation. |
| 2026-08-21T23:33:03-04:00 | Codex S-03 implementation | Security-sensitive automatic resolution is characterized at its public and borrowed-transaction boundaries. | Extended the existing Timeline unit/integration top-level tests; no new top-level test or production failpoint was added. | Focused Timeline unit/integration rows and all four S-03 owner gates. | Viewer authorization is evaluated before superseded-row lifecycle disclosure; authorized lifecycle rejection is `illegal_transition`; invalid target variants use one concealed not-found result; exact replay is zero-write, divergent replay conflicts, and Undo restores the raw unresolved token. | None. | Preserve these orderings through S-04 and S-05. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-21T21:08:52-04:00 | Codex planning session | Planning inventory and freeze map are complete; implementation is unauthorized. | Inspected repository evidence listed in Sections 1 and 2; touched only this tracker. | Read-only discovery commands listed in Section 8. | Worktree baseline was clean `main` at `76fef24e`; no validation success claimed yet. | RB-001 through RB-004. | Authorize and execute S-01 in a later task, then update these handoffs. |
| 2026-08-21T21:54:12-04:00 | Codex NLSpec-style tracker revision | All four prior design questions have normative dispositions; six ordered slices and binary completion gates are ready for handoff. | Inspected evidence listed in Sections 1 through 5; touched only this tracker. | Read-only discovery and tracker verification commands listed in Section 8; `apply_patch`. | No production, test, contract, generated, Core, ADR, harness, or migration file was changed. | Later implementation authority and S-01 owner adoption remain required. | Authorize S-01; do not start a dependent slice early. |
| 2026-08-21T23:56:49-04:00 | Codex S-04 implementation | Contract-first create-only cutover and characterization are complete; S-05 is eligible. | Changed the two authored view schemas, regenerated their Go/TypeScript projections through Make, and extended existing backend/frontend tests. | All S-04 generation, Entities, frontend, browser, and Markdown gates recorded in this handoff. | The correction ships backend and generated web contracts together. No database migration, backfill, route version, alias, feature flag, dual write, or legacy patch shim exists. | Registry consolidation and final validation remain. Rollback posture is to repair the authored contract forward, not restore legacy patch admission; residual risk is duplicate field logic until S-05 completes. | Implement and validate the immutable owner-local field registry. |
| 2026-08-22T00:14:39-04:00 | Codex S-05 implementation | Registry consolidation is complete and S-06 is eligible. | Added two inventoried owner-local files, migrated four consumers, removed superseded field logic, and added one exact harness row plus generated render-index update. | All S-05 format, generation, catalog, harness, boundary, owner-slice, frontend, and Markdown gates recorded in this handoff. | No observable behavior change beyond the S-04 correction; no data or client migration. Markdown passed at `.cartulary/test-results/20260822T041639Z-p1390583`. Rollback is atomic removal of the registry slice and restoration of the former maps/switches, but forward repair is preferred because dual implementations are prohibited. | Final broad gates remain; no owner contradiction or unresolved design blocker. | Run S-06 in its mandated order and record every result root. |
| 2026-08-22T00:45:28-04:00 | Codex S-06 implementation | Final validation and handoff are complete. | Reconciled all changed files, current inventory, generated provenance, compatibility, rollback, and evidence roots in Section 13. | Ordered final gate sequence plus final Markdown lint. | The only transient final failure was the related ST1005 finding recorded above; it was fixed and the entire check reran green. | No known unresolved product, owner, security, harness, or migration risk. Monitor existing `unsupported_field_key` telemetry for the four create-only keys without logging submitted values. | A later task may run `make release-check`; it is outside this remediation's acceptance scope. |

## 11. Open Questions and Blockers

No unresolved design question remains in RB-001 through RB-004. The table
retains the stable IDs and distinguishes a closed planning decision from the
authority or executable evidence still required before implementation can
complete.

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Bounded-root adoption gate. | The no-split topology MUST NOT be treated as adopted until its owner and conformance artifacts agree. | Matching decision, Core 00 REQ-00-073, and Core 04 AC-558, followed by executable S-02 enforcement. | CLOSED; owner adoption completed in S-01 and executable enforcement passed in S-02. |
| RB-002 | Exact harness-accounting implementation gate. | Focused verification remains incomplete until all 21 symbols and every new test have one compatible exact selector. | Authored harness inputs, AST reconciliation, and passing catalog/harness evidence. | CLOSED; final S-06 accounting is 51/51 with zero duplicate selector identity. |
| RB-003 | Timeline/Entities automatic-resolution ownership. | An ownership ambiguity would invalidate transaction and effect planning. | No further authority for this refactor: Timeline retains policy and transaction ownership; Entities retains the typed fact/write boundary. | CLOSED; relocation DROPPED. |
| RB-004 | Field-partition characterization and conformance gate. | Registry implementation is unsafe until exact cells, patch partitions, create-only behavior, and side effects are executable. | Passing 15/15, 8/8, 7/7, create-only, hashing, projection, history, and collaboration assertions. | CLOSED; S-04 characterization and contract projection gates passed. |

## 12. Binary Completion Criteria

| Criterion | Result | Evidence |
| --- | --- | --- |
| Every file in `internal/modules/entities` is inventoried or explicitly out of scope. | PASS | Section 2 accounts for the 68-file baseline plus both S-05 registry files, for 70 current files. |
| Every discovered public contract risk has an owner and test posture. | PASS | Section 4. |
| Every proposed workflow has dependencies and exit criteria. | PASS | Section 6. |
| Every authorized implementation slice is behavior-preserving except for the explicit Core-01 create-only patch correction. | PASS | Section 7 isolates that contract-first correction in S-04 and prohibits compatibility shims. |
| Validation commands are discovered or marked `TODO` with a reason. | PASS | Section 8. |
| Contradictions are marked `BLOCKED: owner contradiction`. | PASS | No contradiction was found; Section 1 records the required future treatment. |
| Advisory material does not displace repository authority. | PASS | Section 1 defines its non-authoritative role; normative requirements are stated directly. |
| Handoff sections are current enough for another agent to continue without rediscovery. | PASS | Sections 9 through 11 record state, evidence, blockers, and next actions. |
| The original planning revision touched only this tracker and passed Markdown validation. | PASS | The planning baseline passed `make lint-markdown` with run root `.cartulary/test-results/20260822T022406Z-p399653`; the authorized implementation now has independent slice checkpoints. |

### Later refactor acceptance

The authorized refactor is complete only when every row below is `PASS` with a
recorded command or artifact. `TODO`, skipped, partial, and inferred results are
not acceptance.

| Refactor criterion | Required evidence | Current result |
| --- | --- | --- |
| The bounded-root whitelist, prohibited responsibilities, import DAG, consumer allowlist, borrowed-transaction rule, and no-registration rule are adopted and enforced. | Matching Entities boundary decision, Core 00 entry, Core 04 criterion, and passing backend boundary gate. | PASS; S-01 adoption plus S-02 boundary run `.cartulary/test-results/20260822T025350Z-p429622`. |
| Every active target-local top-level Go test has exactly one compatible exact selector. | Passing test-catalog and reconciliation evidence with discovered equal to selected, and zero missing, duplicate, overlap, stale, and unresolved selectors. | PASS; final static and AST-backed count is 51/51 at `.cartulary/test-results/20260822T041739Z-p1392676`, with catalog and harness passing. |
| Automatic-resolution transaction participants fail atomically. | Every Section 4 participant-failure case passes with zero durable partial source, mention, link, history, both projection, collaboration, or idempotency effects. | PASS; S-03 Timeline focused `.cartulary/test-results/20260822T031925Z-p663519` and service-backed `.cartulary/test-results/20260822T032400Z-p720139`, plus Entities focused `.cartulary/test-results/20260822T032835Z-p776534` and service-backed `.cartulary/test-results/20260822T033020Z-p829396`. |
| Host and Identity materialization and patch partitions are exact. | For each type, passing 15/15 materialized-cell assertions, 8/8 accepted patch assertions, and 7/7 rejected non-patchable assertions. | PASS; S-04 admission and focused/service-backed Entities evidence is recorded under ENT-009. |
| Create-only identifiers preserve their required asymmetric behavior. | Host `aad_device_id`/`fqdn` and Identity `aad_object_id`/`sid` are accepted on create and visible in rows; ordinary patch returns `unsupported_field_key` before every side effect. | PASS; existing create/upsert and patch integration tests prove visibility and canonical zero-effect rejection. |
| Generated outputs derive only from authored inputs. | Owner-input diff plus passing generation drift, generated-artifact policy, JSON-shape, and harness-contract gates; no hand edit in generated roots. | PASS; final harness, drift, generated-policy, and JSON-shape roots are recorded in Section 13. |
| Observable contracts otherwise remain unchanged. | Passing focused, service-backed, affected frontend/browser, finalization, and full gates covering HTTP routes and OpenAPI operation IDs, authorization/error ordering, WebSocket semantics, saved views, row envelopes, projections, revisions, recovery, portability, imports, reporting, and the grid-adapter boundary. | PASS; every final owner, frontend, browser, finalization, and 642-unit full gate passed. |

The authorized implementation changes only the named S-01 through S-06
surfaces. It does not authorize a package split, migration, Timeline ownership
move, OpenAPI identity change, compatibility shim, or release certification.

## 13. Final Implementation Handoff

### Planning and implementation outcome

S-01 through S-06 completed in the required contract-first order. The adopted
Entities boundary remains a bounded source-owner root without a package split.
Timeline retains automatic-resolution policy and transaction ownership;
Entities retains its typed fact and write ports. No owner contradiction was
found.

The substantive implementation outcomes are:

- Host `aad_device_id` and `fqdn`, plus Identity `aad_object_id` and `sid`, are
  create-only in the authored and generated view contracts. They remain
  accepted on create/upsert and visible in query rows, but ordinary patch and
  conflict resolution reject them as `unsupported_field_key` before value
  interpretation or durable effects.
- One package-private static `hostidentity` descriptor registry now binds all
  15 Host and 15 Identity fields to generated owner metadata. Row
  materialization, create admission, patch decoding, conflict admission, and
  patch application consume it. The former row maps, direct-field allowlist,
  alias-field helper, and per-type patch switches were removed.
- Timeline automatic-resolution policy, replay, disclosure, Undo, lifecycle,
  batch ordering, and all seven post-write rollback participants now have
  complete unit and transaction-backed evidence. No production policy or
  transaction ownership moved.
- Final ordinary-build Entities accounting is 51 discovered, 51 selected, and
  51 unique exact selector identities. The owner-local AST guard makes later
  omissions fail the focused gate.

Domain vocabulary did not change. `docs/domain.md` was inspected as the
vocabulary/navigation owner and was not edited.

### Files changed

- Specification and handoff: `docs/decisions/entities-module-boundary.md`,
  `docs/spec/00_document_set_status_and_precedence.md`,
  `docs/spec/04_security_deployment_and_conformance.md`, and this tracker.
- Authored contracts: the Host and Identity JSON inputs under
  `contracts/view-schemas/`.
- Make-generated projections: `internal/gen/contractviewschemas/artifacts_gen.go`,
  `packages/view-contracts/src/generated/view-contract-projection.ts`, and
  `tools/execution_topology_render_index.json`. None was hand-edited.
- Boundary and harness inputs: `internal/modules/entities/boundary_guard_test.go`,
  `tools/backend_module_boundaries.json`, and
  `tools/test_families/module.entities.json`.
- Entities implementation: `hostidentity/api.go`, `field_registry.go`,
  `patch_api.go`, `patch_store.go`, and `workbook_conflict_admission.go`.
- Entities tests: `hostidentity/field_registry_test.go`, `patch_api_test.go`,
  `workbook_admission_test.go`, and root `resolution_integration_test.go`.
- Timeline tests and test composition:
  `internal/modules/timeline/request_test.go`,
  `resolution_integration_test.go`, and
  `timeline_event_integration_test.go`.
- Frontend contract evidence:
  `apps/web/src/workbook/models/entityWorkbookModel.test.ts`.

### Compatibility, migration, and rollback

The sole intentional public behavior correction is removal of ordinary patch
support and grid editability for the four canonical create-only identifiers.
Clients using that out-of-contract behavior must use the existing sanctioned
create/upsert workflows. Backend and generated web metadata change together.
Existing rows and history remain valid; there is no database migration,
backfill, route or operation change, schema or field-key change, alias,
deprecation shim, feature flag, version fork, dual write, or event-shape
change.

For a pre-merge rollback, revert each completed slice atomically and regenerate
derived outputs from restored authored inputs. Do not retain both the registry
and the superseded maps/switches. After release, repair the authored contract
forward rather than restoring legacy patch admission. Existing
`unsupported_field_key` telemetry may be monitored for these keys without
logging submitted values.

No unresolved product, specification, security, harness, ownership, or data
migration risk is known. `make release-check` was not run because release
certification is explicitly outside this remediation's acceptance scope.

### Final validation evidence

The mandatory S-06 gates ran in the prescribed order:

| Order | Gate | Result and retained root |
| --- | --- | --- |
| 1 | 51/51 selector reconciliation | PASS; AST reconciliation `.cartulary/test-results/20260822T041739Z-p1392676`, with static counts `discovered=51 selected=51 unique_selected=51`. |
| 2 | `make test-catalog-check` | PASS; `.cartulary/test-results/20260822T041754Z-p1393676`. |
| 3 | `make harness-contract` | PASS; `.cartulary/test-results/20260822T041804Z-p1394002`. |
| 4 | `make backend-module-boundary-check` | PASS; `.cartulary/test-results/20260822T041822Z-p1394608`. |
| 5 | `make generate-drift` | PASS; `.cartulary/test-results/20260822T041832Z-p1394993`. |
| 6 | `make generated-artifact-policy-check` | PASS; `.cartulary/test-results/20260822T041843Z-p1397929`. |
| 7 | `make json-shape-check` | PASS; `.cartulary/test-results/20260822T041851Z-p1398380`. |
| 8 | Entities and Timeline focused/service-backed slices | PASS; Entities focused 38/38 `.cartulary/test-results/20260822T041901Z-p1398909`, Timeline focused 51/51 `.cartulary/test-results/20260822T041901Z-p1398914`, Entities service-backed 29/29 `.cartulary/test-results/20260822T042339Z-p1508994`, Timeline service-backed 29/29 `.cartulary/test-results/20260822T042339Z-p1508991`. |
| 9 | `make frontend-unit` | PASS 390/390; `.cartulary/test-results/20260822T042817Z-p1617522`. |
| 10 | `make frontend-typecheck` | PASS; `.cartulary/test-results/20260822T042824Z-p1617910`. |
| 11 | `make frontend-import-boundary-check` | PASS; `.cartulary/test-results/20260822T042837Z-p1618409`. |
| 12 | Four affected browser targets | PASS; webserver-backed 58/58 `.cartulary/test-results/20260822T042848Z-p1619000`, stateful 34/34 `.cartulary/test-results/20260822T042848Z-p1619007`, accessibility 12/12 `.cartulary/test-results/20260822T043301Z-p1714156`, visual 12/12 `.cartulary/test-results/20260822T043300Z-p1714080`. No visual golden changed. |
| 13 | `make agent-finalize` | PASS; `.cartulary/test-results/20260822T043447Z-p1796207`. `RESULTS_DIR` was unset because no qualifying successful full warm-check run had been supplied, so retained-run maintenance was skipped. |
| 14 | `make check` | Initial related failure 641/642 at `.cartulary/test-results/20260822T043505Z-p1799078` on two ST1005 error-message findings; narrow Go lint passed after correction, and the complete rerun passed 642/642 at `.cartulary/test-results/20260822T044053Z-p1929849`. |
| Final tracker | `make lint-markdown` | PASS; final handoff checkpoint `.cartulary/test-results/20260822T044725Z-p2044021`. |

## 14. EPR Scope and Planning Posture

EPR is the Entities Production-Readiness iteration. Sections 1 through 13 are
the completed execution history for the original remediation and remain the
evidence for that work. EPR does not reopen its owner decision, create-only
field correction, automatic-resolution evidence, field registry, or final
validation.

This tracker remains an execution-support artifact rather than product
authority. Adopted Core specifications and
`docs/decisions/entities-module-boundary.md` continue to define the behavior
and owner topology preserved by EPR. The bounded-root, no-package-split
decision remains in force.

EPR-S00 is a documentation-only rebaseline slice. It changes no production,
test, authored contract, generated artifact, migration, or application-
composition file. EPR-S01 explicitly adopts the missing construction,
production-export, Host/Identity capability, and consumer-port rules in the
Entities boundary decision and Core 04 AC-558 before implementation begins.
Core 00 REQ-00-073 already owns the applicable topology rule and receives no
semantic change. `docs/domain.md` remains unchanged.

### 14.1 Planning baseline

| Baseline fact | EPR-S00 record |
| --- | --- |
| Commit | `769131d31ffb4b494ae24e7826ba5e5144323a8a` |
| Repository posture | Clean worktree before this tracker edit |
| Date | 2026-08-22 |
| Entities files | 70 total: 51 production/support and 19 test files |
| Ordinary-build top-level Go tests | 51 discovered and 51 exactly selected |
| Entities verification rows | 33 total owner rows; 21 service-backed rows |
| Prior broad evidence | `make check` passed 642/642 at `.cartulary/test-results/20260822T044053Z-p1929849` |
| Prior tracker evidence | `make lint-markdown` passed at `.cartulary/test-results/20260822T044725Z-p2044021` |

The prior green roots establish the planning baseline only. They do not prove
any EPR implementation work, and a later EPR implementation session must
reconcile live repository state before changing code.

### 14.2 Iteration objective

EPR first adopts the missing owner rules, then removes accidental and test-only
production surface, replaces partial and panic-based construction with
complete fallible capabilities, moves cross-owner merge translation to
application assembly, contracts the Host/Identity merge bridge, and leaves the
package in a cohesive steady-state layout. It preserves all stable public and
persisted behavior because those contracts protect active clients, security
ordering, auditability, recovery, and transaction correctness.

The iteration is repository-internal. Internal Go APIs intentionally break
without aliases, forwarding wrappers, overloads, `Must` constructors, feature
flags, or deprecation windows. No public or persisted compatibility layer is
added solely to retain current implementation shape.

### 14.3 Authority and stop conditions

The sequence is:

`EPR-S00 -> EPR-S01 -> EPR-S02 -> EPR-S03 -> EPR-S04 -> EPR-S05 -> EPR-S06`

Only one slice may be active or validating. A successor is ineligible until
its predecessor passes focused validation, updates this tracker, and passes
`make lint-markdown`. A required check failure, unexplained drift, hand-edited
generated file, incomplete selector account, normative behavior change,
database migration, or owner contradiction marks the owning slice `BLOCKED`.
Later slices do not start until that condition is resolved through the
appropriate owner process.

## 15. EPR Retain, Replace, Remove, and Defer Decisions

| Surface | Disposition | Decision and reason |
| --- | --- | --- |
| Root HTTP registration | Retain | `RouteOptions` and `RegisterRoutes` are the bounded owner HTTP facade consumed by server assembly. |
| Source-owner contributions | Retain | Revision, Recovery, Incident Bundle source/subtype, Reporting, projection, delete/restore, and rollback contributions have live assembly consumers or required interface methods. |
| `workbookprojection` | Retain | It is the active typed Entities/Projections language, not legacy residue. Redesign requires a separate cross-owner plan. |
| Timeline fact and write ports | Retain | Timeline owns policy and its borrowed transaction; Entities owns source facts and writes. |
| Request decoders and hashes used by application assembly | Retain | They are active typed admission boundaries and preserve idempotency identity. |
| Runtime-excluded Entities test support | Retain | It has current owner and cross-owner test consumers and remains excluded from production imports. |
| `StoreOption` and `With*` APIs | Replace | Required capabilities become explicit dependency-struct fields on fallible constructors. No compatibility overload remains. |
| Broad Host/Identity `Store` use by Timeline, Assessments, and Imports | Replace | Workbook retains a complete store; source facts and import mutation use narrow purpose-built capabilities. |
| Concrete Assessments adapter inside `merge` | Replace | A consumer-owned merge port remains in Entities; concrete conversion moves to `internal/app/entitymergeassembly`. |
| Free Host/Identity merge bridge functions and aliases | Replace | One immutable merge capability owns the internal source operation surface. |
| Root `Service` | Privatize | No production caller names the type; route registration is the public capability. |
| Root Incident Bundle codec functions | Privatize | They are used only by the root source-port contribution. |
| Unbounded Host/Identity query wrappers | Remove | Production uses paged queries; the wrappers exist only as test conveniences. Tests use the paged contract directly. |
| Row, payload, alias-item, and clipboard-hash helpers used only in their package | Privatize | Capitalization provides no continuing cross-package value. |
| Mention HTTP error constructors used only by root routes | Remove and relocate | The root HTTP facade owns transport translation; Mention application errors remain typed. |
| Implementation-only type aliases | Remove | They do not provide semantic separation and unnecessarily enlarge the compatibility surface. |
| Stable routes, operation IDs, schema IDs, field keys, bundle version 2, and negative security/upgrade evidence | Retain | These are active contracts or durable evidence, not legacy implementation code. |
| Persisted idempotency-format redesign | Defer | `route_idempotency` is an active shared cross-owner contract. Its redesign requires an Authentication/platform-owned migration plan and is not dead Entities code. |

## 16. EPR Requirements and Gap Register

### 16.1 Requirements

| Requirement | Planned outcome | Primary evidence |
| --- | --- | --- |
| EPR-REQ-001 | Preserve Sections 1 through 13 as completed historical evidence while making Sections 14 through 22 the controlling next-iteration ledger | Tracker review and Markdown lint |
| EPR-REQ-002 | Execute EPR serially and checkpoint this tracker after every slice before starting its successor | Section 20 ledger |
| EPR-REQ-003 | Adopt and enforce module-wide production-export closure: every remaining export has a live consumer, contribution role, or required interface method | Entities boundary decision, Core 04 AC-558, exact AST inventory, and reachability report |
| EPR-REQ-004 | Remove accidental exports and test-only production conveniences without aliases or shims | Zero-reference and export-guard evidence |
| EPR-REQ-005 | Adopt and implement fallible, complete Host/Identity, Mention, and Merge construction, including nil and typed-nil rejection | Entities boundary decision, Core 04 AC-558, and constructor matrix |
| EPR-REQ-006 | Adopt and implement separate Workbook, source-fact, import-create, and merge Host/Identity capabilities | Owner amendments, composition, and capability tests |
| EPR-REQ-007 | Adopt a consumer-owned merge port and move concrete Assessments translation to application assembly | Owner amendments, import graph, and merge port tests |
| EPR-REQ-008 | Replace the free Host/Identity merge bridge with one immutable owner-local capability | Owner amendments, method-set, and retired-symbol evidence |
| EPR-REQ-009 | Preserve HTTP, OpenAPI, authorization, idempotency, row, history, rollback, projection, bundle, and Collaboration behavior | Owner and cross-owner suites |
| EPR-REQ-010 | Keep every active ordinary-build Entities test exactly selected throughout the iteration | AST selector reconciliation |
| EPR-REQ-011 | Improve steady-state file cohesion without a package split or duplicate implementation | File inventory, review, and owner suites |
| EPR-REQ-012 | Complete EPR only after broad developer and release gates pass and the handoff records all evidence | EPR-S06 ledger and checklist |

### 16.2 Gap decisions

| Gap | Areas and remediation | Rationale | Expected long-term benefit | Compatibility or migration impact | Risk if unresolved | Validation criteria |
| --- | --- | --- | --- | --- | --- | --- |
| EPR-G01 — Completed work is still the active tracker posture | Documentation: retain the completed ledger and append the EPR scope, gaps, workstreams, validation, deferrals, and handoff sections. | Rewriting history loses auditability; leaving it controlling directs later work at closed gaps. | One current execution ledger with intact prior proof. | Documentation only. | A later session can repeat closed work or mistake historical green evidence for EPR completion. | Header distinguishes both iterations; EPR-S00 through EPR-S06 agree; scoped diff and Markdown gates pass. |
| EPR-G02 — Production exports and reachability are not closed by the adopted owner | Specification, tests, and tracker: adopt module-wide export closure, add an AST exact disposition inventory and synthetic negative fixture, then privatize or remove accidental surfaces in EPR-S04. | Capitalization and test usage do not establish a durable API contract. | A small, deliberate module surface that resists accidental coupling and makes later decomposition safer. | Repository-internal Go breaks only; no aliases or deprecation shims. | Unsupported helpers can become permanent compatibility obligations. | Root and production-child exports are exact; test support is excluded; every retained export has independent justification; the negative fixture fails. |
| EPR-G03 — Constructors admit partial objects, panics, and delayed nil failures | Specification, implementation, composition, and tests: adopt complete fallible construction, replace variadic options with dependency structs, and reject nil and typed-nil dependencies without panic. | Successful construction must prove every operation on the returned capability is safe to invoke. | Startup-time failure, predictable dependency growth, and simpler composition. | Atomic internal constructor cutover; public and persisted behavior is unchanged. | Missing projections, history, links, or Collaboration capabilities can fail after mutation starts. | Complete dependency matrices return declaration-ordered errors and nil results; valid composition succeeds; no option, `Must`, fallback, or panic path remains. |
| EPR-G04 — Host/Identity consumers depend on an overbroad, inconsistently initialized store | Specification, implementation, app composition, and tests: separate the complete Workbook store, borrowed-transaction source facts, import creation, and merge source operations. | Workbook, Timeline, Assessments, Imports, and merge need materially different capabilities. | Narrow method-set coupling, explicit transaction ownership, and easier workflow expansion. | Internal compile-time migration only. | Optional fields remain hidden invalid state and unrelated consumers inherit Workbook mutation dependencies. | Workbook alone receives the complete store; other consumers receive only their declared capability; no partial Host/Identity store remains. |
| EPR-G05 — Merge owns concrete Assessments translation and an ad hoc Host/Identity bridge | Specification, implementation, composition, and tests: adopt a merge-owned Assessment effects port, translate concrete Assessments types in application assembly, and consolidate source operations behind an immutable owner-local merge capability. | A consumer owns its port language; concrete cross-owner conversion belongs at composition. | Cleaner ownership, isolated error translation, and a stable seam for future merge participants. | Internal Go break only; transaction order, protected-set semantics, rollback, and public errors remain exact. | New merge participants would amplify concrete imports, aliases, and scattered wrappers. | Entities production has no concrete Assessments implementation import; protected-set errors retain their code; old bridges and aliases are absent; rollback evidence passes. |
| EPR-G06 — Mixed files and phase-era layout obscure steady-state responsibilities | Implementation layout and tests: split merge, matching, and oversized test files along existing behavior boundaries; remove transitional names/comments. | The former implementation sequence should not remain encoded in permanent layout. | Smaller review units and clearer change ownership. | Structural only; no new package or verification owner. | Mixed sequencing encourages duplication and unsafe partial edits. | No package split, duplicate sequence, stale path, phase residue, or selector loss remains. |
| EPR-G07 — Production-readiness deletion evidence is not assembled | Validation and handoff: run exact reachability, owner, cross-owner, generated, browser, broad, build, and release gates and record all roots. | Local compilation cannot prove removed APIs lack hidden consumers. | Reproducible production-readiness evidence. | None beyond earlier internal breaks. | Dead shims, stale routing, or behavior drift can survive narrow tests. | EPR-S06 passes, gaps close, and tracker status, evidence, risks, and checklist agree. |

## 17. EPR Execution Policy and Checkpoint Gate

Every authorized slice is atomic. After implementation and focused validation,
and before its successor starts, update this tracker with:

- slice status and requirement/gap changes;
- every authored, moved, generated, and deleted file;
- exact commands, results, and result or artifact roots;
- test and verification-row counts when changed;
- substantive behavior and internal-interface changes;
- compatibility and migration impact;
- rollback boundary;
- failures, causal attribution, skipped checks, and residual risks; and
- the next eligible slice and its authorization state.

`make lint-markdown` is the last gate for each checkpoint. Generated roots and
generated Harness topology are changed only through authored inputs and public
Make targets. Runtime, tests, tools, generators, and conformance code must not
read or derive facts from this tracker.

EPR-S01 is explicitly authorized to amend
`docs/decisions/entities-module-boundary.md` and Core 04 AC-558 with the
construction, export-closure, capability-specific consumer, and consumer-port
rules required by this plan. No other EPR slice may broaden into a Core or
specification change, migration, package split, public behavior change, or
shared idempotency redesign. Such a discovery returns to planning and owner
adoption rather than receiving a tactical compatibility path.

## 18. EPR Serial Workstreams

| Slice | Workstream | Depends on | Status | Authorization | Rollback boundary | Exit criterion |
| --- | --- | --- | --- | --- | --- | --- |
| EPR-S00 | Tracker rebaseline | Completed S-06 | COMPLETE | Authorized and complete | Revert only the EPR planning revision and header posture. | Tracker-only diff and Markdown lint passed; EPR-S01 is ready and inactive. |
| EPR-S01 | Owner adoption, reachability, and behavior characterization | EPR-S00 checkpoint | COMPLETE | Authorized and complete | Revert the owner amendments, characterization, export guard, and routing changes together. | Owner amendments are adopted; current behavior and every production export have executable dispositions; accounting remains 51/51. |
| EPR-S02 | Complete construction and Host/Identity capability separation | EPR-S01 checkpoint | COMPLETE | Authorized and complete | Revert constructors, capability split, callers, tests, authored rows, and generated topology together. | Complete construction is fallible, partial stores/options are absent, and accounting is 52/52. |
| EPR-S03 | Merge ports and owner-boundary cleanup | EPR-S02 checkpoint | COMPLETE | Authorized and complete | Revert merge ports, app adapters, Host/Identity capability, constructor wiring, and tests together. | Concrete Assessments coupling and free bridge wrappers are absent with merge parity green. |
| EPR-S04 | Accidental API removal and final export closure | EPR-S03 checkpoint | COMPLETE | Authorized and complete | Revert removals, callers, tests, and the final export inventory together. | Retired surfaces have zero references and the final inventory is exact. |
| EPR-S05 | Steady-state cohesion cleanup | EPR-S04 checkpoint | COMPLETE | Authorized and complete | Revert file/test moves and terminology changes without changing S04 APIs. | Steady-state files are cohesive and all test identities remain exactly routed. |
| EPR-S06 | Final validation and handoff | EPR-S05 checkpoint | COMPLETE | Authorized and complete | Corrections return to their owning slice; S06 changes evidence and tracker state. | All required developer, browser, security, and release gates pass; EPR-G01 through EPR-G07 are closed. |

### 18.1 EPR-S00 — Tracker rebaseline

- **Areas:** this tracker only.
- **Remediation:** Update the header and append Sections 14 through 22 as the
  controlling delta plan. Preserve the completed Sections 1 through 13 except
  for the header posture needed to distinguish the iterations.
- **Compatibility:** Documentation only. Domain vocabulary is unchanged and
  `docs/domain.md` is not edited.
- **Validation:** `git diff --check --
  docs/handoffs/entities-module-refactor-tracker.md`, `make lint-markdown`, and
  final changed-path review.
- **Exit:** Only this tracker changed; EPR-S00 is `COMPLETE`; EPR-S01 is
  authorized and `READY`; EPR-S02 through EPR-S06 remain authorized and
  dependency-gated as `PLANNED`.

### 18.2 EPR-S01 — Reachability and behavior characterization

- Amend `docs/decisions/entities-module-boundary.md` and Core 04 AC-558 to
  adopt module-wide production-export closure, complete fallible
  construction, capability-specific Host/Identity consumers, and
  consumer-owned cross-owner adapter placement. Do not change Core 00 or
  `docs/domain.md`.
- Inventory every production export in the root and owner-facing subpackages.
  Classify it as retain, privatize, remove, or replace with a typed capability.
- Extend `TestEntitiesProductionImportBoundaries` with an exact AST allowlist
  and a synthetic negative fixture. Exclude owner-local test support from the
  production allowlist.
- Freeze HTTP/OpenAPI identity, authorization and concealment order,
  idempotent replay, row and hash shapes, merge and mention atomicity,
  contribution identities, Incident Bundle bytes, rollback, projection, and
  Collaboration consequences.
- Validate test catalog, Harness contract, backend boundary, and focused and
  service-backed Entities slices.
- Exit at 51 discovered, 51 selected, and no unexplained export or removal
  candidate.

### 18.3 EPR-S02 — Fail-safe construction and capability separation

Adopt these repository-internal construction surfaces:

- `hostidentity.NewStore(hostidentity.StoreDependencies) (*Store, error)` for
  complete Workbook behavior;
- `hostidentity.NewSourceFacts()` for stateless borrowed-transaction alias and
  target facts;
- `hostidentity.NewImportCreateFacade(targetViewSchemaID, facadeID,
  hostidentity.ImportDependencies)` using a private import owner rather than a
  partial store;
- `mentions.NewStore(mentions.StoreDependencies) (*Store, error)`; and
- `merge.NewStore(merge.StoreDependencies) (*Store, error)`.

Remove all `StoreOption` and `With*` APIs. Require Postgres, Revisions,
projection, Timeline, Collaboration, Links, Mentions, Assessments, query, and
conflict capabilities wherever the constructed surface needs them. Reject nil
and typed-nil dependencies deterministically. Do not add `Must` helpers,
overloads, or transitional constructors.

Rename `TestNewStoreRequiresLinkOperations` and
`TestWithMentionStoreRetainsInjectedInstance` to complete-construction
contract names. Add exactly one new top-level test,
`TestHostIdentityStoreCompositionRequiresCompleteDependencies_Unit`, and one
exact authored unit row. Regenerate topology through Make. Expected final
accounting is 52/52, 34 owner rows, and 21 service-backed rows.

Validate Entities, Timeline, Workbook, Imports, Assessments, and app-server
composition, plus catalog, Harness, boundary, generation, and server-build
gates.

### 18.4 EPR-S03 — Merge port and owner-boundary cleanup

- Define `merge.AssessmentEffectsPort` and typed assessment mutation and
  precondition DTOs.
- Move the concrete Assessments adapter, mutation translation, and
  `MergeProtectedSetChangedError` translation into
  `internal/app/entitymergeassembly`.
- Replace exported exact-match precedence, normalization, transactional
  load/update, preserved-identifier synchronization, alias synchronization,
  and bridge aliases with one immutable Host/Identity merge capability.
- Inject both capabilities through `merge.StoreDependencies`. Timeline keeps
  merge coordination and borrowed-transaction assembly; no policy moves.
- Validate Entities, Assessments, Timeline, Revisions, and app-server focused
  and service-backed paths, including protected-set and merge rollback cases.

### 18.5 EPR-S04 — Dead and accidental API removal

Privatize or remove the proven implementation-only surfaces:

- root `Service`, `ExportIncidentBundleFiles`, and
  `ImportIncidentBundleFilesTx`;
- Host/Identity `QueryHostRows` and `QueryIdentityRows` test-convenience
  wrappers;
- `BuildHostRow`, `BuildIdentityRow`, `BuildMutationPayload`,
  `ParseEntityAliasItemRef`, and `EntityClipboardPasteRequestHash`;
- `merge.BuildMergePayload`;
- `ImportCreateCommand`, `MergeTimelineInvalidation`, `entityTimelinePort`,
  and superseded bridge aliases; and
- Mention HTTP error constructors whose sole caller is the root route facade,
  moving their transport mapping to root-owned HTTP helpers.

Update the exact export allowlist in the same slice. Keep all Section 15 retain
decisions. Validate zero references for retired names, 52/52 selector
accounting, affected owner suites, boundary, Harness, generated state, and
`make test-fast`.

### 18.6 EPR-S05 — Cohesion and historical-residue cleanup

- Split `merge_store.go` into transaction coordination, protected-set and
  admission, source carry-forward, collision detection, revision/result, and
  replay units without changing order or behavior.
- Split `hostidentity/match.go` into exact matching, upsert,
  alias synchronization, and preserved-identifier units.
- Split oversized root tests by route or behavior while preserving all
  top-level test names and selectors after the explicit S02 renames.
- Remove phase-era comments and misleading filenames without creating a new
  package, registry, shared kernel, or duplicate implementation.
- Validate format, exact paths and selectors, affected owner slices, boundary,
  catalog, and Harness contract.

### 18.7 EPR-S06 — Final validation and handoff

Run the Section 19 final gates in order. Record every success and failure,
result root, compatibility conclusion, generated-file provenance, rollback
posture, skipped check, and residual risk. A failed mandatory gate remains a
failure and prevents completion.

## 19. EPR Validation Matrix

### 19.1 Per-slice minimums

| Slice | Required minimum validation |
| --- | --- |
| EPR-S00 | Tracker-scoped `git diff --check`; `make lint-markdown`; changed-path review |
| EPR-S01 | `make test-catalog-check`; `make harness-contract`; `make backend-module-boundary-check`; focused and service-backed Entities slices |
| EPR-S02 | `make format`; `make generate`; catalog, Harness, boundary, generation drift; focused/service-backed Entities, Timeline, Workbook, Imports, Assessments, and app-server paths; `make build-server` |
| EPR-S03 | `make format`; boundary; focused/service-backed Entities, Assessments, Timeline, Revisions, and app-server paths; exact old-import and bridge scans |
| EPR-S04 | `make format`; exact export and retired-symbol scans; catalog, Harness, boundary, generated-state gates; affected owner slices; `make test-fast` |
| EPR-S05 | `make format`; file and selector reconciliation; affected owner slices; catalog, Harness, and boundary gates |
| EPR-S06 | Complete ordered final matrix below |

Use `make task-guide ROLE=module-author OWNER=<owner-id>` and owner
explanations at execution time to select the exact affected rows before broad
runs. Do not substitute direct Go, frontend, or browser tool invocations for
their public Make targets.

### 19.2 Ordered final matrix

1. Reconcile exact exports, retired symbols, imports, files, and 52/52 test
   selectors.
2. `make test-catalog-check`
3. `make harness-contract`
4. `make backend-module-boundary-check`
5. `make generate-drift`
6. `make generated-artifact-policy-check`
7. `make json-shape-check`
8. Run focused and service-backed slices for Entities, Timeline, Assessments,
   Workbook, Imports, Revisions, Incident Bundles, and app server where live
   routing reports them affected.
9. `make frontend-unit`
10. `make frontend-typecheck`
11. `make frontend-import-boundary-check`
12. Run affected `browser-e2e-webserver-backed`, `browser-e2e-stateful`,
    `browser-e2e-a11y`, and `browser-e2e-visual` targets.
13. `make build`
14. Run `make agent-finalize`, supplying `RESULTS_DIR` only for a qualifying
    current successful full warm-check root.
15. `make check`
16. `make release-check`
17. Run the final tracker-scoped diff check and `make lint-markdown`.

## 20. EPR Tracker and Checkpoint Ledger

### 20.1 Work tracker

| Work item | Description | Slice | Status | Dependency | Evidence | Completion condition |
| --- | --- | --- | --- | --- | --- | --- |
| EPR-001 | Rebaseline this tracker and establish the controlling EPR plan. | EPR-S00 | DONE | Completed S-06 | Both staged and unstaged tracker-scoped diff checks passed; Markdown passed at `.cartulary/test-results/20260822T061030Z-p2079549`; changed-path review found only this tracker. | Only the tracker changes; scoped diff and Markdown gates pass. |
| EPR-002 | Adopt the owner rules and lock the current production export and behavior surface. | EPR-S01 | DONE | EPR-001 | Boundary decision and Core 04 AC-558 adopted the required rules; the exact AST inventory, package discovery, role justifications, and negative fixture passed; 51/51 selectors and 33/21 rows remained exact. | Owner amendments are adopted and every export and removal candidate has executable disposition. |
| EPR-003 | Replace partial and panic-based construction with complete fallible capabilities. | EPR-S02 | DONE | EPR-002 | Host/Identity, Mention, Merge, and import construction reject every missing and typed-nil dependency without panic, return nil on failure, report declaration-ordered errors, and accept complete composition; all option APIs are absent. | Dependency matrices and affected assembly pass; no option API remains. |
| EPR-004 | Separate complete Workbook, source-fact, and import Host/Identity capabilities. | EPR-S02 | DONE | EPR-002 | Workbook alone receives `Store`; Timeline and Assessments receive stateless `SourceFacts`; Imports receives a facade over the shared private mutation core. | No partial Host/Identity store exists. |
| EPR-005 | Move Assessments merge translation and contract the Host/Identity merge bridge. | EPR-S03 | DONE | EPR-003, EPR-004 | Merge owns `AssessmentEffectsPort` and its command, result, mutation, and protected-set error language; `internal/app/entitymergeassembly` owns concrete Assessments translation and defensive cloning; immutable `hostidentity.MergeCapability` replaces the free bridge; exact old-import and old-symbol scans are empty. | Cross-owner adapter location and exact method set pass. |
| EPR-006 | Remove or privatize accidental production APIs and aliases. | EPR-S04 | DONE | EPR-005 | Root transport and Incident Bundle codecs, Host/Identity serializers/parser/hash, and merge payload are private; unbounded query wrappers, superseded DTO/port aliases, and Mention-owned HTTP mappings are removed; exact scoped scans are empty and the final export inventory accepts retained declarations only. | Zero retired references and exact final export allowlist. |
| EPR-007 | Establish steady-state implementation and test layout. | EPR-S05 | DONE | EPR-006 | Host/Identity is divided into construction, paged query, create mutation, source persistence/hydration, exact matching, upsert, alias synchronization, and preserved-identifier files; merge is divided into transaction/replay, admission, effects, source carry-forward, collision, and history/result files; the three oversized root tests are split by unchanged top-level behavior. S06 reconciliation also corrected the two authored projection-provider references and test-support caller matrix that still named the removed aggregate test, removed five newly dead fixture helpers, and regenerated the derived topology index. | Cohesive files retain exact behavior and routing; every old path is absent from active owner inputs and generated projections. |
| EPR-008 | Reconcile final 52/52 selector and 34/21 row accounting. | EPR-S06 | DONE | EPR-007 | Exact source count is 52 top-level Entities tests; `make explain-test-owner OWNER=module.entities` reports 34 rows and 21 service-backed rows; final export, retired-symbol, concrete-import, and stale-path scans are empty in their production scopes. | Counts are exact with zero anomalies. |
| EPR-009 | Run final validation and publish the production-readiness handoff. | EPR-S06 | DONE | EPR-008 | Section 20.3 records every mandatory final gate, owner root, failure, retry, generated artifact, compatibility conclusion, and rollback boundary; `make check` passed 642/642 and `make release-check` passed 799/799. | Every required gate passes and EPR-G01 through EPR-G07 close. |

### 20.2 Session handoff log

| Time | Session | Current state | Files inspected or touched | Commands and results | Compatibility and rollback | Blockers and next action |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-08-22 | Codex / EPR-S00 planning | EPR tracker rebaseline complete; EPR-S01 ready but unauthorized | Inspected the completed tracker, Entities exports/callers, constructor and merge composition seams, owner task guidance, and current repository status; touched only this tracker | Read-only `rg`, `find`, `sed`, `jq`, `git`, and `make explain-test-owner OWNER=module.entities`; tracker-scoped `git diff --check` passed; `make lint-markdown` passed at `.cartulary/test-results/20260822T054206Z-p2066998`; changed-path review found only this tracker | Documentation only; no runtime, test, contract, generated, migration, or Domain change; rollback reverts the EPR header posture and Sections 14 through 22 | Superseded only as an authorization statement by the later implementation directive; retained as historical evidence |
| 2026-08-22 | Codex / EPR-S00 authorization rebaseline | Approved EPR plan reconciled; EPR-S00 complete and EPR-S01 ready | Reconciled Sections 14 through 22 and the header against the approved end-to-end directive; touched only this tracker | Both staged and unstaged tracker-scoped diff checks passed; `make lint-markdown` passed at `.cartulary/test-results/20260822T061030Z-p2079549`; changed-path review found only this tracker | Documentation only; no product, test, contract, generated, migration, Core, or Domain change; rollback reverts this authorization rebaseline only | None; activate EPR-S01 after this passed checkpoint |
| 2026-08-22 | Codex / EPR-S01 owner adoption and characterization | Owner adoption and current-surface closure complete; EPR-S02 ready | Changed `docs/decisions/entities-module-boundary.md`, Core 04 AC-558, `internal/modules/entities/boundary_guard_test.go`, and this tracker; no routing or generated file changed | `make format` passed at `.cartulary/test-results/20260822T061636Z-p2085999`; focused boundary row passed at `.cartulary/test-results/20260822T061619Z-p2085274`; `make test-catalog-check` passed; `make harness-contract` passed at `.cartulary/test-results/20260822T061653Z-p2090207`; `make backend-module-boundary-check` passed at `.cartulary/test-results/20260822T061709Z-p2090807`; focused Entities passed at `.cartulary/test-results/20260822T061714Z-p2091187`; service-backed Entities passed at `.cartulary/test-results/20260822T061900Z-p2144317`; 51/51 selectors, 33 owner rows, and 21 service-backed rows remained exact | Specification and tests only; no public, persisted, generated-contract, migration, Core 00, or Domain change; rollback reverts the decision, AC-558 amendment, and boundary guard together | Preliminary `.cartulary/test-results/20260822T061334Z-p2082468` failed before selector execution because the new helper accepted `testing.TB` while invoking `Run`; corrected to `*testing.T`. `.cartulary/test-results/20260822T061358Z-p2083215` then failed as the deliberately empty inventory exposed all current exports; the reviewed exact dispositions replaced it. No residual blocker; run tracker checkpoint, then activate EPR-S02 |
| 2026-08-22 | Codex / EPR-S02 complete construction and Host/Identity capability separation | Constructor and capability cutover complete; EPR-S03 ready | Added `internal/modules/entities/hostidentity/store_test.go`; changed Host/Identity store, import facade, source facts, mutation receivers, Mention and Merge constructors/tests, app and test composition, Timeline tests, the boundary inventory, and `tools/test_families/module.entities.json`; generated `tools/execution_topology_render_index.json` only through `make generate`; changed this tracker; no file was deleted | Semantic row IDs derived through `make author-test-row-id`; final `make format` passed at `.cartulary/test-results/20260822T065316Z-p2902466`; `make generate` passed at `.cartulary/test-results/20260822T063336Z-p2224409`; final constructor and boundary rows passed at `.cartulary/test-results/20260822T065320Z-p2906131`; `make test-catalog-check` passed; generation drift passed at `.cartulary/test-results/20260822T063411Z-p2228779`; Harness passed at `.cartulary/test-results/20260822T063420Z-p2231757`; backend boundary passed at `.cartulary/test-results/20260822T063433Z-p2232326`; focused Entities, Timeline, Imports, Assessments, app server, and Workbook passed at `.cartulary/test-results/20260822T063502Z-p2234285`, `.cartulary/test-results/20260822T063502Z-p2234301`, `.cartulary/test-results/20260822T063502Z-p2234319`, `.cartulary/test-results/20260822T063502Z-p2234375`, `.cartulary/test-results/20260822T064036Z-p2509596`, and `.cartulary/test-results/20260822T064136Z-p2548565`; their service-backed slices passed at `.cartulary/test-results/20260822T064405Z-p2605011`, `.cartulary/test-results/20260822T064405Z-p2605015`, `.cartulary/test-results/20260822T064853Z-p2713876`, `.cartulary/test-results/20260822T065117Z-p2810633`, `.cartulary/test-results/20260822T065117Z-p2810632`, and `.cartulary/test-results/20260822T064853Z-p2713867`; `make build-server` passed at `.cartulary/test-results/20260822T065228Z-p2890469`; exact accounting is 52/52 tests, 34 owner rows, and 21 service-backed rows | Internal Go constructor break only; no alias, overload, shim, public or persisted behavior, OpenAPI, generated contract, database migration, Core, or Domain change. Rollback reverts constructors, capability split, all callers and tests, the authored manifest, and generated topology atomically | The first combined format/generate command failed because new authored row IDs were not ASCII-sorted; the manifest was reordered before generation. Concurrent focused runs for app server and Workbook failed at `.cartulary/test-results/20260822T063502Z-p2234388` and `.cartulary/test-results/20260822T063502Z-p2234338` when multiple graphs raced on `tmp/test-service-images/warm.stamp.tmp`; all emitted row results passed, and serial reruns passed. No residual blocker; run tracker checkpoint, then activate EPR-S03 |
| 2026-08-22 | Codex / EPR-S03 merge ports and owner-boundary cleanup | Merge owner boundaries are closed; EPR-S04 ready | Added `internal/app/entitymergeassembly/assessments.go` and `internal/modules/entities/hostidentity/merge_capability.go`; changed Timeline assembly, Merge ports/store/coordinator and protected-set tests, Host/Identity bridge callers, constructor tests, the exact export guard, and this tracker; no generated, contract, migration, lock, frontend, or Domain file changed and no file was deleted | Final `make format` passed at `.cartulary/test-results/20260822T070753Z-p2957875`; targeted adapter, protected-set, constructor, and boundary rows passed at `.cartulary/test-results/20260822T070757Z-p2961570`; backend boundary passed at `.cartulary/test-results/20260822T070854Z-p2977325`; exact production scans found no concrete Assessments import, `MentionStore`, `IdentifierSeed`, free precedence/normalization bridge, or exported transactional bridge call; focused Entities, Assessments, Timeline, Revisions, and app server passed at `.cartulary/test-results/20260822T070901Z-p2977778`, `.cartulary/test-results/20260822T070901Z-p2977786`, `.cartulary/test-results/20260822T071050Z-p3072596`, `.cartulary/test-results/20260822T071050Z-p3072603`, and `.cartulary/test-results/20260822T071539Z-p3173716`; service-backed Entities, Assessments, Timeline, Revisions, and app server passed at `.cartulary/test-results/20260822T071642Z-p3213242`, `.cartulary/test-results/20260822T071830Z-p3266061`, `.cartulary/test-results/20260822T072802Z-p3403654`, `.cartulary/test-results/20260822T073240Z-p3459930`, and `.cartulary/test-results/20260822T073345Z-p3503293`; final `make build-server` passed at `.cartulary/test-results/20260822T073442Z-p3542254`; accounting remains 52/52 tests, 34 owner rows, and 21 service-backed rows | Internal Go port and constructor break only; transaction order, protected-set error code, rollback, public and persisted behavior remain unchanged. No alias or forwarding shim was retained. Rollback reverts the merge port, application adapter, Host/Identity capability, constructor wiring, and tests atomically | The first server build failed at `.cartulary/test-results/20260822T070400Z-p2916092` because four owner-local Host/Identity callers still used retired exported load functions; all were migrated before the passing build. The initial full Timeline service graph failed at `.cartulary/test-results/20260822T071928Z-p3307054` when the last of 99 blank-row measurement samples timed out after 98 successes; the unchanged exact row passed at `.cartulary/test-results/20260822T072506Z-p3364527` and the unchanged full graph passed 29/29 at `.cartulary/test-results/20260822T072802Z-p3403654`, establishing a transient paint qualification failure unrelated to S03. No residual blocker; run the tracker checkpoint, then activate EPR-S04 |
| 2026-08-22 | Codex / EPR-S04 accidental API removal and final export closure | Accidental surfaces are absent and final export closure is enforced; EPR-S05 ready | Changed root Entities routes and Incident Bundle portability, Host/Identity API/query/mutation files and package tests, Mention ports, Merge API/ports/coordinator, root Entities tests, the boundary guard, and this tracker; no file was added, generated, deleted, or moved in this slice | Final `make format` passed at `.cartulary/test-results/20260822T074104Z-p3562975`; retained-only export closure and the synthetic negative fixture passed at `.cartulary/test-results/20260822T074006Z-p3561394`; exact scoped scans found zero retired declarations or references; `make test-catalog-check` passed; Harness passed at `.cartulary/test-results/20260822T074119Z-p3567083`; backend boundary passed at `.cartulary/test-results/20260822T074135Z-p3567688`; generation drift passed at `.cartulary/test-results/20260822T074141Z-p3568053`; focused and service-backed Entities passed at `.cartulary/test-results/20260822T074157Z-p3571043` and `.cartulary/test-results/20260822T074344Z-p3624828`; Incident Bundles passed at `.cartulary/test-results/20260822T074544Z-p3678102` and `.cartulary/test-results/20260822T074643Z-p3693731`; app server passed at `.cartulary/test-results/20260822T074743Z-p3709093` and `.cartulary/test-results/20260822T074842Z-p3748627`; `make test-fast` passed 421/421 at `.cartulary/test-results/20260822T074941Z-p3787817`; `make explain-test-owner OWNER=module.entities` reports 34 owner rows and 21 service-backed rows, and exact selector accounting remains 52/52 | Repository-internal Go removals only. No shim, alias, forwarding wrapper, public route/error, persisted payload, OpenAPI, bundle-byte, generated, migration, database, frontend, Core, or Domain change. Rollback reverts all removals, private callers/tests, and the final export inventory together | No S04 failure or residual blocker. The bounded test query window is 100 rows and preserves the fixture assertions without retaining an unbounded compatibility surface; run the tracker checkpoint, then activate EPR-S05 |
| 2026-08-22 | Codex / EPR-S05 steady-state cohesion cleanup | Cohesive production and test layout complete; EPR-S06 ready | Added eight Host/Identity responsibility files, six Merge responsibility files, four root unit/support files, six resolution integration files, and two support integration files; deleted superseded `hostidentity/store.go`, `hostidentity/match.go`, `merge/merge_store.go`, `unit_test.go`, `resolution_integration_test.go`, and `support_integration_test.go`; updated the boundary guard, `tools/backend_module_boundaries.json`, workbook-projection characterization references, one obsolete migration-era test message, and this tracker; no generated file changed | Final `make format` passed at `.cartulary/test-results/20260822T080652Z-p3931837`; `make test-catalog-check` passed; Harness passed at `.cartulary/test-results/20260822T080140Z-p3822101`; backend boundary passed at `.cartulary/test-results/20260822T080244Z-p3823683`; exact stale-path scans are empty; the export/selector boundary row passed at `.cartulary/test-results/20260822T080033Z-p3817041`; focused Entities passed 38/38 at `.cartulary/test-results/20260822T080257Z-p3824147`; service-backed Entities passed 29/29 at `.cartulary/test-results/20260822T080445Z-p3878470`; 52 top-level tests, 34 owner rows, and 21 service-backed rows remain exact | Structural only. Packages, APIs, transaction sequence, test names, test packages, selectors, public and persisted behavior, generated artifacts, migrations, and Domain vocabulary are unchanged. Rollback restores the six prior aggregate files, old boundary paths and characterization reference without reverting stabilized S04 APIs | Mechanical compile failures at `.cartulary/test-results/20260822T075407Z-p3801872`, `.cartulary/test-results/20260822T075426Z-p3802435`, `.cartulary/test-results/20260822T075536Z-p3808124`, `.cartulary/test-results/20260822T075632Z-p3809841`, and `.cartulary/test-results/20260822T075749Z-p3810920` exposed missing or unused imports while blocks were moved; they were corrected without code changes. `.cartulary/test-results/20260822T075554Z-p3808670` exposed a hard-coded retired merge filename in the boundary test, and `.cartulary/test-results/20260822T080156Z-p3822689` exposed retired SQL-read allowlist paths; both owner inputs now enumerate the new production files and pass. No residual blocker; run the tracker checkpoint, then activate EPR-S06 |
| 2026-08-22 | Codex / EPR-S06 final validation and handoff | Production-readiness remediation complete; all seven gaps and all binary criteria closed | Reconciled the full 79-path change set; corrected S05 projection/test-support path projections and dead helpers; generated the topology index only through `make generate`; finalized this tracker | Every ordered final gate passed. Exact roots, counts, failures, causal attribution, and retries are in Section 20.3. Final `make check` passed 642/642 at `.cartulary/test-results/20260822T090627Z-p883489`; retained-run finalization passed at `.cartulary/test-results/20260822T091101Z-p996639`; `make release-check` passed 799/799 at `.cartulary/test-results/20260822T091123Z-p999735`; post-handoff Markdown lint passed at `.cartulary/test-results/20260822T092800Z-p1211872`, and staged, unstaged, and whole-worktree diff checks passed | Internal Go cutovers only; no public or persisted behavior, migration, backfill, OpenAPI, schema, bundle, event, authorization, frontend, lockfile, or Domain change. Rollback remains atomic by owning slice; the S05 correction reverts with the layout slice and its derived topology. | No blocker or unclosed EPR risk. Persisted idempotency, workbook-projection redesign, package split, and versioned Incident Bundle retirement remain explicitly deferred. |

Every future checkpoint appends a row rather than rewriting prior evidence.

### 20.3 Final production-readiness handoff

#### Outcome and gap closure

EPR-S00 through EPR-S06 are complete. EPR-G01 through EPR-G07 are closed:
the ledger is current, the adopted owner closes all production exports and
construction, Host/Identity and Merge capabilities follow consumer-owned
boundaries, accidental APIs are gone, steady-state files are cohesive, and
current broad/release evidence validates the result. The executable final
export inventory is the retained-only inventory in
`internal/modules/entities/boundary_guard_test.go`; it discovers the root and
all production child packages, excludes explicit test-support packages, and
rejects a synthetic unapproved export.

#### Final changed-file inventory

The final worktree differs from the starting commit
`769131d31ffb4b494ae24e7826ba5e5144323a8a` in 79 paths: 29 added, 44
modified, and 6 deleted. Git records the cohesion moves as deletions plus new
files rather than rename metadata.

- **Added — application adapter:**
  `internal/app/entitymergeassembly/assessments.go`.
- **Added — Host/Identity:** `alias_sync.go`, `construction.go`,
  `create_mutation.go`, `exact_match.go`, `merge_capability.go`,
  `preserved_identifiers.go`, `query.go`, `source_rows.go`, `store_test.go`,
  and `upsert.go` under `internal/modules/entities/hostidentity/`.
- **Added — Merge:** `admission.go`, `collision_detection.go`,
  `effect_coordination.go`, `history_result.go`, `source_carry_forward.go`, and
  `transaction.go` under `internal/modules/entities/merge/`.
- **Added — root tests:** `create_idempotency_integration_test.go`,
  `create_security_integration_test.go`, `exact_match_unit_test.go`,
  `mention_unit_test.go`, `merge_route_integration_test.go`,
  `merge_unit_test.go`, `origin_upsert_integration_test.go`,
  `resolution_route_integration_test.go`, `resolution_support_test.go`,
  `support_contract_integration_test.go`, `support_scenario_test.go`, and
  `unit_support_test.go` under `internal/modules/entities/`.
- **Deleted:** `internal/modules/entities/hostidentity/match.go`,
  `internal/modules/entities/hostidentity/store.go`,
  `internal/modules/entities/merge/merge_store.go`,
  `internal/modules/entities/resolution_integration_test.go`,
  `internal/modules/entities/support_integration_test.go`, and
  `internal/modules/entities/unit_test.go`.
- **Modified — owners and projections:**
  `docs/decisions/entities-module-boundary.md`, Core 04, this tracker,
  `contracts/projection-providers/index.json`,
  `tools/backend_module_boundaries.json`, and
  `tools/test_families/module.entities.json`.
- **Modified — application composition:** the six affected files under
  `internal/app/assessmentassembly`, `internal/app/importassembly`,
  `internal/app/timelineassembly`, and `internal/app/workbookassembly` named
  by the final changed-path audit.
- **Modified — existing Entities files:** the boundary guard, routes, Incident
  Bundle source/codec files, twelve Host/Identity files, two Mention files,
  six Merge files, and two workbook-projection files named by the per-slice
  ledger above.
- **Modified — cross-owner/test support:** two Timeline tests,
  `internal/testutil/appsupport/performancefixture/owners.go`,
  `internal/testutil/appsupport/workbook.go`, and
  `internal/testutil/httptestx/httptestx_test.go`.
- **Generated:** `tools/execution_topology_render_index.json`, derived only
  through `make generate` from the authored Entities row and projection-owner
  inputs. No generated root, lockfile, migration, frontend path, or
  `docs/domain.md` was hand-edited or otherwise changed.

#### Internal interfaces and behavior conclusion

The final code uses complete fallible dependency-struct constructors for
Host/Identity, Mentions, and Merge. Workbook alone receives the complete
Host/Identity store; Timeline and Assessments receive stateless borrowed-
transaction source facts; Imports uses the shared private mutation core; Merge
receives an immutable Host/Identity capability plus consumer-owned Mention and
Assessment effect ports. Concrete Assessment translation and protected-set
error conversion live in application assembly. There is one merge transaction
sequence and no compatibility constructor, option, alias, forwarding shim, or
test-only production convenience.

HTTP routes and operation IDs, OpenAPI and schema identities, field keys,
database state, bundle bytes, event shapes, authorization and concealment
order, idempotent replay, mutation hashes and rows, transaction order,
history/rollback, projections, Collaboration intents, and source-contribution
identities remain unchanged. Frontend and browser gates passed without source
or visual-golden changes. The compatibility impact is therefore limited to an
atomic repository-internal Go compile-time cutover; no migration, backfill,
feature flag, deprecation window, or deployment coordination is required.

#### Exact accounting and owner evidence

Final reconciliation found 52/52 top-level Entities test selectors and
`make explain-test-owner OWNER=module.entities` reports exactly 34 owner rows,
21 of them service-backed. Final affected-owner roots are:

| Owner | Focused result | Service-backed result |
| --- | --- | --- |
| Entities | 38/38, `.cartulary/test-results/20260822T080959Z-p3944475` | 29/29, `.cartulary/test-results/20260822T081148Z-p3997310` |
| Timeline | 51/51, `.cartulary/test-results/20260822T081334Z-p4050137` | 29/29, `.cartulary/test-results/20260822T081810Z-p4107134` |
| Assessments | 27/27, `.cartulary/test-results/20260822T082247Z-p4163702` | 18/18, `.cartulary/test-results/20260822T082346Z-p11508` |
| Workbook | 65/65, `.cartulary/test-results/20260822T082445Z-p52708` | 37/37, `.cartulary/test-results/20260822T082700Z-p109351` |
| Imports | 22/22, `.cartulary/test-results/20260822T082914Z-p165792` | 14/14, `.cartulary/test-results/20260822T083026Z-p207234` |
| Revisions | 27/27, `.cartulary/test-results/20260822T083144Z-p248332` | 20/20, `.cartulary/test-results/20260822T083249Z-p292974` |
| Incident Bundles | 8/8, `.cartulary/test-results/20260822T083356Z-p336688` | 6/6, `.cartulary/test-results/20260822T083455Z-p352500` |
| app server | 24/24, `.cartulary/test-results/20260822T083554Z-p368061` | 17/17, `.cartulary/test-results/20260822T083653Z-p407848` |

#### Ordered final validation evidence

| Gate | Result and current evidence |
| --- | --- |
| Reconciliation | PASS; retired production symbols, concrete Assessment imports, and active stale paths are absent; 52/52 and 34/21 are exact. |
| Catalog | `make test-catalog-check` PASS; the standalone target emits no result root. |
| Harness | PASS at `.cartulary/test-results/20260822T080842Z-p3937804`. |
| Backend boundary | PASS at `.cartulary/test-results/20260822T080858Z-p3938418`. |
| Generation drift | Initial PASS at `.cartulary/test-results/20260822T080906Z-p3938797`; post-correction PASS at `.cartulary/test-results/20260822T090549Z-p876637`. |
| Generated-artifact policy | Initial PASS at `.cartulary/test-results/20260822T080920Z-p3941727`; post-correction PASS at `.cartulary/test-results/20260822T090606Z-p880084`. |
| JSON shape | Initial PASS at `.cartulary/test-results/20260822T080929Z-p3942179`; post-generation PASS at `.cartulary/test-results/20260822T090559Z-p879599`. |
| Frontend unit | 390/390 PASS at `.cartulary/test-results/20260822T083752Z-p447061`. |
| Frontend typecheck | 2/2 PASS at `.cartulary/test-results/20260822T083803Z-p447499`. |
| Frontend import boundary | 2/2 PASS at `.cartulary/test-results/20260822T083816Z-p448067`. |
| Browser webserver-backed | 58/58 PASS at `.cartulary/test-results/20260822T083825Z-p448564`. |
| Browser stateful | 34/34 PASS at `.cartulary/test-results/20260822T084232Z-p500960`. |
| Browser accessibility | Final 12/12 PASS at `.cartulary/test-results/20260822T084859Z-p584917`. |
| Browser visual | 12/12 PASS with no golden refresh at `.cartulary/test-results/20260822T085027Z-p626365`. |
| Build | 7/7 PASS at `.cartulary/test-results/20260822T085210Z-p668197`. |
| S05 correction | Format PASS at `.cartulary/test-results/20260822T090037Z-p827653`; exact projection rows PASS at `.cartulary/test-results/20260822T090046Z-p831430`; backend integration 92/92 PASS at `.cartulary/test-results/20260822T090144Z-p852201`; generation PASS at `.cartulary/test-results/20260822T090529Z-p873564`. |
| Finalization | Pre-check no-retained-root PASS at `.cartulary/test-results/20260822T085232Z-p701180`; corrected current-source PASS at `.cartulary/test-results/20260822T090610Z-p880551`; retained full-check validation PASS with zero generated updates at `.cartulary/test-results/20260822T091101Z-p996639`. |
| Full check | 642/642 PASS at `.cartulary/test-results/20260822T090627Z-p883489`. |
| Release and security | 799/799 PASS at `.cartulary/test-results/20260822T091123Z-p999735`, including targeted and audit Gosec, `govulncheck`, SBOM, release inventory, readiness, and evidence-contract units. |

#### Failures, retries, skips, and residual risk

- The first accessibility run failed 10/12 at
  `.cartulary/test-results/20260822T084448Z-p544693` because the claimed
  Network Flow group timed out waiting for an object-store transport while the
  container was running. No accessibility assertion failed; the exact
  unchanged full target then passed 12/12.
- The first `make check` failed 638/642 at
  `.cartulary/test-results/20260822T085248Z-p704088`. Two projection/accounting
  rows still named the removed `resolution_integration_test.go`, the aggregate
  test-support run inherited the same stale caller-matrix expectation, and
  Go lint found five dead moved fixture helpers. The owner projections and
  expectation now name the cohesive files, the helpers are deleted, and exact
  rows, lint, backend integration, the full check, and release all pass.
- An attempted non-public `make backend-integration-testutil` command had no
  Make rule and no result root. The public `make backend-integration` target
  was used instead and passed 92/92.
- A correction-time `make agent-finalize` failed at
  `.cartulary/test-results/20260822T090500Z-p872621`; the confirming
  `make json-shape-check` failed at
  `.cartulary/test-results/20260822T090512Z-p873017` because the updated
  authored projection-provider index required regenerated topology. No
  generated output was hand-edited: `make generate` produced the exact index,
  and all drift, shape, policy, finalization, check, and release gates passed.
- No mandatory check is skipped. The first finalizer correctly omitted
  retained-run maintenance because no successful current full-check root yet
  existed; the later finalizer consumed the qualifying 642/642 root and
  validated it. Visual goldens were intentionally not refreshed because the
  visual gate passed unchanged.

No EPR implementation risk remains open. The only residual design work is the
four explicit deferrals in Section 21. Rollback is slice-atomic: revert the
owning owner/implementation/tests/projections together, regenerate derived
topology from the reverted authored inputs, and do not introduce a compatibility
shim. No next EPR slice is permitted or required; future feature work starts
from this completed production-ready boundary or from a separately adopted
owner plan.

## 21. EPR Deferrals, Risks, and Rollback Posture

| ID | Decision | Reason | Revisit trigger | Status |
| --- | --- | --- | --- | --- |
| EPR-DEF-01 | Do not redesign persisted route idempotency payloads. | `route_idempotency` is shared active behavior and durable replay history, not dead Entities code. | A separately adopted Authentication/platform migration plan covering all route owners. | DEFERRED |
| EPR-DEF-02 | Do not redesign `workbookprojection`. | It is a live Entities/Projections typed contract with multiple consumers. | A separately authorized cross-owner contract version. | DEFERRED |
| EPR-DEF-03 | Do not split the Entities bounded root into new owner packages. | The adopted boundary found one cohesive source owner and prohibited speculative package movement. | A new owner decision based on a distinct responsibility and migration evidence. | DEFERRED |
| EPR-DEF-04 | Do not remove versioned Incident Bundle behavior or negative security/upgrade evidence. | These surfaces protect active compatibility, security, or recovery semantics. | Explicit retirement evidence and owner adoption. | DEFERRED |

The primary execution risks are typed-nil dependency gaps, accidental removal
of interface-required methods, changed authorization or replay order during
constructor migration, cross-owner merge error drift, stale Harness selectors,
and behavior changes hidden by file movement. Their controls are the S01
characterization, per-slice exact scans and owner gates, atomic checkpoints,
and final broad/release validation.

Each implementation slice has the rollback boundary recorded in Section 18.
Do not retain dual constructors, aliases, or both old and new implementations
as rollback machinery. Before release, revert the affected slice atomically if
needed. After release, repair the clean capability forward unless an adopted
owner explicitly requires compatibility restoration.

## 22. EPR Binary Completion Criteria

EPR is complete only when every row is `PASS` with recorded evidence. `TODO`,
skipped, partial, inferred, or historically green results are not acceptance.

| Criterion | Required evidence | Current result |
| --- | --- | --- |
| The completed remediation remains intact and the EPR ledger is controlling. | Header and Sections 14 through 22; tracker-scoped diff and Markdown results. | PASS; EPR-S00 changed only this tracker, scoped diff passed, and Markdown passed at `.cartulary/test-results/20260822T054206Z-p2066998`. |
| Every production export is deliberate and locked. | Exact AST disposition inventory, production-package discovery, role justifications, and negative fixture. | PASS; the final retained-only inventory closes root and production-child exports and its synthetic unapproved export fails as designed. |
| All production constructors are fallible and complete. | Nil and typed-nil dependency matrices, declaration-order checks, valid composition, and absence of options/panics. | PASS; EPR-S02 constructor rows passed at `.cartulary/test-results/20260822T065320Z-p2906131`. |
| Host/Identity capabilities match their consumers. | Workbook, source-fact, import, Assessment, Timeline, and performance composition tests. | PASS; complete store, stateless source facts, and private import owner passed every affected focused and service-backed slice. |
| Merge depends on typed owner ports without concrete Assessments coupling or free Host/Identity bridge wrappers. | Import graph, adapter tests, old-symbol scans, and merge/rollback suites. | PASS; merge-owned ports, application translation and cloning, immutable Host/Identity capability, exact scans, and all affected focused and service-backed paths passed in EPR-S03. |
| Accidental APIs and aliases are absent without shims. | Exact zero-reference scans and final export guard. | PASS; every named retired surface has zero references in its owning scope, the inventory contains retained exports only, and its negative fixture rejects additions. |
| Steady-state files are cohesive without a package split or selector loss. | File inventory, diff review, format, 52/52 reconciliation, and owner suites. | PASS; responsibility files replace the six aggregate files, stale paths are absent, one transaction sequence remains, and all 52 selectors and 34/21 rows pass unchanged. |
| Public and persisted behavior remains unchanged. | HTTP/OpenAPI, authorization, idempotency, history, rollback, projection, Collaboration, bundle, frontend, and browser evidence. | PASS; all affected owner suites, 390 frontend units, frontend type/import gates, 58 webserver-backed, 34 stateful, 12 accessibility, and 12 visual rows passed without frontend or golden changes; full and release graphs are green. |
| Generated outputs derive only from authored inputs and no lockfile is hand-edited. | Generation drift, generated policy, JSON shape, Harness, and changed-path review. | PASS; `tools/execution_topology_render_index.json` is the only generated delta and came from `make generate`; final drift, policy, shape, and release gates pass; no lockfile changed. |
| Developer and release validation is complete. | Build, finalization, `make check`, `make release-check`, final Markdown lint, and retained roots. | PASS; build passed 7/7, retained-run finalization validated the current 642/642 full check with zero updates, release passed 799/799, and the post-handoff Markdown checkpoint passed at `.cartulary/test-results/20260822T092800Z-p1211872`. |

Domain vocabulary is unchanged. `docs/domain.md` was not edited by EPR-S00
through EPR-S06.

## 23. ESI Scope, Baseline, and Planning Posture

ESI is the Entities Source Integrity iteration. Sections 1 through 13 remain
the completed history for the original remediation, and Sections 14 through
22 remain the completed EPR history. ESI does not reopen their closed owner
topology, constructor, export-closure, cohesion, or final-validation work.

This tracker remains an execution-support artifact. Adopted Core owners and
`docs/decisions/entities-module-boundary.md` continue to outrank it. ESI-S01
must adopt every new behavior named here before an implementation slice may
rely on it. No runtime, test, contract, generator, or verification code may
read this tracker as an executable requirement source.

ESI-S00 was the documentation-only rebaseline requested on 2026-08-22. The
later 2026-08-22 implementation directive authorizes the remaining serial
slices. `docs/domain.md` remains unchanged because ESI introduces no
vocabulary or owner-navigation change.

### 23.1 Planning baseline

| Baseline fact | ESI-S00 record |
| --- | --- |
| Commit | `4998e011fd1a70cbefb02fec136723d8f59505a2` |
| Commit subject | `Entities Production-Readiness` |
| Repository posture | Clean worktree before this tracker edit |
| Date | 2026-08-22 |
| Migration posture | Immutable boundary 29; repository migration head 35 |
| Ordinary-build top-level Entities tests | 52 discovered and exactly selected |
| Entities verification rows | 34 owner rows; 21 service-backed rows |
| Completed predecessor | EPR-S00 through EPR-S06; EPR-G01 through EPR-G07 closed |
| Current authorization | ESI-S01 through ESI-S07 authorized and complete; no ESI successor slice remains |

Historical green results establish the planning baseline only. They do not
prove any ESI implementation work. Each implementation slice must reconcile
the live commit, worktree, migration head, selector count, owner routing, and
owner documents before changing an artifact.

### 23.2 Iteration objective

ESI makes active Host and Identity exact-match identity authoritative at the
database boundary, closes source-table invariants, makes concurrent and
lifecycle mutation deterministic, and validates every declared Entities
Incident Bundle invariant. The target is one indexed, rebuildable active-claim
model shared by all mutation paths without a generic cross-module identity
service or a package split.

Existing successful public behavior remains stable. ESI intentionally rejects
ambiguous multi-identifier input and invalid persisted source state instead of
preserving behavior that can create competing active identities. It adds only
the owner-adopted restore and rollback conflict vocabulary needed to fail those
paths safely.

### 23.3 Sequence and authority

The sequence is:

`ESI-S00 -> ESI-S01 -> ESI-S02 -> ESI-S03 -> ESI-S04 -> ESI-S05 -> ESI-S06 -> ESI-S07`

Only one slice may be active or validating. A successor remains ineligible
until its predecessor completes focused validation, updates this tracker, and
passes the checkpoint gate in Section 26. ESI-S00 through ESI-S07 completed
in that order; Section 29 records every checkpoint and the final handoff.

## 24. ESI Requirements and Gap Register

### 24.1 Requirements

| Requirement | Planned outcome | Primary evidence |
| --- | --- | --- |
| ESI-REQ-001 | Preserve completed Sections 1 through 22 while making Sections 23 through 31 the sole ESI ledger. | Tracker review and Markdown lint |
| ESI-REQ-002 | Execute ESI serially and checkpoint the tracker after every slice. | Section 29 ledger |
| ESI-REQ-003 | Make active exact-match identity database-enforced, indexed, concurrent, and tombstone-aware. | Owner requirements, migrations 36 and 37, claim and concurrency tests |
| ESI-REQ-004 | Reject submitted exact identifiers that resolve to different active records; use precedence only when all matches converge. | Characterization and mutation tests |
| ESI-REQ-005 | Enforce the adopted Host, Identity, alias, preserved-identifier, mention, envelope, and merge-lineage invariants at the source boundary. | Migration preflight, constraints, triggers, and negative fixtures |
| ESI-REQ-006 | Maintain active claims through every creation, mutation, lifecycle, merge, rollback, portability, and recovery path. | Owner and cross-owner integration tests |
| ESI-REQ-007 | Keep active claims derived, rebuildable, least-privileged, and absent from portable and backup-domain content. | Recovery catalog, bundle inventory, privileges, and rebuild evidence |
| ESI-REQ-008 | Replace table-shaped Entities bundle serialization with explicit version-2 portable rows and close all eight declared invariant IDs. | Typed codec and invariant-closure suite |
| ESI-REQ-009 | Preserve routes, operation IDs, request/success shapes, security order, idempotent replay, history, projections, Collaboration, and valid bundle bytes. | Characterization, owner, frontend, browser, and release evidence |
| ESI-REQ-010 | Reach 56/56 Entities selectors, 38 owner rows, and 25 service-backed rows without selector loss. | Catalog and exact accounting |
| ESI-REQ-011 | Complete ESI only after migration, recovery, portability, owner, browser, build, security, and release gates pass. | ESI-S07 handoff |

### 24.2 Gap decisions

| Gap | Areas and remediation | Rationale | Expected long-term benefit | Compatibility or migration impact | Risk if unresolved | Validation criteria |
| --- | --- | --- | --- | --- | --- | --- |
| ESI-G01 — The tracker has no successor posture | Documentation: retain completed history, update the opening posture, and append the ESI gaps, target design, serial slices, validation, ledger, risks, and completion criteria. | Leaving EPR as the terminal posture hides known integrity work; rewriting it would destroy auditability. | One current serial ledger with intact predecessor evidence. | Documentation only in ESI-S00. | Later sessions can infer authority or completion from historical green results. | Header and Sections 23 through 31 agree; only the tracker changes; scoped diff and Markdown lint pass. |
| ESI-G02 — Exact matching is neither authoritative nor concurrency-safe | Specification, implementation, migrations, recovery, and tests: replace incident-wide Go scans with a database-enforced active identifier claim model and deterministic advisory locking. Exclude deleted and merged records. Reject cross-record submitted matches. | Application scans cannot close an empty-incident race or prevent canonical and preserved identifiers from diverging. | Indexed lookup, one concurrency model, predictable expansion to later identity classes, and fewer mutation-time races. | Intentional behavior change for ambiguous multi-key requests; clean databases receive a forward migration. | Duplicate active identities, deleted-record reuse errors, request-time races, and O(N) growth remain possible. | Concurrent creates converge; every mutation path maintains claims; cross-record input returns `entity_match_conflict`; query plans use the claim key. |
| ESI-G03 — Source tables under-enforce adopted invariants | Specification, migrations, tests, and generated ownership projections: add preflight, constraints, and triggers for envelope/type/incident ownership, identifier classes and normalization, mention vocabulary and resolution tuples, and merge lineage. | Source invariants must survive direct SQL, import, recovery, and future callers rather than depending on individual Go paths. | Durable local consistency and simpler application reasoning. | Migration blocks invalid existing rows with safe aggregate findings; it never guesses or repairs them. | Invalid tuples can enter through non-HTTP paths and later break portability, resolution, merge, or recovery. | Clean install and valid upgrade pass; each invalid fixture fails before partial schema change; ownership, privileges, Down, and Up evidence pass. |
| ESI-G04 — Lifecycle and recovery do not own active identity consistently | Specification, Revisions port, implementation, recovery, and tests: prepare both delete and restore transitions, maintain claims through delete, restore, rollback, merge, import, bundle import, and recovery, and add safe public conflict vocabulary. | Identity ownership changes with envelope lifecycle; delete-only preconditions cannot make restore or rollback safe. | Atomic lifecycle semantics and deterministic reconstruction after restore. | Adds `record_restore_blocked` and one rollback reason; successful shapes and routes remain unchanged. | Restore or rollback can recreate a competing active identity or expose a raw database failure. | Delete releases claims; restore reacquires or fails safely; rollback, merge, bundle import, and recovery reproduce the exact claim set with no partial effects. |
| ESI-G05 — Incident Bundle validation declares more than it enforces | Specification, contracts, implementation, and tests: use typed fixed version-2 rows, row-local prepare validation, post-apply cross-row validation, and one negative fixture for each declared invariant. | Generic table-shaped JSON couples bundles to future columns and converts owner failures into database accidents. | Stable portable bytes, exact failure attribution, and safe schema evolution. | Valid v2 bytes remain stable; malformed bundles may fail earlier and more precisely. | Invalid portable state can pass preparation, fail with unsafe storage details, or survive until a later subsystem consumes it. | All eight invariant IDs are independently reachable; invalid imports are atomic; claims are absent from the bundle and rebuilt from authoritative rows. |
| ESI-G06 — Merge and mutation ordering cannot safely hand off unique claims | Specification, implementation, and tests: make merge planning read-only, lock the complete identity set in canonical order, release loser claims before carry-forward, and retain one transaction sequence. | Mutating during planning conflicts with a unique claim model and makes lock ownership difficult to prove. | Clear transaction phases, deadlock-resistant ordering, and a stable seam for later merge participants. | Internal Go refactor only; public merge order, errors, history, projections, Collaboration, replay, and rollback remain exact. | Claim uniqueness can turn existing merge helpers into mid-transaction collisions or duplicated sequencing. | Pure-plan tests, lock-order evidence, claim handoff, protected-set revalidation, merge rollback, and exact history/result tests pass. |
| ESI-G07 — Production-readiness evidence does not cover the new integrity boundary | Validation and handoff: run exact reconciliation plus migration, recovery, portability, owner, frontend, browser, build, full-check, security, and release gates. | Focused tests cannot prove every writer, recovery path, generated projection, and hidden consumer moved together. | Reproducible release evidence for later feature work. | None beyond the adopted changes above. | Stale writers, missing triggers, recovery drift, unsafe errors, or hidden compatibility breaks can ship. | Every mandatory ESI-S07 result is `PASS`; all gaps and binary criteria close with current evidence. |

## 25. ESI Adopted Target Design

ESI-S01 must adopt the following target before implementation. The design is
not runtime authority while ESI-S01 remains unauthorized or incomplete.

### 25.1 Active identifier claims

- Add derived table `entity_active_identifier_claims` with columns
  `incident_id`, `entity_type`, `identifier_type`, `normalized_value`, and
  `record_id`.
- Use `(incident_id, entity_type, identifier_type, normalized_value)` as the
  primary or unique claim key and add a record-oriented lookup index.
- Define the claim domain as the union of non-null canonical exact-match fields
  and active preserved identifiers classified `exact_match_reuse`.
- A claim exists only while the Records envelope is non-deleted and the owner
  subtype state is `stub` or `canonical`. Merged and deleted entities hold no
  active claim.
- Keep the table derived and rebuildable. It is absent from Incident Bundle
  files and backup-domain authoritative content.

### 25.2 Normalization and locking

- Add a private versioned SQL normalizer equivalent to
  `fieldnorm.NormalizeIdentifier`: NFC normalization, Unicode-whitespace trim,
  empty rejection, Cc/Cf rejection, lowercase for `aad_device_id`, `fqdn`,
  `hostname`, `aad_object_id`, `upn`, `email`, and `sam_account_name`, and
  uppercase for `sid`.
- Project a versioned golden corpus under `contracts/entities` and prove Go and
  SQL parity for ordinary, Unicode, whitespace, case, control, empty, and
  invalid-class inputs.
- Acquire transaction advisory locks for every normalized identity tuple in
  canonical lexical tuple order before lookup or mutation. Patch, rollback,
  merge, and carry-forward lock the union of old, current, and proposed tuples.
- Exact precedence selects a record only after all supplied matching claims
  resolve to that same record. Matches distributed across records fail with
  the existing `entity_match_conflict` contract.

### 25.3 Schema integrity and migration posture

- Add append-only `00036_entities_source_integrity.sql` for preflight and
  source constraints or triggers.
- Add append-only `00037_entities_active_identifier_claims.sql` for the
  normalizer, claim relation, indexes, `ENABLE ALWAYS` maintenance triggers,
  deterministic rebuild and validation routines, backfill, and grants.
- Assign versions 36 and 37 to source owner `entities` in the authored
  migration-owner map and derive history, schema ownership, and topology only
  through `make generate`.
- If either version is occupied before ESI-S02 begins, mark the slice
  `BLOCKED: migration head drift` and rebaseline the owner plan; do not rename
  or reorder migrations silently.
- Migration preflight reports safe aggregate counts and a remediation hint for
  invalid or duplicate current state. It never auto-merges, discards, or
  chooses a winner.
- Correct Core 01 so version 29 remains the immutable Production DDL Rebaseline
  v2 boundary while repository head becomes 37. No mixed-version writer window
  is required because server startup requires repository-head schema.

### 25.4 Lifecycle and public conflict surface

- Replace Revisions' delete-only `ValidateDeletePreconditionsTx` with the
  consumer-owned `PrepareStateTransitionTx` port method. It receives a closed
  delete/restore transition kind, may acquire owner identity locks, and returns
  structured blocked state without exposing database errors.
- Invoke the port for both delete and restore before envelope state changes.
  Migrate all providers atomically; retain no forwarding method or dual port.
- Add `409 record_restore_blocked` with
  `reason_code=active_entity_identifier_conflict`.
- After ordinary authentication, visibility, role, destructive-lock, replay,
  and row-version gates, safe conflict details may include `entity_type`,
  `identifier_class`, `normalized_value`, and `blocking_record_id`, matching
  the existing merge collision disclosure boundary.
- Extend `rollback_precondition_failed` with
  `active_entity_identifier_conflict` for rollback rekey collisions.
- Preserve every route, operation ID, request body, success body, schema ID,
  field key, event shape, authorization/concealment order, and idempotent replay
  rule.

### 25.5 Portability and recovery

- Replace `to_jsonb(table)` and `jsonb_populate_record` coupling with
  owner-local typed portable rows and explicit version-2 column lists.
- Preserve the exact valid version-2 logical paths, stable identities, row
  ordering, and serialized member set unless S01 characterization proves the
  current table-shaped output contains non-contract columns. Any required byte
  correction returns to owner adoption rather than being inferred.
- Validate row-local facts during prepare and cross-row facts after apply.
  Every declared failure must use exactly one of:
  `entities.source_identity_admitted`, `entities.mentions_observational`,
  `entities.envelope_type_scope`, `entities.resolution_merge_coherent`,
  `entities.alias_identifier_normalized`,
  `entities.alias_identifier_classified`,
  `entities.alias_identifier_unique`, or
  `entities.alias_identifier_same_incident`.
- Register `entities.restore_active_identifier_claims.v1` as a deterministic
  recovery algorithm. Recovery restores authoritative source rows, rebuilds
  claims, validates the result, and publishes no claim table as authoritative
  backup content.

## 26. ESI Execution Policy and Checkpoint Gate

Every implementation slice is atomic. After implementation and focused
validation, and before a successor begins, update this tracker with:

- slice, requirement, gap, and binary-criterion status;
- every authored, moved, generated, and deleted file;
- exact commands, results, result roots, and relevant summary artifacts;
- selector and verification-row counts when changed;
- public, persisted, internal-interface, migration, and deployment impact;
- generated-file provenance and changed-path review;
- rollback boundary;
- failures, causal attribution, skipped checks, and residual risks; and
- the next eligible slice and its authorization state.

Then run:

1. `git diff --check -- docs/handoffs/entities-module-refactor-tracker.md`
2. `make lint-markdown`

A required check failure, unexplained baseline or selector drift, owner
contradiction, hand-edited generated artifact, undocumented public behavior,
invalid migration data without an adopted disposition, or occupied migration
version marks the active slice `BLOCKED`. Do not begin its successor.

Use public Make targets and current task guidance. Do not invoke direct Go,
frontend, browser, migration, or generator commands as verification evidence.
Generated artifacts are changed only by editing their owner inputs and running
the applicable Make generator.

## 27. ESI Serial Workstreams

| Slice | Workstream | Depends on | Current status | Authorization | Rollback boundary | Exit criterion |
| --- | --- | --- | --- | --- | --- | --- |
| ESI-S00 | Tracker rebaseline | Completed EPR-S06 | COMPLETE | Authorized by the documentation request | Revert the ESI header posture and Sections 23 through 31 only. | Tracker-only diff; scoped diff check and Markdown lint pass. |
| ESI-S01 | Owner adoption and characterization | ESI-S00 checkpoint | COMPLETE | Authorized and completed under the current directive | Revert owner amendments and this checkpoint together. | Owners adopt every new behavior; baseline accounting remains 52/34/21. |
| ESI-S02 | Source-integrity migration | ESI-S01 checkpoint | COMPLETE | Authorized and completed under the current directive | Revert migration 36, owner inputs, generated projections, test, and row together. | Clean/upgrade/invalid/Down/Up evidence passes at 53/35/22. |
| ESI-S03 | Active claims and concurrent matching | ESI-S02 checkpoint | COMPLETE | Authorized and completed under the current directive | Revert migration 37, claim repository, mutation callers, recovery contribution, corpus, tests, rows, and generated projections together. | Claim parity, lifecycle/rebuild, and concurrency pass at 55/37/24. |
| ESI-S04 | Lifecycle, rollback, and merge integration | ESI-S03 checkpoint | COMPLETE | Authorized and completed under the current directive | Revert the Revisions port, providers, public error projection, merge ordering, migration-37 pre-deployment refinements, and tests atomically. | Delete/restore/rollback/merge claim ownership and existing behavior are exact. |
| ESI-S05 | Incident Bundle invariant closure | ESI-S04 checkpoint | COMPLETE | Authorized and completed under the current directive | Revert typed codecs, validation, test, authored row, and generated topology together. | All eight invariant failures and valid round trips pass at 56/38/25. |
| ESI-S06 | Cleanup and exact reconciliation | ESI-S05 checkpoint | COMPLETE | Authorized and completed under the current directive | Revert cleanup without reverting stabilized S05 behavior. | No old scan path, stale index, selector loss, or inventory drift remains. |
| ESI-S07 | Final validation and handoff | ESI-S06 checkpoint | COMPLETE | Authorized and completed under the current directive | Corrections returned to their owning slice, then the complete invalidated matrix restarted; S07 closes evidence and tracker state. | Every mandatory current gate passes and ESI-G01 through ESI-G07 close. |

### 27.1 ESI-S00 — Tracker rebaseline

- Change the opening posture and append Sections 23 through 31.
- Preserve all completed implementation and evidence history.
- Change no product, test, owner, contract, generated, migration, application,
  or domain artifact.
- Exit only after changed-path review, tracker-scoped diff check, and Markdown
  lint pass.

### 27.2 ESI-S01 — Owner adoption and characterization

- Amend Core 02 entity matching, active-state, normalization, conflict, and
  merge carry-forward requirements.
- Amend Core 01 restore errors, rollback reasons, Incident Bundle version-2
  owner behavior, recovery semantics, and immutable-boundary/current-head
  wording.
- Add Core 04 AC-559 as the Entity analogue to AC-533 and include AC-559 in
  the base-profile acceptance manifest.
- Characterize deleted-record exclusion, cross-class multi-record conflict,
  restore collision, rollback rekey collision, merge claim handoff, current
  valid bundle bytes, failure ordering, replay, and atomicity through subtests
  under existing top-level selectors.
- Run catalog, Harness, backend boundary, focused Entities, service-backed
  Entities, Revisions, Incident Bundles, and migration-owner guidance selected
  at execution time.
- Exit with adopted owners and exact 52 tests, 34 owner rows, and 21
  service-backed rows.

### 27.3 ESI-S02 — Source-integrity migration

- Add migration 36 with safe preflight and source enforcement for envelope
  type and incident ownership, valid identifier classes and normalization,
  mention origin/status/resolution tuples, and same-incident non-self merge
  lineage.
- Do not broaden seed-mention semantics unless S01 adopts the missing tuple
  relationship explicitly.
- Add `TestEntitySourceIntegrityMigration_Integration` and one authored
  service-backed Entities row derived through `make author-test-row-id`.
- Update the authored migration-owner input and run `make generate`; do not
  hand-edit history, schema-ownership, or execution-topology projections.
- Validate empty installation, valid 35-to-36 upgrade, every rejected invalid
  fixture, no partial DDL, disposable Down and Up, least privilege, schema
  ownership, migration drift, catalog, Harness, and affected owner slices.
- Exit at 53/53 tests, 35 owner rows, and 22 service-backed rows.

### 27.4 ESI-S03 — Active claims and concurrent matching

- Add migration 37, the SQL normalizer, claim table, indexes, backfill,
  `ENABLE ALWAYS` triggers, rebuild/validation functions, and minimum grants.
- Add the versioned normalization corpus and the Entities recovery
  contribution for `entities.restore_active_identifier_claims.v1`.
- Replace incident-wide exact-match scans with indexed claims and remove the
  redundant second create-path match.
- Apply deterministic tuple locking to create, upsert, patch, import, and
  clipboard mutation paths.
- Add
  `TestEntityActiveIdentifierClaimsFollowLifecycleAndRebuild_Integration` and
  `TestConcurrentEntityExactMatchConverges_Integration`, with two authored
  service-backed Entities rows.
- Validate Go/SQL parity, valid and invalid migration backfill, concurrent
  convergence, deleted/merged exclusion, create/upsert/import/clipboard paths,
  trigger order, recovery rebuild, corruption detection, privileges, query
  plans, drift, and affected owner suites.
- Exit at 55/55 tests, 37 owner rows, and 24 service-backed rows.

### 27.5 ESI-S04 — Lifecycle, rollback, and merge integration

- Replace the Revisions delete-only source method with
  `PrepareStateTransitionTx`, define a closed delete/restore transition kind,
  and migrate every source provider and catalog test atomically.
- Acquire identity locks and prepare owner state before delete or restore
  changes the Records envelope. Release claims on delete; reacquire them or
  return `record_restore_blocked` on restore.
- Make row and collection rollback lock and validate every resulting exact
  identifier and use the adopted rollback conflict reason.
- Make merge planning read-only. Lock survivor, loser, protected records, and
  the complete old/new/carry identifier set; release loser claims before
  survivor carry-forward; then execute the existing single effect sequence.
- Preserve destructive-lock ordering, protected-set revalidation, history,
  projections, Collaboration intents, replay, public merge errors, and
  transaction rollback.
- Add subtests under existing selectors and do not change 55/37/24 accounting.

### 27.6 ESI-S05 — Incident Bundle invariant closure

- Introduce typed owner-local version-2 portable rows with explicit field
  lists and deterministic ordering for the five existing Entities paths.
- Validate row-local invariants during prepare and cross-row invariants after
  apply, using only the eight declared Entities invariant IDs.
- Add one negative fixture per invariant and prove every rejection leaves no
  partial imported state or unsafe database detail.
- Prove valid version-2 round trips retain the characterized bytes, active
  claims are not serialized, and import triggers reconstruct the expected
  claims.
- Add `TestEntityIncidentBundleInvariantClosure_Integration` and one authored
  service-backed Entities row.
- Validate Entities and Incident Bundles focused/service-backed slices,
  portability catalogs, Revisions validation, recovery interaction, catalog,
  Harness, generation drift, generated policy, and JSON shape.
- Exit at 56/56 tests, 38 owner rows, and 25 service-backed rows.

### 27.7 ESI-S06 — Cleanup and exact reconciliation

- Remove superseded full-scan matching helpers only after exact zero-reference
  evidence.
- Remove now-redundant indexes only when current query-plan evidence proves no
  production or recovery query uses them. Otherwise retain and record the live
  role rather than optimizing speculatively.
- Reconcile exact exports, production imports, files, selectors, authored
  rows, migrations, schema ownership, recovery catalog, Incident Bundle
  catalog, and generated topology.
- Do not split packages, create a shared identity kernel, or retain forwarding
  wrappers, dual paths, or compatibility indexes without a live query.
- Exit with exact 56/56 selectors, 38 owner rows, 25 service-backed rows, and
  no stale path or duplicate implementation.

### 27.8 ESI-S07 — Final validation and handoff

Run the Section 28 final matrix in order. Record all current results, result
roots, causal attribution, generated provenance, compatibility conclusions,
rollback posture, skipped checks, and residual risks. A failed mandatory gate
prevents completion and returns corrections to the owning earlier slice.

## 28. ESI Validation Matrix

### 28.1 Test accounting

| Slice | Added top-level Entities test | Expected tests | Expected owner rows | Expected service-backed rows |
| --- | --- | --- | --- | --- |
| ESI-S00 and ESI-S01 | None; characterization uses subtests | 52 | 34 | 21 |
| ESI-S02 | `TestEntitySourceIntegrityMigration_Integration` | 53 | 35 | 22 |
| ESI-S03 | `TestEntityActiveIdentifierClaimsFollowLifecycleAndRebuild_Integration`; `TestConcurrentEntityExactMatchConverges_Integration` | 55 | 37 | 24 |
| ESI-S04 | None; lifecycle and merge cases use subtests | 55 | 37 | 24 |
| ESI-S05 through ESI-S07 | `TestEntityIncidentBundleInvariantClosure_Integration` | 56 | 38 | 25 |

Every new row ID is derived with `make author-test-row-id`. Update the authored
owner manifest, then run `make generate`. Generated topology is never edited
directly.

### 28.2 Required behavioral scenarios

- Concurrent identical creates converge on one active entity and one claim.
- Submitted identifiers that map to different active records fail without
  mutation, while same-record multi-key matches retain precedence behavior.
- Deleted and merged records are excluded; delete releases claims; restore
  reacquires them or returns the exact safe conflict.
- Patch, upsert, import, clipboard, row rollback, collection rollback, merge,
  and bundle import rekey claims atomically.
- Go and SQL normalization agree for NFC, case, Unicode whitespace, Cc/Cf,
  empty, unknown-class, canonical, and preserved-identifier fixtures.
- Empty install, valid 35-to-37 upgrade, invalid-data preflight, Down, Up,
  trigger recovery order, rebuild, corruption detection, ownership, and
  privileges pass.
- Every Entities Incident Bundle invariant has an exact negative fixture,
  deterministic failure ID, and no partial-import effect.
- Existing valid routes, success payloads, authorization and concealment
  ordering, idempotent replay, history, projections, Collaboration, and
  version-2 bundle bytes remain exact.

### 28.3 Per-slice minimums

| Slice | Required minimum validation |
| --- | --- |
| ESI-S00 | Tracker-only changed-path review; tracker-scoped diff check; `make lint-markdown` |
| ESI-S01 | Owner-document reconciliation; catalog; Harness; backend boundary; focused/service-backed Entities and affected owner characterization |
| ESI-S02 | Format; migration input/drift; generation and policy drift; schema ownership; catalog; Harness; migration, Entities, and app startup slices |
| ESI-S03 | Format; generation and drift; normalization parity; recovery catalog; query plans; focused/service-backed Entities, Imports, Workbook, Timeline, Assessments, Recovery, and app server |
| ESI-S04 | Format; Revisions port closure; public error and OpenAPI checks; focused/service-backed Entities, Revisions, Timeline, Assessments, Workbook, Imports, and app server; merge/rollback cases |
| ESI-S05 | Format; bundle catalog and byte reconciliation; all invariant fixtures; Entities, Incident Bundles, Revisions, and Recovery slices; generated policy and JSON shape |
| ESI-S06 | Exact symbol, file, migration, schema, recovery, bundle, export, selector, row, and generated reconciliation; catalog; Harness; boundary; affected owner slices |
| ESI-S07 | Complete ordered final matrix below |

Use `make task-guide ROLE=module-author OWNER=<owner-id>` and explanation
targets immediately before execution to select exact affected rows.

### 28.4 Ordered final matrix

1. Reconcile exports, retired scans, production imports, files, migration head
   37, schema objects, recovery catalog, bundle catalog, 56/56 selectors, 38
   owner rows, and 25 service-backed rows.
2. `make test-catalog-check`
3. `make harness-contract`
4. `make backend-module-boundary-check`
5. `make migration-drift`
6. `make generate-drift`
7. `make generated-artifact-policy-check`
8. `make json-shape-check`
9. Run current focused and service-backed task-guidance selections for
   Entities, Revisions, Records, Timeline, Assessments, Workbook, Imports,
   Incident Bundles, Recovery, Database Migrations, and app server.
10. `make frontend-unit`
11. `make frontend-typecheck`
12. `make frontend-import-boundary-check`
13. `make browser-e2e-webserver-backed`
14. `make browser-e2e-stateful`
15. `make browser-e2e-a11y`
16. `make browser-e2e-visual`; any visual change is a failure and no golden is
    refreshed.
17. `make build`
18. `make agent-finalize`, supplying `RESULTS_DIR` only for a qualifying
    successful current full warm-check root.
19. `make check`
20. `make release-check`
21. Final changed-path review, tracker-scoped diff check, and
    `make lint-markdown`.

## 29. ESI Tracker and Checkpoint Ledger

### 29.1 Work tracker

| Work item | Description | Slice | Status | Dependency | Evidence | Completion condition |
| --- | --- | --- | --- | --- | --- | --- |
| ESI-001 | Rebaseline the tracker and establish the ESI plan. | ESI-S00 | DONE | Completed EPR-S06 | Baseline commit, migration head, counts, authority, target design, workstreams, validation, and handoff rules are recorded in Sections 23 through 31. | Only the tracker changes; scoped diff and Markdown gates pass. |
| ESI-002 | Adopt source-integrity, exact-match, lifecycle, portability, recovery, and public conflict behavior. | ESI-S01 | DONE | ESI-001 | Core 01 owns lifecycle, recovery, migration-head, exact v2 portability, and conflict behavior; Core 02 owns normalization, claims, convergence, and merge sequencing; Core 04 owns AC-559. Catalog, Harness, boundary, and Entities focused/service-backed characterization passed. | Core 01, Core 02, and Core 04 own every behavior before implementation begins. |
| ESI-003 | Enforce source-table integrity through migration 36. | ESI-S02 | DONE | ESI-002 | Migration 36 preflights existing source rows and enforces envelope, lifecycle, child ownership, identifier, normalization, mention, and resolution invariants; clean, upgrade, invalid, atomic, Down/Up, privilege, ownership, catalog, and owner evidence passes at 53/35/22. | Valid install/upgrade and every invalid preflight case pass at 53/35/22. |
| ESI-004 | Add derived claims and concurrent indexed matching through migration 37. | ESI-S03 | DONE | ESI-003 | Migration 37, the v1 corpus, indexed repository, tuple locks, immediate `ENABLE ALWAYS` triggers, least-privilege recovery and handoff routines, recovery dispatch, concurrency and corruption fixtures, and current affected-owner evidence pass at 55/37/24. | Parity, claim lifecycle/rebuild, and concurrent convergence pass at 55/37/24. |
| ESI-005 | Integrate claims with delete, restore, rollback, and merge. | ESI-S04 | DONE | ESI-004 | The lifecycle port is atomically replaced; delete/restore, row and collection rollback, patch, and merge use ordered claim preparation; safe public conflicts, pure merge planning, protected-set revalidation, handoff, replay, and rollback evidence pass at 55/37/24. | Lifecycle and merge ownership are atomic with public behavior exact. |
| ESI-006 | Close all declared Entities Incident Bundle invariants. | ESI-S05 | DONE | ESI-005 | Explicit typed v2 rows, exact member decoding and fixed-column SQL replace table-shaped serialization; all eight closed invariants, atomicity, claim reconstruction, and valid byte-stable round trips pass at 56/38/25. | Eight negative invariants and valid v2 round trips pass at 56/38/25. |
| ESI-007 | Remove superseded paths and reconcile the final surface. | ESI-S06 | DONE | ESI-006 | Dead incident-wide scans and duplicate match calls are absent; nine lookup-only indexes retire in migration 37 with exact Down recreation; file, selector, migration, schema, recovery, bundle, export, and generated inventories agree at 56/38/25. | Exact symbol, file, selector, migration, recovery, bundle, and generated inventories agree. |
| ESI-008 | Run final gates and publish the ESI handoff. | ESI-S07 | DONE | ESI-007 | The ordered matrix passes on the final corrected state; Section 29.10 records compatibility, rollback, failures, generated provenance, and current result roots. | Every mandatory current gate passes and ESI-G01 through ESI-G07 close. |

### 29.2 Session handoff log

| Time | Session | Current state | Files inspected or touched | Commands and results | Compatibility and rollback | Blockers and next action |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-08-22 | Codex / ESI-S00 planning and rebaseline | ESI plan recorded; ESI-S00 complete; ESI-S01 ready but not authorized | Inspected the completed tracker, Core owner rules, Entities matching, source schema, portability, recovery, Revisions lifecycle port, migration catalog, test routing, and repository baseline; touched only this tracker | Read-only repository inspection established commit `4998e011fd1a70cbefb02fec136723d8f59505a2`, migration head 35, clean pre-edit worktree, and 52/34/21 accounting; tracker-scoped `git diff --check` passed; `make lint-markdown` passed at `.cartulary/test-results/20260822T143308Z-p1279330` | Documentation only; no product, test, owner, contract, generated, migration, application, or Domain artifact changed; rollback removes the ESI header posture and Sections 23 through 31 | No implementation blocker established; ESI-S01 requires a later explicit directive |
| 2026-08-22 | Codex / ESI-S01 owner adoption and characterization | ESI-S01 complete; ESI-S02 is the next eligible authorized slice | Touched `docs/spec/01_architecture_storage_and_view_contracts.md`, `docs/spec/02_domain_model_schema_and_history.md`, `docs/spec/04_security_deployment_and_conformance.md`, and this tracker; inspected the current catalog, Harness, boundary, and Entities characterization | `make test-catalog-check` passed; `make harness-contract` passed at `.cartulary/test-results/20260822T151401Z-p1293270`; `make backend-module-boundary-check` passed at `.cartulary/test-results/20260822T151414Z-p1293782`; `make test-slice OWNER=module.entities` passed 38/38 at `.cartulary/test-results/20260822T151416Z-p1294133`; `make service-backed-test-slice OWNER=module.entities` passed 29/29 at `.cartulary/test-results/20260822T151600Z-p1347198`; tracker diff check and `make lint-markdown` passed at `.cartulary/test-results/20260822T152107Z-p1401031`; selector and authored-row accounting remains 52/34/21 | Specification-only product posture: existing successful public and persisted behavior is retained except for the newly adopted ambiguous-match, restore, and rollback conflicts; no runtime, schema, migration, generated, test-routing, or Domain artifact changed; rollback reverts the three owner-document changes and this checkpoint together | No skipped mandatory check or residual S01 blocker; ESI-S02 is authorized and eligible after the checkpoint below |
| 2026-08-22 | Codex / ESI-S02 source-integrity migration | ESI-S02 and ESI-G03 complete; ESI-S03 is the next eligible authorized slice | Added `db/migrations/00036_entities_source_integrity.sql` and `internal/modules/entities/source_integrity_migration_test.go`; changed the Host/Identity delete/restore provider, merge and resolution fixtures, shared Entities test support, Timeline mention-origin writer and fixtures, migration catalog generator and source-hash tests, authored Entities test-family row, Harness fixture-count assertion, generated execution-topology, migration-history, and schema-ownership projections, and this tracker; no file was moved or deleted and `docs/domain.md` is unchanged | Final format passed at `.cartulary/test-results/20260822T155258Z-p1692842`; generation passed at `.cartulary/test-results/20260822T155302Z-p1696548`; Entities focused and service-backed passed 38/38 and 29/29 at `.cartulary/test-results/20260822T154322Z-p1537111` and `.cartulary/test-results/20260822T154514Z-p1591313`; Database Migrations passed 9/9 and 4/4 at `.cartulary/test-results/20260822T155311Z-p1699417` and `.cartulary/test-results/20260822T155416Z-p1715337`; app server passed 24/24 and 17/17 at `.cartulary/test-results/20260822T155518Z-p1730811` and `.cartulary/test-results/20260822T155613Z-p1770850`; final migration drift, generation drift, generated policy, Harness, and boundary passed at `.cartulary/test-results/20260822T155712Z-p1809996`, `.cartulary/test-results/20260822T155719Z-p1812812`, `.cartulary/test-results/20260822T155727Z-p1815698`, `.cartulary/test-results/20260822T155730Z-p1816486`, and `.cartulary/test-results/20260822T155743Z-p1817044`; catalog passed and exact accounting is 53/53 selectors, 35 owner rows, and 22 service-backed rows | Migration 36 is an append-only persisted boundary: invalid pre-existing data blocks atomically with aggregate counts and an operator remediation hint; successful public routes and responses do not change. Internal fixture and Timeline-origin corrections project already adopted vocabulary and retained envelopes. Rollback before deployment reverts migration, owner input, generator support, implementation/test changes, authored row, generated projections, and checkpoint together; the disposable Down/Up path is tested, while a deployed migration is repaired forward | Initial format, generator-security, SQL-cast, and PgError-detail failures were corrected before the dedicated selector passed. The first broad Entities run exposed stale envelope fixtures and `interactive_cell`; the first Harness run exposed its pinned dedicated-row count. Database Migrations then exposed constraint-trigger projection and migration-source-hash fixtures; both were corrected and all invalidated gates rerun. No mandatory check is skipped and no residual S02 blocker remains; ESI-S03 is authorized and eligible after the checkpoint below |
| 2026-08-22 | Codex / ESI-S03 active claims and concurrent matching | ESI-S03 and the active-claim portion of ESI-G02 and ESI-G04 complete; ESI-S04 is the next eligible authorized slice | Added the two `contracts/entities/identifier-normalization-corpus.v1*` inputs, `db/migrations/00037_entities_active_identifier_claims.sql`, `internal/gen/contractentities/artifacts_gen.go`, and `internal/modules/entities/active_identifier_claims_integration_test.go`; changed the contract and Recovery indexes/fixtures/schema, Recovery assembly/catalog/dispatcher and generated projections, migration catalog generator and projections, `fieldnorm`, Host/Identity exact matching, patch/source loading, merge claim lookup/handoff, Entities characterization, Workbook and Assessment envelope fixtures, the Entities authored rows, Harness count, generated topology, and this tracker. No file was moved or deleted; `docs/domain.md` and lockfiles are unchanged | Final format passed at `.cartulary/test-results/20260822T174226Z-p3301010`; generation passed at `.cartulary/test-results/20260822T170820Z-p2513557`; dedicated parity/migration and concurrency/lifecycle selectors passed 4/4 at `.cartulary/test-results/20260822T170833Z-p2520085`; current Entities focused and service-backed passed 39/39 and 30/30 at `.cartulary/test-results/20260822T173630Z-p3193562` and `.cartulary/test-results/20260822T173630Z-p3193564`. Imports passed 22/22 and 14/14 at `.cartulary/test-results/20260822T165657Z-p2320421` and `.cartulary/test-results/20260822T165657Z-p2320423`; Workbook 65/65 and 37/37 at `.cartulary/test-results/20260822T171401Z-p2632609` and `.cartulary/test-results/20260822T171401Z-p2632611`; Timeline 51/51 and 29/29 at `.cartulary/test-results/20260822T171832Z-p2746005` and `.cartulary/test-results/20260822T171832Z-p2746007`; Assessments 27/27 and 18/18 at `.cartulary/test-results/20260822T173127Z-p3024498` and `.cartulary/test-results/20260822T173128Z-p3024500`; current Recovery 24/24 and 19/19 at `.cartulary/test-results/20260822T174230Z-p3304756` and `.cartulary/test-results/20260822T174230Z-p3304758`; app server 24/24 and 17/17 at `.cartulary/test-results/20260822T173326Z-p3107497` and `.cartulary/test-results/20260822T173326Z-p3107499`. Migration drift, generation drift, generated policy, Harness, and boundary passed at `.cartulary/test-results/20260822T173550Z-p3187022`, `.cartulary/test-results/20260822T173550Z-p3187014`, `.cartulary/test-results/20260822T173550Z-p3187016`, `.cartulary/test-results/20260822T173550Z-p3187088`, and `.cartulary/test-results/20260822T173550Z-p3187078`; catalog passed and exact accounting is 55/55 selectors, 37 owner rows, and 24 service-backed rows | Migration 37 is the append-only persisted/deployment boundary. Claims are derived and excluded from authoritative backup rows, Recovery has only SELECT/TRUNCATE plus rebuild/validate execution, runtime has SELECT and the narrow loser-handoff routine but no table DML, and restore rebuild/validation is atomic. Ambiguous multi-key input now returns the adopted conflict; successful APIs remain exact. This pre-production repository adds no ESI compatibility generation for nonexistent old backups; the two pre-existing historical Recovery generations remain outside ESI scope. Rollback before deployment reverts migration 37, corpus, implementation, recovery inputs, tests, rows, generated projections, and checkpoint together; deployed schema is repaired forward | Deferred claim triggers first hid same-transaction writes and collided at merge commit; immediate triggers plus explicit claim handoff and record-column filtering resolved the cause. Recovery reset exposed missing derived-table TRUNCATE privilege. A bulk Timeline fixture then exposed a global per-row projection with quadratic behavior; record-scoped refresh completed the same graph. Stale Workbook and Assessment direct-SQL fixtures were corrected to mirror Records exactly. The production Recovery dispatcher was found and corrected from no-op to rebuild plus validate. All invalidated gates were rerun; no mandatory check is skipped and S04 owns restore/rollback public conflicts, complete tuple locking, pure merge planning, and failure-boundary coverage |
| 2026-08-22 | Codex / ESI-S04 lifecycle, rollback, and merge integration | ESI-S04, ESI-G04, and ESI-G06 complete; ESI-S05 is the next eligible authorized slice | Added `internal/modules/revisions/rollback_identifier_claims.go` and `internal/modules/entities/hostidentity/rollbackprovider/claims.go`. Changed `contracts/errors/index.json`, `contracts/otel/error_class_registry.json`, Core 01, migration 37, all nine delete/restore providers, Revisions lifecycle/rollback contracts, stores, coordinators, error projection, route tests and fixtures, Host/Identity delete/restore, patch, merge and rollback providers, merge planning/apply files, Entities boundary and merge tests, shared Assessment, Workbook, Import, Timeline, and Revisions fixtures, migration characterization/policy tests, source hashes, contract-family and schema-ownership validators, and generated error, schema-ownership, migration-history, and topology projections. No file was moved or deleted; `docs/domain.md` and lockfiles are unchanged | Final format passed at `.cartulary/test-results/20260822T192145Z-p765178`; final generation passed at `.cartulary/test-results/20260822T191934Z-p744868`. Revisions passed 27/27 and 20/20 at `.cartulary/test-results/20260822T191028Z-p519787` and `.cartulary/test-results/20260822T191143Z-p566355`; Entities passed 39/39 and 30/30 at `.cartulary/test-results/20260822T191249Z-p610242` and `.cartulary/test-results/20260822T191444Z-p664286`, with the post-index active-claims selector passing 3/3 at `.cartulary/test-results/20260822T192710Z-p904216`; Timeline passed 51/51 and 29/29 at `.cartulary/test-results/20260822T192803Z-p919698` and `.cartulary/test-results/20260822T193247Z-p978636`; Assessments passed 27/27 and 18/18 at `.cartulary/test-results/20260822T193729Z-p1035423` and `.cartulary/test-results/20260822T193729Z-p1035425`; Workbook passed 65/65 and 37/37 at `.cartulary/test-results/20260822T193926Z-p1118820` and `.cartulary/test-results/20260822T193926Z-p1118822`; Imports passed 22/22 and 14/14 at `.cartulary/test-results/20260822T194840Z-p1318488` and `.cartulary/test-results/20260822T194725Z-p1277289`; app server passed 24/24 and 17/17 at `.cartulary/test-results/20260822T194956Z-p1359710` and `.cartulary/test-results/20260822T195408Z-p1439112`; Recovery passed 24/24 and 19/19 at `.cartulary/test-results/20260822T192415Z-p800626` and `.cartulary/test-results/20260822T192535Z-p852860`; Database Migrations passed 9/9 and 4/4 at `.cartulary/test-results/20260822T192152Z-p768958` and `.cartulary/test-results/20260822T192306Z-p785058`; OpenAPI passed 4/4 at `.cartulary/test-results/20260822T184305Z-p156704`. Final catalog passed; Harness, boundary, migration drift, generation drift, generated policy, and JSON shape passed at `.cartulary/test-results/20260822T195513Z-p1478474`, `.cartulary/test-results/20260822T195513Z-p1478464`, `.cartulary/test-results/20260822T195513Z-p1478408`, `.cartulary/test-results/20260822T195513Z-p1478400`, `.cartulary/test-results/20260822T195513Z-p1478402`, and `.cartulary/test-results/20260822T195513Z-p1478404`; accounting remains exactly 55/55 selectors, 37 owner rows, and 24 service-backed rows | Public behavior adds only the adopted `record_restore_blocked` error and rollback conflict reason; successful routes, operation IDs, schemas, security order, replay, history, projections, Collaboration intents, and merge results remain exact. The internal lifecycle port has no forwarding method or dual interface. Migration 37 was strengthened in place before any deployment because this repository is pre-production and has no old backups; head remains 37 and no compatibility generation was added. Rollback before deployment reverts the S04 port, providers, migration refinement, generated projections, tests, and checkpoint together; after deployment repair is forward | An attempted new top-level collection test violated fixed S04 accounting and was converted to a subtest. JSON shape exposed stale active-family, migration-owner, and constraint-trigger validators; Database Migrations exposed the missing FK support index, Recovery cardinality, approved-definer mirrors, and source hashes; all were corrected and invalidated gates rerun. Two earlier Timeline full runs, the first Imports focused run, and the first app-server service run failed only at service readiness; current complete retries pass, and the isolated Timeline product row also passed at `.cartulary/test-results/20260822T185026Z-p215079`. OpenAPI has no service-backed rows, so its service-backed invocation was non-applicable rather than skipped. No mandatory S04 check or residual S04 blocker remains; S05 owns the typed Incident Bundle codec and eight invariant fixtures |

| 2026-08-22 | Codex / ESI-S05 Incident Bundle invariant closure | ESI-S05, ESI-G05, and ESI-REQ-008 complete; ESI-S06 is the next eligible authorized slice | Added `internal/modules/entities/incident_bundle_portable_model.go`, `incident_bundle_portable_prepare.go`, `incident_bundle_portable_apply.go`, `incident_bundle_portable_encode.go`, `incident_bundle_portable_validate.go`, and `incident_bundle_invariant_closure_integration_test.go`; replaced the generic implementation in `incident_bundle_portability.go` and `incident_bundle_source_port.go`; changed the Entities boundary guard, the Incident Bundles direct-SQL fixtures, the authored Entities test family, PostgreSQL fixture policy, Harness assertion, backend-boundary policy, generated execution topology, and this tracker. No file was moved or deleted; `docs/domain.md`, migrations, public schemas, and lockfiles are unchanged | Final format passed at `.cartulary/test-results/20260822T202721Z-p1634991`; generation passed at `.cartulary/test-results/20260822T202923Z-p1646118`. The new invariant-closure row passed 3/3 focused and service-backed units at `.cartulary/test-results/20260822T202938Z-p1649077` and `.cartulary/test-results/20260822T203022Z-p1664769`. Entities passed 40/40 and 31/31 at `.cartulary/test-results/20260822T203103Z-p1680374` and `.cartulary/test-results/20260822T203304Z-p1734433`; Incident Bundles passed 8/8 and 6/6 at `.cartulary/test-results/20260822T203500Z-p1787691` and `.cartulary/test-results/20260822T203559Z-p1803534`; Revisions passed 27/27 and 20/20 at `.cartulary/test-results/20260822T203744Z-p1819368` and `.cartulary/test-results/20260822T203851Z-p1863893`; Recovery passed 24/24 and 19/19 at `.cartulary/test-results/20260822T203958Z-p1907749` and `.cartulary/test-results/20260822T204117Z-p1959537`. Catalog passed; final Harness, boundary, migration drift, generation drift, generated policy, and JSON shape passed at `.cartulary/test-results/20260822T204322Z-p2012041`, `.cartulary/test-results/20260822T204440Z-p2014062`, `.cartulary/test-results/20260822T204450Z-p2014480`, `.cartulary/test-results/20260822T204500Z-p2017301`, `.cartulary/test-results/20260822T204514Z-p2020274`, and `.cartulary/test-results/20260822T204518Z-p2020724`. Catalog accounting is exactly 56/56 selectors, 38 owner rows, and 25 service-backed rows | Valid version-2 member sets, nullability, timestamps, stable identities, ordering, exports, and re-export bytes remain exact. Malformed bundles now fail earlier through the closed invariant vocabulary; no v3, translation layer, dual codec, or backup compatibility generation exists. Claims remain derived, are never serialized, and are reconstructed from imported source rows. Rollback before deployment reverts the typed codec, dedicated test, authored row, policy inputs, generated topology, fixture corrections, and this checkpoint together; migration head remains 37 | Initial generation identified the missing direct-SQL transaction approval and ordering; initial Incident Bundle runs identified S02-era fixtures that separately mutated retained Records and Host mirrors; initial Harness and boundary runs identified the new authored-row count and exact source-port importer/read policy. Each causal defect was corrected and every invalidated gate rerun. No mandatory S05 check is skipped and no residual S05 blocker remains; ESI-S06 is authorized and eligible after the checkpoint below |

| 2026-08-22 | Codex / ESI-S06 cleanup and exact reconciliation | ESI-S06 and ESI-007 complete; ESI-S07 is the next eligible authorized slice | Changed migration 37, the active-claims migration selector, Host/Identity exact matching, upsert, create, import, and clipboard callers, canonical migration-hash fixtures, generated migration history and schema ownership, and this tracker. Removed three unreferenced incident-wide scan helpers, their private row type, two duplicate snapshot-match helpers, and nine exact-lookup-only indexes from the head schema. No file was added, moved, or deleted; `docs/domain.md`, public contracts, portability catalogs, Recovery catalogs, and lockfiles are unchanged | Final format passed at `.cartulary/test-results/20260822T211929Z-p2640220`; generation passed at `.cartulary/test-results/20260822T205243Z-p2030212`. The migration/index selector passed 3/3 focused and service-backed at `.cartulary/test-results/20260822T205309Z-p2033384` and `.cartulary/test-results/20260822T205401Z-p2049663`. Entities passed 40/40 and 31/31 at `.cartulary/test-results/20260822T205759Z-p2106618` and `.cartulary/test-results/20260822T205956Z-p2161224`; Imports passed 22/22 and 14/14 at `.cartulary/test-results/20260822T210327Z-p2222941` and `.cartulary/test-results/20260822T210441Z-p2265232`; Workbook passed 65/65 and 37/37 at `.cartulary/test-results/20260822T210554Z-p2306506` and `.cartulary/test-results/20260822T210820Z-p2366320`; Revisions passed 27/27 and 20/20 at `.cartulary/test-results/20260822T211037Z-p2422939` and `.cartulary/test-results/20260822T211145Z-p2467885`; Database Migrations passed 9/9 and 4/4 at `.cartulary/test-results/20260822T211435Z-p2528513` and `.cartulary/test-results/20260822T211542Z-p2544514`; app server passed 24/24 and 17/17 at `.cartulary/test-results/20260822T211650Z-p2560171` and `.cartulary/test-results/20260822T211750Z-p2600381`. Catalog passed; Harness, boundary, migration drift, generation drift, generated policy, and JSON shape passed at `.cartulary/test-results/20260822T210219Z-p2215003`, `.cartulary/test-results/20260822T210236Z-p2215636`, `.cartulary/test-results/20260822T210246Z-p2216060`, `.cartulary/test-results/20260822T210257Z-p2218997`, `.cartulary/test-results/20260822T210309Z-p2221956`, and `.cartulary/test-results/20260822T210318Z-p2222447`. Exact reconciliation reports head 37, 56 unique selectors, 38 rows, 25 service-backed rows, five exact bundle paths, derived/excluded/rebuildable claims, zero stale symbols, and zero stale head-schema indexes | Public routes, successes, errors, replay, history, projections, portability v2 bytes, and recovery semantics are unchanged. Each create/upsert/import/clipboard row now matches exactly once and captures its pre-mutation snapshot from that protected result. Migration 37 retires only nine exact-lookup indexes whose production matcher now demonstrably uses the active-claim primary key; display, lineage, foreign-key, and record-local indexes remain. The disposable Down recreates every retired index exactly. This pre-production repository adds no migration generation or old-backup compatibility path. Rollback reverts the S06 implementation, migration refinement, generated projections, tests, hash fixtures, and checkpoint together | The first complete Entities run failed only because the refactor left one unused local variable; the complete retry passed. An initial Database Migrations command used the wrong hyphenated owner ID and was a usage error. The first correctly routed migration-owner run then exposed the two expected source-hash pins; both were updated and all four focused/service-backed units passed on retry. No mandatory S06 check is skipped and no residual S06 blocker remains; ESI-S07 is authorized and eligible after the checkpoint below |
| 2026-08-22 | Codex / ESI-S07 final validation and handoff | ESI-S07, ESI-008, ESI-G01 through ESI-G07, and ESI-REQ-001 through ESI-REQ-011 complete; no successor slice remains | Final validation returned corrections to their owning S01/S03/S06 boundaries. Changed `apps/web/e2e/collaboration.spec.ts`, `db/migrations/00037_entities_active_identifier_claims.sql`, `docs/spec/04_security_deployment_and_conformance.md`, `internal/app/operator/operator_migration_evidence_test.go`, `internal/modules/database_migrations/catalog_characterization_test.go`, `internal/modules/database_migrations/rebaseline_manifest_integration_test.go`, `internal/modules/entities/active_identifier_claims_integration_test.go`, `internal/modules/reporting/reporting_integration_test.go`, `internal/platform/postgres/postgres_role_integration_test.go`, `internal/testutil/pgtest/pgtest_test.go`, `tools/database-migrations/generate-catalog-projections.mjs`, `tools/harness/generated-artifacts/database-contract-drift/schema-object-ownership.mjs`, and this tracker. `make generate` produced `tools/migration_history_manifest.json` and `tools/schema_object_ownership_manifest.json` from migration 37 and the generator. No file was added, moved, or deleted in S07; `docs/domain.md`, lockfiles, public success schemas, bundle version, and visual goldens are unchanged | Final generation passed at `.cartulary/test-results/20260822T224710Z-p200176`. Exact review reports migration head 37, 56/56 top-level selectors, 38 owner rows, 25 service-backed rows, five typed bundle paths, derived/excluded/rebuildable claims, and no retired matcher or head-schema index reference. `make test-catalog-check` passed. Harness, boundary, migration drift, generation drift, generated policy, and JSON shape passed at `.cartulary/test-results/20260822T225323Z-p298913`, `.cartulary/test-results/20260822T225336Z-p299478`, `.cartulary/test-results/20260822T225338Z-p299823`, `.cartulary/test-results/20260822T225346Z-p302759`, `.cartulary/test-results/20260822T225354Z-p305671`, and `.cartulary/test-results/20260822T225356Z-p306096`. Focused/service-backed results were: Entities 40/40 and 31/31 at `.cartulary/test-results/20260822T225403Z-p306582` and `.cartulary/test-results/20260822T225558Z-p360624`; Revisions 27/27 and 20/20 at `.cartulary/test-results/20260822T225757Z-p413954` and `.cartulary/test-results/20260822T225905Z-p458502`; Records 8/8 and 5/5 at `.cartulary/test-results/20260822T230012Z-p502391` and `.cartulary/test-results/20260822T230052Z-p518095`; Timeline 51/51 and 29/29 at `.cartulary/test-results/20260822T230134Z-p533671` and `.cartulary/test-results/20260822T230617Z-p592486`; Assessments 27/27 and 18/18 at `.cartulary/test-results/20260822T231056Z-p649294` and `.cartulary/test-results/20260822T231157Z-p691548`; Workbook 65/65 and 37/37 at `.cartulary/test-results/20260822T231256Z-p733042` and `.cartulary/test-results/20260822T231516Z-p792538`; Imports 22/22 and 14/14 at `.cartulary/test-results/20260822T231730Z-p849205` and `.cartulary/test-results/20260822T231844Z-p891185`; Incident Bundles 8/8 and 6/6 at `.cartulary/test-results/20260822T232011Z-p932921` and `.cartulary/test-results/20260822T232110Z-p948730`; Recovery 24/24 and 19/19 at `.cartulary/test-results/20260822T232208Z-p964377` and `.cartulary/test-results/20260822T232326Z-p1015968`; Database Migrations 9/9 and 4/4 at `.cartulary/test-results/20260822T232445Z-p1067095` and `.cartulary/test-results/20260822T232601Z-p1082761`; app server 24/24 and 17/17 at `.cartulary/test-results/20260822T232721Z-p1098435` and `.cartulary/test-results/20260822T232821Z-p1138264`. Frontend unit, typecheck, and import boundary passed 390/390, 2/2, and 2/2 at `.cartulary/test-results/20260822T232919Z-p1177632`, `.cartulary/test-results/20260822T233049Z-p1213766`, and `.cartulary/test-results/20260822T233059Z-p1214265`. Browser webserver-backed, stateful, accessibility, and visual passed 58/58, 34/34, 12/12, and 12/12 at `.cartulary/test-results/20260822T233107Z-p1214742`, `.cartulary/test-results/20260822T233512Z-p1267291`, `.cartulary/test-results/20260822T233723Z-p1311187`, and `.cartulary/test-results/20260822T233849Z-p1352769`, with no golden refresh. Build and agent-finalize passed 7/7 and 1/1 at `.cartulary/test-results/20260822T234033Z-p1394751` and `.cartulary/test-results/20260822T234054Z-p1427663`; retained-run maintenance was correctly skipped because `RESULTS_DIR` was unset before a qualifying current full warm check existed. `make check` passed 642/642 at `.cartulary/test-results/20260822T234111Z-p1430572`; `make release-check` passed 801/801 at `.cartulary/test-results/20260822T234615Z-p1544218` | Successful APIs, authorization and concealment order, replay, history, projections, Collaboration replay, and valid bundle-v2 bytes remain exact. The only intentional public changes remain the adopted ambiguous-match and restore/rollback conflicts. The final privilege correction gives runtime claim `SELECT`, recovery claim `SELECT`/`TRUNCATE`, and only narrow owner routines; `table_rebuild` now states that least-privilege derived-state posture. Migration 37 remains undeployed head 37 and was refined in place; this pre-production project has no old backups and adds no backup generation, compatibility migration, dual path, or rollback shim. Rollback before deployment reverts each owning slice atomically; after deployment repair is forward | The first webserver-backed run failed 56/58 at `.cartulary/test-results/20260822T215908Z-p3560083` because a redundant pending-queue assertion raced the already-characterized revoked-session shell after a real 401; the isolated Collaboration row and complete retry passed. The first `make check` failed 639/642 at `.cartulary/test-results/20260822T221656Z-p3913935`, exposing a stale Operator evidence digest, a Reporting fixture using retired mention origin `manual`, and incomplete claim-object privilege projection. Follow-on ACL runs identified `PUBLIC` row-type usage, private/application routine misclassification, and the need for recovery-only claim truncation; the owner adopted `table_rebuild`, migration 37 now revokes row-type `PUBLIC` access and grants no direct row DML, generated projections were refreshed, and all invalidated gates restarted. One intermediate Entities run failed 35/40 at `.cartulary/test-results/20260822T224350Z-p139634` only because removing recovery `TRUNCATE` prevented harness reset; the final class preserves the required reset capability without broad DML. An underscored Incident Bundles owner name and an Operator service-backed command with no service-backed row were usage/non-applicable attempts, not skipped mandatory checks. No causal failure, skipped mandatory gate, or residual ESI blocker remains |

### 29.3 ESI-S00 checkpoint

| Check | Required result | Current result |
| --- | --- | --- |
| Changed-path review | Only `docs/handoffs/entities-module-refactor-tracker.md` changed | PASS |
| Tracker-scoped `git diff --check` | No whitespace error | PASS |
| `make lint-markdown` | Current tracker passes the public Markdown gate | PASS at `.cartulary/test-results/20260822T143308Z-p1279330` |
| Authorization transition | ESI-S00 complete; ESI-S01 ready but not authorized | PASS |

### 29.4 ESI-S01 checkpoint

| Check | Required result | Current result |
| --- | --- | --- |
| Owner adoption | Core 01, Core 02, and Core 04 own every ESI implementation behavior | PASS; REQ-01-665 through REQ-01-667, REQ-02-270, and AC-559 adopt the lifecycle, recovery, portability, normalization, claims, concurrency, and merge posture |
| Characterization accounting | No new top-level selector; retain 52/52 selectors, 34 owner rows, and 21 service-backed rows | PASS; accounting is unchanged at 52/52, 34, and 21 |
| Catalog, Harness, and boundary | Current gates pass | PASS; catalog passed, Harness passed at `.cartulary/test-results/20260822T151401Z-p1293270`, and boundary passed at `.cartulary/test-results/20260822T151414Z-p1293782` |
| Focused and service-backed Entities characterization | Current affected selections pass | PASS; focused passed 38/38 at `.cartulary/test-results/20260822T151416Z-p1294133` and service-backed passed 29/29 at `.cartulary/test-results/20260822T151600Z-p1347198` |
| Changed-path and provenance review | Owner documents and tracker only; no generated file or Domain change | PASS; all four files are authored Markdown, `docs/domain.md` is unchanged, and no generated or lockfile path changed |
| Public, persisted, interface, deployment, and migration impact | All changes are owner adoption only | PASS; no implementation interface, public API, persisted state, deployment input, or migration changed in S01 |
| Rollback, failures, skips, and residual risk | Atomic owner/checkpoint rollback; no unexplained failure or skip | PASS; revert the three owner documents and this checkpoint together; S02 remains responsible for enforcing adopted source invariants |
| Authorization transition | ESI-S01 complete; ESI-S02 authorized and ready | PASS |
| Tracker checkpoint | Scoped diff check and current Markdown lint pass before S02 begins | PASS; `make lint-markdown` passed at `.cartulary/test-results/20260822T152107Z-p1401031` |

### 29.5 ESI-S02 checkpoint

| Check | Required result | Current result |
| --- | --- | --- |
| Requirement and gap closure | ESI-REQ-005 and ESI-G03 are enforced at the source boundary | PASS; migration 36 preflights and rejects invalid Host, Identity, alias, preserved-identifier, mention, envelope, lifecycle, and merge-lineage tuples without repair |
| Install, upgrade, invalid data, and rollback | Empty install, valid 35-to-36 upgrade, deterministic invalid fixtures, atomic failure, and disposable Down/Up pass | PASS; `TestEntitySourceIntegrityMigration_Integration` covers the complete migration boundary in the passing Entities owner results |
| Ownership and least privilege | Migration owner, schema objects, routines, triggers, constraints, grants, and runtime/recovery access agree | PASS; owner input assigns 36 to Entities, generated schema ownership models constraint triggers exactly, PUBLIC execute is revoked, and migration/catalog parity passes |
| Accounting | Reach exactly 53/53 selectors, 35 owner rows, and 22 service-backed rows | PASS; selector reconciliation is enforced by the passing boundary row and `make explain-test-owner OWNER=module.entities` reports 35 rows and 22 service-backed rows |
| Affected product owners | Entities, Database Migrations, and app-server startup selections pass | PASS; focused/service-backed results are 38/38 and 29/29, 9/9 and 4/4, and 24/24 and 17/17, with roots recorded in Section 29.2 |
| Generation and repository gates | Format, migration drift, generation drift, generated policy, catalog, Harness, and backend boundary pass | PASS; current result roots are recorded in Section 29.2; `make test-catalog-check` also passed |
| Generated provenance and changed paths | Generated projections come only from owner inputs and `make generate`; no unrelated or Domain change exists | PASS; `tools/execution_topology_render_index.json`, `tools/migration_history_manifest.json`, and `tools/schema_object_ownership_manifest.json` were regenerated; `docs/domain.md` and lockfiles are unchanged; no file was moved or deleted |
| Impact and rollback | Persisted, internal-interface, deployment, and compatibility effects are explicit | PASS; migration 36 is the persisted/deployment boundary, no public success contract changed, no mixed-version window exists, and the slice rollback boundary is recorded in Sections 27 and 29.2 |
| Failures, skips, and residual risk | Every causal failure is resolved; no mandatory gate is skipped | PASS; formatting, migration-policy, fixture, Harness-count, constraint-trigger projection, and source-hash failures are attributed in Section 29.2; all invalidated checks passed and claim enforcement remains explicitly owned by S03 |
| Authorization transition | ESI-S02 complete; ESI-S03 authorized and ready | PASS |
| Tracker checkpoint | Scoped diff check and current Markdown lint pass before S03 begins | PASS; scoped `git diff --check` passed and `make lint-markdown` passed at `.cartulary/test-results/20260822T155849Z-p1817736` |

### 29.6 ESI-S03 checkpoint

| Check | Required result | Current result |
| --- | --- | --- |
| Requirement and gap closure | ESI-REQ-003, ESI-REQ-004, and ESI-REQ-007 active-claim behavior is authoritative; lifecycle/merge completion remains assigned to S04 | PASS; canonical and `exact_match_reuse` identifiers project into one indexed derived claim key, unknown classes fail closed, deleted and merged sources are excluded, and cross-record submitted matches return `entity_match_conflict` |
| Migration, normalization, and rollback | Empty install, valid 36-to-37 upgrade, duplicate preflight, Go/SQL corpus parity, deterministic backfill, and disposable Down/Up pass | PASS; the migration selector covers NFC, Unicode whitespace, case, Cc/Cf, empty and unknown classes, atomic duplicate blocking, object presence, ownership, privileges, and Down/Up |
| Concurrency and indexed lookup | Concurrent identical creates converge; same-record multi-class matches retain precedence; cross-record matches fail atomically; the claim-key plan is indexed | PASS; the concurrency selector proves one record and claim, ordered conflicts, primary-key query-plan use, and no direct runtime table DML |
| Claim maintenance and recovery | Source/envelope triggers, delete/merge exclusion, restore reacquisition, corruption detection, deterministic rebuild, and production Recovery dispatch agree | PASS; immediate `ENABLE ALWAYS` triggers expose same-transaction writes, Recovery owns SELECT/TRUNCATE plus rebuild/validate only, and `entities.restore_active_identifier_claims.v1` rebuilds and validates inside the restore transaction |
| Mutation and handoff coverage | Create, upsert, import, clipboard, patch, and current merge behavior use claim lookup/locks without a competing scan | PASS for S03 paths; patch locks old/current/proposed tuples, create/upsert/import/clipboard share the indexed matcher, and merge collision lookup plus loser release use claim authority. S04 owns rollback locking, pure planning, protected-set revalidation, and exact failure injection |
| Scalability | Trigger refresh is record-scoped and rebuild remains global | PASS; the first Timeline run exposed and was cancelled for a quadratic global per-row projection, then the record-parameterized SQL projection advanced through the large-grid fixture and Timeline passed 51/51 and 29/29 |
| Accounting | Reach exactly 55/55 selectors, 37 owner rows, and 24 service-backed rows | PASS; top-level selector reconciliation is 55, `make explain-test-owner OWNER=module.entities` reports 37 rows and 24 service-backed rows, and catalog/Harness checks pass |
| Affected product owners | Entities, Imports, Workbook, Timeline, Assessments, Recovery, and app-server focused/service-backed selections pass | PASS; current or post-change roots and exact unit counts are recorded in Section 29.2, with the final Entities and Recovery evidence rerun after their last production changes |
| Generation and repository gates | Format, migration drift, generation drift, generated policy, catalog, Harness, and backend boundary pass | PASS; current result roots are recorded in Section 29.2 and `make test-catalog-check` passed |
| Generated provenance and changed paths | Generated files come only from owner inputs and `make generate`; no unrelated, Domain, lockfile, moved, or deleted path exists | PASS; contract, Recovery, SQL, migration-history, schema-ownership, and topology outputs were regenerated from their authored inputs; the changed-path inventory is recorded in Section 29.2 |
| Impact and rollback | Public, persisted, internal-interface, deployment, compatibility, and pre-production backup posture are explicit | PASS; migration 37 and repository head 37 are the persisted/deployment change, successful APIs are stable, ambiguous input is intentionally stricter, no mixed-version window or old-backup ESI generation exists, and rollback is slice-atomic before deployment |
| Failures, skips, and residual risk | Every causal failure is resolved; no mandatory gate is skipped | PASS; trigger visibility/order, recovery privilege, merge collision-fixture, quadratic refresh, stale envelope-fixture, and Recovery-dispatch defects are attributed in Section 29.2; S04 residual work is explicit rather than hidden |
| Authorization transition | ESI-S03 complete; ESI-S04 authorized and ready | PASS |
| Tracker checkpoint | Scoped diff check and current Markdown lint pass before S04 begins | PASS; scoped `git diff --check` passed and `make lint-markdown` passed at `.cartulary/test-results/20260822T174645Z-p3408885` |

### 29.7 ESI-S04 checkpoint

| Check | Required result | Current result |
| --- | --- | --- |
| Requirement and gap closure | ESI-REQ-006 lifecycle, rollback, and merge behavior closes ESI-G04 and ESI-G06 | PASS; delete and restore prepare owner state before envelope mutation, row and exact preserved-identifier rollback batch complete lexical tuple unions, and merge has pure plan, protected-set revalidation, loser release, and one apply sequence |
| Lifecycle port and public errors | One closed delete/restore port with no forwarding method; safe restore and rollback conflicts | PASS; every provider implements `PrepareStateTransitionTx`, the old method has zero references, restore returns exact `record_restore_blocked` detail after ordinary security/version gates, and rollback uses `active_entity_identifier_conflict` without leaking storage errors |
| Claim ordering and atomicity | Patch, delete, restore, row and collection rollback, and merge lock and transfer claims without partial effects | PASS; patch locks old/current/proposed tuples, rollback batches each complete entity-type tuple union before release, exact preserved-identifier inverses nominate their owner record, delete releases, restore reacquires, and failure-injection plus collision tests leave envelope, history, and claims unchanged |
| Merge purity and compatibility | Planning is read-only; claim handoff precedes carry-forward; protected state and effect order remain exact | PASS; structural AST coverage rejects mutation-capable planning calls, the protected set is re-planned under locks, loser claims release before survivor carry-forward, and merge result, history, projection, Collaboration, replay, and rollback cases pass |
| Migration and recovery reconciliation | Pre-deployment migration 37 refinements, schema objects, privileges, Recovery state, and source hashes agree | PASS; deferred rollback refresh is transaction-local, the envelope FK has a supporting index, approved routines remain least-privileged, Recovery has 114 tables with claims excluded/rebuildable, and Database Migrations plus Recovery focused/service-backed suites pass |
| Accounting | Retain exactly 55/55 selectors, 37 owner rows, and 24 service-backed rows | PASS; the new collection case is a subtest, the Entities boundary selector reconciles 55, and `make explain-test-owner OWNER=module.entities` reports 37 rows and 24 service-backed rows |
| Affected product owners | Current Entities, Revisions, Timeline, Assessments, Workbook, Imports, app-server, OpenAPI, migration, and Recovery evidence passes | PASS; exact counts and current result roots are recorded in Section 29.2; infrastructure-only readiness failures were replaced by complete passing retries |
| Generation and repository gates | Format, generation, catalog, Harness, boundary, migration drift, generation drift, generated policy, and JSON shape pass | PASS; final roots are recorded in Section 29.2 and `make test-catalog-check` passed without findings |
| Generated provenance and changed paths | Generated files come only from authored inputs and `make generate`; every S04 path is recorded; no unrelated, Domain, lockfile, moved, or deleted path exists | PASS; generated error, SQL, migration-history, schema-ownership, and topology artifacts came from the error contracts, migration, and generator inputs; `docs/domain.md` and lockfiles are unchanged; no file was moved or deleted |
| Impact and rollback | Public, persisted, internal-interface, deployment, migration, compatibility, and pre-production backup effects are explicit | PASS; only the adopted conflict vocabulary changes public failures, migration head remains 37, the lifecycle and rollback ports change atomically, no mixed writer or dual port exists, and no old-backup compatibility generation is present |
| Failures, skips, and residual risk | Every causal failure is resolved; no mandatory check is skipped | PASS; selector accounting, JSON validators, FK indexing, Recovery count, routine policy, hash fixtures, and service-readiness retries are attributed in Section 29.2; OpenAPI service-backed is non-applicable because it owns no such row; Incident Bundle closure remains explicitly assigned to S05 |
| Authorization transition | ESI-S04 complete; ESI-S05 authorized and ready | PASS |
| Tracker checkpoint | Scoped diff check and current Markdown lint pass before S05 begins | PASS; scoped `git diff --check` passed and `make lint-markdown` passed at `.cartulary/test-results/20260822T195808Z-p1486259` |

### 29.8 ESI-S05 checkpoint

| Check | Required result | Current result |
| --- | --- | --- |
| Requirement and gap closure | ESI-REQ-008 and ESI-G05 close the declared Entities version-2 portability boundary | PASS; all five paths use explicit typed rows, exact decoding, fixed-column SQL, affected-row equality, and deterministic prepare and postapply validation |
| Invariant attribution and atomicity | All eight invariant IDs are independently reachable with deterministic precedence and no partial state | PASS; the dedicated integration row exercises each closed invariant, multi-defect precedence, affected-row atomicity, and postapply typed-set equality |
| Byte and claim behavior | Valid version-2 exports remain byte-stable and active claims are reconstructed rather than serialized | PASS; import, apply, validate, export, and re-export bytes agree for all five paths, bundle bytes contain no claims, and claim rows exactly match imported source state |
| Accounting | Reach exactly 56/56 selectors, 38 owner rows, and 25 service-backed rows | PASS; the single authored service-backed row adds one top-level selector and catalog reconciliation reports exactly 56, 38, and 25 |
| Affected product owners | Current Entities, Incident Bundles, Revisions, and Recovery focused and service-backed selections pass | PASS; exact unit counts and current result roots are recorded in Section 29.2 |
| Generation and repository gates | Format, generation, catalog, Harness, boundary, migration drift, generation drift, generated policy, and JSON shape pass | PASS; current result roots are recorded in Section 29.2 and `make test-catalog-check` passed without findings |
| Generated provenance and changed paths | Generated topology comes only from the authored row and `make generate`; every S05 path is recorded; no unrelated, Domain, migration, lockfile, moved, or deleted path exists | PASS; only `tools/execution_topology_render_index.json` is generated in S05, its owner input is `tools/test_families/module.entities.json`, and the complete changed-path inventory is recorded in Section 29.2 |
| Impact and rollback | Public, persisted, internal-interface, deployment, migration, portability, and pre-production backup effects are explicit | PASS; successful public APIs and valid v2 bytes are stable, malformed bundles gain precise failures, no persisted schema or migration changes, no dual codec or old-backup generation exists, and rollback is slice-atomic before deployment |
| Failures, skips, and residual risk | Every causal failure is resolved; no mandatory check is skipped | PASS; direct-SQL mirror fixtures, transaction approval, Harness accounting, and boundary policies are attributed in Section 29.2 and all invalidated checks passed |
| Authorization transition | ESI-S05 complete; ESI-S06 authorized and ready | PASS |
| Tracker checkpoint | Scoped diff check and current Markdown lint pass before S06 begins | PASS; scoped `git diff --check` passed and `make lint-markdown` passed at `.cartulary/test-results/20260822T204900Z-p2022633` |

### 29.9 ESI-S06 checkpoint

| Check | Required result | Current result |
| --- | --- | --- |
| Superseded path closure | No incident-wide matcher, duplicate create match, forwarding helper, dual codec, or old lifecycle port remains | PASS; zero-reference searches are empty and Host/Identity create, upsert, import, and clipboard each match once before snapshot and mutation |
| Index reconciliation | Remove only lookup indexes made redundant by the measured claim-key plan; retain every independently live index | PASS; the active matcher plan uses `entity_active_identifier_claims_pkey`, nine exact-lookup-only indexes retire, and display, lineage, foreign-key, uniqueness, and record-local indexes remain |
| Migration and rollback | Head stays 37 and disposable Down/Up exactly restores/removes retired indexes | PASS; the dedicated migration selector proves all nine indexes absent after Up, present after Down or failed preflight, and absent after re-Up without partial schema mutation |
| Exact inventories | Files, exports, production imports, migrations, schema, Recovery, bundle paths, selectors, rows, and generated topology agree | PASS; exact reconciliation reports head 37, five v2 paths, derived/excluded/rebuildable claims, no stale schema index, and no stale symbol |
| Accounting | Retain exactly 56/56 selectors, 38 owner rows, and 25 service-backed rows | PASS; selectors are unique and exact, with unchanged authored and service-backed row counts |
| Affected product owners | Current Entities, Imports, Workbook, Revisions, Database Migrations, and app-server focused/service-backed selections pass | PASS; exact counts and current result roots are recorded in Section 29.2 |
| Generation and repository gates | Format, generation, catalog, Harness, boundary, migration drift, generation drift, generated policy, and JSON shape pass | PASS; current result roots are recorded in Section 29.2 and `make test-catalog-check` passed without findings |
| Generated provenance and changed paths | Migration history and schema ownership come only from migration 37 plus `make generate`; no unrelated, Domain, public-contract, lockfile, moved, or deleted path exists | PASS; the complete S06 path inventory and generator root are recorded in Section 29.2 |
| Impact and rollback | Public, persisted, internal-interface, deployment, migration, compatibility, and pre-production backup effects are explicit | PASS; public behavior is unchanged, matching is single-pass internally, only redundant indexes leave persisted state, Down recreates them, and no new migration or backup compatibility generation exists |
| Failures, skips, and residual risk | Every causal failure is resolved; no mandatory check is skipped | PASS; the compile-only local, owner-ID usage error, and two migration-hash pins are attributed in Section 29.2 and all invalidated gates passed |
| Authorization transition | ESI-S06 complete; ESI-S07 authorized and ready | PASS |
| Tracker checkpoint | Scoped diff check and current Markdown lint pass before S07 begins | PASS; scoped `git diff --check` passed and `make lint-markdown` passed at `.cartulary/test-results/20260822T212058Z-p2644606` |

### 29.10 ESI-S07 checkpoint

| Check | Required result | Current result |
| --- | --- | --- |
| Requirement and gap closure | Close ESI-REQ-001 through ESI-REQ-011 and ESI-G01 through ESI-G07 with current evidence | PASS; every work item is `DONE`, every slice is `COMPLETE`, and Section 31 has no pending criterion |
| Exact repository accounting | Head 37, 56/56 selectors, 38 owner rows, 25 service-backed rows, five typed bundle paths, derived claims, and no stale path | PASS; current source count, owner explanation, catalog, symbol review, migration manifest, schema ownership, Recovery catalog, and bundle inventory agree |
| Ordered repository gates | Catalog, Harness, boundary, migration drift, generation drift, generated policy, and JSON shape pass before product suites | PASS; current result roots are recorded in Section 29.2 and catalog completed without findings |
| Owner matrix | Current focused and service-backed task-guidance selections pass for all eleven required owners | PASS; all 22 applicable owner selections pass with exact counts and result roots in Section 29.2 |
| Frontend and browser | Unit, type, import-boundary, webserver-backed, stateful, accessibility, and unchanged visual-golden gates pass | PASS; 390/390, 2/2, 2/2, 58/58, 34/34, 12/12, and 12/12 pass with no golden refresh |
| Build, full check, and release | Build, finalize, complete repository check, and release check pass on the final state | PASS; build 7/7, finalize 1/1, check 642/642, and release 801/801 pass at the roots in Section 29.2 |
| Generated provenance and changed paths | Every S07 path and generated output is recorded; generated files come only from owner inputs and `make generate` | PASS; migration 37 and the catalog generator produced the migration-history and schema-ownership projections; no lockfile, Domain, golden, moved, or deleted path exists |
| Public and persisted compatibility | Only adopted conflicts change; successful interfaces and valid v2 bytes remain exact; migration posture is explicit | PASS; migration head is 37, claims are derived and excluded, `table_rebuild` grants recovery only `SELECT`/`TRUNCATE`, and no mixed writer, compatibility path, or old-backup generation exists |
| Failures, skips, and residual risk | Every failure is causally attributed, corrected at its owning boundary, and followed by all invalidated gates | PASS; the Collaboration race assertion, Reporting vocabulary fixture, evidence pins, privilege projection, and recovery reset class are resolved; only non-applicable/usage attempts were omitted and no mandatory gate was skipped |
| Rollback and handoff | State the rollback boundary and leave no eligible ESI successor | PASS; pre-deployment rollback remains slice-atomic, post-deployment repair is forward, ESI-S07 is complete, and no successor slice or residual blocker remains |
| Tracker checkpoint | Final tracker-scoped diff and current Markdown lint pass | PASS; scoped `git diff --check` passed and `make lint-markdown` passed at `.cartulary/test-results/20260823T000307Z-p1755898` before this terminal status update |

## 30. ESI Compatibility, Risks, Deferrals, and Rollback

### 30.1 Compatibility and migration

- ESI-S00 is documentation-only.
- ESI-S02 and ESI-S03 use forward, append-only migrations. Invalid or
  duplicate source state blocks upgrade and requires explicit operator
  remediation; ESI performs no automatic merge or data loss.
- Migration 37 was refined before deployment during S04. This project is
  pre-production and has no old backups, so ESI adds no historical backup
  generation, migration replacement, or compatibility restoration path.
- ESI-S05 keeps the exact version-2 portable member contract and valid bytes
  while replacing physical-table reflection with typed rows. Claims are never
  serialized; malformed bundles fail earlier through closed safe invariants.
- ESI-S06 refines undeployed migration 37 without changing its version. The
  head schema drops only exact-lookup indexes superseded by the measured claim
  primary-key plan; Down recreates them exactly, and no compatibility indexes
  or historical backup generation remain.
- No rolling mixed-version writer support, feature flag, dual schema, or
  compatibility trigger is added. Server startup already requires repository
  migration head before runtime construction.
- Ambiguous submitted identifiers intentionally change from precedence-based
  selection to `entity_match_conflict` when matches span records.
- Restore identity collisions add `record_restore_blocked` with
  `active_entity_identifier_conflict`; rollback collisions add the same reason
  to `rollback_precondition_failed`.
- Routes, operation IDs, request members, success payloads, schema IDs, field
  keys, event shapes, authorization order, idempotent replay, and valid bundle
  version remain unchanged.

### 30.2 Principal risks and controls

| Risk | Control |
| --- | --- |
| Go and SQL normalization drift | One versioned corpus exercised by both implementations; migration backfill uses the SQL function. |
| Advisory-lock deadlock | Canonical lexical tuple order over the complete identity set; concurrency tests cover opposing mutations. |
| Trigger recursion or recovery-order drift | Owner-local refresh functions, `ENABLE ALWAYS` triggers, deterministic rebuild/validation, recovery-order tests, and corruption fixtures. |
| Invalid existing production data | Preflight before irreversible DDL, aggregate safe report, no automatic repair, and `BLOCKED` status until disposition. |
| Merge collision after uniqueness | Pure planning, complete lock set, loser release before carry-forward, one effect sequence, and rollback tests. |
| Unsafe public database error | Owner-adopted restore and rollback classifiers after ordinary concealment and authorization gates. |
| Portable-byte drift | S01 characterization, explicit fixed v2 members, golden comparison, and no version bump without separate owner adoption. |
| Generated or routing drift | Authored owner inputs, `make generate`, exact counts, catalog, Harness, drift, and policy gates. |

### 30.3 Deferrals

| ID | Decision | Reason | Revisit trigger | Status |
| --- | --- | --- | --- | --- |
| ESI-DEF-01 | Do not redesign persisted route idempotency payloads. | This remains a shared Authentication/platform migration concern. | Separately adopted cross-owner idempotency migration. | DEFERRED |
| ESI-DEF-02 | Do not redesign `workbookprojection`. | It is a live typed cross-owner contract. | Separately adopted contract version. | DEFERRED |
| ESI-DEF-03 | Do not split the Entities bounded root. | No new owner responsibility justifies a package boundary. | New owner decision with decomposition evidence. | DEFERRED |
| ESI-DEF-04 | Do not retire Incident Bundle version 2. | It is an active portability contract; this iteration stabilizes its exact owner shape. | Explicit version retirement or successor adoption. | DEFERRED |
| ESI-DEF-05 | Do not make active claims a generic identity platform. | The claim language and lifecycle are Entities-owned. | A separately adopted cross-module identity requirement. | DEFERRED |

### 30.4 Rollback posture

Rollback is slice-atomic. Revert each slice's owner, migration, implementation,
tests, authored verification rows, and generated projections together. Do not
keep old and new matchers, dual lifecycle ports, forwarding wrappers,
compatibility triggers, or unused indexes as rollback machinery.

Before release, a failed product slice is reverted at its Section 27 boundary.
After a migration reaches an environment, repair forward unless the adopted
migration owner explicitly authorizes and validates a Down transition for that
environment. ESI-S07 changes evidence only; product corrections return to the
earlier owning slice.

## 31. ESI Binary Completion Criteria

ESI is complete only when every row is `PASS` with current recorded evidence.
`TODO`, planned, skipped, partial, inferred, or historical-only results are not
acceptance.

| Criterion | Required evidence | Current result |
| --- | --- | --- |
| The ESI tracker is current without rewriting completed history. | Header, Sections 23 through 31, changed-path review, scoped diff, and Markdown result. | PASS for ESI-S00; Sections 1 through 22 remain historical evidence and only this tracker changed. |
| Owner documents adopt the exact-match, schema, lifecycle, portability, recovery, and public conflict behavior. | Core 01, Core 02, Core 04 AC-559, manifest, and characterization. | PASS; owner adoption and current 52/34/21 characterization completed in ESI-S01. |
| Source schema rejects every invalid owned tuple without guessing. | Migration 36 clean, upgrade, invalid preflight, Down/Up, ownership, and privilege evidence. | PASS; ESI-S02 migration and affected-owner evidence is recorded in Sections 29.2 and 29.5 at exact 53/35/22 accounting. |
| Active exact identity is indexed, unique, concurrent, and rebuildable. | Migration 37, normalization parity, claims, query plans, concurrency, lifecycle, and recovery evidence. | PASS; ESI-S03 migration, corpus, indexed repository, concurrency, lifecycle, rebuild, production Recovery dispatch, affected-owner evidence, and exact 55/37/24 accounting are recorded in Sections 29.2 and 29.6. |
| Delete, restore, rollback, and merge transfer claims atomically. | Revisions port, public error, rollback reason, pure merge planning, claim handoff, history, projection, Collaboration, replay, and rollback evidence. | PASS; ESI-S04 lifecycle, ordered row/collection rollback, pure merge plan/apply, safe conflict, failure-boundary, affected-owner, and exact 55/37/24 evidence is recorded in Sections 29.2 and 29.7. |
| Every declared Entities Incident Bundle invariant is enforced. | Typed v2 codec, eight negative fixtures, valid byte-stable round trip, claim exclusion, and atomic import evidence. | PASS; ESI-S05 explicit typed-row codecs, all eight deterministic failures, zero-partial-state checks, valid byte equality, and exact claim reconstruction are recorded in Sections 29.2 and 29.8. |
| Superseded matching paths and unnecessary indexes are absent without compatibility residue. | Exact symbol, query-plan, index, export, file, selector, migration, recovery, bundle, and generated reconciliation. | PASS; ESI-S06 removes the zero-reference scans and duplicate match calls, retires only nine claim-superseded indexes with exact Down recreation, and records complete inventory and affected-owner evidence in Sections 29.2 and 29.9. |
| Final accounting is exact. | 56/56 selectors, 38 owner rows, and 25 service-backed rows. | PASS; current ESI-S07 reconciliation confirms 56 unique top-level selectors, 38 owner rows, and 25 service-backed rows with catalog, Harness, and owner suites passing. |
| Public compatibility changes are limited to the adopted conflict behavior. | OpenAPI, HTTP mapping, authorization/concealment order, replay, success-shape, frontend, and browser evidence. | PASS; current owner, frontend, browser, full-check, and release evidence preserves successful behavior and limits new failures to the adopted exact-match, restore, and rollback conflicts. |
| Developer, migration, security, and release validation is complete. | Section 28.4 current result roots and final handoff. | PASS; every ordered gate completed on the final state, including 642/642 `make check` and 801/801 `make release-check`; Section 29.2 records all roots and Section 29.10 closes the handoff. |

Domain vocabulary remains unchanged. `docs/domain.md` is not an ESI artifact
and remains unchanged throughout ESI.

## 32. ELR Scope, Baseline, and Planning Posture

Entities Legacy Removal (`ELR`) is the controlling successor to the completed
ESI iteration. It removes proven dead or compatibility-only repository APIs,
replaces the mixed Entities projection aggregate with directional contracts,
and moves mutation admission out of transport-shaped child-package APIs. The
iteration favors one final structure over deprecation layers, forwarding
packages, aliases, dual paths, or speculative extension points.

ELR-S00 is the documentation-only rebaseline requested on 2026-08-29. It
changes only this tracker. ELR-S01 through ELR-S06 require a later explicit
implementation directive and MUST remain inactive until that authority exists.

### 32.1 Current planning baseline

| Fact | Current value | Evidence posture |
| --- | --- | --- |
| Baseline commit | `d3b24ca8dad2f50b446035ca3b4d6c71e652fe05` | Clean worktree before ELR-S00. |
| Database migration head | 40 | `00040_parties_mutation_request_hash_v1.sql`; historical Entities migrations remain immutable. |
| Entities Go files | 101 total: 69 production and 32 test | Current `internal/modules/entities` inventory. |
| Top-level Entities Go tests | 56 | Unique `func Test*` declarations under the owner tree. |
| Authored Entities verification rows | 38 total; 25 service-backed | `tools/test_families/module.entities.json`. |
| Active public HTTP operations | Two | `mergeEntityRecord` and `resolveEntityMention` remain in the current owner OpenAPI source. |
| Active portable source version | Incident Bundle v3 | Every current Entities source descriptor declares version 3 only. |

Counts and repository facts in Sections 1 through 31 remain accurate evidence
for their named historical checkpoints but MUST NOT be treated as the current
ELR baseline. Section 32.1 supersedes those historical counts for ELR planning.

Historical green runs establish prior iteration closure only. Every ELR code
slice requires current evidence after its changes; no EPR or ESI result may be
reused to claim ELR completion.

### 32.2 Authority and source posture

- Adopted subsystem NLSpecs and Core owners continue to govern behavior.
- `docs/decisions/entities-module-boundary.md` and
  `docs/decisions/projections-module-boundary.md` govern the internal topology
  only after their proposed ELR amendments are adopted.
- `docs/domain.md` remains unchanged because ELR introduces no vocabulary or
  owner-navigation change.
- `docs/research/nlspec-spec.md` supplies writing discipline only and creates
  no Entities behavior or implementation authority.
- This tracker records implementation intent and checkpoints; it does not
  supersede an owner or authorize a later implementation slice by itself.

### 32.3 Objective and compatibility posture

ELR is complete only when each retained production export has a live
production consumer, required interface role, or source-contribution role;
each retired surface is absent without a compatibility replacement; projection
consumers receive only the capability they use; admission failures are
transport-neutral until the application boundary; and all affected owner and
release gates pass on the final state.

Public retirement is permitted but evidence-gated. The ELR-S00 audit found no
obsolete externally observable operation, schema, persisted shape, or
historical migration. Therefore ELR carries forward the two active Entities
HTTP operations, their OpenAPI identities, Timeline mention actions, view
schemas and field keys, authorization and concealment order, idempotency,
database schema, migration history, Incident Bundle v3, Recovery, Revisions,
and successful response bytes. The approved retirements in Section 34 are
repository-internal Go surfaces with no qualifying production role.

## 33. ELR Requirements and Gap Register

### 33.1 Requirements

- **ELR-REQ-001:** Amend the adopted Entities and Projections topology owners
  before changing their exact package or export projections.
- **ELR-REQ-002:** Replace the mixed Entities projection aggregate atomically;
  do not retain `workbookprojection`, `Ports`, `EntityPorts`, or forwarding
  aliases.
- **ELR-REQ-003:** Give mutation, query, reporting, and source-contribution
  consumers distinct compile-time projection capabilities.
- **ELR-REQ-004:** Return immutable Entities-owned admission facts from child
  decoders and perform HTTP or Workbook translation exactly once at the
  application boundary.
- **ELR-REQ-005:** Require exactly one duplicate-free JSON object with no
  trailing value for every Entities admission entry; preserve valid request
  hashes, established field/reason/count details, error precedence, and
  successful response bytes.
- **ELR-REQ-006:** Remove every confirmed dead export, forwarding constant,
  alias, marker method, wrapper, and unnecessarily exported concrete provider
  named in Section 34.3.
- **ELR-REQ-007:** Add authored guards that reject the retired package and
  symbols and constrain all replacement packages to their adopted consumers.
- **ELR-REQ-008:** Reconcile generated projection-provider metadata and exact
  Harness routing from authored inputs; never hand-edit generated outputs.
- **ELR-REQ-009:** Keep migrations, schema IDs, field keys, bundle shape,
  routes, authorization, idempotency, Recovery, Revisions, and source behavior
  unchanged.
- **ELR-REQ-010:** Record a complete checkpoint before activating each serial
  successor and return failures to the slice that introduced them.

### 33.2 Gap decisions

#### ELR-G01 — Mixed projection aggregate

- **Remediation:** Replace `workbookprojection` atomically with source-facing
  `projectioncontract` and consumer-facing `projectionports`; expose separate
  mutation, query, and reporting interface views from Projections assembly.
- **Areas:** Specification, implementation, tests, authored contracts,
  generated artifacts, and tracker documentation.
- **Rationale:** Source contribution, mutation, query, and reporting have
  different ownership directions and growth rates. One aggregate violates
  interface segregation and grants consumers unrelated capabilities.
- **Long-term benefit:** Projection behavior can expand without widening every
  consumer, creating import cycles, or coupling source owners to physical
  projection implementation.
- **Compatibility and migration:** Repository-internal Go APIs break
  intentionally with no alias or deprecation layer. Public wire behavior and
  persisted state remain unchanged; provider facade paths and derived fixture
  digests may change through authored generation.
- **Risk if unresolved:** The aggregate becomes a growing coordination API and
  makes each later projection phase more expensive to separate.
- **Validation:** Exact method-set tests, production import guards, unchanged
  projection/query/reporting results, zero old-package imports, and green
  affected owner slices.

#### ELR-G02 — Transport-shaped child admission

- **Remediation:** Introduce immutable `mutationadmission.Failure` values for
  create, patch, clipboard, conflict resolution, merge, and mention action;
  translate once in the Entities HTTP root or Workbook assembly.
- **Areas:** Entities boundary specification, Core 04 conformance,
  implementation, unit/integration tests, and tracker documentation.
- **Rationale:** Child packages currently select HTTP status, code, message,
  and arbitrary detail-map shape even when Workbook is the application owner.
- **Long-term benefit:** Admission remains reusable by future transports and
  workflows while wire policy stays centralized and testable.
- **Compatibility and migration:** Decoder return types change internally.
  Existing valid inputs and established field, reason, and count details remain
  exact.
- **Risk if unresolved:** Transport concerns spread into source-owner code and
  arbitrary detail maps become an accidental compatibility contract.
- **Validation:** Affected child decoders have no `httpapi` admission import;
  failures expose only closed typed facts; HTTP and Workbook translations
  reproduce exact intended envelopes and precedence.

#### ELR-G02A — Ambiguous JSON acceptance

- **Remediation:** Require exactly one duplicate-free JSON object with no
  trailing value throughout Entities admission by using
  `strictjson.DecodeObject`.
- **Areas:** Core 01, Core 04, implementation, security/conformance tests, and
  tracker documentation.
- **Rationale:** Create, merge, and mention action accept ambiguous forms that
  sibling Workbook mutations already reject, weakening deterministic
  validation and replay reasoning.
- **Long-term benefit:** One parse, validation, and hashing model supplies a
  stable security baseline for future fields and transports.
- **Compatibility and migration:** Intentional public tightening: malformed,
  scalar, top-level `null`, duplicate-member, and trailing-value bodies newly
  fail with `400 invalid_mutation_payload` and
  `reason_code=request_not_object`. Valid payloads, normalized hashes,
  idempotent replay, and successful bytes remain unchanged.
- **Risk if unresolved:** Different parsers can interpret the same bytes
  differently, undermining validation, identity, and security guarantees.
- **Validation:** Every decoder rejects malformed, scalar/null, duplicate, and
  trailing fixtures with zero effects; valid and already-rejected golden cases
  remain exact.

#### ELR-G03 — Dead and compatibility-only exports

- **Remediation:** Remove the complete Section 34.3 inventory without
  forwarding substitutes and use `entitycontract` as the canonical schema-ID
  source, including its explicitly allowed Workbook assembly consumer.
- **Areas:** Entities topology specification, implementation, tests, boundary
  policy, and tracker documentation.
- **Rationale:** The targeted surfaces have no production consumer or duplicate
  a canonical type or constant.
- **Long-term benefit:** The exported API communicates only supported
  responsibilities and obsolete shapes cannot influence new work.
- **Compatibility and migration:** Internal compile-time break only; external
  tests must use owner contracts or designated test support instead of removed
  forwarders. No shim is permitted.
- **Risk if unresolved:** Dead APIs retain false authority, attract new
  consumers, and create avoidable future compatibility burden.
- **Validation:** Repeat production reachability immediately before deletion;
  require zero references and retired-token guards. Contrary live-consumer
  evidence blocks and rebaselines that retirement.

#### ELR-G04 — Concrete provider leakage

- **Remediation:** Privatize Host/Identity projection and reporting providers;
  constructors return `projectioncontract.SourceReader` or
  `exportprovider.FieldProvider`; replace mention reporting free functions with
  an infallible interface-returning constructor.
- **Areas:** Entities and Projections boundary specifications, implementation,
  tests, and export inventory.
- **Rationale:** Consumers require behavior, not provider state or convenience
  wrappers.
- **Long-term benefit:** Implementations can change without expanding caller
  coupling or public method sets.
- **Compatibility and migration:** Internal constructor signatures change
  atomically with callers. Provider keys, field order, facts, and support
  references remain unchanged.
- **Risk if unresolved:** Concrete types become accidental contracts and make
  provider evolution needlessly disruptive.
- **Validation:** Export inventory retains only constructors and required
  interface methods; key, order, fact, nil-dependency, and composition tests
  pass.

#### ELR-G05 — Stale executable topology and verification metadata

- **Remediation:** Update backend boundaries, the authored provider manifest,
  owner-local exports, retired-token policy, and exact selectors; regenerate
  only through `make generate`.
- **Areas:** Authored contracts, generated artifacts, Harness tests,
  implementation policy, and tracker documentation.
- **Rationale:** Structural improvements are durable only when executable
  enforcement describes the new topology and rejects the old one.
- **Long-term benefit:** Drift checks prevent reintroduction of aggregate
  ports, forwarding aliases, transport imports, or unauthorized consumers.
- **Compatibility and migration:** Provider facade paths and contract digests
  change; performance-fixture snapshot identities may rotate through
  generation. Product data and public contracts do not migrate.
- **Risk if unresolved:** Product tests could be green beside stale metadata,
  broken routing, or silently reintroduced legacy APIs.
- **Validation:** Boundary, generation drift, generated policy, JSON shape,
  catalog, Harness contract, and exact selector reconciliation all pass.

## 34. ELR Adopted Target Design

The following design is adopted by the Entities and Projections boundary
decisions, Core 01 admission owners, and Core 04 AC-558 through ELR-S01.

### 34.1 Directional projection contracts

`internal/modules/entities/projectioncontract` owns only the source-owner side
of projection contribution: Host and Identity projection input/page values,
`SourceReader`, descriptor construction, `Contribution`, `NewContribution`,
`ProjectionContribution`, and `Source`. Host and Identity surface-intent
builders are package-private. Projection descriptors name this package as the
Entities facade package.

`internal/modules/entities/projectionports` owns only consumer-facing derived
state capabilities:

- `MutationRows` contains Host/Identity refresh and delete methods borrowing
  the caller transaction.
- `QueryReader` contains Host/Identity query selection and their bounded value
  types.
- `ReportingReader` contains Host/Identity derived-fact collection and the
  derived-fact value type.

Projection assembly and its adapter expose `EntityMutationRows`,
`EntityQueryReader`, and `EntityReportingReader`. `EntityPorts`, the aggregate
`Ports` value, and the `internal/modules/entities/workbookprojection` package
are deleted in the same slice. No forwarding package, alias, or transitional
aggregate is permitted.

Host/Identity mutation code, Timeline, Imports, merge, and performance fixtures
receive `MutationRows` only. Host/Identity query composition receives
`QueryReader` only. Reporting receives `ReportingReader` only. The concrete
Entities projection source becomes private and its constructor returns
`projectioncontract.SourceReader`.

### 34.2 Transport-neutral mutation admission

Add `internal/modules/entities/mutationadmission` with an immutable `Failure`
and closed reason-code vocabulary. The failure exposes `ReasonCode`, `Field`,
`CollectionField`, `RequestedCount`, and `MaximumCount` accessors; absent
optional values report `ok == false`. Owner-local constructors create ordinary
and limit failures without arbitrary detail maps.

Host/Identity, mention, and merge JSON decoders use the platform strict-JSON
primitive and return `*mutationadmission.Failure` instead of
`*httpapi.APIError`. Validation ownership and canonical hashing remain in
Entities. The root route facade converts merge and mention failures to the
existing HTTP error once, including the existing invalid loser-record
concealment. Workbook assembly converts Host/Identity failures directly to
`workbook.InvalidPayloadFailure` or `workbook.InvalidPayloadLimitFailure`.

All six entries accept exactly one duplicate-free JSON object with no trailing
value. A malformed body, scalar, top-level `null`, duplicate member, or
trailing value fails with `400 invalid_mutation_payload` and
`reason_code=request_not_object`; this is the sole intentional external
compatibility tightening in ELR.

Entities adapters stop calling `workbook.DecodeMutationFailure`. That Workbook
function remains while Action, Assessment, or Timeline production consumers
still require it; ELR does not broaden its retirement beyond proven reachability.

### 34.3 Exact retirement inventory

| Current surface | Disposition | Replacement or retained source of truth |
| --- | --- | --- |
| `entitycontract.EntityTypeHost` and `EntityTypeIdentity` | Remove | No production consumer; use owner-local literals where already present. |
| Forwarded `hostidentity.HostsViewSchemaID` and `IdentitiesViewSchemaID` | Remove | `entitycontract` remains the single owner-local schema-ID source. |
| `AliasSyncResult.Changed` | Remove | No production consumer. |
| `ExactMatchConflictError.EntityMatchConflictDetails` | Remove | Active callers use the typed conflict fields directly. |
| Mention `InvalidMutationTarget` and `MutationTransitionDetails` marker methods | Remove | Active root and Workbook adapters classify the concrete typed errors. |
| `mentions.TombstoneLinkCommand` alias | Remove | Use `mentions.LinkCommand` directly. |
| Exported `HostSurfaceIntent` and `IdentitySurfaceIntent` | Privatize during package replacement | `projectioncontract` constructs immutable contribution intent internally. |
| Host/Identity and mention reporting `CollectFieldsTx` wrappers | Remove | `exportprovider.FieldProvider.CollectFactsTx` is the live reporting contract. |
| Exported concrete Host/Identity reporting provider | Privatize | `New` returns `exportprovider.FieldProvider`. |
| Free-function mention reporting adapter | Replace | An infallible `New` returns an unexported `exportprovider.FieldProvider`. |
| Exported concrete Host/Identity projection source | Privatize | Constructor returns `projectioncontract.SourceReader`. |
| Entities `workbookprojection`, aggregate `Ports`, and `EntityPorts` | Remove | Directional packages and three explicit runtime accessors. |

ELR-S01 MUST repeat the production-consumer audit before adoption. A changed
live-consumer result blocks the affected retirement and requires this ledger to
be rebaselined; it does not authorize a compatibility shim.

### 34.4 Retained product contracts

Except for the strict ambiguous-JSON rejection in Section 34.2, ELR makes no
database migration and changes no HTTP method, route template, operation ID,
request member, response member, status, error code, safe detail, schema ID,
field key, authorization rule, idempotency scope or valid-request hash, event,
history facet, Recovery contribution, Incident Bundle descriptor or byte
shape, source table, or successful behavior. Migration Down/Up checks bearing
historical names remain active evidence and are not dead code.

## 35. ELR Execution Policy and Serial Workstreams

Only one ELR slice may be active. A successor begins only after its dependency
is `COMPLETE`, its checkpoint in Section 37 is filled, every required check is
`PASS`, changed paths match the slice boundary, and explicit implementation
authority covers the successor. A failure returns to the owning slice; final
validation does not absorb unfinished implementation.

| Slice | Workstream | Depends on | Current status | Authorization | Exit criterion |
| --- | --- | --- | --- | --- | --- |
| ELR-S00 | Tracker rebaseline | Completed ESI-S07 | COMPLETE | Authorized and complete for this tracker only | Only this tracker changed; scoped diff and Markdown lint passed. |
| ELR-S01 | Owner adoption and characterization | ELR-S00 checkpoint | COMPLETE | Authorized and complete | Owners adopt the exact topology; current behavior and export reachability are characterized. |
| ELR-S02 | Directional projection-boundary replacement | ELR-S01 checkpoint | COMPLETE | Authorized and complete | New directional packages and runtime accessors are live; the aggregate package is absent. |
| ELR-S03 | Semantic admission and strict JSON cutover | ELR-S02 checkpoint | COMPLETE | Authorized and complete | Entities decoders return semantic failures; strict ambiguous-form rejection and preserved valid behavior are exact. |
| ELR-S04 | Dead and compatibility-surface removal | ELR-S03 checkpoint | COMPLETE | Authorized and complete | Every Section 34.3 retirement is complete with no replacement shim. |
| ELR-S05 | Boundary, generated, and Harness reconciliation | ELR-S04 checkpoint | COMPLETE | Authorized and complete | Authored topology and selectors describe the final tree; all generated projections are current. |
| ELR-S06 | Final validation and handoff | ELR-S05 checkpoint | COMPLETE | Authorized and complete | Every binary criterion passes on one final state and no successor remains. |

### 35.1 ELR-S00 — Tracker rebaseline

Update the header and append Sections 32 through 38. Record the baseline,
authority, exact target interfaces and retirements, serial dependency graph,
validation, risks, deferrals, and binary completion criteria. Change no other
file. Rollback removes the ELR header posture and Sections 32 through 38 only.

### 35.2 ELR-S01 — Owner adoption and characterization

Amend the adopted Entities and Projections boundary decisions and Core 04
AC-558 to replace the legacy aggregate topology and adopt semantic admission.
Amend Core 01's row-mutation, merge, and mention-action owners to require
strict single-object JSON. Before product edits, characterize malformed JSON,
duplicate and unknown fields, collection and count limits, merge loser-ID
concealment, authorization and error precedence, request hashes, projection
results, reporting keys/order, and the complete production export inventory.

### 35.3 ELR-S02 — Directional projection-boundary replacement

Create the two adopted directional packages, migrate Projections assembly and
adapter accessors, then migrate every production and test consumer atomically.
Update the authored projection descriptor inputs and regenerate their
projections. Delete `workbookprojection` only after no live import remains;
do not leave a forwarding package or aggregate.

### 35.4 ELR-S03 — Semantic admission and strict JSON cutover

Add the owner-local immutable failure, convert the Host/Identity, mention, and
merge decoders, and place the only translations in root HTTP and Workbook
assembly. Preserve exact established membership, reason/detail values,
valid-request hashes, concealment, and precedence while intentionally rejecting
every ambiguous JSON form adopted in ELR-G02A. Remove child decoder imports
that exist solely to construct `httpapi.APIError`.

### 35.5 ELR-S04 — Dead and compatibility-surface removal

Apply the complete Section 34.3 retirement ledger after repeating the live
consumer search. Contract concrete providers behind required interfaces and
remove wrappers, markers, aliases, and forwarded constants without
deprecation. Add focused characterization only where deletion would otherwise
leave observable behavior unproved.

### 35.6 ELR-S05 — Boundary, generated, and Harness reconciliation

Update authored backend-module boundaries for the replacement packages and
their exact consumers. Add retired-package and retired-symbol guards. Refresh
authored test rows only when selectors actually changed, run public Make
generation, and require exact one-row-per-test accounting with no synthetic
rows added merely to preserve historical counts.

### 35.7 ELR-S06 — Final validation and handoff

Run the ordered Section 36 matrix on one final worktree, return any failure to
its owning slice, record all current result roots, reconcile the export and
file inventory, close every gap and requirement, and leave no active successor.

## 36. ELR Validation Matrix

### 36.1 Required scenarios

- Strict JSON rejects non-objects, duplicate members, unknown members, invalid
  values, empty changes, duplicate field keys, and limit overflow with the
  same field, reason, collection field, requested count, and maximum count.
- Invalid merge loser IDs retain existing not-found concealment and ordinary
  authentication, authorization, incident, and decode precedence.
- Canonical request hashes and idempotent replay remain byte-stable across the
  admission cutover.
- Mutation-only consumers cannot query or collect reporting facts; query-only
  and reporting-only consumers cannot mutate projection rows.
- Host/Identity query results, projection refresh/delete effects, reporting
  provider keys, field ordering, and fact values remain exact.
- The retired aggregate package, symbols, aliases, concrete providers, and
  wrappers have zero production references and are rejected by authored
  guards.
- OpenAPI, view schemas, Incident Bundle v3, migrations, Recovery, Revisions,
  routes, successful responses, and browser-visible behavior remain unchanged.

### 36.2 Per-slice minimums

| Slice | Minimum verification before checkpoint |
| --- | --- |
| ELR-S00 | Tracker-scoped `git diff --check`, `make lint-markdown`, and changed-path review. |
| ELR-S01 | Owner lint/consistency checks, export reachability, and current characterization tests. |
| ELR-S02 | `make format`; Entities and Projections focused/service-backed slices; boundary, generation drift, generated policy, and JSON shape checks. |
| ELR-S03 | Entities and Workbook admission unit/integration rows plus affected Timeline and app Server rows; exact HTTP characterization. |
| ELR-S04 | Export and retired-symbol guards, Entities and Reporting focused/service-backed slices, and production build. |
| ELR-S05 | `make generate`, generation drift, generated policy, JSON shape, test catalog, Harness contract, and backend boundary checks. |
| ELR-S06 | Every affected owner slice, frontend/browser rows selected by the Entities owner, build, `make agent-finalize`, `make check`, and `make release-check`. |

### 36.3 Ordered final matrix

Run focused and service-backed slices for `module.entities`,
`module.projections`, `module.workbook`, `module.timeline`, `module.imports`,
`module.reporting`, `app.server`, `module.revisions`, `module.recovery`, and
`module.incidentbundles`. Run the exact Entities-owned frontend, browser,
stateful, accessibility, and visual rows selected by the authored catalog;
visual goldens must not change without separate justification. Then run
boundary/generation/catalog/Harness checks, `make migration-drift`,
`make frontend-typecheck`, `make frontend-unit`, `make build`,
`make go-gosec-targeted`, `make agent-finalize`, `make check`, and
`make release-check`. Use `make task-guide ROLE=module-author OWNER=<owner-id>`
and the explain targets before invoking each affected owner slice. Record every
result root and explain any non-applicable target; no mandatory failure may be
called unrelated without evidence.

## 37. ELR Tracker and Checkpoint Ledger

### 37.1 Work tracker

| ID | Work item | Slice | Status | Completion evidence |
| --- | --- | --- | --- | --- |
| ELR-001 | Rebaseline this tracker and establish the controlling ELR plan. | ELR-S00 | DONE | Baseline, target design, slices, validation, risks, and completion gates are recorded; tracker diff and Markdown lint passed. |
| ELR-002 | Adopt topology and characterize behavior and exports. | ELR-S01 | DONE | Entities/Projections decisions, Core 01, and Core 04 agree; export reachability and focused characterization passed. |
| ELR-003 | Replace the projection aggregate with directional contracts. | ELR-S02 | DONE | Directional packages, three accessors, private implementation, authored metadata, generation, and affected suites passed. |
| ELR-004 | Replace transport-shaped child admission with semantic failures. | ELR-S03 | DONE | Immutable failures, six strict decoders, two sole translators, exact behavior tests, and structural gates passed. |
| ELR-005 | Remove the exact dead and compatibility-surface inventory. | ELR-S04 | DONE | Fresh audit, atomic deletion, direct canonical schema imports, private providers, export inventory, and retired-token guards passed. |
| ELR-006 | Reconcile boundaries, generated metadata, and Harness routing. | ELR-S05 | DONE | Exact allowlists, retired path/accessor guards, generated digest review, selector reconciliation, and Harness gates passed. |
| ELR-007 | Complete final production-readiness validation and handoff. | ELR-S06 | DONE | Owner, frontend/browser, structural, security, full-check, release, inventory, and handoff gates passed. |

### 37.2 Session handoff log

| Date | Slice | State | Changed scope | Verification | Next action |
| --- | --- | --- | --- | --- | --- |
| 2026-08-29 | ELR-S00 planning and rebaseline | ELR-S00 complete; ELR-S01 pending and unauthorized | This tracker only; no product, test, owner, generated, migration, Domain, or application artifact changed | Baseline inspection, changed-path review, and scoped `git diff --check` passed; `make lint-markdown` passed at `.cartulary/test-results/20260829T173955Z-p4060775` | Preserve the checkpoint; do not activate ELR-S01 without a later directive. |
| 2026-08-29 | ELR-S01 owner adoption and characterization | ELR-S01 complete; ELR-S02 active | Changed the Entities and Projections boundary decisions, Core 01 row/merge/mention admission, Core 04 AC-558, and this tracker; no implementation, generated, migration, Domain, or routing file changed | Entities focused 42/42 at `.cartulary/test-results/20260829T181437Z-p4073164`; Projections focused 19/19 at `.cartulary/test-results/20260829T181633Z-p4123832`; backend boundary 3/3 at `.cartulary/test-results/20260829T181729Z-p4141513`; test catalog passed; post-checkpoint Markdown passed at `.cartulary/test-results/20260829T181918Z-p4143472`; scoped diff passed | Replace the projection aggregate atomically in ELR-S02. |
| 2026-08-29 | ELR-S02 directional projection replacement | ELR-S02 complete; ELR-S03 active | Added `projectioncontract` and `projectionports`; deleted `workbookprojection`; migrated Projections/application/owner/test consumers; updated provider manifest, boundary policy, selector, export inventories, and Make-generated topology index | Format `.cartulary/test-results/20260829T183040Z-p30982`; Entities focused 42/42 `.cartulary/test-results/20260829T182705Z-p4151371`, service-backed 33/33 `.cartulary/test-results/20260829T183553Z-p213256`; Projections focused 19/19 `.cartulary/test-results/20260829T182905Z-p9833`, service-backed 12/12 `.cartulary/test-results/20260829T183553Z-p213258`; Workbook 68/68, Timeline 53/53, Imports 23/23, Reporting 6/6; generation and policy gates passed | Implement semantic admission and strict JSON atomically in ELR-S03. |
| 2026-08-29 | ELR-S03 semantic admission and strict JSON | ELR-S03 complete; ELR-S04 active | Added immutable `mutationadmission`; converted create, patch, clipboard, conflict, merge, and mention admission to strict object decoding; centralized HTTP and Workbook translation; added exact malformed-input, translation, and zero-effect tests; reconciled boundary and test metadata | Format `.cartulary/test-results/20260829T184958Z-p342588`; Entities 42/42 `.cartulary/test-results/20260829T185014Z-p347138`, service-backed 33/33 `.cartulary/test-results/20260829T185849Z-p563721`; Workbook 68/68 `.cartulary/test-results/20260829T185014Z-p347148`, service-backed 39/39 `.cartulary/test-results/20260829T185849Z-p563726`; app Server 24/24 `.cartulary/test-results/20260829T185238Z-p455624`; Timeline 53/53 `.cartulary/test-results/20260829T185238Z-p455625`; generation, structural, catalog, and Harness gates passed | Repeat the live-consumer audit and remove the exact retirement inventory in ELR-S04. |
| 2026-08-29 | ELR-S04 dead-surface removal and provider encapsulation | ELR-S04 complete; ELR-S05 active | Removed entity-type constants, schema forwarders, dead methods and marker methods, the mention tombstone alias, reporting field wrappers, concrete reporting provider export, and mention free-function adapter; migrated all schema users to `entitycontract`; added exact retired-token and consumer guards | Format `.cartulary/test-results/20260829T190915Z-p722091`; Entities 42/42 `.cartulary/test-results/20260829T190923Z-p726309`, service-backed 33/33 `.cartulary/test-results/20260829T191123Z-p795521`; Reporting 6/6 `.cartulary/test-results/20260829T190923Z-p726313`, service-backed 4/4 `.cartulary/test-results/20260829T191123Z-p795528`; boundary 3/3 `.cartulary/test-results/20260829T190810Z-p677245`; build 7/7 `.cartulary/test-results/20260829T191123Z-p795815`; initial compile-only failures `.cartulary/test-results/20260829T190810Z-p677077` and `.cartulary/test-results/20260829T190810Z-p677083` were corrected before the clean retries | Reconcile final authored and generated topology and exact selectors in ELR-S05. |
| 2026-08-29 | ELR-S05 boundary, generated, and Harness reconciliation | ELR-S05 complete; ELR-S06 active | Replaced subtree-wide exceptions with exact file allowlists for all four new contracts; guarded the retired package path and aggregate accessors; reviewed the sole Make-generated topology-index digest delta; retained only real selectors and rows | Generate `.cartulary/test-results/20260829T191551Z-p899679`; generation drift 4/4 `.cartulary/test-results/20260829T191612Z-p902654`; generated policy 3/3 `.cartulary/test-results/20260829T191612Z-p902686`; JSON shape 3/3 `.cartulary/test-results/20260829T191612Z-p902723`; backend boundary 3/3 `.cartulary/test-results/20260829T191612Z-p903117`; catalog `.cartulary/test-results/20260829T191613Z-p903561`; Harness 2/2 `.cartulary/test-results/20260829T191612Z-p903244` | Execute the ordered final validation matrix and close the handoff in ELR-S06. |
| 2026-08-29 | ELR-S06 final validation and handoff | ELR-S06 and ELR complete; no successor active | Corrected five stale test imports exposed by final Workbook compilation; ran every routed backend, frontend, browser, structural, security, broad-check, and release gate; reconciled current files, exports, tests, rows, compatibility, rollback, failures, and skips | All clean owner and final roots are recorded in Section 37.9; `make check` passed 671/671 at `.cartulary/test-results/20260829T195918Z-p2416256`; `make release-check` passed 837/837 at `.cartulary/test-results/20260829T200413Z-p2530106` | Preserve this completed handoff; no ELR successor remains. |

### 37.3 ELR-S00 checkpoint

| Gate | Required result | Current result |
| --- | --- | --- |
| Baseline | Commit, migration head, file/test counts, owner rows, and authority recorded | PASS |
| Changed paths | Only this tracker differs from the baseline | PASS |
| Scoped diff | Tracker-scoped `git diff --check` passes | PASS |
| Markdown | `make lint-markdown` passes on the ELR document | PASS at `.cartulary/test-results/20260829T173955Z-p4060775` |
| Domain posture | `docs/domain.md` remains unchanged | PASS |
| Authorization transition | ELR-S00 complete; ELR-S01 pending and unauthorized | PASS |

### 37.4 ELR-S01 checkpoint

| Gate | Required result | Current result |
| --- | --- | --- |
| Owner adoption | Entities, Projections, Core 01, and Core 04 agree on directional contracts, semantic admission, and strict JSON | PASS |
| Retirement audit | Every Section 34.3 target retains zero qualifying production consumers; no contrary target requires rebaselining | PASS |
| Characterization | Existing valid hashes, field/reason/count errors, precedence, projection behavior, and reporting order remain covered | PASS; Entities 42/42 and Projections 19/19 focused units |
| Malformed-input baseline | Create, merge, and mention currently use one ordinary decoder pass and therefore admit some duplicate/trailing forms; strict siblings already reject them | PASS; the intended ELR-G02A delta is explicit |
| Test accounting | 56 top-level Entities tests; 38 owner rows; 25 service-backed rows | PASS; unchanged in this specification slice |
| Executable boundaries | Existing backend boundary projection remains green | PASS at `.cartulary/test-results/20260829T181729Z-p4141513` |
| Changed scope | Five documentation owners only; no implementation, generated, migration, Domain, or routing delta | PASS |
| Compatibility | Only ambiguous invalid JSON forms are approved to tighten in ELR-S03; valid wire and persisted behavior remain fixed | PASS |
| Rollback boundary | Revert the four owner amendments and this S01 tracker checkpoint together | PASS; no data rollback required |
| Residual risk | S02 must rotate provider facade paths and selectors through authored inputs without preserving the aggregate | ACCEPTED for ELR-S02 |
| Scoped diff and Markdown | Tracker-scoped `git diff --check` and `make lint-markdown` pass | PASS; post-checkpoint Markdown root `.cartulary/test-results/20260829T181918Z-p4143472` |
| Successor | ELR-S02 begins only after this checkpoint | PASS; ELR-S02 active |

### 37.5 ELR-S02 checkpoint

| Gate | Required result | Current result |
| --- | --- | --- |
| Directional packages | `projectioncontract` owns contribution/source facts; `projectionports` owns exact 4/2/2 mutation/query/reporting method sets | PASS |
| Aggregate removal | No `workbookprojection` package, `Ports`, `EntityPorts`, `Entities()`, or forwarding alias remains | PASS; zero non-document references |
| Encapsulation | One unexported Projections implementation supplies three interface views; projection source implementation is unexported | PASS |
| Consumer narrowing | Host/Identity receives named mutation/query dependencies; Timeline, Imports, merge, mentions, and fixtures receive mutation only; Reporting receives reporting only | PASS |
| Authored and generated metadata | Provider facade paths, boundary consumers, contribution selector, export inventories, and generated topology digest reflect the new tree | PASS; `make generate` root `.cartulary/test-results/20260829T182956Z-p27849` |
| Focused tests | Entities 42/42; Projections 19/19; Workbook 68/68; Timeline 53/53; Imports 23/23; Reporting 6/6 | PASS; roots recorded in Section 37.2 |
| Service-backed tests | Entities 33/33 and Projections 12/12 | PASS; roots recorded in Section 37.2 |
| Structural gates | Backend boundary 3/3, generation drift 4/4, generated policy 3/3, JSON shape 3/3, catalog, and Harness 2/2 | PASS; roots `.cartulary/test-results/20260829T183044Z-p35121`, `.cartulary/test-results/20260829T183046Z-p35468`, `.cartulary/test-results/20260829T183054Z-p38423`, `.cartulary/test-results/20260829T183056Z-p38827`, and `.cartulary/test-results/20260829T183748Z-p280763` |
| Test accounting | 102 Entities Go files: 70 production and 32 test; 56 top-level tests; 38 rows; 25 service-backed rows | PASS; one package split and one new production file, with selector count unchanged |
| Compatibility | Internal Go composition breaks intentionally; projection/query/reporting outputs and persisted/public behavior remain unchanged | PASS |
| Rollback boundary | Revert directional packages, all consumers/tests, authored manifest/policy/selector, generated topology index, and this checkpoint together | PASS; no dual topology or data rollback |
| Residual risk | S03 must preserve valid hashes and exact translations while applying only the adopted ambiguous-JSON tightening | ACCEPTED for ELR-S03 |
| Successor | Tracker-scoped diff and Markdown pass before S03 product edits | PASS; post-checkpoint Markdown root `.cartulary/test-results/20260829T183910Z-p282021`; ELR-S03 active |

### 37.6 ELR-S03 checkpoint

| Gate | Required result | Current result |
| --- | --- | --- |
| Semantic failure contract | `mutationadmission.Failure` has private state, closed typed reasons, and only ordinary or paired-limit construction | PASS; reflection, accessor, and fail-closed tests are cataloged under the Entities root boundary row |
| Six strict entry points | Create, patch, clipboard, conflict resolution, merge, and mention action require one duplicate-free object with no trailing value | PASS; malformed, scalar, null, top-level duplicate, nested duplicate, and trailing fixtures return `request_not_object` |
| Transport separation | Child decoders return semantic facts and import no `httpapi`; Entities root and Workbook assembly are the sole translators | PASS; production search and backend-module boundary policy agree |
| Exact external behavior | Existing field/reason/count details, loser-ID concealment, authorization/decode precedence, valid hashes, replay, and successful behavior remain exact | PASS; focused Entities, Workbook, app Server, and Timeline suites passed; only adopted ambiguous invalid forms changed |
| Zero-effect rejection | Newly rejected create, merge, and mention bodies create no source, history, projection, or mention effects | PASS; authorized integration cases compare before/after state |
| Focused tests | Entities 42/42; Workbook 68/68; app Server 24/24; Timeline 53/53 | PASS; roots recorded in Section 37.2 |
| Service-backed tests | Entities 33/33 and Workbook 39/39 | PASS; roots recorded in Section 37.2 |
| Authored and generated metadata | Entities and Workbook selectors, boundary consumers, export inventory, and generated topology digest reflect the new tests and package | PASS; `make generate` root `.cartulary/test-results/20260829T185725Z-p555598` |
| Structural gates | Generation drift 4/4, generated policy 3/3, JSON shape 3/3, backend boundary 3/3, catalog, and Harness 2/2 | PASS; roots `.cartulary/test-results/20260829T185828Z-p558750`, `.cartulary/test-results/20260829T185828Z-p558786`, `.cartulary/test-results/20260829T185828Z-p558797`, `.cartulary/test-results/20260829T185828Z-p559209`, `.cartulary/test-results/20260829T185829Z-p559650`, and `.cartulary/test-results/20260829T185828Z-p559254` |
| Test accounting | 104 Entities Go files: 71 production and 33 test; 57 top-level tests; 38 owner rows; 25 service-backed rows | PASS; one semantic-contract production file and one root test file were added; the new Workbook top-level test has one real catalog row |
| Compatibility | Ambiguous scalar/null/duplicate/trailing JSON forms now fail with `400 invalid_mutation_payload` and `request_not_object`; valid wire, hash, replay, route, and persisted behavior is unchanged | PASS; this is the only intended external tightening |
| Rollback boundary | Revert `mutationadmission`, six decoder conversions, two translators, tests, authored boundary/selectors, generated index, and this checkpoint together | PASS; no dual failure type or data rollback |
| Residual risk | S04 deletions must be re-audited against the current tree and blocked individually if a qualifying consumer has appeared | ACCEPTED for ELR-S04 |
| Successor | Tracker-scoped diff and Markdown pass before S04 product edits | PASS; post-checkpoint Markdown root `.cartulary/test-results/20260829T190248Z-p668817`; ELR-S04 active |

### 37.7 ELR-S04 checkpoint

| Gate | Required result | Current result |
| --- | --- | --- |
| Fresh live audit | Every remaining Section 34.3 target has zero qualifying production consumers immediately before deletion | PASS; only the declared definitions, compatibility consumers, and export-inventory entries were present |
| Dead-surface deletion | Entity-type constants, `AliasSyncResult.Changed`, conflict and mention marker methods, and `TombstoneLinkCommand` are absent | PASS; no alias, forwarding method, or deprecation shim remains |
| Canonical schema identity | Host/Identity forwarding constants are absent and every former consumer uses qualified `entitycontract` values | PASS; Workbook assembly is the only new production consumer and its exception is limited to two files |
| Reporting encapsulation | Host/Identity and mention implementations are unexported; constructors return `exportprovider.FieldProvider`; field wrappers and mention free functions are absent | PASS; provider keys, typed fact collection, ordering, support references, nil-dependency rejection, and server composition remain covered by green Reporting and build evidence |
| Executable enforcement | Exact export inventory and authored retired-token/import guards reject every retired declaration or consumer | PASS; backend boundary 3/3 and an explicit zero-result production search passed |
| Focused tests | Entities 42/42 and Reporting 6/6 | PASS; clean retry roots recorded in Section 37.2 |
| Service-backed tests | Entities 33/33 and Reporting 4/4 | PASS; roots recorded in Section 37.2 |
| Production build | All production composition roots compile against interface-returning providers and canonical schema IDs | PASS; 7/7 at `.cartulary/test-results/20260829T191123Z-p795815` |
| Failure and retry accounting | No failed attempt is omitted from the handoff | PASS; initial focused runs failed on one unnamed receiver and one unused import, both corrected before full clean retries |
| Test accounting | 104 Entities Go files: 71 production and 33 test; 57 top-level tests; 38 owner rows; 25 service-backed rows | PASS; no synthetic test or selector was added for deletion-only behavior |
| Compatibility | Repository-internal forwarding APIs break intentionally; reporting keys/facts/order and all public or persisted behavior remain unchanged | PASS; no data or wire migration |
| Rollback boundary | Revert retirements, all qualified schema imports, provider construction, reporting composition, export/retired guards, and this checkpoint together | PASS; no compatibility layer or data rollback |
| Residual risk | S05 must prove the authored and generated tree contains no stale selectors, facade digests, or aggregate paths | ACCEPTED for ELR-S05 |
| Successor | Tracker-scoped diff and Markdown pass before S05 reconciliation | PASS; post-checkpoint Markdown root `.cartulary/test-results/20260829T191411Z-p897673`; ELR-S05 active |

### 37.8 ELR-S05 checkpoint

| Gate | Required result | Current result |
| --- | --- | --- |
| Exact contract consumers | `projectioncontract`, `projectionports`, `mutationadmission`, and `entitycontract` allow only the current importing production files | PASS; all former subtree-wide exceptions were replaced with exact path lists |
| Retirement enforcement | The removed Entities `workbookprojection` path, aggregate accessors, schema forwarders, dead methods, aliases, wrappers, and transport dependencies are guarded | PASS; zero live references and backend boundary 3/3 |
| Authored provider manifest | Host and Identity facade paths resolve through `projectioncontract` and no old aggregate facade remains | PASS; reviewed `contracts/projection-providers/index.json` |
| Selector reconciliation | Every active selector has compatible catalog ownership and no stale selector or count-preservation row remains | PASS; catalog root `.cartulary/test-results/20260829T191613Z-p903561` |
| Generated provenance | Generated files changed only through `make generate` and every delta has an authored input | PASS; only `tools/execution_topology_render_index.json` changed, rotating the provider and Entities/Workbook family digests |
| Generation and shape gates | Generation drift 4/4, generated policy 3/3, and JSON shape 3/3 | PASS; roots recorded in Section 37.2 |
| Boundary and Harness gates | Backend boundary 3/3 and Harness contract 2/2 | PASS; roots recorded in Section 37.2 |
| Test accounting | 104 Entities Go files: 71 production and 33 test; 57 top-level tests; 38 Entities rows, 25 service-backed; 94 Workbook rows, 57 service-backed | PASS; one real Workbook translation test row was added in S03 and no synthetic row was added |
| Compatibility | Facade and catalog digests rotate internally; no product data, public contract, route, field, schema identity, or visual artifact migrates | PASS |
| Rollback boundary | Revert exact allowlists/retired guards, authored manifest and family rows, generated topology index, and this checkpoint with the preceding implementing slices | PASS; never restore a generated file by hand |
| Residual risk | The final ordered matrix may reveal cross-owner or release-only integration failures; any failure returns to its owning slice | ACCEPTED for ELR-S06 |
| Successor | Tracker-scoped diff and Markdown pass before final validation | PASS; post-checkpoint Markdown root `.cartulary/test-results/20260829T191718Z-p907937`; ELR-S06 active |

### 37.9 ELR-S06 final validation and handoff

#### Requirement closure

| Requirement | Current result | Evidence |
| --- | --- | --- |
| ELR-REQ-001 | PASS | S01 owner amendments and checkpoint precede every implementation slice. |
| ELR-REQ-002 | PASS | S02 removed the mixed package and aggregate atomically; retired-path and accessor guards remain green. |
| ELR-REQ-003 | PASS | Exact 4/2/2 directional interfaces and role-specific consumers pass Projections and boundary tests. |
| ELR-REQ-004 | PASS | Six child decoders return immutable semantic failures; only root HTTP and Workbook assembly translate. |
| ELR-REQ-005 | PASS | Strict object matrices, hashes, replay, precedence, concealment, and zero-effect integration cases pass. |
| ELR-REQ-006 | PASS | The complete Section 34.3 inventory is removed or private with zero forbidden production references. |
| ELR-REQ-007 | PASS | Exact file allowlists and retired package, accessor, symbol, wrapper, and transport guards pass. |
| ELR-REQ-008 | PASS | Authored provider/family inputs and Make-generated topology agree; drift, policy, catalog, and Harness pass. |
| ELR-REQ-009 | PASS | Migration drift, Bundle, Recovery, Revisions, route, replay, frontend, and browser evidence pass; no schema or vocabulary artifact changed. |
| ELR-REQ-010 | PASS | S01 through S05 each closed its tracker, scoped diff, and Markdown gate before its successor began; S06 failures and retries are recorded below. |

#### Gap closure

| Gap | Current result | Completion evidence |
| --- | --- | --- |
| ELR-G01 — mixed projection aggregate | PASS | `projectioncontract` and `projectionports` are live; old package and aggregate APIs are absent and guarded. |
| ELR-G02 — transport-shaped child admission | PASS | Closed semantic failures and two sole translators pass exact boundary and application tests. |
| ELR-G02A — ambiguous JSON acceptance | PASS | All six entries reject scalar, null, malformed, duplicate, and trailing forms as adopted, with no effects. |
| ELR-G03 — dead and compatibility-only exports | PASS | Fresh reachability, 237-item retained export inventory, retired-token policy, and zero-result scans pass. |
| ELR-G04 — concrete provider leakage | PASS | Projection and reporting implementations are private; constructors expose only source-reader or field-provider interfaces. |
| ELR-G05 — stale executable topology | PASS | Exact allowlists, provider facade paths, selectors, generated digest, catalog, and Harness all describe the final tree. |

#### Final backend owner evidence

| Owner | Focused result | Service-backed result |
| --- | --- | --- |
| `module.entities` | PASS 42/42 at `.cartulary/test-results/20260829T192910Z-p1474583` | PASS 33/33 at `.cartulary/test-results/20260829T193102Z-p1525069` |
| `module.projections` | PASS 19/19 at `.cartulary/test-results/20260829T191804Z-p913715` | PASS 12/12 at `.cartulary/test-results/20260829T191804Z-p913804` |
| `module.workbook` | PASS 68/68 at `.cartulary/test-results/20260829T192454Z-p1364835` | PASS 39/39 at `.cartulary/test-results/20260829T192703Z-p1419856` |
| `module.timeline` | PASS 53/53 at `.cartulary/test-results/20260829T193304Z-p1575292` | PASS 30/30 at `.cartulary/test-results/20260829T193740Z-p1632802` |
| `module.imports` | PASS 23/23 at `.cartulary/test-results/20260829T191804Z-p914094` | PASS 14/14 at `.cartulary/test-results/20260829T191804Z-p914008` |
| `module.reporting` | PASS 6/6 at `.cartulary/test-results/20260829T194225Z-p1690156` | PASS 4/4 at `.cartulary/test-results/20260829T194329Z-p1749464` |
| `app.server` | PASS 24/24 at `.cartulary/test-results/20260829T194225Z-p1690159` | PASS 17/17 at `.cartulary/test-results/20260829T194329Z-p1749470` |
| `module.revisions` | PASS 27/27 at `.cartulary/test-results/20260829T194433Z-p1808135` | PASS 20/20 at `.cartulary/test-results/20260829T194610Z-p1925361` |
| `module.recovery` | PASS 24/24 at `.cartulary/test-results/20260829T194433Z-p1808138` | PASS 19/19 at `.cartulary/test-results/20260829T194610Z-p1925366` |
| `module.incidentbundles` | PASS 8/8 at `.cartulary/test-results/20260829T194433Z-p1808148` | PASS 6/6 at `.cartulary/test-results/20260829T194610Z-p1925382` |

Every owner task guide and `make explain-test-owner` completed successfully
before these slices were invoked.

#### Final frontend, browser, and repository gates

| Gate | Current result |
| --- | --- |
| Two exact Entities frontend-unit rows | PASS at `.cartulary/test-results/20260829T194807Z-p2040603` |
| Five exact Entities browser rows | PASS at `.cartulary/test-results/20260829T194820Z-p2041047` |
| Exact Entities stateful row | PASS at `.cartulary/test-results/20260829T194949Z-p2084016` |
| Exact Entities accessibility row | PASS at `.cartulary/test-results/20260829T195042Z-p2125141` |
| Two exact Entities visual rows | PASS at `.cartulary/test-results/20260829T195138Z-p2166312`; no golden changed |
| Generation drift, generated policy, JSON shape, backend boundary | PASS 4/4, 3/3, 3/3, and 3/3 at `.cartulary/test-results/20260829T195244Z-p2207705`, `.cartulary/test-results/20260829T195244Z-p2207790`, `.cartulary/test-results/20260829T195244Z-p2207737`, and `.cartulary/test-results/20260829T195244Z-p2208445` |
| Test catalog and Harness contract | PASS at `.cartulary/test-results/20260829T195244Z-p2209328` and 2/2 at `.cartulary/test-results/20260829T195244Z-p2208652` |
| Migration drift | PASS 5/5 at `.cartulary/test-results/20260829T195244Z-p2207999` |
| Frontend typecheck and unit | PASS 2/2 and 390/390 at `.cartulary/test-results/20260829T195244Z-p2208602` and `.cartulary/test-results/20260829T195244Z-p2208693` |
| Production build and targeted gosec | PASS 7/7 and 4/4 at `.cartulary/test-results/20260829T195244Z-p2209330` and `.cartulary/test-results/20260829T195244Z-p2209195` |
| Agent finalize | PASS 1/1 at `.cartulary/test-results/20260829T195315Z-p2281656`; retained-run maintenance was skipped because `RESULTS_DIR` was unset |
| Full check | PASS 671/671 at `.cartulary/test-results/20260829T195918Z-p2416256` |
| Release check | PASS 837/837 at `.cartulary/test-results/20260829T200413Z-p2530106` |
| Final tracker and diff handoff | PASS cached, unstaged tracker, and full-worktree `git diff --check`; Markdown PASS at `.cartulary/test-results/20260829T202056Z-p2746145` |

#### Failure, retry, inventory, and compatibility accounting

- The initial final owner batch failed at
  `.cartulary/test-results/20260829T191804Z-p913728`,
  `.cartulary/test-results/20260829T191804Z-p913774`,
  `.cartulary/test-results/20260829T191804Z-p913845`,
  `.cartulary/test-results/20260829T191804Z-p913908`,
  `.cartulary/test-results/20260829T191804Z-p913922`, and
  `.cartulary/test-results/20260829T191804Z-p914063`. Workbook exposed five
  stale test-only Host/Identity imports after schema-forwarder removal; they
  were deleted and formatted at
  `.cartulary/test-results/20260829T192442Z-p1360692`. Entities and Timeline
  failures were PostgreSQL force-drop timeouts under ten concurrent database
  runs. Every affected owner then passed the isolated clean roots above.
- The first full check reached 667/671 at
  `.cartulary/test-results/20260829T195333Z-p2284564`; all four failures were
  PostgreSQL force-drop timeouts after the relevant test assertions. The
  unchanged canonical retry passed 671/671. S04's earlier compile-only failure
  roots and clean retries remain recorded in Sections 37.2 and 37.7.
- Current repository accounting is 104 Entities Go files: 71 production and
  33 test; 57 ordinary-build top-level Entities tests; 237 retained exported
  declarations or methods across 15 inventoried Entities packages; 38 exact
  Entities rows, 25 service-backed rows, and 11 non-Go rows. Workbook has 94
  exact rows and 57 service-backed rows. These are current measurements, not
  historical targets.
- The worktree contains only explained ELR owner, implementation, test,
  authored contract, Make-generated topology, and tracker deltas. The sole
  staged pre-existing path remains this tracker. `docs/domain.md`, migrations,
  lockfiles, OpenAPI/schema identities, and visual goldens are unchanged.
- Compatibility is intentionally tightened only for ambiguous invalid JSON:
  scalar, top-level null, duplicate-member, and trailing-value forms now fail
  with `400 invalid_mutation_payload` and `request_not_object`. Valid requests,
  hashes, replay, routes, authorization/concealment order, successful bytes,
  projection/reporting results, persisted data, Incident Bundle v3, Recovery,
  and Revisions remain unchanged. No migration or deployment sequencing is
  required.
- Rollback is slice-atomic across owner amendments, implementation, tests,
  authored contracts, generated topology, and tracker checkpoints. No retired
  alias, aggregate, provider, failure type, or forwarding constant should be
  restored as rollback machinery; no data rollback exists.
- No mandatory check is skipped. The only non-product maintenance omission is
  the documented `agent-finalize` retained-run step with unset `RESULTS_DIR`.
  No residual product or release risk remains inside ELR scope, and no
  successor is active.

ELR-S01 through ELR-S06 checkpoints are added only when the active slice has
current evidence; do not pre-fill planned work as passing.

## 38. ELR Risks, Deferrals, Rollback, and Binary Completion

### 38.1 Principal risks and controls

| Risk | Control |
| --- | --- |
| A supposedly dead export has a hidden production consumer | Repeat exact reachability in ELR-S01 and ELR-S04; block that retirement on contrary evidence. |
| Projection replacement broadens rather than narrows coupling | Enforce three directional interfaces and delete the aggregate atomically. |
| Admission cleanup changes observable errors or hashes | Characterize before replacement and compare exact field, reason, limit detail, precedence, and hash behavior. |
| A forwarding package survives as convenience | Add retired-path and retired-token guards with no allowed path. |
| Generated descriptor or Harness topology drifts | Change authored inputs, run Make generation, and require drift, policy, catalog, and Harness gates. |
| Public or persisted behavior is retired without authority | Treat new evidence as a blocker and rebaseline owners and this tracker before implementation. |

### 38.2 Deferrals

| ID | Decision | Revisit trigger |
| --- | --- | --- |
| ELR-DEF-01 | Do not redesign stored idempotency payloads or status-bearing replay records. | Separately adopted cross-owner idempotency migration. |
| ELR-DEF-02 | Do not split the Entities bounded root or move source ownership. | New owner responsibility with adopted decomposition evidence. |
| ELR-DEF-03 | Do not retire either active Entities HTTP operation or Timeline mention action. | Owner evidence that the public capability is obsolete plus explicit route/OpenAPI adoption. |
| ELR-DEF-04 | Do not change database schema, immutable migrations, or active-claim lifecycle. | Separately adopted source-integrity requirement. |
| ELR-DEF-05 | Do not change or retire Incident Bundle v3. | Explicit portability version retirement or successor adoption. |
| ELR-DEF-06 | Do not remove `workbook.DecodeMutationFailure` while non-Entities production consumers remain. | A later reachability audit proves zero production consumers. |

### 38.3 Rollback posture

ELR-S00 rollback removes only the ELR header posture and Sections 32 through
38. Later implementation rollback is slice-atomic: revert owner amendments,
implementation, tests, authored contracts, generated projections, and the
checkpoint together. Do not preserve the old and new projection packages,
admission types, aliases, provider wrappers, or runtime accessors as rollback
machinery. Because ELR changes no persisted schema, it introduces no data
migration or mixed-version rollback path.

### 38.4 Binary completion criteria

ELR is complete only when every criterion is `PASS` with current evidence.
`TODO`, pending, planned, historical, inferred, or skipped results do not count.

| Criterion | Required evidence | Current result |
| --- | --- | --- |
| The ELR tracker is current without rewriting completed history. | Header, Sections 32 through 38, scoped diff, Markdown lint, and changed-path review. | PASS; ELR-S00 is complete and only this tracker changed. |
| Owners adopt the replacement topology and semantic admission boundary. | Adopted Entities/Projections amendments, Core 01, and Core 04 projection. | PASS; ELR-S01 checkpoint. |
| Projection consumers receive only their directional capability. | New package APIs, consumer inventory, compile-time tests, and absence of the aggregate. | PASS; ELR-S02 checkpoint. |
| Entities admission is transport-neutral with only the adopted strictness change. | Typed failure tests and exact root/Workbook HTTP characterization. | PASS; ELR-S03 checkpoint. |
| Every confirmed dead or compatibility-only surface is absent. | Export inventory, retired-token/path guards, and zero production references. | PASS; ELR-S04 checkpoint. |
| Boundary, generated, and Harness projections describe the final tree exactly. | Boundary, generation, catalog, policy, JSON shape, and Harness results. | PASS; ELR-S05 checkpoint. |
| Public and persisted behavior remains within the adopted ELR posture. | OpenAPI, view, route, auth, replay, migration, bundle, Recovery, Revisions, frontend, and browser evidence. | PASS; ELR-S06 final matrix and compatibility accounting. |
| Final production-readiness validation passes on one state. | Affected owner slices, build, agent-finalize, full check, release check, and final handoff. | PASS; ELR-S06 checkpoint with 671/671 check and 837/837 release evidence. |

Domain vocabulary remains unchanged. `docs/domain.md` is not an ELR artifact
and MUST remain unchanged throughout this iteration unless a separately
adopted vocabulary or navigation change requires a new plan.
